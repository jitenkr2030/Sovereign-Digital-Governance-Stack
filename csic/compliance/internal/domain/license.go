// Compliance Management Module - License Models
// License lifecycle management and state machine

package domain

import (
	"time"
)

// LicenseType represents the type of license
type LicenseType string

const (
	LicenseTypeExchange       LicenseType = "EXCHANGE_LICENSE"
	LicenseTypeCustodial      LicenseType = "CUSTODIAL_LICENSE"
	LicenseTypeWallet         LicenseType = "WALLET_LICENSE"
	LicenseTypeMining         LicenseType = "MINING_LICENSE"
	LicenseTypeOTC            LicenseType = "OTC_LICENSE"
	LicenseTypeICO            LicenseType = "ICO_LICENSE"
	LicenseTypeATM            LicenseType = "ATM_LICENSE"
)

// LicenseStatus represents the status of a license
type LicenseStatus string

const (
	LicenseStatusDraft       LicenseStatus = "DRAFT"
	LicenseStatusSubmitted   LicenseStatus = "SUBMITTED"
	LicenseStatusUnderReview LicenseStatus = "UNDER_REVIEW"
	LicenseStatusPendingConditions LicenseStatus = "PENDING_CONDITIONS"
	LicenseStatusApproved    LicenseStatus = "APPROVED"
	LicenseStatusActive      LicenseStatus = "ACTIVE"
	LicenseStatusExpired     LicenseStatus = "EXPIRED"
	LicenseStatusSuspended   LicenseStatus = "SUSPENDED"
	LicenseStatusRevoked     LicenseStatus = "REVOKED"
	LicenseStatusSurrendered LicenseStatus = "SURRENDERED"
)

