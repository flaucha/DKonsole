package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	"github.com/flaucha/DKonsole/backend/internal/api"
	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/helm"
	"github.com/flaucha/DKonsole/backend/internal/k8s"
	"github.com/flaucha/DKonsole/backend/internal/ldap"
	"github.com/flaucha/DKonsole/backend/internal/logo"
	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/pod"
	"github.com/flaucha/DKonsole/backend/internal/prometheus"
	"github.com/flaucha/DKonsole/backend/internal/settings"
)

func newTestRouter(t *testing.T) (*http.ServeMux, *models.Handlers) {
	t.Helper()

	clientset := k8sfake.NewSimpleClientset(&appsv1.Deployment{})
	dyn := fake.NewSimpleDynamicClient(runtime.NewScheme())

	handlersModel := &models.Handlers{
		Clients:     map[string]kubernetes.Interface{"default": clientset},
		Dynamics:    map[string]dynamic.Interface{"default": dyn},
		RESTConfigs: map[string]*rest.Config{},
	}

	clusterService := cluster.NewService(handlersModel)
	k8sService := k8s.NewService(handlersModel, clusterService)
	apiService := api.NewService(clusterService)
	helmService := helm.NewService(handlersModel, clusterService)
	podService := pod.NewService(handlersModel, clusterService)
	promService := prometheus.NewHTTPHandler("", clusterService)
	authService, err := auth.NewService(clientset, "test-secret")
	if err != nil {
		t.Fatalf("failed to init auth service: %v", err)
	}
	ldapService := ldap.NewServiceFactory(clientset, "test-secret").NewService()
	settingsService := settings.NewServiceFactory(clientset, handlersModel, "test-secret", promService).NewService()
	logoService := logo.NewService(clientset, "default")

	tmpDir := t.TempDir()

	router := NewRouter(Dependencies{
		AuthService:       authService,
		LDAPService:       ldapService,
		ClusterService:    clusterService,
		K8sService:        k8sService,
		APIService:        apiService,
		HelmService:       helmService,
		PodService:        podService,
		PrometheusService: promService,
		SettingsService:   settingsService,
		LogoService:       logoService,
		HandlersModel:     handlersModel,
		StaticDir:         filepath.Join(tmpDir, "static"),
	})

	return router, handlersModel
}

func TestRouter_HealthAndCORS(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "https://example.com")
	router, _ := newTestRouter(t)

	// Liveness endpoint
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want 200", rr.Code)
	}
	if rr.Body.String() == "" {
		t.Fatalf("healthz body is empty")
	}

	// CORS preflight should be allowed without hitting auth middleware
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodOptions, "/api/apis", nil)
	req.Header.Set("Origin", "https://example.com")
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("OPTIONS status = %d, want 200", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("CORS header = %q, want https://example.com", got)
	}
}

func TestRouter_ReadyzVariants(t *testing.T) {
	router, handlers := newTestRouter(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("readyz status = %d, want 200", rr.Code)
	}

	handlers.Clients = map[string]kubernetes.Interface{} // simulate missing client
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz status without client = %d, want 503", rr2.Code)
	}
}

func TestContentTypeFixer_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	fixer := &contentTypeFixer{ResponseWriter: rr, path: "/assets/app.js"}
	fixer.WriteHeader(http.StatusOK)
	if ct := rr.Header().Get("Content-Type"); ct != "application/javascript; charset=utf-8" {
		t.Fatalf("unexpected content type: %s", ct)
	}

	// Test default fallthrough
	rr = httptest.NewRecorder()
	fixer = &contentTypeFixer{ResponseWriter: rr, path: "/assets/unknown.xyz"}
	fixer.WriteHeader(http.StatusOK)
	// We just ensure it doesn't panic
}

func TestRouter_StaticFiles(t *testing.T) {
	// We need to create the actual static files in the temp dir
	// The newTestRouter uses a temp dir for static files
	// However, we don't have easy access to it from here unless we modify newTestRouter return signature
	// or recreating the router with known path.
	// Let's create a specific router for this test.
	tmpDir := t.TempDir()
	staticDir := filepath.Join(tmpDir, "static")
	if err := os.MkdirAll(filepath.Join(staticDir, "assets"), 0755); err != nil {
		t.Fatalf("failed to create static dir: %v", err)
	}

	// Create a dummy index.html
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<html></html>"), 0600); err != nil {
		t.Fatalf("failed to create index.html: %v", err)
	}

	// Create a dummy asset
	if err := os.WriteFile(filepath.Join(staticDir, "assets", "test.js"), []byte("console.log('hi')"), 0600); err != nil {
		t.Fatalf("failed to create test.js: %v", err)
	}

	// Re-init router with this static dir
	clientset := k8sfake.NewSimpleClientset()
	handlersModel := &models.Handlers{Clients: map[string]kubernetes.Interface{"default": clientset}}
	nullDep := Dependencies{
		StaticDir:     staticDir,
		HandlersModel: handlersModel,
		AuthService:   &auth.Service{}, // Minimal needed
	}
	// We need a clearer way to init null deps, but let's try to reuse newTestRouter logic manually or modify it.
	// Actually, let's just use NewRouter with minimal deps
	r := NewRouter(nullDep)

	t.Run("serve index.html", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("serve asset", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/test.js", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "application/javascript; charset=utf-8" {
			t.Errorf("expected js content type, got %s", ct)
		}
	})

	t.Run("404 for unknown", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/unknown", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		// It should serve index.html for SPA routing (if configured) or 404
		// Looking at code:
		// if strings.HasPrefix(r.URL.Path, "/api") -> 404
		// else serve index.html
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 (index.html), got %d", rr.Code)
		}
	})

	t.Run("404 for api", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/unknown", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rr.Code)
		}
	})
}

func TestEnableCors(t *testing.T) {
	// Setup a handler wrapped with enableCors
	handler := enableCors(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("no origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		handler(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("same origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", "http://example.com")
		req.Host = "example.com"
		rr := httptest.NewRecorder()
		handler(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
		if rr.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
			t.Errorf("expected ACAO header")
		}
	})

	t.Run("different origin allowed via env", func(t *testing.T) {
		t.Setenv("ALLOWED_ORIGINS", "http://allowed.com")
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", "http://allowed.com")
		rr := httptest.NewRecorder()
		handler(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("different origin denied", func(t *testing.T) {
		t.Setenv("ALLOWED_ORIGINS", "http://allowed.com")
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", "http://denied.com")
		rr := httptest.NewRecorder()
		handler(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rr.Code)
		}
	})
}
