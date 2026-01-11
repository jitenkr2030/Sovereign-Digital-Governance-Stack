# Distributed Lock Coordination API Reference

This document provides a complete API reference for the Distributed State Coordination system. It covers all public interfaces, types, and functions available for integration.

## LockManager Interface

The LockManager is the primary interface for acquiring and releasing distributed locks. It handles retry logic, lock validation, and automatic renewal through the lease watchdog.

### Acquire

```go
func (lm *LockManager) Acquire(ctx context.Context, resourceType, resourceID string) (*Lock, error)
```

Attempts to acquire a lock on the specified resource using the configured retry strategy. The function returns the acquired Lock object on success, or an error if acquisition fails after all retries.

The resourceType parameter identifies the category of resource being locked (e.g., "user", "order"). The resourceID parameter uniquely identifies the specific resource within that category. The lock TTL is determined by the LockConfig.TTL setting.

On success, the returned Lock object contains the unique lock ID, owner ID, fencing token, and expiration time. The lock should be released when no longer needed by calling Release.

On failure, the function returns ErrLockNotAcquired if all retry attempts are exhausted, or a context error if the context is cancelled or times out.

### Release

```go
func (lm *LockManager) Release(ctx context.Context, lock *Lock) error
```

Releases a previously acquired lock. The function validates that the lock is still held by the current owner before releasing it. This prevents releasing locks that have expired or been acquired by another process.

If the lock is valid and successfully released, the function returns nil. If the lock is no longer held (expired or owned by another process), the function returns ErrLockNotHeld. Other errors from the underlying provider are wrapped and returned.

Passing nil for the lock parameter is safe and returns nil immediately without performing any operations.

### Extend

```go
func (lm *LockManager) Extend(ctx context.Context, lock *Lock, additionalTTL time.Duration) error
```

Extends the expiration time of an existing lock by the specified additionalTTL. This is useful for long-running operations that need more time than the original TTL provided.

The function first validates that the lock is still held by the current owner, then attempts to extend the lock through the provider. On success, the lock's ExpiresAt field is updated to reflect the new expiration time.

For locks managed by a LeaseWatchdog, the watchdog will automatically extend the lock before it expires, so manual extension is typically only needed for operations that exceed the watchdog interval.

### WithLock

```go
func (lm *LockManager) WithLock(ctx context.Context, resourceType, resourceID string, callback func() error) error
```

A convenience method that acquires a lock, executes the callback function, and ensures the lock is released. This is the recommended pattern for most lock usage, as it guarantees proper cleanup even if the callback panics or returns an error.

The callback function is executed while holding the lock. If the callback returns an error, the lock is still released before the error is returned. If lock acquisition fails, the callback is never executed.

### WithLockFenced

```go
func (lm *LockManager) WithLockFenced(ctx context.Context, resourceType, resourceID string, callback func(*Lock) error) error
```

Similar to WithLock, but the callback receives the Lock object including the fencing token. This is useful for operations that need to pass the fencing token to external systems for validation.

### Stop

```go
func (lm *LockManager) Stop()
```

Stops the lock manager and its background processes, including the lease watchdog. Call this method during shutdown to ensure clean resource cleanup. After Stop is called, the LockManager should no longer be used.

## Lock Structure

The Lock structure represents an acquired distributed lock.

```go
type Lock struct {
    ID           uuid.UUID
    ResourceType string
    ResourceID   string
    OwnerID      string
    FencingToken int64
    AcquiredAt   time.Time
    ExpiresAt    time.Time
    TTL          time.Duration
}
```

The ID field is a unique identifier for this specific lock acquisition. The ResourceType and ResourceID fields identify the resource being locked. The OwnerID identifies the process that acquired the lock. The FencingToken is a monotonically increasing token that can be used to prevent stale operations.

### IsExpired

```go
func (l *Lock) IsExpired() bool
```

Returns true if the lock has expired (current time is after ExpiresAt). Expired locks should not be used for operations as another process may have acquired the lock.

### RemainingTTL

```go
func (l *Lock) RemainingTTL() time.Duration
```

Returns the remaining time until the lock expires. Returns zero if the lock has already expired.

## LockProvider Interface

The LockProvider interface defines the contract for lock implementations. The RedisLockProvider is the standard implementation.

### Acquire

```go
func (p *RedisLockProvider) Acquire(ctx context.Context, resourceType, resourceID string, ttl time.Duration) (*Lock, error)
```

