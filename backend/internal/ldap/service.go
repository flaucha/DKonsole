package ldap

import (
	"context"
	"fmt"

	"sync"

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

// Methods GetConfigHandler, GetLDAPStatusHandler, UpdateConfigHandler, GetGroupsHandler, UpdateGroupsHandler, GetCredentialsHandler, UpdateCredentialsHandler, TestConnectionHandler are in handlers.go
// Structs TestConnectionRequest, UpdateConfigRequest, UpdateGroupsRequest, UpdateCredentialsRequest are in handlers.go
// func isValidLDAPURL is in validation.go

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

// AuthenticateUser authenticates a user against LDAP
func (s *Service) AuthenticateUser(ctx context.Context, username, password string) error {
	utils.LogInfo("AuthenticateUser called", map[string]interface{}{
		"username": username,
	})

	// Prepare authentication (validate input, get config)
	config, err := s.prepareAuthentication(ctx, username)
	if err != nil {
		return err
	}

	// Connection and Service Bind
	conn, err := s.connectAndBindService(ctx, config, username)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Connect to LDAP server for user bind if needed?
	// Note: conn is bound with service account. We need to find User DN first using this separate connection?
	// Actually, we can reuse 'conn' for searching, then we can try binding with another connection or re-bind?
	// ldap v3 supports re-binding on same connection usually? No, standard says unbind closes connection.
	// But go-ldap `Bind` usually sends bind request.
	// However, usually we use a separate connection for user bind to verify password, or we use the same one if we don't care about the service bind status afterwards.
	// The original code dial new connection: `conn, err := ldap.DialURL(config.URL)`

	// Find User DN
	bindDN, err := s.findUserDN(conn, config, username)
	if err != nil {
		return err
	}

	// Verify User Password (Bind with User DN)
	return s.verifyUserPassword(config, bindDN, password, username)
}

// Methods GetUserGroups, ValidateUserGroup, GetUserPermissions are in groups.go
// Helper methods loginAndBindService, findUserDN, verifyUserPassword are in auth_helpers.go
