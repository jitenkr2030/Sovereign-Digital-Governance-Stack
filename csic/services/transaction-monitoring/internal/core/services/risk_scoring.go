package services

import (
	"context"
	"fmt"
	"time"

	"github.com/csic/monitoring/internal/core/domain"
	"github.com/csic/monitoring/internal/core/ports"
	"go.uber.org/zap"
)

// RiskScoringService calculates risk scores for transactions
type RiskScoringService struct {
	sanctionsRepo   ports.SanctionsRepository
	walletProfileRepo ports.WalletProfileRepository
	logger          *zap.Logger
}

// NewRiskScoringService creates a new risk scoring service
func NewRiskScoringService(
	sanctionsRepo ports.SanctionsRepository,
	walletProfileRepo ports.WalletProfileRepository,
	logger *zap.Logger,
) *RiskScoringService {
	return &RiskScoringService{
		sanctionsRepo:   sanctionsRepo,
		walletProfileRepo: walletProfileRepo,
		logger:          logger,
	}
}

// CalculateRiskScore calculates the risk score for a transaction
func (s *RiskScoringService) CalculateRiskScore(ctx context.Context, tx *domain.Transaction) (*domain.RiskAssessment, error) {
	assessment := &domain.RiskAssessment{
		Factors:           make([]domain.RiskFactor, 0),
		Recommendations:   make([]string, 0),
		RequiresReview:    false,
		Flagged:           false,
	}

	// Check sanctions list first (highest priority)
	sanctionsCheck := s.checkSanctions(tx.FromAddress, tx.Chain)
	if sanctionsCheck.Score > 0 {
		assessment.Factors = append(assessment.Factors, sanctionsCheck)
		assessment.Flagged = true
		assessment.RequiresReview = true
	}

	if tx.ToAddress != nil {
		sanctionsCheck = s.checkSanctions(*tx.ToAddress, tx.Chain)
		if sanctionsCheck.Score > 0 {
			assessment.Factors = append(assessment.Factors, sanctionsCheck)
			assessment.Flagged = true
			assessment.RequiresReview = true
		}
	}

	// Check amount thresholds (structuring detection)
	amountFactor := s.checkAmountThresholds(tx.AmountUSD)
	if amountFactor.Score > 0 {
		assessment.Factors = append(assessment.Factors, amountFactor)
	}

	// Check wallet age
	walletAgeFactor := s.checkWalletAge(tx.FromAddress)
	if walletAgeFactor.Score > 0 {
		assessment.Factors = append(assessment.Factors, walletAgeFactor)
	}

	// Check transaction velocity
	txCountFactor := s.checkTransactionVelocity(tx.FromAddress)
	if txCountFactor.Score > 0 {
		assessment.Factors = append(assessment.Factors, txCountFactor)
	}

	// Check for mixing patterns (simplified)
	mixingFactor := s.checkMixingPatterns(tx)
	if mixingFactor.Score > 0 {
		assessment.Factors = append(assessment.Factors, mixingFactor)
	}

	// Calculate total score from factors
	totalScore := 0
	for _, factor := range assessment.Factors {
		totalScore += factor.Score
	}

	// Cap score at 100
	if totalScore > 100 {
		totalScore = 100
	}

	assessment.OverallScore = totalScore
	assessment.RiskLevel = s.calculateRiskLevel(totalScore)

	// Generate recommendations based on risk factors
	assessment.Recommendations = s.generateRecommendations(assessment)

	// Auto-flag high-risk transactions
	if totalScore >= 80 {
		assessment.Flagged = true
		assessment.RequiresReview = true
	}

	return assessment, nil
}

func (s *RiskScoringService) checkSanctions(address, chain string) domain.RiskFactor {
	// This would typically query a fast lookup (Redis cache or Bloom filter)
	// For now, return a placeholder that would be replaced with actual check
	return domain.RiskFactor{
		Type:       "SANCTIONS_LIST",
		Score:      100,
		Threshold:  0,
		Observed:   1,
		Description: fmt.Sprintf("Address %s found in sanctions list", address[:8]+"..."),
	}
}

