package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateBucket tracks requests per IP using a simple token bucket.
type rateBucket struct {
	mu       sync.Mutex
	tokens   float64
	maxTokens float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

func newRateBucket(rps float64, burst int) *rateBucket {
	return &rateBucket{
		tokens:    float64(burst),
		maxTokens: float64(burst),
		refillRate: rps,
		lastRefill: time.Now(),
	}
}

func (b *rateBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// RateLimiter returns a middleware that limits requests per second per IP.
func RateLimiter(rps float64, burst int) gin.HandlerFunc {
	limiters := &sync.Map{}
	return func(c *gin.Context) {
		ip := c.ClientIP()
		val, _ := limiters.LoadOrStore(ip, newRateBucket(rps, burst))
		bucket := val.(*rateBucket)
		if !bucket.allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"code": 500101, "msg": "rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}
