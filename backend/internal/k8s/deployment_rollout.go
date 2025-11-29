package k8s

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flaucha/DKonsole/backend/internal/permissions"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// RolloutDeployment triggers a rollout/restart of a deployment
// Refactored to use layered architecture: Handler -> Service -> Repository
func (s *Service) RolloutDeployment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Parse HTTP request body
	var req struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Namespace == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing name or namespace")
		return
	}

	// Validate namespace access
	ctx := r.Context()
	canEdit, err := permissions.CanPerformAction(ctx, req.Namespace, "edit")
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check permissions: %v", err))
		return
	}
	if !canEdit {
		utils.ErrorResponse(w, http.StatusForbidden, fmt.Sprintf("Edit permission required for namespace: %s", req.Namespace))
		return
	}

	// Validate parameters
	if err := utils.ValidateK8sName(req.Name, "name"); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := utils.ValidateK8sName(req.Namespace, "namespace"); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create service using factory (dependency injection)
	deploymentService := s.serviceFactory.CreateDeploymentService(client)

	// Create context with timeout
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Call service to rollout deployment (business logic layer)
	err = deploymentService.RolloutDeployment(ctx, req.Namespace, req.Name)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to rollout deployment: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, map[string]string{"message": "Deployment rollout triggered successfully"})
}
