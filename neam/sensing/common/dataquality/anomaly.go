package dataquality

import (
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AnomalyType represents the type of anomaly detected
type AnomalyType string

const (
	AnomalyTypeScoreDrop     AnomalyType = "score_drop"
	AnomalyTypeSpike         AnomalyType = "spike"
	AnomalyTypeSilence       AnomalyType = "silence"
	AnomalyTypeOutlier       AnomalyType = "outlier"
	AnomalyTypeTrendChange   AnomalyType = "trend_change"
	AnomalyTypeSeasonality   AnomalyType = "seasonality_violation"
)

// AnomalySeverity represents the severity of an anomaly
type AnomalySeverity string

const (
	AnomalySeverityLow      AnomalySeverity = "low"
	AnomalySeverityMedium   AnomalySeverity = "medium"
	AnomalySeverityHigh     AnomalySeverity = "high"
	AnomalySeverityCritical AnomalySeverity = "critical"
)

// Anomaly represents a detected data quality anomaly
type Anomaly struct {
	AnomalyID      string            `json:"anomaly_id"`
	Timestamp      time.Time         `json:"timestamp"`
	SourceID       string            `json:"source_id"`
	Type           AnomalyType       `json:"type"`
	Severity       AnomalySeverity   `json:"severity"`
	MetricName     string            `json:"metric_name"`
	CurrentValue   float64           `json:"current_value"`
	ExpectedValue  float64           `json:"expected_value"`
	Deviation      float64           `json:"deviation"`
	Description    string            `json:"description"`
	Context        map[string]interface{} `json:"context,omitempty"`
	Recommendations []string         `json:"recommendations,omitempty"`
}

// AnomalyConfig contains configuration for anomaly detection
type AnomalyConfig struct {
	// Statistical thresholds
	ZScoreThreshold    float64 `json:"z_score_threshold"`
	StdDevThreshold    float64 `json:"std_dev_threshold"`
	PercentileThreshold float64 `json:"percentile_threshold"`

	// Window settings
	WindowSize         int           `json:"window_size"`
	MinWindowSize      int           `json:"min_window_size"`
	SlideInterval      time.Duration `json:"slide_interval"`

	// Detection settings
	SilenceDuration    time.Duration `json:"silence_duration"`
	DropThreshold      float64       `json:"drop_threshold"`
	SpikeMultiplier    float64       `json:"spike_multiplier"`

	// Alert settings
	AlertCooldown      time.Duration `json:"alert_cooldown"`
	EnableAlerts       bool          `json:"enable_alerts"`
}

// DefaultAnomalyConfig returns the default anomaly detection configuration
func DefaultAnomalyConfig() AnomalyConfig {
	return AnomalyConfig{
		ZScoreThreshold:    3.0,
		StdDevThreshold:    2.0,
		PercentileThreshold: 0.95,
		WindowSize:         100,
		MinWindowSize:      10,
		SlideInterval:      time.Minute,
		SilenceDuration:    5 * time.Minute,
		DropThreshold:      0.2,
		SpikeMultiplier:    3.0,
		AlertCooldown:      5 * time.Minute,
		EnableAlerts:       true,
	}
}

// TimeSeriesPoint represents a single data point in a time series
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// TimeSeries maintains a sliding window of time series data
type TimeSeries struct {
	mu        sync.RWMutex
	Points    []TimeSeriesPoint
	MaxSize   int
}

// NewTimeSeries creates a new time series with specified max size
func NewTimeSeries(maxSize int) *TimeSeries {
	return &TimeSeries{
		Points:  make([]TimeSeriesPoint, 0, maxSize),
		MaxSize: maxSize,
	}
}

// Add adds a new point to the time series
func (ts *TimeSeries) Add(point TimeSeriesPoint) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.Points = append(ts.Points, point)

	// Maintain max size
	if len(ts.Points) > ts.MaxSize {
		ts.Points = ts.Points[len(ts.Points)-ts.MaxSize:]
	}
}

// GetRecent returns the most recent points
func (ts *TimeSeries) GetRecent(count int) []TimeSeriesPoint {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	if len(ts.Points) == 0 {
		return nil
	}

	start := len(ts.Points) - count
	if start < 0 {
		start = 0
	}

	result := make([]TimeSeriesPoint, count)
	copy(result, ts.Points[start:])
	return result
}

