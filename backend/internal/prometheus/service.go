package prometheus

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Service provides business logic for Prometheus metrics operations
type Service struct {
	repo Repository
}

// NewService creates a new Prometheus Service
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetDeploymentMetricsRequest represents parameters for getting deployment metrics
type GetDeploymentMetricsRequest struct {
	Deployment string
	Namespace  string
	Range      string
}

// GetDeploymentMetrics fetches CPU and Memory metrics for a deployment
func (s *Service) GetDeploymentMetrics(ctx context.Context, req GetDeploymentMetricsRequest) (*DeploymentMetricsResponse, error) {
	startTime, endTime := parseDuration(req.Range)

	// Validate and escape parameters
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

	return &DeploymentMetricsResponse{
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
func (s *Service) GetPodMetrics(ctx context.Context, req GetPodMetricsRequest) (*PodMetricsResponse, error) {
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

	cpuData, _ := s.repo.QueryRange(ctx, cpuQuery, startTime, endTime, "60s")
	memoryData, _ := s.repo.QueryRange(ctx, memoryQuery, startTime, endTime, "60s")
	networkRxData, _ := s.repo.QueryRange(ctx, networkRxQuery, startTime, endTime, "60s")
	networkTxData, _ := s.repo.QueryRange(ctx, networkTxQuery, startTime, endTime, "60s")
	pvcUsageData, _ := s.repo.QueryRange(ctx, pvcUsageQuery, startTime, endTime, "60s")

	return &PodMetricsResponse{
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
func (s *Service) GetClusterOverview(ctx context.Context, req GetClusterOverviewRequest, client kubernetes.Interface) (*ClusterOverviewResponse, error) {
	// Get node metrics
	nodeMetrics, err := s.getNodeMetrics(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics: %w", err)
	}

	// Calculate cluster stats
	clusterStats := s.calculateClusterStats(nodeMetrics)

	return &ClusterOverviewResponse{
		NodeMetrics:  nodeMetrics,
		ClusterStats: clusterStats,
	}, nil
}

// getNodeMetrics fetches metrics for all nodes
func (s *Service) getNodeMetrics(ctx context.Context, client kubernetes.Interface) ([]NodeMetric, error) {
	var nodes []NodeMetric

	if client == nil {
		return nodes, fmt.Errorf("kubernetes client is nil")
	}

	// Get nodes from Kubernetes API
	k8sNodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nodes, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Get node status from Kubernetes
	nodeStatusMap := make(map[string]string)
	for _, node := range k8sNodes.Items {
		status := "NotReady"
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				status = "Ready"
				break
			}
		}
		nodeStatusMap[node.Name] = status
	}

	// Query all CPU metrics
	cpuQuery := `100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
	cpuData, _ := s.repo.QueryInstant(ctx, cpuQuery)

	// Query all memory metrics
	memQuery := `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`
	memData, _ := s.repo.QueryInstant(ctx, memQuery)

	// Query all disk metrics
	diskQuery := `(1 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"})) * 100`
	diskData, _ := s.repo.QueryInstant(ctx, diskQuery)

	// Query all network RX metrics
	netRxQuery := `sum(rate(node_network_receive_bytes_total[5m])) by (instance) / 1024`
	netRxData, _ := s.repo.QueryInstant(ctx, netRxQuery)

	// Query all network TX metrics
	netTxQuery := `sum(rate(node_network_transmit_bytes_total[5m])) by (instance) / 1024`
	netTxData, _ := s.repo.QueryInstant(ctx, netTxQuery)

	// Build maps for quick lookup
	cpuMap := make(map[string]float64)
	for _, data := range cpuData {
		if inst, ok := data["instance"].(string); ok {
			if val, ok := data["value"].(float64); ok {
				cpuMap[inst] = val
			}
		}
	}

	memMap := make(map[string]float64)
	for _, data := range memData {
		if inst, ok := data["instance"].(string); ok {
			if val, ok := data["value"].(float64); ok {
				memMap[inst] = val
			}
		}
	}

	diskMap := make(map[string]float64)
	for _, data := range diskData {
		if inst, ok := data["instance"].(string); ok {
			if val, ok := data["value"].(float64); ok {
				diskMap[inst] = val
			}
		}
	}

	netRxMap := make(map[string]float64)
	for _, data := range netRxData {
		if inst, ok := data["instance"].(string); ok {
			if val, ok := data["value"].(float64); ok {
				netRxMap[inst] = val
			}
		}
	}

	netTxMap := make(map[string]float64)
	for _, data := range netTxData {
		if inst, ok := data["instance"].(string); ok {
			if val, ok := data["value"].(float64); ok {
				netTxMap[inst] = val
			}
		}
	}

	// Build a map of all available instances
	availableInstances := make(map[string]bool)
	for _, data := range cpuData {
		if inst, ok := data["instance"].(string); ok {
			availableInstances[inst] = true
		}
	}

	// Try to get instance mapping from kube_node_info
	nodeToInstance := make(map[string]string)
	nodesQuery := `kube_node_info`
	nodesData, _ := s.repo.QueryInstant(ctx, nodesQuery)
	for _, nodeData := range nodesData {
		if node, ok := nodeData["node"].(string); ok {
			if inst, ok := nodeData["instance"].(string); ok {
				if availableInstances[inst] {
					nodeToInstance[node] = inst
				}
			}
		}
	}

	// Process each Kubernetes node
	for _, k8sNode := range k8sNodes.Items {
		nodeName := k8sNode.Name
		instance := nodeToInstance[nodeName]

		// If we don't have a mapping, try to find by IP address
		if instance == "" {
			var nodeIP string
			for _, addr := range k8sNode.Status.Addresses {
				if addr.Type == corev1.NodeInternalIP {
					nodeIP = addr.Address
					break
				}
			}

			// Try to match instance by IP
			if nodeIP != "" {
				if availableInstances[nodeIP] {
					instance = nodeIP
				} else {
					// Try IP:port format
					for inst := range availableInstances {
						instParts := strings.Split(inst, ":")
						if len(instParts) > 0 && instParts[0] == nodeIP {
							instance = inst
							break
						}
						if strings.Contains(inst, nodeIP) {
							instance = inst
							break
						}
					}
				}
			}

			// Last resort: try node name as instance
			if instance == "" && availableInstances[nodeName] {
				instance = nodeName
			}
		}

		// Get metrics by instance (with fallback to node name)
		cpuUsage := cpuMap[instance]
		if cpuUsage == 0 {
			cpuUsage = cpuMap[nodeName]
		}

		memUsage := memMap[instance]
		if memUsage == 0 {
			memUsage = memMap[nodeName]
		}

		diskUsage := diskMap[instance]
		if diskUsage == 0 {
			diskUsage = diskMap[nodeName]
		}

		networkRx := netRxMap[instance]
		if networkRx == 0 {
			networkRx = netRxMap[nodeName]
		}

		networkTx := netTxMap[instance]
		if networkTx == 0 {
			networkTx = netTxMap[nodeName]
		}

		status := nodeStatusMap[nodeName]
		if status == "" {
			status = "NotReady"
		}

		nodes = append(nodes, NodeMetric{
			Name:      nodeName,
			CPUUsage:  cpuUsage,
			MemUsage:  memUsage,
			DiskUsage: diskUsage,
			NetworkRx: networkRx,
			NetworkTx: networkTx,
			Status:    status,
		})
	}

	return nodes, nil
}

// calculateClusterStats calculates aggregated cluster statistics
func (s *Service) calculateClusterStats(nodes []NodeMetric) *ClusterStats {
	if len(nodes) == 0 {
		return &ClusterStats{
			TotalNodes:     0,
			AvgCPUUsage:    0.0,
			AvgMemoryUsage: 0.0,
			NetworkTraffic: 0.0,
			CPUTrend:       0.0,
			MemoryTrend:    0.0,
		}
	}

	totalCPU := 0.0
	totalMem := 0.0
	totalNetworkRx := 0.0
	totalNetworkTx := 0.0

	for _, node := range nodes {
		totalCPU += node.CPUUsage
		totalMem += node.MemUsage
		totalNetworkRx += node.NetworkRx
		totalNetworkTx += node.NetworkTx
	}

	avgCPU := totalCPU / float64(len(nodes))
	avgMem := totalMem / float64(len(nodes))
	networkTraffic := (totalNetworkRx + totalNetworkTx) / 1024 // Convert to MB/s

	// Calculate trends (simplified - for now set to 0.0)
	cpuTrend := 0.0
	memoryTrend := 0.0

	return &ClusterStats{
		TotalNodes:     len(nodes),
		AvgCPUUsage:    avgCPU,
		AvgMemoryUsage: avgMem,
		NetworkTraffic: networkTraffic,
		CPUTrend:       cpuTrend,
		MemoryTrend:    memoryTrend,
	}
}
