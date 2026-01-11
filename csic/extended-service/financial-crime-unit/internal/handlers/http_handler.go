package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jitenkr2030/csic-platform/extended-services/financial-crime-unit/internal/domain"
	"github.com/jitenkr2030/csic-platform/extended-services/financial-crime-unit/internal/service"
)

// FCUHandler handles HTTP requests for financial crime operations
type FCUHandler struct {
	service service.FCUService
}

// NewFCUHandler creates a new FCU handler
func NewFCUHandler(svc service.FCUService) *FCUHandler {
	return &FCUHandler{
		service: svc,
	}
}

// Transaction screening endpoints

// ScreenTransaction handles transaction screening requests
func (h *FCUHandler) ScreenTransaction(c *gin.Context) {
	var req service.TransactionScreenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.ScreenTransaction(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "screening_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ScreenEntity handles entity screening requests
func (h *FCUHandler) ScreenEntity(c *gin.Context) {
	var req service.EntityScreenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.ScreenEntity(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "screening_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ScreenBatch handles batch screening requests
func (h *FCUHandler) ScreenBatch(c *gin.Context) {
	var req service.BatchScreenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	results, err := h.service.ScreenBatch(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "screening_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":     results,
		"total_count": len(results),
	})
}

// Sanctions screening endpoints

// CheckSanctions handles sanctions check requests
func (h *FCUHandler) CheckSanctions(c *gin.Context) {
	var req service.SanctionsCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.CheckSanctions(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "sanctions_check_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// BulkSanctionsCheck handles bulk sanctions checking
func (h *FCUHandler) BulkSanctionsCheck(c *gin.Context) {
	var req service.BulkSanctionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	results, err := h.service.BulkSanctionsCheck(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "bulk_check_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":     results,
		"total_count": len(results),
	})
}

// GetSanctionLists retrieves available sanction lists
func (h *FCUHandler) GetSanctionLists(c *gin.Context) {
	lists, err := h.service.GetSanctionLists(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"lists": lists,
	})
}

// SyncSanctionLists triggers a sanctions list sync
func (h *FCUHandler) SyncSanctionLists(c *gin.Context) {
	if err := h.service.SyncSanctionLists(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "sync_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sanctions list sync initiated",
	})
}

// Alerts management endpoints

// GetAlerts retrieves alerts based on filter criteria
func (h *FCUHandler) GetAlerts(c *gin.Context) {
	var filter service.AlertFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 50
	}

	alerts, err := h.service.GetAlerts(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts":     alerts,
		"total_count": len(alerts),
	})
}

// GetAlert retrieves a specific alert
func (h *FCUHandler) GetAlert(c *gin.Context) {
	id := c.Param("id")

	alert, err := h.service.GetAlert(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	if alert == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Alert not found",
		})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// AcknowledgeAlert acknowledges an alert
func (h *FCUHandler) AcknowledgeAlert(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.AcknowledgeAlert(c.Request.Context(), id, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "acknowledgment_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert acknowledged",
	})
}

