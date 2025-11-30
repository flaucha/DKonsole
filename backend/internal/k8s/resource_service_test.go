package k8s

import (
	"context"
	"errors"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

// mockResourceRepository is a mock implementation of ResourceRepository
type mockResourceRepository struct {
	getFunc    func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error)
	createFunc func(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, obj *unstructured.Unstructured, options metav1.CreateOptions) (*unstructured.Unstructured, error)
	patchFunc  func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, patchData []byte, patchType types.PatchType, options metav1.PatchOptions) (*unstructured.Unstructured, error)
	deleteFunc func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, options metav1.DeleteOptions) error
}

func (m *mockResourceRepository) Get(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, gvr, name, namespace, namespaced)
	}
	return nil, errors.New("get not implemented")
}

func (m *mockResourceRepository) Create(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, obj *unstructured.Unstructured, options metav1.CreateOptions) (*unstructured.Unstructured, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, gvr, namespace, namespaced, obj, options)
	}
	return nil, errors.New("create not implemented")
}

func (m *mockResourceRepository) Patch(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, patchData []byte, patchType types.PatchType, options metav1.PatchOptions) (*unstructured.Unstructured, error) {
	if m.patchFunc != nil {
		return m.patchFunc(ctx, gvr, name, namespace, namespaced, patchData, patchType, options)
	}
	return nil, errors.New("patch not implemented")
}

func (m *mockResourceRepository) Delete(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, options metav1.DeleteOptions) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, gvr, name, namespace, namespaced, options)
	}
	return errors.New("delete not implemented")
}

// mockGVRResolver is a mock implementation of GVRResolver
type mockGVRResolver struct {
	resolveFunc func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error)
}

func (m *mockGVRResolver) ResolveGVR(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
	if m.resolveFunc != nil {
		return m.resolveFunc(ctx, kind, apiVersion, namespacedParam)
	}
	return schema.GroupVersionResource{}, models.ResourceMeta{}, errors.New("resolve not implemented")
}

func TestResourceService_DeleteResource(t *testing.T) {
	tests := []struct {
		name       string
		req        DeleteResourceRequest
		setupMocks func() (*mockResourceRepository, *mockGVRResolver)
		wantErr    bool
		errMsg     string
	}{
		{
			name: "successful deletion",
			req: DeleteResourceRequest{
				Kind:      "Deployment",
				Name:      "test-deployment",
				Namespace: "default",
				Force:     false,
			},
			setupMocks: func() (*mockResourceRepository, *mockGVRResolver) {
				repo := &mockResourceRepository{
					deleteFunc: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, options metav1.DeleteOptions) error {
						if name != "test-deployment" || namespace != "default" {
							return errors.New("unexpected parameters")
						}
						return nil
					},
				}
				resolver := &mockGVRResolver{
					resolveFunc: func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
						return schema.GroupVersionResource{
							Group:    "apps",
							Version:  "v1",
							Resource: "deployments",
						}, models.ResourceMeta{Namespaced: true}, nil
					},
				}
				return repo, resolver
			},
			wantErr: false,
		},
		{
			name: "missing namespace for namespaced resource",
			req: DeleteResourceRequest{
				Kind:      "Deployment",
				Name:      "test-deployment",
				Namespace: "",
				Force:     false,
			},
			setupMocks: func() (*mockResourceRepository, *mockGVRResolver) {
				repo := &mockResourceRepository{}
				resolver := &mockGVRResolver{
					resolveFunc: func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
						return schema.GroupVersionResource{}, models.ResourceMeta{Namespaced: true}, nil
					},
				}
				return repo, resolver
			},
			wantErr: true,
			errMsg:  "namespace is required",
		},
		{
			name: "force deletion",
			req: DeleteResourceRequest{
				Kind:      "Pod",
				Name:      "test-pod",
				Namespace: "default",
				Force:     true,
			},
			setupMocks: func() (*mockResourceRepository, *mockGVRResolver) {
				repo := &mockResourceRepository{
					deleteFunc: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, options metav1.DeleteOptions) error {
						if options.GracePeriodSeconds == nil || *options.GracePeriodSeconds != 0 {
							return errors.New("expected grace period to be 0 for force delete")
						}
						return nil
					},
				}
				resolver := &mockGVRResolver{
					resolveFunc: func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
						return schema.GroupVersionResource{
							Group:    "",
							Version:  "v1",
							Resource: "pods",
						}, models.ResourceMeta{Namespaced: true}, nil
					},
				}
				return repo, resolver
			},
			wantErr: false,
		},
		{
			name: "GVR resolution error",
			req: DeleteResourceRequest{
				Kind:      "Unknown",
				Name:      "test",
				Namespace: "default",
			},
			setupMocks: func() (*mockResourceRepository, *mockGVRResolver) {
				repo := &mockResourceRepository{}
				resolver := &mockGVRResolver{
					resolveFunc: func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
						return schema.GroupVersionResource{}, models.ResourceMeta{}, errors.New("failed to resolve GVR")
					},
				}
				return repo, resolver
			},
			wantErr: true,
			errMsg:  "failed to resolve GVR",
		},
		{
			name: "delete repository error",
			req: DeleteResourceRequest{
				Kind:      "Deployment",
				Name:      "test-deployment",
				Namespace: "default",
			},
			setupMocks: func() (*mockResourceRepository, *mockGVRResolver) {
				repo := &mockResourceRepository{
					deleteFunc: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, options metav1.DeleteOptions) error {
						return errors.New("resource not found")
					},
				}
				resolver := &mockGVRResolver{
					resolveFunc: func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
						return schema.GroupVersionResource{
							Group:    "apps",
							Version:  "v1",
							Resource: "deployments",
						}, models.ResourceMeta{Namespaced: true}, nil
					},
				}
				return repo, resolver
			},
			wantErr: true,
			errMsg:  "failed to delete resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, resolver := tt.setupMocks()
			service := NewResourceService(repo, resolver)

			// Create context with user for permission checks
			ctx := context.WithValue(context.Background(), auth.UserContextKey(), &auth.AuthClaims{
				Claims: models.Claims{
					Username: "admin",
					Role:     "admin",
				},
			})

			err := service.DeleteResource(ctx, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("DeleteResource() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestResourceService_GetResourceYAML(t *testing.T) {
	tests := []struct {
		name       string
		req        GetResourceRequest
		setupMocks func() (*mockResourceRepository, *mockGVRResolver)
		wantErr    bool
		checkYAML  bool
	}{
		{
			name: "successful get",
			req: GetResourceRequest{
				Kind:      "Deployment",
				Name:      "test-deployment",
				Namespace: "default",
			},
			setupMocks: func() (*mockResourceRepository, *mockGVRResolver) {
				repo := &mockResourceRepository{
					getFunc: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error) {
						obj := &unstructured.Unstructured{}
						obj.SetName("test-deployment")
						obj.SetNamespace("default")
						obj.SetKind("Deployment")
						obj.Object["apiVersion"] = "apps/v1"
						obj.Object["metadata"] = map[string]interface{}{
							"name":      "test-deployment",
							"namespace": "default",
						}
						return obj, nil
					},
				}
				resolver := &mockGVRResolver{
					resolveFunc: func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
						return schema.GroupVersionResource{
							Group:    "apps",
							Version:  "v1",
							Resource: "deployments",
						}, models.ResourceMeta{Namespaced: true}, nil
					},
				}
				return repo, resolver
			},
			wantErr:   false,
			checkYAML: true,
		},
		{
			name: "resource not found",
			req: GetResourceRequest{
				Kind:      "Deployment",
				Name:      "not-found",
				Namespace: "default",
			},
			setupMocks: func() (*mockResourceRepository, *mockGVRResolver) {
				repo := &mockResourceRepository{
					getFunc: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error) {
						return nil, errors.New("not found")
					},
				}
				resolver := &mockGVRResolver{
					resolveFunc: func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
						return schema.GroupVersionResource{
							Group:    "apps",
							Version:  "v1",
							Resource: "deployments",
						}, models.ResourceMeta{Namespaced: true}, nil
					},
				}
				return repo, resolver
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, resolver := tt.setupMocks()
			service := NewResourceService(repo, resolver)

			yaml, err := service.GetResourceYAML(context.Background(), tt.req, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetResourceYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkYAML {
				if yaml == "" {
					t.Errorf("GetResourceYAML() yaml is empty")
				}
				if !contains(yaml, "test-deployment") {
					t.Errorf("GetResourceYAML() yaml does not contain resource name")
				}
			}
		})
	}
}

