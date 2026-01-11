package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

// TestCircuitBreakerStateTransition tests the circuit breaker state transitions
func TestCircuitBreakerStateTransition(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	metrics := NewMetricsCollector(logger)

	config := CircuitBreakerConfig{
		Enabled:              true,
		Timeout:              100 * time.Millisecond,
		MaxRequests:          3,
		Interval:             1 * time.Second,
		ReadyToTripThreshold: 3,
		SuccessThreshold:     1,
		RequestTimeout:       1 * time.Second,
	}

	manager := NewCircuitBreakerManager(config, metrics, logger)

	// Register a circuit breaker with a fallback
	testErr := errors.New("test error")
	fallbackCalled := false

	fallback := func(ctx context.Context, operation string, err error) (interface{}, error) {
		fallbackCalled = true
		return "fallback_response", nil
	}

	err := manager.RegisterCircuitBreaker("test-circuit", fallback)
	if err != nil {
		t.Fatalf("Failed to register circuit breaker: %v", err)
	}

	// Execute until circuit opens (5 consecutive failures)
	for i := 0; i < 5; i++ {
		result := manager.Execute(context.Background(), "test-circuit", func(ctx context.Context) (interface{}, error) {
			return nil, testErr
		})
		if result.Error == nil {
			t.Error("Expected error but got nil")
		}
	}

	// Verify circuit is open
	state := manager.GetState("test-circuit")
	if state != gobreaker.StateOpen {
		t.Errorf("Expected circuit to be open, got: %v", state)
	}

	// Verify fallback is called when circuit is open
	fallbackCalled = false
	result := manager.Execute(context.Background(), "test-circuit", func(ctx context.Context) (interface{}, error) {
		t.Error("Function should not be called when circuit is open")
		return nil, testErr
	})

	if !fallbackCalled {
		t.Error("Fallback should have been called when circuit is open")
	}

	if result.WasFallback != true {
		t.Error("Result should indicate fallback was used")
	}

	if result.Response != "fallback_response" {
		t.Errorf("Expected fallback response, got: %v", result.Response)
	}
}

// TestCircuitBreakerRecovery tests recovery from open to half-open to closed
func TestCircuitBreakerRecovery(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	metrics := NewMetricsCollector(logger)

	config := CircuitBreakerConfig{
		Enabled:              true,
		Timeout:              50 * time.Millisecond,
		MaxRequests:          1,
		Interval:             1 * time.Second,
		ReadyToTripThreshold: 2,
		SuccessThreshold:     1,
		RequestTimeout:       1 * time.Second,
	}

	manager := NewCircuitBreakerManager(config, metrics, logger)

	fallback := func(ctx context.Context, operation string, err error) (interface{}, error) {
		return "fallback", nil
	}

	err := manager.RegisterCircuitBreaker("recovery-circuit", fallback)
	if err != nil {
		t.Fatalf("Failed to register circuit breaker: %v", err)
	}

	// Open the circuit
	for i := 0; i < 3; i++ {
		manager.Execute(context.Background(), "recovery-circuit", func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("failure")
		})
	}

	state := manager.GetState("recovery-circuit")
	if state != gobreaker.StateOpen {
		t.Errorf("Expected circuit to be open, got: %v", state)
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Try a successful call - should transition to half-open then closed
	result := manager.Execute(context.Background(), "recovery-circuit", func(ctx context.Context) (interface{}, error) {
		return "success", nil
	})

	if result.Error != nil {
		t.Errorf("Expected success, got error: %v", result.Error)
	}

	if result.Response != "success" {
		t.Errorf("Expected 'success' response, got: %v", result.Response)
	}

	// Verify circuit is now closed
	state = manager.GetState("recovery-circuit")
	if state != gobreaker.StateClosed {
		t.Errorf("Expected circuit to be closed after success, got: %v", state)
	}
}

// TestCircuitBreakerFailFast tests that circuit breaker fails fast when open
func TestCircuitBreakerFailFast(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	metrics := NewMetricsCollector(logger)

	config := CircuitBreakerConfig{
		Enabled:              true,
		Timeout:              100 * time.Millisecond,
		MaxRequests:          3,
		Interval:             1 * time.Second,
		ReadyToTripThreshold: 2,
		SuccessThreshold:     1,
		RequestTimeout:       1 * time.Second,
	}

	manager := NewCircuitBreakerManager(config, metrics, logger)

	fallback := func(ctx context.Context, operation string, err error) (interface{}, error) {
		return "fallback", nil
	}

	err := manager.RegisterCircuitBreaker("failfast-circuit", fallback)
	if err != nil {
		t.Fatalf("Failed to register circuit breaker: %v", err)
	}

	// Open the circuit
	for i := 0; i < 2; i++ {
		manager.Execute(context.Background(), "failfast-circuit", func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("failure")
		})
	}

	// Track if the function was called
	fnCalled := false
	startTime := time.Now()

	// Execute when circuit is open - should fail fast
	result := manager.Execute(context.Background(), "failfast-circuit", func(ctx context.Context) (interface{}, error) {
		fnCalled = true
		// Simulate long operation
		time.Sleep(5 * time.Second)
		return nil, nil
	})

	elapsed := time.Since(startTime)

	if fnCalled {
		t.Error("Function should not have been called when circuit is open")
	}

	if result.CircuitOpen != true {
		t.Error("Result should indicate circuit is open")
	}

	// Should fail fast (less than 100ms)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Circuit breaker should fail fast, took: %v", elapsed)
	}
}

