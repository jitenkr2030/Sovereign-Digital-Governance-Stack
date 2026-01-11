package http

import (
	"time"

	"github.com/csic-platform/services/transaction-monitoring/internal/adapters/handler/http/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Router configures HTTP routes for the service
type Router struct {
	handlers *Handlers
	logger   *zap.Logger
}

// NewRouter creates a new router instance
func NewRouter(handlers *Handlers, logger *zap.Logger) *Router {
	return &Router{
		handlers: handlers,
		logger:   logger,
	}
}

// Setup configures all routes
func (r *Router) Setup() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(r.logger))
	router.Use(middleware.Cors())
	router.Use(middleware.Metrics())

	// Health check
	router.GET("/health", r.handlers.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Transaction analysis
		transactions := v1.Group("/transactions")
		{
			transactions.POST("/analyze", r.handlers.AnalyzeTransaction)
			transactions.GET("", r.handlers.GetTransactions)
		}

		// Wallet profiling
		wallets := v1.Group("/wallets")
		{
			wallets.GET("/:address", r.handlers.GetWalletProfile)
			wallets.GET("/:address/transactions", r.handlers.GetWalletTransactions)
			wallets.GET("/:address/sanctions", r.handlers.CheckSanctions)
		}

		// Alert management
		alerts := v1.Group("/alerts")
		{
			alerts.GET("", r.handlers.GetAlerts)
			alerts.GET("/stats", r.handlers.GetAlertStats)
			alerts.PUT("/:id/resolve", r.handlers.ResolveAlert)
		}

		// Monitoring rules
		rules := v1.Group("/rules")
		{
			rules.GET("", r.handlers.GetMonitoringRules)
			rules.POST("", r.handlers.CreateMonitoringRule)
		}

		// Sanctions list
		sanctions := v1.Group("/sanctions")
		{
			sanctions.POST("/check", r.handlers.CheckSanctions)
			sanctions.POST("/import", r.handlers.ImportSanctions)
		}

		// Statistics
		v1.GET("/stats", r.handlers.GetMonitoringStats)
	}

	return router
}

// ConfigureCors configures CORS middleware
func ConfigureCors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
