package prometheus

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

// mockRepository is a mock implementation of Repository
type mockPrometheusRepository struct {
	queryRangeFunc   func(ctx context.Context, query string, start, end time.Time, step string) ([]models.MetricDataPoint, error)
	queryInstantFunc func(ctx context.Context, query string) ([]map[string]interface{}, error)
}

func (m *mockPrometheusRepository) QueryRange(ctx context.Context, query string, start, end time.Time, step string) ([]models.MetricDataPoint, error) {
	if m.queryRangeFunc != nil {
		return m.queryRangeFunc(ctx, query, start, end, step)
	}
	return []models.MetricDataPoint{}, nil
}

func (m *mockPrometheusRepository) QueryInstant(ctx context.Context, query string) ([]map[string]interface{}, error) {
	if m.queryInstantFunc != nil {
		return m.queryInstantFunc(ctx, query)
	}
	return []map[string]interface{}{}, nil
}

func TestIsControlPlaneNode(t *testing.T) {
	tests := []struct {
		name string
		node corev1.Node
		want bool
	}{
		{
			name: "node with control-plane label",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "control-plane-node",
					Labels: map[string]string{
						"node-role.kubernetes.io/control-plane": "true",
					},
				},
			},
			want: true,
		},
		{
			name: "node with master label",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "master-node",
					Labels: map[string]string{
						"node-role.kubernetes.io/master": "true",
					},
				},
			},
			want: true,
		},
		{
			name: "worker node without control-plane labels",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "worker-node",
					Labels: map[string]string{
						"kubernetes.io/os": "linux",
					},
				},
			},
			want: false,
		},
		{
			name: "node with control-plane taint",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "control-plane-node",
				},
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key: "node-role.kubernetes.io/control-plane",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "node with master taint",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "master-node",
				},
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key: "node-role.kubernetes.io/master",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "worker node with other taints",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "worker-node",
				},
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key: "node.kubernetes.io/not-ready",
						},
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isControlPlaneNode(tt.node)
			if result != tt.want {
				t.Errorf("isControlPlaneNode() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestService_GetDeploymentMetrics(t *testing.T) {
	tests := []struct {
		name           string
		request        GetDeploymentMetricsRequest
		queryRangeFunc func(ctx context.Context, query string, start, end time.Time, step string) ([]models.MetricDataPoint, error)
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful get deployment metrics",
			request: GetDeploymentMetricsRequest{
				Deployment: "my-app",
				Namespace:  "default",
				Range:      "1h",
			},
			queryRangeFunc: func(ctx context.Context, query string, start, end time.Time, step string) ([]models.MetricDataPoint, error) {
				return []models.MetricDataPoint{
					{Timestamp: time.Now().Unix() * 1000, Value: 100.0},
				}, nil
			},
			wantErr: false,
		},
		{
			name: "invalid namespace parameter",
			request: GetDeploymentMetricsRequest{
				Deployment: "my-app",
				Namespace:  "invalid namespace", // Contains space
				Range:      "1h",
			},
			wantErr: true,
			errMsg:  "invalid namespace",
		},
		{
			name: "invalid deployment parameter",
			request: GetDeploymentMetricsRequest{
				Deployment: "my@app", // Invalid character
				Namespace:  "default",
				Range:      "1h",
			},
			wantErr: true,
			errMsg:  "invalid deployment",
		},
		{
			name: "CPU query error",
			request: GetDeploymentMetricsRequest{
				Deployment: "my-app",
				Namespace:  "default",
				Range:      "1h",
			},
			queryRangeFunc: func(ctx context.Context, query string, start, end time.Time, step string) ([]models.MetricDataPoint, error) {
				if strings.Contains(query, "cpu") {
					return nil, errors.New("CPU query failed")
				}
				return []models.MetricDataPoint{}, nil
			},
			wantErr: true,
			errMsg:  "failed to query CPU metrics",
		},
		{
			name: "Memory query error",
			request: GetDeploymentMetricsRequest{
				Deployment: "my-app",
				Namespace:  "default",
				Range:      "1h",
			},
			queryRangeFunc: func(ctx context.Context, query string, start, end time.Time, step string) ([]models.MetricDataPoint, error) {
				if strings.Contains(query, "memory") {
					return nil, errors.New("Memory query failed")
				}
				return []models.MetricDataPoint{
					{Timestamp: time.Now().Unix() * 1000, Value: 100.0},
				}, nil
			},
			wantErr: true,
			errMsg:  "failed to query memory metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockPrometheusRepository{
				queryRangeFunc: tt.queryRangeFunc,
			}

			service := NewService(mockRepo)
			ctx := context.Background()

			result, err := service.GetDeploymentMetrics(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeploymentMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetDeploymentMetrics() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("GetDeploymentMetrics() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			if result == nil {
				t.Errorf("GetDeploymentMetrics() result is nil")
				return
			}

			// Verify structure
			if result.CPU == nil {
				t.Errorf("GetDeploymentMetrics() CPU should not be nil")
			}
			if result.Memory == nil {
				t.Errorf("GetDeploymentMetrics() Memory should not be nil")
			}
		})
	}
}

