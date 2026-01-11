package handler

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/csic-platform/services/api-gateway/internal/config"
	"github.com/csic-platform/services/api-gateway/internal/core/domain"
	"github.com/csic-platform/services/api-gateway/internal/core/ports"
	"github.com/gin-gonic/gin"
)

// HTTPHandler contains all HTTP handlers for the API
type HTTPHandler struct {
	service ports.GatewayService
	cfg     *config.Config
}

// NewHTTPHandler creates a new HTTP handler instance
func NewHTTPHandler(service ports.GatewayService, cfg *config.Config) *HTTPHandler {
	return &HTTPHandler{
		service: service,
		cfg:     cfg,
	}
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

// HealthCheck returns the health status of the service
func (h *HTTPHandler) HealthCheck(c *gin.Context) {
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
func (h *HTTPHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get dashboard stats",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    stats,
	})
}

// GetAlerts returns a paginated list of alerts
func (h *HTTPHandler) GetAlerts(c *gin.Context) {
	page, pageSize := h.getPaginationParams(c)

	alerts, err := h.service.GetAlerts(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get alerts",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    alerts.Items,
		Meta: &MetaInfo{
			Page:       alerts.Page,
			PageSize:   alerts.PageSize,
			Total:      alerts.Total,
			TotalPages: alerts.TotalPages,
		},
	})
}

// GetAlertByID returns a specific alert by ID
func (h *HTTPHandler) GetAlertByID(c *gin.Context) {
	alertID := c.Param("id")

	alert, err := h.service.GetAlertByID(c.Request.Context(), alertID)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Alert not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    alert,
	})
}

// AcknowledgeAlert acknowledges an alert
func (h *HTTPHandler) AcknowledgeAlert(c *gin.Context) {
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

	if err := h.service.AcknowledgeAlert(c.Request.Context(), alertID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to acknowledge alert",
			},
		})
		return
	}

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

// GetExchanges returns a paginated list of exchanges
func (h *HTTPHandler) GetExchanges(c *gin.Context) {
	page, pageSize := h.getPaginationParams(c)

	exchanges, err := h.service.GetExchanges(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get exchanges",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    exchanges.Items,
		Meta: &MetaInfo{
			Page:       exchanges.Page,
			PageSize:   exchanges.PageSize,
			Total:      exchanges.Total,
			TotalPages: exchanges.TotalPages,
		},
	})
}

// GetExchangeByID returns a specific exchange by ID
func (h *HTTPHandler) GetExchangeByID(c *gin.Context) {
	exchangeID := c.Param("id")

	exchange, err := h.service.GetExchangeByID(c.Request.Context(), exchangeID)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Exchange not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    exchange,
	})
}

// SuspendExchange suspends an exchange
func (h *HTTPHandler) SuspendExchange(c *gin.Context) {
	exchangeID := c.Param("id")

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

	if err := h.service.SuspendExchange(c.Request.Context(), exchangeID, req.Reason, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to suspend exchange",
			},
		})
		return
	}

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

// GetWallets returns a paginated list of wallets
func (h *HTTPHandler) GetWallets(c *gin.Context) {
	page, pageSize := h.getPaginationParams(c)

	wallets, err := h.service.GetWallets(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get wallets",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    wallets.Items,
		Meta: &MetaInfo{
			Page:       wallets.Page,
			PageSize:   wallets.PageSize,
			Total:      wallets.Total,
			TotalPages: wallets.TotalPages,
		},
	})
}

// GetWalletByID returns a specific wallet by ID
func (h *HTTPHandler) GetWalletByID(c *gin.Context) {
	walletID := c.Param("id")

	wallet, err := h.service.GetWalletByID(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Wallet not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    wallet,
	})
}

// FreezeWallet freezes a wallet
func (h *HTTPHandler) FreezeWallet(c *gin.Context) {
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

	if err := h.service.FreezeWallet(c.Request.Context(), walletID, req.Reason, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to freeze wallet",
			},
		})
		return
	}

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

// GetMiners returns a paginated list of miners
func (h *HTTPHandler) GetMiners(c *gin.Context) {
	page, pageSize := h.getPaginationParams(c)

	miners, err := h.service.GetMiners(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get miners",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    miners.Items,
		Meta: &MetaInfo{
			Page:       miners.Page,
			PageSize:   miners.PageSize,
			Total:      miners.Total,
			TotalPages: miners.TotalPages,
		},
	})
}

// GetMinerByID returns a specific miner by ID
func (h *HTTPHandler) GetMinerByID(c *gin.Context) {
	minerID := c.Param("id")

	miner, err := h.service.GetMinerByID(c.Request.Context(), minerID)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Miner not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    miner,
	})
}

// GetComplianceReports returns compliance reports
func (h *HTTPHandler) GetComplianceReports(c *gin.Context) {
	page, pageSize := h.getPaginationParams(c)

	reports, err := h.service.GetComplianceReports(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get compliance reports",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    reports.Items,
		Meta: &MetaInfo{
			Page:       reports.Page,
			PageSize:   reports.PageSize,
			Total:      reports.Total,
			TotalPages: reports.TotalPages,
		},
	})
}

// GenerateComplianceReport generates a new compliance report
func (h *HTTPHandler) GenerateComplianceReport(c *gin.Context) {
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

	report, err := h.service.GenerateComplianceReport(c.Request.Context(), req.EntityType, req.EntityID, req.Period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to generate compliance report",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    report,
	})
}

// GetAuditLogs returns audit logs
func (h *HTTPHandler) GetAuditLogs(c *gin.Context) {
	page, pageSize := h.getPaginationParams(c)

	logs, err := h.service.GetAuditLogs(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get audit logs",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    logs.Items,
		Meta: &MetaInfo{
			Page:       logs.Page,
			PageSize:   logs.PageSize,
			Total:      logs.Total,
			TotalPages: logs.TotalPages,
		},
	})
}

// GetBlockchainStatus returns blockchain node status
func (h *HTTPHandler) GetBlockchainStatus(c *gin.Context) {
	status, err := h.service.GetBlockchainStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get blockchain status",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    status,
	})
}

// GetCurrentUser returns the current authenticated user
func (h *HTTPHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.service.GetCurrentUser(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "User not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    user,
	})
}

// GetUsers returns all users (admin only)
func (h *HTTPHandler) GetUsers(c *gin.Context) {
	page, pageSize := h.getPaginationParams(c)

	users, err := h.service.GetUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get users",
			},
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    users.Items,
		Meta: &MetaInfo{
			Page:       users.Page,
			PageSize:   users.PageSize,
			Total:      users.Total,
			TotalPages: users.TotalPages,
		},
	})
}

// getPaginationParams extracts pagination parameters from the request
func (h *HTTPHandler) getPaginationParams(c *gin.Context) (int, int) {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	return page, pageSize
}

// ErrorHandler handles panics and returns proper error responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
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

		c.Set("latency", latency)
		c.Set("status", status)
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
