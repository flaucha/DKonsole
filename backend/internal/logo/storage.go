package logo

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const DataDir = "./data"

// LogoStorage defines the interface for logo file storage operations
type LogoStorage interface {
	// Save saves a logo file with the given extension and content
	Save(ctx context.Context, ext string, content io.Reader) error
	// Get returns the file path of the logo with the given extension, if it exists
	Get(ctx context.Context, ext string) (string, error)
	// Remove removes all existing logo files
	RemoveAll(ctx context.Context) error
	// EnsureDataDir ensures the data directory exists
	EnsureDataDir(ctx context.Context) error
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
func (s *FileSystemLogoStorage) Save(ctx context.Context, ext string, content io.Reader) error {
	destPath := filepath.Join(s.dataDir, "logo"+ext)
	absPath, _ := filepath.Abs(destPath)
	
	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, content); err != nil {
		return fmt.Errorf("failed to save file content: %w", err)
	}

	fmt.Printf("Saving logo to: %s\n", absPath)
	return nil
}

// Get returns the file path of the logo with the given extension, if it exists
func (s *FileSystemLogoStorage) Get(ctx context.Context, ext string) (string, error) {
	path := filepath.Join(s.dataDir, "logo"+ext)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("logo not found")
		}
		return "", fmt.Errorf("failed to check logo file: %w", err)
	}
	return path, nil
}

// RemoveAll removes all existing logo files
func (s *FileSystemLogoStorage) RemoveAll(ctx context.Context) error {
	extensions := []string{".png", ".svg"}
	for _, ext := range extensions {
		path := filepath.Join(s.dataDir, "logo"+ext)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			// Log but don't fail if file doesn't exist
			fmt.Printf("Warning: failed to remove logo file %s: %v\n", path, err)
		}
	}
	return nil
}







