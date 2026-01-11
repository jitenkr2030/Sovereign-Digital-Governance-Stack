package middleware

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/csic-platform/shared/logger"
)

// Logger interface for logging (compatible with shared logger)
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// SharedLoggerAdapter adapts the shared logger to our Logger interface
type SharedLoggerAdapter struct {
	*logger.Logger
}

func (a *SharedLoggerAdapter) Info(msg string, fields ...interface{}) {
	a.Logger.Info(msg)
}

func (a *SharedLoggerAdapter) Error(msg string, fields ...interface{}) {
	a.Logger.Error(msg)
}

func (a *SharedLoggerAdapter) Warn(msg string, fields ...interface{}) {
	a.Logger.Warn(msg)
}

// RateLimitMiddleware handles rate limiting
type RateLimitMiddleware struct {
	logger    Logger
	requests  map[string][]time.Time
	limit     int
	window    time.Duration
	mu        sync.RWMutex
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(logger Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		logger:   logger,
		requests: make(map[string][]time.Time),
		limit:    100,           // 100 requests per window
		window:   time.Minute,  // 1 minute window
	}
}

// Middleware returns the rate limiting middleware
func (rl *RateLimitMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()

		now := time.Now()
		windowStart := now.Add(-rl.window)

		// Filter out old requests
		requests := rl.requests[ip]
		rl.requests[ip] = filterRequests(requests, windowStart)

		// Check rate limit
		if len(rl.requests[ip]) >= rl.limit {
			rl.mu.Unlock()

			c.Header("Retry-After", "60")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Too many requests. Please try again later.",
				},
			})
			c.Abort()
			return
		}

		// Add current request
		rl.requests[ip] = append(rl.requests[ip], now)
		rl.mu.Unlock()

		c.Next()
	}
}

// filterRequests removes requests outside the time window
func filterRequests(requests []time.Time, since time.Time) []time.Time {
	filtered := make([]time.Time, 0, len(requests))
	for _, req := range requests {
		if req.After(since) {
			filtered = append(filtered, req)
		}
	}
	return filtered
}

// SetLimit sets the rate limit
func (rl *RateLimitMiddleware) SetLimit(limit int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.limit = limit
}

// SetWindow sets the rate limit window
func (rl *RateLimitMiddleware) SetWindow(window time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.window = window
}

// GetCurrentRequests returns the current request count for an IP
func (rl *RateLimitMiddleware) GetCurrentRequests(ip string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)
	requests := rl.requests[ip]

	count := 0
	for _, req := range requests {
		if req.After(windowStart) {
			count++
		}
	}
	return count
}

// LoggingMiddleware logs all incoming requests
type LoggingMiddleware struct {
	logger Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Middleware returns the logging middleware
func (l *LoggingMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if query != "" {
			path = path + "?" + query
		}

		// Log using standard log (could be enhanced to use structured logging)
		log.Printf("[GIN] %s | %3d | %13v | %15s | %-7s %s",
			time.Now().Format("2006/01/02 - 15:04:05"),
			status,
			latency,
			c.ClientIP(),
			c.Request.Method,
			path,
		)

		// Also log using the logger if available
		if l.logger != nil {
			l.logger.Info("request completed",
				logger.Int("status", status),
				logger.String("method", c.Request.Method),
				logger.String("path", path),
				logger.String("ip", c.ClientIP()),
				logger.Duration("latency", latency),
			)
		}
	}
}

// SecurityHeadersMiddleware adds security headers to all responses
type SecurityHeadersMiddleware struct{}

// NewSecurityHeadersMiddleware creates a new security headers middleware
func NewSecurityHeadersMiddleware() *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{}
}

// Headers returns the security headers middleware
func (s *SecurityHeadersMiddleware) Headers() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")

		c.Next()
	}
}

// CORSMiddleware handles CORS
type CORSMiddleware struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: []string{"*"},
		allowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		allowedHeaders: []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
	}
}

// Middleware returns the CORS middleware
func (c *CORSMiddleware) Middleware() gin.HandlerFunc {
	return func(cxt *gin.Context) {
		cxt.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		cxt.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(c.allowedMethods, ", "))
		cxt.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(c.allowedHeaders, ", "))
		cxt.Writer.Header().Set("Access-Control-Max-Age", "86400")
		cxt.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if cxt.Request.Method == "OPTIONS" {
			cxt.AbortWithStatus(http.StatusNoContent)
			return
		}

		cxt.Next()
	}
}
