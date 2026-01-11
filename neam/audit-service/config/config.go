package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the audit service
type Config struct {
	// Server configuration
	ServerPort          int           `env:"SERVER_PORT" default:"9090"`
	Environment         string        `env:"ENVIRONMENT" default:"development"`
	Debug               bool          `env:"DEBUG" default:"false"`
	ReadTimeout         time.Duration `env:"READ_TIMEOUT" default:"30s"`
	WriteTimeout        time.Duration `env:"WRITE_TIMEOUT" default:"30s"`
	IdleTimeout         time.Duration `env:"IDLE_TIMEOUT" default:"120s"`

	// Database configuration
	DatabaseURL         string `env:"DATABASE_URL" required:"true"`
	MaxOpenConns        int    `env:"MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns        int    `env:"MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime     time.Duration `env:"CONN_MAX_LIFETIME" default:"5m"`

	// WORM Storage configuration
	WORMStoragePath     string `env:"WORM_STORAGE_PATH" default:"/data/audit/worm"`
	WORMRetentionDays   int    `env:"WORM_RETENTION_DAYS" default:"2555"` // 7 years for court-grade
	MinFileSize         int64  `env:"WORM_MIN_FILE_SIZE" default:"1048576"` // 1MB

	// Cryptographic configuration
	HSMEnabled          bool   `env:"HSM_ENABLED" default:"false"`
	HSMPin             string `env:"HSM_PIN" default:""`
	HSMKeyLabel         string `env:"HSM_KEY_LABEL" default:"audit-seal-key"`
	SoftwareKeyPath     string `env:"SOFTWARE_KEY_PATH" default:"/config/keys/signing.key"`
	SealingAlgorithm    string `env:"SEALING_ALGORITHM" default:"SHA256"`

	// Merkle Tree configuration
	MerkleTreeSealingCron string `env:"MERKLE_SEALING_CRON" default:"0 0 * * *"` // Daily at midnight
	MerkleTreeSize        int    `env:"MERKLE_TREE_SIZE" default:"1000"` // Events per tree

	// Integrity verification
	IntegrityCheckInterval time.Duration `env:"INTEGRITY_CHECK_INTERVAL" default:"1h"`

	// Keycloak configuration
	KeycloakEnabled       bool   `env:"KEYCLOAK_ENABLED" default:"false"`
	KeycloakURL          string `env:"KEYCLOAK_URL" default:"http://localhost:8080"`
	KeycloakRealm        string `env:"KEYCLOAK_REALM" default:"neam"`
	KeycloakClientID     string `env:"KEYCLOAK_CLIENT_ID" default:"audit-service"`
	KeycloakClientSecret string `env:"KEYCLOAK_CLIENT_SECRET" default:""`

	// OpenSearch SIEM configuration
	OpenSearchEnabled    bool   `env:"OPENSEARCH_ENABLED" default:"true"`
	OpenSearchURL        string `env:"OPENSEARCH_URL" default:"http://localhost:9200"`
	OpenSearchIndex      string `env:"OPENSEARCH_INDEX" default:"neam-audit-logs"`
	OpenSearchUsername   string `env:"OPENSEARCH_USERNAME" default:""`
	OpenSearchPassword   string `env:"OPENSEARCH_PASSWORD" default:""`

	// JWT configuration
	JWTSecret           string `env:"JWT_SECRET" required:"true"`
	JWTIssuer           string `env:"JWT_ISSUER" default:"neam-platform"`
	JWTAudience         string `env:"JWT_AUDIENCE" default:"neam-audit"`
	TokenExpiry         time.Duration `env:"TOKEN_EXPIRY" default:"1h"`
	RefreshTokenExpiry  time.Duration `env:"REFRESH_TOKEN_EXPIRY" default:"168h"` // 7 days

	// Rate limiting
	RateLimitRequests   int           `env:"RATE_LIMIT_REQUESTS" default:"100"`
	RateLimitWindow     time.Duration `env:"RATE_LIMIT_WINDOW" default:"1m"`

	// Compliance configuration
	ComplianceStandards []string `env:"COMPLIANCE_STANDARDS" default:"ISO27001,SOX,GDPR"`
	AuditRetentionYears int      `env:"AUDIT_RETENTION_YEARS" default:"7"`

	// Logging configuration
	LogLevel            string `env:"LOG_LEVEL" default:"info"`
	LogFormat           string `env:"LOG_FORMAT" default:"json"`

	// Air-gapped deployment
	AirGappedMode       bool   `env:"AIR_GAPPED_MODE" default:"false"`
	OfflineWitnessURL   string `env:"OFFLINE_WITNESS_URL" default:""`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load from .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{}

	// Set defaults
	cfg.ServerPort = 9090
	cfg.Environment = "development"
	cfg.Debug = false
	cfg.WORMStoragePath = "/data/audit/worm"
	cfg.WORMRetentionDays = 2555
	cfg.MerkleTreeSealingCron = "0 0 * * *"
	cfg.MerkleTreeSize = 1000
	cfg.IntegrityCheckInterval = time.Hour
	cfg.KeycloakEnabled = false
	cfg.OpenSearchEnabled = true
	cfg.OpenSearchIndex = "neam-audit-logs"
	cfg.JWTIssuer = "neam-platform"
	cfg.JWTAudience = "neam-audit"
	cfg.TokenExpiry = time.Hour
	cfg.RefreshTokenExpiry = 168 * time.Hour
	cfg.RateLimitRequests = 100
	cfg.RateLimitWindow = time.Minute
	cfg.ComplianceStandards = []string{"ISO27001", "SOX", "GDPR"}
	cfg.AuditRetentionYears = 7
	cfg.LogLevel = "info"
	cfg.LogFormat = "json"
	cfg.AirGappedMode = false

	// Override with environment variables
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.ServerPort = p
		}
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.DatabaseURL = dbURL
	}

	if wormPath := os.Getenv("WORM_STORAGE_PATH"); wormPath != "" {
		cfg.WORMStoragePath = wormPath
	}

	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.JWTSecret = jwtSecret
	}

	if keycloakURL := os.Getenv("KEYCLOAK_URL"); keycloakURL != "" {
		cfg.KeycloakURL = keycloakURL
	}

	if keycloakRealm := os.Getenv("KEYCLOAK_REALM"); keycloakRealm != "" {
		cfg.KeycloakRealm = keycloakRealm
	}

	if opensearchURL := os.Getenv("OPENSEARCH_URL"); opensearchURL != "" {
		cfg.OpenSearchURL = opensearchURL
	}

	if hsmEnabled := os.Getenv("HSM_ENABLED"); hsmEnabled == "true" {
		cfg.HSMEnabled = true
	}

	if hsmPin := os.Getenv("HSM_PIN"); hsmPin != "" {
		cfg.HSMPin = hsmPin
	}

	if airGapped := os.Getenv("AIR_GAPPED_MODE"); airGapped == "true" {
		cfg.AirGappedMode = true
	}

	// Validate required configuration
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

// GetStoragePath returns the WORM storage path
func (c *Config) GetStoragePath() string {
	return c.WORMStoragePath
}

// GetRetentionDuration returns the retention duration
func (c *Config) GetRetentionDuration() time.Duration {
	return time.Duration(c.WORMRetentionDays) * 24 * time.Hour
}

// IsComplianceEnabled checks if a specific compliance standard is enabled
func (c *Config) IsComplianceEnabled(standard string) bool {
	for _, s := range c.ComplianceStandards {
		if strings.EqualFold(s, standard) {
			return true
		}
	}
	return false
}

// GetDatabaseConfig returns database connection configuration
func (c *Config) GetDatabaseConfig() map[string]interface{} {
	return map[string]interface{}{
		"max_open_conns":    c.MaxOpenConns,
		"max_idle_conns":    c.MaxIdleConns,
		"conn_max_lifetime": c.ConnMaxLifetime.String(),
	}
}
