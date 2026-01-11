package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"neam-platform/approval/backend/activities"
	"neam-platform/approval/backend/models"
	"neam-platform/approval/backend/workflows"
	"neam-platform/shared"
)

// ApprovalHandler handles approval-related API endpoints
type ApprovalHandler struct {
	postgreSQL *shared.PostgreSQL
	redis      *shared.Redis
	temporal   shared.TemporalClient
	logger     *shared.Logger
}

// NewApprovalHandler creates a new approval handler
func NewApprovalHandler(
	postgreSQL *shared.PostgreSQL,
	redis *shared.Redis,
	temporal shared.TemporalClient,
	logger *shared.Logger,
) *ApprovalHandler {
	return &ApprovalHandler{
		postgreSQL: postgreSQL,
		redis:      redis,
		temporal:   temporal,
		logger:     logger,
	}
}

// CreateApprovalDefinitionRequest represents a request to create an approval definition
type CreateApprovalDefinitionRequest struct {
	Name          string                 `json:"name" binding:"required"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category" binding:"required"`
	PolicyConfig  models.PolicyConfig    `json:"policy_config" binding:"required"`
	CreatedBy     string                 `json:"created_by" binding:"required"`
}

// CreateApprovalRequestRequest represents a request to create an approval request
type CreateApprovalRequestRequest struct {
	DefinitionID  string                 `json:"definition_id" binding:"required"`
	RequesterID   string                 `json:"requester_id" binding:"required"`
	ResourceType  string                 `json:"resource_type" binding:"required"`
	ResourceID    string                 `json:"resource_id" binding:"required"`
	ContextData   map[string]interface{} `json:"context_data"`
	ApproverIDs   []string               `json:"approver_ids" binding:"required,min=1"`
	Priority      string                 `json:"priority"`
}

// ApprovalDecisionRequest represents a request to submit an approval decision
type ApprovalDecisionRequest struct {
	ApproverID  string `json:"approver_id" binding:"required"`
	Decision    string `json:"decision" binding:"required,oneof=APPROVE REJECT ABSTAIN"`
	Comments    string `json:"comments"`
	DelegatedBy string `json:"delegated_by,omitempty"`
}

// CreateDelegationRequest represents a request to create a delegation
type CreateDelegationRequest struct {
	DelegatorID   string          `json:"delegator_id" binding:"required"`
	DelegateeID   string          `json:"delegatee_id" binding:"required"`
	DefinitionID  string          `json:"definition_id,omitempty"`
	Scope         json.RawMessage `json:"scope"`
	ValidFrom     time.Time       `json:"valid_from"`
	ValidUntil    time.Time       `json:"valid_until" binding:"required"`
	Reason        string          `json:"reason"`
}

// CreateDefinition creates a new approval definition
// POST /api/v1/approval-definitions
func (h *ApprovalHandler) CreateDefinition(c *gin.Context) {
	var req CreateApprovalDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	definition := models.ApprovalDefinition{
		ID:           uuid.New(),
		Name:         req.Name,
		Description:  req.Description,
		Category:     req.Category,
		PolicyConfig: req.PolicyConfig,
		Version:      1,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    uuid.MustParse(req.CreatedBy),
	}

	query := `
		INSERT INTO approval_definitions 
		(id, name, description, category, policy_config, version, is_active, created_at, updated_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := h.postgreSQL.ExecContext(c.Request.Context(), query,
		definition.ID, definition.Name, definition.Description, definition.Category,
		definition.PolicyConfig, definition.Version, definition.IsActive,
		definition.CreatedAt, definition.UpdatedAt, definition.CreatedBy,
	)
	if err != nil {
		h.logger.Error("Failed to create approval definition", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create approval definition"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          definition.ID,
		"name":        definition.Name,
		"message":     "Approval definition created successfully",
	})
}

// GetDefinitions returns all approval definitions
// GET /api/v1/approval-definitions
func (h *ApprovalHandler) GetDefinitions(c *gin.Context) {
	category := c.Query("category")
	activeOnly := c.Query("active") == "true"

	query := `SELECT id, name, description, category, policy_config, version, is_active, created_at, updated_at, created_by 
	          FROM approval_definitions WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, category)
		argIndex++
	}

	if activeOnly {
		query += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, true)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	rows, err := h.postgreSQL.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch definitions"})
		return
	}
	defer rows.Close()

	var definitions []models.ApprovalDefinition
	for rows.Next() {
		var d models.ApprovalDefinition
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.Category, &d.PolicyConfig,
			&d.Version, &d.IsActive, &d.CreatedAt, &d.UpdatedAt, &d.CreatedBy); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan definition"})
			return
		}
		definitions = append(definitions, d)
	}

	c.JSON(http.StatusOK, gin.H{
		"definitions": definitions,
		"total":       len(definitions),
	})
}

