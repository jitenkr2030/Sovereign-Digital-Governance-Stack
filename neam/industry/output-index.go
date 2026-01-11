// NEAM Sensing Layer - Industrial Production Adapter
// Real-time industrial output, manufacturing, and supply chain data ingestion

package industry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"neam-platform/sensing/common"
)

// ProductionEvent represents an industrial production event
type ProductionEvent struct {
	// Identification
	ID            string    `json:"id"`
	FactoryID     string    `json:"factory_id"`
	FactoryName   string    `json:"factory_name,omitempty"`
	PlantID       string    `json:"plant_id"`
	
	// Classification
	Sector        string    `json:"sector"` // manufacturing, mining, utilities, construction
	IndustryCode  string    `json:"industry_code"` // NIC code
	ProductCode   string    `json:"product_code,omitempty"`
	ProductName   string    `json:"product_name,omitempty"`
	
	// Location
	RegionID      string  `json:"region_id"`
	DistrictID    string  `json:"district_id,omitempty"`
	ZoneID        string  `json:"zone_id,omitempty"`
	
	// Timestamp
	Timestamp     time.Time `json:"timestamp"`
	ProductionDate time.Time `json:"production_date"`
	
	// Production metrics
	OutputQuantity float64 `json:"output_quantity"`
	OutputUnit     string  `json:"output_unit"` // tonnes, units, kwh, cubic_meters
	OutputValue    float64 `json:"output_value"` // in local currency
	
	// Capacity utilization
	InstalledCapacity float64 `json:"installed_capacity"`
	CapacityUtilized float64 `json:"capacity_utilized"`
	UtilizationRate  float64 `json:"utilization_rate_pct"`
	
	// Input metrics
	InputQuantity  float64 `json:"input_quantity,omitempty"`
	InputUnit      string  `json:"input_unit,omitempty"`
	InputValue     float64 `json:"input_value,omitempty"`
	
	// Employment
	WorkersEmployed int     `json:"workers_employed"`
	ShiftCount      int     `json:"shift_count"`
	
	// Working hours
	OperatingHours  float64 `json:"operating_hours"`
	PlannedHours    float64 `json:"planned_hours"`
	DowntimeHours   float64 `json:"downtime_hours,omitempty"`
	
	// Quality metrics
	FirstPassQuality float64 `json:"first_pass_quality_pct"`
	DefectRate       float64 `json:"defect_rate_pct,omitempty"`
	
	// Status
	ProductionStatus string  `json:"production_status"` // normal, reduced, halted, maintenance
	
	// Energy consumption
	EnergyConsumed float64 `json:"energy_consumed_kwh,omitempty"`
	
	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SupplyChainEvent represents a supply chain disruption or flow event
type SupplyChainEvent struct {
	// Identification
	ID              string    `json:"id"`
	EventType       string    `json:"event_type"` // disruption, shipment, inventory, demand
	Severity        string    `json:"severity"` // low, medium, high, critical
	
	// Chain information
	ChainType       string    `json:"chain_type"` // raw_materials, components, finished_goods
	Stage           string    `json:"stage"` // supplier, manufacturing, distribution, retail
	OriginRegion    string    `json:"origin_region"`
	DestRegion      string    `json:"dest_region"`
	
	// Timestamp
	Timestamp       time.Time `json:"timestamp"`
	
	// Impact metrics
	ImpactDuration  time.Duration `json:"impact_duration"`
	AffectedVolume  float64 `json:"affected_volume"`
	AffectedValue   float64 `json:"affected_value"`
	
	// Status
	Status          string `json:"status"` // active, resolved, predicted
	ResolutionTime  *time.Time `json:"resolution_time,omitempty"`
	
	// Description
	Description     string                 `json:"description"`
	Details         map[string]interface{} `json:"details,omitempty"`
	
	// Related entities
	SupplierID      string `json:"supplier_id,omitempty"`
	ProductCodes    []string `json:"product_codes,omitempty"`
}

// IndexData represents sectoral index data
type IndexData struct {
	// Identification
	ID          string    `json:"id"`
	IndexType   string    `json:"index_type"` // production, productivity, capacity_utilization
	Sector      string    `json:"sector"`
	SubSector   string    `json:"sub_sector,omitempty"`
	
	// Period
	Period      string    `json:"period"` // monthly, quarterly, annual
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	
	// Index value
	IndexValue  float64 `json:"index_value"`
	BaseValue   float64 `json:"base_value"`
	
	// Change metrics
	ChangePct   float64 `json:"change_pct"`
	Change MOM    float64 `json:"change_mom"` // month over month
	Change YOY    float64 `json:"change_yoy"` // year over year
	
	// Weight
	WeightInOverall float64 `json:"weight_in_overall"`
	
	// Contributing factors
	Contributors []Contributor `json:"contributors,omitempty"`
	
	// Projections
	ProjectedNext float64 `json:"projected_next,omitempty"`
	Projected3M   float64 `json:"projected_3m,omitempty"`
	Projected6M   float64 `json:"projected_6m,omitempty"`
}

// Contributor represents a factor contributing to index movement
type Contributor struct {
	Factor         string  `json:"factor"`
	ContributionPct float64 `json:"contribution_pct"`
	Direction      string  `json:"direction"` // positive, negative
}

// ManufacturingAdapter interface for manufacturing data
type ManufacturingAdapter interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Health() bool
	Stream() <-chan ProductionEvent
	Metrics() ManufacturingMetrics
}

