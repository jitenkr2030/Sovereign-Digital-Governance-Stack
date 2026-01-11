// NEAM Sensing Layer - Data Normalizer
// Standardizes heterogeneous data formats into unified economic signals

package normalizer

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"neam-platform/sensing/common"
	"neam-platform/sensing/payments/schemas"
)

// NormalizedEvent represents a standardized economic event after normalization
type NormalizedEvent struct {
	// Event identification
	ID          string    `json:"id"`
	SourceID    string    `json:"source_id"`
	SourceType  string    `json:"source_type"` // payment, energy, transport, industry
	EventType   string    `json:"event_type"`
	
	// Timestamp
	IngestedAt  time.Time `json:"ingested_at"`
	EventTime   time.Time `json:"event_time"`
	
	// Location
	RegionID    string  `json:"region_id"`
	DistrictID  string  `json:"district_id,omitempty"`
	Location    *Location `json:"location,omitempty"`
	
	// Economic indicators
	Category    string                 `json:"category"`
	SubCategory string                 `json:"sub_category,omitempty"`
	Metric      string                 `json:"metric"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	
	// Context
	Context     map[string]interface{} `json:"context,omitempty"`
	
	// Source reference
	SourceRef   string                 `json:"source_ref,omitempty"`
	
	// Quality indicators
	QualityScore float64               `json:"quality_score"`
	Anonymized  bool                   `json:"anonymized"`
	
	// Original data reference
	OriginalData json.RawMessage       `json:"original_data,omitempty"`
}

// Location represents geographic location
type Location struct {
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
	PinCode    string  `json:"pin_code,omitempty"`
	Zone       string  `json:"zone,omitempty"`
}

// NormalizationConfig holds configuration for the normalizer
type NormalizationConfig struct {
	AnonymizationEnabled  bool     `mapstructure:"anonymization_enabled"`
	QualityThreshold      float64  `mapstructure:"quality_threshold"`
	BatchSize             int      `mapstructure:"batch_size"`
	ProcessingWorkers     int      `mapstructure:"processing_workers"`
	RegionMappingFile     string   `mapstructure:"region_mapping_file"`
	CategoryMappingFile   string   `mapstructure:"category_mapping_file"`
}

// Normalizer provides data normalization services
type Normalizer struct {
	config      *NormalizationConfig
	logger      *common.Logger
	anonymizer  *common.Anonymizer
	regionMap   map[string]string
	categoryMap map[string]string
	
	// Input channels
	paymentChan   <-chan schemas.PaymentEvent
	energyChan    <-chan EnergyReading
	transportChan <-chan TransportEvent
	productionChan <-chan ProductionEvent
	
	// Output channel
	outputChan chan NormalizedEvent
	
	// Metrics
	metrics     NormalizerMetrics
	metricsMu   sync.RWMutex
	
	// Workers
	wg       sync.WaitGroup
	stopChan chan struct{}
}

// NormalizerMetrics holds normalizer performance metrics
type NormalizerMetrics struct {
	EventsProcessed   int64
	EventsAnonymized  int64
	EventsFiltered    int64
	ErrorsCount       int64
	AvgProcessingTime time.Duration
	LastProcessedTime time.Time
}

// EnergyReading import placeholder
type EnergyReading struct {
	ID          string
	RegionID    string
	Timestamp   time.Time
	LoadMW      float64
	// ... other fields
}

// TransportEvent import placeholder  
type TransportEvent struct {
	ID          string
	RegionID    string
	Timestamp   time.Time
	CargoWeightKG float64
	// ... other fields
}

// ProductionEvent import placeholder
type ProductionEvent struct {
	ID          string
	RegionID    string
	Timestamp   time.Time
	OutputValue float64
	// ... other fields
}

// NewNormalizer creates a new normalizer instance
func NewNormalizer(config *NormalizationConfig, logger *common.Logger, anonymizer *common.Anonymizer) *Normalizer {
	return &Normalizer{
		config:       config,
		logger:       logger,
		anonymizer:   anonymizer,
		regionMap:    make(map[string]string),
		categoryMap:  make(map[string]string),
		outputChan:   make(chan NormalizedEvent, config.BatchSize*2),
		metrics:      NormalizerMetrics{},
		stopChan:     make(chan struct{}),
	}
}

// SetInputChannels sets the input channels for different data types
func (n *Normalizer) SetInputChannels(payment <-chan schemas.PaymentEvent, energy <-chan EnergyReading, transport <-chan TransportEvent, production <-chan ProductionEvent) {
	n.paymentChan = payment
	n.energyChan = energy
	n.transportChan = transport
	n.productionChan = production
}

// Start begins the normalization process
func (n *Normalizer) Start(ctx context.Context) error {
	n.logger.Info("Starting normalizer", "workers", n.config.ProcessingWorkers)
	
	// Start processing workers
	for i := 0; i < n.config.ProcessingWorkers; i++ {
		n.wg.Add(1)
		go n.worker(i)
	}
	
	// Start metrics reporter
	go n.metricsReporter()
	
	n.logger.Info("Normalizer started")
	return nil
}

// Stop stops the normalizer
func (n *Normalizer) Stop() {
	close(n.stopChan)
	n.wg.Wait()
	close(n.outputChan)
}

// worker processes events from input channels
func (n *Normalizer) worker(id int) {
	defer n.wg.Done()
	
	for {
		select {
		case <-n.stopChan:
			return
		default:
			// Check all channels for available events
			select {
			case payment := <-n.paymentChan:
				n.normalizePaymentEvent(payment)
			case energy := <-n.energyChan:
				n.normalizeEnergyEvent(energy)
			case transport := <-n.transportChan:
				n.normalizeTransportEvent(transport)
			case production := <-n.productionChan:
				n.normalizeProductionEvent(production)
			default:
				// No events available, wait a bit
				time.Sleep(time.Millisecond)
			}
		}
	}
}

// normalizePaymentEvent normalizes a payment event
func (n *Normalizer) normalizePaymentEvent(event schemas.PaymentEvent) {
	startTime := time.Now()
	
	// Validate event
	if err := event.Validate(); err != nil {
		n.logger.Warn("Invalid payment event", "error", err)
		n.incrementErrors()
		return
	}
	
	// Calculate quality score
	qualityScore := n.calculateQualityScore(&event)
	
	// Check quality threshold
	if qualityScore < n.config.QualityThreshold {
		n.metricsMu.Lock()
		n.metrics.EventsFiltered++
		n.metricsMu.Unlock()
		return
	}
	
	// Anonymize sensitive fields if enabled
	anonymized := false
	if n.config.AnonymizationEnabled && n.anonymizer != nil {
		event.SourceAccount = n.anonymizer.HashIdentifier(event.SourceAccount)
		event.DestAccount = n.anonymizer.HashIdentifier(event.DestAccount)
		event.MerchantID = n.anonymizer.HashIdentifier(event.MerchantID)
		anonymized = true
	}
	
	// Create normalized event
	normalized := NormalizedEvent{
		ID:         fmt.Sprintf("NE-%s", event.TransactionID),
		SourceID:   event.SourceBank,
		SourceType: "payment",
		EventType:  string(event.PaymentType),
		IngestedAt: time.Now(),
		EventTime:  event.Timestamp,
		RegionID:   event.Region,
		Location: &Location{
			Latitude:  event.Latitude,
			Longitude: event.Longitude,
			PinCode:   event.PinCode,
		},
		Category:   "economic_activity",
		SubCategory: event.MerchantCategory,
		Metric:     "transaction_value",
		Value:      event.Amount,
		Unit:       event.Currency,
		Context: map[string]interface{}{
			"channel":         event.Channel,
			"payment_status":  event.Status,
			"processing_time": event.ProcessingTime,
		},
		SourceRef:   event.TransactionID,
		QualityScore: qualityScore,
		Anonymized:  anonymized,
	}
	
	// Send to output channel
	n.sendNormalizedEvent(normalized, startTime)
}

// normalizeEnergyEvent normalizes an energy reading
func (n *Normalizer) normalizeEnergyEvent(event EnergyReading) {
	startTime := time.Now()
	
	// Calculate quality score
	qualityScore := n.calculateEnergyQualityScore(&event)
	
	if qualityScore < n.config.QualityThreshold {
		n.metricsMu.Lock()
		n.metrics.EventsFiltered++
		n.metricsMu.Unlock()
		return
	}
	
	normalized := NormalizedEvent{
		ID:         fmt.Sprintf("NE-ENERGY-%s", event.ID),
		SourceID:   event.StationID,
		SourceType: "energy",
		EventType:  "telemetry",
		IngestedAt: time.Now(),
		EventTime:  event.Timestamp,
		RegionID:   event.RegionID,
		Location: &Location{
			Zone: event.ZoneID,
		},
		Category:    "energy",
		SubCategory: event.GenerationSource,
		Metric:      "grid_load",
		Value:       event.LoadMW,
		Unit:        "MW",
		Context: map[string]interface{}{
			"frequency":        event.Frequency,
			"grid_status":      event.GridStatus,
			"renewable_share":  event.RenewableShare,
			"voltage":          event.Voltage,
		},
		SourceRef:   event.ID,
		QualityScore: qualityScore,
	}
	
	n.sendNormalizedEvent(normalized, startTime)
}

// normalizeTransportEvent normalizes a transport event
func (n *Normalizer) normalizeTransportEvent(event TransportEvent) {
	startTime := time.Now()
	
	qualityScore := n.calculateTransportQualityScore(&event)
	
	if qualityScore < n.config.QualityThreshold {
		n.metricsMu.Lock()
		n.metrics.EventsFiltered++
		n.metricsMu.Unlock()
		return
	}
	
	normalized := NormalizedEvent{
		ID:         fmt.Sprintf("NE-TRANSPORT-%s", event.ID),
		SourceID:   event.PortID,
		SourceType: "transport",
		EventType:  event.EventType,
		IngestedAt: time.Now(),
		EventTime:  event.Timestamp,
		RegionID:   event.RegionID,
		Location: &Location{
			PinCode: event.DistrictID,
		},
		Category:    "logistics",
		SubCategory: event.TransportType,
		Metric:      "cargo_volume",
		Value:       event.CargoWeightKG,
		Unit:        "kg",
		Context: map[string]interface{}{
			"container_teu":   event.ContainerTEU,
			"vehicle_type":    event.VehicleType,
			"status":          event.Status,
			"processing_time": event.ProcessingTimeMin,
		},
		SourceRef:   event.ID,
		QualityScore: qualityScore,
	}
	
	n.sendNormalizedEvent(normalized, startTime)
}

// normalizeProductionEvent normalizes a production event
func (n *Normalizer) normalizeProductionEvent(event ProductionEvent) {
	startTime := time.Now()
	
	qualityScore := n.calculateProductionQualityScore(&event)
	
	if qualityScore < n.config.QualityThreshold {
		n.metricsMu.Lock()
		n.metrics.EventsFiltered++
		n.metricsMu.Unlock()
		return
	}
	
	normalized := NormalizedEvent{
		ID:         fmt.Sprintf("NE-INDUSTRY-%s", event.ID),
		SourceID:   event.FactoryID,
		SourceType: "industry",
		EventType:  "production",
		IngestedAt: time.Now(),
		EventTime:  event.Timestamp,
		RegionID:   event.RegionID,
		Category:    "industrial_production",
		SubCategory: event.Sector,
		Metric:      "output_value",
		Value:       event.OutputValue,
		Unit:        "INR",
		Context: map[string]interface{}{
			"output_quantity":  event.OutputQuantity,
			"utilization_rate": event.UtilizationRate,
			"workers":          event.WorkersEmployed,
			"energy_consumed":  event.EnergyConsumed,
		},
		SourceRef:   event.ID,
		QualityScore: qualityScore,
	}
	
	n.sendNormalizedEvent(normalized, startTime)
}

// sendNormalizedEvent sends a normalized event to the output channel
func (n *Normalizer) sendNormalizedEvent(event NormalizedEvent, startTime time.Time) {
	processingTime := time.Since(startTime)
	
	select {
	case n.outputChan <- event:
		n.metricsMu.Lock()
		n.metrics.EventsProcessed++
		n.metrics.LastProcessedTime = time.Now()
		n.metrics.AvgProcessingTime = (n.metrics.AvgProcessingTime*time.Duration(n.metrics.EventsProcessed-1) + processingTime) / time.Duration(n.metrics.EventsProcessed)
		n.metricsMu.Unlock()
	default:
		n.incrementErrors()
	}
}

// calculateQualityScore calculates quality score for payment event
func (n *Normalizer) calculateQualityScore(event interface{}) float64 {
	// Basic quality score calculation
	// In production, this would be more sophisticated
	score := 1.0
	
	// Adjust based on completeness of data
	switch e := event.(type) {
	case *schemas.PaymentEvent:
		if e.Region == "" {
			score -= 0.2
		}
		if e.MerchantCategory == "" {
			score -= 0.1
		}
		if e.SourceBank == "" {
			score -= 0.1
		}
	}
	
	if score < 0 {
		score = 0
	}
	
	return score
}

// calculateEnergyQualityScore calculates quality score for energy event
func (n *Normalizer) calculateEnergyQualityScore(event *EnergyReading) float64 {
	score := 1.0
	
	if event.RegionID == "" {
		score -= 0.2
	}
	if event.Timestamp.IsZero() {
		score -= 0.3
	}
	if event.LoadMW == 0 {
		score -= 0.2
	}
	
	if score < 0 {
		score = 0
	}
	
	return score
}

// calculateTransportQualityScore calculates quality score for transport event
func (n *Normalizer) calculateTransportQualityScore(event *TransportEvent) float64 {
	score := 1.0
	
	if event.RegionID == "" {
		score -= 0.2
	}
	if event.Timestamp.IsZero() {
		score -= 0.2
	}
	
	if score < 0 {
		score = 0
	}
	
	return score
}

// calculateProductionQualityScore calculates quality score for production event
func (n *Normalizer) calculateProductionQualityScore(event *ProductionEvent) float64 {
	score := 1.0
	
	if event.RegionID == "" {
		score -= 0.2
	}
	if event.Sector == "" {
		score -= 0.2
	}
	if event.OutputValue == 0 {
		score -= 0.1
	}
	
	if score < 0 {
		score = 0
	}
	
	return score
}

// incrementErrors increments error counter
func (n *Normalizer) incrementErrors() {
	n.metricsMu.Lock()
	defer n.metricsMu.Unlock()
	n.metrics.ErrorsCount++
}

// GetOutputChannel returns the output channel
func (n *Normalizer) GetOutputChannel() <-chan NormalizedEvent {
	return n.outputChan
}

// GetMetrics returns current metrics
func (n *Normalizer) GetMetrics() NormalizerMetrics {
	n.metricsMu.RLock()
	defer n.metricsMu.RUnlock()
	return n.metrics
}

// metricsReporter periodically logs metrics
func (n *Normalizer) metricsReporter() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-n.stopChan:
			return
		case <-ticker.C:
			metrics := n.GetMetrics()
			n.logger.Info("Normalizer metrics",
				"processed", metrics.EventsProcessed,
				"anonymized", metrics.EventsAnonymized,
				"filtered", metrics.EventsFiltered,
				"errors", metrics.ErrorsCount,
				"avg_latency_ms", metrics.AvgProcessingTime.Milliseconds(),
			)
		}
	}
}

// BatchNormalize processes events in batches
func (n *Normalizer) BatchNormalize(events []NormalizedEvent) ([]NormalizedEvent, error) {
	normalized := make([]NormalizedEvent, 0, len(events))
	
	for _, event := range events {
		if event.QualityScore >= n.config.QualityThreshold {
			normalized = append(normalized, event)
		}
	}
	
	return normalized, nil
}

// ExportMetrics exports normalizer metrics for monitoring
func (n *Normalizer) ExportMetrics() map[string]interface{} {
	metrics := n.GetMetrics()
	return map[string]interface{}{
		"events_processed":   metrics.EventsProcessed,
		"events_anonymized":  metrics.EventsAnonymized,
		"events_filtered":    metrics.EventsFiltered,
		"errors":             metrics.ErrorsCount,
		"avg_processing_time": metrics.AvgProcessingTime.String(),
		"last_processed":     metrics.LastProcessedTime,
	}
}

// ToJSON converts normalized event to JSON
func (ne *NormalizedEvent) ToJSON() string {
	data, _ := json.Marshal(ne)
	return string(data)
}

// InitializeRegionMapping initializes region code mappings
func (n *Normalizer) InitializeRegionMapping(mapping map[string]string) {
	for k, v := range mapping {
		n.regionMap[k] = v
	}
}

// InitializeCategoryMapping initializes category mappings
func (n *Normalizer) InitializeCategoryMapping(mapping map[string]string) {
	for k, v := range mapping {
		n.categoryMap[k] = v
	}
}

// MapRegion maps a region code to standard format
func (n *Normalizer) MapRegion(code string) string {
	if mapped, ok := n.regionMap[code]; ok {
		return mapped
	}
	return code
}

// MapCategory maps a category to standard format
func (n *Normalizer) MapCategory(category string) string {
	if mapped, ok := n.categoryMap[category]; ok {
		return mapped
	}
	return category
}
