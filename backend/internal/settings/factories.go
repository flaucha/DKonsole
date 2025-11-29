package settings

import (
	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/prometheus"
)

// ServiceFactory creates and configures settings services
type ServiceFactory struct {
	k8sClient        kubernetes.Interface
	handlersModel    *models.Handlers
	secretName       string
	prometheusService *prometheus.HTTPHandler
}

// NewServiceFactory creates a new ServiceFactory
func NewServiceFactory(k8sClient kubernetes.Interface, handlersModel *models.Handlers, secretName string, prometheusService *prometheus.HTTPHandler) *ServiceFactory {
	return &ServiceFactory{
		k8sClient:         k8sClient,
		handlersModel:     handlersModel,
		secretName:        secretName,
		prometheusService: prometheusService,
	}
}

// NewService creates a new settings service
func (f *ServiceFactory) NewService() *Service {
	repo := NewRepository(f.k8sClient, f.secretName)
	return NewService(repo, f.handlersModel, f.prometheusService)
}