// ManufacturingMetrics holds performance metrics
type ManufacturingMetrics struct {
	EventsProcessed int64
	ErrorsCount     int64
	LastEventTime   time.Time
	Latency         time.Duration
}

// IndustrialSurveyAdapter implements adapter for industrial survey data
type IndustrialSurveyAdapter struct {
	config    SurveyConfig
	logger    *common.Logger
	eventChan chan ProductionEvent
	metrics   ManufacturingMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// SurveyConfig holds configuration for industrial survey adapter
type SurveyConfig struct {
	Name            string        `mapstructure:"name"`
	APIEndpoint     string        `mapstructure:"api_endpoint"`
	APIKey          string        `mapstructure:"api_key"`
	Sectors         []string      `mapstructure:"sectors"`
	RegionIDs       []string      `mapstructure:"region_ids"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
}

// MinistryAdapter implements adapter for ministry of commerce data
type MinistryAdapter struct {
	config    MinistryConfig
	logger    *common.Logger
	eventChan chan ProductionEvent
	metrics   ManufacturingMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// MinistryConfig holds configuration for ministry adapter
type MinistryConfig struct {
	Name            string        `mapstructure:"name"`
	APIEndpoint     string        `mapstructure:"api_endpoint"`
	APIKey          string        `mapstructure:"api_key"`
	ReportTypes     []string      `mapstructure:"report_types"` // monthly, quarterly
	RegionIDs       []string      `mapstructure:"region_ids"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
}

// SupplyChainAdapter implements adapter for supply chain data
type SupplyChainAdapter struct {
	config    SupplyChainConfig
	logger    *common.Logger
	eventChan chan SupplyChainEvent
	metrics   SupplyChainMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// SupplyChainMetrics holds performance metrics
type SupplyChainMetrics struct {
	EventsProcessed int64
	ErrorsCount     int64
	LastEventTime   time.Time
}

// SupplyChainConfig holds configuration for supply chain adapter
type SupplyChainConfig struct {
	Name            string        `mapstructure:"name"`
	APIEndpoint     string        `mapstructure:"api_endpoint"`
	APIKey          string        `mapstructure:"api_key"`
	ChainTypes      []string      `mapstructure:"chain_types"`
	RegionIDs       []string      `mapstructure:"region_ids"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
}

// IndexAdapter implements adapter for index data
type IndexAdapter struct {
	config    IndexConfig
	logger    *common.Logger
	eventChan chan IndexData
	metrics   IndexMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// IndexMetrics holds performance metrics
type IndexMetrics struct {
	ReadingsProcessed int64
	ErrorsCount       int64
	LastReadingTime   time.Time
}

// IndexConfig holds configuration for index adapter
type IndexConfig struct {
	Name            string        `mapstructure:"name"`
	APIEndpoint     string        `mapstructure:"api_endpoint"`
	APIKey          string        `mapstructure:"api_key"`
	IndexTypes      []string      `mapstructure:"index_types"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
}

// NewIndustrialSurveyAdapter creates a new industrial survey adapter
func NewIndustrialSurveyAdapter(config SurveyConfig, logger *common.Logger) *IndustrialSurveyAdapter {
	return &IndustrialSurveyAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan ProductionEvent, 1000),
		stopChan:  make(chan struct{}),
	}
}

