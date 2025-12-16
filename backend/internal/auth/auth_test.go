package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func TestLoginHandler(t *testing.T) {
	testPassword := "testpassword123"
	testPasswordHash := generateArgon2Hash(testPassword)
	jwtSecret := []byte("test-secret-key-must-be-at-least-32-characters-long")

	tests := []struct {
		name           string
		setupMode      bool
		authService    *AuthService
		requestBody    interface{}
		wantStatusCode int
		wantErrMsg     string
		checkCookie    bool
		checkResponse  bool
	}{
		{
			name:        "successful login",
			setupMode:   false,
			authService: NewAuthService(&mockUserRepository{adminUser: "admin", adminPassword: testPasswordHash}, jwtSecret),
			requestBody: models.Credentials{
				Username: "admin",
				Password: testPassword,
			},
			wantStatusCode: http.StatusOK,
			checkCookie:    true,
			checkResponse:  true,
		},
		{
			name:           "invalid JSON body",
			setupMode:      false,
			authService:    NewAuthService(&mockUserRepository{adminUser: "admin", adminPassword: testPasswordHash}, jwtSecret),
			requestBody:    "invalid-json",
			wantStatusCode: http.StatusBadRequest,
			wantErrMsg:     "Invalid request body",
		},
		{
			name:        "invalid credentials",
			setupMode:   false,
			authService: NewAuthService(&mockUserRepository{adminUser: "admin", adminPassword: testPasswordHash}, jwtSecret),
			requestBody: models.Credentials{
				Username: "admin",
				Password: "wrongpassword",
			},
			wantStatusCode: http.StatusUnauthorized,
			wantErrMsg:     "Invalid credentials",
		},
		{
			name:           "setup mode blocks login",
			setupMode:      true,
			authService:    nil,
			requestBody:    models.Credentials{Username: "admin", Password: testPassword},
			wantStatusCode: http.StatusPreconditionFailed,
			wantErrMsg:     "Setup required",
		},
		{
			name:           "auth service not initialized",
			setupMode:      false,
			authService:    nil,
			requestBody:    models.Credentials{Username: "admin", Password: testPassword},
			wantStatusCode: http.StatusInternalServerError,
			wantErrMsg:     "Authentication service not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			service := &Service{
				authService: tt.authService,
				setupMode:   tt.setupMode,
			}

			// Prepare request body
			var bodyBytes []byte
			var err error
			if strBody, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(strBody)
			} else {
				bodyBytes, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			// Create request
			req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Call handler
			service.LoginHandler(rr, req)

			// Check status code
			if rr.Code != tt.wantStatusCode {
				t.Errorf("LoginHandler() status code = %v, want %v", rr.Code, tt.wantStatusCode)
			}

			// Check error message if expected
			if tt.wantErrMsg != "" {
				bodyStr := string(rr.Body.Bytes())
				if !strings.Contains(bodyStr, tt.wantErrMsg) {
					t.Errorf("LoginHandler() error message not found, body = %v, want containing %v", bodyStr, tt.wantErrMsg)
				}
			}

			// Check cookie if expected
			if tt.checkCookie {
				cookies := rr.Result().Cookies()
				found := false
				for _, cookie := range cookies {
					if cookie.Name == "token" {
						found = true
						if cookie.Value == "" {
							t.Errorf("LoginHandler() token cookie is empty")
						}
						if !cookie.HttpOnly {
							t.Errorf("LoginHandler() token cookie should be HttpOnly")
						}
						if cookie.SameSite != http.SameSiteNoneMode {
							t.Errorf("LoginHandler() token cookie should be SameSiteNoneMode, got %v", cookie.SameSite)
						}
						break
					}
				}
				if !found {
					t.Errorf("LoginHandler() token cookie not found")
				}
			}

			// Check response if expected
			if tt.checkResponse {
				var response LoginResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("LoginHandler() failed to unmarshal response: %v", err)
				} else {
					if response.Role != "admin" {
						t.Errorf("LoginHandler() role = %v, want admin", response.Role)
					}
				}
			}
		})
	}
}

