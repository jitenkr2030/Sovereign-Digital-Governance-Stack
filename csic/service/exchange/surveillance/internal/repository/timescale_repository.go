package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/csic/surveillance/internal/domain"
	"github.com/csic/surveillance/internal/port"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	_ "github.com/lib/pq"
)

// TimescaleConfig holds TimescaleDB connection configuration
type TimescaleConfig struct {
	PostgresConfig
	ChunkInterval     string
	RetentionInterval string
}

// TimescaleMarketRepository implements MarketRepository for TimescaleDB
type TimescaleMarketRepository struct {
	db        *sql.DB
	chunkDays int
}

// NewTimescaleMarketRepository creates a new TimescaleDB market repository
func NewTimescaleMarketRepository(config PostgresConfig) (*TimescaleMarketRepository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Name, config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &TimescaleMarketRepository{
		db:        db,
		chunkDays: 7, // Default 7-day chunks
	}

	return repo, nil
}

// Close closes the database connection
func (r *TimescaleMarketRepository) Close() error {
	return r.db.Close()
}

// StoreEvent stores a single market event
func (r *TimescaleMarketRepository) StoreEvent(ctx context.Context, event *domain.MarketEvent) error {
	query := `INSERT INTO market_events (
		id, exchange_id, symbol, event_type, order_id, user_id, price, quantity,
		filled_quantity, direction, status, timestamp, received_at, sequence_num, raw_data, checksum
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	ON CONFLICT (id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.ExchangeID, event.Symbol, event.EventType, event.OrderID,
		event.UserID, event.Price, event.Quantity, event.FilledQuantity,
		event.Direction, event.Status, event.Timestamp, event.ReceivedAt,
		event.SequenceNum, event.RawData, event.Checksum,
	)

	if err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	return nil
}

// StoreEvents stores multiple market events
func (r *TimescaleMarketRepository) StoreEvents(ctx context.Context, events []domain.MarketEvent) error {
	return r.BulkStoreEvents(ctx, events)
}

// GetEventByID retrieves a market event by ID
func (r *TimescaleMarketRepository) GetEventByID(ctx context.Context, id uuid.UUID) (*domain.MarketEvent, error) {
	query := `SELECT id, exchange_id, symbol, event_type, order_id, user_id, price, quantity,
		filled_quantity, direction, status, timestamp, received_at, sequence_num, raw_data, checksum
		FROM market_events WHERE id = $1`

	event := &domain.MarketEvent{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID, &event.ExchangeID, &event.Symbol, &event.EventType, &event.OrderID,
		&event.UserID, &event.Price, &event.Quantity, &event.FilledQuantity,
		&event.Direction, &event.Status, &event.Timestamp, &event.ReceivedAt,
		&event.SequenceNum, &event.RawData, &event.Checksum,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return event, nil
}

// GetEventsBySymbol retrieves events for a symbol within a time range
func (r *TimescaleMarketRepository) GetEventsBySymbol(ctx context.Context, symbol string, startTime, endTime time.Time, limit int) ([]domain.MarketEvent, error) {
	query := `SELECT id, exchange_id, symbol, event_type, order_id, user_id, price, quantity,
		filled_quantity, direction, status, timestamp, received_at, sequence_num, raw_data, checksum
		FROM market_events WHERE symbol = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp DESC LIMIT $4`

	rows, err := r.db.QueryContext(ctx, query, symbol, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetEventsByExchange retrieves events for an exchange within a time range
func (r *TimescaleMarketRepository) GetEventsByExchange(ctx context.Context, exchangeID uuid.UUID, startTime, endTime time.Time, limit int) ([]domain.MarketEvent, error) {
	query := `SELECT id, exchange_id, symbol, event_type, order_id, user_id, price, quantity,
		filled_quantity, direction, status, timestamp, received_at, sequence_num, raw_data, checksum
		FROM market_events WHERE exchange_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp DESC LIMIT $4`

	rows, err := r.db.QueryContext(ctx, query, exchangeID, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetRecentEvents retrieves recent events for an exchange and symbol
func (r *TimescaleMarketRepository) GetRecentEvents(ctx context.Context, exchangeID uuid.UUID, symbol string, duration time.Duration) ([]domain.MarketEvent, error) {
	startTime := time.Now().Add(-duration)

	query := `SELECT id, exchange_id, symbol, event_type, order_id, user_id, price, quantity,
		filled_quantity, direction, status, timestamp, received_at, sequence_num, raw_data, checksum
		FROM market_events WHERE exchange_id = $1 AND symbol = $2 AND timestamp >= $3
		ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query, exchangeID, symbol, startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent events: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetTradesByAccount retrieves trades for an account within a time range
func (r *TimescaleMarketRepository) GetTradesByAccount(ctx context.Context, accountID string, startTime, endTime time.Time, limit int) ([]domain.MarketEvent, error) {
	query := `SELECT id, exchange_id, symbol, event_type, order_id, user_id, price, quantity,
		filled_quantity, direction, status, timestamp, received_at, sequence_num, raw_data, checksum
		FROM market_events WHERE user_id = $1 AND event_type = 'TRADE'
		AND timestamp >= $2 AND timestamp <= $3 ORDER BY timestamp DESC LIMIT $4`

	rows, err := r.db.QueryContext(ctx, query, accountID, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades by account: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetTradesByTimeWindow retrieves trades within a time window
func (r *TimescaleMarketRepository) GetTradesByTimeWindow(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) ([]domain.MarketEvent, error) {
	query := `SELECT id, exchange_id, symbol, event_type, order_id, user_id, price, quantity,
		filled_quantity, direction, status, timestamp, received_at, sequence_num, raw_data, checksum
		FROM market_events WHERE exchange_id = $1 AND symbol = $2 AND event_type = 'TRADE'
		AND timestamp >= $3 AND timestamp <= $4 ORDER BY timestamp`

	rows, err := r.db.QueryContext(ctx, query, exchangeID, symbol, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades by time window: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetRelatedTrades retrieves trades involving specific accounts within a time window
func (r *TimescaleMarketRepository) GetRelatedTrades(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time, accountIDs []string) ([]domain.MarketEvent, error) {
	if len(accountIDs) == 0 {
		return []domain.MarketEvent{}, nil
	}

	query := `SELECT id, exchange_id, symbol, event_type, order_id, user_id, price, quantity,
		filled_quantity, direction, status, timestamp, received_at, sequence_num, raw_data, checksum
		FROM market_events WHERE exchange_id = $1 AND symbol = $2 AND event_type = 'TRADE'
		AND timestamp >= $3 AND timestamp <= $4 AND user_id = ANY($5) ORDER BY timestamp`

	rows, err := r.db.QueryContext(ctx, query, exchangeID, symbol, startTime, endTime, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get related trades: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetOrdersByAccount retrieves orders for an account
func (r *TimescaleMarketRepository) GetOrdersByAccount(ctx context.Context, accountID string, startTime, endTime time.Time) ([]domain.MarketEvent, error) {
	query := `SELECT id, exchange_id, symbol, event_type, order_id, user_id, price, quantity,
		filled_quantity, direction, status, timestamp, received_at, sequence_num, raw_data, checksum
		FROM market_events WHERE user_id = $1 AND event_type = 'ORDER'
		AND timestamp >= $2 AND timestamp <= $3 ORDER BY timestamp`

	rows, err := r.db.QueryContext(ctx, query, accountID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by account: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetCancelledOrders retrieves cancelled orders within a time window
func (r *TimescaleMarketRepository) GetCancelledOrders(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) ([]domain.MarketEvent, error) {
	query := `SELECT id, exchange_id, symbol, event_type, order_id, user_id, price, quantity,
		filled_quantity, direction, status, timestamp, received_at, sequence_num, raw_data, checksum
		FROM market_events WHERE exchange_id = $1 AND symbol = $2 AND event_type = 'ORDER'
		AND status = 'CANCELLED' AND timestamp >= $3 AND timestamp <= $4 ORDER BY timestamp`

	rows, err := r.db.QueryContext(ctx, query, exchangeID, symbol, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get cancelled orders: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// GetOrderLifetimes calculates order lifetimes from creation to cancellation/fill
func (r *TimescaleMarketRepository) GetOrderLifetimes(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) ([]port.OrderLifetime, error) {
	query := `
		WITH orders AS (
			SELECT
				order_id,
				user_id,
				quantity,
				filled_quantity,
				timestamp as created_at,
				NULL::timestamp as cancelled_at
			FROM market_events
			WHERE exchange_id = $1 AND symbol = $2 AND event_type = 'ORDER'
			AND status = 'NEW' AND timestamp >= $3 AND timestamp <= $4
		),
		cancelled AS (
			SELECT
				order_id,
				timestamp as cancelled_at
			FROM market_events
			WHERE exchange_id = $1 AND symbol = $2 AND event_type = 'ORDER'
			AND status = 'CANCELLED' AND timestamp >= $3 AND timestamp <= $4
		)
		SELECT
			o.order_id,
			o.user_id,
			o.quantity,
			o.filled_quantity,
			COALESCE(o.quantity - o.filled_quantity, o.quantity) as cancelled_qty,
			COALESCE(c.cancelled_at, NOW()) - o.created_at as lifetime,
			o.created_at,
			c.cancelled_at
		FROM orders o
		LEFT JOIN cancelled c ON o.order_id = c.order_id
	`

	rows, err := r.db.QueryContext(ctx, query, exchangeID, symbol, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get order lifetimes: %w", err)
	}
	defer rows.Close()

	var lifetimes []port.OrderLifetime
	for rows.Next() {
		var lt port.OrderLifetime
		var createdAt, cancelledAt sql.NullTime

		err := rows.Scan(
			&lt.OrderID, &lt.AccountID, &lt.InitialQty, &lt.FilledQty,
			&lt.CancelledQty, &lt.Lifetime, &createdAt, &cancelledAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order lifetime: %w", err)
		}

		if createdAt.Valid {
			lt.CreatedAt = createdAt.Time
		}
		if cancelledAt.Valid {
			lt.CancelledAt = &cancelledAt.Time
		}

		lifetimes = append(lifetimes, lt)
	}

	return lifetimes, nil
}

// GetMarketSummary retrieves market summary statistics for a symbol
func (r *TimescaleMarketRepository) GetMarketSummary(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) (*domain.MarketSummary, error) {
	query := `
		SELECT
			$1 as exchange_id,
			$2 as symbol,
			$3 as period_start,
			$4 as period_end,
			(SELECT price FROM market_events WHERE exchange_id = $1 AND symbol = $2 AND event_type = 'TRADE' AND timestamp >= $3 ORDER BY timestamp ASC LIMIT 1) as open_price,
			MAX(price) as high_price,
			MIN(price) as low_price,
			(SELECT price FROM market_events WHERE exchange_id = $1 AND symbol = $2 AND event_type = 'TRADE' AND timestamp <= $4 ORDER BY timestamp DESC LIMIT 1) as close_price,
			SUM(quantity) as volume,
			COUNT(*) as trade_count,
			AVG(quantity) as avg_trade_size,
			SUM(price * quantity) / SUM(quantity) as vwap,
			0 as price_volatility
		FROM market_events
		WHERE exchange_id = $1 AND symbol = $2 AND event_type = 'TRADE' AND timestamp >= $3 AND timestamp <= $4
	`

	summary := &domain.MarketSummary{
		ExchangeID:  exchangeID,
		Symbol:      symbol,
		PeriodStart: startTime,
		PeriodEnd:   endTime,
	}

	err := r.db.QueryRowContext(ctx, query, exchangeID, symbol, startTime, endTime).Scan(
		&summary.ExchangeID, &summary.Symbol, &summary.PeriodStart, &summary.PeriodEnd,
		&summary.OpenPrice, &summary.HighPrice, &summary.LowPrice, &summary.ClosePrice,
		&summary.Volume, &summary.TradeCount, &summary.AvgTradeSize, &summary.VWAP,
		&summary.PriceVolatility,
	)

	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get market summary: %w", err)
	}

	summary.UpdatedAt = time.Now()

	return summary, nil
}

// GetMarketStats retrieves market statistics for a symbol
func (r *TimescaleMarketRepository) GetMarketStats(ctx context.Context, exchangeID uuid.UUID, symbol string, duration time.Duration) (*domain.MarketStats, error) {
	startTime := time.Now().Add(-duration)

	query := `
		SELECT
			$3 as symbol,
			$1 as exchange_id,
			AVG(price) as moving_average,
			STDDEV(price) as standard_deviation,
			AVG(quantity) as volume_ma,
			STDDEV(quantity) as volume_std_dev,
			(MAX(price) + MIN(price) + AVG(price)) / 3 as typical_price
		FROM market_events
		WHERE exchange_id = $1 AND symbol = $2 AND event_type = 'TRADE' AND timestamp >= $4
	`

	stats := &domain.MarketStats{
		Symbol:     symbol,
		ExchangeID: exchangeID,
	}

	err := r.db.QueryRowContext(ctx, query, exchangeID, symbol, symbol, startTime).Scan(
		&stats.Symbol, &stats.ExchangeID, &stats.MovingAverage, &stats.StandardDeviation,
		&stats.VolumeMA, &stats.VolumeStdDev, &stats.TypicalPrice,
	)

	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get market stats: %w", err)
	}

	stats.UpdatedAt = time.Now()

	return stats, nil
}

// GetPriceHistory retrieves price history for a symbol
func (r *TimescaleMarketRepository) GetPriceHistory(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time, interval string) ([]domain.PricePoint, error) {
	// Simplified implementation
	return []domain.PricePoint{}, nil
}

// GetVolumeHistory retrieves volume history for a symbol
func (r *TimescaleMarketRepository) GetVolumeHistory(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time, interval string) ([]domain.VolumePoint, error) {
	// Simplified implementation
	return []domain.VolumePoint{}, nil
}

// GetGlobalAveragePrice calculates the global average price across all exchanges
func (r *TimescaleMarketRepository) GetGlobalAveragePrice(ctx context.Context, symbol string, startTime, endTime time.Time) (decimal.Decimal, error) {
	query := `
		SELECT AVG(price)
		FROM market_events
		WHERE symbol = $1 AND event_type = 'TRADE' AND timestamp >= $2 AND timestamp <= $3
	`

	var avgPrice sql.NullString
	err := r.db.QueryRowContext(ctx, query, symbol, startTime, endTime).Scan(&avgPrice)

	if err != nil && err != sql.ErrNoRows {
		return decimal.Zero, fmt.Errorf("failed to get global average price: %w", err)
	}

	if !avgPrice.Valid {
		return decimal.Zero, nil
	}

	price, err := decimal.NewFromString(avgPrice.String)
	if err != nil {
		return decimal.Zero, nil
	}

	return price, nil
}

// StoreOrderBook stores an order book snapshot
func (r *TimescaleMarketRepository) StoreOrderBook(ctx context.Context, snapshot *domain.OrderBookSnapshot) error {
	// Placeholder implementation
	return nil
}

// GetOrderBookSnapshot retrieves an order book snapshot by ID
func (r *TimescaleMarketRepository) GetOrderBookSnapshot(ctx context.Context, id uuid.UUID) (*domain.OrderBookSnapshot, error) {
	// Placeholder implementation
	return nil, nil
}

// BulkStoreEvents stores multiple market events efficiently
func (r *TimescaleMarketRepository) BulkStoreEvents(ctx context.Context, events []domain.MarketEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO market_events (
		id, exchange_id, symbol, event_type, order_id, user_id, price, quantity,
		filled_quantity, direction, status, timestamp, received_at, sequence_num, raw_data, checksum
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		_, err := stmt.ExecContext(ctx,
			event.ID, event.ExchangeID, event.Symbol, event.EventType, event.OrderID,
			event.UserID, event.Price, event.Quantity, event.FilledQuantity,
			event.Direction, event.Status, event.Timestamp, event.ReceivedAt,
			event.SequenceNum, event.RawData, event.Checksum,
		)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	return tx.Commit()
}

// scanEvents scans rows into MarketEvent structs
func (r *TimescaleMarketRepository) scanEvents(rows *sql.Rows) ([]domain.MarketEvent, error) {
	var events []domain.MarketEvent

	for rows.Next() {
		event := domain.MarketEvent{}
		err := rows.Scan(
			&event.ID, &event.ExchangeID, &event.Symbol, &event.EventType, &event.OrderID,
			&event.UserID, &event.Price, &event.Quantity, &event.FilledQuantity,
			&event.Direction, &event.Status, &event.Timestamp, &event.ReceivedAt,
			&event.SequenceNum, &event.RawData, &event.Checksum,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// Ensure the interface is satisfied
var _ port.MarketRepository = (*TimescaleMarketRepository)(nil)
