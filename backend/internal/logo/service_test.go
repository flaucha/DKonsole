package logo

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

// mockLogoValidator is a mock implementation of LogoValidator
type mockLogoValidator struct {
	validateFileFunc func(filename string, size int64, content io.Reader) (string, error)
}

func (m *mockLogoValidator) ValidateFile(filename string, size int64, content io.Reader) (string, error) {
	if m.validateFileFunc != nil {
		return m.validateFileFunc(filename, size, content)
	}
	return ".png", nil
}

// mockLogoStorage is a mock implementation of LogoStorage
type mockLogoStorage struct {
	ensureDataDirFunc func(ctx context.Context) error
	saveFunc          func(ctx context.Context, logoType string, ext string, content io.Reader) error
	getFunc           func(ctx context.Context, logoType string, ext string) (string, error)
	removeAllFunc     func(ctx context.Context, logoType string) error
}

func (m *mockLogoStorage) EnsureDataDir(ctx context.Context) error {
	if m.ensureDataDirFunc != nil {
		return m.ensureDataDirFunc(ctx)
	}
	return nil
}

func (m *mockLogoStorage) Save(ctx context.Context, logoType string, ext string, content io.Reader) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, logoType, ext, content)
	}
	return nil
}

func (m *mockLogoStorage) Get(ctx context.Context, logoType string, ext string) (string, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, logoType, ext)
	}
	return "", nil
}

func (m *mockLogoStorage) RemoveAll(ctx context.Context, logoType string) error {
	if m.removeAllFunc != nil {
		return m.removeAllFunc(ctx, logoType)
	}
	return nil
}

