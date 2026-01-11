package adapters

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/google/uuid"
	"github.com/csic/oversight/internal/core/domain"
	"go.uber.org/zap"
)

// PostgresRepository provides PostgreSQL persistence for oversight data
type PostgresRepository struct {
	db          *sql.DB
	logger      *zap.Logger
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger,
	}
}

// InitSchema creates the necessary database tables
func (r *PostgresRepository) InitSchema(ctx context.Context) error {
	queries := []string{
		// Exchange registry table
		`CREATE TABLE IF NOT EXISTS oversight_exchanges (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			api_endpoint VARCHAR(512),
			status VARCHAR(20) DEFAULT 'unknown',
			health_score DECIMAL(5,2) DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Health metrics table (TimescaleDB hypertable)
		`CREATE TABLE IF NOT EXISTS oversight_health_metrics (
			id BIGSERIAL,
			exchange_id VARCHAR(36) REFERENCES oversight_exchanges(id),
			latency_ms INTEGER DEFAULT 0,
			uptime_percent DECIMAL(5,2) DEFAULT 100,
			liquidity_depth DECIMAL(20,8) DEFAULT 0,
			api_status INTEGER DEFAULT 200,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Trade anomalies table
		`CREATE TABLE IF NOT EXISTS oversight_anomalies (
			id VARCHAR(36) PRIMARY KEY,
			exchange_id VARCHAR(36) REFERENCES oversight_exchanges(id),
			symbol VARCHAR(20) NOT NULL,
			abuse_type VARCHAR(50) NOT NULL,
			confidence_score DECIMAL(5,4) DEFAULT 0,
			description TEXT,
			trade_ids TEXT[],
			status VARCHAR(20) DEFAULT 'pending',
			detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Trades table
		`CREATE TABLE IF NOT EXISTS oversight_trades (
			id VARCHAR(36) PRIMARY KEY,
			exchange_id VARCHAR(36) REFERENCES oversight_exchanges(id),
			symbol VARCHAR(20) NOT NULL,
			side VARCHAR(10) NOT NULL,
			price DECIMAL(20,8) NOT NULL,
			quantity DECIMAL(20,8) NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Market depth table
		`CREATE TABLE IF NOT EXISTS oversight_market_depth (
			id BIGSERIAL,
			exchange_id VARCHAR(36) REFERENCES oversight_exchanges(id),
			symbol VARCHAR(20) NOT NULL,
			bids JSONB,
			asks JSONB,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_health_exchange ON oversight_health_metrics(exchange_id, timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_anomalies_exchange ON oversight_anomalies(exchange_id, detected_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_anomalies_status ON oversight_anomalies(status)`,
		`CREATE INDEX IF NOT EXISTS idx_trades_exchange_symbol ON oversight_trades(exchange_id, symbol, timestamp DESC)`,
	}
	
	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}
	
	// Convert to hypertable if TimescaleDB is available
	hypertableQuery := `SELECT create_hypertable('oversight_health_metrics', 'timestamp')`
	if _, err := r.db.ExecContext(ctx, hypertableQuery); err != nil {
		r.logger.Info("TimescaleDB hypertable creation skipped (TimescaleDB may not be available)")
	}
	
	return nil
}

// Exchange operations
func (r *PostgresRepository) CreateExchange(ctx context.Context, exchange *domain.Exchange) error {
	query := `INSERT INTO oversight_exchanges (id, name, api_endpoint, status, health_score, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (id) DO UPDATE SET
		name = EXCLUDED.name,
		api_endpoint = EXCLUDED.api_endpoint,
		updated_at = CURRENT_TIMESTAMP`
	
	_, err := r.db.ExecContext(ctx, query,
		exchange.ID,
		exchange.Name,
		exchange.APIEndpoint,
		exchange.Status,
		exchange.HealthScore,
		exchange.CreatedAt,
		exchange.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetExchange(ctx context.Context, id string) (*domain.Exchange, error) {
	query := `SELECT id, name, api_endpoint, status, health_score, created_at, updated_at
	FROM oversight_exchanges WHERE id = $1`
	
	var exchange domain.Exchange
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&exchange.ID,
		&exchange.Name,
		&exchange.APIEndpoint,
		&exchange.Status,
		&exchange.HealthScore,
		&exchange.CreatedAt,
		&exchange.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &exchange, nil
}

func (r *PostgresRepository) UpdateExchangeStatus(ctx context.Context, id string, status domain.ExchangeStatus) error {
	query := `UPDATE oversight_exchanges SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *PostgresRepository) ListExchanges(ctx context.Context) ([]*domain.Exchange, error) {
	query := `SELECT id, name, api_endpoint, status, health_score, created_at, updated_at
	FROM oversight_exchanges ORDER BY name`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var exchanges []*domain.Exchange
	for rows.Next() {
		var exchange domain.Exchange
		if err := rows.Scan(
			&exchange.ID,
			&exchange.Name,
			&exchange.APIEndpoint,
			&exchange.Status,
			&exchange.HealthScore,
			&exchange.CreatedAt,
			&exchange.UpdatedAt,
		); err != nil {
			return nil, err
		}
		exchanges = append(exchanges, &exchange)
	}
	return exchanges, rows.Err()
}

// Health metrics operations
func (r *PostgresRepository) RecordHealthMetrics(ctx context.Context, metrics *domain.ExchangeHealthMetrics) error {
	query := `INSERT INTO oversight_health_metrics
	(exchange_id, latency_ms, uptime_percent, liquidity_depth, api_status, timestamp)
	VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err := r.db.ExecContext(ctx, query,
		metrics.ExchangeID,
		metrics.LatencyMs,
		metrics.UptimePercent,
		metrics.LiquidityDepth,
		metrics.ApiStatus,
		metrics.Timestamp,
	)
	return err
}

func (r *PostgresRepository) GetHealthMetrics(ctx context.Context, exchangeID string, from, to time.Time) ([]*domain.ExchangeHealthMetrics, error) {
	query := `SELECT id, exchange_id, latency_ms, uptime_percent, liquidity_depth, api_status, timestamp
	FROM oversight_health_metrics
	WHERE exchange_id = $1 AND timestamp BETWEEN $2 AND $3
	ORDER BY timestamp DESC`
	
	rows, err := r.db.QueryContext(ctx, query, exchangeID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var metrics []*domain.ExchangeHealthMetrics
	for rows.Next() {
		var m domain.ExchangeHealthMetrics
		if err := rows.Scan(
			&m.ID,
			&m.ExchangeID,
			&m.LatencyMs,
			&m.UptimePercent,
			&m.LiquidityDepth,
			&m.ApiStatus,
			&m.Timestamp,
		); err != nil {
			return nil, err
		}
		metrics = append(metrics, &m)
	}
	return metrics, rows.Err()
}

func (r *PostgresRepository) CalculateAverageLatency(ctx context.Context, exchangeID string, window time.Duration) (int, error) {
	query := `SELECT COALESCE(AVG(latency_ms), 0) FROM oversight_health_metrics
	WHERE exchange_id = $1 AND timestamp > $2`
	
	var avgLatency float64
	err := r.db.QueryRowContext(ctx, query, exchangeID, time.Now().Add(-window)).Scan(&avgLatency)
	return int(avgLatency), err
}

// Anomaly operations
func (r *PostgresRepository) RecordAnomaly(ctx context.Context, anomaly *domain.TradeAnomaly) error {
	query := `INSERT INTO oversight_anomalies
	(id, exchange_id, symbol, abuse_type, confidence_score, description, trade_ids, status, detected_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT (id) DO NOTHING`
	
	_, err := r.db.ExecContext(ctx, query,
		anomaly.ID,
		anomaly.ExchangeID,
		anomaly.Symbol,
		anomaly.AbuseType,
		anomaly.ConfidenceScore,
		anomaly.Description,
		anomaly.TradeIDs,
		anomaly.Status,
		anomaly.DetectedAt,
	)
	return err
}

func (r *PostgresRepository) GetAnomalies(ctx context.Context, exchangeID string, from, to time.Time) ([]*domain.TradeAnomaly, error) {
	query := `SELECT id, exchange_id, symbol, abuse_type, confidence_score, description, trade_ids, status, detected_at
	FROM oversight_anomalies
	WHERE exchange_id = $1 AND detected_at BETWEEN $2 AND $3
	ORDER BY detected_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, exchangeID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var anomalies []*domain.TradeAnomaly
	for rows.Next() {
		var a domain.TradeAnomaly
		if err := rows.Scan(
			&a.ID,
			&a.ExchangeID,
			&a.Symbol,
			&a.AbuseType,
			&a.ConfidenceScore,
			&a.Description,
			&a.TradeIDs,
			&a.Status,
			&a.DetectedAt,
		); err != nil {
			return nil, err
		}
		anomalies = append(anomalies, &a)
	}
	return anomalies, rows.Err()
}

func (r *PostgresRepository) UpdateAnomalyStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE oversight_anomalies SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *PostgresRepository) GetPendingAnomalies(ctx context.Context) ([]*domain.TradeAnomaly, error) {
	query := `SELECT id, exchange_id, symbol, abuse_type, confidence_score, description, trade_ids, status, detected_at
	FROM oversight_anomalies WHERE status = 'pending' ORDER BY detected_at DESC LIMIT 100`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var anomalies []*domain.TradeAnomaly
	for rows.Next() {
		var a domain.TradeAnomaly
		if err := rows.Scan(
			&a.ID,
			&a.ExchangeID,
			&a.Symbol,
			&a.AbuseType,
			&a.ConfidenceScore,
			&a.Description,
			&a.TradeIDs,
			&a.Status,
			&a.DetectedAt,
		); err != nil {
			return nil, err
		}
		anomalies = append(anomalies, &a)
	}
	return anomalies, rows.Err()
}

// Trade operations
func (r *PostgresRepository) RecordTrade(ctx context.Context, trade *domain.Trade) error {
	if trade.ID == "" {
		trade.ID = uuid.New().String()
	}
	query := `INSERT INTO oversight_trades (id, exchange_id, symbol, side, price, quantity, timestamp)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (id) DO NOTHING`
	
	_, err := r.db.ExecContext(ctx, query,
		trade.ID,
		trade.ExchangeID,
		trade.Symbol,
		trade.Side,
		trade.Price,
		trade.Quantity,
		trade.Timestamp,
	)
	return err
}

func (r *PostgresRepository) GetRecentTrades(ctx context.Context, exchangeID, symbol string, window time.Duration) ([]*domain.Trade, error) {
	query := `SELECT id, exchange_id, symbol, side, price, quantity, timestamp
	FROM oversight_trades
	WHERE exchange_id = $1 AND symbol = $2 AND timestamp > $3
	ORDER BY timestamp DESC`
	
	rows, err := r.db.QueryContext(ctx, query, exchangeID, symbol, time.Now().Add(-window))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var trades []*domain.Trade
	for rows.Next() {
		var t domain.Trade
		if err := rows.Scan(
			&t.ID,
			&t.ExchangeID,
			&t.Symbol,
			&t.Side,
			&t.Price,
			&t.Quantity,
			&t.Timestamp,
		); err != nil {
			return nil, err
		}
		trades = append(trades, &t)
	}
	return trades, rows.Err()
}

// Market depth operations
func (r *PostgresRepository) RecordMarketDepth(ctx context.Context, depth *domain.MarketDepth) error {
	query := `INSERT INTO oversight_market_depth (exchange_id, symbol, bids, asks, timestamp)
	VALUES ($1, $2, $3, $4, $5)`
	
	bidsJSON, _ := json.Marshal(depth.Bids)
	asksJSON, _ := json.Marshal(depth.Asks)
	
	_, err := r.db.ExecContext(ctx, query,
		depth.ExchangeID,
		depth.Symbol,
		bidsJSON,
		asksJSON,
		depth.Timestamp,
	)
	return err
}

func (r *PostgresRepository) GetMarketDepth(ctx context.Context, exchangeID, symbol string, timestamp time.Time) (*domain.MarketDepth, error) {
	query := `SELECT exchange_id, symbol, bids, asks, timestamp FROM oversight_market_depth
	WHERE exchange_id = $1 AND symbol = $2 AND timestamp <= $3 ORDER BY timestamp DESC LIMIT 1`
	
	var depth domain.MarketDepth
	var bidsJSON, asksJSON []byte
	
	err := r.db.QueryRowContext(ctx, query, exchangeID, symbol, timestamp).Scan(
		&depth.ExchangeID,
		&depth.Symbol,
		&bidsJSON,
		&asksJSON,
		&depth.Timestamp,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	json.Unmarshal(bidsJSON, &depth.Bids)
	json.Unmarshal(asksJSON, &depth.Asks)
	
	return &depth, nil
}
