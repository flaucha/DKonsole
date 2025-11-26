package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// HandleError logs detailed error internally and returns sanitized message to user
func HandleError(w http.ResponseWriter, err error, userMessage string, statusCode int) {
	// Log the full error internally with context using structured logging
	LogError(err, userMessage, map[string]interface{}{
		"status_code": statusCode,
	})

	// Send generic message to user (don't expose internal details)
	http.Error(w, userMessage, statusCode)
}

// CreateTimeoutContext creates a context with a timeout for Kubernetes operations
// Default timeout is 30 seconds, but can be customized via environment variable
func CreateTimeoutContext() (context.Context, context.CancelFunc) {
	timeout := 30 * time.Second
	if timeoutStr := os.Getenv("K8S_OPERATION_TIMEOUT"); timeoutStr != "" {
		if parsed, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = parsed
		}
	}
	return context.WithTimeout(context.Background(), timeout)
}

// CreateRequestContext creates a context from HTTP request with timeout
// This ensures that if the user cancels the HTTP request, Kubernetes API calls are also canceled
func CreateRequestContext(r *http.Request) (context.Context, context.CancelFunc) {
	timeout := 30 * time.Second
	if timeoutStr := os.Getenv("K8S_OPERATION_TIMEOUT"); timeoutStr != "" {
		if parsed, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = parsed
		}
	}
	// Use request context as parent so cancellation propagates
	return context.WithTimeout(r.Context(), timeout)
}

// IsSystemNamespace checks if a namespace is a system namespace
func IsSystemNamespace(ns string) bool {
	systemNamespaces := map[string]bool{
		"kube-system":     true,
		"kube-public":     true,
		"kube-node-lease": true,
	}
	return systemNamespaces[ns]
}

// ValidateK8sName validates a Kubernetes resource name according to RFC 1123
func ValidateK8sName(name, paramName string) error {
	if name == "" {
		return fmt.Errorf("%s is required", paramName)
	}

	// Validate length
	if len(name) > 253 {
		return fmt.Errorf("invalid %s: too long (max 253 characters)", paramName)
	}

	// Validate according to RFC 1123 (Kubernetes names) using regex
	// Lowercase alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character.
	var dns1123SubdomainRegexp = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	if !dns1123SubdomainRegexp.MatchString(name) {
		return fmt.Errorf("invalid %s: must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character", paramName)
	}

	return nil
}

// CheckQuotaLimits validates CPU and memory requests/limits against quota
func CheckQuotaLimits(quota corev1.ResourceQuota, requests, limits map[string]interface{}) error {
	// Get quota limits
	hard := quota.Spec.Hard
	_ = quota.Status.Used // Available for future use in full validation

	// Check CPU
	if cpuReq, ok := requests["cpu"].(string); ok && hard != nil {
		if hardCPU, exists := hard[corev1.ResourceRequestsCPU]; exists {
			// Parse and compare (simplified - would need proper quantity parsing in production)
			// For now, we'll let Kubernetes handle the validation and just log
			LogDebug("Checking CPU request against quota limit", map[string]interface{}{
				"cpu_request": cpuReq,
				"quota_limit": hardCPU.String(),
			})
		}
	}

	// Check Memory
	if memReq, ok := requests["memory"].(string); ok && hard != nil {
		if hardMem, exists := hard[corev1.ResourceRequestsMemory]; exists {
			LogDebug("Checking memory request against quota limit", map[string]interface{}{
				"memory_request": memReq,
				"quota_limit":    hardMem.String(),
			})
		}
	}

	// Note: Full validation would require parsing Kubernetes resource quantities
	// For now, we rely on Kubernetes API server to reject if quota is exceeded
	// This function serves as a pre-check and audit point
	return nil
}

// CheckStorageQuota validates storage requests against quota
func CheckStorageQuota(quota corev1.ResourceQuota, storage string) error {
	hard := quota.Spec.Hard
	if hard == nil {
		return nil
	}

	if hardStorage, exists := hard[corev1.ResourceRequestsStorage]; exists {
		// Parse and compare storage (simplified)
		LogDebug("Checking storage request against quota limit", map[string]interface{}{
			"storage_request": storage,
			"quota_limit":     hardStorage.String(),
		})
		// Full validation would require parsing Kubernetes resource quantities
	}

	return nil
}

