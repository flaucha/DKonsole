package permissions

import (
	"context"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/auth"
	"github.com/flaucha/DKonsole/backend/internal/models"
)
func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name             string
		ctx              context.Context
		ldapAdminChecker LDAPAdminChecker
		want             bool
		wantErr          bool
	}{
		{
			name: "core admin is admin",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "admin",
					Role:        "admin",
					Permissions: nil,
				},
			),
			ldapAdminChecker: nil,
			want:             true,
			wantErr:          false,
		},
		{
			name: "user in LDAP admin group is admin",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "ldapuser",
					Role:     "user",
					Permissions: map[string]string{
						"ns1": "view",
					},
				},
			),
			ldapAdminChecker: &mockLDAPAdminChecker{
				userGroups: []string{"admins", "developers"},
				config: &models.LDAPConfig{
					Enabled:     true,
					AdminGroups: []string{"admins"},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "user not in LDAP admin group is not admin",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "ldapuser",
					Role:     "user",
					Permissions: map[string]string{
						"ns1": "view",
					},
				},
			),
			ldapAdminChecker: &mockLDAPAdminChecker{
				userGroups: []string{"developers"},
				config: &models.LDAPConfig{
					Enabled:     true,
					AdminGroups: []string{"admins"},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "regular user is not admin",
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
			ldapAdminChecker: nil,
			want:             false,
			wantErr:          false,
		},
		{
			name: "LDAP disabled means no LDAP admin check",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "ldapuser",
					Role:     "user",
					Permissions: map[string]string{
						"ns1": "view",
					},
				},
			),
			ldapAdminChecker: &mockLDAPAdminChecker{
				userGroups: []string{"admins"},
				config: &models.LDAPConfig{
					Enabled:     false,
					AdminGroups: []string{"admins"},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name:             "no user in context",
			ctx:              context.Background(),
			ldapAdminChecker: nil,
			want:             false,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsAdmin(tt.ctx, tt.ldapAdminChecker)

			if (err != nil) != tt.wantErr {
				t.Errorf("IsAdmin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	tests := []struct {
		name             string
		ctx              context.Context
		ldapAdminChecker LDAPAdminChecker
		wantErr          bool
		errMsg           string
	}{
		{
			name: "core admin passes",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username:    "admin",
					Role:        "admin",
					Permissions: nil,
				},
			),
			ldapAdminChecker: nil,
			wantErr:          false,
		},
		{
			name: "LDAP admin passes",
			ctx: context.WithValue(
				context.Background(),
				auth.UserContextKey(),
				&models.Claims{
					Username: "ldapuser",
					Role:     "user",
					Permissions: map[string]string{
						"ns1": "view",
					},
				},
			),
			ldapAdminChecker: &mockLDAPAdminChecker{
				userGroups: []string{"admins"},
				config: &models.LDAPConfig{
					Enabled:     true,
					AdminGroups: []string{"admins"},
				},
			},
			wantErr: false,
		},
		{
			name: "regular user fails",
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
			ldapAdminChecker: nil,
			wantErr:          true,
			errMsg:           "admin access required",
		},
		{
			name:             "no user in context",
			ctx:              context.Background(),
			ldapAdminChecker: nil,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireAdmin(tt.ctx, tt.ldapAdminChecker)

			if (err != nil) != tt.wantErr {
				t.Errorf("RequireAdmin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("RequireAdmin() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("RequireAdmin() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}
