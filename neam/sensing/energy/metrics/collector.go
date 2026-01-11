package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// PointsProcessed tracks the total number of telemetry points processed
	PointsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "energy_adapter_points_processed_total",
			Help: "Total number of telemetry points processed",
		},
		[]string{"source_id", "point_id", "protocol"},
	)

	// PointsFiltered tracks the total number of filtered points
	PointsFiltered = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "energy_adapter_points_filtered_total",
			Help: "Total number of telemetry points filtered out",
		},
		[]string{"source_id", "point_id", "protocol", "filter_type"},
	)

	// SpikesDetected tracks spike detections
	SpikesDetected = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "energy_adapter_spikes_detected_total",
			Help: "Total number of spikes detected",
		},
	)

	// AnomaliesDetected tracks anomaly detections
	AnomaliesDetected = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "energy_adapter_anomalies_detected_total",
			Help: "Total number of anomalies detected",
		},
	)

	// QualityFiltered tracks quality-filtered points
	QualityFiltered = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "energy_adapter_quality_filtered_total",
			Help: "Total number of points filtered by quality",
		},
	)

	// ProcessingLatency tracks processing latency
	ProcessingLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "energy_adapter_processing_latency_seconds",
			Help:    "Processing latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// ConnectionHealth tracks connection health status
	ConnectionHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "energy_adapter_connection_health",
			Help: "Connection health status (1=healthy, 0=unhealthy)",
		},
		[]string{"connection_id", "protocol"},
	)

	// KafkaMessagesSent tracks Kafka messages sent
	KafkaMessagesSent = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_producer_messages_sent_total",
			Help: "Total number of Kafka messages sent",
		},
	)

	// KafkaMessagesFailed tracks failed Kafka messages
	KafkaMessagesFailed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_producer_messages_failed_total",
			Help: "Total number of failed Kafka messages",
		},
	)

	// KafkaBytesSent tracks bytes sent to Kafka
	KafkaBytesSent = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_producer_bytes_sent_total",
			Help: "Total number of bytes sent to Kafka",
		},
	)

	// KafkaLatency tracks Kafka write latency
	KafkaLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "kafka_producer_write_latency_seconds",
			Help:    "Kafka write latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
	)

	// ActiveConnections tracks active protocol connections
	ActiveConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "energy_adapter_active_connections",
			Help: "Number of active protocol connections",
		},
		[]string{"protocol"},
	)

	// SessionRetries tracks session reconnection attempts
	SessionRetries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "energy_adapter_session_retries_total",
			Help: "Total number of session reconnection attempts",
		},
		[]string{"protocol", "session_id"},
	)

	// SessionState tracks session states
	SessionState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "energy_adapter_session_state",
			Help: "Session state (0=disconnected, 1=connecting, 2=connected, 3=reconnecting, 4=failed)",
		},
		[]string{"session_id", "protocol"},
	)

	// PoolSize tracks connection pool size
	PoolSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "energy_adapter_pool_size",
			Help: "Current connection pool size",
		},
		[]string{"pool_type"}, // "idle" or "active"
	)

	// PoolCapacity tracks connection pool capacity
	PoolCapacity = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "energy_adapter_pool_capacity",
			Help: "Connection pool capacity",
		},
		[]string{"pool_type"},
	)
)

// RecordPointProcessed records a processed telemetry point
func RecordPointProcessed(sourceID, pointID, protocol string) {
	PointsProcessed.WithLabelValues(sourceID, pointID, protocol).Inc()
}

// RecordPointFiltered records a filtered telemetry point
func RecordPointFiltered(sourceID, pointID, protocol, filterType string) {
	PointsFiltered.WithLabelValues(sourceID, pointID, protocol, filterType).Inc()
}

// RecordSpikeDetected records a spike detection
func RecordSpikeDetected() {
	SpikesDetected.Inc()
}

// RecordAnomalyDetected records an anomaly detection
func RecordAnomalyDetected() {
	AnomaliesDetected.Inc()
}

// RecordQualityFiltered records a quality-filtered point
func RecordQualityFiltered() {
	QualityFiltered.Inc()
}

// RecordProcessingLatency records processing latency
func RecordProcessingLatency(operation string, durationSeconds float64) {
	ProcessingLatency.WithLabelValues(operation).Observe(durationSeconds)
}

// SetConnectionHealth sets connection health status
func SetConnectionHealth(connectionID string, protocol string, healthy bool) {
	var value float64
	if healthy {
		value = 1
	}
	ConnectionHealth.WithLabelValues(connectionID, protocol).Set(value)
}

// RecordKafkaMessageSent records a successful Kafka message
func RecordKafkaMessageSent() {
	KafkaMessagesSent.Inc()
}

// RecordKafkaMessageFailed records a failed Kafka message
func RecordKafkaMessageFailed() {
	KafkaMessagesFailed.Inc()
}

// RecordKafkaBytesSent records bytes sent to Kafka
func RecordKafkaBytesSent(bytes int64) {
	KafkaBytesSent.Add(float64(bytes))
}

// RecordKafkaLatency records Kafka write latency
func RecordKafkaLatency(durationSeconds float64) {
	KafkaLatency.Observe(durationSeconds)
}

// SetActiveConnections sets the number of active connections
func SetActiveConnections(protocol string, count float64) {
	ActiveConnections.WithLabelValues(protocol).Set(count)
}

// RecordSessionRetry records a session retry attempt
func RecordSessionRetry(protocol, sessionID string) {
	SessionRetries.WithLabelValues(protocol, sessionID).Inc()
}

// SetSessionState sets the current session state
func SetSessionState(sessionID, protocol string, state float64) {
	SessionState.WithLabelValues(sessionID, protocol).Set(state)
}

// SetPoolSize sets the current pool size
func SetPoolSize(poolType string, size float64) {
	PoolSize.WithLabelValues(poolType).Set(size)
}

// SetPoolCapacity sets the pool capacity
func SetPoolCapacity(poolType string, capacity float64) {
	PoolCapacity.WithLabelValues(poolType).Set(capacity)
}
