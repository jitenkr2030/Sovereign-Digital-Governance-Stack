package services

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/api-gateway/gateway/internal/core/domain"
	"github.com/api-gateway/gateway/internal/core/ports"
)

const (
	// DefaultWindowSize is the default rate limit window size.
	DefaultWindowSize = time.Minute
	// DefaultLimit is the default rate limit.
	DefaultLimit = 100
)

// RateLimitService provides rate limiting functionality.
type RateLimitService struct {
	rateLimitRepo ports.RateLimitRepository

	// In-memory sliding window counter
	counters map[string]*slidingWindow
	mutex    sync.RWMutex
}

type slidingWindow struct {
	Requests  []time.Time
	WindowSec int64
}

// NewRateLimitService creates a new RateLimitService.
func NewRateLimitService(rateLimitRepo ports.RateLimitRepository) *RateLimitService {
	return &RateLimitService{
		rateLimitRepo: rateLimitRepo,
		counters:      make(map[string]*slidingWindow),
	}
}

// CheckRateLimit checks if a request is allowed under rate limits.
func (s *RateLimitService) CheckRateLimit(
	ctx context.Context,
	identifier, routeID string,
	config *domain.RateLimitConfig,
) (bool, int, int, error) {
	if config == nil {
		// No rate limiting configured
		return true, 0, 0, nil
	}

	window := DefaultWindowSize
	limit := config.RequestsPerMinute

	if limit <= 0 {
		limit = DefaultLimit
	}

	// Calculate remaining requests
	used, err := s.countRequests(ctx, identifier, routeID, window)
	if err != nil {
		return false, 0, 0, err
	}

	remaining := limit - used
	if remaining < 0 {
		remaining = 0
	}

	if used >= limit {
		return false, limit, remaining, ErrRateLimitExceeded
	}

	// Increment counter
	if err := s.incrementCounter(ctx, identifier, routeID, window); err != nil {
		// Log error but don't fail the request
		return true, limit, remaining, nil
	}

	return true, limit, remaining, nil
}

// countRequests counts the number of requests in the current window.
func (s *RateLimitService) countRequests(ctx context.Context, identifier, routeID string, window time.Duration) (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	key := s.buildKey(identifier, routeID)
	windowStart := time.Now().Add(-window)

	if sw, ok := s.counters[key]; ok {
		count := 0
		for _, t := range sw.Requests {
			if t.After(windowStart) {
				count++
			}
		}
		return count, nil
	}

	return 0, nil
}

// incrementCounter increments the request counter.
func (s *RateLimitService) incrementCounter(ctx context.Context, identifier, routeID string, window time.Duration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	key := s.buildKey(identifier, routeID)
	now := time.Now()

	if sw, ok := s.counters[key]; ok {
		sw.Requests = append(sw.Requests, now)
	} else {
		s.counters[key] = &slidingWindow{
			Requests:  []time.Time{now},
			WindowSec: int64(window.Seconds()),
		}
	}

	return nil
}

// buildKey builds a unique key for the rate limit counter.
func (s *RateLimitService) buildKey(identifier, routeID string) string {
	return identifier + ":" + routeID
}

// Cleanup removes expired entries from the in-memory counters.
func (s *RateLimitService) Cleanup(maxAge time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)

	for key, sw := range s.counters {
		windowStart := time.Now().Add(-time.Duration(sw.WindowSec) * time.Second)
		if windowStart.Before(cutoff) {
			// Remove expired entries
			var valid []time.Time
			for _, t := range sw.Requests {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(s.counters, key)
			} else {
				sw.Requests = valid
			}
		}
	}
}

// GetUsage returns the current usage statistics for an identifier.
func (s *RateLimitService) GetUsage(ctx context.Context, identifier, routeID string, window time.Duration) (*domain.RateLimitEntry, error) {
	count, err := s.countRequests(ctx, identifier, routeID, window)
	if err != nil {
		return nil, err
	}

	return &domain.RateLimitEntry{
		ID:         domain.NewEntityID(),
		Identifier: identifier,
		RouteID:    routeID,
		Requests:   count,
		WindowStart: time.Now().Add(-window),
		ExpiresAt:  time.Now().Add(window),
	}, nil
}

