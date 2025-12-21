package pod

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

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

func (m *MockClusterService) GetRESTConfig(r *http.Request) (*rest.Config, error) {
	args := m.Called(r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rest.Config), args.Error(1)
}

// MockEventRepository
type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) GetEvents(ctx context.Context, namespace, podName string) ([]EventInfo, error) {
	args := m.Called(ctx, namespace, podName)
	return args.Get(0).([]EventInfo), args.Error(1)
}

// MockLogRepository
type MockLogRepository struct {
	mock.Mock
}

func (m *MockLogRepository) NewK8sLogRepository(client kubernetes.Interface) LogRepository {
	return m
}

func (m *MockLogRepository) GetLogStream(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, namespace, podName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

// MockExecService
type MockExecService struct {
	mock.Mock
}

func (m *MockExecService) CreateExecutor(client kubernetes.Interface, config *rest.Config, req ExecRequest) (remotecommand.Executor, string, error) {
	args := m.Called(client, config, req)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(remotecommand.Executor), args.String(1), args.Error(2)
}

// MockExecutor implements remotecommand.Executor
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Stream(options remotecommand.StreamOptions) error {
	return m.StreamWithContext(context.Background(), options)
}

func (m *MockExecutor) StreamWithContext(ctx context.Context, options remotecommand.StreamOptions) error {
	args := m.Called(ctx, options)
	// Simulate stream interaction if needed
	if options.Stdout != nil {
		options.Stdout.Write([]byte("mock output"))
	}
	// Sleep to allow WebSocket flush
	time.Sleep(100 * time.Millisecond)
	return args.Error(0)
}

func TestService_ExecIntoPod_Mocked(t *testing.T) {
	setup := func() (*MockClusterService, *MockExecService, *httptest.Server, string) {
		mockClusterService := new(MockClusterService)
		// mockEventRepo := new(MockEventRepository) // Unused
		mockExecService := new(MockExecService)

		execFactory := func() execCreator {
			return mockExecService
		}

		service := &Service{
			clusterService: mockClusterService,
			// eventRepo:      mockEventRepo, // Unused
			execFactory: execFactory,
			handlers:    &models.Handlers{},
		}

		// Inject Auth Context to simulate Admin user
		// We need to wrap the service handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), auth.UserContextKey(), &auth.AuthClaims{
				Claims: models.Claims{
					Role: "admin",
				},
			})
			service.ExecIntoPod(w, r.WithContext(ctx))
		})

		server := httptest.NewServer(handler)
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		return mockClusterService, mockExecService, server, wsURL
	}

	t.Run("ExecIntoPod Success", func(t *testing.T) {
		mockClusterService, mockExecService, s, u := setup()
		defer s.Close()

		mockClusterService.On("GetClient", mock.Anything).Return(nil, nil)
		mockClusterService.On("GetRESTConfig", mock.Anything).Return(&rest.Config{}, nil)

		mockExecutor := new(MockExecutor)
		mockExecutor.On("StreamWithContext", mock.Anything, mock.Anything).Return(nil)

		mockExecService.On("CreateExecutor", mock.Anything, mock.Anything, mock.Anything).Return(mockExecutor, "ws://mock", nil)

		wsURL := u + "?namespace=default&pod=pod-1&container=c1"
		ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to dial: %v", err)
		}
		defer ws.Close()

		_, msg, err := ws.ReadMessage()
		assert.NoError(t, err)
		assert.Contains(t, string(msg), "mock output")
	})

	t.Run("ExecIntoPod Missing Params", func(t *testing.T) {
		_, _, s, u := setup()
		defer s.Close()

		wsURL := u
		_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.Error(t, err)
	})

	t.Run("ExecIntoPod Cluster Client Error", func(t *testing.T) {
		mockClusterService, _, s, u := setup()
		defer s.Close()

		mockClusterService.On("GetClient", mock.Anything).Return(nil, errors.New("client error"))

		wsURL := u + "?namespace=default&pod=pod-1&container=c2"
		_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.Error(t, err)
	})
	t.Run("ExecIntoPod Forbidden", func(t *testing.T) {
		_, _, s, _ := setup()
		defer s.Close()

		handler := func(w http.ResponseWriter, r *http.Request) {
			claims := &models.Claims{
				Username:    "guest",
				Role:        "guest",
				Permissions: map[string]string{}, // No permissions
			}
			ctx := context.WithValue(r.Context(), auth.UserContextKey(), &auth.AuthClaims{Claims: *claims}) // Wrap nicely
			service := &Service{
				handlers: &models.Handlers{},
			}
			service.ExecIntoPod(w, r.WithContext(ctx))
		}

		sForbidden := httptest.NewServer(http.HandlerFunc(handler))
		defer sForbidden.Close()

		uForbidden := "ws" + strings.TrimPrefix(sForbidden.URL, "http") + "?namespace=default&pod=pod-1&container=c1"

		_, _, err := websocket.DefaultDialer.Dial(uForbidden, nil)
		assert.Error(t, err) // Should fail with 403 Forbidden
	})

	t.Run("ExecIntoPod Message Filtering", func(t *testing.T) {
		mockClusterService, mockExecService, s, u := setup()
		defer s.Close()

		mockClusterService.On("GetClient", mock.Anything).Return(nil, nil)
		mockClusterService.On("GetRESTConfig", mock.Anything).Return(&rest.Config{}, nil)

		mockExecutor := new(MockExecutor)
		mockExecutor.On("StreamWithContext", mock.Anything, mock.Anything).Return(nil)

		mockExecService.On("CreateExecutor", mock.Anything, mock.Anything, mock.Anything).Return(mockExecutor, "ws://mock", nil)

		wsURL := u + "?namespace=default&pod=pod-1&container=c1"
		ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to dial: %v", err)
		}
		defer ws.Close()

		// Send empty message (should be filtered)
		err = ws.WriteMessage(websocket.BinaryMessage, []byte{})
		assert.NoError(t, err)

		// Send null char (should be filtered)
		err = ws.WriteMessage(websocket.BinaryMessage, []byte{0})
		assert.NoError(t, err)

		// Wait a bit to ensure filtering logic runs
		time.Sleep(50 * time.Millisecond)
	})
	// Test Allowed Origins via Env Var
	t.Run("ExecIntoPod Allowed Origins", func(t *testing.T) {
		t.Setenv("ALLOWED_ORIGINS", "https://trusted.com,https://another.com")

		mockClusterService, mockExecService, s, u := setup()
		defer s.Close()

		mockClusterService.On("GetClient", mock.Anything).Return(nil, nil)
		mockClusterService.On("GetRESTConfig", mock.Anything).Return(&rest.Config{}, nil)

		mockExecutor := new(MockExecutor)
		mockExecutor.On("StreamWithContext", mock.Anything, mock.Anything).Return(nil)

		// For valid origin, creation succeeds
		mockExecService.On("CreateExecutor", mock.Anything, mock.Anything, mock.Anything).Return(mockExecutor, "ws://mock", nil)

		// Valid Origin
		uValid := u + "?namespace=default&pod=pod-1&container=c1"
		header := http.Header{}
		header.Set("Origin", "https://trusted.com")
		ws, _, err := websocket.DefaultDialer.Dial(uValid, header)
		if err != nil {
			t.Fatalf("Failed to dial valid origin: %v", err)
		}
		ws.Close()

		// Invalid Origin
		headerInvalid := http.Header{}
		headerInvalid.Set("Origin", "https://malicious.com")
		_, _, err = websocket.DefaultDialer.Dial(uValid, headerInvalid)
		assert.Error(t, err) // Should fail handshake
	})
}

func TestSetWebSocketCORSHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	setWebSocketCORSHeaders(w, req)

	assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestSetWebSocketCORSHeaders_NoOrigin(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	setWebSocketCORSHeaders(w, req)

	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}
