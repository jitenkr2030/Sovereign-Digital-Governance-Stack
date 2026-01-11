package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/csic/platform/blockchain/nodes/internal/config"
	"github.com/csic/platform/blockchain/nodes/internal/domain"
	"github.com/csic/platform/blockchain/nodes/internal/repository"
)

// HealthService handles node health monitoring
type HealthService struct {
	config   *config.Config
	nodeRepo *repository.NodeRepository
	producer interface {
		Publish(ctx context.Context, topic string, event interface{})
	}
	mu      sync.RWMutex
}

// NewHealthService creates a new HealthService instance
func NewHealthService(cfg *config.Config, nodeRepo *repository.NodeRepository, producer interface{}) *HealthService {
	return &HealthService{
		config:   cfg,
		nodeRepo: nodeRepo,
		producer: producer,
	}
}

// StartHealthCheckLoop starts the background health check loop
func (s *HealthService) StartHealthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(s.config.GetHealthCheckInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runHealthChecks(ctx)
		}
	}
}

// runHealthChecks performs health checks on all active nodes
func (s *HealthService) runHealthChecks(ctx context.Context) {
	filter := &domain.NodeListFilter{Limit: 1000}
	paginatedNodes, err := s.nodeRepo.List(filter)
	if err != nil {
		fmt.Printf("Failed to list nodes: %v\n", err)
		return
	}

	for _, node := range paginatedNodes.Nodes {
		result := s.checkNodeHealth(ctx, node)

		// Update node status in database
		if err := s.nodeRepo.UpdateStatus(node.ID, result.Status, result.Details); err != nil {
			fmt.Printf("Failed to update node status: %v\n", err)
		}

		// Handle status changes
		if node.Status != result.Status {
			s.handleStatusChange(ctx, node, result)
		}
	}
}

// checkNodeHealth performs a health check on a single node
func (s *HealthService) checkNodeHealth(ctx context.Context, node *domain.Node) *domain.HealthCheckResult {
	result := &domain.HealthCheckResult{
		NodeID:    node.ID,
		CheckedAt: time.Now(),
		Details:   make(map[string]interface{}),
	}

	start := time.Now()

	// Perform health check (simplified - real implementation would check RPC endpoint)
	// This is a placeholder for actual health check logic
	result.Latency = time.Since(start)

	// Simulate health check result
	result.Status = domain.NodeStatusOnline
	result.Details["reachable"] = true
	result.Details["version"] = "1.0.0"

	return result
}

// handleStatusChange handles changes in node status
func (s *HealthService) handleStatusChange(ctx context.Context, node *domain.Node, result *domain.HealthCheckResult) {
	var severity string
	var eventType string

	switch result.Status {
	case domain.NodeStatusOffline:
		severity = "error"
		eventType = "node.offline"
	case domain.NodeStatusOnline:
		severity = "info"
		eventType = "node.online"
	case domain.NodeStatusSyncing:
		severity = "warning"
		eventType = "node.syncing"
	default:
		severity = "warning"
		eventType = "node.status_changed"
	}

	event := &domain.NodeEvent{
		NodeID:    node.ID,
		EventType: eventType,
		Severity:  severity,
		Message:   fmt.Sprintf("Node '%s' status changed from %s to %s", node.Name, node.Status, result.Status),
		Timestamp: time.Now(),
	}

	// Update local node status
	s.mu.Lock()
	node.Status = result.Status
	s.mu.Unlock()

	// Publish event
	if s.producer != nil {
		s.producer.Publish(ctx, "csic.nodes.events", event)
	}
}

// GetNodeHealth retrieves the current health status of a node
func (s *HealthService) GetNodeHealth(ctx context.Context, nodeID string) (*domain.HealthCheckResult, error) {
	node, err := s.nodeRepo.GetByID(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	if node == nil {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	return s.checkNodeHealth(ctx, node), nil
}

// GetHealth returns overall system health
func (s *HealthService) GetHealth(ctx context.Context) map[string]interface{} {
	filter := &domain.NodeListFilter{Limit: 1000}
	paginatedNodes, err := s.nodeRepo.List(filter)
	if err != nil {
		return map[string]interface{}{
			"status":  "unhealthy",
			"error":   err.Error(),
			"nodes":   map[string]interface{}{},
		}
	}

	online := 0
	offline := 0
	syncing := 0

	for _, node := range paginatedNodes.Nodes {
		switch node.Status {
		case domain.NodeStatusOnline:
			online++
		case domain.NodeStatusOffline:
			offline++
		case domain.NodeStatusSyncing:
			syncing++
		}
	}

	total := online + offline + syncing
	healthPercent := 0.0
	if total > 0 {
		healthPercent = float64(online) / float64(total) * 100
	}

	status := "healthy"
	if healthPercent < 50 {
		status = "critical"
	} else if healthPercent < 80 {
		status = "degraded"
	}

	return map[string]interface{}{
		"status":        status,
		"total_nodes":   total,
		"online_nodes":  online,
		"offline_nodes": offline,
		"syncing_nodes": syncing,
		"health_score":  healthPercent,
	}
}
