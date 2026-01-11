package intelligence

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"neam-platform/intelligence/anomaly"
	"neam-platform/intelligence/correlation"
	"neam-platform/intelligence/inflation"
	"neam-platform/intelligence/models"
	"neam-platform/intelligence/signals"
)

// Engine is the main orchestrator for the Economic Intelligence Layer
type Engine struct {
	config           *Config
	anomalyEngine    *AnomalyProcessor
	inflationEngine  *inflation.Engine
	correlationEngine *correlation.Engine
	signalRouter     *signals.Router
	store            DataStore
	consumer         MessageConsumer
	mu               sync.RWMutex
	running          bool
	stopChan         chan struct{}
	wg               sync.WaitGroup
}

// Config holds engine-level configuration
type Config struct {
	AnomalyEnabled      bool
	InflationEnabled    bool
	CorrelationEnabled  bool
	RegionID            string
	Workers             int
	KafkaBrokers        []string
	KafkaConsumerGroup  string
	InputTopics         []string
	OutputTopic         string
	UpdateInterval      time.Duration
	HistoryLookbackDays int
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		AnomalyEnabled:     true,
		InflationEnabled:   true,
		CorrelationEnabled: true,
		RegionID:           "NATIONAL",
		Workers:            4,
		KafkaBrokers:       []string{"localhost:9092"},
		KafkaConsumerGroup: "neam-intelligence",
		InputTopics:        []string{
			"neam.payment.events",
			"neam.energy.telemetry",
			"neam.price.data",
		},
		OutputTopic:       "neam.intelligence.signals",
		UpdateInterval:    time.Minute,
		HistoryLookbackDays: 30,
	}
}

// DataStore interface for data persistence
type DataStore interface {
	// Inflation store methods
	GetCPIHistory(regionID string, lookbackDays int) ([]models.InflationData, error)
	GetCommodityPrices(lookbackDays int) ([]models.TimeSeriesPoint, error)
	GetProducerPrices(regionID string, lookbackDays int) ([]models.TimeSeriesPoint, error)
	SavePrediction(prediction *models.InflationPrediction) error
	
	// Correlation store methods
	GetTimeSeriesData(metricType models.MetricType, regionID string, lookbackDays int) ([]models.TimeSeriesPoint, error)
	SaveCorrelationData(data *models.CorrelationData) error
	GetRegionalStressIndex(regionID string) (*models.RegionalStressIndex, error)
	SaveRegionalStressIndex(index *models.RegionalStressIndex) error
	
	// Signal store methods
	SaveSignal(signal *models.Signal) error
	GetSignals(filter signals.SignalFilter) ([]models.Signal, error)
	GetSignalByID(id string) (*models.Signal, error)
}

// MessageConsumer interface for Kafka consumption
type MessageConsumer interface {
	Consume(ctx context.Context, topic string, handler func(message []byte) error) error
}

// MessagePublisher interface for Kafka publishing
type MessagePublisher interface {
	Publish(ctx context.Context, topic string, message []byte) error
}

// NewEngine creates a new intelligence engine
func NewEngine(
	store DataStore,
	consumer MessageConsumer,
	publisher MessagePublisher,
	config *Config,
) *Engine {
	if config == nil {
		config = DefaultConfig()
	}
	
	// Create components
	signalRouter := signals.NewRouter(store, &KafkaPublisher{publisher: publisher}, nil)
	
	anomalyEngine := NewAnomalyProcessor(config, store, signalRouter)
	
	inflEngine := inflation.NewEngine(store, &inflation.Config{
		RegionID:            config.RegionID,
		HistoryLookbackDays: config.HistoryLookbackDays,
	})
	
	corrEngine := correlation.NewEngine(store, &correlation.Config{
		RegionID:    config.RegionID,
		LookbackDays: 7,
	})
	
	return &Engine{
		config:           config,
		anomalyEngine:    anomalyEngine,
		inflationEngine:  inflEngine,
		correlationEngine: corrEngine,
		signalRouter:     signalRouter,
		store:            store,
		consumer:         consumer,
		stopChan:         make(chan struct{}),
	}
}

// Start begins all intelligence processing
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return fmt.Errorf("engine already running")
	}
	e.running = true
	e.mu.Unlock()
	
	log.Println("Starting NEAM Intelligence Engine...")
	
	// Start signal router
	go e.signalRouter.Start(ctx)
	
	// Start component engines
	if e.config.AnomalyEnabled {
		e.wg.Add(1)
		go e.anomalyEngine.Start(ctx)
	}
	
	if e.config.InflationEnabled {
		e.wg.Add(1)
		go e.inflationEngine.Start()
	}
	
	if e.config.CorrelationEnabled {
		e.wg.Add(1)
		go e.correlationEngine.Start()
	}
	
	// Start message consumption
	if e.consumer != nil {
		e.wg.Add(len(e.config.InputTopics))
		for _, topic := range e.config.InputTopics {
			go e.consumeTopic(ctx, topic)
		}
	}
	
	log.Println("NEAM Intelligence Engine started successfully")
	return nil
}

