// NEAM Sensing Layer - Energy Telemetry Adapter
// Real-time energy grid data ingestion and processing

package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"neam-platform/sensing/common"
)

// EnergyReading represents a single energy telemetry reading
type EnergyReading struct {
	// Identification
	ID          string    `json:"id"`
	RegionID    string    `json:"region_id"`
	ZoneID      string    `json:"zone_id,omitempty"`
	StationID   string    `json:"station_id"`
	StationName string    `json:"station_name,omitempty"`
	
	// Timestamp
	Timestamp time.Time `json:"timestamp"`
	
	// Load metrics
	LoadMW           float64 `json:"load_mw"`
	LoadMWPeak       float64 `json:"load_mw_peak,omitempty"`
	LoadMWOffPeak     float64 `json:"load_mw_off_peak,omitempty"`
	LoadFactor       float64 `json:"load_factor,omitempty"`
	
	// Generation metrics
	GenerationMW      float64 `json:"generation_mw,omitempty"`
	GenerationSource  string  `json:"generation_source,omitempty"` // thermal, hydro, solar, wind, nuclear
	RenewableShare    float64 `json:"renewable_share,omitempty"`
	
	// Capacity metrics
	InstalledCapacityMW float64 `json:"installed_capacity_mw,omitempty"`
	AvailableCapacityMW float64 `json:"available_capacity_mw,omitempty"`
	ReserveMargin       float64 `json:"reserve_margin,omitempty"`
	
	// Frequency and voltage
	Frequency float64 `json:"frequency_hz,omitempty"`
	Voltage   float64 `json:"voltage_kv,omitempty"`
	
	// Status
	GridStatus   string `json:"grid_status,omitempty"` // normal, warning, critical
	AlertLevel   string `json:"alert_level,omitempty"` // none, yellow, orange, red
	
	// Weather impact (if available)
	WeatherCondition string  `json:"weather_condition,omitempty"`
	Temperature      float64 `json:"temperature_celsius,omitempty"`
	
	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// EnergyZoneMetrics represents aggregated metrics for a zone
type EnergyZoneMetrics struct {
	ZoneID          string    `json:"zone_id"`
	RegionID        string    `json:"region_id"`
	Period          string    `json:"period"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	
	// Load statistics
	AverageLoadMW   float64 `json:"avg_load_mw"`
	PeakLoadMW      float64 `json:"peak_load_mw"`
	MinLoadMW       float64 `json:"min_load_mw"`
	TotalEnergyGWH  float64 `json:"total_energy_gwh"`
	
	// Generation breakdown
	ThermalGWH      float64 `json:"thermal_gwh"`
	HydroGWH        float64 `json:"hydro_gwh"`
	SolarGWH        float64 `json:"solar_gwh"`
	WindGWH         float64 `json:"wind_gwh"`
	NuclearGWH      float64 `json:"nuclear_gwh"`
	RenewableShare  float64 `json:"renewable_share_pct"`
	
	// Capacity utilization
	AvgUtilization  float64 `json:"avg_utilization_pct"`
	PeakUtilization float64 `json:"peak_utilization_pct"`
	
	// Grid stability
	AvgFrequency    float64 `json:"avg_frequency_hz"`
	MinFrequency    float64 `json:"min_frequency_hz"`
	MaxFrequency    float64 `json:"max_frequency_hz"`
	
	// Alerts
	TotalAlerts     int     `json:"total_alerts"`
	CriticalAlerts  int     `json:"critical_alerts"`
}

// EnergyRegionMetrics represents aggregated metrics for a region
type EnergyRegionMetrics struct {
	RegionID        string    `json:"region_id"`
	RegionName      string    `json:"region_name"`
	Period          string    `json:"period"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	
	// Overall load
	TotalDemandMW   float64 `json:"total_demand_mw"`
	PeakDemandMW    float64 `json:"peak_demand_mw"`
	EnergyConsumedGWH float64 `json:"energy_consumed_gwh"`
	
	// Per source
	GenerationBySource map[string]float64 `json:"generation_by_source"`
	
	// Inter-regional
	ImportsMW        float64 `json:"imports_mw"`
	ExportsMW        float64 `json:"exports_mw"`
	NetInterchangeMW float64 `json:"net_interchange_mw"`
	
	// Comparison with previous
	DemandChangePct  float64 `json:"demand_change_pct"`
	GenerationChangePct float64 `json:"generation_change_pct"`
	
	// Health indicators
	GridReliability float64 `json:"grid_reliability_pct"`
	OutageDurationMin float64 `json:"outage_duration_min"`
}

// EnergyAlert represents an alert from the energy sector
type EnergyAlert struct {
	ID          string    `json:"id"`
	AlertType   string    `json:"alert_type"` // load_shedding, grid_instability, capacity_shortage, maintenance
	Severity    string    `json:"severity"` // warning, critical
	RegionID    string    `json:"region_id"`
	ZoneID      string    `json:"zone_id,omitempty"`
	
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
	
	Timestamp   time.Time  `json:"timestamp"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	
	Impact      string  `json:"impact,omitempty"` // demand_reduction_mw, affected_customers
	Status      string  `json:"status"` // active, resolved, scheduled
	
	Source      string  `json:"source"` // scada, nldc, sldc
}

// GridAdapter interface for energy data sources
type GridAdapter interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Health() bool
	Stream() <-chan EnergyReading
	Metrics() GridMetrics
}

// GridMetrics holds performance metrics for grid adapter
type GridMetrics struct {
	ReadingsProcessed int64
	ErrorsCount       int64
	LastReadingTime   time.Time
	Latency           time.Duration
}

// SCADAAdapter implements adapter for SCADA (Supervisory Control and Data Acquisition) systems
type SCADAAdapter struct {
	config    SCADAConfig
	logger    *common.Logger
	eventChan chan EnergyReading
	metrics   GridMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// SCADAConfig holds configuration for SCADA adapter
type SCADAConfig struct {
	Name            string        `mapstructure:"name"`
	SCADAEndpoint   string        `mapstructure:"scada_endpoint"`
	APIKey          string        `mapstructure:"api_key"`
	RegionIDs       []string      `mapstructure:"region_ids"`
	StationIDs      []string      `mapstructure:"station_ids"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
	RetryAttempts   int           `mapstructure:"retry_attempts"`
}

// NLDCAdapter implements adapter for National Load Dispatch Centre
type NLDCAdapter struct {
	config    NLDConfig
	logger    *common.Logger
	eventChan chan EnergyReading
	metrics   GridMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// NLDConfig holds configuration for NLDC adapter
type NLDConfig struct {
	Name             string        `mapstructure:"name"`
	NLDCEndpoint     string        `mapstructure:"nldc_endpoint"`
	APIKey           string        `mapstructure:"api_key"`
	RegionID         string        `mapstructure:"region_id"`
	PollingInterval  time.Duration `mapstructure:"polling_interval"`
	RequestTimeout   time.Duration `mapstructure:"request_timeout"`
}

// SLDCAdapter implements adapter for State Load Dispatch Centre
type SLDCAdapter struct {
	config    SLDConfig
	logger    *common.Logger
	eventChan chan EnergyReading
	metrics   GridMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// SLDConfig holds configuration for SLDC adapter
type SLDConfig struct {
	Name            string        `mapstructure:"name"`
	SLDCEndpoint    string        `mapstructure:"sldc_endpoint"`
	APIKey          string        `mapstructure:"api_key"`
	StateID         string        `mapstructure:"state_id"`
	RegionID        string        `mapstructure:"region_id"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
}

// RenewableAdapter implements adapter for renewable energy monitoring
type RenewableAdapter struct {
	config    RenewableConfig
	logger    *common.Logger
	eventChan chan EnergyReading
	metrics   GridMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// RenewableConfig holds configuration for renewable adapter
type RenewableConfig struct {
	Name            string        `mapstructure:"name"`
	APIEndpoint     string        `mapstructure:"api_endpoint"`
	APIKey          string        `mapstructure:"api_key"`
	Types           []string      `mapstructure:"types"` // solar, wind, hydro
	RegionIDs       []string      `mapstructure:"region_ids"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
}

// NewSCADAAdapter creates a new SCADA adapter
func NewSCADAAdapter(config SCADAConfig, logger *common.Logger) *SCADAAdapter {
	return &SCADAAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan EnergyReading, 1000),
		stopChan:  make(chan struct{}),
	}
}

// NewNLDCAdapter creates a new NLDC adapter
func NewNLDCAdapter(config NLDConfig, logger *common.Logger) *NLDCAdapter {
	return &NLDCAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan EnergyReading, 1000),
		stopChan:  make(chan struct{}),
	}
}

