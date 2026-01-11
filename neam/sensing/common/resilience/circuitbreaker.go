package resilience

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

// CircuitBreakerConfig holds configuration for circuit breaker
type CircuitBreakerConfig struct {
	// Enabled controls whether circuit breaker is active
	Enabled bool `yaml:"enabled"`

	// Timeout is the duration before circuit breaker switches to half-open
	Timeout time.Duration `yaml:"timeout"`

	// MaxRequests is the maximum number of requests allowed to pass through
	// when the circuit breaker is half-open
	MaxRequests uint32 `yaml:"max_requests"`

	// Interval is the cyclic period of the closed state
	// for the circuit breaker to clear the internal counts
	Interval time.Duration `yaml:"interval"`

	// ReadyToTripThreshold is the number of consecutive failures
	// required to open the circuit
	ReadyToTripThreshold uint32 `yaml:"ready_to_trip_threshold"`

	// SuccessThreshold is the number of successes required to close
	// the circuit from half-open to closed
	SuccessThreshold uint32 `yaml:"success_threshold"`

	// RequestTimeout is the timeout for each individual request
	RequestTimeout time.Duration `yaml:"request_timeout"`
}

// CircuitBreakerManager manages multiple circuit breakers for different operations
type CircuitBreakerManager struct {
	breakers   map[string]*gobreaker.CircuitBreaker
	metrics    *MetricsCollector
	logger     *zap.Logger
	config     CircuitBreakerConfig
	mu         sync.RWMutex
	fallbacks  map[string]FallbackFunc
}

// FallbackFunc defines the function type for fallback behaviors
type FallbackFunc func(ctx context.Context, operation string, err error) (interface{}, error)

// CircuitBreakerResult holds the result of a circuit breaker execution
type CircuitBreakerResult struct {
	Response    interface{}
	Error       error
	WasFallback bool
	CircuitOpen bool
}

// DefaultCircuitBreakerConfig returns the default configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Enabled:              true,
		Timeout:              30 * time.Second,
		MaxRequests:          5,
		Interval:             60 * time.Second,
		ReadyToTripThreshold: 5,
		SuccessThreshold:     1,
		RequestTimeout:       10 * time.Second,
	}
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(
	config CircuitBreakerConfig,
	metrics *MetricsCollector,
	logger *zap.Logger,
) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers:  make(map[string]*gobreaker.CircuitBreaker),
		metrics:   metrics,
		logger:    logger,
		config:    config,
		fallbacks: make(map[string]FallbackFunc),
	}
}

// RegisterCircuitBreaker creates and registers a new circuit breaker for an operation
func (m *CircuitBreakerManager) RegisterCircuitBreaker(
	name string,
	fallback FallbackFunc,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.breakers[name]; exists {
		return fmt.Errorf("circuit breaker already registered for: %s", name)
	}

	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: m.config.MaxRequests,
		Interval:    m.config.Interval,
		Timeout:     m.config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= m.config.ReadyToTripThreshold
		},
		IsSuccess: func(err error) bool {
			return err == nil
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			m.logger.Info("Circuit breaker state changed",
				zap.String("circuit", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()))

			if m.metrics != nil {
				m.metrics.RecordCircuitBreakerState(name, from, to)
			}
		},
		OnRequestComplete: func(name string, request gobreaker.Request) {
			if m.metrics != nil {
				duration := time.Since(request.StartTime)
				m.metrics.RecordCircuitBreakerRequest(name, request.Result, request.Err, duration)
			}
		},
	}

	cb := gobreaker.NewCircuitBreaker(settings)
	m.breakers[name] = cb
	m.fallbacks[name] = fallback

	m.logger.Info("Circuit breaker registered",
		zap.String("name", name),
		zap.Uint32("max_requests", m.config.MaxRequests),
		zap.Duration("timeout", m.config.Timeout),
		zap.Uint32("ready_to_trip_threshold", m.config.ReadyToTripThreshold))

	return nil
}

// Execute runs a function through the circuit breaker
func (m *CircuitBreakerManager) Execute(
	ctx context.Context,
	operation string,
	fn func(ctx context.Context) (interface{}, error),
) *CircuitBreakerResult {
	m.mu.RLock()
	cb, exists := m.breakers[operation]
	fallback := m.fallbacks[operation]
	m.mu.RUnlock()

	if !exists || !m.config.Enabled {
		// No circuit breaker, execute directly
		resp, err := fn(ctx)
		return &CircuitBreakerResult{
			Response: resp,
			Error:    err,
		}
	}

	// Create a context with timeout for the request
	reqCtx, cancel := context.WithTimeout(ctx, m.config.RequestTimeout)
	defer cancel()

	result, err := cb.Execute(func() (interface{}, error) {
		return fn(reqCtx)
	})

	if err != nil {
		// Check if circuit is open and fallback is available
		state := cb.State()
		if state == gobreaker.StateOpen {
			m.logger.Warn("Circuit breaker open, attempting fallback",
				zap.String("operation", operation),
				zap.Error(err))

			if fallback != nil {
				fallbackResp, fallbackErr := fallback(ctx, operation, err)
				return &CircuitBreakerResult{
					Response:    fallbackResp,
					Error:       fallbackErr,
					WasFallback: true,
					CircuitOpen: true,
				}, nil
			}
		}
		return &CircuitBreakerResult{
			Response:  result,
			Error:     err,
			CircuitOpen: state == gobreaker.StateOpen,
		}
	}

	return &CircuitBreakerResult{
		Response: result,
		Error:    nil,
	}
}

// GetState returns the current state of a circuit breaker
func (m *CircuitBreakerManager) GetState(operation string) (gobreaker.State, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cb, exists := m.breakers[operation]
	if !exists {
		return gobreaker.State(0), fmt.Errorf("circuit breaker not found: %s", operation)
	}
	return cb.State(), nil
}

// GetAllStates returns the state of all circuit breakers
func (m *CircuitBreakerManager) GetAllStates() map[string]gobreaker.State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make(map[string]gobreaker.State)
	for name, cb := range m.breakers {
		states[name] = cb.State()
	}
	return states
}

// Reset resets a specific circuit breaker
func (m *CircuitBreakerManager) Reset(operation string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cb, exists := m.breakers[operation]
	if !exists {
		return fmt.Errorf("circuit breaker not found: %s", operation)
	}

	cb.Reset()
	m.logger.Info("Circuit breaker reset", zap.String("operation", operation))
	return nil
}

// ResetAll resets all circuit breakers
func (m *CircuitBreakerManager) ResetAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, cb := range m.breakers {
		cb.Reset()
		m.logger.Info("Circuit breaker reset", zap.String("operation", name))
	}
}

// Close cleans up all circuit breakers
func (m *CircuitBreakerManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name := range m.breakers {
		delete(m.breakers, name)
		delete(m.fallbacks, name)
	}
}

// Common error variables
var (
	ErrCircuitOpen      = errors.New("circuit breaker is open")
	ErrCircuitHalfOpen  = errors.New("circuit breaker is half-open")
	ErrRequestTimeout   = errors.New("request timeout exceeded")
	ErrFallbackFailed   = errors.New("fallback also failed")
)
