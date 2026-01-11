package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"intervention-service/models"
	"intervention-service/repository"
)

// ScenarioService handles scenario simulation business logic
type ScenarioService struct {
	repo  *repository.PostgresRepository
	redis *redis.Client
}

// NewScenarioService creates a new scenario service
func NewScenarioService(repo *repository.PostgresRepository, redisClient *redis.Client) *ScenarioService {
	return &ScenarioService{
		repo:  repo,
		redis: redisClient,
	}
}

// CreateScenario creates a new scenario
func (s *ScenarioService) CreateScenario(ctx context.Context, scenario *models.Scenario) error {
	scenario.ID = uuid.New()
	scenario.Status = models.ScenarioStatusDraft
	scenario.CreatedAt = time.Now()
	scenario.UpdatedAt = time.Now()

	if err := s.validateScenario(scenario); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return s.repo.CreateScenario(ctx, scenario)
}

// GetScenario retrieves a scenario by ID
func (s *ScenarioService) GetScenario(ctx context.Context, id uuid.UUID) (*models.Scenario, error) {
	return s.repo.GetScenario(ctx, id)
}

// ListScenarios retrieves scenarios with filters
func (s *ScenarioService) ListScenarios(ctx context.Context, filters map[string]interface{}) ([]*models.Scenario, error) {
	return s.repo.ListScenarios(ctx, filters)
}

// SimulateScenario runs a scenario simulation
func (s *ScenarioService) SimulateScenario(ctx context.Context, scenario *models.Scenario) (*models.ScenarioResults, error) {
	scenario.ID = uuid.New()
	scenario.Status = models.ScenarioStatusRunning
	scenario.CreatedAt = time.Now()
	scenario.UpdatedAt = time.Now()

	startTime := time.Now()

	// Run simulation based on type
	var results *models.ScenarioResults
	var err error

	switch scenario.Type {
	case models.ScenarioTypeWhatIf:
		results, err = s.runWhatIfSimulation(ctx, scenario)
	case models.ScenarioTypeMonteCarlo:
		results, err = s.runMonteCarloSimulation(ctx, scenario)
	case models.ScenarioTypeSensitivity:
		results, err = s.runSensitivitySimulation(ctx, scenario)
	case models.ScenarioTypeImpact:
		results, err = s.runImpactSimulation(ctx, scenario)
	default:
		return nil, fmt.Errorf("unknown scenario type: %s", scenario.Type)
	}

	if err != nil {
		scenario.Status = models.ScenarioStatusFailed
		s.repo.UpdateScenario(ctx, scenario)
		return nil, err
	}

	scenario.Status = models.ScenarioStatusCompleted
	scenario.Results = results
	scenario.UpdatedAt = time.Now()

	if err := s.repo.CreateScenario(ctx, scenario); err != nil {
		return nil, err
	}

	return results, nil
}

// ListScenarioTemplates retrieves available scenario templates
func (s *ScenarioService) ListScenarioTemplates(ctx context.Context) ([]*models.ScenarioTemplate, error) {
	return s.repo.ListScenarioTemplates(ctx)
}

// CreateScenarioTemplate creates a new scenario template
func (s *ScenarioService) CreateScenarioTemplate(ctx context.Context, template *models.ScenarioTemplate) error {
	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	return s.repo.CreateScenarioTemplate(ctx, template)
}

// Simulation functions

func (s *ScenarioService) runWhatIfSimulation(ctx context.Context, scenario *models.Scenario) (*models.ScenarioResults, error) {
	results := &models.ScenarioResults{
		RunID:          uuid.New(),
		Status:         "completed",
		Summary:        models.ScenarioSummary{},
		Distribution:   make(map[string]interface{}),
		Sensitivity:    make(map[string]interface{}),
		Recommendations: []string{},
		CompletedAt:    time.Now(),
	}

	// Run what-if analysis
	iterations := scenario.Parameters.Iterations
	if iterations == 0 {
		iterations = 1000
	}

	// Generate simulation data
	values := make([]float64, iterations)
	for i := 0; i < iterations; i++ {
		values[i] = s.simulateWhatIfValue(scenario.Parameters)
	}

	results.Summary = s.calculateSummary(values)
	results.Distribution["values"] = values
	results.Recommendations = s.generateRecommendations(results.Summary)

	return results, nil
}

