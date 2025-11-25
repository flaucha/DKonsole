package helm

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// HelmRelease represents a Helm release
type HelmRelease struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Chart       string `json:"chart"`
	Version     string `json:"version"`
	Status      string `json:"status"`
	Revision    int    `json:"revision"`
	Updated     string `json:"updated"`
	AppVersion  string `json:"appVersion,omitempty"`
	Description string `json:"description,omitempty"`
}

// HelmReleaseService provides business logic for Helm release operations
type HelmReleaseService struct {
	repo HelmReleaseRepository
}

// NewHelmReleaseService creates a new HelmReleaseService
func NewHelmReleaseService(repo HelmReleaseRepository) *HelmReleaseService {
	return &HelmReleaseService{repo: repo}
}

// GetHelmReleases fetches and transforms all Helm releases from Secrets and ConfigMaps
func (s *HelmReleaseService) GetHelmReleases(ctx context.Context) ([]HelmRelease, error) {
	releasesMap := make(map[string]*HelmRelease)

	// Process Secrets
	secrets, err := s.repo.ListHelmSecrets(ctx)
	if err == nil {
		for _, secret := range secrets {
			if release := s.parseReleaseFromSecret(secret); release != nil {
				key := fmt.Sprintf("%s/%s", release.Namespace, release.Name)
				// Only update if this is a newer revision or doesn't exist
				if existing, exists := releasesMap[key]; !exists ||
					release.Revision > existing.Revision ||
					(release.Revision == existing.Revision && release.Status == "deployed" && existing.Status != "deployed") {
					releasesMap[key] = release
				}
			}
		}
	}

	// Process ConfigMaps
	configMaps, err := s.repo.ListHelmConfigMaps(ctx)
	if err == nil {
		for _, cm := range configMaps {
			releaseName := cm.Labels["name"]
			if releaseName == "" {
				releaseName = cm.Annotations["meta.helm.sh/release-name"]
			}
			releaseNamespace := cm.Namespace
			if releaseNamespace == "" {
				releaseNamespace = cm.Annotations["meta.helm.sh/release-namespace"]
			}

			if releaseName != "" && releaseNamespace != "" {
				key := fmt.Sprintf("%s/%s", releaseNamespace, releaseName)
				if _, exists := releasesMap[key]; !exists {
					release := s.parseReleaseFromConfigMap(cm)
					if release != nil {
						releasesMap[key] = release
					}
				}
			}
		}
	}

	// Convert map to slice
	releases := make([]HelmRelease, 0, len(releasesMap))
	for _, release := range releasesMap {
		releases = append(releases, *release)
	}

	return releases, nil
}

// parseReleaseFromSecret parses a Helm release from a Secret
func (s *HelmReleaseService) parseReleaseFromSecret(secret corev1.Secret) *HelmRelease {
	releaseName := secret.Labels["name"]
	if releaseName == "" {
		releaseName = secret.Annotations["meta.helm.sh/release-name"]
	}
	releaseNamespace := secret.Namespace
	if releaseNamespace == "" {
		releaseNamespace = secret.Annotations["meta.helm.sh/release-namespace"]
	}

	if releaseName == "" || releaseNamespace == "" {
		return nil
	}

	// Parse release data
	if releaseData, ok := secret.Data["release"]; ok {
		releaseInfo, err := s.DecodeHelmReleaseData(releaseData)
		if err != nil {
			return nil
		}

		chartName, chartVersion, appVersion := s.extractChartInfo(releaseInfo)
		status, revision, updated, description := s.extractReleaseInfo(releaseInfo, secret)

		return &HelmRelease{
			Name:        releaseName,
			Namespace:   releaseNamespace,
			Chart:       chartName,
			Version:     chartVersion,
			Status:      status,
			Revision:    revision,
			Updated:     updated,
			AppVersion:  appVersion,
			Description: description,
		}
	}

	return nil
}

