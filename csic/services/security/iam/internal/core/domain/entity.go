package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user entity in the system
type User struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Username      string     `json:"username" db:"username"`
	Email         string     `json:"email" db:"email"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	RoleID        uuid.UUID   `json:"role_id" db:"role_id"`
	Role          *Role       `json:"role,omitempty"`
	MFAEnabled    bool       `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecret     string      `json:"-" db:"mfa_secret"`
	MFABackupCodes []string   `json:"-"`
	APIKey        string     `json:"-" db:"api_key"`
	APIKeyEnabled bool       `json:"api_key_enabled" db:"api_key_enabled"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	LoginAttempts int        `json:"-" db:"login_attempts"`
	LockedUntil   *time.Time `json:"-" db:"locked_until"`
	Enabled       bool       `json:"enabled" db:"enabled"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// Role represents a role with associated permissions
type Role struct {
	ID          uuid.UUID        `json:"id" db:"id"`
	Name        string           `json:"name" db:"name"`
	Description string           `json:"description" db:"description"`
	Permissions []*Permission    `json:"permissions,omitempty"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}

// Permission represents a system permission
type Permission struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Resource    string     `json:"resource" db:"resource"`
	Action      string     `json:"action" db:"action"`
	Description string     `json:"description" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// Session represents an authenticated session
type Session struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	Token        string     `json:"-" db:"token"`
	RefreshToken string     `json:"-" db:"refresh_token"`
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

// RefreshToken represents a refresh token for session management
type RefreshToken struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Token     string     `json:"-" db:"token"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	Used      bool       `json:"used" db:"used"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	KeyHash     string     `json:"-" db:"key_hash"`
	Name        string     `json:"name" db:"name"`
	Permissions []string   `json:"permissions" db:"-"`
	Scopes      []string   `json:"scopes" db:"-"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	Enabled     bool       `json:"enabled" db:"enabled"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// LoginRequest represents a login attempt
type LoginRequest struct {
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	MFACode     string `json:"mfa_code,omitempty"`
	IPAddress   string `json:"-"`
	UserAgent   string `json:"-"`
}

// LoginResponse represents the response after successful authentication
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	User         *UserInfo `json:"user"`
}

// UserInfo represents public user information
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	RoleID   string `json:"role_id,omitempty"`
}

// AuthContext holds authentication context information
type AuthContext struct {
	UserID     uuid.UUID
	Username   string
	Email      string
	Role       string
	Permissions []string
	APIKey     bool
}

// SetPassword sets the user's password using secure hashing
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// VerifyPassword verifies a password against the stored hash
func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(u.PasswordHash),
		[]byte(password),
	)
	return err == nil
}

// IsLocked checks if the user account is locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return u.LockedUntil.After(time.Now())
}

// Lock locks the user account until the specified time
func (u *User) Lock(until time.Time) {
	u.LockedUntil = &until
}

// Unlock unlocks the user account
func (u *User) Unlock() {
	u.LockedUntil = nil
}
