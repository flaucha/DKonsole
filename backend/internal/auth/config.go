package auth

import (
	"crypto/rand"
	"encoding/hex"
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
		if os.Getenv("GO_ENV") == "production" {
			// In production we expect Kubernetes secret to be present; leave empty to force failure if used.
			utils.LogWarn("JWT_SECRET environment variable not set - relying on cluster secret or setup mode", map[string]interface{}{
				"level": "warning",
			})
		} else {
			// Generate a random secret for non-production to avoid predictable defaults
			buf := make([]byte, 32)
			if _, err := rand.Read(buf); err != nil {
				log.Fatal("CRITICAL: failed to generate JWT secret:", err)
			}
			jwtSecretStr = hex.EncodeToString(buf)
			utils.LogWarn("Generated random JWT_SECRET for non-production (set JWT_SECRET to persist sessions)", map[string]interface{}{
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
