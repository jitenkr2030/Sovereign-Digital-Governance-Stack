package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
)

// PolicyBusinessService defines the interface for policy business logic operations.
// This interface encapsulates all domain-specific operations that can be performed
// on policy entities beyond basic CRUD operations.
type PolicyBusinessService interface {
	// CompilePolicy compiles a policy into an executable format
	CompilePolicy(ctx context.Context, policyID string) (*domain.CompiledPolicy, error)

	// EvaluatePolicy evaluates a compiled policy against given context
	EvaluatePolicy(ctx context.Context, compiledPolicy *domain.CompiledPolicy, context map[string]interface{}) (*domain.PolicyEvaluationResult, error)

	// GetPolicyDependencies analyzes and returns policy dependencies
	GetPolicyDependencies(ctx context.Context, policyID string) ([]string, error)

	// ValidatePolicyStructure validates the structural integrity of a policy
	ValidatePolicyStructure(ctx context.Context, policy *domain.Policy) ([]*domain.ValidationError, error)

	// GetPolicyImpact assesses the impact of policy changes
	GetPolicyImpact(ctx context.Context, policyID string) (*domain.PolicyImpact, error)

	// GetPolicyCoverage analyzes policy coverage across entities
	GetPolicyCoverage(ctx context.Context, policyID string) (*domain.PolicyCoverage, error)

	// GetPolicyOverlap identifies overlapping policies
	GetPolicyOverlap(ctx context.Context, filter *domain.PolicyOverlapFilter) ([]*domain.PolicyOverlap, error)

	// SuggestPolicy creates policy suggestions based on patterns
	SuggestPolicy(ctx context.Context, context *domain.PolicySuggestionContext) ([]*domain.PolicySuggestion, error)

	// GetPolicyEffectiveness measures policy effectiveness
	GetPolicyEffectiveness(ctx context.Context, policyID string, timeRange domain.TimeRange) (*domain.PolicyEffectiveness, error)

	// GetPolicyCompliance calculates compliance rates for a policy
	GetPolicyCompliance(ctx context.Context, policyID string) (*domain.PolicyCompliance, error)

	// GetPolicyViolations retrieves violations of a policy
	GetPolicyViolations(ctx context.Context, policyID string, filter *domain.ViolationFilter) ([]*domain.Violation, error)

	// GetPolicyHistory retrieves the complete history of policy changes
	GetPolicyHistory(ctx context.Context, policyID string) ([]*domain.PolicyHistoryEntry, error)

	// ComparePolicies compares two policies and identifies differences
	ComparePolicies(ctx context.Context, policyID1, policyID2 string) (*domain.PolicyComparison, error)

	// MergePolicies merges multiple policies into a single policy
	MergePolicies(ctx context.Context, policyIDs []string, strategy string) (*domain.Policy, error)

	// ArchivePolicy archives a policy for historical reference
	ArchivePolicy(ctx context.Context, policyID string) error

	// RestorePolicy restores a policy from archive
	RestorePolicy(ctx context.Context, archiveID string) (*domain.Policy, error)
}

