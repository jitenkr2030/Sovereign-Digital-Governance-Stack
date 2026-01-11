package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/csic-licensing/internal/domain/models"
	"github.com/csic-licensing/internal/repository"
	"github.com/csic-licensing/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for the licensing service
type Handler struct {
	licenseService    *service.LicenseService
	complianceService *service.ComplianceService
	reportingService  *service.ReportingService
	logger            *zap.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(
	licenseService *service.LicenseService,
	complianceService *service.ComplianceService,
	reportingService *service.ReportingService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		licenseService:    licenseService,
		complianceService: complianceService,
		reportingService:  reportingService,
		logger:            logger,
	}
}

// SetupRouter sets up the Gin router
func (h *Handler) SetupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(h.loggingMiddleware())

	// Health endpoints
	router.GET("/health", h.healthCheck)
	router.GET("/ready", h.readinessCheck)

	// API v1
	v1 := router.Group("/api/v1")
	{
		// License endpoints
		licenses := v1.Group("/licenses")
		{
			licenses.POST("", h.createLicense)
			licenses.GET("", h.listLicenses)
			licenses.GET("/:id", h.getLicense)
			licenses.GET("/number/:number", h.getLicenseByNumber)
			licenses.PUT("/:id/status", h.updateLicenseStatus)
			licenses.POST("/:id/suspend", h.suspendLicense)
			licenses.POST("/:id/revoke", h.revokeLicense)
			licenses.POST("/:id/renew", h.renewLicense)
			licenses.GET("/:id/compliance", h.getLicenseCompliance)
		}

		// Exchange license endpoints
		exchanges := v1.Group("/exchanges")
		{
			exchanges.GET("/:exchange_id/licenses", h.getExchangeLicenses)
			exchanges.GET("/:exchange_id/compliance-score", h.getExchangeComplianceScore)
			exchanges.GET("/:exchange_id/dashboard", h.getExchangeDashboard)
		}

		// Violation endpoints
		violations := v1.Group("/violations")
		{
			violations.POST("", h.createViolation)
			violations.GET("", h.listViolations)
			violations.GET("/open", h.getOpenViolations)
			violations.GET("/:id", h.getViolation)
			violations.PUT("/:id/status", h.updateViolationStatus)
			violations.POST("/:id/assign", h.assignViolation)
			violations.POST("/:id/escalate", h.escalateViolation)
			violations.POST("/:id/resolve", h.resolveViolation)
		}

		// Report endpoints
		reports := v1.Group("/reports")
		{
			reports.POST("/generate", h.generateReport)
			reports.GET("", h.listReports)
			reports.GET("/:id", h.getReport)
			reports.GET("/:id/download", h.downloadReport)
			reports.POST("/:id/submit", h.submitReport)
			reports.GET("/overdue", h.getOverdueReports)
			reports.GET("/stats", h.getReportStats)
		}

		// Dashboard endpoints
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("/overview", h.getDashboardOverview)
			dashboard.GET("/metrics", h.getDashboardMetrics)
			dashboard.GET("/violation-trends", h.getViolationTrends)
			dashboard.GET("/activity", h.getRecentActivity)
		}

		// Compliance check endpoints
		compliance := v1.Group("/compliance")
		{
			compliance.GET("/checks/:license_id", h.getComplianceChecks)
			compliance.PUT("/checks/:id", h.updateComplianceCheck)
			compliance.POST("/checks", h.createComplianceCheck)
			compliance.GET("/kpis/:exchange_id", h.getKPIs)
		}
	}

	return router
}

// Health and readiness

func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

func (h *Handler) readinessCheck(c *gin.Context) {
	// In production, check database and other dependencies
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

func (h *Handler) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		h.logger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
		)
	}
}

// License endpoints

func (h *Handler) createLicense(c *gin.Context) {
	var req struct {
		ExchangeID    string `json:"exchange_id" binding:"required"`
		LicenseNumber string `json:"license_number" binding:"required"`
		Jurisdiction  string `json:"jurisdiction" binding:"required"`
		Type          string `json:"type" binding:"required"`
		IssuedDate    string `json:"issued_date"`
		ExpiryDate    string `json:"expiry_date" binding:"required"`
		ApprovedBy    string `json:"approved_by"`
		Notes         string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	issuedDate := time.Now()
	if req.IssuedDate != "" {
		issuedDate, _ = time.Parse(time.RFC3339, req.IssuedDate)
	}

	expiryDate, _ := time.Parse(time.RFC3339, req.ExpiryDate)

	license := models.NewLicense(req.ExchangeID, req.LicenseNumber, models.Jurisdiction(req.Jurisdiction), models.LicenseType(req.Type))
	license.IssuedDate = issuedDate
	license.ExpiryDate = expiryDate
	license.ApprovedBy = req.ApprovedBy
	license.Notes = req.Notes

	if err := h.licenseService.CreateLicense(c.Request.Context(), license); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, license)
}

