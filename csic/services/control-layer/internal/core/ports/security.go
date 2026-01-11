package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
)

// Authenticator defines the interface for authentication operations.
// This interface provides methods for verifying user identities.
type Authenticator interface {
	// Authenticate authenticates a user
	Authenticate(ctx context.Context, credentials interface{}) (*domain.User, error)

	// AuthenticateToken authenticates using a token
	AuthenticateToken(ctx context.Context, token string) (*domain.User, error)

	// AuthenticateRequest authenticates an HTTP request
	AuthenticateRequest(ctx context.Context, request interface{}) (*domain.User, error)

	// GetAuthenticationMethod returns the authentication method
	GetAuthenticationMethod() string

	// SupportsMethod checks if a method is supported
	SupportsMethod(method string) bool
}

// AuthenticationService defines the interface for authentication service operations.
type AuthenticationService interface {
	// Login authenticates a user and returns tokens
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error)

	// Logout logs out a user
	Logout(ctx context.Context, req *domain.LogoutRequest) error

	// Refresh refreshes an access token
	Refresh(ctx context.Context, req *domain.RefreshRequest) (*domain.LoginResponse, error)

	// RevokeToken revokes a token
	RevokeToken(ctx context.Context, req *domain.RevokeRequest) error

	// Register registers a new user
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error)

	// VerifyEmail verifies a user's email
	VerifyEmail(ctx context.Context, req *domain.VerifyEmailRequest) error

	// ForgotPassword initiates password reset
	ForgotPassword(ctx context.Context, req *domain.ForgotPasswordRequest) error

	// ResetPassword resets a user's password
	ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest) error

	// ChangePassword changes a user's password
	ChangePassword(ctx context.Context, req *domain.ChangePasswordRequest) error

	// EnableMFA enables multi-factor authentication
	EnableMFA(ctx context.Context, req *domain.EnableMFARequest) (*domain.MFASetup, error)

	// VerifyMFA verifies multi-factor authentication
	VerifyMFA(ctx context.Context, req *domain.VerifyMFARequest) error

	// DisableMFA disables multi-factor authentication
	DisableMFA(ctx context.Context, req *domain.DisableMFARequest) error

	// GetMFAMethods gets available MFA methods
	GetMFAMethods(ctx context.Context) []*domain.MFAMethod

	// GetSession gets a session by ID
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)

	// ListSessions lists user sessions
	ListSessions(ctx context.Context, userID string) ([]*domain.Session, error)

	// TerminateSession terminates a session
	TerminateSession(ctx context.Context, sessionID string) error

	// TerminateAllSessions terminates all user sessions
	TerminateAllSessions(ctx context.Context, userID string) error
}

// Authorizer defines the interface for authorization operations.
type Authorizer interface {
	// Authorize checks if a user is authorized
	Authorize(ctx context.Context, user *domain.User, resource string, action string) error

	// AuthorizeRole checks if a user with a role is authorized
	AuthorizeRole(ctx context.Context, role string, resource string, action string) error

	// AuthorizeMultiple checks multiple permissions
	AuthorizeMultiple(ctx context.Context, user *domain.User, permissions []*domain.PermissionCheck) ([]*domain.PermissionCheck, error)

	// GetPermissions gets permissions for a user
	GetPermissions(ctx context.Context, userID string) ([]*domain.Permission, error)

	// GetRoles gets roles for a user
	GetRoles(ctx context.Context, userID string) ([]*domain.Role, error)

	// HasRole checks if a user has a role
	HasRole(ctx context.Context, userID string, role string) (bool, error)

	// HasPermission checks if a user has a permission
	HasPermission(ctx context.Context, userID string, resource string, action string) (bool, error)
}

