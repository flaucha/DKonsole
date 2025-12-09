package helm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/models"
	
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// MockServiceFactory implements ServiceFactoryInterface
type MockServiceFactory struct {
	HelmReleaseService HelmReleaseServiceInterface
	HelmInstallService HelmInstallServiceInterface
	HelmUpgradeService HelmUpgradeServiceInterface
}

func (m *MockServiceFactory) CreateHelmReleaseService(client kubernetes.Interface) HelmReleaseServiceInterface {
	return m.HelmReleaseService
}

func (m *MockServiceFactory) CreateHelmInstallService(client kubernetes.Interface) HelmInstallServiceInterface {
	return m.HelmInstallService
}

func (m *MockServiceFactory) CreateHelmUpgradeService(client kubernetes.Interface) HelmUpgradeServiceInterface {
	return m.HelmUpgradeService
}

// MockHelmReleaseService implements HelmReleaseServiceInterface
type MockHelmReleaseService struct {
	GetHelmReleasesFunc   func(ctx context.Context) ([]HelmRelease, error)
	DeleteHelmReleaseFunc func(ctx context.Context, req DeleteHelmReleaseRequest) (*DeleteHelmReleaseResponse, error)
	GetChartInfoFunc      func(ctx context.Context, namespace, releaseName string) (*ChartInfo, error)
}

func (m *MockHelmReleaseService) GetHelmReleases(ctx context.Context) ([]HelmRelease, error) {
	if m.GetHelmReleasesFunc != nil {
		return m.GetHelmReleasesFunc(ctx)
	}
	return nil, nil
}

