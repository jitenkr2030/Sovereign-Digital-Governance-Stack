package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PoolStatus represents the status of a mining pool
type PoolStatus string

const (
	PoolStatusPending    PoolStatus = "PENDING"
	PoolStatusActive     PoolStatus = "ACTIVE"
	PoolStatusSuspended  PoolStatus = "SUSPENDED"
	PoolStatusRevoked    PoolStatus = "REVOKED"
	PoolStatusExpired    PoolStatus = "EXPIRED"
)

// EnergySourceType represents the type of energy source
type EnergySourceType string

const (
	EnergySourceGrid      EnergySourceType = "GRID"
	EnergySourceSolar     EnergySourceType = "SOLAR"
	EnergySourceHydro     EnergySourceType = "HYDRO"
	EnergySourceWind      EnergySourceType = "WIND"
	EnergySourceNuclear   EnergySourceType = "NUCLEAR"
	EnergySourceCoal      EnergySourceType = "COAL"
	EnergySourceNaturalGas EnergySourceType = "NATURAL_GAS"
	EnergySourceMixed     EnergySourceType = "MIXED"
)

// MachineType represents the type of mining machine
type MachineType string

const (
	MachineTypeASIC MachineType = "ASIC"
	MachineTypeGPU  MachineType = "GPU"
	MachineTypeFPGA MachineType = "FPGA"
	MachineTypeCPU  MachineType = "CPU"
)

// ViolationSeverity represents the severity of a compliance violation
type ViolationSeverity string

const (
	ViolationSeverityLow      ViolationSeverity = "LOW"
	ViolationSeverityMedium   ViolationSeverity = "MEDIUM"
	ViolationSeverityHigh     ViolationSeverity = "HIGH"
	ViolationSeverityCritical ViolationSeverity = "CRITICAL"
)

// ViolationType represents the type of compliance violation
type ViolationType string

const (
	ViolationTypeExcessEnergy         ViolationType = "EXCESS_ENERGY"
	ViolationTypeUnregisteredHardware  ViolationType = "UNREGISTERED_HARDWARE"
	ViolationTypeBannedRegion          ViolationType = "BANNED_REGION"
	ViolationTypeEnergySourceViolation ViolationType = "ENERGY_SOURCE_VIOLATION"
	ViolationTypeHashRateAnomaly       ViolationType = "HASH_RATE_ANOMALY"
	ViolationTypeLicenseExpired        ViolationType = "LICENSE_EXPIRED"
	ViolationTypeCarbonLimitExceeded   ViolationType = "CARBON_LIMIT_EXCEEDED"
	ViolationTypeReportingViolation    ViolationType = "REPORTING_VIOLATION"
)

// ViolationStatus represents the status of a violation
type ViolationStatus string

const (
	ViolationStatusOpen       ViolationStatus = "OPEN"
	ViolationStatusInvestigating ViolationStatus = "INVESTIGATING"
	ViolationStatusResolved   ViolationStatus = "RESOLVED"
	ViolationStatusDismissed  ViolationStatus = "DISMISSED"
)

// RegionStatus represents the regulatory status of a region
type RegionStatus string

const (
	RegionStatusAllowed    RegionStatus = "ALLOWED"
	RegionStatusRestricted RegionStatus = "RESTRICTED"
	RegionStatusBanned     RegionStatus = "BANNED"
)