func (s *ScenarioService) runMonteCarloSimulation(ctx context.Context, scenario *models.Scenario) (*models.ScenarioResults, error) {
	results := &models.ScenarioResults{
		RunID:          uuid.New(),
		Status:         "completed",
		Summary:        models.ScenarioSummary{},
		Distribution:   make(map[string]interface{}),
		Sensitivity:    make(map[string]interface{}),
		Recommendations: []string{},
		CompletedAt:    time.Now(),
	}

	iterations := scenario.Parameters.Iterations
	if iterations == 0 {
		iterations = 10000
	}

	// Run Monte Carlo simulation
	values := make([]float64, iterations)
	for i := 0; i < iterations; i++ {
		values[i] = s.simulateMonteCarloValue(scenario.Parameters)
	}

	results.Summary = s.calculateSummary(values)
	results.Distribution["histogram"] = s.createHistogram(values, 20)
	results.Distribution["statistics"] = map[string]float64{
		"skewness": s.calculateSkewness(values),
		"kurtosis": s.calculateKurtosis(values),
	}
	results.Recommendations = s.generateMonteCarloRecommendations(results.Summary, scenario.Parameters.ConfidenceLevel)

	return results, nil
}

func (s *ScenarioService) runSensitivitySimulation(ctx context.Context, scenario *models.Scenario) (*models.ScenarioResults, error) {
	results := &models.ScenarioResults{
		RunID:          uuid.New(),
		Status:         "completed",
		Summary:        models.ScenarioSummary{},
		Distribution:   make(map[string]interface{}),
		Sensitivity:    make(map[string]interface{}),
		Recommendations: []string{},
		CompletedAt:    time.Now(),
	}

	// Run sensitivity analysis
	variableImpacts := make(map[string]float64)
	for _, variable := range scenario.Parameters.Variables {
		impact := s.calculateVariableImpact(scenario.Parameters, variable)
		variableImpacts[variable.Name] = impact
	}

	results.Sensitivity["variable_impacts"] = variableImpacts
	results.Summary = s.calculateBaseline(scenario.Parameters)
	results.Recommendations = s.generateSensitivityRecommendations(variableImpacts)

	return results, nil
}

func (s *ScenarioService) runImpactSimulation(ctx context.Context, scenario *models.Scenario) (*models.ScenarioResults, error) {
	results := &models.ScenarioResults{
		RunID:          uuid.New(),
		Status:         "completed",
		Summary:        models.ScenarioSummary{},
		Distribution:   make(map[string]interface{}),
		Sensitivity:    make(map[string]interface{}),
		Recommendations: []string{},
		CompletedAt:    time.Now(),
	}

	// Run impact assessment
	impact := s.calculateImpact(scenario.Parameters)
	results.Summary = impact
	results.Sensitivity["impact_factors"] = s.identifyImpactFactors(scenario.Parameters)
	results.Recommendations = s.generateImpactRecommendations(impact)

	return results, nil
}

// Helper functions

func (s *ScenarioService) validateScenario(scenario *models.Scenario) error {
	if scenario.Name == "" {
		return fmt.Errorf("scenario name is required")
	}
	if scenario.Type == "" {
		return fmt.Errorf("scenario type is required")
	}
	if len(scenario.Parameters.Variables) == 0 {
		return fmt.Errorf("at least one variable is required")
	}
	return nil
}

func (s *ScenarioService) simulateWhatIfValue(params models.ScenarioParameters) float64 {
	baseValue := 100.0
	for _, v := range params.Variables {
		baseValue += (v.CurrentValue - v.MinValue) * 0.1
	}
	return baseValue * (1 + (rand.Float64()-0.5)*0.1)
}

func (s *ScenarioService) simulateMonteCarloValue(params models.ScenarioParameters) float64 {
	baseValue := 100.0
	for _, v := range params.Variables {
		baseValue += s.sampleDistribution(v)
	}
	return baseValue * (1 + (rand.Float64()-0.5)*0.2)
}

func (s *ScenarioService) sampleDistribution(v models.ScenarioVariable) float64 {
	switch v.Distribution {
	case "normal":
		mean := (v.CurrentValue + v.MinValue + v.MaxValue) / 3
		stdDev := (v.MaxValue - v.MinValue) / 6
		return rand.NormFloat64()*stdDev + mean
	case "uniform":
		return v.MinValue + rand.Float64()*(v.MaxValue-v.MinValue)
	case "triangular":
		mode := v.CurrentValue
		u := rand.Float64()
		f := (mode - v.MinValue) / (v.MaxValue - v.MinValue)
		if u < f {
			return v.MinValue + math.Sqrt(u*(v.MaxValue-v.MinValue)*(mode-v.MinValue))
		}
		return v.MaxValue - math.Sqrt((1-u)*(v.MaxValue-v.MinValue)*(v.MaxValue-mode))
	default:
		return v.CurrentValue
	}
}

func (s *ScenarioService) calculateSummary(values []float64) models.ScenarioSummary {
	n := float64(len(values))
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / n

	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= n
	stdDev := math.Sqrt(variance)

	percentiles := make(map[int]float64)
	percentiles[5] = s.percentile(values, 0.05)
	percentiles[25] = s.percentile(values, 0.25)
	percentiles[50] = s.percentile(values, 0.50)
	percentiles[75] = s.percentile(values, 0.75)
	percentiles[95] = s.percentile(values, 0.95)

	return models.ScenarioSummary{
		Mean:        mean,
		Median:      s.percentile(values, 0.50),
		StdDev:      stdDev,
		Min:         values[0],
		Max:         values[len(values)-1],
		Percentiles: percentiles,
	}
}

