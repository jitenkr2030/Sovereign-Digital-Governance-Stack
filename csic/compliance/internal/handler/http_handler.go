// Compliance Management Module - HTTP Handlers
// REST API handlers for compliance operations

package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/csic-platform/compliance/internal/domain"
	"github.com/csic-platform/compliance/internal/port"
	"github.com/csic-platform/compliance/internal/service"
	"github.com/gin-gonic/gin"
)

// ComplianceHandler handles HTTP requests for compliance operations
type ComplianceHandler struct {
	entityService       *service.EntityService
	licensingService    *service.LicensingService
	obligationService   *service.ObligationService
	violationService    *service.ViolationService
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(
	entityService *service.EntityService,
	licensingService *service.LicensingService,
	obligationService *service.ObligationService,
	violationService *service.ViolationService,
) *ComplianceHandler {
	return &ComplianceHandler{
		entityService:     entityService,
		licensingService:  licensingService,
		obligationService: obligationService,
		violationService:  violationService,
	}
}

// HealthCheck returns service health status
func (h *ComplianceHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "compliance-service",
		"timestamp": time.Now().UTC(),
	})
}

// CreateEntity creates a new regulated entity
func (h *ComplianceHandler) CreateEntity(c *gin.Context) {
	var entity domain.RegulatedEntity
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := c.GetString("actor_id")
	if err := h.entityService.CreateEntity(c.Request.Context(), &entity, actorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entity)
}

// GetEntity retrieves an entity by ID
func (h *ComplianceHandler) GetEntity(c *gin.Context) {
	entityID := c.Param("id")
	entity, err := h.entityService.GetEntity(c.Request.Context(), entityID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entity not found"})
		return
	}
	c.JSON(http.StatusOK, entity)
}

// ListEntities lists entities with filters
func (h *ComplianceHandler) ListEntities(c *gin.Context) {
	filter := port.EntityFilter{
		Limit:  100,
		Offset: 0,
	}

	if status := c.QueryArray("status"); len(status) > 0 {
		for _, s := range status {
			filter.Status = append(filter.Status, domain.EntityStatus(s))
		}
	}

	if types := c.QueryArray("type"); len(types) > 0 {
		for _, t := range types {
			filter.Type = append(filter.Type, domain.EntityType(t))
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	entities, err := h.entityService.ListEntities(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entities": entities,
		"count":    len(entities),
	})
}

// UpdateEntity updates an entity
func (h *ComplianceHandler) UpdateEntity(c *gin.Context) {
	entityID := c.Param("id")
	var entity domain.RegulatedEntity
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity.ID = entityID

	actorID := c.GetString("actor_id")
	if err := h.entityService.UpdateEntity(c.Request.Context(), &entity, actorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// ActivateEntity activates an entity
func (h *ComplianceHandler) ActivateEntity(c *gin.Context) {
	entityID := c.Param("id")
	actorID := c.GetString("actor_id")

	if err := h.entityService.ActivateEntity(c.Request.Context(), entityID, actorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "entity activated"})
}

// SuspendEntity suspends an entity
func (h *ComplianceHandler) SuspendEntity(c *gin.Context) {
	entityID := c.Param("id")
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	actorID := c.GetString("actor_id")
	if err := h.entityService.SuspendEntity(c.Request.Context(), entityID, req.Reason, actorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "entity suspended"})
}

// CreateLicense creates a new license application
func (h *ComplianceHandler) CreateLicense(c *gin.Context) {
	var license domain.License
	if err := c.ShouldBindJSON(&license); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := c.GetString("actor_id")
	if err := h.licensingService.CreateLicense(c.Request.Context(), &license, actorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, license)
}

// GetLicense retrieves a license by ID
func (h *ComplianceHandler) GetLicense(c *gin.Context) {
	licenseID := c.Param("id")
	license, err := h.licensingService.GetLicense(c.Request.Context(), licenseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "license not found"})
		return
	}
	c.JSON(http.StatusOK, license)
}

// ListLicenses lists licenses with filters
func (h *ComplianceHandler) ListLicenses(c *gin.Context) {
	filter := port.LicenseFilter{Limit: 100, Offset: 0}

	if entityID := c.Query("entity_id"); entityID != "" {
		filter.EntityID = entityID
	}

	if status := c.QueryArray("status"); len(status) > 0 {
		for _, s := range status {
			filter.Status = append(filter.Status, domain.LicenseStatus(s))
		}
	}

	licenses, err := h.licensingService.ListLicenses(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"licenses": licenses,
		"count":    len(licenses),
	})
}

// ApproveLicense approves a license
func (h *ComplianceHandler) ApproveLicense(c *gin.Context) {
	licenseID := c.Param("id")
	var req struct {
		Conditions []domain.LicenseCondition `json:"conditions"`
	}
	c.ShouldBindJSON(&req)

	actorID := c.GetString("actor_id")
	if err := h.licensingService.ApproveLicense(c.Request.Context(), licenseID, actorID, req.Conditions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "license approved"})
}

