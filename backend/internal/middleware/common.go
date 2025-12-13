package middleware

import (
	"net/http"
	"strings"
)

// StatusRecorder wraps http.ResponseWriter to capture status code
type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *StatusRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// getClientIP extracts the real client IP from request.
// Security: By default this implementation relies on RemoteAddr to avoid IP spoofing via headers.
// If running behind a trusted proxy/Kubernetes Ingress, X-Forwarded-For should be trusted ONLY if
// the trusted proxy list is configured.
// For this remediation scope, we stick to RemoteAddr for safety unless specific trust config is added.
func getClientIP(r *http.Request) string {
	// TODO: implement Trusted Proxies configuration.
	// For now, to mitigate spoofing risks identified in analysis, we default to the direct connection IP.
	// If the user needs to trust headers, they should configure trusted CIDRs in a future update.

	// Fallback to RemoteAddr (remove port if present)
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	return ip
}

// Chain applies middlewares in order
func Chain(h http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}
