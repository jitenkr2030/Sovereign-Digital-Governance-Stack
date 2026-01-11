package activities

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"neam-platform/approval/backend/models"
	"neam-platform/shared"
)

// ApprovalActivities implements all activities for approval workflows
type ApprovalActivities struct {
	postgreSQL *shared.PostgreSQL
	redis      *shared.Redis
	logger     *shared.Logger
}

// NewApprovalActivities creates a new instance of approval activities
func NewApprovalActivities(
	postgreSQL *shared.PostgreSQL,
	redis *shared.Redis,
	logger *shared.Logger,
) *ApprovalActivities {
	return &ApprovalActivities{
		postgreSQL: postgreSQL,
		redis:      redis,
		logger:     logger,
	}
}

// InitializeApprovalRequestActivity initializes a new approval request
func (a *ApprovalActivities) InitializeApprovalRequestActivity(ctx context.Context, input ApprovalWorkflowInput) error {
	log.Printf("Initializing approval request: %s", input.RequestID)

	// Convert context data to JSON
	contextJSON, err := json.Marshal(input.ContextData)
	if err != nil {
		return fmt.Errorf("failed to marshal context data: %w", err)
	}

	// Calculate deadline
	deadline := time.Now().Add(time.Duration(input.PolicyConfig.TimeoutHours) * time.Hour)

	// Create approval request
	request := models.ApprovalRequest{
		ID:           uuid.MustParse(input.RequestID),
		DefinitionID: uuid.MustParse(input.DefinitionID),
		RequesterID:  uuid.MustParse(input.RequesterID),
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		ContextData:  contextJSON,
		Status:       models.StatusPending,
		CurrentStep:  1,
		Priority:     input.Priority,
		Deadline:     deadline,
		WorkflowID:   fmt.Sprintf("approval-%s", input.RequestID),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Insert into database
	query := `
		INSERT INTO approval_requests 
		(id, definition_id, requester_id, resource_type, resource_id, context_data, 
		status, current_step, priority, deadline, workflow_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = a.postgreSQL.ExecContext(ctx, query,
		request.ID, request.DefinitionID, request.RequesterID, request.ResourceType,
		request.ResourceID, request.ContextData, request.Status, request.CurrentStep,
		request.Priority, request.Deadline, request.WorkflowID, request.CreatedAt, request.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert approval request: %w", err)
	}

	// Create approver assignments
	for i, approverID := range input.ApproverIDs {
		assignment := models.ApproverAssignment{
			ID:         uuid.New(),
			RequestID:  request.ID,
			ApproverID: uuid.MustParse(approverID),
			Role:       fmt.Sprintf("approver_level_%d", i+1),
			StepOrder:  i + 1,
			HasActed:   false,
			AssignedAt: time.Now(),
		}

		assignQuery := `
			INSERT INTO approver_assignments
			(id, request_id, approver_id, role, step_order, has_acted, assigned_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		_, err = a.postgreSQL.ExecContext(ctx, assignQuery,
			assignment.ID, assignment.RequestID, assignment.ApproverID,
			assignment.Role, assignment.StepOrder, assignment.HasActed, assignment.AssignedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert approver assignment: %w", err)
		}
	}

	a.logger.Info("Approval request initialized",
		"request_id", input.RequestID,
		"approvers_count", len(input.ApproverIDs),
	)

	return nil
}

// RecordApprovalActionActivity records an approval/rejection action
func (a *ApprovalActivities) RecordApprovalActionActivity(ctx context.Context, input RecordApprovalInput) error {
	log.Printf("Recording approval action: %s by %s", input.RequestID, input.ApproverID)

	requestID := uuid.MustParse(input.RequestID)
	approverID := uuid.MustParse(input.ApproverID)

	// Get previous hash for chain
	var previousHash string
	prevHashQuery := `
		SELECT signature_hash FROM approval_actions 
		WHERE request_id = $1 
		ORDER BY timestamp DESC LIMIT 1
	`
	err := a.postgreSQL.QueryRowContext(ctx, prevHashQuery, requestID).Scan(&previousHash)
	if err != nil {
		previousHash = "" // First action has no previous hash
	}

	// Generate signature hash
	timestamp := time.Now()
	actionData := fmt.Sprintf("%s|%s|%s|%s|%s",
		input.RequestID, input.ApproverID, input.ActionType, timestamp.Format(time.RFC3339), previousHash,
	)
	signatureHash := generateSignatureHash(actionData)

	// Determine actor type (user or delegated)
	actorType := "user"
	var onBehalfOf *uuid.UUID
	if input.DelegatedBy != nil {
		actorType = "delegated"
		delegatedID := uuid.MustParse(*input.DelegatedBy)
		onBehalfOf = &delegatedID
	}

	// Create action record
	action := models.ApprovalAction{
		ID:            uuid.New(),
		RequestID:     requestID,
		ActorID:       approverID,
		ActorType:     actorType,
		ActionType:    models.ActionType(input.ActionType),
		OnBehalfOf:    onBehalfOf,
		StepNumber:    1, // Will be updated based on assignment
		Comments:      input.Comments,
		SignatureHash: signatureHash,
		PreviousHash:  previousHash,
		Timestamp:     timestamp,
	}

	// Insert action
	query := `
		INSERT INTO approval_actions
		(id, request_id, actor_id, actor_type, action_type, on_behalf_of, step_number, 
		comments, signature_hash, previous_hash, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = a.postgreSQL.ExecContext(ctx, query,
		action.ID, action.RequestID, action.ActorID, action.ActorType,
		action.ActionType, action.OnBehalfOf, action.StepNumber,
		action.Comments, action.SignatureHash, action.PreviousHash, action.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to insert approval action: %w", err)
	}

	// Update approver assignment
	updateQuery := `
		UPDATE approver_assignments
		SET has_acted = true, acted_at = $1, action_type = $2
		WHERE request_id = $3 AND approver_id = $4
	`

	_, err = a.postgreSQL.ExecContext(ctx, updateQuery, time.Now(), action.ActionType, requestID, approverID)
	if err != nil {
		return fmt.Errorf("failed to update approver assignment: %w", err)
	}

	// Update request status if needed
	// This will be handled by the workflow based on quorum

	a.logger.Info("Approval action recorded",
		"request_id", input.RequestID,
		"approver_id", input.ApproverID,
		"action", input.ActionType,
	)

	return nil
}

// RecordApprovalInput represents input for recording an approval action
type RecordApprovalInput struct {
	RequestID   string  `json:"request_id"`
	ApproverID  string  `json:"approver_id"`
	ActionType  string  `json:"action_type"`
	Comments    string  `json:"comments"`
	DelegatedBy *string `json:"delegated_by,omitempty"`
}

// SendApprovalRequestNotificationsActivity sends notifications about approval status
func (a *ApprovalActivities) SendApprovalRequestNotificationsActivity(ctx context.Context, requestID string, notificationType string) error {
	log.Printf("Sending approval notifications: %s - Type: %s", requestID, notificationType)

	// Get request details
	request, err := a.getApprovalRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get approval request: %w", err)
	}

	// Get approvers
	approvers, err := a.getApproversForRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get approvers: %w", err)
	}

	// Get requester
	requester, err := a.getUserByID(ctx, request.RequesterID.String())
	if err != nil {
		return fmt.Errorf("failed to get requester: %w", err)
	}

	// Send notifications based on type
	switch notificationType {
	case "approval_required":
		for _, approver := range approvers {
			if !approver.HasActed {
				a.sendNotification(ctx, NotificationInput{
					RecipientID: approver.ApproverID.String(),
					Type:        "approval_required",
					Channel:     "in_app",
					Payload: map[string]interface{}{
						"request_id":    requestID,
						"resource_type": request.ResourceType,
						"resource_id":   request.ResourceID,
						"requester":     requester.Name,
						"deadline":      request.Deadline,
						"priority":      request.Priority,
					},
				})
			}
		}
	case "approved":
		a.sendNotification(ctx, NotificationInput{
			RecipientID: requester.ID.String(),
			Type:        "request_approved",
			Channel:     "in_app",
			Payload: map[string]interface{}{
				"request_id":    requestID,
				"resource_type": request.ResourceType,
				"resource_id":   request.ResourceID,
				"message":       "Your approval request has been approved",
			},
		})
	case "rejected":
		a.sendNotification(ctx, NotificationInput{
			RecipientID: requester.ID.String(),
			Type:        "request_rejected",
			Channel:     "in_app",
			Payload: map[string]interface{}{
				"request_id":    requestID,
				"resource_type": request.ResourceType,
				"resource_id":   request.ResourceID,
				"message":       "Your approval request has been rejected",
			},
		})
	case "expired":
		a.sendNotification(ctx, NotificationInput{
			RecipientID: requester.ID.String(),
			Type:        "request_expired",
			Channel:     "in_app",
			Payload: map[string]interface{}{
				"request_id":    requestID,
				"resource_type": request.ResourceType,
				"resource_id":   request.ResourceID,
				"message":       "Your approval request has expired",
			},
		})
	}

	return nil
}

// ExecuteApprovedActionActivity executes the action after approval is granted
func (a *ApprovalActivities) ExecuteApprovedActionActivity(ctx context.Context, requestID string) error {
	log.Printf("Executing approved action: %s", requestID)

	// Get request details
	request, err := a.getApprovalRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get approval request: %w", err)
	}

	// Based on resource type, execute appropriate action
	switch request.ResourceType {
	case "policy":
		// Activate the policy
		log.Printf("Activating policy: %s", request.ResourceID)
	case "budget":
		// Release the budget
		log.Printf("Releasing budget: %s", request.ResourceID)
	case "transaction":
		// Process the transaction
		log.Printf("Processing transaction: %s", request.ResourceID)
	default:
		log.Printf("Executing custom action for: %s (%s)", request.ResourceID, request.ResourceType)
	}

	a.logger.Info("Approved action executed",
		"request_id", requestID,
		"resource_type", request.ResourceType,
		"resource_id", request.ResourceID,
	)

	return nil
}

// FinalizeApprovalWorkflowActivity finalizes the workflow
func (a *ApprovalActivities) FinalizeApprovalWorkflowActivity(ctx context.Context, requestID string, status models.ApprovalStatus) error {
	log.Printf("Finalizing approval workflow: %s with status %s", requestID, status)

	requestUUID := uuid.MustParse(requestID)
	now := time.Now()

	// Update request status
	query := `
		UPDATE approval_requests
		SET status = $1, updated_at = $2, completed_at = $3
		WHERE id = $4
	`

	_, err := a.postgreSQL.ExecContext(ctx, query, status, now, now, requestUUID)
	if err != nil {
		return fmt.Errorf("failed to update approval request status: %w", err)
	}

	a.logger.Info("Approval workflow finalized",
		"request_id", requestID,
		"status", status,
	)

	return nil
}

// GetApprovalStatusActivity retrieves the current status of an approval request
func (a *ApprovalActivities) GetApprovalStatusActivity(ctx context.Context, requestID string) (*ApprovalStatusOutput, error) {
	log.Printf("Getting approval status: %s", requestID)

	request, err := a.getApprovalRequest(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get approval request: %w", err)
	}

	approvers, err := a.getApproversForRequest(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get approvers: %w", err)
	}

	// Calculate status
	approvalsCount := 0
	rejectionsCount := 0
	pendingCount := 0

	for _, approver := range approvers {
		if approver.HasActed {
			if approver.ActionType != nil {
				switch *approver.ActionType {
				case models.ActionApprove:
					approvalsCount++
				case models.ActionReject:
					rejectionsCount++
				}
			}
		} else {
			pendingCount++
		}
	}

	output := &ApprovalStatusOutput{
		RequestID:       requestID,
		Status:          request.Status,
		Deadline:        request.Deadline,
		ApprovalsCount:  approvalsCount,
		RejectionsCount: rejectionsCount,
		PendingCount:    pendingCount,
		TotalApprovers:  len(approvers),
	}

	// Get policy config
	def, err := a.getDefinitionByID(ctx, request.DefinitionID.String())
	if err == nil && def != nil {
		output.QuorumRequired = def.PolicyConfig.QuorumRequired
		output.QuorumMet = approvalsCount >= def.PolicyConfig.QuorumRequired
	}

	return output, nil
}

// ApprovalStatusOutput represents the current status of an approval request
type ApprovalStatusOutput struct {
	RequestID       string             `json:"request_id"`
	Status          models.ApprovalStatus `json:"status"`
	Deadline        time.Time          `json:"deadline"`
	ApprovalsCount  int                `json:"approvals_count"`
	RejectionsCount int                `json:"rejections_count"`
	PendingCount    int                `json:"pending_count"`
	TotalApprovers  int                `json:"total_approvers"`
	QuorumRequired  int                `json:"quorum_required"`
	QuorumMet       bool               `json:"quorum_met"`
}

// VerifyAuditTrailActivity verifies the integrity of the audit trail
func (a *ApprovalActivities) VerifyAuditTrailActivity(ctx context.Context, requestID string) (*AuditVerificationOutput, error) {
	log.Printf("Verifying audit trail: %s", requestID)

	requestUUID := uuid.MustParse(requestID)

	// Get all actions for the request
	actions, err := a.getActionsForRequest(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}

	output := &AuditVerificationOutput{
		RequestID:      requestID,
		ActionsCount:   len(actions),
		HashChainValid: true,
		BrokenAtAction: "",
	}

	// Verify hash chain
	var previousHash string
	for i, action := range actions {
		expectedHash := generateSignatureHash(fmt.Sprintf("%s|%s|%s|%s|%s",
			action.RequestID, action.ActorID, action.ActionType, action.Timestamp.Format(time.RFC3339), previousHash,
		))

		if action.SignatureHash != expectedHash {
			output.HashChainValid = false
			output.BrokenAtAction = action.ID.String()
			output.VerificationMessage = fmt.Sprintf("Hash mismatch at action %d", i+1)
			break
		}

		previousHash = action.SignatureHash
	}

	if output.HashChainValid {
		output.VerificationMessage = "Audit trail hash chain is valid"
	}

	return output, nil
}

// AuditVerificationOutput represents the result of audit trail verification
type AuditVerificationOutput struct {
	RequestID         string `json:"request_id"`
	ActionsCount      int    `json:"actions_count"`
	HashChainValid    bool   `json:"hash_chain_valid"`
	BrokenAtAction    string `json:"broken_at_action,omitempty"`
	VerificationMessage string `json:"verification_message"`
}

// Helper functions

func generateSignatureHash(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (a *ApprovalActivities) getApprovalRequest(ctx context.Context, requestID string) (*models.ApprovalRequest, error) {
	requestUUID := uuid.MustParse(requestID)
	query := `SELECT * FROM approval_requests WHERE id = $1`
	row := a.postgreSQL.QueryRowContext(ctx, query, requestUUID)
	return scanApprovalRequest(row)
}

func (a *ApprovalActivities) getApproversForRequest(ctx context.Context, requestID string) ([]models.ApproverAssignment, error) {
	requestUUID := uuid.MustParse(requestID)
	query := `SELECT * FROM approver_assignments WHERE request_id = $1 ORDER BY step_order`
	rows, err := a.postgreSQL.QueryContext(ctx, query, requestUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []models.ApproverAssignment
	for rows.Next() {
		var a models.ApproverAssignment
		if err := rows.Scan(&a.ID, &a.RequestID, &a.ApproverID, &a.Role, &a.StepOrder,
			&a.HasActed, &a.ActedAt, &a.ActionType, &a.AssignedAt); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}
	return assignments, nil
}

func (a *ApprovalActivities) getActionsForRequest(ctx context.Context, requestID string) ([]models.ApprovalAction, error) {
	requestUUID := uuid.MustParse(requestID)
	query := `SELECT * FROM approval_actions WHERE request_id = $1 ORDER BY timestamp`
	rows, err := a.postgreSQL.QueryContext(ctx, query, requestUUID)
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
			return nil, err
		}
		actions = append(actions, act)
	}
	return actions, nil
}

func (a *ApprovalActivities) getDefinitionByID(ctx context.Context, definitionID string) (*models.ApprovalDefinition, error) {
	defUUID := uuid.MustParse(definitionID)
	query := `SELECT * FROM approval_definitions WHERE id = $1`
	row := a.postgreSQL.QueryRowContext(ctx, query, defUUID)
	return scanDefinition(row)
}

func (a *ApprovalActivities) getUserByID(ctx context.Context, userID string) (*UserInfo, error) {
	userUUID := uuid.MustParse(userID)
	query := `SELECT id, name, email FROM users WHERE id = $1`
	row := a.postgreSQL.QueryRowContext(ctx, query, userUUID)
	var user UserInfo
	if err := row.Scan(&user.ID, &user.Name, &user.Email); err != nil {
		return nil, err
	}
	return &user, nil
}

func (a *ApprovalActivities) sendNotification(ctx context.Context, input NotificationInput) error {
	query := `
		INSERT INTO approval_notifications
		(id, request_id, recipient_id, notification_type, channel, payload, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'PENDING', NOW())
	`
	payloadJSON, _ := json.Marshal(input.Payload)
	_, err := a.postgreSQL.ExecContext(ctx, query, uuid.New(), input.RequestID, input.RecipientID,
		input.Type, input.Channel, payloadJSON)
	return err
}

// NotificationInput represents input for a notification
type NotificationInput struct {
	RequestID   string                 `json:"request_id"`
	RecipientID string                 `json:"recipient_id"`
	Type        string                 `json:"type"`
	Channel     string                 `json:"channel"` // 'email', 'slack', 'in_app'
	Payload     map[string]interface{} `json:"payload"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

// scanApprovalRequest scans a database row into ApprovalRequest
func scanApprovalRequest(row *sql.Row) (*models.ApprovalRequest, error) {
	var r models.ApprovalRequest
	err := row.Scan(&r.ID, &r.DefinitionID, &r.RequesterID, &r.ResourceType, &r.ResourceID,
		&r.ContextData, &r.Status, &r.CurrentStep, &r.Priority, &r.Deadline,
		&r.WorkflowID, &r.Metadata, &r.CreatedAt, &r.UpdatedAt, &r.CompletedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// scanDefinition scans a database row into ApprovalDefinition
func scanDefinition(row *sql.Row) (*models.ApprovalDefinition, error) {
	var d models.ApprovalDefinition
	err := row.Scan(&d.ID, &d.Name, &d.Description, &d.Category, &d.PolicyConfig,
		&d.Version, &d.IsActive, &d.CreatedAt, &d.UpdatedAt, &d.CreatedBy)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
