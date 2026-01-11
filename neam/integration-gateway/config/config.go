package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the integration gateway
type Config struct {
	// Server configuration
	Server struct {
		Environment string `yaml:"environment" env:"ENVIRONMENT" default:"development"`
		HTTPPort    int    `yaml:"http_port" env:"HTTP_PORT" default:"8080"`
		GRPCPort    int    `yaml:"grpc_port" env:"GRPC_PORT" default:"9090"`
		MetricsPort int    `yaml:"metrics_port" env:"METRICS_PORT" default:"9091"`
		ReadTimeout int    `yaml:"read_timeout" env:"READ_TIMEOUT" default:"30"`
		WriteTimeout int   `yaml:"write_timeout" env:"WRITE_TIMEOUT" default:"30"`
		IdleTimeout  int    `yaml:"idle_timeout" env:"IDLE_TIMEOUT" default:"120"`
	} `yaml:"server"`

	// Kafka configuration
	Kafka struct {
		Brokers         []string `yaml:"brokers" env:"KAFKA_BROKERS" default:"localhost:9092"`
		ConsumerGroup   string   `yaml:"consumer_group" env:"KAFKA_CONSUMER_GROUP" default:"integration-gateway"`
		IngestionTopic  string   `yaml:"ingestion_topic" env:"KAFKA_INGESTION_TOPIC" default:"neam.ingestion.v1"`
		BroadcastTopic  string   `yaml:"broadcast_topic" env:"KAFKA_BROADCAST_TOPIC" default:"neam.broadcast.v1"`
		RequestTopic    string   `yaml:"request_topic" env:"KAFKA_REQUEST_TOPIC" default:"neam.request.v1"`
		ResponseTopic   string   `yaml:"response_topic" env:"KAFKA_RESPONSE_TOPIC" default:"neam.response.v1"`
		TLSEnabled      bool     `yaml:"tls_enabled" env:"KAFKA_TLS_ENABLED" default:"false"`
		SASLEnabled     bool     `yaml:"sasl_enabled" env:"KAFKA_SASL_ENABLED" default:"false"`
		SASLMechanism   string   `yaml:"sasl_mechanism" env:"KAFKA_SASL_MECHANISM" default:"PLAIN"`
		Username        string   `yaml:"username" env:"KAFKA_USERNAME" default:""`
		Password        string   `yaml:"password" env:"KAFKA_PASSWORD" default:""`
		ProducerTimeout int      `yaml:"producer_timeout" env:"KAFKA_PRODUCER_TIMEOUT" default:"10"`
		ConsumerTimeout int      `yaml:"consumer_timeout" env:"KAFKA_CONSUMER_TIMEOUT" default:"10"`
	} `yaml:"kafka"`

	// TLS configuration
	TLS struct {
		Enabled           bool   `yaml:"enabled" env:"TLS_ENABLED" default:"true"`
		CertPath          string `yaml:"cert_path" env:"TLS_CERT_PATH" default:"/certs/server.crt"`
		KeyPath           string `yaml:"key_path" env:"TLS_KEY_PATH" default:"/certs/server.key"`
		CAPath            string `yaml:"ca_path" env:"TLS_CA_PATH" default:"/certs/ca.crt"`
		MinVersion        string `yaml:"min_version" env:"TLS_MIN_VERSION" default:"1.2"`
		ClientAuthEnabled bool   `yaml:"client_auth_enabled" env:"TLS_CLIENT_AUTH_ENABLED" default:"true"`
		Renegotiation     string `yaml:"renegotiation" env:"TLS_RENEGOTIATION" default:"never"`
	} `yaml:"tls"`

	// Rate limiting configuration
	RateLimit struct {
		Requests  int `yaml:"requests" env:"RATE_LIMIT_REQUESTS" default:"1000"`
		WindowSec int `yaml:"window_sec" env:"RATE_LIMIT_WINDOW_SEC" default:"60"`
	} `yaml:"rate_limit"`

	// Adapter configurations
	Adapters struct {
		Ministry struct {
			Endpoint      string   `yaml:"endpoint" env:"MINISTRY_ENDPOINT" default:"http://localhost:8001"`
			Timeout       int      `yaml:"timeout" env:"MINISTRY_TIMEOUT" default:"30"`
			RetryAttempts int      `yaml:"retry_attempts" env:"MINISTRY_RETRY_ATTEMPTS" default:"3"`
			RetryDelay    int      `yaml:"retry_delay" env:"MINISTRY_RETRY_DELAY" default:"1"`
		} `yaml:"ministry"`

		CentralBank struct {
			Endpoint      string   `yaml:"endpoint" env:"CENTRAL_BANK_ENDPOINT" default:"http://localhost:8002"`
			Timeout       int      `yaml:"timeout" env:"CENTRAL_BANK_TIMEOUT" default:"30"`
			RetryAttempts int      `yaml:"retry_attempts" env:"CENTRAL_BANK_RETRY_ATTEMPTS" default:"3"`
			RetryDelay    int      `yaml:"retry_delay" env:"CENTRAL_BANK_RETRY_DELAY" default:"1"`
			APIKey        string   `yaml:"api_key" env:"CENTRAL_BANK_API_KEY" default:""`
		} `yaml:"central_bank"`

		State struct {
			Endpoint       string   `yaml:"endpoint" env:"STATE_ENDPOINT" default:"http://localhost:8003"`
			Timeout        int      `yaml:"timeout" env:"STATE_TIMEOUT" default:"30"`
			RetryAttempts  int      `yaml:"retry_attempts" env:"STATE_RETRY_ATTEMPTS" default:"3"`
			RetryDelay     int      `yaml:"retry_delay" env:"STATE_RETRY_DELAY" default:"1"`
			StateCode      string   `yaml:"state_code" env:"STATE_CODE" default:""`
		} `yaml:"state"`

		District struct {
			Endpoint      string   `yaml:"endpoint" env:"DISTRICT_ENDPOINT" default:"http://localhost:8004"`
			Timeout       int      `yaml:"timeout" env:"DISTRICT_TIMEOUT" default:"30"`
			RetryAttempts int      `yaml:"retry_attempts" env:"DISTRICT_RETRY_ATTEMPTS" default:"3"`
			RetryDelay    int      `yaml:"retry_delay" env:"DISTRICT_RETRY_DELAY" default:"1"`
		} `yaml:"district"`

		Legacy struct {
			Endpoint         string   `yaml:"endpoint" env:"LEGACY_ENDPOINT" default:"http://localhost:8005"`
			Timeout          int      `yaml:"timeout" env:"LEGACY_TIMEOUT" default:"60"`
			RetryAttempts    int      `yaml:"retry_attempts" env:"LEGACY_RETRY_ATTEMPTS" default:"3"`
			RetryDelay       int      `yaml:"retry_delay" env:"LEGACY_RETRY_DELAY" default:"5"`
			FTPServer        string   `yaml:"ftp_server" env:"LEGACY_FTP_SERVER" default:""`
			FTPUser          string   `yaml:"ftp_user" env:"LEGACY_FTP_USER" default:""`
			FTPPassword      string   `yaml:"ftp_password" env:"LEGACY_FTP_PASSWORD" default:""`
			InputDirectory   string   `yaml:"input_directory" env:"LEGACY_INPUT_DIRECTORY" default:"/data/legacy/input"`
			OutputDirectory  string   `yaml:"output_directory" env:"LEGACY_OUTPUT_DIRECTORY" default:"/data/legacy/output"`
			ArchiveDirectory string   `yaml:"archive_directory" env:"LEGACY_ARCHIVE_DIRECTORY" default:"/data/legacy/archive"`
			SupportedFormats []string `yaml:"supported_formats" env:"LEGACY_FORMATS" default:"csv,xml,json"`
		} `yaml:"legacy"`
	} `yaml:"adapters"`

	// Data transformation configuration
	Transformation struct {
		EnablePIIProtection bool     `yaml:"enable_pii_protection" env:"TRANSFORM_PII_PROTECTION" default:"true"`
		DefaultDateFormat   string   `yaml:"default_date_format" env:"TRANSFORM_DATE_FORMAT" default:"2006-01-02"`
		TargetDateFormat    string   `yaml:"target_date_format" env:"TRANSFORM_TARGET_DATE_FORMAT" default:"2006-01-02T15:04:05Z07:00"`
		FieldMappings       []string `yaml:"field_mappings" env:"TRANSFORM_FIELD_MAPPINGS" default:""`
	} `yaml:"transformation"`

	// Security configuration
	Security struct {
		APIKeys          []string `yaml:"api_keys" env:"API_KEYS" default:""`
		AllowedOrigins   []string `yaml:"allowed_origins" env:"ALLOWED_ORIGINS" default:"*"`
		AllowedMethods   []string `yaml:"allowed_methods" env:"ALLOWED_METHODS" default:"GET,POST,PUT,DELETE"`
		AllowedHeaders   []string `yaml:"allowed_headers" env:"ALLOWED_HEADERS" default:"Content-Type,Authorization,X-Request-ID"`
		CORSMaxAge       int      `yaml:"cors_max_age" env:"CORS_MAX_AGE" default:"86400"`
		EnableCORS       bool     `yaml:"enable_cors" env:"ENABLE_CORS" default:"true"`
	} `yaml:"security"`

	// Monitoring configuration
	Monitoring struct {
		EnableMetrics    bool   `yaml:"enable_metrics" env:"ENABLE_METRICS" default:"true"`
		EnableTracing    bool   `yaml:"enable_tracing" env:"ENABLE_TRACING" default:"true"`
		TracingEndpoint  string `yaml:"tracing_endpoint" env:"TRACING_ENDPOINT" default:""`
		MetricsEndpoint  string `yaml:"metrics_endpoint" env:"METRICS_ENDPOINT" default:"/metrics"`
		HealthEndpoint   string `yaml:"health_endpoint" env:"HEALTH_ENDPOINT" default:"/health"`
	} `yaml:"monitoring"`

	// Logging configuration
	Logging struct {
		Level  string `yaml:"level" env:"LOG_LEVEL" default:"info"`
		Format string `yaml:"format" env:"LOG_FORMAT" default:"json"`
		Output string `yaml:"output" env:"LOG_OUTPUT" default:"stdout"`
	} `yaml:"logging"`
}