// GetClientIP extracts the real client IP from request, handling proxies
func GetClientIP(r *http.Request) string {
	// Try X-Real-IP first (set by nginx, traefik, etc.)
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	// Try X-Forwarded-For (may contain multiple IPs, take the first)
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	// Fallback to RemoteAddr (remove port if present)
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	return ip
}

// AuditLogLegacy logs detailed audit information for critical actions
// This is a convenience function that builds an AuditLogEntry from parameters
//
// Deprecated: This function does not preserve Method, Path, Status, and Duration fields.
// Use LogAuditEntry with a complete AuditLogEntry structure instead for full audit information.
// This function is kept for backward compatibility with existing code.
func AuditLogLegacy(r *http.Request, action, resourceKind, resourceName, namespace string, success bool, err error, details map[string]interface{}) {
	user := "anonymous"
	// Try to get user from context - this will need to be adapted based on how Claims is structured
	// For now, we'll use a simple approach
	if userVal := r.Context().Value("user"); userVal != nil {
		// This will need to be updated when we move auth to its own package
		if claims, ok := userVal.(interface{ Username() string }); ok {
			user = claims.Username()
		} else if claimsMap, ok := userVal.(map[string]interface{}); ok {
			if u, ok := claimsMap["username"].(string); ok {
				user = u
			}
		}
	}

	// Get real client IP (handles proxies)
	clientIP := GetClientIP(r)

	resource := resourceKind
	if resourceName != "" {
		resource = fmt.Sprintf("%s/%s", resourceKind, resourceName)
	}

	entry := AuditLogEntry{
		User:      user,
		IP:        clientIP,
		Action:    action,
		Resource:  resource,
		Namespace: namespace,
		Success:   success,
		Details:   details,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// Call the structured auditLogInternal function from logger.go
	auditLogInternal(entry)
}

// AuditLog is a convenience wrapper that maintains backward compatibility
// It calls AuditLogLegacy internally
func AuditLog(r *http.Request, action, resourceKind, resourceName, namespace string, success bool, err error, details map[string]interface{}) {
	AuditLogLegacy(r, action, resourceKind, resourceName, namespace, success, err, details)
}

// JSONResponse writes a JSON response with the given status code and data
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		LogError(err, "Error writing JSON response", map[string]interface{}{
			"status_code": statusCode,
		})
		// Fallback to a plain text error if JSON encoding fails
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ErrorResponse writes a consistent JSON error response
func ErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	JSONResponse(w, statusCode, map[string]string{"error": message})
}

// PodParams holds common parameters for pod-related requests
type PodParams struct {
	Namespace string
	PodName   string
	Container string
	Cluster   string
}

// ParsePodParams extracts and validates pod-related parameters from an HTTP request
// Uses strict validation to prevent path traversal and injection attacks
func ParsePodParams(r *http.Request) (*PodParams, error) {
	ns := r.URL.Query().Get("namespace")
	podName := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")
	cluster := r.URL.Query().Get("cluster")

	if ns == "" || podName == "" {
		return nil, fmt.Errorf("missing namespace or pod parameter")
	}

	// Use strict validation functions
	if err := ValidateNamespace(ns); err != nil {
		return nil, err
	}
	if err := ValidatePodName(podName); err != nil {
		return nil, err
	}
	if container != "" {
		if err := ValidateContainerName(container); err != nil {
			return nil, err
		}
	}

	return &PodParams{
		Namespace: ns,
		PodName:   podName,
		Container: container,
		Cluster:   cluster,
	}, nil
}

// SuccessResponse writes a standardized JSON success response
func SuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	response := map[string]interface{}{
		"status":  "success",
		"message": message,
	}
	if data != nil {
		response["data"] = data
	}
	JSONResponse(w, http.StatusOK, response)
}
