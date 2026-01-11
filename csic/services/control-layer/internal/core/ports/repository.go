package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
)

// GenericRepository defines a generic interface for CRUD operations.
// This interface can be used as a base for type-safe repositories.
type GenericRepository interface {
	// Create creates a new entity
	Create(ctx context.Context, entity interface{}) error

	// GetByID retrieves an entity by ID
	GetByID(ctx context.Context, id string) (interface{}, error)

	// Update updates an entity
	Update(ctx context.Context, entity interface{}) error

	// Delete deletes an entity
	Delete(ctx context.Context, id string) error

	// Exists checks if an entity exists
	Exists(ctx context.Context, id string) (bool, error)

	// List lists all entities
	List(ctx context.Context, offset, limit int) ([]interface{}, error)

	// Count counts all entities
	Count(ctx context.Context) (int64, error)
}

// BaseRepository defines additional operations for all repositories.
// This interface extends GenericRepository with common operations.
type BaseRepository interface {
	GenericRepository

	// FindBy finds entities matching the criteria
	FindBy(ctx context.Context, criteria map[string]interface{}, offset, limit int) ([]interface{}, error)

	// FindOne finds a single entity matching the criteria
	FindOne(ctx context.Context, criteria map[string]interface{}) (interface{}, error)

	// UpdateBy updates entities matching the criteria
	UpdateBy(ctx context.Context, criteria map[string]interface{}, updates map[string]interface{}) (int64, error)

	// DeleteBy deletes entities matching the criteria
	DeleteBy(ctx context.Context, criteria map[string]interface{}) (int64, error)

	// CountBy counts entities matching the criteria
	CountBy(ctx context.Context, criteria map[string]interface{}) (int64, error)
}

// SoftDeleteRepository defines soft delete operations.
type SoftDeleteRepository interface {
	// SoftDelete soft-deletes an entity
	SoftDelete(ctx context.Context, id string) error

	// SoftDeleteBy soft-deletes entities matching the criteria
	SoftDeleteBy(ctx context.Context, criteria map[string]interface{}) (int64, error)

	// Restore restores a soft-deleted entity
	Restore(ctx context.Context, id string) error

	// FindDeleted finds soft-deleted entities
	FindDeleted(ctx context.Context, offset, limit int) ([]interface{}, error)

	// CountDeleted counts soft-deleted entities
	CountDeleted(ctx context.Context) (int64, error)
}

// VersionedRepository defines version control operations for entities.
type VersionedRepository interface {
	// CreateVersion creates a new version of an entity
	CreateVersion(ctx context.Context, entityID string, version interface{}) error

	// GetVersion gets a specific version of an entity
	GetVersion(ctx context.Context, entityID string, versionID string) (interface{}, error)

	// GetVersions gets all versions of an entity
	GetVersions(ctx context.Context, entityID string) ([]interface{}, error)

	// GetCurrentVersion gets the current version of an entity
	GetCurrentVersion(ctx context.Context, entityID string) (interface{}, error)

	// GetVersionHistory gets the version history of an entity
	GetVersionHistory(ctx context.Context, entityID string) ([]*domain.VersionInfo, error)

	// Rollback rolls back to a specific version
	Rollback(ctx context.Context, entityID string, versionID string) error

	// CompareVersions compares two versions of an entity
	CompareVersions(ctx context.Context, entityID string, versionID1, versionID2 string) (*domain.VersionComparison, error)
}

// AuditedRepository defines audit trail operations.
type AuditedRepository interface {
	// CreateAuditEntry creates an audit entry
	CreateAuditEntry(ctx context.Context, entry *domain.AuditEntry) error

	// GetAuditTrail gets the audit trail for an entity
	GetAuditTrail(ctx context.Context, entityType, entityID string) ([]*domain.AuditEntry, error)

	// GetAuditEntry gets a specific audit entry
	GetAuditEntry(ctx context.Context, entryID string) (*domain.AuditEntry, error)

	// QueryAuditLog queries the audit log
	QueryAuditLog(ctx context.Context, query *domain.AuditQuery) ([]*domain.AuditEntry, error)

	// GetAuditStats gets audit statistics
	GetAuditStats(ctx context.Context, entityType string) (*domain.AuditStats, error)
}

// TenantRepository defines tenant-specific operations.
type TenantRepository interface {
	// SetTenant sets the current tenant context
	SetTenant(ctx context.Context, tenantID string) context.Context

	// GetTenant gets the current tenant
	GetTenant(ctx context.Context) string

	// GetByTenant gets entities for a specific tenant
	GetByTenant(ctx context.Context, tenantID string, offset, limit int) ([]interface{}, error)

	// CountByTenant counts entities for a specific tenant
	CountByTenant(ctx context.Context, tenantID string) (int64, error)

	// DeleteByTenant deletes all entities for a tenant
	DeleteByTenant(ctx context.Context, tenantID string) (int64, error)
}

