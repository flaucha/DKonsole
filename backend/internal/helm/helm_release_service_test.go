package helm

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// mockHelmReleaseRepository is a mock implementation of HelmReleaseRepository
type mockHelmReleaseRepository struct {
	listHelmSecretsFunc           func(ctx context.Context) ([]corev1.Secret, error)
	listHelmConfigMapsFunc        func(ctx context.Context) ([]corev1.ConfigMap, error)
	listSecretsInNamespaceFunc    func(ctx context.Context, namespace string) ([]corev1.Secret, error)
	listConfigMapsInNamespaceFunc func(ctx context.Context, namespace string) ([]corev1.ConfigMap, error)
	deleteSecretFunc              func(ctx context.Context, namespace, name string) error
	deleteConfigMapFunc           func(ctx context.Context, namespace, name string) error
}

func buildReleaseDataJSON(status string, revision int, chartName, chartVersion, appVersion, description string) []byte {
	releaseInfo := map[string]interface{}{
		"chart": map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":       chartName,
				"version":    chartVersion,
				"appVersion": appVersion,
			},
		},
		"info": map[string]interface{}{
			"status":      status,
			"revision":    float64(revision),
			"description": description,
			"last_deployed": map[string]interface{}{
				"Time": "2025-01-01T00:00:00Z",
			},
		},
	}
	data, _ := json.Marshal(releaseInfo)
	return data
}

func buildHelmSecretWithRelease(name, namespace, status string, revision int, chart, version, appVersion, description string, encode bool) corev1.Secret {
	data := buildReleaseDataJSON(status, revision, chart, version, appVersion, description)
	if encode {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, _ = gz.Write(data)
		gz.Close()
		data = []byte(base64.StdEncoding.EncodeToString(buf.Bytes()))
	}

	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("sh.helm.release.v1.%s.v%d", name, revision),
			Namespace: namespace,
			Labels: map[string]string{
				"name":    name,
				"version": fmt.Sprintf("%d", revision),
				"status":  "superseded",
			},
			Annotations: map[string]string{
				"meta.helm.sh/release-name":      name,
				"meta.helm.sh/release-namespace": namespace,
			},
		},
		Data: map[string][]byte{
			"release": data,
		},
	}
}

func (m *mockHelmReleaseRepository) ListHelmSecrets(ctx context.Context) ([]corev1.Secret, error) {
	if m.listHelmSecretsFunc != nil {
		return m.listHelmSecretsFunc(ctx)
	}
	return []corev1.Secret{}, nil
}

func (m *mockHelmReleaseRepository) ListHelmConfigMaps(ctx context.Context) ([]corev1.ConfigMap, error) {
	if m.listHelmConfigMapsFunc != nil {
		return m.listHelmConfigMapsFunc(ctx)
	}
	return []corev1.ConfigMap{}, nil
}

func (m *mockHelmReleaseRepository) ListSecretsInNamespace(ctx context.Context, namespace string) ([]corev1.Secret, error) {
	if m.listSecretsInNamespaceFunc != nil {
		return m.listSecretsInNamespaceFunc(ctx, namespace)
	}
	return []corev1.Secret{}, nil
}

func (m *mockHelmReleaseRepository) ListConfigMapsInNamespace(ctx context.Context, namespace string) ([]corev1.ConfigMap, error) {
	if m.listConfigMapsInNamespaceFunc != nil {
		return m.listConfigMapsInNamespaceFunc(ctx, namespace)
	}
	return []corev1.ConfigMap{}, nil
}

func (m *mockHelmReleaseRepository) DeleteSecret(ctx context.Context, namespace, name string) error {
	if m.deleteSecretFunc != nil {
		return m.deleteSecretFunc(ctx, namespace, name)
	}
	return nil
}

func (m *mockHelmReleaseRepository) DeleteConfigMap(ctx context.Context, namespace, name string) error {
	if m.deleteConfigMapFunc != nil {
		return m.deleteConfigMapFunc(ctx, namespace, name)
	}
	return nil
}

