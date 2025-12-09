package helm

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// HelmJobRepository defines the interface for creating Kubernetes Jobs and ConfigMaps
type HelmJobRepository interface {
	CreateConfigMap(ctx context.Context, namespace string, cm *corev1.ConfigMap) error
	CreateJob(ctx context.Context, namespace string, job *batchv1.Job) error
	GetServiceAccount(ctx context.Context, namespace, name string) (*corev1.ServiceAccount, error)
}

// K8sHelmJobRepository implements HelmJobRepository
type K8sHelmJobRepository struct {
	client kubernetes.Interface
}

// NewK8sHelmJobRepository creates a new K8sHelmJobRepository
func NewK8sHelmJobRepository(client kubernetes.Interface) *K8sHelmJobRepository {
	return &K8sHelmJobRepository{client: client}
}

// CreateConfigMap creates a ConfigMap
func (r *K8sHelmJobRepository) CreateConfigMap(ctx context.Context, namespace string, cm *corev1.ConfigMap) error {
	_, err := r.client.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create configmap: %w", err)
	}
	return nil
}

// CreateJob creates a Job
func (r *K8sHelmJobRepository) CreateJob(ctx context.Context, namespace string, job *batchv1.Job) error {
	_, err := r.client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}
	return nil
}

// GetServiceAccount gets a ServiceAccount
func (r *K8sHelmJobRepository) GetServiceAccount(ctx context.Context, namespace, name string) (*corev1.ServiceAccount, error) {
	sa, err := r.client.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get serviceaccount: %w", err)
	}
	return sa, nil
}

// HelmJobService provides business logic for creating Helm Jobs
type HelmJobService struct {
	repo HelmJobRepository
}

// NewHelmJobService creates a new HelmJobService
func NewHelmJobService(repo HelmJobRepository) *HelmJobService {
	return &HelmJobService{repo: repo}
}
