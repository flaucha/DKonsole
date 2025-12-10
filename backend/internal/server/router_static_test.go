package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRegisterStaticRoutes(t *testing.T) {
	// Create temporary static directory structure
	tmpDir, err := os.MkdirTemp("", "static-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.Mkdir(assetsDir, 0755); err != nil {
		t.Fatalf("Failed to create assets dir: %v", err)
	}

	// Create dummy files
	createFile(t, filepath.Join(tmpDir, "index.html"), "<html>index</html>")
	createFile(t, filepath.Join(tmpDir, "favicon.ico"), "icon data")
	createFile(t, filepath.Join(tmpDir, "robots.txt"), "User-agent: *")
	createFile(t, filepath.Join(assetsDir, "script.js"), "console.log('test')")
	createFile(t, filepath.Join(assetsDir, "style.css"), "body { color: red; }")

	// Create RouterConfig with temp dir
	mux := http.NewServeMux()
	config := RouterConfig{
		Mux: mux,
		Deps: Dependencies{
			StaticDir: tmpDir,
		},
	}

	// Register routes
	registerStaticRoutes(config)

	// Tests
	tests := []struct {
		name            string
		path            string
		wantStatus      int
		wantBody        string
		wantContentType string
	}{
		{
			name:       "Index page",
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody:   "<html>index</html>",
		},
		{
			name:       "Robots.txt",
			path:       "/robots.txt",
			wantStatus: http.StatusOK,
			wantBody:   "User-agent: *",
		},
		{
			name:            "Asset JS",
			path:            "/assets/script.js",
			wantStatus:      http.StatusOK,
			wantBody:        "console.log('test')",
			wantContentType: "application/javascript; charset=utf-8",
		},
		{
			name:            "Asset CSS",
			path:            "/assets/style.css",
			wantStatus:      http.StatusOK,
			wantBody:        "body { color: red; }",
			wantContentType: "text/css; charset=utf-8",
		},
		{
			name:       "API path should verify not found (handled by API router usually)",
			path:       "/api/something",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Missing Asset",
			path:       "/assets/missing.js",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && rr.Body.String() != tt.wantBody {
				t.Errorf("handler returned wrong body: got %v want %v", rr.Body.String(), tt.wantBody)
			}

			if tt.wantContentType != "" {
				if ct := rr.Header().Get("Content-Type"); ct != tt.wantContentType {
					t.Errorf("handler returned wrong content type: got %v want %v", ct, tt.wantContentType)
				}
			}

			// Check security headers on index
			if tt.path == "/" && rr.Code == http.StatusOK {
				if rr.Header().Get("X-Frame-Options") == "" {
					t.Error("Security headers missing")
				}
			}
		})
	}
}

func TestRegisterStaticRoutes_DirNotFound(t *testing.T) {
	// Should log warning but not panic
	mux := http.NewServeMux()
	config := RouterConfig{
		Mux: mux,
		Deps: Dependencies{
			StaticDir: "/path/to/non/existent/dir",
		},
	}
	registerStaticRoutes(config)

	// Verify routes are NOT registered (or 404)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected 404 when static dir missing, got %v", rr.Code)
	}
}

func createFile(t *testing.T, path string, content string) {
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create file %s: %v", path, err)
	}
}
