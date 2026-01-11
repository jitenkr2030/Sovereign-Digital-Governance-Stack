package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/csic-platform/services/services/compliance/internal/core/domain"
	"github.com/csic-platform/services/services/compliance/internal/core/ports"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handlers implements the HTTP handlers for compliance operations
type Handlers struct {
	licenseService     ports.LicenseService
	complianceService  ports.ComplianceService
	obligationService  ports.ObligationService
	auditService       ports.AuditService
	log                *zap.Logger
}

// NewHandlers creates a new Handlers instance
func NewHandlers(
	licenseService ports.LicenseService,
	complianceService ports.ComplianceService,
	obligationService ports.ObligationService,
	auditService ports.AuditService,
	log *zap.Logger,
) *Handlers {
	return &Handlers{
		licenseService:    licenseService,
		complianceService: complianceService,
		obligationService: obligationService,
		auditService:      auditService,
		log:               log,
	}
}

// ===== License Application Handlers =====

// SubmitApplication handles POST /api/v1/licenses/applications
func (h *Handlers) SubmitApplication(c *gin.Context) {
	var req ports.SubmitApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid application request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	app, err := h.licenseService.SubmitApplication(c.Request.Context(), req)
	if err != nil {
		h.log.Error("Failed to submit application", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to submit application", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Application submitted", "application": app})
}

// ReviewApplication handles PATCH /api/v1/licenses/applications/:id/review
func (h *Handlers) ReviewApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	var req ports.ApplicationReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	license, err := h.licenseService.ReviewApplication(c.Request.Context(), id, req)
	if err != nil {
		h.log.Error("Failed to review application", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to review application", "details": err.Error()})
		return
	}

	if license != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Application approved, license issued", "license": license})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Application rejected"})
	}
}

// GetApplication handles GET /api/v1/licenses/applications/:id
func (h *Handlers) GetApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	app, err := h.licenseService.GetApplication(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get application", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get application"})
		return
	}

	if app == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	c.JSON(http.StatusOK, app)
}

// ListApplications handles GET /api/v1/licenses/applications
func (h *Handlers) ListApplications(c *gin.Context) {
	var filter ports.ApplicationFilter
	filter.Page = 1
	filter.PageSize = 20

	if s := c.Query("status"); s != "" {
		status := domain.ApplicationStatus(s)
		filter.Status = &status
	}
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			filter.Page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			filter.PageSize = parsed
		}
	}

	apps, err := h.licenseService.ListApplications(c.Request.Context(), filter)
	if err != nil {
		h.log.Error("Failed to list applications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list applications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"applications": apps, "page": filter.Page, "page_size": filter.PageSize})
}

// ===== License Handlers =====

// IssueLicense handles POST /api/v1/licenses
func (h *Handlers) IssueLicense(c *gin.Context) {
	var req ports.IssueLicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid license request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	license, err := h.licenseService.IssueLicense(c.Request.Context(), req)
	if err != nil {
		h.log.Error("Failed to issue license", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to issue license", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "License issued", "license": license})
}

// GetLicense handles GET /api/v1/licenses/:id
func (h *Handlers) GetLicense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid license ID"})
		return
	}

	license, err := h.licenseService.GetLicense(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get license", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get license"})
		return
	}

	if license == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "License not found"})
		return
	}

	c.JSON(http.StatusOK, license)
}

// GetEntityLicenses handles GET /api/v1/entities/:id/licenses
func (h *Handlers) GetEntityLicenses(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	licenses, err := h.licenseService.GetEntityLicenses(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get entity licenses", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get licenses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entity_id": id, "licenses": licenses})
}

// SuspendLicense handles POST /api/v1/licenses/:id/suspend
func (h *Handlers) SuspendLicense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid license ID"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reason required"})
		return
	}

	if err := h.licenseService.SuspendLicense(c.Request.Context(), id, req.Reason); err != nil {
		h.log.Error("Failed to suspend license", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to suspend license", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "License suspended"})
}

// RevokeLicense handles POST /api/v1/licenses/:id/revoke
func (h *Handlers) RevokeLicense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid license ID"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reason required"})
		return
	}

	if err := h.licenseService.RevokeLicense(c.Request.Context(), id, req.Reason); err != nil {
		h.log.Error("Failed to revoke license", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to revoke license", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "License revoked"})
}

// GetExpiringLicenses handles GET /api/v1/licenses/expiring
func (h *Handlers) GetExpiringLicenses(c *gin.Context) {
	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	licenses, err := h.licenseService.GetExpiringLicenses(c.Request.Context(), days)
	if err != nil {
		h.log.Error("Failed to get expiring licenses", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get expiring licenses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"days": days, "licenses": licenses})
}

// ===== Entity Handlers =====

// RegisterEntity handles POST /api/v1/entities
func (h *Handlers) RegisterEntity(c *gin.Context) {
	var req ports.RegisterEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid entity request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	entity, err := h.licenseService.RegisterEntity(c.Request.Context(), req)
	if err != nil {
		h.log.Error("Failed to register entity", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to register entity", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Entity registered", "entity": entity})
}

// GetEntity handles GET /api/v1/entities/:id
func (h *Handlers) GetEntity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	entity, err := h.licenseService.GetEntity(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get entity", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get entity"})
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// ===== Compliance Scoring Handlers =====

