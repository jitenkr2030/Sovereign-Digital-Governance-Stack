package coordination

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"neam-platform/shared"
)

// MockLockProvider is a mock implementation of LockProvider
type MockLockProvider struct {
	mock.Mock
}

func (m *MockLockProvider) Acquire(ctx context.Context, resourceType, resourceID string, ttl time.Duration) (*Lock, error) {
	args := m.Called(ctx, resourceType, resourceID, ttl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Lock), args.Error(1)
}

func (m *MockLockProvider) Release(ctx context.Context, lock *Lock) error {
	args := m.Called(ctx, lock)
	return args.Error(0)
}

func (m *MockLockProvider) Extend(ctx context.Context, lock *Lock, additionalTTL time.Duration) error {
	args := m.Called(ctx, lock, additionalTTL)
	return args.Error(0)
}

func (m *MockLockProvider) Validate(ctx context.Context, lock *Lock) (bool, error) {
	args := m.Called(ctx, lock)
	return args.Bool(0), args.Error(1)
}

func (m *MockLockProvider) GetLock(ctx context.Context, resourceType, resourceID string) (*Lock, error) {
	args := m.Called(ctx, resourceType, resourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Lock), args.Error(1)
}

// MockLogger is a simple mock logger for testing
type MockLogger struct {
	mu       sync.Mutex
	Messages []string
}

func (l *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Messages = append(l.Messages, "DEBUG: "+msg)
}

func (l *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Messages = append(l.Messages, "INFO: "+msg)
}

func (l *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Messages = append(l.Messages, "WARN: "+msg)
}

func (l *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Messages = append(l.Messages, "ERROR: "+msg)
}

// TestLockIsExpired tests the IsExpired method
func TestLockIsExpired(t *testing.T) {
	t.Run("expired lock returns true", func(t *testing.T) {
		lock := &Lock{
			ExpiresAt: time.Now().Add(-1 * time.Second),
		}
		assert.True(t, lock.IsExpired())
	})

	t.Run("valid lock returns false", func(t *testing.T) {
		lock := &Lock{
			ExpiresAt: time.Now().Add(1 * time.Minute),
		}
		assert.False(t, lock.IsExpired())
	})
}

// TestLockRemainingTTL tests the RemainingTTL method
func TestLockRemainingTTL(t *testing.T) {
	t.Run("expired lock returns zero", func(t *testing.T) {
		lock := &Lock{
			ExpiresAt: time.Now().Add(-1 * time.Second),
		}
		assert.Equal(t, time.Duration(0), lock.RemainingTTL())
	})

	t.Run("valid lock returns positive duration", func(t *testing.T) {
		lock := &Lock{
			ExpiresAt: time.Now().Add(30 * time.Second),
		}
		remaining := lock.RemainingTTL()
		assert.True(t, remaining > 0)
		assert.True(t, remaining < 31*time.Second)
	})
}

// TestExponentialBackoff tests the exponential backoff retry strategy
func TestExponentialBackoff(t *testing.T) {
	t.Run("initial backoff value", func(t *testing.T) {
		eb := NewExponentialBackoff(50*time.Millisecond, 2*time.Second, 10, false)
		backoff := eb.NextBackoff()
		assert.Equal(t, 50*time.Millisecond, backoff)
	})

	t.Run("exponential growth", func(t *testing.T) {
		eb := NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond, 10, false)

		backoffs := []time.Duration{}
		for i := 0; i < 5; i++ {
			backoff := eb.NextBackoff()
			if backoff > 0 {
				backoffs = append(backoffs, backoff)
			}
		}

		// Verify exponential growth (each should be double the previous)
		for i := 1; i < len(backoffs); i++ {
			assert.True(t, backoffs[i] >= backoffs[i-1])
		}
	})

	t.Run("max delay cap", func(t *testing.T) {
		eb := NewExponentialBackoff(10*time.Millisecond, 50*time.Millisecond, 10, false)

		for i := 0; i < 10; i++ {
			eb.NextBackoff()
		}

		// After many iterations, should be capped at max delay
		assert.Equal(t, 50*time.Millisecond, eb.currentDelay)
	})

	t.Run("retry limit", func(t *testing.T) {
		eb := NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond, 3, false)

		// First 3 attempts should return values
		assert.Equal(t, 10*time.Millisecond, eb.NextBackoff())
		assert.Equal(t, 20*time.Millisecond, eb.NextBackoff())
		assert.Equal(t, 40*time.Millisecond, eb.NextBackoff())

		// 4th attempt should return 0 (exhausted)
		assert.Equal(t, time.Duration(0), eb.NextBackoff())
		assert.Equal(t, time.Duration(0), eb.NextBackoff())
	})

	t.Run("reset", func(t *testing.T) {
		eb := NewExponentialBackoff(50*time.Millisecond, 2*time.Second, 10, false)

		// Advance the strategy
		eb.NextBackoff()
		eb.NextBackoff()

		// Reset
		eb.Reset()

		// Should start from initial delay again
		assert.Equal(t, 50*time.Millisecond, eb.currentDelay)
		assert.Equal(t, 0, eb.attempt)
	})

	t.Run("attempt count", func(t *testing.T) {
		eb := NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond, 5, false)

		for i := 0; i < 3; i++ {
			eb.NextBackoff()
		}

		assert.Equal(t, 3, eb.Attempt())
	})
}

