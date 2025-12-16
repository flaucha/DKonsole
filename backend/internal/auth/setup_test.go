package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func newSetupService(t *testing.T, secretExists bool) (*Service, *k8sfake.Clientset) {
	t.Helper()
	client := k8sfake.NewSimpleClientset()
	repo := &K8sUserRepository{
		client:     client,
		namespace:  "setup-ns",
		secretName: "dkonsole-auth",
		ClientFactory: func(token string) (kubernetes.Interface, error) {
			return client, nil
		},
	}
	t.Setenv("POD_NAMESPACE", "setup-ns")

	if secretExists {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      repo.secretName,
				Namespace: repo.namespace,
			},
			Data: map[string][]byte{
				"admin-username":      []byte("admin"),
				"admin-password-hash": []byte("hash"),
				"jwt-secret":          []byte(strings.Repeat("a", 32)),
			},
		}
		if _, err := client.CoreV1().Secrets(repo.namespace).Create(t.Context(), secret, metav1.CreateOptions{}); err != nil {
			t.Fatalf("failed to seed secret: %v", err)
		}
		if exists, err := repo.SecretExists(t.Context()); err != nil || !exists {
			t.Fatalf("expected seeded secret, exists=%v err=%v", exists, err)
		}
	}

	svc := &Service{
		k8sRepo:    repo,
		k8sClient:  client,
		secretName: repo.secretName,
		setupMode:  !secretExists,
		ClientFactory: func(token string) (kubernetes.Interface, error) {
			return client, nil
		},
	}
	return svc, client
}

func TestSetupStatusHandler(t *testing.T) {
	svc, _ := newSetupService(t, false)

	req := httptest.NewRequest(http.MethodGet, "/api/setup/status", nil)
	rr := httptest.NewRecorder()
	svc.SetupStatusHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var resp SetupStatusResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.SetupRequired {
		t.Fatalf("expected setupRequired true")
	}

	svc, _ = newSetupService(t, true)
	rr = httptest.NewRecorder()
	svc.SetupStatusHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.SetupRequired {
		t.Fatalf("expected setupRequired false when secret exists")
	}
}

func TestSetupStatusHandler_AuthError(t *testing.T) {
	// Mock client that returns Forbidden error
	svc, client := newSetupService(t, false)

	// Inject a reactor to simulate Forbidden error on Secret Get
	// Note: We need to ensure k8sRepo uses this client
	client.PrependReactor("get", "secrets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, apierrors.NewForbidden(corev1.Resource("secrets"), "dkonsole-auth", fmt.Errorf("forbidden"))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/setup/status", nil)
	rr := httptest.NewRecorder()
	svc.SetupStatusHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var resp SetupStatusResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// Verify that we fallback to SetupRequired = true
	if !resp.SetupRequired {
		t.Fatalf("expected setupRequired true on Auth Error")
	}
	// Verify tokenUpdateRequired is FALSE (because we want Full Setup)
	if resp.TokenUpdateRequired {
		t.Fatalf("expected tokenUpdateRequired FALSE on Auth Error")
	}
}

