package metrics

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
)

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeCounter   MetricType = "counter"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Metric represents a single metric data point
type Metric struct {
	Name       string            `json:"name"`
	Type       MetricType        `json:"type"`
	Value      float64           `json:"value"`
	Labels     map[string]string `json:"labels,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
	Provider   string            `json:"provider"`
}

// MetricsSnapshot represents a snapshot of all collected metrics
type MetricsSnapshot struct {
	Timestamp     time.Time              `json:"timestamp"`
	CPUUsage      float64                `json:"cpu_usage_percent"`
	MemoryUsage   float64                `json:"memory_usage_percent"`
	MemoryUsed    uint64                 `json:"memory_used_bytes"`
	MemoryTotal   uint64                 `json:"memory_total_bytes"`
	Goroutines    int                    `json:"goroutine_count"`
	GCStats       GCStats                `json:"gc_stats"`
	CustomMetrics []Metric               `json:"custom_metrics"`
}

// GCStats represents garbage collection statistics
type GCStats struct {
	NumGC         uint32    `json:"num_gc"`
	PauseTotal    time.Duration `json:"pause_total"`
	LastPause     time.Duration `json:"last_pause"`
	Enable        bool      `json:"enabled"`
}

// CollectorConfig contains configuration for the metrics collector
type CollectorConfig struct {
	Enabled           bool          `json:"enabled"`
	CollectionInterval time.Duration `json:"collection_interval"`
	Detailed          bool          `json:"detailed"`
	Providers         []string      `json:"providers"`
	CustomMetrics     []CustomMetricConfig `json:"custom_metrics"`
}

// CustomMetricConfig defines a custom metric to collect
type CustomMetricConfig struct {
	Name      string            `json:"name"`
	Type      MetricType        `json:"type"`
	Labels    map[string]string `json:"labels"`
	Provider  string            `json:"provider"`
}

// CollectionSession represents an active metrics collection session
type CollectionSession struct {
	ID        string
	Category  string
	TargetID  string
	StartTime time.Time
	EndTime   time.Time
	Metrics   []Metric
}

// Service provides metrics collection capabilities
type Service struct {
	logger          *zap.Logger
	config          *CollectorConfig
	collector       *Collector
	snapshots       []MetricsSnapshot
	sessions        map[string]*CollectionSession
	customProviders map[string]CustomMetricProvider
	mu              sync.RWMutex
}

// CustomMetricProvider is an interface for custom metric providers
type CustomMetricProvider interface {
	GetMetrics() ([]Metric, error)
}

// Collector is the global metrics collector
type Collector struct {
	logger     *zap.Logger
	config     *CollectorConfig
	snapshots  []MetricsSnapshot
	mu         sync.RWMutex
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// Global collector instance
var globalCollector *Collector
var collectorMu sync.RWMutex

// Initialize initializes the metrics service
func Initialize(logger *zap.Logger, cfg interface{}) error {
	collectorMu.Lock()
	defer collectorMu.Unlock()

	// Extract configuration
	config := &CollectorConfig{
		Enabled:            true,
		CollectionInterval: 5 * time.Second,
		Detailed:           true,
		Providers:          []string{"cpu", "memory", "goroutine", "gc"},
	}

	// Try to extract from config
	if c, ok := cfg.(interface {
		Get(name string) interface{}
	}); ok {
		if v := c.Get("metrics"); v != nil {
			// Would parse detailed config here
			_ = v
		}
	}

	globalCollector = &Collector{
		logger:    logger,
		config:    config,
		snapshots: make([]MetricsSnapshot, 0),
		stopChan:  make(chan struct{}),
	}

	logger.Info("Metrics collector initialized",
		zap.Bool("enabled", config.Enabled),
		zap.Duration("interval", config.CollectionInterval))

	return nil
}

// GetCollector returns the global metrics collector
func GetCollector() *Collector {
	collectorMu.RLock()
	defer collectorMu.RUnlock()
	return globalCollector
}

// StartCollection starts collecting metrics for a specific target
func (c *Collector) StartCollection(ctx context.Context, category, targetID string) *CollectionSession {
	session := &CollectionSession{
		ID:       generateSessionID(),
		Category: category,
		TargetID: targetID,
		StartTime: time.Now(),
		Metrics:  make([]Metric, 0),
	}

	c.mu.Lock()
	c.sessions[session.ID] = session
	c.mu.Unlock()

	// Start background collection if enabled
	if c.config.Enabled {
		c.wg.Add(1)
		go c.collectMetrics(ctx, session)
	}

	return session
}

// StopCollection stops a metrics collection session
func (c *Collector) StopCollection(category, targetID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id, session := range c.sessions {
		if session.Category == category && session.TargetID == targetID {
			session.EndTime = time.Now()
			delete(c.sessions, id)
			break
		}
	}
}

// collectMetrics continuously collects metrics for a session
func (c *Collector) collectMetrics(ctx context.Context, session *CollectionSession) {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		case <-ticker.C:
			metrics, err := c.CollectAll()
			if err != nil {
				c.logger.Error("Failed to collect metrics", zap.Error(err))
				continue
			}

			c.mu.Lock()
			session.Metrics = append(session.Metrics, metrics...)
			c.snapshots = append(c.snapshots, c.takeSnapshot())
			c.mu.Unlock()
		}
	}
}

// CollectAll collects all available metrics
func (c *Collector) CollectAll() ([]Metric, error) {
	metrics := make([]Metric, 0)

	// Collect CPU metrics
	cpuMetrics, err := c.collectCPUMetrics()
	if err != nil {
		c.logger.Warn("Failed to collect CPU metrics", zap.Error(err))
	} else {
		metrics = append(metrics, cpuMetrics...)
	}

	// Collect memory metrics
	memMetrics, err := c.collectMemoryMetrics()
	if err != nil {
		c.logger.Warn("Failed to collect memory metrics", zap.Error(err))
	} else {
		metrics = append(metrics, memMetrics...)
	}

	// Collect runtime metrics
	runtimeMetrics, err := c.collectRuntimeMetrics()
	if err != nil {
		c.logger.Warn("Failed to collect runtime metrics", zap.Error(err))
	} else {
		metrics = append(metrics, runtimeMetrics...)
	}

	// Collect custom metrics
	customMetrics, err := c.collectCustomMetrics()
	if err != nil {
		c.logger.Warn("Failed to collect custom metrics", zap.Error(err))
	} else {
		metrics = append(metrics, customMetrics...)
	}

	return metrics, nil
}

// collectCPUMetrics collects CPU-related metrics
func (c *Collector) collectCPUMetrics() ([]Metric, error) {
	metrics := make([]Metric, 0)
	timestamp := time.Now()

	// CPU usage percentage
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}

	var cpuPercent float64
	if len(percent) > 0 {
		cpuPercent = percent[0]
	}

	metrics = append(metrics, Metric{
		Name:      "cpu_usage_percent",
		Type:      MetricTypeGauge,
		Value:     cpuPercent,
		Timestamp: timestamp,
		Provider:  "gopsutil/cpu",
	})

	// CPU count
	count, err := cpu.Counts(true)
	if err == nil {
		metrics = append(metrics, Metric{
			Name:      "cpu_count",
			Type:      MetricTypeGauge,
			Value:     float64(count),
			Timestamp: timestamp,
			Provider:  "gopsutil/cpu",
		})
	}

	return metrics, nil
}

// collectMemoryMetrics collects memory-related metrics
func (c *Collector) collectMemoryMetrics() ([]Metric, error) {
	metrics := make([]Metric, 0)
	timestamp := time.Now()

	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	metrics = append(metrics, Metric{
		Name:      "memory_used_bytes",
		Type:      MetricTypeGauge,
		Value:     float64(v.Used),
		Timestamp: timestamp,
		Provider:  "gopsutil/mem",
	})

	metrics = append(metrics, Metric{
		Name:      "memory_total_bytes",
		Type:      MetricTypeGauge,
		Value:     float64(v.Total),
		Timestamp: timestamp,
		Provider:  "gopsutil/mem",
	})

	metrics = append(metrics, Metric{
		Name:      "memory_usage_percent",
		Type:      MetricTypeGauge,
		Value:     v.UsedPercent,
		Timestamp: timestamp,
		Provider:  "gopsutil/mem",
	})

	metrics = append(metrics, Metric{
		Name:      "memory_available_bytes",
		Type:      MetricTypeGauge,
		Value:     float64(v.Available),
		Timestamp: timestamp,
		Provider:  "gopsutil/mem",
	})

	return metrics, nil
}

// collectRuntimeMetrics collects Go runtime metrics
func (c *Collector) collectRuntimeMetrics() ([]Metric, error) {
	metrics := make([]Metric, 0)
	timestamp := time.Now()

	// Goroutine count
	metrics = append(metrics, Metric{
		Name:      "goroutine_count",
		Type:      MetricTypeGauge,
		Value:     float64(runtime.NumGoroutine()),
		Timestamp: timestamp,
		Provider:  "runtime",
	})

	// GC stats
	stats := &runtime.MemStats{}
	runtime.ReadMemStats(stats)

	metrics = append(metrics, Metric{
		Name:      "gc_num_forced",
		Type:      MetricTypeCounter,
		Value:     float64(stats.NumForcedGC),
		Timestamp: timestamp,
		Provider:  "runtime/gc",
	})

	metrics = append(metrics, Metric{
		Name:      "gc_pause_total_ns",
		Type:      MetricTypeCounter,
		Value:     float64(stats.PauseTotalNs),
		Timestamp: timestamp,
		Provider:  "runtime/gc",
	})

	metrics = append(metrics, Metric{
		Name:      "heap_alloc_bytes",
		Type:      MetricTypeGauge,
		Value:     float64(stats.HeapAlloc),
		Timestamp: timestamp,
		Provider:  "runtime/memory",
	})

	metrics = append(metrics, Metric{
		Name:      "heap_sys_bytes",
		Type:      MetricTypeGauge,
		Value:     float64(stats.HeapSys),
		Timestamp: timestamp,
		Provider:  "runtime/memory",
	})

	metrics = append(metrics, Metric{
		Name:      "heap_idle_bytes",
		Type:      MetricTypeGauge,
		Value:     float64(stats.HeapIdle),
		Timestamp: timestamp,
		Provider:  "runtime/memory",
	})

	metrics = append(metrics, Metric{
		Name:      "heap_inuse_bytes",
		Type:      MetricTypeGauge,
		Value:     float64(stats.HeapInuse),
		Timestamp: timestamp,
		Provider:  "runtime/memory",
	})

	return metrics, nil
}

// collectCustomMetrics collects metrics from custom providers
func (c *Collector) collectCustomMetrics() ([]Metric, error) {
	metrics := make([]Metric, 0)

	c.mu.RLock()
	providers := c.customProviders
	c.mu.RUnlock()

	for name, provider := range providers {
		providerMetrics, err := provider.GetMetrics()
		if err != nil {
			c.logger.Warn("Custom provider failed",
				zap.String("provider", name),
				zap.Error(err))
			continue
		}
		metrics = append(metrics, providerMetrics...)
	}

	return metrics, nil
}

// takeSnapshot takes a snapshot of current system metrics
func (c *Collector) takeSnapshot() MetricsSnapshot {
	stats := &runtime.MemStats{}
	runtime.ReadMemStats(stats)

	cpuPercent, _ := cpu.Percent(0, false)
	var cpu float64
	if len(cpuPercent) > 0 {
		cpu = cpuPercent[0]
	}

	v, _ := mem.VirtualMemory()

	return MetricsSnapshot{
		Timestamp:    time.Now(),
		CPUUsage:     cpu,
		MemoryUsage:  v.UsedPercent,
		MemoryUsed:   v.Used,
		MemoryTotal:  v.Total,
		Goroutines:   runtime.NumGoroutine(),
		GCStats: GCStats{
			NumGC:      stats.NumGC,
			PauseTotal: time.Duration(stats.PauseTotalNs),
			LastPause:  time.Duration(stats.PauseNs[(stats.NumGC+255)%256]),
			Enable:     true,
		},
	}
}

// GetSnapshot returns the latest metrics snapshot
func (c *Collector) GetSnapshot() MetricsSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.snapshots) == 0 {
		return c.takeSnapshot()
	}
	return c.snapshots[len(c.snapshots)-1]
}

// GetHistory returns historical snapshots
func (c *Collector) GetHistory(limit int) []MetricsSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if limit <= 0 || limit > len(c.snapshots) {
		limit = len(c.snapshots)
	}

	snapshots := make([]MetricsSnapshot, limit)
	copy(snapshots, c.snapshots[len(c.snapshots)-limit:])
	return snapshots
}

// GetMetric returns a specific metric by name
func (c *Collector) GetMetric(name string) *Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics, _ := c.CollectAll()
	for _, m := range metrics {
		if m.Name == name {
			return &m
		}
	}
	return nil
}

// RegisterCustomProvider registers a custom metric provider
func (c *Collector) RegisterCustomProvider(name string, provider CustomMetricProvider) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.customProviders == nil {
		c.customProviders = make(map[string]CustomMetricProvider)
	}
	c.customProviders[name] = provider
}

// UnregisterCustomProvider removes a custom metric provider
func (c *Collector) UnregisterCustomProvider(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.customProviders, name)
}

// Stop stops the metrics collector
func (c *Collector) Stop() {
	close(c.stopChan)
	c.wg.Wait()
}

// GetService returns a new metrics service instance
func GetService(logger *zap.Logger, config *CollectorConfig) *Service {
	return &Service{
		logger:  logger,
		config:  config,
		metrics: make([]Metric, 0),
	}
}

func generateSessionID() string {
	return string(time.Now().Format("20060102150405")) + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