// GetValues returns all values in the time series
func (ts *TimeSeries) GetValues() []float64 {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	values := make([]float64, len(ts.Points))
	for i, p := range ts.Points {
		values[i] = p.Value
	}
	return values
}

// Statistics holds calculated statistics
type Statistics struct {
	Mean        float64
	StdDev      float64
	Min         float64
	Max         float64
	Median      float64
	Percentile95 float64
	Count       int
}

// Calculate calculates statistics for the time series
func (ts *TimeSeries) Calculate() Statistics {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	stats := Statistics{
		Min:  math.MaxFloat64,
		Max:  -math.MaxFloat64,
		Count: len(ts.Points),
	}

	if len(ts.Points) == 0 {
		return stats
	}

	var sum float64
	values := make([]float64, len(ts.Points))

	for i, p := range ts.Points {
		values[i] = p.Value
		sum += p.Value

		if p.Value < stats.Min {
			stats.Mean = p.Value
		}
		if p.Value > stats.Max {
			stats.Max = p.Value
		}
	}

	stats.Mean = sum / float64(len(values))

	// Calculate standard deviation
	var sumSq float64
	for _, v := range values {
		diff := v - stats.Mean
		sumSq += diff * diff
	}
	stats.StdDev = math.Sqrt(sumSq / float64(len(values)))

	// Calculate median and percentile
	sorted := make([]float64, len(values))
	copy(sorted, values)
	insertionSort(sorted)

	stats.Median = percentile(sorted, 0.5)
	stats.Percentile95 = percentile(sorted, 0.95)
	stats.Min = sorted[0]
	stats.Max = sorted[len(sorted)-1]

	return stats
}

// insertionSort sorts a slice of floats (simple implementation)
func insertionSort(arr []float64) {
	for i := 1; i < len(arr); i++ {
		key := arr[i]
		j := i - 1
		for j >= 0 && arr[j] > key {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = key
	}
}

// percentile calculates the specified percentile
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}

	index := p * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	return sorted[lower]*(1-index+float64(lower)) + sorted[upper]*(index-float64(lower))
}

// Detector performs anomaly detection on data quality metrics
type Detector struct {
	mu          sync.RWMutex
	config      AnomalyConfig
	timeSeries  map[string]*TimeSeries
	lastAnomaly map[string]time.Time
	alertFunc   func(*Anomaly)
	logger      *zap.Logger
}

// NewDetector creates a new anomaly detector
func NewDetector(config AnomalyConfig, logger *zap.Logger) *Detector {
	return &Detector{
		config:     config,
		timeSeries: make(map[string]*TimeSeries),
		lastAnomaly: make(map[string]time.Time),
		logger:     logger,
	}
}

// SetAlertFunc sets the function to call when an anomaly is detected
func (d *Detector) SetAlertFunc(alertFunc func(*Anomaly)) {
	d.alertFunc = alertFunc
}

// Observe adds an observation and checks for anomalies
func (d *Detector) Observe(sourceID string, metricName string, value float64) *Anomaly {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := fmt.Sprintf("%s:%s", sourceID, metricName)

	// Get or create time series
	ts, exists := d.timeSeries[key]
	if !exists {
		ts = NewTimeSeries(d.config.WindowSize)
		d.timeSeries[key] = ts
	}

	// Add the observation
	ts.Add(TimeSeriesPoint{
		Timestamp: time.Now(),
		Value:     value,
	})

	// Check if we have enough data
	if len(ts.Points) < d.config.MinWindowSize {
		return nil
	}

	// Perform anomaly detection
	return d.detectAnomaly(key, sourceID, metricName, ts, value)
}

