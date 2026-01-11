package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
)

// StateRepository defines the interface for regulatory state data persistence.
// This interface handles all database operations for state entities.
type StateRepository interface {
	// Create adds a new state record to the repository
	Create(ctx context.Context, state *domain.RegulatoryState) error

	// GetByID retrieves a state record by its unique identifier
	GetByID(ctx context.Context, id string) (*domain.RegulatoryState, error)

	// GetByEntity retrieves the current state for a specific entity
	GetByEntity(ctx context.Context, entityType string, entityID string) (*domain.RegulatoryState, error)

	// GetStateHistory retrieves the complete history of state changes for an entity
	GetStateHistory(ctx context.Context, entityType string, entityID string) ([]*domain.StateChange, error)

	// GetByType retrieves all states of a specific state type
	GetByType(ctx context.Context, stateType domain.StateType) ([]*domain.RegulatoryState, error)

	// GetByStatus retrieves all states with a specific status
	GetByStatus(ctx context.Context, status domain.StateStatus) ([]*domain.RegulatoryState, error)

	// Update modifies an existing state record
	Update(ctx context.Context, state *domain.RegulatoryState) error

	// Delete removes a state record
	Delete(ctx context.Context, id string) error

	// List retrieves all state records with optional pagination
	List(ctx context.Context, filter *domain.StateFilter, offset, limit int) ([]*domain.RegulatoryState, error)

	// Count returns the total count of state records matching the filter
	Count(ctx context.Context, filter *domain.StateFilter) (int64, error)

	// GetCurrentState retrieves the current (latest) state for an entity
	GetCurrentState(ctx context.Context, entityType string, entityID string) (*domain.RegulatoryState, error)

	// GetPreviousState retrieves the previous state for an entity
	GetPreviousState(ctx context.Context, entityType string, entityID string) (*domain.RegulatoryState, error)

	// CompareStates compares two states and identifies differences
	CompareStates(ctx context.Context, stateID1, stateID2 string) (*domain.StateComparison, error)

	// GetStateTransitions retrieves all state transitions for analysis
	GetStateTransitions(ctx context.Context, filter *domain.StateTransitionFilter) ([]*domain.StateTransition, error)

	// CreateTransition creates a new state transition record
	CreateTransition(ctx context.Context, transition *domain.StateTransition) error

	// GetActiveAlerts retrieves all active alerts for states
	GetActiveAlerts(ctx context.Context) ([]*domain.StateAlert, error)

	// CreateAlert creates a new state alert
	CreateAlert(ctx context.Context, alert *domain.StateAlert) error

	// ResolveAlert resolves an existing alert
	ResolveAlert(ctx context.Context, alertID string, resolution string) error

	// GetStateMetrics retrieves state-related metrics
	GetStateMetrics(ctx context.Context, filter *domain.StateMetricsFilter) (*domain.StateMetrics, error)

	// BulkUpdateStates performs bulk updates on multiple state records
	BulkUpdateStates(ctx context.Context, updates []*domain.StateUpdate) error

	// Exists checks if a state record exists for an entity
	Exists(ctx context.Context, entityType string, entityID string) (bool, error)

	// GetSnapshot retrieves a snapshot of the state at a specific point in time
	GetSnapshot(ctx context.Context, timestamp time.Time, entityType, entityID string) (*domain.RegulatoryState, error)

	// CreateSnapshot creates a new state snapshot
	CreateSnapshot(ctx context.Context, state *domain.RegulatoryState, metadata *domain.SnapshotMetadata) error
}

