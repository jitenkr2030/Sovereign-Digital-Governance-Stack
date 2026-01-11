package signals

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"neam-platform/intelligence/models"

	"github.com/google/uuid"
)

// Router manages signal generation, routing, and delivery
type Router struct {
	store       SignalStore
	publisher   SignalPublisher
	config      *Config
	mu          sync.RWMutex
	rules       map[string]*models.DetectionRule
	lastSignals map[string]time.Time
	alertQueue  chan *models.Signal
	running     bool
	stopChan    chan struct{}
}

// SignalStore interface for signal persistence
type SignalStore interface {
	SaveSignal(signal *models.Signal) error
	GetSignals(filter SignalFilter) ([]models.Signal, error)
	GetSignalByID(id string) (*models.Signal, error)
}

// SignalPublisher interface for signal delivery
type SignalPublisher interface {
	Publish(ctx context.Context, signal *models.Signal) error
	PublishBatch(ctx context.Context, signals []models.Signal) error
}

// Config holds router configuration
type Config struct {
	AlertCooldown         time.Duration
	HighSeverityThreshold float64
	BatchSize             int
	BatchInterval         time.Duration
	MaxQueueSize          int
	EnableThrottling      bool
	EnableDeduplication   bool
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		AlertCooldown:         time.Hour,
		HighSeverityThreshold: 0.75,
		BatchSize:             100,
		BatchInterval:         time.Minute,
		MaxQueueSize:          10000,
		EnableThrottling:      true,
		EnableDeduplication:   true,
	}
}

// SignalFilter defines filtering criteria for signal queries
type SignalFilter struct {
	StartTime     *time.Time
	EndTime       *time.Time
	SignalTypes   []models.SignalType
	RegionIDs     []string
	SeverityMin   *float64
	SeverityMax   *float64
	Limit         int
	Offset        int
}

// NewRouter creates a new signal router
func NewRouter(store SignalStore, publisher SignalPublisher, config *Config) *Router {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &Router{
		store:        store,
		publisher:    publisher,
		config:       config,
		rules:        make(map[string]*models.DetectionRule),
		lastSignals:  make(map[string]time.Time),
		alertQueue:   make(chan *models.Signal, 10000),
		stopChan:     make(chan struct{}),
	}
}

// Start begins the router's processing loop
func (r *Router) Start(ctx context.Context) {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return
	}
	r.running = true
	r.mu.Unlock()

	// Start batch processor
	batchTicker := time.NewTicker(r.config.BatchInterval)
	defer batchTicker.Stop()

	// Start queue processor
	for {
		select {
		case <-ctx.Done():
			r.flushQueue()
			return
		case <-r.stopChan:
			r.flushQueue()
			return
		case signal := <-r.alertQueue:
			r.processSignal(signal)
		case <-batchTicker.C:
			r.processBatch()
		}
	}
}

// Stop halts the router
func (r *Router) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.running {
		close(r.stopChan)
		r.running = false
	}
}

// RouteSignal receives and routes a signal
func (r *Router) RouteSignal(signal *models.Signal) error {
	// Generate ID if not present
	if signal.ID == "" {
		signal.ID = uuid.New().String()
	}
	
	// Set timestamp if not present
	if signal.Timestamp.IsZero() {
		signal.Timestamp = time.Now()
	}
	
	// Check throttling
	if r.config.EnableThrottling {
		if r.isThrottled(signal) {
			return nil // Signal suppressed
		}
	}
	
	// Add to queue
	select {
	case r.alertQueue <- signal:
		return nil
	default:
		// Queue full, log and drop
		return fmt.Errorf("signal queue full")
	}
}

// processSignal handles a single signal
func (r *Router) processSignal(signal *models.Signal) {
	// Save to store
	if err := r.store.SaveSignal(signal); err != nil {
		// Log error but continue
	}
	
	// Update throttling map
	key := r.getThrottleKey(signal)
	r.lastSignals[key] = time.Now()
	
	// Publish based on severity
	if signal.Severity >= r.config.HighSeverityThreshold {
		_ = r.publisher.Publish(context.Background(), signal)
	}
}

