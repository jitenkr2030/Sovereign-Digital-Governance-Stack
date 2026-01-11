package graph

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/csic/transaction-monitoring/internal/config"
	"github.com/csic/transaction-monitoring/internal/domain/models"
	"github.com/csic/transaction-monitoring/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ClusteringService handles entity clustering operations
type ClusteringService struct {
	cfg          *config.Config
	repo         *repository.Repository
	neo4jRepo    *repository.Neo4jRepository
	logger       *zap.Logger
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.RWMutex
	isRunning    bool
}

// NewClusteringService creates a new clustering service
func NewClusteringService(
	cfg *config.Config,
	repo *repository.Repository,
	neo4jRepo *repository.Neo4jRepository,
	logger *zap.Logger,
) *ClusteringService {
	return &ClusteringService{
		cfg:       cfg,
		repo:      repo,
		neo4jRepo: neo4jRepo,
		logger:    logger,
		stopChan:  make(chan struct{}),
	}
}

// Start begins the clustering process
func (s *ClusteringService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return nil
	}
	s.isRunning = true
	s.mu.Unlock()

	s.logger.Info("Starting entity clustering service")

	if s.cfg.Clustering.Enabled {
		s.wg.Add(1)
		go s.clusteringLoop(ctx)
	}

	return nil
}

// Stop gracefully stops the clustering service
func (s *ClusteringService) Stop() {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = false
	s.mu.Unlock()

	s.logger.Info("Stopping entity clustering service")
	close(s.stopChan)
	s.wg.Wait()
}

// clusteringLoop runs periodic clustering
func (s *ClusteringService) clusteringLoop(ctx context.Context) {
	defer s.wg.Done()

	batchInterval, _ := time.ParseDuration(s.cfg.Clustering.BatchInterval)
	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runClustering(ctx)
		}
	}
}

// runClustering executes clustering algorithms
func (s *ClusteringService) runClustering(ctx context.Context) {
	startTime := time.Now()

	// Process common input clustering (Bitcoin)
	if s.cfg.Clustering.CommonInputEnabled {
		if err := s.processCommonInputClustering(ctx); err != nil {
			s.logger.Error("Common input clustering failed", zap.Error(err))
		}
	}

	// Process deposit address clustering
	if s.cfg.Clustering.DepositClusterEnabled {
		if err := s.processDepositClustering(ctx); err != nil {
			s.logger.Error("Deposit clustering failed", zap.Error(err))
		}
	}

	// Process change address linking
	if s.cfg.Clustering.ChangeLinkEnabled {
		if err := s.processChangeLinkClustering(ctx); err != nil {
			s.logger.Error("Change link clustering failed", zap.Error(err))
		}
	}

	s.logger.Info("Clustering cycle completed",
		zap.Duration("duration", time.Since(startTime)))
}

// processCommonInputClustering identifies wallets that appear together in transaction inputs
func (s *ClusteringService) processCommonInputClustering(ctx context.Context) error {
	s.logger.Info("Processing common input clustering")

	// Get recent transactions with multiple inputs
	txs, err := s.repo.GetMultiInputTransactions(ctx, 1000)
	if err != nil {
		return fmt.Errorf("failed to get multi-input transactions: %w", err)
	}

	clusteredCount := 0
	for _, tx := range txs {
		if len(tx.Inputs) >= 2 {
			// All inputs in a Bitcoin transaction are likely owned by the same entity
			cluster, err := s.findOrCreateCluster(ctx, tx.Inputs, "common-input")
			if err != nil {
				s.logger.Warn("Failed to create cluster for multi-input tx",
					zap.String("tx_hash", tx.TxHash),
					zap.Error(err))
				continue
			}

			// Add members to cluster
			for _, input := range tx.Inputs {
				if err := s.addToCluster(ctx, cluster.ID, input.Address, "common-input", 1.0); err != nil {
					s.logger.Warn("Failed to add address to cluster",
						zap.String("address", input.Address),
						zap.Error(err))
					continue
				}
			}

			clusteredCount++
		}
	}

	s.logger.Info("Common input clustering completed",
		zap.Int("clusters_created", clusteredCount))

	return nil
}

