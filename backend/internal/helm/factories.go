package helm

import (
	"k8s.io/client-go/kubernetes"
)

// ServiceFactory provides factory methods for creating business logic services
type ServiceFactory struct{}

// NewServiceFactory creates a new ServiceFactory
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

// CreateHelmReleaseService creates a HelmReleaseService with the given client
func (f *ServiceFactory) CreateHelmReleaseService(client kubernetes.Interface) *HelmReleaseService {
	helmRepo := NewK8sHelmReleaseRepository(client)
	return NewHelmReleaseService(helmRepo)
}

// CreateHelmJobService creates a HelmJobService with the given client
func (f *ServiceFactory) CreateHelmJobService(client kubernetes.Interface) *HelmJobService {
	jobRepo := NewK8sHelmJobRepository(client)
	return NewHelmJobService(jobRepo)
}

// CreateHelmUpgradeService creates a HelmUpgradeService with the given client
func (f *ServiceFactory) CreateHelmUpgradeService(client kubernetes.Interface) *HelmUpgradeService {
	releaseRepo := NewK8sHelmReleaseRepository(client)
	jobService := f.CreateHelmJobService(client)
	return NewHelmUpgradeService(releaseRepo, jobService)
}

// CreateHelmInstallService creates a HelmInstallService with the given client
func (f *ServiceFactory) CreateHelmInstallService(client kubernetes.Interface) *HelmInstallService {
	jobService := f.CreateHelmJobService(client)
	return NewHelmInstallService(jobService)
}

