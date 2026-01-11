package schema

import (
	"context"
	"encoding/json"
	"time"
)

// ValidationResult represents the outcome of schema validation
type ValidationResult struct {
	Valid     bool                `json:"valid"`
	Errors    []ValidationError   `json:"errors,omitempty"`
	Timestamp time.Time           `json:"timestamp"`
	Duration  time.Duration       `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field       string `json:"field"`
	Message     string `json:"message"`
	Expected    string `json:"expected,omitempty"`
	Actual      string `json:"actual,omitempty"`
	ErrorType   string `json:"error_type"`
	Constraint  string `json:"constraint,omitempty"`
}

// Validator defines the interface for schema validators
type Validator interface {
	// Validate validates the provided data against the schema
	Validate(ctx context.Context, data json.RawMessage) (*ValidationResult, error)

	// ValidateWithContext validates with additional context
	ValidateWithContext(ctx context.Context, data json.RawMessage, metadata map[string]interface{}) (*ValidationResult, error)

	// GetSchemaID returns the schema identifier
	GetSchemaID() string

	// GetSchemaVersion returns the schema version
	GetSchemaVersion() string

	// GetSchemaType returns the type of schema
	GetSchemaType() SchemaType
}

// ValidationMode defines how validation should be performed
type ValidationMode string

const (
	ValidationModeStrict   ValidationMode = "strict"   // Fail on any validation error
	ValidationModePermissive ValidationMode = "permissive" // Continue collecting errors
	ValidationModeAudit    ValidationMode = "audit"    // Log errors but don't fail
)

// ValidatorConfig contains configuration for validators
type ValidatorConfig struct {
	Mode            ValidationMode       `json:"mode"`
	CacheResults    bool                 `json:"cache_results"`
	CacheTTL        time.Duration        `json:"cache_ttl"`
	CustomRules     []CustomRule         `json:"custom_rules,omitempty"`
	RequiredFields  []string             `json:"required_fields,omitempty"`
	FieldOverrides  map[string]FieldConfig `json:"field_overrides,omitempty"`
}

// FieldConfig contains configuration for individual fields
type FieldConfig struct {
	Required    bool     `json:"required"`
	Type        string   `json:"type,omitempty"`
	MinLength   *int     `json:"min_length,omitempty"`
	MaxLength   *int     `json:"max_length,omitempty"`
	Minimum     *float64 `json:"minimum,omitempty"`
	Maximum     *float64 `json:"maximum,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// CustomRule defines a custom validation rule
type CustomRule struct {
	Name       string            `json:"name"`
	Field      string            `json:"field"`
	Predicate  string            `json:"predicate"`
	Message    string            `json:"message"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

// ValidationContext provides context for validation operations
type ValidationContext struct {
	Context     context.Context
	SourceID    string
	Operation   string
	Timestamp   time.Time
	Metadata    map[string]interface{}
	SessionID   string
	RequestID   string
}

// NewValidationContext creates a new validation context
func NewValidationContext(ctx context.Context, sourceID, operation string) *ValidationContext {
	return &ValidationContext{
		Context:   ctx,
		SourceID:  sourceID,
		Operation: operation,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// WithMetadata adds metadata to the validation context
func (vc *ValidationContext) WithMetadata(key string, value interface{}) *ValidationContext {
	vc.Metadata[key] = value
	return vc
}

// WithSession adds session information to the validation context
func (vc *ValidationContext) WithSession(sessionID string) *ValidationContext {
	vc.SessionID = sessionID
	return vc
}

// WithRequest adds request information to the validation context
func (vc *ValidationContext) WithRequest(requestID string) *ValidationContext {
	vc.RequestID = requestID
	return vc
}

// ValidationStats contains statistics about validation operations
type ValidationStats struct {
	TotalValidations  int64             `json:"total_validations"`
	TotalErrors       int64             `json:"total_errors"`
	ErrorsByField     map[string]int64  `json:"errors_by_field"`
	ErrorsByType      map[string]int64  `json:"errors_by_type"`
	LastValidation    time.Time         `json:"last_validation"`
	AverageDuration   time.Duration     `json:"average_duration"`
	P99Duration       time.Duration     `json:"p99_duration"`
}

// NewValidationStats creates a new validation stats instance
func NewValidationStats() *ValidationStats {
	return &ValidationStats{
		ErrorsByField: make(map[string]int64),
		ErrorsByType:  make(map[string]int64),
	}
}

// RecordError records a validation error for statistics
func (vs *ValidationStats) RecordError(err ValidationError) {
	vs.TotalErrors++
	vs.ErrorsByField[err.Field]++
	vs.ErrorsByType[err.ErrorType]++
}

// RecordSuccess records a successful validation
func (vs *ValidationStats) RecordSuccess(duration time.Duration) {
	vs.TotalValidations++
	vs.LastValidation = time.Now()
	
	// Update running average
	total := vs.TotalValidations + vs.TotalErrors
	if total == 1 {
		vs.AverageDuration = duration
	} else {
		vs.AverageDuration = time.Duration(
			(float64(vs.AverageDuration)*(float64(total-1)) + float64(duration)) / float64(total),
		)
	}
}
