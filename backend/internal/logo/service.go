package logo

import (
	"bytes"
	"context"
	"fmt"
	"io"
)

// LogoService provides business logic for logo operations
type LogoService struct {
	validator LogoValidator
	storage   LogoStorage
}

// NewLogoService creates a new LogoService
func NewLogoService(validator LogoValidator, storage LogoStorage) *LogoService {
	return &LogoService{
		validator: validator,
		storage:   storage,
	}
}

// UploadLogoRequest represents parameters for uploading a logo
type UploadLogoRequest struct {
	Filename string
	Size     int64
	Content  io.Reader
	LogoType string // "normal" or "light"
}

// UploadLogo uploads and saves a logo file
func (s *LogoService) UploadLogo(ctx context.Context, req UploadLogoRequest) error {
	// Ensure data directory exists
	if err := s.storage.EnsureDataDir(ctx); err != nil {
		return fmt.Errorf("failed to ensure data directory: %w", err)
	}

	// Read content into memory to allow multiple reads (validation + storage)
	contentBytes, err := io.ReadAll(req.Content)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	// Create readers from the content bytes for validation and storage
	contentReader := bytes.NewReader(contentBytes)
	contentReaderForStorage := bytes.NewReader(contentBytes)

	// Validate file
	ext, err := s.validator.ValidateFile(req.Filename, req.Size, contentReader)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Determine logo type (default to "normal" if not specified)
	logoType := req.LogoType
	if logoType == "" {
		logoType = "normal"
	}

	// Remove existing logos of this type to avoid conflicts
	if err := s.storage.RemoveAll(ctx, logoType); err != nil {
		return fmt.Errorf("failed to remove existing logos: %w", err)
	}

	// Save new logo
	if err := s.storage.Save(ctx, logoType, ext, contentReaderForStorage); err != nil {
		return fmt.Errorf("failed to save logo: %w", err)
	}

	return nil
}

// GetLogoPath returns the path to an existing logo file (for filesystem storage)
// For ConfigMap storage, this returns a special marker
func (s *LogoService) GetLogoPath(ctx context.Context, logoType string) (string, error) {
	extensions := []string{".png", ".svg"}

	// Default to "normal" if not specified
	if logoType == "" {
		logoType = "normal"
	}

	for _, ext := range extensions {
		path, err := s.storage.Get(ctx, logoType, ext)
		if err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("logo not found")
}

// GetLogoContent returns the logo content as bytes (for ConfigMap storage)
func (s *LogoService) GetLogoContent(ctx context.Context, logoType string) ([]byte, string, error) {
	extensions := []string{".png", ".svg"}

	// Default to "normal" if not specified
	if logoType == "" {
		logoType = "normal"
	}

	// Try to get logo content using the interface method
	for _, ext := range extensions {
		content, err := s.storage.GetLogoContent(ctx, logoType, ext)
		if err == nil && content != nil {
			// Determine content type based on extension
			contentType := "image/svg+xml"
			if ext == ".png" {
				contentType = "image/png"
			}
			return content, contentType, nil
		}
	}

	// If GetLogoContent returns nil (filesystem storage), fallback to GetLogoPath
	path, err := s.GetLogoPath(ctx, logoType)
	if err != nil {
		return nil, "", err
	}
	// For filesystem, we return the path and let the HTTP handler serve it
	// This is a marker to indicate filesystem storage
	return []byte(path), "", nil
}
