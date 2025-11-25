package k8s

import (
	"context"
	"fmt"
	"log"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/example/k8s-view/internal/models"
)

// ResourceRepository defines the interface for accessing Kubernetes resources
type ResourceRepository interface {
	Get(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error)
	Patch(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, patchData []byte, patchType types.PatchType, options metav1.PatchOptions) (*unstructured.Unstructured, error)
	Delete(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, options metav1.DeleteOptions) error
}

// K8sResourceRepository implements ResourceRepository using dynamic client
type K8sResourceRepository struct {
	client dynamic.Interface
}

// NewK8sResourceRepository creates a new K8sResourceRepository
func NewK8sResourceRepository(client dynamic.Interface) *K8sResourceRepository {
	return &K8sResourceRepository{client: client}
}

// getResourceInterface returns the appropriate ResourceInterface
func (r *K8sResourceRepository) getResourceInterface(gvr schema.GroupVersionResource, namespace string, namespaced bool) dynamic.ResourceInterface {
	res := r.client.Resource(gvr)
	if namespaced && namespace != "" {
		return res.Namespace(namespace)
	}
	return res
}

// Get retrieves a resource
func (r *K8sResourceRepository) Get(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error) {
	res := r.getResourceInterface(gvr, namespace, namespaced)
	obj, err := res.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}
	return obj, nil
}

// Patch applies a patch to a resource
func (r *K8sResourceRepository) Patch(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, patchData []byte, patchType types.PatchType, options metav1.PatchOptions) (*unstructured.Unstructured, error) {
	res := r.getResourceInterface(gvr, namespace, namespaced)
	obj, err := res.Patch(ctx, name, patchType, patchData, options)
	if err != nil {
		return nil, fmt.Errorf("failed to patch resource: %w", err)
	}
	return obj, nil
}

// Delete deletes a resource
func (r *K8sResourceRepository) Delete(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, options metav1.DeleteOptions) error {
	res := r.getResourceInterface(gvr, namespace, namespaced)
	err := res.Delete(ctx, name, options)
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}
	return nil
}

// GVRResolver resolves GroupVersionResource from kind and apiVersion
type GVRResolver interface {
	ResolveGVR(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error)
}

// K8sGVRResolver implements GVRResolver
type K8sGVRResolver struct {
	discoveryClient kubernetes.Interface
}

// NewK8sGVRResolver creates a new K8sGVRResolver
func NewK8sGVRResolver() *K8sGVRResolver {
	return &K8sGVRResolver{}
}

// NewK8sGVRResolverWithDiscovery creates a new K8sGVRResolver with discovery client
func NewK8sGVRResolverWithDiscovery(discoveryClient interface{}) *K8sGVRResolver {
	if client, ok := discoveryClient.(kubernetes.Interface); ok {
		return &K8sGVRResolver{discoveryClient: client}
	}
	return &K8sGVRResolver{}
}

// ResolveGVR resolves GVR from kind and apiVersion
func (r *K8sGVRResolver) ResolveGVR(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
	normalizedKind := models.NormalizeKind(kind)

	// Parse apiVersion
	group := ""
	version := apiVersion
	if parts := strings.SplitN(apiVersion, "/", 2); len(parts) == 2 {
		group = parts[0]
		version = parts[1]
	}

	// Default metadata
	meta := models.ResourceMeta{Namespaced: true}
	if namespacedParam == "false" {
		meta.Namespaced = false
	}

	// Try static map first
	if staticGVR, ok := models.ResolveGVR(normalizedKind); ok {
		if staticMeta, metaOk := models.ResourceMetaMap[normalizedKind]; metaOk {
			return staticGVR, staticMeta, nil
		}
		return staticGVR, meta, nil
	}

	// Try discovery API
	if r.discoveryClient != nil {
		lists, err := r.discoveryClient.Discovery().ServerPreferredResources()
		if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
			log.Printf("GVRResolver: Discovery error: %v", err)
		} else {
			for _, list := range lists {
				gv, _ := schema.ParseGroupVersion(list.GroupVersion)
				for _, ar := range list.APIResources {
					if ar.Kind == normalizedKind && !strings.Contains(ar.Name, "/") {
						gvr := schema.GroupVersionResource{
							Group:    gv.Group,
							Version:  gv.Version,
							Resource: ar.Name,
						}
						meta.Namespaced = ar.Namespaced
						return gvr, meta, nil
					}
				}
			}
		}
	}

	// Fallback: infer from kind
	resourceName := strings.ToLower(normalizedKind) + "s"
	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resourceName,
	}

	// Special handling for HPA
	if normalizedKind == "HorizontalPodAutoscaler" {
		gvr = schema.GroupVersionResource{
			Group:    "autoscaling",
			Version:  "v2",
			Resource: "horizontalpodautoscalers",
		}
	}

	return gvr, meta, nil
}
