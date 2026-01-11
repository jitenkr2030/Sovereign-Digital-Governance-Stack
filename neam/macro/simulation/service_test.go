package simulation

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
		Redis:  nil,
		Logger: logger,
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
	assert.NotNil(t, svc.runs)
	assert.NotNil(t, svc.templates)
	assert.NotNil(t, svc.mu)
	assert.NotNil(t, svc.ctx)
	assert.NotNil(t, svc.cancel)
}

// TestServiceLifecycle tests service start and stop
func TestServiceLifecycle(t *testing.T) {
	svc := createTestService(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test Start
	svc.Start(ctx)
	assert.NotNil(t, svc.ctx)

	// Verify templates were initialized
	svc.mu.RLock()
	templatesCount := len(svc.templates)
	svc.mu.RUnlock()
	assert.Equal(t, 3, templatesCount) // gdp-stress-test, inflation-shock, oil-shock

	// Test Stop
	svc.Stop()
}

// TestInitializeTemplates tests template initialization
func TestInitializeTemplates(t *testing.T) {
	svc := createTestService(t)

	// Templates should not be initialized until NewService
	svc.mu.RLock()
	initialCount := len(svc.templates)
	svc.mu.RUnlock()
	assert.Equal(t, 3, initialCount)

	// Verify template structure
	svc.mu.RLock()
	gdpTemplate := svc.templates["gdp-stress-test"]
	svc.mu.RUnlock()
	require.NotNil(t, gdpTemplate)

	assert.Equal(t, "GDP Growth Stress Test", gdpTemplate.Name)
	assert.Equal(t, "gdp", gdpTemplate.Category)
	assert.NotEmpty(t, gdpTemplate.Variables)
	assert.NotEmpty(t, gdpTemplate.Scenarios)
}

// TestCreateRun tests simulation run creation
func TestCreateRun(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	params := SimulationParameters{
		Variables: []VariableConfig{
			{Name: "gdp_growth", Distribution: "normal", Mean: 5.6, StdDev: 1.2, Unit: "%"},
		},
		Horizon:   12,
		StartDate: time.Now(),
	}

	run, err := svc.CreateRun(ctx, "Test Run", "gdp-stress-test", 1000, params, "test-user")
	require.NoError(t, err)
	require.NotNil(t, run)

	assert.NotEmpty(t, run.ID)
	assert.Contains(t, run.ID, "sim-run")
	assert.Equal(t, "Test Run", run.Name)
	assert.Equal(t, "gdp-stress-test", run.TemplateID)
	assert.Equal(t, "pending", run.Status)
	assert.Equal(t, 1000, run.Iterations)
	assert.Equal(t, 0.0, run.Progress)
	assert.Equal(t, 0, run.CompletedIterations)
	assert.Equal(t, "test-user", run.CreatedBy)
}

// TestGetRun tests simulation run retrieval
func TestGetRun(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Initially no runs
	_, err := svc.GetRun(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulation run not found")

	// Create a run
	params := SimulationParameters{
		Variables: []VariableConfig{
			{Name: "test", Distribution: "normal", Mean: 5.0, StdDev: 1.0},
		},
	}
	run, err := svc.CreateRun(ctx, "Test", "gdp-stress-test", 100, params, "test")
	require.NoError(t, err)

	// Retrieve the run
	retrieved, err := svc.GetRun(ctx, run.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, run.ID, retrieved.ID)
	assert.Equal(t, "Test", retrieved.Name)
}

// TestListRuns tests listing simulation runs with filters
func TestListRuns(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Create some runs
	params := SimulationParameters{
		Variables: []VariableConfig{
			{Name: "test", Distribution: "normal", Mean: 5.0, StdDev: 1.0},
		},
	}

	run1, _ := svc.CreateRun(ctx, "Run 1", "gdp-stress-test", 100, params, "user1")
	run2, _ := svc.CreateRun(ctx, "Run 2", "inflation-shock", 100, params, "user1")
	run3, _ := svc.CreateRun(ctx, "Run 3", "gdp-stress-test", 100, params, "user2")

	// List all runs
	runs, err := svc.ListRuns(ctx, map[string]string{})
	require.NoError(t, err)
	assert.Len(t, runs, 3)

	// Verify the created runs exist
	assert.NotEmpty(t, run1.ID)
	assert.NotEmpty(t, run2.ID)
	assert.NotEmpty(t, run3.ID)

	// Filter by status
	runs, err = svc.ListRuns(ctx, map[string]string{"status": "pending"})
	require.NoError(t, err)
	assert.Len(t, runs, 3)

	// Filter by template_id
	runs, err = svc.ListRuns(ctx, map[string]string{"template_id": "gdp-stress-test"})
	require.NoError(t, err)
	assert.Len(t, runs, 2)

	// Filter by non-existent template
	runs, err = svc.ListRuns(ctx, map[string]string{"template_id": "non-existent"})
	require.NoError(t, err)
	assert.Len(t, runs, 0)
}

// TestCancelRun tests cancelling a simulation run
func TestCancelRun(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Create a run
	params := SimulationParameters{
		Variables: []VariableConfig{
			{Name: "test", Distribution: "normal", Mean: 5.0, StdDev: 1.0},
		},
	}
	run, err := svc.CreateRun(ctx, "To Cancel", "gdp-stress-test", 100, params, "test")
	require.NoError(t, err)

	// Cancel the run
	err = svc.CancelRun(ctx, run.ID)
	require.NoError(t, err)

	// Verify cancellation
	svc.mu.RLock()
	cancelledRun := svc.runs[run.ID]
	svc.mu.RUnlock()
	require.NotNil(t, cancelledRun)
	assert.Equal(t, "cancelled", cancelledRun.Status)

	// Try to cancel non-existent run
	err = svc.CancelRun(ctx, "non-existent")
	assert.Error(t, err)
}

// TestGetTemplates tests template retrieval
func TestGetTemplates(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	templates, err := svc.GetTemplates(ctx)
	require.NoError(t, err)
	require.NotNil(t, templates)

	assert.Len(t, templates, 3)

	// Verify template names
	templateNames := make(map[string]bool)
	for _, t := range templates {
		templateNames[t.Name] = true
	}
	assert.True(t, templateNames["GDP Growth Stress Test"])
	assert.True(t, templateNames["Inflation Shock Analysis"])
	assert.True(t, templateNames["Oil Price Shock Simulation"])
}

// TestGetTemplate tests specific template retrieval
func TestGetTemplate(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Get existing template
	template, err := svc.GetTemplate(ctx, "gdp-stress-test")
	require.NoError(t, err)
	require.NotNil(t, template)
	assert.Equal(t, "GDP Growth Stress Test", template.Name)
	assert.Equal(t, "gdp", template.Category)

	// Get non-existent template
	_, err = svc.GetTemplate(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")
}

// TestCreateTemplate tests custom template creation
func TestCreateTemplate(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	template := SimulationTemplate{
		Name:        "Custom Template",
		Description: "A custom simulation template",
		Category:    "custom",
		Parameters: SimulationParameters{
			Horizon:   24,
			StartDate: time.Now(),
		},
		Variables: []VariableConfig{
			{Name: "custom_var", Distribution: "normal", Mean: 10.0, StdDev: 2.0},
		},
	}

	created, err := svc.CreateTemplate(ctx, template)
	require.NoError(t, err)
	require.NotNil(t, created)

	assert.NotEmpty(t, created.ID)
	assert.Contains(t, created.ID, "template-")
	assert.Equal(t, "Custom Template", created.Name)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)

	// Verify template was stored
	retrieved, err := svc.GetTemplate(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
}

// TestRunMonteCarloSimulation tests Monte Carlo simulation execution
func TestRunMonteCarloSimulation(t *testing.T) {
	svc := createTestService(t)

	params := SimulationParameters{
		Variables: []VariableConfig{
			{Name: "gdp_growth", Distribution: "normal", Mean: 5.6, StdDev: 1.2, Unit: "%"},
			{Name: "inflation", Distribution: "normal", Mean: 4.5, StdDev: 0.8, Unit: "%"},
		},
		Horizon:   12,
		StartDate: time.Now(),
	}

	results := svc.runMonteCarloSimulation(params, 1000)

	require.NotNil(t, results)
	assert.NotNil(t, results.Variables)
	assert.NotNil(t, results.Statistics)
	assert.NotNil(t, results.Distributions)
	assert.NotNil(t, results.Percentiles)

	// Verify variable results
	assert.Contains(t, results.Variables, "gdp_growth")
	assert.Contains(t, results.Variables, "inflation")

	gdpResults := results.Variables["gdp_growth"]
	assert.Greater(t, gdpResults.Mean, 0.0)
	assert.GreaterOrEqual(t, gdpResults.Min, -10.0) // Reasonable bounds
	assert.LessOrEqual(t, gdpResults.Max, 20.0)

	// Verify distributions
	assert.Len(t, results.Distributions["gdp_growth"], 1000)
	assert.Len(t, results.Distributions["inflation"], 1000)

	// Verify statistics
	assert.Contains(t, results.Statistics, "correlation_gdp_inflation")
}

// TestGenerateSamples tests sample generation for different distributions
func TestGenerateSamples(t *testing.T) {
	svc := createTestService(t)
	count := 1000

	testCases := []struct {
		name     string
		config   VariableConfig
		minVal   float64
		maxVal   float64
		meanVal  float64
	}{
		{
			name:    "Normal Distribution",
			config:  VariableConfig{Name: "test", Distribution: "normal", Mean: 5.0, StdDev: 1.0},
			minVal:  1.0,
			maxVal:  9.0,
			meanVal: 5.0,
		},
		{
			name:    "Lognormal Distribution",
			config:  VariableConfig{Name: "test", Distribution: "lognormal", Mean: 2.0, StdDev: 0.5},
			minVal:  0.5,
			maxVal:  5.0,
			meanVal: 2.0,
		},
		{
			name:    "Uniform Distribution",
			config:  VariableConfig{Name: "test", Distribution: "uniform", Min: 1.0, Max: 10.0},
			minVal:  1.0,
			maxVal:  10.0,
			meanVal: 5.5,
		},
		{
			name:    "Triangular Distribution",
			config:  VariableConfig{Name: "test", Distribution: "triangular", Min: 0.0, Max: 10.0, Mode: 5.0},
			minVal:  0.0,
			maxVal:  10.0,
			meanVal: 5.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			samples := svc.generateSamples(tc.config, count)
			assert.Len(t, samples, count)

			// Calculate statistics
			var sum float64
			min := samples[0]
			max := samples[0]
			for _, s := range samples {
				sum += s
				if s < min {
					min = s
				}
				if s > max {
					max = s
				}
			}
			mean := sum / float64(count)

			// Verify bounds
			if tc.config.Min != 0 {
				assert.GreaterOrEqual(t, min, tc.minVal)
			}
			if tc.config.Max != 0 {
				assert.LessOrEqual(t, max, tc.maxVal)
			}

			// Mean should be approximately correct (with some tolerance)
			assert.InDelta(t, tc.meanVal, mean, 1.0)
		})
	}
}

// TestCalculateVariableResults tests variable statistics calculation
func TestCalculateVariableResults(t *testing.T) {
	svc := createTestService(t)

	// Test with known values
	samples := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

	results := svc.calculateVariableResults(samples)

	assert.Equal(t, 3.0, results.Mean)         // (1+2+3+4+5)/5 = 3.0
	assert.Equal(t, 1.0, results.Min)          // Min of samples
	assert.Equal(t, 5.0, results.Max)          // Max of samples
	assert.Equal(t, 3.0, results.Median)       // Middle value of sorted [1,2,3,4,5]
	assert.Len(t, results.Values, 5)

	// Calculate expected standard deviation (sqrt of mean of squared deviations)
	expectedStdDev := math.Sqrt(2.0) // ≈ 1.414
	assert.InDelta(t, expectedStdDev, results.StdDev, 0.1)
}

// TestCalculateCorrelation tests correlation calculation
func TestCalculateCorrelation(t *testing.T) {
	svc := createTestService(t)

	// Perfect positive correlation
	x1 := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	y1 := []float64{2.0, 4.0, 6.0, 8.0, 10.0}
	cor1 := svc.calculateCorrelation(x1, y1)
	assert.InDelta(t, 1.0, cor1, 0.001)

	// Perfect negative correlation
	x2 := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	y2 := []float64{5.0, 4.0, 3.0, 2.0, 1.0}
	cor2 := svc.calculateCorrelation(x2, y2)
	assert.InDelta(t, -1.0, cor2, 0.001)

	// No correlation
	x3 := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	y3 := []float64{1.0, 1.0, 1.0, 1.0, 1.0}
	cor3 := svc.calculateCorrelation(x3, y3)
	assert.Equal(t, 0.0, cor3)

	// Empty arrays
	cor4 := svc.calculateCorrelation([]float64{}, []float64{})
	assert.Equal(t, 0.0, cor4)
}

// TestExecuteSimulation tests simulation execution
func TestExecuteSimulation(t *testing.T) {
	svc := createTestService(t)

	params := SimulationParameters{
		Variables: []VariableConfig{
			{Name: "gdp_growth", Distribution: "normal", Mean: 5.6, StdDev: 1.2},
			{Name: "inflation", Distribution: "normal", Mean: 4.5, StdDev: 0.8},
		},
		Horizon:   12,
		StartDate: time.Now(),
	}

	run := &SimulationRun{
		ID:         "test-run",
		Name:       "Test Execution",
		TemplateID: "gdp-stress-test",
		Status:     "pending",
		Iterations: 100,
		Parameters: params,
	}

	// Execute simulation
	svc.executeSimulation(run)

	// Verify results
	assert.Equal(t, "completed", run.Status)
	assert.Equal(t, 100, run.CompletedIterations)
	assert.Equal(t, 100.0, run.Progress)
	assert.NotNil(t, run.Results)
	assert.NotNil(t, run.StartedAt)
	assert.NotNil(t, run.CompletedAt)
	assert.Greater(t, run.Duration, time.Duration(0))
}

// TestApplyScenarioWeights tests scenario weight application
func TestApplyScenarioWeights(t *testing.T) {
	svc := createTestService(t)

	params := SimulationParameters{
		Variables: []VariableConfig{
			{Name: "gdp_growth", Distribution: "normal", Mean: 5.6, StdDev: 1.2},
		},
	}

	run := &SimulationRun{
		ID:         "test-run",
		Name:       "Test",
		Parameters: params,
	}

	results := &SimulationResults{
		Variables: make(map[string]VariableResults),
	}

	applied := svc.applyScenarioWeights(run, results)
	assert.NotNil(t, applied)
}

// TestSimulationRunStructure tests simulation run data structure
func TestSimulationRunStructure(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	run := &SimulationRun{
		ID:                  "test-run",
		Name:                "Test Run",
		Description:         "A test simulation run",
		TemplateID:          "gdp-stress-test",
		Status:              "completed",
		Progress:            100.0,
		Iterations:          1000,
		CompletedIterations: 1000,
		Parameters: SimulationParameters{
			Variables: []VariableConfig{
				{Name: "gdp_growth", Distribution: "normal", Mean: 5.6, StdDev: 1.2},
			},
		},
		Results: &SimulationResults{
			Variables:   make(map[string]VariableResults),
			Statistics:  make(map[string]float64),
			Distributions: make(map[string][]float64),
			Percentiles: make(map[string]map[string]float64),
		},
		StartedAt:   &startTime,
		CompletedAt: &endTime,
		CreatedAt:   startTime.Add(-1 * time.Hour),
		CreatedBy:   "test-user",
		Duration:    1 * time.Hour,
	}

	assert.Equal(t, "test-run", run.ID)
	assert.Equal(t, "completed", run.Status)
	assert.Equal(t, 100.0, run.Progress)
	assert.Equal(t, 1000, run.CompletedIterations)
	assert.NotNil(t, run.Results)
	assert.NotNil(t, run.StartedAt)
	assert.NotNil(t, run.CompletedAt)
}

// TestSimulationParametersStructure tests simulation parameters data structure
func TestSimulationParametersStructure(t *testing.T) {
	params := SimulationParameters{
		Variables: []VariableConfig{
			{Name: "gdp_growth", Distribution: "normal", Mean: 5.6, StdDev: 1.2, Unit: "%"},
			{Name: "inflation", Distribution: "normal", Mean: 4.5, StdDev: 0.8, Unit: "%"},
		},
		Correlations: []CorrelationConfig{
			{Variable1: "gdp_growth", Variable2: "inflation", Correlation: -0.5},
		},
		Constraints: []ConstraintConfig{
			{Expression: "gdp_growth > -5", Description: "GDP cannot be negative", Type: "hard"},
		},
		Horizon:   24,
		StartDate: time.Now(),
	}

	assert.Len(t, params.Variables, 2)
	assert.Len(t, params.Correlations, 1)
	assert.Len(t, params.Constraints, 1)
	assert.Equal(t, 24, params.Horizon)
}

// TestVariableConfigStructure tests variable configuration data structure
func TestVariableConfigStructure(t *testing.T) {
	config := VariableConfig{
		Name:        "gdp_growth",
		Distribution: "normal",
		Mean:        5.6,
		StdDev:      1.2,
		Min:         -10.0,
		Max:         20.0,
		Mode:        0.0,
		Unit:        "%",
		Description: "Annual GDP growth rate",
	}

	assert.Equal(t, "gdp_growth", config.Name)
	assert.Equal(t, "normal", config.Distribution)
	assert.InDelta(t, 5.6, config.Mean, 0.001)
	assert.InDelta(t, 1.2, config.StdDev, 0.001)
	assert.Equal(t, "%", config.Unit)
}

// TestCorrelationConfigStructure tests correlation configuration data structure
func TestCorrelationConfigStructure(t *testing.T) {
	config := CorrelationConfig{
		Variable1:   "gdp_growth",
		Variable2:   "inflation",
		Correlation: -0.5,
	}

	assert.Equal(t, "gdp_growth", config.Variable1)
	assert.Equal(t, "inflation", config.Variable2)
	assert.GreaterOrEqual(t, config.Correlation, -1.0)
	assert.LessOrEqual(t, config.Correlation, 1.0)
}

// TestConstraintConfigStructure tests constraint configuration data structure
func TestConstraintConfigStructure(t *testing.T) {
	config := ConstraintConfig{
		Expression:  "gdp_growth > -5",
		Description: "GDP cannot be negative",
		Type:        "hard",
	}

	assert.Equal(t, "gdp_growth > -5", config.Expression)
	assert.Equal(t, "hard", config.Type)
}

// TestSimulationResultsStructure tests simulation results data structure
func TestSimulationResultsStructure(t *testing.T) {
	results := &SimulationResults{
		Variables: map[string]VariableResults{
			"gdp_growth": {Mean: 5.6, StdDev: 1.2, Min: 1.0, Max: 10.0, Median: 5.5},
		},
		Statistics: map[string]float64{
			"correlation": -0.5,
		},
		Distributions: map[string][]float64{
			"gdp_growth": {1.0, 2.0, 3.0, 4.0, 5.0},
		},
		Percentiles: map[string]map[string]float64{
			"gdp_growth": {"p10": 2.5, "p50": 5.5, "p90": 8.5},
		},
		Scenarios: []ScenarioResult{
			{Name: "Baseline", Probability: 0.5, Description: "No shocks", Outcomes: map[string]float64{"gdp_growth": 5.6}},
		},
		Metadata: map[string]interface{}{
			"iterations": 1000,
		},
	}

	assert.Contains(t, results.Variables, "gdp_growth")
	assert.Contains(t, results.Statistics, "correlation")
	assert.Len(t, results.Distributions["gdp_growth"], 5)
	assert.Contains(t, results.Percentiles, "gdp_growth")
	assert.Len(t, results.Scenarios, 1)
}

// TestVariableResultsStructure tests variable results data structure
func TestVariableResultsStructure(t *testing.T) {
	results := VariableResults{
		Mean:     5.6,
		StdDev:   1.2,
		Min:      1.0,
		Max:      10.0,
		Median:   5.5,
		Skewness: 0.1,
		Kurtosis: 3.0,
		Values:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
	}

	assert.InDelta(t, 5.6, results.Mean, 0.001)
	assert.InDelta(t, 1.2, results.StdDev, 0.001)
	assert.Equal(t, 1.0, results.Min)
	assert.Equal(t, 10.0, results.Max)
	assert.Len(t, results.Values, 5)
}

// TestScenarioResultStructure tests scenario result data structure
func TestScenarioResultStructure(t *testing.T) {
	result := ScenarioResult{
		Name:        "Mild Recession",
		Probability: 0.25,
		Description: "Short-term economic contraction",
		Outcomes: map[string]float64{
			"gdp_growth": 3.6,
			"inflation":  6.0,
		},
	}

	assert.Equal(t, "Mild Recession", result.Name)
	assert.InDelta(t, 0.25, result.Probability, 0.001)
	assert.Contains(t, result.Outcomes, "gdp_growth")
	assert.Contains(t, result.Outcomes, "inflation")
}

// TestSimulationTemplateStructure tests simulation template data structure
func TestSimulationTemplateStructure(t *testing.T) {
	template := &SimulationTemplate{
		ID:          "test-template",
		Name:        "Test Template",
		Description: "A test simulation template",
		Category:    "gdp",
		Parameters: SimulationParameters{
			Horizon:   24,
			StartDate: time.Now(),
		},
		Variables: []VariableConfig{
			{Name: "gdp_growth", Distribution: "normal", Mean: 5.6, StdDev: 1.2},
		},
		Scenarios: []ScenarioTemplate{
			{Name: "Baseline", Description: "No shocks", Probability: 0.5, Adjustments: []Adjustment{}},
		},
		CreatedAt: time.Now(),
		CreatedBy: "test-user",
		UpdatedAt: time.Now(),
	}

	assert.Equal(t, "test-template", template.ID)
	assert.Equal(t, "gdp", template.Category)
	assert.Len(t, template.Variables, 1)
	assert.Len(t, template.Scenarios, 1)
}

// TestScenarioTemplateStructure tests scenario template data structure
func TestScenarioTemplateStructure(t *testing.T) {
	template := ScenarioTemplate{
		Name:        "Severe Recession",
		Description: "Extended economic downturn",
		Adjustments: []Adjustment{
			{Variable: "gdp_growth", Type: "add", Value: -4.0},
		},
		Probability: 0.15,
	}

	assert.Equal(t, "Severe Recession", template.Name)
	assert.Len(t, template.Adjustments, 1)
	assert.InDelta(t, 0.15, template.Probability, 0.001)
}

// TestAdjustmentStructure tests adjustment data structure
func TestAdjustmentStructure(t *testing.T) {
	adjustment := Adjustment{
		Variable: "gdp_growth",
		Type:     "add",
		Value:    -4.0,
	}

	assert.Equal(t, "gdp_growth", adjustment.Variable)
	assert.Equal(t, "add", adjustment.Type)
	assert.InDelta(t, -4.0, adjustment.Value, 0.001)
}

// TestThreadSafety tests concurrent access to the service
func TestThreadSafety(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent run creation
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			params := SimulationParameters{
				Variables: []VariableConfig{
					{Name: "test", Distribution: "normal", Mean: 5.0, StdDev: 1.0},
				},
			}
			_, _ = svc.CreateRun(ctx, "Test", "gdp-stress-test", 100, params, "test")
		}()
	}

	// Concurrent template access
	for i := 0; i < iterations/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = svc.GetTemplates(ctx)
		}()
	}

	wg.Wait()
}

