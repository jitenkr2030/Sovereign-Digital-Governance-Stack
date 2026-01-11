package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the Policy Layer
type Config struct {
	General     GeneralConfig     `yaml:"general"`
	Policy      PolicyConfig      `yaml:"policy"`
	Workflow    WorkflowConfig    `yaml:"workflow"`
	State       StateConfig       `yaml:"state"`
	Alerts      AlertsConfig      `yaml:"alerts"`
	Simulation  SimulationConfig  `yaml:"simulation`
	Database    DatabaseConfig    `yaml:"database`
	Redis       RedisConfig       `yaml:"redis`
	Kafka       KafkaConfig       `yaml:"kafka`
	Auth        AuthConfig        `yaml:"auth`
	Logging     LoggingConfig     `yaml:"logging`
}

// GeneralConfig holds general settings
type GeneralConfig struct {
	Environment    string        `yaml:"environment"`
	ServiceName    string        `yaml:"service_name"`
	ServiceVersion string        `yaml:"service_version`
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port`
	MetricsPort    int           `yaml:"metrics_port`
	ReadTimeout    time.Duration `yaml:"read_timeout`
	WriteTimeout   time.Duration `yaml:"write_timeout`
	IdleTimeout    time.Duration `yaml:"idle_timeout`
}

// PolicyConfig holds policy engine settings
type PolicyConfig struct {
	Enabled            bool              `yaml:"enabled"`
	EvaluationInterval time.Duration     `yaml:"evaluation_interval`
	MaxPolicies        int               `yaml:"max_policies`
	CacheTTL           time.Duration     `yaml:"cache_ttl`
	DefaultCooldown    int               `yaml:"default_cooldown_minutes`
	RulesPath          string            `yaml:"rules_path`
	PolicyCategories   []string          `yaml:"policy_categories`
}

// WorkflowConfig holds workflow settings
type WorkflowConfig struct {
	Enabled            bool          `yaml:"enabled"`
	DefaultTimeout     time.Duration `yaml:"default_timeout`
	MaxRetries         int           `yaml:"max_retries`
	ApprovalRequired   bool          `yaml:"approval_required`
	ApprovalTimeout    time.Duration `yaml:"approval_timeout`
	RegistryPath       string        `yaml:"registry_path`
}

// StateConfig holds state management settings
type StateConfig struct {
	RedisHost     string        `yaml:"redis_host`
	RedisPort     int           `yaml:"redis_port`
	RedisDB       int           `yaml:"redis_db`
	RedisPassword string        `yaml:"redis_password`
	PoolSize      int           `yaml:"pool_size`
	MetricTTL     time.Duration `yaml:"metric_ttl`
	StateTTL      time.Duration `yaml:"state_ttl`
}

// AlertsConfig holds alert settings
type AlertsConfig struct {
	Enabled         bool          `yaml:"enabled"`
	MaxRetries      int           `yaml:"max_retries`
	RetryDelay      time.Duration `yaml:"retry_delay`
	BatchSize       int           `yaml:"batch_size`
	BatchInterval   time.Duration `yaml:"batch_interval`
	RateLimit       int           `yaml:"rate_limit`
	EmailEnabled    bool          `yaml:"email_enabled`
	SMSEnabled      bool          `yaml:"sms_enabled`
	WebhookEnabled  bool          `yaml:"webhook_enabled`
	DashboardEnabled bool         `yaml:"dashboard_enabled`
}

// SimulationConfig holds simulation settings
type SimulationConfig struct {
	Enabled        bool          `yaml:"enabled"`
	MaxConcurrent  int           `yaml:"max_concurrent`
	Timeout        time.Duration `yaml:"timeout`
	ResultTTL      time.Duration `yaml:"result_ttl`
	CacheSize      int           `yaml:"cache_size`
	PythonEndpoint string        `yaml:"python_endpoint`
}

// DatabaseConfig holds database settings
type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port`
	Username     string `yaml:"username`
	Password     string `yaml:"password`
	Database     string `yaml:"database`
	MaxOpenConns int    `yaml:"max_open_conns`
	MaxIdleConns int    `yaml:"max_idle_conns`
	SSLMode      string `yaml:"ssl_mode`
}

// RedisConfig holds Redis settings
type RedisConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port`
	Password     string `yaml:"password`
	DB           int    `yaml:"db`
	PoolSize     int    `yaml:"pool_size`
	KeyPrefix    string `yaml:"key_prefix`
}

// KafkaConfig holds Kafka settings
type KafkaConfig struct {
	Brokers       []string `yaml:"brokers`
	ConsumerGroup string   `yaml:"consumer_group`
	InputTopic    string   `yaml:"input_topic`
	OutputTopic   string   `yaml:"output_topic`
	AlertTopic    string   `yaml:"alert_topic`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	Enabled       bool   `yaml:"enabled`
	JWTSecret     string `yaml:"jwt_secret`
	TokenExpiry   int    `yaml:"token_expiry_hours`
	OAuthProvider string `yaml:"oauth_provider`
	OAuthClientID string `yaml:"oauth_client_id`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	OutputPath string `yaml:"output_path"`
	MaxSize    int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age_days"`
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	config.applyDefaults()
	
	return &config, nil
}

