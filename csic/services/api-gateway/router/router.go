// CSIC Platform - API Gateway Router Configuration
// Main router with all API endpoints

package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"csic-platform/services/api-gateway/middleware"
	"csic-platform/services/api-gateway/auth"
	"csic-platform/shared/logger"
)

// Router interface for API routing
type Router interface {
    SetupRoutes(r *gin.Engine)
}

// APIRouter handles all API route configurations
type APIRouter struct {
	authMiddleware   *middleware.AuthMiddleware
	rateLimiter     *middleware.RateLimitMiddleware
	loggingMiddleware *middleware.LoggingMiddleware
	logger           *logger.Logger
}

// NewAPIRouter creates a new API router instance
func NewAPIRouter(
	authMiddleware *middleware.AuthMiddleware,
	rateLimiter *middleware.RateLimitMiddleware,
	loggingMiddleware *middleware.LoggingMiddleware,
) *APIRouter {
	return &APIRouter{
		authMiddleware:   authMiddleware,
		rateLimiter:     rateLimiter,
		loggingMiddleware: loggingMiddleware,
		logger:           nil,
	}
}

// SetupRoutes configures all API routes
func (r *APIRouter) SetupRoutes(router *gin.Engine) {
	// Health check endpoints (public)
	r.setupHealthRoutes(router)

	// API v1 routes (protected)
	v1 := router.Group("/api/v1")
	{
		// Apply authentication middleware
		v1.Use(r.authMiddleware.Authenticate())
		v1.Use(r.loggingMiddleware.Log)

		// Apply rate limiting
		v1.Use(r.rateLimiter.Limit(100, time.Minute))
        
        // Core control layer routes
        r.setupCoreControlRoutes(v1)
        
        // Exchange oversight routes
        r.setupExchangeRoutes(v1)
        
        // Wallet governance routes
        r.setupWalletRoutes(v1)
        
        // Transaction monitoring routes
        r.setupTransactionRoutes(v1)
        
        // Licensing and compliance routes
        r.setupLicenseRoutes(v1)
        
        // Mining control routes
        r.setupMiningRoutes(v1)
        
        // Energy integration routes
        r.setupEnergyRoutes(v1)
        
        // Reporting routes
        r.setupReportingRoutes(v1)
        
        // Security and audit routes
        r.setupSecurityRoutes(v1)
    }
    
    // gRPC proxy routes
    r.setupGRPCRoutes(router)
}

// setupHealthRoutes configures health check endpoints
func (r *APIRouter) setupHealthRoutes(router *gin.Engine) {
    health := router.Group("/health")
    {
        health.GET("", r.healthCheck)
        health.GET("/live", r.livenessCheck)
        health.GET("/ready", r.readinessCheck)
    }
}

// setupCoreControlRoutes configures core control layer routes
func (r *APIRouter) setupCoreControlRoutes(group *gin.RouterGroup) {
    core := group.Group("/core")
    {
        // System status
        core.GET("/status", r.getSystemStatus)
        core.GET("/metrics", r.getSystemMetrics)
        
        // Emergency operations (requires ADMIN role)
        emergency := core.Group("/emergency")
        emergency.Use(r.authMiddleware.RequireRole("ADMIN", "OPERATOR"))
        {
            emergency.POST("/stop", r.emergencyStop)
            emergency.POST("/resume", r.emergencyResume)
        }
        
        // Policy management (requires ADMIN role)
        policy := core.Group("/policy")
        policy.Use(r.authMiddleware.RequireRole("ADMIN"))
        {
            policy.GET("", r.getPolicies)
            policy.POST("", r.createPolicy)
            policy.PUT("/:id", r.updatePolicy)
            policy.DELETE("/:id", r.deletePolicy)
        }
    }
}

// setupExchangeRoutes configures exchange oversight routes
func (r *APIRouter) setupExchangeRoutes(group *gin.RouterGroup) {
    exchanges := group.Group("/exchanges")
    {
        // Public exchange listing
        exchanges.GET("", r.getExchanges)
        exchanges.GET("/:id", r.getExchangeDetails)
        
        // Exchange management (requires OPERATOR role)
        exchangeMgmt := exchanges.Group("/:id")
        exchangeMgmt.Use(r.authMiddleware.RequireRole("ADMIN", "OPERATOR"))
        {
            exchangeMgmt.POST("/freeze", r.freezeExchange)
            exchangeMgmt.POST("/thaw", r.thawExchange)
            exchangeMgmt.GET("/metrics", r.getExchangeMetrics)
            exchangeMgmt.GET("/orders", r.getOrderBook)
            exchangeMgmt.GET("/trades", r.getTradeHistory)
        }
        
        // Exchange reporting
        exchanges.GET("/:id/reports", r.getExchangeReports)
    }
}

