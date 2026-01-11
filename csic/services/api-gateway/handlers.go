// CSIC Platform - Backend API Handlers
// Production-grade HTTP handlers for the CSIC API Gateway

package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/csic-platform/shared/config"
	"github.com/csic-platform/shared/errors"
	"github.com/csic-platform/shared/validation"
	"github.com/gin-gonic/gin"
)

// Handler contains all HTTP handlers for the API
type Handler struct {
	cfg *config.Config
}

// NewHandler creates a new handler instance
func NewHandler(cfg *config.Config) *Handler {
	return &Handler{cfg: cfg}
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *MetaInfo   `json:"meta,omitempty"`
}

// ErrorInfo represents error details
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// MetaInfo represents pagination metadata
type MetaInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PaginationParams represents pagination query parameters
type PaginationParams struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

// HealthCheck returns the health status of the service
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   h.cfg.App.Version,
		},
	})
}

// GetDashboardStats returns dashboard statistics
func (h *Handler) GetDashboardStats(c *gin.Context) {
	// Mock data - in production, this would aggregate from multiple services
	stats := gin.H{
		"system_status": gin.H{
			"status":       "ONLINE",
			"uptime":       time.Since(time.Now().Add(-15 * 24 * time.Hour)).Seconds(),
			"last_heartbeat": time.Now().UTC().Format(time.RFC3339),
		},
		"metrics": gin.H{
			"exchanges": gin.H{
				"total":      24,
				"active":     20,
				"suspended":  2,
				"revoked":    2,
			},
			"transactions": gin.H{
				"total_24h":   15420,
				"volume_24h":  2500000000,
				"flagged_24h": 12,
			},
			"wallets": gin.H{
				"total":   1560,
				"frozen":  23,
				"blacklisted": 45,
			},
			"miners": gin.H{
				"total":       156,
				"online":      142,
				"total_hashrate": 450,
			},
		},
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    stats,
	})
}

// Alerts

// GetAlerts returns a paginated list of alerts
func (h *Handler) GetAlerts(c *gin.Context) {
	var params PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		params.Page = 1
		params.PageSize = 20
	}

	// Mock alerts data
	alerts := []gin.H{
		{
			"id":          "alert_001",
			"title":       "高风险交易模式检测",
			"description": "检测到与已知混币服务相关的异常交易模式",
			"severity":    "CRITICAL",
			"status":      "ACTIVE",
			"category":    "交易监控",
			"source":      "风险引擎",
			"created_at":  time.Now().UTC().Format(time.RFC3339),
		},
		{
			"id":          "alert_002",
			"title":       "交易所合规分数下降",
			"description": "CryptoExchange Pro 的合规分数从 92 下降到 78",
			"severity":    "WARNING",
			"status":      "ACTIVE",
			"category":    "合规监控",
			"source":      "合规系统",
			"created_at":  time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"items": alerts,
		},
		Meta: &MetaInfo{
			Page:       params.Page,
			PageSize:   params.PageSize,
			Total:      int64(len(alerts)),
			TotalPages: 1,
		},
	})
}

// GetAlertByID returns a specific alert by ID
func (h *Handler) GetAlertByID(c *gin.Context) {
	alertID := c.Param("id")

	alert := gin.H{
		"id":          alertID,
		"title":       "高风险交易模式检测",
		"description": "检测到与已知混币服务相关的异常交易模式，涉及地址 1A2B3C4D5E",
		"severity":    "CRITICAL",
		"status":      "ACTIVE",
		"category":    "交易监控",
		"source":      "风险引擎",
		"created_at":  time.Now().UTC().Format(time.RFC3339),
		"evidence": []gin.H{
			{"type": "RISK_SCORE", "value": 95, "threshold": 80},
			{"type": "PATTERN_MATCH", "value": "MIXER_USAGE", "threshold": 1},
		},
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    alert,
	})
}