// CreateApprovalRequest initiates a new approval request
// POST /api/v1/approval-requests
func (h *ApprovalHandler) CreateApprovalRequest(c *gin.Context) {
	var req CreateApprovalRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestID := uuid.New()
	priority := req.Priority
	if priority == "" {
		priority = "normal"
	}

	// Get definition to validate policy
	def, err := h.getDefinitionByID(c, req.DefinitionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid definition ID"})
		return
	}

	// Validate approver count matches policy
	if len(req.ApproverIDs) != def.PolicyConfig.TotalApprovers {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Expected %d approvers, got %d",
				def.PolicyConfig.TotalApprovers, len(req.ApproverIDs)),
		})
		return
	}

	// Create workflow input
	workflowInput := workflows.ApprovalWorkflowInput{
		RequestID:      requestID.String(),
		DefinitionID:   req.DefinitionID,
		RequesterID:    req.RequesterID,
		ResourceType:   req.ResourceType,
		ResourceID:     req.ResourceID,
		ContextData:    req.ContextData,
		ApproverIDs:    req.ApproverIDs,
		PolicyConfig:   def.PolicyConfig,
		Priority:       priority,
		CreatedBy:      req.RequesterID,
	}

	// Start Temporal workflow
	workflowOptions := shared.StartWorkflowOptions{
		ID:        fmt.Sprintf("approval-%s", requestID.String()),
		TaskQueue: "approval-task-queue",
	}

	workflowRun, err := h.temporal.ExecuteWorkflow(c.Request.Context(), workflowOptions,
		workflows.ExecuteApprovalWorkflow, workflowInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start approval workflow"})
		return
	}

	// Store request in database
	contextJSON, _ := json.Marshal(req.ContextData)
	deadline := time.Now().Add(time.Duration(def.PolicyConfig.TimeoutHours) * time.Hour)

	query := `
		INSERT INTO approval_requests
		(id, definition_id, requester_id, resource_type, resource_id, context_data,
		status, current_step, priority, deadline, workflow_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = h.postgreSQL.ExecContext(c.Request.Context(), query,
		requestID, req.DefinitionID, req.RequesterID, req.ResourceType, req.ResourceID,
		contextJSON, models.StatusPending, 1, priority, deadline,
		workflowRun.GetID(), time.Now(), time.Now(),
	)
	if err != nil {
		h.logger.Error("Failed to store approval request", "error", err)
	}

	// Create approver assignments
	for i, approverID := range req.ApproverIDs {
		assignQuery := `
			INSERT INTO approver_assignments
			(id, request_id, approver_id, role, step_order, has_acted, assigned_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err = h.postgreSQL.ExecContext(c.Request.Context(), assignQuery,
			uuid.New(), requestID, approverID,
			fmt.Sprintf("approver_level_%d", i+1), i+1, false, time.Now(),
		)
		if err != nil {
			h.logger.Error("Failed to create approver assignment", "error", err)
		}
	}

	c.JSON(http.StatusAccepted, gin.H{
		"request_id":   requestID,
		"workflow_id":  workflowRun.GetID(),
		"status":       models.StatusPending,
		"deadline":     deadline,
		"message":      "Approval request created and workflow initiated",
	})
}