// setupWalletRoutes configures wallet governance routes
func (r *APIRouter) setupWalletRoutes(group *gin.RouterGroup) {
    wallets := group.Group("/wallets")
    {
        // Wallet queries
        wallets.GET("", r.getWallets)
        wallets.GET("/:id", r.getWalletDetails)
        wallets.GET("/:id/transactions", r.getWalletTransactions)
        wallets.GET("/:id/risk-score", r.getWalletRiskScore)
        
        // Wallet operations (requires OPERATOR role)
        walletOps := wallets.Group("/:id")
        walletOps.Use(r.authMiddleware.RequireRole("ADMIN", "OPERATOR"))
        {
            walletOps.POST("/freeze", r.freezeWallet)
            walletOps.POST("/unfreeze", r.unfreezeWallet)
            walletOps.POST("/transfer", r.transferFromWallet)
            walletOps.POST("/recover", r.recoverWallet)
        }
        
        // Multisig wallet operations
        multisig := wallets.Group("/multisig")
        multisig.Use(r.authMiddleware.RequireRole("ADMIN", "OPERATOR"))
        {
            multisig.POST("/create", r.createMultisigWallet)
            multisig.POST("/:id/sign", r.signTransaction)
            multisig.POST("/:id/execute", r.executeTransaction)
        }
    }
}

// setupTransactionRoutes configures transaction monitoring routes
func (r *APIRouter) setupTransactionRoutes(group *gin.RouterGroup) {
    transactions := group.Group("/transactions")
    {
        // Transaction queries
        transactions.GET("", r.getTransactions)
        transactions.GET("/:id", r.getTransactionDetails)
        transactions.GET("/search", r.searchTransactions)
        
        // Transaction flags (requires OPERATOR role)
        flags := transactions.Group("/flag")
        flags.Use(r.authMiddleware.RequireRole("ADMIN", "OPERATOR", "ANALYST"))
        {
            flags.POST("", r.flagTransaction)
            flags.DELETE("/:id", r.unflagTransaction)
        }
        
        // Risk analysis
        transactions.GET("/risk/analysis", r.analyzeTransactionRisk)
        
        // Blockchain explorer integration
        transactions.GET("/txid/:txid", r.getTransactionByTxID)
    }
}

// setupLicenseRoutes configures licensing and compliance routes
func (r *APIRouter) setupLicenseRoutes(group *gin.RouterGroup) {
    licenses := group.Group("/licenses")
    {
        // License queries
        licenses.GET("", r.getLicenses)
        licenses.GET("/:id", r.getLicenseDetails)
        licenses.GET("/entity/:entity_id", r.getEntityLicense)
        licenses.GET("/expiring", r.getExpiringLicenses)
        
        // License management (requires ADMIN role)
        licenseMgmt := licenses.Group("")
        licenseMgmt.Use(r.authMiddleware.RequireRole("ADMIN"))
        {
            licenseMgmt.POST("", r.createLicense)
            licenseMgmt.PUT("/:id", r.updateLicense)
            licenseMgmt.POST("/:id/revoke", r.revokeLicense)
            licenseMgmt.POST("/:id/suspend", r.suspendLicense)
            licenseMgmt.POST("/:id/renew", r.renewLicense)
        }
        
        // Compliance obligations
        licenses.GET("/:id/obligations", r.getLicenseObligations)
    }
}

// setupMiningRoutes configures mining control routes
func (r *APIRouter) setupMiningRoutes(group *gin.RouterGroup) {
    miners := group.Group("/miners")
    {
        // Miner queries
        miners.GET("", r.getMiners)
        miners.GET("/:id", r.getMinerDetails)
        miners.GET("/:id/telemetry", r.getMinerTelemetry)
        miners.GET("/metrics/summary", r.getMiningSummary)
        
        // Miner control (requires OPERATOR role)
        minerControl := miners.Group("/:id")
        minerControl.Use(r.authMiddleware.RequireRole("ADMIN", "OPERATOR"))
        {
            minerControl.POST("/shutdown", r.shutdownMiner)
            minerControl.POST("/start", r.startMiner)
            minerControl.POST("/throttle", r.throttleMiner)
            minerControl.POST("/config", r.updateMinerConfig)
        }
        
        // Mining registry
        miners.GET("/registry", r.getMiningRegistry)
        miners.POST("/register", r.registerMiner)
    }
}

