package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the access control service
type Config struct {
	// Application configuration
	AppName string `envconfig:"APP_NAME" default:"access-control-service"`
	Host    string `envconfig:"HOST" default:"0.0.0.0"`
	Port    int    `envconfig:"PORT" default:"8080"`
	Env     string `envconfig:"ENV" default:"development"`

	// PostgreSQL configuration
	PostgresHost     string `envconfig:"POSTGRES_HOST" default:"localhost"`
	PostgresPort     int    `envconfig:"POSTGRES_PORT" default:"5432"`
	PostgresUser     string `envconfig:"POSTGRES_USER" default:"postgres"`
	PostgresPassword string `envconfig:"POSTGRES_PASSWORD" default:"postgres"`
	PostgresDB       string `envconfig:"POSTGRES_DB" default:"csic_platform"`
	PostgresSSLMode  string `envconfig:"POSTGRES_SSL_MODE" default:"disable"`
	MaxOpenConns     int    `envconfig:"POSTGRES_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns     int    `envconfig:"POSTGRES_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime  int    `envconfig:"POSTGRES_CONN_MAX_LIFETIME" default:"300"`

	// Redis configuration (for caching)
	RedisHost     string `envconfig:"REDIS_HOST" default:"localhost"`
	RedisPort     int    `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" default:""`
	RedisDB       int    `envconfig:"REDIS_DB" default:"0"`
	RedisKeyPrefix string `envconfig:"REDIS_KEY_PREFIX" default:"ac:"`

	// Kafka configuration
	KafkaBrokers       string `envconfig:"KAFKA_BROKERS" default:"localhost:9092"`
	KafkaTopic         string `envconfig:"KAFKA_TOPIC" default:"access-control-events"`
	KafkaConsumerGroup string `envconfig:"KAFKA_CONSUMER_GROUP" default:"access-control-service"`
	KafkaMinBytes      int    `envconfig:"KAFKA_MIN_BYTES" default:"1"`
	KafkaMaxBytes      int    `envconfig:"KAFKA_MAX_BYTES" default:"10485760"`
	KafkaMaxWait       int    `envconfig:"KAFKA_MAX_WAIT" default:"5"`
	KafkaStartOffset   int64  `envconfig:"KAFKA_START_OFFSET" default:"-2"` // Latest offset

	// Logging configuration
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	LogFormat   string `envconfig:"LOG_FORMAT" default:"json"`

	// Metrics configuration
	MetricsEnabled bool   `envconfig:"METRICS_ENABLED" default:"true"`
	MetricsPort    int    `envconfig:"METRICS_PORT" default:"9090"`

	// Cache configuration
	PolicyCacheTTL int `envconfig:"POLICY_CACHE_TTL" default:"300"` // seconds
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return &cfg, nil
}

// GetPostgresDSN returns the PostgreSQL connection string
func (c *Config) GetPostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresDB, c.PostgresSSLMode,
	)
}

// GetRedisAddr returns the Redis address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.RedisHost, c.RedisPort)
}

// GetConnMaxLifetimeDuration returns the connection max lifetime as a duration
func (c *Config) GetConnMaxLifetimeDuration() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}

// GetPolicyCacheTTLDuration returns the policy cache TTL as a duration
func (c *Config) GetPolicyCacheTTLDuration() time.Duration {
	return time.Duration(c.PolicyCacheTTL) * time.Second
}

// GetKafkaMaxWaitDuration returns the Kafka max wait as a duration
func (c *Config) GetKafkaMaxWaitDuration() time.Duration {
	return time.Duration(c.KafkaMaxWait) * time.Second
}