// Reset resets the rate limit counter for an identifier.
func (s *RateLimitService) Reset(identifier, routeID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	key := s.buildKey(identifier, routeID)
	delete(s.counters, key)
}

// Policy represents a rate limiting policy.
type Policy interface {
	Allow(identifier string) (bool, int, error)
	GetLimit() int
	GetWindow() time.Duration
}

// SlidingWindowPolicy implements a sliding window rate limit policy.
type SlidingWindowPolicy struct {
	Limit  int
	Window time.Duration
	Store  map[string][]time.Time
	Mutex  sync.RWMutex
}

// NewSlidingWindowPolicy creates a new sliding window policy.
func NewSlidingWindowPolicy(limit int, window time.Duration) *SlidingWindowPolicy {
	return &SlidingWindowPolicy{
		Limit:  limit,
		Window: window,
		Store:  make(map[string][]time.Time),
	}
}

// Allow checks if a request is allowed.
func (p *SlidingWindowPolicy) Allow(identifier string) (bool, int, error) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-p.Window)

	// Count requests in window
	var validRequests []time.Time
	count := 0
	for _, t := range p.Store[identifier] {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
			count++
		}
	}

	p.Store[identifier] = validRequests

	if count >= p.Limit {
		return false, p.Limit - count, errors.New("rate limit exceeded")
	}

	// Add current request
	p.Store[identifier] = append(p.Store[identifier], now)

	return true, p.Limit - count - 1, nil
}

// GetLimit returns the rate limit.
func (p *SlidingWindowPolicy) GetLimit() int {
	return p.Limit
}

// GetWindow returns the rate limit window.
func (p *SlidingWindowPolicy) GetWindow() time.Duration {
	return p.Window
}

// TokenBucketPolicy implements a token bucket rate limit policy.
type TokenBucketPolicy struct {
	Capacity    int
	RefillRate  float64 // tokens per second
	Tokens      map[string]float64
	LastRefill  map[string]time.Time
	Mutex       sync.RWMutex
}

// NewTokenBucketPolicy creates a new token bucket policy.
func NewTokenBucketPolicy(capacity int, refillRate float64) *TokenBucketPolicy {
	return &TokenBucketPolicy{
		Capacity:   capacity,
		RefillRate: refillRate,
		Tokens:     make(map[string]float64),
		LastRefill: make(map[string]time.Time),
	}
}

// Allow checks if a request is allowed.
func (p *TokenBucketPolicy) Allow(identifier string) (bool, int, error) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	now := time.Now()

	// Initialize if first request
	if _, exists := p.Tokens[identifier]; !exists {
		p.Tokens[identifier] = float64(p.Capacity)
		p.LastRefill[identifier] = now
		return true, int(p.Tokens[identifier]), nil
	}

	// Calculate tokens to add
	lastRefill := p.LastRefill[identifier]
	elapsed := now.Sub(lastRefill).Seconds()
	tokensToAdd := elapsed * p.RefillRate

	// Add tokens
	p.Tokens[identifier] = p.Tokens[identifier] + tokensToAdd
	if p.Tokens[identifier] > float64(p.Capacity) {
		p.Tokens[identifier] = float64(p.Capacity)
	}
	p.LastRefill[identifier] = now

	// Check if we have tokens
	if p.Tokens[identifier] < 1 {
		return false, 0, errors.New("rate limit exceeded")
	}

	// Consume a token
	p.Tokens[identifier] = p.Tokens[identifier] - 1

	return true, int(p.Tokens[identifier]), nil
}

// GetLimit returns the rate limit (bucket capacity).
func (p *TokenBucketPolicy) GetLimit() int {
	return p.Capacity
}

// GetWindow returns the effective window (for token bucket, this is 0).
func (p *TokenBucketPolicy) GetWindow() time.Duration {
	return 0
}
