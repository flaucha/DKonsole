package permissions

import (
	"context"
	"fmt"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
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

// HasNamespaceAccess checks if the user has access to a specific namespace
// Returns true if:
// - User is admin (role == "admin" means full access, including LDAP admins)
// - User has any permission (view/edit) for the namespace
func HasNamespaceAccess(ctx context.Context, namespace string) (bool, error) {
	claims, err := GetUserFromContext(ctx)
	if err != nil {
		return false, err
	}

	// Admin has full access (role == "admin" includes both core and LDAP admins)
	if claims.Role == "admin" {
		return true, nil
	}

	// Check if user has any permission for this namespace
	// If permissions is empty, user has no access (not admin)
	// Note: len() for nil maps is defined as zero, so we can omit the nil check
	if len(claims.Permissions) == 0 {
		return false, nil
	}

	_, hasAccess := claims.Permissions[namespace]
	return hasAccess, nil
}

// GetPermissionLevel returns the permission level for a namespace
// Returns "edit", "view", or "" if no access
// Admin users return "edit" (full access)
func GetPermissionLevel(ctx context.Context, namespace string) (string, error) {
	claims, err := GetUserFromContext(ctx)
	if err != nil {
		return "", err
	}

	// Admin has full access (equivalent to "edit" for all namespaces)
	// role == "admin" includes both core and LDAP admins
	if claims.Role == "admin" {
		return "edit", nil
	}

	// If permissions is empty, user has no access (not admin)
	// Note: len() for nil maps is defined as zero, so we can omit the nil check
	if len(claims.Permissions) == 0 {
		return "", fmt.Errorf("no access to namespace: %s", namespace)
	}

	// Get permission for namespace
	permission, exists := claims.Permissions[namespace]
	if !exists {
		return "", fmt.Errorf("no access to namespace: %s", namespace)
	}

	return permission, nil
}

// CanPerformAction checks if the user can perform a specific action on a namespace
// Actions: "view", "edit"
// Returns true if user has the required permission level or higher
func CanPerformAction(ctx context.Context, namespace, action string) (bool, error) {
	permission, err := GetPermissionLevel(ctx, namespace)
	if err != nil {
		return false, err
	}

	// Permission hierarchy: edit > view
	permissionLevels := map[string]int{
		"view": 1,
		"edit": 2,
	}

	requiredLevel, ok := permissionLevels[action]
	if !ok {
		return false, fmt.Errorf("invalid action: %s. Must be 'view' or 'edit'", action)
	}

	userLevel, ok := permissionLevels[permission]
	if !ok {
		return false, fmt.Errorf("invalid permission level: %s", permission)
	}

	return userLevel >= requiredLevel, nil
}

// FilterAllowedNamespaces filters a list of namespaces to only include those the user has access to
func FilterAllowedNamespaces(ctx context.Context, namespaces []string) ([]string, error) {
	claims, err := GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Admin has access to all namespaces
	// role == "admin" includes both core and LDAP admins
	if claims.Role == "admin" {
		return namespaces, nil
	}

	// If permissions is empty, user has no access (not admin)
	// Note: len() for nil maps is defined as zero, so we can omit the nil check
	if len(claims.Permissions) == 0 {
		return []string{}, nil
	}

	// Filter namespaces based on permissions
	allowed := make([]string, 0)
	for _, ns := range namespaces {
		if _, hasAccess := claims.Permissions[ns]; hasAccess {
			allowed = append(allowed, ns)
		}
	}

	return allowed, nil
}

// GetAllowedNamespaces returns the list of namespaces the user has access to
func GetAllowedNamespaces(ctx context.Context) ([]string, error) {
	claims, err := GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Admin has access to all namespaces (return empty list to indicate "all")
	// role == "admin" includes both core and LDAP admins
	if claims.Role == "admin" {
		return []string{}, nil // Empty list means "all namespaces"
	}

	// If permissions is empty, user has no access (not admin)
	// Note: len() for nil maps is defined as zero, so we can omit the nil check
	if len(claims.Permissions) == 0 {
		return []string{}, nil // Empty list means no access
	}

	// Return list of allowed namespaces
	allowed := make([]string, 0, len(claims.Permissions))
	for ns := range claims.Permissions {
		allowed = append(allowed, ns)
	}

	return allowed, nil
}

// ValidateNamespaceAccess validates that the user has access to a namespace
// Returns an error if access is denied
func ValidateNamespaceAccess(ctx context.Context, namespace string) error {
	hasAccess, err := HasNamespaceAccess(ctx, namespace)
	if err != nil {
		return err
	}

	if !hasAccess {
		return fmt.Errorf("access denied to namespace: %s", namespace)
	}

	return nil
}

// ValidateAction validates that the user can perform an action on a namespace
// Returns an error if the action is not allowed
func ValidateAction(ctx context.Context, namespace, action string) error {
	canPerform, err := CanPerformAction(ctx, namespace, action)
	if err != nil {
		return err
	}

	if !canPerform {
		return fmt.Errorf("action '%s' not allowed on namespace: %s", action, namespace)
	}

	return nil
}

// FilterResources filters a list of resources to only include those in namespaces the user has access to
func FilterResources(ctx context.Context, resources []models.Resource) ([]models.Resource, error) {
	claims, err := GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Admin has access to all resources
	// role == "admin" includes both core and LDAP admins
	if claims.Role == "admin" {
		return resources, nil
	}

	// Filter resources based on namespace permissions
	filtered := make([]models.Resource, 0)
	for _, resource := range resources {
		// Cluster-scoped resources (no namespace) are allowed for all users
		if resource.Namespace == "" {
			filtered = append(filtered, resource)
			continue
		}

		// Check if user has access to this namespace
		if _, hasAccess := claims.Permissions[resource.Namespace]; hasAccess {
			filtered = append(filtered, resource)
		} else {
			utils.LogWarn("Filtered out resource due to lack of namespace access", map[string]interface{}{
				"namespace": resource.Namespace,
				"kind":      resource.Kind,
				"name":      resource.Name,
			})
		}
	}

	return filtered, nil
}

// IsAdmin checks if the user is an admin (core admin or LDAP admin group member)
func IsAdmin(ctx context.Context, ldapAdminChecker LDAPAdminChecker) (bool, error) {
	claims, err := GetUserFromContext(ctx)
	if err != nil {
		return false, err
	}

	// Core admin has full access
	if claims.Role == "admin" {
		return true, nil
	}

	// Check if user belongs to LDAP admin groups
	if ldapAdminChecker != nil {
		config, err := ldapAdminChecker.GetConfig(ctx)
		if err == nil && config != nil && config.Enabled && len(config.AdminGroups) > 0 {
			// Get user's groups
			userGroups, err := ldapAdminChecker.GetUserGroups(ctx, claims.Username)
			if err == nil {
				// Check if user belongs to any admin group
				for _, adminGroup := range config.AdminGroups {
					for _, userGroup := range userGroups {
						if userGroup == adminGroup {
							utils.LogInfo("User is LDAP admin group member", map[string]interface{}{
								"username":    claims.Username,
								"admin_group": adminGroup,
							})
							return true, nil
						}
					}
				}
			}
		}
	}

	return false, nil
}

// RequireAdmin validates that the user is an admin (core admin or LDAP admin group member)
// Returns an error if the user is not an admin
func RequireAdmin(ctx context.Context, ldapAdminChecker LDAPAdminChecker) error {
	isAdmin, err := IsAdmin(ctx, ldapAdminChecker)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}

	if !isAdmin {
		return fmt.Errorf("admin access required")
	}

	return nil
}
