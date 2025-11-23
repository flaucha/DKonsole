package main

import (
	"log"
	"net/http"
	"sync"
	"time"
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

// AuditMiddleware logs request details
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
		if claims, ok := r.Context().Value("user").(*Claims); ok {
			user = claims.Username
		}

		// Get status code
		status := http.StatusOK
		if recorder != nil {
			status = recorder.Status
		} else if isWebSocket {
			// For WebSocket, we can't easily get the status, assume success if we got here
			status = http.StatusSwitchingProtocols
		}

		// Log format: [AUDIT] | Status | Duration | User | Method | Path
		log.Printf("[AUDIT] | %d | %v | %s | %s %s",
			status,
			duration,
			user,
			r.Method,
			r.URL.Path,
		)
	}
}

// Simple in-memory rate limiter
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
}

type visitor struct {
	lastSeen time.Time
	count    int
}

var limiter = &RateLimiter{
	visitors: make(map[string]*visitor),
}

// RateLimitMiddleware limits requests per IP
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for WebSocket upgrade requests or handle them carefully
		if r.Header.Get("Upgrade") == "websocket" {
			next(w, r)
			return
		}

		ip := r.RemoteAddr
		// In a real app behind proxy, use X-Forwarded-For if trusted

		limiter.mu.Lock()
		v, exists := limiter.visitors[ip]
		if !exists {
			limiter.visitors[ip] = &visitor{lastSeen: time.Now(), count: 1}
			limiter.mu.Unlock()
			next(w, r)
			return
		}

		if time.Since(v.lastSeen) > time.Minute {
			v.count = 1
			v.lastSeen = time.Now()
		} else {
			v.count++
		}

		if v.count > 300 { // 300 requests per minute limit
			limiter.mu.Unlock()
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		limiter.mu.Unlock()

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
