// Compliance Management Module - Domain Models
// Core business entities for regulatory compliance

package domain

import (
	"time"
)

// EntityType represents the type of regulated entity
type EntityType string

const (
	EntityTypeExchange      EntityType = "EXCHANGE"
	EntityTypeMiner         EntityType = "MINER"
	EntityTypeWalletProvider EntityType = "WALLET_PROVIDER"
	EntityTypeCustodian     EntityType = "CUSTODIAN"
	EntityTypeICOIssuer     EntityType = "ICO_ISSUER"
	EntityTypeOTCDesk       EntityType = "OTC_DESK"
)

// EntityStatus represents the operational status of an entity
type EntityStatus string

const (
	EntityStatusPending    EntityStatus = "PENDING"
	EntityStatusActive     EntityStatus = "ACTIVE"
	EntityStatusSuspended  EntityStatus = "SUSPENDED"
	EntityStatusRevoked    EntityStatus = "REVOKED"
	EntityStatusDormant    EntityStatus = "DORMANT"
	EntityStatusTerminated EntityStatus = "TERMINATED"
)

// RegulatedEntity represents a regulated organization
type RegulatedEntity struct {
	ID                 string            `json:"id" db:"id"`
	RegistrationNumber string            `json:"registration_number" db:"registration_number"`
	Name               string            `json:"name" db:"name"`
	Type               EntityType        `json:"type" db:"type"`
	Status             EntityStatus      `json:"status" db:"status"`
	Jurisdiction       string            `json:"jurisdiction" db:"jurisdiction"`
	RegistrationDate   time.Time         `json:"registration_date" db:"registration_date"`
	Address            Address           `json:"address"`
	ContactInfo        ContactInfo       `json:"contact_info"`
	AuthorizedUsers    []AuthorizedUser  `json:"authorized_users"`
	BankAccounts       []BankAccount     `json:"bank_accounts"`
	BlockchainAddresses []BlockchainAddr `json:"blockchain_addresses"`
	LicenseIDs         []string          `json:"license_ids"`
	RiskRating         string            `json:"risk_rating"`
	ComplianceScore    float64           `json:"compliance_score"`
	LastInspectionDate *time.Time        `json:"last_inspection_date,omitempty"`
	CreatedAt          time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at" db:"updated_at"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// Address represents a physical address
type Address struct {
	Street1     string `json:"street1"`
	Street2     string `json:"street2,omitempty"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	Country     string `json:"country"`
}

// ContactInfo represents contact information
type ContactInfo struct {
	PrimaryContact string `json:"primary_contact"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Website        string `json:"website,omitempty"`
}

// AuthorizedUser represents an authorized representative
type AuthorizedUser struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Role        string    `json:"role"`
	StartDate   time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

// BankAccount represents a bank account for the entity
type BankAccount struct {
	ID            string `json:"id"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	SWIFTCode     string `json:"swift_code"`
	Currency      string `json:"currency"`
	IsPrimary     bool   `json:"is_primary"`
}

// BlockchainAddr represents a blockchain address
type BlockchainAddr struct {
	ID        string    `json:"id"`
	Chain     string    `json:"chain"`
	Address   string    `json:"address"`
	Tag       string    `json:"tag,omitempty"`
	Label     string    `json:"label,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Validate validates the entity data
func (e *RegulatedEntity) Validate() error {
	if e.Name == "" {
		return ErrValidationError("entity name is required")
	}
	if e.RegistrationNumber == "" {
		return ErrValidationError("registration number is required")
	}
	if e.Type == "" {
		return ErrValidationError("entity type is required")
	}
	if e.Jurisdiction == "" {
		return ErrValidationError("jurisdiction is required")
	}
	return nil
}

// CanApplyForLicense checks if entity can apply for a license
func (e *RegulatedEntity) CanApplyForLicense() bool {
	return e.Status == EntityStatusPending || e.Status == EntityStatusActive
}

// IsOperational checks if entity can conduct regulated activities
func (e *RegulatedEntity) IsOperational() bool {
	return e.Status == EntityStatusActive
}

// Activate activates the entity
func (e *RegulatedEntity) Activate() error {
	if e.Status != EntityStatusPending && e.Status != EntityStatusDormant {
		return ErrInvalidStateTransition("entity", e.Status, EntityStatusActive)
	}
	e.Status = EntityStatusActive
	return nil
}

// Suspend suspends the entity
func (e *RegulatedEntity) Suspend(reason string) error {
	if e.Status == EntityStatusRevoked || e.Status == EntityStatusTerminated {
		return ErrInvalidStateTransition("entity", e.Status, EntityStatusSuspended)
	}
	e.Status = EntityStatusSuspended
	return nil
}

// Revoke revokes the entity
func (e *RegulatedEntity) Revoke(reason string) error {
	if e.Status == EntityStatusRevoked || e.Status == EntityStatusTerminated {
		return ErrInvalidStateTransition("entity", e.Status, EntityStatusRevoked)
	}
	e.Status = EntityStatusRevoked
	return nil
}

// AddLicense adds a license ID to the entity
func (e *RegulatedEntity) AddLicense(licenseID string) {
	for _, id := range e.LicenseIDs {
		if id == licenseID {
			return
		}
		e.LicenseIDs = append(e.LicenseIDs, licenseID)
	}
}

// RemoveLicense removes a license ID from the entity
func (e *RegulatedEntity) RemoveLicense(licenseID string) {
	newIDs := make([]string, 0)
	for _, id := range e.LicenseIDs {
		if id != licenseID {
			newIDs = append(newIDs, id)
		}
	}
	e.LicenseIDs = newIDs
}

// GetActiveLicenseCount returns the count of active licenses
func (e *RegulatedEntity) GetActiveLicenseCount() int {
	return len(e.LicenseIDs)
}
