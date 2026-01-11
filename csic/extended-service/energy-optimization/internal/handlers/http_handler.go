package handlers

import (
	"net/http"
	"time"

	"energy-optimization/internal/domain"
	"energy-optimization/internal/service"

	"github.com/gin-gonic/gin"
)

// EnergyHandler handles HTTP requests for energy operations
type EnergyHandler struct {
	energyService *service.EnergyService
}

// NewEnergyHandler creates a new energy handler
func NewEnergyHandler(energyService *service.EnergyService) *EnergyHandler {
	return &EnergyHandler{
		energyService: energyService,
	}
}

// RegisterRoutes registers the energy routes
func (h *EnergyHandler) RegisterRoutes(router *gin.RouterGroup) {
	readings := router.Group("/readings")
	{
		readings.POST("", h.IngestReading)
		readings.GET("/:id", h.GetReading)
		readings.GET("/facility/:facility_id", h.GetReadingsByFacility)
	}

	facility := router.Group("/facility")
	{
		facility.GET("/:facility_id/load", h.GetFacilityLoad)
		facility.GET("/:facility_id/consumption", h.GetDailyConsumption)
		facility.GET("/:facility_id/status", h.GetFacilityStatus)
		facility.GET("/:facility_id/zones", h.GetZoneLoads)
	}
}

// IngestReading handles POST /api/v1/readings
func (h *EnergyHandler) IngestReading(c *gin.Context) {
	var reading domain.EnergyReading
	if err := c.ShouldBindJSON(&reading); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.energyService.IngestReading(c.Request.Context(), &reading); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "processing_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        reading.ID,
		"message":   "reading ingested successfully",
		"timestamp": reading.Timestamp,
	})
}

// GetReading handles GET /api/v1/readings/:id
func (h *EnergyHandler) GetReading(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "reading ID is required",
		})
		return
	}

	reading, err := h.energyService.GetReading(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	if reading == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "reading not found",
		})
		return
	}

	c.JSON(http.StatusOK, reading)
}

// GetReadingsByFacility handles GET /api/v1/readings/facility/:facility_id
func (h *EnergyHandler) GetReadingsByFacility(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	// Parse time range parameters
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

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

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := parseInt(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	readings, err := h.energyService.GetReadingsByFacility(c.Request.Context(), facilityID, startTime, endTime, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"facility_id": facilityID,
		"start_time":  startTime,
		"end_time":    endTime,
		"count":       len(readings),
		"readings":    readings,
	})
}

// GetFacilityLoad handles GET /api/v1/facility/:facility_id/load
func (h *EnergyHandler) GetFacilityLoad(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	load, err := h.energyService.GetFacilityCurrentLoad(c.Request.Context(), facilityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"facility_id": facilityID,
		"current_load": load,
		"unit":        "kWh",
	})
}

// GetDailyConsumption handles GET /api/v1/facility/:facility_id/consumption
func (h *EnergyHandler) GetDailyConsumption(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	consumption, err := h.energyService.GetDailyConsumption(c.Request.Context(), facilityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"facility_id": facilityID,
		"date":        time.Now().Format("2006-01-02"),
		"consumption": consumption,
		"unit":        "kWh",
	})
}

// GetFacilityStatus handles GET /api/v1/facility/:facility_id/status
func (h *EnergyHandler) GetFacilityStatus(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	status, err := h.energyService.GetFacilityStatus(c.Request.Context(), facilityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetZoneLoads handles GET /api/v1/facility/:facility_id/zones
func (h *EnergyHandler) GetZoneLoads(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	// This would need the load balancing service
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"facility_id": facilityID,
		"zones":       []interface{}{},
	})
}

