// Database Package - PostgreSQL connection and utilities for CSIC Platform
// Connection pooling, migrations, and query helpers

package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Config holds database configuration
type Config struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	MaxConnRetries  int
	RetryDelay      time.Duration
}

// DB wraps the sql.DB with CSIC-specific functionality
type DB struct {
	*sql.DB
	logger  *zap.Logger
	config  Config
}

// NewDB creates a new database connection
func NewDB(cfg Config, logger *zap.Logger) (*DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		DB:      db,
		logger:  logger,
		config:  cfg,
	}, nil
}

// HealthCheck verifies the database connection is healthy
func (db *DB) HealthCheck(ctx context.Context) error {
	return db.PingContext(ctx)
}

// QueryRow executes a query that returns a single row
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRow(query, args...)
}

// Query executes a query that returns multiple rows
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.Query(query, args...)
}

// Exec executes a query without returning results
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.DB.Exec(query, args...)
}

// Transaction executes a function within a transaction
func (db *DB) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithTransaction returns a transaction context
func (db *DB) WithTransaction(ctx context.Context) (context.Context, *sql.Tx, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return context.WithValue(ctx, "tx", tx), tx, nil
}

// QueryWithContext executes a query with context
func (db *DB) QueryWithContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, query, args...)
}

// ExecWithContext executes a query with context
func (db *DB) ExecWithContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

// Prepared statements cache
type PreparedStatement struct {
	Stmt  *sql.Stmt
	Query string
}

type StatementCache struct {
	mu      sync.Mutex
	stmts   map[string]*sql.Stmt
	db      *DB
}

func NewStatementCache(db *DB) *StatementCache {
	return &StatementCache{
		stmts: make(map[string]*sql.Stmt),
		db:    db,
	}
}

// Prepare creates or retrieves a cached prepared statement
func (c *StatementCache) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if stmt, exists := c.stmts[query]; exists {
		return stmt, nil
	}

	stmt, err := c.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	c.stmts[query] = stmt
	return stmt, nil
}

// Close closes all prepared statements
func (c *StatementCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, stmt := range c.stmts {
		if err := stmt.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Migration represents a database migration
type Migration struct {
	ID          string
	Description string
	Up          string
	Down        string
	AppliedAt   time.Time
}

// Migrate runs migrations
func (db *DB) Migrate(ctx context.Context, migrations []Migration) error {
	// Create migrations table if not exists
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id VARCHAR(255) PRIMARY KEY,
			description TEXT,
			up_sql TEXT NOT NULL,
			down_sql TEXT,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.ExecWithContext(ctx, createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	applied := make(map[string]bool)
	rows, err := db.QueryWithContext(ctx, "SELECT id FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		rows.Scan(&id)
		applied[id] = true
	}

	// Apply pending migrations
	for _, m := range migrations {
		if applied[m.ID] {
			continue
		}

		db.logger.Info("Applying migration", zap.String("id", m.ID), zap.String("description", m.Description))

		if _, err := db.ExecWithContext(ctx, m.Up); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", m.ID, err)
		}

		// Record migration
		insert := `INSERT INTO schema_migrations (id, description, up_sql, down_sql) VALUES ($1, $2, $3, $4)`
		if _, err := db.ExecWithContext(ctx, insert, m.ID, m.Description, m.Up, m.Down); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", m.ID, err)
		}
	}

	return nil
}

// RowScanner provides a convenient interface for scanning rows
type RowScanner struct {
	rows *sql.Rows
	err  error
}

// Scan scans a row into the provided pointers
func (rs *RowScanner) Scan(dest ...interface{}) *RowScanner {
	if rs.err != nil {
		return rs
	}
	rs.err = rs.rows.Scan(dest...)
	return rs
}

// Err returns the first error encountered
func (rs *RowScanner) Err() error {
	return rs.err
}

// Query scans rows into a slice of structs
func (db *DB) QueryScan(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	rows, err := db.QueryWithContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanRows(rows, dest)
}

// scanRows scans sql.Rows into a slice of structs
func scanRows(rows *sql.Rows, dest interface{}) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return err
		}
	}

	return nil
}

// NullTime represents a nullable time value
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Scan implements the sql.Scanner interface
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}
	nt.Valid = true
	switch v := value.(type) {
	case time.Time:
		nt.Time = v
	default:
		return fmt.Errorf("cannot scan type %T into NullTime", value)
	}
	return nil
}

// NullString represents a nullable string value
type NullString struct {
	String string
	Valid  bool
}

// Scan implements the sql.Scanner interface
func (ns *NullString) Scan(value interface{}) error {
	if value == nil {
		ns.String, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	switch v := value.(type) {
	case string:
		ns.String = v
	default:
		return fmt.Errorf("cannot scan type %T into NullString", value)
	}
	return nil
}

// NullInt64 represents a nullable int64 value
type NullInt64 struct {
	Int64 int64
	Valid bool
}

// Scan implements the sql.Scanner interface
func (ni *NullInt64) Scan(value interface{}) error {
	if value == nil {
		ni.Int64, ni.Valid = 0, false
		return nil
	}
	ni.Valid = true
	switch v := value.(type) {
	case int64:
		ni.Int64 = v
	default:
		return fmt.Errorf("cannot scan type %T into NullInt64", value)
	}
	return nil
}

// ConnectionPoolStats returns statistics about the connection pool
func (db *DB) ConnectionPoolStats() ConnectionStats {
	stats := db.DB.Stats()
	return ConnectionStats{
		OpenConnections:   stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxConnections:    db.config.MaxOpenConns,
		MaxIdleConnection: db.config.MaxIdleConns,
	}
}

// ConnectionStats holds connection pool statistics
type ConnectionStats struct {
	OpenConnections   int
	InUse             int
	Idle              int
	WaitCount         int64
	WaitDuration      time.Duration
	MaxConnections    int
	MaxIdleConnection int
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Import required sync package
import "sync"
