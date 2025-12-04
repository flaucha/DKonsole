package k8s

import (
	"context"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func newTestDeployment(name, namespace string, labels map[string]string, created metav1.Time) *appsv1.Deployment {
	replicas := int32(1)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			Labels:            labels,
			CreationTimestamp: created,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", Image: "demo"},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{ReadyReplicas: 1},
	}
}

func TestResourceListService_DeniesNamespaceWithoutAccess(t *testing.T) {
	client := k8sfake.NewClientset()
	service := NewResourceListService(nil, "")

	ctx := context.WithValue(context.Background(), auth.UserContextKey(), &auth.AuthClaims{
		Claims: models.Claims{
			Username:    "user",
			Permissions: map[string]string{"allowed": "view"},
		},
	})

	_, err := service.ListResources(ctx, ListResourcesRequest{
		Kind:      "Deployment",
		Namespace: "forbidden",
		Client:    client,
	})

	if err == nil || !strings.Contains(err.Error(), "access denied to namespace: forbidden") {
		t.Fatalf("expected namespace access error, got %v", err)
	}
}

func TestResourceListService_AllNamespacesFilteredAndSelector(t *testing.T) {
	now := metav1.Now()
	depAllowed := newTestDeployment("allowed-dep", "allowed", map[string]string{"app": "demo"}, now)
	depDenied := newTestDeployment("denied-dep", "denied", map[string]string{"app": "demo"}, now)
	depOtherLabel := newTestDeployment("other-label", "allowed", map[string]string{"app": "other"}, now)

	client := k8sfake.NewClientset(depAllowed, depDenied, depOtherLabel)
	service := NewResourceListService(nil, "")

	ctx := context.WithValue(context.Background(), auth.UserContextKey(), &auth.AuthClaims{
		Claims: models.Claims{
			Username:    "user",
			Permissions: map[string]string{"allowed": "view"},
		},
	})

	resources, err := service.ListResources(ctx, ListResourcesRequest{
		Kind:          "Deployment",
		Namespace:     "all",
		AllNamespaces: true,
		LabelSelector: "app=demo",
		Client:        client,
	})
	if err != nil {
		t.Fatalf("ListResources returned error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource after filtering, got %d", len(resources))
	}
	if resources[0].Namespace != "allowed" || resources[0].Name != "allowed-dep" {
		t.Fatalf("unexpected resource returned: %+v", resources[0])
	}
}

func TestResourceListService_FailsWithoutUserContext(t *testing.T) {
	client := k8sfake.NewClientset()
	service := NewResourceListService(nil, "")

	_, err := service.ListResources(context.Background(), ListResourcesRequest{
		Kind:   "Deployment",
		Client: client,
	})
	if err == nil || !strings.Contains(err.Error(), "failed to get user permissions") {
		t.Fatalf("expected missing user error, got %v", err)
	}
}
