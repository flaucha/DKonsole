package pod

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/permissions"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// Service provides HTTP handlers for pod-specific operations including log streaming and exec.
type Service struct {
	handlers       *models.Handlers
	clusterService *cluster.Service
}

// NewService creates a new pod service with the provided handlers and cluster service.
func NewService(h *models.Handlers, cs *cluster.Service) *Service {
	return &Service{
		handlers:       h,
		clusterService: cs,
		// PodService will be created on-demand in handlers that need it
		// This maintains backward compatibility while allowing refactoring
	}
}

// StreamPodLogs handles HTTP GET requests to stream logs from a Kubernetes pod.
// Query parameters:
//   - namespace: The namespace containing the pod
//   - pod: The pod name
//   - container: Optional container name (if pod has multiple containers)
//
// Returns a streaming text/plain response with pod logs. The connection remains open
// and logs are streamed in real-time as they are generated.
func (s *Service) StreamPodLogs(w http.ResponseWriter, r *http.Request) {
	// Parse and validate HTTP parameters
	params, err := utils.ParsePodParams(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate namespace access
	ctx := r.Context()
	hasAccess, err := permissions.HasNamespaceAccess(ctx, params.Namespace)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check permissions: %v", err))
		return
	}
	if !hasAccess {
		utils.ErrorResponse(w, http.StatusForbidden, fmt.Sprintf("Access denied to namespace: %s", params.Namespace))
		return
	}

	// Get Kubernetes client for this request
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create repository for this request
	logRepo := NewK8sLogRepository(client)

	// Create service with repository
	logService := NewLogService(logRepo)

	// Create context with timeout
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Prepare request
	streamReq := StreamLogsRequest{
		Namespace: params.Namespace,
		PodName:   params.PodName,
		Container: params.Container,
		Follow:    true,
	}

	// Call service to get log stream (business logic layer)
	stream, err := logService.StreamLogs(ctx, streamReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to open log stream: %v", err))
		return
	}
	defer stream.Close()

	// Set HTTP headers for streaming
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Connection", "keep-alive")

	// Check if streaming is supported
	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	// Stream logs (HTTP layer - streaming loop remains here for efficiency)
	buf := make([]byte, 1024)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			if _, wErr := w.Write(buf[:n]); wErr != nil {
				break
			}
			flusher.Flush()
		}
		if err != nil {
			break
		}
	}
}

// GetPodEvents handles HTTP GET requests to retrieve Kubernetes events for a specific pod.
// Query parameters:
//   - namespace: The namespace containing the pod
//   - pod: The pod name
//
// Returns a JSON array of events sorted by LastSeen timestamp (most recent first).
func (s *Service) GetPodEvents(w http.ResponseWriter, r *http.Request) {
	// Parse and validate HTTP parameters
	params, err := utils.ParsePodParams(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate namespace access
	ctx := r.Context()
	hasAccess, err := permissions.HasNamespaceAccess(ctx, params.Namespace)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check permissions: %v", err))
		return
	}
	if !hasAccess {
		utils.ErrorResponse(w, http.StatusForbidden, fmt.Sprintf("Access denied to namespace: %s", params.Namespace))
		return
	}

	// Get Kubernetes client for this request
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create repository for this request
	eventRepo := NewK8sEventRepository(client)

	// Create service with repository
	podService := NewPodService(eventRepo)

	// Create context with timeout
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Call service to get events (business logic layer)
	events, err := podService.GetPodEvents(ctx, params.Namespace, params.PodName)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get pod events: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, events)
}

