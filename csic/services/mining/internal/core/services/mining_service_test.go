package services

import (
	"context"
	"testing"
	"time"

	"github.com/csic-platform/services/services/mining/internal/core/domain"
	"github.com/csic-platform/services/services/mining/internal/core/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MockRepository implements ports.MiningRepository for testing
type MockRepository struct {
	operations      map[uuid.UUID]*domain.MiningOperation
	walletMap       map[string]*domain.MiningOperation
	hashrateRecords map[uuid.UUID][]domain.HashrateRecord
	quotas          map[uuid.UUID]*domain.MiningQuota
	commands        map[uuid.UUID]*domain.ShutdownCommand
	violations      []domain.QuotaViolation
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		operations:      make(map[uuid.UUID]*domain.MiningOperation),
		walletMap:       make(map[string]*domain.MiningOperation),
		hashrateRecords: make(map[uuid.UUID][]domain.HashrateRecord),
		quotas:          make(map[uuid.UUID]*domain.MiningQuota),
		commands:        make(map[uuid.UUID]*domain.ShutdownCommand),
		violations:      make([]domain.QuotaViolation, 0),
	}
}

func (m *MockRepository) CreateOperation(ctx context.Context, op *domain.MiningOperation) error {
	m.operations[op.ID] = op
	m.walletMap[op.WalletAddress] = op
	return nil
}

func (m *MockRepository) GetOperation(ctx context.Context, id uuid.UUID) (*domain.MiningOperation, error) {
	return m.operations[id], nil
}

func (m *MockRepository) GetOperationByWallet(ctx context.Context, walletAddress string) (*domain.MiningOperation, error) {
	return m.walletMap[walletAddress], nil
}

func (m *MockRepository) UpdateOperationStatus(ctx context.Context, id uuid.UUID, status domain.OperationStatus) error {
	if op, ok := m.operations[id]; ok {
		op.Status = status
	}
	return nil
}

func (m *MockRepository) UpdateOperationHashrate(ctx context.Context, id uuid.UUID, hashrate float64, timestamp time.Time) error {
	if op, ok := m.operations[id]; ok {
		op.CurrentHashrate = hashrate
		op.LastReportAt = &timestamp
	}
	return nil
}

func (m *MockRepository) ListOperations(ctx context.Context, status *domain.OperationStatus, limit, offset int) ([]domain.MiningOperation, error) {
	var result []domain.MiningOperation
	count := 0
	for _, op := range m.operations {
		if status == nil || op.Status == *status {
			if count >= offset && (limit == 0 || len(result) < limit) {
				result = append(result, *op)
			}
			count++
		}
	}
	return result, nil
}

func (m *MockRepository) CountOperations(ctx context.Context, status *domain.OperationStatus) (int64, error) {
	var count int64
	for _, op := range m.operations {
		if status == nil || op.Status == *status {
			count++
		}
	}
	return count, nil
}

func (m *MockRepository) RecordHashrate(ctx context.Context, record *domain.HashrateRecord) error {
	m.hashrateRecords[record.OperationID] = append(m.hashrateRecords[record.OperationID], *record)
	return nil
}

func (m *MockRepository) GetHashrateHistory(ctx context.Context, opID uuid.UUID, limit int) ([]domain.HashrateRecord, error) {
	records := m.hashrateRecords[opID]
	if limit > 0 && len(records) > limit {
		return records[:limit], nil
	}
	return records, nil
}

func (m *MockRepository) GetLatestHashrate(ctx context.Context, opID uuid.UUID) (*domain.HashrateRecord, error) {
	records := m.hashrateRecords[opID]
	if len(records) > 0 {
		return &records[len(records)-1], nil
	}
	return nil, nil
}

func (m *MockRepository) CountHashrateRecords(ctx context.Context, opID uuid.UUID) (int64, error) {
	return int64(len(m.hashrateRecords[opID])), nil
}

func (m *MockRepository) SetQuota(ctx context.Context, quota *domain.MiningQuota) error {
	m.quotas[quota.ID] = quota
	return nil
}

func (m *MockRepository) GetCurrentQuota(ctx context.Context, opID uuid.UUID) (*domain.MiningQuota, error) {
	for _, quota := range m.quotas {
		if quota.OperationID == opID {
			return quota, nil
		}
	}
	return nil, nil
}

func (m *MockRepository) GetQuotas(ctx context.Context, opID uuid.UUID) ([]domain.MiningQuota, error) {
	var result []domain.MiningQuota
	for _, quota := range m.quotas {
		if quota.OperationID == opID {
			result = append(result, *quota)
		}
	}
	return result, nil
}

func (m *MockRepository) DeleteQuota(ctx context.Context, id uuid.UUID) error {
	delete(m.quotas, id)
	return nil
}

