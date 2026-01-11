package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"neam-platform/approval/backend/models"
)

// ApprovalWorkflowInput represents the input for an approval workflow
type ApprovalWorkflowInput struct {
	RequestID           string                 `json:"request_id"`
	DefinitionID        string                 `json:"definition_id"`
	RequesterID         string                 `json:"requester_id"`
	ResourceType        string                 `json:"resource_type"`
	ResourceID          string                 `json:"resource_id"`
	ContextData         map[string]interface{} `json:"context_data"`
	ApproverIDs         []string               `json:"approver_ids"`
	PolicyConfig        models.PolicyConfig    `json:"policy_config"`
	Priority            string                 `json:"priority"`
	CreatedBy           string                 `json:"created_by"`
}

// ApprovalWorkflowOutput represents the output of an approval workflow
type ApprovalWorkflowOutput struct {
	RequestID           string                  `json:"request_id"`
	FinalStatus         models.ApprovalStatus   `json:"final_status"`
	Success             bool                    `json:"success"`
	Message             string                  `json:"message"`
	ExecutionTime       time.Duration           `json:"execution_time"`
	ApprovalsReceived   int                     `json:"approvals_received"`
	RejectionsReceived  int                     `json:"rejections_received"`
	ActionsExecuted     []string                `json:"actions_executed"`
	FinalDecision       string                  `json:"final_decision"`
}

// ApprovalSignal represents a signal from an approver
type ApprovalSignal struct {
	ApproverID   string    `json:"approver_id"`
	ActionType   string    `json:"action_type"` // "APPROVE", "REJECT", "ABSTAIN"
	Comments     string    `json:"comments"`
	DelegatedBy  *string   `json:"delegated_by,omitempty"` // If acting on behalf of someone
	Timestamp    time.Time `json:"timestamp"`
}

// ApprovalTimeoutConfig defines timeout settings for approval workflows
type ApprovalTimeoutConfig struct {
	TotalTimeout      time.Duration
	ActivityTimeout   time.Duration
	DecisionTimeout   time.Duration
	HeartbeatTimeout  time.Duration
}

// DefaultTimeoutConfig returns default timeout settings
func DefaultTimeoutConfig() ApprovalTimeoutConfig {
	return ApprovalTimeoutConfig{
		TotalTimeout:     72 * time.Hour,
		ActivityTimeout:  5 * time.Minute,
		DecisionTimeout:  30 * time.Second,
		HeartbeatTimeout: 30 * time.Second,
	}
}

