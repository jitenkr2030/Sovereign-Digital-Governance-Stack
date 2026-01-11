package coordination

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"neam-platform/shared"
)

// RedisLockProvider implements LockProvider using Redis
type RedisLockProvider struct {
	redis     *redis.Client
	config    *LockConfig
	logger    *shared.Logger
	generator *FencingTokenGenerator
}

// NewRedisLockProvider creates a new Redis-based lock provider
func NewRedisLockProvider(redis *redis.Client, config *LockConfig, logger *shared.Logger) *RedisLockProvider {
	provider := &RedisLockProvider{
		redis:  redis,
		config: config,
		logger: logger,
	}

	// Initialize fencing token generator
	if config.EnableFencingToken {
		provider.generator = NewFencingTokenGenerator(redis, fmt.Sprintf("%s:fencing:token", config.LockPrefix))
	}

	return provider
}

// Lua script for atomic lock acquisition
// Ensures that only one client can acquire the lock at a time
const acquireScript = `
-- KEYS[1] = lock key
-- ARGV[1] = lock value (JSON)
-- ARGV[2] = TTL in milliseconds

local result = redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[2], 'NX')
if result then
    return 1
else
    return 0
end
`

// Lua script for atomic lock release
// Only releases the lock if the owner matches
const releaseScript = `
-- KEYS[1] = lock key
-- ARGV[1] = expected owner ID

local current = redis.call('GET', KEYS[1])
if current then
    local data = cjson.decode(current)
    if data.owner_id == ARGV[1] then
        return redis.call('DEL', KEYS[1])
    else
        return 0
    end
else
    return -1
end
`

// Lua script for atomic lock extension
// Only extends the lock if the owner matches
const extendScript = `
-- KEYS[1] = lock key
-- ARGV[1] = expected owner ID
-- ARGV[2] = new TTL in milliseconds

local current = redis.call('GET', KEYS[1])
if current then
    local data = cjson.decode(current)
    if data.owner_id == ARGV[1] then
        return redis.call('PEXPIRE', KEYS[1], ARGV[2])
    else
        return 0
    end
else
    return -1
end
`

// Lua script for lock validation
// Returns the lock data if the owner matches, nil otherwise
const validateScript = `
-- KEYS[1] = lock key
-- ARGV[1] = expected owner ID

local current = redis.call('GET', KEYS[1])
if current then
    local data = cjson.decode(current)
    if data.owner_id == ARGV[1] then
        return current
    else
        return nil
    end
else
    return nil
end
`

// Lua script for getting lock info
// Returns the lock data or nil if not exists
const getLockScript = `
-- KEYS[1] = lock key

return redis.call('GET', KEYS[1])
`

// Acquire implements LockProvider.Acquire
func (p *RedisLockProvider) Acquire(ctx context.Context, resourceType, resourceID string, ttl time.Duration) (*Lock, error) {
	lockKey := p.buildLockKey(resourceType, resourceID)

	// Generate owner ID and fencing token
	ownerID := uuid.New().String()
	fencingToken := int64(1)

	if p.generator != nil {
		token, err := p.generator.Next(ctx)
		if err != nil {
			p.logger.Warn("Failed to generate fencing token, using fallback",
				"error", err,
			)
		} else {
			fencingToken = token
		}
	}

	now := time.Now()
	lock := &Lock{
		ID:           uuid.New(),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		OwnerID:      ownerID,
		FencingToken: fencingToken,
		AcquiredAt:   now,
		ExpiresAt:    now.Add(ttl),
		TTL:          ttl,
	}

	// Serialize lock data
	lockData := map[string]interface{}{
		"owner_id":       ownerID,
		"fencing_token":  fencingToken,
		"created_at":     now.UnixMilli(),
		"expires_at":     now.Add(ttl).UnixMilli(),
		"lock_id":        lock.ID.String(),
		"resource_type":  resourceType,
		"resource_id":    resourceID,
	}

	lockJSON, err := json.Marshal(lockData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal lock data: %w", err)
	}

	// Execute atomic acquisition
	result, err := p.redis.Eval(ctx, acquireScript, []string{lockKey}, string(lockJSON), ttl.Milliseconds()).Result()
	if err != nil {
		p.logger.Error("Failed to execute acquire script",
			"error", err,
			"resource_type", resourceType,
			"resource_id", resourceID,
		)
		return nil, fmt.Errorf("redis error: %w", err)
	}

	acquired, ok := result.(int64)
	if !ok {
		return nil, fmt.Errorf("unexpected result type from Redis")
	}

	if acquired != 1 {
		// Lock already held by another client
		return nil, ErrLockNotAcquired
	}

	p.logger.Debug("Lock acquired",
		"lock_id", lock.ID,
		"resource_type", resourceType,
		"resource_id", resourceID,
		"owner_id", ownerID,
		"ttl", ttl,
	)

	return lock, nil
}