// AuthorizationService defines the interface for authorization service operations.
type AuthorizationService interface {
	// CreateRole creates a new role
	CreateRole(ctx context.Context, role *domain.Role) error

	// UpdateRole updates a role
	UpdateRole(ctx context.Context, id string, role *domain.Role) error

	// DeleteRole deletes a role
	DeleteRole(ctx context.Context, id string) error

	// GetRole gets a role by ID
	GetRole(ctx context.Context, id string) (*domain.Role, error)

	// GetRoleByName gets a role by name
	GetRoleByName(ctx context.Context, name string) (*domain.Role, error)

	// ListRoles lists all roles
	ListRoles(ctx context.Context, filter *domain.RoleFilter) ([]*domain.Role, error)

	// AssignRole assigns a role to a user
	AssignRole(ctx context.Context, userID string, roleID string) error

	// RemoveRole removes a role from a user
	RemoveRole(ctx context.Context, userID string, roleID string) error

	// BulkAssignRoles assigns roles to multiple users
	BulkAssignRoles(ctx context.Context, userIDs []string, roleID string) error

	// CreatePermission creates a new permission
	CreatePermission(ctx context.Context, permission *domain.Permission) error

	// UpdatePermission updates a permission
	UpdatePermission(ctx context.Context, id string, permission *domain.Permission) error

	// DeletePermission deletes a permission
	DeletePermission(ctx context.Context, id string) error

	// GetPermission gets a permission by ID
	GetPermission(ctx context.Context, id string) (*domain.Permission, error)

	// ListPermissions lists all permissions
	ListPermissions(ctx context.Context, filter *domain.PermissionFilter) ([]*domain.Permission, error)

	// AssignPermission assigns a permission to a role
	AssignPermission(ctx context.Context, roleID string, permissionID string) error

	// RemovePermission removes a permission from a role
	RemovePermission(ctx context.Context, roleID string, permissionID string) error

	// CreatePolicy creates an authorization policy
	CreatePolicy(ctx context.Context, policy *domain.AuthorizationPolicy) error

	// UpdatePolicy updates an authorization policy
	UpdatePolicy(ctx context.Context, id string, policy *domain.AuthorizationPolicy) error

	// DeletePolicy deletes an authorization policy
	DeletePolicy(ctx context.Context, id string) error

	// GetPolicy gets an authorization policy
	GetPolicy(ctx context.Context, id string) (*domain.AuthorizationPolicy, error)

	// ListPolicies lists all authorization policies
	ListPolicies(ctx context.Context) ([]*domain.AuthorizationPolicy, error)
}

// TokenService defines the interface for token operations.
type TokenService interface {
	// GenerateAccessToken generates an access token
	GenerateAccessToken(ctx context.Context, user *domain.User) (string, error)

	// GenerateRefreshToken generates a refresh token
	GenerateRefreshToken(ctx context.Context, user *domain.User) (string, error)

	// GenerateIDToken generates an ID token
	GenerateIDToken(ctx context.Context, user *domain.User) (string, error)

	// ValidateAccessToken validates an access token
	ValidateAccessToken(ctx context.Context, token string) (*domain.TokenClaims, error)

	// ValidateRefreshToken validates a refresh token
	ValidateRefreshToken(ctx context.Context, token string) (*domain.TokenClaims, error)

	// ValidateIDToken validates an ID token
	ValidateIDToken(ctx context.Context, token string) (*domain.TokenClaims, error)

	// RevokeToken revokes a token
	RevokeToken(ctx context.Context, token string) error

	// RevokeAllTokens revokes all tokens for a user
	RevokeAllTokens(ctx context.Context, userID string) error

	// GetTokenInfo gets token information
	GetTokenInfo(ctx context.Context, token string) (*domain.TokenInfo, error)

	// GetActiveTokens gets active tokens for a user
	GetActiveTokens(ctx context.Context, userID string) ([]*domain.TokenInfo, error)
}

