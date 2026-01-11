package metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector manages Prometheus metrics for the control layer
type MetricsCollector struct {
	registry      *prometheus.Registry
	metricsPrefix string

	// Policy metrics
	PolicyEvaluationsTotal    *prometheus.CounterVec
	PolicyEvaluationDuration  *prometheus.HistogramVec
	ActivePoliciesTotal       prometheus.Gauge

	// Enforcement metrics
	EnforcementActionsTotal   *prometheus.CounterVec
	EnforcementDuration       *prometheus.HistogramVec
	PendingEnforcements       prometheus.Gauge

	// Intervention metrics
	InterventionsTotal        *prometheus.CounterVec
	ActiveInterventions       prometheus.Gauge
	InterventionDuration      *prometheus.HistogramVec

	// State registry metrics
	StateUpdatesTotal         *prometheus.CounterVec
	StateCacheHitTotal        prometheus.Counter
	StateCacheMissTotal       prometheus.Counter

	// Kafka metrics
	KafkaMessagesProduced     *prometheus.CounterVec
	KafkaMessagesConsumed     *prometheus.CounterVec
	KafkaConsumerLag          prometheus.Gauge

	// HTTP metrics
	HTTPRequestsTotal         *prometheus.CounterVec
	HTTPRequestDuration       *prometheus.HistogramVec
	HTTPRequestsInProgress    prometheus.Gauge

	// gRPC metrics
	GRPCCallsTotal            *prometheus.CounterVec
	GRPCCallDuration          *prometheus.HistogramVec
	GRPCCallsInProgress       prometheus.Gauge
}

// NewPrometheusCollector creates a new metrics collector
func NewPrometheusCollector(prefix string) *MetricsCollector {
	registry := prometheus.NewRegistry()

	m := &MetricsCollector{
		registry:      registry,
		metricsPrefix: prefix,
		PolicyEvaluationsTotal: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_policy_evaluations_total", prefix),
				Help: "Total number of policy evaluations",
			},
			[]string{"policy_id", "result"},
		),
		PolicyEvaluationDuration: promauto.NewHistogramVec(
			registry,
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_policy_evaluation_duration_ms", prefix),
				Help:    "Duration of policy evaluations in milliseconds",
				Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
			},
			[]string{"policy_id"},
		),
		ActivePoliciesTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_active_policies_total", prefix),
				Help: "Total number of active policies",
			},
		),
		EnforcementActionsTotal: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_enforcement_actions_total", prefix),
				Help: "Total number of enforcement actions",
			},
			[]string{"policy_id", "action_type", "target"},
		),
		EnforcementDuration: promauto.NewHistogramVec(
			registry,
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_enforcement_duration_ms", prefix),
				Help:    "Duration of enforcement actions in milliseconds",
				Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
			},
			[]string{"action_type"},
		),
		PendingEnforcements: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_pending_enforcements", prefix),
				Help: "Number of pending enforcement actions",
			},
		),
		InterventionsTotal: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_interventions_total", prefix),
				Help: "Total number of interventions",
			},
			[]string{"policy_id", "severity", "status"},
		),
		ActiveInterventions: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_active_interventions", prefix),
				Help: "Number of active interventions",
			},
		),
		InterventionDuration: promauto.NewHistogramVec(
			registry,
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_intervention_duration_ms", prefix),
				Help:    "Duration of interventions in milliseconds",
				Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 5000},
			},
			[]string{"intervention_type"},
		),
		StateUpdatesTotal: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_state_updates_total", prefix),
				Help: "Total number of state updates",
			},
			[]string{"state_type"},
		),
		StateCacheHitTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_state_cache_hits_total", prefix),
				Help: "Total number of state cache hits",
			},
		),
		StateCacheMissTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_state_cache_misses_total", prefix),
				Help: "Total number of state cache misses",
			},
		),
		KafkaMessagesProduced: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_kafka_messages_produced_total", prefix),
				Help: "Total number of Kafka messages produced",
			},
			[]string{"topic"},
		),
		KafkaMessagesConsumed: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_kafka_messages_consumed_total", prefix),
				Help: "Total number of Kafka messages consumed",
			},
			[]string{"topic"},
		),
		KafkaConsumerLag: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_kafka_consumer_lag", prefix),
				Help: "Current Kafka consumer lag",
			},
		),
		HTTPRequestsTotal: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_http_requests_total", prefix),
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			registry,
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_http_request_duration_ms", prefix),
				Help:    "Duration of HTTP requests in milliseconds",
				Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInProgress: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_http_requests_in_progress", prefix),
				Help: "Number of HTTP requests currently in progress",
			},
		),
		GRPCCallsTotal: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_grpc_calls_total", prefix),
				Help: "Total number of gRPC calls",
			},
			[]string{"method", "status"},
		),
		GRPCCallDuration: promauto.NewHistogramVec(
			registry,
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_grpc_call_duration_ms", prefix),
				Help:    "Duration of gRPC calls in milliseconds",
				Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
			},
			[]string{"method"},
		),
		GRPCCallsInProgress: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_grpc_calls_in_progress", prefix),
				Help: "Number of gRPC calls currently in progress",
			},
		),
	}

	return m
}

