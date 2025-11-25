package pod

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
)

// LogService provides business logic for pod log operations.
type LogService struct {
	logRepo LogRepository
}

// NewLogService creates a new LogService with the provided log repository.
func NewLogService(logRepo LogRepository) *LogService {
	return &LogService{logRepo: logRepo}
}

// StreamLogsRequest represents the parameters for streaming pod logs.
type StreamLogsRequest struct {
	Namespace string // Kubernetes namespace
	PodName   string // Pod name
	Container string // Optional container name (for multi-container pods)
	Follow    bool   // If true, stream logs continuously as they are generated
}

// StreamLogs opens a log stream for a specific pod container.
// Returns an io.ReadCloser that can be used to read log data.
// The caller is responsible for closing the stream.
// Returns an error if the log stream cannot be opened.
func (s *LogService) StreamLogs(ctx context.Context, req StreamLogsRequest) (io.ReadCloser, error) {
	opts := &corev1.PodLogOptions{
		Follow: req.Follow,
	}
	if req.Container != "" {
		opts.Container = req.Container
	}

	stream, err := s.logRepo.GetLogStream(ctx, req.Namespace, req.PodName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get log stream from repository: %w", err)
	}

	return stream, nil
}
