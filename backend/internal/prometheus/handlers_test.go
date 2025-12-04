package prometheus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

)

func newTestHandler(url string) *HTTPHandler {
	h := &HTTPHandler{
		prometheusURL: url,
		repo:          NewHTTPPrometheusRepository(url),
		promService:   NewService(NewHTTPPrometheusRepository(url)),
		clusterService: nil,
	}
	return h
}

func TestHTTPHandler_IsConfigured(t *testing.T) {
	handler := newTestHandler("")
	if handler.IsConfigured() {
		t.Fatalf("expected not configured when URL is empty")
	}

	handler.UpdateURL("http://example.com")
	if !handler.IsConfigured() {
		t.Fatalf("expected configured when URL is set")
	}
}

func TestHTTPHandler_HealthCheck(t *testing.T) {
	t.Run("not configured", func(t *testing.T) {
		h := newTestHandler("")
		if err := h.HealthCheck(context.Background()); err != nil {
			t.Fatalf("expected nil when not configured, got %v", err)
		}
	})

	t.Run("healthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		h := newTestHandler(server.URL)
		if err := h.HealthCheck(context.Background()); err != nil {
			t.Fatalf("expected healthy, got %v", err)
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		h := newTestHandler(server.URL)
		if err := h.HealthCheck(context.Background()); err == nil {
			t.Fatalf("expected error on unhealthy response")
		}
	})
}
