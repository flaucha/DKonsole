package k8s

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/example/k8s-view/internal/cluster"
	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/utils"
)

// Service provides Kubernetes resource operations
type Service struct {
	handlers       *models.Handlers
	clusterService *cluster.Service
	serviceFactory *ServiceFactory
}

// NewService creates a new Kubernetes service
func NewService(h *models.Handlers, cs *cluster.Service) *Service {
	return &Service{
		handlers:       h,
		clusterService: cs,
		serviceFactory: NewServiceFactory(),
	}
}

// GetNamespaces returns a list of all namespaces
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Repository (Data Access)
func (s *Service) GetNamespaces(w http.ResponseWriter, r *http.Request) {
	// Get Kubernetes client for this request
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create service using factory (dependency injection)
	namespaceService := s.serviceFactory.CreateNamespaceService(client)

	// Create context with timeout
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Call service to get namespaces (business logic layer)
	namespaces, err := namespaceService.GetNamespaces(ctx)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get namespaces: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, namespaces)
}

// GetResources lists resources of a specific kind
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Repository (Data Access)
func (s *Service) GetResources(w http.ResponseWriter, r *http.Request) {
	// Parse HTTP parameters
	ns := r.URL.Query().Get("namespace")
	kind := r.URL.Query().Get("kind")
	allNamespaces := ns == "all"

	if kind == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing kind parameter")
		return
	}

	// Normalize kind (handle aliases like HPA -> HorizontalPodAutoscaler)
	normalizedKind := models.NormalizeKind(kind)

	// Get clients
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	metricsClient := s.clusterService.GetMetricsClient(r)

	// Create service
	resourceListService := NewResourceListService(s.clusterService, s.handlers.PrometheusURL)

	// Create context
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	// Prepare request
	req := ListResourcesRequest{
		Kind:          normalizedKind,
		Namespace:     ns,
		AllNamespaces: allNamespaces,
		Client:        client,
		MetricsClient: metricsClient,
	}

	// Call service to get resources (business logic layer)
	resources, err := resourceListService.ListResources(ctx, req)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list resources: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, resources)
}

// ScaleResource scales a deployment
// Refactored to use utils helpers
func (s *Service) ScaleResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	deltaStr := r.URL.Query().Get("delta")

	if kind != "Deployment" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Scaling supported only for Deployments")
		return
	}
	if name == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing name")
		return
	}

	if err := utils.ValidateK8sName(name, "name"); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if namespace != "" {
		if err := utils.ValidateK8sName(namespace, "namespace"); err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if namespace == "" {
		namespace = "default"
	}

	delta, err := strconv.Atoi(deltaStr)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid delta: %v", err))
		return
	}
	if delta == 0 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Delta cannot be zero")
		return
	}

	// Create service using factory (dependency injection)
	deploymentService := s.serviceFactory.CreateDeploymentService(client)

	// Create context
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Call service to scale deployment (business logic layer)
	newReplicas, err := deploymentService.ScaleDeployment(ctx, namespace, name, delta)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to scale deployment: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, map[string]int32{"replicas": newReplicas})
}

// WatchResources is implemented in resource_operations.go

// GetClusterStats returns cluster statistics
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Repository (Data Access)
func (s *Service) GetClusterStats(w http.ResponseWriter, r *http.Request) {
	// Use request context with timeout so cancellation propagates
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	// Get Kubernetes client for this request
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create service using factory (dependency injection)
	statsService := s.serviceFactory.CreateClusterStatsService(client)

	// Call service to get cluster stats (business logic layer)
	stats, err := statsService.GetClusterStats(ctx)
	if err != nil {
		// Note: GetClusterStats handles errors gracefully and returns partial stats
		// If we want to be stricter, we could use GetClusterStatsWithErrors
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get cluster stats: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, stats)
}

// TriggerCronJob is implemented in cronjob.go