// GetApprovalRequest returns details of an approval request
// GET /api/v1/approval-requests/:id
func (h *ApprovalHandler) GetApprovalRequest(c *gin.Context) {
	requestID := c.Param("id")
	requestUUID := uuid.MustParse(requestID)

	query := `SELECT * FROM approval_requests WHERE id = $1`
	row := h.postgreSQL.QueryRowContext(c.Request.Context(), query, requestUUID)
	request, err := scanApprovalRequestRow(row)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Approval request not found"})
		return
	}

	// Get approvers
	approvers, _ := h.getApproversForRequest(c.Request.Context(), requestID)

	// Get actions
	actions, _ := h.getActionsForRequest(c.Request.Context(), requestID)

	c.JSON(http.StatusOK, gin.H{
		"request":      request,
		"approvers":    approvers,
		"actions":      actions,
	})
}

// GetPendingApprovals returns pending approvals for a user
// GET /api/v1/approvals/pending
func (h *ApprovalHandler) GetPendingApprovals(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userUUID := uuid.MustParse(userID)

	query := `
		SELECT ar.* FROM approval_requests ar
		JOIN approver_assignments aa ON ar.id = aa.request_id
		WHERE aa.approver_id = $1 AND aa.has_acted = false AND ar.status = 'PENDING'
		ORDER BY ar.priority DESC, ar.deadline ASC
	`

	rows, err := h.postgreSQL.QueryContext(c.Request.Context(), query, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending approvals"})
		return
	}
	defer rows.Close()

	var requests []models.ApprovalRequest
	for rows.Next() {
		var r models.ApprovalRequest
		if err := rows.Scan(&r.ID, &r.DefinitionID, &r.RequesterID, &r.ResourceType, &r.ResourceID,
			&r.ContextData, &r.Status, &r.CurrentStep, &r.Priority, &r.Deadline,
			&r.WorkflowID, &r.Metadata, &r.CreatedAt, &r.UpdatedAt, &r.CompletedAt); err != nil {
			continue
		}
		requests = append(requests, r)
	}

	c.JSON(http.StatusOK, gin.H{
		"pending_approvals": requests,
		"total":             len(requests),
	})
}

// SubmitDecision submits an approval decision
// POST /api/v1/approval-requests/:id/decide
func (h *ApprovalHandler) SubmitDecision(c *gin.Context) {
	requestID := c.Param("id")
	var req ApprovalDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate approver has not already acted
	alreadyActed, err := h.hasApproverActed(c.Request.Context(), requestID, req.ApproverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check approver status"})
		return
	}
	if alreadyActed {
		c.JSON(http.StatusConflict, gin.H{"error": "Approver has already submitted a decision"})
		return
	}

	// Check delegation if applicable
	if req.DelegatedBy != "" {
		valid, err := h.validateDelegation(c.Request.Context(), req.DelegatedBy, req.ApproverID)
		if err != nil || !valid {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid or expired delegation"})
			return
		}
	}

	// Send signal to Temporal workflow
	signal := workflows.ApprovalSignal{
		ApproverID:  req.ApproverID,
		ActionType:  req.Decision,
		Comments:    req.Comments,
		DelegatedBy: strPtr(req.DelegatedBy),
		Timestamp:   time.Now(),
	}

	err = h.temporal.SignalWorkflow(c.Request.Context(), fmt.Sprintf("approval-%s", requestID), "", "approval_decision", signal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit decision to workflow"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"request_id": requestID,
		"decision":   req.Decision,
		"message":    "Decision submitted successfully",
	})
}