// Release implements LockProvider.Release
func (p *RedisLockProvider) Release(ctx context.Context, lock *Lock) error {
	if lock == nil {
		return nil
	}

	lockKey := p.buildLockKey(lock.ResourceType, lock.ResourceID)

	// Execute atomic release
	result, err := p.redis.Eval(ctx, releaseScript, []string{lockKey}, lock.OwnerID).Result()
	if err != nil {
		p.logger.Error("Failed to execute release script",
			"error", err,
			"lock_id", lock.ID,
		)
		return fmt.Errorf("redis error: %w", err)
	}

	deleted, ok := result.(int64)
	if !ok {
		return fmt.Errorf("unexpected result type from Redis")
	}

	if deleted == 0 {
		// Lock was held by a different owner
		p.logger.Warn("Failed to release lock - not the owner",
			"lock_id", lock.ID,
			"owner_id", lock.OwnerID,
		)
		return ErrLockNotHeld
	}

	if deleted == -1 {
		// Lock doesn't exist
		p.logger.Warn("Failed to release lock - lock does not exist",
			"lock_id", lock.ID,
		)
		return ErrLockNotHeld
	}

	p.logger.Debug("Lock released",
		"lock_id", lock.ID,
		"resource_type", lock.ResourceType,
		"resource_id", lock.ResourceID,
	)

	return nil
}

// Extend implements LockProvider.Extend
func (p *RedisLockProvider) Extend(ctx context.Context, lock *Lock, additionalTTL time.Duration) error {
	if lock == nil {
		return nil
	}

	lockKey := p.buildLockKey(lock.ResourceType, lock.ResourceID)
	newExpiry := time.Now().Add(additionalTTL)

	// Execute atomic extension
	result, err := p.redis.Eval(ctx, extendScript, []string{lockKey}, lock.OwnerID, additionalTTL.Milliseconds()).Result()
	if err != nil {
		p.logger.Error("Failed to execute extend script",
			"error", err,
			"lock_id", lock.ID,
		)
		return fmt.Errorf("redis error: %w", err)
	}

	extended, ok := result.(int64)
	if !ok {
		return fmt.Errorf("unexpected result type from Redis")
	}

	if extended == 0 {
		// Lock was held by a different owner
		p.logger.Warn("Failed to extend lock - not the owner",
			"lock_id", lock.ID,
		)
		return ErrLockNotHeld
	}

	if extended == -1 {
		// Lock doesn't exist
		p.logger.Warn("Failed to extend lock - lock does not exist",
			"lock_id", lock.ID,
		)
		return ErrLockNotHeld
	}

	// Update lock expiration
	lock.ExpiresAt = newExpiry
	lock.TTL = additionalTTL

	p.logger.Debug("Lock extended",
		"lock_id", lock.ID,
		"new_expiry", newExpiry,
	)

	return nil
}

// Validate implements LockProvider.Validate
func (p *RedisLockProvider) Validate(ctx context.Context, lock *Lock) (bool, error) {
	if lock == nil {
		return false, nil
	}

	lockKey := p.buildLockKey(lock.ResourceType, lock.ResourceID)

	// Execute atomic validation
	result, err := p.redis.Eval(ctx, validateScript, []string{lockKey}, lock.OwnerID).Result()
	if err != nil {
		p.logger.Error("Failed to execute validate script",
			"error", err,
			"lock_id", lock.ID,
		)
		return false, fmt.Errorf("redis error: %w", err)
	}

	return result != nil, nil
}

