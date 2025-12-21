package settings

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/flaucha/DKonsole/backend/internal/models"
	"github.com/flaucha/DKonsole/backend/internal/prometheus"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// Service provides business logic for settings operations
type Service struct {
	repo              Repository
	handlersModel     *models.Handlers
	prometheusService *prometheus.HTTPHandler
}

func (s *Service) refreshRepoClient() {
	repo, ok := s.repo.(*K8sRepository)
	if !ok || s.handlersModel == nil {
		return
	}

	s.handlersModel.RLock()
	client := s.handlersModel.Clients["default"]
	s.handlersModel.RUnlock()

	if client != nil {
		repo.client = client
	}
}

// NewService creates a new settings service
func NewService(repo Repository, handlersModel *models.Handlers, prometheusService *prometheus.HTTPHandler) *Service {
	return &Service{
		repo:              repo,
		handlersModel:     handlersModel,
		prometheusService: prometheusService,
	}
}

// UpdatePrometheusURLRequest represents a request to update Prometheus URL
type UpdatePrometheusURLRequest struct {
	URL string `json:"url"`
}

// GetPrometheusURLHandler returns the current Prometheus URL
func (s *Service) GetPrometheusURLHandler(w http.ResponseWriter, r *http.Request) {
	s.refreshRepoClient()
	promURL, err := s.repo.GetPrometheusURL(r.Context())
	if err != nil {
		utils.HandleErrorJSON(w, err, "Failed to get Prometheus URL", http.StatusInternalServerError, nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"url": promURL,
	})
}

// UpdatePrometheusURLHandler updates the Prometheus URL
func (s *Service) UpdatePrometheusURLHandler(w http.ResponseWriter, r *http.Request) {
	s.refreshRepoClient()
	var req UpdatePrometheusURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate URL format
	if req.URL != "" {
		if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
			utils.ErrorResponse(w, http.StatusBadRequest, "invalid URL format. Must start with http:// or https://")
			return
		}
		// Validate URL is parseable
		if _, err := url.Parse(req.URL); err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid URL format: %v", err))
			return
		}
	}

	// Update in repository
	if err := s.repo.UpdatePrometheusURL(r.Context(), req.URL); err != nil {
		utils.HandleErrorJSON(w, err, "Failed to update Prometheus URL", http.StatusInternalServerError, map[string]interface{}{
			"url": req.URL,
		})
		return
	}

	// Update in handlersModel (in-memory)
	s.handlersModel.Lock()
	s.handlersModel.PrometheusURL = req.URL
	s.handlersModel.Unlock()

	// Update Prometheus service with new URL
	if s.prometheusService != nil {
		s.prometheusService.UpdateURL(req.URL)
	}

	utils.LogInfo("Prometheus URL updated", map[string]interface{}{
		"url": req.URL,
	})

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Prometheus URL updated successfully",
		"url":     req.URL,
	})
}
