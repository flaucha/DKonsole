package k8s

import (
	"fmt"
	"net/http"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/flaucha/DKonsole/backend/internal/permissions"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// GetResourceYAML returns the YAML representation of a resource
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Repository (Data Access)
func (s *Service) GetResourceYAML(w http.ResponseWriter, r *http.Request) {
	// Parse HTTP parameters
	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	allNamespaces := namespace == "all"
	namespacedParam := r.URL.Query().Get("namespaced") == "true"
	group := r.URL.Query().Get("group")
	version := r.URL.Query().Get("version")
	resourceName := r.URL.Query().Get("resource")

	if kind == "" || name == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing kind or name parameter")
		return
	}

	// Validate parameters
	if err := utils.ValidateK8sName(name, "name"); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if namespace != "" && namespace != "all" {
		if err := utils.ValidateK8sName(namespace, "namespace"); err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Get dynamic client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get discovery client for GVR resolution
	discoveryClient, err := s.clusterService.GetClient(r)
	var discoveryInterface interface{}
	if err == nil && discoveryClient != nil {
		discoveryInterface = discoveryClient
	}

	// Create repository and resolver
	resourceRepo := NewK8sResourceRepository(dynamicClient)
	gvrResolver := NewK8sGVRResolverWithDiscovery(discoveryInterface)

	// Create service
	resourceService := NewResourceService(resourceRepo, gvrResolver)

	// Create context
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	// Validate namespace access if resource is namespaced
	if namespacedParam && namespace != "" && namespace != "all" {
		hasAccess, err := permissions.HasNamespaceAccess(ctx, namespace)
		if err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check permissions: %v", err))
			return
		}
		if !hasAccess {
			utils.ErrorResponse(w, http.StatusForbidden, fmt.Sprintf("Access denied to namespace: %s", namespace))
			return
		}
	}

	// Prepare request
	getReq := GetResourceRequest{
		Kind:          kind,
		Name:          name,
		Namespace:     namespace,
		AllNamespaces: allNamespaces,
		Group:         group,
		Version:       version,
		Resource:      resourceName,
		Namespaced:    namespacedParam,
	}

	// Call service to get resource YAML (business logic layer)
	yamlData, err := resourceService.GetResourceYAML(ctx, getReq, discoveryInterface)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if apierrors.IsNotFound(err) {
			statusCode = http.StatusNotFound
		} else if apierrors.IsForbidden(err) {
			statusCode = http.StatusForbidden
		} else if apierrors.IsBadRequest(err) {
			statusCode = http.StatusBadRequest
		}
		utils.ErrorResponse(w, statusCode, fmt.Sprintf("Failed to fetch resource: %v", err))
		return
	}

	// Write YAML response (HTTP layer)
	w.Header().Set("Content-Type", "application/x-yaml")
	_, _ = w.Write([]byte(yamlData))
}
