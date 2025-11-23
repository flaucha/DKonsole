package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// PodMetricsResponse includes all pod metrics
type PodMetricsResponse struct {
	CPU       []MetricDataPoint `json:"cpu"`
	Memory    []MetricDataPoint `json:"memory"`
	NetworkRx []MetricDataPoint `json:"networkRx"`
	NetworkTx []MetricDataPoint `json:"networkTx"`
	PVCUsage  []MetricDataPoint `json:"pvcUsage"`
}

func (h *Handlers) GetPrometheusPodMetrics(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetPrometheusPodMetrics: PrometheusURL=%s, pod=%s, namespace=%s", 
		h.PrometheusURL, r.URL.Query().Get("pod"), r.URL.Query().Get("namespace"))
	if h.PrometheusURL == "" {
		http.Error(w, "Prometheus URL not configured", http.StatusServiceUnavailable)
		return
	}

	podName := r.URL.Query().Get("pod")
	namespace := r.URL.Query().Get("namespace")
	rangeParam := r.URL.Query().Get("range") // e.g., "1h", "6h", "12h", "1d", "7d", "15d"

	if podName == "" || namespace == "" {
		http.Error(w, "pod and namespace are required", http.StatusBadRequest)
		return
	}

	// Parse range parameter
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
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "15d":
		startTime = endTime.Add(-15 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-1 * time.Hour)
	}

	// Validate and escape parameters
	validatedNamespace, err := validatePromQLParam(namespace, "namespace")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validatedPodName, err := validatePromQLParam(podName, "pod")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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
	// This query calculates the percentage of used space in PVCs mounted by the pod
	pvcUsageQuery := fmt.Sprintf(
		`(sum(kubelet_volume_stats_used_bytes{namespace="%s",pod="%s"}) / sum(kubelet_volume_stats_capacity_bytes{namespace="%s",pod="%s"})) * 100`,
		validatedNamespace, validatedPodName, validatedNamespace, validatedPodName,
	)

	cpuData := h.queryPrometheusRange(cpuQuery, startTime, endTime)
	memoryData := h.queryPrometheusRange(memoryQuery, startTime, endTime)
	networkRxData := h.queryPrometheusRange(networkRxQuery, startTime, endTime)
	networkTxData := h.queryPrometheusRange(networkTxQuery, startTime, endTime)
	pvcUsageData := h.queryPrometheusRange(pvcUsageQuery, startTime, endTime)

	response := PodMetricsResponse{
		CPU:       cpuData,
		Memory:    memoryData,
		NetworkRx: networkRxData,
		NetworkTx: networkTxData,
		PVCUsage:  pvcUsageData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
