package postgres_repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic-platform/services/security/access-control/internal/core/domain"
	"github.com/csic-platform/services/security/access-control/internal/core/ports"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// PostgresPolicyRepository implements PolicyRepository for PostgreSQL
type PostgresPolicyRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgresPolicyRepository creates a new PostgreSQL policy repository
func NewPostgresPolicyRepository(db *sql.DB, logger *zap.Logger) *PostgresPolicyRepository {
	return &PostgresPolicyRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new policy
func (r *PostgresPolicyRepository) Create(ctx context.Context, policy *domain.Policy) error {
	conditionsJSON, err := json.Marshal(policy.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	query := `
		INSERT INTO access_control.policies (id, name, description, effect, priority, conditions, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.ExecContext(ctx, query,
		policy.ID, policy.Name, policy.Description, policy.Effect,
		policy.Priority, conditionsJSON, policy.Enabled,
		policy.CreatedAt, policy.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	return nil
}

// Update updates an existing policy
func (r *PostgresPolicyRepository) Update(ctx context.Context, policy *domain.Policy) error {
	conditionsJSON, err := json.Marshal(policy.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	query := `
		UPDATE access_control.policies
		SET name = $1, description = $2, effect = $3, priority = $4,
			conditions = $5, enabled = $6, updated_at = $7
		WHERE id = $8
	`

	_, err = r.db.ExecContext(ctx, query,
		policy.Name, policy.Description, policy.Effect,
		policy.Priority, conditionsJSON, policy.Enabled,
		time.Now(), policy.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	return nil
}

// Delete deletes a policy by ID
func (r *PostgresPolicyRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM access_control.policies WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	return nil
}

// FindByID retrieves a policy by ID
func (r *PostgresPolicyRepository) FindByID(ctx context.Context, id string) (*domain.Policy, error) {
	query := `
		SELECT id, name, description, effect, priority, conditions, enabled, created_at, updated_at
		FROM access_control.policies
		WHERE id = $1
	`

	var policy domain.Policy
	var conditionsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&policy.ID, &policy.Name, &policy.Description, &policy.Effect,
		&policy.Priority, &conditionsJSON, &policy.Enabled,
		&policy.CreatedAt, &policy.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	if err := json.Unmarshal(conditionsJSON, &policy.Conditions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
	}

	return &policy, nil
}

// FindAll retrieves all policies
func (r *PostgresPolicyRepository) FindAll(ctx context.Context) ([]*domain.Policy, error) {
	query := `
		SELECT id, name, description, effect, priority, conditions, enabled, created_at, updated_at
		FROM access_control.policies
		ORDER BY priority DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var policy domain.Policy
		var conditionsJSON []byte

		err := rows.Scan(
			&policy.ID, &policy.Name, &policy.Description, &policy.Effect,
			&policy.Priority, &conditionsJSON, &policy.Enabled,
			&policy.CreatedAt, &policy.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		if err := json.Unmarshal(conditionsJSON, &policy.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}

		policies = append(policies, &policy)
	}

	return policies, nil
}

// FindEnabled retrieves all enabled policies
func (r *PostgresPolicyRepository) FindEnabled(ctx context.Context) ([]*domain.Policy, error) {
	query := `
		SELECT id, name, description, effect, priority, conditions, enabled, created_at, updated_at
		FROM access_control.policies
		WHERE enabled = true
		ORDER BY priority DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled policies: %w", err)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var policy domain.Policy
		var conditionsJSON []byte

		err := rows.Scan(
			&policy.ID, &policy.Name, &policy.Description, &policy.Effect,
			&policy.Priority, &conditionsJSON, &policy.Enabled,
			&policy.CreatedAt, &policy.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		if err := json.Unmarshal(conditionsJSON, &policy.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}

		policies = append(policies, &policy)
	}

	return policies, nil
}

// FindApplicable retrieves policies applicable to a request (simplified implementation)
func (r *PostgresPolicyRepository) FindApplicable(ctx context.Context, req *domain.AccessRequest) ([]*domain.Policy, error) {
	// Get all enabled policies and filter in memory
	// In production, this would be optimized with proper indexing and query parameters
	return r.FindEnabled(ctx)
}

// Enable enables a policy
func (r *PostgresPolicyRepository) Enable(ctx context.Context, id string) error {
	query := `UPDATE access_control.policies SET enabled = true, updated_at = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to enable policy: %w", err)
	}

	return nil
}

// Disable disables a policy
func (r *PostgresPolicyRepository) Disable(ctx context.Context, id string) error {
	query := `UPDATE access_control.policies SET enabled = false, updated_at = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to disable policy: %w", err)
	}

	return nil
}

// Compile-time check that PostgresPolicyRepository implements PolicyRepository
var _ ports.PolicyRepository = (*PostgresPolicyRepository)(nil)

// PostgresAuditLogRepository implements AuditLogRepository for PostgreSQL
type PostgresAuditLogRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgresAuditLogRepository creates a new PostgreSQL audit log repository
func NewPostgresAuditLogRepository(db *sql.DB, logger *zap.Logger) *PostgresAuditLogRepository {
	return &PostgresAuditLogRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new audit log entry
func (r *PostgresAuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	requestJSON, err := json.Marshal(log.Request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	responseJSON, err := json.Marshal(log.Response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	query := `
		INSERT INTO access_control.audit_logs (id, request, response, decision, user_id, resource_id, action, ip_address, user_agent, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.ExecContext(ctx, query,
		uuid.New(), requestJSON, responseJSON, log.Decision,
		log.UserID, log.ResourceID, log.Action,
		log.IPAddress, log.UserAgent, log.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// FindByID retrieves an audit log by ID
func (r *PostgresAuditLogRepository) FindByID(ctx context.Context, id string) (*domain.AuditLog, error) {
	query := `
		SELECT id, request, response, decision, user_id, resource_id, action, ip_address, user_agent, timestamp
		FROM access_control.audit_logs
		WHERE id = $1
	`

	var log domain.AuditLog
	var requestJSON, responseJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID, &requestJSON, &responseJSON, &log.Decision,
		&log.UserID, &log.ResourceID, &log.Action,
		&log.IPAddress, &log.UserAgent, &log.Timestamp,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	if err := json.Unmarshal(requestJSON, &log.Request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	if err := json.Unmarshal(responseJSON, &log.Response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &log, nil
}

// FindByUserID retrieves audit logs for a user
func (r *PostgresAuditLogRepository) FindByUserID(ctx context.Context, userID string, limit int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, request, response, decision, user_id, resource_id, action, ip_address, user_agent, timestamp
		FROM access_control.audit_logs
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var requestJSON, responseJSON []byte

		err := rows.Scan(
			&log.ID, &requestJSON, &responseJSON, &log.Decision,
			&log.UserID, &log.ResourceID, &log.Action,
			&log.IPAddress, &log.UserAgent, &log.Timestamp,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if err := json.Unmarshal(requestJSON, &log.Request); err != nil {
			return nil, fmt.Errorf("failed to unmarshal request: %w", err)
		}

		if err := json.Unmarshal(responseJSON, &log.Response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		logs = append(logs, &log)
	}

	return logs, nil
}

// FindByResourceID retrieves audit logs for a resource
func (r *PostgresAuditLogRepository) FindByResourceID(ctx context.Context, resourceID string, limit int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, request, response, decision, user_id, resource_id, action, ip_address, user_agent, timestamp
		FROM access_control.audit_logs
		WHERE resource_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, resourceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var requestJSON, responseJSON []byte

		err := rows.Scan(
			&log.ID, &requestJSON, &responseJSON, &log.Decision,
			&log.UserID, &log.ResourceID, &log.Action,
			&log.IPAddress, &log.UserAgent, &log.Timestamp,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if err := json.Unmarshal(requestJSON, &log.Request); err != nil {
			return nil, fmt.Errorf("failed to unmarshal request: %w", err)
		}

		if err := json.Unmarshal(responseJSON, &log.Response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		logs = append(logs, &log)
	}

	return logs, nil
}

// FindByTimeRange retrieves audit logs within a time range
func (r *PostgresAuditLogRepository) FindByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, request, response, decision, user_id, resource_id, action, ip_address, user_agent, timestamp
		FROM access_control.audit_logs
		WHERE timestamp BETWEEN $1 AND $2
		ORDER BY timestamp DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var requestJSON, responseJSON []byte

		err := rows.Scan(
			&log.ID, &requestJSON, &responseJSON, &log.Decision,
			&log.UserID, &log.ResourceID, &log.Action,
			&log.IPAddress, &log.UserAgent, &log.Timestamp,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if err := json.Unmarshal(requestJSON, &log.Request); err != nil {
			return nil, fmt.Errorf("failed to unmarshal request: %w", err)
		}

		if err := json.Unmarshal(responseJSON, &log.Response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		logs = append(logs, &log)
	}

	return logs, nil
}

// Cleanup deletes audit logs older than the specified duration
func (r *PostgresAuditLogRepository) Cleanup(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	query := `DELETE FROM access_control.audit_logs WHERE timestamp < $1`

	result, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup audit logs: %w", err)
	}

	count, _ := result.RowsAffected()
	r.logger.Info("Cleaned up audit logs", zap.Int64("count", count))

	return nil
}

// Compile-time check that PostgresAuditLogRepository implements AuditLogRepository
var _ ports.AuditLogRepository = (*PostgresAuditLogRepository)(nil)

// PostgresResourceOwnershipRepository implements ResourceOwnershipRepository for PostgreSQL
type PostgresResourceOwnershipRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgresResourceOwnershipRepository creates a new PostgreSQL resource ownership repository
func NewPostgresResourceOwnershipRepository(db *sql.DB, logger *zap.Logger) *PostgresResourceOwnershipRepository {
	return &PostgresResourceOwnershipRepository{
		db:     db,
		logger: logger,
	}
}

// FindByResourceID retrieves ownership by resource ID
func (r *PostgresResourceOwnershipRepository) FindByResourceID(ctx context.Context, resourceID string) (*domain.ResourceOwnership, error) {
	query := `
		SELECT resource_id, owner_id, owner_type, permissions, created_at, updated_at
		FROM access_control.resource_ownership
		WHERE resource_id = $1
	`

	var ownership domain.ResourceOwnership
	var permissionsJSON []byte

	err := r.db.QueryRowContext(ctx, query, resourceID).Scan(
		&ownership.ResourceID, &ownership.OwnerID, &ownership.OwnerType,
		&permissionsJSON, &ownership.CreatedAt, &ownership.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get ownership: %w", err)
	}

	if err := json.Unmarshal(permissionsJSON, &ownership.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	return &ownership, nil
}

// FindByOwnerID retrieves all resources owned by a user
func (r *PostgresResourceOwnershipRepository) FindByOwnerID(ctx context.Context, ownerID string) ([]*domain.ResourceOwnership, error) {
	query := `
		SELECT resource_id, owner_id, owner_type, permissions, created_at, updated_at
		FROM access_control.resource_ownership
		WHERE owner_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query ownership: %w", err)
	}
	defer rows.Close()

	var ownerships []*domain.ResourceOwnership
	for rows.Next() {
		var ownership domain.ResourceOwnership
		var permissionsJSON []byte

		err := rows.Scan(
			&ownership.ResourceID, &ownership.OwnerID, &ownership.OwnerType,
			&permissionsJSON, &ownership.CreatedAt, &ownership.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan ownership: %w", err)
		}

		if err := json.Unmarshal(permissionsJSON, &ownership.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}

		ownerships = append(ownerships, &ownership)
	}

	return ownerships, nil
}

// Create creates a new ownership record
func (r *PostgresResourceOwnershipRepository) Create(ctx context.Context, ownership *domain.ResourceOwnership) error {
	permissionsJSON, err := json.Marshal(ownership.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	query := `
		INSERT INTO access_control.resource_ownership (resource_id, owner_id, owner_type, permissions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = r.db.ExecContext(ctx, query,
		ownership.ResourceID, ownership.OwnerID, ownership.OwnerType,
		permissionsJSON, ownership.CreatedAt, ownership.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create ownership: %w", err)
	}

	return nil
}

// Update updates an ownership record
func (r *PostgresResourceOwnershipRepository) Update(ctx context.Context, ownership *domain.ResourceOwnership) error {
	permissionsJSON, err := json.Marshal(ownership.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	query := `
		UPDATE access_control.resource_ownership
		SET owner_id = $1, owner_type = $2, permissions = $3, updated_at = $4
		WHERE resource_id = $5
	`

	_, err = r.db.ExecContext(ctx, query,
		ownership.OwnerID, ownership.OwnerType, permissionsJSON,
		time.Now(), ownership.ResourceID,
	)

	if err != nil {
		return fmt.Errorf("failed to update ownership: %w", err)
	}

	return nil
}

// Delete deletes an ownership record
func (r *PostgresResourceOwnershipRepository) Delete(ctx context.Context, resourceID string) error {
	query := `DELETE FROM access_control.resource_ownership WHERE resource_id = $1`

	_, err := r.db.ExecContext(ctx, query, resourceID)
	if err != nil {
		return fmt.Errorf("failed to delete ownership: %w", err)
	}

	return nil
}

// Compile-time check that PostgresResourceOwnershipRepository implements ResourceOwnershipRepository
var _ ports.ResourceOwnershipRepository = (*PostgresResourceOwnershipRepository)(nil)
