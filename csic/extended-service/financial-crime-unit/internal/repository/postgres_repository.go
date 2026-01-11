package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jitenkr2030/csic-platform/extended-services/financial-crime-unit/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

// PostgresConfig holds the PostgreSQL connection configuration
type PostgresConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Name            string
	Schema          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime string
}

// RedisConfig holds the Redis connection configuration
type RedisConfig struct {
	Host      string
	Port      int
	Password  string
	DB        int
	KeyPrefix string
	PoolSize  int
}

// PostgresRepository handles all database operations for financial crime detection
type PostgresRepository interface {
	// Alert operations
	CreateAlert(ctx context.Context, alert *domain.Alert) error
	GetAlert(ctx context.Context, id string) (*domain.Alert, error)
	GetAlerts(ctx context.Context, filter AlertFilter) ([]domain.Alert, error)
	UpdateAlert(ctx context.Context, alert *domain.Alert) error
	AcknowledgeAlert(ctx context.Context, id, acknowledgedBy string) error
	ResolveAlert(ctx context.Context, id, resolvedBy, resolution string) error
	EscalateAlert(ctx context.Context, id, escalatedTo string) error
	
	// Case operations
	CreateCase(ctx context.Context, caseObj *domain.Case) error
	GetCase(ctx context.Context, id string) (*domain.Case, error)
	GetCases(ctx context.Context, filter CaseFilter) ([]domain.Case, error)
	UpdateCase(ctx context.Context, caseObj *domain.Case) error
	TransitionCaseWorkflow(ctx context.Context, caseID, newStatus, performedBy string) error
	AssignCase(ctx context.Context, caseID, assigneeID string) error
	AddEvidenceToCase(ctx context.Context, caseID string, evidence *domain.Evidence) error
	
	// Transaction operations
	CreateTransaction(ctx context.Context, tx *domain.Transaction) error
	GetTransaction(ctx context.Context, id string) (*domain.Transaction, error)
	GetTransactionsByEntity(ctx context.Context, entityID string, limit, offset int) ([]domain.Transaction, error)
	UpdateTransactionRisk(ctx context.Context, id string, riskScore int, flagged bool, flagReason *string) error
	
	// Entity operations
	CreateEntity(ctx context.Context, entity *domain.Entity) error
	GetEntity(ctx context.Context, id string) (*domain.Entity, error)
	UpdateEntity(ctx context.Context, entity *domain.Entity) error
	GetEntityByWallet(ctx context.Context, wallet string) (*domain.Entity, error)
	
	// SAR operations
	CreateSAR(ctx context.Context, sar *domain.SAR) error
	GetSAR(ctx context.Context, id string) (*domain.SAR, error)
	GetSARs(ctx context.Context, filter SARFilter) ([]domain.SAR, error)
	UpdateSAR(ctx context.Context, sar *domain.SAR) error
	ValidateSAR(ctx context.Context, id string) error
	SubmitSAR(ctx context.Context, id, submittedBy string) error
	
	// Sanctions operations
	CreateSanctionEntry(ctx context.Context, entry *domain.SanctionList) error
	GetSanctionEntry(ctx context.Context, id string) (*domain.SanctionList, error)
	SearchSanctions(ctx context.Context, query string) ([]domain.SanctionList, error)
	GetSanctionsByWallet(ctx context.Context, wallet string) (*domain.SanctionList, error)
	
	// Watchlist operations
	AddToWatchlist(ctx context.Context, entry *domain.WatchlistEntry) error
	RemoveFromWatchlist(ctx context.Context, entityID string) error
	GetWatchlistStatus(ctx context.Context, entityID string) (*domain.WatchlistEntry, error)
	GetWatchlistEntries(ctx context.Context) ([]domain.WatchlistEntry, error)
	
	// Risk profile operations
	GetRiskProfile(ctx context.Context, entityID string) (*domain.RiskProfile, error)
	SaveRiskProfile(ctx context.Context, profile *domain.RiskProfile) error
	GetRiskHistory(ctx context.Context, entityID string, limit int) ([]domain.HistoricalScore, error)
	
	// Investigation operations
	CreateInvestigation(ctx context.Context, investigation *domain.Investigation) error
	GetInvestigation(ctx context.Context, id string) (*domain.Investigation, error)
	AddInvestigationNote(ctx context.Context, note *domain.InvestigationNote) error
	AddTimelineEvent(ctx context.Context, event *domain.TimelineEvent) error
	
	// Network analysis
	GetEntityConnections(ctx context.Context, entityID string, maxDepth int) ([]domain.Transaction, error)
	GetNetworkPatterns(ctx context.Context, entityID string) ([]domain.NetworkPattern, error)
	
	// Velocity tracking
	RecordVelocityData(ctx context.Context, data *domain.VelocityData) error
	GetVelocityData(ctx context.Context, entityID, metricType string, windowStart, windowEnd time.Time) (*domain.VelocityData, error)
	
	// Audit logging
	CreateAuditLog(ctx context.Context, log *domain.AuditLog) error
	
	// Health check
	HealthCheck(ctx context.Context) error
	Close()
}

// AlertFilter defines filtering options for alerts
type AlertFilter struct {
	EntityID    string
	EntityType  string
	Type        string
	Severity    string
	Status      string
	DateFrom    *time.Time
	DateTo      *time.Time
	AssignedTo  string
	MinScore    int
	MaxScore    int
	Limit       int
	Offset      int
}