// GetRegistry returns the Prometheus registry
func (m *MetricsCollector) GetRegistry() *prometheus.Registry {
	return m.registry
}

// RecordPolicyEvaluation records a policy evaluation
func (m *MetricsCollector) RecordPolicyEvaluation(policyID string, result string, durationMs float64) {
	m.PolicyEvaluationsTotal.WithLabelValues(policyID, result).Inc()
	m.PolicyEvaluationDuration.WithLabelValues(policyID).Observe(durationMs)
}

// RecordEnforcementAction records an enforcement action
func (m *MetricsCollector) RecordEnforcementAction(policyID, actionType, target string, durationMs float64) {
	m.EnforcementActionsTotal.WithLabelValues(policyID, actionType, target).Inc()
	m.EnforcementDuration.WithLabelValues(actionType).Observe(durationMs)
}

// RecordIntervention records an intervention
func (m *MetricsCollector) RecordIntervention(policyID, severity, status string, durationMs float64) {
	m.InterventionsTotal.WithLabelValues(policyID, severity, status).Inc()
	m.InterventionDuration.WithLabelValues(severity).Observe(durationMs)
}

// UpdateStateCacheHit records a state cache hit
func (m *MetricsCollector) UpdateStateCacheHit() {
	m.StateCacheHitTotal.Inc()
}

// UpdateStateCacheMiss records a state cache miss
func (m *MetricsCollector) UpdateStateCacheMiss() {
	m.StateCacheMissTotal.Inc()
}

// RecordKafkaMessageProduced records a Kafka message produced
func (m *MetricsCollector) RecordKafkaMessageProduced(topic string) {
	m.KafkaMessagesProduced.WithLabelValues(topic).Inc()
}

// RecordKafkaMessageConsumed records a Kafka message consumed
func (m *MetricsCollector) RecordKafkaMessageConsumed(topic string) {
	m.KafkaMessagesConsumed.WithLabelValues(topic).Inc()
}

// SetKafkaConsumerLag sets the current Kafka consumer lag
func (m *MetricsCollector) SetKafkaConsumerLag(lag float64) {
	m.KafkaConsumerLag.Set(lag)
}

// RecordHTTPRequest records an HTTP request
func (m *MetricsCollector) RecordHTTPRequest(method, endpoint, status string, durationMs float64) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(durationMs)
}

// IncHTTPRequestsInProgress increments the in-progress HTTP requests counter
func (m *MetricsCollector) IncHTTPRequestsInProgress() {
	m.HTTPRequestsInProgress.Inc()
}

// DecHTTPRequestsInProgress decrements the in-progress HTTP requests counter
func (m *MetricsCollector) DecHTTPRequestsInProgress() {
	m.HTTPRequestsInProgress.Dec()
}

// RecordGRPCCall records a gRPC call
func (m *MetricsCollector) RecordGRPCCall(method, status string, durationMs float64) {
	m.GRPCCallsTotal.WithLabelValues(method, status).Inc()
	m.GRPCCallDuration.WithLabelValues(method).Observe(durationMs)
}

// IncGRPCCallsInProgress increments the in-progress gRPC calls counter
func (m *MetricsCollector) IncGRPCCallsInProgress() {
	m.GRPCCallsInProgress.Inc()
}

// DecGRPCCallsInProgress decrements the in-progress gRPC calls counter
func (m *MetricsCollector) DecGRPCCallsInProgress() {
	m.GRPCCallsInProgress.Dec()
}

// SetActivePolicies sets the number of active policies
func (m *MetricsCollector) SetActivePolicies(count float64) {
	m.ActivePoliciesTotal.Set(count)
}

// SetPendingEnforcements sets the number of pending enforcements
func (m *MetricsCollector) SetPendingEnforcements(count float64) {
	m.PendingEnforcements.Set(count)
}

// SetActiveInterventions sets the number of active interventions
func (m *MetricsCollector) SetActiveInterventions(count float64) {
	m.ActiveInterventions.Set(count)
}

// MetricsCollectorOption is a function that modifies MetricsCollector
type MetricsCollectorOption func(*MetricsCollector)

// WithPrefix sets the metrics prefix
func WithPrefix(prefix string) MetricsCollectorOption {
	return func(m *MetricsCollector) {
		m.metricsPrefix = prefix
	}
}

// NewMockMetricsCollector creates a mock metrics collector for testing
func NewMockMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		registry:      prometheus.NewRegistry(),
		metricsPrefix: "test",
	}
}

var (
	metricsLock     sync.Mutex
	metricsInstance *MetricsCollector
)

// GetMetricsInstance returns a singleton metrics instance
func GetMetricsInstance() *MetricsCollector {
	metricsLock.Lock()
	defer metricsLock.Unlock()

	if metricsInstance == nil {
		metricsInstance = NewPrometheusCollector("control_layer")
	}

	return metricsInstance
}
