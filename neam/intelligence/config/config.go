package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the Intelligence Layer
type Config struct {
	General     GeneralConfig     `yaml:"general"`
	Anomaly     AnomalyConfig     `yaml:"anomaly"`
	Inflation   InflationConfig   `yaml:"inflation"`
	Correlation CorrelationConfig `yaml:"correlation"`
	Signals     SignalsConfig     `yaml:"signals"`
	Database    DatabaseConfig    `yaml:"database"`
	Kafka       KafkaConfig       `yaml:"kafka"`
	Logging     LoggingConfig     `yaml:"logging"`
}

// GeneralConfig holds general settings
type GeneralConfig struct {
	Environment       string        `yaml:"environment"`        // development, staging, production
	RegionID          string        `yaml:"region_id"`          // National or regional identifier
	Workers           int           `yaml:"workers"`            // Number of worker threads
	UpdateInterval    time.Duration `yaml:"update_interval"`    // How often to run batch analysis
	HistoryLookback   int           `yaml:"history_lookback"`   // Days of history to maintain
	EnableMetrics     bool          `yaml:"enable_metrics"`     // Enable Prometheus metrics
	MetricsPort       int           `yaml:"metrics_port"`       // Prometheus metrics port
}

// AnomalyConfig holds anomaly detection settings
type AnomalyConfig struct {
	Enabled             bool              `yaml:"enabled"`
	DetectionAlgorithms []string          `yaml:"detection_algorithms"` // z_score, iqr, moving_avg, ensemble
	DefaultThreshold    float64           `yaml:"default_threshold"`
	LookbackPeriods     int               `yaml:"lookback_periods"`
	BusinessRules       []BusinessRuleConfig `yaml:"business_rules"`
}

// BusinessRuleConfig defines domain-specific detection rules
type BusinessRuleConfig struct {
	Name              string   `yaml:"name"`
	Type              string   `yaml:"type"` // black_economy, supply_chain, decoupling
	Enabled           bool     `yaml:"enabled"`
	Threshold         float64  `yaml:"threshold"`
	SeverityWeight    float64  `yaml:"severity_weight"`
	AlertCooldown     int      `yaml:"alert_cooldown"` // Seconds
	AffectedMetrics   []string `yaml:"affected_metrics"`
}

// InflationConfig holds inflation engine settings
type InflationConfig struct {
	Enabled              bool     `yaml:"enabled"`
	PredictionHorizons   []int    `yaml:"prediction_horizons"` // Days: [30, 90, 180]
	ConfidenceThreshold  float64  `yaml:"confidence_threshold`
	LeadingIndicators    []string `yaml:"leading_indicators"` // commodity_prices, producer_prices
	UpdateInterval       int      `yaml:"update_interval"`    // Hours
	BasketWeights        BasketConfig `yaml:"basket_weights"`
}

// BasketConfig defines CPI basket weights
type BasketConfig struct {
	Food       float64 `yaml:"food"`
	Energy     float64 `yaml:"energy"`
	Housing    float64 `yaml:"housing"`
	Transport  float64 `yaml:"transport"`
	Healthcare float64 `yaml:"healthcare"`
	Education  float64 `yaml:"education"`
	Other      float64 `yaml:"other"`
}

// CorrelationConfig holds correlation analysis settings
type CorrelationConfig struct {
	Enabled             bool     `yaml:"enabled"`
	LookbackDays        int      `yaml:"lookback_days"`
	StrongCorrThreshold float64  `yaml:"strong_correlation_threshold"` // e.g., 0.8
	DecouplingThreshold float64  `yaml:"decoupling_threshold"`         // e.g., 0.3
	MetricPairs         []MetricPairConfig `yaml:"metric_pairs"`
}

// MetricPairConfig defines which metrics to correlate
type MetricPairConfig struct {
	MetricA string  `yaml:"metric_a"`
	MetricB string  `yaml:"metric_b"`
	Weight  float64 `yaml:"weight"` // Importance weight for this pair
}

