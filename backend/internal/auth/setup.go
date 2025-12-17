package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flaucha/DKonsole/backend/internal/middleware"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// SetupStatusResponse represents the response for setup status check.
type SetupStatusResponse struct {
	SetupRequired       bool `json:"setupRequired"`       // true if secret doesn't exist and setup is needed
	TokenUpdateRequired bool `json:"tokenUpdateRequired"` // true if k8s authentication failed (token expired/invalid)
}

// SetupCompleteRequest represents the request to complete setup.
type SetupCompleteRequest struct {
	Username            string `json:"username"`            // Admin username
	Password            string `json:"password"`            // Plain password (will be hashed)
	JWTSecret           string `json:"jwtSecret"`           // JWT secret (optional, will be auto-generated if empty)
	ServiceAccountToken string `json:"serviceAccountToken"` // Token for K8s authentication
}

// SetupStatusHandler handles GET requests to check if setup is required.
// Returns whether the dkonsole-auth secret exists.
func (s *Service) SetupStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := r.Context()

	// Check if secret exists
	exists, err := s.checkSecretExists(ctx)
	tokenUpdateRequired := false

	if err != nil {
		errStr := err.Error()
		// Check for auth errors which indicate token invalidity
		if strings.Contains(errStr, "Unauthorized") || strings.Contains(errStr, "Forbidden") ||
			strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "forbidden") {
			// Do NOT set tokenUpdateRequired = true.
			// If we can't read the secret, we don't know if it's a valid completed setup or an incomplete one.
			// Forcing "SetupRequired" (Full Setup) allows the user to Provide Token + Admin Credentials.
			// This covers:
			// 1. Initial Setup on restricted cluster (can't read secret to verify if it exists)
			// 2. Recovery from expired token (resets admin password)
			// 3. Fixing incomplete secrets (like Helm deploy with missing JWT)
			utils.LogWarn("Setup status check failed with auth error, defaulting to Full Setup mode", map[string]interface{}{
				"error":     errStr,
				"namespace": s.k8sRepo.namespace,
				"client_ok": s.k8sClient != nil,
			})
			// tokenUpdateRequired remains false. !exists (which is true because err != nil) will trigger SetupRequired.
		} else {
			utils.LogError(err, "Failed to check secret existence", map[string]interface{}{
				"endpoint": "/api/setup/status",
			})
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to check setup status")
			return
		}
	}

	// Logic update: Even if secret exists, we must check if Admin User is set.
	// If secret exists but no admin user, it's NOT a completed setup (e.g. Helm created secret with token only).
	adminUser, _ := s.k8sRepo.GetAdminUser()

	// Setup is required if:
	// 1. Secret doesn't exist (!exists)
	// 2. Secret exists but Admin User is missing (adminUser == "")
	setupRequired := !exists || adminUser == ""

	response := SetupStatusResponse{
		SetupRequired:       setupRequired,
		TokenUpdateRequired: tokenUpdateRequired,
	}

	utils.JSONResponse(w, http.StatusOK, response)
}