// TokenGenerator defines the interface for token generation.
type TokenGenerator interface {
	// GenerateToken generates a token
	GenerateToken(ctx context.Context, claims map[string]interface{}) (string, error)

	// GenerateTokenWithExpiry generates a token with expiry
	GenerateTokenWithExpiry(ctx context.Context, claims map[string]interface{}, expiry time.Duration) (string, error)

	// ValidateToken validates a token
	ValidateToken(ctx context.Context, token string) (map[string]interface{}, error)

	// GetTokenExpiry gets the token expiry
	GetTokenExpiry(ctx context.Context, token string) (time.Time, error)
}

// EncryptionService defines the interface for encryption operations.
type EncryptionService interface {
	// Encrypt encrypts data
	Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)

	// Decrypt decrypts data
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)

	// EncryptString encrypts a string
	EncryptString(ctx context.Context, plaintext string) (string, error)

	// DecryptString decrypts a string
	DecryptString(ctx context.Context, ciphertext string) (string, error)

	// GetEncryptionInfo returns encryption information
	GetEncryptionInfo(ctx context.Context) (*domain.EncryptionInfo, error)

	// GetAlgorithm returns the encryption algorithm
	GetAlgorithm() string

	// GetKeySize returns the key size
	GetKeySize() int
}

// Hasher defines the interface for hashing operations.
type Hasher interface {
	// Hash hashes data
	Hash(ctx context.Context, data []byte) ([]byte, error)

	// HashString hashes a string
	HashString(ctx context.Context, data string) (string, error)

	// Verify verifies hashed data
	Verify(ctx context.Context, data []byte, hash []byte) (bool, error)

	// VerifyString verifies a hashed string
	VerifyString(ctx context.Context, data string, hash string) (bool, error)

	// NeedsRehash checks if a hash needs rehashing
	NeedsRehash(ctx context.Context, hash []byte) (bool, error)

	// GetInfo returns hash information
	GetInfo(ctx context.Context) (*domain.HashInfo, error)
}

// KeyManager defines the interface for key management operations.
type KeyManager interface {
	// GenerateKey generates a new key
	GenerateKey(ctx context.Context, keyType string) (*domain.Key, error)

	// GetKey gets a key by ID
	GetKey(ctx context.Context, keyID string) (*domain.Key, error)

	// GetPublicKey gets a public key
	GetPublicKey(ctx context.Context, keyID string) ([]byte, error)

	// ImportKey imports a key
	ImportKey(ctx context.Context, key *domain.Key) error

	// ExportKey exports a key
	ExportKey(ctx context.Context, keyID string, purpose string) ([]byte, error)

	// DeleteKey deletes a key
	DeleteKey(ctx context.Context, keyID string) error

	// RotateKey rotates a key
	RotateKey(ctx context.Context, keyID string) (*domain.Key, error)

	// ListKeys lists all keys
	ListKeys(ctx context.Context, filter *domain.KeyFilter) ([]*domain.Key, error)

	// GetKeyVersion gets a key version
	GetKeyVersion(ctx context.Context, keyID string, versionID string) (*domain.KeyVersion, error)

	// CreateKeyVersion creates a new key version
	CreateKeyVersion(ctx context.Context, keyID string) (*domain.KeyVersion, error)

	// DisableKey disables a key
	DisableKey(ctx context.Context, keyID string) error

	// EnableKey enables a key
	EnableKey(ctx context.Context, keyID string) error
}

// KeyStore defines the interface for secure key storage.
type KeyStore interface {
	// StoreKey stores a key
	StoreKey(ctx context.Context, key []byte, metadata *domain.KeyMetadata) (string, error)

	// RetrieveKey retrieves a key
	RetrieveKey(ctx context.Context, keyID string) ([]byte, error)

	// DeleteKey deletes a key
	DeleteKey(ctx context.Context, keyID string) error

	// ListKeys lists key metadata
	ListKeys(ctx context.Context) ([]*domain.KeyMetadata, error)

	// GetKeyInfo gets key metadata
	GetKeyInfo(ctx context.Context, keyID string) (*domain.KeyMetadata, error)

	// RotateMasterKey rotates the master key
	RotateMasterKey(ctx context.Context) error
}

