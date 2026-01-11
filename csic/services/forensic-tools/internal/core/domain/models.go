package domain

import (
	"time"

	"github.com/google/uuid"
)

// CaseStatus represents the status of a forensic case
type CaseStatus string

const (
	CaseStatusOpen         CaseStatus = "OPEN"
	CaseStatusInProgress   CaseStatus = "IN_PROGRESS"
	CaseStatusUnderReview  CaseStatus = "UNDER_REVIEW"
	CaseStatusPendingEvidence CaseStatus = "PENDING_EVIDENCE"
	CaseStatusResolved     CaseStatus = "RESOLVED"
	CaseStatusClosed       CaseStatus = "CLOSED"
)

// CasePriority represents case priority
type CasePriority string

const (
	CasePriorityLow      CasePriority = "LOW"
	CasePriorityMedium   CasePriority = "MEDIUM"
	CasePriorityHigh     CasePriority = "HIGH"
	CasePriorityCritical CasePriority = "CRITICAL"
)

// Case represents a forensic investigation case
type Case struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	CaseNumber      string        `json:"case_number" db:"case_number"`
	Title           string        `json:"title" db:"title"`
	Description     string        `json:"description" db:"description"`
	CaseType        string        `json:"case_type" db:"case_type"`
	Status          CaseStatus    `json:"status" db:"status"`
	Priority        CasePriority  `json:"priority" db:"priority"`
	InvestigatorID  uuid.UUID     `json:"investigating_user_id" db:"investigating_user_id"`
	InvestigatorName string       `json:"investigating_user_name" db:"investigating_user_name"`
	AssignedTeam    []string      `json:"assigned_team" db:"assigned_team"`
	Tags            []string      `json:"tags" db:"tags"`
	RelatedCases    []uuid.UUID   `json:"related_cases" db:"related_cases"`
	SuspectWallets  []string      `json:"suspect_wallets" db:"suspect_wallets"`
	TotalEvidence   int           `json:"total_evidence" db:"total_evidence"`
	RiskScore       float64       `json:"risk_score" db:"risk_score"`
	Findings        string        `json:"findings,omitempty" db:"findings"`
	Recommendations string        `json:"recommendations,omitempty" db:"recommendations"`
	OpenedAt        time.Time     `json:"opened_at" db:"opened_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
	ResolvedAt      *time.Time    `json:"resolved_at,omitempty" db:"resolved_at"`
	ClosedAt        *time.Time    `json:"closed_at,omitempty" db:"closed_at"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
}

// EvidenceType represents the type of evidence
type EvidenceType string

const (
	EvidenceTypeTransaction    EvidenceType = "TRANSACTION"
	EvidenceTypeWallet         EvidenceType = "WALLET"
	EvidenceTypeGraph          EvidenceType = "GRAPH"
	EvidenceTypeReport         EvidenceType = "REPORT"
	EvidenceTypeDocument       EvidenceType = "DOCUMENT"
	EvidenceTypeScreenshot     EvidenceType = "SCREENSHOT"
	EvidenceTypeBlockchainData EvidenceType = "BLOCKCHAIN_DATA"
	EvidenceTypeCommunication  EvidenceType = "COMMUNICATION"
)

// EvidenceStatus represents the status of evidence
type EvidenceStatus string

const (
	EvidenceStatusCollected  EvidenceStatus = "COLLECTED"
	EvidenceStatusVerified   EvidenceStatus = "VERIFIED"
	EvidenceStatusSealed     EvidenceStatus = "SEALED"
	EvidenceStatusReleased   EvidenceStatus = "RELEASED"
	EvidenceStatusArchived   EvidenceStatus = "ARCHIVED"
)

