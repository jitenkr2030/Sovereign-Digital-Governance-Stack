package domain

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusNew          AlertStatus = "new"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusInvestigating AlertStatus = "investigating"
	AlertStatusEscalated    AlertStatus = "escalated"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusDismissed    AlertStatus = "dismissed"
)

// AlertType represents the type of financial crime alert
type AlertType string

const (
	AlertTypeVelocity          AlertType = "velocity"
	AlertTypeStructuring       AlertType = "structuring"
	AlertTypeSmurfing          AlertType = "smurfing"
	AlertTypeMoneyLaundering   AlertType = "money_laundering"
	AlertTypeTerrorFinancing   AlertType = "terror_financing"
	AlertTypeSanctionsViolation AlertType = "sanctions_violation"
	AlertTypeFraud             AlertType = "fraud"
	AlertTypeMarketManipulation AlertType = "market_manipulation"
	AlertTypeIdentityTheft     AlertType = "identity_theft"
	AlertTypeSuspiciousActivity AlertType = "suspicious_activity"
	AlertTypeGeographicRisk    AlertType = "geographic_risk"
	AlertTypeBehavioralAnomaly AlertType = "behavioral_anomaly"
	AlertTypePEP               AlertType = "pep_match"
	AlertTypeBlacklist         AlertType = "blacklist_hit"
)

// Alert represents a financial crime alert
type Alert struct {
	ID              string         `json:"id" db:"id"`
	CaseID          *string        `json:"case_id,omitempty" db:"case_id"`
	Type            AlertType      `json:"type" db:"type"`
	Severity        AlertSeverity  `json:"severity" db:"severity"`
	Status          AlertStatus    `json:"status" db:"status"`
	Score           int            `json:"score" db:"score"`
	RuleTriggered   string         `json:"rule_triggered" db:"rule_triggered"`
	RuleDetails     *RuleDetails   `json:"rule_details,omitempty" db:"rule_details"`
	EntityID        string         `json:"entity_id" db:"entity_id"`
	EntityType      string         `json:"entity_type" db:"entity_type"`
	TransactionID   *string        `json:"transaction_id,omitempty" db:"transaction_id"`
	Description     string         `json:"description" db:"description"`
	ContextData     *AlertContext  `json:"context_data,omitempty" db:"context_data"`
	RelatedAlerts   []string       `json:"related_alerts,omitempty" db:"related_alerts"`
	AssignedTo      *string        `json:"assigned_to,omitempty" db:"assigned_to"`
	AcknowledgedAt  *time.Time     `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	ResolvedAt      *time.Time     `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy      *string        `json:"resolved_by,omitempty" db:"resolved_by"`
	Resolution      *string        `json:"resolution,omitempty" db:"resolution"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty" db:"expires_at"`
}

// RuleDetails contains details about the detection rule that triggered the alert
type RuleDetails struct {
	RuleID        string                 `json:"rule_id"`
	RuleName      string                 `json:"rule_name"`
	RuleCategory  string                 `json:"rule_category"`
	Threshold     map[string]interface{} `json:"threshold,omitempty"`
	ActualValue   map[string]interface{} `json:"actual_value,omitempty"`
	Deviation     float64                `json:"deviation"`
}

// Value implements driver.Valuer for RuleDetails
func (r RuleDetails) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan implements sql.Scanner for RuleDetails
func (r *RuleDetails) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan RuleDetails")
	}
	return json.Unmarshal(bytes, r)
}

// AlertContext contains contextual information about the alert
type AlertContext struct {
	TransactionHash   *string            `json:"transaction_hash,omitempty"`
	Amount            *float64           `json:"amount,omitempty"`
	Currency          *string            `json:"currency,omitempty"`
	SenderID          *string            `json:"sender_id,omitempty"`
	ReceiverID        *string            `json:"receiver_id,omitempty"`
	IPAddress         *string            `json:"ip_address,omitempty"`
	GeographicData    *GeographicContext `json:"geographic_data,omitempty"`
	PreviousActivity  *ActivitySummary   `json:"previous_activity,omitempty"`
	RelatedParties    []string           `json:"related_parties,omitempty"`
	Tags              []string           `json:"tags,omitempty"`
}

// Value implements driver.Valuer for AlertContext
func (a AlertContext) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements sql.Scanner for AlertContext
func (a *AlertContext) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan AlertContext")
	}
	return json.Unmarshal(bytes, a)
}

