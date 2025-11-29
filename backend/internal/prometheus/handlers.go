package prometheus

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// HTTPHandler handles HTTP requests for Prometheus metrics
type HTTPHandler struct {
	prometheusURL  string
	repo           Repository
	clusterService *cluster.Service
	promService    *Service
	mu             sync.RWMutex // Mutex for thread-safe URL updates
}

// NewHTTPHandler creates a new Prometheus HTTP handler
func NewHTTPHandler(prometheusURL string, clusterService *cluster.Service) *HTTPHandler {
	repo := NewHTTPPrometheusRepository(prometheusURL)
	promService := NewService(repo)

	return &HTTPHandler{
		prometheusURL:  prometheusURL,
		repo:           repo,
		clusterService: clusterService,
		promService:    promService,
	}
}

// GetStatus returns the Prometheus service status
func (h *HTTPHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	url := h.prometheusURL
	h.mu.RUnlock()

	status := StatusResponse{
		Enabled: url != "",
		URL:     url,
	}
	utils.JSONResponse(w, http.StatusOK, status)
}

// GetMetrics handles requests for deployment metrics
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Repository (Data Access)
func (h *HTTPHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	deployment := r.URL.Query().Get("deployment")
	namespace := r.URL.Query().Get("namespace")
	rangeParam := r.URL.Query().Get("range")

	h.mu.RLock()
	url := h.prometheusURL
	promService := h.promService
	h.mu.RUnlock()

	if url == "" {
		utils.LogWarn("Prometheus URL not configured", nil)
		utils.ErrorResponse(w, http.StatusServiceUnavailable, "Prometheus URL not configured")
		return
	}

	// Verify repository is using correct URL
	h.mu.RLock()
	repoNil := h.repo == nil
	h.mu.RUnlock()
	if repoNil {
		utils.LogWarn("Prometheus repository is nil, recreating", nil)
		h.mu.Lock()
		h.repo = NewHTTPPrometheusRepository(url)
		h.promService = NewService(h.repo)
		promService = h.promService
		h.mu.Unlock()
	}

	if deployment == "" || namespace == "" {
		utils.LogWarn("Missing deployment or namespace", map[string]interface{}{
			"deployment": deployment,
			"namespace":  namespace,
		})
		utils.ErrorResponse(w, http.StatusBadRequest, "deployment and namespace are required")
		return
	}

	// Create context
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	// Prepare request
	req := GetDeploymentMetricsRequest{
		Deployment: deployment,
		Namespace:  namespace,
		Range:      rangeParam,
	}

	// Call service (business logic layer)
	response, err := promService.GetDeploymentMetrics(ctx, req)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get metrics: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, response)
}

// GetPodMetrics handles requests for pod metrics
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Repository (Data Access)
func (h *HTTPHandler) GetPodMetrics(w http.ResponseWriter, r *http.Request) {
	podName := r.URL.Query().Get("pod")
	namespace := r.URL.Query().Get("namespace")
	rangeParam := r.URL.Query().Get("range")

	h.mu.RLock()
	url := h.prometheusURL
	promService := h.promService
	h.mu.RUnlock()

	if url == "" {
		utils.ErrorResponse(w, http.StatusServiceUnavailable, "Prometheus URL not configured")
		return
	}

	// Verify repository is using correct URL
	h.mu.RLock()
	repoNil := h.repo == nil
	h.mu.RUnlock()
	if repoNil {
		utils.LogWarn("Prometheus repository is nil, recreating", nil)
		h.mu.Lock()
		h.repo = NewHTTPPrometheusRepository(url)
		h.promService = NewService(h.repo)
		promService = h.promService
		h.mu.Unlock()
	}

	if podName == "" || namespace == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "pod and namespace are required")
		return
	}

	// Create context
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	// Prepare request
	req := GetPodMetricsRequest{
		PodName:   podName,
		Namespace: namespace,
		Range:     rangeParam,
	}

	// Call service (business logic layer)
	response, err := promService.GetPodMetrics(ctx, req)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get pod metrics: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, response)
}

// UpdateURL updates the Prometheus URL and recreates the repository and service
func (h *HTTPHandler) UpdateURL(newURL string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.prometheusURL = newURL
	h.repo = NewHTTPPrometheusRepository(newURL)
	h.promService = NewService(h.repo)
}

// GetClusterOverview handles requests for cluster overview metrics
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Repository (Data Access)
func (h *HTTPHandler) GetClusterOverview(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	url := h.prometheusURL
	promService := h.promService
	h.mu.RUnlock()

	if url == "" {
		utils.ErrorResponse(w, http.StatusServiceUnavailable, "Prometheus URL not configured")
		return
	}

	// Get Kubernetes client
	client, err := h.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	rangeParam := r.URL.Query().Get("range")

	// Create context
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	// Prepare request
	req := GetClusterOverviewRequest{
		Range: rangeParam,
	}

	// Call service (business logic layer)
	response, err := promService.GetClusterOverview(ctx, req, client)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get cluster overview: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, response)
}
