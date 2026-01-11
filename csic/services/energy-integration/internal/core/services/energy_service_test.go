package services

import (
	"context"
	"testing"
	"time"

	"github.com/energy-integration/energy/internal/core/domain"
	"github.com/energy-integration/energy/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEnergyProfileRepository is a mock implementation of EnergyProfileRepository.
type MockEnergyProfileRepository struct {
	mock.Mock
}

func (m *MockEnergyProfileRepository) Create(ctx context.Context, profile *domain.EnergyProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockEnergyProfileRepository) GetByID(ctx context.Context, id string) (*domain.EnergyProfile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EnergyProfile), args.Error(1)
}

func (m *MockEnergyProfileRepository) GetByChainID(ctx context.Context, chainID string) (*domain.EnergyProfile, error) {
	args := m.Called(ctx, chainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EnergyProfile), args.Error(1)
}

func (m *MockEnergyProfileRepository) Update(ctx context.Context, profile *domain.EnergyProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockEnergyProfileRepository) List(ctx context.Context) ([]*domain.EnergyProfile, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.EnergyProfile), args.Error(1)
}

func (m *MockEnergyProfileRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockTransactionFootprintRepository is a mock implementation of TransactionFootprintRepository.
type MockTransactionFootprintRepository struct {
	mock.Mock
}

func (m *MockTransactionFootprintRepository) Create(ctx context.Context, footprint *domain.TransactionFootprint) error {
	args := m.Called(ctx, footprint)
	return args.Error(0)
}

func (m *MockTransactionFootprintRepository) GetByID(ctx context.Context, id string) (*domain.TransactionFootprint, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TransactionFootprint), args.Error(1)
}

func (m *MockTransactionFootprintRepository) GetByTransactionHash(ctx context.Context, txHash string) (*domain.TransactionFootprint, error) {
	args := m.Called(ctx, txHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TransactionFootprint), args.Error(1)
}

func (m *MockTransactionFootprintRepository) Update(ctx context.Context, footprint *domain.TransactionFootprint) error {
	args := m.Called(ctx, footprint)
	return args.Error(0)
}

func (m *MockTransactionFootprintRepository) UpdateOffsetStatus(ctx context.Context, id string, status domain.OffsetStatus, certificateID string) error {
	args := m.Called(ctx, id, status, certificateID)
	return args.Error(0)
}

func (m *MockTransactionFootprintRepository) List(ctx context.Context, filter ports.FootprintFilter) ([]*domain.TransactionFootprint, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.TransactionFootprint), args.Error(1)
}

func (m *MockTransactionFootprintRepository) GetSummary(ctx context.Context, startTime, endTime time.Time, chainID string) (*domain.CarbonFootprintSummary, error) {
	args := m.Called(ctx, startTime, endTime, chainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CarbonFootprintSummary), args.Error(1)
}

func (m *MockTransactionFootprintRepository) Count(ctx context.Context, filter ports.FootprintFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockTransactionFootprintRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockOffsetCertificateRepository is a mock implementation of OffsetCertificateRepository.
type MockOffsetCertificateRepository struct {
	mock.Mock
}

func (m *MockOffsetCertificateRepository) Create(ctx context.Context, cert *domain.OffsetCertificate) error {
	args := m.Called(ctx, cert)
	return args.Error(0)
}

func (m *MockOffsetCertificateRepository) GetByID(ctx context.Context, id string) (*domain.OffsetCertificate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OffsetCertificate), args.Error(1)
}

func (m *MockOffsetCertificateRepository) GetByCertificateNumber(ctx context.Context, number string) (*domain.OffsetCertificate, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OffsetCertificate), args.Error(1)
}

func (m *MockOffsetCertificateRepository) Update(ctx context.Context, cert *domain.OffsetCertificate) error {
	args := m.Called(ctx, cert)
	return args.Error(0)
}

func (m *MockOffsetCertificateRepository) MarkAsRedeemed(ctx context.Context, id string, footprintID string) error {
	args := m.Called(ctx, id, footprintID)
	return args.Error(0)
}

