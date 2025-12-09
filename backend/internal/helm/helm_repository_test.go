package helm

import (
	"context"
	"fmt"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8s_testing "k8s.io/client-go/testing"
)

func TestK8sHelmReleaseRepository(t *testing.T) {
	ctx := context.TODO()

	// Setup fake data
	helmSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sh.helm.release.v1.release-1.v1",
			Namespace: "default",
			Labels:    map[string]string{"owner": "helm"},
		},
	}
	otherSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other-secret",
			Namespace: "default",
		},
	}
	helmCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sh.helm.release.v1.release-1.v1",
			Namespace: "default",
			Labels:    map[string]string{"owner": "helm"},
		},
	}

	client := fake.NewSimpleClientset(helmSecret, otherSecret, helmCM)
	repo := NewK8sHelmReleaseRepository(client)

	// Test ListHelmSecrets
	t.Run("ListHelmSecrets", func(t *testing.T) {
		secrets, err := repo.ListHelmSecrets(ctx)
		if err != nil {
			t.Fatalf("ListHelmSecrets failed: %v", err)
		}
		if len(secrets) != 1 {
			t.Errorf("expected 1 helm secret, got %d", len(secrets))
		}
		if secrets[0].Name != helmSecret.Name {
			t.Errorf("expected secret %s, got %s", helmSecret.Name, secrets[0].Name)
		}
	})

	// Test ListHelmConfigMaps
	t.Run("ListHelmConfigMaps", func(t *testing.T) {
		cms, err := repo.ListHelmConfigMaps(ctx)
		if err != nil {
			t.Fatalf("ListHelmConfigMaps failed: %v", err)
		}
		if len(cms) != 1 {
			t.Errorf("expected 1 helm configmap, got %d", len(cms))
		}
	})

	// Test ListSecretsInNamespace
	t.Run("ListSecretsInNamespace", func(t *testing.T) {
		secrets, err := repo.ListSecretsInNamespace(ctx, "default")
		if err != nil {
			t.Fatalf("ListSecretsInNamespace failed: %v", err)
		}
		if len(secrets) != 2 {
			t.Errorf("expected 2 secrets in default ns, got %d", len(secrets))
		}
	})

	// Test ListConfigMapsInNamespace
	t.Run("ListConfigMapsInNamespace", func(t *testing.T) {
		cms, err := repo.ListConfigMapsInNamespace(ctx, "default")
		if err != nil {
			t.Fatalf("ListConfigMapsInNamespace failed: %v", err)
		}
		if len(cms) != 1 {
			t.Errorf("expected 1 configmap in default ns, got %d", len(cms))
		}
	})

	// Test DeleteSecret
	t.Run("DeleteSecret", func(t *testing.T) {
		err := repo.DeleteSecret(ctx, "default", "other-secret")
		if err != nil {
			t.Fatalf("DeleteSecret failed: %v", err)
		}
		// Verify deletion
		_, err = client.CoreV1().Secrets("default").Get(ctx, "other-secret", metav1.GetOptions{})
		if err == nil {
			t.Error("Secret should be deleted")
		}
	})

	// Test DeleteConfigMap
	t.Run("DeleteConfigMap", func(t *testing.T) {
		err := repo.DeleteConfigMap(ctx, "default", helmCM.Name)
		if err != nil {
			t.Fatalf("DeleteConfigMap failed: %v", err)
		}
		// Verify deletion
		_, err = client.CoreV1().ConfigMaps("default").Get(ctx, helmCM.Name, metav1.GetOptions{})
		if err == nil {
			t.Error("ConfigMap should be deleted")
		}
	})

	// Test error scenarios using Reactors
	t.Run("ListHelmSecretsError", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		client.PrependReactor("list", "secrets", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, fmt.Errorf("k8s error")
		})
		repo := NewK8sHelmReleaseRepository(client)
		_, err := repo.ListHelmSecrets(ctx)
		if err == nil {
			t.Error("expected error listing secrets")
		}
	})

	t.Run("ListHelmConfigMapsError", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		client.PrependReactor("list", "configmaps", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, fmt.Errorf("k8s error")
		})
		repo := NewK8sHelmReleaseRepository(client)
		_, err := repo.ListHelmConfigMaps(ctx)
		if err == nil {
			t.Error("expected error listing configmaps")
		}
	})

	t.Run("ListSecretsInNamespaceError", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		client.PrependReactor("list", "secrets", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, fmt.Errorf("k8s error")
		})
		repo := NewK8sHelmReleaseRepository(client)
		_, err := repo.ListSecretsInNamespace(ctx, "default")
		if err == nil {
			t.Error("expected error listing secrets in namespace")
		}
	})

	t.Run("ListConfigMapsInNamespaceError", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		client.PrependReactor("list", "configmaps", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, fmt.Errorf("k8s error")
		})
		repo := NewK8sHelmReleaseRepository(client)
		_, err := repo.ListConfigMapsInNamespace(ctx, "default")
		if err == nil {
			t.Error("expected error listing configmaps in namespace")
		}
	})

	t.Run("DeleteConfigMapError", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		client.PrependReactor("delete", "configmaps", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, fmt.Errorf("k8s error")
		})
		repo := NewK8sHelmReleaseRepository(client)
		err := repo.DeleteConfigMap(ctx, "default", "foo")
		if err == nil {
			t.Error("expected error deleting configmap")
		}
	})
}

