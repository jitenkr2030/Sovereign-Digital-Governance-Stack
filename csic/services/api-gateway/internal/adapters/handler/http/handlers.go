package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/api-gateway/gateway/internal/core/domain"
	"github.com/api-gateway/gateway/internal/core/ports"
	"github.com/api-gateway/gateway/internal/core/services"
	"github.com/gin-gonic/gin"
)

// GatewayHandler handles HTTP requests for gateway operations.
type GatewayHandler struct {
	gatewayService *services.GatewayService
	authService    *services.AuthService
}

// NewGatewayHandler creates a new GatewayHandler.
func NewGatewayHandler(gatewayService *services.GatewayService, authService *services.AuthService) *GatewayHandler {
	return &GatewayHandler{
		gatewayService: gatewayService,
		authService:    authService,
	}
}

// RegisterRoutes registers all gateway routes.
func (h *GatewayHandler) RegisterRoutes(router *gin.Engine) {
	// Health check endpoints
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Route management
		v1.POST("/routes", h.CreateRoute)
		v1.GET("/routes", h.ListRoutes)
		v1.GET("/routes/:id", h.GetRoute)
		v1.PUT("/routes/:id", h.UpdateRoute)
		v1.DELETE("/routes/:id", h.DeleteRoute)

		// Consumer management
		v1.POST("/consumers", h.CreateConsumer)
		v1.GET("/consumers", h.ListConsumers)
		v1.GET("/consumers/:id", h.GetConsumer)
		v1.POST("/consumers/:id/apikeys", h.GenerateAPIKey)

		// Service management
		v1.POST("/services", h.CreateService)
		v1.GET("/services", h.ListServices)
		v1.GET("/services/:id", h.GetService)

		// Transform rules
		v1.POST("/routes/:id/transforms", h.CreateTransformRule)
		v1.GET("/routes/:id/transforms", h.GetTransformRules)

		// Analytics
		v1.GET("/analytics", h.GetAnalytics)
		v1.GET("/analytics/summary", h.GetAnalyticsSummary)
	}

	// Proxy routes (these would typically be handled by a proxy handler)
	router.Any("/proxy/*path", h.ProxyRequest)
}

// HealthCheck handles GET /health
func (h *GatewayHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "api-gateway",
	})
}

// ReadinessCheck handles GET /ready
func (h *GatewayHandler) ReadinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "api-gateway",
	})
}

// CreateRouteRequest represents the request body for creating a route.
type CreateRouteRequest struct {
	Name         string   `json:"name" binding:"required"`
	Path         string   `json:"path" binding:"required"`
	Methods      []string `json:"methods" binding:"required,min=1"`
	UpstreamURL  string   `json:"upstream_url" binding:"required"`
	UpstreamPath string   `json:"upstream_path"`
	Timeout      int      `json:"timeout"`
	RetryCount   int      `json:"retry_count"`
}

// CreateRoute handles POST /api/v1/routes
func (h *GatewayHandler) CreateRoute(c *gin.Context) {
	var req CreateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 30
	}

	retryCount := req.RetryCount
	if retryCount <= 0 {
		retryCount = 3
	}

	route, err := h.gatewayService.CreateRoute(
		c.Request.Context(),
		req.Name,
		req.Path,
		req.Methods,
		req.UpstreamURL,
		req.UpstreamPath,
		timeout,
		retryCount,
	)

	if err != nil {
		switch err {
		case services.ErrRouteAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case services.ErrInvalidRoutePath:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, route)
}

// ListRoutes handles GET /api/v1/routes
func (h *GatewayHandler) ListRoutes(c *gin.Context) {
	routes, err := h.gatewayService.ListRoutes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   routes,
		"total":  len(routes),
	})
}

// GetRoute handles GET /api/v1/routes/:id
func (h *GatewayHandler) GetRoute(c *gin.Context) {
	id := c.Param("id")

	routes, err := h.gatewayService.ListRoutes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var route *domain.Route
	for _, r := range routes {
		if string(r.ID) == id {
			route = r
			break
		}
	}

	if route == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
		return
	}

	c.JSON(http.StatusOK, route)
}

// UpdateRoute handles PUT /api/v1/routes/:id
func (h *GatewayHandler) UpdateRoute(c *gin.Context) {
	id := c.Param("id")

	var route domain.Route
	if err := c.ShouldBindJSON(&route); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	route.ID = domain.EntityID(id)

	if err := h.gatewayService.UpdateRoute(c.Request.Context(), &route); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, route)
}