// GeographicContext contains geographic risk information
type GeographicContext struct {
	SourceCountry     *string `json:"source_country,omitempty"`
	DestinationCountry *string `json:"destination_country,omitempty"`
	SourceIP          *string `json:"source_ip,omitempty"`
	IPCountry         *string `json:"ip_country,omitempty"`
	RiskLevel         string  `json:"risk_level"`
	HighRiskJurisdiction bool  `json:"high_risk_jurisdiction"`
	SanctionedCountry bool    `json:"sanctioned_country"`
}

// ActivitySummary contains a summary of previous activity
type ActivitySummary struct {
	TransactionCount24h  int     `json:"transaction_count_24h"`
	TransactionVolume24h float64 `json:"transaction_volume_24h"`
	TransactionCount7d   int     `json:"transaction_count_7d"`
	TransactionVolume7d  float64 `json:"transaction_volume_7d"`
	AverageRiskScore     float64 `json:"average_risk_score"`
	LastActivityAt       *time.Time `json:"last_activity_at,omitempty"`
}

// CaseStatus represents the status of an investigation case
type CaseStatus string

const (
	CaseStatusNew          CaseStatus = "new"
	CaseStatusAssigned     CaseStatus = "assigned"
	CaseStatusInvestigating CaseStatus = "investigating"
	CaseStatusSARFiled     CaseStatus = "sar_filed"
	CaseStatusEscalated    CaseStatus = "escalated"
	CaseStatusDismissed    CaseStatus = "dismissed"
	CaseStatusClosed       CaseStatus = "closed"
)

// CasePriority represents the priority level of a case
type CasePriority string

const (
	CasePriorityLow      CasePriority = "low"
	CasePriorityMedium   CasePriority = "medium"
	CasePriorityHigh     CasePriority = "high"
	CasePriorityUrgent   CasePriority = "urgent"
)

