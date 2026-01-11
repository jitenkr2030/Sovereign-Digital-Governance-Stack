package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
)

// ComplianceRepository defines the interface for compliance data persistence.
// This interface handles all database operations for compliance entities.
type ComplianceRepository interface {
	// Create adds a new compliance record to the repository
	Create(ctx context.Context, compliance *domain.ComplianceRecord) error

	// GetByID retrieves a compliance record by its unique identifier
	GetByID(ctx context.Context, id string) (*domain.ComplianceRecord, error)

	// GetByEntity retrieves all compliance records for a specific entity
	GetByEntity(ctx context.Context, entityType string, entityID string) ([]*domain.ComplianceRecord, error)

	// GetByPolicy retrieves all compliance records related to a specific policy
	GetByPolicy(ctx context.Context, policyID string) ([]*domain.ComplianceRecord, error)

	// GetByStatus retrieves all compliance records with the specified status
	GetByStatus(ctx context.Context, status domain.ComplianceStatus) ([]*domain.ComplianceRecord, error)

	// GetCurrentCompliance retrieves the current compliance status for an entity
	GetCurrentCompliance(ctx context.Context, entityType string, entityID string) (*domain.ComplianceSummary, error)

	// Update modifies an existing compliance record
	Update(ctx context.Context, compliance *domain.ComplianceRecord) error

	// Delete removes a compliance record
	Delete(ctx context.Context, id string) error

	// List retrieves all compliance records with optional pagination
	List(ctx context.Context, filter *domain.ComplianceFilter, offset, limit int) ([]*domain.ComplianceRecord, error)

	// Count returns the total count of compliance records matching the filter
	Count(ctx context.Context, filter *domain.ComplianceFilter) (int64, error)

	// GetComplianceHistory retrieves the compliance history for an entity
	GetComplianceHistory(ctx context.Context, entityType string, entityID string) ([]*domain.ComplianceHistoryEntry, error)

	// GetComplianceHistoryByTimeRange retrieves compliance history within a time range
	GetComplianceHistoryByTimeRange(ctx context.Context, entityType string, entityID string, startDate, endDate time.Time) ([]*domain.ComplianceHistoryEntry, error)

	// CreateComplianceSnapshot creates a snapshot of compliance status
	CreateComplianceSnapshot(ctx context.Context, summary *domain.ComplianceSummary) error

	// GetComplianceSnapshots retrieves compliance snapshots
	GetComplianceSnapshots(ctx context.Context, entityType string, entityID string) ([]*domain.ComplianceSnapshot, error)

	// GetViolations retrieves all violations for an entity
	GetViolations(ctx context.Context, entityType string, entityID string) ([]*domain.Violation, error)

	// CreateViolation creates a new violation record
	CreateViolation(ctx context.Context, violation *domain.Violation) error

	// UpdateViolation updates a violation record
	UpdateViolation(ctx context.Context, violation *domain.Violation) error

	// ResolveViolation marks a violation as resolved
	ResolveViolation(ctx context.Context, violationID string, resolution string) error

	// GetExemptions retrieves all exemptions for an entity
	GetExemptions(ctx context.Context, entityType string, entityID string) ([]*domain.Exemption, error)

	// CreateExemption creates a new exemption record
	CreateExemption(ctx context.Context, exemption *domain.Exemption) error

	// UpdateExemption updates an exemption record
	UpdateExemption(ctx context.Context, exemption *domain.Exemption) error

	// RevokeExemption revokes an existing exemption
	RevokeExemption(ctx context.Context, exemptionID string, reason string) error

	// GetAuditTrail retrieves the audit trail for compliance activities
	GetAuditTrail(ctx context.Context, entityType string, entityID string) ([]*domain.ComplianceAuditEntry, error)

	// CreateAuditEntry creates an audit trail entry
	CreateAuditEntry(ctx context.Context, entry *domain.ComplianceAuditEntry) error

	// GetMetrics retrieves compliance metrics
	GetMetrics(ctx context.Context, filter *domain.ComplianceMetricsFilter) (*domain.ComplianceMetrics, error)

	// Exists checks if a compliance record exists
	Exists(ctx context.Context, id string) (bool, error)

	// BulkCreate creates multiple compliance records
	BulkCreate(ctx context.Context, compliances []*domain.ComplianceRecord) error

	// BulkUpdateStatus updates the status of multiple compliance records
	BulkUpdateStatus(ctx context.Context, ids []string, status domain.ComplianceStatus) error
}

