package k8s

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/example/k8s-view/internal/utils"
)

// TriggerCronJob triggers a CronJob manually
// Refactored to use layered architecture: Handler -> Service -> Repository
func (s *Service) TriggerCronJob(w http.ResponseWriter, r *http.Request) {
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
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create repository
	cronJobRepo := NewK8sCronJobRepository(client)

	// Create service
	cronJobService := NewCronJobService(cronJobRepo)

	// Create context
	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Call service to trigger cronjob (business logic layer)
	jobName, err := cronJobService.TriggerCronJob(ctx, req.Namespace, req.Name)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to trigger cronjob: %v", err))
		return
	}

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusCreated, map[string]string{"jobName": jobName})
}

