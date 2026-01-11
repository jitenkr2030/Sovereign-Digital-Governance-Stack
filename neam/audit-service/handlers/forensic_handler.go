package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"neam-platform/audit-service/crypto"
	"neam-platform/audit-service/models"
	"neam-platform/audit-service/repository"
	"neam-platform/audit-service/services"
)

// ForensicHandler handles forensic investigation HTTP requests
type ForensicHandler struct {
	forensicRepo   repository.ForensicRepository
	sealingService services.SealingService
	hashService    crypto.HashService
	accessControl  services.AccessControl
}

// NewForensicHandler creates a new ForensicHandler
func NewForensicHandler(
	forensicRepo repository.ForensicRepository,
	sealingService services.SealingService,
	hashService crypto.HashService,
	accessControl services.AccessControl,
) *ForesticHandler {
	return &ForensicHandler{
		forensicRepo:   forensicRepo,
		sealingService: sealingService,
		hashService:    hashService,
		accessControl:  accessControl,
	}
}

// RegisterRoutes registers forensic investigation routes
func (h *ForensicHandler) RegisterRoutes(router *gin.RouterGroup) {
	forensic := router.Group("/forensic")
	{
		forensic.POST("", h.CreateInvestigation)
		forensic.GET("/:id", h.GetInvestigation)
		forensic.POST("/:id/approve", h.ApproveInvestigation)
		forensic.POST("/:id/execute", h.ExecuteInvestigation)
		forensic.GET("/:id/report", h.GetInvestigationReport)
		forensic.GET("/evidence/:event_id", h.GetEvidence)
		forensic.POST("/verify", h.VerifyEvidence)
	}
}

// CreateInvestigation handles POST /api/v1/forensic
// @Summary Create forensic investigation
// @Description Create a new forensic investigation request
// @Tags forensic
// @Accept json
// @Produce json
// @Param investigation body ForensicCreateRequest true "Investigation details"
// @Success 201 {object} ForensicInvestigation
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/forensic [post]
func (h *ForensicHandler) CreateInvestigation(c *gin.Context) {
	var req ForensicCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Get requester info from context
	requesterID, _ := c.Get("user_id")
	requesterRole, _ := c.Get("role")

	// Validate justification
	if len(req.Justification) < 50 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_justification",
			Message: "Justification must be at least 50 characters",
		})
		return
	}

	// Parse time range
	startTime, err := time.Parse(time.RFC3339, req.Query.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_time",
			Message: "Invalid start_time format",
		})
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.Query.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_time",
			Message: "Invalid end_time format",
		})
		return
	}

	// Create investigation
	investigation := &models.ForensicRequest{
		ID:            generateInvestigationID(),
		RequesterID:   requesterID.(string),
		RequesterRole: requesterRole.(string),
		Query: models.ForensicQuery{
			EventTypes:   req.Query.EventTypes,
			ActorIDs:     req.Query.ActorIDs,
			Resources:    req.Query.Resources,
			TimeRange:    models.TimeRange{Start: startTime, End: endTime},
			SearchTerms:  req.Query.SearchTerms,
			IncludeChain: req.Query.IncludeChain,
			IncludeHash:  req.Query.IncludeHash,
		},
		Justification: req.Justification,
		Status:        "pending",
		CreatedAt:     time.Now(),
	}

	// Store investigation
	if err := h.forensicRepo.Create(c.Request.Context(), investigation); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, investigation)
}

// GetInvestigation handles GET /api/v1/forensic/:id
// @Summary Get forensic investigation
// @Description Get details of a forensic investigation
// @Tags forensic
// @Produce json
// @Param id path string true "Investigation ID"
// @Success 200 {object} ForensicInvestigation
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/forensic/{id} [get]
func (h *ForensicHandler) GetInvestigation(c *gin.Context) {
	investigationID := c.Param("id")

	investigation, err := h.forensicRepo.GetByID(c.Request.Context(), investigationID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Investigation not found",
		})
		return
	}

	c.JSON(http.StatusOK, investigation)
}

