// API Gateway Middleware - Authentication, logging, and rate limiting
// HTTP middleware components for the CSIC Platform API

package middleware

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// ContextUserKey is the context key for user information
	ContextUserKey ContextKey = "user"
	// ContextRequestIDKey is the context key for request ID
	ContextRequestIDKey ContextKey = "request_id"
	// ContextSessionIDKey is the context key for session ID
	ContextSessionIDKey ContextKey = "session_id"
)

// User represents authenticated user information
type User struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	Name      string   `json:"name"`
	Role      string   `json:"role"`
	Permissions []string `json:"permissions"`
	SessionID string   `json:"session_id"`
}

// RequestInfo contains metadata about the request
type RequestInfo struct {
	RequestID   string
	SessionID   string
	UserID      string
	IPAddress   string
	UserAgent   string
	Method      string
	Path        string
	Timestamp   time.Time
}

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	logger     *zap.Logger
	jwtSecret  []byte
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(logger *zap.Logger, jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		logger:    logger,
		jwtSecret: []byte(jwtSecret),
	}
}

// Authenticate validates JWT tokens and populates user context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Allow public routes
			next.ServeHTTP(w, r)
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			m.writeError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		token := parts[1]

		// Validate token (simplified - in production use proper JWT library)
		user, err := m.validateToken(token)
		if err != nil {
			m.logger.Warn("token validation failed",
				zap.Error(err),
				zap.String("path", r.URL.Path))
			m.writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), ContextUserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth requires authentication and returns 401 if not authenticated
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(ContextUserKey)
		if user == nil {
			m.writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePermission checks if user has required permission
func (m *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(ContextUserKey).(*User)
			if !ok {
				m.writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			// Check permission
			hasPermission := false
			for _, p := range user.Permissions {
				if p == permission || p == "write:all" {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				m.logger.Warn("permission denied",
					zap.String("user_id", user.ID),
					zap.String("required_permission", permission),
					zap.String("path", r.URL.Path))
				m.writeError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole checks if user has required role
func (m *AuthMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(ContextUserKey).(*User)
			if !ok {
				m.writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			hasRole := false
			for _, role := range roles {
				if subtle.ConstantTimeCompare([]byte(user.Role), []byte(role)) == 1 {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.writeError(w, http.StatusForbidden, "insufficient role")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// validateToken validates a JWT token and returns user info
func (m *AuthMiddleware) validateToken(token string) (*User, error) {
	// Simplified token validation
	// In production, use proper JWT library with signature verification
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode payload (simplified - in production verify signature)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode token payload: %w", err)
	}

	var user User
	if err := json.Unmarshal(payload, &user); err != nil {
		return nil, fmt.Errorf("failed to parse token payload: %w", err)
	}

	return &user, nil
}

// generateToken generates a JWT token for a user
func (m *AuthMiddleware) generateToken(user *User, expiry time.Time) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	payload := map[string]interface{}{
		"sub":         user.ID,
		"email":       user.Email,
		"name":        user.Name,
		"role":        user.Role,
		"permissions": user.Permissions,
		"exp":         expiry.Unix(),
		"iat":         time.Now().Unix(),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// In production, compute proper HMAC signature
	signature := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%s.%s.%s", header, payloadB64, m.jwtSecret)))

	return fmt.Sprintf("%s.%s.%s", header, payloadB64, signature), nil
}

func (m *AuthMiddleware) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   message,
		"status":  strconv.Itoa(status),
	})
}

// LoggingMiddleware provides request logging
type LoggingMiddleware struct {
	logger *zap.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

// Log logs HTTP requests and responses
func (m *LoggingMiddleware) Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create response wrapper to capture status code
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)

		// Log request
		m.logger.Info("http request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.Int("status", lrw.statusCode),
			zap.Duration("duration", duration),
			zap.String("ip", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.Int("bytes", lrw.bytesWritten),
		)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}

// RateLimitMiddleware provides rate limiting
type RateLimitMiddleware struct {
	logger    *zap.Logger
	limits    map[string]*RateLimiter
}

// RateLimiter tracks request counts for rate limiting
type RateLimiter struct {
	count     int
	window    time.Duration
	limit     int
	resetTime time.Time
	mu        sync.Mutex
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(logger *zap.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		logger: logger,
		limits: make(map[string]*RateLimiter),
	}
}

// Limit creates a rate limiting middleware with specified limits
func (m *RateLimitMiddleware) Limit(requests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP as key
			key := r.RemoteAddr

			limiter := m.getLimiter(key, requests, window)

			if !limiter.Allow() {
				m.logger.Warn("rate limit exceeded",
					zap.String("ip", key),
					zap.Int("limit", requests),
					zap.Duration("window", window))
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
				m.writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getLimiter gets or creates a rate limiter for a key
func (m *RateLimitMiddleware) getLimiter(key string, limit int, window time.Duration) *RateLimiter {
	// In production, use Redis for distributed rate limiting
	m.mu.Lock()
	defer m.mu.Unlock()

	limiter, exists := m.limits[key]
	if !exists {
		limiter = &RateLimiter{
			count:     0,
			limit:     limit,
			window:    window,
			resetTime: time.Now().Add(window),
		}
		m.limits[key] = limiter
	}

	// Reset if window expired
	if time.Now().After(limiter.resetTime) {
		limiter.count = 0
		limiter.resetTime = time.Now().Add(window)
	}

	return limiter
}

// Allow checks if a request is allowed
func (l *RateLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.count >= l.limit {
		return false
	}
	l.count++
	return true
}

func (m *RateLimitMiddleware) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":  message,
		"status": strconv.Itoa(status),
	})
}

// CORSMiddleware handles Cross-Origin Resource Sharing
type CORSMiddleware struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(allowedOrigins []string) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: allowedOrigins,
		allowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		allowedHeaders: []string{"Authorization", "Content-Type", "X-Request-ID", "X-Correlation-ID"},
	}
}

// CORS adds CORS headers to responses
func (m *CORSMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, o := range m.allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(m.allowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(m.allowedHeaders, ", "))
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds security headers
type SecurityHeadersMiddleware struct{}

// NewSecurityHeadersMiddleware creates a new security headers middleware
func NewSecurityHeadersMiddleware() *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{}
}

// Headers adds security headers to responses
func (m *SecurityHeadersMiddleware) Headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent XSS attacks
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")

		// Strict Transport Security (HTTPS only)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// Import required packages
import (
	"sync"
)
