package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a token bucket rate limiter per IP
type RateLimiter struct {
	requests map[string]*clientLimit
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

type clientLimit struct {
	count     int
	resetTime time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientLimit),
		rate:     requestsPerMinute,
		window:   time.Minute,
	}
	// Cleanup old entries periodically
	go rl.cleanup()
	return rl
}

// Limit returns a gin middleware that rate limits requests
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now()
		limit, exists := rl.requests[clientIP]

		if !exists || now.After(limit.resetTime) {
			// First request or window expired
			rl.requests[clientIP] = &clientLimit{
				count:     1,
				resetTime: now.Add(rl.window),
			}
			c.Next()
			return
		}

		if limit.count >= rl.rate {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": limit.resetTime.Format(time.RFC3339),
			})
			return
		}

		limit.count++
		c.Next()
	}
}

// cleanup removes expired entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, limit := range rl.requests {
			if now.After(limit.resetTime) {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// StrictRateLimit is a stricter rate limiter for sensitive endpoints
func StrictRateLimit() gin.HandlerFunc {
	limiter := NewRateLimiter(10) // 10 requests per minute
	return limiter.Limit()
}

// StandardRateLimit is a standard rate limiter for API endpoints
func StandardRateLimit() gin.HandlerFunc {
	limiter := NewRateLimiter(100) // 100 requests per minute
	return limiter.Limit()
}
