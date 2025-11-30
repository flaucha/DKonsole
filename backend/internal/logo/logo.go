package logo

import (
	"net/http"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// Service provides HTTP handlers for logo operations
type Service struct {
	logoService *LogoService
	dataDir     string
}

// NewService creates a new logo service
// If client is provided, uses ConfigMap storage; otherwise uses filesystem storage
func NewService(client kubernetes.Interface, namespace string) *Service {
	// Create dependencies
	validator := NewDefaultLogoValidator(5 << 20) // 5MB max size

	var storage LogoStorage
	if client != nil {
		// Use ConfigMap storage
		storage = NewConfigMapLogoStorage(client, namespace)
	} else {
		// Fallback to filesystem storage
		storage = NewFileSystemLogoStorage("./data")
	}

	logoService := NewLogoService(validator, storage)

	return &Service{
		logoService: logoService,
		dataDir:     "./data", // Keep for backward compatibility
	}
}

// UploadLogo handles logo file uploads
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Storage (Data Access)
func (s *Service) UploadLogo(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		utils.LogError(err, "Error parsing multipart form", nil)
		utils.ErrorResponse(w, http.StatusBadRequest, "Error parsing form")
		return
	}

	file, handler, err := r.FormFile("logo")
	if err != nil {
		utils.LogError(err, "Error retrieving file", nil)
		utils.ErrorResponse(w, http.StatusBadRequest, "Error retrieving file")
		return
	}
	defer file.Close()

	// Get logo type from form (defaults to "normal" if not provided)
	logoType := r.FormValue("type")
	if logoType == "" {
		logoType = "normal"
	}
	if logoType != "normal" && logoType != "light" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid logo type. Must be 'normal' or 'light'")
		return
	}

	// Prepare request
	uploadReq := UploadLogoRequest{
		Filename: handler.Filename,
		Size:     handler.Size,
		Content:  file,
		LogoType: logoType,
	}

	// Call service to upload logo (business logic layer)
	ctx := r.Context()
	if err := s.logoService.UploadLogo(ctx, uploadReq); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.LogDebug("Logo uploaded successfully", map[string]interface{}{
		"filename": handler.Filename,
		"size":     handler.Size,
		"type":     logoType,
		"mime":     handler.Header,
	})

	// Write success response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Logo uploaded successfully",
	})
}

// GetLogo serves the custom logo file
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Storage (Data Access)
func (s *Service) GetLogo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get logo type from query parameter (defaults to "normal" if not provided)
	logoType := r.URL.Query().Get("type")
	if logoType == "" {
		logoType = "normal"
	}
	if logoType != "normal" && logoType != "light" {
		http.Error(w, "Invalid logo type", http.StatusBadRequest)
		return
	}

	// Try to get logo content (works for both ConfigMap and filesystem storage)
	content, contentType, err := s.logoService.GetLogoContent(ctx, logoType)
	if err != nil {
		// Logo not found - try to serve default logo from static directory
		var staticLogoPath string
		if logoType == "light" {
			staticLogoPath = filepath.Join("static", "logo-full-light.svg")
		} else {
			staticLogoPath = filepath.Join("static", "logo-full-dark.svg")
		}

		if absPath, err := filepath.Abs(staticLogoPath); err == nil {
			if _, err := os.Stat(staticLogoPath); err == nil {
				utils.LogDebug("Serving default logo", map[string]interface{}{
					"path": absPath,
					"type": logoType,
				})
				http.ServeFile(w, r, staticLogoPath)
				return
			}
		}
		// Logo not found and no default available - return 404
		// Frontend will handle this gracefully and use default logo
		http.Error(w, "Logo not found", http.StatusNotFound)
		return
	}

	// If contentType is empty, it means we got a filesystem path (backward compatibility)
	if contentType == "" {
		// This is a filesystem path, serve it as a file
		logoPath := string(content)
		absPath, _ := filepath.Abs(logoPath)
		utils.LogDebug("Serving logo from filesystem", map[string]interface{}{
			"type": logoType,
			"path": absPath,
		})
		http.ServeFile(w, r, logoPath)
		return
	}

	// This is ConfigMap storage - serve content from memory
	utils.LogDebug("Serving logo from ConfigMap", map[string]interface{}{
		"type":        logoType,
		"contentType": contentType,
		"size":        len(content),
	})
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}
