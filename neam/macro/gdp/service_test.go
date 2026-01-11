package gdp

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"neam-platform/shared"
)

// TestHelper provides common test utilities
type TestHelper struct{}

// NewTestHelper creates a new test helper
func NewTestHelper() *TestHelper {
	return &TestHelper{}
}

// createTestService creates a service for testing
func createTestService(t *testing.T) *Service {
	t.Helper()

	logger := shared.NewLogger(&shared.LoggerConfig{
		Level:      "debug",
		Format:     "json",
		OutputPath: "stdout",
	})

	cfg := Config{
		PostgreSQL: &shared.PostgreSQL{},
		Redis:      nil,
		Logger:     logger,
	}

	svc := NewService(cfg)
	require.NotNil(t, svc)
	return svc
}

// TestNewService tests service creation
func TestNewService(t *testing.T) {
	svc := createTestService(t)
	require.NotNil(t, svc)
	assert.NotNil(t, svc.config)
	assert.NotNil(t, svc.logger)
	assert.NotNil(t, svc.forecasts)
	assert.NotNil(t, svc.scenarios)
	assert.NotNil(t, svc.models)
}

// TestServiceLifecycle tests service start and stop
func TestServiceLifecycle(t *testing.T) {
	svc := createTestService(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test Start
	svc.Start(ctx)
	assert.NotNil(t, svc.ctx)
	assert.NotNil(t, svc.cancel)

	// Test Stop
	svc.Stop()
	assert.NotNil(t, svc.cancel)
}

// TestVARModelForecast tests VAR model forecast generation
func TestVARModelForecast(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	varModel := &VARModel{logger: svc.logger}
	indicators := []GDPIndicator{
		{ID: "ipi", Name: "Industrial Production Index", Category: "production", LatestValue: 105.5, PreviousValue: 103.2, Change: 2.3, Trend: "up"},
		{ID: "retail", Name: "Retail Sales", Category: "expenditure", LatestValue: 108.2, PreviousValue: 106.5, Change: 1.7, Trend: "up"},
	}

	forecast, err := varModel.Forecast(ctx, indicators, 4)
	require.NoError(t, err)
	require.NotNil(t, forecast)

	assert.NotEmpty(t, forecast.ID)
	assert.Contains(t, forecast.ID, "gdp-var")
	assert.Equal(t, "probabilistic", forecast.Type)
	assert.Equal(t, "quarterly", forecast.Period.Type)
	assert.Greater(t, forecast.Value, 0.0)
	assert.Greater(t, forecast.Confidence, 0.0)
	assert.LessOrEqual(t, forecast.Confidence, 1.0)
	assert.NotNil(t, forecast.Percentiles)
	assert.Len(t, forecast.Percentiles, 5)
	assert.Contains(t, forecast.Percentiles, "p10")
	assert.Contains(t, forecast.Percentiles, "p50")
	assert.Contains(t, forecast.Percentiles, "p90")
}

// TestLSTMModelForecast tests LSTM model forecast generation
func TestLSTMModelForecast(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	lstmModel := &LSTMModel{logger: svc.logger}
	indicators := []GDPIndicator{
		{ID: "gdp", Name: "GDP", Category: "production", LatestValue: 100.0, PreviousValue: 98.0, Change: 2.0, Trend: "up"},
	}

	forecast, err := lstmModel.Forecast(ctx, indicators, 4)
	require.NoError(t, err)
	require.NotNil(t, forecast)

	assert.NotEmpty(t, forecast.ID)
	assert.Contains(t, forecast.ID, "gdp-lstm")
	assert.Equal(t, "LSTM", forecast.Method)
	assert.Greater(t, forecast.Confidence, 0.85)
}

// TestEnsembleModelForecast tests ensemble model forecast generation
func TestEnsembleModelForecast(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	ensembleModel := &EnsembleModel{
		logger:  svc.logger,
		models:  []Model{&VARModel{logger: svc.logger}, &LSTMModel{logger: svc.logger}},
		weights: map[string]float64{"VAR": 0.4, "LSTM": 0.6},
	}

	indicators := []GDPIndicator{
		{ID: "test", Name: "Test Indicator", Category: "production", LatestValue: 100.0, Trend: "up"},
	}

	forecast, err := ensembleModel.Forecast(ctx, indicators, 4)
	require.NoError(t, err)
	require.NotNil(t, forecast)

	assert.NotEmpty(t, forecast.ID)
	assert.Contains(t, forecast.ID, "gdp-ensemble")
	assert.Equal(t, "Ensemble", forecast.Method)
	assert.Greater(t, forecast.Confidence, 0.9)
	assert.GreaterOrEqual(t, forecast.Value, 0.0)
}

// TestModelInterface tests model interface compliance
func TestModelInterface(t *testing.T) {
	svc := createTestService(t)

	varModel := svc.models["var"].(*VARModel)
	lstmModel := svc.models["lstm"].(*LSTMModel)
	ensembleModel := svc.models["ensemble"].(*EnsembleModel)

	assert.Equal(t, "Vector Autoregression (VAR)", varModel.GetName())
	assert.Equal(t, "var", varModel.GetType())

	assert.Equal(t, "Long Short-Term Memory (LSTM)", lstmModel.GetName())
	assert.Equal(t, "lstm", lstmModel.GetType())

	assert.Equal(t, "Ensemble (VAR + LSTM)", ensembleModel.GetName())
	assert.Equal(t, "ensemble", ensembleModel.GetType())
}

// TestGetForecast tests forecast retrieval
func TestGetForecast(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Initially no forecasts
	_, err := svc.GetForecast(ctx, "quarterly")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no forecast found")

	// Add a forecast directly
	svc.mu.Lock()
	svc.forecasts["test-quarterly"] = &GDPEstimate{
		ID:    "test-quarterly",
		Period: TimePeriod{Type: "quarterly", Label: "Q1 2024"},
		Value: 5.6,
	}
	svc.mu.Unlock()

	forecast, err := svc.GetForecast(ctx, "quarterly")
	require.NoError(t, err)
	require.NotNil(t, forecast)
	assert.Equal(t, "test-quarterly", forecast.ID)
}

// TestGetScenario tests scenario retrieval
func TestGetScenario(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Initially no scenarios
	_, err := svc.GetScenario(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scenario not found")

	// Add a scenario directly
	svc.mu.Lock()
	scenario := &GDPScenario{
		ID:   "test-scenario",
		Name: "Test Scenario",
		Type: "baseline",
	}
	svc.scenarios["test-scenario"] = scenario
	svc.mu.Unlock()

	retrieved, err := svc.GetScenario(ctx, "test-scenario")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, "test-scenario", retrieved.ID)
	assert.Equal(t, "Test Scenario", retrieved.Name)
}

// TestCreateScenario tests scenario creation
func TestCreateScenario(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	params := ScenarioParameters{
		InterestRate:       5.0,
		TaxRate:            25.0,
		GovernmentSpending: 100.0,
	}

	scenario, err := svc.CreateScenario(ctx, "New Scenario", "optimistic", params)
	require.NoError(t, err)
	require.NotNil(t, scenario)

	assert.NotEmpty(t, scenario.ID)
	assert.Contains(t, scenario.ID, "scenario-gdp")
	assert.Equal(t, "New Scenario", scenario.Name)
	assert.Equal(t, "optimistic", scenario.Type)
	assert.Equal(t, "draft", scenario.Status)
	assert.Equal(t, params.InterestRate, scenario.Parameters.InterestRate)
	assert.Equal(t, params.TaxRate, scenario.Parameters.TaxRate)
	assert.NotZero(t, scenario.CreatedAt)
}

// TestGetIndicators tests indicator retrieval
func TestGetIndicators(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	indicators, err := svc.GetIndicators(ctx)
	require.NoError(t, err)
	require.NotNil(t, indicators)

	assert.NotEmpty(t, indicators)
	assert.GreaterOrEqual(t, len(indicators), 3)

	// Verify indicator structure
	for _, ind := range indicators {
		assert.NotEmpty(t, ind.ID)
		assert.NotEmpty(t, ind.Name)
		assert.NotEmpty(t, ind.Category)
		assert.NotEmpty(t, ind.Trend)
	}
}

// TestGetSectorContributions tests sector contribution retrieval
func TestGetSectorContributions(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	contributions, err := svc.GetSectorContributions(ctx)
	require.NoError(t, err)
	require.NotNil(t, contributions)

	assert.NotEmpty(t, contributions)

	// Verify sector structure
	for _, contrib := range contributions {
		assert.NotEmpty(t, contrib.Sector)
		assert.NotEmpty(t, contrib.Code)
		assert.GreaterOrEqual(t, contrib.Weight, 0.0)
		assert.LessOrEqual(t, contrib.Weight, 1.0)
	}

	// Verify weights sum to approximately 1.0
	var totalWeight float64
	for _, contrib := range contributions {
		totalWeight += contrib.Weight
	}
	assert.InDelta(t, 1.0, totalWeight, 0.01)
}

// TestSensitivityAnalysis tests sensitivity analysis
func TestSensitivityAnalysis(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	results, err := svc.SensitivityAnalysis(ctx, "interest_rate", 10.0)
	require.NoError(t, err)
	assert.NotNil(t, results) // Method returns empty slice, not nil
	assert.Equal(t, 0, len(results)) // Verify empty results
}

// TestThreadSafety tests concurrent access to the service
func TestThreadSafety(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent forecast access
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = svc.GetIndicators(ctx)
		}()
	}

	// Concurrent scenario creation
	for i := 0; i < iterations/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			params := ScenarioParameters{InterestRate: 5.0}
			_, _ = svc.CreateScenario(ctx, "Test", "baseline", params)
		}()
	}

	wg.Wait()
	// If we get here without panics, the concurrent access is thread-safe
}

