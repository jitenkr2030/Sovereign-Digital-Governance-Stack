// Compliance Management Module - Obligation Models
// Compliance obligations and deadline tracking

package domain

import (
	"time"
)

// ObligationType represents the type of compliance obligation
type ObligationType string

const (
	ObligationTypeReport          ObligationType = "REPORT"
	ObligationTypeAudit           ObligationType = "AUDIT"
	ObligationTypeInspection      ObligationType = "INSPECTION"
	ObligationTypeSubmission      ObligationType = "SUBMISSION"
	ObligationTypeCertification   ObligationType = "CERTIFICATION"
	ObligationTypeFeePayment      ObligationType = "FEE_PAYMENT"
	ObligationTypeReview          ObligationType = "REVIEW"
	ObligationTypeTraining        ObligationType = "TRAINING"
	ObligationTypePolicyUpdate    ObligationType = "POLICY_UPDATE"
)

// ObligationStatus represents the status of an obligation
type ObligationStatus string

const (
	ObligationStatusPending      ObligationStatus = "PENDING"
	ObligationStatusInProgress   ObligationStatus = "IN_PROGRESS"
	ObligationStatusSubmitted    ObligationStatus = "SUBMITTED"
	ObligationStatusUnderReview  ObligationStatus = "UNDER_REVIEW"
	ObligationStatusVerified     ObligationStatus = "VERIFIED"
	ObligationStatusApproved     ObligationStatus = "APPROVED"
	ObligationStatusOverdue      ObligationStatus = "OVERDUE"
	ObligationStatusNonCompliant ObligationStatus = "NON_COMPLIANT"
	ObligationStatusWaived       ObligationStatus = "WAIVED"
	ObligationStatusCancelled    ObligationStatus = "CANCELLED"
)

// ObligationPriority represents the priority of an obligation
type ObligationPriority string

const (
	ObligationPriorityLow      ObligationPriority = "LOW"
	ObligationPriorityMedium   ObligationPriority = "MEDIUM"
	ObligationPriorityHigh     ObligationPriority = "HIGH"
	ObligationPriorityCritical ObligationPriority = "CRITICAL"
)

