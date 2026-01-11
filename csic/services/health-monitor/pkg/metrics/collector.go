package metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector manages Prometheus metrics for the health monitor
type MetricsCollector struct {
	registry      *prometheus.Registry
	metricsPrefix string

	// Heartbeat metrics
	HeartbeatsReceived    *prometheus.CounterVec
	HeartbeatProcessingTime *prometheus.HistogramVec

	// Health check metrics
	HealthChecksTotal     *prometheus.CounterVec
	HealthCheckDuration   *prometheus.HistogramVec
	HealthCheckStatus     *prometheus.GaugeVec

	// Service status metrics
	ServiceStatus         *prometheus.GaugeVec
	ServiceUptime         *prometheus.GaugeVec
	ServiceResponseTime   *prometheus.GaugeVec
	ServiceCPUUsage       *prometheus.GaugeVec
	ServiceMemoryUsage    *prometheus.GaugeVec
	ServiceDiskUsage      *prometheus.GaugeVec

	// Alert metrics
	AlertsFired           *prometheus.CounterVec
	AlertsResolved        *prometheus.CounterVec
	ActiveAlerts          *prometheus.GaugeVec
	AlertCooldowns        *prometheus.GaugeVec

	// Outage metrics
	OutagesDetected       *prometheus.CounterVec
	OutagesResolved       *prometheus.CounterVec
	ActiveOutages         *prometheus.GaugeVec
	OutageDuration        *prometheus.HistogramVec

	// HTTP metrics
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPRequestsInProgress prometheus.Gauge

	// Kafka metrics
	KafkaMessagesConsumed *prometheus.CounterVec
	KafkaConsumerLag      prometheus.Gauge
}

// NewPrometheusCollector creates a new metrics collector
func NewPrometheusCollector(prefix string) *MetricsCollector {
	registry := prometheus.NewRegistry()

	m := &MetricsCollector{
		registry:      registry,
		metricsPrefix: prefix,
		HeartbeatsReceived: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_heartbeats_received_total", prefix),
				Help: "Total number of heartbeats received",
			},
			[]string{"service"},
		),
		HeartbeatProcessingTime: promauto.NewHistogramVec(
			registry,
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_heartbeat_processing_duration_ms", prefix),
				Help:    "Duration of heartbeat processing in milliseconds",
				Buckets: []float64{1, 5, 10, 25, 50, 100},
			},
			[]string{"service"},
		),
		HealthChecksTotal: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_health_checks_total", prefix),
				Help: "Total number of health checks",
			},
			[]string{"service", "status"},
		),
		HealthCheckDuration: promauto.NewHistogramVec(
			registry,
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_health_check_duration_ms", prefix),
				Help:    "Duration of health checks in milliseconds",
				Buckets: []float64{10, 50, 100, 250, 500, 1000, 2500},
			},
			[]string{"service"},
		),
		HealthCheckStatus: promauto.NewGaugeVec(
			registry,
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_health_check_status", prefix),
				Help: "Current health check status (0=unknown, 1=healthy, 2=degraded, 3=unhealthy)",
			},
			[]string{"service"},
		),
		ServiceStatus: promauto.NewGaugeVec(
			registry,
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_service_status", prefix),
				Help: "Current service status",
			},
			[]string{"service"},
		),
		ServiceUptime: promauto.NewGaugeVec(
			registry,
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_service_uptime_percent", prefix),
				Help: "Service uptime percentage",
			},
			[]string{"service"},
		),
		ServiceResponseTime: promauto.NewGaugeVec(
			registry,
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_service_response_time_ms", prefix),
				Help: "Service response time in milliseconds",
			},
			[]string{"service"},
		),
		ServiceCPUUsage: promauto.NewGaugeVec(
			registry,
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_service_cpu_percent", prefix),
				Help: "Service CPU usage percentage",
			},
			[]string{"service"},
		),
		ServiceMemoryUsage: promauto.NewGaugeVec(
			registry,
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_service_memory_percent", prefix),
				Help: "Service memory usage percentage",
			},
			[]string{"service"},
		),
		ServiceDiskUsage: promauto.NewGaugeVec(
			registry,
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_service_disk_percent", prefix),
				Help: "Service disk usage percentage",
			},
			[]string{"service"},
		),
		AlertsFired: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_alerts_fired_total", prefix),
				Help: "Total number of alerts fired",
			},
			[]string{"severity", "service"},
		),
		AlertsResolved: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_alerts_resolved_total", prefix),
				Help: "Total number of alerts resolved",
			},
			[]string{"severity", "service"},
		),
		ActiveAlerts: promauto.NewGaugeVec(
			registry,
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_active_alerts", prefix),
				Help: "Number of active alerts",
			},
			[]string{"severity"},
		),
		AlertCooldowns: promauto.NewGaugeVec(
			registry,
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_alert_cooldowns", prefix),
				Help: "Number of alerts in cooldown",
			},
			[]string{"rule_id"},
		),
		OutagesDetected: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_outages_detected_total", prefix),
				Help: "Total number of outages detected",
			},
			[]string{"service", "severity"},
		),
		OutagesResolved: promauto.NewCounterVec(
			registry,
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_outages_resolved_total", prefix),
				Help: "Total number of outages resolved",
			},
			[]string{"service"},
		),
		ActiveOutages: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_active_outages", prefix),
				Help: "Number of active outages",
			},
		),
		OutageDuration: promauto.NewHistogramVec(
			registry,
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_outage_duration_seconds", prefix),
				Help:    "Duration of outages in seconds",
				Buckets: []float64{60, 300, 600, 1800, 3600, 7200, 86400},
			},
			[]string{"service"},
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
	}

	return m
}

