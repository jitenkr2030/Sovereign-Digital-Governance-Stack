package coordination

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"neam-platform/shared"
)

// ErrLockNotAcquired indicates that a lock could not be acquired after exhausting retries
var ErrLockNotAcquired = errors.New("failed to acquire lock after maximum retries")

// ErrLockNotHeld indicates that the lock is not currently held
var ErrLockNotHeld = errors.New("lock is not held by the current owner")

// ErrLockExpired indicates that the lock expired before the operation completed
var ErrLockExpired = errors.New("lock expired before operation completed")

// ErrDeadlockDetected indicates that a potential deadlock was detected
var ErrDeadlockDetected = errors.New("potential deadlock detected")

// Lock represents a distributed lock acquired on a resource
type Lock struct {
	ID            uuid.UUID    `json:"id"`
	ResourceType  string       `json:"resource_type"`
	ResourceID    string       `json:"resource_id"`
	OwnerID       string       `json:"owner_id"`
	FencingToken  int64        `json:"fencing_token"`
	AcquiredAt    time.Time    `json:"acquired_at"`
	ExpiresAt     time.Time    `json:"expires_at"`
	TTL           time.Duration `json:"ttl"`
	metrics       *lockMetrics
}

// IsExpired checks if the lock has expired
func (l *Lock) IsExpired() bool {
	return time.Now().After(l.ExpiresAt)
}

