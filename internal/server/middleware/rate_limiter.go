package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateLimiter provides simple in-memory rate limiting
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int           // max requests per window
	window   time.Duration // window duration
}

type visitor struct {
	lastSeen time.Time
	count    int
}

// NewRateLimiter creates a new in-memory rate limiter
func NewRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}

	// Background cleanup every 10 minutes
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			rl.cleanup()
		}
	}()

	return rl
}

// cleanup removes expired visitors
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, v := range rl.visitors {
		if now.Sub(v.lastSeen) > rl.window {
			delete(rl.visitors, key)
		}
	}
}

// getVisitor retrieves or creates a visitor entry
func (rl *rateLimiter) getVisitor(key string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	if !exists {
		v = &visitor{lastSeen: time.Now(), count: 0}
		rl.visitors[key] = v
	}
	return v
}

// Limit returns a Gin middleware handler that enforces rate limits
func (rl *rateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			key = c.ClientIP()
		}

		v := rl.getVisitor(key)
		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now()
		if now.Sub(v.lastSeen) > rl.window {
			v.count = 0
			v.lastSeen = now
		}

		v.count++
		if v.count > rl.limit {
			resetIn := rl.window - time.Since(v.lastSeen)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded. Please try again later.",
				"limit":       rl.limit,
				"window":      rl.window.String(),
				"reset_in_s":  int(resetIn.Seconds()),
				"identifier":  key,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetStatus returns remaining requests and time until reset for a given key
func (rl *rateLimiter) GetStatus(key string) (remaining int, reset time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	if !exists {
		// Never seen this key before — full quota available
		return rl.limit, rl.window
	}

	elapsed := time.Since(v.lastSeen)
	if elapsed > rl.window {
		return rl.limit, rl.window
	}

	return rl.limit - v.count, rl.window - elapsed
}
