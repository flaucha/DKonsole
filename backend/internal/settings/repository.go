package settings

import (
	"context"
	"fmt"
	"os"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// Repository defines the interface for settings data access
type Repository interface {
	GetPrometheusURL(ctx context.Context) (string, error)
	UpdatePrometheusURL(ctx context.Context, url string) error
}

// K8sRepository implements Repository using Kubernetes ConfigMap
type K8sRepository struct {
	client     kubernetes.Interface
	namespace  string
	configMapName string
	secretName string
}

// NewRepository creates a new K8sRepository instance
func NewRepository(client kubernetes.Interface, secretName string) *K8sRepository {
	namespace, err := getCurrentNamespace()
	if err != nil {
		// Fallback to default namespace
		namespace = "default"
		utils.LogWarn("Failed to get current namespace, using default", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Use the same name pattern as the secret for consistency
	configMapName := secretName
	if configMapName == "" {
		configMapName = "dkonsole-auth"
	}

	return &K8sRepository{
		client:        client,
		namespace:     namespace,
		configMapName: configMapName,
		secretName:    secretName,
	}
}

// getCurrentNamespace retrieves the current namespace from the pod's service account
func getCurrentNamespace() (string, error) {
	// Try reading from service account namespace file (standard in Kubernetes pods)
	nsFile := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	if data, err := os.ReadFile(nsFile); err == nil {
		namespace := string(data)
		if namespace != "" {
			return namespace, nil
		}
	}

	// Fallback to environment variable
	if namespace := os.Getenv("POD_NAMESPACE"); namespace != "" {
		return namespace, nil
	}

	return "", fmt.Errorf("could not determine namespace: service account file not found and POD_NAMESPACE not set")
}

// GetPrometheusURL retrieves the Prometheus URL from ConfigMap or environment variable
func (r *K8sRepository) GetPrometheusURL(ctx context.Context) (string, error) {
	// First try to get from ConfigMap
	if r.client != nil {
		configMap, err := r.client.CoreV1().ConfigMaps(r.namespace).Get(ctx, r.configMapName, metav1.GetOptions{})
		if err == nil {
			if url, exists := configMap.Data["prometheus-url"]; exists && url != "" {
				return url, nil
			}
		} else if !apierrors.IsNotFound(err) {
			utils.LogWarn("Failed to get ConfigMap for Prometheus URL", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Fallback to environment variable
	if url := os.Getenv("PROMETHEUS_URL"); url != "" {
		return url, nil
	}

	return "", nil
}

// UpdatePrometheusURL updates the Prometheus URL in ConfigMap
func (r *K8sRepository) UpdatePrometheusURL(ctx context.Context, url string) error {
	if r.client == nil {
		return fmt.Errorf("Kubernetes client not available")
	}

	// Try to get existing ConfigMap
	configMap, err := r.client.CoreV1().ConfigMaps(r.namespace).Get(ctx, r.configMapName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new ConfigMap
			configMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      r.configMapName,
					Namespace: r.namespace,
				},
				Data: map[string]string{
					"prometheus-url": url,
				},
			}
			_, err = r.client.CoreV1().ConfigMaps(r.namespace).Create(ctx, configMap, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create ConfigMap: %w", err)
			}
			utils.LogInfo("Created ConfigMap for Prometheus URL", map[string]interface{}{
				"configmap_name": r.configMapName,
				"namespace":      r.namespace,
			})
			return nil
		}
		return fmt.Errorf("failed to get ConfigMap: %w", err)
	}

	// Update existing ConfigMap
	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}
	configMap.Data["prometheus-url"] = url

	_, err = r.client.CoreV1().ConfigMaps(r.namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap: %w", err)
	}

	utils.LogInfo("Updated Prometheus URL in ConfigMap", map[string]interface{}{
		"configmap_name": r.configMapName,
		"namespace":      r.namespace,
		"url":             url,
	})

	return nil
}
