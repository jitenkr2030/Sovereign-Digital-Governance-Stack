package domain

import (
	"time"
)

// EntityID is a type alias for entity identifiers.
type EntityID string

// NewEntityID generates a new unique entity identifier.
func NewEntityID() EntityID {
	return EntityID(generateReportUUID())
}

// ReportStatus represents the status of a regulatory report.
type ReportStatus string

const (
	ReportStatusDraft     ReportStatus = "draft"
	ReportStatusPending   ReportStatus = "pending_review"
	ReportStatusApproved  ReportStatus = "approved"
	ReportStatusSubmitted ReportStatus = "submitted"
	ReportStatusRejected  ReportStatus = "rejected"
	ReportStatusClosed    ReportStatus = "closed"
)

// ReportType represents the type of regulatory report.
type ReportType string

const (
	ReportTypeSAR          ReportType = "sar"          // Suspicious Activity Report
	ReportTypeCTR          ReportType = "ctr"          // Currency Transaction Report
	ReportTypeFBAR         ReportType = "fbar"         // Foreign Bank Account Report
	ReportTypeDOJ          ReportType = "doj"          // Department of Justice Referral
	ReportTypeFinCEN       ReportType = "fincen"       // FinCEN Request
	ReportTypeInternalAlert ReportType = "internal_alert" // Internal compliance alert
)

// SAR represents a Suspicious Activity Report.
type SAR struct {
	ID                EntityID            `json:"id"`
	ReportNumber      string              `json:"report_number"`
	SubjectID         string              `json:"subject_id"`
	SubjectType       string              `json:"subject_type"` // individual, organization, account
	SubjectName       string              `json:"subject_name"`
	SuspiciousActivity string             `json:"suspicious_activity"`
	ActivityDate      time.Time           `json:"activity_date"`
	DollarAmount      float64             `json:"dollar_amount"`
	Currency          string              `json:"currency"`
	TransactionCount  int                 `json:"transaction_count"`
	Narrative         string              `json:"narrative"`
	RiskIndicators    []string            `json:"risk_indicators"`
	FilingInstitution string              `json:"filing_institution"`
	ReporterID        string              `json:"reporter_id"`
	ReviewerID        string              `json:"reviewer_id,omitempty"`
	Status            ReportStatus        `json:"status"`
	SupportingDocs    []SupportingDocument `json:"supporting_docs,omitempty"`
	FieldOffice       string              `json:"field_office"`
	OriginatingAgency string              `json:"originating_agency"`
	SubmittedAt       *time.Time          `json:"submitted_at,omitempty"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

// CTR represents a Currency Transaction Report.
type CTR struct {
	ID                   EntityID            `json:"id"`
	ReportNumber         string              `json:"report_number"`
	PersonName           string              `json:"person_name"`
	PersonIDType         string              `json:"person_id_type"` // SSN, Passport, etc.
	PersonIDNumber       string              `json:"person_id_number"`
	PersonAddress        Address             `json:"person_address"`
	AccountNumber        string              `json:"account_number"`
	TransactionType      string              `json:"transaction_type"` // deposit, withdrawal, transfer
	CashInAmount         float64             `json:"cash_in_amount"`
	CashOutAmount        float64             `json:"cash_out_amount"`
	TotalAmount          float64             `json:"total_amount"`
	Currency             string              `json:"currency"`
	TransactionDate      time.Time           `json:"transaction_date"`
	MethodReceived       string              `json:"method_received"` // in_person, mail, wire
	InstitutionID        string              `json:"institution_id"`
	BranchID             string              `json:"branch_id"`
	TellerID             string              `json:"teller_id"`
	ReviewerID           string              `json:"reviewer_id,omitempty"`
	Status               ReportStatus        `json:"status"`
	SupportingDocs       []SupportingDocument `json:"supporting_docs,omitempty"`
	SubmittedAt          *time.Time          `json:"submitted_at,omitempty"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}

