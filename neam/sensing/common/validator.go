// NEAM Sensing Layer - Common Utilities
// Validator, Anonymizer, Rate Limiter, Logger

package common

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// Config represents the sensing layer configuration
type Config struct {
	// Server configuration
	HTTPPort int `mapstructure:"http_port" yaml:"http_port" env:"NEAM_HTTP_PORT"`

	// Kafka configuration
	KafkaBrokers  string `mapstructure:"kafka_brokers" yaml:"kafka_brokers" env:"NEAM_KAFKA_BROKERS"`
	KafkaPrefix   string `mapstructure:"kafka_prefix" yaml:"kafka_prefix" env:"NEAM_KAFKA_PREFIX"`
	KafkaConsumer string `mapstructure:"kafka_consumer" yaml:"kafka_consumer" env:"NEAM_KAFKA_CONSUMER"`

	// ClickHouse configuration
	ClickHouseDSN     string `mapstructure:"clickhouse_dsn" yaml:"clickhouse_dsn" env:"NEAM_CLICKHOUSE_DSN"`
	ClickHouseDatabase string `mapstructure:"clickhouse_database" yaml:"clickhouse_database" env:"NEAM_CLICKHOUSE_DATABASE"`

	// Redis configuration
	RedisAddr     string `mapstructure:"redis_addr" yaml:"redis_addr" env:"NEAM_REDIS_ADDR"`
	RedisPassword string `mapstructure:"redis_password" yaml:"redis_password" env:"NEAM_REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"redis_db" yaml:"redis_db" env:"NEAM_REDIS_DB"`

	// Anonymization
	AnonymizationSalt string `mapstructure:"anonymization_salt" yaml:"anonymization_salt" env:"NEAM_ANONYMIZATION_SALT"`

	// Rate limiting
	RateLimitRequests int           `mapstructure:"rate_limit_requests" yaml:"rate_limit_requests"`
	RateLimitWindow   time.Duration `mapstructure:"rate_limit_window" yaml:"rate_limit_window"`

	// Data retention
	DataRetentionDays int `mapstructure:"data_retention_days" yaml:"data_retention_days" env:"NEAM_DATA_RETENTION_DAYS"`
}

// Load loads configuration from environment and config file
func (c *Config) Load() error {
	// Configuration loading logic would go here
	// For now, set defaults
	if c.HTTPPort == 0 {
		c.HTTPPort = 8082
	}
	if c.KafkaPrefix == "" {
		c.KafkaPrefix = "neam.sensing"
	}
	if c.RateLimitRequests == 0 {
		c.RateLimitRequests = 1000
	}
	if c.RateLimitWindow == 0 {
		c.RateLimitWindow = time.Minute
	}
	if c.DataRetentionDays == 0 {
		c.DataRetentionDays = 90
	}
	return nil
}

// Logger provides structured logging for the sensing layer
type Logger struct {
	*log.Logger
	mu     sync.Mutex
	fields map[string]interface{}
}

// NewLogger creates a new logger instance
func NewLogger(cfg *Config) *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "[NEAM-SENSING] ", log.LstdFlags|log.Lmicroseconds),
		fields: make(map[string]interface{}),
	}
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	newLogger := &Logger{
		Logger: l.Logger,
		fields: make(map[string]interface{}),
	}
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	newLogger.fields[key] = value
	return newLogger
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.log("DEBUG", msg, keysAndValues...)
}

// Info logs an info message
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.log("INFO", msg, keysAndValues...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.log("WARN", msg, keysAndValues...)
}

// Error logs an error message
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.log("ERROR", msg, keysAndValues...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.log("FATAL", msg, keysAndValues...)
	os.Exit(1)
}

func (l *Logger) log(level, msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	formatted := fmt.Sprintf("[%s] %s", level, msg)
	if len(keysAndValues) > 0 {
		formatted += " |"
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				formatted += fmt.Sprintf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
			}
		}
	}
	l.Logger.Print(formatted)
}

