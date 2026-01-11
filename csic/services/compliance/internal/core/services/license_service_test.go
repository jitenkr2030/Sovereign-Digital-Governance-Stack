package services

import (
	"context"
	"testing"
	"time"

	"github.com/csic-platform/services/services/compliance/internal/core/domain"
	"github.com/csic-platform/services/services/compliance/internal/core/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MockRepository implements all compliance repository interfaces for testing
type MockRepository struct {
	entities      map[uuid.UUID]*domain.Entity
	applications  map[uuid.UUID]*domain.LicenseApplication
	licenses      map[uuid.UUID]*domain.License
	obligations   map[uuid.UUID]*domain.Obligation
	scores        map[uuid.UUID][]domain.ComplianceScore
	auditRecords  map[uuid.UUID]*domain.AuditRecord
	regulations   map[uuid.UUID]*domain.Regulation
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		entities:     make(map[uuid.UUID]*domain.Entity),
		applications: make(map[uuid.UUID]*domain.LicenseApplication),
		licenses:     make(map[uuid.UUID]*domain.License),
		obligations:  make(map[uuid.UUID]*domain.Obligation),
		scores:       make(map[uuid.UUID][]domain.ComplianceScore),
		auditRecords: make(map[uuid.UUID]*domain.AuditRecord),
		regulations:  make(map[uuid.UUID]*domain.Regulation),
	}
}

// License Repository Methods

func (m *MockRepository) CreateLicense(ctx context.Context, license *domain.License) error {
	m.licenses[license.ID] = license
	return nil
}

func (m *MockRepository) GetLicense(ctx context.Context, id uuid.UUID) (*domain.License, error) {
	return m.licenses[id], nil
}

func (m *MockRepository) GetLicenseByNumber(ctx context.Context, number string) (*domain.License, error) {
	for _, lic := range m.licenses {
		if lic.LicenseNumber == number {
			return lic, nil
		}
	}
	return nil, nil
}

func (m *MockRepository) GetLicensesByEntity(ctx context.Context, entityID uuid.UUID) ([]domain.License, error) {
	var licenses []domain.License
	for _, lic := range m.licenses {
		if lic.EntityID == entityID {
			licenses = append(licenses, *lic)
		}
	}
	return licenses, nil
}

func (m *MockRepository) UpdateLicense(ctx context.Context, license *domain.License) error {
	m.licenses[license.ID] = license
	return nil
}

func (m *MockRepository) UpdateLicenseStatus(ctx context.Context, id uuid.UUID, status domain.LicenseStatus) error {
	if lic, ok := m.licenses[id]; ok {
		lic.Status = status
	}
	return nil
}

func (m *MockRepository) GetLicensesExpiringSoon(ctx context.Context, days int) ([]domain.License, error) {
	var licenses []domain.License
	threshold := time.Now().AddDate(0, 0, days)
	for _, lic := range m.licenses {
		if lic.ExpiryDate.Before(threshold) && lic.Status == domain.LicenseStatusActive {
			licenses = append(licenses, *lic)
		}
	}
	return licenses, nil
}

func (m *MockRepository) GetActiveLicenses(ctx context.Context) ([]domain.License, error) {
	var licenses []domain.License
	for _, lic := range m.licenses {
		if lic.Status == domain.LicenseStatusActive {
			licenses = append(licenses, *lic)
		}
	}
	return licenses, nil
}

// Application Repository Methods

func (m *MockRepository) CreateApplication(ctx context.Context, app *domain.LicenseApplication) error {
	m.applications[app.ID] = app
	return nil
}

func (m *MockRepository) GetApplication(ctx context.Context, id uuid.UUID) (*domain.LicenseApplication, error) {
	return m.applications[id], nil
}

func (m *MockRepository) GetApplicationsByEntity(ctx context.Context, entityID uuid.UUID) ([]domain.LicenseApplication, error) {
	var apps []domain.LicenseApplication
	for _, app := range m.applications {
		if app.EntityID == entityID {
			apps = append(apps, *app)
		}
	}
	return apps, nil
}