// EnforcementBusinessService defines the interface for enforcement business logic operations.
type EnforcementBusinessService interface {
	// AssessCase assesses an enforcement case and recommends actions
	AssessCase(ctx context.Context, caseID string) (*domain.CaseAssessment, error)

	// PrioritizeCases prioritizes enforcement cases based on various factors
	PrioritizeCases(ctx context.Context, cases []*domain.CasePriority) ([]*domain.CasePriority, error)

	// CalculatePenalty calculates an appropriate penalty for a violation
	CalculatePenalty(ctx context.Context, violation *domain.Violation) (*domain.PenaltyCalculation, error)

	// NegotiateSettlement negotiates a settlement for an enforcement case
	NegotiateSettlement(ctx context.Context, caseID string, offer *domain.SettlementOffer) (*domain.Settlement, error)

	// ConductInvestigation conducts an investigation for an enforcement case
	ConductInvestigation(ctx context.Context, caseID string, investigation *domain.Investigation) error

	// ScheduleHearing schedules a hearing for an enforcement case
	ScheduleHearing(ctx context.Context, caseID string, hearing *domain.Hearing) (*domain.Hearing, error)

	// IssueRuling issues a ruling for an enforcement case
	IssueRuling(ctx context.Context, caseID string, ruling *domain.Ruling) (*domain.Ruling, error)

	// CollectEvidence collects and processes evidence
	CollectEvidence(ctx context.Context, caseID string, evidence *domain.Evidence) error

	// AnalyzeEvidence analyzes evidence for an enforcement case
	AnalyzeEvidence(ctx context.Context, caseID string, evidenceID string) (*domain.EvidenceAnalysis, error)

	// GetCaseTimeline retrieves the timeline for a case
	GetCaseTimeline(ctx context.Context, caseID string) ([]*domain.TimelineEvent, error)

	// AddTimelineEvent adds an event to the case timeline
	AddTimelineEvent(ctx context.Context, caseID string, event *domain.TimelineEvent) error

	// GetCaseMetrics retrieves metrics for an enforcement case
	GetCaseMetrics(ctx context.Context, caseID string) (*domain.CaseMetrics, error)

	// PredictOutcome predicts the likely outcome of a case
	PredictOutcome(ctx context.Context, caseID string) (*domain.OutcomePrediction, error)

	// GetEnforcementTrends retrieves enforcement trend data
	GetEnforcementTrends(ctx context.Context, filter *domain.EnforcementTrendFilter) (*domain.EnforcementTrends, error)

	// GetEnforcementStats retrieves aggregated enforcement statistics
	GetEnforcementStats(ctx context.Context, filter *domain.EnforcementStatsFilter) (*domain.EnforcementStats, error)

	// BulkAction performs bulk actions on multiple cases
	BulkAction(ctx context.Context, action *domain.BulkEnforcementAction) error
}

// StateBusinessService defines the interface for state business logic operations.
type StateBusinessService interface {
	// AssessState assesses the current state of an entity
	AssessState(ctx context.Context, entityType, entityID string) (*domain.StateAssessment, error)

	// PredictStateTransition predicts likely future state transitions
	PredictStateTransition(ctx context.Context, entityType, entityID string) ([]*domain.StateTransitionPrediction, error)

	// OptimizeStateTransitions optimizes state transition sequences
	OptimizeStateTransitions(ctx context.Context, transitions []*domain.StateTransition) ([]*domain.StateTransition, error)

	// DetectStateAnomaly detects anomalies in state transitions
	DetectStateAnomaly(ctx context.Context, entityType, entityID string) (*domain.StateAnomaly, error)

	// GetStateMetrics retrieves state-related metrics
	GetStateMetrics(ctx context.Context, entityType, entityID string) (*domain.StateMetrics, error)

	// GetStateHistory retrieves detailed state history
	GetStateHistory(ctx context.Context, entityType, entityID string, detailLevel string) ([]*domain.StateChange, error)

	// CompareStateHistories compares state histories of multiple entities
	CompareStateHistories(ctx context.Context, entities []*domain.EntityRef) (*domain.StateComparison, error)

	// GetStatePatterns identifies patterns in state transitions
	GetStatePatterns(ctx context.Context, entityType string, timeRange domain.TimeRange) ([]*domain.StatePattern, error)

	// SuggestStateTransition suggests optimal state transitions
	SuggestStateTransition(ctx context.Context, entityType, entityID string, targetState domain.StateType) (*domain.StateTransitionSuggestion, error)

	// ValidateStateTransition validates if a transition is allowed
	ValidateStateTransition(ctx context.Context, entityType, entityID string, targetState domain.StateType) (*domain.ValidationResult, error)

	// BatchStateTransition performs batch state transitions
	BatchStateTransition(ctx context.Context, transitions []*domain.BatchStateTransition) ([]*domain.StateTransitionResult, error)

	// ScheduleStateTransition schedules a future state transition
	ScheduleStateTransition(ctx context.Context, entityType, entityID string, targetState domain.StateType, scheduledTime time.Time, reason string) error

	// CancelScheduledTransition cancels a scheduled transition
	CancelScheduledTransition(ctx context.Context, transitionID string) error

	// GetScheduledTransitions retrieves scheduled transitions
	GetScheduledTransitions(ctx context.Context, filter *domain.ScheduledTransitionFilter) ([]*domain.ScheduledTransition, error)

	// GetStateRules retrieves rules that govern state transitions
	GetStateRules(ctx context.Context, entityType string) ([]*domain.StateRule, error)

	// UpdateStateRules updates state transition rules
	UpdateStateRules(ctx context.Context, entityType string, rules []*domain.StateRule) error
}

