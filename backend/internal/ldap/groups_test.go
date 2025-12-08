package ldap

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/go-ldap/ldap/v3"
)

// enhancedFakeLDAPClientRepository for groups testing
type enhancedFakeLDAPClientRepository struct {
	searchFunc func(baseDN string, filter string) ([]*ldap.Entry, error)
	bindFunc   func(bindDN, password string) error
	closeFunc  func() error
}

func (f *enhancedFakeLDAPClientRepository) Bind(ctx context.Context, bindDN, password string) error {
	if f.bindFunc != nil {
		return f.bindFunc(bindDN, password)
	}
	return nil
}

func (f *enhancedFakeLDAPClientRepository) Search(ctx context.Context, baseDN string, scope int, filter string, attributes []string) ([]*ldap.Entry, error) {
	if f.searchFunc != nil {
		return f.searchFunc(baseDN, filter)
	}
	return nil, nil
}

func (f *enhancedFakeLDAPClientRepository) Close() error {
	if f.closeFunc != nil {
		return f.closeFunc()
	}
	return nil
}

func TestGetUserGroupsWithRepo(t *testing.T) {
	service := &Service{}
	config := &models.LDAPConfig{
		Enabled: true,
		BaseDN:  "dc=example,dc=com",
		UserDN:  "uid",
		GroupDN: "ou=groups,dc=example,dc=com",
	}

	t.Run("User Already DN", func(t *testing.T) {
		repo := &enhancedFakeLDAPClientRepository{
			searchFunc: func(baseDN, filter string) ([]*ldap.Entry, error) {
				// memberOf search
				return []*ldap.Entry{{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{Name: "memberOf", Values: []string{"cn=dev,ou=groups"}},
					},
				}}, nil
			},
		}

		groups, err := service.getUserGroupsWithRepo(context.Background(), repo, config, "srv", "pass", "uid=test,dc=example,dc=com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(groups) != 1 || groups[0] != "dev" {
			t.Errorf("got groups %v, want [dev]", groups)
		}
	})

	t.Run("Extract Groups Logic", func(t *testing.T) {
		// Test extraction logic from memberOf
		repo := &enhancedFakeLDAPClientRepository{
			searchFunc: func(baseDN, filter string) ([]*ldap.Entry, error) {
				return []*ldap.Entry{{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{Name: "memberOf", Values: []string{
							"cn=admin,ou=groups",
							"ou=editors,ou=groups",
							"cn=viewers,ou=other",
						}},
					},
				}}, nil
			},
		}
		groups, err := service.getUserGroupsWithRepo(context.Background(), repo, config, "srv", "pass", "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := map[string]bool{"admin": true, "editors": true, "viewers": true}
		if len(groups) != 3 {
			t.Errorf("expected 3 groups, got %d", len(groups))
		}
		for _, g := range groups {
			if !expected[g] {
				t.Errorf("unexpected group: %s", g)
			}
		}
	})
	
	t.Run("Fallback Search", func(t *testing.T) {
		repo := &enhancedFakeLDAPClientRepository{
			searchFunc: func(baseDN, filter string) ([]*ldap.Entry, error) {
				if baseDN == config.BaseDN {
					// User search failing to return memberOf attributes or empty
					return nil, nil 
				}
				// Group search
				return []*ldap.Entry{
					{Attributes: []*ldap.EntryAttribute{{Name: "cn", Values: []string{"dev"}}}},
				}, nil
			},
		}
		groups, err := service.getUserGroupsWithRepo(context.Background(), repo, config, "srv", "pass", "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(groups) != 1 || groups[0] != "dev" {
			t.Errorf("got groups %v, want [dev]", groups)
		}
	})
	
	t.Run("Search Error", func(t *testing.T) {
		repo := &enhancedFakeLDAPClientRepository{
			searchFunc: func(baseDN, filter string) ([]*ldap.Entry, error) {
				return nil, errors.New("search fail")
			},
		}
		// Expect fallback search to also fail
		_, err := service.getUserGroupsWithRepo(context.Background(), repo, config, "srv", "pass", "test")
		if err == nil {
			t.Error("expected error")
		}
	})
	
	t.Run("Bind Error", func(t *testing.T) {
		repo := &enhancedFakeLDAPClientRepository{
			bindFunc: func(dn, pass string) error {
				return errors.New("bind fail")
			},
		}
		_, err := service.getUserGroupsWithRepo(context.Background(), repo, config, "srv", "pass", "test")
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestValidateUserGroup(t *testing.T) {
	originalDialer := ldapDialer
	defer func() { ldapDialer = originalDialer }()

	tests := []struct {
		name          string
		config        *models.LDAPConfig
		userGroups    []string
		wantError     bool
	}{
		{
			name:      "No Configured Required Group",
			config:    &models.LDAPConfig{Enabled: true, RequiredGroup: ""},
			wantError: false,
		},
		{
			name:       "Required Group Found",
			config:     &models.LDAPConfig{Enabled: true, RequiredGroup: "admins"},
			userGroups: []string{"devs", "admins"},
			wantError:  false,
		},
		{
			name:       "Required Group Missing",
			config:     &models.LDAPConfig{Enabled: true, RequiredGroup: "superusers"},
			userGroups: []string{"devs", "admins"},
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock Repo
			repo := &mockRepository{
				config: tt.config,
				// Mock credentials for GetUserGroups
				username: "svc",
				password: "pwd",
			}
			
			// Mock LDAP Connection for GetUserGroups
			setupMockSearch(tt.userGroups)
			
			service := NewService(repo)
			
			err := service.ValidateUserGroup(context.Background(), "user")
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateUserGroup() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// Helper to setup mock search for user groups
func setupMockSearch(groups []string) {
	ldapDialer = func(url string, opts ...ldap.DialOpt) (LDAPConnection, error) {
		return &mockLDAPConnection{
			searchFunc: func(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
				// Very basic mock: if request filter looks like group search or user search
				// For simple test, we assume fallback search or memberOf return matches
				
				// Let's assume we return memberOf for user
				attrs := []*ldap.EntryAttribute{}
				vals := []string{}
				for _, g := range groups {
					vals = append(vals, fmt.Sprintf("cn=%s,ou=groups", g))
				}
				if len(vals) > 0 {
					attrs = append(attrs, &ldap.EntryAttribute{Name: "memberOf", Values: vals})
				}
				
				return &ldap.SearchResult{
					Entries: []*ldap.Entry{
						{DN: "cn=user,dc=example", Attributes: attrs},
					},
				}, nil
			},
		}, nil
	}
}
