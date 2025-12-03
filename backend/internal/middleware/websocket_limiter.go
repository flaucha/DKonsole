package middleware

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// WebSocketConnectionLimiter limits the number of concurrent WebSocket connections per IP
type WebSocketConnectionLimiter struct {
	connections map[string]int
	maxPerIP    int
	mu          sync.RWMutex
	cleanup     *time.Ticker
}

var (
	// Global WebSocket connection limiter
	wsLimiter = &WebSocketConnectionLimiter{
		connections: make(map[string]int),
		maxPerIP:    getMaxWebSocketConnections(),
	}
)

func init() {
	// Start cleanup goroutine to reset connection counts periodically
	wsLimiter.cleanup = time.NewTicker(1 * time.Minute)
	go func() {
		for range wsLimiter.cleanup.C {
			wsLimiter.cleanupConnections()
		}
	}()
}

// getMaxWebSocketConnections returns the maximum WebSocket connections per IP from environment
// Default: 20 connections per IP (configurable via MAX_WS_CONNECTIONS)
func getMaxWebSocketConnections() int {
	maxStr := os.Getenv("MAX_WS_CONNECTIONS")
	if maxStr != "" {
		if val, err := strconv.Atoi(maxStr); err == nil && val > 0 {
			return val
		}
	}
	return 20
}

// cleanupConnections resets connection counts (called periodically)
func (l *WebSocketConnectionLimiter) cleanupConnections() {
	l.mu.Lock()
	defer l.mu.Unlock()
	// Reset all counts - connections will be tracked again as they're established
	// This prevents stale entries from accumulating
	for ip := range l.connections {
		if l.connections[ip] <= 0 {
			delete(l.connections, ip)
		}
	}
}

// incrementConnection increments the connection count for an IP
func (l *WebSocketConnectionLimiter) incrementConnection(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	count := l.connections[ip]
	if count >= l.maxPerIP {
		return false // Limit exceeded
	}

	l.connections[ip] = count + 1
	return true
}

// decrementConnection decrements the connection count for an IP
func (l *WebSocketConnectionLimiter) decrementConnection(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	count := l.connections[ip]
	if count > 0 {
		l.connections[ip] = count - 1
		if l.connections[ip] == 0 {
			delete(l.connections, ip)
		}
	}
}

// WebSocketLimitMiddleware limits WebSocket connections per IP
// This should be applied before the WebSocket upgrade
func WebSocketLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only apply to WebSocket upgrade requests
		if r.Header.Get("Upgrade") != "websocket" {
			next(w, r)
			return
		}

		clientIP := getClientIP(r)

		// Check if connection limit is exceeded
		if !wsLimiter.incrementConnection(clientIP) {
			http.Error(w, "WebSocket connection limit exceeded. Maximum 5 concurrent connections per IP.", http.StatusTooManyRequests)
			return
		}

		// Decrement when connection closes (best effort - actual cleanup happens in cleanup goroutine)
		// Note: In a real implementation, you'd want to track active connections more precisely
		// This is a simplified version that uses periodic cleanup
		defer wsLimiter.decrementConnection(clientIP)

		next(w, r)
	}
}
