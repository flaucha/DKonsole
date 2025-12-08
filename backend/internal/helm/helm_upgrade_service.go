package helm

import (
	"context"
	"fmt"
)

// HelmUpgradeService provides business logic for upgrading Helm releases
type HelmUpgradeService struct {
	releaseService HelmReleaseServiceInterface
	jobService     *HelmJobService
}

// NewHelmUpgradeService creates a new HelmUpgradeService
func NewHelmUpgradeService(releaseService HelmReleaseServiceInterface, jobService *HelmJobService) *HelmUpgradeService {
	return &HelmUpgradeService{
		releaseService: releaseService,
		jobService:     jobService,
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
		chartInfo, err := s.releaseService.GetChartInfo(ctx, req.Namespace, req.Name)
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