// InterventionBusinessService defines the interface for intervention business logic operations.
type InterventionBusinessService interface {
	// DesignIntervention designs an intervention plan
	DesignIntervention(ctx context.Context, entityType, entityID string, interventionType domain.InterventionType) (*domain.InterventionPlan, error)

	// SimulateIntervention simulates the effects of an intervention
	SimulateIntervention(ctx context.Context, plan *domain.InterventionPlan, context map[string]interface{}) (*domain.InterventionSimulation, error)

	// OptimizeIntervention optimizes an intervention plan
	OptimizeIntervention(ctx context.Context, plan *domain.InterventionPlan) (*domain.InterventionPlan, error)

	// AssessInterventionNeed assesses if an intervention is needed
	AssessInterventionNeed(ctx context.Context, entityType, entityID string) (*domain.InterventionNeedAssessment, error)

	// RecommendInterventions recommends appropriate interventions
	RecommendInterventions(ctx context.Context, entityType, entityID string) ([]*domain.InterventionRecommendation, error)

	// GetInterventionTemplates retrieves available intervention templates
	GetInterventionTemplates(ctx context.Context, filter *domain.InterventionTemplateFilter) ([]*domain.InterventionTemplate, error)

	// CreateInterventionTemplate creates a new intervention template
	CreateInterventionTemplate(ctx context.Context, template *domain.InterventionTemplate) error

	// UpdateInterventionTemplate updates an intervention template
	UpdateInterventionTemplate(ctx context.Context, templateID string, template *domain.InterventionTemplate) error

	// ExecuteIntervention executes an intervention plan
	ExecuteIntervention(ctx context.Context, planID string) (*domain.InterventionExecution, error)

	// MonitorIntervention monitors intervention progress
	MonitorIntervention(ctx context.Context, interventionID string) (*domain.InterventionProgress, error)

	// AdjustIntervention adjusts an ongoing intervention
	AdjustIntervention(ctx context.Context, interventionID string, adjustment *domain.InterventionAdjustment) error

	// CompleteIntervention completes an intervention
	CompleteIntervention(ctx context.Context, interventionID string, outcome *domain.Outcome) error

	// TerminateIntervention terminates an intervention
	TerminateIntervention(ctx context.Context, interventionID string, reason string) error

	// GetInterventionEffectiveness measures intervention effectiveness
	GetInterventionEffectiveness(ctx context.Context, interventionID string) (*domain.InterventionEffectiveness, error)

	// GetInterventionROI calculates return on investment for an intervention
	GetInterventionROI(ctx context.Context, interventionID string) (*domain.InterventionROI, error)

	// CompareInterventions compares multiple interventions
	CompareInterventions(ctx context.Context, interventionIDs []string) (*domain.InterventionComparison, error)

	// GetInterventionHistory retrieves intervention history for an entity
	GetInterventionHistory(ctx context.Context, entityType, entityID string) ([]*domain.Intervention, error)

	// GetInterventionPatterns identifies patterns in interventions
	GetInterventionPatterns(ctx context.Context, entityType string, timeRange domain.TimeRange) ([]*domain.InterventionPattern, error)
}

