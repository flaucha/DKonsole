package logo

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

const DataDir = "./data"

// LogoStorage defines the interface for logo file storage operations
type LogoStorage interface {
	// Save saves a logo file with the given extension and content
	Save(ctx context.Context, logoType string, ext string, content io.Reader) error
	// Get returns the file path of the logo with the given extension, if it exists
	// For ConfigMap storage, this returns base64-encoded content
	Get(ctx context.Context, logoType string, ext string) (string, error)
	// Remove removes all existing logo files of the specified type
	RemoveAll(ctx context.Context, logoType string) error
	// EnsureDataDir ensures the data directory exists
	EnsureDataDir(ctx context.Context) error
	// GetLogoContent returns the decoded logo content as bytes (for ConfigMap storage)
	// For filesystem storage, this should return nil, nil
	GetLogoContent(ctx context.Context, logoType string, ext string) ([]byte, error)
}

// FileSystemLogoStorage implements LogoStorage using the local filesystem
type FileSystemLogoStorage struct {
	dataDir string
}

// NewFileSystemLogoStorage creates a new FileSystemLogoStorage
func NewFileSystemLogoStorage(dataDir string) *FileSystemLogoStorage {
	return &FileSystemLogoStorage{
		dataDir: dataDir,
	}
}

// EnsureDataDir ensures the data directory exists
func (s *FileSystemLogoStorage) EnsureDataDir(ctx context.Context) error {
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	return nil
}

// Save saves a logo file with the given extension and content
func (s *FileSystemLogoStorage) Save(ctx context.Context, logoType string, ext string, content io.Reader) error {
	var filename string
	if logoType == "light" {
		filename = "logo-light" + ext
	} else {
		filename = "logo" + ext
	}
	destPath := filepath.Join(s.dataDir, filename)
	absPath, _ := filepath.Abs(destPath)

	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, content); err != nil {
		return fmt.Errorf("failed to save file content: %w", err)
	}

	utils.LogInfo("Saving logo", map[string]interface{}{
		"type": logoType,
		"path": absPath,
	})
	return nil
}

// Get returns the file path of the logo with the given extension, if it exists
func (s *FileSystemLogoStorage) Get(ctx context.Context, logoType string, ext string) (string, error) {
	var filename string
	if logoType == "light" {
		filename = "logo-light" + ext
	} else {
		filename = "logo" + ext
	}
	path := filepath.Join(s.dataDir, filename)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("logo not found")
		}
		return "", fmt.Errorf("failed to check logo file: %w", err)
	}
	return path, nil
}

// RemoveAll removes all existing logo files of the specified type
func (s *FileSystemLogoStorage) RemoveAll(ctx context.Context, logoType string) error {
	extensions := []string{".png", ".svg"}
	var filenamePrefix string
	if logoType == "light" {
		filenamePrefix = "logo-light"
	} else {
		filenamePrefix = "logo"
	}
	for _, ext := range extensions {
		path := filepath.Join(s.dataDir, filenamePrefix+ext)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			// Log but don't fail if file doesn't exist
			utils.LogWarn("Failed to remove logo file", map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			})
		}
	}
	return nil
}

// GetLogoContent returns nil for filesystem storage (not used)
func (s *FileSystemLogoStorage) GetLogoContent(ctx context.Context, logoType string, ext string) ([]byte, error) {
	return nil, nil // Filesystem storage doesn't use this method
}
