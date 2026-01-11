package resilience

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestDLQMessageSerialization tests DLQ message serialization
func TestDLQMessageSerialization(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	serializer := NewMessageSerializer(logger)

	originalPayload := json.RawMessage(`{"vehicle_id":"123","speed":65.5}`)

	dlqMsg := &DLQMessage{
		ID:                "test-id-123",
		CorrelationID:     "corr-id-456",
		Timestamp:         time.Now(),
		SourceAdapter:     "transport",
		SourceTopic:       "transport.telemetry",
		ErrorReason:       ErrorReasonProcessing,
		ErrorCode:         "PROCESSING_ERROR",
		ErrorMessage:      "failed to process message",
		OriginalPayload:   originalPayload,
		Metadata:          map[string]interface{}{"key": "value"},
		RetryCount:        3,
		FirstFailureTime:  time.Now().Add(-1 * time.Minute),
		LastFailureTime:   time.Now(),
	}

	// Serialize
	data, err := serializer.Serialize(dlqMsg)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	// Deserialize
	decoded, err := serializer.DeserializeDLQ(data)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	// Verify
	if decoded.ID != dlqMsg.ID {
		t.Errorf("ID mismatch: expected %s, got %s", dlqMsg.ID, decoded.ID)
	}
	if decoded.CorrelationID != dlqMsg.CorrelationID {
		t.Errorf("CorrelationID mismatch")
	}
	if decoded.SourceAdapter != dlqMsg.SourceAdapter {
		t.Errorf("SourceAdapter mismatch")
	}
	if decoded.ErrorReason != dlqMsg.ErrorReason {
		t.Errorf("ErrorReason mismatch: expected %s, got %s", dlqMsg.ErrorReason, decoded.ErrorReason)
	}
	if decoded.RetryCount != dlqMsg.RetryCount {
		t.Errorf("RetryCount mismatch")
	}
}

// TestValidationResult tests validation result structure
func TestValidationResult(t *testing.T) {
	schemaValidator := NewSchemaValidator(zap.NewNop())

	// Test valid data
	validData := map[string]interface{}{
		"vehicle_id": "123e4567-e89b-12d3-a456-426614174000",
		"timestamp":  time.Now(),
		"latitude":   37.7749,
		"longitude":  -122.4194,
		"speed":      65.5,
		"fuel_level": 75.0,
	}

	result := schemaValidator.Validate(context.Background(), validData)
	if !result.Valid {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}
}

// TestResilientError tests resilient error structure
func TestResilientError(t *testing.T) {
	originalErr := errors.New("original error")
	err := NewResilientError(
		"ERR_001",
		"Test error message",
		"correlation-123",
		originalErr,
		true,
	)

	// Test error message
	expectedMsg := "Test error message: original error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test unwrap
	if err.Unwrap() != originalErr {
		t.Error("Unwrap should return original error")
	}

	// Test context
	err.WithContext("key1", "value1")
	if err.Context["key1"] != "value1" {
		t.Error("Context not set correctly")
	}
}

// TestRetryPolicy tests retry policy behavior
func TestRetryPolicy(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	metrics := NewMetricsCollector(logger)

	config := RetryConfig{
		MaxAttempts:    3,
		InitialDelay:   10 * time.Millisecond,
		MaxDelay:       50 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         false,
		RetryableCodes: []string{"retryable"},
	}

	policy := NewRetryPolicy(config, metrics, logger)

	// Test successful retry
	successAttempt := 0
	result := policy.Execute(context.Background(), func(attempt int) (interface{}, error) {
		successAttempt = attempt
		if attempt < 2 {
			return nil, errors.New("retryable error")
		}
		return "success", nil
	})

	if !result.Success {
		t.Error("Expected success after retries")
	}
	if successAttempt != 2 {
		t.Errorf("Expected success on attempt 2, got attempt %d", successAttempt)
	}

	// Test failure after all retries
	policy2 := NewRetryPolicy(config, metrics, logger)
	result2 := policy2.Execute(context.Background(), func(attempt int) (interface{}, error) {
		return nil, errors.New("retryable error")
	})

	if result2.Success {
		t.Error("Expected failure after all retries")
	}
	if result2.Attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", result2.Attempts)
	}
}

