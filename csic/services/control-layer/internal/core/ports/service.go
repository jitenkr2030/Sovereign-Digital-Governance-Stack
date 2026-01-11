package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
)

// GenericService defines a generic interface for service operations.
// This interface can be used as a base for type-safe services.
type GenericService interface {
	// Execute executes a service operation
	Execute(ctx context.Context, request interface{}) (interface{}, error)

	// GetName returns the service name
	GetName() string

	// GetVersion returns the service version
	GetVersion() string
}

// CRUDService defines CRUD service operations.
type CRUDService interface {
	// Create creates a new entity
	Create(ctx context.Context, input interface{}) (interface{}, error)

	// Read retrieves an entity
	Read(ctx context.Context, id string) (interface{}, error)

	// Update updates an entity
	Update(ctx context.Context, id string, input interface{}) (interface{}, error)

	// Delete deletes an entity
	Delete(ctx context.Context, id string) error

	// List lists all entities
	List(ctx context.Context, offset, limit int) ([]interface{}, error)

	// Count counts all entities
	Count(ctx context.Context) (int64, error)
}

// QueryService defines read-only service operations.
type QueryService interface {
	// Get retrieves an entity by ID
	Get(ctx context.Context, id string) (interface{}, error)

	// Find finds entities matching criteria
	Find(ctx context.Context, criteria map[string]interface{}, offset, limit int) ([]interface{}, error)

	// FindOne finds one entity matching criteria
	FindOne(ctx context.Context, criteria map[string]interface{}) (interface{}, error)

	// List lists all entities
	List(ctx context.Context, offset, limit int) ([]interface{}, error)

	// Count counts entities matching criteria
	Count(ctx context.Context, criteria map[string]interface{}) (int64, error)

	// Exists checks if an entity exists
	Exists(ctx context.Context, id string) (bool, error)

	// Aggregate performs aggregation
	Aggregate(ctx context.Context, pipeline interface{}) ([]interface{}, error)
}

// CommandService defines write service operations.
type CommandService interface {
	// Handle handles a command
	Handle(ctx context.Context, command interface{}) (interface{}, error)

	// HandleAsync handles a command asynchronously
	HandleAsync(ctx context.Context, command interface{}) error

	// Validate validates a command
	Validate(ctx context.Context, command interface{}) error

	// GetCommandHandlers returns available command handlers
	GetCommandHandlers() map[string]domain.CommandHandler
}

// DomainService defines domain-specific business logic.
type DomainService interface {
	// GetDomainName returns the domain name
	GetDomainName() string

	// GetAggregateRoot returns the aggregate root type
	GetAggregateRoot() interface{}

	// ExecuteBusinessRule executes a business rule
	ExecuteBusinessRule(ctx context.Context, rule string, aggregate interface{}) error

	// ApplyInvariant applies an invariant
	ApplyInvariant(ctx context.Context, aggregate interface{}) error

	// GetInvariants returns all invariants for the aggregate
	GetInvariants() []domain.Invariant

	// GetBusinessRules returns all business rules
	GetBusinessRules() []domain.BusinessRule
}

// ApplicationService defines application-specific business logic.
type ApplicationService interface {
	// GetUseCases returns available use cases
	GetUseCases() []domain.UseCase

	// ExecuteUseCase executes a use case
	ExecuteUseCase(ctx context.Context, useCase string, input interface{}) (interface{}, error)

	// GetWorkflow returns a workflow by name
	GetWorkflow(ctx context.Context, name string) domain.Workflow

	// ExecuteWorkflow executes a workflow
	ExecuteWorkflow(ctx context.Context, name string, input interface{}) (interface{}, error)
}

// WorkflowService defines workflow operations.
type WorkflowService interface {
	// StartWorkflow starts a new workflow instance
	StartWorkflow(ctx context.Context, workflowID string, input map[string]interface{}) (string, error)

	// GetWorkflowStatus gets the status of a workflow instance
	GetWorkflowStatus(ctx context.Context, instanceID string) (*domain.WorkflowStatus, error)

	// SendWorkflowEvent sends an event to a workflow
	SendWorkflowEvent(ctx context.Context, instanceID string, event string, data map[string]interface{}) error

	// CancelWorkflow cancels a workflow instance
	CancelWorkflow(ctx context.Context, instanceID string, reason string) error

	// GetWorkflowTasks gets tasks for a workflow instance
	GetWorkflowTasks(ctx context.Context, instanceID string) ([]domain.WorkflowTask, error)

	// CompleteTask completes a workflow task
	CompleteTask(ctx context.Context, taskID string, output map[string]interface{}) error

	// FailTask fails a workflow task
	FailTask(ctx context.Context, taskID string, error string) error

	// ListWorkflowDefinitions lists workflow definitions
	ListWorkflowDefinitions(ctx context.Context) ([]domain.WorkflowDefinition, error)

	// GetWorkflowDefinition gets a workflow definition
	GetWorkflowDefinition(ctx context.Context, workflowID string) (*domain.WorkflowDefinition, error)

	// RegisterWorkflowDefinition registers a workflow definition
	RegisterWorkflowDefinition(ctx context.Context, definition *domain.WorkflowDefinition) error
}