// Evidence represents forensic evidence
type Evidence struct {
	ID              uuid.UUID      `json:"id" db:"id"`
	CaseID          uuid.UUID      `json:"case_id" db:"case_id"`
	EvidenceType    EvidenceType   `json:"evidence_type" db:"evidence_type"`
	Title           string         `json:"title" db:"title"`
	Description     string         `json:"description" db:"description"`
	ContentHash     string         `json:"content_hash" db:"content_hash"`
	Content         string         `json:"content,omitempty" db:"content"`
	ContentSize     int64          `json:"content_size" db:"content_size"`
	Source          string         `json:"source" db:"source"`
	SourceURL       string         `json:"source_url,omitempty" db:"source_url"`
	ChainOfCustody  []CustodyEntry `json:"chain_of_custody" db:"chain_of_custody"`
	HashAlgorithm   string         `json:"hash_algorithm" db:"hash_algorithm"`
	IntegrityHash   string         `json:"integrity_hash" db:"integrity_hash"`
	VerificationStatus string     `json:"verification_status" db:"verification_status"`
	CollectedBy     uuid.UUID      `json:"collected_by" db:"collected_by"`
	CollectedAt     time.Time      `json:"collected_at" db:"collected_at"`
	VerifiedAt      *time.Time     `json:"verified_at,omitempty" db:"verified_at"`
	VerifiedBy      *uuid.UUID     `json:"verified_by,omitempty" db:"verified_by"`
	Status          EvidenceStatus `json:"status" db:"status"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

// CustodyEntry represents a chain of custody entry
type CustodyEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Action      string    `json:"action"`
	ActorID     uuid.UUID `json:"actor_id"`
	ActorName   string    `json:"actor_name"`
	Location    string    `json:"location"`
	Notes       string    `json:"notes"`
	Signature   string    `json:"signature"`
}

// GraphNode represents a node in a transaction graph
type GraphNode struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Address      string    `json:"address" db:"address"`
	AddressType  string    `json:"address_type" db:"address_type"`
	EntityLabel  string    `json:"entity_label,omitempty" db:"entity_label"`
	EntityType   string    `json:"entity_type,omitempty" db:"entity_type"`
	Balance      float64   `json:"balance" db:"balance"`
	FirstSeen    time.Time `json:"first_seen" db:"first_seen"`
	LastSeen     time.Time `json:"last_seen" db:"last_seen"`
	Labels       []string  `json:"labels,omitempty" db:"labels"`
	ClusterID    *uuid.UUID `json:"cluster_id,omitempty" db:"cluster_id"`
	RiskScore    float64   `json:"risk_score" db:"risk_score"`
	KnownEntity  bool      `json:"known_entity" db:"known_entity"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// GraphEdge represents an edge in a transaction graph
type GraphEdge struct {
	ID            uuid.UUID `json:"id" db:"id"`
	TransactionID uuid.UUID `json:"transaction_id" db:"transaction_id"`
	SourceNode    string    `json:"source_node" db:"source_node"`
	TargetNode    string    `json:"target_node" db:"target_node"`
	Value         float64   `json:"value" db:"value"`
	AssetType     string    `json:"asset_type" db:"asset_type"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
	BlockNumber   uint64    `json:"block_number" db:"block_number"`
	TxHash        string    `json:"tx_hash" db:"tx_hash"`
	GasFee        float64   `json:"gas_fee" db:"gas_fee"`
	Direction     string    `json:"direction" db:"direction"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// TransactionGraph represents a complete transaction graph
type TransactionGraph struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	CaseID      uuid.UUID   `json:"case_id,omitempty" db:"case_id"`
	RootAddress string      `json:"root_address" db:"root_address"`
	Depth       int         `json:"depth" db:"depth"`
	Nodes       []GraphNode `json:"nodes" db:"nodes"`
	Edges       []GraphEdge `json:"edges" db:"edges"`
	PathCount   int         `json:"path_count" db:"path_count"`
	NodeCount   int         `json:"node_count" db:"node_count"`
	EdgeCount   int         `json:"edge_count" db:"edge_count"`
	TotalVolume float64     `json:"total_volume" db:"total_volume"`
	GeneratedAt time.Time   `json:"generated_at" db:"generated_at"`
}

// WalletCluster represents a cluster of related wallets
type WalletCluster struct {
	ID            uuid.UUID   `json:"id" db:"id"`
	ClusterType   string      `json:"cluster_type" db:"cluster_type"`
	WalletCount   int         `json:"wallet_count" db:"wallet_count"`
	WalletIDs     []string    `json:"wallet_addresses" db:"wallet_addresses"`
	CommonOwner   bool        `json:"common_owner" db:"common_owner"`
	RiskIndicators []string   `json:"risk_indicators" db:"risk_indicators"`
	TotalVolume   float64     `json:"total_volume" db:"total_volume"`
	FirstActivity time.Time   `json:"first_activity" db:"first_activity"`
	LastActivity  time.Time   `json:"last_activity" db:"last_activity"`
	Labels        []string    `json:"labels,omitempty" db:"labels"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
}

// ReportType represents the type of forensic report
type ReportType string

const (
	ReportTypeInvestigation   ReportType = "INVESTIGATION"
	ReportTypeTimeline        ReportType = "TIMELINE"
	ReportTypeAssetTrace      ReportType = "ASSET_TRACE"
	ReportTypeWalletCluster   ReportType = "WALLET_CLUSTER"
	ReportTypeCompliance      ReportType = "COMPLIANCE"
	ReportTypeExecutive       ReportType = "EXECUTIVE"
)

// ReportStatus represents the status of a report
type ReportStatus string

const (
	ReportStatusDraft      ReportStatus = "DRAFT"
	ReportStatusReview     ReportStatus = "IN_REVIEW"
	ReportStatusApproved   ReportStatus = "APPROVED"
	ReportStatusFinal      ReportStatus = "FINAL"
	ReportStatusArchived   ReportStatus = "ARCHIVED"
)

// ForensicReport represents a forensic report
type ForensicReport struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	CaseID        uuid.UUID     `json:"case_id" db:"case_id"`
	ReportType    ReportType    `json:"report_type" db:"report_type"`
	Title         string        `json:"title" db:"title"`
	Version       int           `json:"version" db:"version"`
	Status        ReportStatus  `json:"status" db:"status"`
	Summary       string        `json:"summary" db:"summary"`
	Content       string        `json:"content" db:"content"`
	Attachments   []string      `json:"attachments,omitempty" db:"attachments"`
	EvidenceRefs  []uuid.UUID   `json:"evidence_refs" db:"evidence_refs"`
	GraphRefs     []uuid.UUID   `json:"graph_refs" db:"graph_refs"`
	Conclusions   string        `json:"conclusions" db:"conclusions"`
	Recommendations string      `json:"recommendations" db:"recommendations"`
	ReportHash    string        `json:"report_hash" db:"report_hash"`
	GeneratedBy   uuid.UUID     `json:"generated_by" db:"generated_by"`
	GeneratedByName string      `json:"generated_by_name" db:"generated_by_name"`
	ApprovedBy    *uuid.UUID    `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt    *time.Time    `json:"approved_at,omitempty" db:"approved_at"`
	GeneratedAt   time.Time     `json:"generated_at" db:"generated_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

// GraphTraceRequest represents a request to trace a transaction graph
type GraphTraceRequest struct {
	Address       string `json:"address" db:"address"`
	StartTxHash   string `json:"start_tx_hash,omitempty"`
	MaxDepth      int    `json:"max_depth"`
	MaxNodes      int    `json:"max_nodes"`
	IncludeLabels bool   `json:"include_labels"`
	FilterByValue bool   `json:"filter_by_value"`
	MinValue      float64 `json:"min_value"`
}

// GraphTraceResult represents the result of a graph trace
type GraphTraceResult struct {
	RootAddress   string           `json:"root_address"`
	Graph         *TransactionGraph `json:"graph"`
	Paths         []GraphPath      `json:"paths"`
	TimeTaken     time.Duration    `json:"time_taken"`
	NodesVisited  int              `json:"nodes_visited"`
	EdgesVisited  int              `json:"edges_visited"`
	TracedAt      time.Time        `json:"traced_at"`
}

// GraphPath represents a path through the transaction graph
type GraphPath struct {
	PathID        uuid.UUID   `json:"path_id"`
	Path          []string    `json:"path"`
	PathLength    int         `json:"path_length"`
	TotalValue    float64     `json:"total_value"`
	AssetType     string      `json:"asset_type"`
	StartTime     time.Time   `json:"start_time"`
	EndTime       time.Time   `json:"end_time"`
	ContainsKnown bool        `json:"contains_known_entity"`
	KnownEntities []string    `json:"known_entities"`
}

// TimelineEvent represents an event in a forensic timeline
type TimelineEvent struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	CaseID        uuid.UUID     `json:"case_id" db:"case_id"`
	EventType     string        `json:"event_type" db:"event_type"`
	Timestamp     time.Time     `json:"timestamp" db:"timestamp"`
	Description   string        `json:"description" db:"description"`
	EvidenceIDs   []uuid.UUID   `json:"evidence_ids" db:"evidence_ids"`
	TransactionID *uuid.UUID    `json:"transaction_id,omitempty" db:"transaction_id"`
	WalletAddress *string       `json:"wallet_address,omitempty" db:"wallet_address"`
	Metadata      string        `json:"metadata,omitempty" db:"metadata"`
	CreatedBy     uuid.UUID     `json:"created_by" db:"created_by"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
}

// CaseStats represents statistics for a case
type CaseStats struct {
	TotalCases       int64   `json:"total_cases"`
	OpenCases        int64   `json:"open_cases"`
	ResolvedCases    int64   `json:"resolved_cases"`
	AverageResolution float64 `json:"average_resolution_hours"`
	TotalEvidence    int64   `json:"total_evidence"`
	ReportsGenerated int64   `json:"reports_generated"`
}
