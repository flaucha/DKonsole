package auth

import (
	"os"
)

// UserRepository defines the interface for user data access.
// Implementations should retrieve user credentials from a secure source.
type UserRepository interface {
	GetAdminUser() (string, error)         // Returns the admin username
	GetAdminPasswordHash() (string, error) // Returns the admin password hash (Argon2 format)
}

// EnvUserRepository implements UserRepository using environment variables.
// It reads ADMIN_USER and ADMIN_PASSWORD from the environment.
type EnvUserRepository struct{}

// NewEnvUserRepository creates a new EnvUserRepository instance.
func NewEnvUserRepository() *EnvUserRepository {
	return &EnvUserRepository{}
}

// GetAdminUser retrieves the admin username from the ADMIN_USER environment variable.
// Returns ErrAdminUserNotSet if the environment variable is not set.
func (r *EnvUserRepository) GetAdminUser() (string, error) {
	adminUser := os.Getenv("ADMIN_USER")
	if adminUser == "" {
		return "", ErrAdminUserNotSet
	}
	return adminUser, nil
}

// GetAdminPasswordHash retrieves the admin password hash from the ADMIN_PASSWORD environment variable.
// The hash should be in Argon2 format (e.g., $argon2id$v=19$m=65536,t=3,p=4$salt$hash).
// Returns ErrAdminPasswordNotSet if the environment variable is not set.
func (r *EnvUserRepository) GetAdminPasswordHash() (string, error) {
	adminPassHash := os.Getenv("ADMIN_PASSWORD")
	if adminPassHash == "" {
		return "", ErrAdminPasswordNotSet
	}
	return adminPassHash, nil
}

// Predefined authentication errors.
var (
	// ErrAdminUserNotSet is returned when ADMIN_USER environment variable is not set.
	ErrAdminUserNotSet = &AuthError{Message: "ADMIN_USER environment variable not set"}
	// ErrAdminPasswordNotSet is returned when ADMIN_PASSWORD environment variable is not set.
	ErrAdminPasswordNotSet = &AuthError{Message: "ADMIN_PASSWORD environment variable not set"}
)

// AuthError represents an authentication error.
// It implements the error interface.
type AuthError struct {
	Message string // Human-readable error message
}

func (e *AuthError) Error() string {
	return e.Message
}
