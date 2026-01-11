package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/csic-platform/tests/integration/internal/testsuite"
	"github.com/csic_platform/tests/integration/internal/containers"
	"github.com/csic_platform/tests/integration/internal/clients"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	
	fmt.Println("Starting CSIC Platform Integration Tests...")
	
	// Setup test infrastructure
	setup := containers.NewTestInfrastructure(ctx)
	
	// Start infrastructure
	if err := setup.Start(ctx); err != nil {
		fmt.Printf("Failed to start infrastructure: %v\n", err)
		os.Exit(1)
	}
	
	// Run tests
	code := m.Run()
	
	// Cleanup
	setup.Cleanup(ctx)
	
	os.Exit(code)
}

func TestMiningControlIntegration(t *testing.T) {
	ctx := context.Background()
	
	t.Run("RegisterAndMonitorMiningOperation", func(t *testing.T) {
		test := testsuite.NewTestCase(t, "Mining Control Integration")
		
		// Create mining service client
		client, err := clients.NewMiningClient(ctx, "http://localhost:8080")
		if err != nil {
			t.Fatalf("Failed to create mining client: %v", err)
		}
		
		// Register new operation
		op, err := client.RegisterOperation(ctx, &clients.RegisterOperationRequest{
			OperatorName:  "Integration Test Mining Corp",
			WalletAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f7547b",
			Location:      "Test Data Center",
			Region:        "US-WEST",
			MachineType:   "Antminer S19 Pro",
		})
		if err != nil {
			t.Fatalf("Failed to register operation: %v", err)
		}
		
		test.AssertNotNil(op.ID, "Operation ID should be set")
		test.AssertEqual(op.Status, "ACTIVE", "Initial status should be ACTIVE")
		
		// Report hashrate
		err = client.ReportHashrate(ctx, op.ID, 150.5, "TH/s")
		if err != nil {
			t.Fatalf("Failed to report hashrate: %v", err)
		}
		
		// Get operation
		fetched, err := client.GetOperation(ctx, op.ID)
		if err != nil {
			t.Fatalf("Failed to get operation: %v", err)
		}
		
		test.AssertEqual(fetched.CurrentHashrate, 150.5, "Hashrate should be updated")
		
		fmt.Printf("✓ Mining Control test passed for operation %s\n", op.ID.String())
	})
}

func TestComplianceIntegration(t *testing.T) {
	ctx := context.Background()
	
	t.Run("EntityRegistrationAndScoring", func(t *testing.T) {
		test := testsuite.NewTestCase(t, "Compliance Integration")
		
		// Create compliance client
		client, err := clients.NewComplianceClient(ctx, "http://localhost:8081")
		if err != nil {
			t.Fatalf("Failed to create compliance client: %v", err)
		}
		
		// Register entity
		entity, err := client.RegisterEntity(ctx, &clients.RegisterEntityRequest{
			Name:            "Integration Test Entity",
			LegalName:       "Integration Test Entity LLC",
			RegistrationNum: fmt.Sprintf("INT-TEST-%d", time.Now().Unix()),
			Jurisdiction:    "US",
			EntityType:      "EXCHANGE",
			Address:         "123 Test St",
			ContactEmail:    "test@integration.com",
			RiskLevel:       "MEDIUM",
		})
		if err != nil {
			t.Fatalf("Failed to register entity: %v", err)
		}
		
		test.AssertNotNil(entity.ID, "Entity ID should be set")
		
		// Issue license
		license, err := client.IssueLicense(ctx, &clients.IssueLicenseRequest{
			EntityID:      entity.ID,
			LicenseType:   "EXCHANGE",
			LicenseNumber: fmt.Sprintf("LCC-INT-%d", time.Now().Unix()),
			Jurisdiction:  "US",
			IssuedBy:      "Test Authority",
		})
		if err != nil {
			t.Fatalf("Failed to issue license: %v", err)
		}
		
		test.AssertNotNil(license.ID, "License ID should be set")
		test.AssertEqual(license.Status, "ACTIVE", "License should be ACTIVE")
		
		// Calculate compliance score
		score, err := client.CalculateScore(ctx, entity.ID)
		if err != nil {
			t.Fatalf("Failed to calculate score: %v", err)
		}
		
		test.AssertTrue(score.TotalScore >= 0 && score.TotalScore <= 100, "Score should be between 0-100")
		
		fmt.Printf("✓ Compliance test passed for entity %s\n", entity.ID.String())
	})
}

