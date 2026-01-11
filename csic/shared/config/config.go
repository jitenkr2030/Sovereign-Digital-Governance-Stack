// CSIC Platform - Shared Configuration Package
// Centralized configuration management with environment support

package config

import (
    "fmt"
    "os"
    "sync"
    "time"

    "gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	App        AppConfig        `yaml:"app"`
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Redis      RedisConfig      `yaml:"redis"`
	Kafka      KafkaConfig      `yaml:"kafka"`
	Blockchain BlockchainConfig `yaml:"blockchain"`
	Security   SecurityConfig   `yaml:"security"`
	Logging    LoggingConfig    `yaml:"logging"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
	AuditLog   AuditLogConfig   `yaml:"audit_log"`
}

// AppConfig contains application metadata
type AppConfig struct {
    Name        string `yaml:"name"`
    Version     string `yaml:"version"`
    Environment string `yaml:"environment"` // development, staging, production
    BuildNumber string `yaml:"build_number"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
    Host         string `yaml:"host"`
    Port         int    `yaml:"port"`
    ReadTimeout  int    `yaml:"read_timeout"`  // seconds
    WriteTimeout int    `yaml:"write_timeout"` // seconds
    MaxBodySize  int64  `yaml:"max_body_size"` // bytes
}

// DatabaseConfig contains PostgreSQL connection settings
type DatabaseConfig struct {
    Host            string `yaml:"host"`
    Port            int    `yaml:"port"`
    Username        string `yaml:"username"`
    Password        string `yaml:"password"`
    Name            string `yaml:"name"`
    SSLMode         string `yaml:"ssl_mode"`
    MaxOpenConns    int    `yaml:"max_open_conns"`
    MaxIdleConns    int    `yaml:"max_idle_conns"`
    ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // seconds
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
    Host        string `yaml:"host"`
    Port        int    `yaml:"port"`
    Password    string `yaml:"password"`
    DB          int    `yaml:"db"`
    PoolSize    int    `yaml:"pool_size"`
    MinIdleConns int   `yaml:"min_idle_conns"`
}

// KafkaConfig contains Kafka connection settings
type KafkaConfig struct {
    Brokers       []string     `yaml:"brokers"`
    ConsumerGroup string       `yaml:"consumer_group"`
    Topics        TopicsConfig `yaml:"topics"`
    Security      KafkaSecurity `yaml:"security"`
}

// TopicsConfig contains Kafka topic names
type TopicsConfig struct {
    Transactions  string `yaml:"transactions"`
    Alerts        string `yaml:"alerts"`
    AuditLogs     string `yaml:"audit_logs"`
    ExchangeData  string `yaml:"exchange_data"`
    MiningMetrics string `yaml:"mining_metrics"`
}

// KafkaSecurity contains Kafka security settings
type KafkaSecurity struct {
    SASLMechanism string `yaml:"sasl_mechanism"`
    TLSEnabled    bool   `yaml:"tls_enabled"`
}

// BlockchainConfig contains blockchain node settings
type BlockchainConfig struct {
    Bitcoin  BitcoinConfig  `yaml:"bitcoin"`
    Ethereum EthereumConfig `yaml:"ethereum"`
}

// BitcoinConfig contains Bitcoin node settings
type BitcoinConfig struct {
    RPCURL      string `yaml:"rpc_url"`
    RPCUser     string `yaml:"rpc_user"`
    RPCPassword string `yaml:"rpc_password"`
    ZMQURL      string `yaml:"zmq_url"`
    WalletName  string `yaml:"wallet_name"`
}

// EthereumConfig contains Ethereum node settings
type EthereumConfig struct {
    RPCURL      string `yaml:"rpc_url"`
    WSURL       string `yaml:"ws_url"`
    ChainID     int    `yaml:"chain_id"`
    NetworkID   int    `yaml:"network_id"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
    JWT       JWTConfig       `yaml:"jwt"`
    Password  PasswordConfig  `yaml:"password"`
    MFA       MFAConfig       `yaml:"mfa"`
    Session   SessionConfig   `yaml:"session"`
    HSM       HSMConfig       `yaml:"hsm"`
}

// JWTConfig contains JWT settings
type JWTConfig struct {
    Secret            string `yaml:"secret"`
    ExpiryHours       int    `yaml:"expiry_hours"`
    RefreshExpiryHours int   `yaml:"refresh_expiry_hours"`
    Algorithm         string `yaml:"algorithm"`
}

// PasswordConfig contains password policy settings
type PasswordConfig struct {
    MinLength        int  `yaml:"min_length"`
    RequireUppercase bool `yaml:"require_uppercase"`
    RequireLowercase bool `yaml:"require_lowercase"`
    RequireNumbers   bool `yaml:"require_numbers"`
    RequireSpecial   bool `yaml:"require_special"`
    PasswordHistory  int  `yaml:"password_history"`
    MaxAgeDays       int  `yaml:"max_age_days"`
}

// MFAConfig contains MFA settings
type MFAConfig struct {
    Enabled     bool `yaml:"enabled"`
    Issuer      string `yaml:"issuer"`
    Window      int   `yaml:"window"`
    BackupCodes int   `yaml:"backup_codes"`
}

// SessionConfig contains session settings
type SessionConfig struct {
    IdleTimeout     int `yaml:"idle_timeout"`     // minutes
    AbsoluteTimeout int `yaml:"absolute_timeout"` // hours
    ConcurrentLimit int `yaml:"concurrent_limit"`
}

// HSMConfig contains HSM settings
type HSMConfig struct {
    Provider    string `yaml:"provider"`
    LibraryPath string `yaml:"library_path"`
    Slot        int    `yaml:"slot"`
    KeyLabel    string `yaml:"key_label"`
    KeyType     string `yaml:"key_type"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
    Level   string `yaml:"level"`   // DEBUG, INFO, WARN, ERROR
    Format  string `yaml:"format"`  // json, text
    Output  string `yaml:"output"`  // stdout, file
    Path    string `yaml:"path"`    // file path if output is file
}