// Load loads configuration from YAML file
func Load(configPath string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try default paths
		defaultPaths := []string{
			"config.yaml",
			"config/config.yaml",
			"/etc/neam/integration-gateway.yaml",
		}

		for _, path := range defaultPaths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
	}

	// Load from file
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Return default config if file not found
		return &Config{}, nil
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(&cfg)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides
func applyEnvOverrides(cfg *Config) {
	// Server configuration
	if v := os.Getenv("ENVIRONMENT"); v != "" {
		cfg.Server.Environment = v
	}
	if v := os.Getenv("HTTP_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.HTTPPort)
	}
	if v := os.Getenv("GRPC_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.GRPCPort)
	}

	// Kafka configuration
	if v := os.Getenv("KAFKA_BROKERS"); v != "" {
		cfg.Kafka.Brokers = []string{v}
	}
	if v := os.Getenv("KAFKA_TLS_ENABLED"); v == "true" {
		cfg.Kafka.TLSEnabled = true
	}

	// TLS configuration
	if v := os.Getenv("TLS_ENABLED"); v == "false" {
		cfg.TLS.Enabled = false
	}
	if v := os.Getenv("TLS_CLIENT_AUTH_ENABLED"); v == "false" {
		cfg.TLS.ClientAuthEnabled = false
	}

	// Rate limiting
	if v := os.Getenv("RATE_LIMIT_REQUESTS"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.RateLimit.Requests)
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server configuration
	if c.Server.HTTPPort <= 0 || c.Server.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", c.Server.HTTPPort)
	}
	if c.Server.GRPCPort <= 0 || c.Server.GRPCPort > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", c.Server.GRPCPort)
	}

	// Validate Kafka configuration
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("at least one Kafka broker is required")
	}

	// Validate TLS configuration
	if c.TLS.Enabled {
		if c.TLS.CertPath == "" {
			return fmt.Errorf("TLS certificate path is required when TLS is enabled")
		}
		if c.TLS.KeyPath == "" {
			return fmt.Errorf("TLS key path is required when TLS is enabled")
		}
	}

	// Validate rate limiting
	if c.RateLimit.Requests <= 0 {
		return fmt.Errorf("rate limit requests must be positive")
	}
	if c.RateLimit.WindowSec <= 0 {
		return fmt.Errorf("rate limit window must be positive")
	}

	return nil
}

