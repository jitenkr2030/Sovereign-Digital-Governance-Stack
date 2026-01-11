package services

import (
	"context"
	"fmt"
	"time"

	"github.com/csic/monitoring/internal/core/domain"
	"github.com/csic/monitoring/internal/core/ports"
	"go.uber.org/zap"
)

// TransactionService handles transaction processing and monitoring
type TransactionService struct {
	transactionRepo ports.TransactionRepository
	riskScorer      *RiskScoringService
	sanctionsRepo   ports.SanctionsRepository
	logger          *zap.Logger
}

// NewTransactionService creates a new transaction service
func NewTransactionService(
	transactionRepo ports.TransactionRepository,
	riskScorer *RiskScoringService,
	sanctionsRepo ports.SanctionsRepository,
	logger *zap.Logger,
) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		riskScorer:      riskScorer,
		sanctionsRepo:   sanctionsRepo,
		logger:          logger,
	}
}

// IngestTransaction processes and stores a new transaction
func (s *TransactionService) IngestTransaction(ctx context.Context, tx *domain.Transaction) (*domain.Transaction, error) {
	// Set default values
	if tx.ID == "" {
		tx.ID = fmt.Sprintf("tx_%d", time.Now().UnixNano())
	}
	tx.CreatedAt = time.Now().UTC()

	// Calculate risk score
	riskAssessment, err := s.riskScorer.CalculateRiskScore(ctx, tx)
	if err != nil {
		s.logger.Error("Failed to calculate risk score", zap.Error(err))
		return nil, fmt.Errorf("failed to calculate risk score: %w", err)
	}

	// Apply risk assessment to transaction
	tx.RiskScore = riskAssessment.OverallScore
	tx.RiskFactors = riskAssessment.Factors
	tx.Flagged = riskAssessment.Flagged
	if riskAssessment.Flagged {
		reason := "High risk score detected"
		tx.FlagReason = &reason
	}

	// Store transaction
	if err := s.transactionRepo.Create(ctx, tx); err != nil {
		s.logger.Error("Failed to store transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to store transaction: %w", err)
	}

	s.logger.Info("Transaction ingested",
		zap.String("tx_hash", tx.TxHash),
		zap.String("chain", tx.Chain),
		zap.Int("risk_score", tx.RiskScore),
		zap.Bool("flagged", tx.Flagged),
	)

	return tx, nil
}

// GetTransactionByHash retrieves a transaction by its hash
func (s *TransactionService) GetTransactionByHash(ctx context.Context, txHash string) (*domain.Transaction, error) {
	return s.transactionRepo.GetByHash(ctx, txHash)
}

