package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/utils"
)

// Service provides HTTP handlers for authentication operations
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Repository (Data Access)
type Service struct {
	authService *AuthService
	jwtService  *JWTService
}

// NewService creates a new auth service
func NewService(h *models.Handlers) *Service {
	// Initialize repository
	userRepo := NewEnvUserRepository()

	// Initialize services
	jwtSecret := GetJWTSecret()
	authService := NewAuthService(userRepo, jwtSecret)
	jwtService := NewJWTService(jwtSecret)

	return &Service{
		authService: authService,
		jwtService:  jwtService,
	}
}

// LoginHandler handles user authentication
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Repository (Data Access)
func (s *Service) LoginHandler(w http.ResponseWriter, r *http.Request) {
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
	}

	// Call service (business logic layer)
	result, err := s.authService.Login(ctx, loginReq)
	if err != nil {
		if err == ErrInvalidCredentials {
			utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set cookie (HTTP layer)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    result.Token,
		Expires:  result.Expires,
		HttpOnly: true,
		Secure:   true, // Should be true in production (HTTPS)
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	// Write JSON response (HTTP layer) - Do not return token in body
	utils.JSONResponse(w, http.StatusOK, result.Response)
}

// LogoutHandler clears the session cookie
// Refactored to use utils helpers
func (s *Service) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})
	utils.JSONResponse(w, http.StatusOK, map[string]string{"message": "Logged out"})
}

// MeHandler returns current user info if authenticated
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic)
func (s *Service) MeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Call service (business logic layer)
	claims, err := s.authService.GetCurrentUser(ctx)
	if err != nil {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"username": claims.Username,
		"role":     claims.Role,
	})
}

// AuthMiddleware protects routes
// Refactored to use JWTService
func (s *Service) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow OPTIONS for CORS
		if r.Method == "OPTIONS" {
			next(w, r)
			return
		}

		// Use JWTService to authenticate request
		claims, err := s.jwtService.AuthenticateRequest(r)
		if err != nil {
			utils.ErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), "user", claims)
		next(w, r.WithContext(ctx))
	}
}

// AuthenticateRequest extracts and validates JWT from request
// Delegates to JWTService
func (s *Service) AuthenticateRequest(r *http.Request) (*AuthClaims, error) {
	return s.jwtService.AuthenticateRequest(r)
}
