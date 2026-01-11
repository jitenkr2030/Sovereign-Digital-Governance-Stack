package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic/platform/blockchain/nodes/internal/config"
	"github.com/csic/platform/blockchain/nodes/internal/domain"
	"github.com/csic/platform/blockchain/nodes/internal/messaging"
	"github.com/csic/platform/blockchain/nodes/internal/repository"
)

// MetricsService handles node metrics operations
type MetricsService struct {
	config      *config.Config
	metricsRepo *repository.MetricsRepository
	producer    messaging.KafkaProducer
}

// NewMetricsService creates a new MetricsService instance
func NewMetricsService(cfg *config.Config, metricsRepo *repository.MetricsRepository, producer messaging.KafkaProducer) *MetricsService {
	return &MetricsService{
		config:      cfg,
		metricsRepo: metricsRepo,
		producer:    producer,
	}
}

// GetNodeMetrics retrieves the latest metrics for a node
func (s *MetricsService) GetNodeMetrics(ctx context.Context, nodeID string) (*domain.NodeMetrics, error) {
	metrics, err := s.metricsRepo.GetLatestMetrics(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics: %w", err)
	}
	return metrics, nil
}

// GetNodeMetricsHistory retrieves historical metrics for a node
func (s *MetricsService) GetNodeMetricsHistory(ctx context.Context, nodeID string, startTime, endTime time.Time, limit int) ([]*domain.NodeMetrics, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	metrics, err := s.metricsRepo.GetMetricsHistory(ctx, nodeID, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics history: %w", err)
	}
	return metrics, nil
}

// GetNetworkMetrics retrieves aggregated metrics for a network
func (s *MetricsService) GetNetworkMetrics(ctx context.Context, network string, duration time.Duration) (*domain.NetworkMetrics, error) {
	since := time.Now().Add(-duration)
	metrics, err := s.metricsRepo.GetNetworkMetrics(ctx, network, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get network metrics: %w", err)
	}
	return metrics, nil
}

// GetSystemSummary retrieves aggregated metrics for all nodes
func (s *MetricsService) GetSystemSummary(ctx context.Context) (*domain.SystemSummary, error) {
	summary, err := s.metricsRepo.GetSystemSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system summary: %w", err)
	}
	return summary, nil
}

// RecordMetrics records new metrics for a node
func (s *MetricsService) RecordMetrics(ctx context.Context, metrics *domain.NodeMetrics) error {
	if err := s.metricsRepo.SaveMetrics(ctx, metrics); err != nil {
		return fmt.Errorf("failed to record metrics: %w", err)
	}

	// Publish metrics to Kafka
	if s.producer != nil {
		s.producer.Publish(ctx, s.config.Kafka.Topics.NodeMetrics, metrics)
	}

	return nil
}

// PrometheusMetrics returns metrics in Prometheus format
func (s *MetricsService) PrometheusMetrics(ctx context.Context) (string, error) {
	summary, err := s.metricsRepo.GetSystemSummary(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get system summary: %w", err)
	}

	// This is a simplified version - a real implementation would gather all metrics
	metrics := fmt.Sprintf(`# HELP csic_nodes_total Total number of nodes
# TYPE csic_nodes_total gauge
csic_nodes_total %d
# HELP csic_nodes_online Number of online nodes
# TYPE csic_nodes_online gauge
csic_nodes_online %d
# HELP csic_nodes_offline Number of offline nodes
# TYPE csic_nodes_offline gauge
csic_nodes_offline %d
# HELP csic_nodes_syncing Number of syncing nodes
# TYPE csic_nodes_syncing gauge
csic_nodes_syncing %d
`, summary.TotalNodes, summary.OnlineNodes, summary.OfflineNodes, summary.SyncingNodes)

	return metrics, nil
}
