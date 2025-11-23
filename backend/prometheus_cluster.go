package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ClusterOverviewResponse includes cluster-wide metrics
type ClusterOverviewResponse struct {
	NodeMetrics  []NodeMetric            `json:"nodeMetrics"`
	ClusterStats *PrometheusClusterStats `json:"clusterStats"`
}

// NodeMetric represents metrics for a single node
type NodeMetric struct {
	Name      string  `json:"name"`
	CPUUsage  float64 `json:"cpuUsage"`
	MemUsage  float64 `json:"memoryUsage"`
	DiskUsage float64 `json:"diskUsage"`
	NetworkRx float64 `json:"networkRx"`
	NetworkTx float64 `json:"networkTx"`
	Status    string  `json:"status"`
}

// PrometheusClusterStats represents aggregated cluster statistics from Prometheus
type PrometheusClusterStats struct {
	TotalNodes     int     `json:"totalNodes"`
	AvgCPUUsage    float64 `json:"avgCpuUsage"`
	AvgMemoryUsage float64 `json:"avgMemoryUsage"`
	NetworkTraffic float64 `json:"networkTraffic"`
	CPUTrend       float64 `json:"cpuTrend"`
	MemoryTrend    float64 `json:"memoryTrend"`
}

func (h *Handlers) GetPrometheusClusterOverview(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GetPrometheusClusterOverview called, PrometheusURL: %s\n", h.PrometheusURL)
	
	if h.PrometheusURL == "" {
		fmt.Printf("Prometheus URL not configured\n")
		http.Error(w, "Prometheus URL not configured", http.StatusServiceUnavailable)
		return
	}

	// Get Kubernetes client
	client, err := h.getClient(r)
	if err != nil {
		fmt.Printf("Error getting client: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rangeParam := r.URL.Query().Get("range")
	duration := "1h"
	if rangeParam != "" {
		duration = rangeParam
	}

	// Calculate time range
	endTime := time.Now()
	var startTime time.Time

	switch duration {
	case "1h":
		startTime = endTime.Add(-1 * time.Hour)
	case "6h":
		startTime = endTime.Add(-6 * time.Hour)
	case "12h":
		startTime = endTime.Add(-12 * time.Hour)
	case "1d":
		startTime = endTime.Add(-24 * time.Hour)
	default:
		startTime = endTime.Add(-1 * time.Hour)
	}

	fmt.Printf("Fetching node metrics for range: %s\n", duration)
	// Query for node metrics
	nodeMetrics := h.getNodeMetrics(client, startTime, endTime)
	fmt.Printf("Found %d nodes with metrics\n", len(nodeMetrics))
	
	clusterStats := h.calculateClusterStats(nodeMetrics, startTime, endTime)
	fmt.Printf("Cluster stats: %+v\n", clusterStats)

	response := ClusterOverviewResponse{
		NodeMetrics:  nodeMetrics,
		ClusterStats: clusterStats,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) getNodeMetrics(client *kubernetes.Clientset, startTime, endTime time.Time) []NodeMetric {
	var nodes []NodeMetric

	if client == nil {
		fmt.Printf("Error: Kubernetes client is nil\n")
		return nodes
	}

	// Get nodes from Kubernetes API
	k8sNodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error getting nodes from Kubernetes: %v\n", err)
		return nodes
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

	// Query all CPU metrics - simplified query
	cpuQuery := `100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
	fmt.Printf("Querying CPU metrics: %s\n", cpuQuery)
	cpuData := h.queryPrometheusInstant(cpuQuery)
	fmt.Printf("CPU data returned: %d results\n", len(cpuData))

	// Query all memory metrics
	memQuery := `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`
	fmt.Printf("Querying memory metrics: %s\n", memQuery)
	memData := h.queryPrometheusInstant(memQuery)
	fmt.Printf("Memory data returned: %d results\n", len(memData))

	// Query all disk metrics
	diskQuery := `(1 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"})) * 100`
	diskData := h.queryPrometheusInstant(diskQuery)

	// Query all network RX metrics
	netRxQuery := `sum(rate(node_network_receive_bytes_total[5m])) by (instance) / 1024`
	netRxData := h.queryPrometheusInstant(netRxQuery)

	// Query all network TX metrics
	netTxQuery := `sum(rate(node_network_transmit_bytes_total[5m])) by (instance) / 1024`
	netTxData := h.queryPrometheusInstant(netTxQuery)

	// Build a map of instance to metrics for quick lookup
	cpuMap := make(map[string]float64)
	allInstances := make([]string, 0)
	for _, data := range cpuData {
		if inst, ok := data["instance"].(string); ok {
			if val, ok := data["value"].(float64); ok {
				cpuMap[inst] = val
				allInstances = append(allInstances, inst)
			}
		}
	}
	fmt.Printf("Available instances in CPU metrics: %v\n", allInstances)

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

	// Build a map of all available instances from CPU metrics
	availableInstances := make(map[string]bool)
	for _, data := range cpuData {
		if inst, ok := data["instance"].(string); ok {
			availableInstances[inst] = true
		}
	}
	fmt.Printf("Available instances in CPU metrics: %v\n", availableInstances)

	// Try to get instance mapping from kube_node_info (but it might not have the right instance)
	nodeToInstance := make(map[string]string)
	nodesQuery := `kube_node_info`
	fmt.Printf("Querying kube_node_info: %s\n", nodesQuery)
	nodesData := h.queryPrometheusInstant(nodesQuery)
	fmt.Printf("kube_node_info returned: %d results\n", len(nodesData))
	for _, nodeData := range nodesData {
		if node, ok := nodeData["node"].(string); ok {
			if inst, ok := nodeData["instance"].(string); ok {
				// Only use if this instance appears in our metrics
				if availableInstances[inst] {
					nodeToInstance[node] = inst
					fmt.Printf("Mapped node %s to instance %s (from kube_node_info)\n", node, inst)
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
			
			// Try to match instance by IP (instances are often IP:port)
			if nodeIP != "" {
				// Try exact IP match first (instance might be just IP)
				if availableInstances[nodeIP] {
					instance = nodeIP
					fmt.Printf("Mapped node %s (IP: %s) to instance %s by exact IP match\n", nodeName, nodeIP, instance)
				} else {
					// Try IP:port format (common for node_exporter)
					for inst := range availableInstances {
						// Extract IP from instance (format is usually IP:port)
						instParts := strings.Split(inst, ":")
						if len(instParts) > 0 && instParts[0] == nodeIP {
							instance = inst
							fmt.Printf("Mapped node %s (IP: %s) to instance %s by IP:port match\n", nodeName, nodeIP, instance)
							break
						}
						// Also try contains as fallback
						if strings.Contains(inst, nodeIP) && !strings.Contains(inst, "10.42.3.244:8080") {
							instance = inst
							fmt.Printf("Mapped node %s (IP: %s) to instance %s by IP contains match\n", nodeName, nodeIP, instance)
							break
						}
					}
				}
			}
			
			// Last resort: try node name as instance
			if instance == "" && availableInstances[nodeName] {
				instance = nodeName
				fmt.Printf("Mapped node %s to instance %s by name match\n", nodeName, instance)
			}
		}

		// Try to find metrics by instance
		cpuUsage := cpuMap[instance]
		if cpuUsage == 0 {
			// Try node name as fallback
			cpuUsage = cpuMap[nodeName]
		}
		if cpuUsage == 0 {
			fmt.Printf("Warning: No CPU metrics found for node %s (instance: %s)\n", nodeName, instance)
		}

		memUsage := memMap[instance]
		if memUsage == 0 {
			memUsage = memMap[nodeName]
		}
		if memUsage == 0 {
			fmt.Printf("Warning: No memory metrics found for node %s (instance: %s)\n", nodeName, instance)
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
		
		fmt.Printf("Node %s metrics - CPU: %.2f%%, Mem: %.2f%%, Disk: %.2f%%, NetRx: %.2f, NetTx: %.2f\n", 
			nodeName, cpuUsage, memUsage, diskUsage, networkRx, networkTx)

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

	return nodes
}

func (h *Handlers) calculateClusterStats(nodes []NodeMetric, startTime, endTime time.Time) *PrometheusClusterStats {
	if len(nodes) == 0 {
		return &PrometheusClusterStats{
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

	// Calculate trends (simplified - compare current vs 1 hour ago)
	// For now, we'll set trends to 0.0 as calculating actual trends would require
	// querying historical data which is more complex
	cpuTrend := 0.0
	memoryTrend := 0.0

	return &PrometheusClusterStats{
		TotalNodes:     len(nodes),
		AvgCPUUsage:    avgCPU,
		AvgMemoryUsage: avgMem,
		NetworkTraffic: networkTraffic,
		CPUTrend:       cpuTrend,
		MemoryTrend:    memoryTrend,
	}
}
