package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/flaucha/DKonsole/backend/internal/permissions"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

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

	// Validate namespace access if resource is namespaced
	ctx := r.Context()
	if namespaced && namespace != "" {
		// Check if user has edit permission
		canEdit, err := permissions.CanPerformAction(ctx, namespace, "edit")
		if err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check permissions: %v", err))
			return
		}
		if !canEdit {
			utils.ErrorResponse(w, http.StatusForbidden, fmt.Sprintf("Edit permission required for namespace: %s", namespace))
			return
		}
	}

	// Read YAML from request body
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

	// Create Resource Service
	resourceService := s.serviceFactory.CreateResourceService(dynamicClient)

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
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Write success response
	utils.JSONResponse(w, http.StatusOK, map[string]string{"message": "Resource updated successfully"})
}

// CreateResourceYAML handles HTTP POST requests to create a Kubernetes resource from YAML.
func (s *Service) CreateResourceYAML(w http.ResponseWriter, r *http.Request) {
	// Read YAML from request body
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

	// Create Resource Service
	resourceService := s.serviceFactory.CreateResourceService(dynamicClient)

	// Create the resource
	result, err := resourceService.CreateResource(ctx, yamlData)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Write success response
	utils.JSONResponse(w, http.StatusCreated, result)
}

// StreamResourceCreation handles SSE for resource creation feedback
func (s *Service) StreamResourceCreation(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Read YAML from body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"message\": \"Failed to read body: %v\"}\n\n", err)
		flusher.Flush()
		return
	}
	defer r.Body.Close()

	yamlData := string(body)

	// Send "start" event
	fmt.Fprintf(w, "event: status\ndata: {\"message\": \"Starting creation...\"}\n\n")
	flusher.Flush()

	// Create context
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Get Dynamic Client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		errorData, _ := json.Marshal(map[string]string{"message": err.Error()})
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", errorData)
		flusher.Flush()
		return
	}

	// Create Resource Service
	resourceService := s.serviceFactory.CreateResourceService(dynamicClient)

	// Attempt creation
	result, err := resourceService.CreateResource(ctx, yamlData)

	if err != nil {
		// Send error event
		errorData, _ := json.Marshal(map[string]string{"message": err.Error()})
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", errorData)
		flusher.Flush()
	} else {
		// Send success event
		successData, _ := json.Marshal(result)
		fmt.Fprintf(w, "event: success\ndata: %s\n\n", successData)
		flusher.Flush()
	}
}

// DryRunResourceYAML performs a dry-run creation
func (s *Service) DryRunResourceYAML(w http.ResponseWriter, r *http.Request) {
	// Read YAML
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	yamlData := string(body)
	jsonData, err := yaml.YAMLToJSON([]byte(yamlData))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid YAML: %v", err))
		return
	}

	var obj unstructured.Unstructured
	if err := json.Unmarshal(jsonData, &obj); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Get Dynamic Client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resolver := NewK8sGVRResolver()

	gvk := obj.GroupVersionKind()
	gvr, meta, err := resolver.ResolveGVR(ctx, gvk.Kind, gvk.GroupVersion().String(), "")
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	namespace := obj.GetNamespace()
	if meta.Namespaced && namespace == "" {
		namespace = "default"
		obj.SetNamespace(namespace)
	}

	// Server-side DryRun
	var result *unstructured.Unstructured
	options := metav1.CreateOptions{
		DryRun: []string{metav1.DryRunAll},
	}

	if meta.Namespaced {
		result, err = dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, &obj, options)
	} else {
		result, err = dynamicClient.Resource(gvr).Create(ctx, &obj, options)
	}

	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Dry run failed: %v", err))
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Dry run successful",
		"object":  result.Object,
	})
}

// ValidateResourceYAML validates the YAML content
func (s *Service) ValidateResourceYAML(w http.ResponseWriter, r *http.Request) {
	// Read YAML
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	yamlData := string(body)

	// Basic YAML validation
	jsonData, err := yaml.YAMLToJSON([]byte(yamlData))
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid YAML syntax: %v", err))
		return
	}

	// Unmarshal
	var obj unstructured.Unstructured
	if err := json.Unmarshal(jsonData, &obj); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid resource structure: %v", err))
		return
	}

	gvk := obj.GroupVersionKind()
	if gvk.Kind == "" || gvk.Version == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing Kind or APIVersion")
		return
	}

	// Check if we can resolve the resource type
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	resolver := NewK8sGVRResolver()
	_, _, err = resolver.ResolveGVR(ctx, gvk.Kind, gvk.GroupVersion().String(), "")
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Unknown resource type: %v", err))
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "YAML is valid",
		"data": map[string]string{
			"kind":    gvk.Kind,
			"version": gvk.Version,
			"name":    obj.GetName(),
		},
	})
}

