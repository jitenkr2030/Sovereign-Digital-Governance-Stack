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

// --- Energy Readings Cache ---

// CacheEnergyReading caches an energy reading
func (r *RedisRepository) CacheEnergyReading(ctx context.Context, reading *struct {
	ID             string
	MeterID        string
	FacilityID     string
	EnergySource   string
	Consumption    decimal.Decimal
	Timestamp      time.Time
}, ttl time.Duration) error {
	data, err := json.Marshal(reading)
	if err != nil {
		return fmt.Errorf("failed to marshal reading: %w", err)
	}

	key := r.key("reading", reading.ID)
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedEnergyReading retrieves a cached energy reading
func (r *RedisRepository) GetCachedEnergyReading(ctx context.Context, id string) (*struct {
	ID             string
	MeterID        string
	FacilityID     string
	EnergySource   string
	Consumption    decimal.Decimal
	Timestamp      time.Time
}, error) {
	key := r.key("reading", id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var reading struct {
		ID             string
		MeterID        string
		FacilityID     string
		EnergySource   string
		Consumption    decimal.Decimal
		Timestamp      time.Time
	}
	if err := json.Unmarshal(data, &reading); err != nil {
		return nil, err
	}

	return &reading, nil
}

// GetFacilityLatestReading returns the latest reading ID for a facility
func (r *RedisRepository) GetFacilityLatestReading(ctx context.Context, facilityID string) (string, error) {
	key := r.key("facility", facilityID, "latest_reading")
	return r.client.Get(ctx, key).Result()
}

// SetFacilityLatestReading sets the latest reading for a facility
func (r *RedisRepository) SetFacilityLatestReading(ctx context.Context, facilityID, readingID string, ttl time.Duration) error {
	key := r.key("facility", facilityID, "latest_reading")
	return r.client.Set(ctx, key, readingID, ttl).Err()
}

// --- Carbon Emissions Cache ---

// CacheCarbonEmissions caches carbon emission data
func (r *RedisRepository) CacheCarbonEmissions(ctx context.Context, facilityID string, emissions interface{}, ttl time.Duration) error {
	data, err := json.Marshal(emissions)
	if err != nil {
		return fmt.Errorf("failed to marshal emissions: %w", err)
	}

	key := r.key("carbon", facilityID, "daily")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedCarbonEmissions retrieves cached carbon emissions
func (r *RedisRepository) GetCachedCarbonEmissions(ctx context.Context, facilityID string) ([]byte, error) {
	key := r.key("carbon", facilityID, "daily")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// --- Load Metrics Cache ---

// CacheLoadMetrics caches load metrics for quick access
func (r *RedisRepository) CacheLoadMetrics(ctx context.Context, facilityID string, metrics interface{}, ttl time.Duration) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	key := r.key("load", facilityID, "current")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedLoadMetrics retrieves cached load metrics
func (r *RedisRepository) GetCachedLoadMetrics(ctx context.Context, facilityID string) ([]byte, error) {
	key := r.key("load", facilityID, "current")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// --- Carbon Budget Cache ---

// CacheCarbonBudget caches the current carbon budget
func (r *RedisRepository) CacheCarbonBudget(ctx context.Context, facilityID string, budget interface{}, ttl time.Duration) error {
	data, err := json.Marshal(budget)
	if err != nil {
		return fmt.Errorf("failed to marshal budget: %w", err)
	}

	key := r.key("budget", facilityID, "current")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedCarbonBudget retrieves a cached carbon budget
func (r *RedisRepository) GetCachedCarbonBudget(ctx context.Context, facilityID string) ([]byte, error) {
	key := r.key("budget", facilityID, "current")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// UpdateCarbonBudgetUsage atomically updates the used amount of a carbon budget
func (r *RedisRepository) UpdateCarbonBudgetUsage(ctx context.Context, facilityID string, additionalUsage decimal.Decimal) (decimal.Decimal, error) {
	key := r.key("budget", facilityID, "usage")

	// Increment the usage counter
	newValue, err := r.client.IncrByFloat(ctx, key, additionalUsage.InexactFloat64()).Result()
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromFloat(newValue), nil
}

// GetCarbonBudgetUsage gets the current carbon budget usage for a facility
func (r *RedisRepository) GetCarbonBudgetUsage(ctx context.Context, facilityID string) (decimal.Decimal, error) {
	key := r.key("budget", facilityID, "usage")
	value, err := r.client.Get(ctx, key).Float64()
	if err == redis.Nil {
		return decimal.Zero, nil
	}
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromFloat(value), nil
}

// ResetCarbonBudgetUsage resets the carbon budget usage for a new period
func (r *RedisRepository) ResetCarbonBudgetUsage(ctx context.Context, facilityID string) error {
	key := r.key("budget", facilityID, "usage")
	return r.client.Set(ctx, key, 0, 0).Err()
}

// --- Load Balancing Cache ---

// CacheLoadBalancingDecision caches a pending load balancing decision
func (r *RedisRepository) CacheLoadBalancingDecision(ctx context.Context, decisionID string, decision interface{}, ttl time.Duration) error {
	data, err := json.Marshal(decision)
	if err != nil {
		return fmt.Errorf("failed to marshal decision: %w", err)
	}

	key := r.key("lb_decision", decisionID)
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedLoadBalancingDecision retrieves a cached load balancing decision
func (r *RedisRepository) GetCachedLoadBalancingDecision(ctx context.Context, decisionID string) ([]byte, error) {
	key := r.key("lb_decision", decisionID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// GetPendingLoadBalancingDecisions retrieves all pending decisions for a facility
func (r *RedisRepository) GetPendingLoadBalancingDecisions(ctx context.Context, facilityID string) ([]string, error) {
	pattern := r.key("lb_decision", "*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var pendingIDs []string
	for _, key := range keys {
		// Get the decision data to check status
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var decision struct {
			FacilityID string
			Status     string
		}
		if err := json.Unmarshal(data, &decision); err != nil {
			continue
		}

		if decision.FacilityID == facilityID && decision.Status == "pending" {
			pendingIDs = append(pendingIDs, key)
		}
	}

	return pendingIDs, nil
}

// --- Facility Status Cache ---

// SetFacilityStatus caches the current status of a facility
func (r *RedisRepository) SetFacilityStatus(ctx context.Context, facilityID string, status interface{}, ttl time.Duration) error {
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	key := r.key("facility", facilityID, "status")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetFacilityStatus retrieves the cached status of a facility
func (r *RedisRepository) GetFacilityStatus(ctx context.Context, facilityID string) ([]byte, error) {
	key := r.key("facility", facilityID, "status")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// --- Zone Load Cache ---

// CacheZoneLoad caches the current load for a zone
func (r *RedisRepository) CacheZoneLoad(ctx context.Context, zoneID string, load interface{}, ttl time.Duration) error {
	data, err := json.Marshal(load)
	if err != nil {
		return fmt.Errorf("failed to marshal load: %w", err)
	}

	key := r.key("zone", zoneID, "load")
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedZoneLoad retrieves the cached load for a zone
func (r *RedisRepository) GetCachedZoneLoad(ctx context.Context, zoneID string) ([]byte, error) {
	key := r.key("zone", zoneID, "load")
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// GetAllZoneLoads retrieves cached loads for all zones in a facility
func (r *RedisRepository) GetAllZoneLoads(ctx context.Context, facilityID string) (map[string]interface{}, error) {
	pattern := r.key("zone", "*", "load")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	loads := make(map[string]interface{})
	for _, key := range keys {
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var load interface{}
		if err := json.Unmarshal(data, &load); err != nil {
			continue
		}

		// Extract zone ID from key
		// Key format: prefix:zone:zoneID:load
		loads[key] = load
	}

	return loads, nil
}

// --- Real-time Telemetry ---

// PublishEnergyReading publishes an energy reading to the telemetry stream
func (r *RedisRepository) PublishEnergyReading(ctx context.Context, channel string, reading interface{}) error {
	data, err := json.Marshal(reading)
	if err != nil {
		return fmt.Errorf("failed to marshal reading: %w", err)
	}

	return r.client.Publish(ctx, channel, data).Err()
}

// SubscribeEnergyReading subscribes to energy reading events
func (r *RedisRepository) SubscribeEnergyReading(ctx context.Context, channel string) *redis.PubSub {
	pubsub := r.client.Subscribe(ctx, channel)
	return pubsub
}

// --- Rate Limiting ---

// CheckRateLimit checks if a meter has exceeded its rate limit
func (r *RedisRepository) CheckRateLimit(ctx context.Context, meterID string, limit int, window time.Duration) (bool, error) {
	key := r.key("rate", meterID)
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		r.client.Expire(ctx, key, window)
	}

	return count > int64(limit), nil
}

// --- Aggregation Keys ---

// IncrementDailyConsumption atomically increments daily consumption counter
func (r *RedisRepository) IncrementDailyConsumption(ctx context.Context, facilityID, energySource string, amount decimal.Decimal) error {
	date := time.Now().Format("2006-01-02")
	key := r.key("daily_consumption", facilityID, energySource, date)

	return r.client.HIncrByFloat(ctx, key, "total", amount.InexactFloat64()).Err()
}

// GetDailyConsumption retrieves the daily consumption for a facility
func (r *RedisRepository) GetDailyConsumption(ctx context.Context, facilityID, energySource string) (decimal.Decimal, error) {
	date := time.Now().Format("2006-01-02")
	key := r.key("daily_consumption", facilityID, energySource, date)

	value, err := r.client.HGet(ctx, key, "total").Float64()
	if err == redis.Nil {
		return decimal.Zero, nil
	}
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromFloat(value), nil
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

// --- Health Check ---

// Ping checks if Redis is available
func (r *RedisRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
