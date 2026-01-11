package models

import (
	"time"

	"github.com/google/uuid"
)

// ComplianceScore represents the overall compliance score for an entity
type ComplianceScore struct {
	OverallScore     float64            `json:"overall_score"`
	RiskLevel        string             `json:"risk_level"`
	Breakdown        ScoreBreakdown     `json:"breakdown"`
	LastCalculatedAt time.Time          `json:"last_calculated_at"`
	Trend            ScoreTrend         `json:"trend"`
}

// ScoreBreakdown represents the breakdown of compliance score components
type ScoreBreakdown struct {
	LicenseCompliance    float64 `json:"license_compliance"`    // 30%
	ViolationRecord      float64 `json:"violation_record"`      // 25%
	ReportingCompliance  float64 `json:"reporting_compliance"`  // 20%
	CapitalAdequacy      float64 `json:"capital_adequacy"`      // 15%
	OperationalReadiness float64 `json:"operational_readiness"` // 10%
}

// ScoreTrend represents the trend of compliance score
type ScoreTrend struct {
	Direction      string    `json:"direction"` // improving, stable, declining
	ChangePercent  float64   `json:"change_percent"`
	Period         string    `json:"period"` // weekly, monthly, quarterly
	PreviousScore  float64   `json:"previous_score"`
}

// DashboardMetrics represents aggregated metrics for dashboard display
type DashboardMetrics struct {
	// License metrics
	TotalLicenses       int64 `json:"total_licenses"`
	ActiveLicenses      int64 `json:"active_licenses"`
	SuspendedLicenses   int64 `json:"suspended_licenses"`
	ExpiringSoon        int64 `json:"expiring_soon"`
	PendingRenewals     int64 `json:"pending_renewals"`

	// Violation metrics
	TotalViolations     int64 `json:"total_violations"`
	OpenViolations      int64 `json:"open_violations"`
	CriticalViolations  int64 `json:"critical_violations"`
	HighViolations      int64 `json:"high_violations"`
	ResolvedThisMonth   int64 `json:"resolved_this_month"`

	// Report metrics
	TotalReports        int64 `json:"total_reports"`
	PendingReports      int64 `json:"pending_reports"`
	SubmittedReports    int64 `json:"submitted_reports"`
	OverdueReports      int64 `json:"overdue_reports"`

	// Compliance score
	AverageComplianceScore float64 `json:"average_compliance_score"`
	RiskDistribution       map[string]int64 `json:"risk_distribution"`

	// Timeline data
	RecentActivity      []ActivityItem      `json:"recent_activity"`
	ViolationTrends     []ViolationTrend    `json:"violation_trends"`
	LicenseExpiryTimeline []LicenseExpiry   `json:"license_expiry_timeline"`

	// Generated at
	LastUpdated time.Time `json:"last_updated"`
}

