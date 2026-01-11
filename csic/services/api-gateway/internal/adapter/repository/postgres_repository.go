package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/csic-platform/services/api-gateway/internal/config"
	"github.com/csic-platform/services/api-gateway/internal/core/domain"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresRepository implements the Repository interface using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository instance
func NewPostgresRepository(cfg *config.DatabaseConfig) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.GetConnMaxLifetime())

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// DB returns the underlying database connection
func (r *PostgresRepository) DB() *sql.DB {
	return r.db
}

// Exchange operations

func (r *PostgresRepository) GetExchanges(ctx context.Context, page, pageSize int) ([]*domain.Exchange, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT id, name, license_number, status, jurisdiction, website, contact_email,
		       compliance_score, risk_level, registration_date, last_audit, next_audit,
		       created_at, updated_at
		FROM exchanges
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query exchanges: %w", err)
	}
	defer rows.Close()

	var exchanges []*domain.Exchange
	for rows.Next() {
		var e domain.Exchange
		err := rows.Scan(
			&e.ID, &e.Name, &e.LicenseNumber, &e.Status, &e.Jurisdiction,
			&e.Website, &e.ContactEmail, &e.ComplianceScore, &e.RiskLevel,
			&e.RegistrationDate, &e.LastAudit, &e.NextAudit, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan exchange: %w", err)
		}
		exchanges = append(exchanges, &e)
	}

	return exchanges, nil
}

func (r *PostgresRepository) GetExchangeByID(ctx context.Context, id string) (*domain.Exchange, error) {
	query := `
		SELECT id, name, license_number, status, jurisdiction, website, contact_email,
		       compliance_score, risk_level, registration_date, last_audit, next_audit,
		       created_at, updated_at
		FROM exchanges WHERE id = $1
	`

	var e domain.Exchange
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID, &e.Name, &e.LicenseNumber, &e.Status, &e.Jurisdiction,
		&e.Website, &e.ContactEmail, &e.ComplianceScore, &e.RiskLevel,
		&e.RegistrationDate, &e.LastAudit, &e.NextAudit, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("exchange not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange: %w", err)
	}

	return &e, nil
}

