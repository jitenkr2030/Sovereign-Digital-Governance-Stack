package domain

import "errors"

// Common error definitions for the risk engine domain

var (
	// Rule errors
	ErrRuleNotFound       = errors.New("risk rule not found")
	ErrRuleAlreadyExists  = errors.New("risk rule already exists")
	ErrRuleInvalid        = errors.New("invalid risk rule configuration")
	ErrRuleDisabled       = errors.New("risk rule is disabled")

	// Alert errors
	ErrAlertNotFound      = errors.New("risk alert not found")
	ErrAlertInvalid       = errors.New("invalid alert configuration")
	ErrAlertAlreadyExists = errors.New("alert already exists")

	// Assessment errors
	ErrAssessmentFailed   = errors.New("risk assessment failed")
	ErrInvalidInput       = errors.New("invalid input for risk assessment")

	// Exposure errors
	ErrExposureNotFound   = errors.New("exposure not found")
	ErrExposureExceeded   = errors.New("exposure limit exceeded")
	ErrInvalidExposure    = errors.New("invalid exposure data")

	// Profile errors
	ErrProfileNotFound    = errors.New("risk profile not found")
	ErrProfileInvalid     = errors.New("invalid risk profile")

	// Threshold errors
	ErrThresholdNotFound  = errors.New("threshold configuration not found")
	ErrThresholdInvalid   = errors.New("invalid threshold configuration")

	// Evaluation errors
	ErrEvaluationTimeout  = errors.New("risk evaluation timed out")
	ErrNoRulesMatched     = errors.New("no rules matched the evaluation criteria")

	// Compliance errors
	ErrComplianceCheckFailed = errors.New("compliance check failed")
	ErrNonCompliant          = errors.New("entity is non-compliant")

	// Repository errors
	ErrRepositoryNotFound   = errors.New("record not found in repository")
	ErrRepositoryConflict   = errors.New("record conflict in repository")
	ErrRepositoryUnavailable = errors.New("repository unavailable")
)