// RemainingTTL returns the remaining time until expiration
func (l *Lock) RemainingTTL() time.Duration {
	remaining := time.Until(l.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// LockMetrics contains performance metrics for a lock
type lockMetrics struct {
	mu             sync.RWMutex
	acquireAttempts int64
	acquireSuccess  int64
	acquireFail     int64
	totalWaitTime   time.Duration
	holdDuration    time.Duration
	contentions     int64
}

// RetryStrategy defines how lock acquisition retries should be handled
type RetryStrategy interface {
	// NextBackoff returns the next backoff duration, or 0 if retries are exhausted
	NextBackoff() time.Duration
	// Reset resets the strategy to initial state
	Reset()
	// Attempt returns the current attempt number
	Attempt() int
}

// ExponentialBackoff implements exponential backoff retry strategy
type ExponentialBackoff struct {
	initialDelay  time.Duration
	maxDelay      time.Duration
	maxAttempts   int
	currentDelay  time.Duration
	attempt       int
	jitter        bool
}

// NewExponentialBackoff creates a new exponential backoff strategy
func NewExponentialBackoff(initialDelay, maxDelay time.Duration, maxAttempts int, jitter bool) *ExponentialBackoff {
	return &ExponentialBackoff{
		initialDelay: initialDelay,
		maxDelay:     maxDelay,
		maxAttempts:  maxAttempts,
		currentDelay: initialDelay,
		attempt:      0,
		jitter:       jitter,
	}
}

// NextBackoff implements RetryStrategy
func (eb *ExponentialBackoff) NextBackoff() time.Duration {
	if eb.attempt >= eb.maxAttempts {
		return 0
	}

	delay := eb.currentDelay
	if eb.jitter {
		// Add random jitter between 0.5x and 1.5x of current delay
		jitterFactor := 0.5 + (float64(time.Now().UnixNano()%1000) / 1000.0)
		delay = time.Duration(float64(delay) * jitterFactor)
	}

	// Exponential increase
	eb.currentDelay *= 2
	if eb.currentDelay > eb.maxDelay {
		eb.currentDelay = eb.maxDelay
	}

	eb.attempt++
	return delay
}

// Reset implements RetryStrategy
func (eb *ExponentialBackoff) Reset() {
	eb.currentDelay = eb.initialDelay
	eb.attempt = 0
}

// Attempt implements RetryStrategy
func (eb *ExponentialBackoff) Attempt() int {
	return eb.attempt
}

// LinearBackoff implements linear backoff retry strategy
type LinearBackoff struct {
	delay      time.Duration
	maxAttempts int
	attempt    int
}

// NewLinearBackoff creates a new linear backoff strategy
func NewLinearBackoff(delay time.Duration, maxAttempts int) *LinearBackoff {
	return &LinearBackoff{
		delay:       delay,
		maxAttempts: maxAttempts,
		attempt:     0,
	}
}

// NextBackoff implements RetryStrategy
func (lb *LinearBackoff) NextBackoff() time.Duration {
	if lb.attempt >= lb.maxAttempts {
		return 0
	}
	lb.attempt++
	return lb.delay
}

// Reset implements RetryStrategy
func (lb *LinearBackoff) Reset() {
	lb.attempt = 0
}

// Attempt implements RetryStrategy
func (lb *LinearBackoff) Attempt() int {
	return lb.attempt
}

// LockConfig holds configuration for lock operations
type LockConfig struct {
	// TTL is the default lock timeout duration
	TTL time.Duration

	// RetryStrategy defines how to handle retries on contention
	RetryStrategy RetryStrategy

	// WatchdogInterval is how often to renew locks for long operations
	WatchdogInterval time.Duration

	// EnableFencingToken enables fencing token generation for zombie prevention
	EnableFencingToken bool

	// LockPrefix is the prefix for all lock keys in Redis
	LockPrefix string

	// ContentionThreshold is the number of retries before considering high contention
	ContentionThreshold int

	// DeadlockDetectionTimeout is the timeout for deadlock detection
	DeadlockDetectionTimeout time.Duration
}

// DefaultLockConfig returns the default configuration
func DefaultLockConfig() *LockConfig {
	return &LockConfig{
		TTL:                    30 * time.Second,
		RetryStrategy:          NewExponentialBackoff(50*time.Millisecond, 2*time.Second, 10, true),
		WatchdogInterval:       10 * time.Second,
		EnableFencingToken:     true,
		LockPrefix:             "neam:lock",
		ContentionThreshold:    5,
		DeadlockDetectionTimeout: 10 * time.Second,
	}
}

// LockProvider defines the interface for distributed lock implementations
type LockProvider interface {
	// Acquire attempts to acquire a lock on the specified resource
	Acquire(ctx context.Context, resourceType, resourceID string, ttl time.Duration) (*Lock, error)

	// Release releases a previously acquired lock
	Release(ctx context.Context, lock *Lock) error

	// Extend extends the TTL of an existing lock
	Extend(ctx context.Context, lock *Lock, additionalTTL time.Duration) error

	// Validate validates that the lock is still held by the current owner
	Validate(ctx context.Context, lock *Lock) (bool, error)

	// GetLock returns the current lock state for a resource
	GetLock(ctx context.Context, resourceType, resourceID string) (*Lock, error)
}

// LockManager orchestrates lock operations across providers
type LockManager struct {
	provider  LockProvider
	config    *LockConfig
	logger    *shared.Logger
	metrics   *MetricsCollector
	watchdog  *LeaseWatchdog
	mu        sync.RWMutex
}

// NewLockManager creates a new lock manager
func NewLockManager(provider LockProvider, config *LockConfig, logger *shared.Logger, metrics *MetricsCollector) *LockManager {
	if config == nil {
		config = DefaultLockConfig()
	}

	lm := &LockManager{
		provider: provider,
		config:   config,
		logger:   logger,
		metrics:  metrics,
	}

	// Start lease watchdog for long-running operations
	if config.WatchdogInterval > 0 {
		lm.watchdog = NewLeaseWatchdog(provider, config.WatchdogInterval, logger, metrics)
		lm.watchdog.Start()
	}

	return lm
}

// Acquire attempts to acquire a lock with retry logic
func (lm *LockManager) Acquire(ctx context.Context, resourceType, resourceID string) (*Lock, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	startTime := time.Now()

	lm.config.RetryStrategy.Reset()

	for {
		// Attempt to acquire the lock
		lock, err := lm.provider.Acquire(ctx, resourceType, resourceID, lm.config.TTL)
		if err == nil {
			// Success
			if lm.metrics != nil {
				lm.metrics.RecordAcquisition(resourceType, time.Since(startTime), true)
			}
			return lock, nil
		}

		// Check if we should retry
		backoff := lm.config.RetryStrategy.NextBackoff()
		if backoff == 0 {
			// Retries exhausted
			if lm.metrics != nil {
				lm.metrics.RecordAcquisition(resourceType, time.Since(startTime), false)
			}
			lm.logger.Warn("Failed to acquire lock after retries",
				"resource_type", resourceType,
				"resource_id", resourceID,
				"attempts", lm.config.RetryStrategy.Attempt(),
			)
			return nil, ErrLockNotAcquired
		}

		// Check for high contention
		if lm.config.RetryStrategy.Attempt() >= lm.config.ContentionThreshold {
			if lm.metrics != nil {
				lm.metrics.RecordContention(resourceType)
			}
			lm.logger.Warn("High lock contention detected",
				"resource_type", resourceType,
				"resource_id", resourceID,
				"attempts", lm.config.RetryStrategy.Attempt(),
			)
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			continue
		}
	}
}

// Release safely releases a lock
func (lm *LockManager) Release(ctx context.Context, lock *Lock) error {
	if lock == nil {
		return nil
	}

	// Validate lock is still held by us
	valid, err := lm.provider.Validate(ctx, lock)
	if err != nil {
		return fmt.Errorf("failed to validate lock: %w", err)
	}

	if !valid {
		return ErrLockNotHeld
	}

	// Release the lock
	err = lm.provider.Release(ctx, lock)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if lm.metrics != nil {
		lm.metrics.RecordRelease(lock.ResourceType, time.Since(lock.AcquiredAt))
	}

	return nil
}

// Extend extends the TTL of an existing lock
func (lm *LockManager) Extend(ctx context.Context, lock *Lock, additionalTTL time.Duration) error {
	if lock == nil {
		return nil
	}

	// Validate lock is still held by us
	valid, err := lm.provider.Validate(ctx, lock)
	if err != nil {
		return fmt.Errorf("failed to validate lock: %w", err)
	}

	if !valid {
		return ErrLockNotHeld
	}

	// Extend the lock
	err = lm.provider.Extend(ctx, lock, additionalTTL)
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	return nil
}

// WithLock acquires a lock, executes the callback, and ensures release
func (lm *LockManager) WithLock(ctx context.Context, resourceType, resourceID string, callback func() error) error {
	lock, err := lm.Acquire(ctx, resourceType, resourceID)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	// Ensure release even if callback panics
	defer func() {
		releaseErr := lm.Release(ctx, lock)
		if releaseErr != nil {
			lm.logger.Error("Failed to release lock",
				"error", releaseErr,
				"lock_id", lock.ID,
			)
		}
	}()

	// Execute the callback
	return callback()
}

// WithLockFenced acquires a lock with fencing token and executes callback
func (lm *LockManager) WithLockFenced(ctx context.Context, resourceType, resourceID string, callback func(*Lock) error) error {
	lock, err := lm.Acquire(ctx, resourceType, resourceID)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	// Ensure release
	defer func() {
		releaseErr := lm.Release(ctx, lock)
		if releaseErr != nil {
			lm.logger.Error("Failed to release lock",
				"error", releaseErr,
				"lock_id", lock.ID,
			)
		}
	}()

	// Execute callback with fencing token
	return callback(lock)
}

// Stop stops the lock manager and its background processes
func (lm *LockManager) Stop() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.watchdog != nil {
		lm.watchdog.Stop()
	}
}

// LeaseWatchdog automatically renews locks for long-running operations
type LeaseWatchdog struct {
	provider   LockProvider
	interval   time.Duration
	logger     *shared.Logger
	metrics    *MetricsCollector
	activeLocks sync.Map
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

// NewLeaseWatchdog creates a new lease watchdog
func NewLeaseWatchdog(provider LockProvider, interval time.Duration, logger *shared.Logger, metrics *MetricsCollector) *LeaseWatchdog {
	return &LeaseWatchdog{
		provider: provider,
		interval: interval,
		logger:   logger,
		metrics:  metrics,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the watchdog process
func (lw *LeaseWatchdog) Start() {
	lw.wg.Add(1)
	go lw.run()
}

// Stop stops the watchdog process
func (lw *LeaseWatchdog) Stop() {
	close(lw.stopCh)
	lw.wg.Wait()
}

// Register registers a lock for automatic renewal
func (lw *LeaseWatchdog) Register(lock *Lock) {
	lw.activeLocks.Store(lock.ID, lock)
}

// Unregister removes a lock from automatic renewal
func (lw *LeaseWatchdog) Unregister(lockID uuid.UUID) {
	lw.activeLocks.Delete(lockID)
}

func (lw *LeaseWatchdog) run() {
	defer lw.wg.Done()

	ticker := time.NewTicker(lw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-lw.stopCh:
			return
		case <-ticker.C:
			lw.renewAll()
		}
	}
}

func (lw *LeaseWatchdog) renewAll() {
	lw.activeLocks.Range(func(key, value interface{}) bool {
		lock := value.(*Lock)

		// Calculate remaining time
		remaining := lock.RemainingTTL()
		renewThreshold := lw.interval

		// If lock is about to expire, renew it
		if remaining < renewThreshold {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := lw.provider.Extend(ctx, lock, lock.TTL)
			cancel()

			if err != nil {
				lw.logger.Warn("Failed to renew lock",
					"lock_id", lock.ID,
					"error", err,
				)
				// Lock may have expired, remove from active locks
				lw.activeLocks.Delete(key)
			} else {
				lw.logger.Debug("Renewed lock",
					"lock_id", lock.ID,
					"new_expiry", lock.ExpiresAt,
				)
			}
		}

		return true
	})
}

// FencingTokenGenerator generates monotonically increasing fencing tokens
type FencingTokenGenerator struct {
	mu       sync.Mutex
	current  int64
	redis    *redis.Client
	key      string
}

// NewFencingTokenGenerator creates a new fencing token generator
func NewFencingTokenGenerator(redis *redis.Client, key string) *FencingTokenGenerator {
	return &FencingTokenGenerator{
		redis: redis,
		key:   key,
	}
}

// Next generates the next fencing token
func (ftg *FencingTokenGenerator) Next(ctx context.Context) (int64, error) {
	ftg.mu.Lock()
	defer ftg.mu.Unlock()

	// Try to increment Redis counter first (for distributed generation)
	if ftg.redis != nil {
		token, err := ftg.redis.Incr(ctx, ftg.key).Result()
		if err == nil {
			ftg.current = token
			return token, nil
		}
		// Fallback to local generation on Redis error
	}

	// Local generation
	ftg.current++
	return ftg.current, nil
}

// GetCurrent returns the current fencing token
func (ftg *FencingTokenGenerator) GetCurrent() int64 {
	ftg.mu.Lock()
	defer ftg.mu.Unlock()
	return ftg.current
}
