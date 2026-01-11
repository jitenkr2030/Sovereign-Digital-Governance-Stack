package inflation

import (
	"math"
	"sync"
	"time"

	"neam-platform/intelligence/models"
)

// Engine processes inflation data and generates predictions
type Engine struct {
	store              InflationStore
	config             *Config
	mu                 sync.RWMutex
	running            bool
	stopChan           chan struct{}
}

// InflationStore interface for data persistence
type InflationStore interface {
	GetCPIHistory(regionID string, lookbackDays int) ([]models.InflationData, error)
	GetCommodityPrices(lookbackDays int) ([]models.TimeSeriesPoint, error)
	GetProducerPrices(regionID string, lookbackDays int) ([]models.TimeSeriesPoint, error)
	SavePrediction(prediction *models.InflationPrediction) error
}

// Config holds inflation engine configuration
type Config struct {
	RegionID              string
	UpdateInterval        time.Duration
	HistoryLookbackDays   int
	PredictionHorizons    []int // Days: [30, 90, 180]
	ConfidenceThreshold   float64
	LeadingIndicators     []models.MetricType
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		RegionID:            "NATIONAL",
		UpdateInterval:      time.Hour,
		HistoryLookbackDays: 365,
		PredictionHorizons:  []int{30, 90, 180},
		ConfidenceThreshold: 0.7,
		LeadingIndicators: []models.MetricType{
			models.MetricProducerPrice,
			models.MetricCommodityPrice,
		},
	}
}

// NewEngine creates a new inflation engine
func NewEngine(store InflationStore, config *Config) *Engine {
	if config == nil {
		config = DefaultConfig()
	}
	return &Engine{
		store:    store,
		config:   config,
		stopChan: make(chan struct{}),
	}
}

// Start begins the inflation prediction loop
func (e *Engine) Start() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.mu.Unlock()

	ticker := time.NewTicker(e.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.RunAnalysis()
		case <-e.stopChan:
			return
		}
	}
}

// Stop halts the inflation engine
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.running {
		close(e.stopChan)
		e.running = false
	}
}

// RunAnalysis performs a single inflation analysis cycle
func (e *Engine) RunAnalysis() (*models.InflationPrediction, error) {
	// Fetch historical data
	data, err := e.store.GetCPIHistory(e.config.RegionID, e.config.HistoryLookbackDays)
	if err != nil {
		return nil, err
	}
	
	if len(data) < 30 {
		// Insufficient data for prediction
		return nil, nil
	}
	
	// Get leading indicators
	leadingData, _ := e.store.GetCommodityPrices(30)
	producerData, _ := e.store.GetProducerPrices(e.config.RegionID, 30)
	
	// Calculate current inflation metrics
	currentData := data[len(data)-1]
	currentInflation := e.calculateCurrentInflation(data)
	
	// Generate predictions
	predictions := e.generatePredictions(data, leadingData, producerData)
	
	// Determine risk level
	riskLevel := e.assessRiskLevel(predictions, currentInflation)
	
	// Identify contributing factors
	contributingFactors := e.identifyContributingFactors(data, leadingData, producerData)
	
	// Build prediction object
	prediction := &models.InflationPrediction{
		Timestamp:           time.Now(),
		RegionID:            e.config.RegionID,
		CurrentInflation:    currentInflation,
		Predicted30Day:      predictions[30],
		Predicted90Day:      predictions[90],
		Predicted180Day:     predictions[180],
		Confidence:          e.calculatePredictionConfidence(data),
		RiskLevel:           riskLevel,
		LeadingIndicators:   e.formatLeadingIndicators(leadingData, producerData),
		ContributingFactors: contributingFactors,
	}
	
	// Save prediction
	if err := e.store.SavePrediction(prediction); err != nil {
		// Log but don't fail
	}
	
	return prediction, nil
}

