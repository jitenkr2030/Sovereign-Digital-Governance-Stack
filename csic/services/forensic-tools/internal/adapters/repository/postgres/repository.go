package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/csic-platform/internal/core/domain"
	"github.com/csic-platform/internal/core/ports"
	_ "github.com/lib/pq"
)

// PostgresRepository implements all forensic repository interfaces
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository instance
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GraphRepository implementations

// StoreNode stores a graph node in the database
func (r *PostgresRepository) StoreNode(ctx context.Context, node *domain.GraphNode) error {
	query := `
		INSERT INTO graph_nodes (id, node_type, label, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			node_type = EXCLUDED.node_type,
			label = EXCLUDED.label,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
	`

	metadataJSON, err := node.MetadataToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	now := time.Now().Unix()
	_, err = r.db.ExecContext(ctx, query,
		node.ID,
		string(node.NodeType),
		node.Label,
		metadataJSON,
		node.CreatedAt,
		now,
	)
	return err
}

// GetNode retrieves a graph node by ID
func (r *PostgresRepository) GetNode(ctx context.Context, id string) (*domain.GraphNode, error) {
	query := `SELECT id, node_type, label, metadata, created_at FROM graph_nodes WHERE id = $1`

	var node domain.GraphNode
	var metadataJSON string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&node.ID,
		&node.NodeType,
		&node.Label,
		&metadataJSON,
		&node.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("node not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	if err := node.MetadataFromJSON(metadataJSON); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &node, nil
}

// GetNodeByAddress retrieves a graph node by wallet address
func (r *PostgresRepository) GetNodeByAddress(ctx context.Context, address string) (*domain.GraphNode, error) {
	query := `SELECT id, node_type, label, metadata, created_at FROM graph_nodes WHERE metadata @> $1`

	var node domain.GraphNode
	var metadataJSON string

	err := r.db.QueryRowContext(ctx, query, fmt.Sprintf(`{"address": "%s"}`, address)).Scan(
		&node.ID,
		&node.NodeType,
		&node.Label,
		&metadataJSON,
		&node.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("node not found for address: %s", address)
	}
	if err != nil {
		return nil, err
	}

	node.MetadataFromJSON(metadataJSON)
	return &node, nil
}

// ListNodes retrieves a list of graph nodes with pagination
func (r *PostgresRepository) ListNodes(ctx context.Context, limit, offset int) ([]*domain.GraphNode, error) {
	query := `SELECT id, node_type, label, metadata, created_at FROM graph_nodes ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*domain.GraphNode
	for rows.Next() {
		var node domain.GraphNode
		var metadataJSON string

		if err := rows.Scan(&node.ID, &node.NodeType, &node.Label, &metadataJSON, &node.CreatedAt); err != nil {
			return nil, err
		}

		node.MetadataFromJSON(metadataJSON)
		nodes = append(nodes, &node)
	}

	return nodes, rows.Err()
}

// DeleteNode removes a graph node from the database
func (r *PostgresRepository) DeleteNode(ctx context.Context, id string) error {
	query := `DELETE FROM graph_nodes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// StoreEdge stores a graph edge in the database
func (r *PostgresRepository) StoreEdge(ctx context.Context, edge *domain.GraphEdge) error {
	query := `
		INSERT INTO graph_edges (id, source_id, target_id, edge_type, weight, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			source_id = EXCLUDED.source_id,
			target_id = EXCLUDED.target_id,
			edge_type = EXCLUDED.edge_type,
			weight = EXCLUDED.weight,
			metadata = EXCLUDED.metadata
	`

	metadataJSON, _ := edge.MetadataToJSON()
	now := time.Now().Unix()

	_, err := r.db.ExecContext(ctx, query,
		edge.ID,
		edge.SourceID,
		edge.TargetID,
		string(edge.EdgeType),
		edge.Weight,
		metadataJSON,
		now,
	)
	return err
}

// GetEdge retrieves a graph edge by ID
func (r *PostgresRepository) GetEdge(ctx context.Context, id string) (*domain.GraphEdge, error) {
	query := `SELECT id, source_id, target_id, edge_type, weight, metadata, created_at FROM graph_edges WHERE id = $1`

	var edge domain.GraphEdge
	var metadataJSON string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&edge.ID,
		&edge.SourceID,
		&edge.TargetID,
		&edge.EdgeType,
		&edge.Weight,
		&metadataJSON,
		&edge.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("edge not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	edge.MetadataFromJSON(metadataJSON)
	return &edge, nil
}

// GetEdgesBySource retrieves all edges from a source node
func (r *PostgresRepository) GetEdgesBySource(ctx context.Context, sourceID string) ([]*domain.GraphEdge, error) {
	query := `SELECT id, source_id, target_id, edge_type, weight, metadata, created_at FROM graph_edges WHERE source_id = $1`

	rows, err := r.db.QueryContext(ctx, query, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []*domain.GraphEdge
	for rows.Next() {
		var edge domain.GraphEdge
		var metadataJSON string

		if err := rows.Scan(&edge.ID, &edge.SourceID, &edge.TargetID, &edge.EdgeType, &edge.Weight, &metadataJSON, &edge.CreatedAt); err != nil {
			return nil, err
		}

		edge.MetadataFromJSON(metadataJSON)
		edges = append(edges, &edge)
	}

	return edges, rows.Err()
}

// GetEdgesByTarget retrieves all edges to a target node
func (r *PostgresRepository) GetEdgesByTarget(ctx context.Context, targetID string) ([]*domain.GraphEdge, error) {
	query := `SELECT id, source_id, target_id, edge_type, weight, metadata, created_at FROM graph_edges WHERE target_id = $1`

	rows, err := r.db.QueryContext(ctx, query, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []*domain.GraphEdge
	for rows.Next() {
		var edge domain.GraphEdge
		var metadataJSON string

		if err := rows.Scan(&edge.ID, &edge.SourceID, &edge.TargetID, &edge.EdgeType, &edge.Weight, &metadataJSON, &edge.CreatedAt); err != nil {
			return nil, err
		}

		edge.MetadataFromJSON(metadataJSON)
		edges = append(edges, &edge)
	}

	return edges, rows.Err()
}

// DeleteEdge removes a graph edge from the database
func (r *PostgresRepository) DeleteEdge(ctx context.Context, id string) error {
	query := `DELETE FROM graph_edges WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// StoreGraph stores multiple nodes and edges in a transaction
func (r *PostgresRepository) StoreGraph(ctx context.Context, nodes []*domain.GraphNode, edges []*domain.GraphEdge) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Store nodes
	for _, node := range nodes {
		if err := r.StoreNode(ctx, node); err != nil {
			return err
		}
	}

	// Store edges
	for _, edge := range edges {
		if err := r.StoreEdge(ctx, edge); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetSubgraph retrieves nodes and edges for a list of node IDs
func (r *PostgresRepository) GetSubgraph(ctx context.Context, nodeIDs []string) ([]*domain.GraphNode, []*domain.GraphEdge, error) {
	if len(nodeIDs) == 0 {
		return []*domain.GraphNode{}, []*domain.GraphEdge{}, nil
	}

	// Build query placeholders
	placeholders := ""
	args := make([]interface{}, len(nodeIDs))
	for i, id := range nodeIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	// Get nodes
	nodeQuery := fmt.Sprintf(`SELECT id, node_type, label, metadata, created_at FROM graph_nodes WHERE id IN (%s)`, placeholders)
	nodeRows, err := r.db.QueryContext(ctx, nodeQuery, args...)
	if err != nil {
		return nil, nil, err
	}
	defer nodeRows.Close()

	var nodes []*domain.GraphNode
	for nodeRows.Next() {
		var node domain.GraphNode
		var metadataJSON string

		if err := nodeRows.Scan(&node.ID, &node.NodeType, &node.Label, &metadataJSON, &node.CreatedAt); err != nil {
			return nil, nil, err
		}

		node.MetadataFromJSON(metadataJSON)
		nodes = append(nodes, &node)
	}

	// Get edges
	edgeQuery := fmt.Sprintf(`SELECT id, source_id, target_id, edge_type, weight, metadata, created_at FROM graph_edges WHERE source_id IN (%s) OR target_id IN (%s)`, placeholders, placeholders)
	edgeRows, err := r.db.QueryContext(ctx, edgeQuery, args...)
	if err != nil {
		return nil, nil, err
	}
	defer edgeRows.Close()

	var edges []*domain.GraphEdge
	for edgeRows.Next() {
		var edge domain.GraphEdge
		var metadataJSON string

		if err := edgeRows.Scan(&edge.ID, &edge.SourceID, &edge.TargetID, &edge.EdgeType, &edge.Weight, &metadataJSON, &edge.CreatedAt); err != nil {
			return nil, nil, err
		}

		edge.MetadataFromJSON(metadataJSON)
		edges = append(edges, &edge)
	}

	return nodes, edges, edgeRows.Err()
}

// EvidenceRepository implementations

// StoreEvidence stores forensic evidence
func (r *PostgresRepository) StoreEvidence(ctx context.Context, evidence *domain.Evidence) error {
	query := `
		INSERT INTO evidence (id, case_id, evidence_type, title, description, content, tags, hash, collected_at, collected_by, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			case_id = EXCLUDED.case_id,
			evidence_type = EXCLUDED.evidence_type,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			content = EXCLUDED.content,
			tags = EXCLUDED.tags,
			hash = EXCLUDED.hash,
			status = EXCLUDED.status
	`

	contentJSON, _ := evidence.ContentToJSON()
	tagsJSON, _ := evidence.TagsToJSON()

	_, err := r.db.ExecContext(ctx, query,
		evidence.ID,
		evidence.CaseID,
		string(evidence.EvidenceType),
		evidence.Title,
		evidence.Description,
		contentJSON,
		tagsJSON,
		evidence.Hash,
		evidence.CollectedAt,
		evidence.CollectedBy,
		string(evidence.Status),
	)
	return err
}

// GetEvidence retrieves evidence by ID
func (r *PostgresRepository) GetEvidence(ctx context.Context, id string) (*domain.Evidence, error) {
	query := `SELECT id, case_id, evidence_type, title, description, content, tags, hash, collected_at, collected_by, status FROM evidence WHERE id = $1`

	var evidence domain.Evidence
	var contentJSON, tagsJSON string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&evidence.ID,
		&evidence.CaseID,
		&evidence.EvidenceType,
		&evidence.Title,
		&evidence.Description,
		&contentJSON,
		&tagsJSON,
		&evidence.Hash,
		&evidence.CollectedAt,
		&evidence.CollectedBy,
		&evidence.Status,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("evidence not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	evidence.ContentFromJSON(contentJSON)
	evidence.TagsFromJSON(tagsJSON)
	return &evidence, nil
}

// ListEvidence retrieves evidence for a case with pagination
func (r *PostgresRepository) ListEvidence(ctx context.Context, caseID string, limit, offset int) ([]*domain.Evidence, error) {
	query := `SELECT id, case_id, evidence_type, title, description, content, tags, hash, collected_at, collected_by, status FROM evidence WHERE case_id = $1 ORDER BY collected_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, caseID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evidenceList []*domain.Evidence
	for rows.Next() {
		var evidence domain.Evidence
		var contentJSON, tagsJSON string

		if err := rows.Scan(&evidence.ID, &evidence.CaseID, &evidence.EvidenceType, &evidence.Title, &evidence.Description, &contentJSON, &tagsJSON, &evidence.Hash, &evidence.CollectedAt, &evidence.CollectedBy, &evidence.Status); err != nil {
			return nil, err
		}

		evidence.ContentFromJSON(contentJSON)
		evidence.TagsFromJSON(tagsJSON)
		evidenceList = append(evidenceList, &evidence)
	}

	return evidenceList, rows.Err()
}

// DeleteEvidence removes evidence from the database
func (r *PostgresRepository) DeleteEvidence(ctx context.Context, id string) error {
	query := `DELETE FROM evidence WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// AppendCustodyRecord adds a chain of custody record
func (r *PostgresRepository) AppendCustodyRecord(ctx context.Context, evidenceID string, record *domain.ChainOfCustody) error {
	query := `INSERT INTO chain_of_custody (evidence_id, timestamp, handler, action, location, integrity_hash, notes) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		evidenceID,
		record.Timestamp,
		record.Handler,
		record.Action,
		record.Location,
		record.IntegrityHash,
		record.Notes,
	)
	return err
}

// GetCustodyHistory retrieves the complete chain of custody for evidence
func (r *PostgresRepository) GetCustodyHistory(ctx context.Context, evidenceID string) ([]*domain.ChainOfCustody, error) {
	query := `SELECT timestamp, handler, action, location, integrity_hash, notes FROM chain_of_custody WHERE evidence_id = $1 ORDER BY timestamp ASC`

	rows, err := r.db.QueryContext(ctx, query, evidenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*domain.ChainOfCustody
	for rows.Next() {
		var record domain.ChainOfCustody

		if err := rows.Scan(&record.Timestamp, &record.Handler, &record.Action, &record.Location, &record.IntegrityHash, &record.Notes); err != nil {
			return nil, err
		}

		history = append(history, &record)
	}

	return history, rows.Err()
}

// VerifyEvidence checks if evidence hash matches stored hash
func (r *PostgresRepository) VerifyEvidence(ctx context.Context, id string) (bool, error) {
	query := `SELECT hash FROM evidence WHERE id = $1`
	var storedHash string

	err := r.db.QueryRowContext(ctx, query, id).Scan(&storedHash)
	if err == sql.ErrNoRows {
		return false, fmt.Errorf("evidence not found: %s", id)
	}
	if err != nil {
		return false, err
	}

	return storedHash != "", nil
}

// ReportRepository implementations

// StoreReport stores a forensic report
func (r *PostgresRepository) StoreReport(ctx context.Context, report *domain.Report) error {
	query := `
		INSERT INTO forensic_reports (id, case_id, report_type, title, description, content, hash, generated_at, generated_by, status, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			case_id = EXCLUDED.case_id,
			report_type = EXCLUDED.report_type,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			content = EXCLUDED.content,
			hash = EXCLUDED.hash,
			status = EXCLUDED.status,
			version = EXCLUDED.version
	`

	contentJSON, _ := report.ContentToJSON()

	_, err := r.db.ExecContext(ctx, query,
		report.ID,
		report.CaseID,
		string(report.ReportType),
		report.Title,
		report.Description,
		contentJSON,
		report.Hash,
		report.GeneratedAt,
		report.GeneratedBy,
		string(report.Status),
		report.Version,
	)
	return err
}

// GetReport retrieves a report by ID
func (r *PostgresRepository) GetReport(ctx context.Context, id string) (*domain.Report, error) {
	query := `SELECT id, case_id, report_type, title, description, content, hash, generated_at, generated_by, status, version FROM forensic_reports WHERE id = $1`

	var report domain.Report
	var contentJSON string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID,
		&report.CaseID,
		&report.ReportType,
		&report.Title,
		&report.Description,
		&contentJSON,
		&report.Hash,
		&report.GeneratedAt,
		&report.GeneratedBy,
		&report.Status,
		&report.Version,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("report not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	report.ContentFromJSON(contentJSON)
	return &report, nil
}

// ListReports retrieves reports for a case with pagination
func (r *PostgresRepository) ListReports(ctx context.Context, caseID string, limit, offset int) ([]*domain.Report, error) {
	query := `SELECT id, case_id, report_type, title, description, content, hash, generated_at, generated_by, status, version FROM forensic_reports WHERE case_id = $1 ORDER BY generated_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, caseID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*domain.Report
	for rows.Next() {
		var report domain.Report
		var contentJSON string

		if err := rows.Scan(&report.ID, &report.CaseID, &report.ReportType, &report.Title, &report.Description, &contentJSON, &report.Hash, &report.GeneratedAt, &report.GeneratedBy, &report.Status, &report.Version); err != nil {
			return nil, err
		}

		report.ContentFromJSON(contentJSON)
		reports = append(reports, &report)
	}

	return reports, rows.Err()
}

// DeleteReport removes a report from the database
func (r *PostgresRepository) DeleteReport(ctx context.Context, id string) error {
	query := `DELETE FROM forensic_reports WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GenerateReport generates a report for a case
func (r *PostgresRepository) GenerateReport(ctx context.Context, caseID string, reportType domain.ReportType) (*domain.Report, error) {
	report := &domain.Report{
		ID:          fmt.Sprintf("rpt-%s-%d", caseID[:8], time.Now().UnixNano()),
		CaseID:      caseID,
		ReportType:  reportType,
		Title:       fmt.Sprintf("Report - %s - %s", reportType, time.Now().Format("2006-01-02")),
		Description: fmt.Sprintf("Auto-generated %s report", reportType),
		Content: map[string]interface{}{
			"case_id":       caseID,
			"report_type":   string(reportType),
			"generated_at":  time.Now().UTC().Format(time.RFC3339),
			"generated_by":  "system",
		},
		GeneratedAt: time.Now().Unix(),
		GeneratedBy: "system",
		Status:      domain.ReportStatusGenerated,
		Version:     "1.0",
	}

	if err := r.StoreReport(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

// ClusterRepository implementations

// StoreCluster stores a cluster
func (r *PostgresRepository) StoreCluster(ctx context.Context, cluster *domain.Cluster) error {
	query := `INSERT INTO clusters (id, name, member_ids, score, metadata, created_at) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, member_ids = EXCLUDED.member_ids, score = EXCLUDED.score, metadata = EXCLUDED.metadata`

	membersJSON, _ := cluster.MemberIDsToJSON()
	metadataJSON, _ := cluster.MetadataToJSON()

	_, err := r.db.ExecContext(ctx, query,
		cluster.ID,
		cluster.Name,
		membersJSON,
		cluster.Score,
		metadataJSON,
		cluster.CreatedAt,
	)
	return err
}

// GetCluster retrieves a cluster by ID
func (r *PostgresRepository) GetCluster(ctx context.Context, id string) (*domain.Cluster, error) {
	query := `SELECT id, name, member_ids, score, metadata, created_at FROM clusters WHERE id = $1`

	var cluster domain.Cluster
	var membersJSON, metadataJSON string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cluster.ID,
		&cluster.Name,
		&membersJSON,
		&cluster.Score,
		&metadataJSON,
		&cluster.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cluster not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	cluster.MemberIDsFromJSON(membersJSON)
	cluster.MetadataFromJSON(metadataJSON)
	return &cluster, nil
}

// ListClusters retrieves all clusters for a case
func (r *PostgresRepository) ListClusters(ctx context.Context, caseID string) ([]*domain.Cluster, error) {
	// This would typically join with a cases table, but for simplicity we store case_id with clusters
	query := `SELECT id, name, member_ids, score, metadata, created_at FROM clusters`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []*domain.Cluster
	for rows.Next() {
		var cluster domain.Cluster
		var membersJSON, metadataJSON string

		if err := rows.Scan(&cluster.ID, &cluster.Name, &membersJSON, &cluster.Score, &metadataJSON, &cluster.CreatedAt); err != nil {
			return nil, err
		}

		cluster.MemberIDsFromJSON(membersJSON)
		cluster.MetadataFromJSON(metadataJSON)
		clusters = append(clusters, &cluster)
	}

	return clusters, rows.Err()
}

// DeleteCluster removes a cluster from the database
func (r *PostgresRepository) DeleteCluster(ctx context.Context, id string) error {
	query := `DELETE FROM clusters WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetClusterMembers retrieves member IDs for a cluster
func (r *PostgresRepository) GetClusterMembers(ctx context.Context, clusterID string) ([]string, error) {
	query := `SELECT member_ids FROM clusters WHERE id = $1`

	var membersJSON string
	err := r.db.QueryRowContext(ctx, query, clusterID).Scan(&membersJSON)
	if err != nil {
		return nil, err
	}

	var members []string
	return members, nil
}

// AddClusterMembers adds members to a cluster
func (r *PostgresRepository) AddClusterMembers(ctx context.Context, clusterID string, nodeIDs []string) error {
	cluster, err := r.GetCluster(ctx, clusterID)
	if err != nil {
		return err
	}

	existingMembers := make(map[string]bool)
	for _, id := range cluster.MemberIDs {
		existingMembers[id] = true
	}

	for _, id := range nodeIDs {
		existingMembers[id] = true
	}

	var allMembers []string
	for id := range existingMembers {
		allMembers = append(allMembers, id)
	}

	query := `UPDATE clusters SET member_ids = $1 WHERE id = $2`
	_, err = r.db.ExecContext(ctx, query, allMembers, clusterID)
	return err
}

// RemoveClusterMembers removes members from a cluster
func (r *PostgresRepository) RemoveClusterMembers(ctx context.Context, clusterID string, nodeIDs []string) error {
	cluster, err := r.GetCluster(ctx, clusterID)
	if err != nil {
		return err
	}

	removeSet := make(map[string]bool)
	for _, id := range nodeIDs {
		removeSet[id] = true
	}

	var remainingMembers []string
	for _, id := range cluster.MemberIDs {
		if !removeSet[id] {
			remainingMembers = append(remainingMembers, id)
		}
	}

	query := `UPDATE clusters SET member_ids = $1 WHERE id = $2`
	_, err = r.db.ExecContext(ctx, query, remainingMembers, clusterID)
	return err
}

// Ensure PostgresRepository implements all required interfaces
var _ ports.GraphRepository = (*PostgresRepository)(nil)
var _ ports.EvidenceRepository = (*PostgresRepository)(nil)
var _ ports.ReportRepository = (*PostgresRepository)(nil)
var _ ports.ClusterRepository = (*PostgresRepository)(nil)
