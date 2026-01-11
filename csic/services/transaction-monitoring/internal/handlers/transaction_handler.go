package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/csic/monitoring/internal/core/domain"
	"github.com/csic/monitoring/internal/core/services"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// TransactionHandler handles HTTP requests for transactions
type TransactionHandler struct {
	service *services.TransactionService
	logger  *zap.Logger
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(service *services.TransactionService, logger *zap.Logger) *TransactionHandler {
	return &TransactionHandler{
		service: service,
		logger:  logger,
	}
}

// IngestTransaction handles POST /transactions/ingest
func (h *TransactionHandler) IngestTransaction(w http.ResponseWriter, r *http.Request) {
	var tx domain.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}

	// Validate required fields
	if tx.TxHash == "" {
		h.respondError(w, http.StatusBadRequest, "MISSING_FIELD", "Transaction hash is required", "")
		return
	}
	if tx.Chain == "" {
		h.respondError(w, http.StatusBadRequest, "MISSING_FIELD", "Chain is required", "")
		return
	}
	if tx.FromAddress == "" {
		h.respondError(w, http.StatusBadRequest, "MISSING_FIELD", "From address is required", "")
		return
	}
	if tx.Amount <= 0 {
		h.respondError(w, http.StatusBadRequest, "INVALID_AMOUNT", "Amount must be greater than 0", "")
		return
	}

	ctx := r.Context()
	result, err := h.service.IngestTransaction(ctx, &tx)
	if err != nil {
		h.logger.Error("Failed to ingest transaction", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to process transaction", err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"transaction_id": result.ID,
		"tx_hash":        result.TxHash,
		"chain":          result.Chain,
		"risk_score":     result.RiskScore,
		"risk_level":     h.calculateRiskLevel(result.RiskScore),
		"flagged":        result.Flagged,
		"created_at":     result.CreatedAt.Format(time.RFC3339),
	})
}

// GetTransactionHistory handles GET /transactions/history
func (h *TransactionHandler) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	filter := h.parseTransactionFilter(r)

	ctx := r.Context()
	result, err := h.service.GetTransactionHistory(ctx, filter)
	if err != nil {
		h.logger.Error("Failed to get transaction history", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "QUERY_ERROR", "Failed to query transactions", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// GetTransactionRisk handles GET /transactions/risk/{txHash}
func (h *TransactionHandler) GetTransactionRisk(w http.ResponseWriter, r *http.Request) {
	txHash := mux.Vars(r)["txHash"]
	if txHash == "" {
		h.respondError(w, http.StatusBadRequest, "MISSING_PARAMETER", "Transaction hash is required", "")
		return
	}

	ctx := r.Context()
	risk, err := h.service.GetTransactionRisk(ctx, txHash)
	if err != nil {
		h.logger.Error("Failed to get transaction risk", zap.Error(err))
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Transaction not found", "")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"tx_hash":        txHash,
		"risk_score":     risk.OverallScore,
		"risk_level":     risk.RiskLevel,
		"requires_review": risk.RequiresReview,
		"flagged":        risk.Flagged,
		"factors":        risk.Factors,
		"recommendations": risk.Recommendations,
	})
}

// GetFlaggedTransactions handles GET /transactions/flagged
func (h *TransactionHandler) GetFlaggedTransactions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	ctx := r.Context()
	result, err := h.service.GetFlaggedTransactions(ctx, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get flagged transactions", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "QUERY_ERROR", "Failed to query flagged transactions", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// ScanAddress handles GET /transactions/scan/{address}
func (h *TransactionHandler) ScanAddress(w http.ResponseWriter, r *http.Request) {
	address := mux.Vars(r)["address"]
	chain := r.URL.Query().Get("chain")

	if address == "" {
		h.respondError(w, http.StatusBadRequest, "MISSING_PARAMETER", "Address is required", "")
		return
	}
	if chain == "" {
		h.respondError(w, http.StatusBadRequest, "MISSING_PARAMETER", "Chain is required", "")
		return
	}

	ctx := r.Context()
	profile, err := h.service.ScanAddress(ctx, address, chain)
	if err != nil {
		h.logger.Error("Failed to scan address", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "SCAN_ERROR", "Failed to scan address", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"address":           profile.Address,
		"chain":             profile.Chain,
		"tx_count":          profile.TxCount,
		"total_volume_usd":  profile.TotalVolumeUSD,
		"avg_tx_value_usd":  profile.AvgTxValueUSD,
		"risk_indicators":   profile.RiskIndicators,
		"first_seen":        formatTime(profile.FirstSeen),
		"last_seen":         formatTime(profile.LastSeen),
	})
}

// GetSuspiciousActivityReport handles GET /reports/suspicious-activity
func (h *TransactionHandler) GetSuspiciousActivityReport(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start_time")
	endStr := r.URL.Query().Get("end_time")

	startTime, err := parseTime(startStr, time.Now().AddDate(0, -1, 0))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_TIME", "Invalid start_time format", "")
		return
	}

	endTime, err := parseTime(endStr, time.Now())
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_TIME", "Invalid end_time format", "")
		return
	}

	ctx := r.Context()
	report, err := h.service.GetSuspiciousActivityReport(ctx, startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to generate suspicious activity report", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "REPORT_ERROR", "Failed to generate report", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, report)
}

// GetRiskSummaryReport handles GET /reports/risk-summary
func (h *TransactionHandler) GetRiskSummaryReport(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start_time")
	endStr := r.URL.Query().Get("end_time")

	startTime, err := parseTime(startStr, time.Now().AddDate(0, -1, 0))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_TIME", "Invalid start_time format", "")
		return
	}

	endTime, err := parseTime(endStr, time.Now())
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_TIME", "Invalid end_time format", "")
		return
	}

	ctx := r.Context()
	report, err := h.service.GetRiskSummaryReport(ctx, startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to generate risk summary report", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "REPORT_ERROR", "Failed to generate report", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, report)
}

