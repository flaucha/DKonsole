package k8s

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func newHandlersWithDynamic() *models.Handlers {
	return &models.Handlers{
		Clients:  map[string]kubernetes.Interface{"default": k8sfake.NewSimpleClientset()},
		Dynamics: map[string]dynamic.Interface{"default": dynamicfake.NewSimpleDynamicClient(runtimeScheme())},
	}
}

func runtimeScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	// Register unstructured for dynamic client
	s.AddKnownTypeWithName(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}, &unstructured.Unstructured{})
	s.AddKnownTypeWithName(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}, &unstructured.Unstructured{})
	return s
}

func adminCtx(req *http.Request) *http.Request {
	claims := &auth.AuthClaims{Claims: models.Claims{Role: "admin"}}
	return req.WithContext(context.WithValue(req.Context(), auth.UserContextKey(), claims))
}

func editCtx(req *http.Request, ns string) *http.Request {
	claims := &auth.AuthClaims{Claims: models.Claims{Role: "", Permissions: map[string]string{ns: "edit"}}}
	return req.WithContext(context.WithValue(req.Context(), auth.UserContextKey(), claims))
}

func viewCtx(req *http.Request, ns string) *http.Request {
	claims := &auth.AuthClaims{Claims: models.Claims{Role: "", Permissions: map[string]string{ns: "view"}}}
	return req.WithContext(context.WithValue(req.Context(), auth.UserContextKey(), claims))
}

