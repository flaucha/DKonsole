package ldap

import (
	"testing"

	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func TestConstructors(t *testing.T) {
	t.Run("NewRepository", func(t *testing.T) {
		client := k8sfake.NewSimpleClientset()
		repo := NewRepository(client, "secret")
		if repo == nil {
			t.Error("NewRepository returned nil")
		}
	})

	t.Run("NewService", func(t *testing.T) {
		repo := &mockRepository{}
		svc := NewService(repo)
		if svc == nil {
			t.Error("NewService returned nil")
		}
	})

	t.Run("NewServiceFactory", func(t *testing.T) {
		client := k8sfake.NewSimpleClientset()
		factory := NewServiceFactory(client, "secret")
		if factory == nil {
			t.Error("NewServiceFactory returned nil")
		}

		svc := factory.NewService()
		if svc == nil {
			t.Error("factory.NewService returned nil")
		}
	})

	t.Run("NewLDAPClient", func(t *testing.T) {
		// Nil config
		if _, err := NewLDAPClient(nil); err == nil {
			t.Error("expected error for nil config")
		}

		// Valid
		client, err := NewLDAPClient(&models.LDAPConfig{URL: "ldap://example.com"})
		if err != nil {
			t.Errorf("NewLDAPClient error: %v", err)
		}
		if client == nil {
			t.Error("NewLDAPClient returned nil client")
		}
	})
}
