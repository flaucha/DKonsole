package k8s

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	metricsapi "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned/fake"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func TestResourceListService_ListPods_WithMetrics(t *testing.T) {
	// Create Pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "c1"}},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	// Create Metrics
	cpu := resource.MustParse("100m")
	mem := resource.MustParse("100Mi")

	pm := &metricsapi.PodMetrics{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodMetrics",
			APIVersion: "metrics.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Containers: []metricsapi.ContainerMetrics{
			{
				Name: "c1",
				Usage: corev1.ResourceList{
					corev1.ResourceCPU:    cpu,
					corev1.ResourceMemory: mem,
				},
			},
		},
	}

	client := k8sfake.NewSimpleClientset(pod)
	metricsClient := metricsv.NewSimpleClientset()

	// Prepend Reactor to return metrics
	metricsClient.PrependReactor("list", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &metricsapi.PodMetricsList{
			Items: []metricsapi.PodMetrics{*pm},
		}, nil
	})

	svc := NewResourceListService(nil, "")

	// Test List
	list, err := svc.listPods(context.Background(), client, metricsClient, "default", metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listPods failed: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 pod, got %d", len(list))
	}

	res := list[0]
	details := res.Details.(map[string]interface{})

	// 'metrics' should be present and be models.PodMetric
	// Wait, list_core.go: Line 163: "metrics": metricsMap[i.Name]
	// metricsMap values are models.PodMetric (struct).

	metricsVal, ok := details["metrics"]
	if !ok {
		t.Errorf("metrics key missing in details")
	} else {
		metrics, ok := metricsVal.(models.PodMetric)
		if !ok {
			t.Errorf("metrics value is not models.PodMetric, got %T", metricsVal)
		} else {
			if metrics.CPU != "100m" {
				t.Errorf("expected CPU 100m, got %s", metrics.CPU)
			}
			// Memory formatting is "%.1fMi". 100Mi = 100*1024*1024 bytes.
			// 100*1024*1024 / (1024*1024) = 100.0.
			if metrics.Memory != "100.0Mi" {
				t.Errorf("expected Memory 100.0Mi, got %s", metrics.Memory)
			}
		}
	}
}
