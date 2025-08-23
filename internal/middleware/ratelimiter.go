package middleware

import (
	"fmt"
	"net/http"
	"provolo-api/internal/types"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	RequestsPerMinute int
	BurstSize         int
	CleanupInterval   time.Duration
}

// DefaultRateLimiterConfig returns a sensible default configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerMinute: 60, // 60 requests per minute
		BurstSize:         10, // Allow burst of 10 requests
		CleanupInterval:   time.Minute * 5,
	}
}

// StrictRateLimiterConfig returns a more restrictive configuration
func StrictRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerMinute: 30, // 30 requests per minute
		BurstSize:         5,  // Allow burst of 5 requests
		CleanupInterval:   time.Minute * 5,
	}
}

// clientInfo stores rate limiting information for each client
type clientInfo struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	config  RateLimiterConfig
	clients map[string]*clientInfo
	mu      sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		config:  config,
		clients: make(map[string]*clientInfo),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup removes old client entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, client := range rl.clients {
			client.mu.Lock()
			// Remove clients that haven't been active for 10 minutes
			if now.Sub(client.lastRefill) > time.Minute*10 {
				delete(rl.clients, ip)
			}
			client.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// getClientIP extracts the real client IP from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return c.ClientIP()
}

// Allow checks if a request should be allowed and updates the token bucket
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.RLock()
	client, exists := rl.clients[ip]
	rl.mu.RUnlock()

	if !exists {
		// Create new client with full token bucket
		client = &clientInfo{
			tokens:     rl.config.BurstSize,
			lastRefill: time.Now(),
		}
		rl.mu.Lock()
		rl.clients[ip] = client
		rl.mu.Unlock()
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	now := time.Now()
	timePassed := now.Sub(client.lastRefill)

	// Refill tokens based on time passed
	tokensToAdd := int(timePassed.Minutes() * float64(rl.config.RequestsPerMinute))
	if tokensToAdd > 0 {
		client.tokens += tokensToAdd
		if client.tokens > rl.config.BurstSize {
			client.tokens = rl.config.BurstSize
		}
		client.lastRefill = now
	}

	// Check if request can be allowed
	if client.tokens > 0 {
		client.tokens--
		return true
	}

	return false
}

// RateLimiterMiddleware returns a Gin middleware function for rate limiting
func RateLimiterMiddleware(config RateLimiterConfig) gin.HandlerFunc {
	rl := NewRateLimiter(config)

	return func(c *gin.Context) {
		clientIP := getClientIP(c)

		if !rl.Allow(clientIP) {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "60")

			c.JSON(http.StatusTooManyRequests, types.NewErrorResponse(
				"Rate Limit Exceeded",
				"Too many requests. Please try again later.",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// GlobalRateLimiter creates a rate limiter for all routes
func GlobalRateLimiter() gin.HandlerFunc {
	return RateLimiterMiddleware(DefaultRateLimiterConfig())
}

// StrictRateLimiter creates a more restrictive rate limiter
func StrictRateLimiter() gin.HandlerFunc {
	return RateLimiterMiddleware(StrictRateLimiterConfig())
}
