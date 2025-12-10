package pod

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestK8sLogRepository_GetLogStream(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL
		expectedPath := "/api/v1/namespaces/default/pods/test-pod/log"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify Query Params
		if r.URL.Query().Get("container") != "main" {
			t.Errorf("Expected container=main, got %s", r.URL.Query().Get("container"))
		}
		if r.URL.Query().Get("follow") != "true" {
			t.Errorf("Expected follow=true, got %s", r.URL.Query().Get("follow"))
		}

		// Write log data
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("log line 1\nlog line 2\n"))
	}))
	defer server.Close()

	// Create client pointing to mock server
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

	repo := NewK8sLogRepository(client)
	ctx := context.Background()
	opts := &corev1.PodLogOptions{
		Container: "main",
		Follow:    true,
	}

	stream, err := repo.GetLogStream(ctx, "default", "test-pod", opts)
	if err != nil {
		t.Fatalf("GetLogStream() unexpected error: %v", err)
	}
	defer stream.Close()

	// Read stream
	body, err := io.ReadAll(stream)
	if err != nil {
		t.Fatalf("Failed to read stream: %v", err)
	}

	expected := "log line 1\nlog line 2\n"
	if string(body) != expected {
		t.Errorf("Expected logs %q, got %q", expected, string(body))
	}
}