func TestUpdateResourceYAMLHandler_Success(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	resolver := &fakeGVRResolver{gvr: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, meta: models.ResourceMeta{Namespaced: true}}
	repo := &fakeResourceRepo{}
	mockFactory := newMockServiceFactory()
	mockFactory.resourceSvc = NewResourceService(repo, resolver)
	service.serviceFactory = mockFactory

	body := bytes.NewBufferString("apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: demo\n  namespace: default\nspec: {}")
	req := editCtx(httptest.NewRequest(http.MethodPut, "/api/resource?kind=Deployment&name=demo&namespace=default&namespaced=true", body), "default")
	rr := httptest.NewRecorder()

	service.UpdateResourceYAML(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if repo.patchName != "demo" || repo.patchNamespace != "default" {
		t.Fatalf("patch not called with expected args: name=%s ns=%s", repo.patchName, repo.patchNamespace)
	}
}

func TestUpdateResourceYAMLHandler_MissingParams(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := httptest.NewRequest(http.MethodPut, "/api/resource?kind=&name=", nil)
	rr := httptest.NewRecorder()

	service.UpdateResourceYAML(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestUpdateResourceYAMLHandler_Forbidden(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := viewCtx(httptest.NewRequest(http.MethodPut, "/api/resource?kind=Deployment&name=demo&namespace=default&namespaced=true", bytes.NewBufferString("kind: Deployment")), "default")
	rr := httptest.NewRecorder()

	service.UpdateResourceYAML(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rr.Code)
	}
}

func TestCreateResourceYAMLHandler_Success(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	resolver := &fakeGVRResolver{gvr: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}, meta: models.ResourceMeta{Namespaced: true}}
	repo := &fakeResourceRepo{}
	mockFactory := newMockServiceFactory()
	mockFactory.resourceSvc = NewResourceService(repo, resolver)
	service.serviceFactory = mockFactory

	body := bytes.NewBufferString("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm1\n  namespace: default\ndata:\n  k: v")
	req := adminCtx(httptest.NewRequest(http.MethodPost, "/api/resource", body))
	rr := httptest.NewRecorder()

	service.CreateResourceYAML(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	if repo.createObj == nil || repo.createObj.GetName() != "cm1" {
		t.Fatalf("create not invoked as expected")
	}
}

func TestCreateResourceYAMLHandler_ClusterNotFound(t *testing.T) {
	handlers := &models.Handlers{
		Clients:  map[string]kubernetes.Interface{},
		Dynamics: map[string]dynamic.Interface{},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := adminCtx(httptest.NewRequest(http.MethodPost, "/api/resource?cluster=ghost", bytes.NewBufferString("kind: ConfigMap")))
	rr := httptest.NewRecorder()

	service.CreateResourceYAML(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestDryRunResourceYAML_InvalidYAML(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := adminCtx(httptest.NewRequest(http.MethodPost, "/api/resource/dry-run", bytes.NewBufferString("kind:")))
	rr := httptest.NewRecorder()

	service.DryRunResourceYAML(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if rr.Body.Len() == 0 {
		t.Fatalf("expected error body, got empty")
	}
}

func TestDryRunResourceYAML_ClusterNotFound(t *testing.T) {
	handlers := &models.Handlers{
		Dynamics: map[string]dynamic.Interface{},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := adminCtx(httptest.NewRequest(http.MethodPost, "/api/resource/dry-run?cluster=ghost", bytes.NewBufferString("kind: Pod")))
	rr := httptest.NewRecorder()

	service.DryRunResourceYAML(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestValidateResourceYAML_MissingKind(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := adminCtx(httptest.NewRequest(http.MethodPost, "/api/resource/validate", bytes.NewBufferString("apiVersion: v1\nmetadata:\n  name: demo")))
	rr := httptest.NewRecorder()

	service.ValidateResourceYAML(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestServerSideApply_ClusterError(t *testing.T) {
	handlers := &models.Handlers{
		Dynamics: map[string]dynamic.Interface{},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := adminCtx(httptest.NewRequest(http.MethodPost, "/api/resource/apply?cluster=ghost", bytes.NewBufferString("kind: ConfigMap")))
	rr := httptest.NewRecorder()

	service.ServerSideApply(rr, req)

	if rr.Code != http.StatusOK && rr.Code != 0 {
		t.Fatalf("expected 200-style response, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Cluster error") {
		t.Fatalf("expected cluster error event, got %s", rr.Body.String())
	}
}

func TestImportResourceYAMLHandler_Success(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	resolver := &fakeGVRResolver{gvr: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}, meta: models.ResourceMeta{Namespaced: true}}
	repo := &fakeResourceRepo{}
	importSvc := NewImportService(repo, resolver, handlers.Clients["default"])

	mockFactory := newMockServiceFactory()
	mockFactory.importSvc = importSvc
	service.serviceFactory = mockFactory

	body := bytes.NewBufferString("---\napiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm-imp\n  namespace: default\ndata:\n  k: v\n")
	req := editCtx(httptest.NewRequest(http.MethodPost, "/api/resource/import", body), "default")
	rr := httptest.NewRecorder()

	service.ImportResourceYAML(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if repo.patchCalls == 0 {
		t.Fatalf("expected patch to be called during import")
	}
}

func TestStreamResourceCreation_Success(t *testing.T) {
	scheme := runtimeScheme()
	dyn := dynamicfake.NewSimpleDynamicClient(scheme)
	handlers := &models.Handlers{
		Clients:  map[string]kubernetes.Interface{"default": k8sfake.NewSimpleClientset()},
		Dynamics: map[string]dynamic.Interface{"default": dyn},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	// Replace factory to avoid real GVR resolver; use fake resolver and repo
	resolver := &fakeGVRResolver{gvr: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}, meta: models.ResourceMeta{Namespaced: true}}
	repo := &fakeResourceRepo{createResult: &unstructured.Unstructured{}}
	mockFactory := newMockServiceFactory()
	mockFactory.resourceSvc = NewResourceService(repo, resolver)
	service.serviceFactory = mockFactory

	body := bytes.NewBufferString("apiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\n  namespace: default\nspec:\n  containers:\n  - name: c\n    image: busybox\n")
	req := adminCtx(httptest.NewRequest(http.MethodPost, "/api/resource/stream", body))
	rr := httptest.NewRecorder()

	service.StreamResourceCreation(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	respBody := rr.Body.String()
	if !strings.Contains(respBody, "Starting creation") {
		t.Fatalf("expected start event, got %s", respBody)
	}
	if !strings.Contains(respBody, "success") {
		t.Fatalf("expected success event, got %s", respBody)
	}
}

func TestStreamResourceCreation_Error(t *testing.T) {
	handlers := &models.Handlers{
		Dynamics: map[string]dynamic.Interface{},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := adminCtx(httptest.NewRequest(http.MethodPost, "/api/resource/stream?cluster=ghost", bytes.NewBufferString("kind: Pod")))
	rr := httptest.NewRecorder()

	service.StreamResourceCreation(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "error") {
		t.Fatalf("expected error event, got %s", rr.Body.String())
	}
}

func TestDeleteResourceHandler_Success(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	resolver := &fakeGVRResolver{gvr: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}, meta: models.ResourceMeta{Namespaced: true}}
	repo := &fakeResourceRepo{}
	mockFactory := newMockServiceFactory()
	mockFactory.resourceSvc = NewResourceService(repo, resolver)
	service.serviceFactory = mockFactory

	req := editCtx(httptest.NewRequest(http.MethodDelete, "/api/resource?kind=ConfigMap&name=cm1&namespace=default", nil), "default")
	rr := httptest.NewRecorder()

	service.DeleteResource(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !repo.deleteCalled || repo.deleteName != "cm1" || repo.deleteNamespace != "default" {
		t.Fatalf("delete not invoked correctly")
	}
}

func TestDeleteResourceHandler_Forbidden(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := viewCtx(httptest.NewRequest(http.MethodDelete, "/api/resource?kind=ConfigMap&name=cm1&namespace=default", nil), "default")
	rr := httptest.NewRecorder()

	service.DeleteResource(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestDeleteResourceHandler_MissingParams(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := adminCtx(httptest.NewRequest(http.MethodDelete, "/api/resource?kind=&name=", nil))
	rr := httptest.NewRecorder()

	service.DeleteResource(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestWatchResources_Success(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	fakeWatcher := watch.NewFake()
	watchSvc := &WatchService{
		watchFunc: func(ctx context.Context, client dynamic.Interface, req WatchRequest) (watch.Interface, error) {
			return fakeWatcher, nil
		},
		transformFunc: func(event watch.Event) (*WatchResult, error) {
			obj := event.Object.(*unstructured.Unstructured)
			return &WatchResult{
				Type:      string(event.Type),
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			}, nil
		},
	}

	mockFactory := newMockServiceFactory()
	mockFactory.watchSvc = watchSvc
	service.serviceFactory = mockFactory

	server := httptest.NewServer(http.HandlerFunc(service.WatchResources))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/?kind=Pod&namespace=default"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	fakeWatcher.Add(&unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "pod-1",
			"namespace": "default",
		},
	}})

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message: %v", err)
	}
	if !strings.Contains(string(msg), "\"pod-1\"") {
		t.Fatalf("expected pod name in message, got %s", string(msg))
	}
}

func TestWatchResources_StartWatchError(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	watchSvc := &WatchService{
		watchFunc: func(ctx context.Context, client dynamic.Interface, req WatchRequest) (watch.Interface, error) {
			return nil, errors.New("watch failed")
		},
	}

	mockFactory := newMockServiceFactory()
	mockFactory.watchSvc = watchSvc
	service.serviceFactory = mockFactory

	server := httptest.NewServer(http.HandlerFunc(service.WatchResources))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/?kind=Pod&namespace=default"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message: %v", err)
	}
	if !strings.Contains(string(msg), "watch failed") {
		t.Fatalf("expected error message, got %s", string(msg))
	}
}

func TestDryRunResourceYAML_Success(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	body := bytes.NewBufferString("apiVersion: v1\nkind: Pod\nmetadata:\n  name: dry-pod\n  namespace: default\n")
	req := adminCtx(httptest.NewRequest(http.MethodPost, "/api/resource/dry-run", body))
	rr := httptest.NewRecorder()

	// Mock dynamic client behaviors (create returns object)
	// We need to inject reactor for Create to return the object
	dyn := handlers.Dynamics["default"].(*dynamicfake.FakeDynamicClient)
	dyn.PrependReactor("create", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		createAction := action.(k8stesting.CreateAction)
		return true, createAction.GetObject(), nil
	})

	service.DryRunResourceYAML(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Dry run successful") {
		t.Fatalf("expected success message, got %s", rr.Body.String())
	}
}

func TestValidateResourceYAML_Success(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	body := bytes.NewBufferString("apiVersion: v1\nkind: Pod\nmetadata:\n  name: val-pod\n")
	req := adminCtx(httptest.NewRequest(http.MethodPost, "/api/resource/validate", body))
	rr := httptest.NewRecorder()

	service.ValidateResourceYAML(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "YAML is valid") {
		t.Fatalf("expected valid message, got %s", rr.Body.String())
	}
}

func TestServerSideApply_Success(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	body := bytes.NewBufferString("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: apply-cm\n  namespace: default\n")
	req := adminCtx(httptest.NewRequest(http.MethodPost, "/api/resource/apply", body))
	rr := httptest.NewRecorder()

	// Mock dynamic client behaviors (patch returns object)
	dyn := handlers.Dynamics["default"].(*dynamicfake.FakeDynamicClient)
	dyn.PrependReactor("patch", "configmaps", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		// Return some object
		return true, &unstructured.Unstructured{Object: map[string]interface{}{"metadata": map[string]interface{}{"name": "apply-cm"}}}, nil
	})

	service.ServerSideApply(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	// Check streaming response
	resp := rr.Body.String()
	if !strings.Contains(resp, "Starting Server-Side Apply") {
		t.Fatalf("expected starting message")
	}
	if !strings.Contains(resp, "event: success") {
		t.Fatalf("expected success event")
	}
}

func TestImportResourceYAMLHandler_ClusterError(t *testing.T) {
	handlers := &models.Handlers{
		Clients:  map[string]kubernetes.Interface{},
		Dynamics: map[string]dynamic.Interface{},
	}
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	req := adminCtx(httptest.NewRequest(http.MethodPost, "/api/resource/import?cluster=ghost", bytes.NewBufferString("kind: ConfigMap")))
	rr := httptest.NewRecorder()

	service.ImportResourceYAML(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestImportResourceYAMLHandler_ImportError(t *testing.T) {
	handlers := newHandlersWithDynamic()
	clusterService := cluster.NewService(handlers)
	service := NewService(handlers, clusterService)

	// Mock import failing by ensuring resolve fails or similar?
	// ImportService calls ResolveGVR then Patch/Create.
	// If ResolveGVR fails, ImportResources fails.
	// We can't inject mock ImportService easily into Service unless we replace factory.

	// Better: Use mock factory and return error from ImportSvc.
	// We need to implement a MockImportService or fake it.
	// ImportService is a struct, not interface in factory.
	// But factory returns *ImportService.
	// So we need to control ImportService dependency.

	// ImportService uses ResourceRepository and GVRResolver.
	// Make resolver fail.

	failingResolver := &fakeGVRResolver{err: errors.New("resolve failure")}
	repo := &fakeResourceRepo{}
	importSvc := NewImportService(repo, failingResolver, handlers.Clients["default"])

	mockFactory := newMockServiceFactory()
	mockFactory.importSvc = importSvc
	service.serviceFactory = mockFactory

	body := bytes.NewBufferString("---\napiVersion: v1\nkind: Pod\nmetadata:\n  name: p\n")
	req := adminCtx(httptest.NewRequest(http.MethodPost, "/api/resource/import", body))
	rr := httptest.NewRecorder()

	service.ImportResourceYAML(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Failed to import resources") {
		t.Fatalf("expected sanitized error message, got %s", rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "resolve failure") {
		t.Fatalf("expected internal error to be hidden, got %s", rr.Body.String())
	}
}
