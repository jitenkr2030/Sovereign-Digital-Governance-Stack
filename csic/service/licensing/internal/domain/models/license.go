package models

import (
	"time"

	"github.com/google/uuid"
)

// LicenseStatus represents the operational status of a license
type LicenseStatus string

const (
	LicenseStatusActive          LicenseStatus = "active"
	LicenseStatusSuspended       LicenseStatus = "suspended"
	LicenseStatusRevoked         LicenseStatus = "revoked"
	LicenseStatusExpired         LicenseStatus = "expired"
	LicenseStatusPendingRenewal  LicenseStatus = "pending_renewal"
	LicenseStatusPendingApproval LicenseStatus = "pending_approval"
	LicenseStatusTerminated      LicenseStatus = "terminated"
)

// LicenseType represents the type of license
type LicenseType string

const (
	LicenseTypeOperational   LicenseType = "operational"
	LicenseTypeProvisional   LicenseType = "provisional"
	LicenseTypeRestricted    LicenseType = "restricted"
	LicenseTypeConditional   LicenseType = "conditional"
	LicenseTypeExperimental  LicenseType = "experimental"
)

// Jurisdiction represents regulatory jurisdictions
type Jurisdiction string

const (
	JurisdictionUSSEC    Jurisdiction = "US-SEC"
	JurisdictionUSFinCEN Jurisdiction = "US-FinCEN"
	JurisdictionUKFCA    Jurisdiction = "UK-FCA"
	JurisdictionEUESMA   Jurisdiction = "EU-ESMA"
	JurisdictionJPJFSA   Jurisdiction = "JP-FSA"
	JurisdictionSGSG     Jurisdiction = "SG-MAS"
	JurisdictionHKMA     Jurisdiction = "HK-MA"
	JurisdictionGlobal   Jurisdiction = "global"
)

