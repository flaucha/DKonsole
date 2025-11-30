package ldap

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-ldap/ldap/v3"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// validateLDAPUsername validates username for LDAP search
// Prevents injection attacks and ensures valid format
func validateLDAPUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Check length (reasonable limit)
	if len(username) > 256 {
		return fmt.Errorf("username too long (max 256 characters)")
	}

	// Blacklist dangerous characters for LDAP injection
	// These are already escaped by ldap.EscapeFilter(), but we validate anyway
	dangerousChars := []string{
		"*",    // Wildcard
		"(",    // Filter start
		")",    // Filter end
		"\\",   // Escape character
		"\x00", // Null byte
	}

	for _, char := range dangerousChars {
		if strings.Contains(username, char) {
			return fmt.Errorf("username contains invalid character: %s", char)
		}
	}

	return nil
}

// isValidLDAPDN validates a full LDAP Distinguished Name
// Example: "cn=admin,dc=example,dc=com"
// A valid DN must have at least two components (must contain at least one comma)
func isValidLDAPDN(dn string) bool {
	// Basic DN validation - must contain at least one "="
	if !strings.Contains(dn, "=") {
		return false
	}

	// A valid DN must have at least one comma (at least two components)
	if !strings.Contains(dn, ",") {
		return false
	}

	// Validate DN format with regex
	// DN components: cn, ou, dc, etc.
	// Pattern requires at least one comma (two components minimum)
	dnRegex := regexp.MustCompile(`^(?:[a-zA-Z]+=[^,]+,)+[a-zA-Z]+=[^,]+$`)
	return dnRegex.MatchString(dn)
}

// sanitizeLDAPFilter is a wrapper around ldap.EscapeFilter with logging
// Use this for all user-controlled inputs in LDAP filters
func sanitizeLDAPFilter(input string) string {
	escaped := ldap.EscapeFilter(input)

	// Log if escaping changed the input (potential injection attempt)
	if escaped != input {
		utils.LogWarn("LDAP filter sanitized - potential injection attempt", map[string]interface{}{
			"original": input,
			"escaped":  escaped,
		})
	}

	return escaped
}
