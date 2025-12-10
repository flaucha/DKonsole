package k8s

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

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

// MockClusterService
type MockClusterService struct {
	mock.Mock
}

func (m *MockClusterService) GetClient(r *http.Request) (kubernetes.Interface, error) {
	args := m.Called(r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(kubernetes.Interface), args.Error(1)
}

func (m *MockClusterService) GetDynamicClient(r *http.Request) (dynamic.Interface, error) {
	args := m.Called(r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(dynamic.Interface), args.Error(1)
}

func (m *MockClusterService) GetMetricsClient(r *http.Request) *metricsv.Clientset {
	args := m.Called(r)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*metricsv.Clientset)
}

func (m *MockClusterService) GetRESTConfig(r *http.Request) (*rest.Config, error) {
	args := m.Called(r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rest.Config), args.Error(1)
}
