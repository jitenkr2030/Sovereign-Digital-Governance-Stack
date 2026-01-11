package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/csic/monitoring/internal/core/domain"
	"github.com/csic/monitoring/internal/core/ports"
	"github.com/csic/monitoring/internal/core/services"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// WalletHandler handles HTTP requests for wallet profiles
type WalletHandler struct {
	walletRepo   ports.WalletProfileRepository
	riskScorer   *services.RiskScoringService
	logger       *zap.Logger
}

// NewWalletHandler creates a new wallet handler
func NewWalletHandler(
	walletRepo ports.WalletProfileRepository,
	riskScorer *services.RiskScoringService,
	logger *zap.Logger,
) *WalletHandler {
	return &WalletHandler{
		walletRepo: walletRepo,
		riskScorer: riskScorer,
		logger:     logger,
	}
}

// GetWalletProfile handles GET /wallets/profile/{address}
func (h *WalletHandler) GetWalletProfile(w http.ResponseWriter, r *http.Request) {
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
	profile, err := h.walletRepo.GetByAddress(ctx, address, chain)
	if err != nil {
		h.logger.Error("Failed to get wallet profile", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "QUERY_ERROR", "Failed to get wallet profile", err.Error())
		return
	}

	if profile == nil {
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Wallet profile not found", "")
		return
	}

	h.respondJSON(w, http.StatusOK, profile)
}

// GetWalletRisk handles GET /wallets/risk/{address}
func (h *WalletHandler) GetWalletRisk(w http.ResponseWriter, r *http.Request) {
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
	profile, err := h.riskScorer.CalculateWalletRisk(ctx, address, chain)
	if err != nil {
		h.logger.Error("Failed to calculate wallet risk", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "CALCULATION_ERROR", "Failed to calculate wallet risk", err.Error())
		return
	}

	// Determine overall risk level
	riskLevel := "MINIMAL"
	hasCritical := false
	hasHigh := false

	for _, indicator := range profile.RiskIndicators {
		switch indicator.Severity {
		case "CRITICAL":
			hasCritical = true
		case "HIGH":
			hasHigh = true
		}
	}

	if hasCritical {
		riskLevel = "CRITICAL"
	} else if hasHigh {
		riskLevel = "HIGH"
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"address":          profile.Address,
		"chain":            profile.Chain,
		"risk_level":       riskLevel,
		"risk_indicators":  profile.RiskIndicators,
		"transaction_count": profile.TxCount,
		"total_volume_usd": profile.TotalVolumeUSD,
		"avg_tx_value_usd": profile.AvgTxValueUSD,
		"wallet_age_hours": profile.WalletAgeHours,
		"is_contract":      profile.IsContract,
	})
}

// SearchWallets handles GET /wallets/search
func (h *WalletHandler) SearchWallets(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if query == "" {
		h.respondError(w, http.StatusBadRequest, "MISSING_PARAMETER", "Search query is required", "")
		return
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	ctx := r.Context()
	wallets, total, err := h.walletRepo.Search(ctx, query, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to search wallets", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "SEARCH_ERROR", "Failed to search wallets", err.Error())
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"items":       wallets,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

func (h *WalletHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *WalletHandler) respondError(w http.ResponseWriter, status int, code, message, details string) {
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
