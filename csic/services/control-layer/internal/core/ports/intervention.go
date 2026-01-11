package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
)

// InterventionRepository defines the interface for intervention data persistence.
// This interface handles all database operations for intervention entities.
type InterventionRepository interface {
	// Create adds a new intervention to the repository
	Create(ctx context.Context, intervention *domain.Intervention) error

	// GetByID retrieves an intervention by its unique identifier
	GetByID(ctx context.Context, id string) (*domain.Intervention, error)

	// GetByEntity retrieves all interventions for a specific entity
	GetByEntity(ctx context.Context, entityType string, entityID string) ([]*domain.Intervention, error)

	// GetByPolicy retrieves all interventions related to a specific policy
	GetByPolicy(ctx context.Context, policyID string) ([]*domain.Intervention, error)

	// GetByType retrieves all interventions of a specific type
	GetByType(ctx context.Context, interventionType domain.InterventionType) ([]*domain.Intervention, error)

	// GetByStatus retrieves all interventions with the specified status
	GetByStatus(ctx context.Context, status domain.InterventionStatus) ([]*domain.Intervention, error)

	// GetByInitiator retrieves all interventions initiated by a specific user
	GetByInitiator(ctx context.Context, initiatorID string) ([]*domain.Intervention, error)

	// GetByAssignee retrieves all interventions assigned to a specific user
	GetByAssignee(ctx context.Context, assigneeID string) ([]*domain.Intervention, error)

	// Update modifies an existing intervention
	Update(ctx context.Context, intervention *domain.Intervention) error

	// Delete removes an intervention
	Delete(ctx context.Context, id string) error

	// List retrieves all interventions with optional pagination
	List(ctx context.Context, filter *domain.InterventionFilter, offset, limit int) ([]*domain.Intervention, error)

	// Count returns the total count of interventions matching the filter
	Count(ctx context.Context, filter *domain.InterventionFilter) (int64, error)

	// GetTimeline retrieves the complete timeline for an intervention
	GetTimeline(ctx context.Context, interventionID string) ([]*domain.InterventionTimelineEvent, error)

	// AddTimelineEvent adds a new event to the intervention timeline
	AddTimelineEvent(ctx context.Context, interventionID string, event *domain.InterventionTimelineEvent) error

	// GetDocuments retrieves all documents attached to an intervention
	GetDocuments(ctx context.Context, interventionID string) ([]*domain.Document, error)

	// AttachDocument attaches a document to an intervention
	AttachDocument(ctx context.Context, interventionID string, document *domain.Document) error

	// DetachDocument removes a document from an intervention
	DetachDocument(ctx context.Context, interventionID string, documentID string) error

	// GetActions retrieves all actions for an intervention
	GetActions(ctx context.Context, interventionID string) ([]*domain.InterventionAction, error)

	// AddAction adds an action to an intervention
	AddAction(ctx context.Context, interventionID string, action *domain.InterventionAction) error

	// UpdateAction updates an action for an intervention
	UpdateAction(ctx context.Context, interventionID string, actionID string, action *domain.InterventionAction) error

	// CompleteAction marks an action as completed
	CompleteAction(ctx context.Context, interventionID string, actionID string, result string) error

	// GetOutcomes retrieves all outcomes for an intervention
	GetOutcomes(ctx context.Context, interventionID string) ([]*domain.Outcome, error)

	// AddOutcome adds an outcome to an intervention
	AddOutcome(ctx context.Context, interventionID string, outcome *domain.Outcome) error

	// GetConditions retrieves all conditions for an intervention
	GetConditions(ctx context.Context, interventionID string) ([]*domain.Condition, error)

	// AddCondition adds a condition to an intervention
	AddCondition(ctx context.Context, interventionID string, condition *domain.Condition) error

	// UpdateCondition updates a condition
	UpdateCondition(ctx context.Context, interventionID string, conditionID string, condition *domain.Condition) error

	// VerifyCondition verifies a condition as met or not met
	VerifyCondition(ctx context.Context, interventionID string, conditionID string, verified bool, notes string) error

	// GetHistory retrieves the complete history of changes for an intervention
	GetHistory(ctx context.Context, interventionID string) ([]*domain.HistoryEntry, error)

	// Exists checks if an intervention exists
	Exists(ctx context.Context, id string) (bool, error)

	// Search performs a text search on intervention records
	Search(ctx context.Context, query string, offset, limit int) ([]*domain.Intervention, error)

	// BulkUpdateStatus updates the status of multiple interventions
	BulkUpdateStatus(ctx context.Context, ids []string, status domain.InterventionStatus) error

	// BulkAssign updates the assignee for multiple interventions
	BulkAssign(ctx context.Context, ids []string, assigneeID string) error

	// GetEffectivenessMetrics retrieves effectiveness metrics for interventions
	GetEffectivenessMetrics(ctx context.Context, filter *domain.EffectivenessFilter) (*domain.EffectivenessMetrics, error)
}

