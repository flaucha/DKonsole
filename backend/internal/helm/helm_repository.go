package helm

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// HelmReleaseRepository defines the interface for accessing Helm releases stored in Secrets/ConfigMaps
type HelmReleaseRepository interface {
	ListHelmSecrets(ctx context.Context) ([]corev1.Secret, error)
	ListHelmConfigMaps(ctx context.Context) ([]corev1.ConfigMap, error)
	ListSecretsInNamespace(ctx context.Context, namespace string) ([]corev1.Secret, error)
	ListConfigMapsInNamespace(ctx context.Context, namespace string) ([]corev1.ConfigMap, error)
	DeleteSecret(ctx context.Context, namespace, name string) error
	DeleteConfigMap(ctx context.Context, namespace, name string) error
}

// K8sHelmReleaseRepository implements HelmReleaseRepository using Kubernetes client-go
type K8sHelmReleaseRepository struct {
	client kubernetes.Interface
}

// NewK8sHelmReleaseRepository creates a new K8sHelmReleaseRepository
func NewK8sHelmReleaseRepository(client kubernetes.Interface) *K8sHelmReleaseRepository {
	return &K8sHelmReleaseRepository{client: client}
}

// ListHelmSecrets lists all Secrets with Helm labels across all namespaces
func (r *K8sHelmReleaseRepository) ListHelmSecrets(ctx context.Context) ([]corev1.Secret, error) {
	secrets, err := r.client.CoreV1().Secrets("").List(ctx, metav1.ListOptions{
		LabelSelector: "owner=helm",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list Helm secrets: %w", err)
	}
	return secrets.Items, nil
}

// ListHelmConfigMaps lists all ConfigMaps with Helm labels across all namespaces
func (r *K8sHelmReleaseRepository) ListHelmConfigMaps(ctx context.Context) ([]corev1.ConfigMap, error) {
	configMaps, err := r.client.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{
		LabelSelector: "owner=helm",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list Helm configmaps: %w", err)
	}
	return configMaps.Items, nil
}

// ListSecretsInNamespace lists all Secrets in a specific namespace
func (r *K8sHelmReleaseRepository) ListSecretsInNamespace(ctx context.Context, namespace string) ([]corev1.Secret, error) {
	secrets, err := r.client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets in namespace %s: %w", namespace, err)
	}
	return secrets.Items, nil
}

// ListConfigMapsInNamespace lists all ConfigMaps in a specific namespace
func (r *K8sHelmReleaseRepository) ListConfigMapsInNamespace(ctx context.Context, namespace string) ([]corev1.ConfigMap, error) {
	configMaps, err := r.client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps in namespace %s: %w", namespace, err)
	}
	return configMaps.Items, nil
}

// DeleteSecret deletes a Secret
func (r *K8sHelmReleaseRepository) DeleteSecret(ctx context.Context, namespace, name string) error {
	err := r.client.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete secret %s/%s: %w", namespace, name, err)
	}
	return nil
}

// DeleteConfigMap deletes a ConfigMap
func (r *K8sHelmReleaseRepository) DeleteConfigMap(ctx context.Context, namespace, name string) error {
	err := r.client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete configmap %s/%s: %w", namespace, name, err)
	}
	return nil
}