// Validator validates incoming economic data
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidationResult represents the result of a validation
type ValidationResult struct {
	Valid  bool
	Errors []string
}

// ValidatePayment validates payment data
func (v *Validator) ValidatePayment(data map[string]interface{}) *ValidationResult {
	errors := []string{}

	// Required fields validation
	requiredFields := []string{"timestamp", "amount", "currency", "merchant_category"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			errors = append(errors, fmt.Sprintf("missing required field: %s", field))
		}
	}

	// Amount validation
	if amount, ok := data["amount"].(float64); ok {
		if amount < 0 {
			errors = append(errors, "amount cannot be negative")
		}
	}

	// Currency validation
	if currency, ok := data["currency"].(string); ok {
		if len(currency) != 3 {
			errors = append(errors, "currency must be a 3-letter code")
		}
	}

	return &ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// ValidateEnergy validates energy telemetry data
func (v *Validator) ValidateEnergy(data map[string]interface{}) *ValidationResult {
	errors := []string{}

	requiredFields := []string{"timestamp", "region_id", "load_mw"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			errors = append(errors, fmt.Sprintf("missing required field: %s", field))
		}
	}

	return &ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// ValidateTransport validates transport and logistics data
func (v *Validator) ValidateTransport(data map[string]interface{}) *ValidationResult {
	errors := []string{}

	requiredFields := []string{"timestamp", "volume", "transport_type"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			errors = append(errors, fmt.Sprintf("missing required field: %s", field))
		}
	}

	return &ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// ValidateIndustry validates industrial production data
func (v *Validator) ValidateIndustry(data map[string]interface{}) *ValidationResult {
	errors := []string{}

	requiredFields := []string{"timestamp", "sector", "output_index"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			errors = append(errors, fmt.Sprintf("missing required field: %s", field))
		}
	}

	return &ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// Anonymizer provides data anonymization for privacy preservation
type Anonymizer struct {
	salt []byte
}

// NewAnonymizer creates a new anonymizer instance
func NewAnonymizer(salt string) *Anonymizer {
	if salt == "" {
		salt = generateRandomSalt()
	}
	return &Anonymizer{
		salt: []byte(salt),
	}
}

// HashIdentifier hashes an identifier for anonymization
func (a *Anonymizer) HashIdentifier(identifier string) string {
	hash := sha256.New()
	hash.Write(a.salt)
	hash.Write([]byte(identifier))
	return hex.EncodeToString(hash.Sum(nil))[:16]
}

// AnonymizePayment anonymizes payment data
func (a *Anonymizer) AnonymizePayment(data map[string]interface{}) map[string]interface{} {
	anonymized := make(map[string]interface{})

	for key, value := range data {
		switch key {
		case "account_id", "card_number", "customer_id":
			anonymized[key] = a.HashIdentifier(fmt.Sprintf("%v", value))
		case "merchant_id":
			anonymized[key] = a.HashIdentifier(fmt.Sprintf("%v", value))
		default:
			anonymized[key] = value
		}
	}

	return anonymized
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	limit       int
	window      time.Duration
	redisClient *redis.Client
	localTokens int
	localMu     sync.Mutex
	lastUpdate  time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:       limit,
		window:      window,
		localTokens: limit,
		lastUpdate:  time.Now(),
	}
}

// Middleware returns a Gin middleware for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.allow() {
			c.AbortWithStatusJSON(429, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})
			return
		}
		c.Next()
	}
}

// allow checks if a request should be allowed
func (rl *RateLimiter) allow() bool {
	rl.localMu.Lock()
	defer rl.localMu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate)

	// Refill tokens based on elapsed time
	if elapsed > 0 {
		refill := int(float64(elapsed) / float64(rl.window) * float64(rl.limit))
		rl.localTokens = min(rl.limit, rl.localTokens+refill)
		rl.lastUpdate = now
	}

	if rl.localTokens > 0 {
		rl.localTokens--
		return true
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func generateRandomSalt() string {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}