// CaseFilter defines filtering options for cases
type CaseFilter struct {
	Status       string
	Priority     string
	SubjectID    string
	AssigneeID   string
	DateFrom     *time.Time
	DateTo       *time.Time
	RiskScoreMin int
	RiskScoreMax int
	Limit        int
	Offset       int
}

// SARFilter defines filtering options for SARs
type SARFilter struct {
	CaseID         string
	SubjectID      string
	Status         string
	RegulatoryBody string
	DateFrom       *time.Time
	DateTo         *time.Time
	Limit          int
	Offset         int
}

// pgxRepository implements PostgresRepository using pgx
type pgxRepository struct {
	pool      *pgxpool.Pool
	schema    string
	keyPrefix string
}

// NewPostgresRepository creates a new PostgreSQL repository instance
func NewPostgresRepository(config PostgresConfig) (*pgxRepository, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = int32(config.MaxOpenConns)
	poolConfig.MinConns = int32(config.MaxIdleConns)

	// Parse connection max lifetime
	if config.ConnMaxLifetime != "" {
		duration, err := time.ParseDuration(config.ConnMaxLifetime)
		if err == nil {
			poolConfig.MaxConnLifetime = duration
		}
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	repo := &pgxRepository{
		pool:   pool,
		schema: config.Schema,
	}

	// Initialize schema
	if err := repo.initSchema(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

// initSchema creates the necessary tables
func (r *pgxRepository) initSchema(ctx context.Context) error {
	queries := []string{
		// Create schema
		fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", r.schema),
		
		// Alerts table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.alerts (
				id VARCHAR(36) PRIMARY KEY,
				case_id VARCHAR(36),
				type VARCHAR(50) NOT NULL,
				severity VARCHAR(20) NOT NULL,
				status VARCHAR(20) NOT NULL DEFAULT 'new',
				score INTEGER NOT NULL,
				rule_triggered VARCHAR(100) NOT NULL,
				rule_details JSONB,
				entity_id VARCHAR(100) NOT NULL,
				entity_type VARCHAR(50) NOT NULL,
				transaction_id VARCHAR(100),
				description TEXT NOT NULL,
				context_data JSONB,
				related_alerts JSONB,
				assigned_to VARCHAR(100),
				acknowledged_at TIMESTAMP,
				resolved_at TIMESTAMP,
				resolved_by VARCHAR(100),
				resolution TEXT,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
				expires_at TIMESTAMP
			)
		`, r.schema),
		
		// Cases table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.cases (
				id VARCHAR(36) PRIMARY KEY,
				case_number VARCHAR(50) UNIQUE NOT NULL,
				status VARCHAR(20) NOT NULL DEFAULT 'new',
				priority VARCHAR(20) NOT NULL DEFAULT 'medium',
				title VARCHAR(255) NOT NULL,
				description TEXT NOT NULL,
				alert_ids JSONB,
				subject_id VARCHAR(100) NOT NULL,
				subject_type VARCHAR(50) NOT NULL,
				assignee_id VARCHAR(100),
				team_id VARCHAR(100),
				risk_score INTEGER NOT NULL DEFAULT 0,
				total_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
				currency VARCHAR(10) NOT NULL DEFAULT 'USD',
				transaction_ids JSONB,
				evidence JSONB,
				tags JSONB,
				related_cases JSONB,
				sar_id VARCHAR(36),
				regulatory_body VARCHAR(100),
				deadline TIMESTAMP,
				escalated_at TIMESTAMP,
				closed_at TIMESTAMP,
				closed_reason TEXT,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
				created_by VARCHAR(100) NOT NULL,
				version INTEGER DEFAULT 1
			)
		`, r.schema),
		
		// Transactions table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.transactions (
				id VARCHAR(36) PRIMARY KEY,
				transaction_hash VARCHAR(255) UNIQUE NOT NULL,
				type VARCHAR(50) NOT NULL,
				asset VARCHAR(50) NOT NULL,
				amount DECIMAL(20,8) NOT NULL,
				currency VARCHAR(10) NOT NULL,
				sender_id VARCHAR(100) NOT NULL,
				sender_wallet VARCHAR(255),
				receiver_id VARCHAR(100) NOT NULL,
				receiver_wallet VARCHAR(255),
				fee DECIMAL(20,8) NOT NULL DEFAULT 0,
				fee_currency VARCHAR(10),
				status VARCHAR(20) NOT NULL,
				risk_score INTEGER NOT NULL DEFAULT 0,
				flagged BOOLEAN DEFAULT FALSE,
				flag_reason TEXT,
				block_height BIGINT,
				block_hash VARCHAR(255),
				timestamp TIMESTAMP NOT NULL,
				processed_at TIMESTAMP,
				metadata JSONB,
				created_at TIMESTAMP NOT NULL DEFAULT NOW()
			)
		`, r.schema),
		
		// Entities table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.entities (
				id VARCHAR(36) PRIMARY KEY,
				external_id VARCHAR(100),
				type VARCHAR(50) NOT NULL,
				name VARCHAR(255) NOT NULL,
				email VARCHAR(255),
				phone VARCHAR(50),
				wallet_addresses JSONB,
				risk_score INTEGER NOT NULL DEFAULT 0,
				risk_level VARCHAR(20) NOT NULL DEFAULT 'low',
				kyc_check_status VARCHAR(20) NOT NULL DEFAULT 'pending',
				kyc_level INTEGER NOT NULL DEFAULT 0,
				account_status VARCHAR(20) NOT NULL DEFAULT 'active',
				watchlist_status VARCHAR(20) NOT NULL DEFAULT 'clear',
				watchlist_reason TEXT,
				pep_status VARCHAR(20) NOT NULL DEFAULT 'clear',
				sanctions_status VARCHAR(20) NOT NULL DEFAULT 'clear',
				blacklist_status VARCHAR(20) NOT NULL DEFAULT 'clear',
				high_risk_country BOOLEAN DEFAULT FALSE,
				last_risk_assessment TIMESTAMP,
				account_created_at TIMESTAMP NOT NULL,
				last_activity_at TIMESTAMP,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
				version INTEGER DEFAULT 1
			)
		`, r.schema),
		
		// SARs table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.sars (
				id VARCHAR(36) PRIMARY KEY,
				sar_number VARCHAR(50) UNIQUE NOT NULL,
				case_id VARCHAR(36) NOT NULL,
				status VARCHAR(20) NOT NULL DEFAULT 'draft',
				regulatory_body VARCHAR(100) NOT NULL,
				report_type VARCHAR(50) NOT NULL,
				subject_id VARCHAR(100) NOT NULL,
				subject_name VARCHAR(255) NOT NULL,
				subject_type VARCHAR(50) NOT NULL,
				total_amount DECIMAL(20,8) NOT NULL,
				currency VARCHAR(10) NOT NULL,
				narrative TEXT NOT NULL,
				suspicious_activity TEXT NOT NULL,
				transaction_details JSONB,
				timeline JSONB,
				supporting_docs JSONB,
				filing_institution VARCHAR(255) NOT NULL,
				filing_user_id VARCHAR(100) NOT NULL,
				reviewer_id VARCHAR(100),
				submitted_at TIMESTAMP,
				regulator_ref VARCHAR(100),
				rejection_reason TEXT,
				deadline TIMESTAMP,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
				version INTEGER DEFAULT 1
			)
		`, r.schema),
		
		// Sanctions list table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.sanction_lists (
				id VARCHAR(36) PRIMARY KEY,
				list_name VARCHAR(100) NOT NULL,
				list_type VARCHAR(50) NOT NULL,
				entity_type VARCHAR(50) NOT NULL,
				name VARCHAR(255) NOT NULL,
				alternative_names JSONB,
				date_of_birth DATE,
				nationality VARCHAR(100),
				passport_number VARCHAR(100),
				wallet_addresses JSONB,
				identifiers JSONB,
				remarks TEXT,
				program VARCHAR(255),
				effective_date DATE,
				expiration_date DATE,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP NOT NULL DEFAULT NOW()
			)
		`, r.schema),
		
		// Watchlist table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.watchlist (
				id VARCHAR(36) PRIMARY KEY,
				entity_id VARCHAR(100) NOT NULL,
				entity_type VARCHAR(50) NOT NULL,
				reason TEXT NOT NULL,
				risk_level VARCHAR(20) NOT NULL,
				added_by VARCHAR(100) NOT NULL,
				expiry_date TIMESTAMP,
				notes TEXT,
				tags JSONB,
				related_entries JSONB,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
				UNIQUE(entity_id)
			)
		`, r.schema),
		
		// Risk profiles table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.risk_profiles (
				entity_id VARCHAR(100) PRIMARY KEY,
				overall_score INTEGER NOT NULL DEFAULT 0,
				risk_level VARCHAR(20) NOT NULL DEFAULT 'low',
				components JSONB NOT NULL,
				historical_scores JSONB,
				last_assessment_at TIMESTAMP NOT NULL DEFAULT NOW(),
				next_assessment_at TIMESTAMP,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP NOT NULL DEFAULT NOW()
			)
		`, r.schema),
		
		// Investigations table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.investigations (
				id VARCHAR(36) PRIMARY KEY,
				case_id VARCHAR(36) NOT NULL,
				type VARCHAR(50) NOT NULL,
				status VARCHAR(20) NOT NULL DEFAULT 'active',
				lead_investigator VARCHAR(100) NOT NULL,
				team_members JSONB,
				subject_id VARCHAR(100) NOT NULL,
				subject_type VARCHAR(50) NOT NULL,
				scope TEXT NOT NULL,
				objectives JSONB,
				methodology TEXT,
				timeline JSONB,
				notes JSONB,
				findings JSONB,
				conclusions TEXT,
				recommendations TEXT,
				started_at TIMESTAMP NOT NULL DEFAULT NOW(),
				completed_at TIMESTAMP,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP NOT NULL DEFAULT NOW()
			)
		`, r.schema),
		
		// Velocity data table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.velocity_data (
				id VARCHAR(36) PRIMARY KEY,
				entity_id VARCHAR(100) NOT NULL,
				metric_type VARCHAR(50) NOT NULL,
				window_start TIMESTAMP NOT NULL,
				window_end TIMESTAMP NOT NULL,
				transaction_count INTEGER NOT NULL DEFAULT 0,
				total_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
				unique_counterparties INTEGER NOT NULL DEFAULT 0,
				alert_triggered BOOLEAN DEFAULT FALSE,
				created_at TIMESTAMP NOT NULL DEFAULT NOW()
			)
		`, r.schema),
		
		// Audit log table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.audit_log (
				id VARCHAR(36) PRIMARY KEY,
				timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
				action VARCHAR(100) NOT NULL,
				actor_id VARCHAR(100) NOT NULL,
				actor_type VARCHAR(50) NOT NULL,
				entity_type VARCHAR(50) NOT NULL,
				entity_id VARCHAR(100) NOT NULL,
				old_value TEXT,
				new_value TEXT,
				ip_address VARCHAR(45),
				user_agent TEXT,
				metadata JSONB
			)
		`, r.schema),
		
		// Indexes
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_alerts_entity_id ON %s.alerts(entity_id)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_alerts_status ON %s.alerts(status)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_alerts_severity ON %s.alerts(severity)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_alerts_created_at ON %s.alerts(created_at)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_cases_status ON %s.cases(status)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_cases_subject_id ON %s.cases(subject_id)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_cases_assignee ON %s.cases(assignee_id)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_transactions_sender ON %s.transactions(sender_id)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_transactions_receiver ON %s.transactions(receiver_id)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_transactions_hash ON %s.transactions(transaction_hash)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_entities_wallets ON %s.entities using GIN(wallet_addresses)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_sanctions_name ON %s.sanction_lists using GIN(name gin_trgm_ops)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_sanctions_wallets ON %s.sanction_lists using GIN(wallet_addresses)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_watchlist_entity ON %s.watchlist(entity_id)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_audit_timestamp ON %s.audit_log(timestamp)`, r.schema, r.schema),
	}

	for _, query := range queries {
		if _, err := r.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	return nil
}