// calculateCurrentInflation computes the current inflation rate
func (e *Engine) calculateCurrentInflation(data []models.InflationData) float64 {
	if len(data) < 2 {
		return 0
	}
	
	// Year-over-year CPI change
	recentCPI := data[len(data)-1].CPI
	yearAgoIdx := len(data) - 365
	if yearAgoIdx < 0 {
		yearAgoIdx = 0
	}
	
	yearAgoCPI := data[yearAgoIdx].CPI
	if yearAgoCPI == 0 {
		return 0
	}
	
	return (recentCPI - yearAgoCPI) / yearAgoCPI
}

// generatePredictions generates inflation forecasts for different horizons
func (e *Engine) generatePredictions(
	historicalData []models.InflationData,
	commodityData []models.TimeSeriesPoint,
	producerData []models.TimeSeriesPoint,
) map[int]float64 {
	
	predictions := make(map[int]float64)
	
	// Extract CPI series
	cpiSeries := make([]float64, len(historicalData))
	for i, d := range historicalData {
		cpiSeries[i] = d.CPI
	}
	
	// Calculate current inflation trend
	trend := e.calculateTrend(cpiSeries)
	
	// Get leading indicator influence
	commodityInfluence := e.calculateLeadingIndicatorInfluence(commodityData)
	producerInfluence := e.calculateLeadingIndicatorInfluence(producerData)
	
	// Apply Holt-Winters style exponential smoothing for predictions
	for _, horizon := range e.config.PredictionHorizons {
		predictions[horizon] = e.predictWithInfluences(
			cpiSeries, trend, commodityInfluence, producerInfluence, horizon,
		)
	}
	
	return predictions
}

