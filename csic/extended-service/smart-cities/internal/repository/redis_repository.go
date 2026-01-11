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

// --- Node Latest Reading Cache ---

// CacheNodeLatestReading caches the latest reading for a node
func (r *RedisRepository) CacheNodeLatestReading(ctx context.Context, nodeID string, reading interface{}, ttl time.Duration) error {
	data, err := json.Marshal(reading)
	if err != nil {
		return fmt.Errorf("failed to marshal reading: %w", err)
	}

	key := r.key("node", nodeID, "latest")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedNodeLatestReading retrieves cached latest reading for a node
func (r *RedisRepository) GetCachedNodeLatestReading(ctx context.Context, nodeID string) ([]byte, error) {
	key := r.key("node", nodeID, "latest")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// --- District Aggregates Cache ---

// CacheDistrictAggregates caches aggregated metrics for a district
func (r *RedisRepository) CacheDistrictAggregates(ctx context.Context, districtID string, aggregates interface{}, ttl time.Duration) error {
	data, err := json.Marshal(aggregates)
	if err != nil {
		return fmt.Errorf("failed to marshal aggregates: %w", err)
	}

	key := r.key("district", districtID, "aggregates")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedDistrictAggregates retrieves cached district aggregates
func (r *RedisRepository) GetCachedDistrictAggregates(ctx context.Context, districtID string) ([]byte, error) {
	key := r.key("district", districtID, "aggregates")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// IncrementDistrictMetric increments a metric counter for a district
func (r *RedisRepository) IncrementDistrictMetric(ctx context.Context, districtID, metric string, value float64) error {
	key := r.key("district", districtID, "metrics", metric)
	return r.client.IncrByFloat(ctx, key, value).Err()
}

// GetDistrictMetric retrieves a metric value for a district
func (r *RedisRepository) GetDistrictMetric(ctx context.Context, districtID, metric string) (decimal.Decimal, error) {
	key := r.key("district", districtID, "metrics", metric)
	value, err := r.client.Get(ctx, key).Float64()
	if err == redis.Nil {
		return decimal.Zero, nil
	}
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromFloat(value), nil
}

// --- Incident Cache ---

// CacheIncidents caches recent incidents
func (r *RedisRepository) CacheIncidents(ctx context.Context, incidents interface{}, ttl time.Duration) error {
	data, err := json.Marshal(incidents)
	if err != nil {
		return fmt.Errorf("failed to marshal incidents: %w", err)
	}

	key := r.key("incidents", "recent")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedIncidents retrieves cached incidents
func (r *RedisRepository) GetCachedIncidents(ctx context.Context) ([]byte, error) {
	key := r.key("incidents", "recent")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// IncrementIncidentCounter increments the incident counter for a district
func (r *RedisRepository) IncrementIncidentCounter(ctx context.Context, districtID, incidentType string) error {
	key := r.key("incidents", "counter", districtID, incidentType)
	return r.client.Incr(ctx, key).Err()
}

// GetIncidentCounter retrieves the incident count for a district
func (r *RedisRepository) GetIncidentCounter(ctx context.Context, districtID, incidentType string) (int64, error) {
	key := r.key("incidents", "counter", districtID, incidentType)
	count, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// --- AQI Cache ---

// CacheAQIData caches AQI data for a district
func (r *RedisRepository) CacheAQIData(ctx context.Context, districtID string, aqi interface{}, ttl time.Duration) error {
	data, err := json.Marshal(aqi)
	if err != nil {
		return fmt.Errorf("failed to marshal AQI data: %w", err)
	}

	key := r.key("aqi", districtID, "current")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedAQIData retrieves cached AQI data
func (r *RedisRepository) GetCachedAQIData(ctx context.Context, districtID string) ([]byte, error) {
	key := r.key("aqi", districtID, "current")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// --- Traffic Status Cache ---

// CacheTrafficStatus caches traffic status for a location
func (r *RedisRepository) CacheTrafficStatus(ctx context.Context, locationID string, status interface{}, ttl time.Duration) error {
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal traffic status: %w", err)
	}

	key := r.key("traffic", locationID, "status")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedTrafficStatus retrieves cached traffic status
func (r *RedisRepository) GetCachedTrafficStatus(ctx context.Context, locationID string) ([]byte, error) {
	key := r.key("traffic", locationID, "status")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// --- Resource Usage Cache ---

// CacheResourceUsage caches resource usage for a district
func (r *RedisRepository) CacheResourceUsage(ctx context.Context, districtID string, usage interface{}, ttl time.Duration) error {
	data, err := json.Marshal(usage)
	if err != nil {
		return fmt.Errorf("failed to marshal resource usage: %w", err)
	}

	key := r.key("resources", districtID, "current")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedResourceUsage retrieves cached resource usage
func (r *RedisRepository) GetCachedResourceUsage(ctx context.Context, districtID string) ([]byte, error) {
	key := r.key("resources", districtID, "current")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
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

// --- Real-time Publishing ---

// PublishIoTEvent publishes an IoT event to subscribers
func (r *RedisRepository) PublishIoTEvent(ctx context.Context, channel string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return r.client.Publish(ctx, channel, data).Err()
}

// SubscribeIoT subscribes to IoT events
func (r *RedisRepository) SubscribeIoT(ctx context.Context, channel string) *redis.PubSub {
	pubsub := r.client.Subscribe(ctx, channel)
	return pubsub
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

// --- Node Status Cache ---

// CacheNodeStatus caches the status of a node
func (r *RedisRepository) CacheNodeStatus(ctx context.Context, nodeID string, status interface{}, ttl time.Duration) error {
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal node status: %w", err)
	}

	key := r.key("node", nodeID, "status")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedNodeStatus retrieves cached node status
func (r *RedisRepository) GetCachedNodeStatus(ctx context.Context, nodeID string) ([]byte, error) {
	key := r.key("node", nodeID, "status")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// --- Sustainability Metrics Cache ---

// CacheSustainabilityMetrics caches sustainability metrics
func (r *RedisRepository) CacheSustainabilityMetrics(ctx context.Context, districtID string, metrics interface{}, ttl time.Duration) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal sustainability metrics: %w", err)
	}

	key := r.key("sustainability", districtID, "current")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedSustainabilityMetrics retrieves cached sustainability metrics
func (r *RedisRepository) GetCachedSustainabilityMetrics(ctx context.Context, districtID string) ([]byte, error) {
	key := r.key("sustainability", districtID, "current")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
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
