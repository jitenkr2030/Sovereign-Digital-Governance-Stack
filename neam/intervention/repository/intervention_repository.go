package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"intervention-service/models"

	_ "github.com/lib/pq"
)

// PostgresRepository handles database operations
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(config map[string]string) (*PostgresRepository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config["host"], 5432, config["user"], config["password"], config["name"], "require",
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &PostgresRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return repo, nil
}

// Close closes the database connection
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// migrate runs database migrations
func (r *PostgresRepository) migrate() error {
	migrations := []string{
		// Policies table
		`CREATE TABLE IF NOT EXISTS policies (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			type VARCHAR(50) NOT NULL,
			category VARCHAR(100),
			rules JSONB,
			conditions TEXT,
			enforcement_level VARCHAR(20) NOT NULL,
			scope JSONB,
			priority INTEGER DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'draft',
			version VARCHAR(20) NOT NULL DEFAULT '1.0.0',
			created_by UUID,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			activated_at TIMESTAMP WITH TIME ZONE
		)`,

		// Interventions table
		`CREATE TABLE IF NOT EXISTS interventions (
			id UUID PRIMARY KEY,
			policy_id UUID,
			type VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			parameters JSONB,
			target_scope JSONB,
			trigger JSONB,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			priority INTEGER DEFAULT 0,
			workflow_id UUID,
			executed_by UUID,
			executed_at TIMESTAMP WITH TIME ZONE,
			completed_at TIMESTAMP WITH TIME ZONE,
			rolled_back_at TIMESTAMP WITH TIME ZONE,
			result JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Emergency locks table
		`CREATE TABLE IF NOT EXISTS emergency_locks (
			id UUID PRIMARY KEY,
			lock_type VARCHAR(50) NOT NULL,
			resource_id VARCHAR(255) NOT NULL,
			resource_type VARCHAR(100) NOT NULL,
			scope JSONB,
			reason TEXT NOT NULL,
			authorized_by UUID NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			priority INTEGER DEFAULT 0,
			expires_at TIMESTAMP WITH TIME ZONE,
			released_at TIMESTAMP WITH TIME ZONE,
			released_by UUID,
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Overrides table
		`CREATE TABLE IF NOT EXISTS overrides (
			id UUID PRIMARY KEY,
			type VARCHAR(50) NOT NULL,
			target_id VARCHAR(255) NOT NULL,
			override_value JSONB NOT NULL,
			reason TEXT NOT NULL,
			authorized_by UUID NOT NULL,
			signer_id UUID NOT NULL,
			signature TEXT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			used_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Scenarios table
		`CREATE TABLE IF NOT EXISTS scenarios (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			type VARCHAR(50) NOT NULL,
			parameters JSONB NOT NULL,
			constraints JSONB,
			status VARCHAR(20) NOT NULL DEFAULT 'draft',
			results JSONB,
			created_by UUID NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Scenario templates table
		`CREATE TABLE IF NOT EXISTS scenario_templates (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			type VARCHAR(50) NOT NULL,
			template JSONB NOT NULL,
			created_by UUID NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_policies_type ON policies(type)`,
		`CREATE INDEX IF NOT EXISTS idx_policies_status ON policies(status)`,
		`CREATE INDEX IF NOT EXISTS idx_interventions_policy_id ON interventions(policy_id)`,
		`CREATE INDEX IF NOT EXISTS idx_interventions_status ON interventions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_emergency_locks_resource ON emergency_locks(resource_id, resource_type)`,
		`CREATE INDEX IF NOT EXISTS idx_scenarios_status ON scenarios(status)`,
	}

	for _, migration := range migrations {
		if _, err := r.db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// Policy operations

func (r *PostgresRepository) CreatePolicy(ctx context.Context, policy *models.Policy) error {
	query := `INSERT INTO policies (
		id, name, description, type, category, rules, conditions,
		enforcement_level, scope, priority, status, version, created_by,
		created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	rulesJSON, _ := json.Marshal(policy.Rules)
	scopeJSON, _ := json.Marshal(policy.Scope)

	_, err := r.db.ExecContext(ctx, query,
		policy.ID, policy.Name, policy.Description, policy.Type, policy.Category,
		rulesJSON, policy.Conditions, policy.EnforcementLevel, scopeJSON,
		policy.Priority, policy.Status, policy.Version, policy.CreatedBy,
		policy.CreatedAt, policy.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetPolicy(ctx context.Context, id uuid.UUID) (*models.Policy, error) {
	query := `SELECT id, name, description, type, category, rules, conditions,
		enforcement_level, scope, priority, status, version, created_by,
		created_at, updated_at, activated_at FROM policies WHERE id = $1`

	var policy models.Policy
	var rulesJSON, scopeJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&policy.ID, &policy.Name, &policy.Description, &policy.Type,
		&policy.Category, &rulesJSON, &policy.Conditions, &policy.EnforcementLevel,
		&scopeJSON, &policy.Priority, &policy.Status, &policy.Version,
		&policy.CreatedBy, &policy.CreatedAt, &policy.UpdatedAt, &policy.ActivatedAt,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(rulesJSON, &policy.Rules)
	json.Unmarshal(scopeJSON, &policy.Scope)

	return &policy, nil
}

func (r *PostgresRepository) ListPolicies(ctx context.Context, filters map[string]interface{}) ([]*models.Policy, error) {
	query := `SELECT id, name, description, type, category, rules, conditions,
		enforcement_level, scope, priority, status, version, created_by,
		created_at, updated_at, activated_at FROM policies WHERE status != 'archived'`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*models.Policy
	for rows.Next() {
		var policy models.Policy
		var rulesJSON, scopeJSON []byte

		if err := rows.Scan(
			&policy.ID, &policy.Name, &policy.Description, &policy.Type,
			&policy.Category, &rulesJSON, &policy.Conditions, &policy.EnforcementLevel,
			&scopeJSON, &policy.Priority, &policy.Status, &policy.Version,
			&policy.CreatedBy, &policy.CreatedAt, &policy.UpdatedAt, &policy.ActivatedAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal(rulesJSON, &policy.Rules)
		json.Unmarshal(scopeJSON, &policy.Scope)
		policies = append(policies, &policy)
	}

	return policies, nil
}

func (r *PostgresRepository) UpdatePolicy(ctx context.Context, policy *models.Policy) error {
	query := `UPDATE policies SET
		name = $1, description = $2, type = $3, category = $4, rules = $5,
		conditions = $6, enforcement_level = $7, scope = $8, priority = $9,
		status = $10, version = $11, activated_at = $12, updated_at = $13
		WHERE id = $14`

	rulesJSON, _ := json.Marshal(policy.Rules)
	scopeJSON, _ := json.Marshal(policy.Scope)

	_, err := r.db.ExecContext(ctx, query,
		policy.Name, policy.Description, policy.Type, policy.Category,
		rulesJSON, policy.Conditions, policy.EnforcementLevel, scopeJSON,
		policy.Priority, policy.Status, policy.Version, policy.ActivatedAt,
		policy.UpdatedAt, policy.ID,
	)
	return err
}

// Intervention operations

func (r *PostgresRepository) CreateIntervention(ctx context.Context, intervention *models.Intervention) error {
	query := `INSERT INTO interventions (
		id, policy_id, type, name, description, parameters, target_scope,
		trigger, status, priority, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	paramsJSON, _ := json.Marshal(intervention.Parameters)
	scopeJSON, _ := json.Marshal(intervention.TargetScope)
	triggerJSON, _ := json.Marshal(intervention.Trigger)

	_, err := r.db.ExecContext(ctx, query,
		intervention.ID, intervention.PolicyID, intervention.Type, intervention.Name,
		intervention.Description, paramsJSON, scopeJSON, triggerJSON,
		intervention.Status, intervention.Priority, intervention.CreatedAt, intervention.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetIntervention(ctx context.Context, id uuid.UUID) (*models.Intervention, error) {
	query := `SELECT id, policy_id, type, name, description, parameters, target_scope,
		trigger, status, priority, workflow_id, executed_by, executed_at,
		completed_at, rolled_back_at, result, created_at, updated_at
		FROM interventions WHERE id = $1`

	var intervention models.Intervention
	var paramsJSON, scopeJSON, triggerJSON, resultJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&intervention.ID, &intervention.PolicyID, &intervention.Type,
		&intervention.Name, &intervention.Description, &paramsJSON, &scopeJSON,
		&triggerJSON, &intervention.Status, &intervention.Priority, &intervention.WorkflowID,
		&intervention.ExecutedBy, &intervention.ExecutedAt, &intervention.CompletedAt,
		&intervention.RolledBackAt, &resultJSON, &intervention.CreatedAt, &intervention.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(paramsJSON, &intervention.Parameters)
	json.Unmarshal(scopeJSON, &intervention.TargetScope)
	json.Unmarshal(triggerJSON, &intervention.Trigger)
	if resultJSON != nil {
		json.Unmarshal(resultJSON, &intervention.Result)
	}

	return &intervention, nil
}

func (r *PostgresRepository) ListInterventions(ctx context.Context, filters map[string]interface{}) ([]*models.Intervention, error) {
	query := `SELECT id, policy_id, type, name, description, parameters, target_scope,
		trigger, status, priority, workflow_id, executed_by, executed_at,
		completed_at, rolled_back_at, result, created_at, updated_at
		FROM interventions ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interventions []*models.Intervention
	for rows.Next() {
		var intervention models.Intervention
		var paramsJSON, scopeJSON, triggerJSON, resultJSON []byte

		if err := rows.Scan(
			&intervention.ID, &intervention.PolicyID, &intervention.Type,
			&intervention.Name, &intervention.Description, &paramsJSON, &scopeJSON,
			&triggerJSON, &intervention.Status, &intervention.Priority, &intervention.WorkflowID,
			&intervention.ExecutedBy, &intervention.ExecutedAt, &intervention.CompletedAt,
			&intervention.RolledBackAt, &resultJSON, &intervention.CreatedAt, &intervention.UpdatedAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal(paramsJSON, &intervention.Parameters)
		json.Unmarshal(scopeJSON, &intervention.TargetScope)
		json.Unmarshal(triggerJSON, &intervention.Trigger)
		if resultJSON != nil {
			json.Unmarshal(resultJSON, &intervention.Result)
		}
		interventions = append(interventions, &intervention)
	}

	return interventions, nil
}

func (r *PostgresRepository) UpdateIntervention(ctx context.Context, intervention *models.Intervention) error {
	query := `UPDATE interventions SET
		status = $1, workflow_id = $2, executed_by = $3, executed_at = $4,
		completed_at = $5, rolled_back_at = $6, result = $7, updated_at = $8
		WHERE id = $9`

	resultJSON, _ := json.Marshal(intervention.Result)

	_, err := r.db.ExecContext(ctx, query,
		intervention.Status, intervention.WorkflowID, intervention.ExecutedBy,
		intervention.ExecutedAt, intervention.CompletedAt, intervention.RolledBackAt,
		resultJSON, intervention.UpdatedAt, intervention.ID,
	)
	return err
}

// Emergency lock operations

func (r *PostgresRepository) CreateEmergencyLock(ctx context.Context, lock *models.EmergencyLock) error {
	query := `INSERT INTO emergency_locks (
		id, lock_type, resource_id, resource_type, scope, reason, authorized_by,
		status, priority, expires_at, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	scopeJSON, _ := json.Marshal(lock.Scope)
	metadataJSON, _ := json.Marshal(lock.Metadata)

	_, err := r.db.ExecContext(ctx, query,
		lock.ID, lock.LockType, lock.ResourceID, lock.ResourceType, scopeJSON,
		lock.Reason, lock.AuthorizedBy, lock.Status, lock.Priority, lock.ExpiresAt,
		lock.CreatedAt, lock.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetEmergencyLock(ctx context.Context, id uuid.UUID) (*models.EmergencyLock, error) {
	query := `SELECT id, lock_type, resource_id, resource_type, scope, reason,
		authorized_by, status, priority, expires_at, released_at, released_by,
		metadata, created_at, updated_at FROM emergency_locks WHERE id = $1`

	var lock models.EmergencyLock
	var scopeJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lock.ID, &lock.LockType, &lock.ResourceID, &lock.ResourceType, &scopeJSON,
		&lock.Reason, &lock.AuthorizedBy, &lock.Status, &lock.Priority, &lock.ExpiresAt,
		&lock.ReleasedAt, &lock.ReleasedBy, &metadataJSON, &lock.CreatedAt, &lock.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(scopeJSON, &lock.Scope)
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &lock.Metadata)
	}

	return &lock, nil
}

func (r *PostgresRepository) ListEmergencyLocks(ctx context.Context, filters map[string]interface{}) ([]*models.EmergencyLock, error) {
	query := `SELECT id, lock_type, resource_id, resource_type, scope, reason,
		authorized_by, status, priority, expires_at, released_at, released_by,
		metadata, created_at, updated_at FROM emergency_locks WHERE status = 'active'
		ORDER BY priority DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locks []*models.EmergencyLock
	for rows.Next() {
		var lock models.EmergencyLock
		var scopeJSON, metadataJSON []byte

		if err := rows.Scan(
			&lock.ID, &lock.LockType, &lock.ResourceID, &lock.ResourceType, &scopeJSON,
			&lock.Reason, &lock.AuthorizedBy, &lock.Status, &lock.Priority, &lock.ExpiresAt,
			&lock.ReleasedAt, &lock.ReleasedBy, &metadataJSON, &lock.CreatedAt, &lock.UpdatedAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal(scopeJSON, &lock.Scope)
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &lock.Metadata)
		}
		locks = append(locks, &lock)
	}

	return locks, nil
}

func (r *PostgresRepository) UpdateEmergencyLock(ctx context.Context, lock *models.EmergencyLock) error {
	query := `UPDATE emergency_locks SET
		status = $1, released_at = $2, released_by = $3, updated_at = $4
		WHERE id = $5`

	_, err := r.db.ExecContext(ctx, query,
		lock.Status, lock.ReleasedAt, lock.ReleasedBy, lock.UpdatedAt, lock.ID,
	)
	return err
}

func (r *PostgresRepository) GetActiveLocksCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, 
		"SELECT COUNT(*) FROM emergency_locks WHERE status = 'active'").Scan(&count)
	return count, err
}

func (r *PostgresRepository) CheckLockExists(ctx context.Context, resourceType, resourceID string) (bool, *models.EmergencyLock, error) {
	query := `SELECT id, lock_type, resource_id, resource_type, scope, reason,
		authorized_by, status, priority, expires_at, released_at, released_by,
		metadata, created_at, updated_at FROM emergency_locks 
		WHERE resource_type = $1 AND resource_id = $2 AND status = 'active'
		LIMIT 1`

	var lock models.EmergencyLock
	var scopeJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, resourceType, resourceID).Scan(
		&lock.ID, &lock.LockType, &lock.ResourceID, &lock.ResourceType, &scopeJSON,
		&lock.Reason, &lock.AuthorizedBy, &lock.Status, &lock.Priority, &lock.ExpiresAt,
		&lock.ReleasedAt, &lock.ReleasedBy, &metadataJSON, &lock.CreatedAt, &lock.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}

	json.Unmarshal(scopeJSON, &lock.Scope)
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &lock.Metadata)
	}

	return true, &lock, nil
}

// Override operations

func (r *PostgresRepository) CreateOverride(ctx context.Context, override *models.Override) error {
	query := `INSERT INTO overrides (
		id, type, target_id, override_value, reason, authorized_by, signer_id,
		signature, status, expires_at, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	valueJSON, _ := json.Marshal(override.OverrideValue)

	_, err := r.db.ExecContext(ctx, query,
		override.ID, override.Type, override.TargetID, valueJSON, override.Reason,
		override.AuthorizedBy, override.SignerID, override.Signature, override.Status,
		override.ExpiresAt, override.CreatedAt,
	)
	return err
}

// Scenario operations

func (r *PostgresRepository) CreateScenario(ctx context.Context, scenario *models.Scenario) error {
	query := `INSERT INTO scenarios (
		id, name, description, type, parameters, constraints, status, results,
		created_by, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	paramsJSON, _ := json.Marshal(scenario.Parameters)
	constraintsJSON, _ := json.Marshal(scenario.Constraints)
	var resultsJSON []byte
	if scenario.Results != nil {
		resultsJSON, _ = json.Marshal(scenario.Results)
	}

	_, err := r.db.ExecContext(ctx, query,
		scenario.ID, scenario.Name, scenario.Description, scenario.Type, paramsJSON,
		constraintsJSON, scenario.Status, resultsJSON, scenario.CreatedBy,
		scenario.CreatedAt, scenario.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetScenario(ctx context.Context, id uuid.UUID) (*models.Scenario, error) {
	query := `SELECT id, name, description, type, parameters, constraints, status,
		results, created_by, created_at, updated_at FROM scenarios WHERE id = $1`

	var scenario models.Scenario
	var paramsJSON, constraintsJSON, resultsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&scenario.ID, &scenario.Name, &scenario.Description, &scenario.Type,
		&paramsJSON, &constraintsJSON, &scenario.Status, &resultsJSON,
		&scenario.CreatedBy, &scenario.CreatedAt, &scenario.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(paramsJSON, &scenario.Parameters)
	json.Unmarshal(constraintsJSON, &scenario.Constraints)
	if resultsJSON != nil {
		json.Unmarshal(resultsJSON, &scenario.Results)
	}

	return &scenario, nil
}

func (r *PostgresRepository) ListScenarios(ctx context.Context, filters map[string]interface{}) ([]*models.Scenario, error) {
	query := `SELECT id, name, description, type, parameters, constraints, status,
		results, created_by, created_at, updated_at FROM scenarios
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scenarios []*models.Scenario
	for rows.Next() {
		var scenario models.Scenario
		var paramsJSON, constraintsJSON, resultsJSON []byte

		if err := rows.Scan(
			&scenario.ID, &scenario.Name, &scenario.Description, &scenario.Type,
			&paramsJSON, &constraintsJSON, &scenario.Status, &resultsJSON,
			&scenario.CreatedBy, &scenario.CreatedAt, &scenario.UpdatedAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal(paramsJSON, &scenario.Parameters)
		json.Unmarshal(constraintsJSON, &scenario.Constraints)
		if resultsJSON != nil {
			json.Unmarshal(resultsJSON, &scenario.Results)
		}
		scenarios = append(scenarios, &scenario)
	}

	return scenarios, nil
}

func (r *PostgresRepository) UpdateScenario(ctx context.Context, scenario *models.Scenario) error {
	query := `UPDATE scenarios SET
		status = $1, results = $2, updated_at = $3 WHERE id = $4`

	resultsJSON, _ := json.Marshal(scenario.Results)

	_, err := r.db.ExecContext(ctx, query,
		scenario.Status, resultsJSON, scenario.UpdatedAt, scenario.ID,
	)
	return err
}

func (r *PostgresRepository) GetScenarioResults(ctx context.Context, runID uuid.UUID) (*models.ScenarioResults, error) {
	query := `SELECT results FROM scenarios WHERE id = $1`

	var resultsJSON []byte
	err := r.db.QueryRowContext(ctx, query, runID).Scan(&resultsJSON)
	if err != nil {
		return nil, err
	}

	var results models.ScenarioResults
	json.Unmarshal(resultsJSON, &results)
	return &results, nil
}

// Scenario template operations

func (r *PostgresRepository) CreateScenarioTemplate(ctx context.Context, template *models.ScenarioTemplate) error {
	query := `INSERT INTO scenario_templates (
		id, name, description, type, template, created_by, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	templateJSON, _ := json.Marshal(template.Template)

	_, err := r.db.ExecContext(ctx, query,
		template.ID, template.Name, template.Description, template.Type, templateJSON,
		template.CreatedBy, template.CreatedAt, template.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) ListScenarioTemplates(ctx context.Context) ([]*models.ScenarioTemplate, error) {
	query := `SELECT id, name, description, type, template, created_by,
		created_at, updated_at FROM scenario_templates ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []*models.ScenarioTemplate
	for rows.Next() {
		var template models.ScenarioTemplate
		var templateJSON []byte

		if err := rows.Scan(
			&template.ID, &template.Name, &template.Description, &template.Type,
			&templateJSON, &template.CreatedBy, &template.CreatedAt, &template.UpdatedAt,
		); err != nil {
			return nil, err
		}

		json.Unmarshal(templateJSON, &template.Template)
		templates = append(templates, &template)
	}

	return templates, nil
}