// MiningPool represents a registered cryptocurrency mining operation
type MiningPool struct {
	ID                     uuid.UUID         `json:"id" db:"id"`
	LicenseNumber          string           `json:"license_number" db:"license_number"`
	Name                   string           `json:"name" db:"name"`
	OwnerEntityID          uuid.UUID        `json:"owner_entity_id" db:"owner_entity_id"`
	OwnerEntityName        string           `json:"owner_entity_name" db:"owner_entity_name"`
	Status                 PoolStatus       `json:"status" db:"status"`
	RegionCode             string           `json:"region_code" db:"region_code"`
	MaxAllowedEnergyKW     decimal.Decimal  `json:"max_allowed_energy_kw" db:"max_allowed_energy_kw"`
	CurrentEnergyUsageKW   decimal.Decimal  `json:"current_energy_usage_kw" db:"current_energy_usage_kw"`
	AllowedEnergySources   []EnergySourceType `json:"allowed_energy_sources" db:"allowed_energy_sources"`
	TotalHashRateTH        decimal.Decimal  `json:"total_hash_rate_th" db:"total_hash_rate_th"`
	ActiveMachineCount     int              `json:"active_machine_count" db:"active_machine_count"`
	TotalMachineCount      int              `json:"total_machine_count" db:"total_machine_count"`
	CarbonFootprintKG      decimal.Decimal  `json:"carbon_footprint_kg" db:"carbon_footprint_kg"`
	LicenseIssuedAt        time.Time        `json:"license_issued_at" db:"license_issued_at"`
	LicenseExpiresAt       time.Time        `json:"license_expires_at" db:"license_expires_at"`
	LastInspectionAt       *time.Time       `json:"last_inspection_at,omitempty" db:"last_inspection_at"`
	ComplianceCertificate  *ComplianceCertificate `json:"compliance_certificate,omitempty"`
	ContactEmail           string           `json:"contact_email" db:"contact_email"`
	ContactPhone           string           `json:"contact_phone" db:"contact_phone"`
	FacilityAddress        string           `json:"facility_address" db:"facility_address"`
	GPSCoordinates         *GPSCoordinates  `json:"gps_coordinates,omitempty" db:"gps_coordinates"`
	CreatedAt              time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at" db:"updated_at"`
	Notes                  string           `json:"notes,omitempty" db:"notes"`
}

// GPSCoordinates represents geographic coordinates
type GPSCoordinates struct {
	Latitude  decimal.Decimal `json:"latitude" db:"latitude"`
	Longitude decimal.Decimal `json:"longitude" db:"longitude"`
	Altitude  decimal.Decimal `json:"altitude,omitempty" db:"altitude"`
}

// MiningMachine represents a registered mining machine
type MiningMachine struct {
	ID                 uuid.UUID      `json:"id" db:"id"`
	SerialNumber       string         `json:"serial_number" db:"serial_number"`
	PoolID             uuid.UUID      `json:"pool_id" db:"pool_id"`
	ModelType          MachineType    `json:"model_type" db:"model_type"`
	Manufacturer       string         `json:"manufacturer" db:"manufacturer"`
	ModelName          string         `json:"model_name" db:"model_name"`
	HashRateSpecTH     decimal.Decimal `json:"hash_rate_spec_th" db:"hash_rate_spec_th"`
	PowerSpecWatts     decimal.Decimal `json:"power_spec_watts" db:"power_spec_watts"`
	LocationID         string         `json:"location_id" db:"location_id"`
	LocationDescription string        `json:"location_description" db:"location_description"`
	GPSCoordinates     *GPSCoordinates `json:"gps_coordinates,omitempty" db:"gps_coordinates"`
	InstallationDate   time.Time      `json:"installation_date" db:"installation_date"`
	LastMaintenanceDate *time.Time    `json:"last_maintenance_date,omitempty" db:"last_maintenance_date"`
	Status             MachineStatus  `json:"status" db:"status"`
	IsActive           bool           `json:"is_active" db:"is_active"`
	CurrentHashRateTH  decimal.Decimal `json:"current_hash_rate_th" db:"current_hash_rate_th"`
	CurrentPowerWatts  decimal.Decimal `json:"current_power_watts" db:"current_power_watts"`
	TotalEnergyKWh     decimal.Decimal `json:"total_energy_kwh" db:"total_energy_kwh"`
	UptimeHours        decimal.Decimal `json:"uptime_hours" db:"uptime_hours"`
	CreatedAt          time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at" db:"updated_at"`
	DecommissionedAt   *time.Time     `json:"decommissioned_at,omitempty" db:"decommissioned_at"`
}

// MachineStatus represents the status of a mining machine
type MachineStatus string