func (r *PostgresRepository) CreateExchange(ctx context.Context, exchange *domain.Exchange) error {
	if exchange.ID == "" {
		exchange.ID = uuid.New().String()
	}
	exchange.CreatedAt = time.Now()
	exchange.UpdatedAt = time.Now()

	query := `
		INSERT INTO exchanges (id, name, license_number, status, jurisdiction, website,
		                       contact_email, compliance_score, risk_level, registration_date,
		                       last_audit, next_audit, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.ExecContext(ctx, query,
		exchange.ID, exchange.Name, exchange.LicenseNumber, exchange.Status,
		exchange.Jurisdiction, exchange.Website, exchange.ContactEmail,
		exchange.ComplianceScore, exchange.RiskLevel, exchange.RegistrationDate,
		exchange.LastAudit, exchange.NextAudit, exchange.CreatedAt, exchange.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) UpdateExchange(ctx context.Context, exchange *domain.Exchange) error {
	exchange.UpdatedAt = time.Now()

	query := `
		UPDATE exchanges SET name=$1, status=$2, compliance_score=$3, risk_level=$4,
		       last_audit=$5, updated_at=$6 WHERE id=$7
	`

	_, err := r.db.ExecContext(ctx, query,
		exchange.Name, exchange.Status, exchange.ComplianceScore,
		exchange.RiskLevel, exchange.LastAudit, exchange.UpdatedAt, exchange.ID,
	)
	return err
}

func (r *PostgresRepository) SuspendExchange(ctx context.Context, id, reason, userID string) error {
	query := `
		UPDATE exchanges SET status=$1, updated_at=$2 WHERE id=$3
	`
	_, err := r.db.ExecContext(ctx, query, domain.ExchangeStatusSuspended, time.Now(), id)
	return err
}

// Wallet operations

func (r *PostgresRepository) GetWallets(ctx context.Context, page, pageSize int) ([]*domain.Wallet, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT id, address, label, type, status, risk_score, first_seen, last_activity,
		       created_at, updated_at
		FROM wallets ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query wallets: %w", err)
	}
	defer rows.Close()

	var wallets []*domain.Wallet
	for rows.Next() {
		var w domain.Wallet
		err := rows.Scan(
			&w.ID, &w.Address, &w.Label, &w.Type, &w.Status, &w.RiskScore,
			&w.FirstSeen, &w.LastActivity, &w.CreatedAt, &w.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan wallet: %w", err)
		}
		wallets = append(wallets, &w)
	}

	return wallets, nil
}

func (r *PostgresRepository) GetWalletByID(ctx context.Context, id string) (*domain.Wallet, error) {
	query := `
		SELECT id, address, label, type, status, risk_score, first_seen, last_activity,
		       created_at, updated_at FROM wallets WHERE id=$1
	`

	var w domain.Wallet
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&w.ID, &w.Address, &w.Label, &w.Type, &w.Status, &w.RiskScore,
		&w.FirstSeen, &w.LastActivity, &w.CreatedAt, &w.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("wallet not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return &w, nil
}

func (r *PostgresRepository) GetWalletByAddress(ctx context.Context, address string) (*domain.Wallet, error) {
	query := `
		SELECT id, address, label, type, status, risk_score, first_seen, last_activity,
		       created_at, updated_at FROM wallets WHERE address=$1
	`

	var w domain.Wallet
	err := r.db.QueryRowContext(ctx, query, address).Scan(
		&w.ID, &w.Address, &w.Label, &w.Type, &w.Status, &w.RiskScore,
		&w.FirstSeen, &w.LastActivity, &w.CreatedAt, &w.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet by address: %w", err)
	}

	return &w, nil
}

func (r *PostgresRepository) CreateWallet(ctx context.Context, wallet *domain.Wallet) error {
	if wallet.ID == "" {
		wallet.ID = uuid.New().String()
	}
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()

	query := `
		INSERT INTO wallets (id, address, label, type, status, risk_score, first_seen,
		                    last_activity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		wallet.ID, wallet.Address, wallet.Label, wallet.Type, wallet.Status,
		wallet.RiskScore, wallet.FirstSeen, wallet.LastActivity,
		wallet.CreatedAt, wallet.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) UpdateWallet(ctx context.Context, wallet *domain.Wallet) error {
	wallet.UpdatedAt = time.Now()

	query := `
		UPDATE wallets SET label=$1, status=$2, risk_score=$3, updated_at=$4 WHERE id=$5
	`
	_, err := r.db.ExecContext(ctx, query,
		wallet.Label, wallet.Status, wallet.RiskScore, wallet.UpdatedAt, wallet.ID,
	)
	return err
}

func (r *PostgresRepository) FreezeWallet(ctx context.Context, id, reason, userID string) error {
	query := `
		UPDATE wallets SET status=$1, updated_at=$2 WHERE id=$3
	`
	_, err := r.db.ExecContext(ctx, query, domain.WalletStatusFrozen, time.Now(), id)
	return err
}

// Miner operations

func (r *PostgresRepository) GetMiners(ctx context.Context, page, pageSize int) ([]*domain.Miner, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT id, name, license_number, status, jurisdiction, hash_rate, energy_consumption,
		       energy_source, compliance_status, registration_date, last_inspection,
		       created_at, updated_at
		FROM miners ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query miners: %w", err)
	}
	defer rows.Close()

	var miners []*domain.Miner
	for rows.Next() {
		var m domain.Miner
		err := rows.Scan(
			&m.ID, &m.Name, &m.LicenseNumber, &m.Status, &m.Jurisdiction,
			&m.HashRate, &m.EnergyConsumption, &m.EnergySource, &m.ComplianceStatus,
			&m.RegistrationDate, &m.LastInspection, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan miner: %w", err)
		}
		miners = append(miners, &m)
	}

	return miners, nil
}

func (r *PostgresRepository) GetMinerByID(ctx context.Context, id string) (*domain.Miner, error) {
	query := `
		SELECT id, name, license_number, status, jurisdiction, hash_rate, energy_consumption,
		       energy_source, compliance_status, registration_date, last_inspection,
		       created_at, updated_at FROM miners WHERE id=$1
	`

	var m domain.Miner
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.Name, &m.LicenseNumber, &m.Status, &m.Jurisdiction,
		&m.HashRate, &m.EnergyConsumption, &m.EnergySource, &m.ComplianceStatus,
		&m.RegistrationDate, &m.LastInspection, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("miner not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get miner: %w", err)
	}

	return &m, nil
}

