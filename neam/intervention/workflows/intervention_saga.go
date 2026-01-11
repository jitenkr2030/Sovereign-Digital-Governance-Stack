package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"neam-platform/intervention/activities"
)

// InterventionWorkflowInput represents the input for an intervention workflow
type InterventionWorkflowInput struct {
	InterventionID string            `json:"intervention_id"`
	PolicyID       string            `json:"policy_id"`
	PolicyType     string            `json:"policy_type"`
	TargetEntity   string            `json:"target_entity"`
	TargetRegion   string            `json:"target_region"`
	Amount         float64           `json:"amount"`
	Currency       string            `json:"currency"`
	Parameters     map[string]string `json:"parameters"`
	Priority       string            `json:"priority"`
	CreatedBy      string            `json:"created_by"`
}

// InterventionWorkflowOutput represents the output of an intervention workflow
type InterventionWorkflowOutput struct {
	InterventionID   string                      `json:"intervention_id"`
	FinalStatus      string                      `json:"final_status"`
	Success          bool                        `json:"success"`
	Message          string                      `json:"message"`
	ExecutionTime    time.Duration               `json:"execution_time"`
	ActivitiesExecuted []string                  `json:"activities_executed"`
	CompensationsExecuted []string               `json:"compensations_executed"`
	Result           *activities.InterventionResult `json:"result"`
}

// InterventionWorkflowTimeoutConfig defines timeout settings
type InterventionWorkflowTimeoutConfig struct {
	TotalTimeout       time.Duration
	ActivityTimeout    time.Duration
	DecisionTimeout    time.Duration
	HeartbeatTimeout   time.Duration
}

// DefaultTimeoutConfig returns default timeout settings
func DefaultTimeoutConfig() InterventionWorkflowTimeoutConfig {
	return InterventionWorkflowTimeoutConfig{
		TotalTimeout:     30 * time.Minute,
		ActivityTimeout:  5 * time.Minute,
		DecisionTimeout:  10 * time.Second,
		HeartbeatTimeout: 30 * time.Second,
	}
}