// ExecIntoPod provides WebSocket-based terminal access to a pod
// Refactored to use layered architecture:
// Handler (HTTP/WebSocket) -> Service (Business Logic) -> Repository (Data Access)
func (s *Service) ExecIntoPod(w http.ResponseWriter, r *http.Request) {
	// Parse and validate HTTP parameters
	params, err := utils.ParsePodParams(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate namespace access
	ctx := r.Context()
	canEdit, err := permissions.CanPerformAction(ctx, params.Namespace, "edit")
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check permissions: %v", err))
		return
	}
	if !canEdit {
		utils.ErrorResponse(w, http.StatusForbidden, fmt.Sprintf("Edit permission required for namespace: %s", params.Namespace))
		return
	}

	// Get Kubernetes client for this request
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get REST config for exec
	restConfig, err := s.clusterService.GetRESTConfig(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create exec service
	execService := NewExecService()

	// Prepare exec request
	execReq := ExecRequest{
		Namespace: params.Namespace,
		PodName:   params.PodName,
		Container: params.Container,
	}

	// Create executor (business logic layer)
	executor, _, err := execService.CreateExecutor(client, restConfig, execReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create executor: %v", err))
		return
	}

	// Upgrade HTTP connection to WebSocket (HTTP layer - WebSocket handling remains here)
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")

			// For WebSocket, be more permissive - allow empty origin for same-origin connections
			if origin == "" {
				return true
			}

			originURL, err := url.Parse(origin)
			if err != nil {
				return false
			}

			// Check allowed origins from env
			allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
			if allowedOrigins != "" {
				origins := strings.Split(allowedOrigins, ",")
				for _, allowed := range origins {
					allowed = strings.TrimSpace(allowed)
					allowedURL, err := url.Parse(allowed)
					if err != nil {
						continue
					}
					allowedHost := allowedURL.Host
					originHost := originURL.Host
					if strings.Contains(allowedHost, ":") {
						allowedHost = strings.Split(allowedHost, ":")[0]
					}
					if strings.Contains(originHost, ":") {
						originHost = strings.Split(originHost, ":")[0]
					}
					schemeMatch := (originURL.Scheme == allowedURL.Scheme) ||
						(originURL.Scheme == "https" && allowedURL.Scheme == "wss") ||
						(originURL.Scheme == "wss" && allowedURL.Scheme == "https") ||
						(originURL.Scheme == "http" && allowedURL.Scheme == "ws") ||
						(originURL.Scheme == "ws" && allowedURL.Scheme == "http")
					if schemeMatch && originHost == allowedHost {
						return true
					}
				}
				return false
			}

			// If no ALLOWED_ORIGINS, allow same-origin, localhost with valid scheme
			host := r.Host
			if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
				host = forwardedHost
			}
			if strings.Contains(host, ":") {
				host = strings.Split(host, ":")[0]
			}

			originHost := originURL.Host
			if strings.Contains(originHost, ":") {
				originHost = strings.Split(originHost, ":")[0]
			}

			validScheme := originURL.Scheme == "http" || originURL.Scheme == "https" ||
				originURL.Scheme == "ws" || originURL.Scheme == "wss"

			hostMatch := originHost == host ||
				originHost == "localhost" ||
				originHost == "127.0.0.1" ||
				strings.HasSuffix(originHost, "."+host) ||
				strings.HasSuffix(host, "."+originHost)

			return validScheme && hostMatch
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.LogError(err, "WebSocket upgrade failed", map[string]interface{}{
			"origin": r.Header.Get("Origin"),
			"host":   r.Host,
			"pod":    fmt.Sprintf("%s/%s", params.Namespace, params.PodName),
		})
		return
	}
	defer conn.Close()

	utils.LogInfo("WebSocket upgraded successfully", map[string]interface{}{
		"pod":       fmt.Sprintf("%s/%s", params.Namespace, params.PodName),
		"container": params.Container,
	})

	// Create pipes for stdin/stdout/stderr
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	// Handle WebSocket messages (send to pod stdin) - HTTP layer
	go func() {
		defer stdinWriter.Close()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			stdinWriter.Write(message)
		}
	}()

	// Read from pod stdout and send to WebSocket - HTTP layer
	go func() {
		defer stdoutReader.Close()
		buf := make([]byte, 8192)
		for {
			n, err := stdoutReader.Read(buf)
			if n > 0 {
				if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// Execute the command with a cancelable context (no fixed timeout to avoid dropping long sessions)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  stdinReader,
		Stdout: stdoutWriter,
		Stderr: stdoutWriter,
		Tty:    true,
	})

	if err != nil {
		utils.LogError(err, "Exec stream error", map[string]interface{}{
			"pod":       fmt.Sprintf("%s/%s", params.Namespace, params.PodName),
			"container": params.Container,
		})
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Exec error: %v", err)))
	}
}
