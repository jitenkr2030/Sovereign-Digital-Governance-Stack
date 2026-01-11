package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"csic-platform/control-layer/internal/core/domain"
	"csic-platform/control-layer/internal/core/ports"
	"csic-platform/control-layer/internal/core/services"
	"csic-platform/control-layer/pkg/metrics"
)

// HTTPHandler handles HTTP requests
type HTTPHandler struct {
	policyEngine        services.PolicyEngine
	enforcementHandler  services.EnforcementHandler
	stateRegistry       services.StateRegistry
	interventionService services.InterventionService
	metricsCollector    *metrics.MetricsCollector
	logger              *zap.Logger
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(
	policyEngine services.PolicyEngine,
	enforcementHandler services.EnforcementHandler,
	stateRegistry services.StateRegistry,
	interventionService services.InterventionService,
	metricsCollector *metrics.MetricsCollector,
	logger *zap.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		policyEngine:        policyEngine,
		enforcementHandler:  enforcementHandler,
		stateRegistry:       stateRegistry,
		interventionService: interventionService,
		metricsCollector:    metricsCollector,
		logger:              logger,
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
		// Policy endpoints
		policies := v1.Group("/policies")
		{
			policies.GET("", h.ListPolicies)
			policies.GET("/:id", h.GetPolicy)
			policies.POST("", h.CreatePolicy)
			policies.PUT("/:id", h.UpdatePolicy)
			policies.DELETE("/:id", h.DeletePolicy)
		}

		// Enforcement endpoints
		enforcements := v1.Group("/enforcements")
		{
			enforcements.GET("", h.ListEnforcements)
			enforcements.GET("/:id", h.GetEnforcement)
			enforcements.POST("", h.CreateEnforcement)
			enforcements.PUT("/:id/status", h.UpdateEnforcementStatus)
		}

		// Intervention endpoints
		interventions := v1.Group("/interventions")
		{
			interventions.GET("", h.ListInterventions)
			interventions.GET("/:id", h.GetIntervention)
			interventions.POST("", h.CreateIntervention)
			interventions.PUT("/:id/status", h.UpdateInterventionStatus)
			interventions.POST("/:id/resolve", h.ResolveIntervention)
		}

		// State endpoints
		states := v1.Group("/states")
		{
			states.GET("", h.ListStates)
			states.GET("/:key", h.GetState)
			states.PUT("/:key", h.UpdateState)
			states.DELETE("/:key", h.DeleteState)
		}

		// Evaluation endpoints
		evaluate := v1.Group("/evaluate")
		{
			evaluate.POST("", h.EvaluatePolicy)
		}
	}

	// Metrics endpoint
	router.GET("/metrics", h.MetricsHandler)

	addr := ":" + strconv.Itoa(port)
	h.logger.Info("Starting HTTP server", zap.String("addr", addr))

	return router.Run(addr)
}