// processDepositClustering identifies wallets that deposit to common addresses
func (s *ClusteringService) processDepositClustering(ctx context.Context) error {
	s.logger.Info("Processing deposit address clustering")

	// Find centralized exchange deposit addresses (heuristics: high volume, many senders)
	depositAddresses, err := s.repo.FindHighVolumeAddresses(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to find deposit addresses: %w", err)
	}

	for _, addr := range depositAddresses {
		// Find all senders to this address
		senders, err := s.repo.GetSendersToAddress(ctx, addr.Address, addr.Network)
		if err != nil {
			continue
		}

		if len(senders) >= 5 {
			// These senders likely belong to the same entity
			cluster, err := s.findOrCreateCluster(ctx, senders, "deposit")
			if err != nil {
				continue
			}

			for _, sender := range senders {
				s.addToCluster(ctx, cluster.ID, sender, "deposit", 0.8)
			}
		}
	}

	return nil
}

// processChangeLinkClustering links change addresses to original sender
func (s *ClusteringService) processChangeLinkClustering(ctx context.Context) error {
	s.logger.Info("Processing change address linking")

	// Get transactions with change outputs
	txs, err := s.repo.GetTransactionsWithChangeOutputs(ctx, 10000)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	for _, tx := range txs {
		for _, output := range tx.Outputs {
			if output.IsChange {
				// Link change address to sender
				s.addToCluster(ctx, tx.Sender, output.Address, "change", 0.7)
			}
		}
	}

	return nil
}