// Case represents an investigation case
type Case struct {
	ID              string        `json:"id" db:"id"`
	CaseNumber      string        `json:"case_number" db:"case_number"`
	Status          CaseStatus    `json:"status" db:"status"`
	Priority        CasePriority  `json:"priority" db:"priority"`
	Title           string        `json:"title" db:"title"`
	Description     string        `json:"description" db:"description"`
	AlertIDs        []string      `json:"alert_ids,omitempty" db:"alert_ids"`
	SubjectID       string        `json:"subject_id" db:"subject_id"`
	SubjectType     string        `json:"subject_type" db:"subject_type"`
	AssigneeID      *string       `json:"assignee_id,omitempty" db:"assignee_id"`
	TeamID          *string       `json:"team_id,omitempty" db:"team_id"`
	RiskScore       int           `json:"risk_score" db:"risk_score"`
	TotalAmount     float64       `json:"total_amount" db:"total_amount"`
	Currency        string        `json:"currency" db:"currency"`
	TransactionIDs  []string      `json:"transaction_ids,omitempty" db:"transaction_ids"`
	Evidence        []Evidence    `json:"evidence,omitempty" db:"evidence"`
	Tags            []string      `json:"tags,omitempty" db:"tags"`
	RelatedCases    []string      `json:"related_cases,omitempty" db:"related_cases"`
	SARID           *string       `json:"sar_id,omitempty" db:"sar_id"`
	RegulatoryBody  *string       `json:"regulatory_body,omitempty" db:"regulatory_body"`
	Deadline        *time.Time    `json:"deadline,omitempty" db:"deadline"`
	EscalatedAt     *time.Time    `json:"escalated_at,omitempty" db:"escalated_at"`
	ClosedAt        *time.Time    `json:"closed_at,omitempty" db:"closed_at"`
	ClosedReason    *string       `json:"closed_reason,omitempty" db:"closed_reason"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
	CreatedBy       string        `json:"created_by" db:"created_by"`
	Version         int           `json:"version" db:"version"`
}

// Value implements driver.Valuer for Evidence
func (e Evidence) Value() (driver.Value, error) {
	return json.Marshal(e)
}

// Scan implements sql.Scanner for Evidence
func (e *Evidence) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Evidence")
	}
	return json.Unmarshal(bytes, e)
}

// Evidence represents evidence attached to a case
type Evidence struct {
	ID            string    `json:"id" db:"id"`
	Type          string    `json:"type" db:"type"`
	Name          string    `json:"name" db:"name"`
	Description   *string   `json:"description,omitempty" db:"description"`
	FilePath      *string   `json:"file_path,omitempty" db:"file_path"`
	FileHash      *string   `json:"file_hash,omitempty" db:"file_hash"`
	Metadata      *EvidenceMetadata `json:"metadata,omitempty" db:"metadata"`
	AddedAt       time.Time `json:"added_at" db:"added_at"`
	AddedBy       string    `json:"added_by" db:"added_by"`
}

// EvidenceMetadata contains metadata about evidence
type EvidenceMetadata struct {
	FileSize      int64     `json:"file_size"`
	MimeType      string    `json:"mime_type"`
	Encrypted     bool      `json:"encrypted"`
	RetentionDate *time.Time `json:"retention_date,omitempty"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID              string            `json:"id" db:"id"`
	TransactionHash string            `json:"transaction_hash" db:"transaction_hash"`
	Type            string            `json:"type" db:"type"`
	Asset           string            `json:"asset" db:"asset"`
	Amount          float64           `json:"amount" db:"amount"`
	Currency        string            `json:"currency" db:"currency"`
	SenderID        string            `json:"sender_id" db:"sender_id"`
	SenderWallet    *string           `json:"sender_wallet,omitempty" db:"sender_wallet"`
	ReceiverID      string            `json:"receiver_id" db:"receiver_id"`
	ReceiverWallet  *string           `json:"receiver_wallet,omitempty" db:"receiver_wallet"`
	Fee             float64           `json:"fee" db:"fee"`
	FeeCurrency     *string           `json:"fee_currency,omitempty" db:"fee_currency"`
	Status          string            `json:"status" db:"status"`
	RiskScore       int               `json:"risk_score" db:"risk_score"`
	Flagged         bool              `json:"flagged" db:"flagged"`
	FlagReason      *string           `json:"flag_reason,omitempty" db:"flag_reason"`
	BlockHeight     *int64            `json:"block_height,omitempty" db:"block_height"`
	BlockHash       *string           `json:"block_hash,omitempty" db:"block_hash"`
	Timestamp       time.Time         `json:"timestamp" db:"timestamp"`
	ProcessedAt     *time.Time        `json:"processed_at,omitempty" db:"processed_at"`
	Metadata        *TransactionMetadata `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
}

// TransactionMetadata contains additional transaction information
type TransactionMetadata struct {
	IPAddress       *string            `json:"ip_address,omitempty"`
	GeoLocation     *GeoLocation       `json:"geo_location,omitempty"`
	DeviceInfo      *DeviceInfo        `json:"device_info,omitempty"`
	Memo            *string            `json:"memo,omitempty"`
	Reference       *string            `json:"reference,omitempty"`
	RelatedTXIDs    []string           `json:"related_txids,omitempty"`
	Tags            []string           `json:"tags,omitempty"`
}

// GeoLocation contains geographic coordinates
type GeoLocation struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Country    string  `json:"country"`
	City       *string `json:"city,omitempty"`
	Region     *string `json:"region,omitempty"`
}

// Value implements driver.Valuer for TransactionMetadata
func (t TransactionMetadata) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan implements sql.Scanner for TransactionMetadata
func (t *TransactionMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan TransactionMetadata")
	}
	return json.Unmarshal(bytes, t)
}

// Entity represents an entity being monitored (individual or organization)
type Entity struct {
	ID               string         `json:"id" db:"id"`
	ExternalID       *string        `json:"external_id,omitempty" db:"external_id"`
	Type             string         `json:"type" db:"type"`
	Name             string         `json:"name" db:"name"`
	Email            *string        `json:"email,omitempty" db:"email"`
	Phone            *string        `json:"phone,omitempty" db:"phone"`
	WalletAddresses  []string       `json:"wallet_addresses,omitempty" db:"wallet_addresses"`
	RiskScore        int            `json:"risk_score" db:"risk_score"`
	RiskLevel        string         `json:"risk_level" db:"risk_level"`
	KYCCheckStatus   string         `json:"kyc_check_status" db:"kyc_check_status"`
	KYCLevel         int            `json:"kyc_level" db:"kyc_level"`
	AccountStatus    string         `json:"account_status" db:"account_status"`
	WatchlistStatus  string         `json:"watchlist_status" db:"watchlist_status"`
	WatchlistReason  *string        `json:"watchlist_reason,omitempty" db:"watchlist_reason"`
	PEPStatus        string         `json:"pep_status" db:"pep_status"`
	SanctionsStatus  string         `json:"sanctions_status" db:"sanctions_status"`
	BlacklistStatus  string         `json:"blacklist_status" db:"blacklist_status"`
	HighRiskCountry  bool           `json:"high_risk_country" db:"high_risk_country"`
	LastRiskAssessment *time.Time   `json:"last_risk_assessment,omitempty" db:"last_risk_assessment"`
	AccountCreatedAt time.Time      `json:"account_created_at" db:"account_created_at"`
	LastActivityAt   *time.Time     `json:"last_activity_at,omitempty" db:"last_activity_at"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
	Version          int            `json:"version" db:"version"`
}

