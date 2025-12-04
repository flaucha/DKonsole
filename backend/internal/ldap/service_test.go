package ldap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-ldap/ldap/v3"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

// mockRepository is a mock implementation of Repository for testing
type mockRepository struct {
	config             *models.LDAPConfig
	groups             *models.LDAPGroupsConfig
	username           string
	password           string
	configErr          error
	groupsErr          error
	credsErr           error
	updateErr          error
	updateGroupsErr    error
	updateCredsErr     error
	updateConfigCalled bool
	updateGroupsCalled bool
	updateCredsCalled  bool
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
	m.updateConfigCalled = true
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
		m.updateGroupsCalled = true
		return m.updateGroupsErr
	}
	m.updateGroupsCalled = true
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
		m.updateCredsCalled = true
		return m.updateCredsErr
	}
	m.updateCredsCalled = true
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
			Enabled:       true,
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

func TestGetLDAPStatusHandler_ErrorReturnsDisabled(t *testing.T) {
	service := NewService(&mockRepository{
		configErr: &mockError{msg: "fail"},
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/ldap/status", nil)
	service.GetLDAPStatusHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if enabled, ok := resp["enabled"].(bool); !ok || enabled {
		t.Fatalf("expected enabled=false, got %v", resp["enabled"])
	}
}

func TestUpdateConfigHandler_InvalidURL(t *testing.T) {
	service := NewService(&mockRepository{})
	body := bytes.NewBufferString(`{"config":{"enabled":true,"url":"http://bad"}}`)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/ldap/config", body)

	service.UpdateConfigHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestUpdateGroupsHandler_InvalidPermission(t *testing.T) {
	service := NewService(&mockRepository{})
	body := bytes.NewBufferString(`{"groups":{"groups":[{"name":"dev","permissions":[{"namespace":"default","permission":"admin"}]}]}}`)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/ldap/groups", body)

	service.UpdateGroupsHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestUpdateCredentialsHandler_ValidationPaths(t *testing.T) {
	t.Run("missing username", func(t *testing.T) {
		service := NewService(&mockRepository{})
		body := bytes.NewBufferString(`{"username":"","password":""}`)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/ldap/credentials", body)

		service.UpdateCredentialsHandler(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rr.Code)
		}
	})

	t.Run("username changed without password and none stored", func(t *testing.T) {
		service := NewService(&mockRepository{
			username: "old",
			password: "",
		})
		body := bytes.NewBufferString(`{"username":"new","password":""}`)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/ldap/credentials", body)

		service.UpdateCredentialsHandler(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "password required") {
			t.Fatalf("unexpected body: %s", rr.Body.String())
		}
	})

	t.Run("no changes to save", func(t *testing.T) {
		service := NewService(&mockRepository{
			username: "same",
			password: "secret",
		})
		body := bytes.NewBufferString(`{"username":"same","password":""}`)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/ldap/credentials", body)

		service.UpdateCredentialsHandler(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "No changes") {
			t.Fatalf("unexpected body: %s", rr.Body.String())
		}
	})
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

func TestGetLDAPStatusHandler_Enabled(t *testing.T) {
	service := NewService(&mockRepository{
		config: &models.LDAPConfig{Enabled: true},
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/ldap/status", nil)

	service.GetLDAPStatusHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if enabled, ok := resp["enabled"].(bool); !ok || !enabled {
		t.Fatalf("expected enabled=true, got %v", resp["enabled"])
	}
}

func TestGetConfigHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := &models.LDAPConfig{
			Enabled: true,
			URL:     "ldap://example.com",
			BaseDN:  "dc=example,dc=com",
		}
		repo := &mockRepository{config: expected}
		service := NewService(repo)

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/ldap/config", nil)
		service.GetConfigHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rr.Code)
		}
		var cfg models.LDAPConfig
		if err := json.NewDecoder(rr.Body).Decode(&cfg); err != nil {
			t.Fatalf("decode config: %v", err)
		}
		if cfg.URL != expected.URL || !cfg.Enabled {
			t.Fatalf("unexpected config: %+v", cfg)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		service := NewService(&mockRepository{
			configErr: errors.New("boom"),
		})
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/ldap/config", nil)
		service.GetConfigHandler(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rr.Code)
		}
	})
}

func TestUpdateConfigHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepository{}
		service := NewService(repo)
		body := bytes.NewBufferString(`{"config":{"enabled":false,"url":"ldap://example.com","baseDN":"dc=example,dc=com","userDN":"uid"}}`)

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/ldap/config", body)
		service.UpdateConfigHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rr.Code)
		}
		if !repo.updateConfigCalled {
			t.Fatalf("expected UpdateConfig to be called")
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		service := NewService(&mockRepository{})
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/ldap/config", bytes.NewBufferString("{"))

		service.UpdateConfigHandler(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rr.Code)
		}
	})

	t.Run("update error", func(t *testing.T) {
		service := NewService(&mockRepository{
			updateErr: errors.New("cannot update"),
		})
		body := bytes.NewBufferString(`{"config":{"enabled":false,"url":"ldap://example.com"}}`)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/ldap/config", body)

		service.UpdateConfigHandler(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rr.Code)
		}
	})
}

func TestGetGroupsHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepository{
			groups: &models.LDAPGroupsConfig{
				Groups: []models.LDAPGroup{{Name: "dev"}},
			},
		}
		service := NewService(repo)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/ldap/groups", nil)

		service.GetGroupsHandler(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rr.Code)
		}
		var resp models.LDAPGroupsConfig
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("decode groups: %v", err)
		}
		if len(resp.Groups) != 1 || resp.Groups[0].Name != "dev" {
			t.Fatalf("unexpected groups: %+v", resp.Groups)
		}
	})

	t.Run("error", func(t *testing.T) {
		service := NewService(&mockRepository{groupsErr: errors.New("boom")})
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/ldap/groups", nil)

		service.GetGroupsHandler(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rr.Code)
		}
	})
}

func TestUpdateGroupsHandler(t *testing.T) {
	t.Run("empty group name", func(t *testing.T) {
		repo := &mockRepository{}
		service := NewService(repo)
		body := bytes.NewBufferString(`{"groups":{"groups":[{"name":"","permissions":[{"namespace":"default","permission":"view"}]}]}}`)

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/ldap/groups", body)
		service.UpdateGroupsHandler(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rr.Code)
		}
		if repo.updateGroupsCalled {
			t.Fatalf("expected repository not to be called on validation error")
		}
	})

	t.Run("repo error", func(t *testing.T) {
		service := NewService(&mockRepository{
			updateGroupsErr: errors.New("fail"),
		})
		body := bytes.NewBufferString(`{"groups":{"groups":[{"name":"dev","permissions":[{"namespace":"default","permission":"view"}]}]}}`)

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/ldap/groups", body)
		service.UpdateGroupsHandler(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rr.Code)
		}
	})
}

func TestCredentialsHandlers(t *testing.T) {
	t.Run("get credentials success", func(t *testing.T) {
		service := NewService(&mockRepository{
			username: "admin",
			password: "secret",
		})
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/ldap/credentials", nil)

		service.GetCredentialsHandler(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rr.Code)
		}
		var resp map[string]any
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp["configured"] != true || resp["username"] != "admin" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("get credentials error", func(t *testing.T) {
		service := NewService(&mockRepository{credsErr: errors.New("fail")})
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/ldap/credentials", nil)

		service.GetCredentialsHandler(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rr.Code)
		}
	})

	t.Run("update credentials error", func(t *testing.T) {
		service := NewService(&mockRepository{
			updateCredsErr: errors.New("fail"),
		})
		body := bytes.NewBufferString(`{"username":"admin","password":"secret"}`)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/ldap/credentials", body)

		service.UpdateCredentialsHandler(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rr.Code)
		}
	})
}

