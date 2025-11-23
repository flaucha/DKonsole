package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type PrometheusQueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

type MetricDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type DeploymentMetricsResponse struct {
	CPU    []MetricDataPoint `json:"cpu"`
	Memory []MetricDataPoint `json:"memory"`
}

func (h *Handlers) GetPrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	if h.PrometheusURL == "" {
		http.Error(w, "Prometheus URL not configured", http.StatusServiceUnavailable)
		return
	}

	deployment := r.URL.Query().Get("deployment")
	namespace := r.URL.Query().Get("namespace")
	rangeParam := r.URL.Query().Get("range") // e.g., "1h", "6h", "12h", "1d", "7d", "15d"

	if deployment == "" || namespace == "" {
		http.Error(w, "deployment and namespace are required", http.StatusBadRequest)
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

	validatedDeployment, err := validatePromQLParam(deployment, "deployment")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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

	cpuData := h.queryPrometheusRange(cpuQuery, startTime, endTime)
	memoryData := h.queryPrometheusRange(memoryQuery, startTime, endTime)

	response := DeploymentMetricsResponse{
		CPU:    cpuData,
		Memory: memoryData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// validatePromQLParam validates and escapes PromQL parameters to prevent injection
func validatePromQLParam(param, paramName string) (string, error) {
	// Validate that it only contains alphanumeric characters, hyphens, and dots
	// This covers Kubernetes namespaces, pod names, deployment names, etc.
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validPattern.MatchString(param) {
		return "", fmt.Errorf("invalid %s: contains invalid characters", paramName)
	}

	// Validate max length (Kubernetes limit is usually 253)
	if len(param) > 253 {
		return "", fmt.Errorf("invalid %s: too long", paramName)
	}

	// Escape double quotes just in case (though regex above prevents them)
	escaped := strings.ReplaceAll(param, `"`, `\"`)
	return escaped, nil
}

func (h *Handlers) queryPrometheusRange(query string, start, end time.Time) []MetricDataPoint {
	// Build Prometheus query URL
	promURL := fmt.Sprintf("%s/api/v1/query_range", h.PrometheusURL)

	params := url.Values{}
	params.Add("query", query)
	params.Add("start", fmt.Sprintf("%d", start.Unix()))
	params.Add("end", fmt.Sprintf("%d", end.Unix()))
	params.Add("step", "60s") // 1 minute resolution

	fullURL := fmt.Sprintf("%s?%s", promURL, params.Encode())

	client := createSecureHTTPClient()

	resp, err := client.Get(fullURL)
	if err != nil {
		return []MetricDataPoint{}
	}
	defer resp.Body.Close()

	// Limit response size to 10MB to prevent DoS
	maxResponseSize := int64(10 << 20) // 10MB
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)

	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return []MetricDataPoint{}
	}

	// Check if response was truncated
	if len(body) >= int(maxResponseSize) {
		fmt.Printf("Warning: Prometheus response truncated (max %d bytes)\n", maxResponseSize)
	}

	var result PrometheusQueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return []MetricDataPoint{}
	}

	var dataPoints []MetricDataPoint
	if len(result.Data.Result) > 0 {
		for _, value := range result.Data.Result[0].Values {
			if len(value) >= 2 {
				timestamp, ok1 := value[0].(float64)
				valueStr, ok2 := value[1].(string)

				if ok1 && ok2 {
					var floatValue float64
					fmt.Sscanf(valueStr, "%f", &floatValue)

					dataPoints = append(dataPoints, MetricDataPoint{
						Timestamp: int64(timestamp) * 1000, // Convert to milliseconds
						Value:     floatValue,
					})
				}
			}
		}
	}

	return dataPoints
}

func (h *Handlers) queryPrometheusInstant(query string) []map[string]interface{} {
	// Build Prometheus query URL for instant query
	promURL := fmt.Sprintf("%s/api/v1/query", h.PrometheusURL)

	params := url.Values{}
	params.Add("query", query)

	fullURL := fmt.Sprintf("%s?%s", promURL, params.Encode())

	client := createSecureHTTPClient()

	resp, err := client.Get(fullURL)
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		return []map[string]interface{}{}
	}
	defer resp.Body.Close()

	// Limit response size to 10MB to prevent DoS
	maxResponseSize := int64(10 << 20) // 10MB
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(limitedReader)
		fmt.Printf("Prometheus query failed with status %d: %s\n", resp.StatusCode, string(body))
		return []map[string]interface{}{}
	}

	body, err := io.ReadAll(limitedReader)
	if err != nil {
		fmt.Printf("Error reading Prometheus response: %v\n", err)
		return []map[string]interface{}{}
	}

	// Check if response was truncated
	if len(body) >= int(maxResponseSize) {
		fmt.Printf("Warning: Prometheus response truncated (max %d bytes)\n", maxResponseSize)
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		} `json:"data"`
		Error     string `json:"error"`
		ErrorType string `json:"errorType"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error parsing Prometheus response: %v\n", err)
		return []map[string]interface{}{}
	}

	if result.Status != "success" {
		fmt.Printf("Prometheus query error: %s (type: %s)\n", result.Error, result.ErrorType)
		return []map[string]interface{}{}
	}

	var results []map[string]interface{}
	for _, r := range result.Data.Result {
		resultMap := make(map[string]interface{})
		// Copy metric labels
		for k, v := range r.Metric {
			resultMap[k] = v
		}
		// Add value if present
		if len(r.Value) >= 2 {
			if valueStr, ok := r.Value[1].(string); ok {
				var floatValue float64
				fmt.Sscanf(valueStr, "%f", &floatValue)
				resultMap["value"] = floatValue
			}
		}
		results = append(results, resultMap)
	}

	return results
}

func (h *Handlers) GetPrometheusStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"enabled": h.PrometheusURL != "",
		"url":     h.PrometheusURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// createSecureHTTPClient creates an HTTP client with proper TLS certificate validation
func createSecureHTTPClient() *http.Client {
	// Load system certificate pool
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Optional: load additional certificates from environment variable
	// certPEM := os.Getenv("PROMETHEUS_CA_CERT")
	// if certPEM != "" {
	//     rootCAs.AppendCertsFromPEM([]byte(certPEM))
	// }

	config := &tls.Config{
		RootCAs: rootCAs,
		// In production, do not skip verification
		// Only allow skipping for development/testing
		InsecureSkipVerify: os.Getenv("PROMETHEUS_INSECURE_SKIP_VERIFY") == "true",
	}

	transport := &http.Transport{
		TLSClientConfig: config,
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}
