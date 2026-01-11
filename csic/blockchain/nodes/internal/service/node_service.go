package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/csic/platform/blockchain/nodes/internal/config"
	"github.com/csic/platform/blockchain/nodes/internal/domain"
	"github.com/csic/platform/blockchain/nodes/internal/messaging"
	"github.com/csic/platform/blockchain/nodes/internal/repository"
)

// NodeService handles node management operations
type NodeService struct {
	config     *config.Config
	nodeRepo   *repository.NodeRepository
	producer   messaging.KafkaProducer
}

// NewNodeService creates a new NodeService instance
func NewNodeService(cfg *config.Config, nodeRepo *repository.NodeRepository, producer messaging.KafkaProducer) *NodeService {
	return &NodeService{
		config:   cfg,
		nodeRepo: nodeRepo,
		producer: producer,
	}
}

// CreateNode creates a new blockchain node
func (s *NodeService) CreateNode(ctx context.Context, req *domain.CreateNodeRequest) (*domain.Node, error) {
	// Validate network exists
	networkConfig, exists := s.config.Blockchains[req.Network]
	if !exists {
		return nil, fmt.Errorf("unknown network: %s", req.Network)
	}

	// Validate node limit
	currentNodes, err := s.nodeRepo.GetNodesByNetwork(req.Network)
	if err != nil {
		return nil, fmt.Errorf("failed to check node count: %w", err)
	}
	if len(currentNodes) >= s.config.NodeManager.MaxNodesPerNetwork {
		return nil, fmt.Errorf("maximum nodes per network (%d) reached", s.config.NodeManager.MaxNodesPerNetwork)
	}

	nodeType := req.Type
	if nodeType == "" {
		nodeType = domain.NodeTypeFull
	}

	node := &domain.Node{
		Name:    req.Name,
		Network: req.Network,
		Type:    nodeType,
		RPCURL:  req.RPCURL,
		WSURL:   req.WSURL,
		ChainID: networkConfig.ChainID,
		Metadata: req.Metadata,
		Status:   domain.NodeStatusUnknown,
	}

	if err := s.nodeRepo.Create(node); err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	// Publish node created event
	s.publishEvent(ctx, &domain.NodeEvent{
		NodeID:    node.ID,
		EventType: "node.created",
		Severity:  "info",
		Message:   fmt.Sprintf("Node '%s' created on network '%s'", node.Name, node.Network),
	})

	return node, nil
}

// GetNode retrieves a node by ID
func (s *NodeService) GetNode(ctx context.Context, id string) (*domain.Node, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	if node == nil {
		return nil, fmt.Errorf("node not found: %s", id)
	}
	return node, nil
}

// ListNodes lists all nodes with optional filtering
func (s *NodeService) ListNodes(ctx context.Context, filter *domain.NodeListFilter) (*domain.PaginatedNodes, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return s.nodeRepo.List(filter)
}

// UpdateNode updates an existing node
func (s *NodeService) UpdateNode(ctx context.Context, id string, req *domain.UpdateNodeRequest) (*domain.Node, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	if node == nil {
		return nil, fmt.Errorf("node not found: %s", id)
	}

	if req.Name != nil {
		node.Name = *req.Name
	}
	if req.RPCURL != nil {
		node.RPCURL = *req.RPCURL
	}
	if req.WSURL != nil {
		node.WSURL = *req.WSURL
	}
	if req.Type != nil {
		node.Type = *req.Type
	}
	if req.Metadata != nil {
		node.Metadata = req.Metadata
	}

	if err := s.nodeRepo.Update(node); err != nil {
		return nil, fmt.Errorf("failed to update node: %w", err)
	}

	return node, nil
}

// DeleteNode deletes a node
func (s *NodeService) DeleteNode(ctx context.Context, id string) error {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}
	if node == nil {
		return fmt.Errorf("node not found: %s", id)
	}

	if err := s.nodeRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	s.publishEvent(ctx, &domain.NodeEvent{
		NodeID:    id,
		EventType: "node.deleted",
		Severity:  "info",
		Message:   fmt.Sprintf("Node '%s' deleted from network '%s'", node.Name, node.Network),
	})

	return nil
}

// RestartNode restarts a node
func (s *NodeService) RestartNode(ctx context.Context, id string) error {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}
	if node == nil {
		return fmt.Errorf("node not found: %s", id)
	}

	s.publishEvent(ctx, &domain.NodeEvent{
		NodeID:    id,
		EventType: "node.restarting",
		Severity:  "info",
		Message:   fmt.Sprintf("Restarting node '%s'", node.Name),
	})

	// In a real implementation, this would trigger the node restart via its management API
	return nil
}

// ForceSync forces a synchronization update for a node
func (s *NodeService) ForceSync(ctx context.Context, id string) error {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}
	if node == nil {
		return fmt.Errorf("node not found: %s", id)
	}

	s.publishEvent(ctx, &domain.NodeEvent{
		NodeID:    id,
		EventType: "node.sync_requested",
		Severity:  "info",
		Message:   fmt.Sprintf("Force sync requested for node '%s'", node.Name),
	})

	return nil
}

// GetNetworkNodes retrieves all nodes for a specific network
func (s *NodeService) GetNetworkNodes(ctx context.Context, network string) ([]*domain.Node, error) {
	return s.nodeRepo.GetNodesByNetwork(network)
}

// ListNetworks lists all configured networks
func (s *NodeService) ListNetworks(ctx context.Context) ([]*domain.NetworkInfo, error) {
	networks := make([]*domain.NetworkInfo, 0)

	for key, cfg := range s.config.Blockchains {
		nodes, err := s.nodeRepo.GetNodesByNetwork(key)
		if err != nil {
			return nil, fmt.Errorf("failed to get nodes for network %s: %w", key, err)
		}

		activeNodes := 0
		for _, n := range nodes {
			if n.Status == domain.NodeStatusOnline {
				activeNodes++
			}
		}

		networks = append(networks, &domain.NetworkInfo{
			Name:           cfg.Name,
			ChainID:        cfg.ChainID,
			NativeCurrency: domain.NativeCurrency{
				Name:     cfg.NativeCurrency.Name,
				Symbol:   cfg.NativeCurrency.Symbol,
				Decimals: cfg.NativeCurrency.Decimals,
			},
			BlockTime:     cfg.BlockTime,
			MinPeers:      cfg.MinPeers,
			ExplorerURL:   cfg.ExplorerURL,
			ActiveNodes:   activeNodes,
			TotalNodes:    len(nodes),
		})
	}

	return networks, nil
}

// publishEvent publishes a node event to Kafka
func (s *NodeService) publishEvent(ctx context.Context, event *domain.NodeEvent) {
	event.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	event.Timestamp = time.Now()

	if s.producer != nil {
		s.producer.Publish(ctx, s.config.Kafka.Topics.NodeEvents, event)
	}
}
