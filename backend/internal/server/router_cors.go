package server

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// enableCors applies CORS headers allowing configured origins or host match
func enableCors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

		if origin == "" && r.Method != "OPTIONS" {
			next(w, r)
			return
		}

		allowed := false
		if origin != "" {
			// Exact matches from allowlist
			if allowedOrigins != "" {
				origins := strings.Split(allowedOrigins, ",")
				for _, o := range origins {
					o = strings.TrimSpace(o)
					if o != "" && strings.EqualFold(o, origin) {
						allowed = true
						break
					}
				}
			}

			// Always allow same-origin requests (host match)
			if !allowed {
				originURL, err := url.Parse(origin)
				if err == nil {
					host := r.Host
					if strings.Contains(host, ":") {
						host, _, _ = strings.Cut(host, ":")
					}
					originHost := originURL.Host
					if strings.Contains(originHost, ":") {
						originHost, _, _ = strings.Cut(originHost, ":")
					}
					if strings.EqualFold(strings.TrimSpace(originHost), strings.TrimSpace(host)) &&
						(originURL.Scheme == "http" || originURL.Scheme == "https") {
						allowed = true
					}
				}
			}
		}

		if !allowed && origin != "" {
			utils.LogWarn("CORS: Origin not allowed", map[string]interface{}{
				"origin": origin,
				"host":   r.Host,
				"path":   r.URL.Path,
			})
			http.Error(w, "Origin not allowed", http.StatusForbidden)
			return
		}

		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
