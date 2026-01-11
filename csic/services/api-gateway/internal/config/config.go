package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration
type Config struct {
	App         AppConfig         `mapstructure:"app"`
	Server      ServerConfig      `mapstructure:"server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	Kafka       KafkaConfig       `mapstructure:"kafka"`
	Blockchain  BlockchainConfig  `mapstructure:"blockchain"`
	Security    SecurityConfig    `mapstructure:"security"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	Monitoring  MonitoringConfig  `mapstructure:"monitoring"`
}

// AppConfig contains application metadata
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	BuildNumber string `mapstructure:"build_number"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
	MaxBodySize  int64  `mapstructure:"max_body_size"`
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
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// KafkaConfig contains Kafka broker settings
type KafkaConfig struct {
	Brokers       []string            `mapstructure:"brokers"`
	ConsumerGroup string              `mapstructure:"consumer_group"`
	Topics        KafkaTopicsConfig   `mapstructure:"topics"`
	Security      KafkaSecurityConfig `mapstructure:"security"`
}

// KafkaTopicsConfig contains Kafka topic names
type KafkaTopicsConfig struct {
	Transactions  string `mapstructure:"transactions"`
	Alerts        string `mapstructure:"alerts"`
	AuditLogs     string `mapstructure:"audit_logs"`
	ExchangeData  string `mapstructure:"exchange_data"`
	MiningMetrics string `mapstructure:"mining_metrics"`
}

// KafkaSecurityConfig contains Kafka security settings
type KafkaSecurityConfig struct {
	SASLMechanism string `mapstructure:"sasl_mechanism"`
	TLSEnabled    bool   `mapstructure:"tls_enabled"`
}

// BlockchainConfig contains blockchain node settings
type BlockchainConfig struct {
	Bitcoin  BlockchainNodeConfig  `mapstructure:"bitcoin"`
	Ethereum BlockchainNodeConfig  `mapstructure:"ethereum"`
}

// BlockchainNodeConfig contains settings for a blockchain node
type BlockchainNodeConfig struct {
	RPCURL      string `mapstructure:"rpc_url"`
	RPCUser     string `mapstructure:"rpc_user"`
	RPCPassword string `mapstructure:"rpc_password"`
	ZMQURL      string `mapstructure:"zmq_url"`
	WalletName  string `mapstructure:"wallet_name"`
	WSURL       string `mapstructure:"ws_url"`
	ChainID     int    `mapstructure:"chain_id"`
	NetworkID   int    `mapstructure:"network_id"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	JWT       JWTConfig       `mapstructure:"jwt"`
	Password  PasswordConfig  `mapstructure:"password"`
	MFA       MFAConfig       `mapstructure:"mfa"`
	Session   SessionConfig   `mapstructure:"session"`
	HSM       HSMConfig       `mapstructure:"hsm"`
}

// JWTConfig contains JWT authentication settings
type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	ExpiryHours        int    `mapstructure:"expiry_hours"`
	RefreshExpiryHours int    `mapstructure:"refresh_expiry_hours"`
	Algorithm          string `mapstructure:"algorithm"`
}

// PasswordConfig contains password policy settings
type PasswordConfig struct {
	MinLength      int  `mapstructure:"min_length"`
	RequireUppercase bool `mapstructure:"require_uppercase"`
	RequireLowercase bool `mapstructure:"require_lowercase"`
	RequireNumbers  bool `mapstructure:"require_numbers"`
	RequireSpecial  bool `mapstructure:"require_special"`
	PasswordHistory int  `mapstructure:"password_history"`
	MaxAgeDays      int  `mapstructure:"max_age_days"`
}

// MFAConfig contains multi-factor authentication settings
type MFAConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	Issuer       string `mapstructure:"issuer"`
	Window       int    `mapstructure:"window"`
	BackupCodes  int    `mapstructure:"backup_codes"`
}

// SessionConfig contains session management settings
type SessionConfig struct {
	IdleTimeout      int `mapstructure:"idle_timeout"`
	AbsoluteTimeout  int `mapstructure:"absolute_timeout"`
	ConcurrentLimit  int `mapstructure:"concurrent_limit"`
}

// HSMConfig contains Hardware Security Module settings
type HSMConfig struct {
	Provider    string `mapstructure:"provider"`
	LibraryPath string `mapstructure:"library_path"`
	Slot        int    `mapstructure:"slot"`
	KeyLabel    string `mapstructure:"key_label"`
	KeyType     string `mapstructure:"key_type"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level   string `mapstructure:"level"`
	Format  string `mapstructure:"format"`
	Output  string `mapstructure:"output"`
	Path    string `mapstructure:"path"`
}

// MonitoringConfig contains monitoring settings
type MonitoringConfig struct {
	MetricsEnabled    bool   `mapstructure:"metrics_enabled"`
	TracingEnabled    bool   `mapstructure:"tracing_enabled"`
	HealthCheckInterval int  `mapstructure:"health_check_interval"`
	MetricsPath       string `mapstructure:"metrics_path"`
	HealthPath        string `mapstructure:"health_path"`
}

// ConfigLoader handles loading configuration from files and environment
type ConfigLoader struct {
	configPath string
	vip        *viper.Viper
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(configPath string) *ConfigLoader {
	return &ConfigLoader{
		configPath: configPath,
		vip:        viper.New(),
	}
}

// Load loads configuration from the specified file path
func (cl *ConfigLoader) Load() (*Config, error) {
	cl.vip.SetConfigFile(cl.configPath)
	cl.vip.SetConfigType("yaml")

	// Enable environment variable overrides
	cl.vip.AutomaticEnv()

	// Read configuration file
	if err := cl.vip.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := cl.vip.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// LoadFromBytes loads configuration from a byte slice
func (cl *ConfigLoader) LoadFromBytes(data []byte) (*Config, error) {
	cl.vip.SetConfigType("yaml")
	
	if err := cl.vip.MergeConfigBytes(data); err != nil {
		return nil, fmt.Errorf("failed to merge config: %w", err)
	}

	var cfg Config
	if err := cl.vip.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app name is required")
	}
	if c.Server.Port <= 0 {
		return fmt.Errorf("server port must be positive")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Security.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
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

// GetIdleTimeout returns the idle timeout as a duration
func (c *ServerConfig) GetIdleTimeout() time.Duration {
	return time.Duration(c.IdleTimeout) * time.Second
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
