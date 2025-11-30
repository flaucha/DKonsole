package ldap

import (
	"context"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

// mockRepository is a mock implementation of Repository for testing
type mockRepository struct {
	config      *models.LDAPConfig
	groups      *models.LDAPGroupsConfig
	username    string
	password    string
	configErr   error
	groupsErr   error
	credsErr    error
	updateErr   error
	updateGroupsErr error
	updateCredsErr  error
}

func (m *mockRepository) GetConfig(ctx context.Context) (*models.LDAPConfig, error) {
	if m.configErr != nil {
		return nil, m.configErr
	}
	if m.config != nil {
		return m.config, nil
	}
	return &models.LDAPConfig{Enabled: false}, nil
}

func (m *mockRepository) UpdateConfig(ctx context.Context, config *models.LDAPConfig) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.config = config
	return nil
}

func (m *mockRepository) GetGroups(ctx context.Context) (*models.LDAPGroupsConfig, error) {
	if m.groupsErr != nil {
		return nil, m.groupsErr
	}
	if m.groups != nil {
		return m.groups, nil
	}
	return &models.LDAPGroupsConfig{Groups: []models.LDAPGroup{}}, nil
}

func (m *mockRepository) UpdateGroups(ctx context.Context, groups *models.LDAPGroupsConfig) error {
	if m.updateGroupsErr != nil {
		return m.updateGroupsErr
	}
	m.groups = groups
	return nil
}

func (m *mockRepository) GetCredentials(ctx context.Context) (username, password string, err error) {
	if m.credsErr != nil {
		return "", "", m.credsErr
	}
	return m.username, m.password, nil
}

func (m *mockRepository) UpdateCredentials(ctx context.Context, username, password string) error {
	if m.updateCredsErr != nil {
		return m.updateCredsErr
	}
	m.username = username
	m.password = password
	return nil
}

func TestIsValidLDAPURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "valid ldap URL",
			url:  "ldap://example.com",
			want: true,
		},
		{
			name: "valid ldaps URL",
			url:  "ldaps://example.com",
			want: true,
		},
		{
			name: "valid ldap URL with port",
			url:  "ldap://example.com:389",
			want: true,
		},
		{
			name: "valid ldaps URL with port",
			url:  "ldaps://example.com:636",
			want: true,
		},
		{
			name: "invalid URL - http",
			url:  "http://example.com",
			want: false,
		},
		{
			name: "invalid URL - empty",
			url:  "",
			want: false,
		},
		{
			name: "invalid URL - no protocol",
			url:  "example.com",
			want: false,
		},
		{
			name: "invalid URL - wrong protocol",
			url:  "ldapx://example.com",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle panic for short strings
			if len(tt.url) < 7 && tt.url != "" {
				if isValidLDAPURL(tt.url) {
					t.Errorf("isValidLDAPURL() = %v, want false for short URL", true)
				}
				return
			}

			got := isValidLDAPURL(tt.url)
			if got != tt.want {
				t.Errorf("isValidLDAPURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_AuthenticateUser_LDAPDisabled(t *testing.T) {
	mockRepo := &mockRepository{
		config: &models.LDAPConfig{Enabled: false},
	}

	service := NewService(mockRepo)
	ctx := context.Background()

	err := service.AuthenticateUser(ctx, "testuser", "password")
	if err == nil {
		t.Errorf("AuthenticateUser() expected error when LDAP is disabled")
		return
	}

	if err.Error() != "LDAP is not enabled" {
		t.Errorf("AuthenticateUser() error = %v, want 'LDAP is not enabled'", err)
	}
}

func TestService_AuthenticateUser_InvalidUsername(t *testing.T) {
	mockRepo := &mockRepository{
		config: &models.LDAPConfig{Enabled: true},
	}

	service := NewService(mockRepo)
	ctx := context.Background()

	// Test with invalid username (LDAP injection attempt)
	err := service.AuthenticateUser(ctx, "admin*", "password")
	if err == nil {
		t.Errorf("AuthenticateUser() expected error for invalid username")
		return
	}

	if err.Error() == "" {
		t.Errorf("AuthenticateUser() error should not be empty")
	}
}

func TestService_ValidateUserGroup_NoRequiredGroup(t *testing.T) {
	mockRepo := &mockRepository{
		config: &models.LDAPConfig{
			Enabled:      true,
			RequiredGroup: "", // No required group
		},
	}

	service := NewService(mockRepo)
	ctx := context.Background()

	err := service.ValidateUserGroup(ctx, "testuser")
	if err != nil {
		t.Errorf("ValidateUserGroup() error = %v, want nil when no required group", err)
	}
}

func TestService_ValidateUserGroup_ConfigError(t *testing.T) {
	mockRepo := &mockRepository{
		configErr: &mockError{msg: "config error"},
	}

	service := NewService(mockRepo)
	ctx := context.Background()

	err := service.ValidateUserGroup(ctx, "testuser")
	if err == nil {
		t.Errorf("ValidateUserGroup() expected error when config fails")
	}
}

func TestService_GetConfig(t *testing.T) {
	expectedConfig := &models.LDAPConfig{
		Enabled: true,
		URL:     "ldap://example.com",
		BaseDN:  "dc=example,dc=com",
	}

	mockRepo := &mockRepository{
		config: expectedConfig,
	}

	service := NewService(mockRepo)
	ctx := context.Background()

	config, err := service.GetConfig(ctx)
	if err != nil {
		t.Errorf("GetConfig() error = %v, want nil", err)
		return
	}

	if config == nil {
		t.Errorf("GetConfig() config is nil")
		return
	}

	if config.Enabled != expectedConfig.Enabled {
		t.Errorf("GetConfig() Enabled = %v, want %v", config.Enabled, expectedConfig.Enabled)
	}

	if config.URL != expectedConfig.URL {
		t.Errorf("GetConfig() URL = %v, want %v", config.URL, expectedConfig.URL)
	}
}

func TestService_GetUserPermissions_LDAPDisabled(t *testing.T) {
	mockRepo := &mockRepository{
		config: &models.LDAPConfig{Enabled: false},
	}

	service := NewService(mockRepo)
	ctx := context.Background()

	_, err := service.GetUserPermissions(ctx, "testuser")
	if err == nil {
		t.Errorf("GetUserPermissions() expected error when LDAP is disabled")
		return
	}
}

func TestService_GetUserPermissions_ConfigError(t *testing.T) {
	mockRepo := &mockRepository{
		configErr: &mockError{msg: "config error"},
	}

	service := NewService(mockRepo)
	ctx := context.Background()

	_, err := service.GetUserPermissions(ctx, "testuser")
	if err == nil {
		t.Errorf("GetUserPermissions() expected error when config fails")
	}
}

func TestBuildBindDN(t *testing.T) {
	config := &models.LDAPConfig{
		UserDN: "uid",
		BaseDN: "dc=example,dc=com",
	}

	tests := []struct {
		name     string
		username string
		want     string
	}{
		{
			name:     "username without equals (simple)",
			username: "johndoe",
			want:     "uid=johndoe,dc=example,dc=com",
		},
		{
			name:     "username with equals (full DN)",
			username: "uid=johndoe,dc=example,dc=com",
			want:     "uid=johndoe,dc=example,dc=com",
		},
		{
			name:     "username is already full DN",
			username: "cn=admin,ou=users,dc=example,dc=com",
			want:     "cn=admin,ou=users,dc=example,dc=com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildBindDN(tt.username, config)
			if got != tt.want {
				t.Errorf("buildBindDN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeLDAPFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDiff bool // Whether escaping changes the input
	}{
		{
			name:     "safe input",
			input:    "johndoe",
			wantDiff: false,
		},
		{
			name:     "input with wildcard",
			input:    "admin*",
			wantDiff: true,
		},
		{
			name:     "input with parentheses",
			input:    "admin)(cn=*",
			wantDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeLDAPFilter(tt.input)

			diff := got != tt.input
			if diff != tt.wantDiff {
				t.Errorf("sanitizeLDAPFilter() changed input = %v, want %v (got: %v, input: %v)", diff, tt.wantDiff, got, tt.input)
			}

			// Sanitized output should never be empty for non-empty input
			if tt.input != "" && got == "" {
				t.Errorf("sanitizeLDAPFilter() returned empty string for non-empty input")
			}
		})
	}
}

// mockError is a simple error implementation
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
