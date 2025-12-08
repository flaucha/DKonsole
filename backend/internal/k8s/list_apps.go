package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func (s *ResourceListService) listDeployments(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.AppsV1().Deployments(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		var images []string
		var ports []int32
		var pvcs []string

		// Aggregate requests and limits from all containers
		var totalRequestsCPU, totalRequestsMem, totalLimitsCPU, totalLimitsMem resource.Quantity
		var hasRequestsCPU, hasRequestsMem, hasLimitsCPU, hasLimitsMem bool
		for _, c := range i.Spec.Template.Spec.Containers {
			images = append(images, c.Image)
			for _, p := range c.Ports {
				ports = append(ports, p.ContainerPort)
			}

			// Sum up requests and limits from all containers
			if c.Resources.Requests != nil {
				if cpu, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
					if !hasRequestsCPU {
						totalRequestsCPU = cpu.DeepCopy()
						hasRequestsCPU = true
					} else {
						totalRequestsCPU.Add(cpu)
					}
				}
				if mem, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
					if !hasRequestsMem {
						totalRequestsMem = mem.DeepCopy()
						hasRequestsMem = true
					} else {
						totalRequestsMem.Add(mem)
					}
				}
			}
			if c.Resources.Limits != nil {
				if cpu, ok := c.Resources.Limits[corev1.ResourceCPU]; ok {
					if !hasLimitsCPU {
						totalLimitsCPU = cpu.DeepCopy()
						hasLimitsCPU = true
					} else {
						totalLimitsCPU.Add(cpu)
					}
				}
				if mem, ok := c.Resources.Limits[corev1.ResourceMemory]; ok {
					if !hasLimitsMem {
						totalLimitsMem = mem.DeepCopy()
						hasLimitsMem = true
					} else {
						totalLimitsMem.Add(mem)
					}
				}
			}
		}
		for _, v := range i.Spec.Template.Spec.Volumes {
			if v.PersistentVolumeClaim != nil {
				pvcs = append(pvcs, v.PersistentVolumeClaim.ClaimName)
			}
		}

		var replicas int32
		if i.Spec.Replicas != nil {
			replicas = *i.Spec.Replicas
		}

		// Extract tag from first image (format: registry/repo/image:tag or image:tag)
		var imageTag string
		if len(images) > 0 && images[0] != "" {
			image := images[0]
			// Check if image has SHA256 digest (format: image@sha256:hash)
			if idx := strings.Index(image, "@sha256:"); idx != -1 {
				imageTag = image[idx+8:idx+16] + "..." // Show first 8 chars of SHA
			} else if idx := strings.LastIndex(image, ":"); idx != -1 {
				imageTag = image[idx+1:]
			} else {
				imageTag = "latest"
			}
		}

		// Format requests and limits as strings
		var requestsCPU, requestsMem, limitsCPU, limitsMem string
		if hasRequestsCPU {
			requestsCPU = totalRequestsCPU.String()
		}
		if hasRequestsMem {
			requestsMem = totalRequestsMem.String()
		}
		if hasLimitsCPU {
			limitsCPU = totalLimitsCPU.String()
		}
		if hasLimitsMem {
			limitsMem = totalLimitsMem.String()
		}

		details := models.DeploymentDetails{
			Replicas:    replicas,
			Ready:       i.Status.ReadyReplicas,
			Images:      images,
			ImageTag:    imageTag,
			Ports:       ports,
			PVCs:        pvcs,
			PodLabels:   i.Spec.Selector.MatchLabels,
			Labels:      i.Labels,
			RequestsCPU: requestsCPU,
			RequestsMem: requestsMem,
			LimitsCPU:   limitsCPU,
			LimitsMem:   limitsMem,
		}

		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "Deployment",
			Status:    fmt.Sprintf("%d/%d", i.Status.ReadyReplicas, replicas),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listStatefulSets(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.AppsV1().StatefulSets(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		details := map[string]interface{}{
			"replicas":       i.Status.Replicas,
			"ready":          i.Status.ReadyReplicas,
			"current":        i.Status.CurrentReplicas,
			"update":         i.Status.UpdatedReplicas,
			"serviceName":    i.Spec.ServiceName,
			"podManagement":  i.Spec.PodManagementPolicy,
			"updateStrategy": i.Spec.UpdateStrategy,
			"volumeClaims":   i.Spec.VolumeClaimTemplates,
			"selector":       i.Spec.Selector.MatchLabels,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "StatefulSet",
			Status:    fmt.Sprintf("%d/%d", i.Status.ReadyReplicas, i.Status.Replicas),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listDaemonSets(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.AppsV1().DaemonSets(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		details := map[string]interface{}{
			"desired":      i.Status.DesiredNumberScheduled,
			"current":      i.Status.CurrentNumberScheduled,
			"ready":        i.Status.NumberReady,
			"available":    i.Status.NumberAvailable,
			"updated":      i.Status.UpdatedNumberScheduled,
			"misscheduled": i.Status.NumberMisscheduled,
			"nodeSelector": i.Spec.Template.Spec.NodeSelector,
			"selector":     i.Spec.Selector.MatchLabels,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "DaemonSet",
			Status:    fmt.Sprintf("%d/%d", i.Status.NumberReady, i.Status.DesiredNumberScheduled),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listHPAs(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		status := fmt.Sprintf("%d/%d replicas", i.Status.CurrentReplicas, i.Status.DesiredReplicas)
		details := map[string]interface{}{
			"minReplicas":   i.Spec.MinReplicas,
			"maxReplicas":   i.Spec.MaxReplicas,
			"current":       i.Status.CurrentReplicas,
			"desired":       i.Status.DesiredReplicas,
			"metrics":       i.Spec.Metrics,
			"lastScaleTime": i.Status.LastScaleTime,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "HPA",
			Status:    status,
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}
