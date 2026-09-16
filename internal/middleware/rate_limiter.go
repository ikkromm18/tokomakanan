package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter manages per-IP token bucket rate limiters
type IPRateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*clientLimiter
	rps      rate.Limit
	burst    int
	stopOnce sync.Once
	stopChan chan struct{}
}

// NewIPRateLimiter creates a new IPRateLimiter with periodic idle cleanup
func NewIPRateLimiter(rps, burst int, cleanupInterval, maxIdle time.Duration) *IPRateLimiter {
	l := &IPRateLimiter{
		clients:  make(map[string]*clientLimiter),
		rps:      rate.Limit(rps),
		burst:    burst,
		stopChan: make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				l.cleanup(maxIdle)
			case <-l.stopChan:
				return
			}
		}
	}()

	return l
}

// GetLimiter returns or creates the rate.Limiter for a given IP
func (l *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	client, exists := l.clients[ip]
	if !exists {
		limiter := rate.NewLimiter(l.rps, l.burst)
		l.clients[ip] = &clientLimiter{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	client.lastSeen = time.Now()
	return client.limiter
}

func (l *IPRateLimiter) cleanup(maxIdle time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for ip, client := range l.clients {
		if now.Sub(client.lastSeen) > maxIdle {
			delete(l.clients, ip)
		}
	}
}

// Count returns the number of tracked IP addresses
func (l *IPRateLimiter) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.clients)
}

// Stop shuts down the background cleanup worker
func (l *IPRateLimiter) Stop() {
	l.stopOnce.Do(func() {
		close(l.stopChan)
	})
}

// RateLimiter returns a Gin middleware for IP rate limiting
func RateLimiter(rps, burst int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rps, burst, 5*time.Minute, 10*time.Minute)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		client := limiter.GetLimiter(ip)
		if !client.Allow() {
			response.Error(c, http.StatusTooManyRequests, "Too many requests, please try again later")
			c.Abort()
			return
		}
		c.Next()
	}
}
