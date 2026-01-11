package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
)

// PolicyRepository defines the interface for policy data persistence operations.
// This interface is implemented by infrastructure adapters to provide
// data access functionality for policy entities.
type PolicyRepository interface {
	// Create adds a new policy to the repository
	Create(ctx context.Context, policy *domain.Policy) error

	// GetByID retrieves a policy by its unique identifier
	GetByID(ctx context.Context, id string) (*domain.Policy, error)

	// GetByName retrieves a policy by its name
	GetByName(ctx context.Context, name string) (*domain.Policy, error)

	// GetByCategory retrieves all policies matching the specified category
	GetByCategory(ctx context.Context, category domain.PolicyCategory) ([]*domain.Policy, error)

	// GetByScope retrieves all policies applicable to the specified regulatory scope
	GetByScope(ctx context.Context, scope domain.RegulatoryScope) ([]*domain.Policy, error)

	// Update modifies an existing policy in the repository
	Update(ctx context.Context, policy *domain.Policy) error

	// Delete removes a policy by its unique identifier
	Delete(ctx context.Context, id string) error

	// List retrieves all policies with optional pagination
	List(ctx context.Context, offset, limit int) ([]*domain.Policy, error)

	// ListActive retrieves all currently active policies
	ListActive(ctx context.Context) ([]*domain.Policy, error)

	// ListByEffectiveDate retrieves policies effective within the specified date range
	ListByEffectiveDate(ctx context.Context, startDate, endDate time.Time) ([]*domain.Policy, error)

	// Exists checks if a policy with the given ID exists
	Exists(ctx context.Context, id string) (bool, error)

	// Count returns the total number of policies in the repository
	Count(ctx context.Context) (int64, error)

	// CountByCategory returns the count of policies grouped by category
	CountByCategory(ctx context.Context) (map[domain.PolicyCategory]int64, error)

	// Search performs a text search on policy content and metadata
	Search(ctx context.Context, query string, offset, limit int) ([]*domain.Policy, error)

	// GetVersionHistory retrieves all versions of a specific policy
	GetVersionHistory(ctx context.Context, policyID string) ([]*domain.PolicyVersion, error)

	// CreateVersion creates a new version of an existing policy
	CreateVersion(ctx context.Context, policyID string, version *domain.PolicyVersion) error
}

// PolicyService defines the interface for policy business logic operations.
// This interface encapsulates all domain-specific operations that can be
// performed on policy entities.
type PolicyService interface {
	// CreatePolicy creates a new policy with validation
	CreatePolicy(ctx context.Context, input *domain.CreatePolicyInput) (*domain.Policy, error)

	// GetPolicy retrieves a policy by ID with all related data
	GetPolicy(ctx context.Context, id string) (*domain.Policy, error)

	// GetPolicyByName retrieves a policy by its name
	GetPolicyByName(ctx context.Context, name string) (*domain.Policy, error)

	// ListPolicies retrieves all policies with optional filtering and pagination
	ListPolicies(ctx context.Context, filter *domain.PolicyFilter) ([]*domain.Policy, error)

	// UpdatePolicy updates an existing policy with validation
	UpdatePolicy(ctx context.Context, id string, input *domain.UpdatePolicyInput) (*domain.Policy, error)

	// DeletePolicy soft-deletes a policy (marks as deprecated)
	DeletePolicy(ctx context.Context, id string) error

	// ActivatePolicy activates a previously inactive policy
	ActivatePolicy(ctx context.Context, id string) error

	// DeactivatePolicy deactivates an active policy
	DeactivatePolicy(ctx context.Context, id string) error

	// ValidatePolicy validates a policy against regulatory requirements
	ValidatePolicy(ctx context.Context, policyID string) (*domain.ValidationResult, error)

	// GetApplicablePolicies retrieves all policies applicable to a given scope and entity
	GetApplicablePolicies(ctx context.Context, scope domain.RegulatoryScope, entityType string, entityID string) ([]*domain.Policy, error)

	// CheckCompliance checks if an entity complies with all applicable policies
	CheckCompliance(ctx context.Context, entityType string, entityID string) (*domain.ComplianceCheckResult, error)

	// SearchPolicies performs an advanced search on policies
	SearchPolicies(ctx context.Context, query string, filter *domain.PolicyFilter) ([]*domain.Policy, error)

	// GetPolicyVersion retrieves a specific version of a policy
	GetPolicyVersion(ctx context.Context, policyID string, versionID string) (*domain.PolicyVersion, error)

	// RollbackPolicy reverts a policy to a previous version
	RollbackPolicy(ctx context.Context, policyID string, targetVersionID string) (*domain.Policy, error)

	// GetPolicyStats returns statistics about policies
	GetPolicyStats(ctx context.Context) (*domain.PolicyStats, error)

	// ImportPolicy imports a policy from an external source
	ImportPolicy(ctx context.Context, source string, metadata *domain.ImportMetadata) (*domain.Policy, error)

	// ExportPolicy exports a policy to the specified format
	ExportPolicy(ctx context.Context, policyID string, format string) ([]byte, error)

	// ClonePolicy creates a copy of an existing policy
	ClonePolicy(ctx context.Context, policyID string, newName string) (*domain.Policy, error)

	// DeprecatePolicy marks a policy as deprecated with a replacement
	DeprecatePolicy(ctx context.Context, policyID string, replacementID string, deprecationDate time.Time) error
}

