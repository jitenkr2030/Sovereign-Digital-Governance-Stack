package handlers

import (
	"encoding/json"
	"net/http"
	"smart-cities/internal/repository"
	"smart-cities/internal/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// HTTPHandler handles HTTP requests for the Smart Cities service
type HTTPHandler struct {
	trafficService    *service.TrafficService
	emergencyService  *service.EmergencyService
	wasteService      *service.WasteService
	lightingService   *service.LightingService
	postgresRepo      *repository.PostgresRepository
	redisRepo         *repository.RedisRepository
}

// NewHTTPHandler creates a new HTTP handler instance
func NewHTTPHandler(
	traffic *service.TrafficService,
	emergency *service.EmergencyService,
	waste *service.WasteService,
	lighting *service.LightingService,
	pg *repository.PostgresRepository,
	rd *repository.RedisRepository,
) *HTTPHandler {
	return &HTTPHandler{
		trafficService:   traffic,
		emergencyService: emergency,
		wasteService:     waste,
		lightingService:  lighting,
		postgresRepo:     pg,
		redisRepo:        rd,
	}
}

// SetupRoutes configures all API routes
func (h *HTTPHandler) SetupRoutes(r *gin.Engine) {
	// Health check
	r.GET("/health", h.HealthCheck)
	
	// API v1
	v1 := r.Group("/api/v1")
	{
		// Traffic routes
		traffic := v1.Group("/traffic")
		{
			traffic.GET("/nodes", h.GetTrafficNodes)
			traffic.GET("/nodes/:id", h.GetTrafficNode)
			traffic.GET("/nodes/:id/analytics", h.GetTrafficAnalytics)
			traffic.GET("/incidents", h.GetTrafficIncidents)
			traffic.POST("/incidents", h.CreateTrafficIncident)
			traffic.GET("/predict/:node_id", h.PredictTraffic)
			traffic.POST("/sync-lights", h.SynchronizeLights)
		}
		
		// Emergency routes
		emergency := v1.Group("/emergency")
		{
			emergency.GET("/units", h.GetEmergencyUnits)
			emergency.GET("/units/:id", h.GetEmergencyUnit)
			emergency.GET("/incidents", h.GetEmergencyIncidents)
			emergency.POST("/incidents", h.CreateEmergencyIncident)
			emergency.POST("/dispatch", h.DispatchUnit)
			emergency.POST("/resolve", h.ResolveIncident)
			emergency.GET("/metrics", h.GetEmergencyMetrics)
		}
		
		// Waste management routes
		waste := v1.Group("/waste")
		{
			waste.GET("/bins", h.GetWasteBins)
			waste.GET("/bins/:id", h.GetWasteBin)
			waste.PUT("/bins/:id/fill", h.UpdateBinFillLevel)
			waste.POST("/collect", h.CollectWaste)
			waste.GET("/route/:vehicle_id", h.GetCollectionRoute)
			waste.GET("/analytics", h.GetWasteAnalytics)
		}
		
		// Lighting routes
		lighting := v1.Group("/lighting")
		{
			lighting.GET("/fixtures", h.GetLightingFixtures)
			lighting.GET("/fixtures/:id", h.GetLightingFixture)
			lighting.PUT("/fixtures/:id/brightness", h.SetBrightness)
			lighting.POST("/zones/:id/adjust", h.AdjustZoneLighting)
			lighting.POST("/fixtures/:id/motion", h.HandleMotionEvent)
			lighting.GET("/metrics", h.GetLightingMetrics)
			lighting.GET("/savings", h.GetEnergySavings)
		}
	}
}

// HealthCheck returns service health status
func (h *HTTPHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "smart-cities",
	})
}

// Traffic handlers

