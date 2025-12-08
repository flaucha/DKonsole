package ldap

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

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
		"secret_name":  r.secretName,
		"namespace":    r.namespace,
		"groups_count": len(groups.Groups),
	})

	return nil
}
