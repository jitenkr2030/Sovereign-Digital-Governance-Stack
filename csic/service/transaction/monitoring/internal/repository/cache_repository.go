package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/transaction-monitoring/internal/config"
	"github.com/csic/transaction-monitoring/internal/domain/models"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CacheRepository handles Redis caching operations
type CacheRepository struct {
	client    *redis.Client
	cfg       *config.Config
	logger    *zap.Logger
	keyPrefix string
}

// NewCacheRepository creates a new cache repository instance
func NewCacheRepository(cfg *config.Config, logger *zap.Logger) (*CacheRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &CacheRepository{
		client:    client,
		cfg:       cfg,
		logger:    logger,
		keyPrefix: cfg.Redis.KeyPrefix,
	}, nil
}

// Close closes the Redis connection
func (r *CacheRepository) Close() error {
	return r.client.Close()
}

// key generates a prefixed cache key
func (r *CacheRepository) key(parts ...string) string {
	key := r.keyPrefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

// Wallet risk score operations

// SetWalletRiskScore caches a wallet risk score
func (r *CacheRepository) SetWalletRiskScore(ctx context.Context, address string, network models.Network, score float64) error {
	key := r.key("risk", string(network), address)
	return r.client.Set(ctx, key, score, 24*time.Hour).Err()
}

// GetWalletRiskScore retrieves a cached wallet risk score
func (r *CacheRepository) GetWalletRiskScore(ctx context.Context, address string, network models.Network) (float64, error) {
	key := r.key("risk", string(network), address)
	val, err := r.client.Get(ctx, key).Float64()
	if err == redis.Nil {
		return 0, fmt.Errorf("cache miss")
	}
	return val, err
}

// SetWalletRiskFactors caches wallet risk factors
func (r *CacheRepository) SetWalletRiskFactors(ctx context.Context, factors *models.RiskFactors) error {
	key := r.key("risk:factors", factors.WalletID)
	data, err := json.Marshal(factors)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, 24*time.Hour).Err()
}

// GetWalletRiskFactors retrieves cached risk factors
func (r *CacheRepository) GetWalletRiskFactors(ctx context.Context, walletID string) (*models.RiskFactors, error) {
	key := r.key("risk:factors", walletID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("cache miss")
	}
	if err != nil {
		return nil, err
	}

	var factors models.RiskFactors
	if err := json.Unmarshal(data, &factors); err != nil {
		return nil, err
	}
	return &factors, nil
}

// Sanctions operations

// SetSanctionsStatus caches sanctions status
func (r *CacheRepository) SetSanctionsStatus(ctx context.Context, address string, isSanctioned bool) error {
	key := r.key("sanctions", address)
	if isSanctioned {
		return r.client.Set(ctx, key, "1", 24*time.Hour).Err()
	}
	return r.client.Set(ctx, key, "0", 24*time.Hour).Err()
}

// GetSanctionsStatus retrieves cached sanctions status
func (r *CacheRepository) GetSanctionsStatus(ctx context.Context, address string) (bool, error) {
	key := r.key("sanctions", address)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

// AddToSanctionsBloom adds an address to the sanctions bloom filter representation
func (r *CacheRepository) AddToSanctionsBloom(address string) error {
	// This would add to a Redis Bloom filter module if available
	// For now, we use a simple set
	key := r.key("sanctions:bloom")
	return r.client.SAdd(context.Background(), key, address).Err()
}

// TestSanctionsBloom tests if an address is in the sanctions bloom filter
func (r *CacheRepository) TestSanctionsBloom(address string) (bool, error) {
	key := r.key("sanctions:bloom")
	return r.client.SIsMember(context.Background(), key, address).Result()
}

// Velocity tracking operations

// RecordTransaction records a transaction for velocity tracking
func (r *CacheRepository) RecordTransaction(ctx context.Context, address string, network models.Network, amount float64, timestamp time.Time) error {
	key := r.key("velocity", string(network), address)
	// Add to sorted set with timestamp as score
	member := redis.Z{
		Score:  float64(timestamp.Unix()),
		Member: fmt.Sprintf("%f:%d", amount, timestamp.UnixNano()),
	}
	return r.client.ZAdd(ctx, key, member).Err()
}

// GetVelocity24h calculates 24-hour transaction velocity
func (r *CacheRepository) GetVelocity24h(ctx context.Context, address string, network models.Network) (float64, error) {
	key := r.key("velocity", string(network), address)
	now := time.Now()
	oneDayAgo := now.Add(-24 * time.Hour)

	// Get transactions in the last 24 hours
	results, err := r.client.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", float64(oneDayAgo.Unix())),
		Max: fmt.Sprintf("%f", float64(now.Unix())),
	}).Result()
	if err != nil {
		return 0, err
	}

	var total float64
	for _, z := range results {
		var amount float64
		fmt.Sscanf(z.Member.(string), "%f:", &amount)
		total += amount
	}

	// Clean up old entries
	r.client.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%f", float64(oneDayAgo.Unix()-7*24*3600)))

	return total, nil
}

