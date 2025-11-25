package prometheus

import (
	"fmt"
	"log"
	"net/http"

	"github.com/example/k8s-view/internal/cluster"
	"github.com/example/k8s-view/internal/utils"
)

// HTTPHandler handles HTTP requests for Prometheus metrics
type HTTPHandler struct {
	prometheusURL  string
	repo           Repository
	clusterService *cluster.Service
	promService    *Service
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
	status := StatusResponse{
		Enabled: h.prometheusURL != "",
		URL:     h.prometheusURL,
	}
	utils.JSONResponse(w, http.StatusOK, status)
}

// GetMetrics handles requests for deployment metrics
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Repository (Data Access)
func (h *HTTPHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetPrometheusMetrics: PrometheusURL=%s, deployment=%s, namespace=%s",
		h.prometheusURL, r.URL.Query().Get("deployment"), r.URL.Query().Get("namespace"))

	if h.prometheusURL == "" {
		log.Printf("GetPrometheusMetrics: Prometheus URL not configured")
		utils.ErrorResponse(w, http.StatusServiceUnavailable, "Prometheus URL not configured")
		return
	}

	deployment := r.URL.Query().Get("deployment")
	namespace := r.URL.Query().Get("namespace")
	rangeParam := r.URL.Query().Get("range")

	if deployment == "" || namespace == "" {
		log.Printf("GetPrometheusMetrics: Missing deployment or namespace")
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
	response, err := h.promService.GetDeploymentMetrics(ctx, req)
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
	log.Printf("GetPrometheusPodMetrics: PrometheusURL=%s, pod=%s, namespace=%s",
		h.prometheusURL, r.URL.Query().Get("pod"), r.URL.Query().Get("namespace"))

	if h.prometheusURL == "" {
		utils.ErrorResponse(w, http.StatusServiceUnavailable, "Prometheus URL not configured")
		return
	}

	podName := r.URL.Query().Get("pod")
	namespace := r.URL.Query().Get("namespace")
	rangeParam := r.URL.Query().Get("range")

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
	response, err := h.promService.GetPodMetrics(ctx, req)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get pod metrics: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, response)
}

// GetClusterOverview handles requests for cluster overview metrics
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Repository (Data Access)
func (h *HTTPHandler) GetClusterOverview(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetPrometheusClusterOverview: PrometheusURL=%s", h.prometheusURL)

	if h.prometheusURL == "" {
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
	response, err := h.promService.GetClusterOverview(ctx, req, client)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get cluster overview: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, response)
}