func TestService_GetPodMetrics(t *testing.T) {
	tests := []struct {
		name           string
		request        GetPodMetricsRequest
		queryRangeFunc func(ctx context.Context, query string, start, end time.Time, step string) ([]models.MetricDataPoint, error)
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful get pod metrics",
			request: GetPodMetricsRequest{
				PodName:   "my-pod",
				Namespace: "default",
				Range:     "1h",
			},
			queryRangeFunc: func(ctx context.Context, query string, start, end time.Time, step string) ([]models.MetricDataPoint, error) {
				return []models.MetricDataPoint{
					{Timestamp: time.Now().Unix() * 1000, Value: 50.0},
				}, nil
			},
			wantErr: false,
		},
		{
			name: "invalid namespace parameter",
			request: GetPodMetricsRequest{
				PodName:   "my-pod",
				Namespace: "invalid namespace",
				Range:     "1h",
			},
			wantErr: true,
			errMsg:  "invalid namespace",
		},
		{
			name: "invalid pod name parameter",
			request: GetPodMetricsRequest{
				PodName:   "my@pod",
				Namespace: "default",
				Range:     "1h",
			},
			wantErr: true,
			errMsg:  "invalid pod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockPrometheusRepository{
				queryRangeFunc: tt.queryRangeFunc,
			}

			service := NewService(mockRepo)
			ctx := context.Background()

			result, err := service.GetPodMetrics(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPodMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetPodMetrics() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("GetPodMetrics() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			if result == nil {
				t.Errorf("GetPodMetrics() result is nil")
				return
			}

			// Verify structure
			if result.CPU == nil {
				t.Errorf("GetPodMetrics() CPU should not be nil")
			}
			if result.Memory == nil {
				t.Errorf("GetPodMetrics() Memory should not be nil")
			}
		})
	}
}

func TestService_CalculateClusterStats(t *testing.T) {
	service := NewService(nil)

	tests := []struct {
		name                  string
		nodes                 []models.NodeMetric
		controlPlaneCount     int
		controlPlaneNodes     map[string]bool
		wantTotalNodes        int
		wantControlPlaneNodes int
		wantAvgCPU            float64
		wantAvgMemory         float64
	}{
		{
			name: "single worker node",
			nodes: []models.NodeMetric{
				{
					Name:      "worker-1",
					Role:      "worker",
					CPUUsage:  50.0,
					MemUsage:  60.0,
					NetworkRx: 1000.0,
					NetworkTx: 500.0,
				},
			},
			controlPlaneCount:     0,
			controlPlaneNodes:     map[string]bool{},
			wantTotalNodes:        1,
			wantControlPlaneNodes: 0,
			wantAvgCPU:            50.0,
			wantAvgMemory:         60.0,
		},
		{
			name: "multiple worker nodes",
			nodes: []models.NodeMetric{
				{
					Name:      "worker-1",
					Role:      "worker",
					CPUUsage:  50.0,
					MemUsage:  60.0,
					NetworkRx: 1000.0,
					NetworkTx: 500.0,
				},
				{
					Name:      "worker-2",
					Role:      "worker",
					CPUUsage:  30.0,
					MemUsage:  40.0,
					NetworkRx: 2000.0,
					NetworkTx: 1000.0,
				},
			},
			controlPlaneCount:     0,
			controlPlaneNodes:     map[string]bool{},
			wantTotalNodes:        2,
			wantControlPlaneNodes: 0,
			wantAvgCPU:            40.0, // (50 + 30) / 2
			wantAvgMemory:         50.0, // (60 + 40) / 2
		},
		{
			name: "worker nodes with control plane excluded from stats",
			nodes: []models.NodeMetric{
				{
					Name:      "worker-1",
					Role:      "worker",
					CPUUsage:  50.0,
					MemUsage:  60.0,
					NetworkRx: 1000.0,
					NetworkTx: 500.0,
				},
				{
					Name:      "control-plane-1",
					Role:      "control-plane",
					CPUUsage:  80.0,
					MemUsage:  90.0,
					NetworkRx: 5000.0,
					NetworkTx: 3000.0,
				},
			},
			controlPlaneCount:     1,
			controlPlaneNodes:     map[string]bool{"control-plane-1": true},
			wantTotalNodes:        1, // Only worker nodes count
			wantControlPlaneNodes: 1,
			wantAvgCPU:            50.0, // Only worker node CPU
			wantAvgMemory:         60.0, // Only worker node memory
		},
		{
			name:                  "empty nodes list",
			nodes:                 []models.NodeMetric{},
			controlPlaneCount:     0,
			controlPlaneNodes:     map[string]bool{},
			wantTotalNodes:        0,
			wantControlPlaneNodes: 0,
			wantAvgCPU:            0.0,
			wantAvgMemory:         0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := service.calculateClusterStats(tt.nodes, tt.controlPlaneCount, tt.controlPlaneNodes)

			if stats == nil {
				t.Errorf("calculateClusterStats() returned nil")
				return
			}

			if stats.TotalNodes != tt.wantTotalNodes {
				t.Errorf("calculateClusterStats() TotalNodes = %v, want %v", stats.TotalNodes, tt.wantTotalNodes)
			}

			if stats.ControlPlaneNodes != tt.wantControlPlaneNodes {
				t.Errorf("calculateClusterStats() ControlPlaneNodes = %v, want %v", stats.ControlPlaneNodes, tt.wantControlPlaneNodes)
			}

			if stats.AvgCPUUsage != tt.wantAvgCPU {
				t.Errorf("calculateClusterStats() AvgCPUUsage = %v, want %v", stats.AvgCPUUsage, tt.wantAvgCPU)
			}

			if stats.AvgMemoryUsage != tt.wantAvgMemory {
				t.Errorf("calculateClusterStats() AvgMemoryUsage = %v, want %v", stats.AvgMemoryUsage, tt.wantAvgMemory)
			}
		})
	}
}
