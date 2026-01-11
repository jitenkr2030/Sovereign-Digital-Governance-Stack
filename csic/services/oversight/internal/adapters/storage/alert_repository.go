package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/csic/oversight/internal/core/domain"
	"github.com/csic/oversight/internal/core/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// AlertRepository implements the AlertRepository interface for PostgreSQL
type AlertRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	table  string
}

// NewAlertRepository creates a new AlertRepository
func NewAlertRepository(pool *pgxpool.Pool, logger *zap.Logger) *AlertRepository {
	return &AlertRepository{
		pool:   pool,
		logger: logger,
		table:  "alerts",
	}
}

// SaveAlert saves a market alert to the database
func (r *AlertRepository) SaveAlert(ctx context.Context, alert domain.MarketAlert) error {
	detailsJSON, err := json.Marshal(alert.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal alert details: %w", err)
	}

	evidenceJSON, err := json.Marshal(alert.Evidence)
	if err != nil {
		return fmt.Errorf("failed to marshal alert evidence: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, severity, alert_type, exchange_id, trading_pair, user_id,
			details, evidence, status, assigned_to, resolved_at, resolution,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, r.table)

	_, err = r.pool.Exec(ctx, query,
		alert.ID,
		alert.Severity,
		alert.AlertType,
		alert.ExchangeID,
		alert.TradingPair,
		alert.UserID,
		detailsJSON,
		evidenceJSON,
		alert.Status,
		alert.AssignedTo,
		alert.ResolvedAt,
		alert.Resolution,
		alert.CreatedAt,
		alert.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save alert: %w", err)
	}

	return nil
}

// SaveAlertBatch saves multiple alerts in a single transaction
func (r *AlertRepository) SaveAlertBatch(ctx context.Context, alerts []domain.MarketAlert) error {
	if len(alerts) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	for _, alert := range alerts {
		detailsJSON, err := json.Marshal(alert.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal alert details: %w", err)
		}

		evidenceJSON, err := json.Marshal(alert.Evidence)
		if err != nil {
			return fmt.Errorf("failed to marshal alert evidence: %w", err)
		}

		query := fmt.Sprintf(`
			INSERT INTO %s (
				id, severity, alert_type, exchange_id, trading_pair, user_id,
				details, evidence, status, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (id) DO NOTHING
		`, r.table)

		batch.Queue(query,
			alert.ID,
			alert.Severity,
			alert.AlertType,
			alert.ExchangeID,
			alert.TradingPair,
			alert.UserID,
			detailsJSON,
			evidenceJSON,
			alert.Status,
			alert.CreatedAt,
			alert.UpdatedAt,
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

// GetAlerts retrieves alerts based on filter criteria
func (r *AlertRepository) GetAlerts(ctx context.Context, filter ports.AlertFilter) ([]domain.MarketAlert, error) {
	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	argIdx := 1

	if filter.ExchangeID != "" {
		conditions = append(conditions, fmt.Sprintf("exchange_id = $%d", argIdx))
		args = append(args, filter.ExchangeID)
		argIdx++
	}

	if filter.AlertType != "" {
		conditions = append(conditions, fmt.Sprintf("alert_type = $%d", argIdx))
		args = append(args, filter.AlertType)
		argIdx++
	}

	if filter.Severity != "" {
		conditions = append(conditions, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, filter.Severity)
		argIdx++
	}

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.TradingPair != "" {
		conditions = append(conditions, fmt.Sprintf("trading_pair = $%d", argIdx))
		args = append(args, filter.TradingPair)
		argIdx++
	}

	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, filter.UserID)
		argIdx++
	}

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *filter.StartTime)
		argIdx++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *filter.EndTime)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}

	query := fmt.Sprintf(`
		SELECT id, severity, alert_type, exchange_id, trading_pair, user_id,
			   details, evidence, status, assigned_to, resolved_at, resolution,
			   created_at, updated_at
		FROM %s
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, r.table, whereClause, argIdx, argIdx+1)

	args = append(args, limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()

	alerts, err := r.scanAlerts(rows)
	if err != nil {
		return nil, err
	}

	return alerts, nil
}

// GetAlertByID retrieves a specific alert by ID
func (r *AlertRepository) GetAlertByID(ctx context.Context, alertID string) (*domain.MarketAlert, error) {
	query := fmt.Sprintf(`
		SELECT id, severity, alert_type, exchange_id, trading_pair, user_id,
			   details, evidence, status, assigned_to, resolved_at, resolution,
			   created_at, updated_at
		FROM %s
		WHERE id = $1
	`, r.table)

	row := r.pool.QueryRow(ctx, query, alertID)

	alert, err := r.scanAlert(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("alert not found: %s", alertID)
		}
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return alert, nil
}

// UpdateAlertStatus updates the status of an alert
func (r *AlertRepository) UpdateAlertStatus(ctx context.Context, alertID string, status domain.AlertStatus, resolution string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $1,
			resolution = $2,
			updated_at = $3
		WHERE id = $4
	`, r.table)

	now := time.Now().UTC()
	result, err := r.pool.Exec(ctx, query, status, resolution, now, alertID)
	if err != nil {
		return fmt.Errorf("failed to update alert status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	return nil
}

// GetAlertStats retrieves alert statistics for a time period
func (r *AlertRepository) GetAlertStats(ctx context.Context, start, end time.Time) (ports.AlertStats, error) {
	stats := ports.AlertStats{
		BySeverity: make(map[string]int),
		ByType:     make(map[string]int),
		ByExchange: make(map[string]int),
	}

	query := fmt.Sprintf(`
		SELECT severity, alert_type, exchange_id, COUNT(*)
		FROM %s
		WHERE created_at BETWEEN $1 AND $2
		GROUP BY severity, alert_type, exchange_id
	`, r.table)

	rows, err := r.pool.Query(ctx, query, start, end)
	if err != nil {
		return stats, fmt.Errorf("failed to query alert stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var severity, alertType, exchangeID string
		var count int

		if err := rows.Scan(&severity, &alertType, &exchangeID, &count); err != nil {
			r.logger.Warn("Failed to scan alert stat row", zap.Error(err))
			continue
		}

		stats.BySeverity[severity] += count
		stats.ByType[alertType] += count
		stats.ByExchange[exchangeID] += count
		stats.TotalAlerts += count
	}

	return stats, nil
}

// GetAlertsByExchange retrieves all alerts for a specific exchange
func (r *AlertRepository) GetAlertsByExchange(ctx context.Context, exchangeID string, limit int) ([]domain.MarketAlert, error) {
	query := fmt.Sprintf(`
		SELECT id, severity, alert_type, exchange_id, trading_pair, user_id,
			   details, evidence, status, assigned_to, resolved_at, resolution,
			   created_at, updated_at
		FROM %s
		WHERE exchange_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, r.table)

	rows, err := r.pool.Query(ctx, query, exchangeID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()

	alerts, err := r.scanAlerts(rows)
	if err != nil {
		return nil, err
	}

	return alerts, nil
}

// scanAlerts scans multiple alert rows
func (r *AlertRepository) scanAlerts(rows pgx.Rows) ([]domain.MarketAlert, error) {
	alerts := make([]domain.MarketAlert, 0)

	for rows.Next() {
		alert, err := r.scanAlert(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, *alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error scanning alerts: %w", err)
	}

	return alerts, nil
}

// scanAlert scans a single alert row
func (r *AlertRepository) scanAlert(row pgx.Row) (*domain.MarketAlert, error) {
	var alert domain.MarketAlert
	var detailsJSON, evidenceJSON []byte

	err := row.Scan(
		&alert.ID,
		&alert.Severity,
		&alert.AlertType,
		&alert.ExchangeID,
		&alert.TradingPair,
		&alert.UserID,
		&detailsJSON,
		&evidenceJSON,
		&alert.Status,
		&alert.AssignedTo,
		&alert.ResolvedAt,
		&alert.Resolution,
		&alert.CreatedAt,
		&alert.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(detailsJSON, &alert.Details); err != nil {
		r.logger.Warn("Failed to unmarshal alert details", zap.Error(err))
	}

	if err := json.Unmarshal(evidenceJSON, &alert.Evidence); err != nil {
		r.logger.Warn("Failed to unmarshal alert evidence", zap.Error(err))
	}

	return &alert, nil
}
