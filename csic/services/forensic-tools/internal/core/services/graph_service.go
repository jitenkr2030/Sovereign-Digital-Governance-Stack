package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/csic-platform/internal/core/domain"
	"github.com/csic-platform/internal/core/ports"
)

// GraphAnalysisService implements the GraphService interface
type GraphAnalysisService struct {
	repo ports.GraphRepository
	mu   sync.RWMutex
}

// NewGraphAnalysisService creates a new graph analysis service instance
func NewGraphAnalysisService(repo ports.GraphRepository) *GraphAnalysisService {
	return &GraphAnalysisService{
		repo: repo,
	}
}

// BuildTransactionGraph constructs a transaction graph from a list of transaction hashes
func (s *GraphAnalysisService) BuildTransactionGraph(ctx context.Context, txHashes []string) (*domain.GraphNode, []*domain.GraphEdge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var allNodes []*domain.GraphNode
	var allEdges []*domain.GraphEdge

	// Create root node for the graph
	rootNode := &domain.GraphNode{
		ID:        fmt.Sprintf("root-%d", len(txHashes)),
		NodeType:  domain.NodeTypeCluster,
		Label:     fmt.Sprintf("Transaction Bundle (%d txs)", len(txHashes)),
		Metadata:  map[string]interface{}{"tx_count": len(txHashes)},
		CreatedAt: s.now(),
	}

	for _, txHash := range txHashes {
		txNode := &domain.GraphNode{
			ID:        txHash,
			NodeType:  domain.NodeTypeTransaction,
			Label:     txHash[:16] + "...",
			Metadata:  map[string]interface{}{"tx_hash": txHash},
			CreatedAt: s.now(),
		}

		// Create edge from root to transaction
		edge := &domain.GraphEdge{
			ID:        fmt.Sprintf("edge-%s-root", txHash),
			SourceID:  rootNode.ID,
			TargetID:  txNode.ID,
			EdgeType:  domain.EdgeTypeContains,
			Weight:    1.0,
			CreatedAt: s.now(),
		}

		allNodes = append(allNodes, txNode)
		allEdges = append(allEdges, edge)

		// Build detailed transaction graph (inputs and outputs)
		inputNodes, inputEdges := s.buildInputGraph(ctx, txHash)
		outputNodes, outputEdges := s.buildOutputGraph(ctx, txHash)

		allNodes = append(allNodes, inputNodes...)
		allNodes = append(allNodes, outputNodes...)
		allEdges = append(allEdges, inputEdges...)
		allEdges = append(allEdges, outputEdges...)
	}

	// Store all nodes and edges
	if err := s.repo.StoreGraph(ctx, allNodes, allEdges); err != nil {
		return nil, nil, fmt.Errorf("failed to store graph: %w", err)
	}

	return rootNode, allEdges, nil
}

// buildInputGraph creates graph nodes for transaction inputs
func (s *GraphAnalysisService) buildInputGraph(ctx context.Context, txHash string) ([]*domain.GraphNode, []*domain.GraphEdge) {
	var nodes []*domain.GraphNode
	var edges []*domain.GraphEdge

	// Simulate input wallet nodes
	inputWallet := &domain.GraphNode{
		ID:        fmt.Sprintf("input-%s", txHash),
		NodeType:  domain.NodeTypeWallet,
		Label:     fmt.Sprintf("Input Wallet for %s...", txHash[:8]),
		Metadata:  map[string]interface{}{"tx_hash": txHash, "direction": "input"},
		CreatedAt: s.now(),
	}

	edge := &domain.GraphEdge{
		ID:        fmt.Sprintf("edge-input-%s", txHash),
		SourceID:  inputWallet.ID,
		TargetID:  txHash,
		EdgeType:  domain.EdgeTypeSpentFrom,
		Weight:    1.0,
		CreatedAt: s.now(),
	}

	nodes = append(nodes, inputWallet)
	edges = append(edges, edge)

	return nodes, edges
}

