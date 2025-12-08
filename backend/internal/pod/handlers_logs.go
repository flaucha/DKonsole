package pod

import (
	"fmt"
	"net/http"

	"github.com/flaucha/DKonsole/backend/internal/permissions"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

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
	logRepo := s.logRepoFactory(client)

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
