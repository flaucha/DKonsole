package pod

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

type mockLogRepo struct {
	lastNamespace string
	lastPod       string
	lastOpts      *corev1.PodLogOptions
	stream        io.ReadCloser
	err           error
}

func (m *mockLogRepo) GetLogStream(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error) {
	m.lastNamespace = namespace
	m.lastPod = podName
	m.lastOpts = opts
	return m.stream, m.err
}

type mockExecService struct {
	executor remotecommand.Executor
	err      error
	received ExecRequest
}

func (m *mockExecService) CreateExecutor(client kubernetes.Interface, config *rest.Config, req ExecRequest) (remotecommand.Executor, string, error) {
	m.received = req
	return m.executor, "ws://mock", m.err
}

type mockExecutor struct {
	called bool
	opts   remotecommand.StreamOptions
	err    error
	done   chan struct{}
}

func (m *mockExecutor) Stream(options remotecommand.StreamOptions) error {
	return m.StreamWithContext(context.Background(), options)
}

func (m *mockExecutor) StreamWithContext(ctx context.Context, options remotecommand.StreamOptions) error {
	m.called = true
	m.opts = options
	if m.done != nil {
		select {
		case <-m.done:
		default:
			close(m.done)
		}
	}
	return m.err
}

func newPodServiceForHandlers() *Service {
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{},
	}
	return NewService(handlers, cluster.NewService(handlers))
}

func TestStreamPodLogs_InvalidParams(t *testing.T) {
	service := newPodServiceForHandlers()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/pods/logs", nil)

	service.StreamPodLogs(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestStreamPodLogs_Forbidden(t *testing.T) {
	service := newPodServiceForHandlers()

	req := httptest.NewRequest(http.MethodGet, "/api/pods/logs?namespace=ns1&pod=pod1", nil)
	ctx := authContext(req, models.Claims{
		Username:    "user",
		Permissions: map[string]string{"other": "view"},
	})
	rr := httptest.NewRecorder()
	service.StreamPodLogs(rr, ctx)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
}

func TestGetPodEvents_Forbidden(t *testing.T) {
	service := newPodServiceForHandlers()
	req := httptest.NewRequest(http.MethodGet, "/api/pods/events?namespace=ns1&pod=pod1", nil)
	ctx := authContext(req, models.Claims{
		Username:    "user",
		Permissions: map[string]string{"other": "view"},
	})
	rr := httptest.NewRecorder()
	service.GetPodEvents(rr, ctx)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
}

func TestExecIntoPod_PermissionDenied(t *testing.T) {
	service := newPodServiceForHandlers()
	req := httptest.NewRequest(http.MethodGet, "/api/pods/exec?namespace=ns1&pod=pod1", nil)
	ctx := authContext(req, models.Claims{
		Username:    "user",
		Permissions: map[string]string{"ns1": "view"},
	})
	rr := httptest.NewRecorder()
	service.ExecIntoPod(rr, ctx)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
}

func TestStreamPodLogs_ClientError(t *testing.T) {
	service := newPodServiceForHandlers()

	req := httptest.NewRequest(http.MethodGet, "/api/pods/logs?namespace=ns1&pod=pod1", nil)
	req = authContext(req, models.Claims{
		Username:    "user",
		Permissions: map[string]string{"ns1": "view"},
	})
	rr := httptest.NewRecorder()

	service.StreamPodLogs(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 when client missing", rr.Code)
	}
}

func TestStreamPodLogs_Success(t *testing.T) {
	logRepo := &mockLogRepo{
		stream: io.NopCloser(strings.NewReader("hello logs")),
	}
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": k8sfake.NewSimpleClientset()},
	}
	service := NewService(handlers, cluster.NewService(handlers))
	service.logRepoFactory = func(client kubernetes.Interface) LogRepository {
		return logRepo
	}

	req := httptest.NewRequest(http.MethodGet, "/api/pods/logs?namespace=ns1&pod=pod1", nil)
	req = authContext(req, models.Claims{
		Username:    "user",
		Permissions: map[string]string{"ns1": "view"},
	})
	rr := httptest.NewRecorder()

	service.StreamPodLogs(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if body := rr.Body.String(); !strings.Contains(body, "hello logs") {
		t.Fatalf("unexpected body: %q", body)
	}
	if logRepo.lastOpts == nil || !logRepo.lastOpts.Follow {
		t.Fatalf("expected Follow=true in log options")
	}
}

func TestExecIntoPod_MissingClient(t *testing.T) {
	service := newPodServiceForHandlers()
	req := httptest.NewRequest(http.MethodGet, "/api/pods/exec?namespace=ns1&pod=pod1", nil)
	req = authContext(req, models.Claims{
		Username:    "user",
		Permissions: map[string]string{"ns1": "edit"},
	})
	rr := httptest.NewRecorder()

	service.ExecIntoPod(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 when client missing", rr.Code)
	}
}

func TestExecIntoPod_MissingRESTConfig(t *testing.T) {
	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": k8sfake.NewSimpleClientset()},
	}
	service := NewService(handlers, cluster.NewService(handlers))

	req := httptest.NewRequest(http.MethodGet, "/api/pods/exec?namespace=ns1&pod=pod1", nil)
	req = authContext(req, models.Claims{
		Username:    "user",
		Permissions: map[string]string{"ns1": "edit"},
	})
	rr := httptest.NewRecorder()

	service.ExecIntoPod(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 when REST config missing", rr.Code)
	}
}

func TestExecIntoPod_SuccessWithMockExecutor(t *testing.T) {
	executor := &mockExecutor{done: make(chan struct{})}
	execSvc := &mockExecService{executor: executor}
	handlers := &models.Handlers{
		Clients:     map[string]kubernetes.Interface{"default": k8sfake.NewSimpleClientset()},
		RESTConfigs: map[string]*rest.Config{"default": {}},
	}
	service := NewService(handlers, cluster.NewService(handlers))
	service.execFactory = func() execCreator { return execSvc }

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = authContext(r, models.Claims{
			Username:    "admin",
			Permissions: map[string]string{"ns1": "edit"},
			Role:        "admin",
		})
		service.ExecIntoPod(w, r)
	}))
	defer srv.Close()

	u := strings.Replace(srv.URL, "http", "ws", 1) + "/api/pods/exec?namespace=ns1&pod=pod1"
	conn, resp, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("websocket dial failed: %v", err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("status = %d, want 101", resp.StatusCode)
	}
	select {
	case <-executor.done:
	case <-time.After(2 * time.Second):
		t.Fatalf("executor not called in time")
	}
	if execSvc.received.Namespace != "ns1" || execSvc.received.PodName != "pod1" {
		t.Fatalf("unexpected exec request: %+v", execSvc.received)
	}
}

