package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
)

// EnforcementRepository defines the interface for enforcement action data persistence.
// This interface handles all database operations for enforcement entities.
type EnforcementRepository interface {
	// Create adds a new enforcement action to the repository
	Create(ctx context.Context, enforcement *domain.Enforcement) error

	// GetByID retrieves an enforcement action by its unique identifier
	GetByID(ctx context.Context, id string) (*domain.Enforcement, error)

	// GetByEntity retrieves all enforcement actions for a specific entity
	GetByEntity(ctx context.Context, entityType string, entityID string) ([]*domain.Enforcement, error)

	// GetByPolicy retrieves all enforcement actions related to a specific policy
	GetByPolicy(ctx context.Context, policyID string) ([]*domain.Enforcement, error)

	// GetByStatus retrieves all enforcement actions with the specified status
	GetByStatus(ctx context.Context, status domain.EnforcementStatus) ([]*domain.Enforcement, error)

	// GetByAssignee retrieves all enforcement actions assigned to a specific user
	GetByAssignee(ctx context.Context, assigneeID string) ([]*domain.Enforcement, error)

	// Update modifies an existing enforcement action
	Update(ctx context.Context, enforcement *domain.Enforcement) error

	// Delete removes an enforcement action
	Delete(ctx context.Context, id string) error

	// List retrieves all enforcement actions with optional pagination
	List(ctx context.Context, filter *domain.EnforcementFilter, offset, limit int) ([]*domain.Enforcement, error)

	// Count returns the total count of enforcement actions matching the filter
	Count(ctx context.Context, filter *domain.EnforcementFilter) (int64, error)

	// GetTimeline retrieves the complete timeline for an enforcement action
	GetTimeline(ctx context.Context, enforcementID string) ([]*domain.TimelineEvent, error)

	// AddTimelineEvent adds a new event to the enforcement timeline
	AddTimelineEvent(ctx context.Context, enforcementID string, event *domain.TimelineEvent) error

	// GetDocuments retrieves all documents attached to an enforcement action
	GetDocuments(ctx context.Context, enforcementID string) ([]*domain.Document, error)

	// AttachDocument attaches a document to an enforcement action
	AttachDocument(ctx context.Context, enforcementID string, document *domain.Document) error

	// DetachDocument removes a document from an enforcement action
	DetachDocument(ctx context.Context, enforcementID string, documentID string) error

	// GetNotes retrieves all notes for an enforcement action
	GetNotes(ctx context.Context, enforcementID string) ([]*domain.Note, error)

	// AddNote adds a note to an enforcement action
	AddNote(ctx context.Context, enforcementID string, note *domain.Note) error

	// GetHistory retrieves the complete history of changes for an enforcement action
	GetHistory(ctx context.Context, enforcementID string) ([]*domain.HistoryEntry, error)

	// Exists checks if an enforcement action exists
	Exists(ctx context.Context, id string) (bool, error)

	// Search performs a text search on enforcement records
	Search(ctx context.Context, query string, offset, limit int) ([]*domain.Enforcement, error)

	// BulkUpdateStatus updates the status of multiple enforcement actions
	BulkUpdateStatus(ctx context.Context, ids []string, status domain.EnforcementStatus) error

	// BulkAssign updates the assignee for multiple enforcement actions
	BulkAssign(ctx context.Context, ids []string, assigneeID string) error
}

// EnforcementService defines the interface for enforcement business logic operations.
// This interface encapsulates all domain-specific operations for enforcement actions.
type EnforcementService interface {
	// CreateEnforcement initiates a new enforcement action
	CreateEnforcement(ctx context.Context, input *domain.CreateEnforcementInput) (*domain.Enforcement, error)

	// GetEnforcement retrieves an enforcement action by ID
	GetEnforcement(ctx context.Context, id string) (*domain.Enforcement, error)

	// ListEnforcements retrieves enforcement actions with filtering and pagination
	ListEnforcements(ctx context.Context, filter *domain.EnforcementFilter) ([]*domain.Enforcement, error)

	// UpdateEnforcement updates an existing enforcement action
	UpdateEnforcement(ctx context.Context, id string, input *domain.UpdateEnforcementInput) (*domain.Enforcement, error)

	// CloseEnforcement closes an enforcement action with resolution
	CloseEnforcement(ctx context.Context, id string, resolution *domain.Resolution) (*domain.Enforcement, error)

	// EscalateEnforcement escalates an enforcement action to a higher authority
	EscalateEnforcement(ctx context.Context, id string, reason string, escalateTo string) (*domain.Enforcement, error)

	// AssignEnforcement assigns an enforcement action to an investigator
	AssignEnforcement(ctx context.Context, id string, assigneeID string) (*domain.Enforcement, error)

	// ReassignEnforcement reassigns an enforcement action to a different investigator
	ReassignEnforcement(ctx context.Context, id string, newAssigneeID string, reason string) (*domain.Enforcement, error)

	// AddEvidence adds evidence to an enforcement action
	AddEvidence(ctx context.Context, enforcementID string, evidence *domain.Evidence) error

	// ReviewEvidence reviews evidence submitted for an enforcement action
	ReviewEvidence(ctx context.Context, enforcementID string, evidenceID string, decision string, notes string) error

	// IssueNotice issues an official notice related to an enforcement action
	IssueNotice(ctx context.Context, enforcementID string, noticeType domain.NoticeType, templateID string) (*domain.Notice, error)

	// ScheduleHearing schedules a hearing for an enforcement action
	ScheduleHearing(ctx context.Context, enforcementID string, hearing *domain.Hearing) error

	// RescheduleHearing reschedules an existing hearing
	RescheduleHearing(ctx context.Context, enforcementID string, hearingID string, newDate time.Time, reason string) error

	// IssueRuling issues a ruling for an enforcement action
	IssueRuling(ctx context.Context, enforcementID string, ruling *domain.Ruling) (*domain.Enforcement, error)

	// AppealRuling initiates an appeal against a ruling
	AppealRuling(ctx context.Context, enforcementID string, appeal *domain.Appeal) error

	// CheckCompliance checks entity compliance for enforcement purposes
	CheckCompliance(ctx context.Context, enforcementID string) (*domain.ComplianceCheckResult, error)

	// GenerateReport generates an enforcement report
	GenerateReport(ctx context.Context, enforcementID string, reportType string) ([]byte, error)

	// GetEnforcementStats returns statistics about enforcement actions
	GetEnforcementStats(ctx context.Context, filter *domain.EnforcementStatsFilter) (*domain.EnforcementStats, error)

	// GetWorkflowState returns the current workflow state for an enforcement action
	GetWorkflowState(ctx context.Context, enforcementID string) (*domain.WorkflowState, error)

	// TransitionWorkflow transitions the enforcement action to the next workflow state
	TransitionWorkflow(ctx context.Context, enforcementID string, transition string) (*domain.Enforcement, error)

	// GetAvailableTransitions returns the available workflow transitions for an enforcement action
	GetAvailableTransitions(ctx context.Context, enforcementID string) ([]string, error)

	// NotifyStakeholders sends notifications to all relevant stakeholders
	NotifyStakeholders(ctx context.Context, enforcementID string, notificationType string) error

	// ArchiveEnforcement archives an enforcement action for record keeping
	ArchiveEnforcement(ctx context.Context, id string) error
}

