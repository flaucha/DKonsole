package permissions

import (
	"context"
	"fmt"
)

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
							// Using fmt.Printf here as we don't have utils usage in this file currently
							// Ideally we should import/use a logger if available or pass it
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
