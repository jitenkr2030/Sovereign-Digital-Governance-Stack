package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic-platform/internal/adapters/handler/http"
	"github.com/csic-platform/internal/adapters/repository/postgres"
	"github.com/csic-platform/internal/core/ports"
	"github.com/csic-platform/internal/core/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"github.com/lib/pq"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	App           AppConfig           `yaml:"app"`
	Database      DatabaseConfig      `yaml:"database"`
	Redis         RedisConfig         `yaml:"redis"`
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	Graph         GraphConfig         `yaml:"graph"`
	Evidence      EvidenceConfig      `yaml:"evidence"`
	Security      SecurityConfig      `yaml:"security"`
	Health        HealthConfig        `yaml:"health"`
}

// AppConfig holds application-level settings
type AppConfig struct {
	Name        string `yaml:"name"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Environment string `yaml:"environment"`
	LogLevel    string `yaml:"log_level"`
}

// DatabaseConfig holds PostgreSQL connection settings
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	TablePrefix     string `yaml:"table_prefix"`
	SSLMode         string `yaml:"ssl_mode"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime"`
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Password  string `yaml:"password"`
	DB        int    `yaml:"db"`
	KeyPrefix string `yaml:"key_prefix"`
	PoolSize  int    `yaml:"pool_size"`
}

// ElasticsearchConfig holds Elasticsearch settings
type ElasticsearchConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Index     string `yaml:"index"`
	SessionID string `yaml:"session_id"`
}

// GraphConfig holds graph analysis settings
type GraphConfig struct {
	MaxDepth                    int `yaml:"max_depth"`
	MaxNodes                    int `yaml:"max_nodes"`
	CommunityDetectionLimit     int `yaml:"community_detection_limit"`
	CentralityCalculationLimit  int `yaml:"centrality_calculation_limit"`
}

// EvidenceConfig holds evidence management settings
type EvidenceConfig struct {
	RetentionDays            int    `yaml:"retention_days"`
	MaxEvidenceSizeMB        int    `yaml:"max_evidence_size_mb"`
	HashAlgorithm            string `yaml:"hash_algorithm"`
	ChainOfCustodyEnabled    bool   `yaml:"chain_of_custody_enabled"`
}

// SecurityConfig holds security settings
type SecurityConfig struct {
	APIKeys      []string          `yaml:"api_keys"`
	RateLimit    RateLimitConfig   `yaml:"rate_limit"`
}

// RateLimitConfig holds rate limiting settings
type RateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	Burst             int `yaml:"burst"`
}

// HealthConfig holds health check settings
type HealthConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	Interval string `yaml:"interval"`
	Timeout  string `yaml:"timeout"`
}

