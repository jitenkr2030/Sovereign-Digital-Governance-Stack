package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/energy-integration/energy/internal/core/domain"
	"github.com/energy-integration/energy/internal/core/ports"
	"github.com/energy-integration/energy/internal/core/services"
	"github.com/gin-gonic/gin"
)

// EnergyHandler handles HTTP requests for energy integration operations.
type EnergyHandler struct {
	energyService *services.EnergyService
}

// NewEnergyHandler creates a new EnergyHandler with the required service dependencies.
func NewEnergyHandler(energyService *services.EnergyService) *EnergyHandler {
	return &EnergyHandler{
		energyService: energyService,
	}
}

// RegisterRoutes registers all energy-related routes on the given router group.
func (h *EnergyHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Profile routes
	profiles := router.Group("/energy/profiles")
	{
		profiles.POST("", h.CreateProfile)
		profiles.GET("", h.ListProfiles)
		profiles.GET("/:chain_id", h.GetProfile)
	}

	// Footprint routes
	footprints := router.Group("/energy/footprints")
	{
		footprints.POST("/calculate", h.CalculateFootprint)
		footprints.POST("", h.RecordFootprint)
		footprints.GET("/:tx_hash", h.GetFootprint)
		footprints.GET("", h.ListFootprints)
		footprints.GET("/summary", h.GetFootprintSummary)
		footprints.POST("/:tx_hash/offset", h.ApplyOffset)
	}

	// Certificate routes
	certificates := router.Group("/energy/certificates")
	{
		certificates.POST("", h.CreateCertificate)
		certificates.GET("", h.ListCertificates)
		certificates.GET("/:id", h.GetCertificate)
	}
}

// CreateProfileRequest represents the request body for creating an energy profile.
type CreateProfileRequest struct {
	ChainID        string  `json:"chain_id" binding:"required"`
	ChainName      string  `json:"chain_name" binding:"required"`
	ConsensusType  string  `json:"consensus_type" binding:"required"`
	AvgKWhPerTx    float64 `json:"avg_kwh_per_tx" binding:"required,gt=0"`
	BaseCarbonGrams float64 `json:"base_carbon_grams" binding:"required,gt=0"`
	EnergySource   string  `json:"energy_source" binding:"required"`
}

// CreateProfile handles POST /energy/profiles
func (h *EnergyHandler) CreateProfile(c *gin.Context) {
	var req CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.energyService.CreateProfile(
		c.Request.Context(),
		req.ChainID,
		req.ChainName,
		domain.ConsensusType(req.ConsensusType),
		req.AvgKWhPerTx,
		req.BaseCarbonGrams,
		domain.EnergySource(req.EnergySource),
	)
	if err != nil {
		switch err {
		case services.ErrProfileAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case services.ErrInvalidChainID:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, profile)
}

// ListProfiles handles GET /energy/profiles
func (h *EnergyHandler) ListProfiles(c *gin.Context) {
	profiles, err := h.energyService.ListProfiles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  profiles,
		"total": len(profiles),
	})
}