// CertificateManager defines the interface for certificate management.
type CertificateManager interface {
	// GenerateCertificate generates a certificate
	GenerateCertificate(ctx context.Context, req *domain.CertificateRequest) (*domain.Certificate, error)

	// GetCertificate gets a certificate by ID
	GetCertificate(ctx context.Context, certID string) (*domain.Certificate, error)

	// GetCertificateByFingerprint gets a certificate by fingerprint
	GetCertificateByFingerprint(ctx context.Context, fingerprint string) (*domain.Certificate, error)

	// ImportCertificate imports a certificate
	ImportCertificate(ctx context.Context, cert *domain.Certificate) error

	// ExportCertificate exports a certificate
	ExportCertificate(ctx context.Context, certID string, format string) ([]byte, error)

	// DeleteCertificate deletes a certificate
	DeleteCertificate(ctx context.Context, certID string) error

	// RevokeCertificate revokes a certificate
	RevokeCertificate(ctx context.Context, certID string, reason string) error

	// ListCertificates lists all certificates
	ListCertificates(ctx context.Context, filter *domain.CertificateFilter) ([]*domain.Certificate, error)

	// GetCertificateChain gets the certificate chain
	GetCertificateChain(ctx context.Context, certID string) ([]*domain.Certificate, error)

	// SignCSR signs a certificate signing request
	SignCSR(ctx context.Context, csr []byte, validity time.Duration) ([]byte, error)

	// GetCA returns the certificate authority
	GetCA(ctx context.Context) (*domain.Certificate, error)

	// GetCRL gets the certificate revocation list
	GetCRL(ctx context.Context) ([]byte, error)
}

// SSOManager defines the interface for single sign-on operations.
type SSOManager interface {
	// InitiateLogin initiates SSO login
	InitiateLogin(ctx context.Context, provider string) (string, error)

	// HandleCallback handles SSO callback
	HandleCallback(ctx context.Context, provider string, callbackParams map[string]string) (*domain.SSOResult, error)

	// GetProviders gets available SSO providers
	GetProviders(ctx context.Context) []*domain.SSOProvider

	// RegisterProvider registers an SSO provider
	RegisterProvider(ctx context.Context, provider *domain.SSOProvider) error

	// UpdateProvider updates an SSO provider
	UpdateProvider(ctx context.Context, id string, provider *domain.SSOProvider) error

	// DeleteProvider deletes an SSO provider
	DeleteProvider(ctx context.Context, id string) error

	// GetUserInfo gets user info from provider
	GetUserInfo(ctx context.Context, provider string, accessToken string) (*domain.SSOUserInfo, error)

	// DisconnectProvider disconnects an SSO provider for a user
	DisconnectProvider(ctx context.Context, userID string, provider string) error
}

// OAuthManager defines the interface for OAuth operations.
type OAuthManager interface {
	// Authorize authorizes an OAuth request
	Authorize(ctx context.Context, req *domain.OAuthAuthorizeRequest) (string, error)

	// Token exchanges an authorization code for tokens
	Token(ctx context.Context, req *domain.OAuthTokenRequest) (*domain.OAuthTokenResponse, error)

	// Introspect introspects a token
	Introspect(ctx context.Context, token string) (*domain.TokenIntrospection, error)

	// Revoke revokes a token
	Revoke(ctx context.Context, token string, tokenTypeHint string) error

	// RegisterClient registers an OAuth client
	RegisterClient(ctx context.Context, client *domain.OAuthClient) error

	// UpdateClient updates an OAuth client
	UpdateClient(ctx context.Context, clientID string, client *domain.OAuthClient) error

	// DeleteClient deletes an OAuth client
	DeleteClient(ctx context.Context, clientID string) error

	// GetClient gets an OAuth client
	GetClient(ctx context.Context, clientID string) (*domain.OAuthClient, error)

	// ListClients lists OAuth clients
	ListClients(ctx context.Context) ([]*domain.OAuthClient, error)

	// RotateClientSecret rotates a client secret
	RotateClientSecret(ctx context.Context, clientID string) (string, error)
}

