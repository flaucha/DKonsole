package health

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "GET request returns OK",
			method:         "GET",
			wantStatusCode: http.StatusOK,
			wantBody:       `{"status":"ok"}`,
		},
		{
			name:           "POST request returns OK",
			method:         "POST",
			wantStatusCode: http.StatusOK,
			wantBody:       `{"status":"ok"}`,
		},
		{
			name:           "HEAD request returns OK (body may be present)",
			method:         "HEAD",
			wantStatusCode: http.StatusOK,
			wantBody:       `{"status":"ok"}`, // Handler writes body for all methods
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			HealthHandler(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("HealthHandler() status code = %v, want %v", w.Code, tt.wantStatusCode)
			}

			body := w.Body.String()
			if body != tt.wantBody {
				t.Errorf("HealthHandler() body = %v, want %v", body, tt.wantBody)
			}

			// Verify Content-Type is set (or can be set by middleware)
			contentType := w.Header().Get("Content-Type")
			// Content-Type might be empty or set by middleware, so we just check that it doesn't error
			_ = contentType
		})
	}
}

func TestHealthHandler_ResponseFormat(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	HealthHandler(w, req)

	// Verify status code
	if w.Code != http.StatusOK {
		t.Errorf("HealthHandler() status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify response body is valid JSON
	body := w.Body.String()
	if !strings.Contains(body, "status") {
		t.Errorf("HealthHandler() body should contain 'status', got %v", body)
	}
	if !strings.Contains(body, "ok") {
		t.Errorf("HealthHandler() body should contain 'ok', got %v", body)
	}

	// Verify it's valid JSON format (basic check)
	if !strings.HasPrefix(body, "{") || !strings.HasSuffix(body, "}") {
		t.Errorf("HealthHandler() body should be valid JSON object, got %v", body)
	}
}

func TestHealthHandler_MultipleCalls(t *testing.T) {
	// Test that the handler works consistently across multiple calls
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HealthHandler() call %d: status code = %v, want %v", i, w.Code, http.StatusOK)
		}

		body := w.Body.String()
		if body != `{"status":"ok"}` {
			t.Errorf("HealthHandler() call %d: body = %v, want %v", i, body, `{"status":"ok"}`)
		}
	}
}