func (m *MockOffsetCertificateRepository) List(ctx context.Context, filter ports.CertificateFilter) ([]*domain.OffsetCertificate, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.OffsetCertificate), args.Error(1)
}

func (m *MockOffsetCertificateRepository) Count(ctx context.Context, filter ports.CertificateFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return int64(args.Int(0)), args.Error(1)
}

// MockExternalOracleRepository is a mock implementation of ExternalOracleRepository.
type MockExternalOracleRepository struct {
	mock.Mock
}

func (m *MockExternalOracleRepository) GetCarbonIntensity(ctx context.Context, regionCode string) (float64, error) {
	args := m.Called(ctx, regionCode)
	return float64(args.Float(0)), args.Error(1)
}

func (m *MockExternalOracleRepository) VerifyCertificate(ctx context.Context, certificateNumber string) (*domain.OffsetCertificate, error) {
	args := m.Called(ctx, certificateNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OffsetCertificate), args.Error(1)
}

func (m *MockExternalOracleRepository) GetEnergyMix(ctx context.Context, regionCode string) (map[domain.EnergySource]float64, error) {
	args := m.Called(ctx, regionCode)
	return args.Get(0).(map[domain.EnergySource]float64), args.Error(1)
}

// Helper function to create a test energy profile.
func createTestProfile(chainID string) *domain.EnergyProfile {
	return &domain.EnergyProfile{
		ID:              domain.NewEntityID(),
		ChainID:         chainID,
		ChainName:       "Test Chain " + chainID,
		ConsensusType:   domain.ConsensusProofOfStake,
		AvgKWhPerTx:     0.001,
		BaseCarbonGrams: 0.5,
		EnergySource:    domain.EnergySourceRenewable,
		CarbonIntensity: 0.02,
		IsActive:        true,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
}

// Helper function to create a test footprint.
func createTestFootprint(txHash string, chainID string) *domain.TransactionFootprint {
	return &domain.TransactionFootprint{
		ID:              domain.NewEntityID(),
		TransactionHash: txHash,
		ChainID:         chainID,
		EnergyUsed: domain.EnergyConsumption{
			Value:  0.001,
			Unit:   domain.EnergyUnitKilowattHours,
			Source: domain.EnergySourceRenewable,
		},
		CarbonEmission: domain.CarbonEmission{
			Value: 0.02,
			Unit:  domain.CarbonUnitGrams,
		},
		OffsetStatus: domain.OffsetStatusNone,
		Timestamp:    time.Now().UTC(),
		CreatedAt:    time.Now().UTC(),
	}
}

// Helper function to create a test certificate.
func createTestCertificate(certNumber string) *domain.OffsetCertificate {
	return &domain.OffsetCertificate{
		ID:                domain.NewEntityID(),
		CertificateNumber: certNumber,
		EnergySource:      domain.EnergySourceRenewable,
		EnergyAmount:      100.0,
		EnergyUnit:        domain.EnergyUnitMegawattHours,
		CarbonOffset: domain.CarbonEmission{
			Value: 50000.0, // 50kg of carbon offsets
			Unit:  domain.CarbonUnitGrams,
		},
		Registry:   "I-REC",
		IssueDate:  time.Now().UTC(),
		ExpiryDate: time.Now().UTC().AddDate(1, 0, 0),
		Status:     "active",
		CreatedAt:  time.Now().UTC(),
	}
}

func TestEnergyService_CreateProfile_Success(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	profileRepo.On("GetByChainID", ctx, "ethereum-pos").Return(nil, nil)
	profileRepo.On("Create", ctx, mock.AnythingOfType("*domain.EnergyProfile")).Return(nil)

	profile, err := service.CreateProfile(
		ctx,
		"ethereum-pos",
		"Ethereum Proof of Stake",
		domain.ConsensusProofOfStake,
		0.0006,
		0.3,
		domain.EnergySourceMixed,
	)

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "ethereum-pos", profile.ChainID)
	assert.Equal(t, domain.ConsensusProofOfStake, profile.ConsensusType)
	profileRepo.AssertExpectations(t)
}

