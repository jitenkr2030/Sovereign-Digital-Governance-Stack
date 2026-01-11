package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"csic-platform/service/reporting/internal/domain"
	"csic-platform/service/reporting/internal/service"
)

// Handler handles HTTP requests for the reporting service
type Handler struct {
	reportService  *service.ReportGenerationService
	exportService  *service.ExportService
	schedulerService *service.SchedulerService
}

// NewHandler creates a new handler
func NewHandler(
	reportService *service.ReportGenerationService,
	exportService *service.ExportService,
	schedulerService *service.SchedulerService,
) *Handler {
	return &Handler{
		reportService:    reportService,
		exportService:    exportService,
		schedulerService: schedulerService,
	}
}

// SetupRoutes sets up the API routes
func (h *Handler) SetupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Reports
		reports := v1.Group("/reports")
		{
			reports.GET("", h.ListReports)
			reports.POST("", h.CreateReport)
			reports.GET("/:id", h.GetReport)
			reports.POST("/:id/regenerate", h.RegenerateReport)
			reports.POST("/:id/approve", h.ApproveReport)
			reports.POST("/:id/validate", h.ValidateReport)
			reports.POST("/:id/archive", h.ArchiveReport)
			reports.POST("/:id/submit", h.SubmitReport)
			reports.GET("/:id/download", h.DownloadReport)
		}

		// Templates
		templates := v1.Group("/templates")
		{
			templates.GET("", h.ListTemplates)
			templates.POST("", h.CreateTemplate)
			templates.GET("/:id", h.GetTemplate)
			templates.PUT("/:id", h.UpdateTemplate)
			templates.DELETE("/:id", h.DeleteTemplate)
		}

		// Schedules
		schedules := v1.Group("/schedules")
		{
			schedules.GET("", h.ListSchedules)
			schedules.POST("", h.CreateSchedule)
			schedules.GET("/:id", h.GetSchedule)
			schedules.PUT("/:id", h.UpdateSchedule)
			schedules.DELETE("/:id", h.DeleteSchedule)
			schedules.POST("/:id/trigger", h.TriggerSchedule)
			schedules.POST("/:id/activate", h.ActivateSchedule)
			schedules.POST("/:id/deactivate", h.DeactivateSchedule)
		}

		// Exports
		exports := v1.Group("/exports")
		{
			exports.GET("", h.ListExports)
			exports.GET("/:id", h.GetExport)
			exports.POST("/:id/download", h.DownloadExport)
		}

		// Statistics
		v1.GET("/stats/reports", h.GetReportStats)
		v1.GET("/stats/exports", h.GetExportStats)
		v1.GET("/stats/scheduler", h.GetSchedulerStats)
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "csic-reporting-service",
	})
}

// ReadinessCheck handles readiness check requests
func (h *Handler) ReadinessCheck(c *gin.Context) {
	// In a real implementation, check database and cache connectivity
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// ListReports handles GET /api/v1/reports
func (h *Handler) ListReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := domain.ReportFilter{
		Page:     page,
		PageSize: pageSize,
		SortBy:   c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}

	// Parse query parameters
	if regulatorID := c.Query("regulator_id"); regulatorID != "" {
		if id, err := uuid.Parse(regulatorID); err == nil {
			filter.RegulatorIDs = []uuid.UUID{id}
		}
	}

	if reportType := c.Query("type"); reportType != "" {
		filter.ReportTypes = []domain.ReportType{domain.ReportType(reportType)}
	}

	if status := c.Query("status"); status != "" {
		filter.Statuses = []domain.ReportStatus{domain.ReportStatus(status)}
	}

	reports, total, err := h.reportService.ListReports(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       reports,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": (total + pageSize - 1) / pageSize,
	})
}

// CreateReport handles POST /api/v1/reports
func (h *Handler) CreateReport(c *gin.Context) {
	var req service.GenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set generation method
	req.GenerationMethod = "manual"
	req.GeneratedBy = c.GetString("user_id")

	report, err := h.reportService.GenerateReport(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, report)
}

