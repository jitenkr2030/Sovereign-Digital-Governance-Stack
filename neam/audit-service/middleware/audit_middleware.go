package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditMiddlewareConfig provides configuration for audit middleware
type AuditMiddlewareConfig struct {
	// SkipPaths paths to skip auditing
	SkipPaths []string

	// IncludeRequestBody include request body in audit log
	IncludeRequestBody bool

	// IncludeResponseBody include response body in audit log
	IncludeResponseBody bool

	// SensitiveFields fields to mask in logs
	SensitiveFields []string

	// CustomEventType custom function to determine event type
	CustomEventType func(*http.Request) string
}

// AuditMiddleware creates audit logging middleware with configuration
func AuditMiddlewareConfigurable(config AuditMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if path is in skip list
		if shouldSkipPath(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		// Capture request details
		start := time.Now()
		var requestBody []byte

		if config.IncludeRequestBody && c.Request.Body != nil {
			requestBody, _ = ioutil.ReadAll(c.Request.Body)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create audit context
		auditCtx := &AuditContext{
			RequestID:    getRequestIDFromGin(c),
			TraceID:      getTraceID(c),
			SessionID:    getSessionIDFromGin(c),
			ActorID:      getActorIDFromGin(c),
			ActorType:    getActorType(c),
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Location:     getLocationFromRequest(c.Request),
			RequestPath:  c.Request.URL.Path,
			RequestMethod: c.Request.Method,
			StartTime:    start,
		}

		// Store audit context
		c.Set("audit_context", auditCtx)

		// Process request
		c.Next()

		// Capture response
		auditCtx.Latency = time.Since(start)
		auditCtx.StatusCode = c.Writer.Status()
		auditCtx.Outcome = determineOutcome(auditCtx.StatusCode)

		// Capture response body if configured
		if config.IncludeResponseBody {
			auditCtx.ResponseBody = c.Writer.Body.String()
		}

		// Create and log audit event
		event := createAuditEvent(auditCtx, config)
		if event != nil {
			logAuditEvent(event)
		}
	}
}

// AuditContext contains context for audit logging
type AuditContext struct {
	RequestID     string
	TraceID       string
	SessionID     string
	ActorID       string
	ActorType     string
	IPAddress     string
	UserAgent     string
	Location      string
	RequestPath   string
	RequestMethod string
	StartTime     time.Time
	Latency       time.Duration
	StatusCode    int
	Outcome       string
	ResponseBody  string
}

// shouldSkipPath checks if a path should be skipped
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath || path == skipPath+"*" {
			return true
		}
	}
	return false
}