// TestLinearBackoff tests the linear backoff retry strategy
func TestLinearBackoff(t *testing.T) {
	t.Run("constant backoff values", func(t *testing.T) {
		lb := NewLinearBackoff(100*time.Millisecond, 5)

		backoffs := []time.Duration{}
		for i := 0; i < 5; i++ {
			backoff := lb.NextBackoff()
			if backoff > 0 {
				backoffs = append(backoffs, backoff)
			}
		}

		// All should be equal
		for i := 1; i < len(backoffs); i++ {
			assert.Equal(t, backoffs[0], backoffs[i])
		}
		assert.Equal(t, 5, len(backoffs))
	})

	t.Run("retry limit", func(t *testing.T) {
		lb := NewLinearBackoff(100*time.Millisecond, 3)

		assert.Equal(t, 100*time.Millisecond, lb.NextBackoff())
		assert.Equal(t, 100*time.Millisecond, lb.NextBackoff())
		assert.Equal(t, 100*time.Millisecond, lb.NextBackoff())

		// Exhausted
		assert.Equal(t, time.Duration(0), lb.NextBackoff())
	})

	t.Run("reset", func(t *testing.T) {
		lb := NewLinearBackoff(50*time.Millisecond, 5)

		lb.NextBackoff()
		lb.NextBackoff()
		lb.Reset()

		assert.Equal(t, 0, lb.attempt)
	})
}

// TestDefaultLockConfig tests the default configuration
func TestDefaultLockConfig(t *testing.T) {
	config := DefaultLockConfig()

	assert.Equal(t, 30*time.Second, config.TTL)
	assert.NotNil(t, config.RetryStrategy)
	assert.Equal(t, 10*time.Second, config.WatchdogInterval)
	assert.True(t, config.EnableFencingToken)
	assert.Equal(t, "neam:lock", config.LockPrefix)
	assert.Equal(t, 5, config.ContentionThreshold)
	assert.Equal(t, 10*time.Second, config.DeadlockDetectionTimeout)
}

// TestLockManagerAcquire tests lock acquisition
func TestLockManagerAcquire(t *testing.T) {
	t.Run("successful acquisition", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		expectedLock := &Lock{
			ResourceType: "test",
			ResourceID:   "resource1",
		}

		mockProvider.On("Acquire", mock.Anything, "test", "resource1", mock.Anything).
			Return(expectedLock, nil)

		config := DefaultLockConfig()
		lm := NewLockManager(mockProvider, config, logger, metrics)

		lock, err := lm.Acquire(context.Background(), "test", "resource1")

		assert.NoError(t, err)
		assert.Equal(t, expectedLock, lock)
		mockProvider.AssertExpectations(t)
	})

	t.Run("retry on failure", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		expectedLock := &Lock{
			ResourceType: "test",
			ResourceID:   "resource1",
		}

		// First call fails, second succeeds
		mockProvider.On("Acquire", mock.Anything, "test", "resource1", mock.Anything).
			Return(nil, ErrLockNotAcquired).Once()
		mockProvider.On("Acquire", mock.Anything, "test", "resource1", mock.Anything).
			Return(expectedLock, nil).Once()

		config := DefaultLockConfig()
		lm := NewLockManager(mockProvider, config, logger, metrics)

		lock, err := lm.Acquire(context.Background(), "test", "resource1")

		assert.NoError(t, err)
		assert.Equal(t, expectedLock, lock)
		mockProvider.AssertExpectations(t)
	})

	t.Run("exhausted retries returns error", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		// Always fail
		mockProvider.On("Acquire", mock.Anything, "test", "resource1", mock.Anything).
			Return(nil, ErrLockNotAcquired)

		config := DefaultLockConfig()
		// Set max attempts to 3
		config.RetryStrategy = NewLinearBackoff(10*time.Millisecond, 3)
		lm := NewLockManager(mockProvider, config, logger, metrics)

		lock, err := lm.Acquire(context.Background(), "test", "resource1")

		assert.Error(t, err)
		assert.Equal(t, ErrLockNotAcquired, err)
		assert.Nil(t, lock)
	})

	t.Run("context cancellation", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		// Cancel immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockProvider.On("Acquire", mock.Anything, "test", "resource1", mock.Anything).
			Return(nil, errors.New("context cancelled"))

		config := DefaultLockConfig()
		lm := NewLockManager(mockProvider, config, logger, metrics)

		lock, err := lm.Acquire(ctx, "test", "resource1")

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.Nil(t, lock)
	})
}

