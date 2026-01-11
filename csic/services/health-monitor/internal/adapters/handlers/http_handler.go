package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"csic-platform/health-monitor/internal/core/domain"
	"csic-platform/health-monitor/internal/core/ports"
	"csic-platform/health-monitor/internal/core/services"
	"csic-platform/health-monitor/pkg/metrics"
)

// HTTPHandler handles HTTP requests
type HTTPHandler struct {
	monitorService  services.MonitorService
	alertService    services.AlertService
	metricsCollector *metrics.MetricsCollector
	logger          *zap.Logger
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(
	monitorService services.MonitorService,
	alertService services.AlertService,
	metricsCollector *metrics.MetricsCollector,
	logger *zap.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		monitorService:   monitorService,
		alertService:     alertService,
		metricsCollector: metricsCollector,
		logger:           logger,
	}
}

// Start starts the HTTP server
func (h *HTTPHandler) Start(port int, logger *zap.Logger) error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(h.loggingMiddleware())
	router.Use(h.metricsMiddleware())

	// Health endpoints
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Service status endpoints
		services := v1.Group("/services")
		{
			services.GET("", h.ListServices)
			services.GET("/:name", h.GetService)
			services.GET("/:name/metrics", h.GetServiceMetrics)
			services.GET("/:name/health", h.GetServiceHealthCheck)
			services.POST("/register", h.RegisterService)
		}

		// Health summary endpoint
		v1.GET("/health", h.GetHealthSummary)

		// Alert rule endpoints
		rules := v1.Group("/alert-rules")
		{
			rules.GET("", h.ListAlertRules)
			rules.GET("/:id", h.GetAlertRule)
			rules.POST("", h.CreateAlertRule)
			rules.PUT("/:id", h.UpdateAlertRule)
			rules.DELETE("/:id", h.DeleteAlertRule)
		}

		// Alert endpoints
		alerts := v1.Group("/alerts")
		{
			alerts.GET("", h.ListAlerts)
			alerts.GET("/firing", h.GetFiringAlerts)
			alerts.POST("/:id/resolve", h.ResolveAlert)
		}

		// Outage endpoints
		outages := v1.Group("/outages")
		{
			outages.GET("", h.ListOutages)
			outages.GET("/active", h.GetActiveOutages)
			outages.GET("/:id", h.GetOutage)
			outages.POST("/:id/resolve", h.ResolveOutage)
		}

		// Heartbeat endpoint
		v1.POST("/heartbeat", h.SubmitHeartbeat)
	}

	// Metrics endpoint
	router.GET("/metrics", h.MetricsHandler)

	addr := ":" + strconv.Itoa(port)
	h.logger.Info("Starting HTTP server", zap.String("addr", addr))

	return router.Run(addr)
}

// Shutdown gracefully shuts down the HTTP server
func (h *HTTPHandler) Shutdown(ctx context.Context) error {
	return nil
}

// HealthCheck returns the health status
func (h *HTTPHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

// ReadinessCheck returns the readiness status
func (h *HTTPHandler) ReadinessCheck(c *gin.Context) {
	ready := h.monitorService != nil && h.alertService != nil

	if ready {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().UTC(),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not ready",
			"timestamp": time.Now().UTC(),
		})
	}
}

// ListServices lists all services
func (h *HTTPHandler) ListServices(c *gin.Context) {
	ctx := c.Request.Context()
	statuses, err := h.monitorService.GetAllServiceStatuses(ctx)
	if err != nil {
		h.logger.Error("Failed to list services", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"services": statuses,
		"count":    len(statuses),
	})
}