// buildOutputGraph creates graph nodes for transaction outputs
func (s *GraphAnalysisService) buildOutputGraph(ctx context.Context, txHash string) ([]*domain.GraphNode, []*domain.GraphEdge) {
	var nodes []*domain.GraphNode
	var edges []*domain.GraphEdge

	// Simulate output wallet nodes
	outputWallet := &domain.GraphNode{
		ID:        fmt.Sprintf("output-%s", txHash),
		NodeType:  domain.NodeTypeWallet,
		Label:     fmt.Sprintf("Output Wallet for %s...", txHash[:8]),
		Metadata:  map[string]interface{}{"tx_hash": txHash, "direction": "output"},
		CreatedAt: s.now(),
	}

	edge := &domain.GraphEdge{
		ID:        fmt.Sprintf("edge-output-%s", txHash),
		SourceID:  txHash,
		TargetID:  outputWallet.ID,
		EdgeType:  domain.EdgeTypeTransferredTo,
		Weight:    1.0,
		CreatedAt: s.now(),
	}

	nodes = append(nodes, outputWallet)
	edges = append(edges, edge)

	return nodes, edges
}

// ExpandGraph expands a graph node to the specified depth
func (s *GraphAnalysisService) ExpandGraph(ctx context.Context, nodeID string, depth int) ([]*domain.GraphNode, []*domain.GraphEdge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var allNodes []*domain.GraphNode
	var allEdges []*domain.GraphEdge
	visited := make(map[string]bool)
	queue := []string{nodeID}

	for i := 0; i < depth && len(queue) > 0; i++ {
		var nextQueue []string
		for _, currentID := range queue {
			if visited[currentID] {
				continue
			}
			visited[currentID] = true

			// Get the current node
			node, err := s.repo.GetNode(ctx, currentID)
			if err != nil {
				continue
			}
			allNodes = append(allNodes, node)

			// Get outgoing edges
			edges, err := s.repo.GetEdgesBySource(ctx, currentID)
			if err != nil {
				continue
			}
			for _, edge := range edges {
				allEdges = append(allEdges, edge)
				nextQueue = append(nextQueue, edge.TargetID)

				// Get the target node
				targetNode, err := s.repo.GetNode(ctx, edge.TargetID)
				if err == nil && !visited[edge.TargetID] {
					allNodes = append(allNodes, targetNode)
				}
			}

			// Get incoming edges
			incomingEdges, err := s.repo.GetEdgesByTarget(ctx, currentID)
			if err != nil {
				continue
			}
			for _, edge := range incomingEdges {
				allEdges = append(allEdges, edge)
				nextQueue = append(nextQueue, edge.SourceID)

				// Get the source node
				sourceNode, err := s.repo.GetNode(ctx, edge.SourceID)
				if err == nil && !visited[edge.SourceID] {
					allNodes = append(allNodes, sourceNode)
				}
			}
		}
		queue = nextQueue
	}

	return allNodes, allEdges, nil
}

// DetectCommunities identifies clusters of closely connected nodes using a simple algorithm
func (s *GraphAnalysisService) DetectCommunities(ctx context.Context, nodeIDs []string) ([]*domain.Cluster, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the subgraph
	nodes, edges, err := s.repo.GetSubgraph(ctx, nodeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get subgraph: %w", err)
	}

	// Build adjacency map
	adjacency := make(map[string]map[string]float64)
	for _, node := range nodes {
		adjacency[node.ID] = make(map[string]float64)
	}

	for _, edge := range edges {
		if _, ok := adjacency[edge.SourceID]; ok {
			adjacency[edge.SourceID][edge.TargetID] += edge.Weight
		}
		if _, ok := adjacency[edge.TargetID]; ok {
			adjacency[edge.TargetID][edge.SourceID] += edge.Weight
		}
	}

	// Simple community detection using label propagation
	clusters := s.labelPropagation(nodes, adjacency)

	return clusters, nil
}

