package settings

import (
	"context"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/prometheus"
)

func TestNewRepositoryDefaultsAndEnvNamespace(t *testing.T) {
	defer os.Unsetenv("POD_NAMESPACE")
	os.Setenv("POD_NAMESPACE", "test-ns")

	repo := NewRepository(nil, "")
	if repo.namespace == "" {
		t.Fatalf("expected namespace to be set")
	}
	if repo.configMapName != "dkonsole-auth" {
		t.Fatalf("expected default configmap name, got %s", repo.configMapName)
	}
}

func TestRepository_GetPrometheusURL(t *testing.T) {
	client := k8sfake.NewSimpleClientset(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cfg",
			Namespace: "default",
		},
		Data: map[string]string{
			"prometheus-url": "http://cm",
		},
	})
	repo := &K8sRepository{client: client, namespace: "default", configMapName: "cfg"}

	url, err := repo.GetPrometheusURL(context.Background())
	if err != nil {
		t.Fatalf("GetPrometheusURL error: %v", err)
	}
	if url != "http://cm" {
		t.Fatalf("expected url from configmap, got %s", url)
	}

	// fall back to env when configmap missing
	_ = client.CoreV1().ConfigMaps("default").Delete(context.Background(), "cfg", metav1.DeleteOptions{})
	defer os.Unsetenv("PROMETHEUS_URL")
	os.Setenv("PROMETHEUS_URL", "http://env")
	url, err = repo.GetPrometheusURL(context.Background())
	if err != nil {
		t.Fatalf("GetPrometheusURL error: %v", err)
	}
	if url != "http://env" {
		t.Fatalf("expected env url, got %s", url)
	}
}

func TestRepository_UpdatePrometheusURL_CreateAndUpdate(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	repo := &K8sRepository{client: client, namespace: "default", configMapName: "cfg"}

	if err := repo.UpdatePrometheusURL(context.Background(), "http://first"); err != nil {
		t.Fatalf("UpdatePrometheusURL create error: %v", err)
	}
	cm, _ := client.CoreV1().ConfigMaps("default").Get(context.Background(), "cfg", metav1.GetOptions{})
	if cm.Data["prometheus-url"] != "http://first" {
		t.Fatalf("expected stored url, got %s", cm.Data["prometheus-url"])
	}

	if err := repo.UpdatePrometheusURL(context.Background(), "http://second"); err != nil {
		t.Fatalf("UpdatePrometheusURL update error: %v", err)
	}
	cm, _ = client.CoreV1().ConfigMaps("default").Get(context.Background(), "cfg", metav1.GetOptions{})
	if cm.Data["prometheus-url"] != "http://second" {
		t.Fatalf("expected updated url, got %s", cm.Data["prometheus-url"])
	}
}

func TestRepository_UpdatePrometheusURL_NoClient(t *testing.T) {
	repo := &K8sRepository{}
	if err := repo.UpdatePrometheusURL(context.Background(), "http://url"); err == nil {
		t.Fatalf("expected error when client is nil")
	}
}

func TestServiceFactory_NewService(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	handlers := &models.Handlers{}
	promHandler := &prometheus.HTTPHandler{}

	factory := NewServiceFactory(client, handlers, "secret", promHandler)
	if svc := factory.NewService(); svc == nil {
		t.Fatalf("expected service instance, got nil")
	}
}
