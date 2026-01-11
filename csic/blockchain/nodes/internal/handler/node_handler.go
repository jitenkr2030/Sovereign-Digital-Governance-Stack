package handler

import (
	"net/http"

	"github.com/csic/platform/blockchain/nodes/internal/domain"
	"github.com/csic/platform/blockchain/nodes/internal/service"
	"github.com/gin-gonic/gin"
)

// NodeHandler handles HTTP requests for node operations
type NodeHandler struct {
	nodeService *service.NodeService
}

// NewNodeHandler creates a new NodeHandler instance
func NewNodeHandler(nodeService *service.NodeService) *NodeHandler {
	return &NodeHandler{
		nodeService: nodeService,
	}
}

// CreateNode creates a new blockchain node
// @Summary Create a new blockchain node
// @Description Register a new blockchain node for monitoring
// @Tags nodes
// @Accept json
// @Produce json
// @Param node body domain.CreateNodeRequest true "Node configuration"
// @Success 201 {object} domain.Node
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/nodes [post]
func (h *NodeHandler) CreateNode(c *gin.Context) {
	var req domain.CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node, err := h.nodeService.CreateNode(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, node)
}

// GetNode retrieves a node by ID
// @Summary Get a blockchain node
// @Description Retrieve details of a specific blockchain node
// @Tags nodes
// @Produce json
// @Param id path string true "Node ID"
// @Success 200 {object} domain.Node
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/nodes/{id} [get]
func (h *NodeHandler) GetNode(c *gin.Context) {
	id := c.Param("id")

	node, err := h.nodeService.GetNode(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, node)
}

// ListNodes lists all nodes with optional filtering
// @Summary List blockchain nodes
// @Description Retrieve a paginated list of blockchain nodes
// @Tags nodes
// @Produce json
// @Param network query string false "Filter by network"
// @Param status query string false "Filter by status"
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(20)
// @Success 200 {object} domain.PaginatedNodes
// @Failure 500 {object} map[string]string
// @Router /api/v1/nodes [get]
func (h *NodeHandler) ListNodes(c *gin.Context) {
	filter := &domain.NodeListFilter{
		Network: c.Query("network"),
		Status:  domain.NodeStatus(c.Query("status")),
		Type:    domain.NodeType(c.Query("type")),
		Offset:  0,
		Limit:   20,
	}

	if offset := c.Query("offset"); offset != "" {
		// Parse offset
	}

	if limit := c.Query("limit"); limit != "" {
		// Parse limit
	}

	result, err := h.nodeService.ListNodes(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateNode updates an existing node
// @Summary Update a blockchain node
// @Description Update configuration of an existing blockchain node
// @Tags nodes
// @Accept json
// @Produce json
// @Param id path string true "Node ID"
// @Param node body domain.UpdateNodeRequest true "Node updates"
// @Success 200 {object} domain.Node
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/nodes/{id} [put]
func (h *NodeHandler) UpdateNode(c *gin.Context) {
	id := c.Param("id")

	var req domain.UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node, err := h.nodeService.UpdateNode(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, node)
}

// DeleteNode deletes a node
// @Summary Delete a blockchain node
// @Description Remove a blockchain node from monitoring
// @Tags nodes
// @Produce json
// @Param id path string true "Node ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/nodes/{id} [delete]
func (h *NodeHandler) DeleteNode(c *gin.Context) {
	id := c.Param("id")

	if err := h.nodeService.DeleteNode(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// RestartNode restarts a node
// @Summary Restart a blockchain node
// @Description Trigger a restart of a blockchain node
// @Tags nodes
// @Produce json
// @Param id path string true "Node ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/nodes/{id}/restart [post]
func (h *NodeHandler) RestartNode(c *gin.Context) {
	id := c.Param("id")

	if err := h.nodeService.RestartNode(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Node restart initiated"})
}

// ForceSync forces a synchronization update
// @Summary Force node sync
// @Description Trigger a synchronization update for a node
// @Tags nodes
// @Produce json
// @Param id path string true "Node ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/nodes/{id}/sync [post]
func (h *NodeHandler) ForceSync(c *gin.Context) {
	id := c.Param("id")

	if err := h.nodeService.ForceSync(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sync requested"})
}

// GetNetworkNodes retrieves all nodes for a specific network
// @Summary Get nodes by network
// @Description Retrieve all nodes for a specific blockchain network
// @Tags networks
// @Produce json
// @Param network path string true "Network name"
// @Success 200 {array} domain.Node
// @Failure 500 {object} map[string]string
// @Router /api/v1/networks/{network}/nodes [get]
func (h *NodeHandler) GetNetworkNodes(c *gin.Context) {
	network := c.Param("network")

	nodes, err := h.nodeService.GetNetworkNodes(c.Request.Context(), network)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nodes)
}

// ListNetworks lists all configured networks
// @Summary List networks
// @Description Retrieve information about all configured blockchain networks
// @Tags networks
// @Produce json
// @Success 200 {array} domain.NetworkInfo
// @Failure 500 {object} map[string]string
// @Router /api/v1/networks [get]
func (h *NodeHandler) ListNetworks(c *gin.Context) {
	networks, err := h.nodeService.ListNetworks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, networks)
}
