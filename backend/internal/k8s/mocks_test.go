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
	patchCalls      int
	patchErr        error
	patchName       string
	patchNamespace  string
	patchNamespaced bool
	patchData       []byte
	patchType       types.PatchType

	createObj        *unstructured.Unstructured
	createNamespace  string
	createNamespaced bool
	createResult     *unstructured.Unstructured

	deleteCalled     bool
	deleteName       string
	deleteNamespace  string
	deleteNamespaced bool
	deleteOptions    metav1.DeleteOptions

	getObj     *unstructured.Unstructured
	getErr     error
	getReturns *unstructured.Unstructured // Explicit return for Get

	// Generic error for all methods if set and specific error not set
	err error
}

func (m *mockResourceRepository) Get(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.err != nil {
		return nil, m.err
	}
	if m.getObj != nil {
		return m.getObj.DeepCopy(), nil
	}
	// Fallback to empty if nothing configured, or maybe return nil/error?
	// Preserving behavior: return nil if not set
	return nil, nil
}

func (m *mockResourceRepository) Create(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, obj *unstructured.Unstructured, options metav1.CreateOptions) (*unstructured.Unstructured, error) {
	m.createNamespace = namespace
	m.createNamespaced = namespaced
	m.createObj = obj

	if m.err != nil {
		return nil, m.err
	}
	if m.createResult != nil {
		return m.createResult, nil
	}
	return obj, nil
}

func (m *mockResourceRepository) Patch(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, patchData []byte, patchType types.PatchType, options metav1.PatchOptions) (*unstructured.Unstructured, error) {
	m.patchCalls++
	m.patchName = name
	m.patchNamespace = namespace
	m.patchNamespaced = namespaced
	m.patchData = patchData
	m.patchType = patchType

	if m.patchErr != nil {
		return nil, m.patchErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return &unstructured.Unstructured{}, nil
}

func (m *mockResourceRepository) Delete(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, options metav1.DeleteOptions) error {
	m.deleteCalled = true
	m.deleteName = name
	m.deleteNamespace = namespace
	m.deleteNamespaced = namespaced
	m.deleteOptions = options

	if m.err != nil {
		return m.err
	}
	return nil
}

type fakeResourceRepo = mockResourceRepository

type mockGVRResolver struct {
	resolveFunc func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error)
}

func (m *mockGVRResolver) ResolveGVR(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
	if m.resolveFunc != nil {
		return m.resolveFunc(ctx, kind, apiVersion, namespacedParam)
	}
	return schema.GroupVersionResource{}, models.ResourceMeta{}, nil
}

// fakeGVRResolver for use in tests where we need struct fields
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

// mockServiceFactory implements Factory interface for testing
type mockServiceFactory struct {
	realFactory      *ServiceFactory
	namespaceService *NamespaceService
	clusterStatsSvc  *ClusterStatsService
	deploymentSvc    *DeploymentService
	resourceSvc      *ResourceService
	importSvc        *ImportService
	watchSvc         *WatchService
}

func newMockServiceFactory() *mockServiceFactory {
	return &mockServiceFactory{realFactory: NewServiceFactory()}
}

func (f *mockServiceFactory) CreateResourceService(dynamicClient dynamic.Interface, client kubernetes.Interface) *ResourceService {
	if f.resourceSvc != nil {
		return f.resourceSvc
	}
	return f.realFactory.CreateResourceService(dynamicClient, client)
}

func (f *mockServiceFactory) CreateImportService(dynamicClient dynamic.Interface, client kubernetes.Interface) *ImportService {
	if f.importSvc != nil {
		return f.importSvc
	}
	return f.realFactory.CreateImportService(dynamicClient, client)
}

func (f *mockServiceFactory) CreateNamespaceService(client kubernetes.Interface) *NamespaceService {
	if f.namespaceService != nil {
		return f.namespaceService
	}
	return f.realFactory.CreateNamespaceService(client)
}

func (f *mockServiceFactory) CreateClusterStatsService(client kubernetes.Interface) *ClusterStatsService {
	if f.clusterStatsSvc != nil {
		return f.clusterStatsSvc
	}
	return f.realFactory.CreateClusterStatsService(client)
}

func (f *mockServiceFactory) CreateDeploymentService(client kubernetes.Interface) *DeploymentService {
	if f.deploymentSvc != nil {
		return f.deploymentSvc
	}
	return f.realFactory.CreateDeploymentService(client)
}

func (f *mockServiceFactory) CreateCronJobService(client kubernetes.Interface) *CronJobService {
	return f.realFactory.CreateCronJobService(client)
}

func (f *mockServiceFactory) CreateWatchService() *WatchService {
	if f.watchSvc != nil {
		return f.watchSvc
	}
	return f.realFactory.CreateWatchService()
}