// Stop halts all processing
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if !e.running {
		return
	}
	
	log.Println("Stopping NEAM Intelligence Engine...")
	
	close(e.stopChan)
	e.wg.Wait()
	
	e.anomalyEngine.Stop()
	e.inflationEngine.Stop()
	e.correlationEngine.Stop()
	e.signalRouter.Stop()
	
	e.running = false
	log.Println("NEAM Intelligence Engine stopped")
}

// consumeTopic starts consuming messages from a Kafka topic
func (e *Engine) consumeTopic(ctx context.Context, topic string) {
	defer e.wg.Done()
	
	err := e.consumer.Consume(ctx, topic, func(message []byte) error {
		var event models.EconomicEvent
		if err := unmarshalEvent(message, &event); err != nil {
			return err
		}
		
		// Route to anomaly detection
		if e.config.AnomalyEnabled {
			e.anomalyEngine.ProcessEvent(&event)
		}
		
		// Route to correlation engine
		if e.config.CorrelationEnabled {
			e.correlationEngine.RunAnalysis()
		}
		
		return nil
	})
	
	if err != nil && ctx.Err() == nil {
		log.Printf("Error consuming from topic %s: %v", topic, err)
	}
}

// RunAnalysis triggers a manual analysis cycle
func (e *Engine) RunAnalysis() error {
	var wg sync.WaitGroup
	errors := make(chan error, 3)
	
	if e.config.AnomalyEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := e.anomalyEngine.RunAnalysis(); err != nil {
				errors <- err
			}
		}()
	}
	
	if e.config.InflationEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := e.inflationEngine.RunAnalysis(); err != nil {
				errors <- err
			}
		}()
	}
	
	if e.config.CorrelationEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := e.correlationEngine.RunAnalysis(); err != nil {
				errors <- err
			}
		}()
	}
	
	wg.Wait()
	close(errors)
	
	for err := range errors {
		return err
	}
	
	return nil
}

// GetStatus returns the current engine status
func (e *Engine) GetStatus() EngineStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	return EngineStatus{
		Running:           e.running,
		AnomalyEnabled:    e.config.AnomalyEnabled,
		InflationEnabled:  e.config.InflationEnabled,
		CorrelationEnabled: e.config.CorrelationEnabled,
		RegionID:          e.config.RegionID,
	}
}

// EngineStatus holds engine status information
type EngineStatus struct {
	Running            bool
	AnomalyEnabled     bool
	InflationEnabled   bool
	CorrelationEnabled bool
	RegionID           string
}

// AnomalyProcessor handles real-time anomaly detection
type AnomalyProcessor struct {
	config     *Config
	detectors  map[models.MetricType][]anomaly.Detector
	router     *signals.Router
	store      DataStore
	dataBuffer map[models.MetricType][]models.TimeSeriesPoint
	mu         sync.RWMutex
	running    bool
	stopChan   chan struct{}
}

// NewAnomalyProcessor creates a new anomaly processor
func NewAnomalyProcessor(config *Config, store DataStore, router *signals.Router) *AnomalyProcessor {
	processor := &AnomalyProcessor{
		config:     config,
		router:     router,
		store:      store,
		dataBuffer: make(map[models.MetricType][]models.TimeSeriesPoint),
		stopChan:   make(chan struct{}),
	}
	
	// Initialize detectors for each metric type
	processor.detectors = map[models.MetricType][]anomaly.Detector{
		models.MetricPaymentVolume: {
			anomaly.NewZScoreDetector(3.0, 30),
			anomaly.NewSpendingSurgeDetector(0.5, 3, 14),
			anomaly.NewSpendingDropDetector(0.3, 14, 7),
		},
		models.MetricPaymentValue: {
			anomaly.NewZScoreDetector(3.0, 30),
			anomaly.NewMovingAverageDetector(30, 2.5),
		},
		models.MetricEnergyConsumption: {
			anomaly.NewZScoreDetector(3.0, 30),
			anomaly.NewExponentialSmoothingDetector(0.3, 0.1, 2.0),
		},
		models.MetricRetailSales: {
			anomaly.NewIQRDetector(1.5, 30),
			anomaly.NewSpendingDropDetector(0.25, 14, 7),
		},
		models.MetricCashTransaction: {
			anomaly.NewZScoreDetector(2.5, 14),
		},
		models.MetricDigitalTransaction: {
			anomaly.NewZScoreDetector(3.0, 30),
		},
		models.MetricInventoryLevel: {
			anomaly.NewMovingAverageDetector(14, 2.0),
			anomaly.NewIQRDetector(1.5, 30),
		},
	}
	
	return processor
}

