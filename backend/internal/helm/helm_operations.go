package helm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/example/k8s-view/internal/utils"
)

// UpgradeHelmRelease upgrades a Helm release to a new version
// Refactored to use layered architecture: Handler -> Service -> Repository
func (s *Service) UpgradeHelmRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse HTTP request body
	var req struct {
		Name       string                 `json:"name"`
		Namespace  string                 `json:"namespace"`
		Chart      string                 `json:"chart,omitempty"`
		Version    string                 `json:"version,omitempty"`
		Repo       string                 `json:"repo,omitempty"`
		Values     map[string]interface{} `json:"values,omitempty"`
		ValuesYAML string                 `json:"valuesYaml,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	if req.Name == "" || req.Namespace == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing name or namespace")
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

	// Get Kubernetes clients
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get cluster name
	cluster := r.URL.Query().Get("cluster")
	if cluster == "" {
		cluster = "default"
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Create service using factory (dependency injection)
	upgradeService := s.serviceFactory.CreateHelmUpgradeService(client)

	// Determine dkonsole namespace and service account
	dkonsoleNamespace := "dkonsole"
	saName := "dkonsole"

	// Prepare upgrade request
	upgradeReq := UpgradeHelmReleaseRequest{
		Name:           req.Name,
		Namespace:      req.Namespace,
		Chart:          req.Chart,
		Version:        req.Version,
		Repo:           req.Repo,
		ValuesYAML:     req.ValuesYAML,
		DkonsoleNS:     dkonsoleNamespace,
		ServiceAccount: saName,
	}

	// Call service to upgrade Helm release (business logic layer)
	result, err := upgradeService.UpgradeHelmRelease(ctx, upgradeReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upgrade Helm release: %v", err))
		return
	}

	// Audit log
	utils.AuditLog(r, "upgrade", "HelmRelease", req.Name, req.Namespace, true, nil, map[string]interface{}{
		"chart":   req.Chart,
		"version": req.Version,
		"job":     result.JobName,
	})

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  result.Status,
		"message": result.Message,
		"job":     result.JobName,
	})
}

// InstallHelmRelease installs a new Helm chart
// Refactored to use layered architecture: Handler -> Service -> Repository
func (s *Service) InstallHelmRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse HTTP request body
	var req struct {
		Name       string                 `json:"name"`
		Namespace  string                 `json:"namespace"`
		Chart      string                 `json:"chart"`
		Version    string                 `json:"version,omitempty"`
		Repo       string                 `json:"repo,omitempty"`
		Values     map[string]interface{} `json:"values,omitempty"`
		ValuesYAML string                 `json:"valuesYaml,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	if req.Name == "" || req.Namespace == "" || req.Chart == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing required fields: name, namespace, or chart")
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

	// Get Kubernetes client
	client, err := s.clusterService.GetClient(r)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := utils.CreateTimeoutContext()
	defer cancel()

	// Create service using factory (dependency injection)
	installService := s.serviceFactory.CreateHelmInstallService(client)

	// Determine dkonsole namespace and service account
	dkonsoleNamespace := "dkonsole"
	saName := "dkonsole"

	// Prepare install request
	installReq := InstallHelmReleaseRequest{
		Name:           req.Name,
		Namespace:      req.Namespace,
		Chart:          req.Chart,
		Version:        req.Version,
		Repo:           req.Repo,
		ValuesYAML:     req.ValuesYAML,
		DkonsoleNS:     dkonsoleNamespace,
		ServiceAccount: saName,
	}

	// Call service to install Helm release (business logic layer)
	result, err := installService.InstallHelmRelease(ctx, installReq)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to install Helm release: %v", err))
		return
	}

	// Audit log
	utils.AuditLog(r, "install", "HelmRelease", req.Name, req.Namespace, true, nil, map[string]interface{}{
		"chart":   req.Chart,
		"version": req.Version,
		"job":     result.JobName,
	})

	// Write JSON response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  result.Status,
		"message": result.Message,
		"job":     result.JobName,
	})
}
