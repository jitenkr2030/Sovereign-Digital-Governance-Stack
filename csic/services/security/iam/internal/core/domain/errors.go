package domain

import "errors"

// Common error definitions for the IAM domain

var (
	// Authentication errors
	ErrInvalidCredentials   = errors.New("invalid username or password")
	ErrAccountLocked        = errors.New("account is locked")
	ErrAccountDisabled      = errors.New("account is disabled")
	ErrTokenExpired         = errors.New("token has expired")
	ErrTokenInvalid         = errors.New("invalid token")
	ErrTokenBlacklisted     = errors.New("token has been revoked")
	ErrMFACodeInvalid       = errors.New("invalid MFA code")
	ErrMFACodeExpired       = errors.New("MFA code has expired")
	ErrMFANotEnabled        = errors.New("MFA is not enabled")
	ErrMFANotVerified       = errors.New("MFA verification required")
	ErrBackupCodeUsed       = errors.New("backup code has already been used")
	ErrSessionExpired       = errors.New("session has expired")
	ErrSessionRevoked       = sessionRevokedError{}

	// User errors
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrUserInvalidInput     = errors.New("invalid user input")
	ErrUserCannotDeleteSelf = errors.New("cannot delete your own account")
	ErrUserCannotDisableSelf = errors.New("cannot disable your own account")

	// Role errors
	ErrRoleNotFound         = errors.New("role not found")
	ErrRoleAlreadyExists    = errors.New("role already exists")
	ErrRoleInUse            = errors.New("role is assigned to users")
	ErrRoleCannotDelete     = errors.New("cannot delete system role")

	// Password errors
	ErrPasswordWeak         = errors.New("password does not meet complexity requirements")
	ErrPasswordReused       = errors.New("password has been used recently")
	ErrPasswordSame         = errors.New("new password must be different from current password")

	// API key errors
	ErrAPIKeyNotFound       = errors.New("API key not found")
	ErrAPIKeyExpired        = errors.New("API key has expired")
	ErrAPIKeyDisabled       = errors.New("API key has been disabled")
	ErrAPIKeyInvalid        = errors.New("invalid API key")

	// Session errors
	ErrSessionNotFound      = errors.New("session not found")

	// Permission errors
	ErrPermissionDenied     = errors.New("permission denied")
	ErrInsufficientScope    = errors.New("insufficient scope")

	// Repository errors
	ErrRepositoryNotFound   = errors.New("record not found in repository")
	ErrRepositoryConflict   = errors.New("record conflict in repository")
	ErrRepositoryUnavailable = errors.New("repository unavailable")
)

type sessionRevokedError struct{}

func (e sessionRevokedError) Error() string {
	return "session has been revoked"
}

// IsSessionRevoked checks if the error is a session revoked error
func IsSessionRevoked(err error) bool {
	_, ok := err.(sessionRevokedError)
	return ok
}
