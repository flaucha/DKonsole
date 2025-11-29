package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/permissions"
)

// ResourceService provides business logic for Kubernetes resource operations.
// It handles resource updates, deletions, and YAML retrieval using the dynamic client.
type ResourceService struct {
	resourceRepo ResourceRepository
	gvrResolver  GVRResolver
}

// NewResourceService creates a new ResourceService with the provided repository and GVR resolver.
func NewResourceService(resourceRepo ResourceRepository, gvrResolver GVRResolver) *ResourceService {
	return &ResourceService{
		resourceRepo: resourceRepo,
		gvrResolver:  gvrResolver,
	}
}

// UpdateResourceRequest represents parameters for updating a Kubernetes resource.
type UpdateResourceRequest struct {
	YAMLContent string // YAML content of the resource to update
	Kind        string // Resource kind (e.g., "Pod", "Deployment")
	Name        string // Resource name
	Namespace   string // Namespace (empty for cluster-scoped resources)
	Namespaced  bool   // Whether the resource is namespaced
}

// UpdateResource updates a Kubernetes resource from YAML content.
// It parses the YAML, resolves the GroupVersionResource, and applies the changes using a strategic merge patch.
// Returns an error if the YAML is invalid, the resource cannot be found, or the update fails.
// It validates namespace permissions before updating.
func (s *ResourceService) UpdateResource(ctx context.Context, req UpdateResourceRequest) error {
	// Validate namespace access if resource is namespaced
	if req.Namespace != "" {
		// Check if user has edit permission
		canEdit, err := permissions.CanPerformAction(ctx, req.Namespace, "edit")
		if err != nil {
			return fmt.Errorf("failed to check permissions: %w", err)
		}
		if !canEdit {
			return fmt.Errorf("edit permission required for namespace: %s", req.Namespace)
		}
	}
	// Parse YAML to JSON
	jsonData, err := yaml.YAMLToJSON([]byte(req.YAMLContent))
	if err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	// Unmarshal to unstructured
	var obj unstructured.Unstructured
	if err := json.Unmarshal(jsonData, &obj.Object); err != nil {
		return fmt.Errorf("failed to parse resource: %w", err)
	}

	normalizedKind := models.NormalizeKind(req.Kind)
	if obj.GetKind() != "" {
		normalizedKind = models.NormalizeKind(obj.GetKind())
	}

	// Resolve GVR
	gvr, meta, err := s.gvrResolver.ResolveGVR(ctx, normalizedKind, obj.GetAPIVersion(), fmt.Sprintf("%t", req.Namespaced))
	if err != nil {
		return fmt.Errorf("failed to resolve GVR: %w", err)
	}

	if req.Namespaced {
		meta.Namespaced = true
	}

	// Normalize name and namespace
	if req.Name != "" {
		obj.SetName(req.Name)
	}
	if req.Namespace != "" && meta.Namespaced {
		obj.SetNamespace(req.Namespace)
	} else if !meta.Namespaced {
		obj.SetNamespace("")
	}

	// Cleanup metadata
	unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")
	unstructured.RemoveNestedField(obj.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(obj.Object, "metadata", "uid")

	// Marshal for patch
	patchData, err := json.Marshal(obj.Object)
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	// Apply patch
	force := true
	patchOptions := metav1.PatchOptions{
		FieldManager: "dkonsole",
		Force:        &force,
	}

	_, err = s.resourceRepo.Patch(ctx, gvr, obj.GetName(), obj.GetNamespace(), meta.Namespaced, patchData, types.ApplyPatchType, patchOptions)
	if err != nil {
		return fmt.Errorf("failed to update resource: %w", err)
	}

	return nil
}

// CreateResource creates a Kubernetes resource from YAML content.
// It validates namespace permissions before creating.
func (s *ResourceService) CreateResource(ctx context.Context, yamlContent string) (interface{}, error) {
	// Parse YAML to JSON
	jsonData, err := yaml.YAMLToJSON([]byte(yamlContent))
	if err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	// Unmarshal to unstructured
	var obj unstructured.Unstructured
	if err := json.Unmarshal(jsonData, &obj.Object); err != nil {
		return nil, fmt.Errorf("failed to parse resource: %w", err)
	}

	// Resolve GVR
	gvr, meta, err := s.gvrResolver.ResolveGVR(ctx, obj.GetKind(), obj.GetAPIVersion(), "")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve GVR: %w", err)
	}

	namespace := obj.GetNamespace()
	if meta.Namespaced && namespace == "" {
		namespace = "default"
		obj.SetNamespace(namespace)
	}

	created, err := s.resourceRepo.Create(ctx, gvr, namespace, meta.Namespaced, &obj, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	return created.Object, nil
}

// DeleteResourceRequest represents parameters for deleting a Kubernetes resource.
type DeleteResourceRequest struct {
	Kind      string // Resource kind (e.g., "Pod", "Deployment")
	Name      string // Resource name
	Namespace string // Namespace (required for namespaced resources)
	Force     bool   // If true, force deletion with 0 grace period
}

// DeleteResource deletes a Kubernetes resource.
// By default, it uses a 30-second grace period and foreground deletion propagation.
// If Force is true, it uses 0 grace period and background propagation.
// Returns an error if the resource cannot be found or deletion fails.
// It validates namespace permissions before deleting.
func (s *ResourceService) DeleteResource(ctx context.Context, req DeleteResourceRequest) error {
	normalizedKind := models.NormalizeKind(req.Kind)

	gvr, meta, err := s.gvrResolver.ResolveGVR(ctx, normalizedKind, "", "")
	if err != nil {
		return fmt.Errorf("failed to resolve GVR: %w", err)
	}

	if meta.Namespaced && req.Namespace == "" {
		return fmt.Errorf("namespace is required for namespaced resource deletion")
	}

	// Validate namespace access if resource is namespaced
	if meta.Namespaced && req.Namespace != "" {
		// Check if user has edit permission
		canEdit, err := permissions.CanPerformAction(ctx, req.Namespace, "edit")
		if err != nil {
			return fmt.Errorf("failed to check permissions: %w", err)
		}
		if !canEdit {
			return fmt.Errorf("edit permission required for namespace: %s", req.Namespace)
		}
	}

	var grace int64 = 30
	propagation := metav1.DeletePropagationForeground
	if req.Force {
		grace = 0
		propagation = metav1.DeletePropagationBackground
	}

	delOpts := metav1.DeleteOptions{
		GracePeriodSeconds: &grace,
		PropagationPolicy:  &propagation,
	}

	err = s.resourceRepo.Delete(ctx, gvr, req.Name, req.Namespace, meta.Namespaced, delOpts)
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}
	return nil
}

