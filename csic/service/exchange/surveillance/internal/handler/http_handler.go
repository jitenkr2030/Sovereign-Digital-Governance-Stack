package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/csic/surveillance/internal/domain"
	"github.com/csic/surveillance/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HTTPHandler handles HTTP requests for the surveillance service
type HTTPHandler struct {
	alertSvc   *service.AlertService
	analysisSvc *service.AnalysisService
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(alertSvc *service.AlertService, analysisSvc *service.AnalysisService) *HTTPHandler {
	return &HTTPHandler{
		alertSvc:    alertSvc,
		analysisSvc: analysisSvc,
	}
}

// ListAlerts returns a list of alerts with optional filtering
func (h *HTTPHandler) ListAlerts(c *gin.Context) {
	// Parse query parameters
	filter := domain.AlertFilter{}

	// Types
	if types := c.Query("types"); types != "" {
		for _, t := range splitString(types, ",") {
			filter.Types = append(filter.Types, domain.AlertType(t))
		}
	}

	// Severities
	if severities := c.Query("severities"); severities != "" {
		for _, s := range splitString(severities, ",") {
			filter.Severities = append(filter.Severities, domain.AlertSeverity(s))
		}
	}

	// Statuses
	if statuses := c.Query("statuses"); statuses != "" {
		for _, s := range splitString(statuses, ",") {
			filter.Statuses = append(filter.Statuses, domain.AlertStatus(s))
		}
	}

	// Exchange IDs
	if exchangeIDs := c.Query("exchange_ids"); exchangeIDs != "" {
		for _, idStr := range splitString(exchangeIDs, ",") {
			if id, err := uuid.Parse(idStr); err == nil {
				filter.ExchangeIDs = append(filter.ExchangeIDs, id)
			}
		}
	}

	// Symbols
	if symbols := c.Query("symbols"); symbols != "" {
		filter.Symbols = splitString(symbols, ",")
	}

	// Time range
	if after := c.Query("detected_after"); after != "" {
		if t, err := time.Parse(time.RFC3339, after); err == nil {
			filter.DetectedAfter = &t
		}
	}
	if before := c.Query("detected_before"); before != "" {
		if t, err := time.Parse(time.RFC3339, before); err == nil {
			filter.DetectedBefore = &t
		}
	}

	// Pagination
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	alerts, err := h.alertSvc.ListAlerts(c.Request.Context(), filter, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve alerts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"meta": gin.H{
			"limit":   limit,
			"offset":  offset,
			"count":   len(alerts),
		},
	})
}

// GetAlert returns a single alert by ID
func (h *HTTPHandler) GetAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid alert ID format",
		})
		return
	}

	alert, err := h.alertSvc.GetAlert(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Alert not found",
		})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// UpdateAlertStatus updates the status of an alert
func (h *HTTPHandler) UpdateAlertStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid alert ID format",
		})
		return
	}

	var req struct {
		Status    string `json:"status"`
		Resolution string `json:"resolution"`
		UserID    string `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	switch domain.AlertStatus(req.Status) {
	case domain.AlertStatusResolved:
		if err := h.alertSvc.ResolveAlert(c.Request.Context(), id, req.Resolution, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to resolve alert",
			})
			return
		}
	case domain.AlertStatusDismissed:
		if err := h.alertSvc.DismissAlert(c.Request.Context(), id, req.Resolution, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to dismiss alert",
			})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status. Use 'RESOLVED' or 'DISMISSED'",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert status updated successfully",
		"alert_id": id,
	})
}

// GetMarketSummary returns a summary of market activity for an exchange/symbol
func (h *HTTPHandler) GetMarketSummary(c *gin.Context) {
	exchangeIDStr := c.Param("exchange_id")
	exchangeID, err := uuid.Parse(exchangeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid exchange ID format",
		})
		return
	}

	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Symbol is required",
		})
		return
	}

	// Parse time range
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	// This would call the repository in a real implementation
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"period": gin.H{
			"start": startTime,
			"end":   endTime,
		},
		"summary": gin.H{
			"open_price":   "0",
			"high_price":   "0",
			"low_price":    "0",
			"close_price":  "0",
			"volume":       "0",
			"trade_count":  0,
		},
	})
}

// AnalyzeWashTrading performs wash trade analysis for a given exchange and symbol
func (h *HTTPHandler) AnalyzeWashTrading(c *gin.Context) {
	var req struct {
		ExchangeID string `json:"exchange_id" binding:"required"`
		Symbol     string `json:"symbol" binding:"required"`
		StartTime  string `json:"start_time"`
		EndTime    string `json:"end_time"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	exchangeID, err := uuid.Parse(req.ExchangeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid exchange ID format",
		})
		return
	}

	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	if req.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			startTime = t
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			endTime = t
		}
	}

	candidates, err := h.analysisSvc.DetectWashTrading(c.Request.Context(), exchangeID, req.Symbol, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Analysis failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis_type": "wash_trading",
		"exchange_id":   exchangeID,
		"symbol":        req.Symbol,
		"time_window": gin.H{
			"start": startTime,
			"end":   endTime,
		},
		"candidates": candidates,
		"count":      len(candidates),
	})
}

// AnalyzeSpoofing performs spoofing analysis for a given exchange and symbol
func (h *HTTPHandler) AnalyzeSpoofing(c *gin.Context) {
	var req struct {
		ExchangeID string `json:"exchange_id" binding:"required"`
		Symbol     string `json:"symbol" binding:"required"`
		StartTime  string `json:"start_time"`
		EndTime    string `json:"end_time"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	exchangeID, err := uuid.Parse(req.ExchangeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid exchange ID format",
		})
		return
	}

	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	if req.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			startTime = t
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			endTime = t
		}
	}

	indicators, err := h.analysisSvc.DetectSpoofing(c.Request.Context(), exchangeID, req.Symbol, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Analysis failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis_type": "spoofing",
		"exchange_id":   exchangeID,
		"symbol":        req.Symbol,
		"time_window": gin.H{
			"start": startTime,
			"end":   endTime,
		},
		"indicators": indicators,
		"count":      len(indicators),
	})
}

// GetThresholds returns the current analysis thresholds
func (h *HTTPHandler) GetThresholds(c *gin.Context) {
	thresholds, err := h.alertSvc.GetThresholds(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve thresholds",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"thresholds": thresholds,
	})
}

// UpdateThresholds updates the analysis thresholds
func (h *HTTPHandler) UpdateThresholds(c *gin.Context) {
	var thresholds map[string]map[string]interface{}

	if err := c.ShouldBindJSON(&thresholds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Convert to the format expected by the service
	formattedThresholds := make(map[domain.AlertType]map[string]any)
	for alertType, params := range thresholds {
		formattedThresholds[domain.AlertType(alertType)] = params
	}

	if err := h.analysisSvc.ConfigureThresholds(c.Request.Context(), formattedThresholds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update thresholds",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thresholds updated successfully",
	})
}

// Helper function to split a string by delimiter
func splitString(s string, delimiter string) []string {
	if s == "" {
		return nil
	}
	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == delimiter[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}
