package helm

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// HelmUpgradeService provides business logic for upgrading Helm releases
type HelmUpgradeService struct {
	releaseRepo HelmReleaseRepository
	jobService  *HelmJobService
}

// NewHelmUpgradeService creates a new HelmUpgradeService
func NewHelmUpgradeService(releaseRepo HelmReleaseRepository, jobService *HelmJobService) *HelmUpgradeService {
	return &HelmUpgradeService{
		releaseRepo: releaseRepo,
		jobService:  jobService,
	}
}

// UpgradeHelmReleaseRequest represents the parameters for upgrading a Helm release
type UpgradeHelmReleaseRequest struct {
	Name           string
	Namespace      string
	Chart          string
	Version        string
	Repo           string
	ValuesYAML     string
	DkonsoleNS     string
	ServiceAccount string
}

// UpgradeHelmReleaseResponse represents the result of initiating an upgrade
type UpgradeHelmReleaseResponse struct {
	JobName string
	Status  string
	Message string
}

// UpgradeHelmRelease upgrades a Helm release by creating a Kubernetes Job
func (s *HelmUpgradeService) UpgradeHelmRelease(ctx context.Context, req UpgradeHelmReleaseRequest) (*UpgradeHelmReleaseResponse, error) {
	chartName := req.Chart
	existingRepo := req.Repo

	// If chart not specified, get it from the existing release
	if chartName == "" || req.Repo == "" {
		chartInfo, err := s.getChartInfoFromRelease(ctx, req.Namespace, req.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get chart info from release: %w", err)
		}
		if chartInfo.ChartName == "" {
			return nil, fmt.Errorf("chart name is required for upgrade. Could not determine chart from existing release")
		}
		if chartName == "" {
			chartName = chartInfo.ChartName
		}
		if existingRepo == "" {
			existingRepo = chartInfo.Repo
		}
	}

	// Create values ConfigMap if needed
	valuesCMName := ""
	if req.ValuesYAML != "" {
		cmName, err := s.jobService.CreateValuesConfigMap(ctx, req.DkonsoleNS, fmt.Sprintf("helm-upgrade-%s", req.Name), req.ValuesYAML)
		if err != nil {
			return nil, fmt.Errorf("failed to create values configmap: %w", err)
		}
		valuesCMName = cmName
	}

	// Create Helm Job
	jobName, err := s.jobService.CreateHelmJob(ctx, CreateHelmJobRequest{
		Operation:          "upgrade",
		ReleaseName:        req.Name,
		Namespace:          req.Namespace,
		ChartName:          chartName,
		Version:            req.Version,
		Repo:               existingRepo,
		ValuesYAML:         req.ValuesYAML,
		ValuesCMName:       valuesCMName,
		ServiceAccountName: req.ServiceAccount,
		DkonsoleNamespace:  req.DkonsoleNS,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create helm upgrade job: %w", err)
	}

	return &UpgradeHelmReleaseResponse{
		JobName: jobName,
		Status:  "upgrade_initiated",
		Message: fmt.Sprintf("Helm upgrade job created: %s", jobName),
	}, nil
}

// ChartInfo contains chart information extracted from a Helm release
type ChartInfo struct {
	ChartName string
	Repo      string
}

// getChartInfoFromRelease extracts chart information from an existing Helm release
func (s *HelmUpgradeService) getChartInfoFromRelease(ctx context.Context, namespace, releaseName string) (*ChartInfo, error) {
	secrets, err := s.releaseRepo.ListSecretsInNamespace(ctx, namespace)
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

	// Use HelmReleaseService to decode release data
	releaseService := NewHelmReleaseService(s.releaseRepo)
	releaseData, ok := latestSecret.Data["release"]
	if !ok {
		return &ChartInfo{}, nil
	}

	// Reuse the decode logic from HelmReleaseService
	releaseInfo, err := releaseService.DecodeHelmReleaseData(releaseData)
	if err != nil {
		// If decoding fails, fall back to labels
		chartInfo := &ChartInfo{}
		if name, ok := latestSecret.Labels["name"]; ok {
			chartInfo.ChartName = name
		}
		return chartInfo, nil
	}

	// Extract chart info
	chartInfo := &ChartInfo{}
	if chart, ok := releaseInfo["chart"].(map[string]interface{}); ok {
		if metadata, ok := chart["metadata"].(map[string]interface{}); ok {
			if name, ok := metadata["name"].(string); ok {
				chartInfo.ChartName = name
			}
			if repo, ok := metadata["repository"].(string); ok && repo != "" {
				chartInfo.Repo = repo
			}
		}
		// Try sources if repository not found
		if chartInfo.Repo == "" {
			if sources, ok := chart["sources"].([]interface{}); ok {
				for _, source := range sources {
					if sourceStr, ok := source.(string); ok && sourceStr != "" {
						if strings.HasPrefix(sourceStr, "http://") || strings.HasPrefix(sourceStr, "https://") {
							chartInfo.Repo = sourceStr
							break
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
