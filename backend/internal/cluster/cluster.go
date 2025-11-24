package cluster

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/utils"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Service provides cluster management operations
type Service struct {
	handlers *models.Handlers
}

// NewService creates a new cluster service
func NewService(h *models.Handlers) *Service {
	return &Service{handlers: h}
}

// GetClusters returns a list of all configured clusters
func (s *Service) GetClusters(w http.ResponseWriter, r *http.Request) {
	s.handlers.RLock()
	defer s.handlers.RUnlock()

	var clusters []string
	for name := range s.handlers.Clients {
		clusters = append(clusters, name)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clusters)
}

// AddCluster adds a new cluster configuration
func (s *Service) AddCluster(w http.ResponseWriter, r *http.Request) {
	var config models.ClusterConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if config.Name == "" || config.Host == "" || config.Token == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	s.handlers.Lock()
	defer s.handlers.Unlock()

	if _, exists := s.handlers.Clients[config.Name]; exists {
		http.Error(w, "Cluster with this name already exists", http.StatusConflict)
		return
	}

	k8sConfig := &rest.Config{
		Host:        config.Host,
		BearerToken: config.Token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: config.Insecure,
		},
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		utils.HandleError(w, err, "Failed to create client", http.StatusInternalServerError)
		return
	}

	dynamicClient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		utils.HandleError(w, err, "Failed to create dynamic client", http.StatusInternalServerError)
		return
	}

	metricsClient, _ := metricsv.NewForConfig(k8sConfig)

	s.handlers.Clients[config.Name] = clientset
	s.handlers.Dynamics[config.Name] = dynamicClient
	s.handlers.RESTConfigs[config.Name] = k8sConfig
	if metricsClient != nil {
		s.handlers.Metrics[config.Name] = metricsClient
	}
	w.WriteHeader(http.StatusCreated)
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