// SagaService defines saga pattern operations for distributed transactions.
type SagaService interface {
	// StartSaga starts a new saga
	StartSaga(ctx context.Context, sagaID string, name string, data map[string]interface{}) error

	// GetSagaStatus gets the status of a saga
	GetSagaStatus(ctx context.Context, sagaID string) (*domain.SagaStatus, error)

	// SendSagaEvent sends an event to a saga
	SendSagaEvent(ctx context.Context, sagaID string, event string, data map[string]interface{}) error

	// CompensateSaga compensates a saga
	CompensateSaga(ctx context.Context, sagaID string, reason string) error

	// CompleteSaga completes a saga successfully
	CompleteSaga(ctx context.Context, sagaID string) error

	// GetSagaHistory gets the history of a saga
	GetSagaHistory(ctx context.Context, sagaID string) ([]domain.SagaStep, error)

	// ListSagas lists all sagas
	ListSagas(ctx context.Context, filter *domain.SagaFilter) ([]*domain.SagaStatus, error)
}

// EventDrivenService defines event-driven service operations.
type EventDrivenService interface {
	// PublishEvent publishes an event
	PublishEvent(ctx context.Context, event domain.Event) error

	// PublishEventTo publishes an event to a specific topic
	PublishEventTo(ctx context.Context, topic string, event domain.Event) error

	// SubscribeToEvents subscribes to event types
	SubscribeToEvents(ctx context.Context, eventTypes []string, handler domain.EventHandler) error

	// UnsubscribeFromEvents unsubscribes from event types
	UnsubscribeFromEvents(ctx context.Context, eventTypes []string, handlerID string) error

	// GetEventHandlers returns all event handlers
	GetEventHandlers() map[string][]domain.EventHandler

	// RegisterEventHandler registers an event handler
	RegisterEventHandler(ctx context.Context, eventType string, handler domain.EventHandler) error
}

// AggregateService defines aggregate operations for domain-driven design.
type AggregateService interface {
	// LoadAggregate loads an aggregate by ID
	LoadAggregate(ctx context.Context, aggregateType string, id string) (interface{}, error)

	// SaveAggregate saves an aggregate
	SaveAggregate(ctx context.Context, aggregate interface{}) error

	// DeleteAggregate deletes an aggregate
	DeleteAggregate(ctx context.Context, aggregateType string, id string) error

	// ApplyEvent applies an event to an aggregate
	ApplyEvent(ctx context.Context, aggregate interface{}, event domain.Event) error

	// GetAggregateHistory gets the event history of an aggregate
	GetAggregateHistory(ctx context.Context, aggregateType string, id string) ([]domain.Event, error)

	// GetAggregateVersion gets the current version of an aggregate
	GetAggregateVersion(ctx context.Context, aggregateType string, id string) (int64, error)
}

// ServiceFactory defines factory methods for creating services.
type ServiceFactory interface {
	// Create creates a new service instance
	Create(ctx context.Context, serviceType interface{}) interface{}

	// GetService gets a service by type
	GetService(ctx context.Context, serviceType interface{}) interface{}

	// Register registers a service
	Register(ctx context.Context, serviceType interface{}, service interface{})

	// GetRegisteredServices gets all registered services
	GetRegisteredServices(ctx context.Context) map[string]interface{}
}

// ServiceMiddleware defines middleware for service operations.
type ServiceMiddleware interface {
	// Wrap wraps a service with middleware
	Wrap(service interface{}) interface{}

	// PreProcess pre-processes a service operation
	PreProcess(ctx context.Context, serviceName string, operation string, input interface{}) error

	// PostProcess post-processes a service operation
	PostProcess(ctx context.Context, serviceName string, operation string, input interface{}, output interface{}, err error) error

	// OnError handles errors from service operations
	OnError(ctx context.Context, serviceName string, operation string, err error) error
}

