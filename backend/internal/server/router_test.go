package server

import (
	"net/http"
	"net/http/httptest"
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

func newTestRouter(t *testing.T) *http.ServeMux {
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

	return NewRouter(Dependencies{
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
}

func TestRouter_HealthAndCORS(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "https://example.com")
	router := newTestRouter(t)

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
