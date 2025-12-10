package k8s

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

// mockNamespaceRepository is a mock implementation of NamespaceRepository
type mockNamespaceRepository struct {
	listFunc func(ctx context.Context) ([]corev1.Namespace, error)
}

func (m *mockNamespaceRepository) List(ctx context.Context) ([]corev1.Namespace, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return []corev1.Namespace{}, nil
}

func TestNamespaceService_GetNamespaces(t *testing.T) {
	now := metav1.Now()

	tests := []struct {
		name          string
		repoFunc      func(ctx context.Context) ([]corev1.Namespace, error)
		ctx           context.Context
		wantErr       bool
		wantCount     int
		expectedNames []string
	}{
		{
			name: "successful list all namespaces (admin)",
			repoFunc: func(ctx context.Context) ([]corev1.Namespace, error) {
				return []corev1.Namespace{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "default",
							Labels:            map[string]string{"env": "prod"},
							CreationTimestamp: now,
						},
						Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "kube-system",
							CreationTimestamp: now,
						},
						Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
					},
				}, nil
			},
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&auth.AuthClaims{
					Claims: models.Claims{
						Username:    "admin",
						Role:        "admin",
						Permissions: nil, // Admin has nil permissions
					},
				},
			),
			wantErr:       false,
			wantCount:     2,
			expectedNames: []string{"default", "kube-system"},
		},
		{
			name: "repository error",
			repoFunc: func(ctx context.Context) ([]corev1.Namespace, error) {
				return nil, errors.New("repository error")
			},
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&auth.AuthClaims{
					Claims: models.Claims{
						Username:    "admin",
						Role:        "admin",
						Permissions: nil,
					},
				},
			),
			wantErr:   true,
			wantCount: 0,
		},
		{
			name: "empty namespaces list",
			repoFunc: func(ctx context.Context) ([]corev1.Namespace, error) {
				return []corev1.Namespace{}, nil
			},
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&auth.AuthClaims{
					Claims: models.Claims{
						Username:    "admin",
						Role:        "admin",
						Permissions: nil,
					},
				},
			),
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "restricted user sees only allowed namespaces",
			repoFunc: func(ctx context.Context) ([]corev1.Namespace, error) {
				return []corev1.Namespace{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "default"},
						Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "secret-ns"},
						Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
					},
				}, nil
			},
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&auth.AuthClaims{
					Claims: models.Claims{
						Username: "user",
						Role:     "user",
						Permissions: map[string]string{
							"default": "view",
						},
					},
				},
			),
			wantErr:       false,
			wantCount:     1,
			expectedNames: []string{"default"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockNamespaceRepository{
				listFunc: tt.repoFunc,
			}

			service := NewNamespaceService(mockRepo)
			result, err := service.GetNamespaces(tt.ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result) != tt.wantCount {
					t.Errorf("GetNamespaces() count = %v, want %v", len(result), tt.wantCount)
				}

				if len(tt.expectedNames) > 0 {
					resultNames := make([]string, len(result))
					for i, ns := range result {
						resultNames[i] = ns.Name
					}

					for _, expectedName := range tt.expectedNames {
						found := false
						for _, resultName := range resultNames {
							if resultName == expectedName {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("GetNamespaces() missing namespace %v in result", expectedName)
						}
					}
				}

				// Verify structure of returned namespaces
				for _, ns := range result {
					if ns.Name == "" {
						t.Errorf("GetNamespaces() namespace name is empty")
					}
					if ns.Status == "" {
						t.Errorf("GetNamespaces() namespace status is empty")
					}
					if ns.Created == "" {
						t.Errorf("GetNamespaces() namespace created timestamp is empty")
					}
				}
			}
		})
	}
}
