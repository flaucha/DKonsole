package ldap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/go-ldap/ldap/v3"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// Service provides business logic for LDAP operations
type Service struct {
	repo   Repository
	client *LDAPClient
	mu     sync.RWMutex
}

// GetConfig returns the LDAP configuration (for internal use)
func (s *Service) GetConfig(ctx context.Context) (*models.LDAPConfig, error) {
	return s.repo.GetConfig(ctx)
}

// NewService creates a new LDAP service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// initializeClient initializes or updates the LDAP client with current config
func (s *Service) initializeClient(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, err := s.repo.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get LDAP config: %w", err)
	}

	if !config.Enabled {
		// Close existing client if disabled
		if s.client != nil {
			s.client.Close()
			s.client = nil
		}
		return nil
	}

	// If client exists and config hasn't changed, reuse it
	if s.client != nil {
		// Update config if needed
		if err := s.client.UpdateConfig(config); err != nil {
			utils.LogWarn("Failed to update LDAP client config, recreating", map[string]interface{}{
				"error": err.Error(),
			})
			s.client.Close()
			s.client = nil
		} else {
			return nil
		}
	}

	// Create new client
	client, err := NewLDAPClient(config)
	if err != nil {
		return fmt.Errorf("failed to create LDAP client: %w", err)
	}

	s.client = client
	return nil
}

// getClient gets or initializes the LDAP client
func (s *Service) getClient(ctx context.Context) (*LDAPClient, error) {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()

	if client == nil {
		if err := s.initializeClient(ctx); err != nil {
			return nil, err
		}
		s.mu.RLock()
		client = s.client
		s.mu.RUnlock()
	}

	return client, nil
}