// NewSLDCAdapter creates a new SLDC adapter
func NewSLDCAdapter(config SLDConfig, logger *common.Logger) *SLDCAdapter {
	return &SLDCAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan EnergyReading, 1000),
		stopChan:  make(chan struct{}),
	}
}

// NewRenewableAdapter creates a new renewable energy adapter
func NewRenewableAdapter(config RenewableConfig, logger *common.Logger) *RenewableAdapter {
	return &RenewableAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan EnergyReading, 1000),
		stopChan:  make(chan struct{}),
	}
}

// Connect implements SCADA connection
func (sa *SCADAAdapter) Connect(ctx context.Context) error {
	sa.logger.Info("Connecting to SCADA", "endpoint", sa.config.SCADAEndpoint, "regions", len(sa.config.RegionIDs))
	
	// In production, this would establish real SCADA connection
	sa.logger.Info("SCADA adapter connected successfully")
	return nil
}

// Disconnect closes SCADA connection
func (sa *SCADAAdapter) Disconnect() error {
	sa.running = false
	close(sa.stopChan)
	sa.wg.Wait()
	return nil
}

// Health checks adapter health
func (sa *SCADAAdapter) Health() bool {
	sa.metricsMu.RLock()
	defer sa.metricsMu.RUnlock()
	return sa.running && sa.metrics.ErrorsCount < 100
}

