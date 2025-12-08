package ldap

import (
	"context"
	"errors"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/go-ldap/ldap/v3"
)

func TestNewLDAPClientRepository(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		originalDialer := ldapDialer
		defer func() { ldapDialer = originalDialer }()
		ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
			return &mockLDAPConnection{}, nil
		}

		client, _ := NewLDAPClient(&models.LDAPConfig{URL: "ldap://e.com"})
		repo, err := NewLDAPClientRepository(client)
		if err != nil {
			t.Fatalf("NewLDAPClientRepository error: %v", err)
		}
		if repo == nil {
			t.Error("expected non-nil repo")
		}
		repo.Close()
	})

	t.Run("Dial Fail", func(t *testing.T) {
		originalDialer := ldapDialer
		defer func() { ldapDialer = originalDialer }()
		ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
			return nil, errors.New("dial fail")
		}

		client, _ := NewLDAPClient(&models.LDAPConfig{URL: "ldap://e.com"})
		// NewLDAPClient pre-populates, but logs warning if fails.
		// GetConnection inside NewLDAPClientRepository will try to create new if pool empty logic or if prepopulation failed.
		
		// If prepopulation failed (it did because we mocked dial fail), pool is empty.
		// GetConnection triggers createConnection which calls dialer which fails.
		
		_, err := NewLDAPClientRepository(client)
		if err == nil {
			t.Error("expected error when connection fails")
		}
	})
}

func TestLDAPClientRepository_Methods(t *testing.T) {
	mockConn := &mockLDAPConnection{}
	// We can manually construct repo for testing methods
	repo := &ldapClientRepository{
		conn: mockConn,
		// client needed for Close -> ReturnConnection
		client: &LDAPClient{pool: &connectionPool{connections: make(chan LDAPConnection, 1)}},
	}

	t.Run("Bind", func(t *testing.T) {
		mockConn.bindFunc = func(u, p string) error {
			if u == "u" && p == "p" { return nil }
			return errors.New("fail")
		}
		if err := repo.Bind(context.Background(), "u", "p"); err != nil {
			t.Error("expected success")
		}
		if err := repo.Bind(context.Background(), "u", "wrong"); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Search", func(t *testing.T) {
		mockConn.searchFunc = func(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
			return &ldap.SearchResult{Entries: []*ldap.Entry{{DN: "dn"}}}, nil
		}
		res, err := repo.Search(context.Background(), "base", ldap.ScopeBaseObject, "filter", nil)
		if err != nil {
			t.Error("expected success")
		}
		if len(res) != 1 {
			t.Error("expected 1 result")
		}
	})
	
	t.Run("Close", func(t *testing.T) {
		// Just ensure it doesn't panic
		if err := repo.Close(); err != nil {
			t.Errorf("Close error: %v", err)
		}
		// Connection should be returned to pool (testable if we inspect pool or mock client better)
	})
	
	t.Run("Nil Conn", func(t *testing.T) {
		nilRepo := &ldapClientRepository{conn: nil}
		if err := nilRepo.Bind(context.Background(), "u", "p"); err == nil {
			t.Error("expected error on nil conn")
		}
		if _, err := nilRepo.Search(context.Background(), "b", 0, "f", nil); err == nil {
			t.Error("expected error on nil conn")
		}
	})
}
