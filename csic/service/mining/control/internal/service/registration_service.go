package service

import (
	"context"
	"log"
	"time"

	"github.com/csic/mining-control/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ComplianceRepository defines the interface for compliance verification
type ComplianceRepository interface {
	IsEntityLicensed(ctx context.Context, entityID uuid.UUID) (bool, error)
	GetEntityDetails(ctx context.Context, entityID uuid.UUID) (*EntityDetails, error)
}

// EntityDetails represents entity details from the compliance system
type EntityDetails struct {
	EntityID   uuid.UUID
	Name       string
	Type       string
	KYCStatus  string
	AMLStatus  string
	RiskScore  float64
}

// MiningPoolRepository defines the interface for mining pool persistence
type MiningPoolRepository interface {
	Create(ctx context.Context, pool *domain.MiningPool) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.MiningPool, error)
	GetByLicenseNumber(ctx context.Context, licenseNumber string) (*domain.MiningPool, error)
	Update(ctx context.Context, pool *domain.MiningPool) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PoolStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.PoolFilter, limit, offset int) ([]domain.MiningPool, error)
	Count(ctx context.Context, filter domain.PoolFilter) (int64, error)
	IncrementMachineCount(ctx context.Context, id uuid.UUID) error
	DecrementMachineCount(ctx context.Context, id uuid.UUID) error
	UpdateEnergyUsage(ctx context.Context, id uuid.UUID, usageKW decimal.Decimal) error
	UpdateHashRate(ctx context.Context, id uuid.UUID, hashRateTH decimal.Decimal) error
}

// MachineRepository defines the interface for mining machine persistence
type MachineRepository interface {
	Create(ctx context.Context, machine *domain.MiningMachine) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.MiningMachine, error)
	GetBySerialNumber(ctx context.Context, serialNumber string) (*domain.MiningMachine, error)
	Update(ctx context.Context, machine *domain.MiningMachine) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPool(ctx context.Context, poolID uuid.UUID, activeOnly bool, limit, offset int) ([]domain.MiningMachine, error)
	CountByPool(ctx context.Context, poolID uuid.UUID, activeOnly bool) (int64, error)
	BatchCreate(ctx context.Context, machines []domain.MiningMachine) error
	GetPoolMachinesStats(ctx context.Context, poolID uuid.UUID) (*MachineStats, error)
}

// MachineStats represents statistics for machines in a pool
type MachineStats struct {
	TotalMachines    int
	ActiveMachines   int
	OfflineMachines  int
	MaintenanceCount int
}

// RegistrationService handles mining pool and machine registration
type RegistrationService struct {
	poolRepo    MiningPoolRepository
	machineRepo MachineRepository
	complianceRepo ComplianceRepository
}

// NewRegistrationService creates a new registration service
func NewRegistrationService(poolRepo MiningPoolRepository, machineRepo MachineRepository, complianceRepo ComplianceRepository) *RegistrationService {
	return &RegistrationService{
		poolRepo:    poolRepo,
		machineRepo: machineRepo,
		complianceRepo: complianceRepo,
	}
}

// RegisterMiningPool registers a new mining pool
func (s *RegistrationService) RegisterMiningPool(ctx context.Context, req *domain.PoolRegistrationRequest) (*domain.MiningPool, error) {
	// Verify entity with compliance system
	isLicensed, err := s.complianceRepo.IsEntityLicensed(ctx, req.OwnerEntityID)
	if err != nil {
		return nil, &ServiceError{
			Code:    "COMPLIANCE_CHECK_FAILED",
			Message: "Failed to verify entity with compliance system",
			Err:     err,
		}
	}

	if !isLicensed {
		return nil, &ServiceError{
			Code:    "ENTITY_NOT_LICENSED",
			Message: "Owner entity is not licensed with the compliance system",
		}
	}

	// Get entity details for the pool record
	entityDetails, err := s.complianceRepo.GetEntityDetails(ctx, req.OwnerEntityID)
	if err != nil {
		return nil, &ServiceError{
			Code:    "ENTITY_FETCH_FAILED",
			Message: "Failed to fetch entity details",
			Err:     err,
		}
	}

	// Generate license number
	licenseNumber := s.generateLicenseNumber(req.RegionCode)

	// Create the mining pool
	pool := &domain.MiningPool{
		ID:                   uuid.New(),
		LicenseNumber:        licenseNumber,
		Name:                 req.Name,
		OwnerEntityID:        req.OwnerEntityID,
		OwnerEntityName:      entityDetails.Name,
		Status:               domain.PoolStatusPending,
		RegionCode:           req.RegionCode,
		MaxAllowedEnergyKW:   req.MaxAllowedEnergyKW,
		CurrentEnergyUsageKW: decimal.Zero,
		AllowedEnergySources: req.AllowedEnergySources,
		TotalHashRateTH:      decimal.Zero,
		ActiveMachineCount:   0,
		TotalMachineCount:    0,
		CarbonFootprintKG:    decimal.Zero,
		LicenseIssuedAt:      time.Now(),
		LicenseExpiresAt:     time.Now().AddDate(1, 0, 0),
		ContactEmail:         req.ContactEmail,
		ContactPhone:         req.ContactPhone,
		FacilityAddress:      req.FacilityAddress,
		GPSCoordinates:       req.GPSCoordinates,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Validate the pool
	if err := s.validateMiningPool(pool); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.poolRepo.Create(ctx, pool); err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to create mining pool",
			Err:     err,
		}
	}

	log.Printf("Registered new mining pool: %s (%s)", pool.Name, pool.LicenseNumber)

	return pool, nil
}

