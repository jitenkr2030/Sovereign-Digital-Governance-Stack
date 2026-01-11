package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	App           AppConfig           `yaml:"app"`
	Server        ServerConfig        `yaml:"server"`
	Database      DatabaseConfig      `yaml:"database"`
	Redis         RedisConfig         `yaml:"redis"`
	Kafka         KafkaConfig         `yaml:"kafka"`
	Ingestion     IngestionConfig     `yaml:"ingestion"`
	Analysis      AnalysisConfig      `yaml:"analysis"`
	Alerts        AlertsConfig        `yaml:"alerts"`
	Compliance    ComplianceConfig    `yaml:"compliance"`
	AuditLog      AuditLogConfig      `yaml:"audit_log"`
	Logging       LoggingConfig       `yaml:"logging"`
	Metrics       MetricsConfig       `yaml:"metrics"`
	Security      SecurityConfig      `yaml:"security"`
}

// AppConfig holds application metadata
type AppConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	Version     string `yaml:"version"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	IdleTimeout  int    `yaml:"idle_timeout"`
}

// DatabaseConfig holds PostgreSQL and TimescaleDB settings
type DatabaseConfig struct {
	Host            string               `yaml:"host"`
	Port            int                  `yaml:"port"`
	Username        string               `yaml:"username"`
	Password        string               `yaml:"password"`
	Name            string               `yaml:"name"`
	SSLMode         string               `yaml:"ssl_mode"`
	MaxOpenConns    int                  `yaml:"max_open_conns"`
	MaxIdleConns    int                  `yaml:"max_idle_conns"`
	ConnMaxLifetime int                  `yaml:"conn_max_lifetime"`
	Timescaledb     TimescaleDBConfig    `yaml:"timescaledb"`
}

// TimescaleDBConfig holds TimescaleDB-specific settings
type TimescaleDBConfig struct {
	Enabled            bool   `yaml:"enabled"`
	ChunkInterval      string `yaml:"chunk_interval"`
	RetentionInterval  string `yaml:"retention_interval"`
}

// RedisConfig holds Redis cache settings
type RedisConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Password  string `yaml:"password"`
	DB        int    `yaml:"db"`
	PoolSize  int    `yaml:"pool_size"`
	KeyPrefix string `yaml:"key_prefix"`
}

// KafkaConfig holds Kafka event streaming settings
type KafkaConfig struct {
	Brokers       []string           `yaml:"brokers"`
	TopicPrefix   string             `yaml:"topic_prefix"`
	ConsumerGroup string             `yaml:"consumer_group"`
	Topics        KafkaTopicsConfig  `yaml:"topics"`
}

// KafkaTopicsConfig holds Kafka topic names
type KafkaTopicsConfig struct {
	MarketIngress    string `yaml:"market_ingress"`
	Alerts           string `yaml:"alerts"`
	AnalysisResults  string `yaml:"analysis_results"`
}

// IngestionConfig holds market data ingestion settings
type IngestionConfig struct {
	WebSocket WebSocketConfig `yaml:"websocket"`
	Buffer    BufferConfig    `yaml:"buffer"`
	Auth      AuthConfig      `yaml:"auth"`
}

// WebSocketConfig holds WebSocket connection settings
type WebSocketConfig struct {
	MaxConnections  int   `yaml:"max_connections"`
	PingInterval    int   `yaml:"ping_interval"`
	PongTimeout     int   `yaml:"pong_timeout"`
	MaxMessageSize  int64 `yaml:"max_message_size"`
}

// BufferConfig holds buffer settings for batch processing
type BufferConfig struct {
	MaxSize       int `yaml:"max_size"`
	FlushInterval int `yaml:"flush_interval"`
	BatchSize     int `yaml:"batch_size"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	JWTSecret      string `yaml:"jwt_secret"`
	TokenExpiry    int    `yaml:"token_expiry"`
	RequireMTLS    bool   `yaml:"require_mtls"`
}

// AnalysisConfig holds market analysis settings
type AnalysisConfig struct {
	WashTrade       WashTradeConfig       `yaml:"wash_trade"`
	Spoofing        SpoofingConfig        `yaml:"spoofing"`
	PriceDeviation  PriceDeviationConfig  `yaml:"price_deviation"`
	VolumeAnalysis  VolumeAnalysisConfig  `yaml:"volume_analysis"`
}

// WashTradeConfig holds wash trade detection settings
type WashTradeConfig struct {
	Enabled                      bool    `yaml:"enabled"`
	TimeWindow                   int     `yaml:"time_window"`
	PriceDeviationThreshold      float64 `yaml:"price_deviation_threshold"`
	QuantitySimilarityThreshold  float64 `yaml:"quantity_similarity_threshold"`
	MinTradesForDetection        int     `yaml:"min_trades_for_detection"`
}

// SpoofingConfig holds spoofing detection settings
type SpoofingConfig struct {
	Enabled                bool    `yaml:"enabled"`
	OrderLifetimeWindow    int     `yaml:"order_lifetime_window"`
	LargeOrderThreshold    float64 `yaml:"large_order_threshold"`
	CancellationRateThreshold float64 `yaml:"cancellation_rate_threshold"`
	BaitOrderThreshold     float64 `yaml:"bait_order_threshold"`
}

