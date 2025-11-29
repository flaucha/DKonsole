package auth

import (
	"context"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// K8sUserRepository implements UserRepository using Kubernetes secrets.
// It reads credentials from the "dkonsole-auth" secret in the current namespace.
type K8sUserRepository struct {
	client     kubernetes.Interface
	namespace  string
	secretName string
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
func (r *K8sUserRepository) CreateSecret(ctx context.Context, username, passwordHash, jwtSecret string) error {
	if len(jwtSecret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.secretName,
			Namespace: r.namespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"admin-username":      username,
			"admin-password-hash": passwordHash,
			"jwt-secret":          jwtSecret,
		},
	}

	utils.LogInfo("Creating Kubernetes secret", map[string]interface{}{
		"secret_name": r.secretName,
		"namespace":   r.namespace,
		"username":    username,
	})

	_, err := r.client.CoreV1().Secrets(r.namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		// Check if it's a specific Kubernetes API error
		if apierrors.IsForbidden(err) {
			return fmt.Errorf("forbidden: %w (check RBAC permissions for creating secrets)", err)
		}
		if apierrors.IsAlreadyExists(err) {
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
	if secret.StringData == nil {
		secret.StringData = make(map[string]string)
	}
	secret.StringData["admin-password-hash"] = passwordHash

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
