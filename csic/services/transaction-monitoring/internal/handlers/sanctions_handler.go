package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/csic/monitoring/internal/core/domain"
	"github.com/csic/monitoring/internal/core/services"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// SanctionsHandler handles HTTP requests for sanctions management
type SanctionsHandler struct {
	service *services.SanctionsService
	logger  *zap.Logger
}

// NewSanctionsHandler creates a new sanctions handler
func NewSanctionsHandler(service *services.SanctionsService, logger *zap.Logger) *SanctionsHandler {
	return &SanctionsHandler{
		service: service,
		logger:  logger,
	}
}

// ListSanctions handles GET /sanctions
func (h *SanctionsHandler) ListSanctions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	ctx := r.Context()
	sanctions, total, err := h.service.ListSanctions(ctx, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list sanctions", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "QUERY_ERROR", "Failed to query sanctions", err.Error())
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"items":      sanctions,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": totalPages,
	})
}

// AddSanctionedAddress handles POST /sanctions/add
func (h *SanctionsHandler) AddSanctionedAddress(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Address    string `json:"address"`
		Chain      string `json:"chain"`
		SourceList string `json:"source_list"`
		Reason     string `json:"reason"`
		EntityName string `json:"entity_name"`
		EntityType string `json:"entity_type"`
		Program    string `json:"program"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}

	// Validate required fields
	if req.Address == "" {
		h.respondError(w, http.StatusBadRequest, "MISSING_FIELD", "Address is required", "")
		return
	}
	if req.Chain == "" {
		h.respondError(w, http.StatusBadRequest, "MISSING_FIELD", "Chain is required", "")
		return
	}
	if req.SourceList == "" {
		h.respondError(w, http.StatusBadRequest, "MISSING_FIELD", "Source list is required", "")
		return
	}

	sanction := &domain.SanctionedAddress{
		Address:    req.Address,
		Chain:      req.Chain,
		SourceList: req.SourceList,
		Reason:     req.Reason,
		EntityName: req.EntityName,
		EntityType: req.EntityType,
		Program:    req.Program,
		AddedAt:    time.Now().UTC(),
	}

	ctx := r.Context()
	if err := h.service.AddSanction(ctx, sanction); err != nil {
		h.logger.Error("Failed to add sanction", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INSERT_ERROR", "Failed to add sanction", err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          sanction.ID,
		"address":     sanction.Address,
		"chain":       sanction.Chain,
		"source_list": sanction.SourceList,
		"added_at":    sanction.AddedAt.Format(time.RFC3339),
	})
}

// CheckAddress handles GET /sanctions/check/{address}
func (h *SanctionsHandler) CheckAddress(w http.ResponseWriter, r *http.Request) {
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
	sanctions, err := h.service.CheckAddress(ctx, address, chain)
	if err != nil {
		h.logger.Error("Failed to check address", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "CHECK_ERROR", "Failed to check address", err.Error())
		return
	}

	isSanctioned := len(sanctions) > 0

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"address":       address,
		"chain":         chain,
		"is_sanctioned": isSanctioned,
		"sanctions":     sanctions,
	})
}

// ImportSanctionsList handles POST /sanctions/import
func (h *SanctionsHandler) ImportSanctionsList(w http.ResponseWriter, r *http.Request) {
	var req domain.SanctionsListImport
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}

	if len(req.Addresses) == 0 {
		h.respondError(w, http.StatusBadRequest, "EMPTY_LIST", "No addresses to import", "")
		return
	}

	ctx := r.Context()
	imported, failed, err := h.service.ImportSanctionsList(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to import sanctions list", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "IMPORT_ERROR", "Failed to import sanctions", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"imported_count": imported,
		"failed_count":   failed,
		"source_list":    req.SourceList,
		"imported_by":    req.ImportedBy,
	})
}

func (h *SanctionsHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *SanctionsHandler) respondError(w http.ResponseWriter, status int, code, message, details string) {
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
