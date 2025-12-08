package helm

import (
	"k8s.io/client-go/kubernetes"
)

// ServiceFactory creates instances of Helm services
type ServiceFactory struct{}

// NewServiceFactory creates a new ServiceFactory
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

// Ensure ServiceFactory implements ServiceFactoryInterface
var _ ServiceFactoryInterface = &ServiceFactory{}

// CreateHelmReleaseService creates a new HelmReleaseService
func (f *ServiceFactory) CreateHelmReleaseService(client kubernetes.Interface) HelmReleaseServiceInterface {
	repo := NewK8sHelmReleaseRepository(client)
	return NewHelmReleaseService(repo)
}

// CreateHelmJobService creates a new HelmJobService
// Note: This is internal dependency, not necessarily exposed via interface but used by others
func (f *ServiceFactory) CreateHelmJobService(client kubernetes.Interface) *HelmJobService {
	repo := NewK8sHelmJobRepository(client)
	return NewHelmJobService(repo)
}

// CreateHelmUpgradeService creates a new HelmUpgradeService
func (f *ServiceFactory) CreateHelmUpgradeService(client kubernetes.Interface) HelmUpgradeServiceInterface {
	releaseService := f.CreateHelmReleaseService(client)
	jobService := f.CreateHelmJobService(client)
	return NewHelmUpgradeService(releaseService, jobService)
}

// CreateHelmInstallService creates a new HelmInstallService
func (f *ServiceFactory) CreateHelmInstallService(client kubernetes.Interface) HelmInstallServiceInterface {
	jobService := f.CreateHelmJobService(client)
	return NewHelmInstallService(jobService)
}