// EnforcementEventPublisher defines the interface for publishing enforcement events.
type EnforcementEventPublisher interface {
	// PublishEnforcementCreated publishes an event when a new enforcement is created
	PublishEnforcementCreated(ctx context.Context, enforcement *domain.Enforcement) error

	// PublishEnforcementUpdated publishes an event when an enforcement is updated
	PublishEnforcementUpdated(ctx context.Context, enforcement *domain.Enforcement) error

	// PublishEnforcementStatusChanged publishes an event when enforcement status changes
	PublishEnforcementStatusChanged(ctx context.Context, enforcement *domain.Enforcement, oldStatus, newStatus domain.EnforcementStatus) error

	// PublishEnforcementEscalated publishes an event when an enforcement is escalated
	PublishEnforcementEscalated(ctx context.Context, enforcement *domain.Enforcement, reason string) error

	// PublishHearingScheduled publishes an event when a hearing is scheduled
	PublishHearingScheduled(ctx context.Context, enforcement *domain.Enforcement, hearing *domain.Hearing) error

	// PublishRulingIssued publishes an event when a ruling is issued
	PublishRulingIssued(ctx context.Context, enforcement *domain.Enforcement, ruling *domain.Ruling) error

	// PublishComplianceAlert publishes an event when a compliance issue is detected
	PublishComplianceAlert(ctx context.Context, result *domain.ComplianceCheckResult) error
}

// EnforcementEventSubscriber defines the interface for consuming enforcement events.
type EnforcementEventSubscriber interface {
	// SubscribeToEnforcementEvents starts consuming enforcement-related events
	SubscribeToEnforcementEvents(ctx context.Context) error

	// HandleEnforcementEvent processes a single enforcement event
	HandleEnforcementEvent(ctx context.Context, event *domain.EnforcementEvent) error

	// Close stops the event subscription
	Close() error
}

// EnforcementNotificationService defines the interface for sending notifications
// related to enforcement actions.
type EnforcementNotificationService interface {
	// SendNotification sends a notification about an enforcement action
	SendNotification(ctx context.Context, notification *domain.EnforcementNotification) error

	// SendBulkNotifications sends multiple notifications
	SendBulkNotifications(ctx context.Context, notifications []*domain.EnforcementNotification) error

	// GetNotificationTemplates retrieves all available notification templates
	GetNotificationTemplates(ctx context.Context) ([]*domain.NotificationTemplate, error)

	// CreateNotificationTemplate creates a new notification template
	CreateNotificationTemplate(ctx context.Context, template *domain.NotificationTemplate) error

	// UpdateNotificationTemplate updates an existing notification template
	UpdateNotificationTemplate(ctx context.Context, id string, template *domain.NotificationTemplate) error

	// DeleteNotificationTemplate deletes a notification template
	DeleteNotificationTemplate(ctx context.Context, id string) error
}

// EnforcementSearchRepository defines the interface for searching enforcement records.
type EnforcementSearchRepository interface {
	// Index indexes an enforcement record for search
	Index(ctx context.Context, enforcement *domain.Enforcement) error

	// DeleteIndex removes an enforcement record from the search index
	DeleteIndex(ctx context.Context, id string) error

	// Search performs a full-text search on enforcement records
	Search(ctx context.Context, query string, filter *domain.EnforcementSearchFilter) ([]*domain.Enforcement, error)

	// Suggest provides search suggestions based on partial queries
	Suggest(ctx context.Context, query string, limit int) ([]string, error)

	// Reindex rebuilds the search index for all enforcement records
	Reindex(ctx context.Context) error

	// ReindexOne rebuilds the search index for a single enforcement record
	ReindexOne(ctx context.Context, id string) error
}