// SetupCompleteHandler handles POST requests to complete the initial setup.
// It creates the dkonsole-auth secret with the provided credentials.
func (s *Service) SetupCompleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.LogInfo("SetupCompleteHandler called", map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
		"ip":     r.RemoteAddr,
	})

	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if !middleware.IsRequestOriginOrRefererAllowed(r) {
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		return
	}

	ctx := r.Context()

	// First, verify that secret doesn't exist (security check)
	utils.LogInfo("Checking if secret exists before setup", map[string]interface{}{
		"endpoint": "/api/setup/complete",
	})
	exists, err := s.checkSecretExists(ctx)
	if err != nil {
		utils.LogError(err, "Failed to check secret existence", map[string]interface{}{
			"endpoint":   "/api/setup/complete",
			"error_type": fmt.Sprintf("%T", err),
		})
		// Check if it's a permission error
		errStr := err.Error()
		if strings.Contains(errStr, "Forbidden") || strings.Contains(errStr, "permission") || strings.Contains(errStr, "forbidden") || strings.Contains(errStr, "Unauthorized") {
			// If we can't check if secret exists due to permissions, we assume we are in a restricted environment
			// and proceed to try creating/updating it using the *provided* token in the request.
			// The CreateOrUpdateSecret function will use the new token to perform the check/update.
			utils.LogWarn("Could not check secret existence due to permissions, proceeding with provided token", map[string]interface{}{
				"error": errStr,
			})
			exists = false // Treat as not found (or unknown) so we fall through to CreateOrUpdate
		} else {
			utils.HandleErrorJSON(w, err, "Failed to check setup status", http.StatusInternalServerError, map[string]interface{}{
				"endpoint": "/api/setup/complete",
			})
			return
		}
	}

	utils.LogInfo("Secret existence check completed", map[string]interface{}{
		"exists": exists,
	})

	if exists {
		// Check if it really is a complete setup (has admin user)
		adminUser, _ := s.k8sRepo.GetAdminUser()
		if adminUser != "" {
			utils.LogWarn("Setup attempted but secret already exists and has admin user", map[string]interface{}{
				"endpoint": "/api/setup/complete",
			})
			utils.ErrorResponse(w, http.StatusForbidden, "Setup already completed. Secret already exists.")
			return
		}
		// If adminUser is empty, we allow proceeding to UPDATE the secret
		utils.LogInfo("Secret exists but is incomplete (no admin user), proceeding to update", nil)
	}

	// Parse request body
	var req SetupCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate username
	if req.Username == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Username is required")
		return
	}

	// Validate password
	if req.Password == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Password is required")
		return
	}

	if len(req.Password) < 8 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Password must be at least 8 characters long")
		return
	}

	// Generate password hash using Argon2
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		utils.LogError(err, "Failed to hash password", nil)
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Handle JWT secret
	jwtSecret := req.JWTSecret
	if jwtSecret == "" {
		// Auto-generate JWT secret if not provided
		jwtSecret, err = generateJWTSecret()
		if err != nil {
			utils.LogError(err, "Failed to generate JWT secret", nil)
			utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to generate JWT secret")
			return
		}
	}

	// Validate JWT secret length
	if len(jwtSecret) < 32 {
		utils.ErrorResponse(w, http.StatusBadRequest, "JWT secret must be at least 32 characters long")
		return
	}

	// Validate Service Account Token
	if req.ServiceAccountToken == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Service Account Token is required")
		return
	}

	// Create or Update secret using repository
	utils.LogInfo("Attempting to create/update secret", map[string]interface{}{
		"username":    req.Username,
		"secret_name": s.k8sRepo.secretName,
		"namespace":   s.k8sRepo.namespace,
	})

	if err := s.k8sRepo.CreateOrUpdateSecret(ctx, req.Username, passwordHash, jwtSecret, req.ServiceAccountToken); err != nil {
		errStr := err.Error()
		utils.LogError(err, "Failed to create/update secret", map[string]interface{}{
			"username":      req.Username,
			"error_type":    fmt.Sprintf("%T", err),
			"error_message": errStr,
		})
		// Check if it's a permission error
		if strings.Contains(errStr, "Forbidden") || strings.Contains(errStr, "permission") || strings.Contains(errStr, "forbidden") {
			utils.ErrorResponse(w, http.StatusForbidden, "Permission denied: Unable to create/update secret. Please verify RBAC permissions allow creating secrets in this namespace.")
		} else {
			utils.HandleErrorJSON(w, err, "Failed to create/update secret", http.StatusInternalServerError, map[string]interface{}{
				"username":  req.Username,
				"namespace": s.k8sRepo.namespace,
			})
		}
		return
	}

	utils.LogInfo("Setup completed successfully, attempting to reload service", map[string]interface{}{
		"username": req.Username,
	})

	// Attempt to reload the service configuration
	reloaded, err := s.Reload(ctx, req.ServiceAccountToken)
	if err != nil {
		utils.LogError(err, "Failed to reload service after setup", map[string]interface{}{
			"username": req.Username,
		})
		utils.HandleErrorJSON(w, err, "Setup completed, but failed to reload service", http.StatusInternalServerError, map[string]interface{}{
			"username": req.Username,
		})
		return
	}

	if reloaded {
		utils.LogInfo("Service reloaded successfully after setup", map[string]interface{}{
			"username": req.Username,
		})
		utils.JSONResponse(w, http.StatusOK, map[string]string{
			"message": "Setup completed successfully. The service has been reloaded and is ready to use.",
		})
	} else {
		// This shouldn't happen, but handle it gracefully
		utils.LogWarn("Service reload returned false after setup", map[string]interface{}{
			"username": req.Username,
		})
		utils.JSONResponse(w, http.StatusOK, map[string]string{
			"message": "Setup completed successfully. The service will reload automatically. If you encounter issues, you may need to restart the pod.",
		})
	}
}

// checkSecretExists checks if the dkonsole-auth secret exists.
func (s *Service) checkSecretExists(ctx context.Context) (bool, error) {
	if s.k8sRepo == nil {
		return false, fmt.Errorf("K8s repository not initialized")
	}
	return s.k8sRepo.SecretExists(ctx)
}

// generateJWTSecret generates a secure random JWT secret (32 bytes, base64 encoded = 44 characters).
func generateJWTSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// UpdateTokenRequest represents the request to update the service account token.
type UpdateTokenRequest struct {
	ServiceAccountToken string `json:"serviceAccountToken"`
}

// UpdateTokenHandler handles POST requests to update only the service account token.
func (s *Service) UpdateTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if !middleware.IsRequestOriginOrRefererAllowed(r) {
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		return
	}

	ctx := r.Context()

	// Only allow token updates while setup is still required (e.g. incomplete secret created by Helm).
	// Once setup is completed (admin user present), this endpoint must be disabled.
	if s.k8sRepo == nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "K8s repository not initialized")
		return
	}
	exists, err := s.checkSecretExists(ctx)
	if err != nil {
		utils.LogError(err, "Failed to check secret existence for token update", map[string]interface{}{
			"endpoint": "/api/setup/token",
		})
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to validate setup state")
		return
	}
	if !exists {
		utils.ErrorResponse(w, http.StatusPreconditionFailed, "Setup required. Use /api/setup/complete to initialize the system.")
		return
	}
	adminUser, _ := s.k8sRepo.GetAdminUser()
	if adminUser != "" {
		utils.ErrorResponse(w, http.StatusPreconditionFailed, "Setup already completed.")
		return
	}

	// Parse request
	var req UpdateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ServiceAccountToken == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Service Account Token is required")
		return
	}

	// Update token using repository
	// This will use the NEW token to create a client and update the secret
	if err := s.k8sRepo.UpdateSecretToken(ctx, req.ServiceAccountToken); err != nil {
		utils.LogError(err, "Failed to update token", nil)
		utils.HandleErrorJSON(w, err, "Failed to update token", http.StatusInternalServerError, nil)
		return
	}

	utils.LogInfo("Token updated successfully, attempting to reload service", nil)

	// Attempt reload
	if _, err := s.Reload(ctx, req.ServiceAccountToken); err != nil {
		utils.LogError(err, "Failed to reload service after token update", nil)
		utils.HandleErrorJSON(w, err, "Token updated, but failed to reload service", http.StatusInternalServerError, nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Token updated successfully",
	})
}
