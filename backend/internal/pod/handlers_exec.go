package pod

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/flaucha/DKonsole/backend/internal/permissions"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

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
	execService := s.execFactory()

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

	setWebSocketCORSHeaders(w, r)

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

	// Configure WebSocket connection with ping/pong for keep-alive
	const (
		pongWait   = 60 * time.Second // Time allowed to read the next pong message
		pingPeriod = 30 * time.Second // Send pings with this period (must be less than pongWait)
		writeWait  = 10 * time.Second // Time allowed to write a message
	)

	// Set initial read deadline
	conn.SetReadDeadline(time.Now().Add(pongWait))

	// Handle pong messages - extend the read deadline when we receive a pong
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Mutex to protect concurrent writes to WebSocket
	var writeMu sync.Mutex

	// Create pipes for stdin/stdout/stderr
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	// Context for cleanup coordination
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Ping ticker to keep connection alive
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				writeMu.Lock()
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					writeMu.Unlock()
					cancel() // Signal to close the connection
					return
				}
				writeMu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Handle WebSocket messages (send to pod stdin) - HTTP layer
	go func() {
		defer stdinWriter.Close()
		defer cancel()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			// Filter out null characters used for keep-alive (legacy clients)
			// Only forward non-empty, non-null messages to the shell
			if len(message) == 0 {
				continue
			}
			if len(message) == 1 && message[0] == 0 {
				continue
			}
			stdinWriter.Write(message)
		}
	}()

	// Read from pod stdout and send to WebSocket - HTTP layer
	go func() {
		defer stdoutReader.Close()
		defer cancel()
		buf := make([]byte, 8192)
		for {
			n, err := stdoutReader.Read(buf)
			if n > 0 {
				writeMu.Lock()
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				writeErr := conn.WriteMessage(websocket.BinaryMessage, buf[:n])
				writeMu.Unlock()
				if writeErr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// Execute the command (uses the previously created cancelable context)
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
		writeMu.Lock()
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Exec error: %v", err)))
		writeMu.Unlock()
	}
}

// setWebSocketCORSHeaders writes the standard CORS headers for the WebSocket handshake.
// ExecIntoPod already validates the origin via the Upgrader, so this helper simply mirrors
// the headers enableCors would emit to satisfy the browser’s credentialed requests.
func setWebSocketCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "3600")
}
