package ldap

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// ldapDialer allows mocking ldap.DialURL in tests
var ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
	return ldap.DialURL(url, opts...)
}

// prepareAuthentication validates the request and retrieves LDAP configuration
func (s *Service) prepareAuthentication(ctx context.Context, username string) (*models.LDAPConfig, error) {
	// Validate input
	if err := validateLDAPUsername(username); err != nil {
		utils.LogWarn("Invalid username format", map[string]interface{}{
			"username": username,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("invalid username format: %w", err)
	}

	config, err := s.repo.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get LDAP config: %w", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("LDAP is not enabled")
	}

	return config, nil
}

func (s *Service) connectAndBindService(ctx context.Context, config *models.LDAPConfig, contextUsername string) (LDAPConnection, error) {
	// Get service account credentials for searching
	serviceUsername, servicePassword, err := s.repo.GetCredentials(ctx)
	if err != nil {
		utils.LogWarn("Failed to get LDAP service credentials", map[string]interface{}{
			"username": contextUsername,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("failed to get LDAP service credentials: %w", err)
	}

	// Connect to LDAP server
	conn, err := ldapDialer(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %w", err)
	}

	// Bind with service account
	serviceBindDN := buildBindDN(serviceUsername, config)
	if err := conn.Bind(serviceBindDN, servicePassword); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to bind with service account: %w", err)
	}

	return conn, nil
}

func (s *Service) findUserDN(conn LDAPConnection, config *models.LDAPConfig, username string) (string, error) {
	if strings.Contains(username, "=") {
		// Username is already a full DN
		if !isValidLDAPDN(username) {
			return "", fmt.Errorf("invalid LDAP DN format")
		}
		return username, nil
	}

	// Search for user
	escapedUsername := sanitizeLDAPFilter(username)
	userSearchFilter := fmt.Sprintf("(%s=%s)", config.UserDN, escapedUsername)
	if config.UserFilter != "" {
		userSearchFilter = fmt.Sprintf("(&(%s=%s)%s)", config.UserDN, escapedUsername, config.UserFilter)
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
		// Fallback: construct DN
		return fmt.Sprintf("%s=%s,%s", config.UserDN, escapedUsername, config.BaseDN), nil
	}

	return userSr.Entries[0].DN, nil
}

func (s *Service) verifyUserPassword(config *models.LDAPConfig, bindDN, password, username string) error {
	utils.LogInfo("Attempting to bind with user credentials", map[string]interface{}{
		"username": username,
		"bindDN":   bindDN,
	})

	// New connection for user bind to avoid messing up other connection or if unbind closes it
	// New connection for user bind to avoid messing up other connection or if unbind closes it
	conn, err := ldapDialer(config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to LDAP server for user bind: %w", err)
	}
	defer conn.Close()

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
