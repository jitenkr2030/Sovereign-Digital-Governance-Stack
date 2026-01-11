package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete transport adapter configuration
type Config struct {
	Service    ServiceConfig     `yaml:"service"`
	MQTT       MQTTConfig        `yaml:"mqtt"`
	Kafka      KafkaConfig       `yaml:"kafka"`
	Session    SessionConfig     `yaml:"session"`
	Filtering  FilteringConfig   `yaml:"filtering"`
	Logging    LoggingConfig     `yaml:"logging"`
	Monitoring MonitoringConfig  `yaml:"monitoring"`
	Resilience ResilienceConfig  `yaml:"resilience"`
}

// ServiceConfig contains service-level configuration
type ServiceConfig struct {
	Name        string        `yaml:"name"`
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	Environment string        `yaml:"environment"`
	Shutdown    time.Duration `yaml:"shutdown_timeout"`
}

// MQTTConfig contains MQTT protocol configuration
type MQTTConfig struct {
	Enabled          bool                   `yaml:"enabled"`
	Sources          []MQTTEndPoint         `yaml:"sources"`
	QoSByCriticality map[string]int         `yaml:"qos_by_criticality"`
	KeepAlive        Duration               `yaml:"keep_alive"`
	SessionExpiry    Duration               `yaml:"session_expiry"`
	Timeout          Duration               `yaml:"timeout"`
	RetryInterval    Duration               `yaml:"retry_interval"`
	ReconnectWait    Duration               `yaml:"reconnect_wait"`
}

// MQTTEndPoint represents an MQTT data source
type MQTTEndPoint struct {
	BrokerURL  string   `yaml:"broker_url"`
	ClientID   string   `yaml:"client_id"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	CleanStart bool     `yaml:"clean_start"`
	Topics     []string `yaml:"topics"`
	QoS        int      `yaml:"qos"`
	Criticality string  `yaml:"criticality"`
}

// KafkaConfig contains Kafka output configuration
type KafkaConfig struct {
	Brokers       string   `yaml:"brokers"`
	OutputTopic   string   `yaml:"output_topic"`
	DLQTopic      string   `yaml:"dlq_topic"`
	Version       string   `yaml:"version"`
	MaxOpenReqs   int      `yaml:"max_open_requests"`
	DialTimeout   Duration `yaml:"dial_timeout"`
	WriteTimeout  Duration `yaml:"write_timeout"`
}

// SessionConfig contains session management configuration
type SessionConfig struct {
	MaxRetries          int      `yaml:"max_retries"`
	InitialDelay        Duration `yaml:"initial_delay"`
	MaxDelay            Duration `yaml:"max_delay"`
	Multiplier          float64  `yaml:"multiplier"`
	HealthCheckInterval Duration `yaml:"health_check_interval"`
	ReconnectWait       Duration `yaml:"reconnect_wait"`
	StateStoreType      string   `yaml:"state_store_type"` // redis, memory
	StateStoreURL       string   `yaml:"state_store_url"`
}

// FilteringConfig contains signal processing configuration
type FilteringConfig struct {
	DeadbandPercent      float64   `yaml:"deadband_percent"`
	SpikeThreshold       float64   `yaml:"spike_threshold"`
	SpikeWindowSize      int       `yaml:"spike_window_size"`
	RateOfChangeLimit    float64   `yaml:"rate_of_change_limit"`
	QualityFilter        []string  `yaml:"quality_filter"`
	MinReportingInterval Duration  `yaml:"min_reporting_interval"`
	MaxValue             float64   `yaml:"max_value"`
	MinValue             float64   `yaml:"min_value"`
	OutlierThreshold     float64   `yaml:"outlier_threshold"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	OutputPath string `yaml:"output_path"`
	JSONFormat bool   `yaml:"json_format"`
}

// MonitoringConfig contains monitoring configuration
type MonitoringConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Port       int    `yaml:"port"`
	Path       string `yaml:"path"`
	HealthPath string `yaml:"health_path"`
}

