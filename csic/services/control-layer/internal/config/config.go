package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	// Application
	Environment string `mapstructure:"environment"`
	ServiceName string `mapstructure:"service_name"`
	LogLevel    string `mapstructure:"log_level"`

	// HTTP Server
	HTTPPort int `mapstructure:"http_port"`

	// gRPC Server
	GRPCPort int `mapstructure:"grpc_port"`

	// Database
	DatabaseURL     string `mapstructure:"database_url"`
	MaxOpenConn     int    `mapstructure:"max_open_conn"`
	MaxIdleConn     int    `mapstructure:"max_idle_conn"`
	ConnMaxLife     int    `mapstructure:"conn_max_lifetime"`

	// Redis
	RedisAddr     string `mapstructure:"redis_addr"`
	RedisPassword string `mapstructure:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db"`

	// Kafka
	KafkaBrokers       string `mapstructure:"kafka_brokers"`
	KafkaConsumerGroup string `mapstructure:"kafka_consumer_group"`

	// Policy Engine
	PolicyCacheTTL     int  `mapstructure:"policy_cache_ttl"`
	PolicyHotReload    bool `mapstructure:"policy_hot_reload"`
	EvaluationTimeout  int  `mapstructure:"evaluation_timeout_ms"`

	// Enforcement
	EnforcementRetryAttempts int `mapstructure:"enforcement_retry_attempts"`
	EnforcementRetryDelay    int `mapstructure:"enforcement_retry_delay_ms"`

	// Monitoring
	MetricsEnabled bool   `mapstructure:"metrics_enabled"`
	MetricsPort    int    `mapstructure:"metrics_port"`
	HealthCheckTTL int    `mapstructure:"health_check_ttl"`

	// Security
	EnableAuth     bool   `mapstructure:"enable_auth"`
	JWTSecret      string `mapstructure:"jwt_secret"`
	AllowedOrigins string `mapstructure:"allowed_origins"`
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Allow environment variables to override config
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults and env vars
			return loadFromDefaults(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func loadFromDefaults() *Config {
	setDefaults()

	cfg := &Config{
		Environment:         viper.GetString("environment"),
		ServiceName:         viper.GetString("service_name"),
		LogLevel:            viper.GetString("log_level"),
		HTTPPort:            viper.GetInt("http_port"),
		GRPCPort:            viper.GetInt("grpc_port"),
		DatabaseURL:         viper.GetString("database_url"),
		MaxOpenConn:         viper.GetInt("max_open_conn"),
		MaxIdleConn:         viper.GetInt("max_idle_conn"),
		ConnMaxLife:         viper.GetInt("conn_max_lifetime"),
		RedisAddr:           viper.GetString("redis_addr"),
		RedisPassword:       viper.GetString("redis_password"),
		RedisDB:             viper.GetInt("redis_db"),
		KafkaBrokers:        viper.GetString("kafka_brokers"),
		KafkaConsumerGroup:  viper.GetString("kafka_consumer_group"),
		PolicyCacheTTL:      viper.GetInt("policy_cache_ttl"),
		PolicyHotReload:     viper.GetBool("policy_hot_reload"),
		EvaluationTimeout:   viper.GetInt("evaluation_timeout_ms"),
		EnforcementRetryAttempts: viper.GetInt("enforcement_retry_attempts"),
		EnforcementRetryDelay:    viper.GetInt("enforcement_retry_delay_ms"),
		MetricsEnabled:      viper.GetBool("metrics_enabled"),
		MetricsPort:         viper.GetInt("metrics_port"),
		HealthCheckTTL:      viper.GetInt("health_check_ttl"),
		EnableAuth:          viper.GetBool("enable_auth"),
		JWTSecret:           viper.GetString("jwt_secret"),
		AllowedOrigins:      viper.GetString("allowed_origins"),
	}

	return cfg
}

func setDefaults() {
	viper.SetDefault("environment", "development")
	viper.SetDefault("service_name", "control-layer")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("http_port", 8080)
	viper.SetDefault("grpc_port", 9090)
	viper.SetDefault("max_open_conn", 25)
	viper.SetDefault("max_idle_conn", 5)
	viper.SetDefault("conn_max_lifetime", 300)
	viper.SetDefault("redis_addr", "localhost:6379")
	viper.SetDefault("redis_db", 0)
	viper.SetDefault("kafka_consumer_group", "control-layer-group")
	viper.SetDefault("policy_cache_ttl", 300)
	viper.SetDefault("policy_hot_reload", true)
	viper.SetDefault("evaluation_timeout_ms", 100)
	viper.SetDefault("enforcement_retry_attempts", 3)
	viper.SetDefault("enforcement_retry_delay_ms", 1000)
	viper.SetDefault("metrics_enabled", true)
	viper.SetDefault("metrics_port", 9090)
	viper.SetDefault("health_check_ttl", 30)
	viper.SetDefault("enable_auth", false)
	viper.SetDefault("allowed_origins", "*")
}

func validateConfig(cfg *Config) error {
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("database_url is required")
	}
	if cfg.KafkaBrokers == "" {
		return fmt.Errorf("kafka_brokers is required")
	}
	if cfg.HTTPPort <= 0 || cfg.HTTPPort > 65535 {
		return fmt.Errorf("invalid http_port: %d", cfg.HTTPPort)
	}
	if cfg.GRPCPort <= 0 || cfg.GRPCPort > 65535 {
		return fmt.Errorf("invalid grpc_port: %d", cfg.GRPCPort)
	}
	return nil
}

// GetAllowedOrigins returns a slice of allowed origins
func (c *Config) GetAllowedOrigins() []string {
	if c.AllowedOrigins == "" || c.AllowedOrigins == "*" {
		return []string{"*"}
	}
	return strings.Split(c.AllowedOrigins, ",")
}

// GetRedisAddr returns the Redis address
func (c *Config) GetRedisAddr() string {
	return c.RedisAddr
}

// GetRedisPassword returns the Redis password
func (c *Config) GetRedisPassword() string {
	return c.RedisPassword
}

// GetRedisDB returns the Redis database number
func (c *Config) GetRedisDB() int {
	return c.RedisDB
}
