package anomaly

import (
	"math"
	"sync"
	"time"

	"neam-platform/intelligence/models"
)

// Detector is the interface for all anomaly detection algorithms
type Detector interface {
	// Detect analyzes a new data point against historical data
	// Returns: isAnomaly, anomalyScore, confidence
	Detect(history []models.TimeSeriesPoint, current float64) (bool, float64, float64)
	
	// GetName returns the name of the detector algorithm
	GetName() string
	
	// Configure sets detector-specific parameters
	Configure(params map[string]interface{}) error
}

// ZScoreDetector implements statistical anomaly detection using Z-score
type ZScoreDetector struct {
	Threshold   float64
	Lookback    int
	MinDataPoints int
	mu          sync.RWMutex
}

// NewZScoreDetector creates a new Z-score detector
func NewZScoreDetector(threshold float64, lookback int) *ZScoreDetector {
	return &ZScoreDetector{
		Threshold:     threshold,
		Lookback:      lookback,
		MinDataPoints: 5,
	}
}

// GetName returns the detector name
func (z *ZScoreDetector) GetName() string {
	return "Z_SCORE"
}

// Configure sets detector parameters
func (z *ZScoreDetector) Configure(params map[string]interface{}) error {
	z.mu.Lock()
	defer z.mu.Unlock()
	
	if v, ok := params["threshold"]; ok {
		if f, ok := v.(float64); ok {
			z.Threshold = f
		}
	}
	if v, ok := params["lookback"]; ok {
		if i, ok := v.(int); ok {
			z.Lookback = i
		}
	}
	return nil
}

// Detect identifies anomalies using Z-score methodology
func (z *ZScoreDetector) Detect(history []models.TimeSeriesPoint, current float64) (bool, float64, float64) {
	z.mu.RLock()
	defer z.mu.RUnlock()
	
	if len(history) < z.MinDataPoints {
		return false, 0.0, 0.5 // Low confidence with insufficient data
	}
	
	// Extract values and apply lookback window
	startIdx := len(history) - z.Lookback
	if startIdx < 0 {
		startIdx = 0
	}
	
	window := history[startIdx:]
	values := make([]float64, len(window))
	for i, point := range window {
		values[i] = point.Value
	}
	
	// Calculate statistics
	mean, stdDev := calculateMeanAndStdDev(values)
	
	if stdDev == 0 {
		// No variance, check if current is different from mean
		if current != mean {
			return true, 1.0, 0.8
		}
		return false, 0.0, 0.9
	}
	
	// Calculate Z-score
	zScore := (current - mean) / stdDev
	absZScore := math.Abs(zScore)
	
	// Determine if anomaly and calculate score
	isAnomaly := absZScore > z.Threshold
	anomalyScore := math.Tanh(absZScore / z.Threshold) // Normalize to 0-1
	
	// Confidence decreases as we move further from threshold
	deviationRatio := absZScore / z.Threshold
	confidence := math.Min(0.95, 0.5+(0.45*math.Min(1.0, deviationRatio/2)))
	
	return isAnomaly, anomalyScore, confidence
}

// IQRDetector implements Interquartile Range anomaly detection
type IQRDetector struct {
	Multiplier   float64
	Lookback     int
	MinDataPoints int
	mu           sync.RWMutex
}

// NewIQRDetector creates a new IQR-based detector
func NewIQRDetector(multiplier float64, lookback int) *IQRDetector {
	return &IQRDetector{
		Multiplier:    multiplier,
		Lookback:      lookback,
		MinDataPoints: 10,
	}
}

// GetName returns the detector name
func (i *IQRDetector) GetName() string {
	return "IQR"
}

// Configure sets detector parameters
func (i *IQRDetector) Configure(params map[string]interface{}) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	
	if v, ok := params["multiplier"]; ok {
		if f, ok := v.(float64); ok {
			i.Multiplier = f
		}
	}
	return nil
}

// Detect identifies anomalies using IQR methodology
func (i *IQRDetector) Detect(history []models.TimeSeriesPoint, current float64) (bool, float64, float64) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	
	if len(history) < i.MinDataPoints {
		return false, 0.0, 0.5
	}
	
	// Extract window values
	startIdx := len(history) - i.Lookback
	if startIdx < 0 {
		startIdx = 0
	}
	
	window := history[startIdx:]
	values := make([]float64, len(window))
	for j, point := range window {
		values[j] = point.Value
	}
	
	// Calculate quartiles
	q1, q3 := calculateQuartiles(values)
	iqr := q3 - q1
	
	lowerBound := q1 - i.Multiplier*iqr
	upperBound := q3 + i.Multiplier*iqr
	
	// Calculate anomaly score based on distance from bounds
	var anomalyScore float64
	var isAnomaly bool
	confidence := 0.75
	
	if current < lowerBound {
		distance := lowerBound - current
		iqrRange := q1 - lowerBound
		if iqrRange > 0 {
			anomalyScore = math.Min(1.0, distance/iqrRange)
		} else {
			anomalyScore = 1.0
		}
		isAnomaly = true
		confidence = 0.85
	} else if current > upperBound {
		distance := current - upperBound
		iqrRange := upperBound - q3
		if iqrRange > 0 {
			anomalyScore = math.Min(1.0, distance/iqrRange)
		} else {
			anomalyScore = 1.0
		}
		isAnomaly = true
		confidence = 0.85
	}
	
	return isAnomaly, anomalyScore, confidence
}

