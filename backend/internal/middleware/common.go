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

// getClientIP extracts the real client IP from request, handling proxies
func getClientIP(r *http.Request) string {
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

// Chain applies middlewares in order
func Chain(h http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}