// Stream returns reading channel
func (sa *SCADAAdapter) Stream() <-chan EnergyReading {
	return sa.eventChan
}

// Metrics returns adapter metrics
func (sa *SCADAAdapter) Metrics() GridMetrics {
	sa.metricsMu.RLock()
	defer sa.metricsMu.RUnlock()
	return sa.metrics
}

// Start begins SCADA data collection
func (sa *SCADAAdapter) Start(ctx context.Context) error {
	if err := sa.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect SCADA: %w", err)
	}
	
	sa.running = true
	sa.wg.Add(1)
	go sa.processLoop()
	
	sa.logger.Info("SCADA adapter started")
	return nil
}

// processLoop continuously collects SCADA data
func (sa *SCADAAdapter) processLoop() {
	defer sa.wg.Done()
	
	ticker := time.NewTicker(sa.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-sa.stopChan:
			return
		case <-ticker.C:
			if err := sa.collectReadings(); err != nil {
				sa.logger.Error("Failed to collect SCADA readings", "error", err)
				sa.updateMetricsError()
			}
		}
	}
}

// collectReadings gathers readings from all configured stations
func (sa *SCADAAdapter) collectReadings() error {
	now := time.Now()
	
	// Generate mock readings for each station
	for _, regionID := range sa.config.RegionIDs {
		for _, stationID := range sa.config.StationIDs {
			reading := EnergyReading{
				ID:          fmt.Sprintf("SCADA-%s-%s-%d", regionID, stationID, now.UnixNano()),
				RegionID:    regionID,
				StationID:   stationID,
				Timestamp:   now,
				LoadMW:      4500 + float64(now.Hour())*100,
				GridStatus:  "normal",
				AlertLevel:  "none",
				Frequency:   50.0 + (float64(now.Hour()-12) * 0.01),
				Voltage:     400.0,
				RenewableShare: 25.0,
			}
			
			select {
			case sa.eventChan <- reading:
				sa.updateMetricsSuccess()
			default:
				sa.updateMetricsError()
			}
		}
	}
	
	return nil
}

// updateMetricsSuccess updates metrics on success
func (sa *SCADAAdapter) updateMetricsSuccess() {
	sa.metricsMu.Lock()
	defer sa.metricsMu.Unlock()
	sa.metrics.ReadingsProcessed++
	sa.metrics.LastReadingTime = time.Now()
}

// updateMetricsError updates metrics on error
func (sa *SCADAAdapter) updateMetricsError() {
	sa.metricsMu.Lock()
	defer sa.metricsMu.Unlock()
	sa.metrics.ErrorsCount++
}

// NLDC adapter implementation
func (na *NLDCAdapter) Connect(ctx context.Context) error {
	na.logger.Info("Connecting to NLDC", "endpoint", na.config.NLDCEndpoint)
	return nil
}

