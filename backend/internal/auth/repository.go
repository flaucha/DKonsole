package auth

import (
	"os"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	GetAdminUser() (string, error)
	GetAdminPasswordHash() (string, error)
}

// EnvUserRepository implements UserRepository using environment variables
type EnvUserRepository struct{}

// NewEnvUserRepository creates a new EnvUserRepository
func NewEnvUserRepository() *EnvUserRepository {
	return &EnvUserRepository{}
}

// GetAdminUser retrieves the admin username from environment variables
func (r *EnvUserRepository) GetAdminUser() (string, error) {
	adminUser := os.Getenv("ADMIN_USER")
	if adminUser == "" {
		return "", ErrAdminUserNotSet
	}
	return adminUser, nil
}

// GetAdminPasswordHash retrieves the admin password hash from environment variables
func (r *EnvUserRepository) GetAdminPasswordHash() (string, error) {
	adminPassHash := os.Getenv("ADMIN_PASSWORD")
	if adminPassHash == "" {
		return "", ErrAdminPasswordNotSet
	}
	return adminPassHash, nil
}

// Errors
var (
	ErrAdminUserNotSet     = &AuthError{Message: "ADMIN_USER environment variable not set"}
	ErrAdminPasswordNotSet = &AuthError{Message: "ADMIN_PASSWORD environment variable not set"}
)

// AuthError represents an authentication error
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}





