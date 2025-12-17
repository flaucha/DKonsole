package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func mustGenerateTestCACertPEM(t *testing.T) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}))
}

func TestCreateEphemeralClient_TLSFailClosedWhenNoCA(t *testing.T) {
	origInCluster := inClusterConfigFunc
	defer func() { inClusterConfigFunc = origInCluster }()
	inClusterConfigFunc = func() (*rest.Config, error) { // Force fallback path
		return nil, errors.New("forced")
	}

	t.Setenv("KUBERNETES_SERVICE_HOST", "kubernetes.default.svc")
	t.Setenv("KUBERNETES_SERVICE_PORT", "443")
	t.Setenv("K8S_CA_PEM", "")
	t.Setenv("K8S_INSECURE_SKIP_VERIFY", "")
	t.Setenv("GO_ENV", "")

	_, err := createEphemeralClient("dummy-token")
	if err == nil {
		t.Fatalf("expected error when CA is missing and insecure skip is not enabled")
	}
}

func TestCreateEphemeralClient_AllowsExplicitCA(t *testing.T) {
	origInCluster := inClusterConfigFunc
	defer func() { inClusterConfigFunc = origInCluster }()
	inClusterConfigFunc = func() (*rest.Config, error) { // Force fallback path
		return nil, errors.New("forced")
	}

	t.Setenv("KUBERNETES_SERVICE_HOST", "kubernetes.default.svc")
	t.Setenv("KUBERNETES_SERVICE_PORT", "443")
	t.Setenv("K8S_CA_PEM", mustGenerateTestCACertPEM(t))
	t.Setenv("K8S_INSECURE_SKIP_VERIFY", "")
	t.Setenv("GO_ENV", "")

	client, err := createEphemeralClient("dummy-token")
	if err != nil {
		t.Fatalf("expected success, got err=%v", err)
	}
	if client == nil {
		t.Fatalf("expected non-nil client")
	}
}

func TestCreateEphemeralClient_DisallowsInsecureSkipVerifyInProduction(t *testing.T) {
	origInCluster := inClusterConfigFunc
	defer func() { inClusterConfigFunc = origInCluster }()
	inClusterConfigFunc = func() (*rest.Config, error) { // Force fallback path
		return nil, errors.New("forced")
	}

	t.Setenv("KUBERNETES_SERVICE_HOST", "kubernetes.default.svc")
	t.Setenv("KUBERNETES_SERVICE_PORT", "443")
	t.Setenv("K8S_CA_PEM", "")
	t.Setenv("K8S_INSECURE_SKIP_VERIFY", "true")
	t.Setenv("GO_ENV", "production")

	_, err := createEphemeralClient("dummy-token")
	if err == nil {
		t.Fatalf("expected error when insecure skip verify is enabled in production")
	}
}

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
