package api

import (
	"context"
	"errors"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

// mockAPIResourceRepository is a mock implementation of APIResourceRepository
type mockAPIResourceRepository struct {
	listAPIResourcesFunc func(ctx context.Context) ([]*metav1.APIResourceList, error)
}

func (m *mockAPIResourceRepository) ListAPIResources(ctx context.Context) ([]*metav1.APIResourceList, error) {
	if m.listAPIResourcesFunc != nil {
		return m.listAPIResourcesFunc(ctx)
	}
	return []*metav1.APIResourceList{}, nil
}

func (m *mockAPIResourceRepository) ListAPIResourceObjects(ctx context.Context, gvr schema.GroupVersionResource, namespace string, listOpts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return nil, errors.New("not implemented")
}

// mockDynamicResourceRepository is a mock implementation of DynamicResourceRepository
type mockDynamicResourceRepository struct {
	getFunc  func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error)
	listFunc func(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, limit int64, continueToken string) (*unstructured.UnstructuredList, string, error)
}

func (m *mockDynamicResourceRepository) Get(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, gvr, name, namespace, namespaced)
	}
	return nil, errors.New("get not implemented")
}

func (m *mockDynamicResourceRepository) List(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, limit int64, continueToken string) (*unstructured.UnstructuredList, string, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, gvr, namespace, namespaced, limit, continueToken)
	}
	return nil, "", errors.New("list not implemented")
}

type mockCRDRepository struct {
	listFunc func(ctx context.Context, limit int64, continueToken string) (*unstructured.UnstructuredList, error)
}

func (m *mockCRDRepository) ListCRDs(ctx context.Context, limit int64, continueToken string) (*unstructured.UnstructuredList, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, continueToken)
	}
	return nil, errors.New("list not implemented")
}

func TestAPIService_ListAPIResources(t *testing.T) {
	tests := []struct {
		name                  string
		listAPIResourcesFunc  func(ctx context.Context) ([]*metav1.APIResourceList, error)
		wantErr               bool
		errMsg                string
		expectedResourceCount int
	}{
		{
			name: "successful list API resources",
			listAPIResourcesFunc: func(ctx context.Context) ([]*metav1.APIResourceList, error) {
				return []*metav1.APIResourceList{
					{
						GroupVersion: "v1",
						APIResources: []metav1.APIResource{
							{
								Name:       "pods",
								Kind:       "Pod",
								Namespaced: true,
							},
							{
								Name:       "services",
								Kind:       "Service",
								Namespaced: true,
							},
						},
					},
				}, nil
			},
			wantErr:               false,
			expectedResourceCount: 2,
		},
		{
			name: "filter out subresources (contain /)",
			listAPIResourcesFunc: func(ctx context.Context) ([]*metav1.APIResourceList, error) {
				return []*metav1.APIResourceList{
					{
						GroupVersion: "v1",
						APIResources: []metav1.APIResource{
							{
								Name:       "pods",
								Kind:       "Pod",
								Namespaced: true,
							},
							{
								Name:       "pods/exec", // Should be filtered out
								Kind:       "PodExecOptions",
								Namespaced: true,
							},
						},
					},
				}, nil
			},
			wantErr:               false,
			expectedResourceCount: 1, // pods/exec should be filtered
		},
		{
			name: "repository error",
			listAPIResourcesFunc: func(ctx context.Context) ([]*metav1.APIResourceList, error) {
				return nil, errors.New("repository error")
			},
			wantErr: true,
			errMsg:  "failed to list API resources",
		},
		{
			name:                 "discovery repository not set",
			listAPIResourcesFunc: nil,
			wantErr:              true,
			errMsg:               "discovery repository not set",
		},
		{
			name: "empty API resources list",
			listAPIResourcesFunc: func(ctx context.Context) ([]*metav1.APIResourceList, error) {
				return []*metav1.APIResourceList{}, nil
			},
			wantErr:               false,
			expectedResourceCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockRepo APIResourceRepository
			if tt.listAPIResourcesFunc != nil || tt.name != "discovery repository not set" {
				mockRepo = &mockAPIResourceRepository{
					listAPIResourcesFunc: tt.listAPIResourcesFunc,
				}
			}

			service := NewAPIService(mockRepo, nil)
			ctx := context.Background()

			result, err := service.ListAPIResources(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListAPIResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("ListAPIResources() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ListAPIResources() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			if len(result) != tt.expectedResourceCount {
				t.Errorf("ListAPIResources() count = %v, want %v", len(result), tt.expectedResourceCount)
			}

			// Verify structure
			for _, res := range result {
				if res.Resource == "" {
					t.Errorf("ListAPIResources() resource name is empty")
				}
				if res.Kind == "" {
					t.Errorf("ListAPIResources() kind is empty")
				}
			}
		})
	}
}

