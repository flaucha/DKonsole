package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// Service provides business logic for settings operations
type Service struct {
	repo          Repository
	handlersModel *models.Handlers
}

// NewService creates a new settings service
func NewService(repo Repository, handlersModel *models.Handlers) *Service {
	return &Service{
		repo:          repo,
		handlersModel: handlersModel,
	}
}

// UpdatePrometheusURLRequest represents a request to update Prometheus URL
type UpdatePrometheusURLRequest struct {
	URL string `json:"url"`
}

// GetPrometheusURLHandler returns the current Prometheus URL
func (s *Service) GetPrometheusURLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	url, err := s.repo.GetPrometheusURL(ctx)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get Prometheus URL: %v", err))
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"url": url,
	})
}

// UpdatePrometheusURLHandler updates the Prometheus URL
func (s *Service) UpdatePrometheusURLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req UpdatePrometheusURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate URL format (basic validation)
	if req.URL != "" {
		if !(len(req.URL) > 7 && (req.URL[:7] == "http://" || req.URL[:8] == "https://")) {
			utils.ErrorResponse(w, http.StatusBadRequest, "Invalid URL format. Must start with http:// or https://")
			return
		}
	}

	// Update in repository
	if err := s.repo.UpdatePrometheusURL(ctx, req.URL); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update Prometheus URL: %v", err))
		return
	}

	// Update in handlersModel (in-memory)
	s.handlersModel.Lock()
	s.handlersModel.PrometheusURL = req.URL
	s.handlersModel.Unlock()

	utils.LogInfo("Prometheus URL updated", map[string]interface{}{
		"url": req.URL,
	})

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Prometheus URL updated successfully",
		"url":     req.URL,
	})
}
