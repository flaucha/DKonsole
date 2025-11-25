package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/example/k8s-view/internal/models"
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
func (s *ResourceService) UpdateResource(ctx context.Context, req UpdateResourceRequest) error {
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
func (s *ResourceService) DeleteResource(ctx context.Context, req DeleteResourceRequest) error {
	normalizedKind := models.NormalizeKind(req.Kind)

	gvr, meta, err := s.gvrResolver.ResolveGVR(ctx, normalizedKind, "", "")
	if err != nil {
		return fmt.Errorf("failed to resolve GVR: %w", err)
	}

	if meta.Namespaced && req.Namespace == "" {
		return fmt.Errorf("namespace is required for namespaced resource deletion")
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

	jsonData, err := obj.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal resource: %w", err)
	}

	yamlData, err := yaml.JSONToYAML(jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to convert to YAML: %w", err)
	}

	return string(yamlData), nil
}
