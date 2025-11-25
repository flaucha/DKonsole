package api

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// APIResourceRepository defines the interface for discovering and listing API resources
type APIResourceRepository interface {
	ListAPIResources(ctx context.Context) ([]*metav1.APIResourceList, error)
	ListAPIResourceObjects(ctx context.Context, gvr schema.GroupVersionResource, namespace string, listOpts metav1.ListOptions) (*unstructured.UnstructuredList, error)
}

// K8sDiscoveryRepository implements APIResourceRepository for discovery
type K8sDiscoveryRepository struct {
	client kubernetes.Interface
}

// NewK8sDiscoveryRepository creates a new K8sDiscoveryRepository
func NewK8sDiscoveryRepository(client kubernetes.Interface) *K8sDiscoveryRepository {
	return &K8sDiscoveryRepository{client: client}
}

// ListAPIResources discovers all available API resources
func (r *K8sDiscoveryRepository) ListAPIResources(ctx context.Context) ([]*metav1.APIResourceList, error) {
	lists, err := r.client.Discovery().ServerPreferredResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return nil, fmt.Errorf("failed to discover APIs: %w", err)
	}
	return lists, nil
}

// ListAPIResourceObjects lists instances of a specific API resource
func (r *K8sDiscoveryRepository) ListAPIResourceObjects(ctx context.Context, gvr schema.GroupVersionResource, namespace string, listOpts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return nil, fmt.Errorf("not implemented - use K8sDynamicResourceRepository")
}

// DynamicResourceRepository defines the interface for dynamic resource operations
type DynamicResourceRepository interface {
	Get(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error)
	List(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, limit int64, continueToken string) (*unstructured.UnstructuredList, string, error)
}

// K8sDynamicResourceRepository implements DynamicResourceRepository
type K8sDynamicResourceRepository struct {
	client dynamic.Interface
}

// NewK8sDynamicResourceRepository creates a new K8sDynamicResourceRepository
func NewK8sDynamicResourceRepository(client dynamic.Interface) *K8sDynamicResourceRepository {
	return &K8sDynamicResourceRepository{client: client}
}

// Get retrieves a resource
func (r *K8sDynamicResourceRepository) Get(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error) {
	var res dynamic.ResourceInterface
	if namespaced && namespace != "" {
		res = r.client.Resource(gvr).Namespace(namespace)
	} else {
		res = r.client.Resource(gvr)
	}
	obj, err := res.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}
	return obj, nil
}

// List lists resources
func (r *K8sDynamicResourceRepository) List(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, limit int64, continueToken string) (*unstructured.UnstructuredList, string, error) {
	var res dynamic.ResourceInterface
	if namespaced && namespace != "" && namespace != "all" {
		res = r.client.Resource(gvr).Namespace(namespace)
	} else {
		res = r.client.Resource(gvr)
	}
	listOpts := metav1.ListOptions{
		Limit:    limit,
		Continue: continueToken,
	}
	list, err := res.List(ctx, listOpts)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list resources: %w", err)
	}
	return list, list.GetContinue(), nil
}

// CRDRepository defines the interface for Custom Resource Definitions
type CRDRepository interface {
	ListCRDs(ctx context.Context, limit int64, continueToken string) (*unstructured.UnstructuredList, error)
}

// K8sCRDRepository implements CRDRepository
type K8sCRDRepository struct {
	client dynamic.Interface
}

// NewK8sCRDRepository creates a new K8sCRDRepository
func NewK8sCRDRepository(client dynamic.Interface) *K8sCRDRepository {
	return &K8sCRDRepository{client: client}
}

// ListCRDs lists all Custom Resource Definitions
func (r *K8sCRDRepository) ListCRDs(ctx context.Context, limit int64, continueToken string) (*unstructured.UnstructuredList, error) {
	gvr := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}
	listOpts := metav1.ListOptions{
		Limit:    limit,
		Continue: continueToken,
	}
	unstructuredList, err := r.client.Resource(gvr).List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list CRDs: %w", err)
	}
	return unstructuredList, nil
}
