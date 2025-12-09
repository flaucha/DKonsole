package ldap

import (
	"context"
	"fmt"
	"strings"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// GetUserPermissions retrieves the permissions for a user based on their LDAP groups
// If the user belongs to an admin group, returns empty permissions (admin has full access)
func (s *Service) GetUserPermissions(ctx context.Context, username string) (map[string]string, error) {
	// Get user groups first
	groups, err := s.GetUserGroups(ctx, username)
	if err != nil {
		utils.LogWarn("Failed to get user groups for permissions", map[string]interface{}{
			"username": username,
			"error":    err.Error(),
		})
		return nil, err
	}

	utils.LogInfo("GetUserPermissions: user groups retrieved", map[string]interface{}{
		"username": username,
		"groups":   groups,
	})

	// Get configuration
	config, err := s.repo.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get LDAP config: %w", err)
	}

	// Get groups configuration
	groupsConfig, err := s.repo.GetGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get LDAP groups config: %w", err)
	}

	return s.calculatePermissions(groups, config, groupsConfig), nil
}

func (s *Service) calculatePermissions(groups []string, config *models.LDAPConfig, groupsConfig *models.LDAPGroupsConfig) map[string]string {
	// Check if user belongs to admin groups (admins have full access, no namespace-specific permissions needed)
	if config != nil && config.Enabled && len(config.AdminGroups) > 0 {
		// Check if user belongs to any admin group
		for _, adminGroup := range config.AdminGroups {
			for _, userGroup := range groups {
				if userGroup == adminGroup {
					utils.LogInfo("GetUserPermissions: user is admin group member, returning nil permissions", map[string]interface{}{
						"admin_group": adminGroup,
					})
					// Admin has full access, return nil (not empty map) to indicate admin status
					return nil
				}
			}
		}
	}

	// Build permissions map: namespace -> highest permission
	permissions := make(map[string]string)

	// Create a map of group names for quick lookup (case-insensitive comparison)
	groupMap := make(map[string]bool)
	for _, group := range groups {
		groupMap[strings.ToLower(group)] = true
		groupMap[group] = true // Also keep original case for exact match
	}

	// Permission hierarchy: view < edit
	permissionLevel := map[string]int{
		"view": 1,
		"edit": 2,
	}

	// Find permissions for user's groups
	for _, group := range groupsConfig.Groups {
		// Case-insensitive comparison
		groupNameLower := strings.ToLower(group.Name)
		if groupMap[group.Name] || groupMap[groupNameLower] {
			for _, perm := range group.Permissions {
				currentLevel := permissionLevel[perm.Permission]
				existingLevel := 0
				if existing, exists := permissions[perm.Namespace]; exists {
					existingLevel = permissionLevel[existing]
				}
				// Use the highest permission level
				if currentLevel > existingLevel {
					permissions[perm.Namespace] = perm.Permission
				}
			}
		}
	}

	return permissions
}
