package anomalydetection_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"neam-platform/black-economy/anomalydetection"
)

// MockDetector is a mock implementation of the detector interface
type MockDetector struct {
	mock.Mock
}

func (m *MockDetector) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDetector) Detect(ctx context.Context, data map[string]interface{}) (bool, float64, error) {
	args := m.Called(ctx, data)
	return args.Bool(0), args.Float64(1), args.Error(2)
}

func (m *MockDetector) Threshold() float64 {
	args := m.Called()
	return args.Float64(0)
}

// TestService creation and initialization
func TestNewAnomalyDetectionService(t *testing.T) {
	t.Run("creates service with valid config", func(t *testing.T) {
		config := anomalydetection.ServiceConfig{
			DetectionInterval:   time.Minute,
			MaxConcurrentChecks: 10,
			EnableRealTimeAlerts: true,
		}

		service, err := anomalydetection.NewAnomalyDetectionService(config)
		require.NoError(t, err)
		require.NotNil(t, service)
		assert.Equal(t, config.DetectionInterval, service.GetDetectionInterval())
		assert.Equal(t, config.EnableRealTimeAlerts, service.IsRealTimeAlertsEnabled())
	})

	t.Run("creates service with default config", func(t *testing.T) {
		service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{})
		require.NoError(t, err)
		require.NotNil(t, service)
	})

	t.Run("fails with invalid max concurrent checks", func(t *testing.T) {
		config := anomalydetection.ServiceConfig{
			MaxConcurrentChecks: 0,
		}
		service, err := anomalydetection.NewAnomalyDetectionService(config)
		assert.Error(t, err)
		assert.Nil(t, service)
	})
}

// Test detector registration
func TestServiceRegisterDetector(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{})
	require.NoError(t, err)

	mockDetector := new(MockDetector)
	mockDetector.On("Name").Return("test_detector")
	mockDetector.On("Threshold").Return(0.8)

	err = service.RegisterDetector(mockDetector)
	require.NoError(t, err)

	detectors := service.GetRegisteredDetectors()
	assert.Contains(t, detectors, "test_detector")
}

func TestServiceRegisterDuplicateDetector(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{})
	require.NoError(t, err)

	mockDetector1 := new(MockDetector)
	mockDetector1.On("Name").Return("duplicate_detector")
	mockDetector1.On("Threshold").Return(0.8)

	mockDetector2 := new(MockDetector)
	mockDetector2.On("Name").Return("duplicate_detector")
	mockDetector2.On("Threshold").Return(0.7)

	err = service.RegisterDetector(mockDetector1)
	require.NoError(t, err)

	err = service.RegisterDetector(mockDetector2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

// Test anomaly detection
func TestServiceDetectAnomalies(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{
		MaxConcurrentChecks: 5,
	})
	require.NoError(t, err)

	mockDetector := new(MockDetector)
	mockDetector.On("Name").Return("test_detector")
	mockDetector.On("Threshold").Return(0.8)
	mockDetector.On("Detect", mock.Anything, mock.Anything).Return(true, 0.9, nil)

	err = service.RegisterDetector(mockDetector)
	require.NoError(t, err)

	testData := map[string]interface{}{
		"transaction_amount": 15000.0,
		"account_age":        30,
		"location_risk":      0.7,
	}

	result, err := service.DetectAnomalies(context.Background(), testData)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsAnomaly)
	assert.Greater(t, result.Score, 0.8)
	assert.Len(t, result.DetectorResults, 1)
}

func TestServiceDetectAnomaliesWithNoDetectors(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{})
	require.NoError(t, err)

	testData := map[string]interface{}{
		"transaction_amount": 15000.0,
	}

	result, err := service.DetectAnomalies(context.Background(), testData)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsAnomaly)
	assert.Empty(t, result.DetectorResults)
}

func TestServiceDetectAnomaliesWithMultipleDetectors(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{
		MaxConcurrentChecks: 10,
	})
	require.NoError(t, err)

	// Register multiple detectors
	for i := 1; i <= 3; i++ {
		mockDetector := new(MockDetector)
		mockDetector.On("Name").Return("detector_" + string(rune('0'+i)))
		mockDetector.On("Threshold").Return(0.7)
		mockDetector.On("Detect", mock.Anything, mock.Anything).Return(i%2 == 0, 0.6+float64(i)*0.1, nil)
		
		err = service.RegisterDetector(mockDetector)
		require.NoError(t, err)
	}

	testData := map[string]interface{}{
		"value": 100.0,
	}

	result, err := service.DetectAnomalies(context.Background(), testData)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsAnomaly) // At least one detector returned true
	assert.Len(t, result.DetectorResults, 3)
}

