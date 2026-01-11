package dataquality

import (
	"math"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestTimeSeries_Add(t *testing.T) {
	ts := NewTimeSeries(10)

	// Add some points
	for i := 0; i < 5; i++ {
		ts.Add(TimeSeriesPoint{
			Timestamp: time.Now(),
			Value:     float64(i),
		})
	}

	if len(ts.Points) != 5 {
		t.Errorf("Expected 5 points, got %d", len(ts.Points))
	}
}

func TestTimeSeries_MaintainMaxSize(t *testing.T) {
	ts := NewTimeSeries(5)

	// Add more than max size
	for i := 0; i < 10; i++ {
		ts.Add(TimeSeriesPoint{
			Timestamp: time.Now(),
			Value:     float64(i),
		})
	}

	if len(ts.Points) != 5 {
		t.Errorf("Expected 5 points (max size), got %d", len(ts.Points))
	}

	// Should have the most recent 5 values (5, 6, 7, 8, 9)
	values := ts.GetValues()
	if values[0] != 5 {
		t.Errorf("Expected first value to be 5, got %f", values[0])
	}
}

func TestTimeSeries_GetRecent(t *testing.T) {
	ts := NewTimeSeries(10)

	for i := 0; i < 10; i++ {
		ts.Add(TimeSeriesPoint{
			Timestamp: time.Now(),
			Value:     float64(i),
		})
	}

	recent := ts.GetRecent(3)
	if len(recent) != 3 {
		t.Errorf("Expected 3 recent points, got %d", len(recent))
	}
}

func TestTimeSeries_GetValues(t *testing.T) {
	ts := NewTimeSeries(10)

	expectedValues := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	for _, v := range expectedValues {
		ts.Add(TimeSeriesPoint{Timestamp: time.Now(), Value: v})
	}

	values := ts.GetValues()
	if len(values) != len(expectedValues) {
		t.Errorf("Expected %d values, got %d", len(expectedValues), len(values))
	}
}

func TestTimeSeries_Calculate(t *testing.T) {
	ts := NewTimeSeries(10)

	// Add values: 10, 20, 30, 40, 50
	for i := 1; i <= 5; i++ {
		ts.Add(TimeSeriesPoint{
			Timestamp: time.Now(),
			Value:     float64(i * 10),
		})
	}

	stats := ts.Calculate()

	// Mean should be 30
	if math.Abs(stats.Mean-30.0) > 0.001 {
		t.Errorf("Expected mean 30, got %f", stats.Mean)
	}

	// Min should be 10
	if stats.Min != 10 {
		t.Errorf("Expected min 10, got %f", stats.Min)
	}

	// Max should be 50
	if stats.Max != 50 {
		t.Errorf("Expected max 50, got %f", stats.Max)
	}

	// Count should be 5
	if stats.Count != 5 {
		t.Errorf("Expected count 5, got %d", stats.Count)
	}
}

func TestTimeSeries_StdDev(t *testing.T) {
	ts := NewTimeSeries(10)

	// Add values with known std dev: 10, 20, 30
	// Mean = 20, deviations = -10, 0, 10
	// Variance = (100 + 0 + 100) / 3 = 66.67
	// Std dev = sqrt(66.67) ≈ 8.16
	for _, v := range []float64{10, 20, 30} {
		ts.Add(TimeSeriesPoint{Timestamp: time.Now(), Value: v})
	}

	stats := ts.Calculate()
	expectedStdDev := math.Sqrt(66.66666666666667)

	if math.Abs(stats.StdDev-expectedStdDev) > 0.01 {
		t.Errorf("Expected std dev %f, got %f", expectedStdDev, stats.StdDev)
	}
}

func TestDetector_Observe(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultAnomalyConfig()
	config.MinWindowSize = 3 // Lower for testing
	detector := NewDetector(config, logger)

	// Add some normal values first
	for i := 0; i < 5; i++ {
		anomaly := detector.Observe("test-source", "metric", 50.0+float64(i))
		if anomaly != nil {
			t.Errorf("Expected no anomaly for normal values, got one")
		}
	}

	// Add a spike value
	anomaly := detector.Observe("test-source", "metric", 1000.0)
	if anomaly == nil {
		t.Error("Expected anomaly for spike value")
	}

	if anomaly.Type != AnomalyTypeSpike {
		t.Errorf("Expected spike anomaly, got %s", anomaly.Type)
	}
}

func TestDetector_ObserveScoreDrop(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultAnomalyConfig()
	config.MinWindowSize = 5
	detector := NewDetector(config, logger)

	// Add normal values around 90
	for i := 0; i < 5; i++ {
		detector.Observe("test-source", "quality_score", 90.0)
	}

	// Add a significant drop
	anomaly := detector.Observe("test-source", "quality_score", 50.0)
	if anomaly == nil {
		t.Error("Expected anomaly for score drop")
	}

	if anomaly.Type != AnomalyTypeScoreDrop {
		t.Errorf("Expected score drop anomaly, got %s", anomaly.Type)
	}
}

func TestDetector_Cooldown(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultAnomalyConfig()
	config.MinWindowSize = 3
	config.AlertCooldown = 10 * time.Second
	detector := NewDetector(config, logger)

	// Add normal values
	for i := 0; i < 5; i++ {
		detector.Observe("test-source", "metric", 50.0)
	}

	// Trigger first anomaly
	anomaly1 := detector.Observe("test-source", "metric", 1000.0)
	if anomaly1 == nil {
		t.Error("Expected first anomaly")
	}

	// Trigger second anomaly immediately (should be suppressed by cooldown)
	anomaly2 := detector.Observe("test-source", "metric", 1000.0)
	if anomaly2 != nil {
		t.Error("Expected no anomaly due to cooldown")
	}
}

func TestDetector_GetStatistics(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultAnomalyConfig()
	detector := NewDetector(config, logger)

	// Add some values
	for i := 0; i < 10; i++ {
		detector.Observe("test-source", "metric", float64(i))
	}

	stats := detector.GetStatistics("test-source", "metric")

	if stats.Count != 10 {
		t.Errorf("Expected count 10, got %d", stats.Count)
	}

	if stats.Mean != 4.5 {
		t.Errorf("Expected mean 4.5, got %f", stats.Mean)
	}
}

func TestDetector_Reset(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultAnomalyConfig()
	detector := NewDetector(config, logger)

	// Add some values
	for i := 0; i < 10; i++ {
		detector.Observe("test-source", "metric", float64(i))
	}

	// Reset
	detector.Reset()

	// Statistics should be empty
	stats := detector.GetStatistics("test-source", "metric")
	if stats.Count != 0 {
		t.Errorf("Expected count 0 after reset, got %d", stats.Count)
	}
}

func TestDetector_ResetSource(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultAnomalyConfig()
	detector := NewDetector(config, logger)

	// Add values for two sources
	for i := 0; i < 10; i++ {
		detector.Observe("source-1", "metric", float64(i))
		detector.Observe("source-2", "metric", float64(i))
	}

	// Reset one source
	detector.ResetSource("source-1")

	// Source-1 should be cleared
	stats1 := detector.GetStatistics("source-1", "metric")
	if stats1.Count != 0 {
		t.Errorf("Expected count 0 for source-1 after reset, got %d", stats1.Count)
	}

	// Source-2 should remain
	stats2 := detector.GetStatistics("source-2", "metric")
	if stats2.Count != 10 {
		t.Errorf("Expected count 10 for source-2, got %d", stats2.Count)
	}
}

func TestPercentile(t *testing.T) {
	sorted := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		p       float64
		expected float64
	}{
		{0, 1},
		{0.5, 5.5},
		{1, 10},
		{0.25, 3.25},
		{0.75, 7.75},
	}

	for _, tt := range tests {
		result := percentile(sorted, tt.p)
		if result != tt.expected {
			t.Errorf("percentile(%f) = %f, expected %f", tt.p, result, tt.expected)
		}
	}
}

func TestAnomaly_Recommendations(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultAnomalyConfig()
	config.EnableAlerts = false
	detector := NewDetector(config, logger)

	// Set up alert function
	var alerts []*Anomaly
	detector.SetAlertFunc(func(a *Anomaly) {
		alerts = append(alerts, a)
	})

	// Add normal values
	for i := 0; i < 10; i++ {
		detector.Observe("test-source", "metric", 50.0)
	}

	// Trigger anomaly
	detector.Observe("test-source", "metric", 1000.0)

	if len(alerts) == 0 {
		t.Error("Expected alert to be generated")
	}

	if len(alerts[0].Recommendations) == 0 {
		t.Error("Expected recommendations to be generated")
	}
}

func BenchmarkDetector_Observe(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultAnomalyConfig()
	detector := NewDetector(config, logger)

	// Pre-populate with normal values
	for i := 0; i < 100; i++ {
		detector.Observe("test-source", "metric", 50.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.Observe("test-source", "metric", 50.0+float64(i%10))
	}
}