// TestGDPEstimateStructure tests GDP estimate data structure
func TestGDPEstimateStructure(t *testing.T) {
	estimate := &GDPEstimate{
		ID:          "test-estimate",
		Type:        "probabilistic",
		Period:      TimePeriod{Type: "annual", Start: time.Now(), End: time.Now().AddDate(1, 0, 0), Label: "2024"},
		Value:       5.6,
		LowerBound:  5.1,
		UpperBound:  6.1,
		Confidence:  0.85,
		Percentiles: map[string]float64{"p10": 4.8, "p25": 5.2, "p50": 5.6, "p75": 6.0, "p90": 6.4},
		Method:      "VAR",
		Contributors: []Contributor{
			{Indicator: "Industrial Production", Weight: 0.2, CurrentValue: 105.5, Direction: "up"},
		},
		Assumptions: []string{"Historical trends continue"},
		CreatedAt:   time.Now(),
		ValidUntil:  time.Now().AddDate(0, 0, 7),
	}

	assert.Equal(t, "test-estimate", estimate.ID)
	assert.Equal(t, "probabilistic", estimate.Type)
	assert.Greater(t, estimate.UpperBound, estimate.LowerBound)
	assert.Greater(t, estimate.Confidence, 0.0)
	assert.LessOrEqual(t, estimate.Confidence, 1.0)
	assert.Len(t, estimate.Percentiles, 5)
	assert.Len(t, estimate.Contributors, 1)
	assert.Len(t, estimate.Assumptions, 1)
}

