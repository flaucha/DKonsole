package helm

import (
	"fmt"
	"net/http"

	"github.com/example/k8s-view/internal/cluster"
	"github.com/example/k8s-view/internal/models"
	"github.com/example/k8s-view/internal/utils"
)

// Service provides Helm release operations
type Service struct {
	handlers       *models.Handlers
	clusterService *cluster.Service
	serviceFactory *ServiceFactory
}

// NewService creates a new Helm service
func NewService(h *models.Handlers, cs *cluster.Service) *Service {
	return &Service{
		handlers:       h,
		clusterService: cs,
		serviceFactory: NewServiceFactory(),
	}
}

// GetHelmReleases lists all Helm releases in the cluster
// Refactored to use layered architecture: Handler -> Service -> Repository
func (s *Service) GetHelmReleases(w http.ResponseWriter, r *http.Request) {
	// Get Kubernetes client
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create service using factory (dependency injection)
	helmService := s.serviceFactory.CreateHelmReleaseService(client)

	// Create context
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Call service to get Helm releases (business logic layer)
	releases, err := helmService.GetHelmReleases(ctx)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get Helm releases: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, releases)
}

// DeleteHelmRelease uninstalls a Helm release by deleting its Secrets
// Refactored to use layered architecture: Handler -> Service -> Repository
func (s *Service) DeleteHelmRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse HTTP parameters
	releaseName := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")

	if releaseName == "" || namespace == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing name or namespace parameter")
		return
	}

	// Validate parameters
	if err := utils.ValidateK8sName(releaseName, "name"); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := utils.ValidateK8sName(namespace, "namespace"); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get Kubernetes client
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create service using factory (dependency injection)
	helmService := s.serviceFactory.CreateHelmReleaseService(client)

	// Create context
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Prepare request
	deleteReq := DeleteHelmReleaseRequest{
		Name:      releaseName,
		Namespace: namespace,
	}

	// Call service to delete Helm release (business logic layer)
	result, err := helmService.DeleteHelmRelease(ctx, deleteReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Failed to delete Helm release: %v", err))
		return
	}

	// Audit log
	utils.AuditLog(r, "delete", "HelmRelease", releaseName, namespace, true, nil, map[string]interface{}{
		"secrets_deleted": result.SecretsDeleted,
	})

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":          "deleted",
		"secrets_deleted": result.SecretsDeleted,
	})
}