// GetAuditTrail returns the audit trail for an approval request
// GET /api/v1/approval-requests/:id/audit-trail
func (h *ApprovalHandler) GetAuditTrail(c *gin.Context) {
	requestID := c.Param("id")

	// Get all actions
	actions, err := h.getActionsForRequest(c.Request.Context(), requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit trail"})
		return
	}

	// Verify hash chain
	hashChainValid := true
	var previousHash string
	for i, action := range actions {
		expectedHash := generateActionHash(action.RequestID, action.ActorID, action.ActionType, action.Timestamp, previousHash)
		if action.SignatureHash != expectedHash {
			hashChainValid = false
			actions[i] = action // Mark as invalid in response
			break
		}
		previousHash = action.SignatureHash
	}

	// Get user details for each action
	trail := make([]models.AuditTrailEntry, 0, len(actions))
	for _, action := range actions {
		user, _ := h.getUserByID(c.Request.Context(), action.ActorID.String())
		trail = append(trail, models.AuditTrailEntry{
			ActionID:       action.ID,
			ActorID:        action.ActorID,
			ActorName:      user.Name,
			ActionType:     action.ActionType,
			Comments:       action.Comments,
			Timestamp:      action.Timestamp,
			SignatureValid: hashChainValid,
			DelegatedFrom:  action.OnBehalfOf,
			IPAddress:      action.IPAddress,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"request_id":    requestID,
		"audit_trail":   trail,
		"hash_chain_valid": hashChainValid,
	})
}

// CreateDelegation creates a new delegation
// POST /api/v1/delegations
func (h *ApprovalHandler) CreateDelegation(c *gin.Context) {
	var req CreateDelegationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	delegation := models.Delegation{
		ID:          uuid.New(),
		DelegatorID: uuid.MustParse(req.DelegatorID),
		DelegateeID: uuid.MustParse(req.DelegateeID),
		Scope:       req.Scope,
		ValidFrom:   req.ValidFrom,
		ValidUntil:  req.ValidUntil,
		Reason:      req.Reason,
		IsRevoked:   false,
		CreatedAt:   time.Now(),
	}

	if req.DefinitionID != "" {
		delegation.DefinitionID = uuid.MustParse(req.DefinitionID)
	}

	query := `
		INSERT INTO delegations
		(id, delegator_id, delegatee_id, definition_id, scope, valid_from, valid_until, is_revoked, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := h.postgreSQL.ExecContext(c.Request.Context(), query,
		delegation.ID, delegation.DelegatorID, delegation.DelegateeID,
		delegation.DefinitionID, delegation.Scope, delegation.ValidFrom,
		delegation.ValidUntil, delegation.IsRevoked, delegation.Reason, delegation.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create delegation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"delegation_id": delegation.ID,
		"message":       "Delegation created successfully",
	})
}

// GetDelegations returns delegations for a user
// GET /api/v1/delegations
func (h *ApprovalHandler) GetDelegations(c *gin.Context) {
	userID := c.Query("user_id")
	delegatorOnly := c.Query("as_delegator") == "true"

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userUUID := uuid.MustParse(userID)

	var query string
	var args []interface{}

	if delegatorOnly {
		query = `SELECT * FROM delegations WHERE delegator_id = $1 AND is_revoked = false AND valid_until > NOW()`
		args = []interface{}{userUUID}
	} else {
		query = `SELECT * FROM delegations WHERE delegatee_id = $1 AND is_revoked = false AND valid_from <= NOW() AND valid_until > NOW()`
		args = []interface{}{userUUID}
	}

	rows, err := h.postgreSQL.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch delegations"})
		return
	}
	defer rows.Close()

	var delegations []models.Delegation
	for rows.Next() {
		var d models.Delegation
		if err := rows.Scan(&d.ID, &d.DelegatorID, &d.DelegateeID, &d.DefinitionID,
			&d.Scope, &d.ValidFrom, &d.ValidUntil, &d.IsRevoked, &d.Reason, &d.CreatedAt, &d.RevokedAt); err != nil {
			continue
		}
		delegations = append(delegations, d)
	}

	c.JSON(http.StatusOK, gin.H{
		"delegations": delegations,
		"total":       len(delegations),
	})
}

// Helper functions

func (h *ApprovalHandler) getDefinitionByID(c *gin.Context, definitionID string) (*models.ApprovalDefinition, error) {
	defUUID := uuid.MustParse(definitionID)
	query := `SELECT * FROM approval_definitions WHERE id = $1`
	row := h.postgreSQL.QueryRowContext(c.Request.Context(), query, defUUID)
	return scanDefinitionRow(row)
}

func (h *ApprovalHandler) getApproversForRequest(ctx context.Context, requestID string) ([]models.ApproverAssignment, error) {
	requestUUID := uuid.MustParse(requestID)
	query := `SELECT * FROM approver_assignments WHERE request_id = $1 ORDER BY step_order`
	rows, err := h.postgreSQL.QueryContext(ctx, query, requestUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []models.ApproverAssignment
	for rows.Next() {
		var a models.ApproverAssignment
		if err := rows.Scan(&a.ID, &a.RequestID, &a.ApproverID, &a.Role, &a.StepOrder,
			&a.HasActed, &a.ActedAt, &a.ActionType, &a.AssignedAt); err != nil {
			continue
		}
		assignments = append(assignments, a)
	}
	return assignments, nil
}

func (h *ApprovalHandler) getActionsForRequest(ctx context.Context, requestID string) ([]models.ApprovalAction, error) {
	requestUUID := uuid.MustParse(requestID)
	query := `SELECT * FROM approval_actions WHERE request_id = $1 ORDER BY timestamp`
	rows, err := h.postgreSQL.QueryContext(ctx, query, requestUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []models.ApprovalAction
	for rows.Next() {
		var act models.ApprovalAction
		if err := rows.Scan(&act.ID, &act.RequestID, &act.ActorID, &act.ActorType,
			&act.ActionType, &act.OnBehalfOf, &act.StepNumber, &act.Comments,
			&act.SignatureHash, &act.PreviousHash, &act.IPAddress, &act.UserAgent, &act.Timestamp); err != nil {
			continue
		}
		actions = append(actions, act)
	}
	return actions, nil
}

func (h *ApprovalHandler) hasApproverActed(ctx context.Context, requestID, approverID string) (bool, error) {
	requestUUID := uuid.MustParse(requestID)
	approverUUID := uuid.MustParse(approverID)
	query := `SELECT has_acted FROM approver_assignments WHERE request_id = $1 AND approver_id = $2`
	var hasActed bool
	err := h.postgreSQL.QueryRowContext(ctx, query, requestUUID, approverUUID).Scan(&hasActed)
	return hasActed, err
}

func (h *ApprovalHandler) validateDelegation(ctx context.Context, delegatorID, delegateeID string) (bool, error) {
	delegatorUUID := uuid.MustParse(delegatorID)
	delegateeUUID := uuid.MustParse(delegateeID)
	query := `SELECT COUNT(*) FROM delegations 
	          WHERE delegator_id = $1 AND delegatee_id = $2 
	          AND is_revoked = false AND valid_from <= NOW() AND valid_until > NOW()`
	var count int
	err := h.postgreSQL.QueryRowContext(ctx, query, delegatorUUID, delegateeUUID).Scan(&count)
	return count > 0, err
}

func (h *ApprovalHandler) getUserByID(ctx context.Context, userID string) (*UserInfo, error) {
	userUUID := uuid.MustParse(userID)
	query := `SELECT id, name, email FROM users WHERE id = $1`
	row := h.postgreSQL.QueryRowContext(ctx, query, userUUID)
	var user UserInfo
	if err := row.Scan(&user.ID, &user.Name, &user.Email); err != nil {
		return nil, err
	}
	return &user, nil
}

func generateActionHash(requestID uuid.UUID, actorID uuid.UUID, actionType models.ActionType, timestamp time.Time, previousHash string) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s",
		requestID, actorID, actionType, timestamp.Format(time.RFC3339), previousHash,
	)
	return activities.GenerateSignatureHash(data)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Row scanning helpers
func scanApprovalRequestRow(row *sql.Row) (*models.ApprovalRequest, error) {
	var r models.ApprovalRequest
	err := row.Scan(&r.ID, &r.DefinitionID, &r.RequesterID, &r.ResourceType, &r.ResourceID,
		&r.ContextData, &r.Status, &r.CurrentStep, &r.Priority, &r.Deadline,
		&r.WorkflowID, &r.Metadata, &r.CreatedAt, &r.UpdatedAt, &r.CompletedAt)
	return &r, err
}

func scanDefinitionRow(row *sql.Row) (*models.ApprovalDefinition, error) {
	var d models.ApprovalDefinition
	err := row.Scan(&d.ID, &d.Name, &d.Description, &d.Category, &d.PolicyConfig,
		&d.Version, &d.IsActive, &d.CreatedAt, &d.UpdatedAt, &d.CreatedBy)
	return &d, err
}

// Import for database scanning
import "database/sql"