func (na *NLDCAdapter) Disconnect() error {
	na.running = false
	close(na.stopChan)
	na.wg.Wait()
	return nil
}

func (na *NLDCAdapter) Health() bool {
	return na.running
}

func (na *NLDCAdapter) Stream() <-chan EnergyReading {
	return na.eventChan
}

func (na *NLDCAdapter) Metrics() GridMetrics {
	na.metricsMu.RLock()
	defer na.metricsMu.RUnlock()
	return na.metrics
}

func (na *NLDCAdapter) Start(ctx context.Context) error {
	if err := na.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect NLDC: %w", err)
	}
	
	na.running = true
	na.wg.Add(1)
	go na.processLoop()
	
	na.logger.Info("NLDC adapter started")
	return nil
}

func (na *NLDCAdapter) processLoop() {
	defer na.wg.Done()
	
	ticker := time.NewTicker(na.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-na.stopChan:
			return
		case <-ticker.C:
			na.collectNationalReadings()
		}
	}
}

func (na *NLDCAdapter) collectNationalReadings() {
	now := time.Now()
	
	// Generate national level readings
	reading := EnergyReading{
		ID:              fmt.Sprintf("NLDC-%s-%d", na.config.RegionID, now.UnixNano()),
		RegionID:        na.config.RegionID,
		Timestamp:       now,
		LoadMW:          180000 + float64(now.Hour())*5000,
		GenerationMW:    182000,
		GenerationSource: "mixed",
		RenewableShare:  24.0,
		Frequency:       50.01,
		GridStatus:      "normal",
		AlertLevel:      "none",
	}
	
	select {
	case na.eventChan <- reading:
		na.metrics.ReadingsProcessed++
		na.metrics.LastReadingTime = now
	default:
		na.metrics.ErrorsCount++
	}
}

// SLDC adapter implementation
func (sa *SLDCAdapter) Connect(ctx context.Context) error {
	sa.logger.Info("Connecting to SLDC", "state", sa.config.StateID)
	return nil
}

func (sa *SLDCAdapter) Disconnect() error {
	sa.running = false
	close(sa.stopChan)
	sa.wg.Wait()
	return nil
}

func (sa *SLDCAdapter) Health() bool {
	return sa.running
}

func (sa *SLDCAdapter) Stream() <-chan EnergyReading {
	return sa.eventChan
}

func (sa *SLDCAdapter) Metrics() GridMetrics {
	sa.metricsMu.RLock()
	defer sa.metricsMu.RUnlock()
	return sa.metrics
}

func (sa *SLDCAdapter) Start(ctx context.Context) error {
	if err := sa.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect SLDC: %w", err)
	}
	
	sa.running = true
	sa.wg.Add(1)
	go sa.processLoop()
	
	sa.logger.Info("SLDC adapter started", "state", sa.config.StateID)
	return nil
}

func (sa *SLDCAdapter) processLoop() {
	defer sa.wg.Done()
	
	ticker := time.NewTicker(sa.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-sa.stopChan:
			return
		case <-ticker.C:
			sa.collectStateReadings()
		}
	}
}

func (sa *SLDCAdapter) collectStateReadings() {
	now := time.Now()
	
	reading := EnergyReading{
		ID:          fmt.Sprintf("SLDC-%s-%d", sa.config.StateID, now.UnixNano()),
		RegionID:    sa.config.RegionID,
		Timestamp:   now,
		LoadMW:      15000 + float64(now.Hour())*500,
		GenerationMW: 15200,
		GenerationSource: "thermal",
		Frequency:   49.98,
		GridStatus:  "normal",
		AlertLevel:  "none",
	}
	
	select {
	case sa.eventChan <- reading:
		sa.metrics.ReadingsProcessed++
		sa.metrics.LastReadingTime = now
	default:
		sa.metrics.ErrorsCount++
	}
}

// Renewable adapter implementation
func (ra *RenewableAdapter) Connect(ctx context.Context) error {
	ra.logger.Info("Connecting to renewable energy monitoring", "types", ra.config.Types)
	return nil
}

func (ra *RenewableAdapter) Disconnect() error {
	ra.running = false
	close(ra.stopChan)
	ra.wg.Wait()
	return nil
}

func (ra *RenewableAdapter) Health() bool {
	return ra.running
}

