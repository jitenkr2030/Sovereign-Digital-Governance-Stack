package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	client    *redis.Client
	address   string
	password  string
	db        int
	poolSize  int
	keyPrefix string
	logger    *zap.Logger
}

// Ping checks Redis connectivity
func (c *RedisClient) Ping() error {
	client := redis.NewClient(&redis.Options{
		Addr:     c.address,
		Password: c.password,
		DB:       c.db,
		PoolSize: c.poolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	c.client = client
	return nil
}

// GetKey returns a prefixed key
func (c *RedisClient) GetKey(key string) string {
	return c.keyPrefix + key
}

// Get returns a value from Redis
func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, c.GetKey(key)).Result()
}

// Set sets a value in Redis with optional expiration
func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, c.GetKey(key), value, expiration).Err()
}

// Delete deletes a key from Redis
func (c *RedisClient) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.GetKey(key)).Err()
}

// Increment increments a key
func (c *RedisClient) Increment(ctx context.Context, key string) error {
	return c.client.Incr(ctx, c.GetKey(key)).Err()
}

// SetNX sets a value if it doesn't exist
func (c *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, c.GetKey(key), value, expiration).Result()
}

// CacheBlock caches a block with TTL
func (c *RedisClient) CacheBlock(ctx context.Context, blockNumber int64, data []byte, ttl time.Duration) error {
	return c.Set(ctx, fmt.Sprintf("block:%d", blockNumber), data, ttl)
}

// GetCachedBlock gets a cached block
func (c *RedisClient) GetCachedBlock(ctx context.Context, blockNumber int64) ([]byte, error) {
	data, err := c.Get(ctx, fmt.Sprintf("block:%d", blockNumber))
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

// CacheTransaction caches a transaction with TTL
func (c *RedisClient) CacheTransaction(ctx context.Context, txHash string, data []byte, ttl time.Duration) error {
	return c.Set(ctx, "tx:"+txHash, data, ttl)
}

// GetCachedTransaction gets a cached transaction
func (c *RedisClient) GetCachedTransaction(ctx context.Context, txHash string) ([]byte, error) {
	data, err := c.Get(ctx, "tx:"+txHash)
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

// CacheAddress caches address metadata with TTL
func (c *RedisClient) CacheAddress(ctx context.Context, address string, data []byte, ttl time.Duration) error {
	return c.Set(ctx, "addr:"+address, data, ttl)
}

// GetCachedAddress gets cached address metadata
func (c *RedisClient) GetCachedAddress(ctx context.Context, address string) ([]byte, error) {
	data, err := c.Get(ctx, "addr:"+address)
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

// SetIndexedBlock marks a block as indexed
func (c *RedisClient) SetIndexedBlock(ctx context.Context, blockNumber int64) error {
	return c.Set(ctx, fmt.Sprintf("indexed:%d", blockNumber), "1", 0)
}

// IsBlockIndexed checks if a block is indexed
func (c *RedisClient) IsBlockIndexed(ctx context.Context, blockNumber int64) (bool, error) {
	key := fmt.Sprintf("indexed:%d", blockNumber)
	exists, err := c.client.Exists(ctx, c.GetKey(key)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// GetLastIndexedBlock gets the last indexed block number
func (c *RedisClient) GetLastIndexedBlock(ctx context.Context) (int64, error) {
	val, err := c.Get(ctx, "last_indexed_block")
	if err != nil {
		return 0, err
	}
	var blockNum int64
	_, err = fmt.Sscanf(val, "%d", &blockNum)
	if err != nil {
		return 0, err
	}
	return blockNum, nil
}

// SetLastIndexedBlock sets the last indexed block number
func (c *RedisClient) SetLastIndexedBlock(ctx context.Context, blockNumber int64) error {
	return c.Set(ctx, "last_indexed_block", fmt.Sprintf("%d", blockNumber), 0)
}

// IncrementIndexedBlocks increments the indexed blocks counter
func (c *RedisClient) IncrementIndexedBlocks(ctx context.Context) error {
	return c.Increment(ctx, "total_indexed_blocks")
}

// GetTotalIndexedBlocks gets the total number of indexed blocks
func (c *RedisClient) GetTotalIndexedBlocks(ctx context.Context) (int64, error) {
	val, err := c.Get(ctx, "total_indexed_blocks")
	if err != nil {
		return 0, err
	}
	var count int64
	_, err = fmt.Sscanf(val, "%d", &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// AddBlockToQueue adds a block number to the processing queue
func (c *RedisClient) AddBlockToQueue(ctx context.Context, blockNumbers ...int64) error {
	for _, blockNum := range blockNumbers {
		if err := c.client.LPush(ctx, c.GetKey("block_queue"), blockNum).Err(); err != nil {
			return fmt.Errorf("failed to add block to queue: %w", err)
		}
	}
	return nil
}

// GetBlockFromQueue gets a block number from the processing queue
func (c *RedisClient) GetBlockFromQueue(ctx context.Context) (int64, error) {
	val, err := c.client.RPop(ctx, c.GetKey("block_queue")).Result()
	if err != nil {
		return 0, err
	}
	var blockNum int64
	_, err = fmt.Sscanf(val, "%d", &blockNum)
	if err != nil {
		return 0, err
	}
	return blockNum, nil
}

// CacheTokenTransfer caches a token transfer with TTL
func (c *RedisClient) CacheTokenTransfer(ctx context.Context, txHash string, data []byte, ttl time.Duration) error {
	return c.Set(ctx, "token_transfer:"+txHash, data, ttl)
}

// GetCachedTokenTransfer gets a cached token transfer
func (c *RedisClient) GetCachedTokenTransfer(ctx context.Context, txHash string) ([]byte, error) {
	data, err := c.Get(ctx, "token_transfer:"+txHash)
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

// IncrementMetrics increments various metrics counters
func (c *RedisClient) IncrementMetrics(ctx context.Context, metric string) error {
	return c.Increment(ctx, fmt.Sprintf("metrics:%s", metric))
}

// GetMetrics gets metrics values
func (c *RedisClient) GetMetrics(ctx context.Context, metric string) (int64, error) {
	val, err := c.Get(ctx, fmt.Sprintf("metrics:%s", metric))
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var count int64
	_, err = fmt.Sscanf(val, "%d", &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Close closes the Redis connection
func (c *RedisClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
