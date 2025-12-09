package ldap

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-ldap/ldap/v3"
)

func TestUpdateConfigHandler_Detailed(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockRepo       *mockRepository
		setupMockConn  func()
		wantStatusCode int
	}{
		{
			name:           "Invalid JSON",
			body:           `{invalid}`,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Invalid URL",
			body:           `{"config": {"enabled": true, "url": "invalid://url"}}`,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Repo Error",
			body: `{"config": {"enabled": true, "url": "ldap://example.com"}}`,
			mockRepo: &mockRepository{
				updateErr: errors.New("update fail"),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:     "Success",
			body:     `{"config": {"enabled": true, "url": "ldap://example.com", "insecureSkipVerify": true}}`,
			mockRepo: &mockRepository{},
			setupMockConn: func() {
				ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
					return &mockLDAPConnection{}, nil
				}
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.mockRepo
			if repo == nil {
				repo = &mockRepository{}
			}

			if tt.setupMockConn != nil {
				originalDialer := ldapDialer
				defer func() { ldapDialer = originalDialer }()
				tt.setupMockConn()
			} else {
				// Prevent actual dialing if not set up
				originalDialer := ldapDialer
				defer func() { ldapDialer = originalDialer }()
				ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
					return &mockLDAPConnection{}, nil
				}
			}

			s := NewService(repo)
			req := httptest.NewRequest("POST", "/config", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			s.UpdateConfigHandler(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestUpdateGroupsHandler_Detailed(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockRepo       *mockRepository
		wantStatusCode int
	}{
		{
			name:           "Invalid JSON",
			body:           `{invalid}`,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Empty Group Name",
			body:           `{"groups": {"groups": [{"name": ""}]}}`,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Invalid Permission",
			body:           `{"groups": {"groups": [{"name": "admins", "permissions": [{"namespace": "default", "permission": "invalid"}]}]}}`,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Repo Error",
			body: `{"groups": {"groups": [{"name": "admins"}]}}`,
			mockRepo: &mockRepository{
				updateGroupsErr: errors.New("db error"),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:           "Success with filtering empty namespaces",
			body:           `{"groups": {"groups": [{"name": "admins", "permissions": [{"namespace": "", "permission": "view"}, {"namespace": "default", "permission": "edit"}]}]}}`,
			mockRepo:       &mockRepository{},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.mockRepo
			if repo == nil {
				repo = &mockRepository{}
			}
			s := NewService(repo)
			req := httptest.NewRequest("POST", "/groups", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			s.UpdateGroupsHandler(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestUpdateCredentialsHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockRepo       *mockRepository
		wantStatusCode int
	}{
		{
			name:           "Invalid JSON",
			body:           `{invalid}`,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Empty Username",
			body:           `{"username": ""}`,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Repo Error",
			body: `{"username": "admin", "password": "new"}`,
			mockRepo: &mockRepository{
				updateCredsErr: errors.New("db error"),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:           "Success New Password",
			body:           `{"username": "admin", "password": "new"}`,
			mockRepo:       &mockRepository{},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "Keep Existing Password - Success",
			body: `{"username": "admin", "password": ""}`,
			mockRepo: &mockRepository{
				username: "admin",
				password: "old",
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "Keep Existing Password - Fail (Empty existing)",
			body: `{"username": "new", "password": ""}`,
			mockRepo: &mockRepository{
				username: "old",
				password: "", // No password stored
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Keep Existing Password - Repo Get Error",
			body: `{"username": "admin", "password": ""}`,
			mockRepo: &mockRepository{
				credsErr: errors.New("read error"),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.mockRepo
			if repo == nil {
				repo = &mockRepository{}
			}
			s := NewService(repo)
			req := httptest.NewRequest("POST", "/credentials", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			s.UpdateCredentialsHandler(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestTestConnectionHandler_Errors(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupMockConn  func()
		mockRepo       *mockRepository
		wantStatusCode int
	}{
		{
			name:           "Invalid JSON",
			body:           `{invalid}`,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Connection Fail",
			body: `{"url": "ldap://example.com"}`, // Using provided params
			setupMockConn: func() {
				ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
					return nil, errors.New("conn fail")
				}
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalDialer := ldapDialer
			defer func() { ldapDialer = originalDialer }()

			if tt.setupMockConn != nil {
				tt.setupMockConn()
			} else {
				// Default mock if needed, or if expected to fail before dial
				ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
					return &mockLDAPConnection{}, nil
				}
			}

			repo := tt.mockRepo
			if repo == nil {
				repo = &mockRepository{}
			}

			s := NewService(repo)
			req := httptest.NewRequest("POST", "/test-connection", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			s.TestConnectionHandler(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

// Add test for GetConfigHandler error
func TestGetConfigHandler_Error(t *testing.T) {
	repo := &mockRepository{
		configErr: errors.New("db error"),
	}
	s := NewService(repo)
	req := httptest.NewRequest("GET", "/config", nil)
	rr := httptest.NewRecorder()

	s.GetConfigHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got status %v, want 500", rr.Code)
	}
}

// Add test for GetGroupsHandler error
func TestGetGroupsHandler_Error(t *testing.T) {
	repo := &mockRepository{
		groupsErr: errors.New("db error"),
	}
	s := NewService(repo)
	req := httptest.NewRequest("GET", "/groups", nil)
	rr := httptest.NewRecorder()

	s.GetGroupsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got status %v, want 500", rr.Code)
	}
}

// Add test for GetCredentialsHandler error
func TestGetCredentialsHandler_Error(t *testing.T) {
	repo := &mockRepository{
		credsErr: errors.New("db error"),
	}
	s := NewService(repo)
	req := httptest.NewRequest("GET", "/credentials", nil)
	rr := httptest.NewRecorder()

	s.GetCredentialsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("got status %v, want 500", rr.Code)
	}
}
