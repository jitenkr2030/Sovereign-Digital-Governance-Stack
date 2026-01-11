package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration
type Config struct {
	Environment string         `mapstructure:"environment"`
	Server      ServerConfig   `mapstructure:"server"`
	Postgres    PostgresConfig `mapstructure:"postgres"`
	Redis       RedisConfig    `mapstructure:"redis"`
	Kafka       KafkaConfig    `mapstructure:"kafka"`
	OpenSearch  OpenSearchConfig `mapstructure:"opensearch"`
	Metrics     MetricsConfig  `mapstructure:"metrics"`
	Logging     LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig holds HTTP and gRPC server configuration
type ServerConfig struct {
	HTTPPort         int    `mapstructure:"http_port"`
	GRPCPort         int    `mapstructure:"grpc_port"`
	ReadTimeout      int    `mapstructure:"read_timeout"`
	WriteTimeout     int    `mapstructure:"write_timeout"`
	MaxHeaderBytes   int    `mapstructure:"max_header_bytes"`
	ShutdownTimeout  int    `mapstructure:"shutdown_timeout"`
}

// PostgresConfig holds PostgreSQL database configuration
type PostgresConfig struct {
	URL                  string `mapstructure:"url"`
	Host                 string `mapstructure:"host"`
	Port                 int    `mapstructure:"port"`
	Database             string `mapstructure:"database"`
	User                 string `mapstructure:"user"`
	Password             string `mapstructure:"password"`
	MaxConnections       int    `mapstructure:"max_connections"`
	MinConnections       int    `mapstructure:"min_connections"`
	MaxConnLifetimeMinutes int  `mapstructure:"max_conn_lifetime_minutes"`
	SSLMode              string `mapstructure:"ssl_mode"`
	MigrationPath        string `mapstructure:"migration_path"`
}

// RedisConfig holds Redis cache configuration
type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	KeyPrefix    string `mapstructure:"key_prefix"`
	TTL          int    `mapstructure:"ttl"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers        []string `mapstructure:"brokers"`
	TradeTopic     string   `mapstructure:"trade_topic"`
	AlertTopic     string   `mapstructure:"alert_topic"`
	ThrottleTopic  string   `mapstructure:"throttle_topic"`
	ReportTopic    string   `mapstructure:"report_topic"`
	ConsumerGroup  string   `mapstructure:"consumer_group"`
	ProducerAck    int      `mapstructure:"producer_ack"`
	BatchSize      int      `mapstructure:"batch_size"`
	BatchTimeoutMs int      `mapstructure:"batch_timeout_ms"`
}

// OpenSearchConfig holds OpenSearch configuration
type OpenSearchConfig struct {
	Addresses         []string `mapstructure:"addresses"`
	IndexPrefix       string   `mapstructure:"index_prefix"`
	TradeIndex        string   `mapstructure:"trade_index"`
	Username          string   `mapstructure:"username"`
	Password          string   `mapstructure:"password"`
	MaxRetries        int      `mapstructure:"max_retries"`
	RetryOnStatus     []int    `mapstructure:"retry_on_status"`
	TransportTimeout  int      `mapstructure:"transport_timeout"`
	BulkFlushInterval int      `mapstructure:"bulk_flush_interval"`
}

// MetricsConfig holds metrics collection configuration
type MetricsConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	Endpoint        string `mapstructure:"endpoint"`
	Port            int    `mapstructure:"port"`
	Path            string `mapstructure:"path"`
	HistogramBuckets []float64 `mapstructure:"histogram_buckets"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
	JSONFormat bool   `mapstructure:"json_format"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Support environment variable overrides
	v.SetEnvPrefix("CSIC")
	v.AutomaticEnv()

	// Set configuration file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/csic/oversight")
	v.AddConfigPath("$HOME/.csic/oversight")

	// Try to read configuration file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, use defaults
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.http_port", 8080)
	v.SetDefault("server.grpc_port", 50051)
	v.SetDefault("server.read_timeout", 15)
	v.SetDefault("server.write_timeout", 15)
	v.SetDefault("server.max_header_bytes", 1048576)
	v.SetDefault("server.shutdown_timeout", 30)

	// PostgreSQL defaults
	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.database", "csic_oversight")
	v.SetDefault("postgres.max_connections", 25)
	v.SetDefault("postgres.min_connections", 5)
	v.SetDefault("postgres.max_conn_lifetime_minutes", 30)
	v.SetDefault("postgres.ssl_mode", "require")
	v.SetDefault("postgres.migration_path", "./migrations")

	// Redis defaults
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 100)
	v.SetDefault("redis.min_idle_conns", 10)
	v.SetDefault("redis.key_prefix", "csic:oversight:")
	v.SetDefault("redis.ttl", 3600)

	// Kafka defaults
	v.SetDefault("kafka.brokers", []string{"localhost:9092"})
	v.SetDefault("kafka.trade_topic", "market.trades.raw")
	v.SetDefault("kafka.alert_topic", "risk.alerts")
	v.SetDefault("kafka.throttle_topic", "control.commands")
	v.SetDefault("kafka.report_topic", "reporting.centralbank")
	v.SetDefault("kafka.consumer_group", "oversight-group")
	v.SetDefault("kafka.producer_ack", -1)
	v.SetDefault("kafka.batch_size", 100)
	v.SetDefault("kafka.batch_timeout_ms", 10)

	// OpenSearch defaults
	v.SetDefault("opensearch.addresses", []string{"http://localhost:9200"})
	v.SetDefault("opensearch.index_prefix", "csic")
	v.SetDefault("opensearch.trade_index", "trades")
	v.SetDefault("opensearch.max_retries", 3)
	v.SetDefault("opensearch.retry_on_status", []int{502, 503, 504})
	v.SetDefault("opensearch.transport_timeout", 30)
	v.SetDefault("opensearch.bulk_flush_interval", 5)

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.endpoint", "/metrics")
	v.SetDefault("metrics.port", 9090)
	v.SetDefault("metrics.path", "/metrics")
	v.SetDefault("metrics.histogram_buckets", []float64{0.1, 0.5, 1, 5, 10, 30, 60, 120})

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output_path", "stdout")
	v.SetDefault("logging.json_format", true)

	// Environment defaults
	v.SetDefault("environment", "development")
}

// validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Validate server configuration
	if cfg.Server.HTTPPort <= 0 || cfg.Server.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", cfg.Server.HTTPPort)
	}
	if cfg.Server.GRPCPort <= 0 || cfg.Server.GRPCPort > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", cfg.Server.GRPCPort)
	}

	// Validate PostgreSQL configuration
	if cfg.Postgres.URL == "" {
		if cfg.Postgres.Host == "" {
			return fmt.Errorf("PostgreSQL host is required when URL is not provided")
		}
	}

	// Validate Kafka configuration
	if len(cfg.Kafka.Brokers) == 0 {
		return fmt.Errorf("at least one Kafka broker is required")
	}
	if cfg.Kafka.TradeTopic == "" {
		return fmt.Errorf("Kafka trade topic is required")
	}
	if cfg.Kafka.ConsumerGroup == "" {
		return fmt.Errorf("Kafka consumer group is required")
	}

	// Validate OpenSearch configuration
	if len(cfg.OpenSearch.Addresses) == 0 {
		return fmt.Errorf("at least one OpenSearch address is required")
	}

	return nil
}

// GetPostgresDSN returns the PostgreSQL connection string
func (c *PostgresConfig) GetDSN() string {
	if c.URL != "" {
		return c.URL
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode,
	)
}

// GetConnectionMaxLifetime returns the connection max lifetime as duration
func (c *PostgresConfig) GetConnectionMaxLifetime() time.Duration {
	return time.Duration(c.MaxConnLifetimeMinutes) * time.Minute
}

// GetKeyPrefix returns the Redis key prefix with trailing colon if not present
func (c *RedisConfig) GetKeyPrefix() string {
	if c.KeyPrefix == "" {
		return ""
	}
	if c.KeyPrefix[len(c.KeyPrefix)-1] != ':' {
		return c.KeyPrefix + ":"
	}
	return c.KeyPrefix
}

// GetTTL returns the TTL as duration
func (c *RedisConfig) GetTTL() time.Duration {
	return time.Duration(c.TTL) * time.Second
}

// GetBatchTimeout returns the batch timeout as duration
func (c *KafkaConfig) GetBatchTimeout() time.Duration {
	return time.Duration(c.BatchTimeoutMs) * time.Millisecond
}

// GetBulkFlushInterval returns the bulk flush interval as duration
func (c *OpenSearchConfig) GetBulkFlushInterval() time.Duration {
	return time.Duration(c.BulkFlushInterval) * time.Second
}
