package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// StateManager manages policy state using Redis
type StateManager struct {
	client    *redis.Client
	keyPrefix string
}

// NewStateManager creates a new state manager
func NewStateManager(redisURL string) (*StateManager, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisURL,
		DB:   0,
	})
	
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	
	return &StateManager{
		client:    client,
		keyPrefix: "neam:policy:",
	}, nil
}

// Close closes the Redis connection
func (s *StateManager) Close() error {
	return s.client.Close()
}

// SetState sets a state value with TTL
func (s *StateManager) SetState(key string, value interface{}, ttl time.Duration) error {
	ctx := context.Background()
	
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	
	fullKey := s.keyPrefix + key
	
	if ttl > 0 {
		return s.client.Set(ctx, fullKey, data, ttl).Err()
	}
	
	return s.client.Set(ctx, fullKey, data, 0).Err()
}

// GetState gets a state value
func (s *StateManager) GetState(key string) (interface{}, error) {
	ctx := context.Background()
	
	fullKey := s.keyPrefix + key
	data, err := s.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}
	
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}
	
	return value, nil
}

// GetString gets a string state value
func (s *StateManager) GetString(key string) (string, error) {
	ctx := context.Background()
	
	fullKey := s.keyPrefix + key
	val, err := s.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	
	return val, nil
}

// GetFloat gets a float64 state value
func (s *StateManager) GetFloat(key string) (float64, error) {
	ctx := context.Background()
	
	fullKey := s.keyPrefix + key
	val, err := s.client.Get(ctx, fullKey).Float64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	
	return val, nil
}

// GetInt gets an int64 state value
func (s *StateManager) GetInt(key string) (int64, error) {
	ctx := context.Background()
	
	fullKey := s.keyPrefix + key
	val, err := s.client.Get(ctx, fullKey).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	
	return val, nil
}

// DeleteState deletes a state value
func (s *StateManager) DeleteState(key string) error {
	ctx := context.Background()
	
	fullKey := s.keyPrefix + key
	return s.client.Del(ctx, fullKey).Err()
}

// IncrementCounter increments a counter
func (s *StateManager) IncrementCounter(key string) (int64, error) {
	ctx := context.Background()
	
	fullKey := s.keyPrefix + "counter:" + key
	return s.client.Incr(ctx, fullKey).Result()
}

// GetCounter gets a counter value
func (s *StateManager) GetCounter(key string) (int64, error) {
	ctx := context.Background()
	
	fullKey := s.keyPrefix + "counter:" + key
	val, err := s.client.Get(ctx, fullKey).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	
	return val, nil
}

// GetMetricValue gets a metric value from Redis
func (s *StateManager) GetMetricValue(metric string, regionID string, sectorID string) (float64, error) {
	ctx := context.Background()
	
	// Build metric key
	key := s.buildMetricKey(metric, regionID, sectorID)
	
	// Try to get from Redis
	val, err := s.client.Get(ctx, key).Result()
	if err == nil {
		return strconv.ParseFloat(val, 64)
	}
	
	// Fallback to aggregated metric
	if regionID != "" {
		key = s.buildMetricKey(metric, regionID, "")
		val, err = s.client.Get(ctx, key).Result()
		if err == nil {
			return strconv.ParseFloat(val, 64)
		}
	}
	
	// Fallback to global metric
	key = s.buildMetricKey(metric, "", "")
	val, err = s.client.Get(ctx, key).Result()
	if err == nil {
		return strconv.ParseFloat(val, 64)
	}
	
	// Return 0 if not found
	return 0, nil
}

// buildMetricKey builds a Redis key for a metric
func (s *StateManager) buildMetricKey(metric string, regionID string, sectorID string) string {
	parts := []string{s.keyPrefix + "metric", metric}
	
	if regionID != "" {
		parts = append(parts, regionID)
	}
	if sectorID != "" {
		parts = append(parts, sectorID)
	}
	
	return joinKeys(parts...)
}

// SetMetricValue sets a metric value in Redis
func (s *StateManager) SetMetricValue(metric string, regionID string, sectorID string, value float64, ttl time.Duration) error {
	ctx := context.Background()
	
	key := s.buildMetricKey(metric, regionID, sectorID)
	
	if ttl > 0 {
		return s.client.Set(ctx, key, fmt.Sprintf("%f", value), ttl).Err()
	}
	
	return s.client.Set(ctx, key, fmt.Sprintf("%f", value), 0).Err()
}

// IncrementMetric increments a metric value
func (s *StateManager) IncrementMetric(metric string, regionID string, sectorID string, delta float64) error {
	ctx := context.Background()
	
	key := s.buildMetricKey(metric, regionID, sectorID)
	
	pipe := s.client.Pipeline()
	pipe.IncrByFloat(ctx, key, delta)
	pipe.Expire(ctx, key, 24*time.Hour) // Default TTL for metrics
	_, err := pipe.Exec(ctx)
	
	return err
}