// Test real-time alert configuration
func TestServiceConfigureAlerts(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{})
	require.NoError(t, err)

	alertConfig := anomalydetection.AlertConfig{
		Enabled:          true,
		SeverityThreshold: 0.9,
		Channels:         []string{"email", "slack"},
		CooldownPeriod:   time.Minute * 5,
	}

	err = service.ConfigureAlerts(alertConfig)
	require.NoError(t, err)

	config := service.GetAlertConfig()
	assert.True(t, config.Enabled)
	assert.Equal(t, 0.9, config.SeverityThreshold)
	assert.Contains(t, config.Channels, "email")
	assert.Contains(t, config.Channels, "slack")
}

// Test statistics and metrics
func TestServiceGetStatistics(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{})
	require.NoError(t, err)

	stats := service.GetStatistics()
	require.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.TotalChecks, int64(0))
	assert.GreaterOrEqual(t, stats.AnomaliesDetected, int64(0))
}

// Test detector removal
func TestServiceRemoveDetector(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{})
	require.NoError(t, err)

	mockDetector := new(MockDetector)
	mockDetector.On("Name").Return("removable_detector")
	mockDetector.On("Threshold").Return(0.8)

	err = service.RegisterDetector(mockDetector)
	require.NoError(t, err)

	err = service.RemoveDetector("removable_detector")
	require.NoError(t, err)

	detectors := service.GetRegisteredDetectors()
	assert.NotContains(t, detectors, "removable_detector")
}

func TestServiceRemoveNonExistentDetector(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{})
	require.NoError(t, err)

	err = service.RemoveDetector("non_existent")
	assert.Error(t, err)
}

// Test batch detection
func TestServiceBatchDetectAnomalies(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{
		MaxConcurrentChecks: 10,
	})
	require.NoError(t, err)

	mockDetector := new(MockDetector)
	mockDetector.On("Name").Return("batch_detector")
	mockDetector.On("Threshold").Return(0.7)
	mockDetector.On("Detect", mock.Anything, mock.Anything).Return(false, 0.5, nil)

	err = service.RegisterDetector(mockDetector)
	require.NoError(t, err)

	batchData := []map[string]interface{}{
		{"value": 100},
		{"value": 200},
		{"value": 300},
	}

	results, err := service.BatchDetectAnomalies(context.Background(), batchData)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	for _, result := range results {
		assert.False(t, result.IsAnomaly)
	}
}

// Test context cancellation
func TestServiceDetectWithCancelledContext(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{})
	require.NoError(t, err)

	mockDetector := new(MockDetector)
	mockDetector.On("Name").Return("cancelled_detector")
	mockDetector.On("Threshold").Return(0.8)
	mockDetector.On("Detect", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		// Simulate slow detection
		time.Sleep(time.Second)
	}).Return(true, 0.9, nil)

	err = service.RegisterDetector(mockDetector)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	testData := map[string]interface{}{
		"value": 100,
	}

	result, err := service.DetectAnomalies(ctx, testData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
	assert.Nil(t, result)
}

// Test JSON serialization
func TestAnomalyResultSerialization(t *testing.T) {
	result := anomalydetection.AnomalyResult{
		IsAnomaly:         true,
		Score:             0.85,
		Confidence:        0.92,
		DetectionTime:     time.Now(),
		DetectorResults:   []anomalydetection.DetectorResult{
			{
				DetectorName: "test_detector",
				IsAnomaly:    true,
				Score:        0.85,
				Details:      map[string]interface{}{"reason": "high_value"},
			},
		},
	}

	// Serialize to JSON
	data, err := json.Marshal(result)
	require.NoError(t, err)

	// Deserialize back
	var decoded anomalydetection.AnomalyResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.IsAnomaly, decoded.IsAnomaly)
	assert.Equal(t, result.Score, decoded.Score)
	assert.Len(t, decoded.DetectorResults, 1)
	assert.Equal(t, "test_detector", decoded.DetectorResults[0].DetectorName)
}

// Test detector result struct
func TestDetectorResultCreation(t *testing.T) {
	result := anomalydetection.DetectorResult{
		DetectorName: "test_detector",
		IsAnomaly:    true,
		Score:        0.75,
		Details: map[string]interface{}{
			"threshold": 0.7,
			"observed":  0.75,
		},
		ExecutionTime: 50 * time.Millisecond,
	}

	assert.Equal(t, "test_detector", result.DetectorName)
	assert.True(t, result.IsAnomaly)
	assert.Equal(t, 0.75, result.Score)
	assert.Contains(t, result.Details, "threshold")
	assert.Contains(t, result.Details, "observed")
	assert.Equal(t, 50*time.Millisecond, result.ExecutionTime)
}

