package ports

import (
	"context"

	"github.com/energy-integration/energy/internal/core/domain"
)

// EnergyProfileRepository defines the interface for energy profile persistence.
type EnergyProfileRepository interface {
	// Create creates a new energy profile.
	Create(ctx context.Context, profile *domain.EnergyProfile) error

	// GetByID retrieves an energy profile by ID.
	GetByID(ctx context.Context, id string) (*domain.EnergyProfile, error)

	// GetByChainID retrieves an energy profile by blockchain chain ID.
	GetByChainID(ctx context.Context, chainID string) (*domain.EnergyProfile, error)

	// Update updates an existing energy profile.
	Update(ctx context.Context, profile *domain.EnergyProfile) error

	// List retrieves all active energy profiles.
	List(ctx context.Context) ([]*domain.EnergyProfile, error)

	// Delete soft-deletes an energy profile.
	Delete(ctx context.Context, id string) error
}

// TransactionFootprintRepository defines the interface for transaction footprint persistence.
type TransactionFootprintRepository interface {
	// Create creates a new transaction footprint.
	Create(ctx context.Context, footprint *domain.TransactionFootprint) error

	// GetByID retrieves a transaction footprint by ID.
	GetByID(ctx context.Context, id string) (*domain.TransactionFootprint, error)

	// GetByTransactionHash retrieves footprint by transaction hash.
	GetByTransactionHash(ctx context.Context, txHash string) (*domain.TransactionFootprint, error)

	// Update updates an existing transaction footprint.
	Update(ctx context.Context, footprint *domain.TransactionFootprint) error

	// UpdateOffsetStatus updates only the offset status.
	UpdateOffsetStatus(ctx context.Context, id string, status domain.OffsetStatus, certificateID string) error

	// List retrieves footprints with optional filtering.
	List(ctx context.Context, filter FootprintFilter) ([]*domain.TransactionFootprint, error)

	// GetSummary retrieves aggregated footprint summary.
	GetSummary(ctx context.Context, startTime, endTime time.Time, chainID string) (*domain.CarbonFootprintSummary, error)

	// Count returns the count of footprints matching the filter.
	Count(ctx context.Context, filter FootprintFilter) (int64, error)

	// Delete soft-deletes a transaction footprint.
	Delete(ctx context.Context, id string) error
}

// FootprintFilter represents filtering criteria for footprint queries.
type FootprintFilter struct {
	ChainID      string
	OffsetStatus []domain.OffsetStatus
	StartTime    *time.Time
	EndTime      *time.Time
	MinCarbonGrams float64
	MaxCarbonGrams float64
	Limit        int
	Offset       int
}

// OffsetCertificateRepository defines the interface for offset certificate persistence.
type OffsetCertificateRepository interface {
	// Create creates a new offset certificate.
	Create(ctx context.Context, cert *domain.OffsetCertificate) error

	// GetByID retrieves an offset certificate by ID.
	GetByID(ctx context.Context, id string) (*domain.OffsetCertificate, error)

	// GetByCertificateNumber retrieves certificate by its number.
	GetByCertificateNumber(ctx context.Context, number string) (*domain.OffsetCertificate, error)

	// Update updates an existing offset certificate.
	Update(ctx context.Context, cert *domain.OffsetCertificate) error

	// MarkAsRedeemed marks a certificate as redeemed.
	MarkAsRedeemed(ctx context.Context, id string, footprintID string) error

	// List retrieves certificates with optional filtering.
	List(ctx context.Context, filter CertificateFilter) ([]*domain.OffsetCertificate, error)

	// Count returns the count of certificates matching the filter.
	Count(ctx context.Context, filter CertificateFilter) (int64, error)
}

// CertificateFilter represents filtering criteria for certificate queries.
type CertificateFilter struct {
	Status     []string
	EnergySource domain.EnergySource
	Registry   string
	StartDate  *time.Time
	EndDate    *time.Time
	Limit      int
	Offset     int
}

// ExternalOracleRepository defines the interface for external carbon data oracles.
type ExternalOracleRepository interface {
	// GetCarbonIntensity retrieves current carbon intensity for a region.
	GetCarbonIntensity(ctx context.Context, regionCode string) (float64, error)

	// VerifyCertificate verifies a certificate with the external registry.
	VerifyCertificate(ctx context.Context, certificateNumber string) (*domain.OffsetCertificate, error)

	// GetEnergyMix retrieves the energy mix for a region.
	GetEnergyMix(ctx context.Context, regionCode string) (map[domain.EnergySource]float64, error)
}
