package ports

import (
	"context"
	"time"

	"github.com/csic-platform/internal/core/domain"
)

// UnitOfWork defines the interface for managing transactional operations.
// This interface ensures data consistency across multiple repository operations.
type UnitOfWork interface {
	// Begin starts a new transaction
	Begin(ctx context.Context) error

	// Commit commits the current transaction
	Commit(ctx context.Context) error

	// Rollback rolls back the current transaction
	Rollback(ctx context.Context) error

	// Complete completes the transaction based on the outcome
	Complete(ctx context.Context, err error) error

	// GetTransaction returns the current transaction
	GetTransaction(ctx context.Context) interface{}

	// Register registers a repository to participate in the transaction
	Register(repository interface{}) error

	// GetRepository gets a registered repository
	GetRepository(repositoryType interface{}) interface{}
}

// Repository defines the base interface for all repositories.
// This interface provides common CRUD operations applicable to all entities.
type Repository[T any] interface {
	// Create adds a new entity to the repository
	Create(ctx context.Context, entity *T) error

	// GetByID retrieves an entity by its unique identifier
	GetByID(ctx context.Context, id string) (*T, error)

	// Update modifies an existing entity
	Update(ctx context.Context, entity *T) error

	// Delete removes an entity
	Delete(ctx context.Context, id string) error

	// Exists checks if an entity exists
	Exists(ctx context.Context, id string) (bool, error)

	// List retrieves all entities with optional pagination
	List(ctx context.Context, offset, limit int) ([]*T, error)

	// Count returns the total count of entities
	Count(ctx context.Context) (int64, error)
}

// SearchableRepository extends Repository with search capabilities.
type SearchableRepository[T any] interface {
	Repository[T]

	// Search performs a text search on entities
	Search(ctx context.Context, query string, offset, limit int) ([]*T, error)

	// Suggest provides search suggestions
	Suggest(ctx context.Context, query string, limit int) ([]string, error)
}

// PaginatedRepository extends Repository with pagination support.
type PaginatedRepository[T any] interface {
	Repository[T]

	// ListWithFilter retrieves entities matching the filter with pagination
	ListWithFilter(ctx context.Context, filter interface{}, offset, limit int) ([]*T, error)

	// CountWithFilter returns the count of entities matching the filter
	CountWithFilter(ctx context.Context, filter interface{}) (int64, error)
}

// EventPublisher defines the base interface for publishing domain events.
type EventPublisher interface {
	// Publish publishes a domain event
	Publish(ctx context.Context, event domain.Event) error

	// PublishBatch publishes multiple domain events
	PublishBatch(ctx context.Context, events []domain.Event) error

	// Flush flushes any pending events
	Flush(ctx context.Context) error

	// Close closes the publisher
	Close() error
}

// EventSubscriber defines the base interface for consuming domain events.
type EventSubscriber interface {
	// Subscribe subscribes to events of a specific type
	Subscribe(ctx context.Context, eventType string, handler interface{}) error

	// Unsubscribe unsubscribes from events
	Unsubscribe(eventType string, handlerID string) error

	// Start starts consuming events
	Start(ctx context.Context) error

	// Stop stops consuming events
	Stop(ctx context.Context) error
}

// CacheRepository defines the interface for cached repository operations.
type CacheRepository interface {
	// Get retrieves a cached entity
	Get(ctx context.Context, key string) (interface{}, error)

	// Set stores an entity in the cache
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes an entity from the cache
	Delete(ctx context.Context, key string) error

	// Invalidate invalidates all cached entities matching the pattern
	Invalidate(ctx context.Context, pattern string) error

	// GetStats returns cache statistics
	GetStats(ctx context.Context) (*domain.CacheStats, error)

	// Clear clears all cached data
	Clear(ctx context.Context) error
}

