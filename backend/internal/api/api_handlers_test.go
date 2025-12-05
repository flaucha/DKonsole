package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func apiTestScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add corev1: %v", err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatalf("add appsv1: %v", err)
	}
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"}, &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinitionList"}, &unstructured.UnstructuredList{})
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Widget"}, &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "WidgetList"}, &unstructured.UnstructuredList{})
	return scheme
}

func adminRequest(req *http.Request) *http.Request {
	claims := &auth.AuthClaims{Claims: models.Claims{Role: "admin"}}
	return req.WithContext(context.WithValue(req.Context(), auth.UserContextKey(), claims))
}

func TestListAPIResources_HandlerSuccessAndError(t *testing.T) {
	k8sClient := k8sfake.NewSimpleClientset()
	handlers := &models.Handlers{
		Clients:  map[string]kubernetes.Interface{"default": k8sClient},
		Dynamics: map[string]dynamic.Interface{},
	}
	service := NewService(cluster.NewService(handlers))

	req := httptest.NewRequest(http.MethodGet, "/api/resources", nil)
	rr := httptest.NewRecorder()
	service.ListAPIResources(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// missing client
	serviceErr := NewService(cluster.NewService(&models.Handlers{}))
	rr2 := httptest.NewRecorder()
	serviceErr.ListAPIResources(rr2, req)
	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr2.Code)
	}
}

func TestListAPIResourceObjects_Handler(t *testing.T) {
	scheme := apiTestScheme(t)
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "default"}}
	dyn := dynamicfake.NewSimpleDynamicClient(scheme, pod)

	handlers := &models.Handlers{
		Clients:  map[string]kubernetes.Interface{"default": k8sfake.NewSimpleClientset()},
		Dynamics: map[string]dynamic.Interface{"default": dyn},
	}
	service := NewService(cluster.NewService(handlers))

	req := adminRequest(httptest.NewRequest(http.MethodGet, "/api/resource/objects?resource=pods&version=v1&namespace=default&namespaced=true", nil))
	rr := httptest.NewRecorder()

	service.ListAPIResourceObjects(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", rr.Code, rr.Body.String())
	}
}

func TestGetAPIResourceYAML_Handler(t *testing.T) {
	scheme := apiTestScheme(t)
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}}
	dyn := dynamicfake.NewSimpleDynamicClient(scheme, node)
	handlers := &models.Handlers{
		Dynamics: map[string]dynamic.Interface{"default": dyn},
	}
	service := NewService(cluster.NewService(handlers))

	req := adminRequest(httptest.NewRequest(http.MethodGet, "/api/resource/yaml?resource=nodes&version=v1&name=node1&namespaced=false", nil))
	rr := httptest.NewRecorder()

	service.GetAPIResourceYAML(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", rr.Code, rr.Body.String())
	}
	if rr.Body.Len() == 0 {
		t.Fatalf("expected YAML body")
	}
}

func TestGetCRDs_Handler(t *testing.T) {
	scheme := apiTestScheme(t)
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "widgets.example.com",
			},
		},
	}
	crd.SetGroupVersionKind(schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"})
	dyn := dynamicfake.NewSimpleDynamicClient(scheme, crd)
	handlers := &models.Handlers{Dynamics: map[string]dynamic.Interface{"default": dyn}}
	service := NewService(cluster.NewService(handlers))

	req := httptest.NewRequest(http.MethodGet, "/api/crds", nil)
	rr := httptest.NewRecorder()

	service.GetCRDs(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetCRDResourcesAndYAML_Handler(t *testing.T) {
	scheme := apiTestScheme(t)
	widget := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "example.com/v1",
			"kind":       "Widget",
			"metadata": map[string]interface{}{
				"name":      "w1",
				"namespace": "default",
			},
		},
	}
	widget.SetGroupVersionKind(schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Widget"})

	dyn := dynamicfake.NewSimpleDynamicClient(scheme, widget)
	handlers := &models.Handlers{Dynamics: map[string]dynamic.Interface{"default": dyn}}
	service := NewService(cluster.NewService(handlers))

	listReq := adminRequest(httptest.NewRequest(http.MethodGet, "/api/crd/resources?group=example.com&version=v1&resource=widgets&namespace=default&namespaced=true", nil))
	rr := httptest.NewRecorder()
	service.GetCRDResources(rr, listReq)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	yamlReq := adminRequest(httptest.NewRequest(http.MethodGet, "/api/crd/yaml?group=example.com&version=v1&resource=widgets&name=w1&namespace=default&namespaced=true", nil))
	rr2 := httptest.NewRecorder()
	service.GetCRDYaml(rr2, yamlReq)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}
	if rr2.Body.Len() == 0 {
		t.Fatalf("expected YAML body")
	}
}
