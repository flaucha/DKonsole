package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/flaucha/DKonsole/backend/internal/middleware"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return middleware.IsRequestOriginAllowed(r)
	},
}

// StreamResourceCreation handles SSE for resource creation feedback
func (s *Service) StreamResourceCreation(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin != "" && !middleware.IsRequestOriginAllowed(r) {
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		return
	}
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Read YAML from body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"message\": \"Failed to read body: %v\"}\n\n", err)
		flusher.Flush()
		return
	}
	defer r.Body.Close()

	yamlData := string(body)

	// Send "start" event
	fmt.Fprintf(w, "event: status\ndata: {\"message\": \"Starting creation...\"}\n\n")
	flusher.Flush()

	// Create context
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Get Dynamic Client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		errorData, _ := json.Marshal(map[string]string{"message": err.Error()})
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", errorData)
		flusher.Flush()
		return
	}

	// Get Client (for discovery)
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		errorData, _ := json.Marshal(map[string]string{"message": err.Error()})
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", errorData)
		flusher.Flush()
		return
	}

	// Create Resource Service
	resourceService := s.serviceFactory.CreateResourceService(dynamicClient, client)

	// Attempt creation
	result, err := resourceService.CreateResource(ctx, yamlData)

	if err != nil {
		// Send error event
		errorData, _ := json.Marshal(map[string]string{"message": err.Error()})
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", errorData)
		flusher.Flush()
	} else {
		// Send success event
		successData, _ := json.Marshal(result)
		fmt.Fprintf(w, "event: success\ndata: %s\n\n", successData)
		flusher.Flush()
	}
}

// WatchResources handles WebSocket connections for watching Kubernetes resources
func (s *Service) WatchResources(w http.ResponseWriter, r *http.Request) {
	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.LogError(err, "Failed to upgrade to WebSocket", nil)
		return
	}
	defer conn.Close()

	// Parse parameters
	kind := r.URL.Query().Get("kind")
	namespace := r.URL.Query().Get("namespace")
	allNamespaces := r.URL.Query().Get("namespace") == "all"

	// Get Dynamic Client
	dynamicClient, err := s.clusterService.GetDynamicClient(r)
	if err != nil {
		_ = conn.WriteJSON(map[string]string{"type": "ERROR", "message": err.Error()})
		return
	}

	// Create WatchService
	watchService := s.serviceFactory.CreateWatchService()

	// Create context that is canceled when connection closes
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Handle client disconnect
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				cancel()
				return
			}
		}
	}()

	// Start Watch
	req := WatchRequest{
		Kind:          kind,
		Namespace:     namespace,
		AllNamespaces: allNamespaces,
	}

	watcher, err := watchService.StartWatch(ctx, dynamicClient, req)
	if err != nil {
		_ = conn.WriteJSON(map[string]string{"type": "ERROR", "message": err.Error()})
		return
	}
	defer watcher.Stop()

	// Send events to client
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return
			}

			// Transform event
			result, err := watchService.TransformEvent(event)
			if err != nil {
				continue
			}

			// Send to client
			if err := conn.WriteJSON(result); err != nil {
				return
			}
		}
	}
}