// ActivityItem represents a recent activity item for the dashboard
type ActivityItem struct {
	ID          string    `json:"id"`
	ActivityType string   `json:"activity_type"` // license_issued, violation_detected, report_submitted, etc.
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EntityID    string    `json:"entity_id"`
	EntityType  string    `json:"entity_type"` // license, violation, report
	Severity    string    `json:"severity,omitempty"`
	ActorID     string    `json:"actor_id"`
	ActorName   string    `json:"actor_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// LicenseExpiry represents an upcoming license expiry
type LicenseExpiry struct {
	LicenseID     string    `json:"license_id"`
	LicenseNumber string    `json:"license_number"`
	ExchangeID    string    `json:"exchange_id"`
	ExchangeName  string    `json:"exchange_name"`
	Jurisdiction  string    `json:"jurisdiction"`
	ExpiryDate    time.Time `json:"expiry_date"`
	DaysRemaining int       `json:"days_remaining"`
	Status        string    `json:"status"`
}

// KPIMetric represents a key performance indicator metric
type KPIMetric struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Target      float64   `json:"target"`
	Threshold   float64   `json:"threshold"`
	Trend       string    `json:"trend"` // up, down, stable
	Status      string    `json:"status"` // on_track, at_risk, below_target
	Period      string    `json:"period"` // daily, weekly, monthly
	LastUpdated time.Time `json:"last_updated"`
}

// ComplianceCheck represents a compliance check or audit item
type ComplianceCheck struct {
	ID             string    `json:"id" db:"id"`
	LicenseID      string    `json:"license_id" db:"license_id"`
	CheckType      string    `json:"check_type" db:"check_type"` // aml, kyc, capital, technical, governance
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description" db:"description"`
	Category       string    `json:"category" db:"category"`
	Frequency      string    `json:"frequency" db:"frequency"` // daily, weekly, monthly, quarterly, annually
	LastCheckAt    *time.Time `json:"last_check_at" db:"last_check_at"`
	NextCheckAt    *time.Time `json:"next_check_at" db:"next_check_at"`
	Status         string    `json:"status" db:"status"` // compliant, non_compliant, pending, not_applicable
	Findings       string    `json:"findings" db:"findings"`
	Remediation    string    `json:"remediation" db:"remediation"`
	CheckedBy      string    `json:"checked_by" db:"checked_by"`
	ApprovedBy     string    `json:"approved_by" db:"approved_by"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID          string    `json:"id" db:"id"`
	EntityType  string    `json:"entity_type" db:"entity_type"` // license, violation, report, compliance_check
	EntityID    string    `json:"entity_id" db:"entity_id"`
	Action      string    `json:"action" db:"action"` // created, updated, deleted, status_changed, submitted, approved
	ActorID     string    `json:"actor_id" db:"actor_id"`
	ActorName   string    `json:"actor_name" db:"actor_name"`
	ActorType   string    `json:"actor_type" db:"actor_type"` // user, system, api
	OldValue    string    `json:"old_value" db:"old_value"`
	NewValue    string    `json:"new_value" db:"new_value"`
	Reason      string    `json:"reason" db:"reason"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	UserAgent   string    `json:"user_agent" db:"user_agent"`
	Metadata    string    `json:"metadata" db:"metadata"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Regulator represents a regulatory authority
type Regulator struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"`
	Jurisdiction Jurisdiction `json:"jurisdiction" db:"jurisdiction"`
	Website     string    `json:"website" db:"website"`
	ContactEmail string   `json:"contact_email" db:"contact_email"`
	Address     string    `json:"address" db:"address"`
	ReportingURL string   `json:"reporting_url" db:"reporting_url"`
	Requirements string   `json:"requirements" db:"requirements"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ComplianceAlert represents a proactive compliance alert
type ComplianceAlert struct {
	ID          string    `json:"id" db:"id"`
	AlertType   string    `json:"alert_type" db:"alert_type"` // upcoming_expiry, potential_violation, deadline_approaching
	Severity    string    `json:"severity" db:"severity"` // low, medium, high, critical
	Title       string    `json:"title" db:"title"`
	Message     string    `json:"message" db:"message"`
	EntityType  string    `json:"entity_type" db:"entity_type"`
	EntityID    string    `json:"entity_id" db:"entity_id"`
	ActionRequired string `json:"action_required" db:"action_required"`
	Deadline    *time.Time `json:"deadline" db:"deadline"`
	IsRead      bool      `json:"is_read" db:"is_read"`
	IsResolved  bool      `json:"is_resolved" db:"is_resolved"`
	ResolvedAt  *time.Time `json:"resolved_at" db:"resolved_at"`
	ResolvedBy  string    `json:"resolved_by" db:"resolved_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// NewDashboardMetrics creates a new dashboard metrics instance
func NewDashboardMetrics() *DashboardMetrics {
	return &DashboardMetrics{
		RiskDistribution:   make(map[string]int64),
		RecentActivity:     []ActivityItem{},
		ViolationTrends:    []ViolationTrend{},
		LicenseExpiryTimeline: []LicenseExpiry{},
		LastUpdated:        time.Now(),
	}
}

// NewActivityItem creates a new activity item
func NewActivityItem(activityType, title, description, entityID, entityType, actorID, actorName string) *ActivityItem {
	return &ActivityItem{
		ID:            uuid.New().String(),
		ActivityType:  activityType,
		Title:         title,
		Description:   description,
		EntityID:      entityID,
		EntityType:    entityType,
		ActorID:       actorID,
		ActorName:     actorName,
		CreatedAt:     time.Now(),
	}
}
