package pod

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func TestExecService_CreateExecutor(t *testing.T) {
	// Backup and restore the original function
	originalFunc := spdyExecutorFunc
	defer func() { spdyExecutorFunc = originalFunc }()

	// Start a local HTTP server to mock K8s API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify expected calls
		if r.Method == "POST" && r.URL.Path == "/api/v1/namespaces/default/pods/test-pod/exec" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == "POST" && r.URL.Path == "/api/v1/namespaces/error-ns/pods/error-pod/exec" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Create a real client pointing to the mock server
	config := &rest.Config{
		Host: server.URL,
		ContentConfig: rest.ContentConfig{
			GroupVersion: &corev1.SchemeGroupVersion,
		},
		APIPath: "/api",
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Mock SPDY executor creation
	spdyExecutorFunc = func(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error) {
		// Verify URL contains expected parameters
		if url.String() == "" {
			return nil, errors.New("empty url")
		}
		if url.Path == "/api/v1/namespaces/error-ns/pods/error-pod/exec" {
			return nil, errors.New("executor creation failed")
		}
		return &mockExecutor{}, nil
	}

	service := NewExecService()

	t.Run("Success", func(t *testing.T) {
		req := ExecRequest{
			Namespace: "default",
			PodName:   "test-pod",
			Container: "test-container",
		}

		executor, urlStr, err := service.CreateExecutor(client, config, req)
		if err != nil {
			t.Fatalf("CreateExecutor() unexpected error: %v", err)
		}
		if executor == nil {
			t.Error("CreateExecutor() executor is nil")
		}
		if urlStr == "" {
			t.Error("CreateExecutor() url is empty")
		}
	})

	t.Run("Executor Creation Fail", func(t *testing.T) {
		req := ExecRequest{
			Namespace: "error-ns",  // Trigger mock error
			PodName:   "error-pod", // Trigger mock error
			Container: "test-container",
		}

		_, _, err := service.CreateExecutor(client, config, req)
		if err == nil {
			t.Error("CreateExecutor() expected error when executor creation fails")
		}
	})
}
