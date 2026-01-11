package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/services/services/mining/internal/core/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Repository implements the MiningRepository interface using PostgreSQL
type Repository struct {
	conn   Conn
	log    *zap.Logger
}

// NewRepository creates a new Repository instance
func NewRepository(conn Conn, log *zap.Logger) *Repository {
	return &Repository{
		conn: conn,
		log:  log,
	}
}

// Helper function to scan a row into a MiningOperation
func scanOperation(row RowScanner) (*domain.MiningOperation, error) {
	op := &domain.MiningOperation{}
	err := row.Scan(
		&op.ID,
		&op.OperatorName,
		&op.WalletAddress,
		&op.CurrentHashrate,
		&op.Status,
		&op.QuotaID,
		&op.Location,
		&op.Region,
		&op.MachineType,
		&op.RegisteredAt,
		&op.UpdatedAt,
		&op.LastReportAt,
		&op.ViolationCount,
		&op.Metadata,
	)
	if err != nil {
		return nil, err
	}
	return op, nil
}

// Helper function to scan a row into a HashrateRecord
func scanHashrateRecord(row RowScanner) (*domain.HashrateRecord, error) {
	record := &domain.HashrateRecord{}
	err := row.Scan(
		&record.ID,
		&record.OperationID,
		&record.Hashrate,
		&record.Unit,
		&record.BlockHeight,
		&record.Timestamp,
		&record.SubmittedAt,
	)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// Helper function to scan a row into a MiningQuota
func scanQuota(row RowScanner) (*domain.MiningQuota, error) {
	quota := &domain.MiningQuota{}
	err := row.Scan(
		&quota.ID,
		&quota.OperationID,
		&quota.MaxHashrate,
		&quota.QuotaType,
		&quota.ValidFrom,
		&quota.ValidTo,
		&quota.Region,
		&quota.Priority,
		&quota.CreatedAt,
		&quota.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return quota, nil
}

// Helper function to scan a row into a ShutdownCommand
func scanShutdownCommand(row RowScanner) (*domain.ShutdownCommand, error) {
	cmd := &domain.ShutdownCommand{}
	err := row.Scan(
		&cmd.ID,
		&cmd.OperationID,
		&cmd.CommandType,
		&cmd.Reason,
		&cmd.Status,
		&cmd.IssuedAt,
		&cmd.ExecutedAt,
		&cmd.AckedAt,
		&cmd.IssuedBy,
		&cmd.ExpiresAt,
		&cmd.Result,
	)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

// CreateOperation creates a new mining operation
func (r *Repository) CreateOperation(ctx context.Context, op *domain.MiningOperation) error {
	query := `
		INSERT INTO mining_operations (
			id, operator_name, wallet_address, current_hashrate, status,
			quota_id, location, region, machine_type, registered_at,
			updated_at, last_report_at, violation_count, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`
	_, err := r.conn.Exec(ctx, query,
		op.ID,
		op.OperatorName,
		op.WalletAddress,
		op.CurrentHashrate,
		op.Status,
		op.QuotaID,
		op.Location,
		op.Region,
		op.MachineType,
		op.RegisteredAt,
		op.UpdatedAt,
		op.LastReportAt,
		op.ViolationCount,
		op.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create operation: %w", err)
	}
	return nil
}

// GetOperation retrieves a mining operation by ID
func (r *Repository) GetOperation(ctx context.Context, id uuid.UUID) (*domain.MiningOperation, error) {
	query := `
		SELECT id, operator_name, wallet_address, current_hashrate, status,
			   quota_id, location, region, machine_type, registered_at,
			   updated_at, last_report_at, violation_count, metadata
		FROM mining_operations
		WHERE id = $1
	`
	row := r.conn.QueryRow(ctx, query, id)
	return scanOperation(row)
}

// GetOperationByWallet retrieves a mining operation by wallet address
func (r *Repository) GetOperationByWallet(ctx context.Context, walletAddress string) (*domain.MiningOperation, error) {
	query := `
		SELECT id, operator_name, wallet_address, current_hashrate, status,
			   quota_id, location, region, machine_type, registered_at,
			   updated_at, last_report_at, violation_count, metadata
		FROM mining_operations
		WHERE wallet_address = $1
	`
	row := r.conn.QueryRow(ctx, query, walletAddress)
	return scanOperation(row)
}

// UpdateOperationStatus updates the status of a mining operation
func (r *Repository) UpdateOperationStatus(ctx context.Context, id uuid.UUID, status domain.OperationStatus) error {
	query := `
		UPDATE mining_operations
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	result, err := r.conn.Exec(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update operation status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("operation not found: %s", id.String())
	}
	return nil
}

// UpdateOperationHashrate updates the hashrate and last report time
func (r *Repository) UpdateOperationHashrate(ctx context.Context, id uuid.UUID, hashrate float64, timestamp time.Time) error {
	query := `
		UPDATE mining_operations
		SET current_hashrate = $1, last_report_at = $2, updated_at = $3
		WHERE id = $4
	`
	_, err := r.conn.Exec(ctx, query, hashrate, timestamp, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update operation hashrate: %w", err)
	}
	return nil
}

// ListOperations retrieves all operations with optional status filter
func (r *Repository) ListOperations(ctx context.Context, status *domain.OperationStatus, limit, offset int) ([]domain.MiningOperation, error) {
	var query string
	var args []interface{}

	if status != nil {
		query = `
			SELECT id, operator_name, wallet_address, current_hashrate, status,
				   quota_id, location, region, machine_type, registered_at,
				   updated_at, last_report_at, violation_count, metadata
			FROM mining_operations
			WHERE status = $1
			ORDER BY registered_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{*status, limit, offset}
	} else {
		query = `
			SELECT id, operator_name, wallet_address, current_hashrate, status,
				   quota_id, location, region, machine_type, registered_at,
				   updated_at, last_report_at, violation_count, metadata
			FROM mining_operations
			ORDER BY registered_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list operations: %w", err)
	}
	defer rows.Close()

	var operations []domain.MiningOperation
	for rows.Next() {
		op, err := scanOperation(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan operation: %w", err)
		}
		operations = append(operations, *op)
	}

	return operations, nil
}

// CountOperations counts operations with optional status filter
func (r *Repository) CountOperations(ctx context.Context, status *domain.OperationStatus) (int64, error) {
	var query string
	var args []interface{}

	if status != nil {
		query = "SELECT COUNT(*) FROM mining_operations WHERE status = $1"
		args = []interface{}{*status}
	} else {
		query = "SELECT COUNT(*) FROM mining_operations"
	}

	var count int64
	err := r.conn.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count operations: %w", err)
	}
	return count, nil
}

// RecordHashrate records a hashrate telemetry record
func (r *Repository) RecordHashrate(ctx context.Context, record *domain.HashrateRecord) error {
	query := `
		INSERT INTO hashrate_records (
			operation_id, hashrate, unit, block_height, timestamp, submitted_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.conn.Exec(ctx, query,
		record.OperationID,
		record.Hashrate,
		record.Unit,
		record.BlockHeight,
		record.Timestamp,
		record.SubmittedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to record hashrate: %w", err)
	}
	return nil
}

// GetHashrateHistory retrieves historical hashrate records
func (r *Repository) GetHashrateHistory(ctx context.Context, opID uuid.UUID, limit int) ([]domain.HashrateRecord, error) {
	query := `
		SELECT id, operation_id, hashrate, unit, block_height, timestamp, submitted_at
		FROM hashrate_records
		WHERE operation_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`
	rows, err := r.conn.Query(ctx, query, opID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get hashrate history: %w", err)
	}
	defer rows.Close()

	var records []domain.HashrateRecord
	for rows.Next() {
		record, err := scanHashrateRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan hashrate record: %w", err)
		}
		records = append(records, *record)
	}

	return records, nil
}

// GetLatestHashrate retrieves the latest hashrate record
func (r *Repository) GetLatestHashrate(ctx context.Context, opID uuid.UUID) (*domain.HashrateRecord, error) {
	query := `
		SELECT id, operation_id, hashrate, unit, block_height, timestamp, submitted_at
		FROM hashrate_records
		WHERE operation_id = $1
		ORDER BY submitted_at DESC
		LIMIT 1
	`
	row := r.conn.QueryRow(ctx, query, opID)
	return scanHashrateRecord(row)
}

// CountHashrateRecords counts hashrate records for an operation
func (r *Repository) CountHashrateRecords(ctx context.Context, opID uuid.UUID) (int64, error) {
	query := "SELECT COUNT(*) FROM hashrate_records WHERE operation_id = $1"
	var count int64
	err := r.conn.QueryRow(ctx, query, opID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count hashrate records: %w", err)
	}
	return count, nil
}

// SetQuota creates or updates a quota
func (r *Repository) SetQuota(ctx context.Context, quota *domain.MiningQuota) error {
	query := `
		INSERT INTO mining_quotas (
			id, operation_id, max_hashrate, quota_type, valid_from, valid_to,
			region, priority, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		ON CONFLICT (id) DO UPDATE SET
			max_hashrate = EXCLUDED.max_hashrate,
			quota_type = EXCLUDED.quota_type,
			valid_from = EXCLUDED.valid_from,
			valid_to = EXCLUDED.valid_to,
			region = EXCLUDED.region,
			priority = EXCLUDED.priority,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.conn.Exec(ctx, query,
		quota.ID,
		quota.OperationID,
		quota.MaxHashrate,
		quota.QuotaType,
		quota.ValidFrom,
		quota.ValidTo,
		quota.Region,
		quota.Priority,
		quota.CreatedAt,
		quota.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to set quota: %w", err)
	}
	return nil
}

// GetCurrentQuota retrieves the current active quota for an operation
func (r *Repository) GetCurrentQuota(ctx context.Context, opID uuid.UUID) (*domain.MiningQuota, error) {
	query := `
		SELECT id, operation_id, max_hashrate, quota_type, valid_from, valid_to,
			   region, priority, created_at, updated_at
		FROM mining_quotas
		WHERE operation_id = $1
		  AND (valid_to IS NULL OR valid_to > $2)
		  AND valid_from <= $2
		ORDER BY priority DESC, created_at DESC
		LIMIT 1
	`
	row := r.conn.QueryRow(ctx, query, opID, time.Now().UTC())
	return scanQuota(row)
}

// GetQuotas retrieves all quotas for an operation
func (r *Repository) GetQuotas(ctx context.Context, opID uuid.UUID) ([]domain.MiningQuota, error) {
	query := `
		SELECT id, operation_id, max_hashrate, quota_type, valid_from, valid_to,
			   region, priority, created_at, updated_at
		FROM mining_quotas
		WHERE operation_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.conn.Query(ctx, query, opID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quotas: %w", err)
	}
	defer rows.Close()

	var quotas []domain.MiningQuota
	for rows.Next() {
		quota, err := scanQuota(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quota: %w", err)
		}
		quotas = append(quotas, *quota)
	}

	return quotas, nil
}

// DeleteQuota deletes a quota by ID
func (r *Repository) DeleteQuota(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM mining_quotas WHERE id = $1"
	result, err := r.conn.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete quota: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("quota not found: %s", id.String())
	}
	return nil
}

// CreateShutdownCommand creates a new shutdown command
func (r *Repository) CreateShutdownCommand(ctx context.Context, cmd *domain.ShutdownCommand) error {
	query := `
		INSERT INTO shutdown_commands (
			id, operation_id, command_type, reason, status, issued_at,
			executed_at, acked_at, issued_by, expires_at, result
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.conn.Exec(ctx, query,
		cmd.ID,
		cmd.OperationID,
		cmd.CommandType,
		cmd.Reason,
		cmd.Status,
		cmd.IssuedAt,
		cmd.ExecutedAt,
		cmd.AckedAt,
		cmd.IssuedBy,
		cmd.ExpiresAt,
		cmd.Result,
	)
	if err != nil {
		return fmt.Errorf("failed to create shutdown command: %w", err)
	}
	return nil
}

// GetShutdownCommand retrieves a shutdown command by ID
func (r *Repository) GetShutdownCommand(ctx context.Context, id uuid.UUID) (*domain.ShutdownCommand, error) {
	query := `
		SELECT id, operation_id, command_type, reason, status, issued_at,
			   executed_at, acked_at, issued_by, expires_at, result
		FROM shutdown_commands
		WHERE id = $1
	`
	row := r.conn.QueryRow(ctx, query, id)
	return scanShutdownCommand(row)
}

// GetPendingCommands retrieves pending shutdown commands for an operation
func (r *Repository) GetPendingCommands(ctx context.Context, opID uuid.UUID) ([]domain.ShutdownCommand, error) {
	query := `
		SELECT id, operation_id, command_type, reason, status, issued_at,
			   executed_at, acked_at, issued_by, expires_at, result
		FROM shutdown_commands
		WHERE operation_id = $1 AND status IN ('ISSUED', 'ACKNOWLEDGED')
		ORDER BY issued_at DESC
	`
	rows, err := r.conn.Query(ctx, query, opID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending commands: %w", err)
	}
	defer rows.Close()

	var commands []domain.ShutdownCommand
	for rows.Next() {
		cmd, err := scanShutdownCommand(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shutdown command: %w", err)
		}
		commands = append(commands, *cmd)
	}

	return commands, nil
}

// GetAllPendingCommands retrieves all pending shutdown commands
func (r *Repository) GetAllPendingCommands(ctx context.Context) ([]domain.ShutdownCommand, error) {
	query := `
		SELECT id, operation_id, command_type, reason, status, issued_at,
			   executed_at, acked_at, issued_by, expires_at, result
		FROM shutdown_commands
		WHERE status IN ('ISSUED', 'ACKNOWLEDGED')
		ORDER BY issued_at DESC
	`
	rows, err := r.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all pending commands: %w", err)
	}
	defer rows.Close()

	var commands []domain.ShutdownCommand
	for rows.Next() {
		cmd, err := scanShutdownCommand(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shutdown command: %w", err)
		}
		commands = append(commands, *cmd)
	}

	return commands, nil
}

// UpdateCommandStatus updates the status of a shutdown command
func (r *Repository) UpdateCommandStatus(ctx context.Context, id uuid.UUID, status domain.CommandStatus, executedAt *time.Time) error {
	var query string
	var args []interface{}

	if executedAt != nil {
		query = `
			UPDATE shutdown_commands
			SET status = $1, executed_at = $2
			WHERE id = $3
		`
		args = []interface{}{status, *executedAt, id}
	} else {
		query = `
			UPDATE shutdown_commands
			SET status = $1, acked_at = $2
			WHERE id = $3
		`
		args = []interface{}{status, time.Now().UTC(), id}
	}

	result, err := r.conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update command status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("command not found: %s", id.String())
	}
	return nil
}

// RecordViolation records a quota violation
func (r *Repository) RecordViolation(ctx context.Context, violation *domain.QuotaViolation) error {
	query := `
		INSERT INTO quota_violations (
			id, operation_id, quota_id, reported_hashrate, max_hashrate,
			recorded_at, consecutive, automatic_shutdown
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.conn.Exec(ctx, query,
		violation.ID,
		violation.OperationID,
		violation.QuotaID,
		violation.ReportedHashrate,
		violation.MaxHashrate,
		violation.RecordedAt,
		violation.Consecutive,
		violation.AutomaticShutdown,
	)
	if err != nil {
		return fmt.Errorf("failed to record violation: %w", err)
	}
	return nil
}

// GetViolations retrieves quota violations for an operation
func (r *Repository) GetViolations(ctx context.Context, opID uuid.UUID, limit int) ([]domain.QuotaViolation, error) {
	query := `
		SELECT id, operation_id, quota_id, reported_hashrate, max_hashrate,
			   recorded_at, consecutive, automatic_shutdown
		FROM quota_violations
		WHERE operation_id = $1
		ORDER BY recorded_at DESC
		LIMIT $2
	`
	rows, err := r.conn.Query(ctx, query, opID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get violations: %w", err)
	}
	defer rows.Close()

	var violations []domain.QuotaViolation
	for rows.Next() {
		v := &domain.QuotaViolation{}
		err := rows.Scan(
			&v.ID,
			&v.OperationID,
			&v.QuotaID,
			&v.ReportedHashrate,
			&v.MaxHashrate,
			&v.RecordedAt,
			&v.Consecutive,
			&v.AutomaticShutdown,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan violation: %w", err)
		}
		violations = append(violations, *v)
	}

	return violations, nil
}

// GetRegistryStats retrieves registry statistics
func (r *Repository) GetRegistryStats(ctx context.Context) (*domain.RegistryStats, error) {
	stats := &domain.RegistryStats{}

	// Total operations
	total, err := r.CountOperations(ctx, nil)
	if err != nil {
		return nil, err
	}
	stats.TotalOperations = total

	// Active operations
	active, err := r.CountOperations(ctx, ptr(domain.StatusActive))
	if err != nil {
		return nil, err
	}
	stats.ActiveOperations = active

	// Suspended operations
	suspended, err := r.CountOperations(ctx, ptr(domain.StatusSuspended))
	if err != nil {
		return nil, err
	}
	stats.SuspendedOperations = suspended

	// Shutdown operations
	shutdown, err := r.CountOperations(ctx, ptr(domain.StatusShutdownExecuted))
	if err != nil {
		return nil, err
	}
	stats.ShutdownOperations = shutdown

	// Total hashrate
	query := "SELECT COALESCE(SUM(current_hashrate), 0) FROM mining_operations WHERE status = 'ACTIVE'"
	err = r.conn.QueryRow(ctx, query).Scan(&stats.TotalHashrate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total hashrate: %w", err)
	}

	// Quota violations count
	query = "SELECT COUNT(*) FROM quota_violations"
	err = r.conn.QueryRow(ctx, query).Scan(&stats.QuotaViolations)
	if err != nil {
		return nil, fmt.Errorf("failed to count violations: %w", err)
	}

	// Pending commands count
	query = "SELECT COUNT(*) FROM shutdown_commands WHERE status IN ('ISSUED', 'ACKNOWLEDGED')"
	err = r.conn.QueryRow(ctx, query).Scan(&stats.PendingCommands)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending commands: %w", err)
	}

	return stats, nil
}

// Helper function to get pointer to value
func ptr[T any](v T) *T {
	return &v
}
