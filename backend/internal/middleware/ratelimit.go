package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

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
