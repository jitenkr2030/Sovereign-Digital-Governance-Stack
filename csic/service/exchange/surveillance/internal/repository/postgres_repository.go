package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/surveillance/internal/domain"
	"github.com/csic/surveillance/internal/port"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresConfig holds PostgreSQL connection configuration
type PostgresConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// PostgresAlertRepository implements AlertRepository for PostgreSQL
type PostgresAlertRepository struct {
	db          *sql.DB
	stmtCache   map[string]*sql.Stmt
	stmtCacheMu sync.RWMutex
}

// NewPostgresAlertRepository creates a new PostgreSQL alert repository
func NewPostgresAlertRepository(config PostgresConfig) (*PostgresAlertRepository, error) {
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

	repo := &PostgresAlertRepository{
		db:        db,
		stmtCache: make(map[string]*sql.Stmt),
	}

	// Initialize prepared statements
	if err := repo.initStatements(); err != nil {
		return nil, fmt.Errorf("failed to initialize statements: %w", err)
	}

	return repo, nil
}

// initStatements creates prepared statements for common operations
func (r *PostgresAlertRepository) initStatements() error {
	statements := []string{
		`SELECT id, type, severity, status, exchange_id, symbol, account_id, order_id, trade_id,
			title, description, evidence, risk_score, pattern_confidence, detected_at, updated_at,
			resolved_at, resolved_by, resolution, assigned_to, tags, related_alert_ids, audit_hash
		 FROM alerts WHERE id = $1`,
		`INSERT INTO alerts (id, type, severity, status, exchange_id, symbol, account_id, order_id, trade_id,
			title, description, evidence, risk_score, pattern_confidence, detected_at, updated_at, tags, related_alert_ids, audit_hash)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		`UPDATE alerts SET status = $1, updated_at = $2, resolved_at = $3, resolved_by = $4, resolution = $5 WHERE id = $6`,
	}

	stmtNames := []string{"getByID", "create", "updateStatus"}

	for i, stmt := range statements {
		preparedStmt, err := r.db.Prepare(stmt)
		if err != nil {
			return err
		}
		r.stmtCacheMu.Lock()
		r.stmtCache[stmtNames[i]] = preparedStmt
		r.stmtCacheMu.Unlock()
	}

	return nil
}

// Close closes the database connection
func (r *PostgresAlertRepository) Close() error {
	r.stmtCacheMu.RLock()
	defer r.stmtCacheMu.RUnlock()

	for _, stmt := range r.stmtCache {
		stmt.Close()
	}

	return r.db.Close()
}

// Create creates a new alert
func (r *PostgresAlertRepository) Create(ctx context.Context, alert *domain.Alert) error {
	evidenceJSON, err := json.Marshal(alert.Evidence)
	if err != nil {
		return fmt.Errorf("failed to marshal evidence: %w", err)
	}

	tagsJSON, err := json.Marshal(alert.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	relatedAlertsJSON, err := json.Marshal(alert.RelatedAlertIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal related alerts: %w", err)
	}

	query := `INSERT INTO alerts (
		id, type, severity, status, exchange_id, symbol, account_id, order_id, trade_id,
		title, description, evidence, risk_score, pattern_confidence, detected_at, updated_at,
		resolved_at, resolved_by, resolution, assigned_to, tags, related_alert_ids, audit_hash
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`

	_, err = r.db.ExecContext(ctx, query,
		alert.ID, alert.Type, alert.Severity, alert.Status, alert.ExchangeID, alert.Symbol,
		alert.AccountID, alert.OrderID, alert.TradeID, alert.Title, alert.Description,
		evidenceJSON, alert.RiskScore, alert.PatternConfidence, alert.DetectedAt, alert.UpdatedAt,
		alert.ResolvedAt, alert.ResolvedBy, alert.Resolution, alert.AssignedTo,
		tagsJSON, relatedAlertsJSON, alert.AuditHash,
	)

	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

// GetByID retrieves an alert by ID
func (r *PostgresAlertRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Alert, error) {
	query := `SELECT id, type, severity, status, exchange_id, symbol, account_id, order_id, trade_id,
		title, description, evidence, risk_score, pattern_confidence, detected_at, updated_at,
		resolved_at, resolved_by, resolution, assigned_to, tags, related_alert_ids, audit_hash
		FROM alerts WHERE id = $1`

	alert := &domain.Alert{}
	var evidenceJSON, tagsJSON, relatedAlertsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&alert.ID, &alert.Type, &alert.Severity, &alert.Status, &alert.ExchangeID, &alert.Symbol,
		&alert.AccountID, &alert.OrderID, &alert.TradeID, &alert.Title, &alert.Description,
		&evidenceJSON, &alert.RiskScore, &alert.PatternConfidence, &alert.DetectedAt, &alert.UpdatedAt,
		&alert.ResolvedAt, &alert.ResolvedBy, &alert.Resolution, &alert.AssignedTo,
		&tagsJSON, &relatedAlertsJSON, &alert.AuditHash,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(evidenceJSON, &alert.Evidence); err != nil {
		return nil, fmt.Errorf("failed to unmarshal evidence: %w", err)
	}
	if err := json.Unmarshal(tagsJSON, &alert.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}
	if err := json.Unmarshal(relatedAlertsJSON, &alert.RelatedAlertIDs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal related alerts: %w", err)
	}

	return alert, nil
}

// Update updates an existing alert
func (r *PostgresAlertRepository) Update(ctx context.Context, alert *domain.Alert) error {
	evidenceJSON, err := json.Marshal(alert.Evidence)
	if err != nil {
		return fmt.Errorf("failed to marshal evidence: %w", err)
	}

	tagsJSON, err := json.Marshal(alert.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	relatedAlertsJSON, err := json.Marshal(alert.RelatedAlertIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal related alerts: %w", err)
	}

	query := `UPDATE alerts SET
		type = $1, severity = $2, status = $3, exchange_id = $4, symbol = $5, account_id = $6,
		order_id = $7, trade_id = $8, title = $9, description = $10, evidence = $11,
		risk_score = $12, pattern_confidence = $13, detected_at = $14, updated_at = $15,
		resolved_at = $16, resolved_by = $17, resolution = $18, assigned_to = $19,
		tags = $20, related_alert_ids = $21, audit_hash = $22 WHERE id = $23`

	_, err = r.db.ExecContext(ctx, query,
		alert.Type, alert.Severity, alert.Status, alert.ExchangeID, alert.Symbol,
		alert.AccountID, alert.OrderID, alert.TradeID, alert.Title, alert.Description,
		evidenceJSON, alert.RiskScore, alert.PatternConfidence, alert.DetectedAt, alert.UpdatedAt,
		alert.ResolvedAt, alert.ResolvedBy, alert.Resolution, alert.AssignedTo,
		tagsJSON, relatedAlertsJSON, alert.AuditHash, alert.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	return nil
}

// Delete deletes an alert
func (r *PostgresAlertRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM alerts WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}
	return nil
}

// List retrieves alerts based on filter criteria
func (r *PostgresAlertRepository) List(ctx context.Context, filter domain.AlertFilter, limit, offset int) ([]domain.Alert, error) {
	query := `SELECT id, type, severity, status, exchange_id, symbol, account_id, order_id, trade_id,
		title, description, evidence, risk_score, pattern_confidence, detected_at, updated_at,
		resolved_at, resolved_by, resolution, assigned_to, tags, related_alert_ids, audit_hash
		FROM alerts WHERE 1=1`

	args := []interface{}{}
	argNum := 1

	// Apply filters
	if len(filter.Types) > 0 {
		query += fmt.Sprintf(" AND type = ANY($%d)", argNum)
		args = append(args, filter.Types)
		argNum++
	}

	if len(filter.Severities) > 0 {
		query += fmt.Sprintf(" AND severity = ANY($%d)", argNum)
		args = append(args, filter.Severities)
		argNum++
	}

	if len(filter.Statuses) > 0 {
		query += fmt.Sprintf(" AND status = ANY($%d)", argNum)
		args = append(args, filter.Statuses)
		argNum++
	}

	if len(filter.ExchangeIDs) > 0 {
		query += fmt.Sprintf(" AND exchange_id = ANY($%d)", argNum)
		args = append(args, filter.ExchangeIDs)
		argNum++
	}

	if len(filter.Symbols) > 0 {
		query += fmt.Sprintf(" AND symbol = ANY($%d)", argNum)
		args = append(args, filter.Symbols)
		argNum++
	}

	if filter.DetectedAfter != nil {
		query += fmt.Sprintf(" AND detected_at >= $%d", argNum)
		args = append(args, *filter.DetectedAfter)
		argNum++
	}

	if filter.DetectedBefore != nil {
		query += fmt.Sprintf(" AND detected_at <= $%d", argNum)
		args = append(args, *filter.DetectedBefore)
		argNum++
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY detected_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []domain.Alert
	for rows.Next() {
		alert := &domain.Alert{}
		var evidenceJSON, tagsJSON, relatedAlertsJSON []byte

		err := rows.Scan(
			&alert.ID, &alert.Type, &alert.Severity, &alert.Status, &alert.ExchangeID, &alert.Symbol,
			&alert.AccountID, &alert.OrderID, &alert.TradeID, &alert.Title, &alert.Description,
			&evidenceJSON, &alert.RiskScore, &alert.PatternConfidence, &alert.DetectedAt, &alert.UpdatedAt,
			&alert.ResolvedAt, &alert.ResolvedBy, &alert.Resolution, &alert.AssignedTo,
			&tagsJSON, &relatedAlertsJSON, &alert.AuditHash,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}

		if err := json.Unmarshal(evidenceJSON, &alert.Evidence); err != nil {
			return nil, fmt.Errorf("failed to unmarshal evidence: %w", err)
		}
		if err := json.Unmarshal(tagsJSON, &alert.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
		if err := json.Unmarshal(relatedAlertsJSON, &alert.RelatedAlertIDs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal related alerts: %w", err)
		}

		alerts = append(alerts, *alert)
	}

	return alerts, nil
}

// Count returns the count of alerts matching the filter
func (r *PostgresAlertRepository) Count(ctx context.Context, filter domain.AlertFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM alerts WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if len(filter.Types) > 0 {
		query += fmt.Sprintf(" AND type = ANY($%d)", argNum)
		args = append(args, filter.Types)
		argNum++
	}

	if len(filter.Severities) > 0 {
		query += fmt.Sprintf(" AND severity = ANY($%d)", argNum)
		args = append(args, filter.Severities)
		argNum++
	}

	if len(filter.Statuses) > 0 {
		query += fmt.Sprintf(" AND status = ANY($%d)", argNum)
		args = append(args, filter.Statuses)
		argNum++
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	return count, nil
}

// GetByStatus retrieves alerts by status
func (r *PostgresAlertRepository) GetByStatus(ctx context.Context, status domain.AlertStatus) ([]domain.Alert, error) {
	return r.List(ctx, domain.AlertFilter{Statuses: []domain.AlertStatus{status}}, 100, 0)
}

// GetByExchange retrieves alerts for a specific exchange
func (r *PostgresAlertRepository) GetByExchange(ctx context.Context, exchangeID uuid.UUID, limit, offset int) ([]domain.Alert, error) {
	return r.List(ctx, domain.AlertFilter{ExchangeIDs: []uuid.UUID{exchangeID}}, limit, offset)
}

// GetBySymbol retrieves alerts for a specific symbol
func (r *PostgresAlertRepository) GetBySymbol(ctx context.Context, symbol string, limit, offset int) ([]domain.Alert, error) {
	return r.List(ctx, domain.AlertFilter{Symbols: []string{symbol}}, limit, offset)
}

// GetOpenAlerts retrieves all open alerts
func (r *PostgresAlertRepository) GetOpenAlerts(ctx context.Context) ([]domain.Alert, error) {
	return r.List(ctx, domain.AlertFilter{
		Statuses: []domain.AlertStatus{
			domain.AlertStatusOpen,
			domain.AlertStatusInvestigating,
		},
	}, 500, 0)
}

// GetRecentAlerts retrieves alerts since a specific time
func (r *PostgresAlertRepository) GetRecentAlerts(ctx context.Context, since time.Time) ([]domain.Alert, error) {
	return r.List(ctx, domain.AlertFilter{DetectedAfter: &since}, 100, 0)
}

// GetStatistics retrieves alert statistics for a time period
func (r *PostgresAlertRepository) GetStatistics(ctx context.Context, startTime, endTime time.Time) (*domain.AlertStatistics, error) {
	// Placeholder implementation
	stats := &domain.AlertStatistics{
		AlertsByType:     make(map[domain.AlertType]int64),
		AlertsBySeverity: make(map[domain.AlertSeverity]int64),
		AlertsByStatus:   make(map[domain.AlertStatus]int64),
	}

	return stats, nil
}

// GetAlertTrend retrieves alert trend data
func (r *PostgresAlertRepository) GetAlertTrend(ctx context.Context, interval string, startTime, endTime time.Time) ([]domain.AlertTrendPoint, error) {
	// Placeholder implementation
	return []domain.AlertTrendPoint{}, nil
}

// BulkCreate creates multiple alerts
func (r *PostgresAlertRepository) BulkCreate(ctx context.Context, alerts []domain.Alert) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO alerts (
		id, type, severity, status, exchange_id, symbol, account_id, order_id, trade_id,
		title, description, evidence, risk_score, pattern_confidence, detected_at, updated_at,
		tags, related_alert_ids, audit_hash
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, alert := range alerts {
		evidenceJSON, _ := json.Marshal(alert.Evidence)
		tagsJSON, _ := json.Marshal(alert.Tags)
		relatedAlertsJSON, _ := json.Marshal(alert.RelatedAlertIDs)

		_, err := stmt.ExecContext(ctx,
			alert.ID, alert.Type, alert.Severity, alert.Status, alert.ExchangeID, alert.Symbol,
			alert.AccountID, alert.OrderID, alert.TradeID, alert.Title, alert.Description,
			evidenceJSON, alert.RiskScore, alert.PatternConfidence, alert.DetectedAt, alert.UpdatedAt,
			tagsJSON, relatedAlertsJSON, alert.AuditHash,
		)
		if err != nil {
			return fmt.Errorf("failed to insert alert: %w", err)
		}
	}

	return tx.Commit()
}

// BulkUpdateStatus updates the status of multiple alerts
func (r *PostgresAlertRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status domain.AlertStatus) error {
	query := `UPDATE alerts SET status = $1, updated_at = $2 WHERE id = ANY($3)`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), ids)
	if err != nil {
		return fmt.Errorf("failed to bulk update status: %w", err)
	}
	return nil
}

// Ensure the interface is satisfied
var _ port.AlertRepository = (*PostgresAlertRepository)(nil)
