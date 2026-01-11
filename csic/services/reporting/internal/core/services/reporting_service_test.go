package services

import (
	"context"
	"testing"
	"time"

	"github.com/reporting-service/reporting/internal/core/domain"
	"github.com/reporting-service/reporting/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSARRepository is a mock implementation of SARRepository.
type MockSARRepository struct {
	mock.Mock
}

func (m *MockSARRepository) Create(ctx context.Context, sar *domain.SAR) error {
	args := m.Called(ctx, sar)
	return args.Error(0)
}

func (m *MockSARRepository) GetByID(ctx context.Context, id string) (*domain.SAR, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SAR), args.Error(1)
}

func (m *MockSARRepository) GetByReportNumber(ctx context.Context, number string) (*domain.SAR, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SAR), args.Error(1)
}

func (m *MockSARRepository) Update(ctx context.Context, sar *domain.SAR) error {
	args := m.Called(ctx, sar)
	return args.Error(0)
}

func (m *MockSARRepository) List(ctx context.Context, filter ports.SARFilter) ([]*domain.SAR, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.SAR), args.Error(1)
}

func (m *MockSARRepository) Count(ctx context.Context, filter ports.SARFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockSARRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockCTRRepository is a mock implementation of CTRRepository.
type MockCTRRepository struct {
	mock.Mock
}

func (m *MockCTRRepository) Create(ctx context.Context, ctr *domain.CTR) error {
	args := m.Called(ctx, ctr)
	return args.Error(0)
}

func (m *MockCTRRepository) GetByID(ctx context.Context, id string) (*domain.CTR, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CTR), args.Error(1)
}

func (m *MockCTRRepository) GetByReportNumber(ctx context.Context, number string) (*domain.CTR, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CTR), args.Error(1)
}

func (m *MockCTRRepository) Update(ctx context.Context, ctr *domain.CTR) error {
	args := m.Called(ctx, ctr)
	return args.Error(0)
}

func (m *MockCTRRepository) List(ctx context.Context, filter ports.CTRFilter) ([]*domain.CTR, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CTR), args.Error(1)
}

func (m *MockCTRRepository) Count(ctx context.Context, filter ports.CTRFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockCTRRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockComplianceRuleRepository is a mock implementation of ComplianceRuleRepository.
type MockComplianceRuleRepository struct {
	mock.Mock
}

func (m *MockComplianceRuleRepository) Create(ctx context.Context, rule *domain.ComplianceRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockComplianceRuleRepository) GetByID(ctx context.Context, id string) (*domain.ComplianceRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ComplianceRule), args.Error(1)
}

func (m *MockComplianceRuleRepository) GetByCode(ctx context.Context, code string) (*domain.ComplianceRule, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ComplianceRule), args.Error(1)
}

func (m *MockComplianceRuleRepository) Update(ctx context.Context, rule *domain.ComplianceRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockComplianceRuleRepository) List(ctx context.Context, filter ports.ComplianceRuleFilter) ([]*domain.ComplianceRule, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ComplianceRule), args.Error(1)
}

func (m *MockComplianceRuleRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockAlertRepository is a mock implementation of AlertRepository.
type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) Create(ctx context.Context, alert *domain.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) GetByID(ctx context.Context, id string) (*domain.Alert, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetByAlertNumber(ctx context.Context, number string) (*domain.Alert, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Alert), args.Error(1)
}

func (m *MockAlertRepository) Update(ctx context.Context, alert *domain.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) List(ctx context.Context, filter ports.AlertFilter) ([]*domain.Alert, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Alert), args.Error(1)
}

func (m *MockAlertRepository) Count(ctx context.Context, filter ports.AlertFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockAlertRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockFilingRecordRepository is a mock implementation of FilingRecordRepository.
type MockFilingRecordRepository struct {
	mock.Mock
}

func (m *MockFilingRecordRepository) Create(ctx context.Context, filing *domain.FilingRecord) error {
	args := m.Called(ctx, filing)
	return args.Error(0)
}

func (m *MockFilingRecordRepository) GetByID(ctx context.Context, id string) (*domain.FilingRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FilingRecord), args.Error(1)
}

func (m *MockFilingRecordRepository) GetByReportID(ctx context.Context, reportID string) ([]*domain.FilingRecord, error) {
	args := m.Called(ctx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FilingRecord), args.Error(1)
}

func (m *MockFilingRecordRepository) Update(ctx context.Context, filing *domain.FilingRecord) error {
	args := m.Called(ctx, filing)
	return args.Error(0)
}

func (m *MockFilingRecordRepository) List(ctx context.Context, filter ports.FilingRecordFilter) ([]*domain.FilingRecord, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FilingRecord), args.Error(1)
}

