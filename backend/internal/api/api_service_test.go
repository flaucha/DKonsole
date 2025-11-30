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