func (h *Handler) getLicense(c *gin.Context) {
	id := c.Param("id")

	license, err := h.licenseService.GetLicense(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "License not found"})
		return
	}

	c.JSON(http.StatusOK, license)
}

func (h *Handler) getLicenseByNumber(c *gin.Context) {
	number := c.Param("number")

	license, err := h.licenseService.GetLicenseByNumber(c.Request.Context(), number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "License not found"})
		return
	}

	c.JSON(http.StatusOK, license)
}

func (h *Handler) listLicenses(c *gin.Context) {
	var filter repository.LicenseFilter
	filter.ExchangeID = c.Query("exchange_id")
	filter.Jurisdiction = models.Jurisdiction(c.Query("jurisdiction"))
	filter.Status = models.LicenseStatus(c.Query("status"))
	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	licenses, total, err := h.licenseService.ListLicenses(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  licenses,
		"total": total,
		"page":  filter.Page,
		"size":  filter.PageSize,
	})
}

func (h *Handler) updateLicenseStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status   string `json:"status" binding:"required"`
		ActorID  string `json:"actor_id" binding:"required"`
		Reason   string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.licenseService.UpdateLicenseStatus(c.Request.Context(), id, models.LicenseStatus(req.Status), req.ActorID, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *Handler) suspendLicense(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Reason  string `json:"reason" binding:"required"`
		ActorID string `json:"actor_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.licenseService.SuspendLicense(c.Request.Context(), id, req.Reason, req.ActorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "suspended"})
}

func (h *Handler) revokeLicense(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Reason  string `json:"reason" binding:"required"`
		ActorID string `json:"actor_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.licenseService.RevokeLicense(c.Request.Context(), id, req.Reason, req.ActorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "revoked"})
}

func (h *Handler) renewLicense(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		NewExpiryDate string `json:"new_expiry_date" binding:"required"`
		ActorID       string `json:"actor_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expiryDate, _ := time.Parse(time.RFC3339, req.NewExpiryDate)

	err := h.licenseService.RenewLicense(c.Request.Context(), id, expiryDate, req.ActorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "renewed"})
}

func (h *Handler) getLicenseCompliance(c *gin.Context) {
	id := c.Param("id")

	status, err := h.licenseService.GetLicenseComplianceStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *Handler) getExchangeLicenses(c *gin.Context) {
	exchangeID := c.Param("exchange_id")

	licenses, err := h.licenseService.GetLicensesByExchange(c.Request.Context(), exchangeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exchange_id": exchangeID,
		"licenses":    licenses,
		"count":       len(licenses),
	})
}

func (h *Handler) getExchangeComplianceScore(c *gin.Context) {
	exchangeID := c.Param("exchange_id")

	score, err := h.complianceService.GetComplianceScore(c.Request.Context(), exchangeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, score)
}

func (h *Handler) getExchangeDashboard(c *gin.Context) {
	exchangeID := c.Param("exchange_id")

	metrics, err := h.complianceService.GetDashboardMetrics(c.Request.Context(), exchangeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// Violation endpoints

func (h *Handler) createViolation(c *gin.Context) {
	var req struct {
		LicenseID  string `json:"license_id" binding:"required"`
		Severity   string `json:"severity" binding:"required"`
		Type       string `json:"type" binding:"required"`
		Title      string `json:"title" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	violation := models.NewViolation(req.LicenseID, "", models.ViolationSeverity(req.Severity), models.ViolationType(req.Type), req.Title, req.Description)

	if err := h.complianceService.CreateViolation(c.Request.Context(), violation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, violation)
}

func (h *Handler) getViolation(c *gin.Context) {
	id := c.Param("id")

	violations, _, err := h.complianceService.GetViolationsByFilter(c.Request.Context(), repository.ViolationFilter{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Find the specific violation
	for _, v := range violations {
		if v.ID == id {
			c.JSON(http.StatusOK, v)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Violation not found"})
}

func (h *Handler) listViolations(c *gin.Context) {
	var filter repository.ViolationFilter
	filter.LicenseID = c.Query("license_id")
	filter.Status = models.ViolationStatus(c.Query("status"))
	filter.Severity = models.ViolationSeverity(c.Query("severity"))
	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	violations, total, err := h.complianceService.GetViolationsByFilter(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  violations,
		"total": total,
		"page":  filter.Page,
		"size":  filter.PageSize,
	})
}

func (h *Handler) getOpenViolations(c *gin.Context) {
	violations, err := h.complianceService.GetOpenViolations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  violations,
		"count": len(violations),
	})
}

func (h *Handler) updateViolationStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status    string `json:"status" binding:"required"`
		UpdatedBy string `json:"updated_by" binding:"required"`
		Reason    string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.complianceService.UpdateViolationStatus(c.Request.Context(), id, models.ViolationStatus(req.Status), req.UpdatedBy, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *Handler) assignViolation(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Assignee string `json:"assignee" binding:"required"`
		Team     string `json:"team" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.complianceService.AssignViolation(c.Request.Context(), id, req.Assignee, req.Team)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "assigned"})
}

func (h *Handler) escalateViolation(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		EscalateTo string `json:"escalate_to" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.complianceService.EscalateViolation(c.Request.Context(), id, req.EscalateTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "escalated"})
}