// parseInt parses a string to int
func parseInt(s string) (int, error) {
	var result int
	_, err := gin.Mode()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// CarbonHandler handles HTTP requests for carbon operations
type CarbonHandler struct {
	carbonService *service.CarbonService
}

// NewCarbonHandler creates a new carbon handler
func NewCarbonHandler(carbonService *service.CarbonService) *CarbonHandler {
	return &CarbonHandler{
		carbonService: carbonService,
	}
}

// RegisterRoutes registers the carbon routes
func (h *CarbonHandler) RegisterRoutes(router *gin.RouterGroup) {
	carbon := router.Group("/carbon")
	{
		carbon.POST("/process", h.ProcessEmission)
		carbon.GET("/emissions/:facility_id", h.GetEmissions)
		carbon.GET("/emissions/:facility_id/by-source", h.GetEmissionsBySource)
		carbon.GET("/emissions/:facility_id/total", h.GetTotalEmissions)
		carbon.GET("/budget/:facility_id", h.GetBudget)
		carbon.POST("/budget", h.CreateBudget)
		carbon.GET("/status/:facility_id", h.GetBudgetStatus)
		carbon.GET("/intensity/:facility_id", h.GetCarbonIntensity)
	}
}

// ProcessEmission handles POST /api/v1/carbon/process
func (h *CarbonHandler) ProcessEmission(c *gin.Context) {
	var reading domain.EnergyReading
	if err := c.ShouldBindJSON(&reading); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	emission, err := h.carbonService.ProcessReading(c.Request.Context(), &reading)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "processing_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":              emission.ID,
		"total_emission":  emission.TotalEmission,
		"carbon_intensity": emission.CarbonIntensity,
		"unit":            "kg CO2",
	})
}

// GetEmissions handles GET /api/v1/carbon/emissions/:facility_id
func (h *CarbonHandler) GetEmissions(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

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

	emissions, err := h.carbonService.GetCarbonEmissions(c.Request.Context(), facilityID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"facility_id": facilityID,
		"start_time":  startTime,
		"end_time":    endTime,
		"count":       len(emissions),
		"emissions":   emissions,
	})
}

// GetEmissionsBySource handles GET /api/v1/carbon/emissions/:facility_id/by-source
func (h *CarbonHandler) GetEmissionsBySource(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

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

	emissions, err := h.carbonService.GetEmissionsBySource(c.Request.Context(), facilityID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"facility_id": facilityID,
		"start_time":  startTime,
		"end_time":    endTime,
		"by_source":   emissions,
		"unit":        "kg CO2",
	})
}

// GetTotalEmissions handles GET /api/v1/carbon/emissions/:facility_id/total
func (h *CarbonHandler) GetTotalEmissions(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

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

	total, err := h.carbonService.GetTotalEmissions(c.Request.Context(), facilityID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"facility_id":  facilityID,
		"start_time":   startTime,
		"end_time":     endTime,
		"total":        total,
		"unit":         "kg CO2",
	})
}

// GetBudget handles GET /api/v1/carbon/budget/:facility_id
func (h *CarbonHandler) GetBudget(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	budget, err := h.carbonService.GetBudget(c.Request.Context(), facilityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	if budget == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "no budget configured for this facility",
		})
		return
	}

	c.JSON(http.StatusOK, budget)
}

// CreateBudget handles POST /api/v1/carbon/budget
func (h *CarbonHandler) CreateBudget(c *gin.Context) {
	var budget domain.CarbonBudget
	if err := c.ShouldBindJSON(&budget); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.carbonService.CreateBudget(c.Request.Context(), &budget); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "creation_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        budget.ID,
		"message":   "budget created successfully",
		"budget":    budget,
	})
}

// GetBudgetStatus handles GET /api/v1/carbon/status/:facility_id
func (h *CarbonHandler) GetBudgetStatus(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	status, err := h.carbonService.CheckBudgetStatus(c.Request.Context(), facilityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	budget, _ := h.carbonService.GetBudget(c.Request.Context(), facilityID)

	c.JSON(http.StatusOK, gin.H{
		"facility_id": facilityID,
		"status":      status,
		"budget":      budget,
	})
}

// GetCarbonIntensity handles GET /api/v1/carbon/intensity/:facility_id
func (h *CarbonHandler) GetCarbonIntensity(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

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

	intensity, err := h.carbonService.GetCarbonIntensity(c.Request.Context(), facilityID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"facility_id":   facilityID,
		"carbon_intensity": intensity,
		"unit":         "g CO2/kWh",
	})
}

// LoadBalancingHandler handles HTTP requests for load balancing operations
type LoadBalancingHandler struct {
	loadBalanceService *service.LoadBalancingService
}

// NewLoadBalancingHandler creates a new load balancing handler
func NewLoadBalancingHandler(loadBalanceService *service.LoadBalancingService) *LoadBalancingHandler {
	return &LoadBalancingHandler{
		loadBalanceService: loadBalanceService,
	}
}

