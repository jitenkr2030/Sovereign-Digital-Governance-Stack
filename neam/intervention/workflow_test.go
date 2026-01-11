package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"neam-platform/intervention/activities"
	"neam-platform/intervention/workflows"
)

func TestInterventionWorkflow_ValidationError(t *testing.T) {
	s := testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()

	// Register workflow
	env.RegisterWorkflow(workflows.ExecuteInterventionWorkflow)

	// Create test input
	input := workflows.InterventionWorkflowInput{
		InterventionID: "test-intervention-001",
		PolicyID:       "INVALID",
		PolicyType:     "economic_stabilization",
		TargetEntity:   "ENT-001",
		TargetRegion:   "NORTH",
		Amount:         100000,
		Currency:       "USD",
		Parameters:     make(map[string]string),
		Priority:       "high",
		CreatedBy:      "test-user",
	}

	// Setup activity mocks
	acts := &mockInterventionActivities{}
	env.RegisterActivity(acts)

	// Run workflow
	env.ExecuteWorkflow(workflows.ExecuteInterventionWorkflow, input)

	// Verify workflow completed with validation error
	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())

	var result *workflows.InterventionWorkflowOutput
	env.GetWorkflowResult(&result)

	assert.NotNil(t, result)
	assert.Equal(t, "VALIDATION_FAILED", result.FinalStatus)
	assert.False(t, result.Success)
	assert.Contains(t, result.ActivitiesExecuted, "ValidatePolicy:SUCCESS")
}

func TestInterventionWorkflow_Success(t *testing.T) {
	s := testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(workflows.ExecuteInterventionWorkflow)

	input := workflows.InterventionWorkflowInput{
		InterventionID: "test-intervention-002",
		PolicyID:       "POLICY-001",
		PolicyType:     "economic_stabilization",
		TargetEntity:   "ENT-002",
		TargetRegion:   "SOUTH",
		Amount:         50000,
		Currency:       "USD",
		Parameters:     make(map[string]string),
		Priority:       "medium",
		CreatedBy:      "test-user",
	}

	acts := &mockInterventionActivities{
		shouldFail: false,
	}
	env.RegisterActivity(acts)

	env.ExecuteWorkflow(workflows.ExecuteInterventionWorkflow, input)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())

	var result *workflows.InterventionWorkflowOutput
	env.GetWorkflowResult(&result)

	assert.NotNil(t, result)
	assert.Equal(t, "COMPLETED", result.FinalStatus)
	assert.True(t, result.Success)
	assert.Contains(t, result.ActivitiesExecuted, "ValidatePolicy:SUCCESS")
	assert.Contains(t, result.ActivitiesExecuted, "ReserveBudget:SUCCESS")
	assert.Contains(t, result.ActivitiesExecuted, "ExecuteIntervention:SUCCESS")
}

func TestInterventionWorkflow_BudgetReservationFails(t *testing.T) {
	s := testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(workflows.ExecuteInterventionWorkflow)

	input := workflows.InterventionWorkflowInput{
		InterventionID: "test-intervention-003",
		PolicyID:       "POLICY-001",
		PolicyType:     "economic_stabilization",
		TargetEntity:   "ENT-003",
		TargetRegion:   "EAST",
		Amount:         50000,
		Currency:       "USD",
		Parameters:     make(map[string]string),
		Priority:       "low",
		CreatedBy:      "test-user",
	}

	acts := &mockInterventionActivities{
		failAtStep: "ReserveBudget",
	}
	env.RegisterActivity(acts)

	env.ExecuteWorkflow(workflows.ExecuteInterventionWorkflow, input)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())

	var result *workflows.InterventionWorkflowOutput
	env.GetWorkflowResult(&result)

	assert.NotNil(t, result)
	assert.Equal(t, "BUDGET_RESERVATION_FAILED", result.FinalStatus)
	assert.False(t, result.Success)
}

func TestInterventionWorkflow_CompensationExecuted(t *testing.T) {
	s := testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(workflows.ExecuteInterventionWorkflow)

	input := workflows.InterventionWorkflowInput{
		InterventionID: "test-intervention-004",
		PolicyID:       "POLICY-001",
		PolicyType:     "economic_stabilization",
		TargetEntity:   "ENT-004",
		TargetRegion:   "WEST",
		Amount:         50000,
		Currency:       "USD",
		Parameters:     make(map[string]string),
		Priority:       "high",
		CreatedBy:      "test-user",
	}

	acts := &mockInterventionActivities{
		shouldFail:   true,
		failAtStep:   "ExecuteIntervention",
		compensationExecuted: false,
	}
	env.RegisterActivity(acts)

	env.ExecuteWorkflow(workflows.ExecuteInterventionWorkflow, input)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())

	var result *workflows.InterventionWorkflowOutput
	env.GetWorkflowResult(&result)

	assert.NotNil(t, result)
	assert.Equal(t, "EXECUTION_FAILED", result.FinalStatus)
	assert.False(t, result.Success)
	assert.True(t, acts.compensationExecuted)
}