func TestEnergyService_CreateProfile_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	existingProfile := createTestProfile("ethereum-pos")
	profileRepo.On("GetByChainID", ctx, "ethereum-pos").Return(existingProfile, nil)

	_, err := service.CreateProfile(
		ctx,
		"ethereum-pos",
		"Ethereum Proof of Stake",
		domain.ConsensusProofOfStake,
		0.0006,
		0.3,
		domain.EnergySourceMixed,
	)

	assert.Error(t, err)
	assert.Equal(t, ErrProfileAlreadyExists, err)
	profileRepo.AssertExpectations(t)
}

func TestEnergyService_CreateProfile_InvalidChainID(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	_, err := service.CreateProfile(
		ctx,
		"",
		"Test Chain",
		domain.ConsensusProofOfStake,
		0.001,
		0.5,
		domain.EnergySourceRenewable,
	)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidChainID, err)
}

func TestEnergyService_GetProfile_Success(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	expectedProfile := createTestProfile("ethereum")
	profileRepo.On("GetByChainID", ctx, "ethereum").Return(expectedProfile, nil)

	profile, err := service.GetProfile(ctx, "ethereum")

	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "ethereum", profile.ChainID)
	profileRepo.AssertExpectations(t)
}

func TestEnergyService_GetProfile_NotFound(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	profileRepo.On("GetByChainID", ctx, "unknown-chain").Return(nil, nil)

	_, err := service.GetProfile(ctx, "unknown-chain")

	assert.Error(t, err)
	assert.Equal(t, ErrProfileNotFound, err)
	profileRepo.AssertExpectations(t)
}

func TestEnergyService_CalculateAndRecordFootprint_Success(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	profile := createTestProfile("ethereum-pos")
	profileRepo.On("GetByChainID", ctx, "ethereum-pos").Return(profile, nil)
	footprintRepo.On("Create", ctx, mock.AnythingOfType("*domain.TransactionFootprint")).Return(nil)

	footprint, err := service.CalculateAndRecordFootprint(
		ctx,
		"0x123abc",
		"ethereum-pos",
		map[string]interface{}{"gas_used": 50000.0},
	)

	assert.NoError(t, err)
	assert.NotNil(t, footprint)
	assert.Equal(t, "0x123abc", footprint.TransactionHash)
	assert.Equal(t, "ethereum-pos", footprint.ChainID)
	assert.Equal(t, domain.OffsetStatusNone, footprint.OffsetStatus)
	profileRepo.AssertExpectations(t)
	footprintRepo.AssertExpectations(t)
}

func TestEnergyService_GetFootprint_Success(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	expectedFootprint := createTestFootprint("0x456def", "ethereum")
	footprintRepo.On("GetByTransactionHash", ctx, "0x456def").Return(expectedFootprint, nil)

	footprint, err := service.GetFootprint(ctx, "0x456def")

	assert.NoError(t, err)
	assert.NotNil(t, footprint)
	assert.Equal(t, "0x456def", footprint.TransactionHash)
	footprintRepo.AssertExpectations(t)
}

func TestEnergyService_GetFootprint_NotFound(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	footprintRepo.On("GetByTransactionHash", ctx, "0xnonexistent").Return(nil, nil)

	_, err := service.GetFootprint(ctx, "0xnonexistent")

	assert.Error(t, err)
	assert.Equal(t, ErrFootprintNotFound, err)
	footprintRepo.AssertExpectations(t)
}

