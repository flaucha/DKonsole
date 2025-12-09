package auth

import (
	"context"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestNewService(t *testing.T) {
	// Setup K8s fake client
	client := k8sfake.NewSimpleClientset()
	namespace := "default"
	secretName := "dkonsole-auth" //nolint:gosec // Test secret name

	// Set env for namespace detection
	os.Setenv("POD_NAMESPACE", namespace)
	defer os.Unsetenv("POD_NAMESPACE")

	t.Run("Initialize with K8s client - Secret Missing (Setup Mode)", func(t *testing.T) {
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
		// Create secret
		client.CoreV1().Secrets(namespace).Create(context.Background(), &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"admin-username":      []byte("admin"),
				"admin-password-hash": []byte("hash"),
				"jwt-secret":          []byte("secret-key-must-be-32-bytes-long-123"),
			},
		}, metav1.CreateOptions{})

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

	t.Run("Initialize without K8s client (Env Fallback)", func(t *testing.T) {
		os.Setenv("ADMIN_USER", "envadmin")
		os.Setenv("ADMIN_PASSWORD_HASH", "envhash")
		os.Setenv("JWT_SECRET", "envsecret")
		defer os.Unsetenv("ADMIN_USER")
		defer os.Unsetenv("ADMIN_PASSWORD_HASH")
		defer os.Unsetenv("JWT_SECRET")

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
}
