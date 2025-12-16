package auth

import (
	"context"
	"os"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestK8sUserRepository_UpdateSecretToken(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	repo := &K8sUserRepository{
		client:     client,
		namespace:  "testns",
		secretName: "dkonsole-auth",
		ClientFactory: func(token string) (kubernetes.Interface, error) {
			return client, nil
		},
	}

	// Secret missing
	if _, err := repo.GetAdminUser(); err == nil {
		t.Fatalf("expected error when secret missing")
	}
}

func TestK8sUserRepository_GetAdminUserAndPassword(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	repo := &K8sUserRepository{client: client, namespace: "testns", secretName: "dkonsole-auth"}

	// Secret missing
	if _, err := repo.GetAdminUser(); err == nil {
		t.Fatalf("expected error when secret missing")
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dkonsole-auth",
			Namespace: "testns",
		},
		Data: map[string][]byte{
			"admin-username":      []byte("admin"),
			"admin-password-hash": []byte("hash"),
			"jwt-secret":          []byte("a-very-long-secret-value-with-32-chars"),
		},
	}
	if _, err := client.CoreV1().Secrets("testns").Create(context.Background(), secret, metav1.CreateOptions{}); err != nil {
		t.Fatalf("failed to seed secret: %v", err)
	}
	if _, err := client.CoreV1().Secrets("testns").Get(context.Background(), "dkonsole-auth", metav1.GetOptions{}); err != nil {
		t.Fatalf("seeded secret not found: %v", err)
	}

	user, err := repo.GetAdminUser()
	if err != nil || user != "admin" {
		t.Fatalf("GetAdminUser = %q, err=%v", user, err)
	}
	pass, err := repo.GetAdminPasswordHash()
	if err != nil || pass != "hash" {
		t.Fatalf("GetAdminPasswordHash = %q, err=%v", pass, err)
	}

	// Remove password hash key to trigger error
	if err := client.CoreV1().Secrets("testns").Delete(context.Background(), "dkonsole-auth", metav1.DeleteOptions{}); err != nil {
		t.Fatalf("failed to delete secret: %v", err)
	}
	secret.Data = map[string][]byte{
		"admin-username": []byte("admin"),
		// password missing
		"jwt-secret": []byte("a-very-long-secret-value-with-32-chars"),
	}
	if _, err := client.CoreV1().Secrets("testns").Create(context.Background(), secret, metav1.CreateOptions{}); err != nil {
		t.Fatalf("failed to seed secret without password: %v", err)
	}

	if _, err := repo.GetAdminPasswordHash(); err == nil {
		t.Fatalf("expected error when password hash missing")
	}
}

func TestK8sUserRepository_SecretExistsAndCreateSecret(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	repo := &K8sUserRepository{
		client:     client,
		namespace:  "ns-create",
		secretName: "dkonsole-auth",
		ClientFactory: func(token string) (kubernetes.Interface, error) {
			return client, nil
		},
	}

	exists, err := repo.SecretExists(context.Background())
	if err != nil {
		t.Fatalf("SecretExists returned error: %v", err)
	}
	if exists {
		t.Fatalf("expected secret to not exist")
	}

	if err := repo.CreateSecret(context.Background(), "admin", "hash", "short", "dummy-token"); err == nil {
		t.Fatalf("expected error for short JWT secret")
	}

	if err := repo.CreateSecret(context.Background(), "admin", "hash", strings.Repeat("a", 32), "dummy-token"); err != nil {
		t.Fatalf("CreateSecret returned error: %v", err)
	}

	exists, err = repo.SecretExists(context.Background())
	if err != nil || !exists {
		t.Fatalf("SecretExists after create = %v, err=%v", exists, err)
	}
}

func TestEnvUserRepository(t *testing.T) {
	repo := NewEnvUserRepository()
	os.Unsetenv("ADMIN_USER")
	os.Unsetenv("ADMIN_PASSWORD")
	if _, err := repo.GetAdminUser(); err == nil {
		t.Fatalf("expected error when ADMIN_USER unset")
	}
	if _, err := repo.GetAdminPasswordHash(); err == nil {
		t.Fatalf("expected error when ADMIN_PASSWORD unset")
	}

	os.Setenv("ADMIN_USER", "admin")
	os.Setenv("ADMIN_PASSWORD", "hash")
	user, _ := repo.GetAdminUser()
	pass, _ := repo.GetAdminPasswordHash()
	if user != "admin" || pass != "hash" {
		t.Fatalf("env repo returned %q/%q", user, pass)
	}
}
