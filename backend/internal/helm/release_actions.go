package helm

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// DeleteHelmReleaseRequest represents the parameters for deleting a Helm release
type DeleteHelmReleaseRequest struct {
	Name      string
	Namespace string
}

// DeleteHelmReleaseResponse represents the result of deleting a Helm release
type DeleteHelmReleaseResponse struct {
	SecretsDeleted int
}

// DeleteHelmRelease deletes all Secrets and ConfigMaps related to a Helm release
func (s *HelmReleaseService) DeleteHelmRelease(ctx context.Context, req DeleteHelmReleaseRequest) (*DeleteHelmReleaseResponse, error) {
	deletedCount := 0

	// Find and delete Secrets
	secrets, err := s.repo.ListSecretsInNamespace(ctx, req.Namespace)
	if err == nil {
		for _, secret := range secrets {
			if s.isSecretRelatedToRelease(secret, req.Name, req.Namespace) {
				if err := s.repo.DeleteSecret(ctx, req.Namespace, secret.Name); err == nil {
					deletedCount++
				}
			}
		}
	}

	// Find and delete ConfigMaps
	configMaps, err := s.repo.ListConfigMapsInNamespace(ctx, req.Namespace)
	if err == nil {
		for _, cm := range configMaps {
			releaseNameFromAnnotation := cm.Annotations["meta.helm.sh/release-name"]
			releaseNamespaceFromAnnotation := cm.Annotations["meta.helm.sh/release-namespace"]

			if releaseNameFromAnnotation == req.Name && releaseNamespaceFromAnnotation == req.Namespace {
				if err := s.repo.DeleteConfigMap(ctx, req.Namespace, cm.Name); err == nil {
					deletedCount++
				}
			}
		}
	}

	if deletedCount == 0 {
		return nil, fmt.Errorf("no Helm release secrets found")
	}

	return &DeleteHelmReleaseResponse{
		SecretsDeleted: deletedCount,
	}, nil
}

// isSecretRelatedToRelease checks if a Secret is related to a Helm release
func (s *HelmReleaseService) isSecretRelatedToRelease(secret corev1.Secret, releaseName, namespace string) bool {
	releaseNameFromAnnotation := secret.Annotations["meta.helm.sh/release-name"]
	releaseNamespaceFromAnnotation := secret.Annotations["meta.helm.sh/release-namespace"]

	secretNameMatches := false
	if len(secret.Name) > len(releaseName) {
		prefix := fmt.Sprintf("sh.helm.release.v1.%s.v", releaseName)
		secretNameMatches = len(secret.Name) >= len(prefix) && secret.Name[:len(prefix)] == prefix
	}

	return (releaseNameFromAnnotation == releaseName && releaseNamespaceFromAnnotation == namespace) || secretNameMatches
}