// PriceDeviationConfig holds price deviation detection settings
type PriceDeviationConfig struct {
	Enabled             bool `yaml:"enabled"`
	GlobalAverageWindow int  `yaml:"global_average_window"`
	DeviationThreshold  float64 `yaml:"deviation_threshold"`
	MinDataPoints       int  `yaml:"min_data_points"`
}

// VolumeAnalysisConfig holds volume analysis settings
type VolumeAnalysisConfig struct {
	Enabled          bool    `yaml:"enabled"`
	AnomalyThreshold float64 `yaml:"anomaly_threshold"`
	RollingWindow    int     `yaml:"rolling_window"`
}

// AlertsConfig holds alert management settings
type AlertsConfig struct {
	SeverityLevels  []SeverityLevelConfig  `yaml:"severity_levels"`
	Notification    NotificationConfig     `yaml:"notification"`
	Escalation      EscalationConfig       `yaml:"escalation"`
}

// SeverityLevelConfig holds severity level settings
type SeverityLevelConfig struct {
	Name         string `yaml:"name"`
	AutoResolve  bool   `yaml:"auto_resolve"`
	RetentionDays int   `yaml:"retention_days"`
}

// NotificationConfig holds notification settings
type NotificationConfig struct {
	Enabled       bool   `yaml:"enabled"`
	EmailEnabled  bool   `yaml:"email_enabled"`
	SMSEnabled    bool   `yaml:"sms_enabled"`
	WebhookEnabled bool  `yaml:"webhook_enabled"`
	WebhookURL    string `yaml:"webhook_url"`
}

// EscalationConfig holds escalation rule settings
type EscalationConfig struct {
	Enabled bool                  `yaml:"enabled"`
	Rules   []EscalationRuleConfig `yaml:"rules"`
}

// EscalationRuleConfig holds individual escalation rule settings
type EscalationRuleConfig struct {
	AfterHours int      `yaml:"after_hours"`
	Notify     []string `yaml:"notify"`
}

// ComplianceConfig holds compliance integration settings
type ComplianceConfig struct {
	Enabled     bool   `yaml:"enabled"`
	APIEndpoint string `yaml:"api_endpoint"`
	APITimeout  int    `yaml:"api_timeout"`
	CacheTTL    int    `yaml:"cache_ttl"`
}

// AuditLogConfig holds audit log integration settings
type AuditLogConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Endpoint      string `yaml:"endpoint"`
	BatchSize     int    `yaml:"batch_size"`
	FlushInterval int    `yaml:"flush_interval"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level          string            `yaml:"level"`
	Format         string            `yaml:"format"`
	Output         string            `yaml:"output"`
	IncludeCaller  bool              `yaml:"include_caller"`
	Fields         map[string]string `yaml:"fields"`
}

// MetricsConfig holds metrics collection settings
type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Endpoint  string `yaml:"endpoint"`
	Port      int    `yaml:"port"`
	Custom    CustomMetricsConfig `yaml:"custom"`
}

// CustomMetricsConfig holds custom metrics settings
type CustomMetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Prefix  string `yaml:"prefix"`
}

// SecurityConfig holds security settings
type SecurityConfig struct {
	TLS       TLSConfig       `yaml:"tls"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	CORS      CORSConfig      `yaml:"cors"`
}

// TLSConfig holds TLS settings
type TLSConfig struct {
	Enabled   bool   `yaml:"enabled"`
	CertFile  string `yaml:"cert_file"`
	KeyFile   string `yaml:"key_file"`
}

// RateLimitConfig holds rate limiting settings
type RateLimitConfig struct {
	Enabled            bool `yaml:"enabled"`
	RequestsPerMinute  int  `yaml:"requests_per_minute"`
	BurstSize          int  `yaml:"burst_size"`
}

// CORSConfig holds CORS settings
type CORSConfig struct {
	Enabled         bool     `yaml:"enabled"`
	AllowedOrigins  []string `yaml:"allowed_origins"`
	AllowedMethods  []string `yaml:"allowed_methods"`
	AllowedHeaders  []string `yaml:"allowed_headers"`
	MaxAge          int      `yaml:"max_age"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides to configuration
func (c *Config) applyEnvOverrides() {
	if host := os.Getenv("DB_HOST"); host != "" {
		c.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		var p int
		if _, err := fmt.Sscanf(port, "%d", &p); err == nil {
			c.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USERNAME"); user != "" {
		c.Database.Username = user
	}
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		c.Database.Password = pass
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		c.Database.Name = name
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		c.Ingestion.Auth.JWTSecret = secret
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		var p int
		if _, err := fmt.Sscanf(port, "%d", &p); err == nil {
			c.Server.Port = p
		}
	}
}

// GetConnMaxLifetime returns the connection max lifetime as a duration
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}

// GetBufferFlushInterval returns the buffer flush interval as a duration
func (c *IngestionConfig) GetBufferFlushInterval() time.Duration {
	return time.Duration(c.Buffer.FlushInterval) * time.Second
}
