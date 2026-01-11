package schema

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// SchemaMetrics contains all Prometheus metrics for schema validation
type SchemaMetrics struct {
	// Validation counters
	ValidationTotal    *prometheus.CounterVec
	ValidationSuccess  prometheus.Counter
	ValidationFailure  *prometheus.CounterVec

	// Validation timing
	ValidationDuration *prometheus.HistogramVec

	// Error metrics
	ErrorsByField     *prometheus.CounterVec
	ErrorsByType      *prometheus.CounterVec
	ErrorsByCategory  *prometheus.CounterVec

	// Schema metrics
	SchemaRegistered   prometheus.Counter
	SchemaActive       prometheus.GaugeVec

	// Quality metrics
	QualityScore       *prometheus.GaugeVec
	QualityScoreHist   prometheus.Histogram

	// Anomaly metrics
	AnomalyDetected    *prometheus.CounterVec
	AnomalyScore       *prometheus.GaugeVec
}

// NewSchemaMetrics creates and registers all schema validation metrics
func NewSchemaMetrics(namespace string) *SchemaMetrics {
	return &SchemaMetrics{
		ValidationTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "schema_validation_total",
				Help:      "Total number of schema validation attempts",
			},
			[]string{"source", "schema_id", "schema_version", "status"},
		),

		ValidationSuccess: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "schema_validation_success_total",
				Help:      "Total number of successful schema validations",
			},
		),

		ValidationFailure: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "schema_validation_failure_total",
				Help:      "Total number of failed schema validations",
			},
			[]string{"source", "schema_id", "schema_version", "error_type"},
		),

		ValidationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "schema_validation_duration_seconds",
				Help:      "Schema validation duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"source", "schema_id", "schema_type"},
		),

		ErrorsByField: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "schema_validation_errors_by_field_total",
				Help:      "Total validation errors by field",
			},
			[]string{"source", "schema_id", "field", "error_type"},
		),

		ErrorsByType: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "schema_validation_errors_by_type_total",
				Help:      "Total validation errors by error type",
			},
			[]string{"source", "schema_id", "error_type"},
		),

		ErrorsByCategory: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "schema_validation_errors_by_category_total",
				Help:      "Total validation errors by error category",
			},
			[]string{"source", "schema_id", "category"},
		),

		SchemaRegistered: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "schema_registered_total",
				Help:      "Total number of schemas registered",
			},
		),

		SchemaActive: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "schema_active",
				Help:      "Number of active schemas",
			},
			[]string{"schema_id"},
		),

		QualityScore: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "data_quality_score",
				Help:      "Current data quality score (0-100)",
			},
			[]string{"source", "dimension"},
		),

		QualityScoreHist: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "data_quality_score_histogram",
				Help:      "Distribution of data quality scores",
				Buckets:   []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
			},
		),

		AnomalyDetected: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "data_anomaly_detected_total",
				Help:      "Total anomalies detected",
			},
			[]string{"source", "anomaly_type", "severity"},
		),

		AnomalyScore: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "data_anomaly_score",
				Help:      "Current anomaly score for data source",
			},
			[]string{"source"},
		),
	}
}

// RecordValidation records a validation attempt
func (m *SchemaMetrics) RecordValidation(source, schemaID, schemaVersion string, success bool, durationSeconds float64, errorType string) {
	status := "success"
	if !success {
		status = "failure"
	}

	m.ValidationTotal.WithLabelValues(source, schemaID, schemaVersion, status).Inc()

	if success {
		m.ValidationSuccess.Inc()
	} else {
		m.ValidationFailure.WithLabelValues(source, schemaID, schemaVersion, errorType).Inc()
	}

	m.ValidationDuration.WithLabelValues(source, schemaID, "json").Observe(durationSeconds)
}

// RecordError records a validation error
func (m *SchemaMetrics) RecordError(source, schemaID, field, errorType string) {
	m.ErrorsByField.WithLabelValues(source, schemaID, field, errorType).Inc()
	m.ErrorsByType.WithLabelValues(source, schemaID, errorType).Inc()
}

// RecordErrorByCategory records an error with its category
func (m *SchemaMetrics) RecordErrorByCategory(source, schemaID string, category ErrorCategory) {
	m.ErrorsByCategory.WithLabelValues(source, schemaID, string(category)).Inc()
}

// RecordSchemaRegistered records a new schema registration
func (m *SchemaMetrics) RecordSchemaRegistered(schemaID string) {
	m.SchemaRegistered.Inc()
	m.SchemaActive.WithLabelValues(schemaID).Inc()
}

// RecordSchemaDeleted records a schema deletion
func (m *SchemaMetrics) RecordSchemaDeleted(schemaID string) {
	m.SchemaActive.WithLabelValues(schemaID).Dec()
}

// RecordQualityScore records a data quality score
func (m *SchemaMetrics) RecordQualityScore(source, dimension string, score float64) {
	m.QualityScore.WithLabelValues(source, dimension).Set(score)
	m.QualityScoreHist.Observe(score)
}

// RecordAnomaly records an anomaly detection
func (m *SchemaMetrics) RecordAnomaly(source, anomalyType, severity string) {
	m.AnomalyDetected.WithLabelValues(source, anomalyType, severity).Inc()
	m.AnomalyScore.WithLabelValues(source).Set(1.0)
}

// ClearAnomalyScore clears the anomaly score for a source
func (m *SchemaMetrics) ClearAnomalyScore(source string) {
	m.AnomalyScore.WithLabelValues(source).Set(0.0)
}

// SchemaMetricsRegistry holds all metrics for the schema system
type SchemaMetricsRegistry struct {
	SchemaMetrics *SchemaMetrics
	DataQualityMetrics *DataQualityMetrics
}

