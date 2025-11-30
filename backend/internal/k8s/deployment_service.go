package k8s

import (
	"context"
	"fmt"
	"time"
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

// RolloutDeployment triggers a rollout/restart of a deployment by updating an annotation
func (s *DeploymentService) RolloutDeployment(ctx context.Context, namespace, name string) error {
	// Get the deployment
	deployment, err := s.repo.GetDeployment(ctx, namespace, name)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	// Ensure annotations map exists
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}

	// Update the restart annotation to trigger a rollout
	// Using kubectl.kubernetes.io/restartAt annotation
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartAt"] = time.Now().Format(time.RFC3339)

	// Update the deployment
	_, err = s.repo.UpdateDeployment(ctx, namespace, deployment)
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	return nil
}
