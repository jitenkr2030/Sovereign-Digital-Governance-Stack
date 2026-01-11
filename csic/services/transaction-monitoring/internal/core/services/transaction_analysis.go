package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/csic-platform/services/transaction-monitoring/internal/core/domain"
	"github.com/csic-platform/services/transaction-monitoring/internal/core/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TransactionAnalysisService handles transaction analysis and risk assessment
type TransactionAnalysisService struct {
	transactionRepo  ports.TransactionRepository
	walletRepo       ports.WalletProfileRepository
	sanctionsRepo    ports.SanctionsRepository
	ruleRepo         ports.MonitoringRuleRepository
	logger           *zap.Logger
}

// NewTransactionAnalysisService creates a new transaction analysis service
func NewTransactionAnalysisService(
	transactionRepo ports.TransactionRepository,
	walletRepo ports.WalletProfileRepository,
	sanctionsRepo ports.SanctionsRepository,
	ruleRepo ports.MonitoringRuleRepository,
	logger *zap.Logger,
) *TransactionAnalysisService {
	return &TransactionAnalysisService{
		transactionRepo: transactionRepo,
		walletRepo:      walletRepo,
		sanctionsRepo:   sanctionsRepo,
		ruleRepo:        ruleRepo,
		logger:          logger,
	}
}

// AnalyzeTransaction performs comprehensive transaction analysis
func (s *TransactionAnalysisService) AnalyzeTransaction(ctx context.Context, tx *domain.Transaction) (*domain.TransactionAnalysisResult, error) {
	result := &domain.TransactionAnalysisResult{
		TransactionID:  uuid.MustParse(tx.ID),
		AnalyzedAt:     time.Now(),
		TriggeredRules: []domain.RuleMatch{},
		Recommendations: []string{},
	}

	// Step 1: Check sanctions list
	sanctionsMatches, err := s.checkSanctions(ctx, tx)
	if err != nil {
		s.logger.Error("Failed to check sanctions", zap.Error(err))
	}
	result.SanctionsMatches = sanctionsMatches

	// Step 2: Evaluate monitoring rules
	rules, err := s.ruleRepo.GetActiveRules(ctx)
	if err != nil {
		s.logger.Error("Failed to get active rules", zap.Error(err))
		rules = []domain.MonitoringRule{}
	}

	for _, rule := range rules {
		matched, detail, err := s.evaluateRule(ctx, &rule, tx)
		if err != nil {
			s.logger.Warn("Rule evaluation failed", zap.String("rule", rule.Name), zap.Error(err))
			continue
		}
		if matched {
			result.TriggeredRules = append(result.TriggeredRules, domain.RuleMatch{
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				RuleType:    rule.RuleType,
				Severity:    rule.Severity,
				RiskWeight:  rule.RiskWeight,
				MatchDetail: detail,
			})
		}
	}

	// Step 3: Calculate risk score
	riskScore := s.calculateRiskScore(tx, result.SanctionsMatches, result.TriggeredRules)
	result.RiskScore = riskScore
	result.RiskLevel = s.determineRiskLevel(riskScore)

	// Step 4: Calculate wallet risk delta
	walletRiskDelta, err := s.calculateWalletRiskDelta(ctx, tx)
	if err == nil {
		result.WalletRiskDelta = walletRiskDelta
	}

	// Step 5: Generate recommendations
	result.Recommendations = s.generateRecommendations(tx, result)

	// Step 6: Determine actions
	result.ShouldBlock = result.RiskLevel == domain.RiskCritical || len(result.SanctionsMatches) > 0
	result.ShouldAlert = result.RiskLevel == domain.RiskHigh || result.RiskLevel == domain.RiskCritical || len(result.SanctionsMatches) > 0

	// Step 7: Update transaction with risk assessment
	tx.RiskScore = riskScore
	tx.RiskLevel = result.RiskLevel
	if result.ShouldBlock {
		tx.Status = domain.TxStatusBlocked
	} else if result.ShouldAlert {
		tx.Status = domain.TxStatusFlagged
	}

	if err := s.transactionRepo.UpdateTransaction(ctx, tx); err != nil {
		s.logger.Error("Failed to update transaction", zap.Error(err))
	}

	return result, nil
}

// AnalyzeTransactionSync performs synchronous transaction analysis
func (s *TransactionAnalysisService) AnalyzeTransactionSync(ctx context.Context, txHash string) (*domain.RiskAssessment, error) {
	tx, err := s.transactionRepo.GetTransactionByHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	result, err := s.AnalyzeTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	return &domain.RiskAssessment{
		OverallScore:       int(result.RiskScore),
		RiskLevel:          string(result.RiskLevel),
		Factors:            s.convertToRiskFactors(result.TriggeredRules),
		TotalRiskFromFactors: result.RiskScore,
		Recommendations:    result.Recommendations,
		RequiresReview:     result.ShouldAlert,
		Flagged:            result.ShouldBlock || result.ShouldAlert,
	}, nil
}

// ProcessTransactionStream processes a stream of transactions
func (s *TransactionAnalysisService) ProcessTransactionStream(ctx context.Context, txs []*domain.Transaction) error {
	for _, tx := range txs {
		if _, err := s.AnalyzeTransaction(ctx, tx); err != nil {
			s.logger.Error("Failed to analyze transaction",
				zap.String("tx_hash", tx.TxHash),
				zap.Error(err))
			continue
		}
	}
	return nil
}