// processBatch handles batched signals
func (r *Router) processBatch() {
	// Collect signals from queue up to batch size
	batch := make([]models.Signal, 0, r.config.BatchSize)
	
	for len(batch) < r.config.BatchSize && len(r.alertQueue) > 0 {
		select {
		case signal := <-r.alertQueue:
			batch = append(batch, *signal)
		default:
			break
		}
	}
	
	if len(batch) > 0 {
		_ = r.publisher.PublishBatch(context.Background(), batch)
	}
}

// flushQueue drains the signal queue
func (r *Router) flushQueue() {
	close(r.alertQueue)
	
	for signal := range r.alertQueue {
		r.processSignal(signal)
	}
}

// isThrottled checks if a signal should be throttled
func (r *Router) isThrottled(signal *models.Signal) bool {
	key := r.getThrottleKey(signal)
	
	r.mu.RLock()
	lastTime, exists := r.lastSignals[key]
	r.mu.RUnlock()
	
	if !exists {
		return false
	}
	
	return time.Since(lastTime) < r.config.AlertCooldown
}

// getThrottleKey generates a unique key for throttling
func (r *Router) getThrottleKey(signal *models.Signal) string {
	return fmt.Sprintf("%s:%s:%s", signal.SignalType, signal.RegionID, signal.SectorID)
}

// LoadRules loads detection rules from the store
func (r *Router) LoadRules(rules []models.DetectionRule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for i := range rules {
		key := fmt.Sprintf("%s:%s:%s", rules[i].MetricType, rules[i].RegionID, rules[i].SectorID)
		r.rules[key] = &rules[i]
	}
}

// CreateSignalFromAnomaly creates a signal from anomaly detection results
func (r *Router) CreateSignalFromAnomaly(
	event *models.EconomicEvent,
	anomalyScore float64,
	confidence float64,
	detectorName string,
) *models.Signal {
	
	signalType := models.SignalAnomaly
	payload := models.SignalPayload{
		AnomalyScore: anomalyScore,
	}
	
	// Add specific context based on metric type
	switch event.MetricType {
	case models.MetricCashTransaction, models.MetricDigitalTransaction:
		signalType = models.SignalBlackEconomyRisk
	case models.MetricInventoryLevel:
		signalType = models.SignalSupplyChainShock
	case models.MetricEnergyConsumption:
		signalType = models.SignalDecouplingAlert
	}
	
	return &models.Signal{
		ID:              uuid.New().String(),
		Timestamp:       time.Now(),
		SignalType:      signalType,
		Severity:        anomalyScore,
		Confidence:      confidence,
		RegionID:        event.RegionID,
		SectorID:        event.SectorID,
		TriggeringEvent: event,
		Payload:         payload,
		Metadata: map[string]interface{}{
			"detector":    detectorName,
			"metric_type": event.MetricType,
			"source_id":   event.SourceID,
		},
	}
}

// CreateInflationSignal creates a signal from inflation prediction
func (r *Router) CreateInflationSignal(prediction *models.InflationPrediction) *models.Signal {
	
	signalType := models.SignalInflationWarning
	riskLevel := prediction.RiskLevel
	
	// Determine signal type based on risk level
	if riskLevel == "CRITICAL" || riskLevel == "HIGH" {
		signalType = models.SignalInflationWarning
	} else if riskLevel == "LOW" {
		// Could generate positive signals for low inflation
		signalType = models.SignalSlowdownWarning
	}
	
	return &models.Signal{
		ID:         uuid.New().String(),
		Timestamp:  time.Now(),
		SignalType: signalType,
		Severity:   r.riskLevelToSeverity(riskLevel),
		Confidence: prediction.Confidence,
		RegionID:   prediction.RegionID,
		Payload: models.SignalPayload{
			Prediction:       prediction.CurrentInflation,
			PredictionWindow: "current",
			LeadingIndicators: prediction.LeadingIndicators,
			Recommendations:  r.generateInflationRecommendations(prediction),
		},
		Metadata: map[string]interface{}{
			"predicted_30_day":  prediction.Predicted30Day,
			"predicted_90_day":  prediction.Predicted90Day,
			"risk_level":        riskLevel,
		},
	}
}

