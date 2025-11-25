package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/example/k8s-view/internal/models"
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
func (s *NamespaceService) GetNamespaces(ctx context.Context) ([]models.Namespace, error) {
	k8sNamespaces, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}
	var result []models.Namespace
	for _, ns := range k8sNamespaces {
		result = append(result, models.Namespace{
			Name:    ns.Name,
			Status:  string(ns.Status.Phase),
			Labels:  ns.Labels,
			Created: ns.CreationTimestamp.Format(time.RFC3339),
		})
	}
	return result, nil
}
