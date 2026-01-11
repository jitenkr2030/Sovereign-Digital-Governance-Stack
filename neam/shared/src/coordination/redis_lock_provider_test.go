package coordination

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"neam-platform/shared"
)

// createTestRedis creates a miniredis instance for testing
func createTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	return client, s
}

// TestRedisLockProviderAcquire tests lock acquisition with real Redis
func TestRedisLockProviderAcquire(t *testing.T) {
	client, miniredis := createTestRedis(t)
	defer client.Close()
	defer miniredis.Close()

	logger := shared.NewLogger(nil)
	config := DefaultLockConfig()
	provider := NewRedisLockProvider(client, config, logger)

	t.Run("successful acquisition", func(t *testing.T) {
		miniredis.FlushAll()

		lock, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)

		assert.NoError(t, err)
		assert.NotNil(t, lock)
		assert.Equal(t, "test", lock.ResourceType)
		assert.Equal(t, "resource1", lock.ResourceID)
		assert.NotEmpty(t, lock.OwnerID)
		assert.True(t, lock.FencingToken > 0)
	})

	t.Run("duplicate acquisition fails", func(t *testing.T) {
		miniredis.FlushAll()

		// First acquisition succeeds
		lock1, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)
		assert.NotNil(t, lock1)

		// Second acquisition fails
		lock2, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.Error(t, err)
		assert.Equal(t, ErrLockNotAcquired, err)
		assert.Nil(t, lock2)
	})

	t.Run("different resources can be acquired", func(t *testing.T) {
		miniredis.FlushAll()

		lock1, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)

		lock2, err := provider.Acquire(context.Background(), "test", "resource2", 10*time.Second)
		assert.NoError(t, err)

		assert.NotEqual(t, lock1.ID, lock2.ID)
	})

	t.Run("different resource types can be acquired", func(t *testing.T) {
		miniredis.FlushAll()

		lock1, err := provider.Acquire(context.Background(), "type1", "resource1", 10*time.Second)
		assert.NoError(t, err)

		lock2, err := provider.Acquire(context.Background(), "type2", "resource1", 10*time.Second)
		assert.NoError(t, err)

		assert.NotEqual(t, lock1.ID, lock2.ID)
	})
}

// TestRedisLockProviderRelease tests lock release with real Redis
func TestRedisLockProviderRelease(t *testing.T) {
	client, miniredis := createTestRedis(t)
	defer client.Close()
	defer miniredis.Close()

	logger := shared.NewLogger(nil)
	config := DefaultLockConfig()
	provider := NewRedisLockProvider(client, config, logger)

	t.Run("successful release", func(t *testing.T) {
		miniredis.FlushAll()

		lock, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)

		err = provider.Release(context.Background(), lock)
		assert.NoError(t, err)

		// Should be able to acquire again
		lock2, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)
		assert.NotNil(t, lock2)
	})

	t.Run("release with wrong owner fails", func(t *testing.T) {
		miniredis.FlushAll()

		lock, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)

		// Create a lock with different owner
		wrongLock := &Lock{
			ID:           lock.ID,
			ResourceType: lock.ResourceType,
			ResourceID:   lock.ResourceID,
			OwnerID:      "wrong-owner",
			FencingToken: lock.FencingToken,
		}

		err = provider.Release(context.Background(), wrongLock)
		assert.Error(t, err)
		assert.Equal(t, ErrLockNotHeld, err)
	})

	t.Run("release expired lock fails", func(t *testing.T) {
		miniredis.FlushAll()

		lock, err := provider.Acquire(context.Background(), "test", "resource1", 1*time.Second)
		assert.NoError(t, err)

		// Wait for expiration
		time.Sleep(2 * time.Second)

		err = provider.Release(context.Background(), lock)
		assert.Error(t, err)
		assert.Equal(t, ErrLockNotHeld, err)
	})

	t.Run("release nil lock does nothing", func(t *testing.T) {
		miniredis.FlushAll()

		err := provider.Release(context.Background(), nil)
		assert.NoError(t, err)
	})
}

// TestRedisLockProviderExtend tests lock extension with real Redis
func TestRedisLockProviderExtend(t *testing.T) {
	client, miniredis := createTestRedis(t)
	defer client.Close()
	defer miniredis.Close()

	logger := shared.NewLogger(nil)
	config := DefaultLockConfig()
	provider := NewRedisLockProvider(client, config, logger)

	t.Run("successful extension", func(t *testing.T) {
		miniredis.FlushAll()

		lock, err := provider.Acquire(context.Background(), "test", "resource1", 5*time.Second)
		assert.NoError(t, err)

		originalExpiry := lock.ExpiresAt

		err = provider.Extend(context.Background(), lock, 10*time.Second)
		assert.NoError(t, err)

		// Expiry should be extended
		assert.True(t, lock.ExpiresAt.After(originalExpiry))
	})

	t.Run("extend with wrong owner fails", func(t *testing.T) {
		miniredis.FlushAll()

		lock, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)

		wrongLock := &Lock{
			ID:           lock.ID,
			ResourceType: lock.ResourceType,
			ResourceID:   lock.ResourceID,
			OwnerID:      "wrong-owner",
			FencingToken: lock.FencingToken,
		}

		err = provider.Extend(context.Background(), wrongLock, 10*time.Second)
		assert.Error(t, err)
		assert.Equal(t, ErrLockNotHeld, err)
	})

	t.Run("extend nil lock does nothing", func(t *testing.T) {
		miniredis.FlushAll()

		err := provider.Extend(context.Background(), nil, 10*time.Second)
		assert.NoError(t, err)
	})
}

