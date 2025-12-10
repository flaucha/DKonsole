package prometheus

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

// MockRepo mocks the repository
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) QueryInstant(ctx context.Context, query string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockRepo) QueryRange(ctx context.Context, query string, start, end time.Time, step string) ([]models.MetricDataPoint, error) {
	args := m.Called(ctx, query, start, end, step)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.MetricDataPoint), args.Error(1)
}

func TestGetNodeMetrics(t *testing.T) {
	// Setup
	mockRepo := new(MockRepo)
	service := &Service{repo: mockRepo}

	// Create fake k8s client with a node
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-1",
			Labels: map[string]string{
				"node-role.kubernetes.io/worker": "true",
			},
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
			},
			Addresses: []corev1.NodeAddress{
				{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
			},
		},
	}
	k8sClient := testclient.NewSimpleClientset(node)

	// Mock Prometheus responses
	// We expect CPU, Mem, Disk, NetRx, NetTx queries

	// Mock fetchMetricMap responses
	// Note: We need to match the specific queries used in node_metrics.go

	// Use Anything for query string to simplify mock setup for now,
	// or be specific if we want to valid queries.
	mockRepo.On("QueryInstant", mock.Anything, mock.AnythingOfType("string")).Return(
		[]map[string]interface{}{
			{
				"instance": "10.0.0.1:9100",
				"value":    float64(50.0),
			},
		}, nil,
	)

	// Mock node info query for buildNodeToInstanceMap
	/*
		nodesQuery := `kube_node_info`
		mockRepo.On("QueryInstant", mock.Anything, nodesQuery).Return(
			[]map[string]interface{}{
				{
					"node":     "node-1",
					"instance": "10.0.0.1:9100",
				},
			}, nil,
		)
	*/

	ctx := context.Background()
	metrics, count, _, err := service.getNodeMetrics(ctx, k8sClient)

	assert.NoError(t, err)
	assert.Equal(t, 0, count) // 0 control plane nodes
	assert.Len(t, metrics, 1)
	assert.Equal(t, "node-1", metrics[0].Name)
	assert.Equal(t, "Ready", metrics[0].Status)
	// Since we returned 50.0 for all queries (CPU, Mem, etc.), check if it mapped correctly
	// Note: The mapping logic relies on instance usage.
	// Our mock returns instance "10.0.0.1:9100".
	// The node internal IP is "10.0.0.1".
	// The resolveNodeInstance logic should match "10.0.0.1:9100" to "node-1" via IP prefix.
	assert.Equal(t, 50.0, metrics[0].CPUUsage)
}