Acquires a lock using Redis SET NX PX for atomic acquisition. The lock key is constructed using the configured LockPrefix. On success, the lock data is stored in Redis with the owner ID, fencing token, and timestamps serialized as JSON.

### Release

```go
func (p *RedisLockProvider) Release(ctx context.Context, lock *Lock) error
```

Releases the lock using an atomic Lua script that verifies the owner ID before deletion. This prevents accidentally releasing locks owned by other processes.

### Extend

```go
func (p *RedisLockProvider) Extend(ctx context.Context, lock *Lock, additionalTTL time.Duration) error
```

Extends the lock TTL using Redis PEXPIRE with owner verification. The lock's ExpiresAt field is updated on success.

### Validate

```go
func (p *RedisLockProvider) Validate(ctx context.Context, lock *Lock) (bool, error)
```

Validates that the lock is still held by the current owner. This is used by the LockManager before releasing or extending a lock.

### GetLock

```go
func (p *RedisLockProvider) GetLock(ctx context.Context, resourceType, resourceID string) (*Lock, error)
```

Retrieves the current state of a lock from Redis. Returns nil if the lock does not exist or has expired.

### GetLockStats

```go
func (p *RedisLockProvider) GetLockStats(ctx context.Context) (*LockStats, error)
```

Returns statistics about lock usage, including total locks and active locks with remaining TTL.

## Retry Strategies

The retry strategy interface defines how the LockManager handles lock acquisition retries.

### RetryStrategy Interface

```go
type RetryStrategy interface {
    NextBackoff() time.Duration
    Reset()
    Attempt() int
}
```

### ExponentialBackoff

```go
func NewExponentialBackoff(initialDelay, maxDelay time.Duration, maxAttempts int, jitter bool) *ExponentialBackoff
```

Creates an exponential backoff strategy with the specified parameters. The initialDelay is the starting backoff duration, maxDelay caps the backoff growth, and maxAttempts limits the total number of attempts. When jitter is true, random variation is added to prevent thundering herd problems.

### LinearBackoff

```go
func NewLinearBackoff(delay time.Duration, maxAttempts int) *LinearBackoff
```

Creates a linear backoff strategy that uses a constant delay between attempts.

## Deadlock Detector

The DeadlockDetector monitors the wait-for graph and resolves deadlocks using the configured strategy.

### NewDeadlockDetector

```go
func NewDeadlockDetector(
    provider LockProvider,
    config *DeadlockConfig,
    logger *shared.Logger,
    detectedDeadlocks chan<- *DeadlockInfo,
) *DeadlockDetector
```

Creates a new deadlock detector. The detectedDeadlocks channel receives deadlock events when deadlocks are detected and resolved.

### Start

```go
func (dd *DeadlockDetector) Start()
```

Starts the background deadlock detection loop. The detector periodically scans for cycles in the wait-for graph.

### Stop

```go
func (dd *DeadlockDetector) Stop()
```

Stops the deadlock detection loop.

### RegisterWait

```go
func (dd *DeadlockDetector) RegisterWait(txID, resourceKey string, priority int)
```

Registers that a transaction is waiting for a resource. This is called by the LockManager when a lock acquisition fails and the transaction begins waiting.

### RegisterHold

```go
func (dd *DeadlockDetector) RegisterHold(txID, resourceKey string)
```

Registers that a transaction holds a resource.

### RegisterRelease

```go
func (dd *DeadlockDetector) RegisterRelease(txID, resourceKey string)
```

Registers that a transaction has released a resource.

## Metrics Collector

The MetricsCollector tracks performance metrics for the locking system.

### NewMetricsCollector

```go
func NewMetricsCollector(logger *shared.Logger) *MetricsCollector
```

Creates a new metrics collector.

### RecordAcquisition

```go
func (mc *MetricsCollector) RecordAcquisition(resourceType string, waitTime time.Duration, success bool)
```

Records a lock acquisition attempt.

### RecordRelease

```go
func (mc *MetricsCollector) RecordRelease(resourceType string, holdTime time.Duration)
```

Records a lock release with the hold duration.

### RecordContention

```go
func (mc *MetricsCollector) RecordContention(resourceType string)
```

Records a lock contention event.

### RecordDeadlock

```go
func (mc *MetricsCollector) RecordDeadlock(resourceType string)
```

Records a deadlock detection event.

### GetGlobalMetrics

```go
func (mc *MetricsCollector) GetGlobalMetrics() GlobalMetrics
```

