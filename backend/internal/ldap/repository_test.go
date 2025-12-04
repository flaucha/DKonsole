package ldap

import (
	"context"
	"encoding/json"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

func TestK8sRepository_GetConfig_DefaultsWithoutClient(t *testing.T) {
	repo := &K8sRepository{}

	cfg, err := repo.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("GetConfig returned error: %v", err)
	}
	if cfg.Enabled {
		t.Fatalf("expected disabled config when client is nil")
	}
}

func TestK8sRepository_GetConfig_InvalidJSON(t *testing.T) {
	client := k8sfake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ldap-config",
			Namespace: "ns",
		},
		Data: map[string][]byte{
			"ldap-config": []byte("{not-json"),
		},
	})
	repo := &K8sRepository{
		client:     client,
		namespace:  "ns",
		secretName: "ldap-config",
	}

	if _, err := repo.GetConfig(context.Background()); err == nil {
		t.Fatalf("expected error for invalid config JSON")
	}
}

func TestK8sRepository_UpdateConfig(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	repo := &K8sRepository{
		client:     client,
		namespace:  "ns",
		secretName: "ldap-config",
	}
	cfg := &models.LDAPConfig{Enabled: true, URL: "ldap://example.com"}

	if err := repo.UpdateConfig(context.Background(), cfg); err != nil {
		t.Fatalf("UpdateConfig returned error: %v", err)
	}
	secret, err := client.CoreV1().Secrets("ns").Get(context.Background(), "ldap-config", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("secret not created: %v", err)
	}
	var stored models.LDAPConfig
	if err := json.Unmarshal(secret.Data["ldap-config"], &stored); err != nil {
		t.Fatalf("config not stored as JSON: %v", err)
	}
	if !stored.Enabled || stored.URL != cfg.URL {
		t.Fatalf("stored config mismatch: %+v", stored)
	}
}

func TestK8sRepository_UpdateConfig_NoClient(t *testing.T) {
	repo := &K8sRepository{}
	if err := repo.UpdateConfig(context.Background(), &models.LDAPConfig{}); err == nil {
		t.Fatalf("expected error when client is nil")
	}
}

func TestK8sRepository_GetGroups_Defaults(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	repo := &K8sRepository{
		client:     client,
		namespace:  "ns",
		secretName: "ldap-config",
	}

	groups, err := repo.GetGroups(context.Background())
	if err != nil {
		t.Fatalf("GetGroups returned error: %v", err)
	}
	if len(groups.Groups) != 0 {
		t.Fatalf("expected empty groups, got %v", groups.Groups)
	}
}

func TestK8sRepository_UpdateGroups(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	repo := &K8sRepository{
		client:     client,
		namespace:  "ns",
		secretName: "ldap-config",
	}
	groups := &models.LDAPGroupsConfig{
		Groups: []models.LDAPGroup{{Name: "dev"}},
	}

	if err := repo.UpdateGroups(context.Background(), groups); err != nil {
		t.Fatalf("UpdateGroups returned error: %v", err)
	}
	secret, err := client.CoreV1().Secrets("ns").Get(context.Background(), "ldap-config", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("secret not created: %v", err)
	}
	if len(secret.Data["ldap-groups"]) == 0 {
		t.Fatalf("expected groups data to be stored")
	}
}

func TestK8sRepository_UpdateGroups_NoClient(t *testing.T) {
	repo := &K8sRepository{}
	if err := repo.UpdateGroups(context.Background(), &models.LDAPGroupsConfig{}); err == nil {
		t.Fatalf("expected error when client is nil")
	}
}

func TestK8sRepository_GetCredentials(t *testing.T) {
	client := k8sfake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ldap-config",
			Namespace: "ns",
		},
		Data: map[string][]byte{
			"ldap-username": []byte("admin"),
			"ldap-password": []byte("secret"),
		},
	})
	repo := &K8sRepository{
		client:     client,
		namespace:  "ns",
		secretName: "ldap-config",
	}

	user, pass, err := repo.GetCredentials(context.Background())
	if err != nil {
		t.Fatalf("GetCredentials returned error: %v", err)
	}
	if user != "admin" || pass != "secret" {
		t.Fatalf("unexpected credentials: %s/%s", user, pass)
	}
}

func TestK8sRepository_GetCredentials_NoClient(t *testing.T) {
	repo := &K8sRepository{}
	if _, _, err := repo.GetCredentials(context.Background()); err == nil {
		t.Fatalf("expected error when client is nil")
	}
}

func TestK8sRepository_UpdateCredentials(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	repo := &K8sRepository{
		client:     client,
		namespace:  "ns",
		secretName: "ldap-config",
	}

	if err := repo.UpdateCredentials(context.Background(), "admin", "secret"); err != nil {
		t.Fatalf("UpdateCredentials returned error: %v", err)
	}
	secret, err := client.CoreV1().Secrets("ns").Get(context.Background(), "ldap-config", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("secret not created: %v", err)
	}
	if secret.StringData["ldap-username"] != "admin" || secret.StringData["ldap-password"] != "secret" {
		t.Fatalf("credentials not stored in StringData: %+v", secret.StringData)
	}
}

func TestK8sRepository_UpdateCredentials_NoClient(t *testing.T) {
	repo := &K8sRepository{}
	if err := repo.UpdateCredentials(context.Background(), "admin", "secret"); err == nil {
		t.Fatalf("expected error when client is nil")
	}
}