// detectAnomaly performs the actual anomaly detection
func (d *Detector) detectAnomaly(key, sourceID, metricName string, ts *TimeSeries, currentValue float64) *Anomaly {
	stats := ts.Calculate()

	// Check for score drop
	if currentValue < stats.Mean*(1-d.config.DropThreshold) {
		return d.createAnomaly(sourceID, metricName, AnomalyTypeScoreDrop, currentValue, stats.Mean, stats.StdDev, stats)
	}

	// Check for spike
	if currentValue > stats.Mean*(1+d.config.SpikeMultiplier*stats.StdDev/stats.Mean) {
		return d.createAnomaly(sourceID, metricName, AnomalyTypeSpike, currentValue, stats.Mean, stats.StdDev, stats)
	}

	// Check Z-score
	if stats.StdDev > 0 {
		zScore := (currentValue - stats.Mean) / stats.StdDev
		if math.Abs(zScore) > d.config.ZScoreThreshold {
			return d.createAnomaly(sourceID, metricName, AnomalyTypeOutlier, currentValue, stats.Mean, stats.StdDev, stats)
		}
	}

	return nil
}

// createAnomaly creates an anomaly record
func (d *Detector) createAnomaly(sourceID, metricName string, anomalyType AnomalyType, currentValue, expectedValue, stdDev float64, stats Statistics) *Anomaly {
	deviation := 0.0
	if expectedValue != 0 {
		deviation = (currentValue - expectedValue) / expectedValue * 100
	}

	severity := d.calculateSeverity(anomalyType, deviation)

	anomaly := &Anomaly{
		AnomalyID:     fmt.Sprintf("anomaly_%d", time.Now().UnixNano()),
		Timestamp:     time.Now(),
		SourceID:      sourceID,
		Type:          anomalyType,
		Severity:      severity,
		MetricName:    metricName,
		CurrentValue:  currentValue,
		ExpectedValue: expectedValue,
		Deviation:     deviation,
		Description:   d.generateDescription(anomalyType, sourceID, metricName, currentValue, expectedValue, deviation),
		Context: map[string]interface{}{
			"std_dev":      stdDev,
			"mean":         stats.Mean,
			"min":          stats.Min,
			"max":          stats.Max,
			"percentile_95": stats.Percentile95,
		},
		Recommendations: d.generateRecommendations(anomalyType, sourceID, metricName),
	}

	// Check cooldown
	if lastTime, exists := d.lastAnomaly[key]; exists {
		if time.Since(lastTime) < d.config.AlertCooldown {
			return nil // Skip due to cooldown
		}
	}

	d.lastAnomaly[key] = time.Now()

	// Call alert function if set
	if d.config.EnableAlerts && d.alertFunc != nil {
		d.alertFunc(anomaly)
	}

	d.logger.Warn("Anomaly detected",
		zap.String("anomaly_id", anomaly.AnomalyID),
		zap.String("source_id", sourceID),
		zap.String("type", string(anomalyType)),
		zap.String("severity", string(severity)),
		zap.String("metric", metricName),
		zap.Float64("current_value", currentValue),
		zap.Float64("expected_value", expectedValue),
		zap.Float64("deviation", deviation))

	return anomaly
}

// calculateSeverity calculates the severity of an anomaly
func (d *Detector) calculateSeverity(anomalyType AnomalyType, deviation float64) AnomalySeverity {
	absDeviation := math.Abs(deviation)

	switch {
	case absDeviation > 50:
		return AnomalySeverityCritical
	case absDeviation > 30:
		return AnomalySeverityHigh
	case absDeviation > 20:
		return AnomalySeverityMedium
	default:
		return AnomalySeverityLow
	}
}

// generateDescription generates a human-readable description
func (d *Detector) generateDescription(anomalyType AnomalyType, sourceID, metricName string, currentValue, expectedValue, deviation float64) string {
	switch anomalyType {
	case AnomalyTypeScoreDrop:
		return fmt.Sprintf("Quality score dropped significantly for %s:%s. Current: %.2f, Expected: %.2f (%.1f%% deviation)",
			sourceID, metricName, currentValue, expectedValue, deviation)
	case AnomalyTypeSpike:
		return fmt.Sprintf("Unexpected spike detected in %s:%s. Current: %.2f, Expected: %.2f (%.1f%% deviation)",
			sourceID, metricName, currentValue, expectedValue, deviation)
	case AnomalyTypeOutlier:
		return fmt.Sprintf("Outlier detected in %s:%s. Value: %.2f is outside expected range (%.2f ± %.2f)",
			sourceID, metricName, currentValue, expectedValue, expectedValue)
	case AnomalyTypeSilence:
		return fmt.Sprintf("No data received from %s:%s for %v",
			sourceID, metricName, d.config.SilenceDuration)
	case AnomalyTypeTrendChange:
		return fmt.Sprintf("Trend change detected in %s:%s",
			sourceID, metricName)
	default:
		return fmt.Sprintf("Anomaly detected in %s:%s", sourceID, metricName)
	}
}

