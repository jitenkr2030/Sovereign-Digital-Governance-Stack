package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic-platform/services/security/key-management/internal/config"
	"github.com/csic-platform/services/security/key-management/internal/core/domain"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresRepository implements the KeyRepository interface using PostgreSQL
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

// CreateKey creates a new key in the vault
func (r *PostgresRepository) CreateKey(ctx context.Context, key *domain.Key) error {
	metadata, _ := json.Marshal(key.Metadata)

	query := `
		INSERT INTO key_vault (id, alias, algorithm, current_version, status, usage,
		                     description, metadata, created_at, updated_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		key.ID, key.Alias, key.Algorithm, key.CurrentVersion, key.Status,
		key.Usage, key.Description, metadata, key.CreatedAt, key.UpdatedAt, key.ExpiresAt,
	)
	return err
}

// GetKey retrieves a key by ID
func (r *PostgresRepository) GetKey(ctx context.Context, id string) (*domain.Key, error) {
	query := `
		SELECT id, alias, algorithm, current_version, status, usage, description,
		       metadata, created_at, updated_at, rotated_at, expires_at
		FROM key_vault WHERE id = $1
	`

	var key domain.Key
	var usage []byte
	var metadata []byte
	var rotatedAt sql.NullTime
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&key.ID, &key.Alias, &key.Algorithm, &key.CurrentVersion, &key.Status,
		&usage, &key.Description, &metadata, &key.CreatedAt, &key.UpdatedAt,
		&rotatedAt, &expiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	json.Unmarshal(usage, &key.Usage)
	json.Unmarshal(metadata, &key.Metadata)

	if rotatedAt.Valid {
		key.RotatedAt = &rotatedAt.Time
	}
	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}

	return &key, nil
}

// GetKeyByAlias retrieves a key by alias
func (r *PostgresRepository) GetKeyByAlias(ctx context.Context, alias string) (*domain.Key, error) {
	query := `
		SELECT id, alias, algorithm, current_version, status, usage, description,
		       metadata, created_at, updated_at, rotated_at, expires_at
		FROM key_vault WHERE alias = $1
	`

	var key domain.Key
	var usage []byte
	var metadata []byte
	var rotatedAt sql.NullTime
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, alias).Scan(
		&key.ID, &key.Alias, &key.Algorithm, &key.CurrentVersion, &key.Status,
		&usage, &key.Description, &metadata, &key.CreatedAt, &key.UpdatedAt,
		&rotatedAt, &expiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key by alias: %w", err)
	}

	json.Unmarshal(usage, &key.Usage)
	json.Unmarshal(metadata, &key.Metadata)

	if rotatedAt.Valid {
		key.RotatedAt = &rotatedAt.Time
	}
	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}

	return &key, nil
}

// ListKeys retrieves a paginated list of keys
func (r *PostgresRepository) ListKeys(ctx context.Context, page, pageSize int) ([]*domain.Key, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	countQuery := "SELECT COUNT(*) FROM key_vault"
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count keys: %w", err)
	}

	// Get keys
	query := `
		SELECT id, alias, algorithm, current_version, status, usage, description,
		       metadata, created_at, updated_at, rotated_at, expires_at
		FROM key_vault ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list keys: %w", err)
	}
	defer rows.Close()

	var keys []*domain.Key
	for rows.Next() {
		var key domain.Key
		var usage []byte
		var metadata []byte
		var rotatedAt sql.NullTime
		var expiresAt sql.NullTime

		if err := rows.Scan(
			&key.ID, &key.Alias, &key.Algorithm, &key.CurrentVersion, &key.Status,
			&usage, &key.Description, &metadata, &key.CreatedAt, &key.UpdatedAt,
			&rotatedAt, &expiresAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan key: %w", err)
		}

		json.Unmarshal(usage, &key.Usage)
		json.Unmarshal(metadata, &key.Metadata)

		if rotatedAt.Valid {
			key.RotatedAt = &rotatedAt.Time
		}
		if expiresAt.Valid {
			key.ExpiresAt = &expiresAt.Time
		}

		keys = append(keys, &key)
	}

	return keys, total, nil
}

// UpdateKey updates a key
func (r *PostgresRepository) UpdateKey(ctx context.Context, key *domain.Key) error {
	metadata, _ := json.Marshal(key.Metadata)

	query := `
		UPDATE key_vault SET alias=$1, current_version=$2, status=$3, usage=$4,
		       description=$5, metadata=$6, updated_at=$7, rotated_at=$8, expires_at=$9
		WHERE id=$10
	`

	_, err := r.db.ExecContext(ctx, query,
		key.Alias, key.CurrentVersion, key.Status, key.Usage,
		key.Description, metadata, key.UpdatedAt, key.RotatedAt, key.ExpiresAt, key.ID,
	)
	return err
}

// DeleteKey deletes a key (soft delete by archiving)
func (r *PostgresRepository) DeleteKey(ctx context.Context, id string) error {
	query := `UPDATE key_vault SET status=$1, updated_at=$2 WHERE id=$3`
	_, err := r.db.ExecContext(ctx, query, domain.KeyStatusArchived, time.Now(), id)
	return err
}

