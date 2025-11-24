package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/example/k8s-view/internal/models"
)

// Handlers is kept for backward compatibility with handlers not yet migrated
// New code should use the service modules directly
// This embeds models.Handlers to maintain compatibility
type Handlers struct {
	*models.Handlers
}

// getClient, getDynamicClient, getMetricsClient are kept for backward compatibility
// They are used by Prometheus handlers
func (h *Handlers) getClient(r *http.Request) (*kubernetes.Clientset, error) {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}
	h.RLock()
	defer h.RUnlock()
	client, ok := h.Clients[cluster]
	if !ok {
		return nil, fmt.Errorf("cluster not found: %s", cluster)
	}
	return client, nil
}

func (h *Handlers) getDynamicClient(r *http.Request) (dynamic.Interface, error) {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}
	h.RLock()
	defer h.RUnlock()
	client, ok := h.Dynamics[cluster]
	if !ok {
		return nil, fmt.Errorf("cluster not found: %s", cluster)
	}
	return client, nil
}

func (h *Handlers) getMetricsClient(r *http.Request) *metricsv.Clientset {
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}
	h.RLock()
	defer h.RUnlock()
	client, ok := h.Metrics[cluster]
	if !ok {
		return nil
	}
	return client
}

// Type aliases for backward compatibility
type ClusterConfig = models.ClusterConfig
type Namespace = models.Namespace
type Resource = models.Resource
type DeploymentDetails = models.DeploymentDetails
type PodMetric = models.PodMetric
type ResourceMeta = models.ResourceMeta

var resourceMeta = models.ResourceMetaMap
var kindAliases = models.KindAliases

// Functions for backward compatibility
func normalizeKind(kind string) string {
	return models.NormalizeKind(kind)
}

func resolveGVR(kind string) (schema.GroupVersionResource, bool) {
	return models.ResolveGVR(kind)
}

// HealthHandler is an unauthenticated liveness endpoint
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

const DataDir = "./data"

// GetLogo serves the custom logo file
func (h *Handlers) GetLogo(w http.ResponseWriter, r *http.Request) {
	// Check for supported extensions
	extensions := []string{".png", ".svg"}
	var foundPath string

	for _, ext := range extensions {
		path := filepath.Join(DataDir, "logo"+ext)
		if _, err := os.Stat(path); err == nil {
			foundPath = path
			break
		}
	}

	if foundPath == "" {
		// Logo not found - return 404
		// Frontend will handle this gracefully and use default logo
		http.Error(w, "Logo not found", http.StatusNotFound)
		return
	}

	absPath, _ := filepath.Abs(foundPath)
	fmt.Printf("Serving logo from: %s\n", absPath)
	http.ServeFile(w, r, foundPath)
}

// UploadLogo handles logo file uploads
func (h *Handlers) UploadLogo(w http.ResponseWriter, r *http.Request) {
	// Limit upload size to 5MB
	r.ParseMultipartForm(5 << 20)

	file, handler, err := r.FormFile("logo")
	if err != nil {
		fmt.Printf("Error retrieving file: %v\n", err)
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if handler.Size > 5<<20 {
		http.Error(w, "File too large (max 5MB)", http.StatusBadRequest)
		return
	}

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}
	file.Seek(0, 0) // Reset pointer

	contentType := http.DetectContentType(buffer)

	// Validate extension
	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if ext != ".png" && ext != ".svg" {
		http.Error(w, "Invalid file type. Only .png and .svg are allowed", http.StatusBadRequest)
		return
	}

	if ext == ".png" && contentType != "image/png" {
		http.Error(w, "Invalid file content (not a PNG)", http.StatusBadRequest)
		return
	}
	// For SVG, we perform a basic check for script tags to prevent XSS
	if ext == ".svg" {
		// Read the whole file to check content
		content, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusBadRequest)
			return
		}
		// Reset file pointer for later copy
		file.Seek(0, 0)

		contentStr := strings.ToLower(string(content))
		if strings.Contains(contentStr, "<script") || 
		   strings.Contains(contentStr, "javascript:") || 
		   strings.Contains(contentStr, "onload=") || 
		   strings.Contains(contentStr, "onerror=") {
			http.Error(w, "Invalid SVG content: script tags or event handlers are not allowed", http.StatusBadRequest)
			return
		}
	}

	// Ensure data directory exists
	if err := os.MkdirAll(DataDir, 0755); err != nil {
		fmt.Printf("Error creating data directory: %v\n", err)
		http.Error(w, "Error creating data directory", http.StatusInternalServerError)
		return
	}

	// Remove existing logos to avoid conflicts
	os.Remove(filepath.Join(DataDir, "logo.png"))
	os.Remove(filepath.Join(DataDir, "logo.svg"))

	// Create destination file
	destPath := filepath.Join(DataDir, "logo"+ext)
	absPath, _ := filepath.Abs(destPath)
	fmt.Printf("Saving logo to: %s\n", absPath)

	dst, err := os.Create(destPath)
	if err != nil {
		fmt.Printf("Error creating destination file: %v\n", err)
		http.Error(w, "Error creating destination file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		fmt.Printf("Error saving file content: %v\n", err)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logo uploaded successfully"))
}