// findOrCreateCluster finds or creates a cluster for addresses
func (s *ClusteringService) findOrCreateCluster(ctx context.Context, addresses []string, clusterType string) (*models.EntityCluster, error) {
	// Check if any address already belongs to a cluster
	for _, addr := range addresses {
		wallet, err := s.repo.GetWalletByAddress(ctx, addr, models.NetworkBitcoin)
		if err != nil {
			continue
		}
		if wallet.ClusterID != nil {
			return s.repo.GetClusterByID(ctx, *wallet.ClusterID)
		}
	}

	// Create new cluster
	cluster := &models.EntityCluster{
		ID:              uuid.New().String(),
		ClusterType:     clusterType,
		PrimaryAddress:  addresses[0],
		WalletCount:     len(addresses),
		ConfidenceScore: 1.0,
		Tags:            []string{clusterType},
		DiscoveryMethod: clusterType,
		FirstSeen:       time.Now(),
		LastActivity:    time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.CreateCluster(ctx, cluster); err != nil {
		return nil, err
	}

	// Store in Neo4j for graph operations
	if s.neo4jRepo != nil {
		if err := s.neo4jRepo.CreateCluster(ctx, cluster); err != nil {
			s.logger.Warn("Failed to create cluster in Neo4j", zap.Error(err))
		}
	}

	return cluster, nil
}

// addToCluster adds an address to a cluster
func (s *ClusteringService) addToCluster(ctx context.Context, clusterID, address, linkType string, linkStrength float64) error {
	member := &models.ClusterMember{
		ID:            uuid.New().String(),
		ClusterID:     clusterID,
		WalletAddress: address,
		MemberType:    "linked",
		LinkType:      linkType,
		LinkStrength:  linkStrength,
		FirstLinkedAt: time.Now(),
		CreatedAt:     time.Now(),
	}

	if err := s.repo.AddClusterMember(ctx, member); err != nil {
		return err
	}

	// Update wallet with cluster ID
	if err := s.repo.UpdateWalletCluster(ctx, address, clusterID); err != nil {
		return err
	}

	// Add to Neo4j graph
	if s.neo4jRepo != nil {
		s.neo4jRepo.AddToCluster(ctx, clusterID, address)
	}

	return nil
}

// GetCluster returns a cluster by ID
func (s *ClusteringService) GetCluster(ctx context.Context, clusterID string) (*models.EntityCluster, error) {
	return s.repo.GetClusterByID(ctx, clusterID)
}

// GetClusterMembers returns all members of a cluster
func (s *ClusteringService) GetClusterMembers(ctx context.Context, clusterID string) ([]models.ClusterMember, error) {
	return s.repo.GetClusterMembers(ctx, clusterID)
}

// GetWalletCluster returns the cluster for a wallet
func (s *ClusteringService) GetWalletCluster(ctx context.Context, address string, network models.Network) (*models.EntityCluster, error) {
	wallet, err := s.repo.GetWalletByAddress(ctx, address, network)
	if err != nil {
		return nil, err
	}

	if wallet.ClusterID == nil {
		return nil, nil
	}

	return s.repo.GetClusterByID(ctx, *wallet.ClusterID)
}

// FindRelatedClusters finds clusters related to a given cluster
func (s *ClusteringService) FindRelatedClusters(ctx context.Context, clusterID string) ([]models.ClusterRelationship, error) {
	return s.repo.GetClusterRelationships(ctx, clusterID)
}

// AnalyzeClusterRisk analyzes and updates risk for all wallets in a cluster
func (s *ClusteringService) AnalyzeClusterRisk(ctx context.Context, clusterID string) error {
	cluster, err := s.repo.GetClusterByID(ctx, clusterID)
	if err != nil {
		return err
	}

	members, err := s.repo.GetClusterMembers(ctx, clusterID)
	if err != nil {
		return err
	}

	// Calculate cluster metrics
	var totalScore float64
	maxScore := 0.0
	highRiskCount := 0

	for _, member := range members {
		wallet, err := s.repo.GetWalletByAddress(ctx, member.WalletAddress, member.Network)
		if err != nil {
			continue
		}

		totalScore += wallet.RiskScore
		if wallet.RiskScore > maxScore {
			maxScore = wallet.RiskScore
		}
		if wallet.RiskScore >= 60 {
			highRiskCount++
		}
	}

	if len(members) > 0 {
		cluster.TotalVolume = fmt.Sprintf("%f", totalScore)
		cluster.TotalTxCount = int64(len(members))
		cluster.RiskScore = totalScore / float64(len(members))

		// Cluster risk increases if any member has high risk
		if maxScore > cluster.RiskScore {
			cluster.RiskScore = maxScore
		}

		// Update risk level
		switch {
		case cluster.RiskScore >= 80:
			cluster.RiskLevel = "critical"
		case cluster.RiskScore >= 60:
			cluster.RiskLevel = "high"
		case cluster.RiskScore >= 40:
			cluster.RiskLevel = "medium"
		default:
			cluster.RiskLevel = "low"
		}
	}

	cluster.LastActivity = time.Now()

	return s.repo.UpdateCluster(ctx, cluster)
}

// MergeClusters merges two clusters into one
func (s *ClusteringService) MergeClusters(ctx context.Context, sourceID, targetID string) error {
	sourceCluster, err := s.repo.GetClusterByID(ctx, sourceID)
	if err != nil {
		return err
	}

	targetCluster, err := s.repo.GetClusterByID(ctx, targetID)
	if err != nil {
		return err
	}

	// Move all members from source to target
	members, err := s.repo.GetClusterMembers(ctx, sourceID)
	if err != nil {
		return err
	}

	for _, member := range members {
		member.ClusterID = targetID
		if err := s.repo.UpdateClusterMember(ctx, &member); err != nil {
			return err
		}
	}

	// Update cluster stats
	targetCluster.WalletCount += sourceCluster.WalletCount
	targetCluster.TotalTxCount += sourceCluster.TotalTxCount
	targetCluster.RiskScore = (targetCluster.RiskScore + sourceCluster.RiskScore) / 2

	// Archive source cluster
	sourceCluster.IsVerified = true
	if err := s.repo.UpdateCluster(ctx, sourceCluster); err != nil {
		return err
	}

	return s.repo.UpdateCluster(ctx, targetCluster)
}

// GetClusteringStats returns clustering statistics
func (s *ClusteringService) GetClusteringStats(ctx context.Context) (*ClusteringStats, error) {
	stats := &ClusteringStats{}

	clusters, err := s.repo.GetAllClusters(ctx)
	if err != nil {
		return nil, err
	}

	stats.TotalClusters = len(clusters)
	for _, c := range clusters {
		switch c.RiskLevel {
		case "critical":
			stats.CriticalClusters++
		case "high":
			stats.HighRiskClusters++
		case "medium":
			stats.MediumRiskClusters++
		}
	}

	return stats, nil
}

// ClusteringStats contains clustering operation statistics
type ClusteringStats struct {
	TotalClusters     int `json:"total_clusters"`
	CriticalClusters  int `json:"critical_clusters"`
	HighRiskClusters  int `json:"high_risk_clusters"`
	MediumRiskClusters int `json:"medium_risk_clusters"`
}

// HealthCheck verifies the service is healthy
func (s *ClusteringService) HealthCheck(ctx context.Context) error {
	return nil
}
