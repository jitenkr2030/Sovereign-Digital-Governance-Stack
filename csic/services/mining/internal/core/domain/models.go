package domain

import (
	"time"

	"github.com/google/uuid"
)

// OperationStatus represents the status of a mining operation
type OperationStatus string

const (
	StatusActive             OperationStatus = "ACTIVE"
	StatusSuspended          OperationStatus = "SUSPENDED"
	StatusShutdownOrdered    OperationStatus = "SHUTDOWN_ORDERED"
	StatusNonCompliant       OperationStatus = "NON_COMPLIANT"
	StatusShutdownExecuted   OperationStatus = "SHUTDOWN_EXECUTED"
	StatusPendingRegistration OperationStatus = "PENDING_REGISTRATION"
)

// QuotaType represents the type of mining quota
type QuotaType string

const (
	QuotaFixed      QuotaType = "FIXED"
	QuotaDynamicGrid QuotaType = "DYNAMIC_GRID"
)

// CommandType represents the type of shutdown command
type CommandType string

const (
	CommandGraceful   CommandType = "GRACEFUL"
	CommandImmediate  CommandType = "IMMEDIATE"
	CommandForceKill  CommandType = "FORCE_KILL"
)

// CommandStatus represents the status of a shutdown command
type CommandStatus string

const (
	CommandIssued      CommandStatus = "ISSUED"
	CommandAcknowledged CommandStatus = "ACKNOWLEDGED"
	CommandExecuted    CommandStatus = "EXECUTED"
	CommandFailed      CommandStatus = "FAILED"
)

// MiningOperation represents a registered mining operation in the national registry
type MiningOperation struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	OperatorName   string          `json:"operator_name" db:"operator_name"`
	WalletAddress  string          `json:"wallet_address" db:"wallet_address"`
	CurrentHashrate float64        `json:"current_hashrate" db:"current_hashrate"`
	Status         OperationStatus `json:"status" db:"status"`
	QuotaID        *uuid.UUID      `json:"quota_id,omitempty" db:"quota_id"`
	Location       string          `json:"location" db:"location"`
	Region         string          `json:"region" db:"region"`
	MachineType    string          `json:"machine_type" db:"machine_type"`
	RegisteredAt   time.Time       `json:"registered_at" db:"registered_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	LastReportAt   *time.Time      `json:"last_report_at,omitempty" db:"last_report_at"`
	ViolationCount int             `json:"violation_count" db:"violation_count"`
	Metadata       string          `json:"metadata,omitempty" db:"metadata"`
}

// HashrateRecord represents a hashrate telemetry record
type HashrateRecord struct {
	ID           int64     `json:"id" db:"id"`
	OperationID  uuid.UUID `json:"operation_id" db:"operation_id"`
	Hashrate     float64   `json:"hashrate" db:"hashrate"`
	Unit         string    `json:"unit" db:"unit"`
	BlockHeight  uint64    `json:"block_height" db:"block_height"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	SubmittedAt  time.Time `json:"submitted_at" db:"submitted_at"`
}

// MiningQuota represents a hashrate quota assigned to a mining operation
type MiningQuota struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OperationID  uuid.UUID  `json:"operation_id" db:"operation_id"`
	MaxHashrate  float64    `json:"max_hashrate" db:"max_hashrate"`
	QuotaType    QuotaType  `json:"quota_type" db:"quota_type"`
	ValidFrom    time.Time  `json:"valid_from" db:"valid_from"`
	ValidTo      *time.Time `json:"valid_to,omitempty" db:"valid_to"`
	Region       string     `json:"region" db:"region"`
	Priority     int        `json:"priority" db:"priority"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// ShutdownCommand represents a remote shutdown command issued to a mining operation
type ShutdownCommand struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	OperationID uuid.UUID     `json:"operation_id" db:"operation_id"`
	CommandType CommandType   `json:"command_type" db:"command_type"`
	Reason      string        `json:"reason" db:"reason"`
	Status      CommandStatus `json:"status" db:"status"`
	IssuedAt    time.Time     `json:"issued_at" db:"issued_at"`
	ExecutedAt  *time.Time    `json:"executed_at,omitempty" db:"executed_at"`
	AckedAt     *time.Time    `json:"acked_at,omitempty" db:"acked_at"`
	IssuedBy    string        `json:"issued_by" db:"issued_by"`
	ExpiresAt   *time.Time    `json:"expires_at,omitempty" db:"expires_at"`
	Result      string        `json:"result,omitempty" db:"result"`
}

// QuotaViolation represents a quota violation record
type QuotaViolation struct {
	ID          uuid.UUID `json:"id" db:"id"`
	OperationID uuid.UUID `json:"operation_id" db:"operation_id"`
	QuotaID     uuid.UUID `json:"quota_id" db:"quota_id"`
	ReportedHashrate float64 `json:"reported_hashrate" db:"reported_hashrate"`
	MaxHashrate float64    `json:"max_hashrate" db:"max_hashrate"`
	RecordedAt  time.Time  `json:"recorded_at" db:"recorded_at"`
	Consecutive bool       `json:"consecutive" db:"consecutive"`
	AutomaticShutdown bool `json:"automatic_shutdown" db:"automatic_shutdown"`
}

// Location represents geographic location of a mining operation
type Location struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Address    string  `json:"address"`
	FacilityID string  `json:"facility_id"`
	GridZone   string  `json:"grid_zone"`
}

// TelemetryData represents incoming hashrate telemetry from mining operations
type TelemetryData struct {
	OperationID uuid.UUID `json:"operation_id"`
	Hashrate    float64   `json:"hashrate"`
	Unit        string    `json:"unit"`
	BlockHeight uint64    `json:"block_height"`
	Timestamp   time.Time `json:"timestamp"`
}

// RegistryStats represents statistics for the national mining registry
type RegistryStats struct {
	TotalOperations    int64   `json:"total_operations"`
	ActiveOperations   int64   `json:"active_operations"`
	SuspendedOperations int64  `json:"suspended_operations"`
	ShutdownOperations int64   `json:"shutdown_operations"`
	TotalHashrate      float64 `json:"total_hashrate"`
	QuotaViolations    int64   `json:"quota_violations"`
	PendingCommands    int64   `json:"pending_commands"`
}