// Helper functions

func (h *TransactionHandler) parseTransactionFilter(r *http.Request) *domain.TransactionFilter {
	filter := &domain.TransactionFilter{
		Page:     1,
		PageSize: 20,
	}

	if chain := r.URL.Query().Get("chain"); chain != "" {
		filter.Chain = chain
	}
	if from := r.URL.Query().Get("from_address"); from != "" {
		filter.FromAddress = from
	}
	if to := r.URL.Query().Get("to_address"); to != "" {
		filter.ToAddress = to
	}
	if minAmt := r.URL.Query().Get("min_amount_usd"); minAmt != "" {
		if val, err := strconv.ParseFloat(minAmt, 64); err == nil {
			filter.MinAmountUSD = val
		}
	}
	if maxAmt := r.URL.Query().Get("max_amount_usd"); maxAmt != "" {
		if val, err := strconv.ParseFloat(maxAmt, 64); err == nil {
			filter.MaxAmountUSD = val
		}
	}
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil {
			filter.Page = val
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if val, err := strconv.Atoi(pageSizeStr); err == nil {
			filter.PageSize = val
		}
	}

	return filter
}

func (h *TransactionHandler) calculateRiskLevel(score int) string {
	switch {
	case score >= 80:
		return "CRITICAL"
	case score >= 60:
		return "HIGH"
	case score >= 40:
		return "MEDIUM"
	case score >= 20:
		return "LOW"
	default:
		return "MINIMAL"
	}
}

func (h *TransactionHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *TransactionHandler) respondError(w http.ResponseWriter, status int, code, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	if details != "" {
		resp.(map[string]interface{})["error"].(map[string]interface{})["details"] = details
	}

	json.NewEncoder(w).Encode(resp)
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func parseTime(s string, defaultVal time.Time) (time.Time, error) {
	if s == "" {
		return defaultVal, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return defaultVal, err
	}
	return t, nil
}
