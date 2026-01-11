package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"csic-platform/control-layer/internal/core/domain"
	"csic-platform/control-layer/internal/core/ports"
)

// PostgresPolicyRepository implements PolicyRepository using PostgreSQL
type PostgresPolicyRepository struct {
	db          *sql.DB
	tablePrefix string
}

// NewPostgresPolicyRepository creates a new PostgreSQL policy repository
func NewPostgresPolicyRepository(databaseURL string) (ports.PolicyRepository, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresPolicyRepository{
		db:          db,
		tablePrefix: "control_layer_",
	}, nil
}

// Close closes the database connection
func (r *PostgresPolicyRepository) Close() error {
	return r.db.Close()
}

// tableName returns the prefixed table name
func (r *PostgresPolicyRepository) tableName(name string) string {
	return r.tablePrefix + name
}

// GetPolicyByID retrieves a policy by its ID
func (r *PostgresPolicyRepository) GetPolicyByID(ctx context.Context, id uuid.UUID) (*domain.Policy, error) {
	query := fmt.Sprintf(`
		SELECT id, name, description, rule_json, priority, is_active, created_at, updated_at
		FROM %s
		WHERE id = $1
	`, r.tableName("policies"))

	var policy domain.Policy
	var ruleJSON []byte
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&policy.ID,
		&policy.Name,
		&description,
		&ruleJSON,
		&policy.Priority,
		&policy.IsActive,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	if description.Valid {
		policy.Description = description.String
	}

	// Parse rule JSON
	if err := json.Unmarshal(ruleJSON, &policy.Rule); err != nil {
		return nil, fmt.Errorf("failed to parse policy rule: %w", err)
	}

	return &policy, nil
}

// GetAllPolicies retrieves all policies
func (r *PostgresPolicyRepository) GetAllPolicies(ctx context.Context) ([]*domain.Policy, error) {
	query := fmt.Sprintf(`
		SELECT id, name, description, rule_json, priority, is_active, created_at, updated_at
		FROM %s
		ORDER BY priority DESC, created_at ASC
	`, r.tableName("policies"))

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()

	var policies []*domain.Policy

	for rows.Next() {
		var policy domain.Policy
		var ruleJSON []byte
		var description sql.NullString

		if err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&description,
			&ruleJSON,
			&policy.Priority,
			&policy.IsActive,
			&policy.CreatedAt,
			&policy.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		if description.Valid {
			policy.Description = description.String
		}

		// Parse rule JSON
		if err := json.Unmarshal(ruleJSON, &policy.Rule); err != nil {
			return nil, fmt.Errorf("failed to parse policy rule: %w", err)
		}

		policies = append(policies, &policy)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating policies: %w", err)
	}

	return policies, nil
}

// GetActivePolicies retrieves all active policies
func (r *PostgresPolicyRepository) GetActivePolicies(ctx context.Context) ([]*domain.Policy, error) {
	query := fmt.Sprintf(`
		SELECT id, name, description, rule_json, priority, is_active, created_at, updated_at
		FROM %s
		WHERE is_active = true
		ORDER BY priority DESC, created_at ASC
	`, r.tableName("policies"))

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active policies: %w", err)
	}
	defer rows.Close()

	var policies []*domain.Policy

	for rows.Next() {
		var policy domain.Policy
		var ruleJSON []byte
		var description sql.NullString

		if err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&description,
			&ruleJSON,
			&policy.Priority,
			&policy.IsActive,
			&policy.CreatedAt,
			&policy.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		if description.Valid {
			policy.Description = description.String
		}

		// Parse rule JSON
		if err := json.Unmarshal(ruleJSON, &policy.Rule); err != nil {
			return nil, fmt.Errorf("failed to parse policy rule: %w", err)
		}

		policies = append(policies, &policy)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating policies: %w", err)
	}

	return policies, nil
}

// CreatePolicy creates a new policy
func (r *PostgresPolicyRepository) CreatePolicy(ctx context.Context, policy *domain.Policy) error {
	ruleJSON, err := json.Marshal(policy.Rule)
	if err != nil {
		return fmt.Errorf("failed to marshal policy rule: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, name, description, rule_json, priority, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, r.tableName("policies"))

	_, err = r.db.ExecContext(ctx, query,
		policy.ID,
		policy.Name,
		policy.Description,
		ruleJSON,
		policy.Priority,
		policy.IsActive,
		policy.CreatedAt,
		policy.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	return nil
}

// UpdatePolicy updates an existing policy
func (r *PostgresPolicyRepository) UpdatePolicy(ctx context.Context, policy *domain.Policy) error {
	ruleJSON, err := json.Marshal(policy.Rule)
	if err != nil {
		return fmt.Errorf("failed to marshal policy rule: %w", err)
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET name = $1, description = $2, rule_json = $3, priority = $4, is_active = $5, updated_at = $6
		WHERE id = $7
	`, r.tableName("policies"))

	result, err := r.db.ExecContext(ctx, query,
		policy.Name,
		policy.Description,
		ruleJSON,
		policy.Priority,
		policy.IsActive,
		time.Now(),
		policy.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("policy not found: %s", policy.ID)
	}

	return nil
}

// DeletePolicy deletes a policy by ID
func (r *PostgresPolicyRepository) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, r.tableName("policies"))

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("policy not found: %s", id)
	}

	return nil
}

// GetPoliciesByTarget retrieves policies targeting a specific service
func (r *PostgresPolicyRepository) GetPoliciesByTarget(ctx context.Context, target string) ([]*domain.Policy, error) {
	query := fmt.Sprintf(`
		SELECT id, name, description, rule_json, priority, is_active, created_at, updated_at
		FROM %s
		WHERE is_active = true AND rule_json::text LIKE $1
		ORDER BY priority DESC
	`, r.tableName("policies"))

	pattern := fmt.Sprintf("%%\"target\":\\s*\"%s\"%%", target)
	rows, err := r.db.QueryContext(ctx, query, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies by target: %w", err)
	}
	defer rows.Close()

	var policies []*domain.Policy

	for rows.Next() {
		var policy domain.Policy
		var ruleJSON []byte
		var description sql.NullString

		if err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&description,
			&ruleJSON,
			&policy.Priority,
			&policy.IsActive,
			&policy.CreatedAt,
			&policy.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		if description.Valid {
			policy.Description = description.String
		}

		if err := json.Unmarshal(ruleJSON, &policy.Rule); err != nil {
			return nil, fmt.Errorf("failed to parse policy rule: %w", err)
		}

		policies = append(policies, &policy)
	}

	return policies, nil
}