func (m *MockRepository) GetApplicationsByStatus(ctx context.Context, status domain.ApplicationStatus) ([]domain.LicenseApplication, error) {
	var apps []domain.LicenseApplication
	for _, app := range m.applications {
		if app.Status == status {
			apps = append(apps, *app)
		}
	}
	return apps, nil
}

func (m *MockRepository) UpdateApplication(ctx context.Context, app *domain.LicenseApplication) error {
	m.applications[app.ID] = app
	return nil
}

func (m *MockRepository) UpdateApplicationStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, reviewerID *uuid.UUID, notes string) error {
	if app, ok := m.applications[id]; ok {
		app.Status = status
		if reviewerID != nil {
			app.ReviewerID = reviewerID
		}
		app.ReviewerNotes = notes
		now := time.Now().UTC()
		app.ReviewedAt = &now
	}
	return nil
}

// Entity Repository Methods

func (m *MockRepository) CreateEntity(ctx context.Context, entity *domain.Entity) error {
	m.entities[entity.ID] = entity
	return nil
}

func (m *MockRepository) GetEntity(ctx context.Context, id uuid.UUID) (*domain.Entity, error) {
	return m.entities[id], nil
}

func (m *MockRepository) GetEntityByRegistration(ctx context.Context, regNum string) (*domain.Entity, error) {
	for _, entity := range m.entities {
		if entity.RegistrationNum == regNum {
			return entity, nil
		}
	}
	return nil, nil
}

func (m *MockRepository) UpdateEntity(ctx context.Context, entity *domain.Entity) error {
	m.entities[entity.ID] = entity
	return nil
}

// Regulation Repository Methods

func (m *MockRepository) CreateRegulation(ctx context.Context, reg *domain.Regulation) error {
	m.regulations[reg.ID] = reg
	return nil
}

func (m *MockRepository) GetRegulation(ctx context.Context, id uuid.UUID) (*domain.Regulation, error) {
	return m.regulations[id], nil
}

func (m *MockRepository) GetRegulationsByCategory(ctx context.Context, category string) ([]domain.Regulation, error) {
	var regs []domain.Regulation
	for _, reg := range m.regulations {
		if reg.Category == category {
			regs = append(regs, *reg)
		}
	}
	return regs, nil
}

// Compliance Repository Methods

func (m *MockRepository) CreateScore(ctx context.Context, score *domain.ComplianceScore) error {
	m.scores[score.EntityID] = append(m.scores[score.EntityID], *score)
	return nil
}

func (m *MockRepository) GetCurrentScore(ctx context.Context, entityID uuid.UUID) (*domain.ComplianceScore, error) {
	scores := m.scores[entityID]
	if len(scores) > 0 {
		return &scores[len(scores)-1], nil
	}
	return nil, nil
}

func (m *MockRepository) GetScoreHistory(ctx context.Context, entityID uuid.UUID, limit int) ([]domain.ComplianceScore, error) {
	return m.scores[entityID], nil
}

func (m *MockRepository) GetScoresByTier(ctx context.Context, tier domain.ComplianceTier) ([]domain.ComplianceScore, error) {
	var scores []domain.ComplianceScore
	for _, entityScores := range m.scores {
		for _, s := range entityScores {
			if s.Tier == tier {
				scores = append(scores, s)
			}
		}
	}
	return scores, nil
}

func (m *MockRepository) CalculateAverageScore(ctx context.Context) (float64, error) {
	return 75.0, nil
}

// Obligation Repository Methods

func (m *MockRepository) CreateObligation(ctx context.Context, obligation *domain.Obligation) error {
	m.obligations[obligation.ID] = obligation
	return nil
}

func (m *MockRepository) GetObligation(ctx context.Context, id uuid.UUID) (*domain.Obligation, error) {
	return m.obligations[id], nil
}

func (m *MockRepository) GetObligationsByEntity(ctx context.Context, entityID uuid.UUID) ([]domain.Obligation, error) {
	var obs []domain.Obligation
	for _, ob := range m.obligations {
		if ob.EntityID == entityID {
			obs = append(obs, *ob)
		}
	}
	return obs, nil
}