// SearchRepository defines the interface for search operations.
type SearchRepository interface {
	// Index indexes a document
	Index(ctx context.Context, id string, document interface{}) error

	// BulkIndex indexes multiple documents
	BulkIndex(ctx context.Context, documents []interface{}) error

	// DeleteIndex removes a document from the index
	DeleteIndex(ctx context.Context, id string) error

	// Search performs a search query
	Search(ctx context.Context, query string, filters map[string]interface{}, offset, limit int) ([]interface{}, error)

	// Suggest provides search suggestions
	Suggest(ctx context.Context, field string, query string, limit int) ([]string, error)

	// Reindex rebuilds the search index
	Reindex(ctx context.Context) error

	// ReindexOne rebuilds the index for a single document
	ReindexOne(ctx context.Context, id string) error

	// GetIndexStats returns index statistics
	GetIndexStats(ctx context.Context) (*domain.SearchIndexStats, error)
}

// MetricsCollector defines the interface for collecting metrics.
type MetricsCollector interface {
	// Increment increments a counter metric
	Increment(ctx context.Context, name string, labels map[string]string) error

	// Gauge sets a gauge metric
	Gauge(ctx context.Context, name string, value float64, labels map[string]string) error

	// Histogram records a histogram metric
	Histogram(ctx context.Context, name string, value float64, labels map[string]string) error

	// Summary records a summary metric
	Summary(ctx context.Context, name string, value float64, labels map[string]string) error

	// Timing records a timing metric
	Timing(ctx context.Context, name string, duration time.Duration, labels map[string]string) error

	// GetCounter gets the value of a counter metric
	GetCounter(ctx context.Context, name string, labels map[string]string) (float64, error)

	// GetGauge gets the value of a gauge metric
	GetGauge(ctx context.Context, name string, labels map[string]string) (float64, error)
}

// Logger defines the interface for logging operations.
type Logger interface {
	// Debug logs a debug message
	Debug(ctx context.Context, message string, fields map[string]interface{})

	// Info logs an info message
	Info(ctx context.Context, message string, fields map[string]interface{})

	// Warn logs a warning message
	Warn(ctx context.Context, message string, fields map[string]interface{})

	// Error logs an error message
	Error(ctx context.Context, message string, fields map[string]interface{})

	// Fatal logs a fatal message and exits
	Fatal(ctx context.Context, message string, fields map[string]interface{})

	// WithFields returns a new logger with additional fields
	WithFields(fields map[string]interface{}) Logger

	// WithError returns a new logger with an error field
	WithError(err error) Logger

	// WithRequestID returns a new logger with a request ID field
	WithRequestID(requestID string) Logger
}

// ConfigurationLoader defines the interface for loading configuration.
type ConfigurationLoader interface {
	// Load loads configuration from the specified source
	Load(ctx context.Context, source string) (map[string]interface{}, error)

	// LoadFile loads configuration from a file
	LoadFile(ctx context.Context, path string) (map[string]interface{}, error)

	// LoadEnv loads configuration from environment variables
	LoadEnv(ctx context.Context, prefix string) (map[string]interface{}, error)

	// Get gets a configuration value
	Get(ctx context.Context, key string) (interface{}, error)

	// GetString gets a string configuration value
	GetString(ctx context.Context, key string) (string, error)

	// GetInt gets an int configuration value
	GetInt(ctx context.Context, key string) (int, error)

	// GetFloat gets a float configuration value
	GetFloat(ctx context.Context, key string) (float64, error)

	// GetBool gets a bool configuration value
	GetBool(ctx context.Context, key string) (bool, error)

	// GetSlice gets a slice configuration value
	GetSlice(ctx context.Context, key string) ([]interface{}, error)

	// GetMap gets a map configuration value
	GetMap(ctx context.Context, key string) (map[string]interface{}, error)

	// Set sets a configuration value
	Set(ctx context.Context, key string, value interface{}) error

	// Reload reloads configuration
	Reload(ctx context.Context) error
}

// HealthChecker defines the interface for health check operations.
type HealthChecker interface {
	// Check performs a health check
	Check(ctx context.Context) (*domain.HealthStatus, error)

	// CheckDependency checks a dependency's health
	CheckDependency(ctx context.Context, dependency string) (*domain.DependencyHealth, error)

	// GetDependencies returns all dependencies
	GetDependencies(ctx context.Context) ([]string, error)

	// RegisterDependency registers a dependency
	RegisterDependency(ctx context.Context, name string, checker interface{}) error
}