func (h *HTTPHandler) GetTrafficNodes(c *gin.Context) {
	nodes, err := h.postgresRepo.GetAllTrafficNodes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

func (h *HTTPHandler) GetTrafficNode(c *gin.Context) {
	id := c.Param("id")
	node, err := h.postgresRepo.GetTrafficNode(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	c.JSON(http.StatusOK, node)
}

func (h *HTTPHandler) GetTrafficAnalytics(c *gin.Context) {
	id := c.Param("id")
	timeRange := 24 * time.Hour
	if tr := c.Query("time_range"); tr != "" {
		if parsed, err := time.ParseDuration(tr); err == nil {
			timeRange = parsed
		}
	}
	
	analytics, err := h.trafficService.GetTrafficAnalytics(c.Request.Context(), id, timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, analytics)
}

func (h *HTTPHandler) GetTrafficIncidents(c *gin.Context) {
	incidents, err := h.postgresRepo.GetAllTrafficIncidents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, incidents)
}

func (h *HTTPHandler) CreateTrafficIncident(c *gin.Context) {
	var incident service.TrafficIncident
	if err := c.ShouldBindJSON(&incident); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.trafficService.HandleTrafficIncident(c.Request.Context(), incident.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, incident)
}

func (h *HTTPHandler) PredictTraffic(c *gin.Context) {
	nodeID := c.Param("node_id")
	hour := time.Now().Hour()
	if h := c.Query("hour"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil {
			hour = parsed
		}
	}
	
	prediction, confidence, err := h.trafficService.PredictTrafficPattern(c.Request.Context(), nodeID, hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"node_id":     nodeID,
		"predicted_count": prediction,
		"confidence":  confidence,
		"hour":        hour,
	})
}

