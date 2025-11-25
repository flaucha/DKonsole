package helm

import (
	"context"
	"fmt"
)

// HelmInstallService provides business logic for installing Helm releases
type HelmInstallService struct {
	jobService *HelmJobService
}

// NewHelmInstallService creates a new HelmInstallService
func NewHelmInstallService(jobService *HelmJobService) *HelmInstallService {
	return &HelmInstallService{
		jobService: jobService,
	}
}

// InstallHelmReleaseRequest represents the parameters for installing a Helm release
type InstallHelmReleaseRequest struct {
	Name            string
	Namespace       string
	Chart           string
	Version         string
	Repo            string
	ValuesYAML      string
	DkonsoleNS      string
	ServiceAccount  string
}

// InstallHelmReleaseResponse represents the result of initiating an install
type InstallHelmReleaseResponse struct {
	JobName string
	Status  string
	Message string
}

// InstallHelmRelease installs a Helm release by creating a Kubernetes Job
func (s *HelmInstallService) InstallHelmRelease(ctx context.Context, req InstallHelmReleaseRequest) (*InstallHelmReleaseResponse, error) {
	if req.Chart == "" {
		return nil, fmt.Errorf("chart name is required for installation")
	}

	// Create values ConfigMap if needed
	valuesCMName := ""
	if req.ValuesYAML != "" {
		cmName, err := s.jobService.CreateValuesConfigMap(ctx, req.DkonsoleNS, fmt.Sprintf("helm-install-%s", req.Name), req.ValuesYAML)
		if err != nil {
			return nil, fmt.Errorf("failed to create values configmap: %w", err)
		}
		valuesCMName = cmName
	}

	// Create Helm Job
	jobName, err := s.jobService.CreateHelmJob(ctx, CreateHelmJobRequest{
		Operation:          "install",
		ReleaseName:        req.Name,
		Namespace:          req.Namespace,
		ChartName:          req.Chart,
		Version:            req.Version,
		Repo:               req.Repo,
		ValuesYAML:         req.ValuesYAML,
		ValuesCMName:       valuesCMName,
		ServiceAccountName: req.ServiceAccount,
		DkonsoleNamespace:  req.DkonsoleNS,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create helm install job: %w", err)
	}

	return &InstallHelmReleaseResponse{
		JobName: jobName,
		Status:  "install_initiated",
		Message: fmt.Sprintf("Helm install job created: %s", jobName),
	}, nil
}