// EventHandler defines the interface for event handlers.
type EventHandler interface {
	// Handle handles an event
	Handle(ctx context.Context, event domain.Event) error

	// Priority returns the handler priority
	Priority() int

	// EventTypes returns the event types this handler handles
	EventTypes() []string
}

// QueryBus defines the interface for query dispatching.
type QueryBus interface {
	// Register registers a query handler
	Register(queryType interface{}, handler interface{}) error

	// Execute executes a query
	Execute(ctx context.Context, query interface{}) (interface{}, error)
}

// CommandBus defines the interface for command dispatching.
type CommandBus interface {
	// Register registers a command handler
	Register(commandType interface{}, handler interface{}) error

	// Execute executes a command
	Execute(ctx context.Context, command interface{}) (interface{}, error)

	// ExecuteAsync executes a command asynchronously
	ExecuteAsync(ctx context.Context, command interface{}) error
}

// NotificationService defines the interface for sending notifications.
type NotificationService interface {
	// Send sends a notification
	Send(ctx context.Context, notification *domain.Notification) error

	// SendBatch sends multiple notifications
	SendBatch(ctx context.Context, notifications []*domain.Notification) error

	// GetTemplate gets a notification template
	GetTemplate(ctx context.Context, templateID string) (*domain.NotificationTemplate, error)

	// CreateTemplate creates a notification template
	CreateTemplate(ctx context.Context, template *domain.NotificationTemplate) error

	// UpdateTemplate updates a notification template
	UpdateTemplate(ctx context.Context, template *domain.NotificationTemplate) error

	// DeleteTemplate deletes a notification template
	DeleteTemplate(ctx context.Context, templateID string) error

	// ListTemplates lists all notification templates
	ListTemplates(ctx context.Context) ([]*domain.NotificationTemplate, error)
}

// FileStorage defines the interface for file storage operations.
type FileStorage interface {
	// Upload uploads a file
	Upload(ctx context.Context, path string, content []byte, metadata map[string]string) error

	// Download downloads a file
	Download(ctx context.Context, path string) ([]byte, error)

	// Delete deletes a file
	Delete(ctx context.Context, path string) error

	// Exists checks if a file exists
	Exists(ctx context.Context, path string) (bool, error)

	// List lists files in a directory
	List(ctx context.Context, path string) ([]*domain.FileInfo, error)

	// GetMetadata gets file metadata
	GetMetadata(ctx context.Context, path string) (*domain.FileMetadata, error)

	// GetSignedURL gets a signed URL for file access
	GetSignedURL(ctx context.Context, path string, expiration time.Duration) (string, error)

	// Copy copies a file
	Copy(ctx context.Context, sourcePath, destPath string) error

	// Move moves a file
	Move(ctx context.Context, sourcePath, destPath string) error
}

// EmailService defines the interface for email operations.
type EmailService interface {
	// Send sends an email
	Send(ctx context.Context, email *domain.Email) error

	// SendBatch sends multiple emails
	SendBatch(ctx context.Context, emails []*domain.Email) error

	// GetTemplate gets an email template
	GetTemplate(ctx context.Context, templateID string) (*domain.EmailTemplate, error)

	// CreateTemplate creates an email template
	CreateTemplate(ctx context.Context, template *domain.EmailTemplate) error

	// UpdateTemplate updates an email template
	UpdateTemplate(ctx context.Context, template *domain.EmailTemplate) error

	// DeleteTemplate deletes an email template
	DeleteTemplate(ctx context.Context, templateID string) error
}

// SMSService defines the interface for SMS operations.
type SMSService interface {
	// Send sends an SMS message
	Send(ctx context.Context, sms *domain.SMS) error

	// SendBatch sends multiple SMS messages
	SendBatch(ctx context.Context, smsList []*domain.SMS) error

	// GetStatus gets the status of an SMS message
	GetStatus(ctx context.Context, messageID string) (*domain.SMSStatus, error)
}

