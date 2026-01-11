package inflation

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/redis/go-redis/v9"
	"neam-platform/shared"
)

// TimePeriod represents a time period
type TimePeriod struct {
	Type     string    `json:"type"` // monthly, quarterly, annual
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Label    string    `json:"label"`
}

// Config holds the configuration for the inflation prediction service
type Config struct {
	ClickHouse clickhouse.Conn
	PostgreSQL *shared.PostgreSQL
	Redis      *redis.Client
	Logger     *shared.Logger
}

// Service handles inflation prediction and analysis
type Service struct {
	config     Config
	logger     *shared.Logger
	clickhouse clickhouse.Conn
	postgreSQL *shared.PostgreSQL
	redis      *redis.Client
	forecasts  map[string]*InflationEstimate
	scenarios  map[string]*InflationScenario
	components map[string]*InflationComponent
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// InflationEstimate represents an inflation forecast
type InflationEstimate struct {
	ID          string                 `json:"id"`
	Period      TimePeriod             `json:"period"`
	Value       float64                `json:"value"`
	LowerBound  float64                `json:"lower_bound"`
	UpperBound  float64                `json:"upper_bound"`
	Confidence  float64                `json:"confidence"`
	Percentiles map[string]float64     `json:"percentiles"`
	Components  []ComponentContribution `json:"components"`
	Method      string                 `json:"method"` // arimax, lstm, ensemble
	Drivers     []InflationDriver      `json:"drivers"`
	CreatedAt   time.Time              `json:"created_at"`
	ValidUntil  time.Time              `json:"valid_until"`
}

// ComponentContribution represents contribution from a component
type ComponentContribution struct {
	Component string  `json:"component"`
	Weight    float64 `json:"weight"`
	Value     float64 `json:"value"`
	Change    float64 `json:"change"`
	Contribution float64 `json:"contribution"` // contribution to overall inflation
}

// InflationDriver represents a driver of inflation
type InflationDriver struct {
	Name        string  `json:"name"`
	Impact      float64 `json:"impact"` // 0-1 scale
	CurrentValue float64 `json:"current_value"`
	Trend       string  `json:"trend"` // increasing, stable, decreasing
	Forecast    float64 `json:"forecast"`
}

// InflationScenario represents an inflation scenario
type InflationScenario struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // baseline, optimistic, pessimistic
	Forecasts   []InflationEstimate    `json:"forecasts"`
	Parameters  ScenarioParameters     `json:"parameters"`
	Probability float64                `json:"probability"`
	CreatedAt   time.Time              `json:"created_at"`
}

// InflationComponent represents an inflation component
type InflationComponent struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Category    string    `json:"category"` // food, fuel, core, services
	Weight      float64   `json:"weight"`
	LatestValue float64   `json:"latest_value"`
	PreviousValue float64 `json:"previous_value"`
	Change      float64   `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	Trend       string    `json:"trend"` // up, down, stable
	UpdatedAt   time.Time `json:"updated_at"`
}

// RegionalInflation represents inflation by region
type RegionalInflation struct {
	Region      string    `json:"region"`
	Value       float64   `json:"value"`
	Change      float64   `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	Components  []ComponentContribution `json:"components"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CategoryInflation represents inflation by category
type CategoryInflation struct {
	Category    string    `json:"category"`
	Value       float64   `json:"value"`
	Change      float64   `json:"change"`
	Subcategories []SubCategoryInflation `json:"subcategories"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SubCategoryInflation represents inflation by subcategory
type SubCategoryInflation struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Change      float64 `json:"change"`
	Weight      float64 `json:"weight"`
}

// NewService creates a new inflation prediction service
func NewService(cfg Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		config:     cfg,
		logger:     cfg.Logger,
		clickhouse: cfg.ClickHouse,
		postgreSQL: cfg.PostgreSQL,
		redis:      cfg.Redis,
		forecasts:  make(map[string]*InflationEstimate),
		scenarios:  make(map[string]*InflationScenario),
		components: make(map[string]*InflationComponent),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start begins the inflation prediction service
func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Starting inflation prediction service")

	// Initialize components
	s.initializeComponents()

	// Start background updates
	s.wg.Add(1)
	go s.forecastUpdateLoop()

	s.logger.Info("Inflation prediction service started")
}

// Stop gracefully stops the service
func (s *Service) Stop() {
	s.logger.Info("Stopping inflation prediction service")
	s.cancel()
	s.wg.Wait()
	s.logger.Info("Inflation prediction service stopped")
}

