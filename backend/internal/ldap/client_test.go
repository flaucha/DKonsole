package ldap

import (
	"errors"
	"testing"
	"time"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/go-ldap/ldap/v3"
)

func TestExtractServerName(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{url: "ldap://example.com:389", want: "example.com"},
		{url: "ldaps://ldap.example.com", want: "ldap.example.com"},
		{url: "ldap://example.com", want: "example.com"},
		{url: "bad", want: ""},
		{url: "", want: ""},
	}

	for _, tt := range tests {
		if got := extractServerName(tt.url); got != tt.want {
			t.Fatalf("extractServerName(%q)=%q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestBuildTLSConfig(t *testing.T) {
	config := &models.LDAPConfig{
		URL:                "ldaps://ldap.example.com:636",
		InsecureSkipVerify: true,
	}

	tlsConfig, err := buildTLSConfig(config)
	if err != nil {
		t.Fatalf("buildTLSConfig returned error: %v", err)
	}
	if tlsConfig == nil {
		t.Fatalf("expected tlsConfig to be non-nil")
	}
	if !tlsConfig.InsecureSkipVerify {
		t.Fatalf("expected InsecureSkipVerify to be true")
	}
	if tlsConfig.ServerName != "ldap.example.com" {
		t.Fatalf("server name = %s, want ldap.example.com", tlsConfig.ServerName)
	}
}

func TestBuildTLSConfig_InvalidCA(t *testing.T) {
	config := &models.LDAPConfig{
		URL:    "ldaps://ldap.example.com",
		CACert: "invalid pem",
	}

	if _, err := buildTLSConfig(config); err == nil {
		t.Fatalf("expected error for invalid CA cert")
	}
}

func TestNewLDAPClient_NilConfig(t *testing.T) {
	if _, err := NewLDAPClient(nil); err == nil {
		t.Fatalf("expected error when config is nil")
	}
}

func TestNewLDAPClient_InvalidTLS(t *testing.T) {
	config := &models.LDAPConfig{
		URL:    "ldaps://example.com",
		CACert: "invalid",
	}
	if _, err := NewLDAPClient(config); err == nil {
		t.Error("expected error with invalid CA cert")
	}
}

func TestLDAPClient_UpdateConfig(t *testing.T) {
	// Setup mock dialer
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()
	
	ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
		return &mockLDAPConnection{}, nil
	}

	config := &models.LDAPConfig{
		URL: "ldap://old.com",
	}
	client, err := NewLDAPClient(config)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	if client.config.URL != "ldap://old.com" {
		t.Errorf("initial url = %s, want ldap://old.com", client.config.URL)
	}

	newConfig := &models.LDAPConfig{
		URL:                "ldaps://new.com",
		InsecureSkipVerify: true,
	}
	if err := client.UpdateConfig(newConfig); err != nil {
		t.Fatalf("UpdateConfig failed: %v", err)
	}

	if client.config.URL != "ldaps://new.com" {
		t.Errorf("updated url = %s, want ldaps://new.com", client.config.URL)
	}
	if !client.tlsConfig.InsecureSkipVerify {
		t.Error("expected TLS config to be updated")
	}
}

func TestLDAPClient_UpdateConfig_Error(t *testing.T) {
	// Setup mock dialer
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()
	ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
		return &mockLDAPConnection{}, nil
	}

	config := &models.LDAPConfig{URL: "ldap://example.com"}
	client, _ := NewLDAPClient(config)
	
	// Try updating with invalid CA cert which causes buildTLSConfig to fail
	badConfig := &models.LDAPConfig{
		URL: "ldaps://example.com",
		CACert: "bad",
	}
	
	if err := client.UpdateConfig(badConfig); err == nil {
		t.Error("expected error when updating with invalid config")
	}
}

func TestConnectionPool_ReturnConnection_Full(t *testing.T) {
	// Setup mock dialer
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()
	
	closeCalled := false
	mockConn := &mockLDAPConnection{
		closeFunc: func() {
			closeCalled = true
		},
	}
	ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
		return mockConn, nil
	}

	// Create pool with size 1
	pool := newConnectionPool("ldap://example.com", nil, time.Second, 1) // Pre-populates 1 (capped at max 1)
	
	// Get connection out (channel now empty)
	conn, _ := pool.getConnection()
	
	// Return it (channel has 1)
	pool.returnConnection(conn)
	
	// Create another mock conn to simulate "extra"
	extraConn := &mockLDAPConnection{
		closeFunc: func() {
			closeCalled = true
		},
	}
	
	// Try to return extra conn (channel full)
	pool.returnConnection(extraConn)
	
	if !closeCalled {
		t.Error("expected connection to be closed when returned to full pool")
	}
}

func TestConnectionPool_CreateConnection_Error(t *testing.T) {
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()
	
	ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
		return nil, errors.New("dial fail")
	}
	
	pool := &connectionPool{
		url: "ldaps://example.com",
		tlsConfig: nil,
	}
	
	if _, err := pool.createConnection(); err == nil {
		t.Error("expected error on dial failure")
	}
}

func TestConnectionPool_ReturnConnection_Nil(t *testing.T) {
	pool := &connectionPool{}
	// Should not panic or do anything
	pool.returnConnection(nil)
}

func TestLDAPClient_Delegation(t *testing.T) {
	// Tests simple delegation methods
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()
	ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
		return &mockLDAPConnection{}, nil
	}
	
	config := &models.LDAPConfig{URL: "ldap://example.com"}
	client, _ := NewLDAPClient(config)
	
	conn, err := client.GetConnection()
	if err != nil {
		t.Fatalf("GetConnection fail: %v", err)
	}
	
	client.ReturnConnection(conn)
}