// Close closes the database connection pool
func (r *pgxRepository) Close() {
	r.pool.Close()
}

// CreateAlert creates a new alert
func (r *pgxRepository) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()
	alert.Status = domain.AlertStatusNew

	ruleDetailsJSON, err := json.Marshal(alert.RuleDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal rule details: %w", err)
	}

	contextDataJSON, err := json.Marshal(alert.ContextData)
	if err != nil {
		return fmt.Errorf("failed to marshal context data: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.alerts (
			id, case_id, type, severity, status, score, rule_triggered,
			rule_details, entity_id, entity_type, transaction_id,
			description, context_data, related_alerts, assigned_to,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`, r.schema)

	_, err = r.pool.Exec(ctx, query,
		alert.ID,
		alert.CaseID,
		alert.Type,
		alert.Severity,
		alert.Status,
		alert.Score,
		alert.RuleTriggered,
		ruleDetailsJSON,
		alert.EntityID,
		alert.EntityType,
		alert.TransactionID,
		alert.Description,
		contextDataJSON,
		alert.RelatedAlerts,
		alert.AssignedTo,
		alert.CreatedAt,
		alert.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

// GetAlert retrieves an alert by ID
func (r *pgxRepository) GetAlert(ctx context.Context, id string) (*domain.Alert, error) {
	query := fmt.Sprintf(`
		SELECT id, case_id, type, severity, status, score, rule_triggered,
			rule_details, entity_id, entity_type, transaction_id,
			description, context_data, related_alerts, assigned_to,
			acknowledged_at, resolved_at, resolved_by, resolution,
			created_at, updated_at, expires_at
		FROM %s.alerts
		WHERE id = $1
	`, r.schema)

	var alert domain.Alert
	var ruleDetailsJSON, contextDataJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&alert.ID,
		&alert.CaseID,
		&alert.Type,
		&alert.Severity,
		&alert.Status,
		&alert.Score,
		&alert.RuleTriggered,
		&ruleDetailsJSON,
		&alert.EntityID,
		&alert.EntityType,
		&alert.TransactionID,
		&alert.Description,
		&contextDataJSON,
		&alert.RelatedAlerts,
		&alert.AssignedTo,
		&alert.AcknowledgedAt,
		&alert.ResolvedAt,
		&alert.ResolvedBy,
		&alert.Resolution,
		&alert.CreatedAt,
		&alert.UpdatedAt,
		&alert.ExpiresAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	if ruleDetailsJSON != nil {
		if err := json.Unmarshal(ruleDetailsJSON, &alert.RuleDetails); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule details: %w", err)
		}
	}
	if contextDataJSON != nil {
		if err := json.Unmarshal(contextDataJSON, &alert.ContextData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal context data: %w", err)
		}
	}

	return &alert, nil
}

// GetAlerts retrieves alerts based on filter criteria
func (r *pgxRepository) GetAlerts(ctx context.Context, filter AlertFilter) ([]domain.Alert, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.EntityID != "" {
		conditions = append(conditions, fmt.Sprintf("entity_id = $%d", argIndex))
		args = append(args, filter.EntityID)
		argIndex++
	}
	if filter.EntityType != "" {
		conditions = append(conditions, fmt.Sprintf("entity_type = $%d", argIndex))
		args = append(args, filter.EntityType)
		argIndex++
	}
	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, filter.Type)
		argIndex++
	}
	if filter.Severity != "" {
		conditions = append(conditions, fmt.Sprintf("severity = $%d", argIndex))
		args = append(args, filter.Severity)
		argIndex++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}
	if filter.AssignedTo != "" {
		conditions = append(conditions, fmt.Sprintf("assigned_to = $%d", argIndex))
		args = append(args, filter.AssignedTo)
		argIndex++
	}
	if filter.MinScore > 0 {
		conditions = append(conditions, fmt.Sprintf("score >= $%d", argIndex))
		args = append(args, filter.MinScore)
		argIndex++
	}
	if filter.MaxScore > 0 {
		conditions = append(conditions, fmt.Sprintf("score <= $%d", argIndex))
		args = append(args, filter.MaxScore)
		argIndex++
	}

	query := fmt.Sprintf(`
		SELECT id, case_id, type, severity, status, score, rule_triggered,
			rule_details, entity_id, entity_type, transaction_id,
			description, context_data, related_alerts, assigned_to,
			acknowledged_at, resolved_at, resolved_by, resolution,
			created_at, updated_at, expires_at
		FROM %s.alerts
	`, r.schema)

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	defer rows.Close()

	var alerts []domain.Alert
	for rows.Next() {
		var alert domain.Alert
		var ruleDetailsJSON, contextDataJSON []byte

		if err := rows.Scan(
			&alert.ID,
			&alert.CaseID,
			&alert.Type,
			&alert.Severity,
			&alert.Status,
			&alert.Score,
			&alert.RuleTriggered,
			&ruleDetailsJSON,
			&alert.EntityID,
			&alert.EntityType,
			&alert.TransactionID,
			&alert.Description,
			&contextDataJSON,
			&alert.RelatedAlerts,
			&alert.AssignedTo,
			&alert.AcknowledgedAt,
			&alert.ResolvedAt,
			&alert.ResolvedBy,
			&alert.Resolution,
			&alert.CreatedAt,
			&alert.UpdatedAt,
			&alert.ExpiresAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}

		if ruleDetailsJSON != nil {
			json.Unmarshal(ruleDetailsJSON, &alert.RuleDetails)
		}
		if contextDataJSON != nil {
			json.Unmarshal(contextDataJSON, &alert.ContextData)
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// CreateCase creates a new case
func (r *pgxRepository) CreateCase(ctx context.Context, caseObj *domain.Case) error {
	if caseObj.ID == "" {
		caseObj.ID = uuid.New().String()
	}
	if caseObj.CaseNumber == "" {
		caseObj.CaseNumber = fmt.Sprintf("FCU-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:8])
	}
	caseObj.CreatedAt = time.Now()
	caseObj.UpdatedAt = time.Now()
	caseObj.Status = domain.CaseStatusNew
	caseObj.Version = 1

	alertIDsJSON, err := json.Marshal(caseObj.AlertIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal alert IDs: %w", err)
	}

	txIDsJSON, err := json.Marshal(caseObj.TransactionIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction IDs: %w", err)
	}

	evidenceJSON, err := json.Marshal(caseObj.Evidence)
	if err != nil {
		return fmt.Errorf("failed to marshal evidence: %w", err)
	}

	tagsJSON, err := json.Marshal(caseObj.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.cases (
			id, case_number, status, priority, title, description,
			alert_ids, subject_id, subject_type, assignee_id, team_id,
			risk_score, total_amount, currency, transaction_ids,
			evidence, tags, related_cases, created_at, updated_at,
			created_by, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
	`, r.schema)

	_, err = r.pool.Exec(ctx, query,
		caseObj.ID,
		caseObj.CaseNumber,
		caseObj.Status,
		caseObj.Priority,
		caseObj.Title,
		caseObj.Description,
		alertIDsJSON,
		caseObj.SubjectID,
		caseObj.SubjectType,
		caseObj.AssigneeID,
		caseObj.TeamID,
		caseObj.RiskScore,
		caseObj.TotalAmount,
		caseObj.Currency,
		txIDsJSON,
		evidenceJSON,
		tagsJSON,
		caseObj.RelatedCases,
		caseObj.CreatedAt,
		caseObj.UpdatedAt,
		caseObj.CreatedBy,
		caseObj.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to create case: %w", err)
	}

	return nil
}