func (m *MockRepository) CreateShutdownCommand(ctx context.Context, cmd *domain.ShutdownCommand) error {
	m.commands[cmd.ID] = cmd
	return nil
}

func (m *MockRepository) GetShutdownCommand(ctx context.Context, id uuid.UUID) (*domain.ShutdownCommand, error) {
	return m.commands[id], nil
}

func (m *MockRepository) GetPendingCommands(ctx context.Context, opID uuid.UUID) ([]domain.ShutdownCommand, error) {
	var result []domain.ShutdownCommand
	for _, cmd := range m.commands {
		if cmd.OperationID == opID && (cmd.Status == domain.CommandIssued || cmd.Status == domain.CommandAcknowledged) {
			result = append(result, *cmd)
		}
	}
	return result, nil
}

func (m *MockRepository) UpdateCommandStatus(ctx context.Context, id uuid.UUID, status domain.CommandStatus, executedAt *time.Time) error {
	if cmd, ok := m.commands[id]; ok {
		cmd.Status = status
		cmd.ExecutedAt = executedAt
	}
	return nil
}

func (m *MockRepository) GetAllPendingCommands(ctx context.Context) ([]domain.ShutdownCommand, error) {
	var result []domain.ShutdownCommand
	for _, cmd := range m.commands {
		if cmd.Status == domain.CommandIssued || cmd.Status == domain.CommandAcknowledged {
			result = append(result, *cmd)
		}
	}
	return result, nil
}

func (m *MockRepository) RecordViolation(ctx context.Context, violation *domain.QuotaViolation) error {
	m.violations = append(m.violations, *violation)
	return nil
}