func TestGetPodEvents_Success(t *testing.T) {
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "evt1",
			Namespace: "ns1",
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      "pod1",
			Namespace: "ns1",
		},
	}
	client := k8sfake.NewSimpleClientset(event)

	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": client},
	}
	service := NewService(handlers, cluster.NewService(handlers))

	req := httptest.NewRequest(http.MethodGet, "/api/pods/events?namespace=ns1&pod=pod1", nil)
	req = authContext(req, models.Claims{
		Username:    "user",
		Permissions: map[string]string{"ns1": "view"},
	})

	rr := httptest.NewRecorder()
	service.GetPodEvents(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var events []EventInfo
	if err := json.NewDecoder(rr.Body).Decode(&events); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestGetPodEvents_RepoError(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	client.Fake.PrependReactor("list", "events", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("boom")
	})

	handlers := &models.Handlers{
		Clients: map[string]kubernetes.Interface{"default": client},
	}
	service := NewService(handlers, cluster.NewService(handlers))

	req := httptest.NewRequest(http.MethodGet, "/api/pods/events?namespace=ns1&pod=pod1", nil)
	req = authContext(req, models.Claims{
		Username:    "user",
		Permissions: map[string]string{"ns1": "view"},
	})

	rr := httptest.NewRecorder()
	service.GetPodEvents(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}
}

func authContext(req *http.Request, claims models.Claims) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), auth.UserContextKey(), &auth.AuthClaims{
		Claims: claims,
	}))
}
