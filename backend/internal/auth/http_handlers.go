package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// LoginHandler handles HTTP POST requests for user authentication.
// It expects a JSON body with "username" and "password" fields.
// On success, returns a JWT token in the response body and sets it as an HTTP-only cookie.
//
// @Summary Autenticar usuario
// @Description Autentica un usuario y retorna un JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.Credentials true "Credenciales de usuario"
// @Success 200 {object} LoginResponse "Autenticación exitosa"
// @Failure 400 {object} map[string]string "Cuerpo de solicitud inválido"
// @Failure 401 {object} map[string]string "Credenciales inválidas"
// @Router /api/login [post]
//
// Example request body:
//
//	{"username": "admin", "password": "password123"}
//
// Example response:
//
//	{"role": "admin"}
func (s *Service) LoginHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	setupMode := s.setupMode
	authService := s.authService
	s.mu.RUnlock()

	// Check if in setup mode
	if setupMode {
		utils.ErrorResponse(w, http.StatusPreconditionFailed, "Setup required. Please complete the initial setup first.")
		return
	}

	var creds models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create context
	ctx := r.Context()

	// Prepare request
	loginReq := LoginRequest{
		Username: creds.Username,
		Password: creds.Password,
		IDP:      creds.IDP, // "core", "ldap", or "" for auto-detect
	}

	// Call service (business logic layer)
	if authService == nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Authentication service not initialized")
		return
	}

	result, err := authService.Login(ctx, loginReq)
	if err != nil {
		if err == ErrInvalidCredentials {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set cookie (HTTP layer)
	// Use SameSite=None to allow cross-origin WebSocket handshakes after CORS restrictions.
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    result.Token,
		Expires:  result.Expires,
		HttpOnly: true,
		Secure:   true, // Required for SameSite=None
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	})

	// Write JSON response (HTTP layer) - Do not return token in body
	utils.JSONResponse(w, http.StatusOK, result.Response)
}

// LogoutHandler handles HTTP requests to log out the current user.
// It clears the authentication cookie by setting it to expire immediately.
// Returns a JSON response with a success message.
func (s *Service) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
		MaxAge:   -1,
	})
	utils.JSONResponse(w, http.StatusOK, map[string]string{"message": "Logged out"})
}

// MeHandler returns the current authenticated user's information.
// It extracts user claims from the request context (set by AuthMiddleware).
// Returns a JSON response with username and role, or 401 Unauthorized if not authenticated.
//
// @Summary Obtener usuario actual
// @Description Retorna la información del usuario autenticado
// @Tags auth
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]string "Información del usuario"
// @Failure 401 {object} map[string]string "No autenticado"
// @Router /api/me [get]
//
// Example response:
//
//	{"username": "admin", "role": "admin"}
func (s *Service) MeHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	setupMode := s.setupMode
	authService := s.authService
	s.mu.RUnlock()

	// Check if in setup mode
	if setupMode {
		utils.ErrorResponse(w, http.StatusPreconditionFailed, "Setup required. Please complete the initial setup first.")
		return
	}

	ctx := r.Context()

	// Call service (business logic layer)
	if authService == nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Authentication service not initialized")
		return
	}

	claims, err := authService.GetCurrentUser(ctx)
	if err != nil {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Write JSON response (HTTP layer)
	response := map[string]interface{}{
		"username": claims.Username,
		"role":     claims.Role,
	}
	// Include IDP if available
	if claims.IDP != "" {
		response["idp"] = claims.IDP
	}
	// Include permissions if available (even if empty map)
	if claims.Permissions != nil {
		response["permissions"] = claims.Permissions
	} else {
		// Explicitly set empty permissions for non-admin users
		response["permissions"] = make(map[string]string)
	}

	utils.LogInfo("MeHandler: returning user info", map[string]interface{}{
		"username":    claims.Username,
		"role":        claims.Role,
		"idp":         claims.IDP,
		"permissions": response["permissions"],
	})

	utils.JSONResponse(w, http.StatusOK, response)
}

// ChangePasswordRequest represents a request to change password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ChangePasswordHandler handles password change requests
func (s *Service) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	setupMode := s.setupMode
	authService := s.authService
	k8sRepo := s.k8sRepo
	s.mu.RUnlock()

	// Check if in setup mode
	if setupMode {
		utils.ErrorResponse(w, http.StatusPreconditionFailed, "Setup required. Please complete the initial setup first.")
		return
	}

	// Check if K8s repository is available
	if k8sRepo == nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Password change not available in this configuration")
		return
	}

	ctx := r.Context()

	// Parse request body
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate new password
	if len(req.NewPassword) < 8 {
		utils.ErrorResponse(w, http.StatusBadRequest, "New password must be at least 8 characters long")
		return
	}

	// Verify current password
	if authService == nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Authentication service not initialized")
		return
	}

	// Get current username from context
	claims, err := authService.GetCurrentUser(ctx)
	if err != nil {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Verify current password
	loginReq := LoginRequest{
		Username: claims.Username,
		Password: req.CurrentPassword,
	}
	_, err = authService.Login(ctx, loginReq)
	if err != nil {
		if err == ErrInvalidCredentials {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Current password is incorrect")
			return
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Hash new password
	newPasswordHash, err := HashPassword(req.NewPassword)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("failed to hash password: %v", err))
		return
	}

	// Update password in secret
	if err := k8sRepo.UpdatePassword(ctx, newPasswordHash); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("failed to update password: %v", err))
		return
	}

	utils.LogInfo("Password changed successfully", map[string]interface{}{
		"username": claims.Username,
	})

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}