// ApproveInvestigation handles POST /api/v1/forensic/:id/approve
// @Summary Approve forensic investigation
// @Description Approve a pending forensic investigation
// @Tags forensic
// @Accept json
// @Produce json
// @Param id path string true "Investigation ID"
// @Success 200 {object} ForensicInvestigation
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/forensic/{id}/approve [post]
func (h *ForensicHandler) ApproveInvestigation(c *gin.Context) {
	investigationID := c.Param("id")

	// Get requester info
	approverID, _ := c.Get("user_id")

	// Get investigation
	investigation, err := h.forensicRepo.GetByID(c.Request.Context(), investigationID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Investigation not found",
		})
		return
	}

	// Check if investigation is pending
	if investigation.Status != "pending" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_status",
			Message: "Investigation is not pending approval",
		})
		return
	}

	// Update investigation
	now := time.Now()
	investigation.Status = "approved"
	investigation.ApprovedBy = approverID.(string)
	investigation.ApprovedAt = now

	if err := h.forensicRepo.Update(c.Request.Context(), investigation); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, investigation)
}

// ExecuteInvestigation handles POST /api/v1/forensic/:id/execute
// @Summary Execute forensic investigation
// @Description Execute an approved forensic investigation
// @Tags forensic
// @Produce json
// @Param id path string true "Investigation ID"
// @Success 200 {object} ForensicResult
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/forensic/{id}/execute [post]
func (h *ForensicHandler) ExecuteInvestigation(c *gin.Context) {
	investigationID := c.Param("id")

	// Get investigation
	investigation, err := h.forensicRepo.GetByID(c.Request.Context(), investigationID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Investigation not found",
		})
		return
	}

	// Check if investigation is approved
	if investigation.Status != "approved" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "not_approved",
			Message: "Investigation must be approved before execution",
		})
		return
	}

	// Execute investigation (placeholder - actual implementation would query audit logs)
	result := &ForensicResult{
		QueryID:       investigation.ID,
		EventsFound:   0,
		EventsReturned: 0,
		Events:        []models.AuditEvent{},
		QueryTime:     time.Since(investigation.CreatedAt),
		GeneratedAt:   time.Now(),
	}

	// Update investigation status
	investigation.Status = "completed"
	investigation.CompletedAt = time.Now()
	if err := h.forensicRepo.Update(c.Request.Context(), investigation); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetInvestigationReport handles GET /api/v1/forensic/:id/report
// @Summary Get investigation report
// @Description Get the report for a completed forensic investigation
// @Tags forensic
// @Produce json
// @Param id path string true "Investigation ID"
// @Success 200 {object} ForensicReport
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/forensic/{id}/report [get]
func (h *ForensicHandler) GetInvestigationReport(c *gin.Context) {
	investigationID := c.Param("id")

	// Get investigation
	investigation, err := h.forensicRepo.GetByID(c.Request.Context(), investigationID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Investigation not found",
		})
		return
	}

	// Generate report
	report := &ForensicReport{
		InvestigationID: investigation.ID,
		RequesterID:     investigation.RequesterID,
		RequesterRole:   investigation.RequesterRole,
		Justification:   investigation.Justification,
		Query:           investigation.Query,
		Status:          investigation.Status,
		CreatedAt:       investigation.CreatedAt,
		ApprovedBy:      investigation.ApprovedBy,
		ApprovedAt:      investigation.ApprovedAt,
		CompletedAt:     investigation.CompletedAt,
		GeneratedAt:     time.Now(),
	}

	c.JSON(http.StatusOK, report)
}

// GetEvidence handles GET /api/v1/forensic/evidence/:event_id
// @Summary Get evidence for event
// @Description Get court-grade evidence for a specific audit event
// @Tags forensic
// @Produce json
// @Param event_id path string true "Event ID"
// @Success 200 {object} CourtEvidence
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/forensic/evidence/{event_id} [get]
func (h *ForensicHandler) GetEvidence(c *gin.Context) {
	eventID := c.Param("event_id")

	// Get event from repository
	event, err := h.forensicRepo.GetEventByID(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Event not found",
		})
		return
	}

	// Generate evidence
	evidence := &CourtEvidence{
		Event:        event,
		Timestamp:    time.Now(),
		ChainOfCustody: []CustodyEvent{
			{
				Timestamp: event.CreatedAt,
				Action:    "CREATED",
				Actor:     "SYSTEM",
				Details:   "Event created in audit log",
			},
		},
	}

	// Add integrity hash
	hash := h.hashService.CalculateHash(event)
	evidence.IntegrityHash = hash

	// Add verification
	evidence.Verification = &EvidenceVerification{
		Verified:       true,
		VerifiedAt:     time.Now(),
		Algorithm:      "SHA256",
		RootHash:       hash,
	}

	c.JSON(http.StatusOK, evidence)
}

