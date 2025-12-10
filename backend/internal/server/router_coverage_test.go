package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func addAuthHeader(t *testing.T, req *http.Request, claims models.Claims) *http.Request {
	authClaims := &auth.AuthClaims{
		Claims:           claims,
		RegisteredClaims: jwt.RegisteredClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, authClaims)
	tokenString, err := token.SignedString([]byte("test-secret")) // Matches newTestRouter secret
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("Origin", "http://example.com") // Satisfy CSRF check
	return req
}

func TestRouter_HelmRoutes_Coverage(t *testing.T) {
	router, _ := newTestRouter(t)

	// GET /api/helm/releases
	req := httptest.NewRequest(http.MethodGet, "/api/helm/releases", nil)
	req = addAuthHeader(t, req, models.Claims{Username: "admin", Role: "admin"})
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	// Even if it returns error (due to empty context/fake client), it covers the route registration
	// We check for status code != 404
	if rr.Code == http.StatusNotFound {
		t.Errorf("GET /api/helm/releases returned 404")
	}

	// POST /api/helm/releases/install
	req = httptest.NewRequest(http.MethodPost, "/api/helm/releases/install", nil)
	req = addAuthHeader(t, req, models.Claims{Username: "admin", Role: "admin"})
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code == http.StatusNotFound {
		t.Errorf("POST /api/helm/releases/install returned 404")
	}

	// Method Not Allowed checks (indirectly covers the "else" branches in handlers)
	req = httptest.NewRequest(http.MethodPut, "/api/helm/releases/install", nil)
	req = addAuthHeader(t, req, models.Claims{Username: "admin", Role: "admin"})
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("PUT /api/helm/releases/install should be 405, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/helm/releases", nil) // Only GET, DELETE, POST allowed
	req = addAuthHeader(t, req, models.Claims{Username: "admin", Role: "admin"})
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("PUT /api/helm/releases should be 405, got %d", rr.Code)
	}
}

func TestRouter_OtherRoutes_Coverage(t *testing.T) {
	router, _ := newTestRouter(t)

	// Logo
	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/logo"},
		{http.MethodPost, "/api/logo"},   // Upload
		{http.MethodDelete, "/api/logo"}, // Delete
		{http.MethodGet, "/api/prometheus/status"},
		{http.MethodGet, "/api/prometheus/metrics"},
		{http.MethodGet, "/api/prometheus/pod-metrics"},
		{http.MethodGet, "/api/prometheus/cluster-overview"},
		{http.MethodGet, "/api/settings/prometheus/url"},
		{http.MethodPut, "/api/settings/prometheus/url"},
		{http.MethodGet, "/api/ldap/status"},
		{http.MethodGet, "/api/ldap/config"},
		{http.MethodPut, "/api/ldap/config"},
		{http.MethodGet, "/api/ldap/groups"},
		{http.MethodPut, "/api/ldap/groups"},
		{http.MethodGet, "/api/ldap/credentials"},
		{http.MethodPut, "/api/ldap/credentials"},
		{http.MethodPost, "/api/ldap/test"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, nil)
			// Mock admin permission if needed?
			// newTestRouter uses real Auth middleware with mock client.
			// Ideally we just want to hit the router function.
			// Auth middleware might return 401 but that counts as covered route.
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			if rr.Code == http.StatusNotFound {
				t.Errorf("%s %s returned 404", ep.method, ep.path)
			}
		})
	}

	// Method Not Allowed for Logo (PUT not allowed)
	req := httptest.NewRequest(http.MethodPut, "/api/logo", nil)
	req = addAuthHeader(t, req, models.Claims{Username: "admin", Role: "admin"})
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("PUT /api/logo should be 405, got %d", rr.Code)
	}
}