// applyDefaults sets default values
func (c *Config) applyDefaults() {
	// General defaults
	if c.General.Environment == "" {
		c.General.Environment = "development"
	}
	if c.General.ServiceName == "" {
		c.General.ServiceName = "neam-policy"
	}
	if c.General.Host == "" {
		c.General.Host = "0.0.0.0"
	}
	if c.General.Port == 0 {
		c.General.Port = 8080
	}
	if c.General.MetricsPort == 0 {
		c.General.MetricsPort = 9090
	}
	if c.General.ReadTimeout == 0 {
		c.General.ReadTimeout = 30 * time.Second
	}
	if c.General.WriteTimeout == 0 {
		c.General.WriteTimeout = 30 * time.Second
	}
	
	// Policy defaults
	if c.Policy.EvaluationInterval == 0 {
		c.Policy.EvaluationInterval = time.Minute
	}
	if c.Policy.MaxPolicies == 0 {
		c.Policy.MaxPolicies = 1000
	}
	if c.Policy.CacheTTL == 0 {
		c.Policy.CacheTTL = 5 * time.Minute
	}
	if c.Policy.DefaultCooldown == 0 {
		c.Policy.DefaultCooldown = 60
	}
	
	// Workflow defaults
	if c.Workflow.DefaultTimeout == 0 {
		c.Workflow.DefaultTimeout = 30 * time.Minute
	}
	if c.Workflow.MaxRetries == 0 {
		c.Workflow.MaxRetries = 3
	}
	
	// State defaults
	if c.State.RedisHost == "" {
		c.State.RedisHost = "localhost"
	}
	if c.State.RedisPort == 0 {
		c.State.RedisPort = 6379
	}
	if c.State.RedisDB == 0 {
		c.State.RedisDB = 0
	}
	if c.State.PoolSize == 0 {
		c.State.PoolSize = 10
	}
	if c.State.MetricTTL == 0 {
		c.State.MetricTTL = 24 * time.Hour
	}
	if c.State.StateTTL == 0 {
		c.State.StateTTL = 7 * 24 * time.Hour
	}
	
	// Alerts defaults
	if c.Alerts.MaxRetries == 0 {
		c.Alerts.MaxRetries = 3
	}
	if c.Alerts.RetryDelay == 0 {
		c.Alerts.RetryDelay = 5 * time.Second
	}
	if c.Alerts.BatchSize == 0 {
		c.Alerts.BatchSize = 100
	}
	if c.Alerts.BatchInterval == 0 {
		c.Alerts.BatchInterval = time.Minute
	}
	if c.Alerts.RateLimit == 0 {
		c.Alerts.RateLimit = 1000
	}
	
	// Simulation defaults
	if c.Simulation.MaxConcurrent == 0 {
		c.Simulation.MaxConcurrent = 10
	}
	if c.Simulation.Timeout == 0 {
		c.Simulation.Timeout = 10 * time.Minute
	}
	if c.Simulation.ResultTTL == 0 {
		c.Simulation.ResultTTL = 7 * 24 * time.Hour
	}
	if c.Simulation.CacheSize == 0 {
		c.Simulation.CacheSize = 100
	}
	if c.Simulation.PythonEndpoint == "" {
		c.Simulation.PythonEndpoint = "http://localhost:8001"
	}
	
	// Database defaults
	if c.Database.Host == "" {
		c.Database.Host = "localhost"
	}
	if c.Database.Port == 0 {
		c.Database.Port = 5432
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 25
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 5
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	
	// Redis defaults
	if c.Redis.Host == "" {
		c.Redis.Host = "localhost"
	}
	if c.Redis.Port == 0 {
		c.Redis.Port = 6379
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 10
	}
	if c.Redis.KeyPrefix == "" {
		c.Redis.KeyPrefix = "neam:"
	}
	
	// Kafka defaults
	if len(c.Kafka.Brokers) == 0 {
		c.Kafka.Brokers = []string{"localhost:9092"}
	}
	if c.Kafka.ConsumerGroup == "" {
		c.Kafka.ConsumerGroup = "neam-policy"
	}
	if c.Kafka.InputTopic == "" {
		c.Kafka.InputTopic = "neam.intelligence.signals"
	}
	if c.Kafka.OutputTopic == "" {
		c.Kafka.OutputTopic = "neam.policy.interventions"
	}
	if c.Kafka.AlertTopic == "" {
		c.Kafka.AlertTopic = "neam.alerts"
	}
	
	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Logging.MaxSize == 0 {
		c.Logging.MaxSize = 100
	}
	if c.Logging.MaxBackups == 0 {
		c.Logging.MaxBackups = 5
	}
	if c.Logging.MaxAge == 0 {
		c.Logging.MaxAge = 30
	}
}

// Save saves configuration to a file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate ports
	if c.General.Port < 1 || c.General.Port > 65535 {
		return newValidationError("general.port", "must be between 1 and 65535")
	}
	
	if c.General.MetricsPort < 1 || c.General.MetricsPort > 65535 {
		return newValidationError("general.metrics_port", "must be between 1 and 65535")
	}
	
	// Validate timeouts
	if c.General.ReadTimeout < 0 {
		return newValidationError("general.read_timeout", "must be non-negative")
	}
	
	if c.Policy.DefaultCooldown < 0 {
		return newValidationError("policy.default_cooldown_minutes", "must be non-negative")
	}
	
	return nil
}

// GetDSN returns PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return "host=" + c.Host +
		" port=" + itoa(c.Port) +
		" user=" + c.Username +
		" password=" + c.Password +
		" dbname=" + c.Database +
		" sslmode=" + c.SSLMode
}

// GetRedisAddr returns Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return c.Host + ":" + itoa(c.Port)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func newValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func itoa(i int) string {
	return string(rune('0'+i/1000%10)) + string(rune('0'+i/100%10)) + string(rune('0'+i/10%10)) + string(rune('0'+i%10))
}
