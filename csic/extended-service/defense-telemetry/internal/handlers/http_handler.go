package handlers

import (
	"net/http"
	"time"

	"defense-telemetry/internal/domain"
	"defense-telemetry/internal/service"

	"github.com/gin-gonic/gin"
)

// TelemetryHandler handles HTTP requests for telemetry operations
type TelemetryHandler struct {
	telemetryService *service.TelemetryService
}

// NewTelemetryHandler creates a new telemetry handler
func NewTelemetryHandler(telemetryService *service.TelemetryService) *TelemetryHandler {
	return &TelemetryHandler{
		telemetryService: telemetryService,
	}
}

// RegisterRoutes registers the telemetry routes
func (h *TelemetryHandler) RegisterRoutes(router *gin.RouterGroup) {
	telemetry := router.Group("/telemetry")
	{
		telemetry.POST("/ingest", h.IngestTelemetry)
		telemetry.POST("/batch", h.IngestBatch)
		telemetry.GET("/unit/:unit_id", h.GetLatestTelemetry)
		telemetry.GET("/unit/:unit_id/status", h.GetUnitStatus)
		telemetry.GET("/units", h.GetAllUnits)
		telemetry.POST("/units", h.CreateUnit)
		telemetry.GET("/units/nearby", h.GetNearbyUnits)
	}
}

// IngestTelemetry handles POST /api/v1/telemetry/ingest
func (h *TelemetryHandler) IngestTelemetry(c *gin.Context) {
	var data domain.TelemetryData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.telemetryService.IngestTelemetry(c.Request.Context(), &data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "processing_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// IngestBatch handles POST /api/v1/telemetry/batch
func (h *TelemetryHandler) IngestBatch(c *gin.Context) {
	var batch domain.TelemetryBatch
	if err := c.ShouldBindJSON(&batch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.telemetryService.IngestBatch(c.Request.Context(), &batch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "processing_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetLatestTelemetry handles GET /api/v1/telemetry/unit/:unit_id
func (h *TelemetryHandler) GetLatestTelemetry(c *gin.Context) {
	unitID := c.Param("unit_id")
	if unitID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "unit_id is required",
		})
		return
	}

	telemetry, err := h.telemetryService.GetLatestTelemetry(c.Request.Context(), unitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	if telemetry == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "telemetry not found",
		})
		return
	}

	c.JSON(http.StatusOK, telemetry)
}

// GetUnitStatus handles GET /api/v1/telemetry/unit/:unit_id/status
func (h *TelemetryHandler) GetUnitStatus(c *gin.Context) {
	unitID := c.Param("unit_id")
	if unitID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "unit_id is required",
		})
		return
	}

	status, err := h.telemetryService.GetUnitStatus(c.Request.Context(), unitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	if status == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "unit not found",
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetAllUnits handles GET /api/v1/telemetry/units
func (h *TelemetryHandler) GetAllUnits(c *gin.Context) {
	units, err := h.telemetryService.GetAllUnits(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(units),
		"units": units,
	})
}

// CreateUnit handles POST /api/v1/telemetry/units
func (h *TelemetryHandler) CreateUnit(c *gin.Context) {
	var unit domain.DefenseUnit
	if err := c.ShouldBindJSON(&unit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.telemetryService.CreateDefenseUnit(c.Request.Context(), &unit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "creation_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      unit.ID,
		"message": "unit created successfully",
		"unit":    unit,
	})
}

// GetNearbyUnits handles GET /api/v1/telemetry/units/nearby
func (h *TelemetryHandler) GetNearbyUnits(c *gin.Context) {
	lat := c.Query("lat")
	lng := c.Query("lng")
	radius := c.DefaultQuery("radius", "50")

	if lat == "" || lng == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "lat and lng parameters are required",
		})
		return
	}

	var latF, lngF, radiusF float64
	if _, err := parseFloat(lat, &latF); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_lat"})
		return
	}
	if _, err := parseFloat(lng, &lngF); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_lng"})
		return
	}
	if _, err := parseFloat(radius, &radiusF); err != nil {
		radiusF = 50
	}

	unitIDs, err := h.telemetryService.GetNearbyUnits(c.Request.Context(), latF, lngF, radiusF)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"lat":      latF,
		"lng":      lngF,
		"radius":   radiusF,
		"unit_ids": unitIDs,
	})
}