const (
	MachineStatusPending      MachineStatus = "PENDING"
	MachineStatusActive       MachineStatus = "ACTIVE"
	MachineStatusMaintenance  MachineStatus = "MAINTENANCE"
	MachineStatusOffline      MachineStatus = "OFFLINE"
	MachineStatusDecommissioned MachineStatus = "DECOMMISSIONED"
)

// EnergyConsumptionLog represents energy consumption telemetry
type EnergyConsumptionLog struct {
	ID                  uuid.UUID         `json:"id" db:"id"`
	PoolID              uuid.UUID         `json:"pool_id" db:"pool_id"`
	Timestamp           time.Time         `json:"timestamp" db:"timestamp"`
	DurationMinutes     int               `json:"duration_minutes" db:"duration_minutes"`
	PowerUsageKWh       decimal.Decimal   `json:"power_usage_kwh" db:"power_usage_kwh"`
	PeakPowerKW         decimal.Decimal   `json:"peak_power_kw" db:"peak_power_kw"`
	AveragePowerKW      decimal.Decimal   `json:"average_power_kw" db:"average_power_kw"`
	EnergySource        EnergySourceType  `json:"energy_source" db:"energy_source"`
	SourceMixPercentage map[string]int    `json:"source_mix_percentage,omitempty" db:"source_mix_percentage"`
	CarbonEmissionsKG   decimal.Decimal   `json:"carbon_emissions_kg" db:"carbon_emissions_kg"`
	GridRegionCode      string            `json:"grid_region_code" db:"grid_region_code"`
	TemperatureC        decimal.Decimal   `json:"temperature_c,omitempty" db:"temperature_c"`
	HumidityPercent     decimal.Decimal   `json:"humidity_percent,omitempty" db:"humidity_percent"`
	DataQuality         DataQuality       `json:"data_quality" db:"data_quality"`
	SubmittedAt         time.Time         `json:"submitted_at" db:"submitted_at"`
	Checksum            string            `json:"checksum" db:"checksum"`
}

// DataQuality represents the quality of telemetry data
type DataQuality string

const (
	DataQualityGood     DataQuality = "GOOD"
	DataQualitySuspect  DataQuality = "SUSPECT"
	DataQualityMissing  DataQuality = "MISSING"
	DataQualityEstimated DataQuality = "ESTIMATED"
)

// HashRateMetric represents hash rate telemetry
type HashRateMetric struct {
	ID               uuid.UUID      `json:"id" db:"id"`
	PoolID           uuid.UUID      `json:"pool_id" db:"pool_id"`
	Timestamp        time.Time      `json:"timestamp" db:"timestamp"`
	ReportedHashRate decimal.Decimal `json:"reported_hash_rate_th" db:"reported_hash_rate_th"`
	CalculatedHashRate decimal.Decimal `json:"calculated_hash_rate_th" db:"calculated_hash_rate_th"`
	ActiveWorkerCount int           `json:"active_worker_count" db:"active_worker_count"`
	StaleWorkerCount  int           `json:"stale_worker_count" db:"stale_worker_count"`
	TotalShares      decimal.Decimal `json:"total_shares" db:"total_shares"`
	AcceptedShares   decimal.Decimal `json:"accepted_shares" db:"accepted_shares"`
	RejectedShares   decimal.Decimal `json:"rejected_shares" db:"rejected_shares"`
	Difficulty       decimal.Decimal `json:"difficulty" db:"difficulty"`
	PoolFeePercent   decimal.Decimal `json:"pool_fee_percent" db:"pool_fee_percent"`
	DataQuality      DataQuality     `json:"data_quality" db:"data_quality"`
	SubmittedAt      time.Time       `json:"submitted_at" db:"submitted_at"`
	Checksum         string          `json:"checksum" db:"checksum"`
}

