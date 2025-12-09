package k8s

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func (s *ResourceListService) listConfigMaps(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().ConfigMaps(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	resources := make([]models.Resource, 0, len(list.Items))
	for _, item := range list.Items {
		resources = append(resources, models.Resource{
			UID:       string(item.UID),
			Name:      item.Name,
			Namespace: item.Namespace,
			Kind:      "ConfigMap",
			Created:   item.CreationTimestamp.Format(time.RFC3339),
			Status:    "Active",
			Details: map[string]interface{}{
				"data":      item.Data,
				"dataCount": len(item.Data),
			},
		})
	}
	return resources, nil
}

func (s *ResourceListService) listSecrets(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().Secrets(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	resources := make([]models.Resource, 0, len(list.Items))
	for _, item := range list.Items {
		// Convert []byte values to base64 strings for JSON serialization
		dataMap := make(map[string]string)
		for k, v := range item.Data {
			dataMap[k] = string(v)
		}
		resources = append(resources, models.Resource{
			UID:       string(item.UID),
			Name:      item.Name,
			Namespace: item.Namespace,
			Kind:      "Secret",
			Created:   item.CreationTimestamp.Format(time.RFC3339),
			Status:    string(item.Type),
			Details: map[string]interface{}{
				"Type":      string(item.Type),
				"data":      dataMap,
				"dataCount": len(item.Data),
			},
		})
	}
	return resources, nil
}

func (s *ResourceListService) listServiceAccounts(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().ServiceAccounts(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	resources := make([]models.Resource, 0, len(list.Items))
	for _, item := range list.Items {
		resources = append(resources, models.Resource{
			UID:       string(item.UID),
			Name:      item.Name,
			Namespace: item.Namespace,
			Kind:      "ServiceAccount",
			Created:   item.CreationTimestamp.Format(time.RFC3339),
			Status:    "Active",
			Details: map[string]interface{}{
				"Secrets": len(item.Secrets),
			},
		})
	}
	return resources, nil
}