// ServerSideApply performs a server-side apply
func (s *Service) ServerSideApply(w http.ResponseWriter, r *http.Request) {
	// SSE setup
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	body, _ := io.ReadAll(r.Body)
	yamlData := string(body)

	fmt.Fprintf(w, "data: {\"message\": \"Starting Server-Side Apply...\"}\n\n")
	flusher.Flush()

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	jsonData, err := yaml.YAMLToJSON([]byte(yamlData))
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"message\": \"Invalid YAML: %v\"}\n\n", err)
		flusher.Flush()
		return
	}

	var obj unstructured.Unstructured
	json.Unmarshal(jsonData, &obj)

	// Get Dynamic Client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"message\": \"Cluster error: %v\"}\n\n", err)
		flusher.Flush()
		return
	}

	resolver := NewK8sGVRResolver()
	gvk := obj.GroupVersionKind()
	gvr, meta, err := resolver.ResolveGVR(ctx, gvk.Kind, gvk.GroupVersion().String(), "")
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"message\": \"Resolution error: %v\"}\n\n", err)
		flusher.Flush()
		return
	}

	namespace := obj.GetNamespace()
	if meta.Namespaced && namespace == "" {
		namespace = "default"
		obj.SetNamespace(namespace)
	}

	force := true
	options := metav1.PatchOptions{
		FieldManager: "dkonsole-apply",
		Force:        &force,
	}

	var result *unstructured.Unstructured
	if meta.Namespaced {
		result, err = dynamicClient.Resource(gvr).Namespace(namespace).Patch(ctx, obj.GetName(), types.ApplyPatchType, jsonData, options)
	} else {
		result, err = dynamicClient.Resource(gvr).Patch(ctx, obj.GetName(), types.ApplyPatchType, jsonData, options)
	}

	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"message\": \"Apply failed: %v\"}\n\n", err)
		flusher.Flush()
	} else {
		resBytes, _ := json.Marshal(result)
		fmt.Fprintf(w, "event: success\ndata: %s\n\n", resBytes)
		flusher.Flush()
	}
}

// WatchResources handles WebSocket connections for watching Kubernetes resources
func (s *Service) WatchResources(w http.ResponseWriter, r *http.Request) {
	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.LogError(err, "Failed to upgrade to WebSocket", nil)
		return
	}
	defer conn.Close()

	// Parse parameters
	kind := r.URL.Query().Get("kind")
	namespace := r.URL.Query().Get("namespace")
	allNamespaces := r.URL.Query().Get("namespace") == "all"

	// Get Dynamic Client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		_ = conn.WriteJSON(map[string]string{"type": "ERROR", "message": err.Error()})
		return
	}

	// Create WatchService
	watchService := s.serviceFactory.CreateWatchService()

	// Create context that is canceled when connection closes
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Handle client disconnect
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				cancel()
				return
			}
		}
	}()

	// Start Watch
	req := WatchRequest{
		Kind:          kind,
		Namespace:     namespace,
		AllNamespaces: allNamespaces,
	}

	watcher, err := watchService.StartWatch(ctx, dynamicClient, req)
	if err != nil {
		_ = conn.WriteJSON(map[string]string{"type": "ERROR", "message": err.Error()})
		return
	}
	defer watcher.Stop()

	// Send events to client
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return
			}

			// Transform event
			result, err := watchService.TransformEvent(event)
			if err != nil {
				continue
			}

			// Send to client
			if err := conn.WriteJSON(result); err != nil {
				return
			}
		}
	}
}

// ImportResourceYAML handles HTTP POST requests to import multiple Kubernetes resources from YAML.
func (s *Service) ImportResourceYAML(w http.ResponseWriter, r *http.Request) {
	// Read YAML from request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Failed to read request body: %v", err))
		return
	}
	defer r.Body.Close()

	// Create context
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

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
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to import resources: %v", err))
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
			utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check permissions: %v", err))
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

	// Create Resource Service
	resourceService := s.serviceFactory.CreateResourceService(dynamicClient)

	// Delete resource
	req := DeleteResourceRequest{
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
		Force:     force,
	}

	if err := resourceService.DeleteResource(ctx, req); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Write success response
	utils.JSONResponse(w, http.StatusOK, map[string]string{"message": "Resource deleted successfully"})
}