// Value implements driver.Valuer for Entity
func (e Entity) Value() (driver.Value, error) {
	return json.Marshal(e)
}

// Scan implements sql.Scanner for Entity
func (e *Entity) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Entity")
	}
	return json.Unmarshal(bytes, e)
}

// SARStatus represents the status of a Suspicious Activity Report
type SARStatus string

const (
	SARStatusDraft     SARStatus = "draft"
	SARStatusReview    SARStatus = "review"
	SARStatusValidated SARStatus = "validated"
	SARStatusSubmitted SARStatus = "submitted"
	SARStatusRejected  SARStatus = "rejected"
	SARStatusAccepted  SARStatus = "accepted"
	SARStatusClosed    SARStatus = "closed"
)

// SAR represents a Suspicious Activity Report
type SAR struct {
	ID               string         `json:"id" db:"id"`
	SARNumber        string         `json:"sar_number" db:"sar_number"`
	CaseID           string         `json:"case_id" db:"case_id"`
	Status           SARStatus      `json:"status" db:"status"`
	RegulatoryBody   string         `json:"regulatory_body" db:"regulatory_body"`
	ReportType       string         `json:"report_type" db:"report_type"`
	SubjectID        string         `json:"subject_id" db:"subject_id"`
	SubjectName      string         `json:"subject_name" db:"subject_name"`
	SubjectType      string         `json:"subject_type" db:"subject_type"`
	TotalAmount      float64        `json:"total_amount" db:"total_amount"`
	Currency         string         `json:"currency" db:"currency"`
	Narrative        string         `json:"narrative" db:"narrative"`
	SuspiciousActivity string       `json:"suspicious_activity" db:"suspicious_activity"`
	TransactionDetails []TransactionSummary `json:"transaction_details,omitempty" db:"transaction_details"`
	Timeline         []TimelineEvent `json:"timeline,omitempty" db:"timeline"`
	SupportingDocs   []string       `json:"supporting_docs,omitempty" db:"supporting_docs"`
	FilingInstitution string        `json:"filing_institution" db:"filing_institution"`
	FilingUserID     string         `json:"filing_user_id" db:"filing_user_id"`
	ReviewerID       *string        `json:"reviewer_id,omitempty" db:"reviewer_id"`
	SubmittedAt      *time.Time     `json:"submitted_at,omitempty" db:"submitted_at"`
	RegulatorRef     *string        `json:"regulator_ref,omitempty" db:"regulator_ref"`
	RejectionReason  *string        `json:"rejection_reason,omitempty" db:"rejection_reason"`
	Deadline         *time.Time     `json:"deadline,omitempty" db:"deadline"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
	Version          int            `json:"version" db:"version"`
}

// TransactionSummary contains a summary of a transaction for SAR
type TransactionSummary struct {
	TransactionID   string    `json:"transaction_id"`
	Date            time.Time `json:"date"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	TransactionType string    `json:"transaction_type"`
	Description     string    `json:"description"`
}

// Value implements driver.Valuer for TransactionSummary
func (t TransactionSummary) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan implements sql.Scanner for TransactionSummary
func (t *TransactionSummary) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan TransactionSummary")
	}
	return json.Unmarshal(bytes, t)
}

// TimelineEvent represents an event in the investigation timeline
type TimelineEvent struct {
	ID          string    `json:"id" db:"id"`
	CaseID      string    `json:"case_id" db:"case_id"`
	EventType   string    `json:"event_type" db:"event_type"`
	Description string    `json:"description" db:"description"`
	PerformedBy string    `json:"performed_by" db:"performed_by"`
	EntityID    *string   `json:"entity_id,omitempty" db:"entity_id"`
	Metadata    *json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
}