func TestK8sHelmJobRepository(t *testing.T) {
	ctx := context.TODO()
	client := fake.NewSimpleClientset()
	repo := NewK8sHelmJobRepository(client)

	// Test CreateConfigMap
	t.Run("CreateConfigMap", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: "default",
			},
		}
		err := repo.CreateConfigMap(ctx, "default", cm)
		if err != nil {
			t.Fatalf("CreateConfigMap failed: %v", err)
		}
		// Verify
		_, err = client.CoreV1().ConfigMaps("default").Get(ctx, "test-cm", metav1.GetOptions{})
		if err != nil {
			t.Error("ConfigMap was not created")
		}
	})

	// Test CreateJob
	t.Run("CreateJob", func(t *testing.T) {
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job",
				Namespace: "default",
			},
		}
		err := repo.CreateJob(ctx, "default", job)
		if err != nil {
			t.Fatalf("CreateJob failed: %v", err)
		}
		// Verify
		_, err = client.BatchV1().Jobs("default").Get(ctx, "test-job", metav1.GetOptions{})
		if err != nil {
			t.Error("Job was not created")
		}
	})

	// Test GetServiceAccount
	t.Run("GetServiceAccount", func(t *testing.T) {
		// Create SA first
		sa := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sa",
				Namespace: "default",
			},
		}
		_, err := client.CoreV1().ServiceAccounts("default").Create(ctx, sa, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create prerequisite SA: %v", err)
		}

		// Test Get
		got, err := repo.GetServiceAccount(ctx, "default", "test-sa")
		if err != nil {
			t.Fatalf("GetServiceAccount failed: %v", err)
		}
		if got.Name != "test-sa" {
			t.Errorf("expected sa name test-sa, got %s", got.Name)
		}
	})

	// Test Error Scenarios
	t.Run("GetMissingSA", func(t *testing.T) {
		_, err := repo.GetServiceAccount(ctx, "default", "missing-sa")
		if err == nil {
			t.Error("GetServiceAccount should fail for missing SA")
		}
	})

	t.Run("CreateConfigMapError", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		client.PrependReactor("create", "configmaps", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, fmt.Errorf("create error")
		})
		repo := NewK8sHelmJobRepository(client)
		err := repo.CreateConfigMap(ctx, "default", &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "foo"}})
		if err == nil {
			t.Error("expected error creating configmap")
		}
	})

	t.Run("CreateJobError", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		client.PrependReactor("create", "jobs", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, fmt.Errorf("create error")
		})
		repo := NewK8sHelmJobRepository(client)
		err := repo.CreateJob(ctx, "default", &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "foo"}})
		if err == nil {
			t.Error("expected error creating job")
		}
	})

	t.Run("DeleteSecretError", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		client.PrependReactor("delete", "secrets", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, fmt.Errorf("k8s error")
		})
		repo := NewK8sHelmReleaseRepository(client)
		err := repo.DeleteSecret(ctx, "default", "foo")
		if err == nil {
			t.Error("expected error deleting secret")
		}
	})
}
