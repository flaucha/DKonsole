package middleware

import (
	"net/http"
)

// CSRFMiddleware checks Origin/Referer for state-changing requests
func CSRFMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip for safe methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" || r.Method == "TRACE" {
			next(w, r)
			return
		}

		// Check Origin or Referer
		origin := r.Header.Get("Origin")
		referer := r.Header.Get("Referer")

		if origin == "" && referer == "" {
			http.Error(w, "Missing Origin or Referer header", http.StatusForbidden)
			return
		}

		if !IsRequestOriginOrRefererAllowed(r) {
			http.Error(w, "Origin not allowed", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}
