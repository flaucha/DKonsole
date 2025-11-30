package permissions

import (
	"context"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func TestValidateNamespaceAccess(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		namespace string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "admin has access",
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
			wantErr:   false,
		},
		{
			name: "user with permission has access",
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
			wantErr:   false,
		},
		{
			name: "user without permission denied",
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
			wantErr:   true,
			errMsg:    "access denied to namespace",
		},
		{
			name:      "no user in context",
			ctx:       context.Background(),
			namespace: "namespace1",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNamespaceAccess(tt.ctx, tt.namespace)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNamespaceAccess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateNamespaceAccess() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateNamespaceAccess() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateAction(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		namespace string
		action    string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "admin can perform any action",
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
			wantErr:   false,
		},
		{
			name: "user with edit permission can perform edit",
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
			wantErr:   false,
		},
		{
			name: "user with view permission can perform view",
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
			wantErr:   false,
		},
		{
			name: "user with view permission cannot perform edit",
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
			wantErr:   true,
			errMsg:    "not allowed",
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
			wantErr:   true,
		},
		{
			name:      "no user in context",
			ctx:       context.Background(),
			namespace: "namespace1",
			action:    "view",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAction(tt.ctx, tt.namespace, tt.action)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateAction() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateAction() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}