// TestFallbackRegistry tests fallback registry
func TestFallbackRegistry(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewFallbackRegistry(logger)

	// Register fallback
	callback := func(ctx context.Context, operation string, err error) (interface{}, error) {
		return "fallback_result", nil
	}

	registry.RegisterFallback("test-adapter", "test-operation", callback)

	// Retrieve fallback
	fn, ok := registry.GetFallback("test-adapter", "test-operation")
	if !ok {
		t.Error("Fallback not found")
	}

	result, err := fn(context.Background(), "test-operation", nil)
	if err != nil {
		t.Errorf("Fallback returned error: %v", err)
	}
	if result != "fallback_result" {
		t.Errorf("Unexpected fallback result: %v", result)
	}
}

// TestTransportTelemetryValidation tests transport telemetry validation
func TestTransportTelemetryValidation(t *testing.T) {
	validator := NewSchemaValidator(zap.NewNop())

	// Valid telemetry
	telemetry := &TransportTelemetry{
		VehicleID: "123e4567-e89b-12d3-a456-426614174000",
		Timestamp: time.Now(),
		Latitude:  37.7749,
		Longitude: -122.4194,
		Speed:     65.5,
		FuelLevel: 75.0,
	}

	result := validator.Validate(context.Background(), telemetry)
	if !result.Valid {
		t.Errorf("Expected valid telemetry, got errors: %v", result.Errors)
	}

	// Invalid telemetry - invalid UUID
	telemetry2 := &TransportTelemetry{
		VehicleID: "invalid-uuid",
		Timestamp: time.Now(),
		Latitude:  37.7749,
		Longitude: -122.4194,
	}

	result2 := validator.Validate(context.Background(), telemetry2)
	if result2.Valid {
		t.Error("Expected invalid telemetry due to UUID validation")
	}

	// Invalid latitude
	telemetry3 := &TransportTelemetry{
		VehicleID: "123e4567-e89b-12d3-a456-426614174000",
		Timestamp: time.Now(),
		Latitude:  100.0, // Invalid: > 90
		Longitude: -122.4194,
	}

	result3 := validator.Validate(context.Background(), telemetry3)
	if result3.Valid {
		t.Error("Expected invalid telemetry due to latitude validation")
	}
}

// TestIndustrialTelemetryValidation tests industrial telemetry validation
func TestIndustrialTelemetryValidation(t *testing.T) {
	validator := NewSchemaValidator(zap.NewNop())

	// Valid telemetry
	telemetry := &IndustrialTelemetry{
		NodeID:    "node-001",
		AssetID:   "asset-001",
		Timestamp: time.Now(),
		Value:     75.5,
		Quality:   "Good",
		DataType:  "Float",
	}

	result := validator.Validate(context.Background(), telemetry)
	if !result.Valid {
		t.Errorf("Expected valid telemetry, got errors: %v", result.Errors)
	}

	// Invalid quality
	telemetry2 := &IndustrialTelemetry{
		NodeID:    "node-001",
		AssetID:   "asset-001",
		Timestamp: time.Now(),
		Value:     75.5,
		Quality:   "InvalidQuality", // Not in oneof list
		DataType:  "Float",
	}

	result2 := validator.Validate(context.Background(), telemetry2)
	if result2.Valid {
		t.Error("Expected invalid telemetry due to quality validation")
	}
}

// TestEnergyTelemetryValidation tests energy telemetry validation
func TestEnergyTelemetryValidation(t *testing.T) {
	validator := NewSchemaValidator(zap.NewNop())

	// Valid telemetry
	telemetry := &EnergyTelemetry{
		MeterID:      "meter-001",
		Timestamp:    time.Now(),
		ActivePower:  1000.0,
		Voltage:      240.0,
		Current:      4.2,
		Frequency:    60.0,
		PowerFactor:  0.95,
		Quality:      "Good",
	}

	result := validator.Validate(context.Background(), telemetry)
	if !result.Valid {
		t.Errorf("Expected valid telemetry, got errors: %v", result.Errors)
	}

	// Invalid frequency (out of range)
	telemetry2 := &EnergyTelemetry{
		MeterID:      "meter-001",
		Timestamp:    time.Now(),
		ActivePower:  1000.0,
		Frequency:    70.0, // Invalid: > 65
		Quality:      "Good",
	}

	result2 := validator.Validate(context.Background(), telemetry2)
	if result2.Valid {
		t.Error("Expected invalid telemetry due to frequency validation")
	}
}
