package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPPrometheusRepository_QueryRange(t *testing.T) {
	// Mock Prometheus server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check URL path
		if r.URL.Path != "/api/v1/query_range" {
			t.Errorf("expected path /api/v1/query_range, got %s", r.URL.Path)
		}

		// Check parameters
		q := r.URL.Query()
		if q.Get("query") == "" {
			t.Errorf("missing query parameter")
		}
		if q.Get("start") == "" {
			t.Errorf("missing start parameter")
		}
		if q.Get("end") == "" {
			t.Errorf("missing end parameter")
		}
		if q.Get("step") == "" {
			t.Errorf("missing step parameter")
		}

		// Return success response
		response := `{
			"status": "success",
			"data": {
				"resultType": "matrix",
				"result": [
					{
						"metric": {"foo": "bar"},
						"values": [
							[1620000000.123, "1.23"],
							[1620000060.456, "4.56"]
						]
					}
				]
			}
		}`
		fmt.Fprintln(w, response)
	}))
	defer ts.Close()

	repo := NewHTTPPrometheusRepository(ts.URL)
	ctx := context.Background()
	start := time.Unix(1620000000, 0)
	end := time.Unix(1620003600, 0)

	// Test success
	points, err := repo.QueryRange(ctx, "up", start, end, "60s")
	if err != nil {
		t.Fatalf("QueryRange failed: %v", err)
	}

	if len(points) != 2 {
		t.Errorf("expected 2 points, got %d", len(points))
	}

	if points[0].Value != 1.23 {
		t.Errorf("expected point 1 value 1.23, got %f", points[0].Value)
	}

	// Test server error
	tsError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer tsError.Close()

	repoError := NewHTTPPrometheusRepository(tsError.URL)
	_, err = repoError.QueryRange(ctx, "up", start, end, "60s")
	if err == nil {
		t.Error("expected error from server error, got nil")
	}

	// Test invalid JSON
	tsInvalid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "invalid json")
	}))
	defer tsInvalid.Close()

	repoInvalid := NewHTTPPrometheusRepository(tsInvalid.URL)
	_, err = repoInvalid.QueryRange(ctx, "up", start, end, "60s")
	if err == nil {
		t.Error("expected error from invalid JSON, got nil")
	}
}

func TestHTTPPrometheusRepository_QueryInstant(t *testing.T) {
	// Mock Prometheus server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/query" {
			t.Errorf("expected path /api/v1/query, got %s", r.URL.Path)
		}

		response := `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{
						"metric": {"foo": "bar"},
						"value": [1620000000.123, "1.23"]
					}
				]
			}
		}`
		fmt.Fprintln(w, response)
	}))
	defer ts.Close()

	repo := NewHTTPPrometheusRepository(ts.URL)
	ctx := context.Background()

	// Test success
	results, err := repo.QueryInstant(ctx, "up")
	if err != nil {
		t.Fatalf("QueryInstant failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	if val, ok := results[0]["value"].(float64); !ok || val != 1.23 {
		t.Errorf("expected value 1.23, got %v", results[0]["value"])
	}

	if val, ok := results[0]["foo"].(string); !ok || val != "bar" {
		t.Errorf("expected metric foo=bar, got %v", results[0]["foo"])
	}

	// Test Prometheus API error
	tsAPIError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"status": "error",
			"errorType": "bad_data",
			"error": "invalid query"
		}`
		fmt.Fprintln(w, response)
	}))
	defer tsAPIError.Close()

	repoAPIError := NewHTTPPrometheusRepository(tsAPIError.URL)
	_, err = repoAPIError.QueryInstant(ctx, "invalid_query")
	if err == nil {
		t.Error("expected error from API error, got nil")
	}
}
