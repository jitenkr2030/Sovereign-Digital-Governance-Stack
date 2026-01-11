package ports

import (
	"context"

	"github.com/csic-platform/services/security/access-control/internal/core/domain"
)

// PolicyRepository defines the interface for policy data operations
type PolicyRepository interface {
	// FindByID retrieves a policy by its unique identifier
	FindByID(ctx context.Context, id string) (*domain.Policy, error)

	// FindAll retrieves all policies
	FindAll(ctx context.Context) ([]*domain.Policy, error)

	// FindEnabled retrieves all enabled policies
	FindEnabled(ctx context.Context) ([]*domain.Policy, error)

	// FindApplicable retrieves policies applicable to a request
	FindApplicable(ctx context.Context, req *domain.AccessRequest) ([]*domain.Policy, error)

	// Create creates a new policy
	Create(ctx context.Context, policy *domain.Policy) error

	// Update updates an existing policy
	Update(ctx context.Context, policy *domain.Policy) error

	// Delete deletes a policy
	Delete(ctx context.Context, id string) error

	// Enable enables a policy
	Enable(ctx context.Context, id string) error

	// Disable disables a policy
	Disable(ctx context.Context, id string) error
}

// ResourceOwnershipRepository defines the interface for resource ownership operations
type ResourceOwnershipRepository interface {
	// FindByResourceID retrieves ownership by resource ID
	FindByResourceID(ctx context.Context, resourceID string) (*domain.ResourceOwnership, error)

	// FindByOwnerID retrieves all resources owned by a user
	FindByOwnerID(ctx context.Context, ownerID string) ([]*domain.ResourceOwnership, error)

	// Create creates a new ownership record
	Create(ctx context.Context, ownership *domain.ResourceOwnership) error

	// Update updates an ownership record
	Update(ctx context.Context, ownership *domain.ResourceOwnership) error

	// Delete deletes an ownership record
	Delete(ctx context.Context, resourceID string) error
}

// AuditLogRepository defines the interface for audit log operations
type AuditLogRepository interface {
	// Create creates a new audit log entry
	Create(ctx context.Context, log *domain.AuditLog) error

	// FindByID retrieves an audit log by ID
	FindByID(ctx context.Context, id string) (*domain.AuditLog, error)

	// FindByUserID retrieves audit logs for a user
	FindByUserID(ctx context.Context, userID string, limit int) ([]*domain.AuditLog, error)

	// FindByResourceID retrieves audit logs for a resource
	FindByResourceID(ctx context.Context, resourceID string, limit int) ([]*domain.AuditLog, error)

	// FindByTimeRange retrieves audit logs within a time range
	FindByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*domain.AuditLog, error)

	// Cleanup deletes audit logs older than the specified duration
	Cleanup(ctx context.Context, olderThan time.Duration) error
}

// AccessControlService defines the business logic for access control
type AccessControlService interface {
	// CheckAccess performs access control check
	CheckAccess(ctx context.Context, req *domain.AccessRequest) (*domain.AccessResponse, error)

	// IsOwner checks if a subject is the owner of a resource
	IsOwner(ctx context.Context, resourceID string, subjectID string) (bool, error)

	// CanAccessResource checks if a subject can access a resource
	CanAccessResource(ctx context.Context, subjectID string, resourceID string, actionName string) (bool, error)
}

// PolicyManagementService defines the business logic for policy management
type PolicyManagementService interface {
	// CreatePolicy creates a new policy
	CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*CreatePolicyResponse, error)

	// GetPolicy retrieves a policy by ID
	GetPolicy(ctx context.Context, id string) (*domain.Policy, error)

	// ListPolicies lists all policies
	ListPolicies(ctx context.Context, req *ListPoliciesRequest) (*ListPoliciesResponse, error)

	// UpdatePolicy updates an existing policy
	UpdatePolicy(ctx context.Context, req *UpdatePolicyRequest) error

	// DeletePolicy deletes a policy
	DeletePolicy(ctx context.Context, id string) error

	// EnablePolicy enables a policy
	EnablePolicy(ctx context.Context, id string) error

	// DisablePolicy disables a policy
	DisablePolicy(ctx context.Context, id string) error
}

// Request and response types for PolicyManagementService

type CreatePolicyRequest struct {
	Name        string                    `json:"name" validate:"required"`
	Description string                    `json:"description"`
	Effect      domain.PolicyEffect       `json:"effect" validate:"required"`
	Priority    int                       `json:"priority"`
	Conditions  domain.PolicyConditions   `json:"conditions"`
}

type CreatePolicyResponse struct {
	PolicyID string `json:"policy_id"`
	Name     string `json:"name"`
	Success  bool   `json:"success"`
}

type ListPoliciesRequest struct {
	EnabledOnly *bool  `json:"enabled_only,omitempty"`
	Effect      string `json:"effect,omitempty"`
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
}

type ListPoliciesResponse struct {
	Policies []*domain.Policy `json:"policies"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

type UpdatePolicyRequest struct {
	ID          string               `json:"id" validate:"required"`
	Name        *string              `json:"name,omitempty"`
	Description *string              `json:"description,omitempty"`
	Effect      *domain.PolicyEffect `json:"effect,omitempty"`
	Priority    *int                 `json:"priority,omitempty"`
	Conditions  *domain.PolicyConditions `json:"conditions,omitempty"`
}
