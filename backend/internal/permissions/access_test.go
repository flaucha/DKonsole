package permissions

import (
	"context"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func TestHasNamespaceAccess(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		namespace string
		want      bool
		wantErr   bool
	}{
		{
			name: "admin has access to any namespace",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "admin",
					Role:        "admin",
					Permissions: nil,
				},
			),
			namespace: "any-namespace",
			want:      true,
			wantErr:   false,
		},
		{
			name: "user with view permission has access",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"namespace1": "view",
					},
				},
			),
			namespace: "namespace1",
			want:      true,
			wantErr:   false,
		},
		{
			name: "user with edit permission has access",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"namespace1": "edit",
					},
				},
			),
			namespace: "namespace1",
			want:      true,
			wantErr:   false,
		},
		{
			name: "user without permission has no access",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"namespace1": "view",
					},
				},
			),
			namespace: "namespace2",
			want:      false,
			wantErr:   false,
		},
		{
			name: "empty permissions means admin access",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "admin",
					Role:        "user",
					Permissions: map[string]string{},
				},
			),
			namespace: "any-namespace",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "no user in context",
			ctx:       context.Background(),
			namespace: "namespace1",
			want:      false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HasNamespaceAccess(tt.ctx, tt.namespace)

			if (err != nil) != tt.wantErr {
				t.Errorf("HasNamespaceAccess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("HasNamespaceAccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPermissionLevel(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		namespace string
		want      string
		wantErr   bool
	}{
		{
			name: "admin returns edit",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "admin",
					Role:        "admin",
					Permissions: nil,
				},
			),
			namespace: "any-namespace",
			want:      "edit",
			wantErr:   false,
		},
		{
			name: "user with view permission returns view",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"namespace1": "view",
					},
				},
			),
			namespace: "namespace1",
			want:      "view",
			wantErr:   false,
		},
		{
			name: "user with edit permission returns edit",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"namespace1": "edit",
					},
				},
			),
			namespace: "namespace1",
			want:      "edit",
			wantErr:   false,
		},
		{
			name: "user without access returns error",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"namespace1": "view",
					},
				},
			),
			namespace: "namespace2",
			want:      "",
			wantErr:   true,
		},
		{
			name: "empty permissions means edit (admin)",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "admin",
					Role:        "user",
					Permissions: map[string]string{},
				},
			),
			namespace: "any-namespace",
			want:      "edit",
			wantErr:   false,
		},
		{
			name:      "no user in context",
			ctx:       context.Background(),
			namespace: "namespace1",
			want:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPermissionLevel(tt.ctx, tt.namespace)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPermissionLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("GetPermissionLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanPerformAction(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		namespace string
		action    string
		want      bool
		wantErr   bool
	}{
		{
			name: "admin can perform view action",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "admin",
					Role:        "admin",
					Permissions: nil,
				},
			),
			namespace: "namespace1",
			action:    "view",
			want:      true,
			wantErr:   false,
		},
		{
			name: "admin can perform edit action",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "admin",
					Role:        "admin",
					Permissions: nil,
				},
			),
			namespace: "namespace1",
			action:    "edit",
			want:      true,
			wantErr:   false,
		},
		{
			name: "user with view permission can perform view action",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"namespace1": "view",
					},
				},
			),
			namespace: "namespace1",
			action:    "view",
			want:      true,
			wantErr:   false,
		},
		{
			name: "user with view permission cannot perform edit action",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"namespace1": "view",
					},
				},
			),
			namespace: "namespace1",
			action:    "edit",
			want:      false,
			wantErr:   false,
		},
		{
			name: "user with edit permission can perform edit action",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"namespace1": "edit",
					},
				},
			),
			namespace: "namespace1",
			action:    "edit",
			want:      true,
			wantErr:   false,
		},
		{
			name: "user with edit permission can perform view action",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"namespace1": "edit",
					},
				},
			),
			namespace: "namespace1",
			action:    "view",
			want:      true,
			wantErr:   false,
		},
		{
			name: "invalid action returns error",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "testuser",
					Role:     "user",
					Permissions: map[string]string{
						"namespace1": "view",
					},
				},
			),
			namespace: "namespace1",
			action:    "invalid",
			want:      false,
			wantErr:   true,
		},
		{
			name:      "no user in context",
			ctx:       context.Background(),
			namespace: "namespace1",
			action:    "view",
			want:      false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CanPerformAction(tt.ctx, tt.namespace, tt.action)

			if (err != nil) != tt.wantErr {
				t.Errorf("CanPerformAction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("CanPerformAction() = %v, want %v", got, tt.want)
			}
		})
	}
}