// Start begins anomaly processing
func (a *AnomalyProcessor) Start(ctx context.Context) {
	defer a.cleanup()
	
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopChan:
			return
		case <-ticker.C:
			a.RunAnalysis()
		}
	}
}

// Stop halts anomaly processing
func (a *AnomalyProcessor) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.running {
		close(a.stopChan)
		a.running = false
	}
}

// ProcessEvent processes a single economic event
func (a *AnomalyProcessor) ProcessEvent(event *models.EconomicEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Add to buffer
	point := models.TimeSeriesPoint{
		Timestamp: event.Timestamp,
		Value:     event.Value,
	}
	a.dataBuffer[event.MetricType] = append(a.dataBuffer[event.MetricType], point)
	
	// Keep buffer size manageable
	maxBufferSize := 365 // One year of daily data
	for metricType, buffer := range a.dataBuffer {
		if len(buffer) > maxBufferSize {
			a.dataBuffer[metricType] = buffer[len(buffer)-maxBufferSize:]
		}
	}
	
	// Get detectors for this metric type
	detectors, ok := a.detectors[event.MetricType]
	if !ok || len(detectors) == 0 {
		return
	}
	
	// Run detection
	history := a.dataBuffer[event.MetricType]
	for _, detector := range detectors {
		isAnomaly, score, confidence := detector.Detect(history, event.Value)
		
		if isAnomaly {
			signal := a.router.CreateSignalFromAnomaly(
				event, score, confidence, detector.GetName(),
			)
			_ = a.router.RouteSignal(signal)
		}
	}
}

// RunAnalysis performs a full analysis cycle
func (a *AnomalyProcessor) RunAnalysis() error {
	// This would typically fetch data from store and process
	// For now, it's triggered by event processing
	return nil
}

func (a *AnomalyProcessor) cleanup() {
	a.Stop()
}

// KafkaPublisher implements SignalPublisher for Kafka
type KafkaPublisher struct {
	publisher MessagePublisher
}

// Publish sends a signal to Kafka
func (k *KafkaPublisher) Publish(ctx context.Context, signal *models.Signal) error {
	data, err := signals.SignalToJSON(signal)
	if err != nil {
		return err
	}
	
	// This would use the configured output topic
	return k.publisher.Publish(ctx, "neam.intelligence.signals", data)
}

// PublishBatch sends multiple signals to Kafka
func (k *KafkaPublisher) PublishBatch(ctx context.Context, signals []models.Signal) error {
	for _, signal := range signals {
		if err := k.Publish(ctx, &signal); err != nil {
			return err
		}
	}
	return nil
}

// unmarshalEvent unmarshals JSON into an EconomicEvent
func unmarshalEvent(data []byte, event *models.EconomicEvent) error {
	return json.Unmarshal(data, event)
}

// CreateRegionalReport generates a comprehensive regional analysis report
func (e *Engine) CreateRegionalReport(regionID string, lookbackDays int) (*RegionalReport, error) {
	report := &RegionalReport{
		RegionID:     regionID,
		GeneratedAt:  time.Now(),
		LookbackDays: lookbackDays,
		Metrics:      make(map[models.MetricType]MetricReport),
	}
	
	// Get inflation prediction
	inflationData, err := e.store.GetCPIHistory(regionID, lookbackDays)
	if err == nil && len(inflationData) > 0 {
		latestInflation := inflationData[len(inflationData)-1]
		report.InflationRate = latestInflation.CPIYoY
	}
	
	// Get signals
	now := time.Now()
	start := now.AddDate(0, 0, -lookbackDays)
	startTime := start
	signals, err := e.store.GetSignals(signals.SignalFilter{
		StartTime: &startTime,
		RegionIDs: []string{regionID},
		Limit:     1000,
	})
	if err == nil {
		report.TotalSignals = len(signals)
		for _, sig := range signals {
			report.TotalSeverity += sig.Severity
		}
		if report.TotalSignals > 0 {
			report.AvgSeverity = report.TotalSeverity / float64(report.TotalSignals)
		}
	}
	
	// Get regional stress
	stress, err := e.store.GetRegionalStressIndex(regionID)
	if err == nil && stress != nil {
		report.StressIndex = stress.OverallScore
	}
	
	return report, nil
}

// RegionalReport holds comprehensive regional analysis
type RegionalReport struct {
	RegionID       string
	GeneratedAt    time.Time
	LookbackDays   int
	InflationRate  float64
	TotalSignals   int
	TotalSeverity  float64
	AvgSeverity    float64
	StressIndex    float64
	Metrics        map[models.MetricType]MetricReport
}

// MetricReport holds analysis for a specific metric
type MetricReport struct {
	MetricType     models.MetricType
	Mean           float64
	StdDev         float64
	Trend          string
	AnomalyCount   int
	LastValue      float64
	LastUpdate     time.Time
}