func TestAPIService_GetResourceYAML(t *testing.T) {
	validYAMLObject := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-cm",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"key": "value",
			},
		},
	}

	tests := []struct {
		name       string
		request    GetResourceYAMLRequest
		ctx        context.Context
		getFunc    func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error)
		wantErr    bool
		errMsg     string
		expectYAML bool
	}{
		{
			name: "successful get resource YAML",
			request: GetResourceYAMLRequest{
				Group:      "",
				Version:    "v1",
				Resource:   "configmaps",
				Name:       "test-cm",
				Namespace:  "default",
				Namespaced: true,
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
			getFunc: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error) {
				return validYAMLObject, nil
			},
			wantErr:    false,
			expectYAML: true,
		},
		{
			name: "resource repository not set",
			request: GetResourceYAMLRequest{
				Group:      "",
				Version:    "v1",
				Resource:   "configmaps",
				Name:       "test-cm",
				Namespace:  "default",
				Namespaced: true,
			},
			ctx:     context.Background(),
			getFunc: nil,
			wantErr: true,
			errMsg:  "resource repository not set",
		},
		{
			name: "repository get error",
			request: GetResourceYAMLRequest{
				Group:      "",
				Version:    "v1",
				Resource:   "configmaps",
				Name:       "test-cm",
				Namespace:  "default",
				Namespaced: true,
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
			getFunc: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, namespaced bool) (*unstructured.Unstructured, error) {
				return nil, errors.New("resource not found")
			},
			wantErr: true,
			errMsg:  "failed to get resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockRepo DynamicResourceRepository
			if tt.getFunc != nil {
				mockRepo = &mockDynamicResourceRepository{
					getFunc: tt.getFunc,
				}
			}

			service := NewAPIService(nil, mockRepo)

			yamlData, err := service.GetResourceYAML(tt.ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetResourceYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetResourceYAML() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("GetResourceYAML() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			if tt.expectYAML {
				if yamlData == "" {
					t.Errorf("GetResourceYAML() yamlData is empty")
				}
				// Verify YAML contains expected content
				if !strings.Contains(yamlData, "kind: ConfigMap") {
					t.Errorf("GetResourceYAML() yamlData doesn't contain expected kind")
				}
			}
		})
	}
}

func TestAPIService_ListAPIResourceObjects(t *testing.T) {
	allowedCtx := context.WithValue(context.Background(), auth.UserContextKey(), &auth.AuthClaims{
		Claims: models.Claims{
			Username:    "viewer",
			Permissions: map[string]string{"allowed": "view"},
		},
	})

	allowedItem := unstructured.Unstructured{}
	allowedItem.SetName("ok")
	allowedItem.SetNamespace("allowed")
	allowedItem.SetCreationTimestamp(metav1.Now())
	allowedItem.SetKind("Demo")
	allowedItem.Object["status"] = map[string]interface{}{"phase": "Running"}

	deniedItem := unstructured.Unstructured{}
	deniedItem.SetName("nope")
	deniedItem.SetNamespace("denied")
	deniedItem.SetCreationTimestamp(metav1.Now())
	deniedItem.SetKind("Demo")

	tests := []struct {
		name         string
		ctx          context.Context
		req          ListAPIResourceObjectsRequest
		listFunc     func(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, limit int64, continueToken string) (*unstructured.UnstructuredList, string, error)
		wantErr      bool
		errMsg       string
		wantCount    int
		wantContinue string
	}{
		{
			name: "happy path filters unauthorized namespaces",
			ctx:  allowedCtx,
			req: ListAPIResourceObjectsRequest{
				Group:      "demo.k8s",
				Version:    "v1",
				Resource:   "demos",
				Namespaced: true,
			},
			listFunc: func(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, limit int64, continueToken string) (*unstructured.UnstructuredList, string, error) {
				return &unstructured.UnstructuredList{
					Items: []unstructured.Unstructured{allowedItem, deniedItem},
				}, "token123", nil
			},
			wantErr:      false,
			wantCount:    1,
			wantContinue: "token123",
		},
		{
			name: "deny namespace without permission",
			ctx:  allowedCtx,
			req: ListAPIResourceObjectsRequest{
				Group:      "demo.k8s",
				Version:    "v1",
				Resource:   "demos",
				Namespace:  "denied",
				Namespaced: true,
			},
			listFunc: func(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, limit int64, continueToken string) (*unstructured.UnstructuredList, string, error) {
				return &unstructured.UnstructuredList{}, "", nil
			},
			wantErr: true,
			errMsg:  "access denied",
		},
		{
			name:    "resource repository not set",
			ctx:     allowedCtx,
			req:     ListAPIResourceObjectsRequest{Resource: "demos"},
			wantErr: true,
			errMsg:  "resource repository not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var repo DynamicResourceRepository
			if tt.listFunc != nil {
				repo = &mockDynamicResourceRepository{listFunc: tt.listFunc}
			}
			service := NewAPIService(nil, repo)

			resp, err := service.ListAPIResourceObjects(tt.ctx, tt.req)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ListAPIResourceObjects() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.errMsg != "" && (err == nil || !strings.Contains(err.Error(), tt.errMsg)) {
					t.Fatalf("expected error containing %q, got %v", tt.errMsg, err)
				}
				return
			}
			if resp == nil {
				t.Fatalf("response is nil")
			}
			if len(resp.Resources) != tt.wantCount {
				t.Fatalf("expected %d resources, got %d", tt.wantCount, len(resp.Resources))
			}
			if resp.Continue != tt.wantContinue {
				t.Fatalf("expected continue token %q, got %q", tt.wantContinue, resp.Continue)
			}
			if resp.Resources[0].Name != "ok" || resp.Resources[0].Namespace != "allowed" {
				t.Fatalf("unexpected resource %+v", resp.Resources[0])
			}
		})
	}
}

