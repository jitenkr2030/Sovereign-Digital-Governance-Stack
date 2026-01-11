package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/csic-platform/services/exchange-ingestion/internal/core/domain"
	"github.com/csic-platform/services/exchange-ingestion/internal/core/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	_ "github.com/lib/pq"
)

// PostgresDataSourceRepository implements DataSourceRepository using PostgreSQL
type PostgresDataSourceRepository struct {
	db     *sql.DB
	logger interface{ Debug(msg string) }
}

// NewPostgresDataSourceRepository creates a new PostgresDataSourceRepository
func NewPostgresDataSourceRepository(db *sql.DB) *PostgresDataSourceRepository {
	return &PostgresDataSourceRepository{
		db: db,
	}
}

// FindByID retrieves a data source configuration by its unique identifier
func (r *PostgresDataSourceRepository) FindByID(ctx context.Context, id string) (*domain.DataSourceConfig, error) {
	query := `
		SELECT id, name, exchange_type, endpoint, api_key, api_secret, symbols,
		       enabled, polling_rate, ws_enabled, auth_type, timeout, retry_attempts,
		       created_at, updated_at
		FROM data_sources
		WHERE id = $1
	`

	var config domain.DataSourceConfig
	var apiKey, apiSecret sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&config.ID,
		&config.Name,
		&config.ExchangeType,
		&config.Endpoint,
		&apiKey,
		&apiSecret,
		&config.Symbols,
		&config.Enabled,
		&config.PollingRate,
		&config.WSEnabled,
		&config.AuthType,
		&config.Timeout,
		&config.RetryAttempts,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrDataSourceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find data source: %w", err)
	}

	if apiKey.Valid {
		config.APIKey = apiKey.String
	}
	if apiSecret.Valid {
		config.APISecret = apiSecret.String
	}

	return &config, nil
}

// FindByName retrieves a data source configuration by its name
func (r *PostgresDataSourceRepository) FindByName(ctx context.Context, name string) (*domain.DataSourceConfig, error) {
	query := `
		SELECT id, name, exchange_type, endpoint, api_key, api_secret, symbols,
		       enabled, polling_rate, ws_enabled, auth_type, timeout, retry_attempts,
		       created_at, updated_at
		FROM data_sources
		WHERE name = $1
	`

	var config domain.DataSourceConfig
	var apiKey, apiSecret sql.NullString

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&config.ID,
		&config.Name,
		&config.ExchangeType,
		&config.Endpoint,
		&apiKey,
		&apiSecret,
		&config.Symbols,
		&config.Enabled,
		&config.PollingRate,
		&config.WSEnabled,
		&config.AuthType,
		&config.Timeout,
		&config.RetryAttempts,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find data source by name: %w", err)
	}

	if apiKey.Valid {
		config.APIKey = apiKey.String
	}
	if apiSecret.Valid {
		config.APISecret = apiSecret.String
	}

	return &config, nil
}