// TestLockManagerRelease tests lock release
func TestLockManagerRelease(t *testing.T) {
	t.Run("successful release", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		lock := &Lock{
			ResourceType: "test",
			ResourceID:   "resource1",
			OwnerID:      "owner1",
		}

		mockProvider.On("Validate", mock.Anything, lock).Return(true, nil)
		mockProvider.On("Release", mock.Anything, lock).Return(nil)

		config := DefaultLockConfig()
		lm := NewLockManager(mockProvider, config, logger, metrics)

		err := lm.Release(context.Background(), lock)

		assert.NoError(t, err)
		mockProvider.AssertExpectations(t)
	})

	t.Run("release nil lock does nothing", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		config := DefaultLockConfig()
		lm := NewLockManager(mockProvider, config, logger, metrics)

		err := lm.Release(context.Background(), nil)

		assert.NoError(t, err)
		mockProvider.AssertNotCalled(t, "Release")
	})

	t.Run("release invalid lock returns error", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		lock := &Lock{
			ResourceType: "test",
			ResourceID:   "resource1",
			OwnerID:      "owner1",
		}

		mockProvider.On("Validate", mock.Anything, lock).Return(false, nil)

		config := DefaultLockConfig()
		lm := NewLockManager(mockProvider, config, logger, metrics)

		err := lm.Release(context.Background(), lock)

		assert.Error(t, err)
		assert.Equal(t, ErrLockNotHeld, err)
	})
}

// TestLockManagerExtend tests lock extension
func TestLockManagerExtend(t *testing.T) {
	t.Run("successful extension", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		lock := &Lock{
			ResourceType: "test",
			ResourceID:   "resource1",
			OwnerID:      "owner1",
		}

		mockProvider.On("Validate", mock.Anything, lock).Return(true, nil)
		mockProvider.On("Extend", mock.Anything, lock, 30*time.Second).Return(nil)

		config := DefaultLockConfig()
		lm := NewLockManager(mockProvider, config, logger, metrics)

		err := lm.Extend(context.Background(), lock, 30*time.Second)

		assert.NoError(t, err)
		mockProvider.AssertExpectations(t)
	})

	t.Run("extend nil lock does nothing", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		config := DefaultLockConfig()
		lm := NewLockManager(mockProvider, config, logger, metrics)

		err := lm.Extend(context.Background(), nil, 30*time.Second)

		assert.NoError(t, err)
	})
}

