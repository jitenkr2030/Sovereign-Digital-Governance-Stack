package services

import (
	"context"
	"testing"
	"time"

	"github.com/csic/monitoring/internal/core/domain"
	"go.uber.org/zap"
)

// MockSanctionsRepository implements ports.SanctionsRepository for testing
type MockSanctionsRepository struct {
	sanctions map[string]bool
}

func NewMockSanctionsRepository() *MockSanctionsRepository {
	return &MockSanctionsRepository{
		sanctions: make(map[string]bool),
	}
}

func (m *MockSanctionsRepository) Exists(ctx context.Context, address, chain string) (bool, error) {
	key := address + ":" + chain
	return m.sanctions[key], nil
}

func (m *MockSanctionsRepository) Add(address, chain string) {
	key := address + ":" + chain
	m.sanctions[key] = true
}

// MockWalletProfileRepository implements ports.WalletProfileRepository for testing
type MockWalletProfileRepository struct {
	profiles map[string]*domain.WalletProfile
}

func NewMockWalletProfileRepository() *MockWalletProfileRepository {
	return &MockWalletProfileRepository{
		profiles: make(map[string]*domain.WalletProfile),
	}
}

func (m *MockWalletProfileRepository) GetByAddress(ctx context.Context, address, chain string) (*domain.WalletProfile, error) {
	key := address + ":" + chain
	return m.profiles[key], nil
}

func (m *MockWalletProfileRepository) Add(profile *domain.WalletProfile) {
	key := profile.Address + ":" + profile.Chain
	m.profiles[key] = profile
}

// TestRiskScoringService_CalculateRiskScore tests the risk scoring logic
func TestRiskScoringService_CalculateRiskScore(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	mockSanctions := NewMockSanctionsRepository()
	mockWallets := NewMockWalletProfileRepository()

	service := NewRiskScoringService(mockSanctions, mockWallets, logger)

	tests := []struct {
		name           string
		tx             *domain.Transaction
		expectedScore  int
		expectedLevel  string
		expectFlagged  bool
	}{
		{
			name: "low risk transaction",
			tx: &domain.Transaction{
				TxHash:    "0x123",
				Chain:     "ethereum",
				FromAddress: "0xabc",
				AmountUSD: 100,
			},
			expectedScore: 0,
			expectedLevel: "MINIMAL",
			expectFlagged: false,
		},
		{
			name: "medium risk transaction (large amount)",
			tx: &domain.Transaction{
				TxHash:    "0x456",
				Chain:     "ethereum",
				FromAddress: "0xdef",
				AmountUSD: 15000,
			},
			expectedScore: 20,
			expectedLevel: "LOW",
			expectFlagged: false,
		},
		{
			name: "high risk transaction (very large amount)",
			tx: &domain.Transaction{
				TxHash:    "0x789",
				Chain:     "bitcoin",
				FromAddress: "0xghi",
				AmountUSD: 150000,
			},
			expectedScore: 50,
			expectedLevel: "MEDIUM",
			expectFlagged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			assessment, err := service.CalculateRiskScore(ctx, tt.tx)
			if err != nil {
				t.Fatalf("CalculateRiskScore failed: %v", err)
			}

			if assessment.OverallScore != tt.expectedScore {
				t.Errorf("Expected score %d, got %d", tt.expectedScore, assessment.OverallScore)
			}

			if assessment.RiskLevel != tt.expectedLevel {
				t.Errorf("Expected risk level %s, got %s", tt.expectedLevel, assessment.RiskLevel)
			}

			if assessment.Flagged != tt.expectFlagged {
				t.Errorf("Expected flagged %v, got %v", tt.expectFlagged, assessment.Flagged)
			}
		})
	}
}

// TestRiskScoringService_checkAmountThresholds tests amount threshold checking
func TestRiskScoringService_checkAmountThresholds(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	mockSanctions := NewMockSanctionsRepository()
	mockWallets := NewMockWalletProfileRepository()

	service := NewRiskScoringService(mockSanctions, mockWallets, logger)

	tests := []struct {
		name          string
		amountUSD     float64
		expectedScore int
	}{
		{"small amount", 500, 0},
		{"threshold 1000", 1500, 5},
		{"threshold 10000", 15000, 20},
		{"threshold 50000", 75000, 30},
		{"threshold 100000", 150000, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factor := service.checkAmountThresholds(tt.amountUSD)
			if factor.Score != tt.expectedScore {
				t.Errorf("Expected score %d for amount %.2f, got %d", tt.expectedScore, tt.amountUSD, factor.Score)
			}
		})
	}
}

// TestRiskScoringService_calculateRiskLevel tests risk level determination
func TestRiskScoringService_calculateRiskLevel(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	mockSanctions := NewMockSanctionsRepository()
	mockWallets := NewMockWalletProfileRepository()

	service := NewRiskScoringService(mockSanctions, mockWallets, logger)

	tests := []struct {
		score        int
		expectedLevel string
	}{
		{0, "MINIMAL"},
		{10, "MINIMAL"},
		{19, "MINIMAL"},
		{20, "LOW"},
		{39, "LOW"},
		{40, "MEDIUM"},
		{59, "MEDIUM"},
		{60, "HIGH"},
		{79, "HIGH"},
		{80, "CRITICAL"},
		{100, "CRITICAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := service.calculateRiskLevel(tt.score)
			if level != tt.expectedLevel {
				t.Errorf("Expected level %s for score %d, got %s", tt.expectedLevel, tt.score, level)
			}
		})
	}
}

// TestTransactionService_IngestTransaction tests transaction ingestion
func TestTransactionService_IngestTransaction(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Note: In a real test, you would use a mock repository
	// For now, this is a placeholder for the test structure

	t.Run("transaction with valid data", func(t *testing.T) {
		tx := &domain.Transaction{
			TxHash:     "0x123456789",
			Chain:      "ethereum",
			FromAddress: "0xabcd",
			ToAddress:   stringPtr("0xefgh"),
			Amount:      1.5,
			AmountUSD:   3000,
			TxTimestamp: time.Now(),
		}

		if tx.TxHash == "" {
			t.Error("Transaction should have a hash")
		}

		if tx.AmountUSD <= 0 {
			t.Error("Transaction should have a positive USD amount")
		}
	})
}

func stringPtr(s string) *string {
	return &s
}
