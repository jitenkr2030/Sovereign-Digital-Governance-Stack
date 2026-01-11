package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration
type Config struct {
	App            AppConfig            `yaml:"app"`
	Server         ServerConfig         `yaml:"server"`
	Database       DatabaseConfig       `yaml:"database"`
	Redis          RedisConfig          `yaml:"redis"`
	Kafka          KafkaConfig          `yaml:"kafka"`
	Registration   RegistrationConfig   `yaml:"registration"`
	EnergyMonitoring EnergyMonitoringConfig `yaml:"energy_monitoring"`
	HashRateMonitoring HashRateMonitoringConfig `yaml:"hashrate_monitoring"`
	Compliance     ComplianceConfig     `yaml:"compliance"`
	Enforcement    EnforcementConfig    `yaml:"enforcement"`
	Regional       RegionalConfig       `yaml:"regional"`
	Integration    IntegrationConfig    `yaml:"integration"`
	Logging        LoggingConfig        `yaml:"logging"`
	Metrics        MetricsConfig        `yaml:"metrics"`
	Security       SecurityConfig       `yaml:"security"`
}

// AppConfig contains application metadata
type AppConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	Version     string `yaml:"version"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	IdleTimeout  int    `yaml:"idle_timeout"`
}

// DatabaseConfig contains PostgreSQL connection settings
type DatabaseConfig struct {
	Host            string            `yaml:"host"`
	Port            int               `yaml:"port"`
	Username        string            `yaml:"username"`
	Password        string            `yaml:"password"`
	Name            string            `yaml:"name"`
	SSLMode         string            `yaml:"ssl_mode"`
	MaxOpenConns    int               `yaml:"max_open_conns"`
	MaxIdleConns    int               `yaml:"max_idle_conns"`
	ConnMaxLifetime int               `yaml:"conn_max_lifetime"`
	TimescaleDB     TimescaleDBConfig `yaml:"timescaledb"`
}

// TimescaleDBConfig contains TimescaleDB specific settings
type TimescaleDBConfig struct {
	Enabled           bool   `yaml:"enabled"`
	ChunkInterval     string `yaml:"chunk_interval"`
	RetentionInterval string `yaml:"retention_interval"`
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Password  string `yaml:"password"`
	DB        int    `yaml:"db"`
	PoolSize  int    `yaml:"pool_size"`
	KeyPrefix string `yaml:"key_prefix"`
}

// KafkaConfig contains Kafka connection settings
type KafkaConfig struct {
	Brokers       []string          `yaml:"brokers"`
	TopicPrefix   string            `yaml:"topic_prefix"`
	ConsumerGroup string            `yaml:"consumer_group"`
	Topics        KafkaTopicsConfig `yaml:"topics"`
}

// KafkaTopicsConfig contains Kafka topic names
type KafkaTopicsConfig struct {
	Telemetry  string `yaml:"telemetry"`
	Violations string `yaml:"violations"`
	Compliance string `yaml:"compliance"`
	Events     string `yaml:"events"`
}

// RegistrationConfig contains registration settings
type RegistrationConfig struct {
	GeoRestrictions GeoRestrictionsConfig `yaml:"geo_restrictions"`
	Machine         MachineConfig         `yaml:"machine"`
	License         LicenseConfig         `yaml:"license"`
}

// GeoRestrictionsConfig contains geographic restrictions
type GeoRestrictionsConfig struct {
	Enabled              bool     `yaml:"enabled"`
	BannedRegions        []string `yaml:"banned_regions"`
	RestrictedRegions    []string `yaml:"restricted_regions"`
	AllowedEnergySources []string `yaml:"allowed_energy_sources"`
}

// MachineConfig contains machine registration settings
type MachineConfig struct {
	MaxMachinesPerPool    int      `yaml:"max_machines_per_pool"`
	RequireModelRegistration bool   `yaml:"require_model_registration"`
	AllowedMachineTypes   []string `yaml:"allowed_machine_types"`
}

// LicenseConfig contains license settings
type LicenseConfig struct {
	DefaultValidityDays    int  `yaml:"default_validity_days"`
	RenewalNotificationDays int  `yaml:"renewal_notification_days"`
	RequireComplianceCheck bool `yaml:"require_compliance_check"`
}

// EnergyMonitoringConfig contains energy monitoring settings
type EnergyMonitoringConfig struct {
	Thresholds     EnergyThresholdsConfig `yaml:"thresholds"`
	Carbon         CarbonConfig           `yaml:"carbon"`
	SourceTypes    []EnergySourceConfig   `yaml:"source_types"`
	Telemetry      TelemetryConfig        `yaml:"telemetry"`
}

// EnergyThresholdsConfig contains energy threshold settings
type EnergyThresholdsConfig struct {
	WarningLimitPercent    int `yaml:"warning_limit_percent"`
	CriticalLimitPercent   int `yaml:"critical_limit_percent"`
	OverageGracePeriodMinutes int `yaml:"overage_grace_period_minutes"`
}

// CarbonConfig contains carbon footprint calculation settings
type CarbonConfig struct {
	GridEmissionFactors map[string]float64 `yaml:"grid_emission_factors"`
	RenewableCredit     float64            `yaml:"renewable_credit"`
}

// EnergySourceConfig contains energy source settings
type EnergySourceConfig struct {
	Name         string  `yaml:"name"`
	CarbonFactor float64 `yaml:"carbon_factor"`
}

// TelemetryConfig contains telemetry reporting settings
type TelemetryConfig struct {
	MinIntervalSeconds int `yaml:"min_interval_seconds"`
	MaxIntervalSeconds int `yaml:"max_interval_seconds"`
	BatchSize          int `yaml:"batch_size"`
}

// HashRateMonitoringConfig contains hash rate monitoring settings
type HashRateMonitoringConfig struct {
	Thresholds  HashRateThresholdsConfig `yaml:"thresholds"`
	Validation  HashRateValidationConfig `yaml:"validation"`
	Telemetry   TelemetryConfig          `yaml:"telemetry"`
}

// HashRateThresholdsConfig contains hash rate threshold settings
type HashRateThresholdsConfig struct {
	AnomalyStdDev       float64 `yaml:"anomaly_stddev"`
	MinReportedHashRate float64 `yaml:"min_reported_hashrate"`
	MaxReportedHashRate float64 `yaml:"max_reported_hashrate"`
}

// HashRateValidationConfig contains hash rate validation settings
type HashRateValidationConfig struct {
	PowerHashRateRatioMin float64 `yaml:"power_hashrate_ratio_min"`
	PowerHashRateRatioMax float64 `yaml:"power_hashrate_ratio_max"`
	WorkerCountMin        int     `yaml:"worker_count_min"`
	WorkerCountMax        int     `yaml:"worker_count_max"`
}

// ComplianceConfig contains compliance settings
type ComplianceConfig struct {
	CheckInterval    int                  `yaml:"check_interval"`
	DailyReportTime  string               `yaml:"daily_report_time"`
	Certificate      CertificateConfig    `yaml:"certificate"`
	ViolationSeverity []ViolationSeverityConfig `yaml:"violation_severity"`
}

// CertificateConfig contains certificate settings
type CertificateConfig struct {
	ValidityDays  int  `yaml:"validity_days"`
	AutoRenewal   bool `yaml:"auto_renewal"`
	RequireInspection bool `yaml:"require_inspection"`
}

// ViolationSeverityConfig contains violation severity settings
type ViolationSeverityConfig struct {
	Name           string `yaml:"name"`
	AutoResolve    bool   `yaml:"auto_resolve"`
	RetentionDays  int    `yaml:"retention_days"`
}

// EnforcementConfig contains enforcement settings
type EnforcementConfig struct {
	AutoSuspend     AutoSuspendConfig `yaml:"auto_suspend"`
	Alert           AlertConfig       `yaml:"alert"`
	ExchangeSurveillance ExchangeSurveillanceConfig `yaml:"exchange_surveillance"`
}

// AutoSuspendConfig contains auto-suspend settings
type AutoSuspendConfig struct {
	Enabled          bool   `yaml:"enabled"`
	Threshold        string `yaml:"threshold"`
	GracePeriodMinutes int  `yaml:"grace_period_minutes"`
}

// AlertConfig contains alert settings
type AlertConfig struct {
	Enabled        bool   `yaml:"enabled"`
	EmailEnabled   bool   `yaml:"email_enabled"`
	WebhookEnabled bool   `yaml:"webhook_enabled"`
	WebhookURL     string `yaml:"webhook_url"`
}

// ExchangeSurveillanceConfig contains exchange surveillance integration settings
type ExchangeSurveillanceConfig struct {
	Enabled            bool   `yaml:"enabled"`
	WalletFreezeThreshold string `yaml:"wallet_freeze_threshold"`
}

// RegionalConfig contains regional settings
type RegionalConfig struct {
	EnergyLimits RegionalEnergyLimits `yaml:"energy_limits"`
	Restrictions []RegionalRestriction `yaml:"restrictions"`
}

// RegionalEnergyLimits contains regional energy limit settings
type RegionalEnergyLimits struct {
	Default float64          `yaml:"default"`
	ByRegion map[string]float64 `yaml:"by_region"`
}

// RegionalRestriction contains regional restriction settings
type RegionalRestriction struct {
	RegionCode        string   `yaml:"region_code"`
	Status            string   `yaml:"status"`
	Reason            string   `yaml:"reason"`
	MaxEnergyKW       float64  `yaml:"max_energy_kw"`
	AllowedEnergySources []string `yaml:"allowed_energy_sources"`
}

// IntegrationConfig contains integration settings
type IntegrationConfig struct {
	ComplianceService ComplianceServiceConfig `yaml:"compliance_service"`
	ExchangeSurveillance ExchangeSurveillanceServiceConfig `yaml:"exchange_surveillance"`
	AuditLog          AuditLogConfig          `yaml:"audit_log"`
}

// ComplianceServiceConfig contains compliance service settings
type ComplianceServiceConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Endpoint  string `yaml:"endpoint"`
	Timeout   int    `yaml:"timeout"`
	CacheTTL  int    `yaml:"cache_ttl"`
}

// ExchangeSurveillanceServiceConfig contains exchange surveillance service settings
type ExchangeSurveillanceServiceConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	Timeout  int    `yaml:"timeout"`
}

// AuditLogConfig contains audit log settings
type AuditLogConfig struct {
	Enabled      bool `yaml:"enabled"`
	Endpoint     string `yaml:"endpoint"`
	BatchSize    int  `yaml:"batch_size"`
	FlushInterval int `yaml:"flush_interval"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level          string            `yaml:"level"`
	Format         string            `yaml:"format"`
	Output         string            `yaml:"output"`
	IncludeCaller  bool              `yaml:"include_caller"`
	Fields         map[string]string `yaml:"fields"`
}

