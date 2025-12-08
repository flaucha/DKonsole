package ldap

import (
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

// We need a way to mock GetUserGroups inside GetUserPermissions.
// Since we refactored GetUserGroups to delegate to getUserGroupsWithRepo,
// but GetUserPermissions calls GetUserGroups (the public method),
// we still run into the issue of it creating a real repo and client.

// However, we can use a small trick: GetUserPermissions just needs the group list.
// If we can control GetUserGroups, we are golden.
// Since we can't easily mock the method on the struct, we will use a different approach.
// We can make GetUserGroups logic swappable or split the permission logic.
// The permission logic calculates the map based on a list of groups.
// That part is pure logic and easy to test if extracted.

// Let's refactor permissions.go to extract `calculatePermissions(groups []string, config *models.LDAPConfig, groupsConfig *models.LDAPGroupsConfig)`.
// This is the cleanest way and avoids needing to mock the whole LDAP stack for permission logic testing.

func TestCalculatePermissions(t *testing.T) {
	config := &models.LDAPConfig{
		Enabled:     true,
		AdminGroups: []string{"admin-group"},
	}
	groupsConfig := &models.LDAPGroupsConfig{
		Groups: []models.LDAPGroup{
			{
				Name: "dev-group",
				Permissions: []models.LDAPGroupPermission{
					{Namespace: "dev", Permission: "edit"},
					{Namespace: "prod", Permission: "view"},
				},
			},
			{
				Name: "prod-viewers",
				Permissions: []models.LDAPGroupPermission{
					{Namespace: "prod", Permission: "view"},
				},
			},
		},
	}

	service := &Service{}

	t.Run("Admin Group", func(t *testing.T) {
		userGroups := []string{"other", "admin-group"}
		perms := service.calculatePermissions(userGroups, config, groupsConfig)
		if perms != nil {
			t.Errorf("expected nil permissions for admin, got %v", perms)
		}
	})

	t.Run("Regular User", func(t *testing.T) {
		userGroups := []string{"dev-group", "prod-viewers"}
		perms := service.calculatePermissions(userGroups, config, groupsConfig)
		
		if perms["dev"] != "edit" {
			t.Errorf("dev namespace: got %s, want edit", perms["dev"])
		}
		if perms["prod"] != "view" {
			t.Errorf("prod namespace: got %s, want view", perms["prod"])
		}
	})
	
	t.Run("Permission Upgrade", func(t *testing.T) {
		// Test that higher permission wins
		groupsConfigWithOverlap := &models.LDAPGroupsConfig{
			Groups: []models.LDAPGroup{
				{
					Name: "team-a",
					Permissions: []models.LDAPGroupPermission{
						{Namespace: "shared", Permission: "view"},
					},
				},
				{
					Name: "team-b",
					Permissions: []models.LDAPGroupPermission{
						{Namespace: "shared", Permission: "edit"},
					},
				},
			},
		}
		userGroups := []string{"team-a", "team-b"}
		perms := service.calculatePermissions(userGroups, config, groupsConfigWithOverlap)
		if perms["shared"] != "edit" {
			t.Errorf("shared namespace: got %s, want edit", perms["shared"])
		}
	})
	
	t.Run("No Groups", func(t *testing.T) {
		perms := service.calculatePermissions([]string{}, config, groupsConfig)
		if len(perms) != 0 {
			t.Error("expected empty permissions")
		}
	})
}