// TestRedisLockProviderValidate tests lock validation with real Redis
func TestRedisLockProviderValidate(t *testing.T) {
	client, miniredis := createTestRedis(t)
	defer client.Close()
	defer miniredis.Close()

	logger := shared.NewLogger(nil)
	config := DefaultLockConfig()
	provider := NewRedisLockProvider(client, config, logger)

	t.Run("validate owned lock returns true", func(t *testing.T) {
		miniredis.FlushAll()

		lock, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)

		valid, err := provider.Validate(context.Background(), lock)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("validate wrong owner returns false", func(t *testing.T) {
		miniredis.FlushAll()

		lock, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)

		wrongLock := &Lock{
			ID:           lock.ID,
			ResourceType: lock.ResourceType,
			ResourceID:   lock.ResourceID,
			OwnerID:      "wrong-owner",
			FencingToken: lock.FencingToken,
		}

		valid, err := provider.Validate(context.Background(), wrongLock)
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("validate nil lock returns false", func(t *testing.T) {
		miniredis.FlushAll()

		valid, err := provider.Validate(context.Background(), nil)
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("validate expired lock returns false", func(t *testing.T) {
		miniredis.FlushAll()

		lock, err := provider.Acquire(context.Background(), "test", "resource1", 1*time.Second)
		assert.NoError(t, err)

		time.Sleep(2 * time.Second)

		valid, err := provider.Validate(context.Background(), lock)
		assert.NoError(t, err)
		assert.False(t, valid)
	})
}

// TestRedisLockProviderGetLock tests getting lock information
func TestRedisLockProviderGetLock(t *testing.T) {
	client, miniredis := createTestRedis(t)
	defer client.Close()
	defer miniredis.Close()

	logger := shared.NewLogger(nil)
	config := DefaultLockConfig()
	provider := NewRedisLockProvider(client, config, logger)

	t.Run("get existing lock", func(t *testing.T) {
		miniredis.FlushAll()

		originalLock, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)

		lock, err := provider.GetLock(context.Background(), "test", "resource1")
		assert.NoError(t, err)
		assert.NotNil(t, lock)
		assert.Equal(t, originalLock.ID, lock.ID)
		assert.Equal(t, originalLock.ResourceType, lock.ResourceType)
		assert.Equal(t, originalLock.ResourceID, lock.ResourceID)
		assert.Equal(t, originalLock.OwnerID, lock.OwnerID)
		assert.Equal(t, originalLock.FencingToken, lock.FencingToken)
	})

	t.Run("get non-existent lock returns nil", func(t *testing.T) {
		miniredis.FlushAll()

		lock, err := provider.GetLock(context.Background(), "test", "nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, lock)
	})
}

// TestRedisLockProviderGetLockStats tests getting lock statistics
func TestRedisLockProviderGetLockStats(t *testing.T) {
	client, miniredis := createTestRedis(t)
	defer client.Close()
	defer miniredis.Close()

	logger := shared.NewLogger(nil)
	config := DefaultLockConfig()
	provider := NewRedisLockProvider(client, config, logger)

	t.Run("empty stats", func(t *testing.T) {
		miniredis.FlushAll()

		stats, err := provider.GetLockStats(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(0), stats.TotalLocks)
		assert.Equal(t, int64(0), stats.ActiveLocks)
	})

	t.Run("with active locks", func(t *testing.T) {
		miniredis.FlushAll()

		_, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)

		_, err = provider.Acquire(context.Background(), "test", "resource2", 10*time.Second)
		assert.NoError(t, err)

		stats, err := provider.GetLockStats(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(2), stats.TotalLocks)
		assert.Equal(t, int64(2), stats.ActiveLocks)
	})
}

// TestBuildLockKey tests the lock key building
func TestBuildLockKey(t *testing.T) {
	client, miniredis := createTestRedis(t)
	defer client.Close()
	defer miniredis.Close()

	logger := shared.NewLogger(nil)
	config := DefaultLockConfig()
	provider := NewRedisLockProvider(client, config, logger)

	t.Run("build key with prefix", func(t *testing.T) {
		key := provider.buildLockKey("test", "resource1")
		assert.Equal(t, "neam:lock:test:resource1", key)
	})

	t.Run("build key with special characters", func(t *testing.T) {
		key := provider.buildLockKey("type-a", "resource/b+c")
		assert.Equal(t, "neam:lock:type-a:resource/b+c", key)
	})
}

// TestExtractResourceID tests resource ID extraction from lock keys
func TestExtractResourceID(t *testing.T) {
	t.Run("valid extraction", func(t *testing.T) {
		resourceID := extractResourceID("neam:lock:test:resource1", "neam:lock", "test")
		assert.Equal(t, "resource1", resourceID)
	})

	t.Run("invalid prefix", func(t *testing.T) {
		resourceID := extractResourceID("different:lock:test:resource1", "neam:lock", "test")
		assert.Empty(t, resourceID)
	})

	t.Run("empty resource ID", func(t *testing.T) {
		resourceID := extractResourceID("neam:lock:test:", "neam:lock", "test")
		assert.Empty(t, resourceID)
	})
}

// TestFencingTokenGeneratorWithRedis tests fencing token generation with Redis
func TestFencingTokenGeneratorWithRedis(t *testing.T) {
	client, miniredis := createTestRedis(t)
	defer client.Close()
	defer miniredis.Close()

	logger := shared.NewLogger(nil)
	config := DefaultLockConfig()
	config.EnableFencingToken = true
	provider := NewRedisLockProvider(client, config, logger)

	t.Run("fencing tokens are monotonically increasing", func(t *testing.T) {
		miniredis.FlushAll()

		lock1, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)

		lock2, err := provider.Acquire(context.Background(), "test", "resource2", 10*time.Second)
		assert.NoError(t, err)

		lock3, err := provider.Acquire(context.Background(), "test", "resource3", 10*time.Second)
		assert.NoError(t, err)

		assert.True(t, lock1.FencingToken < lock2.FencingToken)
		assert.True(t, lock2.FencingToken < lock3.FencingToken)
	})

	t.Run("releasing and reacquiring gives higher fencing token", func(t *testing.T) {
		miniredis.FlushAll()

		lock1, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)

		err = provider.Release(context.Background(), lock1)
		assert.NoError(t, err)

		lock2, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)

		assert.True(t, lock2.FencingToken > lock1.FencingToken)
	})
}

