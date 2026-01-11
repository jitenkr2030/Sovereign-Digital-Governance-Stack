package activities

import "fmt"

// PolicyValidationError is returned when policy validation fails
type PolicyValidationError struct {
	PolicyID string
}

func (e *PolicyValidationError) Error() string {
	return fmt.Sprintf("policy validation failed for policy: %s", e.PolicyID)
}

// BudgetReservationError is returned when budget reservation fails
type BudgetReservationError struct {
	InterventionID string
	Amount         float64
}

func (e *BudgetReservationError) Error() string {
	return fmt.Sprintf("budget reservation failed for intervention: %s (amount: %.2f)", e.InterventionID, e.Amount)
}

// InterventionExecutionError is returned when intervention execution fails
type InterventionExecutionError struct {
	InterventionID string
	Reason         string
}

func (e *InterventionExecutionError) Error() string {
	return fmt.Sprintf("intervention execution failed: %s (reason: %s)", e.InterventionID, e.Reason)
}

// InsufficientBudgetError is returned when there's not enough budget
type InsufficientBudgetError struct {
	RequestedAmount float64
	AvailableAmount float64
}

func (e *InsufficientBudgetError) Error() string {
	return fmt.Sprintf("insufficient budget: requested %.2f but only %.2f available", e.RequestedAmount, e.AvailableAmount)
}

// PolicyNotFoundError is returned when a policy cannot be found
type PolicyNotFoundError struct {
	PolicyID string
}

func (e *PolicyNotFoundError) Error() string {
	return fmt.Sprintf("policy not found: %s", e.PolicyID)
}

// StateSyncError is returned when state synchronization fails
type StateSyncError struct {
	NodeID    string
	Operation string
}

func (e *StateSyncError) Error() string {
	return fmt.Sprintf("state synchronization failed for node %s during operation %s", e.NodeID, e.Operation)
}
