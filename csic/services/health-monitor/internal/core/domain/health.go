package domain

import (
	"time"
)

// ServiceStatus represents the health status of a service
type ServiceStatus struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"` // "healthy", "degraded", "unhealthy", "unknown"
	PreviousStatus  string    `json:"previous_status,omitempty"`
	LastHeartbeat   time.Time `json:"last_heartbeat"`
	LastCheck       time.Time `json:"last_check"`
	Uptime          float64   `json:"uptime_percent"`
	ResponseTime    float64   `json:"response_time_ms"`
	ErrorRate       float64   `json:"error_rate_percent"`
	CPUUsage        float64   `json:"cpu_percent"`
	MemoryUsage     float64   `json:"memory_percent"`
	DiskUsage       float64   `json:"disk_percent"`
	Message         string    `json:"message,omitempty"`
	Metadata        Metadata  `json:"metadata,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// HealthStatus constants
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusUnknown   = "unknown"
)

// Heartbeat represents a heartbeat message from a service
type Heartbeat struct {
	ServiceName   string            `json:"service_name"`
	InstanceID    string            `json:"instance_id"`
	Timestamp     time.Time         `json:"timestamp"`
	Status        string            `json:"status"`
	Metrics       map[string]float64 `json:"metrics"`
	Version       string            `json:"version"`
	Hostname      string            `json:"hostname"`
	Uptime        float64           `json:"uptime_seconds"`
	Metadata      Metadata          `json:"metadata,omitempty"`
}

// HealthCheck represents a health check result
type HealthCheck struct {
	ServiceName string    `json:"service_name"`
	Endpoint    string    `json:"endpoint"`
	Status      string    `json:"status"`
	ResponseTime float64  `json:"response_time_ms"`
	StatusCode  int       `json:"status_code"`
	Message     string    `json:"message,omitempty"`
	CheckedAt   time.Time `json:"checked_at"`
}

// AlertRule represents a rule for generating alerts
type AlertRule struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	ServiceName    string    `json:"service_name"`
	Condition      string    `json:"condition"`      // e.g., "cpu > 80", "memory > 85"
	Threshold      float64   `json:"threshold"`
	Duration       int       `json:"duration_seconds"` // How long condition must persist
	Severity       string    `json:"severity"`         // "info", "warning", "critical"
	Enabled        bool      `json:"enabled"`
	Cooldown       int       `json:"cooldown_seconds"`
	NotificationChannels []string `json:"notification_channels"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AlertSeverity constants
const (
	AlertSeverityInfo     = "info"
	AlertSeverityWarning  = "warning"
	AlertSeverityCritical = "critical"
)

// Alert represents a health alert
type Alert struct {
	ID           string    `json:"id"`
	RuleID       string    `json:"rule_id"`
	ServiceName  string    `json:"service_name"`
	Severity     string    `json:"severity"`
	Condition    string    `json:"condition"`
	CurrentValue float64   `json:"current_value"`
	Threshold    float64   `json:"threshold"`
	Status       string    `json:"status"` // "firing", "resolved", "pending"
	Message      string    `json:"message"`
	FiredAt      time.Time `json:"fired_at"`
	ResolvedAt   time.Time `json:"resolved_at,omitempty"`
	Metadata     Metadata  `json:"metadata,omitempty"`
}

// AlertStatus constants
const (
	AlertStatusFiring  = "firing"
	AlertStatusPending = "pending"
	AlertStatusResolved = "resolved"
)

// Outage represents a service outage
type Outage struct {
	ID            string    `json:"id"`
	ServiceName   string    `json:"service_name"`
	InstanceID    string    `json:"instance_id,omitempty"`
	Severity      string    `json:"severity"`
	Status        string    `json:"status"` // "detected", "investigating", "identified", "monitoring", "resolved"
	StartedAt     time.Time `json:"started_at"`
	DetectedAt    time.Time `json:"detected_at"`
	ResolvedAt    time.Time `json:"resolved_at,omitempty"`
	Duration      int       `json:"duration_seconds"`
	Impact        string    `json:"impact"`
	Description   string    `json:"description"`
	RootCause     string    `json:"root_cause,omitempty"`
	AffectedUsers int       `json:"affected_users,omitempty"`
	Metadata      Metadata  `json:"metadata,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
)

// OutageStatus constants
const (
	OutageStatusDetected      = "detected"
	OutageStatusInvestigating = "investigating"
	OutageStatusIdentified    = "identified"
	OutageStatusMonitoring    = "monitoring"
	OutageStatusResolved      = "resolved"
)

// HealthSummary represents a summary of all service health statuses
type HealthSummary struct {
	TotalServices   int                   `json:"total_services"`
	HealthyServices int                   `json:"healthy_services"`
	DegradedServices int                  `json:"degraded_services"`
	UnhealthyServices int                 `json:"unhealthy_services"`
	UnknownServices int                   `json:"unknown_services"`
	LastUpdated     time.Time             `json:"last_updated"`
	Services        []*ServiceStatus      `json:"services"`
	RecentAlerts    []*Alert              `json:"recent_alerts"`
	RecentOutages   []*Outage             `json:"recent_outages"`
}

// Metadata represents flexible key-value metadata
type Metadata map[string]interface{}

// ServiceHealth represents the health overview of a single service
type ServiceHealth struct {
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	Uptime        float64   `json:"uptime_percent"`
	Incidents     int       `json:"incident_count_last_24h"`
	LastIncident  time.Time `json:"last_incident_at,omitempty"`
}

// HealthCheckRequest represents a request to check service health
type HealthCheckRequest struct {
	ServiceName string `json:"service_name"`
	Endpoint    string `json:"endpoint"`
	Timeout     int    `json:"timeout_ms"`
}

// HealthCheckResponse represents the response from a health check
type HealthCheckResponse struct {
	ServiceName string    `json:"service_name"`
	Status      string    `json:"status"`
	ResponseTime float64  `json:"response_time_ms"`
	Message     string    `json:"message,omitempty"`
	CheckedAt   time.Time `json:"checked_at"`
}

// RegisterServiceRequest represents a request to register a new service
type RegisterServiceRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Endpoints   []string `json:"endpoints"`
	Tags        []string `json:"tags"`
	Metadata    Metadata `json:"metadata"`
}

// RegisterServiceResponse represents the response from registering a service
type RegisterServiceResponse struct {
	ServiceID string    `json:"service_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