// SignalsConfig holds signal routing settings
type SignalsConfig struct {
	Enabled              bool          `yaml:"enabled"`
	HighSeverityThreshold float64      `yaml:"high_severity_threshold`
	AlertCooldown        time.Duration `yaml:"alert_cooldown`
	BatchSize            int           `yaml:"batch_size`
	BatchInterval        time.Duration `yaml:"batch_interval`
	EnableDeduplication  bool          `yaml:"enable_deduplication
	EnableThrottling     bool          `yaml:"enable_throttling
	OutputTopics         []string      `yaml:"output_topics`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Postgres   PostgresConfig  `yaml:"postgres"`
	ClickHouse ClickHouseConfig `yaml:"clickhouse"`
}

// PostgresConfig holds PostgreSQL settings
type PostgresConfig struct {
	Host               string `yaml:"host"`
	Port               int    `yaml:"port`
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	Database           string `yaml:"database"`
	MaxOpenConns       int    `yaml:"max_open_conns`
	MaxIdleConns       int    `yaml:"max_idle_conns`
	ConnMaxLifetime    int    `yaml:"conn_max_lifetime"` // Minutes
	SSLMode            string `yaml:"ssl_mode`
}

// ClickHouseConfig holds ClickHouse settings
type ClickHouseConfig struct {
	Host               string `yaml:"host"`
	Port               int    `yaml:"port`
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	Database           string `yaml:"database`
	MaxExecutionTime   int    `yaml:"max_execution_time"` // Seconds
	PoolSize           int    `yaml:"pool_size`
}

// KafkaConfig holds Kafka settings
type KafkaConfig struct {
	Brokers            []string    `yaml:"brokers`
	ConsumerGroup      string      `yaml:"consumer_group`
	InputTopics        []string    `yaml:"input_topics`
	OutputTopic        string      `yaml:"output_topic`
	CriticalAlertTopic string      `yaml:"critical_alert_topic`
	SessionTimeout     int         `yaml:"session_timeout"` // Milliseconds
	HeartbeatInterval  int         `yaml:"heartbeat_interval"` // Milliseconds
	AutoOffsetReset    string      `yaml:"auto_offset_reset"` // earliest, latest
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level      string `yaml:"level"` // debug, info, warn, error
	Format     string `yaml:"format"` // json, text
	OutputPath string `yaml:"output_path`
	MaxSize    int    `yaml:"max_size"` // MB
	MaxBackups int    `yaml:"max_backups`
	MaxAge     int    `yaml:"max_age"` // Days
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Apply defaults
	config.applyDefaults()
	
	return &config, nil
}

