package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"

	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/utils"
)

// UpdateResourceYAML handles HTTP PUT requests to update a Kubernetes resource from YAML.
// Query parameters:
//   - kind: The resource kind
//   - name: The resource name
//   - namespace: The namespace (if namespaced)
//   - namespaced: "true" if the resource is namespaced
//
// Request body should contain the YAML representation of the resource.
// Returns a success status on completion.
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

	// Read YAML from request body (HTTP layer)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Failed to read request body: %v", err))
		return
	}
	defer r.Body.Close()

	yamlData := string(body)

	// Create context with timeout
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Update the resource (Business Logic)
	err = s.UpdateResource(ctx, kind, name, namespace, namespaced, yamlData)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Write success response (HTTP layer)
	utils.SuccessResponse(w, "Resource updated successfully", nil)
}

// UpdateResource implements the business logic for updating a resource
func (s *Service) UpdateResource(ctx context.Context, kind, name, namespace string, namespaced bool, yamlData string) error {
	// Convert YAML to JSON for Kubernetes API
	jsonData, err := utils.YamlToJson(yamlData)
	if err != nil {
		return fmt.Errorf("invalid YAML: %v", err)
	}

	// Resolve GVR
	gvr, _, err := s.resolver.ResolveGVR(ctx, kind, "", "")
	if err != nil {
		return err
	}

	// Parse JSON into unstructured object to get metadata
	var obj unstructured.Unstructured
	if err := json.Unmarshal(jsonData, &obj); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Ensure metadata matches request parameters
	if obj.GetName() != name {
		return fmt.Errorf("resource name in YAML (%s) does not match query parameter (%s)", obj.GetName(), name)
	}
	if namespaced && obj.GetNamespace() != namespace {
		return fmt.Errorf("resource namespace in YAML (%s) does not match query parameter (%s)", obj.GetNamespace(), namespace)
	}

	// Get current resource to verify it exists and for patch generation
	currentObj, err := s.repo.Get(ctx, gvr, name, namespace, namespaced)
	if err != nil {
		return fmt.Errorf("failed to get resource: %v", err)
	}

	// Update metadata resource version to ensure consistency
	obj.SetResourceVersion(currentObj.GetResourceVersion())

	// Re-serialize to JSON with resource version
	updatedJSON, err := obj.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal updated object: %v", err)
	}

	// Apply the update using Patch (Strategic Merge Patch or Merge Patch)
	// Note: Using Apply or Update would be cleaner, but Patch is often safer for partial updates
	// Here we use Update mechanism via Patch or native Update if repository supported it
	// Since our repo has Patch, let's use Patch with MergePatchType which overwrites fields
	// Or better, implement Update in repository. For now, assuming repo only has Patch.
	// Actually, repo.Patch is available.
	// To fully replace, we should use application/json-patch+json or just Update.
	// Given the constraints, let's try to use the repository's methods.
	// The repository interface in resource_repository.go only shows Get, Patch, Delete.
	// So we must use Patch.

	// Ideally we would use types.ApplyPatchType if server supports it, or just update.
	// But without Update method, we can try to use Patch with the full content.
	_, err = s.repo.Patch(ctx, gvr, name, namespace, namespaced, updatedJSON, "application/merge-patch+json", metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to update resource: %v", err)
	}

	return nil
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

	// Create the resource
	result, err := s.CreateResource(ctx, yamlData)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Write SSE event if this is a stream, but standard POST returns JSON
	// The frontend seems to expect a simple success or the created object
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// CreateResource implements logic to create a resource from YAML
func (s *Service) CreateResource(ctx context.Context, yamlData string) (interface{}, error) {
	// Convert YAML to JSON
	jsonData, err := utils.YamlToJson(yamlData)
	if err != nil {
		return nil, fmt.Errorf("invalid YAML: %v", err)
	}

	// Parse into unstructured
	var obj unstructured.Unstructured
	if err := json.Unmarshal(jsonData, &obj); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	gvk := obj.GroupVersionKind()
	kind := gvk.Kind
	apiVersion := gvk.GroupVersion().String()
	namespace := obj.GetNamespace()
	name := obj.GetName()

	// Resolve GVR
	gvr, meta, err := s.resolver.ResolveGVR(ctx, kind, apiVersion, "")
	if err != nil {
		return nil, err
	}

	// Validate namespace if namespaced
	if meta.Namespaced && namespace == "" {
		namespace = "default"
		obj.SetNamespace(namespace)
	}

	// Get dynamic client
	// Note: Repository interface doesn't have Create. We need access to dynamic client.
	// This is a limitation of the current repository abstraction.
	// We'll have to access the client directly or extend the repository.
	// The Service struct has `client` which is `kubernetes.Interface` (typed)
	// and `dynamicClient` which is `dynamic.Interface`.
	// Check Service definition in resource_service.go
	// It seems Service has `client kubernetes.Interface` and `dynamicClient dynamic.Interface`.

	// Using dynamic client to create
	// We need to re-marshal the object to pass to Create
	// Actually Create takes *unstructured.Unstructured

	var created *unstructured.Unstructured

	if meta.Namespaced {
		created, err = s.dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, &obj, metav1.CreateOptions{})
	} else {
		created, err = s.dynamicClient.Resource(gvr).Create(ctx, &obj, metav1.CreateOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	return created.Object, nil
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

	// Attempt creation
	result, err := s.CreateResource(ctx, yamlData)

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
	// Similar to Create but with DryRun option
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	yamlData := string(body)
	jsonData, err := utils.YamlToJson(yamlData)
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

	gvk := obj.GroupVersionKind()
	gvr, meta, err := s.resolver.ResolveGVR(ctx, gvk.Kind, gvk.GroupVersion().String(), "")
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
		result, err = s.dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, &obj, options)
	} else {
		result, err = s.dynamicClient.Resource(gvr).Create(ctx, &obj, options)
	}

	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Dry run failed: %v", err))
		return
	}

	utils.SuccessResponse(w, "Dry run successful", result.Object)
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
	jsonData, err := utils.YamlToJson(yamlData)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid YAML syntax: %v", err))
		return
	}

	// Schema validation could go here (using openapi schema)
	// For now, we check if we can unmarshal into unstructured and resolve GVK
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

	_, _, err = s.resolver.ResolveGVR(ctx, gvk.Kind, gvk.GroupVersion().String(), "")
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Unknown resource type: %v", err))
		return
	}

	utils.SuccessResponse(w, "YAML is valid", map[string]string{
		"kind":    gvk.Kind,
		"version": gvk.Version,
		"name":    obj.GetName(),
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

	jsonData, err := utils.YamlToJson(yamlData)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"message\": \"Invalid YAML: %v\"}\n\n", err)
		flusher.Flush()
		return
	}

	var obj unstructured.Unstructured
	json.Unmarshal(jsonData, &obj)

	gvk := obj.GroupVersionKind()
	gvr, meta, err := s.resolver.ResolveGVR(ctx, gvk.Kind, gvk.GroupVersion().String(), "")
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
		result, err = s.dynamicClient.Resource(gvr).Namespace(namespace).Patch(ctx, obj.GetName(), types.ApplyPatchType, jsonData, options)
	} else {
		result, err = s.dynamicClient.Resource(gvr).Patch(ctx, obj.GetName(), types.ApplyPatchType, jsonData, options)
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
