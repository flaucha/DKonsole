package auth

import (
	"context"
	"testing"
	"time"

	"github.com/example/k8s-view/internal/models"
)

// mockUserRepository is a mock implementation of UserRepository for testing
type mockUserRepository struct {
	adminUser     string
	adminPassword string
}

func (m *mockUserRepository) GetAdminUser() (string, error) {
	if m.adminUser == "" {
		return "", ErrAdminUserNotSet
	}
	return m.adminUser, nil
}

func (m *mockUserRepository) GetAdminPasswordHash() (string, error) {
	if m.adminPassword == "" {
		return "", ErrAdminPasswordNotSet
	}
	return m.adminPassword, nil
}

func TestAuthService_Login(t *testing.T) {
	testPassword := "testpassword123"
	// Generate a valid hash for testing
	testPasswordHash := generateArgon2Hash(testPassword)

	tests := []struct {
		name          string
		adminUser     string
		adminPassword string
		req           LoginRequest
		wantErr       bool
		errMsg        string
		checkToken    bool
	}{
		{
			name:          "successful login",
			adminUser:     "admin",
			adminPassword: testPasswordHash,
			req: LoginRequest{
				Username: "admin",
				Password: testPassword,
			},
			wantErr:    false,
			checkToken: true,
		},
		{
			name:          "invalid username",
			adminUser:     "admin",
			adminPassword: testPasswordHash,
			req: LoginRequest{
				Username: "wronguser",
				Password: testPassword,
			},
			wantErr: true,
			errMsg:  "Invalid credentials",
		},
		{
			name:          "invalid password",
			adminUser:     "admin",
			adminPassword: testPasswordHash,
			req: LoginRequest{
				Username: "admin",
				Password: "wrongpassword",
			},
			wantErr: true,
			errMsg:  "Invalid credentials",
		},
		{
			name:          "missing admin user",
			adminUser:     "",
			adminPassword: testPasswordHash,
			req: LoginRequest{
				Username: "admin",
				Password: testPassword,
			},
			wantErr: true,
			errMsg:  "server configuration error",
		},
		{
			name:          "missing admin password",
			adminUser:     "admin",
			adminPassword: "",
			req: LoginRequest{
				Username: "admin",
				Password: testPassword,
			},
			wantErr: true,
			errMsg:  "server configuration error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := &mockUserRepository{
				adminUser:     tt.adminUser,
				adminPassword: tt.adminPassword,
			}

			// Create service with test JWT secret
			jwtSecret := []byte("test-secret-key-must-be-at-least-32-characters-long")
			service := NewAuthService(mockRepo, jwtSecret)

			// Attempt login
			result, err := service.Login(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("Login() expected error but got nil")
					return
				}
				if tt.errMsg != "" {
					// Check if error message contains expected substring
					errStr := err.Error()
					found := false
					for i := 0; i <= len(errStr)-len(tt.errMsg); i++ {
						if errStr[i:i+len(tt.errMsg)] == tt.errMsg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Login() error = %v, want error containing %v", err, tt.errMsg)
					}
				}
				return
			}

			// Check successful login result
			if result == nil {
				t.Errorf("Login() result is nil")
				return
			}

			if tt.checkToken {
				if result.Token == "" {
					t.Errorf("Login() token is empty")
				}
				if result.Response.Role != "admin" {
					t.Errorf("Login() role = %v, want admin", result.Response.Role)
				}
				if result.Expires.Before(time.Now()) {
					t.Errorf("Login() expiration time is in the past")
				}
				if result.Expires.After(time.Now().Add(25 * time.Hour)) {
					t.Errorf("Login() expiration time is too far in the future")
				}
			}
		})
	}
}

func TestAuthService_GetCurrentUser(t *testing.T) {
	jwtSecret := []byte("test-secret-key-must-be-at-least-32-characters-long")
	mockRepo := &mockUserRepository{
		adminUser:     "admin",
		adminPassword: "$argon2id$v=19$m=65536,t=3,p=2$c29tZXNhbHQ$RdescudvJCsgt3ub+b+dWRWJTmaaJObG",
	}
	service := NewAuthService(mockRepo, jwtSecret)

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "valid user in context",
			ctx:     context.WithValue(context.Background(), "user", &AuthClaims{Claims: models.Claims{Username: "admin", Role: "admin"}}),
			wantErr: false,
		},
		{
			name:    "no user in context",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "invalid user type in context",
			ctx:     context.WithValue(context.Background(), "user", "invalid"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.GetCurrentUser(tt.ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if user == nil {
					t.Errorf("GetCurrentUser() user is nil")
					return
				}
				if user.Username == "" {
					t.Errorf("GetCurrentUser() username is empty")
				}
			}
		})
	}
}
