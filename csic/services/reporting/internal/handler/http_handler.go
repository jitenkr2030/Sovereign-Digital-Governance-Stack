package handler

import (
	"encoding/json"
	"net/http"

	"github.com/csic-platform/services/reporting/internal/core/ports"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// HTTPHandler handles HTTP requests for the reporting service
type HTTPHandler struct {
	reportService   ports.ReportService
	templateService ports.TemplateService
	schedulerService ports.SchedulerService
	logger          *zap.Logger
}

// NewHTTPHandler creates a new HTTPHandler
func NewHTTPHandler(
	reportService ports.ReportService,
	templateService ports.TemplateService,
	schedulerService ports.SchedulerService,
	logger *zap.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		reportService:    reportService,
		templateService:  templateService,
		schedulerService: schedulerService,
		logger:           logger,
	}
}

// RegisterRoutes registers all HTTP routes
func (h *HTTPHandler) RegisterRoutes(r *chi.Mux) {
	// Report routes
	r.Post("/api/v1/reports", h.GenerateReport)
	r.Get("/api/v1/reports", h.ListReports)
	r.Get("/api/v1/reports/{id}", h.GetReport)
	r.Get("/api/v1/reports/{id}/download", h.DownloadReport)
	r.Delete("/api/v1/reports/{id}", h.DeleteReport)
	r.Post("/api/v1/reports/{id}/archive", h.ArchiveReport)
	r.Get("/api/v1/reports/statistics", h.GetStatistics)

	// Template routes
	r.Post("/api/v1/templates", h.CreateTemplate)
	r.Get("/api/v1/templates", h.ListTemplates)
	r.Get("/api/v1/templates/{id}", h.GetTemplate)
	r.Put("/api/v1/templates/{id}", h.UpdateTemplate)
	r.Delete("/api/v1/templates/{id}", h.DeleteTemplate)

	// Scheduled report routes
	r.Post("/api/v1/scheduled", h.CreateScheduledReport)
	r.Get("/api/v1/scheduled", h.ListScheduledReports)
	r.Get("/api/v1/scheduled/{id}", h.GetScheduledReport)
	r.Put("/api/v1/scheduled/{id}", h.UpdateScheduledReport)
	r.Delete("/api/v1/scheduled/{id}", h.DeleteScheduledReport)
	r.Post("/api/v1/scheduled/{id}/enable", h.EnableScheduledReport)
	r.Post("/api/v1/scheduled/{id}/disable", h.DisableScheduledReport)
	r.Post("/api/v1/scheduled/{id}/execute", h.ExecuteScheduledReport)

	// Health check
	r.Get("/health", h.HealthCheck)
}

// Report handlers

func (h *HTTPHandler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	var req ports.GenerateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.reportService.GenerateReport(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to generate report", err)
		return
	}

	h.writeJSON(w, http.StatusAccepted, resp)
}

func (h *HTTPHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	var req ports.ListReportsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.Body != http.NoBody {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.reportService.ListReports(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list reports", err)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	report, err := h.reportService.GetReport(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Report not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, report)
}

func (h *HTTPHandler) DownloadReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	reader, err := h.reportService.DownloadReport(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Report not found or not ready", err)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Disposition", "attachment; filename=report.pdf")
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeContent(w, r, "report.pdf", nil, reader)
}

func (h *HTTPHandler) DeleteReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.reportService.DeleteReport(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to delete report", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Report deleted"})
}

func (h *HTTPHandler) ArchiveReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.reportService.ArchiveReport(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to archive report", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Report archived"})
}

func (h *HTTPHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.reportService.GetStatistics(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get statistics", err)
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// Template handlers

func (h *HTTPHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req ports.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.templateService.CreateTemplate(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to create template", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

func (h *HTTPHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	var req ports.ListTemplatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.Body != http.NoBody {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.templateService.ListTemplates(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list templates", err)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	template, err := h.templateService.GetTemplate(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Template not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, template)
}

func (h *HTTPHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req ports.UpdateTemplateRequest
	req.ID = id

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.templateService.UpdateTemplate(r.Context(), &req); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update template", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Template updated"})
}

func (h *HTTPHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.templateService.DeleteTemplate(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to delete template", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Template deleted"})
}

// Scheduled report handlers

func (h *HTTPHandler) CreateScheduledReport(w http.ResponseWriter, r *http.Request) {
	var req ports.CreateScheduledRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.schedulerService.CreateScheduledReport(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to create scheduled report", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

func (h *HTTPHandler) ListScheduledReports(w http.ResponseWriter, r *http.Request) {
	reports, err := h.schedulerService.ListScheduledReports(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list scheduled reports", err)
		return
	}

	h.writeJSON(w, http.StatusOK, reports)
}

func (h *HTTPHandler) GetScheduledReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	report, err := h.schedulerService.GetScheduledReport(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Scheduled report not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, report)
}

func (h *HTTPHandler) UpdateScheduledReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req ports.UpdateScheduledRequest
	req.ID = id

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.schedulerService.UpdateScheduledReport(r.Context(), &req); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update scheduled report", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Scheduled report updated"})
}

func (h *HTTPHandler) DeleteScheduledReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.schedulerService.DeleteScheduledReport(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to delete scheduled report", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Scheduled report deleted"})
}

func (h *HTTPHandler) EnableScheduledReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.schedulerService.EnableScheduledReport(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to enable scheduled report", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Scheduled report enabled"})
}

func (h *HTTPHandler) DisableScheduledReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.schedulerService.DisableScheduledReport(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to disable scheduled report", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Scheduled report disabled"})
}

func (h *HTTPHandler) ExecuteScheduledReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.schedulerService.ExecuteScheduledReport(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to execute scheduled report", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Scheduled report execution started"})
}

// Health check

func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "reporting",
	})
}

// Helper methods

func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *HTTPHandler) writeError(w http.ResponseWriter, status int, message string, err error) {
	h.logger.Error(message, zap.Error(err))
	h.writeJSON(w, status, map[string]interface{}{
		"error":   message,
		"details": err.Error(),
	})
}
