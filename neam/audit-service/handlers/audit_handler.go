package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"neam-platform/audit-service/services"
)

// AuditHandler handles audit-related HTTP requests
type AuditHandler struct {
	auditService services.AuditService
}

// NewAuditHandler creates a new AuditHandler
func NewAuditHandler(auditService services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// RegisterRoutes registers audit-related routes
func (h *AuditHandler) RegisterRoutes(router *gin.RouterGroup) {
	audit := router.Group("/audit")
	{
		audit.GET("", h.QueryEvents)
		audit.GET("/:id", h.GetEvent)
		audit.GET("/statistics", h.GetStatistics)
		audit.POST("/verify", h.VerifyIntegrity)
		audit.POST("/seal", h.SealEvents)
		audit.GET("/chain/:id", h.GetChain)
		audit.GET("/export", h.ExportAuditLogs)
	}
}

// QueryEvents handles GET /api/v1/audit
// @Summary Query audit events
// @Description Query audit events with filters
// @Tags audit
// @Accept json
// @Produce json
// @Param event_types query string false "Comma-separated event types"
// @Param actor_ids query string false "Comma-separated actor IDs"
// @Param resources query string false "Comma-separated resources"
// @Param outcomes query string false "Comma-separated outcomes"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Param limit query int false "Limit (default 100, max 1000)"
// @Param offset query int false "Offset"
// @Success 200 {object} QueryEventsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/audit [get]
func (h *AuditHandler) QueryEvents(c *gin.Context) {
	var query AuditQueryRequest
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_query",
			Message: err.Error(),
		})
		return
	}

	// Parse limit and offset
	limit := 100
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Parse time range
	var timeRange *TimeRange
	if startTime := c.Query("start_time"); startTime != "" {
		start, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_time",
				Message: "Invalid start_time format",
			})
			return
		}
		endTime := c.Query("end_time")
		var end time.Time
		if endTime != "" {
			end, err = time.Parse(time.RFC3339, endTime)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "invalid_time",
					Message: "Invalid end_time format",
				})
				return
			}
		} else {
			end = time.Now()
		}
		timeRange = &TimeRange{Start: start, End: end}
	}

	// Build query
	auditQuery := services.AuditQuery{
		Limit:         limit,
		Offset:        offset,
		TimeRange:     timeRange,
		IncludeSealed: true,
	}

	// Execute query
	events, err := h.auditService.QueryEvents(c.Request.Context(), auditQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, QueryEventsResponse{
		Events:  events,
		Total:   len(events),
		Limit:   limit,
		Offset:  offset,
	})
}

// GetEvent handles GET /api/v1/audit/:id
// @Summary Get audit event by ID
// @Description Get a single audit event by its ID
// @Tags audit
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} services.AuditEvent
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/audit/{id} [get]
func (h *AuditHandler) GetEvent(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_id",
			Message: "Event ID is required",
		})
		return
	}

	event, err := h.auditService.GetEvent(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Event not found",
		})
		return
	}

	c.JSON(http.StatusOK, event)
}

// GetStatistics handles GET /api/v1/audit/statistics
// @Summary Get audit statistics
// @Description Get statistics about audit events
// @Tags audit
// @Produce json
// @Success 200 {object} services.AuditStatistics
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/audit/statistics [get]
func (h *AuditHandler) GetStatistics(c *gin.Context) {
	stats, err := h.auditService.GetStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// VerifyIntegrity handles POST /api/v1/audit/verify
// @Summary Verify audit integrity
// @Description Verify the integrity of audit records
// @Tags audit
// @Accept json
// @Produce json
// @Success 200 {object} services.VerificationReport
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/audit/verify [post]
func (h *AuditHandler) VerifyIntegrity(c *gin.Context) {
	report, err := h.auditService.VerifyIntegrity(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "verification_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// SealEvents handles POST /api/v1/audit/seal
// @Summary Seal audit events
// @Description Seal a batch of audit events cryptographically
// @Tags audit
// @Accept json
// @Produce json
// @Param event_ids body SealRequest true "Event IDs to seal"
// @Success 200 {object} services.SealRecord
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/audit/seal [post]
func (h *AuditHandler) SealEvents(c *gin.Context) {
	var req SealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if len(req.EventIDs) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_event_ids",
			Message: "Event IDs are required",
		})
		return
	}

	seal, err := h.auditService.SealEvents(c.Request.Context(), req.EventIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "sealing_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, seal)
}

// GetChain handles GET /api/v1/audit/chain/:id
// @Summary Get integrity chain for event
// @Description Get the cryptographic chain for an audit event
// @Tags audit
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} ChainResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/audit/chain/{id} [get]
func (h *AuditHandler) GetChain(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_id",
			Message: "Event ID is required",
		})
		return
	}

	chain, err := h.auditService.GetChain(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ChainResponse{
		EventID:  eventID,
		Seals:    chain,
		Length:   len(chain),
	})
}

// ExportAuditLogs handles GET /api/v1/audit/export
// @Summary Export audit logs
// @Description Export audit logs in various formats
// @Tags audit
// @Produce json
// @Param format query string false "Export format (json, csv, pdf)"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Success 200 {object} ExportResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/audit/export [get]
func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// Parse time range
	var timeRange *TimeRange
	if startTime != "" {
		start, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_time",
				Message: "Invalid start_time format",
			})
			return
		}
		var end time.Time
		if endTime != "" {
			end, err = time.Parse(time.RFC3339, endTime)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "invalid_time",
					Message: "Invalid end_time format",
				})
				return
			}
		} else {
			end = time.Now()
		}
		timeRange = &TimeRange{Start: start, End: end}
	}

	// Build query
	query := services.AuditQuery{
		TimeRange:     timeRange,
		IncludeSealed: true,
		Limit:         10000, // Maximum export limit
	}

	// Get events
	events, err := h.auditService.QueryEvents(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "export_failed",
			Message: err.Error(),
		})
		return
	}

	// Export based on format
	switch format {
	case "json":
		c.JSON(http.StatusOK, ExportResponse{
			Format:     "json",
			Events:     events,
			TotalCount: len(events),
			ExportTime: time.Now(),
		})
	case "csv":
		// TODO: Implement CSV export
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "not_implemented",
			Message: "CSV export not yet implemented",
		})
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_format",
			Message: "Invalid export format",
		})
	}
}

// Request/Response types

// AuditQueryRequest represents a query request
type AuditQueryRequest struct {
	EventTypes []string `form:"event_types"`
	ActorIDs   []string `form:"actor_ids"`
	Resources  []string `form:"resources"`
	Outcomes   []string `form:"outcomes"`
	StartTime  string   `form:"start_time"`
	EndTime    string   `form:"end_time"`
}

// TimeRange represents a time range
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// QueryEventsResponse represents the response for query events
type QueryEventsResponse struct {
	Events []interface{} `json:"events"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// SealRequest represents a seal request
type SealRequest struct {
	EventIDs []string `json:"event_ids"`
}

// ChainResponse represents a chain response
type ChainResponse struct {
	EventID string          `json:"event_id"`
	Seals   []interface{}   `json:"seals"`
	Length  int             `json:"length"`
}

// ExportResponse represents an export response
type ExportResponse struct {
	Format     string      `json:"format"`
	Events     []interface{} `json:"events"`
	TotalCount int         `json:"total_count"`
	ExportTime time.Time   `json:"export_time"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
