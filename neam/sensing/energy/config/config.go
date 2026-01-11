package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete energy adapter configuration
type Config struct {
	Service    ServiceConfig     `yaml:"service"`
	ICCP       ICCPConfig        `yaml:"iccp"`
	IEC61850   IEC61850Config    `yaml:"iec61850"`
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

// ICCPConfig contains ICCP/TASE.2 protocol configuration
type ICCPConfig struct {
	Enabled         bool             `yaml:"enabled"`
	Sources         []ICCPEndPoint   `yaml:"sources"`
	PointsOfInterest []string        `yaml:"points_of_interest"`
	Timeout         Duration         `yaml:"timeout"`
	RetryInterval   Duration         `yaml:"retry_interval"`
	KeepAlive       Duration         `yaml:"keep_alive"`
}

// ICCPEndPoint represents an ICCP data source
type ICCPEndPoint struct {
	EndPoint        string `yaml:"endpoint"`
	LocalTSAP       string `yaml:"local_tsap"`
	RemoteTSAP      string `yaml:"remote_tsap"`
	BilateralTable  string `yaml:"bilateral_table"`
	LocalIP         string `yaml:"local_ip"`
	VCCSelector     string `yaml:"vcc_selector"`
	DomainSelector  string `yaml:"domain_selector"`
}

// IEC61850Config contains IEC 61850/MMS protocol configuration
type IEC61850Config struct {
	Enabled          bool               `yaml:"enabled"`
	Sources          []IEC61850EndPoint `yaml:"sources"`
	PointsOfInterest []string           `yaml:"points_of_interest"`
	Timeout          Duration           `yaml:"timeout"`
	RetryInterval    Duration           `yaml:"retry_interval"`
	KeepAlive        Duration           `yaml:"keep_alive"`
}

// IEC61850EndPoint represents an IEC 61850 data source
type IEC61850EndPoint struct {
	EndPoint     string `yaml:"endpoint"`
	Port         int    `yaml:"port"`
	IEDName      string `yaml:"ied_name"`
	APTitle      string `yaml:"ap_title"`
	AEQualifier  int    `yaml:"ae_qualifier"`
	ReportCBName string `yaml:"report_cb_name"`
	RCBEnabled   bool   `yaml:"rcb_enabled"`
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
	MaxRetries       int      `yaml:"max_retries"`
	InitialDelay     Duration `yaml:"initial_delay"`
	MaxDelay         Duration `yaml:"max_delay"`
	Multiplier       float64  `yaml:"multiplier"`
	HealthCheckInterval Duration `yaml:"health_check_interval"`
	ReconnectWait    Duration `yaml:"reconnect_wait"`
}

// FilteringConfig contains signal processing configuration
type FilteringConfig struct {
	DeadbandPercent     float64 `yaml:"deadband_percent"`
	SpikeThreshold      float64 `yaml:"spike_threshold"`
	SpikeWindowSize     int     `yaml:"spike_window_size"`
	RateOfChangeLimit   float64 `yaml:"rate_of_change_limit"`
	QualityFilter       []string `yaml:"quality_filter"`
	MinReportingInterval Duration `yaml:"min_reporting_interval"`
	MaxValue            float64 `yaml:"max_value"`
	MinValue            float64 `yaml:"min_value"`
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
	Enabled              bool      `yaml:"enabled"`
	Timeout              Duration  `yaml:"timeout"`
	MaxRequests          uint32    `yaml:"max_requests"`
	Interval             Duration  `yaml:"interval"`
	ReadyToTripThreshold float64   `yaml:"ready_to_trip_threshold"`
	SuccessThreshold     float64   `yaml:"success_threshold"`
	RequestTimeout       Duration  `yaml:"request_timeout"`
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
	Enabled             bool     `yaml:"enabled"`
	CacheMaxAge         Duration `yaml:"cache_max_age"`
	DegradedModeEnabled bool     `yaml:"degraded_mode_enabled"`
	DegradedMaxLatency  Duration `yaml:"degraded_max_latency"`
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
	if endpoint := os.Getenv("ICCP_ENDPOINT"); endpoint != "" {
		if len(c.ICCP.Sources) > 0 {
			c.ICCP.Sources[0].EndPoint = endpoint
		}
	}
	if endpoint := os.Getenv("IEC61850_ENDPOINT"); endpoint != "" {
		if len(c.IEC61850.Sources) > 0 {
			c.IEC61850.Sources[0].EndPoint = endpoint
		}
	}
}

// Validate checks that the configuration is valid
func (c *Config) Validate() error {
	if c.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if !c.ICCP.Enabled && !c.IEC61850.Enabled {
		return fmt.Errorf("at least one protocol must be enabled")
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
			Name:        "energy-adapter",
			Host:        "0.0.0.0",
			Port:        8085,
			Environment: "development",
			Shutdown:    30 * time.Second,
		},
		ICCP: ICCPConfig{
			Enabled:          true,
			Sources:          []ICCPEndPoint{},
			PointsOfInterest: []string{"Frequency", "Voltage", "Current", "Power"},
			Timeout:          Duration(30 * time.Second),
			RetryInterval:    Duration(5 * time.Second),
			KeepAlive:        Duration(10 * time.Second),
		},
		IEC61850: IEC61850Config{
			Enabled:          true,
			Sources:          []IEC61850EndPoint{},
			PointsOfInterest: []string{"MMXU1.Vol", "MMXU1.A", "MMXU1.W", "XCBR.Pos"},
			Timeout:          Duration(30 * time.Second),
			RetryInterval:    Duration(5 * time.Second),
			KeepAlive:        Duration(10 * time.Second),
		},
		Kafka: KafkaConfig{
			Brokers:      "localhost:9092",
			OutputTopic:  "energy.telemetry",
			DLQTopic:     "energy.dlq",
			Version:      "2.8.0",
			MaxOpenReqs:  100,
			DialTimeout:  Duration(10 * time.Second),
			WriteTimeout: Duration(30 * time.Second),
		},
		Session: SessionConfig{
			MaxRetries:         10,
			InitialDelay:       Duration(1 * time.Second),
			MaxDelay:           Duration(60 * time.Second),
			Multiplier:         2.0,
			HealthCheckInterval: Duration(30 * time.Second),
			ReconnectWait:      Duration(5 * time.Second),
		},
		Filtering: FilteringConfig{
			DeadbandPercent:     0.1, // 0.1%
			SpikeThreshold:      20.0, // 20% change
			SpikeWindowSize:     5,
			RateOfChangeLimit:   5.0, // 5% per second max
			QualityFilter:       []string{"Good"},
			MinReportingInterval: Duration(100 * time.Millisecond),
			MaxValue:            1000000.0,
			MinValue:            -1000000.0,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "/var/log/neam",
			JSONFormat: true,
		},
		Monitoring: MonitoringConfig{
			Enabled:    true,
			Port:       9091,
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
