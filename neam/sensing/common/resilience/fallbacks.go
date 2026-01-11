package resilience

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// FallbackRegistry manages fallback functions for different operations
type FallbackRegistry struct {
	fallbacks map[string]map[string]FallbackFunc
	logger    *zap.Logger
}

// NewFallbackRegistry creates a new fallback registry
func NewFallbackRegistry(logger *zap.Logger) *FallbackRegistry {
	return &FallbackRegistry{
		fallbacks: make(map[string]map[string]FallbackFunc),
		logger:    logger,
	}
}

// RegisterFallback registers a fallback function for an adapter and operation
func (r *FallbackRegistry) RegisterFallback(adapter, operation string, fn FallbackFunc) {
	if _, exists := r.fallbacks[adapter]; !exists {
		r.fallbacks[adapter] = make(map[string]FallbackFunc)
	}
	r.fallbacks[adapter][operation] = fn
	r.logger.Info("Fallback registered",
		zap.String("adapter", adapter),
		zap.String("operation", operation))
}

// GetFallback retrieves a fallback function
func (r *FallbackRegistry) GetFallback(adapter, operation string) (FallbackFunc, bool) {
	if ops, exists := r.fallbacks[adapter]; exists {
		if fn, exists := ops[operation]; exists {
			return fn, true
		}
	}
	return nil, false
}

// TransportFallbacks provides fallback behaviors for the transport adapter
type TransportFallbacks struct {
	logger *zap.Logger
}

// NewTransportFallbacks creates transport fallback functions
func NewTransportFallbacks(logger *zap.Logger) *TransportFallbacks {
	return &TransportFallbacks{logger: logger}
}

// CachedDataFallback returns cached vehicle telemetry data
func (f *TransportFallbacks) CachedDataFallback(ctx context.Context, operation string, err error) (interface{}, error) {
	f.logger.Warn("Transport fallback triggered",
		zap.String("operation", operation),
		zap.Error(err))

	// Return cached/empty telemetry data
	return map[string]interface{}{
		"fallback":     true,
		"operation":    operation,
		"cached":       false,
		"message":      "Circuit breaker open, using degraded mode",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"vehicle_data": []interface{}{},
	}, nil
}

// DegradedModeFallback provides degraded service mode
func (f *TransportFallbacks) DegradedModeFallback(ctx context.Context, operation string, err error) (interface{}, error) {
	f.logger.Warn("Transport degraded mode fallback",
		zap.String("operation", operation),
		zap.Error(err))

	return map[string]interface{}{
		"fallback":     true,
		"degraded":     true,
		"operation":    operation,
		"message":      "Service operating in degraded mode",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"available":    false,
	}, nil
}

// IndustrialFallbacks provides fallback behaviors for the industrial adapter
type IndustrialFallbacks struct {
	lastKnownValues map[string]interface{}
	logger          *zap.Logger
}

// CachedDataFallback returns cached industrial sensor data
func (f *IndustrialFallbacks) CachedDataFallback(ctx context.Context, operation string, err error) (interface{}, error) {
	f.logger.Warn("Industrial fallback triggered",
		zap.String("operation", operation),
		zap.Error(err))

	// Return cached/empty sensor data
	return map[string]interface{}{
		"fallback":   true,
		"operation":  operation,
		"cached":     false,
		"message":    "Circuit breaker open, using cached data",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"sensor_data": f.lastKnownValues,
	}, nil
}

// DegradedModeFallback provides degraded service mode for industrial operations
func (f *IndustrialFallbacks) DegradedModeFallback(ctx context.Context, operation string, err error) (interface{}, error) {
	f.logger.Warn("Industrial degraded mode fallback",
		zap.String("operation", operation),
		zap.Error(err))

	return map[string]interface{}{
		"fallback":   true,
		"degraded":   true,
		"operation":  operation,
		"message":    "Service operating in degraded mode",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"available":  false,
	}, nil
}

// NewIndustrialFallbacks creates industrial fallback functions
func NewIndustrialFallbacks(logger *zap.Logger) *IndustrialFallbacks {
	return &IndustrialFallbacks{
		lastKnownValues: make(map[string]interface{}),
		logger:          logger,
	}
}