// TestConnectionRequest represents a request to test LDAP connection
type TestConnectionRequest struct {
	URL      string `json:"url"`
	BaseDN   string `json:"baseDN"`
	UserDN   string `json:"userDN"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// UpdateConfigRequest represents a request to update LDAP configuration
type UpdateConfigRequest struct {
	Config models.LDAPConfig `json:"config"`
}

// UpdateGroupsRequest represents a request to update LDAP groups
type UpdateGroupsRequest struct {
	Groups models.LDAPGroupsConfig `json:"groups"`
}

// UpdateCredentialsRequest represents a request to update LDAP credentials
type UpdateCredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// TestConnection tests the LDAP connection with provided credentials
func (s *Service) TestConnection(ctx context.Context, req TestConnectionRequest) error {
	// Create a temporary config for testing
	testConfig := &models.LDAPConfig{
		URL:                req.URL,
		BaseDN:             req.BaseDN,
		UserDN:             req.UserDN,
		InsecureSkipVerify: false, // Default to secure for testing
	}

	// Create temporary client for testing
	client, err := NewLDAPClient(testConfig)
	if err != nil {
		return fmt.Errorf("failed to create LDAP client: %w", err)
	}
	defer client.Close()

	// Get connection from pool
	repo, err := NewLDAPClientRepository(client)
	if err != nil {
		return fmt.Errorf("failed to get LDAP connection: %w", err)
	}
	defer repo.Close()

	// Determine bind DN: if username contains "=", it's already a DN, otherwise construct it
	bindDN := buildBindDN(req.Username, testConfig)

	if err := repo.Bind(ctx, bindDN, req.Password); err != nil {
		return fmt.Errorf("failed to bind to LDAP server: %w", err)
	}

	return nil
}

// GetConfigHandler returns the current LDAP configuration
func (s *Service) GetConfigHandler(w http.ResponseWriter, r *http.Request) {
	config, err := s.repo.GetConfig(r.Context())
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("failed to get LDAP config: %v", err))
		return
	}

	// Don't return credentials in the response
	utils.JSONResponse(w, http.StatusOK, config)
}

// GetLDAPStatusHandler returns whether LDAP is enabled (public endpoint for login page)
func (s *Service) GetLDAPStatusHandler(w http.ResponseWriter, r *http.Request) {
	config, err := s.repo.GetConfig(r.Context())
	if err != nil {
		// If error, assume LDAP is not enabled
		utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
			"enabled": false,
		})
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"enabled": config != nil && config.Enabled,
	})
}

// UpdateConfigHandler updates the LDAP configuration
func (s *Service) UpdateConfigHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate URL format if enabled
	if req.Config.Enabled && req.Config.URL != "" {
		if !isValidLDAPURL(req.Config.URL) {
			utils.ErrorResponse(w, http.StatusBadRequest, "invalid LDAP URL format. Must start with ldap:// or ldaps://")
			return
		}
	}

	// Warn if InsecureSkipVerify is enabled
	if req.Config.InsecureSkipVerify {
		utils.LogWarn("LDAP InsecureSkipVerify enabled - TLS certificate verification is disabled", map[string]interface{}{
			"url": req.Config.URL,
		})
	}

	// Update in repository
	if err := s.repo.UpdateConfig(r.Context(), &req.Config); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("failed to update LDAP config: %v", err))
		return
	}

	// Reinitialize client with new config
	if err := s.initializeClient(r.Context()); err != nil {
		utils.LogWarn("Failed to reinitialize LDAP client after config update", map[string]interface{}{
			"error": err.Error(),
		})
		// Don't fail the request, just log the warning
	}

	utils.LogInfo("LDAP config updated", map[string]interface{}{
		"enabled":            req.Config.Enabled,
		"url":                req.Config.URL,
		"insecureSkipVerify": req.Config.InsecureSkipVerify,
		"hasCACert":          req.Config.CACert != "",
	})

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "LDAP configuration updated successfully",
	})
}

// GetGroupsHandler returns the current LDAP groups configuration
func (s *Service) GetGroupsHandler(w http.ResponseWriter, r *http.Request) {
	groups, err := s.repo.GetGroups(r.Context())
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("failed to get LDAP groups: %v", err))
		return
	}

	utils.JSONResponse(w, http.StatusOK, groups)
}

// UpdateGroupsHandler updates the LDAP groups configuration
func (s *Service) UpdateGroupsHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdateGroupsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogWarn("Failed to decode LDAP groups request", map[string]interface{}{
			"error": err.Error(),
		})
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Validate groups and filter out empty permissions
	for i := range req.Groups.Groups {
		group := &req.Groups.Groups[i]
		if group.Name == "" {
			utils.ErrorResponse(w, http.StatusBadRequest, "group name cannot be empty")
			return
		}

		// Filter out permissions with empty namespace (incomplete permissions)
		validPermissions := make([]models.LDAPGroupPermission, 0)
		for _, perm := range group.Permissions {
			if perm.Namespace == "" {
				// Skip empty permissions (user hasn't selected a namespace yet)
				continue
			}
			if perm.Permission != "view" && perm.Permission != "edit" {
				utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid permission: %s. Must be 'view' or 'edit'", perm.Permission))
				return
			}
			validPermissions = append(validPermissions, perm)
		}
		// Update group with filtered permissions
		group.Permissions = validPermissions
	}

	// Update in repository
	if err := s.repo.UpdateGroups(r.Context(), &req.Groups); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("failed to update LDAP groups: %v", err))
		return
	}

	utils.LogInfo("LDAP groups updated", map[string]interface{}{
		"groups_count": len(req.Groups.Groups),
	})

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "LDAP groups updated successfully",
	})
}

// GetCredentialsHandler returns whether credentials are set and the username (but not the password)
func (s *Service) GetCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	username, _, err := s.repo.GetCredentials(r.Context())
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("failed to get LDAP credentials: %v", err))
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"configured": username != "",
		"username":   username,
	})
}

// UpdateCredentialsHandler updates the LDAP credentials
func (s *Service) UpdateCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdateCredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "username cannot be empty")
		return
	}

	// If password is empty, get existing password to keep it
	if req.Password == "" {
		existingUsername, existingPassword, err := s.repo.GetCredentials(r.Context())
		if err != nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("failed to get existing credentials: %v", err))
			return
		}

		// If username unchanged and password empty, nothing to update
		if existingUsername == req.Username {
			utils.JSONResponse(w, http.StatusOK, map[string]string{
				"message": "No changes to save",
			})
			return
		}

		// Username changed but password empty - use existing password
		if existingPassword == "" {
			utils.ErrorResponse(w, http.StatusBadRequest, "password required for new credentials")
			return
		}
		req.Password = existingPassword
	}

	// Update in repository
	if err := s.repo.UpdateCredentials(r.Context(), req.Username, req.Password); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("failed to update LDAP credentials: %v", err))
		return
	}

	utils.LogInfo("LDAP credentials updated", map[string]interface{}{
		"username": req.Username,
	})

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "LDAP credentials updated successfully",
	})
}

// TestConnectionHandler tests the LDAP connection
func (s *Service) TestConnectionHandler(w http.ResponseWriter, r *http.Request) {
	var req TestConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.TestConnection(r.Context(), req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("LDAP connection test failed: %v", err))
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "LDAP connection test successful",
	})
}

// AuthenticateUser authenticates a user against LDAP
func (s *Service) AuthenticateUser(ctx context.Context, username, password string) error {
	utils.LogInfo("AuthenticateUser called", map[string]interface{}{
		"username": username,
	})

	// Validar input antes de usar
	if err := validateLDAPUsername(username); err != nil {
		utils.LogWarn("Invalid username format", map[string]interface{}{
			"username": username,
			"error":    err.Error(),
		})
		return fmt.Errorf("invalid username format: %w", err)
	}

	config, err := s.repo.GetConfig(ctx)
	if err != nil {
		utils.LogWarn("Failed to get LDAP config", map[string]interface{}{
			"username": username,
			"error":    err.Error(),
		})
		return fmt.Errorf("failed to get LDAP config: %w", err)
	}

	if !config.Enabled {
		utils.LogWarn("LDAP is not enabled", map[string]interface{}{
			"username": username,
		})
		return fmt.Errorf("LDAP is not enabled")
	}

	// Get service account credentials for searching
	serviceUsername, servicePassword, err := s.repo.GetCredentials(ctx)
	if err != nil {
		utils.LogWarn("Failed to get LDAP service credentials", map[string]interface{}{
			"username": username,
			"error":    err.Error(),
		})
		return fmt.Errorf("failed to get LDAP service credentials: %w", err)
	}

	utils.LogInfo("Got service credentials for authentication", map[string]interface{}{
		"username":         username,
		"service_username": serviceUsername,
	})

	// Connect to LDAP server
	conn, err := ldap.DialURL(config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to LDAP server: %w", err)
	}
	defer conn.Close()

	// Bind with service account first to search for user
	var serviceBindDN string
	if strings.Contains(serviceUsername, "=") {
		serviceBindDN = serviceUsername
	} else {
		serviceBindDN = fmt.Sprintf("%s=%s,%s", config.UserDN, serviceUsername, config.BaseDN)
	}
	if err := conn.Bind(serviceBindDN, servicePassword); err != nil {
		return fmt.Errorf("failed to bind with service account: %w", err)
	}

	// Determine user DN: if username contains "=", it's already a DN, otherwise search for it
	var bindDN string
	if strings.Contains(username, "=") {
		// Username is already a full DN - validar que sea DN válido
		if !isValidLDAPDN(username) {
			return fmt.Errorf("invalid LDAP DN format")
		}
		bindDN = username
	} else {
		// Search for user first to get the full DN
		// SANITIZAR username con sanitizeLDAPFilter()
		escapedUsername := sanitizeLDAPFilter(username)
		userSearchFilter := fmt.Sprintf("(%s=%s)", config.UserDN, escapedUsername)
		if config.UserFilter != "" {
			userSearchFilter = fmt.Sprintf("(&(%s=%s)%s)", config.UserDN, escapedUsername, config.UserFilter)
		}

		utils.LogInfo("LDAP search filter (sanitized)", map[string]interface{}{
			"filter": userSearchFilter,
		})

		userSearchRequest := ldap.NewSearchRequest(
			config.BaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			userSearchFilter,
			[]string{"dn"},
			nil,
		)
		userSr, err := conn.Search(userSearchRequest)
		if err != nil || len(userSr.Entries) == 0 {
			// Fallback: construct DN from username, userDN attribute, and baseDN
			bindDN = fmt.Sprintf("%s=%s,%s", config.UserDN, escapedUsername, config.BaseDN)
		} else {
			// Use the found DN
			bindDN = userSr.Entries[0].DN
		}
	}

	// Now bind with user credentials
	utils.LogInfo("Attempting to bind with user credentials", map[string]interface{}{
		"username": username,
		"bindDN":   bindDN,
	})
	if err := conn.Bind(bindDN, password); err != nil {
		utils.LogWarn("Failed to bind with user credentials", map[string]interface{}{
			"username": username,
			"bindDN":   bindDN,
			"error":    err.Error(),
		})
		return fmt.Errorf("failed to bind to LDAP server: %w", err)
	}

	utils.LogInfo("User authenticated successfully", map[string]interface{}{
		"username": username,
		"bindDN":   bindDN,
	})
	return nil
}

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

	// Check if user belongs to admin groups (admins have full access, no namespace-specific permissions needed)
	config, err := s.repo.GetConfig(ctx)
	if err == nil && config != nil && config.Enabled && len(config.AdminGroups) > 0 {
		// Check if user belongs to any admin group
		for _, adminGroup := range config.AdminGroups {
			for _, userGroup := range groups {
				if userGroup == adminGroup {
					utils.LogInfo("GetUserPermissions: user is admin group member, returning empty permissions", map[string]interface{}{
						"username":    username,
						"admin_group": adminGroup,
					})
					// Admin has full access, return empty permissions
					return make(map[string]string), nil
				}
			}
		}
	}

	// Get groups configuration
	groupsConfig, err := s.repo.GetGroups(ctx)
	if err != nil {
		utils.LogWarn("Failed to get LDAP groups config", map[string]interface{}{
			"username": username,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("failed to get LDAP groups config: %w", err)
	}

	// Log configured groups for debugging
	configuredGroupNames := make([]string, 0, len(groupsConfig.Groups))
	for _, group := range groupsConfig.Groups {
		configuredGroupNames = append(configuredGroupNames, group.Name)
	}
	utils.LogInfo("GetUserPermissions: configured groups", map[string]interface{}{
		"username":          username,
		"configured_groups": configuredGroupNames,
	})

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
	matchedGroups := make([]string, 0)
	for _, group := range groupsConfig.Groups {
		// Case-insensitive comparison
		groupNameLower := strings.ToLower(group.Name)
		if groupMap[group.Name] || groupMap[groupNameLower] {
			matchedGroups = append(matchedGroups, group.Name)
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

	utils.LogInfo("GetUserPermissions: permissions calculated", map[string]interface{}{
		"username":       username,
		"matched_groups": matchedGroups,
		"permissions":    permissions,
	})

	return permissions, nil
}

// isValidLDAPURL validates LDAP URL format
func isValidLDAPURL(url string) bool {
	return len(url) > 0 && (url[:7] == "ldap://" || url[:8] == "ldaps://")
}
