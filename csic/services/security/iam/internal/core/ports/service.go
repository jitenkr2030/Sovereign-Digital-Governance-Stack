package ports

import (
	"context"

	"github.com/csic-platform/services/security/iam/internal/core/domain"
)

// AuthenticationService defines the business logic for authentication operations
type AuthenticationService interface {
	// Authenticate authenticates a user with username and password
	Authenticate(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error)

	// Register creates a new user account
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error)

	// Refresh refreshes an access token using a refresh token
	Refresh(ctx context.Context, refreshToken string) (*domain.LoginResponse, error)

	// Logout revokes the current session
	Logout(ctx context.Context, token string) error

	// LogoutAll revokes all sessions for a user
	LogoutAll(ctx context.Context, userID string) error

	// ValidateToken validates an access token
	ValidateToken(ctx context.Context, token string) (*domain.AuthContext, error)
}

// UserService defines the business logic for user management
type UserService interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error)

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id string) (*domain.User, error)

	// ListUsers lists all users with filtering
	ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, req *UpdateUserRequest) error

	// DeleteUser soft-deletes a user
	DeleteUser(ctx context.Context, id string) error

	// EnableUser enables a disabled user
	EnableUser(ctx context.Context, id string) error

	// DisableUser disables an active user
	DisableUser(ctx context.Context, id string) error

	// ChangePassword changes a user's password
	ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error

	// ResetPassword resets a user's password (admin action)
	ResetPassword(ctx context.Context, id string) (*ResetPasswordResponse, error)
}

// RoleService defines the business logic for role management
type RoleService interface {
	// CreateRole creates a new role
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*CreateRoleResponse, error)

	// GetRole retrieves a role by ID
	GetRole(ctx context.Context, id string) (*domain.Role, error)

	// ListRoles lists all roles
	ListRoles(ctx context.Context) ([]*domain.Role, error)

	// UpdateRole updates an existing role
	UpdateRole(ctx context.Context, req *UpdateRoleRequest) error

	// DeleteRole deletes a role
	DeleteRole(ctx context.Context, id string) error

	// AssignPermissions assigns permissions to a role
	AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error
}

// SessionService defines the business logic for session management
type SessionService interface {
	// GetSessions retrieves all sessions for a user
	GetSessions(ctx context.Context, userID string) ([]*domain.Session, error)

	// RevokeSession revokes a specific session
	RevokeSession(ctx context.Context, sessionID string) error

	// RevokeAllSessions revokes all sessions for a user
	RevokeAllSessions(ctx context.Context, userID string) error

	// CleanupExpired removes all expired sessions
	CleanupExpired(ctx context.Context) error
}

// APIKeyService defines the business logic for API key management
type APIKeyService interface {
	// CreateAPIKey creates a new API key for a user
	CreateAPIKey(ctx context.Context, userID string, req *CreateAPIKeyRequest) (*CreateAPIKeyResponse, error)

	// GetAPIKeys retrieves all API keys for a user
	GetAPIKeys(ctx context.Context, userID string) ([]*domain.APIKey, error)

	// RevokeAPIKey revokes an API key
	RevokeAPIKey(ctx context.Context, id string) error

	// ValidateAPIKey validates an API key and returns the associated user
	ValidateAPIKey(ctx context.Context, key string) (*domain.User, error)
}

// MFAService defines the business logic for multi-factor authentication
type MFAService interface {
	// SetupMFA sets up MFA for a user
	SetupMFA(ctx context.Context, userID string) (*MFASetupResponse, error)

	// EnableMFA enables MFA for a user after verification
	EnableMFA(ctx context.Context, userID string, code string) error

	// DisableMFA disables MFA for a user
	DisableMFA(ctx context.Context, userID string, code string) error

	// VerifyMFA verifies an MFA code
	VerifyMFA(ctx context.Context, userID string, code string) bool

	// GenerateBackupCodes generates new backup codes for a user
	GenerateBackupCodes(ctx context.Context, userID string) ([]string, error)
}

// Request and response types for UserService

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	RoleID   string `json:"role_id"`
}

type CreateUserResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	ID       string  `json:"id" validate:"required"`
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	RoleID   *string `json:"role_id,omitempty"`
}

type ResetPasswordResponse struct {
	NewPassword string `json:"new_password"`
}

// Request and response types for RoleService

type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type CreateRoleResponse struct {
	RoleID string `json:"role_id"`
	Name   string `json:"name"`
}

type UpdateRoleRequest struct {
	ID          string   `json:"id" validate:"required"`
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// Request and response types for APIKeyService

type CreateAPIKeyRequest struct {
	Name        string   `json:"name" validate:"required"`
	Permissions []string `json:"permissions"`
	Scopes      []string `json:"scopes"`
	ExpiresAt   string   `json:"expires_at,omitempty"`
}

type CreateAPIKeyResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	APIKey    string `json:"api_key"` // Only returned once
	CreatedAt string `json:"created_at"`
}

// Request and response types for MFAService

type MFASetupResponse struct {
	Secret     string   `json:"secret"`
	QRCodeURL  string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}