// SanctionList represents an entry in a sanctions or watchlist
type SanctionList struct {
	ID              string         `json:"id" db:"id"`
	ListName        string         `json:"list_name" db:"list_name"`
	ListType        string         `json:"list_type" db:"list_type"`
	EntityType      string         `json:"entity_type" db:"entity_type"`
	Name            string         `json:"name" db:"name"`
	AlternativeNames []string      `json:"alternative_names,omitempty" db:"alternative_names"`
	DateOfBirth     *time.Time     `json:"date_of_birth,omitempty" db:"date_of_birth"`
	Nationality     *string        `json:"nationality,omitempty" db:"nationality"`
	PassportNumber  *string        `json:"passport_number,omitempty" db:"passport_number"`
	WalletAddresses []string       `json:"wallet_addresses,omitempty" db:"wallet_addresses"`
	Identifiers     []string       `json:"identifiers,omitempty" db:"identifiers"`
	Remarks         *string        `json:"remarks,omitempty" db:"remarks"`
	Program         *string        `json:"program,omitempty" db:"program"`
	EffectiveDate   *time.Time     `json:"effective_date,omitempty" db:"effective_date"`
	ExpirationDate  *time.Time     `json:"expiration_date,omitempty" db:"expiration_date"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

// WatchlistEntry represents an entity on the internal watchlist
type WatchlistEntry struct {
	ID              string         `json:"id" db:"id"`
	EntityID        string         `json:"entity_id" db:"entity_id"`
	EntityType      string         `json:"entity_type" db:"entity_type"`
	Reason          string         `json:"reason" db:"reason"`
	RiskLevel       string         `json:"risk_level" db:"risk_level"`
	AddedBy         string         `json:"added_by" db:"added_by"`
	ExpiryDate      *time.Time     `json:"expiry_date,omitempty" db:"expiry_date"`
	Notes           *string        `json:"notes,omitempty" db:"notes"`
	Tags            []string       `json:"tags,omitempty" db:"tags"`
	RelatedEntries  []string       `json:"related_entries,omitempty" db:"related_entries"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

// RiskProfile represents the risk profile of an entity
type RiskProfile struct {
	EntityID           string           `json:"entity_id" db:"entity_id"`
	OverallScore       int              `json:"overall_score" db:"overall_score"`
	RiskLevel          string           `json:"risk_level" db:"risk_level"`
	Components         RiskComponents   `json:"components" db:"components"`
	HistoricalScores   []HistoricalScore `json:"historical_scores,omitempty" db:"historical_scores"`
	LastAssessmentAt   time.Time        `json:"last_assessment_at" db:"last_assessment_at"`
	NextAssessmentAt   *time.Time       `json:"next_assessment_at,omitempty" db:"next_assessment_at"`
	CreatedAt          time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at" db:"updated_at"`
}

// RiskComponents contains the individual risk components
type RiskComponents struct {
	TransactionRisk     float64 `json:"transaction_risk"`
	VelocityRisk        float64 `json:"velocity_risk"`
	GeographicRisk      float64 `json:"geographic_risk"`
	EntityRisk          float64 `json:"entity_risk"`
	BehavioralRisk      float64 `json:"behavioral_risk"`
	SanctionsRisk       float64 `json:"sanctions_risk"`
	PepRisk             float64 `json:"pep_risk"`
	BlacklistRisk       float64 `json:"blacklist_risk"`
}

// Value implements driver.Valuer for RiskComponents
func (r RiskComponents) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan implements sql.Scanner for RiskComponents
func (r *RiskComponents) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan RiskComponents")
	}
	return json.Unmarshal(bytes, r)
}

// HistoricalScore represents a historical risk score
type HistoricalScore struct {
	Score      int       `json:"score" db:"score"`
	RiskLevel  string    `json:"risk_level" db:"risk_level"`
	RecordedAt time.Time `json:"recorded_at" db:"recorded_at"`
}

// Value implements driver.Valuer for HistoricalScore
func (h HistoricalScore) Value() (driver.Value, error) {
	return json.Marshal(h)
}

// Scan implements sql.Scanner for HistoricalScore
func (h *HistoricalScore) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan HistoricalScore")
	}
	return json.Unmarshal(bytes, h)
}

