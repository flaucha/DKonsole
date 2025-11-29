package ldap

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// Repository defines the interface for LDAP configuration data access
type Repository interface {
	GetConfig(ctx context.Context) (*models.LDAPConfig, error)
	UpdateConfig(ctx context.Context, config *models.LDAPConfig) error
	GetGroups(ctx context.Context) (*models.LDAPGroupsConfig, error)
	UpdateGroups(ctx context.Context, groups *models.LDAPGroupsConfig) error
	GetCredentials(ctx context.Context) (username, password string, err error)
	UpdateCredentials(ctx context.Context, username, password string) error
}

// K8sRepository implements Repository using Kubernetes Secret
type K8sRepository struct {
	client     kubernetes.Interface
	namespace  string
	secretName string
}

// NewRepository creates a new K8sRepository instance
func NewRepository(client kubernetes.Interface, secretName string) *K8sRepository {
	namespace, err := getCurrentNamespace()
	if err != nil {
		// Fallback to default namespace
		namespace = "default"
		utils.LogWarn("Failed to get current namespace, using default", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Use "ldap-config" as the secret name for all LDAP configuration
	ldapSecretName := "ldap-config"

	return &K8sRepository{
		client:     client,
		namespace:  namespace,
		secretName: ldapSecretName,
	}
}

// getCurrentNamespace retrieves the current namespace from the pod's service account
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

// GetConfig retrieves the LDAP configuration from Secret
func (r *K8sRepository) GetConfig(ctx context.Context) (*models.LDAPConfig, error) {
	if r.client == nil {
		// Return default disabled config if no client
		return &models.LDAPConfig{Enabled: false}, nil
	}

	secret, err := r.client.CoreV1().Secrets(r.namespace).Get(ctx, r.secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Secret doesn't exist - return default disabled config
			return &models.LDAPConfig{Enabled: false}, nil
		}
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Get LDAP config from Secret
	configBytes, exists := secret.Data["ldap-config"]
	if !exists || len(configBytes) == 0 {
		return &models.LDAPConfig{Enabled: false}, nil
	}

	var config models.LDAPConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal LDAP config: %w", err)
	}

	return &config, nil
}

// UpdateConfig updates the LDAP configuration in Secret
func (r *K8sRepository) UpdateConfig(ctx context.Context, config *models.LDAPConfig) error {
	if r.client == nil {
		return fmt.Errorf("kubernetes client not available")
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal LDAP config: %w", err)
	}

	// Try to get existing Secret
	secret, err := r.client.CoreV1().Secrets(r.namespace).Get(ctx, r.secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new Secret
			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      r.secretName,
					Namespace: r.namespace,
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					"ldap-config": configJSON,
				},
			}
			_, err = r.client.CoreV1().Secrets(r.namespace).Create(ctx, secret, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create secret: %w", err)
			}
			utils.LogInfo("Created Secret for LDAP config", map[string]interface{}{
				"secret_name": r.secretName,
				"namespace":   r.namespace,
			})
			return nil
		}
		return fmt.Errorf("failed to get secret: %w", err)
	}

	// Update existing Secret
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data["ldap-config"] = configJSON

	_, err = r.client.CoreV1().Secrets(r.namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	utils.LogInfo("Updated LDAP config in Secret", map[string]interface{}{
		"secret_name": r.secretName,
		"namespace":   r.namespace,
		"enabled":     config.Enabled,
	})

	return nil
}

// GetGroups retrieves the LDAP groups configuration from Secret
func (r *K8sRepository) GetGroups(ctx context.Context) (*models.LDAPGroupsConfig, error) {
	if r.client == nil {
		// Return empty groups if no client
		return &models.LDAPGroupsConfig{Groups: []models.LDAPGroup{}}, nil
	}

	secret, err := r.client.CoreV1().Secrets(r.namespace).Get(ctx, r.secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Secret doesn't exist - return empty groups
			return &models.LDAPGroupsConfig{Groups: []models.LDAPGroup{}}, nil
		}
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Get LDAP groups from Secret
	groupsBytes, exists := secret.Data["ldap-groups"]
	if !exists || len(groupsBytes) == 0 {
		return &models.LDAPGroupsConfig{Groups: []models.LDAPGroup{}}, nil
	}

	var groups models.LDAPGroupsConfig
	if err := json.Unmarshal(groupsBytes, &groups); err != nil {
		return nil, fmt.Errorf("failed to unmarshal LDAP groups: %w", err)
	}

	return &groups, nil
}

