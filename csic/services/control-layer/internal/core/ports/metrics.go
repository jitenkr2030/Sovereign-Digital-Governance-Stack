package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
)

// MetricsCollector defines the interface for collecting metrics.
// This interface provides methods for recording various types of metrics.
type MetricsCollector interface {
	// Counter creates and returns a counter metric
	Counter(name string, opts ...MetricsOption) Counter

	// Gauge creates and returns a gauge metric
	Gauge(name string, opts ...MetricsOption) Gauge

	// Histogram creates and returns a histogram metric
	Histogram(name string, opts ...MetricsOption) Histogram

	// Summary creates and returns a summary metric
	Summary(name string, opts ...MetricsOption) Summary

	// Timer creates and returns a timer metric
	Timer(name string, opts ...MetricsOption) Timer

	// SetDefaultTags sets default tags for all metrics
	SetDefaultTags(tags map[string]string)

	// GetDefaultTags returns the default tags
	GetDefaultTags() map[string]string

	// Describe returns metric descriptions
	Describe() []*domain.MetricDescriptor

	// Collect returns collected metrics
	Collect() []*domain.MetricFamily
}

// Counter defines the interface for counter metrics.
type Counter interface {
	// Add adds a value to the counter
	Add(v float64)

	// Inc increments the counter by 1
	Inc()

	// Value returns the current value
	Value() float64

	// WithTags returns a counter with additional tags
	WithTags(tags map[string]string) Counter

	// Reset resets the counter
	Reset()
}

// Gauge defines the interface for gauge metrics.
type Gauge interface {
	// Set sets the gauge value
	Set(v float64)

	// Add adds a value to the gauge
	Add(v float64)

	// Sub subtracts a value from the gauge
	Sub(v float64)

	// Inc increments the gauge by 1
	Inc()

	// Dec decrements the gauge by 1
	Dec()

	// Value returns the current value
	Value() float64

	// WithTags returns a gauge with additional tags
	WithTags(tags map[string]string) Gauge

	// Reset resets the gauge
	Reset()
}

// Histogram defines the interface for histogram metrics.
type Histogram interface {
	// Observe observes a value
	Observe(v float64)

	// WithTags returns a histogram with additional tags
	WithTags(tags map[string]string) Histogram

	// GetCount returns the observation count
	GetCount() int64

	// GetSum returns the sum of observations
	GetSum() float64

	// GetBuckets returns the bucket values
	GetBuckets() map[float64]int64

	// Reset resets the histogram
	Reset()
}

// Summary defines the interface for summary metrics.
type Summary interface {
	// Observe observes a value
	Observe(v float64)

	// WithTags returns a summary with additional tags
	WithTags(tags map[string]string) Summary

	// GetCount returns the observation count
	GetCount() int64

	// GetSum returns the sum of observations
	GetSum() float64

	// GetQuantiles returns the quantile values
	GetQuantiles() map[float64]float64

	// Reset resets the summary
	Reset()
}

// Timer defines the interface for timer metrics.
type Timer interface {
	// ObserveDuration observes a duration
	ObserveDuration(d time.Duration)

	// Time executes a function and observes its duration
	Time(f func())

	// TimeContext executes a function with context and observes its duration
	TimeContext(ctx context.Context, f func(ctx context.Context)) error

	// WithTags returns a timer with additional tags
	WithTags(tags map[string]string) Timer

	// GetCount returns the observation count
	GetCount() int64

	// GetSum returns the total duration
	GetSum() time.Duration

	// GetMin returns the minimum duration
	GetMin() time.Duration

	// GetMax returns the maximum duration
	GetMax() time.Duration

	// GetMean returns the mean duration
	GetMean() time.Duration

	// Reset resets the timer
	Reset()
}

// MetricsOption defines options for metrics.
type MetricsOption interface {
	Apply(opts *domain.MetricOptions)
}

// MetricsStorage defines the interface for storing metrics.
type MetricsStorage interface {
	// Store stores a metric
	Store(ctx context.Context, metric *domain.Metric) error

	// StoreBatch stores multiple metrics
	StoreBatch(ctx context.Context, metrics []*domain.Metric) error

	// Query queries metrics
	Query(ctx context.Context, query *domain.MetricsQuery) ([]*domain.Metric, error)

	// GetLatest gets the latest value of a metric
	GetLatest(ctx context.Context, name string, tags map[string]string) (*domain.Metric, error)

	// GetRange gets metric values in a range
	GetRange(ctx context.Context, name string, tags map[string]string, start, end time.Time) ([]*domain.Metric, error)

	// Delete deletes metrics
	Delete(ctx context.Context, query *domain.MetricsDeleteQuery) error

	// GetStorageStats returns storage statistics
	GetStorageStats(ctx context.Context) (*domain.MetricsStorageStats, error)
}