// ComplianceService defines the interface for compliance business logic operations.
// This interface integrates policy, enforcement, state, and intervention operations.
type ComplianceService interface {
	// CheckCompliance performs a comprehensive compliance check for an entity
	CheckCompliance(ctx context.Context, entityType string, entityID string) (*domain.ComplianceCheckResult, error)

	// CheckPolicyCompliance checks compliance against a specific policy
	CheckPolicyCompliance(ctx context.Context, entityType string, entityID string, policyID string) (*domain.PolicyComplianceResult, error)

	// GetComplianceStatus retrieves the current compliance status for an entity
	GetComplianceStatus(ctx context.Context, entityType string, entityID string) (*domain.ComplianceSummary, error)

	// ListComplianceRecords retrieves compliance records with filtering
	ListComplianceRecords(ctx context.Context, filter *domain.ComplianceFilter) ([]*domain.ComplianceRecord, error)

	// CreateComplianceRecord creates a new compliance record
	CreateComplianceRecord(ctx context.Context, input *domain.CreateComplianceInput) (*domain.ComplianceRecord, error)

	// UpdateComplianceRecord updates an existing compliance record
	UpdateComplianceRecord(ctx context.Context, id string, input *domain.UpdateComplianceInput) (*domain.ComplianceRecord, error)

	// GetComplianceHistory retrieves the compliance history for an entity
	GetComplianceHistory(ctx context.Context, entityType string, entityID string, startDate, endDate time.Time) ([]*domain.ComplianceHistoryEntry, error)

	// GetComplianceTrends retrieves compliance trend data
	GetComplianceTrends(ctx context.Context, entityType string, entityID string, timeRange domain.TimeRange) (*domain.ComplianceTrends, error)

	// GetComplianceMetrics retrieves compliance metrics
	GetComplianceMetrics(ctx context.Context, filter *domain.ComplianceMetricsFilter) (*domain.ComplianceMetrics, error)

	// GetComplianceScore calculates and returns the compliance score
	GetComplianceScore(ctx context.Context, entityType string, entityID string) (*domain.ComplianceScore, error)

	// ReportViolation reports a violation for an entity
	ReportViolation(ctx context.Context, input *domain.ReportViolationInput) (*domain.Violation, error)

	// UpdateViolation updates a violation
	UpdateViolation(ctx context.Context, id string, input *domain.UpdateViolationInput) (*domain.Violation, error)

	// ResolveViolation resolves a violation
	ResolveViolation(ctx context.Context, id string, resolution string) (*domain.Violation, error)

	// WaiveViolation waives a violation
	WaiveViolation(ctx context.Context, id string, reason string) (*domain.Violation, error)

	// RequestExemption requests an exemption for a policy requirement
	RequestExemption(ctx context.Context, input *domain.RequestExemptionInput) (*domain.Exemption, error)

	// ApproveExemption approves an exemption request
	ApproveExemption(ctx context.Context, id string, notes string) (*domain.Exemption, error)

	// RejectExemption rejects an exemption request
	RejectExemption(ctx context.Context, id string, reason string) (*domain.Exemption, error)

	// RevokeExemption revokes an existing exemption
	RevokeExemption(ctx context.Context, id string, reason string) (*domain.Exemption, error)

	// CreateSnapshot creates a compliance snapshot
	CreateSnapshot(ctx context.Context, entityType string, entityID string) (*domain.ComplianceSnapshot, error)

	// CompareSnapshots compares two compliance snapshots
	CompareSnapshots(ctx context.Context, snapshotID1, snapshotID2 string) (*domain.SnapshotComparison, error)

	// GenerateComplianceReport generates a compliance report
	GenerateComplianceReport(ctx context.Context, entityType string, entityID string, reportType string) ([]byte, error)

	// ScheduleComplianceCheck schedules an automated compliance check
	ScheduleComplianceCheck(ctx context.Context, entityType string, entityID string, schedule *domain.CheckSchedule) error

	// GetScheduledChecks retrieves all scheduled compliance checks
	GetScheduledChecks(ctx context.Context) ([]*domain.CheckSchedule, error)

	// CancelScheduledCheck cancels a scheduled check
	CancelScheduledCheck(ctx context.Context, scheduleID string) error

	// TriggerImmediateCheck triggers an immediate compliance check
	TriggerImmediateCheck(ctx context.Context, entityType string, entityID string) error

	// GetComplianceAlerts retrieves all active compliance alerts
	GetComplianceAlerts(ctx context.Context) ([]*domain.ComplianceAlert, error)

	// AcknowledgeAlert acknowledges a compliance alert
	AcknowledgeAlert(ctx context.Context, alertID string, userID string) error

	// DismissAlert dismisses a compliance alert
	DismissAlert(ctx context.Context, alertID string, reason string) error

	// GetApplicablePolicies retrieves all policies applicable to an entity
	GetApplicablePolicies(ctx context.Context, entityType string, entityID string) ([]*domain.Policy, error)

	// AssessPolicyApplicability assesses if a policy applies to an entity
	AssessPolicyApplicability(ctx context.Context, entityType string, entityID string, policyID string) (*domain.ApplicabilityResult, error)

	// GetComplianceRequirements retrieves all compliance requirements for an entity
	GetComplianceRequirements(ctx context.Context, entityType string, entityID string) ([]*domain.ComplianceRequirement, error)

	// SelfAssessCompliance performs a self-assessment compliance check
	SelfAssessCompliance(ctx context.Context, entityType string, entityID string, assessment *domain.SelfAssessment) (*domain.SelfAssessmentResult, error)

	// SubmitSelfAssessment submits a self-assessment for review
	SubmitSelfAssessment(ctx context.Context, entityType string, entityID string) error

	// ReviewSelfAssessment reviews and approves/rejects a self-assessment
	ReviewSelfAssessment(ctx context.Context, assessmentID string, decision string, notes string) error

	// ExportComplianceData exports compliance data for external review
	ExportComplianceData(ctx context.Context, entityType string, entityID string, format string) ([]byte, error)

	// ImportComplianceData imports compliance data from an external source
	ImportComplianceData(ctx context.Context, source string, metadata *domain.ImportMetadata) error
}

