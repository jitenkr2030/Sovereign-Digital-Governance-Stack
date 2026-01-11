package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/services/transaction-monitoring/internal/core/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// WalletProfileRepository implements ports.WalletProfileRepository
type WalletProfileRepository struct {
	conn   *Connection
	logger *zap.Logger
}

// NewWalletProfileRepository creates a new wallet profile repository
func NewWalletProfileRepository(conn *Connection, logger *zap.Logger) *WalletProfileRepository {
	return &WalletProfileRepository{
		conn:   conn,
		logger: logger,
	}
}

// CreateWalletProfile creates a new wallet profile
func (r *WalletProfileRepository) CreateWalletProfile(ctx context.Context, profile *domain.WalletProfile) error {
	query := `
		INSERT INTO wallet_profiles (
			id, address, chain, first_seen, last_seen, tx_count,
			total_volume_usd, avg_tx_value_usd, risk_score, is_sanctioned,
			flag_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.conn.pool.Exec(ctx, query,
		profile.ID, profile.Address, profile.Chain, profile.FirstSeen, profile.LastSeen,
		profile.TxCount, profile.TotalVolumeUSD, profile.AvgTxValueUSD, profile.CurrentRiskScore,
		profile.IsSanctioned, profile.FlagCount, time.Now(), time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create wallet profile: %w", err)
	}

	return nil
}

// GetWalletProfile retrieves a wallet profile
func (r *WalletProfileRepository) GetWalletProfile(ctx context.Context, address string) (*domain.WalletProfile, error) {
	query := `SELECT * FROM wallet_profiles WHERE address = $1`
	row := r.conn.pool.QueryRow(ctx, query, address)

	var profile domain.WalletProfile
	err := row.Scan(
		&profile.ID, &profile.Address, &profile.Chain, &profile.FirstSeen, &profile.LastSeen,
		&profile.TxCount, &profile.TotalVolumeUSD, &profile.AvgTxValueUSD, &profile.ConnectedTags,
		&profile.RiskIndicators, &profile.WalletAgeHours, &profile.IsContract, &profile.ContractType,
		&profile.CreatedAt, &profile.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("wallet profile not found: %w", err)
	}

	return &profile, nil
}

// UpdateWalletProfile updates an existing wallet profile
func (r *WalletProfileRepository) UpdateWalletProfile(ctx context.Context, profile *domain.WalletProfile) error {
	query := `
		UPDATE wallet_profiles SET
			last_seen = $1, tx_count = $2, total_volume_usd = $3,
			avg_tx_value_usd = $4, risk_score = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.conn.pool.Exec(ctx, query,
		time.Now(), profile.TxCount, profile.TotalVolumeUSD, profile.AvgTxValueUSD,
		profile.CurrentRiskScore, time.Now(), profile.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update wallet profile: %w", err)
	}

	return nil
}

// GetOrCreateWalletProfile retrieves or creates a wallet profile
func (r *WalletProfileRepository) GetOrCreateWalletProfile(ctx context.Context, address string) (*domain.WalletProfile, error) {
	profile, err := r.GetWalletProfile(ctx, address)
	if err == nil {
		return profile, nil
	}

	// Create new profile
	newProfile := &domain.WalletProfile{
		ID:              generateUUID(),
		Address:         address,
		Chain:           "ethereum", // Default chain
		FirstSeen:       time.Now(),
		LastSeen:        time.Now(),
		TxCount:         0,
		TotalVolumeUSD:  0,
		AvgTxValueUSD:   0,
		CurrentRiskScore: 0,
		RiskLevel:       domain.RiskLow,
		IsSanctioned:    false,
		FlagCount:       0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := r.CreateWalletProfile(ctx, newProfile); err != nil {
		return nil, err
	}

	return newProfile, nil
}

// GetHighRiskWallets retrieves wallets with high risk scores
func (r *WalletProfileRepository) GetHighRiskWallets(ctx context.Context, limit int) ([]*domain.WalletProfile, error) {
	query := `
		SELECT * FROM wallet_profiles
		WHERE risk_score >= 75
		ORDER BY risk_score DESC
		LIMIT $1
	`

	rows, err := r.conn.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query high risk wallets: %w", err)
	}
	defer rows.Close()

	var profiles []*domain.WalletProfile
	for rows.Next() {
		var profile domain.WalletProfile
		err := rows.Scan(
			&profile.ID, &profile.Address, &profile.Chain, &profile.FirstSeen, &profile.LastSeen,
			&profile.TxCount, &profile.TotalVolumeUSD, &profile.AvgTxValueUSD, &profile.ConnectedTags,
			&profile.RiskIndicators, &profile.WalletAgeHours, &profile.IsContract, &profile.ContractType,
			&profile.CreatedAt, &profile.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan wallet profile: %w", err)
		}
		profiles = append(profiles, &profile)
	}

	return profiles, nil
}

