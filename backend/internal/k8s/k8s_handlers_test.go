package k8s

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

type mockServiceFactory struct {
	realFactory      *ServiceFactory
	namespaceService *NamespaceService
	clusterStatsSvc  *ClusterStatsService
	deploymentSvc    *DeploymentService
}

func newMockServiceFactory() *mockServiceFactory {
	return &mockServiceFactory{realFactory: NewServiceFactory()}
}

func (f *mockServiceFactory) CreateResourceService(dynamicClient dynamic.Interface) *ResourceService {
	return f.realFactory.CreateResourceService(dynamicClient)
}

func (f *mockServiceFactory) CreateImportService(dynamicClient dynamic.Interface, client kubernetes.Interface) *ImportService {
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
	return f.realFactory.CreateWatchService()
}

type stubDeploymentRepository struct {
	replicas  int32
	getErr    error
	updateErr error
}

func (s *stubDeploymentRepository) GetScale(ctx context.Context, namespace, name string) (*autoscalingv1.Scale, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return &autoscalingv1.Scale{Spec: autoscalingv1.ScaleSpec{Replicas: s.replicas}}, nil
}

func (s *stubDeploymentRepository) UpdateScale(ctx context.Context, namespace, name string, scale *autoscalingv1.Scale) (*autoscalingv1.Scale, error) {
	if s.updateErr != nil {
		return nil, s.updateErr
	}
	return scale, nil
}

func (s *stubDeploymentRepository) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return &appsv1.Deployment{}, nil
}

func (s *stubDeploymentRepository) UpdateDeployment(ctx context.Context, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return deployment, nil
}

type stubClusterStatsRepository struct {
	stats models.ClusterStats
	err   error
}

func (s *stubClusterStatsRepository) GetNodeCount(ctx context.Context) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	return s.stats.Nodes, nil
}

func (s *stubClusterStatsRepository) GetNamespaceCount(ctx context.Context) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	return s.stats.Namespaces, nil
}

func (s *stubClusterStatsRepository) GetPodCount(ctx context.Context) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	return s.stats.Pods, nil
}

func (s *stubClusterStatsRepository) GetDeploymentCount(ctx context.Context) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	return s.stats.Deployments, nil
}

func (s *stubClusterStatsRepository) GetServiceCount(ctx context.Context) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	return s.stats.Services, nil
}

func (s *stubClusterStatsRepository) GetIngressCount(ctx context.Context) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	return s.stats.Ingresses, nil
}

func (s *stubClusterStatsRepository) GetPVCCount(ctx context.Context) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	return s.stats.PVCs, nil
}

func (s *stubClusterStatsRepository) GetPVCount(ctx context.Context) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	return s.stats.PVs, nil
}

func withUser(req *http.Request, claims models.Claims) *http.Request {
	ctx := context.WithValue(req.Context(), auth.UserContextKey(), &auth.AuthClaims{Claims: claims})
	return req.WithContext(ctx)
}

func newDeployment(name, namespace string, labels map[string]string, created metav1.Time) *appsv1.Deployment {
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

func TestGetNamespacesHandler_Success(t *testing.T) {
	client := k8sfake.NewClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}, Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "dev"}, Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive}},
	)
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": client},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := withUser(httptest.NewRequest(http.MethodGet, "/api/namespaces", nil), models.Claims{
		Username: "admin",
		Role:     "admin",
	})
	rr := httptest.NewRecorder()

	service.GetNamespaces(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var namespaces []models.Namespace
	if err := json.NewDecoder(rr.Body).Decode(&namespaces); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(namespaces) != 2 {
		t.Fatalf("expected 2 namespaces, got %d", len(namespaces))
	}
}

func TestGetNamespacesHandler_ClusterNotFound(t *testing.T) {
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := httptest.NewRequest(http.MethodGet, "/api/namespaces", nil)
	rr := httptest.NewRecorder()

	service.GetNamespaces(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "cluster not found") {
		t.Fatalf("expected cluster error message, got %s", rr.Body.String())
	}
}