// TestLockManagerWithLock tests the WithLock helper method
func TestLockManagerWithLock(t *testing.T) {
	t.Run("successful lock and callback", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		expectedLock := &Lock{
			ResourceType: "test",
			ResourceID:   "resource1",
			OwnerID:      "owner1",
		}

		callbackExecuted := false
		callbackErr := error(nil)

		mockProvider.On("Acquire", mock.Anything, "test", "resource1", mock.Anything).
			Return(expectedLock, nil)
		mockProvider.On("Validate", mock.Anything, expectedLock).Return(true, nil)
		mockProvider.On("Release", mock.Anything, expectedLock).Return(nil)

		config := DefaultLockConfig()
		lm := NewLockManager(mockProvider, config, logger, metrics)

		callback := func() error {
			callbackExecuted = true
			return callbackErr
		}

		err := lm.WithLock(context.Background(), "test", "resource1", callback)

		assert.NoError(t, err)
		assert.True(t, callbackExecuted)
		mockProvider.AssertExpectations(t)
	})

	t.Run("lock acquisition error", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		mockProvider.On("Acquire", mock.Anything, "test", "resource1", mock.Anything).
			Return(nil, ErrLockNotAcquired)

		config := DefaultLockConfig()
		lm := NewLockManager(mockProvider, config, logger, metrics)

		callback := func() error {
			return errors.New("should not be called")
		}

		err := lm.WithLock(context.Background(), "test", "resource1", callback)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to acquire lock")
	})

	t.Run("callback error still releases lock", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		expectedLock := &Lock{
			ResourceType: "test",
			ResourceID:   "resource1",
			OwnerID:      "owner1",
		}

		callbackErr := errors.New("callback failed")

		mockProvider.On("Acquire", mock.Anything, "test", "resource1", mock.Anything).
			Return(expectedLock, nil)
		mockProvider.On("Validate", mock.Anything, expectedLock).Return(true, nil)
		mockProvider.On("Release", mock.Anything, expectedLock).Return(nil)

		config := DefaultLockConfig()
		lm := NewLockManager(mockProvider, config, logger, metrics)

		callback := func() error {
			return callbackErr
		}

		err := lm.WithLock(context.Background(), "test", "resource1", callback)

		assert.Equal(t, callbackErr, err)
		mockProvider.AssertExpectations(t)
	})
}

// TestLeaseWatchdog tests the lease watchdog functionality
func TestLeaseWatchdog(t *testing.T) {
	t.Run("register and unregister", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		watchdog := NewLeaseWatchdog(mockProvider, 10*time.Second, logger, metrics)

		lock := &Lock{
			ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			ExpiresAt: time.Now().Add(5 * time.Second),
		}

		watchdog.Register(lock)
		watchdog.Unregister(lock.ID)
	})

	t.Run("start and stop", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		watchdog := NewLeaseWatchdog(mockProvider, 100*time.Millisecond, logger, metrics)
		watchdog.Start()

		// Let it run briefly
		time.Sleep(150 * time.Millisecond)

		watchdog.Stop()
	})

	t.Run("renewal on expiration", func(t *testing.T) {
		mockProvider := new(MockLockProvider)
		logger := shared.NewLogger(nil)
		metrics := NewMetricsCollector(logger)

		lock := &Lock{
			ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			ExpiresAt: time.Now().Add(1 * time.Second), // Will expire soon
			TTL:       10 * time.Second,
		}

		mockProvider.On("Extend", mock.Anything, lock, lock.TTL).Return(nil)

		watchdog := NewLeaseWatchdog(mockProvider, 100*time.Millisecond, logger, metrics)
		watchdog.Start()
		watchdog.Register(lock)

		// Wait for renewal
		time.Sleep(2 * time.Second)

		watchdog.Stop()
		mockProvider.AssertNumberOfCalls(t, "Extend", 1)
	})
}

// TestFencingTokenGenerator tests the fencing token generator
func TestFencingTokenGenerator(t *testing.T) {
	t.Run("local token generation", func(t *testing.T) {
		generator := NewFencingTokenGenerator(nil, "test-key")

		token1, err := generator.Next(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(1), token1)

		token2, err := generator.Next(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(2), token2)

		token3, err := generator.Next(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(3), token3)
	})

	t.Run("get current token", func(t *testing.T) {
		generator := NewFencingTokenGenerator(nil, "test-key")

		generator.Next(context.Background())
		generator.Next(context.Background())

		current := generator.GetCurrent()
		assert.Equal(t, int64(2), current)
	})
}

// TestLockManagerStop tests the stop functionality
func TestLockManagerStop(t *testing.T) {
	mockProvider := new(MockLockProvider)
	logger := shared.NewLogger(nil)
	metrics := NewMetricsCollector(logger)

	config := DefaultLockConfig()
	lm := NewLockManager(mockProvider, config, logger, metrics)

	// Should not panic
	lm.Stop()
}

// TestLockErrors tests custom error types
func TestLockErrors(t *testing.T) {
	assert.Equal(t, "failed to acquire lock after maximum retries", ErrLockNotAcquired.Error())
	assert.Equal(t, "lock is not held by the current owner", ErrLockNotHeld.Error())
	assert.Equal(t, "lock expired before operation completed", ErrLockExpired.Error())
	assert.Equal(t, "potential deadlock detected", ErrDeadlockDetected.Error())
}
