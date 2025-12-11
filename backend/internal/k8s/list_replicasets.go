package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func (s *ResourceListService) listReplicaSets(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.AppsV1().ReplicaSets(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		var images []string

		for _, c := range i.Spec.Template.Spec.Containers {
			images = append(images, c.Image)
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

		details := models.ReplicaSetDetails{
			Replicas:          replicas,
			Ready:             i.Status.ReadyReplicas,
			AvailableReplicas: i.Status.AvailableReplicas,
			Images:            images,
			ImageTag:          imageTag,
			Labels:            i.Labels,
			PodLabels:         i.Spec.Selector.MatchLabels,
		}

		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "ReplicaSet",
			Status:    fmt.Sprintf("%d/%d", i.Status.ReadyReplicas, replicas),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}