func (s *ScenarioService) percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values)) * p)
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}

func (s *ScenarioService) createHistogram(values []float64, bins int) map[int]int {
	histogram := make(map[int]int)
	minVal := values[0]
	maxVal := values[len(values)-1]
	binWidth := (maxVal - minVal) / float64(bins)

	for _, v := range values {
		binIndex := int((v - minVal) / binWidth)
		if binIndex >= bins {
			binIndex = bins - 1
		}
		histogram[binIndex]++
	}

	return histogram
}

func (s *ScenarioService) calculateSkewness(values []float64) float64 {
	n := float64(len(values))
	mean := s.calculateMean(values)
	stdDev := s.calculateStdDev(values, mean)

	sum := 0.0
	for _, v := range values {
		sum += math.Pow((v-mean)/stdDev, 3)
	}

	return sum / n
}

func (s *ScenarioService) calculateKurtosis(values []float64) float64 {
	n := float64(len(values))
	mean := s.calculateMean(values)
	stdDev := s.calculateStdDev(values, mean)

	sum := 0.0
	for _, v := range values {
		sum += math.Pow((v-mean)/stdDev, 4)
	}

	return (sum / n) - 3
}

func (s *ScenarioService) calculateMean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (s *ScenarioService) calculateStdDev(values []float64, mean float64) float64 {
	n := float64(len(values))
	sum := 0.0
	for _, v := range values {
		sum += (v - mean) * (v - mean)
	}
	return math.Sqrt(sum / n)
}

func (s *ScenarioService) calculateVariableImpact(params models.ScenarioVariable, variable models.ScenarioVariable) float64 {
	range_ := variable.MaxValue - variable.MinValue
	if range_ == 0 {
		return 0
	}
	return (variable.CurrentValue - variable.MinValue) / range_
}

func (s *ScenarioService) calculateBaseline(params models.ScenarioParameters) models.ScenarioSummary {
	baseValue := 100.0
	for _, v := range params.Variables {
		baseValue += v.CurrentValue * 0.1
	}

	return models.ScenarioSummary{
		Mean:   baseValue,
		Median: baseValue,
		StdDev: baseValue * 0.1,
		Min:    baseValue * 0.8,
		Max:    baseValue * 1.2,
	}
}

func (s *ScenarioService) calculateImpact(params models.ScenarioParameters) models.ScenarioSummary {
	baseValue := 100.0
	for _, v := range params.Variables {
		baseValue += v.CurrentValue * 0.2
	}

	return models.ScenarioSummary{
		Mean:   baseValue,
		Median: baseValue,
		StdDev: baseValue * 0.15,
		Min:    baseValue * 0.7,
		Max:    baseValue * 1.3,
	}
}

func (s *ScenarioService) identifyImpactFactors(params models.ScenarioParameters) map[string]float64 {
	factors := make(map[string]float64)
	for _, v := range params.Variables {
		factors[v.Name] = (v.CurrentValue - v.MinValue) / (v.MaxValue - v.MinValue + 0.001)
	}
	return factors
}

func (s *ScenarioService) generateRecommendations(summary models.ScenarioSummary) []string {
	return []string{
		"Consider monitoring key indicators closely",
		"Review historical patterns for context",
		"Prepare contingency measures for outlier scenarios",
	}
}

func (s *ScenarioService) generateMonteCarloRecommendations(summary models.ScenarioSummary, confidenceLevel float64) []string {
	return []string{
		fmt.Sprintf("At %.0f%% confidence level, expect values between %.2f and %.2f", 
			confidenceLevel*100, summary.Percentiles[5], summary.Percentiles[95]),
		"High variance suggests multiple scenarios to consider",
		"Review distribution tails for extreme outcomes",
	}
}

func (s *ScenarioService) generateSensitivityRecommendations(variableImpacts map[string]float64) []string {
	var recommendations []string
	for name, impact := range variableImpacts {
		if impact > 0.7 {
			recommendations = append(recommendations, 
				fmt.Sprintf("High sensitivity to %s - prioritize monitoring", name))
		}
	}
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Low overall sensitivity to individual variables")
	}
	return recommendations
}

func (s *ScenarioService) generateImpactRecommendations(summary models.ScenarioSummary) []string {
	return []string{
		fmt.Sprintf("Expected impact range: %.2f - %.2f", summary.Min, summary.Max),
		"Consider phased implementation to monitor effects",
		"Establish key metrics to track actual vs projected impact",
	}
}

func (s *ScenarioService) GetScenarioResults(ctx context.Context, runID uuid.UUID) (*models.ScenarioResults, error) {
	return s.repo.GetScenarioResults(ctx, runID)
}
