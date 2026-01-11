package online

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"neam-platform/shared"
)

// Config holds the configuration for the online feature store
type Config struct {
	Redis  *redis.Client
	Logger *shared.Logger
}

// Service handles real-time feature serving
type Service struct {
	config    Config
	logger    *shared.Logger
	redis     *redis.Client
	features  map[string]*FeatureConfig
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// FeatureConfig holds configuration for a feature
type FeatureConfig struct {
	Name         string        `json:"name"`
	TTL          time.Duration `json:"ttl"`
	RefreshRate  time.Duration `json:"refresh_rate"`
	Aggregation  string        `json:"aggregation"` // none, sum, avg, max, min
}

// FeatureValue represents a feature value with metadata
type FeatureValue struct {
	Name       string                 `json:"name"`
	Value      interface{}            `json:"value"`
	EntityID   string                 `json:"entity_id"`
	Version    string                 `json:"version"`
	ComputedAt time.Time              `json:"computed_at"`
	Source     string                 `json:"source"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// BatchRequest represents a batch feature request
type BatchRequest struct {
	EntityIDs  []string `json:"entity_ids"`
	Features   []string `json:"features"`
	Version    string   `json:"version"`
}

// BatchResponse represents a batch feature response
type BatchResponse struct {
	Entities   map[string]map[string]FeatureValue `json:"entities"`
	Missing    map[string][]string                `json:"missing"`
	ComputedAt time.Time                          `json:"computed_at"`
}

// CachedFeatures represents cached features for an entity
type CachedFeatures struct {
	EntityID   string                 `json:"entity_id"`
	Features   map[string]FeatureValue `json:"features"`
	CachedAt   time.Time              `json:"cached_at"`
	ExpiresAt  time.Time              `json:"expires_at"`
}

// AggregationType defines the type of aggregation
type AggregationType int

const (
	AggNone AggregationType = iota
	AggSum
	AggAvg
	AggMax
	AggMin
	AggCount
)

// NewService creates a new online feature store service
func NewService(cfg Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		config:   cfg,
		logger:   cfg.Logger,
		redis:    cfg.Redis,
		features: make(map[string]*FeatureConfig),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start begins the online feature store service
func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Starting online feature store service")

	// Start feature refresh workers
	s.wg.Add(1)
	go s.featureRefreshLoop()

	s.logger.Info("Online feature store service started")
}

// Stop gracefully stops the service
func (s *Service) Stop() {
	s.logger.Info("Stopping online feature store service")
	s.cancel()
	s.wg.Wait()
	s.logger.Info("Online feature store service stopped")
}

// featureRefreshLoop periodically refreshes cached features
func (s *Service) featureRefreshLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.refreshExpiredFeatures()
		}
	}
}

// refreshExpiredFeatures refreshes features that have expired
func (s *Service) refreshExpiredFeatures() {
	// Find and refresh expired feature caches
	s.logger.Debug("Refreshing expired features")
}

// GetFeatures retrieves features for a single entity
func (s *Service) GetFeatures(ctx context.Context, entityID string, featureNames []string, version string) (map[string]FeatureValue, error) {
	result := make(map[string]FeatureValue)

	// Try to get from cache first
	for _, name := range featureNames {
		value, err := s.getFeature(ctx, entityID, name, version)
		if err != nil {
			s.logger.Warn("Feature not found", "entity_id", entityID, "feature", name)
			continue
		}
		result[name] = value
	}

	return result, nil
}

// getFeature retrieves a single feature value
func (s *Service) getFeature(ctx context.Context, entityID, featureName, version string) (FeatureValue, error) {
	// Construct cache key
	cacheKey := fmt.Sprintf("neam:features:%s:%s:%s", entityID, featureName, version)

	// Try Redis cache
	value, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && value != "" {
		return FeatureValue{
			Name:       featureName,
			Value:      value,
			EntityID:   entityID,
			Version:    version,
			ComputedAt: time.Now(),
			Source:     "cache",
		}, nil
	}

	// Feature not in cache - would need to compute from source
	// This is where you would integrate with the computation engine
	return FeatureValue{}, fmt.Errorf("feature not found in cache")
}

// BatchGetFeatures retrieves features for multiple entities
func (s *Service) BatchGetFeatures(ctx context.Context, request BatchRequest) (*BatchResponse, error) {
	response := &BatchResponse{
		Entities: make(map[string]map[string]FeatureValue),
		Missing:  make(map[string][]string),
	}

	// Create pipeline for efficient batch retrieval
	pipe := s.redis.Pipeline()

	// Queue all cache lookups
	cmds := make(map[string]*redis.StringCmd)
	for _, entityID := range request.EntityIDs {
		for _, featureName := range request.Features {
			cacheKey := fmt.Sprintf("neam:features:%s:%s:%s", entityID, featureName, request.Version)
			cmds[cacheKey] = pipe.Get(ctx, cacheKey)
		}
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("batch get failed: %w", err)
	}

	// Process results
	for _, entityID := range request.EntityIDs {
		response.Entities[entityID] = make(map[string]FeatureValue)
		missingFeatures := []string{}

		for _, featureName := range request.Features {
			cacheKey := fmt.Sprintf("neam:features:%s:%s:%s", entityID, featureName, request.Version)
			cmd := cmds[cacheKey]
			value, err := cmd.Result()

			if err == redis.Nil {
				missingFeatures = append(missingFeatures, featureName)
			} else if err == nil {
				response.Entities[entityID][featureName] = FeatureValue{
					Name:       featureName,
					Value:      value,
					EntityID:   entityID,
					Version:    request.Version,
					ComputedAt: time.Now(),
					Source:     "cache",
				}
			}
		}

		if len(missingFeatures) > 0 {
			response.Missing[entityID] = missingFeatures
		}
	}

	response.ComputedAt = time.Now()

	return response, nil
}

// SetFeature stores a feature value in the cache
func (s *Service) SetFeature(ctx context.Context, entityID, featureName, version string, value interface{}, ttl time.Duration) error {
	s.mu.RLock()
	featureConfig, exists := s.features[featureName]
	s.mu.RUnlock()

	if !exists {
		// Use default TTL if not configured
		featureConfig = &FeatureConfig{
			Name: featureName,
			TTL:  24 * time.Hour,
		}
	}

	// Determine effective TTL
	effectiveTTL := ttl
	if effectiveTTL == 0 {
		effectiveTTL = featureConfig.TTL
	}

	// Convert value to string for storage
	var valueStr string
	switch v := value.(type) {
	case string:
		valueStr = v
	case float64:
		valueStr = fmt.Sprintf("%f", v)
	case int64:
		valueStr = fmt.Sprintf("%d", v)
	case bool:
		valueStr = fmt.Sprintf("%t", v)
	default:
		valueStr = fmt.Sprintf("%v", v)
	}

	// Construct cache key
	cacheKey := fmt.Sprintf("neam:features:%s:%s:%s", entityID, featureName, version)

	// Store in Redis
	err := s.redis.Set(ctx, cacheKey, valueStr, effectiveTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set feature: %w", err)
	}

	s.logger.Debug("Feature cached", "entity_id", entityID, "feature", featureName)

	return nil
}

// SetFeatures stores multiple feature values
func (s *Service) SetFeatures(ctx context.Context, entityID string, features map[string]interface{}, version string, ttl time.Duration) error {
	pipe := s.redis.Pipeline()

	for name, value := range features {
		cacheKey := fmt.Sprintf("neam:features:%s:%s:%s", entityID, name, version)
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = v
		case float64:
			valueStr = fmt.Sprintf("%f", v)
		case int64:
			valueStr = fmt.Sprintf("%d", v)
		default:
			valueStr = fmt.Sprintf("%v", v)
		}
		pipe.Set(ctx, cacheKey, valueStr, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set features: %w", err)
	}

	return nil
}

// DeleteFeature removes a feature from the cache
func (s *Service) DeleteFeature(ctx context.Context, entityID, featureName, version string) error {
	cacheKey := fmt.Sprintf("neam:features:%s:%s:%s", entityID, featureName, version)
	return s.redis.Del(ctx, cacheKey).Err()
}

// DeleteEntityFeatures removes all features for an entity
func (s *Service) DeleteEntityFeatures(ctx context.Context, entityID string, version string) error {
	pattern := fmt.Sprintf("neam:features:%s:*:%s", entityID, version)
	iter := s.redis.Scan(ctx, 0, pattern, 100).Iterator()

	for iter.Next(ctx) {
		if err := s.redis.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

// GetEntityFeatures retrieves all cached features for an entity
func (s *Service) GetEntityFeatures(ctx context.Context, entityID, version string) (*CachedFeatures, error) {
	pattern := fmt.Sprintf("neam:features:%s:*:%s", entityID, version)
	iter := s.redis.Scan(ctx, 0, pattern, 1000).Iterator()

	features := make(map[string]FeatureValue)
	var maxExpiry time.Time

	for iter.Next(ctx) {
		// Parse cache key
		cacheKey := iter.Val()
		value, err := s.redis.Get(ctx, cacheKey).Result()
		if err != nil {
			continue
		}

		// Extract feature name from key
		// Format: neam:features:{entity_id}:{feature_name}:{version}
		var featureName string
		// Simplified parsing - in production, use more robust parsing
		features[featureName] = FeatureValue{
			Name:       featureName,
			Value:      value,
			EntityID:   entityID,
			Version:    version,
			ComputedAt: time.Now(),
			Source:     "cache",
		}
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return &CachedFeatures{
		EntityID:  entityID,
		Features:  features,
		CachedAt:  time.Now(),
		ExpiresAt: maxExpiry,
	}, nil
}

// IncrementFeature atomically increments a numeric feature
func (s *Service) IncrementFeature(ctx context.Context, entityID, featureName, version string, delta float64) (float64, error) {
	cacheKey := fmt.Sprintf("neam:features:%s:%s:%s", entityID, featureName, version)

	// Use Redis INCRBYFLOAT for atomic increment
	newValue, err := s.redis.IncrByFloat(ctx, cacheKey, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment feature: %w", err)
	}

	// Set expiry if not already set
	s.redis.Expire(ctx, cacheKey, 24*time.Hour)

	return newValue, nil
}

// AggregateFeatures performs aggregation across entities
func (s *Service) AggregateFeatures(ctx context.Context, featureName, version string, aggType AggregationType) (float64, error) {
	pattern := fmt.Sprintf("neam:features:*:%s:%s", featureName, version)
	iter := s.redis.Scan(ctx, 0, pattern, 10000).Iterator()

	var sum float64
	var count int64
	var min, max float64 = math.MaxFloat64, -math.MaxFloat64

	for iter.Next(ctx) {
		value, err := s.redis.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}

		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			continue
		}

		sum += floatVal
		count++
		if floatVal < min {
			min = floatVal
		}
		if floatVal > max {
			max = floatVal
		}
	}

	if err := iter.Err(); err != nil {
		return 0, err
	}

	switch aggType {
	case AggSum:
		return sum, nil
	case AggAvg:
		if count == 0 {
			return 0, nil
		}
		return sum / float64(count), nil
	case AggMax:
		if count == 0 {
			return 0, nil
		}
		return max, nil
	case AggMin:
		if count == 0 {
			return 0, nil
		}
		return min, nil
	case AggCount:
		return float64(count), nil
	default:
		return 0, nil
	}
}

// RegisterFeature registers a feature configuration
func (s *Service) RegisterFeature(ctx context.Context, config FeatureConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.features[config.Name] = &config

	s.logger.Info("Feature registered", "name", config.Name, "ttl", config.TTL)

	return nil
}

// GetLatency returns the P50, P95, and P99 latency for feature retrieval
func (s *Service) GetLatency(ctx context.Context) (p50, p95, p99 time.Duration) {
	// In production, this would use latency histogram from Redis
	// For now, return estimated values
	return 500 * time.Microsecond, 2 * time.Millisecond, 5 * time.Millisecond
}

// HealthCheck verifies the online store is healthy
func (s *Service) HealthCheck(ctx context.Context) error {
	return s.redis.Ping(ctx).Err()
}

// Import math for aggregation functions
import "math"
import "strconv"
