package pod

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/example/k8s-view/internal/cluster"
	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/utils"
)

// Service provides pod-specific operations
type Service struct {
	handlers       *models.Handlers
	clusterService *cluster.Service
}

// NewService creates a new pod service
func NewService(h *models.Handlers, cs *cluster.Service) *Service {
	return &Service{
		handlers:       h,
		clusterService: cs,
	}
}

// StreamPodLogs streams logs from a pod
func (s *Service) StreamPodLogs(w http.ResponseWriter, r *http.Request) {
	// Note: authenticateRequest is handled by middleware

	client, err := s.clusterService.GetClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ns := r.URL.Query().Get("namespace")
	podName := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")

	if ns == "" || podName == "" {
		http.Error(w, "Missing namespace or pod parameter", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateK8sName(ns, "namespace"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.ValidateK8sName(podName, "pod"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if container != "" {
		if err := utils.ValidateK8sName(container, "container"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	opts := &corev1.PodLogOptions{
		Follow: true,
	}
	if container != "" {
		opts.Container = container
	}

	req := client.CoreV1().Pods(ns).GetLogs(podName, opts)
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()
	stream, err := req.Stream(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open log stream: %v", err), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

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

// EventInfo represents pod event information
type EventInfo struct {
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Count     int32     `json:"count"`
	FirstSeen time.Time `json:"firstSeen"`
	LastSeen  time.Time `json:"lastSeen"`
	Source    string    `json:"source,omitempty"`
}

// GetPodEvents returns events for a specific pod
func (s *Service) GetPodEvents(w http.ResponseWriter, r *http.Request) {
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ns := r.URL.Query().Get("namespace")
	podName := r.URL.Query().Get("pod")

	if ns == "" || podName == "" {
		http.Error(w, "Missing namespace or pod parameter", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateK8sName(ns, "namespace"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.ValidateK8sName(podName, "pod"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Get events for the pod
	events, err := client.CoreV1().Events(ns).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podName),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get pod events: %v", err), http.StatusInternalServerError)
		return
	}

	var eventList []EventInfo
	for _, event := range events.Items {
		eventList = append(eventList, EventInfo{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Count:     event.Count,
			FirstSeen: event.FirstTimestamp.Time,
			LastSeen:  event.LastTimestamp.Time,
			Source:    fmt.Sprintf("%s/%s", event.Source.Component, event.Source.Host),
		})
	}

	// Sort by last seen (most recent first)
	sort.Slice(eventList, func(i, j int) bool {
		return eventList[i].LastSeen.After(eventList[j].LastSeen)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eventList)
}

// ExecIntoPod provides WebSocket-based terminal access to a pod
func (s *Service) ExecIntoPod(w http.ResponseWriter, r *http.Request) {
	// Authenticate before WebSocket upgrade
	// Note: authenticateRequest is set up in main.go
	// For now, we'll handle auth through middleware
	// This will be updated when auth is fully migrated

	client, err := s.clusterService.GetClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ns := r.URL.Query().Get("namespace")
	podName := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")

	if ns == "" || podName == "" {
		http.Error(w, "Missing namespace or pod parameter", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateK8sName(ns, "namespace"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := utils.ValidateK8sName(podName, "pod"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if container != "" {
		if err := utils.ValidateK8sName(container, "container"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Upgrade HTTP connection to WebSocket
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
		log.Printf("WebSocket upgrade failed: %v, Origin: %s, Host: %s",
			err, r.Header.Get("Origin"), r.Host)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket upgraded successfully for pod: %s/%s", ns, podName)

	// Create exec request
	req := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(ns).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   []string{"/bin/sh", "-c", "TERM=xterm-256color; export TERM; [ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c \"/bin/bash\" /dev/null || exec /bin/bash) || exec /bin/sh"},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, runtime.NewParameterCodec(clientgoscheme.Scheme))

	// Get REST config for exec
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}
	s.handlers.RLock()
	restConfig := s.handlers.RESTConfigs[cluster]
	s.handlers.RUnlock()

	if restConfig == nil {
		conn.WriteMessage(websocket.TextMessage, []byte("REST config not found for cluster"))
		return
	}

	exec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", req.URL())
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to create executor: %v", err)))
		return
	}

	// Create pipes for stdin/stdout/stderr
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	// Handle WebSocket messages (send to pod stdin)
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

	// Read from pod stdout and send to WebSocket
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

	// Execute the command
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdinReader,
		Stdout: stdoutWriter,
		Stderr: stdoutWriter,
		Tty:    true,
	})

	if err != nil {
		log.Printf("Exec stream error: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Exec error: %v", err)))
	}
}