// GetCase retrieves a case by ID
func (r *pgxRepository) GetCase(ctx context.Context, id string) (*domain.Case, error) {
	query := fmt.Sprintf(`
		SELECT id, case_number, status, priority, title, description,
			alert_ids, subject_id, subject_type, assignee_id, team_id,
			risk_score, total_amount, currency, transaction_ids,
			evidence, tags, related_cases, sar_id, regulatory_body,
			deadline, escalated_at, closed_at, closed_reason,
			created_at, updated_at, created_by, version
		FROM %s.cases
		WHERE id = $1
	`, r.schema)

	var caseObj domain.Case
	var alertIDsJSON, txIDsJSON, evidenceJSON, tagsJSON, relatedCasesJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&caseObj.ID,
		&caseObj.CaseNumber,
		&caseObj.Status,
		&caseObj.Priority,
		&caseObj.Title,
		&caseObj.Description,
		&alertIDsJSON,
		&caseObj.SubjectID,
		&caseObj.SubjectType,
		&caseObj.AssigneeID,
		&caseObj.TeamID,
		&caseObj.RiskScore,
		&caseObj.TotalAmount,
		&caseObj.Currency,
		&txIDsJSON,
		&evidenceJSON,
		&tagsJSON,
		&relatedCasesJSON,
		&caseObj.SARID,
		&caseObj.RegulatoryBody,
		&caseObj.Deadline,
		&caseObj.EscalatedAt,
		&caseObj.ClosedAt,
		&caseObj.ClosedReason,
		&caseObj.CreatedAt,
		&caseObj.UpdatedAt,
		&caseObj.CreatedBy,
		&caseObj.Version,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get case: %w", err)
	}

	if alertIDsJSON != nil {
		json.Unmarshal(alertIDsJSON, &caseObj.AlertIDs)
	}
	if txIDsJSON != nil {
		json.Unmarshal(txIDsJSON, &caseObj.TransactionIDs)
	}
	if evidenceJSON != nil {
		json.Unmarshal(evidenceJSON, &caseObj.Evidence)
	}
	if tagsJSON != nil {
		json.Unmarshal(tagsJSON, &caseObj.Tags)
	}
	if relatedCasesJSON != nil {
		json.Unmarshal(relatedCasesJSON, &caseObj.RelatedCases)
	}

	return &caseObj, nil
}