// GetRegistry returns the Prometheus registry
func (m *MetricsCollector) GetRegistry() *prometheus.Registry {
	return m.registry
}

// RecordHeartbeatProcessed records a processed heartbeat
func (m *MetricsCollector) RecordHeartbeatProcessed(serviceName string, durationMs float64) {
	m.HeartbeatsReceived.WithLabelValues(serviceName).Inc()
	m.HeartbeatProcessingTime.WithLabelValues(serviceName).Observe(durationMs)
}

// RecordHealthCheck records a health check result
func (m *MetricsCollector) RecordHealthCheck(serviceName, status string, durationMs float64) {
	m.HealthChecksTotal.WithLabelValues(serviceName, status).Inc()
	m.HealthCheckDuration.WithLabelValues(serviceName).Observe(durationMs)
}

// UpdateServiceStatus updates service status metrics
func (m *MetricsCollector) UpdateServiceStatus(serviceName, status string) {
	m.ServiceStatus.WithLabelValues(serviceName).Set(m.statusToFloat(status))
}

// UpdateServiceUptime updates service uptime metric
func (m *MetricsCollector) UpdateServiceUptime(serviceName string, uptimePercent float64) {
	m.ServiceUptime.WithLabelValues(serviceName).Set(uptimePercent)
}

// UpdateServiceResponseTime updates service response time metric
func (m *MetricsCollector) UpdateServiceResponseTime(serviceName string, responseTimeMs float64) {
	m.ServiceResponseTime.WithLabelValues(serviceName).Set(responseTimeMs)
}

// UpdateServiceCPUUsage updates service CPU usage metric
func (m *MetricsCollector) UpdateServiceCPUUsage(serviceName string, cpuPercent float64) {
	m.ServiceCPUUsage.WithLabelValues(serviceName).Set(cpuPercent)
}

// UpdateServiceMemoryUsage updates service memory usage metric
func (m *MetricsCollector) UpdateServiceMemoryUsage(serviceName string, memoryPercent float64) {
	m.ServiceMemoryUsage.WithLabelValues(serviceName).Set(memoryPercent)
}

// UpdateServiceDiskUsage updates service disk usage metric
func (m *MetricsCollector) UpdateServiceDiskUsage(serviceName string, diskPercent float64) {
	m.ServiceDiskUsage.WithLabelValues(serviceName).Set(diskPercent)
}

// RecordAlertFired records a fired alert
func (m *MetricsCollector) RecordAlertFired(severity, serviceName string) {
	m.AlertsFired.WithLabelValues(severity, serviceName).Inc()
	m.ActiveAlerts.WithLabelValues(severity).Inc()
}

// RecordAlertResolved records a resolved alert
func (m *MetricsCollector) RecordAlertResolved(severity, serviceName string) {
	m.AlertsResolved.WithLabelValues(severity, serviceName).Inc()
	m.ActiveAlerts.WithLabelValues(severity).Dec()
}

// RecordOutageDetected records a detected outage
func (m *MetricsCollector) RecordOutageDetected(serviceName, severity string) {
	m.OutagesDetected.WithLabelValues(serviceName, severity).Inc()
	m.ActiveOutages.Inc()
}

// RecordOutageResolved records a resolved outage
func (m *MetricsCollector) RecordOutageResolved(serviceName string, durationSeconds float64) {
	m.OutagesResolved.WithLabelValues(serviceName).Inc()
	m.ActiveOutages.Dec()
	m.OutageDuration.WithLabelValues(serviceName).Observe(durationSeconds)
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

// RecordKafkaMessageConsumed records a Kafka message consumed
func (m *MetricsCollector) RecordKafkaMessageConsumed(topic string) {
	m.KafkaMessagesConsumed.WithLabelValues(topic).Inc()
}

// SetKafkaConsumerLag sets the current Kafka consumer lag
func (m *MetricsCollector) SetKafkaConsumerLag(lag float64) {
	m.KafkaConsumerLag.Set(lag)
}

// statusToFloat converts status to float value
func (m *MetricsCollector) statusToFloat(status string) float64 {
	switch status {
	case "healthy":
		return 1
	case "degraded":
		return 2
	case "unhealthy":
		return 3
	default:
		return 0
	}
}

// MetricsCollectorOption is a function that modifies MetricsCollector
type MetricsCollectorOption func(*MetricsCollector)

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
		metricsInstance = NewPrometheusCollector("health_monitor")
	}

	return metricsInstance
}
