package k8s

import (
	"context"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestListRoles(t *testing.T) {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "role-1",
			Namespace: "default",
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "list"},
				Resources: []string{"pods"},
			},
		},
	}

	client := k8sfake.NewClientset(role)
	service := NewResourceListService(nil, "")

	resources, err := service.listRoles(context.Background(), client, "default", metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listRoles failed: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 role, got %d", len(resources))
	}

	r := resources[0]
	if r.Name != "role-1" {
		t.Errorf("expected name role-1, got %s", r.Name)
	}
}
