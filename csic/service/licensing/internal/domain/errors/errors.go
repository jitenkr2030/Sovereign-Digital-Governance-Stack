package errors

import "errors"

// License-related errors
var (
	ErrLicenseNotFound             = errors.New("license not found")
	ErrLicenseAlreadyExists        = errors.New("license already exists")
	ErrInvalidStatusTransition     = errors.New("invalid license status transition")
	ErrLicenseExpired              = errors.New("license has expired")
	ErrLicenseSuspended            = errors.New("license is suspended")
	ErrLicenseRevoked              = errors.New("license has been revoked")
	ErrInvalidJurisdiction         = errors.New("invalid jurisdiction")
	ErrInvalidLicenseType          = errors.New("invalid license type")
	ErrMissingRequiredField        = errors.New("required field is missing")
	ErrLicenseNumberConflict       = errors.New("license number already exists")
	ErrExchangeAlreadyLicensed     = errors.New("exchange already has an active license for this jurisdiction")
)

// Violation-related errors
var (
	ErrViolationNotFound   = errors.New("violation not found")
	ErrInvalidViolationID  = errors.New("invalid violation ID")
	ErrCannotResolveOpenViolation = errors.New("cannot resolve an open violation without resolution details")
)

// Report-related errors
var (
	ErrReportNotFound      = errors.New("report not found")
	ErrReportGeneration    = errors.New("report generation failed")
	ErrInvalidReportType   = errors.New("invalid report type")
	ErrReportNotReady      = errors.New("report is not ready for download")
	ErrReportTemplateNotFound = errors.New("report template not found")
)

// Application-related errors
var (
	ErrApplicationNotFound    = errors.New("application not found")
	ErrApplicationIncomplete  = errors.New("application is incomplete")
	ErrInvalidApplicationStatus = errors.New("invalid application status transition")
)

// Compliance-related errors
var (
	ErrComplianceCheckFailed = errors.New("compliance check failed")
	ErrInvalidComplianceScore = errors.New("invalid compliance score")
	ErrComplianceThresholdExceeded = errors.New("compliance threshold exceeded")
)

// Document-related errors
var (
	ErrDocumentNotFound    = errors.New("document not found")
	ErrDocumentUploadFailed = errors.New("document upload failed")
	ErrInvalidDocumentType = errors.New("invalid document type")
	ErrDocumentTooLarge    = errors.New("document exceeds maximum size")
)

// Authorization errors
var (
	ErrUnauthorizedAccess = errors.New("unauthorized access")
	ErrInsufficientPermission = errors.New("insufficient permission")
)

// Validation errors
var (
	ErrInvalidDateRange    = errors.New("invalid date range")
	ErrInvalidUUID         = errors.New("invalid UUID format")
	ErrValueOutOfRange     = errors.New("value is out of allowed range")
)

// CustomError represents a structured error response
type CustomError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *CustomError) Error() string {
	return e.Message
}

// NewCustomError creates a new custom error
func NewCustomError(code, message, details string) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Error codes
const (
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeInternal           = "INTERNAL_ERROR"
	ErrCodeBadRequest         = "BAD_REQUEST"
	ErrCodeTooManyRequests    = "TOO_MANY_REQUESTS"
)