func TestHelmReleaseService_GetHelmReleases(t *testing.T) {
	tests := []struct {
		name                   string
		listHelmSecretsFunc    func(ctx context.Context) ([]corev1.Secret, error)
		listHelmConfigMapsFunc func(ctx context.Context) ([]corev1.ConfigMap, error)
		wantErr                bool
		expectedReleaseCount   int
	}{
		{
			name: "successful list with secrets",
			listHelmSecretsFunc: func(ctx context.Context) ([]corev1.Secret, error) {
				return []corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "sh.helm.release.v1.my-app.v1",
							Namespace: "default",
							Labels: map[string]string{
								"owner":   "helm",
								"name":    "my-app",
								"status":  "deployed",
								"version": "1",
							},
							Annotations: map[string]string{
								"meta.helm.sh/release-name":      "my-app",
								"meta.helm.sh/release-namespace": "default",
							},
						},
					},
				}, nil
			},
			listHelmConfigMapsFunc: func(ctx context.Context) ([]corev1.ConfigMap, error) {
				return []corev1.ConfigMap{}, nil
			},
			wantErr:              false,
			expectedReleaseCount: 0, // Will be 0 because we need release data to parse
		},
		{
			name: "empty releases list",
			listHelmSecretsFunc: func(ctx context.Context) ([]corev1.Secret, error) {
				return []corev1.Secret{}, nil
			},
			listHelmConfigMapsFunc: func(ctx context.Context) ([]corev1.ConfigMap, error) {
				return []corev1.ConfigMap{}, nil
			},
			wantErr:              false,
			expectedReleaseCount: 0,
		},
		{
			name: "error listing secrets is ignored",
			listHelmSecretsFunc: func(ctx context.Context) ([]corev1.Secret, error) {
				return nil, errors.New("secrets error")
			},
			listHelmConfigMapsFunc: func(ctx context.Context) ([]corev1.ConfigMap, error) {
				return []corev1.ConfigMap{}, nil
			},
			wantErr:              false,
			expectedReleaseCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockHelmReleaseRepository{
				listHelmSecretsFunc:    tt.listHelmSecretsFunc,
				listHelmConfigMapsFunc: tt.listHelmConfigMapsFunc,
			}

			service := NewHelmReleaseService(mockRepo)
			ctx := context.Background()

			releases, err := service.GetHelmReleases(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetHelmReleases() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if releases == nil {
					t.Errorf("GetHelmReleases() releases is nil")
					return
				}
				if len(releases) != tt.expectedReleaseCount {
					t.Errorf("GetHelmReleases() count = %v, want %v", len(releases), tt.expectedReleaseCount)
				}
			}
		})
	}
}

func TestHelmReleaseService_DeleteHelmRelease(t *testing.T) {
	tests := []struct {
		name                          string
		request                       DeleteHelmReleaseRequest
		listSecretsInNamespaceFunc    func(ctx context.Context, namespace string) ([]corev1.Secret, error)
		listConfigMapsInNamespaceFunc func(ctx context.Context, namespace string) ([]corev1.ConfigMap, error)
		deleteSecretFunc              func(ctx context.Context, namespace, name string) error
		deleteConfigMapFunc           func(ctx context.Context, namespace, name string) error
		wantErr                       bool
		errMsg                        string
		expectedDeletedCount          int
	}{
		{
			name: "successful delete with secrets and configmaps",
			request: DeleteHelmReleaseRequest{
				Name:      "my-app",
				Namespace: "default",
			},
			listSecretsInNamespaceFunc: func(ctx context.Context, namespace string) ([]corev1.Secret, error) {
				return []corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "sh.helm.release.v1.my-app.v1",
							Namespace: namespace,
							Annotations: map[string]string{
								"meta.helm.sh/release-name":      "my-app",
								"meta.helm.sh/release-namespace": namespace,
							},
						},
					},
				}, nil
			},
			listConfigMapsInNamespaceFunc: func(ctx context.Context, namespace string) ([]corev1.ConfigMap, error) {
				return []corev1.ConfigMap{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-app.v1",
							Namespace: namespace,
							Annotations: map[string]string{
								"meta.helm.sh/release-name":      "my-app",
								"meta.helm.sh/release-namespace": namespace,
							},
						},
					},
				}, nil
			},
			deleteSecretFunc: func(ctx context.Context, namespace, name string) error {
				return nil
			},
			deleteConfigMapFunc: func(ctx context.Context, namespace, name string) error {
				return nil
			},
			wantErr:              false,
			expectedDeletedCount: 2, // 1 secret + 1 configmap
		},
		{
			name: "no release found",
			request: DeleteHelmReleaseRequest{
				Name:      "non-existent",
				Namespace: "default",
			},
			listSecretsInNamespaceFunc: func(ctx context.Context, namespace string) ([]corev1.Secret, error) {
				return []corev1.Secret{}, nil
			},
			listConfigMapsInNamespaceFunc: func(ctx context.Context, namespace string) ([]corev1.ConfigMap, error) {
				return []corev1.ConfigMap{}, nil
			},
			wantErr: true,
			errMsg:  "no Helm release secrets found",
		},
		{
			name: "error listing secrets is ignored",
			request: DeleteHelmReleaseRequest{
				Name:      "my-app",
				Namespace: "default",
			},
			listSecretsInNamespaceFunc: func(ctx context.Context, namespace string) ([]corev1.Secret, error) {
				return nil, errors.New("list error")
			},
			listConfigMapsInNamespaceFunc: func(ctx context.Context, namespace string) ([]corev1.ConfigMap, error) {
				return []corev1.ConfigMap{}, nil
			},
			wantErr: true,
			errMsg:  "no Helm release secrets found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockHelmReleaseRepository{
				listSecretsInNamespaceFunc:    tt.listSecretsInNamespaceFunc,
				listConfigMapsInNamespaceFunc: tt.listConfigMapsInNamespaceFunc,
				deleteSecretFunc:              tt.deleteSecretFunc,
				deleteConfigMapFunc:           tt.deleteConfigMapFunc,
			}

			service := NewHelmReleaseService(mockRepo)
			ctx := context.Background()

			result, err := service.DeleteHelmRelease(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteHelmRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("DeleteHelmRelease() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("DeleteHelmRelease() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			if result == nil {
				t.Errorf("DeleteHelmRelease() result is nil")
				return
			}

			if result.SecretsDeleted != tt.expectedDeletedCount {
				t.Errorf("DeleteHelmRelease() SecretsDeleted = %v, want %v", result.SecretsDeleted, tt.expectedDeletedCount)
			}
		})
	}
}

