package helm

import (
	"context"
	"k8s.io/client-go/kubernetes"
)

// HelmReleaseServiceInterface defines the interface for Helm release operations
type HelmReleaseServiceInterface interface {
	GetHelmReleases(ctx context.Context) ([]HelmRelease, error)
	DeleteHelmRelease(ctx context.Context, req DeleteHelmReleaseRequest) (*DeleteHelmReleaseResponse, error)
	GetChartInfo(ctx context.Context, namespace, releaseName string) (*ChartInfo, error)
}

// HelmInstallServiceInterface defines the interface for Helm install operations
type HelmInstallServiceInterface interface {
	InstallHelmRelease(ctx context.Context, req InstallHelmReleaseRequest) (*InstallHelmReleaseResponse, error)
}

// HelmUpgradeServiceInterface defines the interface for Helm upgrade operations
type HelmUpgradeServiceInterface interface {
	UpgradeHelmRelease(ctx context.Context, req UpgradeHelmReleaseRequest) (*UpgradeHelmReleaseResponse, error)
}

// ServiceFactoryInterface defines the interface for creating Helm services
type ServiceFactoryInterface interface {
	CreateHelmReleaseService(client kubernetes.Interface) HelmReleaseServiceInterface
	CreateHelmInstallService(client kubernetes.Interface) HelmInstallServiceInterface
	CreateHelmUpgradeService(client kubernetes.Interface) HelmUpgradeServiceInterface
}