// GetTransactionCount24h gets transaction count in last 24 hours
func (r *CacheRepository) GetTransactionCount24h(ctx context.Context, address string, network models.Network) (int64, error) {
	key := r.key("velocity", string(network), address)
	now := time.Now()
	oneDayAgo := now.Add(-24 * time.Hour)

	return r.client.ZCount(ctx, key,
		fmt.Sprintf("%f", float64(oneDayAgo.Unix())),
		fmt.Sprintf("%f", float64(now.Unix())),
	).Result()
}

// Rate limiting operations

// CheckRateLimit checks if an address has exceeded rate limits
func (r *CacheRepository) CheckRateLimit(ctx context.Context, identifier string, limit int, window time.Duration) (bool, error) {
	key := r.key("rate", identifier, fmt.Sprintf("%d", window))
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		r.client.Expire(ctx, key, window)
	}

	return count > int64(limit), nil
}

// Alert caching operations

// CacheAlert caches an alert for deduplication
func (r *CacheRepository) CacheAlert(ctx context.Context, alertID string, ttl time.Duration) error {
	key := r.key("alert:seen", alertID)
	return r.client.Set(ctx, key, "1", ttl).Err()
}

// IsAlertCached checks if an alert has been processed
func (r *CacheRepository) IsAlertCached(ctx context.Context, alertID string) (bool, error) {
	key := r.key("alert:seen", alertID)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// Cluster cache operations

// SetClusterCache caches cluster data
func (r *CacheRepository) SetClusterCache(ctx context.Context, clusterID string, data interface{}) error {
	key := r.key("cluster", clusterID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, jsonData, 1*time.Hour).Err()
}

// GetClusterCache retrieves cached cluster data
func (r *CacheRepository) GetClusterCache(ctx context.Context, clusterID string) ([]byte, error) {
	key := r.key("cluster", clusterID)
	return r.client.Get(ctx, key).Bytes()
}

// Invalidation operations

// InvalidateWalletCache invalidates all cached data for a wallet
func (r *CacheRepository) InvalidateWalletCache(ctx context.Context, address string, network models.Network) error {
	pattern := r.key("risk", string(network), address) + "*"
	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// InvalidateClusterCache invalidates cached cluster data
func (r *CacheRepository) InvalidateClusterCache(ctx context.Context, clusterID string) error {
	key := r.key("cluster", clusterID)
	return r.client.Del(ctx, key).Err()
}

// Health check operations

// Ping checks Redis connectivity
func (r *CacheRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// GetStats returns Redis stats
func (r *CacheRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	pool := r.client.PoolStats()
	return map[string]interface{}{
		"pool_size":      pool.TotalConns,
		"idle_conns":     pool.IdleConns,
		"hits":           pool.Hits,
		"misses":         pool.Misses,
		"Timeouts":       pool.Timeouts,
	}, nil
}

// Bulk operations

// BulkSetWalletScores bulk sets wallet risk scores
func (r *CacheRepository) BulkSetWalletScores(ctx context.Context, scores map[string]float64, ttl time.Duration) error {
	if len(scores) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	for key, score := range scores {
		pipe.Set(ctx, key, score, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// BulkGetWalletScores bulk retrieves wallet risk scores
func (r *CacheRepository) BulkGetWalletScores(ctx context.Context, keys []string) (map[string]float64, error) {
	if len(keys) == 0 {
		return make(map[string]float64), nil
	}

	cmds, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	scores := make(map[string]float64)
	for i, cmd := range cmds {
		if cmd != nil {
			if val, ok := cmd.(string); ok {
				var score float64
				fmt.Sscanf(val, "%f", &score)
				scores[keys[i]] = score
			}
		}
	}
	return scores, nil
}
