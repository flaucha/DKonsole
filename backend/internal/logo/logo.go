package logo

import (
	"net/http"
	"path/filepath"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// Service provides HTTP handlers for logo operations
type Service struct {
	logoService *LogoService
	dataDir     string
}

// NewService creates a new logo service
func NewService(dataDir string) *Service {
	// Create dependencies
	validator := NewDefaultLogoValidator(5 << 20) // 5MB max size
	storage := NewFileSystemLogoStorage(dataDir)
	logoService := NewLogoService(validator, storage)

	return &Service{
		logoService: logoService,
		dataDir:     dataDir,
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

	// Call service to get logo path (business logic layer)
	logoPath, err := s.logoService.GetLogoPath(ctx, logoType)
	if err != nil {
		// Logo not found - return 404
		// Frontend will handle this gracefully and use default logo
		http.Error(w, "Logo not found", http.StatusNotFound)
		return
	}

	absPath, _ := filepath.Abs(logoPath)
	utils.LogDebug("Serving logo", map[string]interface{}{
		"type": logoType,
		"path": absPath,
	})
	http.ServeFile(w, r, logoPath)
}