// Investigation represents an ongoing investigation
type Investigation struct {
	ID              string         `json:"id" db:"id"`
	CaseID          string         `json:"case_id" db:"case_id"`
	Type            string         `json:"type" db:"type"`
	Status          string         `json:"status" db:"status"`
	LeadInvestigator string        `json:"lead_investigator" db:"lead_investigator"`
	TeamMembers     []string       `json:"team_members,omitempty" db:"team_members"`
	SubjectID       string         `json:"subject_id" db:"subject_id"`
	SubjectType     string         `json:"subject_type" db:"subject_type"`
	Scope           string         `json:"scope" db:"scope"`
	Objectives      []string       `json:"objectives,omitempty" db:"objectives"`
	Methodology     *string        `json:"methodology,omitempty" db:"methodology"`
	Timeline        []TimelineEvent `json:"timeline,omitempty" db:"timeline"`
	Notes           []InvestigationNote `json:"notes,omitempty" db:"notes"`
	Findings        *json.RawMessage `json:"findings,omitempty" db:"findings"`
	Conclusions     *string        `json:"conclusions,omitempty" db:"conclusions"`
	Recommendations *string        `json:"recommendations,omitempty" db:"recommendations"`
	StartedAt       time.Time      `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time     `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

// InvestigationNote represents a note in an investigation
type InvestigationNote struct {
	ID            string    `json:"id" db:"id"`
	InvestigationID string  `json:"investigation_id" db:"investigation_id"`
	NoteType      string    `json:"note_type" db:"note_type"`
	Content       string    `json:"content" db:"content"`
	AuthorID      string    `json:"author_id" db:"author_id"`
	AuthorName    string    `json:"author_name" db:"author_name"`
	Confidential  bool      `json:"confidential" db:"confidential"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// NetworkGraph represents a network analysis result
type NetworkGraph struct {
	CentralEntityID  string            `json:"central_entity_id"`
	Nodes            []NetworkNode     `json:"nodes"`
	Edges            []NetworkEdge     `json:"edges"`
	Patterns         []NetworkPattern  `json:"patterns"`
	RiskScore        int               `json:"risk_score"`
	AnalysisDepth    int               `json:"analysis_depth"`
	AnalyzedAt       time.Time         `json:"analyzed_at"`
}

// NetworkNode represents a node in the network graph
type NetworkNode struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Label       string  `json:"label"`
	RiskScore   int     `json:"risk_score"`
	WalletCount int     `json:"wallet_count"`
	TransactionCount int `json:"transaction_count"`
	TotalVolume float64 `json:"total_volume"`
}

// NetworkEdge represents a connection between nodes
type NetworkEdge struct {
	SourceID      string  `json:"source_id"`
	TargetID      string  `json:"target_id"`
	TransactionCount int  `json:"transaction_count"`
	TotalVolume   float64 `json:"total_volume"`
	FirstSeenAt   time.Time `json:"first_seen_at"`
	LastSeenAt    time.Time `json:"last_seen_at"`
	Relationship  string  `json:"relationship"`
}

// NetworkPattern represents a detected network pattern
type NetworkPattern struct {
	Type          string  `json:"type"`
	Description   string  `json:"description"`
	Severity      string  `json:"severity"`
	EntitiesInvolved []string `json:"entities_involved"`
	Confidence    float64 `json:"confidence"`
}

// VelocityData represents velocity tracking data
type VelocityData struct {
	EntityID       string    `json:"entity_id" db:"entity_id"`
	MetricType     string    `json:"metric_type" db:"metric_type"`
	WindowStart    time.Time `json:"window_start" db:"window_start"`
	WindowEnd      time.Time `json:"window_end" db:"window_end"`
	TransactionCount int     `json:"transaction_count" db:"transaction_count"`
	TotalAmount    float64   `json:"total_amount" db:"total_amount"`
	UniqueCounterparties int `json:"unique_counterparties" db:"unique_counterparties"`
	AlertTriggered bool      `json:"alert_triggered" db:"alert_triggered"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID            string          `json:"id" db:"id"`
	Timestamp     time.Time       `json:"timestamp" db:"timestamp"`
	Action        string          `json:"action" db:"action"`
	ActorID       string          `json:"actor_id" db:"actor_id"`
	ActorType     string          `json:"actor_type" db:"actor_type"`
	EntityType    string          `json:"entity_type" db:"entity_type"`
	EntityID      string          `json:"entity_id" db:"entity_id"`
	OldValue      *string         `json:"old_value,omitempty" db:"old_value"`
	NewValue      *string         `json:"new_value,omitempty" db:"new_value"`
	IPAddress     *string         `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent     *string         `json:"user_agent,omitempty" db:"user_agent"`
	Metadata      *json.RawMessage `json:"metadata,omitempty" db:"metadata"`
}
