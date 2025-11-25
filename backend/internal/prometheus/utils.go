package prometheus

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// validatePromQLParam validates and escapes PromQL parameters to prevent injection
func validatePromQLParam(param, paramName string) (string, error) {
	// Validate that it only contains alphanumeric characters, hyphens, and dots
	// This covers Kubernetes namespaces, pod names, deployment names, etc.
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validPattern.MatchString(param) {
		return "", fmt.Errorf("invalid %s: contains invalid characters", paramName)
	}

	// Validate max length (Kubernetes limit is usually 253)
	if len(param) > 253 {
		return "", fmt.Errorf("invalid %s: too long", paramName)
	}

	// Escape double quotes just in case (though regex above prevents them)
	escaped := strings.ReplaceAll(param, `"`, `\"`)
	return escaped, nil
}

// parseDuration parses a duration string and returns start and end times
func parseDuration(rangeParam string) (time.Time, time.Time) {
	duration := "1h"
	if rangeParam != "" {
		duration = rangeParam
	}

	endTime := time.Now()
	var startTime time.Time

	switch duration {
	case "1h":
		startTime = endTime.Add(-1 * time.Hour)
	case "6h":
		startTime = endTime.Add(-6 * time.Hour)
	case "12h":
		startTime = endTime.Add(-12 * time.Hour)
	case "1d":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "15d":
		startTime = endTime.Add(-15 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-1 * time.Hour)
	}

	return startTime, endTime
}
