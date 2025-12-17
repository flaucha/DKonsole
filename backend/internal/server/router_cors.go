package server

import (
	"net/http"
	"os"
	"strings"

	"github.com/flaucha/DKonsole/backend/internal/middleware"
	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// enableCors applies CORS headers allowing configured origins or host match
func enableCors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

		if strings.Contains(allowedOrigins, "*") {
			utils.LogWarn("Security Warning: ALLOWED_ORIGINS contains wildcard '*'. This is insecure for authenticated sessions.", nil)
		}

		if origin == "" && r.Method != "OPTIONS" {
			next(w, r)
			return
		}

		if origin != "" && !middleware.IsRequestOriginAllowed(r) {
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