// InterventionService defines the interface for intervention business logic operations.
// This interface encapsulates all domain-specific operations for regulatory interventions.
type InterventionService interface {
	// CreateIntervention creates a new intervention
	CreateIntervention(ctx context.Context, input *domain.CreateInterventionInput) (*domain.Intervention, error)

	// GetIntervention retrieves an intervention by ID
	GetIntervention(ctx context.Context, id string) (*domain.Intervention, error)

	// ListInterventions retrieves interventions with filtering and pagination
	ListInterventions(ctx context.Context, filter *domain.InterventionFilter) ([]*domain.Intervention, error)

	// UpdateIntervention updates an existing intervention
	UpdateIntervention(ctx context.Context, id string, input *domain.UpdateInterventionInput) (*domain.Intervention, error)

	// InitiateIntervention initiates an intervention process
	InitiateIntervention(ctx context.Context, id string) (*domain.Intervention, error)

	// ExecuteIntervention executes an intervention action
	ExecuteIntervention(ctx context.Context, id string, actionType domain.InterventionActionType, parameters map[string]interface{}) (*domain.Intervention, error)

	// SuspendIntervention suspends an ongoing intervention
	SuspendIntervention(ctx context.Context, id string, reason string) (*domain.Intervention, error)

	// ResumeIntervention resumes a suspended intervention
	ResumeIntervention(ctx context.Context, id string) (*domain.Intervention, error)

	// CompleteIntervention marks an intervention as completed
	CompleteIntervention(ctx context.Context, id string, outcome *domain.Outcome) (*domain.Intervention, error)

	// CancelIntervention cancels an intervention
	CancelIntervention(ctx context.Context, id string, reason string) (*domain.Intervention, error)

	// AssignIntervention assigns an intervention to a handler
	AssignIntervention(ctx context.Context, id string, assigneeID string) (*domain.Intervention, error)

	// ReassignIntervention reassigns an intervention to a different handler
	ReassignIntervention(ctx context.Context, id string, newAssigneeID string, reason string) (*domain.Intervention, error)

	// AddAction adds an action to an intervention
	AddAction(ctx context.Context, interventionID string, action *domain.InterventionAction) error

	// ExecuteAction executes a specific action within an intervention
	ExecuteAction(ctx context.Context, interventionID string, actionID string) error

	// AddCondition adds a condition to an intervention
	AddCondition(ctx context.Context, interventionID string, condition *domain.Condition) error

	// VerifyCondition verifies if a condition has been met
	VerifyCondition(ctx context.Context, interventionID string, conditionID string, verified bool, notes string) error

	// AddNote adds a note to an intervention
	AddNote(ctx context.Context, interventionID string, note string) error

	// AttachDocument attaches a document to an intervention
	AttachDocument(ctx context.Context, interventionID string, document *domain.Document) error

	// EscalateIntervention escalates an intervention to a higher authority
	EscalateIntervention(ctx context.Context, id string, reason string, escalateTo string) (*domain.Intervention, error)

	// CheckInterventionProgress checks the progress of an intervention
	CheckInterventionProgress(ctx context.Context, id string) (*domain.ProgressReport, error)

	// CalculateEffectiveness calculates the effectiveness score for an intervention
	CalculateEffectiveness(ctx context.Context, id string) (*domain.EffectivenessScore, error)

	// GenerateReport generates an intervention report
	GenerateReport(ctx context.Context, interventionID string, reportType string) ([]byte, error)

	// GetInterventionStats returns statistics about interventions
	GetInterventionStats(ctx context.Context, filter *domain.InterventionStatsFilter) (*domain.InterventionStats, error)

	// GetWorkflowState returns the current workflow state for an intervention
	GetWorkflowState(ctx context.Context, interventionID string) (*domain.WorkflowState, error)

	// TransitionWorkflow transitions the intervention to the next workflow state
	TransitionWorkflow(ctx context.Context, interventionID string, transition string) (*domain.Intervention, error)

	// GetAvailableTransitions returns the available workflow transitions for an intervention
	GetAvailableTransitions(ctx context.Context, interventionID string) ([]string, error)

	// CloneIntervention creates a copy of an existing intervention
	CloneIntervention(ctx context.Context, id string, newEntityType, newEntityID string) (*domain.Intervention, error)

	// ArchiveIntervention archives an intervention for record keeping
	ArchiveIntervention(ctx context.Context, id string) error

	// GetRecommendedInterventions returns recommended interventions based on entity state
	GetRecommendedInterventions(ctx context.Context, entityType string, entityID string) ([]*domain.InterventionRecommendation, error)

	// SimulateIntervention simulates the effects of an intervention
	SimulateIntervention(ctx context.Context, input *domain.SimulateInterventionInput) (*domain.SimulationResult, error)
}