// labelPropagation performs simple community detection
func (s *GraphAnalysisService) labelPropagation(nodes []*domain.GraphNode, adjacency map[string]map[string]float64) []*domain.Cluster {
	labels := make(map[string]string)
	for i, node := range nodes {
		labels[node.ID] = fmt.Sprintf("community-%d", i%5) // Start with random labels
	}

	// Iterative label propagation
	for iteration := 0; iteration < 10; iteration++ {
		for _, node := range nodes {
			frequency := make(map[string]int)
			for neighbor := range adjacency[node.ID] {
				if label, ok := labels[neighbor]; ok {
					frequency[label]++
				}
			}

			maxFreq := 0
			var bestLabel string
			for label, freq := range frequency {
				if freq > maxFreq {
					maxFreq = freq
					bestLabel = label
				}
			}
			if bestLabel != "" {
				labels[node.ID] = bestLabel
			}
		}
	}

	// Group nodes by label
	clusterMap := make(map[string][]string)
	for nodeID, label := range labels {
		clusterMap[label] = append(clusterMap[label], nodeID)
	}

	var clusters []*domain.Cluster
	for label, memberIDs := range clusterMap {
		cluster := &domain.Cluster{
			ID:        label,
			Name:      label,
			MemberIDs: memberIDs,
			Score:     float64(len(memberIDs)) / float64(len(nodes)),
			CreatedAt: s.now(),
		}
		clusters = append(clusters, cluster)
	}

	return clusters
}

// FindCentralNodes identifies the most central nodes in a graph
func (s *GraphAnalysisService) FindCentralNodes(ctx context.Context, nodeIDs []string, limit int) ([]*domain.GraphNode, error) {
	nodes, edges, err := s.repo.GetSubgraph(ctx, nodeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get subgraph: %w", err)
	}

	// Calculate degree centrality
	centrality := make(map[string]float64)
	for _, node := range nodes {
		centrality[node.ID] = 0
	}

	for _, edge := range edges {
		centrality[edge.SourceID]++
		centrality[edge.TargetID]++
	}

	// Sort by centrality
	type nodeScore struct {
		node  *domain.GraphNode
		score float64
	}
	var scoredNodes []nodeScore
	for _, node := range nodes {
		scoredNodes = append(scoredNodes, nodeScore{
			node:  node,
			score: centrality[node.ID],
		})
	}

	// Simple bubble sort for small sets
	for i := 0; i < len(scoredNodes); i++ {
		for j := i + 1; j < len(scoredNodes); j++ {
			if scoredNodes[j].score > scoredNodes[i].score {
				scoredNodes[i], scoredNodes[j] = scoredNodes[j], scoredNodes[i]
			}
		}
	}

	if limit > len(scoredNodes) {
		limit = len(scoredNodes)
	}

	var result []*domain.GraphNode
	for i := 0; i < limit; i++ {
		result = append(result, scoredNodes[i].node)
	}

	return result, nil
}

// FindPath finds a path between two nodes using BFS
func (s *GraphAnalysisService) FindPath(ctx context.Context, sourceID, targetID string) ([]*domain.GraphNode, []*domain.GraphEdge, error) {
	visited := make(map[string]bool)
	queue := []string{sourceID}
	parents := make(map[string]string)
	edgeParents := make(map[string]string)

	visited[sourceID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == targetID {
			// Reconstruct path
			var pathNodes []*domain.GraphNode
			var pathEdges []*domain.GraphEdge

			currentID := targetID
			for currentID != sourceID {
				node, err := s.repo.GetNode(ctx, currentID)
				if err != nil {
					break
				}
				pathNodes = append([]*domain.GraphNode{node}, pathNodes...)

				if edgeID, ok := edgeParents[currentID]; ok {
					edge, _ := s.repo.GetEdge(ctx, edgeID)
					if edge != nil {
						pathEdges = append([]*domain.GraphEdge{edge}, pathEdges...)
					}
				}
				currentID = parents[currentID]
			}

			// Add source node
			sourceNode, _ := s.repo.GetNode(ctx, sourceID)
			if sourceNode != nil {
				pathNodes = append([]*domain.GraphNode{sourceNode}, pathNodes...)
			}

			return pathNodes, pathEdges, nil
		}

		// Check outgoing edges
		edges, _ := s.repo.GetEdgesBySource(ctx, current)
		for _, edge := range edges {
			if !visited[edge.TargetID] {
				visited[edge.TargetID] = true
				parents[edge.TargetID] = current
				edgeParents[edge.TargetID] = edge.ID
				queue = append(queue, edge.TargetID)
			}
		}
	}

	return nil, nil, fmt.Errorf("no path found between %s and %s", sourceID, targetID)
}

