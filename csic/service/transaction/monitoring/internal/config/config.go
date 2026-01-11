package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration
type Config struct {
	App         AppConfig        `yaml:"app"`
	Server      ServerConfig     `yaml:"server"`
	Database    DatabaseConfig   `yaml:"database"`
	Neo4j       Neo4jConfig      `yaml:"neo4j"`
	Redis       RedisConfig      `yaml:"redis"`
	Blockchain  BlockchainConfig `yaml:"blockchain"`
	RiskScoring RiskScoringConfig `yaml:"risk_scoring"`
	Clustering  ClusteringConfig `yaml:"clustering"`
	Alerting    AlertingConfig   `yaml:"alerting"`
	Logging     LoggingConfig    `yaml:"logging"`
	Metrics     MetricsConfig    `yaml:"metrics"`
	Security    SecurityConfig   `yaml:"security"`
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
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	SSLMode         string `yaml:"ssl_mode"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

// Neo4jConfig contains Neo4j connection settings
type Neo4jConfig struct {
	URI            string `yaml:"uri"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	Database       string `yaml:"database"`
	MaxConnPool    int    `yaml:"max_conn_pool"`
	ConnectionTimeout int `yaml:"connection_timeout"`
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

// BlockchainConfig contains blockchain node settings
type BlockchainConfig struct {
	Bitcoin  BitcoinConfig  `yaml:"bitcoin"`
	Ethereum EthereumConfig `yaml:"ethereum"`
	Chains   []ChainConfig  `yaml:"chains"`
}

// BitcoinConfig contains Bitcoin node settings
type BitcoinConfig struct {
	Enabled     bool   `yaml:"enabled"`
	RPCUser     string `yaml:"rpc_user"`
	RPCPassword string `yaml:"rpc_password"`
	RPCHost     string `yaml:"rpc_host"`
	RPCPort     int    `yaml:"rpc_port"`
	UseTLS      bool   `yaml:"use_tls"`
	WalletName  string `yaml:"wallet_name"`
}

// EthereumConfig contains Ethereum node settings
type EthereumConfig struct {
	Enabled      bool   `yaml:"enabled"`
	RPCEndpoint  string `yaml:"rpc_endpoint"`
	WSEndpoint   string `yaml:"ws_endpoint"`
	ChainID      int    `yaml:"chain_id"`
	ContractAddr string `yaml:"contract_addr"`
}

// ChainConfig contains generic chain settings
type ChainConfig struct {
	ID            string `yaml:"id"`
	Name          string `yaml:"name"`
	Type          string `yaml:"type"`
	RPCEndpoint   string `yaml:"rpc_endpoint"`
	StartBlock    int64  `yaml:"start_block"`
	Confirmations int    `yaml:"confirmations"`
}

// RiskScoringConfig contains risk scoring settings
type RiskScoringConfig struct {
	CriticalThreshold  int             `yaml:"critical_threshold"`
	HighThreshold      int             `yaml:"high_threshold"`
	MediumThreshold    int             `yaml:"medium_threshold"`
	Rules              []RiskRuleConfig `yaml:"rules"`
	SanctionsListURL   string          `yaml:"sanctions_list_url"`
	RefreshInterval    string          `yaml:"refresh_interval"`
	EnableRealTime     bool            `yaml:"enable_real_time"`
}

// RiskRuleConfig contains individual risk rule settings
type RiskRuleConfig struct {
	ID          string  `yaml:"id"`
	Name        string  `yaml:"name"`
	Category    string  `yaml:"category"`
	Score       int     `yaml:"score"`
	Enabled     bool    `yaml:"enabled"`
	Description string  `yaml:"description"`
	Parameters  map[string]interface{} `yaml:"parameters"`
}

// ClusteringConfig contains entity clustering settings
type ClusteringConfig struct {
	Enabled              bool   `yaml:"enabled"`
	CommonInputEnabled   bool   `yaml:"common_input_enabled"`
	DepositClusterEnabled bool  `yaml:"deposit_cluster_enabled"`
	ChangeLinkEnabled    bool   `yaml:"change_link_enabled"`
	BatchInterval        string `yaml:"batch_interval"`
	RetentionDays        int    `yaml:"retention_days"`
	MaxClusterDepth      int    `yaml:"max_cluster_depth"`
}

// AlertingConfig contains alert settings
type AlertingConfig struct {
	Enabled       bool          `yaml:"enabled"`
	CriticalWebhooks []string   `yaml:"critical_webhooks"`
	HighWebhooks    []string   `yaml:"high_webhooks"`
	EmailEnabled    bool        `yaml:"email_enabled"`
	SMTPConfig      SMTPConfig  `yaml:"smtp_config"`
	RetentionDays   int         `yaml:"retention_days"`
}

// SMTPConfig contains SMTP settings for email alerts
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level         string            `yaml:"level"`
	Format        string            `yaml:"format"`
	Output        string            `yaml:"output"`
	IncludeCaller bool              `yaml:"include_caller"`
	Fields        map[string]string `yaml:"fields"`
}

// MetricsConfig contains metrics settings
type MetricsConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	Port     int    `yaml:"port"`
	Prefix   string `yaml:"prefix"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	TLS       TLSServerConfig   `yaml:"tls"`
	RateLimit RateLimitConfig   `yaml:"rate_limit"`
	CORS      CORSConfig        `yaml:"cors"`
	Auth      AuthConfig        `yaml:"auth"`
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

// AuthConfig contains authentication settings
type AuthConfig struct {
	Enabled      bool   `yaml:"enabled"`
	JWTSecret    string `yaml:"jwt_secret"`
	TokenExpiry  string `yaml:"token_expiry"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
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

	// Neo4j overrides
	if v := os.Getenv("NEO4J_URI"); v != "" {
		cfg.Neo4j.URI = v
	}
	if v := os.Getenv("NEO4J_PASSWORD"); v != "" {
		cfg.Neo4j.Password = v
	}

	// Redis overrides
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}

	// Blockchain overrides
	if v := os.Getenv("BTC_RPC_HOST"); v != "" {
		cfg.Blockchain.Bitcoin.RPCHost = v
	}
	if v := os.Getenv("ETH_RPC_ENDPOINT"); v != "" {
		cfg.Blockchain.Ethereum.RPCEndpoint = v
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
