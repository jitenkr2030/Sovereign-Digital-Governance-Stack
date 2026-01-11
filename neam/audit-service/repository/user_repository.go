package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"neam-platform/audit-service/models"
)

// UserRepository provides user data access
type UserRepository interface {
	// CRUD operations
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error

	// Query operations
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByRole(ctx context.Context, role string) ([]models.User, error)
	List(ctx context.Context, limit, offset int) ([]models.User, error)

	// Session operations
	CreateSession(ctx context.Context, session *models.Session) error
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	UpdateSessionActivity(ctx context.Context, sessionID string) error

	// Authentication operations
	IncrementFailedLogins(ctx context.Context, userID string) error
	ResetFailedLogins(ctx context.Context, userID string) error
	UpdateLastLogin(ctx context.Context, userID string, loginTime *time.Time) error

	// Policy operations
	CreatePolicy(ctx context.Context, policy *models.Policy) error
	GetPolicy(ctx context.Context, policyID string) (*models.Policy, error)
	UpdatePolicy(ctx context.Context, policy *models.Policy) error
	DeletePolicy(ctx context.Context, policyID string) error
	ListPolicies(ctx context.Context) ([]models.Policy, error)
}

// userRepository implements UserRepository
type userRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, username, email, password_hash, role, permissions,
			groups, active, mfa_enabled, mfa_secret, locked,
			locked_until, failed_logins, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.Role, user.Permissions,
		user.Groups, user.Active, user.MFEnabled, user.MFSecret, user.Locked,
		user.LockedUntil, user.FailedLogins, user.CreatedAt, user.UpdatedAt,
	)

	return err
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, permissions,
			groups, active, mfa_enabled, mfa_secret, locked,
			locked_until, failed_logins, created_at, updated_at, last_login, password_changed
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.Permissions,
		&user.Groups, &user.Active, &user.MFEnabled, &user.MFSecret, &user.Locked,
		&user.LockedUntil, &user.FailedLogins, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.PasswordChanged,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return &user, err
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET
			username = $2, email = $3, password_hash = $4, role = $5, permissions = $6,
			groups = $7, active = $8, mfa_enabled = $9, mfa_secret = $10, locked = $11,
			locked_until = $12, failed_logins = $13, updated_at = $14, last_login = $15, password_changed = $16
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.Role, user.Permissions,
		user.Groups, user.Active, user.MFEnabled, user.MFSecret, user.Locked,
		user.LockedUntil, user.FailedLogins, user.UpdatedAt, user.LastLogin, user.PasswordChanged,
	)

	return err
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := "UPDATE users SET active = false, updated_at = $2 WHERE id = $1"
	_, err := r.pool.Exec(ctx, query, id, time.Now())
	return err
}

// FindByUsername finds a user by username
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, permissions,
			groups, active, mfa_enabled, mfa_secret, locked,
			locked_until, failed_logins, created_at, updated_at, last_login, password_changed
		FROM users
		WHERE username = $1 AND active = true
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.Permissions,
		&user.Groups, &user.Active, &user.MFEnabled, &user.MFSecret, &user.Locked,
		&user.LockedUntil, &user.FailedLogins, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.PasswordChanged,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return &user, err
}

// FindByEmail finds a user by email
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, permissions,
			groups, active, mfa_enabled, mfa_secret, locked,
			locked_until, failed_logins, created_at, updated_at, last_login, password_changed
		FROM users
		WHERE email = $1 AND active = true
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.Permissions,
		&user.Groups, &user.Active, &user.MFEnabled, &user.MFSecret, &user.Locked,
		&user.LockedUntil, &user.FailedLogins, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.PasswordChanged,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return &user, err
}