// NewSchemaMetricsRegistry creates a new metrics registry
func NewSchemaMetricsRegistry(namespace string) *SchemaMetricsRegistry {
	return &SchemaMetricsRegistry{
		SchemaMetrics:     NewSchemaMetrics(namespace),
		DataQualityMetrics: NewDataQualityMetrics(namespace),
	}
}

// DataQualityMetrics contains Prometheus metrics for data quality monitoring
type DataQualityMetrics struct {
	// Completeness metrics
	CompletenessScore    *prometheus.GaugeVec
	CompletenessTotal    *prometheus.CounterVec
	CompletenessMissing  *prometheus.CounterVec

	// Validity metrics
	ValidityScore       *prometheus.GaugeVec
	ValidityTotal       *prometheus.CounterVec
	ValidityFailed      *prometheus.CounterVec

	// Consistency metrics
	ConsistencyScore    *prometheus.GaugeVec
	ConsistencyTotal    *prometheus.CounterVec
	ConsistencyViolated *prometheus.CounterVec

	// Timeliness metrics
	TimelinessScore     *prometheus.GaugeVec
	TimelinessTotal     *prometheus.CounterVec
	TimelinessLate      *prometheus.CounterVec
	TimelinessLag       *prometheus.HistogramVec

	// Overall quality
	OverallQualityScore *prometheus.GaugeVec
	OverallQualityHist  prometheus.Histogram
}

// NewDataQualityMetrics creates data quality metrics
func NewDataQualityMetrics(namespace string) *DataQualityMetrics {
	return &DataQualityMetrics{
		CompletenessScore: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "data_quality_completeness_score",
				Help:      "Data completeness score (0-100) per source",
			},
			[]string{"source", "field"},
		),

		CompletenessTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "data_quality_completeness_total",
				Help:      "Total records checked for completeness",
			},
			[]string{"source"},
		),

		CompletenessMissing: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "data_quality_completeness_missing_total",
				Help:      "Total missing values detected",
			},
			[]string{"source", "field"},
		),

		ValidityScore: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "data_quality_validity_score",
				Help:      "Data validity score (0-100) per source",
			},
			[]string{"source", "rule"},
		),

		ValidityTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "data_quality_validity_total",
				Help:      "Total records validated for validity",
			},
			[]string{"source"},
		),

		ValidityFailed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "data_quality_validity_failed_total",
				Help:      "Total validity validation failures",
			},
			[]string{"source", "rule"},
		),

		ConsistencyScore: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "data_quality_consistency_score",
				Help:      "Data consistency score (0-100) per source",
			},
			[]string{"source", "check"},
		),

		ConsistencyTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "data_quality_consistency_total",
				Help:      "Total consistency checks performed",
			},
			[]string{"source"},
		),

		ConsistencyViolated: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "data_quality_consistency_violated_total",
				Help:      "Total consistency violations",
			},
			[]string{"source", "check"},
		),

		TimelinessScore: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "data_quality_timeliness_score",
				Help:      "Data timeliness score (0-100) per source",
			},
			[]string{"source"},
		),

		TimelinessTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "data_quality_timeliness_total",
				Help:      "Total timeliness checks performed",
			},
			[]string{"source"},
		),

		TimelinessLate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "data_quality_timeliness_late_total",
				Help:      "Total late data events",
			},
			[]string{"source"},
		),

		TimelinessLag: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "data_quality_timeliness_lag_seconds",
				Help:      "Time lag between event and ingestion in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"source"},
		),

		OverallQualityScore: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "data_quality_overall_score",
				Help:      "Overall data quality score (0-100) per source",
			},
			[]string{"source"},
		),

		OverallQualityHist: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "data_quality_overall_histogram",
				Help:      "Distribution of overall quality scores",
				Buckets:   []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
			},
		),
	}
}

// RecordCompleteness records completeness metrics
func (m *DataQualityMetrics) RecordCompleteness(source, field string, score float64, isMissing bool) {
	m.CompletenessScore.WithLabelValues(source, field).Set(score)
	m.CompletenessTotal.WithLabelValues(source).Inc()
	if isMissing {
		m.CompletenessMissing.WithLabelValues(source, field).Inc()
	}
}

// RecordValidity records validity metrics
func (m *DataQualityMetrics) RecordValidity(source, rule string, score float64, isValid bool) {
	m.ValidityScore.WithLabelValues(source, rule).Set(score)
	m.ValidityTotal.WithLabelValues(source).Inc()
	if !isValid {
		m.ValidityFailed.WithLabelValues(source, rule).Inc()
	}
}

// RecordConsistency records consistency metrics
func (m *DataQualityMetrics) RecordConsistency(source, check string, score float64, isConsistent bool) {
	m.ConsistencyScore.WithLabelValues(source, check).Set(score)
	m.ConsistencyTotal.WithLabelValues(source).Inc()
	if !isConsistent {
		m.ConsistencyViolated.WithLabelValues(source, check).Inc()
	}
}

// RecordTimeliness records timeliness metrics
func (m *DataQualityMetrics) RecordTimeliness(source string, score, lagSeconds float64, isLate bool) {
	m.TimelinessScore.WithLabelValues(source).Set(score)
	m.TimelinessTotal.WithLabelValues(source).Inc()
	m.TimelinessLag.WithLabelValues(source).Observe(lagSeconds)
	if isLate {
		m.TimelinessLate.WithLabelValues(source).Inc()
	}
}

// RecordOverallQuality records the overall quality score
func (m *DataQualityMetrics) RecordOverallQuality(source string, score float64) {
	m.OverallQualityScore.WithLabelValues(source).Set(score)
	m.OverallQualityHist.Observe(score)
}