// WebhookService defines the interface for webhook operations.
type WebhookService interface {
	// Register registers a webhook
	Register(ctx context.Context, webhook *domain.Webhook) error

	// Unregister unregisters a webhook
	Unregister(ctx context.Context, webhookID string) error

	// Update updates a webhook
	Update(ctx context.Context, webhook *domain.Webhook) error

	// Get gets a webhook
	Get(ctx context.Context, webhookID string) (*domain.Webhook, error)

	// List lists all webhooks
	List(ctx context.Context, eventType string) ([]*domain.Webhook, error)

	// Trigger triggers a webhook manually
	Trigger(ctx context.Context, webhookID string, payload interface{}) error

	// Test tests a webhook
	Test(ctx context.Context, webhookID string) error

	// GetLogs gets webhook delivery logs
	GetLogs(ctx context.Context, webhookID string, offset, limit int) ([]*domain.WebhookLog, error)
}

// Scheduler defines the interface for scheduling operations.
type Scheduler interface {
	// Schedule schedules a job
	Schedule(ctx context.Context, job *domain.Job) error

	// ScheduleCron schedules a cron job
	ScheduleCron(ctx context.Context, job *domain.Job, cronExpression string) error

	// Unschedule unschedules a job
	Unschedule(ctx context.Context, jobID string) error

	// GetJob gets a job
	GetJob(ctx context.Context, jobID string) (*domain.Job, error)

	// ListJobs lists all jobs
	ListJobs(ctx context.Context) ([]*domain.Job, error)

	// PauseJob pauses a job
	PauseJob(ctx context.Context, jobID string) error

	// ResumeJob resumes a job
	ResumeJob(ctx context.Context, jobID string) error

	// TriggerJob triggers a job manually
	TriggerJob(ctx context.Context, jobID string) error

	// GetJobLogs gets job execution logs
	GetJobLogs(ctx context.Context, jobID string, offset, limit int) ([]*domain.JobLog, error)
}

// RateLimiter defines the interface for rate limiting operations.
type RateLimiter interface {
	// Allow checks if an action is allowed
	Allow(ctx context.Context, key string) (bool, error)

	// Consume consumes a rate limit token
	Consume(ctx context.Context, key string) (bool, error)

	// Reset resets the rate limit for a key
	Reset(ctx context.Context, key string) error

	// GetLimit gets the rate limit for a key
	GetLimit(ctx context.Context, key string) (*domain.RateLimit, error)

	// SetLimit sets a custom rate limit
	SetLimit(ctx context.Context, key string, limit *domain.RateLimit) error
}

// CircuitBreaker defines the interface for circuit breaker operations.
type CircuitBreaker interface {
	// Execute executes a function with circuit breaker protection
	Execute(ctx context.Context, name string, fn func() error) error

	// State returns the current state of the circuit breaker
	State(ctx context.Context, name string) (domain.CircuitState, error)

	// Reset resets the circuit breaker
	Reset(ctx context.Context, name string) error

	// GetMetrics gets circuit breaker metrics
	GetMetrics(ctx context.Context, name string) (*domain.CircuitBreakerMetrics, error)
}

// PaginationService defines the interface for pagination operations.
type PaginationService interface {
	// Paginate paginates a query result
	Paginate(ctx context.Context, query interface{}, page, pageSize int) (*domain.PaginatedResult, error)

	// CalculatePages calculates the total number of pages
	CalculatePages(ctx context.Context, totalItems, pageSize int) int

	// CreatePaginationParams creates pagination parameters from request
	CreatePaginationParams(ctx context.Context, page, pageSize int, sortBy, sortOrder string) *domain.PaginationParams
}

// ValidationService defines the interface for validation operations.
type ValidationService interface {
	// Validate validates an object
	Validate(ctx context.Context, object interface{}) (*domain.ValidationResult, error)

	// ValidateStruct validates a struct
	ValidateStruct(ctx context.Context, structPtr interface{}) ([]*domain.ValidationError, error)

	// ValidateJSON validates a JSON object against a schema
	ValidateJSON(ctx context.Context, jsonData []byte, schemaID string) (*domain.ValidationResult, error)

	// RegisterValidator registers a custom validator
	RegisterValidator(validatorType interface{}, validator interface{}) error

	// GetValidator gets a registered validator
	GetValidator(validatorType interface{}) interface{}
}
