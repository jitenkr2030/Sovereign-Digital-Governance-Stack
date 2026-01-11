package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/csic-platform/services/transaction-monitoring/internal/core/domain"
	"github.com/csic-platform/services/transaction-monitoring/internal/core/ports"
	"go.uber.org/zap"
)

// WalletProfilingService handles wallet profiling and risk assessment
type WalletProfilingService struct {
	walletRepo       ports.WalletProfileRepository
	transactionRepo  ports.TransactionRepository
	logger           *zap.Logger
}

// NewWalletProfilingService creates a new wallet profiling service
func NewWalletProfilingService(
	walletRepo ports.WalletProfileRepository,
	transactionRepo ports.TransactionRepository,
	logger *zap.Logger,
) *WalletProfilingService {
	return &WalletProfilingService{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		logger:          logger,
	}
}

// GetOrCreateProfile retrieves or creates a wallet profile
func (s *WalletProfilingService) GetOrCreateProfile(ctx context.Context, address string) (*domain.WalletProfile, error) {
	return s.walletRepo.GetOrCreateWalletProfile(ctx, address)
}

// UpdateProfileRiskScore updates the risk score for a wallet
func (s *WalletProfilingService) UpdateProfileRiskScore(ctx context.Context, address string, riskScore float64) error {
	profile, err := s.walletRepo.GetWalletProfile(ctx, address)
	if err != nil {
		return err
	}

	// Calculate new weighted average
	oldScore := profile.CurrentRiskScore
	newScore := (oldScore*0.7 + riskScore*0.3)

	profile.CurrentRiskScore = newScore
	profile.LastSeenAt = time.Now()

	if newScore >= 75 {
		profile.RiskLevel = domain.RiskHigh
	} else if newScore >= 50 {
		profile.RiskLevel = domain.RiskMedium
	} else {
		profile.RiskLevel = domain.RiskLow
	}

	return s.walletRepo.UpdateWalletProfile(ctx, profile)
}

// GetWalletRiskProfile retrieves the full risk profile for a wallet
func (s *WalletProfilingService) GetWalletRiskProfile(ctx context.Context, address string) (*domain.WalletProfile, error) {
	profile, err := s.walletRepo.GetOrCreateWalletProfile(ctx, address)
	if err != nil {
		return nil, err
	}

	// Get recent transactions
	recentTx, err := s.walletRepo.GetWalletHistory(ctx, address, 100)
	if err != nil {
		s.logger.Warn("Failed to get wallet history", zap.Error(err))
	}

	// Calculate velocity metrics
	if len(recentTx) > 1 {
		var totalVolume float64
		var txCount int
		for _, tx := range recentTx {
			totalVolume += tx.AmountUSD
			txCount++
		}
		profile.AverageTransactionValue = totalVolume / float64(txCount)
		profile.TotalTransactions = int64(txCount)
		profile.TotalVolume = totalVolume
	}

	return profile, nil
}

