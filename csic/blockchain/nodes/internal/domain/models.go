package domain

import (
	"time"
)

// NodeStatus represents the operational status of a blockchain node
type NodeStatus string

const (
	NodeStatusOnline  NodeStatus = "online"
	NodeStatusOffline NodeStatus = "offline"
	NodeStatusSyncing NodeStatus = "syncing"
	NodeStatusError   NodeStatus = "error"
	NodeStatusUnknown NodeStatus = "unknown"
)

// NodeType represents the type of blockchain node
type NodeType string

const (
	NodeTypeFull    NodeType = "full"
	NodeTypeArchive NodeType = "archive"
	NodeTypeLight   NodeType = "light"
	NodeTypeValidator NodeType = "validator"
)

// Node represents a blockchain node instance
type Node struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Network         string            `json:"network"`
	Type            NodeType          `json:"type"`
	RPCURL          string            `json:"rpc_url"`
	WSURL           string            `json:"ws_url"`
	Status          NodeStatus        `json:"status"`
	Version         string            `json:"version"`
	ChainID         int64             `json:"chain_id"`
	BlockNumber     int64             `json:"block_number"`
	PeerCount       int               `json:"peer_count"`
	LastSyncTime    time.Time         `json:"last_sync_time"`
	LastHealthCheck time.Time         `json:"last_health_check"`
	Metadata        map[string]string `json:"metadata"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// NodeMetrics represents performance metrics for a blockchain node
type NodeMetrics struct {
	ID            string    `json:"id"`
	NodeID        string    `json:"node_id"`
	Timestamp     time.Time `json:"timestamp"`
	BlockHeight   int64     `json:"block_height"`
	BlockTime     float64   `json:"block_time"`
	PeerCount     int       `json:"peer_count"`
	TransactionPoolSize int  `json:"transaction_pool_size"`
	GasPrice      float64   `json:"gas_price"`
	MemoryUsage   int64     `json:"memory_usage"`
	CPUUsage      float64   `json:"cpu_usage"`
	DiskUsage     int64     `json:"disk_usage"`
	NetworkIn     int64     `json:"network_in"`
	NetworkOut    int64     `json:"network_out"`
	RPCLatency    float64   `json:"rpc_latency"`
	SyncProgress  float64   `json:"sync_progress"`
}

// NetworkInfo represents information about a blockchain network
type NetworkInfo struct {
	Name           string            `json:"name"`
	ChainID        int64             `json:"chain_id"`
	NativeCurrency NativeCurrency    `json:"native_currency"`
	BlockTime      int               `json:"block_time"`
	MinPeers       int               `json:"min_peers"`
	ExplorerURL    string            `json:"explorer_url"`
	ActiveNodes    int               `json:"active_nodes"`
	TotalNodes     int               `json:"total_nodes"`
}

// NativeCurrency represents the native currency of a blockchain
type NativeCurrency struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

// NodeEvent represents an event related to a blockchain node
type NodeEvent struct {
	ID        string    `json:"id"`
	NodeID    string    `json:"node_id"`
	EventType string    `json:"event_type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	NodeID       string            `json:"node_id"`
	Status       NodeStatus        `json:"status"`
	Latency      time.Duration     `json:"latency"`
	Error        string            `json:"error"`
	Details      map[string]interface{} `json:"details"`
	CheckedAt    time.Time         `json:"checked_at"`
}

// CreateNodeRequest represents a request to create a new node
type CreateNodeRequest struct {
	Name    string            `json:"name" binding:"required"`
	Network string            `json:"network" binding:"required"`
	Type    NodeType          `json:"type"`
	RPCURL  string            `json:"rpc_url" binding:"required"`
	WSURL   string            `json:"ws_url"`
	Metadata map[string]string `json:"metadata"`
}

// UpdateNodeRequest represents a request to update an existing node
type UpdateNodeRequest struct {
	Name    *string            `json:"name"`
	RPCURL  *string            `json:"rpc_url"`
	WSURL   *string            `json:"ws_url"`
	Type    *NodeType          `json:"type"`
	Metadata map[string]string `json:"metadata"`
}

// NodeListFilter represents filter options for listing nodes
type NodeListFilter struct {
	Network   string     `json:"network"`
	Status    NodeStatus `json:"status"`
	Type      NodeType   `json:"type"`
	Offset    int        `json:"offset"`
	Limit     int        `json:"limit"`
}

// PaginatedNodes represents a paginated list of nodes
type PaginatedNodes struct {
	Nodes      []*Node `json:"nodes"`
	Total      int     `json:"total"`
	Offset     int     `json:"offset"`
	Limit      int     `json:"limit"`
}

// PaginatedMetrics represents a paginated list of metrics
type PaginatedMetrics struct {
	Metrics  []*NodeMetrics `json:"metrics"`
	Total    int            `json:"total"`
	Offset   int            `json:"offset"`
	Limit    int            `json:"limit"`
}

// SystemSummary represents a summary of all nodes and networks
type SystemSummary struct {
	TotalNodes     int               `json:"total_nodes"`
	OnlineNodes    int               `json:"online_nodes"`
	OfflineNodes   int               `json:"offline_nodes"`
	SyncingNodes   int               `json:"syncing_nodes"`
	Networks       []*NetworkSummary `json:"networks"`
}

// NetworkSummary represents a summary for a specific network
type NetworkSummary struct {
	Network     string `json:"network"`
	Name        string `json:"name"`
	TotalNodes  int    `json:"total_nodes"`
	OnlineNodes int    `json:"online_nodes"`
	AvgBlockHeight int64 `json:"avg_block_height"`
}
