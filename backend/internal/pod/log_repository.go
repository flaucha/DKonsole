package pod

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// LogRepository defines the interface for streaming pod logs from Kubernetes
type LogRepository interface {
	GetLogStream(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error)
}

// K8sLogRepository is an implementation of LogRepository that interacts with the Kubernetes API
type K8sLogRepository struct {
	client kubernetes.Interface
}

// NewK8sLogRepository creates a new K8sLogRepository
func NewK8sLogRepository(client kubernetes.Interface) *K8sLogRepository {
	return &K8sLogRepository{client: client}
}

// GetLogStream fetches a log stream for a specific pod from the Kubernetes API
func (r *K8sLogRepository) GetLogStream(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error) {
	req := r.client.CoreV1().Pods(namespace).GetLogs(podName, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open log stream from K8s API: %w", err)
	}
	return stream, nil
}