func TestResourceService_UpdateResource(t *testing.T) {
	validYAML := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: default
spec:
  replicas: 3
`

	tests := []struct {
		name       string
		req        UpdateResourceRequest
		setupMocks func() (*mockResourceRepository, *mockGVRResolver)
		wantErr    bool
		errMsg     string
	}{
		{
			name: "successful update",
			req: UpdateResourceRequest{
				YAMLContent: validYAML,
				Kind:        "Deployment",
				Name:        "test-deployment",
				Namespace:   "default",
				Namespaced:  true,
			},
			setupMocks: func() (*mockResourceRepository, *mockGVRResolver) {
				repo := &mockResourceRepository{
					patchFunc: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool, patchData []byte, patchType types.PatchType, options metav1.PatchOptions) (*unstructured.Unstructured, error) {
						if name != "test-deployment" || namespace != "default" {
							return nil, errors.New("unexpected parameters")
						}
						obj := &unstructured.Unstructured{}
						obj.SetName(name)
						obj.SetNamespace(namespace)
						return obj, nil
					},
				}
				resolver := &mockGVRResolver{
					resolveFunc: func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
						return schema.GroupVersionResource{
							Group:    "apps",
							Version:  "v1",
							Resource: "deployments",
						}, models.ResourceMeta{Namespaced: true}, nil
					},
				}
				return repo, resolver
			},
			wantErr: false,
		},
		{
			name: "invalid YAML",
			req: UpdateResourceRequest{
				YAMLContent: "invalid: yaml: content: [",
				Kind:        "Deployment",
			},
			setupMocks: func() (*mockResourceRepository, *mockGVRResolver) {
				return &mockResourceRepository{}, &mockGVRResolver{}
			},
			wantErr: true,
			errMsg:  "invalid YAML",
		},
		{
			name: "GVR resolution error",
			req: UpdateResourceRequest{
				YAMLContent: validYAML,
				Kind:        "Unknown",
			},
			setupMocks: func() (*mockResourceRepository, *mockGVRResolver) {
				repo := &mockResourceRepository{}
				resolver := &mockGVRResolver{
					resolveFunc: func(ctx context.Context, kind, apiVersion string, namespacedParam string) (schema.GroupVersionResource, models.ResourceMeta, error) {
						return schema.GroupVersionResource{}, models.ResourceMeta{}, errors.New("failed to resolve GVR")
					},
				}
				return repo, resolver
			},
			wantErr: true,
			errMsg:  "failed to resolve GVR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, resolver := tt.setupMocks()
			service := NewResourceService(repo, resolver)

			// Create context with user for permission checks
			ctx := context.WithValue(context.Background(), auth.UserContextKey(), &auth.AuthClaims{
				Claims: models.Claims{
					Username: "admin",
					Role:     "admin",
				},
			})

			err := service.UpdateResource(ctx, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("UpdateResource() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
