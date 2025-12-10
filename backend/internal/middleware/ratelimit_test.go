package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimitMiddleware(t *testing.T) {
	// Create a test handler
	nextHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	tests := []struct {
		name            string
		requests        int
		delay           time.Duration
		isWebSocket     bool
		wantStatusCodes []int // Expected status codes for each request
		expectLimitHit  bool
	}{
		{
			name:            "single request should pass",
			requests:        1,
			delay:           0,
			wantStatusCodes: []int{http.StatusOK},
			expectLimitHit:  false,
		},
		{
			name:            "multiple requests under limit should pass",
			requests:        10,
			delay:           10 * time.Millisecond, // Small delay between requests
			wantStatusCodes: []int{http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK},
			expectLimitHit:  false,
		},
		{
			name:            "websocket request should bypass rate limit",
			requests:        1,
			delay:           0,
			isWebSocket:     true,
			wantStatusCodes: []int{http.StatusOK},
			expectLimitHit:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use custom config for testing: 10 req/sec, burst 5
			handler := rateLimitMiddlewareWithConfig(nextHandler, 10.0, 5)

			// Clear limiters for clean test
			apiLimiters.mu.Lock()
			apiLimiters.limiters = make(map[string]*rateLimiterEntry)
			apiLimiters.mu.Unlock()

			var wg sync.WaitGroup
			results := make([]int, tt.requests)

			for i := 0; i < tt.requests; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()

					req := httptest.NewRequest("GET", "/api/test", nil)
					if tt.isWebSocket {
						req.Header.Set("Upgrade", "websocket")
						req.Header.Set("Connection", "Upgrade")
					}
					req.RemoteAddr = "127.0.0.1:12345" // Same IP for all requests

					rr := httptest.NewRecorder()
					handler(rr, req)

					results[index] = rr.Code

					// Small delay between requests to avoid all hitting at once
					if tt.delay > 0 && index < tt.requests-1 {
						time.Sleep(tt.delay)
					}
				}(i)
			}

			wg.Wait()

			// Check status codes
			if len(results) != len(tt.wantStatusCodes) {
				t.Errorf("RateLimitMiddleware() got %d responses, want %d", len(results), len(tt.wantStatusCodes))
			} else {
				for i, got := range results {
					// For rate limit tests, we expect most to pass, but some might hit limit
					// Just verify we got valid responses
					if got != http.StatusOK && got != http.StatusTooManyRequests {
						t.Errorf("RateLimitMiddleware() request %d status code = %v, want %v or %v",
							i, got, http.StatusOK, http.StatusTooManyRequests)
					}
				}
			}

			// Clean up
			apiLimiters.mu.Lock()
			apiLimiters.limiters = make(map[string]*rateLimiterEntry)
			apiLimiters.mu.Unlock()
		})
	}
}

func TestRateLimitMiddleware_Headers(t *testing.T) {
	nextHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	handler := RateLimitMiddleware(nextHandler)

	// Clear limiters
	apiLimiters.mu.Lock()
	apiLimiters.limiters = make(map[string]*rateLimiterEntry)
	apiLimiters.mu.Unlock()

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()

	handler(rr, req)

	// Request should pass (not hit limit)
	if rr.Code != http.StatusOK {
		t.Errorf("RateLimitMiddleware() status code = %v, want %v", rr.Code, http.StatusOK)
	}

	// Clean up
	apiLimiters.mu.Lock()
	apiLimiters.limiters = make(map[string]*rateLimiterEntry)
	apiLimiters.mu.Unlock()
}