// FindAll retrieves all data source configurations
func (r *PostgresDataSourceRepository) FindAll(ctx context.Context) ([]*domain.DataSourceConfig, error) {
	query := `
		SELECT id, name, exchange_type, endpoint, api_key, api_secret, symbols,
		       enabled, polling_rate, ws_enabled, auth_type, timeout, retry_attempts,
		       created_at, updated_at
		FROM data_sources
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query data sources: %w", err)
	}
	defer rows.Close()

	var configs []*domain.DataSourceConfig
	for rows.Next() {
		var config domain.DataSourceConfig
		var apiKey, apiSecret sql.NullString

		if err := rows.Scan(
			&config.ID,
			&config.Name,
			&config.ExchangeType,
			&config.Endpoint,
			&apiKey,
			&apiSecret,
			&config.Symbols,
			&config.Enabled,
			&config.PollingRate,
			&config.WSEnabled,
			&config.AuthType,
			&config.Timeout,
			&config.RetryAttempts,
			&config.CreatedAt,
			&config.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan data source: %w", err)
		}

		if apiKey.Valid {
			config.APIKey = apiKey.String
		}
		if apiSecret.Valid {
			config.APISecret = apiSecret.String
		}

		configs = append(configs, &config)
	}

	return configs, nil
}

// FindEnabled retrieves all enabled data source configurations
func (r *PostgresDataSourceRepository) FindEnabled(ctx context.Context) ([]*domain.DataSourceConfig, error) {
	query := `
		SELECT id, name, exchange_type, endpoint, api_key, api_secret, symbols,
		       enabled, polling_rate, ws_enabled, auth_type, timeout, retry_attempts,
		       created_at, updated_at
		FROM data_sources
		WHERE enabled = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled data sources: %w", err)
	}
	defer rows.Close()

	var configs []*domain.DataSourceConfig
	for rows.Next() {
		var config domain.DataSourceConfig
		var apiKey, apiSecret sql.NullString

		if err := rows.Scan(
			&config.ID,
			&config.Name,
			&config.ExchangeType,
			&config.Endpoint,
			&apiKey,
			&apiSecret,
			&config.Symbols,
			&config.Enabled,
			&config.PollingRate,
			&config.WSEnabled,
			&config.AuthType,
			&config.Timeout,
			&config.RetryAttempts,
			&config.CreatedAt,
			&config.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan data source: %w", err)
		}

		if apiKey.Valid {
			config.APIKey = apiKey.String
		}
		if apiSecret.Valid {
			config.APISecret = apiSecret.String
		}

		configs = append(configs, &config)
	}

	return configs, nil
}

// Save creates or updates a data source configuration
func (r *PostgresDataSourceRepository) Save(ctx context.Context, config *domain.DataSourceConfig) error {
	query := `
		INSERT INTO data_sources (id, name, exchange_type, endpoint, api_key, api_secret, symbols,
		                          enabled, polling_rate, ws_enabled, auth_type, timeout, retry_attempts,
		                          created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			exchange_type = EXCLUDED.exchange_type,
			endpoint = EXCLUDED.endpoint,
			api_key = EXCLUDED.api_key,
			api_secret = EXCLUDED.api_secret,
			symbols = EXCLUDED.symbols,
			enabled = EXCLUDED.enabled,
			polling_rate = EXCLUDED.polling_rate,
			ws_enabled = EXCLUDED.ws_enabled,
			auth_type = EXCLUDED.auth_type,
			timeout = EXCLUDED.timeout,
			retry_attempts = EXCLUDED.retry_attempts,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		config.ID,
		config.Name,
		config.ExchangeType,
		config.Endpoint,
		config.APIKey,
		config.APISecret,
		config.Symbols,
		config.Enabled,
		config.PollingRate,
		config.WSEnabled,
		config.AuthType,
		config.Timeout,
		config.RetryAttempts,
		config.CreatedAt,
		config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save data source: %w", err)
	}

	return nil
}

// Delete removes a data source configuration
func (r *PostgresDataSourceRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM data_sources WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete data source: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrDataSourceNotFound
	}

	return nil
}

// UpdateStatus updates the connection status of a data source
func (r *PostgresDataSourceRepository) UpdateStatus(ctx context.Context, id string, status domain.ConnectionStatus) error {
	query := "UPDATE data_sources SET updated_at = $1 WHERE id = $2"

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update data source status: %w", err)
	}

	return nil
}

// PostgresMarketDataRepository implements MarketDataRepository using PostgreSQL
type PostgresMarketDataRepository struct {
	db *sql.DB
}

// NewPostgresMarketDataRepository creates a new PostgresMarketDataRepository
func NewPostgresMarketDataRepository(db *sql.DB) *PostgresMarketDataRepository {
	return &PostgresMarketDataRepository{db: db}
}

// Store saves market data to the repository
func (r *PostgresMarketDataRepository) Store(ctx context.Context, data *domain.MarketData) error {
	query := `
		INSERT INTO market_data (id, source_id, symbol, base_symbol, quote_symbol, price, volume,
		                        high_24h, low_24h, change_24h, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		data.ID,
		data.SourceID,
		data.Symbol,
		data.BaseSymbol,
		data.QuoteSymbol,
		data.Price,
		data.Volume,
		data.High24h,
		data.Low24h,
		data.Change24h,
		data.Timestamp,
		data.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store market data: %w", err)
	}

	return nil
}

