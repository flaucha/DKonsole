package prometheus

import (
	"context"
	"fmt"

	"github.com/flaucha/DKonsole/backend/internal/models"

	"k8s.io/client-go/kubernetes"
)

// Service provides business logic for Prometheus metrics operations
type Service struct {
	repo Repository
}

// NewService creates a new Prometheus Service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GetDeploymentMetricsRequest represents parameters for getting deployment metrics
type GetDeploymentMetricsRequest struct {
	Deployment string
	Namespace  string
	Range      string
}

// GetDeploymentMetrics returns metrics for a specific deployment
func (s *Service) GetDeploymentMetrics(ctx context.Context, req GetDeploymentMetricsRequest) (*models.DeploymentMetricsResponse, error) {
	startTime, endTime := parseDuration(req.Range)

	validatedNamespace, err := validatePromQLParam(req.Namespace, "namespace")
	if err != nil {
		return nil, err
	}

	validatedDeployment, err := validatePromQLParam(req.Deployment, "deployment")
	if err != nil {
		return nil, err
	}

	// Query CPU metrics
	cpuQuery := fmt.Sprintf(
		`sum(rate(container_cpu_usage_seconds_total{namespace="%s",pod=~"%s-.*"}[5m])) * 1000`,
		validatedNamespace, validatedDeployment,
	)

	// Query Memory metrics
	memoryQuery := fmt.Sprintf(
		`sum(container_memory_working_set_bytes{namespace="%s",pod=~"%s-.*"}) / 1024 / 1024`,
		validatedNamespace, validatedDeployment,
	)

	cpuData, err := s.repo.QueryRange(ctx, cpuQuery, startTime, endTime, "60s")
	if err != nil {
		return nil, fmt.Errorf("failed to query CPU metrics: %w", err)
	}

	memoryData, err := s.repo.QueryRange(ctx, memoryQuery, startTime, endTime, "60s")
	if err != nil {
		return nil, fmt.Errorf("failed to query memory metrics: %w", err)
	}

	return &models.DeploymentMetricsResponse{
		CPU:    cpuData,
		Memory: memoryData,
	}, nil
}

// GetPodMetricsRequest represents parameters for getting pod metrics
type GetPodMetricsRequest struct {
	PodName   string
	Namespace string
	Range     string
}

// GetPodMetrics fetches all metrics for a specific pod
func (s *Service) GetPodMetrics(ctx context.Context, req GetPodMetricsRequest) (*models.PodMetricsResponse, error) {
	startTime, endTime := parseDuration(req.Range)

	// Validate and escape parameters
	validatedNamespace, err := validatePromQLParam(req.Namespace, "namespace")
	if err != nil {
		return nil, err
	}

	validatedPodName, err := validatePromQLParam(req.PodName, "pod")
	if err != nil {
		return nil, err
	}

	// Query CPU metrics for specific pod
	cpuQuery := fmt.Sprintf(
		`sum(rate(container_cpu_usage_seconds_total{namespace="%s",pod="%s",container!=""}[5m])) * 1000`,
		validatedNamespace, validatedPodName,
	)

	// Query Memory metrics for specific pod
	memoryQuery := fmt.Sprintf(
		`sum(container_memory_working_set_bytes{namespace="%s",pod="%s",container!=""}) / 1024 / 1024`,
		validatedNamespace, validatedPodName,
	)

	// Query Network RX (receive) metrics
	networkRxQuery := fmt.Sprintf(
		`sum(rate(container_network_receive_bytes_total{namespace="%s",pod="%s"}[5m])) / 1024`,
		validatedNamespace, validatedPodName,
	)

	// Query Network TX (transmit) metrics
	networkTxQuery := fmt.Sprintf(
		`sum(rate(container_network_transmit_bytes_total{namespace="%s",pod="%s"}[5m])) / 1024`,
		validatedNamespace, validatedPodName,
	)

	// Query PVC usage percentage
	pvcUsageQuery := fmt.Sprintf(
		`(sum(kubelet_volume_stats_used_bytes{namespace="%s",pod="%s"}) / sum(kubelet_volume_stats_capacity_bytes{namespace="%s",pod="%s"})) * 100`,
		validatedNamespace, validatedPodName, validatedNamespace, validatedPodName,
	)

	cpuData, err := s.repo.QueryRange(ctx, cpuQuery, startTime, endTime, "60s")
	if err != nil {
		return nil, fmt.Errorf("failed to query CPU metrics: %w", err)
	}

	memoryData, err := s.repo.QueryRange(ctx, memoryQuery, startTime, endTime, "60s")
	if err != nil {
		return nil, fmt.Errorf("failed to query memory metrics: %w", err)
	}

	networkRxData, err := s.repo.QueryRange(ctx, networkRxQuery, startTime, endTime, "60s")
	if err != nil {
		return nil, fmt.Errorf("failed to query network RX metrics: %w", err)
	}

	networkTxData, err := s.repo.QueryRange(ctx, networkTxQuery, startTime, endTime, "60s")
	if err != nil {
		return nil, fmt.Errorf("failed to query network TX metrics: %w", err)
	}

	pvcUsageData, err := s.repo.QueryRange(ctx, pvcUsageQuery, startTime, endTime, "60s")
	if err != nil {
		return nil, fmt.Errorf("failed to query PVC usage metrics: %w", err)
	}

	return &models.PodMetricsResponse{
		CPU:       cpuData,
		Memory:    memoryData,
		NetworkRx: networkRxData,
		NetworkTx: networkTxData,
		PVCUsage:  pvcUsageData,
	}, nil
}

// GetClusterOverviewRequest represents parameters for getting cluster overview metrics
type GetClusterOverviewRequest struct {
	Range string
}

// GetClusterOverview fetches cluster-wide metrics including node metrics and cluster stats
func (s *Service) GetClusterOverview(ctx context.Context, req GetClusterOverviewRequest, client kubernetes.Interface) (*models.ClusterOverviewResponse, error) {
	// Get node metrics
	nodeMetrics, controlPlaneCount, controlPlaneNodes, err := s.getNodeMetrics(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics: %w", err)
	}

	// Calculate cluster stats
	clusterStats := s.calculateClusterStats(nodeMetrics, controlPlaneCount, controlPlaneNodes)

	return &models.ClusterOverviewResponse{
		NodeMetrics:  nodeMetrics,
		ClusterStats: clusterStats,
	}, nil
}
