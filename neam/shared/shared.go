// Shared Utilities Package
// Common utilities for the NEAM platform

package shared

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
)

// Config holds application configuration
type Config struct {
	// Application settings
	AppName      string `mapstructure:"app_name" json:"app_name"`
	Environment  string `mapstructure:"environment" json:"environment"`
	Debug        bool   `mapstructure:"debug" json:"debug"`
	
	// HTTP settings
	HTTPPort     int    `mapstructure:"http_port" json:"http_port"`
	ReadTimeout  int    `mapstructure:"read_timeout" json:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout" json:"write_timeout"`
	
	// ClickHouse settings
	ClickHouseDSN    string `mapstructure:"clickhouse_dsn" json:"clickhouse_dsn"`
	ClickHouseHost   string `mapstructure:"clickhouse_host" json:"clickhouse_host"`
	ClickHousePort   int    `mapstructure:"clickhouse_port" json:"clickhouse_port"`
	ClickHouseUser   string `mapstructure:"clickhouse_user" json:"clickhouse_user"`
	ClickHousePass   string `mapstructure:"clickhouse_password" json:"clickhouse_password"`
	ClickHouseDB     string `mapstructure:"clickhouse_database" json:"clickhouse_database"`
	
	// PostgreSQL settings
	PostgresDSN    string `mapstructure:"postgres_dsn" json:"postgres_dsn"`
	PostgresHost   string `mapstructure:"postgres_host" json:"postgres_host"`
	PostgresPort   int    `mapstructure:"postgres_port" json:"postgres_port"`
	PostgresUser   string `mapstructure:"postgres_user" json:"postgres_user"`
	PostgresPass   string `mapstructure:"postgres_password" json:"postgres_password"`
	PostgresDB     string `mapstructure:"postgres_database" json:"postgres_database"`
	
	// Redis settings
	RedisAddr     string `mapstructure:"redis_addr" json:"redis_addr"`
	RedisPassword string `mapstructure:"redis_password" json:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db" json:"redis_db"`
	
	// Kafka settings
	KafkaBrokers  string `mapstructure:"kafka_brokers" json:"kafka_brokers"`
	KafkaGroup    string `mapstructure:"kafka_group" json:"kafka_group"`
	
	// Neo4j settings
	Neo4jURI      string `mapstructure:"neo4j_uri" json:"neo4j_uri"`
	Neo4jUser     string `mapstructure:"neo4j_user" json:"neo4j_user"`
	Neo4jPass     string `mapstructure:"neo4j_password" json:"neo4j_password"`
	
	// Elasticsearch settings
	ElasticsearchAddrs string `mapstructure:"elasticsearch_addrs" json:"elasticsearch_addrs"`
	
	// MinIO settings
	MinIOEndpoint   string `mapstructure:"minio_endpoint" json:"minio_endpoint"`
	MinIOAccessKey  string `mapstructure:"minio_access_key" json:"minio_access_key"`
	MinIOSecretKey  string `mapstructure:"minio_secret_key" json:"minio_secret_key"`
	
	// Export settings
	ExportDir    string `mapstructure:"export_dir" json:"export_dir"`
	ArchiveDir   string `mapstructure:"archive_dir" json:"archive_dir"`

	mu sync.RWMutex
	loaded bool
}

// NewConfig creates a new configuration instance
func NewConfig() *Config {
	return &Config{
		AppName:      "neam-platform",
		Environment:  "development",
		Debug:        true,
		HTTPPort:     8080,
		ReadTimeout:  15,
		WriteTimeout: 15,
		ClickHouseDSN: "clickhouse://localhost:9000",
		PostgresDSN:  "postgres://localhost:5432",
		RedisAddr:    "localhost:6379",
		KafkaBrokers: "localhost:9092",
		Neo4jURI:     "bolt://localhost:7687",
		ElasticsearchAddrs: "http://localhost:9200",
		MinIOEndpoint: "localhost:9000",
		ExportDir:    "/var/lib/neam/exports",
		ArchiveDir:   "/var/lib/neam/archives",
	}
}

// Load loads configuration from file and environment
func (c *Config) Load(serviceName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.loaded {
		return nil
	}

	// Try to load from config file
	configPath := os.Getenv("NEAM_CONFIG_PATH")
	if configPath == "" {
		configPath = fmt.Sprintf("/etc/neam/%s.yaml", serviceName)
	}

	if _, err := os.Stat(configPath); err == nil {
		if err := c.loadFromFile(configPath); err != nil {
			log.Printf("Warning: Failed to load config from file: %v", err)
		}
	}

	// Override with environment variables
	c.loadFromEnv()

	c.loaded = true
	return nil
}

// loadFromFile loads configuration from YAML file
func (c *Config) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}

