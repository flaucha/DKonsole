package auth

import (
	"context"
	"os"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestNewService(t *testing.T) {
	namespace := "default"
	secretName := "dkonsole-auth" //nolint:gosec // Test secret name

	// Make the test robust in environments where the serviceaccount namespace file exists.
	if ns, err := getCurrentNamespace(); err == nil && ns != "" {
		namespace = ns
	} else {
		os.Setenv("POD_NAMESPACE", namespace)
		defer os.Unsetenv("POD_NAMESPACE")
	}

	t.Run("Initialize with K8s client - Secret Missing (Setup Mode)", func(t *testing.T) {
		client := k8sfake.NewSimpleClientset()
		service, err := NewService(client, secretName)
		if err != nil {
			t.Fatalf("NewService failed: %v", err)
		}
		if !service.IsSetupMode() {
			t.Error("Expected setup mode to be true")
		}
		if service.k8sRepo == nil {
			t.Error("Expected K8s repo to be initialized")
		}
	})

	t.Run("Initialize with K8s client - Secret Exists", func(t *testing.T) {
		client := k8sfake.NewSimpleClientset()
		// Create secret
		if _, err := client.CoreV1().Secrets(namespace).Create(context.Background(), &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"admin-username":      []byte("admin"),
				"admin-password-hash": []byte("hash"),
				"jwt-secret":          []byte("secret-key-must-be-32-bytes-long-123"),
			},
		}, metav1.CreateOptions{}); err != nil {
			t.Fatalf("failed to create secret: %v", err)
		}

		service, err := NewService(client, secretName)
		if err != nil {
			t.Fatalf("NewService failed: %v", err)
		}
		if service.IsSetupMode() {
			t.Error("Expected setup mode to be false")
		}
		if service.authService == nil {
			t.Error("Expected authService to be initialized")
		}
	})

	t.Run("Initialize with K8s client - Secret Exists but Missing JWT Secret (Setup Mode)", func(t *testing.T) {
		client := k8sfake.NewSimpleClientset()
		// Create secret without jwt-secret (placeholder/incomplete)
		if _, err := client.CoreV1().Secrets(namespace).Create(context.Background(), &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"admin-username":      []byte("admin"),
				"admin-password-hash": []byte("hash"),
			},
		}, metav1.CreateOptions{}); err != nil {
			t.Fatalf("failed to create secret: %v", err)
		}

		service, err := NewService(client, secretName)
		if err != nil {
			t.Fatalf("NewService failed: %v", err)
		}
		if !service.IsSetupMode() {
			t.Error("Expected setup mode to be true")
		}
		if service.k8sRepo == nil {
			t.Error("Expected K8s repo to be initialized")
		}
		if service.authService != nil || service.jwtService != nil {
			t.Error("Expected auth services to be nil in setup mode")
		}
	})

	t.Run("Initialize without K8s client (Env Fallback)", func(t *testing.T) {
		os.Setenv("ADMIN_USER", "envadmin")
		os.Setenv("ADMIN_PASSWORD", "envhash")
		defer os.Unsetenv("ADMIN_USER")
		defer os.Unsetenv("ADMIN_PASSWORD")

		origJWT := jwtSecret
		jwtSecret = []byte(strings.Repeat("b", 32))
		t.Cleanup(func() { jwtSecret = origJWT })

		service, err := NewService(nil, "")
		if err != nil {
			t.Fatalf("NewService failed: %v", err)
		}
		if service.IsSetupMode() {
			t.Error("Expected setup mode to be false")
		}
		if service.k8sRepo != nil {
			t.Error("Expected K8s repo to be nil")
		}
	})

	t.Run("Production without configured JWT secret fails fast", func(t *testing.T) {
		t.Setenv("GO_ENV", "production")

		origJWT := jwtSecret
		jwtSecret = nil
		t.Cleanup(func() { jwtSecret = origJWT })

		_, err := NewService(nil, "")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