// ComplianceBusinessService defines the interface for compliance business logic operations.
type ComplianceBusinessService interface {
	// AssessCompliance assesses compliance status for an entity
	AssessCompliance(ctx context.Context, entityType, entityID string, scope []string) (*domain.ComplianceAssessment, error)

	// PerformAudit performs a compliance audit
	PerformAudit(ctx context.Context, audit *domain.ComplianceAudit) (*domain.AuditResult, error)

	// ScheduleAudit schedules a compliance audit
	ScheduleAudit(ctx context.Context, audit *domain.ScheduledAudit) error

	// GenerateComplianceReport generates a compliance report
	GenerateComplianceReport(ctx context.Context, entityType, entityID string, reportType string, scope []string) ([]byte, error)

	// GetComplianceScore calculates compliance score
	GetComplianceScore(ctx context.Context, entityType, entityID string) (*domain.ComplianceScore, error)

	// GetComplianceTrend calculates compliance trends
	GetComplianceTrend(ctx context.Context, entityType, entityID string, timeRange domain.TimeRange) (*domain.ComplianceTrend, error)

	// GetComplianceMetrics retrieves comprehensive compliance metrics
	GetComplianceMetrics(ctx context.Context, filter *domain.ComplianceMetricsFilter) (*domain.ComplianceMetrics, error)

	// GetComplianceGaps identifies compliance gaps
	GetComplianceGaps(ctx context.Context, entityType, entityID string) ([]*domain.ComplianceGap, error)

	// RecommendRemediations recommends remediation actions
	RecommendRemediations(ctx context.Context, gaps []*domain.ComplianceGap) ([]*domain.RemediationRecommendation, error)

	// TrackRemediation tracks remediation progress
	TrackRemediation(ctx context.Context, remediationID string) (*domain.RemediationProgress, error)

	// CompleteRemediation completes a remediation
	CompleteRemediation(ctx context.Context, remediationID string, result *domain.RemediationResult) error

	// GetComplianceHistory retrieves compliance history
	GetComplianceHistory(ctx context.Context, entityType, entityID string, timeRange domain.TimeRange) ([]*domain.ComplianceHistoryEntry, error)

	// CompareCompliance compares compliance across multiple entities
	CompareCompliance(ctx context.Context, entities []*domain.EntityRef) (*domain.ComplianceComparison, error)

	// GetComplianceBenchmark compares against benchmarks
	GetComplianceBenchmark(ctx context.Context, entityType, entityID string, benchmarkID string) (*domain.ComplianceBenchmark, error)

	// GetComplianceRequirements retrieves requirements for an entity
	GetComplianceRequirements(ctx context.Context, entityType, entityID string) ([]*domain.ComplianceRequirement, error)

	// AssessRequirement assesses compliance with a specific requirement
	AssessRequirement(ctx context.Context, entityType, entityID string, requirementID string) (*domain.RequirementAssessment, error)

	// GetExceptions gets compliance exceptions
	GetExceptions(ctx context.Context, entityType, entityID string) ([]*domain.ComplianceException, error)

	// RequestException requests a compliance exception
	RequestException(ctx context.Context, request *domain.ExceptionRequest) (*domain.ComplianceException, error)

	// ApproveException approves an exception request
	ApproveException(ctx context.Context, requestID string, approverID string, notes string) error

	// RejectException rejects an exception request
	RejectException(ctx context.Context, requestID string, reason string) error
}

// CrossCuttingService defines interfaces for cross-cutting concerns.
type CrossCuttingService interface {
	// GetIdempotencyService returns the idempotency service
	GetIdempotencyService() IdempotencyService

	// GetCorrelationService returns the correlation service
	GetCorrelationService() CorrelationService

	// GetDistributedLockService returns the distributed lock service
	GetDistributedLockService() DistributedLockService

	// GetCircuitBreakerService returns the circuit breaker service
	GetCircuitBreakerService() CircuitBreakerService

	// GetRetryService returns the retry service
	GetRetryService() RetryService
}

