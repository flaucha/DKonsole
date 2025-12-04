package k8s

import (
	"context"
	"errors"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

type fakeGVRResolver struct {
	gvr        schema.GroupVersionResource
	meta       models.ResourceMeta
	err        error
	resolveLog []struct {
		kind       string
		api        string
		namespaced string
	}
}

func (f *fakeGVRResolver) ResolveGVR(_ context.Context, kind, apiVersion, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
	f.resolveLog = append(f.resolveLog, struct {
		kind       string
		api        string
		namespaced string
	}{kind: kind, api: apiVersion, namespaced: namespacedParam})
	if f.err != nil {
		return schema.GroupVersionResource{}, models.ResourceMeta{}, f.err
	}
	return f.gvr, f.meta, nil
}

type fakeResourceRepo struct {
	patchCalled     bool
	patchName       string
	patchNamespace  string
	patchNamespaced bool
	patchData       []byte
	patchType       types.PatchType

	deleteCalled     bool
	deleteOptions    metav1.DeleteOptions
	deleteNamespace  string
	deleteNamespaced bool
	deleteName       string

	createNamespace  string
	createNamespaced bool
	createObj        *unstructured.Unstructured
	createResult     *unstructured.Unstructured

	err error
}

func (f *fakeResourceRepo) Get(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error) {
	return nil, errors.New("not implemented: Get")
}

func (f *fakeResourceRepo) Create(_ context.Context, _ schema.GroupVersionResource, namespace string, namespaced bool, obj *unstructured.Unstructured, _ metav1.CreateOptions) (*unstructured.Unstructured, error) {
	f.createNamespace = namespace
	f.createNamespaced = namespaced
	f.createObj = obj
	if f.err != nil {
		return nil, f.err
	}
	if f.createResult != nil {
		return f.createResult, nil
	}
	return obj, nil
}

func (f *fakeResourceRepo) Patch(_ context.Context, _ schema.GroupVersionResource, name, namespace string, namespaced bool, patchData []byte, patchType types.PatchType, _ metav1.PatchOptions) (*unstructured.Unstructured, error) {
	f.patchCalled = true
	f.patchName = name
	f.patchNamespace = namespace
	f.patchNamespaced = namespaced
	f.patchData = patchData
	f.patchType = patchType
	if f.err != nil {
		return nil, f.err
	}
	return &unstructured.Unstructured{}, nil
}

func (f *fakeResourceRepo) Delete(_ context.Context, _ schema.GroupVersionResource, name, namespace string, namespaced bool, options metav1.DeleteOptions) error {
	f.deleteCalled = true
	f.deleteName = name
	f.deleteNamespace = namespace
	f.deleteNamespaced = namespaced
	f.deleteOptions = options
	return f.err
}

func ctxWithPermissions(perms map[string]string, role string) context.Context {
	return context.WithValue(context.Background(), auth.UserContextKey(), &models.Claims{
		Role:        role,
		Permissions: perms,
	})
}

func TestResourceService_UpdateResource_Success(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	resolver := &fakeGVRResolver{gvr: gvr, meta: models.ResourceMeta{Namespaced: true}}
	repo := &fakeResourceRepo{}
	svc := NewResourceService(repo, resolver)

	ctx := ctxWithPermissions(map[string]string{"default": "edit"}, "")
	req := UpdateResourceRequest{
		YAMLContent: "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: my-dep\n  namespace: default\nspec: {}",
		Kind:        "Deployment",
		Namespace:   "default",
		Namespaced:  true,
	}

	if err := svc.UpdateResource(ctx, req); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if !repo.patchCalled {
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

	ctx := ctxWithPermissions(map[string]string{"default": "edit"}, "")
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
	if repo.patchCalled {
		t.Fatalf("expected no patch call on invalid YAML")
	}
}

func TestResourceService_UpdateResource_PermissionDenied(t *testing.T) {
	resolver := &fakeGVRResolver{meta: models.ResourceMeta{Namespaced: true}}
	repo := &fakeResourceRepo{}
	svc := NewResourceService(repo, resolver)

	ctx := ctxWithPermissions(map[string]string{}, "")
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
	if repo.patchCalled {
		t.Fatalf("expected no patch call on permission failure")
	}
}

func TestResourceService_DeleteResource_RequiresNamespace(t *testing.T) {
	resolver := &fakeGVRResolver{meta: models.ResourceMeta{Namespaced: true}}
	repo := &fakeResourceRepo{}
	svc := NewResourceService(repo, resolver)

	ctx := ctxWithPermissions(map[string]string{"default": "edit"}, "")
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

	ctx := ctxWithPermissions(map[string]string{"default": "edit"}, "")
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