// License represents a regulatory license
type License struct {
	ID               string            `json:"id" db:"id"`
	LicenseNumber    string            `json:"license_number" db:"license_number"`
	EntityID         string            `json:"entity_id" db:"entity_id"`
	EntityName       string            `json:"entity_name" db:"entity_name"`
	Type             LicenseType       `json:"type" db:"type"`
	Status           LicenseStatus     `json:"status" db:"status"`
	Jurisdiction     string            `json:"jurisdiction" db:"jurisdiction"`
	IssuedAt         time.Time         `json:"issued_at" db:"issued_at"`
	EffectiveDate    time.Time         `json:"effective_date" db:"effective_date"`
	ExpiresAt        time.Time         `json:"expires_at" db:"expires_at"`
	RenewalDueDate   *time.Time        `json:"renewal_due_date,omitempty" db:"renewal_due_date"`
	ApprovalDate     *time.Time        `json:"approval_date,omitempty" db:"approval_date"`
	ApprovalOfficer  string            `json:"approval_officer" db:"approval_officer"`
	Conditions       []LicenseCondition `json:"conditions"`
	Scope            LicenseScope      `json:"scope"`
	Fee              LicenseFee        `json:"fee"`
	PreviousLicense  string            `json:"previous_license,omitempty"`
	LastAuditDate    *time.Time        `json:"last_audit_date,omitempty"`
	CreatedAt        time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at" db:"updated_at"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// LicenseCondition represents a license condition or restriction
type LicenseCondition struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // REQUIRED, RESTRICTED, PROHIBITED
	Category    string    `json:"category"`
	Description string    `json:"description"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Status      string    `json:"status"` // PENDING, COMPLIANT, NON_COMPLIANT
}

// LicenseScope represents the scope of the license
type LicenseScope struct {
	AllowedActivities   []string `json:"allowed_activities"`
	RestrictedActivities []string `json:"restricted_activities"`
	GeographicScope     string    `json:"geographic_scope"`
	AssetClasses       []string  `json:"asset_classes"`
	Jurisdictions      []string  `json:"jurisdictions"`
	TransactionLimits  *TransactionLimits `json:"transaction_limits,omitempty"`
}

// TransactionLimits represents transaction limits for the license
type TransactionLimits struct {
	DailyLimit      float64 `json:"daily_limit"`
	MonthlyLimit    float64 `json:"monthly_limit"`
	SingleTxLimit   float64 `json:"single_tx_limit"`
	Currency        string  `json:"currency"`
}

// LicenseFee represents license fees
type LicenseFee struct {
	ApplicationFee   float64 `json:"application_fee"`
	AnnualFee        float64 `json:"annual_fee"`
	AssessmentFee    float64 `json:"assessment_fee"`
	FeeStatus        string  `json:"fee_status"` // PAID, PENDING, OVERDUE
	LastPaymentDate  *time.Time `json:"last_payment_date,omitempty"`
}

// Validate validates the license data
func (l *License) Validate() error {
	if l.EntityID == "" {
		return ErrValidationError("entity ID is required")
	}
	if l.Type == "" {
		return ErrValidationError("license type is required")
	}
	if l.Jurisdiction == "" {
		return ErrValidationError("jurisdiction is required")
	}
	return nil
}

// IsActive checks if the license is active
func (l *License) IsActive() bool {
	return l.Status == LicenseStatusActive
}

// IsExpired checks if the license has expired
func (l *License) IsExpired() bool {
	return time.Now().After(l.ExpiresAt)
}

// DaysUntilExpiry returns days until license expires
func (l *License) DaysUntilExpiry() int {
	days := l.ExpiresAt.Sub(time.Now()) / (24 * time.Hour)
	return int(days)
}

// IsRenewalDue checks if renewal is due within given days
func (l *License) IsRenewalDue(withinDays int) bool {
	if l.RenewalDueDate == nil {
		return false
	}
	renewalWindow := time.Now().AddDate(0, 0, withinDays)
	return l.RenewalDueDate.Before(renewalWindow) || l.RenewalDueDate.Equal(renewalWindow)
}

// CanTransitionTo checks if status transition is valid
func (l *License) CanTransitionTo(newStatus LicenseStatus) bool {
	validTransitions := map[LicenseStatus][]LicenseStatus{
		LicenseStatusDraft:       {LicenseStatusSubmitted},
		LicenseStatusSubmitted:   {LicenseStatusUnderReview, LicenseStatusDraft},
		LicenseStatusUnderReview: {LicenseStatusApproved, LicenseStatusPendingConditions, LicenseStatusRejected, LicenseStatusDraft},
		LicenseStatusPendingConditions: {LicenseStatusApproved, LicenseStatusRejected, LicenseStatusDraft},
		LicenseStatusApproved:    {LicenseStatusActive, LicenseStatusDraft},
		LicenseStatusActive:      {LicenseStatusSuspended, LicenseStatusExpired, LicenseStatusRevoked, LicenseStatusSurrendered},
		LicenseStatusSuspended:   {LicenseStatusActive, LicenseStatusRevoked},
		LicenseStatusExpired:     {LicenseStatusSubmitted, LicenseStatusRevoked},
		LicenseStatusRevoked:     {}, // Terminal state
		LicenseStatusSurrendered: {}, // Terminal state
		LicenseStatusRejected:    {LicenseStatusDraft, LicenseStatusSubmitted},
	}

	allowed, exists := validTransitions[l.Status]
	if !exists {
		return false
	}

	for _, s := range allowed {
		if s == newStatus {
			return true
		}
	}
	return false
}

// TransitionTo performs a status transition
func (l *License) TransitionTo(newStatus LicenseStatus) error {
	if !l.CanTransitionTo(newStatus) {
		return ErrInvalidStateTransition("license", l.Status, newStatus)
	}
	l.Status = newStatus
	return nil
}

// Submit submits the license application
func (l *License) Submit() error {
	return l.TransitionTo(LicenseStatusSubmitted)
}

// Approve approves the license
func (l *License) Approve(officerID string) error {
	if err := l.TransitionTo(LicenseStatusApproved); err != nil {
		return err
	}
	now := time.Now()
	l.ApprovalDate = &now
	l.ApprovalOfficer = officerID
	l.IssuedAt = now
	l.EffectiveDate = now
	return nil
}

// Activate activates the license
func (l *License) Activate() error {
	return l.TransitionTo(LicenseStatusActive)
}

// Suspend suspends the license
func (l *License) Suspend(reason string) error {
	return l.TransitionTo(LicenseStatusSuspended)
}

// Revoke revokes the license
func (l *License) Revoke(reason string) error {
	return l.TransitionTo(LicenseStatusRevoked)
}

// Expire expires the license
func (l *License) Expire() error {
	return l.TransitionTo(LicenseStatusExpired)
}

// Surrender surrenders the license
func (l *License) Surrender() error {
	return l.TransitionTo(LicenseStatusSurrendered)
}

// AddCondition adds a condition to the license
func (l *License) AddCondition(condition LicenseCondition) {
	l.Conditions = append(l.Conditions, condition)
}

// GenerateLicenseNumber generates a license number
func (l *License) GenerateLicenseNumber() string {
	return generateLicenseNumber(l.Type, l.Jurisdiction)
}

// generateLicenseNumber creates a unique license number
func generateLicenseNumber(licenseType LicenseType, jurisdiction string) string {
	prefix := getLicensePrefix(licenseType)
	year := time.Now().Year()
	sequence := generateSequenceNumber()
	return fmt.Sprintf("%s-%s-%d-%s", prefix, jurisdiction, year, sequence)
}

// getLicensePrefix returns the prefix for a license type
func getLicensePrefix(licenseType LicenseType) string {
	prefixes := map[LicenseType]string{
		LicenseTypeExchange:   "EXC",
		LicenseTypeCustodial:  "CUS",
		LicenseTypeWallet:     "WAL",
		LicenseTypeMining:     "MIN",
		LicenseTypeOTC:        "OTC",
		LicenseTypeICO:        "ICO",
		LicenseTypeATM:        "ATM",
	}
	return prefixes[licenseType]
}

// generateSequenceNumber generates a random sequence number
func generateSequenceNumber() string {
	return randomString(6)
}

// randomString generates a random alphanumeric string
func randomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