// ComplianceEventPublisher defines the interface for publishing compliance events.
type ComplianceEventPublisher interface {
	// PublishComplianceCheckCompleted publishes an event when a compliance check is completed
	PublishComplianceCheckCompleted(ctx context.Context, result *domain.ComplianceCheckResult) error

	// PublishComplianceStatusChanged publishes an event when compliance status changes
	PublishComplianceStatusChanged(ctx context.Context, entityType, entityID string, oldStatus, newStatus domain.ComplianceStatus) error

	// PublishViolationDetected publishes an event when a violation is detected
	PublishViolationDetected(ctx context.Context, violation *domain.Violation) error

	// PublishViolationResolved publishes an event when a violation is resolved
	PublishViolationResolved(ctx context.Context, violation *domain.Violation) error

	// PublishExemptionRequested publishes an event when an exemption is requested
	PublishExemptionRequested(ctx context.Context, exemption *domain.Exemption) error

	// PublishExemptionApproved publishes an event when an exemption is approved
	PublishExemptionApproved(ctx context.Context, exemption *domain.Exemption) error

	// PublishExemptionRejected publishes an event when an exemption is rejected
	PublishExemptionRejected(ctx context.Context, exemption *domain.Exemption) error

	// PublishComplianceAlert publishes an event when a compliance alert is generated
	PublishComplianceAlert(ctx context.Context, alert *domain.ComplianceAlert) error

	// PublishSelfAssessmentSubmitted publishes an event when a self-assessment is submitted
	PublishSelfAssessmentSubmitted(ctx context.Context, entityType, entityID string) error

	// PublishSnapshotCreated publishes an event when a compliance snapshot is created
	PublishSnapshotCreated(ctx context.Context, snapshot *domain.ComplianceSnapshot) error

	// PublishComplianceReportGenerated publishes an event when a compliance report is generated
	PublishComplianceReportGenerated(ctx context.Context, entityType, entityID string, reportType string) error
}

// ComplianceEventSubscriber defines the interface for consuming compliance events.
type ComplianceEventSubscriber interface {
	// SubscribeToComplianceEvents starts consuming compliance-related events
	SubscribeToComplianceEvents(ctx context.Context) error

	// HandleComplianceEvent processes a single compliance event
	HandleComplianceEvent(ctx context.Context, event *domain.ComplianceEvent) error

	// Close stops the event subscription
	Close() error
}

// ComplianceCacheRepository defines the interface for caching compliance data.
type ComplianceCacheRepository interface {
	// Get retrieves a cached compliance record
	Get(ctx context.Context, id string) (*domain.ComplianceRecord, error)

	// Set stores a compliance record in the cache
	Set(ctx context.Context, compliance *domain.ComplianceRecord, ttl time.Duration) error

	// Delete removes a compliance record from the cache
	Delete(ctx context.Context, id string) error

	// Invalidate invalidates all cached compliance records
	Invalidate(ctx context.Context) error

	// InvalidateByEntity invalidates all cached compliance records for an entity
	InvalidateByEntity(ctx context.Context, entityType, entityID string) error

	// InvalidateByPolicy invalidates all cached compliance records for a policy
	InvalidateByPolicy(ctx context.Context, policyID string) error

	// GetCurrentCompliance retrieves the current compliance summary for an entity
	GetCurrentCompliance(ctx context.Context, entityType, entityID string) (*domain.ComplianceSummary, error)

	// SetCurrentCompliance stores the current compliance summary for an entity
	SetCurrentCompliance(ctx context.Context, summary *domain.ComplianceSummary, ttl time.Duration) error

	// InvalidateCurrentCompliance invalidates the current compliance summary
	InvalidateCurrentCompliance(ctx context.Context, entityType, entityID string) error
}

