package k8s

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func (s *ResourceListService) listJobs(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.BatchV1().Jobs(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		status := "Running"
		if i.Status.Succeeded > 0 {
			status = "Completed"
		} else if i.Status.Failed > 0 {
			status = "Failed"
		}
		details := map[string]interface{}{
			"active":         i.Status.Active,
			"succeeded":      i.Status.Succeeded,
			"failed":         i.Status.Failed,
			"startTime":      i.Status.StartTime,
			"completionTime": i.Status.CompletionTime,
			"parallelism":    i.Spec.Parallelism,
			"completions":    i.Spec.Completions,
			"backoffLimit":   i.Spec.BackoffLimit,
			"activeDeadline": i.Spec.ActiveDeadlineSeconds,
			"podSelector":    i.Spec.Selector,
			"podTemplate":    i.Spec.Template.Spec,
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "Job",
			Status:    status,
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listCronJobs(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.BatchV1().CronJobs(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	for _, i := range list.Items {
		var lastSchedule string
		if i.Status.LastScheduleTime != nil {
			lastSchedule = i.Status.LastScheduleTime.Format(time.RFC3339)
		}

		// Calculate status based on CronJob state
		status := "Active"
		if i.Spec.Suspend != nil && *i.Spec.Suspend {
			status = "Suspended"
		} else if len(i.Status.Active) > 0 {
			status = "Running"
		} else if i.Status.LastSuccessfulTime != nil {
			// Check if there's a recent failure by comparing LastScheduleTime with LastSuccessfulTime
			// If LastScheduleTime exists but LastSuccessfulTime is nil or older, it might have failed
			if i.Status.LastScheduleTime != nil {
				if i.Status.LastSuccessfulTime.Before(i.Status.LastScheduleTime) {
					// Last schedule was not successful
					status = "Failed"
				} else {
					status = "Succeeded"
				}
			} else {
				status = "Succeeded"
			}
		} else if i.Status.LastScheduleTime != nil {
			// Has been scheduled but no successful time recorded - likely failed
			status = "Failed"
		}

		details := map[string]interface{}{
			"schedule":         i.Spec.Schedule,
			"suspend":          i.Spec.Suspend,
			"concurrency":      i.Spec.ConcurrencyPolicy,
			"startingDeadline": i.Spec.StartingDeadlineSeconds,
			"lastSchedule":     lastSchedule,
			"succeeded":        len(i.Status.Active) == 0 && i.Status.LastSuccessfulTime != nil,
			"failed":           i.Status.LastScheduleTime != nil && (i.Status.LastSuccessfulTime == nil || i.Status.LastSuccessfulTime.Before(i.Status.LastScheduleTime)),
		}
		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "CronJob",
			Status:    status,
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}
