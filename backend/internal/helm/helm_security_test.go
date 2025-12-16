package helm

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

// Mocks

type MockServiceFactorySecurity struct {
	mock.Mock
}

func (m *MockServiceFactorySecurity) CreateHelmReleaseService(client kubernetes.Interface) HelmReleaseServiceInterface {
	args := m.Called(client)
	return args.Get(0).(HelmReleaseServiceInterface)
}

func (m *MockServiceFactorySecurity) CreateHelmUpgradeService(client kubernetes.Interface) HelmUpgradeServiceInterface {
	args := m.Called(client)
	return args.Get(0).(HelmUpgradeServiceInterface)
}

func (m *MockServiceFactorySecurity) CreateHelmInstallService(client kubernetes.Interface) HelmInstallServiceInterface {
	args := m.Called(client)
	return args.Get(0).(HelmInstallServiceInterface)
}

type MockHelmReleaseServiceSecurity struct {
	mock.Mock
}

func (m *MockHelmReleaseServiceSecurity) GetHelmReleases(ctx context.Context) ([]HelmRelease, error) {
	args := m.Called(ctx)
	return args.Get(0).([]HelmRelease), args.Error(1)
}

func (m *MockHelmReleaseServiceSecurity) GetChartInfo(ctx context.Context, namespace, releaseName string) (*ChartInfo, error) {
	args := m.Called(ctx, namespace, releaseName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ChartInfo), args.Error(1)
}

func (m *MockHelmReleaseServiceSecurity) DeleteHelmRelease(ctx context.Context, req DeleteHelmReleaseRequest) (*DeleteHelmReleaseResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DeleteHelmReleaseResponse), args.Error(1)
}

type MockHelmUpgradeServiceSecurity struct {
	mock.Mock
}

func (m *MockHelmUpgradeServiceSecurity) UpgradeHelmRelease(ctx context.Context, req UpgradeHelmReleaseRequest) (*UpgradeHelmReleaseResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UpgradeHelmReleaseResponse), args.Error(1)
}

type MockHelmInstallServiceSecurity struct {
	mock.Mock
}

func (m *MockHelmInstallServiceSecurity) InstallHelmRelease(ctx context.Context, req InstallHelmReleaseRequest) (*InstallHelmReleaseResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*InstallHelmReleaseResponse), args.Error(1)
}

type MockClusterServiceSecurity struct {
	mock.Mock
}