// License represents an exchange operating license
type License struct {
	ID             string                 `json:"id" db:"id"`
	ExchangeID     string                 `json:"exchange_id" db:"exchange_id"`
	LicenseNumber  string                 `json:"license_number" db:"license_number"`
	Jurisdiction   Jurisdiction           `json:"jurisdiction" db:"jurisdiction"`
	Type           LicenseType            `json:"type" db:"type"`
	Status         LicenseStatus          `json:"status" db:"status"`
	IssuedDate     time.Time              `json:"issued_date" db:"issued_date"`
	ExpiryDate     time.Time              `json:"expiry_date" db:"expiry_date"`
	Conditions     map[string]interface{} `json:"conditions" db:"conditions"`
	Requirements   []LicenseRequirement   `json:"requirements" db:"-"`
	Documents      []DocumentReference    `json:"documents" db:"-"`
	Notes          string                 `json:"notes" db:"notes"`
	ApprovedBy     string                 `json:"approved_by" db:"approved_by"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// LicenseRequirement represents a compliance requirement attached to a license
type LicenseRequirement struct {
	ID             string    `json:"id" db:"id"`
	LicenseID      string    `json:"license_id" db:"license_id"`
	Requirement    string    `json:"requirement" db:"requirement"`
	Category       string    `json:"category" db:"category"`
	Frequency      string    `json:"frequency" db:"frequency"` // daily, weekly, monthly, quarterly, annually
	DueDate        time.Time `json:"due_date" db:"due_date"`
	Status         string    `json:"status" db:"status"` // pending, in_progress, submitted, approved, rejected
	LastSubmission time.Time `json:"last_submission" db:"last_submission"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// DocumentReference represents a reference to a stored document
type DocumentReference struct {
	ID          string    `json:"id" db:"id"`
	LicenseID   string    `json:"license_id" db:"license_id"`
	DocumentType string   `json:"document_type" db:"document_type"` // application, approval, audit_report, compliance_report
	FileName    string    `json:"file_name" db:"file_name"`
	FilePath    string    `json:"file_path" db:"file_path"`
	FileSize    int64     `json:"file_size" db:"file_size"`
	MimeType    string    `json:"mime_type" db:"mime_type"`
	UploadedBy  string    `json:"uploaded_by" db:"uploaded_by"`
	UploadedAt  time.Time `json:"uploaded_at" db:"uploaded_at"`
	VerifiedAt  *time.Time `json:"verified_at" db:"verified_at"`
	VerifiedBy  string    `json:"verified_by" db:"verified_by"`
}

// LicenseApplication represents a license application
type LicenseApplication struct {
	ID            string                 `json:"id" db:"id"`
	ExchangeID    string                 `json:"exchange_id" db:"exchange_id"`
	Jurisdiction  Jurisdiction           `json:"jurisdiction" db:"jurisdiction"`
	Type          LicenseType            `json:"type" db:"type"`
	Status        string                 `json:"status" db:"status"` // submitted, under_review, pending_info, approved, rejected
	SubmittedAt   time.Time              `json:"submitted_at" db:"submitted_at"`
	ReviewedAt    *time.Time             `json:"reviewed_at" db:"reviewed_at"`
	ReviewedBy    string                 `json:"reviewed_by" db:"reviewed_by"`
	Decision      string                 `json:"decision" db:"decision"`
	DecisionNotes string                 `json:"decision_notes" db:"decision_notes"`
	Documents     []DocumentReference    `json:"documents" db:"-"`
	Answers       map[string]interface{} `json:"answers" db:"answers"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// NewLicense creates a new license with generated ID
func NewLicense(exchangeID, licenseNumber string, jurisdiction Jurisdiction, licenseType LicenseType) *License {
	return &License{
		ID:            uuid.New().String(),
		ExchangeID:    exchangeID,
		LicenseNumber: licenseNumber,
		Jurisdiction:  jurisdiction,
		Type:          licenseType,
		Status:        LicenseStatusPendingApproval,
		Requirements:  []LicenseRequirement{},
		Documents:     []DocumentReference{},
		Conditions:    make(map[string]interface{}),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// IsValid checks if the license is currently valid
func (l *License) IsValid() bool {
	return l.Status == LicenseStatusActive &&
		time.Now().After(l.IssuedDate) &&
		time.Now().Before(l.ExpiryDate)
}

// IsExpiringSoon checks if license expires within given days
func (l *License) IsExpiringSoon(days int) bool {
	return time.Until(l.ExpiryDate).Hours()/24 <= float64(days)
}

// CanTrade checks if the license permits trading operations
func (l *License) CanTrade() bool {
	return l.IsValid() &&
		(l.Type == LicenseTypeOperational || l.Type == LicenseTypeConditional)
}

// TransitionStatus changes the license status with validation
func (l *License) TransitionStatus(newStatus LicenseStatus) error {
	// Define valid transitions
	validTransitions := map[LicenseStatus][]LicenseStatus{
		LicenseStatusPendingApproval: {LicenseStatusActive, LicenseStatusSuspended, LicenseStatusPendingRenewal},
		LicenseStatusActive:          {LicenseStatusSuspended, LicenseStatusRevoked, LicenseStatusPendingRenewal, LicenseStatusExpired},
		LicenseStatusSuspended:       {LicenseStatusActive, LicenseStatusRevoked, LicenseStatusTerminated},
		LicenseStatusPendingRenewal:  {LicenseStatusActive, LicenseStatusExpired},
		LicenseStatusExpired:         {LicenseStatusPendingRenewal, LicenseStatusRevoked},
	}

	allowed, ok := validTransitions[l.Status]
	if !ok {
		return ErrInvalidStatusTransition
	}

	for _, allowedStatus := range allowed {
		if newStatus == allowedStatus {
			l.Status = newStatus
			l.UpdatedAt = time.Now()
			return nil
		}
	}

	return ErrInvalidStatusTransition
}

// LicenseEvent represents an event in the license lifecycle
type LicenseEvent struct {
	ID        string    `json:"id" db:"id"`
	LicenseID string    `json:"license_id" db:"license_id"`
	EventType string    `json:"event_type" db:"event_type"` // issued, suspended, revoked, renewed, conditions_updated
	OldStatus string    `json:"old_status" db:"old_status"`
	NewStatus string    `json:"new_status" db:"new_status"`
	Details   string    `json:"details" db:"details"`
	ActorID   string    `json:"actor_id" db:"actor_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