// SessionManager defines the interface for session management.
type SessionManager interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, userID string, data map[string]interface{}) (*domain.Session, error)

	// GetSession gets a session by ID
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)

	// UpdateSession updates a session
	UpdateSession(ctx context.Context, sessionID string, data map[string]interface{}) error

	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, sessionID string) error

	// DeleteUserSessions deletes all sessions for a user
	DeleteUserSessions(ctx context.Context, userID string) error

	// RefreshSession refreshes a session
	RefreshSession(ctx context.Context, sessionID string) error

	// ValidateSession validates a session
	ValidateSession(ctx context.Context, sessionID string) (bool, error)

	// GetUserSessions gets all sessions for a user
	GetUserSessions(ctx context.Context, userID string) ([]*domain.Session, error)

	// CountUserSessions counts sessions for a user
	CountUserSessions(ctx context.Context, userID string) (int, error)
}

// SecurityPolicyManager defines the interface for security policy management.
type SecurityPolicyManager interface {
	// CreatePolicy creates a security policy
	CreatePolicy(ctx context.Context, policy *domain.SecurityPolicy) error

	// UpdatePolicy updates a security policy
	UpdatePolicy(ctx context.Context, id string, policy *domain.SecurityPolicy) error

	// DeletePolicy deletes a security policy
	DeletePolicy(ctx context.Context, id string) error

	// GetPolicy gets a security policy
	GetPolicy(ctx context.Context, id string) (*domain.SecurityPolicy, error)

	// ListPolicies lists all security policies
	ListPolicies(ctx context.Context, filter *domain.SecurityPolicyFilter) ([]*domain.SecurityPolicy, error)

	// ActivatePolicy activates a security policy
	ActivatePolicy(ctx context.Context, id string) error

	// DeactivatePolicy deactivates a security policy
	DeactivatePolicy(ctx context.Context, id string) error

	// EvaluatePolicy evaluates a security policy
	EvaluatePolicy(ctx context.Context, policyID string, context map[string]interface{}) (*domain.PolicyEvaluation, error)

	// GetPolicyStatus gets the policy status
	GetPolicyStatus(ctx context.Context, id string) (*domain.PolicyStatus, error)
}

// AuditLogger defines the interface for security audit logging.
type AuditLogger interface {
	// Log logs an audit event
	Log(ctx context.Context, event *domain.AuditEvent) error

	// LogBatch logs multiple audit events
	LogBatch(ctx context.Context, events []*domain.AuditEvent) error

	// GetAuditTrail gets the audit trail
	GetAuditTrail(ctx context.Context, query *domain.AuditQuery) ([]*domain.AuditEvent, error)

	// GetUserActivity gets user activity
	GetUserActivity(ctx context.Context, userID string, startTime, endTime time.Time) ([]*domain.AuditEvent, error)

	// GetSecurityEvents gets security events
	GetSecurityEvents(ctx context.Context, query *domain.SecurityEventQuery) ([]*domain.AuditEvent, error)

	// GetLoginHistory gets login history
	GetLoginHistory(ctx context.Context, userID string, limit int) ([]*domain.AuditEvent, error)

	// ExportAuditLog exports the audit log
	ExportAuditLog(ctx context.Context, query *domain.AuditQuery, format string) ([]byte, error)
}

