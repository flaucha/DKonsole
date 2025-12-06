package k8s

import (
	"context"
	"fmt"

	"github.com/flaucha/DKonsole/backend/internal/permissions"
)

// validateNamespaceAccess checks if the user has permission to access the requested namespace
// It returns a list of effective namespaces to list and an error if access is denied
func (s *ResourceListService) validateNamespaceAccess(ctx context.Context, ns string, allNamespaces bool) ([]string, error) {
	// Get user's allowed namespaces
	allowedNamespaces, err := permissions.GetAllowedNamespaces(ctx)
	if err != nil {
		// If we can't get permissions, deny access (fail secure)
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	if ns == "" {
		ns = "default"
	}

	// Case 1: All Namespaces requested
	if allNamespaces {
		if len(allowedNamespaces) > 0 {
			// User has restricted permissions - query ONLY allowed namespaces
			return allowedNamespaces, nil
		}
		// User has no restrictions (admin) - query all (represented by empty string)
		return []string{""}, nil
	}

	// Case 2: Specific Namespace requested
	if len(allowedNamespaces) > 0 {
		// User has restricted permissions - check if requested namespace is allowed
		hasAccess, err := permissions.HasNamespaceAccess(ctx, ns)
		if err != nil {
			return nil, fmt.Errorf("failed to check namespace access: %w", err)
		}
		if !hasAccess {
			return nil, fmt.Errorf("access denied to namespace: %s", ns)
		}
	}

	return []string{ns}, nil
}
