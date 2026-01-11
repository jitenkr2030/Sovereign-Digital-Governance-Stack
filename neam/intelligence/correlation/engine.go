package correlation

import (
	"math"
	"sync"
	"time"

	"neam-platform/intelligence/models"
)

// Engine performs cross-sector correlation analysis
type Engine struct {
	store    CorrelationStore
	config   *Config
	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}
}

// CorrelationStore interface for data persistence
type CorrelationStore interface {
	GetTimeSeriesData(metricType models.MetricType, regionID string, lookbackDays int) ([]models.TimeSeriesPoint, error)
	SaveCorrelationData(data *models.CorrelationData) error
	GetRegionalStressIndex(regionID string) (*models.RegionalStressIndex, error)
	SaveRegionalStressIndex(index *models.RegionalStressIndex) error
}

// Config holds correlation engine configuration
type Config struct {
	RegionID              string
	UpdateInterval        time.Duration
	LookbackDays          int
	CorrelationThreshold  float64 // Strong correlation threshold
	DecouplingThreshold   float64 // Decoupling detection threshold
	AlertCooldown         time.Duration
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		RegionID:             "NATIONAL",
		UpdateInterval:       time.Hour,
		LookbackDays:         7,
		CorrelationThreshold: 0.8,
		DecouplingThreshold:  0.3,
		AlertCooldown:        time.Hour,
	}
}

// NewEngine creates a new correlation engine
func NewEngine(store CorrelationStore, config *Config) *Engine {
	if config == nil {
		config = DefaultConfig()
	}
	return &Engine{
		store:    store,
		config:   config,
		stopChan: make(chan struct{}),
	}
}

// Start begins the correlation analysis loop
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

// Stop halts the correlation engine
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.running {
		close(e.stopChan)
		e.running = false
	}
}

// RunAnalysis performs a single correlation analysis cycle
func (e *Engine) RunAnalysis() ([]models.CorrelationData, error) {
	correlations := make([]models.CorrelationData, 0)
	
	// Define metric pairs to correlate
	pairs := []struct {
		A models.MetricType
		B models.MetricType
	}{
		{models.MetricEnergyConsumption, models.MetricPaymentVolume},
		{models.MetricEnergyConsumption, models.MetricPaymentValue},
		{models.MetricRetailSales, models.MetricEmployment},
		{models.MetricManufacturingOutput, models.MetricEnergyConsumption},
		{models.MetricInventoryLevel, models.MetricCommodityPrice},
		{models.MetricCashTransaction, models.MetricDigitalTransaction},
	}
	
	for _, pair := range pairs {
		corr := e.analyzePair(pair.A, pair.B)
		if corr != nil {
			correlations = append(correlations, *corr)
			
			// Save correlation data
			_ = e.store.SaveCorrelationData(corr)
			
			// Check for decoupling alerts
			if corr.Deviation < -e.config.DecouplingThreshold {
				// Trigger decoupling alert
				_ = e.handleDecoupling(corr)
			}
		}
	}
	
	// Update regional stress indices
	_ = e.updateRegionalStress()
	
	return correlations, nil
}

// analyzePair calculates correlation between two metrics
func (e *Engine) analyzePair(metricA, metricB models.MetricType) *models.CorrelationData {
	// Fetch time series data
	seriesA, err := e.store.GetTimeSeriesData(metricA, e.config.RegionID, e.config.LookbackDays)
	if err != nil || len(seriesA) == 0 {
		return nil
	}
	
	seriesB, err := e.store.GetTimeSeriesData(metricB, e.config.RegionID, e.config.LookbackDays)
	if err != nil || len(seriesB) == 0 {
		return nil
	}
	
	// Align series by timestamp
	alignedA, alignedB := alignSeries(seriesA, seriesB)
	
	if len(alignedA) < 3 {
		return nil
	}
	
	// Calculate Pearson correlation
	pearsonCorr := calculatePearsonCorrelation(alignedA, alignedB)
	spearmanCorr := calculateSpearmanCorrelation(alignedA, alignedB)
	
	// Calculate deviation from historical average (mock historical for now)
	historicalAvg := e.getHistoricalCorrelation(metricA, metricB)
	deviation := pearsonCorr - historicalAvg
	
	return &models.CorrelationData{
		Timestamp:      time.Now(),
		MetricA:        metricA,
		MetricB:        metricB,
		RegionID:       e.config.RegionID,
		PearsonCoeff:   pearsonCorr,
		SpearmanCoeff:  spearmanCorr,
		WindowDays:     e.config.LookbackDays,
		HistoricalAvg:  historicalAvg,
		Deviation:      deviation,
	}
}

