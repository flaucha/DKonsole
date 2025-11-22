package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (h *Handlers) GetPrometheusPodMetrics(w http.ResponseWriter, r *http.Request) {
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

	// Query CPU metrics for specific pod
	cpuQuery := fmt.Sprintf(
		`sum(rate(container_cpu_usage_seconds_total{namespace="%s",pod="%s",container!=""}[5m])) * 1000`,
		namespace, podName,
	)

	// Query Memory metrics for specific pod
	memoryQuery := fmt.Sprintf(
		`sum(container_memory_working_set_bytes{namespace="%s",pod="%s",container!=""}) / 1024 / 1024`,
		namespace, podName,
	)

	cpuData := h.queryPrometheusRange(cpuQuery, startTime, endTime)
	memoryData := h.queryPrometheusRange(memoryQuery, startTime, endTime)

	response := DeploymentMetricsResponse{
		CPU:    cpuData,
		Memory: memoryData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