// GetMiningPool retrieves a mining pool by ID
func (s *RegistrationService) GetMiningPool(ctx context.Context, id uuid.UUID) (*domain.MiningPool, error) {
	pool, err := s.poolRepo.GetByID(ctx, id)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining pool",
			Err:     err,
		}
	}

	if pool == nil {
		return nil, &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Mining pool not found",
		}
	}

	return pool, nil
}

// UpdateMiningPool updates an existing mining pool
func (s *RegistrationService) UpdateMiningPool(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*domain.MiningPool, error) {
	pool, err := s.poolRepo.GetByID(ctx, id)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining pool",
			Err:     err,
		}
	}

	if pool == nil {
		return nil, &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Mining pool not found",
		}
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		pool.Name = name
	}
	if email, ok := updates["contact_email"].(string); ok {
		pool.ContactEmail = email
	}
	if phone, ok := updates["contact_phone"].(string); ok {
		pool.ContactPhone = phone
	}
	if address, ok := updates["facility_address"].(string); ok {
		pool.FacilityAddress = address
	}
	if maxEnergy, ok := updates["max_allowed_energy_kw"].(decimal.Decimal); ok {
		pool.MaxAllowedEnergyKW = maxEnergy
	}

	pool.UpdatedAt = time.Now()

	if err := s.poolRepo.Update(ctx, pool); err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to update mining pool",
			Err:     err,
		}
	}

	return pool, nil
}

// SuspendMiningPool suspends a mining pool
func (s *RegistrationService) SuspendMiningPool(ctx context.Context, id uuid.UUID, reason string) (*domain.MiningPool, error) {
	pool, err := s.poolRepo.GetByID(ctx, id)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining pool",
			Err:     err,
		}
	}

	if pool == nil {
		return nil, &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Mining pool not found",
		}
	}

	if pool.Status == domain.PoolStatusSuspended {
		return nil, &ServiceError{
			Code:    "INVALID_STATE",
			Message: "Mining pool is already suspended",
		}
	}

	if err := s.poolRepo.UpdateStatus(ctx, id, domain.PoolStatusSuspended); err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to suspend mining pool",
			Err:     err,
		}
	}

	pool.Status = domain.PoolStatusSuspended
	log.Printf("Suspended mining pool: %s (%s). Reason: %s", pool.Name, pool.LicenseNumber, reason)

	return pool, nil
}

// ActivateMiningPool activates a suspended mining pool
func (s *RegistrationService) ActivateMiningPool(ctx context.Context, id uuid.UUID) (*domain.MiningPool, error) {
	pool, err := s.poolRepo.GetByID(ctx, id)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining pool",
			Err:     err,
		}
	}

	if pool == nil {
		return nil, &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Mining pool not found",
		}
	}

	if pool.Status != domain.PoolStatusSuspended {
		return nil, &ServiceError{
			Code:    "INVALID_STATE",
			Message: "Mining pool is not suspended",
		}
	}

	if err := s.poolRepo.UpdateStatus(ctx, id, domain.PoolStatusActive); err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to activate mining pool",
			Err:     err,
		}
	}

	pool.Status = domain.PoolStatusActive
	log.Printf("Activated mining pool: %s (%s)", pool.Name, pool.LicenseNumber)

	return pool, nil
}

