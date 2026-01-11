package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
)

// RedisRepository handles all Redis cache operations
type RedisRepository struct {
	client    *redis.Client
	keyPrefix string
}

// RedisConfig matches the config.Redis structure
type RedisConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Password  string `yaml:"password"`
	DB        int    `yaml:"db"`
	KeyPrefix string `yaml:"key_prefix"`
	PoolSize  int    `yaml:"pool_size"`
}

// NewRedisClient creates a new Redis client
func NewRedisClient(config *RedisConfig) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisRepository{
		client:    client,
		keyPrefix: config.KeyPrefix,
	}, nil
}

// Close closes the Redis connection
func (r *RedisRepository) Close() error {
	return r.client.Close()
}

// key returns the full key with prefix
func (r *RedisRepository) key(parts ...string) string {
	result := r.keyPrefix
	for _, part := range parts {
		result += part + ":"
	}
	return result[:len(result)-1] // Remove trailing colon
}

// --- Unit Location Cache (Geo-Spatial) ---

// CacheUnitLocation caches the current location of a unit for geo-queries
func (r *RedisRepository) CacheUnitLocation(ctx context.Context, unitID string, lat, lng float64) error {
	key := r.key("unit_locations")
	return r.client.GeoAdd(ctx, key, &redis.GeoLocation{
		Name:      unitID,
		Longitude: lng,
		Latitude:  lat,
	}).Err()
}

// GetNearbyUnits returns units within a radius of a location
func (r *RedisRepository) GetNearbyUnits(ctx context.Context, lat, lng float64, radius float64) ([]string, error) {
	key := r.key("unit_locations")
	results, err := r.client.GeoRadius(ctx, key, lng, lat, &redis.GeoRadiusQuery{
		Radius:     radius,
		Unit:       "km",
		WithCoord:  false,
		WithDist:   true,
		WithGeoHash: false,
		Count:      100,
		Sort:       "ASC",
	}).Result()
	if err != nil {
		return nil, err
	}

	var unitIDs []string
	for _, loc := range results {
		unitIDs = append(unitIDs, loc.Name)
	}

	return unitIDs, nil
}

// --- Telemetry Deduplication ---

// CheckAndSetDedupKey checks if a telemetry packet is a duplicate
func (r *RedisRepository) CheckAndSetDedupKey(ctx context.Context, hash string, ttl time.Duration) (bool, error) {
	key := r.key("dedup", hash)
	return r.client.SetNX(ctx, key, time.Now().Unix(), ttl).Result()
}

// --- Unit Status Cache ---