func TestLogoService_UploadLogo(t *testing.T) {
	validPNGHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	tests := []struct {
		name               string
		request            UploadLogoRequest
		ensureDataDirFunc  func(ctx context.Context) error
		validateFileFunc   func(filename string, size int64, content io.Reader) (string, error)
		removeAllFunc      func(ctx context.Context, logoType string) error
		saveFunc           func(ctx context.Context, logoType string, ext string, content io.Reader) error
		wantErr            bool
		errMsg             string
	}{
		{
			name: "successful upload PNG normal",
			request: UploadLogoRequest{
				Filename: "logo.png",
				Size:     1024,
				Content:  bytes.NewReader(validPNGHeader),
				LogoType: "normal",
			},
			ensureDataDirFunc: func(ctx context.Context) error {
				return nil
			},
			validateFileFunc: func(filename string, size int64, content io.Reader) (string, error) {
				return ".png", nil
			},
			removeAllFunc: func(ctx context.Context, logoType string) error {
				return nil
			},
			saveFunc: func(ctx context.Context, logoType string, ext string, content io.Reader) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "successful upload SVG light",
			request: UploadLogoRequest{
				Filename: "logo.svg",
				Size:     500,
				Content:  bytes.NewReader([]byte("<svg></svg>")),
				LogoType: "light",
			},
			ensureDataDirFunc: func(ctx context.Context) error {
				return nil
			},
			validateFileFunc: func(filename string, size int64, content io.Reader) (string, error) {
				return ".svg", nil
			},
			removeAllFunc: func(ctx context.Context, logoType string) error {
				return nil
			},
			saveFunc: func(ctx context.Context, logoType string, ext string, content io.Reader) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "default to normal logo type",
			request: UploadLogoRequest{
				Filename: "logo.png",
				Size:     1024,
				Content:  bytes.NewReader(validPNGHeader),
				LogoType: "", // Empty should default to "normal"
			},
			ensureDataDirFunc: func(ctx context.Context) error {
				return nil
			},
			validateFileFunc: func(filename string, size int64, content io.Reader) (string, error) {
				return ".png", nil
			},
			removeAllFunc: func(ctx context.Context, logoType string) error {
				// Verify it defaults to "normal"
				if logoType != "normal" {
					t.Errorf("Expected logoType 'normal', got %v", logoType)
				}
				return nil
			},
			saveFunc: func(ctx context.Context, logoType string, ext string, content io.Reader) error {
				if logoType != "normal" {
					t.Errorf("Expected logoType 'normal', got %v", logoType)
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "error ensuring data directory",
			request: UploadLogoRequest{
				Filename: "logo.png",
				Size:     1024,
				Content:  bytes.NewReader(validPNGHeader),
				LogoType: "normal",
			},
			ensureDataDirFunc: func(ctx context.Context) error {
				return io.EOF // Simulate error
			},
			wantErr: true,
			errMsg:  "failed to ensure data directory",
		},
		{
			name: "validation error",
			request: UploadLogoRequest{
				Filename: "logo.png",
				Size:     1024,
				Content:  bytes.NewReader(validPNGHeader),
				LogoType: "normal",
			},
			ensureDataDirFunc: func(ctx context.Context) error {
				return nil
			},
			validateFileFunc: func(filename string, size int64, content io.Reader) (string, error) {
				return "", io.EOF // Simulate validation error
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name: "error removing existing logos",
			request: UploadLogoRequest{
				Filename: "logo.png",
				Size:     1024,
				Content:  bytes.NewReader(validPNGHeader),
				LogoType: "normal",
			},
			ensureDataDirFunc: func(ctx context.Context) error {
				return nil
			},
			validateFileFunc: func(filename string, size int64, content io.Reader) (string, error) {
				return ".png", nil
			},
			removeAllFunc: func(ctx context.Context, logoType string) error {
				return io.EOF // Simulate error
			},
			wantErr: true,
			errMsg:  "failed to remove existing logos",
		},
		{
			name: "error saving logo",
			request: UploadLogoRequest{
				Filename: "logo.png",
				Size:     1024,
				Content:  bytes.NewReader(validPNGHeader),
				LogoType: "normal",
			},
			ensureDataDirFunc: func(ctx context.Context) error {
				return nil
			},
			validateFileFunc: func(filename string, size int64, content io.Reader) (string, error) {
				return ".png", nil
			},
			removeAllFunc: func(ctx context.Context, logoType string) error {
				return nil
			},
			saveFunc: func(ctx context.Context, logoType string, ext string, content io.Reader) error {
				return io.EOF // Simulate error
			},
			wantErr: true,
			errMsg:  "failed to save logo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := &mockLogoValidator{
				validateFileFunc: tt.validateFileFunc,
			}

			mockStorage := &mockLogoStorage{
				ensureDataDirFunc: tt.ensureDataDirFunc,
				removeAllFunc:     tt.removeAllFunc,
				saveFunc:          tt.saveFunc,
			}

			service := NewLogoService(mockValidator, mockStorage)
			ctx := context.Background()

			err := service.UploadLogo(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("UploadLogo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("UploadLogo() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("UploadLogo() error = %v, want containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestLogoService_GetLogoPath(t *testing.T) {
	tests := []struct {
		name        string
		logoType    string
		getFunc     func(ctx context.Context, logoType string, ext string) (string, error)
		wantErr     bool
		wantPath    string
	}{
		{
			name:     "get existing normal logo PNG",
			logoType: "normal",
			getFunc: func(ctx context.Context, logoType string, ext string) (string, error) {
				if ext == ".png" {
					return "/path/to/logo.png", nil
				}
				return "", io.EOF
			},
			wantErr:  false,
			wantPath: "/path/to/logo.png",
		},
		{
			name:     "get existing light logo SVG",
			logoType: "light",
			getFunc: func(ctx context.Context, logoType string, ext string) (string, error) {
				if ext == ".svg" {
					return "/path/to/logo-light.svg", nil
				}
				return "", io.EOF
			},
			wantErr:  false,
			wantPath: "/path/to/logo-light.svg",
		},
		{
			name:     "default to normal when logo type is empty",
			logoType: "",
			getFunc: func(ctx context.Context, logoType string, ext string) (string, error) {
				if logoType == "normal" && ext == ".png" {
					return "/path/to/logo.png", nil
				}
				return "", io.EOF
			},
			wantErr:  false,
			wantPath: "/path/to/logo.png",
		},
		{
			name:     "logo not found",
			logoType: "normal",
			getFunc: func(ctx context.Context, logoType string, ext string) (string, error) {
				return "", io.EOF
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &mockLogoStorage{
				getFunc: tt.getFunc,
			}

			service := NewLogoService(nil, mockStorage)
			ctx := context.Background()

			path, err := service.GetLogoPath(ctx, tt.logoType)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLogoPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if path != tt.wantPath {
					t.Errorf("GetLogoPath() path = %v, want %v", path, tt.wantPath)
				}
			}
		})
	}
}
