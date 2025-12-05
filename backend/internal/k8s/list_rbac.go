package k8s

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func (s *ResourceListService) listRoles(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.RbacV1().Roles(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		details := map[string]interface{}{
			"rules": i.Rules,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "Role",
			Status:    "Active",
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listClusterRoles(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.RbacV1().ClusterRoles().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		details := map[string]interface{}{
			"rules": i.Rules,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: "",
			Kind:      "ClusterRole",
			Status:    "Active",
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listRoleBindings(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.RbacV1().RoleBindings(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		details := map[string]interface{}{
			"subjects": i.Subjects,
			"roleRef":  i.RoleRef,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "RoleBinding",
			Status:    "Active",
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listClusterRoleBindings(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.RbacV1().ClusterRoleBindings().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		details := map[string]interface{}{
			"subjects": i.Subjects,
			"roleRef":  i.RoleRef,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: "",
			Kind:      "ClusterRoleBinding",
			Status:    "Active",
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}
