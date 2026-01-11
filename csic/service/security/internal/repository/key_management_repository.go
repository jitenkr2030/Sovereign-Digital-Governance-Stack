package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"csic-platform/service/security/internal/domain/models"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgreSQL key management repository for cryptographic key lifecycle management.
type PostgresKeyManagementRepository struct {
	db          *sql.DB
	tablePrefix string
}

// NewPostgresKeyManagementRepository creates a new PostgreSQL key management repository.
func NewPostgresKeyManagementRepository(db *sql.DB, tablePrefix string) *PostgresKeyManagementRepository {
	return &PostgresKeyManagementRepository{
		db:          db,
		tablePrefix: tablePrefix,
	}
}

// tableName returns the full table name with prefix.
func (r *PostgresKeyManagementRepository) tableName() string {
	return r.tablePrefix + "_crypto_keys"
}

// StoreKey securely stores a cryptographic key with its metadata.
func (r *PostgresKeyManagementRepository) StoreKey(ctx context.Context, key *models.CryptoKey) error {
	if key.ID == "" {
		key.ID = uuid.New().String()
	}

	key.CreatedAt = time.Now().UTC()
	key.Version = 1

	// Serialize key metadata for storage.
	metadataJSON, err := json.Marshal(key.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize key metadata: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, name, purpose, algorithm, key_type, key_material,
			public_key, version, status, created_at, last_used_at,
			expires_at, metadata, hsm_key_id, rotation_policy
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			purpose = EXCLUDED.purpose,
			algorithm = EXCLUDED.algorithm,
			key_type = EXCLUDED.key_type,
			key_material = EXCLUDED.key_material,
			public_key = EXCLUDED.public_key,
			status = EXCLUDED.status,
			last_used_at = EXCLUDED.last_used_at,
			expires_at = EXCLUDED.expires_at,
			metadata = EXCLUDED.metadata,
			hsm_key_id = EXCLUDED.hsm_key_id,
			rotation_policy = EXCLUDED.rotation_policy
	`, r.tableName())

	_, err = r.db.ExecContext(ctx, query,
		key.ID,
		key.Name,
		key.Purpose,
		key.Algorithm,
		key.KeyType,
		key.KeyMaterial,
		key.PublicKey,
		key.Version,
		key.Status,
		key.CreatedAt,
		key.LastUsedAt,
		key.ExpiresAt,
		string(metadataJSON),
		key.HSMKeyID,
		key.RotationPolicy,
	)

	if err != nil {
		return fmt.Errorf("failed to store key: %w", err)
	}

	return nil
}

// RetrieveKey retrieves a cryptographic key by its identifier.
func (r *PostgresKeyManagementRepository) RetrieveKey(ctx context.Context, keyID string) (*models.CryptoKey, error) {
	query := fmt.Sprintf(`
		SELECT id, name, purpose, algorithm, key_type, key_material,
			   public_key, version, status, created_at, last_used_at,
			   expires_at, metadata, hsm_key_id, rotation_policy
		FROM %s WHERE id = $1
	`, r.tableName())

	key := &models.CryptoKey{}
	var metadataJSON, rotationPolicy string

	err := r.db.QueryRowContext(ctx, query, keyID).Scan(
		&key.ID,
		&key.Name,
		&key.Purpose,
		&key.Algorithm,
		&key.KeyType,
		&key.KeyMaterial,
		&key.PublicKey,
		&key.Version,
		&key.Status,
		&key.CreatedAt,
		&key.LastUsedAt,
		&key.ExpiresAt,
		&metadataJSON,
		&key.HSMKeyID,
		&rotationPolicy,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve key: %w", err)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &key.Metadata); err != nil {
		key.Metadata = make(map[string]string)
	}
	key.RotationPolicy = rotationPolicy

	return key, nil
}

// GetKeyVersion retrieves a specific version of a key.
func (r *PostgresKeyManagementRepository) GetKeyVersion(ctx context.Context, keyID string, version int) (*models.CryptoKey, error) {
	query := fmt.Sprintf(`
		SELECT id, name, purpose, algorithm, key_type, key_material,
			   public_key, version, status, created_at, last_used_at,
			   expires_at, metadata, hsm_key_id, rotation_policy
		FROM %s WHERE id = $1 AND version = $2
	`, r.tableName())

	key := &models.CryptoKey{}
	var metadataJSON, rotationPolicy string

	err := r.db.QueryRowContext(ctx, query, keyID, version).Scan(
		&key.ID,
		&key.Name,
		&key.Purpose,
		&key.Algorithm,
		&key.KeyType,
		&key.KeyMaterial,
		&key.PublicKey,
		&key.Version,
		&key.Status,
		&key.CreatedAt,
		&key.LastUsedAt,
		&key.ExpiresAt,
		&metadataJSON,
		&key.HSMKeyID,
		&rotationPolicy,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve key version: %w", err)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &key.Metadata); err != nil {
		key.Metadata = make(map[string]string)
	}
	key.RotationPolicy = rotationPolicy

	return key, nil
}

// ListKeys lists all active keys for a given purpose.
func (r *PostgresKeyManagementRepository) ListKeys(ctx context.Context, purpose models.KeyPurpose) ([]*models.CryptoKey, error) {
	query := fmt.Sprintf(`
		SELECT id, name, purpose, algorithm, key_type, key_material,
			   public_key, version, status, created_at, last_used_at,
			   expires_at, metadata, hsm_key_id, rotation_policy
		FROM %s
		WHERE purpose = $1 AND status = $2
		ORDER BY created_at DESC
	`, r.tableName())

	rows, err := r.db.QueryContext(ctx, query, purpose, models.KeyStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	defer rows.Close()

	return r.scanKeys(rows)
}

// MarkKeyRevoked marks a key as revoked and optionally archives it.
func (r *PostgresKeyManagementRepository) MarkKeyRevoked(ctx context.Context, keyID string, reason string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $1, metadata = metadata || $2::jsonb
		WHERE id = $3
	`, r.tableName())

	metadataUpdate := fmt.Sprintf(`{"revoked_at": "%s", "revocation_reason": "%s"}`,
		time.Now().UTC().Format(time.RFC3339), reason)

	_, err := r.db.ExecContext(ctx, query, models.KeyStatusRevoked, metadataUpdate, keyID)
	if err != nil {
		return fmt.Errorf("failed to revoke key: %w", err)
	}

	return nil
}