// loadFromEnv loads configuration from environment variables
func (c *Config) loadFromEnv() {
	// ClickHouse overrides
	if v := os.Getenv("CLICKHOUSE_HOST"); v != "" {
		c.ClickHouseHost = v
	}
	if v := os.Getenv("CLICKHOUSE_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &c.ClickHousePort)
	}
	if v := os.Getenv("CLICKHOUSE_USER"); v != "" {
		c.ClickHouseUser = v
	}
	if v := os.Getenv("CLICKHOUSE_PASSWORD"); v != "" {
		c.ClickHousePass = v
	}

	// PostgreSQL overrides
	if v := os.Getenv("POSTGRES_HOST"); v != "" {
		c.PostgresHost = v
	}
	if v := os.Getenv("POSTGRES_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &c.PostgresPort)
	}
	if v := os.Getenv("POSTGRES_USER"); v != "" {
		c.PostgresUser = v
	}
	if v := os.Getenv("POSTGRES_PASSWORD"); v != "" {
		c.PostgresPass = v
	}

	// Redis overrides
	if v := os.Getenv("REDIS_HOST"); v != "" {
		c.RedisAddr = v + ":6379"
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		c.RedisPassword = v
	}

	// Kafka overrides
	if v := os.Getenv("KAFKA_BROKERS"); v != "" {
		c.KafkaBrokers = v
	}
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() *Config {
	return NewConfig()
}

// Logger provides structured logging
type Logger struct {
	name   string
	fields map[string]interface{}
	mu     sync.Mutex
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      string `mapstructure:"level" json:"level"`
	Format     string `mapstructure:"format" json:"format"`
	OutputPath string `mapstructure:"output_path" json:"output_path"`
}

// NewLogger creates a new logger instance
func NewLogger(config *LoggerConfig) *Logger {
	name := "neam"
	if config != nil {
		name = config.OutputPath
	}
	return &Logger{
		name:   name,
		fields: make(map[string]interface{}),
	}
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newLogger := &Logger{
		name:   l.name,
		fields: make(map[string]interface{}),
	}
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	newLogger.fields[key] = value
	return newLogger
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log("INFO", msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log("ERROR", msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log("WARN", msg, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log("DEBUG", msg, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log("FATAL", msg, args...)
	os.Exit(1)
}

// log implements the logging logic
func (l *Logger) log(level, msg string, args ...interface{}) {
	fields := make(map[string]interface{})
	l.mu.Lock()
	for k, v := range l.fields {
		fields[k] = v
	}
	l.mu.Unlock()

	fields["level"] = level
	fields["service"] = l.name
	fields["timestamp"] = time.Now().Format(time.RFC3339)

	// Parse args as key-value pairs
	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			fields[key] = args[i+1]
		}
	}

	jsonFields, _ := json.Marshal(fields)
	fmt.Printf("[%s] %s %s\n", level, msg, string(jsonFields))
}

// PostgreSQL wrapper
type PostgreSQL struct {
	db        *sql.DB
	mu        sync.RWMutex
	connected bool
}

// NewPostgreSQL creates a new PostgreSQL connection
func NewPostgreSQL(dsn string) (*PostgreSQL, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &PostgreSQL{db: db, connected: true}, nil
}

// Close closes the database connection
func (p *PostgreSQL) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.connected {
		p.connected = false
		return p.db.Close()
	}
	return nil
}

// Query executes a query that returns rows
func (p *PostgreSQL) Query(ctx context.Context, query string, dest interface{}) error {
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil // Simplified - in production, scan results to dest
}

// Execute executes a query without returning results
func (p *PostgreSQL) Execute(ctx context.Context, query string, args ...interface{}) error {
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

// ClickHouse wrapper
type ClickHouse struct {
	conn clickhouse.Conn
	mu   sync.RWMutex
}

// NewClickHouse creates a new ClickHouse connection
func NewClickHouse(dsn string) (ClickHouse, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{dsn},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
		TLS: &tls.Config{},
	})
	if err != nil {
		return ClickHouse{}, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return ClickHouse{}, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return ClickHouse{conn: conn}, nil
}

// Close closes the connection
func (c *ClickHouse) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
	}
}