// DeleteRoute handles DELETE /api/v1/routes/:id
func (h *GatewayHandler) DeleteRoute(c *gin.Context) {
	id := c.Param("id")

	if err := h.gatewayService.DeleteRoute(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// CreateConsumerRequest represents the request body for creating a consumer.
type CreateConsumerRequest struct {
	Username string               `json:"username" binding:"required"`
	Groups   []string             `json:"groups"`
	Quota    *domain.QuotaConfig  `json:"quota"`
}

// CreateConsumer handles POST /api/v1/consumers
func (h *GatewayHandler) CreateConsumer(c *gin.Context) {
	var req CreateConsumerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	consumer, err := h.gatewayService.CreateConsumer(
		c.Request.Context(),
		req.Username,
		req.Groups,
		req.Quota,
	)

	if err != nil {
		switch err {
		case services.ErrConsumerAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, consumer)
}

// ListConsumers handles GET /api/v1/consumers
func (h *GatewayHandler) ListConsumers(c *gin.Context) {
	// This would normally call a ListConsumers method on the service
	c.JSON(http.StatusOK, gin.H{
		"data":  []interface{}{},
		"total": 0,
	})
}

// GetConsumer handles GET /api/v1/consumers/:id
func (h *GatewayHandler) GetConsumer(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// GenerateAPIKeyRequest represents the request body for generating an API key.
type GenerateAPIKeyRequest struct {
	Name      string     `json:"name" binding:"required"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// GenerateAPIKey handles POST /api/v1/consumers/:id/apikeys
func (h *GatewayHandler) GenerateAPIKey(c *gin.Context) {
	consumerID := c.Param("id")

	var req GenerateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := h.gatewayService.GenerateAPIKey(
		c.Request.Context(),
		consumerID,
		req.Name,
		req.Scopes,
		req.ExpiresAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, key)
}

// CreateServiceRequest represents the request body for creating a service.
type CreateServiceRequest struct {
	Name        string                `json:"name" binding:"required"`
	Host        string                `json:"host" binding:"required"`
	Port        int                   `json:"port" binding:"required"`
	Protocol    string                `json:"protocol" binding:"required"`
	HealthCheck *domain.HealthCheck   `json:"health_check"`
}

// CreateService handles POST /api/v1/services
func (h *GatewayHandler) CreateService(c *gin.Context) {
	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service, err := h.gatewayService.CreateService(
		c.Request.Context(),
		req.Name,
		req.Host,
		req.Port,
		req.Protocol,
		req.HealthCheck,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, service)
}

// ListServices handles GET /api/v1/services
func (h *GatewayHandler) ListServices(c *gin.Context) {
	services, err := h.gatewayService.ListServices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   services,
		"total":  len(services),
	})
}

// GetService handles GET /api/v1/services/:id
func (h *GatewayHandler) GetService(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// CreateTransformRuleRequest represents the request body for creating a transform rule.
type CreateTransformRuleRequest struct {
	Name   string `json:"name" binding:"required"`
	Type   string `json:"type" binding:"required"`
	Action string `json:"action" binding:"required"`
	Target string `json:"target" binding:"required"`
	Value  string `json:"value"`
}

// CreateTransformRule handles POST /api/v1/routes/:id/transforms
func (h *GatewayHandler) CreateTransformRule(c *gin.Context) {
	routeID := c.Param("id")

	var req CreateTransformRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.gatewayService.CreateTransformRule(
		c.Request.Context(),
		req.Name,
		routeID,
		req.Type,
		req.Action,
		req.Target,
		req.Value,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// GetTransformRules handles GET /api/v1/routes/:id/transforms
func (h *GatewayHandler) GetTransformRules(c *gin.Context) {
	routeID := c.Param("id")

	rules, err := h.gatewayService.GetTransformRules(c.Request.Context(), routeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  rules,
		"total": len(rules),
	})
}

// GetAnalytics handles GET /api/v1/analytics
func (h *GatewayHandler) GetAnalytics(c *gin.Context) {
	var filter ports.AnalyticsFilter

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	// This would normally query analytics data
	c.JSON(http.StatusOK, gin.H{
		"data":   []interface{}{},
		"total":  0,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetAnalyticsSummary handles GET /api/v1/analytics/summary
func (h *GatewayHandler) GetAnalyticsSummary(c *gin.Context) {
	startTime := time.Now().AddDate(0, -1, 0) // Default: 1 month ago
	endTime := time.Now().UTC()

	if startStr := c.Query("start_time"); startStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = parsed
		}
	}

	if endStr := c.Query("end_time"); endStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = parsed
		}
	}

	summary, err := h.gatewayService.GetAnalyticsSummary(c.Request.Context(), startTime, endTime, "day")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// ProxyRequest handles all proxy requests.
func (h *GatewayHandler) ProxyRequest(c *gin.Context) {
	path := c.Param("path")
	method := c.Request.Method

	// Get route matching the request
	route, err := h.gatewayService.GetRoute(c.Request.Context(), path, method)
	if err != nil {
		switch err {
		case services.ErrRouteNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
		case services.ErrRouteNotActive:
			c.JSON(http.StatusForbidden, gin.H{"error": "route is not active"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// In a real implementation, this would forward the request to the upstream service
	// For now, return a mock response
	response := gin.H{
		"message":    "Proxy request",
		"route":      route.Name,
		"upstream":   route.UpstreamURL,
		"path":       path,
		"method":     method,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// MetricsHandler returns JSON metrics for Prometheus scraping.
func (h *GatewayHandler) MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := gin.H{
			"total_requests": 0,
			"total_errors":   0,
			"uptime_seconds": 0,
		}

		c.JSON(http.StatusOK, metrics)
	}
}

// EncodeJSON encodes a value to JSON.
func EncodeJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// DecodeJSON decodes JSON to a value.
func DecodeJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