// applyDefaults sets default values for unset configuration
func (c *Config) applyDefaults() {
	// General defaults
	if c.General.Environment == "" {
		c.General.Environment = "development"
	}
	if c.General.RegionID == "" {
		c.General.RegionID = "NATIONAL"
	}
	if c.General.Workers <= 0 {
		c.General.Workers = 4
	}
	if c.General.UpdateInterval == 0 {
		c.General.UpdateInterval = time.Minute
	}
	if c.General.HistoryLookback <= 0 {
		c.General.HistoryLookback = 365
	}
	if c.General.MetricsPort <= 0 {
		c.General.MetricsPort = 9090
	}
	
	// Anomaly defaults
	if c.Anomaly.DefaultThreshold == 0 {
		c.Anomaly.DefaultThreshold = 3.0
	}
	if c.Anomaly.LookbackPeriods <= 0 {
		c.Anomaly.LookbackPeriods = 30
	}
	
	// Inflation defaults
	if len(c.Inflation.PredictionHorizons) == 0 {
		c.Inflation.PredictionHorizons = []int{30, 90, 180}
	}
	if c.Inflation.ConfidenceThreshold == 0 {
		c.Inflation.ConfidenceThreshold = 0.7
	}
	if c.Inflation.UpdateInterval <= 0 {
		c.Inflation.UpdateInterval = 24
	}
	
	// Default basket weights
	if c.Inflation.BasketWeights.Food == 0 {
		c.Inflation.BasketWeights = BasketConfig{
			Food:       0.25,
			Energy:     0.10,
			Housing:    0.25,
			Transport:  0.15,
			Healthcare: 0.10,
			Education:  0.05,
			Other:      0.10,
		}
	}
	
	// Correlation defaults
	if c.Correlation.LookbackDays <= 0 {
		c.Correlation.LookbackDays = 7
	}
	if c.Correlation.StrongCorrThreshold == 0 {
		c.Correlation.StrongCorrThreshold = 0.8
	}
	if c.Correlation.DecouplingThreshold == 0 {
		c.Correlation.DecouplingThreshold = 0.3
	}
	
	// Default metric pairs
	if len(c.Correlation.MetricPairs) == 0 {
		c.Correlation.MetricPairs = []MetricPairConfig{
			{MetricA: "ENERGY_KWH", MetricB: "PAYMENT_VOL", Weight: 1.0},
			{MetricA: "RETAIL_SALES", MetricB: "EMPLOYMENT", Weight: 0.8},
			{MetricA: "INVENTORY_LEVEL", MetricB: "COMMODITY_PRICE", Weight: 0.9},
		}
	}
	
	// Signals defaults
	if c.Signals.HighSeverityThreshold == 0 {
		c.Signals.HighSeverityThreshold = 0.75
	}
	if c.Signals.AlertCooldown == 0 {
		c.Signals.AlertCooldown = time.Hour
	}
	if c.Signals.BatchSize <= 0 {
		c.Signals.BatchSize = 100
	}
	if c.Signals.BatchInterval == 0 {
		c.Signals.BatchInterval = time.Minute
	}
	
	// Default output topics
	if len(c.Signals.OutputTopics) == 0 {
		c.Signals.OutputTopics = []string{
			"neam.intelligence.signals",
			"neam.alerts.critical",
		}
	}
	
	// Kafka defaults
	if len(c.Kafka.Brokers) == 0 {
		c.Kafka.Brokers = []string{"localhost:9092"}
	}
	if c.Kafka.ConsumerGroup == "" {
		c.Kafka.ConsumerGroup = "neam-intelligence"
	}
	if c.Kafka.InputTopics == nil {
		c.Kafka.InputTopics = []string{
			"neam.payment.events",
			"neam.energy.telemetry",
			"neam.price.data",
		}
	}
	if c.Kafka.OutputTopic == "" {
		c.Kafka.OutputTopic = "neam.intelligence.signals"
	}
	if c.Kafka.SessionTimeout == 0 {
		c.Kafka.SessionTimeout = 30000
	}
	if c.Kafka.HeartbeatInterval == 0 {
		c.Kafka.HeartbeatInterval = 3000
	}
	if c.Kafka.AutoOffsetReset == "" {
		c.Kafka.AutoOffsetReset = "earliest"
	}
	
	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
}

// Save saves configuration to a YAML file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	return os.WriteFile(path, data, 0644)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate basket weights sum to 1.0
	weightSum := c.Inflation.BasketWeights.Food +
		c.Inflation.BasketWeights.Energy +
		c.Inflation.BasketWeights.Housing +
		c.Inflation.BasketWeights.Transport +
		c.Inflation.BasketWeights.Healthcare +
		c.Inflation.BasketWeights.Education +
		c.Inflation.BasketWeights.Other
	
	if weightSum < 0.99 || weightSum > 1.01 {
		return fmt.Errorf("basket weights must sum to 1.0, got %f", weightSum)
	}
	
	// Validate Kafka brokers
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("at least one Kafka broker must be configured")
	}
	
	// Validate thresholds
	if c.Anomaly.DefaultThreshold <= 0 {
		return fmt.Errorf("anomaly threshold must be positive")
	}
	
	if c.Correlation.StrongCorrThreshold > 1 || c.Correlation.StrongCorrThreshold < -1 {
		return fmt.Errorf("correlation threshold must be between -1 and 1")
	}
	
	return nil
}

// GetDSN returns PostgreSQL connection string
func (c *PostgresConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode,
	)
}

// GetClickHouseDSN returns ClickHouse connection string
func (c *ClickHouseConfig) GetDSN() string {
	return fmt.Sprintf(
		"clickhouse://%s:%s@%s:%d/%s",
		c.Username, c.Password, c.Host, c.Port, c.Database,
	)
}

// GetBusinessRules converts config to detection rules
func (c *AnomalyConfig) GetBusinessRules() []BusinessRuleConfig {
	return c.BusinessRules
}
