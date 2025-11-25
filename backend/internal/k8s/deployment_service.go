package k8s

import (
	"context"
	"fmt"
)

// DeploymentService provides business logic for Deployment operations
type DeploymentService struct {
	repo DeploymentRepository
}

// NewDeploymentService creates a new DeploymentService
func NewDeploymentService(repo DeploymentRepository) *DeploymentService {
	return &DeploymentService{repo: repo}
}

// ScaleDeployment scales a deployment by a given delta
func (s *DeploymentService) ScaleDeployment(ctx context.Context, namespace, name string, delta int) (int32, error) {
	// Get current scale
	scale, err := s.repo.GetScale(ctx, namespace, name)
	if err != nil {
		return 0, fmt.Errorf("failed to get current scale: %w", err)
	}

	// Calculate new replicas
	currentReplicas := int(scale.Spec.Replicas)
	newReplicas := currentReplicas + delta
	if newReplicas < 0 {
		newReplicas = 0
	}

	// Update only the replicas, preserving all metadata (including ResourceVersion)
	scale.Spec.Replicas = int32(newReplicas)

	// Update scale
	updatedScale, err := s.repo.UpdateScale(ctx, namespace, name, scale)
	if err != nil {
		return 0, fmt.Errorf("failed to update scale: %w", err)
	}

	return updatedScale.Spec.Replicas, nil
}
