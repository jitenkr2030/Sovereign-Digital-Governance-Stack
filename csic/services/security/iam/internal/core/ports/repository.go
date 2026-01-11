package ports

import (
	"context"

	"github.com/csic-platform/services/security/iam/internal/core/domain"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// FindByID retrieves a user by their unique identifier
	FindByID(ctx context.Context, id string) (*domain.User, error)

	// FindByUsername retrieves a user by their username
	FindByUsername(ctx context.Context, username string) (*domain.User, error)

	// FindByEmail retrieves a user by their email address
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	// FindByAPIKey retrieves a user by their API key
	FindByAPIKey(ctx context.Context, apiKey string) (*domain.User, error)

	// FindAll retrieves all users with optional filtering
	FindAll(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error)

	// Create creates a new user
	Create(ctx context.Context, user *domain.User) error

	// Update updates an existing user
	Update(ctx context.Context, user *domain.User) error

	// Delete soft-deletes a user
	Delete(ctx context.Context, id string) error

	// UpdateLogin updates the last login timestamp
	UpdateLogin(ctx context.Context, id string) error

	// IncrementLoginAttempts increments the login attempts counter
	IncrementLoginAttempts(ctx context.Context, id string) error

	// ResetLoginAttempts resets the login attempts counter
	ResetLoginAttempts(ctx context.Context, id string) error
}

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	// FindByID retrieves a role by its unique identifier
	FindByID(ctx context.Context, id string) (*domain.Role, error)

	// FindByName retrieves a role by its name
	FindByName(ctx context.Context, name string) (*domain.Role, error)

	// FindAll retrieves all roles
	FindAll(ctx context.Context) ([]*domain.Role, error)

	// Create creates a new role
	Create(ctx context.Context, role *domain.Role) error

	// Update updates an existing role
	Update(ctx context.Context, role *domain.Role) error

	// Delete deletes a role
	Delete(ctx context.Context, id string) error
}

// PermissionRepository defines the interface for permission data operations
type PermissionRepository interface {
	// FindByID retrieves a permission by its unique identifier
	FindByID(ctx context.Context, id string) (*domain.Permission, error)

	// FindByResource retrieves permissions for a specific resource
	FindByResource(ctx context.Context, resource string) ([]*domain.Permission, error)

	// FindAll retrieves all permissions
	FindAll(ctx context.Context) ([]*domain.Permission, error)

	// Create creates a new permission
	Create(ctx context.Context, permission *domain.Permission) error

	// Delete deletes a permission
	Delete(ctx context.Context, id string) error
}

// SessionRepository defines the interface for session data operations
type SessionRepository interface {
	// FindByID retrieves a session by its unique identifier
	FindByID(ctx context.Context, id string) (*domain.Session, error)

	// FindByToken retrieves a session by its token
	FindByToken(ctx context.Context, token string) (*domain.Session, error)

	// FindByUserID retrieves all sessions for a user
	FindByUserID(ctx context.Context, userID string) ([]*domain.Session, error)

	// Create creates a new session
	Create(ctx context.Context, session *domain.Session) error

	// Revoke revokes a session
	Revoke(ctx context.Context, id string) error

	// RevokeAll revokes all sessions for a user
	RevokeAll(ctx context.Context, userID string) error

	// CleanupExpired removes all expired sessions
	CleanupExpired(ctx context.Context) error
}

// RefreshTokenRepository defines the interface for refresh token operations
type RefreshTokenRepository interface {
	// FindByToken retrieves a refresh token by its token value
	FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error)

	// Create creates a new refresh token
	Create(ctx context.Context, token *domain.RefreshToken) error

	// MarkUsed marks a refresh token as used
	MarkUsed(ctx context.Context, id string) error

	// CleanupExpired removes all expired refresh tokens
	CleanupExpired(ctx context.Context) error
}

// APIKeyRepository defines the interface for API key operations
type APIKeyRepository interface {
	// FindByID retrieves an API key by its unique identifier
	FindByID(ctx context.Context, id string) (*domain.APIKey, error)

	// FindByKeyHash retrieves an API key by its key hash
	FindByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error)

	// FindByUserID retrieves all API keys for a user
	FindByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error)

	// Create creates a new API key
	Create(ctx context.Context, apiKey *domain.APIKey) error

	// Update updates an existing API key
	Update(ctx context.Context, apiKey *domain.APIKey) error

	// Revoke revokes an API key
	Revoke(ctx context.Context, id string) error

	// UpdateLastUsed updates the last used timestamp
	UpdateLastUsed(ctx context.Context, id string) error
}

// Request and response types for UserRepository

type ListUsersRequest struct {
	RoleID    string `json:"role_id,omitempty"`
	Enabled   *bool  `json:"enabled,omitempty"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

type ListUsersResponse struct {
	Users     []*domain.User `json:"users"`
	Total     int            `json:"total"`
	Page      int            `json:"page"`
	PageSize  int            `json:"page_size"`
}
