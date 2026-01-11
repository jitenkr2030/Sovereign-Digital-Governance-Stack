package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Database wraps the sql.DB connection with additional functionality
type Database struct {
	*sql.DB
}

// NewDatabase creates a new Database connection
func NewDatabase(config interface{}) (*Database, error) {
	// Type assertion to get config values
	cfg, ok := config.(struct {
		Host            string
		Port            int
		Username        string
		Password        string
		Name            string
		SSLMode         string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime int
	})
	if !ok {
		return nil, fmt.Errorf("invalid database config type")
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{db}, nil
}

// RunMigrations applies database migrations
func RunMigrations(db *Database) error {
	migrations := []string{
		// Nodes table
		`CREATE TABLE IF NOT EXISTS nodes (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			network VARCHAR(100) NOT NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'full',
			rpc_url VARCHAR(500) NOT NULL,
			ws_url VARCHAR(500),
			status VARCHAR(50) NOT NULL DEFAULT 'unknown',
			version VARCHAR(100),
			chain_id BIGINT,
			block_number BIGINT DEFAULT 0,
			peer_count INT DEFAULT 0,
			last_sync_time TIMESTAMP,
			last_health_check TIMESTAMP,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		// Node metrics table
		`CREATE TABLE IF NOT EXISTS node_metrics (
			id VARCHAR(36) PRIMARY KEY,
			node_id VARCHAR(36) NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
			timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
			block_height BIGINT,
			block_time DOUBLE PRECISION,
			peer_count INT,
			transaction_pool_size INT,
			gas_price DOUBLE PRECISION,
			memory_usage BIGINT,
			cpu_usage DOUBLE PRECISION,
			disk_usage BIGINT,
			network_in BIGINT,
			network_out BIGINT,
			rpc_latency DOUBLE PRECISION,
			sync_progress DOUBLE PRECISION
		)`,

		// Node events table
		`CREATE TABLE IF NOT EXISTS node_events (
			id VARCHAR(36) PRIMARY KEY,
			node_id VARCHAR(36) NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
			event_type VARCHAR(100) NOT NULL,
			severity VARCHAR(20) NOT NULL DEFAULT 'info',
			message TEXT,
			metadata JSONB DEFAULT '{}',
			timestamp TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_nodes_network ON nodes(network)`,
		`CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status)`,
		`CREATE INDEX IF NOT EXISTS idx_node_metrics_node_id ON node_metrics(node_id)`,
		`CREATE INDEX IF NOT EXISTS idx_node_metrics_timestamp ON node_metrics(timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_node_events_node_id ON node_events(node_id)`,
		`CREATE INDEX IF NOT EXISTS idx_node_events_timestamp ON node_events(timestamp DESC)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