// MultiTenantRepository defines multi-tenant operations.
type MultiTenantRepository interface {
	TenantRepository

	// GetTenants gets all tenants
	GetTenants(ctx context.Context) ([]string, error)

	// CreateTenant creates a new tenant
	CreateTenant(ctx context.Context, tenantID string, config *domain.TenantConfig) error

	// UpdateTenant updates tenant configuration
	UpdateTenant(ctx context.Context, tenantID string, config *domain.TenantConfig) error

	// DeleteTenant deletes a tenant
	DeleteTenant(ctx context.Context, tenantID string) error

	// GetTenantConfig gets tenant configuration
	GetTenantConfig(ctx context.Context, tenantID string) (*domain.TenantConfig, error)
}

// ReadOnlyRepository defines read-only operations.
type ReadOnlyRepository interface {
	// GetByID retrieves an entity by ID (read-only)
	GetByID(ctx context.Context, id string) (interface{}, error)

	// FindBy finds entities matching the criteria (read-only)
	FindBy(ctx context.Context, criteria map[string]interface{}, offset, limit int) ([]interface{}, error)

	// FindOne finds a single entity (read-only)
	FindOne(ctx context.Context, criteria map[string]interface{}) (interface{}, error)

	// List lists all entities (read-only)
	List(ctx context.Context, offset, limit int) ([]interface{}, error)

	// Count counts all entities (read-only)
	Count(ctx context.Context) (int64, error)

	// CountBy counts entities matching the criteria (read-only)
	CountBy(ctx context.Context, criteria map[string]interface{}) (int64, error)
}

// TransactionalRepository defines transactional operations.
type TransactionalRepository interface {
	// WithTransaction executes operations within a transaction
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error

	// BeginTx begins a new transaction
	BeginTx(ctx context.Context) (context.Context, error)

	// CommitTx commits the current transaction
	CommitTx(ctx context.Context) error

	// RollbackTx rolls back the current transaction
	RollbackTx(ctx context.Context) error

	// GetTx returns the current transaction
	GetTx(ctx context.Context) interface{}
}

// BulkRepository defines bulk operation capabilities.
type BulkRepository interface {
	// BulkCreate creates multiple entities
	BulkCreate(ctx context.Context, entities []interface{}) error

	// BulkUpdate updates multiple entities
	BulkUpdate(ctx context.Context, entities []interface{}) error

	// BulkDelete deletes multiple entities
	BulkDelete(ctx context.Context, ids []string) error

	// BulkUpsert creates or updates multiple entities
	BulkUpsert(ctx context.Context, entities []interface{}) error

	// BulkCreateWithTransaction creates multiple entities within a transaction
	BulkCreateWithTransaction(ctx context.Context, entities []interface{}) error

	// GetBulk retrieves multiple entities by IDs
	GetBulk(ctx context.Context, ids []string) ([]interface{}, error)
}

// CompositeRepository combines multiple repositories.
type CompositeRepository interface {
	// AddRepository adds a sub-repository
	AddRepository(name string, repository interface{})

	// GetRepository gets a sub-repository
	GetRepository(name string) interface{}

	// HasRepository checks if a sub-repository exists
	HasRepository(name string) bool

	// GetRepositories gets all sub-repositories
	GetRepositories() map[string]interface{}
}

// RepositoryFactory defines factory methods for creating repositories.
type RepositoryFactory interface {
	// Create creates a new repository instance
	Create(ctx context.Context, entityType interface{}) interface{}

	// GetRepository gets a repository for an entity type
	GetRepository(ctx context.Context, entityType interface{}) interface{}

	// Register registers a repository for an entity type
	Register(ctx context.Context, entityType interface{}, repository interface{})
}

// CachedRepository defines caching operations.
type CachedRepository interface {
	// GetCached gets an entity from cache
	GetCached(ctx context.Context, id string) (interface{}, bool, error)

	// SetCached sets an entity in cache
	SetCached(ctx context.Context, id string, entity interface{}, ttl time.Duration) error

	// InvalidateCache invalidates cache for an entity
	InvalidateCache(ctx context.Context, id string) error

	// InvalidateAllCache invalidates all cached entities
	InvalidateAllCache(ctx context.Context) error

	// InvalidatePatternCache invalidates cache matching a pattern
	InvalidatePatternCache(ctx context.Context, pattern string) error

	// GetCacheStats gets cache statistics
	GetCacheStats(ctx context.Context) (*domain.CacheStats, error)
}

// SearchableRepository defines search operations.
type SearchableRepository interface {
	// Search performs a full-text search
	Search(ctx context.Context, query string, offset, limit int) ([]interface{}, error)

	// SearchWithFilter performs a search with filters
	SearchWithFilter(ctx context.Context, query string, filters map[string]interface{}, offset, limit int) ([]interface{}, error)

	// Suggest provides search suggestions
	Suggest(ctx context.Context, query string, limit int) ([]string, error)

	// Reindex rebuilds the search index
	Reindex(ctx context.Context) error

	// ReindexOne rebuilds the search index for an entity
	ReindexOne(ctx context.Context, id string) error

	// Index indexes an entity
	Index(ctx context.Context, entity interface{}) error

	// RemoveIndex removes an entity from the index
	RemoveIndex(ctx context.Context, id string) error
}

