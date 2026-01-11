package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/csic/mining-control/internal/domain"
	"github.com/csic/mining-control/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// HTTPHandler handles HTTP requests for the mining control service
type HTTPHandler struct {
	registrationSvc *service.RegistrationService
	monitoringSvc   *service.MonitoringService
	enforcementSvc  *service.EnforcementService
	reportingSvc    *service.ReportingService
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(registrationSvc *service.RegistrationService, monitoringSvc *service.MonitoringService, enforcementSvc *service.EnforcementService, reportingSvc *service.ReportingService) *HTTPHandler {
	return &HTTPHandler{
		registrationSvc: registrationSvc,
		monitoringSvc:   monitoringSvc,
		enforcementSvc:  enforcementSvc,
		reportingSvc:    reportingSvc,
	}
}

// Pool Registration Handlers

// RegisterMiningPool registers a new mining pool
func (h *HTTPHandler) RegisterMiningPool(c *gin.Context) {
	var req domain.PoolRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	pool, err := h.registrationSvc.RegisterMiningPool(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to register mining pool",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, pool)
}

// GetMiningPool retrieves a mining pool by ID
func (h *HTTPHandler) GetMiningPool(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pool ID format",
		})
		return
	}

	pool, err := h.registrationSvc.GetMiningPool(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Mining pool not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve mining pool",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, pool)
}

// UpdateMiningPool updates a mining pool
func (h *HTTPHandler) UpdateMiningPool(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pool ID format",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	pool, err := h.registrationSvc.UpdateMiningPool(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update mining pool",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, pool)
}

// SuspendMiningPool suspends a mining pool
func (h *HTTPHandler) SuspendMiningPool(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pool ID format",
		})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Reason is required",
			"details": err.Error(),
		})
		return
	}

	pool, err := h.registrationSvc.SuspendMiningPool(c.Request.Context(), id, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to suspend mining pool",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mining pool suspended",
		"pool":    pool,
	})
}

// ActivateMiningPool activates a suspended mining pool
func (h *HTTPHandler) ActivateMiningPool(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pool ID format",
		})
		return
	}

	pool, err := h.registrationSvc.ActivateMiningPool(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to activate mining pool",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mining pool activated",
		"pool":    pool,
	})
}

// RevokeMiningPoolLicense revokes a mining pool license
func (h *HTTPHandler) RevokeMiningPoolLicense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pool ID format",
		})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Reason is required",
			"details": err.Error(),
		})
		return
	}

	if err := h.registrationSvc.RevokeMiningPoolLicense(c.Request.Context(), id, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to revoke mining pool license",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mining pool license revoked Machine Registration Handlers",
	})
}

//

// RegisterMiningMachine registers a new mining machine
func (h *HTTPHandler) RegisterMiningMachine(c *gin.Context) {
	var req domain.MachineRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	poolIDStr := c.Query("pool_id")
	if poolIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "pool_id query parameter is required",
		})
		return
	}

	poolID, err := uuid.Parse(poolIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pool_id format",
		})
		return
	}

	machine, err := h.registrationSvc.RegisterMiningMachine(c.Request.Context(), poolID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to register mining machine",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, machine)
}

// BatchRegisterMachines registers multiple mining machines
func (h *HTTPHandler) BatchRegisterMachines(c *gin.Context) {
	var reqs []domain.MachineRegistrationRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	poolIDStr := c.Query("pool_id")
	if poolIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "pool_id query parameter is required",
		})
		return
	}

	poolID, err := uuid.Parse(poolIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pool_id format",
		})
		return
	}

	machines, err := h.registrationSvc.BatchRegisterMachines(c.Request.Context(), poolID, reqs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to batch register mining machines",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"count":    len(machines),
		"machines": machines,
	})
}

// GetMiningMachine retrieves a mining machine by ID
func (h *HTTPHandler) GetMiningMachine(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid machine ID format",
		})
		return
	}

	machine, err := h.registrationSvc.GetMiningMachine(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Mining machine not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve mining machine",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, machine)
}

