package k8s

import (
	"context"
	"fmt"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// ClusterStatsService provides business logic for cluster statistics
type ClusterStatsService struct {
	repo ClusterStatsRepository
}

// NewClusterStatsService creates a new ClusterStatsService
func NewClusterStatsService(repo ClusterStatsRepository) *ClusterStatsService {
	return &ClusterStatsService{repo: repo}
}

// GetClusterStats fetches cluster statistics
func (s *ClusterStatsService) GetClusterStats(ctx context.Context) (models.ClusterStats, error) {
	stats := models.ClusterStats{}
	var errors []error

	if count, err := s.repo.GetNodeCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("nodes: %w", err))
		utils.LogError(err, "Error fetching node count", map[string]interface{}{"resource": "nodes"})
	} else {
		stats.Nodes = count
	}

	if count, err := s.repo.GetNamespaceCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("namespaces: %w", err))
		utils.LogError(err, "Error fetching namespace count", map[string]interface{}{"resource": "namespaces"})
	} else {
		stats.Namespaces = count
	}

	if count, err := s.repo.GetPodCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("pods: %w", err))
		utils.LogError(err, "Error fetching pod count", map[string]interface{}{"resource": "pods"})
	} else {
		stats.Pods = count
	}

	if count, err := s.repo.GetDeploymentCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("deployments: %w", err))
		utils.LogError(err, "Error fetching deployment count", map[string]interface{}{"resource": "deployments"})
	} else {
		stats.Deployments = count
	}

	if count, err := s.repo.GetServiceCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("services: %w", err))
		utils.LogError(err, "Error fetching service count", map[string]interface{}{"resource": "services"})
	} else {
		stats.Services = count
	}

	if count, err := s.repo.GetIngressCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("ingresses: %w", err))
		utils.LogError(err, "Error fetching ingress count", map[string]interface{}{"resource": "ingresses"})
	} else {
		stats.Ingresses = count
	}

	if count, err := s.repo.GetPVCCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("pvcs: %w", err))
		utils.LogError(err, "Error fetching PVC count", map[string]interface{}{"resource": "pvcs"})
	} else {
		stats.PVCs = count
	}

	if count, err := s.repo.GetPVCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("pvs: %w", err))
		utils.LogError(err, "Error fetching PV count", map[string]interface{}{"resource": "pvs"})
	} else {
		stats.PVs = count
	}

	if len(errors) > 0 {
		return stats, fmt.Errorf("encountered errors while fetching some cluster stats: %v", errors)
	}

	return stats, nil
}
