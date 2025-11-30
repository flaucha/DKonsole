package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/permissions"
)

// NamespaceService provides business logic for namespace operations
type NamespaceService struct {
	repo NamespaceRepository
}

// NewNamespaceService creates a new NamespaceService
func NewNamespaceService(repo NamespaceRepository) *NamespaceService {
	return &NamespaceService{repo: repo}
}

// GetNamespaces fetches and transforms namespaces
// It filters namespaces based on user permissions
func (s *NamespaceService) GetNamespaces(ctx context.Context) ([]models.Namespace, error) {
	k8sNamespaces, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	// Get user's allowed namespaces
	allowedNamespaces, err := permissions.GetAllowedNamespaces(ctx)
	if err != nil {
		// If we can't get permissions, deny access (fail secure)
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	var result []models.Namespace
	for _, ns := range k8sNamespaces {
		// If user has restricted permissions, filter namespaces
		if len(allowedNamespaces) > 0 {
			// Check if user has access to this namespace
			hasAccess, err := permissions.HasNamespaceAccess(ctx, ns.Name)
			if err != nil {
				// Skip if we can't check permissions
				continue
			}
			if !hasAccess {
				// Skip namespaces the user doesn't have access to
				continue
			}
		}
		// User has access (admin or has permission for this namespace)
		result = append(result, models.Namespace{
			Name:    ns.Name,
			Status:  string(ns.Status.Phase),
			Labels:  ns.Labels,
			Created: ns.CreationTimestamp.Format(time.RFC3339),
		})
	}
	return result, nil
}
