package risk

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/csic/transaction-monitoring/internal/config"
	"github.com/csic/transaction-monitoring/internal/domain/models"
	"github.com/csic/transaction-monitoring/internal/repository"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// RiskScoringService handles wallet risk assessment
type RiskScoringService struct {
	cfg      *config.Config
	repo     *repository.Repository
	cache    *repository.CacheRepository
	alertSvc interface{} // Alert service dependency
	logger   *zap.Logger
	mu       sync.RWMutex
}

// NewRiskScoringService creates a new risk scoring service
func NewRiskScoringService(
	cfg *config.Config,
	repo *repository.Repository,
	cache *repository.CacheRepository,
	logger *zap.Logger,
) *RiskScoringService {
	return &RiskScoringService{
		cfg:    cfg,
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// CalculateRiskScore calculates the comprehensive risk score for a wallet
func (s *RiskScoringService) CalculateRiskScore(ctx context.Context, address string, network models.Network) (*models.RiskFactors, error) {
	// Get wallet data
	wallet, err := s.repo.GetWalletByAddress(ctx, address, network)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Get historical metrics
	metrics, err := s.repo.GetWalletMetrics(ctx, wallet.ID)
	if err != nil {
		s.logger.Warn("Failed to get wallet metrics", zap.Error(err))
		metrics = &models.WalletMetrics{}
	}

	// Calculate individual risk components
	riskFactors := &models.WalletFactors{
		WalletID: wallet.ID,
	}

	// 1. Historical Risk Score
	riskFactors.HistoricalScore = s.calculateHistoricalScore(wallet, metrics)

	// 2. Velocity Score
	riskFactors.VelocityScore = s.calculateVelocityScore(wallet, metrics)

	// 3. Pattern Score
	riskFactors.PatternScore = s.calculatePatternScore(wallet, metrics)

	// 4. Link Analysis Score
	riskFactors.LinkScore = s.calculateLinkScore(wallet)

	// 5. Sanctions Score
	riskFactors.SanctionsScore = s.calculateSanctionsScore(wallet)

	// 6. Behavioral Score
	riskFactors.BehavioralScore = s.calculateBehavioralScore(wallet, metrics)

	// Calculate weighted total score
	riskFactors.TotalScore = s.calculateWeightedScore(riskFactors)

	// Get triggered rules
	riskFactors.TriggeredRules = s.getTriggeredRules(riskFactors)

	return riskFactors, nil
}

// calculateHistoricalScore evaluates historical risk factors
func (s *RiskScoringService) calculateHistoricalScore(wallet *models.Wallet, metrics *models.WalletMetrics) float64 {
	score := 0.0

	// Wallet age factor (newer wallets are riskier)
	walletAgeDays := int(time.Since(wallet.FirstSeen).Hours() / 24)
	if walletAgeDays < 7 {
		score += 25
	} else if walletAgeDays < 30 {
		score += 15
	} else if walletAgeDays < 90 {
		score += 10
	}

	// Transaction count factor
	if wallet.TxCount < 5 {
		score += 20
	} else if wallet.TxCount < 20 {
		score += 10
	}

	// Volume factor (large volumes are riskier)
	if wallet.TotalReceived.GreaterThan(decimal.NewFromFloat(1000000)) {
		score += 25
	} else if wallet.TotalReceived.GreaterThan(decimal.NewFromFloat(100000)) {
		score += 15
	} else if wallet.TotalReceived.GreaterThan(decimal.NewFromFloat(10000)) {
		score += 5
	}

	return math.Min(score, 50)
}

// calculateVelocityScore evaluates transaction velocity
func (s *RiskScoringService) calculateVelocityScore(wallet *models.Wallet, metrics *models.WalletMetrics) float64 {
	score := 0.0

	// 24h velocity
	velocity24h := metrics.Velocity24h
	if velocity24h.GreaterThan(decimal.NewFromFloat(100000)) {
		score += 30
	} else if velocity24h.GreaterThan(decimal.NewFromFloat(10000)) {
		score += 20
	} else if velocity24h.GreaterThan(decimal.NewFromFloat(1000)) {
		score += 10
	}

	// Transaction frequency
	if wallet.TxCount > 100 {
		score += 15
	} else if wallet.TxCount > 50 {
		score += 10
	}

	// Active days in last 30
	if metrics.ActiveDaysLast30 > 25 {
		score += 15
	} else if metrics.ActiveDaysLast30 > 15 {
		score += 10
	} else if metrics.ActiveDaysLast30 > 5 {
		score += 5
	}

	return math.Min(score, 60)
}

// calculatePatternScore evaluates transaction patterns
func (s *RiskScoringService) calculatePatternScore(wallet *models.Wallet, metrics *models.WalletMetrics) float64 {
	score := 0.0

	// Check for round amounts (potential structuring)
	avgTxSize := metrics.AverageTxSize
	roundAmountThreshold := decimal.NewFromFloat(0.95)

	// Check for unusual patterns
	if metrics.IncomingRatio > 0.9 {
		score += 20 // Mostly receiving (potential collecting for laundering)
	} else if metrics.OutgoingRatio > 0.9 {
		score += 15 // Mostly sending (potential distribution)
	}

	// Check for large transactions
	maxTx := metrics.MaxTxSize
	if maxTx.GreaterThan(decimal.NewFromFloat(500000)) {
		score += 25
	} else if maxTx.GreaterThan(decimal.NewFromFloat(100000)) {
		score += 15
	} else if maxTx.GreaterThan(decimal.NewFromFloat(10000)) {
		score += 5
	}

	return math.Min(score, 50)
}

// calculateLinkScore evaluates links to risky entities
func (s *RiskScoringService) calculateLinkScore(wallet *models.Wallet) float64 {
	score := 0.0

	if wallet.IsBlacklisted {
		return 100
	}

	if wallet.ClusterID != nil {
		cluster, _ := s.repo.GetClusterByID(context.Background(), *wallet.ClusterID)
		if cluster != nil {
			if cluster.RiskLevel == "critical" {
				score += 60
			} else if cluster.RiskLevel == "high" {
				score += 40
			} else if cluster.RiskLevel == "medium" {
				score += 20
			}
		}
	}

	return math.Min(score, 60)
}

// calculateSanctionsScore checks sanctions status
func (s *RiskScoringService) calculateSanctionsScore(wallet *models.Wallet) float64 {
	if wallet.IsSanctioned {
		return 100
	}
	return 0
}

// calculateBehavioralScore evaluates behavioral patterns
func (s *RiskScoringService) calculateBehavioralScore(wallet *models.Wallet, metrics *models.WalletMetrics) float64 {
	score := 0.0

	// Check for标签
	for _, tag := range []string{"mixer", "tumbler", "darknet", "gambling"} {
		// This would check against wallet tags
		_ = tag
	}

	// Check unique counterparties
	if metrics.UniqueCounterparties < 2 {
		score += 20 // Very limited interaction
	} else if metrics.UniqueCounterparties > 100 {
		score += 15 // High interaction
	}

	return math.Min(score, 35)
}

// calculateWeightedScore calculates the final weighted risk score
func (s *RiskScoringService) calculateWeightedScore(factors *models.RiskFactors) float64 {
	weights := map[string]float64{
		"historical":   0.25,
		"velocity":     0.25,
		"pattern":      0.20,
		"link":         0.15,
		"sanctions":    0.10,
		"behavioral":   0.05,
	}

	totalScore := factors.HistoricalScore*weights["historical"] +
		factors.VelocityScore*weights["velocity"] +
		factors.PatternScore*weights["pattern"] +
		factors.LinkScore*weights["link"] +
		factors.SanctionsScore*weights["sanctions"] +
		factors.BehavioralScore*weights["behavioral"]

	return math.Round(totalScore*100) / 100
}

// getTriggeredRules returns list of risk rules that were triggered
func (s *RiskScoringService) getTriggeredRules(factors *models.RiskFactors) []string {
	var triggered []string

	if factors.HistoricalScore > 30 {
		triggered = append(triggered, "historical_risk_high")
	}
	if factors.VelocityScore > 30 {
		triggered = append(triggered, "velocity_risk_high")
	}
	if factors.PatternScore > 20 {
		triggered = append(triggered, "pattern_anomaly_detected")
	}
	if factors.LinkScore > 30 {
		triggered = append(triggered, "linked_to_risky_entity")
	}
	if factors.SanctionsScore > 0 {
		triggered = append(triggered, "sanctions_match")
	}

	return triggered
}

// UpdateWalletRiskScore updates the risk score for a wallet
func (s *RiskScoringService) UpdateWalletRiskScore(ctx context.Context, address string, network models.Network) error {
	wallet, err := s.repo.GetWalletByAddress(ctx, address, network)
	if err != nil {
		return err
	}

	factors, err := s.CalculateRiskScore(ctx, address, network)
	if err != nil {
		return err
	}

	wallet.RiskScore = factors.TotalScore

	// Update risk level
	switch {
	case factors.TotalScore >= s.cfg.RiskScoring.CriticalThreshold:
		wallet.RiskLevel = "critical"
	case factors.TotalScore >= s.cfg.RiskScoring.HighThreshold:
		wallet.RiskLevel = "high"
	case factors.TotalScore >= s.cfg.RiskScoring.MediumThreshold:
		wallet.RiskLevel = "medium"
	default:
		wallet.RiskLevel = "low"
	}

	// Save to database
	if err := s.repo.UpdateWalletRiskScore(ctx, wallet); err != nil {
		return err
	}

	// Cache the score
	if err := s.cache.SetWalletRiskScore(ctx, address, network, factors.TotalScore); err != nil {
		s.logger.Warn("Failed to cache risk score", zap.Error(err))
	}

	// Trigger alert if above threshold
	if factors.TotalScore >= float64(s.cfg.RiskScoring.HighThreshold) {
		s.triggerHighRiskAlert(ctx, wallet, factors)
	}

	return nil
}

// triggerHighRiskAlert creates an alert for high-risk wallet
func (s *RiskScoringService) triggerHighRiskAlert(ctx context.Context, wallet *models.Wallet, factors *models.RiskFactors) {
	alert := models.NewAlert(
		models.AlertTypeHighRiskScore,
		models.SeverityHigh,
		"wallet",
		wallet.ID,
		fmt.Sprintf("High risk wallet detected: %s", wallet.Address),
	)

	alert.RuleName = "risk_score_threshold"
	alert.Score = factors.TotalScore
	alert.Evidence = factors.TriggeredRules

	if err := s.repo.CreateAlert(ctx, alert); err != nil {
		s.logger.Error("Failed to create high risk alert", zap.Error(err))
	}
}

// GetRiskScore returns cached or calculated risk score
func (s *RiskScoringService) GetRiskScore(ctx context.Context, address string, network models.Network) (float64, error) {
	// Try cache first
	score, err := s.cache.GetWalletRiskScore(ctx, address, network)
	if err == nil {
		return score, nil
	}

	// Calculate and cache
	if err := s.UpdateWalletRiskScore(ctx, address, network); err != nil {
		return 0, err
	}

	return s.cache.GetWalletRiskScore(ctx, address, network)
}

// EvaluateTransactionRisk evaluates risk for a transaction
func (s *RiskScoringService) EvaluateTransactionRisk(ctx context.Context, tx *models.NormalizedTransaction) (float64, []string) {
	var riskScore float64
	var reasons []string

	// Check sender risk
	senderScore, _ := s.GetRiskScore(ctx, tx.Inputs[0].Address, tx.Network)
	riskScore += senderScore * 0.4

	// Check recipient risk
	for _, output := range tx.Outputs {
		if output.IsChange {
			continue
		}
		recipientScore, _ := s.GetRiskScore(ctx, output.Address, tx.Network)
		riskScore += recipientScore * 0.3
	}

	// Transaction amount risk
	if tx.TotalValue.GreaterThan(decimal.NewFromFloat(100000)) {
		riskScore += 20
		reasons = append(reasons, "large_transaction_amount")
	}

	// Check for structuring
	if s.detectStructuring(tx) {
		riskScore += 30
		reasons = append(reasons, "potential_structuring_detected")
	}

	// Check for layering
	if s.detectLayering(tx) {
		riskScore += 25
		reasons = append(reasons, "potential_layering_detected")
	}

	return math.Min(riskScore, 100), reasons
}

// detectStructuring detects potential structuring patterns
func (s *RiskScoringService) detectStructuring(tx *models.NormalizedTransaction) bool {
	// Check for round amounts near reporting thresholds
	thresholds := []float64{9000, 10000, 49000, 50000}
	for _, threshold := range thresholds {
		if tx.TotalValue.GreaterThan(decimal.NewFromFloat(threshold-100)) &&
			tx.TotalValue.LessThan(decimal.NewFromFloat(threshold+100)) {
			return true
		}
	}
	return false
}

// detectLayering detects potential layering patterns
func (s *RiskScoringService) detectLayering(tx *models.NormalizedTransaction) bool {
	// Multiple outputs with similar amounts
	if len(tx.Outputs) >= 3 {
		var amounts []float64
		for _, output := range tx.Outputs {
			f, _ := output.Amount.Float64()
			amounts = append(amounts, f)
		}
		// Check for similar amounts (potential distribution)
		avg := amounts[0]
		similarCount := 0
		for i := 1; i < len(amounts); i++ {
			if math.Abs(amounts[i]-avg)/avg < 0.1 {
				similarCount++
			}
		}
		if similarCount >= len(amounts)-1 {
			return true
		}
	}
	return false
}

// HealthCheck verifies the service is healthy
func (s *RiskScoringService) HealthCheck(ctx context.Context) error {
	return nil
}
