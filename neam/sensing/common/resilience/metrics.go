package resilience

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

// MetricsCollector collects resilience-related Prometheus metrics
type MetricsCollector struct {
	// Circuit breaker metrics
	circuitBreakerState       *prometheus.GaugeVec
	circuitBreakerRequests    *prometheus.CounterVec
	circuitBreakerFailures    *prometheus.CounterVec
	circuitBreakerLatency     *prometheus.HistogramVec
	circuitBreakerTransitions *prometheus.CounterVec

	// DLQ metrics
	dlqMessages       *prometheus.CounterVec
	dlqReprocess      *prometheus.CounterVec
	dlqProcessingTime *prometheus.HistogramVec
	dlqSize           *prometheus.GaugeVec

	// Validation metrics
	validationChecks   *prometheus.CounterVec
	validationFailures *prometheus.CounterVec
	validationLatency  *prometheus.HistogramVec

	// Error metrics
	errors *prometheus.CounterVec

	logger *zap.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	m := &MetricsCollector{
		logger: logger,

		circuitBreakerState: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "neam_resilience_circuit_breaker_state",
				Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
			},
			[]string{"adapter", "circuit"},
		),

		circuitBreakerRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "neam_resilience_circuit_breaker_requests_total",
				Help: "Total number of circuit breaker requests",
			},
			[]string{"adapter", "circuit", "result"}, // result: success, failure, fallback
		),

		circuitBreakerFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "neam_resilience_circuit_breaker_failures_total",
				Help: "Total number of circuit breaker failures",
			},
			[]string{"adapter", "circuit", "reason"}, // reason: timeout, error, circuit_open
		),

		circuitBreakerLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "neam_resilience_circuit_breaker_latency_seconds",
				Help:    "Circuit breaker request latency",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"adapter", "circuit"},
		),

		circuitBreakerTransitions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "neam_resilience_circuit_breaker_transitions_total",
				Help: "Total number of circuit breaker state transitions",
			},
			[]string{"adapter", "circuit", "from_state", "to_state"},
		),

		dlqMessages: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "neam_resilience_dlq_messages_total",
				Help: "Total number of messages sent to DLQ",
			},
			[]string{"adapter", "reason"}, // reason: schema, processing, timeout, etc.
		),

		dlqReprocess: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "neam_resilience_dlq_reprocess_total",
				Help: "Total number of DLQ message reprocess attempts",
			},
			[]string{"adapter", "result"}, // result: success, failure
		),

		dlqProcessingTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "neam_resilience_dlq_processing_seconds",
				Help:    "DLQ message processing time",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
			},
			[]string{"adapter", "operation"}, // operation: consume, reprocess, discard
		),

		dlqSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "neam_resilience_dlq_size",
				Help: "Current size of DLQ (approximate)",
			},
			[]string{"adapter"},
		),

		validationChecks: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "neam_resilience_validation_checks_total",
				Help: "Total number of validation checks",
			},
			[]string{"adapter", "schema_type"}, // schema_type: transport, industrial, energy
		),

		validationFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "neam_resilience_validation_failures_total",
				Help: "Total number of validation failures",
			},
			[]string{"adapter", "schema_type", "field", "reason"},
		),

		validationLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "neam_resilience_validation_latency_seconds",
				Help:    "Validation latency",
				Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
			},
			[]string{"adapter", "schema_type"},
		),

		errors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "neam_resilience_errors_total",
				Help: "Total number of resilience errors",
			},
			[]string{"adapter", "error_type", "operation"},
		),
	}

	return m
}

// RecordCircuitBreakerState records circuit breaker state changes
func (m *MetricsCollector) RecordCircuitBreakerState(name string, from, to gobreaker.State) {
	m.circuitBreakerState.WithLabelValues(name, name).Set(float64(to))
	m.circuitBreakerTransitions.WithLabelValues(
		name,
		name,
		from.String(),
		to.String(),
	).Inc()

	m.logger.Debug("Circuit breaker state recorded",
		zap.String("circuit", name),
		zap.String("from", from.String()),
		zap.String("to", to.String()))
}