// getHistoricalCorrelation retrieves historical average correlation
func (e *Engine) getHistoricalCorrelation(metricA, metricB models.MetricType) float64 {
	// In production, this would query historical correlation data
	// For now, return reasonable defaults based on known relationships
	
	// Energy and payments typically have strong positive correlation
	if metricA == models.MetricEnergyConsumption && metricB == models.MetricPaymentVolume {
		return 0.75
	}
	
	// Cash and digital transactions typically have moderate negative correlation
	if metricA == models.MetricCashTransaction && metricB == models.MetricDigitalTransaction {
		return -0.3
	}
	
	// Retail and employment typically have positive correlation
	if metricA == models.MetricRetailSales && metricB == models.MetricEmployment {
		return 0.65
	}
	
	// Default moderate positive correlation
	return 0.5
}

// handleDecoupling processes detected decoupling events
func (e *Engine) handleDecoupling(corr *models.CorrelationData) error {
	// This would trigger alerts and store decoupling events
	return nil
}

// updateRegionalStress calculates and updates regional stress indices
func (e *Engine) updateRegionalStress() error {
	// In production, this would aggregate data from multiple regions
	return nil
}

// calculatePearsonCorrelation computes Pearson correlation coefficient
func calculatePearsonCorrelation(seriesA, seriesB []float64) float64 {
	n := float64(len(seriesA))
	
	// Calculate means
	meanA := mean(seriesA)
	meanB := mean(seriesB)
	
	// Calculate covariance and variances
	covariance := 0.0
	varA := 0.0
	varB := 0.0
	
	for i := range seriesA {
		diffA := seriesA[i] - meanA
		diffB := seriesB[i] - meanB
		
		covariance += diffA * diffB
		varA += diffA * diffA
		varB += diffB * diffB
	}
	
	// Calculate correlation
	if varA == 0 || varB == 0 {
		return 0
	}
	
	return covariance / math.Sqrt(varA*varB)
}

// calculateSpearmanCorrelation computes Spearman rank correlation
func calculateSpearmanCorrelation(seriesA, seriesB []float64) float64 {
	// Convert to ranks
	ranksA := toRanks(seriesA)
	ranksB := toRanks(seriesB)
	
	// Calculate Pearson on ranks
	return calculatePearsonCorrelation(ranksA, ranksB)
}

// toRanks converts values to their ranks
func toRanks(values []float64) []float64 {
	n := len(values)
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	
	// Sort by value
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if values[i] > values[j] {
				values[i], values[j] = values[j], values[i]
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}
	
	// Assign ranks (average for ties)
	ranks := make([]float64, n)
	visited := make(map[float64][]int)
	
	for i, v := range values {
		visited[v] = append(visited[v], i)
	}
	
	for _, indices := range visited {
		rank := float64(0)
		for _, idx := range indices {
			rank += float64(idx + 1)
		}
		rank /= float64(len(indices))
		for _, idx := range indices {
			ranks[idx] = rank
		}
	}
	
	return ranks
}

// alignSeries aligns two time series by timestamp
func alignSeries(seriesA, seriesB []models.TimeSeriesPoint) ([]float64, []float64) {
	// Create maps by timestamp
	mapA := make(map[time.Time]float64)
	for _, p := range seriesA {
		mapA[p.Timestamp] = p.Value
	}
	
	mapB := make(map[time.Time]float64)
	for _, p := range seriesB {
		mapB[p.Timestamp] = p.Value
	}
	
	// Find common timestamps
	resultA := make([]float64, 0)
	resultB := make([]float64, 0)
	
	for t, vA := range mapA {
		if vB, ok := mapB[t]; ok {
			resultA = append(resultA, vA)
			resultB = append(resultB, vB)
		}
	}
	
	return resultA, resultB
}