// TestProcessPendingSimulations tests processing pending simulations
func TestProcessPendingSimulations(t *testing.T) {
	svc := createTestService(t)
	ctx := context.Background()

	// Create a pending run
	params := SimulationParameters{
		Variables: []VariableConfig{
			{Name: "gdp_growth", Distribution: "normal", Mean: 5.6, StdDev: 1.2},
		},
	}
	run, _ := svc.CreateRun(ctx, "Pending Run", "gdp-stress-test", 100, params, "test")

	// Process pending simulations
	svc.processPendingSimulations()

	// Verify run was processed
	svc.mu.RLock()
	processedRun := svc.runs[run.ID]
	svc.mu.RUnlock()

	require.NotNil(t, processedRun)
	assert.Equal(t, "completed", processedRun.Status)
	assert.Equal(t, 100, processedRun.CompletedIterations)
}

// TestSimulationExecutionLoop tests the simulation execution loop
func TestSimulationExecutionLoop(t *testing.T) {
	svc := createTestService(t)

	// Start the service which starts the loop
	ctx, cancel := context.WithCancel(context.Background())
	svc.Start(ctx)

	// Process a simulation
	params := SimulationParameters{
		Variables: []VariableConfig{
			{Name: "test", Distribution: "normal", Mean: 5.0, StdDev: 1.0},
		},
	}
	run, _ := svc.CreateRun(ctx, "Loop Test", "gdp-stress-test", 100, params, "test")

	svc.processPendingSimulations()

	svc.mu.RLock()
	processedRun := svc.runs[run.ID]
	svc.mu.RUnlock()

	assert.Equal(t, "completed", processedRun.Status)

	cancel()
	svc.Stop()
}

// TestSumFloats tests the sumFloats utility function
func TestSumFloats(t *testing.T) {
	result := sumFloats([]float64{1.0, 2.0, 3.0, 4.0, 5.0})
	assert.Equal(t, 15.0, result)

	result = sumFloats([]float64{})
	assert.Equal(t, 0.0, result)

	result = sumFloats([]float64{-1.0, 1.0})
	assert.Equal(t, 0.0, result)
}

// TestDefaultDistribution tests the default distribution behavior
func TestDefaultDistribution(t *testing.T) {
	svc := createTestService(t)

	config := VariableConfig{
		Name:        "unknown",
		Distribution: "unknown_distribution", // Should default to normal
		Mean:        10.0,
		StdDev:      2.0,
	}

	samples := svc.generateSamples(config, 100)

	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))

	// Should still generate samples around the mean
	assert.InDelta(t, 10.0, mean, 2.0)
}
