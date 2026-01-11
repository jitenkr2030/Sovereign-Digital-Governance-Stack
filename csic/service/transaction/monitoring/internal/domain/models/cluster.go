package models

import (
	"time"

	"github.com/google/uuid"
)

// EntityCluster represents a cluster of related wallet addresses
type EntityCluster struct {
	ID               string    `json:"id" db:"id"`
	ClusterType      string    `json:"cluster_type" db:"cluster_type"` // behavioral, common-input, deposit, change
	PrimaryAddress   string    `json:"primary_address" db:"primary_address"`
	Label            string    `json:"label" db:"label"`
	Description      string    `json:"description" db:"description"`
	WalletCount      int       `json:"wallet_count" db:"wallet_count"`
	TotalVolume      string    `json:"total_volume" db:"total_volume"`
	TotalTxCount     int64     `json:"total_tx_count" db:"total_tx_count"`
	RiskScore        float64   `json:"risk_score" db:"risk_score"`
	RiskLevel        string    `json:"risk_level" db:"risk_level"`
	ConfidenceScore  float64   `json:"confidence_score" db:"confidence_score"`
	Tags             []string  `json:"tags" db:"-"`
	IsVerified       bool      `json:"is_verified" db:"is_verified"`
	VerifiedBy       *string   `json:"verified_by" db:"verified_by"`
	SuspectedEntity  string    `json:"suspected_entity" db:"suspected_entity"`
	DiscoveryMethod  string    `json:"discovery_method" db:"discovery_method"`
	FirstSeen        time.Time `json:"first_seen" db:"first_seen"`
	LastActivity     time.Time `json:"last_activity" db:"last_activity"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// ClusterMember represents a wallet address that is a member of a cluster
type ClusterMember struct {
	ID            string    `json:"id" db:"id"`
	ClusterID     string    `json:"cluster_id" db:"cluster_id"`
	WalletAddress string    `json:"wallet_address" db:"wallet_address"`
	Network       Network   `json:"network" db:"network"`
	MemberType    string    `json:"member_type" db:"member_type"` // primary, linked
	LinkType      string    `json:"link_type" db:"link_type"`     // common-input, deposit, change, behavioral
	LinkStrength  float64   `json:"link_strength" db:"link_strength"`
	TxCount       int64     `json:"tx_count" db:"tx_count"`
	Volume        string    `json:"volume" db:"volume"`
	FirstLinkedAt time.Time `json:"first_linked_at" db:"first_linked_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// ClusterRelationship represents relationships between clusters
type ClusterRelationship struct {
	ID              string    `json:"id" db:"id"`
	SourceClusterID string    `json:"source_cluster_id" db:"source_cluster_id"`
	TargetClusterID string    `json:"target_cluster_id" db:"target_cluster_id"`
	RelationshipType string   `json:"relationship_type" db:"relationship_type"` // transaction, ownership, behavioral
	TxCount         int64     `json:"tx_count" db:"tx_count"`
	TotalVolume     string    `json:"total_volume" db:"total_volume"`
	FirstSeen       time.Time `json:"first_seen" db:"first_seen"`
	LastSeen        time.Time `json:"last_seen" db:"last_seen"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// ClusterActivity represents cluster activity for monitoring
type ClusterActivity struct {
	ID              string    `json:"id" db:"id"`
	ClusterID       string    `json:"cluster_id" db:"cluster_id"`
	ActivityType    string    `json:"activity_type" db:"activity_type"` // new_member, risk_increase, tx_detected
	Details         string    `json:"details" db:"details"`
	TxHash          *string   `json:"tx_hash" db:"tx_hash"`
	RelatedWallet   *string   `json:"related_wallet" db:"related_wallet"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// NewEntityCluster creates a new entity cluster with generated ID
func NewEntityCluster(clusterType, primaryAddress string) *EntityCluster {
	return &EntityCluster{
		ID:              uuid.New().String(),
		ClusterType:     clusterType,
		PrimaryAddress:  primaryAddress,
		WalletCount:     1,
		ConfidenceScore: 1.0,
		Tags:            []string{},
		FirstSeen:       time.Now(),
		LastActivity:    time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// AddMember adds a wallet address to the cluster
func (c *EntityCluster) AddMember(address string, network Network, linkType string, linkStrength float64) {
	c.WalletCount++
	c.ConfidenceScore = c.ConfidenceScore * linkStrength
	c.LastActivity = time.Now()
}

// UpdateRiskScore recalculates risk level based on score
func (c *EntityCluster) UpdateRiskScore(score float64) {
	c.RiskScore = score
	switch {
	case score >= 80:
		c.RiskLevel = "critical"
	case score >= 60:
		c.RiskLevel = "high"
	case score >= 40:
		c.RiskLevel = "medium"
	case score >= 20:
		c.RiskLevel = "low"
	default:
		c.RiskLevel = "minimal"
	}
}

// EntityClusterGraph represents the graph structure for clustering
type EntityClusterGraph struct {
	Nodes     map[string]*GraphNode
	Edges     []*GraphEdge
	Adjacency map[string][]string
}

// GraphNode represents a node in the clustering graph
type GraphNode struct {
	Address   string          `json:"address"`
	Network   Network         `json:"network"`
	TxCount   int64           `json:"tx_count"`
	Volume    string          `json:"volume"`
	Features  []float64       `json:"features"`
	ClusterID *string         `json:"cluster_id,omitempty"`
}

// GraphEdge represents an edge in the clustering graph
type GraphEdge struct {
	Source      string  `json:"source"`
	Target      string  `json:"target"`
	Weight      float64 `json:"weight"`
	LinkType    string  `json:"link_type"`
	TxCount     int64   `json:"tx_count"`
	FirstSeen   string  `json:"first_seen"`
	LastSeen    string  `json:"last_seen"`
}
