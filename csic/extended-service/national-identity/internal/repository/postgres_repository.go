package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jitenkr2030/csic-platform/extended-services/national-identity/internal/domain"

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

// PostgresRepository handles all database operations for identity management
type PostgresRepository interface {
	// Identity operations
	CreateIdentity(ctx context.Context, identity *domain.Identity) error
	GetIdentity(ctx context.Context, id string) (*domain.Identity, error)
	GetIdentityByNationalID(ctx context.Context, nationalID string) (*domain.Identity, error)
	UpdateIdentity(ctx context.Context, identity *domain.Identity) error
	DeleteIdentity(ctx context.Context, id string) error
	GetIdentityHistory(ctx context.Context, id string, limit, offset int) ([]domain.IdentityHistory, error)
	
	// Verification operations
	CreateVerification(ctx context.Context, verification *domain.IdentityVerification) error
	GetVerifications(ctx context.Context, identityID string) ([]domain.IdentityVerification, error)
	UpdateVerification(ctx context.Context, verification *domain.IdentityVerification) error
	
	// Biometric operations
	SaveBiometricTemplate(ctx context.Context, template *domain.BiometricTemplate) error
	GetBiometricTemplate(ctx context.Context, identityID, biometricType string) (*domain.BiometricTemplate, error)
	DeleteBiometricTemplate(ctx context.Context, id string) error
	
	// Behavioral profile operations
	SaveBehavioralProfile(ctx context.Context, profile *domain.BehavioralProfile) error
	GetBehavioralProfile(ctx context.Context, identityID string) (*domain.BehavioralProfile, error)
	SaveBehavioralDataPoint(ctx context.Context, dataPoint *domain.BehavioralDataPoint) error
	
	// Fraud operations
	CreateFraudAlert(ctx context.Context, alert *domain.FraudAlert) error
	GetFraudAlerts(ctx context.Context, identityID string, status string) ([]domain.FraudAlert, error)
	ResolveFraudAlert(ctx context.Context, alertID, resolvedBy, resolution string) error
	GetFraudScore(ctx context.Context, identityID string) (int, error)
	
	// Authentication operations
	CreateSession(ctx context.Context, session *domain.AuthenticationSession) error
	GetSession(ctx context.Context, sessionToken string) (*domain.AuthenticationSession, error)
	UpdateSession(ctx context.Context, session *domain.AuthenticationSession) error
	TerminateSession(ctx context.Context, sessionToken, reason string) error
	TerminateAllSessions(ctx context.Context, identityID, reason string) error
	GetActiveSessions(ctx context.Context, identityID string) ([]domain.AuthenticationSession, error)
	
	// MFA operations
	SaveMFAToken(ctx context.Context, token *domain.MFAToken) error
	GetMFAToken(ctx context.Context, identityID, tokenType string) (*domain.MFAToken, error)
	VerifyMFAToken(ctx context.Context, tokenID string) error
	InvalidateMFAToken(ctx context.Context, tokenID string) error
	UpdateMFATokenAttempts(ctx context.Context, tokenID string, failed bool) error
	
	// Federated learning operations
	SaveModelUpdate(ctx context.Context, update *domain.FederatedModelUpdate) error
	GetModelUpdates(ctx context.Context, modelVersion string) ([]domain.FederatedModelUpdate, error)
	GetModelStatus(ctx context.Context) (*domain.ModelStatus, error)
	GetClientContributions(ctx context.Context, clientID string) ([]domain.ClientContribution, error)
	
	// Consent operations
	SaveConsentRecord(ctx context.Context, record *domain.ConsentRecord) error
	GetConsentRecord(ctx context.Context, identityID, consentType string) (*domain.ConsentRecord, error)
	GetAllConsentRecords(ctx context.Context, identityID string) ([]domain.ConsentRecord, error)
	
	// Event operations
	PublishEvent(ctx context.Context, event *domain.IdentityEvent) error
	GetEvents(ctx context.Context, identityID string, limit int) ([]domain.IdentityEvent, error)
	
	// Health and metrics
	HealthCheck(ctx context.Context) error
	Close()
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

// initSchema creates the necessary tables if they don't exist
func (r *pgxRepository) initSchema(ctx context.Context) error {
	queries := []string{
		// Create schema if not exists
		fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", r.schema),
		
		// Identities table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.identities (
				id VARCHAR(36) PRIMARY KEY,
				national_id VARCHAR(50) UNIQUE NOT NULL,
				legacy_id VARCHAR(50),
				first_name VARCHAR(100) NOT NULL,
				middle_name VARCHAR(100),
				last_name VARCHAR(100) NOT NULL,
				date_of_birth DATE NOT NULL,
				place_of_birth VARCHAR(200) NOT NULL,
				gender VARCHAR(20) NOT NULL,
				nationality VARCHAR(100) NOT NULL,
				ethnicity VARCHAR(100),
				marital_status VARCHAR(50),
				current_address JSONB NOT NULL,
				permanent_address JSONB,
				email VARCHAR(255),
				phone VARCHAR(50),
				mobile VARCHAR(50),
				passport_number VARCHAR(50),
				passport_expiry DATE,
				driver_license VARCHAR(50),
				biometric_id VARCHAR(36),
				photo_url TEXT,
				fingerprint_hash VARCHAR(256),
				iris_scan_hash VARCHAR(256),
				facial_template_hash VARCHAR(256),
				status VARCHAR(20) NOT NULL DEFAULT 'pending',
				verification_level VARCHAR(20) NOT NULL DEFAULT 'none',
				verified_at TIMESTAMP,
				verified_by VARCHAR(100),
				fraud_score INTEGER DEFAULT 0,
				risk_level VARCHAR(20) DEFAULT 'low',
				last_risk_assessment TIMESTAMP,
				consent_given BOOLEAN DEFAULT FALSE,
				consent_date TIMESTAMP,
				data_opt_out BOOLEAN DEFAULT FALSE,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
				created_by VARCHAR(100) NOT NULL,
				updated_by VARCHAR(100),
				version INTEGER DEFAULT 1,
				deleted_at TIMESTAMP,
				is_deleted BOOLEAN DEFAULT FALSE
			)
		`, r.schema),

		// Identity history table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.identity_history (
				id VARCHAR(36) PRIMARY KEY,
				identity_id VARCHAR(36) NOT NULL,
				action VARCHAR(50) NOT NULL,
				field_changed VARCHAR(100),
				old_value TEXT,
				new_value TEXT,
				changed_by VARCHAR(100) NOT NULL,
				changed_at TIMESTAMP NOT NULL DEFAULT NOW(),
				reason TEXT,
				ip_address VARCHAR(45),
				user_agent TEXT
			)
		`, r.schema),

		// Identity verifications table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.identity_verifications (
				id VARCHAR(36) PRIMARY KEY,
				identity_id VARCHAR(36) NOT NULL,
				verification_type VARCHAR(50) NOT NULL,
				method VARCHAR(50) NOT NULL,
				status VARCHAR(20) NOT NULL,
				score DECIMAL(5,4) NOT NULL,
				threshold DECIMAL(5,4) NOT NULL,
				passed BOOLEAN NOT NULL,
				failure_reason TEXT,
				evidence JSONB,
				verified_at TIMESTAMP NOT NULL DEFAULT NOW(),
				expires_at TIMESTAMP,
				verified_by VARCHAR(100),
				created_at TIMESTAMP NOT NULL DEFAULT NOW()
			)
		`, r.schema),

		// Biometric templates table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.biometric_templates (
				id VARCHAR(36) PRIMARY KEY,
				identity_id VARCHAR(36) NOT NULL,
				biometric_type VARCHAR(50) NOT NULL,
				template_hash VARCHAR(256) NOT NULL,
				template_data BYTEA,
				encryption_key_id VARCHAR(36),
				quality_score DECIMAL(5,4) NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
				expires_at TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE,
				UNIQUE(identity_id, biometric_type)
			)
		`, r.schema),

		// Behavioral profiles table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.behavioral_profiles (
				id VARCHAR(36) PRIMARY KEY,
				identity_id VARCHAR(36) UNIQUE NOT NULL,
				keystroke_profile JSONB,
				mouse_profile JSONB,
				touch_profile JSONB,
				gait_profile JSONB,
				overall_score DECIMAL(5,4) NOT NULL,
				anomaly_count INTEGER DEFAULT 0,
				last_anomaly_at TIMESTAMP,
				baseline_established_at TIMESTAMP NOT NULL,
				updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
				sample_count INTEGER DEFAULT 0
			)
		`, r.schema),

		// Behavioral data points table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.behavioral_data_points (
				id VARCHAR(36) PRIMARY KEY,
				identity_id VARCHAR(36) NOT NULL,
				data_type VARCHAR(50) NOT NULL,
				raw_data BYTEA,
				processed_data JSONB,
				anomaly_score DECIMAL(5,4) NOT NULL,
				is_anomaly BOOLEAN NOT NULL,
				device_info JSONB,
				session_id VARCHAR(36),
				submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
				processed_at TIMESTAMP
			)
		`, r.schema),

		// Fraud alerts table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.fraud_alerts (
				id VARCHAR(36) PRIMARY KEY,
				identity_id VARCHAR(36) NOT NULL,
				alert_type VARCHAR(50) NOT NULL,
				severity VARCHAR(20) NOT NULL,
				score INTEGER NOT NULL,
				indicator VARCHAR(100) NOT NULL,
				description TEXT NOT NULL,
				evidence TEXT,
				related_event_id VARCHAR(36),
				status VARCHAR(20) NOT NULL DEFAULT 'open',
				resolved_at TIMESTAMP,
				resolved_by VARCHAR(100),
				resolution TEXT,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP NOT NULL DEFAULT NOW()
			)
		`, r.schema),

		// Authentication sessions table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.authentication_sessions (
				id VARCHAR(36) PRIMARY KEY,
				identity_id VARCHAR(36) NOT NULL,
				session_token VARCHAR(255) UNIQUE NOT NULL,
				refresh_token VARCHAR(255) UNIQUE NOT NULL,
				ip_address VARCHAR(45) NOT NULL,
				user_agent TEXT NOT NULL,
				device_info JSONB,
				mfa_enabled BOOLEAN DEFAULT FALSE,
				mfa_passed BOOLEAN DEFAULT FALSE,
				behavioral_score DECIMAL(5,4) NOT NULL,
				risk_score INTEGER NOT NULL,
				expires_at TIMESTAMP NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				last_activity_at TIMESTAMP NOT NULL DEFAULT NOW(),
				terminated_at TIMESTAMP,
				termination_reason TEXT
			)
		`, r.schema),

		// MFA tokens table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.mfa_tokens (
				id VARCHAR(36) PRIMARY KEY,
				identity_id VARCHAR(36) NOT NULL,
				token_type VARCHAR(50) NOT NULL,
				token_hash VARCHAR(255) NOT NULL,
				secret TEXT,
				phone_number VARCHAR(50),
				email_address VARCHAR(255),
				is_active BOOLEAN DEFAULT TRUE,
				verified_at TIMESTAMP,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				expires_at TIMESTAMP NOT NULL,
				last_used_at TIMESTAMP,
				failed_attempts INTEGER DEFAULT 0,
				locked_until TIMESTAMP
			)
		`, r.schema),

		// Federated model updates table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.federated_model_updates (
				id VARCHAR(36) PRIMARY KEY,
				client_id VARCHAR(100) NOT NULL,
				model_version VARCHAR(50) NOT NULL,
				update_data BYTEA,
				update_hash VARCHAR(256) NOT NULL,
				sample_count INTEGER NOT NULL,
				accuracy DECIMAL(5,4) NOT NULL,
				client_version VARCHAR(50),
				device_info JSONB,
				submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
				status VARCHAR(20) NOT NULL DEFAULT 'pending',
				processed_at TIMESTAMP,
				weight DECIMAL(5,4) DEFAULT 1.0
			)
		`, r.schema),

		// Consent records table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.consent_records (
				id VARCHAR(36) PRIMARY KEY,
				identity_id VARCHAR(36) NOT NULL,
				consent_type VARCHAR(50) NOT NULL,
				purpose TEXT NOT NULL,
				data_categories JSONB NOT NULL,
				third_parties JSONB,
				geo_scope VARCHAR(100),
				retention_period VARCHAR(50) NOT NULL,
				consent_given BOOLEAN NOT NULL,
				consent_date TIMESTAMP NOT NULL,
				withdrawal_date TIMESTAMP,
				ip_address VARCHAR(45),
				user_agent TEXT,
				version INTEGER DEFAULT 1,
				UNIQUE(identity_id, consent_type)
			)
		`, r.schema),

		// Identity events table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.identity_events (
				id VARCHAR(36) PRIMARY KEY,
				identity_id VARCHAR(36) NOT NULL,
				event_type VARCHAR(50) NOT NULL,
				event_data JSONB,
				source_system VARCHAR(100) NOT NULL,
				correlation_id VARCHAR(36),
				actor_id VARCHAR(100),
				ip_address VARCHAR(45),
				created_at TIMESTAMP NOT NULL DEFAULT NOW()
			)
		`, r.schema),

		// Create indexes
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_identities_national_id ON %s.identities(national_id)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_identities_status ON %s.identities(status)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_identities_created_at ON %s.identities(created_at)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_history_identity_id ON %s.identity_history(identity_id)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_verifications_identity_id ON %s.identity_verifications(identity_id)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_biometric_identity_id ON %s.biometric_templates(identity_id)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_fraud_identity_id ON %s.fraud_alerts(identity_id)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_fraud_status ON %s.fraud_alerts(status)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_sessions_token ON %s.authentication_sessions(session_token)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_sessions_identity_id ON %s.authentication_sessions(identity_id)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_sessions_expires ON %s.authentication_sessions(expires_at)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_events_identity_id ON %s.identity_events(identity_id)`, r.schema, r.schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_events_created_at ON %s.identity_events(created_at)`, r.schema, r.schema),
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