// getRequestIDFromGin extracts request ID from gin context
func getRequestIDFromGin(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// getTraceID extracts trace ID from request
func getTraceID(c *gin.Context) string {
	return c.GetHeader("X-Trace-ID")
}

// getSessionIDFromGin extracts session ID from gin context
func getSessionIDFromGin(c *gin.Context) string {
	if sessionID, exists := c.Get("session_id"); exists {
		return sessionID.(string)
	}
	return ""
}

// getActorIDFromGin extracts actor ID from gin context
func getActorIDFromGin(c *gin.Context) string {
	if actorID, exists := c.Get("user_id"); exists {
		return actorID.(string)
	}
	return "anonymous"
}

// getActorType determines the type of actor
func getActorType(c *gin.Context) string {
	if _, exists := c.Get("service_account"); exists {
		return "SERVICE"
	}
	if actorID := getActorIDFromGin(c); actorID == "anonymous" {
		return "EXTERNAL"
	}
	return "USER"
}

// getLocationFromRequest extracts location from request
func getLocationFromRequest(r *http.Request) string {
	// Could use GeoIP lookup here
	// For now, return empty or extract from headers
	return r.Header.Get("X-Country-Code")
}

// determineOutcome determines the outcome based on status code
func determineOutcome(statusCode int) string {
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

// createAuditEvent creates an audit event from the context
func createAuditEvent(ctx *AuditContext, config AuditMiddlewareConfig) *AuditEvent {
	// Determine event type
	eventType := determineEventType(ctx, config)

	// Mask sensitive fields if configured
	maskedBody := ""
	if ctx.RequestBody != "" && len(config.SensitiveFields) > 0 {
		maskedBody = maskSensitiveData(ctx.RequestBody, config.SensitiveFields)
	}

	event := &AuditEvent{
		ID:            generateEventID(),
		Timestamp:     ctx.StartTime,
		EventType:     eventType,
		ActorID:       ctx.ActorID,
		ActorType:     ctx.ActorType,
		Action:        ctx.RequestMethod,
		Resource:      ctx.RequestPath,
		Outcome:       ctx.Outcome,
		IPAddress:     ctx.IPAddress,
		UserAgent:     ctx.UserAgent,
		Location:      ctx.Location,
		SessionID:     ctx.SessionID,
		IntegrityHash: "", // Will be calculated by audit service
	}

	// Add details
	details := map[string]interface{}{
		"request_id":  ctx.RequestID,
		"trace_id":    ctx.TraceID,
		"latency_ms":  ctx.Latency.Milliseconds(),
		"status_code": ctx.StatusCode,
	}

	if maskedBody != "" {
		details["request_body_hash"] = hashData(maskedBody)
	}

	event.Details = details

	return event
}

// determineEventType determines the audit event type
func determineEventType(ctx *AuditContext, config AuditMiddlewareConfig) string {
	if config.CustomEventType != nil {
		return config.CustomEventType(nil)
	}

	// Default event type determination
	switch {
	case contains(ctx.RequestPath, "/auth/"):
		return "AUTHENTICATION"
	case contains(ctx.RequestPath, "/policy/"):
		return "POLICY"
	case contains(ctx.RequestPath, "/audit/") && ctx.RequestMethod != "GET":
		return "DATA_MODIFICATION"
	case contains(ctx.RequestPath, "/audit/"):
		return "DATA_ACCESS"
	case ctx.RequestMethod == "DELETE":
		return "DATA_DELETION"
	case ctx.RequestMethod == "GET":
		return "DATA_ACCESS"
	case ctx.RequestMethod == "POST" || ctx.RequestMethod == "PUT" || ctx.RequestMethod == "PATCH":
		return "DATA_MODIFICATION"
	default:
		return "SYSTEM"
	}
}

// maskSensitiveData masks sensitive fields in data
func maskSensitiveData(data string, sensitiveFields []string) string {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return data
	}

	for _, field := range sensitiveFields {
		if val, ok := result[field]; ok {
			result[field] = "***MASKED***"
		}
	}

	jsonData, _ := json.Marshal(result)
	return string(jsonData)
}

// hashData calculates SHA-256 hash of data
func hashData(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// logAuditEvent logs the audit event (to be connected to audit service)
func logAuditEvent(event *AuditEvent) {
	// In production, this would send to the audit service
	eventJSON, _ := json.Marshal(event)
	fmt.Printf("AUDIT_EVENT: %s\n", string(eventJSON))
}

// AuditEvent represents an audit event
type AuditEvent struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	EventType     string                 `json:"event_type"`
	ActorID       string                 `json:"actor_id"`
	ActorType     string                 `json:"actor_type"`
	Action        string                 `json:"action"`
	Resource      string                 `json:"resource"`
	Outcome       string                 `json:"outcome"`
	Details       map[string]interface{} `json:"details,omitempty"`
	IPAddress     string                 `json:"ip_address"`
	UserAgent     string                 `json:"user_agent"`
	Location      string                 `json:"location"`
	SessionID     string                 `json:"session_id"`
	IntegrityHash string                 `json:"integrity_hash"`
}

// GenerateEventID generates a unique event ID
func generateEventID() string {
	hash := sha256.Sum256([]byte(time.Now().String() + randomString(16)))
	return hex.EncodeToString(hash[:8])
}

// RandomString generates a random string
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ForensicAuditMiddleware provides specialized middleware for forensic investigation
func ForensicAuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Capture all request details for forensic investigation
		start := time.Now()

		// Read complete request
		body, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		// Store original request
		originalRequest := &ForensicRequestData{
			Timestamp:     start,
			Method:        c.Request.Method,
			Path:          c.Request.URL.Path,
			Query:         c.Request.URL.RawQuery,
			Headers:       copyHeaders(c.Request.Header),
			Body:          string(body),
			ClientIP:      c.ClientIP(),
			UserAgent:     c.Request.UserAgent,
			ContentLength: c.Request.ContentLength,
		}

		c.Set("forensic_request", originalRequest)
		c.Next()

		// Capture response
		originalRequest.Latency = time.Since(start)
		originalRequest.StatusCode = c.Writer.Status()

		// Store for forensic investigation
		if forensicData, exists := c.Get("forensic_request"); exists {
			c.Set("forensic_data", forensicData)
		}
	}
}

