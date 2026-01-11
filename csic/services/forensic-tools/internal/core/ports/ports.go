package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
)

// GraphRepository defines the interface for persisting graph data
type GraphRepository interface {
	// Node operations
	StoreNode(ctx context.Context, node *domain.GraphNode) error
	GetNode(ctx context.Context, id string) (*domain.GraphNode, error)
	GetNodeByAddress(ctx context.Context, address string) (*domain.GraphNode, error)
	ListNodes(ctx context.Context, limit, offset int) ([]*domain.GraphNode, error)
	DeleteNode(ctx context.Context, id string) error

	// Edge operations
	StoreEdge(ctx context.Context, edge *domain.GraphEdge) error
	GetEdge(ctx context.Context, id string) (*domain.GraphEdge, error)
	GetEdgesBySource(ctx context.Context, sourceID string) ([]*domain.GraphEdge, error)
	GetEdgesByTarget(ctx context.Context, targetID string) ([]*domain.GraphEdge, error)
	DeleteEdge(ctx context.Context, id string) error

	// Batch operations
	StoreGraph(ctx context.Context, nodes []*domain.GraphNode, edges []*domain.GraphEdge) error
	GetSubgraph(ctx context.Context, nodeIDs []string) ([]*domain.GraphNode, []*domain.GraphEdge, error)
}

// EvidenceRepository defines the interface for persisting forensic evidence
type EvidenceRepository interface {
	// Evidence CRUD
	StoreEvidence(ctx context.Context, evidence *domain.Evidence) error
	GetEvidence(ctx context.Context, id string) (*domain.Evidence, error)
	ListEvidence(ctx context.Context, caseID string, limit, offset int) ([]*domain.Evidence, error)
	DeleteEvidence(ctx context.Context, id string) error

	// Chain of custody
	AppendCustodyRecord(ctx context.Context, evidenceID string, record *domain.ChainOfCustody) error
	GetCustodyHistory(ctx context.Context, evidenceID string) ([]*domain.ChainOfCustody, error)

	// Evidence verification
	VerifyEvidence(ctx context.Context, id string) (bool, error)
}

// ReportRepository defines the interface for persisting forensic reports
type ReportRepository interface {
	StoreReport(ctx context.Context, report *domain.Report) error
	GetReport(ctx context.Context, id string) (*domain.Report, error)
	ListReports(ctx context.Context, caseID string, limit, offset int) ([]*domain.Report, error)
	DeleteReport(ctx context.Context, id string) error

	// Report generation
	GenerateReport(ctx context.Context, caseID string, reportType domain.ReportType) (*domain.Report, error)
}

// ClusterRepository defines the interface for persisting clustering results
type ClusterRepository interface {
	StoreCluster(ctx context.Context, cluster *domain.Cluster) error
	GetCluster(ctx context.Context, id string) (*domain.Cluster, error)
	ListClusters(ctx context.Context, caseID string) ([]*domain.Cluster, error)
	DeleteCluster(ctx context.Context, id string) error

	// Cluster membership
	GetClusterMembers(ctx context.Context, clusterID string) ([]string, error)
	AddClusterMembers(ctx context.Context, clusterID string, nodeIDs []string) error
	RemoveClusterMembers(ctx context.Context, clusterID string, nodeIDs []string) error
}

// GraphService defines the input port for graph analysis operations
type GraphService interface {
	// Transaction graph construction
	BuildTransactionGraph(ctx context.Context, txHashes []string) (*domain.GraphNode, []*domain.GraphEdge, error)
	ExpandGraph(ctx context.Context, nodeID string, depth int) ([]*domain.GraphNode, []*domain.GraphEdge, error)

	// Graph analysis
	DetectCommunities(ctx context.Context, nodeIDs []string) ([]*domain.Cluster, error)
	FindCentralNodes(ctx context.Context, nodeIDs []string, limit int) ([]*domain.GraphNode, error)
	FindPath(ctx context.Context, sourceID, targetID string) ([]*domain.GraphNode, []*domain.GraphEdge, error)

	// Graph metrics
	CalculateCentrality(ctx context.Context, nodeID string) (float64, error)
	CalculateClusteringCoefficient(ctx context.Context, nodeID string) (float64, error)
	GetConnectedComponents(ctx context.Context, nodeIDs []string) ([]*domain.Cluster, error)
}

// EvidenceService defines the input port for evidence management operations
type EvidenceService interface {
	// Evidence collection
	CollectTransactionEvidence(ctx context.Context, txHash string) (*domain.Evidence, error)
	CollectWalletEvidence(ctx context.Context, address string) (*domain.Evidence, error)
	CollectClusterEvidence(ctx context.Context, clusterID string) (*domain.Evidence, error)

	// Evidence verification
	VerifyHashIntegrity(ctx context.Context, evidenceID string) (bool, error)
	CalculateEvidenceHash(ctx context.Context, evidence *domain.Evidence) (string, error)

	// Chain of custody
	RecordCustody(ctx context.Context, evidenceID string, handler string, action string) error
	GetEvidenceTimeline(ctx context.Context, evidenceID string) ([]*domain.ChainOfCustody, error)

	// Evidence export
	ExportEvidence(ctx context.Context, evidenceID string, format string) ([]byte, error)
	PackageForLegal(ctx context.Context, evidenceIDs []string) (*domain.Report, error)
}

// ForensicRepository is the composite interface for all forensic data access
type ForensicRepository interface {
	GraphRepository
	EvidenceRepository
	ReportRepository
	ClusterRepository
}
