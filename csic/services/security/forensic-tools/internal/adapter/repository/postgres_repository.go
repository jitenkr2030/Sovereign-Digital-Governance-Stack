package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic-platform/services/security/forensic-tools/internal/config"
	"github.com/csic-platform/services/security/forensic-tools/internal/core/domain"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresRepository implements the EvidenceRepository interface using PostgreSQL
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

// CreateEvidence creates new evidence
func (r *PostgresRepository) CreateEvidence(ctx context.Context, evidence *domain.Evidence) error {
	metadata, _ := json.Marshal(evidence.Metadata)

	query := `
		INSERT INTO evidence (id, case_id, file_name, file_hash, file_type, size_bytes,
		                    storage_path, evidence_type, status, metadata, uploaded_by,
		                    uploaded_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		evidence.ID, evidence.CaseID, evidence.FileName, evidence.FileHash,
		evidence.FileType, evidence.SizeBytes, evidence.StoragePath, evidence.EvidenceType,
		evidence.Status, metadata, evidence.UploadedBy, evidence.UploadedAt, evidence.UpdatedAt,
	)
	return err
}

// GetEvidence retrieves evidence by ID
func (r *PostgresRepository) GetEvidence(ctx context.Context, id string) (*domain.Evidence, error) {
	query := `
		SELECT id, case_id, file_name, file_hash, file_type, size_bytes, storage_path,
		       evidence_type, status, metadata, uploaded_by, uploaded_at, updated_at,
		       archived_at, deleted_at
		FROM evidence WHERE id = $1 AND deleted_at IS NULL
	`

	var evidence domain.Evidence
	var metadata []byte
	var archivedAt, deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&evidence.ID, &evidence.CaseID, &evidence.FileName, &evidence.FileHash,
		&evidence.FileType, &evidence.SizeBytes, &evidence.StoragePath, &evidence.EvidenceType,
		&evidence.Status, &metadata, &evidence.UploadedBy, &evidence.UploadedAt, &evidence.UpdatedAt,
		&archivedAt, &deletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrEvidenceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get evidence: %w", err)
	}

	json.Unmarshal(metadata, &evidence.Metadata)

	if archivedAt.Valid {
		evidence.ArchivedAt = &archivedAt.Time
	}
	if deletedAt.Valid {
		evidence.DeletedAt = &deletedAt.Time
	}

	return &evidence, nil
}

// GetEvidenceByHash retrieves evidence by file hash
func (r *PostgresRepository) GetEvidenceByHash(ctx context.Context, hash string) (*domain.Evidence, error) {
	query := `
		SELECT id, case_id, file_name, file_hash, file_type, size_bytes, storage_path,
		       evidence_type, status, metadata, uploaded_by, uploaded_at, updated_at,
		       archived_at, deleted_at
		FROM evidence WHERE file_hash = $1 AND deleted_at IS NULL
	`

	var evidence domain.Evidence
	var metadata []byte
	var archivedAt, deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&evidence.ID, &evidence.CaseID, &evidence.FileName, &evidence.FileHash,
		&evidence.FileType, &evidence.SizeBytes, &evidence.StoragePath, &evidence.EvidenceType,
		&evidence.Status, &metadata, &evidence.UploadedBy, &evidence.UploadedAt, &evidence.UpdatedAt,
		&archivedAt, &deletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get evidence by hash: %w", err)
	}

	json.Unmarshal(metadata, &evidence.Metadata)

	if archivedAt.Valid {
		evidence.ArchivedAt = &archivedAt.Time
	}
	if deletedAt.Valid {
		evidence.DeletedAt = &deletedAt.Time
	}

	return &evidence, nil
}

// ListEvidence retrieves a paginated list of evidence
func (r *PostgresRepository) ListEvidence(ctx context.Context, caseID string, page, pageSize int) ([]*domain.Evidence, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	countQuery := "SELECT COUNT(*) FROM evidence WHERE deleted_at IS NULL"
	if caseID != "" {
		countQuery += " AND case_id = $1"
	}
	countArgs := []interface{}{}
	if caseID != "" {
		countArgs = append(countArgs, caseID)
	}
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count evidence: %w", err)
	}

	// Get evidence
	query := `
		SELECT id, case_id, file_name, file_hash, file_type, size_bytes, storage_path,
		       evidence_type, status, metadata, uploaded_by, uploaded_at, updated_at,
		       archived_at, deleted_at
		FROM evidence WHERE deleted_at IS NULL
	`
	args := []interface{}{}

	if caseID != "" {
		query += " AND case_id = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, caseID)
	}

	query += fmt.Sprintf(" ORDER BY uploaded_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list evidence: %w", err)
	}
	defer rows.Close()

	var evidenceList []*domain.Evidence
	for rows.Next() {
		var evidence domain.Evidence
		var metadata []byte
		var archivedAt, deletedAt sql.NullTime

		if err := rows.Scan(
			&evidence.ID, &evidence.CaseID, &evidence.FileName, &evidence.FileHash,
			&evidence.FileType, &evidence.SizeBytes, &evidence.StoragePath, &evidence.EvidenceType,
			&evidence.Status, &metadata, &evidence.UploadedBy, &evidence.UploadedAt, &evidence.UpdatedAt,
			&archivedAt, &deletedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan evidence: %w", err)
		}

		json.Unmarshal(metadata, &evidence.Metadata)

		if archivedAt.Valid {
			evidence.ArchivedAt = &archivedAt.Time
		}
		if deletedAt.Valid {
			evidence.DeletedAt = &deletedAt.Time
		}

		evidenceList = append(evidenceList, &evidence)
	}

	return evidenceList, total, nil
}

// UpdateEvidence updates evidence
func (r *PostgresRepository) UpdateEvidence(ctx context.Context, evidence *domain.Evidence) error {
	metadata, _ := json.Marshal(evidence.Metadata)

	query := `
		UPDATE evidence SET case_id=$1, file_name=$2, file_type=$3, status=$4,
		       metadata=$5, updated_at=$6, archived_at=$7, deleted_at=$8
		WHERE id=$9
	`

	_, err := r.db.ExecContext(ctx, query,
		evidence.CaseID, evidence.FileName, evidence.FileType, evidence.Status,
		metadata, evidence.UpdatedAt, evidence.ArchivedAt, evidence.DeletedAt, evidence.ID,
	)
	return err
}

// DeleteEvidence soft-deletes evidence
func (r *PostgresRepository) DeleteEvidence(ctx context.Context, id string) error {
	query := `UPDATE evidence SET deleted_at=$1 WHERE id=$2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

// AddCustodyEntry adds an entry to the chain of custody
func (r *PostgresRepository) AddCustodyEntry(ctx context.Context, entry *domain.ChainOfCustody) error {
	details, _ := json.Marshal(entry.Details)

	query := `
		INSERT INTO chain_of_custody (id, evidence_id, actor_id, action, details, ip_address, timestamp, signature)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.EvidenceID, entry.ActorID, entry.Action, details,
		entry.IPAddress, entry.Timestamp, entry.Signature,
	)
	return err
}

// GetCustodyHistory retrieves the chain of custody for evidence
func (r *PostgresRepository) GetCustodyHistory(ctx context.Context, evidenceID string) ([]*domain.ChainOfCustody, error) {
	query := `
		SELECT id, evidence_id, actor_id, action, details, ip_address, timestamp, signature
		FROM chain_of_custody WHERE evidence_id=$1 ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, evidenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get custody history: %w", err)
	}
	defer rows.Close()

	var entries []*domain.ChainOfCustody
	for rows.Next() {
		var entry domain.ChainOfCustody
		var details []byte

		if err := rows.Scan(
			&entry.ID, &entry.EvidenceID, &entry.ActorID, &entry.Action,
			&details, &entry.IPAddress, &entry.Timestamp, &entry.Signature,
		); err != nil {
			return nil, fmt.Errorf("failed to scan custody entry: %w", err)
		}

		json.Unmarshal(details, &entry.Details)
		entries = append(entries, &entry)
	}

	return entries, nil
}

