package schema

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestJSONValidator_Validate(t *testing.T) {
	logger := zap.NewNop()

	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "string"},
			"value": {"type": "number"},
			"enabled": {"type": "boolean"}
		},
		"required": ["id"]
	}`

	validator, err := NewJSONValidator(schemaJSON)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	ctx := context.Background()

	// Test valid data
	validData := `{"id": "test-123", "value": 42.5, "enabled": true}`
	result, err := validator.Validate(ctx, []byte(validData))
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}

	// Test missing required field
	invalidData := `{"value": 42.5}`
	result, err = validator.Validate(ctx, []byte(invalidData))
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result for missing required field")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected at least one validation error")
	}

	// Check that error is for the 'id' field
	found := false
	for _, e := range result.Errors {
		if e.Field == "id" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for 'id' field")
	}

	// Test wrong type
	wrongTypeData := `{"id": 123}`
	result, err = validator.Validate(ctx, []byte(wrongTypeData))
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result for wrong type")
	}
}

func TestJSONValidator_WithContext(t *testing.T) {
	logger := zap.NewNop()

	schemaJSON := `{"type": "object", "properties": {"name": {"type": "string"}}}`

	validator, err := NewJSONValidator(schemaJSON)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	ctx := context.Background()
	metadata := map[string]interface{}{
		"source": "test",
		"request_id": "req-123",
	}

	validData := `{"name": "test"}`
	result, err := validator.ValidateWithContext(ctx, []byte(validData), metadata)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}

	if result.Metadata == nil {
		t.Error("Expected metadata to be set")
	}

	if result.Metadata["source"] != "test" {
		t.Error("Expected metadata source to be 'test'")
	}
}

func TestJSONValidator_ValidationDuration(t *testing.T) {
	logger := zap.NewNop()

	schemaJSON := `{"type": "object"}`

	validator, err := NewJSONValidator(schemaJSON)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	ctx := context.Background()
	validData := `{"key": "value"}`

	start := time.Now()
	result, _ := validator.Validate(ctx, []byte(validData))
	elapsed := time.Since(start)

	if result.Duration > elapsed {
		t.Error("Recorded duration should be less than or equal to elapsed time")
	}

	if result.Duration < 0 {
		t.Error("Duration should not be negative")
	}
}

func TestJSONValidator_InvalidJSON(t *testing.T) {
	logger := zap.NewNop()

	schemaJSON := `{"type": "object"}`

	validator, err := NewJSONValidator(schemaJSON)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	ctx := context.Background()

	// Test invalid JSON
	invalidJSON := `{invalid json}`
	result, err := validator.Validate(ctx, []byte(invalidJSON))
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result for invalid JSON")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected at least one error for invalid JSON")
	}

	// Check for parse error
	found := false
	for _, e := range result.Errors {
		if e.ErrorType == "parse_error" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected parse_error type")
	}
}

func TestJSONValidator_RequiredFields(t *testing.T) {
	logger := zap.NewNop()

	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "string"},
			"name": {"type": "string"},
			"email": {"type": "string"}
		},
		"required": ["id", "email"]
	}`

	validator, err := NewJSONValidatorWithConfig("test", "1.0.0", schemaJSON, ValidatorConfig{
		RequiredFields: []string{"id", "email"},
	}, logger)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	ctx := context.Background()

	// Test with missing required field
	data := `{"id": "123"}`
	result, _ := validator.Validate(ctx, []byte(data))

	if result.Valid {
		t.Error("Expected invalid result for missing 'email' field")
	}

	// Test with all required fields
	data = `{"id": "123", "email": "test@example.com"}`
	result, _ = validator.Validate(ctx, []byte(data))

	if !result.Valid {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}
}

func TestJSONValidator_CustomRules(t *testing.T) {
	logger := zap.NewNop()

	schemaJSON := `{"type": "object", "properties": {"age": {"type": "number"}}}`

	config := ValidatorConfig{
		CustomRules: []CustomRule{
			{
				Name:      "age_positive",
				Field:     "age",
				Predicate: "greater_than",
				Message:   "Age must be positive",
				Parameters: map[string]string{"min": "0"},
			},
		},
	}

	validator, err := NewJSONValidatorWithConfig("test", "1.0.0", schemaJSON, config, logger)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	ctx := context.Background()

	// Test with negative age
	data := `{"age": -5}`
	result, _ := validator.Validate(ctx, []byte(data))

	if result.Valid {
		t.Error("Expected invalid result for negative age")
	}

	// Test with positive age
	data = `{"age": 25}`
	result, _ = validator.Validate(ctx, []byte(data))

	if !result.Valid {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}
}

func TestValidationResult_ToJSON(t *testing.T) {
	result := &ValidationResult{
		Valid:     false,
		Timestamp: time.Now(),
		Errors: []ValidationError{
			{
				Field:     "field1",
				Message:   "error message",
				ErrorType: "required",
			},
		},
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if parsed["valid"] != false {
		t.Error("Expected valid to be false")
	}
}

func BenchmarkJSONValidator_Validate(b *testing.B) {
	logger := zap.NewNop()

	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "string"},
			"value": {"type": "number"},
			"enabled": {"type": "boolean"},
			"tags": {"type": "array"}
		},
		"required": ["id"]
	}`

	validator, _ := NewJSONValidator(schemaJSON)

	ctx := context.Background()
	validData := []byte(`{"id": "test-123", "value": 42.5, "enabled": true, "tags": ["a", "b"]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ctx, validData)
	}
}
