package auth

import (
	"context"
	"net/http"

	"github.com/flaucha/DKonsole/backend/internal/utils"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const userContextKey contextKey = "user"

// UserContextKey returns the context key for user information
// This is exported so other packages can access user information from context
func UserContextKey() contextKey {
	return userContextKey
}

// AuthMiddleware is an HTTP middleware that protects routes by requiring valid JWT authentication.
// It extracts and validates the JWT token from the request, and if valid, adds the user claims
// to the request context for use by subsequent handlers.
//
// OPTIONS requests are allowed through for CORS preflight.
// Unauthenticated requests receive a 401 Unauthorized response.
func (s *Service) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow OPTIONS for CORS
		if r.Method == "OPTIONS" {
			next(w, r)
			return
		}

		s.mu.RLock()
		setupMode := s.setupMode
		jwtService := s.jwtService
		s.mu.RUnlock()

		// If in setup mode, block authenticated routes
		if setupMode {
			utils.ErrorResponse(w, http.StatusPreconditionFailed, "Setup required. Please complete the initial setup first.")
			return
		}

		// Use JWTService to authenticate request
		if jwtService == nil {
			utils.ErrorResponse(w, http.StatusInternalServerError, "JWT service not initialized")
			return
		}

		claims, err := jwtService.AuthenticateRequest(r)
		if err != nil {
			utils.ErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		// Log claims for debugging - REMOVED for security (P1)
		// utils.LogInfo("AuthMiddleware: extracted claims from JWT", map[string]interface{}{...})

		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next(w, r.WithContext(ctx))
	}
}