// StoreBatch saves multiple market data records in a single operation
func (r *PostgresMarketDataRepository) StoreBatch(ctx context.Context, data []*domain.MarketData) error {
	if len(data) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO market_data (id, source_id, symbol, base_symbol, quote_symbol, price, volume,
		                        high_24h, low_24h, change_24h, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, d := range data {
		_, err := stmt.ExecContext(ctx,
			d.ID,
			d.SourceID,
			d.Symbol,
			d.BaseSymbol,
			d.QuoteSymbol,
			d.Price,
			d.Volume,
			d.High24h,
			d.Low24h,
			d.Change24h,
			d.Timestamp,
			d.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert market data: %w", err)
		}
	}

	return tx.Commit()
}

// FindLatest retrieves the latest market data for a specific symbol and source
func (r *PostgresMarketDataRepository) FindLatest(ctx context.Context, sourceID, symbol string) (*domain.MarketData, error) {
	query := `
		SELECT id, source_id, symbol, base_symbol, quote_symbol, price, volume,
		       high_24h, low_24h, change_24h, timestamp, created_at
		FROM market_data
		WHERE source_id = $1 AND symbol = $2
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var data domain.MarketData
	err := r.db.QueryRowContext(ctx, query, sourceID, symbol).Scan(
		&data.ID,
		&data.SourceID,
		&data.Symbol,
		&data.BaseSymbol,
		&data.QuoteSymbol,
		&data.Price,
		&data.Volume,
		&data.High24h,
		&data.Low24h,
		&data.Change24h,
		&data.Timestamp,
		&data.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find latest market data: %w", err)
	}

	return &data, nil
}

// FindByTimeRange retrieves market data within a specified time range
func (r *PostgresMarketDataRepository) FindByTimeRange(
	ctx context.Context,
	sourceID, symbol string,
	start, end time.Time,
) ([]*domain.MarketData, error) {
	query := `
		SELECT id, source_id, symbol, base_symbol, quote_symbol, price, volume,
		       high_24h, low_24h, change_24h, timestamp, created_at
		FROM market_data
		WHERE source_id = $1 AND symbol = $2 AND timestamp >= $3 AND timestamp <= $4
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, sourceID, symbol, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query market data: %w", err)
	}
	defer rows.Close()

	var dataList []*domain.MarketData
	for rows.Next() {
		var data domain.MarketData
		if err := rows.Scan(
			&data.ID,
			&data.SourceID,
			&data.Symbol,
			&data.BaseSymbol,
			&data.QuoteSymbol,
			&data.Price,
			&data.Volume,
			&data.High24h,
			&data.Low24h,
			&data.Change24h,
			&data.Timestamp,
			&data.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan market data: %w", err)
		}
		dataList = append(dataList, &data)
	}

	return dataList, nil
}

