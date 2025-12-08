package helm

import (
	"context"
	"fmt"
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

// ChartInfo contains chart information extracted from a Helm release
type ChartInfo struct {
	ChartName string
	Repo      string
}

// GetChartInfo extracts chart information from an existing Helm release
func (s *HelmReleaseService) GetChartInfo(ctx context.Context, namespace, releaseName string) (*ChartInfo, error) {
	secrets, err := s.repo.ListSecretsInNamespace(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	// Find the latest revision secret
	var latestSecret *corev1.Secret
	latestRevision := 0
	for i := range secrets {
		secret := &secrets[i]
		releaseNameFromAnnotation := secret.Annotations["meta.helm.sh/release-name"]

		if releaseNameFromAnnotation == releaseName {
			if revStr, ok := secret.Labels["version"]; ok {
				if rev, err := strconv.Atoi(revStr); err == nil && rev > latestRevision {
					latestRevision = rev
					latestSecret = secret
				}
			}
		}
	}

	if latestSecret == nil {
		return &ChartInfo{}, nil
	}

	var releaseInfo map[string]interface{}
	
	if releaseData, ok := latestSecret.Data["release"]; ok {
		var err error
		releaseInfo, err = s.DecodeHelmReleaseData(releaseData)
		if err != nil {
			// Log error?
			// Proceed to fallback
		}
	}

	// Extract chart info
	chartInfo := &ChartInfo{}
	
	if releaseInfo != nil {
		// This extraction logic ideally should be in extractChartInfo but that returns 3 strings.
		// We can reuse extractChartInfo partially or duplicate extraction for Repo.
		// extractChartInfo currently extracts Name, Version, AppVersion. It DOES NOT extract Repo.
		// So we keep the logic here or enhance extractChartInfo.
		// Keeping logic here for now.
		
		if chart, ok := releaseInfo["chart"].(map[string]interface{}); ok {
			if metadata, ok := chart["metadata"].(map[string]interface{}); ok {
				if name, ok := metadata["name"].(string); ok {
					chartInfo.ChartName = name
				}
				// Repository field in metadata is not standard in all helm versions or charts but we check it.
				// Often repo is not stored in the chart metadata itself if it's from a repo.
				// But let's keep the existing logic.
				if repo, ok := metadata["repository"].(string); ok && repo != "" {
					chartInfo.Repo = repo
				}
			}
			
			// Try sources if repository not found
			if chartInfo.Repo == "" {
				if sources, ok := chart["sources"].([]interface{}); ok {
					for _, source := range sources {
						if sourceStr, ok := source.(string); ok && sourceStr != "" {
							// Basic heuristic for URL
							if len(sourceStr) > 4 && (sourceStr[:7] == "http://" || sourceStr[:8] == "https://") {
								chartInfo.Repo = sourceStr
								break
							}
						}
					}
				}
			}
		}
	}
	
	// Fallback to label if chart name not found
	if chartInfo.ChartName == "" {
		if name, ok := latestSecret.Labels["name"]; ok {
			chartInfo.ChartName = name
		}
	}

	return chartInfo, nil
}