func (r *PostgresRepository) CreateMiner(ctx context.Context, miner *domain.Miner) error {
	if miner.ID == "" {
		miner.ID = uuid.New().String()
	}
	miner.CreatedAt = time.Now()
	miner.UpdatedAt = time.Now()

	query := `
		INSERT INTO miners (id, name, license_number, status, jurisdiction, hash_rate,
		                   energy_consumption, energy_source, compliance_status,
		                   registration_date, last_inspection, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		miner.ID, miner.Name, miner.LicenseNumber, miner.Status, miner.Jurisdiction,
		miner.HashRate, miner.EnergyConsumption, miner.EnergySource, miner.ComplianceStatus,
		miner.RegistrationDate, miner.LastInspection, miner.CreatedAt, miner.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) UpdateMiner(ctx context.Context, miner *domain.Miner) error {
	miner.UpdatedAt = time.Now()

	query := `
		UPDATE miners SET name=$1, status=$2, hash_rate=$3, compliance_status=$4,
		       updated_at=$5 WHERE id=$6
	`
	_, err := r.db.ExecContext(ctx, query,
		miner.Name, miner.Status, miner.HashRate, miner.ComplianceStatus,
		miner.UpdatedAt, miner.ID,
	)
	return err
}

// Alert operations

func (r *PostgresRepository) GetAlerts(ctx context.Context, page, pageSize int) ([]*domain.Alert, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT id, title, description, severity, status, category, source, evidence,
		       acknowledged_by, acknowledged_at, created_at, updated_at
		FROM alerts ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*domain.Alert
	for rows.Next() {
		var a domain.Alert
		var evidence sql.NullString
		var acknowledgedBy, acknowledgedAt sql.NullString
		err := rows.Scan(
			&a.ID, &a.Title, &a.Description, &a.Severity, &a.Status,
			&a.Category, &a.Source, &evidence, &acknowledgedBy, &acknowledgedAt,
			&a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		if evidence.Valid {
			_ = json.Unmarshal([]byte(evidence.String), &a.Evidence)
		}
		if acknowledgedBy.Valid {
			a.AcknowledgedBy = acknowledgedBy.String
		}
		if acknowledgedAt.Valid {
			a.AcknowledgedAt = acknowledgedAt.String
		}
		alerts = append(alerts, &a)
	}

	return alerts, nil
}

func (r *PostgresRepository) GetAlertByID(ctx context.Context, id string) (*domain.Alert, error) {
	query := `
		SELECT id, title, description, severity, status, category, source, evidence,
		       acknowledged_by, acknowledged_at, created_at, updated_at
		FROM alerts WHERE id=$1
	`

	var a domain.Alert
	var evidence sql.NullString
	var acknowledgedBy, acknowledgedAt sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.Title, &a.Description, &a.Severity, &a.Status,
		&a.Category, &a.Source, &evidence, &acknowledgedBy, &acknowledgedAt,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("alert not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	if evidence.Valid {
		_ = json.Unmarshal([]byte(evidence.String), &a.Evidence)
	}
	if acknowledgedBy.Valid {
		a.AcknowledgedBy = acknowledgedBy.String
	}
	if acknowledgedAt.Valid {
		a.AcknowledgedAt = acknowledgedAt.String
	}

	return &a, nil
}

func (r *PostgresRepository) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()

	evidenceJSON, _ := json.Marshal(alert.Evidence)

	query := `
		INSERT INTO alerts (id, title, description, severity, status, category, source,
		                   evidence, acknowledged_by, acknowledged_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		alert.ID, alert.Title, alert.Description, alert.Severity, alert.Status,
		alert.Category, alert.Source, evidenceJSON, alert.AcknowledgedBy,
		alert.AcknowledgedAt, alert.CreatedAt, alert.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) UpdateAlert(ctx context.Context, alert *domain.Alert) error {
	alert.UpdatedAt = time.Now()
	evidenceJSON, _ := json.Marshal(alert.Evidence)

	query := `
		UPDATE alerts SET title=$1, description=$2, severity=$3, status=$4,
		       evidence=$5, acknowledged_by=$6, acknowledged_at=$7, updated_at=$8 WHERE id=$9
	`
	_, err := r.db.ExecContext(ctx, query,
		alert.Title, alert.Description, alert.Severity, alert.Status,
		evidenceJSON, alert.AcknowledgedBy, alert.AcknowledgedAt,
		alert.UpdatedAt, alert.ID,
	)
	return err
}

func (r *PostgresRepository) AcknowledgeAlert(ctx context.Context, id, userID string) error {
	query := `
		UPDATE alerts SET status=$1, acknowledged_by=$2, acknowledged_at=$3, updated_at=$4
		WHERE id=$5
	`
	_, err := r.db.ExecContext(ctx, query,
		domain.AlertStatusAcknowledged, userID,
		time.Now().UTC().Format(time.RFC3339), time.Now(), id,
	)
	return err
}

// Compliance report operations

func (r *PostgresRepository) GetComplianceReports(ctx context.Context, page, pageSize int) ([]*domain.ComplianceReport, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT id, entity_type, entity_id, entity_name, period, status, score,
		       generated_at, created_at, updated_at
		FROM compliance_reports ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query compliance reports: %w", err)
	}
	defer rows.Close()

	var reports []*domain.ComplianceReport
	for rows.Next() {
		var rpt domain.ComplianceReport
		err := rows.Scan(
			&rpt.ID, &rpt.EntityType, &rpt.EntityID, &rpt.EntityName,
			&rpt.Period, &rpt.Status, &rpt.Score, &rpt.GeneratedAt,
			&rpt.CreatedAt, &rpt.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compliance report: %w", err)
		}
		reports = append(reports, &rpt)
	}

	return reports, nil
}

