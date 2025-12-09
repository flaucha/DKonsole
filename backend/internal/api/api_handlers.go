package api

import (
	"fmt"
	"net/http"

	"github.com/flaucha/DKonsole/backend/internal/cluster"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// Service provides API resource and CRD operations
type Service struct {
	clusterService *cluster.Service
}

// NewService creates a new API service
func NewService(cs *cluster.Service) *Service {
	return &Service{
		clusterService: cs,
	}
}

// ListAPIResources lists all available API resources in the cluster
// Refactored to use layered architecture: Handler -> Service -> Repository
func (s *Service) ListAPIResources(w http.ResponseWriter, r *http.Request) {
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create repository
	discoveryRepo := NewK8sDiscoveryRepository(client)

	// Create service
	apiService := NewAPIService(discoveryRepo, nil)

	// Create context
	ctx, cancel := utils.CreateRequestContext(r)
	defer cancel()

	// Call service (business logic layer)
	result, err := apiService.ListAPIResources(ctx)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to discover APIs: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, result)
}