// SecurityScanner defines the interface for security scanning.
type SecurityScanner interface {
	// ScanVulnerabilities scans for vulnerabilities
	ScanVulnerabilities(ctx context.Context, target string) ([]*domain.Vulnerability, error)

	// ScanDependencies scans dependencies for vulnerabilities
	ScanDependencies(ctx context.Context) ([]*domain.Vulnerability, error)

	// ScanSecrets scans for secrets
	ScanSecrets(ctx context.Context, content []byte) ([]*domain.SecretFinding, error)

	// ScanCode scans code for security issues
	ScanCode(ctx context.Context, repoURL string) ([]*domain.CodeFinding, error)

	// GetVulnerabilities gets vulnerabilities
	GetVulnerabilities(ctx context.Context, query *domain.VulnerabilityQuery) ([]*domain.Vulnerability, error)

	// GetVulnerability gets a vulnerability
	GetVulnerability(ctx context.Context, id string) (*domain.Vulnerability, error)

	// UpdateVulnerabilityStatus updates vulnerability status
	UpdateVulnerabilityStatus(ctx context.Context, id string, status string) error
}

// IntrusionDetection defines the interface for intrusion detection operations.
type IntrusionDetection interface {
	// DetectIntrusion detects intrusions
	DetectIntrusion(ctx context.Context, event *domain.SecurityEvent) (*domain.IntrusionDetectionResult, error)

	// GetThreats gets detected threats
	GetThreats(ctx context.Context, query *domain.ThreatQuery) ([]*domain.Threat, error)

	// GetThreat gets a threat
	GetThreat(ctx context.Context, id string) (*domain.Threat, error)

	// UpdateThreatStatus updates threat status
	UpdateThreatStatus(ctx context.Context, id string, status string) error

	// BlockIP blocks an IP address
	BlockIP(ctx context.Context, ip string, reason string, duration time.Duration) error

	// UnblockIP unblocks an IP address
	UnblockIP(ctx context.Context, ip string) error

	// GetBlockedIPs gets blocked IPs
	GetBlockedIPs(ctx context.Context) ([]*domain.BlockedIP, error)

	// DetectAnomaly detects anomalies
	DetectAnomaly(ctx context.Context, event *domain.AnomalyEvent) (*domain.AnomalyDetectionResult, error)
}

// RateLimiter defines the interface for rate limiting operations.
type RateLimiter interface {
	// Allow checks if an action is allowed
	Allow(ctx context.Context, identifier string, operation string) (bool, error)

	// Consume consumes a rate limit token
	Consume(ctx context.Context, identifier string, operation string) (bool, error)

	// GetLimit gets the rate limit
	GetLimit(ctx context.Context, identifier string, operation string) (*domain.RateLimit, error)

	// SetLimit sets the rate limit
	SetLimit(ctx context.Context, identifier string, operation string, limit *domain.RateLimit) error

	// ResetLimit resets the rate limit
	ResetLimit(ctx context.Context, identifier string, operation string) error

	// GetRateLimitStats gets rate limit statistics
	GetRateLimitStats(ctx context.Context) (*domain.RateLimitStats, error)
}

// WAFManager defines the interface for web application firewall operations.
type WAFManager interface {
	// CheckRequest checks a request against WAF rules
	CheckRequest(ctx context.Context, request *domain.WAFRequest) (*domain.WAFResult, error)

	// GetRules gets WAF rules
	GetRules(ctx context.Context, filter *domain.WAFRuleFilter) ([]*domain.WAFRule, error)

	// CreateRule creates a WAF rule
	CreateRule(ctx context.Context, rule *domain.WAFRule) error

	// UpdateRule updates a WAF rule
	UpdateRule(ctx context.Context, id string, rule *domain.WAFRule) error

	// DeleteRule deletes a WAF rule
	DeleteRule(ctx context.Context, id string) error

	// EnableRule enables a WAF rule
	EnableRule(ctx context.Context, id string) error

	// DisableRule disables a WAF rule
	DisableRule(ctx context.Context, id string) error

	// GetProfiles gets WAF profiles
	GetProfiles(ctx context.Context) ([]*domain.WAFProfile, error)

	// CreateProfile creates a WAF profile
	CreateProfile(ctx context.Context, profile *domain.WAFProfile) error

	// ApplyProfile applies a WAF profile
	ApplyProfile(ctx context.Context, profileID string) error
}
