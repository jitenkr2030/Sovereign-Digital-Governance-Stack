package config

import (
	"fmt"
	"os"
	"time"
)

// Config represents the complete application configuration
type Config struct {
	App      AppConfig      `yaml:"app"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	HSM      HSMConfig      `yaml:"hsm"`
	Governance GovernanceConfig `yaml:"governance"`
	Signing  SigningConfig  `yaml:"signing"`
	Logging  LoggingConfig  `yaml:"logging"`
	Metrics  MetricsConfig  `yaml:"metrics"`
	Security SecurityConfig `yaml:"security"`
}

// AppConfig contains application metadata
type AppConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	Version     string `yaml:"version"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	IdleTimeout  int    `yaml:"idle_timeout"`
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
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Password  string `yaml:"password"`
	DB        int    `yaml:"db"`
	PoolSize  int    `yaml:"pool_size"`
	KeyPrefix string `yaml:"key_prefix"`
}

// KafkaConfig contains Kafka connection settings
type KafkaConfig struct {
	Brokers       []string          `yaml:"brokers"`
	TopicPrefix   string            `yaml:"topic_prefix"`
	ConsumerGroup string            `yaml:"consumer_group"`
	Topics        KafkaTopicsConfig `yaml:"topics"`
}

// KafkaTopicsConfig contains Kafka topic names
type KafkaTopicsConfig struct {
	Transactions string `yaml:"transactions"`
	Freezes      string `yaml:"freezes"`
	Signatures   string `yaml:"signatures"`
	Alerts       string `yaml:"alerts"`
}

// HSMConfig contains HSM connection settings
type HSMConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Provider    string `yaml:"provider"` // "pkcs11", "soft", "cloud"
	Endpoint    string `yaml:"endpoint"`
	SlotID      int    `yaml:"slot_id"`
	Pin         string `yaml:"pin"`
	KeyLabel    string `yaml:"key_label"`
	Timeout     int    `yaml:"timeout"`
}

// GovernanceConfig contains multi-signature governance settings
type GovernanceConfig struct {
	MinSigners              int `yaml:"min_signers"`
	DefaultThreshold        int `yaml:"default_threshold"`
	TransactionExpiryHours  int `yaml:"transaction_expiry_hours"`
	MaxPendingTransactions  int `yaml:"max_pending_transactions"`
	EmergencySignerRequired bool `yaml:"emergency_signer_required"`
	EmergencyThreshold      int `yaml:"emergency_threshold"`
}

// SigningConfig contains transaction signing settings
type SigningConfig struct {
	TimeoutSeconds     int  `yaml:"timeout_seconds"`
	RetryAttempts      int  `yaml:"retry_attempts"`
	RequireMultiSig    bool `yaml:"require_multi_sig"`
	AllowPartialSign   bool `yaml:"allow_partial_sign"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level         string            `yaml:"level"`
	Format        string            `yaml:"format"`
	Output        string            `yaml:"output"`
	IncludeCaller bool              `yaml:"include_caller"`
	Fields        map[string]string `yaml:"fields"`
}

// MetricsConfig contains metrics settings
type MetricsConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	Port     int    `yaml:"port"`
	Prefix   string `yaml:"prefix"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	TLS       TLSServerConfig   `yaml:"tls"`
	RateLimit RateLimitConfig   `yaml:"rate_limit"`
	CORS      CORSConfig        `yaml:"cors"`
}

// TLSServerConfig contains TLS settings
type TLSServerConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// RateLimitConfig contains rate limit settings
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
	BurstSize         int  `yaml:"burst_size"`
}

// CORSConfig contains CORS settings
type CORSConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
	MaxAge         int      `yaml:"max_age"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(&cfg)

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides to the configuration
func applyEnvOverrides(cfg *Config) {
	// Database overrides
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil {
			cfg.Database.Port = port
		}
	}
	if v := os.Getenv("DB_USERNAME"); v != "" {
		cfg.Database.Username = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.Name = v
	}

	// Redis overrides
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil {
			cfg.Redis.Port = port
		}
	}

	// HSM overrides
	if v := os.Getenv("HSM_PIN"); v != "" {
		cfg.HSM.Pin = v
	}

	// Server overrides
	if v := os.Getenv("APP_PORT"); v != "" {
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil {
			cfg.Server.Port = port
		}
	}
}

// GetConnMaxLifetime returns the database connection max lifetime as a duration
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}