// GetReadTimeout returns read timeout as duration
func (c *Config) GetReadTimeout() time.Duration {
	return time.Duration(c.Server.ReadTimeout) * time.Second
}

// GetWriteTimeout returns write timeout as duration
func (c *Config) GetWriteTimeout() time.Duration {
	return time.Duration(c.Server.WriteTimeout) * time.Second
}

// GetIdleTimeout returns idle timeout as duration
func (c *Config) GetIdleTimeout() time.Duration {
	return time.Duration(c.Server.IdleTimeout) * time.Second
}

// GetRateLimitWindow returns rate limit window as duration
func (c *Config) GetRateLimitWindow() time.Duration {
	return time.Duration(c.RateLimit.WindowSec) * time.Second
}

// GetAdapterTimeout returns timeout for specific adapter
func (c *Config) GetAdapterTimeout(adapterType string) time.Duration {
	var timeout int
	switch adapterType {
	case "ministry":
		timeout = c.Adapters.Ministry.Timeout
	case "central_bank":
		timeout = c.Adapters.CentralBank.Timeout
	case "state":
		timeout = c.Adapters.State.Timeout
	case "district":
		timeout = c.Adapters.District.Timeout
	case "legacy":
		timeout = c.Adapters.Legacy.Timeout
	default:
		timeout = 30
	}
	return time.Duration(timeout) * time.Second
}
