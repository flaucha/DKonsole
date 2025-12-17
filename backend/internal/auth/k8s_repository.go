package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

var (
	inClusterConfigFunc = rest.InClusterConfig
	newForConfigFunc    = kubernetes.NewForConfig
)

// K8sUserRepository implements UserRepository using Kubernetes secrets.
// It reads credentials from the "dkonsole-auth" secret in the current namespace.
type K8sUserRepository struct {
	client        kubernetes.Interface
	namespace     string
	secretName    string
	ClientFactory func(token string) (kubernetes.Interface, error) // Optional factory for tests
}

// NewK8sUserRepository creates a new K8sUserRepository instance.
// It automatically detects the namespace from the pod's service account.
func NewK8sUserRepository(client kubernetes.Interface, secretName string) (*K8sUserRepository, error) {
	namespace, err := getCurrentNamespace()
	if err != nil {
		return nil, fmt.Errorf("failed to get current namespace: %w", err)
	}

	return &K8sUserRepository{
		client:     client,
		namespace:  namespace,
		secretName: secretName,
	}, nil
}

// getCurrentNamespace retrieves the current namespace from the pod's service account.
// It checks the service account namespace file first, then falls back to environment variable.
func getCurrentNamespace() (string, error) {
	// Try reading from service account namespace file (standard in Kubernetes pods)
	nsFile := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	if data, err := os.ReadFile(nsFile); err == nil {
		namespace := string(data)
		if namespace != "" {
			return namespace, nil
		}
	}

	// Fallback to environment variable
	if namespace := os.Getenv("POD_NAMESPACE"); namespace != "" {
		return namespace, nil
	}

	return "", fmt.Errorf("could not determine namespace: service account file not found and POD_NAMESPACE not set")
}