// PolicyEventPublisher defines the interface for publishing policy-related events.
// This interface is used to publish domain events for async processing and integration.
type PolicyEventPublisher interface {
	// PublishPolicyCreated publishes an event when a new policy is created
	PublishPolicyCreated(ctx context.Context, policy *domain.Policy) error

	// PublishPolicyUpdated publishes an event when a policy is updated
	PublishPolicyUpdated(ctx context.Context, policy *domain.Policy, changes []string) error

	// PublishPolicyActivated publishes an event when a policy is activated
	PublishPolicyActivated(ctx context.Context, policy *domain.Policy) error

	// PublishPolicyDeactivated publishes an event when a policy is deactivated
	PublishPolicyDeactivated(ctx context.Context, policy *domain.Policy) error

	// PublishPolicyDeleted publishes an event when a policy is deleted
	PublishPolicyDeleted(ctx context.Context, policyID string) error

	// PublishComplianceCheck publishes an event with compliance check results
	PublishComplianceCheck(ctx context.Context, result *domain.ComplianceCheckResult) error

	// PublishValidationResult publishes an event with policy validation results
	PublishValidationResult(ctx context.Context, result *domain.ValidationResult) error
}

// PolicyEventSubscriber defines the interface for consuming policy-related events.
// This interface allows the service to react to events from other services.
type PolicyEventSubscriber interface {
	// SubscribeToPolicyEvents starts consuming policy-related events
	SubscribeToPolicyEvents(ctx context.Context) error

	// HandlePolicyEvent processes a single policy event
	HandlePolicyEvent(ctx context.Context, event *domain.PolicyEvent) error

	// Close stops the event subscription
	Close() error
}

// PolicyCacheRepository defines the interface for caching policy data.
// This interface provides caching capabilities for improved performance.
type PolicyCacheRepository interface {
	// Get retrieves a cached policy by ID
	Get(ctx context.Context, id string) (*domain.Policy, error)

	// Set stores a policy in the cache
	Set(ctx context.Context, policy *domain.Policy, ttl time.Duration) error

	// Delete removes a policy from the cache
	Delete(ctx context.Context, id string) error

	// Invalidate invalidates all cached policies
	Invalidate(ctx context.Context) error

	// InvalidateByCategory invalidates all cached policies in a category
	InvalidateByCategory(ctx context.Context, category domain.PolicyCategory) error

	// InvalidateByScope invalidates all cached policies for a scope
	InvalidateByScope(ctx context.Context, scope domain.RegulatoryScope) error

	// GetStats returns cache statistics
	GetStats(ctx context.Context) (*domain.CacheStats, error)
}