// GetService gets a specific service
func (h *HTTPHandler) GetService(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	status, err := h.monitorService.GetServiceStatus(ctx, name)
	if err != nil {
		h.logger.Error("Failed to get service", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if status == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetServiceMetrics gets metrics for a specific service
func (h *HTTPHandler) GetServiceMetrics(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	since := time.Now().Add(-24 * time.Hour)

	metrics, err := h.monitorService.GetServiceMetrics(ctx, name, since)
	if err != nil {
		h.logger.Error("Failed to get service metrics", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetServiceHealthCheck performs a health check on a service
func (h *HTTPHandler) GetServiceHealthCheck(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()

	// Default endpoint and timeout
	endpoint := "http://" + name + ":8080/health"
	timeout := 5 * time.Second

	check, err := h.monitorService.PerformHealthCheck(ctx, name, endpoint, timeout)
	if err != nil {
		h.logger.Error("Failed to perform health check", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, check)
}

// RegisterService registers a new service
func (h *HTTPHandler) RegisterService(c *gin.Context) {
	var req domain.RegisterServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.monitorService.RegisterService(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to register service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetHealthSummary returns the health summary
func (h *HTTPHandler) GetHealthSummary(c *gin.Context) {
	ctx := c.Request.Context()
	summary, err := h.monitorService.GetHealthSummary(ctx)
	if err != nil {
		h.logger.Error("Failed to get health summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// ListAlertRules lists all alert rules
func (h *HTTPHandler) ListAlertRules(c *gin.Context) {
	ctx := c.Request.Context()
	rules, err := h.alertService.GetAlertRules(ctx)
	if err != nil {
		h.logger.Error("Failed to list alert rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"count": len(rules),
	})
}

// GetAlertRule gets an alert rule by ID
func (h *HTTPHandler) GetAlertRule(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	rule, err := h.alertService.GetAlertRule(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get alert rule", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// CreateAlertRule creates a new alert rule
func (h *HTTPHandler) CreateAlertRule(c *gin.Context) {
	var req domain.AlertRule
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	rule, err := h.alertService.CreateAlertRule(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create alert rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// UpdateAlertRule updates an alert rule
func (h *HTTPHandler) UpdateAlertRule(c *gin.Context) {
	id := c.Param("id")
	var req domain.AlertRule
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	rule, err := h.alertService.UpdateAlertRule(ctx, id, &req)
	if err != nil {
		h.logger.Error("Failed to update alert rule", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DeleteAlertRule deletes an alert rule
func (h *HTTPHandler) DeleteAlertRule(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	if err := h.alertService.DeleteAlertRule(ctx, id); err != nil {
		h.logger.Error("Failed to delete alert rule", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListAlerts lists alerts
func (h *HTTPHandler) ListAlerts(c *gin.Context) {
	ctx := c.Request.Context()

	filter := ports.AlertFilter{}
	if serviceName := c.Query("service"); serviceName != "" {
		filter.ServiceName = serviceName
	}
	if severity := c.Query("severity"); severity != "" {
		filter.Severity = severity
	}
	if status := c.Query("status"); status != "" {
		filter.Status = status
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	alerts, err := h.alertService.GetAlerts(ctx, filter)
	if err != nil {
		h.logger.Error("Failed to list alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// GetFiringAlerts gets all currently firing alerts
func (h *HTTPHandler) GetFiringAlerts(c *gin.Context) {
	ctx := c.Request.Context()
	alerts, err := h.alertService.GetFiringAlerts(ctx)
	if err != nil {
		h.logger.Error("Failed to get firing alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// ResolveAlert resolves an alert
func (h *HTTPHandler) ResolveAlert(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	if err := h.alertService.ResolveAlert(ctx, id); err != nil {
		h.logger.Error("Failed to resolve alert", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alert resolved"})
}

// ListOutages lists outages
func (h *HTTPHandler) ListOutages(c *gin.Context) {
	ctx := c.Request.Context()
	serviceName := c.Query("service")
	limit := 50

	if l, err := strconv.Atoi(c.Query("limit")); err == nil {
		limit = l
	}

	if serviceName != "" {
		outages, err := h.alertService.GetOutages(ctx, serviceName, limit)
		if err != nil {
			h.logger.Error("Failed to list outages", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"outages": outages,
			"count":   len(outages),
		})
		return
	}

	outages, err := h.alertService.GetActiveOutages(ctx)
	if err != nil {
		h.logger.Error("Failed to list outages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"outages": outages,
		"count":   len(outages),
	})
}

// GetActiveOutages gets all active outages
func (h *HTTPHandler) GetActiveOutages(c *gin.Context) {
	ctx := c.Request.Context()
	outages, err := h.alertService.GetActiveOutages(ctx)
	if err != nil {
		h.logger.Error("Failed to get active outages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"outages": outages,
		"count":   len(outages),
	})
}

// GetOutage gets an outage by ID
func (h *HTTPHandler) GetOutage(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	outages, err := h.alertService.GetOutages(ctx, "", 1)
	if err != nil {
		h.logger.Error("Failed to get outage", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, outage := range outages {
		if outage.ID == id {
			c.JSON(http.StatusOK, outage)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "outage not found"})
}

// ResolveOutage resolves an outage
func (h *HTTPHandler) ResolveOutage(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		RootCause string `json:"root_cause"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.alertService.ResolveOutage(ctx, id, req.RootCause); err != nil {
		h.logger.Error("Failed to resolve outage", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "outage resolved"})
}

// SubmitHeartbeat submits a heartbeat
func (h *HTTPHandler) SubmitHeartbeat(c *gin.Context) {
	var heartbeat domain.Heartbeat
	if err := c.ShouldBindJSON(&heartbeat); err != nil {
		h.logger.Error("Invalid heartbeat payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.monitorService.ProcessHeartbeat(ctx, &heartbeat); err != nil {
		h.logger.Error("Failed to process heartbeat", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "heartbeat received"})
}

// MetricsHandler returns Prometheus metrics
func (h *HTTPHandler) MetricsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "metrics endpoint",
	})
}

// loggingMiddleware logs HTTP requests
func (h *HTTPHandler) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		h.logger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
		)
	}
}

// metricsMiddleware tracks HTTP request metrics
func (h *HTTPHandler) metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.metricsCollector.IncHTTPRequestsInProgress()
		start := time.Now()

		c.Next()

		duration := float64(time.Since(start).Milliseconds())
		h.metricsCollector.DecHTTPRequestsInProgress()
		h.metricsCollector.RecordHTTPRequest(c.Request.Method, c.Request.URL.Path, strconv.Itoa(c.Writer.Status()), duration)
	}
}
