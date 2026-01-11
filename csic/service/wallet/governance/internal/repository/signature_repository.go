package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/wallet-governance/internal/config"
	"github.com/csic/wallet-governance/internal/domain/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresSignatureRepository handles signature request data access
type PostgresSignatureRepository struct {
	db *sql.DB
}

// NewPostgresSignatureRepository creates a new signature repository
func NewPostgresSignatureRepository(cfg config.DatabaseConfig) (*PostgresSignatureRepository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresSignatureRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresSignatureRepository) Close() error {
	return r.db.Close()
}

// Create creates a new signature request
func (r *PostgresSignatureRepository) Create(ctx context.Context, req *models.SignatureRequest) error {
	query := `
		INSERT INTO signature_requests (
			id, request_id, wallet_id, signer_id, message_hash, message,
			signature_type, status, signature, public_key, expires_at,
			retry_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	req.ID = uuid.New()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		req.ID, req.RequestID, req.WalletID, req.SignerID, req.MessageHash, req.Message,
		req.SignatureType, req.Status, req.Signature, req.PublicKey, req.ExpiresAt,
		req.RetryCount, req.CreatedAt, req.UpdatedAt,
	)

	return err
}

// GetByID retrieves a signature request by ID
func (r *PostgresSignatureRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SignatureRequest, error) {
	query := `
		SELECT id, request_id, wallet_id, signer_id, message_hash, message,
			signature_type, status, signature, public_key, expires_at,
			completed_at, failure_reason, retry_count, created_at, updated_at
		FROM signature_requests WHERE id = $1
	`

	var req models.SignatureRequest
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID, &req.RequestID, &req.WalletID, &req.SignerID, &req.MessageHash, &req.Message,
		&req.SignatureType, &req.Status, &req.Signature, &req.PublicKey, &req.ExpiresAt,
		&completedAt, &req.FailureReason, &req.RetryCount, &req.CreatedAt, &req.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if completedAt.Valid {
		req.CompletedAt = &completedAt.Time
	}

	return &req, nil
}

// Update updates a signature request
func (r *PostgresSignatureRepository) Update(ctx context.Context, req *models.SignatureRequest) error {
	query := `
		UPDATE signature_requests SET
			status = $1, signature = $2, completed_at = $3, failure_reason = $4,
			retry_count = $5, updated_at = $6
		WHERE id = $7
	`

	req.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		req.Status, req.Signature, req.CompletedAt, req.FailureReason,
		req.RetryCount, req.UpdatedAt, req.ID,
	)

	return err
}

// GetPendingByWallet retrieves pending signature requests for a wallet
func (r *PostgresSignatureRepository) GetPendingByWallet(ctx context.Context, walletID uuid.UUID) ([]*models.SignatureRequest, error) {
	query := `
		SELECT id, request_id, wallet_id, signer_id, message_hash, message,
			signature_type, status, signature, public_key, expires_at,
			completed_at, failure_reason, retry_count, created_at, updated_at
		FROM signature_requests
		WHERE wallet_id = $1 AND status = 'PENDING'
		AND expires_at > NOW()
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []*models.SignatureRequest
	for rows.Next() {
		var req models.SignatureRequest
		var completedAt sql.NullTime

		err := rows.Scan(
			&req.ID, &req.RequestID, &req.WalletID, &req.SignerID, &req.MessageHash, &req.Message,
			&req.SignatureType, &req.Status, &req.Signature, &req.PublicKey, &req.ExpiresAt,
			&completedAt, &req.FailureReason, &req.RetryCount, &req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if completedAt.Valid {
			req.CompletedAt = &completedAt.Time
		}

		reqs = append(reqs, &req)
	}

	return reqs, rows.Err()
}

// GetExpired retrieves expired signature requests
func (r *PostgresSignatureRepository) GetExpired(ctx context.Context) ([]*models.SignatureRequest, error) {
	query := `
		SELECT id, request_id, wallet_id, signer_id, message_hash, message,
			signature_type, status, signature, public_key, expires_at,
			completed_at, failure_reason, retry_count, created_at, updated_at
		FROM signature_requests
		WHERE status = 'PENDING'
		AND expires_at <= NOW()
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []*models.SignatureRequest
	for rows.Next() {
		var req models.SignatureRequest
		var completedAt sql.NullTime

		err := rows.Scan(
			&req.ID, &req.RequestID, &req.WalletID, &req.SignerID, &req.MessageHash, &req.Message,
			&req.SignatureType, &req.Status, &req.Signature, &req.PublicKey, &req.ExpiresAt,
			&completedAt, &req.FailureReason, &req.RetryCount, &req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if completedAt.Valid {
			req.CompletedAt = &completedAt.Time
		}

		reqs = append(reqs, &req)
	}

	return reqs, rows.Err()
}