func TestLogoutHandler(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name           string
		method         string
		wantStatusCode int
	}{
		{
			name:           "successful logout with POST",
			method:         "POST",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid method GET",
			method:         "GET",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid method PUT",
			method:         "PUT",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/logout", nil)
			rr := httptest.NewRecorder()

			service.LogoutHandler(rr, req)

			// Check status code
			if rr.Code != tt.wantStatusCode {
				t.Errorf("LogoutHandler() status code = %v, want %v", rr.Code, tt.wantStatusCode)
			}

			// For successful logout, check cookie and response
			if tt.wantStatusCode == http.StatusOK {
				// Check cookie is cleared
				cookies := rr.Result().Cookies()
				found := false
				for _, cookie := range cookies {
					if cookie.Name == "token" {
						found = true
						if cookie.Value != "" {
							t.Errorf("LogoutHandler() token cookie should be empty, got %v", cookie.Value)
						}
						if cookie.MaxAge != -1 && !cookie.Expires.Before(time.Now()) {
							t.Errorf("LogoutHandler() token cookie should be expired")
						}
						break
					}
				}
				if !found {
					t.Errorf("LogoutHandler() token cookie not found")
				}

				// Check response
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("LogoutHandler() failed to unmarshal response: %v", err)
				} else {
					if message, ok := response["message"]; !ok || message != "Logged out" {
						t.Errorf("LogoutHandler() message = %v, want 'Logged out'", response)
					}
				}
			}
		})
	}
}

func TestMeHandler(t *testing.T) {
	jwtSecret := []byte("test-secret-key-must-be-at-least-32-characters-long")
	mockRepo := &mockUserRepository{
		adminUser:     "admin",
		adminPassword: generateArgon2Hash("testpass"),
	}
	authService := NewAuthService(mockRepo, jwtSecret)

	tests := []struct {
		name           string
		setupMode      bool
		ctx            context.Context
		wantStatusCode int
		wantErrMsg     string
		checkResponse  bool
	}{
		{
			name:      "authenticated user",
			setupMode: false,
			ctx: context.WithValue(
				context.Background(),
				userContextKey,
				&AuthClaims{
					Claims: models.Claims{
						Username:    "admin",
						Role:        "admin",
						Permissions: nil,
					},
				},
			),
			wantStatusCode: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "no user in context",
			setupMode:      false,
			ctx:            context.Background(),
			wantStatusCode: http.StatusUnauthorized,
			wantErrMsg:     "Unauthorized",
		},
		{
			name:           "setup mode blocks access",
			setupMode:      true,
			ctx:            context.Background(),
			wantStatusCode: http.StatusPreconditionFailed,
			wantErrMsg:     "Setup required",
		},
		{
			name:           "auth service not initialized",
			setupMode:      false,
			ctx:            context.Background(),
			wantStatusCode: http.StatusInternalServerError,
			wantErrMsg:     "Authentication service not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				authService: authService,
				setupMode:   tt.setupMode,
			}

			// Override authService if needed for this test
			if tt.name == "auth service not initialized" {
				service.authService = nil
			}

			req := httptest.NewRequest("GET", "/api/me", nil)
			req = req.WithContext(tt.ctx)
			rr := httptest.NewRecorder()

			service.MeHandler(rr, req)

			// Check status code
			if rr.Code != tt.wantStatusCode {
				t.Errorf("MeHandler() status code = %v, want %v", rr.Code, tt.wantStatusCode)
			}

			// Check error message if expected
			if tt.wantErrMsg != "" {
				bodyStr := string(rr.Body.Bytes())
				if !strings.Contains(bodyStr, tt.wantErrMsg) {
					t.Errorf("MeHandler() error message not found, body = %v", bodyStr)
				}
			}

			// Check response if expected
			if tt.checkResponse {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("MeHandler() failed to unmarshal response: %v", err)
				} else {
					if username, ok := response["username"].(string); !ok || username != "admin" {
						t.Errorf("MeHandler() username = %v, want admin", response["username"])
					}
					if role, ok := response["role"].(string); !ok || role != "admin" {
						t.Errorf("MeHandler() role = %v, want admin", response["role"])
					}
					if _, ok := response["permissions"]; !ok {
						t.Errorf("MeHandler() permissions field missing")
					}
				}
			}
		})
	}
}

