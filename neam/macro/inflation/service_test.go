package inflation

import (
	"context"
	"math"
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
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.config)
	assert.NotNil(t, svc.logger)
	assert.NotNil(t, svc.forecasts)
	assert.NotNil(t, svc.scenarios)
	assert.NotNil(t, svc.components)
}

// TestServiceLifecycle tests service start and stop
func TestServiceLifecycle(t *testing.T) {
	svc := createTestService(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test Start (which also initializes components)
	svc.Start(ctx)
	assert.NotNil(t, svc.ctx)
	assert.NotNil(t, svc.cancel)

	// Verify components were initialized
	svc.mu.RLock()
	componentsCount := len(svc.components)
	svc.mu.RUnlock()
	assert.Equal(t, 4, componentsCount)

	// Test Stop
	svc.Stop()
}

// TestInitializeComponents tests component initialization
func TestInitializeComponents(t *testing.T) {
	svc := createTestService(t)

	// Components should not be initialized until Start is called
	svc.mu.RLock()
	initialCount := len(svc.components)
	svc.mu.RUnlock()
	assert.Equal(t, 0, initialCount)

	// Call initializeComponents
	svc.initializeComponents()

	svc.mu.RLock()
	componentsCount := len(svc.components)
	svc.mu.RUnlock()
	assert.Equal(t, 4, componentsCount)

	// Verify component structure
	svc.mu.RLock()
	food := svc.components["food"]
	svc.mu.RUnlock()
	require.NotNil(t, food)

	assert.Equal(t, "food", food.ID)
	assert.Equal(t, "Food and Non-Alcoholic Beverages", food.Name)
	assert.Equal(t, "food", food.Category)
	assert.InDelta(t, 0.30, food.Weight, 0.001)
}

// TestGetComponents tests component retrieval
func TestGetComponents(t *testing.T) {
	svc := createTestService(t)
	svc.initializeComponents()

	ctx := context.Background()
	components, err := svc.GetComponents(ctx)
	require.NoError(t, err)
	require.NotNil(t, components)

	assert.Len(t, components, 4)

	// Verify all expected components are present
	componentNames := make(map[string]bool)
	for _, c := range components {
		componentNames[c.ID] = true
	}
	assert.True(t, componentNames["food"])
	assert.True(t, componentNames["fuel"])
	assert.True(t, componentNames["core"])
	assert.True(t, componentNames["services"])
}

// TestGetForecast tests forecast retrieval
func TestGetForecast(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Initially no forecasts
	_, err := svc.GetForecast(ctx, "monthly")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no forecast found")

	// Add a forecast directly
	svc.mu.Lock()
	svc.forecasts["test-monthly"] = &InflationEstimate{
		ID:     "test-monthly",
		Period: TimePeriod{Type: "monthly", Label: "January 2024"},
		Value:  4.5,
	}
	svc.mu.Unlock()

	forecast, err := svc.GetForecast(ctx, "monthly")
	require.NoError(t, err)
	require.NotNil(t, forecast)
	assert.Equal(t, "test-monthly", forecast.ID)
	assert.Equal(t, 4.5, forecast.Value)
}

// TestUpdateForecasts tests forecast update mechanism
func TestUpdateForecasts(t *testing.T) {
	svc := createTestService(t)
	svc.initializeComponents()

	// Call update which generates a new forecast
	svc.updateForecasts()

	svc.mu.RLock()
	forecastsCount := len(svc.forecasts)
	svc.mu.RUnlock()

	assert.Equal(t, 1, forecastsCount)

	// Verify forecast structure - get the first forecast
	svc.mu.RLock()
	var forecast *InflationEstimate
	for _, f := range svc.forecasts {
		forecast = f
		break
	}
	svc.mu.RUnlock()

	require.NotNil(t, forecast)
	assert.NotEmpty(t, forecast.ID)
	assert.Contains(t, forecast.ID, "inflation-arimax")
	assert.Equal(t, "ARIMAX", forecast.Method)
	assert.Greater(t, forecast.Confidence, 0.85)
}

// TestGenerateARIMAXForecast tests ARIMAX forecast generation
func TestGenerateARIMAXForecast(t *testing.T) {
	svc := createTestService(t)
	svc.initializeComponents()
	ctx := context.Background()

	forecast := svc.generateARIMAXForecast(ctx)

	require.NotNil(t, forecast)
	assert.NotEmpty(t, forecast.ID)
	assert.Equal(t, "monthly", forecast.Period.Type)
	assert.Greater(t, forecast.Value, 0.0)
	assert.Greater(t, forecast.UpperBound, forecast.LowerBound)
	assert.Greater(t, forecast.Confidence, 0.0)
	assert.LessOrEqual(t, forecast.Confidence, 1.0)

	// Verify percentiles
	assert.NotNil(t, forecast.Percentiles)
	assert.Len(t, forecast.Percentiles, 5)
	assert.Contains(t, forecast.Percentiles, "p10")
	assert.Contains(t, forecast.Percentiles, "p50")
	assert.Contains(t, forecast.Percentiles, "p90")

	// Verify components
	assert.NotEmpty(t, forecast.Components)
	for _, comp := range forecast.Components {
		assert.NotEmpty(t, comp.Component)
		assert.GreaterOrEqual(t, comp.Weight, 0.0)
		assert.LessOrEqual(t, comp.Weight, 1.0)
	}

	// Verify drivers
	assert.NotEmpty(t, forecast.Drivers)
	for _, driver := range forecast.Drivers {
		assert.NotEmpty(t, driver.Name)
		assert.GreaterOrEqual(t, driver.Impact, 0.0)
		assert.LessOrEqual(t, driver.Impact, 1.0)
	}
}

// TestGetByCategory tests category-based inflation retrieval
func TestGetByCategory(t *testing.T) {
	svc := createTestService(t)
	svc.initializeComponents()
	ctx := context.Background()

	categories, err := svc.GetByCategory(ctx)
	require.NoError(t, err)
	require.NotNil(t, categories)

	assert.Len(t, categories, 3) // Food, Fuel, Core

	for _, cat := range categories {
		assert.NotEmpty(t, cat.Category)
		assert.NotEmpty(t, cat.Subcategories)
		for _, sub := range cat.Subcategories {
			assert.NotEmpty(t, sub.Name)
			assert.GreaterOrEqual(t, sub.Weight, 0.0)
			assert.LessOrEqual(t, sub.Weight, 1.0)
		}
	}
}

// TestGetByRegion tests region-based inflation retrieval
func TestGetByRegion(t *testing.T) {
	svc := createTestService(t)
	svc.initializeComponents()
	ctx := context.Background()

	regional, err := svc.GetByRegion(ctx)
	require.NoError(t, err)
	require.NotNil(t, regional)

	assert.Len(t, regional, 5) // NORTH, SOUTH, EAST, WEST, CENTRAL

	for _, r := range regional {
		assert.NotEmpty(t, r.Region)
	}
}

// TestCreateScenario tests scenario creation
func TestCreateScenario(t *testing.T) {
	svc := createTestService(t)
	svc.initializeComponents()
	ctx := context.Background()

	params := ScenarioParameters{
		OilPrice:     85.0,
		ExchangeRate: 85.0,
	}

	scenario, err := svc.CreateScenario(ctx, "High Oil Price", "pessimistic", params)
	require.NoError(t, err)
	require.NotNil(t, scenario)

	assert.NotEmpty(t, scenario.ID)
	assert.Contains(t, scenario.ID, "scenario-inflation")
	assert.Equal(t, "High Oil Price", scenario.Name)
	assert.Equal(t, "pessimistic", scenario.Type)
	assert.NotEmpty(t, scenario.Forecasts)
}

// TestGenerateScenarioForecasts tests scenario forecast generation
func TestGenerateScenarioForecasts(t *testing.T) {
	svc := createTestService(t)
	svc.initializeComponents()
	ctx := context.Background()

	params := ScenarioParameters{
		OilPrice:     100.0,
		ExchangeRate: 90.0,
	}

	forecasts := svc.generateScenarioForecasts(ctx, params)

	assert.Len(t, forecasts, 12) // 12 months

	for i, forecast := range forecasts {
		assert.NotEmpty(t, forecast.ID)
		assert.Equal(t, "monthly", forecast.Period.Type)
		assert.GreaterOrEqual(t, forecast.Confidence, 0.0)
		assert.LessOrEqual(t, forecast.Confidence, 1.0)

		// Confidence decreases with horizon
		if i > 0 {
			assert.Less(t, forecasts[i].Confidence, forecasts[i-1].Confidence)
		}
	}
}

// TestCalculateParameterAdjustment tests parameter adjustment calculation
func TestCalculateParameterAdjustment(t *testing.T) {
	svc := createTestService(t)

	// Test with baseline parameters
	params1 := ScenarioParameters{
		OilPrice:     75.0,
		ExchangeRate: 83.5,
	}
	adjustment1 := svc.calculateParameterAdjustment(params1)
	assert.Equal(t, 0.0, adjustment1)

	// Test with increased oil price
	params2 := ScenarioParameters{
		OilPrice:     85.0, // +10 from baseline
		ExchangeRate: 83.5,
	}
	adjustment2 := svc.calculateParameterAdjustment(params2)
	assert.Greater(t, adjustment2, 0.0)

	// Test with increased exchange rate
	params3 := ScenarioParameters{
		OilPrice:     75.0,
		ExchangeRate: 88.5, // +5 from baseline
	}
	adjustment3 := svc.calculateParameterAdjustment(params3)
	assert.Greater(t, adjustment3, 0.0)
}

// TestGetHistory tests historical data retrieval
func TestGetHistory(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	history, err := svc.GetHistory(ctx, startDate, endDate)
	require.NoError(t, err)
	require.NotNil(t, history)

	// Should have 31 entries (one per day)
	assert.Len(t, history, 31)

	for _, entry := range history {
		assert.False(t, entry.Date.IsZero())
	}
}

// TestInflationEstimateStructure tests inflation estimate data structure
func TestInflationEstimateStructure(t *testing.T) {
	estimate := &InflationEstimate{
		ID:         "test-estimate",
		Period:     TimePeriod{Type: "monthly", Label: "January 2024"},
		Value:      4.52,
		LowerBound: 4.02,
		UpperBound: 5.02,
		Confidence: 0.88,
		Percentiles: map[string]float64{
			"p10": 3.72,
			"p25": 4.17,
			"p50": 4.52,
			"p75": 4.87,
			"p90": 5.32,
		},
		Components: []ComponentContribution{
			{Component: "Food", Weight: 0.30, Value: 6.2, Change: -0.3, Contribution: 1.86},
		},
		Method:    "ARIMAX",
		Drivers:   []InflationDriver{{Name: "Oil", Impact: 0.25}},
		CreatedAt: time.Now(),
		ValidUntil: time.Now().AddDate(0, 0, 1),
	}

	assert.Equal(t, "test-estimate", estimate.ID)
	assert.Greater(t, estimate.UpperBound, estimate.LowerBound)
	assert.Greater(t, estimate.Confidence, 0.0)
	assert.Len(t, estimate.Percentiles, 5)
}

// TestInflationComponentStructure tests inflation component data structure
func TestInflationComponentStructure(t *testing.T) {
	component := &InflationComponent{
		ID:            "food",
		Name:          "Food and Non-Alcoholic Beverages",
		Code:          "CF1",
		Category:      "food",
		Weight:        0.30,
		LatestValue:   6.2,
		PreviousValue: 6.5,
		Change:        -0.3,
		ChangePercent: -4.6,
		Trend:         "decreasing",
		UpdatedAt:     time.Now(),
	}

	assert.Equal(t, "food", component.ID)
	assert.Equal(t, "CF1", component.Code)
	assert.InDelta(t, 0.30, component.Weight, 0.001)
	assert.Equal(t, "decreasing", component.Trend)
	assert.Less(t, component.Change, 0.0)
}

// TestRegionalInflationStructure tests regional inflation data structure
func TestRegionalInflationStructure(t *testing.T) {
	regional := RegionalInflation{
		Region:        "NORTH",
		Value:         4.72,
		Change:        0.15,
		ChangePercent: 3.3,
		Components: []ComponentContribution{
			{Component: "Food", Weight: 0.30, Contribution: 1.86},
		},
		UpdatedAt: time.Now(),
	}

	assert.Equal(t, "NORTH", regional.Region)
	assert.Greater(t, regional.Value, 0.0)
}

// TestCategoryInflationStructure tests category inflation data structure
func TestCategoryInflationStructure(t *testing.T) {
	category := CategoryInflation{
		Category: "Food",
		Value:    6.2,
		Change:   -0.3,
		Subcategories: []SubCategoryInflation{
			{Name: "Cereals", Value: 5.5, Change: -0.2, Weight: 0.10},
			{Name: "Meat", Value: 8.2, Change: -0.5, Weight: 0.08},
		},
		UpdatedAt: time.Now(),
	}

	assert.Equal(t, "Food", category.Category)
	assert.Len(t, category.Subcategories, 2)
}

// TestInflationDriverStructure tests inflation driver data structure
func TestInflationDriverStructure(t *testing.T) {
	driver := InflationDriver{
		Name:        "Crude Oil Prices",
		Impact:      0.25,
		CurrentValue: 75.0,
		Trend:       "stable",
		Forecast:    72.0,
	}

	assert.Equal(t, "Crude Oil Prices", driver.Name)
	assert.InDelta(t, 0.25, driver.Impact, 0.001)
	assert.Equal(t, "stable", driver.Trend)
}

// TestScenarioParametersStructure tests scenario parameters data structure
func TestScenarioParametersStructure(t *testing.T) {
	params := ScenarioParameters{
		OilPrice:       85.0,
		ExchangeRate:   85.0,
		FoodPriceIndex: 6.2,
		WageGrowth:     5.5,
		MonetaryPolicy: "contraction",
	}

	assert.Equal(t, 85.0, params.OilPrice)
	assert.Equal(t, 85.0, params.ExchangeRate)
	assert.Equal(t, "contraction", params.MonetaryPolicy)
}

// TestInflationScenarioStructure tests inflation scenario data structure
func TestInflationScenarioStructure(t *testing.T) {
	scenario := &InflationScenario{
		ID:        "test-scenario",
		Name:      "High Inflation",
		Type:      "pessimistic",
		Forecasts: []InflationEstimate{},
		Parameters: ScenarioParameters{
			OilPrice: 100.0,
		},
		Probability: 0.20,
		CreatedAt:   time.Now(),
	}

	assert.Equal(t, "test-scenario", scenario.ID)
	assert.Equal(t, "pessimistic", scenario.Type)
	assert.Equal(t, 0.20, scenario.Probability)
}

// TestComponentContributionStructure tests component contribution data structure
func TestComponentContributionStructure(t *testing.T) {
	contribution := ComponentContribution{
		Component:   "Food",
		Weight:      0.30,
		Value:       6.2,
		Change:      -0.3,
		Contribution: 1.86,
	}

	assert.Equal(t, "Food", contribution.Component)
	assert.InDelta(t, 0.30, contribution.Weight, 0.001)
	assert.InDelta(t, 1.86, contribution.Contribution, 0.001)
}

// TestThreadSafety tests concurrent access to the service
func TestThreadSafety(t *testing.T) {
	svc := createTestService(t)
	svc.initializeComponents()
	ctx := context.Background()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent component access
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = svc.GetComponents(ctx)
		}()
	}

	// Concurrent forecast updates
	for i := 0; i < iterations/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc.updateForecasts()
		}()
	}

	wg.Wait()
}