// AcknowledgeAlert acknowledges an alert
func (h *Handler) AcknowledgeAlert(c *gin.Context) {
	alertID := c.Param("id")

	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	log.Printf("Alert %s acknowledged by user %s", alertID, req.UserID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"id":             alertID,
			"status":         "ACKNOWLEDGED",
			"acknowledged_by": req.UserID,
			"acknowledged_at": time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// Exchanges

// GetExchanges returns a paginated list of exchanges
func (h *Handler) GetExchanges(c *gin.Context) {
	var params PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		params.Page = 1
		params.PageSize = 20
	}

	exchanges := []gin.H{
		{
			"id":              "ex_001",
			"name":            "CryptoExchange Pro",
			"license_number":  "CSIC-2024-001",
			"status":          "ACTIVE",
			"jurisdiction":    "Singapore",
			"compliance_score": 92,
			"risk_level":      "LOW",
			"registration_date": "2024-01-15T00:00:00Z",
			"last_audit":      "2024-11-01T00:00:00Z",
		},
		{
			"id":              "ex_002",
			"name":            "Digital Asset Hub",
			"license_number":  "CSIC-2024-002",
			"status":          "ACTIVE",
			"jurisdiction":    "Switzerland",
			"compliance_score": 88,
			"risk_level":      "LOW",
			"registration_date": "2024-02-20T00:00:00Z",
			"last_audit":      "2024-10-15T00:00:00Z",
		},
		{
			"id":              "ex_003",
			"name":            "BlockTrade Global",
			"license_number":  "CSIC-2024-003",
			"status":          "SUSPENDED",
			"jurisdiction":    "British Virgin Islands",
			"compliance_score": 45,
			"risk_level":      "HIGH",
			"registration_date": "2024-03-10T00:00:00Z",
			"last_audit":      "2024-09-20T00:00:00Z",
		},
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"items": exchanges,
		},
		Meta: &MetaInfo{
			Page:       params.Page,
			PageSize:   params.PageSize,
			Total:      int64(len(exchanges)),
			TotalPages: 1,
		},
	})
}

// GetExchangeByID returns a specific exchange by ID
func (h *Handler) GetExchangeByID(c *gin.Context) {
	exchangeID := c.Param("id")

	exchange := gin.H{
		"id":              exchangeID,
		"name":            "CryptoExchange Pro",
		"license_number":  "CSIC-2024-001",
		"status":          "ACTIVE",
		"jurisdiction":    "Singapore",
		"website":         "https://cryptopro.exchange",
		"contact_email":   "compliance@cryptopro.exchange",
		"compliance_score": 92,
		"risk_level":      "LOW",
		"registration_date": "2024-01-15T00:00:00Z",
		"last_audit":      "2024-11-01T00:00:00Z",
		"next_audit":      "2025-02-01T00:00:00Z",
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    exchange,
	})
}

// SuspendExchange suspends an exchange
func (h *Handler) SuspendExchange(c *gin.Context) {
	exchangeID := c.Param("id")

	var req struct {
		Reason   string `json:"reason" binding:"required"`
		UserID   string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	log.Printf("Exchange %s suspended by user %s. Reason: %s", exchangeID, req.UserID, req.Reason)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"id":          exchangeID,
			"status":      "SUSPENDED",
			"suspended_by": req.UserID,
			"suspended_at": time.Now().UTC().Format(time.RFC3339),
			"reason":      req.Reason,
		},
	})
}

// Wallets

// GetWallets returns a paginated list of wallets
func (h *Handler) GetWallets(c *gin.Context) {
	var params PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		params.Page = 1
		params.PageSize = 20
	}

	wallets := []gin.H{
		{
			"id":              "w_001",
			"address":         "1A2B3C4D5E6F7890ABCDEF1234567890",
			"label":           "可疑地址 #4521",
			"type":            "MIXER",
			"status":          "BLACKLISTED",
			"risk_score":      95,
			"first_seen":      "2023-06-15T00:00:00Z",
			"last_activity":   time.Now().UTC().Format(time.RFC3339),
		},
		{
			"id":              "w_002",
			"address":         "0x1234567890ABCDEF1234567890ABCDEF12345678",
			"label":           "交易所热钱包 B",
			"type":            "EXCHANGE",
			"status":          "ACTIVE",
			"risk_score":      15,
			"first_seen":      "2024-01-20T00:00:00Z",
			"last_activity":   time.Now().UTC().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"items": wallets,
		},
		Meta: &MetaInfo{
			Page:       params.Page,
			PageSize:   params.PageSize,
			Total:      int64(len(wallets)),
			TotalPages: 1,
		},
	})
}