// InterventionEventPublisher defines the interface for publishing intervention events.
type InterventionEventPublisher interface {
	// PublishInterventionCreated publishes an event when a new intervention is created
	PublishInterventionCreated(ctx context.Context, intervention *domain.Intervention) error

	// PublishInterventionUpdated publishes an event when an intervention is updated
	PublishInterventionUpdated(ctx context.Context, intervention *domain.Intervention) error

	// PublishInterventionStatusChanged publishes an event when intervention status changes
	PublishInterventionStatusChanged(ctx context.Context, intervention *domain.Intervention, oldStatus, newStatus domain.InterventionStatus) error

	// PublishInterventionAction publishes an event when an action is executed
	PublishInterventionAction(ctx context.Context, intervention *domain.Intervention, action *domain.InterventionAction) error

	// PublishInterventionConditionMet publishes an event when a condition is met
	PublishInterventionConditionMet(ctx context.Context, intervention *domain.Intervention, condition *domain.Condition) error

	// PublishInterventionCompleted publishes an event when an intervention is completed
	PublishInterventionCompleted(ctx context.Context, intervention *domain.Intervention, outcome *domain.Outcome) error

	// PublishInterventionEscalated publishes an event when an intervention is escalated
	PublishInterventionEscalated(ctx context.Context, intervention *domain.Intervention, reason string) error

	// PublishInterventionEffectiveness publishes effectiveness metrics
	PublishInterventionEffectiveness(ctx context.Context, interventionID string, score *domain.EffectivenessScore) error
}

// InterventionEventSubscriber defines the interface for consuming intervention events.
type InterventionEventSubscriber interface {
	// SubscribeToInterventionEvents starts consuming intervention-related events
	SubscribeToInterventionEvents(ctx context.Context) error

	// HandleInterventionEvent processes a single intervention event
	HandleInterventionEvent(ctx context.Context, event *domain.InterventionEvent) error

	// Close stops the event subscription
	Close() error
}

// InterventionCacheRepository defines the interface for caching intervention data.
type InterventionCacheRepository interface {
	// Get retrieves a cached intervention
	Get(ctx context.Context, id string) (*domain.Intervention, error)

	// Set stores an intervention in the cache
	Set(ctx context.Context, intervention *domain.Intervention, ttl time.Duration) error

	// Delete removes an intervention from the cache
	Delete(ctx context.Context, id string) error

	// Invalidate invalidates all cached interventions
	Invalidate(ctx context.Context) error

	// InvalidateByEntity invalidates all cached interventions for an entity
	InvalidateByEntity(ctx context.Context, entityType, entityID string) error

	// InvalidateByType invalidates all cached interventions of a type
	InvalidateByType(ctx context.Context, interventionType domain.InterventionType) error

	// InvalidateByStatus invalidates all cached interventions with a status
	InvalidateByStatus(ctx context.Context, status domain.InterventionStatus) error
}

// InterventionNotificationService defines the interface for intervention-related notifications.
type InterventionNotificationService interface {
	// SendNotification sends a notification about an intervention
	SendNotification(ctx context.Context, notification *domain.InterventionNotification) error

	// SendBulkNotifications sends multiple notifications
	SendBulkNotifications(ctx context.Context, notifications []*domain.InterventionNotification) error

	// SendDeadlineReminder sends a deadline reminder for an intervention
	SendDeadlineReminder(ctx context.Context, interventionID string, daysRemaining int) error

	// SendEscalationNotification sends an escalation notification
	SendEscalationNotification(ctx context.Context, interventionID string, escalateTo string, reason string) error

	// GetNotificationTemplates retrieves all available notification templates
	GetNotificationTemplates(ctx context.Context) ([]*domain.NotificationTemplate, error)
}

// InterventionSearchRepository defines the interface for searching intervention records.
type InterventionSearchRepository interface {
	// Index indexes an intervention record for search
	Index(ctx context.Context, intervention *domain.Intervention) error

	// DeleteIndex removes an intervention record from the search index
	DeleteIndex(ctx context.Context, id string) error

	// Search performs a full-text search on intervention records
	Search(ctx context.Context, query string, filter *domain.InterventionSearchFilter) ([]*domain.Intervention, error)

	// Suggest provides search suggestions based on partial queries
	Suggest(ctx context.Context, query string, limit int) ([]string, error)

	// Reindex rebuilds the search index for all intervention records
	Reindex(ctx context.Context) error

	// ReindexOne rebuilds the search index for a single intervention record
	ReindexOne(ctx context.Context, id string) error
}

// InterventionTemplateRepository defines the interface for intervention template management.
type InterventionTemplateRepository interface {
	// Create creates a new intervention template
	Create(ctx context.Context, template *domain.InterventionTemplate) error

	// GetByID retrieves a template by ID
	GetByID(ctx context.Context, id string) (*domain.InterventionTemplate, error)

	// GetByType retrieves templates by type
	GetByType(ctx context.Context, interventionType domain.InterventionType) ([]*domain.InterventionTemplate, error)

	// List lists all templates
	List(ctx context.Context, offset, limit int) ([]*domain.InterventionTemplate, error)

	// Update updates a template
	Update(ctx context.Context, template *domain.InterventionTemplate) error

	// Delete deletes a template
	Delete(ctx context.Context, id string) error

	// Clone clones a template
	Clone(ctx context.Context, id string, newName string) (*domain.InterventionTemplate, error)

	// UseTemplate creates an intervention from a template
	UseTemplate(ctx context.Context, templateID string, entityType, entityID string) (*domain.Intervention, error)
}
