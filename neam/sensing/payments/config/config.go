package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete payment adapter configuration
type Config struct {
	Service     ServiceConfig     `yaml:"service"`
	Kafka       KafkaConfig       `yaml:"kafka"`
	Redis       RedisConfig       `yaml:"redis"`
	Processing  ProcessingConfig  `yaml:"processing"`
	Retry       RetryConfig       `yaml:"retry"`
	Logging     LoggingConfig     `yaml:"logging"`
	Monitoring  MonitoringConfig  `yaml:"monitoring"`
	Batch       BatchConfig       `yaml:"batch"`
}

// ServiceConfig contains service-level configuration
type ServiceConfig struct {
	Name        string        `yaml:"name"`
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	Environment string        `yaml:"environment"`
	Shutdown    time.Duration `yaml:"shutdown_timeout"`
}

// KafkaConfig contains Kafka connection and topic configuration
type KafkaConfig struct {
	Brokers        string   `yaml:"brokers"`
	ConsumerGroup  string   `yaml:"consumer_group"`
	InputTopic     string   `yaml:"input_topic"`
	OutputTopic    string   `yaml:"output_topic"`
	DLQTopic       string   `yaml:"dlq_topic"`
	RetryTopic     string   `yaml:"retry_topic"`
	Version        string   `yaml:"version"`
	MaxOpenReqs    int      `yaml:"max_open_requests"`
	DialTimeout    Duration `yaml:"dial_timeout"`
	ReadTimeout    Duration `yaml:"read_timeout"`
	WriteTimeout   Duration `yaml:"write_timeout"`
	SessionTimeout Duration `yaml:"session_timeout"`
	HeartbeatInterval Duration `yaml:"heartbeat_interval"`
}

// RedisConfig contains Redis connection configuration
type RedisConfig struct {
	Addr            string   `yaml:"addr"`
	Password        string   `yaml:"password"`
	DB              int      `yaml:"db"`
	PoolSize        int      `yaml:"pool_size"`
	MinIdleConns    int      `yaml:"min_idle_conns"`
	DialTimeout     Duration `yaml:"dial_timeout"`
	ReadTimeout     Duration `yaml:"read_timeout"`
	WriteTimeout    Duration `yaml:"write_timeout"`
	IdempotencyTTL  Duration `yaml:"idempotency_ttl"`
}

// ProcessingConfig contains message processing configuration
type ProcessingConfig struct {
	Mode          string `yaml:"mode"`
	WorkerCount   int    `yaml:"worker_count"`
	BatchSize     int    `yaml:"batch_size"`
	BatchTimeout  Duration `yaml:"batch_timeout"`
	BatchDirectory string `yaml:"batch_directory"`
	ParallelFiles int    `yaml:"parallel_files"`
}

// RetryConfig contains retry policy configuration
type RetryConfig struct {
	MaxRetries      int           `yaml:"max_retries"`
	InitialDelay    Duration      `yaml:"initial_delay"`
	MaxDelay        Duration      `yaml:"max_delay"`
	Multiplier      float64       `yaml:"multiplier"`
	Jitter          bool          `yaml:"jitter"`
	JitterFactor    float64       `yaml:"jitter_factor"`
	DLQMaxRetries   int           `yaml:"dlq_max_retries"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	OutputPath string `yaml:"output_path"`
	JSONFormat bool   `yaml:"json_format"`
}

// MonitoringConfig contains monitoring and metrics configuration
type MonitoringConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Port       int    `yaml:"port"`
	Path       string `yaml:"path"`
	HealthPath string `yaml:"health_path"`
}

// BatchConfig contains batch processing configuration
type BatchConfig struct {
	SupportedFormats []string `yaml:"supported_formats"`
	MaxFileSize      int64    `yaml:"max_file_size"`
	ScanInterval     Duration `yaml:"scan_interval"`
	ArchiveProcessed bool     `yaml:"archive_processed"`
	ArchivePath      string   `yaml:"archive_path"`
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
	if port := os.Getenv("SERVICE_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &c.Service.Port)
	}
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		c.Kafka.Brokers = brokers
	}
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		c.Redis.Addr = addr
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		c.Redis.Password = password
	}
	if mode := os.Getenv("PROCESSING_MODE"); mode != "" {
		c.Processing.Mode = mode
	}
}

// Validate checks that the configuration is valid
func (c *Config) Validate() error {
	if c.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if c.Kafka.Brokers == "" {
		return fmt.Errorf("Kafka brokers is required")
	}
	if c.Kafka.InputTopic == "" {
		return fmt.Errorf("Kafka input topic is required")
	}
	if c.Kafka.OutputTopic == "" {
		return fmt.Errorf("Kafka output topic is required")
	}
	if c.Redis.Addr == "" {
		return fmt.Errorf("Redis address is required")
	}
	if c.Processing.WorkerCount <= 0 {
		return fmt.Errorf("worker count must be positive")
	}
	if c.Processing.Mode != "realtime" && c.Processing.Mode != "batch" && c.Processing.Mode != "hybrid" {
		return fmt.Errorf("invalid processing mode: %s", c.Processing.Mode)
	}
	if c.Retry.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative")
	}
	return nil
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Service: ServiceConfig{
			Name:        "payment-adapter",
			Host:        "0.0.0.0",
			Port:        8084,
			Environment: "development",
			Shutdown:    30 * time.Second,
		},
		Kafka: KafkaConfig{
			Brokers:         "localhost:9092",
			ConsumerGroup:   "payment-adapter-group",
			InputTopic:      "payments.raw",
			OutputTopic:     "payments.normalized",
			DLQTopic:        "payments.dlq",
			RetryTopic:      "payments.retry",
			Version:         "2.8.0",
			MaxOpenReqs:     100,
			DialTimeout:     Duration(10 * time.Second),
			ReadTimeout:     Duration(30 * time.Second),
			WriteTimeout:    Duration(30 * time.Second),
			SessionTimeout:  Duration(10 * time.Second),
			HeartbeatInterval: Duration(3 * time.Second),
		},
		Redis: RedisConfig{
			Addr:           "localhost:6379",
			Password:       "",
			DB:             0,
			PoolSize:       100,
			MinIdleConns:   10,
			DialTimeout:    Duration(5 * time.Second),
			ReadTimeout:    Duration(3 * time.Second),
			WriteTimeout:   Duration(3 * time.Second),
			IdempotencyTTL: Duration(24 * time.Hour),
		},
		Processing: ProcessingConfig{
			Mode:           "realtime",
			WorkerCount:    10,
			BatchSize:      100,
			BatchTimeout:   Duration(5 * time.Second),
			BatchDirectory: "/data/batch/payments",
			ParallelFiles:  5,
		},
		Retry: RetryConfig{
			MaxRetries:    5,
			InitialDelay:  Duration(100 * time.Millisecond),
			MaxDelay:      Duration(10 * time.Second),
			Multiplier:    2.0,
			Jitter:        true,
			JitterFactor:  0.1,
			DLQMaxRetries: 3,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "/var/log/neam",
			JSONFormat: true,
		},
		Monitoring: MonitoringConfig{
			Enabled:    true,
			Port:       9090,
			Path:       "/metrics",
			HealthPath: "/health",
		},
		Batch: BatchConfig{
			SupportedFormats: []string{"xml", "txt", "json"},
			MaxFileSize:      100 * 1024 * 1024, // 100MB
			ScanInterval:     Duration(1 * time.Minute),
			ArchiveProcessed: true,
			ArchivePath:      "/data/batch/archive",
		},
	}
}
