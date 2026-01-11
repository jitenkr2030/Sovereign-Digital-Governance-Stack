package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete industrial adapter configuration
type Config struct {
	Service    ServiceConfig        `yaml:"service"`
	OPCUA      OPCUAConfig          `yaml:"opcua"`
	Discovery  DiscoveryConfig      `yaml:"discovery"`
	Kafka      KafkaConfig          `yaml:"kafka"`
	Session    SessionConfig        `yaml:"session"`
	Filtering  FilteringConfig      `yaml:"filtering"`
	Logging    LoggingConfig        `yaml:"logging"`
	Monitoring MonitoringConfig     `yaml:"monitoring"`
	Resilience ResilienceConfig     `yaml:"resilience"`
}

// ServiceConfig contains service-level configuration
type ServiceConfig struct {
	Name        string        `yaml:"name"`
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	Environment string        `yaml:"environment"`
	Shutdown    time.Duration `yaml:"shutdown_timeout"`
}

// OPCUAConfig contains OPC UA protocol configuration
type OPCUAConfig struct {
	Enabled          bool               `yaml:"enabled"`
	Sources          []OPCUAEndPoint    `yaml:"sources"`
	NodesOfInterest  []string           `yaml:"nodes_of_interest"`
	Timeout          Duration           `yaml:"timeout"`
	RetryInterval    Duration           `yaml:"retry_interval"`
	ReconnectWait    Duration           `yaml:"reconnect_wait"`
	SubscriptionRate Duration           `yaml:"subscription_rate"`
	KeepAlive        Duration           `yaml:"keep_alive"`
}

// OPCUAEndPoint represents an OPC UA data source
type OPCUAEndPoint struct {
	EndpointURL     string `yaml:"endpoint_url"`
	ApplicationURI  string `yaml:"application_uri"`
	ProductURI      string `yaml:"product_uri"`
	ApplicationName string `yaml:"application_name"`
	SecurityPolicy  string `yaml:"security_policy"`  // None, Basic256Sha256
	SecurityMode    string `yaml:"security_mode"`    // None, Sign, SignAndEncrypt
	AuthMode        string `yaml:"auth_mode"`        // Anonymous, UserName, Certificate
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	CertificatePath string `yaml:"certificate_path"`
	PrivateKeyPath  string `yaml:"private_key_path"`
	NodesOfInterest []string `yaml:"nodes_of_interest"`
}

// DiscoveryConfig contains asset discovery configuration
type DiscoveryConfig struct {
	Enabled          bool         `yaml:"enabled"`
	DiscoveryInterval Duration    `yaml:"discovery_interval"`
	BrowseDepth      int          `yaml:"browse_depth"`
	FilterByType     []string     `yaml:"filter_by_type"`  // e.g., BaseAnalogType, BaseDataVariable
	IncludeChildren  bool         `yaml:"include_children"`
	CacheTimeout     Duration     `yaml:"cache_timeout"`
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
	StateStoreType      string   `yaml:"state_store_type"`
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
	MinValue             float64   `json:"min_value"`
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
	if endpoint := os.Getenv("OPCUA_ENDPOINT_URL"); endpoint != "" {
		if len(c.OPCUA.Sources) > 0 {
			c.OPCUA.Sources[0].EndpointURL = endpoint
		}
	}
	if username := os.Getenv("OPCUA_USERNAME"); username != "" {
		if len(c.OPCUA.Sources) > 0 {
			c.OPCUA.Sources[0].Username = username
		}
	}
	if password := os.Getenv("OPCUA_PASSWORD"); password != "" {
		if len(c.OPCUA.Sources) > 0 {
			c.OPCUA.Sources[0].Password = password
		}
	}
}

// Validate checks that the configuration is valid
func (c *Config) Validate() error {
	if c.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if !c.OPCUA.Enabled {
		return fmt.Errorf("OPC UA protocol must be enabled")
	}
	if len(c.OPCUA.Sources) == 0 {
		return fmt.Errorf("at least one OPC UA source is required")
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
			Name:        "industrial-adapter",
			Host:        "0.0.0.0",
			Port:        8087,
			Environment: "development",
			Shutdown:    30 * time.Second,
		},
		OPCUA: OPCUAConfig{
			Enabled: true,
			Sources: []OPCUAEndPoint{
				{
					EndpointURL:     "opc.tcp://localhost:4840",
					ApplicationURI:  "urn:neam:industrial:adapter",
					ProductURI:      "urn:neam:industrial:adapter",
					ApplicationName: "NEAM Industrial Adapter",
					SecurityPolicy:  "Basic256Sha256",
					SecurityMode:    "SignAndEncrypt",
					AuthMode:        "Anonymous",
					NodesOfInterest: []string{
						"ns=2;s=Channel1.Device1.Temperature",
						"ns=2;s=Channel1.Device1.Pressure",
						"ns=2;s=Channel1.Device1.Flow",
					},
				},
			},
			NodesOfInterest: []string{
				"ns=2;s=Channel1.Device1.Temperature",
				"ns=2;s=Channel1.Device1.Pressure",
			},
			Timeout:          Duration(30 * time.Second),
			RetryInterval:    Duration(5 * time.Second),
			ReconnectWait:    Duration(1 * time.Second),
			SubscriptionRate: Duration(100 * time.Millisecond),
			KeepAlive:        Duration(10 * time.Second),
		},
		Discovery: DiscoveryConfig{
			Enabled:           true,
			DiscoveryInterval: Duration(300 * time.Second), // 5 minutes
			BrowseDepth:       3,
			FilterByType:      []string{"BaseAnalogType", "BaseDataVariable"},
			IncludeChildren:   true,
			CacheTimeout:      Duration(600 * time.Second), // 10 minutes
		},
		Kafka: KafkaConfig{
			Brokers:      "localhost:9092",
			OutputTopic:  "industrial.telemetry",
			DLQTopic:     "industrial.dlq",
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
			Port:       9093,
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
