package middleware

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// RequestLogger logs all requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Log request details
		logEntry := map[string]interface{}{
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"status":      statusCode,
			"method":      c.Request.Method,
			"path":        path,
			"query":       query,
			"latency":     latency.String(),
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"request_id":  getRequestID(c),
		}

		// Add user info if available
		if userID, exists := c.Get("user_id"); exists {
			logEntry["user_id"] = userID
		}

		// Log to stdout (in production, send to log aggregation)
		if statusCode >= 400 {
			log.Printf("ERROR: %s", formatLogEntry(logEntry))
		} else {
			log.Printf("INFO: %s", formatLogEntry(logEntry))
		}
	}
}

// Recovery handles panics and returns 500 error
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":   "internal server error",
					"message": "An unexpected error occurred",
				})
			}
		}()
		c.Next()
	}
}

// SecurityHeaders adds security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "SAMEORIGIN")
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		// XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")
		// Strict Transport Security
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// Cache control for sensitive data
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Header("Pragma", "no-cache")

		c.Next()
	}
}

// ZeroTrustAuth provides zero-trust authentication middleware
func ZeroTrustAuth(keycloakClient interface{}, accessControl interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health check and public endpoints
		if isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract and validate JWT token
		token, err := extractToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": err.Error(),
			})
			return
		}

		// Validate token
		claims, err := validateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid token",
			})
			return
		}

		// Check if token is expired
		if claims.ExpiresAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "token expired",
			})
			return
		}

		// Verify session is still valid (not revoked)
		if isSessionRevoked(claims.SessionID) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "session revoked",
			})
			return
		}

		// Check IP consistency (zero-trust)
		if !verifyIPConsistency(c, claims) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "IP address mismatch",
			})
			return
		}

		// Set user info in context
		c.Set("user_id", claims.Subject)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("permissions", claims.Permissions)
		c.Set("session_id", claims.SessionID)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequirePermission checks if user has required permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "no permissions found",
			})
			return
		}

		permList, ok := permissions.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "internal error",
				"message": "invalid permissions format",
			})
			return
		}

		if !hasPermission(permList, permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":      "forbidden",
				"message":    "insufficient permissions",
				"required":   permission,
				"permission": "Access denied",
			})
			return
		}

		c.Next()
	}
}

// RequireRole checks if user has required role
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "no role found",
			})
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "internal error",
				"message": "invalid role format",
			})
			return
		}

		for _, r := range roles {
			if roleStr == r {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":      "forbidden",
			"message":    "insufficient role",
			"required":   roles,
			"permission": "Access denied",
		})
	}
}

// AuditMiddleware logs all access to audit trail
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store start time
		start := time.Now()

		// Read request body for detailed logging
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = ioutil.ReadAll(c.Request.Body)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Process request
		c.Next()

		// Create audit event
		event := map[string]interface{}{
			"event_type":   mapRequestToEventType(c.Request.Method, c.Request.URL.Path),
			"actor_id":     getUserIDFromContext(c),
			"actor_type":   "USER",
			"action":       c.Request.Method,
			"resource":     c.Request.URL.Path,
			"outcome":      getOutcome(c.Writer.Status()),
			"ip_address":   c.ClientIP(),
			"user_agent":   c.Request.UserAgent(),
			"request_id":   getRequestID(c),
			"session_id":   getSessionIDFromContext(c),
			"latency_ms":   time.Since(start).Milliseconds(),
			"status_code":  c.Writer.Status(),
		}

		// Log the event
		log.Printf("AUDIT: %s", formatLogEntry(event))
	}
}

// RateLimiter provides rate limiting middleware
func RateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	// Simplified rate limiter - in production use Redis or similar
	return func(c *gin.Context) {
		// TODO: Implement proper rate limiting with sliding window
		c.Next()
	}
}

// CORS handles Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Helper functions

func isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/health",
		"/ready",
		"/metrics",
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
	}

	for _, endpoint := range publicEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}
	return false
}

func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

func validateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("YOUR_JWT_SECRET"), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func isSessionRevoked(sessionID string) bool {
	// TODO: Check against revoked sessions database
	return false
}

func verifyIPConsistency(c *gin.Context, claims *TokenClaims) bool {
	// In zero-trust mode, verify IP matches
	// Can be relaxed for legitimate proxy scenarios
	currentIP := c.ClientIP()
	storedIP := claims.IPAddress

	if storedIP == "" {
		return true // No IP stored, allow
	}

	return currentIP == storedIP
}

func hasPermission(permissions []string, required string) bool {
	for _, p := range permissions {
		if p == "*" || p == required {
			return true
		}
		if strings.HasSuffix(p, ":*") {
			prefix := strings.TrimSuffix(p, ":*")
			if strings.HasPrefix(required, prefix+":") {
				return true
			}
		}
	}
	return false
}

func mapRequestToEventType(method, path string) string {
	if strings.Contains(path, "/auth/") {
		return "AUTHENTICATION"
	}
	if strings.Contains(path, "/audit/") {
		return "DATA_ACCESS"
	}
	if strings.HasPrefix(path, "/api/") {
		if method == "POST" || method == "PUT" || method == "PATCH" {
			return "DATA_MODIFICATION"
		}
		return "DATA_ACCESS"
	}
	return "SYSTEM"
}

func getOutcome(statusCode int) string {
	if statusCode >= 200 && statusCode < 300 {
		return "SUCCESS"
	}
	if statusCode == 401 || statusCode == 403 {
		return "DENIED"
	}
	if statusCode >= 400 {
		return "FAILURE"
	}
	return "UNKNOWN"
}

func getUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}

func getSessionIDFromContext(c *gin.Context) string {
	if sessionID, exists := c.Get("session_id"); exists {
		return sessionID.(string)
	}
	return ""
}

func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

func generateRequestID() string {
	b := make([]byte, 16)
	base64.StdEncoding.EncodeToString(b)
	return strings.ReplaceAll(strings.ReplaceAll(string(b), "+", "-"), "/", "_")
}

func formatLogEntry(entry map[string]interface{}) string {
	data, _ := json.Marshal(entry)
	return string(data)
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	jwt.RegisteredClaims
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	SessionID   string   `json:"session_id"`
	IPAddress   string   `json:"ip_address"`
}