func main() {
	// Load configuration
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := httplog.NewLogger(config.App.Name, httplog.Options{
		JSONFields: map[string]string{
			"service": config.App.Name,
			"env":     config.App.Environment,
		},
	})

	// Initialize database connection
	db, err := initDatabase(config.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Initialize repositories
	repo := postgres.NewPostgresRepository(db)

	// Initialize services
	graphService := services.NewGraphAnalysisService(repo)
	evidenceService := services.NewEvidenceManagementService(repo, repo)

	// Initialize HTTP handler
	handler := http.NewForensicHandler(graphService, evidenceService)

	// Setup router
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	// Add health check endpoint
	if config.Health.Enabled {
		router.Get(config.Health.Endpoint, healthHandler)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.App.Host, config.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info().Msgf("Starting %s service on %s:%d", config.App.Name, config.App.Host, config.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited")
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(&config)

	return &config, nil
}

func applyEnvOverrides(config *Config) {
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("APP_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &config.App.Port)
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.App.LogLevel = logLevel
	}
}

func initDatabase(cfg DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	connMaxLifetime, err := time.ParseDuration(cfg.ConnMaxLifetime)
	if err != nil {
		connMaxLifetime = 5 * time.Minute
	}
	db.SetConnMaxLifetime(connMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db, cfg.TablePrefix); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB, tablePrefix string) error {
	migrations := []string{
		// Graph nodes table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %sgraph_nodes (
				id VARCHAR(255) PRIMARY KEY,
				node_type VARCHAR(50) NOT NULL,
				label TEXT,
				metadata JSONB,
				created_at BIGINT,
				updated_at BIGINT
			)
		`, tablePrefix),
		// Graph edges table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %sgraph_edges (
				id VARCHAR(255) PRIMARY KEY,
				source_id VARCHAR(255) NOT NULL,
				target_id VARCHAR(255) NOT NULL,
				edge_type VARCHAR(50) NOT NULL,
				weight DECIMAL(10, 2),
				metadata JSONB,
				created_at BIGINT,
				FOREIGN KEY (source_id) REFERENCES %sgraph_nodes(id) ON DELETE CASCADE,
				FOREIGN KEY (target_id) REFERENCES %sgraph_nodes(id) ON DELETE CASCADE
			)
		`, tablePrefix, tablePrefix, tablePrefix),
		// Evidence table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %sevidence (
				id VARCHAR(255) PRIMARY KEY,
				case_id VARCHAR(255),
				evidence_type VARCHAR(50) NOT NULL,
				title TEXT,
				description TEXT,
				content JSONB,
				tags JSONB,
				hash VARCHAR(255),
				collected_at BIGINT,
				collected_by VARCHAR(255),
				status VARCHAR(50) DEFAULT 'collected'
			)
		`, tablePrefix),
		// Chain of custody table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %schain_of_custody (
				id SERIAL PRIMARY KEY,
				evidence_id VARCHAR(255) NOT NULL,
				timestamp BIGINT NOT NULL,
				handler VARCHAR(255) NOT NULL,
				action VARCHAR(255) NOT NULL,
				location VARCHAR(255),
				integrity_hash VARCHAR(255),
				notes TEXT,
				FOREIGN KEY (evidence_id) REFERENCES %sevidence(id) ON DELETE CASCADE
			)
		`, tablePrefix, tablePrefix),
		// Forensic reports table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %sforensic_reports (
				id VARCHAR(255) PRIMARY KEY,
				case_id VARCHAR(255),
				report_type VARCHAR(50) NOT NULL,
				title TEXT,
				description TEXT,
				content JSONB,
				hash VARCHAR(255),
				generated_at BIGINT,
				generated_by VARCHAR(255),
				status VARCHAR(50) DEFAULT 'generated',
				version VARCHAR(50) DEFAULT '1.0'
			)
		`, tablePrefix),
		// Clusters table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %sclusters (
				id VARCHAR(255) PRIMARY KEY,
				name VARCHAR(255),
				member_ids JSONB,
				score DECIMAL(10, 4),
				metadata JSONB,
				created_at BIGINT
			)
		`, tablePrefix),
		// Create indexes
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %sidx_graph_nodes_type ON %sgraph_nodes(node_type)`, tablePrefix, tablePrefix),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %sidx_graph_edges_source ON %sgraph_edges(source_id)`, tablePrefix, tablePrefix),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %sidx_graph_edges_target ON %sgraph_edges(target_id)`, tablePrefix, tablePrefix),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %sidx_evidence_case ON %sevidence(case_id)`, tablePrefix, tablePrefix),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %sidx_evidence_type ON %sevidence(evidence_type)`, tablePrefix, tablePrefix),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %sidx_custody_evidence ON %schain_of_custody(evidence_id)`, tablePrefix, tablePrefix),
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			// Check if it's a duplicate key error (already exists)
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "42710" {
				continue
			}
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "forensic-tools"}`))
}

// Ensure repository implements all interfaces
var _ ports.GraphRepository = (*postgres.PostgresRepository)(nil)
var _ ports.EvidenceRepository = (*postgres.PostgresRepository)(nil)
var _ ports.ReportRepository = (*postgres.PostgresRepository)(nil)
var _ ports.ClusterRepository = (*postgres.PostgresRepository)(nil)
