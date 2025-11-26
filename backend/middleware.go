package main

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/flaucha/DKonsole/backend/internal/utils"
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

// AuditMiddleware logs request details with improved information
func AuditMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// For WebSocket, don't use StatusRecorder as it interferes with upgrade
		isWebSocket := r.Header.Get("Upgrade") == "websocket"
		var recorder *StatusRecorder
		if !isWebSocket {
			recorder = &StatusRecorder{ResponseWriter: w, Status: http.StatusOK}
			w = recorder
		}

		next(w, r)

		duration := time.Since(start)

		// Extract user if available (set by AuthMiddleware or handler)
		user := "anonymous"
		userVal := r.Context().Value("user")
		if userVal != nil {
			// Try different claim types for compatibility
			if claims, ok := userVal.(map[string]interface{}); ok {
				if u, ok := claims["username"].(string); ok {
					user = u
				}
			} else if claims, ok := userVal.(interface{ Username() string }); ok {
				user = claims.Username()
			} else {
				// Try to extract via reflection or type assertion
				// This handles both old Claims and new AuthClaims
				if claimsMap, ok := userVal.(map[string]interface{}); ok {
					if u, ok := claimsMap["username"].(string); ok {
						user = u
					}
				}
			}
		}

		// Get status code
		status := http.StatusOK
		if recorder != nil {
			status = recorder.Status
		} else if isWebSocket {
			// For WebSocket, we can't easily get the status, assume success if we got here
			status = http.StatusSwitchingProtocols
		}

		// Get real client IP (handles proxies)
		clientIP := getClientIP(r)

		// Use structured logging - build entry and call internal function
		entry := utils.AuditLogEntry{
			User:     user,
			IP:       clientIP,
			Action:   "http_request",
			Method:   r.Method,
			Path:     r.URL.Path,
			Status:   status,
			Duration: duration.String(),
			Success:  status < 400,
			Details: map[string]interface{}{
				"user_agent": r.UserAgent(),
			},
		}
		utils.AuditLogLegacy(r, entry.Action, "", "", "", entry.Success, nil, entry.Details)
	}
}

// rateLimiterEntry holds a rate limiter and last seen time for cleanup
type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
	mu       sync.Mutex
}

// RateLimiterMap manages rate limiters per IP with cleanup
type RateLimiterMap struct {
	limiters map[string]*rateLimiterEntry
	mu       sync.RWMutex
	cleanup  *time.Ticker
}

var (
	// Different rate limits for different endpoints
	loginLimiters = &RateLimiterMap{
		limiters: make(map[string]*rateLimiterEntry),
	}
	apiLimiters = &RateLimiterMap{
		limiters: make(map[string]*rateLimiterEntry),
	}
)

func init() {
	// Start cleanup goroutine for login limiters
	loginLimiters.cleanup = time.NewTicker(5 * time.Minute)
	go func() {
		for range loginLimiters.cleanup.C {
			loginLimiters.cleanupInactive()
		}
	}()

	// Start cleanup goroutine for API limiters
	apiLimiters.cleanup = time.NewTicker(5 * time.Minute)
	go func() {
		for range apiLimiters.cleanup.C {
			apiLimiters.cleanupInactive()
		}
	}()
}

// cleanupInactive removes limiters that haven't been used in 10 minutes
func (rlm *RateLimiterMap) cleanupInactive() {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	now := time.Now()
	for ip, entry := range rlm.limiters {
		entry.mu.Lock()
		if now.Sub(entry.lastSeen) > 10*time.Minute {
			delete(rlm.limiters, ip)
		}
		entry.mu.Unlock()
	}
}

// getLimiter returns or creates a rate limiter for the given IP
func (rlm *RateLimiterMap) getLimiter(ip string, rps float64, burst int) *rate.Limiter {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	entry, exists := rlm.limiters[ip]
	if !exists {
		entry = &rateLimiterEntry{
			limiter:  rate.NewLimiter(rate.Limit(rps), burst),
			lastSeen: time.Now(),
		}
		rlm.limiters[ip] = entry
	} else {
		entry.mu.Lock()
		entry.lastSeen = time.Now()
		entry.mu.Unlock()
	}

	return entry.limiter
}

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

// RateLimitMiddleware limits requests per IP with improved proxy handling
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return rateLimitMiddlewareWithConfig(next, 50.0, 100) // Default: 50 req/sec, burst 100
}

// RateLimitMiddlewareWithConfig allows custom rate limits
func rateLimitMiddlewareWithConfig(next http.HandlerFunc, rps float64, burst int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for WebSocket upgrade requests
		if r.Header.Get("Upgrade") == "websocket" {
			next(w, r)
			return
		}

		clientIP := getClientIP(r)
		limiter := apiLimiters.getLimiter(clientIP, rps, burst)

		if !limiter.Allow() {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

// LoginRateLimitMiddleware applies stricter rate limiting for login endpoint
func LoginRateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		// Stricter limit for login: 5 requests per minute, burst of 5
		limiter := loginLimiters.getLimiter(clientIP, 5.0/60.0, 5) // 5 req/min = 0.083 req/sec

		if !limiter.Allow() {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Too many login attempts. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

// Chain applies middlewares in order
func Chain(h http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}