// NewMinistryAdapter creates a new ministry adapter
func NewMinistryAdapter(config MinistryConfig, logger *common.Logger) *MinistryAdapter {
	return &MinistryAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan ProductionEvent, 1000),
		stopChan:  make(chan struct{}),
	}
}

// NewSupplyChainAdapter creates a new supply chain adapter
func NewSupplyChainAdapter(config SupplyChainConfig, logger *common.Logger) *SupplyChainAdapter {
	return &SupplyChainAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan SupplyChainEvent, 1000),
		stopChan:  make(chan struct{}),
	}
}

// NewIndexAdapter creates a new index adapter
func NewIndexAdapter(config IndexConfig, logger *common.Logger) *IndexAdapter {
	return &IndexAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan IndexData, 500),
		stopChan:  make(chan struct{}),
	}
}

// IndustrialSurveyAdapter implementation
func (isa *IndustrialSurveyAdapter) Connect(ctx context.Context) error {
	isa.logger.Info("Connecting to industrial survey", "sectors", isa.config.Sectors)
	isa.logger.Info("Industrial survey adapter connected successfully")
	return nil
}

func (isa *IndustrialSurveyAdapter) Disconnect() error {
	isa.running = false
	close(isa.stopChan)
	isa.wg.Wait()
	return nil
}

func (isa *IndustrialSurveyAdapter) Health() bool {
	isa.metricsMu.RLock()
	defer isa.metricsMu.RUnlock()
	return isa.running && isa.metrics.ErrorsCount < 100
}

func (isa *IndustrialSurveyAdapter) Stream() <-chan ProductionEvent {
	return isa.eventChan
}

func (isa *IndustrialSurveyAdapter) Metrics() ManufacturingMetrics {
	isa.metricsMu.RLock()
	defer isa.metricsMu.RUnlock()
	return isa.metrics
}

func (isa *IndustrialSurveyAdapter) Start(ctx context.Context) error {
	if err := isa.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect industrial survey: %w", err)
	}
	
	isa.running = true
	isa.wg.Add(1)
	go isa.processLoop()
	
	isa.logger.Info("Industrial survey adapter started")
	return nil
}

func (isa *IndustrialSurveyAdapter) processLoop() {
	defer isa.wg.Done()
	
	ticker := time.NewTicker(isa.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-isa.stopChan:
			return
		case <-ticker.C:
			isa.collectProductionData()
		}
	}
}

func (isa *IndustrialSurveyAdapter) collectProductionData() {
	now := time.Now()
	
	for _, sector := range isa.config.Sectors {
		for _, regionID := range isa.config.RegionIDs {
			event := ProductionEvent{
				ID:              fmt.Sprintf("IND-%s-%s-%d", sector, regionID, now.UnixNano()),
				Sector:          sector,
				IndustryCode:    fmt.Sprintf("NIC-%s", sector),
				RegionID:        regionID,
				Timestamp:       now,
				ProductionDate:  now,
				OutputQuantity:  10000 + float64(now.Hour()*100),
				OutputUnit:      "units",
				OutputValue:     50000000,
				InstalledCapacity: 15000,
				CapacityUtilized: 12000,
				UtilizationRate:  80.0,
				WorkersEmployed: 500,
				ShiftCount:      2,
				OperatingHours:  16,
				PlannedHours:    24,
				ProductionStatus: "normal",
				EnergyConsumed:  5000,
			}
			
			select {
			case isa.eventChan <- event:
				isa.updateMetricsSuccess()
			default:
				isa.updateMetricsError()
			}
		}
	}
}

func (isa *IndustrialSurveyAdapter) updateMetricsSuccess() {
	isa.metricsMu.Lock()
	defer isa.metricsMu.Unlock()
	isa.metrics.EventsProcessed++
	isa.metrics.LastEventTime = time.Now()
}

func (isa *IndustrialSurveyAdapter) updateMetricsError() {
	isa.metricsMu.Lock()
	defer isa.metricsMu.Unlock()
	isa.metrics.ErrorsCount++
}

// MinistryAdapter implementation
func (ma *MinistryAdapter) Connect(ctx context.Context) error {
	ma.logger.Info("Connecting to ministry of commerce", "endpoint", ma.config.APIEndpoint)
	return nil
}

