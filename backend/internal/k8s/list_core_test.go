package k8s

import (
	"context"
	"testing"


	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestListPods(t *testing.T) {
	now := metav1.Now()
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pod-1",
			Namespace:         "default",
			CreationTimestamp: now,
			UID:               "uid-1",
			Labels:            map[string]string{"app": "demo"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "c1",
					Ready: true,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{
							StartedAt: now,
						},
					},
				},
			},
		},
		Spec: corev1.PodSpec{
			NodeName: "node-1",
			Containers: []corev1.Container{
				{Name: "c1", Image: "nginx"},
			},
		},
	}

	client := k8sfake.NewClientset(pod1)
	service := NewResourceListService(nil, "")

	// We can call listPods directly since it's an internal method we want to unit test
	// Normally we'd go through ListResources, but testing the transformation logic directly is also valid for these unit tests.
	// However, since they are unexported, we must be in the same package (k8s).
	// ListResources is the public entry point which calls listPods.

	// Let's test via ListResources to be safe and integration-style, 
	// or call the method directly from this test since it is in package k8s.
	
	resources, err := service.listPods(context.Background(), client, nil, "default", metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listPods failed: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 pod, got %d", len(resources))
	}

	r := resources[0]
	if r.Name != "pod-1" {
		t.Errorf("expected name pod-1, got %s", r.Name)
	}
	if r.Kind != "Pod" {
		t.Errorf("expected kind Pod, got %s", r.Kind)
	}
	if r.Status != "Running" {
		t.Errorf("expected status Running, got %s", r.Status)
	}

	details, ok := r.Details.(map[string]interface{})
	if !ok {
		t.Fatalf("details is not map[string]interface{}")
	}

	if details["node"] != "node-1" {
		t.Errorf("expected node node-1, got %v", details["node"])
	}
}

func TestListServices(t *testing.T) {
	svc1 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc-1",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: "10.0.0.1",
			Ports: []corev1.ServicePort{
				{Port: 80, Protocol: corev1.ProtocolTCP},
			},
		},
	}

	client := k8sfake.NewClientset(svc1)
	service := NewResourceListService(nil, "")

	resources, err := service.listServices(context.Background(), client, "default", metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listServices failed: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 service, got %d", len(resources))
	}
	
	r := resources[0]
	if r.Name != "svc-1" {
		t.Errorf("expected name svc-1, got %s", r.Name)
	}
}