// MonitoringConfig contains monitoring settings
type MonitoringConfig struct {
	MetricsEnabled  bool   `yaml:"metrics_enabled"`
	TracingEnabled  bool   `yaml:"tracing_enabled"`
	HealthCheckIntv int    `yaml:"health_check_interval"` // seconds
	MetricsPath     string `yaml:"metrics_path"`
	HealthPath      string `yaml:"health_path"`
}

// AuditLogConfig contains audit log service settings
type AuditLogConfig struct {
	StoragePath      string `yaml:"storage_path"`
	SealInterval     int    `yaml:"seal_interval"` // seconds
	ChainFilePath    string `yaml:"chain_file_path"`
	VerificationPath string `yaml:"verification_path"`
	RetentionDays    int    `yaml:"retention_days"`
	EnableWORM       bool   `yaml:"enable_worm"`
	MaxFileSize      int64  `yaml:"max_file_size"`
	EntriesPerFile   int    `yaml:"entries_per_file"`
}

// ConfigLoader handles configuration loading
type ConfigLoader struct {
    config   *Config
    mu       sync.RWMutex
    filePath string
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(filePath string) *ConfigLoader {
    return &ConfigLoader{
        filePath: filePath,
    }
}

// Load loads configuration from file
func (l *ConfigLoader) Load() (*Config, error) {
    l.mu.Lock()
    defer l.mu.Unlock()

    data, err := os.ReadFile(l.filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config file: %w", err)
    }

    // Apply environment variable overrides
    l.applyEnvOverrides(&cfg)

    // Validate configuration
    if err := l.validate(&cfg); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }

    l.config = &cfg
    return l.config, nil
}

// Get returns the current configuration
func (l *ConfigLoader) Get() *Config {
    l.mu.RLock()
    defer l.mu.RUnlock()
    return l.config
}

// Reload reloads configuration from file
func (l *ConfigLoader) Reload() error {
    return l.Load()
}

// applyEnvOverrides applies environment variable overrides
func (l *ConfigLoader) applyEnvOverrides(cfg *Config) {
    // Database overrides
    if host := os.Getenv("DB_HOST"); host != "" {
        cfg.Database.Host = host
    }
    if port := os.Getenv("DB_PORT"); port != "" {
        fmt.Sscanf(port, "%d", &cfg.Database.Port)
    }
    if user := os.Getenv("DB_USERNAME"); user != "" {
        cfg.Database.Username = user
    }
    if password := os.Getenv("DB_PASSWORD"); password != "" {
        cfg.Database.Password = password
    }

    // Redis overrides
    if host := os.Getenv("REDIS_HOST"); host != "" {
        cfg.Redis.Host = host
    }
    if port := os.Getenv("REDIS_PORT"); port != "" {
        fmt.Sscanf(port, "%d", &cfg.Redis.Port)
    }
    if password := os.Getenv("REDIS_PASSWORD"); password != "" {
        cfg.Redis.Password = password
    }

    // JWT overrides
    if secret := os.Getenv("JWT_SECRET"); secret != "" {
        cfg.Security.JWT.Secret = secret
    }

    // Server overrides
    if port := os.Getenv("SERVER_PORT"); port != "" {
        fmt.Sscanf(port, "%d", &cfg.Server.Port)
    }
}

// validate validates the configuration
func (l *ConfigLoader) validate(cfg *Config) error {
    if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
        return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
    }

    if cfg.Database.Host == "" {
        return fmt.Errorf("database host is required")
    }

    if cfg.Database.Name == "" {
        return fmt.Errorf("database name is required")
    }

    if cfg.Security.JWT.Secret == "" {
        return fmt.Errorf("JWT secret is required")
    }

    if cfg.App.Environment == "production" && cfg.Security.JWT.Secret == "default-secret-change-in-production" {
        return fmt.Errorf("JWT secret must be changed in production")
    }

    return nil
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
    return fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.Username, c.Password, c.Name, c.SSLMode,
    )
}

// GetConnMaxLifetime returns connection max lifetime as duration
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
    return time.Duration(c.ConnMaxLifetime) * time.Second
}

// GetRedisAddr returns the Redis address
func (c *RedisConfig) GetRedisAddr() string {
    return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetReadTimeout returns read timeout as duration
func (c *ServerConfig) GetReadTimeout() time.Duration {
    return time.Duration(c.ReadTimeout) * time.Second
}

// GetWriteTimeout returns write timeout as duration
func (c *ServerConfig) GetWriteTimeout() time.Duration {
    return time.Duration(c.WriteTimeout) * time.Second
}

// GetIdleTimeout returns idle timeout as duration
func (c *ServerConfig) GetIdleTimeout() time.Duration {
    return time.Duration(60) * time.Second
}
