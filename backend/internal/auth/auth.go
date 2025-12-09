package auth

import (
	"context"
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// Service provides HTTP handlers for authentication operations.
// It follows a layered architecture:
//   - Handler (HTTP): Handles HTTP requests/responses
//   - Service (Business Logic): AuthService and JWTService
//   - Repository (Data Access): UserRepository for credential retrieval
type Service struct {
	authService *AuthService
	jwtService  *JWTService
	k8sRepo     *K8sUserRepository // K8s repository for secret management (may be nil if not using K8s)
	setupMode   bool               // true if running in setup mode (secret doesn't exist)
	mu          sync.RWMutex       // Mutex for thread-safe reload
	k8sClient   kubernetes.Interface
	secretName  string
}

// SetLDAPAuthenticator sets the LDAP authenticator for the auth service
func (s *Service) SetLDAPAuthenticator(ldapAuth LDAPAuthenticator) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.authService != nil {
		s.authService.SetLDAPAuthenticator(ldapAuth)
	}
}

// IsSetupMode reports whether the service is waiting for initial setup (secret missing)
func (s *Service) IsSetupMode() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.setupMode
}

// NewService creates a new authentication service with default configuration.
// It initializes the user repository and JWT services.
// If k8sClient is provided, it will try to use Kubernetes secrets.
// If k8sClient is nil, it falls back to environment variables.
// secretName is the name of the Kubernetes secret to use (default: "dkonsole-auth").
func NewService(k8sClient kubernetes.Interface, secretName string) (*Service, error) {
	var userRepo UserRepository
	var jwtSecret []byte
	var setupMode bool
	var k8sRepo *K8sUserRepository

	if k8sClient != nil && secretName != "" {
		// Try to use Kubernetes secrets
		repo, err := NewK8sUserRepository(k8sClient, secretName)
		if err != nil {
			utils.LogWarn("Failed to initialize K8s repository, falling back to environment variables", map[string]interface{}{
				"error": err.Error(),
			})
			// Fall back to environment variables
			userRepo = NewEnvUserRepository()
			jwtSecret = GetJWTSecret()
		} else {
			k8sRepo = repo
			// Check if secret exists
			ctx := context.Background()
			exists, err := repo.SecretExists(ctx)
			if err != nil {
				utils.LogWarn("Failed to check secret existence, falling back to environment variables", map[string]interface{}{
					"error": err.Error(),
				})
				userRepo = NewEnvUserRepository()
				jwtSecret = GetJWTSecret()
			} else if !exists {
				// Secret doesn't exist - setup mode
				setupMode = true
				utils.LogInfo("Running in setup mode - secret does not exist", map[string]interface{}{
					"secret_name": secretName,
				})
				// Don't initialize authService in setup mode - it will fail without credentials
				return &Service{
					authService: nil, // Will be nil in setup mode
					jwtService:  nil, // Will be nil in setup mode
					k8sRepo:     k8sRepo,
					setupMode:   true,
					k8sClient:   k8sClient,
					secretName:  secretName,
				}, nil
			} else {
				// Secret exists - use K8s repository
				userRepo = k8sRepo
				// Get JWT secret from the secret
				secret, err := k8sClient.CoreV1().Secrets(repo.namespace).Get(ctx, secretName, metav1.GetOptions{})
				if err != nil {
					return nil, fmt.Errorf("failed to get JWT secret from Kubernetes secret: %w", err)
				}
				jwtSecretBytes, exists := secret.Data["jwt-secret"]
				if !exists || len(jwtSecretBytes) == 0 {
					return nil, fmt.Errorf("jwt-secret key not found in secret")
				}
				jwtSecret = jwtSecretBytes
			}
		}
	} else {
		// No K8s client provided - use environment variables
		userRepo = NewEnvUserRepository()
		jwtSecret = GetJWTSecret()
	}

	// Initialize services (only if not in setup mode)
	authService := NewAuthService(userRepo, jwtSecret)
	jwtService := NewJWTService(jwtSecret)

	return &Service{
		authService: authService,
		jwtService:  jwtService,
		k8sRepo:     k8sRepo,
		setupMode:   setupMode,
		k8sClient:   k8sClient,
		secretName:  secretName,
	}, nil
}

// Reload attempts to reload the service configuration if the secret now exists.
// This allows the service to transition from setup mode to normal mode without restarting.
// Returns true if reload was successful, false otherwise.
func (s *Service) Reload(ctx context.Context) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Only reload if we're in setup mode and have K8s client
	if !s.setupMode || s.k8sClient == nil || s.k8sRepo == nil {
		return false, nil
	}

	// Check if secret now exists
	exists, err := s.k8sRepo.SecretExists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check secret existence during reload: %w", err)
	}

	if !exists {
		// Secret still doesn't exist, no reload needed
		return false, nil
	}

	// Secret exists now - reload configuration
	utils.LogInfo("Reloading auth service - secret now exists", map[string]interface{}{
		"secret_name": s.secretName,
	})

	// Get credentials from secret
	userRepo := s.k8sRepo

	// Get JWT secret from the secret
	secret, err := s.k8sClient.CoreV1().Secrets(s.k8sRepo.namespace).Get(ctx, s.secretName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get JWT secret from Kubernetes secret during reload: %w", err)
	}

	jwtSecretBytes, exists := secret.Data["jwt-secret"]
	if !exists || len(jwtSecretBytes) == 0 {
		return false, fmt.Errorf("jwt-secret key not found in secret during reload")
	}

	// Initialize services with new credentials
	authService := NewAuthService(userRepo, jwtSecretBytes)
	jwtService := NewJWTService(jwtSecretBytes)

	// Update service state
	s.authService = authService
	s.jwtService = jwtService
	s.setupMode = false

	utils.LogInfo("Auth service reloaded successfully", map[string]interface{}{
		"secret_name": s.secretName,
	})

	return true, nil
}

// Methods LoginHandler, LogoutHandler, MeHandler, ChangePasswordHandler are in http_handlers.go
// Method AuthMiddleware and context keys are in middleware.go
