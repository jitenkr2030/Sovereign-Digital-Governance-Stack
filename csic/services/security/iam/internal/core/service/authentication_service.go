package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/services/security/iam/internal/core/domain"
	"github.com/csic-platform/services/security/iam/internal/core/ports"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuthenticationServiceImpl implements the AuthenticationService interface
type AuthenticationServiceImpl struct {
	userRepo          ports.UserRepository
	roleRepo          ports.RoleRepository
	sessionRepo       ports.SessionRepository
	refreshTokenRepo  ports.RefreshTokenRepository
	tokenGen          ports.TokenGenerator
	tokenBlacklist    ports.TokenBlacklist
	metrics           ports.MetricsCollector
	logger            *zap.Logger
}

// NewAuthenticationService creates a new AuthenticationServiceImpl
func NewAuthenticationService(
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	sessionRepo ports.SessionRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	tokenGen ports.TokenGenerator,
	tokenBlacklist ports.TokenBlacklist,
	metrics ports.MetricsCollector,
	logger *zap.Logger,
) *AuthenticationServiceImpl {
	return &AuthenticationServiceImpl{
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		sessionRepo:      sessionRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenGen:         tokenGen,
		tokenBlacklist:   tokenBlacklist,
		metrics:          metrics,
		logger:           logger,
	}
}

// Authenticate authenticates a user with username and password
func (s *AuthenticationServiceImpl) Authenticate(
	ctx context.Context,
	req *domain.LoginRequest,
) (*domain.LoginResponse, error) {
	start := time.Now()
	s.logger.Info("Authentication attempt",
		zap.String("username", req.Username),
		zap.String("ip", req.IPAddress))

	// Find user by username
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Warn("Authentication failed - user not found",
			zap.String("username", req.Username))
		if s.metrics != nil {
			s.metrics.IncrementLoginAttempts(false)
		}
		return nil, domain.ErrInvalidCredentials
	}

	// Check if account is locked
	if user.IsLocked() {
		s.logger.Warn("Authentication failed - account locked",
			zap.String("username", req.Username))
		if s.metrics != nil {
			s.metrics.IncrementLoginAttempts(false)
		}
		return nil, domain.ErrAccountLocked
	}

	// Check if account is enabled
	if !user.Enabled {
		s.logger.Warn("Authentication failed - account disabled",
			zap.String("username", req.Username))
		if s.metrics != nil {
			s.metrics.IncrementLoginAttempts(false)
		}
		return nil, domain.ErrAccountDisabled
	}

	// Verify password
	if !user.VerifyPassword(req.Password) {
		s.logger.Warn("Authentication failed - invalid password",
			zap.String("username", req.Username))

		// Increment login attempts
		s.userRepo.IncrementLoginAttempts(ctx, user.ID.String())

		// Lock account after 5 failed attempts
		if user.LoginAttempts >= 4 {
			lockUntil := time.Now().Add(30 * time.Minute)
			user.Lock(lockUntil)
			s.userRepo.Update(ctx, user)
			s.logger.Warn("Account locked due to failed attempts",
				zap.String("username", req.Username))
		}

		if s.metrics != nil {
			s.metrics.IncrementLoginAttempts(false)
		}
		return nil, domain.ErrInvalidCredentials
	}

	// Reset login attempts on successful login
	s.userRepo.ResetLoginAttempts(ctx, user.ID.String())

	// Load role with permissions
	if user.Role == nil && user.RoleID != uuid.Nil {
		role, err := s.roleRepo.FindByID(ctx, user.RoleID.String())
		if err == nil {
			user.Role = role
		}
	}

	// Check MFA if enabled
	if user.MFAEnabled {
		if req.MFACode == "" {
			s.logger.Info("MFA code required",
				zap.String("username", req.Username))
			return &domain.LoginResponse{
				MFARequired: true,
				MFAHint:     "Enter your MFA code",
			}, nil
		}

		// Verify MFA code (simplified - would use MFA service in production)
		if !verifyMFACode(user.MFASecret, req.MFACode) {
			s.logger.Warn("Authentication failed - invalid MFA code",
				zap.String("username", req.Username))
			if s.metrics != nil {
				s.metrics.IncrementLoginAttempts(false)
			}
			return nil, domain.ErrMFACodeInvalid
		}
	}

	// Generate tokens
	accessToken, err := s.tokenGen.GenerateAccessToken(ctx, user)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenGen.GenerateRefreshToken(ctx, user)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session
	session := &domain.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		Token:        accessToken,
		RefreshToken: refreshToken,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
	}
	s.sessionRepo.Create(ctx, session)

	// Store refresh token
	refreshTokenRecord := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}
	s.refreshTokenRepo.Create(ctx, refreshTokenRecord)

	// Update last login
	s.userRepo.UpdateLogin(ctx, user.ID.String())

	// Build permissions list
	permissions := make([]string, 0)
	if user.Role != nil {
		for _, perm := range user.Role.Permissions {
			permissions = append(permissions, fmt.Sprintf("%s:%s", perm.Resource, perm.Action))
		}
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.IncrementLoginAttempts(true)
		s.metrics.RecordAuthLatency(time.Since(start).String())
	}

	s.logger.Info("Authentication successful",
		zap.String("username", req.Username),
		zap.String("user_id", user.ID.String()))

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.tokenGen.GetAccessTokenExpiry(),
		User: &domain.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role.Name,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

