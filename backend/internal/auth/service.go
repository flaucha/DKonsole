package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/example/k8s-view/internal/models"
)

// AuthService provides business logic for authentication operations.
// It handles user authentication, password verification, and JWT token generation.
type AuthService struct {
	userRepo  UserRepository
	jwtSecret []byte
}

// NewAuthService creates a new AuthService with the provided user repository and JWT secret.
// The JWT secret should be a secure random byte array, typically loaded from environment variables.
func NewAuthService(userRepo UserRepository, jwtSecret []byte) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// LoginRequest represents login credentials provided by the user.
type LoginRequest struct {
	Username string // Username for authentication
	Password string // Password for authentication
}

// LoginResponse represents the response after successful login.
type LoginResponse struct {
	Role string `json:"role"` // User role (typically "admin")
}

// LoginResult represents the complete login result including JWT token and expiration.
type LoginResult struct {
	Response LoginResponse // Login response with user role
	Token    string        // JWT token for subsequent authenticated requests
	Expires  time.Time     // Token expiration time
}

// Login authenticates a user and generates a JWT token.
// It verifies the username and password against the configured admin credentials,
// and returns a JWT token valid for 24 hours if authentication succeeds.
//
// Returns ErrInvalidCredentials if username or password is incorrect.
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	// Get admin credentials from repository
	adminUser, err := s.userRepo.GetAdminUser()
	if err != nil {
		return nil, fmt.Errorf("server configuration error: %w", err)
	}

	adminPassHash, err := s.userRepo.GetAdminPasswordHash()
	if err != nil {
		return nil, fmt.Errorf("server configuration error: %w", err)
	}

	// Verify username
	if req.Username != adminUser {
		return nil, ErrInvalidCredentials
	}

	// Verify password using Argon2
	match, err := VerifyPassword(req.Password, adminPassHash)
	if err != nil {
		return nil, fmt.Errorf("password verification error: %w", err)
	}
	if !match {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &AuthClaims{
		Claims: models.Claims{
			Username: req.Username,
			Role:     "admin",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResult{
		Response: LoginResponse{
			Role: "admin",
		},
		Token:   tokenString,
		Expires: expirationTime,
	}, nil
}

// GetCurrentUser extracts user information from the request context.
// The context should contain a "user" value set by the AuthMiddleware.
//
// Returns ErrUnauthorized if no user is found in the context.
func (s *AuthService) GetCurrentUser(ctx context.Context) (*models.Claims, error) {
	userVal := ctx.Value("user")
	if userVal == nil {
		return nil, ErrUnauthorized
	}

	var username, role string
	if claims, ok := userVal.(*AuthClaims); ok {
		username = claims.Username
		role = claims.Role
	} else if claims, ok := userVal.(map[string]interface{}); ok {
		if u, ok := claims["username"].(string); ok {
			username = u
		}
		if r, ok := claims["role"].(string); ok {
			role = r
		}
	}

	if username == "" {
		return nil, ErrUnauthorized
	}

	return &models.Claims{
		Username: username,
		Role:     role,
	}, nil
}

// Errors
var (
	ErrInvalidCredentials = &AuthError{Message: "Invalid credentials"}
	ErrUnauthorized       = &AuthError{Message: "Unauthorized"}
)
