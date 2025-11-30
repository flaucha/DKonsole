package logo

import (
	"bytes"
	"strings"
	"testing"
)

// Valid PNG file header (first 8 bytes)
var validPNGHeader = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

// Valid SVG content
var validSVGContent = `<?xml version="1.0"?>
<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <circle cx="50" cy="50" r="40" fill="blue"/>
</svg>`

func TestDefaultLogoValidator_ValidateFile(t *testing.T) {
	maxSize := int64(5 << 20) // 5MB
	validator := NewDefaultLogoValidator(maxSize)

	tests := []struct {
		name     string
		filename string
		size     int64
		content  []byte
		wantExt  string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid PNG file",
			filename: "logo.png",
			size:     1024,
			content:  validPNGHeader,
			wantExt:  ".png",
			wantErr:  false,
		},
		{
			name:     "valid SVG file",
			filename: "logo.svg",
			size:     int64(len(validSVGContent)),
			content:  []byte(validSVGContent),
			wantExt:  ".svg",
			wantErr:  false,
		},
		{
			name:     "file too large",
			filename: "large.png",
			size:     6 << 20, // 6MB > 5MB max
			content:  validPNGHeader,
			wantErr:  true,
			errMsg:   "file too large",
		},
		{
			name:     "invalid extension - jpg",
			filename: "logo.jpg",
			size:     1024,
			content:  validPNGHeader,
			wantErr:  true,
			errMsg:   "invalid file type",
		},
		{
			name:     "invalid extension - gif",
			filename: "logo.gif",
			size:     1024,
			content:  validPNGHeader,
			wantErr:  true,
			errMsg:   "invalid file type",
		},
		{
			name:     "PNG file with wrong content type",
			filename: "logo.png",
			size:     1024,
			content:  []byte("fake content"),
			wantErr:  true,
			errMsg:   "invalid file content",
		},
		{
			name:     "case insensitive extension - PNG",
			filename: "logo.PNG",
			size:     1024,
			content:  validPNGHeader,
			wantExt:  ".png",
			wantErr:  false,
		},
		{
			name:     "case insensitive extension - SVG",
			filename: "logo.SVG",
			size:     int64(len(validSVGContent)),
			content:  []byte(validSVGContent),
			wantExt:  ".svg",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentReader := bytes.NewReader(tt.content)

			ext, err := validator.ValidateFile(tt.filename, tt.size, contentReader)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateFile() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateFile() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			if ext != tt.wantExt {
				t.Errorf("ValidateFile() ext = %v, want %v", ext, tt.wantExt)
			}
		})
	}
}

func TestDefaultLogoValidator_ValidateSVGSecurity(t *testing.T) {
	validator := NewDefaultLogoValidator(5 << 20)

	dangerousSVGs := []struct {
		name    string
		content string
		errMsg  string
	}{
		{
			name:    "SVG with script tag",
			content: `<svg><script>alert('xss')</script></svg>`,
			errMsg:  "script tags",
		},
		{
			name:    "SVG with javascript: protocol",
			content: `<svg><a href="javascript:alert('xss')">click</a></svg>`,
			errMsg:  "", // Error can be generic "script tags or event handlers are not allowed"
		},
		{
			name:    "SVG with onload handler",
			content: `<svg onload="alert('xss')"></svg>`,
			errMsg:  "event handlers",
		},
		{
			name:    "SVG with onerror handler",
			content: `<svg><img src="x" onerror="alert('xss')"></svg>`,
			errMsg:  "event handlers",
		},
		{
			name:    "SVG with script in uppercase",
			content: `<svg><SCRIPT>alert('xss')</SCRIPT></svg>`,
			errMsg:  "script tags",
		},
	}

	for _, tt := range dangerousSVGs {
		t.Run(tt.name, func(t *testing.T) {
			contentReader := bytes.NewReader([]byte(tt.content))

			_, err := validator.ValidateFile("test.svg", int64(len(tt.content)), contentReader)

			if err == nil {
				t.Errorf("ValidateFile() expected error for dangerous SVG but got nil")
				return
			}

			if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateFile() error = %v, want containing %v", err, tt.errMsg)
			}
		})
	}

	// Test valid SVG passes security check
	t.Run("valid SVG passes security check", func(t *testing.T) {
		validSVG := `<?xml version="1.0"?>
<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <circle cx="50" cy="50" r="40" fill="blue"/>
</svg>`
		contentReader := bytes.NewReader([]byte(validSVG))

		ext, err := validator.ValidateFile("test.svg", int64(len(validSVG)), contentReader)

		if err != nil {
			t.Errorf("ValidateFile() unexpected error for valid SVG: %v", err)
			return
		}

		if ext != ".svg" {
			t.Errorf("ValidateFile() ext = %v, want .svg", ext)
		}
	})
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name        string
		buffer      []byte
		wantContent string
	}{
		{
			name:        "valid PNG header",
			buffer:      validPNGHeader,
			wantContent: "image/png",
		},
		{
			name:        "SVG with XML declaration",
			buffer:      []byte(`<?xml version="1.0"?><svg>`),
			wantContent: "image/svg+xml",
		},
		{
			name:        "SVG without XML declaration",
			buffer:      []byte(`<svg xmlns="http://www.w3.org/2000/svg">`),
			wantContent: "image/svg+xml",
		},
		{
			name:        "empty buffer",
			buffer:      []byte{},
			wantContent: "",
		},
		{
			name:        "invalid content",
			buffer:      []byte("invalid content"),
			wantContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectContentType(tt.buffer)
			if result != tt.wantContent {
				t.Errorf("detectContentType() = %v, want %v", result, tt.wantContent)
			}
		})
	}
}