// GetTimeSeries gets a time series of metric values
func (s *StateManager) GetTimeSeries(metric string, regionID string, sectorID string, from time.Time, to time.Time) ([]TimeSeriesValue, error) {
	ctx := context.Background()
	
	pattern := s.buildMetricKey(metric, regionID, sectorID) + ":*"
	
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	
	values := make([]TimeSeriesValue, 0, len(keys))
	
	for _, key := range keys {
		data, err := s.client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}
		
		if val, ok := data["value"]; ok {
			if timestamp, ok := data["timestamp"]; ok {
				t, _ := time.Parse(time.RFC3339, timestamp)
				floatVal, _ := strconv.ParseFloat(val, 64)
				
				if t.After(from) && t.Before(to) {
					values = append(values, TimeSeriesValue{
						Timestamp: t,
						Value:     floatVal,
					})
				}
			}
		}
	}
	
	return values, nil
}

// PushTimeSeriesValue pushes a value to a time series
func (s *StateManager) PushTimeSeriesValue(metric string, regionID string, sectorID string, value float64) error {
	ctx := context.Background()
	
	timestamp := time.Now().Format(time.RFC3339)
	key := s.buildMetricKey(metric, regionID, sectorID) + ":" + strconv.FormatInt(time.Now().Unix(), 10)
	
	pipe := s.client.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"value":     fmt.Sprintf("%f", value),
		"timestamp": timestamp,
	})
	pipe.Expire(ctx, key, 7*24*time.Hour) // Keep 7 days of time series
	_, err := pipe.Exec(ctx)
	
	return err
}

// SetAlertLevel sets the alert level for a region/sector
func (s *StateManager) SetAlertLevel(regionID string, level string) error {
	key := s.keyPrefix + "alert_level:" + regionID
	return s.SetState(key, level, 24*time.Hour)
}

// GetAlertLevel gets the alert level for a region
func (s *StateManager) GetAlertLevel(regionID string) (string, error) {
	key := s.keyPrefix + "alert_level:" + regionID
	return s.GetString(key)
}

// Lock acquires a distributed lock
func (s *StateManager) Lock(lockName string, ttl time.Duration) (bool, error) {
	ctx := context.Background()
	
	fullKey := s.keyPrefix + "lock:" + lockName
	return s.client.SetNX(ctx, fullKey, time.Now().Unix(), ttl).Result()
}

// Unlock releases a distributed lock
func (s *StateManager) Unlock(lockName string) error {
	ctx := context.Background()
	
	fullKey := s.keyPrefix + "lock:" + lockName
	return s.client.Del(ctx, fullKey).Err()
}

// GetPolicyState gets the current state of a policy
func (s *StateManager) GetPolicyState(policyID string) (*PolicyState, error) {
	key := s.keyPrefix + "policy_state:" + policyID
	
	data, err := s.GetState(key)
	if err != nil {
		return nil, err
	}
	
	if data == nil {
		return &PolicyState{
			PolicyID:      policyID,
			CurrentStatus: "INACTIVE",
		}, nil
	}
	
	jsonData, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("invalid state data type")
	}
	
	var state PolicyState
	if err := json.Unmarshal([]byte(jsonData), &state); err != nil {
		return nil, err
	}
	
	return &state, nil
}

// SetPolicyState sets the current state of a policy
func (s *StateManager) SetPolicyState(policyID string, status string) error {
	key := s.keyPrefix + "policy_state:" + policyID
	
	state := PolicyState{
		PolicyID:      policyID,
		CurrentStatus: status,
		LastUpdated:   time.Now(),
	}
	
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	
	return s.client.Set(context.Background(), key, data, 0).Err()
}

// GetInterventionState gets the current state of an intervention
func (s *StateManager) GetInterventionState(interventionID string) (*InterventionState, error) {
	key := s.keyPrefix + "intervention_state:" + interventionID
	
	data, err := s.GetState(key)
	if err != nil {
		return nil, err
	}
	
	if data == nil {
		return nil, nil
	}
	
	jsonData, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("invalid state data type")
	}
	
	var state InterventionState
	if err := json.Unmarshal([]byte(jsonData), &state); err != nil {
		return nil, err
	}
	
	return &state, nil
}

// SetInterventionState sets the current state of an intervention
func (s *StateManager) SetInterventionState(interventionID string, status string, progress float64) error {
	key := s.keyPrefix + "intervention_state:" + interventionID
	
	state := InterventionState{
		InterventionID: interventionID,
		CurrentStatus:  status,
		Progress:       progress,
		LastUpdated:    time.Now(),
	}
	
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	
	return s.client.Set(context.Background(), key, data, 0).Err()
}

