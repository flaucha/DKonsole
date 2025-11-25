package logo

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// LogoValidator defines the interface for logo file validation
type LogoValidator interface {
	// ValidateFile validates a file based on its name, size, and content
	ValidateFile(filename string, size int64, content io.Reader) (string, error)
}

// DefaultLogoValidator implements LogoValidator with standard validation rules
type DefaultLogoValidator struct {
	maxSize int64
}

// NewDefaultLogoValidator creates a new DefaultLogoValidator
func NewDefaultLogoValidator(maxSize int64) *DefaultLogoValidator {
	return &DefaultLogoValidator{
		maxSize: maxSize,
	}
}

// ValidateFile validates a file based on its name, size, and content
// Returns the file extension if valid, or an error if invalid
func (v *DefaultLogoValidator) ValidateFile(filename string, size int64, content io.Reader) (string, error) {
	// Validate size
	if size > v.maxSize {
		return "", fmt.Errorf("file too large (max %d bytes)", v.maxSize)
	}

	// Validate extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".png" && ext != ".svg" {
		return "", fmt.Errorf("invalid file type. Only .png and .svg are allowed")
	}

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	n, err := content.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error reading file: %w", err)
	}
	
	// Reset reader position for later use
	if seeker, ok := content.(io.Seeker); ok {
		seeker.Seek(0, 0)
	}

	contentType := detectContentType(buffer[:n])

	// Validate content type matches extension
	if ext == ".png" {
		if contentType != "image/png" {
			return "", fmt.Errorf("invalid file content (not a PNG)")
		}
	}

	// For SVG, perform additional security checks
	if ext == ".svg" {
		if err := v.validateSVGSecurity(content); err != nil {
			return "", fmt.Errorf("invalid SVG content: %w", err)
		}
	}

	return ext, nil
}

// validateSVGSecurity performs deep inspection of SVG content to prevent XSS
func (v *DefaultLogoValidator) validateSVGSecurity(content io.Reader) error {
	// Read the whole file to check content
	fileContent, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Reset reader position
	if seeker, ok := content.(io.Seeker); ok {
		seeker.Seek(0, 0)
	}

	contentStr := strings.ToLower(string(fileContent))
	
	// Check for dangerous patterns
	dangerousPatterns := []string{
		"<script",
		"javascript:",
		"onload=",
		"onerror=",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(contentStr, pattern) {
			return fmt.Errorf("script tags or event handlers are not allowed")
		}
	}

	return nil
}

// detectContentType detects the MIME type from file content
func detectContentType(buffer []byte) string {
	if len(buffer) < 2 {
		return ""
	}

	// PNG signature: 89 50 4E 47
	if len(buffer) >= 8 &&
		buffer[0] == 0x89 && buffer[1] == 0x50 &&
		buffer[2] == 0x4E && buffer[3] == 0x47 {
		return "image/png"
	}

	// SVG might start with XML declaration or <svg
	content := string(buffer)
	if strings.HasPrefix(strings.TrimSpace(content), "<?xml") ||
		strings.HasPrefix(strings.TrimSpace(content), "<svg") {
		return "image/svg+xml"
	}

	return ""
}