// Query executes a query
func (c *ClickHouse) Query(ctx context.Context, query string, dest interface{}) error {
	rows, err := c.conn.Query(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

// QueryRow executes a query that returns a single row
func (c *ClickHouse) QueryRow(ctx context.Context, query string, dest ...interface{}) error {
	return c.conn.QueryRow(ctx, query).Scan(dest...)
}

// Exec executes a query without returning results
func (c *ClickHouse) Exec(ctx context.Context, query string, args ...interface{}) error {
	return c.conn.Exec(ctx, query, args...)
}

// Redis wrapper
type Redis struct {
	client *redis.Client
}

// NewRedis creates a new Redis connection
func NewRedis(addr, password string) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Redis{client: client}, nil
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	return r.client.Close()
}

// Client returns the underlying Redis client
func (r *Redis) Client() *redis.Client {
	return r.client
}

// Neo4j wrapper
type Neo4j struct {
	driver neo4j.Driver
}

// NewNeo4j creates a new Neo4j connection
func NewNeo4j(uri, user, password string) (*Neo4j, error) {
	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	if err := driver.VerifyConnectivity(); err != nil {
		return nil, fmt.Errorf("failed to verify Neo4j connectivity: %w", err)
	}

	return &Neo4j{driver: driver}, nil
}

// Close closes the Neo4j connection
func (n *Neo4j) Close() error {
	return n.driver.Close()
}

// Driver returns the underlying Neo4j driver
func (n *Neo4j) Driver() neo4j.Driver {
	return n.driver
}

// Elasticsearch wrapper (simplified)
type Elasticsearch struct {
	addrs []string
}

// NewElasticsearch creates a new Elasticsearch connection
func NewElasticsearch(addrs string) (*Elasticsearch, error) {
	return &Elasticsearch{addrs: []string{addrs}}, nil
}

// MinIO wrapper (simplified)
type MinIO struct {
	endpoint   string
	accessKey  string
	secretKey  string
}

// NewMinIO creates a new MinIO connection
func NewMinIO(endpoint, accessKey, secretKey string) (*MinIO, error) {
	return &MinIO{
		endpoint:  endpoint,
		accessKey: accessKey,
		secretKey: secretKey,
	}, nil
}

// UploadFile uploads a file to MinIO
func (m *MinIO) UploadFile(bucket, objectName, filePath string) error {
	return nil // Simplified
}

// GetPresignedURL generates a presigned URL
func (m *MinIO) GetPresignedURL(bucket, objectName string, expiration time.Duration) (string, error) {
	return "", nil // Simplified
}

// Kafka producer wrapper (simplified)
type KafkaProducer struct {
	brokers []string
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers string) (*KafkaProducer, error) {
	return &KafkaProducer{brokers: []string{brokers}}, nil
}

// Close closes the producer
func (k *KafkaProducer) Close() error {
	return nil
}

// Publish publishes a message to a topic
func (k *KafkaProducer) Publish(topic string, message interface{}) error {
	return nil // Simplified
}

// Kafka consumer wrapper (simplified)
type KafkaConsumer struct {
	brokers []string
	group   string
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(brokers, group string) (*KafkaConsumer, error) {
	return &KafkaConsumer{brokers: []string{brokers}, group: group}, nil
}

// Close closes the consumer
func (k *KafkaConsumer) Close() error {
	return nil
}

// Middleware provides common HTTP middleware
type Middleware struct{}

// NewMiddleware creates a new middleware instance
func NewMiddleware() *Middleware {
	return &Middleware{}
}

// Recovery returns a recovery middleware
func (m *Middleware) Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// Logging returns a logging middleware
func (m *Middleware) Logging() gin.HandlerFunc {
	return gin.Logger()
}

// CORS returns a CORS middleware
func (m *Middleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimit returns a rate limiting middleware
func (m *Middleware) RateLimit(requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simplified rate limiting
		c.Next()
	}
}

// Authenticate returns an authentication middleware
func (m *Middleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simplified authentication
		c.Next()
	}
}

// EconomicIndicator represents a standardized economic metric
type EconomicIndicator struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Region      string                 `json:"region"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Anomaly represents a detected economic anomaly
type Anomaly struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Region      string                 `json:"region"`
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Resolved    bool                   `json:"resolved"`
}

// Policy represents an intervention policy
type Policy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Rules       []PolicyRule           `json:"rules"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PolicyRule represents a single rule within a policy
type PolicyRule struct {
	ID          string                 `json:"id"`
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
}

// Intervention represents a policy intervention
type Intervention struct {
	ID          string                 `json:"id"`
	PolicyID    string                 `json:"policy_id"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Region      string                 `json:"region"`
	Parameters  map[string]interface{} `json:"parameters"`
	ExecutedAt  time.Time              `json:"executed_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
}

// Region represents a geographic region
type Region struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Code        string     `json:"code"`
	ParentID    string     `json:"parent_id,omitempty"`
	Coordinates []float64  `json:"coordinates"`
}

// TimeSeriesData represents a time series dataset
type TimeSeriesData struct {
	Metric     string          `json:"metric"`
	Region     string          `json:"region"`
	Frequency  string          `json:"frequency"`
	DataPoints []DataPoint     `json:"data_points"`
}

// DataPoint represents a single data point in a time series
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// ExportRequest represents a request to export data
type ExportRequest struct {
	Type      string          `json:"type"`
	Format    string          `json:"format"`
	Period    TimePeriod      `json:"period"`
	Regions   []string        `json:"regions"`
	Metrics   []string        `json:"metrics"`
	Options   map[string]bool `json:"options"`
}

// TimePeriod represents a time period
type TimePeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// AuditRecord represents an audit log entry
type AuditRecord struct {
	ID          string                 `json:"id"`
	Actor       string                 `json:"actor"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id"`
	Details     map[string]interface{} `json:"details"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Timestamp   time.Time              `json:"timestamp"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	ChainHash   string                 `json:"chain_hash,omitempty"`
}

// OpenSearch connection placeholder
type OpenSearch struct {
	connected bool
	mu        sync.Mutex
}

// NewOpenSearch creates a new OpenSearch connection
func NewOpenSearch(url string) (*OpenSearch, error) {
	return &OpenSearch{connected: true}, nil
}

// Close closes the connection
func (o *OpenSearch) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.connected = false
	return nil
}
