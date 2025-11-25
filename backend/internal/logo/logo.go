package logo

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/example/k8s-view/internal/utils"
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
		fmt.Printf("Error parsing multipart form: %v\n", err)
		utils.ErrorResponse(w, http.StatusBadRequest, "Error parsing form")
		return
	}

	file, handler, err := r.FormFile("logo")
	if err != nil {
		fmt.Printf("Error retrieving file: %v\n", err)
		utils.ErrorResponse(w, http.StatusBadRequest, "Error retrieving file")
		return
	}
	defer file.Close()

	// Prepare request
	uploadReq := UploadLogoRequest{
		Filename: handler.Filename,
		Size:     handler.Size,
		Content:  file,
	}

	// Call service to upload logo (business logic layer)
	ctx := r.Context()
	if err := s.logoService.UploadLogo(ctx, uploadReq); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Write success response (HTTP layer)
	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"status": "success",
		"message": "Logo uploaded successfully",
	})
}

// GetLogo serves the custom logo file
// Refactored to use layered architecture:
// Handler (HTTP) -> Service (Business Logic) -> Storage (Data Access)
func (s *Service) GetLogo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Call service to get logo path (business logic layer)
	logoPath, err := s.logoService.GetLogoPath(ctx)
	if err != nil {
		// Logo not found - return 404
		// Frontend will handle this gracefully and use default logo
		http.Error(w, "Logo not found", http.StatusNotFound)
		return
	}

	absPath, _ := filepath.Abs(logoPath)
	fmt.Printf("Serving logo from: %s\n", absPath)
	http.ServeFile(w, r, logoPath)
}







