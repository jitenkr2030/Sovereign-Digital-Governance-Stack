package repository

import (
	"context"
	"github.com/csic-platform/services/security/access-control/internal/core/domain"
)

// PolicyRepository defines the interface for policy data operations
type PolicyRepository interface {
	Create(ctx context.Context, policy *domain.Policy) error
	Update(ctx context.Context, policy *domain.Policy) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Policy, error)
	GetAll(ctx context.Context) ([]*domain.Policy, error)
	GetEnabled(ctx context.Context) ([]*domain.Policy, error)
	GetByName(ctx context.Context, name string) (*domain.Policy, error)
	Search(ctx context.Context, query string) ([]*domain.Policy, error)
}

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Role, error)
	GetAll(ctx context.Context) ([]*domain.Role, error)
	GetByName(ctx context.Context, name string) (*domain.Role, error)
	GetUserRoles(ctx context.Context, userID string) ([]*domain.Role, error)
}

// UserRoleRepository defines the interface for user-role association operations
type UserRoleRepository interface {
	AssignRole(ctx context.Context, userRole *domain.UserRole) error
	RevokeRole(ctx context.Context, userID, roleID, scope, scopeID string) error
	RevokeAllUserRoles(ctx context.Context, userID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*domain.UserRole, error)
	GetRoleUsers(ctx context.Context, roleID string) ([]*domain.UserRole, error)
	HasRole(ctx context.Context, userID, roleID string) (bool, error)
}

// DecisionRepository defines the interface for access decision logging
type DecisionRepository interface {
	StoreDecision(ctx context.Context, decision *domain.AccessDecision, request *domain.AccessRequest) error
	GetDecisions(ctx context.Context, filters map[string]string) ([]*domain.AccessDecision, error)
	GetRecentDecisions(ctx context.Context, limit int) ([]*domain.AccessDecision, error)
}
