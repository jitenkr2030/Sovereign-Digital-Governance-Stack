package resilience

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxAttempts    int           `yaml:"max_attempts"`
	InitialDelay   time.Duration `yaml:"initial_delay"`
	MaxDelay       time.Duration `yaml:"max_delay"`
	Multiplier     float64       `yaml:"multiplier"`
	Jitter         bool          `yaml:"jitter"`
	RetryableCodes []string      `yaml:"retryable_codes"`
}

// RetryResult holds the result of a retry operation
type RetryResult struct {
	Attempts   int
	LastError  error
	Duration   time.Duration
	Success    bool
}

// RetryPolicy manages retry behavior
type RetryPolicy struct {
	config  RetryConfig
	metrics *MetricsCollector
	logger  *zap.Logger
}

// NewRetryPolicy creates a new retry policy
func NewRetryPolicy(config RetryConfig, metrics *MetricsCollector, logger *zap.Logger) *RetryPolicy {
	return &RetryPolicy{
		config:  config,
		metrics: metrics,
		logger:  logger,
	}
}

// Execute retries a function with exponential backoff
func (p *RetryPolicy) Execute(
	ctx context.Context,
	fn func(attempt int) (interface{}, error),
) *RetryResult {
	result := &RetryResult{
		Success: false,
	}

	startTime := time.Now()

	for attempt := 1; attempt <= p.config.MaxAttempts; attempt++ {
		result.Attempts = attempt

		resp, err := fn(attempt)

		if err == nil {
			result.Success = true
			result.Duration = time.Since(startTime)
			p.logger.Info("Retry operation succeeded",
				zap.Int("attempt", attempt),
				zap.Duration("duration", result.Duration))
			return result
		}

		result.LastError = err

		// Check if error is retryable
		if !p.isRetryable(err) {
			p.logger.Warn("Non-retryable error",
				zap.Int("attempt", attempt),
				zap.Error(err))
			return result
		}

		// Don't delay after the last attempt
		if attempt < p.config.MaxAttempts {
			delay := p.calculateDelay(attempt)
			p.logger.Warn("Retrying after error",
				zap.Int("attempt", attempt),
				zap.Duration("delay", delay),
				zap.Error(err))

			select {
			case <-ctx.Done():
				result.LastError = ctx.Err()
				return result
			case <-time.After(delay):
			}
		}
	}

	result.Duration = time.Since(startTime)
	p.logger.Error("Retry operation failed after all attempts",
		zap.Int("attempts", result.Attempts),
		zap.Duration("duration", result.Duration),
		zap.Error(result.LastError))

	if p.metrics != nil {
		p.metrics.RecordError("global", "retry_exhausted", "retry_policy")
	}

	return result
}

// ExecuteWithRecovery retries with recovery function on final failure
func (p *RetryPolicy) ExecuteWithRecovery(
	ctx context.Context,
	fn func(attempt int) (interface{}, error),
	recovery func(attempt int, lastError error) (interface{}, error),
) *RetryResult {
	result := p.Execute(ctx, fn)

	if !result.Success && recovery != nil {
		recoveryResult, err := recovery(result.Attempts, result.LastError)
		if err == nil {
			result.Success = true
			result.LastError = nil
			_ = recoveryResult // Use the recovered result
			p.logger.Info("Recovery succeeded after retry failure")
		}
	}

	return result
}

func (p *RetryPolicy) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	for _, code := range p.config.RetryableCodes {
		if contains(errStr, code) {
			return true
		}
	}

	// Default: network errors and timeouts are retryable
	return contains(errStr, "timeout") ||
		contains(errStr, "connection") ||
		contains(errStr, "network") ||
		contains(errStr, "EOF")
}

func (p *RetryPolicy) calculateDelay(attempt int) time.Duration {
	delay := float64(p.config.InitialDelay) * pow(p.config.Multiplier, float64(attempt-1))

	if delay > float64(p.config.MaxDelay) {
		delay = float64(p.config.MaxDelay)
	}

	if p.config.Jitter {
		// Add jitter: +/- 25%
		jitterRange := delay * 0.25
		delay = delay - jitterRange + (time.Duration(float64(time.Now().UnixNano())%uint64(2*jitterRange*float64(time.Second))) / float64(time.Second))
	}

	return time.Duration(delay)
}

func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:    3,
		InitialDelay:   100 * time.Millisecond,
		MaxDelay:       30 * time.Second,
		Multiplier:     2.0,
		Jitter:         true,
		RetryableCodes: []string{"timeout", "connection", "network", "EOF", "temporary"},
	}
}
