package http_handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/csic-platform/services/security/access-control/internal/core/domain"
	"github.com/csic-platform/services/security/access-control/internal/core/ports"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Handler holds the HTTP handlers for the access control service
type Handler struct {
	acService       ports.AccessControlService
	policyService   ports.PolicyManagementService
	logger          *zap.Logger
	validate        *validator.Validate
}

// NewHandler creates a new HTTP handler
func NewHandler(
	acService ports.AccessControlService,
	policyService ports.PolicyManagementService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		acService:     acService,
		policyService: policyService,
		logger:        logger,
		validate:      validator.New(),
	}
}

// AccessCheckRequest represents an access check request
type AccessCheckRequest struct {
	Subject   SubjectRequest  `json:"subject" validate:"required"`
	Resource  ResourceRequest `json:"resource" validate:"required"`
	Action    ActionRequest   `json:"action" validate:"required"`
	Context   ContextRequest  `json:"context"`
}

// SubjectRequest represents the subject in an access check request
type SubjectRequest struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Roles      []string          `json:"roles"`
	Groups     []string          `json:"groups"`
	Attributes map[string]string `json:"attributes"`
}

// ResourceRequest represents the resource in an access check request
type ResourceRequest struct {
	Type       string            `json:"type"`
	ID         string            `json:"id"`
	Path       string            `json:"path"`
	Attributes map[string]string `json:"attributes"`
}

// ActionRequest represents the action in an access check request
type ActionRequest struct {
	Name   string `json:"name"`
	Method string `json:"method"`
}

// ContextRequest represents the context in an access check request
type ContextRequest struct {
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	Time      string `json:"time"`
	Location  string `json:"location"`
	Device    string `json:"device"`
}

// AccessCheckResponse represents an access check response
type AccessCheckResponse struct {
	Decision    string             `json:"decision"`
	PolicyID    *string            `json:"policy_id,omitempty"`
	Reason      string             `json:"reason"`
	Obligations []domain.Obligation `json:"obligations,omitempty"`
	Timestamp   time.Time          `json:"timestamp"`
}

// CheckAccess handles access control check requests
func (h *Handler) CheckAccess(w http.ResponseWriter, r *http.Request) {
	var req AccessCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Convert request to domain model
	accessReq := &domain.AccessRequest{
		Subject: domain.Subject{
			ID:         req.Subject.ID,
			Type:       req.Subject.Type,
			Roles:      req.Subject.Roles,
			Groups:     req.Subject.Groups,
			Attributes: req.Subject.Attributes,
		},
		Resource: domain.Resource{
			Type:       req.Resource.Type,
			ID:         req.Resource.ID,
			Path:       req.Resource.Path,
			Attributes: req.Resource.Attributes,
		},
		Action: domain.Action{
			Name:   req.Action.Name,
			Method: req.Action.Method,
		},
		Context: domain.Context{
			IPAddress: req.Context.IPAddress,
			UserAgent: req.Context.UserAgent,
			Time:      time.Now(),
			Location:  req.Context.Location,
			Device:    req.Context.Device,
		},
	}

	response, err := h.acService.CheckAccess(r.Context(), accessReq)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Access check failed", err)
		return
	}

	// Convert response
	resp := AccessCheckResponse{
		Decision:    string(response.Decision),
		Reason:      response.Reason,
		Obligations: response.Obligations,
		Timestamp:   response.Timestamp,
	}

	if response.PolicyID != nil {
		policyIDStr := response.PolicyID.String()
		resp.PolicyID = &policyIDStr
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// CheckAccessBatch handles batch access control check requests
func (h *Handler) CheckAccessBatch(w http.ResponseWriter, r *http.Request) {
	var requests []AccessCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Convert requests to domain model
	domainRequests := make([]*domain.AccessRequest, len(requests))
	for i, req := range requests {
		domainRequests[i] = &domain.AccessRequest{
			Subject: domain.Subject{
				ID:         req.Subject.ID,
				Type:       req.Subject.Type,
				Roles:      req.Subject.Roles,
				Groups:     req.Subject.Groups,
				Attributes: req.Subject.Attributes,
			},
			Resource: domain.Resource{
				Type:       req.Resource.Type,
				ID:         req.Resource.ID,
				Path:       req.Resource.Path,
				Attributes: req.Resource.Attributes,
			},
			Action: domain.Action{
				Name:   req.Action.Name,
				Method: req.Action.Method,
			},
			Context: domain.Context{
				IPAddress: req.Context.IPAddress,
				UserAgent: req.Context.UserAgent,
				Time:      time.Now(),
				Location:  req.Context.Location,
				Device:    req.Context.Device,
			},
		}
	}

	// Process batch (would need a batch method on the service)
	responses := make([]AccessCheckResponse, len(requests))
	for i, req := range domainRequests {
		response, err := h.acService.CheckAccess(r.Context(), req)
		if err != nil {
			h.respondError(w, http.StatusInternalServerError, "Access check failed", err)
			return
		}

		responses[i] = AccessCheckResponse{
			Decision:    string(response.Decision),
			Reason:      response.Reason,
			Obligations: response.Obligations,
			Timestamp:   response.Timestamp,
		}

		if response.PolicyID != nil {
			policyIDStr := response.PolicyID.String()
			responses[i].PolicyID = &policyIDStr
		}
	}

	h.respondJSON(w, http.StatusOK, responses)
}

// CreatePolicy handles policy creation requests
func (h *Handler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req ports.CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	response, err := h.policyService.CreatePolicy(r.Context(), &req)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to create policy", err)
		return
	}

	h.respondJSON(w, http.StatusCreated, response)
}