func TestChangePasswordHandler(t *testing.T) {
	testPassword := "testpassword123"
	testPasswordHash := generateArgon2Hash(testPassword)
	jwtSecret := []byte("test-secret-key-must-be-at-least-32-characters-long")
	mockRepo := &mockUserRepository{
		adminUser:     "admin",
		adminPassword: testPasswordHash,
	}
	authService := NewAuthService(mockRepo, jwtSecret)

	tests := []struct {
		name           string
		setupMode      bool
		k8sRepo        *K8sUserRepository // nil for tests without K8s
		ctx            context.Context
		requestBody    ChangePasswordRequest
		wantStatusCode int
		wantErrMsg     string
		checkResponse  bool
	}{
		{
			name:           "setup mode blocks access",
			setupMode:      true,
			k8sRepo:        nil,
			ctx:            context.Background(),
			requestBody:    ChangePasswordRequest{CurrentPassword: testPassword, NewPassword: "newpass123"},
			wantStatusCode: http.StatusPreconditionFailed,
			wantErrMsg:     "Setup required",
		},
		{
			name:      "K8s repo not available",
			setupMode: false,
			k8sRepo:   nil,
			ctx: context.WithValue(
				context.Background(),
				userContextKey,
				&AuthClaims{
					Claims: models.Claims{
						Username:    "admin",
						Role:        "admin",
						Permissions: nil,
					},
				},
			),
			requestBody:    ChangePasswordRequest{CurrentPassword: testPassword, NewPassword: "newpass123"},
			wantStatusCode: http.StatusInternalServerError,
			wantErrMsg:     "Password change not available",
		},
		{
			name:           "unauthorized user - K8s repo checked first",
			setupMode:      false,
			k8sRepo:        nil,
			ctx:            context.Background(),
			requestBody:    ChangePasswordRequest{CurrentPassword: testPassword, NewPassword: "newpass123"},
			wantStatusCode: http.StatusInternalServerError,
			wantErrMsg:     "Password change not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				authService: authService,
				k8sRepo:     tt.k8sRepo,
				setupMode:   tt.setupMode,
			}

			// Prepare request body
			var bodyBytes []byte
			var err error
			if tt.name == "invalid JSON body" {
				bodyBytes = []byte("invalid-json")
			} else {
				bodyBytes, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/api/auth/change-password", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(tt.ctx)
			rr := httptest.NewRecorder()

			service.ChangePasswordHandler(rr, req)

			// Check status code
			if rr.Code != tt.wantStatusCode {
				t.Errorf("ChangePasswordHandler() status code = %v, want %v", rr.Code, tt.wantStatusCode)
			}

			// Check error message if expected
			if tt.wantErrMsg != "" {
				bodyStr := string(rr.Body.Bytes())
				if !strings.Contains(bodyStr, tt.wantErrMsg) {
					t.Errorf("ChangePasswordHandler() error message not found, body = %v, want containing %v", bodyStr, tt.wantErrMsg)
				}
			}

			// Check response if expected
			if tt.checkResponse {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("ChangePasswordHandler() failed to unmarshal response: %v", err)
				} else {
					if message, ok := response["message"]; !ok || !strings.Contains(message, "Password changed successfully") {
						t.Errorf("ChangePasswordHandler() message = %v, want containing 'Password changed successfully'", response)
					}
				}
			}
		})
	}
}
