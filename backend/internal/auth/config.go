package auth

import (
	"log"
	"os"

	"github.com/example/k8s-view/internal/utils"
)

var (
	// jwtSecret is the secret key for signing tokens
	jwtSecret []byte
)

func init() {
	jwtSecretStr := os.Getenv("JWT_SECRET")
	if len(jwtSecretStr) == 0 {
		utils.LogWarn("JWT_SECRET environment variable must be set", map[string]interface{}{
			"level": "critical",
		})
		if os.Getenv("GO_ENV") == "production" {
			log.Fatal("JWT_SECRET is required in production")
		}
		if len(jwtSecretStr) == 0 {
			log.Fatal("CRITICAL: JWT_SECRET environment variable must be set")
		}
	}
	if len(jwtSecretStr) < 32 {
		log.Fatal("CRITICAL: JWT_SECRET must be at least 32 characters long")
	}
	jwtSecret = []byte(jwtSecretStr)
}

// GetJWTSecret returns the JWT secret key
func GetJWTSecret() []byte {
	return jwtSecret
}