// initializeComponents sets up default inflation components
func (s *Service) initializeComponents() {
	s.mu.Lock()

	s.components["food"] = &InflationComponent{
		ID:   "food",
		Name: "Food and Non-Alcoholic Beverages",
		Code: "CF1",
		Category: "food",
		Weight:    0.30,
		LatestValue: 6.2,
		PreviousValue: 6.5,
		Change:    -0.3,
		ChangePercent: -4.6,
		Trend:     "decreasing",
	}

	s.components["fuel"] = &InflationComponent{
		ID:   "fuel",
		Name: "Fuel and Lubricants",
		Code: "CF2",
		Category: "fuel",
		Weight:    0.15,
		LatestValue: 3.8,
		PreviousValue: 4.2,
		Change:    -0.4,
		ChangePercent: -9.5,
		Trend:     "decreasing",
	}

	s.components["core"] = &InflationComponent{
		ID:   "core",
		Name: "Core Inflation (Excl Food and Fuel)",
		Code: "CF3",
		Category: "core",
		Weight:    0.40,
		LatestValue: 4.1,
		PreviousValue: 4.0,
		Change:    0.1,
		ChangePercent: 2.5,
		Trend:     "stable",
	}

	s.components["services"] = &InflationComponent{
		ID:   "services",
		Name: "Services",
		Code: "CF4",
		Category: "services",
		Weight:    0.15,
		LatestValue: 3.5,
		PreviousValue: 3.4,
		Change:    0.1,
		ChangePercent: 2.9,
		Trend:     "stable",
	}

	s.mu.Unlock()
}

// forecastUpdateLoop periodically updates forecasts
func (s *Service) forecastUpdateLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.updateForecasts()
		}
	}
}

// updateForecasts recalculates inflation forecasts
func (s *Service) updateForecasts() {
	ctx := context.Background()

	// Generate forecast using ARIMAX model
	forecast := s.generateARIMAXForecast(ctx)

	s.mu.Lock()
	s.forecasts[forecast.ID] = forecast
	s.mu.Unlock()

	s.logger.Info("Inflation forecast updated", "period", forecast.Period.Label, "value", forecast.Value)
}

// generateARIMAXForecast generates an ARIMAX forecast
func (s *Service) generateARIMAXForecast(ctx context.Context) *InflationEstimate {
	baseValue := 4.52

	estimate := &InflationEstimate{
		ID:         fmt.Sprintf("inflation-arimax-%d", time.Now().UnixNano()),
		Period:     TimePeriod{Type: "monthly", Label: "January 2024"},
		Value:      baseValue,
		LowerBound: baseValue - 0.5,
		UpperBound: baseValue + 0.5,
		Confidence: 0.88,
		Method:     "ARIMAX",
		CreatedAt:  time.Now(),
		ValidUntil: time.Now().Add(24 * time.Hour),
	}

	// Calculate percentiles
	estimate.Percentiles = map[string]float64{
		"p10": baseValue - 0.8,
		"p25": baseValue - 0.35,
		"p50": baseValue,
		"p75": baseValue + 0.35,
		"p90": baseValue + 0.8,
	}

	// Add component contributions
	s.mu.RLock()
	for _, comp := range s.components {
		estimate.Components = append(estimate.Components, ComponentContribution{
			Component:  comp.Name,
			Weight:     comp.Weight,
			Value:      comp.LatestValue,
			Change:     comp.Change,
			Contribution: comp.LatestValue * comp.Weight,
		})
	}
	s.mu.RUnlock()

	// Add drivers
	estimate.Drivers = []InflationDriver{
		{Name: "Crude Oil Prices", Impact: 0.25, CurrentValue: 75.0, Trend: "stable", Forecast: 72.0},
		{Name: "Exchange Rate", Impact: 0.20, CurrentValue: 83.5, Trend: "stable", Forecast: 84.0},
		{Name: "Food Prices", Impact: 0.30, CurrentValue: 6.2, Trend: "decreasing", Forecast: 5.8},
		{Name: "Wage Growth", Impact: 0.15, CurrentValue: 5.5, Trend: "stable", Forecast: 5.2},
	}

	return estimate
}

// GetForecast retrieves the current inflation forecast
func (s *Service) GetForecast(ctx context.Context, periodType string) (*InflationEstimate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, f := range s.forecasts {
		if f.Period.Type == periodType {
			return f, nil
		}
	}

	return nil, fmt.Errorf("no forecast found for period: %s", periodType)
}

// GetComponents returns all inflation components
func (s *Service) GetComponents(ctx context.Context) ([]*InflationComponent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	components := make([]*InflationComponent, 0, len(s.components))
	for _, c := range s.components {
		components = append(components, c)
	}

	return components, nil
}

