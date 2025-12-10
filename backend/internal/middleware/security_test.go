package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	// Create a test handler
	handler := SecurityHeadersMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name            string
		path            string
		isHTTPS         bool
		expectedHeaders map[string]string
	}{
		{
			name:    "API route with HTTP",
			path:    "/api/login",
			isHTTPS: false,
			expectedHeaders: map[string]string{
				"X-Content-Type-Options": "nosniff",
				"X-Frame-Options":        "DENY",
				"X-XSS-Protection":       "1; mode=block",
				"Referrer-Policy":        "strict-origin-when-cross-origin",
				"Permissions-Policy":     "geolocation=(), microphone=(), camera=()",
			},
		},
		{
			name:    "API route with HTTPS",
			path:    "/api/resources",
			isHTTPS: true,
			expectedHeaders: map[string]string{
				"X-Content-Type-Options":    "nosniff",
				"X-Frame-Options":           "DENY",
				"X-XSS-Protection":          "1; mode=block",
				"Referrer-Policy":           "strict-origin-when-cross-origin",
				"Permissions-Policy":        "geolocation=(), microphone=(), camera=()",
				"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
			},
		},
		{
			name:    "Static route",
			path:    "/",
			isHTTPS: false,
			expectedHeaders: map[string]string{
				"X-Content-Type-Options": "nosniff",
				"X-Frame-Options":        "DENY",
				"X-XSS-Protection":       "1; mode=block",
				"Referrer-Policy":        "strict-origin-when-cross-origin",
				"Permissions-Policy":     "geolocation=(), microphone=(), camera=()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.isHTTPS {
				req.Header.Set("X-Forwarded-Proto", "https")
			}

			rr := httptest.NewRecorder()
			handler(rr, req)

			// Check status code
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
			}

			// Check all expected headers
			for header, expectedValue := range tt.expectedHeaders {
				actualValue := rr.Header().Get(header)
				if actualValue != expectedValue {
					t.Errorf("header %s: got %v want %v", header, actualValue, expectedValue)
				}
			}

			// Verify HSTS is NOT set for HTTP requests
			if !tt.isHTTPS {
				if hsts := rr.Header().Get("Strict-Transport-Security"); hsts != "" {
					t.Errorf("HSTS header should not be set for HTTP requests, got %v", hsts)
				}
			}
		})
	}
}

func TestBuildCSP(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "API route",
			path:     "/api/login",
			expected: "default-src 'self'; script-src 'none'; style-src 'none'; img-src 'none'; font-src 'none'; connect-src 'self' ws: wss:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		},
		{
			name:     "Static route",
			path:     "/",
			expected: "default-src 'self'; script-src 'self' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net; img-src 'self' data: https:; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self' ws: wss: https://cdn.jsdelivr.net; worker-src 'self' blob:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		},
		{
			name:     "Assets route",
			path:     "/assets/main.js",
			expected: "default-src 'self'; script-src 'self' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net; img-src 'self' data: https:; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self' ws: wss: https://cdn.jsdelivr.net; worker-src 'self' blob:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			csp := buildCSP(req)
			if csp != tt.expected {
				t.Errorf("CSP for path %s: got %v want %v", tt.path, csp, tt.expected)
			}
		})
	}
}

func TestSecurityHeadersHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeadersHandler(mux)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("SecurityHeadersHandler status = %v, want %v", rr.Code, http.StatusOK)
	}

	// Verify at least one security header
	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("SecurityHeadersHandler missing security headers")
	}
}
