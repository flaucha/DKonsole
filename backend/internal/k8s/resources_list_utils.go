package k8s

import (
	"context"
	"fmt"

	"github.com/flaucha/DKonsole/backend/internal/permissions"
)

// validateNamespaceAccess checks if the user has permission to access the requested namespace
// It returns the effective namespace to list (empty if all namespaces) and an error if access is denied
func (s *ResourceListService) validateNamespaceAccess(ctx context.Context, ns string, allNamespaces bool) (string, error) {
	// Get user's allowed namespaces
	allowedNamespaces, err := permissions.GetAllowedNamespaces(ctx)
	if err != nil {
		// If we can't get permissions, deny access (fail secure)
		return "", fmt.Errorf("failed to get user permissions: %w", err)
	}

	if ns == "" {
		ns = "default"
	}

	// If user is not admin and has restricted permissions, validate namespace access
	if len(allowedNamespaces) > 0 {
		// User has restricted permissions - check if requested namespace is allowed
		if !allNamespaces && ns != "" {
			hasAccess, err := permissions.HasNamespaceAccess(ctx, ns)
			if err != nil {
				return "", fmt.Errorf("failed to check namespace access: %w", err)
			}
			if !hasAccess {
				return "", fmt.Errorf("access denied to namespace: %s", ns)
			}
		}
		// If allNamespaces is requested but user has restrictions, we'll query all
		// and filter the results afterwards (this is acceptable for now)
		// TODO: Optimize by querying only allowed namespaces when allNamespaces=true
	}

	listNamespace := ns
	if allNamespaces {
		listNamespace = ""
	}

	return listNamespace, nil
}