// CreateAnalysisJob creates a new analysis job
func (r *PostgresRepository) CreateAnalysisJob(ctx context.Context, job *domain.AnalysisJob) error {
	query := `
		INSERT INTO analysis_jobs (id, evidence_id, tool_name, status, priority, requested_by, submitted_at, progress)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		job.ID, job.EvidenceID, job.ToolName, job.Status, job.Priority,
		job.RequestedBy, job.SubmittedAt, job.Progress,
	)
	return err
}

// GetAnalysisJob retrieves an analysis job by ID
func (r *PostgresRepository) GetAnalysisJob(ctx context.Context, id string) (*domain.AnalysisJob, error) {
	query := `
		SELECT id, evidence_id, tool_name, status, priority, requested_by, submitted_at,
		       started_at, completed_at, error_message, progress
		FROM analysis_jobs WHERE id = $1
	`

	var job domain.AnalysisJob
	var startedAt, completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID, &job.EvidenceID, &job.ToolName, &job.Status, &job.Priority,
		&job.RequestedBy, &job.SubmittedAt, &startedAt, &completedAt,
		&job.ErrorMessage, &job.Progress,
	)
	if err == sql.ErrNoRows {
		return nil, ErrAnalysisJobNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis job: %w", err)
	}

	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

// UpdateAnalysisJob updates an analysis job
func (r *PostgresRepository) UpdateAnalysisJob(ctx context.Context, job *domain.AnalysisJob) error {
	query := `
		UPDATE analysis_jobs SET status=$1, started_at=$2, completed_at=$3, error_message=$4, progress=$5
		WHERE id=$6
	`

	_, err := r.db.ExecContext(ctx, query,
		job.Status, job.StartedAt, job.CompletedAt, job.ErrorMessage, job.Progress, job.ID,
	)
	return err
}

// ListAnalysisJobs retrieves analysis jobs for evidence
func (r *PostgresRepository) ListAnalysisJobs(ctx context.Context, evidenceID string, page, pageSize int) ([]*domain.AnalysisJob, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	countQuery := "SELECT COUNT(*) FROM analysis_jobs WHERE evidence_id = $1"
	if err := r.db.QueryRowContext(ctx, countQuery, evidenceID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count analysis jobs: %w", err)
	}

	// Get jobs
	query := `
		SELECT id, evidence_id, tool_name, status, priority, requested_by, submitted_at,
		       started_at, completed_at, error_message, progress
		FROM analysis_jobs WHERE evidence_id=$1 ORDER BY submitted_at DESC LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, evidenceID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list analysis jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*domain.AnalysisJob
	for rows.Next() {
		var job domain.AnalysisJob
		var startedAt, completedAt sql.NullTime

		if err := rows.Scan(
			&job.ID, &job.EvidenceID, &job.ToolName, &job.Status, &job.Priority,
			&job.RequestedBy, &job.SubmittedAt, &startedAt, &completedAt,
			&job.ErrorMessage, &job.Progress,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan analysis job: %w", err)
		}

		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}

		jobs = append(jobs, &job)
	}

	return jobs, total, nil
}