// FindByRole finds users by role
func (r *userRepository) FindByRole(ctx context.Context, role string) ([]models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, permissions,
			groups, active, mfa_enabled, mfa_secret, locked,
			locked_until, failed_logins, created_at, updated_at, last_login, password_changed
		FROM users
		WHERE role = $1 AND active = true
	`

	rows, err := r.pool.Query(ctx, query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.Permissions,
			&user.Groups, &user.Active, &user.MFEnabled, &user.MFSecret, &user.Locked,
			&user.LockedUntil, &user.FailedLogins, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.PasswordChanged,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// List lists users with pagination
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, permissions,
			groups, active, mfa_enabled, mfa_secret, locked,
			locked_until, failed_logins, created_at, updated_at, last_login, password_changed
		FROM users
		WHERE active = true
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.Permissions,
			&user.Groups, &user.Active, &user.MFEnabled, &user.MFSecret, &user.Locked,
			&user.LockedUntil, &user.FailedLogins, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.PasswordChanged,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// CreateSession creates a new session
func (r *userRepository) CreateSession(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (
			id, user_id, token, ip_address, user_agent, location,
			expires_at, last_activity, created_at, revoked, revoked_at, revoke_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.pool.Exec(ctx, query,
		session.ID, session.UserID, session.Token, session.IPAddress, session.UserAgent, session.Location,
		session.ExpiresAt, session.LastActivity, session.CreatedAt, session.Revoked, session.RevokedAt, session.RevokeReason,
	)

	return err
}

// GetSession retrieves a session by ID
func (r *userRepository) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	query := `
		SELECT id, user_id, token, ip_address, user_agent, location,
			expires_at, last_activity, created_at, revoked, revoked_at, revoke_reason
		FROM sessions
		WHERE id = $1 AND revoked = false AND expires_at > NOW()
	`

	var session models.Session
	err := r.pool.QueryRow(ctx, query, sessionID).Scan(
		&session.ID, &session.UserID, &session.Token, &session.IPAddress, &session.UserAgent, &session.Location,
		&session.ExpiresAt, &session.LastActivity, &session.CreatedAt, &session.Revoked, &session.RevokedAt, &session.RevokeReason,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("session not found or expired")
	}
	return &session, err
}

// RevokeSession revokes a session
func (r *userRepository) RevokeSession(ctx context.Context, sessionID string) error {
	query := `
		UPDATE sessions SET
			revoked = true, revoked_at = $2, revoke_reason = $3
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, sessionID, time.Now(), "user_logout")
	return err
}

// UpdateSessionActivity updates session last activity
func (r *userRepository) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	query := `
		UPDATE sessions SET last_activity = $2 WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, sessionID, time.Now())
	return err
}

// IncrementFailedLogins increments failed login count
func (r *userRepository) IncrementFailedLogins(ctx context.Context, userID string) error {
	query := `
		UPDATE users SET failed_logins = failed_logins + 1 WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// ResetFailedLogins resets failed login count
func (r *userRepository) ResetFailedLogins(ctx context.Context, userID string) error {
	query := `
		UPDATE users SET failed_logins = 0 WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// UpdateLastLogin updates last login time
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID string, loginTime *time.Time) error {
	query := `
		UPDATE users SET last_login = $2, failed_logins = 0 WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, userID, loginTime)
	return err
}

// CreatePolicy creates a new policy
func (r *userRepository) CreatePolicy(ctx context.Context, policy *models.Policy) error {
	query := `
		INSERT INTO policies (
			id, name, description, effect, priority, active, conditions,
			resources, actions, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.pool.Exec(ctx, query,
		policy.ID, policy.Name, policy.Description, policy.Effect, policy.Priority, policy.Active,
		policy.Conditions, policy.Resources, policy.Actions, policy.CreatedAt, policy.UpdatedAt,
	)

	return err
}

// GetPolicy retrieves a policy by ID
func (r *userRepository) GetPolicy(ctx context.Context, policyID string) (*models.Policy, error) {
	query := `
		SELECT id, name, description, effect, priority, active, conditions,
			resources, actions, created_at, updated_at
		FROM policies
		WHERE id = $1
	`

	var policy models.Policy
	err := r.pool.QueryRow(ctx, query, policyID).Scan(
		&policy.ID, &policy.Name, &policy.Description, &policy.Effect, &policy.Priority, &policy.Active,
		&policy.Conditions, &policy.Resources, &policy.Actions, &policy.CreatedAt, &policy.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("policy not found")
	}
	return &policy, err
}

// UpdatePolicy updates a policy
func (r *userRepository) UpdatePolicy(ctx context.Context, policy *models.Policy) error {
	query := `
		UPDATE policies SET
			name = $2, description = $3, effect = $4, priority = $5, active = $6,
			conditions = $7, resources = $8, actions = $9, updated_at = $10
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		policy.ID, policy.Name, policy.Description, policy.Effect, policy.Priority, policy.Active,
		policy.Conditions, policy.Resources, policy.Actions, policy.UpdatedAt,
	)

	return err
}

// DeletePolicy deletes a policy
func (r *userRepository) DeletePolicy(ctx context.Context, policyID string) error {
	query := "DELETE FROM policies WHERE id = $1"
	_, err := r.pool.Exec(ctx, query, policyID)
	return err
}

// ListPolicies lists all policies
func (r *userRepository) ListPolicies(ctx context.Context) ([]models.Policy, error) {
	query := `
		SELECT id, name, description, effect, priority, active, conditions,
			resources, actions, created_at, updated_at
		FROM policies
		ORDER BY priority DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []models.Policy
	for rows.Next() {
		var policy models.Policy
		err := rows.Scan(
			&policy.ID, &policy.Name, &policy.Description, &policy.Effect, &policy.Priority, &policy.Active,
			&policy.Conditions, &policy.Resources, &policy.Actions, &policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, rows.Err()
}
