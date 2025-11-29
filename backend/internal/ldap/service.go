package ldap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// Service provides business logic for LDAP operations
type Service struct {
	repo Repository
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

// TestConnectionRequest represents a request to test LDAP connection
type TestConnectionRequest struct {
	URL        string `json:"url"`
	BaseDN     string `json:"baseDN"`
	UserDN     string `json:"userDN"`
	Username   string `json:"username"`
	Password   string `json:"password"`
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
	// Connect to LDAP server
	conn, err := ldap.DialURL(req.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to LDAP server: %w", err)
	}
	defer conn.Close()

	// Determine bind DN: if username contains "=", it's already a DN, otherwise construct it
	var bindDN string
	if strings.Contains(req.Username, "=") {
		// Username is already a full DN
		bindDN = req.Username
	} else {
		// Construct DN from username, userDN attribute, and baseDN
		bindDN = fmt.Sprintf("%s=%s,%s", req.UserDN, req.Username, req.BaseDN)
	}

	if err := conn.Bind(bindDN, req.Password); err != nil {
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

	// Update in repository
	if err := s.repo.UpdateConfig(r.Context(), &req.Config); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("failed to update LDAP config: %v", err))
		return
	}

	utils.LogInfo("LDAP config updated", map[string]interface{}{
		"enabled": req.Config.Enabled,
		"url":     req.Config.URL,
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
			if perm.Permission != "view" && perm.Permission != "edit" && perm.Permission != "admin" {
				utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid permission: %s. Must be 'view', 'edit', or 'admin'", perm.Permission))
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
	config, err := s.repo.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get LDAP config: %w", err)
	}

	if !config.Enabled {
		return fmt.Errorf("LDAP is not enabled")
	}

	// Get service account credentials for searching
	serviceUsername, servicePassword, err := s.repo.GetCredentials(ctx)
	if err != nil {
		return fmt.Errorf("failed to get LDAP service credentials: %w", err)
	}

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
		// Username is already a full DN
		bindDN = username
	} else {
		// Search for user first to get the full DN
		userSearchFilter := fmt.Sprintf("(%s=%s)", config.UserDN, username)
		if config.UserFilter != "" {
			userSearchFilter = fmt.Sprintf("(&(%s=%s)%s)", config.UserDN, username, config.UserFilter)
		}
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
			bindDN = fmt.Sprintf("%s=%s,%s", config.UserDN, username, config.BaseDN)
		} else {
			// Use the found DN
			bindDN = userSr.Entries[0].DN
		}
	}

	// Now bind with user credentials
	if err := conn.Bind(bindDN, password); err != nil {
		return fmt.Errorf("failed to bind to LDAP server: %w", err)
	}

	return nil
}

// GetUserGroups retrieves the groups for a user from LDAP
func (s *Service) GetUserGroups(ctx context.Context, username string) ([]string, error) {
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

	// Connect to LDAP server
	conn, err := ldap.DialURL(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %w", err)
	}
	defer conn.Close()

	// Bind with service account credentials
	// Determine bind DN: if serviceUsername contains "=", it's already a DN, otherwise construct it
	var bindDN string
	if strings.Contains(serviceUsername, "=") {
		// Service username is already a full DN
		bindDN = serviceUsername
	} else {
		// Construct DN from service username, userDN attribute, and baseDN
		bindDN = fmt.Sprintf("%s=%s,%s", config.UserDN, serviceUsername, config.BaseDN)
	}
	if err := conn.Bind(bindDN, servicePassword); err != nil {
		return nil, fmt.Errorf("failed to bind to LDAP server: %w", err)
	}

	// Determine user DN for search
	var userDN string
	if strings.Contains(username, "=") {
		// Username is already a full DN
		userDN = username
	} else {
		// Search for user first to get the full DN
		userSearchFilter := fmt.Sprintf("(%s=%s)", config.UserDN, username)
		if config.UserFilter != "" {
			userSearchFilter = fmt.Sprintf("(&(%s=%s)%s)", config.UserDN, username, config.UserFilter)
		}
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
			userDN = fmt.Sprintf("%s=%s,%s", config.UserDN, username, config.BaseDN)
		} else {
			// Use the found DN
			userDN = userSr.Entries[0].DN
		}
	}

	// First, try to get groups from memberOf attribute on the user entry (AD-style)
	userSearchRequest := ldap.NewSearchRequest(
		userDN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)",
		[]string{"memberOf"},
		nil,
	)

	userSr, err := conn.Search(userSearchRequest)
	if err == nil && len(userSr.Entries) > 0 {
		// Extract groups from memberOf attribute
		memberOf := userSr.Entries[0].GetAttributeValues("memberOf")
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
		if len(groups) > 0 {
			return groups, nil
		}
	}

	// Fallback: Search for groups with member attribute (standard LDAP)
	searchFilter := fmt.Sprintf("(&(objectClass=groupOfNames)(member=%s))", userDN)
	if config.UserFilter != "" {
		searchFilter = fmt.Sprintf("(&(objectClass=groupOfNames)(member=%s)%s)", userDN, config.UserFilter)
	}

	searchRequest := ldap.NewSearchRequest(
		config.GroupDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		[]string{"cn"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search LDAP: %w", err)
	}

	groups := make([]string, 0, len(sr.Entries))
	for _, entry := range sr.Entries {
		if cn := entry.GetAttributeValue("cn"); cn != "" {
			groups = append(groups, cn)
		}
	}

	return groups, nil
}

// GetUserPermissions retrieves the permissions for a user based on their LDAP groups
func (s *Service) GetUserPermissions(ctx context.Context, username string) (map[string]string, error) {
	// Get user groups
	groups, err := s.GetUserGroups(ctx, username)
	if err != nil {
		return nil, err
	}

	// Get groups configuration
	groupsConfig, err := s.repo.GetGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get LDAP groups config: %w", err)
	}

	// Build permissions map: namespace -> highest permission
	permissions := make(map[string]string)

	// Create a map of group names for quick lookup
	groupMap := make(map[string]bool)
	for _, group := range groups {
		groupMap[group] = true
	}

	// Permission hierarchy: view < edit < admin
	permissionLevel := map[string]int{
		"view":  1,
		"edit":  2,
		"admin": 3,
	}

	// Find permissions for user's groups
	for _, group := range groupsConfig.Groups {
		if groupMap[group.Name] {
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

	return permissions, nil
}

// isValidLDAPURL validates LDAP URL format
func isValidLDAPURL(url string) bool {
	return len(url) > 0 && (url[:7] == "ldap://" || url[:8] == "ldaps://")
}
