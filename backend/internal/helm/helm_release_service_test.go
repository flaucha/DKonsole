package helm

import (
	"context"
	"errors"
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
