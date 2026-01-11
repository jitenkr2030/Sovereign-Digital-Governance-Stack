// Compliance Management Module - Violation and Penalty Models
// Violation tracking, penalties, and enforcement actions

package domain

import (
	"time"
)

// ViolationType represents the type of violation
type ViolationType string

const (
	ViolationTypeAdministrative    ViolationType = "ADMINISTRATIVE"
	ViolationTypeProcedural        ViolationType = "PROCEDURAL"
	ViolationTypeFinancial         ViolationType = "FINANCIAL"
	ViolationTypeSecurity          ViolationType = "SECURITY"
	ViolationTypeReporting         ViolationType = "REPORTING"
	ViolationTypeAML               ViolationType = "AML"
	ViolationTypeKYC               ViolationType = "KYC"
	ViolationTypeDataProtection    ViolationType = "DATA_PROTECTION"
	ViolationTypeConsumerProtection ViolationType = "CONSUMER_PROTECTION"
	ViolationTypeMarketIntegrity   ViolationType = "MARKET_INTEGRITY"
	ViolationTypeLicensing         ViolationType = "LICENSING"
	ViolationTypeOperational       ViolationType = "OPERATIONAL"
)

// ViolationSeverity represents the severity level of a violation
type ViolationSeverity string

const (
	ViolationSeverityInfo     ViolationSeverity = "INFO"
	ViolationSeverityMinor    ViolationSeverity = "MINOR"
	ViolationSeverityModerate ViolationSeverity = "MODERATE"
	ViolationSeverityMajor    ViolationSeverity = "MAJOR"
	ViolationSeverityCritical ViolationSeverity = "CRITICAL"
)

// ViolationStatus represents the status of a violation
type ViolationStatus string

const (
	ViolationStatusDetected   ViolationStatus = "DETECTED"
	ViolationStatusReported   ViolationStatus = "REPORTED"
	ViolationStatusInvestigating ViolationStatus = "INVESTIGATING"
	ViolationStatusConfirmed  ViolationStatus = "CONFIRMED"
	ViolationStatusPendingAction ViolationStatus = "PENDING_ACTION"
	ViolationStatusRemediation ViolationStatus = "REMEDIATION"
	ViolationStatusResolved   ViolationStatus = "RESOLVED"
	ViolationStatusClosed     ViolationStatus = "CLOSED"
	ViolationStatusAppealed   ViolationStatus = "APPEALED"
	ViolationStatusDismissed  ViolationStatus = "DISMISSED"
)

// PenaltyType represents the type of penalty
type PenaltyType string

const (
	PenaltyTypeWarning           PenaltyType = "WARNING"
	PenaltyTypeReprimand         PenaltyType = "REPRIMAND"
	PenaltyTypeFine              PenaltyType = "FINE"
	PenaltyTypeConfiscation      PenaltyType = "CONFISCATION"
	PenaltyTypeSuspension        PenaltyType = "SUSPENSION"
	PenaltyTypeRestriction       PenaltyType = "RESTRICTION"
	PenaltyTypeRevocation        PenaltyType = "REVOCATION"
	PenaltyTypeProhibition       PenaltyType = "PROHIBITION"
	PenaltyTypeCeaseDesist       PenaltyType = "CEASE_DESIST"
	PenaltyTypeCorrectiveAction  PenaltyType = "CORRECTIVE_ACTION"
)

// PenaltyStatus represents the status of a penalty
type PenaltyStatus string

