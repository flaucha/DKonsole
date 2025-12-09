package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/flaucha/DKonsole/backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestChangePasswordHandler_WithK8s(t *testing.T) {
	// Setup K8s fake client
	client := k8sfake.NewSimpleClientset()
	namespace := "default"
	secretName := "dkonsole-auth" //nolint:gosec // Test secret name

	// Set env for namespace detection
	os.Setenv("POD_NAMESPACE", namespace)
	defer os.Unsetenv("POD_NAMESPACE")

	// Create initial secret
	password := "oldpassword123"
	hash, _ := HashPassword(password)

	client.CoreV1().Secrets(namespace).Create(context.Background(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"admin-username":      []byte("admin"),
			"admin-password-hash": []byte(hash),
			"jwt-secret":          []byte("secret-key-must-be-32-bytes-long-123"),
		},
	}, metav1.CreateOptions{})

	k8sRepo, err := NewK8sUserRepository(client, secretName)
	if err != nil {
		t.Fatalf("failed to create k8s repo: %v", err)
	}

	authService := NewAuthService(k8sRepo, []byte("secret-key-must-be-32-bytes-long-123"))
	jwtService := NewJWTService([]byte("secret-key-must-be-32-bytes-long-123"))

	service := &Service{
		authService: authService,
		jwtService:  jwtService,
		k8sRepo:     k8sRepo,
		setupMode:   false,
	}

	// Helper to create valid token for admin
	createToken := func() string {
		claims := &AuthClaims{
			Claims: models.Claims{
				Username: "admin",
				Role:     "admin",
			},
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				Issuer:    "dkonsole",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		ss, _ := token.SignedString([]byte("secret-key-must-be-32-bytes-long-123"))
		return ss
	}

	tests := []struct {
		name           string
		reqBody        map[string]string
		token          string
		expectedStatus int
	}{
		{
			name: "Success",
			reqBody: map[string]string{
				"currentPassword": "oldpassword123",
				"newPassword":     "newpassword123",
			},
			token:          createToken(),
			expectedStatus: http.StatusOK,
		},
		{
			name: "Wrong Current Password",
			reqBody: map[string]string{
				"currentPassword": "wrongpassword",
				"newPassword":     "newpassword123",
			},
			token:          createToken(),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Short New Password",
			reqBody: map[string]string{
				"currentPassword": "oldpassword123",
				"newPassword":     "short",
			},
			token:          createToken(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "No Token",
			reqBody: map[string]string{
				"currentPassword": "oldpassword123",
				"newPassword":     "newpassword123",
			},
			token:          "",
			expectedStatus: http.StatusOK, // AuthMiddleware is skipped here as we test handler logic directly... wait.
			// The handler calls GetCurrentUser(ctx). If no user in context -> Unauthorized.
			// But here we are calling Handler directly without Middleware.
			// So context will be empty.
			// authService.GetCurrentUser(ctx) will fail.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest("POST", "/api/change-password", bytes.NewBuffer(body))

			// Inject user into context if token is "valid" (simulating middleware)
			if tt.token != "" {
				claims := &AuthClaims{
					Claims: models.Claims{
						Username: "admin",
						Role:     "admin",
					},
				}
				ctx := context.WithValue(req.Context(), UserContextKey(), claims)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			service.ChangePasswordHandler(rr, req)

			if tt.name == "No Token" { // Expect Unauthorized because user not in context
				if rr.Code != http.StatusUnauthorized {
					t.Errorf("expected 401, got %v", rr.Code)
				}
			} else if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v. Body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}
