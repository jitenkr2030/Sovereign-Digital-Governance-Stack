package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/energy-integration/energy/internal/core/domain"
	"github.com/energy-integration/energy/internal/core/ports"
)

// Common errors for energy service operations.
var (
	ErrProfileNotFound       = errors.New("energy profile not found")
	ErrProfileAlreadyExists  = errors.New("energy profile already exists")
	ErrFootprintNotFound     = errors.New("transaction footprint not found")
	ErrInvalidChainID        = errors.New("invalid blockchain chain ID")
	ErrCertificateNotFound   = errors.New("offset certificate not found")
	ErrCertificateExpired    = errors.New("offset certificate has expired")
	ErrCertificateRedeemed   = errors.New("offset certificate has already been redeemed")
	ErrInsufficientOffset    = errors.New("insufficient carbon offset capacity")
)

// EnergyService provides the core business logic for energy integration.
type EnergyService struct {
	profileRepo       ports.EnergyProfileRepository
	footprintRepo     ports.TransactionFootprintRepository
	certificateRepo   ports.OffsetCertificateRepository
	oracleRepo        ports.ExternalOracleRepository
	calculatorService *CarbonCalculatorService
}

// NewEnergyService creates a new EnergyService with the required dependencies.
func NewEnergyService(
	profileRepo ports.EnergyProfileRepository,
	footprintRepo ports.TransactionFootprintRepository,
	certificateRepo ports.OffsetCertificateRepository,
	oracleRepo ports.ExternalOracleRepository,
) *EnergyService {
	calcService := NewCarbonCalculatorService(profileRepo, oracleRepo)

	return &EnergyService{
		profileRepo:       profileRepo,
		footprintRepo:     footprintRepo,
		certificateRepo:   certificateRepo,
		oracleRepo:        oracleRepo,
		calculatorService: calcService,
	}
}

// CreateProfile creates a new energy profile for a blockchain.
func (s *EnergyService) CreateProfile(
	ctx context.Context,
	chainID, chainName string,
	consensusType domain.ConsensusType,
	avgKWhPerTx, baseCarbonGrams float64,
	energySource domain.EnergySource,
) (*domain.EnergyProfile, error) {
	// Validate chain ID
	if chainID == "" {
		return nil, ErrInvalidChainID
	}

	// Check for existing profile
	existing, err := s.profileRepo.GetByChainID(ctx, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing profile: %w", err)
	}
	if existing != nil {
		return nil, ErrProfileAlreadyExists
	}

	// Validate consensus type
	if !domain.IsValidConsensusType(consensusType) {
		return nil, fmt.Errorf("invalid consensus type: %s", consensusType)
	}

	// Set default carbon intensity if not provided
	carbonIntensity := s.getDefaultCarbonIntensity(energySource)

	profile := &domain.EnergyProfile{
		ID:              domain.NewEntityID(),
		ChainID:         chainID,
		ChainName:       chainName,
		ConsensusType:   consensusType,
		AvgKWhPerTx:     avgKWhPerTx,
		BaseCarbonGrams: baseCarbonGrams,
		EnergySource:    energySource,
		CarbonIntensity: carbonIntensity,
		IsActive:        true,
		Metadata:        make(map[string]interface{}),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	if err := s.profileRepo.Create(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	return profile, nil
}

// GetProfile retrieves an energy profile by chain ID.
func (s *EnergyService) GetProfile(ctx context.Context, chainID string) (*domain.EnergyProfile, error) {
	profile, err := s.profileRepo.GetByChainID(ctx, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}
	return profile, nil
}

// ListProfiles retrieves all active energy profiles.
func (s *EnergyService) ListProfiles(ctx context.Context) ([]*domain.EnergyProfile, error) {
	return s.profileRepo.List(ctx)
}

// CalculateAndRecordFootprint calculates footprint and records it.
func (s *EnergyService) CalculateAndRecordFootprint(
	ctx context.Context,
	txHash, chainID string,
	txData map[string]interface{},
) (*domain.TransactionFootprint, error) {
	// Calculate footprint
	result, err := s.calculatorService.CalculateFootprint(ctx, domain.EnergyCalculationRequest{
		ChainID:        chainID,
		TransactionData: txData,
		Timestamp:      time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to calculate footprint: %w", err)
	}

	// Create footprint record
	footprint := &domain.TransactionFootprint{
		ID:               domain.NewEntityID(),
		TransactionHash:  txHash,
		ChainID:          chainID,
		EnergyUsed:       result.EnergyUsed,
		CarbonEmission:   result.CarbonEmission,
		OffsetStatus:     domain.OffsetStatusNone,
		Timestamp:        time.Now().UTC(),
		Metadata:         txData,
		CreatedAt:        time.Now().UTC(),
	}

	// Persist footprint
	if err := s.footprintRepo.Create(ctx, footprint); err != nil {
		return nil, fmt.Errorf("failed to record footprint: %w", err)
	}

	return footprint, nil
}

// GetFootprint retrieves a transaction footprint by transaction hash.
func (s *EnergyService) GetFootprint(ctx context.Context, txHash string) (*domain.TransactionFootprint, error) {
	footprint, err := s.footprintRepo.GetByTransactionHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get footprint: %w", err)
	}
	if footprint == nil {
		return nil, ErrFootprintNotFound
	}
	return footprint, nil
}

// ListFootprints retrieves footprints with optional filtering.
func (s *EnergyService) ListFootprints(ctx context.Context, filter ports.FootprintFilter) ([]*domain.TransactionFootprint, error) {
	return s.footprintRepo.List(ctx, filter)
}

// GetFootprintSummary retrieves aggregated footprint summary.
func (s *EnergyService) GetFootprintSummary(
	ctx context.Context,
	startTime, endTime time.Time,
	chainID string,
) (*domain.CarbonFootprintSummary, error) {
	return s.footprintRepo.GetSummary(ctx, startTime, endTime, chainID)
}

// ApplyOffset applies a carbon offset certificate to a transaction footprint.
func (s *EnergyService) ApplyOffset(
	ctx context.Context,
	txHash, certificateID string,
) (*domain.TransactionFootprint, error) {
	// Get footprint
	footprint, err := s.footprintRepo.GetByTransactionHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get footprint: %w", err)
	}
	if footprint == nil {
		return nil, ErrFootprintNotFound
	}

	// Check if already fully offset
	if footprint.OffsetStatus == domain.OffsetStatusFull {
		return nil, errors.New("transaction is already fully offset")
	}

	// Get certificate
	cert, err := s.certificateRepo.GetByID(ctx, certificateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}
	if cert == nil {
		return nil, ErrCertificateNotFound
	}

	// Validate certificate
	if cert.Status != "active" {
		return nil, ErrCertificateExpired
	}

	if cert.ExpiryDate.Before(time.Now().UTC()) {
		return nil, ErrCertificateExpired
	}

	// Calculate offset amount
	availableOffset := cert.CarbonOffset.Value
	requiredOffset := footprint.CarbonEmission.Value

	var newStatus domain.OffsetStatus
	var offsetAmount domain.CarbonEmission

	if availableOffset >= requiredOffset {
		// Full offset
		newStatus = domain.OffsetStatusFull
		offsetAmount = footprint.CarbonEmission
	} else {
		// Partial offset
		newStatus = domain.OffsetStatusPartial
		offsetAmount = cert.CarbonOffset
	}

	// Update footprint
	footprint.OffsetStatus = newStatus
	footprint.OffsetAmount = &offsetAmount
	footprint.CertificateID = certificateID

	if err := s.footprintRepo.Update(ctx, footprint); err != nil {
		return nil, fmt.Errorf("failed to update footprint: %w", err)
	}

	// Mark certificate as redeemed (partially or fully)
	if err := s.certificateRepo.MarkAsRedeemed(ctx, certificateID, txHash); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to mark certificate as redeemed: %v\n", err)
	}

	return footprint, nil
}