// FindLatestBySymbol retrieves the latest market data for a symbol across all sources
func (r *PostgresMarketDataRepository) FindLatestBySymbol(ctx context.Context, symbol string) ([]*domain.MarketData, error) {
	query := `
		SELECT DISTINCT ON (source_id) id, source_id, symbol, base_symbol, quote_symbol, price, volume,
		       high_24h, low_24h, change_24h, timestamp, created_at
		FROM market_data
		WHERE symbol = $1
		ORDER BY source_id, timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to query market data: %w", err)
	}
	defer rows.Close()

	var dataList []*domain.MarketData
	for rows.Next() {
		var data domain.MarketData
		if err := rows.Scan(
			&data.ID,
			&data.SourceID,
			&data.Symbol,
			&data.BaseSymbol,
			&data.QuoteSymbol,
			&data.Price,
			&data.Volume,
			&data.High24h,
			&data.Low24h,
			&data.Change24h,
			&data.Timestamp,
			&data.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan market data: %w", err)
		}
		dataList = append(dataList, &data)
	}

	return dataList, nil
}

// PostgresIngestionStatsRepository implements IngestionStatsRepository using PostgreSQL
type PostgresIngestionStatsRepository struct {
	db *sql.DB
}

// NewPostgresIngestionStatsRepository creates a new PostgresIngestionStatsRepository
func NewPostgresIngestionStatsRepository(db *sql.DB) *PostgresIngestionStatsRepository {
	return &PostgresIngestionStatsRepository{db: db}
}

// GetStats retrieves current statistics for a data source
func (r *PostgresIngestionStatsRepository) GetStats(ctx context.Context, sourceID string) (*domain.IngestionStats, error) {
	query := `
		SELECT source_id, total_messages, processed_messages, failed_messages,
		       last_message_time, average_latency, last_error, status
		FROM ingestion_stats
		WHERE source_id = $1
	`

	var stats domain.IngestionStats
	var lastError sql.NullString

	err := r.db.QueryRowContext(ctx, query, sourceID).Scan(
		&stats.SourceID,
		&stats.TotalMessages,
		&stats.ProcessedMessages,
		&stats.FailedMessages,
		&stats.LastMessageTime,
		&stats.AverageLatency,
		&lastError,
		&stats.Status,
	)

	if err == sql.ErrNoRows {
		// Return default stats for new source
		return &domain.IngestionStats{
			SourceID: sourceID,
			Status:   domain.ConnectionStatusDisconnected,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	if lastError.Valid {
		stats.LastError = lastError.String
	}

	return &stats, nil
}

// UpdateStats updates statistics for a data source
func (r *PostgresIngestionStatsRepository) UpdateStats(ctx context.Context, stats *domain.IngestionStats) error {
	query := `
		INSERT INTO ingestion_stats (source_id, total_messages, processed_messages, failed_messages,
		                            last_message_time, average_latency, last_error, status, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (source_id) DO UPDATE SET
			total_messages = EXCLUDED.total_messages,
			processed_messages = EXCLUDED.processed_messages,
			failed_messages = EXCLUDED.failed_messages,
			last_message_time = EXCLUDED.last_message_time,
			average_latency = EXCLUDED.average_latency,
			last_error = EXCLUDED.last_error,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		stats.SourceID,
		stats.TotalMessages,
		stats.ProcessedMessages,
		stats.FailedMessages,
		stats.LastMessageTime,
		stats.AverageLatency,
		stats.LastError,
		stats.Status,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update stats: %w", err)
	}

	return nil
}

// IncrementMessages increments message counters
func (r *PostgresIngestionStatsRepository) IncrementMessages(ctx context.Context, sourceID string, processed, failed int64) error {
	query := `
		UPDATE ingestion_stats SET
			total_messages = total_messages + $2,
			processed_messages = processed_messages + $3,
			failed_messages = failed_messages + $4,
			updated_at = $5
		WHERE source_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, sourceID, processed+failed, processed, failed, time.Now())
	if err != nil {
		return fmt.Errorf("failed to increment messages: %w", err)
	}

	return nil
}

// RecordLatency records the processing latency
func (r *PostgresIngestionStatsRepository) RecordLatency(ctx context.Context, sourceID string, latency time.Duration) error {
	query := `
		UPDATE ingestion_stats SET
			average_latency = ($2::interval + average_latency * processed_messages) / (processed_messages + 1),
			updated_at = $3
		WHERE source_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, sourceID, latency.String(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to record latency: %w", err)
	}

	return nil
}

// Ensure ports interfaces are implemented
var _ ports.DataSourceRepository = (*PostgresDataSourceRepository)(nil)
var _ ports.MarketDataRepository = (*PostgresMarketDataRepository)(nil)
var _ ports.IngestionStatsRepository = (*PostgresIngestionStatsRepository)(nil)
