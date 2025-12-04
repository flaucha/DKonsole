package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func newSetupService(t *testing.T, secretExists bool) (*Service, *k8sfake.Clientset) {
	t.Helper()
	client := k8sfake.NewSimpleClientset()
	repo := &K8sUserRepository{
		client:     client,
		namespace:  "setup-ns",
		secretName: "dkonsole-auth",
	}

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

func TestSetupCompleteHandler_Success(t *testing.T) {
	svc, client := newSetupService(t, false)

	body := bytes.NewBufferString(`{"username":"admin","password":"strongpass","jwtSecret":"` + strings.Repeat("b", 32) + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/setup/complete", body)
	rr := httptest.NewRecorder()

	svc.SetupCompleteHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rr.Code, rr.Body.String())
	}
	if _, err := client.CoreV1().Secrets("setup-ns").Get(t.Context(), "dkonsole-auth", metav1.GetOptions{}); err != nil {
		t.Fatalf("secret not created: %v", err)
	}
}

func TestSetupCompleteHandler_ForbiddenWhenExists(t *testing.T) {
	svc, _ := newSetupService(t, true)

	body := bytes.NewBufferString(`{"username":"admin","password":"strongpass","jwtSecret":"` + strings.Repeat("c", 32) + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/setup/complete", body)
	rr := httptest.NewRecorder()

	svc.SetupCompleteHandler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
}