// GetPoolMachines retrieves all machines for a pool
func (h *HTTPHandler) GetPoolMachines(c *gin.Context) {
	poolIDStr := c.Param("pool_id")
	poolID, err := uuid.Parse(poolIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pool ID format",
		})
		return
	}

	activeOnly, _ := strconv.ParseBool(c.DefaultQuery("active_only", "false"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	machines, err := h.registrationSvc.GetPoolMachines(c.Request.Context(), poolID, activeOnly, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve pool machines",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":   len(machines),
		"machines": machines,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// UpdateMiningMachine updates a mining machine
func (h *HTTPHandler) UpdateMiningMachine(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid machine ID format",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	machine, err := h.registrationSvc.UpdateMiningMachine(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update mining machine",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, machine)
}

// DecommissionMachine decommission a mining machine
func (h *HTTPHandler) DecommissionMachine(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid machine ID format",
		})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Reason is required",
			"details": err.Error(),
		})
		return
	}

	if err := h.registrationSvc.DecommissionMachine(c.Request.Context(), id, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to decommission mining machine",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mining machine decommissioned",
	})
}

// Telemetry Handlers

// ReportEnergyConsumption reports energy consumption
func (h *HTTPHandler) ReportEnergyConsumption(c *gin.Context) {
	var packet domain.EnergyTelemetryPacket
	if err := c.ShouldBindJSON(&packet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	logEntry, err := h.monitoringSvc.ReportEnergyConsumption(c.Request.Context(), &packet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to record energy consumption",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, logEntry)
}

// BatchReportEnergyConsumption batch reports energy consumption
func (h *HTTPHandler) BatchReportEnergyConsumption(c *gin.Context) {
	var packets []domain.EnergyTelemetryPacket
	if err := c.ShouldBindJSON(&packets); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	logs, err := h.monitoringSvc.BatchReportEnergyConsumption(c.Request.Context(), packets)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to batch record energy consumption",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"count": len(logs),
		"logs":  logs,
	})
}

// ReportHashRate reports hash rate metrics
func (h *HTTPHandler) ReportHashRate(c *gin.Context) {
	var packet domain.HashRateTelemetryPacket
	if err := c.ShouldBindJSON(&packet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	metric, err := h.monitoringSvc.ReportHashRate(c.Request.Context(), &packet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to record hash rate",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, metric)
}

// BatchReportHashRate batch reports hash rate metrics
func (h *HTTPHandler) BatchReportHashRate(c *gin.Context) {
	var packets []domain.HashRateTelemetryPacket
	if err := c.ShouldBindJSON(&packets); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	metrics, err := h.monitoringSvc.BatchReportHashRate(c.Request.Context(), packets)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to batch record hash rate",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"count":   len(metrics),
		"metrics": metrics,
	})
}

// Monitoring Handlers

// GetEnergyUsage retrieves energy usage for a pool
func (h *HTTPHandler) GetEnergyUsage(c *gin.Context) {
	poolIDStr := c.Param("pool_id")
	poolID, err := uuid.Parse(poolIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pool ID format",
		})
		return
	}

	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -7) // Default 7 days

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	logs, err := h.monitoringSvc.GetEnergyUsage(c.Request.Context(), poolID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve energy usage",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pool_id":   poolID,
		"period":    gin.H{"start": startTime, "end": endTime},
		"count":     len(logs),
		"energy":    logs,
	})
}

// GetHashRateMetrics retrieves hash rate metrics for a pool
func (h *HTTPHandler) GetHashRateMetrics(c *gin.Context) {
	poolIDStr := c.Param("pool_id")
	poolID, err := uuid.Parse(poolIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pool ID format",
		})
		return
	}

	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -7) // Default 7 days

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	metrics, err := h.monitoringSvc.GetHashRateMetrics(c.Request.Context(), poolID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve hash rate metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pool_id":   poolID,
		"period":    gin.H{"start": startTime, "end": endTime},
		"count":     len(metrics),
		"hash_rate": metrics,
	})
}

