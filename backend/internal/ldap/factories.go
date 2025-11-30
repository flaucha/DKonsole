package ldap

import (
	"k8s.io/client-go/kubernetes"
)

// ServiceFactory creates and configures LDAP services
type ServiceFactory struct {
	k8sClient kubernetes.Interface
	secretName string
}

// NewServiceFactory creates a new ServiceFactory
func NewServiceFactory(k8sClient kubernetes.Interface, secretName string) *ServiceFactory {
	return &ServiceFactory{
		k8sClient:  k8sClient,
		secretName: secretName,
	}
}

// NewService creates a new LDAP service
func (f *ServiceFactory) NewService() *Service {
	repo := NewRepository(f.k8sClient, f.secretName)
	return NewService(repo)
}
