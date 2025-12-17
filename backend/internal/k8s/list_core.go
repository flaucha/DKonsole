package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func (s *ResourceListService) listNodes(ctx context.Context, client kubernetes.Interface, metricsClient metricsv.Interface, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().Nodes().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Build metrics map
	metricsMap := make(map[string]models.NodeUsage)
	if metricsClient != nil {
		if nmList, mErr := metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, opts); mErr == nil {
			for _, nm := range nmList.Items {
				cpuMilli := nm.Usage.Cpu().MilliValue()
				memBytes := nm.Usage.Memory().Value()
				metricsMap[nm.Name] = models.NodeUsage{
					CPU:    fmt.Sprintf("%dm", cpuMilli),
					Memory: fmt.Sprintf("%.1fMi", float64(memBytes)/(1024*1024)),
				}
			}
		}
	}

	var resources []models.Resource
	for _, i := range list.Items {
		conditions := make(map[string]string)
		for _, c := range i.Status.Conditions {
			conditions[string(c.Type)] = string(c.Status)
		}

		var images []string
		for _, img := range i.Status.Images {
			if len(img.Names) > 0 {
				images = append(images, img.Names[0])
			}
		}

		details := map[string]interface{}{
			"addresses":     i.Status.Addresses,
			"nodeInfo":      i.Status.NodeInfo,
			"capacity":      i.Status.Capacity,
			"allocatable":   i.Status.Allocatable,
			"conditions":    conditions,
			"images":        images,
			"taints":        i.Spec.Taints,
			"podCIDR":       i.Spec.PodCIDR,
			"podCIDRs":      i.Spec.PodCIDRs,
			"unschedulable": i.Spec.Unschedulable,
			"labels":        i.Labels,
			"annotations":   i.Annotations,
			"metrics":       metricsMap[i.Name],
		}

		status := "Ready"
		for _, c := range i.Status.Conditions {
			if c.Type == corev1.NodeReady && c.Status != corev1.ConditionTrue {
				status = "NotReady"
			}
		}
		if i.Spec.Unschedulable {
			status += ",SchedulingDisabled"
		}

		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: "",
			Kind:      "Node",
			Status:    status,
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details:   details,
		})
	}

	return resources, nil
}

func (s *ResourceListService) listPods(ctx context.Context, client kubernetes.Interface, metricsClient metricsv.Interface, namespace string, opts metav1.ListOptions) ([]models.Resource, error) {
	list, err := client.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Build metrics map
	metricsMap := make(map[string]models.PodMetric)
	if metricsClient != nil {
		if pmList, mErr := metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, opts); mErr == nil {
			for _, pm := range pmList.Items {
				var cpuMilli int64
				var memBytes int64
				for _, c := range pm.Containers {
					cpuMilli += c.Usage.Cpu().MilliValue()
					memBytes += c.Usage.Memory().Value()
				}
				metricsMap[pm.Name] = models.PodMetric{
					CPU:    fmt.Sprintf("%dm", cpuMilli),
					Memory: fmt.Sprintf("%.1fMi", float64(memBytes)/(1024*1024)),
				}
			}
		}
	}

	var resources []models.Resource
	for _, i := range list.Items {
		var containers []string

		var initContainerNames []string
		// Include InitContainers in the list of containers
		for _, c := range i.Spec.InitContainers {
			containers = append(containers, c.Name)
			initContainerNames = append(initContainerNames, c.Name)
		}
		for _, c := range i.Spec.Containers {
			containers = append(containers, c.Name)
		}

		restarts := int32(0)
		readyCount := int32(0)
		totalContainers := int32(len(i.Spec.Containers) + len(i.Spec.InitContainers))
		var containerStatuses []map[string]interface{}

		// processStatus is a helper to process container statuses
		processStatus := func(statuses []corev1.ContainerStatus) {
			for _, cs := range statuses {
				restarts += cs.RestartCount
				if cs.Ready {
					readyCount++
				}

				containerStatus := map[string]interface{}{
					"name":         cs.Name,
					"ready":        cs.Ready,
					"restartCount": cs.RestartCount,
					"image":        cs.Image,
				}

				if cs.State.Waiting != nil {
					containerStatus["state"] = "Waiting"
					containerStatus["reason"] = cs.State.Waiting.Reason
					containerStatus["message"] = cs.State.Waiting.Message
				} else if cs.State.Running != nil {
					containerStatus["state"] = "Running"
					containerStatus["startedAt"] = cs.State.Running.StartedAt.Format(time.RFC3339)
				} else if cs.State.Terminated != nil {
					containerStatus["state"] = "Terminated"
					containerStatus["reason"] = cs.State.Terminated.Reason
					containerStatus["exitCode"] = cs.State.Terminated.ExitCode
					if !cs.State.Terminated.StartedAt.IsZero() {
						containerStatus["startedAt"] = cs.State.Terminated.StartedAt.Format(time.RFC3339)
					}
					if !cs.State.Terminated.FinishedAt.IsZero() {
						containerStatus["finishedAt"] = cs.State.Terminated.FinishedAt.Format(time.RFC3339)
					}
				}

				containerStatuses = append(containerStatuses, containerStatus)
			}
		}

		// Process init container statuses first
		processStatus(i.Status.InitContainerStatuses)
		// Process regular container statuses
		processStatus(i.Status.ContainerStatuses)

		resources = append(resources, models.Resource{
			Name:      i.Name,
			Namespace: i.Namespace,
			Kind:      "Pod",
			Status:    string(i.Status.Phase),
			Created:   i.CreationTimestamp.Format(time.RFC3339),
			UID:       string(i.UID),
			Details: map[string]interface{}{
				"node":              i.Spec.NodeName,
				"ip":                i.Status.PodIP,
				"restarts":          restarts,
				"ready":             fmt.Sprintf("%d/%d", readyCount, totalContainers),
				"readyCount":        readyCount,
				"totalContainers":   totalContainers,
				"containers":        containers,
				"initContainers":    initContainerNames,
				"containerStatuses": containerStatuses,
				"metrics":           metricsMap[i.Name],
				"labels":            i.Labels,
			},
		})
	}

	return resources, nil
}
