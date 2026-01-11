package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"csic-platform/control-layer/internal/core/domain"
	"csic-platform/control-layer/internal/core/ports"
	"csic-platform/control-layer/pkg/metrics"
)

// StateRegistry manages system state
type StateRegistry interface {
	GetState(ctx context.Context, key string) (*domain.State, error)
	UpdateState(ctx context.Context, key string, value string, ttl time.Duration) (*domain.State, error)
	DeleteState(ctx context.Context, key string) error
	ListStates(ctx context.Context) ([]*domain.State, error)
	UpdateStateFromEvent(ctx context.Context, event *domain.StateUpdate) error
}

// StateRegistryService implements the StateRegistry interface
type StateRegistryService struct {
	repositories ports.Repositories
	cachePort    ports.CachePort
	logger       *zap.Logger
	metrics      *metrics.MetricsCollector
}

// NewStateRegistry creates a new state registry
func NewStateRegistry(
	repositories ports.Repositories,
	cachePort ports.CachePort,
	logger *zap.Logger,
	metricsCollector *metrics.MetricsCollector,
) StateRegistry {
	return &StateRegistryService{
		repositories: repositories,
		cachePort:    cachePort,
		logger:       logger,
		metrics:      metricsCollector,
	}
}

// GetState gets a state value by key
func (s *StateRegistryService) GetState(ctx context.Context, key string) (*domain.State, error) {
	value, err := s.cachePort.Client.GetState(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	if value == "" {
		return nil, nil
	}

	s.metrics.UpdateStateCacheHit()

	return &domain.State{
		Key:   key,
		Value: value,
	}, nil
}

// UpdateState updates a state value
func (s *StateRegistryService) UpdateState(ctx context.Context, key string, value string, ttl time.Duration) (*domain.State, error) {
	if err := s.cachePort.Client.SetState(ctx, key, value, ttl); err != nil {
		return nil, fmt.Errorf("failed to set state: %w", err)
	}

	s.metrics.StateUpdatesTotal.WithLabelValues("state").Inc()

	state := &domain.State{
		Key:   key,
		Value: value,
	}

	s.logger.Debug("Updated state",
		logger.String("key", key),
	)

	return state, nil
}

// DeleteState deletes a state value
func (s *StateRegistryService) DeleteState(ctx context.Context, key string) error {
	if err := s.cachePort.Client.DeleteState(ctx, key); err != nil {
		return fmt.Errorf("failed to delete state: %w", err)
	}

	s.logger.Debug("Deleted state", logger.String("key", key))

	return nil
}

// ListStates lists all states (with a prefix)
func (s *StateRegistryService) ListStates(ctx context.Context) ([]*domain.State, error) {
	// In a real implementation, this would use SCAN to iterate over keys
	// For now, return empty list
	return []*domain.State{}, nil
}

// UpdateStateFromEvent updates state from a telemetry or state update event
func (s *StateRegistryService) UpdateStateFromEvent(ctx context.Context, event *domain.StateUpdate) error {
	// Convert event data to JSON for storage
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal state data: %w", err)
	}

	// Set state with appropriate TTL based on service type
	ttl := 5 * time.Minute
	if event.ServiceType == "heartbeat" {
		ttl = 2 * time.Minute
	}

	key := fmt.Sprintf("service:%s:state", event.ServiceName)
	if err := s.cachePort.Client.SetState(ctx, key, string(data), ttl); err != nil {
		return fmt.Errorf("failed to set state from event: %w", err)
	}

	// Also update last_seen for heartbeat events
	if event.ServiceType == "heartbeat" {
		lastSeenKey := fmt.Sprintf("service:%s:last_seen", event.ServiceName)
		if err := s.cachePort.Client.SetState(ctx, lastSeenKey, time.Now().Format(time.RFC3339), 2*time.Minute); err != nil {
			return fmt.Errorf("failed to update last_seen: %w", err)
		}
	}

	s.metrics.StateUpdatesTotal.WithLabelValues(event.ServiceType).Inc()

	s.logger.Debug("Updated state from event",
		logger.String("service", event.ServiceName),
		logger.String("type", event.ServiceType),
	)

	return nil
}

// GetServiceHealth checks if a service is healthy based on last seen timestamp
func (s *StateRegistryService) GetServiceHealth(ctx context.Context, serviceName string, maxAge time.Duration) (bool, time.Time, error) {
	lastSeenKey := fmt.Sprintf("service:%s:last_seen", serviceName)

	lastSeenStr, err := s.cachePort.Client.GetState(ctx, lastSeenKey)
	if err != nil {
		return false, time.Time{}, fmt.Errorf("failed to get last_seen: %w", err)
	}

	if lastSeenStr == "" {
		return false, time.Time{}, nil
	}

	lastSeen, err := time.Parse(time.RFC3339, lastSeenStr)
	if err != nil {
		return false, time.Time{}, fmt.Errorf("failed to parse last_seen: %w", err)
	}

	healthy := time.Since(lastSeen) < maxAge

	return healthy, lastSeen, nil
}

// GetAllServiceHealth returns health status for all registered services
func (s *StateRegistryService) GetAllServiceHealth(ctx context.Context, maxAge time.Duration) (map[string]domain.ServiceHealth, error) {
	// In a real implementation, this would iterate over all service keys
	return make(map[string]domain.ServiceHealth), nil
}
