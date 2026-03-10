package middlewares

import (
	"net/http"
	"sync"
	"time"

	"backend/utils"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// visitor represents a rate limiter for a specific IP address
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// rateLimiterStore manages rate limiters for different IP addresses
type rateLimiterStore struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// newRateLimiterStore creates a new rate limiter store
func newRateLimiterStore(r rate.Limit, b int) *rateLimiterStore {
	store := &rateLimiterStore{
		visitors: make(map[string]*visitor),
		rate:     r,
		burst:    b,
	}

	// Clean up old visitors every 5 minutes
	go store.cleanupVisitors()

	return store
}

// getVisitor retrieves or creates a rate limiter for an IP address
func (s *rateLimiterStore) getVisitor(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, exists := s.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(s.rate, s.burst)
		s.visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupVisitors removes visitors that haven't been seen in the last 10 minutes
func (s *rateLimiterStore) cleanupVisitors() {
	for {
		time.Sleep(5 * time.Minute)

		s.mu.Lock()
		for ip, v := range s.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(s.visitors, ip)
			}
		}
		s.mu.Unlock()
	}
}

// GetIPAddress extracts IP address from request
func GetIPAddress(c echo.Context) string {
	// Check X-Forwarded-For header (if behind proxy)
	ip := c.Request().Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For can contain multiple IPs, get the first one
		return ip
	}

	// Check X-Real-IP header
	ip = c.Request().Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Fall back to RemoteAddr
	return c.RealIP()
}

// RateLimitMiddleware creates a rate limiting middleware with specified limits
func RateLimitMiddleware(requestsPerMinute int, burst int) echo.MiddlewareFunc {
	// Convert requests per minute to rate.Limit (requests per second)
	rateLimit := rate.Limit(float64(requestsPerMinute) / 60.0)
	store := newRateLimiterStore(rateLimit, burst)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := GetIPAddress(c)
			limiter := store.getVisitor(ip)

			if !limiter.Allow() {
				return utils.ErrorResponse(c, utils.NewAppError(
					utils.ErrInternal,
					"Too many requests. Please try again later.",
					http.StatusTooManyRequests,
					"Rate limit exceeded for IP: "+ip,
				))
			}

			return next(c)
		}
	}
}

// AuthRateLimiter creates a rate limiter for auth endpoints
// Allows 5 requests per minute per IP
func AuthRateLimiter() echo.MiddlewareFunc {
	return RateLimitMiddleware(5, 5)
}

// StrictAuthRateLimiter creates a stricter rate limiter for sensitive operations
// Allows 3 requests per 5 minutes per IP (0.6 requests per minute)
func StrictAuthRateLimiter() echo.MiddlewareFunc {
	// For 3 requests per 5 minutes, that's 0.6 requests per minute
	// We'll use 1 request per minute with burst of 3 to achieve similar behavior
	rateLimit := rate.Limit(1.0 / 60.0) // 1 request per minute
	store := newRateLimiterStore(rateLimit, 3)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := GetIPAddress(c)
			limiter := store.getVisitor(ip)

			if !limiter.Allow() {
				return utils.ErrorResponse(c, utils.NewAppError(
					utils.ErrInternal,
					"Too many requests. Please try again later.",
					http.StatusTooManyRequests,
					"Strict rate limit exceeded for IP: "+ip,
				))
			}

			return next(c)
		}
	}
}
