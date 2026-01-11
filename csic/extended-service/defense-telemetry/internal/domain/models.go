package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// DefenseUnit represents a military unit (air, land, or sea)
type DefenseUnit struct {
	ID             string        `json:"id" db:"id"`
	UnitCode       string        `json:"unit_code" db:"unit_code"`
	UnitName       string        `json:"unit_name" db:"unit_name"`
	UnitType       UnitType      `json:"unit_type" db:"unit_type"`
	Status         UnitStatus    `json:"status" db:"status"`
	CommandID      string        `json:"command_id,omitempty" db:"command_id"`
	BaseLocation   Coordinates   `json:"base_location" db:"base_location"`
	LastContactAt  time.Time     `json:"last_contact_at" db:"last_contact_at"`
	IsDeployed     bool          `json:"is_deployed" db:"is_deployed"`
	DeploymentZone string        `json:"deployment_zone,omitempty" db:"deployment_zone"`
	Capabilities   []string      `json:"capabilities,omitempty" db:"-"`
	Metadata       JSONMap       `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
}

// UnitType represents the type of defense unit
type UnitType string

const (
	UnitTypeAir   UnitType = "air"
	UnitTypeLand  UnitType = "land"
	UnitTypeSea   UnitType = "sea"
	UnitTypeNaval UnitType = "naval"
	UnitTypeCyber UnitType = "cyber"
	UnitTypeSpace UnitType = "space"
)

// UnitStatus represents the operational status of a unit
type UnitStatus string

const (
	UnitStatusActive     UnitStatus = "active"
	UnitStatusStandby    UnitStatus = "standby"
	UnitStatusDeployed   UnitStatus = "deployed"
	UnitStatusMaintenance UnitStatus = "maintenance"
	UnitStatusDamaged    UnitStatus = "damaged"
	UnitStatusDestroyed  UnitStatus = "destroyed"
	UnitStatusOffline    UnitStatus = "offline"
	UnitStatusCompromised UnitStatus = "compromised"
)

// Coordinates represents a geographic location
type Coordinates struct {
	Latitude  decimal.Decimal `json:"latitude"`
	Longitude decimal.Decimal `json:"longitude"`
	Altitude  decimal.Decimal `json:"altitude,omitempty"` // in meters
	Heading   decimal.Decimal `json:"heading,omitempty"`  // in degrees
}

// TelemetryData represents a single telemetry data point from a unit
type TelemetryData struct {
	ID              string          `json:"id" db:"id"`
	UnitID          string          `json:"unit_id" db:"unit_id"`
	Timestamp       time.Time       `json:"timestamp" db:"timestamp"`
	Coordinates     Coordinates     `json:"coordinates" db:"coordinates"`
	Velocity        decimal.Decimal `json:"velocity" db:"velocity"`           // in m/s
	VelocityVector  VelocityVector  `json:"velocity_vector,omitempty" db:"-"`
	FuelLevel       decimal.Decimal `json:"fuel_level" db:"fuel_level"`       // percentage 0-100
	FuelConsumption decimal.Decimal `json:"fuel_consumption" db:"fuel_consumption"` // liters/hour
	AmmunitionCount int             `json:"ammunition_count" db:"ammunition_count"`
	HullIntegrity   decimal.Decimal `json:"hull_integrity" db:"hull_integrity"` // percentage 0-100
	EngineTemp      decimal.Decimal `json:"engine_temp" db:"engine_temp"`       // in Celsius
	InternalTemp    decimal.Decimal `json:"internal_temp" db:"internal_temp"`   // in Celsius
	Humidity        decimal.Decimal `json:"humidity" db:"humidity"`             // percentage
	PowerOutput     decimal.Decimal `json:"power_output" db:"power_output"`     // kW
	PowerConsumption decimal.Decimal `json:"power_consumption" db:"power_consumption"` // kW
	BatteryLevel    decimal.Decimal `json:"battery_level" db:"battery_level"`   // percentage
	SignalStrength  decimal.Decimal `json:"signal_strength" db:"signal_strength"` // dBm
	EncryptionLevel int             `json:"encryption_level" db:"encryption_level"`
	ComponentHealth ComponentHealth `json:"component_health,omitempty" db:"-"`
	Payload         JSONMap         `json:"payload,omitempty" db:"metadata"`
	ProcessedAt     time.Time       `json:"processed_at" db:"processed_at"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

// VelocityVector represents the velocity components
type VelocityVector struct {
	X decimal.Decimal `json:"x"` // m/s
	Y decimal.Decimal `json:"y"` // m/s
	Z decimal.Decimal `json:"z"` // m/s
}

// ComponentHealth represents the health status of various components
type ComponentHealth struct {
	Engine         decimal.Decimal `json:"engine"`
	Hull           decimal.Decimal `json:"hull"`
	Weapons        decimal.Decimal `json:"weapons"`
	Communications decimal.Decimal `json:"communications"`
	Navigation     decimal.Decimal `json:"navigation"`
	PowerSystems   decimal.Decimal `json:"power_systems"`
	Sensors        decimal.Decimal `json:"sensors"`
	Armor          decimal.Decimal `json:"armor"`
}

// ThreatEvent represents a detected threat
type ThreatEvent struct {
	ID              string        `json:"id" db:"id"`
	UnitID          string        `json:"unit_id" db:"unit_id"`
	ThreatType      ThreatType    `json:"threat_type" db:"threat_type"`
	ThreatSource    string        `json:"threat_source" db:"threat_source"`
	Severity        int           `json:"severity" db:"severity"` // 1-5
	Status          ThreatStatus  `json:"status" db:"status"`
	SourceLocation  Coordinates   `json:"source_location" db:"source_location"`
	TargetLocation  Coordinates   `json:"target_location,omitempty" db:"target_location"`
	Velocity        decimal.Decimal `json:"velocity,omitempty" db:"velocity"` // m/s
	EstimatedImpact decimal.Decimal `json:"estimated_impact" db:"estimated_impact"` // damage estimate
	DetectionTime   time.Time     `json:"detection_time" db:"detection_time"`
	FirstSeenAt     time.Time     `json:"first_seen_at" db:"first_seen_at"`
	LastUpdatedAt   time.Time     `json:"last_updated_at" db:"last_updated_at"`
	ResolvedAt      *time.Time    `json:"resolved_at,omitempty" db:"resolved_at"`
	Resolution      string        `json:"resolution,omitempty" db:"resolution"`
	Countermeasures []string      `json:"countermeasures,omitempty" db:"-"`
	EscalationLevel int           `json:"escalation_level" db:"escalation_level"`
	AssignedTo      string        `json:"assigned_to,omitempty" db:"assigned_to"`
	Notes           string        `json:"notes,omitempty" db:"notes"`
	Metadata        JSONMap       `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
}

// ThreatType represents the type of threat
type ThreatType string

const (
	ThreatTypeMissile     ThreatType = "missile"
	ThreatTypeDrone       ThreatType = "drone"
	ThreatTypeCyber       ThreatType = "cyber"
	ThreatTypeElectronic  ThreatType = "electronic_jamming"
	ThreatTypeMaritime    ThreatType = "maritime_vessel"
	ThreatTypeAerial      ThreatType = "aerial"
	ThreatTypeGround      ThreatType = "ground_assault"
	ThreatTypeSubmarine   ThreatType = "submarine"
	ThreatTypeIed         ThreatType = "ied"
	ThreatTypeChemical    ThreatType = "chemical"
	ThreatTypeBiological  ThreatType = "biological"
	ThreatTypeNuclear     ThreatType = "nuclear"
	ThreatTypeUnknown     ThreatType = "unknown"
)

// ThreatStatus represents the status of a threat
type ThreatStatus string

const (
	ThreatStatusDetected   ThreatStatus = "detected"
	ThreatStatusTracking   ThreatStatus = "tracking"
	ThreatStatusAssessing  ThreatStatus = "assessing"
	ThreatStatusEngaging   ThreatStatus = "engaging"
	ThreatStatusNeutralized ThreatStatus = "neutralized"
	ThreatStatusResolved   ThreatStatus = "resolved"
	ThreatStatusFalseAlarm ThreatStatus = "false_alarm"
)

// MaintenanceLog represents a maintenance record
type MaintenanceLog struct {
	ID                   string          `json:"id" db:"id"`
	UnitID               string          `json:"unit_id" db:"unit_id"`
	Component            string          `json:"component" db:"component"`
	MaintenanceType      MaintenanceType `json:"maintenance_type" db:"maintenance_type"`
	ConditionScore       decimal.Decimal `json:"condition_score" db:"condition_score"` // 0-100
	PreviousCondition    decimal.Decimal `json:"previous_condition,omitempty" db:"previous_condition"`
	WorkPerformed        string          `json:"work_performed" db:"work_performed"`
	PartsReplaced        []string        `json:"parts_replaced,omitempty" db:"-"`
	TechnicianID         string          `json:"technician_id" db:"technician_id"`
	EstimatedLife        time.Duration   `json:"estimated_life,omitempty" db:"estimated_life"`
	ScheduledAt          time.Time       `json:"scheduled_at,omitempty" db:"scheduled_at"`
	StartedAt            time.Time       `json:"started_at,omitempty" db:"started_at"`
	CompletedAt          *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
	NextServiceDue       *time.Time      `json:"next_service_due,omitempty" db:"next_service_due"`
	PredictedFailureDate *time.Time      `json:"predicted_failure_date,omitempty" db:"predicted_failure_date"`
	Status               MaintenanceStatus `json:"status" db:"status"`
	Cost                 decimal.Decimal `json:"cost,omitempty" db:"cost"`
	Notes                string          `json:"notes,omitempty" db:"notes"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

// MaintenanceType represents the type of maintenance
type MaintenanceType string

const (
	MaintenanceTypePreventive  MaintenanceType = "preventive"
	MaintenanceTypeCorrective  MaintenanceType = "corrective"
	MaintenanceTypePredictive  MaintenanceType = "predictive"
	MaintenanceTypeEmergency   MaintenanceType = "emergency"
	MaintenanceTypeInspection  MaintenanceType = "inspection"
	MaintenanceTypeOverhaul    MaintenanceType = "overhaul"
)

// MaintenanceStatus represents the status of maintenance
type MaintenanceStatus string

const (
	MaintenanceStatusScheduled MaintenanceStatus = "scheduled"
	MaintenanceStatusInProgress MaintenanceStatus = "in_progress"
	MaintenanceStatusCompleted MaintenanceStatus = "completed"
	MaintenanceStatusCancelled MaintenanceStatus = "cancelled"
	MaintenanceStatusPending   MaintenanceStatus = "pending_parts"
)

// MaintenancePrediction represents predicted maintenance needs
type MaintenancePrediction struct {
	ID                   string          `json:"id" db:"id"`
	UnitID               string          `json:"unit_id" db:"unit_id"`
	Component            string          `json:"component" db:"component"`
	CurrentHealth        decimal.Decimal `json:"current_health" db:"current_health"`
	PredictedFailureDate time.Time       `json:"predicted_failure_date" db:"predicted_failure_date"`
	ConfidenceLevel      decimal.Decimal `json:"confidence_level" db:"confidence_level"` // 0-100
	RiskLevel            string          `json:"risk_level" db:"risk_level"`
	Recommendations      []string        `json:"recommendations,omitempty" db:"-"`
	ModelVersion         string          `json:"model_version" db:"model_version"`
	PredictionDate       time.Time       `json:"prediction_date" db:"prediction_date"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
}

// Deployment represents a unit deployment
type Deployment struct {
	ID            string          `json:"id" db:"id"`
	UnitID        string          `json:"unit_id" db:"unit_id"`
	MissionID     string          `json:"mission_id" db:"mission_id"`
	ZoneID        string          `json:"zone_id" db:"zone_id"`
	StartTime     time.Time       `json:"start_time" db:"start_time"`
	EndTime       *time.Time      `json:"end_time,omitempty" db:"end_time"`
	Status        DeploymentStatus `json:"status" db:"status"`
	StartLocation Coordinates     `json:"start_location" db:"start_location"`
	TargetLocation Coordinates    `json:"target_location" db:"target_location"`
	ActualLocation Coordinates    `json:"actual_location,omitempty" db:"actual_location"`
	RoutePlanned  RouteData       `json:"route_planned,omitempty" db:"-"`
	ResourceUsage ResourceUsage   `json:"resource_usage,omitempty" db:"-"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// DeploymentStatus represents the status of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPlanned     DeploymentStatus = "planned"
	DeploymentStatusEnRoute     DeploymentStatus = "en_route"
	DeploymentStatusOnStation   DeploymentStatus = "on_station"
	DeploymentStatusReturning   DeploymentStatus = "returning"
	DeploymentStatusCompleted   DeploymentStatus = "completed"
	DeploymentStatusAborted     DeploymentStatus = "aborted"
	DeploymentStatusDelayed     DeploymentStatus = "delayed"
)

// RouteData represents planned route information
type RouteData struct {
	TotalDistance   decimal.Decimal `json:"total_distance"` // km
	EstimatedTime   time.Duration   `json:"estimated_time"`
	Waypoints       []Coordinates   `json:"waypoints"`
	FuelRequired    decimal.Decimal `json:"fuel_required"` // liters
	Hazards         []string        `json:"hazards,omitempty"`
}

// ResourceUsage represents resource consumption during deployment
type ResourceUsage struct {
	FuelConsumed    decimal.Decimal `json:"fuel_consumed"`    // liters
	AmmoConsumed    int             `json:"ammo_consumed"`
	SuppliesUsed    decimal.Decimal `json:"supplies_used"`    // kg
	PowerConsumed   decimal.Decimal `json:"power_consumed"`   // kWh
}

// StrategicZone represents a strategic operational zone
type StrategicZone struct {
	ID          string      `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	ZoneType    ZoneType    `json:"zone_type" db:"zone_type"`
	Priority    int         `json:"priority" db:"priority"`
	Boundary    Polygon     `json:"boundary,omitempty" db:"-"`
	CenterPoint Coordinates `json:"center_point" db:"center_point"`
	CoverageRadius decimal.Decimal `json:"coverage_radius" db:"coverage_radius"` // km
	MaxUnits    int         `json:"max_units" db:"max_units"`
	CurrentUnits int        `json:"current_units" db:"current_units"`
	Status      ZoneStatus  `json:"status" db:"status"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// ZoneType represents the type of strategic zone
type ZoneType string

const (
	ZoneTypeCombat     ZoneType = "combat"
	ZoneTypeDefense    ZoneType = "defense"
	ZoneTypeSupport    ZoneType = "support"
	ZoneTypeLogistics  ZoneType = "logistics"
	ZoneTypeRecon      ZoneType = "reconnaissance"
	ZoneTypeRest       ZoneType = "rest_and_recovery"
)

// ZoneStatus represents the status of a zone
type ZoneStatus string

const (
	ZoneStatusActive    ZoneStatus = "active"
	ZoneStatusStandby   ZoneStatus = "standby"
	ZoneStatusInactive  ZoneStatus = "inactive"
	ZoneStatusContested ZoneStatus = "contested"
)

// Polygon represents a geographic polygon
type Polygon struct {
	Points []Coordinates `json:"points"`
}

// Alert represents a system alert
type Alert struct {
	ID          string      `json:"id" db:"id"`
	AlertType   AlertType   `json:"alert_type" db:"alert_type"`
	Severity    AlertSeverity `json:"severity" db:"severity"`
	Title       string      `json:"title" db:"title"`
	Message     string      `json:"message" db:"message"`
	Source      string      `json:"source" db:"source"`
	EntityID    string      `json:"entity_id,omitempty" db:"entity_id"`
	EntityType  string      `json:"entity_type,omitempty" db:"entity_type"`
	IsRead      bool        `json:"is_read" db:"is_read"`
	IsAcked     bool        `json:"is_acked" db:"is_acked"`
	AckedBy     string      `json:"acked_by,omitempty" db:"acked_by"`
	AckedAt     *time.Time  `json:"acked_at,omitempty" db:"acked_at"`
	ExpiresAt   *time.Time  `json:"expires_at,omitempty" db:"expires_at"`
	Metadata    JSONMap     `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeThreat       AlertType = "threat"
	AlertTypeMaintenance  AlertType = "maintenance"
	AlertTypeSystem       AlertType = "system"
	AlertTypeCommunication AlertType = "communication"
	AlertTypeResource     AlertType = "resource"
	AlertTypeSecurity     AlertType = "security"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityEmergency AlertSeverity = "emergency"
)

// Command represents a command center
type Command struct {
	ID          string        `json:"id" db:"id"`
	Name        string        `json:"name" db:"name"`
	CommandType string        `json:"command_type" db:"command_type"`
	ParentID    string        `json:"parent_id,omitempty" db:"parent_id"`
	Location    Coordinates   `json:"location" db:"location"`
	Status      CommandStatus `json:"status" db:"status"`
	ContactInfo JSONMap       `json:"contact_info,omitempty" db:"-"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

// CommandStatus represents the status of a command
type CommandStatus string

const (
	CommandStatusActive   CommandStatus = "active"
	CommandStatusStandby  CommandStatus = "standby"
	CommandStatusOffline  CommandStatus = "offline"
	CommandStatusRelocated CommandStatus = "relocated"
)

// JSONMap is a generic type for JSON metadata
type JSONMap map[string]interface{}

// TelemetryBatch represents a batch of telemetry data
type TelemetryBatch struct {
	UnitID     string           `json:"unit_id"`
	Readings   []TelemetryData  `json:"readings"`
	ReceivedAt time.Time        `json:"received_at"`
}

// TelemetryStats represents aggregated telemetry statistics
type TelemetryStats struct {
	UnitID           string          `json:"unit_id"`
	PeriodStart      time.Time       `json:"period_start"`
	PeriodEnd        time.Time       `json:"period_end"`
	ReadingCount     int             `json:"reading_count"`
	AvgFuelLevel     decimal.Decimal `json:"avg_fuel_level"`
	MinHullIntegrity decimal.Decimal `json:"min_hull_integrity"`
	AvgEngineTemp    decimal.Decimal `json:"avg_engine_temp"`
	TotalDistance    decimal.Decimal `json:"total_distance"` // km
	AlertsGenerated  int             `json:"alerts_generated"`
}
