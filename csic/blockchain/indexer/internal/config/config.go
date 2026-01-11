package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the blockchain indexer service
type Config struct {
	App           AppConfig            `mapstructure:"app"`
	Database      DatabaseConfig       `mapstructure:"database"`
	Redis         RedisConfig          `mapstructure:"redis"`
	Kafka         KafkaConfig          `mapstructure:"kafka"`
	Blockchain    BlockchainConfig     `mapstructure:"blockchain"`
	Indexer       IndexerConfig        `mapstructure:"indexer"`
	Elasticsearch ElasticsearchConfig  `mapstructure:"elasticsearch"`
	GraphQL       GraphQLConfig        `mapstructure:"graphql"`
	Health        HealthConfig         `mapstructure:"health"`
	Metrics       MetricsConfig        `mapstructure:"metrics"`
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
	LogLevel    string `mapstructure:"log_level"`
}

// DatabaseConfig holds PostgreSQL database configuration
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"name"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Password  string `mapstructure:"password"`
	DB        int    `mapstructure:"db"`
	KeyPrefix string `mapstructure:"key_prefix"`
	PoolSize  int    `mapstructure:"pool_size"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers       []string            `mapstructure:"brokers"`
	ConsumerGroup string              `mapstructure:"consumer_group"`
	Topics        KafkaTopicsConfig   `mapstructure:"topics"`
	Producer      KafkaProducerConfig `mapstructure:"producer"`
}

// KafkaTopicsConfig holds Kafka topic names
type KafkaTopicsConfig struct {
	Blocks       string `mapstructure:"blocks"`
	Transactions string `mapstructure:"transactions"`
	Events       string `mapstructure:"events"`
}

// KafkaProducerConfig holds Kafka producer configuration
type KafkaProducerConfig struct {
	RequiredAcks string `mapstructure:"required_acks"`
	RetryMax     int    `mapstructure:"retry_max"`
}

// BlockchainConfig holds blockchain node configuration
type BlockchainConfig struct {
	Network            string `mapstructure:"network"`
	NodeURL            string `mapstructure:"node_url"`
	WSURL              string `mapstructure:"ws_url"`
	ChainID            int    `mapstructure:"chain_id"`
	StartBlock         int64  `mapstructure:"start_block"`
	BatchSize          int    `mapstructure:"batch_size"`
	ConfirmationBlocks int    `mapstructure:"confirmation_blocks"`
}

// IndexerConfig holds indexer-specific configuration
type IndexerConfig struct {
	Workers           int  `mapstructure:"workers"`
	BatchInterval     int  `mapstructure:"batch_interval"`
	ReindexEnabled    bool `mapstructure:"reindex_enabled"`
	ReindexBatchSize  int  `mapstructure:"reindex_batch_size"`
	CacheTTL          int  `mapstructure:"cache_ttl"`
}

// ElasticsearchConfig holds Elasticsearch configuration
type ElasticsearchConfig struct {
	Enabled   bool     `mapstructure:"enabled"`
	Addresses []string `mapstructure:"addresses"`
	Index     string   `mapstructure:"index"`
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
}

// GraphQLConfig holds GraphQL configuration
type GraphQLConfig struct {
	Enabled       bool `mapstructure:"enabled"`
	Playground    bool `mapstructure:"playground"`
	Introspection bool `mapstructure:"introspection"`
}

// HealthConfig holds health check configuration
type HealthConfig struct {
	Interval int `mapstructure:"interval"`
	Timeout  int `mapstructure:"timeout"`
	Retries  int `mapstructure:"retries"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
	Port     int    `mapstructure:"port"`
}

// Load reads configuration from file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Read from config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./internal/config")
	v.AddConfigPath("/etc/csic/indexer/")

	// Read environment variables
	v.SetEnvPrefix("CSIC")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal configuration
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply defaults for unset values
	cfg.applyDefaults()

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "blockchain-indexer")
	v.SetDefault("app.host", "0.0.0.0")
	v.SetDefault("app.port", 8087)
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.log_level", "info")

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.username", "csic_user")
	v.SetDefault("database.password", "csic_password")
	v.SetDefault("database.name", "csic_blockchain")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 50)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", 300)

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.key_prefix", "csic:indexer:")
	v.SetDefault("redis.pool_size", 20)

	// Kafka defaults
	v.SetDefault("kafka.brokers", []string{"localhost:9092"})
	v.SetDefault("kafka.consumer_group", "indexer-consumer")
	v.SetDefault("kafka.topics.blocks", "csic.blocks")
	v.SetDefault("kafka.topics.transactions", "csic.transactions")
	v.SetDefault("kafka.topics.events", "csic.events")
	v.SetDefault("kafka.producer.required_acks", "all")
	v.SetDefault("kafka.producer.retry_max", 3)

	// Blockchain defaults
	v.SetDefault("blockchain.network", "mainnet")
	v.SetDefault("blockchain.node_url", "http://localhost:8545")
	v.SetDefault("blockchain.ws_url", "ws://localhost:8546")
	v.SetDefault("blockchain.chain_id", 1)
	v.SetDefault("blockchain.start_block", 0)
	v.SetDefault("blockchain.batch_size", 100)
	v.SetDefault("blockchain.confirmation_blocks", 12)

	// Indexer defaults
	v.SetDefault("indexer.workers", 4)
	v.SetDefault("indexer.batch_interval", 1000)
	v.SetDefault("indexer.reindex_enabled", false)
	v.SetDefault("indexer.reindex_batch_size", 1000)
	v.SetDefault("indexer.cache_ttl", 3600)

	// Elasticsearch defaults
	v.SetDefault("elasticsearch.enabled", false)
	v.SetDefault("elasticsearch.addresses", []string{"http://localhost:9200"})
	v.SetDefault("elasticsearch.index", "csic-blocks")
	v.SetDefault("elasticsearch.username", "")
	v.SetDefault("elasticsearch.password", "")

	// GraphQL defaults
	v.SetDefault("graphql.enabled", true)
	v.SetDefault("graphql.playground", true)
	v.SetDefault("graphql.introspection", true)

	// Health defaults
	v.SetDefault("health.interval", 30)
	v.SetDefault("health.timeout", 10)
	v.SetDefault("health.retries", 3)

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.endpoint", "/metrics")
	v.SetDefault("metrics.port", 9090)
}

// applyDefaults applies default values for unset configuration
func (c *Config) applyDefaults() {
	if c.App.Name == "" {
		c.App.Name = "blockchain-indexer"
	}
	if c.App.Host == "" {
		c.App.Host = "0.0.0.0"
	}
	if c.App.Port == 0 {
		c.App.Port = 8087
	}
	if c.App.Environment == "" {
		c.App.Environment = "development"
	}
	if c.App.LogLevel == "" {
		c.App.LogLevel = "info"
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 50
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 10
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 20
	}
	if c.Kafka.Producer.RetryMax == 0 {
		c.Kafka.Producer.RetryMax = 3
	}
	if c.Indexer.Workers == 0 {
		c.Indexer.Workers = 4
	}
	if c.Blockchain.BatchSize == 0 {
		c.Blockchain.BatchSize = 100
	}
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode,
	)
}

// GetAddress returns the Redis address
func (c *RedisConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddress returns the server address
func (c *AppConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