// RevokeMiningPoolLicense revokes a mining pool license
func (s *RegistrationService) RevokeMiningPoolLicense(ctx context.Context, id uuid.UUID, reason string) error {
	pool, err := s.poolRepo.GetByID(ctx, id)
	if err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining pool",
			Err:     err,
		}
	}

	if pool == nil {
		return &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Mining pool not found",
		}
	}

	if err := s.poolRepo.UpdateStatus(ctx, id, domain.PoolStatusRevoked); err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to revoke mining pool license",
			Err:     err,
		}
	}

	log.Printf("Revoked mining pool license: %s (%s). Reason: %s", pool.Name, pool.LicenseNumber, reason)

	return nil
}

// RegisterMiningMachine registers a new mining machine
func (s *RegistrationService) RegisterMiningMachine(ctx context.Context, poolID uuid.UUID, req *domain.MachineRegistrationRequest) (*domain.MiningMachine, error) {
	// Verify pool exists and is active
	pool, err := s.poolRepo.GetByID(ctx, poolID)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining pool",
			Err:     err,
		}
	}

	if pool == nil {
		return nil, &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Mining pool not found",
		}
	}

	if pool.Status != domain.PoolStatusActive {
		return nil, &ServiceError{
			Code:    "INVALID_STATE",
			Message: "Mining pool is not active",
		}
	}

	// Check for duplicate serial number
	existing, err := s.machineRepo.GetBySerialNumber(ctx, req.SerialNumber)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to check for existing machine",
			Err:     err,
		}
	}

	if existing != nil {
		return nil, &ServiceError{
			Code:    "DUPLICATE_ENTRY",
			Message: "Machine with this serial number already exists",
		}
	}

	// Create the machine
	machine := &domain.MiningMachine{
		ID:                 uuid.New(),
		SerialNumber:       req.SerialNumber,
		PoolID:             poolID,
		ModelType:          req.ModelType,
		Manufacturer:       req.Manufacturer,
		ModelName:          req.ModelName,
		HashRateSpecTH:     req.HashRateSpecTH,
		PowerSpecWatts:     req.PowerSpecWatts,
		LocationID:         req.LocationID,
		LocationDescription: req.LocationDescription,
		GPSCoordinates:     req.GPSCoordinates,
		InstallationDate:   req.InstallationDate,
		Status:             domain.MachineStatusPending,
		IsActive:           false,
		CurrentHashRateTH:  decimal.Zero,
		CurrentPowerWatts:  decimal.Zero,
		TotalEnergyKWh:     decimal.Zero,
		UptimeHours:        decimal.Zero,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.machineRepo.Create(ctx, machine); err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to create mining machine",
			Err:     err,
		}
	}

	// Increment machine count for pool
	if err := s.poolRepo.IncrementMachineCount(ctx, poolID); err != nil {
		log.Printf("Warning: Failed to increment machine count for pool %s: %v", poolID, err)
	}

	log.Printf("Registered mining machine: %s for pool %s", machine.SerialNumber, pool.Name)

	return machine, nil
}

// BatchRegisterMachines registers multiple mining machines
func (s *RegistrationService) BatchRegisterMachines(ctx context.Context, poolID uuid.UUID, reqs []domain.MachineRegistrationRequest) ([]domain.MiningMachine, error) {
	// Verify pool exists and is active
	pool, err := s.poolRepo.GetByID(ctx, poolID)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining pool",
			Err:     err,
		}
	}

	if pool == nil {
		return nil, &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Mining pool not found",
		}
	}

	if pool.Status != domain.PoolStatusActive {
		return nil, &ServiceError{
			Code:    "INVALID_STATE",
			Message: "Mining pool is not active",
		}
	}

	var machines []domain.MiningMachine
	for _, req := range reqs {
		machine := &domain.MiningMachine{
			ID:                  uuid.New(),
			SerialNumber:        req.SerialNumber,
			PoolID:              poolID,
			ModelType:           req.ModelType,
			Manufacturer:        req.Manufacturer,
			ModelName:           req.ModelName,
			HashRateSpecTH:      req.HashRateSpecTH,
			PowerSpecWatts:      req.PowerSpecWatts,
			LocationID:          req.LocationID,
			LocationDescription: req.LocationDescription,
			GPSCoordinates:      req.GPSCoordinates,
			InstallationDate:    req.InstallationDate,
			Status:              domain.MachineStatusPending,
			IsActive:            false,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}
		machines = append(machines, *machine)
	}

	if err := s.machineRepo.BatchCreate(ctx, machines); err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to batch create mining machines",
			Err:     err,
		}
	}

	// Update pool machine count
	for range reqs {
		if err := s.poolRepo.IncrementMachineCount(ctx, poolID); err != nil {
			log.Printf("Warning: Failed to increment machine count for pool %s: %v", poolID, err)
		}
	}

	log.Printf("Batch registered %d machines for pool %s", len(machines), pool.Name)

	return machines, nil
}