// CreateCertificate creates a new carbon offset certificate.
func (s *EnergyService) CreateCertificate(
	ctx context.Context,
	certNumber string,
	energySource domain.EnergySource,
	energyAmount float64,
	carbonOffsetGrams float64,
	registry string,
) (*domain.OffsetCertificate, error) {
	// Check for existing certificate
	existing, err := s.certificateRepo.GetByCertificateNumber(ctx, certNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing certificate: %w", err)
	}
	if existing != nil {
		return nil, errors.New("certificate number already exists")
	}

	cert := &domain.OffsetCertificate{
		ID:                domain.NewEntityID(),
		CertificateNumber: certNumber,
		EnergySource:      energySource,
		EnergyAmount:      energyAmount,
		EnergyUnit:        domain.EnergyUnitMegawattHours,
		CarbonOffset: domain.CarbonEmission{
			Value: carbonOffsetGrams,
			Unit:  domain.CarbonUnitGrams,
		},
		Registry:    registry,
		IssueDate:   time.Now().UTC(),
		ExpiryDate:  time.Now().UTC().AddDate(1, 0, 0), // 1 year validity
		Status:      "active",
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.certificateRepo.Create(ctx, cert); err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	return cert, nil
}

// ListCertificates retrieves offset certificates.
func (s *EnergyService) ListCertificates(ctx context.Context, filter ports.CertificateFilter) ([]*domain.OffsetCertificate, error) {
	return s.certificateRepo.List(ctx, filter)
}

// getDefaultCarbonIntensity returns default carbon intensity based on energy source.
func (s *EnergyService) getDefaultCarbonIntensity(source domain.EnergySource) float64 {
	switch source {
	case domain.EnergySourceRenewable, domain.EnergySourceSolar, domain.EnergySourceWind, domain.EnergySourceHydro:
		return 0.02 // Very low
	case domain.EnergySourceNuclear:
		return 0.01 // Extremely low
	case domain.EnergySourceMixed:
		return 0.5 // Global average
	case domain.EnergySourceFossilFuel:
		return 0.9 // High
	default:
		return 0.5
	}
}
