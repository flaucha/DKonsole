package helm

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

func TestServiceFactory(t *testing.T) {
	client := fake.NewSimpleClientset()
	factory := NewServiceFactory()

	t.Run("CreateHelmReleaseService", func(t *testing.T) {
		service := factory.CreateHelmReleaseService(client)
		if service == nil {
			t.Error("CreateHelmReleaseService returned nil")
		}
	})

	t.Run("CreateHelmInstallService", func(t *testing.T) {
		service := factory.CreateHelmInstallService(client)
		if service == nil {
			t.Error("CreateHelmInstallService returned nil")
		}
	})

	t.Run("CreateHelmUpgradeService", func(t *testing.T) {
		service := factory.CreateHelmUpgradeService(client)
		if service == nil {
			t.Error("CreateHelmUpgradeService returned nil")
		}
	})

	t.Run("CreateHelmJobService", func(t *testing.T) {
		service := factory.CreateHelmJobService(client)
		if service == nil {
			t.Error("CreateHelmJobService returned nil")
		}
	})
}