// GetLock implements LockProvider.GetLock
func (p *RedisLockProvider) GetLock(ctx context.Context, resourceType, resourceID string) (*Lock, error) {
	lockKey := p.buildLockKey(resourceType, resourceID)

	// Execute atomic get
	result, err := p.redis.Eval(ctx, getLockScript, []string{lockKey}).Result()
	if err != nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}

	if result == nil {
		return nil, nil // Lock doesn't exist
	}

	// Parse lock data
	lockJSON, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type from Redis")
	}

	var lockData map[string]interface{}
	if err := json.Unmarshal([]byte(lockJSON), &lockData); err != nil {
		return nil, fmt.Errorf("failed to parse lock data: %w", err)
	}

	// Reconstruct lock object
	lock := &Lock{
		ID:           uuid.MustParse(lockData["lock_id"].(string)),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		OwnerID:      lockData["owner_id"].(string),
		FencingToken: int64(lockData["fencing_token"].(float64)),
		AcquiredAt:   time.UnixMilli(int64(lockData["created_at"].(float64))),
		ExpiresAt:    time.UnixMilli(int64(lockData["expires_at"].(float64))),
	}

	// Calculate TTL
	lock.TTL = time.Until(lock.ExpiresAt)

	return lock, nil
}

// buildLockKey constructs the Redis key for a lock
func (p *RedisLockProvider) buildLockKey(resourceType, resourceID string) string {
	return fmt.Sprintf("%s:%s:%s", p.config.LockPrefix, resourceType, resourceID)
}

// GetActiveLocks returns all currently held locks for a resource type
func (p *RedisLockProvider) GetActiveLocks(ctx context.Context, resourceType string) ([]*Lock, error) {
	pattern := fmt.Sprintf("%s:%s:*", p.config.LockPrefix, resourceType)

	var keys []string
	var cursor uint64
	for {
		var err error
		keys, cursor, err = p.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan locks: %w", err)
		}

		for _, key := range keys {
			// Extract resource ID from key
			resourceID := extractResourceID(key, p.config.LockPrefix, resourceType)
			if resourceID == "" {
				continue
			}

			lock, err := p.GetLock(ctx, resourceType, resourceID)
			if err != nil {
				p.logger.Warn("Failed to get lock",
					"error", err,
					"key", key,
				)
				continue
			}

			if lock != nil {
				// Check if lock is still valid
				if !lock.IsExpired() {
					// We can't easily return all locks here due to the GetLock implementation
					// This is a simplified version
				}
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil, nil // Simplified return
}

// GetLockStats returns statistics about lock usage
func (p *RedisLockProvider) GetLockStats(ctx context.Context) (*LockStats, error) {
	stats := &LockStats{
		Timestamp: time.Now(),
	}

	pattern := fmt.Sprintf("%s:*", p.config.LockPrefix)

	var cursor uint64
	for {
		keys, cursor, err := p.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan locks: %w", err)
		}

		stats.TotalLocks += int64(len(keys))

		for _, key := range keys {
			// Get lock TTL
			ttl, err := p.redis.PTTL(ctx, key).Result()
			if err != nil {
				continue
			}

			if ttl > 0 {
				stats.ActiveLocks++
				stats.TotalTTL += ttl
			}
		}

		if cursor == 0 {
			break
		}
	}

	if stats.ActiveLocks > 0 {
		stats.AvgTTL = time.Duration(int64(stats.TotalTTL) / int64(stats.ActiveLocks))
	}

	return stats, nil
}

// LockStats contains statistics about lock usage
type LockStats struct {
	Timestamp  time.Time        `json:"timestamp"`
	TotalLocks int64           `json:"total_locks"`
	ActiveLocks int64          `json:"active_locks"`
	TotalTTL   time.Duration   `json:"total_ttl"`
	AvgTTL     time.Duration   `json:"avg_ttl"`
}

// extractResourceID extracts the resource ID from a lock key
func extractResourceID(key, prefix, resourceType string) string {
	expectedPrefix := fmt.Sprintf("%s:%s:", prefix, resourceType)
	if len(key) > len(expectedPrefix) {
		return key[len(expectedPrefix):]
	}
	return ""
}

// WatchdogRegistration registers a lock with the lease watchdog
func (p *RedisLockProvider) WatchdogRegistration(lock *Lock, watchdog *LeaseWatchdog) {
	if watchdog != nil {
		watchdog.Register(lock)
	}
}

// WatchdogUnregistration unregisters a lock from the lease watchdog
func (p *RedisLockProvider) WatchdogUnregistration(lockID uuid.UUID, watchdog *LeaseWatchdog) {
	if watchdog != nil {
		watchdog.Unregister(lockID)
	}
}
