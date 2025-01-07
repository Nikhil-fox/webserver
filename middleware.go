package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// rateLimiter is a structure to track client request counts.
type rateLimiter struct {
	mu         sync.Mutex
	clients    map[string]*clientData
	maxReq     int
	timeWindow time.Duration
}

type clientData struct {
	requests int
	expires  time.Time
}

// newRateLimiter initializes a new rateLimiter.
func newRateLimiter(maxReq int, timeWindow time.Duration) *rateLimiter {
	return &rateLimiter{
		clients:    make(map[string]*clientData),
		maxReq:     maxReq,
		timeWindow: timeWindow,
	}
}

// allowRequest checks if a request from a client is allowed.
func (rl *rateLimiter) allowRequest(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[ip]

	// If the client is new or their limit has expired, reset their data
	if !exists || client.expires.Before(now) {
		rl.clients[ip] = &clientData{
			requests: 1,
			expires:  now.Add(rl.timeWindow),
		}
		log.Printf("[RATE LIMITER] New client or reset: IP=%s, Requests=1, Expires=%s", ip, rl.clients[ip].expires)
		return true
	}

	// If the client exceeds the max requests, log and deny access
	if client.requests >= rl.maxReq {
		log.Printf("[RATE LIMITER] Rate limit exceeded: IP=%s, Requests=%d, Expires=%s", ip, client.requests, client.expires)
		return false
	}

	// Increment the request count
	client.requests++
	log.Printf("[RATE LIMITER] Client allowed: IP=%s, Requests=%d, Expires=%s", ip, client.requests, client.expires)
	return true
}

// getClientIP retrieves the client IP address from the request.
// It first checks the X-Forwarded-For header, then falls back to RemoteAddr.
func getClientIP(r *http.Request) string {
	// Check if X-Forwarded-For header is present
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For can contain a comma-separated list of IPs; take the first one
		ip := forwarded
		if idx := len(forwarded) - len(forwarded[:1]); idx != -1 {
			ip = forwarded[:idx]
		}
		fmt.Printf("[DEBUG] X-Forwarded-For: %s, Extracted IP: %s", forwarded, ip)
		return ip
	}

	// Fallback to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	log.Printf("[DEBUG] Using RemoteAddr: %s", ip)
	return ip
}

// RateLimitingMiddleware returns a middleware that enforces rate limiting.
func RateLimitingMiddleware(rl *rateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[DEBUG] Middleware executed for path: %s", r.URL.Path)

			// Get the client IP (handle proxy scenarios)
			ip := getClientIP(r)

			// Check if the request is allowed
			if !rl.allowRequest(ip) {
				log.Printf("[RATE LIMITER] Request denied: IP=%s, Path=%s", ip, r.URL.Path)
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			log.Printf("[RATE LIMITER] Request allowed: IP=%s, Path=%s", ip, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}
