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
	if client == nil {
		return nil, 0, nil, fmt.Errorf("kubernetes client is nil")
	}

	// 1. Get nodes from Kubernetes API
	k8sNodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// 2. Identify control plane nodes
	controlPlaneNodes := make(map[string]bool)
	controlPlaneCount := 0
	for _, node := range k8sNodes.Items {
		if isControlPlaneNode(node) {
			controlPlaneCount++
			controlPlaneNodes[node.Name] = true
		}
	}

	// 3. Get node status
	nodeStatusMap := getNodeStatusMap(k8sNodes.Items)

	// 4. Fetch Prometheus metrics
	cpuMap, err := s.fetchMetricMap(ctx, `100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`)
	if err != nil {
		// Log error but continue with partial data, or return error?
		// The analysis criticized silent failure. We should at least return strict errors if critical,
		// but providing partial metrics with zero values is often better for UI than total failure.
		// However, the "Reliability" point 2.4 demanded we stop ignoring errors.
		// A compromise: We log the error internally (if we had a logger here) and return the error
		// ONLY if it's a connection failure, but here we will return the error to be strict as requested.
		return nil, 0, nil, fmt.Errorf("failed to query CPU metrics: %w", err)
	}

	memMap, err := s.fetchMetricMap(ctx, `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to query Memory metrics: %w", err)
	}

	diskMap, err := s.fetchMetricMap(ctx, `(1 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"})) * 100`)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to query Disk metrics: %w", err)
	}

	netRxMap, err := s.fetchMetricMap(ctx, `sum(rate(node_network_receive_bytes_total[5m])) by (instance) / 1024`)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to query Network RX metrics: %w", err)
	}

	netTxMap, err := s.fetchMetricMap(ctx, `sum(rate(node_network_transmit_bytes_total[5m])) by (instance) / 1024`)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to query Network TX metrics: %w", err)
	}

	// 5. Build instance mapping
	availableInstances := make(map[string]bool)
	for inst := range cpuMap {
		availableInstances[inst] = true
	}
	nodeToInstance := s.buildNodeToInstanceMap(ctx, availableInstances)

	// 6. Map metrics to nodes
	var nodes []models.NodeMetric
	for _, k8sNode := range k8sNodes.Items {
		nodeName := k8sNode.Name
		instance := resolveNodeInstance(nodeName, k8sNode, nodeToInstance, availableInstances)

		// Determine node role
		role := "worker"
		if controlPlaneNodes[nodeName] {
			role = "control-plane"
		}

		nodes = append(nodes, models.NodeMetric{
			Name:      nodeName,
			Role:      role,
			CPUUsage:  getMetricValue(instance, nodeName, cpuMap),
			MemUsage:  getMetricValue(instance, nodeName, memMap),
			DiskUsage: getMetricValue(instance, nodeName, diskMap),
			NetworkRx: getMetricValue(instance, nodeName, netRxMap),
			NetworkTx: getMetricValue(instance, nodeName, netTxMap),
			Status:    nodeStatusMap[nodeName],
		})
	}

	return nodes, controlPlaneCount, controlPlaneNodes, nil
}

// Helper: Fetch and map metric values
func (s *Service) fetchMetricMap(ctx context.Context, query string) (map[string]float64, error) {
	data, err := s.repo.QueryInstant(ctx, query)
	if err != nil {
		return nil, err
	}

	result := make(map[string]float64)
	for _, d := range data {
		inst, okInst := d["instance"].(string)
		val, okVal := d["value"].(float64)
		if okInst && okVal {
			result[inst] = val
		}
	}
	return result, nil
}

// Helper: Build node status map
func getNodeStatusMap(nodes []corev1.Node) map[string]string {
	statusMap := make(map[string]string)
	for _, node := range nodes {
		status := "NotReady"
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				status = "Ready"
				break
			}
		}
		statusMap[node.Name] = status
	}
	return statusMap
}

// Helper: Build map from K8s node name to Prometheus instance
func (s *Service) buildNodeToInstanceMap(ctx context.Context, availableInstances map[string]bool) map[string]string {
	nodeToInstance := make(map[string]string)
	nodesQuery := `kube_node_info`
	nodesData, err := s.repo.QueryInstant(ctx, nodesQuery)
	if err != nil {
		return nodeToInstance // Return empty map if query fails, fallback logic will handle it
	}

	for _, nodeData := range nodesData {
		node, okNode := nodeData["node"].(string)
		inst, okInst := nodeData["instance"].(string)
		if okNode && okInst && availableInstances[inst] {
			nodeToInstance[node] = inst
		}
	}
	return nodeToInstance
}

// Helper: Resolve instance string for a node
func resolveNodeInstance(nodeName string, node corev1.Node, nodeToInstance map[string]string, availableInstances map[string]bool) string {
	if inst, ok := nodeToInstance[nodeName]; ok {
		return inst
	}

	// Fallback 1: Internal IP
	var nodeIP string
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			nodeIP = addr.Address
			break
		}
	}

	if nodeIP != "" {
		if availableInstances[nodeIP] {
			return nodeIP
		}
		// Fallback 2: IP:port matching
		for inst := range availableInstances {
			if strings.HasPrefix(inst, nodeIP+":") {
				return inst
			}
		}
	}

	// Fallback 3: Node Name
	if availableInstances[nodeName] {
		return nodeName
	}

	return ""
}

// Helper: Safe metric retrieval
func getMetricValue(primaryKey, secondaryKey string, metricMap map[string]float64) float64 {
	if val, ok := metricMap[primaryKey]; ok {
		return val
	}
	if val, ok := metricMap[secondaryKey]; ok {
		return val
	}
	return 0.0
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
	var workerNodes []models.NodeMetric
	for _, node := range nodes {
		// Only include worker nodes in stats calculation
		if !controlPlaneNodes[node.Name] {
			workerNodes = append(workerNodes, node)
		}
	}

	// Count worker nodes
	workerNodeCount := len(workerNodes)

	if workerNodeCount == 0 {
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
	avgCPU := totalCPU / float64(workerNodeCount)
	avgMem := totalMem / float64(workerNodeCount)
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