// ResilienceConfig contains resilience and error handling configuration
type ResilienceConfig struct {
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	DLQ            DLQConfig            `yaml:"dlq"`
	Retry          RetryConfig          `yaml:"retry"`
	Fallbacks      FallbackConfig       `yaml:"fallbacks"`
}

// CircuitBreakerConfig contains circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled              bool    `yaml:"enabled"`
	Timeout              Duration `yaml:"timeout"`
	MaxRequests          uint32  `yaml:"max_requests"`
	Interval             Duration `yaml:"interval"`
	ReadyToTripThreshold float64 `yaml:"ready_to_trip_threshold"`
	SuccessThreshold     float64 `yaml:"success_threshold"`
	RequestTimeout       Duration `yaml:"request_timeout"`
}

// DLQConfig contains dead letter queue configuration
type DLQConfig struct {
	Enabled                bool      `yaml:"enabled"`
	TopicSuffix            string    `yaml:"topic_suffix"`
	RetryCount             int       `yaml:"retry_count"`
	RetryDelay             Duration  `yaml:"retry_delay"`
	MaxMessageSize         int       `yaml:"max_message_size"`
	EnableSchemaValidation bool      `yaml:"enable_schema_validation"`
}

// RetryConfig contains retry policy configuration
type RetryConfig struct {
	MaxAttempts    int      `yaml:"max_attempts"`
	InitialDelay   Duration `yaml:"initial_delay"`
	MaxDelay       Duration `yaml:"max_delay"`
	Multiplier     float64  `yaml:"multiplier"`
	Jitter         float64  `yaml:"jitter"`
	RetryableCodes []string `yaml:"retryable_codes"`
}

// FallbackConfig contains fallback behavior configuration
type FallbackConfig struct {
	Enabled              bool    `yaml:"enabled"`
	CacheMaxAge          Duration `yaml:"cache_max_age"`
	DegradedModeEnabled  bool    `yaml:"degraded_mode_enabled"`
	DegradedMaxLatency   Duration `yaml:"degraded_max_latency"`
}

// Duration is a custom type for YAML unmarshaling of time.Duration
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}

	duration, err := time.ParseDuration(str)
	if err != nil {
		return fmt.Errorf("invalid duration format: %s", str)
	}

	*d = Duration(duration)
	return nil
}

// MarshalYAML implements yaml.Marshaler
func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// Load reads configuration from the specified file path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides to configuration
func (c *Config) applyEnvOverrides() {
	if host := os.Getenv("SERVICE_HOST"); host != "" {
		c.Service.Host = host
	}
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		c.Kafka.Brokers = brokers
	}
	if brokerURL := os.Getenv("MQTT_BROKER_URL"); brokerURL != "" {
		if len(c.MQTT.Sources) > 0 {
			c.MQTT.Sources[0].BrokerURL = brokerURL
		}
	}
	if clientID := os.Getenv("MQTT_CLIENT_ID"); clientID != "" {
		if len(c.MQTT.Sources) > 0 {
			c.MQTT.Sources[0].ClientID = clientID
		}
	}
	if username := os.Getenv("MQTT_USERNAME"); username != "" {
		if len(c.MQTT.Sources) > 0 {
			c.MQTT.Sources[0].Username = username
		}
	}
	if password := os.Getenv("MQTT_PASSWORD"); password != "" {
		if len(c.MQTT.Sources) > 0 {
			c.MQTT.Sources[0].Password = password
		}
	}
	if stateStoreURL := os.Getenv("STATE_STORE_URL"); stateStoreURL != "" {
		c.Session.StateStoreURL = stateStoreURL
	}
}