func (h *Handler) resolveViolation(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		ResolvedBy      string `json:"resolved_by" binding:"required"`
		Resolution      string `json:"resolution" binding:"required"`
		RootCause       string `json:"root_cause"`
		CorrectiveAction string `json:"corrective_action"`
		PreventiveAction string `json:"preventive_action"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.complianceService.ResolveViolation(c.Request.Context(), id, req.ResolvedBy, req.Resolution, req.RootCause, req.CorrectiveAction, req.PreventiveAction)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "resolved"})
}

// Report endpoints

func (h *Handler) generateReport(c *gin.Context) {
	var req struct {
		ReportType   string `json:"report_type" binding:"required"`
		PeriodStart  string `json:"period_start" binding:"required"`
		PeriodEnd    string `json:"period_end" binding:"required"`
		LicenseID    string `json:"license_id" binding:"required"`
		GeneratedBy  string `json:"generated_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	periodStart, _ := time.Parse(time.RFC3339, req.PeriodStart)
	periodEnd, _ := time.Parse(time.RFC3339, req.PeriodEnd)

	report, err := h.reportingService.GenerateReport(c.Request.Context(), models.ReportType(req.ReportType), periodStart, periodEnd, req.LicenseID, req.GeneratedBy, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":   "Report generation initiated",
		"report_id": report.ID,
	})
}

func (h *Handler) getReport(c *gin.Context) {
	id := c.Param("id")

	report, err := h.reportingService.GetReport(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *Handler) listReports(c *gin.Context) {
	var filter repository.ReportFilter
	filter.LicenseID = c.Query("license_id")
	filter.Status = models.ReportStatus(c.Query("status"))
	filter.ReportType = models.ReportType(c.Query("type"))
	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	reports, total, err := h.reportingService.GetReportsByFilter(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  reports,
		"total": total,
		"page":  filter.Page,
		"size":  filter.PageSize,
	})
}

func (h *Handler) downloadReport(c *gin.Context) {
	id := c.Param("id")

	url, err := h.reportingService.DownloadReport(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": url,
	})
}

func (h *Handler) submitReport(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		SubmittedTo string `json:"submitted_to" binding:"required"`
		SubmittedBy string `json:"submitted_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.reportingService.SubmitReport(c.Request.Context(), id, req.SubmittedTo, req.SubmittedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "submitted"})
}

func (h *Handler) getOverdueReports(c *gin.Context) {
	reports, err := h.reportingService.GetOverdueReports(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  reports,
		"count": len(reports),
	})
}

func (h *Handler) getReportStats(c *gin.Context) {
	stats, err := h.reportingService.GetReportStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Dashboard endpoints

func (h *Handler) getDashboardOverview(c *gin.Context) {
	metrics, err := h.complianceService.GetDashboardMetrics(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) getDashboardMetrics(c *gin.Context) {
	exchangeID := c.Query("exchange_id")

	metrics, err := h.complianceService.GetDashboardMetrics(c.Request.Context(), exchangeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) getViolationTrends(c *gin.Context) {
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	trends, err := h.complianceService.GetViolationTrends(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"start_date": startDate,
		"end_date":   endDate,
		"trends":     trends,
	})
}

func (h *Handler) getRecentActivity(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	activity, err := h.complianceService.GetRecentActivity(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  activity,
		"count": len(activity),
	})
}

// Compliance check endpoints

func (h *Handler) getComplianceChecks(c *gin.Context) {
	licenseID := c.Param("license_id")

	checks, err := h.complianceService.GetComplianceChecks(c.Request.Context(), licenseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"license_id": licenseID,
		"checks":     checks,
		"count":      len(checks),
	})
}

func (h *Handler) updateComplianceCheck(c *gin.Context) {
	var req models.ComplianceCheck

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.complianceService.UpdateComplianceCheck(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *Handler) createComplianceCheck(c *gin.Context) {
	var req models.ComplianceCheck

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.complianceService.CreateComplianceCheck(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "created"})
}

func (h *Handler) getKPIs(c *gin.Context) {
	exchangeID := c.Param("exchange_id")

	kpis, err := h.complianceService.GetKPIMetrics(c.Request.Context(), exchangeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exchange_id": exchangeID,
		"kpis":        kpis,
		"count":       len(kpis),
	})
}
