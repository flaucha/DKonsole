package pod

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

// mockLogRepository is a mock implementation of LogRepository
type mockLogRepository struct {
	getLogStreamFunc func(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error)
}

func (m *mockLogRepository) GetLogStream(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error) {
	if m.getLogStreamFunc != nil {
		return m.getLogStreamFunc(ctx, namespace, podName, opts)
	}
	return nil, errors.New("get log stream not implemented")
}

func TestLogService_StreamLogs(t *testing.T) {
	tests := []struct {
		name              string
		request           StreamLogsRequest
		getLogStreamFunc  func(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error)
		wantErr           bool
		errMsg            string
		expectedFollow    bool
		expectedContainer string
	}{
		{
			name: "successful log stream",
			request: StreamLogsRequest{
				Namespace: "default",
				PodName:   "test-pod",
				Container: "",
				Follow:    false,
			},
			getLogStreamFunc: func(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error) {
				if namespace != "default" {
					t.Errorf("GetLogStream() namespace = %v, want default", namespace)
				}
				if podName != "test-pod" {
					t.Errorf("GetLogStream() podName = %v, want test-pod", podName)
				}
				if opts.Follow != false {
					t.Errorf("GetLogStream() Follow = %v, want false", opts.Follow)
				}
				return io.NopCloser(strings.NewReader("test log content")), nil
			},
			wantErr:        false,
			expectedFollow: false,
		},
		{
			name: "stream logs with follow",
			request: StreamLogsRequest{
				Namespace: "default",
				PodName:   "test-pod",
				Container: "",
				Follow:    true,
			},
			getLogStreamFunc: func(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error) {
				if opts.Follow != true {
					t.Errorf("GetLogStream() Follow = %v, want true", opts.Follow)
				}
				return io.NopCloser(strings.NewReader("streaming logs...")), nil
			},
			wantErr:        false,
			expectedFollow: true,
		},
		{
			name: "stream logs with container specified",
			request: StreamLogsRequest{
				Namespace: "default",
				PodName:   "test-pod",
				Container: "sidecar",
				Follow:    false,
			},
			getLogStreamFunc: func(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error) {
				if opts.Container != "sidecar" {
					t.Errorf("GetLogStream() Container = %v, want sidecar", opts.Container)
				}
				return io.NopCloser(strings.NewReader("sidecar logs")), nil
			},
			wantErr:           false,
			expectedContainer: "sidecar",
		},
		{
			name: "repository error",
			request: StreamLogsRequest{
				Namespace: "default",
				PodName:   "test-pod",
				Container: "",
				Follow:    false,
			},
			getLogStreamFunc: func(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error) {
				return nil, errors.New("pod not found")
			},
			wantErr: true,
			errMsg:  "failed to get log stream from repository",
		},
		{
			name: "stream logs with container and follow",
			request: StreamLogsRequest{
				Namespace: "default",
				PodName:   "test-pod",
				Container: "app",
				Follow:    true,
			},
			getLogStreamFunc: func(ctx context.Context, namespace, podName string, opts *corev1.PodLogOptions) (io.ReadCloser, error) {
				if opts.Container != "app" {
					t.Errorf("GetLogStream() Container = %v, want app", opts.Container)
				}
				if opts.Follow != true {
					t.Errorf("GetLogStream() Follow = %v, want true", opts.Follow)
				}
				return io.NopCloser(strings.NewReader("app logs streaming...")), nil
			},
			wantErr:           false,
			expectedFollow:    true,
			expectedContainer: "app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockLogRepository{
				getLogStreamFunc: tt.getLogStreamFunc,
			}

			service := NewLogService(mockRepo)
			ctx := context.Background()

			stream, err := service.StreamLogs(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("StreamLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("StreamLogs() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("StreamLogs() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			if stream == nil {
				t.Errorf("StreamLogs() stream is nil")
				return
			}

			// Clean up - close the stream
			defer stream.Close()

			// Verify we can read from the stream
			buf := make([]byte, 1024)
			n, err := stream.Read(buf)
			if err != nil && err != io.EOF {
				t.Errorf("StreamLogs() error reading from stream: %v", err)
				return
			}
			if n == 0 {
				t.Errorf("StreamLogs() stream is empty")
			}
		})
	}
}
