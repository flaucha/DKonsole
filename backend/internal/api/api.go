package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

const (
	// DefaultListLimit is the maximum number of items to fetch in a single list operation
	// This prevents OOM issues in large clusters
	DefaultListLimit = int64(500)
)

// Service provides API resource and CRD operations
type Service struct {
	clusterService *cluster.Service
}

// NewService creates a new API service
func NewService(cs *cluster.Service) *Service {
	return &Service{
		clusterService: cs,
	}
}

// APIResourceInfo represents information about an API resource
type APIResourceInfo struct {
	Group      string `json:"group"`
	Version    string `json:"version"`
	Resource   string `json:"resource"`
	Kind       string `json:"kind"`
	Namespaced bool   `json:"namespaced"`
}

// APIResourceObject represents an instance of an API resource
type APIResourceObject struct {
	Name      string      `json:"name"`
	Namespace string      `json:"namespace,omitempty"`
	Kind      string      `json:"kind"`
	Status    string      `json:"status,omitempty"`
	Created   string      `json:"created,omitempty"`
	Raw       interface{} `json:"raw,omitempty"`
}

// ListAPIResources lists all available API resources in the cluster
// Refactored to use layered architecture: Handler -> Service -> Repository
func (s *Service) ListAPIResources(w http.ResponseWriter, r *http.Request) {
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create repository
	discoveryRepo := NewK8sDiscoveryRepository(client)

	// Create service
	apiService := NewAPIService(discoveryRepo, nil)

	// Create context
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	// Call service (business logic layer)
	result, err := apiService.ListAPIResources(ctx)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to discover APIs: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, result)
}

// ListAPIResourceObjects lists instances of a specific API resource
// Refactored to use layered architecture: Handler -> Service -> Repository
func (s *Service) ListAPIResourceObjects(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Parse HTTP parameters
	group := r.URL.Query().Get("group")
	version := r.URL.Query().Get("version")
	resource := r.URL.Query().Get("resource")
	namespace := r.URL.Query().Get("namespace")
	namespacedParam := r.URL.Query().Get("namespaced") == "true"

	if namespacedParam && namespace == "" {
		namespace = "default"
	}

	if resource == "" || version == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing resource or version")
		return
	}

	// Get pagination parameters
	limitStr := r.URL.Query().Get("limit")
	limit := DefaultListLimit
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 64); err == nil && parsedLimit > 0 && parsedLimit <= 5000 {
			limit = parsedLimit
		}
	}
	continueToken := r.URL.Query().Get("continue")

	// Create repository
	resourceRepo := NewK8sDynamicResourceRepository(dynamicClient)

	// Create service
	apiService := NewAPIService(nil, resourceRepo)

	// Prepare request
	listReq := ListAPIResourceObjectsRequest{
		Group:         group,
		Version:       version,
		Resource:      resource,
		Namespace:     namespace,
		Namespaced:    namespacedParam,
		Limit:         limit,
		ContinueToken: continueToken,
	}

	// Call service (business logic layer)
	result, err := apiService.ListAPIResourceObjects(ctx, listReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list resource: %v", err))
		return
	}

	// Build response with pagination
	response := map[string]interface{}{
		"resources": result.Resources,
	}
	if result.Continue != "" {
		response["continue"] = result.Continue
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, response)
}

// GetAPIResourceYAML returns the YAML representation of an API resource
// Refactored to use layered architecture: Handler -> Service -> Repository
func (s *Service) GetAPIResourceYAML(w http.ResponseWriter, r *http.Request) {
	// Parse HTTP parameters
	group := r.URL.Query().Get("group")
	version := r.URL.Query().Get("version")
	resource := r.URL.Query().Get("resource")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	namespaced := r.URL.Query().Get("namespaced") == "true"

	if resource == "" || version == "" || name == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing resource, version, or name")
		return
	}

	// Get dynamic client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create repository
	resourceRepo := NewK8sDynamicResourceRepository(dynamicClient)

	// Create service
	apiService := NewAPIService(nil, resourceRepo)

	// Create context
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	// Prepare request
	getReq := GetResourceYAMLRequest{
		Group:      group,
		Version:    version,
		Resource:   resource,
		Name:       name,
		Namespace:  namespace,
		Namespaced: namespaced,
	}

	// Call service to get resource YAML (business logic layer)
	yamlData, err := apiService.GetResourceYAML(ctx, getReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch resource YAML: %v", err))
		return
	}

	// Write YAML response (HTTP layer)
	w.Header().Set("Content-Type", "application/x-yaml")
	_, _ = w.Write([]byte(yamlData))
}

// GetCRDs returns all Custom Resource Definitions
func (s *Service) GetCRDs(w http.ResponseWriter, r *http.Request) {
	// Get dynamic client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create repository
	crdRepo := NewK8sCRDRepository(dynamicClient)
	crdService := NewCRDService(crdRepo)

	// Get pagination parameters
	limitStr := r.URL.Query().Get("limit")
	limit := DefaultListLimit
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 64); err == nil && parsedLimit > 0 && parsedLimit <= 5000 {
			limit = parsedLimit
		}
	}
	continueToken := r.URL.Query().Get("continue")

	req := GetCRDsRequest{
		Limit:         limit,
		ContinueToken: continueToken,
	}

	result, err := crdService.GetCRDs(r.Context(), req)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list CRDs: %v", err))
		return
	}

	utils.JSONResponse(w, http.StatusOK, result)
}

// GetCRDResources lists instances of a specific Custom Resource Definition
func (s *Service) GetCRDResources(w http.ResponseWriter, r *http.Request) {
	// Delegate to generic ListAPIResourceObjects but with CRD specific params if needed
	// Actually GetCRDResources in Service uses APIService.GetCRDResources
	
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	group := r.URL.Query().Get("group")
	version := r.URL.Query().Get("version")
	resource := r.URL.Query().Get("resource")
	namespace := r.URL.Query().Get("namespace")
	namespaced := r.URL.Query().Get("namespaced") == "true"

	// Get pagination parameters
	limitStr := r.URL.Query().Get("limit")
	limit := DefaultListLimit
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 64); err == nil && parsedLimit > 0 && parsedLimit <= 5000 {
			limit = parsedLimit
		}
	}
	continueToken := r.URL.Query().Get("continue")

	resourceRepo := NewK8sDynamicResourceRepository(dynamicClient)
	apiService := NewAPIService(nil, resourceRepo)

	req := GetCRDResourcesRequest{
		Group:         group,
		Version:       version,
		Resource:      resource,
		Namespace:     namespace,
		Namespaced:    namespaced,
		Limit:         limit,
		ContinueToken: continueToken,
	}

	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	result, err := apiService.GetCRDResources(ctx, req)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list CRD resources: %v", err))
		return
	}

	response := map[string]interface{}{
		"resources": result.Resources,
	}
	if result.Continue != "" {
		response["continue"] = result.Continue
	}

	utils.JSONResponse(w, http.StatusOK, response)
}

// GetCRDYaml returns the YAML representation of a Custom Resource
func (s *Service) GetCRDYaml(w http.ResponseWriter, r *http.Request) {
	// Re-use GetAPIResourceYAML
	s.GetAPIResourceYAML(w, r)
}


