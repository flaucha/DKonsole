package ldap

import (
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/models"
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
