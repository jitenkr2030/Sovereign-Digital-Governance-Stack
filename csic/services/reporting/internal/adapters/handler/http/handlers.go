package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/reporting-service/reporting/internal/core/domain"
	"github.com/reporting-service/reporting/internal/core/ports"
	"github.com/reporting-service/reporting/internal/core/services"
	"github.com/gin-gonic/gin"
)

// ReportingHandler handles HTTP requests for reporting operations.
type ReportingHandler struct {
	reportingService *services.ReportingService
}

// NewReportingHandler creates a new ReportingHandler.
func NewReportingHandler(reportingService *services.ReportingService) *ReportingHandler {
	return &ReportingHandler{
		reportingService: reportingService,
	}
}

// RegisterRoutes registers all reporting routes.
func (h *ReportingHandler) RegisterRoutes(router *gin.Engine) {
	// Health check endpoints
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// SAR endpoints
		v1.POST("/sar", h.CreateSAR)
		v1.GET("/sar", h.ListSARs)
		v1.GET("/sar/:id", h.GetSAR)
		v1.PUT("/sar/:id/status", h.UpdateSARStatus)
		v1.POST("/sar/:id/submit", h.SubmitSAR)

		// CTR endpoints
		v1.POST("/ctr", h.CreateCTR)
		v1.GET("/ctr", h.ListCTRs)
		v1.GET("/ctr/:id", h.GetCTR)

		// Compliance rules endpoints
		v1.POST("/rules", h.CreateComplianceRule)
		v1.GET("/rules", h.ListComplianceRules)
		v1.GET("/rules/:id", h.GetComplianceRule)

		// Alert endpoints
		v1.POST("/alerts", h.CreateAlert)
		v1.GET("/alerts", h.ListAlerts)
		v1.GET("/alerts/:id", h.GetAlert)
		v1.POST("/alerts/:id/resolve", h.ResolveAlert)
		v1.POST("/alerts/:id/escalate", h.EscalateAlert)

		// Compliance check endpoints
		v1.POST("/compliance/check", h.PerformComplianceCheck)

		// Report generation endpoints
		v1.POST("/reports", h.GenerateComplianceReport)
		v1.GET("/reports", h.ListReports)
		v1.GET("/reports/:id", h.GetReport)

		// Filing endpoints
		v1.GET("/filings", h.ListFilings)
	}
}

// HealthCheck handles GET /health
func (h *ReportingHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "reporting-service",
	})
}

// ReadinessCheck handles GET /ready
func (h *ReportingHandler) ReadinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "reporting-service",
	})
}

// ==================== SAR Handlers ====================

// CreateSARRequest represents the request body for creating a SAR.
type CreateSARRequest struct {
	SubjectID         string   `json:"subject_id" binding:"required"`
	SubjectType       string   `json:"subject_type" binding:"required"`
	SubjectName       string   `json:"subject_name" binding:"required"`
	SuspiciousActivity string  `json:"suspicious_activity" binding:"required"`
	ActivityDate      string   `json:"activity_date" binding:"required"`
	DollarAmount      float64  `json:"dollar_amount" binding:"required,gt=0"`
	Currency          string   `json:"currency"`
	TransactionCount  int      `json:"transaction_count"`
	Narrative         string   `json:"narrative" binding:"required"`
	RiskIndicators    []string `json:"risk_indicators"`
	FilingInstitution string   `json:"filing_institution" binding:"required"`
	ReporterID        string   `json:"reporter_id" binding:"required"`
	FieldOffice       string   `json:"field_office"`
	OriginatingAgency string   `json:"originating_agency"`
}

// CreateSAR handles POST /api/v1/sar
func (h *ReportingHandler) CreateSAR(c *gin.Context) {
	var req CreateSARRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	activityDate, err := time.Parse(time.RFC3339, req.ActivityDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity_date format"})
		return
	}

	sar := &domain.SAR{
		SubjectID:          req.SubjectID,
		SubjectType:        req.SubjectType,
		SubjectName:        req.SubjectName,
		SuspiciousActivity: req.SuspiciousActivity,
		ActivityDate:       activityDate,
		DollarAmount:       req.DollarAmount,
		Currency:           req.Currency,
		TransactionCount:   req.TransactionCount,
		Narrative:          req.Narrative,
		RiskIndicators:     req.RiskIndicators,
		FilingInstitution:  req.FilingInstitution,
		ReporterID:         req.ReporterID,
		FieldOffice:        req.FieldOffice,
		OriginatingAgency:  req.OriginatingAgency,
	}

	if err := h.reportingService.CreateSAR(c.Request.Context(), sar); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sar)
}

