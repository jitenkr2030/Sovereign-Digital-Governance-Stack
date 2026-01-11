package risk_scoring_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"neam-platform/black-economy/risk_scoring"
)

// MockRiskModel is a mock implementation of the risk model interface
type MockRiskModel struct {
	mock.Mock
}

func (m *MockRiskModel) CalculateScore(ctx context.Context, entityData map[string]interface{}) (float64, error) {
	args := m.Called(ctx, entityData)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRiskModel) GetModelVersion() string {
	args := m.Called()
	return args.String(0)
}

// TestRiskScore creation and initialization
func TestNewRiskScoringService(t *testing.T) {
	t.Run("creates service with valid config", func(t *testing.T) {
		config := risk_scoring.ServiceConfig{
			ScoreRange:        risk_scoring.ScoreRange{Min: 0, Max: 100},
			CacheTTL:          time.Hour,
			EnableAggregation: true,
		}

		service, err := risk_scoring.NewRiskScoringService(config)
		require.NoError(t, err)
		require.NotNil(t, service)
		assert.Equal(t, 0.0, service.GetMinScore())
		assert.Equal(t, 100.0, service.GetMaxScore())
	})

	t.Run("creates service with custom range", func(t *testing.T) {
		config := risk_scoring.ServiceConfig{
			ScoreRange:        risk_scoring.ScoreRange{Min: 300, Max: 850},
			CacheTTL:          30 * time.Minute,
			EnableAggregation: false,
		}

		service, err := risk_scoring.NewRiskScoringService(config)
		require.NoError(t, err)
		require.NotNil(t, service)
		assert.Equal(t, 300.0, service.GetMinScore())
		assert.Equal(t, 850.0, service.GetMaxScore())
	})

	t.Run("fails with invalid score range", func(t *testing.T) {
		config := risk_scoring.ServiceConfig{
			ScoreRange: risk_scoring.ScoreRange{Min: 100, Max: 50}, // Min > Max
		}
		service, err := risk_scoring.NewRiskScoringService(config)
		assert.Error(t, err)
		assert.Nil(t, service)
	})
}

// Test model registration
func TestServiceRegisterModel(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{})
	require.NoError(t, err)

	mockModel := new(MockRiskModel)
	mockModel.On("GetModelVersion").Return("v1.0.0")

	err = service.RegisterModel("credit_model", mockModel)
	require.NoError(t, err)

	models := service.GetRegisteredModels()
	assert.Contains(t, models, "credit_model")
}

func TestServiceRegisterDuplicateModel(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{})
	require.NoError(t, err)

	mockModel1 := new(MockRiskModel)
	mockModel1.On("GetModelVersion").Return("v1.0.0")

	mockModel2 := new(MockRiskModel)
	mockModel2.On("GetModelVersion").Return("v2.0.0")

	err = service.RegisterModel("duplicate_model", mockModel1)
	require.NoError(t, err)

	err = service.RegisterModel("duplicate_model", mockModel2)
	assert.Error(t, err)
}

// Test risk calculation
func TestServiceCalculateRisk(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{
		ScoreRange: risk_scoring.ScoreRange{Min: 0, Max: 100},
	})
	require.NoError(t, err)

	mockModel := new(MockRiskModel)
	mockModel.On("CalculateScore", mock.Anything, mock.Anything).Return(75.0, nil)

	err = service.RegisterModel("test_model", mockModel)
	require.NoError(t, err)

	entityData := map[string]interface{}{
		"credit_score":     650,
		"income":           50000,
		"debt_to_income":   0.3,
		"payment_history":  0.95,
	}

	result, err := service.CalculateRisk(context.Background(), entityData)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 75.0, result.Score)
	assert.Equal(t, "medium", result.RiskLevel)
	assert.Equal(t, "v1.0.0", result.ModelVersion)
}

func TestServiceCalculateRiskWithMultipleModels(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{
		ScoreRange:        risk_scoring.ScoreRange{Min: 0, Max: 100},
		EnableAggregation: true,
	})
	require.NoError(t, err)

	// Register multiple models
	for i := 1; i <= 3; i++ {
		mockModel := new(MockRiskModel)
		mockModel.On("GetModelVersion").Return("v1." + string(rune('0'+i)))
		mockModel.On("CalculateScore", mock.Anything, mock.Anything).Return(float64(50+i*10), nil)
		
		err = service.RegisterModel("model_"+string(rune('0'+i)), mockModel)
		require.NoError(t, err)
	}

	entityData := map[string]interface{}{"value": 100}

	result, err := service.CalculateRisk(context.Background(), entityData)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 80.0, result.Score) // Average of 60, 70, 80, 90
	assert.Len(t, result.ModelResults, 3)
}

