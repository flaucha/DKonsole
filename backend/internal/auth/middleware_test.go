package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flaucha/DKonsole/backend/internal/models"
	
	"github.com/golang-jwt/jwt/v5"
)

func TestUserContextKey(t *testing.T) {
	key := UserContextKey()
	if string(key) != "user" {
		t.Errorf("expected context key 'user', got '%s'", key)
	}
}

func TestAuthMiddleware(t *testing.T) {
	// Setup test JWT service
	secret := []byte("secret")
	jwtService := NewJWTService(secret)

	// Helper to create token
	createToken := func(username string, role string) string {
		claims := &AuthClaims{
			Claims: models.Claims{
				Username: username,
				Role:     role,
			},
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				Issuer:    "dkonsole",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		ss, _ := token.SignedString(secret)
		return ss
	}

	tests := []struct {
		name           string
		setupMode      bool
		jwtService     *JWTService
		method         string
		token          string
		expectedStatus int
		checkContext   bool
	}{
		{
			name:           "Setup Mode Active",
			setupMode:      true,
			jwtService:     jwtService,
			method:         "GET",
			expectedStatus: http.StatusPreconditionFailed,
		},
		{
			name:           "JWT Service Not Initialized",
			setupMode:      false,
			jwtService:     nil,
			method:         "GET",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "OPTIONS Request",
			setupMode:      false,
			jwtService:     jwtService,
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No Token",
			setupMode:      false,
			jwtService:     jwtService,
			method:         "GET",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Token",
			setupMode:      false,
			jwtService:     jwtService,
			method:         "GET",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid Admin Token",
			setupMode:      false,
			jwtService:     jwtService,
			method:         "GET",
			token:          createToken("admin", "admin"),
			expectedStatus: http.StatusOK,
			checkContext:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				setupMode:  tt.setupMode,
				jwtService: tt.jwtService,
			}

			// Test handler that checks context
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkContext {
					claims, ok := r.Context().Value(UserContextKey()).(*AuthClaims)
					if !ok || claims == nil {
						t.Error("Claims not found in context")
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					if claims.Username != "admin" {
						t.Errorf("Expected username admin, got %s", claims.Username)
					}
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, "/api/protected", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			rr := httptest.NewRecorder()

			handler := s.AuthMiddleware(nextHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, tt.expectedStatus)
			}
		})
	}
}
