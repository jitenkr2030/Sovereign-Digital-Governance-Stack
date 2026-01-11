package http

import (
	"net/http"

	"github.com/csic-platform/services/services/mining/internal/core/domain"
	"github.com/csic-platform/services/services/mining/internal/core/ports"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handlers implements the HTTP handlers for mining operations
type Handlers struct {
	service ports.MiningService
	log     *zap.Logger
}

// NewHandlers creates a new Handlers instance
func NewHandlers(service ports.MiningService, log *zap.Logger) *Handlers {
	return &Handlers{
		service: service,
		log:     log,
	}
}

// RegisterOperation handles POST /api/v1/operations
func (h *Handlers) RegisterOperation(c *gin.Context) {
	var req ports.RegisterOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid registration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	op, err := h.service.RegisterOperation(c.Request.Context(), req)
	if err != nil {
		h.log.Error("Failed to register operation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to register operation",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Operation registered successfully",
		"operation": op,
	})
}

// GetOperation handles GET /api/v1/operations/:id
func (h *Handlers) GetOperation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid operation ID",
		})
		return
	}

	op, err := h.service.GetOperation(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get operation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get operation",
		})
		return
	}

	if op == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Operation not found",
		})
		return
	}

	c.JSON(http.StatusOK, op)
}

// GetOperationByWallet handles GET /api/v1/operations/wallet/:wallet
func (h *Handlers) GetOperationByWallet(c *gin.Context) {
	wallet := c.Param("wallet")

	op, err := h.service.GetOperationByWallet(c.Request.Context(), wallet)
	if err != nil {
		h.log.Error("Failed to get operation by wallet", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get operation",
		})
		return
	}

	if op == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Operation not found",
		})
		return
	}

	c.JSON(http.StatusOK, op)
}

// ListOperations handles GET /api/v1/operations
func (h *Handlers) ListOperations(c *gin.Context) {
	var status *domain.OperationStatus
	if s := c.Query("status"); s != "" {
		st := domain.OperationStatus(s)
		status = &st
	}

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := parseInt(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := parseInt(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	operations, err := h.service.ListOperations(c.Request.Context(), status, page, pageSize)
	if err != nil {
		h.log.Error("Failed to list operations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list operations",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"operations": operations,
		"page":       page,
		"page_size":  pageSize,
	})
}

// UpdateOperationStatus handles PUT /api/v1/operations/:id/status
func (h *Handlers) UpdateOperationStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid operation ID",
		})
		return
	}

	var req struct {
		Status domain.OperationStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	if err := h.service.UpdateOperationStatus(c.Request.Context(), id, req.Status); err != nil {
		h.log.Error("Failed to update operation status", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to update status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status updated successfully",
	})
}

// ReportHashrate handles POST /api/v1/operations/:id/telemetry
func (h *Handlers) ReportHashrate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid operation ID",
		})
		return
	}

	var req ports.TelemetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid telemetry request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Override operation ID from path
	req.OperationID = id

	if err := h.service.ReportHashrate(c.Request.Context(), req); err != nil {
		h.log.Error("Failed to report hashrate", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process telemetry",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hashrate recorded successfully",
	})
}

// GetHashrateHistory handles GET /api/v1/operations/:id/metrics
func (h *Handlers) GetHashrateHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid operation ID",
		})
		return
	}

	limit := 1000
	if l := c.Query("limit"); l != "" {
		if parsed, err := parseInt(l); err == nil && parsed > 0 && parsed <= 10000 {
			limit = parsed
		}
	}

	records, err := h.service.GetHashrateHistory(c.Request.Context(), id, limit)
	if err != nil {
		h.log.Error("Failed to get hashrate history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get metrics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"operation_id": id,
		"records":      records,
		"count":        len(records),
	})
}

// AssignQuota handles POST /api/v1/quotas
func (h *Handlers) AssignQuota(c *gin.Context) {
	var req ports.QuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid quota request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	quota, err := h.service.AssignQuota(c.Request.Context(), req)
	if err != nil {
		h.log.Error("Failed to assign quota", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to assign quota",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Quota assigned successfully",
		"quota":   quota,
	})
}

// GetQuota handles GET /api/v1/operations/:id/quota
func (h *Handlers) GetQuota(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid operation ID",
		})
		return
	}

	quota, err := h.service.GetCurrentQuota(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get quota", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get quota",
		})
		return
	}

	if quota == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No active quota found",
		})
		return
	}

	c.JSON(http.StatusOK, quota)
}

// RevokeQuota handles DELETE /api/v1/quotas/:id
func (h *Handlers) RevokeQuota(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid quota ID",
		})
		return
	}

	if err := h.service.RevokeQuota(c.Request.Context(), id); err != nil {
		h.log.Error("Failed to revoke quota", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to revoke quota",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Quota revoked successfully",
	})
}

// IssueShutdownCommand handles POST /api/v1/operations/:id/shutdown
func (h *Handlers) IssueShutdownCommand(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid operation ID",
		})
		return
	}

	var req ports.ShutdownRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid shutdown request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	req.OperationID = id

	cmd, err := h.service.IssueShutdownCommand(c.Request.Context(), req)
	if err != nil {
		h.log.Error("Failed to issue shutdown command", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to issue command",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Shutdown command issued",
		"command": cmd,
	})
}

// AcknowledgeCommand handles POST /api/v1/commands/:id/ack
func (h *Handlers) AcknowledgeCommand(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid command ID",
		})
		return
	}

	if err := h.service.AcknowledgeCommand(c.Request.Context(), id); err != nil {
		h.log.Error("Failed to acknowledge command", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to acknowledge command",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Command acknowledged",
	})
}

// ConfirmShutdown handles POST /api/v1/commands/:id/confirm
func (h *Handlers) ConfirmShutdown(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid command ID",
		})
		return
	}

	var req struct {
		Success bool   `json:"success"`
		Result  string `json:"result,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	if err := h.service.ConfirmShutdown(c.Request.Context(), id, req.Result); err != nil {
		h.log.Error("Failed to confirm shutdown", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to confirm shutdown",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Shutdown confirmed",
	})
}

// GetPendingCommands handles GET /api/v1/operations/:id/commands
func (h *Handlers) GetPendingCommands(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid operation ID",
		})
		return
	}

	commands, err := h.service.GetPendingCommands(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get pending commands", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get commands",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"operation_id": id,
		"commands":     commands,
		"count":        len(commands),
	})
}

// GetRegistryStats handles GET /api/v1/stats
func (h *Handlers) GetRegistryStats(c *gin.Context) {
	stats, err := h.service.GetRegistryStats(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get registry stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// HealthCheck handles GET /health
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "mining-control-platform",
	})
}

// Helper function to parse integers
func parseInt(s string) (int, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	return int(i), err
}