// Address represents a physical address.
type Address struct {
	Street1    string `json:"street1"`
	Street2    string `json:"street2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// SupportingDocument represents a supporting document for a report.
type SupportingDocument struct {
	ID          EntityID    `json:"id"`
	DocumentType string     `json:"document_type"`
	FileName    string      `json:"file_name"`
	FileSize    int64       `json:"file_size"`
	MimeType    string      `json:"mime_type"`
	FilePath    string      `json:"file_path"`
	Description string      `json:"description,omitempty"`
	UploadedAt  time.Time   `json:"uploaded_at"`
}

// ComplianceRule represents a compliance rule.
type ComplianceRule struct {
	ID            EntityID            `json:"id"`
	RuleCode      string              `json:"rule_code"`
	RuleName      string              `json:"rule_name"`
	Category      string              `json:"category"` // aml, kyc, sanctions, transaction
	Regulation    string              `json:"regulation"` // bsa, aml, etc.
	Description   string              `json:"description"`
	Threshold     *RuleThreshold      `json:"threshold,omitempty"`
	Logic         string              `json:"logic"` // JSON logic for rule evaluation
	RiskScore     int                 `json:"risk_score"` // 1-10
	Severity      string              `json:"severity"` // low, medium, high, critical
	IsActive      bool                `json:"is_active"`
	EffectiveDate time.Time           `json:"effective_date"`
	ExpiryDate    *time.Time          `json:"expiry_date,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// RuleThreshold represents a threshold condition for a rule.
type RuleThreshold struct {
	Field       string      `json:"field"`
	Operator    string      `json:"operator"` // gt, lt, gte, lte, eq, ne, in, between
	Value       interface{} `json:"value"`
	SecondaryValue interface{} `json:"secondary_value,omitempty"`
	TimeWindow  *time.Duration `json:"time_window,omitempty"`
}

// Alert represents a compliance alert.
type Alert struct {
	ID            EntityID          `json:"id"`
	AlertNumber   string            `json:"alert_number"`
	AlertType     string            `json:"alert_type"` // sar_filing, ctr_filing, pattern_detected
	Severity      string            `json:"severity"` // low, medium, high, critical
	SubjectID     string            `json:"subject_id"`
	SubjectType   string            `json:"subject_type"`
	SubjectName   string            `json:"subject_name"`
	TransactionIDs []string         `json:"transaction_ids"`
	Description   string            `json:"description"`
	TriggeredRules []string         `json:"triggered_rules"`
	RiskScore     int               `json:"risk_score"`
	AssignedTo    string            `json:"assigned_to,omitempty"`
	Status        string            `json:"status"` // open, investigating, resolved, escalated
	Resolution    string            `json:"resolution,omitempty"`
	ResolvedAt    *time.Time        `json:"resolved_at,omitempty"`
	ResolvedBy    string            `json:"resolved_by,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// ComplianceCheck represents a compliance check result.
type ComplianceCheck struct {
	ID            EntityID    `json:"id"`
	SubjectID     string      `json:"subject_id"`
	SubjectType   string      `json:"subject_type"`
	CheckType     string      `json:"check_type"` // aml_screening, sanctions_screening, pep_screening
	CheckResult   string      `json:"check_result"` // pass, fail, review_required, pending
	RiskLevel     string      `json:"risk_level"` // low, medium, high
	MatchScore    float64     `json:"match_score"`
	MatchedEntity string      `json:"matched_entity,omitempty"`
	MatchDetails  string      `json:"match_details,omitempty"`
	ScreeningList string      `json:"screening_list,omitempty"`
	CheckedAt     time.Time   `json:"checked_at"`
	ExpiresAt     *time.Time  `json:"expires_at,omitempty"`
}

// ComplianceReport represents a generated compliance report.
type ComplianceReport struct {
	ID              EntityID          `json:"id"`
	ReportType      ReportType        `json:"report_type"`
	ReportName      string            `json:"report_name"`
	PeriodStart     time.Time         `json:"period_start"`
	PeriodEnd       time.Time         `json:"period_end"`
	Summary         ReportSummary     `json:"summary"`
	Details         interface{}       `json:"details"`
	GeneratedBy     string            `json:"generated_by"`
	GeneratedAt     time.Time         `json:"generated_at"`
	Status          ReportStatus      `json:"status"`
	FilePath        string            `json:"file_path,omitempty"`
}

// ReportSummary represents a summary of report data.
type ReportSummary struct {
	TotalTransactions     int64   `json:"total_transactions"`
	TotalVolume           float64 `json:"total_volume"`
	Currency              string  `json:"currency"`
	FlaggedTransactions   int64   `json:"flagged_transactions"`
	SuspiciousActivity    int64   `json:"suspicious_activity"`
	AlertsGenerated       int64   `json:"alerts_generated"`
	AlertsResolved        int64   `json:"alerts_resolved"`
	AverageRiskScore      float64 `json:"average_risk_score"`
	HighRiskCount         int64   `json:"high_risk_count"`
	MediumRiskCount       int64   `json:"medium_risk_count"`
	LowRiskCount          int64   `json:"low_risk_count"`
}

// FilingRecord represents a record of a regulatory filing.
type FilingRecord struct {
	ID            EntityID    `json:"id"`
	ReportID      EntityID    `json:"report_id"`
	ReportType    ReportType  `json:"report_type"`
	ReportNumber  string      `json:"report_number"`
	FilingMethod  string      `json:"filing_method"` // electronic, mail, secure_portal
	FilingAgency  string      `json:"filing_agency"` // FinCEN, SEC, etc.
	FilingStatus  string      `json:"filing_status"` // submitted, accepted, rejected
	Confirmation  string      `json:"confirmation_number,omitempty"`
	SubmittedAt   time.Time   `json:"submitted_at"`
	AcceptedAt    *time.Time  `json:"accepted_at,omitempty"`
	ResponseDue   *time.Time  `json:"response_due,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
}

// generateReportUUID generates a UUID-like identifier.
func generateReportUUID() string {
	return time.Now().UTC().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random alphanumeric string.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
