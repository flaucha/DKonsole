package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

type mockResourceRepository struct {
	patchCalls int
	patchErr   error
}

func (m *mockResourceRepository) Get(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockResourceRepository) Create(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, obj *unstructured.Unstructured, options metav1.CreateOptions) (*unstructured.Unstructured, error) {
	return obj, nil
}

func (m *mockResourceRepository) Patch(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, patchData []byte, patchType types.PatchType, options metav1.PatchOptions) (*unstructured.Unstructured, error) {
	m.patchCalls++
	if m.patchErr != nil {
		return nil, m.patchErr
	}
	return &unstructured.Unstructured{}, nil
}

func (m *mockResourceRepository) Delete(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, options metav1.DeleteOptions) error {
	return nil
}

type mockGVRResolver struct {
	resolveFunc func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error)
}

func (m *mockGVRResolver) ResolveGVR(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
	if m.resolveFunc != nil {
		return m.resolveFunc(ctx, kind, apiVersion, namespacedParam)
	}
	return schema.GroupVersionResource{}, models.ResourceMeta{}, nil
}
