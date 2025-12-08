package permissions

import (
	"context"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

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