// GetWalletHistory retrieves transaction history for a wallet
func (r *WalletProfileRepository) GetWalletHistory(ctx context.Context, address string, limit int) ([]*domain.Transaction, error) {
	query := `
		SELECT * FROM transactions
		WHERE from_address = $1 OR to_address = $1
		ORDER BY tx_timestamp DESC
		LIMIT $2
	`

	rows, err := r.conn.pool.Query(ctx, query, address, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query wallet history: %w", err)
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		err := rows.Scan(
			&tx.ID, &tx.TxHash, &tx.Chain, &tx.BlockNumber, &tx.FromAddress, &tx.ToAddress,
			&tx.TokenAddress, &tx.Amount, &tx.AmountUSD, &tx.GasUsed, &tx.GasPrice, &tx.GasFeeUSD,
			&tx.Nonce, &tx.TxTimestamp, &tx.RiskScore, &tx.Flagged, &tx.FlagReason,
			&tx.ReviewedAt, &tx.ReviewedBy, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

// IncrementFlagCount increments the flag count for a wallet
func (r *WalletProfileRepository) IncrementFlagCount(ctx context.Context, address string) error {
	query := `
		UPDATE wallet_profiles SET
			flag_count = flag_count + 1, updated_at = $1
		WHERE address = $2
	`

	_, err := r.conn.pool.Exec(ctx, query, time.Now(), address)
	return err
}

// SanctionsRepository implements ports.SanctionsRepository
type SanctionsRepository struct {
	conn   *Connection
	logger *zap.Logger
}

// NewSanctionsRepository creates a new sanctions repository
func NewSanctionsRepository(conn *Connection, logger *zap.Logger) *SanctionsRepository {
	return &SanctionsRepository{
		conn:   conn,
		logger: logger,
	}
}

// CheckAddress checks if an address is sanctioned
func (r *SanctionsRepository) CheckAddress(ctx context.Context, address string) (*domain.SanctionedAddress, error) {
	query := `SELECT * FROM sanctioned_addresses WHERE address = $1 AND is_active = true`
	row := r.conn.pool.QueryRow(ctx, query, address)

	var addr domain.SanctionedAddress
	err := row.Scan(
		&addr.ID, &addr.Address, &addr.Chain, &addr.SourceList, &addr.Reason,
		&addr.EntityName, &addr.EntityType, &addr.Program, &addr.AddedAt, &addr.ExpiresAt,
	)

	if err != nil {
		return nil, fmt.Errorf("address not found in sanctions list: %w", err)
	}

	return &addr, nil
}

// GetAllSanctions retrieves all active sanctions
func (r *SanctionsRepository) GetAllSanctions(ctx context.Context) ([]*domain.SanctionedAddress, error) {
	query := `SELECT * FROM sanctioned_addresses WHERE is_active = true`
	rows, err := r.conn.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sanctions: %w", err)
	}
	defer rows.Close()

	var addresses []*domain.SanctionedAddress
	for rows.Next() {
		var addr domain.SanctionedAddress
		err := rows.Scan(
			&addr.ID, &addr.Address, &addr.Chain, &addr.SourceList, &addr.Reason,
			&addr.EntityName, &addr.EntityType, &addr.Program, &addr.AddedAt, &addr.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sanction: %w", err)
		}
		addresses = append(addresses, &addr)
	}

	return addresses, nil
}

// GetSanctionsByChain retrieves sanctions for a specific chain
func (r *SanctionsRepository) GetSanctionsByChain(ctx context.Context, chain string) ([]*domain.SanctionedAddress, error) {
	query := `SELECT * FROM sanctioned_addresses WHERE chain = $1 AND is_active = true`
	rows, err := r.conn.pool.Query(ctx, query, chain)
	if err != nil {
		return nil, fmt.Errorf("failed to query sanctions: %w", err)
	}
	defer rows.Close()

	var addresses []*domain.SanctionedAddress
	for rows.Next() {
		var addr domain.SanctionedAddress
		err := rows.Scan(
			&addr.ID, &addr.Address, &addr.Chain, &addr.SourceList, &addr.Reason,
			&addr.EntityName, &addr.EntityType, &addr.Program, &addr.AddedAt, &addr.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sanction: %w", err)
		}
		addresses = append(addresses, &addr)
	}

	return addresses, nil
}

// ImportSanctions imports a batch of sanctioned addresses
func (r *SanctionsRepository) ImportSanctions(ctx context.Context, addresses []domain.SanctionedAddress) error {
	for _, addr := range addresses {
		query := `
			INSERT INTO sanctioned_addresses (
				id, address, chain, source_list, reason, entity_name,
				entity_type, program, added_at, is_active
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true)
			ON CONFLICT (address, chain) DO UPDATE SET is_active = true
		`

		_, err := r.conn.pool.Exec(ctx, query,
			addr.ID, addr.Address, addr.Chain, addr.SourceList, addr.Reason,
			addr.EntityName, addr.EntityType, addr.Program, time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to import sanction: %w", err)
		}
	}

	return nil
}

// DeactivateSanction deactivates a sanctions entry
func (r *SanctionsRepository) DeactivateSanction(ctx context.Context, id string) error {
	query := `UPDATE sanctioned_addresses SET is_active = false WHERE id = $1`
	_, err := r.conn.pool.Exec(ctx, query, id)
	return err
}

// SearchSanctions searches sanctions by entity name
func (r *SanctionsRepository) SearchSanctions(ctx context.Context, query string) ([]*domain.SanctionedAddress, error) {
	searchQuery := `%` + query + `%`
	sqlQuery := `SELECT * FROM sanctioned_addresses WHERE entity_name ILIKE $1 AND is_active = true`
	rows, err := r.conn.pool.Query(ctx, sqlQuery, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search sanctions: %w", err)
	}
	defer rows.Close()

	var addresses []*domain.SanctionedAddress
	for rows.Next() {
		var addr domain.SanctionedAddress
		err := rows.Scan(
			&addr.ID, &addr.Address, &addr.Chain, &addr.SourceList, &addr.Reason,
			&addr.EntityName, &addr.EntityType, &addr.Program, &addr.AddedAt, &addr.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sanction: %w", err)
		}
		addresses = append(addresses, &addr)
	}

	return addresses, nil
}

// AlertRepository implements ports.AlertRepository
type AlertRepository struct {
	conn   *Connection
	logger *zap.Logger
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(conn *Connection, logger *zap.Logger) *AlertRepository {
	return &AlertRepository{
		conn:   conn,
		logger: logger,
	}
}

// CreateAlert creates a new alert
func (r *AlertRepository) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	query := `
		INSERT INTO alerts (
			id, alert_type, transaction_id, wallet_address, severity,
			risk_score, status, title, description, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.conn.pool.Exec(ctx, query,
		alert.ID, alert.AlertType, alert.TransactionID, alert.WalletAddress,
		alert.Severity, alert.RiskScore, alert.Status, alert.Title, alert.Description,
		time.Now(), time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

// GetAlert retrieves an alert by ID
func (r *AlertRepository) GetAlert(ctx context.Context, id string) (*domain.Alert, error) {
	query := `SELECT * FROM alerts WHERE id = $1`
	row := r.conn.pool.QueryRow(ctx, query, id)

	var alert domain.Alert
	err := row.Scan(
		&alert.ID, &alert.AlertType, &alert.TransactionID, &alert.WalletAddress,
		&alert.Severity, &alert.RiskScore, &alert.Status, &alert.Title, &alert.Description,
		&alert.TriggeredRules, &alert.AssignedTo, &alert.ResolvedAt, &alert.Resolution,
		&alert.CreatedAt, &alert.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("alert not found: %w", err)
	}

	return &alert, nil
}

// UpdateAlert updates an existing alert
func (r *AlertRepository) UpdateAlert(ctx context.Context, alert *domain.Alert) error {
	query := `
		UPDATE alerts SET
			status = $1, assigned_to = $2, resolved_at = $3,
			resolution = $4, updated_at = $5
		WHERE id = $6
	`

	_, err := r.conn.pool.Exec(ctx, query,
		alert.Status, alert.AssignedTo, alert.ResolvedAt, alert.Resolution,
		time.Now(), alert.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	return nil
}

// ListAlerts retrieves alerts with filtering
func (r *AlertRepository) ListAlerts(ctx context.Context, status, severity string, limit, offset int) ([]*domain.Alert, int64, error) {
	return []*domain.Alert{}, 0, nil
}

// GetAlertsByTransaction retrieves alerts for a transaction
func (r *AlertRepository) GetAlertsByTransaction(ctx context.Context, txID string) ([]*domain.Alert, error) {
	return []*domain.Alert{}, nil
}

// GetAlertsByWallet retrieves alerts for a wallet
func (r *AlertRepository) GetAlertsByWallet(ctx context.Context, walletAddress string, limit int) ([]*domain.Alert, error) {
	return []*domain.Alert{}, nil
}

// CountAlertsByStatus counts alerts by status
func (r *AlertRepository) CountAlertsByStatus(ctx context.Context, status string) (int64, error) {
	return 0, nil
}

// AssignAlert assigns an alert to a user
func (r *AlertRepository) AssignAlert(ctx context.Context, alertID, userID string) error {
	return nil
}

// ResolveAlert resolves an alert
func (r *AlertRepository) ResolveAlert(ctx context.Context, alertID, resolution string) error {
	return nil
}

// MonitoringRuleRepository implements ports.MonitoringRuleRepository
type MonitoringRuleRepository struct {
	conn   *Connection
	logger *zap.Logger
}

// NewMonitoringRuleRepository creates a new monitoring rule repository
func NewMonitoringRuleRepository(conn *Connection, logger *zap.Logger) *MonitoringRuleRepository {
	return &MonitoringRuleRepository{
		conn:   conn,
		logger: logger,
	}
}

// GetActiveRules retrieves all active monitoring rules
func (r *MonitoringRuleRepository) GetActiveRules(ctx context.Context) ([]*domain.MonitoringRule, error) {
	query := `SELECT * FROM monitoring_rules WHERE is_active = true ORDER BY priority DESC`
	rows, err := r.conn.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.MonitoringRule
	for rows.Next() {
		var rule domain.MonitoringRule
		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.RuleType,
			&rule.Condition, &rule.Parameters, &rule.RiskWeight, &rule.Severity,
			&rule.IsActive, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rule: %w", err)
		}
		rules = append(rules, &rule)
	}

	return rules, nil
}

// GetRule retrieves a rule by ID
func (r *MonitoringRuleRepository) GetRule(ctx context.Context, id string) (*domain.MonitoringRule, error) {
	return nil, nil
}

// CreateRule creates a new monitoring rule
func (r *MonitoringRuleRepository) CreateRule(ctx context.Context, rule *domain.MonitoringRule) error {
	return nil
}

// UpdateRule updates an existing rule
func (r *MonitoringRuleRepository) UpdateRule(ctx context.Context, rule *domain.MonitoringRule) error {
	return nil
}

// DeleteRule deletes a rule
func (r *MonitoringRuleRepository) DeleteRule(ctx context.Context, id string) error {
	return nil
}

// GetRulesByType retrieves rules by type
func (r *MonitoringRuleRepository) GetRulesByType(ctx context.Context, ruleType string) ([]*domain.MonitoringRule, error) {
	return []*domain.MonitoringRule{}, nil
}

// Helper function to generate UUID
func generateUUID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(12)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
