package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/platform/blockchain/nodes/internal/config"
	"github.com/csic/platform/blockchain/nodes/internal/domain"
	"github.com/redis/go-redis/v9"
)

// MetricsRepository handles database operations for node metrics
type MetricsRepository struct {
	db        *sql.DB
	redis     *redis.Client
	keyPrefix string
}

// NewMetricsRepository creates a new MetricsRepository instance
func NewMetricsRepository(db *sql.DB, redis *redis.Client, keyPrefix string) *MetricsRepository {
	return &MetricsRepository{
		db:        db,
		redis:     redis,
		keyPrefix: keyPrefix,
	}
}

// SaveMetrics saves node metrics to both database and cache
func (r *MetricsRepository) SaveMetrics(ctx context.Context, metrics *domain.NodeMetrics) error {
	// Save to database
	query := `
		INSERT INTO node_metrics (
			id, node_id, timestamp, block_height, block_time, peer_count,
			transaction_pool_size, gas_price, memory_usage, cpu_usage,
			disk_usage, network_in, network_out, rpc_latency, sync_progress
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	metrics.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	_, err := r.db.Exec(query,
		metrics.ID, metrics.NodeID, metrics.Timestamp, metrics.BlockHeight, metrics.BlockTime,
		metrics.PeerCount, metrics.TransactionPoolSize, metrics.GasPrice, metrics.MemoryUsage,
		metrics.CPUUsage, metrics.DiskUsage, metrics.NetworkIn, metrics.NetworkOut,
		metrics.RPCLatency, metrics.SyncProgress,
	)
	if err != nil {
		return fmt.Errorf("failed to save metrics to database: %w", err)
	}

	// Cache the latest metrics in Redis
	cacheKey := fmt.Sprintf("%smetrics:latest:%s", r.keyPrefix, metrics.NodeID)
	data, _ := json.Marshal(metrics)
	r.redis.Set(ctx, cacheKey, data, time.Hour)

	return nil
}

// GetLatestMetrics retrieves the latest metrics for a node
func (r *MetricsRepository) GetLatestMetrics(ctx context.Context, nodeID string) (*domain.NodeMetrics, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("%smetrics:latest:%s", r.keyPrefix, nodeID)
	data, err := r.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var metrics domain.NodeMetrics
		if json.Unmarshal(data, &metrics) == nil {
			return &metrics, nil
		}
	}

	// Fall back to database
	query := `
		SELECT id, node_id, timestamp, block_height, block_time, peer_count,
			transaction_pool_size, gas_price, memory_usage, cpu_usage,
			disk_usage, network_in, network_out, rpc_latency, sync_progress
		FROM node_metrics WHERE node_id = $1
		ORDER BY timestamp DESC LIMIT 1
	`

	var metrics domain.NodeMetrics
	err = r.db.QueryRow(query, nodeID).Scan(
		&metrics.ID, &metrics.NodeID, &metrics.Timestamp, &metrics.BlockHeight,
		&metrics.BlockTime, &metrics.PeerCount, &metrics.TransactionPoolSize, &metrics.GasPrice,
		&metrics.MemoryUsage, &metrics.CPUUsage, &metrics.DiskUsage, &metrics.NetworkIn,
		&metrics.NetworkOut, &metrics.RPCLatency, &metrics.SyncProgress,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &metrics, nil
}

// GetMetricsHistory retrieves historical metrics for a node
func (r *MetricsRepository) GetMetricsHistory(ctx context.Context, nodeID string, startTime, endTime time.Time, limit int) ([]*domain.NodeMetrics, error) {
	query := `
		SELECT id, node_id, timestamp, block_height, block_time, peer_count,
			transaction_pool_size, gas_price, memory_usage, cpu_usage,
			disk_usage, network_in, network_out, rpc_latency, sync_progress
		FROM node_metrics
		WHERE node_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp DESC
		LIMIT $4
	`

	rows, err := r.db.Query(query, nodeID, startTime, endTime, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make([]*domain.NodeMetrics, 0)
	for rows.Next() {
		var m domain.NodeMetrics
		err := rows.Scan(
			&m.ID, &m.NodeID, &m.Timestamp, &m.BlockHeight, &m.BlockTime,
			&m.PeerCount, &m.TransactionPoolSize, &m.GasPrice, &m.MemoryUsage,
			&m.CPUUsage, &m.DiskUsage, &m.NetworkIn, &m.NetworkOut,
			&m.RPCLatency, &m.SyncProgress,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, &m)
	}

	return metrics, nil
}

// GetNetworkMetrics aggregates metrics for all nodes in a network
func (r *MetricsRepository) GetNetworkMetrics(ctx context.Context, network string, since time.Time) (*domain.NetworkMetrics, error) {
	query := `
		SELECT
			COUNT(DISTINCT nm.node_id) as node_count,
			AVG(nm.block_height) as avg_block_height,
			AVG(nm.block_time) as avg_block_time,
			AVG(nm.peer_count) as avg_peer_count,
			AVG(nm.cpu_usage) as avg_cpu,
			AVG(nm.memory_usage) as avg_memory,
			AVG(nm.gas_price) as avg_gas_price
		FROM node_metrics nm
		JOIN nodes n ON n.id = nm.node_id
		WHERE n.network = $1 AND nm.timestamp >= $2
	`

	var result domain.NetworkMetrics
	result.Network = network
	result.Since = since

	err := r.db.QueryRow(query, network, since).Scan(
		&result.NodeCount, &result.AvgBlockHeight, &result.AvgBlockTime,
		&result.AvgPeerCount, &result.AvgCPU, &result.AvgMemory, &result.AvgGasPrice,
	)

	if err == sql.ErrNoRows {
		return &result, nil
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSystemSummary retrieves aggregated metrics for all nodes
func (r *MetricsRepository) GetSystemSummary(ctx context.Context) (*domain.SystemSummary, error) {
	query := `
		SELECT
			COUNT(*) as total_nodes,
			COUNT(CASE WHEN status = 'online' THEN 1 END) as online_nodes,
			COUNT(CASE WHEN status = 'offline' THEN 1 END) as offline_nodes,
			COUNT(CASE WHEN status = 'syncing' THEN 1 END) as syncing_nodes
		FROM nodes
	`

	var summary domain.SystemSummary
	err := r.db.QueryRow(query).Scan(
		&summary.TotalNodes, &summary.OnlineNodes, &summary.OfflineNodes, &summary.SyncingNodes,
	)

	if err != nil {
		return nil, err
	}

	// Get network-level stats
	networkStats, err := r.GetNetworkStats()
	if err != nil {
		return nil, err
	}

	summary.Networks = make([]*domain.NetworkSummary, 0)
	for _, stat := range networkStats {
		summary.Networks = append(summary.Networks, stat)
	}

	return &summary, nil
}

// NetworkMetrics represents aggregated metrics for a network
type NetworkMetrics struct {
	Network       string  `json:"network"`
	NodeCount     int     `json:"node_count"`
	AvgBlockHeight float64 `json:"avg_block_height"`
	AvgBlockTime   float64 `json:"avg_block_time"`
	AvgPeerCount   float64 `json:"avg_peer_count"`
	AvgCPU         float64 `json:"avg_cpu"`
	AvgMemory      float64 `json:"avg_memory"`
	AvgGasPrice    float64 `json:"avg_gas_price"`
	Since          time.Time `json:"since"`
}
