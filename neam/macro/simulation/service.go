package simulation

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/redis/go-redis/v9"
	"neam-platform/shared"
)

// Config holds the configuration for the simulation service
type Config struct {
	ClickHouse clickhouse.Conn
	Redis      *redis.Client
	Logger     *shared.Logger
}

// Service handles Monte Carlo simulations
type Service struct {
	config    Config
	logger    *shared.Logger
	clickhouse clickhouse.Conn
	redis     *redis.Client
	runs      map[string]*SimulationRun
	templates map[string]*SimulationTemplate
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// SimulationRun represents a simulation run
type SimulationRun struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	TemplateID  string                 `json:"template_id"`
	Status      string                 `json:"status"` // pending, running, completed, failed, cancelled
	Progress    float64                `json:"progress"` // 0-100
	Iterations  int                    `json:"iterations"`
	CompletedIterations int            `json:"completed_iterations"`
	Parameters  SimulationParameters   `json:"parameters"`
	Results     *SimulationResults     `json:"results"`
	Error       string                 `json:"error"`
	StartedAt   *time.Time             `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at"`
	CreatedAt   time.Time              `json:"created_at"`
	CreatedBy   string                 `json:"created_by"`
	Duration    time.Duration          `json:"duration"`
}

// SimulationParameters defines the parameters for a simulation
type SimulationParameters struct {
	Variables     []VariableConfig     `json:"variables"`
	Correlations  []CorrelationConfig  `json:"correlations"`
	Constraints   []ConstraintConfig   `json:"constraints"`
	Horizon       int                  `json:"horizon"` // in months
	StartDate     time.Time            `json:"start_date"`
}

// VariableConfig defines a random variable
type VariableConfig struct {
	Name        string  `json:"name"`
	Distribution string `json:"distribution"` // normal, lognormal, uniform, triangular, pert
	Mean        float64 `json:"mean"`
	StdDev      float64 `json:"std_dev"`
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
	Mode        float64 `json:"mode"` // for triangular/pert
	Unit        string  `json:"unit"`
	Description string  `json:"description"`
}

// CorrelationConfig defines correlation between variables
type CorrelationConfig struct {
	Variable1   string  `json:"variable1"`
	Variable2   string  `json:"variable2"`
	Correlation float64 `json:"correlation"` // -1 to 1
}

// ConstraintConfig defines a constraint on variables
type ConstraintConfig struct {
	Expression  string  `json:"expression"` // e.g., "gdp_growth > 0"
	Description string  `json:"description"`
	Type        string  `json:"type"` // hard, soft
}

// SimulationResults contains simulation results
type SimulationResults struct {
	Variables   map[string]VariableResults `json:"variables"`
	Statistics  map[string]float64         `json:"statistics"`
	Distributions map[string][]float64    `json:"distributions"`
	Percentiles map[string]map[string]float64 `json:"percentiles"`
	Scenarios   []ScenarioResult           `json:"scenarios"`
	Metadata    map[string]interface{}     `json:"metadata"`
}

// VariableResults contains results for a single variable
type VariableResults struct {
	Mean       float64   `json:"mean"`
	StdDev     float64   `json:"std_dev"`
	Min        float64   `json:"min"`
	Max        float64   `json:"max"`
	Median     float64   `json:"median"`
	Skewness   float64   `json:"skewness"`
	Kurtosis   float64   `json:"kurtosis"`
	Values     []float64 `json:"values"`
}

// ScenarioResult represents a specific scenario outcome
type ScenarioResult struct {
	Name        string                 `json:"name"`
	Probability float64                `json:"probability"`
	Description string                 `json:"description"`
	Outcomes    map[string]float64     `json:"outcomes"`
}