func TestBatchInterventionWorkflow(t *testing.T) {
	s := testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(workflows.BatchInterventionWorkflow)

	inputs := []workflows.InterventionWorkflowInput{
		{
			InterventionID: "batch-001",
			PolicyID:       "POLICY-001",
			TargetEntity:   "ENT-001",
			Amount:         10000,
		},
		{
			InterventionID: "batch-002",
			PolicyID:       "POLICY-001",
			TargetEntity:   "ENT-002",
			Amount:         20000,
		},
		{
			InterventionID: "batch-003",
			PolicyID:       "INVALID", // Will fail
			TargetEntity:   "ENT-003",
			Amount:         30000,
		},
	}

	acts := &mockInterventionActivities{shouldFail: false}
	env.RegisterActivity(acts)

	env.ExecuteWorkflow(workflows.BatchInterventionWorkflow, inputs)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())

	var result *workflows.BatchWorkflowOutput
	env.GetWorkflowResult(&result)

	assert.NotNil(t, result)
	assert.Equal(t, 3, result.TotalProcessed)
	assert.Equal(t, 2, result.Successful)
	assert.Equal(t, 1, result.Failed)
}

func TestWorkflowTimeouts(t *testing.T) {
	s := testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(workflows.ExecuteInterventionWorkflow)

	input := workflows.InterventionWorkflowInput{
		InterventionID: "timeout-test",
		PolicyID:       "POLICY-001",
		TargetEntity:   "ENT-001",
		Amount:         50000,
	}

	acts := &mockInterventionActivities{shouldFail: false}
	env.RegisterActivity(acts)

	// Set workflow timeout
	env.SetWorkflowTimeout(1 * time.Minute)

	env.ExecuteWorkflow(workflows.ExecuteInterventionWorkflow, input)

	assert.True(t, env.IsWorkflowCompleted())
}

// Mock activities for testing
type mockInterventionActivities struct {
	shouldFail           bool
	failAtStep           string
	compensationExecuted bool
}

func (m *mockInterventionActivities) ValidatePolicyActivity(ctx context.Context, policyID string) error {
	if policyID == "INVALID" {
		return &activities.PolicyValidationError{PolicyID: policyID}
	}
	return nil
}

func (m *mockInterventionActivities) ReserveBudgetActivity(ctx context.Context, interventionID string, amount float64, currency string) error {
	if m.shouldFail && m.failAtStep == "ReserveBudget" {
		return &activities.BudgetReservationError{InterventionID: interventionID}
	}
	return nil
}

func (m *mockInterventionActivities) ReleaseBudgetActivity(ctx context.Context, interventionID string) error {
	m.compensationExecuted = true
	return nil
}

func (m *mockInterventionActivities) ExecuteInterventionActivity(ctx context.Context, request activities.InterventionRequest) (*activities.InterventionResult, error) {
	if m.shouldFail && m.failAtStep == "ExecuteIntervention" {
		return nil, &activities.InterventionExecutionError{InterventionID: request.InterventionID}
	}
	return &activities.InterventionResult{
		Success:    true,
		Message:    "Success",
		Outcome:    "SUCCESS",
		ExecutedAt: time.Now(),
	}, nil
}

func (m *mockInterventionActivities) UpdateFeatureStoreActivity(ctx context.Context, interventionID string, status string, result *activities.InterventionResult) error {
	return nil
}

func (m *mockInterventionActivities) NotifyStakeholdersActivity(ctx context.Context, interventionID string, notificationType string, message string) error {
	return nil
}

func (m *mockInterventionActivities) RecordOutcomeActivity(ctx context.Context, interventionID string, result *activities.InterventionResult) error {
	return nil
}

func (m *mockInterventionActivities) GetPolicyDetailsActivity(ctx context.Context, policyID string) (*activities.PolicyDetails, error) {
	return &activities.PolicyDetails{
		PolicyID:   policyID,
		MaxAmount:  1000000,
		Currency:   "USD",
		ValidFrom:  time.Now().Add(-24 * time.Hour),
		ValidUntil: time.Now().Add(30 * 24 * time.Hour),
	}, nil
}

func (m *mockInterventionActivities) HealthCheckActivity(ctx context.Context) error {
	return nil
}

// Test helper functions
func TestInterventionInputValidation(t *testing.T) {
	testCases := []struct {
		name        string
		input       workflows.InterventionWorkflowInput
		shouldError bool
	}{
		{
			name: "valid input",
			input: workflows.InterventionWorkflowInput{
				InterventionID: "INT-001",
				PolicyID:       "POL-001",
				TargetEntity:   "ENT-001",
				Amount:         10000,
			},
			shouldError: false,
		},
		{
			name: "empty intervention ID",
			input: workflows.InterventionWorkflowInput{
				PolicyID:     "POL-001",
				TargetEntity: "ENT-001",
				Amount:       10000,
			},
			shouldError: true,
		},
		{
			name: "zero amount",
			input: workflows.InterventionWorkflowInput{
				InterventionID: "INT-001",
				PolicyID:       "POL-001",
				TargetEntity:   "ENT-001",
				Amount:         0,
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Validate input
			err := validateInput(&tc.input)
			if tc.shouldError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func validateInput(input *workflows.InterventionWorkflowInput) error {
	if input.InterventionID == "" {
		return &ValidationError{Field: "InterventionID", Message: "cannot be empty"}
	}
	if input.PolicyID == "" {
		return &ValidationError{Field: "PolicyID", Message: "cannot be empty"}
	}
	if input.TargetEntity == "" {
		return &ValidationError{Field: "TargetEntity", Message: "cannot be empty"}
	}
	if input.Amount <= 0 {
		return &ValidationError{Field: "Amount", Message: "must be positive"}
	}
	return nil
}

// Custom error types for testing
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s", e.Field, e.Message)
}