func (s *RiskScoringService) checkAmountThresholds(amountUSD float64) domain.RiskFactor {
	factor := domain.RiskFactor{
		Type:       "AMOUNT_THRESHOLD",
		Observed:   amountUSD,
	}

	if amountUSD >= 100000 {
		factor.Score = 50
		factor.Threshold = 100000
		factor.Description = "Transaction amount exceeds $100,000 threshold"
	} else if amountUSD >= 50000 {
		factor.Score = 30
		factor.Threshold = 50000
		factor.Description = "Transaction amount exceeds $50,000 threshold"
	} else if amountUSD >= 10000 {
		factor.Score = 20
		factor.Threshold = 10000
		factor.Description = "Transaction amount exceeds $10,000 threshold (potential structuring)"
	} else if amountUSD >= 1000 {
		factor.Score = 5
		factor.Threshold = 1000
		factor.Description = "Transaction amount exceeds $1,000 threshold"
	}

	return factor
}

func (s *RiskScoringService) checkWalletAge(address string) domain.RiskFactor {
	// In production, this would query the wallet profile repository
	return domain.RiskFactor{
		Type:        "WALLET_AGE",
		Score:       20,
		Threshold:   24, // hours
		Observed:    0,  // Would be actual wallet age
		Description: "Wallet age less than 24 hours",
	}
}

func (s *RiskScoringService) checkTransactionVelocity(address string) domain.RiskFactor {
	// In production, this would query recent transaction count
	return domain.RiskFactor{
		Type:        "TX_VELOCITY",
		Score:       15,
		Threshold:   10,
		Observed:    0,  // Would be actual transaction count
		Description: "High transaction velocity detected",
	}
}

func (s *RiskScoringService) checkMixingPatterns(tx *domain.Transaction) domain.RiskFactor {
	// Simplified mixing detection - in production would use more sophisticated analysis
	return domain.RiskFactor{
		Type:        "MIXING_PATTERN",
		Score:       0,
		Threshold:   0,
		Observed:    0,
		Description: "",
	}
}

func (s *RiskScoringService) calculateRiskLevel(score int) string {
	switch {
	case score >= 80:
		return "CRITICAL"
	case score >= 60:
		return "HIGH"
	case score >= 40:
		return "MEDIUM"
	case score >= 20:
		return "LOW"
	default:
		return "MINIMAL"
	}
}

func (s *RiskScoringService) generateRecommendations(assessment *domain.RiskAssessment) []string {
	recommendations := make([]string, 0)

	for _, factor := range assessment.Factors {
		switch factor.Type {
		case "SANCTIONS_LIST":
			recommendations = append(recommendations, "IMMEDIATE: Report to compliance officer - address matches sanctions list")
			recommendations = append(recommendations, "BLOCK: Consider blocking transaction until investigation complete")
		case "AMOUNT_THRESHOLD":
			recommendations = append(recommendations, "REVIEW: Enhanced due diligence recommended for large transactions")
		case "WALLET_AGE":
			recommendations = append(recommendations, "VERIFY: New wallet - consider identity verification")
		case "TX_VELOCITY":
			recommendations = append(recommendations, "MONITOR: Unusual transaction frequency - monitor for patterns")
		case "MIXING_PATTERN":
			recommendations = append(recommendations, "INVESTIGATE: Potential mixing pattern - review for money laundering")
		}
	}

	if assessment.Flagged {
		recommendations = append(recommendations, "FLAG: Transaction flagged for manual review")
	}

	return recommendations
}

// CalculateWalletRisk calculates risk score for a wallet address
func (s *RiskScoringService) CalculateWalletRisk(ctx context.Context, address, chain string) (*domain.WalletProfile, error) {
	profile, err := s.walletProfileRepo.GetByAddress(ctx, address, chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet profile: %w", err)
	}

	if profile == nil {
		// Create new profile
		profile = &domain.WalletProfile{
			Address: address,
			Chain:   chain,
		}
	}

	// Check if wallet is sanctioned
	sanctionsCheck := s.checkSanctions(address, chain)
	if sanctionsCheck.Score > 0 {
		profile.RiskIndicators = append(profile.RiskIndicators, domain.RiskIndicator{
			Indicator:    "SANCTIONED",
			Severity:     "CRITICAL",
			Description:  "Address found in sanctions list",
			LastObserved: time.Now(),
			Count:        1,
		})
	}

	return profile, nil
}