// GetDashboardMetrics gets aggregated metrics for dashboard
func (s *StateManager) GetDashboardMetrics() (*DashboardMetrics, error) {
	ctx := context.Background()
	
	metrics := &DashboardMetrics{
		UpdatedAt: time.Now(),
	}
	
	// Get active alerts count
	alertsKey := s.keyPrefix + "counter:active_alerts"
	alerts, _ := s.client.Get(ctx, alertsKey).Int64()
	metrics.ActiveAlerts = int(alerts)
	
	// Get active interventions count
	interventionsKey := s.keyPrefix + "counter:active_interventions"
	interventions, _ := s.client.Get(ctx, interventionsKey).Int64()
	metrics.ActiveInterventions = int(interventions)
	
	// Get regions in alert
	pattern := s.keyPrefix + "alert_level:*"
	keys, _ := s.client.Keys(ctx, pattern).Result()
	metrics.RegionsWithAlerts = len(keys)
	
	// Get system health
	healthKey := s.keyPrefix + "system_health"
	health, _ := s.client.Get(ctx, healthKey).Result()
	metrics.SystemHealth = health
	
	return metrics, nil
}

// SetDashboardMetrics updates dashboard metrics
func (s *StateManager) SetDashboardMetrics(metrics *DashboardMetrics) error {
	ctx := context.Background()
	
	pipe := s.client.Pipeline()
	
	pipe.Set(ctx, s.keyPrefix+"counter:active_alerts", metrics.ActiveAlerts, 0)
	pipe.Set(ctx, s.keyPrefix+"counter:active_interventions", metrics.ActiveInterventions, 0)
	pipe.Set(ctx, s.keyPrefix+"system_health", metrics.SystemHealth, 0)
	
	_, err := pipe.Exec(ctx)
	return err
}

// Helper functions

func joinKeys(parts ...string) string {
	result := parts[0]
	for _, part := range parts[1:] {
		result += ":" + part
	}
	return result
}

// TimeSeriesValue represents a value in a time series
type TimeSeriesValue struct {
	Timestamp time.Time
	Value     float64
}

// PolicyState represents the state of a policy
type PolicyState struct {
	PolicyID      string    `json:"policy_id"`
	CurrentStatus string    `json:"current_status"`
	LastUpdated   time.Time `json:"last_updated"`
	TriggerCount  int       `json:"trigger_count"`
}

// InterventionState represents the state of an intervention
type InterventionState struct {
	InterventionID string    `json:"intervention_id"`
	CurrentStatus  string    `json:"current_status"`
	Progress       float64   `json:"progress"`
	LastUpdated    time.Time `json:"last_updated"`
}

// DashboardMetrics holds aggregated dashboard metrics
type DashboardMetrics struct {
	ActiveAlerts        int       `json:"active_alerts"`
	ActiveInterventions int       `json:"active_interventions"`
	RegionsWithAlerts   int       `json:"regions_with_alerts"`
	SystemHealth        string    `json:"system_health"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// HealthCheck checks Redis connectivity
func (s *StateManager) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	return s.client.Ping(ctx).Err()
}

// SetHealth sets the system health status
func (s *StateManager) SetHealth(status string) error {
	key := s.keyPrefix + "system_health"
	return s.SetState(key, status, 0)
}

// PubSubSubscribe subscribes to a channel
func (s *StateManager) PubSubSubscribe(channel string) *redis.PubSub {
	return s.client.Subscribe(context.Background(), s.keyPrefix+channel)
}

// PubSubPublish publishes to a channel
func (s *StateManager) PubSubPublish(channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	return s.client.Publish(context.Background(), s.keyPrefix+channel, data).Err()
}

// GetRegionMetrics gets all metrics for a region
func (s *StateManager) GetRegionMetrics(regionID string) (map[string]float64, error) {
	ctx := context.Background()
	
	pattern := s.keyPrefix + "metric:*:" + regionID + ":*"
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	
	metrics := make(map[string]float64)
	
	for _, key := range keys {
		val, err := s.client.Get(ctx, key).Float64()
		if err == nil {
			// Extract metric name from key
			parts := splitKeys(key)
			if len(parts) >= 3 {
				metricName := parts[2]
				metrics[metricName] = val
			}
		}
	}
	
	return metrics, nil
}

func splitKeys(key string) []string {
	var parts []string
	current := ""
	
	for _, c := range key {
		if c == ':' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	
	parts = append(parts, current)
	return parts
}

func (s *StateManager) Log(ctx context.Context, format string, args ...interface{}) {
	log.Printf("[StateManager] "+format, args...)
}
