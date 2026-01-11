package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	GRPC     GRPCConfig     `mapstructure:"grpc"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Security SecurityConfig `mapstructure:"security"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
}

// AppConfig contains application metadata
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	HTTPPort     int    `mapstructure:"http_port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// GRPCConfig contains gRPC server settings
type GRPCConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	MaxRecvMsgSize int    `mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize int    `mapstructure:"max_send_msg_size"`
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
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	KeyCacheTTL  int    `mapstructure:"key_cache_ttl"`
}

// KafkaConfig contains Kafka broker settings
type KafkaConfig struct {
	Brokers       []string         `mapstructure:"brokers"`
	ConsumerGroup string           `mapstructure:"consumer_group"`
	Topics        KafkaTopicsConfig `mapstructure:"topics"`
	Producer      KafkaProducerConfig `mapstructure:"producer"`
}

// KafkaTopicsConfig contains Kafka topic names
type KafkaTopicsConfig struct {
	Audit string `mapstructure:"audit"`
}

// KafkaProducerConfig contains Kafka producer settings
type KafkaProducerConfig struct {
	Acks   string `mapstructure:"acks"`
	Retries int   `mapstructure:"retries"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	MasterKey MasterKeyConfig `mapstructure:"master_key"`
	Rotation  RotationConfig  `mapstructure:"rotation"`
	Algorithms []string       `mapstructure:"algorithms"`
	KeyUsage  []string        `mapstructure:"key_usage"`
}

// MasterKeyConfig contains master key settings
type MasterKeyConfig struct {
	Key string `mapstructure:"key"`
}

// RotationConfig contains key rotation settings
type RotationConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	DefaultPeriodDays int  `mapstructure:"default_period_days"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// MonitoringConfig contains monitoring settings
type MonitoringConfig struct {
	MetricsEnabled     bool   `mapstructure:"metrics_enabled"`
	HealthCheckInterval int   `mapstructure:"health_check_interval"`
	HealthPath         string `mapstructure:"health_path"`
	MetricsPath        string `mapstructure:"metrics_path"`
}

// ConfigLoader handles loading configuration from files and environment
type ConfigLoader struct {
	configPath string
	vip        *viper.Viper
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(configPath string) *ConfigLoader {
	vip := viper.New()
	vip.SetConfigFile(configPath)
	vip.SetConfigType("yaml")
	vip.AutomaticEnv()
	
	return &ConfigLoader{
		configPath: configPath,
		vip:        vip,
	}
}

// Load loads configuration from the specified file path
func (cl *ConfigLoader) Load() (*Config, error) {
	if err := cl.vip.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := cl.vip.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app name is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Security.MasterKey.Key == "" {
		return fmt.Errorf("master key is required")
	}
	return nil
}

// GetReadTimeout returns the read timeout as a duration
func (c *ServerConfig) GetReadTimeout() time.Duration {
	return time.Duration(c.ReadTimeout) * time.Second
}

// GetWriteTimeout returns the write timeout as a duration
func (c *ServerConfig) GetWriteTimeout() time.Duration {
	return time.Duration(c.WriteTimeout) * time.Second
}

// GetConnMaxLifetime returns the connection max lifetime as a duration
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Name, c.SSLMode,
	)
}

// GetRedisAddr returns the Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetKeyCacheTTL returns the key cache TTL as a duration
func (c *RedisConfig) GetKeyCacheTTL() time.Duration {
	return time.Duration(c.KeyCacheTTL) * time.Second
}
