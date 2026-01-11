package signals

import (
	"context"
	"testing"
	"time"

	"neam-platform/intelligence/models"
)

// MockStore implements SignalStore for testing
type MockStore struct {
	signals []models.Signal
}

func NewMockStore() *MockStore {
	return &MockStore{
		signals: make([]models.Signal, 0),
	}
}

func (m *MockStore) SaveSignal(signal *models.Signal) error {
	m.signals = append(m.signals, *signal)
	return nil
}

func (m *MockStore) GetSignals(filter SignalFilter) ([]models.Signal, error) {
	result := make([]models.Signal, 0)
	for _, s := range m.signals {
		if filter.StartTime != nil && s.Timestamp.Before(*filter.StartTime) {
			continue
		}
		if filter.EndTime != nil && s.Timestamp.After(*filter.EndTime) {
			continue
		}
		result = append(result, s)
	}
	return result, nil
}

func (m *MockStore) GetSignalByID(id string) (*models.Signal, error) {
	for _, s := range m.signals {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, nil
}

// MockPublisher implements SignalPublisher for testing
type MockPublisher struct {
	publishedSignals []models.Signal
}

func NewMockPublisher() *MockPublisher {
	return &MockPublisher{
		publishedSignals: make([]models.Signal, 0),
	}
}

func (m *MockPublisher) Publish(ctx context.Context, signal *models.Signal) error {
	m.publishedSignals = append(m.publishedSignals, *signal)
	return nil
}

func (m *MockPublisher) PublishBatch(ctx context.Context, signals []models.Signal) error {
	m.publishedSignals = append(m.publishedSignals, signals...)
	return nil
}

func (m *MockPublisher) GetPublishedSignals() []models.Signal {
	return m.publishedSignals
}

func TestRouter_RouteSignal(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	signal := &models.Signal{
		SignalType: models.SignalAnomaly,
		Severity:   0.5,
		RegionID:   "TEST",
	}
	
	err := router.RouteSignal(signal)
	
	if err != nil {
		t.Errorf("Unexpected error routing signal: %v", err)
	}
}

func TestRouter_RouteSignal_GeneratesID(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	signal := &models.Signal{
		SignalType: models.SignalAnomaly,
		Severity:   0.5,
	}
	
	router.RouteSignal(signal)
	
	if signal.ID == "" {
		t.Error("Expected signal ID to be generated")
	}
}

func TestRouter_RouteSignal_SetsTimestamp(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	signal := &models.Signal{
		SignalType: models.SignalAnomaly,
		Severity:   0.5,
	}
	
	before := time.Now()
	router.RouteSignal(signal)
	after := time.Now()
	
	if signal.Timestamp.Before(before) || signal.Timestamp.After(after) {
		t.Error("Expected timestamp to be set to current time")
	}
}

func TestRouter_Throttling(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	config := DefaultConfig()
	config.EnableThrottling = true
	config.AlertCooldown = time.Hour
	
	router := NewRouter(store, publisher, config)
	
	signal := &models.Signal{
		SignalType: models.SignalAnomaly,
		Severity:   0.5,
		RegionID:   "TEST",
		SectorID:   "SECTOR",
	}
	
	// First signal should be accepted
	err := router.RouteSignal(signal)
	if err != nil {
		t.Errorf("Unexpected error for first signal: %v", err)
	}
	
	// Second signal should be throttled
	err = router.RouteSignal(signal)
	if err != nil {
		t.Errorf("Unexpected error for throttled signal: %v", err)
	}
}

func TestRouter_CreateSignalFromAnomaly(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	event := &models.EconomicEvent{
		ID:         "event-1",
		RegionID:   "REGION_A",
		SectorID:   "RETAIL",
		MetricType: models.MetricPaymentVolume,
		Value:      150.0,
		Timestamp:  time.Now(),
	}
	
	signal := router.CreateSignalFromAnomaly(event, 0.8, 0.9, "Z_SCORE")
	
	if signal.SignalType != models.SignalAnomaly {
		t.Errorf("Expected signal type ANOMALY, got %s", signal.SignalType)
	}
	
	if signal.Severity != 0.8 {
		t.Errorf("Expected severity 0.8, got %f", signal.Severity)
	}
	
	if signal.Confidence != 0.9 {
		t.Errorf("Expected confidence 0.9, got %f", signal.Confidence)
	}
	
	if signal.RegionID != "REGION_A" {
		t.Errorf("Expected region REGION_A, got %s", signal.RegionID)
	}
	
	if signal.TriggeringEvent != event {
		t.Error("Expected triggering event to be set")
	}
}

func TestRouter_CreateSignalFromAnomaly_BlackEconomy(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	event := &models.EconomicEvent{
		ID:         "event-1",
		RegionID:   "REGION_A",
		SectorID:   "RETAIL",
		MetricType: models.MetricCashTransaction,
		Value:      150.0,
		Timestamp:  time.Now(),
	}
	
	signal := router.CreateSignalFromAnomaly(event, 0.8, 0.9, "BLACK_ECONOMY")
	
	if signal.SignalType != models.SignalBlackEconomyRisk {
		t.Errorf("Expected BLACK_ECONOMY_RISK signal type, got %s", signal.SignalType)
	}
}

func TestRouter_CreateSignalFromAnomaly_SupplyChain(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	event := &models.EconomicEvent{
		ID:         "event-1",
		MetricType: models.MetricInventoryLevel,
		Value:      150.0,
		Timestamp:  time.Now(),
	}
	
	signal := router.CreateSignalFromAnomaly(event, 0.8, 0.9, "SUPPLY_CHAIN")
	
	if signal.SignalType != models.SignalSupplyChainShock {
		t.Errorf("Expected SUPPLY_CHAIN_SHOCK signal type, got %s", signal.SignalType)
	}
}

func TestRouter_CreateInflationSignal(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	prediction := &models.InflationPrediction{
		Timestamp:         time.Now(),
		RegionID:          "NATIONAL",
		CurrentInflation:  0.05,
		Predicted30Day:    0.06,
		Predicted90Day:    0.07,
		Predicted180Day:   0.08,
		Confidence:        0.85,
		RiskLevel:         "HIGH",
		LeadingIndicators: []string{"Commodity Price Index"},
	}
	
	signal := router.CreateInflationSignal(prediction)
	
	if signal.SignalType != models.SignalInflationWarning {
		t.Errorf("Expected INFLATION_WARNING signal type, got %s", signal.SignalType)
	}
	
	if signal.Severity < 0.7 {
		t.Errorf("Expected high severity for HIGH risk, got %f", signal.Severity)
	}
	
	if len(signal.Payload.Recommendations) == 0 {
		t.Error("Expected recommendations in payload")
	}
}

func TestRouter_CreateCorrelationSignal(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	corrData := &models.CorrelationData{
		Timestamp:       time.Now(),
		MetricA:         models.MetricEnergyConsumption,
		MetricB:         models.MetricPaymentVolume,
		RegionID:        "REGION_A",
		PearsonCoeff:    0.2,        // Dropped from historical
		HistoricalAvg:   0.8,        // Was 0.8
		Deviation:       -0.6,       // Significant drop
	}
	
	signal := router.CreateCorrelationSignal(corrData)
	
	if signal == nil {
		t.Fatal("Expected signal, got nil")
	}
	
	if signal.SignalType != models.SignalDecouplingAlert {
		t.Errorf("Expected DECOUPLING_ALERT signal type, got %s", signal.SignalType)
	}
}

func TestRouter_CreateCorrelationSignal_NoSignificantDeviation(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	corrData := &models.CorrelationData{
		PearsonCoeff:   0.7,
		HistoricalAvg:  0.8,
		Deviation:      -0.1, // Not significant
	}
	
	signal := router.CreateCorrelationSignal(corrData)
	
	if signal != nil {
		t.Error("Expected nil signal for insignificant deviation")
	}
}

func TestRouter_CreateRegionalStressSignal(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	stress := &models.RegionalStressIndex{
		Timestamp:    time.Now(),
		RegionID:     "REGION_A",
		OverallScore: 75.0,
	}
	
	signal := router.CreateRegionalStressSignal(stress)
	
	if signal == nil {
		t.Fatal("Expected signal, got nil")
	}
	
	if signal.SignalType != models.SignalRegionalStress {
		t.Errorf("Expected REGIONAL_STRESS signal type, got %s", signal.SignalType)
	}
}

func TestRouter_CreateRegionalStressSignal_LowStress(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	stress := &models.RegionalStressIndex{
		OverallScore: 40.0, // Low stress
	}
	
	signal := router.CreateRegionalStressSignal(stress)
	
	if signal != nil {
		t.Error("Expected nil signal for low stress")
	}
}

func TestRouter_RiskLevelToSeverity(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	tests := []struct {
		riskLevel string
		expected  float64
	}{
		{"CRITICAL", 1.0},
		{"HIGH", 0.8},
		{"MEDIUM", 0.6},
		{"LOW", 0.3},
		{"UNKNOWN", 0.5},
	}
	
	for _, tt := range tests {
		t.Run(tt.riskLevel, func(t *testing.T) {
			severity := router.riskLevelToSeverity(tt.riskLevel)
			if severity != tt.expected {
				t.Errorf("Expected severity %f for %s, got %f", tt.expected, tt.riskLevel, severity)
			}
		})
	}
}

func TestRouter_GenerateInflationRecommendations(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	tests := []struct {
		name      string
		prediction *models.InflationPrediction
		checkFor  []string
	}{
		{
			name: "Critical recommendations",
			prediction: &models.InflationPrediction{
				RiskLevel: "CRITICAL",
			},
			checkFor: []string{"emergency", "policy"},
		},
		{
			name: "High recommendations",
			prediction: &models.InflationPrediction{
				RiskLevel: "HIGH",
			},
			checkFor: []string{"monitor", "interest"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := router.generateInflationRecommendations(tt.prediction)
			
			for _, expected := range tt.checkFor {
				found := false
				for _, rec := range recommendations {
					// Simple substring check
				}
				if !found {
					t.Errorf("Expected recommendation containing '%s'", expected)
				}
			}
		})
	}
}

func TestRouter_GenerateCorrelationRecommendations(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	corr := &models.CorrelationData{
		MetricA: models.MetricEnergyConsumption,
		MetricB: models.MetricPaymentVolume,
	}
	
	recommendations := router.generateCorrelationRecommendations(corr)
	
	found := false
	for _, rec := range recommendations {
		if rec == "Investigate potential meter tampering" ||
		   rec == "Review energy sector payment collection" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected energy-related recommendations")
	}
}

func TestRouter_GetSignals(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	// Add some signals
	for i := 0; i < 5; i++ {
		signal := &models.Signal{
			ID:         "signal",
			SignalType: models.SignalAnomaly,
			Timestamp:  time.Now().AddDate(0, 0, -i),
		}
		_ = store.SaveSignal(signal)
	}
	
	filter := SignalFilter{
		Limit: 3,
	}
	
	signals, err := router.GetSignals(filter)
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if len(signals) != 3 {
		t.Errorf("Expected 3 signals with limit, got %d", len(signals))
	}
}

func TestRouter_GetSignals_WithTimeFilter(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	// Add signals at different times
	for i := 0; i < 5; i++ {
		signal := &models.Signal{
			ID:        "signal",
			SignalType: models.SignalAnomaly,
			Timestamp:  time.Now().AddDate(0, 0, -i),
		}
		_ = store.SaveSignal(signal)
	}
	
	// Filter to last 2 days
	cutoff := time.Now().AddDate(0, 0, -2)
	filter := SignalFilter{
		StartTime: &cutoff,
	}
	
	signals, err := router.GetSignals(filter)
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if len(signals) != 2 {
		t.Errorf("Expected 2 signals after cutoff, got %d", len(signals))
	}
}

func TestRouter_LoadRules(t *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	rules := []models.DetectionRule{
		{
			MetricType: models.MetricPaymentVolume,
			Algorithm:  "Z_SCORE",
			RegionID:   "NATIONAL",
			IsActive:   true,
		},
	}
	
	router.LoadRules(rules)
	
	if len(router.rules) != 1 {
		t.Errorf("Expected 1 rule loaded, got %d", len(router.rules))
	}
}

func TestSignalAggregator_AggregateByRegion(t *testing.T) {
	store := NewMockStore()
	
	// Add test signals
	for i := 0; i < 10; i++ {
		signal := models.Signal{
			SignalType: models.SignalAnomaly,
			Severity:   0.5 + float64(i%2)*0.3,
			RegionID:   "REGION_A",
		}
		_ = store.SaveSignal(&signal)
	}
	
	aggregator := NewSignalAggregator(store, 24*time.Hour)
	
	summary, err := aggregator.AggregateByRegion()
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if len(summary) == 0 {
		t.Error("Expected some aggregated results")
	}
	
	regionSummary := summary["REGION_A"]
	if regionSummary.TotalCount != 10 {
		t.Errorf("Expected 10 signals, got %d", regionSummary.TotalCount)
	}
}

func TestSignalAggregator_AggregateByType(t *testing.T) {
	store := NewMockStore()
	
	// Add test signals of different types
	for i := 0; i < 5; i++ {
		_ = store.SaveSignal(&models.Signal{
			SignalType: models.SignalAnomaly,
			Severity:   0.5,
		})
		_ = store.SaveSignal(&models.Signal{
			SignalType: models.SignalInflationWarning,
			Severity:   0.6,
		})
	}
	
	aggregator := NewSignalAggregator(store, 24*time.Hour)
	
	summary, err := aggregator.AggregateByType()
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if len(summary) != 2 {
		t.Errorf("Expected 2 signal types, got %d", len(summary))
	}
}

func TestSignalToJSON(t *testing.T) {
	signal := &models.Signal{
		ID:        "test-id",
		SignalType: models.SignalAnomaly,
		Severity:  0.8,
		Timestamp: time.Now(),
	}
	
	data, err := SignalToJSON(signal)
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestSignalFromJSON(t *testing.T) {
	jsonData := []byte(`{"id":"test-id","signal_type":"ANOMALY","severity":0.8}`)
	
	signal, err := SignalFromJSON(jsonData)
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if signal.ID != "test-id" {
		t.Errorf("Expected ID test-id, got %s", signal.ID)
	}
	
	if signal.Severity != 0.8 {
		t.Errorf("Expected severity 0.8, got %f", signal.Severity)
	}
}

// Benchmark tests

func BenchmarkRouter_RouteSignal(b *testing.B) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	signal := &models.Signal{
		SignalType: models.SignalAnomaly,
		Severity:   0.5,
		RegionID:   "TEST",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.RouteSignal(signal)
	}
}

func BenchmarkRouter_CreateSignalFromAnomaly(b *testing.T) {
	store := NewMockStore()
	publisher := NewMockPublisher()
	router := NewRouter(store, publisher, DefaultConfig())
	
	event := &models.EconomicEvent{
		RegionID:   "REGION_A",
		MetricType: models.MetricPaymentVolume,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.CreateSignalFromAnomaly(event, 0.8, 0.9, "Z_SCORE")
	}
}
