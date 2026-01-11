// NEAM Sensing Layer - Transport & Logistics Adapter
// Real-time transport throughput and logistics data ingestion

package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"neam-platform/sensing/common"
)

// TransportEvent represents a transport/logistics event
type TransportEvent struct {
	// Identification
	ID            string    `json:"id"`
	EventType     string    `json:"event_type"` // arrival, departure, crossing, loading, unloading
	TransportType string    `json:"transport_type"` // road, rail, air, sea, pipeline
	OperationType string    `json:"operation_type"` // cargo, passenger, freight
	
	// Location
	PortID        string  `json:"port_id,omitempty"`
	PortName      string  `json:"port_name,omitempty"`
	RegionID      string  `json:"region_id"`
	DistrictID    string  `json:"district_id,omitempty"`
	RouteID       string  `json:"route_id,omitempty"`
	
	// Timestamp
	Timestamp time.Time `json:"timestamp"`
	
	// Cargo information
	ContainerCount   int     `json:"container_count,omitempty"`
	ContainerTEU     float64 `json:"container_teu,omitempty"`
	CargoWeightKG    float64 `json:"cargo_weight_kg,omitempty"`
	CargoType        string  `json:"cargo_type,omitempty"` // containerized, bulk, liquid, dry
	CargoValue       float64 `json:"cargo_value,omitempty"`
	
	// Vehicle information
	VehicleID       string  `json:"vehicle_id,omitempty"`
	VehicleType     string  `json:"vehicle_type,omitempty"` // truck, train, ship, aircraft
	RegistrationNo  string  `json:"registration_no,omitempty"`
	
	// Passenger information
	PassengerCount int `json:"passenger_count,omitempty"`
	
	// Throughput metrics
	ProcessingTimeMin float64 `json:"processing_time_min,omitempty"`
	WaitTimeMin       float64 `json:"wait_time_min,omitempty"`
	
	// Status
	Status     string `json:"status"` // completed, in_progress, delayed, cancelled
	DelayHours float64 `json:"delay_hours,omitempty"`
	
	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// LogisticsMetrics represents aggregated logistics metrics
type LogisticsMetrics struct {
	PortID          string    `json:"port_id"`
	PortName        string    `json:"port_name"`
	RegionID        string    `json:"region_id"`
	Period          string    `json:"period"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	
	// Container throughput
	TotalContainers     int64   `json:"total_containers"`
	TotalTEU            float64 `json:"total_teu"`
	ImportTEU           float64 `json:"import_teu"`
	ExportTEU           float64 `json:"export_teu"`
	TransshipmentTEU    float64 `json:"transshipment_teu"`
	
	// Vessel statistics
	VesselArrivals      int     `json:"vessel_arrivals"`
	VesselDepartures    int     `json:"vessel_departures"`
	AverageDwellDays    float64 `json:"avg_dwell_days"`
	
	// Cargo throughput
	TotalCargoWeight    float64 `json:"total_cargo_weight_kg"`
	CargoByType         map[string]float64 `json:"cargo_by_type"`
	CargoByDestination  map[string]float64 `json:"cargo_by_destination"`
	
	// Performance
	AverageWaitTime     float64 `json:"avg_wait_time_min"`
	AverageProcessingTime float64 `json:"avg_processing_time_min"`
	UtilizationRate     float64 `json:"utilization_rate_pct"`
	
	// Delays
	DelayedShipments    int     `json:"delayed_shipments"`
	DelayTotalHours     float64 `json:"total_delay_hours"`
}

// FuelData represents fuel consumption and pricing data
type FuelData struct {
	// Identification
	ID          string    `json:"id"`
	RegionID    string    `json:"region_id"`
	StationID   string    `json:"station_id"`
	StationType string    `json:"station_type"` // petrol_pump, gas_station, depot
	
	// Timestamp
	Timestamp   time.Time `json:"timestamp"`
	
	// Fuel types
	PetrolSoldLiters   float64 `json:"petrol_sold_liters"`
	DieselSoldLiters   float64 `json:"diesel_sold_liters"`
	CNGSoldKG          float64 `json:"cng_sold_kg,omitempty"`
	LPGSoldKG          float64 `json:"lpg_sold_kg,omitempty"`
	
	// Pricing
	PetrolPricePerLiter float64 `json:"petrol_price"`
	DieselPricePerLiter float64 `json:"diesel_price"`
	CNGPricePerKG       float64 `json:"cng_price,omitempty"`
	LPGPricePerKG       float64 `json:"lpg_price,omitempty"`
	
	// Inventory
	PetrolInventoryLiters float64 `json:"petrol_inventory,omitempty"`
	DieselInventoryLiters float64 `json:"diesel_inventory,omitempty"`
	
	// Vehicle count
	VehicleTransactions int `json:"vehicle_transactions"`
	
	// Status
	SupplyStatus string `json:"supply_status"` // normal, constrained, critical
}

// PortAdapter interface for port and maritime data
type PortAdapter interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Health() bool
	Stream() <-chan TransportEvent
	Metrics() PortMetrics
}

// PortMetrics holds performance metrics for port adapter
type PortMetrics struct {
	EventsProcessed int64
	ErrorsCount     int64
	LastEventTime   time.Time
	Latency         time.Duration
}

// PortAuthorityAdapter implements adapter for port authority systems
type PortAuthorityAdapter struct {
	config    PortConfig
	logger    *common.Logger
	eventChan chan TransportEvent
	metrics   PortMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// PortConfig holds configuration for port adapter
type PortConfig struct {
	Name            string        `mapstructure:"name"`
	APIEndpoint     string        `mapstructure:"api_endpoint"`
	APIKey          string        `mapstructure:"api_key"`
	PortIDs         []string      `mapstructure:"port_ids"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
}

// CustomsAdapter implements adapter for customs data
type CustomsAdapter struct {
	config    CustomsConfig
	logger    *common.Logger
	eventChan chan TransportEvent
	metrics   PortMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// CustomsConfig holds configuration for customs adapter
type CustomsConfig struct {
	Name            string        `mapstructure:"name"`
	APIEndpoint     string        `mapstructure:"api_endpoint"`
	APIKey          string        `mapstructure:"api_key"`
	PortIDs         []string      `mapstructure:"port_ids"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
}

// FuelAdapter implements adapter for fuel data
type FuelAdapter struct {
	config    FuelConfig
	logger    *common.Logger
	eventChan chan FuelData
	metrics   FuelMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// FuelMetrics holds performance metrics for fuel adapter
type FuelMetrics struct {
	ReadingsProcessed int64
	ErrorsCount       int64
	LastReadingTime   time.Time
}

// FuelConfig holds configuration for fuel adapter
type FuelConfig struct {
	Name            string        `mapstructure:"name"`
	APIEndpoint     string        `mapstructure:"api_endpoint"`
	APIKey          string        `mapstructure:"api_key"`
	RegionIDs       []string      `mapstructure:"region_ids"`
	StationIDs      []string      `mapstructure:"station_ids"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
}

// LogisticsAdapter implements adapter for logistics/freight data
type LogisticsAdapter struct {
	config    LogisticsConfig
	logger    *common.Logger
	eventChan chan TransportEvent
	metrics   PortMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// LogisticsConfig holds configuration for logistics adapter
type LogisticsConfig struct {
	Name            string        `mapstructure:"name"`
	APIEndpoint     string        `mapstructure:"api_endpoint"`
	APIKey          string        `mapstructure:"api_key"`
	TransportTypes  []string      `mapstructure:"transport_types"` // road, rail, air
	RegionIDs       []string      `mapstructure:"region_ids"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
}

// NewPortAuthorityAdapter creates a new port authority adapter
func NewPortAuthorityAdapter(config PortConfig, logger *common.Logger) *PortAuthorityAdapter {
	return &PortAuthorityAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan TransportEvent, 1000),
		stopChan:  make(chan struct{}),
	}
}

// NewCustomsAdapter creates a new customs adapter
func NewCustomsAdapter(config CustomsConfig, logger *common.Logger) *CustomsAdapter {
	return &CustomsAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan TransportEvent, 1000),
		stopChan:  make(chan struct{}),
	}
}

// NewFuelAdapter creates a new fuel adapter
func NewFuelAdapter(config FuelConfig, logger *common.Logger) *FuelAdapter {
	return &FuelAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan FuelData, 1000),
		stopChan:  make(chan struct{}),
	}
}

// NewLogisticsAdapter creates a new logistics adapter
func NewLogisticsAdapter(config LogisticsConfig, logger *common.Logger) *LogisticsAdapter {
	return &LogisticsAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan TransportEvent, 1000),
		stopChan:  make(chan struct{}),
	}
}

// PortAuthorityAdapter implementation
func (pa *PortAuthorityAdapter) Connect(ctx context.Context) error {
	pa.logger.Info("Connecting to port authority", "endpoint", pa.config.APIEndpoint, "ports", len(pa.config.PortIDs))
	pa.logger.Info("Port authority adapter connected successfully")
	return nil
}

func (pa *PortAuthorityAdapter) Disconnect() error {
	pa.running = false
	close(pa.stopChan)
	pa.wg.Wait()
	return nil
}

func (pa *PortAuthorityAdapter) Health() bool {
	pa.metricsMu.RLock()
	defer pa.metricsMu.RUnlock()
	return pa.running && pa.metrics.ErrorsCount < 100
}

func (pa *PortAuthorityAdapter) Stream() <-chan TransportEvent {
	return pa.eventChan
}

func (pa *PortAuthorityAdapter) Metrics() PortMetrics {
	pa.metricsMu.RLock()
	defer pa.metricsMu.RUnlock()
	return pa.metrics
}

func (pa *PortAuthorityAdapter) Start(ctx context.Context) error {
	if err := pa.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect port authority: %w", err)
	}
	
	pa.running = true
	pa.wg.Add(1)
	go pa.processLoop()
	
	pa.logger.Info("Port authority adapter started")
	return nil
}

func (pa *PortAuthorityAdapter) processLoop() {
	defer pa.wg.Done()
	
	ticker := time.NewTicker(pa.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-pa.stopChan:
			return
		case <-ticker.C:
			pa.collectPortEvents()
		}
	}
}

func (pa *PortAuthorityAdapter) collectPortEvents() {
	now := time.Now()
	
	for _, portID := range pa.config.PortIDs {
		// Simulate vessel arrival
		event := TransportEvent{
			ID:            fmt.Sprintf("PORT-%s-%d", portID, now.UnixNano()),
			EventType:     "arrival",
			TransportType: "sea",
			OperationType: "cargo",
			PortID:        portID,
			RegionID:      "COASTAL",
			Timestamp:     now,
			ContainerCount: 150 + (now.Hour() % 10) * 10,
			ContainerTEU:  225,
			CargoWeightKG: 2500000,
			CargoType:     "containerized",
			Status:        "completed",
			ProcessingTimeMin: 45.0,
			WaitTimeMin:   30.0,
		}
		
		select {
		case pa.eventChan <- event:
			pa.updateMetricsSuccess()
		default:
			pa.updateMetricsError()
		}
		
		// Simulate container handling
		for i := 0; i < 5; i++ {
			handlingEvent := TransportEvent{
				ID:            fmt.Sprintf("PORT-HANDLE-%s-%d", portID, now.UnixNano()+int64(i)),
				EventType:     "loading",
				TransportType: "sea",
				PortID:        portID,
				RegionID:      "COASTAL",
				Timestamp:     now.Add(time.Duration(i) * time.Minute),
				ContainerTEU:  20,
				CargoWeightKG: 400000,
				Status:        "completed",
			}
			
			select {
			case pa.eventChan <- handlingEvent:
				pa.updateMetricsSuccess()
			default:
				pa.updateMetricsError()
			}
		}
	}
}

func (pa *PortAuthorityAdapter) updateMetricsSuccess() {
	pa.metricsMu.Lock()
	defer pa.metricsMu.Unlock()
	pa.metrics.EventsProcessed++
	pa.metrics.LastEventTime = time.Now()
}

func (pa *PortAuthorityAdapter) updateMetricsError() {
	pa.metricsMu.Lock()
	defer pa.metricsMu.Unlock()
	pa.metrics.ErrorsCount++
}

// CustomsAdapter implementation
func (ca *CustomsAdapter) Connect(ctx context.Context) error {
	ca.logger.Info("Connecting to customs", "endpoint", ca.config.APIEndpoint)
	return nil
}

func (ca *CustomsAdapter) Disconnect() error {
	ca.running = false
	close(ca.stopChan)
	ca.wg.Wait()
	return nil
}

func (ca *CustomsAdapter) Health() bool {
	return ca.running
}

func (ca *CustomsAdapter) Stream() <-chan TransportEvent {
	return ca.eventChan
}

func (ca *CustomsAdapter) Metrics() PortMetrics {
	ca.metricsMu.RLock()
	defer ca.metricsMu.RUnlock()
	return ca.metrics
}

func (ca *CustomsAdapter) Start(ctx context.Context) error {
	if err := ca.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect customs: %w", err)
	}
	
	ca.running = true
	ca.wg.Add(1)
	go ca.processLoop()
	
	ca.logger.Info("Customs adapter started")
	return nil
}

func (ca *CustomsAdapter) processLoop() {
	defer ca.wg.Done()
	
	ticker := time.NewTicker(ca.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ca.stopChan:
			return
		case <-ticker.C:
			ca.collectCustomsEvents()
		}
	}
}

func (ca *CustomsAdapter) collectCustomsEvents() {
	now := time.Now()
	
	for _, portID := range ca.config.PortIDs {
		event := TransportEvent{
			ID:            fmt.Sprintf("CUSTOMS-%s-%d", portID, now.UnixNano()),
			EventType:     "clearance",
			TransportType: "sea",
			PortID:        portID,
			RegionID:      "COASTAL",
			Timestamp:     now,
			CargoValue:    500000 + float64(now.Hour()*10000),
			CargoType:     "import",
			Status:        "completed",
			ProcessingTimeMin: 120.0,
		}
		
		select {
		case ca.eventChan <- event:
			ca.metrics.EventsProcessed++
			ca.metrics.LastEventTime = now
		default:
			ca.metrics.ErrorsCount++
		}
	}
}

// FuelAdapter implementation
func (fa *FuelAdapter) Connect(ctx context.Context) error {
	fa.logger.Info("Connecting to fuel monitoring", "endpoint", fa.config.APIEndpoint)
	return nil
}

func (fa *FuelAdapter) Disconnect() error {
	fa.running = false
	close(fa.stopChan)
	fa.wg.Wait()
	return nil
}

func (fa *FuelAdapter) Health() bool {
	return fa.running
}

func (fa *FuelAdapter) Stream() <-chan FuelData {
	return fa.eventChan
}

func (fa *FuelAdapter) Metrics() FuelMetrics {
	fa.metricsMu.RLock()
	defer fa.metricsMu.RUnlock()
	return fa.metrics
}

func (fa *FuelAdapter) Start(ctx context.Context) error {
	if err := fa.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect fuel adapter: %w", err)
	}
	
	fa.running = true
	fa.wg.Add(1)
	go fa.processLoop()
	
	fa.logger.Info("Fuel adapter started")
	return nil
}

func (fa *FuelAdapter) processLoop() {
	defer fa.wg.Done()
	
	ticker := time.NewTicker(fa.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-fa.stopChan:
			return
		case <-ticker.C:
			fa.collectFuelData()
		}
	}
}

func (fa *FuelAdapter) collectFuelData() {
	now := time.Now()
	
	for _, regionID := range fa.config.RegionIDs {
		for _, stationID := range fa.config.StationIDs {
			data := FuelData{
				ID:               fmt.Sprintf("FUEL-%s-%s-%d", regionID, stationID, now.UnixNano()),
				RegionID:         regionID,
				StationID:        stationID,
				StationType:      "petrol_pump",
				Timestamp:        now,
				PetrolSoldLiters: 5000 + float64(now.Hour()*500),
				DieselSoldLiters: 4000 + float64(now.Hour()*400),
				PetrolPricePerLiter: 105.5 + float64(now.Hour()%10)*0.1,
				DieselPricePerLiter: 89.5 + float64(now.Hour()%10)*0.1,
				VehicleTransactions: 150 + now.Hour()*10,
				SupplyStatus:      "normal",
			}
			
			select {
			case fa.eventChan <- data:
				fa.metrics.ReadingsProcessed++
				fa.metrics.LastReadingTime = now
			default:
				fa.metrics.ErrorsCount++
			}
		}
	}
}

// LogisticsAdapter implementation
func (la *LogisticsAdapter) Connect(ctx context.Context) error {
	la.logger.Info("Connecting to logistics", "endpoint", la.config.APIEndpoint)
	return nil
}

func (la *LogisticsAdapter) Disconnect() error {
	la.running = false
	close(la.stopChan)
	la.wg.Wait()
	return nil
}

func (la *LogisticsAdapter) Health() bool {
	return la.running
}

func (la *LogisticsAdapter) Stream() <-chan TransportEvent {
	return la.eventChan
}

func (la *LogisticsAdapter) Metrics() PortMetrics {
	la.metricsMu.RLock()
	defer la.metricsMu.RUnlock()
	return la.metrics
}

func (la *LogisticsAdapter) Start(ctx context.Context) error {
	if err := la.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect logistics: %w", err)
	}
	
	la.running = true
	la.wg.Add(1)
	go la.processLoop()
	
	la.logger.Info("Logistics adapter started")
	return nil
}

func (la *LogisticsAdapter) processLoop() {
	defer la.wg.Done()
	
	ticker := time.NewTicker(la.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-la.stopChan:
			return
		case <-ticker.C:
			la.collectLogisticsEvents()
		}
	}
}

func (la *LogisticsAdapter) collectLogisticsEvents() {
	now := time.Now()
	
	for _, transportType := range la.config.TransportTypes {
		for _, regionID := range la.config.RegionIDs {
			event := TransportEvent{
				ID:            fmt.Sprintf("LOG-%s-%s-%d", transportType, regionID, now.UnixNano()),
				EventType:     "movement",
				TransportType: transportType,
				OperationType: "freight",
				RegionID:      regionID,
				Timestamp:     now,
				CargoWeightKG: 10000 + float64(now.Hour()*1000),
				VehicleType:   transportType,
				Status:        "completed",
			}
			
			select {
			case la.eventChan <- event:
				la.metrics.EventsProcessed++
				la.metrics.LastEventTime = now
			default:
				la.metrics.ErrorsCount++
			}
		}
	}
}

// TransportAggregator combines transport events from multiple adapters
type TransportAggregator struct {
	adapters   []interface{ Stream() <-chan TransportEvent }
	outputChan chan TransportEvent
	logger     *common.Logger
	wg         sync.WaitGroup
	stopChan   chan struct{}
}

// NewTransportAggregator creates a new transport aggregator
func NewTransportAggregator(logger *common.Logger, adapters ...interface{ Stream() <-chan TransportEvent }) *TransportAggregator {
	return &TransportAggregator{
		adapters:   adapters,
		outputChan: make(chan TransportEvent, 5000),
		logger:     logger,
		stopChan:   make(chan struct{}),
	}
}

// Start begins aggregating transport events
func (ta *TransportAggregator) Start(ctx context.Context) error {
	ta.wg.Add(1)
	go ta.aggregate()
	
	ta.logger.Info("Transport aggregator started", "adapters", len(ta.adapters))
	return nil
}

// aggregate merges events from all adapters
func (ta *TransportAggregator) aggregate() {
	defer ta.wg.Done()
	
	for {
		select {
		case <-ta.stopChan:
			return
		default:
			for _, adapter := range ta.adapters {
				select {
				case event := <-adapter.Stream():
					ta.outputChan <- event
				default:
					continue
				}
			}
			time.Sleep(time.Millisecond)
		}
	}
}

// Stop stops all adapters and aggregation
func (ta *TransportAggregator) Stop() {
	close(ta.stopChan)
	ta.wg.Wait()
}

// Stream returns the aggregated event stream
func (ta *TransportAggregator) Stream() <-chan TransportEvent {
	return ta.outputChan
}

// ToJSON converts event to JSON
func (te *TransportEvent) ToJSON() string {
	data, _ := json.Marshal(te)
	return string(data)
}
