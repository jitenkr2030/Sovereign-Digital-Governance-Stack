package resilience

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string `json:"field"`
	Value   interface{} `json:"value"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// ValidationResult holds the result of schema validation
type ValidationResult struct {
	Valid       bool            `json:"valid"`
	Errors      []ValidationError `json:"errors,omitempty"`
	ValidatedAt time.Time       `json:"validated_at"`
}

// SchemaValidator provides JSON schema validation capabilities
type SchemaValidator struct {
	validate   *validator.Validate
	logger     *zap.Logger
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator(logger *zap.Logger) *SchemaValidator {
	return &SchemaValidator{
		validate: validator.New(),
		logger:   logger,
	}
}

// Validate validates a struct against its tags
func (v *SchemaValidator) Validate(ctx context.Context, structPtr interface{}) *ValidationResult {
	result := &ValidationResult{
		Valid:       true,
		ValidatedAt: time.Now(),
		Errors:      make([]ValidationError, 0),
	}

	err := v.validate.StructCtx(ctx, structPtr)
	if err != nil {
		result.Valid = false

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				ve := ValidationError{
					Field:   e.Field(),
					Value:   fmt.Sprintf("%v", e.Value()),
					Tag:     e.Tag(),
					Message: e.Error(),
				}
				result.Errors = append(result.Errors, ve)
			}
		}
	}

	return result
}

// ValidateJSON validates raw JSON data against a struct
func (v *SchemaValidator) ValidateJSON(ctx context.Context, data []byte, structPtr interface{}) *ValidationResult {
	result := &ValidationResult{
		Valid:       true,
		ValidatedAt: time.Now(),
		Errors:      make([]ValidationError, 0),
	}

	// First, unmarshal the JSON
	if err := json.Unmarshal(data, structPtr); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "root",
			Value:   string(data),
			Tag:     "json_parse",
			Message: fmt.Sprintf("JSON parse error: %v", err),
		})
		return result
	}

	// Then validate the struct
	if err := v.validate.StructCtx(ctx, structPtr); err != nil {
		result.Valid = false

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				ve := ValidationError{
					Field:   e.Field(),
					Value:   fmt.Sprintf("%v", e.Value()),
					Tag:     e.Tag(),
					Message: e.Error(),
				}
				result.Errors = append(result.Errors, ve)
			}
		}
	}

	return result
}

// TransportTelemetry represents a telemetry message for validation
type TransportTelemetry struct {
	VehicleID   string    `json:"vehicle_id" validate:"required,uuid4"`
	Timestamp   time.Time `json:"timestamp" validate:"required"`
	Latitude    float64   `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude   float64   `json:"longitude" validate:"required,min=-180,max=180"`
	Speed       float64   `json:"speed" validate:"omitempty,gte=0"`
	FuelLevel   float64   `json:"fuel_level" validate:"omitempty,min=0,max=100"`
	EngineTemp  float64   `json:"engine_temp" validate:"omitempty"`
	Alerts      []string  `json:"alerts,omitempty"`
}

// IndustrialTelemetry represents industrial sensor data for validation
type IndustrialTelemetry struct {
	NodeID      string    `json:"node_id" validate:"required"`
	AssetID     string    `json:"asset_id" validate:"required"`
	Timestamp   time.Time `json:"timestamp" validate:"required"`
	Value       float64   `json:"value" validate:"required"`
	Quality     string    `json:"quality" validate:"required,oneof=Good Bad Uncertain"`
	Unit        string    `json:"unit,omitempty"`
	DataType    string    `json:"data_type" validate:"required,oneof=Float Double Int Bool String"`
	ScanRate    int       `json:"scan_rate" validate:"omitempty,min=1"`
	AlarmState  string    `json:"alarm_state,omitempty" validate:"omitempty,oneof=Normal Warning Alarm"`
}

// EnergyTelemetry represents energy meter data for validation
type EnergyTelemetry struct {
	MeterID     string    `json:"meter_id" validate:"required"`
	Timestamp   time.Time `json:"timestamp" validate:"required"`
	ActivePower float64   `json:"active_power" validate:"required"`
	ReactivePower float64 `json:"reactive_power,omitempty"`
	Voltage     float64   `json:"voltage,omitempty" validate:"omitempty,min=0"`
	Current     float64   `json:"current,omitempty" validate:"omitempty,min=0"`
	Frequency   float64   `json:"frequency,omitempty" validate:"omitempty,min=45,max=65"`
	PowerFactor float64   `json:"power_factor,omitempty" validate:"omitempty,min=-1,max=1"`
	TotalEnergy float64   `json:"total_energy,omitempty"`
	Quality     string    `json:"quality" validate:"required,oneof=Good Bad Uncertain"`
}

// ErrorInfo represents structured error information
type ErrorInfo struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
	Service     string                 `json:"service"`
	Operation   string                 `json:"operation"`
}

// ResilientError is a structured error type for resilience operations
type ResilientError struct {
	Code          string
	Message       string
	CorrelationID string
	OriginalError error
	Retryable     bool
	Timestamp     time.Time
	Context       map[string]interface{}
}

// Error returns the error message
func (e *ResilientError) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.OriginalError)
	}
	return e.Message
}

// Unwrap returns the original error
func (e *ResilientError) Unwrap() error {
	return e.OriginalError
}

// NewResilientError creates a new resilient error
func NewResilientError(
	code string,
	message string,
	correlationID string,
	originalError error,
	retryable bool,
) *ResilientError {
	return &ResilientError{
		Code:          code,
		Message:       message,
		CorrelationID: correlationID,
		OriginalError: originalError,
		Retryable:     retryable,
		Timestamp:     time.Now(),
		Context:       make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *ResilientError) WithContext(key string, value interface{}) *ResilientError {
	e.Context[key] = value
	return e
}

// Common error variables
var (
	ErrNilInput           = errors.New("input cannot be nil")
	ErrEmptyPayload       = errors.New("payload cannot be empty")
	ErrInvalidJSON        = errors.New("invalid JSON format")
	ErrValidationFailed   = errors.New("validation failed")
	ErrSchemaMismatch     = errors.New("schema mismatch")
	ErrCircuitOpen        = errors.New("circuit breaker is open")
	ErrTimeoutExceeded    = errors.New("operation timeout exceeded")
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")
)
