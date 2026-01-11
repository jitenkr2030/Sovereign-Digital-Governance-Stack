package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Conn defines the interface for database operations
type Conn interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgxconn, error)
	Query(ctx context.Context, sql string, arguments ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...interface{}) RowScanner
}

// pgxconn wraps pgconn.CommandTag to provide RowsAffected method
type pgxconn interface {
	RowsAffected() int64
}

// RowScanner is an interface for scanning rows
type RowScanner interface {
	Scan(dest ...interface{}) error
}

// Rows is an interface for iterating over rows
type Rows interface {
	Close()
	Next() bool
	Scan(dest ...interface{}) error
}

// connection implements the Conn interface using pgxpool
type connection struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

// NewConnection creates a new database connection
func NewConnection(log *zap.Logger) (*connection, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	poolConfig, err := pgxpool.ParseConfig(config.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = int32(config.MaxOpenConns)
	poolConfig.MinConns = int32(config.MaxIdleConns)
	poolConfig.MaxConnLifetime = time.Duration(config.ConnMaxLifetime) * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Successfully connected to PostgreSQL database",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Name),
	)

	return &connection{
		pool: pool,
		log:  log,
	}, nil
}

// Exec executes a SQL command
func (c *connection) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgxconn, error) {
	result, err := c.pool.Exec(ctx, sql, arguments...)
	if err != nil {
		return nil, err
	}
	return &tagWrapper{tag: result}, nil
}

// Query executes a SQL query that returns rows
func (c *connection) Query(ctx context.Context, sql string, arguments ...interface{}) (Rows, error) {
	rows, err := c.pool.Query(ctx, sql, arguments...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// QueryRow executes a SQL query that returns a single row
func (c *connection) QueryRow(ctx context.Context, sql string, arguments ...interface{}) RowScanner {
	return c.pool.QueryRow(ctx, sql, arguments...)
}

// Close closes the database connection
func (c *connection) Close() {
	c.pool.Close()
	c.log.Info("Database connection pool closed")
}

// tagWrapper wraps pgconn.CommandTag
type tagWrapper struct {
	tag pgxconn
}

func (t *tagWrapper) RowsAffected() int64 {
	return t.tag.RowsAffected()
}