func (m *MockHelmReleaseService) DeleteHelmRelease(ctx context.Context, req DeleteHelmReleaseRequest) (*DeleteHelmReleaseResponse, error) {
	if m.DeleteHelmReleaseFunc != nil {
		return m.DeleteHelmReleaseFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockHelmReleaseService) GetChartInfo(ctx context.Context, namespace, releaseName string) (*ChartInfo, error) {
	if m.GetChartInfoFunc != nil {
		return m.GetChartInfoFunc(ctx, namespace, releaseName)
	}
	return &ChartInfo{}, nil
}

// MockHelmInstallService
type MockHelmInstallService struct {
	InstallHelmReleaseFunc func(ctx context.Context, req InstallHelmReleaseRequest) (*InstallHelmReleaseResponse, error)
}

func (m *MockHelmInstallService) InstallHelmRelease(ctx context.Context, req InstallHelmReleaseRequest) (*InstallHelmReleaseResponse, error) {
	if m.InstallHelmReleaseFunc != nil {
		return m.InstallHelmReleaseFunc(ctx, req)
	}
	return nil, nil
}

// MockHelmUpgradeService
type MockHelmUpgradeService struct {
	UpgradeHelmReleaseFunc func(ctx context.Context, req UpgradeHelmReleaseRequest) (*UpgradeHelmReleaseResponse, error)
}

func (m *MockHelmUpgradeService) UpgradeHelmRelease(ctx context.Context, req UpgradeHelmReleaseRequest) (*UpgradeHelmReleaseResponse, error) {
	if m.UpgradeHelmReleaseFunc != nil {
		return m.UpgradeHelmReleaseFunc(ctx, req)
	}
	return nil, nil
}

func setupTestService() (*Service, *MockServiceFactory) {
	// Setup fake K8s client
	clientset := fake.NewSimpleClientset()
	handlers := &models.Handlers{
		Clients:     make(map[string]kubernetes.Interface),
		Dynamics:    make(map[string]dynamic.Interface),
		Metrics:     make(map[string]*metricsv.Clientset),
		RESTConfigs: make(map[string]*rest.Config),
	}
	handlers.Clients["default"] = clientset

	// Setup ClusterService
	clusterService := cluster.NewService(handlers)

	// Setup mocks
	mockReleaseService := &MockHelmReleaseService{}
	mockInstallService := &MockHelmInstallService{}
	mockUpgradeService := &MockHelmUpgradeService{}

	mockFactory := &MockServiceFactory{
		HelmReleaseService: mockReleaseService,
		HelmInstallService: mockInstallService,
		HelmUpgradeService: mockUpgradeService,
	}

	// Create service with mock factory
	service := NewServiceWithFactory(handlers, clusterService, mockFactory)

	return service, mockFactory
}

func TestGetHelmReleases(t *testing.T) {
	service, mockFactory := setupTestService()
	mockReleaseService := mockFactory.HelmReleaseService.(*MockHelmReleaseService)

	tests := []struct {
		name           string
		mockReleases   []HelmRelease
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			mockReleases: []HelmRelease{
				{Name: "release-1", Namespace: "default", Version: "1", Status: "deployed"},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"name":"release-1","namespace":"default","version":"1","status":"deployed"}]`,
		},
		{
			name:           "Service Error",
			mockReleases:   nil,
			mockError:      fmt.Errorf("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to get Helm releases: service error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReleaseService.GetHelmReleasesFunc = func(ctx context.Context) ([]HelmRelease, error) {
				return tt.mockReleases, tt.mockError
			}

			req := httptest.NewRequest("GET", "/api/helm/releases", nil)
			w := httptest.NewRecorder()

			service.GetHelmReleases(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestDeleteHelmRelease(t *testing.T) {
	service, mockFactory := setupTestService()
	mockReleaseService := mockFactory.HelmReleaseService.(*MockHelmReleaseService)

	tests := []struct {
		name           string
		method         string
		params         string
		mockResponse   *DeleteHelmReleaseResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Success",
			method:         "DELETE",
			params:         "?name=my-release&namespace=default",
			mockResponse:   &DeleteHelmReleaseResponse{SecretsDeleted: 3},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Method Not Allowed",
			method:         "GET",
			params:         "?name=my-release&namespace=default",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Missing Parameters",
			method:         "DELETE",
			params:         "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Name",
			method:         "DELETE",
			params:         "?name=Invalid_Name&namespace=default",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Service Error",
			method:         "DELETE",
			params:         "?name=my-release&namespace=default",
			mockResponse:   nil,
			mockError:      fmt.Errorf("delete error"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReleaseService.DeleteHelmReleaseFunc = func(ctx context.Context, req DeleteHelmReleaseRequest) (*DeleteHelmReleaseResponse, error) {
				return tt.mockResponse, tt.mockError
			}

			req := httptest.NewRequest(tt.method, "/api/helm/releases"+tt.params, nil)
			w := httptest.NewRecorder()

			service.DeleteHelmRelease(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	handlers := &models.Handlers{
		Clients:     make(map[string]kubernetes.Interface),
		Dynamics:    make(map[string]dynamic.Interface),
		Metrics:     make(map[string]*metricsv.Clientset),
		RESTConfigs: make(map[string]*rest.Config),
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)
	if service == nil {
		t.Error("NewService returned nil")
	}
	if service.serviceFactory == nil {
		t.Error("ServiceFactory not initialized")
	}
}

func TestInstallHelmReleaseHandler(t *testing.T) {
	service, mockFactory := setupTestService()
	mockInstallService := mockFactory.HelmInstallService.(*MockHelmInstallService)

	tests := []struct {
		name           string
		method         string
		body           map[string]interface{}
		mockResponse   *InstallHelmReleaseResponse
		mockError      error
		expectedStatus int
	}{
		// ... (body of tests)
		{
			name:   "Success",
			method: "POST",
			body: map[string]interface{}{
				"name":      "my-release",
				"namespace": "default",
				"chart":     "stable/chart",
			},
			mockResponse:   &InstallHelmReleaseResponse{Status: "success", Message: "Installed", JobName: "job-123"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Method Not Allowed",
			method:         "GET",
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "Missing Required Fields",
			method: "POST",
			body: map[string]interface{}{
				"name": "my-release",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid Name",
			method: "POST",
			body: map[string]interface{}{
				"name":      "Invalid_Name",
				"namespace": "default",
				"chart":     "stable/chart",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Service Error",
			method: "POST",
			body: map[string]interface{}{
				"name":      "my-release",
				"namespace": "default",
				"chart":     "stable/chart",
			},
			mockResponse:   nil,
			mockError:      fmt.Errorf("install error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInstallService.InstallHelmReleaseFunc = func(ctx context.Context, req InstallHelmReleaseRequest) (*InstallHelmReleaseResponse, error) {
				return tt.mockResponse, tt.mockError
			}

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/api/helm/install", bytes.NewBuffer(bodyBytes))
			w := httptest.NewRecorder()

			service.InstallHelmRelease(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestUpgradeHelmReleaseHandler(t *testing.T) {
	service, mockFactory := setupTestService()
	mockUpgradeService := mockFactory.HelmUpgradeService.(*MockHelmUpgradeService)

	tests := []struct {
		name           string
		method         string
		body           map[string]interface{}
		mockResponse   *UpgradeHelmReleaseResponse
		mockError      error
		expectedStatus int
	}{
		// ... (body of tests)
		{
			name:   "Success",
			method: "POST",
			body: map[string]interface{}{
				"name":      "my-release",
				"namespace": "default",
				"chart":     "stable/chart",
			},
			mockResponse:   &UpgradeHelmReleaseResponse{Status: "success", Message: "Upgraded", JobName: "job-123"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Method Not Allowed",
			method:         "GET",
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "Missing Required Fields",
			method: "POST",
			body: map[string]interface{}{
				"name": "my-release",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid Name",
			method: "POST",
			body: map[string]interface{}{
				"name":      "Invalid_Name",
				"namespace": "default",
				"chart":     "stable/chart",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Service Error",
			method: "POST",
			body: map[string]interface{}{
				"name":      "my-release",
				"namespace": "default",
			},
			mockResponse:   nil,
			mockError:      fmt.Errorf("upgrade error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUpgradeService.UpgradeHelmReleaseFunc = func(ctx context.Context, req UpgradeHelmReleaseRequest) (*UpgradeHelmReleaseResponse, error) {
				return tt.mockResponse, tt.mockError
			}

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/api/helm/upgrade", bytes.NewBuffer(bodyBytes))
			w := httptest.NewRecorder()

			service.UpgradeHelmRelease(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}
