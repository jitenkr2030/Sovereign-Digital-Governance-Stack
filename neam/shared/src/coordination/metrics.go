package coordination

import (
	"sync"
	"sync/atomic"
	"time"

	"neam-platform/shared"
)

// MetricsCollector collects and manages lock contention metrics
type MetricsCollector struct {
	logger *shared.Logger

	// Per-resource type metrics
	mu             sync.RWMutex
	resourceMetrics map[string]*ResourceMetrics

	// Global counters
	totalAcquisitions   int64
	totalReleases       int64
	totalContentions    int64
	totalDeadlocks      int64
	totalAcquisitionTime int64
	totalHoldTime       int64
}

// ResourceMetrics contains metrics for a specific resource type
type ResourceMetrics struct {
	ResourceType string

	// Acquisition metrics
	acquisitions      int64
	acquisitionFails  int64
	acquisitionTime   int64
	contentionCount   int64

	// Hold time metrics
	holdDuration      int64
	holdCount         int64
	maxHoldDuration   int64

	// Retry metrics
	retryAttempts     int64
	maxRetryAttempts  int64

	// Contention hotspots
	mu                sync.RWMutex
	contentiousResources map[string]*ContentiousResource
}

// ContentiousResource tracks metrics for a specific resource that experiences contention
type ContentiousResource struct {
	ResourceID        string
	ContentionCount   int64
	TotalWaitTime     time.Duration
	MaxWaitTime       time.Duration
	LastContention    time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *shared.Logger) *MetricsCollector {
	return &MetricsCollector{
		logger:           logger,
		resourceMetrics:  make(map[string]*ResourceMetrics),
	}
}

// RecordAcquisition records a lock acquisition attempt
func (mc *MetricsCollector) RecordAcquisition(resourceType string, waitTime time.Duration, success bool) {
	atomic.AddInt64(&mc.totalAcquisitions, 1)
	atomic.AddInt64(&mc.totalAcquisitionTime, int64(waitTime))

	metrics := mc.getOrCreateResourceMetrics(resourceType)
	atomic.AddInt64(&metrics.acquisitions, 1)

	if !success {
		atomic.AddInt64(&metrics.acquisitionFails, 1)
	}

	atomic.AddInt64(&metrics.acquisitionTime, int64(waitTime))
}

// RecordRelease records a lock release
func (mc *MetricsCollector) RecordRelease(resourceType string, holdTime time.Duration) {
	atomic.AddInt64(&mc.totalReleases, 1)
	atomic.AddInt64(&mc.totalHoldTime, int64(holdTime))

	metrics := mc.getOrCreateResourceMetrics(resourceType)
	atomic.AddInt64(&metrics.holdCount, 1)
	atomic.AddInt64(&metrics.holdDuration, int64(holdTime))

	// Update max hold duration
	for {
		currentMax := atomic.LoadInt64(&metrics.maxHoldDuration)
		newValue := int64(holdTime)
		if newValue <= currentMax {
			break
		}
		if atomic.CompareAndSwapInt64(&metrics.maxHoldDuration, currentMax, newValue) {
			break
		}
	}
}

// RecordContention records a lock contention event
func (mc *MetricsCollector) RecordContention(resourceType string) {
	atomic.AddInt64(&mc.totalContentions, 1)

	metrics := mc.getOrCreateResourceMetrics(resourceType)
	atomic.AddInt64(&metrics.contentionCount, 1)
}

// RecordRetry records a retry attempt
func (mc *MetricsCollector) RecordRetry(resourceType string, attemptNumber int) {
	metrics := mc.getOrCreateResourceMetrics(resourceType)
	atomic.AddInt64(&metrics.retryAttempts, 1)

	// Update max retry attempts
	for {
		currentMax := atomic.LoadInt64(&metrics.maxRetryAttempts)
		if int64(attemptNumber) <= currentMax {
			break
		}
		if atomic.CompareAndSwapInt64(&metrics.maxRetryAttempts, int64(attemptNumber), currentMax) {
			break
		}
	}
}

// RecordDeadlock records a deadlock detection event
func (mc *MetricsCollector) RecordDeadlock(resourceType string) {
	atomic.AddInt64(&mc.totalDeadlocks, 1)
}

// RecordContentiousResource records contention on a specific resource
func (mc *MetricsCollector) RecordContentiousResource(resourceType, resourceID string, waitTime time.Duration) {
	metrics := mc.getOrCreateResourceMetrics(resourceType)

	metrics.mu.Lock()
	contentious, exists := metrics.contentiousResources[resourceID]
	if !exists {
		contentious = &ContentiousResource{
			ResourceID:     resourceID,
			ContentionCount: 0,
			TotalWaitTime:  0,
			MaxWaitTime:    0,
			LastContention: time.Now(),
		}
		metrics.contentiousResources[resourceID] = contentious
	}

	contentious.ContentionCount++
	contentious.TotalWaitTime += waitTime
	if waitTime > contentious.MaxWaitTime {
		contentious.MaxWaitTime = waitTime
	}
	contentious.LastContention = time.Now()
	metrics.mu.Unlock()
}

