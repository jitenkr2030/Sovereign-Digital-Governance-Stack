package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/csic/transaction-monitoring/internal/config"
	"github.com/csic/transaction-monitoring/internal/domain/models"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// Repository handles all database operations
type Repository struct {
	db    *sql.DB
	cfg   *config.Config
	logger *zap.Logger
}

// NewRepository creates a new repository instance
func NewRepository(cfg *config.Config, logger *zap.Logger) (*Repository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.GetConnMaxLifetime())

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Repository{
		db:    db,
		cfg:   cfg,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (r *Repository) Close() error {
	return r.db.Close()
}

// Transaction represents a database transaction
type Transaction struct {
	*sql.Tx
}

// BeginTx starts a new database transaction
func (r *Repository) BeginTx(ctx context.Context) (*Transaction, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Transaction{Tx: tx}, nil
}

// SaveTransaction saves a normalized transaction to the database
func (r *Repository) SaveTransaction(ctx context.Context, tx *models.NormalizedTransaction) error {
	query := `
		INSERT INTO transactions (
			tx_hash, network, block_number, block_hash, timestamp,
			total_value, asset, fee, input_count, output_count, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (tx_hash) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query,
		tx.TxHash,
		tx.Network,
		tx.BlockNumber,
		tx.BlockHash,
		tx.Timestamp,
		tx.TotalValue,
		tx.Asset,
		tx.Fee,
		tx.InputCount,
		tx.OutputCount,
		time.Now(),
	)

	return err
}

// GetTransaction retrieves a transaction by hash
func (r *Repository) GetTransaction(ctx context.Context, txHash string) (*models.Transaction, error) {
	query := `
		SELECT tx_hash, network, block_number, block_hash, timestamp,
			   sender, receiver, amount, asset, status, created_at
		FROM transactions WHERE tx_hash = $1
	`

	var tx models.Transaction
	err := r.db.QueryRowContext(ctx, query, txHash).Scan(
		&tx.TxHash, &tx.Network, &tx.BlockNumber, &tx.BlockHash, &tx.Timestamp,
		&tx.Sender, &tx.Receiver, &tx.Amount, &tx.Asset, &tx.Status, &tx.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

// GetMultiInputTransactions returns transactions with multiple inputs
func (r *Repository) GetMultiInputTransactions(ctx context.Context, limit int) ([]models.Transaction, error) {
	query := `
		SELECT tx_hash, network, block_number, block_hash, timestamp,
			   sender, receiver, amount, asset, status, created_at
		FROM transactions
		WHERE input_count >= 2
		ORDER BY timestamp DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(
			&tx.TxHash, &tx.Network, &tx.BlockNumber, &tx.BlockHash, &tx.Timestamp,
			&tx.Sender, &tx.Receiver, &tx.Amount, &tx.Asset, &tx.Status, &tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}

	return txs, rows.Err()
}

// GetTransactionHistory returns transaction history for an address
func (r *Repository) GetTransactionHistory(ctx context.Context, address string, network models.Network, limit int) ([]models.Transaction, error) {
	query := `
		SELECT tx_hash, network, block_number, block_hash, timestamp,
			   sender, receiver, amount, asset, status, created_at
		FROM transactions
		WHERE (sender = $1 OR receiver = $1) AND network = $2
		ORDER BY timestamp DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, address, network, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(
			&tx.TxHash, &tx.Network, &tx.BlockNumber, &tx.BlockHash, &tx.Timestamp,
			&tx.Sender, &tx.Receiver, &tx.Amount, &tx.Asset, &tx.Status, &tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}

	return txs, rows.Err()
}

// Wallet operations

// GetWalletByAddress retrieves a wallet by address and network
func (r *Repository) GetWalletByAddress(ctx context.Context, address string, network models.Network) (*models.Wallet, error) {
	query := `
		SELECT id, address, network, wallet_type, label, first_seen, last_seen,
			   tx_count, total_received, total_sent, current_balance,
			   risk_score, risk_level, is_sanctioned, is_blacklisted, is_whitelisted,
			   cluster_id, created_at, updated_at
		FROM wallets WHERE address = $1 AND network = $2
	`

	var wallet models.Wallet
	var label, clusterID sql.NullString

	err := r.db.QueryRowContext(ctx, query, address, network).Scan(
		&wallet.ID, &wallet.Address, &wallet.Network, &wallet.WalletType, &label,
		&wallet.FirstSeen, &wallet.LastSeen, &wallet.TxCount,
		&wallet.TotalReceived, &wallet.TotalSent, &wallet.CurrentBalance,
		&wallet.RiskScore, &wallet.RiskLevel, &wallet.IsSanctioned,
		&wallet.IsBlacklisted, &wallet.IsWhitelisted, &clusterID,
		&wallet.CreatedAt, &wallet.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if label.Valid {
		wallet.Label = label.String
	}
	if clusterID.Valid {
		wallet.ClusterID = &clusterID.String
	}

	return &wallet, nil
}

// CreateWallet creates a new wallet
func (r *Repository) CreateWallet(ctx context.Context, wallet *models.Wallet) error {
	query := `
		INSERT INTO wallets (
			id, address, network, wallet_type, label, first_seen, last_seen,
			tx_count, total_received, total_sent, current_balance,
			risk_score, risk_level, is_sanctioned, is_blacklisted, is_whitelisted,
			cluster_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	_, err := r.db.ExecContext(ctx, query,
		wallet.ID, wallet.Address, wallet.Network, wallet.WalletType, wallet.Label,
		wallet.FirstSeen, wallet.LastSeen, wallet.TxCount,
		wallet.TotalReceived, wallet.TotalSent, wallet.CurrentBalance,
		wallet.RiskScore, wallet.RiskLevel, wallet.IsSanctioned,
		wallet.IsBlacklisted, wallet.IsWhitelisted, wallet.ClusterID,
		wallet.CreatedAt, wallet.UpdatedAt,
	)

	return err
}

// UpdateWalletRiskScore updates the risk score for a wallet
func (r *Repository) UpdateWalletRiskScore(ctx context.Context, wallet *models.Wallet) error {
	query := `
		UPDATE wallets SET
			risk_score = $1, risk_level = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query,
		wallet.RiskScore, wallet.RiskLevel, time.Now(), wallet.ID,
	)

	return err
}

// UpdateWalletCluster updates the cluster ID for a wallet
func (r *Repository) UpdateWalletCluster(ctx context.Context, address, clusterID string) error {
	query := `UPDATE wallets SET cluster_id = $1, updated_at = $2 WHERE address = $3`

	_, err := r.db.ExecContext(ctx, query, clusterID, time.Now(), address)
	return err
}

// GetWalletMetrics retrieves wallet metrics
func (r *Repository) GetWalletMetrics(ctx context.Context, walletID string) (*models.WalletMetrics, error) {
	// This would aggregate metrics from historical data
	return &models.WalletMetrics{}, nil
}

// Cluster operations

// CreateCluster creates a new entity cluster
func (r *Repository) CreateCluster(ctx context.Context, cluster *models.EntityCluster) error {
	query := `
		INSERT INTO entity_clusters (
			id, cluster_type, primary_address, label, description, wallet_count,
			total_volume, total_tx_count, risk_score, risk_level, confidence_score,
			tags, is_verified, verified_by, suspected_entity, discovery_method,
			first_seen, last_activity, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`

	tags := pq.Array(cluster.Tags)
	var verifiedBy sql.NullString

	if cluster.VerifiedBy != nil {
		verifiedBy.String = *cluster.VerifiedBy
		verifiedBy.Valid = true
	}

	_, err := r.db.ExecContext(ctx, query,
		cluster.ID, cluster.ClusterType, cluster.PrimaryAddress, cluster.Label, cluster.Description,
		cluster.WalletCount, cluster.TotalVolume, cluster.TotalTxCount,
		cluster.RiskScore, cluster.RiskLevel, cluster.ConfidenceScore,
		tags, cluster.IsVerified, verifiedBy, cluster.SuspectedEntity, cluster.DiscoveryMethod,
		cluster.FirstSeen, cluster.LastActivity, cluster.CreatedAt, cluster.UpdatedAt,
	)

	return err
}

// GetClusterByID retrieves a cluster by ID
func (r *Repository) GetClusterByID(ctx context.Context, clusterID string) (*models.EntityCluster, error) {
	query := `
		SELECT id, cluster_type, primary_address, label, description, wallet_count,
			   total_volume, total_tx_count, risk_score, risk_level, confidence_score,
			   tags, is_verified, verified_by, suspected_entity, discovery_method,
			   first_seen, last_activity, created_at, updated_at
		FROM entity_clusters WHERE id = $1
	`

	var cluster models.EntityCluster
	var verifiedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, clusterID).Scan(
		&cluster.ID, &cluster.ClusterType, &cluster.PrimaryAddress, &cluster.Label, &cluster.Description,
		&cluster.WalletCount, &cluster.TotalVolume, &cluster.TotalTxCount,
		&cluster.RiskScore, &cluster.RiskLevel, &cluster.ConfidenceScore,
		&cluster.Tags, &cluster.IsVerified, &verifiedBy, &cluster.SuspectedEntity, &cluster.DiscoveryMethod,
		&cluster.FirstSeen, &cluster.LastActivity, &cluster.CreatedAt, &cluster.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if verifiedBy.Valid {
		cluster.VerifiedBy = &verifiedBy.String
	}

	return &cluster, nil
}

// GetAllClusters retrieves all clusters
func (r *Repository) GetAllClusters(ctx context.Context) ([]models.EntityCluster, error) {
	query := `
		SELECT id, cluster_type, primary_address, label, description, wallet_count,
			   total_volume, total_tx_count, risk_score, risk_level, confidence_score,
			   tags, is_verified, verified_by, suspected_entity, discovery_method,
			   first_seen, last_activity, created_at, updated_at
		FROM entity_clusters ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []models.EntityCluster
	for rows.Next() {
		var cluster models.EntityCluster
		var verifiedBy sql.NullString

		if err := rows.Scan(
			&cluster.ID, &cluster.ClusterType, &cluster.PrimaryAddress, &cluster.Label, &cluster.Description,
			&cluster.WalletCount, &cluster.TotalVolume, &cluster.TotalTxCount,
			&cluster.RiskScore, &cluster.RiskLevel, &cluster.ConfidenceScore,
			&cluster.Tags, &cluster.IsVerified, &verifiedBy, &cluster.SuspectedEntity, &cluster.DiscoveryMethod,
			&cluster.FirstSeen, &cluster.LastActivity, &cluster.CreatedAt, &cluster.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if verifiedBy.Valid {
			cluster.VerifiedBy = &verifiedBy.String
		}

		clusters = append(clusters, cluster)
	}

	return clusters, rows.Err()
}

// UpdateCluster updates an existing cluster
func (r *Repository) UpdateCluster(ctx context.Context, cluster *models.EntityCluster) error {
	query := `
		UPDATE entity_clusters SET
			wallet_count = $1, total_volume = $2, total_tx_count = $3,
			risk_score = $4, risk_level = $5, last_activity = $6, updated_at = $7
		WHERE id = $8
	`

	_, err := r.db.ExecContext(ctx, query,
		cluster.WalletCount, cluster.TotalVolume, cluster.TotalTxCount,
		cluster.RiskScore, cluster.RiskLevel, cluster.LastActivity, time.Now(), cluster.ID,
	)

	return err
}

// Cluster member operations

// AddClusterMember adds a member to a cluster
func (r *Repository) AddClusterMember(ctx context.Context, member *models.ClusterMember) error {
	query := `
		INSERT INTO cluster_members (
			id, cluster_id, wallet_address, network, member_type, link_type,
			link_strength, tx_count, volume, first_linked_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		member.ID, member.ClusterID, member.WalletAddress, member.Network,
		member.MemberType, member.LinkType, member.LinkStrength,
		member.TxCount, member.Volume, member.FirstLinkedAt, member.CreatedAt,
	)

	return err
}

// GetClusterMembers retrieves all members of a cluster
func (r *Repository) GetClusterMembers(ctx context.Context, clusterID string) ([]models.ClusterMember, error) {
	query := `
		SELECT id, cluster_id, wallet_address, network, member_type, link_type,
			   link_strength, tx_count, volume, first_linked_at, created_at
		FROM cluster_members WHERE cluster_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.ClusterMember
	for rows.Next() {
		var member models.ClusterMember
		if err := rows.Scan(
			&member.ID, &member.ClusterID, &member.WalletAddress, &member.Network,
			&member.MemberType, &member.LinkType, &member.LinkStrength,
			&member.TxCount, &member.Volume, &member.FirstLinkedAt, &member.CreatedAt,
		); err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, rows.Err()
}

// UpdateClusterMember updates a cluster member
func (r *Repository) UpdateClusterMember(ctx context.Context, member *models.ClusterMember) error {
	query := `UPDATE cluster_members SET cluster_id = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, member.ClusterID, member.ID)
	return err
}

// GetClusterRelationships retrieves cluster relationships
func (r *Repository) GetClusterRelationships(ctx context.Context, clusterID string) ([]models.ClusterRelationship, error) {
	query := `
		SELECT id, source_cluster_id, target_cluster_id, relationship_type,
			   tx_count, total_volume, first_seen, last_seen, created_at
		FROM cluster_relationships
		WHERE source_cluster_id = $1 OR target_cluster_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relationships []models.ClusterRelationship
	for rows.Next() {
		var rel models.ClusterRelationship
		if err := rows.Scan(
			&rel.ID, &rel.SourceClusterID, &rel.TargetClusterID, &rel.RelationshipType,
			&rel.TxCount, &rel.TotalVolume, &rel.FirstSeen, &rel.LastSeen, &rel.CreatedAt,
		); err != nil {
			return nil, err
		}
		relationships = append(relationships, rel)
	}

	return relationships, rows.Err()
}

// Alert operations

// CreateAlert creates a new alert
func (r *Repository) CreateAlert(ctx context.Context, alert *models.Alert) error {
	query := `
		INSERT INTO alerts (
			id, alert_type, severity, status, target_type, target_id, target_value,
			title, description, rule_id, rule_name, triggered_at, resolved_at,
			assigned_to, assigned_team, case_id, score, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	var ruleID, resolvedAt, assignedTo, assignedTeam, caseID sql.NullString

	if alert.RuleID != nil {
		ruleID.String = *alert.RuleID
		ruleID.Valid = true
	}
	if alert.ResolvedAt != nil {
		resolvedAt.Time = *alert.ResolvedAt
		resolvedAt.Valid = true
	}
	if alert.AssignedTo != nil {
		assignedTo.String = *alert.AssignedTo
		assignedTo.Valid = true
	}
	if alert.AssignedTeam != "" {
		assignedTeam.String = alert.AssignedTeam
		assignedTeam.Valid = true
	}
	if alert.CaseID != nil {
		caseID.String = *alert.CaseID
		caseID.Valid = true
	}

	_, err := r.db.ExecContext(ctx, query,
		alert.ID, alert.AlertType, alert.Severity, alert.Status,
		alert.TargetType, alert.TargetID, alert.TargetValue,
		alert.Title, alert.Description, ruleID, alert.RuleName, alert.TriggeredAt,
		resolvedAt, assignedTo, assignedTeam, caseID, alert.Score,
		alert.CreatedAt, alert.UpdatedAt,
	)

	return err
}

// GetAlertsByType retrieves alerts by type
func (r *Repository) GetAlertsByType(ctx context.Context, alertType models.AlertType) ([]models.Alert, error) {
	query := `
		SELECT id, alert_type, severity, status, target_type, target_id, target_value,
			   title, description, rule_id, rule_name, triggered_at, resolved_at,
			   assigned_to, assigned_team, case_id, score, created_at, updated_at
		FROM alerts WHERE alert_type = $1 AND status NOT IN ('resolved', 'dismissed', 'false_positive')
	`

	rows, err := r.db.QueryContext(ctx, query, alertType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var alert models.Alert
		var ruleID, resolvedAt, assignedTo, caseID sql.NullString
		var assignedTeam sql.NullString

		if err := rows.Scan(
			&alert.ID, &alert.AlertType, &alert.Severity, &alert.Status,
			&alert.TargetType, &alert.TargetID, &alert.TargetValue,
			&alert.Title, &alert.Description, &ruleID, &alert.RuleName, &alert.TriggeredAt,
			&resolvedAt, &assignedTo, &assignedTeam, &caseID, &alert.Score,
			&alert.CreatedAt, &alert.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if ruleID.Valid {
			alert.RuleID = &ruleID.String
		}
		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}
		if assignedTo.Valid {
			alert.AssignedTo = &assignedTo.String
		}
		if caseID.Valid {
			alert.CaseID = &caseID.String
		}

		alerts = append(alerts, alert)
	}

	return alerts, rows.Err()
}

// Utility queries

// FindHighVolumeAddresses finds addresses with high transaction volume
func (r *Repository) FindHighVolumeAddresses(ctx context.Context, minTxCount int) ([]models.Wallet, error) {
	query := `
		SELECT id, address, network, wallet_type, label, first_seen, last_seen,
			   tx_count, total_received, total_sent, current_balance,
			   risk_score, risk_level, is_sanctioned, is_blacklisted, is_whitelisted,
			   cluster_id, created_at, updated_at
		FROM wallets WHERE tx_count >= $1 ORDER BY total_received DESC LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, minTxCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []models.Wallet
	for rows.Next() {
		var wallet models.Wallet
		if err := rows.Scan(
			&wallet.ID, &wallet.Address, &wallet.Network, &wallet.WalletType, &wallet.Label,
			&wallet.FirstSeen, &wallet.LastSeen, &wallet.TxCount,
			&wallet.TotalReceived, &wallet.TotalSent, &wallet.CurrentBalance,
			&wallet.RiskScore, &wallet.RiskLevel, &wallet.IsSanctioned,
			&wallet.IsBlacklisted, &wallet.IsWhitelisted, &wallet.ClusterID,
			&wallet.CreatedAt, &wallet.UpdatedAt,
		); err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}

	return wallets, rows.Err()
}

// GetSendersToAddress returns all sender addresses that sent to a given address
func (r *Repository) GetSendersToAddress(ctx context.Context, address string, network models.Network) ([]string, error) {
	query := `
		SELECT DISTINCT sender FROM transactions
		WHERE receiver = $1 AND network = $2
	`

	rows, err := r.db.QueryContext(ctx, query, address, network)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var senders []string
	for rows.Next() {
		var sender string
		if err := rows.Scan(&sender); err != nil {
			return nil, err
		}
		senders = append(senders, sender)
	}

	return senders, rows.Err()
}

// GetTransactionsWithChangeOutputs returns transactions with change outputs
func (r *Repository) GetTransactionsWithChangeOutputs(ctx context.Context, limit int) ([]models.Transaction, error) {
	query := `
		SELECT tx_hash, network, block_number, block_hash, timestamp,
			   sender, receiver, amount, asset, status, created_at
		FROM transactions
		WHERE output_count > 1
		ORDER BY timestamp DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(
			&tx.TxHash, &tx.Network, &tx.BlockNumber, &tx.BlockHash, &tx.Timestamp,
			&tx.Sender, &tx.Receiver, &tx.Amount, &tx.Asset, &tx.Status, &tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}

	return txs, rows.Err()
}