// mean calculates the arithmetic mean
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// HeatMapper generates regional heat maps based on economic stress
type HeatMapper struct {
	store CorrelationStore
	mu    sync.RWMutex
}

// NewHeatMapper creates a new heat mapper
func NewHeatMapper(store CorrelationStore) *HeatMapper {
	return &HeatMapper{
		store: store,
	}
}

// GenerateHeatMap creates a regional stress heat map
func (h *HeatMapper) GenerateHeatMap(regions []string) ([]models.RegionalStressIndex, error) {
	indices := make([]models.RegionalStressIndex, 0, len(regions))
	
	for _, region := range regions {
		index, err := h.calculateRegionalStress(region)
		if err != nil {
			continue
		}
		indices = append(indices, *index)
	}
	
	return indices, nil
}

// calculateRegionalStress computes the stress index for a region
func (h *HeatMapper) calculateRegionalStress(regionID string) (*models.RegionalStressIndex, error) {
	// Aggregate anomaly signals for the region
	// In production, this would query the signal store
	
	stress := &models.RegionalStressIndex{
		Timestamp:    time.Now(),
		RegionID:     regionID,
		OverallScore: 50.0, // Default middle score
		AnomalyCount: 0,
		AnomalyScore: 0.0,
		InflationScore: 50.0,
		CorrelationBreaks: 0,
	}
	
	// Fetch existing data
	existing, err := h.store.GetRegionalStressIndex(regionID)
	if err == nil && existing != nil {
		stress = existing
	}
	
	// Update score based on current signals
	h.updateStressFromSignals(stress)
	
	// Save updated index
	_ = h.store.SaveRegionalStressIndex(stress)
	
	return stress, nil
}

// updateStressFromSignals updates stress index based on recent signals
func (h *HeatMapper) updateStressFromSignals(stress *models.RegionalStressIndex) {
	// Normalize components to 0-100 scale
	anomalyComponent := stress.AnomalyScore * 100
	inflationComponent := stress.InflationScore
	correlationComponent := float64(stress.CorrelationBreaks) * 20
	
	// Weighted average
	stress.OverallScore = (anomalyComponent*0.4 + inflationComponent*0.4 + correlationComponent*0.2)
	
	// Clamp to 0-100
	stress.OverallScore = math.Max(0, math.Min(100, stress.OverallScore))
}

// RegionalAnalyzer provides deep analysis of regional economic data
type RegionalAnalyzer struct {
	store CorrelationStore
	mu    sync.RWMutex
}

// NewRegionalAnalyzer creates a new regional analyzer
func NewRegionalAnalyzer(store CorrelationStore) *RegionalAnalyzer {
	return &RegionalAnalyzer{
		store: store,
	}
}

// AnalyzeRegion performs comprehensive regional analysis
func (r *RegionalAnalyzer) AnalyzeRegion(regionID string, lookbackDays int) (*RegionalAnalysis, error) {
	analysis := &RegionalAnalysis{
		RegionID:   regionID,
		Timestamp:  time.Now(),
		Metrics:    make(map[models.MetricType]RegionalMetricAnalysis),
	}
}
	
	// Analyze each metric type
	metrics := []models.MetricType{
		models.MetricPaymentVolume,
		models.MetricPaymentValue,
		models.MetricEnergyConsumption,
		models.MetricRetailSales,
		models.MetricInventoryLevel,
	}
	
	for _, metric := range metrics {
		data, err := r.store.GetTimeSeriesData(metric, regionID, lookbackDays)
		if err != nil || len(data) == 0 {
			continue
		}
		
		analysis.Metrics[metric] = r.analyzeMetric(data)
	}
	
	// Calculate overall regional health
	analysis.HealthScore = r.calculateRegionalHealth(analysis.Metrics)
	analysis.Trend = r.determineTrend(analysis.Metrics)
	
	return analysis, nil
}