// CreateCorrelationSignal creates a signal from correlation analysis
func (r *Router) CreateCorrelationSignal(correlation *models.CorrelationData) *models.Signal {
	
	// Only create signals for significant correlation breaks
	if correlation.Deviation > -0.3 {
		return nil
	}
	
	signalType := models.SignalCorrelationBreak
	
	// Check for decoupling specifically
	if correlation.PearsonCoeff < 0.3 && correlation.HistoricalAvg > 0.7 {
		signalType = models.SignalDecouplingAlert
	}
	
	// Determine severity based on deviation magnitude
	severity := min(1.0, -correlation.Deviation)
	
	return &models.Signal{
		ID:         uuid.New().String(),
		Timestamp:  time.Now(),
		SignalType: signalType,
		Severity:   severity,
		Confidence: 0.8,
		RegionID:   correlation.RegionID,
		Payload: models.SignalPayload{
			CorrelationCoeff: correlation.PearsonCoeff,
			Recommendations:  r.generateCorrelationRecommendations(correlation),
		},
		Metadata: map[string]interface{}{
			"metric_a":           correlation.MetricA,
			"metric_b":           correlation.MetricB,
			"historical_avg":     correlation.HistoricalAvg,
			"deviation":          correlation.Deviation,
			"window_days":        correlation.WindowDays,
		},
	}
}

// CreateRegionalStressSignal creates a signal from regional stress analysis
func (r *Router) CreateRegionalStressSignal(index *models.RegionalStressIndex) *models.Signal {
	
	if index.OverallScore < 60 {
		return nil // No significant stress
	}
	
	signalType := models.SignalRegionalStress
	severity := index.OverallScore / 100
	
	return &models.Signal{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		SignalType: signalType,
		Severity:  severity,
		Confidence: 0.85,
		RegionID:  index.RegionID,
		Payload: models.SignalPayload{
			AnomalyScore: index.AnomalyScore,
		},
		Metadata: map[string]interface{}{
			"overall_score":      index.OverallScore,
			"anomaly_count":      index.AnomalyCount,
			"correlation_breaks": index.CorrelationBreaks,
		},
	}
}

// GetSignals retrieves signals based on filter criteria
func (r *Router) GetSignals(filter SignalFilter) ([]models.Signal, error) {
	return r.store.GetSignals(filter)
}

// GetSignalByID retrieves a specific signal
func (r *Router) GetSignalByID(id string) (*models.Signal, error) {
	return r.store.GetSignalByID(id)
}

// riskLevelToSeverity converts risk level to severity score
func (r *Router) riskLevelToSeverity(riskLevel string) float64 {
	switch riskLevel {
	case "CRITICAL":
		return 1.0
	case "HIGH":
		return 0.8
	case "MEDIUM":
		return 0.6
	case "LOW":
		return 0.3
	default:
		return 0.5
	}
}

// generateInflationRecommendations generates recommendations based on inflation prediction
func (r *Router) generateInflationRecommendations(prediction *models.InflationPrediction) []string {
	recommendations := make([]string, 0)
	
	switch prediction.RiskLevel {
	case "CRITICAL":
		recommendations = append(recommendations,
			"Consider emergency monetary policy measures",
			"Review price controls for essential goods",
			"Prepare public communication strategy")
	case "HIGH":
		recommendations = append(recommendations,
			"Monitor key commodity prices closely",
			"Review interest rate policy stance",
			"Assess impact on vulnerable populations")
	case "MEDIUM":
		recommendations = append(recommendations,
			"Continue monitoring trend",
			"Review supply chain pressures")
	case "LOW":
		recommendations = append(recommendations,
			"Maintain current policy stance",
			"Focus on structural reforms")
	}
	
	return recommendations
}