// GetPendingJobs retrieves pending analysis jobs
func (r *PostgresRepository) GetPendingJobs(ctx context.Context, limit int) ([]*domain.AnalysisJob, error) {
	query := `
		SELECT id, evidence_id, tool_name, status, priority, requested_by, submitted_at,
		       started_at, completed_at, error_message, progress
		FROM analysis_jobs WHERE status IN ('PENDING', 'QUEUED')
		ORDER BY priority DESC, submitted_at ASC LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*domain.AnalysisJob
	for rows.Next() {
		var job domain.AnalysisJob
		var startedAt, completedAt sql.NullTime

		if err := rows.Scan(
			&job.ID, &job.EvidenceID, &job.ToolName, &job.Status, &job.Priority,
			&job.RequestedBy, &job.SubmittedAt, &startedAt, &completedAt,
			&job.ErrorMessage, &job.Progress,
		); err != nil {
			return nil, fmt.Errorf("failed to scan analysis job: %w", err)
		}

		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// CreateAnalysisResult creates an analysis result
func (r *PostgresRepository) CreateAnalysisResult(ctx context.Context, result *domain.AnalysisResult) error {
	resultData, _ := json.Marshal(result.ResultData)

	query := `
		INSERT INTO analysis_results (id, job_id, evidence_id, tool_name, result_type, result_data, summary, severity, tags, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		result.ID, result.JobID, result.EvidenceID, result.ToolName, result.ResultType,
		resultData, result.Summary, result.Severity, result.Tags, result.CreatedAt,
	)
	return err
}

// GetAnalysisResults retrieves results for an analysis job
func (r *PostgresRepository) GetAnalysisResults(ctx context.Context, jobID string) ([]*domain.AnalysisResult, error) {
	query := `
		SELECT id, job_id, evidence_id, tool_name, result_type, result_data, summary, severity, tags, created_at
		FROM analysis_results WHERE job_id=$1 ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis results: %w", err)
	}
	defer rows.Close()

	var results []*domain.AnalysisResult
	for rows.Next() {
		var result domain.AnalysisResult
		var resultData []byte

		if err := rows.Scan(
			&result.ID, &result.JobID, &result.EvidenceID, &result.ToolName,
			&result.ResultType, &resultData, &result.Summary, &result.Severity,
			&result.Tags, &result.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan analysis result: %w", err)
		}

		json.Unmarshal(resultData, &result.ResultData)
		results = append(results, &result)
	}

	return results, nil
}