// CacheUnitStatus caches the current status of a unit
func (r *RedisRepository) CacheUnitStatus(ctx context.Context, unitID string, status interface{}, ttl time.Duration) error {
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	key := r.key("unit", unitID, "status")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedUnitStatus retrieves cached unit status
func (r *RedisRepository) GetCachedUnitStatus(ctx context.Context, unitID string) ([]byte, error) {
	key := r.key("unit", unitID, "status")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// --- Latest Telemetry Cache ---

// CacheLatestTelemetry caches the latest telemetry for a unit
func (r *RedisRepository) CacheLatestTelemetry(ctx context.Context, unitID string, telemetry interface{}, ttl time.Duration) error {
	data, err := json.Marshal(telemetry)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry: %w", err)
	}

	key := r.key("unit", unitID, "telemetry", "latest")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedLatestTelemetry retrieves cached latest telemetry
func (r *RedisRepository) GetCachedLatestTelemetry(ctx context.Context, unitID string) ([]byte, error) {
	key := r.key("unit", unitID, "telemetry", "latest")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// --- Active Threats Cache ---

// CacheActiveThreats caches active threats for quick access
func (r *RedisRepository) CacheActiveThreats(ctx context.Context, threats interface{}, ttl time.Duration) error {
	data, err := json.Marshal(threats)
	if err != nil {
		return fmt.Errorf("failed to marshal threats: %w", err)
	}

	key := r.key("threats", "active")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedActiveThreats retrieves cached active threats
func (r *RedisRepository) GetCachedActiveThreats(ctx context.Context) ([]byte, error) {
	key := r.key("threats", "active")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// AddThreatToActiveSet adds a threat ID to the active threats set
func (r *RedisRepository) AddThreatToActiveSet(ctx context.Context, threatID string) error {
	key := r.key("threats", "active_set")
	return r.client.SAdd(ctx, key, threatID).Err()
}

// RemoveThreatFromActiveSet removes a threat ID from the active threats set
func (r *RedisRepository) RemoveThreatFromActiveSet(ctx context.Context, threatID string) error {
	key := r.key("threats", "active_set")
	return r.client.SRem(ctx, key, threatID).Err()
}

// GetActiveThreatCount returns the count of active threats
func (r *RedisRepository) GetActiveThreatCount(ctx context.Context) (int64, error) {
	key := r.key("threats", "active_set")
	return r.client.SCard(ctx, key).Result()
}

// --- Alerts Cache ---

// CacheAlerts caches recent alerts
func (r *RedisRepository) CacheAlerts(ctx context.Context, alerts interface{}, ttl time.Duration) error {
	data, err := json.Marshal(alerts)
	if err != nil {
		return fmt.Errorf("failed to marshal alerts: %w", err)
	}

	key := r.key("alerts", "recent")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedAlerts retrieves cached alerts
func (r *RedisRepository) GetCachedAlerts(ctx context.Context) ([]byte, error) {
	key := r.key("alerts", "recent")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// IncrementAlertCounter increments the alert counter for a severity level
func (r *RedisRepository) IncrementAlertCounter(ctx context.Context, severity string) error {
	key := r.key("alerts", "counter", severity)
	return r.client.Incr(ctx, key).Err()
}

// GetAlertCounter returns the alert count for a severity level
func (r *RedisRepository) GetAlertCounter(ctx context.Context, severity string) (int64, error) {
	key := r.key("alerts", "counter", severity)
	return r.client.Get(ctx, key).Int64()
}

// --- Equipment Health Cache ---

// CacheEquipmentHealth caches equipment health scores
func (r *RedisRepository) CacheEquipmentHealth(ctx context.Context, unitID string, health map[string]decimal.Decimal, ttl time.Duration) error {
	data, err := json.Marshal(health)
	if err != nil {
		return fmt.Errorf("failed to marshal health: %w", err)
	}

	key := r.key("unit", unitID, "health")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedEquipmentHealth retrieves cached equipment health
func (r *RedisRepository) GetCachedEquipmentHealth(ctx context.Context, unitID string) ([]byte, error) {
	key := r.key("unit", unitID, "health")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// --- Deployment Cache ---

// CacheDeployment caches deployment information
func (r *RedisRepository) CacheDeployment(ctx context.Context, deploymentID string, deployment interface{}, ttl time.Duration) error {
	data, err := json.Marshal(deployment)
	if err != nil {
		return fmt.Errorf("failed to marshal deployment: %w", err)
	}

	key := r.key("deployment", deploymentID)
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedDeployment retrieves cached deployment
func (r *RedisRepository) GetCachedDeployment(ctx context.Context, deploymentID string) ([]byte, error) {
	key := r.key("deployment", deploymentID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// --- Maintenance Predictions Cache ---

// CacheMaintenancePredictions caches maintenance predictions
func (r *RedisRepository) CacheMaintenancePredictions(ctx context.Context, unitID string, predictions interface{}, ttl time.Duration) error {
	data, err := json.Marshal(predictions)
	if err != nil {
		return fmt.Errorf("failed to marshal predictions: %w", err)
	}

	key := r.key("unit", unitID, "maintenance_predictions")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedMaintenancePredictions retrieves cached maintenance predictions
func (r *RedisRepository) GetCachedMaintenancePredictions(ctx context.Context, unitID string) ([]byte, error) {
	key := r.key("unit", unitID, "maintenance_predictions")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// --- Zone Status Cache ---

// CacheZoneStatus caches zone status information
func (r *RedisRepository) CacheZoneStatus(ctx context.Context, zoneID string, status interface{}, ttl time.Duration) error {
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal zone status: %w", err)
	}

	key := r.key("zone", zoneID, "status")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedZoneStatus retrieves cached zone status
func (r *RedisRepository) GetCachedZoneStatus(ctx context.Context, zoneID string) ([]byte, error) {
	key := r.key("zone", zoneID, "status")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// IncrementZoneUnitCount increments the unit count for a zone
func (r *RedisRepository) IncrementZoneUnitCount(ctx context.Context, zoneID string) error {
	key := r.key("zone", zoneID, "unit_count")
	return r.client.Incr(ctx, key).Err()
}

// DecrementZoneUnitCount decrements the unit count for a zone
func (r *RedisRepository) DecrementZoneUnitCount(ctx context.Context, zoneID string) error {
	key := r.key("zone", zoneID, "unit_count")
	return r.client.Decr(ctx, key).Err()
}

// GetZoneUnitCount returns the current unit count for a zone
func (r *RedisRepository) GetZoneUnitCount(ctx context.Context, zoneID string) (int64, error) {
	key := r.key("zone", zoneID, "unit_count")
	count, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// --- Rate Limiting ---

// CheckRateLimit checks if a request exceeds the rate limit
func (r *RedisRepository) CheckRateLimit(ctx context.Context, identifier string, limit int, window time.Duration) (bool, error) {
	key := r.key("ratelimit", identifier)
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		r.client.Expire(ctx, key, window)
	}

	return count > int64(limit), nil
}

// --- Locking ---

// AcquireLock attempts to acquire a distributed lock
func (r *RedisRepository) AcquireLock(ctx context.Context, lockName string, ttl time.Duration) (bool, error) {
	key := r.key("lock", lockName)
	return r.client.SetNX(ctx, key, time.Now().Unix(), ttl).Result()
}

// ReleaseLock releases a distributed lock
func (r *RedisRepository) ReleaseLock(ctx context.Context, lockName string) error {
	key := r.key("lock", lockName)
	return r.client.Del(ctx, key).Err()
}

// --- Real-time Publishing ---

// PublishTelemetryEvent publishes a telemetry event to subscribers
func (r *RedisRepository) PublishTelemetryEvent(ctx context.Context, channel string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return r.client.Publish(ctx, channel, data).Err()
}

// SubscribeTelemetry subscribes to telemetry events
func (r *RedisRepository) SubscribeTelemetry(ctx context.Context, channel string) *redis.PubSub {
	pubsub := r.client.Subscribe(ctx, channel)
	return pubsub
}

// --- Health Check ---

// Ping checks if Redis is available
func (r *RedisRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// --- Counter for Metrics ---

// IncrementCounter increments a counter for metrics
func (r *RedisRepository) IncrementCounter(ctx context.Context, counterName string) error {
	key := r.key("metrics", "counter", counterName)
	return r.client.Incr(ctx, key).Err()
}

// GetCounter returns the value of a counter
func (r *RedisRepository) GetCounter(ctx context.Context, counterName string) (int64, error) {
	key := r.key("metrics", "counter", counterName)
	value, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return value, err
}

// ResetCounter resets a counter
func (r *RedisRepository) ResetCounter(ctx context.Context, counterName string) error {
	key := r.key("metrics", "counter", counterName)
	return r.client.Set(ctx, key, 0, 0).Err()
}