// RegisterRoutes registers the load balancing routes
func (h *LoadBalancingHandler) RegisterRoutes(router *gin.RouterGroup) {
	load := router.Group("/load")
	{
		load.POST("/metrics", h.RecordLoadMetric)
		load.GET("/evaluate/:facility_id", h.EvaluateLoad)
		load.GET("/zones/:facility_id", h.GetZoneLoads)
	}

	balancing := router.Group("/balancing")
	{
		balancing.POST("/decisions", h.CreateDecision)
		balancing.GET("/decisions/:facility_id", h.GetPendingDecisions)
		balancing.POST("/decisions/:id/execute", h.ExecuteDecision)
		balancing.POST("/decisions/:id/cancel", h.CancelDecision)
		balancing.POST("/auto/:facility_id", h.AutoBalance)
	}

	optimization := router.Group("/optimization")
	{
		optimization.GET("/recommend/:facility_id", h.GetRecommendations)
		optimization.GET("/report/:facility_id", h.GetOptimizationReport)
	}
}

// RecordLoadMetric handles POST /api/v1/load/metrics
func (h *LoadBalancingHandler) RecordLoadMetric(c *gin.Context) {
	var metric domain.LoadMetric
	if err := c.ShouldBindJSON(&metric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.loadBalanceService.RecordLoadMetric(c.Request.Context(), &metric); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "processing_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      metric.ID,
		"message": "load metric recorded successfully",
	})
}

// EvaluateLoad handles GET /api/v1/load/evaluate/:facility_id
func (h *LoadBalancingHandler) EvaluateLoad(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	evaluation, err := h.loadBalanceService.EvaluateLoad(c.Request.Context(), facilityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "evaluation_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, evaluation)
}

// GetZoneLoads handles GET /api/v1/load/zones/:facility_id
func (h *LoadBalancingHandler) GetZoneLoads(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	zones, err := h.loadBalanceService.GetZoneLoads(c.Request.Context(), facilityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"facility_id": facilityID,
		"zones":       zones,
	})
}

// CreateDecision handles POST /api/v1/balancing/decisions
func (h *LoadBalancingHandler) CreateDecision(c *gin.Context) {
	var decision domain.LoadBalancingDecision
	if err := c.ShouldBindJSON(&decision); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.loadBalanceService.CreateLoadBalancingDecision(c.Request.Context(), &decision); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "creation_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      decision.ID,
		"message": "load balancing decision created",
	})
}

// GetPendingDecisions handles GET /api/v1/balancing/decisions/:facility_id
func (h *LoadBalancingHandler) GetPendingDecisions(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	decisions, err := h.loadBalanceService.GetPendingDecisions(c.Request.Context(), facilityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"facility_id": facilityID,
		"count":       len(decisions),
		"decisions":   decisions,
	})
}

// ExecuteDecision handles POST /api/v1/balancing/decisions/:id/execute
func (h *LoadBalancingHandler) ExecuteDecision(c *gin.Context) {
	decisionID := c.Param("id")
	if decisionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "decision ID is required",
		})
		return
	}

	if err := h.loadBalanceService.ExecuteDecision(c.Request.Context(), decisionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "execution_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      decisionID,
		"message": "decision executed successfully",
	})
}

// CancelDecision handles POST /api/v1/balancing/decisions/:id/cancel
func (h *LoadBalancingHandler) CancelDecision(c *gin.Context) {
	decisionID := c.Param("id")
	if decisionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "decision ID is required",
		})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	if err := h.loadBalanceService.CancelDecision(c.Request.Context(), decisionID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "cancellation_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      decisionID,
		"message": "decision cancelled",
	})
}

// AutoBalance handles POST /api/v1/balancing/auto/:facility_id
func (h *LoadBalancingHandler) AutoBalance(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	decisions, err := h.loadBalanceService.AutoBalance(c.Request.Context(), facilityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "balancing_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"facility_id": facilityID,
		"count":       len(decisions),
		"decisions":   decisions,
	})
}

// GetRecommendations handles GET /api/v1/optimization/recommend/:facility_id
func (h *LoadBalancingHandler) GetRecommendations(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	recommendations, err := h.loadBalanceService.AnalyzeAndRecommend(c.Request.Context(), facilityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "analysis_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"facility_id":   facilityID,
		"count":         len(recommendations),
		"recommendations": recommendations,
	})
}

// GetOptimizationReport handles GET /api/v1/optimization/report/:facility_id
func (h *LoadBalancingHandler) GetOptimizationReport(c *gin.Context) {
	facilityID := c.Param("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "facility_id is required",
		})
		return
	}

	startTime := time.Now().Add(-7 * 24 * time.Hour)
	endTime := time.Now()

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

	report, err := h.loadBalanceService.GetOptimizationReport(c.Request.Context(), facilityID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "report_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}