// ExecuteInterventionWorkflow is the main saga workflow for intervention execution
// This implements the Saga pattern with compensation transactions
func ExecuteInterventionWorkflow(ctx workflow.Context, input InterventionWorkflowInput) (*InterventionWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)

	// Setup activity options with timeouts and retry policy
	timeoutConfig := DefaultTimeoutConfig()
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: timeoutConfig.ActivityTimeout,
		HeartbeatTimeout:    timeoutConfig.HeartbeatTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 5,
			MaximumAttempts:    3,
			NonRetryableErrorTypes: []string{
				"ValidationError",
				"PolicyNotFound",
				"InsufficientBudget",
			},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Initialize output
	output := &InterventionWorkflowOutput{
		InterventionID:     input.InterventionID,
		ActivitiesExecuted: make([]string, 0),
		CompensationsExecuted: make([]string, 0),
	}

	startTime := workflow.Now(ctx)

	// Activity references
	acts := &activities.InterventionActivities{}

	// Compensation functions storage (LIFO order for execution)
	var compensations []func() error

	// Helper function to execute compensation
	executeCompensation := func(name string, compensation func() error) {
		if err := compensation(); err != nil {
			logger.Error("Compensation failed", "activity", name, "error", err)
			output.CompensationsExecuted = append(output.CompensationsExecuted, name+":FAILED")
		} else {
			output.CompensationsExecuted = append(output.CompensationsExecuted, name+":SUCCESS")
		}
	}

	// Defer cleanup to handle any panics
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Workflow panicked, executing compensations", "panic", r)
			for i := len(compensations) - 1; i >= 0; i-- {
				executeCompensation("panic-recovery", compensations[i])
			}
			output.FinalStatus = "PANIC"
			output.Success = false
			output.Message = "Workflow panicked during execution"
		}
	}()

	logger.Info("Starting intervention workflow",
		"intervention_id", input.InterventionID,
		"policy_id", input.PolicyID,
		"target_entity", input.TargetEntity,
	)

	// ===== STEP 1: Validate Policy =====
	logger.Info("Executing activity: ValidatePolicy")
	err := workflow.ExecuteActivity(ctx, acts.ValidatePolicyActivity, input.PolicyID).Get(ctx, nil)
	if err != nil {
		output.FinalStatus = "VALIDATION_FAILED"
		output.Message = "Policy validation failed: " + err.Error()
		output.Success = false
		return output, nil
	}
	output.ActivitiesExecuted = append(output.ActivitiesExecuted, "ValidatePolicy:SUCCESS")

	// ===== STEP 2: Get Policy Details =====
	logger.Info("Executing activity: GetPolicyDetails")
	policyDetails, err := workflow.ExecuteActivity(ctx, acts.GetPolicyDetailsActivity, input.PolicyID).Get(ctx, nil)
	if err != nil {
		output.FinalStatus = "POLICY_FETCH_FAILED"
		output.Message = "Failed to fetch policy details: " + err.Error()
		output.Success = false
		return output, nil
	}
	output.ActivitiesExecuted = append(output.ActivitiesExecuted, "GetPolicyDetails:SUCCESS")

	// Validate amount against policy limits
	if input.Amount > policyDetails.MaxAmount {
		output.FinalStatus = "AMOUNT_EXCEEDS_LIMIT"
		output.Message = "Intervention amount exceeds policy maximum"
		output.Success = false
		return output, nil
	}

	// ===== STEP 3: Reserve Budget (with compensation) =====
	logger.Info("Executing activity: ReserveBudget")
	err = workflow.ExecuteActivity(ctx, acts.ReserveBudgetActivity, input.InterventionID, input.Amount, input.Currency).Get(ctx, nil)
	if err != nil {
		output.FinalStatus = "BUDGET_RESERVATION_FAILED"
		output.Message = "Failed to reserve budget: " + err.Error()
		output.Success = false
		return output, nil
	}
	output.ActivitiesExecuted = append(output.ActivitiesExecuted, "ReserveBudget:SUCCESS")

	// Add compensation for budget reservation
	compensations = append(compensations, func() error {
		return acts.ReleaseBudgetActivity(ctx, input.InterventionID)
	})

	// ===== STEP 4: Execute Intervention =====
	logger.Info("Executing activity: ExecuteIntervention")

	// Create activity request
	request := activities.InterventionRequest{
		InterventionID: input.InterventionID,
		PolicyID:       input.PolicyID,
		PolicyType:     input.PolicyType,
		TargetEntity:   input.TargetEntity,
		TargetRegion:   input.TargetRegion,
		Amount:         input.Amount,
		Currency:       input.Currency,
		Parameters:     input.Parameters,
		Priority:       input.Priority,
		CreatedBy:      input.CreatedBy,
		CreatedAt:      time.Now(),
	}

	var result *activities.InterventionResult
	err = workflow.ExecuteActivity(ctx, acts.ExecuteInterventionActivity, request).Get(ctx, &result)
	if err != nil {
		output.FinalStatus = "EXECUTION_FAILED"
		output.Message = "Intervention execution failed: " + err.Error()
		output.Success = false

		// Execute compensations in reverse order
		for i := len(compensations) - 1; i >= 0; i-- {
			executeCompensation("execution-failure-recovery", compensations[i])
		}

		return output, nil
	}
	output.ActivitiesExecuted = append(output.ActivitiesExecuted, "ExecuteIntervention:SUCCESS")
	output.Result = result

	if !result.Success {
		output.FinalStatus = "INTERVENTION_FAILED"
		output.Message = result.Message
		output.Success = false

		// Execute compensations
		for i := len(compensations) - 1; i >= 0; i-- {
			executeCompensation("result-failure-recovery", compensations[i])
		}

		return output, nil
	}

	// ===== STEP 5: Update Feature Store =====
	logger.Info("Executing activity: UpdateFeatureStore")
	err = workflow.ExecuteActivity(ctx, acts.UpdateFeatureStoreActivity, input.InterventionID, "COMPLETED", result).Get(ctx, nil)
	if err != nil {
		logger.Warn("Feature store update failed, but intervention was successful", "error", err)
		output.ActivitiesExecuted = append(output.ActivitiesExecuted, "UpdateFeatureStore:FAILED")
	} else {
		output.ActivitiesExecuted = append(output.ActivitiesExecuted, "UpdateFeatureStore:SUCCESS")
	}

	// ===== STEP 6: Notify Stakeholders =====
	logger.Info("Executing activity: NotifyStakeholders")
	notificationMessage := fmt.Sprintf("Intervention %s completed successfully", input.InterventionID)
	err = workflow.ExecuteActivity(ctx, acts.NotifyStakeholdersActivity, input.InterventionID, "SUCCESS", notificationMessage).Get(ctx, nil)
	if err != nil {
		logger.Warn("Notification failed", "error", err)
		output.ActivitiesExecuted = append(output.ActivitiesExecuted, "NotifyStakeholders:FAILED")
	} else {
		output.ActivitiesExecuted = append(output.ActivitiesExecuted, "NotifyStakeholders:SUCCESS")
	}

	// ===== STEP 7: Record Outcome for Analytics =====