// GetMiningMachine retrieves a mining machine by ID
func (s *RegistrationService) GetMiningMachine(ctx context.Context, id uuid.UUID) (*domain.MiningMachine, error) {
	machine, err := s.machineRepo.GetByID(ctx, id)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining machine",
			Err:     err,
		}
	}

	if machine == nil {
		return nil, &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Mining machine not found",
		}
	}

	return machine, nil
}

// GetPoolMachines retrieves all machines for a pool
func (s *RegistrationService) GetPoolMachines(ctx context.Context, poolID uuid.UUID, activeOnly bool, limit, offset int) ([]domain.MiningMachine, error) {
	machines, err := s.machineRepo.ListByPool(ctx, poolID, activeOnly, limit, offset)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve pool machines",
			Err:     err,
		}
	}

	return machines, nil
}

// UpdateMiningMachine updates a mining machine
func (s *RegistrationService) UpdateMiningMachine(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*domain.MiningMachine, error) {
	machine, err := s.machineRepo.GetByID(ctx, id)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining machine",
			Err:     err,
		}
	}

	if machine == nil {
		return nil, &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Mining machine not found",
		}
	}

	// Apply updates
	if status, ok := updates["status"].(domain.MachineStatus); ok {
		machine.Status = status
		if status == domain.MachineStatusActive {
			machine.IsActive = true
		} else if status == domain.MachineStatusOffline || status == domain.MachineStatusDecommissioned {
			machine.IsActive = false
		}
	}
	if locDesc, ok := updates["location_description"].(string); ok {
		machine.LocationDescription = locDesc
	}
	if maintDate, ok := updates["last_maintenance_date"].(time.Time); ok {
		machine.LastMaintenanceDate = &maintDate
	}

	machine.UpdatedAt = time.Now()

	if err := s.machineRepo.Update(ctx, machine); err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to update mining machine",
			Err:     err,
		}
	}

	return machine, nil
}

// DecommissionMachine decommission a mining machine
func (s *RegistrationService) DecommissionMachine(ctx context.Context, id uuid.UUID, reason string) error {
	machine, err := s.machineRepo.GetByID(ctx, id)
	if err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining machine",
			Err:     err,
		}
	}

	if machine == nil {
		return &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Mining machine not found",
		}
	}

	now := time.Now()
	machine.Status = domain.MachineStatusDecommissioned
	machine.IsActive = false
	machine.DecommissionedAt = &now
	machine.UpdatedAt = now

	if err := s.machineRepo.Update(ctx, machine); err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to decommission mining machine",
			Err:     err,
		}
	}

	// Decrement machine count for pool
	if err := s.poolRepo.DecrementMachineCount(ctx, machine.PoolID); err != nil {
		log.Printf("Warning: Failed to decrement machine count for pool %s: %v", machine.PoolID, err)
	}

	log.Printf("Decommissioned mining machine: %s. Reason: %s", machine.SerialNumber, reason)

	return nil
}

// validateMiningPool validates a mining pool
func (s *RegistrationService) validateMiningPool(pool *domain.MiningPool) error {
	if pool.Name == "" {
		return &ServiceError{
			Code:    "VALIDATION_ERROR",
			Message: "Pool name is required",
		}
	}

	if pool.RegionCode == "" {
		return &ServiceError{
			Code:    "VALIDATION_ERROR",
			Message: "Region code is required",
		}
	}

	if pool.MaxAllowedEnergyKW.LessThanOrEqual(decimal.Zero) {
		return &ServiceError{
			Code:    "VALIDATION_ERROR",
			Message: "Max allowed energy must be greater than zero",
		}
	}

	if len(pool.AllowedEnergySources) == 0 {
		return &ServiceError{
			Code:    "VALIDATION_ERROR",
			Message: "At least one energy source must be specified",
		}
	}

	return nil
}

// generateLicenseNumber generates a unique license number
func (s *RegistrationService) generateLicenseNumber(regionCode string) string {
	timestamp := time.Now().Format("20060102")
	uuidPart := uuid.New().String()[:8]
	return fmt.Sprintf("MC-%s-%s-%s", regionCode, timestamp, uuidPart)
}

import "fmt"
