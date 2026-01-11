package config

import (
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the ingestion service
type Config struct {
	// Application settings
	AppHost    string `envconfig:"APP_HOST" default:"0.0.0.0"`
	AppPort    int    `envconfig:"APP_PORT" default:"8080"`
	Debug      bool   `envconfig:"DEBUG" default:"false"`
	Env        string `envconfig:"ENV" default:"development"`

	// Database settings
	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     int    `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER" default:"csic"`
	DBPassword string `envconfig:"DB_PASSWORD" default:"csic_secret"`
	DBName     string `envconfig:"DB_NAME" default:"csic_platform"`
	DBSSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`

	// Kafka settings
	KafkaBrokers     string `envconfig:"KAFKA_BROKERS" default:"localhost:9092"`
	KafkaTopicPrefix string `envconfig:"KAFKA_TOPIC_PREFIX" default:"csic"`

	// Ingestion settings
	DefaultPollingRate time.Duration `envconfig:"DEFAULT_POLLING_RATE" default:"5s"`
	DefaultTimeout     time.Duration `envconfig:"DEFAULT_TIMEOUT" default:"30s"`
	MaxRetryAttempts   int           `envconfig:"MAX_RETRY_ATTEMPTS" default:"3"`

	// Logging settings
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := parseYAML(data, cfg); err != nil {
		return nil, err
	}

	// Override with environment variables
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// parseYAML is a placeholder for YAML parsing
// In a real implementation, you would use a YAML library like gopkg.in/yaml.v3
func parseYAML(data []byte, cfg *Config) error {
	// Simplified - in production use proper YAML parsing
	return nil
}

// DSN returns the PostgreSQL connection string
func (c *Config) DSN() string {
	return "host=" + c.DBHost +
		" port=" + itoa(c.DBPort) +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

// itoa converts int to string
func itoa(i int) string {
	return string(rune('0'+i/1000%10)) + string(rune('0'+i/100%10)) + string(rune('0'+i/10%10)) + string(rune('0'+i%10))
}