func (ra *RenewableAdapter) Stream() <-chan EnergyReading {
	return ra.eventChan
}

func (ra *RenewableAdapter) Metrics() GridMetrics {
	ra.metricsMu.RLock()
	defer ra.metricsMu.RUnlock()
	return ra.metrics
}

func (ra *RenewableAdapter) Start(ctx context.Context) error {
	if err := ra.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect renewable adapter: %w", err)
	}
	
	ra.running = true
	ra.wg.Add(1)
	go ra.processLoop()
	
	ra.logger.Info("Renewable adapter started")
	return nil
}

func (ra *RenewableAdapter) processLoop() {
	defer ra.wg.Done()
	
	ticker := time.NewTicker(ra.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ra.stopChan:
			return
		case <-ticker.C:
			ra.collectRenewableReadings()
		}
	}
}

func (ra *RenewableAdapter) collectRenewableReadings() {
	now := time.Now()
	
	for _, regionID := range ra.config.RegionIDs {
		for _, genType := range ra.config.Types {
			var generationMW float64
			var stationName string
			
			switch genType {
			case "solar":
				// Solar generation depends on time of day
				hour := now.Hour()
				if hour >= 6 && hour <= 18 {
					generationMW = float64(hour-6) * 500
				} else {
					generationMW = 0
				}
				stationName = fmt.Sprintf("Solar Plant %s", regionID)
			case "wind":
				generationMW = 2000 + float64(now.Hour()*50)
				stationName = fmt.Sprintf("Wind Farm %s", regionID)
			case "hydro":
				generationMW = 3000
				stationName = fmt.Sprintf("Hydro Station %s", regionID)
			default:
				generationMW = 1000
				stationName = fmt.Sprintf("Renewable %s %s", genType, regionID)
			}
			
			reading := EnergyReading{
				ID:               fmt.Sprintf("REN-%s-%s-%d", regionID, genType, now.UnixNano()),
				RegionID:         regionID,
				Timestamp:        now,
				GenerationMW:     generationMW,
				GenerationSource: genType,
				StationName:      stationName,
				Frequency:        50.0,
			}
			
			select {
			case ra.eventChan <- reading:
				ra.metrics.ReadingsProcessed++
				ra.metrics.LastReadingTime = now
			default:
				ra.metrics.ErrorsCount++
			}
		}
	}
}

// EnergyAggregator combines readings from multiple grid adapters
type EnergyAggregator struct {
	adapters   []GridAdapter
	outputChan chan EnergyReading
	logger     *common.Logger
	wg         sync.WaitGroup
	stopChan   chan struct{}
}

// NewEnergyAggregator creates a new energy aggregator
func NewEnergyAggregator(logger *common.Logger, adapters ...GridAdapter) *EnergyAggregator {
	return &EnergyAggregator{
		adapters:   adapters,
		outputChan: make(chan EnergyReading, 5000),
		logger:     logger,
		stopChan:   make(chan struct{}),
	}
}

// Start begins aggregating energy readings
func (ea *EnergyAggregator) Start(ctx context.Context) error {
	for _, adapter := range ea.adapters {
		if err := adapter.Start(ctx); err != nil {
			return fmt.Errorf("failed to start adapter: %w", err)
		}
	}
	
	ea.wg.Add(1)
	go ea.aggregate()
	
	ea.logger.Info("Energy aggregator started", "adapters", len(ea.adapters))
	return nil
}

// aggregate merges readings from all adapters
func (ea *EnergyAggregator) aggregate() {
	defer ea.wg.Done()
	
	for {
		select {
		case <-ea.stopChan:
			return
		default:
			for _, adapter := range ea.adapters {
				select {
				case reading := <-adapter.Stream():
					ea.outputChan <- reading
				default:
					continue
				}
			}
			time.Sleep(time.Millisecond)
		}
	}
}

// Stop stops all adapters and aggregation
func (ea *EnergyAggregator) Stop() {
	close(ea.stopChan)
	
	for _, adapter := range ea.adapters {
		adapter.Disconnect()
	}
	
	ea.wg.Wait()
}

// Stream returns the aggregated reading stream
func (ea *EnergyAggregator) Stream() <-chan EnergyReading {
	return ea.outputChan
}

// ToJSON converts reading to JSON
func (er *EnergyReading) ToJSON() string {
	data, _ := json.Marshal(er)
	return string(data)
}