// MovingAverageDetector detects anomalies based on moving average deviation
type MovingAverageDetector struct {
	Lookback       int
	StdDevMultiplier float64
	MinDataPoints   int
	mu              sync.RWMutex
}

// NewMovingAverageDetector creates a new moving average detector
func NewMovingAverageDetector(lookback int, stdDevMultiplier float64) *MovingAverageDetector {
	return &MovingAverageDetector{
		Lookback:         lookback,
		StdDevMultiplier: stdDevMultiplier,
		MinDataPoints:    5,
	}
}

// GetName returns the detector name
func (m *MovingAverageDetector) GetName() string {
	return "MOVING_AVERAGE"
}

// Configure sets detector parameters
func (m *MovingAverageDetector) Configure(params map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if v, ok := params["stddev_multiplier"]; ok {
		if f, ok := v.(float64); ok {
			m.StdDevMultiplier = f
		}
	}
	return nil
}

// Detect identifies anomalies using moving average methodology
func (m *MovingAverageDetector) Detect(history []models.TimeSeriesPoint, current float64) (bool, float64, float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(history) < m.MinDataPoints {
		return false, 0.0, 0.5
	}
	
	startIdx := len(history) - m.Lookback
	if startIdx < 0 {
		startIdx = 0
	}
	
	window := history[startIdx:]
	values := make([]float64, len(window))
	for k, point := range window {
		values[k] = point.Value
	}
	
	mean, stdDev := calculateMeanAndStdDev(values)
	
	// Calculate deviation from moving average in terms of std dev
	deviation := (current - mean)
	
	if stdDev == 0 {
		if deviation != 0 {
			return true, 1.0, 0.8
		}
		return false, 0.0, 0.9
	}
	
	zScore := deviation / stdDev
	absZScore := math.Abs(zScore)
	
	isAnomaly := absZScore > m.StdDevMultiplier
	anomalyScore := math.Min(1.0, absZScore/m.StdDevMultiplier)
	confidence := math.Min(0.9, 0.6+(0.3*math.Min(1.0, absZScore/m.StdDevMultiplier)))
	
	return isAnomaly, anomalyScore, confidence
}

// ExponentialSmoothingDetector uses double exponential smoothing for anomaly detection
type ExponentialSmoothingDetector struct {
	Alpha           float64 // Level smoothing factor
	Beta            float64 // Trend smoothing factor
	Threshold       float64
	MinDataPoints   int
	level           float64
	trend           float64
	initialized     bool
	mu              sync.RWMutex
}

// NewExponentialSmoothingDetector creates a new exponential smoothing detector
func NewExponentialSmoothingDetector(alpha, beta, threshold float64) *ExponentialSmoothingDetector {
	return &ExponentialSmoothingDetector{
		Alpha:         alpha,
		Beta:          beta,
		Threshold:     threshold,
		MinDataPoints: 3,
	}
}

// GetName returns the detector name
func (e *ExponentialSmoothingDetector) GetName() string {
	return "EXPONENTIAL_SMOOTHING"
}

// Configure sets detector parameters
func (e *ExponentialSmoothingDetector) Configure(params map[string]interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if v, ok := params["alpha"]; ok {
		if f, ok := v.(float64); ok {
			e.Alpha = f
		}
	}
	if v, ok := params["beta"]; ok {
		if f, ok := v.(float64); ok {
			e.Beta = f
		}
	}
	return nil
}

// Detect identifies anomalies using exponential smoothing prediction
func (e *ExponentialSmoothingDetector) Detect(history []models.TimeSeriesPoint, current float64) (bool, float64, float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if len(history) < e.MinDataPoints {
		return false, 0.0, 0.5
	}
	
	// Initialize or update level and trend
	if !e.initialized {
		// Initialize with first two points
		if len(history) >= 2 {
			e.level = history[0].Value
			e.trend = history[1].Value - history[0].Value
			e.initialized = true
		}
	} else {
		// Update level and trend using double exponential smoothing
		lastLevel := e.level
		e.level = e.Alpha*current + (1-e.Alpha)*(e.level+e.trend)
		e.trend = e.Beta*(e.level-lastLevel) + (1-e.Beta)*e.trend
	}
	
	// Predict next value
	predicted := e.level + e.trend
	
	// Calculate deviation
	deviation := math.Abs(current - predicted)
	
	// Calculate normalized score
	isAnomaly := deviation > e.Threshold
	anomalyScore := math.Min(1.0, deviation/e.Threshold)
	confidence := 0.8
	
	return isAnomaly, anomalyScore, confidence
}

