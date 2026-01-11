# Quick Start Guide

This guide provides a quick introduction to the Distributed State Coordination system. After reading this guide, you should be able to integrate distributed locking into your application within minutes.

## Simple Example

The fastest way to get started is to use the WithLock helper method, which handles all the complexity of acquisition, validation, and release in a single call.

```go
package main

import (
   "context"
   "fmt"
   "time"

    "github.com/redis/go-redis/v9"
    "neam-platform/shared"
    "neam-platform/shared/src/coordination"
)

func main() {
    // Create Redis client
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    defer redisClient.Close()

    // Create logger
    logger := shared.NewLogger("coordination-demo")

    // Create lock configuration
    config := coordination.DefaultLockConfig()
    config.TTL = 30 * time.Second

    // Create lock provider and manager
    provider := coordination.NewRedisLockProvider(redisClient, config, logger)
    lockManager := coordination.NewLockManager(provider, config, logger, nil)
    defer lockManager.Stop()

    // Use the lock
    ctx := context.Background()
    err := lockManager.WithLock(ctx, "user", "user-123", func() error {
        // This code has exclusive access to user-123
        fmt.Println("Performing critical operation on user-123")
        time.Sleep(100 * time.Millisecond) // Simulate work
        fmt.Println("Operation complete")
        return nil
    })

    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

This example demonstrates the essential patterns for using distributed locks. The WithLock method ensures the lock is always released, even if the callback function returns an error. The lock prevents other processes from accessing the same resource until the callback completes.

## Step-by-Step Setup

### Step 1: Create the Redis Client

First, establish a connection to your Redis instance. For production deployments, use connection pooling and appropriate timeouts.

```go
redisClient := redis.NewClient(&redis.Options{
    Addr:            "localhost:6379",
    PoolSize:        100,
    MinIdleConns:    10,
    DialTimeout:     5 * time.Second,
    ReadTimeout:     3 * time.Second,
    WriteTimeout:    3 * time.Second,
    PoolTimeout:     10 * time.Second,
})
```

### Step 2: Configure the Lock System

Create a configuration structure with the desired behavior. The default configuration is suitable for most use cases, but you can customize it as needed.

```go
config := coordination.DefaultLockConfig()
config.TTL = 30 * time.Second
config.WatchdogInterval = 10 * time.Second
config.EnableFencingToken = true
```

### Step 3: Create the Lock Manager

Instantiate the lock provider and lock manager with your configuration. The lock manager orchestrates all lock operations.

```go
provider := coordination.NewRedisLockProvider(redisClient, config, logger)
metricsCollector := coordination.NewMetricsCollector(logger)
lockManager := coordination.NewLockManager(provider, config, logger, metricsCollector)
```

### Step 4: Use the Lock

Use the WithLock method for simple operations that should be protected by a lock.

```go
err := lockManager.WithLock(ctx, "resource-type", "resource-id", func() error {
    // Protected code here
    return nil
})
```

For complex operations that require the lock object directly, use Acquire and Release.

```go
lock, err := lockManager.Acquire(ctx, "resource-type", "resource-id")
if err != nil {
    return err
}
defer lockManager.Release(ctx, lock)

// Use lock.FencingToken if needed for external operations
```

## Common Patterns

### Pattern 1: Database Updates

When updating database records, use fencing tokens to prevent stale writes.

```go
lock, err := lockManager.Acquire(ctx, "account", accountID, 30*time.Second)
if err != nil {
    return fmt.Errorf("failed to acquire lock: %w", err)
}
defer lockManager.Release(ctx, lock)

// Perform database update with fencing token
err = db.ExecuteWithFencingToken(ctx, &db.Request{
    Query:        "UPDATE accounts SET balance = ? WHERE id = ?",
    Arguments:    []interface{}{newBalance, accountID},
    FencingToken: lock.FencingToken,
})
```

### Pattern 2: Long-Running Operations

For operations that may exceed the lock TTL, use the watchdog for automatic renewal.

```go
lock, err := lockManager.Acquire(ctx, "processing", jobID, 30*time.Second)
if err != nil {
    return err
}

// Register with watchdog
lockManager.watchdog.Register(lock)
defer func() {
    lockManager.watchdog.Unregister(lock.ID)
    lockManager.Release(ctx, lock)
}()

// Perform long-running operation
err = processJob(ctx, jobID)
```

### Pattern 3: Retry with Custom Strategy

For operations that may experience brief contention, configure a custom retry strategy.

```go
config := coordination.DefaultLockConfig()
config.RetryStrategy = coordination.NewExponentialBackoff(
    initialDelay:  10 * time.Millisecond,
    maxDelay:      500 * time.Millisecond,
    maxAttempts:   10,
    jitter:        true,
)
```

### Pattern 4: Multiple Resource Locks

When locking multiple resources, acquire locks in a consistent order to prevent deadlocks.

```go
// Always acquire user lock before account lock
userLock, err := lockManager.Acquire(ctx, "user", userID, 30*time.Second)
if err != nil {
    return err
}
defer lockManager.Release(ctx, userLock)

accountLock, err := lockManager.Acquire(ctx, "account", accountID, 30*time.Second)
if err != nil {
    return err
}
defer lockManager.Release(ctx, accountLock)

// Now both locks are held, perform the operation
```

## Monitoring

Enable metrics collection to monitor lock performance and identify contention.

```go
metricsCollector := coordination.NewMetricsCollector(logger)
lockManager := coordination.NewLockManager(provider, config, logger, metricsCollector)

// Periodically check metrics
go func() {
    for range time.Tick(30 * time.Second) {
        summary := metricsCollector.GetSummary()
        for _, s := range summary {
            if s.ContentionScore > 0.1 {
                logger.Warn("High contention detected",
                    "resource_type", s.ResourceType,
                    "contention_score", s.ContentionScore,
                )
            }
        }
    }
}()
```

## Error Handling

Handle the common error types appropriately in your application.

```go
lock, err := lockManager.Acquire(ctx, "resource", "id", 30*time.Second)
if err != nil {
    switch err {
    case coordination.ErrLockNotAcquired:
        // Lock could not be acquired after retries
        // Consider retrying later or degrading gracefully
        return fmt.Errorf("resource is busy, please try again")
    default:
        // Other errors (Redis unavailable, etc.)
        return fmt.Errorf("lock error: %w", err)
    }
}
defer lockManager.Release(ctx, lock)
```

## Next Steps

After getting the basic examples working, explore these topics:

- Review the full API documentation in API_REFERENCE.md for advanced features
- Configure deadlock detection for production deployments
- Set up contention monitoring and alerts
- Implement fencing tokens for safety-critical operations
- Tune retry strategies for your specific workload

## Troubleshooting

If you encounter issues, check the following:

1. **Redis connectivity**: Verify Redis is running and accessible
2. **Lock TTL**: Ensure operations complete within the lock TTL
3. **Resource naming**: Use consistent naming conventions
4. **Metrics**: Check metrics for contention patterns

For more detailed information, see the full documentation in the tests/coordination/README.md file.
