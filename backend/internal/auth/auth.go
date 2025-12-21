package auth

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

const minJWTSecretLenBytes = 32

// Service provides HTTP handlers for authentication operations.
// It follows a layered architecture:
//   - Handler (HTTP): Handles HTTP requests/responses
//   - Service (Business Logic): AuthService and JWTService
//   - Repository (Data Access): UserRepository for credential retrieval
type Service struct {
	authService   *AuthService
	jwtService    *JWTService
	k8sRepo       *K8sUserRepository // K8s repository for secret management (may be nil if not using K8s)
	setupMode     bool               // true if running in setup mode (secret doesn't exist)
	mu            sync.RWMutex       // Mutex for thread-safe reload
	k8sClient     kubernetes.Interface
	secretName    string
	ClientFactory func(token string) (kubernetes.Interface, error) // Factory for creating K8s clients
	OnReload      func(token string)                               // Callback to notify about reload (e.g. to update global clients)
}

func validateJWTSecret(secret []byte) error {
	if len(secret) == 0 {
		return fmt.Errorf("jwt secret is not configured")
	}
	if len(secret) < minJWTSecretLenBytes {
		return fmt.Errorf("jwt secret must be at least %d bytes", minJWTSecretLenBytes)
	}
	return nil
}

func newSetupModeService(k8sClient kubernetes.Interface, secretName string, k8sRepo *K8sUserRepository) *Service {
	return &Service{
		authService: nil, // Will be nil in setup mode
		jwtService:  nil, // Will be nil in setup mode
		k8sRepo:     k8sRepo,
		setupMode:   true,
		k8sClient:   k8sClient,
		secretName:  secretName,
	}
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

	isProduction := os.Getenv("GO_ENV") == "production"

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
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			exists, err := repo.SecretExists(ctx)
			if err != nil {
				utils.LogWarn("Failed to check secret existence, falling back to environment variables", map[string]interface{}{
					"error": err.Error(),
				})
				userRepo = NewEnvUserRepository()
				jwtSecret = GetJWTSecret()
			} else if !exists {
				// Secret doesn't exist - setup mode
				utils.LogInfo("Running in setup mode - secret does not exist", map[string]interface{}{
					"secret_name": secretName,
				})
				// Don't initialize authService in setup mode - it will fail without credentials
				return newSetupModeService(k8sClient, secretName, k8sRepo), nil
			} else {
				// Secret exists - use K8s repository
				userRepo = k8sRepo
				// Get JWT secret from the secret. If the app cannot read secrets (RBAC),
				// we fall back to env vars (typically injected by the pod spec).
				secret, err := k8sClient.CoreV1().Secrets(repo.namespace).Get(ctx, secretName, metav1.GetOptions{})
				if err != nil {
					utils.LogWarn("Failed to read auth secret, falling back to environment variables", map[string]interface{}{
						"error":       err.Error(),
						"secret_name": secretName,
					})
					userRepo = NewEnvUserRepository()
					jwtSecret = GetJWTSecret()
				} else {
					jwtSecretBytes := secret.Data["jwt-secret"]
					if len(jwtSecretBytes) == 0 {
						// Common case: Helm (or another installer) created an empty secret to satisfy mounts/env refs.
						// Treat it as incomplete setup and allow the setup flow to populate it.
						utils.LogWarn("Auth secret exists but is missing jwt-secret; entering setup mode", map[string]interface{}{
							"secret_name": secretName,
						})
						return newSetupModeService(k8sClient, secretName, k8sRepo), nil
					}
					jwtSecret = jwtSecretBytes
				}
			}
		}
	} else {
		// No K8s client provided - use environment variables
		userRepo = NewEnvUserRepository()
		jwtSecret = GetJWTSecret()
	}

	if err := validateJWTSecret(jwtSecret); err != nil {
		if isProduction {
			return nil, fmt.Errorf("invalid JWT secret (set JWT_SECRET or complete setup): %w", err)
		}
		return nil, fmt.Errorf("invalid JWT secret: %w", err)
	}

	// Initialize services (only if not in setup mode)
	authService := NewAuthService(userRepo, jwtSecret)
	jwtService := NewJWTService(jwtSecret)

	return &Service{
		authService:   authService,
		jwtService:    jwtService,
		k8sRepo:       k8sRepo,
		setupMode:     setupMode,
		k8sClient:     k8sClient,
		secretName:    secretName,
		ClientFactory: createEphemeralClient,
	}, nil
}

// Reload attempts to reload the service configuration if the secret now exists.
// This allows the service to transition from setup mode to normal mode without restarting.
// tokenOverride: Optional token to use for creating the K8s client (e.g. from setup/token update)
// Returns true if reload was successful, false otherwise.
func (s *Service) Reload(ctx context.Context, tokenOverride string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Only reload if we're in setup mode and have K8s client OR if we have a token override
	if (!s.setupMode || s.k8sClient == nil || s.k8sRepo == nil) && tokenOverride == "" {
		return false, nil
	}

	// Update K8s Client if token is provided
	if tokenOverride != "" {
		utils.LogInfo("Reloading with new token", nil)
		var newClient kubernetes.Interface
		var err error
		if s.ClientFactory != nil {
			newClient, err = s.ClientFactory(tokenOverride)
		} else {
			newClient, err = createEphemeralClient(tokenOverride)
		}

		if err != nil {
			return false, fmt.Errorf("failed to create client from token: %w", err)
		}

		s.k8sClient = newClient
		// Re-initialize repo with new client
		repo, err := NewK8sUserRepository(s.k8sClient, s.secretName)
		if err != nil {
			return false, fmt.Errorf("failed to recreate repo with new client: %w", err)
		}
		// Preserve the factory in the repo if needed for internal updates (k8sRepo has its own factory field)
		// We could try to copy it if K8sUserRepository exposed it, but NewK8sUserRepository creates a fresh one.
		// If we are in tests, we might lose the mock factory inside k8sRepo?
		// K8sUserRepository struct has ClientFactory field. NewK8sUserRepository does NOT set it (it stays nil).
		// We should probably propagate our factory if possible, but K8sUserRepository field is exported?
		// Yes ClientFactory in K8sUserRepository is exported.
		repo.ClientFactory = s.ClientFactory

		s.k8sRepo = repo
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
	if err := validateJWTSecret(jwtSecretBytes); err != nil {
		return false, fmt.Errorf("invalid jwt-secret in secret during reload: %w", err)
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

	// Notify listeners about reload
	if s.OnReload != nil {
		// We need to pass the token that is currently effective.
		// If we used tokenOverride, it's that.
		// If not, we need to extract it from the secret we just read?
		// Actually, standard Reload (without override) reads secret from disk/api.
		// But in this specific setup flow, we usually rely on tokenOverride.
		// If we reloaded from Secret, we can get the token from the secret data.
		tokenToBroadcast := tokenOverride
		if tokenToBroadcast == "" {
			// Extract from secret
			if tokenBytes, ok := secret.Data["service-account-token"]; ok {
				tokenToBroadcast = string(tokenBytes)
			}
		}
		if tokenToBroadcast != "" {
			s.OnReload(tokenToBroadcast)
		}
	}

	return true, nil
}

// Methods LoginHandler, LogoutHandler, MeHandler, ChangePasswordHandler are in http_handlers.go
// Method AuthMiddleware and context keys are in middleware.go