// CreateSAR creates a new SAR
func (r *pgxRepository) CreateSAR(ctx context.Context, sar *domain.SAR) error {
	if sar.ID == "" {
		sar.ID = uuid.New().String()
	}
	if sar.SARNumber == "" {
		sar.SARNumber = fmt.Sprintf("SAR-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:8])
	}
	sar.CreatedAt = time.Now()
	sar.UpdatedAt = time.Now()
	sar.Status = domain.SARStatusDraft

	txDetailsJSON, err := json.Marshal(sar.TransactionDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction details: %w", err)
	}

	timelineJSON, err := json.Marshal(sar.Timeline)
	if err != nil {
		return fmt.Errorf("failed to marshal timeline: %w", err)
	}

	docsJSON, err := json.Marshal(sar.SupportingDocs)
	if err != nil {
		return fmt.Errorf("failed to marshal supporting docs: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.sars (
			id, sar_number, case_id, status, regulatory_body, report_type,
			subject_id, subject_name, subject_type, total_amount, currency,
			narrative, suspicious_activity, transaction_details, timeline,
			supporting_docs, filing_institution, filing_user_id,
			created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`, r.schema)

	_, err = r.pool.Exec(ctx, query,
		sar.ID,
		sar.SARNumber,
		sar.CaseID,
		sar.Status,
		sar.RegulatoryBody,
		sar.ReportType,
		sar.SubjectID,
		sar.SubjectName,
		sar.SubjectType,
		sar.TotalAmount,
		sar.Currency,
		sar.Narrative,
		sar.SuspiciousActivity,
		txDetailsJSON,
		timelineJSON,
		docsJSON,
		sar.FilingInstitution,
		sar.FilingUserID,
		sar.CreatedAt,
		sar.UpdatedAt,
		sar.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to create SAR: %w", err)
	}

	return nil
}

// GetSAR retrieves a SAR by ID
func (r *pgxRepository) GetSAR(ctx context.Context, id string) (*domain.SAR, error) {
	query := fmt.Sprintf(`
		SELECT id, sar_number, case_id, status, regulatory_body, report_type,
			subject_id, subject_name, subject_type, total_amount, currency,
			narrative, suspicious_activity, transaction_details, timeline,
			supporting_docs, filing_institution, filing_user_id, reviewer_id,
			submitted_at, regulator_ref, rejection_reason, deadline,
			created_at, updated_at, version
		FROM %s.sars
		WHERE id = $1
	`, r.schema)

	var sar domain.SAR
	var txDetailsJSON, timelineJSON, docsJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&sar.ID,
		&sar.SARNumber,
		&sar.CaseID,
		&sar.Status,
		&sar.RegulatoryBody,
		&sar.ReportType,
		&sar.SubjectID,
		&sar.SubjectName,
		&sar.SubjectType,
		&sar.TotalAmount,
		&sar.Currency,
		&sar.Narrative,
		&sar.SuspiciousActivity,
		&txDetailsJSON,
		&timelineJSON,
		&docsJSON,
		&sar.FilingInstitution,
		&sar.FilingUserID,
		&sar.ReviewerID,
		&sar.SubmittedAt,
		&sar.RegulatorRef,
		&sar.RejectionReason,
		&sar.Deadline,
		&sar.CreatedAt,
		&sar.UpdatedAt,
		&sar.Version,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SAR: %w", err)
	}

	if txDetailsJSON != nil {
		json.Unmarshal(txDetailsJSON, &sar.TransactionDetails)
	}
	if timelineJSON != nil {
		json.Unmarshal(timelineJSON, &sar.Timeline)
	}
	if docsJSON != nil {
		json.Unmarshal(docsJSON, &sar.SupportingDocs)
	}

	return &sar, nil
}

// CreateEntity creates a new entity
func (r *pgxRepository) CreateEntity(ctx context.Context, entity *domain.Entity) error {
	if entity.ID == "" {
		entity.ID = uuid.New().String()
	}
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
	entity.Version = 1

	walletsJSON, err := json.Marshal(entity.WalletAddresses)
	if err != nil {
		return fmt.Errorf("failed to marshal wallet addresses: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.entities (
			id, external_id, type, name, email, phone, wallet_addresses,
			risk_score, risk_level, kyc_check_status, kyc_level,
			account_status, watchlist_status, watchlist_reason,
			pep_status, sanctions_status, blacklist_status, high_risk_country,
			account_created_at, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
	`, r.schema)

	_, err = r.pool.Exec(ctx, query,
		entity.ID,
		entity.ExternalID,
		entity.Type,
		entity.Name,
		entity.Email,
		entity.Phone,
		walletsJSON,
		entity.RiskScore,
		entity.RiskLevel,
		entity.KYCCheckStatus,
		entity.KYCLevel,
		entity.AccountStatus,
		entity.WatchlistStatus,
		entity.WatchlistReason,
		entity.PEPStatus,
		entity.SanctionsStatus,
		entity.BlacklistStatus,
		entity.HighRiskCountry,
		entity.AccountCreatedAt,
		entity.CreatedAt,
		entity.UpdatedAt,
		entity.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}

	return nil
}

// GetEntity retrieves an entity by ID
func (r *pgxRepository) GetEntity(ctx context.Context, id string) (*domain.Entity, error) {
	query := fmt.Sprintf(`
		SELECT id, external_id, type, name, email, phone, wallet_addresses,
			risk_score, risk_level, kyc_check_status, kyc_level,
			account_status, watchlist_status, watchlist_reason,
			pep_status, sanctions_status, blacklist_status, high_risk_country,
			last_risk_assessment, account_created_at, last_activity_at,
			created_at, updated_at, version
		FROM %s.entities
		WHERE id = $1
	`, r.schema)

	var entity domain.Entity
	var walletsJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&entity.ID,
		&entity.ExternalID,
		&entity.Type,
		&entity.Name,
		&entity.Email,
		&entity.Phone,
		&walletsJSON,
		&entity.RiskScore,
		&entity.RiskLevel,
		&entity.KYCCheckStatus,
		&entity.KYCLevel,
		&entity.AccountStatus,
		&entity.WatchlistStatus,
		&entity.WatchlistReason,
		&entity.PEPStatus,
		&entity.SanctionsStatus,
		&entity.BlacklistStatus,
		&entity.HighRiskCountry,
		&entity.LastRiskAssessment,
		&entity.AccountCreatedAt,
		&entity.LastActivityAt,
		&entity.CreatedAt,
		&entity.UpdatedAt,
		&entity.Version,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	if walletsJSON != nil {
		json.Unmarshal(walletsJSON, &entity.WalletAddresses)
	}

	return &entity, nil
}

// CreateTransaction creates a new transaction
func (r *pgxRepository) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	if tx.ID == "" {
		tx.ID = uuid.New().String()
	}
	tx.CreatedAt = time.Now()

	metadataJSON, err := json.Marshal(tx.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.transactions (
			id, transaction_hash, type, asset, amount, currency,
			sender_id, sender_wallet, receiver_id, receiver_wallet,
			fee, fee_currency, status, risk_score, flagged, flag_reason,
			block_height, block_hash, timestamp, processed_at, metadata,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
	`, r.schema)

	_, err = r.pool.Exec(ctx, query,
		tx.ID,
		tx.TransactionHash,
		tx.Type,
		tx.Asset,
		tx.Amount,
		tx.Currency,
		tx.SenderID,
		tx.SenderWallet,
		tx.ReceiverID,
		tx.ReceiverWallet,
		tx.Fee,
		tx.FeeCurrency,
		tx.Status,
		tx.RiskScore,
		tx.Flagged,
		tx.FlagReason,
		tx.BlockHeight,
		tx.BlockHash,
		tx.Timestamp,
		tx.ProcessedAt,
		metadataJSON,
		tx.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetTransaction retrieves a transaction by ID
func (r *pgxRepository) GetTransaction(ctx context.Context, id string) (*domain.Transaction, error) {
	query := fmt.Sprintf(`
		SELECT id, transaction_hash, type, asset, amount, currency,
			sender_id, sender_wallet, receiver_id, receiver_wallet,
			fee, fee_currency, status, risk_score, flagged, flag_reason,
			block_height, block_hash, timestamp, processed_at, metadata,
			created_at
		FROM %s.transactions
		WHERE id = $1
	`, r.schema)

	var tx domain.Transaction
	var metadataJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tx.ID,
		&tx.TransactionHash,
		&tx.Type,
		&tx.Asset,
		&tx.Amount,
		&tx.Currency,
		&tx.SenderID,
		&tx.SenderWallet,
		&tx.ReceiverID,
		&tx.ReceiverWallet,
		&tx.Fee,
		&tx.FeeCurrency,
		&tx.Status,
		&tx.RiskScore,
		&tx.Flagged,
		&tx.FlagReason,
		&tx.BlockHeight,
		&tx.BlockHash,
		&tx.Timestamp,
		&tx.ProcessedAt,
		&metadataJSON,
		&tx.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &tx.Metadata)
	}

	return &tx, nil
}

// GetTransactionsByEntity retrieves transactions for an entity
func (r *pgxRepository) GetTransactionsByEntity(ctx context.Context, entityID string, limit, offset int) ([]domain.Transaction, error) {
	query := fmt.Sprintf(`
		SELECT id, transaction_hash, type, asset, amount, currency,
			sender_id, sender_wallet, receiver_id, receiver_wallet,
			fee, fee_currency, status, risk_score, flagged, flag_reason,
			block_height, block_hash, timestamp, processed_at, metadata,
			created_at
		FROM %s.transactions
		WHERE sender_id = $1 OR receiver_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`, r.schema)

	rows, err := r.pool.Query(ctx, query, entityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		var metadataJSON []byte

		if err := rows.Scan(
			&tx.ID,
			&tx.TransactionHash,
			&tx.Type,
			&tx.Asset,
			&tx.Amount,
			&tx.Currency,
			&tx.SenderID,
			&tx.SenderWallet,
			&tx.ReceiverID,
			&tx.ReceiverWallet,
			&tx.Fee,
			&tx.FeeCurrency,
			&tx.Status,
			&tx.RiskScore,
			&tx.Flagged,
			&tx.FlagReason,
			&tx.BlockHeight,
			&tx.BlockHash,
			&tx.Timestamp,
			&tx.ProcessedAt,
			&metadataJSON,
			&tx.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &tx.Metadata)
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// AddToWatchlist adds an entity to the watchlist
func (r *pgxRepository) AddToWatchlist(ctx context.Context, entry *domain.WatchlistEntry) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	tagsJSON, err := json.Marshal(entry.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	relatedJSON, err := json.Marshal(entry.RelatedEntries)
	if err != nil {
		return fmt.Errorf("failed to marshal related entries: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.watchlist (
			id, entity_id, entity_type, reason, risk_level,
			added_by, expiry_date, notes, tags, related_entries,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (entity_id)
		DO UPDATE SET
			reason = EXCLUDED.reason,
			risk_level = EXCLUDED.risk_level,
			updated_at = EXCLUDED.updated_at
	`, r.schema)

	_, err = r.pool.Exec(ctx, query,
		entry.ID,
		entry.EntityID,
		entry.EntityType,
		entry.Reason,
		entry.RiskLevel,
		entry.AddedBy,
		entry.ExpiryDate,
		entry.Notes,
		tagsJSON,
		relatedJSON,
		entry.CreatedAt,
		entry.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add to watchlist: %w", err)
	}

	return nil
}

// GetWatchlistStatus retrieves watchlist status for an entity
func (r *pgxRepository) GetWatchlistStatus(ctx context.Context, entityID string) (*domain.WatchlistEntry, error) {
	query := fmt.Sprintf(`
		SELECT id, entity_id, entity_type, reason, risk_level,
			added_by, expiry_date, notes, tags, related_entries,
			created_at, updated_at
		FROM %s.watchlist
		WHERE entity_id = $1
	`, r.schema)

	var entry domain.WatchlistEntry
	var tagsJSON, relatedJSON []byte

	err := r.pool.QueryRow(ctx, query, entityID).Scan(
		&entry.ID,
		&entry.EntityID,
		&entry.EntityType,
		&entry.Reason,
		&entry.RiskLevel,
		&entry.AddedBy,
		&entry.ExpiryDate,
		&entry.Notes,
		&tagsJSON,
		&relatedJSON,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get watchlist status: %w", err)
	}

	if tagsJSON != nil {
		json.Unmarshal(tagsJSON, &entry.Tags)
	}
	if relatedJSON != nil {
		json.Unmarshal(relatedJSON, &entry.RelatedEntries)
	}

	return &entry, nil
}

// GetRiskProfile retrieves the risk profile for an entity
func (r *pgxRepository) GetRiskProfile(ctx context.Context, entityID string) (*domain.RiskProfile, error) {
	query := fmt.Sprintf(`
		SELECT entity_id, overall_score, risk_level, components,
			historical_scores, last_assessment_at, next_assessment_at,
			created_at, updated_at
		FROM %s.risk_profiles
		WHERE entity_id = $1
	`, r.schema)

	var profile domain.RiskProfile
	var componentsJSON, historicalScoresJSON []byte

	err := r.pool.QueryRow(ctx, query, entityID).Scan(
		&profile.EntityID,
		&profile.OverallScore,
		&profile.RiskLevel,
		&componentsJSON,
		&historicalScoresJSON,
		&profile.LastAssessmentAt,
		&profile.NextAssessmentAt,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get risk profile: %w", err)
	}

	if componentsJSON != nil {
		json.Unmarshal(componentsJSON, &profile.Components)
	}
	if historicalScoresJSON != nil {
		json.Unmarshal(historicalScoresJSON, &profile.HistoricalScores)
	}

	return &profile, nil
}

// SaveRiskProfile saves or updates a risk profile
func (r *pgxRepository) SaveRiskProfile(ctx context.Context, profile *domain.RiskProfile) error {
	profile.UpdatedAt = time.Now()

	componentsJSON, err := json.Marshal(profile.Components)
	if err != nil {
		return fmt.Errorf("failed to marshal components: %w", err)
	}

	historicalJSON, err := json.Marshal(profile.HistoricalScores)
	if err != nil {
		return fmt.Errorf("failed to marshal historical scores: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.risk_profiles (
			entity_id, overall_score, risk_level, components,
			historical_scores, last_assessment_at, next_assessment_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (entity_id)
		DO UPDATE SET
			overall_score = EXCLUDED.overall_score,
			risk_level = EXCLUDED.risk_level,
			components = EXCLUDED.components,
			historical_scores = EXCLUDED.historical_scores,
			last_assessment_at = EXCLUDED.last_assessment_at,
			next_assessment_at = EXCLUDED.next_assessment_at,
			updated_at = EXCLUDED.updated_at
	`, r.schema)

	_, err = r.pool.Exec(ctx, query,
		profile.EntityID,
		profile.OverallScore,
		profile.RiskLevel,
		componentsJSON,
		historicalJSON,
		profile.LastAssessmentAt,
		profile.NextAssessmentAt,
		profile.CreatedAt,
		profile.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save risk profile: %w", err)
	}

	return nil
}

// CreateAuditLog creates an audit log entry
func (r *pgxRepository) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	metadataJSON, err := json.Marshal(log.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.audit_log (
			id, timestamp, action, actor_id, actor_type,
			entity_type, entity_id, old_value, new_value,
			ip_address, user_agent, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, r.schema)

	_, err = r.pool.Exec(ctx, query,
		log.ID,
		log.Timestamp,
		log.Action,
		log.ActorID,
		log.ActorType,
		log.EntityType,
		log.EntityID,
		log.OldValue,
		log.NewValue,
		log.IPAddress,
		log.UserAgent,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// HealthCheck verifies the database connection is healthy
func (r *pgxRepository) HealthCheck(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

// RedisCache implements caching using Redis
type RedisCache struct {
	client    *redis.Client
	keyPrefix string
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(config RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client:    client,
		keyPrefix: config.KeyPrefix,
	}, nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Get retrieves a value from cache
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, r.keyPrefix+key).Result()
}

// Set sets a value in cache with expiration
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(ctx, r.keyPrefix+key, data, expiration).Err()
}

// Delete deletes a key from cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, r.keyPrefix+key).Err()
}

// Increment increments a counter
func (r *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, r.keyPrefix+key).Result()
}

// Expire sets expiration on a key
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, r.keyPrefix+key, expiration).Err()
}

// Log provides a simple logging function for debugging
func Log(format string, args ...interface{}) {
	log.Printf("[FCURepo] "+format, args...)
}
