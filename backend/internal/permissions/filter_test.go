package permissions

import (
	"context"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func TestFilterAllowedNamespaces(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		namespaces []string
		want       []string
		wantErr    bool
	}{
		{
			name: "admin sees all namespaces",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "admin",
					Role:        "admin",
					Permissions: nil,
				},
			),
			namespaces: []string{"ns1", "ns2", "ns3"},
			want:       []string{"ns1", "ns2", "ns3"},
			wantErr:    false,
		},
		{
			name: "user sees only allowed namespaces",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"ns1": "view",
						"ns3": "edit",
					},
				},
			),
			namespaces: []string{"ns1", "ns2", "ns3"},
			want:       []string{"ns1", "ns3"},
			wantErr:    false,
		},
		{
			name: "user with no permissions sees none",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "testuser",
					Role:        "user",
					Permissions: map[string]string{},
				},
			),
			namespaces: []string{"ns1", "ns2", "ns3"},
			want:       []string{},
			wantErr:    false,
		},
		{
			name: "user with empty permissions sees none",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "user",
					Role:        "user",
					Permissions: map[string]string{},
				},
			),
			namespaces: []string{"ns1", "ns2", "ns3"},
			want:       []string{},
			wantErr:    false,
		},
		{
			name:       "no user in context",
			ctx:        context.Background(),
			namespaces: []string{"ns1", "ns2"},
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilterAllowedNamespaces(tt.ctx, tt.namespaces)

			if (err != nil) != tt.wantErr {
				t.Errorf("FilterAllowedNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("FilterAllowedNamespaces() length = %v, want %v", len(got), len(tt.want))
					return
				}

				// Create maps for comparison
				gotMap := make(map[string]bool)
				for _, ns := range got {
					gotMap[ns] = true
				}
				wantMap := make(map[string]bool)
				for _, ns := range tt.want {
					wantMap[ns] = true
				}

				for ns := range gotMap {
					if !wantMap[ns] {
						t.Errorf("FilterAllowedNamespaces() got unexpected namespace: %v", ns)
					}
				}
				for ns := range wantMap {
					if !gotMap[ns] {
						t.Errorf("FilterAllowedNamespaces() missing namespace: %v", ns)
					}
				}
			}
		})
	}
}

func TestGetAllowedNamespaces(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		want    []string
		wantErr bool
	}{
		{
			name: "admin returns empty list (means all)",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "admin",
					Role:        "admin",
					Permissions: nil,
				},
			),
			want:    []string{},
			wantErr: false,
		},
		{
			name: "user returns list of allowed namespaces",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"ns1": "view",
						"ns2": "edit",
						"ns3": "view",
					},
				},
			),
			want:    []string{"ns1", "ns2", "ns3"},
			wantErr: false,
		},
		{
			name: "user with no permissions returns empty list",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "testuser",
					Role:        "user",
					Permissions: map[string]string{},
				},
			),
			want:    []string{},
			wantErr: false,
		},
		{
			name: "user with empty permissions returns empty list",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "user",
					Role:        "user",
					Permissions: map[string]string{},
				},
			),
			want:    []string{},
			wantErr: false,
		},
		{
			name:    "no user in context",
			ctx:     context.Background(),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAllowedNamespaces(tt.ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllowedNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("GetAllowedNamespaces() length = %v, want %v", len(got), len(tt.want))
					return
				}

				// Create maps for comparison (order doesn't matter)
				gotMap := make(map[string]bool)
				for _, ns := range got {
					gotMap[ns] = true
				}
				wantMap := make(map[string]bool)
				for _, ns := range tt.want {
					wantMap[ns] = true
				}

				for ns := range gotMap {
					if !wantMap[ns] {
						t.Errorf("GetAllowedNamespaces() got unexpected namespace: %v", ns)
					}
				}
				for ns := range wantMap {
					if !gotMap[ns] {
						t.Errorf("GetAllowedNamespaces() missing namespace: %v", ns)
					}
				}
			}
		})
	}
}
func TestFilterResources(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		resources []models.Resource
		want      []models.Resource
		wantErr   bool
	}{
		{
			name: "admin sees all resources",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "admin",
					Role:        "admin",
					Permissions: nil,
				},
			),
			resources: []models.Resource{
				{Name: "res1", Namespace: "ns1", Kind: "Pod"},
				{Name: "res2", Namespace: "ns2", Kind: "Pod"},
			},
			want: []models.Resource{
				{Name: "res1", Namespace: "ns1", Kind: "Pod"},
				{Name: "res2", Namespace: "ns2", Kind: "Pod"},
			},
			wantErr: false,
		},
		{
			name: "user sees only resources in allowed namespaces",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"ns1": "view",
					},
				},
			),
			resources: []models.Resource{
				{Name: "res1", Namespace: "ns1", Kind: "Pod"},
				{Name: "res2", Namespace: "ns2", Kind: "Pod"},
				{Name: "res3", Namespace: "ns1", Kind: "Deployment"},
			},
			want: []models.Resource{
				{Name: "res1", Namespace: "ns1", Kind: "Pod"},
				{Name: "res3", Namespace: "ns1", Kind: "Deployment"},
			},
			wantErr: false,
		},
		{
			name: "cluster-scoped resources are visible to all users",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"ns1": "view",
					},
				},
			),
			resources: []models.Resource{
				{Name: "node1", Namespace: "", Kind: "Node"},
				{Name: "res1", Namespace: "ns1", Kind: "Pod"},
				{Name: "res2", Namespace: "ns2", Kind: "Pod"},
			},
			want: []models.Resource{
				{Name: "node1", Namespace: "", Kind: "Node"},
				{Name: "res1", Namespace: "ns1", Kind: "Pod"},
			},
			wantErr: false,
		},
		{
			name: "user with no permissions sees only cluster-scoped resources",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "testuser",
					Role:        "user",
					Permissions: map[string]string{},
				},
			),
			resources: []models.Resource{
				{Name: "node1", Namespace: "", Kind: "Node"},
				{Name: "res1", Namespace: "ns1", Kind: "Pod"},
			},
			want: []models.Resource{
				{Name: "node1", Namespace: "", Kind: "Node"},
			},
			wantErr: false,
		},
		{
			name:      "no user in context",
			ctx:       context.Background(),
			resources: []models.Resource{{Name: "res1", Namespace: "ns1", Kind: "Pod"}},
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilterResources(tt.ctx, tt.resources)

			if (err != nil) != tt.wantErr {
				t.Errorf("FilterResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("FilterResources() length = %v, want %v", len(got), len(tt.want))
					return
				}

				// Compare resources
				for i := range got {
					if got[i].Name != tt.want[i].Name || got[i].Namespace != tt.want[i].Namespace {
						t.Errorf("FilterResources()[%d] = %+v, want %+v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
