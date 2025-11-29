package ldap

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"

	"github.com/flaucha/DKonsole/backend/internal/models"
)

// LDAPClientRepository encapsulates LDAP data access operations
type LDAPClientRepository interface {
	// Bind authenticates with the given DN and password
	Bind(ctx context.Context, bindDN, password string) error

	// Search performs an LDAP search operation
	Search(ctx context.Context, baseDN string, scope int, filter string, attributes []string) ([]*ldap.Entry, error)

	// Close closes the underlying connection
	Close() error
}

// ldapClientRepository implements LDAPClientRepository using the connection pool
type ldapClientRepository struct {
	client *LDAPClient
	conn   *ldap.Conn
}

// NewLDAPClientRepository creates a new LDAP client repository
func NewLDAPClientRepository(client *LDAPClient) (LDAPClientRepository, error) {
	conn, err := client.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get LDAP connection: %w", err)
	}

	return &ldapClientRepository{
		client: client,
		conn:   conn,
	}, nil
}

// Bind authenticates with the given DN and password
func (r *ldapClientRepository) Bind(ctx context.Context, bindDN, password string) error {
	if r.conn == nil {
		return fmt.Errorf("LDAP connection is nil")
	}
	return r.conn.Bind(bindDN, password)
}

// Search performs an LDAP search operation
func (r *ldapClientRepository) Search(ctx context.Context, baseDN string, scope int, filter string, attributes []string) ([]*ldap.Entry, error) {
	if r.conn == nil {
		return nil, fmt.Errorf("LDAP connection is nil")
	}

	searchRequest := ldap.NewSearchRequest(
		baseDN,
		scope,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		attributes,
		nil,
	)

	sr, err := r.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("LDAP search failed: %w", err)
	}

	return sr.Entries, nil
}

// Close closes the connection and returns it to the pool
func (r *ldapClientRepository) Close() error {
	if r.conn != nil {
		r.client.ReturnConnection(r.conn)
		r.conn = nil
	}
	return nil
}

// buildBindDN constructs a bind DN from username and config
func buildBindDN(username string, config *models.LDAPConfig) string {
	if strings.Contains(username, "=") {
		// Username is already a full DN
		return username
	}
	// Construct DN from username, userDN attribute, and baseDN
	return fmt.Sprintf("%s=%s,%s", config.UserDN, username, config.BaseDN)
}

// searchUserDN searches for a user's DN in LDAP
func searchUserDN(ctx context.Context, repo LDAPClientRepository, config *models.LDAPConfig, username string) (string, error) {
	// Search for user first to get the full DN
	userSearchFilter := fmt.Sprintf("(%s=%s)", config.UserDN, username)
	if config.UserFilter != "" {
		userSearchFilter = fmt.Sprintf("(&(%s=%s)%s)", config.UserDN, username, config.UserFilter)
	}

	entries, err := repo.Search(ctx, config.BaseDN, ldap.ScopeWholeSubtree, userSearchFilter, []string{"dn"})
	if err != nil || len(entries) == 0 {
		// Fallback: construct DN from username, userDN attribute, and baseDN
		return buildBindDN(username, config), nil
	}

	// Use the found DN
	return entries[0].DN, nil
}
