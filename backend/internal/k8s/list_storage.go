package k8s

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func (s *ResourceListService) listPVCs(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		requested := ""
		if i.Spec.Resources.Requests != nil {
			if storage, ok := i.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
				requested = storage.String()
			}
		}

		allocated := ""
		if i.Status.Capacity != nil {
			if storage, ok := i.Status.Capacity[corev1.ResourceStorage]; ok {
				allocated = storage.String()
			}
		}

		// Note: Prometheus metrics for PVC usage would require queryPrometheusInstant
		// For now, we skip Prometheus metrics in the refactored version
		// This can be added later if needed by injecting a Prometheus client

		details := map[string]interface{}{
			"accessModes":      i.Spec.AccessModes,
			"capacity":         allocated,
			"requested":        requested,
			"storageClassName": i.Spec.StorageClassName,
			"volumeName":       i.Spec.VolumeName,
		}

		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "PersistentVolumeClaim",
			Status:    string(i.Status.Phase),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listPVs(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().PersistentVolumes().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: "",
			Kind:      "PersistentVolume",
			Status:    string(i.Status.Phase),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"accessModes":      i.Spec.AccessModes,
				"capacity":         i.Spec.Capacity.Storage().String(),
				"storageClassName": i.Spec.StorageClassName,
				"claimRef":         i.Spec.ClaimRef,
			},
		})
	}

	return resources, nil
}

func (s *ResourceListService) listStorageClasses(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.StorageV1().StorageClasses().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		reclaim := ""
		if i.ReclaimPolicy != nil {
			reclaim = string(*i.ReclaimPolicy)
		}
		volumeBinding := ""
		if i.VolumeBindingMode != nil {
			volumeBinding = string(*i.VolumeBindingMode)
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: "",
			Kind:      "StorageClass",
			Status:    reclaim,
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"provisioner":          i.Provisioner,
				"reclaimPolicy":        reclaim,
				"volumeBindingMode":    volumeBinding,
				"allowVolumeExpansion": i.AllowVolumeExpansion,
				"parameters":           i.Parameters,
				"mountOptions":         i.MountOptions,
			},
		})
	}

	return resources, nil
}

func (s *ResourceListService) listResourceQuotas(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().ResourceQuotas(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "ResourceQuota",
			Status:    "Active",
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"hard": i.Status.Hard,
				"used": i.Status.Used,
			},
		})
	}

	return resources, nil
}

func (s *ResourceListService) listLimitRanges(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().LimitRanges(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "LimitRange",
			Status:    "Active",
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"limits": i.Spec.Limits,
			},
		})
	}

	return resources, nil
}
