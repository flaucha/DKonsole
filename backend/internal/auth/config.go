package auth

import (
	"log"
	"os"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

var (
	// jwtSecret is the secret key for signing tokens
	jwtSecret []byte
)

func init() {
	jwtSecretStr := os.Getenv("JWT_SECRET")
	if len(jwtSecretStr) == 0 {
		// JWT_SECRET is optional in setup mode (will be created during setup)
		// In non-production environments, use a default for testing/development
		if os.Getenv("GO_ENV") == "production" {
			// In production, we'll allow it to be empty if we're in setup mode
			// The actual validation will happen when we try to use it
			utils.LogWarn("JWT_SECRET environment variable not set - may be in setup mode", map[string]interface{}{
				"level": "warning",
			})
		} else {
			// In non-production, use a default for testing/development
			jwtSecretStr = "default-secret-key-for-testing-only-change-in-production"
			utils.LogWarn("Using default JWT_SECRET (INSECURE - for testing only)", map[string]interface{}{
				"level": "warning",
			})
		}
	}
	if len(jwtSecretStr) > 0 && len(jwtSecretStr) < 32 {
		log.Fatal("CRITICAL: JWT_SECRET must be at least 32 characters long")
	}
	if len(jwtSecretStr) > 0 {
		jwtSecret = []byte(jwtSecretStr)
	}
}

// GetJWTSecret returns the JWT secret key
func GetJWTSecret() []byte {
	return jwtSecret
}
