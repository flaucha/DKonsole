package middleware

import (
	"net/http"
	"time"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// AuditMiddleware logs request details with improved information
func AuditMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// For WebSocket, don't use StatusRecorder as it interferes with upgrade
		isWebSocket := r.Header.Get("Upgrade") == "websocket"
		var recorder *StatusRecorder
		if !isWebSocket {
			recorder = &StatusRecorder{ResponseWriter: w, Status: http.StatusOK}
			w = recorder
		}

		next(w, r)

		duration := time.Since(start)

		// Extract user if available (set by AuthMiddleware or handler)
		user := "anonymous"
		// Use the same context key as auth package to avoid import cycle
		// This matches auth.userContextKey = "user"
		userVal := r.Context().Value("user")
		if userVal != nil {
			// Try different claim types for compatibility
			if claims, ok := userVal.(map[string]interface{}); ok {
				if u, ok := claims["username"].(string); ok {
					user = u
				}
			} else if claims, ok := userVal.(interface{ Username() string }); ok {
				user = claims.Username()
			} else {
				// Try to extract via reflection or type assertion
				// This handles both old Claims and new AuthClaims
				if claimsMap, ok := userVal.(map[string]interface{}); ok {
					if u, ok := claimsMap["username"].(string); ok {
						user = u
					}
				}
			}
		}

		// Get status code
		status := http.StatusOK
		if recorder != nil {
			status = recorder.Status
		} else if isWebSocket {
			// For WebSocket, we can't easily get the status, assume success if we got here
			status = http.StatusSwitchingProtocols
		}

		// Get real client IP (handles proxies)
		clientIP := getClientIP(r)

		// Use structured logging - build entry and log directly using LogAuditEntry
		// This preserves all fields including Method, Path, Status, and Duration
		entry := utils.AuditLogEntry{
			User:     user,
			IP:       clientIP,
			Action:   "http_request",
			Method:   r.Method,
			Path:     r.URL.Path,
			Status:   status,
			Duration: duration.String(),
			Success:  status < 400,
			Details: map[string]interface{}{
				"user_agent": r.UserAgent(),
			},
		}
		utils.LogAuditEntry(entry)
	}
}
