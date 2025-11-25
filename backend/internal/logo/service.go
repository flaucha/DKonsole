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

	// Remove existing logos to avoid conflicts
	if err := s.storage.RemoveAll(ctx); err != nil {
		return fmt.Errorf("failed to remove existing logos: %w", err)
	}

	// Save new logo
	if err := s.storage.Save(ctx, ext, contentReaderForStorage); err != nil {
		return fmt.Errorf("failed to save logo: %w", err)
	}

	return nil
}

// GetLogoPath returns the path to an existing logo file
func (s *LogoService) GetLogoPath(ctx context.Context) (string, error) {
	extensions := []string{".png", ".svg"}

	for _, ext := range extensions {
		path, err := s.storage.Get(ctx, ext)
		if err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("logo not found")
}