// TimeSeriesDB defines the interface for time series database operations.
type TimeSeriesDB interface {
	MetricsStorage

	// CreateRetentionPolicy creates a retention policy
	CreateRetentionPolicy(ctx context.Context, policy *domain.RetentionPolicy) error

	// DeleteRetentionPolicy deletes a retention policy
	DeleteRetentionPolicy(ctx context.Context, name string) error

	// GetRetentionPolicies returns retention policies
	GetRetentionPolicies(ctx context.Context) ([]*domain.RetentionPolicy, error)

	// CreateContinuousQuery creates a continuous query
	CreateContinuousQuery(ctx context.Context, query *domain.ContinuousQuery) error

	// DeleteContinuousQuery deletes a continuous query
	DeleteContinuousQuery(ctx context.Context, name string) error

	// GetContinuousQueries returns continuous queries
	GetContinuousQueries(ctx context.Context) ([]*domain.ContinuousQuery, error)
}

// AlertManager defines the interface for alert management.
type AlertManager interface {
	// CreateAlertRule creates an alert rule
	CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error

	// UpdateAlertRule updates an alert rule
	UpdateAlertRule(ctx context.Context, id string, rule *domain.AlertRule) error

	// DeleteAlertRule deletes an alert rule
	DeleteAlertRule(ctx context.Context, id string) error

	// GetAlertRule gets an alert rule
	GetAlertRule(ctx context.Context, id string) (*domain.AlertRule, error)

	// ListAlertRules lists all alert rules
	ListAlertRules(ctx context.Context, filter *domain.AlertRuleFilter) ([]*domain.AlertRule, error)

	// EnableAlertRule enables an alert rule
	EnableAlertRule(ctx context.Context, id string) error

	// DisableAlertRule disables an alert rule
	DisableAlertRule(ctx context.Context, id string) error

	// GetActiveAlerts gets active alerts
	GetActiveAlerts(ctx context.Context, filter *domain.AlertFilter) ([]*domain.Alert, error)

	// GetAlertHistory gets alert history
	GetAlertHistory(ctx context.Context, filter *domain.AlertHistoryFilter) ([]*domain.Alert, error)

	// AcknowledgeAlert acknowledges an alert
	AcknowledgeAlert(ctx context.Context, id string, userID string) error

	// ResolveAlert resolves an alert
	ResolveAlert(ctx context.Context, id string, userID string, resolution string) error

	// SilenceAlert silences an alert
	SilenceAlert(ctx context.Context, id string, until time.Time, reason string) error

	// GetAlertMetrics gets alert metrics
	GetAlertMetrics(ctx context.Context) (*domain.AlertMetrics, error)
}

// AlertNotifier defines the interface for alert notifications.
type AlertNotifier interface {
	// Notify sends an alert notification
	Notify(ctx context.Context, alert *domain.Alert) error

	// NotifyBatch sends batch notifications
	NotifyBatch(ctx context.Context, alerts []*domain.Alert) error

	// GetNotifiers returns available notifiers
	GetNotifiers(ctx context.Context) ([]*domain.NotifierConfig, error)

	// CreateNotifier creates a notifier
	CreateNotifier(ctx context.Context, notifier *domain.NotifierConfig) error

	// UpdateNotifier updates a notifier
	UpdateNotifier(ctx context.Context, id string, notifier *domain.NotifierConfig) error

	// DeleteNotifier deletes a notifier
	DeleteNotifier(ctx context.Context, id string) error

	// TestNotifier tests a notifier
	TestNotifier(ctx context.Context, id string) error
}