// CalculateCentrality calculates the degree centrality of a node
func (s *GraphAnalysisService) CalculateCentrality(ctx context.Context, nodeID string) (float64, error) {
	node, err := s.repo.GetNode(ctx, nodeID)
	if err != nil {
		return 0, err
	}

	edges, err := s.repo.GetEdgesBySource(ctx, nodeID)
	if err != nil {
		return 0, err
	}

	incomingEdges, err := s.repo.GetEdgesByTarget(ctx, nodeID)
	if err != nil {
		return 0, err
	}

	totalConnections := float64(len(edges) + len(incomingEdges))
	node.Metadata["centrality"] = totalConnections

	return totalConnections, nil
}

// CalculateClusteringCoefficient calculates the local clustering coefficient
func (s *GraphAnalysisService) CalculateClusteringCoefficient(ctx context.Context, nodeID string) (float64, error) {
	edges, _ := s.repo.GetEdgesBySource(ctx, nodeID)
	incomingEdges, _ := s.repo.GetEdgesByTarget(ctx, nodeID)

	// Get neighbors
	neighbors := make(map[string]bool)
	for _, edge := range edges {
		neighbors[edge.TargetID] = true
	}
	for _, edge := range incomingEdges {
		neighbors[edge.SourceID] = true
	}

	neighborList := make([]string, 0, len(neighbors))
	for n := range neighbors {
		neighborList = append(neighborList, n)
	}

	if len(neighborList) < 2 {
		return 0, nil
	}

	// Count triangles
	triangles := 0
	for i, n1 := range neighborList {
		for j := i + 1; j < len(neighborList); j++ {
			n2 := neighborList[j]
			// Check if n1 and n2 are connected
			n1ToN2, _ := s.repo.GetEdgesBySource(ctx, n1)
			for _, e := range n1ToN2 {
				if e.TargetID == n2 {
					triangles++
					break
				}
			}
		}
	}

	// Calculate clustering coefficient
	maxTriangles := float64(len(neighborList) * (len(neighborList) - 1) / 2)
	if maxTriangles == 0 {
		return 0, nil
	}

	return float64(triangles) / maxTriangles, nil
}

// GetConnectedComponents identifies connected components in the graph
func (s *GraphAnalysisService) GetConnectedComponents(ctx context.Context, nodeIDs []string) ([]*domain.Cluster, error) {
	visited := make(map[string]bool)
	var components []*domain.Cluster

	for _, nodeID := range nodeIDs {
		if visited[nodeID] {
			continue
		}

		// BFS to find component
		var component []string
		queue := []string{nodeID}
		visited[nodeID] = true

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			component = append(component, current)

			// Check neighbors
			edges, _ := s.repo.GetEdgesBySource(ctx, current)
			for _, edge := range edges {
				if !visited[edge.TargetID] {
					visited[edge.TargetID] = true
					queue = append(queue, edge.TargetID)
				}
			}
			incomingEdges, _ := s.repo.GetEdgesByTarget(ctx, current)
			for _, edge := range incomingEdges {
				if !visited[edge.SourceID] {
					visited[edge.SourceID] = true
					queue = append(queue, edge.SourceID)
				}
			}
		}

		cluster := &domain.Cluster{
			ID:        fmt.Sprintf("component-%d", len(components)),
			Name:      fmt.Sprintf("Connected Component %d", len(components)+1),
			MemberIDs: component,
			Score:     float64(len(component)),
			CreatedAt: s.now(),
		}
		components = append(components, cluster)
	}

	return components, nil
}

// now returns current timestamp (for testing purposes)
func (s *GraphAnalysisService) now() int64 {
	return 0 // Will be replaced with actual timestamp in production
}
