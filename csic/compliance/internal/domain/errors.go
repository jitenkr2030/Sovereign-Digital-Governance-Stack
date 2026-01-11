// Compliance Management Module - Custom Errors
// Error types and utilities for compliance operations

package domain

import (
	"errors"
	"fmt"
)

// Common error types
var (
	// Entity errors
	ErrEntityNotFound       = errors.New("entity not found")
	ErrEntityAlreadyExists  = errors.New("entity already exists")
	ErrEntityInactive       = errors.New("entity is not active")

	// License errors
	ErrLicenseNotFound      = errors.New("license not found")
	ErrLicenseInactive      = errors.New("license is not active")
	ErrLicenseExpired       = errors.New("license has expired")
	ErrLicenseSuspended     = errors.New("license is suspended")
	ErrLicenseRevoked       = errors.New("license has been revoked")
	ErrDuplicateLicense     = errors.New("entity already has an active license of this type")

	// Obligation errors
	ErrObligationNotFound   = errors.New("obligation not found")
	ErrObligationOverdue    = errors.New("obligation is overdue")

	// Violation errors
	ErrViolationNotFound    = errors.New("violation not found")
	ErrPenaltyNotFound      = errors.New("penalty not found")

	// State errors
	ErrInvalidStateTransition = errors.New("invalid state transition")

	// Validation errors
	ErrValidationError = errors.New("validation error")

	// Authorization errors
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrInsufficientRole   = errors.New("insufficient permissions")

	// Audit errors
	ErrAuditLogFailed     = errors.New("audit log operation failed")
)

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// ErrValidationError creates a validation error (for simple cases)
func ErrValidationError(message string) error {
	return &ValidationError{
		Message: message,
	}
}

// StateTransitionError represents an invalid state transition
type StateTransitionError struct {
	Entity    string
	Current   string
	Attempted string
	Reason    string
}

func (e *StateTransitionError) Error() string {
	return fmt.Sprintf("invalid state transition: %s from %s to %s - %s",
		e.Entity, e.Current, e.Attempted, e.Reason)
}

// ErrInvalidStateTransition creates a state transition error
func ErrInvalidStateTransition(entity, current, attempted string) error {
	return &StateTransitionError{
		Entity:    entity,
		Current:   string(current),
		Attempted: string(attempted),
		Reason:    "transition not allowed",
	}
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	ResourceType string
	ResourceID   string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.ResourceType, e.ResourceID)
}

// ErrNotFound creates a not found error
func ErrNotFound(resourceType, resourceID string) error {
	return &NotFoundError{
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
}

// ConflictError represents a resource conflict error
type ConflictError struct {
	ResourceType string
	Message      string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict: %s - %s", e.ResourceType, e.Message)
}

// ErrConflict creates a conflict error
func ErrConflict(resourceType, message string) error {
	return &ConflictError{
		ResourceType: resourceType,
		Message:      message,
	}
}

// DomainError represents a domain-specific error
type DomainError struct {
	Code    string
	Message string
	Details interface{}
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("domain error [%s]: %s", e.Code, e.Message)
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string, details interface{}) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Error codes for compliance module
const (
	ErrCodeEntityNotFound       = "ENTITY_NOT_FOUND"
	ErrCodeLicenseNotFound      = "LICENSE_NOT_FOUND"
	ErrCodeInvalidTransition    = "INVALID_STATE_TRANSITION"
	ErrCodeValidationFailed     = "VALIDATION_FAILED"
	ErrCodeDuplicateEntry       = "DUPLICATE_ENTRY"
	ErrCodeUnauthorized         = "UNAUTHORIZED"
	ErrCodeAuditFailed          = "AUDIT_FAILED"
	ErrCodeDatabaseError        = "DATABASE_ERROR"
	ErrCodeExternalServiceError = "EXTERNAL_SERVICE_ERROR"
)
