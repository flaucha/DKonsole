package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func TestJWTService_ValidateToken(t *testing.T) {
	jwtSecret := []byte("test-secret-key-must-be-at-least-32-characters-long")
	service := NewJWTService(jwtSecret)

	tests := []struct {
		name      string
		tokenFunc func() string
		wantErr   bool
		checkUser bool
		username  string
	}{
		{
			name: "valid token",
			tokenFunc: func() string {
				claims := &AuthClaims{
					Claims: models.Claims{Username: "admin", Role: "admin"},
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(jwtSecret)
				return tokenString
			},
			wantErr:   false,
			checkUser: true,
			username:  "admin",
		},
		{
			name: "expired token",
			tokenFunc: func() string {
				claims := &AuthClaims{
					Claims: models.Claims{Username: "admin", Role: "admin"},
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(jwtSecret)
				return tokenString
			},
			wantErr: true,
		},
		{
			name: "invalid signature",
			tokenFunc: func() string {
				wrongSecret := []byte("wrong-secret-key-must-be-at-least-32-characters-long")
				claims := &AuthClaims{
					Claims: models.Claims{Username: "admin", Role: "admin"},
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(wrongSecret)
				return tokenString
			},
			wantErr: true,
		},
		{
			name: "malformed token",
			tokenFunc: func() string {
				return "not.a.valid.token"
			},
			wantErr: true,
		},
		{
			name: "empty token",
			tokenFunc: func() string {
				return ""
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := tt.tokenFunc()
			claims, err := service.ValidateToken(tokenString)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkUser {
				if claims == nil {
					t.Errorf("ValidateToken() claims is nil")
					return
				}
				if claims.Username != tt.username {
					t.Errorf("ValidateToken() username = %v, want %v", claims.Username, tt.username)
				}
			}
		})
	}
}

func TestJWTService_ExtractToken(t *testing.T) {
	jwtSecret := []byte("test-secret-key-must-be-at-least-32-characters-long")
	service := NewJWTService(jwtSecret)
	testToken := "test-token-string"

	tests := []struct {
		name      string
		reqFunc   func() *http.Request
		wantToken string
		wantErr   bool
	}{
		{
			name: "token in Authorization header",
			reqFunc: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Authorization", "Bearer "+testToken)
				return req
			},
			wantToken: testToken,
			wantErr:   false,
		},
		{
			name: "token in query parameter",
			reqFunc: func() *http.Request {
				req := httptest.NewRequest("GET", "/?token="+testToken, nil)
				return req
			},
			wantToken: testToken,
			wantErr:   false,
		},
		{
			name: "token in cookie",
			reqFunc: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.AddCookie(&http.Cookie{Name: "token", Value: testToken})
				return req
			},
			wantToken: testToken,
			wantErr:   false,
		},
		{
			name: "no token",
			reqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/", nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.reqFunc()
			token, err := service.ExtractToken(req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if token != tt.wantToken {
					t.Errorf("ExtractToken() token = %v, want %v", token, tt.wantToken)
				}
			}
		})
	}
}

func TestJWTService_AuthenticateRequest(t *testing.T) {
	jwtSecret := []byte("test-secret-key-must-be-at-least-32-characters-long")
	service := NewJWTService(jwtSecret)

	// Create a valid token
	claims := &AuthClaims{
		Claims: models.Claims{Username: "admin", Role: "admin"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)

	tests := []struct {
		name    string
		reqFunc func() *http.Request
		wantErr bool
	}{
		{
			name: "valid token in header",
			reqFunc: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Authorization", "Bearer "+tokenString)
				return req
			},
			wantErr: false,
		},
		{
			name: "no token",
			reqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/", nil)
			},
			wantErr: true,
		},
		{
			name: "invalid token",
			reqFunc: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Authorization", "Bearer invalid-token")
				return req
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.reqFunc()
			claims, err := service.AuthenticateRequest(req)

			if (err != nil) != tt.wantErr {
				t.Errorf("AuthenticateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if claims == nil {
					t.Errorf("AuthenticateRequest() claims is nil")
				}
			}
		})
	}
}
