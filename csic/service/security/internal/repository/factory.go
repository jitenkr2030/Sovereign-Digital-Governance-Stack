package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"csic-platform/service/security/internal/domain/models"

	_ "github.com/lib/pq"
)

// RepositoryFactory provides methods to create and configure repository instances.
type RepositoryFactory struct {
	db          *sql.DB
	tablePrefix string
	config      *models.RepositoryConfig
}

// NewRepositoryFactory creates a new repository factory with the provided configuration.
func NewRepositoryFactory(db *sql.DB, tablePrefix string, config *models.RepositoryConfig) *RepositoryFactory {
	if config == nil {
		config = &models.RepositoryConfig{
			HSMProvider:    "software",
			EnableSoftHSM:  true,
			MaxConnections: 25,
			ConnMaxLifetime: time.Hour,
		}
	}

	return &RepositoryFactory{
		db:          db,
		tablePrefix: tablePrefix,
		config:      config,
	}
}

// NewAuditRepository creates a new audit repository instance.
func (f *RepositoryFactory) NewAuditRepository() AuditRepository {
	return NewPostgresAuditRepository(f.db, f.tablePrefix)
}

// NewKeyManagementRepository creates a new key management repository instance.
func (f *RepositoryFactory) NewKeyManagementRepository() KeyManagementRepository {
	return NewPostgresKeyManagementRepository(f.db, f.tablePrefix)
}

// NewSIEMEventRepository creates a new SIEM event repository instance.
func (f *RepositoryFactory) NewSIEMEventRepository() SIEMEventRepository {
	return NewPostgresSIEMEventRepository(f.db, f.tablePrefix)
}

// NewHSMRepository creates a new HSM repository instance based on configuration.
func (f *RepositoryFactory) NewHSMRepository(ctx context.Context) (HSMRepository, error) {
	var hsm HSMRepository

	switch f.config.HSMProvider {
	case "software", "softHSM", "":
		softHSM := NewSoftwareHSMRepository()
		config := &models.HSMConfig{
			Provider:       "software",
			SoftHSMEnabled: f.config.EnableSoftHSM,
		}
		if err := softHSM.Initialize(ctx, config); err != nil {
			return nil, fmt.Errorf("failed to initialize software HSM: %w", err)
		}
		hsm = softHSM
	default:
		return nil, fmt.Errorf("unsupported HSM provider: %s", f.config.HSMProvider)
	}

	return hsm, nil
}

// NewAllRepositories creates and initializes all repositories.
func (f *RepositoryFactory) NewAllRepositories(ctx context.Context) (*AllRepositories, error) {
	auditRepo := f.NewAuditRepository()
	keyMgmtRepo := f.NewKeyManagementRepository()
	siemRepo := f.NewSIEMEventRepository()

	hsmRepo, err := f.NewHSMRepository(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create HSM repository: %w", err)
	}

	return &AllRepositories{
		AuditRepository:      auditRepo,
		KeyManagement:        keyMgmtRepo,
		SIEMEvent:            siemRepo,
		HSM:                  hsmRepo,
		tablePrefix:          f.tablePrefix,
	}, nil
}

// AllRepositories holds references to all repository instances.
type AllRepositories struct {
	AuditRepository      AuditRepository
	KeyManagement        KeyManagementRepository
	SIEMEvent            SIEMEventRepository
	HSM                  HSMRepository
	tablePrefix          string
}

// Close releases all resources held by the repositories.
func (r *AllRepositories) Close() error {
	if hsm, ok := r.HSM.(interface{ Close() error }); ok {
		if err := hsm.Close(); err != nil {
			return fmt.Errorf("failed to close HSM: %w", err)
		}
	}
	return nil
}