// ComplianceObligation represents a compliance obligation
type ComplianceObligation struct {
	ID                string              `json:"id" db:"id"`
	LicenseID         string              `json:"license_id" db:"license_id"`
	EntityID          string              `json:"entity_id" db:"entity_id"`
	EntityName        string              `json:"entity_name" db:"entity_name"`
	Type              ObligationType      `json:"type" db:"type"`
	Category          string              `json:"category" db:"category"`
	Title             string              `json:"title" db:"title"`
	Description       string              `json:"description" db:"description"`
	ComplianceRef     string              `json:"compliance_ref" db:"compliance_ref"` // e.g., "FATF-2024-001"
	RegulatoryBody    string              `json:"regulatory_body" db:"regulatory_body"`
	Priority          ObligationPriority  `json:"priority" db:"priority"`
	Status            ObligationStatus    `json:"status" db:"status"`
	CreatedAt         time.Time           `json:"created_at" db:"created_at"`
	AssignedTo        string              `json:"assigned_to" db:"assigned_to"`
	DueDate           time.Time           `json:"due_date" db:"due_date"`
	ReminderDate      *time.Time          `json:"reminder_date,omitempty" db:"reminder_date"`
	SubmittedDate     *time.Time          `json:"submitted_date,omitempty" db:"submitted_date"`
	ReviewDate        *time.Time          `json:"review_date,omitempty" db:"review_date"`
	CompletedDate     *time.Time          `json:"completed_date,omitempty" db:"completed_date"`
	Frequency         string              `json:"frequency"` // ONCE, DAILY, WEEKLY, MONTHLY, QUARTERLY, ANNUALLY
	Recurring         bool                `json:"recurring"`
	NextOccurrence    *time.Time          `json:"next_occurrence,omitempty"`
	DocumentIDs       []string            `json:"document_ids"`
	Attachments       []Attachment        `json:"attachments"`
	ReviewNotes       string              `json:"review_notes,omitempty"`
	VerificationNotes string              `json:"verification_notes,omitempty"`
	WaiverReason      string              `json:"waiver_reason,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// Attachment represents a document attachment
type Attachment struct {
	ID          string    `json:"id"`
	FileName    string    `json:"file_name"`
	FileType    string    `json:"file_type"`
	FileSize    int64     `json:"file_size"`
	UploadedAt  time.Time `json:"uploaded_at"`
	UploadedBy  string    `json:"uploaded_by"`
}

// Validate validates the obligation data
func (o *ComplianceObligation) Validate() error {
	if o.EntityID == "" {
		return ErrValidationError("entity ID is required")
	}
	if o.Type == "" {
		return ErrValidationError("obligation type is required")
	}
	if o.Title == "" {
		return ErrValidationError("title is required")
	}
	if o.DueDate.IsZero() {
		return ErrValidationError("due date is required")
	}
	return nil
}

// IsOverdue checks if the obligation is overdue
func (o *ComplianceObligation) IsOverdue() bool {
	if o.Status == ObligationStatusCompleted() || o.Status == ObligationStatusVerified() ||
	   o.Status == ObligationStatusApproved() || o.Status == ObligationStatusWaived() ||
	   o.Status == ObligationStatusCancelled() {
		return false
	}
	return time.Now().After(o.DueDate)
}

// IsDueSoon checks if the obligation is due within given days
func (o *ComplianceObligation) IsDueSoon(withinDays int) bool {
	window := time.Now().AddDate(0, 0, withinDays)
	return time.Now().Before(o.DueDate) && o.DueDate.Before(window) || o.DueDate.Equal(window)
}

// DaysUntilDue returns days until obligation is due
func (o *ComplianceObligation) DaysUntilDue() int {
	if o.DueDate.Before(time.Now()) {
		return -1
	}
	days := o.DueDate.Sub(time.Now()) / (24 * time.Hour)
	return int(days)
}

// Status methods for obligation
func ObligationStatusPending() ObligationStatus      { return ObligationStatusPending }
func ObligationStatusInProgress() ObligationStatus   { return ObligationStatusInProgress }
func ObligationStatusSubmitted() ObligationStatus    { return ObligationStatusSubmitted }
func ObligationStatusUnderReview() ObligationStatus  { return ObligationStatusUnderReview }
func ObligationStatusVerified() ObligationStatus     { return ObligationStatusVerified }
func ObligationStatusApproved() ObligationStatus     { return ObligationStatusApproved }
func ObligationStatusOverdue() ObligationStatus      { return ObligationStatusOverdue }
func ObligationStatusNonCompliant() ObligationStatus { return ObligationStatusNonCompliant }
func ObligationStatusWaived() ObligationStatus       { return ObligationStatusWaived }
func ObligationStatusCancelled() ObligationStatus    { return ObligationStatusCancelled }

// Complete marks the obligation as completed
func (o *ComplianceObligation) Complete() error {
	if o.Status == ObligationStatusOverdue {
		return ErrValidationError("cannot complete overdue obligation without review")
	}
	now := time.Now()
	o.CompletedDate = &now
	o.Status = ObligationStatusSubmitted
	return nil
}

// Verify marks the obligation as verified
func (o *ComplianceObligation) Verify(notes string) error {
	o.Status = ObligationStatusVerified
	o.VerificationNotes = notes
	now := time.Now()
	o.CompletedDate = &now
	return nil
}

// Approve marks the obligation as approved
func (o *ComplianceObligation) Approve(notes string) error {
	o.Status = ObligationStatusApproved
	o.VerificationNotes = notes
	return nil
}

// Reject marks the obligation as non-compliant
func (o *ComplianceObligation) Reject(notes string) error {
	o.Status = ObligationStatusNonCompliant
	o.ReviewNotes = notes
	return nil
}

// MarkOverdue marks the obligation as overdue
func (o *ComplianceObligation) MarkOverdue() error {
	o.Status = ObligationStatusOverdue
	return nil
}

// Waive waives the obligation
func (o *ComplianceObligation) Waive(reason string) error {
	o.Status = ObligationStatusWaived
	o.WaiverReason = reason
	return nil
}

// Cancel cancels the obligation
func (o *ComplianceObligation) Cancel(reason string) error {
	o.Status = ObligationStatusCancelled
	return nil
}

// StartProgress marks the obligation as in progress
func (o *ComplianceObligation) StartProgress() error {
	if o.Status != ObligationStatusPending {
		return ErrInvalidStateTransition("obligation", o.Status, ObligationStatusInProgress)
	}
	o.Status = ObligationStatusInProgress
	return nil
}

// Submit submits the obligation documentation
func (o *ComplianceObligation) Submit() error {
	if o.Status != ObligationStatusInProgress {
		return ErrInvalidStateTransition("obligation", o.Status, ObligationStatusSubmitted)
	}
	now := time.Now()
	o.SubmittedDate = &now
	o.Status = ObligationStatusSubmitted
	return nil
}

// AddAttachment adds an attachment to the obligation
func (o *ComplianceObligation) AddAttachment(attachment Attachment) {
	o.Attachments = append(o.Attachments, attachment)
	o.DocumentIDs = append(o.DocumentIDs, attachment.ID)
}

// SetNextOccurrence sets the next occurrence for recurring obligations
func (o *ComplianceObligation) SetNextOccurrence(interval string) error {
	if !o.Recurring {
		return nil
	}

	var nextDate time.Time
	now := time.Now()

	switch interval {
	case "DAILY":
		nextDate = now.AddDate(0, 0, 1)
	case "WEEKLY":
		nextDate = now.AddDate(0, 0, 7)
	case "MONTHLY":
		nextDate = now.AddDate(0, 1, 0)
	case "QUARTERLY":
		nextDate = now.AddDate(0, 3, 0)
	case "ANNUALLY":
		nextDate = now.AddDate(1, 0, 0)
	default:
		return ErrValidationError("invalid frequency")
	}

	o.NextOccurrence = &nextDate
	return nil
}

// ObligationTemplate represents a template for creating obligations
type ObligationTemplate struct {
	ID              string             `json:"id"`
	LicenseType     LicenseType        `json:"license_type"`
	Type            ObligationType     `json:"type"`
	Category        string             `json:"category"`
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	ComplianceRef   string             `json:"compliance_ref"`
	RegulatoryBody  string             `json:"regulatory_body"`
	Priority        ObligationPriority `json:"priority"`
	Frequency       string             `json:"frequency"`
	Recurring       bool               `json:"recurring"`
	DueInDays       int                `json:"due_in_days"`
}
