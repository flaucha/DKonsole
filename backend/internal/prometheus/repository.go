package prometheus

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// PrometheusQueryResult represents the response from Prometheus query API
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

// PrometheusInstantQueryResult represents the response from Prometheus instant query API
type PrometheusInstantQueryResult struct {
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

// Repository defines the interface for querying Prometheus
type Repository interface {
	QueryRange(query string, start, end time.Time, step string) ([]MetricDataPoint, error)
	QueryInstant(query string) ([]map[string]interface{}, error)
}

// HTTPPrometheusRepository implements Repository using HTTP client
type HTTPPrometheusRepository struct {
	baseURL string
	client  *http.Client
}

// NewHTTPPrometheusRepository creates a new HTTPPrometheusRepository
func NewHTTPPrometheusRepository(baseURL string) *HTTPPrometheusRepository {
	return &HTTPPrometheusRepository{
		baseURL: baseURL,
		client:  createSecureHTTPClient(),
	}
}

// QueryRange executes a Prometheus range query
func (r *HTTPPrometheusRepository) QueryRange(query string, start, end time.Time, step string) ([]MetricDataPoint, error) {
	if step == "" {
		step = "60s"
	}

	// Build Prometheus query URL
	promURL := fmt.Sprintf("%s/api/v1/query_range", r.baseURL)

	params := url.Values{}
	params.Add("query", query)
	params.Add("start", fmt.Sprintf("%d", start.Unix()))
	params.Add("end", fmt.Sprintf("%d", end.Unix()))
	params.Add("step", step)

	fullURL := fmt.Sprintf("%s?%s", promURL, params.Encode())

	resp, err := r.client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query Prometheus: %w", err)
	}
	defer resp.Body.Close()

	// Limit response size to 10MB to prevent DoS
	maxResponseSize := int64(10 << 20) // 10MB
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)

	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read Prometheus response: %w", err)
	}

	// Check if response was truncated
	if len(body) >= int(maxResponseSize) {
		fmt.Printf("Warning: Prometheus response truncated (max %d bytes)\n", maxResponseSize)
	}

	var result PrometheusQueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Prometheus response: %w", err)
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

	return dataPoints, nil
}

// QueryInstant executes a Prometheus instant query
func (r *HTTPPrometheusRepository) QueryInstant(query string) ([]map[string]interface{}, error) {
	// Build Prometheus query URL for instant query
	promURL := fmt.Sprintf("%s/api/v1/query", r.baseURL)

	params := url.Values{}
	params.Add("query", query)

	fullURL := fmt.Sprintf("%s?%s", promURL, params.Encode())

	resp, err := r.client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query Prometheus: %w", err)
	}
	defer resp.Body.Close()

	// Limit response size to 10MB to prevent DoS
	maxResponseSize := int64(10 << 20) // 10MB
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(limitedReader)
		return nil, fmt.Errorf("Prometheus query failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read Prometheus response: %w", err)
	}

	// Check if response was truncated
	if len(body) >= int(maxResponseSize) {
		fmt.Printf("Warning: Prometheus response truncated (max %d bytes)\n", maxResponseSize)
	}

	var result PrometheusInstantQueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Prometheus response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("Prometheus query error: %s (type: %s)", result.Error, result.ErrorType)
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

	return results, nil
}

// createSecureHTTPClient creates an HTTP client with proper TLS certificate validation
func createSecureHTTPClient() *http.Client {
	// Load system certificate pool
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

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