// GetWalletNetworkGraph retrieves the transaction network for a wallet
func (s *WalletProfilingService) GetWalletNetworkGraph(ctx context.Context, address string, depth int) ([]*domain.Transaction, error) {
	if depth > 3 {
		depth = 3 // Limit depth to prevent excessive queries
	}

	transactions, err := s.transactionRepo.GetTransactionsByAddress(ctx, address, 1000, 0)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

// RiskScoringService handles risk scoring calculations
type RiskScoringService struct {
	walletRepo      ports.WalletProfileRepository
	transactionRepo ports.TransactionRepository
	ruleRepo        ports.MonitoringRuleRepository
	logger          *zap.Logger
}

// NewRiskScoringService creates a new risk scoring service
func NewRiskScoringService(
	walletRepo ports.WalletProfileRepository,
	transactionRepo ports.TransactionRepository,
	ruleRepo ports.MonitoringRuleRepository,
	logger *zap.Logger,
) *RiskScoringService {
	return &RiskScoringService{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		ruleRepo:        ruleRepo,
		logger:          logger,
	}
}

// CalculateTransactionRisk calculates risk score for a transaction
func (s *RiskScoringService) CalculateTransactionRisk(ctx context.Context, tx *domain.Transaction) (float64, []domain.RiskIndicator, error) {
	score := 0.0
	indicators := []domain.RiskIndicator{}

	// Amount-based risk
	if tx.AmountUSD > 100000 {
		score += 25
		indicators = append(indicators, domain.RiskIndicator{
			Indicator:   "LARGE_TRANSACTION",
			Severity:    "HIGH",
			Description: fmt.Sprintf("Transaction amount $%.2f exceeds $100,000", tx.AmountUSD),
			Count:       1,
		})
	} else if tx.AmountUSD > 10000 {
		score += 15
		indicators = append(indicators, domain.RiskIndicator{
			Indicator:   "MEDIUM_TRANSACTION",
			Severity:    "MEDIUM",
			Description: fmt.Sprintf("Transaction amount $%.2f exceeds $10,000", tx.AmountUSD),
			Count:       1,
		})
	}

	// Wallet age risk
	profile, err := s.walletRepo.GetOrCreateWalletProfile(ctx, tx.FromAddress)
	if err == nil {
		if profile.WalletAgeHours < 24 {
			score += 20
			indicators = append(indicators, domain.RiskIndicator{
				Indicator:    "NEW_WALLET",
				Severity:     "HIGH",
				Description:  "Wallet is less than 24 hours old",
				FirstObserved: time.Now(),
				LastObserved:  time.Now(),
				Count:         1,
			})
		}

		// Flag count risk
		if profile.FlagCount > 0 {
			score += float64(profile.FlagCount) * 5
			indicators = append(indicators, domain.RiskIndicator{
				Indicator:    "PREVIOUSLY_FLAGGED",
				Severity:     "MEDIUM",
				Description:  fmt.Sprintf("Wallet has been flagged %d times", profile.FlagCount),
				LastObserved: time.Now(),
				Count:        profile.FlagCount,
			})
		}
	}

	// Sanctioned address check
	if profile.IsSanctioned {
		score += 100
		indicators = append(indicators, domain.RiskIndicator{
			Indicator:    "SANCTIONED_ADDRESS",
			Severity:     "CRITICAL",
			Description:  "Address is on sanctions list",
			LastObserved: time.Now(),
			Count:        1,
		})
	}

	// Cap score at 100
	if score > 100 {
		score = 100
	}

	return score, indicators, nil
}

// CalculateWalletRisk calculates overall risk for a wallet
func (s *RiskScoringService) CalculateWalletRisk(ctx context.Context, address string) (float64, []domain.RiskIndicator, error) {
	profile, err := s.walletRepo.GetWalletProfile(ctx, address)
	if err != nil {
		return 0, nil, err
	}

	score := profile.CurrentRiskScore
	indicators := []domain.RiskIndicator{}

	// Transaction velocity risk
	velocityScore, velocityIndicators, err := s.CalculateVelocityRisk(ctx, address, 24*time.Hour)
	if err == nil {
		score += velocityScore
		indicators = append(indicators, velocityIndicators...)
	}

	// Update profile
	profile.CurrentRiskScore = score
	if err := s.walletRepo.UpdateWalletProfile(ctx, profile); err != nil {
		s.logger.Error("Failed to update wallet profile", zap.Error(err))
	}

	return score, indicators, nil
}

// CalculateVelocityRisk calculates risk based on transaction velocity
func (s *RiskScoringService) CalculateVelocityRisk(ctx context.Context, address string, timeWindow time.Duration) (float64, []domain.RiskIndicator, error) {
	now := time.Now()
	startTime := now.Add(-timeWindow)

	txs, err := s.transactionRepo.GetTransactionsByTimeRange(ctx, startTime, now)
	if err != nil {
		return 0, nil, err
	}

	if len(txs) == 0 {
		return 0, nil, nil
	}

	// Calculate velocity metrics
	var totalAmount float64
	for _, tx := range txs {
		totalAmount += tx.AmountUSD
	}

	// High velocity threshold (e.g., > 50 transactions or > $1M in 24h)
	if len(txs) > 50 || totalAmount > 1000000 {
		indicators := []domain.RiskIndicator{
			{
				Indicator:    "HIGH_VELOCITY",
				Severity:     "HIGH",
				Description:  fmt.Sprintf("%d transactions totaling $%.2f in time window", len(txs), totalAmount),
				FirstObserved: startTime,
				LastObserved:  now,
				Count:         len(txs),
			},
		}
		return 30, indicators, nil
	}

	return 0, nil, nil
}

// CalculatePatternRisk calculates risk based on transaction patterns
func (s *RiskScoringService) CalculatePatternRisk(ctx context.Context, tx *domain.Transaction) (float64, error) {
	// Check for round number patterns (potential structuring)
	if tx.AmountUSD > 0 {
		// Simple round number check
		if tx.AmountUSD == float64(int(tx.AmountUSD)) {
			// Could be round number, but not necessarily suspicious
			return 5, nil
		}
	}

	// Check for structuring patterns (multiple just-below-threshold transactions)
	// This would require historical analysis

	return 0, nil
}

// AlertService handles alert generation and management
type AlertService struct {
	alertRepo     ports.AlertRepository
	kafkaProducer interface{} // Would be actual Kafka producer type
	logger        *zap.Logger
}

// NewAlertService creates a new alert service
func NewAlertService(alertRepo ports.AlertRepository, kafkaProducer interface{}, logger *zap.Logger) *AlertService {
	return &AlertService{
		alertRepo:     alertRepo,
		kafkaProducer: kafkaProducer,
		logger:        logger,
	}
}

// GenerateAlert creates a new alert for suspicious activity
func (s *AlertService) GenerateAlert(ctx context.Context, alertType domain.AlertType, tx *domain.Transaction, riskScore float64, reason string) (*domain.Alert, error) {
	severity := domain.RuleSeverityInfo
	if riskScore >= 75 {
		severity = domain.RuleSeverityCritical
	} else if riskScore >= 50 {
		severity = domain.RuleSeverityAlert
	} else if riskScore >= 25 {
		severity = domain.RuleSeverityWarning
	}

	alert := &domain.Alert{
		ID:            uuid.New(),
		AlertType:     alertType,
		TransactionID: uuid.MustParse(tx.ID),
		WalletAddress: tx.FromAddress,
		Severity:      severity,
		RiskScore:     riskScore,
		Status:        domain.AlertStatusOpen,
		Title:         s.generateAlertTitle(alertType, riskScore),
		Description:   reason,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.alertRepo.CreateAlert(ctx, alert); err != nil {
		return nil, err
	}

	s.logger.Info("Alert generated",
		zap.String("alert_id", alert.ID.String()),
		zap.String("alert_type", string(alertType)),
		zap.Float64("risk_score", riskScore))

	return alert, nil
}

// ProcessHighRiskTransaction processes a high-risk transaction
func (s *AlertService) ProcessHighRiskTransaction(ctx context.Context, tx *domain.Transaction, riskScore float64) error {
	alertType := domain.AlertTypeHighRiskTx
	if riskScore >= 75 {
		alertType = domain.AlertTypeCritical
	}

	reason := fmt.Sprintf("Transaction flagged with risk score %.2f", riskScore)
	_, err := s.GenerateAlert(ctx, alertType, tx, riskScore, reason)
	return err
}

// GetOpenAlerts retrieves open alerts with pagination
func (s *AlertService) GetOpenAlerts(ctx context.Context, limit, offset int) ([]*domain.Alert, int64, error) {
	return s.alertRepo.ListAlerts(ctx, string(domain.AlertStatusOpen), "", limit, offset)
}

// GetAlertStats retrieves alert statistics
func (s *AlertService) GetAlertStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	openCount, err := s.alertRepo.CountAlertsByStatus(ctx, string(domain.AlertStatusOpen))
	if err != nil {
		return nil, err
	}
	stats["open"] = openCount

	inProgressCount, err := s.alertRepo.CountAlertsByStatus(ctx, string(domain.AlertStatusInProgress))
	if err != nil {
		return nil, err
	}
	stats["in_progress"] = inProgressCount

	resolvedCount, err := s.alertRepo.CountAlertsByStatus(ctx, string(domain.AlertStatusResolved))
	if err != nil {
		return nil, err
	}
	stats["resolved"] = resolvedCount

	return stats, nil
}

func (s *AlertService) generateAlertTitle(alertType domain.AlertType, riskScore float64) string {
	titles := map[domain.AlertType]string{
		domain.AlertTypeSanctionsMatch:    "Sanctions List Match Detected",
		domain.AlertTypeHighRiskTx:        "High-Risk Transaction Detected",
		domain.AlertTypeVelocityBreach:    "Transaction Velocity Threshold Exceeded",
		domain.AlertTypePatternMatch:      "Suspicious Transaction Pattern Detected",
		domain.AlertTypeUnusualActivity:   "Unusual Activity Detected",
		domain.AlertTypeWalletFlagged:     "Wallet Flagged for Review",
		domain.AlertTypeLargeTransaction:  "Large Transaction Detected",
		domain.AlertTypeGeographicRisk:    "High-Risk Geographic Region",
	}

	if title, ok := titles[alertType]; ok {
		return title
	}
	return "Alert Generated"
}

// RuleEngineService handles monitoring rule evaluation
type RuleEngineService struct {
	ruleRepo ports.MonitoringRuleRepository
	logger   *zap.Logger
}

// NewRuleEngineService creates a new rule engine service
func NewRuleEngineService(ruleRepo ports.MonitoringRuleRepository, logger *zap.Logger) *RuleEngineService {
	return &RuleEngineService{
		ruleRepo: ruleRepo,
		logger:   logger,
	}
}

// EvaluateRules evaluates all active rules against a transaction
func (s *RuleEngineService) EvaluateRules(ctx context.Context, tx *domain.Transaction) ([]domain.RuleMatch, error) {
	rules, err := s.ruleRepo.GetActiveRules(ctx)
	if err != nil {
		return nil, err
	}

	matches := []domain.RuleMatch{}
	for _, rule := range rules {
		matched, detail, err := s.ExecuteRule(ctx, rule, tx)
		if err != nil {
			s.logger.Warn("Rule execution failed", zap.String("rule", rule.Name), zap.Error(err))
			continue
		}
		if matched {
			matches = append(matches, domain.RuleMatch{
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				RuleType:    rule.RuleType,
				Severity:    rule.Severity,
				RiskWeight:  rule.RiskWeight,
				MatchDetail: detail,
			})
		}
	}

	return matches, nil
}

// GetApplicableRules returns rules applicable to a transaction
func (s *RuleEngineService) GetApplicableRules(ctx context.Context, tx *domain.Transaction) ([]*domain.MonitoringRule, error) {
	return s.ruleRepo.GetActiveRules(ctx)
}

// ExecuteRule executes a single rule against a transaction
func (s *RuleEngineService) ExecuteRule(ctx context.Context, rule *domain.MonitoringRule, tx *domain.Transaction) (bool, string, error) {
	var condition map[string]interface{}
	if err := json.Unmarshal([]byte(rule.Condition), &condition); err != nil {
		return false, "", err
	}

	switch rule.RuleType {
	case domain.RuleTypeThreshold:
		return s.executeThresholdRule(condition, tx)
	case domain.RuleTypeVelocity:
		return s.executeVelocityRule(condition, tx)
	case domain.RuleTypeSanctions:
		return s.executeSanctionsRule(condition, tx)
	default:
		return false, "", nil
	}
}

func (s *RuleEngineService) executeThresholdRule(condition map[string]interface{}, tx *domain.Transaction) (bool, string, error) {
	if maxAmount, ok := condition["max_amount"].(float64); ok {
		if tx.AmountUSD > maxAmount {
			return true, fmt.Sprintf("Amount $%.2f exceeds threshold $%.2f", tx.AmountUSD, maxAmount), nil
		}
	}
	return false, "", nil
}

func (s *RuleEngineService) executeVelocityRule(condition map[string]interface{}, tx *domain.Transaction) (bool, string, error) {
	// Velocity rule implementation
	return false, "", nil
}

func (s *RuleEngineService) executeSanctionsRule(condition map[string]interface{}, tx *domain.Transaction) (bool, string, error) {
	// Sanctions rule is handled separately
	return false, "", nil
}
