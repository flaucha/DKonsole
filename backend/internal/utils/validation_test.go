package utils

import (
	"strings"
	"testing"
)

func TestValidateNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid namespace",
			namespace: "default",
			wantErr:   false,
		},
		{
			name:      "valid namespace with hyphen",
			namespace: "my-namespace",
			wantErr:   false,
		},
		{
			name:      "empty namespace",
			namespace: "",
			wantErr:   true,
			errMsg:    "namespace is required",
		},
		{
			name:      "namespace with uppercase",
			namespace: "Default",
			wantErr:   true,
			errMsg:    "invalid namespace format",
		},
		{
			name:      "namespace with dots",
			namespace: "my.namespace",
			wantErr:   true,
			errMsg:    "invalid namespace format",
		},
		{
			name:      "namespace starting with hyphen",
			namespace: "-namespace",
			wantErr:   true,
			errMsg:    "invalid namespace format",
		},
		{
			name:      "namespace too long",
			namespace: string(make([]byte, 64)), // 64 characters
			wantErr:   true,
			errMsg:    "namespace too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNamespace(tt.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
					t.Errorf("ValidateNamespace() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateResourceName(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid resource name",
			resource: "my-resource",
			wantErr:  false,
		},
		{
			name:     "valid resource name with dots",
			resource: "my.resource.name",
			wantErr:  false,
		},
		{
			name:     "empty resource name",
			resource: "",
			wantErr:  true,
			errMsg:   "resource name is required",
		},
		{
			name:     "resource name with uppercase",
			resource: "MyResource",
			wantErr:  true,
			errMsg:   "invalid resource name format",
		},
		{
			name:     "resource name starting with hyphen",
			resource: "-resource",
			wantErr:  true,
			errMsg:   "invalid resource name format",
		},
		{
			name:     "resource name too long",
			resource: string(make([]byte, 254)), // 254 characters
			wantErr:  true,
			errMsg:   "resource name too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceName(tt.resource)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateResourceName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateResourceName() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid relative path",
			path:    "config/app.yaml",
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errMsg:  "path is required",
		},
		{
			name:    "path with ..",
			path:    "../../etc/passwd",
			wantErr: true,
			errMsg:  "contains dangerous pattern",
		},
		{
			name:    "path with /",
			path:    "/etc/passwd",
			wantErr: true,
			errMsg:  "absolute paths are not allowed",
		},
		{
			name:    "path with \\",
			path:    "..\\windows\\system32",
			wantErr: true,
			errMsg:  "contains dangerous pattern",
		},
		{
			name:    "path with URL encoded ..",
			path:    "%2e%2e/etc/passwd",
			wantErr: true,
			errMsg:  "contains dangerous pattern",
		},
		{
			name:    "path with protocol",
			path:    "http://example.com/file",
			wantErr: true,
			errMsg:  "protocol schemes are not allowed",
		},
		{
			name:    "path with file protocol",
			path:    "file:///etc/passwd",
			wantErr: true,
			errMsg:  "protocol schemes are not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePath() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidatePodName(t *testing.T) {
	tests := []struct {
		name    string
		podName string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid pod name",
			podName: "my-pod",
			wantErr: false,
		},
		{
			name:    "empty pod name",
			podName: "",
			wantErr: true,
			errMsg:  "pod name is required",
		},
		{
			name:    "pod name with dots",
			podName: "my.pod",
			wantErr: true,
			errMsg:  "invalid pod name format",
		},
		{
			name:    "pod name with uppercase",
			podName: "MyPod",
			wantErr: true,
			errMsg:  "invalid pod name format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePodName(tt.podName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePodName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePodName() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateContainerName(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "valid container name",
			containerName: "my-container",
			wantErr:       false,
		},
		{
			name:          "empty container name",
			containerName: "",
			wantErr:       true,
			errMsg:        "container name is required",
		},
		{
			name:          "container name with dots",
			containerName: "my.container",
			wantErr:       true,
			errMsg:        "invalid container name format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContainerName(tt.containerName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContainerName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateContainerName() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}