// setupEnergyRoutes configures energy integration routes
func (r *APIRouter) setupEnergyRoutes(group *gin.RouterGroup) {
    energy := group.Group("/energy")
    {
        // Energy queries
        energy.GET("/grid", r.getGridStatus)
        energy.GET("/consumption", r.getEnergyConsumption)
        energy.GET("/correlation", r.getEnergyCorrelation)
        
        // Energy control (requires ADMIN role)
        energyControl := energy.Group("/control")
        energyControl.Use(r.authMiddleware.RequireRole("ADMIN"))
        {
            energyControl.POST("/load-shedding", r.triggerLoadShedding)
            energyControl.POST("/allocation", r.updateEnergyAllocation)
        }
        
        // Energy reports
        energy.GET("/reports", r.getEnergyReports)
    }
}

// setupReportingRoutes configures reporting routes
func (r *APIRouter) setupReportingRoutes(group *gin.RouterGroup) {
    reports := group.Group("/reports")
    {
        // Report queries
        reports.GET("", r.getReports)
        reports.GET("/:id", r.getReportDetails)
        reports.GET("/types", r.getReportTypes)
        
        // Report generation (requires appropriate role)
        reportGen := reports.Group("")
        reportGen.Use(r.authMiddleware.RequireRole("ADMIN", "OPERATOR", "ANALYST"))
        {
            reportGen.POST("/generate", r.generateReport)
            reportGen.POST("/schedule", r.scheduleReport)
        }
        
        // Report download
        reports.GET("/:id/download", r.downloadReport)
        
        // Audit-specific reports
        reports.GET("/audit/summary", r.getAuditSummary)
        reports.GET("/audit/timeline", r.getAuditTimeline)
    }
}

// setupSecurityRoutes configures security and audit routes
func (r *APIRouter) setupSecurityRoutes(group *gin.RouterGroup) {
    security := group.Group("/security")
    {
        // HSM status
        security.GET("/hsm/status", r.getHSMStatus)
        security.GET("/hsm/metrics", r.getHSMMetrics)
        
        // Audit logs
        audit := security.Group("/audit")
        {
            audit.GET("/logs", r.getAuditLogs)
            audit.GET("/logs/:id", r.getAuditLogDetails)
            audit.GET("/verify", r.verifyAuditChain)
            audit.POST("/export", r.exportAuditLogs)
        }
        
        // Key management (requires ADMIN role)
        keys := security.Group("/keys")
        keys.Use(r.authMiddleware.RequireRole("ADMIN"))
        {
            keys.GET("", r.getKeys)
            keys.POST("/rotate", r.rotateKeys)
            keys.POST("/generate", r.generateKey)
        }
        
        // Access control
        security.GET("/users", r.getUsers)
        security.GET("/roles", r.getRoles)
        security.GET("/permissions", r.getPermissions)
        
        // Incident response (requires ADMIN role)
        incidents := security.Group("/incidents")
        incidents.Use(r.authMiddleware.RequireRole("ADMIN"))
        {
            incidents.GET("", r.getIncidents)
            incidents.POST("", r.reportIncident)
            incidents.PUT("/:id", r.updateIncident)
        }
    }
}

// setupGRPCRoutes configures gRPC proxy routes
func (r *APIRouter) setupGRPCRoutes(router *gin.Engine) {
    router.Any("/grpc/*path", r.grpcProxy)
}

// Health check handlers
func (r *APIRouter) healthCheck(c *gin.Context) {
    c.JSON(200, gin.H{
        "status":    "healthy",
        "timestamp": time.Now().UTC(),
        "version":   "1.0.0",
    })
}

func (r *APIRouter) livenessCheck(c *gin.Context) {
    c.JSON(200, gin.H{"status": "alive"})
}

func (r *APIRouter) readinessCheck(c *gin.Context) {
    // Check all dependencies
    if !r.checkDependencies() {
        c.JSON(503, gin.H{
            "status": "not ready",
            "reason": "dependencies not available",
        })
        return
    }
    c.JSON(200, gin.H{"status": "ready"})
}

func (r *APIRouter) checkDependencies() bool {
	// Implement dependency checks
	return true
}