// ComplianceViolation represents a regulatory compliance violation
type ComplianceViolation struct {
	ID            uuid.UUID          `json:"id" db:"id"`
	PoolID        uuid.UUID          `json:"pool_id" db:"pool_id"`
	PoolName      string             `json:"pool_name" db:"pool_name"`
	ViolationType ViolationType      `json:"violation_type" db:"violation_type"`
	Severity      ViolationSeverity  `json:"severity" db:"severity"`
	Status        ViolationStatus    `json:"status" db:"status"`
	Title         string             `json:"title" db:"title"`
	Description   string             `json:"description" db:"description"`
	Details       map[string]interface{} `json:"details,omitempty" db:"details"`
	Evidence      []string           `json:"evidence,omitempty" db:"evidence"`
	DetectedAt    time.Time          `json:"detected_at" db:"detected_at"`
	ResolvedAt    *time.Time         `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy    *uuid.UUID         `json:"resolved_by,omitempty" db:"resolved_by"`
	Resolution    *string            `json:"resolution,omitempty" db:"resolution"`
	AssignedTo    *uuid.UUID         `json:"assigned_to,omitempty" db:"assigned_to"`
	Notes         []ViolationNote    `json:"notes,omitempty"`
	AlertsSent    int                `json:"alerts_sent" db:"alerts_sent"`
	CreatedAt     time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" db:"updated_at"`
}

// ViolationNote represents a note added to a violation
type ViolationNote struct {
	ID         uuid.UUID `json:"id"`
	ViolationID uuid.UUID `json:"violation_id"`
	AuthorID   uuid.UUID `json:"author_id"`
	AuthorName string    `json:"author_name"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	IsInternal bool      `json:"is_internal"`
}

// ComplianceCertificate represents a compliance certificate for a mining pool
type ComplianceCertificate struct {
	ID                 uuid.UUID   `json:"id" db:"id"`
	PoolID             uuid.UUID   `json:"pool_id" db:"pool_id"`
	CertificateNumber  string      `json:"certificate_number" db:"certificate_number"`
	Status             CertificateStatus `json:"status" db:"status"`
	IssueDate          time.Time   `json:"issue_date" db:"issue_date"`
	ExpiryDate         time.Time   `json:"expiry_date" db:"expiry_date"`
	InspectionDate     *time.Time  `json:"inspection_date,omitempty" db:"inspection_date"`
	InspectorID        *uuid.UUID  `json:"inspector_id,omitempty" db:"inspector_id"`
	EnergyRating       string      `json:"energy_rating" db:"energy_rating"`
	CarbonRating       string      `json:"carbon_rating" db:"carbon_rating"`
	ComplianceScore    decimal.Decimal `json:"compliance_score" db:"compliance_score"`
	TermsConditions    []string    `json:"terms_conditions" db:"terms_conditions"`
	Restrictions       []string    `json:"restrictions" db:"restrictions"`
	DigitalSignature   string      `json:"digital_signature" db:"digital_signature"`
	CreatedAt          time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at" db:"updated_at"`
}

// CertificateStatus represents the status of a compliance certificate
type CertificateStatus string

const (
	CertificateStatusValid     CertificateStatus = "VALID"
	CertificateStatusExpired   CertificateStatus = "EXPIRED"
	CertificateStatusSuspended CertificateStatus = "SUSPENDED"
	CertificateStatusRevoked   CertificateStatus = "REVOKED"
	CertificateStatusPending   CertificateStatus = "PENDING"
)

// EnergySourceMix represents the mix of energy sources
type EnergySourceMix struct {
	PoolID       uuid.UUID              `json:"pool_id" db:"pool_id"`
	PeriodStart  time.Time              `json:"period_start" db:"period_start"`
	PeriodEnd    time.Time              `json:"period_end" db:"period_end"`
	Mix          map[EnergySourceType]int `json:"mix" db:"mix"`
	TotalKWh     decimal.Decimal        `json:"total_kwh" db:"total_kwh"`
	RenewablePercent decimal.Decimal    `json:"renewable_percent" db:"renewable_percent"`
	CarbonIntensity decimal.Decimal     `json:"carbon_intensity" db:"carbon_intensity"`
	CalculatedAt time.Time              `json:"calculated_at" db:"calculated_at"`
}