logger.Info("Executing activity: RecordOutcome")
	err = workflow.ExecuteActivity(ctx, acts.RecordOutcomeActivity, input.InterventionID, result).Get(ctx, nil)
	if err != nil {
		logger.Warn("Outcome recording failed", "error", err)
		output.ActivitiesExecuted = append(output.ActivitiesExecuted, "RecordOutcome:FAILED")
	} else {
		output.ActivitiesExecuted = append(output.ActivitiesExecuted, "RecordOutcome:SUCCESS")
	}

	// Mark workflow as successful
	output.FinalStatus = "COMPLETED"
	output.Success = true
	output.Message = "Intervention workflow completed successfully"
	output.ExecutionTime = workflow.Now(ctx).Sub(startTime)

	logger.Info("Intervention workflow completed",
		"intervention_id", input.InterventionID,
		"status", output.FinalStatus,
		"execution_time", output.ExecutionTime,
	)

	return output, nil
}

// Signal-based workflow for human-in-the-loop approval
type ApprovalSignal struct {
	Approved   bool
	ApprovedBy string
	Comment    string
	Timestamp  time.Time
}

// InterventionApprovalWorkflow handles interventions requiring human approval
func InterventionApprovalWorkflow(ctx workflow.Context, input InterventionWorkflowInput) (*InterventionWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)

	// Setup activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	acts := &activities.InterventionActivities{}

	// Create signal channel for approval
	approvalChannel := workflow.GetSignalChannel(ctx, "approval")

	// Set workflow timeout for approval waiting
	ctx, cancel := workflow.WithTimeout(ctx, 24*time.Hour)
	defer cancel()

	logger.Info("Waiting for human approval", "intervention_id", input.InterventionID)

	// Wait for approval signal
	var signal ApprovalSignal
	approved := approvalChannel.Receive(ctx, &signal)

	if !approved {
		// Timeout or signal channel closed
		return &InterventionWorkflowOutput{
			InterventionID: input.InterventionID,
			FinalStatus:    "APPROVAL_TIMEOUT",
			Success:        false,
			Message:        "Approval timeout - no response received",
		}, nil
	}

	if !signal.Approved {
		return &InterventionWorkflowOutput{
			InterventionID: input.InterventionID,
			FinalStatus:    "REJECTED",
			Success:        false,
			Message:        "Intervention rejected by " + signal.ApprovedBy + ": " + signal.Comment,
		}, nil
	}

	logger.Info("Intervention approved", "by", signal.ApprovedBy)

	// Execute the main intervention workflow
	return ExecuteInterventionWorkflow(ctx, input)
}

// BatchInterventionWorkflow processes multiple interventions sequentially
func BatchInterventionWorkflow(ctx workflow.Context, inputs []InterventionWorkflowInput) (*BatchWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)

	output := &BatchWorkflowOutput{
		Results: make([]*InterventionWorkflowOutput, 0, len(inputs)),
	}

	logger.Info("Starting batch intervention workflow", "count", len(inputs))

	for i, input := range inputs {
		logger.Info("Processing batch item", "index", i, "intervention_id", input.InterventionID)

		result, err := ExecuteInterventionWorkflow(ctx, input)
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

	logger.Info("Batch workflow completed",
		"total", output.TotalProcessed,
		"successful", output.Successful,
		"failed", output.Failed,
	)

	return output, nil
}

// BatchWorkflowOutput represents the result of a batch workflow
type BatchWorkflowOutput struct {
	Results       []*InterventionWorkflowOutput `json:"results"`
	TotalProcessed int                          `json:"total_processed"`
	Successful    int                          `json:"successful"`
	Failed        int                          `json:"failed"`
}