// GetProfile handles GET /energy/profiles/:chain_id
func (h *EnergyHandler) GetProfile(c *gin.Context) {
	chainID := c.Param("chain_id")

	profile, err := h.energyService.GetProfile(c.Request.Context(), chainID)
	if err != nil {
		switch err {
		case services.ErrProfileNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, profile)
}

// CalculateFootprintRequest represents the request body for calculating footprint.
type CalculateFootprintRequest struct {
	ChainID         string                 `json:"chain_id" binding:"required"`
	TransactionData map[string]interface{} `json:"transaction_data"`
	NodeLocation    string                 `json:"node_location"`
}

// CalculateFootprint handles POST /energy/footprints/calculate
func (h *EnergyHandler) CalculateFootprint(c *gin.Context) {
	var req CalculateFootprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.energyService.CalculateFootprint(
		c.Request.Context(),
		domain.EnergyCalculationRequest{
			ChainID:         req.ChainID,
			TransactionData: req.TransactionData,
			NodeLocation:    req.NodeLocation,
			Timestamp:       time.Now().UTC(),
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RecordFootprintRequest represents the request body for recording footprint.
type RecordFootprintRequest struct {
	TransactionHash string                 `json:"transaction_hash" binding:"required"`
	ChainID         string                 `json:"chain_id" binding:"required"`
	TransactionData map[string]interface{} `json:"transaction_data"`
}

// RecordFootprint handles POST /energy/footprints
func (h *EnergyHandler) RecordFootprint(c *gin.Context) {
	var req RecordFootprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	footprint, err := h.energyService.CalculateAndRecordFootprint(
		c.Request.Context(),
		req.TransactionHash,
		req.ChainID,
		req.TransactionData,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, footprint)
}

// GetFootprint handles GET /energy/footprints/:tx_hash
func (h *EnergyHandler) GetFootprint(c *gin.Context) {
	txHash := c.Param("tx_hash")

	footprint, err := h.energyService.GetFootprint(c.Request.Context(), txHash)
	if err != nil {
		switch err {
		case services.ErrFootprintNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "footprint not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, footprint)
}

// ListFootprints handles GET /energy/footprints
func (h *EnergyHandler) ListFootprints(c *gin.Context) {
	filter := ports.FootprintFilter{
		Limit:  100,
		Offset: 0,
	}

	// Parse query parameters
	if chainID := c.Query("chain_id"); chainID != "" {
		filter.ChainID = chainID
	}

	if statusStr := c.Query("status"); statusStr != "" {
		filter.OffsetStatus = []domain.OffsetStatus{domain.OffsetStatus(statusStr)}
	}

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

	footprints, err := h.energyService.ListFootprints(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   footprints,
		"total":  len(footprints),
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetFootprintSummary handles GET /energy/footprints/summary
func (h *EnergyHandler) GetFootprintSummary(c *gin.Context) {
	// Parse time parameters
	startTime := time.Now().UTC().AddDate(0, -1, 0) // Default: 1 month ago
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

	chainID := c.Query("chain_id")

	summary, err := h.energyService.GetFootprintSummary(c.Request.Context(), startTime, endTime, chainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// ApplyOffsetRequest represents the request body for applying offset.
type ApplyOffsetRequest struct {
	CertificateID string `json:"certificate_id" binding:"required"`
}

// ApplyOffset handles POST /energy/footprints/:tx_hash/offset
func (h *EnergyHandler) ApplyOffset(c *gin.Context) {
	txHash := c.Param("tx_hash")

	var req ApplyOffsetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	footprint, err := h.energyService.ApplyOffset(c.Request.Context(), txHash, req.CertificateID)
	if err != nil {
		switch err {
		case services.ErrFootprintNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "footprint not found"})
		case services.ErrCertificateNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "certificate not found"})
		case services.ErrCertificateExpired:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, footprint)
}

// CreateCertificateRequest represents the request body for creating a certificate.
type CreateCertificateRequest struct {
	CertificateNumber string  `json:"certificate_number" binding:"required"`
	EnergySource      string  `json:"energy_source" binding:"required"`
	EnergyAmount      float64 `json:"energy_amount" binding:"required,gt=0"`
	CarbonOffsetGrams float64 `json:"carbon_offset_grams" binding:"required,gt=0"`
	Registry          string  `json:"registry" binding:"required"`
}

// CreateCertificate handles POST /energy/certificates
func (h *EnergyHandler) CreateCertificate(c *gin.Context) {
	var req CreateCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cert, err := h.energyService.CreateCertificate(
		c.Request.Context(),
		req.CertificateNumber,
		domain.EnergySource(req.EnergySource),
		req.EnergyAmount,
		req.CarbonOffsetGrams,
		req.Registry,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cert)
}

// ListCertificates handles GET /energy/certificates
func (h *EnergyHandler) ListCertificates(c *gin.Context) {
	filter := ports.CertificateFilter{
		Limit:  100,
		Offset: 0,
	}

	// Parse query parameters
	if statusStr := c.Query("status"); statusStr != "" {
		filter.Status = []string{statusStr}
	}

	if sourceStr := c.Query("energy_source"); sourceStr != "" {
		filter.EnergySource = domain.EnergySource(sourceStr)
	}

	if registry := c.Query("registry"); registry != "" {
		filter.Registry = registry
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	certificates, err := h.energyService.ListCertificates(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   certificates,
		"total":  len(certificates),
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetCertificate handles GET /energy/certificates/:id
func (h *EnergyHandler) GetCertificate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