// UpdateGroups updates the LDAP groups configuration in Secret
func (r *K8sRepository) UpdateGroups(ctx context.Context, groups *models.LDAPGroupsConfig) error {
	if r.client == nil {
		return fmt.Errorf("kubernetes client not available")
	}

	groupsJSON, err := json.Marshal(groups)
	if err != nil {
		return fmt.Errorf("failed to marshal LDAP groups: %w", err)
	}

	// Try to get existing Secret
	secret, err := r.client.CoreV1().Secrets(r.namespace).Get(ctx, r.secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new Secret
			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      r.secretName,
					Namespace: r.namespace,
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					"ldap-groups": groupsJSON,
				},
			}
			_, err = r.client.CoreV1().Secrets(r.namespace).Create(ctx, secret, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create secret: %w", err)
			}
			utils.LogInfo("Created Secret for LDAP groups", map[string]interface{}{
				"secret_name": r.secretName,
				"namespace":   r.namespace,
			})
			return nil
		}
		return fmt.Errorf("failed to get secret: %w", err)
	}

	// Update existing Secret
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data["ldap-groups"] = groupsJSON

	_, err = r.client.CoreV1().Secrets(r.namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	utils.LogInfo("Updated LDAP groups in Secret", map[string]interface{}{
		"secret_name": r.secretName,
		"namespace":   r.namespace,
		"groups_count": len(groups.Groups),
	})

	return nil
}

// GetCredentials retrieves LDAP credentials from Secret
func (r *K8sRepository) GetCredentials(ctx context.Context) (username, password string, err error) {
	if r.client == nil {
		return "", "", fmt.Errorf("kubernetes client not available")
	}

	secret, err := r.client.CoreV1().Secrets(r.namespace).Get(ctx, r.secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", "", nil // No credentials set yet
		}
		return "", "", fmt.Errorf("failed to get secret: %w", err)
	}

	usernameBytes, exists := secret.Data["ldap-username"]
	if !exists {
		return "", "", nil // No username set
	}

	passwordBytes, exists := secret.Data["ldap-password"]
	if !exists {
		return "", "", nil // No password set
	}

	return string(usernameBytes), string(passwordBytes), nil
}

// UpdateCredentials updates LDAP credentials in Secret
func (r *K8sRepository) UpdateCredentials(ctx context.Context, username, password string) error {
	if r.client == nil {
		return fmt.Errorf("kubernetes client not available")
	}

	// Try to get existing Secret
	secret, err := r.client.CoreV1().Secrets(r.namespace).Get(ctx, r.secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new Secret
			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      r.secretName,
					Namespace: r.namespace,
				},
				Type: corev1.SecretTypeOpaque,
				StringData: map[string]string{
					"ldap-username": username,
					"ldap-password": password,
				},
			}
			_, err = r.client.CoreV1().Secrets(r.namespace).Create(ctx, secret, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create secret: %w", err)
			}
			utils.LogInfo("Created Secret for LDAP credentials", map[string]interface{}{
				"secret_name": r.secretName,
				"namespace":   r.namespace,
			})
			return nil
		}
		return fmt.Errorf("failed to get secret: %w", err)
	}

	// Update existing Secret
	// Use StringData to update, which will be converted to Data automatically
	if secret.StringData == nil {
		secret.StringData = make(map[string]string)
	}
	secret.StringData["ldap-username"] = username
	secret.StringData["ldap-password"] = password

	_, err = r.client.CoreV1().Secrets(r.namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	utils.LogInfo("Updated LDAP credentials in Secret", map[string]interface{}{
		"secret_name": r.secretName,
		"namespace":   r.namespace,
	})

	return nil
}