// generateRecommendations generates recommended actions
func (d *Detector) generateRecommendations(anomalyType AnomalyType, sourceID, metricName string) []string {
	switch anomalyType {
	case AnomalyTypeScoreDrop:
		return []string{
			"Check data source connectivity",
			"Review recent configuration changes",
			"Verify upstream data quality",
			"Consider running data validation",
		}
	case AnomalyTypeSpike:
		return []string{
			"Verify data source accuracy",
			"Check for data source malfunction",
			"Review recent events that may have caused the spike",
			"Consider filtering extreme values",
		}
	case AnomalyTypeOutlier:
		return []string{
			"Investigate the root cause of the outlier",
			"Verify data integrity",
			"Check for sensor or data source issues",
			"Consider adjusting thresholds",
		}
	case AnomalyTypeSilence:
		return []string{
			"Check data source connectivity",
			"Verify data source is operational",
			"Review network connectivity",
			"Check for scheduled maintenance",
		}
	default:
		return []string{
			"Investigate the anomaly",
			"Review recent changes",
			"Check system logs",
		}
	}
}

// CheckSilence checks for silence anomalies (no data received)
func (d *Detector) CheckSilence(sourceID string, lastUpdate time.Time) *Anomaly {
	d.mu.Lock()
	defer d.mu.Unlock()

	if time.Since(lastUpdate) > d.config.SilenceDuration {
		key := fmt.Sprintf("%s:silence", sourceID)

		// Check cooldown
		if lastTime, exists := d.lastAnomaly[key]; exists {
			if time.Since(lastTime) < d.config.AlertCooldown {
				return nil
			}
		}

		d.lastAnomaly[key] = time.Now()

		return &Anomaly{
			AnomalyID:     fmt.Sprintf("anomaly_%d", time.Now().UnixNano()),
			Timestamp:     time.Now(),
			SourceID:      sourceID,
			Type:          AnomalyTypeSilence,
			Severity:      AnomalySeverityHigh,
			MetricName:    "data_flow",
			CurrentValue:  0,
			ExpectedValue: 1,
			Deviation:     -100,
			Description:   fmt.Sprintf("No data received from %s for %v", sourceID, d.config.SilenceDuration),
			Recommendations: []string{
				"Check data source connectivity",
				"Verify data source is operational",
				"Review network connectivity",
			},
		}
	}

	return nil
}

// GetStatistics returns statistics for a metric
func (d *Detector) GetStatistics(sourceID, metricName string) Statistics {
	d.mu.RLock()
	defer d.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", sourceID, metricName)
	if ts, exists := d.timeSeries[key]; exists {
		return ts.Calculate()
	}

	return Statistics{}
}

// GetTimeSeries returns the time series for a metric
func (d *Detector) GetTimeSeries(sourceID, metricName string) []TimeSeriesPoint {
	d.mu.RLock()
	defer d.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", sourceID, metricName)
	if ts, exists := d.timeSeries[key]; exists {
		ts.mu.RLock()
		defer ts.mu.RUnlock()
		points := make([]TimeSeriesPoint, len(ts.Points))
		copy(points, ts.Points)
		return points
	}

	return nil
}

// Reset clears all time series data
func (d *Detector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.timeSeries = make(map[string]*TimeSeries)
	d.lastAnomaly = make(map[string]time.Time)
}

// ResetSource clears data for a specific source
func (d *Detector) ResetSource(sourceID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for key := range d.timeSeries {
		if len(key) > len(sourceID) && key[:len(sourceID)] == sourceID {
			delete(d.timeSeries, key)
		}
	}

	for key := range d.lastAnomaly {
		if len(key) > len(sourceID) && key[:len(sourceID)] == sourceID {
			delete(d.lastAnomaly, key)
		}
	}
}
