package models

import (
	"time"
)

// User represents a user in the access control system
type User struct {
	ID             string       `json:"id" db:"id"`
	Username       string       `json:"username" db:"username"`
	Email          string       `json:"email" db:"email"`
	PasswordHash   string       `json:"-" db:"password_hash"`
	Role           string       `json:"role" db:"role"`
	Permissions    []string     `json:"permissions" db:"permissions"`
	Groups         []string     `json:"groups" db:"groups"`
	Active         bool         `json:"active" db:"active"`
	MFEnabled      bool         `json:"mfa_enabled" db:"mfa_enabled"`
	MFSecret       string       `json:"-" db:"mfa_secret"`
	Locked         bool         `json:"locked" db:"locked"`
	LockedUntil    *time.Time   `json:"locked_until,omitempty" db:"locked_until"`
	FailedLogins   int          `json:"failed_logins" db:"failed_logins"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
	LastLogin      *time.Time   `json:"last_login,omitempty" db:"last_login"`
	PasswordChanged *time.Time  `json:"password_changed,omitempty" db:"password_changed"`
}

// Role represents a role in the RBAC system
type Role struct {
	ID           string     `json:"id" db:"id"`
	Name         string     `json:"name" db:"name"`
	Description  string     `json:"description" db:"description"`
	Permissions  []string   `json:"permissions" db:"permissions"`
	IsSystem     bool       `json:"is_system" db:"is_system"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Policy represents an access control policy
type Policy struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Effect      string                 `json:"effect" db:"effect"` // "allow" or "deny"
	Priority    int                    `json:"priority" db:"priority"`
	Active      bool                   `json:"active" db:"active"`
	Conditions  map[string]interface{} `json:"conditions" db:"conditions"`
	Subjects    PolicySubjects         `json:"subjects" db:"-"`
	Resources   []string               `json:"resources" db:"resources"`
	Actions     []string               `json:"actions" db:"actions"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// PolicySubjects defines which subjects a policy applies to
type PolicySubjects struct {
	Roles     []string `json:"roles,omitempty"`
	Users     []string `json:"users,omitempty"`
	Groups    []string `json:"groups,omitempty"`
	Services  []string `json:"services,omitempty"`
}

// Session represents an authenticated session
type Session struct {
	ID           string       `json:"id" db:"id"`
	UserID       string       `json:"user_id" db:"user_id"`
	Token        string       `json:"-" db:"token"`
	IPAddress    string       `json:"ip_address" db:"ip_address"`
	UserAgent    string       `json:"user_agent" db:"user_agent"`
	Location     string       `json:"location" db:"location"`
	ExpiresAt    time.Time    `json:"expires_at" db:"expires_at"`
	LastActivity time.Time    `json:"last_activity" db:"last_activity"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	Revoked      bool         `json:"revoked" db:"revoked"`
	RevokedAt    *time.Time   `json:"revoked_at,omitempty" db:"revoked_at"`
	RevokeReason string       `json:"revoke_reason,omitempty" db:"revoke_reason"`
}

// AccessToken represents a JWT access token
type AccessToken struct {
	TokenID     string   `json:"jti"`
	UserID      string   `json:"sub"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	Groups      []string `json:"groups"`
	SessionID   string   `json:"session_id"`
	IPAddress   string   `json:"ip_address"`
	IssuedAt    time.Time `json:"iat"`
	ExpiresAt   time.Time `json:"exp"`
	Issuer      string   `json:"iss"`
	Audience    string   `json:"aud"`
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	TokenID    string    `json:"jti"`
	UserID     string    `json:"sub"`
	SessionID  string    `json:"session_id"`
	ExpiresAt  time.Time `json:"exp"`
	IssuedAt   time.Time `json:"iat"`
	Used       bool      `json:"used"`
	Revoked    bool      `json:"revoked"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

// AuthenticationRequest represents a login request
type AuthenticationRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	MFACode     string `json:"mfa_code,omitempty"`
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	Location    string `json:"location"`
}

// AuthenticationResponse represents the response after successful authentication
type AuthenticationResponse struct {
	AccessToken    string       `json:"access_token"`
	RefreshToken   string       `json:"refresh_token"`
	TokenType      string       `json:"token_type"`
	ExpiresIn      int          `json:"expires_in"`
	User           UserSummary  `json:"user"`
	SessionID      string       `json:"session_id"`
}

// UserSummary represents a summary of user information
type UserSummary struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	Groups      []string `json:"groups"`
	MFEnabled   bool     `json:"mfa_enabled"`
}

// AuthorizationRequest represents an authorization check request
type AuthorizationRequest struct {
	UserID      string   `json:"user_id"`
	Resource    string   `json:"resource"`
	Action      string   `json:"action"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

// AuthorizationResponse represents the response of an authorization check
type AuthorizationResponse struct {
	Allowed   bool     `json:"allowed"`
	Reason    string   `json:"reason"`
	PolicyID  string   `json:"policy_id,omitempty"`
	DeniedBy  string   `json:"denied_by,omitempty"`
}

// AuditContext provides context for audit logging during authorization
type AuditContext struct {
	RequestID   string
	TraceID     string
	SessionID   string
	ActorID     string
	ActorType   string
	IPAddress   string
	UserAgent   string
	Location    string
	RequestPath string
	RequestMethod string
}
