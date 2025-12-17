package k8s

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/flaucha/DKonsole/backend/internal/permissions"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

const (
	// MaxBodySize limits the size of request bodies to 1MB to prevent DoS
	MaxBodySize = 1048576
)

// UpdateResourceYAML handles HTTP PUT requests to update a Kubernetes resource from YAML.
func (s *Service) UpdateResourceYAML(w http.ResponseWriter, r *http.Request) {
	// Parse HTTP parameters
	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	namespacedParam := r.URL.Query().Get("namespaced")
	namespaced := namespacedParam == "true"

	if kind == "" || name == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing required parameters: kind, name")
		return
	}

	// Read YAML from request body
	// Limit request body size to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodySize)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Failed to read request body: %v", err))
		return
	}
	defer r.Body.Close()

	yamlData := string(body)

	// Use request context (already has permissions validated above)
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	// Get Dynamic Client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get Client (for discovery)
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create Resource Service
	resourceService := s.serviceFactory.CreateResourceService(dynamicClient, client)

	// Update the resource
	req := UpdateResourceRequest{
		YAMLContent: yamlData,
		Kind:        kind,
		Name:        name,
		Namespace:   namespace,
		Namespaced:  namespaced,
	}

	err = resourceService.UpdateResource(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNamespaceMismatch), errors.Is(err, ErrNamespaceRequired):
			utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrForbidden), errors.Is(err, ErrAdminRequired):
			utils.ErrorResponse(w, http.StatusForbidden, err.Error())
		default:
			if strings.HasPrefix(err.Error(), "invalid YAML") {
				utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			utils.HandleErrorJSON(w, err, "Failed to update resource", http.StatusInternalServerError, map[string]interface{}{
				"kind":       kind,
				"name":       name,
				"namespace":  namespace,
				"namespaced": namespaced,
			})
		}
		return
	}

	// Write success response
	utils.JSONResponse(w, http.StatusOK, map[string]string{"message": "Resource updated successfully"})
}

// CreateResourceYAML handles HTTP POST requests to create a Kubernetes resource from YAML.
func (s *Service) CreateResourceYAML(w http.ResponseWriter, r *http.Request) {
	// Read YAML from request body
	// Limit request body size to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodySize)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Failed to read request body: %v", err))
		return
	}
	defer r.Body.Close()

	yamlData := string(body)

	// Create context
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Get Dynamic Client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get Client (for discovery)
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create Resource Service
	resourceService := s.serviceFactory.CreateResourceService(dynamicClient, client)

	// Create the resource
	result, err := resourceService.CreateResource(ctx, yamlData)
	if err != nil {
		utils.HandleErrorJSON(w, err, "Failed to create resource", http.StatusInternalServerError, nil)
		return
	}

	// Write success response
	utils.JSONResponse(w, http.StatusCreated, result)
}

// ImportResourceYAML handles HTTP POST requests to import multiple Kubernetes resources from YAML.
func (s *Service) ImportResourceYAML(w http.ResponseWriter, r *http.Request) {
	// Read YAML from request body
	// Limit request body size to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodySize)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Failed to read request body: %v", err))
		return
	}
	defer r.Body.Close()

	// Use request context (has user permissions)
	ctx := r.Context()

	// Parse YAML to check namespaces before importing
	// We need to validate permissions for each resource's namespace
	// For now, we'll validate during the import process in ImportService

	// Get Clients
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create Import Service
	importService := s.serviceFactory.CreateImportService(dynamicClient, client)

	// Prepare request
	req := ImportResourceRequest{
		YAMLContent: body,
	}

	// Import resources
	result, err := importService.ImportResources(ctx, req)
	if err != nil {
		utils.HandleErrorJSON(w, err, "Failed to import resources", http.StatusInternalServerError, nil)
		return
	}

	// Write success response
	utils.JSONResponse(w, http.StatusOK, result)
}

// DeleteResource handles HTTP DELETE requests to delete a Kubernetes resource.
func (s *Service) DeleteResource(w http.ResponseWriter, r *http.Request) {
	// Parse HTTP parameters
	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	force := r.URL.Query().Get("force") == "true"

	if kind == "" || name == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing required parameters: kind, name")
		return
	}

	// Validate namespace access if namespace is provided
	ctx := r.Context()
	if namespace != "" {
		// Check if user has edit permission
		canEdit, err := permissions.CanPerformAction(ctx, namespace, "edit")
		if err != nil {
			utils.HandleErrorJSON(w, err, "Failed to check permissions", http.StatusInternalServerError, map[string]interface{}{
				"namespace": namespace,
				"action":    "edit",
			})
			return
		}
		if !canEdit {
			utils.ErrorResponse(w, http.StatusForbidden, fmt.Sprintf("Edit permission required for namespace: %s", namespace))
			return
		}
	}

	// Create context
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	// Get Dynamic Client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get Client (for discovery)
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create Resource Service
	resourceService := s.serviceFactory.CreateResourceService(dynamicClient, client)

	// Delete resource
	req := DeleteResourceRequest{
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
		Force:     force,
	}

	if err := resourceService.DeleteResource(ctx, req); err != nil {
		utils.HandleErrorJSON(w, err, "Failed to delete resource", http.StatusInternalServerError, map[string]interface{}{
			"kind":      kind,
			"name":      name,
			"namespace": namespace,
			"force":     force,
		})
		return
	}

	// Write success response
	utils.JSONResponse(w, http.StatusOK, map[string]string{"message": "Resource deleted successfully"})
}