// CreateIdentity creates a new identity record
func (r *pgxRepository) CreateIdentity(ctx context.Context, identity *domain.Identity) error {
	if identity.ID == "" {
		identity.ID = uuid.New().String()
	}
	
	identity.CreatedAt = time.Now()
	identity.UpdatedAt = time.Now()

	currentAddressJSON, err := json.Marshal(identity.CurrentAddress)
	if err != nil {
		return fmt.Errorf("failed to marshal current address: %w", err)
	}

	var permanentAddressJSON []byte
	if identity.PermanentAddress != nil {
		permanentAddressJSON, err = json.Marshal(identity.PermanentAddress)
		if err != nil {
			return fmt.Errorf("failed to marshal permanent address: %w", err)
		}
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.identities (
			id, national_id, legacy_id, first_name, middle_name, last_name,
			date_of_birth, place_of_birth, gender, nationality, ethnicity,
			marital_status, current_address, permanent_address, email, phone,
			mobile, passport_number, passport_expiry, driver_license,
			biometric_id, photo_url, fingerprint_hash, iris_scan_hash,
			facial_template_hash, status, verification_level, verified_at,
			verified_by, fraud_score, risk_level, last_risk_assessment,
			consent_given, consent_date, data_opt_out, created_at, updated_at,
			created_by, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28,
			$29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41
		)
	`, r.schema)

	_, err = r.pool.Exec(ctx, query,
		identity.ID,
		identity.NationalID,
		identity.LegacyID,
		identity.FirstName,
		identity.MiddleName,
		identity.LastName,
		identity.DateOfBirth,
		identity.PlaceOfBirth,
		identity.Gender,
		identity.Nationality,
		identity.Ethnicity,
		identity.MaritalStatus,
		currentAddressJSON,
		permanentAddressJSON,
		identity.Email,
		identity.Phone,
		identity.Mobile,
		identity.PassportNumber,
		identity.PassportExpiry,
		identity.DriverLicense,
		identity.BiometricID,
		identity.PhotoURL,
		identity.FingerprintHash,
		identity.IrisScanHash,
		identity.FacialTemplateHash,
		identity.Status,
		identity.VerificationLevel,
		identity.VerifiedAt,
		identity.VerifiedBy,
		identity.FraudScore,
		identity.RiskLevel,
		identity.LastRiskAssessment,
		identity.ConsentGiven,
		identity.ConsentDate,
		identity.DataOptOut,
		identity.CreatedAt,
		identity.UpdatedAt,
		identity.CreatedBy,
		identity.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to create identity: %w", err)
	}

	// Create initial history entry
	history := &domain.IdentityHistory{
		ID:          uuid.New().String(),
		IdentityID:  identity.ID,
		Action:      "CREATE",
		ChangedBy:   identity.CreatedBy,
		ChangedAt:   time.Now(),
		Reason:      sql.NullString{String: "Initial creation", Valid: true},
	}

	return r.CreateIdentityHistory(ctx, history)
}

// GetIdentity retrieves an identity by ID
func (r *pgxRepository) GetIdentity(ctx context.Context, id string) (*domain.Identity, error) {
	query := fmt.Sprintf(`
		SELECT id, national_id, legacy_id, first_name, middle_name, last_name,
			date_of_birth, place_of_birth, gender, nationality, ethnicity,
			marital_status, current_address, permanent_address, email, phone,
			mobile, passport_number, passport_expiry, driver_license,
			biometric_id, photo_url, fingerprint_hash, iris_scan_hash,
			facial_template_hash, status, verification_level, verified_at,
			verified_by, fraud_score, risk_level, last_risk_assessment,
			consent_given, consent_date, data_opt_out, created_at, updated_at,
			created_by, updated_by, version, deleted_at, is_deleted
		FROM %s.identities
		WHERE id = $1 AND is_deleted = FALSE
	`, r.schema)

	var identity domain.Identity
	var currentAddressJSON, permanentAddressJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&identity.ID,
		&identity.NationalID,
		&identity.LegacyID,
		&identity.FirstName,
		&identity.MiddleName,
		&identity.LastName,
		&identity.DateOfBirth,
		&identity.PlaceOfBirth,
		&identity.Gender,
		&identity.Nationality,
		&identity.Ethnicity,
		&identity.MaritalStatus,
		&currentAddressJSON,
		&permanentAddressJSON,
		&identity.Email,
		&identity.Phone,
		&identity.Mobile,
		&identity.PassportNumber,
		&identity.PassportExpiry,
		&identity.DriverLicense,
		&identity.BiometricID,
		&identity.PhotoURL,
		&identity.FingerprintHash,
		&identity.IrisScanHash,
		&identity.FacialTemplateHash,
		&identity.Status,
		&identity.VerificationLevel,
		&identity.VerifiedAt,
		&identity.VerifiedBy,
		&identity.FraudScore,
		&identity.RiskLevel,
		&identity.LastRiskAssessment,
		&identity.ConsentGiven,
		&identity.ConsentDate,
		&identity.DataOptOut,
		&identity.CreatedAt,
		&identity.UpdatedAt,
		&identity.CreatedBy,
		&identity.UpdatedBy,
		&identity.Version,
		&identity.DeletedAt,
		&identity.IsDeleted,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}

	// Unmarshal addresses
	if err := json.Unmarshal(currentAddressJSON, &identity.CurrentAddress); err != nil {
		return nil, fmt.Errorf("failed to unmarshal current address: %w", err)
	}
	if permanentAddressJSON != nil {
		if err := json.Unmarshal(permanentAddressJSON, &identity.PermanentAddress); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permanent address: %w", err)
		}
	}

	return &identity, nil
}

// GetIdentityByNationalID retrieves an identity by national ID number
func (r *pgxRepository) GetIdentityByNationalID(ctx context.Context, nationalID string) (*domain.Identity, error) {
	query := fmt.Sprintf(`
		SELECT id FROM %s.identities WHERE national_id = $1 AND is_deleted = FALSE
	`, r.schema)

	var id string
	err := r.pool.QueryRow(ctx, query, nationalID).Scan(&id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get identity by national ID: %w", err)
	}

	return r.GetIdentity(ctx, id)
}

// UpdateIdentity updates an existing identity record
func (r *pgxRepository) UpdateIdentity(ctx context.Context, identity *domain.Identity) error {
	identity.UpdatedAt = time.Now()
	identity.Version++

	currentAddressJSON, err := json.Marshal(identity.CurrentAddress)
	if err != nil {
		return fmt.Errorf("failed to marshal current address: %w", err)
	}

	var permanentAddressJSON []byte
	if identity.PermanentAddress != nil {
		permanentAddressJSON, err = json.Marshal(identity.PermanentAddress)
		if err != nil {
			return fmt.Errorf("failed to marshal permanent address: %w", err)
		}
	}

	query := fmt.Sprintf(`
		UPDATE %s.identities SET
			national_id = $2, legacy_id = $3, first_name = $4, middle_name = $5,
			last_name = $6, date_of_birth = $7, place_of_birth = $8, gender = $9,
			nationality = $10, ethnicity = $11, marital_status = $12,
			current_address = $13, permanent_address = $14, email = $15,
			phone = $16, mobile = $17, passport_number = $18, passport_expiry = $19,
			driver_license = $20, biometric_id = $21, photo_url = $22,
			fingerprint_hash = $23, iris_scan_hash = $24, facial_template_hash = $25,
			status = $26, verification_level = $27, verified_at = $28,
			verified_by = $29, fraud_score = $30, risk_level = $31,
			last_risk_assessment = $32, consent_given = $33, consent_date = $34,
			data_opt_out = $35, updated_at = $36, updated_by = $37, version = $38
		WHERE id = $1 AND is_deleted = FALSE
	`, r.schema)

	result, err := r.pool.Exec(ctx, query,
		identity.ID,
		identity.NationalID,
		identity.LegacyID,
		identity.FirstName,
		identity.MiddleName,
		identity.LastName,
		identity.DateOfBirth,
		identity.PlaceOfBirth,
		identity.Gender,
		identity.Nationality,
		identity.Ethnicity,
		identity.MaritalStatus,
		currentAddressJSON,
		permanentAddressJSON,
		identity.Email,
		identity.Phone,
		identity.Mobile,
		identity.PassportNumber,
		identity.PassportExpiry,
		identity.DriverLicense,
		identity.BiometricID,
		identity.PhotoURL,
		identity.FingerprintHash,
		identity.IrisScanHash,
		identity.FacialTemplateHash,
		identity.Status,
		identity.VerificationLevel,
		identity.VerifiedAt,
		identity.VerifiedBy,
		identity.FraudScore,
		identity.RiskLevel,
		identity.LastRiskAssessment,
		identity.ConsentGiven,
		identity.ConsentDate,
		identity.DataOptOut,
		identity.UpdatedAt,
		identity.UpdatedBy,
		identity.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to update identity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("identity not found")
	}

	return nil
}

// DeleteIdentity soft-deletes an identity record
func (r *pgxRepository) DeleteIdentity(ctx context.Context, id string) error {
	query := fmt.Sprintf(`
		UPDATE %s.identities SET
			is_deleted = TRUE, deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND is_deleted = FALSE
	`, r.schema)

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete identity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("identity not found")
	}

	return nil
}

// GetIdentityHistory retrieves the history of an identity
func (r *pgxRepository) GetIdentityHistory(ctx context.Context, id string, limit, offset int) ([]domain.IdentityHistory, error) {
	query := fmt.Sprintf(`
		SELECT id, identity_id, action, field_changed, old_value, new_value,
			changed_by, changed_at, reason, ip_address, user_agent
		FROM %s.identity_history
		WHERE identity_id = $1
		ORDER BY changed_at DESC
		LIMIT $2 OFFSET $3
	`, r.schema)

	rows, err := r.pool.Query(ctx, query, id, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity history: %w", err)
	}
	defer rows.Close()

	var history []domain.IdentityHistory
	for rows.Next() {
		var h domain.IdentityHistory
		if err := rows.Scan(
			&h.ID, &h.IdentityID, &h.Action, &h.FieldChanged,
			&h.OldValue, &h.NewValue, &h.ChangedBy, &h.ChangedAt,
			&h.Reason, &h.IPAddress, &h.UserAgent,
		); err != nil {
			return nil, fmt.Errorf("failed to scan history record: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

// CreateIdentityHistory creates a new history entry
func (r *pgxRepository) CreateIdentityHistory(ctx context.Context, history *domain.IdentityHistory) error {
	if history.ID == "" {
		history.ID = uuid.New().String()
	}
	if history.ChangedAt.IsZero() {
		history.ChangedAt = time.Now()
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.identity_history (
			id, identity_id, action, field_changed, old_value, new_value,
			changed_by, changed_at, reason, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, r.schema)

	_, err := r.pool.Exec(ctx, query,
		history.ID,
		history.IdentityID,
		history.Action,
		history.FieldChanged,
		history.OldValue,
		history.NewValue,
		history.ChangedBy,
		history.ChangedAt,
		history.Reason,
		history.IPAddress,
		history.UserAgent,
	)

	if err != nil {
		return fmt.Errorf("failed to create identity history: %w", err)
	}

	return nil
}

// CreateVerification creates a new verification record
func (r *pgxRepository) CreateVerification(ctx context.Context, verification *domain.IdentityVerification) error {
	if verification.ID == "" {
		verification.ID = uuid.New().String()
	}
	verification.CreatedAt = time.Now()
	verification.VerifiedAt = time.Now()

	evidenceJSON, err := json.Marshal(verification.Evidence)
	if err != nil {
		return fmt.Errorf("failed to marshal evidence: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.identity_verifications (
			id, identity_id, verification_type, method, status, score,
			threshold, passed, failure_reason, evidence, verified_at,
			expires_at, verified_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, r.schema)

	_, err = r.pool.Exec(ctx, query,
		verification.ID,
		verification.IdentityID,
		verification.VerificationType,
		verification.Method,
		verification.Status,
		verification.Score,
		verification.Threshold,
		verification.Passed,
		verification.FailureReason,
		evidenceJSON,
		verification.VerifiedAt,
		verification.ExpiresAt,
		verification.VerifiedBy,
		verification.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create verification: %w", err)
	}

	return nil
}

// GetVerifications retrieves all verifications for an identity
func (r *pgxRepository) GetVerifications(ctx context.Context, identityID string) ([]domain.IdentityVerification, error) {
	query := fmt.Sprintf(`
		SELECT id, identity_id, verification_type, method, status, score,
			threshold, passed, failure_reason, evidence, verified_at,
			expires_at, verified_by, created_at
		FROM %s.identity_verifications
		WHERE identity_id = $1
		ORDER BY verified_at DESC
	`, r.schema)

	rows, err := r.pool.Query(ctx, query, identityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get verifications: %w", err)
	}
	defer rows.Close()

	var verifications []domain.IdentityVerification
	for rows.Next() {
		var v domain.IdentityVerification
		var evidenceJSON []byte

		if err := rows.Scan(
			&v.ID, &v.IdentityID, &v.VerificationType, &v.Method, &v.Status,
			&v.Score, &v.Threshold, &v.Passed, &v.FailureReason, &evidenceJSON,
			&v.VerifiedAt, &v.ExpiresAt, &v.VerifiedBy, &v.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan verification: %w", err)
		}

		if evidenceJSON != nil {
			if err := json.Unmarshal(evidenceJSON, &v.Evidence); err != nil {
				return nil, fmt.Errorf("failed to unmarshal evidence: %w", err)
			}
		}

		verifications = append(verifications, v)
	}

	return verifications, nil
}

// SaveBiometricTemplate saves a biometric template
func (r *pgxRepository) SaveBiometricTemplate(ctx context.Context, template *domain.BiometricTemplate) error {
	if template.ID == "" {
		template.ID = uuid.New().String()
	}
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	template.IsActive = true

	query := fmt.Sprintf(`
		INSERT INTO %s.biometric_templates (
			id, identity_id, biometric_type, template_hash, template_data,
			encryption_key_id, quality_score, created_at, updated_at,
			expires_at, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (identity_id, biometric_type)
		DO UPDATE SET
			template_hash = EXCLUDED.template_hash,
			template_data = EXCLUDED.template_data,
			encryption_key_id = EXCLUDED.encryption_key_id,
			quality_score = EXCLUDED.quality_score,
			updated_at = EXCLUDED.updated_at,
			is_active = EXCLUDED.is_active
	`, r.schema)

	_, err := r.pool.Exec(ctx, query,
		template.ID,
		template.IdentityID,
		template.BiometricType,
		template.TemplateHash,
		template.TemplateData,
		template.EncryptionKeyID,
		template.QualityScore,
		template.CreatedAt,
		template.UpdatedAt,
		template.ExpiresAt,
		template.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to save biometric template: %w", err)
	}

	return nil
}

// GetBiometricTemplate retrieves a biometric template
func (r *pgxRepository) GetBiometricTemplate(ctx context.Context, identityID, biometricType string) (*domain.BiometricTemplate, error) {
	query := fmt.Sprintf(`
		SELECT id, identity_id, biometric_type, template_hash, template_data,
			encryption_key_id, quality_score, created_at, updated_at,
			expires_at, is_active
		FROM %s.biometric_templates
		WHERE identity_id = $1 AND biometric_type = $2 AND is_active = TRUE
	`, r.schema)

	var template domain.BiometricTemplate
	err := r.pool.QueryRow(ctx, query, identityID, biometricType).Scan(
		&template.ID,
		&template.IdentityID,
		&template.BiometricType,
		&template.TemplateHash,
		&template.TemplateData,
		&template.EncryptionKeyID,
		&template.QualityScore,
		&template.CreatedAt,
		&template.UpdatedAt,
		&template.ExpiresAt,
		&template.IsActive,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get biometric template: %w", err)
	}

	return &template, nil
}

// SaveBehavioralProfile saves or updates a behavioral profile
func (r *pgxRepository) SaveBehavioralProfile(ctx context.Context, profile *domain.BehavioralProfile) error {
	if profile.ID == "" {
		profile.ID = uuid.New().String()
	}
	profile.UpdatedAt = time.Now()

	keystrokeJSON, err := json.Marshal(profile.KeystrokeProfile)
	if err != nil {
		return fmt.Errorf("failed to marshal keystroke profile: %w", err)
	}

	mouseJSON, err := json.Marshal(profile.MouseProfile)
	if err != nil {
		return fmt.Errorf("failed to marshal mouse profile: %w", err)
	}

	touchJSON, err := json.Marshal(profile.TouchProfile)
	if err != nil {
		return fmt.Errorf("failed to marshal touch profile: %w", err)
	}

	gaitJSON, err := json.Marshal(profile.GaitProfile)
	if err != nil {
		return fmt.Errorf("failed to marshal gait profile: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.behavioral_profiles (
			id, identity_id, keystroke_profile, mouse_profile, touch_profile,
			gait_profile, overall_score, anomaly_count, last_anomaly_at,
			baseline_established_at, updated_at, sample_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (identity_id)
		DO UPDATE SET
			keystroke_profile = EXCLUDED.keystroke_profile,
			mouse_profile = EXCLUDED.mouse_profile,
			touch_profile = EXCLUDED.touch_profile,
			gait_profile = EXCLUDED.gait_profile,
			overall_score = EXCLUDED.overall_score,
			anomaly_count = EXCLUDED.anomaly_count,
			last_anomaly_at = EXCLUDED.last_anomaly_at,
			baseline_established_at = EXCLUDED.baseline_established_at,
			updated_at = EXCLUDED.updated_at,
			sample_count = EXCLUDED.sample_count
	`, r.schema)

	_, err = r.pool.Exec(ctx, query,
		profile.ID,
		profile.IdentityID,
		keystrokeJSON,
		mouseJSON,
		touchJSON,
		gaitJSON,
		profile.OverallScore,
		profile.AnomalyCount,
		profile.LastAnomalyAt,
		profile.BaselineEstablishedAt,
		profile.UpdatedAt,
		profile.SampleCount,
	)

	if err != nil {
		return fmt.Errorf("failed to save behavioral profile: %w", err)
	}

	return nil
}

// GetBehavioralProfile retrieves a behavioral profile
func (r *pgxRepository) GetBehavioralProfile(ctx context.Context, identityID string) (*domain.BehavioralProfile, error) {
	query := fmt.Sprintf(`
		SELECT id, identity_id, keystroke_profile, mouse_profile, touch_profile,
			gait_profile, overall_score, anomaly_count, last_anomaly_at,
			baseline_established_at, updated_at, sample_count
		FROM %s.behavioral_profiles
		WHERE identity_id = $1
	`, r.schema)

	var profile domain.BehavioralProfile
	var keystrokeJSON, mouseJSON, touchJSON, gaitJSON []byte

	err := r.pool.QueryRow(ctx, query, identityID).Scan(
		&profile.ID,
		&profile.IdentityID,
		&keystrokeJSON,
		&mouseJSON,
		&touchJSON,
		&gaitJSON,
		&profile.OverallScore,
		&profile.AnomalyCount,
		&profile.LastAnomalyAt,
		&profile.BaselineEstablishedAt,
		&profile.UpdatedAt,
		&profile.SampleCount,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get behavioral profile: %w", err)
	}

	// Unmarshal profiles
	if keystrokeJSON != nil {
		if err := json.Unmarshal(keystrokeJSON, &profile.KeystrokeProfile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal keystroke profile: %w", err)
		}
	}
	if mouseJSON != nil {
		if err := json.Unmarshal(mouseJSON, &profile.MouseProfile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal mouse profile: %w", err)
		}
	}
	if touchJSON != nil {
		if err := json.Unmarshal(touchJSON, &profile.TouchProfile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal touch profile: %w", err)
		}
	}
	if gaitJSON != nil {
		if err := json.Unmarshal(gaitJSON, &profile.GaitProfile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal gait profile: %w", err)
		}
	}

	return &profile, nil
}

// CreateFraudAlert creates a new fraud alert
func (r *pgxRepository) CreateFraudAlert(ctx context.Context, alert *domain.FraudAlert) error {
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()
	alert.Status = "open"

	query := fmt.Sprintf(`
		INSERT INTO %s.fraud_alerts (
			id, identity_id, alert_type, severity, score, indicator,
			description, evidence, related_event_id, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, r.schema)

	_, err := r.pool.Exec(ctx, query,
		alert.ID,
		alert.IdentityID,
		alert.AlertType,
		alert.Severity,
		alert.Score,
		alert.Indicator,
		alert.Description,
		alert.Evidence,
		alert.RelatedEventID,
		alert.Status,
		alert.CreatedAt,
		alert.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create fraud alert: %w", err)
	}

	return nil
}

// GetFraudAlerts retrieves fraud alerts for an identity
func (r *pgxRepository) GetFraudAlerts(ctx context.Context, identityID, status string) ([]domain.FraudAlert, error) {
	var query strings.Builder
	query.WriteString(fmt.Sprintf(`
		SELECT id, identity_id, alert_type, severity, score, indicator,
			description, evidence, related_event_id, status, resolved_at,
			resolved_by, resolution, created_at, updated_at
		FROM %s.fraud_alerts
		WHERE identity_id = $1
	`, r.schema))

	args := []interface{}{identityID}
	if status != "" {
		query.WriteString(" AND status = $2")
		args = append(args, status)
	}

	query.WriteString(" ORDER BY created_at DESC")

	rows, err := r.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get fraud alerts: %w", err)
	}
	defer rows.Close()

	var alerts []domain.FraudAlert
	for rows.Next() {
		var alert domain.FraudAlert
		if err := rows.Scan(
			&alert.ID, &alert.IdentityID, &alert.AlertType, &alert.Severity,
			&alert.Score, &alert.Indicator, &alert.Description, &alert.Evidence,
			&alert.RelatedEventID, &alert.Status, &alert.ResolvedAt,
			&alert.ResolvedBy, &alert.Resolution, &alert.CreatedAt, &alert.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan fraud alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// ResolveFraudAlert resolves a fraud alert
func (r *pgxRepository) ResolveFraudAlert(ctx context.Context, alertID, resolvedBy, resolution string) error {
	query := fmt.Sprintf(`
		UPDATE %s.fraud_alerts SET
			status = 'resolved', resolved_at = NOW(), resolved_by = $2,
			resolution = $3, updated_at = NOW()
		WHERE id = $1 AND status = 'open'
	`, r.schema)

	result, err := r.pool.Exec(ctx, query, alertID, resolvedBy, resolution)
	if err != nil {
		return fmt.Errorf("failed to resolve fraud alert: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("alert not found or already resolved")
	}

	return nil
}

// GetFraudScore retrieves the fraud score for an identity
func (r *pgxRepository) GetFraudScore(ctx context.Context, identityID string) (int, error) {
	query := fmt.Sprintf(`
		SELECT COALESCE(MAX(score), 0) FROM %s.fraud_alerts
		WHERE identity_id = $1 AND status = 'open'
	`, r.schema)

	var score int
	err := r.pool.QueryRow(ctx, query, identityID).Scan(&score)
	if err != nil {
		return 0, fmt.Errorf("failed to get fraud score: %w", err)
	}

	return score, nil
}

// CreateSession creates a new authentication session
func (r *pgxRepository) CreateSession(ctx context.Context, session *domain.AuthenticationSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	session.CreatedAt = time.Now()
	session.LastActivityAt = time.Now()

	deviceInfoJSON, err := json.Marshal(session.DeviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal device info: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.authentication_sessions (
			id, identity_id, session_token, refresh_token, ip_address,
			user_agent, device_info, mfa_enabled, mfa_passed, behavioral_score,
			risk_score, expires_at, created_at, last_activity_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, r.schema)

	_, err = r.pool.Exec(ctx, query,
		session.ID,
		session.IdentityID,
		session.SessionToken,
		session.RefreshToken,
		session.IPAddress,
		session.UserAgent,
		deviceInfoJSON,
		session.MFAEnabled,
		session.MFAPassed,
		session.BehavioralScore,
		session.RiskScore,
		session.ExpiresAt,
		session.CreatedAt,
		session.LastActivityAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSession retrieves a session by token
func (r *pgxRepository) GetSession(ctx context.Context, sessionToken string) (*domain.AuthenticationSession, error) {
	query := fmt.Sprintf(`
		SELECT id, identity_id, session_token, refresh_token, ip_address,
			user_agent, device_info, mfa_enabled, mfa_passed, behavioral_score,
			risk_score, expires_at, created_at, last_activity_at,
			terminated_at, termination_reason
		FROM %s.authentication_sessions
		WHERE session_token = $1 AND terminated_at IS NULL AND expires_at > NOW()
	`, r.schema)

	var session domain.AuthenticationSession
	var deviceInfoJSON []byte

	err := r.pool.QueryRow(ctx, query, sessionToken).Scan(
		&session.ID,
		&session.IdentityID,
		&session.SessionToken,
		&session.RefreshToken,
		&session.IPAddress,
		&session.UserAgent,
		&deviceInfoJSON,
		&session.MFAEnabled,
		&session.MFAPassed,
		&session.BehavioralScore,
		&session.RiskScore,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastActivityAt,
		&session.TerminatedAt,
		&session.TerminationReason,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if deviceInfoJSON != nil {
		if err := json.Unmarshal(deviceInfoJSON, &session.DeviceInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
		}
	}

	return &session, nil
}

// UpdateSession updates a session
func (r *pgxRepository) UpdateSession(ctx context.Context, session *domain.AuthenticationSession) error {
	session.LastActivityAt = time.Now()

	query := fmt.Sprintf(`
		UPDATE %s.authentication_sessions SET
			last_activity_at = $2, behavioral_score = $3, risk_score = $4
		WHERE id = $1
	`, r.schema)

	_, err := r.pool.Exec(ctx, query, session.ID, session.LastActivityAt, session.BehavioralScore, session.RiskScore)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// TerminateSession terminates a specific session
func (r *pgxRepository) TerminateSession(ctx context.Context, sessionToken, reason string) error {
	query := fmt.Sprintf(`
		UPDATE %s.authentication_sessions SET
			terminated_at = NOW(), termination_reason = $2
		WHERE session_token = $1 AND terminated_at IS NULL
	`, r.schema)

	_, err := r.pool.Exec(ctx, query, sessionToken, reason)
	if err != nil {
		return fmt.Errorf("failed to terminate session: %w", err)
	}

	return nil
}

// TerminateAllSessions terminates all sessions for an identity
func (r *pgxRepository) TerminateAllSessions(ctx context.Context, identityID, reason string) error {
	query := fmt.Sprintf(`
		UPDATE %s.authentication_sessions SET
			terminated_at = NOW(), termination_reason = $2
		WHERE identity_id = $1 AND terminated_at IS NULL
	`, r.schema)

	_, err := r.pool.Exec(ctx, query, identityID, reason)
	if err != nil {
		return fmt.Errorf("failed to terminate all sessions: %w", err)
	}

	return nil
}

// SaveMFAToken saves an MFA token
func (r *pgxRepository) SaveMFAToken(ctx context.Context, token *domain.MFAToken) error {
	if token.ID == "" {
		token.ID = uuid.New().String()
	}
	token.CreatedAt = time.Now()

	query := fmt.Sprintf(`
		INSERT INTO %s.mfa_tokens (
			id, identity_id, token_type, token_hash, secret, phone_number,
			email_address, is_active, verified_at, created_at, expires_at,
			last_used_at, failed_attempts, locked_until
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, r.schema)

	_, err := r.pool.Exec(ctx, query,
		token.ID,
		token.IdentityID,
		token.TokenType,
		token.TokenHash,
		token.Secret,
		token.PhoneNumber,
		token.EmailAddress,
		token.IsActive,
		token.VerifiedAt,
		token.CreatedAt,
		token.ExpiresAt,
		token.LastUsedAt,
		token.FailedAttempts,
		token.LockedUntil,
	)

	if err != nil {
		return fmt.Errorf("failed to save MFA token: %w", err)
	}

	return nil
}

// GetMFAToken retrieves an MFA token
func (r *pgxRepository) GetMFAToken(ctx context.Context, identityID, tokenType string) (*domain.MFAToken, error) {
	query := fmt.Sprintf(`
		SELECT id, identity_id, token_type, token_hash, secret, phone_number,
			email_address, is_active, verified_at, created_at, expires_at,
			last_used_at, failed_attempts, locked_until
		FROM %s.mfa_tokens
		WHERE identity_id = $1 AND token_type = $2 AND is_active = TRUE
			AND expires_at > NOW()
	`, r.schema)

	var token domain.MFAToken
	err := r.pool.QueryRow(ctx, query, identityID, tokenType).Scan(
		&token.ID,
		&token.IdentityID,
		&token.TokenType,
		&token.TokenHash,
		&token.Secret,
		&token.PhoneNumber,
		&token.EmailAddress,
		&token.IsActive,
		&token.VerifiedAt,
		&token.CreatedAt,
		&token.ExpiresAt,
		&token.LastUsedAt,
		&token.FailedAttempts,
		&token.LockedUntil,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get MFA token: %w", err)
	}

	return &token, nil
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

// GetMultiple retrieves multiple values
func (r *RedisCache) GetMultiple(ctx context.Context, keys []string) ([]string, error) {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = r.keyPrefix + key
	}

	results, err := r.client.MGet(ctx, prefixedKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple values: %w", err)
	}

	values := make([]string, len(results))
	for i, result := range results {
		if result != nil {
			values[i] = result.(string)
		}
	}

	return values, nil
}

// Log provides a simple logging function for debugging
func Log(format string, args ...interface{}) {
	log.Printf("[IdentityRepo] "+format, args...)
}