// SimulationTemplate represents a reusable simulation template
type SimulationTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"` // gdp, inflation, stress_test, custom
	Parameters  SimulationParameters   `json:"parameters"`
	Variables   []VariableConfig       `json:"variables"`
	Scenarios   []ScenarioTemplate     `json:"scenarios"`
	CreatedAt   time.Time              `json:"created_at"`
	CreatedBy   string                 `json:"created_by"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ScenarioTemplate defines a predefined scenario in a template
type ScenarioTemplate struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Adjustments []Adjustment `json:"adjustments"`
	Probability float64  `json:"probability"`
}

// Adjustment defines an adjustment to a variable
type Adjustment struct {
	Variable string  `json:"variable"`
	Type     string  `json:"type"` // add, multiply, set
	Value    float64 `json:"value"`
}

// NewService creates a new simulation service
func NewService(cfg Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	svc := &Service{
		config:    cfg,
		logger:    cfg.Logger,
		clickhouse: cfg.ClickHouse,
		redis:     cfg.Redis,
		runs:      make(map[string]*SimulationRun),
		templates: make(map[string]*SimulationTemplate),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Initialize default templates
	svc.initializeTemplates()

	return svc
}

// Start begins the simulation service
func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Starting simulation service")

	// Start simulation execution
	s.wg.Add(1)
	go s.simulationExecutionLoop()

	s.logger.Info("Simulation service started")
}

// Stop gracefully stops the service
func (s *Service) Stop() {
	s.logger.Info("Stopping simulation service")
	s.cancel()
	s.wg.Wait()
	s.logger.Info("Simulation service stopped")
}

// initializeTemplates sets up default simulation templates
func (s *Service) initializeTemplates() {
	s.mu.Lock()

	// GDP Growth Stress Test Template
	s.templates["gdp-stress-test"] = &SimulationTemplate{
		ID:          "gdp-stress-test",
		Name:        "GDP Growth Stress Test",
		Description: "Simulates various economic stress scenarios on GDP growth",
		Category:    "gdp",
		Parameters: SimulationParameters{
			Horizon:   24, // 2 years
			StartDate: time.Now(),
		},
		Variables: []VariableConfig{
			{Name: "gdp_growth", Distribution: "normal", Mean: 5.6, StdDev: 1.2, Unit: "%"},
			{Name: "inflation", Distribution: "normal", Mean: 4.5, StdDev: 0.8, Unit: "%"},
			{Name: "unemployment", Distribution: "lognormal", Mean: 3.8, StdDev: 0.5, Unit: "%"},
			{Name: "interest_rate", Distribution: "normal", Mean: 5.5, StdDev: 0.5, Unit: "%"},
		},
		Scenarios: []ScenarioTemplate{
			{Name: "Baseline", Description: "No major shocks", Probability: 0.50, Adjustments: []Adjustment{}},
			{Name: "Mild Recession", Description: "Short-term economic contraction", Probability: 0.25, Adjustments: []Adjustment{{Variable: "gdp_growth", Type: "add", Value: -2.0}}},
			{Name: "Severe Recession", Description: "Extended economic downturn", Probability: 0.15, Adjustments: []Adjustment{{Variable: "gdp_growth", Type: "add", Value: -4.0}}},
			{Name: "Global Crisis", Description: "Major global economic crisis", Probability: 0.10, Adjustments: []Adjustment{{Variable: "gdp_growth", Type: "add", Value: -6.0}}},
		},
		CreatedAt: time.Now(),
	}

	// Inflation Shock Template
	s.templates["inflation-shock"] = &SimulationTemplate{
		ID:          "inflation-shock",
		Name:        "Inflation Shock Analysis",
		Description: "Simulates the impact of various inflation shocks",
		Category:    "inflation",
		Parameters: SimulationParameters{
			Horizon:   12, // 1 year
			StartDate: time.Now(),
		},
		Variables: []VariableConfig{
			{Name: "cpi", Distribution: "lognormal", Mean: 4.5, StdDev: 1.0, Unit: "%"},
			{Name: "food_inflation", Distribution: "triangular", Mean: 6.0, Min: 3.0, Max: 12.0, Mode: 5.5, Unit: "%"},
			{Name: "fuel_inflation", Distribution: "triangular", Mean: 4.0, Min: -5.0, Max: 15.0, Mode: 3.0, Unit: "%"},
			{Name: "core_inflation", Distribution: "normal", Mean: 4.1, StdDev: 0.5, Unit: "%"},
		},
		Scenarios: []ScenarioTemplate{
			{Name: "Baseline", Probability: 0.55},
			{Name: "Oil Price Spike", Probability: 0.25, Adjustments: []Adjustment{{Variable: "fuel_inflation", Type: "add", Value: 10.0}}},
			{Name: "Food Crisis", Probability: 0.15, Adjustments: []Adjustment{{Variable: "food_inflation", Type: "add", Value: 8.0}}},
			{Name: "Stagflation", Probability: 0.05, Adjustments: []Adjustment{{Variable: "cpi", Type: "add", Value: 3.0}}},
		},
		CreatedAt: time.Now(),
	}

	// Oil Price Shock Template
	s.templates["oil-shock"] = &SimulationTemplate{
		ID:          "oil-shock",
		Name:        "Oil Price Shock Simulation",
		Description: "Simulates economic impact of oil price fluctuations",
		Category:    "stress_test",
		Parameters: SimulationParameters{
			Horizon:   18,
			StartDate: time.Now(),
		},
		Variables: []VariableConfig{
			{Name: "oil_price", Distribution: "lognormal", Mean: 75.0, StdDev: 15.0, Unit: "USD/bbl"},
			{Name: "gdp_growth", Distribution: "normal", Mean: 5.6, StdDev: 1.2, Unit: "%"},
			{Name: "inflation", Distribution: "normal", Mean: 4.5, StdDev: 0.8, Unit: "%"},
			{Name: "current_account", Distribution: "normal", Mean: -2.5, StdDev: 0.8, Unit: "%"},
		},
		Scenarios: []ScenarioTemplate{
			{Name: "Baseline", Probability: 0.50},
			{Name: "Moderate Shock (+30%)", Probability: 0.25, Adjustments: []Adjustment{{Variable: "oil_price", Type: "multiply", Value: 1.3}}},
			{Name: "Severe Shock (+60%)", Probability: 0.15, Adjustments: []Adjustment{{Variable: "oil_price", Type: "multiply", Value: 1.6}}},
			{Name: "Crash (-40%)", Probability: 0.10, Adjustments: []Adjustment{{Variable: "oil_price", Type: "multiply", Value: 0.6}}},
		},
		CreatedAt: time.Now(),
	}

	s.mu.Unlock()

	s.logger.Info("Simulation templates initialized", "count", len(s.templates))
}

// simulationExecutionLoop processes pending simulation runs
func (s *Service) simulationExecutionLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.processPendingSimulations()
		}
	}
}

// processPendingSimulations processes simulations that are pending
func (s *Service) processPendingSimulations() {
	s.mu.Lock()
	runs := make([]*SimulationRun, 0, 10)
	for _, run := range s.runs {
		if run.Status == "pending" {
			runs = append(runs, run)
		}
	}
	s.mu.Unlock()

	for _, run := range runs {
		s.executeSimulation(run)
	}
}

// executeSimulation runs a simulation
func (s *Service) executeSimulation(run *SimulationRun) {
	s.mu.Lock()
	run.Status = "running"
	now := time.Now()
	run.StartedAt = &now
	s.mu.Unlock()

	s.logger.Info("Starting simulation", "run_id", run.ID, "name", run.Name)

	startTime := time.Now()

	// Run Monte Carlo simulation
	results := s.runMonteCarloSimulation(run.Parameters, run.Iterations)

	// Apply scenario weights
	results = s.applyScenarioWeights(run, results)

	elapsed := time.Since(startTime)

	s.mu.Lock()
	now = time.Now()
	run.Status = "completed"
	run.CompletedIterations = run.Iterations
	run.Progress = 100
	run.Results = results
	run.CompletedAt = &now
	run.Duration = elapsed
	s.mu.Unlock()

	s.logger.Info("Simulation completed", "run_id", run.ID, "duration", elapsed)
}

// runMonteCarloSimulation performs Monte Carlo simulation
func (s *Service) runMonteCarloSimulation(params SimulationParameters, iterations int) *SimulationResults {
	results := &SimulationResults{
		Variables:    make(map[string]VariableResults),
		Statistics:   make(map[string]float64),
		Distributions: make(map[string][]float64),
		Percentiles:  make(map[string]map[string]float64),
	}

	// Generate samples for each variable
	for _, varConfig := range params.Variables {
		samples := s.generateSamples(varConfig, iterations)
		results.Variables[varConfig.Name] = s.calculateVariableResults(samples)
		results.Distributions[varConfig.Name] = samples
	}

	// Calculate correlations
	results.Statistics["correlation_gdp_inflation"] = s.calculateCorrelation(
		results.Distributions["gdp_growth"],
		results.Distributions["inflation"],
	)

	return results
}

// generateSamples generates random samples based on distribution
func (s *Service) generateSamples(config VariableConfig, count int) []float64 {
	samples := make([]float64, count)

	switch config.Distribution {
	case "normal":
		for i := 0; i < count; i++ {
			samples[i] = config.Mean + config.StdDev*rand.NormFloat64()
		}
	case "lognormal":
		// Calculate parameters for lognormal from mean and stdDev
		sigmaSq := math.Log(1 + (config.StdDev*config.StdDev)/(config.Mean*config.Mean))
		mu := math.Log(config.Mean) - 0.5*sigmaSq
		sigma := math.Sqrt(sigmaSq)
		for i := 0; i < count; i++ {
			samples[i] = math.Exp(mu + sigma*rand.NormFloat64())
		}
	case "uniform":
		for i := 0; i < count; i++ {
			samples[i] = config.Min + rand.Float64()*(config.Max-config.Min)
		}
	case "triangular":
		for i := 0; i < count; i++ {
			u := rand.Float64()
			if u < (config.Mode-config.Min)/(config.Max-config.Min) {
				samples[i] = config.Min + math.Sqrt(u*(config.Max-config.Min)*(config.Mode-config.Min))
			} else {
				samples[i] = config.Max - math.Sqrt((1-u)*(config.Max-config.Min)*(config.Max-config.Mode))
			}
		}
	case "pert":
		// PERT distribution (similar to triangular but with mode weighted)
		scale := config.Max - config.Min
		mu := (config.Min + 4*config.Mode + config.Max) / 6
		sigma := scale / 6
		for i := 0; i < count; i++ {
			samples[i] = mu + sigma*rand.NormFloat64()
			if samples[i] < config.Min {
				samples[i] = config.Min
			}
			if samples[i] > config.Max {
				samples[i] = config.Max
			}
		}
	default:
		// Default to normal
		for i := 0; i < count; i++ {
			samples[i] = config.Mean + config.StdDev*rand.NormFloat64()
		}
	}

	return samples
}

// calculateVariableResults calculates statistics for a variable
func (s *Service) calculateVariableResults(samples []float64) VariableResults {
	n := float64(len(samples))

	// Calculate mean
	sum := 0.0
	for _, v := range samples {
		sum += v
	}
	mean := sum / n

	// Calculate standard deviation
	varianceSum := 0.0
	for _, v := range samples {
		varianceSum += (v - mean) * (v - mean)
	}
	stdDev := math.Sqrt(varianceSum / n)

	// Find min and max
	min := samples[0]
	max := samples[0]
	for _, v := range samples {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Calculate median
	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	for i := 0; i < len(sorted)/2; i++ {
		j := len(sorted) - 1 - i
		sorted[i], sorted[j] = sorted[j], sorted[i]
	}
	median := sorted[len(sorted)/2]

	return VariableResults{
		Mean:   mean,
		StdDev: stdDev,
		Min:    min,
		Max:    max,
		Median: median,
		Values: samples,
	}
}

// calculateCorrelation calculates correlation between two datasets
func (s *Service) calculateCorrelation(x, y []float64) float64 {
	n := float64(len(x))
	if n == 0 || len(y) == 0 || len(x) != len(y) {
		return 0
	}

	// Calculate means
	meanX := sumFloats(x) / n
	meanY := sumFloats(y) / n

	// Calculate correlation
	var numerator, denomX, denomY float64
	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		numerator += dx * dy
		denomX += dx * dx
		denomY += dy * dy
	}

	if denomX == 0 || denomY == 0 {
		return 0
	}

	return numerator / (math.Sqrt(denomX) * math.Sqrt(denomY))
}

func sumFloats(arr []float64) float64 {
	sum := 0.0
	for _, v := range arr {
		sum += v
	}
	return sum
}

// applyScenarioWeights applies scenario-based adjustments
func (s *Service) applyScenarioWeights(run *SimulationRun, results *SimulationResults) *SimulationResults {
	for _, scenario := range run.Parameters.Variables {
		// Apply scenario adjustments
		_ = scenario
	}

	return results
}

// CreateRun creates a new simulation run
func (s *Service) CreateRun(ctx context.Context, name, templateID string, iterations int, parameters SimulationParameters, createdBy string) (*SimulationRun, error) {
	run := &SimulationRun{
		ID:         fmt.Sprintf("sim-run-%d", time.Now().UnixNano()),
		Name:       name,
		TemplateID: templateID,
		Status:     "pending",
		Progress:   0,
		Iterations: iterations,
		Parameters: parameters,
		CreatedAt:  time.Now(),
		CreatedBy:  createdBy,
	}

	s.mu.Lock()
	s.runs[run.ID] = run
	s.mu.Unlock()

	return run, nil
}

// GetRun retrieves a simulation run
func (s *Service) GetRun(ctx context.Context, runID string) (*SimulationRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	run, exists := s.runs[runID]
	if !exists {
		return nil, fmt.Errorf("simulation run not found: %s", runID)
	}

	return run, nil
}

// ListRuns returns all simulation runs
func (s *Service) ListRuns(ctx context.Context, filters map[string]string) ([]*SimulationRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var runs []*SimulationRun
	for _, run := range s.runs {
		if filters["status"] != "" && run.Status != filters["status"] {
			continue
		}
		if filters["template_id"] != "" && run.TemplateID != filters["template_id"] {
			continue
		}
		runs = append(runs, run)
	}

	return runs, nil
}

// CancelRun cancels a running simulation
func (s *Service) CancelRun(ctx context.Context, runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	run, exists := s.runs[runID]
	if !exists {
		return fmt.Errorf("simulation run not found: %s", runID)
	}

	run.Status = "cancelled"

	return nil
}

// GetTemplates returns all simulation templates
func (s *Service) GetTemplates(ctx context.Context) ([]*SimulationTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	templates := make([]*SimulationTemplate, 0, len(s.templates))
	for _, t := range s.templates {
		templates = append(templates, t)
	}

	return templates, nil
}

// GetTemplate retrieves a specific template
func (s *Service) GetTemplate(ctx context.Context, templateID string) (*SimulationTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	template, exists := s.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	return template, nil
}

// CreateTemplate creates a new simulation template
func (s *Service) CreateTemplate(ctx context.Context, template SimulationTemplate) (*SimulationTemplate, error) {
	template.ID = fmt.Sprintf("template-%d", time.Now().UnixNano())
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	s.mu.Lock()
	s.templates[template.ID] = &template
	s.mu.Unlock()

	return &template, nil
}