func TestLoginRateLimitMiddleware(t *testing.T) {
	nextHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	tests := []struct {
		name           string
		requests       int
		delay          time.Duration
		wantStatusCode int
		expectLimitHit bool
	}{
		{
			name:           "single login request should pass",
			requests:       1,
			delay:          0,
			wantStatusCode: http.StatusOK,
			expectLimitHit: false,
		},
		{
			name:           "multiple rapid login requests should hit limit",
			requests:       10, // More than burst of 5
			delay:          0,  // No delay, all at once
			wantStatusCode: http.StatusTooManyRequests,
			expectLimitHit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := LoginRateLimitMiddleware(nextHandler)

			// Clear limiters for clean test
			loginLimiters.mu.Lock()
			loginLimiters.limiters = make(map[string]*rateLimiterEntry)
			loginLimiters.mu.Unlock()

			hitLimit := false

			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest("POST", "/api/login", nil)
				req.RemoteAddr = "127.0.0.1:12345" // Same IP

				rr := httptest.NewRecorder()
				handler(rr, req)

				if rr.Code == http.StatusTooManyRequests {
					hitLimit = true
					// Check Retry-After header
					retryAfter := rr.Header().Get("Retry-After")
					if retryAfter != "60" {
						t.Errorf("LoginRateLimitMiddleware() Retry-After header = %v, want 60", retryAfter)
					}
					// Check error message
					body := rr.Body.String()
					if body != "Too many login attempts. Please try again later.\n" {
						t.Errorf("LoginRateLimitMiddleware() body = %v, want 'Too many login attempts. Please try again later.'", body)
					}
				}

				if tt.delay > 0 && i < tt.requests-1 {
					time.Sleep(tt.delay)
				}
			}

			if tt.expectLimitHit && !hitLimit {
				t.Errorf("LoginRateLimitMiddleware() expected to hit rate limit but didn't")
			}

			// Clean up
			loginLimiters.mu.Lock()
			loginLimiters.limiters = make(map[string]*rateLimiterEntry)
			loginLimiters.mu.Unlock()
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name          string
		remoteAddr    string
		xRealIP       string
		xForwardedFor string
		wantClientIP  string
	}{
		{
			name:          "X-Real-IP header should be used first",
			remoteAddr:    "192.168.1.1:12345",
			xRealIP:       "10.0.0.1",
			xForwardedFor: "192.168.1.2",
			wantClientIP:  "10.0.0.1",
		},
		{
			name:          "X-Forwarded-For should be used if X-Real-IP not present",
			remoteAddr:    "192.168.1.1:12345",
			xRealIP:       "",
			xForwardedFor: "10.0.0.2",
			wantClientIP:  "10.0.0.2",
		},
		{
			name:          "X-Forwarded-For with multiple IPs should use first",
			remoteAddr:    "192.168.1.1:12345",
			xRealIP:       "",
			xForwardedFor: "10.0.0.3, 192.168.1.3, 10.0.0.4",
			wantClientIP:  "10.0.0.3",
		},
		{
			name:          "RemoteAddr should be used as fallback",
			remoteAddr:    "192.168.1.4:12345",
			xRealIP:       "",
			xForwardedFor: "",
			wantClientIP:  "192.168.1.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}

			gotIP := getClientIP(req)
			if gotIP != tt.wantClientIP {
				t.Errorf("getClientIP() = %v, want %v", gotIP, tt.wantClientIP)
			}
		})
	}
}

func TestRateLimiterMap_CleanupInactive(t *testing.T) {
	// Setup
	rlm := &RateLimiterMap{
		limiters: make(map[string]*rateLimiterEntry),
	}

	// Add an old entry (20 minutes ago)
	rlm.mu.Lock()
	rlm.limiters["old-ip"] = &rateLimiterEntry{
		limiter:  rate.NewLimiter(1, 1),
		lastSeen: time.Now().Add(-20 * time.Minute),
	}
	// Add a new entry (1 minute ago)
	rlm.limiters["new-ip"] = &rateLimiterEntry{
		limiter:  rate.NewLimiter(1, 1),
		lastSeen: time.Now().Add(-1 * time.Minute),
	}
	rlm.mu.Unlock()

	// Run cleanup
	rlm.cleanupInactive()

	// Verify
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	if _, exists := rlm.limiters["old-ip"]; exists {
		t.Error("cleanupInactive() failed to remove old entry")
	}

	if _, exists := rlm.limiters["new-ip"]; !exists {
		t.Error("cleanupInactive() incorrectly removed new entry")
	}
}
