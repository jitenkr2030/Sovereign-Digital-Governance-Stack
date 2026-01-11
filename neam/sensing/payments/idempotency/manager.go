package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Status represents the idempotency check status
type Status string

const (
	StatusNotFound   Status = "NOT_FOUND"
	StatusProcessing Status = "PROCESSING"
	StatusCompleted  Status = "COMPLETED"
	StatusFailed     Status = "FAILED"
)

// CheckResult represents the result of an idempotency check
type CheckResult struct {
	Status   Status
	Existing interface{}
	TTL      time.Duration
}

// Manager handles idempotency checking using Redis
type Manager struct {
	client    *redis.Client
	logger    *zap.Logger
	keyPrefix string
	defaultTTL time.Duration
}

// NewManager creates a new idempotency manager
func NewManager(cfg RedisConfig, logger *zap.Logger) (*Manager, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  time.Duration(cfg.DialTimeout),
		ReadTimeout:  time.Duration(cfg.ReadTimeout),
		WriteTimeout: time.Duration(cfg.WriteTimeout),
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Manager{
		client:    client,
		logger:    logger,
		keyPrefix: "neam:idempotency:",
		defaultTTL: time.Duration(cfg.IdempotencyTTL),
	}, nil
}

// Check checks if a message has been processed
func (m *Manager) Check(ctx context.Context, messageHash string) (*CheckResult, error) {
	key := m.buildKey(messageHash)

	// Get existing entry
	data, err := m.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get idempotency entry: %w", err)
	}

	if len(data) == 0 {
		return &CheckResult{Status: StatusNotFound}, nil
	}

	// Parse status
	status := Status(data["status"])
	result := &CheckResult{
		Status: status,
		TTL:    m.defaultTTL,
	}

	// Parse existing data if present
	if data["data"] != "" {
		var existing interface{}
		if err := json.Unmarshal([]byte(data["data"]), &existing); err == nil {
			result.Existing = existing
		}
	}

	// Calculate remaining TTL
	if ttlStr, ok := data["ttl"]; ok {
		if ttl, err := time.ParseDuration(ttlStr); err == nil {
			result.TTL = ttl
		}
	}

	return result, nil
}

// MarkProcessing marks a message as being processed
func (m *Manager) MarkProcessing(ctx context.Context, messageHash string) error {
	key := m.buildKey(messageHash)
	now := time.Now()

	pipe := m.client.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"status":    StatusProcessing,
		"started":   now.Unix(),
		"updated":   now.Unix(),
		"data":      "",
		"attempts":  "1",
	})
	pipe.Expire(ctx, key, m.defaultTTL)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to mark message as processing: %w", err)
	}

	return nil
}

// MarkCompleted marks a message as successfully completed
func (m *Manager) MarkCompleted(ctx context.Context, messageHash string, resultData interface{}) error {
	key := m.buildKey(messageHash)
	now := time.Now()

	// Serialize result data
	resultJSON, err := json.Marshal(resultData)
	if err != nil {
		return fmt.Errorf("failed to serialize result data: %w", err)
	}

	pipe := m.client.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"status":    StatusCompleted,
		"completed": now.Unix(),
		"updated":   now.Unix(),
		"data":      string(resultJSON),
		"ttl":       m.defaultTTL.String(),
	})
	pipe.Expire(ctx, key, m.defaultTTL)
	_, err = pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to mark message as completed: %w", err)
	}

	return nil
}

// MarkFailed marks a message as failed
func (m *Manager) MarkFailed(ctx context.Context, messageHash string, errorMsg string) error {
	key := m.buildKey(messageHash)
	now := time.Now()

	pipe := m.client.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"status":    StatusFailed,
		"error":     errorMsg,
		"completed": now.Unix(),
		"updated":   now.Unix(),
		"ttl":       (m.defaultTTL / 2).String(), // Failed entries expire faster
	})
	pipe.Expire(ctx, key, m.defaultTTL/2)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to mark message as failed: %w", err)
	}

	return nil
}

// Delete removes an idempotency entry
func (m *Manager) Delete(ctx context.Context, messageHash string) error {
	key := m.buildKey(messageHash)
	err := m.client.Del(ctx, key).Err()

	if err != nil {
		return fmt.Errorf("failed to delete idempotency entry: %w", err)
	}

	return nil
}

// IncrementAttempts increments the attempt count for a message
func (m *Manager) IncrementAttempts(ctx context.Context, messageHash string) (int, error) {
	key := m.buildKey(messageHash)

	count, err := m.client.HIncrBy(ctx, key, "attempts", 1).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment attempts: %w", err)
	}

	return int(count), nil
}

// GetAttempts gets the current attempt count
func (m *Manager) GetAttempts(ctx context.Context, messageHash string) (int, error) {
	key := m.buildKey(messageHash)

	count, err := m.client.HGet(ctx, key, "attempts").Int()
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Close closes the Redis connection
func (m *Manager) Close() error {
	return m.client.Close()
}

// buildKey constructs the Redis key for a message hash
func (m *Manager) buildKey(messageHash string) string {
	return m.keyPrefix + messageHash
}

// CalculateHash calculates a SHA-256 hash for message content
func CalculateHash(message []byte) string {
	hash := sha256.Sum256(message)
	return hex.EncodeToString(hash[:])
}

// CalculateHashWithContext calculates a hash including context information
func CalculateHashWithContext(message []byte, contextInfo map[string]interface{}) string {
	// Serialize context info
	contextJSON, _ := json.Marshal(contextInfo)

	// Combine message and context
	combined := append(message, contextJSON...)

	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:])
}

// IsRetryable checks if an error allows retry
func IsRetryable(err error) bool {
	// Check for transient Redis errors
	errStr := err.Error()
	transientPatterns := []string{
		"connection refused",
		"connection timeout",
		"server busy",
		"loading dataset",
		"READONLY",
		"MOVED",
		"ASK",
	}

	for _, pattern := range transientPatterns {
		if containsPattern(errStr, pattern) {
			return true
		}
	}

	return false
}

func containsPattern(s, pattern string) bool {
	return len(s) >= len(pattern) && (s == pattern || containsSubstring(s, pattern))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// RedisConfig contains Redis configuration for idempotency manager
type RedisConfig struct {
	Addr           string
	Password       string
	DB             int
	PoolSize       int
	MinIdleConns   int
	DialTimeout    Duration
	ReadTimeout    Duration
	WriteTimeout   Duration
	IdempotencyTTL Duration
}

// Duration is a wrapper for time.Duration
type Duration time.Duration

// String returns the duration string
func (d Duration) String() string {
	return time.Duration(d).String()
}

// Import for additional packages
import "encoding/binary"
