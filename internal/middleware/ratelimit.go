package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Simple in-memory rate limiter per IP
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}

	// Cleanup old entries background routine
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = limiter
	}

	return limiter
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		// Simple cleanup: strictly speaking we should track last seen time,
		// but for a personal blog, clearing the map periodically is fine and safe.
		// A more robust solution would track access time.
		// For now, let's just reset the map if it gets too big to prevent memory leaks
		if len(rl.visitors) > 10000 {
			rl.visitors = make(map[string]*rate.Limiter)
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Priority: X-Real-IP > X-Forwarded-For > RemoteAddr
		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			ip = r.Header.Get("X-Forwarded-For")
			if ip != "" {
				// X-Forwarded-For can be a comma-separated list of IPs.
				// The first one is the original client IP.
				if i := strings.Index(ip, ","); i != -1 {
					ip = ip[:i]
				}
			}
		}
		if ip == "" {
			// Fallback to RemoteAddr
			ip = r.RemoteAddr
			// RemoteAddr usually contains port (e.g., "127.0.0.1:1234"), strip it for consistency
			if host, _, err := net.SplitHostPort(ip); err == nil {
				ip = host
			}
		}

		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
