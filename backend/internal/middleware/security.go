package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeadersMiddleware adds security headers to all HTTP responses
// Similar to Helmet.js for Node.js, this middleware implements OWASP security best practices
func SecurityHeadersMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w, r)
		next(w, r)
	}
}

// SecurityHeadersHandler wraps an http.Handler with security headers
func SecurityHeadersHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w, r)
		next.ServeHTTP(w, r)
	})
}

// setSecurityHeaders sets all security headers on the response
func setSecurityHeaders(w http.ResponseWriter, r *http.Request) {
	// X-Content-Type-Options: Prevents MIME type sniffing
	// Prevents browsers from interpreting files as a different MIME type
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// X-Frame-Options: Prevents clickjacking attacks
	// DENY prevents the page from being displayed in a frame
	w.Header().Set("X-Frame-Options", "DENY")

	// X-XSS-Protection: Enables XSS filtering in older browsers
	// Modern browsers have built-in XSS protection, but this helps with legacy support
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	// Referrer-Policy: Controls referrer information sent with requests
	// strict-origin-when-cross-origin: Send full URL for same-origin, only origin for cross-origin HTTPS
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	// Permissions-Policy: Restricts browser features and APIs
	// Disable geolocation, microphone, and camera by default
	w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

	// Strict-Transport-Security (HSTS): Force HTTPS connections
	// Only set if the request is already over HTTPS or behind a proxy that indicates HTTPS
	if r.Header.Get("X-Forwarded-Proto") == "https" || r.TLS != nil {
		// max-age=31536000 = 1 year, includeSubDomains applies to all subdomains
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	}

	// Content-Security-Policy: Prevents XSS, data injection, and other attacks
	// This is a basic policy - can be customized per route if needed
	csp := buildCSP(r)
	if csp != "" {
		w.Header().Set("Content-Security-Policy", csp)
	}
}

// buildCSP constructs a Content Security Policy based on the request
// For API routes, we use a more restrictive policy
// For static files, we allow needed resources without enabling inline execution
func buildCSP(r *http.Request) string {
	path := r.URL.Path

	// For API routes, use strict CSP
	if strings.HasPrefix(path, "/api") {
		return "default-src 'self'; script-src 'none'; style-src 'none'; " +
			"img-src 'none'; font-src 'none'; connect-src 'self' ws: wss:; " +
			"frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
	}

	// For static files and frontend routes, allow required resources while avoiding inline execution.
	// Nonces/hashes can be added when wiring the HTML template if inline assets become necessary.
	return "default-src 'self'; script-src 'self' https://cdn.jsdelivr.net; style-src 'self' https://fonts.googleapis.com https://cdn.jsdelivr.net; img-src 'self' data: https:; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self' ws: wss: https://cdn.jsdelivr.net; worker-src 'self' blob:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
}
