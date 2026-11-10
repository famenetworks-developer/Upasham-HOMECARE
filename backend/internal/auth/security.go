package auth

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type rateBucket struct {
	count int
	reset time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]rateBucket
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{buckets: make(map[string]rateBucket), limit: limit, window: window}
}

func (r *RateLimiter) Middleware(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		key := scope + ":" + c.ClientIP()
		r.mu.Lock()
		bucket := r.buckets[key]
		if bucket.reset.IsZero() || now.After(bucket.reset) {
			bucket = rateBucket{reset: now.Add(r.window)}
		}
		bucket.count++
		r.buckets[key] = bucket
		remaining := r.limit - bucket.count
		reset := bucket.reset
		r.mu.Unlock()

		c.Header("X-RateLimit-Limit", fmtInt(r.limit))
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Remaining", fmtInt(remaining))
		c.Header("X-RateLimit-Reset", fmtInt(int(time.Until(reset).Seconds())))
		if bucket.count > r.limit {
			c.Header("Retry-After", fmtInt(int(time.Until(reset).Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please try again shortly."})
			return
		}
		c.Next()
	}
}

func fmtInt(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	buf := [20]byte{}
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.NewString()
		}
		c.Set("correlation_id", correlationID)
		c.Header("X-Correlation-ID", correlationID)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(self), microphone=(self), geolocation=()")
		c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'; script-src 'self' https://meet.jit.si; frame-src https:; connect-src 'self' https: wss:")
		if c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}
