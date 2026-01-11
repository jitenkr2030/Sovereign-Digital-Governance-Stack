package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/oversight/internal/core/domain"
	"github.com/csic/oversight/internal/core/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// RuleRepository implements the RuleRepository interface for PostgreSQL
type RuleRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	table  string
}

// NewRuleRepository creates a new RuleRepository
func NewRuleRepository(pool *pgxpool.Pool, logger *zap.Logger) *RuleRepository {
	return &RuleRepository{
		pool:   pool,
		logger: logger,
		table:  "detection_rules",
	}
}

// SaveRule saves a detection rule to the database
func (r *RuleRepository) SaveRule(ctx context.Context, rule domain.DetectionRule) error {
	logicJSON, err := json.Marshal(rule.Logic)
	if err != nil {
		return fmt.Errorf("failed to marshal rule logic: %w", err)
	}

	exchangesJSON, err := json.Marshal(rule.Exchanges)
	if err != nil {
		return fmt.Errorf("failed to marshal rule exchanges: %w", err)
	}

	tradingPairsJSON, err := json.Marshal(rule.TradingPairs)
	if err != nil {
		return fmt.Errorf("failed to marshal rule trading pairs: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, name, description, alert_type, severity, logic,
			time_window_ms, threshold, cooldown_secs, is_active,
			exchanges, trading_pairs, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			logic = EXCLUDED.logic,
			time_window_ms = EXCLUDED.time_window_ms,
			threshold = EXCLUDED.threshold,
			cooldown_secs = EXCLUDED.cooldown_secs,
			is_active = EXCLUDED.is_active,
			exchanges = EXCLUDED.exchanges,
			trading_pairs = EXCLUDED.trading_pairs,
			updated_at = EXCLUDED.updated_at
	`, r.table)

	_, err = r.pool.Exec(ctx, query,
		rule.ID,
		rule.Name,
		rule.Description,
		rule.AlertType,
		rule.Severity,
		logicJSON,
		rule.TimeWindow,
		rule.Threshold,
		rule.CooldownSecs,
		rule.Enabled,
		exchangesJSON,
		tradingPairsJSON,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save rule: %w", err)
	}

	return nil
}

// SaveRuleBatch saves multiple detection rules
func (r *RuleRepository) SaveRuleBatch(ctx context.Context, rules []domain.DetectionRule) error {
	if len(rules) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	for _, rule := range rules {
		logicJSON, _ := json.Marshal(rule.Logic)
		exchangesJSON, _ := json.Marshal(rule.Exchanges)
		tradingPairsJSON, _ := json.Marshal(rule.TradingPairs)

		query := fmt.Sprintf(`
			INSERT INTO %s (
				id, name, description, alert_type, severity, logic,
				time_window_ms, threshold, cooldown_secs, is_active,
				exchanges, trading_pairs, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				is_active = EXCLUDED.is_active,
				updated_at = EXCLUDED.updated_at
		`, r.table)

		batch.Queue(query,
			rule.ID,
			rule.Name,
			rule.Description,
			rule.AlertType,
			rule.Severity,
			logicJSON,
			rule.TimeWindow,
			rule.Threshold,
			rule.CooldownSecs,
			rule.Enabled,
			exchangesJSON,
			tradingPairsJSON,
			rule.CreatedAt,
			rule.UpdatedAt,
		)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	_, err := results.Exec()
	if err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	return nil
}

// GetRuleByID retrieves a specific rule by ID
func (r *RuleRepository) GetRuleByID(ctx context.Context, ruleID string) (*domain.DetectionRule, error) {
	query := fmt.Sprintf(`
		SELECT id, name, description, alert_type, severity, logic,
			   time_window_ms, threshold, cooldown_secs, is_active,
			   exchanges, trading_pairs, created_at, updated_at
		FROM %s
		WHERE id = $1
	`, r.table)

	row := r.pool.QueryRow(ctx, query, ruleID)

	rule, err := r.scanRule(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("rule not found: %s", ruleID)
		}
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}

	return rule, nil
}

// GetAllRules retrieves all detection rules
func (r *RuleRepository) GetAllRules(ctx context.Context) ([]domain.DetectionRule, error) {
	query := fmt.Sprintf(`
		SELECT id, name, description, alert_type, severity, logic,
			   time_window_ms, threshold, cooldown_secs, is_active,
			   exchanges, trading_pairs, created_at, updated_at
		FROM %s
		ORDER BY created_at DESC
	`, r.table)

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query rules: %w", err)
	}
	defer rows.Close()

	return r.scanRules(rows)
}

// GetActiveRules retrieves all active detection rules
func (r *RuleRepository) GetActiveRules(ctx context.Context) ([]domain.DetectionRule, error) {
	query := fmt.Sprintf(`
		SELECT id, name, description, alert_type, severity, logic,
			   time_window_ms, threshold, cooldown_secs, is_active,
			   exchanges, trading_pairs, created_at, updated_at
		FROM %s
		WHERE is_active = true
		ORDER BY created_at DESC
	`, r.table)

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active rules: %w", err)
	}
	defer rows.Close()

	return r.scanRules(rows)
}

// GetRulesByType retrieves rules by alert type
func (r *RuleRepository) GetRulesByType(ctx context.Context, alertType domain.AlertType) ([]domain.DetectionRule, error) {
	query := fmt.Sprintf(`
		SELECT id, name, description, alert_type, severity, logic,
			   time_window_ms, threshold, cooldown_secs, is_active,
			   exchanges, trading_pairs, created_at, updated_at
		FROM %s
		WHERE alert_type = $1 AND is_active = true
		ORDER BY created_at DESC
	`, r.table)

	rows, err := r.pool.Query(ctx, query, alertType)
	if err != nil {
		return nil, fmt.Errorf("failed to query rules by type: %w", err)
	}
	defer rows.Close()

	return r.scanRules(rows)
}

// UpdateRule updates an existing detection rule
func (r *RuleRepository) UpdateRule(ctx context.Context, rule domain.DetectionRule) error {
	rule.UpdatedAt = rule.CreatedAt // Simplified, should be time.Now() in production

	return r.SaveRule(ctx, rule)
}

// DeleteRule soft deletes a detection rule (sets enabled to false)
func (r *RuleRepository) DeleteRule(ctx context.Context, ruleID string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`, r.table)

	result, err := r.pool.Exec(ctx, query, ruleID)
	if err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	return nil
}

// HardDeleteRule permanently removes a detection rule
func (r *RuleRepository) HardDeleteRule(ctx context.Context, ruleID string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, r.table)

	result, err := r.pool.Exec(ctx, query, ruleID)
	if err != nil {
		return fmt.Errorf("failed to hard delete rule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	return nil
}

// scanRules scans multiple rule rows
func (r *RuleRepository) scanRules(rows pgx.Rows) ([]domain.DetectionRule, error) {
	rules := make([]domain.DetectionRule, 0)

	for rows.Next() {
		rule, err := r.scanRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error scanning rules: %w", err)
	}

	return rules, nil
}

// scanRule scans a single rule row
func (r *RuleRepository) scanRule(row pgx.Row) (*domain.DetectionRule, error) {
	var rule domain.DetectionRule
	var logicJSON, exchangesJSON, tradingPairsJSON []byte

	err := row.Scan(
		&rule.ID,
		&rule.Name,
		&rule.Description,
		&rule.AlertType,
		&rule.Severity,
		&logicJSON,
		&rule.TimeWindow,
		&rule.Threshold,
		&rule.CooldownSecs,
		&rule.Enabled,
		&exchangesJSON,
		&tradingPairsJSON,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(logicJSON, &rule.Logic); err != nil {
		r.logger.Warn("Failed to unmarshal rule logic", zap.Error(err))
	}

	if err := json.Unmarshal(exchangesJSON, &rule.Exchanges); err != nil {
		r.logger.Warn("Failed to unmarshal rule exchanges", zap.Error(err))
	}

	if err := json.Unmarshal(tradingPairsJSON, &rule.TradingPairs); err != nil {
		r.logger.Warn("Failed to unmarshal rule trading pairs", zap.Error(err))
	}

	return &rule, nil
}

// ExchangeRepository implements the ExchangeRepository interface for PostgreSQL
type ExchangeRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	table  string
}

// NewExchangeRepository creates a new ExchangeRepository
func NewExchangeRepository(pool *pgxpool.Pool, logger *zap.Logger) *ExchangeRepository {
	return &ExchangeRepository{
		pool:   pool,
		logger: logger,
		table:  "exchange_profiles",
	}
}

// SaveExchange saves an exchange profile
func (r *ExchangeRepository) SaveExchange(ctx context.Context, exchange domain.ExchangeProfile) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, exchange_id, name, license_number, jurisdiction,
			api_endpoint, websocket_endpoint, rate_limit_rps,
			throttle_enabled, max_latency_ms, is_active,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			license_number = EXCLUDED.license_number,
			api_endpoint = EXCLUDED.api_endpoint,
			websocket_endpoint = EXCLUDED.websocket_endpoint,
			rate_limit_rps = EXCLUDED.rate_limit_rps,
			throttle_enabled = EXCLUDED.throttle_enabled,
			max_latency_ms = EXCLUDED.max_latency_ms,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at
	`, r.table)

	_, err := r.pool.Exec(ctx, query,
		exchange.ID,
		exchange.ExchangeID,
		exchange.Name,
		exchange.LicenseNumber,
		exchange.Jurisdiction,
		exchange.APIEndpoint,
		exchange.WebSocketEndpoint,
		exchange.RateLimitRps,
		exchange.ThrottleEnabled,
		exchange.MaxLatencyMs,
		exchange.IsActive,
		exchange.CreatedAt,
		exchange.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save exchange: %w", err)
	}

	return nil
}

// GetExchangeByID retrieves an exchange by ID
func (r *ExchangeRepository) GetExchangeByID(ctx context.Context, exchangeID string) (*domain.ExchangeProfile, error) {
	query := fmt.Sprintf(`
		SELECT id, exchange_id, name, license_number, jurisdiction,
			   api_endpoint, websocket_endpoint, rate_limit_rps,
			   throttle_enabled, max_latency_ms, is_active,
			   created_at, updated_at
		FROM %s
		WHERE exchange_id = $1
	`, r.table)

	row := r.pool.QueryRow(ctx, query, exchangeID)

	exchange, err := r.scanExchange(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("exchange not found: %s", exchangeID)
		}
		return nil, fmt.Errorf("failed to get exchange: %w", err)
	}

	return exchange, nil
}

// GetAllExchanges retrieves all registered exchanges
func (r *ExchangeRepository) GetAllExchanges(ctx context.Context) ([]domain.ExchangeProfile, error) {
	query := fmt.Sprintf(`
		SELECT id, exchange_id, name, license_number, jurisdiction,
			   api_endpoint, websocket_endpoint, rate_limit_rps,
			   throttle_enabled, max_latency_ms, is_active,
			   created_at, updated_at
		FROM %s
		ORDER BY name
	`, r.table)

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query exchanges: %w", err)
	}
	defer rows.Close()

	return r.scanExchanges(rows)
}

// GetActiveExchanges retrieves all active exchanges
func (r *ExchangeRepository) GetActiveExchanges(ctx context.Context) ([]domain.ExchangeProfile, error) {
	query := fmt.Sprintf(`
		SELECT id, exchange_id, name, license_number, jurisdiction,
			   api_endpoint, websocket_endpoint, rate_limit_rps,
			   throttle_enabled, max_latency_ms, is_active,
			   created_at, updated_at
		FROM %s
		WHERE is_active = true
		ORDER BY name
	`, r.table)

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active exchanges: %w", err)
	}
	defer rows.Close()

	return r.scanExchanges(rows)
}

// UpdateExchange updates an existing exchange profile
func (r *ExchangeRepository) UpdateExchange(ctx context.Context, exchange domain.ExchangeProfile) error {
	return r.SaveExchange(ctx, exchange)
}

// DeleteExchange soft deletes an exchange
func (r *ExchangeRepository) DeleteExchange(ctx context.Context, exchangeID string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET is_active = false, updated_at = NOW()
		WHERE exchange_id = $1
	`, r.table)

	result, err := r.pool.Exec(ctx, query, exchangeID)
	if err != nil {
		return fmt.Errorf("failed to delete exchange: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("exchange not found: %s", exchangeID)
	}

	return nil
}

// scanExchanges scans multiple exchange rows
func (r *ExchangeRepository) scanExchanges(rows pgx.Rows) ([]domain.ExchangeProfile, error) {
	exchanges := make([]domain.ExchangeProfile, 0)

	for rows.Next() {
		exchange, err := r.scanExchange(rows)
		if err != nil {
			return nil, err
		}
		exchanges = append(exchanges, *exchange)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error scanning exchanges: %w", err)
	}

	return exchanges, nil
}

// scanExchange scans a single exchange row
func (r *ExchangeRepository) scanExchange(row pgx.Row) (*domain.ExchangeProfile, error) {
	var exchange domain.ExchangeProfile

	err := row.Scan(
		&exchange.ID,
		&exchange.ExchangeID,
		&exchange.Name,
		&exchange.LicenseNumber,
		&exchange.Jurisdiction,
		&exchange.APIEndpoint,
		&exchange.WebSocketEndpoint,
		&exchange.RateLimitRps,
		&exchange.ThrottleEnabled,
		&exchange.MaxLatencyMs,
		&exchange.IsActive,
		&exchange.CreatedAt,
		&exchange.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &exchange, nil
}

// HealthRepository implements the HealthRepository interface for PostgreSQL
type HealthRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	table  string
}

// NewHealthRepository creates a new HealthRepository
func NewHealthRepository(pool *pgxpool.Pool, logger *zap.Logger) *HealthRepository {
	return &HealthRepository{
		pool:   pool,
		logger: logger,
		table:  "exchange_health",
	}
}

// SaveHealthRecord saves an exchange health record
func (r *HealthRepository) SaveHealthRecord(ctx context.Context, health domain.ExchangeHealth) error {
	metricsJSON, err := json.Marshal(health.Metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal health metrics: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, exchange_id, health_score, latency_ms, error_rate,
			trade_volume_24h, trade_count_24h, uptime_percent,
			last_trade_at, last_heartbeat_at, status, metrics, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (id) DO UPDATE SET
			health_score = EXCLUDED.health_score,
			latency_ms = EXCLUDED.latency_ms,
			error_rate = EXCLUDED.error_rate,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, r.table)

	_, err = r.pool.Exec(ctx, query,
		health.ID,
		health.ExchangeID,
		health.HealthScore,
		health.LatencyMs,
		health.ErrorRate,
		health.TradeVolume24h,
		health.TradeCount24h,
		health.UptimePercent,
		health.LastTradeAt,
		health.LastHeartbeatAt,
		health.Status,
		metricsJSON,
		health.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save health record: %w", err)
	}

	return nil
}

// SaveHealthRecordBatch saves multiple health records
func (r *HealthRepository) SaveHealthRecordBatch(ctx context.Context, healthRecords []domain.ExchangeHealth) error {
	if len(healthRecords) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	for _, health := range healthRecords {
		metricsJSON, _ := json.Marshal(health.Metrics)

		query := fmt.Sprintf(`
			INSERT INTO %s (
				id, exchange_id, health_score, latency_ms, error_rate,
				trade_volume_24h, trade_count_24h, uptime_percent,
				last_trade_at, last_heartbeat_at, status, metrics, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (id) DO UPDATE SET
				health_score = EXCLUDED.health_score,
				status = EXCLUDED.status
		`, r.table)

		batch.Queue(query,
			health.ID,
			health.ExchangeID,
			health.HealthScore,
			health.LatencyMs,
			health.ErrorRate,
			health.TradeVolume24h,
			health.TradeCount24h,
			health.UptimePercent,
			health.LastTradeAt,
			health.LastHeartbeatAt,
			health.Status,
			metricsJSON,
			health.UpdatedAt,
		)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	_, err := results.Exec()
	if err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	return nil
}

// GetLatestHealth retrieves the latest health record for an exchange
func (r *HealthRepository) GetLatestHealth(ctx context.Context, exchangeID string) (*domain.ExchangeHealth, error) {
	query := fmt.Sprintf(`
		SELECT id, exchange_id, health_score, latency_ms, error_rate,
			   trade_volume_24h, trade_count_24h, uptime_percent,
			   last_trade_at, last_heartbeat_at, status, metrics, updated_at
		FROM %s
		WHERE exchange_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`, r.table)

	row := r.pool.QueryRow(ctx, query, exchangeID)

	health, err := r.scanHealth(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("health record not found for exchange: %s", exchangeID)
		}
		return nil, fmt.Errorf("failed to get health record: %w", err)
	}

	return health, nil
}

// GetHealthHistory retrieves health history for an exchange
func (r *HealthRepository) GetHealthHistory(ctx context.Context, exchangeID string, start, end time.Time) ([]domain.ExchangeHealth, error) {
	query := fmt.Sprintf(`
		SELECT id, exchange_id, health_score, latency_ms, error_rate,
			   trade_volume_24h, trade_count_24h, uptime_percent,
			   last_trade_at, last_heartbeat_at, status, metrics, updated_at
		FROM %s
		WHERE exchange_id = $1 AND updated_at BETWEEN $2 AND $3
		ORDER BY updated_at DESC
	`, r.table)

	rows, err := r.pool.Query(ctx, query, exchangeID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query health history: %w", err)
	}
	defer rows.Close()

	return r.scanHealthRecords(rows)
}

// GetHealthStats retrieves aggregated health statistics
func (r *HealthRepository) GetHealthStats(ctx context.Context, exchangeID string, period time.Duration) (ports.HealthStats, error) {
	query := fmt.Sprintf(`
		SELECT
			exchange_id,
			AVG(health_score) as avg_score,
			MIN(health_score) as min_score,
			MAX(health_score) as max_score,
			AVG(latency_ms) as avg_latency,
			AVG(error_rate) as avg_error,
			AVG(uptime_percent) as avg_uptime
		FROM %s
		WHERE exchange_id = $1 AND updated_at >= NOW() - INTERVAL '%d seconds'
		GROUP BY exchange_id
	`, r.table, int(period.Seconds()))

	row := r.pool.QueryRow(ctx, query, exchangeID)

	var stats ports.HealthStats
	var minScore, maxScore, avgLatency, avgError, avgUptime *float64

	err := row.Scan(
		&stats.ExchangeID,
		&stats.AvgHealthScore,
		&minScore,
		&maxScore,
		&avgLatency,
		&avgError,
		&avgUptime,
	)

	if err != nil {
		return stats, fmt.Errorf("failed to get health stats: %w", err)
	}

	if minScore != nil {
		stats.MinHealthScore = *minScore
	}
	if maxScore != nil {
		stats.MaxHealthScore = *maxScore
	}
	if avgLatency != nil {
		stats.AvgLatencyMs = *avgLatency
	}
	if avgError != nil {
		stats.AvgErrorRate = *avgError
	}
	if avgUptime != nil {
		stats.AvgUptimePercent = *avgUptime
	}

	return stats, nil
}

// scanHealthRecords scans multiple health record rows
func (r *HealthRepository) scanHealthRecords(rows pgx.Rows) ([]domain.ExchangeHealth, error) {
	records := make([]domain.ExchangeHealth, 0)

	for rows.Next() {
		record, err := r.scanHealth(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, *record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error scanning health records: %w", err)
	}

	return records, nil
}

// scanHealth scans a single health record row
func (r *HealthRepository) scanHealth(row pgx.Row) (*domain.ExchangeHealth, error) {
	var health domain.ExchangeHealth
	var metricsJSON []byte

	err := row.Scan(
		&health.ID,
		&health.ExchangeID,
		&health.HealthScore,
		&health.LatencyMs,
		&health.ErrorRate,
		&health.TradeVolume24h,
		&health.TradeCount24h,
		&health.UptimePercent,
		&health.LastTradeAt,
		&health.LastHeartbeatAt,
		&health.Status,
		&metricsJSON,
		&health.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(metricsJSON, &health.Metrics); err != nil {
		r.logger.Warn("Failed to unmarshal health metrics", zap.Error(err))
	}

	return &health, nil
}
