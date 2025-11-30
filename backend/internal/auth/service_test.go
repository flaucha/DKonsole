package auth

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/flaucha/DKonsole/backend/internal/models"
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

// mockLDAPAuthenticator is a mock implementation of LDAPAuthenticator for testing
type mockLDAPAuthenticator struct {
	authenticateUserErr    error
	validateUserGroupErr   error
	getUserPermissionsFunc func(ctx context.Context, username string) (map[string]string, error)
	getUserGroupsFunc      func(ctx context.Context, username string) ([]string, error)
	getConfigFunc          func(ctx context.Context) (*models.LDAPConfig, error)
}

func (m *mockLDAPAuthenticator) AuthenticateUser(ctx context.Context, username, password string) error {
	return m.authenticateUserErr
}

func (m *mockLDAPAuthenticator) ValidateUserGroup(ctx context.Context, username string) error {
	return m.validateUserGroupErr
}

func (m *mockLDAPAuthenticator) GetUserPermissions(ctx context.Context, username string) (map[string]string, error) {
	if m.getUserPermissionsFunc != nil {
		return m.getUserPermissionsFunc(ctx, username)
	}
	return nil, nil
}

func (m *mockLDAPAuthenticator) GetUserGroups(ctx context.Context, username string) ([]string, error) {
	if m.getUserGroupsFunc != nil {
		return m.getUserGroupsFunc(ctx, username)
	}
	return nil, nil
}

func (m *mockLDAPAuthenticator) GetConfig(ctx context.Context) (*models.LDAPConfig, error) {
	if m.getConfigFunc != nil {
		return m.getConfigFunc(ctx)
	}
	return nil, nil
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
			ctx:     context.WithValue(context.Background(), userContextKey, &AuthClaims{Claims: models.Claims{Username: "admin", Role: "admin"}}),
			wantErr: false,
		},
		{
			name:    "no user in context",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "invalid user type in context",
			ctx:     context.WithValue(context.Background(), userContextKey, "invalid"),
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

func TestAuthService_LoginWithLDAP(t *testing.T) {
	jwtSecret := []byte("test-secret-key-must-be-at-least-32-characters-long")
	mockRepo := &mockUserRepository{}

	tests := []struct {
		name              string
		ldapAuth          *mockLDAPAuthenticator
		req               LoginRequest
		wantErr           bool
		errMsg            string
		checkToken        bool
		expectedRole      string
		expectedPerms     map[string]string
	}{
		{
			name: "successful LDAP login with no permissions (empty map)",
			ldapAuth: &mockLDAPAuthenticator{
				authenticateUserErr:  nil,
				validateUserGroupErr: nil,
				getUserPermissionsFunc: func(ctx context.Context, username string) (map[string]string, error) {
					return map[string]string{}, nil // Empty permissions = user with no access
				},
			},
			req: LoginRequest{
				Username: "ldapuser",
				Password: "ldappass",
				IDP:      "ldap",
			},
			wantErr:      false,
			checkToken:   true,
			expectedRole: "user",
			expectedPerms: map[string]string{}, // User with empty permissions
		},
		{
			name: "successful LDAP login with admin permissions (nil)",
			ldapAuth: &mockLDAPAuthenticator{
				authenticateUserErr:  nil,
				validateUserGroupErr: nil,
				getUserPermissionsFunc: func(ctx context.Context, username string) (map[string]string, error) {
					return nil, nil // Nil permissions = admin
				},
			},
			req: LoginRequest{
				Username: "ldapuser",
				Password: "ldappass",
				IDP:      "ldap",
			},
			wantErr:      false,
			checkToken:   true,
			expectedRole: "admin",
			expectedPerms: nil, // Admin has nil permissions
		},
		{
			name: "successful LDAP login with user permissions",
			ldapAuth: &mockLDAPAuthenticator{
				authenticateUserErr:  nil,
				validateUserGroupErr: nil,
				getUserPermissionsFunc: func(ctx context.Context, username string) (map[string]string, error) {
					return map[string]string{
						"namespace1": "view",
						"namespace2": "edit",
					}, nil
				},
			},
			req: LoginRequest{
				Username: "ldapuser",
				Password: "ldappass",
				IDP:      "ldap",
			},
			wantErr:      false,
			checkToken:   true,
			expectedRole: "user",
			expectedPerms: map[string]string{
				"namespace1": "view",
				"namespace2": "edit",
			},
		},
		{
			name: "LDAP authentication failed",
			ldapAuth: &mockLDAPAuthenticator{
				authenticateUserErr:  fmt.Errorf("invalid credentials"),
				validateUserGroupErr: nil,
			},
			req: LoginRequest{
				Username: "ldapuser",
				Password: "wrongpass",
				IDP:      "ldap",
			},
			wantErr: true,
			errMsg:  "Invalid credentials",
		},
		{
			name:     "LDAP disabled",
			ldapAuth: nil,
			req: LoginRequest{
				Username: "ldapuser",
				Password: "ldappass",
				IDP:      "ldap",
			},
			wantErr: true,
			errMsg:  "Invalid credentials",
		},
		{
			name: "user not in required group",
			ldapAuth: &mockLDAPAuthenticator{
				authenticateUserErr:  nil,
				validateUserGroupErr: fmt.Errorf("user not in required group"),
			},
			req: LoginRequest{
				Username: "ldapuser",
				Password: "ldappass",
				IDP:      "ldap",
			},
			wantErr: true,
			errMsg:  "Invalid credentials",
		},
		{
			name: "failed to get permissions, continues as user",
			ldapAuth: &mockLDAPAuthenticator{
				authenticateUserErr:  nil,
				validateUserGroupErr: nil,
				getUserPermissionsFunc: func(ctx context.Context, username string) (map[string]string, error) {
					return nil, fmt.Errorf("failed to get permissions")
				},
			},
			req: LoginRequest{
				Username: "ldapuser",
				Password: "ldappass",
				IDP:      "ldap",
			},
			wantErr:      false,
			checkToken:   true,
			expectedRole: "user",
			expectedPerms: map[string]string{}, // Empty map when error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewAuthService(mockRepo, jwtSecret)
			if tt.ldapAuth != nil {
				service.SetLDAPAuthenticator(tt.ldapAuth)
			}

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
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Login() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if result == nil {
				t.Errorf("Login() result is nil")
				return
			}

			if tt.checkToken {
				if result.Token == "" {
					t.Errorf("Login() token is empty")
				}
				if result.Response.Role != tt.expectedRole {
					t.Errorf("Login() role = %v, want %v", result.Response.Role, tt.expectedRole)
				}
				if result.Expires.Before(time.Now()) {
					t.Errorf("Login() expiration time is in the past")
				}
			}
		})
	}
}