// StateService defines the interface for state management business logic operations.
// This interface encapsulates all domain-specific operations for regulatory states.
type StateService interface {
	// CreateState creates a new regulatory state
	CreateState(ctx context.Context, input *domain.CreateStateInput) (*domain.RegulatoryState, error)

	// GetState retrieves a state by ID
	GetState(ctx context.Context, id string) (*domain.RegulatoryState, error)

	// GetEntityState retrieves the current state for an entity
	GetEntityState(ctx context.Context, entityType string, entityID string) (*domain.RegulatoryState, error)

	// ListStates retrieves states with filtering and pagination
	ListStates(ctx context.Context, filter *domain.StateFilter) ([]*domain.RegulatoryState, error)

	// UpdateState updates an existing state
	UpdateState(ctx context.Context, id string, input *domain.UpdateStateInput) (*domain.RegulatoryState, error)

	// TransitionState transitions an entity to a new state
	TransitionState(ctx context.Context, entityType string, entityID string, newStateType domain.StateType, reason string) (*domain.RegulatoryState, error)

	// BatchTransitionState transitions multiple entities to new states
	BatchTransitionState(ctx context.Context, transitions []*domain.BatchStateTransition) ([]*domain.RegulatoryState, error)

	// ValidateTransition validates if a state transition is allowed
	ValidateTransition(ctx context.Context, entityType string, entityID string, targetState domain.StateType) (*domain.ValidationResult, error)

	// GetStateHistory retrieves the history of state changes for an entity
	GetStateHistory(ctx context.Context, entityType string, entityID string) ([]*domain.StateChange, error)

	// GetStateTransitions retrieves state transition patterns
	GetStateTransitions(ctx context.Context, filter *domain.StateTransitionFilter) ([]*domain.StateTransition, error)

	// GetAvailableTransitions returns valid next states for an entity
	GetAvailableTransitions(ctx context.Context, entityType string, entityID string) ([]domain.StateType, error)

	// CompareStates compares two states
	CompareStates(ctx context.Context, stateID1, stateID2 string) (*domain.StateComparison, error)

	// SnapshotState creates a snapshot of the current state
	SnapshotState(ctx context.Context, entityType string, entityID string, description string) (*domain.RegulatoryState, error)

	// RestoreState restores a state from a snapshot
	RestoreState(ctx context.Context, snapshotID string) (*domain.RegulatoryState, error)

	// GetStateMetrics retrieves metrics about states
	GetStateMetrics(ctx context.Context, filter *domain.StateMetricsFilter) (*domain.StateMetrics, error)

	// GetStateAlerts retrieves all active alerts
	GetStateAlerts(ctx context.Context) ([]*domain.StateAlert, error)

	// CreateStateAlert creates a new alert
	CreateStateAlert(ctx context.Context, alert *domain.StateAlert) error

	// ResolveStateAlert resolves an existing alert
	ResolveStateAlert(ctx context.Context, alertID string, resolution string) error

	// AcknowledgeAlert acknowledges an alert
	AcknowledgeAlert(ctx context.Context, alertID string, userID string) error

	// GetPendingActions retrieves all actions pending for states
	GetPendingActions(ctx context.Context, filter *domain.PendingActionsFilter) ([]*domain.PendingAction, error)

	// CompleteAction marks an action as completed
	CompleteAction(ctx context.Context, actionID string, result string) error

	// ScheduleStateTransition schedules a future state transition
	ScheduleStateTransition(ctx context.Context, entityType string, entityID string, targetState domain.StateType, scheduledTime time.Time, reason string) error

	// CancelScheduledTransition cancels a scheduled transition
	CancelScheduledTransition(ctx context.Context, transitionID string) error

	// GetScheduledTransitions retrieves all scheduled transitions
	GetScheduledTransitions(ctx context.Context, filter *domain.ScheduledTransitionFilter) ([]*domain.ScheduledTransition, error)

	// AnalyzeStateTrends analyzes state transition trends
	AnalyzeStateTrends(ctx context.Context, entityType string, timeRange domain.TimeRange) (*domain.StateTrendAnalysis, error)
}