// GetReport handles GET /api/v1/reports/:id
func (h *Handler) GetReport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	report, err := h.reportService.GetReport(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if report == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// RegenerateReport handles POST /api/v1/reports/:id/regenerate
func (h *Handler) RegenerateReport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	userID := c.GetString("user_id")
	report, err := h.reportService.RegenerateReport(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// ApproveReport handles POST /api/v1/reports/:id/approve
func (h *Handler) ApproveReport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	userID := c.GetString("user_id")
	if err := h.reportService.ApproveReport(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "report approved"})
}

// ValidateReport handles POST /api/v1/reports/:id/validate
func (h *Handler) ValidateReport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	userID := c.GetString("user_id")
	if err := h.reportService.ValidateReport(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "report validated"})
}

// ArchiveReport handles POST /api/v1/reports/:id/archive
func (h *Handler) ArchiveReport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	if err := h.reportService.ArchiveReport(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "report archived"})
}

// SubmitReport handles POST /api/v1/reports/:id/submit
func (h *Handler) SubmitReport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	var req struct {
		SubmissionRef string `json:"submission_ref"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if err := h.reportService.SubmitReport(c.Request.Context(), id, userID, req.SubmissionRef); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "report submitted"})
}

// DownloadReport handles GET /api/v1/reports/:id/download
func (h *Handler) DownloadReport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	format := domain.ReportFormat(c.DefaultQuery("format", "pdf"))

	// Get client IP
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	sessionID := c.GetHeader("X-Session-ID")

	req := &service.ExportRequest{
		ReportID:   id,
		Format:     format,
		IPAddress:  clientIP,
		UserAgent:  userAgent,
		SessionID:  sessionID,
		Encryption: domain.EncryptionNone,
	}

	result, err := h.exportService.ExportReport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer result.FileReader.Close()

	// Record download
	h.exportService.RecordDownload(c.Request.Context(), result.ExportLog.ID)

	// Set headers
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+result.FileName)
	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Length", strconv.FormatInt(result.FileSize, 10))
	c.Header("X-Export-Log-ID", result.ExportLog.ID.String())
	c.Header("X-Checksum", result.ExportLog.Checksum)

	c.Data(http.StatusOK, result.ContentType, result.FileReader)
}

// ListTemplates handles GET /api/v1/templates
func (h *Handler) ListTemplates(c *gin.Context) {
	// Placeholder - would call template repository
	c.JSON(http.StatusOK, gin.H{"message": "templates list"})
}

// CreateTemplate handles POST /api/v1/templates
func (h *Handler) CreateTemplate(c *gin.Context) {
	// Placeholder
	c.JSON(http.StatusCreated, gin.H{"message": "template created"})
}

// GetTemplate handles GET /api/v1/templates/:id
func (h *Handler) GetTemplate(c *gin.Context) {
	// Placeholder
	c.JSON(http.StatusOK, gin.H{"message": "template details"})
}

// UpdateTemplate handles PUT /api/v1/templates/:id
func (h *Handler) UpdateTemplate(c *gin.Context) {
	// Placeholder
	c.JSON(http.StatusOK, gin.H{"message": "template updated"})
}

// DeleteTemplate handles DELETE /api/v1/templates/:id
func (h *Handler) DeleteTemplate(c *gin.Context) {
	// Placeholder
	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}

// ListSchedules handles GET /api/v1/schedules
func (h *Handler) ListSchedules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	schedules, total, err := h.schedulerService.ListSchedules(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       schedules,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": (total + pageSize - 1) / pageSize,
	})
}

// CreateSchedule handles POST /api/v1/schedules
func (h *Handler) CreateSchedule(c *gin.Context) {
	var schedule domain.ReportSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schedule.CreatedBy = c.GetString("user_id")
	if err := h.schedulerService.AddSchedule(c.Request.Context(), &schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

// GetSchedule handles GET /api/v1/schedules/:id
func (h *Handler) GetSchedule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule ID"})
		return
	}

	schedule, err := h.schedulerService.GetSchedule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if schedule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// UpdateSchedule handles PUT /api/v1/schedules/:id
func (h *Handler) UpdateSchedule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule ID"})
		return
	}

	var schedule domain.ReportSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schedule.ID = id
	if err := h.schedulerService.UpdateSchedule(c.Request.Context(), &schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// DeleteSchedule handles DELETE /api/v1/schedules/:id
func (h *Handler) DeleteSchedule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule ID"})
		return
	}

	if err := h.schedulerService.RemoveSchedule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "schedule deleted"})
}

// TriggerSchedule handles POST /api/v1/schedules/:id/trigger
func (h *Handler) TriggerSchedule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule ID"})
		return
	}

	if err := h.schedulerService.TriggerSchedule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "schedule triggered"})
}

// ActivateSchedule handles POST /api/v1/schedules/:id/activate
func (h *Handler) ActivateSchedule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule ID"})
		return
	}

	if err := h.schedulerService.ActivateSchedule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "schedule activated"})
}

// DeactivateSchedule handles POST /api/v1/schedules/:id/deactivate
func (h *Handler) DeactivateSchedule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule ID"})
		return
	}

	if err := h.schedulerService.DeactivateSchedule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "schedule deactivated"})
}

// ListExports handles GET /api/v1/exports
func (h *Handler) ListExports(c *gin.Context) {
	userID := c.GetString("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	exports, total, err := h.exportService.GetExportLogsByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       exports,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": (total + pageSize - 1) / pageSize,
	})
}

// GetExport handles GET /api/v1/exports/:id
func (h *Handler) GetExport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid export ID"})
		return
	}

	filter := domain.ExportFilter{
		SearchQuery: id.String(),
		Page:        1,
		PageSize:    1,
	}

	exports, _, err := h.exportService.GetExportLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(exports) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "export not found"})
		return
	}

	c.JSON(http.StatusOK, exports[0])
}

// DownloadExport handles POST /api/v1/exports/:id/download
func (h *Handler) DownloadExport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid export ID"})
		return
	}

	format := domain.ReportFormat(c.DefaultQuery("format", "pdf"))

	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	req := &service.ExportRequest{
		ReportID:  id,
		Format:    format,
		IPAddress: clientIP,
		UserAgent: userAgent,
	}

	result, err := h.exportService.ExportReport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer result.FileReader.Close()

	h.exportService.RecordDownload(c.Request.Context(), result.ExportLog.ID)

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+result.FileName)
	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Length", strconv.FormatInt(result.FileSize, 10))

	c.Data(http.StatusOK, result.ContentType, result.FileReader)
}

// GetReportStats handles GET /api/v1/stats/reports
func (h *Handler) GetReportStats(c *gin.Context) {
	// Placeholder
	c.JSON(http.StatusOK, gin.H{
		"total_reports":   0,
		"completed":       0,
		"pending":         0,
		"failed":          0,
	})
}

// GetExportStats handles GET /api/v1/stats/exports
func (h *Handler) GetExportStats(c *gin.Context) {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			start = time.Now().AddDate(0, -1, 0)
		}
	}

	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			end = time.Now()
		}
	}

	stats, err := h.exportService.GetExportStats(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetSchedulerStats handles GET /api/v1/stats/scheduler
func (h *Handler) GetSchedulerStats(c *gin.Context) {
	stats, err := h.schedulerService.GetSchedulerStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CORSMiddleware handles CORS
func (h *Handler) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggerMiddleware handles logging
func (h *Handler) LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health", "/ready"},
	})
}

// RecoveryMiddleware handles panics
func (h *Handler) RecoveryMiddleware() gin.HandlerFunc {
	return gin.Recovery()
}

// AuthenticationMiddleware handles authentication
func (h *Handler) AuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real implementation, validate JWT token
		token := c.GetHeader("Authorization")
		if token != "" && strings.HasPrefix(token, "Bearer ") {
			// Extract user info from token
			c.Set("user_id", "user-123")
			c.Set("user_email", "user@example.com")
			c.Set("user_role", "admin")
		}
		c.Next()
	}
}
