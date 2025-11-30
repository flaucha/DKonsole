package ldap

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"github.com/go-ldap/ldap/v3"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// LDAPClient manages LDAP connections with pooling and TLS support
type LDAPClient struct {
	config    *models.LDAPConfig
	pool      *connectionPool
	mu        sync.RWMutex
	tlsConfig *tls.Config
}

// connectionPool manages a pool of LDAP connections
type connectionPool struct {
	url         string
	tlsConfig   *tls.Config
	timeout     time.Duration
	connections chan *ldap.Conn
	maxSize     int
	mu          sync.Mutex
}

// NewLDAPClient creates a new LDAP client with connection pooling
func NewLDAPClient(config *models.LDAPConfig) (*LDAPClient, error) {
	if config == nil {
		return nil, fmt.Errorf("LDAP config cannot be nil")
	}

	// Configure TLS
	tlsConfig, err := buildTLSConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build TLS config: %w", err)
	}

	// Create connection pool
	pool := newConnectionPool(config.URL, tlsConfig, 10*time.Second, 5)

	return &LDAPClient{
		config:    config,
		pool:      pool,
		tlsConfig: tlsConfig,
	}, nil
}

// buildTLSConfig builds TLS configuration from LDAP config
func buildTLSConfig(config *models.LDAPConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		ServerName:         extractServerName(config.URL),
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	// If CACert is provided, use it
	if config.CACert != "" {
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM([]byte(config.CACert)) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}

// extractServerName extracts server name from LDAP URL
func extractServerName(url string) string {
	// Simple extraction - remove ldap:// or ldaps:// and port
	if len(url) < 7 {
		return ""
	}
	if url[:7] == "ldap://" {
		url = url[7:]
	} else if len(url) >= 8 && url[:8] == "ldaps://" {
		url = url[8:]
	}
	// Remove port if present
	if idx := len(url); idx > 0 {
		for i := 0; i < len(url); i++ {
			if url[i] == ':' {
				return url[:i]
			}
		}
		return url
	}
	return ""
}

// newConnectionPool creates a new connection pool
func newConnectionPool(url string, tlsConfig *tls.Config, timeout time.Duration, maxSize int) *connectionPool {
	pool := &connectionPool{
		url:         url,
		tlsConfig:   tlsConfig,
		timeout:     timeout,
		connections: make(chan *ldap.Conn, maxSize),
		maxSize:     maxSize,
	}

	// Pre-populate pool with a few connections
	for i := 0; i < 2 && i < maxSize; i++ {
		conn, err := pool.createConnection()
		if err != nil {
			// Log but don't fail - pool will create connections on demand
			utils.LogWarn("Failed to pre-populate LDAP connection pool", map[string]interface{}{
				"error": err.Error(),
			})
			break
		}
		pool.connections <- conn
	}

	return pool
}

// createConnection creates a new LDAP connection
func (p *connectionPool) createConnection() (*ldap.Conn, error) {
	var conn *ldap.Conn
	var err error

	if p.tlsConfig != nil && (len(p.url) >= 8 && p.url[:8] == "ldaps://") {
		// Use TLS for ldaps://
		conn, err = ldap.DialURL(p.url, ldap.DialWithTLSConfig(p.tlsConfig))
	} else {
		// Plain connection for ldap://
		conn, err = ldap.DialURL(p.url)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to dial LDAP server: %w", err)
	}

	// Set connection timeout
	conn.SetTimeout(p.timeout)

	return conn, nil
}

// getConnection gets a connection from the pool or creates a new one
func (p *connectionPool) getConnection() (*ldap.Conn, error) {
	select {
	case conn := <-p.connections:
		// Check if connection is still valid
		if conn != nil {
			return conn, nil
		}
	default:
		// No connection available, create new one
	}

	return p.createConnection()
}

// returnConnection returns a connection to the pool
func (p *connectionPool) returnConnection(conn *ldap.Conn) {
	if conn == nil {
		return
	}

	// Check if pool is full
	select {
	case p.connections <- conn:
		// Connection returned to pool
	default:
		// Pool is full, close the connection
		conn.Close()
	}
}

// close closes all connections in the pool
func (p *connectionPool) close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.connections)
	for conn := range p.connections {
		if conn != nil {
			conn.Close()
		}
	}
}

// GetConnection gets a connection from the pool
func (c *LDAPClient) GetConnection() (*ldap.Conn, error) {
	return c.pool.getConnection()
}

// ReturnConnection returns a connection to the pool
func (c *LDAPClient) ReturnConnection(conn *ldap.Conn) {
	c.pool.returnConnection(conn)
}

// Close closes the client and all connections
func (c *LDAPClient) Close() {
	if c.pool != nil {
		c.pool.close()
	}
}

// UpdateConfig updates the client configuration
func (c *LDAPClient) UpdateConfig(config *models.LDAPConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close old pool
	if c.pool != nil {
		c.pool.close()
	}

	// Update config
	c.config = config

	// Rebuild TLS config
	tlsConfig, err := buildTLSConfig(config)
	if err != nil {
		return fmt.Errorf("failed to rebuild TLS config: %w", err)
	}
	c.tlsConfig = tlsConfig

	// Create new pool
	pool := newConnectionPool(config.URL, tlsConfig, 10*time.Second, 5)
	c.pool = pool

	return nil
}
