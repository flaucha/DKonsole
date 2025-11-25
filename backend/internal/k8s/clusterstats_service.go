package k8s

import (
	"context"
	"fmt"
	"log"

	"github.com/example/k8s-view/internal/models"
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
		log.Printf("Error fetching node count: %v", err)
	} else {
		stats.Nodes = count
	}

	if count, err := s.repo.GetNamespaceCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("namespaces: %w", err))
		log.Printf("Error fetching namespace count: %v", err)
	} else {
		stats.Namespaces = count
	}

	if count, err := s.repo.GetPodCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("pods: %w", err))
		log.Printf("Error fetching pod count: %v", err)
	} else {
		stats.Pods = count
	}

	if count, err := s.repo.GetDeploymentCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("deployments: %w", err))
		log.Printf("Error fetching deployment count: %v", err)
	} else {
		stats.Deployments = count
	}

	if count, err := s.repo.GetServiceCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("services: %w", err))
		log.Printf("Error fetching service count: %v", err)
	} else {
		stats.Services = count
	}

	if count, err := s.repo.GetIngressCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("ingresses: %w", err))
		log.Printf("Error fetching ingress count: %v", err)
	} else {
		stats.Ingresses = count
	}

	if count, err := s.repo.GetPVCCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("pvcs: %w", err))
		log.Printf("Error fetching PVC count: %v", err)
	} else {
		stats.PVCs = count
	}

	if count, err := s.repo.GetPVCount(ctx); err != nil {
		errors = append(errors, fmt.Errorf("pvs: %w", err))
		log.Printf("Error fetching PV count: %v", err)
	} else {
		stats.PVs = count
	}

	if len(errors) > 0 {
		return stats, fmt.Errorf("encountered errors while fetching some cluster stats: %v", errors)
	}

	return stats, nil
}
