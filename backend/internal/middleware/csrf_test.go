package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCSRFMiddleware(t *testing.T) {
	// Create a test handler that always succeeds
	nextHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	tests := []struct {
		name           string
		method         string
		origin         string
		referer        string
		wantStatusCode int
		wantBody       string
		shouldAllow    bool
	}{
		{
			name:           "GET request should pass (safe method)",
			method:         "GET",
			origin:         "",
			referer:        "",
			wantStatusCode: http.StatusOK,
			shouldAllow:    true,
		},
		{
			name:           "HEAD request should pass (safe method)",
			method:         "HEAD",
			origin:         "",
			referer:        "",
			wantStatusCode: http.StatusOK,
			shouldAllow:    true,
		},
		{
			name:           "OPTIONS request should pass (safe method)",
			method:         "OPTIONS",
			origin:         "",
			referer:        "",
			wantStatusCode: http.StatusOK,
			shouldAllow:    true,
		},
		{
			name:           "POST with Origin header should pass",
			method:         "POST",
			origin:         "https://example.com",
			referer:        "",
			wantStatusCode: http.StatusOK,
			shouldAllow:    true,
		},
		{
			name:           "POST with Referer header should pass",
			method:         "POST",
			origin:         "",
			referer:        "https://example.com",
			wantStatusCode: http.StatusOK,
			shouldAllow:    true,
		},
		{
			name:           "POST with both Origin and Referer should pass",
			method:         "POST",
			origin:         "https://example.com",
			referer:        "https://example.com/page",
			wantStatusCode: http.StatusOK,
			shouldAllow:    true,
		},
		{
			name:           "POST without Origin or Referer should be rejected",
			method:         "POST",
			origin:         "",
			referer:        "",
			wantStatusCode: http.StatusForbidden,
			wantBody:       "Missing Origin or Referer header",
			shouldAllow:    false,
		},
		{
			name:           "PUT without Origin or Referer should be rejected",
			method:         "PUT",
			origin:         "",
			referer:        "",
			wantStatusCode: http.StatusForbidden,
			wantBody:       "Missing Origin or Referer header",
			shouldAllow:    false,
		},
		{
			name:           "DELETE without Origin or Referer should be rejected",
			method:         "DELETE",
			origin:         "",
			referer:        "",
			wantStatusCode: http.StatusForbidden,
			wantBody:       "Missing Origin or Referer header",
			shouldAllow:    false,
		},
		{
			name:           "PATCH with Origin should pass",
			method:         "PATCH",
			origin:         "https://example.com",
			referer:        "",
			wantStatusCode: http.StatusOK,
			shouldAllow:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CSRFMiddleware(nextHandler)

			req := httptest.NewRequest(tt.method, "/api/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.referer != "" {
				req.Header.Set("Referer", tt.referer)
			}

			rr := httptest.NewRecorder()
			handler(rr, req)

			// Check status code
			if rr.Code != tt.wantStatusCode {
				t.Errorf("CSRFMiddleware() status code = %v, want %v", rr.Code, tt.wantStatusCode)
			}

			// Check body content if specified
			if tt.wantBody != "" {
				body := rr.Body.String()
				if body != tt.wantBody+"\n" {
					t.Errorf("CSRFMiddleware() body = %v, want %v", body, tt.wantBody+"\n")
				}
			}

			// Verify handler was called when should allow
			if tt.shouldAllow && rr.Body.String() != "OK" {
				t.Errorf("CSRFMiddleware() should have called next handler, but body = %v", rr.Body.String())
			}
		})
	}
}
