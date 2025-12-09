package prometheus

import (
	"context"
	"fmt"
	"strings"

	"github.com/flaucha/DKonsole/backend/internal/models"
	
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getNodeMetrics fetches metrics for all nodes
func (s *Service) getNodeMetrics(ctx context.Context, client kubernetes.Interface) ([]models.NodeMetric, int, map[string]bool, error) {
	var nodes []models.NodeMetric
	controlPlaneCount := 0
	controlPlaneNodes := make(map[string]bool)

	if client == nil {
		return nodes, 0, controlPlaneNodes, fmt.Errorf("kubernetes client is nil")
	}

	// Get nodes from Kubernetes API
	k8sNodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nodes, 0, controlPlaneNodes, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Count and mark control plane nodes
	for _, node := range k8sNodes.Items {
		if isControlPlaneNode(node) {
			controlPlaneCount++
			controlPlaneNodes[node.Name] = true
		}
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

	// Process each Kubernetes node (including control plane nodes)
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

		// Determine node role
		role := "worker"
		if controlPlaneNodes[nodeName] {
			role = "control-plane"
		}

		nodes = append(nodes, models.NodeMetric{
			Name:      nodeName,
			Role:      role,
			CPUUsage:  cpuUsage,
			MemUsage:  memUsage,
			DiskUsage: diskUsage,
			NetworkRx: networkRx,
			NetworkTx: networkTx,
			Status:    status,
		})
	}

	return nodes, controlPlaneCount, controlPlaneNodes, nil
}

// isControlPlaneNode checks if a node is a control plane/master node
func isControlPlaneNode(node corev1.Node) bool {
	// Check labels
	if val, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok && val != "" {
		return true
	}
	if val, ok := node.Labels["node-role.kubernetes.io/master"]; ok && val != "" {
		return true
	}

	// Check taints
	for _, taint := range node.Spec.Taints {
		if taint.Key == "node-role.kubernetes.io/control-plane" || taint.Key == "node-role.kubernetes.io/master" {
			return true
		}
	}

	return false
}

// calculateClusterStats calculates aggregated cluster statistics
func (s *Service) calculateClusterStats(nodes []models.NodeMetric, controlPlaneCount int, controlPlaneNodes map[string]bool) *models.PrometheusClusterStats {
	// Separate worker nodes from control plane nodes for stats calculation
	workerNodes := []models.NodeMetric{}
	for _, node := range nodes {
		// Only include worker nodes in stats calculation
		if !controlPlaneNodes[node.Name] {
			workerNodes = append(workerNodes, node)
		}
	}

	// Count worker nodes
	workerNodeCount := len(workerNodes)

	if len(workerNodes) == 0 {
		return &models.PrometheusClusterStats{
			TotalNodes:        0,
			ControlPlaneNodes: controlPlaneCount,
			AvgCPUUsage:       0.0,
			AvgMemoryUsage:    0.0,
			NetworkTraffic:    0.0,
			CPUTrend:          0.0,
			MemoryTrend:       0.0,
		}
	}

	totalCPU := 0.0
	totalMem := 0.0
	totalNetworkRx := 0.0
	totalNetworkTx := 0.0

	// Calculate averages for worker nodes only (excluding control plane from averages)
	for _, node := range workerNodes {
		totalCPU += node.CPUUsage
		totalMem += node.MemUsage
		totalNetworkRx += node.NetworkRx
		totalNetworkTx += node.NetworkTx
	}

	// Use worker node count for averages
	avgCPU := totalCPU / float64(len(workerNodes))
	avgMem := totalMem / float64(len(workerNodes))
	networkTraffic := (totalNetworkRx + totalNetworkTx) / 1024 // Convert to MB/s

	// Calculate trends (simplified - for now set to 0.0)
	cpuTrend := 0.0
	memoryTrend := 0.0

	return &models.PrometheusClusterStats{
		TotalNodes:        workerNodeCount,
		ControlPlaneNodes: controlPlaneCount,
		AvgCPUUsage:       avgCPU,
		AvgMemoryUsage:    avgMem,
		NetworkTraffic:    networkTraffic,
		CPUTrend:          cpuTrend,
		MemoryTrend:       memoryTrend,
	}
}