// EnsembleDetector combines multiple detectors for robust detection
type EnsembleDetector struct {
	Detectors    []Detector
	Weights      []float64
	VoteThreshold float64
	mu           sync.RWMutex
}

// NewEnsembleDetector creates a detector ensemble
func NewEnsembleDetector(detectors []Detector, weights []float64, voteThreshold float64) *EnsembleDetector {
	if len(weights) != len(detectors) {
		weights = make([]float64, len(detectors))
		for i := range weights {
			weights[i] = 1.0 / float64(len(detectors))
		}
	}
	return &EnsembleDetector{
		Detectors:     detectors,
		Weights:       weights,
		VoteThreshold: voteThreshold,
	}
}

// GetName returns the detector name
func (e *EnsembleDetector) GetName() string {
	return "ENSEMBLE"
}

// Configure propagates configuration to all detectors
func (e *EnsembleDetector) Configure(params map[string]interface{}) error {
	for _, detector := range e.Detectors {
		if err := detector.Configure(params); err != nil {
			return err
		}
	}
	return nil
}

// Detect combines results from all detectors
func (e *EnsembleDetector) Detect(history []models.TimeSeriesPoint, current float64) (bool, float64, float64) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if len(e.Detectors) == 0 {
		return false, 0.0, 0.5
	}
	
	totalWeight := 0.0
	weightedScore := 0.0
	anomalyVotes := 0.0
	avgConfidence := 0.0
	
	for i, detector := range e.Detectors {
		isAnomaly, score, confidence := detector.Detect(history, current)
		weight := e.Weights[i]
		
		weightedScore += score * weight
		totalWeight += weight
		
		if isAnomaly {
			anomalyVotes += weight
		}
		avgConfidence += confidence
	}
	
	normalizedScore := weightedScore / totalWeight
	anomalyProbability := anomalyVotes / totalWeight
	avgConfidence /= float64(len(e.Detectors))
	
	// Determine if anomaly based on vote threshold
	isAnomaly := anomalyProbability >= e.VoteThreshold
	
	return isAnomaly, normalizedScore, avgConfidence
}

// Helper Functions

// calculateMeanAndStdDev calculates mean and standard deviation
func calculateMeanAndStdDev(values []float64) (mean, stdDev float64) {
	if len(values) == 0 {
		return 0, 0
	}
	
	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))
	
	// Calculate standard deviation
	if len(values) == 1 {
		return mean, 0
	}
	
	varianceSum := 0.0
	for _, v := range values {
		varianceSum += (v - mean) * (v - mean)
	}
	stdDev = math.Sqrt(varianceSum / float64(len(values)-1))
	
	return mean, stdDev
}

// calculateQuartiles calculates Q1 and Q3 for a sorted slice
func calculateQuartiles(sortedValues []float64) (q1, q3 float64) {
	n := len(sortedValues)
	if n < 4 {
		if n > 0 {
			return sortedValues[0], sortedValues[n-1]
		}
		return 0, 0
	}
	
	// Sort values
	sorted := make([]float64, len(sortedValues))
	copy(sorted, sortedValues)
	
	for i := 0; i < n/2; i++ {
		sorted[i], sorted[n-1-i] = sorted[n-1-i], sorted[i]
	}
	
	// Find median index
	mid := n / 2
	if n%2 == 0 {
		// Even number of elements
		q1 = sorted[mid/2-1]
		q3 = sorted[mid+n%2+mid/2]
	} else {
		// Odd number of elements
		q1 = sorted[mid/2-1]
		q3 = sorted[mid+mid/2]
	}
	
	return q1, q3
}

// DetectorFactory creates detectors based on algorithm name
func DetectorFactory(algorithm string, params map[string]interface{}) Detector {
	switch algorithm {
	case "Z_SCORE":
		detector := NewZScoreDetector(3.0, 30)
		detector.Configure(params)
		return detector
	case "IQR":
		detector := NewIQRDetector(1.5, 30)
		detector.Configure(params)
		return detector
	case "MOVING_AVERAGE":
		detector := NewMovingAverageDetector(30, 2.0)
		detector.Configure(params)
		return detector
	case "EXPONENTIAL_SMOOTHING":
		detector := NewExponentialSmoothingDetector(0.3, 0.1, 2.0)
		detector.Configure(params)
		return detector
	default:
		// Default to Z-Score
		detector := NewZScoreDetector(3.0, 30)
		detector.Configure(params)
		return detector
	}
}

// CreateDefaultEnsemble creates a well-tuned ensemble detector
func CreateDefaultEnsemble() *EnsembleDetector {
	detectors := []Detector{
		NewZScoreDetector(3.0, 30),
		NewIQRDetector(1.5, 30),
		NewMovingAverageDetector(30, 2.5),
	}
	weights := []float64{0.4, 0.3, 0.3}
	return NewEnsembleDetector(detectors, weights, 0.5)
}