// FreezeWallet freezes a wallet
func (h *Handler) FreezeWallet(c *gin.Context) {
	walletID := c.Param("id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	log.Printf("Wallet %s frozen by user %s. Reason: %s", walletID, req.UserID, req.Reason)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"id":        walletID,
			"status":    "FROZEN",
			"frozen_by": req.UserID,
			"frozen_at": time.Now().UTC().Format(time.RFC3339),
			"reason":    req.Reason,
		},
	})
}

// Miners

// GetMiners returns a paginated list of miners
func (h *Handler) GetMiners(c *gin.Context) {
	var params PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		params.Page = 1
		params.PageSize = 20
	}

	miners := []gin.H{
		{
			"id":                "m_001",
			"name":              "Northern Mining Pool",
			"license_number":    "CSIC-MIN-2024-001",
			"status":            "ACTIVE",
			"jurisdiction":      "Canada",
			"hash_rate":         450,
			"energy_consumption": 520,
			"energy_source":     "Hydroelectric",
			"compliance_status": "COMPLIANT",
			"registration_date": "2023-08-10T00:00:00Z",
			"last_inspection":   "2024-10-01T00:00:00Z",
		},
		{
			"id":                "m_002",
			"name":              "GreenHash Energy",
			"license_number":    "CSIC-MIN-2024-002",
			"status":            "ACTIVE",
			"jurisdiction":      "Iceland",
			"hash_rate":         320,
			"energy_consumption": 380,
			"energy_source":     "Geothermal",
			"compliance_status": "COMPLIANT",
			"registration_date": "2023-12-05T00:00:00Z",
			"last_inspection":   "2024-09-15T00:00:00Z",
		},
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"items": miners,
		},
		Meta: &MetaInfo{
			Page:       params.Page,
			PageSize:   params.PageSize,
			Total:      int64(len(miners)),
			TotalPages: 1,
		},
	})
}

// Compliance

// GetComplianceReports returns compliance reports
func (h *Handler) GetComplianceReports(c *gin.Context) {
	var params PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		params.Page = 1
		params.PageSize = 20
	}

	reports := []gin.H{
		{
			"id":           "rpt_001",
			"entity_type":  "EXCHANGE",
			"entity_id":    "ex_001",
			"entity_name":  "CryptoExchange Pro",
			"period":       "Q4 2024",
			"status":       "COMPLIANT",
			"score":        92,
			"generated_at": "2024-11-01T00:00:00Z",
		},
		{
			"id":           "rpt_002",
			"entity_type":  "MINER",
			"entity_id":    "m_001",
			"entity_name":  "Northern Mining Pool",
			"period":       "Q4 2024",
			"status":       "COMPLIANT",
			"score":        95,
			"generated_at": "2024-10-15T00:00:00Z",
		},
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"items": reports,
		},
		Meta: &MetaInfo{
			Page:       params.Page,
			PageSize:   params.PageSize,
			Total:      int64(len(reports)),
			TotalPages: 1,
		},
	})
}

