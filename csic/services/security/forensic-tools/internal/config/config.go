package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Kafka     KafkaConfig     `mapstructure:"kafka"`
	Analysis  AnalysisConfig  `mapstructure:"analysis"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
}

// AppConfig contains application metadata
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	MaxBodySize  int64  `mapstructure:"max_body_size"`
}

// DatabaseConfig contains PostgreSQL connection settings
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// StorageConfig contains blob storage settings
type StorageConfig struct {
	Type  string        `mapstructure:"type"`
	Local LocalStorage  `mapstructure:"local"`
	S3    S3Storage     `mapstructure:"s3"`
}

// LocalStorage contains local filesystem storage settings
type LocalStorage struct {
	BasePath   string `mapstructure:"base_path"`
	MaxFileSize int64 `mapstructure:"max_file_size"`
}

// S3Storage contains S3-compatible storage settings
type S3Storage struct {
	Endpoint  string `mapstructure:"endpoint"`
	Bucket    string `mapstructure:"bucket"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Region    string `mapstructure:"region"`
}

// KafkaConfig contains Kafka broker settings
type KafkaConfig struct {
	Brokers       []string          `mapstructure:"brokers"`
	ConsumerGroup string            `mapstructure:"consumer_group"`
	Topics        KafkaTopicsConfig `mapstructure:"topics"`
	Producer      KafkaProducerConfig `mapstructure:"producer"`
}

// KafkaTopicsConfig contains Kafka topic names
type KafkaTopicsConfig struct {
	AnalysisJobs    string `mapstructure:"analysis_jobs"`
	AnalysisResults string `mapstructure:"analysis_results"`
}

// KafkaProducerConfig contains Kafka producer settings
type KafkaProducerConfig struct {
	Acks   string `mapstructure:"acks"`
	Retries int   `mapstructure:"retries"`
}

// AnalysisConfig contains analysis tool settings
type AnalysisConfig struct {
	Tools            []AnalysisTool `mapstructure:"tools"`
	MaxConcurrentJobs int           `mapstructure:"max_concurrent_jobs"`
	JobTimeout       int           `mapstructure:"job_timeout"`
}

// AnalysisTool represents an analysis tool configuration
type AnalysisTool struct {
	Name    string `mapstructure:"name"`
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// MonitoringConfig contains monitoring settings
type MonitoringConfig struct {
	MetricsEnabled     bool   `mapstructure:"metrics_enabled"`
	HealthCheckInterval int   `mapstructure:"health_check_interval"`
	HealthPath         string `mapstructure:"health_path"`
	MetricsPath        string `mapstructure:"metrics_path"`
}

// ConfigLoader handles loading configuration from files and environment
type ConfigLoader struct {
	configPath string
	vip        *viper.Viper
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(configPath string) *ConfigLoader {
	vip := viper.New()
	vip.SetConfigFile(configPath)
	vip.SetConfigType("yaml")
	vip.AutomaticEnv()

	return &ConfigLoader{
		configPath: configPath,
		vip:        vip,
	}
}

// Load loads configuration from the specified file path
func (cl *ConfigLoader) Load() (*Config, error) {
	if err := cl.vip.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := cl.vip.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app name is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Storage.Type == "" {
		return fmt.Errorf("storage type is required")
	}
	return nil
}

// GetReadTimeout returns the read timeout as a duration
func (c *ServerConfig) GetReadTimeout() time.Duration {
	return time.Duration(c.ReadTimeout) * time.Second
}

// GetWriteTimeout returns the write timeout as a duration
func (c *ServerConfig) GetWriteTimeout() time.Duration {
	return time.Duration(c.WriteTimeout) * time.Second
}

// GetConnMaxLifetime returns the connection max lifetime as a duration
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Name, c.SSLMode,
	)
}
