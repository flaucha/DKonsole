package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// LDAPAuthenticator defines the interface for LDAP authentication
type LDAPAuthenticator interface {
	AuthenticateUser(ctx context.Context, username, password string) error
	GetUserPermissions(ctx context.Context, username string) (map[string]string, error)
	ValidateUserGroup(ctx context.Context, username string) error
	GetUserGroups(ctx context.Context, username string) ([]string, error)
	GetConfig(ctx context.Context) (*models.LDAPConfig, error)
}

// AuthService provides business logic for authentication operations.
// It handles user authentication, password verification, and JWT token generation.
type AuthService struct {
	userRepo  UserRepository
	jwtSecret []byte
	ldapAuth  LDAPAuthenticator // Optional LDAP authenticator
}

// NewAuthService creates a new AuthService with the provided user repository and JWT secret.
// The JWT secret should be a secure random byte array, typically loaded from environment variables.
func NewAuthService(userRepo UserRepository, jwtSecret []byte) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		ldapAuth:  nil,
	}
}

// SetLDAPAuthenticator sets the LDAP authenticator for the service
func (s *AuthService) SetLDAPAuthenticator(ldapAuth LDAPAuthenticator) {
	s.ldapAuth = ldapAuth
}

// LoginRequest represents login credentials provided by the user.
type LoginRequest struct {
	Username string // Username for authentication
	Password string // Password for authentication
	IDP      string // Identity Provider: "core" for admin, "ldap" for LDAP, "" for auto-detect
}

// LoginResponse represents the response after successful login.
type LoginResponse struct {
	Role string `json:"role"` // User role (typically "admin")
}

// LoginResult represents the complete login result including JWT token and expiration.
type LoginResult struct {
	Response LoginResponse // Login response with user role
	Token    string        // JWT token for subsequent authenticated requests
	Expires  time.Time     // Expiration time of the JWT token
}

// GetCurrentUser extracts user information from the request context.
// The context should contain a "user" value set by the AuthMiddleware.
//
// Returns ErrUnauthorized if no user is found in the context.
func (s *AuthService) GetCurrentUser(ctx context.Context) (*models.Claims, error) {
	userVal := ctx.Value(userContextKey)
	if userVal == nil {
		return nil, ErrUnauthorized
	}

	var username, role string
	var permissions map[string]string
	if claims, ok := userVal.(*AuthClaims); ok {
		username = claims.Username
		role = claims.Role
		permissions = claims.Permissions

		utils.LogInfo("GetCurrentUser: extracted from AuthClaims", map[string]interface{}{
			"username":    username,
			"role":        role,
			"permissions": permissions,
		})
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

	claims := &models.Claims{
		Username:    username,
		Role:        role,
		Permissions: permissions,
	}

	utils.LogInfo("GetCurrentUser: returning claims", map[string]interface{}{
		"username":    claims.Username,
		"role":        claims.Role,
		"permissions": claims.Permissions,
	})

	return claims, nil
}

// Errors
var (
	ErrInvalidCredentials = &AuthError{Message: "Invalid credentials"}
	ErrUnauthorized       = &AuthError{Message: "Unauthorized"}
)
// generateToken creates a new JWT token for the user
func (s *AuthService) generateToken(username, role, idp string, permissions map[string]string, expiration time.Time) (string, error) {
	claims := &AuthClaims{
		Claims: models.Claims{
			Username:    username,
			Role:        role,
			IDP:         idp,
			Permissions: permissions,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