// generateCorrelationRecommendations generates recommendations for correlation issues
func (r *Router) generateCorrelationRecommendations(correlation *models.CorrelationData) []string {
	recommendations := make([]string, 0)
	
	// Energy-payment decoupling
	if (correlation.MetricA == models.MetricEnergyConsumption && correlation.MetricB == models.MetricPaymentVolume) ||
	   (correlation.MetricB == models.MetricEnergyConsumption && correlation.MetricA == models.MetricPaymentVolume) {
		recommendations = append(recommendations,
			"Investigate potential meter tampering",
			"Review energy sector payment collection",
			"Check for industrial consumption anomalies")
	}
	
	// Inventory-price decoupling
	if correlation.MetricA == models.MetricInventoryLevel || correlation.MetricB == models.MetricInventoryLevel {
		recommendations = append(recommendations,
			"Assess supply chain disruption extent",
			"Review inventory management practices",
			"Monitor for hoarding behavior")
	}
	
	return recommendations
}

// SignalAggregator aggregates signals for reporting
type SignalAggregator struct {
	store  SignalStore
	window time.Duration
	mu     sync.RWMutex
}

// NewSignalAggregator creates a new signal aggregator
func NewSignalAggregator(store SignalStore, window time.Duration) *SignalAggregator {
	return &SignalAggregator{
		store:  store,
		window: window,
	}
}

// AggregateByRegion aggregates signals by region
func (a *SignalAggregator) AggregateByRegion() (map[string]SignalSummary, error) {
	now := time.Now()
	start := now.Add(-a.window)
	
	startTime := start
	filter := SignalFilter{
		StartTime: &startTime,
		Limit:     10000,
	}
	
	signals, err := a.store.GetSignals(filter)
	if err != nil {
		return nil, err
	}
	
	summary := make(map[string]SignalSummary)
	
	for _, signal := range signals {
		s := summary[signal.RegionID]
		s.RegionID = signal.RegionID
		s.TotalCount++
		s.TotalSeverity += signal.Severity
		
		switch signal.SignalType {
		case models.SignalAnomaly:
			s.AnomalyCount++
		case models.SignalInflationWarning:
			s.InflationSignals++
		case models.SignalCorrelationBreak, models.SignalDecouplingAlert:
			s.CorrelationIssues++
		}
		
		summary[signal.RegionID] = s
	}
	
	// Calculate averages
	for id, s := range summary {
		if s.TotalCount > 0 {
			s.AvgSeverity = s.TotalSeverity / float64(s.TotalCount)
		}
		summary[id] = s
	}
	
	return summary, nil
}

// AggregateByType aggregates signals by type
func (a *SignalAggregator) AggregateByType() (map[models.SignalType]SignalSummary, error) {
	now := time.Now()
	start := now.Add(-a.window)
	
	startTime := start
	filter := SignalFilter{
		StartTime: &startTime,
		Limit:     10000,
	}
	
	signals, err := a.store.GetSignals(filter)
	if err != nil {
		return nil, err
	}
	
	summary := make(map[models.SignalType]SignalSummary)
	
	for _, signal := range signals {
		s := summary[signal.SignalType]
		s.SignalType = signal.SignalType
		s.TotalCount++
		s.TotalSeverity += signal.Severity
		s.AvgSeverity = s.TotalSeverity / float64(s.TotalCount)
		
		summary[signal.SignalType] = s
	}
	
	return summary, nil
}

// SignalSummary holds aggregated signal statistics
type SignalSummary struct {
	SignalType       models.SignalType
	RegionID         string
	TotalCount       int
	AnomalyCount     int
	InflationSignals int
	CorrelationIssues int
	TotalSeverity    float64
	AvgSeverity      float64
}

// SignalToJSON converts a signal to JSON bytes
func SignalToJSON(signal *models.Signal) ([]byte, error) {
	return json.Marshal(signal)
}

// SignalFromJSON parses JSON bytes into a signal
func SignalFromJSON(data []byte) (*models.Signal, error) {
	var signal models.Signal
	if err := json.Unmarshal(data, &signal); err != nil {
		return nil, err
	}
	return &signal, nil
}

// Helper functions

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