Returns aggregated metrics across all resource types.

### GetResourceMetrics

```go
func (mc *MetricsCollector) GetResourceMetrics(resourceType string) *ResourceMetrics
```

Returns metrics for a specific resource type.

### GetSummary

```go
func (mc *MetricsCollector) GetSummary() []*ResourceMetricsSummary
```

Returns a summary of all metrics organized by resource type.

### GetContentionScore

```go
func (mc *MetricsCollector) GetContentionScore(resourceType string) float64
```

Calculates a contention score for a resource type. Higher scores indicate more contention.

## Contention Analyzer

The ContentionAnalyzer identifies resources with excessive contention.

### NewContentionAnalyzer

```go
func NewContentionAnalyzer(collector *MetricsCollector, threshold float64, window time.Duration) *ContentionAnalyzer
```

Creates a new contention analyzer with the specified threshold and time window.

### AnalyzeContentiousResources

```go
func (ca *ContentionAnalyzer) AnalyzeContentiousResources() []*ContentionAlert
```

Analyzes resource metrics and returns alerts for resources exceeding the contention threshold.

## Wait-Die Policy

The WaitDiePolicy implements the wait-die deadlock prevention strategy.

### NewWaitDiePolicy

```go
func NewWaitDiePolicy(logger *shared.Logger) *WaitDiePolicy
```

Creates a new wait-die policy.

### CanWait

```go
func (wp *WaitDiePolicy) CanWait(requestingTx, holdingTx string) bool
```

Determines if a requesting transaction should wait for a holding transaction. Returns true if the requesting transaction is older (lower priority), false if it should die (be aborted).

### RegisterTransaction

```go
func (wp *WaitDiePolicy) RegisterTransaction(txID string)
```

Registers a new transaction with the current timestamp as its priority.

## Fencing Token Handler

The FencingTokenHandler manages fencing tokens for preventing stale operations.

### NewFencingTokenHandler

```go
func NewFencingTokenHandler() *FencingTokenHandler
```

Creates a new fencing token handler.

### ValidateToken

```go
func (fth *FencingTokenHandler) ValidateToken(resource string, token int64) bool
```

Validates that a fencing token is still valid for a resource. Returns true if the token is the latest or newer than the latest recorded token.

### UpdateToken

```go
func (fth *FencingTokenHandler) UpdateToken(resource string, token int64)
```

Updates the latest valid token for a resource.

### IncrementToken

```go
func (fth *FencingTokenHandler) IncrementToken(resource string) int64
```

Increments and returns the current fencing token for a resource.

## Error Types

The following error types are defined in the coordination package:

- `ErrLockNotAcquired`: Returned when lock acquisition fails after exhausting retries
- `ErrLockNotHeld`: Returned when attempting to release or extend a lock not held by the current owner
- `ErrLockExpired`: Returned when a lock expires during an operation
- `ErrDeadlockDetected`: Returned or logged when a potential deadlock is detected

## Configuration

### LockConfig

```go
type LockConfig struct {
    TTL                    time.Duration
    RetryStrategy          RetryStrategy
    WatchdogInterval       time.Duration
    EnableFencingToken     bool
    LockPrefix             string
    ContentionThreshold    int
    DeadlockDetectionTimeout time.Duration
}
```

### DeadlockConfig

```go
type DeadlockConfig struct {
    DetectionInterval     time.Duration
    WaitDieEnabled        bool
    WoundWaitEnabled      bool
    TimeoutEnabled        bool
    DefaultTimeout        time.Duration
    MaxWaitTime           time.Duration
    EnableFencingToken    bool
}
```

## Constants

### Default Values

The package provides helper functions for default configurations:

- `DefaultLockConfig() *LockConfig`: Returns configuration with sensible defaults
- `DefaultDeadlockConfig() *DeadlockConfig`: Returns deadlock configuration with defaults

### Alert Severity Levels

The following severity levels are defined for contention alerts:

- `SeverityLow`: Contention score is slightly above threshold
- `SeverityMedium`: Contention score is 1.5x threshold
- `SeverityHigh`: Contention score is 2x threshold
- `SeverityCritical`: Contention score is 3x threshold

### Deadlock Resolution Types

The following resolution types indicate how deadlocks were resolved:

- `ResolutionAborted`: Transaction was aborted
- `ResolutionRolledBack`: Transaction was rolled back
- `ResolutionTimeout`: Transaction timed out
- `ResolutionRecovered`: Deadlock was resolved without aborting transactions