// RecordCircuitBreakerRequest records a circuit breaker request
func (m *MetricsCollector) RecordCircuitBreakerRequest(
	name string,
	result interface{},
	err error,
	duration interface{},
) {
	var resultLabel string
	switch {
	case err == nil:
		resultLabel = "success"
	default:
		resultLabel = "failure"
	}

	m.circuitBreakerRequests.WithLabelValues(name, name, resultLabel).Inc()

	// Record latency if duration is available
	if d, ok := duration.(float64); ok {
		m.circuitBreakerLatency.WithLabelValues(name, name).Observe(d)
	}
}

// RecordCircuitBreakerFailure records a circuit breaker failure
func (m *MetricsCollector) RecordCircuitBreakerFailure(name string, reason string) {
	m.circuitBreakerFailures.WithLabelValues(name, name, reason).Inc()
}

// RecordDLQMessage records a DLQ message
func (m *MetricsCollector) RecordDLQMessage(adapter string, reason string) {
	m.dlqMessages.WithLabelValues(adapter, string(reason)).Inc()
}

// RecordDLQReprocess records a DLQ reprocess attempt
func (m *MetricsCollector) RecordDLQReprocess(adapter string, result string) {
	m.dlqReprocess.WithLabelValues(adapter, result).Inc()
}

// RecordDLQProcessingTime records DLQ processing time
func (m *MetricsCollector) RecordDLQProcessingTime(adapter string, operation string, durationSeconds float64) {
	m.dlqProcessingTime.WithLabelValues(adapter, operation).Observe(durationSeconds)
}

// SetDLQSize sets the current DLQ size
func (m *MetricsCollector) SetDLQSize(adapter string, size float64) {
	m.dlqSize.WithLabelValues(adapter).Set(size)
}

// RecordValidationCheck records a validation check
func (m *MetricsCollector) RecordValidationCheck(adapter string, schemaType string) {
	m.validationChecks.WithLabelValues(adapter, schemaType).Inc()
}

// RecordValidationFailure records a validation failure
func (m *MetricsCollector) RecordValidationFailure(adapter string, schemaType string, field string, reason string) {
	m.validationFailures.WithLabelValues(adapter, schemaType, field, reason).Inc()
}

// RecordValidationLatency records validation latency
func (m *MetricsCollector) RecordValidationLatency(adapter string, schemaType string, durationSeconds float64) {
	m.validationLatency.WithLabelValues(adapter, schemaType).Observe(durationSeconds)
}

// RecordError records a resilience error
func (m *MetricsCollector) RecordError(adapter string, errorType string, operation string) {
	m.errors.WithLabelValues(adapter, errorType, operation).Inc()
}

// MetricsServer provides HTTP endpoints for metrics
type MetricsServer struct {
	collector *MetricsCollector
	logger    *zap.Logger
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(collector *MetricsCollector, logger *zap.Logger) *MetricsServer {
	return &MetricsServer{
		collector: collector,
		logger:    logger,
	}
}

// Describe implements prometheus.Collector
func (m *MetricsServer) Describe(ch chan<- *prometheus.Desc) {
	m.circuitBreakerState.Describe(ch)
	m.circuitBreakerRequests.Describe(ch)
	m.circuitBreakerFailures.Describe(ch)
	m.circuitBreakerLatency.Describe(ch)
	m.circuitBreakerTransitions.Describe(ch)
	m.dlqMessages.Describe(ch)
	m.dlqReprocess.Describe(ch)
	m.dlqProcessingTime.Describe(ch)
	m.dlqSize.Describe(ch)
	m.validationChecks.Describe(ch)
	m.validationFailures.Describe(ch)
	m.validationLatency.Describe(ch)
	m.errors.Describe(ch)
}

// Collect implements prometheus.Collector
func (m *MetricsServer) Collect(ch chan<- prometheus.Metric) {
	m.circuitBreakerState.Collect(ch)
	m.circuitBreakerRequests.Collect(ch)
	m.circuitBreakerFailures.Collect(ch)
	m.circuitBreakerLatency.Collect(ch)
	m.circuitBreakerTransitions.Collect(ch)
	m.dlqMessages.Collect(ch)
	m.dlqReprocess.Collect(ch)
	m.dlqProcessingTime.Collect(ch)
	m.dlqSize.Collect(ch)
	m.validationChecks.Collect(ch)
	m.validationFailures.Collect(ch)
	m.validationLatency.Collect(ch)
	m.errors.Collect(ch)
}
