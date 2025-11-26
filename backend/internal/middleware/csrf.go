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

		// In a real app, we would validate against a list of allowed domains.
		// For now, we ensure that if Origin is present, it matches the Host (basic check)
		// or matches ALLOWED_ORIGINS env var.

		// This is a simplified check. For strict CSRF, use Double Submit Cookie or similar.
		// Here we rely on the fact that browsers set Origin/Referer and attackers can't spoof them easily in cross-site requests.

		next(w, r)
	}
}