// GetComplianceStatus retrieves compliance status for a pool
func (h *HTTPHandler) GetComplianceStatus(c *gin.Context) {
	poolIDStr := c.Param("pool_id")
	poolID, err := uuid.Parse(poolIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pool ID format",
		})
		return
	}

	status, err := h.monitoringSvc.GetComplianceStatus(c.Request.Context(), poolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve compliance status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// Violation Handlers

// ListViolations lists violations with filtering
func (h *HTTPHandler) ListViolations(c *gin.Context) {
	var filter domain.ViolationFilter

	if types := c.Query("types"); types != "" {
		for _, t := range splitString(types, ",") {
			filter.Types = append(filter.Types, domain.ViolationType(t))
		}
	}

	if severities := c.Query("severities"); severities != "" {
		for _, s := range splitString(severities, ",") {
			filter.Severities = append(filter.Severities, domain.ViolationSeverity(s))
		}
	}

	if statuses := c.Query("statuses"); statuses != "" {
		for _, s := range splitString(statuses, ",") {
			filter.Statuses = append(filter.Statuses, domain.ViolationStatus(s))
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Note: This would need to be implemented in the violation repository
	c.JSON(http.StatusOK, gin.H{
		"count":      0,
		"violations": []domain.ComplianceViolation{},
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetOpenViolations retrieves all open violations
func (h *HTTPHandler) GetOpenViolations(c *gin.Context) {
	poolIDStr := c.Query("pool_id")

	var violations []domain.ComplianceViolation

	if poolIDStr != "" {
		poolID, err := uuid.Parse(poolIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid pool ID format",
			})
			return
		}
		// Would call enforcement service
	} else {
		// Would get all open violations
	}

	c.JSON(http.StatusOK, gin.H{
		"count":      len(violations),
		"violations": violations,
	})
}

// GetViolation retrieves a violation by ID
func (h *HTTPHandler) GetViolation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid violation ID format",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

// ResolveViolation resolves a violation
func (h *HTTPHandler) ResolveViolation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid violation ID format",
		})
		return
	}

	var req struct {
		Resolution string    `json:"resolution" binding:"required"`
		UserID     uuid.UUID `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.enforcementSvc.ResolveViolation(c.Request.Context(), id, req.Resolution, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to resolve violation",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Violation resolved",
	})
}

// Reporting Handlers

// GetRegionalEnergyLoad retrieves regional energy load statistics
func (h *HTTPHandler) GetRegionalEnergyLoad(c *gin.Context) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, -1, 0) // Default 1 month

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	stats, err := h.reportingSvc.GetRegionalEnergyLoad(c.Request.Context(), startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve regional energy load",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period":     gin.H{"start": startTime, "end": endTime},
		"regions":    stats,
		"regionCount": len(stats),
	})
}

// GetCarbonFootprintReport retrieves carbon footprint report
func (h *HTTPHandler) GetCarbonFootprintReport(c *gin.Context) {
	poolIDStr := c.Param("pool_id")
	poolID, err := uuid.Parse(poolIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pool ID format",
		})
		return
	}

	endTime := time.Now()
	startTime := endTime.AddDate(0, -1, 0) // Default 1 month

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	report, err := h.reportingSvc.GetCarbonFootprintReport(c.Request.Context(), poolID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate carbon footprint report",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetMiningSummaryReport retrieves mining summary report
func (h *HTTPHandler) GetMiningSummaryReport(c *gin.Context) {
	poolIDStr := c.Param("pool_id")
	poolID, err := uuid.Parse(poolIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pool ID format",
		})
		return
	}

	endTime := time.Now()
	startTime := endTime.AddDate(0, -1, 0) // Default 1 month

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	report, err := h.reportingSvc.GetMiningSummaryReport(c.Request.Context(), poolID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate mining summary report",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetDashboardStats retrieves dashboard statistics
func (h *HTTPHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.reportingSvc.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve dashboard stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Compliance Certificate Handlers

// GetComplianceCertificate retrieves compliance certificate for a pool
func (h *HTTPHandler) GetComplianceCertificate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Compliance certificate endpoint - to be implemented",
	})
}

// RenewComplianceCertificate renews a compliance certificate
func (h *HTTPHandler) RenewComplianceCertificate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Renew compliance certificate endpoint - to be implemented",
	})
}

// Helper function to split a string by delimiter
func splitString(s string, delimiter string) []string {
	if s == "" {
		return nil
	}
	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == delimiter[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}
