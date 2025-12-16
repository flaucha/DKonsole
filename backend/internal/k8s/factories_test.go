package k8s

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestServiceFactory_CreateServices(t *testing.T) {
	factory := NewServiceFactory()
	dynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	k8sClient := k8sfake.NewSimpleClientset()

	if svc := factory.CreateResourceService(dynamicClient, k8sClient); svc == nil {
		t.Fatalf("CreateResourceService returned nil")
	}
	if svc := factory.CreateImportService(dynamicClient, k8sClient); svc == nil {
		t.Fatalf("CreateImportService returned nil")
	}
	if svc := factory.CreateNamespaceService(k8sClient); svc == nil {
		t.Fatalf("CreateNamespaceService returned nil")
	}
	if svc := factory.CreateClusterStatsService(k8sClient); svc == nil {
		t.Fatalf("CreateClusterStatsService returned nil")
	}
	if svc := factory.CreateDeploymentService(k8sClient); svc == nil {
		t.Fatalf("CreateDeploymentService returned nil")
	}
	if svc := factory.CreateCronJobService(k8sClient); svc == nil {
		t.Fatalf("CreateCronJobService returned nil")
	}
	if svc := factory.CreateWatchService(); svc == nil {
		t.Fatalf("CreateWatchService returned nil")
	}
}
