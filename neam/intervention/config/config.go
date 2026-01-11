package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Temporal TemporalConfig `yaml:"temporal"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	Security SecurityConfig `yaml:"security"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port     int    `yaml:"port"`
	Mode     string `yaml:"mode"` // debug, release, test
	LogLevel string `yaml:"log_level"`
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
	MaxConns int    `yaml:"max_conns"`
}

// ConnectionString returns the PostgreSQL connection string
func (d *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// Address returns the Redis address
func (r *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// TemporalConfig holds Temporal workflow engine configuration
type TemporalConfig struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	Namespace     string `yaml:"namespace"`
	TaskQueue     string `yaml:"task_queue"`
	MaxConcurrent int    `yaml:"max_concurrent"`
}

// Address returns the Temporal address
func (t *TemporalConfig) Address() string {
	return fmt.Sprintf("%s:%d", t.Host, t.Port)
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers   []string `yaml:"brokers"`
	Topic     string   `yaml:"topic"`
	GroupID   string   `yaml:"group_id"`
	TLSEnabled bool    `yaml:"tls_enabled"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret      string `yaml:"jwt_secret"`
	EncryptionKey  string `yaml:"encryption_key"`
	ApprovalExpiry int    `yaml:"approval_expiry_hours"`
}

// LoadConfig loads configuration from the specified file path
func LoadConfig(configPath ...string) (*Config, error) {
	path := "config.yaml"
	if len(configPath) > 0 {
		path = configPath[0]
	}

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

	// Set defaults
	setDefaults(&cfg)

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides to configuration
func applyEnvOverrides(cfg *Config) {
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Database.Port)
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		cfg.Database.Name = name
	}
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		cfg.Redis.Host = redisHost
	}
	if temporalHost := os.Getenv("TEMPORAL_HOST"); temporalHost != "" {
		cfg.Temporal.Host = temporalHost
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.Security.JWTSecret = jwtSecret
	}
}

// setDefaults sets default values for configuration
func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "release"
	}
	if cfg.Server.LogLevel == "" {
		cfg.Server.LogLevel = "info"
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "require"
	}
	if cfg.Database.MaxConns == 0 {
		cfg.Database.MaxConns = 25
	}
	if cfg.Redis.PoolSize == 0 {
		cfg.Redis.PoolSize = 100
	}
	if cfg.Temporal.TaskQueue == "" {
		cfg.Temporal.TaskQueue = "intervention-task-queue"
	}
	if cfg.Temporal.Namespace == "" {
		cfg.Temporal.Namespace = "neam-production"
	}
	if cfg.Temporal.MaxConcurrent == 0 {
		cfg.Temporal.MaxConcurrent = 100
	}
	if cfg.Security.ApprovalExpiry == 0 {
		cfg.Security.ApprovalExpiry = 72
	}
}