// TestGDPScenarioStructure tests GDP scenario data structure
func TestGDPScenarioStructure(t *testing.T) {
	scenario := &GDPScenario{
		ID:          "test-scenario",
		Name:        "Optimistic Growth",
		Description: "Higher than expected growth scenario",
		Type:        "optimistic",
		Parameters: ScenarioParameters{
			InterestRate:       4.5,
			TaxRate:            22.0,
			GovernmentSpending: 120.0,
			ExportGrowth:       8.0,
			ImportGrowth:       6.0,
			InvestmentGrowth:   10.0,
			ConsumerSpending:   7.0,
			OilPrice:           70.0,
			ExchangeRate:       82.0,
		},
		Probability: 0.25,
		Status:      "active",
		CreatedAt:   time.Now(),
		CreatedBy:   "test-user",
	}

	assert.Equal(t, "test-scenario", scenario.ID)
	assert.Equal(t, "optimistic", scenario.Type)
	assert.Equal(t, 0.25, scenario.Probability)
	assert.Equal(t, "active", scenario.Status)
	assert.Equal(t, 4.5, scenario.Parameters.InterestRate)
}

// TestTimePeriodStructure tests time period data structure
func TestTimePeriodStructure(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	period := TimePeriod{
		Type:  "annual",
		Start: start,
		End:   end,
		Label: "2024",
	}

	assert.Equal(t, "annual", period.Type)
	assert.Equal(t, start, period.Start)
	assert.Equal(t, end, period.End)
	assert.Equal(t, "2024", period.Label)
}

// TestContributorStructure tests contributor data structure
func TestContributorStructure(t *testing.T) {
	contributor := Contributor{
		Indicator:       "Industrial Production Index",
		Weight:          0.25,
		CurrentValue:    105.5,
		ForecastedImpact: 0.5,
		Direction:       "positive",
	}

	assert.Equal(t, "Industrial Production Index", contributor.Indicator)
	assert.InDelta(t, 0.25, contributor.Weight, 0.001)
	assert.Equal(t, "positive", contributor.Direction)
}

// TestSectorContributionStructure tests sector contribution data structure
func TestSectorContributionStructure(t *testing.T) {
	contribution := SectorContribution{
		Sector:               "Manufacturing",
		Code:                 "C",
		Weight:               0.22,
		CurrentContribution:  0.22,
		PreviousContribution: 0.21,
		Change:               0.01,
		GrowthRate:           5.8,
	}

	assert.Equal(t, "Manufacturing", contribution.Sector)
	assert.Equal(t, "C", contribution.Code)
	assert.InDelta(t, 0.22, contribution.Weight, 0.001)
	assert.Greater(t, contribution.GrowthRate, 0.0)
}