// CreateTables creates all required database tables for the security service.
func (r *AllRepositories) CreateTables(ctx context.Context) error {
	// Create audit logs table with hash chain support.
	auditTable := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_audit_logs (
			id VARCHAR(36) PRIMARY KEY,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			event_type VARCHAR(100) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			entity_type VARCHAR(100),
			entity_id VARCHAR(255),
			user_id VARCHAR(255),
			user_email VARCHAR(255),
			action VARCHAR(255),
			resource_type VARCHAR(100),
			resource_id VARCHAR(255),
			old_value TEXT,
			new_value TEXT,
			metadata TEXT,
			ip_address INET,
			user_agent TEXT,
			correlation_id VARCHAR(255),
			session_id VARCHAR(255),
			prev_hash VARCHAR(64) NOT NULL,
			current_hash VARCHAR(64) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS %s_audit_logs_timestamp_idx ON %s_audit_logs(timestamp DESC);
		CREATE INDEX IF NOT EXISTS %s_audit_logs_entity_idx ON %s_audit_logs(entity_type, entity_id);
		CREATE INDEX IF NOT EXISTS %s_audit_logs_user_idx ON %s_audit_logs(user_id);
		CREATE INDEX IF NOT EXISTS %s_audit_logs_prev_hash_idx ON %s_audit_logs(prev_hash);
		CREATE INDEX IF NOT EXISTS %s_audit_logs_event_type_idx ON %s_audit_logs(event_type);
		CREATE INDEX IF NOT EXISTS %s_audit_logs_severity_idx ON %s_audit_logs(severity);
	`, r.tablePrefix, r.tablePrefix, r.tablePrefix, r.tablePrefix, r.tablePrefix,
		r.tablePrefix, r.tablePrefix, r.tablePrefix, r.tablePrefix, r.tablePrefix, r.tablePrefix)

	if _, err := r.db.ExecContext(ctx, auditTable); err != nil {
		return fmt.Errorf("failed to create audit logs table: %w", err)
	}

	// Create chain metadata table.
	chainMetaTable := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_chain_metadata (
			id SERIAL PRIMARY KEY DEFAULT 1,
			last_entry_id VARCHAR(36) NOT NULL,
			last_hash VARCHAR(64) NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			CONSTRAINT single_row CHECK (id = 1)
		);
	`, r.tablePrefix)

	if _, err := r.db.ExecContext(ctx, chainMetaTable); err != nil {
		return fmt.Errorf("failed to create chain metadata table: %w", err)
	}

	// Create crypto keys table.
	keysTable := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_crypto_keys (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			purpose VARCHAR(50) NOT NULL,
			algorithm VARCHAR(50) NOT NULL,
			key_type VARCHAR(50) NOT NULL,
			key_material BYTEA NOT NULL,
			public_key BYTEA,
			version INTEGER NOT NULL DEFAULT 1,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			last_used_at TIMESTAMP WITH TIME ZONE,
			expires_at TIMESTAMP WITH TIME ZONE,
			metadata JSONB DEFAULT '{}',
			hsm_key_id VARCHAR(255),
			rotation_policy TEXT
		);

		CREATE INDEX IF NOT EXISTS %s_crypto_keys_purpose_idx ON %s_crypto_keys(purpose, status);
		CREATE INDEX IF NOT EXISTS %s_crypto_keys_status_idx ON %s_crypto_keys(status);
		CREATE INDEX IF NOT EXISTS %s_crypto_keys_created_idx ON %s_crypto_keys(created_at DESC);
	`, r.tablePrefix, r.tablePrefix, r.tablePrefix, r.tablePrefix, r.tablePrefix, r.tablePrefix, r.tablePrefix)

	if _, err := r.db.ExecContext(ctx, keysTable); err != nil {
		return fmt.Errorf("failed to create crypto keys table: %w", err)
	}

	// Create SIEM events table.
	siemTable := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_siem_events (
			id VARCHAR(36) PRIMARY KEY,
			event_type VARCHAR(100) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			source VARCHAR(255),
			source_ip INET,
			destination_ip INET,
			user_id VARCHAR(255),
			session_id VARCHAR(255),
			action VARCHAR(255),
			outcome VARCHAR(50),
			event_data JSONB NOT NULL,
			raw_event TEXT,
			correlation_id VARCHAR(255),
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			retry_count INTEGER NOT NULL DEFAULT 0,
			last_retry_at TIMESTAMP WITH TIME ZONE,
			last_error TEXT,
			remote_event_id VARCHAR(255),
			forwarded_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS %s_siem_events_status_idx ON %s_siem_events(status, created_at);
		CREATE INDEX IF NOT EXISTS %s_siem_events_type_idx ON %s_siem_events(event_type);
		CREATE INDEX IF NOT EXISTS %s_siem_events_severity_idx ON %s_siem_events(severity);
		CREATE INDEX IF NOT EXISTS %s_siem_events_timestamp_idx ON %s_siem_events(timestamp DESC);
		CREATE INDEX IF NOT EXISTS %s_siem_events_user_idx ON %s_siem_events(user_id);
	`, r.tablePrefix, r.tablePrefix, r.tablePrefix, r.tablePrefix, r.tablePrefix,
		r.tablePrefix, r.tablePrefix, r.tablePrefix, r.tablePrefix, r.tablePrefix)

	if _, err := r.db.ExecContext(ctx, siemTable); err != nil {
		return fmt.Errorf("failed to create SIEM events table: %w", err)
	}

	// Create full-text search index for audit logs.
	ftsIndex := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s_audit_logs_fts_idx ON %s_audit_logs
		USING GIN (TO_TSVECTOR('english', COALESCE(metadata, '') || ' ' || COALESCE(action, '') || ' ' || COALESCE(entity_id, '')))
	`, r.tablePrefix, r.tablePrefix)

	if _, err := r.db.ExecContext(ctx, ftsIndex); err != nil {
		// Ignore error if extension doesn't exist, continue without FTS.
		fmt.Printf("Warning: full-text search index creation skipped: %v\n", err)
	}

	return nil
}