// RecordExtendedHold records an extended lock hold duration
func (mc *MetricsCollector) RecordExtendedHold(resourceType string, duration time.Duration, threshold time.Duration) {
	mc.logger.Warn("Extended lock hold detected",
		"resource_type", resourceType,
		"hold_duration", duration,
		"threshold", threshold,
	)
}

// GetGlobalMetrics returns global metrics
func (mc *MetricsCollector) GetGlobalMetrics() GlobalMetrics {
	acquisitions := atomic.LoadInt64(&mc.totalAcquisitions)
	return GlobalMetrics{
		Timestamp:          time.Now(),
		TotalAcquisitions:  acquisitions,
		TotalReleases:      atomic.LoadInt64(&mc.totalReleases),
		TotalContentions:   atomic.LoadInt64(&mc.totalContentions),
		TotalDeadlocks:     atomic.LoadInt64(&mc.totalDeadlocks),
		AvgAcquisitionTime: time.Duration(mc.totalAcquisitionTime) / time.Duration(max(acquisitions, 1)),
		AvgHoldTime:        time.Duration(mc.totalHoldTime) / time.Duration(max(atomic.LoadInt64(&mc.totalReleases), 1)),
	}
}

// GetResourceMetrics returns metrics for a specific resource type
func (mc *MetricsCollector) GetResourceMetrics(resourceType string) *ResourceMetrics {
	return mc.getOrCreateResourceMetrics(resourceType)
}

// GetAllResourceMetrics returns all resource type metrics
func (mc *MetricsCollector) GetAllResourceMetrics() []*ResourceMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics := make([]*ResourceMetrics, 0, len(mc.resourceMetrics))
	for _, m := range mc.resourceMetrics {
		metrics = append(metrics, m)
	}
	return metrics
}

// GetTopContentiousResources returns the top N most contentious resources
func (mc *MetricsCollector) GetTopContentiousResources(resourceType string, n int) []*ContentiousResource {
	metrics := mc.getOrCreateResourceMetrics(resourceType)

	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	contentious := make([]*ContentiousResource, 0, len(metrics.contentiousResources))
	for _, c := range metrics.contentiousResources {
		contentious = append(contentious, c)
	}

	// Sort by contention count descending
	for i := 0; i < len(contentious); i++ {
		for j := i + 1; j < len(contentious); j++ {
			if contentious[j].ContentionCount > contentious[i].ContentionCount {
				contentious[i], contentious[j] = contentious[j], contentious[i]
			}
		}
	}

	if n > 0 && n < len(contentious) {
		return contentious[:n]
	}
	return contentious
}

// GetContentionScore calculates a contention score for a resource type
// Higher score indicates more contention
func (mc *MetricsCollector) GetContentionScore(resourceType string) float64 {
	metrics := mc.getOrCreateResourceMetrics(resourceType)

	contentionCount := atomic.LoadInt64(&metrics.contentionCount)
	acquisitions := atomic.LoadInt64(&metrics.acquisitions)

	if acquisitions == 0 {
		return 0
	}

	// Contention score = contention rate * log(total acquisitions)
	contentionRate := float64(contentionCount) / float64(acquisitions)
	volumeFactor := 1.0 + float64(acquisitions)/1000.0

	return contentionRate * volumeFactor
}

// Reset resets all metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.resourceMetrics = make(map[string]*ResourceMetrics)
	atomic.StoreInt64(&mc.totalAcquisitions, 0)
	atomic.StoreInt64(&mc.totalReleases, 0)
	atomic.StoreInt64(&mc.totalContentions, 0)
	atomic.StoreInt64(&mc.totalDeadlocks, 0)
	mc.totalAcquisitionTime = 0
	mc.totalHoldTime = 0
}

