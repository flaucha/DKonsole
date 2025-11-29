package k8s

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/permissions"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// Service provides HTTP handlers for Kubernetes resource operations.
// It follows a layered architecture pattern with dependency injection via ServiceFactory.
type Service struct {
	handlers       *models.Handlers
	clusterService *cluster.Service
	serviceFactory *ServiceFactory
}

// NewService creates a new Kubernetes service with the provided handlers and cluster service.
func NewService(h *models.Handlers, cs *cluster.Service) *Service {
	return &Service{
		handlers:       h,
		clusterService: cs,
		serviceFactory: NewServiceFactory(),
	}
}

// GetNamespaces handles HTTP GET requests to retrieve all Kubernetes namespaces.
// Returns a JSON array of namespace objects.
//
// @Summary Listar namespaces
// @Description Retorna todos los namespaces de Kubernetes
// @Tags k8s
// @Security Bearer
// @Produce json
// @Success 200 {array} object "Lista de namespaces"
// @Failure 400 {object} map[string]string "Error en la solicitud"
// @Failure 500 {object} map[string]string "Error del servidor"
// @Router /api/namespaces [get]
//
// Example response:
//
//	[{"name": "default", "status": "Active"}, ...]
func (s *Service) GetNamespaces(w http.ResponseWriter, r *http.Request) {
	// Get Kubernetes client for this request
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create service using factory (dependency injection)
	namespaceService := s.serviceFactory.CreateNamespaceService(client)

	// Create context from request (to access user permissions)
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	// Call service to get namespaces (business logic layer)
	// The service will filter namespaces based on user permissions
	namespaces, err := namespaceService.GetNamespaces(ctx)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get namespaces: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, namespaces)
}

// GetResources handles HTTP GET requests to list Kubernetes resources of a specific kind.
// Query parameters:
//   - kind: The resource kind (e.g., "Pod", "Deployment", "Service")
//   - namespace: The namespace to filter by, or "all" for all namespaces
//
// Returns a JSON array of resource objects with metadata and status information.
func (s *Service) GetResources(w http.ResponseWriter, r *http.Request) {
	// Parse HTTP parameters
	ns := r.URL.Query().Get("namespace")
	kind := r.URL.Query().Get("kind")
	labelSelector := r.URL.Query().Get("labelSelector")
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
		LabelSelector: labelSelector,
		Client:        client,
		MetricsClient: metricsClient,
	}

	// Call service to get resources (business logic layer)
	// The service will filter resources based on permissions
	resources, err := resourceListService.ListResources(ctx, req)
	if err != nil {
		// Check if it's a permission error
		if err.Error() == fmt.Sprintf("access denied to namespace: %s", req.Namespace) {
			utils.ErrorResponse(w, http.StatusForbidden, err.Error())
			return
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list resources: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, resources)
}

// ScaleResource handles HTTP POST requests to scale a Kubernetes Deployment.
// Query parameters:
//   - kind: Must be "Deployment"
//   - name: The deployment name
//   - namespace: The namespace (defaults to "default" if empty)
//   - delta: The number of replicas to add or subtract (positive or negative integer)
//
// Returns the new replica count on success.
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

	// Validate namespace access
	ctx := r.Context()
	canEdit, err := permissions.CanPerformAction(ctx, namespace, "edit")
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check permissions: %v", err))
		return
	}
	if !canEdit {
		utils.ErrorResponse(w, http.StatusForbidden, fmt.Sprintf("Edit permission required for namespace: %s", namespace))
		return
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

// GetClusterStats handles HTTP GET requests to retrieve cluster-wide statistics.
// Returns counts of nodes, namespaces, pods, deployments, services, ingresses, PVCs, and PVs.
//
// Example response:
//
//	{
//	  "nodes": 3,
//	  "namespaces": 10,
//	  "pods": 45,
//	  ...
//	}
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