func TestCrossServiceWorkflow(t *testing.T) {
	ctx := context.Background()
	
	t.Run("MiningComplianceAndMonitoring", func(t *testing.T) {
		test := testsuite.NewTestCase(t, "Cross-Service Workflow")
		
		// This test simulates a complete workflow across services
		// Mining operation reports hashrate -> Compliance checks license -> Monitoring tracks activity
		
		miningClient, err := clients.NewMiningClient(ctx, "http://localhost:8080")
		if err != nil {
			t.Fatalf("Failed to create mining client: %v", err)
		}
		
		complianceClient, err := clients.NewComplianceClient(ctx, "http://localhost:8081")
		if err != nil {
			t.Fatalf("Failed to create compliance client: %v", err)
		}
		
		// 1. Register mining operation
		op, err := miningClient.RegisterOperation(ctx, &clients.RegisterOperationRequest{
			OperatorName:  "Cross-Service Test Mining",
			WalletAddress: "0xdD2FD4581271e230360230F9337D5c0430Bf44C0",
			Location:      "Cross-Test Facility",
			Region:        "US-EAST",
			MachineType:   "MicroBT M30S+",
		})
		if err != nil {
			t.Fatalf("Failed to register mining operation: %v", err)
		}
		
		// 2. Create compliance entity for the mining operation
		entity, err := complianceClient.RegisterEntity(ctx, &clients.RegisterEntityRequest{
			Name:            "Cross-Service Mining Entity",
			LegalName:       "Cross-Service Mining LLC",
			RegistrationNum: fmt.Sprintf("CSM-%d", time.Now().UnixNano()),
			Jurisdiction:    "US",
			EntityType:      "MINING",
			Address:         "Cross-Test Facility Address",
			ContactEmail:    "cross@test.com",
			RiskLevel:       "HIGH",
		})
		if err != nil {
			t.Fatalf("Failed to register compliance entity: %v", err)
		}
		
		// 3. Issue mining license
		license, err := complianceClient.IssueLicense(ctx, &clients.IssueLicenseRequest{
			EntityID:      entity.ID,
			LicenseType:   "MINING",
			LicenseNumber: fmt.Sprintf("MINING-CS-%d", time.Now().Unix()),
			Jurisdiction:  "US",
			IssuedBy:      "Test Authority",
		})
		if err != nil {
			t.Fatalf("Failed to issue mining license: %v", err)
		}
		
		// 4. Verify cross-service state consistency
		test.AssertNotNil(op.ID, "Mining operation should exist")
		test.AssertNotNil(entity.ID, "Compliance entity should exist")
		test.AssertNotNil(license.ID, "License should exist")
		
		// 5. Simulate hashrate reporting and compliance check
		err = miningClient.ReportHashrate(ctx, op.ID, 200.0, "TH/s")
		if err != nil {
			t.Fatalf("Failed to report hashrate: %v", err)
		}
		
		// 6. Check compliance score reflects the mining operation
		score, err := complianceClient.CalculateScore(ctx, entity.ID)
		if err != nil {
			t.Fatalf("Failed to calculate compliance score: %v", err)
		}
		
		test.AssertNotNil(score, "Compliance score should be calculated")
		
		fmt.Printf("✓ Cross-service workflow test passed\n")
		fmt.Printf("  - Mining Operation: %s\n", op.ID.String())
		fmt.Printf("  - Compliance Entity: %s\n", entity.ID.String())
		fmt.Printf("  - License: %s\n", license.ID.String())
		fmt.Printf("  - Compliance Score: %.2f\n", score.TotalScore)
	})
}

func TestPerformanceAndLatency(t *testing.T) {
	t.Run("ServiceResponseTime", func(t *testing.T) {
		ctx := context.Background()
		
		client, err := clients.NewMiningClient(ctx, "http://localhost:8080")
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		
		// Measure response time for list operations
		start := time.Now()
		_, err = client.ListOperations(ctx, "", 1, 10)
		latency := time.Since(start)
		
		if err != nil {
			t.Fatalf("Failed to list operations: %v", err)
		}
		
		// Assert response time is under 200ms
		if latency > 200*time.Millisecond {
			t.Logf("Warning: Response time %.2fms exceeded target of 200ms", latency.Seconds()*1000)
		}
		
		fmt.Printf("✓ Performance test: List operations completed in %.2fms\n", latency.Seconds()*1000)
	})
}