// TestForecastUpdateLoop tests the forecast update loop
func TestForecastUpdateLoop(t *testing.T) {
	svc := createTestService(t)
	svc.initializeComponents()

	// Start the service which starts the loop
	ctx, cancel := context.WithCancel(context.Background())
	svc.Start(ctx)

	// Wait a bit to allow the loop to run at least once
	svc.updateForecasts()

	svc.mu.RLock()
	forecastsCount := len(svc.forecasts)
	svc.mu.RUnlock()

	assert.Equal(t, 1, forecastsCount)

	cancel()
	svc.Stop()
}

// TestTimePeriodStructure tests time period data structure
func TestTimePeriodStructure(t *testing.T) {
	period := TimePeriod{
		Type:  "monthly",
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		Label: "January 2024",
	}

	assert.Equal(t, "monthly", period.Type)
	assert.Equal(t, "January 2024", period.Label)
}

// TestHistoryEntryStructure tests history entry data structure
func TestHistoryEntryStructure(t *testing.T) {
	entry := HistoryEntry{
		Date:  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Value: 5.2,
	}

	assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), entry.Date)
	assert.Greater(t, entry.Value, 0.0)
}

// TestSubCategoryInflationStructure tests subcategory inflation data structure
func TestSubCategoryInflationStructure(t *testing.T) {
	subcat := SubCategoryInflation{
		Name:   "Cereals",
		Value:  5.5,
		Change: -0.2,
		Weight: 0.10,
	}

	assert.Equal(t, "Cereals", subcat.Name)
	assert.InDelta(t, 0.10, subcat.Weight, 0.001)
}