// ServiceHealthChecker defines health check operations for services.
type ServiceHealthChecker interface {
	// HealthCheck performs a health check
	HealthCheck(ctx context.Context) (*domain.HealthStatus, error)

	// GetDependencies gets service dependencies
	GetDependencies(ctx context.Context) []string

	// CheckDependency checks a dependency
	CheckDependency(ctx context.Context, name string) (*domain.DependencyHealth, error)

	// GetMetrics gets service metrics
	GetMetrics(ctx context.Context) (*domain.ServiceMetrics, error)
}

// ServiceMetrics defines metrics collection for services.
type ServiceMetrics interface {
	// RecordOperation records a service operation
	RecordOperation(ctx context.Context, serviceName string, operation string, duration time.Duration, success bool)

	// RecordRequest records a request
	RecordRequest(ctx context.Context, serviceName string, status string)

	// RecordError records an error
	RecordError(ctx context.Context, serviceName string, errorType string)

	// ObserveGauge observes a gauge value
	ObserveGauge(ctx context.Context, serviceName string, metric string, value float64)

	// GetMetrics gets service metrics
	GetMetrics(ctx context.Context, serviceName string) (*domain.ServiceMetrics, error)

	// ResetMetrics resets metrics
	ResetMetrics(ctx context.Context, serviceName string)
}

// ServiceLogger defines logging operations for services.
type ServiceLogger interface {
	// Log logs a message
	Log(ctx context.Context, level string, serviceName string, message string, fields map[string]interface{})

	// LogOperation logs a service operation
	LogOperation(ctx context.Context, serviceName string, operation string, duration time.Duration, success bool, fields map[string]interface{})

	// LogError logs an error
	LogError(ctx context.Context, serviceName string, operation string, err error, fields map[string]interface{})

	// GetLogger returns a logger for a specific service
	GetLogger(serviceName string) Logger
}

// ServiceConfig defines configuration operations for services.
type ServiceConfig interface {
	// GetConfig gets service configuration
	GetConfig(ctx context.Context, serviceName string) (map[string]interface{}, error)

	// UpdateConfig updates service configuration
	UpdateConfig(ctx context.Context, serviceName string, config map[string]interface{}) error

	// ReloadConfig reloads service configuration
	ReloadConfig(ctx context.Context, serviceName string) error

	// GetFeatureFlags gets feature flags for a service
	GetFeatureFlags(ctx context.Context, serviceName string) (map[string]bool, error)

	// SetFeatureFlag sets a feature flag
	SetFeatureFlag(ctx context.Context, serviceName string, flag string, enabled bool) error
}

// ServiceRegistry defines service registration and discovery operations.
type ServiceRegistry interface {
	// Register registers a service
	Register(ctx context.Context, service *domain.ServiceRegistration) error

	// Deregister deregisters a service
	Deregister(ctx context.Context, serviceID string) error

	// GetService gets a service by ID
	GetService(ctx context.Context, serviceID string) (*domain.ServiceRegistration, error)

	// GetServices gets all registered services
	GetServices(ctx context.Context) ([]*domain.ServiceRegistration, error)

	// FindServices finds services by name
	FindServices(ctx context.Context, name string) ([]*domain.ServiceRegistration, error)

	// Heartbeat sends a heartbeat
	Heartbeat(ctx context.Context, serviceID string) error

	// GetServiceHealth gets the health of a service
	GetServiceHealth(ctx context.Context, serviceID string) (*domain.HealthStatus, error)
}

// ServiceDiscovery defines service discovery operations.
type ServiceDiscovery interface {
	// Discover discovers a service
	Discover(ctx context.Context, serviceName string) ([]string, error)

	// DiscoverOne discovers a single service instance
	DiscoverOne(ctx context.Context, serviceName string) (string, error)

	// DiscoverWithFilter discovers services with filters
	DiscoverWithFilter(ctx context.Context, serviceName string, filters map[string]string) ([]string, error)

	// GetServiceInstances gets all instances of a service
	GetServiceInstances(ctx context.Context, serviceName string) ([]*domain.ServiceInstance, error)

	// RegisterInstance registers a service instance
	RegisterInstance(ctx context.Context, instance *domain.ServiceInstance) error

	// DeregisterInstance deregisters a service instance
	DeregisterInstance(ctx context.Context, instanceID string) error

	// GetAllServices gets all registered services
	GetAllServices(ctx context.Context) ([]string, error)
}