const (
	PenaltyStatusIssued    PenaltyStatus = "ISSUED"
	PenaltyStatusPending   PenaltyStatus = "PENDING"
	PenaltyStatusPaid      PenaltyStatus = "PAID"
	PenaltyStatusPartial   PenaltyStatus = "PARTIAL"
	PenaltyStatusWaived    PenaltyStatus = "WAIVED"
	PenaltyStatusAppealed  PenaltyStatus = "APPEALED"
	PenaltyStatusOverdue   PenaltyStatus = "OVERDUE"
	PenaltyStatusEnforced  PenaltyStatus = "ENFORCED"
)

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	ID                string             `json:"id" db:"id"`
	EntityID          string             `json:"entity_id" db:"entity_entity_id"`
	EntityName        string             `json:"entity_name" db:"entity_name"`
	LicenseID         string             `json:"license_id" db:"license_id"`
	ViolationNumber   string             `json:"violation_number" db:"violation_number"`
	Type              ViolationType      `json:"type" db:"type"`
	Severity          ViolationSeverity  `json:"severity" db:"severity"`
	Status            ViolationStatus    `json:"status" db:"status"`
	Title             string             `json:"title" db:"title"`
	Description       string             `json:"description" db:"description"`
	IncidentDate      time.Time          `json:"incident_date" db:"incident_date"`
	DetectionDate     time.Time          `json:"detection_date" db:"detection_date"`
	ReportedDate      *time.Time         `json:"reported_date,omitempty" db:"reported_date"`
	InvestigatedDate  *time.Time         `json:"investigated_date,omitempty" db:"investigated_date"`
	ConfirmedDate     *time.Time         `json:"confirmed_date,omitempty" db:"confirmed_date"`
	ResolvedDate      *time.Time         `json:"resolved_date,omitempty" db:"resolved_date"`
	ClosedDate        *time.Time         `json:"closed_date,omitempty" db:"closed_date"`
	DetectionSource   string             `json:"detection_source" db:"detection_source"`
	RegulatoryRef     string             `json:"regulatory_ref" db:"regulatory_ref"`
	AffectedParties   []string           `json:"affected_parties"`
	FinancialImpact   *FinancialImpact   `json:"financial_impact,omitempty"`
	RootCause         string             `json:"root_cause,omitempty"`
	CorrectiveAction  string             `json:"corrective_action,omitempty"`
	PreventiveAction  string             `json:"preventive_action,omitempty"`
	AssignedInvestigator string          `json:"assigned_investigator" db:"assigned_investigator"`
	ReviewerID        string             `json:"reviewer_id" db:"reviewer_id"`
	Appealed          bool               `json:"appealed"`
	AppealReason      string             `json:"appeal_reason,omitempty"`
	PenaltyIDs        []string           `json:"penalty_ids"`
	EvidenceIDs       []string           `json:"evidence_ids"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" db:"updated_at"`
}

// FinancialImpact represents the financial impact of a violation
type FinancialImpact struct {
	EstimatedLoss      float64 `json:"estimated_loss"`
	ActualLoss         float64 `json:"actual_loss"`
	CustomerImpact     int     `json:"customer_impact"`
	MarketImpact       string  `json:"market_impact"`
	ReputationImpact   string  `json:"reputation_impact"`
}