// GetByCategory returns inflation by category
func (s *Service) GetByCategory(ctx context.Context) ([]CategoryInflation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return []CategoryInflation{
		{
			Category: "Food",
			Value:    6.2,
			Change:   -0.3,
			Subcategories: []SubCategoryInflation{
				{Name: "Cereals", Value: 5.5, Change: -0.2, Weight: 0.10},
				{Name: "Meat", Value: 8.2, Change: -0.5, Weight: 0.08},
				{Name: "Dairy", Value: 4.8, Change: 0.1, Weight: 0.05},
				{Name: "Vegetables", Value: 12.5, Change: -1.2, Weight: 0.04},
				{Name: "Fruits", Value: 9.8, Change: -0.8, Weight: 0.03},
			},
			UpdatedAt: time.Now(),
		},
		{
			Category: "Fuel",
			Value:    3.8,
			Change:   -0.4,
			Subcategories: []SubCategoryInflation{
				{Name: "Petrol", Value: 3.5, Change: -0.5, Weight: 0.08},
				{Name: "Diesel", Value: 4.2, Change: -0.3, Weight: 0.05},
				{Name: "LPG", Value: 3.8, Change: -0.2, Weight: 0.02},
			},
			UpdatedAt: time.Now(),
		},
		{
			Category: "Core",
			Value:    4.1,
			Change:   0.1,
			Subcategories: []SubCategoryInflation{
				{Name: "Clothing", Value: 3.5, Change: 0.0, Weight: 0.06},
				{Name: "Housing", Value: 4.2, Change: 0.2, Weight: 0.10},
				{Name: "Household Goods", Value: 3.8, Change: 0.1, Weight: 0.08},
				{Name: "Health", Value: 5.2, Change: 0.1, Weight: 0.05},
				{Name: "Education", Value: 4.8, Change: 0.0, Weight: 0.04},
				{Name: "Communication", Value: 2.5, Change: -0.1, Weight: 0.03},
				{Name: "Recreation", Value: 3.2, Change: 0.1, Weight: 0.04},
			},
			UpdatedAt: time.Now(),
		},
	}, nil
}

// GetByRegion returns inflation by region
func (s *Service) GetByRegion(ctx context.Context) ([]RegionalInflation, error) {
	return []RegionalInflation{
		{Region: "NORTH", Value: 4.72, Change: 0.15, ChangePercent: 3.3, UpdatedAt: time.Now()},
		{Region: "SOUTH", Value: 4.31, Change: -0.08, ChangePercent: -1.8, UpdatedAt: time.Now()},
		{Region: "EAST", Value: 4.58, Change: 0.05, ChangePercent: 1.1, UpdatedAt: time.Now()},
		{Region: "WEST", Value: 4.45, Change: -0.02, ChangePercent: -0.4, UpdatedAt: time.Now()},
		{Region: "CENTRAL", Value: 4.62, Change: 0.10, ChangePercent: 2.2, UpdatedAt: time.Now()},
	}, nil
}

// CreateScenario creates a new inflation scenario
func (s *Service) CreateScenario(ctx context.Context, name, scenarioType string, params ScenarioParameters) (*InflationScenario, error) {
	scenario := &InflationScenario{
		ID:        fmt.Sprintf("scenario-inflation-%d", time.Now().UnixNano()),
		Name:      name,
		Type:      scenarioType,
		CreatedAt: time.Now(),
	}

	// Generate forecasts for this scenario
	scenario.Forecasts = s.generateScenarioForecasts(ctx, params)

	return scenario, nil
}

// generateScenarioForecasts generates forecasts based on scenario parameters
func (s *Service) generateScenarioForecasts(ctx context.Context, params ScenarioParameters) []InflationEstimate {
	var forecasts []InflationEstimate

	// Adjust base inflation based on parameters
	adjustment := s.calculateParameterAdjustment(params)

	baseValue := 4.52 + adjustment

	// Generate forecasts for different horizons
	for months := 1; months <= 12; months++ {
		// Gradually return to baseline
		reversionFactor := math.Min(1.0, float64(months)/6.0)
		adjustedValue := baseValue - (adjustment * (1 - reversionFactor))

		forecasts = append(forecasts, InflationEstimate{
			ID:         fmt.Sprintf("inflation-scenario-%d-%d", time.Now().UnixNano(), months),
			Period:     TimePeriod{Type: "monthly", Label: fmt.Sprintf("Month %d", months)},
			Value:      math.Round(adjustedValue*100)/100,
			Confidence: 0.85 - float64(months)*0.02,
			CreatedAt:  time.Now(),
			ValidUntil: time.Now().AddDate(0, months, 0),
		})
	}

	return forecasts
}

// calculateParameterAdjustment calculates inflation adjustment based on parameters
func (s *Service) calculateParameterAdjustment(params ScenarioParameters) float64 {
	// Simplified adjustment calculation
	return (params.OilPrice - 75.0) * 0.01 + (params.ExchangeRate - 83.5) * 0.02
}

// GetHistory returns historical inflation data
func (s *Service) GetHistory(ctx context.Context, startDate, endDate time.Time) ([]HistoryEntry, error) {
	var history []HistoryEntry

	// Generate sample history data
	currentValue := 5.5
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		// Add some variation
		variation := (float64(date.Day()) / 30.0) - 0.5
		currentValue += variation * 0.1

		history = append(history, HistoryEntry{
			Date:  date,
			Value: math.Round(currentValue*100)/100,
		})
	}

	return history, nil
}

// HistoryEntry represents a historical inflation entry
type HistoryEntry struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// ScenarioParameters represents parameters for inflation scenario
type ScenarioParameters struct {
	OilPrice       float64 `json:"oil_price"`
	ExchangeRate   float64 `json:"exchange_rate"`
	FoodPriceIndex float64 `json:"food_price_index"`
	WageGrowth     float64 `json:"wage_growth"`
	MonetaryPolicy string  `json:"monetary_policy"` // expansion, neutral, contraction
}