// Validate checks that the configuration is valid
func (c *Config) Validate() error {
	if c.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if !c.MQTT.Enabled {
		return fmt.Errorf("MQTT protocol must be enabled")
	}
	if len(c.MQTT.Sources) == 0 {
		return fmt.Errorf("at least one MQTT source is required")
	}
	if c.Kafka.Brokers == "" {
		return fmt.Errorf("Kafka brokers is required")
	}
	if c.Kafka.OutputTopic == "" {
		return fmt.Errorf("Kafka output topic is required")
	}
	if c.Session.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative")
	}
	return nil
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Service: ServiceConfig{
			Name:        "transport-adapter",
			Host:        "0.0.0.0",
			Port:        8086,
			Environment: "development",
			Shutdown:    30 * time.Second,
		},
		MQTT: MQTTConfig{
			Enabled: true,
			Sources: []MQTTEndPoint{
				{
					BrokerURL:  "mqtt://localhost:1883",
					ClientID:   "neam-transport-adapter",
					CleanStart: false,
					Topics:     []string{"transport/+/telemetry", "transport/+/status"},
					QoS:        1,
					Criticality: "standard",
				},
			},
			QoSByCriticality: map[string]int{
				"critical":   2, // Exactly-once delivery
				"important":  1, // At-least-once delivery
				"standard":   1, // At-least-once delivery
				"background": 0, // At-most-once delivery
			},
			KeepAlive:     Duration(30 * time.Second),
			SessionExpiry: Duration(3600 * time.Second), // 1 hour session expiry
			Timeout:       Duration(10 * time.Second),
			RetryInterval: Duration(5 * time.Second),
			ReconnectWait: Duration(1 * time.Second),
		},
		Kafka: KafkaConfig{
			Brokers:      "localhost:9092",
			OutputTopic:  "transport.telemetry",
			DLQTopic:     "transport.dlq",
			Version:      "2.8.0",
			MaxOpenReqs:  100,
			DialTimeout:  Duration(10 * time.Second),
			WriteTimeout: Duration(30 * time.Second),
		},
		Session: SessionConfig{
			MaxRetries:          10,
			InitialDelay:        Duration(1 * time.Second),
			MaxDelay:            Duration(60 * time.Second),
			Multiplier:          2.0,
			HealthCheckInterval: Duration(30 * time.Second),
			ReconnectWait:       Duration(5 * time.Second),
			StateStoreType:      "redis",
			StateStoreURL:       "redis://localhost:6379",
		},
		Filtering: FilteringConfig{
			DeadbandPercent:      0.1,
			SpikeThreshold:       20.0,
			SpikeWindowSize:      5,
			RateOfChangeLimit:    10.0,
			QualityFilter:        []string{"Good"},
			MinReportingInterval: Duration(100 * time.Millisecond),
			MaxValue:             1e9,
			MinValue:             -1e9,
			OutlierThreshold:     3.0,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "/var/log/neam",
			JSONFormat: true,
		},
		Monitoring: MonitoringConfig{
			Enabled:    true,
			Port:       9092,
			Path:       "/metrics",
			HealthPath: "/health",
		},
		Resilience: ResilienceConfig{
			CircuitBreaker: CircuitBreakerConfig{
				Enabled:              true,
				Timeout:              Duration(30 * time.Second),
				MaxRequests:          100,
				Interval:             Duration(60 * time.Second),
				ReadyToTripThreshold: 0.6,
				SuccessThreshold:     2,
				RequestTimeout:       Duration(10 * time.Second),
			},
			DLQ: DLQConfig{
				Enabled:                true,
				TopicSuffix:            "_dlq",
				RetryCount:             3,
				RetryDelay:             Duration(5 * time.Second),
				MaxMessageSize:         1048576,
				EnableSchemaValidation: true,
			},
			Retry: RetryConfig{
				MaxAttempts:    5,
				InitialDelay:   Duration(1 * time.Second),
				MaxDelay:       Duration(60 * time.Second),
				Multiplier:     2.0,
				Jitter:         0.2,
				RetryableCodes: []string{"5xx", "timeout", "network"},
			},
			Fallbacks: FallbackConfig{
				Enabled:             true,
				CacheMaxAge:         Duration(5 * time.Minute),
				DegradedModeEnabled: true,
				DegradedMaxLatency:  Duration(500 * time.Millisecond),
			},
		},
	}
}
