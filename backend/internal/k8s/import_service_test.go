package k8s

import (
	"context"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func TestImportService_ImportResources(t *testing.T) {
	missingKindYAML := `
apiVersion: v1
metadata:
  name: test
`

	systemNamespaceYAML := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: kube-system
data:
  key: value
`

	tests := []struct {
		name           string
		yamlContent    string
		ctx            context.Context
		resolveFunc    func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error)
		wantErr        bool
		errMsg         string
		expectedCount  int
		expectedStatus string
	}{
		{
			name:        "empty YAML",
			yamlContent: "",
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
			wantErr: true,
			errMsg:  "no resources found",
		},
		{
			name:        "missing kind",
			yamlContent: missingKindYAML,
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
			wantErr: true,
			errMsg:  "resource kind missing",
		},
		{
			name:        "system namespace blocked",
			yamlContent: systemNamespaceYAML,
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
			resolveFunc: func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
				return schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "configmaps",
				}, models.ResourceMeta{Namespaced: true}, nil
			},
			wantErr: true,
			errMsg:  "cannot create resources in system namespace",
		},
	}

		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockResourceRepository{}

			mockResolver := &mockGVRResolver{
				resolveFunc: tt.resolveFunc,
			}

			// Default resolver if not provided
			if tt.resolveFunc == nil {
				mockResolver.resolveFunc = func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
					return schema.GroupVersionResource{
						Group:    "",
						Version:  "v1",
						Resource: "configmaps",
					}, models.ResourceMeta{Namespaced: true}, nil
				}
			}

			service := NewImportService(mockRepo, mockResolver, nil)

			req := ImportResourceRequest{
				YAMLContent: []byte(tt.yamlContent),
			}

			result, err := service.ImportResources(tt.ctx, req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ImportResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("ImportResources() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ImportResources() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			if result == nil {
				t.Errorf("ImportResources() result is nil")
				return
			}

			if tt.expectedStatus != "" && result.Status != tt.expectedStatus {
				t.Errorf("ImportResources() Status = %v, want %v", result.Status, tt.expectedStatus)
			}

			if tt.expectedCount > 0 && result.Count != tt.expectedCount {
				t.Errorf("ImportResources() Count = %v, want %v", result.Count, tt.expectedCount)
			}
		})
	}
}