// ServiceLoadBalancer defines load balancing operations for services.
type ServiceLoadBalancer interface {
	// Select selects a service instance
	Select(ctx context.Context, serviceName string) (string, error)

	// SelectWithContext selects a service instance with context
	SelectWithContext(ctx context.Context, serviceName string, ctx interface{}) (string, error)

	// AddInstance adds a service instance
	AddInstance(ctx context.Context, serviceName string, instance string)

	// RemoveInstance removes a service instance
	RemoveInstance(ctx context.Context, serviceName string, instance string)

	// GetInstances gets all instances for a service
	GetInstances(ctx context.Context, serviceName string) ([]string, error)

	// GetStats gets load balancer statistics
	GetStats(ctx context.Context, serviceName string) (*domain.LoadBalancerStats, error)
}

// RateLimitService defines rate limiting operations.
type RateLimitService interface {
	// Allow checks if an action is allowed
	Allow(ctx context.Context, key string, operation string) (bool, error)

	// Consume consumes a rate limit token
	Consume(ctx context.Context, key string, operation string) (bool, error)

	// GetLimit gets the rate limit for a key
	GetLimit(ctx context.Context, key string, operation string) (*domain.RateLimit, error)

	// SetLimit sets a rate limit
	SetLimit(ctx context.Context, key string, operation string, limit *domain.RateLimit) error

	// ResetLimit resets a rate limit
	ResetLimit(ctx context.Context, key string, operation string) error

	// GetRateLimitStats gets rate limit statistics
	GetRateLimitStats(ctx context.Context) (*domain.RateLimitStats, error)
}

// CircuitBreakerService defines circuit breaker operations.
type CircuitBreakerService interface {
	// Execute executes a function with circuit breaker protection
	Execute(ctx context.Context, key string, fn func() error) error

	// ExecuteWithResult executes a function with circuit breaker protection and returns a result
	ExecuteWithResult(ctx context.Context, key string, fn func() (interface{}, error)) (interface{}, error)

	// GetState gets the circuit breaker state
	GetState(ctx context.Context, key string) (domain.CircuitState, error)

	// Reset resets the circuit breaker
	Reset(ctx context.Context, key string) error

	// ForceOpen forces the circuit breaker to open
	ForceOpen(ctx context.Context, key string) error

	// ForceClose forces the circuit breaker to close
	ForceClose(ctx context.Context, key string) error

	// GetMetrics gets circuit breaker metrics
	GetMetrics(ctx context.Context, key string) (*domain.CircuitBreakerMetrics, error)
}

// BulkheadService defines bulkhead isolation operations.
type BulkheadService interface {
	// Execute executes a function with bulkhead isolation
	Execute(ctx context.Context, key string, fn func() error) error

	// ExecuteWithResult executes a function with bulkhead isolation and returns a result
	ExecuteWithResult(ctx context.Context, key string, fn func() (interface{}, error)) (interface{}, error)

	// GetStats gets bulkhead statistics
	GetStats(ctx context.Context, key string) (*domain.BulkheadStats, error)

	// UpdateConfig updates bulkhead configuration
	UpdateConfig(ctx context.Context, key string, config *domain.BulkheadConfig) error
}

// RetryService defines retry operations.
type RetryService interface {
	// Execute executes a function with retry
	Execute(ctx context.Context, fn func() error, opts ...RetryOption) error

	// ExecuteWithResult executes a function with retry and returns a result
	ExecuteWithResult(ctx context.Context, fn func() (interface{}, error), opts ...RetryOption) (interface{}, error)

	// GetRetryConfig gets the retry configuration
	GetRetryConfig() *domain.RetryConfig

	// UpdateConfig updates the retry configuration
	UpdateConfig(config *domain.RetryConfig)
}

// RetryOption defines retry configuration options.
type RetryOption interface {
	Apply(config *domain.RetryConfig)
}

// TimeoutService defines timeout operations.
type TimeoutService interface {
	// Execute executes a function with timeout
	Execute(ctx context.Context, fn func() error, timeout time.Duration) error

	// ExecuteWithResult executes a function with timeout and returns a result
	ExecuteWithResult(ctx context.Context, fn func() (interface{}, error), timeout time.Duration) (interface{}, error)

	// ExecuteWithContext executes a function with timeout from context
	ExecuteWithContext(ctx context.Context, fn func(ctx context.Context) error) error

	// ExecuteWithContextAndResult executes a function with timeout from context and returns a result
	ExecuteWithContextAndResult(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)

	// GetDefaultTimeout gets the default timeout
	GetDefaultTimeout() time.Duration

	// SetDefaultTimeout sets the default timeout
	SetDefaultTimeout(timeout time.Duration)
}
