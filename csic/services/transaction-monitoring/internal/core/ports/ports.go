package ports

import (
	"context"

	"github.com/csic-platform/services/transaction-monitoring/internal/core/domain"
)

// TransactionRepository interface for transaction data access
type TransactionRepository interface {
	CreateTransaction(ctx context.Context, tx *domain.Transaction) error
	GetTransaction(ctx context.Context, id string) (*domain.Transaction, error)
	GetTransactionByHash(ctx context.Context, txHash string) (*domain.Transaction, error)
	UpdateTransaction(ctx context.Context, tx *domain.Transaction) error
	ListTransactions(ctx context.Context, filter domain.TransactionFilter) ([]*domain.Transaction, int64, error)
	GetTransactionsByAddress(ctx context.Context, address string, limit, offset int) ([]*domain.Transaction, error)
	GetTransactionsByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.Transaction, error)
	CountTransactions(ctx context.Context, filter domain.TransactionFilter) (int64, error)
}

// WalletProfileRepository interface for wallet profile data access
type WalletProfileRepository interface {
	CreateWalletProfile(ctx context.Context, profile *domain.WalletProfile) error
	GetWalletProfile(ctx context.Context, address string) (*domain.WalletProfile, error)
	UpdateWalletProfile(ctx context.Context, profile *domain.WalletProfile) error
	GetOrCreateWalletProfile(ctx context.Context, address string) (*domain.WalletProfile, error)
	GetHighRiskWallets(ctx context.Context, limit int) ([]*domain.WalletProfile, error)
	GetWalletHistory(ctx context.Context, address string, limit int) ([]*domain.Transaction, error)
	IncrementFlagCount(ctx context.Context, address string) error
}

// SanctionsRepository interface for sanctions list data access
type SanctionsRepository interface {
	CheckAddress(ctx context.Context, address string) (*domain.SanctionedAddress, error)
	GetAllSanctions(ctx context.Context) ([]*domain.SanctionedAddress, error)
	GetSanctionsByChain(ctx context.Context, chain string) ([]*domain.SanctionedAddress, error)
	ImportSanctions(ctx context.Context, addresses []domain.SanctionedAddress) error
	DeactivateSanction(ctx context.Context, id string) error
	SearchSanctions(ctx context.Context, query string) ([]*domain.SanctionedAddress, error)
}

// AlertRepository interface for alert data access
type AlertRepository interface {
	CreateAlert(ctx context.Context, alert *domain.Alert) error
	GetAlert(ctx context.Context, id string) (*domain.Alert, error)
	UpdateAlert(ctx context.Context, alert *domain.Alert) error
	ListAlerts(ctx context.Context, status string, severity string, limit, offset int) ([]*domain.Alert, int64, error)
	GetAlertsByTransaction(ctx context.Context, txID string) ([]*domain.Alert, error)
	GetAlertsByWallet(ctx context.Context, walletAddress string, limit int) ([]*domain.Alert, error)
	CountAlertsByStatus(ctx context.Context, status string) (int64, error)
	AssignAlert(ctx context.Context, alertID, userID string) error
	ResolveAlert(ctx context.Context, alertID, resolution string) error
}

// MonitoringRuleRepository interface for monitoring rule data access
type MonitoringRuleRepository interface {
	GetActiveRules(ctx context.Context) ([]*domain.MonitoringRule, error)
	GetRule(ctx context.Context, id string) (*domain.MonitoringRule, error)
	CreateRule(ctx context.Context, rule *domain.MonitoringRule) error
	UpdateRule(ctx context.Context, rule *domain.MonitoringRule) error
	DeleteRule(ctx context.Context, id string) error
	GetRulesByType(ctx context.Context, ruleType string) ([]*domain.MonitoringRule, error)
}

// TransactionAnalysisService interface for transaction analysis
type TransactionAnalysisService interface {
	AnalyzeTransaction(ctx context.Context, tx *domain.Transaction) (*domain.TransactionAnalysisResult, error)
	AnalyzeTransactionSync(ctx context.Context, txHash string) (*domain.RiskAssessment, error)
	ProcessTransactionStream(ctx context.Context, txs []*domain.Transaction) error
}

// WalletProfilingService interface for wallet profiling
type WalletProfilingService interface {
	GetOrCreateProfile(ctx context.Context, address string) (*domain.WalletProfile, error)
	UpdateProfileRiskScore(ctx context.Context, address string, riskScore float64) error
	GetWalletRiskProfile(ctx context.Context, address string) (*domain.WalletProfile, error)
	GetWalletNetworkGraph(ctx context.Context, address string, depth int) ([]*domain.Transaction, error)
}

// RiskScoringService interface for risk scoring
type RiskScoringService interface {
	CalculateTransactionRisk(ctx context.Context, tx *domain.Transaction) (float64, []domain.RiskFactor, error)
	CalculateWalletRisk(ctx context.Context, address string) (float64, []domain.RiskIndicator, error)
	CalculateVelocityRisk(ctx context.Context, address string, timeWindow time.Duration) (float64, error)
	CalculatePatternRisk(ctx context.Context, tx *domain.Transaction) (float64, error)
}

// AlertService interface for alert generation and management
type AlertService interface {
	GenerateAlert(ctx context.Context, alertType domain.AlertType, tx *domain.Transaction, riskScore float64, reason string) (*domain.Alert, error)
	ProcessHighRiskTransaction(ctx context.Context, tx *domain.Transaction, riskScore float64) error
	GetOpenAlerts(ctx context.Context, limit, offset int) ([]*domain.Alert, int64, error)
	GetAlertStats(ctx context.Context) (map[string]int64, error)
}

// RuleEngineService interface for monitoring rule engine
type RuleEngineService interface {
	EvaluateRules(ctx context.Context, tx *domain.Transaction) ([]domain.RuleMatch, error)
	GetApplicableRules(ctx context.Context, tx *domain.Transaction) ([]*domain.MonitoringRule, error)
	ExecuteRule(ctx context.Context, rule *domain.MonitoringRule, tx *domain.Transaction) (bool, string, error)
}
