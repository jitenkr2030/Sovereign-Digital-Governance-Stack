package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NewRouter creates a new Gin router with all routes configured
func NewRouter(handlers *Handlers, log *zap.Logger) *gin.Engine {
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(Logger(log))
	router.Use(RequestID())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// License Application routes
		applications := v1.Group("/licenses/applications")
		{
			applications.POST("", handlers.SubmitApplication)
			applications.GET("", handlers.ListApplications)
			applications.GET("/:id", handlers.GetApplication)
			applications.PATCH("/:id/review", handlers.ReviewApplication)
		}

		// License routes
		licenses := v1.Group("/licenses")
		{
			licenses.POST("", handlers.IssueLicense)
			licenses.GET("/:id", handlers.GetLicense)
			licenses.POST("/:id/suspend", handlers.SuspendLicense)
			licenses.POST("/:id/revoke", handlers.RevokeLicense)
			licenses.GET("/expiring", handlers.GetExpiringLicenses)
		}

		// Entity routes
		entities := v1.Group("/entities")
		{
			entities.POST("", handlers.RegisterEntity)
			entities.GET("/:id", handlers.GetEntity)
			entities.GET("/:id/licenses", handlers.GetEntityLicenses)
			entities.GET("/:id/obligations", handlers.GetEntityObligations)
			entities.GET("/:id/compliance/score", handlers.GetComplianceScore)
			entities.POST("/:id/compliance/score/recalculate", handlers.RecalculateScore)
			entities.GET("/:id/compliance/score/history", handlers.GetScoreHistory)
			entities.GET("/:id/audit-trail", handlers.GetEntityAuditTrail)
		}

		// Compliance routes
		compliance := v1.Group("/compliance")
		{
			compliance.GET("/stats", handlers.GetComplianceStats)
		}

		// Obligation routes
		obligations := v1.Group("/obligations")
		{
			obligations.POST("", handlers.CreateObligation)
			obligations.GET("/:id", handlers.GetObligation)
			obligations.GET("/overdue", handlers.GetOverdueObligations)
			obligations.POST("/:id/fulfill", handlers.FulfillObligation)
			obligations.POST("/check-overdue", handlers.CheckOverdueObligations)
		}

		// Audit routes
		audit := v1.Group("/audit-logs")
		{
			audit.GET("", handlers.GetAuditLogs)
		}
	}

	return router
}