// GetTransactionHistory retrieves transactions based on filter criteria
func (s *TransactionService) GetTransactionHistory(ctx context.Context, filter *domain.TransactionFilter) (*domain.PaginationResult, error) {
	// Set defaults
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	transactions, total, err := s.transactionRepo.List(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to list transactions", zap.Error(err))
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &domain.PaginationResult{
		Items:      transactions,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetTransactionRisk retrieves the risk assessment for a specific transaction
func (s *TransactionService) GetTransactionRisk(ctx context.Context, txHash string) (*domain.RiskAssessment, error) {
	tx, err := s.transactionRepo.GetByHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}
	if tx == nil {
		return nil, fmt.Errorf("transaction not found")
	}

	riskAssessment := &domain.RiskAssessment{
		OverallScore: tx.RiskScore,
		RiskLevel:    s.riskScorer.calculateRiskLevel(tx.RiskScore),
		Factors:      tx.RiskFactors,
		Flagged:      tx.Flagged,
	}

	return riskAssessment, nil
}

// GetFlaggedTransactions retrieves all flagged transactions
func (s *TransactionService) GetFlaggedTransactions(ctx context.Context, page, pageSize int) (*domain.PaginationResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	flagged := true
	filter := &domain.TransactionFilter{
		Flagged:  &flagged,
		Page:     page,
		PageSize: pageSize,
	}

	return s.GetTransactionHistory(ctx, filter)
}

// ScanAddress performs a risk scan on a wallet address
func (s *TransactionService) ScanAddress(ctx context.Context, address, chain string) (*domain.WalletProfile, error) {
	// Get recent transactions for this address
	filter := &domain.TransactionFilter{
		Chain:       chain,
		Page:        1,
		PageSize:    100,
	}

	// Check both from and to addresses
	if filter.FromAddress == "" {
		filter.FromAddress = address
	}

	result, err := s.transactionRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions for address: %w", err)
	}

	// Create wallet profile
	profile := &domain.WalletProfile{
		Address:        address,
		Chain:          chain,
		TxCount:        len(result.Items.([]*domain.Transaction)),
		TotalVolumeUSD: 0,
		AvgTxValueUSD:  0,
		RiskIndicators: make([]domain.RiskIndicator, 0),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	// Calculate volume and risk indicators
	if txs, ok := result.Items.([]*domain.Transaction); ok {
		var totalUSD float64
		highRiskCount := 0
		for _, tx := range txs {
			totalUSD += tx.AmountUSD
			if tx.RiskScore >= 60 {
				highRiskCount++
			}
		}

		profile.TotalVolumeUSD = totalUSD
		if len(txs) > 0 {
			profile.AvgTxValueUSD = totalUSD / float64(len(txs))
		}

		// Add risk indicators
		if highRiskCount > 0 {
			profile.RiskIndicators = append(profile.RiskIndicators, domain.RiskIndicator{
				Indicator:    "HIGH_RISK_TRANSACTIONS",
				Severity:     "HIGH",
				Description:  fmt.Sprintf("%d high-risk transactions detected", highRiskCount),
				LastObserved: time.Now(),
				Count:        highRiskCount,
			})
		}
	}

	// Check sanctions
	sanctionsCheck := s.riskScorer.checkSanctions(address, chain)
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

// GetSuspiciousActivityReport generates a report of suspicious activity
func (s *TransactionService) GetSuspiciousActivityReport(ctx context.Context, startTime, endTime time.Time) (*domain.RiskReport, error) {
	flagged := true
	filter := &domain.TransactionFilter{
		Flagged:  &flagged,
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      1,
		PageSize:  10000, // Get all for reporting
	}

	result, err := s.transactionRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get flagged transactions: %w", err)
	}

	txs := result.Items.([]*domain.Transaction)

	report := &domain.RiskReport{
		PeriodStart:      startTime,
		PeriodEnd:        endTime,
		TotalTransactions: result.Total,
		FlaggedCount:     result.Total,
		TopRiskFactors:   make([]domain.RiskFactorSummary, 0),
		ByChain:          make(map[string]domain.ChainStats),
		GeneratedAt:      time.Now().UTC(),
	}

	// Calculate statistics
	var totalRiskScore float64
	riskFactorCount := make(map[string]int)

	for _, tx := range txs {
		report.TotalVolumeUSD += tx.AmountUSD
		totalRiskScore += float64(tx.RiskScore)

		// Count risk factors
		for _, factor := range tx.RiskFactors {
			riskFactorCount[factor.Type]++
		}

		// Aggregate by chain
		chain := tx.Chain
		if _, ok := report.ByChain[chain]; !ok {
			report.ByChain[chain] = domain.ChainStats{
				Transactions: 0,
				VolumeUSD:    0,
				Flagged:      0,
			}
		}
		stats := report.ByChain[chain]
		stats.Transactions++
		stats.VolumeUSD += tx.AmountUSD
		stats.Flagged++
		report.ByChain[chain] = stats
	}

	if result.Total > 0 {
		report.AverageRiskScore = totalRiskScore / float64(result.Total)
	}

	// Summarize risk factors
	totalFactors := 0
	for _, count := range riskFactorCount {
		totalFactors += count
	}
	for factorType, count := range riskFactorCount {
		report.TopRiskFactors = append(report.TopRiskFactors, domain.RiskFactorSummary{
			Type:       factorType,
			Count:      int64(count),
			Percentage: float64(count) / float64(totalFactors) * 100,
		})
	}

	return report, nil
}

// GetRiskSummaryReport generates a summary risk report
func (s *TransactionService) GetRiskSummaryReport(ctx context.Context, startTime, endTime time.Time) (*domain.RiskReport, error) {
	filter := &domain.TransactionFilter{
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      1,
		PageSize:  10000,
	}

	result, err := s.transactionRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	txs := result.Items.([]*domain.Transaction)

	report := &domain.RiskReport{
		PeriodStart:      startTime,
		PeriodEnd:        endTime,
		TotalTransactions: result.Total,
		TopRiskFactors:   make([]domain.RiskFactorSummary, 0),
		ByChain:          make(map[string]domain.ChainStats),
		GeneratedAt:      time.Now().UTC(),
	}

	var totalRiskScore float64
	var highRiskCount int64
	riskFactorCount := make(map[string]int)

	for _, tx := range txs {
		report.TotalVolumeUSD += tx.AmountUSD
		totalRiskScore += float64(tx.RiskScore)

		if tx.RiskScore >= 60 {
			highRiskCount++
		}

		// Count risk factors
		for _, factor := range tx.RiskFactors {
			riskFactorCount[factor.Type]++
		}

		// Aggregate by chain
		chain := tx.Chain
		if _, ok := report.ByChain[chain]; !ok {
			report.ByChain[chain] = domain.ChainStats{
				Transactions: 0,
				VolumeUSD:    0,
			}
		}
		stats := report.ByChain[chain]
		stats.Transactions++
		stats.VolumeUSD += tx.AmountUSD
		if tx.Flagged {
			stats.Flagged++
		}
		report.ByChain[chain] = stats
	}

	report.FlaggedCount = highRiskCount
	if result.Total > 0 {
		report.AverageRiskScore = totalRiskScore / float64(result.Total)
	}

	// Summarize risk factors
	totalFactors := 0
	for _, count := range riskFactorCount {
		totalFactors += count
	}
	for factorType, count := range riskFactorCount {
		report.TopRiskFactors = append(report.TopRiskFactors, domain.RiskFactorSummary{
			Type:       factorType,
			Count:      int64(count),
			Percentage: float64(count) / float64(totalFactors) * 100,
		})
	}

	return report, nil
}