func (m *MockRepository) GetViolations(ctx context.Context, opID uuid.UUID, limit int) ([]domain.QuotaViolation, error) {
	var result []domain.QuotaViolation
	for _, v := range m.violations {
		if v.OperationID == opID {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *MockRepository) GetRegistryStats(ctx context.Context) (*domain.RegistryStats, error) {
	stats := &domain.RegistryStats{}
	for _, op := range m.operations {
		stats.TotalOperations++
		switch op.Status {
		case domain.StatusActive:
			stats.ActiveOperations++
		case domain.StatusSuspended:
			stats.SuspendedOperations++
		case domain.StatusShutdownExecuted:
			stats.ShutdownOperations++
		}
		stats.TotalHashrate += op.CurrentHashrate
	}
	stats.QuotaViolations = int64(len(m.violations))
	stats.PendingCommands = int64(len(m.commands))
	return stats, nil
}

// Test setup helper
func setupTestService() (*MiningService, *MockRepository) {
	log, _ := zap.NewDevelopment()
	repo := NewMockRepository()
	service := NewMiningService(repo, log)
	return service, repo
}

// Test helper to create a valid operation
func createTestOperation() *domain.MiningOperation {
	return &domain.MiningOperation{
		ID:             uuid.New(),
		OperatorName:   "Test Operator",
		WalletAddress:  "0x742d35Cc6634C0532925a3b844Bc9e7595f7547b",
		CurrentHashrate: 100.0,
		Status:         domain.StatusActive,
		Location:       "Test Location",
		Region:         "US-WEST",
		MachineType:    "Antminer S19 Pro",
		RegisteredAt:   time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// Test helper to create a valid quota
func createTestQuota(opID uuid.UUID) *domain.MiningQuota {
	return &domain.MiningQuota{
		ID:          uuid.New(),
		OperationID: opID,
		MaxHashrate: 150.0,
		QuotaType:   domain.QuotaFixed,
		ValidFrom:   time.Now(),
		Region:      "US-WEST",
		Priority:    1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestRegisterOperation_Success(t *testing.T) {
	service, repo := setupTestService()

	req := ports.RegisterOperationRequest{
		OperatorName:   "Test Mining Corp",
		WalletAddress:  "0x742d35Cc6634C0532925a3b844Bc9e7595f7547b",
		Location:       "Data Center A",
		Region:         "US-WEST",
		MachineType:    "Antminer S19 Pro",
		InitialHashrate: 100.0,
	}

	op, err := service.RegisterOperation(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if op.ID == uuid.Nil {
		t.Error("Expected operation ID to be set")
	}

	if op.Status != domain.StatusActive {
		t.Errorf("Expected status ACTIVE, got: %s", op.Status)
	}

	if _, exists := repo.operations[op.ID]; !exists {
		t.Error("Expected operation to be stored in repository")
	}
}

func TestRegisterOperation_DuplicateWallet(t *testing.T) {
	service, repo := setupTestService()

	// Create first operation
	existingOp := createTestOperation()
	repo.operations[existingOp.ID] = existingOp
	repo.walletMap[existingOp.WalletAddress] = existingOp

	req := ports.RegisterOperationRequest{
		OperatorName:   "Another Operator",
		WalletAddress:  existingOp.WalletAddress, // Same wallet
		Location:       "Different Location",
		Region:         "US-EAST",
		MachineType:    "Antminer S19 Pro",
	}

	_, err := service.RegisterOperation(context.Background(), req)
	if err == nil {
		t.Error("Expected error for duplicate wallet, got nil")
	}
}

func TestGetOperation_Success(t *testing.T) {
	service, repo := setupTestService()

	expectedOp := createTestOperation()
	repo.operations[expectedOp.ID] = expectedOp

	op, err := service.GetOperation(context.Background(), expectedOp.ID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if op == nil {
		t.Fatal("Expected operation, got nil")
	}

	if op.OperatorName != expectedOp.OperatorName {
		t.Errorf("Expected operator name %s, got: %s", expectedOp.OperatorName, op.OperatorName)
	}
}

func TestGetOperation_NotFound(t *testing.T) {
	service, _ := setupTestService()

	op, err := service.GetOperation(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if op != nil {
		t.Error("Expected nil operation for non-existent ID")
	}
}

func TestReportHashrate_Success(t *testing.T) {
	service, repo := setupTestService()

	op := createTestOperation()
	repo.operations[op.ID] = op

	req := ports.TelemetryRequest{
		OperationID: op.ID,
		Hashrate:    120.5,
		Unit:        "TH/s",
		BlockHeight: 1000000,
	}

	err := service.ReportHashrate(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check hashrate was recorded
	records := repo.hashrateRecords[op.ID]
	if len(records) != 1 {
		t.Errorf("Expected 1 hashrate record, got: %d", len(records))
	}

	if records[0].Hashrate != 120.5 {
		t.Errorf("Expected hashrate 120.5, got: %f", records[0].Hashrate)
	}
}

func TestReportHashrate_UnitConversion(t *testing.T) {
	service, repo := setupTestService()

	op := createTestOperation()
	repo.operations[op.ID] = op

	req := ports.TelemetryRequest{
		OperationID: op.ID,
		Hashrate:    1.5, // 1.5 PH/s = 1500 TH/s
		Unit:        "PH/s",
	}

	err := service.ReportHashrate(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	records := repo.hashrateRecords[op.ID]
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got: %d", len(records))
	}

	expectedHashrate := 1500.0
	if records[0].Hashrate != expectedHashrate {
		t.Errorf("Expected normalized hashrate %f TH/s, got: %f", expectedHashrate, records[0].Hashrate)
	}
}

func TestReportHashrate_QuotaViolation(t *testing.T) {
	service, repo := setupTestService()

	op := createTestOperation()
	repo.operations[op.ID] = op

	// Set a quota of 100 TH/s
	quota := createTestQuota(op.ID)
	quota.MaxHashrate = 100.0
	repo.quotas[quota.ID] = quota

	req := ports.TelemetryRequest{
		OperationID: op.ID,
		Hashrate:    150.0, // Exceeds quota
		Unit:        "TH/s",
	}

	err := service.ReportHashrate(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error (violation is logged, not returned), got: %v", err)
	}

	// Check violation was recorded
	if len(repo.violations) != 1 {
		t.Errorf("Expected 1 violation, got: %d", len(repo.violations))
	}
}

func TestAssignQuota_Success(t *testing.T) {
	service, repo := setupTestService()

	op := createTestOperation()
	repo.operations[op.ID] = op

	req := ports.QuotaRequest{
		OperationID: op.ID,
		MaxHashrate: 200.0,
		QuotaType:   domain.QuotaFixed,
		Region:      "US-WEST",
		Priority:    1,
	}

	quota, err := service.AssignQuota(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if quota.ID == uuid.Nil {
		t.Error("Expected quota ID to be set")
	}

	if quota.MaxHashrate != 200.0 {
		t.Errorf("Expected max hashrate 200.0, got: %f", quota.MaxHashrate)
	}

	if _, exists := repo.quotas[quota.ID]; !exists {
		t.Error("Expected quota to be stored")
	}
}

func TestIssueShutdownCommand_Success(t *testing.T) {
	service, repo := setupTestService()

	op := createTestOperation()
	repo.operations[op.ID] = op

	req := ports.ShutdownRequest{
		OperationID: op.ID,
		CommandType: domain.CommandGraceful,
		Reason:      "Grid emergency",
		IssuedBy:    "Grid Operator",
	}

	cmd, err := service.IssueShutdownCommand(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cmd.ID == uuid.Nil {
		t.Error("Expected command ID to be set")
	}

	if cmd.Status != domain.CommandIssued {
		t.Errorf("Expected status ISSUED, got: %s", cmd.Status)
	}

	// Check operation status was updated
	if repo.operations[op.ID].Status != domain.StatusShutdownOrdered {
		t.Errorf("Expected operation status SHUTDOWN_ORDERED, got: %s", repo.operations[op.ID].Status)
	}
}

func TestAcknowledgeCommand_Success(t *testing.T) {
	service, repo := setupTestService()

	op := createTestOperation()
	repo.operations[op.ID] = op

	cmd := &domain.ShutdownCommand{
		ID:          uuid.New(),
		OperationID: op.ID,
		CommandType: domain.CommandGraceful,
		Reason:      "Test",
		Status:      domain.CommandIssued,
		IssuedAt:    time.Now(),
		IssuedBy:    "System",
	}
	repo.commands[cmd.ID] = cmd

	err := service.AcknowledgeCommand(context.Background(), cmd.ID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if repo.commands[cmd.ID].Status != domain.CommandAcknowledged {
		t.Errorf("Expected status ACKNOWLEDGED, got: %s", repo.commands[cmd.ID].Status)
	}
}

func TestConfirmShutdown_Success(t *testing.T) {
	service, repo := setupTestService()

	op := createTestOperation()
	repo.operations[op.ID] = op

	cmd := &domain.ShutdownCommand{
		ID:          uuid.New(),
		OperationID: op.ID,
		CommandType: domain.CommandGraceful,
		Reason:      "Test",
		Status:      domain.CommandAcknowledged,
		IssuedAt:    time.Now(),
		IssuedBy:    "System",
	}
	repo.commands[cmd.ID] = cmd

	err := service.ConfirmShutdown(context.Background(), cmd.ID, "Success - graceful shutdown completed")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if repo.commands[cmd.ID].Status != domain.CommandExecuted {
		t.Errorf("Expected status EXECUTED, got: %s", repo.commands[cmd.ID].Status)
	}

	if repo.operations[op.ID].Status != domain.StatusShutdownExecuted {
		t.Errorf("Expected operation status SHUTDOWN_EXECUTED, got: %s", repo.operations[op.ID].Status)
	}
}

func TestGetRegistryStats_Success(t *testing.T) {
	service, repo := setupTestService()

	// Create test operations
	op1 := createTestOperation()
	op1.Status = domain.StatusActive
	op2 := createTestOperation()
	op2.ID = uuid.New()
	op2.WalletAddress = "0xdD2FD4581271e230360230F9337D5c0430Bf44C0"
	op2.Status = domain.StatusSuspended

	repo.operations[op1.ID] = op1
	repo.operations[op2.ID] = op2
	repo.walletMap[op1.WalletAddress] = op1
	repo.walletMap[op2.WalletAddress] = op2

	stats, err := service.GetRegistryStats(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if stats.TotalOperations != 2 {
		t.Errorf("Expected 2 total operations, got: %d", stats.TotalOperations)
	}

	if stats.ActiveOperations != 1 {
		t.Errorf("Expected 1 active operation, got: %d", stats.ActiveOperations)
	}

	if stats.SuspendedOperations != 1 {
		t.Errorf("Expected 1 suspended operation, got: %d", stats.SuspendedOperations)
	}
}

func TestIsValidStatusTransition(t *testing.T) {
	service, _ := setupTestService()

	testCases := []struct {
		from   domain.OperationStatus
		to     domain.OperationStatus
		expect bool
	}{
		{domain.StatusActive, domain.StatusSuspended, true},
		{domain.StatusActive, domain.StatusShutdownOrdered, true},
		{domain.StatusActive, domain.StatusNonCompliant, true},
		{domain.StatusSuspended, domain.StatusActive, true},
		{domain.StatusSuspended, domain.StatusShutdownOrdered, true},
		{domain.StatusActive, domain.StatusActive, false}, // Same status not allowed
		{domain.StatusShutdownExecuted, domain.StatusActive, true},
		{domain.StatusActive, domain.StatusShutdownExecuted, false}, // Direct transition not allowed
	}

	for _, tc := range testCases {
		result := service.isValidStatusTransition(tc.from, tc.to)
		if result != tc.expect {
			t.Errorf("isValidStatusTransition(%s, %s) = %v, expected %v",
				tc.from, tc.to, result, tc.expect)
		}
	}
}

func TestNormalizeHashrate(t *testing.T) {
	service, _ := setupTestService()

	testCases := []struct {
		hashrate float64
		unit     string
		expected float64
	}{
		{100.0, "TH/s", 100.0},
		{1.0, "PH/s", 1000.0},
		{100000.0, "GH/s", 100.0},
	}

	for _, tc := range testCases {
		result := service.normalizeHashrate(tc.hashrate, tc.unit)
		if result != tc.expected {
			t.Errorf("normalizeHashrate(%f, %s) = %f, expected %f",
				tc.hashrate, tc.unit, result, tc.expected)
		}
	}
}