// StateEventPublisher defines the interface for publishing state-related events.
type StateEventPublisher interface {
	// PublishStateCreated publishes an event when a new state is created
	PublishStateCreated(ctx context.Context, state *domain.RegulatoryState) error

	// PublishStateChanged publishes an event when a state changes
	PublishStateChanged(ctx context.Context, change *domain.StateChange) error

	// PublishStateTransition publishes an event when a state transition occurs
	PublishStateTransition(ctx context.Context, transition *domain.StateTransition) error

	// PublishStateAlert publishes an event when a new alert is created
	PublishStateAlert(ctx context.Context, alert *domain.StateAlert) error

	// PublishStateAlertResolved publishes an event when an alert is resolved
	PublishStateAlertResolved(ctx context.Context, alert *domain.StateAlert) error

	// PublishSnapshotCreated publishes an event when a snapshot is created
	PublishSnapshotCreated(ctx context.Context, snapshot *domain.RegulatoryState) error

	// PublishScheduledTransition publishes an event about a scheduled transition
	PublishScheduledTransition(ctx context.Context, transition *domain.ScheduledTransition) error
}

// StateEventSubscriber defines the interface for consuming state-related events.
type StateEventSubscriber interface {
	// SubscribeToStateEvents starts consuming state-related events
	SubscribeToStateEvents(ctx context.Context) error

	// HandleStateEvent processes a single state event
	HandleStateEvent(ctx context.Context, event *domain.StateEvent) error

	// Close stops the event subscription
	Close() error
}

// StateCacheRepository defines the interface for caching state data.
type StateCacheRepository interface {
	// Get retrieves a cached state
	Get(ctx context.Context, entityType, entityID string) (*domain.RegulatoryState, error)

	// Set stores a state in the cache
	Set(ctx context.Context, state *domain.RegulatoryState, ttl time.Duration) error

	// Delete removes a state from the cache
	Delete(ctx context.Context, entityType, entityID string) error

	// Invalidate invalidates all cached states
	Invalidate(ctx context.Context) error

	// InvalidateByType invalidates all cached states of a type
	InvalidateByType(ctx context.Context, stateType domain.StateType) error

	// InvalidateByEntityType invalidates all cached states for an entity type
	InvalidateByEntityType(ctx context.Context, entityType string) error

	// GetCurrentState retrieves the current state for an entity
	GetCurrentState(ctx context.Context, entityType, entityID string) (*domain.RegulatoryState, error)

	// SetCurrentState stores the current state for an entity
	SetCurrentState(ctx context.Context, state *domain.RegulatoryState, ttl time.Duration) error
}

// StateAuditRepository defines the interface for audit trail persistence.
type StateAuditRepository interface {
	// CreateAuditEntry creates a new audit entry
	CreateAuditEntry(ctx context.Context, entry *domain.AuditEntry) error

	// GetAuditLog retrieves the audit log for an entity
	GetAuditLog(ctx context.Context, entityType, entityID string) ([]*domain.AuditEntry, error)

	// GetAuditLogByTimeRange retrieves the audit log within a time range
	GetAuditLogByTimeRange(ctx context.Context, entityType, entityID string, startTime, endTime time.Time) ([]*domain.AuditEntry, error)

	// GetAuditEntry retrieves a specific audit entry
	GetAuditEntry(ctx context.Context, id string) (*domain.AuditEntry, error)

	// QueryAuditLog queries the audit log with filters
	QueryAuditLog(ctx context.Context, query *domain.AuditQuery) ([]*domain.AuditEntry, error)

	// ExportAuditLog exports the audit log for compliance
	ExportAuditLog(ctx context.Context, query *domain.AuditQuery, format string) ([]byte, error)
}

// StateNotificationService defines the interface for state-related notifications.
type StateNotificationService interface {
	// NotifyStateChange sends a notification about a state change
	NotifyStateChange(ctx context.Context, change *domain.StateChange) error

	// NotifyStateAlert sends a notification about a state alert
	NotifyStateAlert(ctx context.Context, alert *domain.StateAlert) error

	// NotifyScheduledTransition sends a notification about a scheduled transition
	NotifyScheduledTransition(ctx context.Context, transition *domain.ScheduledTransition) error

	// GetNotificationPreferences retrieves notification preferences for a user
	GetNotificationPreferences(ctx context.Context, userID string) (*domain.NotificationPreferences, error)

	// UpdateNotificationPreferences updates notification preferences
	UpdateNotificationPreferences(ctx context.Context, userID string, preferences *domain.NotificationPreferences) error
}