// DashboardManager defines the interface for dashboard management.
type DashboardManager interface {
	// CreateDashboard creates a dashboard
	CreateDashboard(ctx context.Context, dashboard *domain.Dashboard) error

	// UpdateDashboard updates a dashboard
	UpdateDashboard(ctx context.Context, id string, dashboard *domain.Dashboard) error

	// DeleteDashboard deletes a dashboard
	DeleteDashboard(ctx context.Context, id string) error

	// GetDashboard gets a dashboard
	GetDashboard(ctx context.Context, id string) (*domain.Dashboard, error)

	// ListDashboards lists all dashboards
	ListDashboards(ctx context.Context, filter *domain.DashboardFilter) ([]*domain.Dashboard, error)

	// CloneDashboard clones a dashboard
	CloneDashboard(ctx context.Context, id string, newName string) (*domain.Dashboard, error)

	// ExportDashboard exports a dashboard
	ExportDashboard(ctx context.Context, id string, format string) ([]byte, error)

	// ImportDashboard imports a dashboard
	ImportDashboard(ctx context.Context, data []byte, format string) (*domain.Dashboard, error)

	// ShareDashboard shares a dashboard
	ShareDashboard(ctx context.Context, id string, shareType string, users []string) error
}

// PanelRenderer defines the interface for rendering dashboard panels.
type PanelRenderer interface {
	// Render renders a panel
	Render(ctx context.Context, panel *domain.Panel, data map[string]interface{}) ([]byte, error)

	// GetPanelTypes returns available panel types
	GetPanelTypes(ctx context.Context) []string

	// ValidatePanel validates a panel configuration
	ValidatePanel(panel *domain.Panel) error

	// GetDefaultPanel returns a default panel configuration
	GetDefaultPanel(panelType string) *domain.Panel
}

// QueryEngine defines the interface for metrics query operations.
type QueryEngine interface {
	// ExecuteQuery executes a query
	ExecuteQuery(ctx context.Context, query *domain.Query) (*domain.QueryResult, error)

	// ParseQuery parses a query string
	ParseQuery(query string) (*domain.Query, error)

	// GetFunctions returns available query functions
	GetFunctions(ctx context.Context) []*domain.QueryFunction

	// GetAggregations returns available aggregations
	GetAggregations(ctx context.Context) []*domain.Aggregation

	// ValidateQuery validates a query
	ValidateQuery(ctx context.Context, query *domain.Query) error
}

// LogManager defines the interface for log management.
type LogManager interface {
	// Write writes a log entry
	Write(ctx context.Context, entry *domain.LogEntry) error

	// WriteBatch writes batch log entries
	WriteBatch(ctx context.Context, entries []*domain.LogEntry) error

	// Query queries log entries
	Query(ctx context.Context, query *domain.LogQuery) ([]*domain.LogEntry, error)

	// GetLogStream returns a log stream
	GetLogStream(ctx context.Context, query *domain.LogQuery) (<-chan *domain.LogEntry, error)

	// GetLogStats returns log statistics
	GetLogStats(ctx context.Context, query *domain.LogQuery) (*domain.LogStats, error)

	// GetLogFields returns available log fields
	GetLogFields(ctx context.Context) []*domain.LogField

	// CreateIndex creates a log index
	CreateIndex(ctx context.Context, index *domain.LogIndex) error

	// DeleteIndex deletes a log index
	DeleteIndex(ctx context.Context, name string) error

	// GetIndices returns all indices
	GetIndices(ctx context.Context) ([]*domain.LogIndex, error)
}

// TraceManager defines the interface for distributed tracing.
type TraceManager interface {
	// StartSpan starts a new span
	StartSpan(ctx context.Context, operationName string, opts ...TraceOption) (context.Context, Span)

	// GetSpan returns the current span
	GetSpan(ctx context.Context) Span

	// Inject injects trace context
	Inject(ctx context.Context, carrier interface{}) error

	// Extract extracts trace context
	Extract(ctx context.Context, carrier interface{}) (context.Context, error)

	// RecordRecord records a trace event
	RecordRecord(ctx context.Context, event *domain.TraceEvent) error

	// GetTrace gets a trace by ID
	GetTrace(ctx context.Context, traceID string) (*domain.Trace, error)

	// QueryTraces queries traces
	QueryTraces(ctx context.Context, query *domain.TraceQuery) ([]*domain.Trace, error)

	// GetServiceGraph returns the service graph
	GetServiceGraph(ctx context.Context, query *domain.ServiceGraphQuery) (*domain.ServiceGraph, error)
}

// Span defines the interface for trace spans.
type Span interface {
	// End ends the span
	End()

	// SetTag sets a tag
	SetTag(key string, value interface{})

	// SetTags sets multiple tags
	SetTags(tags map[string]interface{})

	// SetBaggageItem sets a baggage item
	SetBaggageItem(key, value string)

	// GetBaggageItem gets a baggage item
	GetBaggageItem(key string) string

	// LogFields logs fields
	LogFields(fields ...interface{})

	// LogKV logs key-value pairs
	LogKV(alternatingKeyValues ...interface{})

	// AddLink adds a link
	AddLink(link SpanContext)

	// Context returns the span context
	Context() SpanContext

	// GetOperationName returns the operation name
	GetOperationName() string

	// GetDuration returns the duration
	GetDuration() time.Duration

	// GetStartTime returns the start time
	GetStartTime() time.Time
}

