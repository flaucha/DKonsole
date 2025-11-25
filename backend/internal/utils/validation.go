package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// DNS-1123 subdomain regex: RFC 1123 compliant Kubernetes names
	// Must start and end with alphanumeric, can contain '-' and '.'
	dns1123SubdomainRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

	// DNS-1123 label regex: For individual labels (no dots)
	// Must start and end with alphanumeric, can contain '-'
	dns1123LabelRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

	// Path traversal patterns to detect (note: forward slash / is allowed as path separator)
	// We check for these patterns explicitly in ValidatePath
)

// ValidateNamespace validates a Kubernetes namespace name
// Namespaces must follow DNS-1123 label format (no dots allowed)
func ValidateNamespace(namespace string) error {
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	// Check length (max 63 characters for labels)
	if len(namespace) > 63 {
		return fmt.Errorf("namespace too long (max 63 characters): %s", namespace)
	}

	// Validate format: DNS-1123 label (no dots in namespace names)
	if !dns1123LabelRegex.MatchString(namespace) {
		return fmt.Errorf("invalid namespace format: must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character: %s", namespace)
	}

	return nil
}

// ValidateResourceName validates a Kubernetes resource name
// Resource names can be DNS-1123 subdomain (allows dots) or label (no dots)
// This function allows both formats
func ValidateResourceName(name string) error {
	if name == "" {
		return fmt.Errorf("resource name is required")
	}

	// Check length (max 253 characters for subdomains)
	if len(name) > 253 {
		return fmt.Errorf("resource name too long (max 253 characters): %s", name)
	}

	// Validate format: DNS-1123 subdomain (allows dots for subresources)
	if !dns1123SubdomainRegex.MatchString(name) {
		return fmt.Errorf(
			"invalid resource name format: must consist of lower case alphanumeric characters, '-' or '.', "+
				"and must start and end with an alphanumeric character: %s", name)
	}

	return nil
}

// ValidatePath validates a file path to prevent path traversal attacks
// Returns error if path contains dangerous patterns
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}

	// Normalize path (decode URL encoding if present)
	normalized := strings.ToLower(path)

	// Check for protocol schemes first (http://, file://, etc.)
	if strings.Contains(normalized, "://") {
		return fmt.Errorf("invalid path: protocol schemes are not allowed: %s", path)
	}

	// Check for absolute paths
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		return fmt.Errorf("invalid path: absolute paths are not allowed: %s", path)
	}

	// Check for path traversal patterns (but allow forward slash as path separator)
	// Check for .. (parent directory)
	if strings.Contains(normalized, "..") || strings.Contains(normalized, "%2e%2e") {
		return fmt.Errorf("invalid path: contains dangerous pattern '..': %s", path)
	}

	// Check for backslash (Windows path separator, but also used in attacks)
	if strings.Contains(normalized, "\\") || strings.Contains(normalized, "%5c") {
		return fmt.Errorf("invalid path: backslashes are not allowed: %s", path)
	}

	return nil
}

// Note: ValidateK8sName is defined in utils.go for backward compatibility
// This file provides stricter validation functions: ValidateNamespace, ValidateResourceName, etc.

// ValidatePodName validates a pod name specifically
// Pod names follow DNS-1123 label format (no dots)
func ValidatePodName(podName string) error {
	if podName == "" {
		return fmt.Errorf("pod name is required")
	}

	// Check length
	if len(podName) > 63 {
		return fmt.Errorf("pod name too long (max 63 characters): %s", podName)
	}

	// Validate format: DNS-1123 label (no dots)
	if !dns1123LabelRegex.MatchString(podName) {
		return fmt.Errorf("invalid pod name format: must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character: %s", podName)
	}

	return nil
}

// ValidateContainerName validates a container name
// Container names follow DNS-1123 label format (no dots)
func ValidateContainerName(containerName string) error {
	if containerName == "" {
		return fmt.Errorf("container name is required")
	}

	// Check length
	if len(containerName) > 63 {
		return fmt.Errorf("container name too long (max 63 characters): %s", containerName)
	}

	// Validate format: DNS-1123 label (no dots)
	if !dns1123LabelRegex.MatchString(containerName) {
		return fmt.Errorf(
			"invalid container name format: must consist of lower case alphanumeric characters or '-', "+
				"and must start and end with an alphanumeric character: %s", containerName)
	}

	return nil
}