func TestServiceCalculateRiskWithNoModels(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{})
	require.NoError(t, err)

	entityData := map[string]interface{}{"value": 100}

	result, err := service.CalculateRisk(context.Background(), entityData)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no models registered")
}

// Test risk levels
func TestRiskLevels(t *testing.T) {
	testCases := []struct {
		name           string
		score          float64
		expectedLevel  string
		expectedColor  string
	}{
		{"very low risk", 10.0, "very_low", "green"},
		{"low risk", 25.0, "low", "green"},
		{"medium-low risk", 40.0, "medium_low", "yellow"},
		{"medium risk", 50.0, "medium", "yellow"},
		{"medium-high risk", 65.0, "medium_high", "orange"},
		{"high risk", 80.0, "high", "red"},
		{"very high risk", 95.0, "very_high", "red"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			level := risk_scoring.GetRiskLevel(tc.score)
			assert.Equal(t, tc.expectedLevel, level.Level)
			assert.Equal(t, tc.expectedColor, level.Color)
		})
	}
}

// Test score normalization
func TestServiceNormalizeScore(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{
		ScoreRange: risk_scoring.ScoreRange{Min: 0, Max: 100},
	})
	require.NoError(t, err)

	testCases := []struct {
		name     string
		rawScore float64
		expected float64
	}{
		{"minimum score", 0.0, 0.0},
		{"maximum score", 100.0, 100.0},
		{"mid score", 50.0, 50.0},
		{"quarter score", 25.0, 25.0},
		{"three-quarter score", 75.0, 75.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normalized := service.NormalizeScore(tc.rawScore)
			assert.Equal(t, tc.expected, normalized)
		})
	}
}

func TestServiceNormalizeScoreOutOfRange(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{
		ScoreRange: risk_scoring.ScoreRange{Min: 0, Max: 100},
	})
	require.NoError(t, err)

	// Test below minimum
	normalized := service.NormalizeScore(-10.0)
	assert.Equal(t, 0.0, normalized)

	// Test above maximum
	normalized = service.NormalizeScore(150.0)
	assert.Equal(t, 100.0, normalized)
}

// Test score aggregation
func TestServiceAggregateScores(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{
		EnableAggregation: true,
	})
	require.NoError(t, err)

	scores := []float64{30.0, 50.0, 70.0, 90.0}
	
	result := service.AggregateScores(scores)
	assert.Equal(t, 60.0, result) // Average
	assert.Equal(t, 30.0, result.Min)
	assert.Equal(t, 90.0, result.Max)
	assert.Equal(t, 24.08, result.StdDev)
}

// Test confidence calculation
func TestServiceCalculateConfidence(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{})
	require.NoError(t, err)

	testCases := []struct {
		name           string
		factorCount    int
		dataQuality    float64
		expectedRange  [2]float64
	}{
		{"few factors, low quality", 3, 0.5, [2]float64{0.3, 0.6}},
		{"few factors, high quality", 3, 0.9, [2]float64{0.5, 0.8}},
		{"many factors, low quality", 20, 0.5, [2]float64{0.6, 0.85}},
		{"many factors, high quality", 20, 0.95, [2]float64{0.8, 0.95}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			confidence := service.CalculateConfidence(tc.factorCount, tc.dataQuality)
			assert.GreaterOrEqual(t, confidence, tc.expectedRange[0])
			assert.LessOrEqual(t, confidence, tc.expectedRange[1])
		})
	}
}

// Test risk trend analysis
func TestServiceAnalyzeTrend(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{})
	require.NoError(t, err)

	// Increasing trend
	scores := []float64{30.0, 40.0, 50.0, 60.0, 70.0}
	trend := service.AnalyzeTrend(scores)
	assert.Equal(t, risk_scoring.TrendIncreasing, trend.Direction)
	assert.Equal(t, 0.95, trend.Confidence)

	// Decreasing trend
	scores = []float64{70.0, 60.0, 50.0, 40.0, 30.0}
	trend = service.AnalyzeTrend(scores)
	assert.Equal(t, risk_scoring.TrendDecreasing, trend.Direction)

	// Stable trend
	scores = []float64{50.0, 51.0, 49.0, 50.0, 50.0}
	trend = service.AnalyzeTrend(scores)
	assert.Equal(t, risk_scoring.TrendStable, trend.Direction)

	// Insufficient data
	scores = []float64{50.0}
	trend = service.AnalyzeTrend(scores)
	assert.Equal(t, risk_scoring.TrendUnknown, trend.Direction)
}