// SpanContext defines the interface for span context.
type SpanContext interface {
	// GetTraceID returns the trace ID
	GetTraceID() string

	// GetSpanID returns the span ID
	GetSpanID() string

	// GetParentID returns the parent span ID
	GetParentID() string

	// GetBaggage returns baggage items
	GetBaggage() map[string]string

	// IsSampled returns whether the span is sampled
	IsSampled() bool

	// WithSampled returns a new context with sampling decision
	WithSampled(sampled bool) SpanContext
}

// TraceOption defines options for tracing.
type TraceOption interface {
	Apply(opts *domain.TraceOptions)
}

// ServiceMonitor defines the interface for service monitoring.
type ServiceMonitor interface {
	// GetServiceHealth gets the health of a service
	GetServiceHealth(ctx context.Context, serviceName string) (*domain.ServiceHealth, error)

	// GetAllServicesHealth gets health of all services
	GetAllServicesHealth(ctx context.Context) ([]*domain.ServiceHealth, error)

	// GetServiceMetrics gets metrics for a service
	GetServiceMetrics(ctx context.Context, serviceName string, query *domain.MetricsQuery) ([]*domain.Metric, error)

	// GetServiceDependencies gets service dependencies
	GetServiceDependencies(ctx context.Context, serviceName string) ([]string, error)

	// GetServiceTopology gets the service topology
	GetServiceTopology(ctx context.Context) (*domain.ServiceTopology, error)

	// GetServiceAlerts gets alerts for a service
	GetServiceAlerts(ctx context.Context, serviceName string) ([]*domain.Alert, error)
}

// HealthMonitor defines the interface for system health monitoring.
type HealthMonitor interface {
	// CheckHealth performs a health check
	CheckHealth(ctx context.Context) (*domain.SystemHealth, error)

	// GetComponentHealth gets health of a component
	GetComponentHealth(ctx context.Context, componentName string) (*domain.ComponentHealth, error)

	// RegisterComponent registers a component for monitoring
	RegisterComponent(ctx context.Context, component *domain.MonitoredComponent) error

	// UnregisterComponent unregisters a component
	UnregisterComponent(ctx context.Context, componentName string) error

	// GetComponents returns all registered components
	GetComponents(ctx context.Context) ([]*domain.MonitoredComponent, error)

	// GetHealthHistory gets health history
	GetHealthHistory(ctx context.Context, query *domain.HealthHistoryQuery) ([]*domain.HealthHistoryEntry, error)

	// GetUptime gets system uptime
	GetUptime(ctx context.Context) (time.Duration, error)
}

// EventRecorder defines the interface for recording events.
type EventRecorder interface {
	// RecordEvent records an event
	RecordEvent(ctx context.Context, event *domain.Event) error

	// RecordEventBatch records batch events
	RecordEventBatch(ctx context.Context, events []*domain.Event) error

	// GetEvents gets events
	GetEvents(ctx context.Context, query *domain.EventQuery) ([]*domain.Event, error)

	// GetEvent gets an event by ID
	GetEvent(ctx context.Context, id string) (*domain.Event, error)

	// GetEventStats returns event statistics
	GetEventStats(ctx context.Context, query *domain.EventQuery) (*domain.EventStats, error)

	// Subscribe subscribes to events
	Subscribe(ctx context.Context, query *domain.EventQuery) (<-chan *domain.Event, error)

	// Unsubscribe unsubscribes from events
	Unsubscribe(subscriptionID string) error
}

