package k8s

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestListPVCs(t *testing.T) {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pvc-1",
			Namespace: "default",
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimBound,
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("1Gi"),
			},
		},
	}

	client := k8sfake.NewClientset(pvc)
	service := NewResourceListService(nil, "")

	resources, err := service.listPVCs(context.Background(), client, "default", metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listPVCs failed: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 PVC, got %d", len(resources))
	}

	r := resources[0]
	if r.Name != "pvc-1" {
		t.Errorf("expected name pvc-1, got %s", r.Name)
	}
	if r.Status != "Bound" {
		t.Errorf("expected status Bound, got %s", r.Status)
	}
}
