package domain

import (
	"time"
)

// Entity represents the base entity interface
type Entity interface {
	GetID() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

// BaseEntity contains common fields for all entities
type BaseEntity struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetID returns the entity ID
func (e *BaseEntity) GetID() string {
	return e.ID
}

// GetCreatedAt returns the creation timestamp
func (e *BaseEntity) GetCreatedAt() time.Time {
	return e.CreatedAt
}

// GetUpdatedAt returns the last update timestamp
func (e *BaseEntity) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

// Exchange represents a cryptocurrency exchange entity
type Exchange struct {
	BaseEntity
	Name             string `json:"name"`
	LicenseNumber    string `json:"license_number"`
	Status           string `json:"status"`
	Jurisdiction     string `json:"jurisdiction"`
	Website          string `json:"website,omitempty"`
	ContactEmail     string `json:"contact_email,omitempty"`
	ComplianceScore  int    `json:"compliance_score"`
	RiskLevel        string `json:"risk_level"`
	RegistrationDate string `json:"registration_date"`
	LastAudit        string `json:"last_audit,omitempty"`
	NextAudit        string `json:"next_audit,omitempty"`
}

// ExchangeStatus represents the possible statuses of an exchange
type ExchangeStatus string

const (
	ExchangeStatusActive     ExchangeStatus = "ACTIVE"
	ExchangeStatusSuspended  ExchangeStatus = "SUSPENDED"
	ExchangeStatusRevoked    ExchangeStatus = "REVOKED"
	ExchangeStatusPending    ExchangeStatus = "PENDING"
)

// Wallet represents a cryptocurrency wallet entity
type Wallet struct {
	BaseEntity
	Address       string `json:"address"`
	Label         string `json:"label"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	RiskScore     int    `json:"risk_score"`
	FirstSeen     string `json:"first_seen"`
	LastActivity  string `json:"last_activity"`
}

// WalletStatus represents the possible statuses of a wallet
type WalletStatus string

const (
	WalletStatusActive      WalletStatus = "ACTIVE"
	WalletStatusFrozen      WalletStatus = "FROZEN"
	WalletStatusBlacklisted WalletStatus = "BLACKLISTED"
)

// Miner represents a mining pool entity
type Miner struct {
	BaseEntity
	Name              string `json:"name"`
	LicenseNumber     string `json:"license_number"`
	Status            string `json:"status"`
	Jurisdiction      string `json:"jurisdiction"`
	HashRate          float64 `json:"hash_rate"`
	EnergyConsumption float64 `json:"energy_consumption"`
	EnergySource      string `json:"energy_source"`
	ComplianceStatus  string `json:"compliance_status"`
	RegistrationDate  string `json:"registration_date"`
	LastInspection    string `json:"last_inspection"`
}

// MinerStatus represents the possible statuses of a miner
type MinerStatus string

const (
	MinerStatusActive    MinerStatus = "ACTIVE"
	MinerStatusSuspended MinerStatus = "SUSPENDED"
	MinerStatusInactive  MinerStatus = "INACTIVE"
)

// Alert represents a security alert entity
type Alert struct {
	BaseEntity
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Severity    string       `json:"severity"`
	Status      string       `json:"status"`
	Category    string       `json:"category"`
	Source      string       `json:"source"`
	Evidence    []AlertEvidence `json:"evidence,omitempty"`
	AcknowledgedBy string    `json:"acknowledged_by,omitempty"`
	AcknowledgedAt string    `json:"acknowledged_at,omitempty"`
}

// AlertEvidence represents evidence supporting an alert
type AlertEvidence struct {
	Type      string `json:"type"`
	Value     interface{} `json:"value"`
	Threshold interface{} `json:"threshold,omitempty"`
}

// AlertSeverity represents the possible severity levels of an alert
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "CRITICAL"
	AlertSeverityWarning  AlertSeverity = "WARNING"
	AlertSeverityInfo     AlertSeverity = "INFO"
)

// AlertStatus represents the possible statuses of an alert
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "ACTIVE"
	AlertStatusAcknowledged AlertStatus = "ACKNOWLEDGED"
	AlertStatusResolved     AlertStatus = "RESOLVED"
	AlertStatusDismissed    AlertStatus = "DISMISSED"
}

// ComplianceReport represents a compliance report entity
type ComplianceReport struct {
	BaseEntity
	EntityType  string `json:"entity_type"`
	EntityID    string `json:"entity_id"`
	EntityName  string `json:"entity_name"`
	Period      string `json:"period"`
	Status      string `json:"status"`
	Score       int    `json:"score"`
	GeneratedAt string `json:"generated_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	BaseEntity
	Timestamp     string `json:"timestamp"`
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	Action        string `json:"action"`
	ResourceType  string `json:"resource_type"`
	Details       string `json:"details"`
	IPAddress     string `json:"ip_address"`
	Status        string `json:"status"`
}

// User represents a user entity
type User struct {
	BaseEntity
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	Role       string   `json:"role"`
	Status     string   `json:"status"`
	Password   string   `json:"-"`
	LastLogin  string   `json:"last_login,omitempty"`
	MFAEnabled bool     `json:"mfa_enabled"`
	Permissions []string `json:"permissions"`
}

// UserRole represents the possible roles for a user
type UserRole string

const (
	UserRoleAdmin    UserRole = "ADMIN"
	UserRoleRegulator UserRole = "REGULATOR"
	UserRoleOperator UserRole = "OPERATOR"
	UserRoleViewer   UserRole = "VIEWER"
)

// BlockchainStatus represents the status of a blockchain node
type BlockchainStatus struct {
	NodeType      string  `json:"node_type"`
	Connected     bool    `json:"connected"`
	BlockHeight   int64   `json:"block_height"`
	HeadersHeight int64   `json:"headers_height"`
	Peers         int     `json:"peers"`
	LatencyMS     int     `json:"latency_ms"`
	SyncProgress  float64 `json:"sync_progress"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	SystemStatus SystemStatus  `json:"system_status"`
	Metrics      Metrics       `json:"metrics"`
}

// SystemStatus represents the system status
type SystemStatus struct {
	Status       string `json:"status"`
	Uptime       float64 `json:"uptime"`
	LastHeartbeat string `json:"last_heartbeat"`
}

// Metrics represents system metrics
type Metrics struct {
	Exchanges    ExchangeMetrics  `json:"exchanges"`
	Transactions TransactionMetrics `json:"transactions"`
	Wallets      WalletMetrics    `json:"wallets"`
	Miners       MinerMetrics     `json:"miners"`
}

// ExchangeMetrics represents exchange-related metrics
type ExchangeMetrics struct {
	Total     int `json:"total"`
	Active    int `json:"active"`
	Suspended int `json:"suspended"`
	Revoked   int `json:"revoked"`
}

// TransactionMetrics represents transaction-related metrics
type TransactionMetrics struct {
	Total24h  int64   `json:"total_24h"`
	Volume24h float64 `json:"volume_24h"`
	Flagged24h int    `json:"flagged_24h"`
}

// WalletMetrics represents wallet-related metrics
type WalletMetrics struct {
	Total      int `json:"total"`
	Frozen     int `json:"frozen"`
	Blacklisted int `json:"blacklisted"`
}

// MinerMetrics represents miner-related metrics
type MinerMetrics struct {
	Total        int     `json:"total"`
	Online       int     `json:"online"`
	TotalHashrate float64 `json:"total_hashrate"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int64 `json:"total"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(items interface{}, page, pageSize int, total int64) *PaginatedResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return &PaginatedResponse{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}
