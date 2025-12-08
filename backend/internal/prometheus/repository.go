package prometheus

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"time"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
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
	QueryRange(ctx context.Context, query string, start, end time.Time, step string) ([]models.MetricDataPoint, error)
	QueryInstant(ctx context.Context, query string) ([]map[string]interface{}, error)
}

// HTTPPrometheusRepository implements Repository using HTTP client
type HTTPPrometheusRepository struct {
	baseURL string
	client  *http.Client
}

// NewHTTPPrometheusRepository creates a new HTTPPrometheusRepository
func NewHTTPPrometheusRepository(baseURL string) *HTTPPrometheusRepository {
	timeout := getPrometheusTimeout()
	return &HTTPPrometheusRepository{
		baseURL: baseURL,
		client:  createSecureHTTPClient(timeout),
	}
}

// getPrometheusTimeout returns the timeout for Prometheus queries from environment variable
// Default: 30 seconds
func getPrometheusTimeout() time.Duration {
	timeoutStr := os.Getenv("PROMETHEUS_QUERY_TIMEOUT")
	if timeoutStr == "" {
		return 30 * time.Second
	}
	if timeout, err := time.ParseDuration(timeoutStr); err == nil {
		return timeout
	}
	return 30 * time.Second
}

// QueryRange executes a Prometheus range query with context timeout
func (r *HTTPPrometheusRepository) QueryRange(ctx context.Context, query string, start, end time.Time, step string) ([]models.MetricDataPoint, error) {
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

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add tracing for debugging
	trace := &httptrace.ClientTrace{}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	resp, err := r.client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("prometheus query timeout: %w", err)
		}
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
		utils.LogWarn("Prometheus response truncated", map[string]interface{}{
			"max_bytes": maxResponseSize,
		})
	}

	var result PrometheusQueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Prometheus response: %w", err)
	}

	var dataPoints []models.MetricDataPoint
	if len(result.Data.Result) > 0 {
		for _, value := range result.Data.Result[0].Values {
			if len(value) >= 2 {
				timestamp, ok1 := value[0].(float64)
				valueStr, ok2 := value[1].(string)

				if ok1 && ok2 {
					var floatValue float64
					fmt.Sscanf(valueStr, "%f", &floatValue)

					dataPoints = append(dataPoints, models.MetricDataPoint{
						Timestamp: int64(timestamp) * 1000, // Convert to milliseconds
						Value:     floatValue,
					})
				}
			}
		}
	}

	return dataPoints, nil
}

// QueryInstant executes a Prometheus instant query with context timeout
func (r *HTTPPrometheusRepository) QueryInstant(ctx context.Context, query string) ([]map[string]interface{}, error) {
	// Build Prometheus query URL for instant query
	promURL := fmt.Sprintf("%s/api/v1/query", r.baseURL)

	params := url.Values{}
	params.Add("query", query)

	fullURL := fmt.Sprintf("%s?%s", promURL, params.Encode())

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("prometheus query timeout: %w", err)
		}
		return nil, fmt.Errorf("failed to query Prometheus: %w", err)
	}
	defer resp.Body.Close()

	// Limit response size to 10MB to prevent DoS
	maxResponseSize := int64(10 << 20) // 10MB
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(limitedReader)
		return nil, fmt.Errorf("prometheus query failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read Prometheus response: %w", err)
	}

	// Check if response was truncated
	if len(body) >= int(maxResponseSize) {
		utils.LogWarn("Prometheus response truncated", map[string]interface{}{
			"max_bytes": maxResponseSize,
		})
	}

	var result PrometheusInstantQueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Prometheus response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("prometheus query error: %s (type: %s)", result.Error, result.ErrorType)
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
// The timeout parameter allows customizing the client timeout
func createSecureHTTPClient(timeout time.Duration) *http.Client {
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
		Timeout:   timeout,
		Transport: transport,
	}
}
