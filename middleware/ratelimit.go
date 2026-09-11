package middleware

import (
"net/http"
"sync"
"time"

"github.com/nexrouter/nexrouter/core"
)

// RateLimiter implements per-key token bucket rate limiting
type RateLimiter struct {
tokens map[string]*bucket
mu     sync.RWMutex
rate   float64
burst  int
}

type bucket struct {
tokens    float64
lastCheck time.Time
}

// NewRateLimiter creates a limiter: rate = tokens added per second, burst = max tokens
func NewRateLimiter(rate float64, burst int) *RateLimiter {
rl := &RateLimiter{tokens: make(map[string]*bucket), rate: rate, burst: burst}
go func() {
for {
time.Sleep(time.Minute)
rl.mu.Lock()
now := time.Now()
for key, b := range rl.tokens {
if now.Sub(b.lastCheck) > 10*time.Minute {
delete(rl.tokens, key)
}
}
rl.mu.Unlock()
}
}()
return rl
}

// Allow reports whether one request is allowed for key
func (rl *RateLimiter) Allow(key string) bool {
rl.mu.Lock()
defer rl.mu.Unlock()
now := time.Now()
b, exists := rl.tokens[key]
if !exists {
rl.tokens[key] = &bucket{tokens: float64(rl.burst) - 1, lastCheck: now}
return true
}
elapsed := now.Sub(b.lastCheck).Seconds()
b.tokens += elapsed * rl.rate
if b.tokens > float64(rl.burst) {
b.tokens = float64(rl.burst)
}
b.lastCheck = now
if b.tokens < 1 {
return false
}
b.tokens--
return true
}

// RateLimit returns middleware limiting requests per window per IP
func RateLimit(requests int, window time.Duration) core.HandlerFunc {
limiter := NewRateLimiter(float64(requests)/window.Seconds(), requests)
return func(c *core.Context) {
if !limiter.Allow(c.ClientIP()) {
c.SetHeader("Retry-After", "60")
c.JSON(http.StatusTooManyRequests, core.H{
"error":       "rate limit exceeded",
"retry_after": int(window.Seconds()),
})
c.Abort()
return
}
c.Next()
}
}