// getOrCreateResourceMetrics gets or creates metrics for a resource type
func (mc *MetricsCollector) getOrCreateResourceMetrics(resourceType string) *ResourceMetrics {
	mc.mu.RLock()
	metrics, exists := mc.resourceMetrics[resourceType]
	mc.mu.RUnlock()

	if exists {
		return metrics
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Double-check after acquiring write lock
	if metrics, exists = mc.resourceMetrics[resourceType]; exists {
		return metrics
	}

	metrics = &ResourceMetrics{
		ResourceType:       resourceType,
		contentiousResources: make(map[string]*ContentiousResource),
	}
	mc.resourceMetrics[resourceType] = metrics
	return metrics
}

// GlobalMetrics contains global lock metrics
type GlobalMetrics struct {
	Timestamp          time.Time       `json:"timestamp"`
	TotalAcquisitions  int64           `json:"total_acquisitions"`
	TotalReleases      int64           `json:"total_releases"`
	TotalContentions   int64           `json:"total_contentions"`
	TotalDeadlocks     int64           `json:"total_deadlocks"`
	AvgAcquisitionTime time.Duration   `json:"avg_acquisition_time"`
	AvgHoldTime        time.Duration   `json:"avg_hold_time"`
}

// ResourceMetricsSummary provides a summary of resource metrics
type ResourceMetricsSummary struct {
	ResourceType       string         `json:"resource_type"`
	TotalLocks         int64          `json:"total_locks"`
	Contentions        int64          `json:"contentions"`
	AvgHoldTime        time.Duration  `json:"avg_hold_time"`
	MaxHoldTime        time.Duration  `json:"max_hold_time"`
	ContentionScore    float64        `json:"contention_score"`
	TopContentious     []string       `json:"top_contentious_resources"`
}

// GetSummary returns a summary of all metrics
func (mc *MetricsCollector) GetSummary() []*ResourceMetricsSummary {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	summaries := make([]*ResourceMetricsSummary, 0, len(mc.resourceMetrics))
	for _, metrics := range mc.resourceMetrics {
		summary := &ResourceMetricsSummary{
			ResourceType:    metrics.ResourceType,
			TotalLocks:      atomic.LoadInt64(&metrics.acquisitions),
			Contentions:     atomic.LoadInt64(&metrics.contentionCount),
			AvgHoldTime:     time.Duration(atomic.LoadInt64(&metrics.holdDuration)) / time.Duration(max(atomic.LoadInt64(&metrics.holdCount), 1)),
			MaxHoldTime:     time.Duration(atomic.LoadInt64(&metrics.maxHoldDuration)),
			ContentionScore: mc.GetContentionScore(metrics.ResourceType),
		}

		// Get top 5 contentious resources
		metrics.mu.RLock()
		topContentious := make([]string, 0, 5)
		for id := range metrics.contentiousResources {
			topContentious = append(topContentious, id)
			if len(topContentious) >= 5 {
				break
			}
		}
		metrics.mu.RUnlock()
		summary.TopContentious = topContentious

		summaries = append(summaries, summary)
	}

	return summaries
}

// Helper function for max
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// ContentionAnalyzer analyzes contention patterns
type ContentionAnalyzer struct {
	collector *MetricsCollector
	threshold float64
	window    time.Duration
}

// NewContentionAnalyzer creates a new contention analyzer
func NewContentionAnalyzer(collector *MetricsCollector, threshold float64, window time.Duration) *ContentionAnalyzer {
	return &ContentionAnalyzer{
		collector: collector,
		threshold: threshold,
		window:    window,
	}
}

// AnalyzeContentiousResources returns resources that exceed the contention threshold
func (ca *ContentionAnalyzer) AnalyzeContentiousResources() []*ContentionAlert {
	var alerts []*ContentionAlert

	metrics := ca.collector.GetAllResourceMetrics()
	now := time.Now()

	for _, m := range metrics {
		score := ca.collector.GetContentionScore(m.ResourceType)
		if score > ca.threshold {
			alert := &ContentionAlert{
				ResourceType:     m.ResourceType,
				ContentionScore:  score,
				Threshold:        ca.threshold,
				ContentionCount:  atomic.LoadInt64(&m.contentionCount),
				AvgWaitTime:      time.Duration(atomic.LoadInt64(&m.acquisitionTime)) / time.Duration(max(atomic.LoadInt64(&m.acquisitions), 1)),
				DetectedAt:       now,
				Severity:         ca.determineSeverity(score),
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// determineSeverity determines the alert severity based on contention score
func (ca *ContentionAnalyzer) determineSeverity(score float64) AlertSeverity {
	if score > ca.threshold*3 {
		return SeverityCritical
	} else if score > ca.threshold*2 {
		return SeverityHigh
	} else if score > ca.threshold*1.5 {
		return SeverityMedium
	}
	return SeverityLow
}

// ContentionAlert represents an alert for high contention
type ContentionAlert struct {
	ResourceType    string         `json:"resource_type"`
	ContentionScore float64        `json:"contention_score"`
	Threshold       float64        `json:"threshold"`
	ContentionCount int64          `json:"contention_count"`
	AvgWaitTime     time.Duration  `json:"avg_wait_time"`
	DetectedAt      time.Time      `json:"detected_at"`
	Severity        AlertSeverity  `json:"severity"`
}

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	SeverityLow      AlertSeverity = "low"
	SeverityMedium   AlertSeverity = "medium"
	SeverityHigh     AlertSeverity = "high"
	SeverityCritical AlertSeverity = "critical"
)