func (h *HTTPHandler) SynchronizeLights(c *gin.Context) {
	var req struct {
		RouteID string `json:"route_id"`
		Offset  string `json:"offset"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	offset := 0 * time.Second
	if req.Offset != "" {
		if parsed, err := time.ParseDuration(req.Offset); err == nil {
			offset = parsed
		}
	}
	
	if err := h.trafficService.SynchronizeTrafficLights(c.Request.Context(), req.RouteID, offset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "synchronized"})
}

// Emergency handlers

func (h *HTTPHandler) GetEmergencyUnits(c *gin.Context) {
	units, err := h.postgresRepo.GetAllEmergencyUnits(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, units)
}

func (h *HTTPHandler) GetEmergencyUnit(c *gin.Context) {
	id := c.Param("id")
	unit, err := h.postgresRepo.GetEmergencyUnit(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "unit not found"})
		return
	}
	c.JSON(http.StatusOK, unit)
}

func (h *HTTPHandler) GetEmergencyIncidents(c *gin.Context) {
	incidents, err := h.postgresRepo.GetAllEmergencyIncidents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, incidents)
}

func (h *HTTPHandler) CreateEmergencyIncident(c *gin.Context) {
	var incident service.EmergencyIncident
	if err := c.ShouldBindJSON(&incident); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	result, err := h.emergencyService.ProcessIncident(c.Request.Context(), &incident)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *HTTPHandler) DispatchUnit(c *gin.Context) {
	var req struct {
		UnitID     string `json:"unit_id"`
		IncidentID string `json:"incident_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.emergencyService.DispatchUnit(c.Request.Context(), req.UnitID, req.IncidentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "dispatched"})
}

func (h *HTTPHandler) ResolveIncident(c *gin.Context) {
	var req struct {
		IncidentID string `json:"incident_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.emergencyService.ResolveIncident(c.Request.Context(), req.IncidentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "resolved"})
}

func (h *HTTPHandler) GetEmergencyMetrics(c *gin.Context) {
	timeRange := 24 * time.Hour
	if tr := c.Query("time_range"); tr != "" {
		if parsed, err := time.ParseDuration(tr); err == nil {
			timeRange = parsed
		}
	}
	
	metrics, err := h.emergencyService.GetEmergencyMetrics(c.Request.Context(), timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

// Waste handlers

func (h *HTTPHandler) GetWasteBins(c *gin.Context) {
	bins, err := h.postgresRepo.GetAllWasteBins(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bins)
}

func (h *HTTPHandler) GetWasteBin(c *gin.Context) {
	id := c.Param("id")
	bin, err := h.postgresRepo.GetWasteBin(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bin not found"})
		return
	}
	c.JSON(http.StatusOK, bin)
}

func (h *HTTPHandler) UpdateBinFillLevel(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		FillLevel int `json:"fill_level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.wasteService.UpdateBinFillLevel(c.Request.Context(), id, req.FillLevel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *HTTPHandler) CollectWaste(c *gin.Context) {
	var req struct {
		BinID     string `json:"bin_id"`
		VehicleID string `json:"vehicle_id"`
		Weight    int    `json:"weight"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.wasteService.CollectWaste(c.Request.Context(), req.BinID, req.VehicleID, req.Weight); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "collected"})
}

func (h *HTTPHandler) GetCollectionRoute(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")
	maxBins := 20
	if mb := c.Query("max_bins"); mb != "" {
		if parsed, err := strconv.Atoi(mb); err == nil {
			maxBins = parsed
		}
	}
	
	route, err := h.wasteService.GetOptimalCollectionRoute(c.Request.Context(), vehicleID, maxBins)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"vehicle_id": vehicleID,
		"route":      route,
	})
}

func (h *HTTPHandler) GetWasteAnalytics(c *gin.Context) {
	timeRange := 24 * time.Hour
	if tr := c.Query("time_range"); tr != "" {
		if parsed, err := time.ParseDuration(tr); err == nil {
			timeRange = parsed
		}
	}
	
	analytics, err := h.wasteService.GetWasteAnalytics(c.Request.Context(), timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, analytics)
}

// Lighting handlers

func (h *HTTPHandler) GetLightingFixtures(c *gin.Context) {
	fixtures, err := h.postgresRepo.GetAllLightingFixtures(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fixtures)
}

func (h *HTTPHandler) GetLightingFixture(c *gin.Context) {
	id := c.Param("id")
	fixture, err := h.postgresRepo.GetLightingFixture(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "fixture not found"})
		return
	}
	c.JSON(http.StatusOK, fixture)
}

func (h *HTTPHandler) SetBrightness(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Brightness int `json:"brightness"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.lightingService.SetFixtureBrightness(c.Request.Context(), id, req.Brightness); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "brightness set", "brightness": req.Brightness})
}

func (h *HTTPHandler) AdjustZoneLighting(c *gin.Context) {
	zoneID := c.Param("id")
	var req struct {
		Mode string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.lightingService.AdjustZoneLighting(c.Request.Context(), zoneID, req.Mode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "zone adjusted", "zone_id": zoneID, "mode": req.Mode})
}

func (h *HTTPHandler) HandleMotionEvent(c *gin.Context) {
	id := c.Param("id")
	if err := h.lightingService.HandleMotionEvent(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "motion detected", "fixture_id": id})
}

func (h *HTTPHandler) GetLightingMetrics(c *gin.Context) {
	timeRange := 24 * time.Hour
	if tr := c.Query("time_range"); tr != "" {
		if parsed, err := time.ParseDuration(tr); err == nil {
			timeRange = parsed
		}
	}
	
	metrics, err := h.lightingService.GetLightingMetrics(c.Request.Context(), timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (h *HTTPHandler) GetEnergySavings(c *gin.Context) {
	baseline := 1000.0 // Default baseline kWh
	if b := c.Query("baseline"); b != "" {
		if parsed, err := strconv.ParseFloat(b, 64); err == nil {
			baseline = parsed
		}
	}
	
	savings, err := h.lightingService.GetEnergySavings(c.Request.Context(), baseline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, savings)
}

// Prometheus metrics endpoint
func (h *HTTPHandler) PrometheusMetrics(c *gin.Context) {
	// This would return Prometheus-formatted metrics
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("# Smart Cities Service Metrics\n"))
}

// Utility function to parse request body
func parseRequestBody(c *gin.Context, v interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(v)
}
