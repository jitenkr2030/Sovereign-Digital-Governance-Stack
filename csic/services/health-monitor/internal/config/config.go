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

	// Health Monitor Configuration
	HeartbeatTTL       int  `mapstructure:"heartbeat_ttl"`
	HealthCheckInterval int `mapstructure:"health_check_interval"`
	AlertCooldown      int  `mapstructure:"alert_cooldown"`

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
		Environment:        viper.GetString("environment"),
		ServiceName:        viper.GetString("service_name"),
		LogLevel:           viper.GetString("log_level"),
		HTTPPort:           viper.GetInt("http_port"),
		DatabaseURL:        viper.GetString("database_url"),
		MaxOpenConn:        viper.GetInt("max_open_conn"),
		MaxIdleConn:        viper.GetInt("max_idle_conn"),
		ConnMaxLife:        viper.GetInt("conn_max_lifetime"),
		RedisAddr:          viper.GetString("redis_addr"),
		RedisPassword:      viper.GetString("redis_password"),
		RedisDB:            viper.GetInt("redis_db"),
		KafkaBrokers:       viper.GetString("kafka_brokers"),
		KafkaConsumerGroup: viper.GetString("kafka_consumer_group"),
		HeartbeatTTL:       viper.GetInt("heartbeat_ttl"),
		HealthCheckInterval: viper.GetInt("health_check_interval"),
		AlertCooldown:      viper.GetInt("alert_cooldown"),
		MetricsEnabled:     viper.GetBool("metrics_enabled"),
		MetricsPort:        viper.GetInt("metrics_port"),
		HealthCheckTTL:     viper.GetInt("health_check_ttl"),
		EnableAuth:         viper.GetBool("enable_auth"),
		JWTSecret:          viper.GetString("jwt_secret"),
		AllowedOrigins:     viper.GetString("allowed_origins"),
	}

	return cfg
}

func setDefaults() {
	viper.SetDefault("environment", "development")
	viper.SetDefault("service_name", "health-monitor")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("http_port", 8081)
	viper.SetDefault("max_open_conn", 25)
	viper.SetDefault("max_idle_conn", 5)
	viper.SetDefault("conn_max_lifetime", 300)
	viper.SetDefault("redis_addr", "localhost:6379")
	viper.SetDefault("redis_db", 0)
	viper.SetDefault("kafka_consumer_group", "health-monitor-group")
	viper.SetDefault("heartbeat_ttl", 60)
	viper.SetDefault("health_check_interval", 30)
	viper.SetDefault("alert_cooldown", 300)
	viper.SetDefault("metrics_enabled", true)
	viper.SetDefault("metrics_port", 9091)
	viper.SetDefault("health_check_ttl", 30)
	viper.SetDefault("enable_auth", false)
	viper.SetDefault("allowed_origins", "*")
}

func validateConfig(cfg *Config) error {
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("database_url is required")
	}
	if cfg.HTTPPort <= 0 || cfg.HTTPPort > 65535 {
		return fmt.Errorf("invalid http_port: %d", cfg.HTTPPort)
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
