package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// InfrastructureNode represents an IoT-enabled infrastructure node
type InfrastructureNode struct {
	ID            string          `json:"id" db:"id"`
	NodeType      NodeType        `json:"node_type" db:"node_type"`
	Name          string          `json:"name" db:"name"`
	DistrictID    string          `json:"district_id" db:"district_id"`
	Location      Coordinates     `json:"location" db:"location"`
	IsOperational bool            `json:"is_operational" db:"is_operational"`
	LastPingAt    time.Time       `json:"last_ping_at" db:"last_ping_at"`
	Metadata      JSONMap         `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// NodeType represents the type of infrastructure node
type NodeType string

const (
	NodeTypeTrafficLight     NodeType = "traffic_light"
	NodeTypeTrafficCamera    NodeType = "traffic_camera"
	NodeTypeWaterPump        NodeType = "water_pump"
	NodeTypePowerMeter       NodeType = "power_meter"
	NodeTypeWasteContainer   NodeType = "waste_container"
	NodeTypeStreetLight      NodeType = "street_light"
	NodeTypeParkingSensor    NodeType = "parking_sensor"
	NodeTypeAirQualitySensor NodeType = "air_quality_sensor"
	NodeTypeNoiseSensor      NodeType = "noise_sensor"
	NodeTypeWeatherStation   NodeType = "weather_station"
	NodeTypeGasMeter         NodeType = "gas_meter"
	NodeTypeEVCharger        NodeType = "ev_charger"
)

// Coordinates represents a geographic location
type Coordinates struct {
	Latitude  decimal.Decimal `json:"latitude"`
	Longitude decimal.Decimal `json:"longitude"`
	Altitude  decimal.Decimal `json:"altitude,omitempty"`
}

// SensorReading represents a single sensor reading
type SensorReading struct {
	ID         string          `json:"id" db:"id"`
	NodeID     string          `json:"node_id" db:"node_id"`
	DistrictID string          `json:"district_id" db:"district_id"`
	MetricType MetricType      `json:"metric_type" db:"metric_type"`
	Value      decimal.Decimal `json:"value" db:"value"`
	Unit       string          `json:"unit" db:"unit"`
	Quality    ReadingQuality  `json:"quality" db:"quality"`
	Timestamp  time.Time       `json:"timestamp" db:"timestamp"`
	Metadata   JSONMap         `json:"metadata,omitempty" db:"metadata"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// MetricType represents the type of metric being measured
type MetricType string

const (
	MetricTypeEnergyConsumption    MetricType = "energy_consumption"
	MetricTypeWaterFlow            MetricType = "water_flow"
	MetricTypeWaterLevel           MetricType = "water_level"
	MetricTypeWasteLevel           MetricType = "waste_level"
	MetricTypeAirQualityIndex      MetricType = "air_quality_index"
	MetricTypeNoiseLevel           MetricType = "noise_level"
	MetricTypeTemperature          MetricType = "temperature"
	MetricTypeHumidity             MetricType = "humidity"
	MetricTypeTrafficCount         MetricType = "traffic_count"
	MetricTypeParkingOccupancy     MetricType = "parking_occupancy"
	MetricTypeWindSpeed            MetricType = "wind_speed"
	MetricTypePrecipitation       MetricType = "precipitation"
	MetricTypePressure            MetricType = "pressure"
	MetricTypeCO2Level            MetricType = "co2_level"
	MetricTypePM25                MetricType = "pm25"
	MetricTypePM10                MetricType = "pm10"
)

// ReadingQuality represents the quality of a sensor reading
type ReadingQuality string

const (
	ReadingQualityGood    ReadingQuality = "good"
	ReadingQualitySuspect ReadingQuality = "suspect"
	ReadingQualityBad     ReadingQuality = "bad"
	ReadingQualityEstimated ReadingQuality = "estimated"
)

// CitizenIncident represents a reported incident from citizens
type CitizenIncident struct {
	ID            string          `json:"id" db:"id"`
	IncidentType  IncidentType    `json:"incident_type" db:"incident_type"`
	Category      string          `json:"category" db:"category"`
	DistrictID    string          `json:"district_id" db:"district_id"`
	Location      Coordinates     `json:"location" db:"location"`
	Description   string          `json:"description" db:"description"`
	Status        IncidentStatus  `json:"status" db:"status"`
	Priority      int             `json:"priority" db:"priority"`
	ReporterHash  string          `json:"reporter_hash" db:"reporter_hash"` // Anonymous identifier
	ImageURLs     []string        `json:"image_urls,omitempty" db:"-"`
	AssignedTo    string          `json:"assigned_to,omitempty" db:"assigned_to"`
	ResolvedAt    *time.Time      `json:"resolved_at,omitempty" db:"resolved_at"`
	Resolution    string          `json:"resolution,omitempty" db:"resolution"`
	ResponseTime  *time.Duration  `json:"response_time,omitempty" db:"response_time"`
	Upvotes       int             `json:"upvotes" db:"upvotes"`
	Downvotes     int             `json:"downvotes" db:"downvotes"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// IncidentType represents the type of incident
type IncidentType string

const (
	IncidentTypePothole        IncidentType = "pothole"
	IncidentTypeWaterLeak      IncidentType = "water_leak"
	IncidentTypeStreetLightOut IncidentType = "street_light_out"
	IncidentTypeTrafficSignal  IncidentType = "traffic_signal"
	IncidentTypeGarbageCollection IncidentType = "garbage_collection"
	IncidentTypeNoiseComplaint IncidentType = "noise_complaint"
	IncidentTypeAirQuality     IncidentType = "air_quality"
	IncidentTypeFlooding       IncidentType = "flooding"
	IncidentTypePowerOutage    IncidentType = "power_outage"
	IncidentTypeVandalism      IncidentType = "vandalism"
	IncidentTypeTreeHazard     IncidentType = "tree_hazard"
	IncidentTypeOther          IncidentType = "other"
)

// IncidentStatus represents the status of an incident
type IncidentStatus string

const (
	IncidentStatusReported   IncidentStatus = "reported"
	IncidentStatusDispatched IncidentStatus = "dispatched"
	IncidentStatusInProgress IncidentStatus = "in_progress"
	IncidentStatusPending    IncidentStatus = "pending_parts"
	IncidentStatusResolved   IncidentStatus = "resolved"
	IncidentStatusClosed     IncidentStatus = "closed"
	IncidentStatusDismissed  IncidentStatus = "dismissed"
)

// District represents a city district
type District struct {
	ID          string          `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Code        string          `json:"code" db:"code"`
	Population  int             `json:"population" db:"population"`
	AreaKm2     decimal.Decimal `json:"area_km2" db:"area_km2"`
	Coordinates Coordinates     `json:"coordinates" db:"coordinates"`
	Priority    int             `json:"priority" db:"priority"`
	Metadata    JSONMap         `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// ResourceMetrics represents aggregated resource consumption for a district
type ResourceMetrics struct {
	ID            string          `json:"id" db:"id"`
	DistrictID    string          `json:"district_id" db:"district_id"`
	PeriodType    PeriodType      `json:"period_type" db:"period_type"`
	PeriodStart   time.Time       `json:"period_start" db:"period_start"`
	PeriodEnd     time.Time       `json:"period_end" db:"period_end"`
	EnergyUsed    decimal.Decimal `json:"energy_used" db:"energy_used"`       // kWh
	WaterConsumed decimal.Decimal `json:"water_consumed" db:"water_consumed"` // liters
	WasteGenerated decimal.Decimal `json:"waste_generated" db:"waste_generated"` // kg
	PeakDemand    decimal.Decimal `json:"peak_demand" db:"peak_demand"`       // kW
	AvgAQI        decimal.Decimal `json:"avg_aqi" db:"avg_aqi"`
	Cost          decimal.Decimal `json:"cost" db:"cost"`
	CarbonEmission decimal.Decimal `json:"carbon_emission" db:"carbon_emission"` // kg CO2
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// PeriodType represents the type of time period
type PeriodType string

const (
	PeriodTypeHourly   PeriodType = "hourly"
	PeriodTypeDaily    PeriodType = "daily"
	PeriodTypeWeekly   PeriodType = "weekly"
	PeriodTypeMonthly  PeriodType = "monthly"
	PeriodTypeQuarterly PeriodType = "quarterly"
	PeriodTypeYearly   PeriodType = "yearly"
)

// ServiceRequest represents a service request for city services
type ServiceRequest struct {
	ID            string          `json:"id" db:"id"`
	RequestType   ServiceType    `json:"request_type" db:"request_type"`
	DistrictID    string          `json:"district_id" db:"district_id"`
	Location      Coordinates     `json:"location" db:"location"`
	Description   string          `json:"description" db:"description"`
	Status        RequestStatus   `json:"status" db:"status"`
	Priority      int             `json:"priority" db:"priority"`
	ScheduledDate *time.Time      `json:"scheduled_date,omitempty" db:"scheduled_date"`
	CompletedDate *time.Time      `json:"completed_date,omitempty" db:"completed_date"`
	AssignedTeam  string          `json:"assigned_team" db:"assigned_team"`
	EstimatedCost decimal.Decimal `json:"estimated_cost" db:"estimated_cost"`
	ActualCost    decimal.Decimal `json:"actual_cost" db:"actual_cost"`
	Notes         string          `json:"notes" db:"notes"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// ServiceType represents the type of service request
type ServiceType string

const (
	ServiceTypeStreetCleaning   ServiceType = "street_cleaning"
	ServiceTypeWasteCollection  ServiceType = "waste_collection"
	ServiceTypeTreePruning      ServiceType = "tree_pruning"
	ServiceTypeRoadRepair       ServiceType = "road_repair"
	ServiceTypePavementRepair   ServiceType = "pavement_repair"
	ServiceTypeDrainCleaning    ServiceType = "drain_cleaning"
	ServiceTypeParkMaintenance  ServiceType = "park_maintenance"
	ServiceTypeBuildingRepair   ServiceType = "building_repair"
	ServiceTypeSignageRepair    ServiceType = "signage_repair"
	ServiceTypeOther            ServiceType = "other"
)

// RequestStatus represents the status of a service request
type RequestStatus string

const (
	RequestStatusCreated    RequestStatus = "created"
	RequestStatusScheduled  RequestStatus = "scheduled"
	RequestStatusAssigned   RequestStatus = "assigned"
	RequestStatusInProgress RequestStatus = "in_progress"
	RequestStatusCompleted  RequestStatus = "completed"
	RequestStatusCancelled  RequestStatus = "cancelled"
	RequestStatusOnHold     RequestStatus = "on_hold"
)

// TrafficData represents aggregated traffic data
type TrafficData struct {
	ID            string          `json:"id" db:"id"`
	LocationID    string          `json:"location_id" db:"location_id"`
	DistrictID    string          `json:"district_id" db:"district_id"`
	PeriodStart   time.Time       `json:"period_start" db:"period_start"`
	PeriodEnd     time.Time       `json:"period_end" db:"period_end"`
	VehicleCount  int             `json:"vehicle_count" db:"vehicle_count"`
	AvgSpeed      decimal.Decimal `json:"avg_speed" db:"avg_speed"`         // km/h
	MaxSpeed      decimal.Decimal `json:"max_speed" db:"max_speed"`         // km/h
	CongestionLevel CongestionLevel `json:"congestion_level" db:"congestion_level"`
	Incidents     int             `json:"incidents" db:"incidents"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// CongestionLevel represents the level of traffic congestion
type CongestionLevel string

const (
	CongestionLevelFree    CongestionLevel = "free"
	CongestionLevelLight   CongestionLevel = "light"
	CongestionLevelModerate CongestionLevel = "moderate"
	CongestionLevelHeavy   CongestionLevel = "heavy"
	CongestionLevelSevere  CongestionLevel = "severe"
)

// EnvironmentalMetrics represents environmental monitoring data
type EnvironmentalMetrics struct {
	ID            string          `json:"id" db:"id"`
	DistrictID    string          `json:"district_id" db:"district_id"`
	Timestamp     time.Time       `json:"timestamp" db:"timestamp"`
	AQI           int             `json:"aqi" db:"aqi"`
	AQICategory   AQICategory     `json:"aqi_category" db:"aqi_category"`
	PM25          decimal.Decimal `json:"pm25" db:"pm25"`           // μg/m³
	PM10          decimal.Decimal `json:"pm10" db:"pm10"`           // μg/m³
	CO2           decimal.Decimal `json:"co2" db:"co2"`             // ppm
	NO2           decimal.Decimal `json:"no2" db:"no2"`             // ppb
	O3            decimal.Decimal `json:"o3" db:"o3"`               // ppb
	SO2           decimal.Decimal `json:"so2" db:"so2"`             // ppb
	Temperature   decimal.Decimal `json:"temperature" db:"temperature"` // °C
	Humidity      decimal.Decimal `json:"humidity" db:"humidity"`       // %
	NoiseLevel    decimal.Decimal `json:"noise_level" db:"noise_level"` // dB
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// AQICategory represents the air quality category
type AQICategory string

const (
	AQICategoryGood         AQICategory = "good"
	AQICategoryModerate     AQICategory = "moderate"
	AQICategoryUnhealthySG  AQICategory = "unhealthy_sensitive"
	AQICategoryUnhealthy    AQICategory = "unhealthy"
	AQICategoryVeryUnhealthy AQICategory = "very_unhealthy"
	AQICategoryHazardous    AQICategory = "hazardous"
)

// Alert represents a system alert
type Alert struct {
	ID          string        `json:"id" db:"id"`
	AlertType   AlertType     `json:"alert_type" db:"alert_type"`
	Severity    AlertSeverity `json:"severity" db:"severity"`
	Title       string        `json:"title" db:"title"`
	Message     string        `json:"message" db:"message"`
	Source      string        `json:"source" db:"source"`
	EntityID    string        `json:"entity_id,omitempty" db:"entity_id"`
	EntityType  string        `json:"entity_type,omitempty" db:"entity_type"`
	IsRead      bool          `json:"is_read" db:"is_read"`
	ExpiresAt   *time.Time    `json:"expires_at,omitempty" db:"expires_at"`
	Metadata    JSONMap       `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeResource      AlertType = "resource"
	AlertTypeEnvironmental AlertType = "environmental"
	AlertTypeIncident      AlertType = "incident"
	AlertTypeMaintenance   AlertType = "maintenance"
	AlertTypeSystem        AlertType = "system"
	AlertTypeTraffic       AlertType = "traffic"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityEmergency AlertSeverity = "emergency"
)

// SustainabilityReport represents a sustainability metrics report
type SustainabilityReport struct {
	ID                string          `json:"id" db:"id"`
	DistrictID        string          `json:"district_id" db:"district_id"`
	PeriodStart       time.Time       `json:"period_start" db:"period_start"`
	PeriodEnd         time.Time       `json:"period_end" db:"period_end"`
	TotalEnergyUsed   decimal.Decimal `json:"total_energy_used" db:"total_energy_used"`     // kWh
	EnergySavings     decimal.Decimal `json:"energy_savings" db:"energy_savings"`           // kWh
	EnergySavingsPct  decimal.Decimal `json:"energy_savings_pct" db:"energy_savings_pct"`   // percentage
	TotalWaterUsed    decimal.Decimal `json:"total_water_used" db:"total_water_used"`       // liters
	WaterSaved        decimal.Decimal `json:"water_saved" db:"water_saved"`                 // liters
	WaterSavedPct     decimal.Decimal `json:"water_saved_pct" db:"water_saved_pct"`         // percentage
	TotalWaste        decimal.Decimal `json:"total_waste" db:"total_waste"`                 // kg
	WasteRecycled     decimal.Decimal `json:"waste_recycled" db:"waste_recycled"`           // kg
	RecyclingRate     decimal.Decimal `json:"recycling_rate" db:"recycling_rate"`           // percentage
	CarbonEmissions   decimal.Decimal `json:"carbon_emissions" db:"carbon_emissions"`       // kg CO2
	CarbonReduced     decimal.Decimal `json:"carbon_reduced" db:"carbon_reduced"`           // kg CO2
	CarbonReducedPct  decimal.Decimal `json:"carbon_reduced_pct" db:"carbon_reduced_pct"`   // percentage
	GreenSpacesAdded  decimal.Decimal `json:"green_spaces_added" db:"green_spaces_added"`   // m²
	GeneratedAt       time.Time       `json:"generated_at" db:"generated_at"`
}

// JSONMap is a generic type for JSON metadata
type JSONMap map[string]interface{}

// IoTBatch represents a batch of IoT readings
type IoTBatch struct {
	DeviceID  string         `json:"device_id"`
	Readings  []SensorReading `json:"readings"`
	ReceivedAt time.Time      `json:"received_at"`
}

// DistrictAnalytics represents analytics for a district
type DistrictAnalytics struct {
	DistrictID      string          `json:"district_id"`
	PeriodStart     time.Time       `json:"period_start"`
	PeriodEnd       time.Time       `json:"period_end"`
	TotalNodes      int             `json:"total_nodes"`
	ActiveNodes     int             `json:"active_nodes"`
	TotalReadings   int64           `json:"total_readings"`
	AvgEnergyUsage  decimal.Decimal `json:"avg_energy_usage"`
	AvgWaterUsage   decimal.Decimal `json:"avg_water_usage"`
	AvgWasteLevel   decimal.Decimal `json:"avg_waste_level"`
	AvgAQI          decimal.Decimal `json:"avg_aqi"`
	IncidentCount   int             `json:"incident_count"`
	ResolvedCount   int             `json:"resolved_count"`
	ResolutionRate  decimal.Decimal `json:"resolution_rate"`
	AlertsGenerated int             `json:"alerts_generated"`
}