// Penalty represents a penalty for a violation
type Penalty struct {
	ID              string         `json:"id" db:"id"`
	ViolationID     string         `json:"violation_id" db:"violation_id"`
	PenaltyNumber   string         `json:"penalty_number" db:"penalty_number"`
	Type            PenaltyType    `json:"type" db:"type"`
	Status          PenaltyStatus  `json:"status" db:"status"`
	Amount          float64        `json:"amount" db:"amount"`
	Currency        string         `json:"currency" db:"currency"`
	Description     string         `json:"description" db:"description"`
	LegalBasis      string         `json:"legal_basis" db:"legal_basis"`
	IssuedDate      time.Time      `json:"issued_date" db:"issued_date"`
	DueDate         time.Time      `json:"due_date" db:"due_date"`
	PaidDate        *time.Time     `json:"paid_date,omitempty" db:"paid_date"`
	PaymentRef      string         `json:"payment_ref,omitempty" db:"payment_ref"`
	AppealDeadline  *time.Time     `json:"appeal_deadline,omitempty" db:"appeal_deadline"`
	AppealStatus    string         `json:"appeal_status,omitempty"`
	Notes           string         `json:"notes,omitempty"`
	EnforcementDate *time.Time     `json:"enforcement_date,omitempty" db:"enforcement_date"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

// EnforcementAction represents an enforcement action
type EnforcementAction struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"` // SUSPENSION, RESTRICTION, REVOCATION
	Description     string    `json:"description"`
	EffectiveDate   time.Time `json:"effective_date"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	Reason          string    `json:"reason"`
	LegalBasis      string    `json:"legal_basis"`
	Status          string    `json:"status"`
	ViolationID     string    `json:"violation_id"`
	PenaltyID       string    `json:"penalty_id"`
}

// Validate validates the violation data
func (v *ComplianceViolation) Validate() error {
	if v.EntityID == "" {
		return ErrValidationError("entity ID is required")
	}
	if v.Type == "" {
		return ErrValidationError("violation type is required")
	}
	if v.Title == "" {
		return ErrValidationError("title is required")
	}
	return nil
}

// GenerateViolationNumber generates a violation number
func (v *ComplianceViolation) GenerateViolationNumber() string {
	year := time.Now().Year()
	sequence := generateSequenceNumber()
	return fmt.Sprintf("V-%d-%s", year, sequence)
}

// Status transition methods for violations
func (v *ComplianceViolation) Confirm() error {
	if v.Status != ViolationStatusInvestigating {
		return ErrInvalidStateTransition("violation", v.Status, ViolationStatusConfirmed)
	}
	v.Status = ViolationStatusConfirmed
	now := time.Now()
	v.ConfirmedDate = &now
	return nil
}

func (v *ComplianceViolation) Resolve() error {
	if v.Status != ViolationStatusRemediation {
		return ErrInvalidStateTransition("violation", v.Status, ViolationStatusResolved)
	}
	v.Status = ViolationStatusResolved
	now := time.Now()
	v.ResolvedDate = &now
	return nil
}

func (v *ComplianceViolation) Close() error {
	if v.Status != ViolationStatusResolved && v.Status != ViolationStatusAppealed &&
	   v.Status != ViolationStatusDismissed {
		return ErrInvalidStateTransition("violation", v.Status, ViolationStatusClosed)
	}
	v.Status = ViolationStatusClosed
	now := time.Now()
	v.ClosedDate = &now
	return nil
}

func (v *ComplianceViolation) Dismiss(reason string) error {
	v.Status = ViolationStatusDismissed
	v.CorrectiveAction = reason
	return nil
}

func (v *ComplianceViolation) AddPenalty(penaltyID string) {
	v.PenaltyIDs = append(v.PenaltyIDs, penaltyID)
}

// Penalty methods
func (p *Penalty) IsOverdue() bool {
	return p.Status != PenaltyStatusPaid && p.Status != PenaltyStatusWaived &&
	   p.Status != PenaltyStatusAppealed && time.Now().After(p.DueDate)
}

func (p *Penalty) GeneratePenaltyNumber() string {
	year := time.Now().Year()
	sequence := generateSequenceNumber()
	return fmt.Sprintf("P-%d-%s", year, sequence)
}

func (p *Penalty) MarkPaid(paymentRef string) error {
	if p.Status != PenaltyStatusPending && p.Status != PenaltyStatusPartial {
		return ErrInvalidStateTransition("penalty", p.Status, PenaltyStatusPaid)
	}
	p.Status = PenaltyStatusPaid
	p.PaymentRef = paymentRef
	now := time.Now()
	p.PaidDate = &now
	return nil
}

func (p *Penalty) MarkPartial(amount float64) error {
	p.Status = PenaltyStatusPartial
	p.Amount -= amount
	if p.Amount <= 0 {
		return p.MarkPaid(paymentRef)
	}
	return nil
}

// EscalationMatrix defines severity to action mapping
type EscalationMatrix struct {
	Info     []PenaltyType
	Minor    []PenaltyType
	Moderate []PenaltyType
	Major    []PenaltyType
	Critical []PenaltyType
}

// GetDefaultEscalationMatrix returns the default escalation matrix
func GetDefaultEscalationMatrix() EscalationMatrix {
	return EscalationMatrix{
		Info:     []PenaltyType{PenaltyTypeWarning},
		Minor:    []PenaltyType{PenaltyTypeWarning, PenaltyTypeReprimand},
		Moderate: []PenaltyType{PenaltyTypeFine, PenaltyTypeReprimand},
		Major:    []PenaltyType{PenaltyTypeFine, PenaltyTypeSuspension, PenaltyTypeRestriction},
		Critical: []PenaltyType{PenaltyTypeSuspension, PenaltyTypeRevocation, PenaltyTypeProhibition, PenaltyTypeCeaseDesist},
	}
}

// ViolationReport represents a violation summary report
type ViolationReport struct {
	EntityID       string            `json:"entity_id"`
	EntityName     string            `json:"entity_name"`
	PeriodStart    time.Time         `json:"period_start"`
	PeriodEnd      time.Time         `json:"period_end"`
	TotalCount     int               `json:"total_count"`
	ByType         map[ViolationType]int `json:"by_type"`
	BySeverity     map[ViolationSeverity]int `json:"by_severity"`
	ByStatus       map[ViolationStatus]int `json:"by_status"`
	TotalPenalties float64           `json:"total_penalties"`
	ResolvedCount  int               `json:"resolved_count"`
	OpenCount      int               `json:"open_count"`
}
