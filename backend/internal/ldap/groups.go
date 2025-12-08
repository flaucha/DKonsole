package ldap

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// GetUserGroups retrieves the groups for a user from LDAP
func (s *Service) GetUserGroups(ctx context.Context, username string) ([]string, error) {
	utils.LogInfo("GetUserGroups called", map[string]interface{}{
		"username": username,
	})
	config, err := s.repo.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get LDAP config: %w", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("LDAP is not enabled")
	}

	serviceUsername, servicePassword, err := s.repo.GetCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get LDAP credentials: %w", err)
	}

	utils.LogInfo("Got service credentials", map[string]interface{}{
		"service_username": serviceUsername,
	})

	// Get or initialize client
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get LDAP client: %w", err)
	}

	// Get connection from pool
	repo, err := NewLDAPClientRepository(client)
	if err != nil {
		return nil, fmt.Errorf("failed to get LDAP connection: %w", err)
	}
	defer repo.Close()

	return s.getUserGroupsWithRepo(ctx, repo, config, serviceUsername, servicePassword, username)
}

func (s *Service) getUserGroupsWithRepo(ctx context.Context, repo LDAPClientRepository, config *models.LDAPConfig, serviceUsername, servicePassword, username string) ([]string, error) {
	// Bind with service account credentials
	bindDN := buildBindDN(serviceUsername, config)
	if err := repo.Bind(ctx, bindDN, servicePassword); err != nil {
		return nil, fmt.Errorf("failed to bind to LDAP server: %w", err)
	}

	// Determine user DN for search
	var userDN string
	if strings.Contains(username, "=") {
		// Username is already a full DN
		userDN = username
		utils.LogInfo("Username is already a DN", map[string]interface{}{
			"username": username,
			"userDN":   userDN,
		})
	} else {
		// Search for user first to get the full DN
		userDN = searchUserDN(ctx, repo, config, username)
		utils.LogInfo("Found user DN", map[string]interface{}{
			"username": username,
			"userDN":   userDN,
		})
	}

	utils.LogInfo("Searching for user groups", map[string]interface{}{
		"username": username,
		"userDN":   userDN,
	})

	// First, try to get groups from memberOf attribute on the user entry (AD-style)
	// Some LDAP servers (like Active Directory) store group membership in the memberOf attribute of the user
	// This is more efficient than searching for groups that contain the user
	utils.LogInfo("Searching for memberOf attribute", map[string]interface{}{
		"username": username,
		"userDN":   userDN,
		"scope":    "base",
	})
	// Request all attributes first to see what's available
	entries, err := repo.Search(ctx, userDN, ldap.ScopeBaseObject, "(objectClass=*)", []string{"*"})
	if err != nil {
		utils.LogWarn("Failed to search user for memberOf", map[string]interface{}{
			"username": username,
			"userDN":   userDN,
			"error":    err.Error(),
		})
	} else if len(entries) == 0 {
		utils.LogWarn("User entry not found when searching for memberOf", map[string]interface{}{
			"username": username,
			"userDN":   userDN,
		})
	} else {
		// Log all available attributes for debugging
		entry := entries[0]
		allAttrs := entry.Attributes
		attrNames := make([]string, 0, len(allAttrs))
		for _, attr := range allAttrs {
			attrNames = append(attrNames, attr.Name)
		}
		utils.LogInfo("User entry attributes", map[string]interface{}{
			"username":   username,
			"userDN":     userDN,
			"attributes": attrNames,
		})

		// Extract groups from memberOf attribute
		memberOf := entry.GetAttributeValues("memberOf")
		utils.LogInfo("Found memberOf attributes", map[string]interface{}{
			"username":       username,
			"userDN":         userDN,
			"memberOf":       memberOf,
			"memberOf_count": len(memberOf),
		})
		if len(memberOf) == 0 {
			utils.LogWarn("User has no memberOf attributes", map[string]interface{}{
				"username":             username,
				"userDN":               userDN,
				"available_attributes": attrNames,
			})
		}
		groups := make([]string, 0, len(memberOf))
		for _, groupDN := range memberOf {
			// Extract group name from DN
			// Try CN first (e.g., "cn=admin,ou=groups,dc=glauth,dc=com" -> "admin")
			if strings.HasPrefix(groupDN, "cn=") {
				cn := strings.Split(strings.Split(groupDN, ",")[0], "=")[1]
				groups = append(groups, cn)
			} else if strings.Contains(groupDN, "cn=") {
				// Handle cases where CN is in a different position
				parts := strings.Split(groupDN, ",")
				for _, part := range parts {
					if strings.HasPrefix(part, "cn=") {
						cn := strings.Split(part, "=")[1]
						groups = append(groups, cn)
						break
					}
				}
			} else if strings.HasPrefix(groupDN, "ou=") {
				// Handle cases where group DN starts with OU (e.g., "ou=devsA,ou=groups,dc=glauth,dc=com" -> "devsA")
				ou := strings.Split(strings.Split(groupDN, ",")[0], "=")[1]
				groups = append(groups, ou)
			}
		}
		utils.LogInfo("Extracted groups from memberOf", map[string]interface{}{
			"username": username,
			"groups":   groups,
		})
		if len(groups) > 0 {
			return groups, nil
		}
	}

	utils.LogWarn("No groups found from memberOf, trying fallback search", map[string]interface{}{
		"username": username,
		"userDN":   userDN,
	})

	// Fallback: Search for groups with member or uniqueMember attribute
	// Try both groupOfNames (member) and groupOfUniqueNames (uniqueMember)
	// Sanitizar userDN antes de usarlo en el filtro
	escapedUserDN := sanitizeLDAPFilter(userDN)
	searchFilter := fmt.Sprintf("(|(member=%s)(uniqueMember=%s))", escapedUserDN, escapedUserDN)
	if config.UserFilter != "" {
		searchFilter = fmt.Sprintf("(&(|(member=%s)(uniqueMember=%s))%s)", escapedUserDN, escapedUserDN, config.UserFilter)
	}

	utils.LogInfo("Trying fallback search for groups", map[string]interface{}{
		"username":     username,
		"userDN":       userDN,
		"groupDN":      config.GroupDN,
		"searchFilter": searchFilter,
	})

	fallbackEntries, err := repo.Search(ctx, config.GroupDN, ldap.ScopeWholeSubtree, searchFilter, []string{"cn"})
	if err != nil {
		utils.LogWarn("Fallback search failed", map[string]interface{}{
			"username": username,
			"userDN":   userDN,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("failed to search LDAP: %w", err)
	}

	utils.LogInfo("Fallback search completed", map[string]interface{}{
		"username":    username,
		"userDN":      userDN,
		"entry_count": len(fallbackEntries),
	})

	groups := make([]string, 0, len(fallbackEntries))
	for _, entry := range fallbackEntries {
		if cn := entry.GetAttributeValue("cn"); cn != "" {
			groups = append(groups, cn)
		}
	}

	utils.LogInfo("Groups extracted from fallback search", map[string]interface{}{
		"username": username,
		"groups":   groups,
	})

	return groups, nil
}

// ValidateUserGroup checks if the user belongs to the required group (if configured)
func (s *Service) ValidateUserGroup(ctx context.Context, username string) error {
	utils.LogInfo("ValidateUserGroup called", map[string]interface{}{
		"username": username,
	})
	config, err := s.repo.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get LDAP config: %w", err)
	}

	// If no required group is configured, allow all authenticated users
	if config.RequiredGroup == "" {
		utils.LogInfo("No required group configured, allowing user", map[string]interface{}{
			"username": username,
		})
		return nil
	}

	utils.LogInfo("Required group is configured", map[string]interface{}{
		"username":       username,
		"required_group": config.RequiredGroup,
	})

	// Get user groups
	groups, err := s.GetUserGroups(ctx, username)
	if err != nil {
		utils.LogWarn("Failed to get user groups for validation", map[string]interface{}{
			"username": username,
			"error":    err.Error(),
		})
		return fmt.Errorf("failed to get user groups: %w", err)
	}

	utils.LogInfo("Validating user group membership", map[string]interface{}{
		"username":       username,
		"user_groups":    groups,
		"required_group": config.RequiredGroup,
	})

	// Check if user belongs to required group
	for _, group := range groups {
		if group == config.RequiredGroup {
			utils.LogInfo("User belongs to required group", map[string]interface{}{
				"username":       username,
				"required_group": config.RequiredGroup,
			})
			return nil
		}
	}

	// User doesn't belong to required group
	utils.LogWarn("User does not belong to required group", map[string]interface{}{
		"username":       username,
		"user_groups":    groups,
		"required_group": config.RequiredGroup,
	})
	return fmt.Errorf("user is not a member of required group: %s", config.RequiredGroup)
}