func (s *TransactionAnalysisService) checkSanctions(ctx context.Context, tx *domain.Transaction) ([]domain.SanctionsMatch, error) {
	matches := []domain.SanctionsMatch{}

	// Check sender address
	sender, err := s.sanctionsRepo.CheckAddress(ctx, tx.FromAddress)
	if err == nil && sender != nil {
		matches = append(matches, domain.SanctionsMatch{
			EntryID:    uuid.MustParse(sender.ID),
			EntityName: sender.EntityName,
			Address:    sender.Address,
			Program:    sender.Program,
			SourceList: sender.SourceList,
			MatchType:  "SENDER",
			MatchScore: 1.0,
		})
	}

	// Check receiver address
	if tx.ToAddress != "" {
		receiver, err := s.sanctionsRepo.CheckAddress(ctx, tx.ToAddress)
		if err == nil && receiver != nil {
			matches = append(matches, domain.SanctionsMatch{
				EntryID:    uuid.MustParse(receiver.ID),
				EntityName: receiver.EntityName,
				Address:    receiver.Address,
				Program:    receiver.Program,
				SourceList: receiver.SourceList,
				MatchType:  "RECEIVER",
				MatchScore: 1.0,
			})
		}
	}

	return matches, nil
}

func (s *TransactionAnalysisService) evaluateRule(ctx context.Context, rule *domain.MonitoringRule, tx *domain.Transaction) (bool, string, error) {
	var condition map[string]interface{}
	if err := json.Unmarshal([]byte(rule.Condition), &condition); err != nil {
		return false, "", err
	}

	switch rule.RuleType {
	case domain.RuleTypeThreshold:
		return s.evaluateThresholdRule(condition, tx)
	case domain.RuleTypeVelocity:
		return s.evaluateVelocityRule(condition, tx)
	case domain.RuleTypePattern:
		return s.evaluatePatternRule(condition, tx)
	case domain.RuleTypeGeographic:
		return s.evaluateGeographicRule(condition, tx)
	default:
		return false, "", nil
	}
}

func (s *TransactionAnalysisService) evaluateThresholdRule(condition map[string]interface{}, tx *domain.Transaction) (bool, string, error) {
	if maxAmount, ok := condition["max_amount"].(float64); ok {
		if tx.Amount > maxAmount {
			return true, fmt.Sprintf("Transaction amount %.2f exceeds threshold %.2f", tx.Amount, maxAmount), nil
		}
	}
	return false, "", nil
}

func (s *TransactionAnalysisService) evaluateVelocityRule(condition map[string]interface{}, tx *domain.Transaction) (bool, string, error) {
	return false, "", nil
}

func (s *TransactionAnalysisService) evaluatePatternRule(condition map[string]interface{}, tx *domain.Transaction) (bool, string, error) {
	return false, "", nil
}

func (s *TransactionAnalysisService) evaluateGeographicRule(condition map[string]interface{}, tx *domain.Transaction) (bool, string, error) {
	return false, "", nil
}

func (s *TransactionAnalysisService) calculateRiskScore(tx *domain.Transaction, sanctions []domain.SanctionsMatch, rules []domain.RuleMatch) float64 {
	score := 0.0

	// Sanctions match adds 100 points
	score += float64(len(sanctions)) * 100.0

	// Rule-based risk
	for _, rule := range rules {
		score += rule.RiskWeight
	}

	// Transaction amount risk
	if tx.AmountUSD > 100000 {
		score += 20
	} else if tx.AmountUSD > 10000 {
		score += 10
	} else if tx.AmountUSD > 1000 {
		score += 5
	}

	// Cap score at 100
	return math.Min(score, 100)
}

func (s *TransactionAnalysisService) determineRiskLevel(score float64) domain.RiskLevel {
	switch {
	case score >= 75:
		return domain.RiskCritical
	case score >= 50:
		return domain.RiskHigh
	case score >= 25:
		return domain.RiskMedium
	default:
		return domain.RiskLow
	}
}

func (s *TransactionAnalysisService) calculateWalletRiskDelta(ctx context.Context, tx *domain.Transaction) (float64, error) {
	profile, err := s.walletRepo.GetOrCreateWalletProfile(ctx, tx.FromAddress)
	if err != nil {
		return 0, err
	}

	oldScore := profile.CurrentRiskScore
	newScore := tx.RiskScore
	return newScore - oldScore, nil
}

func (s *TransactionAnalysisService) generateRecommendations(tx *domain.Transaction, result *domain.TransactionAnalysisResult) []string {
	recommendations := []string{}

	if len(result.SanctionsMatches) > 0 {
		recommendations = append(recommendations,
			"Transaction involves sanctioned entity - immediate escalation required",
			"Freeze all associated accounts pending investigation",
			"File SAR with relevant authorities")
	}

	if result.RiskLevel == domain.RiskCritical {
		recommendations = append(recommendations,
			"High-risk transaction requires manual review",
			"Verify sender and receiver identities",
			"Consider blocking transaction pending investigation")
	}

	if result.RiskLevel == domain.RiskHigh {
		recommendations = append(recommendations,
			"Enhanced due diligence recommended",
			"Review transaction patterns for the involved addresses")
	}

	return recommendations
}

func (s *TransactionAnalysisService) convertToRiskFactors(rules []domain.RuleMatch) []domain.RiskFactor {
	factors := []domain.RiskFactor{}
	for _, rule := range rules {
		factors = append(factors, domain.RiskFactor{
			Type:        string(rule.RuleType),
			Score:       int(rule.RiskWeight),
			Description: rule.MatchDetail,
		})
	}
	return factors
}