// CarbonFootprintReport represents a carbon footprint report
type CarbonFootprintReport struct {
	PoolID             uuid.UUID         `json:"pool_id" db:"pool_id"`
	PoolName           string            `json:"pool_name" db:"pool_name"`
	ReportPeriodStart  time.Time         `json:"report_period_start" db:"report_period_start"`
	ReportPeriodEnd    time.Time         `json:"report_period_end" db:"report_period_end"`
	TotalEnergyKWh     decimal.Decimal   `json:"total_energy_kwh" db:"total_energy_kwh"`
	TotalCarbonKG      decimal.Decimal   `json:"total_carbon_kg" db:"total_carbon_kg"`
	CarbonPerKWh       decimal.Decimal   `json:"carbon_per_kwh" db:"carbon_per_kwh"`
	RenewableEnergyKWh decimal.Decimal   `json:"renewable_energy_kwh" db:"renewable_energy_kwh"`
	RenewablePercent   decimal.Decimal   `json:"renewable_percent" db:"renewable_percent"`
	BySource           map[string]decimal.Decimal `json:"by_source" db:"by_source"`
	ComparedToAverage  decimal.Decimal   `json:"compared_to_average" db:"compared_to_average"`
	Rating             string            `json:"rating" db:"rating"`
	GeneratedAt        time.Time         `json:"generated_at" db:"generated_at"`
}

// MiningSummaryReport represents a comprehensive mining summary report
type MiningSummaryReport struct {
	PoolID              uuid.UUID         `json:"pool_id" db:"pool_id"`
	ReportPeriodStart   time.Time         `json:"report_period_start" db:"report_period_start"`
	ReportPeriodEnd     time.Time         `json:"report_period_end" db:"report_period_end"`

	// Pool Information
	PoolName            string            `json:"pool_name" db:"pool_name"`
	LicenseNumber       string            `json:"license_number" db:"license_number"`
	RegionCode          string            `json:"region_code" db:"region_code"`
	Status              PoolStatus        `json:"status" db:"status"`

	// Machine Summary
	TotalMachines       int               `json:"total_machines" db:"total_machines"`
	ActiveMachines      int               `json:"active_machines" db:"active_machines"`
	OfflineMachines     int               `json:"offline_machines" db:"offline_machines"`
	TotalHashRateTH     decimal.Decimal   `json:"total_hash_rate_th" db:"total_hash_rate_th"`
	AverageEfficiency   decimal.Decimal   `json:"average_efficiency" db:"average_efficiency"`

	// Energy Summary
	TotalEnergyKWh      decimal.Decimal   `json:"total_energy_kwh" db:"total_energy_kwh"`
	PeakPowerKW         decimal.Decimal   `json:"peak_power_kw" db:"peak_power_kw"`
	AveragePowerKW      decimal.Decimal   `json:"average_power_kw" db:"average_power_kw"`
	EnergyLimitUtilization decimal.Decimal `json:"energy_limit_utilization" db:"energy_limit_utilization"`

	// Carbon Summary
	TotalCarbonKG       decimal.Decimal   `json:"total_carbon_kg" db:"total_carbon_kg"`
	CarbonPerTH         decimal.Decimal   `json:"carbon_per_th" db:"carbon_per_th"`
	RenewablePercent    decimal.Decimal   `json:"renewable_percent" db:"renewable_percent"`

	// Compliance Summary
	OpenViolations      int               `json:"open_violations" db:"open_violations"`
	ResolvedViolations  int               `json:"resolved_violations" db:"resolved_violations"`
	ComplianceScore     decimal.Decimal   `json:"compliance_score" db:"compliance_score"`
	CertificateStatus   CertificateStatus `json:"certificate_status" db:"certificate_status"`

	GeneratedAt         time.Time         `json:"generated_at" db:"generated_at"`
}

// PoolFilter represents filter criteria for listing mining pools
type PoolFilter struct {
	Statuses          []PoolStatus       `json:"statuses,omitempty"`
	RegionCodes       []string           `json:"region_codes,omitempty"`
	EntityIDs         []uuid.UUID        `json:"entity_ids,omitempty"`
	MinEnergyUsage    *decimal.Decimal   `json:"min_energy_usage,omitempty"`
	MaxEnergyUsage    *decimal.Decimal   `json:"max_energy_usage,omitempty"`
	MinHashRate       *decimal.Decimal   `json:"min_hash_rate,omitempty"`
	LicenseExpired    *bool              `json:"license_expired,omitempty"`
	ComplianceRating  *string            `json:"compliance_rating,omitempty"`
}

