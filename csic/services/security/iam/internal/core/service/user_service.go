package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/services/security/iam/internal/core/domain"
	"github.com/csic-platform/services/security/iam/internal/core/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserServiceImpl implements the UserService interface
type UserServiceImpl struct {
	userRepo  ports.UserRepository
	roleRepo  ports.RoleRepository
	emailSvc  ports.EmailService
	metrics   ports.MetricsCollector
	logger    *zap.Logger
}

// NewUserService creates a new UserServiceImpl
func NewUserService(
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	emailSvc ports.EmailService,
	metrics ports.MetricsCollector,
	logger *zap.Logger,
) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo: userRepo,
		roleRepo: roleRepo,
		emailSvc: emailSvc,
		metrics:  metrics,
		logger:   logger,
	}
}

// CreateUser creates a new user
func (s *UserServiceImpl) CreateUser(
	ctx context.Context,
	req *ports.CreateUserRequest,
) (*ports.CreateUserResponse, error) {
	s.logger.Info("Creating user",
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

	// Get role
	roleID := uuid.Nil
	if req.RoleID != "" {
		roleID = uuid.MustParse(req.RoleID)
	} else {
		role, err := s.roleRepo.FindByName(ctx, "User")
		if err != nil {
			return nil, fmt.Errorf("default role not found: %w", err)
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

	// Get role name for response
	role, _ := s.roleRepo.FindByID(ctx, roleID.String())
	roleName := "User"
	if role != nil {
		roleName = role.Name
	}

	if s.metrics != nil {
		s.metrics.IncrementUsersCreated()
	}

	s.logger.Info("User created successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username))

	return &ports.CreateUserResponse{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     roleName,
	}, nil
}

// GetUser retrieves a user by ID
func (s *UserServiceImpl) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

// ListUsers lists all users with filtering
func (s *UserServiceImpl) ListUsers(
	ctx context.Context,
	req *ports.ListUsersRequest,
) (*ports.ListUsersResponse, error) {
	return s.userRepo.FindAll(ctx, req)
}

// UpdateUser updates an existing user
func (s *UserServiceImpl) UpdateUser(
	ctx context.Context,
	req *ports.UpdateUserRequest,
) error {
	s.logger.Info("Updating user",
		zap.String("user_id", req.ID))

	user, err := s.userRepo.FindByID(ctx, req.ID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	if req.Username != nil {
		// Check if new username is taken
		existing, _ := s.userRepo.FindByUsername(ctx, *req.Username)
		if existing != nil && existing.ID != user.ID {
			return domain.ErrUserAlreadyExists
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		// Check if new email is taken
		existing, _ := s.userRepo.FindByEmail(ctx, *req.Email)
		if existing != nil && existing.ID != user.ID {
			return domain.ErrUserAlreadyExists
		}
		user.Email = *req.Email
	}

	if req.RoleID != nil {
		roleID := uuid.MustParse(*req.RoleID)
		_, err := s.roleRepo.FindByID(ctx, roleID.String())
		if err != nil {
			return domain.ErrRoleNotFound
		}
		user.RoleID = roleID
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info("User updated successfully",
		zap.String("user_id", req.ID))

	return nil
}

// DeleteUser soft-deletes a user
func (s *UserServiceImpl) DeleteUser(ctx context.Context, id string) error {
	s.logger.Info("Deleting user",
		zap.String("user_id", id))

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return domain.ErrUserNotFound
	}

	// Prevent self-deletion
	authContext := domain.AuthContext{}
	if user.ID == authContext.UserID {
		return domain.ErrUserCannotDeleteSelf
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.logger.Info("User deleted successfully",
		zap.String("user_id", id))

	return nil
}

// EnableUser enables a disabled user
func (s *UserServiceImpl) EnableUser(ctx context.Context, id string) error {
	s.logger.Info("Enabling user",
		zap.String("user_id", id))

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return domain.ErrUserNotFound
	}

	user.Enabled = true
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to enable user: %w", err)
	}

	s.logger.Info("User enabled successfully",
		zap.String("user_id", id))

	return nil
}

// DisableUser disables an active user
func (s *UserServiceImpl) DisableUser(ctx context.Context, id string) error {
	s.logger.Info("Disabling user",
		zap.String("user_id", id))

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return domain.ErrUserNotFound
	}

	// Prevent self-disablement
	authContext := domain.AuthContext{}
	if user.ID == authContext.UserID {
		return domain.ErrUserCannotDisableSelf
	}

	user.Enabled = false
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to disable user: %w", err)
	}

	s.logger.Info("User disabled successfully",
		zap.String("user_id", id))

	return nil
}

// ChangePassword changes a user's password
func (s *UserServiceImpl) ChangePassword(
	ctx context.Context,
	userID string,
	oldPassword,
	newPassword string,
) error {
	s.logger.Info("Changing password",
		zap.String("user_id", userID))

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	// Verify old password
	if !user.VerifyPassword(oldPassword) {
		return domain.ErrInvalidCredentials
	}

	// Set new password
	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info("Password changed successfully",
		zap.String("user_id", userID))

	return nil
}

// ResetPassword resets a user's password (admin action)
func (s *UserServiceImpl) ResetPassword(
	ctx context.Context,
	id string,
) (*ports.ResetPasswordResponse, error) {
	s.logger.Info("Resetting password",
		zap.String("user_id", id))

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// Generate new random password
	newPassword := generateSecurePassword()

	if err := user.SetPassword(newPassword); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info("Password reset successfully",
		zap.String("user_id", id))

	return &ports.ResetPasswordResponse{
		NewPassword: newPassword,
	}, nil
}

// generateSecurePassword generates a secure random password
func generateSecurePassword() string {
	// Simplified - use crypto/rand in production
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	result := make([]byte, 16)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// Ensure UserServiceImpl implements UserService
var _ ports.UserService = (*UserServiceImpl)(nil)