// SuspendLicense suspends a license
func (h *ComplianceHandler) SuspendLicense(c *gin.Context) {
	licenseID := c.Param("id")
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	actorID := c.GetString("actor_id")
	if err := h.licensingService.SuspendLicense(c.Request.Context(), licenseID, req.Reason, actorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "license suspended"})
}

// RevokeLicense revokes a license
func (h *ComplianceHandler) RevokeLicense(c *gin.Context) {
	licenseID := c.Param("id")
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	actorID := c.GetString("actor_id")
	if err := h.licensingService.RevokeLicense(c.Request.Context(), licenseID, req.Reason, actorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "license revoked"})
}

// CreateObligation creates a new obligation
func (h *ComplianceHandler) CreateObligation(c *gin.Context) {
	var obligation domain.ComplianceObligation
	if err := c.ShouldBindJSON(&obligation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := c.GetString("actor_id")
	if err := h.obligationService.CreateObligation(c.Request.Context(), &obligation, actorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, obligation)
}

// GetObligation retrieves an obligation by ID
func (h *ComplianceHandler) GetObligation(c *gin.Context) {
	obligationID := c.Param("id")
	obligation, err := h.obligationService.GetObligation(c.Request.Context(), obligationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "obligation not found"})
		return
	}
	c.JSON(http.StatusOK, obligation)
}

// ListObligations lists obligations with filters
func (h *ComplianceHandler) ListObligations(c *gin.Context) {
	filter := port.ObligationFilter{Limit: 100, Offset: 0}

	if entityID := c.Query("entity_id"); entityID != "" {
		filter.EntityID = entityID
	}

	obligations, err := h.obligationService.ListObligations(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"obligations": obligations,
		"count":       len(obligations),
	})
}

// VerifyObligation verifies a submitted obligation
func (h *ComplianceHandler) VerifyObligation(c *gin.Context) {
	obligationID := c.Param("id")
	var req struct {
		Notes string `json:"notes"`
	}
	c.ShouldBindJSON(&req)

	actorID := c.GetString("actor_id")
	if err := h.obligationService.VerifyObligation(c.Request.Context(), obligationID, req.Notes, actorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "obligation verified"})
}

// GetOverdueObligations retrieves all overdue obligations
func (h *ComplianceHandler) GetOverdueObligations(c *gin.Context) {
	obligations, err := h.obligationService.GetOverdueObligations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"obligations": obligations,
		"count":       len(obligations),
	})
}

// CreateViolation creates a new violation
func (h *ComplianceHandler) CreateViolation(c *gin.Context) {
	var violation domain.ComplianceViolation
	if err := c.ShouldBindJSON(&violation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := c.GetString("actor_id")
	if err := h.violationService.CreateViolation(c.Request.Context(), &violation, actorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, violation)
}

// GetViolation retrieves a violation by ID
func (h *ComplianceHandler) GetViolation(c *gin.Context) {
	violationID := c.Param("id")
	violation, err := h.violationService.GetViolation(c.Request.Context(), violationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "violation not found"})
		return
	}
	c.JSON(http.StatusOK, violation)
}

// ListViolations lists violations with filters
func (h *ComplianceHandler) ListViolations(c *gin.Context) {
	filter := port.ViolationFilter{Limit: 100, Offset: 0}

	if entityID := c.Query("entity_id"); entityID != "" {
		filter.EntityID = entityID
	}

	violations, err := h.violationService.ListViolations(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"violations": violations,
		"count":      len(violations),
	})
}

// IssuePenalty issues a penalty for a violation
func (h *ComplianceHandler) IssuePenalty(c *gin.Context) {
	violationID := c.Param("id")
	var penalty domain.Penalty
	if err := c.ShouldBindJSON(&penalty); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	penalty.ViolationID = violationID

	actorID := c.GetString("actor_id")
	if err := h.violationService.IssuePenalty(c.Request.Context(), violationID, &penalty, actorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, penalty)
}

// ResolveViolation resolves a violation
func (h *ComplianceHandler) ResolveViolation(c *gin.Context) {
	violationID := c.Param("id")
	var req struct {
		CorrectiveAction string `json:"corrective_action"`
		PreventiveAction string `json:"preventive_action"`
	}
	c.ShouldBindJSON(&req)

	actorID := c.GetString("actor_id")
	if err := h.violationService.ResolveViolation(c.Request.Context(), violationID, req.CorrectiveAction, req.PreventiveAction, actorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "violation resolved"})
}

// GetOpenViolations retrieves all open violations for an entity
func (h *ComplianceHandler) GetOpenViolations(c *gin.Context) {
	entityID := c.Param("id")
	violations, err := h.violationService.GetOpenViolations(c.Request.Context(), entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"violations": violations,
		"count":      len(violations),
	})
}
