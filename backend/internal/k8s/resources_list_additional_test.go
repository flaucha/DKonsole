package k8s

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestResourceListService_listNodes(t *testing.T) {
	now := metav1.Now()
	readyNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "ready-node",
			CreationTimestamp: now,
			UID:               "uid-1",
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
			},
		},
	}
	notReadyNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "not-ready",
			CreationTimestamp: now,
			UID:               "uid-2",
		},
		Spec: corev1.NodeSpec{
			Unschedulable: true,
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeReady, Status: corev1.ConditionFalse},
			},
		},
	}

	client := k8sfake.NewSimpleClientset(readyNode, notReadyNode)
	service := NewResourceListService(nil, "")

	resources, err := service.listNodes(context.Background(), client, nil, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listNodes returned error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(resources))
	}
	if resources[0].Name == resources[1].Name {
		t.Fatalf("node names should differ")
	}
}

func TestResourceListService_listPods(t *testing.T) {
	now := metav1.Now()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pod-1",
			Namespace:         "ns",
			UID:               "uid-1",
			CreationTimestamp: now,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "c1"},
			},
		},
		Status: corev1.PodStatus{
			Phase:     corev1.PodRunning,
			StartTime: &now,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "c1",
					Ready:        true,
					RestartCount: 1,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{
							StartedAt: metav1.NewTime(time.Now()),
						},
					},
				},
			},
		},
	}
	client := k8sfake.NewSimpleClientset(pod)
	service := NewResourceListService(nil, "")

	resources, err := service.listPods(context.Background(), client, nil, "ns", metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listPods returned error: %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("expected 1 pod, got %d", len(resources))
	}
	if resources[0].Name != "pod-1" || resources[0].Namespace != "ns" {
		t.Fatalf("unexpected pod resource: %+v", resources[0])
	}
	if resources[0].Details == nil {
		t.Fatalf("expected pod details to be populated")
	}
}

func TestResourceListService_listConfigMaps(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "config",
			Namespace: "ns",
			UID:       "uid-config",
		},
	}
	client := k8sfake.NewSimpleClientset(cm)
	service := NewResourceListService(nil, "")

	resources, err := service.listConfigMaps(context.Background(), client, "ns", metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listConfigMaps returned error: %v", err)
	}
	if len(resources) != 1 || resources[0].Name != "config" {
		t.Fatalf("unexpected configmaps: %+v", resources)
	}
}

func TestResourceListService_listServices(t *testing.T) {
	svcObj := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc",
			Namespace: "ns",
			UID:       "uid-svc",
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: "10.0.0.1",
		},
	}
	client := k8sfake.NewSimpleClientset(svcObj)
	service := NewResourceListService(nil, "")

	resources, err := service.listServices(context.Background(), client, "ns", metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listServices returned error: %v", err)
	}
	if len(resources) != 1 || resources[0].Name != "svc" {
		t.Fatalf("unexpected services: %+v", resources)
	}
}