// EscalateAlert escalates an alert
func (h *FCUHandler) EscalateAlert(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		EscalateTo string `json:"escalate_to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.EscalateAlert(c.Request.Context(), id, req.EscalateTo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "escalation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert escalated",
	})
}

// GetAlertsBySeverity retrieves alerts by severity
func (h *FCUHandler) GetAlertsBySeverity(c *gin.Context) {
	severity := c.Param("severity")

	alerts, err := h.service.GetAlerts(c.Request.Context(), service.AlertFilterRequest{
		Severity: severity,
		Limit:    100,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"severity":    severity,
		"alerts":      alerts,
		"total_count": len(alerts),
	})
}

// GetAlertsByStatus retrieves alerts by status
func (h *FCUHandler) GetAlertsByStatus(c *gin.Context) {
	status := c.Param("status")

	alerts, err := h.service.GetAlerts(c.Request.Context(), service.AlertFilterRequest{
		Status: status,
		Limit:  100,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      status,
		"alerts":      alerts,
		"total_count": len(alerts),
	})
}

// Case management endpoints

// GetCases retrieves cases based on filter criteria
func (h *FCUHandler) GetCases(c *gin.Context) {
	var filter service.CaseFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 50
	}

	cases, err := h.service.GetCases(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cases":       cases,
		"total_count": len(cases),
	})
}

// CreateCase creates a new case
func (h *FCUHandler) CreateCase(c *gin.Context) {
	var req service.CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	caseObj, err := h.service.CreateCase(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "creation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Case created successfully",
		"case_id":    caseObj.ID,
		"case_number": caseObj.CaseNumber,
	})
}

// GetCase retrieves a specific case
func (h *FCUHandler) GetCase(c *gin.Context) {
	id := c.Param("id")

	caseObj, err := h.service.GetCase(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	if caseObj == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Case not found",
		})
		return
	}

	c.JSON(http.StatusOK, caseObj)
}

// UpdateCase updates a case
func (h *FCUHandler) UpdateCase(c *gin.Context) {
	id := c.Param("id")

	var caseObj domain.Case
	if err := c.ShouldBindJSON(&caseObj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}
	caseObj.ID = id

	if err := h.service.UpdateCase(c.Request.Context(), &caseObj); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "update_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Case updated successfully",
	})
}

// TransitionCaseWorkflow transitions a case to a new workflow state
func (h *FCUHandler) TransitionCaseWorkflow(c *gin.Context) {
	caseID := c.Param("id")

	var req struct {
		NewStatus string `json:"new_status" binding:"required"`
		UserID    string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.TransitionCaseWorkflow(c.Request.Context(), caseID, req.NewStatus, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "transition_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Case workflow transitioned",
		"new_status": req.NewStatus,
	})
}

// AssignCase assigns a case to an investigator
func (h *FCUHandler) AssignCase(c *gin.Context) {
	caseID := c.Param("id")

	var req struct {
		AssigneeID string `json:"assignee_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.AssignCase(c.Request.Context(), caseID, req.AssigneeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "assignment_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Case assigned",
		"assignee_id": req.AssigneeID,
	})
}

// AddEvidenceToCase adds evidence to a case
func (h *FCUHandler) AddEvidenceToCase(c *gin.Context) {
	caseID := c.Param("id")

	var evidence domain.Evidence
	if err := c.ShouldBindJSON(&evidence); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.AddEvidenceToCase(c.Request.Context(), caseID, &evidence); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "addition_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Evidence added",
		"evidence_id": evidence.ID,
	})
}

// GetCasesByAssignee retrieves cases assigned to a specific user
func (h *FCUHandler) GetCasesByAssignee(c *gin.Context) {
	assignee := c.Param("assignee")

	cases, err := h.service.GetCases(c.Request.Context(), service.CaseFilterRequest{
		AssigneeID: assignee,
		Limit:      100,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assignee":    assignee,
		"cases":       cases,
		"total_count": len(cases),
	})
}

// GetCasesByStatus retrieves cases by status
func (h *FCUHandler) GetCasesByStatus(c *gin.Context) {
	status := c.Param("status")

	cases, err := h.service.GetCases(c.Request.Context(), service.CaseFilterRequest{
		Status: status,
		Limit:  100,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      status,
		"cases":       cases,
		"total_count": len(cases),
	})
}

// SAR endpoints

// GetSARs retrieves SARs based on filter criteria
func (h *FCUHandler) GetSARs(c *gin.Context) {
	var filter service.SARFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 50
	}

	sars, err := h.service.GetSARs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sars":        sars,
		"total_count": len(sars),
	})
}

// CreateSAR creates a new SAR
func (h *FCUHandler) CreateSAR(c *gin.Context) {
	var req service.CreateSARRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	sar, err := h.service.CreateSAR(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "creation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "SAR created successfully",
		"sar_id":    sar.ID,
		"sar_number": sar.SARNumber,
	})
}

// GetSAR retrieves a specific SAR
func (h *FCUHandler) GetSAR(c *gin.Context) {
	id := c.Param("id")

	sar, err := h.service.GetSAR(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	if sar == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "SAR not found",
		})
		return
	}

	c.JSON(http.StatusOK, sar)
}