// ViolationFilter represents filter criteria for listing violations
type ViolationFilter struct {
	Types            []ViolationType     `json:"types,omitempty"`
	Severities       []ViolationSeverity `json:"severities,omitempty"`
	Statuses         []ViolationStatus   `json:"statuses,omitempty"`
	PoolIDs          []uuid.UUID         `json:"pool_ids,omitempty"`
	DetectedAfter    *time.Time          `json:"detected_after,omitempty"`
	DetectedBefore   *time.Time          `json:"detected_before,omitempty"`
	MinSeverity      *ViolationSeverity  `json:"min_severity,omitempty"`
}

// RegionalEnergyStats represents energy statistics by region
type RegionalEnergyStats struct {
	RegionCode          string            `json:"region_code" db:"region_code"`
	TotalPools          int               `json:"total_pools" db:"total_pools"`
	ActivePools         int               `json:"active_pools" db:"active_pools"`
	TotalEnergyKWh      decimal.Decimal   `json:"total_energy_kwh" db:"total_energy_kwh"`
	PeakPowerKW         decimal.Decimal   `json:"peak_power_kw" db:"peak_power_kw"`
	AveragePowerKW      decimal.Decimal   `json:"average_power_kw" db:"average_power_kw"`
	TotalCarbonKG       decimal.Decimal   `json:"total_carbon_kg" db:"total_carbon_kg"`
	AverageCarbonIntensity decimal.Decimal `json:"average_carbon_intensity" db:"average_carbon_intensity"`
	RenewablePercent    decimal.Decimal   `json:"renewable_percent" db:"renewable_percent"`
	OpenViolations      int               `json:"open_violations" db:"open_violations"`
	PeriodStart         time.Time         `json:"period_start" db:"period_start"`
	PeriodEnd           time.Time         `json:"period_end" db:"period_end"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalPools            int                        `json:"total_pools" db:"total_pools"`
	ActivePools           int                        `json:"active_pools" db:"active_pools"`
	SuspendedPools        int                        `json:"suspended_pools" db:"suspended_pools"`
	TotalMachines         int                        `json:"total_machines" db:"total_machines"`
	ActiveMachines        int                        `json:"active_machines" db:"active_machines"`
	TotalEnergyKWh        decimal.Decimal            `json:"total_energy_kwh" db:"total_energy_kwh"`
	PeakPowerKW           decimal.Decimal            `json:"peak_power_kw" db:"peak_power_kw"`
	TotalHashRateTH       decimal.Decimal            `json:"total_hash_rate_th" db:"total_hash_rate_th"`
	TotalCarbonKG         decimal.Decimal            `json:"total_carbon_kg" db:"total_carbon_kg"`
	AverageRenewablePercent decimal.Decimal          `json:"average_renewable_percent" db:"average_renewable_percent"`
	OpenViolations        int                        `json:"open_violations" db:"open_violations"`
	CriticalViolations    int                        `json:"critical_violations" db:"critical_violations"`
	ExpiringCertificates  int                        `json:"expiring_certificates" db:"expiring_certificates"`
	ByRegion              []RegionalEnergyStats      `json:"by_region" db:"by_region"`
	ViolationTrend        []TrendPoint               `json:"violation_trend" db:"violation_trend"`
	EnergyTrend           []TrendPoint               `json:"energy_trend" db:"energy_trend"`
	UpdatedAt             time.Time                  `json:"updated_at" db:"updated_at"`
}

// TrendPoint represents a data point for trend analysis
type TrendPoint struct {
	Timestamp time.Time       `json:"timestamp" db:"timestamp"`
	Value     decimal.Decimal `json:"value" db:"value"`
	Label     string          `json:"label,omitempty" db:"label"`
}

// EnergyTelemetryPacket represents an incoming energy telemetry packet
type EnergyTelemetryPacket struct {
	PoolID              uuid.UUID         `json:"pool_id" db:"pool_id"`
	Timestamp           time.Time         `json:"timestamp" db:"timestamp"`
	DurationMinutes     int               `json:"duration_minutes" db:"duration_minutes"`
	PowerUsageKWh       decimal.Decimal   `json:"power_usage_kwh" db:"power_usage_kwh"`
	PeakPowerKW         decimal.Decimal   `json:"peak_power_kw" db:"peak_power_kw"`
	AveragePowerKW      decimal.Decimal   `json:"average_power_kw" db:"average_power_kw"`
	EnergySource        EnergySourceType  `json:"energy_source" db:"energy_source"`
	SourceMixPercentage map[string]int    `json:"source_mix_percentage,omitempty" db:"source_mix_percentage"`
	GridRegionCode      string            `json:"grid_region_code" db:"grid_region_code"`
	TemperatureC        decimal.Decimal   `json:"temperature_c,omitempty" db:"temperature_c"`
	HumidityPercent     decimal.Decimal   `json:"humidity_percent,omitempty" db:"humidity_percent"`
}

// HashRateTelemetryPacket represents an incoming hash rate telemetry packet
type HashRateTelemetryPacket struct {
	PoolID              uuid.UUID      `json:"pool_id" db:"pool_id"`
	Timestamp           time.Time      `json:"timestamp" db:"timestamp"`
	ReportedHashRate    decimal.Decimal `json:"reported_hash_rate_th" db:"reported_hash_rate_th"`
	ActiveWorkerCount   int            `json:"active_worker_count" db:"active_worker_count"`
	StaleWorkerCount    int            `json:"stale_worker_count" db:"stale_worker_count"`
	TotalShares         decimal.Decimal `json:"total_shares" db:"total_shares"`
	AcceptedShares      decimal.Decimal `json:"accepted_shares" db:"accepted_shares"`
	RejectedShares      decimal.Decimal `json:"rejected_shares" db:"rejected_shares"`
	Difficulty          decimal.Decimal `json:"difficulty" db:"difficulty"`
	PoolFeePercent      decimal.Decimal `json:"pool_fee_percent" db:"pool_fee_percent"`
}

// MachineRegistrationRequest represents a request to register a mining machine
type MachineRegistrationRequest struct {
	SerialNumber        string         `json:"serial_number" binding:"required"`
	ModelType           MachineType    `json:"model_type" binding:"required"`
	Manufacturer        string         `json:"manufacturer" binding:"required"`
	ModelName           string         `json:"model_name" binding:"required"`
	HashRateSpecTH      decimal.Decimal `json:"hash_rate_spec_th" binding:"required,gt=0"`
	PowerSpecWatts      decimal.Decimal `json:"power_spec_watts" binding:"required,gt=0"`
	LocationID          string         `json:"location_id" binding:"required"`
	LocationDescription string         `json:"location_description" binding:"required"`
	GPSCoordinates      *GPSCoordinates `json:"gps_coordinates,omitempty"`
	InstallationDate    time.Time      `json:"installation_date" binding:"required"`
}

// PoolRegistrationRequest represents a request to register a mining pool
type PoolRegistrationRequest struct {
	Name                 string           `json:"name" binding:"required,min=3,max=200"`
	OwnerEntityID        uuid.UUID        `json:"owner_entity_id" binding:"required"`
	RegionCode           string           `json:"region_code" binding:"required"`
	MaxAllowedEnergyKW   decimal.Decimal  `json:"max_allowed_energy_kw" binding:"required,gt=0"`
	AllowedEnergySources []EnergySourceType `json:"allowed_energy_sources" binding:"required,min=1"`
	ContactEmail         string           `json:"contact_email" binding:"required,email"`
	ContactPhone         string           `json:"contact_phone" binding:"required"`
	FacilityAddress      string           `json:"facility_address" binding:"required"`
	GPSCoordinates       *GPSCoordinates  `json:"gps_coordinates,omitempty"`
}
