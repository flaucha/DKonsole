package pod

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/flaucha/DKonsole/backend/internal/middleware"
	"github.com/flaucha/DKonsole/backend/internal/permissions"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// ExecIntoPod provides WebSocket-based terminal access to a pod
// Refactored to use layered architecture:
// Handler (HTTP/WebSocket) -> Service (Business Logic) -> Repository (Data Access)
// ExecIntoPod provides WebSocket-based terminal access to a pod
// Refactored to use helper functions for better maintainability and testing
func (s *Service) ExecIntoPod(w http.ResponseWriter, r *http.Request) {
	// 1. Validation & Setup
	params, err := utils.ParsePodParams(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.validateExecPermissions(r.Context(), params.Namespace); err != nil {
		utils.ErrorResponse(w, http.StatusForbidden, err.Error())
		return
	}

	client, restConfig, err := s.getExecClients(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 2. Business Logic (Executor creation)
	execService := s.execFactory()
	execReq := ExecRequest{
		Namespace: params.Namespace,
		PodName:   params.PodName,
		Container: params.Container,
	}
	executor, _, err := execService.CreateExecutor(client, restConfig, execReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create executor: %v", err))
		return
	}

	// 3. Transport Layer (WebSocket)
	conn, err := s.upgradeToWebSocket(w, r, *params)
	if err != nil {
		// upgradeToWebSocket handles logging
		return
	}
	defer conn.Close()

	// 4. Streaming (IO)
	s.streamPodConnection(r.Context(), conn, executor, *params)
}

func (s *Service) validateExecPermissions(ctx context.Context, namespace string) error {
	canEdit, err := permissions.CanPerformAction(ctx, namespace, "edit")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %v", err)
	}
	if !canEdit {
		return fmt.Errorf("edit permission required for namespace: %s", namespace)
	}
	return nil
}

func (s *Service) getExecClients(r *http.Request) (kubernetes.Interface, *rest.Config, error) {
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		return nil, nil, err
	}
	restConfig, err := s.clusterService.GetRESTConfig(r)
	if err != nil {
		return nil, nil, err
	}
	return client, restConfig, nil
}

func (s *Service) upgradeToWebSocket(w http.ResponseWriter, r *http.Request, params utils.PodParams) (*websocket.Conn, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkWebSocketOrigin,
	}

	setWebSocketCORSHeaders(w, r)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.LogError(err, "WebSocket upgrade failed", map[string]interface{}{
			"origin": r.Header.Get("Origin"),
			"host":   r.Host,
			"pod":    fmt.Sprintf("%s/%s", params.Namespace, params.PodName),
		})
		return nil, err
	}

	utils.LogInfo("WebSocket upgraded successfully", map[string]interface{}{
		"pod":       fmt.Sprintf("%s/%s", params.Namespace, params.PodName),
		"container": params.Container,
	})

	return conn, nil
}

func checkWebSocketOrigin(r *http.Request) bool {
	return middleware.IsRequestOriginAllowed(r)
}

func (s *Service) streamPodConnection(ctx context.Context, conn *websocket.Conn, executor remotecommand.Executor, params utils.PodParams) {
	// Configure WebSocket connection
	const (
		pongWait   = 60 * time.Second
		pingPeriod = 30 * time.Second
		writeWait  = 10 * time.Second
	)

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	var writeMu sync.Mutex
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start ping loop
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
					cancel()
					return
				}
				writeMu.Unlock()
			case <-streamCtx.Done():
				return
			}
		}
	}()

	// Read from WebSocket -> Pod Stdin
	go func() {
		defer stdinWriter.Close()
		defer cancel()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if len(message) == 0 || (len(message) == 1 && message[0] == 0) {
				continue
			}
			stdinWriter.Write(message)
		}
	}()

	// Read from Pod Stdout -> WebSocket
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

	err := executor.StreamWithContext(streamCtx, remotecommand.StreamOptions{
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
