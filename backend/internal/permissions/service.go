package permissions

import (
	"context"
	"fmt"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

// LDAPAdminChecker defines interface for checking LDAP admin groups
type LDAPAdminChecker interface {
	GetUserGroups(ctx context.Context, username string) ([]string, error)
	GetConfig(ctx context.Context) (*models.LDAPConfig, error)
}

// Service provides permission checking and filtering functionality
type Service struct {
	ldapAdminChecker LDAPAdminChecker
}

// NewService creates a new permissions service
func NewService() *Service {
	return &Service{}
}

// SetLDAPAdminChecker sets the LDAP admin checker
func (s *Service) SetLDAPAdminChecker(checker LDAPAdminChecker) {
	s.ldapAdminChecker = checker
}

// GetUserFromContext extracts user information from the request context
// Handles both AuthClaims and models.Claims types
func GetUserFromContext(ctx context.Context) (*models.Claims, error) {
	userVal := ctx.Value(auth.UserContextKey())
	if userVal == nil {
		return nil, fmt.Errorf("user not found in context")
	}

	// Try to extract from AuthClaims (used in auth middleware)
	if authClaims, ok := userVal.(*auth.AuthClaims); ok {
		return &authClaims.Claims, nil
	}

	// Try to extract from models.Claims directly
	if claims, ok := userVal.(*models.Claims); ok {
		return claims, nil
	}

	// Try to extract from map (legacy format)
	if claimsMap, ok := userVal.(map[string]interface{}); ok {
		claims := &models.Claims{}
		if username, ok := claimsMap["username"].(string); ok {
			claims.Username = username
		}
		if role, ok := claimsMap["role"].(string); ok {
			claims.Role = role
		}
		if perms, ok := claimsMap["permissions"].(map[string]interface{}); ok {
			claims.Permissions = make(map[string]string)
			for k, v := range perms {
				if perm, ok := v.(string); ok {
					claims.Permissions[k] = perm
				}
			}
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid user type in context: %T", userVal)
}