// GenerateComplianceReport generates a new compliance report
func (h *Handler) GenerateComplianceReport(c *gin.Context) {
	var req struct {
		EntityType string `json:"entity_type" binding:"required"`
		EntityID   string `json:"entity_id" binding:"required"`
		Period     string `json:"period" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	log.Printf("Generating compliance report for %s %s for period %s", req.EntityType, req.EntityID, req.Period)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"id":           fmt.Sprintf("rpt_%d", time.Now().Unix()),
			"entity_type":  req.EntityType,
			"entity_id":    req.EntityID,
			"period":       req.Period,
			"status":       "GENERATING",
			"generated_at": time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// Audit Logs

// GetAuditLogs returns audit logs
func (h *Handler) GetAuditLogs(c *gin.Context) {
	var params PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		params.Page = 1
		params.PageSize = 20
	}

	logs := []gin.H{
		{
			"id":           "log_001",
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
			"user_id":      "u_001",
			"username":     "admin",
			"action":       "LOGIN",
			"resource_type": "session",
			"details":      "用户登录成功",
			"ip_address":   "192.168.1.100",
			"status":       "SUCCESS",
		},
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"items": logs,
		},
		Meta: &MetaInfo{
			Page:       params.Page,
			PageSize:   params.PageSize,
			Total:      int64(len(logs)),
			TotalPages: 1,
		},
	})
}

// Blockchain

// GetBlockchainStatus returns blockchain node status
func (h *Handler) GetBlockchainStatus(c *gin.Context) {
	status := gin.H{
		"bitcoin": gin.H{
			"node_type":     "bitcoin",
			"connected":     true,
			"block_height":  820000,
			"headers_height": 820000,
			"peers":         125,
			"latency_ms":    45,
			"sync_progress": 100.0,
		},
		"ethereum": gin.H{
			"node_type":     "ethereum",
			"connected":     true,
			"block_height":  18500000,
			"headers_height": 18500000,
			"peers":         45,
			"latency_ms":    120,
			"sync_progress": 100.0,
		},
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    status,
	})
}

// Users

// GetCurrentUser returns the current authenticated user
func (h *Handler) GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"id":           userID,
			"username":     "admin",
			"role":         "ADMIN",
			"permissions":  []string{"*"},
			"last_login":   time.Now().UTC().Format(time.RFC3339),
			"mfa_enabled":  true,
		},
	})
}

// GetUsers returns all users (admin only)
func (h *Handler) GetUsers(c *gin.Context) {
	users := []gin.H{
		{
			"id":          "u_001",
			"username":    "admin",
			"email":       "admin@csic.gov",
			"role":        "ADMIN",
			"status":      "ACTIVE",
			"last_login":  time.Now().UTC().Format(time.RFC3339),
			"mfa_enabled": true,
		},
		{
			"id":          "u_002",
			"username":    "regulator_1",
			"email":       "regulator1@csic.gov",
			"role":        "REGULATOR",
			"status":      "ACTIVE",
			"last_login":  time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339),
			"mfa_enabled": true,
		},
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"items": users,
		},
	})
}

// ErrorHandler handles panics and returns proper error responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)

				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Error: &ErrorInfo{
						Code:    "INTERNAL_ERROR",
						Message: "An internal error occurred",
					},
				})
			}
		}()

		c.Next()
	}
}

// RequestLogger logs all incoming requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf("%s %s %d %v",
			c.Request.Method,
			c.Request.URL.Path,
			status,
			latency,
		)
	}
}

// RateLimiter implements rate limiting middleware
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    requestsPerMinute,
		window:   time.Minute,
	}
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now()
		windowStart := now.Add(-rl.window)

		// Filter out old requests
		requests := rl.requests[ip]
		rl.requests[ip] = filterRequests(requests, windowStart)

		if len(rl.requests[ip]) >= rl.limit {
			c.JSON(http.StatusTooManyRequests, Response{
				Success: false,
				Error: &ErrorInfo{
					Code:    "RATE_LIMIT_EXCEEDED",
					Message: "Too many requests",
				},
			})
			c.Abort()
			return
		}

		rl.requests[ip] = append(rl.requests[ip], now)
		c.Next()
	}
}

func filterRequests(requests []time.Time, since time.Time) []time.Time {
	filtered := make([]time.Time, 0, len(requests))
	for _, req := range requests {
		if req.After(since) {
			filtered = append(filtered, req)
		}
	}
	return filtered
}

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeaders adds security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'")
		
		c.Next()
	}
}
