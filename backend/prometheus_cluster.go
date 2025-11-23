package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ClusterOverviewResponse includes cluster-wide metrics
type ClusterOverviewResponse struct {
	NodeMetrics  []NodeMetric  `json:"nodeMetrics"`
	ClusterStats *ClusterStats `json:"clusterStats"`
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

// ClusterStats represents aggregated cluster statistics
type ClusterStats struct {
	TotalNodes     int     `json:"totalNodes"`
	AvgCPUUsage    float64 `json:"avgCpuUsage"`
	AvgMemoryUsage float64 `json:"avgMemoryUsage"`
	NetworkTraffic float64 `json:"networkTraffic"`
	CPUTrend       float64 `json:"cpuTrend"`
	MemoryTrend    float64 `json:"memoryTrend"`
}

func (h *Handlers) GetPrometheusClusterOverview(w http.ResponseWriter, r *http.Request) {
	if h.PrometheusURL == "" {
		http.Error(w, "Prometheus URL not configured", http.StatusServiceUnavailable)
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

	// Query for node metrics
	nodeMetrics := h.getNodeMetrics(startTime, endTime)
	clusterStats := h.calculateClusterStats(nodeMetrics, startTime, endTime)

	response := ClusterOverviewResponse{
		NodeMetrics:  nodeMetrics,
		ClusterStats: clusterStats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) getNodeMetrics(startTime, endTime time.Time) []NodeMetric {
	var nodes []NodeMetric

	// Query to get list of nodes
	nodesQuery := `count by (node) (kube_node_info)`
	nodesData := h.queryPrometheusInstant(nodesQuery)

	for _, nodeData := range nodesData {
		nodeName := ""
		if node, ok := nodeData["node"].(string); ok {
			nodeName = node
		}

		if nodeName == "" {
			continue
		}

		// CPU usage per node (percentage)
		cpuQuery := fmt.Sprintf(
			`100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle",instance=~"%s.*"}[5m])) * 100)`,
			nodeName,
		)
		cpuData := h.queryPrometheusInstant(cpuQuery)
		cpuUsage := 0.0
		if len(cpuData) > 0 {
			if val, ok := cpuData[0]["value"].(float64); ok {
				cpuUsage = val
			}
		}

		// Memory usage per node (percentage)
		memQuery := fmt.Sprintf(
			`(1 - (node_memory_MemAvailable_bytes{instance=~"%s.*"} / node_memory_MemTotal_bytes{instance=~"%s.*"})) * 100`,
			nodeName, nodeName,
		)
		memData := h.queryPrometheusInstant(memQuery)
		memUsage := 0.0
		if len(memData) > 0 {
			if val, ok := memData[0]["value"].(float64); ok {
				memUsage = val
			}
		}

		// Disk usage per node (percentage)
		diskQuery := fmt.Sprintf(
			`(1 - (node_filesystem_avail_bytes{instance=~"%s.*",mountpoint="/"} / node_filesystem_size_bytes{instance=~"%s.*",mountpoint="/"})) * 100`,
			nodeName, nodeName,
		)
		diskData := h.queryPrometheusInstant(diskQuery)
		diskUsage := 0.0
		if len(diskData) > 0 {
			if val, ok := diskData[0]["value"].(float64); ok {
				diskUsage = val
			}
		}

		// Network RX (KB/s)
		netRxQuery := fmt.Sprintf(
			`sum(rate(node_network_receive_bytes_total{instance=~"%s.*"}[5m])) / 1024`,
			nodeName,
		)
		netRxData := h.queryPrometheusInstant(netRxQuery)
		networkRx := 0.0
		if len(netRxData) > 0 {
			if val, ok := netRxData[0]["value"].(float64); ok {
				networkRx = val
			}
		}

		// Network TX (KB/s)
		netTxQuery := fmt.Sprintf(
			`sum(rate(node_network_transmit_bytes_total{instance=~"%s.*"}[5m])) / 1024`,
			nodeName,
		)
		netTxData := h.queryPrometheusInstant(netTxQuery)
		networkTx := 0.0
		if len(netTxData) > 0 {
			if val, ok := netTxData[0]["value"].(float64); ok {
				networkTx = val
			}
		}

		// Node status
		statusQuery := fmt.Sprintf(`kube_node_status_condition{node="%s",condition="Ready",status="true"}`, nodeName)
		statusData := h.queryPrometheusInstant(statusQuery)
		status := "NotReady"
		if len(statusData) > 0 {
			status = "Ready"
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

func (h *Handlers) calculateClusterStats(nodes []NodeMetric, startTime, endTime time.Time) *ClusterStats {
	if len(nodes) == 0 {
		return &ClusterStats{
			TotalNodes: 0,
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
	cpuTrend := 0.0
	memTrend := 0.0

	// Query historical data for trend calculation
	cpuHistQuery := `avg(rate(node_cpu_seconds_total{mode!="idle"}[5m])) * 100`
	cpuHistData := h.queryPrometheusRange(cpuHistQuery, startTime.Add(-1*time.Hour), startTime)
	if len(cpuHistData) > 0 {
		cpuTrend = avgCPU - cpuHistData[0].Value
	}

	memHistQuery := `avg(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`
	memHistData := h.queryPrometheusRange(memHistQuery, startTime.Add(-1*time.Hour), startTime)
	if len(memHistData) > 0 {
		memTrend = avgMem - memHistData[0].Value
	}

	return &ClusterStats{
		TotalNodes:     len(nodes),
		AvgCPUUsage:    avgCPU,
		AvgMemoryUsage: avgMem,
		NetworkTraffic: networkTraffic,
		CPUTrend:       cpuTrend,
		MemoryTrend:    memTrend,
	}
}

// Helper function to query Prometheus for instant values
func (h *Handlers) queryPrometheusInstant(query string) []map[string]interface{} {
	// This is a simplified version - you'll need to implement the actual Prometheus instant query
	// For now, return empty slice
	return []map[string]interface{}{}
}