// analyzeMetric performs statistical analysis on a single metric
func (r *RegionalAnalyzer) analyzeMetric(data []models.TimeSeriesPoint) RegionalMetricAnalysis {
	values := make([]float64, len(data))
	for i, p := range data {
		values[i] = p.Value
	}
	
	return RegionalMetricAnalysis{
		Mean:           mean(values),
		StdDev:         standardDeviation(values),
		Min:            min(values),
		Max:            max(values),
		RecentChange:   r.calculateRecentChange(values),
		Volatility:     r.calculateVolatility(values),
		TrendDirection: r.determineTrendDirection(values),
	}
}

// calculateRecentChange calculates percentage change in recent period
func (r *RegionalAnalyzer) calculateRecentChange(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	
	// Compare last 20% with previous 20%
	n := len(values)
	split := n / 5
	
	recentAvg := mean(values[n-split:])
	previousAvg := mean(values[n-2*split : n-split])
	
	if previousAvg == 0 {
		return 0
	}
	
	return (recentAvg - previousAvg) / previousAvg
}

// calculateVolatility calculates coefficient of variation
func (r *RegionalAnalyzer) calculateVolatility(values []float64) float64 {
	avg := mean(values)
	if avg == 0 {
		return 0
	}
	
	return standardDeviation(values) / avg
}

// determineTrendDirection determines if metric is trending up, down, or stable
func (r *RegionalAnalyzer) determineTrendDirection(values []float64) string {
	if len(values) < 10 {
		return "INSUFFICIENT_DATA"
	}
	
	// Use linear regression
	n := float64(len(values))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0
	
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	avgY := sumY / n
	
	if avgY == 0 {
		return "STABLE"
	}
	
	normalizedSlope := slope / avgY
	
	switch {
	case normalizedSlope > 0.01:
		return "UP"
	case normalizedSlope < -0.01:
		return "DOWN"
	default:
		return "STABLE"
	}
}

// calculateRegionalHealth calculates overall regional health score
func (r *RegionalAnalyzer) calculateRegionalHealth(metrics map[models.MetricType]RegionalMetricAnalysis) float64 {
	if len(metrics) == 0 {
		return 50.0
	}
	
	total := 0.0
	count := 0.0
	
	for _, m := range metrics {
		// Score based on recent change and volatility
		changeScore := 50.0 + m.RecentChange*100
		
		// Lower volatility is better
		volatilityScore := 100 - m.Volatility*50
		
		combined := (changeScore + volatilityScore) / 2
		total += combined
		count++
	}
	
	return total / count
}

// determineTrend determines overall regional trend
func (r *RegionalAnalyzer) determineTrend(metrics map[models.MetricType]RegionalMetricAnalysis) string {
	upCount := 0
	downCount := 0
	
	for _, m := range metrics {
		switch m.TrendDirection {
		case "UP":
			upCount++
		case "DOWN":
			downCount++
		}
	}
	
	if len(metrics) == 0 {
		return "UNKNOWN"
	}
	
	ratio := float64(upCount-downCount) / float64(len(metrics))
	
	switch {
	case ratio > 0.3:
		return "EXPANDING"
	case ratio < -0.3:
		return "CONTRACTING"
	default:
		return "STABLE"
	}
}

// RegionalAnalysis holds comprehensive regional analysis results
type RegionalAnalysis struct {
	RegionID   string
	Timestamp  time.Time
	HealthScore float64
	Trend      string
	Metrics    map[models.MetricType]RegionalMetricAnalysis
}

// RegionalMetricAnalysis holds analysis results for a single metric
type RegionalMetricAnalysis struct {
	Mean           float64
	StdDev         float64
	Min            float64
	Max            float64
	RecentChange   float64
	Volatility     float64
	TrendDirection string
}

// Helper functions

func standardDeviation(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	avg := mean(values)
	variance := 0.0
	
	for _, v := range values {
		diff := v - avg
		variance += diff * diff
	}
	
	return math.Sqrt(variance / float64(len(values)))
}

func min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	m := values[0]
	for _, v := range values {
		if v < m {
			m = v
		}
	}
	return m
}

func max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	m := values[0]
	for _, v := range values {
		if v > m {
			m = v
		}
	}
	return m
}