// TestCircuitBreakerDisabled tests behavior when circuit breaker is disabled
func TestCircuitBreakerDisabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	metrics := NewMetricsCollector(logger)

	config := CircuitBreakerConfig{
		Enabled:              false,
		Timeout:              100 * time.Millisecond,
		MaxRequests:          3,
		Interval:             1 * time.Second,
		ReadyToTripThreshold: 2,
		SuccessThreshold:     1,
		RequestTimeout:       1 * time.Second,
	}

	manager := NewCircuitBreakerManager(config, metrics, logger)

	fallback := func(ctx context.Context, operation string, err error) (interface{}, error) {
		return "fallback", nil
	}

	err := manager.RegisterCircuitBreaker("disabled-circuit", fallback)
	if err != nil {
		t.Fatalf("Failed to register circuit breaker: %v", err)
	}

	fnCalled := false

	// Should execute directly without circuit breaker
	result := manager.Execute(context.Background(), "disabled-circuit", func(ctx context.Context) (interface{}, error) {
		fnCalled = true
		return "direct_result", nil
	})

	if !fnCalled {
		t.Error("Function should have been called")
	}

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}

	if result.Response != "direct_result" {
		t.Errorf("Expected 'direct_result', got: %v", result.Response)
	}
}

// TestCircuitBreakerGetAllStates tests getting all circuit breaker states
func TestCircuitBreakerGetAllStates(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	metrics := NewMetricsCollector(logger)

	config := DefaultCircuitBreakerConfig()
	manager := NewCircuitBreakerManager(config, metrics, logger)

	manager.RegisterCircuitBreaker("circuit-1", nil)
	manager.RegisterCircuitBreaker("circuit-2", nil)
	manager.RegisterCircuitBreaker("circuit-3", nil)

	states := manager.GetAllStates()

	if len(states) != 3 {
		t.Errorf("Expected 3 circuit breakers, got: %d", len(states))
	}

	for name, state := range states {
		if state != gobreaker.StateClosed {
			t.Errorf("Expected %s to be closed, got: %v", name, state)
		}
	}
}

// TestCircuitBreakerReset tests resetting a specific circuit breaker
func TestCircuitBreakerReset(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	metrics := NewMetricsCollector(logger)

	config := CircuitBreakerConfig{
		Enabled:              true,
		Timeout:              100 * time.Millisecond,
		MaxRequests:          3,
		Interval:             1 * time.Second,
		ReadyToTripThreshold: 2,
		SuccessThreshold:     1,
		RequestTimeout:       1 * time.Second,
	}

	manager := NewCircuitBreakerManager(config, metrics, logger)
	manager.RegisterCircuitBreaker("reset-circuit", nil)

	// Open the circuit
	for i := 0; i < 2; i++ {
		manager.Execute(context.Background(), "reset-circuit", func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("failure")
		})
	}

	state := manager.GetState("reset-circuit")
	if state != gobreaker.StateOpen {
		t.Errorf("Expected circuit to be open, got: %v", state)
	}

	// Reset the circuit
	err := manager.Reset("reset-circuit")
	if err != nil {
		t.Errorf("Failed to reset circuit: %v", err)
	}

	state = manager.GetState("reset-circuit")
	if state != gobreaker.StateClosed {
		t.Errorf("Expected circuit to be closed after reset, got: %v", state)
	}
}

// TestCircuitBreakerResetAll tests resetting all circuit breakers
func TestCircuitBreakerResetAll(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	metrics := NewMetricsCollector(logger)

	config := CircuitBreakerConfig{
		Enabled:              true,
		Timeout:              100 * time.Millisecond,
		MaxRequests:          3,
		Interval:             1 * time.Second,
		ReadyToTripThreshold: 2,
		SuccessThreshold:     1,
		RequestTimeout:       1 * time.Second,
	}

	manager := NewCircuitBreakerManager(config, metrics, logger)
	manager.RegisterCircuitBreaker("reset-all-1", nil)
	manager.RegisterCircuitBreaker("reset-all-2", nil)

	// Open both circuits
	for i := 0; i < 2; i++ {
		manager.Execute(context.Background(), "reset-all-1", func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("failure")
		})
		manager.Execute(context.Background(), "reset-all-2", func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("failure")
		})
	}

	// Reset all
	manager.ResetAll()

	states := manager.GetAllStates()
	for _, state := range states {
		if state != gobreaker.StateClosed {
			t.Errorf("Expected all circuits to be closed after reset, got: %v", state)
		}
	}
}
