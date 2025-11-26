package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const userContextKey contextKey = "user"

// Service provides HTTP handlers for authentication operations.
// It follows a layered architecture:
//   - Handler (HTTP): Handles HTTP requests/responses
//   - Service (Business Logic): AuthService and JWTService
//   - Repository (Data Access): UserRepository for credential retrieval
type Service struct {
	authService *AuthService
	jwtService  *JWTService
}

// NewService creates a new authentication service with default configuration.
// It initializes the user repository (EnvUserRepository) and JWT services.
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

// LogoutHandler handles HTTP requests to log out the current user.
// It clears the authentication cookie by setting it to expire immediately.
// Returns a JSON response with a success message.
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

// AuthMiddleware is an HTTP middleware that protects routes by requiring valid JWT authentication.
// It extracts and validates the JWT token from the request, and if valid, adds the user claims
// to the request context for use by subsequent handlers.
//
// OPTIONS requests are allowed through for CORS preflight.
// Unauthenticated requests receive a 401 Unauthorized response.
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

		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next(w, r.WithContext(ctx))
	}
}

// AuthenticateRequest extracts and validates JWT from request
// Delegates to JWTService
func (s *Service) AuthenticateRequest(r *http.Request) (*AuthClaims, error) {
	return s.jwtService.AuthenticateRequest(r)
}
