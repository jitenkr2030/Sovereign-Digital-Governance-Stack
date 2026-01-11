package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"csic-platform/control-layer/internal/core/ports"
)

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{client: client}, nil
}

// Close closes the Redis connection
func (c *RedisClient) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying Redis client
func (c *RedisClient) GetClient() *redis.Client {
	return c.client
}

// RedisStateRegistryRepository implements StateRepository using Redis
type RedisStateRegistryRepository struct {
	client    *RedisClient
	keyPrefix string
}

// NewRedisStateRegistryRepository creates a new Redis state registry repository
func NewRedisStateRegistryRepository(client *RedisClient) ports.StateRepository {
	return &RedisStateRegistryRepository{
		client:    client,
		keyPrefix: "control-layer:state:",
	}
}

// key returns the prefixed key
func (r *RedisStateRegistryRepository) key(k string) string {
	return r.keyPrefix + k
}

// GetState retrieves a state value
func (r *RedisStateRegistryRepository) GetState(ctx context.Context, key string) (string, error) {
	val, err := r.client.GetClient().Get(ctx, r.key(key)).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get state: %w", err)
	}
	return val, nil
}

// SetState sets a state value with TTL
func (r *RedisStateRegistryRepository) SetState(ctx context.Context, key string, value string, ttl time.Duration) error {
	if ttl == 0 {
		ttl = 5 * time.Minute // Default TTL
	}

	err := r.client.GetClient().Set(ctx, r.key(key), value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set state: %w", err)
	}
	return nil
}

// DeleteState deletes a state value
func (r *RedisStateRegistryRepository) DeleteState(ctx context.Context, key string) error {
	err := r.client.GetClient().Del(ctx, r.key(key)).Err()
	if err != nil {
		return fmt.Errorf("failed to delete state: %w", err)
	}
	return nil
}

// GetStateTTL returns the TTL of a state value
func (r *RedisStateRegistryRepository) GetStateTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.GetClient().TTL(ctx, r.key(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get state TTL: %w", err)
	}
	return ttl, nil
}

// SetStateIfNotExists sets a state value only if it doesn't exist
func (r *RedisStateRegistryRepository) SetStateIfNotExists(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}

	result, err := r.client.GetClient().SetNX(ctx, r.key(key), value, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set state if not exists: %w", err)
	}
	return result, nil
}

// IncrementState increments a numeric state value
func (r *RedisStateRegistryRepository) IncrementState(ctx context.Context, key string) (int64, error) {
	val, err := r.client.GetClient().Incr(ctx, r.key(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment state: %w", err)
	}
	return val, nil
}

// DecrementState decrements a numeric state value
func (r *RedisStateRegistryRepository) DecrementState(ctx context.Context, key string) (int64, error) {
	val, err := r.client.GetClient().Decr(ctx, r.key(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement state: %w", err)
	}
	return val, nil
}

// GetHashState retrieves all fields from a hash
func (r *RedisStateRegistryRepository) GetHashState(ctx context.Context, key string) (map[string]string, error) {
	val, err := r.client.GetClient().HGetAll(ctx, r.key(key)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash state: %w", err)
	}
	return val, nil
}

// SetHashField sets a field in a hash
func (r *RedisStateRegistryRepository) SetHashField(ctx context.Context, key, field, value string) error {
	err := r.client.GetClient().HSet(ctx, r.key(key), field, value).Err()
	if err != nil {
		return fmt.Errorf("failed to set hash field: %w", err)
	}
	return nil
}

// GetHashField gets a field from a hash
func (r *RedisStateRegistryRepository) GetHashField(ctx context.Context, key, field string) (string, error) {
	val, err := r.client.GetClient().HGet(ctx, r.key(key), field).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get hash field: %w", err)
	}
	return val, nil
}

// DeleteHashField deletes a field from a hash
func (r *RedisStateRegistryRepository) DeleteHashField(ctx context.Context, key, field string) error {
	err := r.client.GetClient().HDel(ctx, r.key(key), field).Err()
	if err != nil {
		return fmt.Errorf("failed to delete hash field: %w", err)
	}
	return nil
}

// AddToSet adds a value to a set
func (r *RedisStateRegistryRepository) AddToSet(ctx context.Context, key string, values ...string) error {
	err := r.client.GetClient().SAdd(ctx, r.key(key), values).Err()
	if err != nil {
		return fmt.Errorf("failed to add to set: %w", err)
	}
	return nil
}

// RemoveFromSet removes a value from a set
func (r *RedisStateRegistryRepository) RemoveFromSet(ctx context.Context, key string, values ...string) error {
	err := r.client.GetClient().SRem(ctx, r.key(key), values).Err()
	if err != nil {
		return fmt.Errorf("failed to remove from set: %w", err)
	}
	return nil
}

// GetSetMembers gets all members of a set
func (r *RedisStateRegistryRepository) GetSetMembers(ctx context.Context, key string) ([]string, error) {
	val, err := r.client.GetClient().SMembers(ctx, r.key(key)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get set members: %w", err)
	}
	return val, nil
}

// IsSetMember checks if a value is a member of a set
func (r *RedisStateRegistryRepository) IsSetMember(ctx context.Context, key, value string) (bool, error) {
	val, err := r.client.GetClient().SIsMember(ctx, r.key(key), value).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check set membership: %w", err)
	}
	return val, nil
}

// Publish publishes a message to a channel
func (r *RedisStateRegistryRepository) Publish(ctx context.Context, channel string, message string) error {
	err := r.client.GetClient().Publish(ctx, r.key(channel), message).Err()
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

// Subscribe subscribes to a channel and returns a channel of messages
func (r *RedisStateRegistryRepository) Subscribe(ctx context.Context, channel string) <-chan string {
	pubsub := r.client.GetClient().Subscribe(ctx, r.key(channel))
	ch := pubsub.Channel()

	result := make(chan string)
	go func() {
		for msg := range ch {
			result <- msg.Payload
		}
		close(result)
	}()

	return result
}