func TestEnergyService_ApplyOffset_Success(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	footprint := createTestFootprint("0x789ghi", "ethereum")
	footprint.CarbonEmission.Value = 10.0 // 10g of carbon
	footprintRepo.On("GetByTransactionHash", ctx, "0x789ghi").Return(footprint, nil)

	certificate := createTestCertificate("CERT-001")
	certificate.CarbonOffset.Value = 50.0 // 50g available
	certificateRepo.On("GetByID", ctx, "cert-123").Return(certificate, nil)
	footprintRepo.On("Update", ctx, mock.AnythingOfType("*domain.TransactionFootprint")).Return(nil)
	certificateRepo.On("MarkAsRedeemed", ctx, "cert-123", mock.AnythingOfType("string")).Return(nil)

	result, err := service.ApplyOffset(ctx, "0x789ghi", "cert-123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.OffsetStatusFull, result.OffsetStatus)
	assert.Equal(t, "cert-123", result.CertificateID)
	footprintRepo.AssertExpectations(t)
	certificateRepo.AssertExpectations(t)
}

func TestEnergyService_ApplyOffset_CertificateExpired(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	footprint := createTestFootprint("0xabc", "ethereum")
	footprintRepo.On("GetByTransactionHash", ctx, "0xabc").Return(footprint, nil)

	certificate := createTestCertificate("CERT-002")
	certificate.ExpiryDate = time.Now().UTC().AddDate(-1, 0, 0) // Expired yesterday
	certificateRepo.On("GetByID", ctx, "cert-expired").Return(certificate, nil)

	_, err := service.ApplyOffset(ctx, "0xabc", "cert-expired")

	assert.Error(t, err)
	assert.Equal(t, ErrCertificateExpired, err)
	footprintRepo.AssertExpectations(t)
	certificateRepo.AssertExpectations(t)
}

func TestEnergyService_ApplyOffset_AlreadyFullyOffset(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	footprint := createTestFootprint("0xdef", "ethereum")
	footprint.OffsetStatus = domain.OffsetStatusFull
	footprintRepo.On("GetByTransactionHash", ctx, "0xdef").Return(footprint, nil)

	_, err := service.ApplyOffset(ctx, "0xdef", "cert-new")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already fully offset")
	footprintRepo.AssertExpectations(t)
}

func TestEnergyService_CreateCertificate_Success(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	certificateRepo.On("GetByCertificateNumber", ctx, "I-REC-2024-001").Return(nil, nil)
	certificateRepo.On("Create", ctx, mock.AnythingOfType("*domain.OffsetCertificate")).Return(nil)

	cert, err := service.CreateCertificate(
		ctx,
		"I-REC-2024-001",
		domain.EnergySourceRenewable,
		100.0,
		50000.0,
		"I-REC",
	)

	assert.NoError(t, err)
	assert.NotNil(t, cert)
	assert.Equal(t, "I-REC-2024-001", cert.CertificateNumber)
	assert.Equal(t, "active", cert.Status)
	certificateRepo.AssertExpectations(t)
}

func TestEnergyService_ListProfiles_Success(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	expectedProfiles := []*domain.EnergyProfile{
		createTestProfile("ethereum"),
		createTestProfile("bitcoin"),
		createTestProfile("polygon"),
	}
	profileRepo.On("List", ctx).Return(expectedProfiles, nil)

	profiles, err := service.ListProfiles(ctx)

	assert.NoError(t, err)
	assert.Len(t, profiles, 3)
	profileRepo.AssertExpectations(t)
}

func TestEnergyService_ListFootprints_WithFilter(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	footprintRepo := new(MockTransactionFootprintRepository)
	certificateRepo := new(MockOffsetCertificateRepository)
	oracleRepo := new(MockExternalOracleRepository)

	service := NewEnergyService(profileRepo, footprintRepo, certificateRepo, oracleRepo)

	expectedFootprints := []*domain.TransactionFootprint{
		createTestFootprint("0xabc1", "ethereum"),
		createTestFootprint("0xabc2", "ethereum"),
	}

	filter := ports.FootprintFilter{
		ChainID:      "ethereum",
		OffsetStatus: []domain.OffsetStatus{domain.OffsetStatusNone},
		Limit:        100,
	}
	footprintRepo.On("List", ctx, filter).Return(expectedFootprints, nil)

	footprints, err := service.ListFootprints(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, footprints, 2)
	footprintRepo.AssertExpectations(t)
}
