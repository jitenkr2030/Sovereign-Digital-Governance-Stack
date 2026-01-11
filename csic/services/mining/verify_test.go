package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/csic-platform/services/mining/internal/core/domain"
	"github.com/csic-platform/services/mining/internal/core/services"
)

// MockRepository implements MiningRepository for testing
type MockRepository struct {
	operations map[string]*domain.MiningOperation
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		operations: make(map[string]*domain.MiningOperation),
	}
}

func (m *MockRepository) CreateOperation(ctx context.Context, op *domain.MiningOperation) error {
	m.operations[op.ID.String()] = op
	return nil
}

func (m *MockRepository) GetOperation(ctx context.Context, id string) (*domain.MiningOperation, error) {
	if op, ok := m.operations[id]; ok {
		return op, nil
	}
	return nil, fmt.Errorf("operation not found")
}

func (m *MockRepository) UpdateOperation(ctx context.Context, op *domain.MiningOperation) error {
	m.operations[op.ID.String()] = op
	return nil
}

func (m *MockRepository) ListOperations(ctx context.Context, status string, limit, offset int) ([]*domain.MiningOperation, error) {
	var result []*domain.MiningOperation
	for _, op := range m.operations {
		if status == "" || op.Status == status {
			result = append(result, op)
		}
	}
	return result, nil
}

func (m *MockRepository) RecordHashrate(ctx context.Context, opID string, hashrate float64, unit string) error {
	return nil
}

func main() {
	fmt.Println("=== Mining Service Verification ===")
	fmt.Println()

	// Create mock repository and service
	repo := NewMockRepository()
	svc := services.NewMiningService(repo, nil)

	ctx := context.Background()

	// Test 1: Register a new operation
	fmt.Println("Test 1: Registering new mining operation...")
	op := &domain.MiningOperation{
		OperatorName:  "Test Mining Corp",
		WalletAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f7547b",
		Location:      "Test Data Center",
		Region:        "US-WEST",
		MachineType:   "Antminer S19 Pro",
		Status:        "ACTIVE",
		CurrentHashrate: 0,
		RegisteredAt: time.Now(),
	}

	err := svc.RegisterOperation(ctx, op)
	if err != nil {
		fmt.Printf("FAIL: Failed to register operation: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PASS: Operation registered with ID %s\n", op.ID.String())
	fmt.Println()

	// Test 2: Report hashrate
	fmt.Println("Test 2: Reporting hashrate...")
	err = svc.ReportHashrate(ctx, op.ID.String(), 150.5, "TH/s")
	if err != nil {
		fmt.Printf("FAIL: Failed to report hashrate: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("PASS: Hashrate reported successfully")
	fmt.Println()

	// Test 3: Check quota violations
	fmt.Println("Test 3: Checking quota violations...")
	violations, err := svc.CheckQuotaViolations(ctx, op.ID.String())
	if err != nil {
		fmt.Printf("FAIL: Failed to check violations: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PASS: Found %d quota violations\n", violations)
	fmt.Println()

	// Test 4: List operations
	fmt.Println("Test 4: Listing operations...")
	ops, err := svc.ListOperations(ctx, "", 10, 0)
	if err != nil {
		fmt.Printf("FAIL: Failed to list operations: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PASS: Listed %d operations\n", len(ops))
	fmt.Println()

	// Test 5: Get operation details
	fmt.Println("Test 5: Getting operation details...")
	fetched, err := svc.GetOperation(ctx, op.ID.String())
	if err != nil {
		fmt.Printf("FAIL: Failed to get operation: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PASS: Retrieved operation - %s (Hashrate: %.1f %s)\n", fetched.OperatorName, fetched.CurrentHashrate, "TH/s")
	fmt.Println()

	fmt.Println("=== All Mining Service Tests Passed ===")
}