func TestSetupCompleteHandler_Success(t *testing.T) {
	svc, client := newSetupService(t, false)

	body := bytes.NewBufferString(`{"username":"admin","password":"strongpass","jwtSecret":"` + strings.Repeat("b", 32) + `", "serviceAccountToken":"test-token"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/setup/complete", body)
	rr := httptest.NewRecorder()

	svc.SetupCompleteHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rr.Code, rr.Body.String())
	}
	secret, err := client.CoreV1().Secrets("setup-ns").Get(t.Context(), "dkonsole-auth", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("secret not created: %v", err)
	}
	if string(secret.Data["service-account-token"]) != "test-token" {
		t.Fatalf("expected token test-token, got %s", string(secret.Data["service-account-token"]))
	}
}

func TestSetupCompleteHandler_ForbiddenWhenExists(t *testing.T) {
	svc, _ := newSetupService(t, true)

	body := bytes.NewBufferString(`{"username":"admin","password":"strongpass","jwtSecret":"` + strings.Repeat("c", 32) + `", "serviceAccountToken":"test-token"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/setup/complete", body)
	rr := httptest.NewRecorder()

	svc.SetupCompleteHandler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
}

func TestSetupCompleteHandler_AuthError_PreCheck(t *testing.T) {
	// Scenario: Initial check fails with Forbidden, but we provide a valid token.
	// We expect the handler to PROCEED and use the new token to create/update.
	svc, _ := newSetupService(t, false)

	// Simulate Forbidden on STARTUP check (using default client)
	// We use the same reactor trick but we need to ensure it applies to the 'default' client check
	// and NOT the 'new valid token' check.
	// Since both use the same fake client in our test setup (via ClientFactory mock), this is tricky.
	// However, CreateOrUpdateSecret uses a NEW client (created by factory).
	// In our test harness:
	//   repo.client is 'client' (fake).
	//   ClientFactory returns 'client' (same fake).
	// So a reactor on 'client' affects both.

	// Complex reactor: Fail on "Get Secret" IF we don't have the context or some marker?
	// K8s fake client doesn't support context-based differentiation well for Reactor.
	// Alternative: We can modify newSetupService to return a svc where k8sRepo uses a DIFFERENT client instance than what Factory returns.

	// Let's manually construct the service for this test
	defaultClient := k8sfake.NewSimpleClientset()
	setupClient := k8sfake.NewSimpleClientset() // Client simulating the "New Token" access

	// Default client fails on GET
	defaultClient.PrependReactor("get", "secrets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, apierrors.NewForbidden(corev1.Resource("secrets"), "dkonsole-auth", fmt.Errorf("forbidden"))
	})

	// Setup client allows Everything
	// (It's a fresh fake client, so it starts empty. CreateOrUpdate will see empty -> Create, or if we seed it...)

	repo := &K8sUserRepository{
		client:     defaultClient,
		namespace:  "setup-ns",
		secretName: "dkonsole-auth",
		ClientFactory: func(token string) (kubernetes.Interface, error) {
			if token == "valid-token" {
				return setupClient, nil
			}
			return nil, fmt.Errorf("invalid token")
		},
	}

	svc = &Service{
		k8sRepo:       repo,
		k8sClient:     defaultClient,
		secretName:    repo.secretName,
		setupMode:     true,
		ClientFactory: repo.ClientFactory,
	}

	body := bytes.NewBufferString(`{"username":"admin","password":"strongpass","jwtSecret":"` + strings.Repeat("e", 32) + `", "serviceAccountToken":"valid-token"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/setup/complete", body)
	rr := httptest.NewRecorder()

	svc.SetupCompleteHandler(rr, req)

	// Should SUCCEED (200), not Forbidden (403)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rr.Code, rr.Body.String())
	}

	// Verify secret was created in the SETUP client (not default)
	secret, err := setupClient.CoreV1().Secrets("setup-ns").Get(t.Context(), "dkonsole-auth", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("secret not created in setup client: %v", err)
	}
	if string(secret.Data["service-account-token"]) != "valid-token" {
		t.Errorf("token match error")
	}
}

func TestUpdateTokenHandler(t *testing.T) {
	svc, client := newSetupService(t, true)

	body := bytes.NewBufferString(`{"serviceAccountToken":"new-token"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/setup/token", body)
	rr := httptest.NewRecorder()

	svc.UpdateTokenHandler(rr, req)

	// Check results.
	if rr.Code != http.StatusOK {
		t.Fatalf("UpdateTokenHandler failed: %v", rr.Body.String())
	}

	secret, _ := client.CoreV1().Secrets("setup-ns").Get(t.Context(), "dkonsole-auth", metav1.GetOptions{})
	if string(secret.Data["service-account-token"]) != "new-token" {
		t.Errorf("expected token new-token, got %s", string(secret.Data["service-account-token"]))
	}
}

func TestSetupIncompleteSecret(t *testing.T) {
	// Create service with existing secret BUT missing admin user
	svc, client := newSetupService(t, true)

	// Corrupt the secret to remove admin-username
	secret, _ := client.CoreV1().Secrets("setup-ns").Get(t.Context(), "dkonsole-auth", metav1.GetOptions{})
	delete(secret.Data, "admin-username")
	client.CoreV1().Secrets("setup-ns").Update(t.Context(), secret, metav1.UpdateOptions{})

	// 1. Check Status
	req := httptest.NewRequest(http.MethodGet, "/api/setup/status", nil)
	rr := httptest.NewRecorder()
	svc.SetupStatusHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var resp SetupStatusResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	// Crucial check: Should be TRUE even though secret exists
	if !resp.SetupRequired {
		t.Fatalf("expected setupRequired true for incomplete secret")
	}

	// 2. Complete Setup (should succeed and update secret)
	body := bytes.NewBufferString(`{"username":"admin","password":"strongpass","jwtSecret":"` + strings.Repeat("d", 32) + `", "serviceAccountToken":"updated-token"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/setup/complete", body)
	rr = httptest.NewRecorder()

	svc.SetupCompleteHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("setup complete failed status = %d, want 200, body=%s", rr.Code, rr.Body.String())
	}

	// Verify secret is updated
	updatedSecret, _ := client.CoreV1().Secrets("setup-ns").Get(t.Context(), "dkonsole-auth", metav1.GetOptions{})
	if string(updatedSecret.Data["admin-username"]) != "admin" {
		t.Fatal("admin username not updated in secret")
	}
	if string(updatedSecret.Data["service-account-token"]) != "updated-token" {
		t.Fatal("token not updated in secret")
	}
}