// ComplianceNotificationService defines the interface for compliance-related notifications.
type ComplianceNotificationService interface {
	// NotifyComplianceStatusChange sends a notification about compliance status change
	NotifyComplianceStatusChange(ctx context.Context, entityType, entityID string, oldStatus, newStatus domain.ComplianceStatus) error

	// NotifyViolationDetected sends a notification about a detected violation
	NotifyViolationDetected(ctx context.Context, violation *domain.Violation) error

	// NotifyViolationResolved sends a notification about a resolved violation
	NotifyViolationResolved(ctx context.Context, violation *domain.Violation) error

	// NotifyExemptionRequest sends a notification about an exemption request
	NotifyExemptionRequest(ctx context.Context, exemption *domain.Exemption) error

	// NotifyExemptionDecision sends a notification about an exemption decision
	NotifyExemptionDecision(ctx context.Context, exemption *domain.Exemption) error

	// NotifyComplianceAlert sends a compliance alert notification
	NotifyComplianceAlert(ctx context.Context, alert *domain.ComplianceAlert) error

	// NotifyDeadlineApproaching sends a notification about an approaching deadline
	NotifyDeadlineApproaching(ctx context.Context, entityType, entityID string, deadline time.Time, daysRemaining int) error

	// NotifySelfAssessmentDue sends a notification about an upcoming self-assessment
	NotifySelfAssessmentDue(ctx context.Context, entityType, entityID string, dueDate time.Time) error

	// GetNotificationPreferences retrieves notification preferences for a user
	GetNotificationPreferences(ctx context.Context, userID string) (*domain.NotificationPreferences, error)

	// UpdateNotificationPreferences updates notification preferences
	UpdateNotificationPreferences(ctx context.Context, userID string, preferences *domain.NotificationPreferences) error
}

// ComplianceSearchRepository defines the interface for searching compliance records.
type ComplianceSearchRepository interface {
	// Index indexes a compliance record for search
	Index(ctx context.Context, compliance *domain.ComplianceRecord) error

	// DeleteIndex removes a compliance record from the search index
	DeleteIndex(ctx context.Context, id string) error

	// Search performs a full-text search on compliance records
	Search(ctx context.Context, query string, filter *domain.ComplianceSearchFilter) ([]*domain.ComplianceRecord, error)

	// Suggest provides search suggestions based on partial queries
	Suggest(ctx context.Context, query string, limit int) ([]string, error)

	// Reindex rebuilds the search index for all compliance records
	Reindex(ctx context.Context) error

	// ReindexOne rebuilds the search index for a single compliance record
	ReindexOne(ctx context.Context, id string) error

	// SearchByEntity performs a search for an entity's compliance records
	SearchByEntity(ctx context.Context, entityType, entityID string) ([]*domain.ComplianceRecord, error)

	// SearchByViolation performs a search for compliance records with violations
	SearchByViolation(ctx context.Context, query string, offset, limit int) ([]*domain.ComplianceRecord, error)
}

// ComplianceExternalService defines the interface for external compliance services.
type ComplianceExternalService interface {
	// VerifyExternalCompliance verifies compliance with external regulatory bodies
	VerifyExternalCompliance(ctx context.Context, entityType, entityID string, externalBody string) (*domain.ExternalComplianceResult, error)

	// SubmitComplianceReport submits a compliance report to an external body
	SubmitComplianceReport(ctx context.Context, entityType, entityID string, externalBody string, report []byte) error

	// FetchExternalGuidelines fetches compliance guidelines from an external body
	FetchExternalGuidelines(ctx context.Context, externalBody string) ([]*domain.ExternalGuideline, error)

	// SyncComplianceStandards syncs compliance standards from an external source
	SyncComplianceStandards(ctx context.Context, source string) error

	// GetExternalCertifications retrieves external certifications for an entity
	GetExternalCertifications(ctx context.Context, entityType, entityID string) ([]*domain.ExternalCertification, error)

	// RequestExternalAudit requests an external audit for an entity
	RequestExternalAudit(ctx context.Context, entityType, entityID string, externalBody string, auditType string) error

	// ReceiveExternalAuditResult receives the result of an external audit
	ReceiveExternalAuditResult(ctx context.Context, auditID string, result *domain.ExternalAuditResult) error
}