// EventSourcedRepository defines event sourcing operations.
type EventSourcedRepository interface {
	// SaveEvents saves domain events
	SaveEvents(ctx context.Context, entityID string, events []domain.Event, expectedVersion int64) error

	// LoadEvents loads domain events for an entity
	LoadEvents(ctx context.Context, entityID string, fromVersion int64) ([]domain.Event, error)

	// GetCurrentVersion gets the current version of an entity
	GetCurrentVersion(ctx context.Context, entityID string) (int64, error)

	// GetEntity retrieves the current state of an entity
	GetEntity(ctx context.Context, entityID string) (interface{}, error)

	// Snapshot creates a snapshot of an entity
	Snapshot(ctx context.Context, entityID string, entity interface{}) error

	// GetSnapshot gets a snapshot of an entity
	GetSnapshot(ctx context.Context, entityID string) (interface{}, error)
}

// CQRSRepository defines CQRS read and write operations.
type CQRSRepository interface {
	// WriteRepository returns the write repository
	WriteRepository() WriteRepository

	// ReadRepository returns the read repository
	ReadRepository() ReadRepository

	// Synchronize synchronizes read and write models
	Synchronize(ctx context.Context, entityID string) error

	// Rebuild rebuilds the read model
	Rebuild(ctx context.Context, entityID string) error

	// RebuildAll rebuilds all read models
	RebuildAll(ctx context.Context) error
}

// WriteRepository defines write operations.
type WriteRepository interface {
	// Save saves an aggregate
	Save(ctx context.Context, aggregate interface{}) error

	// Load loads an aggregate by ID
	Load(ctx context.Context, id string) (interface{}, error)

	// Delete deletes an aggregate
	Delete(ctx context.Context, id string) error
}

// ReadRepository defines read operations.
type ReadRepository interface {
	// GetByID gets an entity by ID
	GetByID(ctx context.Context, id string) (interface{}, error)

	// Find finds entities matching criteria
	Find(ctx context.Context, criteria interface{}, offset, limit int) ([]interface{}, error)

	// FindOne finds one entity matching criteria
	FindOne(ctx context.Context, criteria interface{}) (interface{}, error)

	// List lists all entities
	List(ctx context.Context, offset, limit int) ([]interface{}, error)

	// Count counts entities
	Count(ctx context.Context, criteria interface{}) (int64, error)

	// Aggregate performs aggregation
	Aggregate(ctx context.Context, pipeline interface{}) ([]interface{}, error)
}

// ProjectionRepository defines projection operations.
type ProjectionRepository interface {
	// CreateProjection creates a projection
	CreateProjection(ctx context.Context, name string, projection interface{}) error

	// GetProjection gets a projection by name
	GetProjection(ctx context.Context, name string) (interface{}, error)

	// UpdateProjection updates a projection
	UpdateProjection(ctx context.Context, name string, projection interface{}) error

	// DeleteProjection deletes a projection
	DeleteProjection(ctx context.Context, name string) error

	// GetProjectionNames gets all projection names
	GetProjectionNames(ctx context.Context) ([]string, error)

	// ResetProjection resets a projection
	ResetProjection(ctx context.Context, name string) error

	// GetProjectionStatus gets the status of a projection
	GetProjectionStatus(ctx context.Context, name string) (*domain.ProjectionStatus, error)
}

// RepositoryMiddleware defines middleware for repository operations.
type RepositoryMiddleware interface {
	// Wrap wraps a repository with middleware
	Wrap(repository interface{}) interface{}

	// PreProcess pre-processes an operation
	PreProcess(ctx context.Context, operation string, args map[string]interface{}) error

	// PostProcess post-processes an operation
	PostProcess(ctx context.Context, operation string, args map[string]interface{}, result interface{}, err error) error

	// OnError handles errors
	OnError(ctx context.Context, operation string, err error) error
}

// RepositoryHealthChecker defines health check operations for repositories.
type RepositoryHealthChecker interface {
	// HealthCheck performs a health check
	HealthCheck(ctx context.Context) (*domain.HealthStatus, error)

	// GetDependencies gets repository dependencies
	GetDependencies(ctx context.Context) ([]string, error)

	// CheckDependency checks a dependency
	CheckDependency(ctx context.Context, name string) (*domain.DependencyHealth, error)
}

// RepositoryMetrics defines metrics collection for repositories.
type RepositoryMetrics interface {
	// RecordOperation records an operation
	RecordOperation(ctx context.Context, operation string, duration time.Duration, success bool)

	// RecordCacheHit records a cache hit
	RecordCacheHit(ctx context.Context, operation string)

	// RecordCacheMiss records a cache miss
	RecordCacheMiss(ctx context.Context, operation string)

	// GetMetrics gets metrics
	GetMetrics(ctx context.Context) (*domain.RepositoryMetrics, error)

	// ResetMetrics resets metrics
	ResetMetrics(ctx context.Context)
}
