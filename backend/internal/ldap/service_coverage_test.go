package ldap

import (
	"context"
	"errors"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/go-ldap/ldap/v3"
)

func TestService_InitializeClient_Error(t *testing.T) {
	// Repo failing to get config
	repo := &mockRepository{
		configErr: errors.New("config error"),
	}

	service := NewService(repo)

	// Ensure we don't hit network if it somehow progresses
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()
	ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
		return &mockLDAPConnection{}, nil
	}

	err := service.initializeClient(context.Background())
	if err == nil {
		t.Error("expected error when repo fails")
	}
}

func TestService_GetClient_Reinit(t *testing.T) {
	// Mock dialer to prevent network call
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()
	ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
		return &mockLDAPConnection{}, nil
	}

	// Test race condition protection or lazy init if applicable
	repo := &mockRepository{
		config: &models.LDAPConfig{URL: "ldap://example.com", Enabled: true},
	}
	service := NewService(repo)

	client, err := service.getClient(context.Background())
	if err != nil {
		t.Fatalf("getClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("expected client")
	}
}

func TestService_InitializeClient_UpdateError(t *testing.T) {
	// Test case: Client exists, config update fails (fake it), should recreate client
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()
	ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
		return &mockLDAPConnection{}, nil
	}

	repo := &mockRepository{
		config: &models.LDAPConfig{URL: "ldap://old", Enabled: true},
	}
	service := NewService(repo)

	// Init first time
	service.initializeClient(context.Background())

	// Corrupt config to cause failure?
	// UpdateConfig in client_test didn't fail easily except for TLS.
	// We need invalid replacement config.
	repo.config = &models.LDAPConfig{URL: "ldaps://new", CACert: "bad", Enabled: true}

	// Call initializeClient again (triggered by e.g. getClient or explicit)
	// initializeClient will call s.client.UpdateConfig with new config.
	// client.UpdateConfig calls buildTLSConfig which fails for bad CACert.
	// So UpdateConfig returns error.
	// initializeClient logs warning and recreates client calling NewLDAPClient.
	// NewLDAPClient will ALSO fail because of bad config!
	// So initializeClient should return error ultimately?

	err := service.initializeClient(context.Background())
	if err == nil {
		t.Error("expected error when recreation fails too")
	}
}

func TestService_TestConnection_BindError(t *testing.T) {
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()
	ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
		return &mockLDAPConnection{
			bindFunc: func(u, p string) error { return errors.New("bind fail") },
		}, nil
	}

	service := NewService(&mockRepository{})
	req := TestConnectionRequest{
		URL:      "ldap://example.com",
		Username: "user",
		Password: "pass",
	}

	err := service.TestConnection(context.Background(), req)
	if err == nil {
		t.Error("expected error on bind failure")
	}
}
