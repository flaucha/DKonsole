package prometheus

import (
	"strings"
	"testing"
	"time"
)

func TestValidatePromQLParam(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		paramName string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid namespace",
			param:     "default",
			paramName: "namespace",
			wantErr:   false,
		},
		{
			name:      "valid deployment name with hyphens",
			param:     "my-app-deployment",
			paramName: "deployment",
			wantErr:   false,
		},
		{
			name:      "valid pod name with dots",
			param:     "pod.name.example",
			paramName: "pod",
			wantErr:   false,
		},
		{
			name:      "invalid characters - spaces",
			param:     "my deployment",
			paramName: "deployment",
			wantErr:   true,
			errMsg:    "invalid characters",
		},
		{
			name:      "invalid characters - special chars",
			param:     "deployment@123",
			paramName: "deployment",
			wantErr:   true,
			errMsg:    "invalid characters",
		},
		{
			name:      "empty param",
			param:     "",
			paramName: "namespace",
			wantErr:   true, // Empty string doesn't match regex
			errMsg:    "invalid characters",
		},
		{
			name:      "too long param",
			param:     strings.Repeat("a", 254), // 254 characters
			paramName: "namespace",
			wantErr:   true,
			errMsg:    "too long",
		},
		{
			name:      "param at max length (253)",
			param:     strings.Repeat("a", 253), // 253 characters
			paramName: "namespace",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validatePromQLParam(tt.param, tt.paramName)

			if (err != nil) != tt.wantErr {
				t.Errorf("validatePromQLParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePromQLParam() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("validatePromQLParam() error = %v, want containing %v", err, tt.errMsg)
				}
				return
			}

			// Verify result is escaped (if it contained quotes, they should be escaped)
			if tt.param != "" {
				// Result should be the same or escaped version
				if result == "" && tt.param != "" {
					t.Errorf("validatePromQLParam() result is empty for valid param")
				}
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name       string
		rangeParam string
		expectDiff time.Duration // Expected difference between end and start
	}{
		{
			name:       "1h range",
			rangeParam: "1h",
			expectDiff: 1 * time.Hour,
		},
		{
			name:       "6h range",
			rangeParam: "6h",
			expectDiff: 6 * time.Hour,
		},
		{
			name:       "12h range",
			rangeParam: "12h",
			expectDiff: 12 * time.Hour,
		},
		{
			name:       "1d range",
			rangeParam: "1d",
			expectDiff: 24 * time.Hour,
		},
		{
			name:       "7d range",
			rangeParam: "7d",
			expectDiff: 7 * 24 * time.Hour,
		},
		{
			name:       "15d range",
			rangeParam: "15d",
			expectDiff: 15 * 24 * time.Hour,
		},
		{
			name:       "empty range defaults to 1h",
			rangeParam: "",
			expectDiff: 1 * time.Hour,
		},
		{
			name:       "invalid range defaults to 1h",
			rangeParam: "invalid",
			expectDiff: 1 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime, endTime := parseDuration(tt.rangeParam)

			// Verify endTime is approximately now (within 1 second tolerance)
			diff := time.Since(endTime)
			if diff < 0 {
				diff = -diff
			}
			if diff > 1*time.Second {
				t.Errorf("parseDuration() endTime is too far from now: %v", diff)
			}

			// Verify the duration between start and end is approximately correct
			actualDiff := endTime.Sub(startTime)
			diffTolerance := 1 * time.Second
			if actualDiff < tt.expectDiff-diffTolerance || actualDiff > tt.expectDiff+diffTolerance {
				t.Errorf("parseDuration() duration = %v, want approximately %v", actualDiff, tt.expectDiff)
			}

			// Verify startTime is before endTime
			if !startTime.Before(endTime) {
				t.Errorf("parseDuration() startTime should be before endTime")
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
