package middleware

import (
	"net/http"
	"net/url"
	"os"
	"strings"
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

		if !validateOrigin(origin, referer, r.Host) {
			http.Error(w, "Origin not allowed", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// validateOrigin checks Origin/Referer against allowed list or the request host
func validateOrigin(origin, referer, requestHost string) bool {
	allowedList := parseAllowedOrigins()
	requestHost = normalizeHost(requestHost)

	// Prefer Origin header when present
	if origin != "" {
		return isAllowed(origin, allowedList, requestHost)
	}

	// Fallback to Referer host when Origin is missing
	if referer != "" {
		refURL, err := url.Parse(referer)
		if err != nil {
			return false
		}
		refOrigin := refURL.Scheme + "://" + refURL.Host
		return isAllowed(refOrigin, allowedList, requestHost)
	}

	return false
}

func parseAllowedOrigins() []string {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var allowed []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			allowed = append(allowed, trimmed)
		}
	}
	return allowed
}

func isAllowed(origin string, allowedList []string, requestHost string) bool {
	originURL, err := url.Parse(origin)
	if err != nil || originURL.Host == "" {
		return false
	}
	originHost := normalizeHost(originURL.Host)

	// If allowlist is set, match against it (full origin or host match)
	if len(allowedList) > 0 {
		for _, allowed := range allowedList {
			if strings.EqualFold(allowed, origin) {
				return true
			}
			if normalizeHost(allowed) == originHost {
				return true
			}
		}
		return false
	}

	// Without allowlist, require same host
	return originHost == requestHost
}

func normalizeHost(hostport string) string {
	host := strings.TrimSpace(hostport)
	if strings.Contains(host, ":") {
		host, _, _ = strings.Cut(host, ":")
	}
	return strings.ToLower(host)
}
