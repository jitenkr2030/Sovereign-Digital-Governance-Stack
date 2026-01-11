package anomaly

import (
	"math"
	"testing"
	"time"

	"neam-platform/intelligence/models"
)

func TestZScoreDetector(t *testing.T) {
	detector := NewZScoreDetector(3.0, 30)
	
	// Test with normal data
	history := generateTestHistory(30, 100, 10)
	current := 105.0 // Within normal range
	
	isAnomaly, score, confidence := detector.Detect(history, current)
	
	if isAnomaly {
		t.Errorf("Expected no anomaly for value within normal range, got anomaly")
	}
	
	if score < 0 || score > 1 {
		t.Errorf("Expected score between 0 and 1, got %f", score)
	}
}

func TestZScoreDetectorAnomaly(t *testing.T) {
	detector := NewZScoreDetector(2.0, 30)
	
	// Test with anomalous data
	history := generateTestHistory(30, 100, 10)
	current := 250.0 // Significant outlier
	
	isAnomaly, score, confidence := detector.Detect(history, current)
	
	if !isAnomaly {
		t.Errorf("Expected anomaly for significant outlier")
	}
	
	if score < 0.5 {
		t.Errorf("Expected high anomaly score for significant outlier, got %f", score)
	}
}

func TestZScoreDetectorInsufficientData(t *testing.T) {
	detector := NewZScoreDetector(3.0, 30)
	
	// Test with insufficient data
	history := generateTestHistory(3, 100, 10)
	current := 150.0
	
	isAnomaly, score, confidence := detector.Detect(history, current)
	
	if isAnomaly {
		t.Errorf("Expected no anomaly with insufficient data")
	}
	
	if confidence > 0.6 {
		t.Errorf("Expected lower confidence with insufficient data, got %f", confidence)
	}
}

func TestIQRDetector(t *testing.T) {
	detector := NewIQRDetector(1.5, 30)
	
	// Test with normal data
	history := generateTestHistory(30, 100, 10)
	current := 105.0
	
	isAnomaly, score, confidence := detector.Detect(history, current)
	
	if isAnomaly {
		t.Errorf("Expected no anomaly for value within IQR bounds")
	}
}

func TestIQRDetectorOutlier(t *testing.T) {
	detector := NewIQRDetector(1.5, 30)
	
	// Create history with specific distribution
	history := []models.TimeSeriesPoint{
		{Value: 100}, {Value: 101}, {Value: 99}, {Value: 102}, {Value: 98},
		{Value: 100}, {Value: 101}, {Value: 99}, {Value: 102}, {Value: 98},
		{Value: 100}, {Value: 101}, {Value: 99}, {Value: 102}, {Value: 98},
		{Value: 100}, {Value: 101}, {Value: 99}, {Value: 102}, {Value: 98},
		{Value: 100}, {Value: 101}, {Value: 99}, {Value: 102}, {Value: 98},
		{Value: 100}, {Value: 101}, {Value: 99}, {Value: 102}, {Value: 98},
	}
	
	// Test with extreme outlier
	current := 200.0
	
	isAnomaly, score, confidence := detector.Detect(history, current)
	
	if !isAnomaly {
		t.Errorf("Expected anomaly for extreme outlier")
	}
}

func TestMovingAverageDetector(t *testing.T) {
	detector := NewMovingAverageDetector(30, 2.0)
	
	// Test with normal deviation
	history := generateTestHistory(30, 100, 5)
	current := 108.0 // Within 2 std dev
	
	isAnomaly, score, confidence := detector.Detect(history, current)
	
	if isAnomaly {
		t.Errorf("Expected no anomaly for value within normal deviation")
	}
}

func TestMovingAverageDetectorAnomaly(t *testing.T) {
	detector := NewMovingAverageDetector(30, 2.0)
	
	// Test with significant deviation
	history := generateTestHistory(30, 100, 5)
	current := 150.0 // Significant deviation
	
	isAnomaly, score, confidence := detector.Detect(history, current)
	
	if !isAnomaly {
		t.Errorf("Expected anomaly for significant deviation")
	}
}

