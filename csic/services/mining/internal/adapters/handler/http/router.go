package http

import (
	"time"

	"github.com/csic-platform/services/services/mining/internal/adapters/handler/http/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NewRouter creates a new Gin router with all routes configured
func NewRouter(handlers *Handlers, log *zap.Logger) *gin.Engine {
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(log))
	router.Use(middleware.RequestID())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Registry endpoints
		operations := v1.Group("/operations")
		{
			operations.POST("", handlers.RegisterOperation)
			operations.GET("", handlers.ListOperations)
			operations.GET("/wallet/:wallet", handlers.GetOperationByWallet)
			operations.GET("/:id", handlers.GetOperation)
			operations.PUT("/:id/status", handlers.UpdateOperationStatus)
			operations.POST("/:id/telemetry", handlers.ReportHashrate)
			operations.GET("/:id/metrics", handlers.GetHashrateHistory)
			operations.GET("/:id/quota", handlers.GetQuota)
			operations.GET("/:id/commands", handlers.GetPendingCommands)
			operations.POST("/:id/shutdown", handlers.IssueShutdownCommand)
		}

		// Quota management endpoints
		quotas := v1.Group("/quotas")
		{
			quotas.POST("", handlers.AssignQuota)
			quotas.DELETE("/:id", handlers.RevokeQuota)
		}

		// Command endpoints
		commands := v1.Group("/commands")
		{
			commands.POST("/:id/ack", handlers.AcknowledgeCommand)
			commands.POST("/:id/confirm", handlers.ConfirmShutdown)
		}

		// Statistics endpoint
		v1.GET("/stats", handlers.GetRegistryStats)
	}

	return router
}