// GetResourceRequest represents parameters for retrieving a Kubernetes resource.
type GetResourceRequest struct {
	Kind          string // Resource kind (e.g., "Pod", "Deployment")
	Name          string // Resource name
	Namespace     string // Namespace (defaults to "default" for namespaced resources)
	AllNamespaces bool   // If true, ignore namespace filter
	Group         string // API group (optional)
	Version       string // API version (optional)
	Resource      string // Resource name (optional)
	Namespaced    bool   // Whether the resource is namespaced
}

// GetResourceYAML fetches a Kubernetes resource and returns its YAML representation.
// It resolves the GroupVersionResource, retrieves the resource, and converts it to YAML format.
// Returns the YAML string or an error if the resource cannot be found or retrieved.
func (s *ResourceService) GetResourceYAML(ctx context.Context, req GetResourceRequest, discoveryClient interface{}) (string, error) {
	normalizedKind := models.NormalizeKind(req.Kind)

	apiVersion := ""
	if req.Group != "" && req.Version != "" {
		apiVersion = fmt.Sprintf("%s/%s", req.Group, req.Version)
	} else if req.Version != "" {
		apiVersion = req.Version
	}

	gvr, meta, err := s.gvrResolver.ResolveGVR(ctx, normalizedKind, apiVersion, fmt.Sprintf("%t", req.Namespaced))
	if err != nil {
		return "", fmt.Errorf("failed to resolve GVR: %w", err)
	}

	if req.Namespaced {
		meta.Namespaced = true
	}

	namespace := req.Namespace
	if meta.Namespaced && namespace == "" && !req.AllNamespaces {
		namespace = "default"
	}

	obj, err := s.resourceRepo.Get(ctx, gvr, req.Name, namespace, meta.Namespaced)
	if err != nil {
		return "", fmt.Errorf("failed to get resource: %w", err)
	}

	// Clean up the object to ensure clean YAML output
	// Remove metadata fields that clutter the view
	unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")
	unstructured.RemoveNestedField(obj.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(obj.Object, "metadata", "uid")
	unstructured.RemoveNestedField(obj.Object, "metadata", "generation")
	unstructured.RemoveNestedField(obj.Object, "metadata", "selfLink")

	jsonData, err := obj.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal resource: %w", err)
	}

	// Parse JSON to ensure clean structure
	var cleanObj map[string]interface{}
	if err := json.Unmarshal(jsonData, &cleanObj); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Ensure apiVersion appears only once
	if apiVer, exists := cleanObj["apiVersion"]; exists {
		// Remove and re-add to ensure it's at the root level only
		delete(cleanObj, "apiVersion")
		cleanObj["apiVersion"] = apiVer
	}

	// Re-marshal to JSON
	cleanJSON, err := json.Marshal(cleanObj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal clean JSON: %w", err)
	}

	yamlData, err := yaml.JSONToYAML(cleanJSON)
	if err != nil {
		return "", fmt.Errorf("failed to convert to YAML: %w", err)
	}

	return string(yamlData), nil
}