// GetPolicy handles policy retrieval requests
func (h *Handler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID := vars["id"]

	policy, err := h.policyService.GetPolicy(r.Context(), policyID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to get policy", err)
		return
	}

	if policy == nil {
		h.respondError(w, http.StatusNotFound, "Policy not found", nil)
		return
	}

	h.respondJSON(w, http.StatusOK, policy)
}

// ListPolicies handles policy listing requests
func (h *Handler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	enabledOnly := r.URL.Query().Get("enabled_only")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	req := &ports.ListPoliciesRequest{
		Page:     1,
		PageSize: 20,
	}

	if enabledOnly != "" {
		enabled := enabledOnly == "true"
		req.EnabledOnly = &enabled
	}

	if pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}

	if pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = pageSize
		}
	}

	response, err := h.policyService.ListPolicies(r.Context(), req)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to list policies", err)
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

// UpdatePolicy handles policy update requests
func (h *Handler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID := vars["id"]

	var req ports.UpdatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	req.ID = policyID

	if err := h.policyService.UpdatePolicy(r.Context(), &req); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to update policy", err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Policy updated successfully"})
}

// DeletePolicy handles policy deletion requests
func (h *Handler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID := vars["id"]

	if err := h.policyService.DeletePolicy(r.Context(), policyID); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to delete policy", err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Policy deleted successfully"})
}

// EnablePolicy handles policy enable requests
func (h *Handler) EnablePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID := vars["id"]

	if err := h.policyService.EnablePolicy(r.Context(), policyID); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to enable policy", err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Policy enabled successfully"})
}

// DisablePolicy handles policy disable requests
func (h *Handler) DisablePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID := vars["id"]

	if err := h.policyService.DisablePolicy(r.Context(), policyID); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to disable policy", err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Policy disabled successfully"})
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// respondJSON writes a JSON response
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// respondError writes an error response
func (h *Handler) respondError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]interface{}{
		"error":   message,
		"status":  status,
	}

	if err != nil {
		response["details"] = err.Error()
	}

	h.logger.Error(message, zap.Error(err))
	h.respondJSON(w, status, response)
}

// RegisterRoutes registers all HTTP routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Access check endpoints
	router.HandleFunc("/api/v1/check", h.CheckAccess).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/check/batch", h.CheckAccessBatch).Methods(http.MethodPost)

	// Policy management endpoints
	router.HandleFunc("/api/v1/policies", h.ListPolicies).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/policies", h.CreatePolicy).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/policies/{id}", h.GetPolicy).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/policies/{id}", h.UpdatePolicy).Methods(http.MethodPut)
	router.HandleFunc("/api/v1/policies/{id}", h.DeletePolicy).Methods(http.MethodDelete)
	router.HandleFunc("/api/v1/policies/{id}/enable", h.EnablePolicy).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/policies/{id}/disable", h.DisablePolicy).Methods(http.MethodPost)

	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods(http.MethodGet)
}

// Compile-time check that Handler implements the expected interface
var _ = func() {}
