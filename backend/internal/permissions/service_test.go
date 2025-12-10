package permissions

import (
	"context"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

// MockLDAPAdminChecker mocks the LDAPAdminChecker interface
type MockLDAPAdminChecker struct {
	Groups []string
}

func (m *MockLDAPAdminChecker) GetUserGroups(ctx context.Context, username string) ([]string, error) {
	return m.Groups, nil
}

func (m *MockLDAPAdminChecker) GetConfig(ctx context.Context) (*models.LDAPConfig, error) {
	return nil, nil
}

func TestNewService(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Error("NewService() returned nil")
	}
}

func TestSetLDAPAdminChecker(t *testing.T) {
	svc := NewService()
	checker := &MockLDAPAdminChecker{Groups: []string{"admin"}}

	svc.SetLDAPAdminChecker(checker)

	if svc.ldapAdminChecker == nil {
		t.Error("SetLDAPAdminChecker() failed to set checker")
	}
}