func TestExponentialSmoothingDetector(t *testing.T) {
	detector := NewExponentialSmoothingDetector(0.3, 0.1, 2.0)
	
	// Build up history
	history := generateTestHistory(10, 100, 5)
	
	// Test with value close to prediction
	current := 102.0
	
	isAnomaly, score, confidence := detector.Detect(history, current)
	
	// Should not be anomaly if close to predicted value
	if score > 0.5 {
		t.Errorf("Expected low score for value close to prediction, got %f", score)
	}
}

func TestEnsembleDetector(t *testing.T) {
	detectors := []Detector{
		NewZScoreDetector(3.0, 30),
		NewIQRDetector(1.5, 30),
		NewMovingAverageDetector(30, 2.5),
	}
	weights := []float64{0.4, 0.3, 0.3}
	ensemble := NewEnsembleDetector(detectors, weights, 0.5)
	
	// Test with normal data
	history := generateTestHistory(30, 100, 10)
	current := 105.0
	
	isAnomaly, score, confidence := ensemble.Detect(history, current)
	
	if isAnomaly {
		t.Errorf("Expected no anomaly for normal data with ensemble")
	}
	
	// Test with anomaly
	current = 250.0
	
	isAnomaly, score, confidence = ensemble.Detect(history, current)
	
	if !isAnomaly {
		t.Errorf("Expected anomaly detected by ensemble")
	}
}

func TestDetectorFactory(t *testing.T) {
	tests := []struct {
		algorithm string
		expected  string
	}{
		{"Z_SCORE", "Z_SCORE"},
		{"IQR", "IQR"},
		{"MOVING_AVERAGE", "MOVING_AVERAGE"},
		{"EXPONENTIAL_SMOOTHING", "EXPONENTIAL_SMOOTHING"},
		{"UNKNOWN", "Z_SCORE"}, // Default
	}
	
	for _, tt := range tests {
		detector := DetectorFactory(tt.algorithm, nil)
		if detector.GetName() != tt.expected {
			t.Errorf("Expected detector %s, got %s", tt.expected, detector.GetName())
		}
	}
}

func TestCreateDefaultEnsemble(t *testing.T) {
	ensemble := CreateDefaultEnsemble()
	
	if ensemble == nil {
		t.Fatal("Expected non-nil ensemble")
	}
	
	if len(ensemble.Detectors) != 3 {
		t.Errorf("Expected 3 detectors in ensemble, got %d", len(ensemble.Detectors))
	}
	
	// Verify weights sum to 1.0
	var weightSum float64
	for _, w := range ensemble.Weights {
		weightSum += w
	}
	
	if math.Abs(weightSum-1.0) > 0.001 {
		t.Errorf("Expected weights to sum to 1.0, got %f", weightSum)
	}
}

// Benchmark tests

func BenchmarkZScoreDetector(b *testing.B) {
	detector := NewZScoreDetector(3.0, 30)
	history := generateTestHistory(100, 1000, 100)
	current := 1050.0
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.Detect(history, current)
	}
}

func BenchmarkIQRDetector(b *testing.B) {
	detector := NewIQRDetector(1.5, 30)
	history := generateTestHistory(100, 1000, 100)
	current := 1200.0
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.Detect(history, current)
	}
}

func BenchmarkEnsembleDetector(b *testing.B) {
	ensemble := CreateDefaultEnsemble()
	history := generateTestHistory(100, 1000, 100)
	current := 1050.0
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ensemble.Detect(history, current)
	}
}

// Helper functions

func generateTestHistory(n int, mean float64, stdDev float64) []models.TimeSeriesPoint {
	history := make([]models.TimeSeriesPoint, n)
	for i := 0; i < n; i++ {
		// Generate normal-ish distribution
		value := mean + (float64(i%10)-4.5)*stdDev/2
		history[i] = models.TimeSeriesPoint{
			Timestamp: time.Now().Add(-time.Duration(n-i) * time.Hour),
			Value:     value,
		}
	}
	return history
}
