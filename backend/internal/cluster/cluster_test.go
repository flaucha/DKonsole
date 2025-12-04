package cluster

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func newRequest(cluster string) *http.Request {
	return httptest.NewRequest("GET", "/?cluster="+cluster, nil)
}

func TestService_GetClient_DefaultAndCustomCluster(t *testing.T) {
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{
			"default": k8sfake.NewSimpleClientset(),
			"other":   k8sfake.NewSimpleClientset(),
		},
	}
	service := NewService(handlers)

	req := newRequest("")
	client, err := service.GetClient(req)
	if err != nil {
		t.Fatalf("GetClient returned error: %v", err)
	}
	if client != handlers.Clients["default"] {
		t.Fatalf("expected default client")
	}

	req = newRequest("other")
	client, err = service.GetClient(req)
	if err != nil {
		t.Fatalf("GetClient returned error: %v", err)
	}
	if client != handlers.Clients["other"] {
		t.Fatalf("expected other client")
	}
}

func TestService_GetClient_NotFound(t *testing.T) {
	service := NewService(&models.Handlers{Clients: map[string]kubernetes.Interface{}})

	if _, err := service.GetClient(newRequest("missing")); err == nil {
		t.Fatalf("expected error for missing cluster")
	}
}

func TestService_GetDynamicClient(t *testing.T) {
	dynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	service := NewService(&models.Handlers{
		Dynamics: map[string]dynamic.Interface{
			"default": dynamicClient,
		},
	})

	client, err := service.GetDynamicClient(newRequest(""))
	if err != nil {
		t.Fatalf("GetDynamicClient returned error: %v", err)
	}
	if client != dynamicClient {
		t.Fatalf("expected dynamic client")
	}
	if _, err := service.GetDynamicClient(newRequest("missing")); err == nil {
		t.Fatalf("expected error for missing cluster")
	}
}

func TestService_GetMetricsClient(t *testing.T) {
	metricsClient := &metricsv.Clientset{}
	service := NewService(&models.Handlers{
		Metrics: map[string]*metricsv.Clientset{
			"default": metricsClient,
		},
	})

	if got := service.GetMetricsClient(newRequest("")); got != metricsClient {
		t.Fatalf("expected default metrics client")
	}
	if got := service.GetMetricsClient(newRequest("missing")); got != nil {
		t.Fatalf("expected nil for missing cluster")
	}
}

func TestService_GetRESTConfig(t *testing.T) {
	config := &rest.Config{Host: "https://example.com"}
	service := NewService(&models.Handlers{
		RESTConfigs: map[string]*rest.Config{
			"default": config,
		},
	})

	got, err := service.GetRESTConfig(newRequest(""))
	if err != nil {
		t.Fatalf("GetRESTConfig returned error: %v", err)
	}
	if got != config {
		t.Fatalf("expected default config")
	}
	if _, err := service.GetRESTConfig(newRequest("missing")); err == nil {
		t.Fatalf("expected error for missing cluster")
	}
}
