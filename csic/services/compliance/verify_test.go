package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/csic-platform/services/compliance/internal/core/domain"
	"github.com/csic-platform/services/compliance/internal/core/services"
)

// MockComplianceRepository implements ComplianceRepository for testing
type MockComplianceRepository struct {
	entities    map[string]*domain.Entity
	licenses    map[string]*domain.License
	obligations map[string]*domain.Obligation
	scores      map[string]*domain.ComplianceScore
}

func NewMockComplianceRepository() *MockComplianceRepository {
	return &MockComplianceRepository{
		entities:    make(map[string]*domain.Entity),
		licenses:    make(map[string]*domain.License),
		obligations: make(map[string]*domain.Obligation),
		scores:      make(map[string]*domain.ComplianceScore),
	}
}

func (m *MockComplianceRepository) CreateEntity(ctx context.Context, entity *domain.Entity) error {
	m.entities[entity.ID.String()] = entity
	return nil
}

func (m *MockComplianceRepository) GetEntity(ctx context.Context, id string) (*domain.Entity, error) {
	if e, ok := m.entities[id]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("entity not found")
}

func (m *MockComplianceRepository) UpdateEntity(ctx context.Context, entity *domain.Entity) error {
	m.entities[entity.ID.String()] = entity
	return nil
}

