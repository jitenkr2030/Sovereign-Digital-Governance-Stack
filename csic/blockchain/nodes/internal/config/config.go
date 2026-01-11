package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	App          AppConfig          `mapstructure:"app"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Kafka        KafkaConfig        `mapstructure:"kafka"`
	NodeManager  NodeManagerConfig  `mapstructure:"node_manager"`
	Blockchains  map[string]BlockchainConfig `mapstructure:"blockchains"`
	Metrics      MetricsConfig      `mapstructure:"metrics"`
	Logging      LoggingConfig      `mapstructure:"logging"`
}

// AppConfig contains application-level settings
type AppConfig struct {
	Name        string   `mapstructure:"name"`
	Host        string   `mapstructure:"host"`
	Port        int      `mapstructure:"port"`
	Environment string   `mapstructure:"environment"`
	LogLevel    string   `mapstructure:"log_level"`
	APIKeys     []string `mapstructure:"api_keys"`
}

// DatabaseConfig contains PostgreSQL connection settings
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Password  string `mapstructure:"password"`
	DB        int    `mapstructure:"db"`
	KeyPrefix string `mapstructure:"key_prefix"`
	PoolSize  int    `mapstructure:"pool_size"`
}

// KafkaConfig contains Kafka settings
type KafkaConfig struct {
	Brokers       []string             `mapstructure:"brokers"`
	ConsumerGroup string               `mapstructure:"consumer_group"`
	Topics        KafkaTopicsConfig    `mapstructure:"topics"`
	Producer      KafkaProducerConfig  `mapstructure:"producer"`
}

// KafkaTopicsConfig contains Kafka topic names
type KafkaTopicsConfig struct {
	NodeEvents   string `mapstructure:"node_events"`
	NodeMetrics  string `mapstructure:"node_metrics"`
}

// KafkaProducerConfig contains Kafka producer settings
type KafkaProducerConfig struct {
	Acks    string `mapstructure:"acks"`
	Retries int    `mapstructure:"retries"`
}

// NodeManagerConfig contains node management settings
type NodeManagerConfig struct {
	HealthCheckInterval  int                   `mapstructure:"health_check_interval"`
	MaxNodesPerNetwork   int                   `mapstructure:"max_nodes_per_network"`
	RPCTimeout           int                   `mapstructure:"rpc_timeout"`
	MaxConcurrentCalls   int                   `mapstructure:"max_concurrent_calls"`
	OfflineThreshold     int                   `mapstructure:"offline_threshold"`
	AutoRecovery         AutoRecoveryConfig    `mapstructure:"auto_recovery"`
}

// AutoRecoveryConfig contains automatic recovery settings
type AutoRecoveryConfig struct {
	Enabled     bool `mapstructure:"enabled"`
	MaxRetries  int  `mapstructure:"max_retries"`
	RetryDelay  int  `mapstructure:"retry_delay"`
}

// BlockchainConfig contains blockchain network settings
type BlockchainConfig struct {
	Name            string                 `mapstructure:"name"`
	ChainID         int64                  `mapstructure:"chain_id"`
	NetworkID       int64                  `mapstructure:"network_id"`
	NativeCurrency  NativeCurrencyConfig   `mapstructure:"native_currency"`
	RPCURLs         []string               `mapstructure:"rpc_urls"`
	WSURL           string                 `mapstructure:"ws_url"`
	ExplorerURL     string                 `mapstructure:"explorer_url"`
	MinPeers        int                    `mapstructure:"min_peers"`
	BlockTime       int                    `mapstructure:"block_time"`
}

// NativeCurrencyConfig contains native currency information
type NativeCurrencyConfig struct {
	Name     string `mapstructure:"name"`
	Symbol   string `mapstructure:"symbol"`
	Decimals int    `mapstructure:"decimals"`
}

// MetricsConfig contains metrics collection settings
type MetricsConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	Port           int    `mapstructure:"port"`
	Path           string `mapstructure:"path"`
	CollectInterval int   `mapstructure:"collect_interval"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Format         string `mapstructure:"format"`
	Output         string `mapstructure:"output"`
	Level          string `mapstructure:"level"`
	IncludeCaller  bool   `mapstructure:"include_caller"`
}

// Load reads configuration from file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Read from config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./internal/config")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/csic/nodes/")

	// Read from environment variables
	v.SetEnvPrefix("CSIC")
	v.SetEnvKeyReplacer(nil)
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "blockchain-nodes")
	v.SetDefault("app.host", "0.0.0.0")
	v.SetDefault("app.port", 8081)
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.log_level", "info")

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 300)

	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.key_prefix", "csic:nodes:")
	v.SetDefault("redis.pool_size", 10)

	v.SetDefault("node_manager.health_check_interval", 30)
	v.SetDefault("node_manager.max_nodes_per_network", 10)
	v.SetDefault("node_manager.rpc_timeout", 10)
	v.SetDefault("node_manager.max_concurrent_calls", 100)
	v.SetDefault("node_manager.offline_threshold", 3)
	v.SetDefault("node_manager.auto_recovery.enabled", true)
	v.SetDefault("node_manager.auto_recovery.max_retries", 3)
	v.SetDefault("node_manager.auto_recovery.retry_delay", 60)

	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.port", 9090)
	v.SetDefault("metrics.path", "/metrics")
	v.SetDefault("metrics.collect_interval", 15)

	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("logging.level", "info")
}

// GetHealthCheckInterval returns the health check interval as a duration
func (c *Config) GetHealthCheckInterval() time.Duration {
	return time.Duration(c.NodeManager.HealthCheckInterval) * time.Second
}

// GetRPCTimeout returns the RPC timeout as a duration
func (c *Config) GetRPCTimeout() time.Duration {
	return time.Duration(c.NodeManager.RPCTimeout) * time.Second
}

// GetRetryDelay returns the retry delay as a duration
func (c *Config) GetRetryDelay() time.Duration {
	return time.Duration(c.NodeManager.AutoRecovery.RetryDelay) * time.Second
}
