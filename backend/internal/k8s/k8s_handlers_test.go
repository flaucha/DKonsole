package k8s

import (
	"context"
	"encoding/json"
	"errors"
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

type stubNamespaceRepository struct {
	namespaces []corev1.Namespace
	err        error
}

func (s *stubNamespaceRepository) List(ctx context.Context) ([]corev1.Namespace, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.namespaces, nil
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

type stubRolloutDeploymentRepository struct {
	getErr    error
	updateErr error
}

func (s *stubRolloutDeploymentRepository) GetScale(ctx context.Context, namespace, name string) (*autoscalingv1.Scale, error) {
	return &autoscalingv1.Scale{}, nil
}

func (s *stubRolloutDeploymentRepository) UpdateScale(ctx context.Context, namespace, name string, scale *autoscalingv1.Scale) (*autoscalingv1.Scale, error) {
	return scale, nil
}

func (s *stubRolloutDeploymentRepository) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{},
		},
	}, nil
}

func (s *stubRolloutDeploymentRepository) UpdateDeployment(ctx context.Context, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	if s.updateErr != nil {
		return nil, s.updateErr
	}
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

func TestGetNamespacesHandler_ServiceError(t *testing.T) {
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": k8sfake.NewClientset()},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	mockFactory := newMockServiceFactory()
	mockFactory.namespaceService = NewNamespaceService(&stubNamespaceRepository{err: errors.New("boom")})
	service.serviceFactory = mockFactory

	req := withUser(httptest.NewRequest(http.MethodGet, "/api/namespaces", nil), models.Claims{
		Username: "ops",
		Role:     "admin",
	})
	rr := httptest.NewRecorder()

	service.GetNamespaces(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Failed to get namespaces") {
		t.Fatalf("expected namespace failure message, got %s", rr.Body.String())
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

func TestGetResourcesHandler_ClusterNotFound(t *testing.T) {
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := withUser(httptest.NewRequest(http.MethodGet, "/api/resources?kind=Deployment&cluster=ghost", nil), models.Claims{
		Username: "admin",
		Role:     "admin",
	})
	rr := httptest.NewRecorder()

	service.GetResources(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "cluster not found: ghost") {
		t.Fatalf("expected cluster not found message, got %s", rr.Body.String())
	}
}

func TestGetResourcesHandler_InternalError(t *testing.T) {
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": k8sfake.NewClientset()},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := httptest.NewRequest(http.MethodGet, "/api/resources?kind=Deployment&namespace=default", nil)
	rr := httptest.NewRecorder()

	service.GetResources(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Failed to list resources") {
		t.Fatalf("expected resources failure message, got %s", rr.Body.String())
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

func TestScaleResourceHandler_ValidationErrors(t *testing.T) {
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": k8sfake.NewClientset()},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	admin := models.Claims{Username: "admin", Role: "admin"}
	tests := []struct {
		name     string
		method   string
		url      string
		code     int
		contains string
	}{
		{
			name:     "method not allowed",
			method:   http.MethodGet,
			url:      "/api/scale?kind=Deployment&name=demo&namespace=default&delta=1",
			code:     http.StatusMethodNotAllowed,
			contains: "Method not allowed",
		},
		{
			name:     "unsupported kind",
			method:   http.MethodPost,
			url:      "/api/scale?kind=Pod&name=demo&namespace=default&delta=1",
			code:     http.StatusBadRequest,
			contains: "Scaling supported only for Deployments",
		},
		{
			name:     "missing name",
			method:   http.MethodPost,
			url:      "/api/scale?kind=Deployment&namespace=default&delta=1",
			code:     http.StatusBadRequest,
			contains: "Missing name",
		},
		{
			name:     "invalid name",
			method:   http.MethodPost,
			url:      "/api/scale?kind=Deployment&name=BadName&namespace=default&delta=1",
			code:     http.StatusBadRequest,
			contains: "invalid name",
		},
		{
			name:     "invalid namespace",
			method:   http.MethodPost,
			url:      "/api/scale?kind=Deployment&name=demo&namespace=BadNS&delta=1",
			code:     http.StatusBadRequest,
			contains: "invalid namespace",
		},
		{
			name:     "missing delta",
			method:   http.MethodPost,
			url:      "/api/scale?kind=Deployment&name=demo&namespace=default",
			code:     http.StatusBadRequest,
			contains: "Invalid delta",
		},
		{
			name:     "zero delta",
			method:   http.MethodPost,
			url:      "/api/scale?kind=Deployment&name=demo&namespace=default&delta=0",
			code:     http.StatusBadRequest,
			contains: "Delta cannot be zero",
		},
		{
			name:     "cluster not found",
			method:   http.MethodPost,
			url:      "/api/scale?cluster=ghost&kind=Deployment&name=demo&namespace=default&delta=1",
			code:     http.StatusBadRequest,
			contains: "cluster not found: ghost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := withUser(httptest.NewRequest(tt.method, tt.url, nil), admin)

			service.ScaleResource(rr, req)

			if rr.Code != tt.code {
				t.Fatalf("%s: expected status %d, got %d", tt.name, tt.code, rr.Code)
			}
			if tt.contains != "" && !strings.Contains(rr.Body.String(), tt.contains) {
				t.Fatalf("%s: expected body to contain %q, got %s", tt.name, tt.contains, rr.Body.String())
			}
		})
	}
}

func TestScaleResourceHandler_ServiceError(t *testing.T) {
	client := k8sfake.NewClientset()
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": client},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	mockFactory := newMockServiceFactory()
	mockFactory.deploymentSvc = NewDeploymentService(&stubDeploymentRepository{
		getErr: errors.New("cannot fetch scale"),
	})
	service.serviceFactory = mockFactory

	req := withUser(httptest.NewRequest(http.MethodPost, "/api/scale?kind=Deployment&name=demo&namespace=default&delta=1", nil), models.Claims{
		Username: "ops",
		Role:     "admin",
	})
	rr := httptest.NewRecorder()

	service.ScaleResource(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Failed to scale deployment") {
		t.Fatalf("expected scale failure message, got %s", rr.Body.String())
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

func TestGetClusterStatsHandler_ClusterNotFound(t *testing.T) {
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := httptest.NewRequest(http.MethodGet, "/api/cluster-stats?cluster=ghost", nil)
	rr := httptest.NewRecorder()

	service.GetClusterStats(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "cluster not found") {
		t.Fatalf("expected cluster not found message, got %s", rr.Body.String())
	}
}

func TestRolloutDeploymentHandler(t *testing.T) {
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": k8sfake.NewClientset()},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	mockFactory := newMockServiceFactory()
	mockFactory.deploymentSvc = NewDeploymentService(&stubRolloutDeploymentRepository{})
	service.serviceFactory = mockFactory

	admin := models.Claims{Username: "ops", Role: "admin"}

	t.Run("success", func(t *testing.T) {
		body := strings.NewReader(`{"namespace":"default","name":"demo"}`)
		req := withUser(httptest.NewRequest(http.MethodPost, "/api/rollout", body), admin)
		rr := httptest.NewRecorder()

		service.RolloutDeployment(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Deployment rollout triggered successfully") {
			t.Fatalf("unexpected response body: %s", rr.Body.String())
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := withUser(httptest.NewRequest(http.MethodGet, "/api/rollout", nil), admin)
		rr := httptest.NewRecorder()

		service.RolloutDeployment(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status 405, got %d", rr.Code)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		req := withUser(httptest.NewRequest(http.MethodPost, "/api/rollout", strings.NewReader("{invalid")), admin)
		rr := httptest.NewRecorder()

		service.RolloutDeployment(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Invalid request body") {
			t.Fatalf("expected invalid body message, got %s", rr.Body.String())
		}
	})

	t.Run("missing params", func(t *testing.T) {
		req := withUser(httptest.NewRequest(http.MethodPost, "/api/rollout", strings.NewReader(`{"namespace":""}`)), admin)
		rr := httptest.NewRecorder()

		service.RolloutDeployment(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Missing name or namespace") {
			t.Fatalf("expected missing params message, got %s", rr.Body.String())
		}
	})

	t.Run("invalid name", func(t *testing.T) {
		req := withUser(httptest.NewRequest(http.MethodPost, "/api/rollout", strings.NewReader(`{"namespace":"default","name":"BadName"}`)), admin)
		rr := httptest.NewRecorder()

		service.RolloutDeployment(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "invalid name") {
			t.Fatalf("expected invalid name message, got %s", rr.Body.String())
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		req := withUser(httptest.NewRequest(http.MethodPost, "/api/rollout", strings.NewReader(`{"namespace":"default","name":"demo"}`)), models.Claims{
			Username:    "viewer",
			Permissions: map[string]string{"default": "view"},
		})
		rr := httptest.NewRecorder()

		service.RolloutDeployment(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rr.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		mockFactory.deploymentSvc = NewDeploymentService(&stubRolloutDeploymentRepository{getErr: errors.New("boom")})
		service.serviceFactory = mockFactory

		req := withUser(httptest.NewRequest(http.MethodPost, "/api/rollout", strings.NewReader(`{"namespace":"default","name":"demo"}`)), admin)
		rr := httptest.NewRecorder()

		service.RolloutDeployment(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Failed to rollout deployment") {
			t.Fatalf("expected rollout failure message, got %s", rr.Body.String())
		}
	})

	t.Run("cluster not found", func(t *testing.T) {
		req := withUser(httptest.NewRequest(http.MethodPost, "/api/rollout?cluster=ghost", strings.NewReader(`{"namespace":"default","name":"demo"}`)), admin)
		rr := httptest.NewRecorder()

		service.RolloutDeployment(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rr.Code)
		}
	})
}

type assertError struct{}

func (assertError) Error() string { return "boom" }