func TestCRDService_GetCRDs(t *testing.T) {
	validCRD := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{"name": "demos.demo.k8s.io"},
			"spec": map[string]interface{}{
				"group": "demo.k8s.io",
				"names": map[string]interface{}{
					"kind": "Demo",
				},
				"scope": "Namespaced",
				"versions": []interface{}{
					map[string]interface{}{
						"name":   "v1",
						"served": true,
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		listFunc func(ctx context.Context, limit int64, continueToken string) (*unstructured.UnstructuredList, error)
		wantErr  bool
		wantLen  int
	}{
		{
			name: "happy path",
			listFunc: func(ctx context.Context, limit int64, continueToken string) (*unstructured.UnstructuredList, error) {
				return &unstructured.UnstructuredList{
					Items: []unstructured.Unstructured{validCRD},
				}, nil
			},
			wantLen: 1,
		},
		{
			name: "repo error",
			listFunc: func(ctx context.Context, limit int64, continueToken string) (*unstructured.UnstructuredList, error) {
				return nil, errors.New("boom")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCRDService(&mockCRDRepository{listFunc: tt.listFunc})
			resp, err := service.GetCRDs(context.Background(), GetCRDsRequest{Limit: 10})

			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCRDs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if resp == nil {
				t.Fatalf("response is nil")
			}
			if len(resp.CRDs) != tt.wantLen {
				t.Fatalf("expected %d CRDs, got %d", tt.wantLen, len(resp.CRDs))
			}
			if resp.CRDs[0].Name != "demos.demo.k8s.io" || resp.CRDs[0].Version != "v1" {
				t.Fatalf("unexpected CRD %+v", resp.CRDs[0])
			}
		})
	}
}

func TestAPIService_GetCRDResources(t *testing.T) {
	listItem := unstructured.Unstructured{}
	listItem.SetName("demo1")
	listItem.SetNamespace("default")
	listItem.SetCreationTimestamp(metav1.Now())

	tests := []struct {
		name     string
		listFunc func(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, limit int64, continueToken string) (*unstructured.UnstructuredList, string, error)
		wantErr  bool
		wantLen  int
	}{
		{
			name: "happy path",
			listFunc: func(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, limit int64, continueToken string) (*unstructured.UnstructuredList, string, error) {
				return &unstructured.UnstructuredList{
					Items: []unstructured.Unstructured{listItem},
				}, "next", nil
			},
			wantLen: 1,
		},
		{
			name: "repository error",
			listFunc: func(ctx context.Context, gvr schema.GroupVersionResource, namespace string, namespaced bool, limit int64, continueToken string) (*unstructured.UnstructuredList, string, error) {
				return nil, "", errors.New("fail list")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewAPIService(nil, &mockDynamicResourceRepository{listFunc: tt.listFunc})
			resp, err := service.GetCRDResources(context.Background(), GetCRDResourcesRequest{
				Group:    "demo.k8s.io",
				Version:  "v1",
				Resource: "demos",
			})

			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCRDResources() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if resp == nil {
				t.Fatalf("response is nil")
			}
			if len(resp.Resources) != tt.wantLen {
				t.Fatalf("expected %d resources, got %d", tt.wantLen, len(resp.Resources))
			}
			if resp.Continue != "next" {
				t.Fatalf("expected continue token next, got %q", resp.Continue)
			}
		})
	}
}
