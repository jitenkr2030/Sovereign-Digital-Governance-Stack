// Audit Log HTTP Handlers
// REST API handlers for the Audit Log Service

package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/csic-platform/services/audit-log"
	"github.com/gin-gonic/gin"
)

// AuditLogHandler handles HTTP requests for audit log operations
type AuditLogHandler struct {
	service *AuditLogService
}

// NewAuditLogHandler creates a new audit log HTTP handler
func NewAuditLogHandler(service *AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{
		service: service,
	}
}

// LoggingMiddleware returns a gin middleware for request logging
func (h *AuditLogHandler) LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// Log request (simplified - in production use proper logging)
		if status >= 400 {
			println("ERROR:", path, status, latency.String())
		}
	}
}

// HealthCheck returns the service health status
func (h *AuditLogHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "audit-log-service",
	})
}

// ReadinessCheck returns the service readiness status
func (h *AuditLogHandler) ReadinessCheck(c *gin.Context) {
	// Check if service is running
	if !h.service.running {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not ready",
			"reason": "service not running",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// WriteEntry handles writing a single audit log entry
func (h *AuditLogHandler) WriteEntry(c *gin.Context) {
	var entry audit.AuditLogEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.service.WriteLog(c.Request.Context(), &entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to write entry",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "entry created",
		"entry_id":  entry.EntryID,
		"sequence":  entry.SequenceNum,
		"timestamp": entry.Timestamp,
	})
}

// WriteBatch handles writing multiple audit log entries
func (h *AuditLogHandler) WriteBatch(c *gin.Context) {
	var entries []*audit.AuditLogEntry
	if err := c.ShouldBindJSON(&entries); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	if len(entries) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "empty batch",
		})
		return
	}

	if len(entries) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "batch size exceeds maximum (1000)",
		})
		return
	}

	if err := h.service.WriteBatch(c.Request.Context(), entries); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to write batch",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "batch created",
		"entries":     len(entries),
		"first_seq":   entries[0].SequenceNum,
		"last_seq":    entries[len(entries)-1].SequenceNum,
	})
}

// QueryEntries handles querying audit log entries
func (h *AuditLogHandler) QueryEntries(c *gin.Context) {
	query := &audit.AuditQuery{}

	// Parse query parameters
	if startTime := c.Query("start_time"); startTime != "" {
		t, err := time.Parse(time.RFC3339, startTime)
		if err == nil {
			query.StartTime = t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		t, err := time.Parse(time.RFC3339, endTime)
		if err == nil {
			query.EndTime = t
		}
	}

	query.ActorID = c.Query("actor_id")
	query.Service = c.Query("service")
	query.Operation = c.Query("operation")
	query.ActionType = c.Query("action_type")
	query.Resource = c.Query("resource")
	query.Result = c.Query("result")
	query.RiskLevel = c.Query("risk_level")

	if seqFrom := c.Query("sequence_from"); seqFrom != "" {
		if v, err := strconv.ParseUint(seqFrom, 10, 64); err == nil {
			query.SequenceFrom = v
		}
	}

	if seqTo := c.Query("sequence_to"); seqTo != "" {
		if v, err := strconv.ParseUint(seqTo, 10, 64); err == nil {
			query.SequenceTo = v
		}
	}

	query.Limit = 100 // Default limit
	if limit := c.Query("limit"); limit != "" {
		if v, err := strconv.Atoi(limit); err == nil && v > 0 && v <= 1000 {
			query.Limit = v
		}
	}

	query.Offset = 0
	if offset := c.Query("offset"); offset != "" {
		if v, err := strconv.Atoi(offset); err == nil && v >= 0 {
			query.Offset = v
		}
	}

	entries, err := h.service.QueryEntries(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to query entries",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"count":   len(entries),
		"limit":   query.Limit,
		"offset":  query.Offset,
	})
}

// GetEntry handles retrieving a specific audit log entry
func (h *AuditLogHandler) GetEntry(c *gin.Context) {
	entryID := c.Param("id")
	if entryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "entry ID is required",
		})
		return
	}

	entry, err := h.service.GetEntry(c.Request.Context(), entryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "entry not found",
		})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// VerifyChain handles verification of the audit log chain
func (h *AuditLogHandler) VerifyChain(c *gin.Context) {
	result, err := h.service.Verify(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "verification failed",
			"details": err.Error(),
		})
		return
	}

	status := http.StatusOK
	if !result.Valid {
		status = http.StatusUnprocessableEntity
	}

	c.JSON(status, result)
}

// GetVerificationReport handles generating a detailed verification report
func (h *AuditLogHandler) GetVerificationReport(c *gin.Context) {
	report, err := h.service.verifier.GenerateVerificationReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to generate report",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// ListChains handles listing all audit log chains
func (h *AuditLogHandler) ListChains(c *gin.Context) {
	summaries, err := h.service.verifier.GetChainSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to list chains",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chains": summaries,
		"count":  len(summaries),
	})
}

// GetChain handles retrieving a specific chain record
func (h *AuditLogHandler) GetChain(c *gin.Context) {
	chainID := c.Param("id")
	if chainID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "chain ID is required",
		})
		return
	}

	chain, err := h.service.GetChain(chainID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "chain not found",
		})
		return
	}

	c.JSON(http.StatusOK, chain)
}

// ExportChain handles exporting a sealed chain for legal discovery
func (h *AuditLogHandler) ExportChain(c *gin.Context) {
	chainID := c.Param("id")
	if chainID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "chain ID is required",
		})
		return
	}

	data, err := h.service.ExportChain(c.Request.Context(), chainID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "chain not found",
		})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=audit-chain-"+chainID+".json")
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, data)
}

// GetSummary handles retrieving audit log statistics
func (h *AuditLogHandler) GetSummary(c *gin.Context) {
	sequenceNum := h.service.writer.GetSequenceNumber()

	c.JSON(http.StatusOK, gin.H{
		"total_entries":    sequenceNum,
		"current_sequence": sequenceNum,
		"service_running":  h.service.running,
	})
}