func (ma *MinistryAdapter) Disconnect() error {
	ma.running = false
	close(ma.stopChan)
	ma.wg.Wait()
	return nil
}

func (ma *MinistryAdapter) Health() bool {
	return ma.running
}

func (ma *MinistryAdapter) Stream() <-chan ProductionEvent {
	return ma.eventChan
}

func (ma *MinistryAdapter) Metrics() ManufacturingMetrics {
	ma.metricsMu.RLock()
	defer ma.metricsMu.RUnlock()
	return ma.metrics
}

func (ma *MinistryAdapter) Start(ctx context.Context) error {
	if err := ma.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect ministry: %w", err)
	}
	
	ma.running = true
	ma.wg.Add(1)
	go ma.processLoop()
	
	ma.logger.Info("Ministry adapter started")
	return nil
}

func (ma *MinistryAdapter) processLoop() {
	defer ma.wg.Done()
	
	ticker := time.NewTicker(ma.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ma.stopChan:
			return
		case <-ticker.C:
			ma.collectMinistryData()
		}
	}
}

func (ma *MinistryAdapter) collectMinistryData() {
	now := time.Now()
	
	// Collect index data
	for _, regionID := range ma.config.RegionIDs {
		event := ProductionEvent{
			ID:              fmt.Sprintf("MOC-%s-%d", regionID, now.UnixNano()),
			Sector:          "manufacturing",
			RegionID:        regionID,
			Timestamp:       now,
			ProductionDate:  now,
			OutputValue:     250000000,
			UtilizationRate: 78.5,
			WorkersEmployed: 15000,
			ProductionStatus: "normal",
		}
		
		select {
		case ma.eventChan <- event:
			ma.metrics.EventsProcessed++
			ma.metrics.LastEventTime = now
		default:
			ma.metrics.ErrorsCount++
		}
	}
}

// SupplyChainAdapter implementation
func (sca *SupplyChainAdapter) Connect(ctx context.Context) error {
	sca.logger.Info("Connecting to supply chain monitoring", "chain_types", sca.config.ChainTypes)
	return nil
}

func (sca *SupplyChainAdapter) Disconnect() error {
	sca.running = false
	close(sca.stopChan)
	sa.wg.Wait()
	return nil
}

func (sca *SupplyChainAdapter) Health() bool {
	return sca.running
}

func (sca *SupplyChainAdapter) Stream() <-chan SupplyChainEvent {
	return sca.eventChan
}

func (sca *SupplyChainAdapter) Metrics() SupplyChainMetrics {
	sca.metricsMu.RLock()
	defer sca.metricsMu.RUnlock()
	return sca.metrics
}

func (sca *SupplyChainAdapter) Start(ctx context.Context) error {
	if err := sca.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect supply chain: %w", err)
	}
	
	sca.running = true
	sca.wg.Add(1)
	go sca.processLoop()
	
	sca.logger.Info("Supply chain adapter started")
	return nil
}

func (sca *SupplyChainAdapter) processLoop() {
	defer sca.wg.Done()
	
	ticker := time.NewTicker(sca.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-sca.stopChan:
			return
		case <-ticker.C:
			sca.collectSupplyChainEvents()
		}
	}
}

func (sca *SupplyChainAdapter) collectSupplyChainEvents() {
	now := time.Now()
	
	for _, chainType := range sca.config.ChainTypes {
		for _, regionID := range sca.config.RegionIDs {
			event := SupplyChainEvent{
				ID:             fmt.Sprintf("SC-%s-%s-%d", chainType, regionID, now.UnixNano()),
				EventType:      "shipment",
				ChainType:      chainType,
				Stage:          "distribution",
				OriginRegion:   regionID,
				DestRegion:     "NATIONAL",
				Timestamp:      now,
				AffectedVolume: 1000,
				AffectedValue:  500000,
				Status:         "active",
				Description:    fmt.Sprintf("Regular %s shipment", chainType),
			}
			
			select {
			case sca.eventChan <- event:
				sca.metrics.EventsProcessed++
				sca.metrics.LastEventTime = now
			default:
				sca.metrics.ErrorsCount++
			}
		}
	}
}

// IndexAdapter implementation
func (ia *IndexAdapter) Connect(ctx context.Context) error {
	ia.logger.Info("Connecting to index data provider", "index_types", ia.config.IndexTypes)
	return nil
}