// Register creates a new user account
func (s *AuthenticationServiceImpl) Register(
	ctx context.Context,
	req *domain.RegisterRequest,
) (*domain.User, error) {
	s.logger.Info("User registration",
		zap.String("username", req.Username),
		zap.String("email", req.Email))

	// Check if username already exists
	existing, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err == nil && existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Check if email already exists
	existing, err = s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Get default role if not specified
	roleID := uuid.Nil
	if req.RoleID != "" {
		roleID = uuid.MustParse(req.RoleID)
	} else {
		// Find default role (User)
		role, err := s.roleRepo.FindByName(ctx, "User")
		if err != nil {
			s.logger.Error("Default role not found", zap.Error(err))
			return nil, fmt.Errorf("default role not found")
		}
		roleID = role.ID
	}

	// Create user
	user := &domain.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		RoleID:    roleID,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set password
	if err := user.SetPassword(req.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Save user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if s.metrics != nil {
		s.metrics.IncrementUsersCreated()
	}

	s.logger.Info("User registered successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username))

	return user, nil
}

// Refresh refreshes an access token using a refresh token
func (s *AuthenticationServiceImpl) Refresh(
	ctx context.Context,
	refreshToken string,
) (*domain.LoginResponse, error) {
	s.logger.Debug("Token refresh attempt")

	// Find refresh token
	token, err := s.refreshTokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		s.logger.Warn("Refresh token not found")
		return nil, domain.ErrTokenInvalid
	}

	// Check if token is used
	if token.Used {
		return nil, domain.ErrTokenInvalid
	}

	// Check if token is expired
	if token.ExpiresAt.Before(time.Now()) {
		return nil, domain.ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, token.UserID.String())
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// Check if user is enabled
	if !user.Enabled {
		return nil, domain.ErrAccountDisabled
	}

	// Mark old token as used
	s.refreshTokenRepo.MarkUsed(ctx, token.ID.String())

	// Generate new tokens
	accessToken, err := s.tokenGen.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.tokenGen.GenerateRefreshToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store new refresh token
	newToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}
	s.refreshTokenRepo.Create(ctx, newToken)

	s.logger.Info("Token refreshed successfully",
		zap.String("user_id", user.ID.String()))

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.tokenGen.GetAccessTokenExpiry(),
		User: &domain.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role.Name,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

// Logout revokes the current session
func (s *AuthenticationServiceImpl) Logout(ctx context.Context, token string) error {
	s.logger.Debug("Logout attempt")

	// Find session by token
	session, err := s.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return nil // Session may not exist
	}

	// Revoke session
	now := time.Now()
	session.RevokedAt = &now
	s.sessionRepo.Revoke(ctx, session.ID.String())

	// Add token to blacklist
	if err := s.tokenBlacklist.Add(ctx, token, int64(time.Until(session.ExpiresAt).Seconds())); err != nil {
		s.logger.Error("Failed to blacklist token", zap.Error(err))
	}

	// Mark refresh token as used
	s.refreshTokenRepo.MarkUsed(ctx, session.RefreshToken)

	if s.metrics != nil {
		s.metrics.IncrementSessionsRevoked()
	}

	s.logger.Info("Logout successful",
		zap.String("session_id", session.ID.String()))

	return nil
}

// LogoutAll revokes all sessions for a user
func (s *AuthenticationServiceImpl) LogoutAll(ctx context.Context, userID string) error {
	s.logger.Info("Logout all sessions",
		zap.String("user_id", userID))

	s.sessionRepo.RevokeAll(ctx, userID)

	if s.metrics != nil {
		s.metrics.IncrementSessionsRevoked()
	}

	return nil
}

// ValidateToken validates an access token
func (s *AuthenticationServiceImpl) ValidateToken(
	ctx context.Context,
	token string,
) (*domain.AuthContext, error) {
	// Check if token is blacklisted
	blacklisted, err := s.tokenBlacklist.IsBlacklisted(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if blacklisted {
		return nil, domain.ErrTokenBlacklisted
	}

	// Validate token
	authContext, err := s.tokenGen.ValidateAccessToken(ctx, token)
	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.ErrTokenInvalid
	}

	return authContext, nil
}

// verifyMFACode verifies an MFA code (simplified implementation)
func verifyMFACode(secret, code string) bool {
	// In production, use a proper TOTP library
	return len(code) == 6
}

// Ensure AuthenticationServiceImpl implements AuthenticationService
var _ ports.AuthenticationService = (*AuthenticationServiceImpl)(nil)