// parseReleaseFromConfigMap parses a Helm release from a ConfigMap
func (s *HelmReleaseService) parseReleaseFromConfigMap(cm corev1.ConfigMap) *HelmRelease {
	releaseName := cm.Labels["name"]
	if releaseName == "" {
		releaseName = cm.Annotations["meta.helm.sh/release-name"]
	}
	releaseNamespace := cm.Namespace
	if releaseNamespace == "" {
		releaseNamespace = cm.Annotations["meta.helm.sh/release-namespace"]
	}

	if releaseName == "" || releaseNamespace == "" {
		return nil
	}

	chartName := cm.Labels["name"]
	chartVersion := cm.Labels["version"]
	status := cm.Labels["status"]
	if status == "" {
		status = "unknown"
	}

	revision := 0
	if revStr, ok := cm.Labels["version"]; ok {
		if rev, err := strconv.Atoi(revStr); err == nil {
			revision = rev
		}
	}

	return &HelmRelease{
		Name:      releaseName,
		Namespace: releaseNamespace,
		Chart:     chartName,
		Version:   chartVersion,
		Status:    status,
		Revision:  revision,
		Updated:   cm.CreationTimestamp.Format(time.RFC3339),
	}
}

// DecodeHelmReleaseData decodes and decompresses Helm release data from a Secret (exported for reuse)
func (s *HelmReleaseService) DecodeHelmReleaseData(releaseData []byte) (map[string]interface{}, error) {
	decoded := releaseData

	// Try to decode as base64 string first
	if decodedStr, err := base64.StdEncoding.DecodeString(string(releaseData)); err == nil && len(decodedStr) > 0 {
		decoded = decodedStr
	}

	// Try to decompress gzip
	reader := bytes.NewReader(decoded)
	gzReader, err := gzip.NewReader(reader)
	if err == nil {
		decompressed, err := io.ReadAll(gzReader)
		gzReader.Close()
		if err == nil {
			decoded = decompressed
		}
	}

	// Parse as JSON
	var releaseInfo map[string]interface{}
	if err := json.Unmarshal(decoded, &releaseInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal release JSON: %w", err)
	}

	return releaseInfo, nil
}

// extractChartInfo extracts chart information from release data
func (s *HelmReleaseService) extractChartInfo(releaseInfo map[string]interface{}) (chartName, chartVersion, appVersion string) {
	if chart, ok := releaseInfo["chart"].(map[string]interface{}); ok {
		if metadata, ok := chart["metadata"].(map[string]interface{}); ok {
			if name, ok := metadata["name"].(string); ok {
				chartName = name
			}
			if version, ok := metadata["version"].(string); ok {
				chartVersion = version
			}
			if av, ok := metadata["appVersion"].(string); ok {
				appVersion = av
			}
		}
	}
	return
}

// extractReleaseInfo extracts release information from release data and secret labels
func (s *HelmReleaseService) extractReleaseInfo(releaseInfo map[string]interface{}, secret corev1.Secret) (status string, revision int, updated, description string) {
	// Get status and revision from labels first
	status = secret.Labels["status"]
	if status == "" {
		status = "unknown"
	}

	if revStr, ok := secret.Labels["version"]; ok {
		if rev, err := strconv.Atoi(revStr); err == nil {
			revision = rev
		}
	}

	// Override from release info if available
	if info, ok := releaseInfo["info"].(map[string]interface{}); ok {
		if status == "" || status == "superseded" {
			if s, ok := info["status"].(string); ok {
				status = s
			}
		}

		if revision == 0 {
			if r, ok := info["revision"].(float64); ok {
				revision = int(r)
			}
		}

		if u, ok := info["last_deployed"].(map[string]interface{}); ok {
			if uStr, ok := u["Time"].(string); ok {
				updated = uStr
			} else if uStr, ok := info["last_deployed"].(string); ok {
				updated = uStr
			}
		} else if uStr, ok := info["last_deployed"].(string); ok {
			updated = uStr
		}

		if d, ok := info["description"].(string); ok {
			description = d
		}
	}

	return status, revision, updated, description
}

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
