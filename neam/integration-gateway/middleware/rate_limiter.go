package middleware

import (
	"sync"
	"time"

	"neam-platform/integration-gateway/models"
)

// RateLimiter provides rate limiting functionality
type RateLimiter interface {
	// Allow checks if a request is allowed
	Allow(identity string) bool

	// GetRemaining returns remaining requests
	GetRemaining(identity string) int

	// Reset resets the rate limiter for an identity
	Reset(identity string)

	// SetLimit sets the rate limit for an identity
	SetLimit(identity string, requests int, window time.Duration)
}

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	mu       sync.RWMutex
	buckets  map[string]*bucket
	requests int
	window   time.Duration
}

// bucket represents a token bucket for an identity
type bucket struct {
	tokens    float64
	lastFill  time.Time
	limit     int
	window    time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requests int, window time.Duration) RateLimiter {
	return &TokenBucket{
		buckets:  make(map[string]*bucket),
		requests: requests,
		window:   window,
	}
}

// Allow checks if a request is allowed
func (r *TokenBucket) Allow(identity string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	b := r.getBucket(identity)
	return b.consume()
}

// GetRemaining returns remaining requests
func (r *TokenBucket) GetRemaining(identity string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b := r.getBucket(identity)
	r.mu.RUnlock()

	return b.remaining()
}

// Reset resets the rate limiter for an identity
func (r *TokenBucket) Reset(identity string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.buckets, identity)
}

// SetLimit sets the rate limit for an identity
func (r *TokenBucket) SetLimit(identity string, requests int, window time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b := r.getBucket(identity)
	b.limit = requests
	b.window = window
}

// getBucket gets or creates a bucket for an identity
func (r *TokenBucket) getBucket(identity string) *bucket {
	if b, exists := r.buckets[identity]; exists {
		return b
	}

	b := &bucket{
		tokens:   float64(r.requests),
		lastFill: time.Now(),
		limit:    r.requests,
		window:   r.window,
	}
	r.buckets[identity] = b

	return b
}

// consume attempts to consume a token
func (b *bucket) consume() bool {
	b.fill()

	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}

// remaining returns remaining tokens
func (b *bucket) remaining() int {
	b.fill()

	if b.tokens < 0 {
		return 0
	}

	return int(b.tokens)
}

// fill fills the bucket with tokens
func (b *bucket) fill() {
	now := time.Now()
	elapsed := now.Sub(b.lastFill)

	if elapsed >= b.window {
		b.tokens = float64(b.limit)
		b.lastFill = now
		return
	}

	// Calculate tokens to add
	tokensPerSecond := float64(b.limit) / b.window.Seconds()
	tokensToAdd := tokensPerSecond * elapsed.Seconds()

	b.tokens += tokensToAdd
	if b.tokens > float64(b.limit) {
		b.tokens = float64(b.limit)
	}

	b.lastFill = now
}

// DistributedRateLimiter provides distributed rate limiting
type DistributedRateLimiter struct {
	localLimiter RateLimiter
	// In production, Redis would be used for distributed limiting
	// For now, falls back to local limiting
}

// NewDistributedRateLimiter creates a distributed rate limiter
func NewDistributedRateLimiter(requests int, window time.Duration) RateLimiter {
	return &DistributedRateLimiter{
		localLimiter: NewRateLimiter(requests, window),
	}
}

// Allow checks if a request is allowed
func (d *DistributedRateLimiter) Allow(identity string) bool {
	// Try distributed check first (Redis), fall back to local
	return d.localLimiter.Allow(identity)
}

// GetRemaining returns remaining requests
func (d *DistributedRateLimiter) GetRemaining(identity string) int {
	return d.localLimiter.GetRemaining(identity)
}

// Reset resets the rate limiter for an identity
func (d *DistributedRateLimiter) Reset(identity string) {
	d.localLimiter.Reset(identity)
}

// SetLimit sets the rate limit for an identity
func (d *DistributedRateLimiter) SetLimit(identity string, requests int, window time.Duration) {
	d.localLimiter.SetLimit(identity, requests, window)
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	GlobalLimit  int
	PerUserLimit int
	Window       time.Duration
	Whitelist    []string
	Blacklist    []string
}

// ConfigurableRateLimiter provides configurable rate limiting
type ConfigurableRateLimiter struct {
	config     *RateLimitConfig
	limiters   map[string]RateLimiter
	globalLim  RateLimiter
	defaultLim RateLimiter
	mu         sync.RWMutex
}

// NewConfigurableRateLimiter creates a configurable rate limiter
func NewConfigurableRateLimiter(config *RateLimitConfig) *ConfigurableRateLimiter {
	rl := &ConfigurableRateLimiter{
		config:     config,
		limiters:   make(map[string]RateLimiter),
		globalLim:  NewRateLimiter(config.GlobalLimit, config.Window),
		defaultLim: NewRateLimiter(config.PerUserLimit, config.Window),
	}

	// Add whitelisted identities
	for _, identity := range config.Whitelist {
		rl.limiters[identity] = NewRateLimiter(config.GlobalLimit*10, config.Window)
	}

	// Add blacklisted identities (infinite limit of 0)
	for _, identity := range config.Blacklist {
		rl.limiters[identity] = NewRateLimiter(0, config.Window)
	}

	return rl
}

// Allow checks if a request is allowed
func (c *ConfigurableRateLimiter) Allow(identity string) bool {
	c.mu.RLock()
	limiter, exists := c.limiters[identity]
	c.mu.RUnlock()

	if exists {
		return limiter.Allow(identity)
	}

	// Check global limit
	if !c.globalLim.Allow("global") {
		return false
	}

	// Use default limiter for unknown identities
	return c.defaultLim.Allow(identity)
}

// GetRemaining returns remaining requests
func (c *ConfigurableRateLimiter) GetRemaining(identity string) int {
	c.mu.RLock()
	limiter, exists := c.limiters[identity]
	c.mu.RUnlock()

	if exists {
		return limiter.GetRemaining(identity)
	}

	return c.defaultLim.GetRemaining(identity)
}

// Reset resets the rate limiter for an identity
func (c *ConfigurableRateLimiter) Reset(identity string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.limiters, identity)
}

// SetLimit sets the rate limit for an identity
func (c *ConfigurableRateLimiter) SetLimit(identity string, requests int, window time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if limiter, exists := c.limiters[identity]; exists {
		limiter.SetLimit(identity, requests, window)
	} else {
		c.limiters[identity] = NewRateLimiter(requests, window)
	}
}

// AddToWhitelist adds an identity to whitelist
func (c *ConfigurableRateLimiter) AddToWhitelist(identity string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.limiters[identity] = NewRateLimiter(c.config.GlobalLimit*10, c.config.Window)
}

// AddToBlacklist adds an identity to blacklist
func (c *ConfigurableRateLimiter) AddToBlacklist(identity string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.limiters[identity] = NewRateLimiter(0, c.config.Window)
}

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed    bool          `json:"allowed"`
	Remaining  int           `json:"remaining"`
	Limit      int           `json:"limit"`
	ResetAfter time.Duration `json:"reset_after"`
	RetryAfter time.Duration `json:"retry_after,omitempty"`
}

// CheckRateLimit checks rate limit and returns detailed result
func CheckRateLimit(limiter RateLimiter, identity string, limit int, window time.Duration) *RateLimitResult {
	allowed := limiter.Allow(identity)
	remaining := limiter.GetRemaining(identity)

	result := &RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		Limit:     limit,
	}

	if !allowed {
		result.RetryAfter = time.Second
	}

	return result
}
