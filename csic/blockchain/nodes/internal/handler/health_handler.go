package handler

import (
	"net/http"

	"github.com/csic/platform/blockchain/nodes/internal/service"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles HTTP requests for health check operations
type HealthHandler struct {
	healthService *service.HealthService
}

// NewHealthHandler creates a new HealthHandler instance
func NewHealthHandler(healthService *service.HealthService) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

// GetHealth returns overall system health
// @Summary Get system health
// @Description Retrieve overall system health status
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/health [get]
func (h *HealthHandler) GetHealth(c *gin.Context) {
	health := h.healthService.GetHealth(c.Request.Context())
	c.JSON(http.StatusOK, health)
}

// GetNodeHealth returns the health status of a specific node
// @Summary Get node health
// @Description Retrieve health status of a specific blockchain node
// @Tags health
// @Produce json
// @Param id path string true "Node ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/health/node/{id} [get]
func (h *HealthHandler) GetNodeHealth(c *gin.Context) {
	nodeID := c.Param("id")

	result, err := h.healthService.GetNodeHealth(c.Request.Context(), nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"node_id":   result.NodeID,
		"status":    result.Status,
		"latency":   result.Latency.String(),
		"details":   result.Details,
		"checked_at": result.CheckedAt,
	})
}

// LivenessCheck returns whether the service is alive
// @Summary Liveness check
// @Description Check if the service is running
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/health/live [get]
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// ReadinessCheck returns whether the service is ready to accept traffic
// @Summary Readiness check
// @Description Check if the service is ready to accept traffic
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/v1/health/ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// In a real implementation, check database connectivity, Kafka connectivity, etc.
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