// IdempotencyService defines the interface for idempotency handling.
type IdempotencyService interface {
	// GetIdempotencyKey gets an idempotency key from request
	GetIdempotencyKey(ctx context.Context, request interface{}) (string, error)

	// Check checks if a request is idempotent
	Check(ctx context.Context, key string) (bool, error)

	// Record records an idempotent request
	Record(ctx context.Context, key string, response interface{}, ttl time.Duration) error

	// GetResponse gets a recorded response
	GetResponse(ctx context.Context, key string) (interface{}, bool, error)

	// Invalidate invalidates an idempotency key
	Invalidate(ctx context.Context, key string) error

	// Cleanup cleans up expired keys
	Cleanup(ctx context.Context) error
}

// CorrelationService defines the interface for correlation tracking.
type CorrelationService interface {
	// StartCorrelation starts a new correlation
	StartCorrelation(ctx context.Context, operation string) (context.Context, string)

	// GetCorrelationID gets the current correlation ID
	GetCorrelationID(ctx context.Context) string

	// AddContext adds context to correlation
	AddContext(ctx context.Context, key string, value interface{}) error

	// GetContext gets context from correlation
	GetContext(ctx context.Context) map[string]interface{}

	// EndCorrelation ends a correlation
	EndCorrelation(ctx context.Context, status string) error

	// GetCorrelations gets correlations matching a filter
	GetCorrelations(ctx context.Context, filter *domain.CorrelationFilter) ([]*domain.Correlation, error)
}

// DistributedLockService defines the interface for distributed locking.
type DistributedLockService interface {
	// Acquire acquires a lock
	Acquire(ctx context.Context, name string, ttl time.Duration) (string, error)

	// AcquireWithRetry acquires a lock with retry
	AcquireWithRetry(ctx context.Context, name string, ttl time.Duration, maxRetries int, retryDelay time.Duration) (string, error)

	// Release releases a lock
	Release(ctx context.Context, name, token string) error

	// Extend extends a lock
	Extend(ctx context.Context, name, token string, ttl time.Duration) error

	// IsLocked checks if a lock is held
	IsLocked(ctx context.Context, name string) (bool, error)

	// GetLocks gets all held locks
	GetLocks(ctx context.Context) ([]*domain.LockInfo, error)

	// ForceRelease forces the release of a lock
	ForceRelease(ctx context.Context, name string) error
}

// IntegrationService defines interfaces for external system integration.
type IntegrationService interface {
	// GetExternalAPIService returns the external API service
	GetExternalAPIService() ExternalAPIService

	// GetMessageQueueService returns the message queue service
	GetMessageQueueService() MessageQueueService

	// GetWebhookService returns the webhook service
	GetWebhookService() WebhookService

	// GetFileTransferService returns the file transfer service
	GetFileTransferService() FileTransferService
}

// ExternalAPIService defines the interface for external API calls.
type ExternalAPIService interface {
	// Call makes an external API call
	Call(ctx context.Context, request *domain.ExternalAPIRequest) (*domain.ExternalAPIResponse, error)

	// CallWithRetry makes an external API call with retry
	CallWithRetry(ctx context.Context, request *domain.ExternalAPIRequest, maxRetries int, backoff time.Duration) (*domain.ExternalAPIResponse, error)

	// GetRateLimit gets rate limit status for an API
	GetRateLimit(ctx context.Context, apiID string) (*domain.RateLimitStatus, error)

	// RegisterAPI registers an external API
	RegisterAPI(ctx context.Context, api *domain.ExternalAPIConfig) error

	// UpdateAPI updates an external API configuration
	UpdateAPI(ctx context.Context, apiID string, api *domain.ExternalAPIConfig) error

	// DeleteAPI deletes an external API configuration
	DeleteAPI(ctx context.Context, apiID string) error

	// ListAPIs lists all registered external APIs
	ListAPIs(ctx context.Context) ([]*domain.ExternalAPIConfig, error)
}

