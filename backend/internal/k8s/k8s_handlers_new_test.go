package k8s

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func TestGetResourceYAML_Handler(t *testing.T) {
	// Setup Mocks
	mockClusterService := new(MockClusterService)
	handlers := &models.Handlers{}

	// Prepare Service
	service := NewService(handlers, mockClusterService)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/k8s/yaml?kind=Deployment&name=dep1&namespace=ns1&namespaced=true", nil)
	rr := httptest.NewRecorder()

	// Inject User Context with permissions
	claims := &models.Claims{
		Username: "admin",
		Role:     "admin",
		Permissions: map[string]string{
			"ns1": "view",
		},
	}
	ctx := context.WithValue(req.Context(), auth.UserContextKey(), &auth.AuthClaims{Claims: *claims})
	req = req.WithContext(ctx)

	// Mock Expectations
	// 1. GetDynamicClient
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)
	mockClusterService.On("GetDynamicClient", mock.Anything).Return(dynamicClient, nil)

	// 2. GetClient (for discovery)
	k8sClient := k8sfake.NewSimpleClientset()
	mockClusterService.On("GetClient", mock.Anything).Return(k8sClient, nil)

	// Execute
	service.GetResourceYAML(rr, req)

	// Assert
	// Initially it might fail with NotFound because fake dynamic client is empty
	// Or InternalServerError if GVR resolution fails?
	// ResourceYAML handler creates Repo and calls Service.
	// Service resolves GVR. If discovery fails (empty), it infers.
	// Then calls Repo.Get.
	// Repo.Get uses dynamic Client. Client has no object "dep1".
	// Should return NotFound (404).

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (resource missing)", rr.Code)
	}
}

func TestGetResourceYAML_Handler_Success(t *testing.T) {
	mockClusterService := new(MockClusterService)
	service := NewService(&models.Handlers{}, mockClusterService)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/k8s/yaml?kind=Deployment&name=dep1&namespace=ns1&namespaced=true", nil)
	rr := httptest.NewRecorder()

	claims := &models.Claims{
		Username:    "admin",
		Role:        "admin",
		Permissions: map[string]string{"ns1": "view"},
	}
	ctx := context.WithValue(req.Context(), auth.UserContextKey(), &auth.AuthClaims{Claims: *claims})
	req = req.WithContext(ctx)

	// Mock Dynamic Client with resource
	scheme := runtime.NewScheme()

	// Need to register Deployment kind? "apps/v1"
	// Without scheme registration, unstructured might work fine?
	// Fake client handles GVR automatically usually.

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)
	gvr := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "dep1",
				"namespace": "ns1",
			},
		},
	}
	_, _ = dynamicClient.Resource(gvr).Namespace("ns1").Create(context.Background(), obj, metav1.CreateOptions{})

	mockClusterService.On("GetDynamicClient", mock.Anything).Return(dynamicClient, nil)

	k8sClient := k8sfake.NewSimpleClientset()
	mockClusterService.On("GetClient", mock.Anything).Return(k8sClient, nil)

	service.GetResourceYAML(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/x-yaml" {
		t.Errorf("content-type = %s, want application/x-yaml", rr.Header().Get("Content-Type"))
	}
}

func TestGetResourceYAML_MissingParams(t *testing.T) {
	mockClusterService := new(MockClusterService)
	service := NewService(&models.Handlers{}, mockClusterService)

	req := httptest.NewRequest(http.MethodGet, "/api/k8s/yaml", nil) // Missing params
	rr := httptest.NewRecorder()

	service.GetResourceYAML(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
