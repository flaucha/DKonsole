package k8s

import (
	"context"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func ctxWithPermissions(perms map[string]string) context.Context {
	return context.WithValue(context.Background(), auth.UserContextKey(), &models.Claims{
		Role:        "user",
		Permissions: perms,
	})
}

func TestResourceService_GetResourceYAML(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	obj := &unstructured.Unstructured{}
	obj.SetName("demo")
	obj.SetNamespace("default")
	obj.SetKind("Deployment")
	obj.SetAPIVersion("apps/v1")
	obj.Object["spec"] = map[string]interface{}{"replicas": int64(1)}
	obj.Object["metadata"] = map[string]interface{}{
		"managedFields": []interface{}{"noise"},
		"uid":           "abc",
	}

	resolver := &fakeGVRResolver{gvr: gvr, meta: models.ResourceMeta{Namespaced: true}}
	repo := &fakeResourceRepo{getObj: obj}
	svc := NewResourceService(repo, resolver)

	req := GetResourceRequest{
		Kind:       "Deployment",
		Name:       "demo",
		Namespace:  "default",
		Namespaced: true,
	}

	yamlStr, err := svc.GetResourceYAML(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if yamlStr == "" {
		t.Fatalf("expected yaml output")
	}
	if contains := strings.Contains(yamlStr, "managedFields"); contains {
		t.Fatalf("expected managedFields removed, got %s", yamlStr)
	}
}

func TestResourceService_UpdateResource_Success(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	resolver := &fakeGVRResolver{gvr: gvr, meta: models.ResourceMeta{Namespaced: true}}
	repo := &fakeResourceRepo{}
	svc := NewResourceService(repo, resolver)

	ctx := ctxWithPermissions(map[string]string{"default": "edit"})
	req := UpdateResourceRequest{
		YAMLContent: "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: my-dep\n  namespace: default\nspec: {}",
		Kind:        "Deployment",
		Namespace:   "default",
		Namespaced:  true,
	}

	if err := svc.UpdateResource(ctx, req); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if repo.patchCalls == 0 {
		t.Fatalf("expected Patch to be called")
	}
	if repo.patchName != "my-dep" {
		t.Errorf("expected patch name 'my-dep', got %q", repo.patchName)
	}
	if repo.patchNamespace != "default" || !repo.patchNamespaced {
		t.Errorf("expected namespaced patch to default, got namespace=%q namespaced=%v", repo.patchNamespace, repo.patchNamespaced)
	}
	if repo.patchType != types.ApplyPatchType {
		t.Errorf("expected apply patch type, got %v", repo.patchType)
	}
}

func TestResourceService_UpdateResource_InvalidYAML(t *testing.T) {
	resolver := &fakeGVRResolver{}
	repo := &fakeResourceRepo{}
	svc := NewResourceService(repo, resolver)

	ctx := ctxWithPermissions(map[string]string{"default": "edit"})
	req := UpdateResourceRequest{
		YAMLContent: "kind: Deployment\nmetadata: : :",
		Kind:        "Deployment",
		Namespace:   "default",
		Namespaced:  true,
	}

	err := svc.UpdateResource(ctx, req)
	if err == nil || !strings.Contains(err.Error(), "invalid YAML") {
		t.Fatalf("expected invalid YAML error, got: %v", err)
	}
	if repo.patchCalls > 0 {
		t.Fatalf("expected no patch call on invalid YAML")
	}
}

func TestResourceService_UpdateResource_PermissionDenied(t *testing.T) {
	resolver := &fakeGVRResolver{meta: models.ResourceMeta{Namespaced: true}}
	repo := &fakeResourceRepo{}
	svc := NewResourceService(repo, resolver)

	ctx := ctxWithPermissions(map[string]string{})
	req := UpdateResourceRequest{
		YAMLContent: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm\n  namespace: default\ndata:\n  k: v",
		Kind:        "ConfigMap",
		Namespace:   "default",
		Namespaced:  true,
	}

	err := svc.UpdateResource(ctx, req)
	if err == nil || !strings.Contains(err.Error(), "failed to check permissions") {
		t.Fatalf("expected permission error, got: %v", err)
	}
	if repo.patchCalls > 0 {
		t.Fatalf("expected no patch call on permission failure")
	}
}

func TestResourceService_DeleteResource_RequiresNamespace(t *testing.T) {
	resolver := &fakeGVRResolver{meta: models.ResourceMeta{Namespaced: true}}
	repo := &fakeResourceRepo{}
	svc := NewResourceService(repo, resolver)

	ctx := ctxWithPermissions(map[string]string{"default": "edit"})
	err := svc.DeleteResource(ctx, DeleteResourceRequest{
		Kind: "Deployment",
		Name: "dep",
		// Namespace intentionally empty
	})

	if err == nil || !strings.Contains(err.Error(), "namespace is required") {
		t.Fatalf("expected namespace error, got: %v", err)
	}
	if repo.deleteCalled {
		t.Fatalf("expected no delete call without namespace")
	}
}

func TestResourceService_DeleteResource_Success(t *testing.T) {
	resolver := &fakeGVRResolver{gvr: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, meta: models.ResourceMeta{Namespaced: true}}
	repo := &fakeResourceRepo{}
	svc := NewResourceService(repo, resolver)

	ctx := ctxWithPermissions(map[string]string{"default": "edit"})
	err := svc.DeleteResource(ctx, DeleteResourceRequest{
		Kind:      "Deployment",
		Name:      "dep",
		Namespace: "default",
		Force:     true,
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if !repo.deleteCalled {
		t.Fatalf("expected delete to be called")
	}
	if repo.deleteNamespace != "default" || !repo.deleteNamespaced {
		t.Errorf("expected namespaced delete to default, got namespace=%q namespaced=%v", repo.deleteNamespace, repo.deleteNamespaced)
	}
	if repo.deleteOptions.PropagationPolicy == nil || *repo.deleteOptions.PropagationPolicy != metav1.DeletePropagationBackground {
		t.Errorf("expected background propagation, got %v", repo.deleteOptions.PropagationPolicy)
	}
	if repo.deleteOptions.GracePeriodSeconds == nil || *repo.deleteOptions.GracePeriodSeconds != 0 {
		t.Errorf("expected zero grace period, got %v", repo.deleteOptions.GracePeriodSeconds)
	}
}

func TestResourceService_CreateResource_DefaultNamespace(t *testing.T) {
	resolver := &fakeGVRResolver{
		gvr:  schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
		meta: models.ResourceMeta{Namespaced: true},
	}
	repo := &fakeResourceRepo{
		createResult: &unstructured.Unstructured{Object: map[string]interface{}{"kind": "ConfigMap", "metadata": map[string]interface{}{"name": "cm", "namespace": "default"}}},
	}
	svc := NewResourceService(repo, resolver)

	ctx := context.Background()
	result, err := svc.CreateResource(ctx, "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm\nspec: {}")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if repo.createNamespace != "default" || !repo.createNamespaced {
		t.Errorf("expected create namespace default, got namespace=%q namespaced=%v", repo.createNamespace, repo.createNamespaced)
	}
	if repo.createObj.GetNamespace() != "default" {
		t.Errorf("expected object namespace default, got %q", repo.createObj.GetNamespace())
	}
	if objKind, ok := result.(map[string]interface{})["kind"]; !ok || objKind != "ConfigMap" {
		t.Fatalf("expected returned object kind ConfigMap, got %v", result)
	}
}