// MessageQueueService defines the interface for message queue operations.
type MessageQueueService interface {
	// Publish publishes a message to a queue
	Publish(ctx context.Context, queue string, message interface{}, options *domain.QueueOptions) error

	// PublishBatch publishes multiple messages to a queue
	PublishBatch(ctx context.Context, queue string, messages []interface{}, options *domain.QueueOptions) error

	// Subscribe subscribes to a queue
	Subscribe(ctx context.Context, queue string, handler func(interface{}) error, options *domain.SubscribeOptions) (string, error)

	// Unsubscribe unsubscribes from a queue
	Unsubscribe(ctx context.Context, subscriptionID string) error

	// Consume consumes messages from a queue
	Consume(ctx context.Context, queue string, handler func(interface{}) error, options *domain.ConsumeOptions) error

	// GetQueueInfo gets information about a queue
	GetQueueInfo(ctx context.Context, queue string) (*domain.QueueInfo, error)

	// ListQueues lists all queues
	ListQueues(ctx context.Context) ([]*domain.QueueInfo, error)

	// PurgeQueue purges all messages from a queue
	PurgeQueue(ctx context.Context, queue string) error

	// DeleteQueue deletes a queue
	DeleteQueue(ctx context.Context, queue string) error
}

// WebhookService defines the interface for webhook operations.
type WebhookService interface {
	// Register registers a webhook
	Register(ctx context.Context, webhook *domain.WebhookRegistration) error

	// Update updates a webhook
	Update(ctx context.Context, webhookID string, webhook *domain.WebhookRegistration) error

	// Delete deletes a webhook
	Delete(ctx context.Context, webhookID string) error

	// Get gets a webhook
	Get(ctx context.Context, webhookID string) (*domain.WebhookRegistration, error)

	// List lists all webhooks
	List(ctx context.Context, eventType string) ([]*domain.WebhookRegistration, error)

	// Test tests a webhook
	Test(ctx context.Context, webhookID string, payload interface{}) error

	// Trigger triggers a webhook manually
	Trigger(ctx context.Context, webhookID string, payload interface{}) error

	// GetLogs gets webhook delivery logs
	GetLogs(ctx context.Context, webhookID string, filter *domain.WebhookLogFilter) ([]*domain.WebhookLog, error)

	// Retry retries a failed webhook delivery
	Retry(ctx context.Context, deliveryID string) error
}

// FileTransferService defines the interface for file transfer operations.
type FileTransferService interface {
	// Upload uploads a file
	Upload(ctx context.Context, file *domain.FileUpload) error

	// Download downloads a file
	Download(ctx context.Context, fileID string) (*domain.FileDownload, error)

	// Delete deletes a file
	Delete(ctx context.Context, fileID string) error

	// GetInfo gets file information
	GetInfo(ctx context.Context, fileID string) (*domain.FileInfo, error)

	// List lists files
	List(ctx context.Context, filter *domain.FileFilter) ([]*domain.FileInfo, error)

	// GetDownloadURL gets a download URL for a file
	GetDownloadURL(ctx context.Context, fileID string, expiration time.Duration) (string, error)

	// GetUploadURL gets an upload URL for a file
	GetUploadURL(ctx context.Context, file *domain.FileUpload) (string, error)

	// ValidateFile validates a file
	ValidateFile(ctx context.Context, file *domain.FileUpload) ([]*domain.ValidationError, error)

	// ProcessFile processes a file
	ProcessFile(ctx context.Context, fileID string, processor string, options map[string]interface{}) error
}

