package settings

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/prometheus"
)

// mockRepository is a mock implementation of Repository
type mockRepository struct {
	getPrometheusURLFunc    func(ctx context.Context) (string, error)
	updatePrometheusURLFunc func(ctx context.Context, url string) error
}

func (m *mockRepository) GetPrometheusURL(ctx context.Context) (string, error) {
	if m.getPrometheusURLFunc != nil {
		return m.getPrometheusURLFunc(ctx)
	}
	return "", nil
}

func (m *mockRepository) UpdatePrometheusURL(ctx context.Context, url string) error {
	if m.updatePrometheusURLFunc != nil {
		return m.updatePrometheusURLFunc(ctx, url)
	}
	return nil
}

// mockPrometheusService is a mock that implements the UpdateURL method
// We'll test that it's called, but we can't easily mock the full prometheus.HTTPHandler
// For now, we'll pass nil and just verify the service doesn't crash

func TestService_GetPrometheusURLHandler(t *testing.T) {
	tests := []struct {
		name           string
		getURLFunc     func(ctx context.Context) (string, error)
		wantStatusCode int
		wantURL        string
		wantErrMsg     string
	}{
		{
			name: "successful get URL",
			getURLFunc: func(ctx context.Context) (string, error) {
				return "http://prometheus:9090", nil
			},
			wantStatusCode: http.StatusOK,
			wantURL:        "http://prometheus:9090",
		},
		{
			name: "empty URL",
			getURLFunc: func(ctx context.Context) (string, error) {
				return "", nil
			},
			wantStatusCode: http.StatusOK,
			wantURL:        "",
		},
		{
			name: "repository error",
			getURLFunc: func(ctx context.Context) (string, error) {
				return "", context.DeadlineExceeded
			},
			wantStatusCode: http.StatusInternalServerError,
			wantErrMsg:     "Failed to get Prometheus URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockRepository{
				getPrometheusURLFunc: tt.getURLFunc,
			}

			service := NewService(mockRepo, &models.Handlers{}, nil)

			req := httptest.NewRequest("GET", "/api/settings/prometheus/url", nil)
			w := httptest.NewRecorder()

			service.GetPrometheusURLHandler(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("GetPrometheusURLHandler() status code = %v, want %v", w.Code, tt.wantStatusCode)
				return
			}

			if tt.wantStatusCode == http.StatusOK {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("GetPrometheusURLHandler() failed to decode response: %v", err)
					return
				}

				if response["url"] != tt.wantURL {
					t.Errorf("GetPrometheusURLHandler() url = %v, want %v", response["url"], tt.wantURL)
				}
			} else {
				body := w.Body.String()
				if tt.wantErrMsg != "" && !strings.Contains(body, tt.wantErrMsg) {
					t.Errorf("GetPrometheusURLHandler() error message should contain %v, got %v", tt.wantErrMsg, body)
				}
			}
		})
	}
}

func TestService_UpdatePrometheusURLHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		updateURLFunc  func(ctx context.Context, url string) error
		wantStatusCode int
		wantErrMsg     string
		expectedURL    string
		expectPromCall bool
	}{
		{
			name: "successful update URL",
			requestBody: map[string]string{
				"url": "http://prometheus:9090",
			},
			updateURLFunc: func(ctx context.Context, url string) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
			expectedURL:    "http://prometheus:9090",
			expectPromCall: true,
		},
		{
			name: "successful update HTTPS URL",
			requestBody: map[string]string{
				"url": "https://prometheus.example.com",
			},
			updateURLFunc: func(ctx context.Context, url string) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
			expectedURL:    "https://prometheus.example.com",
		},
		{
			name: "empty URL (reset)",
			requestBody: map[string]string{
				"url": "",
			},
			updateURLFunc: func(ctx context.Context, url string) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
			expectedURL:    "",
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			wantStatusCode: http.StatusBadRequest,
			wantErrMsg:     "invalid request body",
		},
		{
			name: "URL without protocol",
			requestBody: map[string]string{
				"url": "prometheus:9090",
			},
			wantStatusCode: http.StatusBadRequest,
			wantErrMsg:     "invalid URL format. Must start with http:// or https://",
		},
		{
			name: "invalid URL format",
			requestBody: map[string]string{
				"url": "http://[invalid",
			},
			updateURLFunc: func(ctx context.Context, url string) error {
				return nil
			},
			wantStatusCode: http.StatusBadRequest,
			wantErrMsg:     "invalid URL format",
		},
		{
			name: "repository update error",
			requestBody: map[string]string{
				"url": "http://prometheus:9090",
			},
			updateURLFunc: func(ctx context.Context, url string) error {
				return context.DeadlineExceeded
			},
			wantStatusCode: http.StatusInternalServerError,
			wantErrMsg:     "Failed to update Prometheus URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockRepository{
				updatePrometheusURLFunc: tt.updateURLFunc,
			}

			handlersModel := &models.Handlers{}
			promHandler := prometheus.NewHTTPHandler("", nil)

			service := NewService(mockRepo, handlersModel, promHandler)

			var bodyBytes []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/api/settings/prometheus/url", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			service.UpdatePrometheusURLHandler(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("UpdatePrometheusURLHandler() status code = %v, want %v", w.Code, tt.wantStatusCode)
				return
			}

			if tt.wantStatusCode == http.StatusOK {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("UpdatePrometheusURLHandler() failed to decode response: %v", err)
					return
				}

				if response["url"] != tt.expectedURL {
					t.Errorf("UpdatePrometheusURLHandler() url = %v, want %v", response["url"], tt.expectedURL)
				}

				// Verify handlersModel was updated
				handlersModel.RLock()
				if handlersModel.PrometheusURL != tt.expectedURL {
					t.Errorf("UpdatePrometheusURLHandler() handlersModel.PrometheusURL = %v, want %v", handlersModel.PrometheusURL, tt.expectedURL)
				}
				handlersModel.RUnlock()

				if tt.expectPromCall {
					statusRR := httptest.NewRecorder()
					statusReq := httptest.NewRequest(http.MethodGet, "/prom/status", nil)
					promHandler.GetStatus(statusRR, statusReq)
					var statusResp map[string]any
					if err := json.NewDecoder(statusRR.Body).Decode(&statusResp); err != nil {
						t.Fatalf("failed to decode prom status: %v", err)
					}
					if statusResp["url"] != tt.expectedURL {
						t.Errorf("UpdatePrometheusURLHandler() prometheus url = %v, want %v", statusResp["url"], tt.expectedURL)
					}
				}
			} else {
				body := w.Body.String()
				if tt.wantErrMsg != "" && !strings.Contains(body, tt.wantErrMsg) {
					t.Errorf("UpdatePrometheusURLHandler() error message should contain %v, got %v", tt.wantErrMsg, body)
				}
			}
		})
	}
}