// TestRedisLockProviderAtomicity tests the atomic nature of lock operations
func TestRedisLockProviderAtomicity(t *testing.T) {
	client, miniredis := createTestRedis(t)
	defer client.Close()
	defer miniredis.Close()

	logger := shared.NewLogger(nil)
	config := DefaultLockConfig()
	provider := NewRedisLockProvider(client, config, logger)

	t.Run("concurrent acquisitions are atomic", func(t *testing.T) {
		miniredis.FlushAll()

		numGoroutines := 10
		successCount := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				_, err := provider.Acquire(context.Background(), "test", "resource1", 5*time.Second)
				successCount <- err == nil
			}()
		}

		successes := 0
		for i := 0; i < numGoroutines; i++ {
			if <-successCount {
				successes++
			}
		}

		assert.Equal(t, 1, successes)
	})

	t.Run("lock data is correctly stored", func(t *testing.T) {
		miniredis.FlushAll()

		lock, err := provider.Acquire(context.Background(), "test", "resource1", 10*time.Second)
		assert.NoError(t, err)

		// Get the raw key data
		key := provider.buildLockKey("test", "resource1")
		data, err := client.Get(context.Background(), key).Result()
		assert.NoError(t, err)

		var lockData map[string]interface{}
		err = json.Unmarshal([]byte(data), &lockData)
		assert.NoError(t, err)

		assert.Equal(t, lock.OwnerID, lockData["owner_id"])
		assert.Equal(t, lock.FencingToken, int64(lockData["fencing_token"].(float64)))
		assert.Equal(t, "test", lockData["resource_type"])
		assert.Equal(t, "resource1", lockData["resource_id"])
	})
}

// TestRedisLockProviderWithCustomConfig tests lock provider with custom configuration
func TestRedisLockProviderWithCustomConfig(t *testing.T) {
	client, miniredis := createTestRedis(t)
	defer client.Close()
	defer miniredis.Close()

	logger := shared.NewLogger(nil)
	config := &LockConfig{
		TTL:            5 * time.Second,
		RetryStrategy:  nil,
		WatchdogInterval: 0,
		EnableFencingToken: false,
		LockPrefix:     "custom:lock",
		ContentionThreshold: 3,
		DeadlockDetectionTimeout: 5 * time.Second,
	}
	provider := NewRedisLockProvider(client, config, logger)

	t.Run("custom prefix is used", func(t *testing.T) {
		miniredis.FlushAll()

		_, err := provider.Acquire(context.Background(), "test", "resource1", 5*time.Second)
		assert.NoError(t, err)

		// Check that the key uses the custom prefix
		exists, err := client.Exists(context.Background(), "custom:lock:test:resource1").Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), exists)
	})

	t.Run("fencing token disabled", func(t *testing.T) {
		miniredis.FlushAll()

		lock, err := provider.Acquire(context.Background(), "test", "resource1", 5*time.Second)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), lock.FencingToken) // Still set to 1 as fallback
	})
}