func (m *MockClusterServiceSecurity) GetClient(r *http.Request) (kubernetes.Interface, error) {
	args := m.Called(r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(kubernetes.Interface), args.Error(1)
}

func TestGetHelmReleases_Security(t *testing.T) {
	tests := []struct {
		name             string
		userRole         string
		permissions      map[string]string
		foundReleases    []HelmRelease
		expectedReleases []HelmRelease
		expectedStatus   int
	}{
		{
			name:        "Unauthorized - No permissions",
			userRole:    "user",
			permissions: map[string]string{},
			foundReleases: []HelmRelease{
				{Name: "rel1", Namespace: "ns1"},
				{Name: "rel2", Namespace: "ns2"},
			},
			expectedReleases: []HelmRelease{}, // Empty
			expectedStatus:   http.StatusOK,
		},
		{
			name:        "Authorized - View ns1 only",
			userRole:    "user",
			permissions: map[string]string{"ns1": "view"},
			foundReleases: []HelmRelease{
				{Name: "rel1", Namespace: "ns1"},
				{Name: "rel2", Namespace: "ns2"},
			},
			expectedReleases: []HelmRelease{
				{Name: "rel1", Namespace: "ns1"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Authorized - Admin sees all",
			userRole:    "admin",
			permissions: map[string]string{},
			foundReleases: []HelmRelease{
				{Name: "rel1", Namespace: "ns1"},
				{Name: "rel2", Namespace: "ns2"},
			},
			expectedReleases: []HelmRelease{
				{Name: "rel1", Namespace: "ns1"},
				{Name: "rel2", Namespace: "ns2"},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClusterService := new(MockClusterServiceSecurity)
			mockFactory := new(MockServiceFactorySecurity)
			mockReleaseService := new(MockHelmReleaseServiceSecurity)

			// Mock Client
			k8sClient := k8sfake.NewSimpleClientset()
			mockClusterService.On("GetClient", mock.Anything).Return(k8sClient, nil)

			// Mock Factory
			mockFactory.On("CreateHelmReleaseService", k8sClient).Return(mockReleaseService)

			// Mock Service calls
			mockReleaseService.On("GetHelmReleases", mock.Anything).Return(tt.foundReleases, nil)

			service := NewServiceWithFactory(&models.Handlers{}, mockClusterService, mockFactory)

			req := httptest.NewRequest(http.MethodGet, "/api/helm/releases", nil)
			rr := httptest.NewRecorder()

			// Inject User Context
			claims := &models.Claims{
				Username:    "testuser",
				Role:        tt.userRole,
				Permissions: tt.permissions,
			}
			ctx := context.WithValue(req.Context(), auth.UserContextKey(), &auth.AuthClaims{Claims: *claims})
			req = req.WithContext(ctx)

			service.GetHelmReleases(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.expectedStatus)
			}

			if rr.Code == http.StatusOK {
				var gotReleases []HelmRelease
				_ = json.NewDecoder(rr.Body).Decode(&gotReleases)
				if len(gotReleases) != len(tt.expectedReleases) {
					t.Errorf("got %d releases, want %d", len(gotReleases), len(tt.expectedReleases))
				}
			}
		})
	}
}

func TestDeleteHelmRelease_Security(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		permissions    map[string]string
		reqNamespace   string
		expectedStatus int
	}{
		{
			name:           "Unauthorized - No permissions",
			userRole:       "user",
			permissions:    map[string]string{},
			reqNamespace:   "ns1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Unauthorized - View permissions",
			userRole:       "user",
			permissions:    map[string]string{"ns1": "view"},
			reqNamespace:   "ns1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Authorized - Edit permissions",
			userRole:       "user",
			permissions:    map[string]string{"ns1": "edit"},
			reqNamespace:   "ns1",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClusterService := new(MockClusterServiceSecurity)
			mockFactory := new(MockServiceFactorySecurity)
			mockReleaseService := new(MockHelmReleaseServiceSecurity)

			if tt.expectedStatus == http.StatusOK {
				k8sClient := k8sfake.NewSimpleClientset()
				mockClusterService.On("GetClient", mock.Anything).Return(k8sClient, nil)
				mockFactory.On("CreateHelmReleaseService", k8sClient).Return(mockReleaseService)
				mockReleaseService.On("DeleteHelmRelease", mock.Anything, mock.Anything).Return(&DeleteHelmReleaseResponse{}, nil)
			}
			// If forbidden, GetClient/CreateService shouldn't be reached or don't matter

			service := NewServiceWithFactory(&models.Handlers{}, mockClusterService, mockFactory)

			req := httptest.NewRequest(http.MethodDelete, "/api/helm/releases?name=rel1&namespace="+tt.reqNamespace, nil)
			rr := httptest.NewRecorder()

			claims := &models.Claims{
				Username:    "testuser",
				Role:        tt.userRole,
				Permissions: tt.permissions,
			}
			ctx := context.WithValue(req.Context(), auth.UserContextKey(), &auth.AuthClaims{Claims: *claims})
			req = req.WithContext(ctx)

			service.DeleteHelmRelease(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestUpgradeHelmRelease_Security(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		permissions    map[string]string
		reqNamespace   string
		expectedStatus int
	}{
		{
			name:           "Unauthorized - ReadOnly",
			userRole:       "user",
			permissions:    map[string]string{"ns1": "view"},
			reqNamespace:   "ns1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Authorized - Edit",
			userRole:       "user",
			permissions:    map[string]string{"ns1": "edit"},
			reqNamespace:   "ns1",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClusterService := new(MockClusterServiceSecurity)
			mockFactory := new(MockServiceFactorySecurity)
			mockUpgradeService := new(MockHelmUpgradeServiceSecurity)

			if tt.expectedStatus == http.StatusOK {
				k8sClient := k8sfake.NewSimpleClientset()
				mockClusterService.On("GetClient", mock.Anything).Return(k8sClient, nil)
				mockFactory.On("CreateHelmUpgradeService", k8sClient).Return(mockUpgradeService)
				mockUpgradeService.On("UpgradeHelmRelease", mock.Anything, mock.Anything).Return(&UpgradeHelmReleaseResponse{Status: "deployed"}, nil)
			}

			service := NewServiceWithFactory(&models.Handlers{}, mockClusterService, mockFactory)

			body := map[string]string{
				"name":      "rel1",
				"namespace": tt.reqNamespace,
			}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/api/helm/releases", bytes.NewBuffer(jsonBody))
			rr := httptest.NewRecorder()

			claims := &models.Claims{
				Username:    "testuser",
				Role:        tt.userRole,
				Permissions: tt.permissions,
			}
			ctx := context.WithValue(req.Context(), auth.UserContextKey(), &auth.AuthClaims{Claims: *claims})
			req = req.WithContext(ctx)

			service.UpgradeHelmRelease(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestInstallHelmRelease_Security(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		permissions    map[string]string
		reqNamespace   string
		expectedStatus int
	}{
		{
			name:           "Unauthorized - ReadOnly",
			userRole:       "user",
			permissions:    map[string]string{"ns1": "view"},
			reqNamespace:   "ns1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Authorized - Edit",
			userRole:       "user",
			permissions:    map[string]string{"ns1": "edit"},
			reqNamespace:   "ns1",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClusterService := new(MockClusterServiceSecurity)
			mockFactory := new(MockServiceFactorySecurity)
			mockInstallService := new(MockHelmInstallServiceSecurity)

			if tt.expectedStatus == http.StatusOK {
				k8sClient := k8sfake.NewSimpleClientset()
				mockClusterService.On("GetClient", mock.Anything).Return(k8sClient, nil)
				mockFactory.On("CreateHelmInstallService", k8sClient).Return(mockInstallService)
				mockInstallService.On("InstallHelmRelease", mock.Anything, mock.Anything).Return(&InstallHelmReleaseResponse{Status: "deployed"}, nil)
			}

			service := NewServiceWithFactory(&models.Handlers{}, mockClusterService, mockFactory)

			body := map[string]string{
				"name":      "rel1",
				"namespace": tt.reqNamespace,
				"chart":     "nginx",
			}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/api/helm/releases/install", bytes.NewBuffer(jsonBody))
			rr := httptest.NewRecorder()

			claims := &models.Claims{
				Username:    "testuser",
				Role:        tt.userRole,
				Permissions: tt.permissions,
			}
			ctx := context.WithValue(req.Context(), auth.UserContextKey(), &auth.AuthClaims{Claims: *claims})
			req = req.WithContext(ctx)

			service.InstallHelmRelease(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}