// ExecuteApprovalWorkflow is the main workflow for handling multi-level approvals
// This implements the Saga pattern with compensation and quorum enforcement
func ExecuteApprovalWorkflow(ctx workflow.Context, input ApprovalWorkflowInput) (*ApprovalWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)

	// Setup activity options with timeouts and retry policy
	timeoutConfig := DefaultTimeoutConfig()
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: timeoutConfig.ActivityTimeout,
		HeartbeatTimeout:    timeoutConfig.HeartbeatTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    5 * time.Minute,
			MaximumAttempts:    3,
			NonRetryableErrorTypes: []string{
				"ValidationError",
				"PolicyNotFound",
				"InsufficientPermissions",
			},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Initialize output
	output := &ApprovalWorkflowOutput{
		RequestID:          input.RequestID,
		ActionsExecuted:    make([]string, 0),
	}

	startTime := workflow.Now(ctx)

	// Calculate deadline
	deadline := startTime.Add(time.Duration(input.PolicyConfig.TimeoutHours) * time.Hour)

	// Activity references
	acts := &ApprovalActivities{}

	logger.Info("Starting approval workflow",
		"request_id", input.RequestID,
		"definition_id", input.DefinitionID,
		"quorum_required", input.PolicyConfig.QuorumRequired,
		"total_approvers", input.PolicyConfig.TotalApprovers,
	)

	// ===== STEP 1: Initialize Approval Request =====
	logger.Info("Executing activity: InitializeApprovalRequest")
	err := workflow.ExecuteActivity(ctx, acts.InitializeApprovalRequestActivity, input).Get(ctx, nil)
	if err != nil {
		output.FinalStatus = models.StatusCancelled
		output.Success = false
		output.Message = "Failed to initialize approval request: " + err.Error()
		return output, nil
	}
	output.ActionsExecuted = append(output.ActionsExecuted, "InitializeRequest:SUCCESS")

	// ===== STEP 2: Send Initial Notifications =====
	logger.Info("Executing activity: SendApprovalRequestNotifications")
	err = workflow.ExecuteActivity(ctx, acts.SendApprovalRequestNotificationsActivity, input.RequestID, "approval_required").Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send notifications", "error", err)
		output.ActionsExecuted = append(output.ActionsExecuted, "SendNotifications:FAILED")
	} else {
		output.ActionsExecuted = append(output.ActionsExecuted, "SendNotifications:SUCCESS")
	}

	// ===== STEP 3: Main Approval Loop with Signal and Timer =====
	selector := workflow.NewSelector(ctx)
	signalChannel := workflow.GetSignalChannel(ctx, "approval_decision")
	timeoutChannel := workflow.NewChannel(ctx)

	// Set up timeout timer
	var timerFuture workflow.Future
	if deadline.After(startTime) {
		timerFuture, _ = workflow.NewTimer(ctx, deadline.Sub(startTime))
	} else {
		// Immediate timeout if deadline has passed
		output.FinalStatus = models.StatusExpired
		output.Success = false
		output.Message = "Approval request expired before any action was taken"
		return output, nil
	}

	// Variables for tracking
	approvalsReceived := 0
	rejectionsReceived := 0
	actionsMap := make(map[string]string) // approverID -> action
	quorumMet := false
	workflowComplete := false

	selector.AddReceive(signalChannel, func(channel workflow.Channel, more bool) {
		var signal ApprovalSignal
		channel.Receive(ctx, &signal)

		logger.Info("Received approval signal",
			"approver_id", signal.ApproverID,
			"action", signal.ActionType,
		)

		// Validate approver
		isValidApprover := false
		for _, approverID := range input.ApproverIDs {
			if approverID == signal.ApproverID {
				isValidApprover = true
				break
			}
		}

		if !isValidApprover {
			logger.Warn("Invalid approver", "approver_id", signal.ApproverID)
			return
		}

		// Check if already acted
		if _, exists := actionsMap[signal.ApproverID]; exists {
			logger.Warn("Approver already acted", "approver_id", signal.ApproverID)
			return
		}

		// Record action
		actionsMap[signal.ApproverID] = signal.ActionType

		if signal.ActionType == "APPROVE" {
			approvalsReceived++
		} else if signal.ActionType == "REJECT" {
			rejectionsReceived++
		}

		// Execute activity to record action
		recordInput := RecordActionInput{
			RequestID:   input.RequestID,
			ApproverID:  signal.ApproverID,
			ActionType:  signal.ActionType,
			Comments:    signal.Comments,
			DelegatedBy: signal.DelegatedBy,
		}
		workflow.ExecuteActivity(ctx, acts.RecordApprovalActionActivity, recordInput).Get(ctx, nil)
		output.ActionsExecuted = append(output.ActionsExecuted, "RecordAction:"+signal.ApproverID+":"+signal.ActionType)

		// Check quorum
		if approvalsReceived >= input.PolicyConfig.QuorumRequired {
			quorumMet = true
			workflowComplete = true
			output.FinalStatus = models.StatusApproved
			output.Success = true
			output.Message = "Approval quorum met"
		} else if rejectionsReceived > (input.PolicyConfig.TotalApprovers - input.PolicyConfig.QuorumRequired) {
			workflowComplete = true
			output.FinalStatus = models.StatusRejected
			output.Success = false
			output.Message = "Approval request rejected - insufficient approvals"
		} else if (approvalsReceived + rejectionsReceived) >= input.PolicyConfig.TotalApprovers {
			// All approvers have acted but quorum not met
			workflowComplete = true
			output.FinalStatus = models.StatusRejected
			output.Success = false
			output.Message = "Approval request expired - quorum not met"
		}
	})

	selector.AddReceive(timerFuture, func(channel workflow.Channel, more bool) {
		// Timeout reached
		workflowComplete = true
		output.FinalStatus = models.StatusExpired
		output.Success = false
		output.Message = "Approval request expired - deadline reached"
	})

	// Wait for completion
	for !workflowComplete {
		selector.Select(ctx)
	}

	// ===== STEP 4: Execute Post-Approval Activities =====
	if output.FinalStatus == models.StatusApproved {
		logger.Info("Executing post-approval activities")
		
		// Execute the approved action
		err = workflow.ExecuteActivity(ctx, acts.ExecuteApprovedActionActivity, input.RequestID).Get(ctx, nil)
		if err != nil {
			logger.Warn("Post-approval action failed", "error", err)
			output.ActionsExecuted = append(output.ActionsExecuted, "ExecuteAction:FAILED")
		} else {
			output.ActionsExecuted = append(output.ActionsExecuted, "ExecuteAction:SUCCESS")
		}

		// Send approval notifications
		workflow.ExecuteActivity(ctx, acts.SendApprovalRequestNotificationsActivity, input.RequestID, "approved").Get(ctx, nil)
	} else if output.FinalStatus == models.StatusRejected {
		logger.Info("Executing post-rejection activities")
		
		// Send rejection notifications
		workflow.ExecuteActivity(ctx, acts.SendApprovalRequestNotificationsActivity, input.RequestID, "rejected").Get(ctx, nil)
	} else if output.FinalStatus == models.StatusExpired {
		logger.Info("Executing post-expiry activities")
		
		// Send expiry notifications
		workflow.ExecuteActivity(ctx, acts.SendApprovalRequestNotificationsActivity, input.RequestID, "expired").Get(ctx, nil)
	}

	// ===== STEP 5: Finalize Workflow =====
	err = workflow.ExecuteActivity(ctx, acts.FinalizeApprovalWorkflowActivity, input.RequestID, output.FinalStatus).Get(ctx, nil)
	if err != nil {
		logger.Warn("Finalization failed", "error", err)
	}

	output.ApprovalsReceived = approvalsReceived
	output.RejectionsReceived = rejectionsReceived
	output.ExecutionTime = workflow.Now(ctx).Sub(startTime)
	output.FinalDecision = string(output.FinalStatus)

	logger.Info("Approval workflow completed",
		"request_id", input.RequestID,
		"status", output.FinalStatus,
		"approvals", approvalsReceived,
		"rejections", rejectionsReceived,
		"execution_time", output.ExecutionTime,
	)

	return output, nil
}