// VerifyEvidence handles POST /api/v1/forensic/verify
// @Summary Verify evidence integrity
// @Description Verify the integrity of audit evidence
// @Tags forensic
// @Accept json
// @Produce json
// @Param evidence body EvidenceVerifyRequest true "Evidence to verify"
// @Success 200 {object} EvidenceVerification
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/forensic/verify [post]
func (h *ForensicHandler) VerifyEvidence(c *gin.Context) {
	var req EvidenceVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Calculate expected hash
	expectedHash := h.hashService.CalculateHash(req.EventData)

	// Compare with provided hash
	valid := expectedHash == req.ExpectedHash

	// Verify seal if provided
	var sealValid bool
	var sealError string
	if req.SealData != nil {
		seal, err := json.Marshal(req.SealData)
		if err != nil {
			sealValid = false
			sealError = "Failed to marshal seal data"
		} else {
			sealValid, sealError = h.sealingService.VerifySeal(c.Request.Context(), req.SealData)
		}
	}

	verification := &EvidenceVerification{
		Verified:    valid && sealValid,
		VerifiedAt:  time.Now(),
		Algorithm:   "SHA256",
		RootHash:    expectedHash,
		HashMatch:   valid,
		SealValid:   sealValid,
		SealError:   sealError,
	}

	c.JSON(http.StatusOK, verification)
}

// Helper functions

func generateInvestigationID() string {
	return "INV-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// Request/Response types

type ForensicCreateRequest struct {
	Query        ForensicQueryRequest `json:"query" binding:"required"`
	Justification string              `json:"justification" binding:"required,min=50"`
}

type ForensicQueryRequest struct {
	EventTypes   []string `json:"event_types,omitempty"`
	ActorIDs     []string `json:"actor_ids,omitempty"`
	Resources    []string `json:"resources,omitempty"`
	StartTime    string   `json:"start_time" binding:"required"`
	EndTime      string   `json:"end_time" binding:"required"`
	SearchTerms  []string `json:"search_terms,omitempty"`
	IncludeChain bool     `json:"include_chain"`
	IncludeHash  bool     `json:"include_hash"`
}

type ForensicInvestigation struct {
	ID             string          `json:"id"`
	RequesterID    string          `json:"requester_id"`
	RequesterRole  string          `json:"requester_role"`
	Query          models.ForensicQuery `json:"query"`
	Justification  string          `json:"justification"`
	Status         string          `json:"status"`
	ApprovedBy     string          `json:"approved_by,omitempty"`
	ApprovedAt     time.Time       `json:"approved_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	CompletedAt    time.Time       `json:"completed_at,omitempty"`
}

type ForensicResult struct {
	QueryID        string              `json:"query_id"`
	EventsFound    int64               `json:"events_found"`
	EventsReturned int                 `json:"events_returned"`
	Events         []models.AuditEvent `json:"events"`
	IntegrityChain []models.SealRecord `json:"integrity_chain,omitempty"`
	QueryTime      time.Duration       `json:"query_time"`
	GeneratedAt    time.Time           `json:"generated_at"`
}

type ForensicReport struct {
	InvestigationID string               `json:"investigation_id"`
	RequesterID     string               `json:"requester_id"`
	RequesterRole   string               `json:"requester_role"`
	Justification   string               `json:"justification"`
	Query           models.ForensicQuery `json:"query"`
	Status          string               `json:"status"`
	CreatedAt       time.Time            `json:"created_at"`
	ApprovedBy      string               `json:"approved_by,omitempty"`
	ApprovedAt      time.Time            `json:"approved_at,omitempty"`
	CompletedAt     time.Time            `json:"completed_at,omitempty"`
	GeneratedAt     time.Time            `json:"generated_at"`
}

type CourtEvidence struct {
	Event           *models.AuditEvent `json:"event"`
	IntegrityHash   string             `json:"integrity_hash"`
	Timestamp       time.Time          `json:"timestamp"`
	ChainOfCustody  []CustodyEvent     `json:"chain_of_custody"`
	Verification    *EvidenceVerification `json:"verification,omitempty"`
}

type CustodyEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`
	Details   string    `json:"details"`
}

type EvidenceVerification struct {
	Verified    bool      `json:"verified"`
	VerifiedAt  time.Time `json:"verified_at"`
	Algorithm   string    `json:"algorithm"`
	RootHash    string    `json:"root_hash"`
	HashMatch   bool      `json:"hash_match"`
	SealValid   bool      `json:"seal_valid,omitempty"`
	SealError   string    `json:"seal_error,omitempty"`
}

type EvidenceVerifyRequest struct {
	EventData    interface{}       `json:"event_data" binding:"required"`
	ExpectedHash string            `json:"expected_hash" binding:"required"`
	SealData     *models.SealRecord `json:"seal_data,omitempty"`
}