// LastKnownValueFallback returns the last known good value for a sensor
func (f *IndustrialFallbacks) LastKnownValueFallback(ctx context.Context, operation string, err error) (interface{}, error) {
	f.logger.Warn("Industrial fallback to last known value",
		zap.String("operation", operation),
		zap.Error(err))

	// Return last known value with quality flag
	return map[string]interface{}{
		"fallback":   true,
		"source":     "last_known_value",
		"quality":    "Uncertain",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"value":      f.lastKnownValues[operation],
		"message":    "Using last known value, quality uncertain",
	}, nil
}

// DefaultValueFallback returns a default value for missing data
func (f *IndustrialFallbacks) DefaultValueFallback(ctx context.Context, operation string, err error) (interface{}, error) {
	f.logger.Warn("Industrial default value fallback",
		zap.String("operation", operation),
		zap.Error(err))

	return map[string]interface{}{
		"fallback":   true,
		"source":     "default_value",
		"quality":    "Bad",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"value":      0.0,
		"message":    "Using default value, quality bad",
	}, nil
}

// SetLastKnownValue updates the last known value for an operation
func (f *IndustrialFallbacks) SetLastKnownValue(operation string, value interface{}) {
	f.lastKnownValues[operation] = value
}

// EnergyFallbacks provides fallback behaviors for the energy adapter
type EnergyFallbacks struct {
	logger *zap.Logger
}

// CachedDataFallback returns cached grid telemetry data
func (f *EnergyFallbacks) CachedDataFallback(ctx context.Context, operation string, err error) (interface{}, error) {
	f.logger.Warn("Energy fallback triggered",
		zap.String("operation", operation),
		zap.Error(err))

	// Return cached/empty grid data
	return map[string]interface{}{
		"fallback":     true,
		"operation":    operation,
		"cached":       false,
		"message":      "Circuit breaker open, using cached grid data",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"grid_data":    map[string]interface{}{},
	}, nil
}

// DegradedModeFallback provides degraded service mode for energy operations
func (f *EnergyFallbacks) DegradedModeFallback(ctx context.Context, operation string, err error) (interface{}, error) {
	f.logger.Warn("Energy degraded mode fallback",
		zap.String("operation", operation),
		zap.Error(err))

	return map[string]interface{}{
		"fallback":     true,
		"degraded":     true,
		"operation":    operation,
		"message":      "Grid service operating in degraded mode",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"available":    false,
	}, nil
}

// NewEnergyFallbacks creates energy fallback functions
func NewEnergyFallbacks(logger *zap.Logger) *EnergyFallbacks {
	return &EnergyFallbacks{logger: logger}
}

// ReducedSamplingFallback implements reduced sampling rate fallback
func (f *EnergyFallbacks) ReducedSamplingFallback(ctx context.Context, operation string, err error) (interface{}, error) {
	f.logger.Warn("Energy reduced sampling fallback",
		zap.String("operation", operation),
		zap.Error(err))

	return map[string]interface{}{
		"fallback":       true,
		"sampling_rate":  "reduced",
		"interval":       "60s",
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"message":        "Operating at reduced sampling rate",
		"available":      false,
	}, nil
}

// ThrottledFallback implements throttled operation fallback
func (f *EnergyFallbacks) ThrottledFallback(ctx context.Context, operation string, err error) (interface{}, error) {
	f.logger.Warn("Energy throttled fallback",
		zap.String("operation", operation),
		zap.Error(err))

	return map[string]interface{}{
		"fallback":    true,
		"mode":        "throttled",
		"max_rate":    "1采样/分钟",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"message":     "Operation throttled due to circuit breaker",
		"retry_after": "30s",
	}, nil
}

// NullDataFallback returns null/empty data for energy readings
func (f *EnergyFallbacks) NullDataFallback(ctx context.Context, operation string, err error) (interface{}, error) {
	f.logger.Warn("Energy null data fallback",
		zap.String("operation", operation),
		zap.Error(err))

	return map[string]interface{}{
		"fallback":    true,
		"source":      "null_data",
		"quality":     "Bad",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"active_power": 0.0,
		"reactive_power": 0.0,
		"voltage":     0.0,
		"current":     0.0,
		"message":     "No data available, circuit breaker open",
	}, nil
}