// parseFloat helper
func parseFloat(s string, f *float64) (bool, error) {
	_, err := gin.Mode()
	if err != nil {
		return false, err
	}
	return true, nil
}

// ThreatHandler handles HTTP requests for threat operations
type ThreatHandler struct {
	threatService *service.ThreatService
}

// NewThreatHandler creates a new threat handler
func NewThreatHandler(threatService *service.ThreatService) *ThreatHandler {
	return &ThreatHandler{
		threatService: threatService,
	}
}

// RegisterRoutes registers the threat routes
func (h *ThreatHandler) RegisterRoutes(router *gin.RouterGroup) {
	threats := router.Group("/threats")
	{
		threats.POST("", h.CreateThreat)
		threats.GET("/active", h.GetActiveThreats)
		threats.GET("/stats", h.GetThreatStats)
		threats.GET("/:id", h.GetThreat)
		threats.POST("/:id/resolve", h.ResolveThreat)
		threats.POST("/:id/escalate", h.EscalateThreat)
	}
}

// CreateThreat handles POST /api/v1/threats
func (h *ThreatHandler) CreateThreat(c *gin.Context) {
	var threat domain.ThreatEvent
	if err := c.ShouldBindJSON(&threat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.threatService.CreateThreatEvent(c.Request.Context(), &threat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "creation_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      threat.ID,
		"message": "threat created successfully",
	})
}

// GetActiveThreats handles GET /api/v1/threats/active
func (h *ThreatHandler) GetActiveThreats(c *gin.Context) {
	minSeverity := c.DefaultQuery("min_severity", "1")

	var minSev int
	if _, err := parseInt(minSeverity); err != nil {
		minSev = 1
	}

	threats, err := h.threatService.GetActiveThreats(c.Request.Context(), minSev)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":   len(threats),
		"threats": threats,
	})
}

// GetThreatStats handles GET /api/v1/threats/stats
func (h *ThreatHandler) GetThreatStats(c *gin.Context) {
	stats, err := h.threatService.GetThreatStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetThreat handles GET /api/v1/threats/:id
func (h *ThreatHandler) GetThreat(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_required"})
		return
	}

	threat, err := h.threatService.GetThreat(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, threat)
}

// ResolveThreat handles POST /api/v1/threats/:id/resolve
func (h *ThreatHandler) ResolveThreat(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_required"})
		return
	}

	var req struct {
		Resolution string `json:"resolution"`
	}
	c.ShouldBindJSON(&req)

	if err := h.threatService.ResolveThreat(c.Request.Context(), id, req.Resolution); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "update_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"message": "threat resolved",
	})
}

// EscalateThreat handles POST /api/v1/threats/:id/escalate
func (h *ThreatHandler) EscalateThreat(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_required"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	if err := h.threatService.EscalateThreat(c.Request.Context(), id, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "update_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"message": "threat escalated",
	})
}

// MaintenanceHandler handles HTTP requests for maintenance operations
type MaintenanceHandler struct {
	maintenanceService *service.MaintenanceService
}

// NewMaintenanceHandler creates a new maintenance handler
func NewMaintenanceHandler(maintenanceService *service.MaintenanceService) *MaintenanceHandler {
	return &MaintenanceHandler{
		maintenanceService: maintenanceService,
	}
}

// RegisterRoutes registers the maintenance routes
func (h *MaintenanceHandler) RegisterRoutes(router *gin.RouterGroup) {
	maintenance := router.Group("/maintenance")
	{
		maintenance.POST("/logs", h.CreateMaintenanceLog)
		maintenance.GET("/unit/:unit_id/predictions", h.GetPredictions)
		maintenance.GET("/unit/:unit_id/health", h.AnalyzeHealth)
		maintenance.GET("/upcoming", h.GetUpcomingMaintenance)
	}
}

// CreateMaintenanceLog handles POST /api/v1/maintenance/logs
func (h *MaintenanceHandler) CreateMaintenanceLog(c *gin.Context) {
	var log domain.MaintenanceLog
	if err := c.ShouldBindJSON(&log); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.maintenanceService.CreateMaintenanceLog(c.Request.Context(), &log); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "creation_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      log.ID,
		"message": "maintenance log created",
	})
}