func (m *MockRepository) GetObligationsByStatus(ctx context.Context, status domain.ObligationStatus) ([]domain.Obligation, error) {
	var obs []domain.Obligation
	for _, ob := range m.obligations {
		if ob.Status == status {
			obs = append(obs, *ob)
		}
	}
	return obs, nil
}

func (m *MockRepository) GetOverdueObligations(ctx context.Context) ([]domain.Obligation, error) {
	var obs []domain.Obligation
	now := time.Now()
	for _, ob := range m.obligations {
		if ob.Status == domain.ObligationPending && ob.DueDate.Before(now) {
			obs = append(obs, *ob)
		}
	}
	return obs, nil
}

func (m *MockRepository) UpdateObligation(ctx context.Context, obligation *domain.Obligation) error {
	m.obligations[obligation.ID] = obligation
	return nil
}

func (m *MockRepository) UpdateObligationStatus(ctx context.Context, id uuid.UUID, status domain.ObligationStatus) error {
	if ob, ok := m.obligations[id]; ok {
		ob.Status = status
	}
	return nil
}

func (m *MockRepository) MarkObligationFulfilled(ctx context.Context, id uuid.UUID, evidence string) error {
	if ob, ok := m.obligations[id]; ok {
		ob.Status = domain.ObligationFulfilled
		now := time.Now().UTC()
		ob.FulfilledAt = &now
		ob.FulfilledEvidence = evidence
	}
	return nil
}

func (m *MockRepository) GetUpcomingObligations(ctx context.Context, days int) ([]domain.Obligation, error) {
	var obs []domain.Obligation
	threshold := time.Now().AddDate(0, 0, days)
	for _, ob := range m.obligations {
		if ob.Status == domain.ObligationPending && ob.DueDate.Before(threshold) {
			obs = append(obs, *ob)
		}
	}
	return obs, nil
}

// Audit Repository Methods

func (m *MockRepository) CreateAuditRecord(ctx context.Context, record *domain.AuditRecord) error {
	m.auditRecords[record.ID] = record
	return nil
}

func (m *MockRepository) GetAuditRecord(ctx context.Context, id uuid.UUID) (*domain.AuditRecord, error) {
	return m.auditRecords[id], nil
}

func (m *MockRepository) GetAuditRecordsByEntity(ctx context.Context, entityID uuid.UUID, limit, offset int) ([]domain.AuditRecord, error) {
	var records []domain.AuditRecord
	count := 0
	for _, record := range m.auditRecords {
		if record.EntityID == entityID {
			if count >= offset && (limit == 0 || len(records) < limit) {
				records = append(records, *record)
			}
			count++
		}
	}
	return records, nil
}

func (m *MockRepository) GetAuditRecordsByResource(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]domain.AuditRecord, error) {
	var records []domain.AuditRecord
	for _, record := range m.auditRecords {
		if record.ResourceType == resourceType && record.ResourceID == resourceID {
			records = append(records, *record)
		}
	}
	return records, nil
}

func (m *MockRepository) GetAuditRecordsByActor(ctx context.Context, actorID uuid.UUID, limit int) ([]domain.AuditRecord, error) {
	var records []domain.AuditRecord
	for _, record := range m.auditRecords {
		if record.ActorID == actorID {
			records = append(records, *record)
		}
	}
	return records, nil
}

func (m *MockRepository) GetAuditRecordsByDateRange(ctx context.Context, from, to time.Time, limit, offset int) ([]domain.AuditRecord, error) {
	var records []domain.AuditRecord
	count := 0
	for _, record := range m.auditRecords {
		if (record.Timestamp.Equal(from) || record.Timestamp.After(from)) &&
			(record.Timestamp.Equal(to) || record.Timestamp.Before(to)) {
			if count >= offset && (limit == 0 || len(records) < limit) {
				records = append(records, *record)
			}
			count++
		}
	}
	return records, nil
}

