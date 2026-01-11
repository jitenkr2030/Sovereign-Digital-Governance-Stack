// CSIC Platform - Security Module (IAM)
// Identity and Access Management for the government-grade platform

package iam

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/csic-platform/shared/config"
	"github.com/csic-platform/shared/errors"
)

// User represents a system user
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	Permissions  []string  `json:"permissions"`
	Status       string    `json:"status"` // ACTIVE, INACTIVE, SUSPENDED
	MFAEnabled   bool      `json:"mfa_enabled"`
	MFASecret    string    `json:"-"`
	LastLogin    time.Time `json:"last_login"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	FailedLogins int       `json:"-"`
	LockedUntil  time.Time `json:"-"`
}

// Session represents an active user session
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastActivity time.Time `json:"last_activity"`
}

// Role represents a user role with permissions
type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a system permission
type Permission struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// MFAConfig contains MFA configuration
type MFAConfig struct {
	Enabled     bool   `json:"enabled"`
	Issuer      string `json:"issuer"`
	Window      int    `json:"window"` // Time window for TOTP verification
	BackupCodes int    `json:"backup_codes"`
}

// Service provides IAM functionality
type Service struct {
	cfg       *config.SecurityConfig
	users     map[string]*User
	sessions  map[string]*Session
	roles     map[string]*Role
	permissions map[string]*Permission
	mfaConfig *MFAConfig
	mu        sync.RWMutex
}

// NewService creates a new IAM service
func NewService(cfg *config.SecurityConfig) *Service {
	svc := &Service{
		cfg:          cfg,
		users:        make(map[string]*User),
		sessions:     make(map[string]*Session),
		roles:        make(map[string]*Role),
		permissions:  make(map[string]*Permission),
		mfaConfig:    &MFAConfig{
			Enabled:     cfg.MFA.Enabled,
			Issuer:      cfg.MFA.Issuer,
			Window:      cfg.MFA.Window,
			BackupCodes: cfg.MFA.BackupCodes,
		},
	}

	// Initialize default roles and permissions
	svc.initializeDefaults()

	return svc
}

// initializeDefaults sets up default roles and permissions
func (s *Service) initializeDefaults() {
	// Define permissions
	permissions := []Permission{
		// Dashboard permissions
		{"view:dashboard", "View Dashboard", "Access to dashboard overview", "dashboard", "view"},
		
		// Alert permissions
		{"view:alerts", "View Alerts", "View system alerts", "alerts", "view"},
		{"manage:alerts", "Manage Alerts", "Acknowledge and resolve alerts", "alerts", "manage"},
		
		// Exchange permissions
		{"view:exchanges", "View Exchanges", "View exchange list and details", "exchanges", "view"},
		{"manage:exchanges", "Manage Exchanges", "Suspend/revoke exchanges", "exchanges", "manage"},
		
		// Wallet permissions
		{"view:wallets", "View Wallets", "View wallet list and details", "wallets", "view"},
		{"manage:wallets", "Manage Wallets", "Freeze/blacklist wallets", "wallets", "manage"},
		
		// Miner permissions
		{"view:miners", "View Miners", "View miner list and details", "miners", "view"},
		{"manage:miners", "Manage Miners", "Suspend miners", "miners", "manage"},
		
		// Transaction permissions
		{"view:transactions", "View Transactions", "View transaction history", "transactions", "view"},
		{"manage:transactions", "Manage Transactions", "Flag/block transactions", "transactions", "manage"},
		
		// Report permissions
		{"view:reports", "View Reports", "View generated reports", "reports", "view"},
		{"create:reports", "Create Reports", "Generate new reports", "reports", "create"},
		
		// Compliance permissions
		{"view:compliance", "View Compliance", "View compliance status", "compliance", "view"},
		{"manage:compliance", "Manage Compliance", "Manage compliance obligations", "compliance", "manage"},
		
		// Audit permissions
		{"view:audit", "View Audit Logs", "View system audit logs", "audit", "view"},
		
		// Settings permissions
		{"view:settings", "View Settings", "View system settings", "settings", "view"},
		{"manage:settings", "Manage Settings", "Modify system settings", "settings", "manage"},
		
		// User management
		{"view:users", "View Users", "View user list", "users", "view"},
		{"manage:users", "Manage Users", "Create/update/delete users", "users", "manage"},
	}

	for _, perm := range permissions {
		s.permissions[perm.Code] = &perm
	}

	// Define roles
	roles := []Role{
		{
			Name:        "ADMIN",
			Description: "System Administrator with full access",
			Permissions: []string{"*"}, // Wildcard for all permissions
		},
		{
			Name:        "REGULATOR",
			Description: "Regulator with operational access",
			Permissions: []string{
				"view:dashboard", "view:alerts", "manage:alerts",
				"view:exchanges", "manage:exchanges",
				"view:wallets", "manage:wallets",
				"view:miners", "manage:miners",
				"view:transactions", "manage:transactions",
				"view:reports", "create:reports",
				"view:compliance", "manage:compliance",
				"view:audit",
			},
		},
		{
			Name:        "AUDITOR",
			Description: "Compliance Auditor",
			Permissions: []string{
				"view:dashboard",
				"view:alerts",
				"view:exchanges", "view:wallets", "view:miners",
				"view:transactions",
				"view:reports", "create:reports",
				"view:compliance",
				"view:audit",
			},
		},
		{
			Name:        "ANALYST",
			Description: "Data Analyst with read-only access",
			Permissions: []string{
				"view:dashboard",
				"view:alerts",
				"view:exchanges", "view:wallets", "view:miners",
				"view:transactions",
				"view:reports",
				"view:compliance",
			},
		},
	}

	for _, role := range roles {
		s.roles[role.Name] = &role
	}

	// Create default admin user
	adminPassword, _ := s.HashPassword("admin123")
	s.users["admin"] = &User{
		ID:           "usr_001",
		Username:     "admin",
		Email:        "admin@csic.gov",
		PasswordHash: adminPassword,
		Role:         "ADMIN",
		Permissions:  s.roles["ADMIN"].Permissions,
		Status:       "ACTIVE",
		MFAEnabled:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// Authenticate validates user credentials
func (s *Service) Authenticate(username, password, ipAddress, userAgent string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[username]
	if !exists {
		return nil, errors.New("invalid credentials")
	}

	// Check if account is locked
	if time.Now().Before(user.LockedUntil) {
		return nil, errors.New("account locked")
	}

	// Verify password
	if !s.VerifyPassword(password, user.PasswordHash) {
		user.FailedLogins++
		
		// Lock account after 5 failed attempts
		if user.FailedLogins >= 5 {
			user.LockedUntil = time.Now().Add(30 * time.Minute)
			log.Printf("User %s locked due to failed login attempts", username)
		}
		
		s.users[username] = user
		return nil, errors.New("invalid credentials")
	}

	// Check if MFA is required
	if user.MFAEnabled {
		// In production, MFA verification would happen here
		return nil, errors.New("mfa_required")
	}

	// Reset failed login count on success
	user.FailedLogins = 0
	user.LockedUntil = time.Time{}
	user.LastLogin = time.Now()
	s.users[username] = user

	// Create session
	session := s.createSession(user, ipAddress, userAgent)
	s.sessions[session.Token] = session

	return session, nil
}

// CreateSession creates a new session for a user
func (s *Service) createSession(user *User, ipAddress, userAgent string) *Session {
	token := s.generateSecureToken(32)
	refreshToken := s.generateSecureToken(32)

	session := &Session{
		ID:           fmt.Sprintf("sess_%d", time.Now().UnixNano()),
		UserID:       user.ID,
		Token:        token,
		RefreshToken: refreshToken,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(s.cfg.Session.IdleTimeout) * time.Minute),
		LastActivity: time.Now(),
	}

	return session
}

// ValidateSession validates a session token
func (s *Service) ValidateSession(token string) (*Session, *User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[token]
	if !exists {
		return nil, nil, errors.New("invalid session")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, token)
		return nil, nil, errors.New("session expired")
	}

	user, exists := s.users["admin"] // Simplified - would lookup by UserID
	if !exists || user.Status != "ACTIVE" {
		return nil, nil, errors.New("user not found or inactive")
	}

	return session, user, nil
}

// RefreshSession refreshes a session with a refresh token
func (s *Service) RefreshSession(refreshToken, ipAddress, userAgent string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find session with matching refresh token
	var oldSession *Session
	for _, session := range s.sessions {
		if session.RefreshToken == refreshToken {
			oldSession = session
			break
		}
	}

	if oldSession == nil {
		return nil, errors.New("invalid refresh token")
	}

	// Delete old session
	delete(s.sessions, oldSession.Token)

	// Get user
	user := s.users["admin"] // Simplified
	if user == nil || user.Status != "ACTIVE" {
		return nil, errors.New("user not found or inactive")
	}

	// Create new session
	session := s.createSession(user, ipAddress, userAgent)
	s.sessions[session.Token] = session

	return session, nil
}

// InvalidateSession removes a session
func (s *Service) InvalidateSession(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, token)
	return nil
}

// InvalidateUserSessions removes all sessions for a user
func (s *Service) InvalidateUserSessions(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for token, session := range s.sessions {
		if session.UserID == userID {
			delete(s.sessions, token)
		}
	}
	return nil
}

// HasPermission checks if a user has a specific permission
func (s *Service) HasPermission(user *User, permission string) bool {
	// Admin has all permissions
	for _, perm := range user.Permissions {
		if perm == "*" {
			return true
		}
		if perm == permission {
			return true
		}
	}

	return false
}

// HasAnyPermission checks if a user has any of the specified permissions
func (s *Service) HasAnyPermission(user *User, permissions ...string) bool {
	for _, permission := range permissions {
		if s.HasPermission(user, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a user has all of the specified permissions
func (s *Service) HasAllPermissions(user *User, permissions ...string) bool {
	for _, permission := range permissions {
		if !s.HasPermission(user, permission) {
			return false
		}
	}
	return true
}

// CreateUser creates a new user
func (s *Service) CreateUser(username, email, password, role string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already exists
	if _, exists := s.users[username]; exists {
		return nil, errors.New("user already exists")
	}

	// Validate role
	roleConfig, exists := s.roles[role]
	if !exists {
		return nil, errors.New("invalid role")
	}

	// Hash password
	passwordHash, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user		ID:           fmt.Sprintf("usr := &User{
_%d", time.Now().UnixNano()),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		Permissions:  roleConfig.Permissions,
		Status:       "ACTIVE",
		MFAEnabled:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	s.users[username] = user
	return user, nil
}

// UpdateUser updates an existing user
func (s *Service) UpdateUser(username string, updates map[string]interface{}) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[username]
	if !exists {
		return nil, errors.New("user not found")
	}

	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}
	if role, ok := updates["role"].(string); ok {
		if roleConfig, exists := s.roles[role]; exists {
			user.Role = role
			user.Permissions = roleConfig.Permissions
		}
	}
	if status, ok := updates["status"].(string); ok {
		user.Status = status
	}

	user.UpdatedAt = time.Now()
	s.users[username] = user

	return user, nil
}

// ChangePassword changes a user's password
func (s *Service) ChangePassword(username, currentPassword, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[username]
	if !exists {
		return errors.New("user not found")
	}

	// Verify current password
	if !s.VerifyPassword(currentPassword, user.PasswordHash) {
		return errors.New("invalid current password")
	}

	// Hash new password
	newHash, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	user.UpdatedAt = time.Now()
	s.users[username] = user

	// Invalidate all sessions for security
	s.InvalidateUserSessions(user.ID)

	return nil
}

// ResetPassword resets a user's password (admin function)
func (s *Service) ResetPassword(username string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[username]
	if !exists {
		return "", errors.New("user not found")
	}

	// Generate new random password
	newPassword := s.generateSecureToken(12)
	hash, err := s.HashPassword(newPassword)
	if err != nil {
		return "", err
	}

	user.PasswordHash = hash
	user.UpdatedAt = time.Now()
	s.users[username] = user

	// Invalidate all sessions
	s.InvalidateUserSessions(user.ID)

	return newPassword, nil
}

// EnableMFA enables MFA for a user
func (s *Service) EnableMFA(username string) (string, []string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[username]
	if !exists {
		return "", nil, errors.New("user not found")
	}

	// Generate MFA secret
	secret := s.generateSecureToken(16)

	// Generate backup codes
	backupCodes := make([]string, s.mfaConfig.BackupCodes)
	for i := range backupCodes {
		backupCodes[i] = s.generateSecureToken(8)
	}

	user.MFASecret = secret
	user.UpdatedAt = time.Now()
	s.users[username] = user

	return secret, backupCodes, nil
}

// VerifyMFA verifies an MFA code
func (s *Service) VerifyMFA(username, code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists || !user.MFAEnabled {
		return false
	}

	// In production, this would verify TOTP code
	// Simplified for demonstration
	return code == "123456"
}

// HashPassword hashes a password using bcrypt-like approach
func (s *Service) HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(append([]byte(password), salt...))
	return fmt.Sprintf("%s:%s", base64.StdEncoding.EncodeToString(salt), hex.EncodeToString(hash[:])), nil
}

// VerifyPassword verifies a password against a hash
func (s *Service) VerifyPassword(password, hash string) bool {
	// Parse the hash
	var saltB64, hashHex string
	fmt.Sscanf(hash, "%s:%s", &saltB64, &hashHex)

	salt, err := base64.StdEncoding.DecodeString(saltB64)
	if err != nil {
		return false
	}

	computedHash := sha256.Sum256(append([]byte(password), salt...))
	return subtle.ConstantTimeCompare([]byte(hashHex), []byte(hex.EncodeToString(computedHash[:]))) == 1
}

// generateSecureToken generates a cryptographically secure token
func (s *Service) generateSecureToken(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Printf("Failed to generate secure token: %v", err)
		// Fallback to time-based token
		return fmt.Sprintf("%d_%d", time.Now().UnixNano(), length)
	}
	return hex.EncodeToString(bytes)
}

// GetUser retrieves a user by username
func (s *Service) GetUser(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUsers retrieves all users
func (s *Service) GetUsers() ([]*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	return users, nil
}

// GetRoles retrieves all roles
func (s *Service) GetRoles() []*Role {
	roles := make([]*Role, 0, len(s.roles))
	for _, role := range s.roles {
		roles = append(roles, role)
	}
	return roles
}

// CleanupExpiredSessions removes expired sessions
func (s *Service) CleanupExpiredSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, token)
		}
	}
}