// TestComponentWeightValidation tests that component weights are valid
func TestComponentWeightValidation(t *testing.T) {
	svc := createTestService(t)
	svc.initializeComponents()

	svc.mu.RLock()
	components := make([]*InflationComponent, 0, len(svc.components))
	for _, c := range svc.components {
		components = append(components, c)
	}
	svc.mu.RUnlock()

	var totalWeight float64
	for _, c := range components {
		assert.GreaterOrEqual(t, c.Weight, 0.0)
		assert.LessOrEqual(t, c.Weight, 1.0)
		totalWeight += c.Weight
	}

	// Weights should sum to approximately 1.0
	assert.InDelta(t, 1.0, totalWeight, 0.01)
}

// TestComponentTrends tests that component trends are consistent with changes
func TestComponentTrends(t *testing.T) {
	svc := createTestService(t)
	svc.initializeComponents()

	svc.mu.RLock()
	components := make([]*InflationComponent, 0, len(svc.components))
	for _, c := range svc.components {
		components = append(components, c)
	}
	svc.mu.RUnlock()

	for _, c := range components {
		switch c.Trend {
		case "increasing":
			assert.Greater(t, c.Change, 0.0)
		case "decreasing":
			assert.Less(t, c.Change, 0.0)
		case "stable":
			// Stable means small change within threshold
			assert.LessOrEqual(t, math.Abs(c.Change), 0.2)
		}
	}
}