// CreateKeyVersion creates a new key version
func (r *PostgresRepository) CreateKeyVersion(ctx context.Context, version *domain.KeyVersion) error {
	query := `
		INSERT INTO key_versions (id, key_id, version, encrypted_material, iv, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		version.ID, version.KeyID, version.Version,
		version.EncryptedMaterial, version.IV, version.CreatedAt,
	)
	return err
}

// GetKeyVersion retrieves a specific version of a key
func (r *PostgresRepository) GetKeyVersion(ctx context.Context, keyID string, version int) (*domain.KeyVersion, error) {
	query := `
		SELECT id, key_id, version, encrypted_material, iv, created_at
		FROM key_versions WHERE key_id=$1 AND version=$2
	`

	var v domain.KeyVersion
	err := r.db.QueryRowContext(ctx, query, keyID, version).Scan(
		&v.ID, &v.KeyID, &v.Version, &v.EncryptedMaterial, &v.IV, &v.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("key version not found: %s v%d", keyID, version)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key version: %w", err)
	}

	return &v, nil
}

// GetLatestKeyVersion retrieves the latest version of a key
func (r *PostgresRepository) GetLatestKeyVersion(ctx context.Context, keyID string) (*domain.KeyVersion, error) {
	query := `
		SELECT id, key_id, version, encrypted_material, iv, created_at
		FROM key_versions WHERE key_id=$1 ORDER BY version DESC LIMIT 1
	`

	var v domain.KeyVersion
	err := r.db.QueryRowContext(ctx, query, keyID).Scan(
		&v.ID, &v.KeyID, &v.Version, &v.EncryptedMaterial, &v.IV, &v.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no key versions found for: %s", keyID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest key version: %w", err)
	}

	return &v, nil
}

// GetAllKeyVersions retrieves all versions of a key
func (r *PostgresRepository) GetAllKeyVersions(ctx context.Context, keyID string) ([]*domain.KeyVersion, error) {
	query := `
		SELECT id, key_id, version, encrypted_material, iv, created_at
		FROM key_versions WHERE key_id=$1 ORDER BY version DESC
	`

	rows, err := r.db.QueryContext(ctx, query, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key versions: %w", err)
	}
	defer rows.Close()

	var versions []*domain.KeyVersion
	for rows.Next() {
		var v domain.KeyVersion
		if err := rows.Scan(
			&v.ID, &v.KeyID, &v.Version, &v.EncryptedMaterial, &v.IV, &v.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan key version: %w", err)
		}
		versions = append(versions, &v)
	}

	return versions, nil
}

// GetKeyPolicy retrieves the policy for a key
func (r *PostgresRepository) GetKeyPolicy(ctx context.Context, keyID string) (*domain.KeyPolicy, error) {
	query := `
		SELECT id, key_id, allowed_roles, allowed_services, max_operations, rate_limit
		FROM key_policies WHERE key_id=$1
	`

	var policy domain.KeyPolicy
	var allowedRoles, allowedServices []byte

	err := r.db.QueryRowContext(ctx, query, keyID).Scan(
		&policy.ID, &policy.KeyID, &allowedRoles, &allowedServices,
		&policy.MaxOperations, &policy.RateLimit,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key policy: %w", err)
	}

	json.Unmarshal(allowedRoles, &policy.AllowedRoles)
	json.Unmarshal(allowedServices, &policy.AllowedServices)

	return &policy, nil
}

// UpdateKeyPolicy updates the policy for a key
func (r *PostgresRepository) UpdateKeyPolicy(ctx context.Context, policy *domain.KeyPolicy) error {
	allowedRoles, _ := json.Marshal(policy.AllowedRoles)
	allowedServices, _ := json.Marshal(policy.AllowedServices)

	query := `
		INSERT INTO key_policies (id, key_id, allowed_roles, allowed_services, max_operations, rate_limit)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (key_id) DO UPDATE SET
			allowed_roles=$3, allowed_services=$4, max_operations=$5, rate_limit=$6
	`

	_, err := r.db.ExecContext(ctx, query,
		policy.ID, policy.KeyID, allowedRoles, allowedServices,
		policy.MaxOperations, policy.RateLimit,
	)
	return err
}

// AuditKeyOperation logs a key operation
func (r *PostgresRepository) AuditKeyOperation(ctx context.Context, operation *domain.KeyOperation) error {
	metadata, _ := json.Marshal(operation.Metadata)

	query := `
		INSERT INTO key_audit_log (id, key_id, operation, actor_id, service_id, success, error, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		operation.ID, operation.KeyID, operation.Operation, operation.ActorID,
		operation.ServiceID, operation.Success, operation.Error, metadata, operation.Timestamp,
	)
	return err
}

// GetKeyAuditLog retrieves the audit log for a key
func (r *PostgresRepository) GetKeyAuditLog(ctx context.Context, keyID string, limit int) ([]*domain.KeyOperation, error) {
	query := `
		SELECT id, key_id, operation, actor_id, service_id, success, error, metadata, timestamp
		FROM key_audit_log WHERE key_id=$1 ORDER BY timestamp DESC LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, keyID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}
	defer rows.Close()

	var operations []*domain.KeyOperation
	for rows.Next() {
		var op domain.KeyOperation
		var metadata []byte
		var errorMsg sql.NullString

		if err := rows.Scan(
			&op.ID, &op.KeyID, &op.Operation, &op.ActorID, &op.ServiceID,
			&op.Success, &errorMsg, &metadata, &op.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if errorMsg.Valid {
			op.Error = errorMsg.String
		}
		json.Unmarshal(metadata, &op.Metadata)

		operations = append(operations, &op)
	}

	return operations, nil
}