func TestHelmReleaseService_IsSecretRelatedToRelease(t *testing.T) {
	service := NewHelmReleaseService(nil)

	tests := []struct {
		name        string
		secret      corev1.Secret
		releaseName string
		namespace   string
		want        bool
	}{
		{
			name: "secret with matching annotations",
			secret: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.my-app.v1",
					Namespace: "default",
					Annotations: map[string]string{
						"meta.helm.sh/release-name":      "my-app",
						"meta.helm.sh/release-namespace": "default",
					},
				},
			},
			releaseName: "my-app",
			namespace:   "default",
			want:        true,
		},
		{
			name: "secret with matching name prefix",
			secret: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.my-app.v1",
					Namespace: "default",
				},
			},
			releaseName: "my-app",
			namespace:   "default",
			want:        true,
		},
		{
			name: "secret not related to release",
			secret: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "other-secret",
					Namespace: "default",
				},
			},
			releaseName: "my-app",
			namespace:   "default",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isSecretRelatedToRelease(tt.secret, tt.releaseName, tt.namespace)
			if result != tt.want {
				t.Errorf("isSecretRelatedToRelease() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestHelmReleaseService_ParseReleaseFromSecretAndConfigMap(t *testing.T) {
	service := NewHelmReleaseService(nil)

	secret := buildHelmSecretWithRelease("demo", "ns1", "superseded", 2, "chart-demo", "1.2.3", "2.0.0", "first deploy", true)
	result := service.parseReleaseFromSecret(secret)
	if result == nil {
		t.Fatalf("parseReleaseFromSecret returned nil")
	}
	if result.Name != "demo" || result.Namespace != "ns1" {
		t.Fatalf("unexpected name/namespace: %+v", result)
	}
	if result.Chart != "chart-demo" || result.Version != "1.2.3" || result.AppVersion != "2.0.0" {
		t.Fatalf("unexpected chart fields: %+v", result)
	}
	if result.Description != "first deploy" || result.Updated == "" {
		t.Fatalf("description/updated not set: %+v", result)
	}
	if result.Revision != 2 || result.Status != "superseded" {
		t.Fatalf("revision/status not taken from labels/info: %+v", result)
	}

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo.v3",
			Namespace: "ns1",
			Labels: map[string]string{
				"name":    "demo",
				"version": "3",
				"status":  "deployed",
			},
			Annotations: map[string]string{
				"meta.helm.sh/release-name":      "demo",
				"meta.helm.sh/release-namespace": "ns1",
			},
		},
	}
	cmResult := service.parseReleaseFromConfigMap(cm)
	if cmResult == nil {
		t.Fatalf("parseReleaseFromConfigMap returned nil")
	}
	if cmResult.Revision != 3 || cmResult.Status != "deployed" {
		t.Fatalf("expected revision 3 deployed from labels, got %+v", cmResult)
	}
}

func TestHelmReleaseService_GetHelmReleasesPrefersNewestDeployed(t *testing.T) {
	secretOld := buildHelmSecretWithRelease("app", "ns1", "deployed", 1, "old", "0.1.0", "1.0", "old deploy", true)
	secretNewSuperseded := buildHelmSecretWithRelease("app", "ns1", "superseded", 3, "new", "0.2.0", "1.1", "new deploy", true)
	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app.v2",
			Namespace: "ns1",
			Labels: map[string]string{
				"name":    "app",
				"version": "2",
				"status":  "failed",
			},
			Annotations: map[string]string{
				"meta.helm.sh/release-name":      "app",
				"meta.helm.sh/release-namespace": "ns1",
			},
		},
	}

	service := NewHelmReleaseService(&mockHelmReleaseRepository{
		listHelmSecretsFunc: func(ctx context.Context) ([]corev1.Secret, error) {
			return []corev1.Secret{secretOld, secretNewSuperseded}, nil
		},
		listHelmConfigMapsFunc: func(ctx context.Context) ([]corev1.ConfigMap, error) {
			return []corev1.ConfigMap{cm}, nil
		},
	})

	releases, err := service.GetHelmReleases(context.Background())
	if err != nil {
		t.Fatalf("GetHelmReleases returned error: %v", err)
	}
	if len(releases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(releases))
	}
	if releases[0].Revision != 3 {
		t.Fatalf("expected highest revision 3, got %d", releases[0].Revision)
	}
	if releases[0].Chart != "new" {
		t.Fatalf("expected chart new, got %s", releases[0].Chart)
	}
}
