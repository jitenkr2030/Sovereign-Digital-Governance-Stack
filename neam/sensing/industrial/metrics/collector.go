package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// PointsProcessed tracks the total number of telemetry points processed
	PointsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "industrial_adapter_points_processed_total",
			Help: "Total number of telemetry points processed",
		},
		[]string{"source_id", "node_id", "protocol"},
	)

	// PointsFiltered tracks the total number of filtered points
	PointsFiltered = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "industrial_adapter_points_filtered_total",
			Help: "Total number of telemetry points filtered out",
		},
		[]string{"source_id", "node_id", "protocol", "filter_type"},
	)

	// SpikesDetected tracks spike detections
	SpikesDetected = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "industrial_adapter_spikes_detected_total",
			Help: "Total number of spikes detected",
		},
	)

	// OutliersDetected tracks outlier detections
	OutliersDetected = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "industrial_adapter_outliers_detected_total",
			Help: "Total number of outliers detected",
		},
	)

	// QualityFiltered tracks quality-filtered points
	QualityFiltered = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "industrial_adapter_quality_filtered_total",
			Help: "Total number of points filtered by quality",
		},
	)

	// ProcessingLatency tracks processing latency
	ProcessingLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "industrial_adapter_processing_latency_seconds",
			Help:    "Processing latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// OPCUAConnected tracks OPC UA connection status
	OPCUAConnected = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "industrial_adapter_opcua_connected",
			Help: "OPC UA connection status (1=connected, 0=disconnected)",
		},
		[]string{"endpoint_url", "server_name"},
	)

	// KafkaHealthy tracks Kafka producer health
	KafkaHealthy = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "industrial_adapter_kafka_healthy",
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

	// SubscriptionCount tracks active OPC UA subscriptions
	SubscriptionCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "industrial_adapter_subscription_count",
			Help: "Number of active OPC UA subscriptions",
		},
		[]string{"endpoint_url"},
	)

	// DiscoveryNodesFound tracks nodes discovered during asset discovery
	DiscoveryNodesFound = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "industrial_adapter_discovery_nodes_found_total",
			Help: "Total number of nodes found during discovery",
		},
		[]string{"endpoint_url", "node_type"},
	)

	// DiscoveryAssetsFound tracks assets discovered during asset discovery
	DiscoveryAssetsFound = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "industrial_adapter_discovery_assets_found_total",
			Help: "Total number of assets found during discovery",
		},
	)

	// DiscoveryErrors tracks discovery errors
	DiscoveryErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "industrial_adapter_discovery_errors_total",
			Help: "Total number of discovery errors",
		},
	)

	// DiscoveryDuration tracks discovery operation duration
	DiscoveryDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "industrial_adapter_discovery_duration_seconds",
			Help:    "Asset discovery operation duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
	)

	// SessionRetries tracks session reconnection attempts
	SessionRetries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "industrial_adapter_session_retries_total",
			Help: "Total number of session reconnection attempts",
		},
		[]string{"endpoint_url"},
	)

	// SessionState tracks session states
	SessionState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "industrial_adapter_session_state",
			Help: "Session state (0=disconnected, 1=connecting, 2=connected, 3=reconnecting, 4=failed)",
		},
		[]string{"endpoint_url"},
	)

	// MonitoredItemsCount tracks monitored items per subscription
	MonitoredItemsCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "industrial_adapter_monitored_items_count",
			Help: "Number of monitored items per subscription",
		},
		[]string{"endpoint_url", "subscription_id"},
	)

	// DataChanges tracks data change notifications received
	DataChanges = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "industrial_adapter_data_changes_total",
			Help: "Total number of data change notifications",
		},
		[]string{"endpoint_url", "node_id"},
	)
)

// RecordPointProcessed records a processed telemetry point
func RecordPointProcessed(sourceID, nodeID, protocol string) {
	PointsProcessed.WithLabelValues(sourceID, nodeID, protocol).Inc()
}

// RecordPointFiltered records a filtered telemetry point
func RecordPointFiltered(sourceID, nodeID, protocol, filterType string) {
	PointsFiltered.WithLabelValues(sourceID, nodeID, protocol, filterType).Inc()
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

// SetOPCUAConnected sets OPC UA connection status
func SetOPCUAConnected(endpointURL, serverName string, connected bool) {
	var value float64
	if connected {
		value = 1
	}
	OPCUAConnected.WithLabelValues(endpointURL, serverName).Set(value)
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

// SetSubscriptionCount sets the number of active subscriptions
func SetSubscriptionCount(endpointURL string, count float64) {
	SubscriptionCount.WithLabelValues(endpointURL).Set(count)
}

// RecordDiscoveryNodeFound records a discovered node
func RecordDiscoveryNodeFound(endpointURL, nodeType string) {
	DiscoveryNodesFound.WithLabelValues(endpointURL, nodeType).Inc()
}

// RecordDiscoveryAssetFound records a discovered asset
func RecordDiscoveryAssetFound() {
	DiscoveryAssetsFound.Inc()
}

// RecordDiscoveryError records a discovery error
func RecordDiscoveryError() {
	DiscoveryErrors.Inc()
}

// RecordDiscoveryDuration records discovery operation duration
func RecordDiscoveryDuration(durationSeconds float64) {
	DiscoveryDuration.Observe(durationSeconds)
}

// RecordSessionRetry records a session retry attempt
func RecordSessionRetry(endpointURL string) {
	SessionRetries.WithLabelValues(endpointURL).Inc()
}

// SetSessionState sets the current session state
func SetSessionState(endpointURL string, state float64) {
	SessionState.WithLabelValues(endpointURL).Set(state)
}

// SetMonitoredItemsCount sets the number of monitored items
func SetMonitoredItemsCount(endpointURL, subscriptionID string, count float64) {
	MonitoredItemsCount.WithLabelValues(endpointURL, subscriptionID).Set(count)
}

// RecordDataChange records a data change notification
func RecordDataChange(endpointURL, nodeID string) {
	DataChanges.WithLabelValues(endpointURL, nodeID).Inc()
}