func (m *MockRepository) GetAuditRecords(ctx context.Context, filter ports.AuditFilter) ([]domain.AuditRecord, error) {
	var records []domain.AuditRecord
	for _, record := range m.auditRecords {
		if filter.EntityID != nil && record.EntityID != *filter.EntityID {
			continue
		}
		if filter.ActorID != nil && record.ActorID != *filter.ActorID {
			continue
		}
		if filter.ResourceType != "" && record.ResourceType != filter.ResourceType {
			continue
		}
		records = append(records, *record)
	}
	return records, nil
}

func (m *MockRepository) CountAuditRecords(ctx context.Context, filter ports.AuditFilter) (int64, error) {
	records, _ := m.GetAuditRecords(ctx, filter)
	return int64(len(records)), nil
}

// Test setup helper

func setupLicenseTestService() (*LicenseService, *MockRepository) {
	log, _ := zap.NewDevelopment()
	repo := NewMockRepository()
	service := NewLicenseService(repo, log)
	return service, repo
}

func createTestEntity() *domain.Entity {
	return &domain.Entity{
		ID:              uuid.New(),
		Name:            "Test Entity",
		LegalName:       "Test Entity LLC",
		RegistrationNum: "TEST-2024-001",
		Jurisdiction:    "US",
		EntityType:      "EXCHANGE",
		Address:         "123 Test St",
		ContactEmail:    "test@test.com",
		Status:          "ACTIVE",
		RiskLevel:       "MEDIUM",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func createTestApplication(entityID uuid.UUID) *domain.LicenseApplication {
	return &domain.LicenseApplication{
		ID:             uuid.New(),
		EntityID:       entityID,
		Type:           domain.AppTypeNew,
		LicenseType:    domain.LicenseTypeExchange,
		Status:         domain.AppStatusSubmitted,
		SubmittedAt:    time.Now(),
		RequestedTerms: "Exchange license for trading",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func createTestLicense(entityID uuid.UUID) *domain.License {
	return &domain.License{
		ID:            uuid.New(),
		EntityID:      entityID,
		Type:          domain.LicenseTypeExchange,
		Status:        domain.LicenseStatusActive,
		LicenseNumber: "LCC-2024-000001",
		IssuedDate:    time.Now(),
		ExpiryDate:    time.Now().AddDate(1, 0, 0),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// Unit Tests

func TestSubmitApplication_Success(t *testing.T) {
	service, repo := setupLicenseTestService()
	entity := createTestEntity()
	repo.entities[entity.ID] = entity

	req := ports.SubmitApplicationRequest{
		EntityID:       entity.ID,
		Type:           domain.AppTypeNew,
		LicenseType:    domain.LicenseTypeExchange,
		RequestedTerms: "Exchange license",
		Documents:      []string{"doc1.pdf"},
	}

	app, err := service.SubmitApplication(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if app.Status != domain.AppStatusSubmitted {
		t.Errorf("Expected status SUBMITTED, got: %s", app.Status)
	}

	if _, exists := repo.applications[app.ID]; !exists {
		t.Error("Expected application to be stored")
	}
}

func TestSubmitApplication_EntityNotFound(t *testing.T) {
	service, _ := setupLicenseTestService()

	req := ports.SubmitApplicationRequest{
		EntityID:       uuid.New(),
		Type:           domain.AppTypeNew,
		LicenseType:    domain.LicenseTypeExchange,
		RequestedTerms: "Exchange license",
	}

	_, err := service.SubmitApplication(context.Background(), req)
	if err == nil {
		t.Error("Expected error for non-existent entity")
	}
}

func TestReviewApplication_Approve(t *testing.T) {
	service, repo := setupLicenseTestService()
	entity := createTestEntity()
	repo.entities[entity.ID] = entity
	app := createTestApplication(entity.ID)
	repo.applications[app.ID] = app

	reviewerID := uuid.New()
	req := ports.ApplicationReviewRequest{
		Approved:   true,
		ReviewerID: reviewerID,
		Notes:      "All requirements met",
		GrantedTerms: "Full exchange license",
		Conditions: "Standard conditions",
	}

	license, err := service.ReviewApplication(context.Background(), app.ID, req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if license == nil {
		t.Fatal("Expected license to be returned on approval")
	}

	if repo.applications[app.ID].Status != domain.AppStatusApproved {
		t.Errorf("Expected application status APPROVED, got: %s", repo.applications[app.ID].Status)
	}
}

func TestReviewApplication_Reject(t *testing.T) {
	service, repo := setupLicenseTestService()
	entity := createTestEntity()
	repo.entities[entity.ID] = entity
	app := createTestApplication(entity.ID)
	repo.applications[app.ID] = app

	reviewerID := uuid.New()
	req := ports.ApplicationReviewRequest{
		Approved:   false,
		ReviewerID: reviewerID,
		Notes:      "Missing documentation",
	}

	license, err := service.ReviewApplication(context.Background(), app.ID, req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if license != nil {
		t.Error("Expected nil license on rejection")
	}

	if repo.applications[app.ID].Status != domain.AppStatusRejected {
		t.Errorf("Expected application status REJECTED, got: %s", repo.applications[app.ID].Status)
	}
}

func TestIssueLicense_Success(t *testing.T) {
	service, repo := setupLicenseTestService()
	entity := createTestEntity()
	repo.entities[entity.ID] = entity

	req := ports.IssueLicenseRequest{
		EntityID:      entity.ID,
		LicenseType:   domain.LicenseTypeExchange,
		LicenseNumber: "LCC-2024-000002",
		ExpiryDays:    365,
		Jurisdiction:  "US",
		IssuedBy:      "Admin",
	}

	license, err := service.IssueLicense(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if license.Status != domain.LicenseStatusActive {
		t.Errorf("Expected status ACTIVE, got: %s", license.Status)
	}
}

func TestSuspendLicense_Success(t *testing.T) {
	service, repo := setupLicenseTestService()
	entity := createTestEntity()
	repo.entities[entity.ID] = entity
	license := createTestLicense(entity.ID)
	repo.licenses[license.ID] = license

	err := service.SuspendLicense(context.Background(), license.ID, "Compliance investigation")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if repo.licenses[license.ID].Status != domain.LicenseStatusSuspended {
		t.Errorf("Expected status SUSPENDED, got: %s", repo.licenses[license.ID].Status)
	}
}

func TestSuspendLicense_AlreadySuspended(t *testing.T) {
	service, repo := setupLicenseTestService()
	entity := createTestEntity()
	repo.entities[entity.ID] = entity
	license := createTestLicense(entity.ID)
	license.Status = domain.LicenseStatusSuspended
	repo.licenses[license.ID] = license

	err := service.SuspendLicense(context.Background(), license.ID, "Another reason")
	if err == nil {
		t.Error("Expected error for already suspended license")
	}
}

func TestRegisterEntity_Success(t *testing.T) {
	service, repo := setupLicenseTestService()

	req := ports.RegisterEntityRequest{
		Name:            "New Entity",
		LegalName:       "New Entity LLC",
		RegistrationNum: "NEW-2024-001",
		Jurisdiction:    "US",
		EntityType:      "EXCHANGE",
		Address:         "456 New St",
		ContactEmail:    "new@entity.com",
		RiskLevel:       "LOW",
	}

	entity, err := service.RegisterEntity(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if entity.RiskLevel != "LOW" {
		t.Errorf("Expected risk level LOW, got: %s", entity.RiskLevel)
	}
}

func TestRegisterEntity_DuplicateRegistration(t *testing.T) {
	service, repo := setupLicenseTestService()
	entity := createTestEntity()
	repo.entities[entity.ID] = entity

	req := ports.RegisterEntityRequest{
		Name:            "Duplicate",
		LegalName:       "Duplicate LLC",
		RegistrationNum: entity.RegistrationNum, // Same registration number
		Jurisdiction:    "US",
		EntityType:      "EXCHANGE",
		Address:         "123 Test St",
		ContactEmail:    "duplicate@test.com",
	}

	_, err := service.RegisterEntity(context.Background(), req)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}
}
