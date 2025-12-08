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
				"Data": len(item.Data),
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
		resources = append(resources, models.Resource{
			UID:       string(item.UID),
			Name:      item.Name,
			Namespace: item.Namespace,
			Kind:      "Secret",
			Created:   item.CreationTimestamp.Format(time.RFC3339),
			Status:    "Active",
			Details: map[string]interface{}{
				"Type": string(item.Type),
				"Data": len(item.Data),
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
