package k8s

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Factory defines the service factory contract to allow test doubles.
type Factory interface {
	CreateResourceService(dynamicClient dynamic.Interface) *ResourceService
	CreateImportService(dynamicClient dynamic.Interface, client kubernetes.Interface) *ImportService
	CreateNamespaceService(client kubernetes.Interface) *NamespaceService
	CreateClusterStatsService(client kubernetes.Interface) *ClusterStatsService
	CreateDeploymentService(client kubernetes.Interface) *DeploymentService
	CreateCronJobService(client kubernetes.Interface) *CronJobService
	CreateWatchService() *WatchService
}

// ServiceFactory provides factory methods for creating business logic services
type ServiceFactory struct{}

// NewServiceFactory creates a new ServiceFactory
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

// CreateResourceService creates a ResourceService with the given dynamic client
func (f *ServiceFactory) CreateResourceService(dynamicClient dynamic.Interface) *ResourceService {
	resourceRepo := NewK8sResourceRepository(dynamicClient)
	gvrResolver := NewK8sGVRResolver()
	return NewResourceService(resourceRepo, gvrResolver)
}

// CreateImportService creates an ImportService with the given clients
func (f *ServiceFactory) CreateImportService(dynamicClient dynamic.Interface, client kubernetes.Interface) *ImportService {
	resourceRepo := NewK8sResourceRepository(dynamicClient)
	gvrResolver := NewK8sGVRResolver()
	return NewImportService(resourceRepo, gvrResolver, client)
}

// CreateNamespaceService creates a NamespaceService with the given client
func (f *ServiceFactory) CreateNamespaceService(client kubernetes.Interface) *NamespaceService {
	namespaceRepo := NewK8sNamespaceRepository(client)
	return NewNamespaceService(namespaceRepo)
}

// CreateClusterStatsService creates a ClusterStatsService with the given client
func (f *ServiceFactory) CreateClusterStatsService(client kubernetes.Interface) *ClusterStatsService {
	statsRepo := NewK8sClusterStatsRepository(client)
	return NewClusterStatsService(statsRepo)
}

// CreateDeploymentService creates a DeploymentService with the given client
func (f *ServiceFactory) CreateDeploymentService(client kubernetes.Interface) *DeploymentService {
	deploymentRepo := NewK8sDeploymentRepository(client)
	return NewDeploymentService(deploymentRepo)
}

// CreateCronJobService creates a CronJobService with the given client
func (f *ServiceFactory) CreateCronJobService(client kubernetes.Interface) *CronJobService {
	cronJobRepo := NewK8sCronJobRepository(client)
	return NewCronJobService(cronJobRepo)
}

// CreateWatchService creates a WatchService
func (f *ServiceFactory) CreateWatchService() *WatchService {
	gvrResolver := NewK8sGVRResolver()
	return NewWatchService(gvrResolver)
}