// GetAdminUser retrieves the admin username from the Kubernetes secret.
// Returns ErrAdminUserNotSet if the secret or key does not exist.
func (r *K8sUserRepository) GetAdminUser() (string, error) {
	secret, err := r.client.CoreV1().Secrets(r.namespace).Get(context.Background(), r.secretName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	username, exists := secret.Data["admin-username"]
	if !exists || len(username) == 0 {
		return "", ErrAdminUserNotSet
	}

	return string(username), nil
}

// GetAdminPasswordHash retrieves the admin password hash from the Kubernetes secret.
// Returns ErrAdminPasswordNotSet if the secret or key does not exist.
func (r *K8sUserRepository) GetAdminPasswordHash() (string, error) {
	secret, err := r.client.CoreV1().Secrets(r.namespace).Get(context.Background(), r.secretName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	passwordHash, exists := secret.Data["admin-password-hash"]
	if !exists || len(passwordHash) == 0 {
		return "", ErrAdminPasswordNotSet
	}

	return string(passwordHash), nil
}

// SecretExists checks if the dkonsole-auth secret exists in the namespace.
// Returns (true, nil) if secret exists
// Returns (false, nil) if secret doesn't exist (NotFound error)
// Returns (false, error) if there's a permission error or other API error
func (r *K8sUserRepository) SecretExists(ctx context.Context) (bool, error) {
	if r.client == nil {
		// No client means we haven't authenticated yet, so secret "doesn't exist" from our perspective
		return false, nil
	}
	_, err := r.client.CoreV1().Secrets(r.namespace).Get(ctx, r.secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Secret doesn't exist - this is expected in setup mode
			return false, nil
		}
		// Other error (permission denied, etc.) - return the error
		return false, fmt.Errorf("failed to check secret existence: %w", err)
	}
	return true, nil
}

// CreateSecret creates the dkonsole-auth secret with the provided credentials.
// username: admin username
// passwordHash: Argon2 password hash
// jwtSecret: JWT secret key (must be at least 32 characters)
// token: Service Account Token for K8s authentication
func (r *K8sUserRepository) CreateSecret(ctx context.Context, username, passwordHash, jwtSecret, token string) error {
	if len(jwtSecret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	// Determine which client to use
	var client kubernetes.Interface
	if token != "" {
		if r.ClientFactory != nil {
			// Use injected factory (for tests)
			c, err := r.ClientFactory(token)
			if err != nil {
				return fmt.Errorf("failed to create ephemeral client from factory: %w", err)
			}
			client = c
		} else {
			// Create real ephemeral client
			c, err := createEphemeralClient(token)
			if err != nil {
				return fmt.Errorf("failed to create ephemeral client: %w", err)
			}
			client = c
		}
	} else {
		if r.client == nil {
			return fmt.Errorf("K8s repository not initialized")
		}
		client = r.client
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.secretName,
			Namespace: r.namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"admin-username":        []byte(username),
			"admin-password-hash":   []byte(passwordHash),
			"jwt-secret":            []byte(jwtSecret),
			"service-account-token": []byte(token),
		},
	}

	utils.LogInfo("Creating Kubernetes secret", map[string]interface{}{
		"secret_name": r.secretName,
		"namespace":   r.namespace,
		"username":    username,
		"using_token": token != "",
	})

	_, err := client.CoreV1().Secrets(r.namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		// Check if it's a specific Kubernetes API error
		if apierrors.IsForbidden(err) {
			return fmt.Errorf("forbidden: %w (check RBAC permissions for creating secrets)", err)
		}
		if apierrors.IsAlreadyExists(err) {
			// If it exists, try to update it? Or return error?
			// The original code returned error. But now with token, we might want to update.
			// However, SetupCompleteHandler checks if secret exists beforehand.
			return fmt.Errorf("secret already exists: %w", err)
		}
		return fmt.Errorf("failed to create secret: %w", err)
	}

	utils.LogInfo("Secret created successfully", map[string]interface{}{
		"secret_name": r.secretName,
		"namespace":   r.namespace,
	})

	return nil
}

// UpdatePassword updates the admin password hash in the Kubernetes secret
func (r *K8sUserRepository) UpdatePassword(ctx context.Context, passwordHash string) error {
	// Get existing secret
	secret, err := r.client.CoreV1().Secrets(r.namespace).Get(ctx, r.secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	// Update password hash
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data["admin-password-hash"] = []byte(passwordHash)
	secret.StringData = nil

	// Update secret
	_, err = r.client.CoreV1().Secrets(r.namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	utils.LogInfo("Password updated in Kubernetes secret", map[string]interface{}{
		"secret_name": r.secretName,
		"namespace":   r.namespace,
	})

	return nil
}

// UpdateSecretToken updates the service account token in the secret.
// uses the NEW token to perform the update (validating it and ensuring permission).
func (r *K8sUserRepository) UpdateSecretToken(ctx context.Context, newToken string) error {
	var client kubernetes.Interface
	var err error

	if r.ClientFactory != nil {
		client, err = r.ClientFactory(newToken)
	} else {
		client, err = createEphemeralClient(newToken)
	}

	if err != nil {
		return fmt.Errorf("failed to create client with new token: %w", err)
	}

	// Get existing secret
	secret, err := client.CoreV1().Secrets(r.namespace).Get(ctx, r.secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get secret (using new token): %w", err)
	}

	// Update token
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data["service-account-token"] = []byte(newToken)
	// Clear StringData to avoid confusion if it was set
	secret.StringData = nil

	// Update secret
	_, err = client.CoreV1().Secrets(r.namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret with new token: %w", err)
	}

	utils.LogInfo("Token updated in Kubernetes secret", map[string]interface{}{
		"secret_name": r.secretName,
		"namespace":   r.namespace,
	})

	return nil
}

// createEphemeralClient creates a K8s client using the provided token.
func createEphemeralClient(token string) (kubernetes.Interface, error) {
	config, err := inClusterConfigFunc()
	if err != nil {
		host := strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_HOST"))
		port := strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_PORT"))
		if host == "" || port == "" {
			return nil, fmt.Errorf("failed to build in-cluster config and KUBERNETES_SERVICE_HOST/PORT are not set")
		}

		// If InClusterConfig fails (e.g. no SA mounted), manually construct config.
		// Security: verify TLS by default (fail closed). Do not disable verification unless explicitly requested.
		config = &rest.Config{
			Host:            "https://" + host + ":" + port,
			BearerToken:     token,
			TLSClientConfig: rest.TLSClientConfig{Insecure: false},
		}

		// Try to load the cluster CA certificate (preferred).
		if caData, readErr := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"); readErr == nil && len(caData) > 0 {
			config.TLSClientConfig.CAData = caData
		} else if caPEM := strings.TrimSpace(os.Getenv("K8S_CA_PEM")); caPEM != "" {
			// Optional: allow supplying CA PEM explicitly (e.g. for non-standard mounts).
			config.TLSClientConfig.CAData = []byte(caPEM)
		}

		// Explicit dev-only escape hatch. Defaults to secure verification.
		if strings.EqualFold(strings.TrimSpace(os.Getenv("K8S_INSECURE_SKIP_VERIFY")), "true") {
			if os.Getenv("GO_ENV") == "production" {
				return nil, fmt.Errorf("K8S_INSECURE_SKIP_VERIFY is not allowed in production")
			}
			utils.LogWarn("K8S_INSECURE_SKIP_VERIFY enabled - TLS certificate verification is disabled for Kubernetes client", nil)
			config.TLSClientConfig.Insecure = true
			config.TLSClientConfig.CAData = nil
		}

		// Fail closed: if we are not explicitly skipping verification, a CA must be provided.
		if !config.TLSClientConfig.Insecure && len(config.TLSClientConfig.CAData) == 0 {
			return nil, fmt.Errorf("kubernetes CA certificate is required (set K8S_CA_PEM or ensure /var/run/secrets/kubernetes.io/serviceaccount/ca.crt is mounted)")
		}
	} else {
		config.BearerToken = token
		config.BearerTokenFile = ""
	}

	client, newErr := newForConfigFunc(config)
	if newErr != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", newErr)
	}
	if client == nil {
		return nil, errors.New("failed to create kubernetes client: nil client")
	}
	return client, nil
}

// CreateOrUpdateSecret creates or updates the dkonsole-auth secret with the provided credentials.
func (r *K8sUserRepository) CreateOrUpdateSecret(ctx context.Context, username, passwordHash, jwtSecret, token string) error {
	if len(jwtSecret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	// Determine which client to use
	var client kubernetes.Interface
	if token != "" {
		if r.ClientFactory != nil {
			// Use injected factory (for tests)
			c, err := r.ClientFactory(token)
			if err != nil {
				return fmt.Errorf("failed to create ephemeral client from factory: %w", err)
			}
			client = c
		} else {
			// Create real ephemeral client
			c, err := createEphemeralClient(token)
			if err != nil {
				return fmt.Errorf("failed to create ephemeral client: %w", err)
			}
			client = c
		}
	} else {
		if r.client == nil {
			return fmt.Errorf("K8s repository not initialized")
		}
		client = r.client
	}

	// Prepare secret data
	secretData := map[string][]byte{
		"admin-username":        []byte(username),
		"admin-password-hash":   []byte(passwordHash),
		"jwt-secret":            []byte(jwtSecret),
		"service-account-token": []byte(token),
	}

	// Try to get existing secret
	existingSecret, err := client.CoreV1().Secrets(r.namespace).Get(ctx, r.secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new secret
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      r.secretName,
					Namespace: r.namespace,
				},
				Type: corev1.SecretTypeOpaque,
				Data: secretData,
			}

			_, err := client.CoreV1().Secrets(r.namespace).Create(ctx, secret, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create secret: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to check secret existence: %w", err)
	}

	// Update existing secret
	existingSecret.Data = secretData
	existingSecret.StringData = nil // Clear StringData

	_, err = client.CoreV1().Secrets(r.namespace).Update(ctx, existingSecret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return nil
}
