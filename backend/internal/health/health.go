package health

import (
	"net/http"
)

// HealthHandler is an unauthenticated liveness endpoint
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
