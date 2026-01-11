package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// PointsProcessed tracks the total number of telemetry points processed
	PointsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transport_adapter_points_processed_total",
			Help: "Total number of telemetry points processed",
		},
		[]string{"source_id", "topic", "protocol"},
	)

	// PointsFiltered tracks the total number of filtered points
	PointsFiltered = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transport_adapter_points_filtered_total",
			Help: "Total number of telemetry points filtered out",
		},
		[]string{"source_id", "topic", "protocol", "filter_type"},
	)

	// SpikesDetected tracks spike detections
	SpikesDetected = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "transport_adapter_spikes_detected_total",
			Help: "Total number of spikes detected",
		},
	)

	// OutliersDetected tracks outlier detections
	OutliersDetected = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "transport_adapter_outliers_detected_total",
			Help: "Total number of outliers detected",
		},
	)

	// QualityFiltered tracks quality-filtered points
	QualityFiltered = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "transport_adapter_quality_filtered_total",
			Help: "Total number of points filtered by quality",
		},
	)

	// ProcessingLatency tracks processing latency
	ProcessingLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "transport_adapter_processing_latency_seconds",
			Help:    "Processing latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// MQTTConnected tracks MQTT connection status
	MQTTConnected = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "transport_adapter_mqtt_connected",
			Help: "MQTT connection status (1=connected, 0=disconnected)",
		},
		[]string{"broker_url", "client_id"},
	)

	// KafkaHealthy tracks Kafka producer health
	KafkaHealthy = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "transport_adapter_kafka_healthy",
			Help: "Kafka producer health status (1=healthy, 0=unhealthy)",
		},
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

	// MQTTMessagesReceived tracks MQTT messages received
	MQTTMessagesReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transport_adapter_mqtt_messages_received_total",
			Help: "Total number of MQTT messages received",
		},
		[]string{"broker_url", "topic", "qos"},
	)

	// ActiveSubscriptions tracks active MQTT subscriptions
	ActiveSubscriptions = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "transport_adapter_active_subscriptions",
			Help: "Number of active MQTT subscriptions",
		},
		[]string{"broker_url"},
	)

	// SessionRetries tracks session reconnection attempts
	SessionRetries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transport_adapter_session_retries_total",
			Help: "Total number of session reconnection attempts",
		},
		[]string{"broker_url", "client_id"},
	)

	// SessionState tracks session states
	SessionState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "transport_adapter_session_state",
			Help: "Session state (0=disconnected, 1=connecting, 2=connected, 3=reconnecting, 4=failed)",
		},
		[]string{"broker_url", "client_id"},
	)

	// CriticalityDistribution tracks messages by criticality level
	CriticalityDistribution = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transport_adapter_criticality_distribution_total",
			Help: "Distribution of messages by criticality level",
		},
		[]string{"criticality", "qos"},
	)

	// QoSByCriticality tracks QoS level distribution by criticality
	QoSByCriticality = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "transport_adapter_qos_by_criticality",
			Help: "Configured QoS levels by message criticality",
		},
		[]string{"criticality"},
	)
)

// RecordPointProcessed records a processed telemetry point
func RecordPointProcessed(sourceID, topic, protocol string) {
	PointsProcessed.WithLabelValues(sourceID, topic, protocol).Inc()
}

// RecordPointFiltered records a filtered telemetry point
func RecordPointFiltered(sourceID, topic, protocol, filterType string) {
	PointsFiltered.WithLabelValues(sourceID, topic, protocol, filterType).Inc()
}

// RecordSpikeDetected records a spike detection
func RecordSpikeDetected() {
	SpikesDetected.Inc()
}

// RecordOutlierDetected records an outlier detection
func RecordOutlierDetected() {
	OutliersDetected.Inc()
}

// RecordQualityFiltered records a quality-filtered point
func RecordQualityFiltered() {
	QualityFiltered.Inc()
}

// RecordProcessingLatency records processing latency
func RecordProcessingLatency(operation string, durationSeconds float64) {
	ProcessingLatency.WithLabelValues(operation).Observe(durationSeconds)
}

// SetMQTTConnected sets MQTT connection status
func SetMQTTConnected(brokerURL, clientID string, connected bool) {
	var value float64
	if connected {
		value = 1
	}
	MQTTConnected.WithLabelValues(brokerURL, clientID).Set(value)
}

// SetKafkaHealthy sets Kafka producer health status
func SetKafkaHealthy(healthy bool) {
	var value float64
	if healthy {
		value = 1
	}
	KafkaHealthy.Set(value)
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

// RecordMQTTMessageReceived records a received MQTT message
func RecordMQTTMessageReceived(brokerURL, topic string, qos int) {
	MQTTMessagesReceived.WithLabelValues(brokerURL, topic, string(rune('0'+qos))).Inc()
}

// SetActiveSubscriptions sets the number of active subscriptions
func SetActiveSubscriptions(brokerURL string, count float64) {
	ActiveSubscriptions.WithLabelValues(brokerURL).Set(count)
}

// RecordSessionRetry records a session retry attempt
func RecordSessionRetry(brokerURL, clientID string) {
	SessionRetries.WithLabelValues(brokerURL, clientID).Inc()
}

// SetSessionState sets the current session state
func SetSessionState(brokerURL, clientID string, state float64) {
	SessionState.WithLabelValues(brokerURL, clientID).Set(state)
}

// RecordCriticalityMessage records a message by its criticality level
func RecordCriticalityMessage(criticality string, qos int) {
	CriticalityDistribution.WithLabelValues(criticality, string(rune('0'+qos))).Inc()
}

// SetQoSByCriticality sets the configured QoS level for a criticality
func SetQoSByCriticality(criticality string, qos int) {
	QoSByCriticality.WithLabelValues(criticality).Set(float64(qos))
}