func (m *MockFilingRecordRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockTransactionRepository is a mock implementation of TransactionRepository.
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) GetByID(ctx context.Context, id string) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

func (m *MockTransactionRepository) GetByAccount(ctx context.Context, accountID string, startDate, endDate time.Time) ([]interface{}, error) {
	args := m.Called(ctx, accountID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockTransactionRepository) GetByUser(ctx context.Context, userID string, startDate, endDate time.Time) ([]interface{}, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockTransactionRepository) GetHighRiskTransactions(ctx context.Context, startDate, endDate time.Time, limit int) ([]interface{}, error) {
	args := m.Called(ctx, startDate, endDate, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockTransactionRepository) GetTransactionPatterns(ctx context.Context, subjectID string) (interface{}, error) {
	args := m.Called(ctx, subjectID)
	return args.Get(0), args.Error(1)
}

// MockScreeningRepository is a mock implementation of ScreeningRepository.
type MockScreeningRepository struct {
	mock.Mock
}

func (m *MockScreeningRepository) Search(ctx context.Context, query string, lists []string) ([]interface{}, error) {
	args := m.Called(ctx, query, lists)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockScreeningRepository) GetByList(ctx context.Context, listName string) ([]interface{}, error) {
	args := m.Called(ctx, listName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockScreeningRepository) AddToList(ctx context.Context, listName string, entity interface{}) error {
	args := m.Called(ctx, listName, entity)
	return args.Error(0)
}

func (m *MockScreeningRepository) RemoveFromList(ctx context.Context, listName string, entityID string) error {
	args := m.Called(ctx, listName, entityID)
	return args.Error(0)
}

// Helper function to create a test SAR.
func createTestSAR() *domain.SAR {
	return &domain.SAR{
		SubjectID:           "subject-001",
		SubjectType:         "individual",
		SubjectName:         "John Doe",
		SuspiciousActivity:  "Unusual wire transfer patterns",
		ActivityDate:        time.Now().UTC(),
		DollarAmount:        50000.00,
		TransactionCount:    5,
		Narrative:           "Multiple wire transfers to high-risk jurisdictions",
		RiskIndicators:      []string{"structuring", "high-risk jurisdiction"},
		FilingInstitution:   "Test Bank",
		ReporterID:          "reporter-001",
		FieldOffice:         "New York",
		OriginatingAgency:   "FBI",
	}
}

// Helper function to create a test CTR.
func createTestCTR() *domain.CTR {
	return &domain.CTR{
		PersonName:       "Jane Smith",
		PersonIDType:     "Passport",
		PersonIDNumber:   "AB1234567",
		PersonAddress:    domain.Address{Street1: "123 Main St", City: "New York", State: "NY", PostalCode: "10001", Country: "USA"},
		AccountNumber:    "ACC-001",
		TransactionType:  "deposit",
		CashInAmount:     15000.00,
		CashOutAmount:    0,
		TotalAmount:      15000.00,
		Currency:         "USD",
		TransactionDate:  time.Now().UTC(),
		MethodReceived:   "in_person",
		InstitutionID:    "INST-001",
		BranchID:         "BR-001",
		TellerID:         "TELLER-001",
	}
}

// Helper function to create a test alert.
func createTestAlert() *domain.Alert {
	return &domain.Alert{
		AlertType:      "sar_filing",
		Severity:       "high",
		SubjectID:      "subject-001",
		SubjectType:    "individual",
		SubjectName:    "John Doe",
		TransactionIDs: []string{"tx-001", "tx-002"},
		Description:    "Multiple high-value transactions detected",
		TriggeredRules: []string{"AML-001", "AML-002"},
		RiskScore:      8,
	}
}

func TestReportingService_CreateSAR_Success(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	sar := createTestSAR()
	sarRepo.On("Create", ctx, mock.AnythingOfType("*domain.SAR")).Return(nil)

	err := service.CreateSAR(ctx, sar)

	assert.NoError(t, err)
	assert.NotEmpty(t, sar.ReportNumber)
	assert.Equal(t, domain.ReportStatusDraft, sar.Status)
	sarRepo.AssertExpectations(t)
}

func TestReportingService_GetSAR_Success(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	expectedSAR := createTestSAR()
	sarRepo.On("GetByID", ctx, "sar-001").Return(expectedSAR, nil)

	sar, err := service.GetSAR(ctx, "sar-001")

	assert.NoError(t, err)
	assert.NotNil(t, sar)
	assert.Equal(t, "John Doe", sar.SubjectName)
	sarRepo.AssertExpectations(t)
}

func TestReportingService_GetSAR_NotFound(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	sarRepo.On("GetByID", ctx, "nonexistent").Return(nil, nil)

	_, err := service.GetSAR(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Equal(t, ErrReportNotFound, err)
	sarRepo.AssertExpectations(t)
}

func TestReportingService_UpdateSARStatus_ValidTransition(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	sar := createTestSAR()
	sar.Status = domain.ReportStatusDraft
	sarRepo.On("GetByID", ctx, "sar-001").Return(sar, nil)
	sarRepo.On("Update", ctx, mock.AnythingOfType("*domain.SAR")).Return(nil)

	err := service.UpdateSARStatus(ctx, "sar-001", domain.ReportStatusPending, "reviewer-001")

	assert.NoError(t, err)
	assert.Equal(t, domain.ReportStatusPending, sar.Status)
	assert.Equal(t, "reviewer-001", sar.ReviewerID)
	sarRepo.AssertExpectations(t)
}

func TestReportingService_UpdateSARStatus_InvalidTransition(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	sar := createTestSAR()
	sar.Status = domain.ReportStatusDraft
	sarRepo.On("GetByID", ctx, "sar-001").Return(sar, nil)

	err := service.UpdateSARStatus(ctx, "sar-001", domain.ReportStatusSubmitted, "reviewer-001")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidReportStatus, err)
	sarRepo.AssertExpectations(t)
}

func TestReportingService_SubmitSAR_Success(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	sar := createTestSAR()
	sar.Status = domain.ReportStatusApproved
	sarRepo.On("GetByID", ctx, "sar-001").Return(sar, nil)
	filingRepo.On("Create", ctx, mock.AnythingOfType("*domain.FilingRecord")).Return(nil)
	sarRepo.On("Update", ctx, mock.AnythingOfType("*domain.SAR")).Return(nil)

	err := service.SubmitSAR(ctx, "sar-001", "electronic", "FinCEN")

	assert.NoError(t, err)
	assert.Equal(t, domain.ReportStatusSubmitted, sar.Status)
	assert.NotNil(t, sar.SubmittedAt)
	sarRepo.AssertExpectations(t)
	filingRepo.AssertExpectations(t)
}

func TestReportingService_SubmitSAR_NotApproved(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	sar := createTestSAR()
	sar.Status = domain.ReportStatusDraft
	sarRepo.On("GetByID", ctx, "sar-001").Return(sar, nil)

	err := service.SubmitSAR(ctx, "sar-001", "electronic", "FinCEN")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be approved")
	sarRepo.AssertExpectations(t)
}

func TestReportingService_CreateCTR_Success(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	ctr := createTestCTR()
	ctrRepo.On("Create", ctx, mock.AnythingOfType("*domain.CTR")).Return(nil)

	err := service.CreateCTR(ctx, ctr)

	assert.NoError(t, err)
	assert.NotEmpty(t, ctr.ReportNumber)
	assert.Equal(t, domain.ReportStatusDraft, ctr.Status)
	ctrRepo.AssertExpectations(t)
}

func TestReportingService_CreateComplianceRule_Success(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	rule := &domain.ComplianceRule{
		RuleCode:      "AML-001",
		RuleName:      "Large Transaction Monitoring",
		Category:      "aml",
		Regulation:    "BSA",
		RiskScore:     8,
		Severity:      "high",
		IsActive:      true,
		EffectiveDate: time.Now().UTC(),
	}

	ruleRepo.On("GetByCode", ctx, "AML-001").Return(nil, nil)
	ruleRepo.On("Create", ctx, mock.AnythingOfType("*domain.ComplianceRule")).Return(nil)

	err := service.CreateComplianceRule(ctx, rule)

	assert.NoError(t, err)
	ruleRepo.AssertExpectations(t)
}

func TestReportingService_CreateComplianceRule_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	existingRule := &domain.ComplianceRule{RuleCode: "AML-001"}
	ruleRepo.On("GetByCode", ctx, "AML-001").Return(existingRule, nil)

	rule := &domain.ComplianceRule{RuleCode: "AML-001", RuleName: "New Rule"}
	err := service.CreateComplianceRule(ctx, rule)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	ruleRepo.AssertExpectations(t)
}

func TestReportingService_CreateAlert_Success(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	alert := createTestAlert()
	alertRepo.On("Create", ctx, mock.AnythingOfType("*domain.Alert")).Return(nil)

	err := service.CreateAlert(ctx, alert)

	assert.NoError(t, err)
	assert.NotEmpty(t, alert.AlertNumber)
	assert.Equal(t, "open", alert.Status)
	alertRepo.AssertExpectations(t)
}

func TestReportingService_ResolveAlert_Success(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	alert := createTestAlert()
	alertRepo.On("GetByID", ctx, "alert-001").Return(alert, nil)
	alertRepo.On("Update", ctx, mock.AnythingOfType("*domain.Alert")).Return(nil)

	err := service.ResolveAlert(ctx, "alert-001", "Reviewed and cleared", "analyst-001")

	assert.NoError(t, err)
	assert.Equal(t, "resolved", alert.Status)
	assert.Equal(t, "Reviewed and cleared", alert.Resolution)
	assert.NotNil(t, alert.ResolvedAt)
	assert.Equal(t, "analyst-001", alert.ResolvedBy)
	alertRepo.AssertExpectations(t)
}

func TestReportingService_EscalateAlert_Success(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	alert := createTestAlert()
	alertRepo.On("GetByID", ctx, "alert-001").Return(alert, nil)
	alertRepo.On("Update", ctx, mock.AnythingOfType("*domain.Alert")).Return(nil)

	err := service.EscalateAlert(ctx, "alert-001", "senior-officer-001")

	assert.NoError(t, err)
	assert.Equal(t, "escalated", alert.Status)
	assert.Equal(t, "senior-officer-001", alert.AssignedTo)
	alertRepo.AssertExpectations(t)
}

func TestReportingService_PerformComplianceCheck_NoMatch(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	screeningRepo.On("Search", ctx, "subject-001", mock.Anything).Return([]interface{}{}, nil)
	checkRepo.On("Create", ctx, mock.AnythingOfType("*domain.ComplianceCheck")).Return(nil)

	check, err := service.PerformComplianceCheck(ctx, "subject-001", "individual", "aml_screening")

	assert.NoError(t, err)
	assert.NotNil(t, check)
	assert.Equal(t, "pass", check.CheckResult)
	assert.Equal(t, "low", check.RiskLevel)
	screeningRepo.AssertExpectations(t)
	checkRepo.AssertExpectations(t)
}

func TestReportingService_PerformComplianceCheck_MatchFound(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	matchedEntity := map[string]interface{}{"name": "Blocked Entity"}
	screeningRepo.On("Search", ctx, "subject-002", mock.Anything).Return([]interface{}{matchedEntity}, nil)
	checkRepo.On("Create", ctx, mock.AnythingOfType("*domain.ComplianceCheck")).Return(nil)

	check, err := service.PerformComplianceCheck(ctx, "subject-002", "individual", "sanctions_screening")

	assert.NoError(t, err)
	assert.NotNil(t, check)
	assert.Equal(t, "review_required", check.CheckResult)
	assert.Equal(t, "high", check.RiskLevel)
	assert.Greater(t, check.MatchScore, float64(0))
	screeningRepo.AssertExpectations(t)
	checkRepo.AssertExpectations(t)
}

func TestReportingService_GenerateComplianceReport_SAR(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	sars := []*domain.SAR{createTestSAR(), createTestSAR()}
	sarRepo.On("List", ctx, mock.Anything).Return(sars, nil)
	reportRepo.On("Create", ctx, mock.AnythingOfType("*domain.ComplianceReport")).Return(nil)

	startTime := time.Now().AddDate(0, -1, 0)
	endTime := time.Now()

	report, err := service.GenerateComplianceReport(ctx, domain.ReportTypeSAR, startTime, endTime, "analyst-001")

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, domain.ReportTypeSAR, report.ReportType)
	assert.Equal(t, 2, int(report.Summary.SuspiciousActivity))
	sarRepo.AssertExpectations(t)
	reportRepo.AssertExpectations(t)
}

func TestReportingService_GenerateComplianceReport_CTR(t *testing.T) {
	ctx := context.Background()
	sarRepo := new(MockSARRepository)
	ctrRepo := new(MockCTRRepository)
	ruleRepo := new(MockComplianceRuleRepository)
	alertRepo := new(MockAlertRepository)
	checkRepo := new(MockComplianceCheckRepository)
	reportRepo := new(MockComplianceReportRepository)
	filingRepo := new(MockFilingRecordRepository)
	transactionRepo := new(MockTransactionRepository)
	screeningRepo := new(MockScreeningRepository)

	service := NewReportingService(
		sarRepo, ctrRepo, ruleRepo, alertRepo, checkRepo, reportRepo, filingRepo, transactionRepo, screeningRepo,
	)

	ctrs := []*domain.CTR{createTestCTR(), createTestCTR(), createTestCTR()}
	ctrRepo.On("List", ctx, mock.Anything).Return(ctrs, nil)
	reportRepo.On("Create", ctx, mock.AnythingOfType("*domain.ComplianceReport")).Return(nil)

	startTime := time.Now().AddDate(0, -1, 0)
	endTime := time.Now()

	report, err := service.GenerateComplianceReport(ctx, domain.ReportTypeCTR, startTime, endTime, "analyst-001")

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, domain.ReportTypeCTR, report.ReportType)
	assert.Equal(t, 3, int(report.Summary.TotalTransactions)
	assert.Equal(t, 45000.0, report.Summary.TotalVolume) // 3 CTRs × $15,000 each
	ctrRepo.AssertExpectations(t)
	reportRepo.AssertExpectations(t)
}