// RecordActionInput represents input for recording an approval action
type RecordActionInput struct {
	RequestID   string  `json:"request_id"`
	ApproverID  string  `json:"approver_id"`
	ActionType  string  `json:"action_type"`
	Comments    string  `json:"comments"`
	DelegatedBy *string `json:"delegated_by,omitempty"`
}

// BatchApprovalWorkflow processes multiple approval requests sequentially
func BatchApprovalWorkflow(ctx workflow.Context, inputs []ApprovalWorkflowInput) (*BatchApprovalOutput, error) {
	logger := workflow.GetLogger(ctx)

	output := &BatchApprovalOutput{
		Results: make([]*ApprovalWorkflowOutput, 0, len(inputs)),
	}

	logger.Info("Starting batch approval workflow", "count", len(inputs))

	for i, input := range inputs {
		logger.Info("Processing batch item", "index", i, "request_id", input.RequestID)

		result, err := ExecuteApprovalWorkflow(ctx, input)
		if err != nil {
			logger.Error("Batch item failed", "index", i, "error", err)
			result.Success = false
			result.Message = "Workflow error: " + err.Error()
		}

		output.Results = append(output.Results, result)
		output.TotalProcessed++

		if result.Success {
			output.Successful++
		} else {
			output.Failed++
		}
	}

	logger.Info("Batch approval workflow completed",
		"total", output.TotalProcessed,
		"successful", output.Successful,
		"failed", output.Failed,
	)

	return output, nil
}

// BatchApprovalOutput represents the result of a batch workflow
type BatchApprovalOutput struct {
	Results        []*ApprovalWorkflowOutput `json:"results"`
	TotalProcessed int                       `json:"total_processed"`
	Successful     int                       `json:"successful"`
	Failed         int                       `json:"failed"`
}

// HumanInTheLoopApprovalWorkflow handles approvals requiring human verification
func HumanInTheLoopApprovalWorkflow(ctx workflow.Context, input ApprovalWorkflowInput) (*ApprovalWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)

	// Setup activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 24 * time.Hour,
		HeartbeatTimeout:    30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Create signal channel for approval
	approvalChannel := workflow.GetSignalChannel(ctx, "approval_decision")

	// Set workflow timeout for approval waiting
	ctx, cancel := workflow.WithTimeout(ctx, time.Duration(input.PolicyConfig.TimeoutHours)*time.Hour)
	defer cancel()

	logger.Info("Waiting for human approval", "request_id", input.RequestID)

	// Wait for approval signal
	var signal ApprovalSignal
	approved := approvalChannel.Receive(ctx, &signal)

	if !approved {
		return &ApprovalWorkflowOutput{
			RequestID:   input.RequestID,
			FinalStatus: models.StatusExpired,
			Success:     false,
			Message:     "Approval timeout - no response received",
		}, nil
	}

	if signal.ActionType != "APPROVE" {
		return &ApprovalWorkflowOutput{
			RequestID:   input.RequestID,
			FinalStatus: models.StatusRejected,
			Success:     false,
			Message:     "Approval rejected by " + signal.ApproverID,
		}, nil
	}

	logger.Info("Approval received from human", "approver_id", signal.ApproverID)

	// Execute the main approval workflow
	return ExecuteApprovalWorkflow(ctx, input)
}