func (m *MockComplianceRepository) ListEntities(ctx context.Context, status string, limit, offset int) ([]*domain.Entity, error) {
	var result []*domain.Entity
	for _, e := range m.entities {
		if status == "" || e.Status == status {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *MockComplianceRepository) CreateLicense(ctx context.Context, license *domain.License) error {
	m.licenses[license.ID.String()] = license
	return nil
}

func (m *MockComplianceRepository) GetLicense(ctx context.Context, id string) (*domain.License, error) {
	if l, ok := m.licenses[id]; ok {
		return l, nil
	}
	return nil, fmt.Errorf("license not found")
}

func (m *MockComplianceRepository) GetLicensesByEntity(ctx context.Context, entityID string) ([]*domain.License, error) {
	var result []*domain.License
	for _, l := range m.licenses {
		if l.EntityID.String() == entityID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *MockComplianceRepository) UpdateLicense(ctx context.Context, license *domain.License) error {
	m.licenses[license.ID.String()] = license
	return nil
}

func (m *MockComplianceRepository) CreateObligation(ctx context.Context, obligation *domain.Obligation) error {
	m.obligations[obligation.ID.String()] = obligation
	return nil
}

func (m *MockComplianceRepository) GetObligation(ctx context.Context, id string) (*domain.Obligation, error) {
	if o, ok := m.obligations[id]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("obligation not found")
}

func (m *MockComplianceRepository) GetObligationsByEntity(ctx context.Context, entityID string) ([]*domain.Obligation, error) {
	var result []*domain.Obligation
	for _, o := range m.obligations {
		if o.EntityID.String() == entityID {
			result = append(result, o)
		}
	}
	return result, nil
}

func (m *MockComplianceRepository) UpdateObligation(ctx context.Context, obligation *domain.Obligation) error {
	m.obligations[obligation.ID.String()] = obligation
	return nil
}

func (m *MockComplianceRepository) GetOverdueObligations(ctx context.Context) ([]*domain.Obligation, error) {
	var result []*domain.Obligation
	now := time.Now()
	for _, o := range m.obligations {
		if o.DueDate.Before(now) && o.Status != "FULFILLED" {
			result = append(result, o)
		}
	}
	return result, nil
}

func (m *MockComplianceRepository) CreateComplianceScore(ctx context.Context, score *domain.ComplianceScore) error {
	m.scores[score.ID.String()] = score
	return nil
}

func (m *MockComplianceRepository) GetLatestComplianceScore(ctx context.Context, entityID string) (*domain.ComplianceScore, error) {
	for _, s := range m.scores {
		if s.EntityID.String() == entityID {
			return s, nil
		}
	}
	return nil, fmt.Errorf("score not found")
}

func (m *MockComplianceRepository) CreateAuditLog(ctx context.Context, log *domain.AuditRecord) error {
	return nil
}

func (m *MockComplianceRepository) GetAuditLogs(ctx context.Context, entityID string, limit int) ([]*domain.AuditRecord, error) {
	return nil, nil
}

func main() {
	fmt.Println("=== Compliance Service Verification ===")
	fmt.Println()

	// Create mock repository and services
	repo := NewMockComplianceRepository()
	licenseSvc := services.NewLicenseService(repo, nil)
	complianceSvc := services.NewComplianceService(repo, nil)
	obligationSvc := services.NewObligationService(repo, nil)

	ctx := context.Background()

	// Test 1: Register a new entity
	fmt.Println("Test 1: Registering new entity...")
	entity := &domain.Entity{
		Name:            "Test Exchange LLC",
		LegalName:       "Test Exchange Limited Liability Company",
		RegistrationNum: "TEST-12345",
		Jurisdiction:    "US",
		EntityType:      "EXCHANGE",
		Status:          "ACTIVE",
		RiskLevel:       "MEDIUM",
		CreatedAt:       time.Now(),
	}

	err := licenseSvc.RegisterEntity(ctx, entity)
	if err != nil {
		fmt.Printf("FAIL: Failed to register entity: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PASS: Entity registered with ID %s\n", entity.ID.String())
	fmt.Println()

	// Test 2: Issue a license
	fmt.Println("Test 2: Issuing license...")
	license := &domain.License{
		EntityID:      entity.ID,
		Type:          "EXCHANGE",
		LicenseNumber: fmt.Sprintf("LIC-%d", time.Now().Unix()),
		Status:        "ACTIVE",
		IssuedDate:    time.Now(),
		ExpiryDate:    time.Now().AddDate(1, 0, 0),
		IssuedBy:      "Test Authority",
	}

	err = licenseSvc.IssueLicense(ctx, license)
	if err != nil {
		fmt.Printf("FAIL: Failed to issue license: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PASS: License issued with ID %s\n", license.ID.String())
	fmt.Println()

	// Test 3: Create an obligation
	fmt.Println("Test 3: Creating obligation...")
	obligation := &domain.Obligation{
		EntityID:   entity.ID,
		Type:       "REPORTING",
		Description: "Submit quarterly compliance report",
		DueDate:    time.Now().AddDate(0, 3, 0),
		Status:     "PENDING",
		CreatedAt:  time.Now(),
	}

	err = obligationSvc.CreateObligation(ctx, obligation)
	if err != nil {
		fmt.Printf("FAIL: Failed to create obligation: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PASS: Obligation created with ID %s\n", obligation.ID.String())
	fmt.Println()

	// Test 4: Calculate compliance score
	fmt.Println("Test 4: Calculating compliance score...")
	score, err := complianceSvc.CalculateScore(ctx, entity.ID.String())
	if err != nil {
		fmt.Printf("FAIL: Failed to calculate score: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PASS: Compliance score calculated: %.2f (Tier: %s)\n", score.TotalScore, score.Tier)
	fmt.Println()

	// Test 5: Get entity licenses
	fmt.Println("Test 5: Getting entity licenses...")
	licenses, err := licenseSvc.GetEntityLicenses(ctx, entity.ID.String())
	if err != nil {
		fmt.Printf("FAIL: Failed to get licenses: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PASS: Found %d license(s) for entity\n", len(licenses))
	fmt.Println()

	// Test 6: Get overdue obligations
	fmt.Println("Test 6: Checking for overdue obligations...")
	overdue, err := obligationSvc.GetOverdueObligations(ctx)
	if err != nil {
		fmt.Printf("FAIL: Failed to check overdue obligations: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PASS: Found %d overdue obligation(s)\n", len(overdue))
	fmt.Println()

	fmt.Println("=== All Compliance Service Tests Passed ===")
}