// GetComplianceScore handles GET /api/v1/entities/:id/compliance/score
func (h *Handlers) GetComplianceScore(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	score, err := h.complianceService.GetCurrentScore(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get compliance score", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get score"})
		return
	}

	if score == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No score available"})
		return
	}

	c.JSON(http.StatusOK, score)
}

// RecalculateScore handles POST /api/v1/entities/:id/compliance/score/recalculate
func (h *Handlers) RecalculateScore(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	score, err := h.complianceService.CalculateScore(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to recalculate score", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalculate score"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Score recalculated", "score": score})
}

// GetScoreHistory handles GET /api/v1/entities/:id/compliance/score/history
func (h *Handlers) GetScoreHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	history, err := h.complianceService.GetScoreHistory(c.Request.Context(), id, limit)
	if err != nil {
		h.log.Error("Failed to get score history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entity_id": id, "history": history})
}

// GetComplianceStats handles GET /api/v1/compliance/stats
func (h *Handlers) GetComplianceStats(c *gin.Context) {
	stats, err := h.complianceService.GetComplianceStats(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get compliance stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ===== Obligation Handlers =====

// CreateObligation handles POST /api/v1/obligations
func (h *Handlers) CreateObligation(c *gin.Context) {
	var req ports.CreateObligationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("Invalid obligation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	obligation, err := h.obligationService.CreateObligation(c.Request.Context(), req)
	if err != nil {
		h.log.Error("Failed to create obligation", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create obligation", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Obligation created", "obligation": obligation})
}

// GetObligation handles GET /api/v1/obligations/:id
func (h *Handlers) GetObligation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid obligation ID"})
		return
	}

	obligation, err := h.obligationService.GetObligation(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get obligation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get obligation"})
		return
	}

	if obligation == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Obligation not found"})
		return
	}

	c.JSON(http.StatusOK, obligation)
}

// GetEntityObligations handles GET /api/v1/entities/:id/obligations
func (h *Handlers) GetEntityObligations(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	obligations, err := h.obligationService.GetEntityObligations(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get entity obligations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get obligations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entity_id": id, "obligations": obligations})
}

// GetOverdueObligations handles GET /api/v1/obligations/overdue
func (h *Handlers) GetOverdueObligations(c *gin.Context) {
	obligations, err := h.obligationService.GetOverdueObligations(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get overdue obligations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get overdue obligations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"obligations": obligations})
}

// FulfillObligation handles POST /api/v1/obligations/:id/fulfill
func (h *Handlers) FulfillObligation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid obligation ID"})
		return
	}

	var req struct {
		Evidence string `json:"evidence" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Evidence required"})
		return
	}

	if err := h.obligationService.FulfillObligation(c.Request.Context(), id, req.Evidence); err != nil {
		h.log.Error("Failed to fulfill obligation", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to fulfill obligation", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Obligation fulfilled"})
}

// CheckOverdueObligations handles POST /api/v1/obligations/check-overdue
func (h *Handlers) CheckOverdueObligations(c *gin.Context) {
	if err := h.obligationService.CheckAndUpdateOverdueObligations(c.Request.Context()); err != nil {
		h.log.Error("Failed to check overdue obligations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check overdue obligations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Overdue check completed"})
}

// ===== Audit Handlers =====

// GetAuditLogs handles GET /api/v1/audit-logs
func (h *Handlers) GetAuditLogs(c *gin.Context) {
	var filter ports.AuditLogFilter
	filter.Page = 1
	filter.PageSize = 100

	if entityID := c.Query("entity_id"); entityID != "" {
		if parsed, err := uuid.Parse(entityID); err == nil {
			filter.EntityID = &parsed
		}
	}
	if actorID := c.Query("actor_id"); actorID != "" {
		if parsed, err := uuid.Parse(actorID); err == nil {
			filter.ActorID = &parsed
		}
	}
	if resourceType := c.Query("resource_type"); resourceType != "" {
		filter.ResourceType = resourceType
	}
	if actionType := c.Query("action_type"); actionType != "" {
		filter.ActionType = actionType
	}
	if from := c.Query("from"); from != "" {
		if parsed, err := time.Parse(time.RFC3339, from); err == nil {
			filter.From = &parsed
		}
	}
	if to := c.Query("to"); to != "" {
		if parsed, err := time.Parse(time.RFC3339, to); err == nil {
			filter.To = &parsed
		}
	}
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			filter.Page = parsed
		}
	}

	logs, err := h.auditService.GetAuditLogs(c.Request.Context(), filter)
	if err != nil {
		h.log.Error("Failed to get audit logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit logs"})
		return
	}

	count, _ := h.auditService.CountAuditLogs(c.Request.Context(), filter)

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"page":      filter.Page,
		"page_size": filter.PageSize,
		"total":     count,
	})
}

// GetEntityAuditTrail handles GET /api/v1/entities/:id/audit-trail
func (h *Handlers) GetEntityAuditTrail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

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

	trail, err := h.auditService.GetEntityAuditTrail(c.Request.Context(), id, limit, offset)
	if err != nil {
		h.log.Error("Failed to get audit trail", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit trail"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entity_id": id, "audit_trail": trail})
}

// ===== Health Check =====

func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "compliance-module",
	})
}