func TestTestConnectionHandler_InvalidBody(t *testing.T) {
	service := NewService(&mockRepository{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/ldap/test-connection", bytes.NewBufferString("not-json"))

	service.TestConnectionHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestService_initializeClient(t *testing.T) {
	t.Run("config error", func(t *testing.T) {
		service := NewService(&mockRepository{
			configErr: errors.New("fail"),
		})
		if err := service.initializeClient(context.Background()); err == nil {
			t.Fatalf("expected error from initializeClient")
		}
	})

	t.Run("disabled closes existing client", func(t *testing.T) {
		pool := &connectionPool{
			connections: make(chan *ldap.Conn, 1),
		}
		service := NewService(&mockRepository{
			config: &models.LDAPConfig{Enabled: false},
		})
		service.client = &LDAPClient{pool: pool}

		if err := service.initializeClient(context.Background()); err != nil {
			t.Fatalf("initializeClient returned error: %v", err)
		}
		if service.client != nil {
			t.Fatalf("expected client to be cleared when disabled")
		}

		defer func() {
			if recover() == nil {
				t.Fatalf("expected channel to be closed")
			}
		}()
		pool.connections <- nil
	})
}

func TestService_getClient(t *testing.T) {
	t.Run("returns existing client without hitting repo", func(t *testing.T) {
		service := NewService(&mockRepository{
			configErr: errors.New("should not be called"),
		})
		existing := &LDAPClient{}
		service.client = existing

		client, err := service.getClient(context.Background())
		if err != nil {
			t.Fatalf("getClient returned error: %v", err)
		}
		if client != existing {
			t.Fatalf("expected existing client to be returned")
		}
	})

	t.Run("initialize error is propagated", func(t *testing.T) {
		service := NewService(&mockRepository{
			configErr: errors.New("fail"),
		})

		_, err := service.getClient(context.Background())
		if err == nil {
			t.Fatalf("expected error from getClient")
		}
	})
}

type fakeLDAPClientRepository struct {
	entries    []*ldap.Entry
	err        error
	lastFilter string
}

func (f *fakeLDAPClientRepository) Bind(ctx context.Context, bindDN, password string) error {
	return nil
}

func (f *fakeLDAPClientRepository) Search(ctx context.Context, baseDN string, scope int, filter string, attributes []string) ([]*ldap.Entry, error) {
	f.lastFilter = filter
	if f.err != nil {
		return nil, f.err
	}
	return f.entries, nil
}

func (f *fakeLDAPClientRepository) Close() error {
	return nil
}

func TestSearchUserDN(t *testing.T) {
	t.Run("uses found DN and filter includes user filter", func(t *testing.T) {
		repo := &fakeLDAPClientRepository{
			entries: []*ldap.Entry{{DN: "uid=john,dc=example,dc=com"}},
		}
		config := &models.LDAPConfig{
			UserDN:     "uid",
			BaseDN:     "dc=example,dc=com",
			UserFilter: "(objectClass=person)",
		}

		dn := searchUserDN(context.Background(), repo, config, "john")
		if dn != "uid=john,dc=example,dc=com" {
			t.Fatalf("got %s, want entry DN", dn)
		}
		if repo.lastFilter != "(&(uid=john)(objectClass=person))" {
			t.Fatalf("unexpected filter: %s", repo.lastFilter)
		}
	})

	t.Run("fallback to buildBindDN on search error", func(t *testing.T) {
		repo := &fakeLDAPClientRepository{
			err: errors.New("search failed"),
		}
		config := &models.LDAPConfig{
			UserDN: "uid",
			BaseDN: "dc=example,dc=com",
		}

		dn := searchUserDN(context.Background(), repo, config, "john")
		if dn != "uid=john,dc=example,dc=com" {
			t.Fatalf("unexpected DN: %s", dn)
		}
	})
}

// mockError is a simple error implementation
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
