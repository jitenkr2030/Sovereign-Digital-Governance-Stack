package ports

import (
	"context"

	"github.com/csic-platform/services/security/iam/internal/core/domain"
)

// TokenGenerator defines the interface for token generation
type TokenGenerator interface {
	// GenerateAccessToken generates an access token for a user
	GenerateAccessToken(ctx context.Context, user *domain.User) (string, error)

	// GenerateRefreshToken generates a refresh token for a user
	GenerateRefreshToken(ctx context.Context, user *domain.User) (string, error)

	// ValidateAccessToken validates an access token
	ValidateAccessToken(ctx context.Context, token string) (*domain.AuthContext, error)

	// GetAccessTokenExpiry returns the access token expiry duration
	GetAccessTokenExpiry() int
}

// TokenBlacklist defines the interface for token blacklist management
type TokenBlacklist interface {
	// Add adds a token to the blacklist
	Add(ctx context.Context, token string, expiry int64) error

	// IsBlacklisted checks if a token is blacklisted
	IsBlacklisted(ctx context.Context, token string) (bool, error)

	// Cleanup removes all expired tokens from the blacklist
	Cleanup(ctx context.Context) error
}

// PasswordHasher defines the interface for password hashing
type PasswordHasher interface {
	// Hash generates a hash for the given password
	Hash(ctx context.Context, password string) (string, error)

	// Verify compares a password with a hash
	Verify(ctx context.Context, password, hash string) bool
}

// MFAGenerator defines the interface for MFA code generation
type MFAGenerator interface {
	// GenerateSecret generates a new MFA secret
	GenerateSecret(ctx context.Context) (string, error)

	// GenerateQRCode generates a QR code for MFA setup
	GenerateQRCode(ctx context.Context, issuer, account, secret string) (string, error)

	// GenerateCode generates a TOTP code
	GenerateCode(ctx context.Context, secret string) (string, error)

	// VerifyCode verifies a TOTP code
	VerifyCode(ctx context.Context, secret, code string) bool

	// GenerateBackupCodes generates backup codes
	GenerateBackupCodes(ctx context.Context, count int) ([]string, error)
}

// EmailService defines the interface for sending emails
type EmailService interface {
	// SendEmail sends an email
	SendEmail(ctx context.Context, req *EmailRequest) error

	// SendPasswordResetEmail sends a password reset email
	SendPasswordResetEmail(ctx context.Context, to, resetURL string) error

	// SendWelcomeEmail sends a welcome email
	SendWelcomeEmail(ctx context.Context, to, username string) error
}

// EmailRequest represents an email request
type EmailRequest struct {
	To        []string `json:"to"`
	Subject   string   `json:"subject"`
	Body      string   `json:"body"`
	BodyHTML  string   `json:"body_html,omitempty"`
}

// MetricsCollector defines the interface for metrics collection
type MetricsCollector interface {
	// IncrementLoginAttempts increments the login attempts counter
	IncrementLoginAttempts(success bool)

	// IncrementUsersCreated increments the users created counter
	IncrementUsersCreated()

	// IncrementSessionsCreated increments the sessions created counter
	IncrementSessionsCreated()

	// IncrementSessionsRevoked increments the sessions revoked counter
	IncrementSessionsRevoked()

	// RecordAuthLatency records the authentication latency
	RecordAuthLatency(duration string)

	// Close closes the metrics collector
	Close() error
}

// HealthChecker defines the interface for health checks
type HealthChecker interface {
	// Check checks the health of the service
	Check(ctx context.Context) (*HealthStatus, error)
}

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status    string            `json:"status"`
	Checks    map[string]string `json:"checks"`
	Timestamp string            `json:"timestamp"`
}
