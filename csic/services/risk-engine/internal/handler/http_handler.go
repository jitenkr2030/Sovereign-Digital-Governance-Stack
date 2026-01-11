package handler

import (
	"encoding/json"
	"net/http"

	"github.com/csic-platform/services/risk-engine/internal/core/ports"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// HTTPHandler handles HTTP requests for the risk engine service
type HTTPHandler struct {
	ruleService    ports.RiskRuleService
	alertService   ports.AlertService
	evaluationSvc  ports.RiskEvaluationService
	thresholdSvc   ports.ThresholdService
	logger         *zap.Logger
}

// NewHTTPHandler creates a new HTTPHandler
func NewHTTPHandler(
	ruleService ports.RiskRuleService,
	alertService ports.AlertService,
	evaluationSvc ports.RiskEvaluationService,
	thresholdSvc ports.ThresholdService,
	logger *zap.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		ruleService:   ruleService,
		alertService:  alertService,
		evaluationSvc: evaluationSvc,
		thresholdSvc:  thresholdSvc,
		logger:        logger,
	}
}

// RegisterRoutes registers all HTTP routes
func (h *HTTPHandler) RegisterRoutes(r *chi.Mux) {
	// Risk rules routes
	r.Post("/api/v1/rules", h.CreateRule)
	r.Get("/api/v1/rules", h.ListRules)
	r.Get("/api/v1/rules/{id}", h.GetRule)
	r.Put("/api/v1/rules/{id}", h.UpdateRule)
	r.Delete("/api/v1/rules/{id}", h.DeleteRule)
	r.Post("/api/v1/rules/{id}/enable", h.EnableRule)
	r.Post("/api/v1/rules/{id}/disable", h.DisableRule)

	// Alert routes
	r.Get("/api/v1/alerts", h.ListAlerts)
	r.Get("/api/v1/alerts/{id}", h.GetAlert)
	r.Get("/api/v1/alerts/active", h.GetActiveAlerts)
	r.Post("/api/v1/alerts/{id}/acknowledge", h.AcknowledgeAlert)
	r.Post("/api/v1/alerts/{id}/resolve", h.ResolveAlert)
	r.Post("/api/v1/alerts/{id}/dismiss", h.DismissAlert)
	r.Get("/api/v1/alerts/stats", h.GetAlertStats)

	// Risk assessment routes
	r.Post("/api/v1/assessments", h.PerformAssessment)
	r.Get("/api/v1/assessments/{entity_id}", h.GetAssessment)

	// Threshold routes
	r.Get("/api/v1/thresholds", h.GetThresholds)
	r.Put("/api/v1/thresholds", h.UpdateThresholds)

	// Health check
	r.Get("/health", h.HealthCheck)
}

// Rule handlers

func (h *HTTPHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var req ports.CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.ruleService.CreateRule(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to create rule", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

func (h *HTTPHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	var req ports.ListRulesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.Body != http.NoBody {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.ruleService.ListRules(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list rules", err)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	rule, err := h.ruleService.GetRule(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Rule not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, rule)
}

func (h *HTTPHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req ports.UpdateRuleRequest
	req.ID = id

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.ruleService.UpdateRule(r.Context(), &req); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update rule", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Rule updated"})
}

func (h *HTTPHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.ruleService.DeleteRule(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to delete rule", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Rule deleted"})
}

func (h *HTTPHandler) EnableRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.ruleService.EnableRule(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to enable rule", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Rule enabled"})
}

func (h *HTTPHandler) DisableRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.ruleService.DisableRule(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to disable rule", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Rule disabled"})
}

// Alert handlers

func (h *HTTPHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	var req ports.ListAlertsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.Body != http.NoBody {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.alertService.ListAlerts(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list alerts", err)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) GetAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	alert, err := h.alertService.GetAlert(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Alert not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, alert)
}

func (h *HTTPHandler) GetActiveAlerts(w http.ResponseWriter, r *http.Request) {
	alerts, err := h.alertService.GetActiveAlerts(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get active alerts", err)
		return
	}

	h.writeJSON(w, http.StatusOK, alerts)
}

func (h *HTTPHandler) AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		UserID string `json:"user_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.alertService.AcknowledgeAlert(r.Context(), id, req.UserID); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to acknowledge alert", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Alert acknowledged"})
}

func (h *HTTPHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.alertService.ResolveAlert(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to resolve alert", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Alert resolved"})
}

func (h *HTTPHandler) DismissAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.alertService.DismissAlert(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to dismiss alert", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Alert dismissed"})
}

func (h *HTTPHandler) GetAlertStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.alertService.GetStats(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get stats", err)
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// Assessment handlers

func (h *HTTPHandler) PerformAssessment(w http.ResponseWriter, r *http.Request) {
	var req ports.RiskAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	assessment, err := h.evaluationSvc.AssessRisk(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to perform assessment", err)
		return
	}

	h.writeJSON(w, http.StatusOK, assessment)
}

func (h *HTTPHandler) GetAssessment(w http.ResponseWriter, r *http.Request) {
	entityID := chi.URLParam(r, "entity_id")

	assessment, err := h.evaluationSvc.AssessRisk(r.Context(), &ports.RiskAssessmentRequest{
		EntityID:   entityID,
		EntityType: "entity",
	})
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get assessment", err)
		return
	}

	h.writeJSON(w, http.StatusOK, assessment)
}

// Threshold handlers

func (h *HTTPHandler) GetThresholds(w http.ResponseWriter, r *http.Request) {
	config, err := h.thresholdSvc.GetConfig(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get thresholds", err)
		return
	}

	h.writeJSON(w, http.StatusOK, config)
}

func (h *HTTPHandler) UpdateThresholds(w http.ResponseWriter, r *http.Request) {
	var req ports.UpdateThresholdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.thresholdSvc.UpdateConfig(r.Context(), &req); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update thresholds", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Thresholds updated"})
}

// Health check

func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "risk-engine",
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