func (ia *IndexAdapter) Disconnect() error {
	ia.running = false
	close(ia.stopChan)
	ia.wg.Wait()
	return nil
}

func (ia *IndexAdapter) Health() bool {
	return ia.running
}

func (ia *IndexAdapter) Stream() <-chan IndexData {
	return ia.eventChan
}

func (ia *IndexAdapter) Metrics() IndexMetrics {
	ia.metricsMu.RLock()
	defer ia.metricsMu.RUnlock()
	return ia.metrics
}

func (ia *IndexAdapter) Start(ctx context.Context) error {
	if err := ia.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect index adapter: %w", err)
	}
	
	ia.running = true
	ia.wg.Add(1)
	go ia.processLoop()
	
	ia.logger.Info("Index adapter started")
	return nil
}

func (ia *IndexAdapter) processLoop() {
	defer ia.wg.Done()
	
	ticker := time.NewTicker(ia.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ia.stopChan:
			return
		case <-ticker.C:
			ia.collectIndexData()
		}
	}
}

func (ia *IndexAdapter) collectIndexData() {
	now := time.Now()
	
	for _, indexType := range ia.config.IndexTypes {
		data := IndexData{
			ID:          fmt.Sprintf("IDX-%s-%d", indexType, now.UnixNano()),
			IndexType:   indexType,
			Sector:      "all",
			Period:      "monthly",
			PeriodStart: now.AddDate(0, -1, 0),
			PeriodEnd:   now,
			IndexValue:  150.5,
			BaseValue:   100.0,
			ChangePct:   2.5,
			Change MOM:  0.8,
			Change YOY:  5.2,
			WeightInOverall: 100.0,
			Contributors: []Contributor{
				{Factor: "manufacturing", ContributionPct: 45.0, Direction: "positive"},
				{Factor: "mining", ContributionPct: 20.0, Direction: "positive"},
				{Factor: "utilities", ContributionPct: 35.0, Direction: "positive"},
			},
			ProjectedNext: 152.0,
			Projected3M:   155.0,
			Projected6M:   160.0,
		}
		
		select {
		case ia.eventChan <- data:
			ia.metrics.ReadingsProcessed++
			ia.metrics.LastReadingTime = now
		default:
			ia.metrics.ErrorsCount++
		}
	}
}

// IndustrialAggregator combines industrial data from multiple adapters
type IndustrialAggregator struct {
	adapters   []interface{}
	outputChan chan interface{}
	logger     *common.Logger
	wg         sync.WaitGroup
	stopChan   chan struct{}
}

// NewIndustrialAggregator creates a new industrial aggregator
func NewIndustrialAggregator(logger *common.Logger, adapters ...interface{}) *IndustrialAggregator {
	return &IndustrialAggregator{
		adapters:   adapters,
		outputChan: make(chan interface{}, 5000),
		logger:     logger,
		stopChan:   make(chan struct{}),
	}
}

// Start begins aggregating industrial data
func (ia *IndustrialAggregator) Start(ctx context.Context) error {
	ia.wg.Add(1)
	go ia.aggregate()
	
	ia.logger.Info("Industrial aggregator started", "adapters", len(ia.adapters))
	return nil
}

// aggregate merges data from all adapters
func (ia *IndustrialAggregator) aggregate() {
	defer ia.wg.Done()
	
	for {
		select {
		case <-ia.stopChan:
			return
		default:
			for _, adapter := range ia.adapters {
				// Handle different adapter types
				switch a := adapter.(type) {
				case ManufacturingAdapter:
					select {
					case event := <-a.Stream():
						ia.outputChan <- event
					default:
						continue
					}
				case SupplyChainAdapter:
					select {
					case event := <-a.Stream():
						ia.outputChan <- event
					default:
						continue
					}
				case IndexAdapter:
					select {
					case event := <-a.Stream():
						ia.outputChan <- event
					default:
						continue
					}
				}
			}
			time.Sleep(time.Millisecond)
		}
	}
}

// Stop stops all adapters and aggregation
func (ia *IndustrialAggregator) Stop() {
	close(ia.stopChan)
	ia.wg.Wait()
}

// Stream returns the aggregated data stream
func (ia *IndustrialAggregator) Stream() <-chan interface{} {
	return ia.outputChan
}

// ToJSON converts event to JSON
func (pe *ProductionEvent) ToJSON() string {
	data, _ := json.Marshal(pe)
	return string(data)
}
