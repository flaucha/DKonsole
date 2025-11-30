package permissions

import (
	"context"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)

func TestGetUserFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		wantErr  bool
		wantUser *models.Claims
		errMsg   string
	}{
		{
			name: "valid user from AuthClaims",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&auth.AuthClaims{
					Claims: models.Claims{
						Username:    "testuser",
						Role:        "admin",
						Permissions: nil,
					},
				},
			),
			wantErr: false,
			wantUser: &models.Claims{
				Username:    "testuser",
				Role:        "admin",
				Permissions: nil,
			},
		},
		{
			name: "valid user from models.Claims",
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
			wantErr: false,
			wantUser: &models.Claims{
				Username: "testuser",
				Role:     "user",
				Permissions: map[string]string{
					"namespace1": "view",
				},
			},
		},
		{
			name:    "no user in context",
			ctx:     context.Background(),
			wantErr: true,
			errMsg:  "user not found in context",
		},
		{
			name: "invalid user type in context",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				"invalid-type",
			),
			wantErr: true,
			errMsg:  "invalid user type in context",
		},
		{
			name: "user from map (legacy format)",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				map[string]interface{}{
					"username": "legacyuser",
					"role":     "user",
					"permissions": map[string]interface{}{
						"namespace1": "edit",
					},
				},
			),
			wantErr: false,
			wantUser: &models.Claims{
				Username: "legacyuser",
				Role:     "user",
				Permissions: map[string]string{
					"namespace1": "edit",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserFromContext(tt.ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserFromContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetUserFromContext() expected error but got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg && !contains(err.Error(), tt.errMsg) {
					t.Errorf("GetUserFromContext() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if got == nil {
				t.Errorf("GetUserFromContext() got = nil, want %+v", tt.wantUser)
				return
			}

			if got.Username != tt.wantUser.Username {
				t.Errorf("GetUserFromContext() username = %v, want %v", got.Username, tt.wantUser.Username)
			}
			if got.Role != tt.wantUser.Role {
				t.Errorf("GetUserFromContext() role = %v, want %v", got.Role, tt.wantUser.Role)
			}

			// Check permissions
			if len(got.Permissions) != len(tt.wantUser.Permissions) {
				t.Errorf("GetUserFromContext() permissions length = %v, want %v", len(got.Permissions), len(tt.wantUser.Permissions))
			}
			for ns, perm := range tt.wantUser.Permissions {
				if got.Permissions[ns] != perm {
					t.Errorf("GetUserFromContext() permissions[%s] = %v, want %v", ns, got.Permissions[ns], perm)
				}
			}
		})
	}
}
