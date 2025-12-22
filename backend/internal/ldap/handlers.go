package ldap

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

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

// GetConfigHandler returns the current LDAP configuration
func (s *Service) GetConfigHandler(w http.ResponseWriter, r *http.Request) {
	s.refreshRepoClient()
	config, err := s.repo.GetConfig(r.Context())
	if err != nil {
		utils.HandleErrorJSON(w, err, "Failed to get LDAP config", http.StatusInternalServerError, nil)
		return
	}

	// Don't return credentials in the response
	utils.JSONResponse(w, http.StatusOK, config)
}

// GetLDAPStatusHandler returns whether LDAP is enabled (public endpoint for login page)
func (s *Service) GetLDAPStatusHandler(w http.ResponseWriter, r *http.Request) {
	s.refreshRepoClient()
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
	s.refreshRepoClient()
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
		utils.HandleErrorJSON(w, err, "Failed to update LDAP config", http.StatusInternalServerError, map[string]interface{}{
			"enabled": req.Config.Enabled,
		})
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
	s.refreshRepoClient()
	groups, err := s.repo.GetGroups(r.Context())
	if err != nil {
		utils.HandleErrorJSON(w, err, "Failed to get LDAP groups", http.StatusInternalServerError, nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, groups)
}

// UpdateGroupsHandler updates the LDAP groups configuration
func (s *Service) UpdateGroupsHandler(w http.ResponseWriter, r *http.Request) {
	s.refreshRepoClient()
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
		utils.HandleErrorJSON(w, err, "Failed to update LDAP groups", http.StatusInternalServerError, map[string]interface{}{
			"groups_count": len(req.Groups.Groups),
		})
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
	s.refreshRepoClient()
	username, _, err := s.repo.GetCredentials(r.Context())
	if err != nil {
		utils.HandleErrorJSON(w, err, "Failed to get LDAP credentials", http.StatusInternalServerError, nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"configured": username != "",
		"username":   username,
	})
}

// UpdateCredentialsHandler updates the LDAP credentials
func (s *Service) UpdateCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	s.refreshRepoClient()
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
			utils.HandleErrorJSON(w, err, "Failed to get LDAP credentials", http.StatusInternalServerError, nil)
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
		utils.HandleErrorJSON(w, err, "Failed to update LDAP credentials", http.StatusInternalServerError, map[string]interface{}{
			"username": req.Username,
		})
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
