package k8s

import (
	"context"
	"testing"


	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func TestListDeployments(t *testing.T) {
	replicas := int32(2)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dep-1",
			Namespace: "default",
			Labels:    map[string]string{"app": "demo"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "demo"},
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", Image: "nginx:latest"},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 2,
		},
	}

	client := k8sfake.NewClientset(dep)
	service := NewResourceListService(nil, "")

	resources, err := service.listDeployments(context.Background(), client, "default", metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listDeployments failed: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 deployment, got %d", len(resources))
	}

	r := resources[0]
	if r.Name != "dep-1" {
		t.Errorf("expected name dep-1, got %s", r.Name)
	}
	if r.Status != "2/2" {
		t.Errorf("expected status 2/2, got %s", r.Status)
	}

	details, ok := r.Details.(models.DeploymentDetails)
	if !ok {
		t.Fatalf("details is not models.DeploymentDetails")
	}

	if len(details.Images) != 1 || details.Images[0] != "nginx:latest" {
		t.Errorf("unexpected images: %v", details.Images)
	}
}