// ListSARs handles GET /api/v1/sar
func (h *ReportingHandler) ListSARs(c *gin.Context) {
	var filter ports.SARFilter

	if statusStr := c.Query("status"); statusStr != "" {
		filter.Status = []domain.ReportStatus{domain.ReportStatus(statusStr)}
	}

	if subjectID := c.Query("subject_id"); subjectID != "" {
		filter.SubjectID = subjectID
	}

	if reporterID := c.Query("reporter_id"); reporterID != "" {
		filter.ReporterID = reporterID
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	sars, err := h.reportingService.ListSARs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   sars,
		"total":  len(sars),
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetSAR handles GET /api/v1/sar/:id
func (h *ReportingHandler) GetSAR(c *gin.Context) {
	id := c.Param("id")

	sar, err := h.reportingService.GetSAR(c.Request.Context(), id)
	if err != nil {
		switch err {
		case services.ErrReportNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "SAR not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, sar)
}

// UpdateSARStatusRequest represents the request body for updating SAR status.
type UpdateSARStatusRequest struct {
	Status     string `json:"status" binding:"required"`
	ReviewerID string `json:"reviewer_id" binding:"required"`
}

// UpdateSARStatus handles PUT /api/v1/sar/:id/status
func (h *ReportingHandler) UpdateSARStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateSARStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.reportingService.UpdateSARStatus(
		c.Request.Context(),
		id,
		domain.ReportStatus(req.Status),
		req.ReviewerID,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SAR status updated"})
}

// SubmitSARRequest represents the request body for submitting a SAR.
type SubmitSARRequest struct {
	FilingMethod string `json:"filing_method" binding:"required"`
	FilingAgency string `json:"filing_agency" binding:"required"`
}

// SubmitSAR handles POST /api/v1/sar/:id/submit
func (h *ReportingHandler) SubmitSAR(c *gin.Context) {
	id := c.Param("id")

	var req SubmitSARRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.reportingService.SubmitSAR(c.Request.Context(), id, req.FilingMethod, req.FilingAgency)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SAR submitted successfully"})
}

// ==================== CTR Handlers ====================

// CreateCTRRequest represents the request body for creating a CTR.
type CreateCTRRequest struct {
	PersonName      string          `json:"person_name" binding:"required"`
	PersonIDType    string          `json:"person_id_type" binding:"required"`
	PersonIDNumber  string          `json:"person_id_number" binding:"required"`
	PersonAddress   domain.Address  `json:"person_address" binding:"required"`
	AccountNumber   string          `json:"account_number" binding:"required"`
	TransactionType string          `json:"transaction_type" binding:"required"`
	CashInAmount    float64         `json:"cash_in_amount"`
	CashOutAmount   float64         `json:"cash_out_amount"`
	Currency        string          `json:"currency"`
	TransactionDate string          `json:"transaction_date" binding:"required"`
	MethodReceived  string          `json:"method_received" binding:"required"`
	InstitutionID   string          `json:"institution_id" binding:"required"`
	BranchID        string          `json:"branch_id"`
	TellerID        string          `json:"teller_id"`
}

// CreateCTR handles POST /api/v1/ctr
func (h *ReportingHandler) CreateCTR(c *gin.Context) {
	var req CreateCTRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transactionDate, err := time.Parse(time.RFC3339, req.TransactionDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction_date format"})
		return
	}

	ctr := &domain.CTR{
		PersonName:       req.PersonName,
		PersonIDType:     req.PersonIDType,
		PersonIDNumber:   req.PersonIDNumber,
		PersonAddress:    req.PersonAddress,
		AccountNumber:    req.AccountNumber,
		TransactionType:  req.TransactionType,
		CashInAmount:     req.CashInAmount,
		CashOutAmount:    req.CashOutAmount,
		TotalAmount:      req.CashInAmount + req.CashOutAmount,
		Currency:         req.Currency,
		TransactionDate:  transactionDate,
		MethodReceived:   req.MethodReceived,
		InstitutionID:    req.InstitutionID,
		BranchID:         req.BranchID,
		TellerID:         req.TellerID,
	}

	if err := h.reportingService.CreateCTR(c.Request.Context(), ctr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ctr)
}

// ListCTRs handles GET /api/v1/ctr
func (h *ReportingHandler) ListCTRs(c *gin.Context) {
	var filter ports.CTRFilter

	if statusStr := c.Query("status"); statusStr != "" {
		filter.Status = []domain.ReportStatus{domain.ReportStatus(statusStr)}
	}

	if personID := c.Query("person_id"); personID != "" {
		filter.PersonID = personID
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	ctrs, err := h.reportingService.ListCTRs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   ctrs,
		"total":  len(ctrs),
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetCTR handles GET /api/v1/ctr/:id
func (h *ReportingHandler) GetCTR(c *gin.Context) {
	id := c.Param("id")

	ctr, err := h.reportingService.GetCTR(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "CTR not found"})
		return
	}

	c.JSON(http.StatusOK, ctr)
}

// ==================== Compliance Rule Handlers ====================

// CreateComplianceRuleRequest represents the request body for creating a compliance rule.
type CreateComplianceRuleRequest struct {
	RuleCode      string               `json:"rule_code" binding:"required"`
	RuleName      string               `json:"rule_name" binding:"required"`
	Category      string               `json:"category" binding:"required"`
	Regulation    string               `json:"regulation" binding:"required"`
	Description   string               `json:"description"`
	Threshold     *domain.RuleThreshold `json:"threshold"`
	Logic         string               `json:"logic"`
	RiskScore     int                  `json:"risk_score"`
	Severity      string               `json:"severity"`
	EffectiveDate string               `json:"effective_date" binding:"required"`
}

// CreateComplianceRule handles POST /api/v1/rules
func (h *ReportingHandler) CreateComplianceRule(c *gin.Context) {
	var req CreateComplianceRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	effectiveDate, err := time.Parse(time.RFC3339, req.EffectiveDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid effective_date format"})
		return
	}

	rule := &domain.ComplianceRule{
		RuleCode:      req.RuleCode,
		RuleName:      req.RuleName,
		Category:      req.Category,
		Regulation:    req.Regulation,
		Description:   req.Description,
		Threshold:     req.Threshold,
		Logic:         req.Logic,
		RiskScore:     req.RiskScore,
		Severity:      req.Severity,
		IsActive:      true,
		EffectiveDate: effectiveDate,
	}

	if err := h.reportingService.CreateComplianceRule(c.Request.Context(), rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// ListComplianceRules handles GET /api/v1/rules
func (h *ReportingHandler) ListComplianceRules(c *gin.Context) {
	var filter ports.ComplianceRuleFilter

	if category := c.Query("category"); category != "" {
		filter.Category = category
	}

	if regulation := c.Query("regulation"); regulation != "" {
		filter.Regulation = regulation
	}

	if severity := c.Query("severity"); severity != "" {
		filter.Severity = severity
	}

	rules, err := h.reportingService.ListComplianceRules(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   rules,
		"total":  len(rules),
	})
}

// GetComplianceRule handles GET /api/v1/rules/:id
func (h *ReportingHandler) GetComplianceRule(c *gin.Context) {
	id := c.Param("id")

	rule, err := h.reportingService.GetComplianceRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// ==================== Alert Handlers ====================

// CreateAlertRequest represents the request body for creating an alert.
type CreateAlertRequest struct {
	AlertType      string   `json:"alert_type" binding:"required"`
	Severity       string   `json:"severity" binding:"required"`
	SubjectID      string   `json:"subject_id" binding:"required"`
	SubjectType    string   `json:"subject_type" binding:"required"`
	SubjectName    string   `json:"subject_name" binding:"required"`
	TransactionIDs []string `json:"transaction_ids"`
	Description    string   `json:"description" binding:"required"`
	TriggeredRules []string `json:"triggered_rules"`
	RiskScore      int      `json:"risk_score"`
}

// CreateAlert handles POST /api/v1/alerts
func (h *ReportingHandler) CreateAlert(c *gin.Context) {
	var req CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alert := &domain.Alert{
		AlertType:      req.AlertType,
		Severity:       req.Severity,
		SubjectID:      req.SubjectID,
		SubjectType:    req.SubjectType,
		SubjectName:    req.SubjectName,
		TransactionIDs: req.TransactionIDs,
		Description:    req.Description,
		TriggeredRules: req.TriggeredRules,
		RiskScore:      req.RiskScore,
	}

	if err := h.reportingService.CreateAlert(c.Request.Context(), alert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, alert)
}

// ListAlerts handles GET /api/v1/alerts
func (h *ReportingHandler) ListAlerts(c *gin.Context) {
	var filter ports.AlertFilter

	if statusStr := c.Query("status"); statusStr != "" {
		filter.Status = []string{statusStr}
	}

	if severityStr := c.Query("severity"); severityStr != "" {
		filter.Severity = []string{severityStr}
	}

	if alertType := c.Query("alert_type"); alertType != "" {
		filter.AlertType = alertType
	}

	if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		filter.AssignedTo = assignedTo
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	alerts, err := h.reportingService.ListAlerts(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   alerts,
		"total":  len(alerts),
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetAlert handles GET /api/v1/alerts/:id
func (h *ReportingHandler) GetAlert(c *gin.Context) {
	id := c.Param("id")

	alert, err := h.reportingService.GetAlert(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// ResolveAlertRequest represents the request body for resolving an alert.
type ResolveAlertRequest struct {
	Resolution string `json:"resolution" binding:"required"`
	ResolvedBy string `json:"resolved_by" binding:"required"`
}

// ResolveAlert handles POST /api/v1/alerts/:id/resolve
func (h *ReportingHandler) ResolveAlert(c *gin.Context) {
	id := c.Param("id")

	var req ResolveAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.reportingService.ResolveAlert(c.Request.Context(), id, req.Resolution, req.ResolvedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert resolved"})
}

// EscalateAlertRequest represents the request body for escalating an alert.
type EscalateAlertRequest struct {
	AssignedTo string `json:"assigned_to" binding:"required"`
}

// EscalateAlert handles POST /api/v1/alerts/:id/escalate
func (h *ReportingHandler) EscalateAlert(c *gin.Context) {
	id := c.Param("id")

	var req EscalateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.reportingService.EscalateAlert(c.Request.Context(), id, req.AssignedTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert escalated"})
}

// ==================== Compliance Check Handler ====================

// PerformComplianceCheckRequest represents the request body for performing a compliance check.
type PerformComplianceCheckRequest struct {
	SubjectID   string `json:"subject_id" binding:"required"`
	SubjectType string `json:"subject_type" binding:"required"`
	CheckType   string `json:"check_type" binding:"required"`
}

// PerformComplianceCheck handles POST /api/v1/compliance/check
func (h *ReportingHandler) PerformComplianceCheck(c *gin.Context) {
	var req PerformComplianceCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	check, err := h.reportingService.PerformComplianceCheck(
		c.Request.Context(),
		req.SubjectID,
		req.SubjectType,
		req.CheckType,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, check)
}

// ==================== Report Generation Handlers ====================

// GenerateComplianceReportRequest represents the request body for generating a report.
type GenerateComplianceReportRequest struct {
	ReportType   string `json:"report_type" binding:"required"`
	PeriodStart  string `json:"period_start" binding:"required"`
	PeriodEnd    string `json:"period_end" binding:"required"`
	GeneratedBy  string `json:"generated_by" binding:"required"`
}

// GenerateComplianceReport handles POST /api/v1/reports
func (h *ReportingHandler) GenerateComplianceReport(c *gin.Context) {
	var req GenerateComplianceReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	periodStart, err := time.Parse(time.RFC3339, req.PeriodStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_start format"})
		return
	}

	periodEnd, err := time.Parse(time.RFC3339, req.PeriodEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_end format"})
		return
	}

	report, err := h.reportingService.GenerateComplianceReport(
		c.Request.Context(),
		domain.ReportType(req.ReportType),
		periodStart,
		periodEnd,
		req.GeneratedBy,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, report)
}

// ListReports handles GET /api/v1/reports
func (h *ReportingHandler) ListReports(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// GetReport handles GET /api/v1/reports/:id
func (h *ReportingHandler) GetReport(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// ==================== Filing Handlers ====================

// ListFilings handles GET /api/v1/filings
func (h *ReportingHandler) ListFilings(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