// Shutdown gracefully shuts down the HTTP server
func (h *HTTPHandler) Shutdown(ctx context.Context) error {
	// In production, use http.Server.Shutdown()
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
	// Check if all dependencies are ready
	ready := h.policyEngine.IsReady()

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

// ListPolicies lists all policies
func (h *HTTPHandler) ListPolicies(c *gin.Context) {
	ctx := c.Request.Context()
	policies, err := h.policyEngine.ListPolicies(ctx)
	if err != nil {
		h.logger.Error("Failed to list policies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"policies": policies,
		"count":    len(policies),
	})
}

// GetPolicy gets a policy by ID
func (h *HTTPHandler) GetPolicy(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	policy, err := h.policyEngine.GetPolicy(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get policy", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if policy == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// CreatePolicy creates a new policy
func (h *HTTPHandler) CreatePolicy(c *gin.Context) {
	var req domain.CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	policy, err := h.policyEngine.CreatePolicy(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, policy)
}

// UpdatePolicy updates a policy
func (h *HTTPHandler) UpdatePolicy(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	policy, err := h.policyEngine.UpdatePolicy(ctx, id, &req)
	if err != nil {
		h.logger.Error("Failed to update policy", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// DeletePolicy deletes a policy
func (h *HTTPHandler) DeletePolicy(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	if err := h.policyEngine.DeletePolicy(ctx, id); err != nil {
		h.logger.Error("Failed to delete policy", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListEnforcements lists all enforcements
func (h *HTTPHandler) ListEnforcements(c *gin.Context) {
	ctx := c.Request.Context()
	enforcements, err := h.enforcementHandler.ListEnforcements(ctx)
	if err != nil {
		h.logger.Error("Failed to list enforcements", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enforcements": enforcements,
		"count":        len(enforcements),
	})
}

// GetEnforcement gets an enforcement by ID
func (h *HTTPHandler) GetEnforcement(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	enforcement, err := h.enforcementHandler.GetEnforcement(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get enforcement", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if enforcement == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "enforcement not found"})
		return
	}

	c.JSON(http.StatusOK, enforcement)
}

// CreateEnforcement creates a new enforcement
func (h *HTTPHandler) CreateEnforcement(c *gin.Context) {
	var req domain.CreateEnforcementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	enforcement, err := h.enforcementHandler.CreateEnforcement(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create enforcement", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, enforcement)
}

// UpdateEnforcementStatus updates the status of an enforcement
func (h *HTTPHandler) UpdateEnforcementStatus(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateEnforcementStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.enforcementHandler.UpdateStatus(ctx, id, req.Status); err != nil {
		h.logger.Error("Failed to update enforcement status", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

// ListInterventions lists all interventions
func (h *HTTPHandler) ListInterventions(c *gin.Context) {
	ctx := c.Request.Context()
	interventions, err := h.interventionService.ListInterventions(ctx)
	if err != nil {
		h.logger.Error("Failed to list interventions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"interventions": interventions,
		"count":         len(interventions),
	})
}

// GetIntervention gets an intervention by ID
func (h *HTTPHandler) GetIntervention(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	intervention, err := h.interventionService.GetIntervention(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get intervention", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if intervention == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "intervention not found"})
		return
	}

	c.JSON(http.StatusOK, intervention)
}

// CreateIntervention creates a new intervention
func (h *HTTPHandler) CreateIntervention(c *gin.Context) {
	var req domain.CreateInterventionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	intervention, err := h.interventionService.CreateIntervention(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create intervention", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, intervention)
}

// UpdateInterventionStatus updates the status of an intervention
func (h *HTTPHandler) UpdateInterventionStatus(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateInterventionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.interventionService.UpdateStatus(ctx, id, req.Status); err != nil {
		h.logger.Error("Failed to update intervention status", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

// ResolveIntervention resolves an intervention
func (h *HTTPHandler) ResolveIntervention(c *gin.Context) {
	id := c.Param("id")
	var req domain.ResolveInterventionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.interventionService.Resolve(ctx, id, req.Resolution); err != nil {
		h.logger.Error("Failed to resolve intervention", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "intervention resolved"})
}

// ListStates lists all states
func (h *HTTPHandler) ListStates(c *gin.Context) {
	ctx := c.Request.Context()
	states, err := h.stateRegistry.ListStates(ctx)
	if err != nil {
		h.logger.Error("Failed to list states", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"states": states,
		"count":  len(states),
	})
}

// GetState gets a state by key
func (h *HTTPHandler) GetState(c *gin.Context) {
	key := c.Param("key")
	ctx := c.Request.Context()
	state, err := h.stateRegistry.GetState(ctx, key)
	if err != nil {
		h.logger.Error("Failed to get state", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if state == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "state not found"})
		return
	}

	c.JSON(http.StatusOK, state)
}

// UpdateState updates a state
func (h *HTTPHandler) UpdateState(c *gin.Context) {
	key := c.Param("key")
	var req domain.UpdateStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	state, err := h.stateRegistry.UpdateState(ctx, key, req.Value, req.TTL)
	if err != nil {
		h.logger.Error("Failed to update state", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, state)
}

// DeleteState deletes a state
func (h *HTTPHandler) DeleteState(c *gin.Context) {
	key := c.Param("key")
	ctx := c.Request.Context()
	if err := h.stateRegistry.DeleteState(ctx, key); err != nil {
		h.logger.Error("Failed to delete state", zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// EvaluatePolicy evaluates a policy against provided data
func (h *HTTPHandler) EvaluatePolicy(c *gin.Context) {
	var req domain.EvaluatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	result, err := h.policyEngine.EvaluatePolicy(ctx, req.PolicyID, req.Data)
	if err != nil {
		h.logger.Error("Failed to evaluate policy", zap.String("policy_id", req.PolicyID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// MetricsHandler returns Prometheus metrics
func (h *HTTPHandler) MetricsHandler(c *gin.Context) {
	// This would typically use promhttp.Handler() in production
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
