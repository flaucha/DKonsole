package cluster

import (
	"fmt"
	"net/http"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/example/k8s-view/internal/models"
)

// Service provides cluster management operations
type Service struct {
	handlers *models.Handlers
}

// NewService creates a new cluster service
func NewService(h *models.Handlers) *Service {
	return &Service{handlers: h}
}

// GetClient returns the Kubernetes client for the specified cluster
func (s *Service) GetClient(r *http.Request) (*kubernetes.Clientset, error) {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	s.handlers.RLock()
	defer s.handlers.RUnlock()

	client, ok := s.handlers.Clients[cluster]
	if !ok {
		return nil, fmt.Errorf("cluster not found: %s", cluster)
	}
	return client, nil
}

// GetDynamicClient returns the dynamic client for the specified cluster
func (s *Service) GetDynamicClient(r *http.Request) (dynamic.Interface, error) {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	s.handlers.RLock()
	defer s.handlers.RUnlock()

	client, ok := s.handlers.Dynamics[cluster]
	if !ok {
		return nil, fmt.Errorf("cluster not found: %s", cluster)
	}
	return client, nil
}

// GetMetricsClient returns the metrics client for the specified cluster
func (s *Service) GetMetricsClient(r *http.Request) *metricsv.Clientset {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	s.handlers.RLock()
	defer s.handlers.RUnlock()

	client, ok := s.handlers.Metrics[cluster]
	if !ok {
		return nil
	}
	return client
}

// GetRESTConfig returns the REST config for the specified cluster
func (s *Service) GetRESTConfig(r *http.Request) (*rest.Config, error) {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	s.handlers.RLock()
	defer s.handlers.RUnlock()

	config, ok := s.handlers.RESTConfigs[cluster]
	if !ok {
		return nil, fmt.Errorf("cluster not found: %s", cluster)
	}
	return config, nil
}