// calculateTrend calculates the overall trend in the data
func (e *Engine) calculateTrend(series []float64) float64 {
	if len(series) < 2 {
		return 0
	}
	
	// Use linear regression for trend
	n := float64(len(series))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0
	
	for i, y := range series {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	// Calculate slope (trend)
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	
	// Normalize to percentage trend
	avgY := sumY / n
	if avgY != 0 {
		return slope / avgY
	}
	
	return 0
}

// calculateLeadingIndicatorInfluence calculates the influence of leading indicators
func (e *Engine) calculateLeadingIndicatorInfluence(data []models.TimeSeriesPoint) float64 {
	if len(data) < 2 {
		return 0
	}
	
	// Calculate recent change in leading indicators
	recent := data[len(data)-1].Value
	prev := data[len(data)-2].Value
	
	if prev == 0 {
		return 0
	}
	
	change := (recent - prev) / prev
	return change
}

// predictWithInfluences combines trend and leading indicators for prediction
func (e *Engine) predictWithInfluences(
	series []float64,
	trend float64,
	commodityInfluence float64,
	producerInfluence float64,
	horizon int,
) float64 {
	
	currentValue := series[len(series)-1]
	
	// Base prediction from trend
	basePrediction := currentValue * (1 + trend*float64(horizon)/365)
	
	// Adjust for leading indicators (with horizon-dependent weighting)
	leadAdj := 0.0
	if len(series) >= 30 {
		recentAvg := average(series[len(series)-30:])
		leadAdj = (commodityInfluence + producerInfluence) * 0.5 * (currentValue / recentAvg)
	}
	
	// Combine predictions
	predictedValue := basePrediction * (1 + leadAdj)
	
	// Convert to inflation rate
	if currentValue != 0 {
		return (predictedValue - currentValue) / currentValue
	}
	
	return 0
}

// calculatePredictionConfidence determines confidence level for predictions
func (e *Engine) calculatePredictionConfidence(data []models.InflationData) float64 {
	// Base confidence on data quality and stability
	if len(data) < 90 {
		return 0.5
	}
	
	// Calculate volatility
	changes := make([]float64, 0, len(data)-1)
	for i := 1; i < len(data); i++ {
		if data[i-1].CPI != 0 {
			change := (data[i].CPI - data[i-1].CPI) / data[i-1].CPI
			changes = append(changes, change)
		}
	}
	
	if len(changes) == 0 {
		return 0.5
	}
	
	// Lower volatility = higher confidence
	volatility := standardDeviation(changes)
	
	// Confidence inversely proportional to volatility
	confidence := 1.0 - min(1.0, volatility*10)
	
	return confidence
}

// assessRiskLevel determines the inflation risk category
func (e *Engine) assessRiskLevel(predictions map[int]float64, currentInflation float64) string {
	avgPrediction := (predictions[30] + predictions[90] + predictions[180]) / 3
	
	switch {
	case avgPrediction > 0.10 || currentInflation > 0.10:
		return "CRITICAL"
	case avgPrediction > 0.07 || currentInflation > 0.07:
		return "HIGH"
	case avgPrediction > 0.04 || currentInflation > 0.04:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// identifyContributingFactors identifies factors contributing to inflation
func (e *Engine) identifyContributingFactors(
	historicalData []models.InflationData,
	commodityData []models.TimeSeriesPoint,
	producerData []models.TimeSeriesPoint,
) []string {
	
	factors := make([]string, 0)
	
	if len(historicalData) < 2 {
		return factors
	}
	
	// Check commodity prices
	if len(commodityData) >= 2 {
		commodityChange := (commodityData[len(commodityData)-1].Value - commodityData[0].Value) / commodityData[0].Value
		if commodityChange > 0.1 {
			factors = append(factors, "Rising commodity prices")
		} else if commodityChange < -0.1 {
			factors = append(factors, "Falling commodity prices")
		}
	}
	
	// Check producer prices
	if len(producerData) >= 2 {
		producerChange := (producerData[len(producerData)-1].Value - producerData[0].Value) / producerData[0].Value
		if producerChange > 0.08 {
			factors = append(factors, "Rising producer costs")
		}
	}
	
	// Check core CPI vs overall CPI
	recent := historicalData[len(historicalData)-1]
	if recent.CoreCPI > recent.CPI {
		factors = append(factors, "Core inflation pressures")
	}
	
	// Check PPI influence
	if recent.PPI > recent.CPI*1.5 {
		factors = append(factors, "Producer price passthrough")
	}
	
	return factors
}

// formatLeadingIndicators formats leading indicator names
func (e *Engine) formatLeadingIndicators(commodityData, producerData []models.TimeSeriesPoint) []string {
	indicators := make([]string, 0)
	
	if len(commodityData) > 0 {
		indicators = append(indicators, "Commodity Price Index")
	}
	if len(producerData) > 0 {
		indicators = append(indicators, "Producer Price Index")
	}
	
	return indicators
}

// SimplePredictionModel provides basic inflation prediction without external dependencies
type SimplePredictionModel struct {
	Alpha float64 // Smoothing factor
	Beta  float64 // Trend factor
}

// NewSimplePredictionModel creates a simple prediction model
func NewSimplePredictionModel() *SimplePredictionModel {
	return &SimplePredictionModel{
		Alpha: 0.3,
		Beta:  0.1,
	}
}

// Predict generates a simple inflation forecast
func (m *SimplePredictionModel) Predict(series []float64, horizon int) (float64, float64) {
	if len(series) < 2 {
		return 0, 0
	}
	
	// Double exponential smoothing
	level := series[0]
	trend := series[1] - series[0]
	
	for _, value := range series[1:] {
		newLevel := m.Alpha*value + (1-m.Alpha)*(level+trend)
		newTrend := m.Beta*(newLevel-level) + (1-m.Beta)*trend
		level = newLevel
		trend = newTrend
	}
	
	// Forecast
	forecast := level + float64(horizon)*trend
	
	// Calculate confidence based on data stability
	confidence := m.calculateConfidence(series)
	
	// Return inflation rate
	if level != 0 {
		return (forecast - level) / level, confidence
	}
	
	return 0, confidence
}

// calculateConfidence determines model confidence
func (m *SimplePredictionModel) calculateConfidence(series []float64) float64 {
	if len(series) < 10 {
		return 0.5
	}
	
	// Calculate coefficient of variation
	avg := average(series)
	if avg == 0 {
		return 0.5
	}
	
	cv := standardDeviation(series) / avg
	confidence := 1.0 - min(1.0, cv)
	
	return confidence
}

// Helper functions

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func standardDeviation(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	avg := average(values)
	variance := 0.0
	
	for _, v := range values {
		diff := v - avg
		variance += diff * diff
	}
	
	return math.Sqrt(variance / float64(len(values)))
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
