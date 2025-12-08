package pod

import (
	"fmt"
	"net/http"

	"github.com/flaucha/DKonsole/backend/internal/permissions"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// GetPodEvents handles HTTP GET requests to retrieve Kubernetes events for a specific pod.
// Query parameters:
//   - namespace: The namespace containing the pod
//   - pod: The pod name
//
// Returns a JSON array of events sorted by LastSeen timestamp (most recent first).
func (s *Service) GetPodEvents(w http.ResponseWriter, r *http.Request) {
	// Parse and validate HTTP parameters
	params, err := utils.ParsePodParams(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate namespace access
	ctx := r.Context()
	hasAccess, err := permissions.HasNamespaceAccess(ctx, params.Namespace)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check permissions: %v", err))
		return
	}
	if !hasAccess {
		utils.ErrorResponse(w, http.StatusForbidden, fmt.Sprintf("Access denied to namespace: %s", params.Namespace))
		return
	}

	// Get Kubernetes client for this request
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create repository for this request
	eventRepo := NewK8sEventRepository(client)

	// Create service with repository
	podService := NewPodService(eventRepo)

	// Create context with timeout
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Call service to get events (business logic layer)
	events, err := podService.GetPodEvents(ctx, params.Namespace, params.PodName)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get pod events: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, events)
}
