package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// SetupStatusResponse represents the response for setup status check.
type SetupStatusResponse struct {
	SetupRequired bool `json:"setupRequired"` // true if secret doesn't exist and setup is needed
}

// SetupCompleteRequest represents the request to complete setup.
type SetupCompleteRequest struct {
	Username  string `json:"username"`  // Admin username
	Password  string `json:"password"`  // Plain password (will be hashed)
	JWTSecret string `json:"jwtSecret"` // JWT secret (optional, will be auto-generated if empty)
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
	if err != nil {
		utils.LogError(err, "Failed to check secret existence", map[string]interface{}{
			"endpoint": "/api/setup/status",
		})
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to check setup status")
		return
	}

	response := SetupStatusResponse{
		SetupRequired: !exists,
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
		if strings.Contains(errStr, "Forbidden") || strings.Contains(errStr, "permission") || strings.Contains(errStr, "forbidden") {
			utils.ErrorResponse(w, http.StatusForbidden, "Permission denied: Unable to check secret existence. Please verify RBAC permissions allow reading secrets in this namespace.")
		} else {
			utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check setup status: %v", err))
		}
		return
	}

	utils.LogInfo("Secret existence check completed", map[string]interface{}{
		"exists": exists,
	})

	if exists {
		utils.LogWarn("Setup attempted but secret already exists", map[string]interface{}{
			"endpoint": "/api/setup/complete",
		})
		utils.ErrorResponse(w, http.StatusForbidden, "Setup already completed. Secret already exists.")
		return
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

	// Create secret using repository
	utils.LogInfo("Attempting to create secret", map[string]interface{}{
		"username":    req.Username,
		"secret_name": s.k8sRepo.secretName,
		"namespace":   s.k8sRepo.namespace,
	})
	if err := s.createSecret(ctx, req.Username, passwordHash, jwtSecret); err != nil {
		errStr := err.Error()
		utils.LogError(err, "Failed to create secret", map[string]interface{}{
			"username":      req.Username,
			"error_type":    fmt.Sprintf("%T", err),
			"error_message": errStr,
		})
		// Check if it's a permission error
		if strings.Contains(errStr, "Forbidden") || strings.Contains(errStr, "permission") || strings.Contains(errStr, "forbidden") {
			utils.ErrorResponse(w, http.StatusForbidden, "Permission denied: Unable to create secret. Please verify RBAC permissions allow creating secrets in this namespace.")
		} else {
			utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create secret: %v", err))
		}
		return
	}

	utils.LogInfo("Setup completed successfully", map[string]interface{}{
		"username": req.Username,
	})

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Setup completed successfully. Please restart the pod for changes to take effect.",
	})
}

// checkSecretExists checks if the dkonsole-auth secret exists.
func (s *Service) checkSecretExists(ctx context.Context) (bool, error) {
	if s.k8sRepo == nil {
		return false, fmt.Errorf("K8s repository not initialized")
	}
	return s.k8sRepo.SecretExists(ctx)
}

// createSecret creates the dkonsole-auth secret.
func (s *Service) createSecret(ctx context.Context, username, passwordHash, jwtSecret string) error {
	if s.k8sRepo == nil {
		return fmt.Errorf("K8s repository not initialized")
	}
	return s.k8sRepo.CreateSecret(ctx, username, passwordHash, jwtSecret)
}

// IsSetupMode returns true if the service is running in setup mode.
func (s *Service) IsSetupMode() bool {
	return s.setupMode
}

// generateJWTSecret generates a secure random JWT secret (32 bytes, base64 encoded = 44 characters).
func generateJWTSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