// GetPredictions handles GET /api/v1/maintenance/unit/:unit_id/predictions
func (h *MaintenanceHandler) GetPredictions(c *gin.Context) {
	unitID := c.Param("unit_id")
	if unitID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unit_id_required"})
		return
	}

	predictions, err := h.maintenanceService.GetMaintenancePredictions(c.Request.Context(), unitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unit_id":     unitID,
		"count":       len(predictions),
		"predictions": predictions,
	})
}

// AnalyzeHealth handles GET /api/v1/maintenance/unit/:unit_id/health
func (h *MaintenanceHandler) AnalyzeHealth(c *gin.Context) {
	unitID := c.Param("unit_id")
	if unitID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unit_id_required"})
		return
	}

	analysis, err := h.maintenanceService.AnalyzeEquipmentHealth(c.Request.Context(), unitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "analysis_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetUpcomingMaintenance handles GET /api/v1/maintenance/upcoming
func (h *MaintenanceHandler) GetUpcomingMaintenance(c *gin.Context) {
	days := c.DefaultQuery("days", "7")

	logs, err := h.maintenanceService.GetUpcomingMaintenance(c.Request.Context(), 7)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"days":      days,
		"count":     len(logs),
		"maintenance": logs,
	})
}

// DeploymentHandler handles HTTP requests for deployment operations
type DeploymentHandler struct {
	deploymentService *service.DeploymentService
}

// NewDeploymentHandler creates a new deployment handler
func NewDeploymentHandler(deploymentService *service.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{
		deploymentService: deploymentService,
	}
}

// RegisterRoutes registers the deployment routes
func (h *DeploymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	deployment := router.Group("/deployment")
	{
		deployment.POST("", h.CreateDeployment)
		deployment.GET("/active", h.GetActiveDeployments)
		deployment.GET("/stats", h.GetDeploymentStats)
		deployment.GET("/optimize/:unit_id", h.OptimizeDeployment)
		deployment.GET("/zones", h.GetStrategicZones)
		deployment.POST("/zones", h.CreateStrategicZone)
	}
}

// CreateDeployment handles POST /api/v1/deployment
func (h *DeploymentHandler) CreateDeployment(c *gin.Context) {
	var deployment domain.Deployment
	if err := c.ShouldBindJSON(&deployment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.deploymentService.CreateDeployment(c.Request.Context(), &deployment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "creation_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        deployment.ID,
		"message":   "deployment created",
		"deployment": deployment,
	})
}

// GetActiveDeployments handles GET /api/v1/deployment/active
func (h *DeploymentHandler) GetActiveDeployments(c *gin.Context) {
	deployments, err := h.deploymentService.GetActiveDeployments(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":       len(deployments),
		"deployments": deployments,
	})
}

// GetDeploymentStats handles GET /api/v1/deployment/stats
func (h *DeploymentHandler) GetDeploymentStats(c *gin.Context) {
	stats, err := h.deploymentService.GetDeploymentStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// OptimizeDeployment handles GET /api/v1/deployment/optimize/:unit_id
func (h *DeploymentHandler) OptimizeDeployment(c *gin.Context) {
	unitID := c.Param("unit_id")
	if unitID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unit_id_required"})
		return
	}

	recommendation, err := h.deploymentService.OptimizeDeployment(c.Request.Context(), unitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "optimization_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, recommendation)
}

// GetStrategicZones handles GET /api/v1/deployment/zones
func (h *DeploymentHandler) GetStrategicZones(c *gin.Context) {
	zones, err := h.deploymentService.GetStrategicZones(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "retrieval_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(zones),
		"zones": zones,
	})
}

// CreateStrategicZone handles POST /api/v1/deployment/zones
func (h *DeploymentHandler) CreateStrategicZone(c *gin.Context) {
	var zone domain.StrategicZone
	if err := c.ShouldBindJSON(&zone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.deploymentService.CreateStrategicZone(c.Request.Context(), &zone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "creation_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      zone.ID,
		"message": "zone created",
		"zone":    zone,
	})
}