// NotificationManager defines the interface for notification management.
type NotificationManager interface {
	// SendNotification sends a notification
	SendNotification(ctx context.Context, notification *domain.Notification) error

	// SendBatch sends batch notifications
	SendBatch(ctx context.Context, notifications []*domain.Notification) error

	// GetChannels returns available notification channels
	GetChannels(ctx context.Context) []*domain.NotificationChannel

	// CreateChannel creates a notification channel
	CreateChannel(ctx context.Context, channel *domain.NotificationChannel) error

	// UpdateChannel updates a notification channel
	UpdateChannel(ctx context.Context, id string, channel *domain.NotificationChannel) error

	// DeleteChannel deletes a notification channel
	DeleteChannel(ctx context.Context, id string) error

	// GetNotifications gets notifications
	GetNotifications(ctx context.Context, query *domain.NotificationQuery) ([]*domain.Notification, error)

	// MarkAsRead marks a notification as read
	MarkAsRead(ctx context.Context, id string) error

	// MarkAllAsRead marks all notifications as read
	MarkAllAsRead(ctx context.Context, userID string) error

	// GetPreferences gets notification preferences
	GetPreferences(ctx context.Context, userID string) (*domain.NotificationPreferences, error)

	// UpdatePreferences updates notification preferences
	UpdatePreferences(ctx context.Context, userID string, preferences *domain.NotificationPreferences) error
}

// ReportGenerator defines the interface for generating reports.
type ReportGenerator interface {
	// GenerateReport generates a report
	GenerateReport(ctx context.Context, config *domain.ReportConfig) ([]byte, error)

	// GetReportTemplate gets a report template
	GetReportTemplate(ctx context.Context, templateID string) (*domain.ReportTemplate, error)

	// CreateReportTemplate creates a report template
	CreateReportTemplate(ctx context.Context, template *domain.ReportTemplate) error

	// UpdateReportTemplate updates a report template
	UpdateReportTemplate(ctx context.Context, id string, template *domain.ReportTemplate) error

	// DeleteReportTemplate deletes a report template
	DeleteReportTemplate(ctx context.Context, id string) error

	// ListReportTemplates lists report templates
	ListReportTemplates(ctx context.Context, filter *domain.ReportTemplateFilter) ([]*domain.ReportTemplate, error)

	// ScheduleReport schedules a report
	ScheduleReport(ctx context.Context, schedule *domain.ReportSchedule) error

	// GetScheduledReports gets scheduled reports
	GetScheduledReports(ctx context.Context) ([]*domain.ReportSchedule, error)

	// CancelScheduledReport cancels a scheduled report
	CancelScheduledReport(ctx context.Context, scheduleID string) error
}

// ConfigurationManager defines the interface for configuration management.
type ConfigurationManager interface {
	// GetConfig gets a configuration value
	GetConfig(ctx context.Context, key string) (interface{}, error)

	// SetConfig sets a configuration value
	SetConfig(ctx context.Context, key string, value interface{}) error

	// DeleteConfig deletes a configuration value
	DeleteConfig(ctx context.Context, key string) error

	// ListConfigs lists configurations
	ListConfigs(ctx context.Context, prefix string) ([]*domain.ConfigEntry, error)

	// GetConfigHistory gets configuration history
	GetConfigHistory(ctx context.Context, key string) ([]*domain.ConfigHistoryEntry, error)

	// ValidateConfig validates a configuration
	ValidateConfig(ctx context.Context, key string, value interface{}) error

	// ExportConfig exports configuration
	ExportConfig(ctx context.Context, prefix string) ([]byte, error)

	// ImportConfig imports configuration
	ImportConfig(ctx context.Context, data []byte, overwrite bool) error
}

// SecurityAuditor defines the interface for security auditing.
type SecurityAuditor interface {
	// AuditLog logs an audit event
	AuditLog(ctx context.Context, event *domain.SecurityAuditEvent) error

	// GetAuditTrail gets the audit trail
	GetAuditTrail(ctx context.Context, query *domain.AuditQuery) ([]*domain.SecurityAuditEvent, error)

	// GetAuditEvents gets security audit events
	GetAuditEvents(ctx context.Context, query *domain.AuditEventQuery) ([]*domain.SecurityAuditEvent, error)

	// AnalyzeThreats analyzes security threats
	AnalyzeThreats(ctx context.Context, query *domain.ThreatQuery) ([]*domain.Threat, error)

	// GetSecurityScore gets the security score
	GetSecurityScore(ctx context.Context) (*domain.SecurityScore, error)

	// GetVulnerabilities gets vulnerabilities
	GetVulnerabilities(ctx context.Context, query *domain.VulnerabilityQuery) ([]*domain.Vulnerability, error)

	// ReportIncident reports a security incident
	ReportIncident(ctx context.Context, incident *domain.SecurityIncident) error

	// GetIncidents gets security incidents
	GetIncidents(ctx context.Context, query *domain.IncidentQuery) ([]*domain.SecurityIncident, error)
}