// Test service cleanup
func TestServiceClose(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{})
	require.NoError(t, err)

	// Close should not error
	err = service.Close()
	require.NoError(t, err)

	// Operations after close should fail
	_, err = service.DetectAnomalies(context.Background(), map[string]interface{}{"value": 100})
	assert.Error(t, err)
}

// Test with real detector implementations
func TestWithRealDetectorImplementations(t *testing.T) {
	t.Run("threshold detector", func(t *testing.T) {
		detector := anomalydetection.NewThresholdDetector(0.8)
		assert.Equal(t, "threshold", detector.Name())
		assert.Equal(t, 0.8, detector.Threshold())

		result, score, err := detector.Detect(context.Background(), map[string]interface{}{"value": 0.9})
		require.NoError(t, err)
		assert.True(t, result)
		assert.GreaterOrEqual(t, score, 0.8)
	})

	t.Run("threshold detector below threshold", func(t *testing.T) {
		detector := anomalydetection.NewThresholdDetector(0.8)
		result, score, err := detector.Detect(context.Background(), map[string]interface{}{"value": 0.5})
		require.NoError(t, err)
		assert.False(t, result)
		assert.Less(t, score, 0.8)
	})

	t.Run("z-score detector", func(t *testing.T) {
		detector := anomalydetection.NewZScoreDetector(3.0)
		assert.Equal(t, "zscore", detector.Name())

		// Normal value
		result, _, err := detector.Detect(context.Background(), map[string]interface{}{"value": 0.0, "mean": 0.0, "stddev": 1.0})
		require.NoError(t, err)
		assert.False(t, result)

		// Extreme value (3+ stddev)
		result, _, err = detector.Detect(context.Background(), map[string]interface{}{"value": 4.0, "mean": 0.0, "stddev": 1.0})
		require.NoError(t, err)
		assert.True(t, result)
	})
}

// Test statistical detector
func TestStatisticalDetector(t *testing.T) {
	detector := anomalydetection.NewStatisticalDetector(anomalydetection.StatisticalConfig{
		WindowSize:   100,
		Sensitivity:  2.0,
		MinDataPoints: 10,
	})

	assert.Equal(t, "statistical", detector.Name())

	// Test with insufficient data
	result, score, err := detector.Detect(context.Background(), map[string]interface{}{"value": 100.0})
	require.NoError(t, err)
	assert.False(t, result) // Insufficient data
	assert.Equal(t, 0.0, score)

	// Add enough data points
	for i := 0; i < 15; i++ {
		_, _, err := detector.Detect(context.Background(), map[string]interface{}{"value": float64(i)})
		require.NoError(t, err)
	}

	// Test with normal value
	result, _, err = detector.Detect(context.Background(), map[string]interface{}{"value": 5.0})
	require.NoError(t, err)
	assert.False(t, result)
}

// Test service health check
func TestServiceHealthCheck(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{})
	require.NoError(t, err)

	health := service.HealthCheck()
	assert.True(t, health.Healthy)
	assert.Equal(t, "operational", health.Status)
	assert.Empty(t, health.ErrorMessage)
}

func TestServiceHealthCheckWithFailedDetector(t *testing.T) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{})
	require.NoError(t, err)

	mockDetector := new(MockDetector)
	mockDetector.On("Name").Return("failing_detector")
	mockDetector.On("Threshold").Return(0.8)
	mockDetector.On("Detect", mock.Anything, mock.Anything).Return(false, 0.0, assert.AnError)

	err = service.RegisterDetector(mockDetector)
	require.NoError(t, err)

	health := service.HealthCheck()
	assert.False(t, health.Healthy)
	assert.Equal(t, "degraded", health.Status)
	assert.Contains(t, health.ErrorMessage, "failing_detector")
}

// Benchmark tests
func BenchmarkDetectorDetect(b *testing.B) {
	detector := anomalydetection.NewThresholdDetector(0.8)
	ctx := context.Background()
	data := map[string]interface{}{"value": 0.85}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.Detect(ctx, data)
	}
}

func BenchmarkServiceDetectAnomalies(b *testing.B) {
	service, err := anomalydetection.NewAnomalyDetectionService(anomalydetection.ServiceConfig{
		MaxConcurrentChecks: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		detector := anomalydetection.NewThresholdDetector(0.7 + float64(i)*0.05)
		service.RegisterDetector(detector)
	}

	ctx := context.Background()
	data := map[string]interface{}{"value": 0.85}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.DetectAnomalies(ctx, data)
	}
}
