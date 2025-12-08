package ldap

import (
	"context"
	"errors"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/go-ldap/ldap/v3"
)

// mockLDAPConnection implements LDAPConnection
type mockLDAPConnection struct {
	bindFunc   func(username, password string) error
	closeFunc  func()
	searchFunc func(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error)
}

func (m *mockLDAPConnection) Bind(username, password string) error {
	if m.bindFunc != nil {
		return m.bindFunc(username, password)
	}
	return nil
}

func (m *mockLDAPConnection) Close() error {
	if m.closeFunc != nil {
		m.closeFunc()
	}
	return nil
}

func (m *mockLDAPConnection) Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	if m.searchFunc != nil {
		return m.searchFunc(searchRequest)
	}
	return &ldap.SearchResult{}, nil
}

func TestConnectAndBindService(t *testing.T) {
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()

	t.Run("GetCredentials Error", func(t *testing.T) {
		service := NewService(&mockRepository{
			credsErr: errors.New("creds error"),
		})
		_, err := service.connectAndBindService(context.Background(), &models.LDAPConfig{}, "user")
		if err == nil {
			t.Error("expected error when getting credentials fails")
		}
	})

	t.Run("Dial Error", func(t *testing.T) {
		service := NewService(&mockRepository{username: "admin", password: "password"})
		ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
			return nil, errors.New("dial error")
		}
		_, err := service.connectAndBindService(context.Background(), &models.LDAPConfig{}, "user")
		if err == nil {
			t.Error("expected error when dial fails")
		}
	})

	t.Run("Bind Error", func(t *testing.T) {
		service := NewService(&mockRepository{username: "admin", password: "password"})
		ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
			return &mockLDAPConnection{
				bindFunc: func(u, p string) error { return errors.New("bind error") },
			}, nil
		}
		_, err := service.connectAndBindService(context.Background(), &models.LDAPConfig{}, "user")
		if err == nil {
			t.Error("expected error when bind fails")
		}
	})

	t.Run("Success", func(t *testing.T) {
		service := NewService(&mockRepository{username: "admin", password: "password"})
		mockConn := &mockLDAPConnection{}
		ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
			return mockConn, nil
		}
		conn, err := service.connectAndBindService(context.Background(), &models.LDAPConfig{}, "user")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if conn != mockConn {
			t.Error("expected mock connection")
		}
	})
}

func TestVerifyUserPassword(t *testing.T) {
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()
	config := &models.LDAPConfig{URL: "ldap://example.com"}

	t.Run("Dial Error", func(t *testing.T) {
		service := NewService(&mockRepository{})
		ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
			return nil, errors.New("dial error")
		}
		err := service.verifyUserPassword(config, "cn=user", "pass", "user")
		if err == nil {
			t.Error("expected error when dial fails")
		}
	})
	
	t.Run("Bind Error", func(t *testing.T) {
		service := NewService(&mockRepository{})
		ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
			return &mockLDAPConnection{
				bindFunc: func(u, p string) error { return errors.New("invalid password") },
			}, nil
		}
		err := service.verifyUserPassword(config, "cn=user", "pass", "user")
		if err == nil {
			t.Error("expected error on bind failure")
		}
	})

	t.Run("Success", func(t *testing.T) {
		service := NewService(&mockRepository{})
		ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
			return &mockLDAPConnection{}, nil
		}
		err := service.verifyUserPassword(config, "cn=user", "pass", "user")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestFindUserDN(t *testing.T) {
	service := NewService(&mockRepository{})
	config := &models.LDAPConfig{UserDN: "uid", BaseDN: "dc=example,dc=com"}

	t.Run("Already Full DN", func(t *testing.T) {
		dn, err := service.findUserDN(nil, config, "uid=john,dc=example,dc=com")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if dn != "uid=john,dc=example,dc=com" {
			t.Errorf("expected DN to be returned as is")
		}
	})

	t.Run("Search Success", func(t *testing.T) {
		conn := &mockLDAPConnection{
			searchFunc: func(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
				return &ldap.SearchResult{
					Entries: []*ldap.Entry{{DN: "cn=john,ou=users,dc=example,dc=com"}},
				}, nil
			},
		}
		dn, err := service.findUserDN(conn, config, "john")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dn != "cn=john,ou=users,dc=example,dc=com" {
			t.Errorf("got %s, want cn=john...", dn)
		}
	})
	
	t.Run("Search Fail (Fallback)", func(t *testing.T) {
		conn := &mockLDAPConnection{
			searchFunc: func(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
				return nil, errors.New("not found")
			},
		}
		dn, err := service.findUserDN(conn, config, "john")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dn != "uid=john,dc=example,dc=com" {
			t.Errorf("got %s, want fallback DN", dn)
		}
	})
}
