package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"csic-platform/service/reporting/internal/domain"
)

// CacheRepository provides caching functionality for the reporting service
type CacheRepository struct {
	client    *redis.Client
	defaultTTL time.Duration
}

// NewCacheRepository creates a new cache repository
func NewCacheRepository(client *redis.Client, defaultTTLSeconds int) *CacheRepository {
	return &CacheRepository{
		client:    client,
		defaultTTL: time.Duration(defaultTTLSeconds) * time.Second,
	}
}

// TemplateCacheKey returns the cache key for a report template
func TemplateCacheKey(id string) string {
	return fmt.Sprintf("report:template:%s", id)
}

// ReportCacheKey returns the cache key for a generated report
func ReportCacheKey(id string) string {
	return fmt.Sprintf("report:generated:%s", id)
}

// RegulatorCacheKey returns the cache key for a regulator
func RegulatorCacheKey(id string) string {
	return fmt.Sprintf("report:regulator:%s", id)
}

// SetTemplate caches a report template
func (r *CacheRepository) SetTemplate(ctx context.Context, template *domain.ReportTemplate) error {
	data, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	key := TemplateCacheKey(template.ID.String())
	return r.client.Set(ctx, key, data, r.defaultTTL).Err()
}

// GetTemplate retrieves a cached report template
func (r *CacheRepository) GetTemplate(ctx context.Context, id string) (*domain.ReportTemplate, error) {
	key := TemplateCacheKey(id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get template from cache: %w", err)
	}

	var template domain.ReportTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	return &template, nil
}

// DeleteTemplate removes a template from cache
func (r *CacheRepository) DeleteTemplate(ctx context.Context, id string) error {
	key := TemplateCacheKey(id)
	return r.client.Del(ctx, key).Err()
}

// InvalidateTemplateCache invalidates all template cache entries
func (r *CacheRepository) InvalidateTemplateCache(ctx context.Context) error {
	pattern := "report:template:*"
	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", iter.Val(), err)
		}
	}
	return iter.Err()
}

// SetReport caches a generated report
func (r *CacheRepository) SetReport(ctx context.Context, report *domain.GeneratedReport) error {
	data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	key := ReportCacheKey(report.ID.String())
	// Use shorter TTL for generated reports
	ttl := 5 * time.Minute
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetReport retrieves a cached generated report
func (r *CacheRepository) GetReport(ctx context.Context, id string) (*domain.GeneratedReport, error) {
	key := ReportCacheKey(id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get report from cache: %w", err)
	}

	var report domain.GeneratedReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report: %w", err)
	}

	return &report, nil
}

// DeleteReport removes a report from cache
func (r *CacheRepository) DeleteReport(ctx context.Context, id string) error {
	key := ReportCacheKey(id)
	return r.client.Del(ctx, key).Err()
}

// SetRegulator caches a regulator
func (r *CacheRepository) SetRegulator(ctx context.Context, regulator *domain.Regulator) error {
	data, err := json.Marshal(regulator)
	if err != nil {
		return fmt.Errorf("failed to marshal regulator: %w", err)
	}

	key := RegulatorCacheKey(regulator.ID.String())
	return r.client.Set(ctx, key, data, r.defaultTTL).Err()
}

// GetRegulator retrieves a cached regulator
func (r *CacheRepository) GetRegulator(ctx context.Context, id string) (*domain.Regulator, error) {
	key := RegulatorCacheKey(id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get regulator from cache: %w", err)
	}

	var regulator domain.Regulator
	if err := json.Unmarshal(data, &regulator); err != nil {
		return nil, fmt.Errorf("failed to unmarshal regulator: %w", err)
	}

	return &regulator, nil
}

// DeleteRegulator removes a regulator from cache
func (r *CacheRepository) DeleteRegulator(ctx context.Context, id string) error {
	key := RegulatorCacheKey(id)
	return r.client.Del(ctx, key).Err()
}

// SetReportList caches a list of reports for a regulator
func (r *CacheRepository) SetReportList(ctx context.Context, regulatorID string, reports []*domain.GeneratedReport) error {
	data, err := json.Marshal(reports)
	if err != nil {
		return fmt.Errorf("failed to marshal report list: %w", err)
	}

	key := fmt.Sprintf("report:list:%s", regulatorID)
	ttl := 2 * time.Minute
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetReportList retrieves a cached list of reports for a regulator
func (r *CacheRepository) GetReportList(ctx context.Context, regulatorID string) ([]*domain.GeneratedReport, error) {
	key := fmt.Sprintf("report:list:%s", regulatorID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get report list from cache: %w", err)
	}

	var reports []*domain.GeneratedReport
	if err := json.Unmarshal(data, &reports); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report list: %w", err)
	}

	return reports, nil
}

// DeleteReportList removes a report list from cache
func (r *CacheRepository) DeleteReportList(ctx context.Context, regulatorID string) error {
	key := fmt.Sprintf("report:list:%s", regulatorID)
	return r.client.Del(ctx, key).Err()
}

// IncrementCounter increments a counter in Redis
func (r *CacheRepository) IncrementCounter(ctx context.Context, key string) error {
	return r.client.Incr(ctx, key).Err()
}

// GetCounter gets a counter value from Redis
func (r *CacheRepository) GetCounter(ctx context.Context, key string) (int64, error) {
	return r.client.Get(ctx, key).Int64()
}

// SetWithTTL sets a key with a specific TTL
func (r *CacheRepository) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a value from cache
func (r *CacheRepository) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set sets a string value in cache
func (r *CacheRepository) Set(ctx context.Context, key, value string) error {
	return r.client.Set(ctx, key, value, r.defaultTTL).Err()
}

// Delete deletes a key from cache
func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// AddToSet adds a member to a set
func (r *CacheRepository) AddToSet(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members).Err()
}

// GetSetMembers gets all members of a set
func (r *CacheRepository) GetSetMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SetExpiry sets an expiry on a key
func (r *CacheRepository) SetExpiry(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// HealthCheck checks if Redis is healthy
func (r *CacheRepository) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (r *CacheRepository) Close() error {
	return r.client.Close()
}
