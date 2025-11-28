package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

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
	k8sRepo     *K8sUserRepository // K8s repository for secret management (may be nil if not using K8s)
	setupMode   bool                // true if running in setup mode (secret doesn't exist)
}

// NewService creates a new authentication service with default configuration.
// It initializes the user repository and JWT services.
// If k8sClient is provided, it will try to use Kubernetes secrets.
// If k8sClient is nil, it falls back to environment variables.
// secretName is the name of the Kubernetes secret to use (default: "dkonsole-auth").
func NewService(k8sClient kubernetes.Interface, secretName string) (*Service, error) {
	var userRepo UserRepository
	var jwtSecret []byte
	var setupMode bool
	var k8sRepo *K8sUserRepository

	if k8sClient != nil && secretName != "" {
		// Try to use Kubernetes secrets
		repo, err := NewK8sUserRepository(k8sClient, secretName)
		if err != nil {
			utils.LogWarn("Failed to initialize K8s repository, falling back to environment variables", map[string]interface{}{
				"error": err.Error(),
			})
			// Fall back to environment variables
			userRepo = NewEnvUserRepository()
			jwtSecret = GetJWTSecret()
		} else {
			k8sRepo = repo
			// Check if secret exists
			ctx := context.Background()
			exists, err := repo.SecretExists(ctx)
			if err != nil {
				utils.LogWarn("Failed to check secret existence, falling back to environment variables", map[string]interface{}{
					"error": err.Error(),
				})
				userRepo = NewEnvUserRepository()
				jwtSecret = GetJWTSecret()
			} else if !exists {
				// Secret doesn't exist - setup mode
				setupMode = true
				utils.LogInfo("Running in setup mode - secret does not exist", map[string]interface{}{
					"secret_name": secretName,
				})
				// Don't initialize authService in setup mode - it will fail without credentials
				return &Service{
					authService: nil, // Will be nil in setup mode
					jwtService:   nil, // Will be nil in setup mode
					k8sRepo:     k8sRepo,
					setupMode:   true,
				}, nil
			} else {
				// Secret exists - use K8s repository
				userRepo = k8sRepo
				// Get JWT secret from the secret
				secret, err := k8sClient.CoreV1().Secrets(repo.namespace).Get(ctx, secretName, metav1.GetOptions{})
				if err != nil {
					return nil, fmt.Errorf("failed to get JWT secret from Kubernetes secret: %w", err)
				}
				jwtSecretBytes, exists := secret.Data["jwt-secret"]
				if !exists || len(jwtSecretBytes) == 0 {
					return nil, fmt.Errorf("jwt-secret key not found in secret")
				}
				jwtSecret = jwtSecretBytes
			}
		}
	} else {
		// No K8s client provided - use environment variables
		userRepo = NewEnvUserRepository()
		jwtSecret = GetJWTSecret()
	}

	// Initialize services (only if not in setup mode)
	authService := NewAuthService(userRepo, jwtSecret)
	jwtService := NewJWTService(jwtSecret)

	return &Service{
		authService: authService,
		jwtService:  jwtService,
		k8sRepo:     k8sRepo,
		setupMode:   setupMode,
	}, nil
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
	// Check if in setup mode
	if s.setupMode {
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
	}

	// Call service (business logic layer)
	if s.authService == nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Authentication service not initialized")
		return
	}

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
	// Check if in setup mode
	if s.setupMode {
		utils.ErrorResponse(w, http.StatusPreconditionFailed, "Setup required. Please complete the initial setup first.")
		return
	}

	ctx := r.Context()

	// Call service (business logic layer)
	if s.authService == nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Authentication service not initialized")
		return
	}

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

		// If in setup mode, block authenticated routes
		if s.setupMode {
			utils.ErrorResponse(w, http.StatusPreconditionFailed, "Setup required. Please complete the initial setup first.")
			return
		}

		// Use JWTService to authenticate request
		if s.jwtService == nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "JWT service not initialized")
			return
		}

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
