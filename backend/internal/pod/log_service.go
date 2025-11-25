package pod

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
)

// LogService provides business logic for pod log operations
type LogService struct {
	logRepo LogRepository
}

// NewLogService creates a new LogService
func NewLogService(logRepo LogRepository) *LogService {
	return &LogService{logRepo: logRepo}
}

// StreamLogsRequest represents the parameters for streaming pod logs
type StreamLogsRequest struct {
	Namespace string
	PodName   string
	Container string
	Follow    bool
}

// StreamLogs opens a log stream for a specific pod
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



