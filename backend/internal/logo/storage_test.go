package logo

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileSystemLogoStorage_EnsureDataDir(t *testing.T) {
	// Create temporary directory for tests
	tmpDir := t.TempDir()
	storage := NewFileSystemLogoStorage(tmpDir)

	ctx := context.Background()

	err := storage.EnsureDataDir(ctx)
	if err != nil {
		t.Errorf("EnsureDataDir() error = %v, want nil", err)
	}

	// Verify directory exists
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Errorf("EnsureDataDir() directory was not created")
	}
}

func TestFileSystemLogoStorage_Save(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemLogoStorage(tmpDir)
	ctx := context.Background()

	// Ensure directory exists
	if err := storage.EnsureDataDir(ctx); err != nil {
		t.Fatalf("Failed to ensure data directory: %v", err)
	}

	tests := []struct {
		name        string
		logoType    string
		ext         string
		content     string
		wantErr     bool
		expectedFile string
	}{
		{
			name:        "save normal logo PNG",
			logoType:    "normal",
			ext:         ".png",
			content:     "fake png content",
			wantErr:     false,
			expectedFile: "logo.png",
		},
		{
			name:        "save light logo PNG",
			logoType:    "light",
			ext:         ".png",
			content:     "fake png content",
			wantErr:     false,
			expectedFile: "logo-light.png",
		},
		{
			name:        "save normal logo SVG",
			logoType:    "normal",
			ext:         ".svg",
			content:     "<svg></svg>",
			wantErr:     false,
			expectedFile: "logo.svg",
		},
		{
			name:        "save light logo SVG",
			logoType:    "light",
			ext:         ".svg",
			content:     "<svg></svg>",
			wantErr:     false,
			expectedFile: "logo-light.svg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentReader := bytes.NewReader([]byte(tt.content))

			err := storage.Save(ctx, tt.logoType, tt.ext, contentReader)

			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was created
				expectedPath := filepath.Join(tmpDir, tt.expectedFile)
				if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
					t.Errorf("Save() file was not created at %v", expectedPath)
					return
				}

				// Verify file content
				fileContent, err := os.ReadFile(expectedPath)
				if err != nil {
					t.Errorf("Save() failed to read saved file: %v", err)
					return
				}

				if string(fileContent) != tt.content {
					t.Errorf("Save() file content = %v, want %v", string(fileContent), tt.content)
				}
			}
		})
	}
}

func TestFileSystemLogoStorage_Get(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemLogoStorage(tmpDir)
	ctx := context.Background()

	// Ensure directory exists
	if err := storage.EnsureDataDir(ctx); err != nil {
		t.Fatalf("Failed to ensure data directory: %v", err)
	}

	tests := []struct {
		name        string
		setupFile   bool
		logoType    string
		ext         string
		filename    string
		wantErr     bool
		errMsg      string
	}{
		{
			name:      "get existing normal logo PNG",
			setupFile: true,
			logoType:  "normal",
			ext:       ".png",
			filename:  "logo.png",
			wantErr:   false,
		},
		{
			name:      "get existing light logo SVG",
			setupFile: true,
			logoType:  "light",
			ext:       ".svg",
			filename:  "logo-light.svg",
			wantErr:   false,
		},
		{
			name:      "get non-existent logo",
			setupFile: false,
			logoType:  "normal",
			ext:       ".png",
			wantErr:   true,
			errMsg:    "logo not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a separate temp dir for each test to avoid interference
			testDir := t.TempDir()
			testStorage := NewFileSystemLogoStorage(testDir)

			// Ensure directory exists
			if err := testStorage.EnsureDataDir(ctx); err != nil {
				t.Fatalf("Failed to ensure data directory: %v", err)
			}

			// Setup: create file if needed
			if tt.setupFile {
				filePath := filepath.Join(testDir, tt.filename)
				if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			path, err := testStorage.Get(ctx, tt.logoType, tt.ext)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("Get() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Get() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			// Verify path is correct
			expectedPath := filepath.Join(testDir, tt.filename)
			if path != expectedPath {
				t.Errorf("Get() path = %v, want %v", path, expectedPath)
			}
		})
	}
}

func TestFileSystemLogoStorage_RemoveAll(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileSystemLogoStorage(tmpDir)
	ctx := context.Background()

	// Ensure directory exists
	if err := storage.EnsureDataDir(ctx); err != nil {
		t.Fatalf("Failed to ensure data directory: %v", err)
	}

	tests := []struct {
		name      string
		setupFiles []string
		logoType  string
		wantErr   bool
	}{
		{
			name:      "remove all normal logos",
			setupFiles: []string{"logo.png", "logo.svg"},
			logoType:  "normal",
			wantErr:   false,
		},
		{
			name:      "remove all light logos",
			setupFiles: []string{"logo-light.png", "logo-light.svg"},
			logoType:  "light",
			wantErr:   false,
		},
		{
			name:      "remove when no files exist (should not error)",
			setupFiles: []string{},
			logoType:  "normal",
			wantErr:   false,
		},
		{
			name:      "remove only removes files of correct type",
			setupFiles: []string{"logo.png", "logo-light.png"},
			logoType:  "normal",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: create files
			for _, filename := range tt.setupFiles {
				filePath := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			err := storage.RemoveAll(ctx, tt.logoType)

			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify files were removed (only for the specified type)
			expectedRemoved := []string{}
			if tt.logoType == "normal" {
				expectedRemoved = []string{"logo.png", "logo.svg"}
			} else {
				expectedRemoved = []string{"logo-light.png", "logo-light.svg"}
			}

			for _, filename := range expectedRemoved {
				filePath := filepath.Join(tmpDir, filename)
				if _, err := os.Stat(filePath); !os.IsNotExist(err) {
					t.Errorf("RemoveAll() file %v should have been removed", filename)
				}
			}

			// Verify other files still exist
			for _, filename := range tt.setupFiles {
				shouldExist := true
				for _, removed := range expectedRemoved {
					if filename == removed {
						shouldExist = false
						break
					}
				}

				filePath := filepath.Join(tmpDir, filename)
				_, err := os.Stat(filePath)
				exists := !os.IsNotExist(err)

				if shouldExist && !exists {
					t.Errorf("RemoveAll() file %v should still exist", filename)
				} else if !shouldExist && exists {
					t.Errorf("RemoveAll() file %v should have been removed", filename)
				}
			}
		})
	}
}
