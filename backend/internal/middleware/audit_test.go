package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuditMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		isWebSocket    bool
		userInContext  interface{}
		handlerStatus  int
		wantStatusCode int
	}{
		{
			name:           "GET request without user",
			method:         "GET",
			path:           "/api/test",
			isWebSocket:    false,
			userInContext:  nil,
			handlerStatus:  http.StatusOK,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "POST request with user in context (map format)",
			method:         "POST",
			path:           "/api/resources",
			isWebSocket:    false,
			userInContext:  map[string]interface{}{"username": "admin"},
			handlerStatus:  http.StatusOK,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "request with 404 status",
			method:         "GET",
			path:           "/api/notfound",
			isWebSocket:    false,
			userInContext:  nil,
			handlerStatus:  http.StatusNotFound,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "request with 500 status",
			method:         "POST",
			path:           "/api/error",
			isWebSocket:    false,
			userInContext:  nil,
			handlerStatus:  http.StatusInternalServerError,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:           "websocket request should pass",
			method:         "GET",
			path:           "/ws/pods",
			isWebSocket:    true,
			userInContext:  nil,
			handlerStatus:  http.StatusSwitchingProtocols,
			wantStatusCode: http.StatusSwitchingProtocols,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler that sets specific status
			nextHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.handlerStatus)
				w.Write([]byte("Response"))
			}

			handler := AuditMiddleware(nextHandler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.isWebSocket {
				req.Header.Set("Upgrade", "websocket")
				req.Header.Set("Connection", "Upgrade")
			}

			// Add user to context if provided
			if tt.userInContext != nil {
				ctx := context.WithValue(req.Context(), userContextKeyStr, tt.userInContext)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			start := time.Now()
			handler(rr, req)
			duration := time.Since(start)

			// Check status code
			if rr.Code != tt.wantStatusCode {
				t.Errorf("AuditMiddleware() status code = %v, want %v", rr.Code, tt.wantStatusCode)
			}

			// Verify response body
			if rr.Body.String() != "Response" {
				t.Errorf("AuditMiddleware() body = %v, want 'Response'", rr.Body.String())
			}

			// Verify reasonable duration (should be very fast)
			if duration > 100*time.Millisecond {
				t.Errorf("AuditMiddleware() took too long: %v", duration)
			}
		})
	}
}

func TestAuditMiddleware_UserExtraction(t *testing.T) {
	nextHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	handler := AuditMiddleware(nextHandler)

	tests := []struct {
		name          string
		userInContext interface{}
		expectPass    bool
	}{
		{
			name:          "user as map[string]interface{}",
			userInContext: map[string]interface{}{"username": "admin"},
			expectPass:    true,
		},
		{
			name:          "no user in context",
			userInContext: nil,
			expectPass:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/test", nil)

			if tt.userInContext != nil {
				ctx := context.WithValue(req.Context(), userContextKeyStr, tt.userInContext)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("AuditMiddleware() status code = %v, want %v", rr.Code, http.StatusOK)
			}
		})
	}
}

func TestStatusRecorder(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		wantStatusCode int
	}{
		{
			name:           "record 200 status",
			statusCode:     http.StatusOK,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "record 404 status",
			statusCode:     http.StatusNotFound,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "record 500 status",
			statusCode:     http.StatusInternalServerError,
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			recorder := &StatusRecorder{
				ResponseWriter: rr,
				Status:         http.StatusOK,
			}

			recorder.WriteHeader(tt.statusCode)

			if recorder.Status != tt.wantStatusCode {
				t.Errorf("StatusRecorder.Status = %v, want %v", recorder.Status, tt.wantStatusCode)
			}

			if rr.Code != tt.wantStatusCode {
				t.Errorf("ResponseWriter.Code = %v, want %v", rr.Code, tt.wantStatusCode)
			}
		})
	}
}
