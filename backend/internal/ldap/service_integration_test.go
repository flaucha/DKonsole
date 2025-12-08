package ldap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/go-ldap/ldap/v3"
)

func TestService_AuthenticateUser_Integration(t *testing.T) {
	// ... (same content, just renamed function)
	// Mock repo with valid config and credentials
	mockRepo := &mockRepository{
		config: &models.LDAPConfig{
			Enabled: true,
			URL:     "ldap://example.com",
			BaseDN:  "dc=example,dc=com",
			UserDN:  "uid",
		},
		username: "admin",
		password: "password",
	}
	service := NewService(mockRepo)
	ctx := context.Background()

	// Backup and restore dialer
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()

	t.Run("Success", func(t *testing.T) {
		mockConn := &mockLDAPConnection{
			bindFunc: func(u, p string) error {
				return nil
			},
			searchFunc: func(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
				// Return user entry
				return &ldap.SearchResult{
					Entries: []*ldap.Entry{{DN: "uid=testuser,dc=example,dc=com"}},
				}, nil
			},
		}
		ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
			return mockConn, nil
		}

		if err := service.AuthenticateUser(ctx, "testuser", "password"); err != nil {
			t.Errorf("AuthenticateUser() unexpected error: %v", err)
		}
	})

	t.Run("Service Bind Error", func(t *testing.T) {
		mockConn := &mockLDAPConnection{
			bindFunc: func(u, p string) error {
				return errors.New("bind fail")
			},
		}
		ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
			return mockConn, nil
		}

		if err := service.AuthenticateUser(ctx, "testuser", "password"); err == nil {
			t.Error("expected error on service bind fail")
		}
	})

	t.Run("User Bind Error (Invalid Password)", func(t *testing.T) {
		mockConn := &mockLDAPConnection{
			bindFunc: func(u, p string) error {
				if u == "uid=testuser,dc=example,dc=com" {
					return errors.New("invalid credentials")
				}
				return nil // Service bind success
			},
			searchFunc: func(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
				return &ldap.SearchResult{
					Entries: []*ldap.Entry{{DN: "uid=testuser,dc=example,dc=com"}},
				}, nil
			},
		}
		ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
			return mockConn, nil
		}

		err := service.AuthenticateUser(ctx, "testuser", "wrong")
		if err == nil {
			t.Error("expected error on user bind fail")
		}
	})
	
	t.Run("Dial Error", func(t *testing.T) {
		ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
			return nil, errors.New("dial fail")
		}
		if err := service.AuthenticateUser(ctx, "testuser", "password"); err == nil {
			t.Error("expected error on dial fail")
		}
	})
}

func TestTestConnectionHandler_Integration(t *testing.T) {
	service := NewService(&mockRepository{})
	
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()

	t.Run("Success", func(t *testing.T) {
		ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
			return &mockLDAPConnection{}, nil
		}
		
		body := bytes.NewBufferString(`{"config":{"enabled":true,"url":"ldap://example.com","userDN":"uid","baseDN":"dc=com"},"username":"admin","password":"password"}`)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/ldap/test-connection", body)

		service.TestConnectionHandler(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rr.Code)
		}
		
		var resp map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &resp)
		if msg, ok := resp["message"].(string); !ok || msg != "LDAP connection test successful" {
				t.Errorf("expected success message, got %v", resp)
		}
	})
}

func TestService_GetUserGroups_Integration(t *testing.T) {
	mockRepo := &mockRepository{
		config: &models.LDAPConfig{Enabled: true, URL: "l://e", BaseDN: "dc=c", UserDN: "uid"},
		username: "admin", password: "pwd",
	}
	service := NewService(mockRepo)
	ctx := context.Background()

	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()
	
	// Create mock connection for search
	mockConn := &mockLDAPConnection{
		searchFunc: func(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
			// 1. searchUserDN -> returns user entry with memberOf?
			// 2. searchGroups -> returns groups?
			
			// If request filter contains (uid=john), return user
			// We can verify filter string
			return &ldap.SearchResult{Entries: []*ldap.Entry{
				{DN: "uid=john,dc=c", Attributes: []*ldap.EntryAttribute{
					{Name: "memberOf", Values: []string{"cn=admins,ou=groups,dc=c"}},
				}},
			}}, nil
		},
	}
	ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
		return mockConn, nil
	}

	groups, err := service.GetUserGroups(ctx, "john")
	if err != nil {
		t.Fatalf("GetUserGroups error: %v", err)
	}
	if len(groups) == 0 {
		t.Error("expected groups")
	}
	// Check content
	if groups[0] != "admins" {
		t.Errorf("got group %s, want admins", groups[0])
	}
}

func TestService_GetUserPermissions_Integration(t *testing.T) {
	// This wrapper needs GetUserGroups success + config mapping
	mockRepo := &mockRepository{
		config: &models.LDAPConfig{
			Enabled: true, URL: "l://e", BaseDN: "dc=c", UserDN: "uid",
			AdminGroups: []string{"admins"},
		},
		groups: &models.LDAPGroupsConfig{},
		username: "admin", password: "pwd",
	}
	service := NewService(mockRepo)
	ctx := context.Background()

	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()
	
	mockConn := &mockLDAPConnection{
		searchFunc: func(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
			return &ldap.SearchResult{Entries: []*ldap.Entry{
				{DN: "uid=john,dc=c", Attributes: []*ldap.EntryAttribute{
					{Name: "memberOf", Values: []string{"cn=admins,ou=groups,dc=c"}},
				}},
			}}, nil
		},
	}
	ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
		return mockConn, nil
	}

	perms, err := service.GetUserPermissions(ctx, "john")
	if err != nil {
		t.Fatalf("GetUserPermissions error: %v", err)
	}
	if perms != nil {
		t.Errorf("expected nil perms for admin user (full access), got %v", perms)
	}
}
