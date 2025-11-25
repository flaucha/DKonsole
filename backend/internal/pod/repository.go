package pod

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EventRepository defines the interface for fetching pod events from Kubernetes
type EventRepository interface {
	GetEvents(ctx context.Context, namespace, podName string) ([]corev1.Event, error)
}

// K8sEventRepository implements EventRepository using Kubernetes client-go
type K8sEventRepository struct {
	client kubernetes.Interface
}

// NewK8sEventRepository creates a new K8sEventRepository
func NewK8sEventRepository(client kubernetes.Interface) *K8sEventRepository {
	return &K8sEventRepository{client: client}
}

// GetEvents fetches events for a specific pod
func (r *K8sEventRepository) GetEvents(ctx context.Context, namespace, podName string) ([]corev1.Event, error) {
	events, err := r.client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pod events: %w", err)
	}
	return events.Items, nil
}