// RotateKey creates a new version of an existing key for rotation purposes.
func (r *PostgresKeyManagementRepository) RotateKey(ctx context.Context, keyID string) (*models.CryptoKey, error) {
	// Get the current key.
	currentKey, err := r.RetrieveKey(ctx, keyID)
	if err != nil {
		return nil, err
	}
	if currentKey == nil {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	// Mark the current key as superseded.
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $1
		WHERE id = $2
	`, r.tableName())

	_, err = r.db.ExecContext(ctx, query, models.KeyStatusSuperseded, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to mark key as superseded: %w", err)
	}

	// Create a new version of the key.
	newKey := &models.CryptoKey{
		Name:        currentKey.Name,
		Purpose:     currentKey.Purpose,
		Algorithm:   currentKey.Algorithm,
		KeyType:     currentKey.KeyType,
		Status:      models.KeyStatusActive,
		ExpiresAt:   currentKey.ExpiresAt,
		RotationPolicy: currentKey.RotationPolicy,
		Metadata:    make(map[string]string),
	}

	// Preserve original metadata and add rotation info.
	for k, v := range currentKey.Metadata {
		newKey.Metadata[k] = v
	}
	newKey.Metadata["rotated_from"] = keyID
	newKey.Metadata["rotated_at"] = time.Now().UTC().Format(time.RFC3339)

	// Store the new key.
	if err := r.StoreKey(ctx, newKey); err != nil {
		return nil, err
	}

	return newKey, nil
}

// GetActiveKey retrieves the active key for a specific purpose.
func (r *PostgresKeyManagementRepository) GetActiveKey(ctx context.Context, purpose models.KeyPurpose) (*models.CryptoKey, error) {
	query := fmt.Sprintf(`
		SELECT id, name, purpose, algorithm, key_type, key_material,
			   public_key, version, status, created_at, last_used_at,
			   expires_at, metadata, hsm_key_id, rotation_policy
		FROM %s
		WHERE purpose = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, r.tableName())

	key := &models.CryptoKey{}
	var metadataJSON, rotationPolicy string

	err := r.db.QueryRowContext(ctx, query, purpose, models.KeyStatusActive).Scan(
		&key.ID,
		&key.Name,
		&key.Purpose,
		&key.Algorithm,
		&key.KeyType,
		&key.KeyMaterial,
		&key.PublicKey,
		&key.Version,
		&key.Status,
		&key.CreatedAt,
		&key.LastUsedAt,
		&key.ExpiresAt,
		&metadataJSON,
		&key.HSMKeyID,
		&rotationPolicy,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active key: %w", err)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &key.Metadata); err != nil {
		key.Metadata = make(map[string]string)
	}
	key.RotationPolicy = rotationPolicy

	return key, nil
}

// scanKeys scans database rows into CryptoKey structs.
func (r *PostgresKeyManagementRepository) scanKeys(rows *sql.Rows) ([]*models.CryptoKey, error) {
	var keys []*models.CryptoKey

	for rows.Next() {
		key := &models.CryptoKey{}
		var metadataJSON, rotationPolicy string

		err := rows.Scan(
			&key.ID,
			&key.Name,
			&key.Purpose,
			&key.Algorithm,
			&key.KeyType,
			&key.KeyMaterial,
			&key.PublicKey,
			&key.Version,
			&key.Status,
			&key.CreatedAt,
			&key.LastUsedAt,
			&key.ExpiresAt,
			&metadataJSON,
			&key.HSMKeyID,
			&rotationPolicy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan key: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataJSON), &key.Metadata); err != nil {
			key.Metadata = make(map[string]string)
		}
		key.RotationPolicy = rotationPolicy
		keys = append(keys, key)
	}

	return keys, rows.Err()
}
