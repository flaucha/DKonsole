package prometheus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/models"
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

type mockRepo struct {
	queryRangeErr   error
	queryInstantErr error
}

func (m *mockRepo) QueryRange(ctx context.Context, query string, start, end time.Time, step string) ([]MetricDataPoint, error) {
	if m.queryRangeErr != nil {
		return nil, m.queryRangeErr
	}
	return []MetricDataPoint{{Timestamp: start.Unix(), Value: 1.23}}, nil
}

func (m *mockRepo) QueryInstant(ctx context.Context, query string) ([]map[string]interface{}, error) {
	if m.queryInstantErr != nil {
		return nil, m.queryInstantErr
	}
	return []map[string]interface{}{{"status": "up"}}, nil
}

func TestHTTPHandler_GetMetrics(t *testing.T) {
	handler := &HTTPHandler{
		prometheusURL: "http://prom",
		repo:          &mockRepo{},
		promService:   NewService(&mockRepo{}),
	}

	t.Run("missing params", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/prometheus/metrics", nil)
		handler.GetMetrics(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rr.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/prometheus/metrics?deployment=demo&namespace=default", nil)
		handler.GetMetrics(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", rr.Code, rr.Body.String())
		}
	})
}

func TestHTTPHandler_GetClusterOverview(t *testing.T) {
	client := k8sfake.NewSimpleClientset(&corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "worker-1",
		},
	})
	handlersModel := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": client},
	}
	clusterSvc := cluster.NewService(handlersModel)

	handler := &HTTPHandler{
		prometheusURL:  "http://prom",
		repo:           &mockRepo{},
		promService:    NewService(&mockRepo{}),
		clusterService: clusterSvc,
	}

	t.Run("not configured", func(t *testing.T) {
		handler.prometheusURL = ""
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/prometheus/cluster-overview", nil)
		handler.GetClusterOverview(rr, req)
		if rr.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", rr.Code)
		}
		handler.prometheusURL = "http://prom"
	})

	t.Run("success", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/prometheus/cluster-overview", nil)
		handler.GetClusterOverview(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", rr.Code, rr.Body.String())
		}
	})
}
