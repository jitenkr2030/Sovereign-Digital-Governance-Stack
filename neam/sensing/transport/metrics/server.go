package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Server provides HTTP server for metrics and health endpoints
type Server struct {
	cfg     ServerConfig
	logger  *zap.Logger
	server  *http.Server
}

// ServerConfig contains metrics server configuration
type ServerConfig struct {
	Port       int
	Path       string
	HealthPath string
}

// NewServer creates a new metrics server
func NewServer(cfg ServerConfig, logger *zap.Logger) *Server {
	mux := http.NewServeMux()

	// Metrics endpoint
	mux.Handle(cfg.Path, promhttp.Handler())

	// Health endpoint
	mux.HandleFunc(cfg.HealthPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	})

	// Readiness endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Add readiness checks here if needed
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ready"}`))
	})

	// Liveness endpoint
	mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "alive"}`))
	})

	return &Server{
		cfg: cfg,
		logger: logger,
		server: &http.Server{
			Addr:    ":" + itoa(cfg.Port),
			Handler: mux,
		},
	}
}

// Start begins serving metrics
func (s *Server) Start() error {
	s.logger.Info("Starting metrics server",
		zap.Int("port", s.cfg.Port),
		zap.String("path", s.cfg.Path))

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop gracefully shuts down the metrics server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping metrics server")
	return s.server.Shutdown(ctx)
}

// itoa converts int to string
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}

// DefaultServer creates a metrics server with default configuration
func DefaultServer(logger *zap.Logger) *Server {
	return NewServer(ServerConfig{
		Port:       9092,
		Path:       "/metrics",
		HealthPath: "/health",
	}, logger)
}

// Collector provides an interface for custom metrics collection
type Collector interface {
	Collect(ch chan<- prometheus.Metric)
	Describe(ch chan<- *prometheus.Desc)
}

// AutoCollector wraps a function to implement prometheus.Collector
type AutoCollector struct {
	collectFunc  func(ch chan<- prometheus.Metric)
	describeFunc func(ch chan<- *prometheus.Desc)
}

// NewAutoCollector creates a collector from functions
func NewAutoCollector(
	collectFunc func(ch chan<- prometheus.Metric),
	describeFunc func(ch chan<- *prometheus.Desc),
) *AutoCollector {
	return &AutoCollector{
		collectFunc:  collectFunc,
		describeFunc: describeFunc,
	}
}

// Collect implements prometheus.Collector
func (c *AutoCollector) Collect(ch chan<- prometheus.Metric) {
	if c.collectFunc != nil {
		c.collectFunc(ch)
	}
}

// Describe implements prometheus.Collector
func (c *AutoCollector) Describe(ch chan<- *prometheus.Desc) {
	if c.describeFunc != nil {
		c.describeFunc(ch)
	}
}

// Timer measures execution time and records it
type Timer struct {
	start time.Time
}

// NewTimer creates a new timer
func NewTimer() *Timer {
	return &Timer{start: time.Now()}
}

// ObserveDuration records the elapsed time
func (t *Timer) ObserveDuration() float64 {
	return time.Since(t.start).Seconds()
}

// LatencyRecorder records latency metrics
type LatencyRecorder struct {
	metric *prometheus.HistogramVec
}

// NewLatencyRecorder creates a new latency recorder
func NewLatencyRecorder(name, help string, buckets []float64) *LatencyRecorder {
	recorder := &LatencyRecorder{
		metric: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    name,
				Help:    help,
				Buckets: buckets,
			},
			[]string{"operation"},
		),
	}
	prometheus.MustRegister(recorder.metric)
	return recorder
}

// Record records latency for an operation
func (r *LatencyRecorder) Record(operation string, durationSeconds float64) {
	r.metric.WithLabelValues(operation).Observe(durationSeconds)
}

// CounterWrapper wraps a counter with labels
type CounterWrapper struct {
	counter *prometheus.CounterVec
}

// NewCounterWrapper creates a new labeled counter wrapper
func NewCounterWrapper(name, help string, labels []string) *CounterWrapper {
	wrapper := &CounterWrapper{
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: name,
				Help: help,
			},
			labels,
		),
	}
	prometheus.MustRegister(wrapper.counter)
	return wrapper
}

// Inc increments the counter for given label values
func (w *CounterWrapper) Inc(labelValues ...string) {
	w.counter.WithLabelValues(labelValues...).Inc()
}

// Add adds a value to the counter
func (w *CounterWrapper) Add(v float64, labelValues ...string) {
	w.counter.WithLabelValues(labelValues...).Add(v)
}

// GaugeWrapper wraps a gauge with labels
type GaugeWrapper struct {
	gauge *prometheus.GaugeVec
}

// NewGaugeWrapper creates a new labeled gauge wrapper
func NewGaugeWrapper(name, help string, labels []string) *GaugeWrapper {
	wrapper := &GaugeWrapper{
		gauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: name,
				Help: help,
			},
			labels,
		),
	}
	prometheus.MustRegister(wrapper.gauge)
	return wrapper
}

// Set sets the gauge value
func (w *GaugeWrapper) Set(v float64, labelValues ...string) {
	w.gauge.WithLabelValues(labelValues...).Set(v)
}

// Inc increments the gauge
func (w *GaugeWrapper) Inc(labelValues ...string) {
	w.gauge.WithLabelValues(labelValues...).Inc()
}

// Dec decrements the gauge
func (w *GaugeWrapper) Dec(labelValues ...string) {
	w.gauge.WithLabelValues(labelValues...).Dec()
}