// Test JSON serialization
func TestRiskScoreResultSerialization(t *testing.T) {
	result := risk_scoring.RiskScoreResult{
		Score:         75.0,
		RiskLevel:     "medium",
		Confidence:    0.85,
		ModelVersion:  "v1.0.0",
		CalculationTime: time.Now(),
		ModelResults:   []risk_scoring.ModelResult{
			{
				ModelName: "credit_model",
				Score:     75.0,
				Weight:    1.0,
			},
		},
	}

	// Serialize
	data, err := json.Marshal(result)
	require.NoError(t, err)

	// Deserialize
	var decoded risk_scoring.RiskScoreResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.Score, decoded.Score)
	assert.Equal(t, result.RiskLevel, decoded.RiskLevel)
	assert.Equal(t, result.Confidence, decoded.Confidence)
	assert.Len(t, decoded.ModelResults, 1)
}

// Test batch risk calculation
func TestServiceBatchCalculateRisk(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{
		EnableAggregation: true,
	})
	require.NoError(t, err)

	mockModel := new(MockRiskModel)
	mockModel.On("GetModelVersion").Return("v1.0.0")
	mockModel.On("CalculateScore", mock.Anything, mock.Anything).Return(50.0, nil)

	err = service.RegisterModel("batch_model", mockModel)
	require.NoError(t, err)

	batchData := []map[string]interface{}{
		{"entity": "entity1"},
		{"entity": "entity2"},
		{"entity": "entity3"},
	}

	results, err := service.BatchCalculateRisk(context.Background(), batchData)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	for _, result := range results {
		assert.Equal(t, 50.0, result.Score)
	}
}

// Test model removal
func TestServiceRemoveModel(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{})
	require.NoError(t, err)

	mockModel := new(MockRiskModel)
	mockModel.On("GetModelVersion").Return("v1.0.0")

	err = service.RegisterModel("removable_model", mockModel)
	require.NoError(t, err)

	err = service.RemoveModel("removable_model")
	require.NoError(t, err)

	models := service.GetRegisteredModels()
	assert.NotContains(t, models, "removable_model")
}

// Test statistics
func TestServiceGetStatistics(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{})
	require.NoError(t, err)

	mockModel := new(MockRiskModel)
	mockModel.On("GetModelVersion").Return("v1.0.0")
	mockModel.On("CalculateScore", mock.Anything, mock.Anything).Return(60.0, nil)

	err = service.RegisterModel("stats_model", mockModel)
	require.NoError(t, err)

	// Perform some calculations
	for i := 0; i < 10; i++ {
		service.CalculateRisk(context.Background(), map[string]interface{}{"value": i})
	}

	stats := service.GetStatistics()
	assert.Equal(t, int64(10), stats.TotalCalculations)
	assert.Len(t, stats.ModelsUsed, 1)
}

// Test cache functionality
func TestServiceCacheManagement(t *testing.T) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{
		CacheTTL: time.Minute,
	})
	require.NoError(t, err)

	mockModel := new(MockRiskModel)
	mockModel.On("GetModelVersion").Return("v1.0.0")
	mockModel.On("CalculateScore", mock.Anything, mock.Anything).Return(65.0, nil)

	err = service.RegisterModel("cache_model", mockModel)
	require.NoError(t, err)

	// First call - should hit model
	entityData := map[string]interface{}{"entity_id": "test_entity"}
	result1, err := service.CalculateRisk(context.Background(), entityData)
	require.NoError(t, err)

	// Second call with same data - should hit cache
	result2, err := service.CalculateRisk(context.Background(), entityData)
	require.NoError(t, err)

	// Results should be the same
	assert.Equal(t, result1.Score, result2.Score)

	// Check cache statistics
	cacheStats := service.GetCacheStats()
	assert.Equal(t, int64(1), cacheStats.Hits)
	assert.Equal(t, int64(1), cacheStats.Misses)
}

// Benchmark tests
func BenchmarkCalculateRisk(b *testing.B) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{
		MaxConcurrentCalculations: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		model := risk_scoring.NewSimpleRiskModel("model_" + string(rune('0'+i)))
		service.RegisterModel("model_"+string(rune('0'+i)), model)
	}

	ctx := context.Background()
	data := map[string]interface{}{
		"credit_score":    650,
		"income":          50000,
		"debt_to_income":  0.3,
		"payment_history": 0.95,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CalculateRisk(ctx, data)
	}
}

func BenchmarkBatchCalculateRisk(b *testing.B) {
	service, err := risk_scoring.NewRiskScoringService(risk_scoring.ServiceConfig{
		EnableAggregation: true,
	})
	if err != nil {
		b.Fatal(err)
	}

	model := risk_scoring.NewSimpleRiskModel("batch_model")
	service.RegisterModel("batch_model", model)

	ctx := context.Background()
	batch := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		batch[i] = map[string]interface{}{"entity": i}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.BatchCalculateRisk(ctx, batch)
	}
}