// ReportingService defines the interface for reporting operations.
type ReportingService interface {
	// GenerateReport generates a report
	GenerateReport(ctx context.Context, request *domain.ReportRequest) ([]byte, error)

	// ScheduleReport schedules a report
	ScheduleReport(ctx context.Context, schedule *domain.ReportSchedule) error

	// GetScheduledReports gets scheduled reports
	GetScheduledReports(ctx context.Context, filter *domain.ReportScheduleFilter) ([]*domain.ReportSchedule, error)

	// CancelScheduledReport cancels a scheduled report
	CancelScheduledReport(ctx context.Context, scheduleID string) error

	// GetReportHistory gets report history
	GetReportHistory(ctx context.Context, filter *domain.ReportHistoryFilter) ([]*domain.ReportHistoryEntry, error)

	// GetReportTemplate gets a report template
	GetReportTemplate(ctx context.Context, templateID string) (*domain.ReportTemplate, error)

	// CreateReportTemplate creates a report template
	CreateReportTemplate(ctx context.Context, template *domain.ReportTemplate) error

	// UpdateReportTemplate updates a report template
	UpdateReportTemplate(ctx context.Context, templateID string, template *domain.ReportTemplate) error

	// DeleteReportTemplate deletes a report template
	DeleteReportTemplate(ctx context.Context, templateID string) error

	// ListReportTemplates lists report templates
	ListReportTemplates(ctx context.Context, filter *domain.ReportTemplateFilter) ([]*domain.ReportTemplate, error)
}

// NotificationService defines the interface for notification operations.
type NotificationService interface {
	// Send sends a notification
	Send(ctx context.Context, notification *domain.NotificationRequest) error

	// SendBatch sends batch notifications
	SendBatch(ctx context.Context, notifications []*domain.NotificationRequest) error

	// GetNotification gets a notification
	GetNotification(ctx context.Context, notificationID string) (*domain.Notification, error)

	// GetNotifications gets notifications
	GetNotifications(ctx context.Context, filter *domain.NotificationFilter) ([]*domain.Notification, error)

	// MarkAsRead marks a notification as read
	MarkAsRead(ctx context.Context, notificationID string) error

	// MarkAllAsRead marks all notifications as read
	MarkAllAsRead(ctx context.Context, userID string) error

	// GetPreferences gets notification preferences
	GetPreferences(ctx context.Context, userID string) (*domain.NotificationPreferences, error)

	// UpdatePreferences updates notification preferences
	UpdatePreferences(ctx context.Context, userID string, preferences *domain.NotificationPreferences) error

	// GetChannels gets available notification channels
	GetChannels(ctx context.Context) ([]*domain.NotificationChannel, error)
}

// TemplateService defines the interface for template management.
type TemplateService interface {
	// Create creates a template
	Create(ctx context.Context, template *domain.Template) error

	// Update updates a template
	Update(ctx context.Context, templateID string, template *domain.Template) error

	// Delete deletes a template
	Delete(ctx context.Context, templateID string) error

	// Get gets a template
	Get(ctx context.Context, templateID string) (*domain.Template, error)

	// List lists templates
	List(ctx context.Context, filter *domain.TemplateFilter) ([]*domain.Template, error)

	// Render renders a template
	Render(ctx context.Context, templateID string, data map[string]interface{}) (string, error)

	// RenderString renders a template string
	RenderString(ctx context.Context, templateString string, data map[string]interface{}) (string, error)

	// Validate validates a template
	Validate(ctx context.Context, template *domain.Template) ([]*domain.ValidationError, error)

	// Clone clones a template
	Clone(ctx context.Context, templateID string, newName string) (*domain.Template, error)

	// GetVersion gets a template version
	GetVersion(ctx context.Context, templateID string, versionID string) (*domain.TemplateVersion, error)

	// GetVersions gets all versions of a template
	GetVersions(ctx context.Context, templateID string) ([]*domain.TemplateVersion, error)
}