// UpdateSAR updates a SAR
func (h *FCUHandler) UpdateSAR(c *gin.Context) {
	id := c.Param("id")

	var sar domain.SAR
	if err := c.ShouldBindJSON(&sar); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}
	sar.ID = id

	if err := h.service.UpdateSAR(c.Request.Context(), &sar); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "update_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SAR updated successfully",
	})
}

// ValidateSAR validates a SAR
func (h *FCUHandler) ValidateSAR(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.ValidateSAR(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "validation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SAR validated successfully",
	})
}

// SubmitSAR submits a SAR
func (h *FCUHandler) SubmitSAR(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		SubmittedBy string `json:"submitted_by" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.SubmitSAR(c.Request.Context(), id, req.SubmittedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "submission_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SAR submitted successfully",
	})
}

// GenerateAutoSAR generates an automated SAR for a case
func (h *FCUHandler) GenerateAutoSAR(c *gin.Context) {
	caseID := c.Param("id")

	sar, err := h.service.GenerateAutoSAR(c.Request.Context(), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "generation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Auto-SAR generated successfully",
		"sar_id":    sar.ID,
		"sar_number": sar.SARNumber,
	})
}

// GetSARsByCase retrieves SARs for a specific case
func (h *FCUHandler) GetSARsByCase(c *gin.Context) {
	caseID := c.Param("case_id")

	sars, err := h.service.GetSARs(c.Request.Context(), service.SARFilterRequest{
		CaseID: caseID,
		Limit:  50,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"case_id":    caseID,
		"sars":       sars,
		"total_count": len(sars),
	})
}

// Risk scoring endpoints

// GetRiskScore retrieves the risk score for an entity
func (h *FCUHandler) GetRiskScore(c *gin.Context) {
	entityID := c.Param("id")

	result, err := h.service.GetRiskScore(c.Request.Context(), entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetRiskProfile retrieves the risk profile for an entity
func (h *FCUHandler) GetRiskProfile(c *gin.Context) {
	entityID := c.Param("id")

	profile, err := h.service.GetRiskProfile(c.Request.Context(), entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	if profile == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Risk profile not found",
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// GetRiskHistory retrieves the risk history for an entity
func (h *FCUHandler) GetRiskHistory(c *gin.Context) {
	entityID := c.Param("id")
	
	limit := 30
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	history, err := h.service.GetRiskHistory(c.Request.Context(), entityID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id":   entityID,
		"risk_history": history,
		"total_count": len(history),
	})
}

// CalculateRisk calculates risk for an entity
func (h *FCUHandler) CalculateRisk(c *gin.Context) {
	var req service.CalculateRiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.CalculateRisk(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "calculation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Network analysis endpoints

// AnalyzeNetwork performs network analysis for an entity
func (h *FCUHandler) AnalyzeNetwork(c *gin.Context) {
	var req service.NetworkAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	graph, err := h.service.AnalyzeNetwork(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "analysis_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, graph)
}

// TraceTransactionFlow traces the transaction flow
func (h *FCUHandler) TraceTransactionFlow(c *gin.Context) {
	var req service.TraceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	flows, err := h.service.TraceTransactionFlow(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "tracing_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction_id": req.TransactionID,
		"flows":          flows,
		"total_count":    len(flows),
	})
}

// GetNetworkPatterns retrieves network patterns for an entity
func (h *FCUHandler) GetNetworkPatterns(c *gin.Context) {
	entityID := c.Param("id")

	patterns, err := h.service.GetNetworkPatterns(c.Request.Context(), entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id": entityID,
		"patterns":  patterns,
	})
}

// GetEntityConnections retrieves connections for an entity
func (h *FCUHandler) GetEntityConnections(c *gin.Context) {
	entityID := c.Param("id")

	connections, err := h.service.GetEntityConnections(c.Request.Context(), entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id":   entityID,
		"connections": connections,
		"total_count": len(connections),
	})
}

// Entity management endpoints

// GetEntity retrieves an entity
func (h *FCUHandler) GetEntity(c *gin.Context) {
	id := c.Param("id")

	entity, err := h.service.GetEntity(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Entity not found",
		})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// UpdateEntity updates an entity
func (h *FCUHandler) UpdateEntity(c *gin.Context) {
	id := c.Param("id")

	var entity domain.Entity
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}
	entity.ID = id

	if err := h.service.UpdateEntity(c.Request.Context(), &entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "update_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Entity updated successfully",
	})
}

// GetEntityTransactions retrieves transactions for an entity
func (h *FCUHandler) GetEntityTransactions(c *gin.Context) {
	entityID := c.Param("id")

	limit := 100
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	transactions, err := h.service.GetEntityTransactions(c.Request.Context(), entityID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id":   entityID,
		"transactions": transactions,
		"total_count": len(transactions),
	})
}

// GetEntityAlerts retrieves alerts for an entity
func (h *FCUHandler) GetEntityAlerts(c *gin.Context) {
	entityID := c.Param("id")

	alerts, err := h.service.GetEntityAlerts(c.Request.Context(), entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id":   entityID,
		"alerts":      alerts,
		"total_count": len(alerts),
	})
}

// AddToWatchlist adds an entity to the watchlist
func (h *FCUHandler) AddToWatchlist(c *gin.Context) {
	var req service.AddToWatchlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.AddToWatchlist(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "addition_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Entity added to watchlist",
	})
}

// RemoveFromWatchlist removes an entity from the watchlist
func (h *FCUHandler) RemoveFromWatchlist(c *gin.Context) {
	entityID := c.Param("id")

	if err := h.service.RemoveFromWatchlist(c.Request.Context(), entityID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "removal_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Entity removed from watchlist",
	})
}

// GetWatchlistStatus retrieves watchlist status for an entity
func (h *FCUHandler) GetWatchlistStatus(c *gin.Context) {
	entityID := c.Param("id")

	status, err := h.service.GetWatchlistStatus(c.Request.Context(), entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	if status == nil {
		c.JSON(http.StatusOK, gin.H{
			"entity_id": entityID,
			"status":    "clear",
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// Investigation endpoints

// GetInvestigation retrieves an investigation
func (h *FCUHandler) GetInvestigation(c *gin.Context) {
	id := c.Param("id")

	investigation, err := h.service.GetInvestigation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	if investigation == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "Investigation not found",
		})
		return
	}

	c.JSON(http.StatusOK, investigation)
}

// StartInvestigation starts a new investigation
func (h *FCUHandler) StartInvestigation(c *gin.Context) {
	var req service.StartInvestigationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	investigation, err := h.service.StartInvestigation(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "creation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "Investigation started",
		"investigation_id": investigation.ID,
	})
}

// AddInvestigationNote adds a note to an investigation
func (h *FCUHandler) AddInvestigationNote(c *gin.Context) {
	var req service.AddNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.AddInvestigationNote(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "addition_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Note added to investigation",
	})
}

// AddTimelineEvent adds a timeline event to a case
func (h *FCUHandler) AddTimelineEvent(c *gin.Context) {
	var req service.AddTimelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.AddTimelineEvent(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "addition_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Timeline event added",
	})
}

// Reports endpoints

// GetReportSummary retrieves a report summary
func (h *FCUHandler) GetReportSummary(c *gin.Context) {
	var req service.ReportSummaryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	summary, err := h.service.GetReportSummary(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "generation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetTrendAnalysis retrieves trend analysis data
func (h *FCUHandler) GetTrendAnalysis(c *gin.Context) {
	var req service.TrendAnalysisRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	analysis, err := h.service.GetTrendAnalysis(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "analysis_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetDetectionEffectiveness retrieves detection effectiveness metrics
func (h *FCUHandler) GetDetectionEffectiveness(c *gin.Context) {
	var req service.EffectivenessRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	effectiveness, err := h.service.GetDetectionEffectiveness(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, effectiveness)
}

// GetComplianceReport retrieves a compliance report
func (h *FCUHandler) GetComplianceReport(c *gin.Context) {
	var req service.ComplianceReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	report, err := h.service.GetComplianceReport(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "generation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// HealthHandler handles health check endpoints
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Liveness handles liveness probe requests
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// Readiness handles readiness probe requests
func (h *HealthHandler) Readiness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// Health handles detailed health check requests
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "financial-crime-unit-service",
		"version":   "1.0.0",
	})
}