func TestGetResourcesHandler_Success(t *testing.T) {
	now := metav1.Now()
	client := k8sfake.NewClientset(
		newDeployment("demo", "default", map[string]string{"app": "demo"}, now),
		newDeployment("other", "default", map[string]string{"app": "other"}, now),
	)
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": client},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := withUser(httptest.NewRequest(http.MethodGet, "/api/resources?kind=Deployment&namespace=default&labelSelector=app=demo", nil), models.Claims{
		Username: "admin",
		Role:     "admin",
	})
	rr := httptest.NewRecorder()

	service.GetResources(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resources []models.Resource
	if err := json.NewDecoder(rr.Body).Decode(&resources); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resources) != 1 || resources[0].Name != "demo" {
		t.Fatalf("expected filtered deployment, got %+v", resources)
	}
}

func TestGetResourcesHandler_AccessDenied(t *testing.T) {
	client := k8sfake.NewClientset()
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": client},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := withUser(httptest.NewRequest(http.MethodGet, "/api/resources?kind=Deployment&namespace=forbidden", nil), models.Claims{
		Username:    "user",
		Permissions: map[string]string{"allowed": "view"},
	})
	rr := httptest.NewRecorder()

	service.GetResources(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "access denied") {
		t.Fatalf("expected access denied message, got %s", rr.Body.String())
	}
}

func TestGetResourcesHandler_MissingKind(t *testing.T) {
	client := k8sfake.NewClientset()
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": client},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := httptest.NewRequest(http.MethodGet, "/api/resources?namespace=default", nil)
	rr := httptest.NewRecorder()

	service.GetResources(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestScaleResourceHandler_Success(t *testing.T) {
	client := k8sfake.NewClientset()
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": client},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	mockFactory := newMockServiceFactory()
	mockFactory.deploymentSvc = NewDeploymentService(&stubDeploymentRepository{replicas: 3})
	service.serviceFactory = mockFactory

	req := withUser(httptest.NewRequest(http.MethodPost, "/api/scale?kind=Deployment&name=demo&namespace=default&delta=2", nil), models.Claims{
		Username: "admin",
		Role:     "admin",
	})
	rr := httptest.NewRecorder()

	service.ScaleResource(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp map[string]int32
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["replicas"] != 5 {
		t.Fatalf("expected 5 replicas, got %d", resp["replicas"])
	}
}

func TestScaleResourceHandler_InvalidDelta(t *testing.T) {
	client := k8sfake.NewClientset()
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": client},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := withUser(httptest.NewRequest(http.MethodPost, "/api/scale?kind=Deployment&name=demo&namespace=default&delta=abc", nil), models.Claims{
		Username: "admin",
		Role:     "admin",
	})
	rr := httptest.NewRecorder()

	service.ScaleResource(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestScaleResourceHandler_Forbidden(t *testing.T) {
	client := k8sfake.NewClientset()
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": client},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := withUser(httptest.NewRequest(http.MethodPost, "/api/scale?kind=Deployment&name=demo&namespace=default&delta=1", nil), models.Claims{
		Username:    "viewer",
		Permissions: map[string]string{"default": "view"},
	})
	rr := httptest.NewRecorder()

	service.ScaleResource(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Edit permission required") {
		t.Fatalf("expected edit permission message, got %s", rr.Body.String())
	}
}

func TestGetClusterStatsHandler_Success(t *testing.T) {
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": k8sfake.NewClientset()},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	mockFactory := newMockServiceFactory()
	mockFactory.clusterStatsSvc = NewClusterStatsService(&stubClusterStatsRepository{
		stats: models.ClusterStats{
			Nodes:       1,
			Namespaces:  2,
			Pods:        3,
			Deployments: 4,
			Services:    5,
			Ingresses:   6,
			PVCs:        7,
			PVs:         8,
		},
	})
	service.serviceFactory = mockFactory

	req := httptest.NewRequest(http.MethodGet, "/api/cluster-stats", nil)
	rr := httptest.NewRecorder()

	service.GetClusterStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var stats models.ClusterStats
	if err := json.NewDecoder(rr.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if stats.Nodes != 1 || stats.PVs != 8 {
		t.Fatalf("unexpected stats payload: %+v", stats)
	}
}

func TestGetClusterStatsHandler_Error(t *testing.T) {
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": k8sfake.NewClientset()},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	mockFactory := newMockServiceFactory()
	mockFactory.clusterStatsSvc = NewClusterStatsService(&stubClusterStatsRepository{
		err: assertError{},
	})
	service.serviceFactory = mockFactory

	req := httptest.NewRequest(http.MethodGet, "/api/cluster-stats", nil)
	rr := httptest.NewRecorder()

	service.GetClusterStats(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}

type assertError struct{}

func (assertError) Error() string { return "boom" }