// TestScenarioParametersStructure tests scenario parameters data structure
func TestScenarioParametersStructure(t *testing.T) {
	params := ScenarioParameters{
		InterestRate:       5.0,
		TaxRate:            25.0,
		GovernmentSpending: 100.0,
		ExportGrowth:       5.0,
		ImportGrowth:       4.0,
		InvestmentGrowth:   6.0,
		ConsumerSpending:   5.5,
		OilPrice:           75.0,
		ExchangeRate:       83.5,
	}

	assert.Equal(t, 5.0, params.InterestRate)
	assert.Equal(t, 25.0, params.TaxRate)
	assert.Equal(t, 100.0, params.GovernmentSpending)
	assert.Equal(t, 75.0, params.OilPrice)
	assert.Equal(t, 83.5, params.ExchangeRate)
}

// TestForecastValidation tests forecast validation logic
func TestForecastValidation(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Create a scenario which generates a forecast
	params := ScenarioParameters{InterestRate: 5.0}
	scenario, err := svc.CreateScenario(ctx, "Validation Test", "pessimistic", params)
	require.NoError(t, err)
	require.NotNil(t, scenario)

	// Verify forecast was generated
	assert.NotEmpty(t, scenario.Forecasts)
}

// TestIndicatorsTrend tests trend calculation in indicators
func TestIndicatorsTrend(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	indicators, err := svc.GetIndicators(ctx)
	require.NoError(t, err)

	for _, ind := range indicators {
		switch ind.Trend {
		case "up":
			assert.Greater(t, ind.Change, 0.0)
		case "down":
			assert.Less(t, ind.Change, 0.0)
		case "stable":
			assert.Equal(t, 0.0, ind.Change)
		}
	}
}

// TestApplyScenarioParameters tests scenario parameter application
func TestApplyScenarioParameters(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	indicators := []GDPIndicator{
		{ID: "test", Name: "Test", Category: "production", LatestValue: 100.0},
	}

	params := ScenarioParameters{
		InterestRate:       3.0,
		GovernmentSpending: 150.0,
	}

	result := svc.applyScenarioParameters(ctx, indicators, params)
	assert.Equal(t, indicators, result) // Currently returns unchanged
}

// TestUpdateForecasts tests the forecast update mechanism
func TestUpdateForecasts(t *testing.T) {
	svc := createTestService(t)

	// Call update which generates a new forecast
	svc.updateForecasts()

	svc.mu.RLock()
	forecastsCount := len(svc.forecasts)
	svc.mu.RUnlock()

	assert.Equal(t, 1, forecastsCount)
}

// TestServiceInitialization verifies service is properly initialized
func TestServiceInitialization(t *testing.T) {
	svc := createTestService(t)

	assert.NotNil(t, svc.forecasts)
	assert.NotNil(t, svc.scenarios)
	assert.NotNil(t, svc.models)
	assert.NotNil(t, svc.mu)
	assert.NotNil(t, svc.ctx)
	assert.NotNil(t, svc.cancel)

	// Verify models are registered
	assert.Contains(t, svc.models, "var")
	assert.Contains(t, svc.models, "lstm")
	assert.Contains(t, svc.models, "ensemble")
}

// TestEnsembleModelWeightedAverage tests ensemble model weighted average calculation
func TestEnsembleModelWeightedAverage(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Create ensemble with known weights
	ensembleModel := &EnsembleModel{
		logger: svc.logger,
		models: []Model{
			&mockForecaster{value: 5.0},
			&mockForecaster{value: 7.0},
		},
		weights: map[string]float64{"Mock1": 0.5, "Mock2": 0.5},
	}

	indicators := []GDPIndicator{
		{ID: "test", Name: "Test", Category: "production", LatestValue: 100.0},
	}

	forecast, err := ensembleModel.Forecast(ctx, indicators, 4)
	require.NoError(t, err)
	require.NotNil(t, forecast)

	// Weighted average of 5.0 and 7.0 with equal weights should be 6.0
	assert.InDelta(t, 6.0, forecast.Value, 0.1)
}

// mockForecaster is a mock model for testing ensemble calculations
type mockForecaster struct {
	value float64
}

func (m *mockForecaster) Forecast(ctx context.Context, indicators []GDPIndicator, horizon int) (*GDPEstimate, error) {
	return &GDPEstimate{
		ID:          "mock",
		Type:        "probabilistic",
		Period:      TimePeriod{Type: "quarterly"},
		Value:       m.value,
		Confidence:  0.9,
		Method:      "Mock1",
		CreatedAt:   time.Now(),
		ValidUntil:  time.Now().AddDate(0, 0, 7),
	}, nil
}

func (m *mockForecaster) GetName() string {
	return "Mock Forecaster"
}

func (m *mockForecaster) GetType() string {
	return "mock"
}
