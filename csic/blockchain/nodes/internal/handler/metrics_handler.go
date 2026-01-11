package handler

import (
	"net/http"
	"time"

	"github.com/csic/platform/blockchain/nodes/internal/domain"
	"github.com/csic/platform/blockchain/nodes/internal/service"
	"github.com/gin-gonic/gin"
)

// MetricsHandler handles HTTP requests for metrics operations
type MetricsHandler struct {
	metricsService *service.MetricsService
}

// NewMetricsHandler creates a new MetricsHandler instance
func NewMetricsHandler(metricsService *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{
		metricsService: metricsService,
	}
}

// GetNodeMetrics retrieves the latest metrics for a node
// @Summary Get node metrics
// @Description Retrieve the latest performance metrics for a blockchain node
// @Tags metrics
// @Produce json
// @Param id path string true "Node ID"
// @Success 200 {object} domain.NodeMetrics
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/metrics/node/{id} [get]
func (h *MetricsHandler) GetNodeMetrics(c *gin.Context) {
	nodeID := c.Param("id")

	metrics, err := h.metricsService.GetNodeMetrics(c.Request.Context(), nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if metrics == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no metrics available for node"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetNodeMetricsHistory retrieves historical metrics for a node
// @Summary Get node metrics history
// @Description Retrieve historical performance metrics for a blockchain node
// @Tags metrics
// @Produce json
// @Param id path string true "Node ID"
// @Param start query string false "Start time (RFC3339)"
// @Param end query string false "End time (RFC3339)"
// @Param limit query int false "Maximum number of records" default(100)
// @Success 200 {object} domain.PaginatedMetrics
// @Failure 500 {object} map[string]string
// @Router /api/v1/metrics/node/{id}/history [get]
func (h *MetricsHandler) GetNodeMetricsHistory(c *gin.Context) {
	nodeID := c.Param("id")

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()
	limit := 100

	if start := c.Query("start"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			startTime = t
		}
	}

	if end := c.Query("end"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			endTime = t
		}
	}

	metrics, err := h.metricsService.GetNodeMetricsHistory(c.Request.Context(), nodeID, startTime, endTime, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
		"total":   len(metrics),
	})
}

// GetNetworkMetrics retrieves aggregated metrics for a network
// @Summary Get network metrics
// @Description Retrieve aggregated metrics for a blockchain network
// @Tags metrics
// @Produce json
// @Param network path string true "Network name"
// @Param duration query string false "Duration (e.g., 1h, 24h, 7d)" default(24h)
// @Success 200 {object} domain.NetworkMetrics
// @Failure 500 {object} map[string]string
// @Router /api/v1/metrics/network/{network} [get]
func (h *MetricsHandler) GetNetworkMetrics(c *gin.Context) {
	network := c.Param("network")

	duration := 24 * time.Hour
	if d := c.Query("duration"); d != "" {
		if parsed, err := time.ParseDuration(d); err == nil {
			duration = parsed
		}
	}

	metrics, err := h.metricsService.GetNetworkMetrics(c.Request.Context(), network, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetSystemSummary retrieves aggregated metrics for all nodes
// @Summary Get system summary
// @Description Retrieve a summary of all nodes and networks
// @Tags metrics
// @Produce json
// @Success 200 {object} domain.SystemSummary
// @Failure 500 {object} map[string]string
// @Router /api/v1/metrics/summary [get]
func (h *MetricsHandler) GetSystemSummary(c *gin.Context) {
	summary, err := h.metricsService.GetSystemSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// PrometheusMetrics returns metrics in Prometheus format
// @Summary Prometheus metrics
// @Description Return metrics in Prometheus format
// @Tags metrics
// @Produce plain
// @Success 200 {string} string
// @Failure 500 {object} map[string]string
// @Router /internal/metrics [get]
func (h *MetricsHandler) PrometheusMetrics(c *gin.Context) {
	metrics, err := h.metricsService.PrometheusMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(metrics))
}
