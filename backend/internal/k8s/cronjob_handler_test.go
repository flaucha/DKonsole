package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/mock"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

// MockServiceFactory is a mock implementation of Factory
type MockServiceFactory struct {
	mock.Mock
}

func (m *MockServiceFactory) CreateResourceService(dynamicClient dynamic.Interface, client kubernetes.Interface) *ResourceService {
	args := m.Called(dynamicClient, client)
	return args.Get(0).(*ResourceService)
}

func (m *MockServiceFactory) CreateImportService(dynamicClient dynamic.Interface, client kubernetes.Interface) *ImportService {
	args := m.Called(dynamicClient, client)
	return args.Get(0).(*ImportService)
}

func (m *MockServiceFactory) CreateNamespaceService(client kubernetes.Interface) *NamespaceService {
	args := m.Called(client)
	return args.Get(0).(*NamespaceService)
}

func (m *MockServiceFactory) CreateDeploymentService(client kubernetes.Interface) *DeploymentService {
	args := m.Called(client)
	return args.Get(0).(*DeploymentService)
}

func (m *MockServiceFactory) CreateCronJobService(client kubernetes.Interface) *CronJobService {
	args := m.Called(client)
	return args.Get(0).(*CronJobService)
}

func (m *MockServiceFactory) CreateClusterStatsService(client kubernetes.Interface) *ClusterStatsService {
	args := m.Called(client)
	return args.Get(0).(*ClusterStatsService)
}

func (m *MockServiceFactory) CreateWatchService() *WatchService {
	args := m.Called()
	return args.Get(0).(*WatchService)
}

// MockCronJobRepository for testing handler integration
type MockRefCronJobRepository struct {
	mock.Mock
}

func (m *MockRefCronJobRepository) GetCronJob(ctx context.Context, namespace, name string) (*batchv1.CronJob, error) {
	args := m.Called(ctx, namespace, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*batchv1.CronJob), args.Error(1)
}

func (m *MockRefCronJobRepository) CreateJob(ctx context.Context, namespace string, job *batchv1.Job) (*batchv1.Job, error) {
	args := m.Called(ctx, namespace, job)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*batchv1.Job), args.Error(1)
}

func TestTriggerCronJob_Handler(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		permissions    map[string]string
		reqNamespace   string
		reqName        string
		expectedStatus int
		mockSetup      func(*MockClusterService, *MockServiceFactory)
	}{
		{
			name:           "Unauthorized - No permissions",
			userRole:       "user",
			permissions:    map[string]string{}, // No access
			reqNamespace:   "default",
			reqName:        "test-cron",
			expectedStatus: http.StatusForbidden,
			mockSetup: func(mcs *MockClusterService, msf *MockServiceFactory) {
				mcs.On("GetClient", mock.Anything).Return(k8sfake.NewSimpleClientset(), nil)
			},
		},
		{
			name:           "Unauthorized - Read only permissions",
			userRole:       "user",
			permissions:    map[string]string{"default": "view"},
			reqNamespace:   "default",
			reqName:        "test-cron",
			expectedStatus: http.StatusForbidden,
			mockSetup: func(mcs *MockClusterService, msf *MockServiceFactory) {
				mcs.On("GetClient", mock.Anything).Return(k8sfake.NewSimpleClientset(), nil)
			},
		},
		{
			name:           "Authorized - Edit permissions",
			userRole:       "user",
			permissions:    map[string]string{"default": "edit"},
			reqNamespace:   "default",
			reqName:        "test-cron",
			expectedStatus: http.StatusCreated,
			mockSetup: func(mcs *MockClusterService, msf *MockServiceFactory) {
				fakeClient := k8sfake.NewSimpleClientset()
				mcs.On("GetClient", mock.Anything).Return(fakeClient, nil)

				// Create a real service with a mock repository to control the behavior
				mockRepo := new(MockRefCronJobRepository)
				// Setup mock repo expectations
				cronJob := &batchv1.CronJob{
					ObjectMeta: metav1.ObjectMeta{Name: "test-cron", Namespace: "default"},
					Spec: batchv1.CronJobSpec{
						JobTemplate: batchv1.JobTemplateSpec{
							Spec: batchv1.JobSpec{},
						},
					},
				}
				mockRepo.On("GetCronJob", mock.Anything, "default", "test-cron").Return(cronJob, nil)
				mockRepo.On("CreateJob", mock.Anything, "default", mock.AnythingOfType("*v1.Job")).Return(&batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{Name: "test-cron-manual-123", Namespace: "default"},
				}, nil)

				cronJobService := NewCronJobService(mockRepo)
				msf.On("CreateCronJobService", fakeClient).Return(cronJobService)
			},
		},
		{
			name:           "Authorized - Admin",
			userRole:       "admin",
			permissions:    map[string]string{}, // Admin implies all
			reqNamespace:   "kube-system",
			reqName:        "system-cron",
			expectedStatus: http.StatusCreated,
			mockSetup: func(mcs *MockClusterService, msf *MockServiceFactory) {
				fakeClient := k8sfake.NewSimpleClientset()
				mcs.On("GetClient", mock.Anything).Return(fakeClient, nil)

				mockRepo := new(MockRefCronJobRepository)
				cronJob := &batchv1.CronJob{
					ObjectMeta: metav1.ObjectMeta{Name: "system-cron", Namespace: "kube-system"},
					Spec: batchv1.CronJobSpec{
						JobTemplate: batchv1.JobTemplateSpec{
							Spec: batchv1.JobSpec{},
						},
					},
				}
				mockRepo.On("GetCronJob", mock.Anything, "kube-system", "system-cron").Return(cronJob, nil)
				mockRepo.On("CreateJob", mock.Anything, "kube-system", mock.AnythingOfType("*v1.Job")).Return(&batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{Name: "system-cron-manual-123", Namespace: "kube-system"},
				}, nil)

				cronJobService := NewCronJobService(mockRepo)
				msf.On("CreateCronJobService", fakeClient).Return(cronJobService)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Mocks
			mockClusterService := new(MockClusterService)
			mockServiceFactory := new(MockServiceFactory)
			handlers := &models.Handlers{}

			if tt.mockSetup != nil {
				tt.mockSetup(mockClusterService, mockServiceFactory)
			}

			// Create Service
			service := NewService(handlers, mockClusterService)
			// Swizzle factory
			service.serviceFactory = mockServiceFactory

			// Create request body
			reqBody, _ := json.Marshal(map[string]string{
				"namespace": tt.reqNamespace,
				"name":      tt.reqName,
			})
			req := httptest.NewRequest(http.MethodPost, "/api/cronjobs/trigger", bytes.NewBuffer(reqBody))
			rr := httptest.NewRecorder()

			// Inject User Context with permissions
			claims := &models.Claims{
				Username:    "testuser",
				Role:        tt.userRole,
				Permissions: tt.permissions,
			}
			ctx := context.WithValue(req.Context(), auth.UserContextKey(), &auth.AuthClaims{Claims: *claims})
			req = req.WithContext(ctx)

			// Execute
			service.TriggerCronJob(rr, req)

			// Assert
			if rr.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d. Body: %s", rr.Code, tt.expectedStatus, rr.Body.String())
			}
		})
	}
}
