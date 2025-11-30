package auth

import (
	"context"
	"errors"
	"fmt"
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
	Expires  time.Time     // Token expiration time
}

// Login authenticates a user and generates a JWT token.
// It first tries admin authentication, then falls back to LDAP if enabled.
// If IDP is specified ("core" or "ldap"), only that method is tried.
// Returns a JWT token valid for 24 hours if authentication succeeds.
//
// Returns ErrInvalidCredentials if username or password is incorrect.
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	var role string
	var permissions map[string]string
	var idp string // Identity Provider: "core" or "ldap"

	// If IDP is specified, only try that method
	if req.IDP == "core" {
		idp = "core"
		// Only try admin authentication
		adminUser, err := s.userRepo.GetAdminUser()
		if err != nil {
			// Check if it's a configuration error
			if errors.Is(err, ErrAdminUserNotSet) || errors.Is(err, ErrAdminPasswordNotSet) {
				return nil, fmt.Errorf("server configuration error: %w", err)
			}
			return nil, ErrInvalidCredentials
		}
		adminPassHash, err := s.userRepo.GetAdminPasswordHash()
		if err != nil {
			// Check if it's a configuration error
			if errors.Is(err, ErrAdminUserNotSet) || errors.Is(err, ErrAdminPasswordNotSet) {
				return nil, fmt.Errorf("server configuration error: %w", err)
			}
			return nil, ErrInvalidCredentials
		}
		// Verify username
		if req.Username != adminUser {
			return nil, ErrInvalidCredentials
		}
		// Verify password using Argon2
		match, err := VerifyPassword(req.Password, adminPassHash)
		if err != nil || !match {
			return nil, ErrInvalidCredentials
		}
		// Admin authentication successful
		role = "admin"
		permissions = nil // Admin has full access
		idp = "core"
	} else if req.IDP == "ldap" {
		idp = "ldap"
		// Only try LDAP authentication
		if s.ldapAuth == nil {
			return nil, ErrInvalidCredentials
		}
		err := s.ldapAuth.AuthenticateUser(ctx, req.Username, req.Password)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
		// Check if user belongs to required group (if configured)
		if err := s.ldapAuth.ValidateUserGroup(ctx, req.Username); err != nil {
			return nil, ErrInvalidCredentials
		}
		// LDAP authentication successful
		// Get user permissions from LDAP groups first to check if user is admin
		permissions, err = s.ldapAuth.GetUserPermissions(ctx, req.Username)
		if err != nil {
			// Log error but continue - user is authenticated
			// Use fmt.Errorf to include error details in logs
			utils.LogWarn("Failed to get user permissions, continuing with empty permissions", map[string]interface{}{
				"username": req.Username,
				"error":    err.Error(),
			})
			permissions = make(map[string]string)
			role = "user"
		} else {
			// Check if user is in admin group (permissions will be nil for admin)
			// If permissions is nil, user is admin (has full access)
			// If permissions is empty map, user has no permissions (should not see anything)
			// If permissions has entries, user has limited permissions
			if permissions == nil {
				role = "admin"
				utils.LogInfo("User is LDAP admin, setting role to admin", map[string]interface{}{
					"username": req.Username,
				})
			} else if len(permissions) == 0 {
				// User has no permissions - should not see anything
				role = "user"
				utils.LogInfo("User has no permissions, setting role to user with empty permissions", map[string]interface{}{
					"username": req.Username,
				})
			} else {
				role = "user"
				utils.LogInfo("User permissions retrieved successfully", map[string]interface{}{
					"username":    req.Username,
					"permissions": permissions,
				})
			}
		}
	} else {
		// Auto-detect: try admin first, then LDAP
		// Try admin authentication first
		adminUser, err := s.userRepo.GetAdminUser()
		if err != nil {
			// Check if it's a configuration error
			if errors.Is(err, ErrAdminUserNotSet) || errors.Is(err, ErrAdminPasswordNotSet) {
				return nil, fmt.Errorf("server configuration error: %w", err)
			}
			// For other errors, continue to LDAP
		} else {
			adminPassHash, err := s.userRepo.GetAdminPasswordHash()
			if err != nil {
				// Check if it's a configuration error
				if errors.Is(err, ErrAdminUserNotSet) || errors.Is(err, ErrAdminPasswordNotSet) {
					return nil, fmt.Errorf("server configuration error: %w", err)
				}
				// For other errors, continue to LDAP
			} else {
				// Verify username
				if req.Username == adminUser {
					// Verify password using Argon2
					match, err := VerifyPassword(req.Password, adminPassHash)
					if err == nil && match {
						// Admin authentication successful
						role = "admin"
						permissions = nil // Admin has full access
						idp = "core"
					}
				}
			}
		}

		// If admin auth failed and LDAP is available, try LDAP
		if role == "" && s.ldapAuth != nil {
			idp = "ldap"
			err := s.ldapAuth.AuthenticateUser(ctx, req.Username, req.Password)
			if err == nil {
				// Check if user belongs to required group (if configured)
				if err := s.ldapAuth.ValidateUserGroup(ctx, req.Username); err != nil {
					// User doesn't belong to required group, continue to try other methods or fail
					// Don't set role, authentication will fail
				} else {
					// LDAP authentication successful
					// Get user permissions from LDAP groups first to check if user is admin
					permissions, err = s.ldapAuth.GetUserPermissions(ctx, req.Username)
					if err != nil {
						// Log error but continue - user is authenticated
						utils.LogWarn("Failed to get user permissions, continuing with empty permissions", map[string]interface{}{
							"username": req.Username,
							"error":    err.Error(),
						})
						permissions = make(map[string]string)
						role = "user"
					} else {
						// Check if user is in admin group (permissions will be nil for admin)
						// If permissions is nil, user is admin (has full access)
						// If permissions is empty map, user has no permissions (should not see anything)
						// If permissions has entries, user has limited permissions
						if permissions == nil {
							role = "admin"
							utils.LogInfo("User is LDAP admin, setting role to admin", map[string]interface{}{
								"username": req.Username,
							})
						} else if len(permissions) == 0 {
							// User has no permissions - should not see anything
							role = "user"
							utils.LogInfo("User has no permissions, setting role to user with empty permissions", map[string]interface{}{
								"username": req.Username,
							})
						} else {
							role = "user"
							utils.LogInfo("User permissions retrieved successfully", map[string]interface{}{
								"username":    req.Username,
								"permissions": permissions,
							})
						}
					}
				}
			}
		}
	}

	// If still no role, authentication failed
	if role == "" {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT
	expirationTime := time.Now().Add(24 * time.Hour)

	// Log permissions before creating JWT
	utils.LogInfo("Login: creating JWT with permissions", map[string]interface{}{
		"username":    req.Username,
		"role":        role,
		"permissions": permissions,
	})

	claims := &AuthClaims{
		Claims: models.Claims{
			Username:    req.Username,
			Role:        role,
			IDP:         idp,
			Permissions: permissions,
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

	// Log token creation success
	utils.LogInfo("Login: JWT token created successfully", map[string]interface{}{
		"username":    req.Username,
		"role":        role,
		"permissions": permissions,
	})

	return &LoginResult{
		Response: LoginResponse{
			Role: role,
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
