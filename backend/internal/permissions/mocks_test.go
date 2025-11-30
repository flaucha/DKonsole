package permissions

import (
	"context"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

// mockLDAPAdminChecker is a mock implementation of LDAPAdminChecker for testing
type mockLDAPAdminChecker struct {
	userGroups   []string
	config       *models.LDAPConfig
	getGroupsErr error
	getConfigErr error
}

func (m *mockLDAPAdminChecker) GetUserGroups(ctx context.Context, username string) ([]string, error) {
	if m.getGroupsErr != nil {
		return nil, m.getGroupsErr
	}
	return m.userGroups, nil
}

func (m *mockLDAPAdminChecker) GetConfig(ctx context.Context) (*models.LDAPConfig, error) {
	if m.getConfigErr != nil {
		return nil, m.getConfigErr
	}
	return m.config, nil
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
