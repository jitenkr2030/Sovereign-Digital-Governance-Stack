package http

import (
	"encoding/json"
	"net/http"

	"github.com/csic-platform/internal/core/domain"
	"github.com/csic-platform/internal/core/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

// ForensicHandler handles HTTP requests for forensic tools
type ForensicHandler struct {
	graphService    *services.GraphAnalysisService
	evidenceService *services.EvidenceManagementService
}

// NewForensicHandler creates a new forensic HTTP handler
func NewForensicHandler(
	graphService *services.GraphAnalysisService,
	evidenceService *services.EvidenceManagementService,
) *ForensicHandler {
	return &ForensicHandler{
		graphService:    graphService,
		evidenceService: evidenceService,
	}
}

// RegisterRoutes registers all forensic API routes
func (h *ForensicHandler) RegisterRoutes(r chi.Router) {
	// Enable CORS
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	r.Use(cors.Handler)

	// Graph analysis endpoints
	r.Post("/api/v1/graph/build", h.BuildGraphHandler)
	r.Post("/api/v1/graph/expand", h.ExpandGraphHandler)
	r.Post("/api/v1/graph/communities", h.DetectCommunitiesHandler)
	r.Post("/api/v1/graph/centrality", h.FindCentralNodesHandler)
	r.Post("/api/v1/graph/path", h.FindPathHandler)
	r.Get("/api/v1/graph/centrality/{nodeID}", h.CalculateCentralityHandler)
	r.Get("/api/v1/graph/clustering/{nodeID}", h.CalculateClusteringHandler)
	r.Post("/api/v1/graph/components", h.GetConnectedComponentsHandler)

	// Evidence management endpoints
	r.Post("/api/v1/evidence/transaction", h.CollectTransactionEvidenceHandler)
	r.Post("/api/v1/evidence/wallet", h.CollectWalletEvidenceHandler)
	r.Post("/api/v1/evidence/cluster", h.CollectClusterEvidenceHandler)
	r.Get("/api/v1/evidence/{id}", h.GetEvidenceHandler)
	r.Get("/api/v1/evidence/{id}/timeline", h.GetEvidenceTimelineHandler)
	r.Post("/api/v1/evidence/{id}/verify", h.VerifyEvidenceHandler)
	r.Post("/api/v1/evidence/{id}/export", h.ExportEvidenceHandler)
	r.Post("/api/v1/evidence/package", h.PackageEvidenceHandler)
	r.Post("/api/v1/evidence/case", h.CreateCaseEvidenceHandler)

	// Report endpoints
	r.Post("/api/v1/reports/legal", h.GenerateLegalReportHandler)
}

// BuildGraphHandler handles requests to build transaction graphs
func (h *ForensicHandler) BuildGraphHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TxHashes []string `json:"tx_hashes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.TxHashes) == 0 {
		http.Error(w, "At least one transaction hash is required", http.StatusBadRequest)
		return
	}

	rootNode, edges, err := h.graphService.BuildTransactionGraph(r.Context(), request.TxHashes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"root_node": rootNode,
		"edges":     edges,
		"edge_count": len(edges),
	}

	w.Header().Set("Content-Type", "application/json")
	json.New(response)
}

// ExpandEncoder(w).EncodeGraphHandler handles requests to expand graph depth
func (h *ForensicHandler) ExpandGraphHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		NodeID string `json:"node_id"`
		Depth  int    `json:"depth"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.NodeID == "" {
		http.Error(w, "Node ID is required", http.StatusBadRequest)
		return
	}

	if request.Depth <= 0 {
		request.Depth = 1
	}

	nodes, edges, err := h.graphService.ExpandGraph(r.Context(), request.NodeID, request.Depth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"nodes":      nodes,
		"edges":      edges,
		"node_count": len(nodes),
		"edge_count": len(edges),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DetectCommunitiesHandler handles community detection requests
func (h *ForensicHandler) DetectCommunitiesHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		NodeIDs []string `json:"node_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	clusters, err := h.graphService.DetectCommunities(r.Context(), request.NodeIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"clusters":      clusters,
		"cluster_count": len(clusters),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// FindCentralNodesHandler handles requests to find central nodes
func (h *ForensicHandler) FindCentralNodesHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		NodeIDs []string `json:"node_ids"`
		Limit   int      `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Limit <= 0 {
		request.Limit = 10
	}

	nodes, err := h.graphService.FindCentralNodes(r.Context(), request.NodeIDs, request.Limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"central_nodes": nodes,
		"count":         len(nodes),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// FindPathHandler handles path finding requests
func (h *ForensicHandler) FindPathHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		SourceID string `json:"source_id"`
		TargetID string `json:"target_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	nodes, edges, err := h.graphService.FindPath(r.Context(), request.SourceID, request.TargetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"nodes":      nodes,
		"edges":      edges,
		"node_count": len(nodes),
		"edge_count": len(edges),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CalculateCentralityHandler handles centrality calculation requests
func (h *ForensicHandler) CalculateCentralityHandler(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "nodeID")

	centrality, err := h.graphService.CalculateCentrality(r.Context(), nodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"node_id":    nodeID,
		"centrality": centrality,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CalculateClusteringHandler handles clustering coefficient requests
func (h *ForensicHandler) CalculateClusteringHandler(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "nodeID")

	coefficient, err := h.graphService.CalculateClusteringCoefficient(r.Context(), nodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"node_id":           nodeID,
		"clustering_coeff":  coefficient,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetConnectedComponentsHandler handles connected component requests
func (h *ForensicHandler) GetConnectedComponentsHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		NodeIDs []string `json:"node_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	components, err := h.graphService.GetConnectedComponents(r.Context(), request.NodeIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"components":      components,
		"component_count": len(components),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CollectTransactionEvidenceHandler handles transaction evidence collection
func (h *ForensicHandler) CollectTransactionEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TxHash string `json:"tx_hash"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.TxHash == "" {
		http.Error(w, "Transaction hash is required", http.StatusBadRequest)
		return
	}

	evidence, err := h.evidenceService.CollectTransactionEvidence(r.Context(), request.TxHash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evidence)
}

// CollectWalletEvidenceHandler handles wallet evidence collection
func (h *ForensicHandler) CollectWalletEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Address string `json:"address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Address == "" {
		http.Error(w, "Wallet address is required", http.StatusBadRequest)
		return
	}

	evidence, err := h.evidenceService.CollectWalletEvidence(r.Context(), request.Address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evidence)
}

// CollectClusterEvidenceHandler handles cluster evidence collection
func (h *ForensicHandler) CollectClusterEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID string `json:"cluster_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.ClusterID == "" {
		http.Error(w, "Cluster ID is required", http.StatusBadRequest)
		return
	}

	evidence, err := h.evidenceService.CollectClusterEvidence(r.Context(), request.ClusterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evidence)
}

// GetEvidenceHandler handles evidence retrieval
func (h *ForensicHandler) GetEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	evidenceID := chi.URLParam(r, "id")

	evidence, err := h.evidenceService.VerifyEvidence(r.Context(), evidenceID)
	if err != nil {
		http.Error(w, "Evidence not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evidence)
}

// GetEvidenceTimelineHandler handles evidence timeline requests
func (h *ForensicHandler) GetEvidenceTimelineHandler(w http.ResponseWriter, r *http.Request) {
	evidenceID := chi.URLParam(r, "id")

	timeline, err := h.evidenceService.GetEvidenceTimeline(r.Context(), evidenceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"evidence_id": evidenceID,
		"timeline":    timeline,
		"record_count": len(timeline),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyEvidenceHandler handles evidence verification requests
func (h *ForensicHandler) VerifyEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	evidenceID := chi.URLParam(r, "id")

	valid, err := h.evidenceService.VerifyHashIntegrity(r.Context(), evidenceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"evidence_id": evidenceID,
		"valid":       valid,
		"verified_at": domain.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExportEvidenceHandler handles evidence export requests
func (h *ForensicHandler) ExportEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	evidenceID := chi.URLParam(r, "id")

	var request struct {
		Format string `json:"format"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		request.Format = "json"
	}

	data, err := h.evidenceService.ExportEvidence(r.Context(), evidenceID, request.Format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// PackageEvidenceHandler handles legal evidence packaging
func (h *ForensicHandler) PackageEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		EvidenceIDs []string `json:"evidence_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	report, err := h.evidenceService.PackageForLegal(r.Context(), request.EvidenceIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// CreateCaseEvidenceHandler handles case evidence package creation
func (h *ForensicHandler) CreateCaseEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		CaseID     string   `json:"case_id"`
		EvidenceIDs []string `json:"evidence_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.CaseID == "" {
		http.Error(w, "Case ID is required", http.StatusBadRequest)
		return
	}

	evidence, err := h.evidenceService.CreateCaseEvidence(r.Context(), request.CaseID, request.EvidenceIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evidence)
}

// GenerateLegalReportHandler handles legal report generation
func (h *ForensicHandler) GenerateLegalReportHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		CaseID string `json:"case_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.CaseID == "" {
		http.Error(w, "Case ID is required", http.StatusBadRequest)
		return
	}

	report := &domain.Report{
		ID:          "",
		CaseID:      request.CaseID,
		ReportType:  domain.ReportTypeLegal,
		Title:       "Legal Investigation Report",
		Description: "Comprehensive legal investigation report",
		Content:     map[string]interface{}{},
		GeneratedAt: domain.Now(),
		GeneratedBy: "system",
		Status:      domain.ReportStatusGenerated,
		Version:     "1.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
