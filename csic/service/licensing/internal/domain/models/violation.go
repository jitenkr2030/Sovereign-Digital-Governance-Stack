package models

import (
	"time"

	"github.com/google/uuid"
)

// ViolationSeverity represents the severity level of a violation
type ViolationSeverity string

const (
	SeverityLow      ViolationSeverity = "low"
	SeverityMedium   ViolationSeverity = "medium"
	SeverityHigh     ViolationSeverity = "high"
	SeverityCritical ViolationSeverity = "critical"
)

// ViolationType represents categories of compliance violations
type ViolationType string

const (
	ViolationTypeReportingDelay       ViolationType = "reporting_delay"
	ViolationTypeCapitalAdequacy      ViolationType = "capital_adequacy"
	ViolationTypeAMLFlag              ViolationType = "aml_flag"
	ViolationTypeKYCFailure           ViolationType = "kyc_failure"
	ViolationTypeTradeViolation       ViolationType = "trade_violation"
	ViolationTypeDataBreach           ViolationType = "data_breach"
	ViolationTypeMarketManipulation   ViolationType = "market_manipulation"
	ViolationTypeSanctionsViolation   ViolationType = "sanctions_violation"
	ViolationTypePrivacyViolation     ViolationType = "privacy_violation"
	ViolationTypeTechnicalFailure     ViolationType = "technical_failure"
	ViolationTypeGovernanceBreach     ViolationType = "governance_breach"
	ViolationTypeClientAssetProtection ViolationType = "client_asset_protection"
)

// ViolationStatus represents the current status of a violation
type ViolationStatus string

const (
	ViolationStatusOpen         ViolationStatus = "open"
	ViolationStatusInvestigating ViolationStatus = "investigating"
	ViolationStatusPendingReview ViolationStatus = "pending_review"
	ViolationStatusResolved     ViolationStatus = "resolved"
	ViolationStatusWaived       ViolationStatus = "waived"
	ViolationStatusEscalated    ViolationStatus = "escalated"
)

// ComplianceViolation represents a compliance violation record
type ComplianceViolation struct {
	ID             string             `json:"id" db:"id"`
	LicenseID      string             `json:"license_id" db:"license_id"`
	ExchangeID     string             `json:"exchange_id" db:"exchange_id"`
	Severity       ViolationSeverity  `json:"severity" db:"severity"`
	Type           ViolationType      `json:"type" db:"type"`
	Title          string             `json:"title" db:"title"`
	Description    string             `json:"description" db:"description"`
	Details        map[string]interface{} `json:"details" db:"details"`
	Status         ViolationStatus    `json:"status" db:"status"`
	DetectedAt     time.Time          `json:"detected_at" db:"detected_at"`
	ResolvedAt     *time.Time         `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy     string             `json:"resolved_by" db:"resolved_by"`
	Resolution     string             `json:"resolution" db:"resolution"`
	RootCause      string             `json:"root_cause" db:"root_cause"`
	CorrectiveAction string           `json:"corrective_action" db:"corrective_action"`
	PreventiveAction string           `json:"preventive_action" db:"preventive_action"`
	AssignedTo     string             `json:"assigned_to" db:"assigned_to"`
	AssignedTeam   string             `json:"assigned_team" db:"assigned_team"`
	EscalatedTo    string             `json:"escalated_to" db:"escalated_to"`
	ExternalReport bool               `json:"external_report" db:"external_report"`
	RegulatorNotified bool            `json:"regulator_notified" db:"regulator_notified"`
	RelatedAlerts  []string           `json:"related_alerts" db:"-"`
	Evidence       []EvidenceItem     `json:"evidence" db:"-"`
	CreatedAt      time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" db:"updated_at"`
}

// EvidenceItem represents evidence attached to a violation
type EvidenceItem struct {
	ID          string    `json:"id" db:"id"`
	ViolationID string    `json:"violation_id" db:"violation_id"`
	EvidenceType string   `json:"evidence_type" db:"evidence_type"` // screenshot, document, log, transaction
	FileName    string    `json:"file_name" db:"file_name"`
	FilePath    string    `json:"file_path" db:"file_path"`
	Description string    `json:"description" db:"description"`
	UploadedBy  string    `json:"uploaded_by" db:"uploaded_by"`
	UploadedAt  time.Time `json:"uploaded_at" db:"created_at"`
}

// ViolationEvent represents an event in the violation lifecycle
type ViolationEvent struct {
	ID           string    `json:"id" db:"id"`
	ViolationID  string    `json:"violation_id" db:"violation_id"`
	EventType    string    `json:"event_type" db:"event_type"` // created, status_changed, assigned, escalated, resolved
	OldStatus    string    `json:"old_status" db:"old_status"`
	NewStatus    string    `json:"new_status" db:"new_status"`
	ChangedBy    string    `json:"changed_by" db:"changed_by"`
	ChangeReason string    `json:"change_reason" db:"change_reason"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// NewViolation creates a new violation record
func NewViolation(licenseID, exchangeID string, severity ViolationSeverity, violationType ViolationType, title, description string) *ComplianceViolation {
	return &ComplianceViolation{
		ID:          uuid.New().String(),
		LicenseID:   licenseID,
		ExchangeID:  exchangeID,
		Severity:    severity,
		Type:        violationType,
		Title:       title,
		Description: description,
		Status:      ViolationStatusOpen,
		DetectedAt:  time.Now(),
		RelatedAlerts: []string{},
		Evidence:    []EvidenceItem{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Escalate escalates the violation to a higher authority
func (v *ComplianceViolation) Escalate(escalateTo string) {
	v.Status = ViolationStatusEscalated
	v.EscalatedTo = escalateTo
	v.UpdatedAt = time.Now()
}

// Assign assigns the violation to an investigator
func (v *ComplianceViolation) Assign(assignee, team string) {
	v.Status = ViolationStatusInvestigating
	v.AssignedTo = assignee
	v.AssignedTeam = team
	v.UpdatedAt = time.Now()
}

// Resolve marks the violation as resolved
func (v *ComplianceViolation) Resolve(resolvedBy, resolution, rootCause, correctiveAction, preventiveAction string) {
	v.Status = ViolationStatusResolved
	now := time.Now()
	v.ResolvedAt = &now
	v.ResolvedBy = resolvedBy
	v.Resolution = resolution
	v.RootCause = rootCause
	v.CorrectiveAction = correctiveAction
	v.PreventiveAction = preventiveAction
	v.UpdatedAt = now
}

// ViolationStats represents aggregated violation statistics
type ViolationStats struct {
	TotalViolations      int64                      `json:"total_violations"`
	OpenViolations       int64                      `json:"open_violations"`
	ResolvedViolations   int64                      `json:"resolved_violations"`
	BySeverity           map[ViolationSeverity]int64 `json:"by_severity"`
	ByType               map[ViolationType]int64    `json:"by_type"`
	AverageResolutionTime float64                   `json:"average_resolution_time_hours"`
	CriticalCount        int64                      `json:"critical_count"`
	HighCount            int64                      `json:"high_count"`
	OverdueCount         int64                      `json:"overdue_count"`
}

// ViolationTrend represents violation trends over time
type ViolationTrend struct {
	Date            time.Time `json:"date"`
	NewViolations   int64     `json:"new_violations"`
	ResolvedViolations int64   `json:"resolved_violations"`
	OpenViolations  int64     `json:"open_violations"`
	CriticalCount   int64     `json:"critical_count"`
}