// ForensicRequestData contains complete request data for forensics
type ForensicRequestData struct {
	Timestamp     time.Time
	Method        string
	Path          string
	Query         string
	Headers       map[string][]string
	Body          string
	ClientIP      string
	UserAgent     string
	ContentLength int64
	Latency       time.Duration
	StatusCode    int
}

// copyHeaders copies request headers
func copyHeaders(header map[string][]string) map[string][]string {
	result := make(map[string][]string)
	for k, v := range header {
		result[k] = make([]string, len(v))
		copy(result[k], v)
	}
	return result
}

// ZeroTrustDecisionMiddleware enforces zero-trust decisions
func ZeroTrustDecisionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create risk assessment
		riskAssessment := &RiskAssessment{
			Timestamp: time.Now(),
			Score:     0,
			Factors:   []RiskFactor{},
		}

		// Assess IP reputation
		ipRisk := assessIPRisk(c.ClientIP())
		riskAssessment.Score += ipRisk.Score
		riskAssessment.Factors = append(riskAssessment.Factors, ipRisk)

		// Assess user behavior
		userRisk := assessUserRisk(c)
		riskAssessment.Score += userRisk.Score
		riskAssessment.Factors = append(riskAssessment.Factors, userRisk)

		// Assess request pattern
		patternRisk := assessPatternRisk(c)
		riskAssessment.Score += patternRisk.Score
		riskAssessment.Factors = append(riskAssessment.Factors, patternRisk)

		// Store risk assessment
		c.Set("risk_assessment", riskAssessment)

		// Enforce decision
		if riskAssessment.Score > 100 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "request blocked due to security policy",
				"risk":    riskAssessment,
			})
			return
		}

		// Log risk assessment
		logRiskAssessment(riskAssessment)

		c.Next()
	}
}

// RiskAssessment represents a security risk assessment
type RiskAssessment struct {
	Timestamp    time.Time
	Score        int
	Factors      []RiskFactor
	Recommendation string
}

// RiskFactor represents a factor in risk assessment
type RiskFactor struct {
	Name        string
	Score       int
	Description string
}

// assessIPRisk assesses risk based on IP address
func assessIPRisk(ip string) RiskFactor {
	// Simplified - in production use IP reputation service
	return RiskFactor{
		Name:        "IP Reputation",
		Score:       0,
		Description: "IP address reputation assessment",
	}
}

// assessUserRisk assesses risk based on user behavior
func assessUserRisk(c *gin.Context) RiskFactor {
	return RiskFactor{
		Name:        "User Behavior",
		Score:       0,
		Description: "User behavior analysis",
	}
}

// assessPatternRisk assesses risk based on request patterns
func assessPatternRisk(c *gin.Context) RiskFactor {
	return RiskFactor{
		Name:        "Request Pattern",
		Score:       0,
		Description: "Request pattern analysis",
	}
}

// logRiskAssessment logs the risk assessment
func logRiskAssessment(assessment *RiskAssessment) {
	data, _ := json.Marshal(assessment)
	fmt.Printf("RISK_ASSESSMENT: %s\n", string(data))
}