func (r *PostgresRepository) GetComplianceReportByID(ctx context.Context, id string) (*domain.ComplianceReport, error) {
	query := `
		SELECT id, entity_type, entity_id, entity_name, period, status, score,
		       generated_at, created_at, updated_at
		FROM compliance_reports WHERE id=$1
	`

	var rpt domain.ComplianceReport
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rpt.ID, &rpt.EntityType, &rpt.EntityID, &rpt.EntityName,
		&rpt.Period, &rpt.Status, &rpt.Score, &rpt.GeneratedAt,
		&rpt.CreatedAt, &rpt.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("compliance report not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance report: %w", err)
	}

	return &rpt, nil
}

func (r *PostgresRepository) CreateComplianceReport(ctx context.Context, report *domain.ComplianceReport) error {
	if report.ID == "" {
		report.ID = uuid.New().String()
	}
	report.CreatedAt = time.Now()
	report.UpdatedAt = time.Now()

	query := `
		INSERT INTO compliance_reports (id, entity_type, entity_id, entity_name, period,
		                               status, score, generated_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		report.ID, report.EntityType, report.EntityID, report.EntityName,
		report.Period, report.Status, report.Score, report.GeneratedAt,
		report.CreatedAt, report.UpdatedAt,
	)
	return err
}

// Audit log operations

func (r *PostgresRepository) GetAuditLogs(ctx context.Context, page, pageSize int) ([]*domain.AuditLog, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT id, timestamp, user_id, username, action, resource_type, details,
		       ip_address, status, created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		err := rows.Scan(
			&log.ID, &log.Timestamp, &log.UserID, &log.Username,
			&log.Action, &log.ResourceType, &log.Details, &log.IPAddress,
			&log.Status, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

func (r *PostgresRepository) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	log.CreatedAt = time.Now()

	query := `
		INSERT INTO audit_logs (id, timestamp, user_id, username, action, resource_type,
		                       details, ip_address, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.Timestamp, log.UserID, log.Username, log.Action,
		log.ResourceType, log.Details, log.IPAddress, log.Status, log.CreatedAt,
	)
	return err
}

// User operations

func (r *PostgresRepository) GetUsers(ctx context.Context, page, pageSize int) ([]*domain.User, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT id, username, email, role, status, password, last_login, mfa_enabled,
		       created_at, updated_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		var password sql.NullString
		var lastLogin sql.NullString
		err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.Role, &u.Status,
			&password, &lastLogin, &u.MFAEnabled, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		if password.Valid {
			u.Password = password.String
		}
		if lastLogin.Valid {
			u.LastLogin = lastLogin.String
		}
		users = append(users, &u)
	}

	return users, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, username, email, role, status, password, last_login, mfa_enabled,
		       permissions, created_at, updated_at
		FROM users WHERE id=$1
	`

	var u domain.User
	var password sql.NullString
	var lastLogin sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.Role, &u.Status,
		&password, &lastLogin, &u.MFAEnabled, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if password.Valid {
		u.Password = password.String
	}
	if lastLogin.Valid {
		u.LastLogin = lastLogin.String
	}

	return &u, nil
}

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, role, status, password, last_login, mfa_enabled,
		       permissions, created_at, updated_at
		FROM users WHERE username=$1
	`

	var u domain.User
	var password sql.NullString
	var lastLogin sql.NullString
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&u.ID, &u.Username, &u.Email, &u.Role, &u.Status,
		&password, &lastLogin, &u.MFAEnabled, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	if password.Valid {
		u.Password = password.String
	}
	if lastLogin.Valid {
		u.LastLogin = lastLogin.String
	}

	return &u, nil
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *domain.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, username, email, role, status, password, last_login,
		                  mfa_enabled, permissions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.Role, user.Status,
		user.Password, user.LastLogin, user.MFAEnabled, user.Permissions,
		user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users SET email=$1, role=$2, status=$3, last_login=$4, mfa_enabled=$5,
		       updated_at=$6 WHERE id=$7
	`
	_, err := r.db.ExecContext(ctx, query,
		user.Email, user.Role, user.Status, user.LastLogin,
		user.MFAEnabled, user.UpdatedAt, user.ID,
	)
	return err
}
