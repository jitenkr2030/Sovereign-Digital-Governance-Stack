package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/csic/platform/blockchain/nodes/internal/config"
	"github.com/csic/platform/blockchain/nodes/internal/domain"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// NodeRepository handles database operations for nodes
type NodeRepository struct {
	db        *sql.DB
	keyPrefix string
}

// NewNodeRepository creates a new NodeRepository instance
func NewNodeRepository(db *sql.DB, keyPrefix string) *NodeRepository {
	return &NodeRepository{
		db:        db,
		keyPrefix: keyPrefix,
	}
}

// Create creates a new node record
func (r *NodeRepository) Create(node *domain.Node) error {
	query := `
		INSERT INTO nodes (
			id, name, network, type, rpc_url, ws_url, status,
			version, chain_id, block_number, peer_count, last_sync_time,
			last_health_check, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	metadata, err := marshalMetadata(node.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	node.ID = uuid.New().String()
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()
	if node.Status == "" {
		node.Status = domain.NodeStatusUnknown
	}

	_, err = r.db.Exec(query,
		node.ID, node.Name, node.Network, node.Type, node.RPCURL, node.WSURL,
		node.Status, node.Version, node.ChainID, node.BlockNumber, node.PeerCount,
		node.LastSyncTime, node.LastHealthCheck, metadata, node.CreatedAt, node.UpdatedAt,
	)

	return err
}

// GetByID retrieves a node by its ID
func (r *NodeRepository) GetByID(id string) (*domain.Node, error) {
	query := `
		SELECT id, name, network, type, rpc_url, ws_url, status,
			version, chain_id, block_number, peer_count, last_sync_time,
			last_health_check, metadata, created_at, updated_at
		FROM nodes WHERE id = $1
	`

	var node domain.Node
	var metadataJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&node.ID, &node.Name, &node.Network, &node.Type, &node.RPCURL, &node.WSURL,
		&node.Status, &node.Version, &node.ChainID, &node.BlockNumber, &node.PeerCount,
		&node.LastSyncTime, &node.LastHealthCheck, &metadataJSON, &node.CreatedAt, &node.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	node.Metadata, err = unmarshalMetadata(metadataJSON)
	return &node, err
}

// Update updates an existing node record
func (r *NodeRepository) Update(node *domain.Node) error {
	query := `
		UPDATE nodes SET
			name = $1, rpc_url = $2, ws_url = $3, type = $4,
			metadata = $5, updated_at = $6
		WHERE id = $7
	`

	metadata, err := marshalMetadata(node.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	node.UpdatedAt = time.Now()

	_, err = r.db.Exec(query,
		node.Name, node.RPCURL, node.WSURL, node.Type,
		metadata, node.UpdatedAt, node.ID,
	)

	return err
}

// UpdateStatus updates the status of a node
func (r *NodeRepository) UpdateStatus(id string, status domain.NodeStatus, details map[string]interface{}) error {
	query := `
		UPDATE nodes SET
			status = $1, last_health_check = $2, updated_at = $3
		WHERE id = $4
	`

	now := time.Now()
	_, err := r.db.Exec(query, status, now, now, id)
	return err
}

// UpdateSyncInfo updates synchronization information for a node
func (r *NodeRepository) UpdateSyncInfo(id string, blockNumber int64, peerCount int) error {
	query := `
		UPDATE nodes SET
			block_number = $1, peer_count = $2, last_sync_time = $3, updated_at = $4
		WHERE id = $5
	`

	now := time.Now()
	_, err := r.db.Exec(query, blockNumber, peerCount, now, now, id)
	return err
}

// Delete deletes a node by its ID
func (r *NodeRepository) Delete(id string) error {
	query := "DELETE FROM nodes WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}

// List retrieves all nodes with optional filtering
func (r *NodeRepository) List(filter *domain.NodeListFilter) (*domain.PaginatedNodes, error) {
	baseQuery := "SELECT id, name, network, type, rpc_url, ws_url, status, version, chain_id, block_number, peer_count, last_sync_time, last_health_check, metadata, created_at, updated_at FROM nodes"
	countQuery := "SELECT COUNT(*) FROM nodes"

	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if filter.Network != "" {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("network = $%d", argIndex)
		args = append(args, filter.Network)
		argIndex++
	}

	if filter.Status != "" {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.Type != "" {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("type = $%d", argIndex)
		args = append(args, filter.Type)
		argIndex++
	}

	if whereClause != "" {
		baseQuery += " WHERE " + whereClause
		countQuery += " WHERE " + whereClause
	}

	// Add pagination
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.Limit, filter.Offset)

	// Get total count
	var total int
	countArgs := args[:len(args)-2]
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Execute query
	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := make([]*domain.Node, 0)
	for rows.Next() {
		var node domain.Node
		var metadataJSON []byte

		err := rows.Scan(
			&node.ID, &node.Name, &node.Network, &node.Type, &node.RPCURL, &node.WSURL,
			&node.Status, &node.Version, &node.ChainID, &node.BlockNumber, &node.PeerCount,
			&node.LastSyncTime, &node.LastHealthCheck, &metadataJSON, &node.CreatedAt, &node.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		node.Metadata, _ = unmarshalMetadata(metadataJSON)
		nodes = append(nodes, &node)
	}

	return &domain.PaginatedNodes{
		Nodes:  nodes,
		Total:  total,
		Offset: filter.Offset,
		Limit:  filter.Limit,
	}, nil
}

// GetNodesByNetwork retrieves all nodes for a specific network
func (r *NodeRepository) GetNodesByNetwork(network string) ([]*domain.Node, error) {
	query := `
		SELECT id, name, network, type, rpc_url, ws_url, status, version, chain_id,
			block_number, peer_count, last_sync_time, last_health_check, metadata, created_at, updated_at
		FROM nodes WHERE network = $1 ORDER BY name ASC
	`

	rows, err := r.db.Query(query, network)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := make([]*domain.Node, 0)
	for rows.Next() {
		var node domain.Node
		var metadataJSON []byte

		err := rows.Scan(
			&node.ID, &node.Name, &node.Network, &node.Type, &node.RPCURL, &node.WSURL,
			&node.Status, &node.Version, &node.ChainID, &node.BlockNumber, &node.PeerCount,
			&node.LastSyncTime, &node.LastHealthCheck, &metadataJSON, &node.CreatedAt, &node.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		node.Metadata, _ = unmarshalMetadata(metadataJSON)
		nodes = append(nodes, &node)
	}

	return nodes, nil
}

// GetNetworkStats retrieves statistics for all networks
func (r *NodeRepository) GetNetworkStats() (map[string]*domain.NetworkSummary, error) {
	query := `
		SELECT network, COUNT(*) as total,
			COUNT(CASE WHEN status = 'online' THEN 1 END) as online,
			AVG(block_number) as avg_height
		FROM nodes GROUP BY network
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]*domain.NetworkSummary)
	for rows.Next() {
		var summary domain.NetworkSummary
		var avgHeight sql.NullFloat64

		err := rows.Scan(&summary.Network, &summary.TotalNodes, &summary.OnlineNodes, &avgHeight)
		if err != nil {
			return nil, err
		}

		if avgHeight.Valid {
			summary.AvgBlockHeight = int64(avgHeight.Float64)
		}

		stats[summary.Network] = &summary
	}

	return stats, nil
}

// marshalMetadata converts map to JSON bytes
func marshalMetadata(m map[string]string) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return []byte(`{}`), nil // Simplified for now
}

// unmarshalMetadata converts JSON bytes to map
func unmarshalMetadata(data []byte) (map[string]string, error) {
	return make(map[string]string), nil // Simplified for now
}