// MetricsConfig contains metrics settings
type MetricsConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	Port     int    `yaml:"port"`
	Custom   CustomMetricsConfig `yaml:"custom"`
}

// CustomMetricsConfig contains custom metrics settings
type CustomMetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Prefix  string `yaml:"prefix"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	TLS       TLSServerConfig   `yaml:"tls"`
	RateLimit RateLimitConfig   `yaml:"rate_limit"`
	CORS      CORSConfig        `yaml:"cors"`
}

// TLSServerConfig contains TLS settings
type TLSServerConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// RateLimitConfig contains rate limit settings
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
	BurstSize         int  `yaml:"burst_size"`
}

// CORSConfig contains CORS settings
type CORSConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
	MaxAge         int      `yaml:"max_age"`
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
	applyEnvOverrides(&cfg)

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides to the configuration
func applyEnvOverrides(cfg *Config) {
	// Database overrides
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil {
			cfg.Database.Port = port
		}
	}
	if v := os.Getenv("DB_USERNAME"); v != "" {
		cfg.Database.Username = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.Name = v
	}

	// Redis overrides
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil {
			cfg.Redis.Port = port
		}
	}

	// Server overrides
	if v := os.Getenv("APP_PORT"); v != "" {
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil {
			cfg.Server.Port = port
		}
	}
}

// GetConnMaxLifetime returns the database connection max lifetime as a duration
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}

// GetConnMaxLifetime returns the database connection max lifetime as a duration
func (c *Config) GetConnMaxLifetime() time.Duration {
	return c.Database.GetConnMaxLifetime()
}
