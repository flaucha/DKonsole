package k8s

import (
	"context"
	"errors"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

// mockClusterStatsRepository is a mock implementation of ClusterStatsRepository
type mockClusterStatsRepository struct {
	getNodeCountFunc        func(ctx context.Context) (int, error)
	getNamespaceCountFunc   func(ctx context.Context) (int, error)
	getPodCountFunc         func(ctx context.Context) (int, error)
	getDeploymentCountFunc  func(ctx context.Context) (int, error)
	getServiceCountFunc     func(ctx context.Context) (int, error)
	getIngressCountFunc     func(ctx context.Context) (int, error)
	getPVCCountFunc         func(ctx context.Context) (int, error)
	getPVCountFunc          func(ctx context.Context) (int, error)
}

func (m *mockClusterStatsRepository) GetNodeCount(ctx context.Context) (int, error) {
	if m.getNodeCountFunc != nil {
		return m.getNodeCountFunc(ctx)
	}
	return 3, nil
}

func (m *mockClusterStatsRepository) GetNamespaceCount(ctx context.Context) (int, error) {
	if m.getNamespaceCountFunc != nil {
		return m.getNamespaceCountFunc(ctx)
	}
	return 5, nil
}

func (m *mockClusterStatsRepository) GetPodCount(ctx context.Context) (int, error) {
	if m.getPodCountFunc != nil {
		return m.getPodCountFunc(ctx)
	}
	return 10, nil
}

func (m *mockClusterStatsRepository) GetDeploymentCount(ctx context.Context) (int, error) {
	if m.getDeploymentCountFunc != nil {
		return m.getDeploymentCountFunc(ctx)
	}
	return 7, nil
}

func (m *mockClusterStatsRepository) GetServiceCount(ctx context.Context) (int, error) {
	if m.getServiceCountFunc != nil {
		return m.getServiceCountFunc(ctx)
	}
	return 8, nil
}

func (m *mockClusterStatsRepository) GetIngressCount(ctx context.Context) (int, error) {
	if m.getIngressCountFunc != nil {
		return m.getIngressCountFunc(ctx)
	}
	return 2, nil
}

func (m *mockClusterStatsRepository) GetPVCCount(ctx context.Context) (int, error) {
	if m.getPVCCountFunc != nil {
		return m.getPVCCountFunc(ctx)
	}
	return 4, nil
}

func (m *mockClusterStatsRepository) GetPVCount(ctx context.Context) (int, error) {
	if m.getPVCountFunc != nil {
		return m.getPVCountFunc(ctx)
	}
	return 6, nil
}

func TestClusterStatsService_GetClusterStats(t *testing.T) {
	tests := []struct {
		name        string
		repo        *mockClusterStatsRepository
		wantErr     bool
		wantStats   models.ClusterStats
		expectError bool
	}{
		{
			name: "successful stats retrieval",
			repo: &mockClusterStatsRepository{},
			wantErr: false,
			wantStats: models.ClusterStats{
				Nodes:       3,
				Namespaces:  5,
				Pods:        10,
				Deployments: 7,
				Services:    8,
				Ingresses:   2,
				PVCs:        4,
				PVs:         6,
			},
			expectError: false,
		},
		{
			name: "partial errors - some stats succeed",
			repo: &mockClusterStatsRepository{
				getNodeCountFunc: func(ctx context.Context) (int, error) {
					return 3, nil
				},
				getNamespaceCountFunc: func(ctx context.Context) (int, error) {
					return 5, nil
				},
				getPodCountFunc: func(ctx context.Context) (int, error) {
					return 0, errors.New("failed to get pod count")
				},
				getDeploymentCountFunc: func(ctx context.Context) (int, error) {
					return 7, nil
				},
				getServiceCountFunc: func(ctx context.Context) (int, error) {
					return 0, errors.New("failed to get service count")
				},
				getIngressCountFunc: func(ctx context.Context) (int, error) {
					return 2, nil
				},
				getPVCCountFunc: func(ctx context.Context) (int, error) {
					return 4, nil
				},
				getPVCountFunc: func(ctx context.Context) (int, error) {
					return 6, nil
				},
			},
			wantErr: true, // Should return error but with partial stats
			wantStats: models.ClusterStats{
				Nodes:       3,
				Namespaces:  5,
				Pods:        0, // Error, default to 0
				Deployments: 7,
				Services:    0, // Error, default to 0
				Ingresses:   2,
				PVCs:        4,
				PVs:         6,
			},
			expectError: true,
		},
		{
			name: "all stats succeed",
			repo: &mockClusterStatsRepository{
				getNodeCountFunc: func(ctx context.Context) (int, error) {
					return 1, nil
				},
				getNamespaceCountFunc: func(ctx context.Context) (int, error) {
					return 2, nil
				},
				getPodCountFunc: func(ctx context.Context) (int, error) {
					return 3, nil
				},
				getDeploymentCountFunc: func(ctx context.Context) (int, error) {
					return 4, nil
				},
				getServiceCountFunc: func(ctx context.Context) (int, error) {
					return 5, nil
				},
				getIngressCountFunc: func(ctx context.Context) (int, error) {
					return 6, nil
				},
				getPVCCountFunc: func(ctx context.Context) (int, error) {
					return 7, nil
				},
				getPVCountFunc: func(ctx context.Context) (int, error) {
					return 8, nil
				},
			},
			wantErr: false,
			wantStats: models.ClusterStats{
				Nodes:       1,
				Namespaces:  2,
				Pods:        3,
				Deployments: 4,
				Services:    5,
				Ingresses:   6,
				PVCs:        7,
				PVs:         8,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewClusterStatsService(tt.repo)
			ctx := context.Background()

			stats, err := service.GetClusterStats(ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("GetClusterStats() expected error but got nil")
					return
				}
			} else {
				if err != nil {
					t.Errorf("GetClusterStats() error = %v, want nil", err)
					return
				}
			}

			// Check stats values
			if stats.Nodes != tt.wantStats.Nodes {
				t.Errorf("GetClusterStats() Nodes = %v, want %v", stats.Nodes, tt.wantStats.Nodes)
			}
			if stats.Namespaces != tt.wantStats.Namespaces {
				t.Errorf("GetClusterStats() Namespaces = %v, want %v", stats.Namespaces, tt.wantStats.Namespaces)
			}
			if stats.Pods != tt.wantStats.Pods {
				t.Errorf("GetClusterStats() Pods = %v, want %v", stats.Pods, tt.wantStats.Pods)
			}
			if stats.Deployments != tt.wantStats.Deployments {
				t.Errorf("GetClusterStats() Deployments = %v, want %v", stats.Deployments, tt.wantStats.Deployments)
			}
			if stats.Services != tt.wantStats.Services {
				t.Errorf("GetClusterStats() Services = %v, want %v", stats.Services, tt.wantStats.Services)
			}
			if stats.Ingresses != tt.wantStats.Ingresses {
				t.Errorf("GetClusterStats() Ingresses = %v, want %v", stats.Ingresses, tt.wantStats.Ingresses)
			}
			if stats.PVCs != tt.wantStats.PVCs {
				t.Errorf("GetClusterStats() PVCs = %v, want %v", stats.PVCs, tt.wantStats.PVCs)
			}
			if stats.PVs != tt.wantStats.PVs {
				t.Errorf("GetClusterStats() PVs = %v, want %v", stats.PVs, tt.wantStats.PVs)
			}
		})
	}
}
