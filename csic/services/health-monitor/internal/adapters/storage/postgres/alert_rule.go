package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"csic-platform/health-monitor/internal/core/domain"
	"csic-platform/health-monitor/internal/core/ports"
)

// PostgresAlertRuleRepository implements AlertRuleRepository using PostgreSQL
type PostgresAlertRuleRepository struct {
	db          *sql.DB
	tablePrefix string
}

// NewPostgresAlertRuleRepository creates a new PostgreSQL alert rule repository
func NewPostgresAlertRuleRepository(databaseURL string) (ports.AlertRuleRepository, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresAlertRuleRepository{
		db:          db,
		tablePrefix: "health_monitor_",
	}, nil
}

// Close closes the database connection
func (r *PostgresAlertRuleRepository) Close() error {
	return r.db.Close()
}

// tableName returns the prefixed table name
func (r *PostgresAlertRuleRepository) tableName(name string) string {
	return r.tablePrefix + name
}

// GetAlertRule retrieves an alert rule by ID
func (r *PostgresAlertRuleRepository) GetAlertRule(ctx context.Context, id string) (*domain.AlertRule, error) {
	query := fmt.Sprintf(`
		SELECT id, name, service_name, condition, threshold, duration, severity,
		       enabled, cooldown, notification_channels, created_at, updated_at
		FROM %s
		WHERE id = $1
	`, r.tableName("alert_rules"))

	var rule domain.AlertRule
	var notificationChannels []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID,
		&rule.Name,
		&rule.ServiceName,
		&rule.Condition,
		&rule.Threshold,
		&rule.Duration,
		&rule.Severity,
		&rule.Enabled,
		&rule.Cooldown,
		&notificationChannels,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	if notificationChannels != nil {
		_ = notificationChannels // Would parse JSON in production
	}

	return &rule, nil
}

// GetAllAlertRules retrieves all alert rules
func (r *PostgresAlertRuleRepository) GetAllAlertRules(ctx context.Context) ([]*domain.AlertRule, error) {
	query := fmt.Sprintf(`
		SELECT id, name, service_name, condition, threshold, duration, severity,
		       enabled, cooldown, notification_channels, created_at, updated_at
		FROM %s
		ORDER BY name ASC
	`, r.tableName("alert_rules"))

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.AlertRule

	for rows.Next() {
		var rule domain.AlertRule
		var notificationChannels []byte

		if err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.ServiceName,
			&rule.Condition,
			&rule.Threshold,
			&rule.Duration,
			&rule.Severity,
			&rule.Enabled,
			&rule.Cooldown,
			&notificationChannels,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan alert rule: %w", err)
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}

// GetAlertRulesByService retrieves alert rules for a specific service
func (r *PostgresAlertRuleRepository) GetAlertRulesByService(ctx context.Context, serviceName string) ([]*domain.AlertRule, error) {
	query := fmt.Sprintf(`
		SELECT id, name, service_name, condition, threshold, duration, severity,
		       enabled, cooldown, notification_channels, created_at, updated_at
		FROM %s
		WHERE service_name = $1
		ORDER BY severity DESC, name ASC
	`, r.tableName("alert_rules"))

	rows, err := r.db.QueryContext(ctx, query, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.AlertRule

	for rows.Next() {
		var rule domain.AlertRule
		var notificationChannels []byte

		if err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.ServiceName,
			&rule.Condition,
			&rule.Threshold,
			&rule.Duration,
			&rule.Severity,
			&rule.Enabled,
			&rule.Cooldown,
			&notificationChannels,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan alert rule: %w", err)
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}

// GetEnabledAlertRules retrieves all enabled alert rules
func (r *PostgresAlertRuleRepository) GetEnabledAlertRules(ctx context.Context) ([]*domain.AlertRule, error) {
	query := fmt.Sprintf(`
		SELECT id, name, service_name, condition, threshold, duration, severity,
		       enabled, cooldown, notification_channels, created_at, updated_at
		FROM %s
		WHERE enabled = true
		ORDER BY severity DESC, name ASC
	`, r.tableName("alert_rules"))

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.AlertRule

	for rows.Next() {
		var rule domain.AlertRule
		var notificationChannels []byte

		if err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.ServiceName,
			&rule.Condition,
			&rule.Threshold,
			&rule.Duration,
			&rule.Severity,
			&rule.Enabled,
			&rule.Cooldown,
			&notificationChannels,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan alert rule: %w", err)
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}

// CreateAlertRule creates a new alert rule
func (r *PostgresAlertRuleRepository) CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, name, service_name, condition, threshold, duration, severity,
		                enabled, cooldown, notification_channels, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, r.tableName("alert_rules"))

	_, err := r.db.ExecContext(ctx, query,
		rule.ID,
		rule.Name,
		rule.ServiceName,
		rule.Condition,
		rule.Threshold,
		rule.Duration,
		rule.Severity,
		rule.Enabled,
		rule.Cooldown,
		[]byte("[]"), // notification_channels as JSON
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create alert rule: %w", err)
	}

	return nil
}

// UpdateAlertRule updates an alert rule
func (r *PostgresAlertRuleRepository) UpdateAlertRule(ctx context.Context, id string, rule *domain.AlertRule) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET name = $1, service_name = $2, condition = $3, threshold = $4, duration = $5,
		    severity = $6, enabled = $7, cooldown = $8, notification_channels = $9, updated_at = $10
		WHERE id = $11
	`, r.tableName("alert_rules"))

	result, err := r.db.ExecContext(ctx, query,
		rule.Name,
		rule.ServiceName,
		rule.Condition,
		rule.Threshold,
		rule.Duration,
		rule.Severity,
		rule.Enabled,
		rule.Cooldown,
		[]byte("[]"),
		time.Now(),
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update alert rule: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("alert rule not found: %s", id)
	}

	return nil
}

// DeleteAlertRule deletes an alert rule
func (r *PostgresAlertRuleRepository) DeleteAlertRule(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, r.tableName("alert_rules"))

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	return nil
}
