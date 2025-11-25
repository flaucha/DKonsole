package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"

	"github.com/example/k8s-view/internal/models"
)

// WatchService provides business logic for watching Kubernetes resources
type WatchService struct {
	gvrResolver GVRResolver
}

// NewWatchService creates a new WatchService
func NewWatchService(gvrResolver GVRResolver) *WatchService {
	return &WatchService{
		gvrResolver: gvrResolver,
	}
}

// WatchRequest represents parameters for watching resources
type WatchRequest struct {
	Kind          string
	Namespace     string
	AllNamespaces bool
}

// WatchResult represents a single watch event
type WatchResult struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// StartWatch initializes a Kubernetes watch for the specified resource type
// Returns a watcher interface that can be used to receive events
func (s *WatchService) StartWatch(ctx context.Context, client dynamic.Interface, req WatchRequest) (watch.Interface, error) {
	// Normalize kind (handle aliases like HPA -> HorizontalPodAutoscaler)
	normalizedKind := models.NormalizeKind(req.Kind)

	// Get resource metadata
	meta, ok := models.ResourceMetaMap[normalizedKind]
	if !ok {
		return nil, fmt.Errorf("unsupported kind: %s", normalizedKind)
	}

	// Resolve GVR
	gvr, ok := models.ResolveGVR(normalizedKind)
	if !ok {
		return nil, fmt.Errorf("failed to resolve GVR for kind: %s", normalizedKind)
	}

	// Determine namespace
	namespace := req.Namespace
	if namespace == "" {
		namespace = "default"
	}

	// Get the appropriate ResourceInterface
	var res dynamic.ResourceInterface
	if meta.Namespaced {
		if req.AllNamespaces {
			res = client.Resource(gvr)
		} else {
			res = client.Resource(gvr).Namespace(namespace)
		}
	} else {
		res = client.Resource(gvr)
	}

	// Create watcher
	watcher, err := res.Watch(ctx, metav1.ListOptions{
		Watch: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start watch: %w", err)
	}

	return watcher, nil
}

// TransformEvent transforms a Kubernetes watch event to a WatchResult DTO
func (s *WatchService) TransformEvent(event watch.Event) (*WatchResult, error) {
	obj, ok := event.Object.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("event object is not unstructured")
	}

	return &WatchResult{
		Type:      string(event.Type),
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, nil
}
