package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"go.uber.org/zap"
)

// JSONValidator implements Validator interface for JSON Schema validation
type JSONValidator struct {
	schemaID      string
	schemaVersion string
	schema        *jsonschema.Schema
	config        ValidatorConfig
	logger        *zap.Logger
}

// NewJSONValidator creates a new JSON schema validator
func NewJSONValidator(schemaJSON string) (*JSONValidator, error) {
	// Parse the JSON schema
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true

	schema, err := compiler.Compile(strings.NewReader(schemaJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to compile JSON schema: %w", err)
	}

	return &JSONValidator{
		schema:        schema,
		schemaID:      "",
		schemaVersion: "1.0.0",
		config: ValidatorConfig{
			Mode:         ValidationModeStrict,
			CacheResults: false,
		},
		logger: nil,
	}, nil
}

// NewJSONValidatorWithConfig creates a new JSON schema validator with configuration
func NewJSONValidatorWithConfig(schemaID, schemaVersion string, schemaJSON string, config ValidatorConfig, logger *zap.Logger) (*JSONValidator, error) {
	// Parse the JSON schema
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true

	schema, err := compiler.Compile(strings.NewReader(schemaJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to compile JSON schema: %w", err)
	}

	return &JSONValidator{
		schemaID:      schemaID,
		schemaVersion: schemaVersion,
		schema:        schema,
		config:        config,
		logger:        logger,
	}, nil
}

// Validate validates the provided JSON data against the schema
func (v *JSONValidator) Validate(ctx context.Context, data json.RawMessage) (*ValidationResult, error) {
	startTime := time.Now()

	// Parse the JSON data
	var parsedData interface{}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return &ValidationResult{
			Valid:     false,
			Errors:    []ValidationError{{Field: "root", Message: "invalid JSON format", ErrorType: "parse_error", Actual: err.Error()}},
			Timestamp: startTime,
			Duration:  time.Since(startTime),
		}, nil
	}

	// Perform validation
	var errors []ValidationError
	err := v.schema.Validate(parsedData)
	if err != nil {
		validationErr, ok := err.(*jsonschema.ValidationError)
		if ok {
			errors = v.extractValidationErrors(validationErr, "")
		} else {
			errors = []ValidationError{{Field: "root", Message: err.Error(), ErrorType: "validation_error"}}
		}
	}

	result := &ValidationResult{
		Valid:     len(errors) == 0,
		Errors:    errors,
		Timestamp: startTime,
		Duration:  time.Since(startTime),
	}

	// Apply custom validation rules if configured
	if len(v.config.CustomRules) > 0 && result.Valid {
		customErrors := v.applyCustomRules(parsedData)
		result.Valid = len(customErrors) == 0
		result.Errors = append(result.Errors, customErrors...)
	}

	return result, nil
}

// ValidateWithContext validates with additional context
func (v *JSONValidator) ValidateWithContext(ctx context.Context, data json.RawMessage, metadata map[string]interface{}) (*ValidationResult, error) {
	result, err := v.Validate(ctx, data)
	if result != nil {
		result.Metadata = metadata
	}
	return result, err
}

// extractValidationErrors extracts structured validation errors from jsonschema.ValidationError
func (v *JSONValidator) extractValidationErrors(ve *jsonschema.ValidationError, prefix string) []ValidationError {
	var errors []ValidationError

	for _, err := range ve.BasicMessage() {
		field := prefix
		if len(err.InstanceLocation) > 0 {
			if len(prefix) > 0 {
				field = prefix + "." + err.InstanceLocation
			} else {
				field = err.InstanceLocation
			}
		}

		errorType := "validation_error"
		expected := ""
		constraint := ""

		// Extract constraint information from the error
		if keyword, ok := err.Keyword.(string); ok {
			errorType = keyword
			constraint = keyword
		}

		// Extract expected value for enum errors
		if len(err.Expected) > 0 {
			expected = fmt.Sprintf("%v", err.Expected)
		}

		validationErr := ValidationError{
			Field:      field,
			Message:    err.Message,
			Expected:   expected,
			ErrorType:  errorType,
			Constraint: constraint,
		}

		// Try to extract actual value
		if len(err.Value) > 0 {
			validationErr.Actual = fmt.Sprintf("%v", err.Value)
		}

		errors = append(errors, validationErr)
	}

	// Handle nested errors (for complex types)
	for _, nested := range ve.Nested() {
		nestedErrors := v.extractValidationErrors(nested, prefix)
		errors = append(errors, nestedErrors...)
	}

	return errors
}

// applyCustomRules applies custom validation rules
func (v *JSONValidator) applyCustomRules(data interface{}) []ValidationError {
	var errors []ValidationError

	for _, rule := range v.config.CustomRules {
		if err := v.evaluateCustomRule(data, rule); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// evaluateCustomRule evaluates a single custom rule
func (v *JSONValidator) evaluateCustomRule(data interface{}, rule CustomRule) *ValidationError {
	// Get the field value using simple path navigation
	value, found := v.getFieldValue(data, rule.Field)
	if !found {
		// Field not found - check if it's required
		if v.isFieldRequired(rule.Field) {
			return &ValidationError{
				Field:     rule.Field,
				Message:   fmt.Sprintf("field is required: %s", rule.Message),
				ErrorType: "required",
			}
		}
		return nil
	}

	// Evaluate the predicate
	if !v.evaluatePredicate(value, rule.Predicate, rule.Parameters) {
		return &ValidationError{
			Field:     rule.Field,
			Message:   rule.Message,
			ErrorType: "custom_rule",
		}
	}

	return nil
}

// getFieldValue retrieves a field value from the data using dot notation
func (v *JSONValidator) getFieldValue(data interface{}, fieldPath string) (interface{}, bool) {
	parts := strings.Split(fieldPath, ".")
	current := data

	for _, part := range parts {
		switch c := current.(type) {
		case map[string]interface{}:
			if val, ok := c[part]; ok {
				current = val
			} else {
				return nil, false
			}
		case []interface{}:
			// Handle array indexing
			if part == "*" {
				// Return first element for wildcard
				if len(c) > 0 {
					current = c[0]
				} else {
					return nil, false
				}
			} else {
				return nil, false
			}
		default:
			return nil, false
		}
	}

	return current, true
}

// isFieldRequired checks if a field is in the required list
func (v *JSONValidator) isFieldRequired(field string) bool {
	for _, f := range v.config.RequiredFields {
		if f == field {
			return true
		}
	}
	return false
}

// evaluatePredicate evaluates a predicate against a value
func (v *JSONValidator) evaluatePredicate(value interface{}, predicate string, params map[string]string) bool {
	switch predicate {
	case "not_empty":
		return value != nil && value != ""
	case "is_number":
		_, ok := value.(float64)
		return ok
	case "is_integer":
		_, ok := value.(float64)
		if ok {
			f := value.(float64)
			return f == float64(int(f))
		}
		return false
	case "is_string":
		_, ok := value.(string)
		return ok
	case "is_boolean":
		_, ok := value.(bool)
		return ok
	case "greater_than":
		if f, ok := value.(float64); ok {
			if min, ok := params["min"]; ok {
				var minVal float64
				fmt.Sscanf(min, "%f", &minVal)
				return f > minVal
			}
		}
	case "less_than":
		if f, ok := value.(float64); ok {
			if max, ok := params["max"]; ok {
				var maxVal float64
				fmt.Sscanf(max, "%f", &maxVal)
				return f < maxVal
			}
		}
	case "matches_pattern":
		if s, ok := value.(string); ok {
			if pattern, ok := params["pattern"]; ok {
				if re, err := regexp.Compile(pattern); err == nil {
					return re.MatchString(s)
				}
			}
		}
	case "in_list":
		if s, ok := value.(string); ok {
			if values, ok := params["values"]; ok {
				for _, v := range strings.Split(values, ",") {
					if strings.TrimSpace(v) == s {
						return true
					}
				}
			}
		}
	}

	return true
}

// GetSchemaID returns the schema identifier
func (v *JSONValidator) GetSchemaID() string {
	return v.schemaID
}

// GetSchemaVersion returns the schema version
func (v *JSONValidator) GetSchemaVersion() string {
	return v.schemaVersion
}

// GetSchemaType returns the type of schema
func (v *JSONValidator) GetSchemaType() SchemaType {
	return SchemaTypeJSON
}

// GetCompiledSchema returns the compiled JSON schema for inspection
func (v *JSONValidator) GetCompiledSchema() *jsonschema.Schema {
	return v.schema
}
