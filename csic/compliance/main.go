// Compliance Management Service - Main Entry Point
// Government-grade regulatory compliance management service

package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic-platform/compliance/internal/domain"
	"github.com/csic-platform/compliance/internal/handler"
	"github.com/csic-platform/compliance/internal/port"
	"github.com/csic-platform/compliance/internal/repository"
	"github.com/csic-platform/compliance/internal/service"
	"github.com/csic-platform/shared/config"
	"github.com/csic-platform/shared/logger"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "internal/config/config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	configLoader := config.NewConfigLoader(*configPath)
	cfg, err := configLoader.Load()
	if err != nil {
		fmt.Printf("Warning: Failed to load config file, using defaults: %v\n", err)
		cfg = getDefaultConfig()
	}

	// Initialize logger
	logConfig := logger.Config{
		ServiceName:  "compliance-service",
		LogLevel:     cfg.Logging.Level,
		OutputPath:   cfg.Logging.Output,
		AuditLogPath: cfg.Logging.Path,
		Development:  cfg.App.Environment == "development",
		JSONOutput:   cfg.Logging.Format == "json",
	}

	appLogger, err := logger.NewLogger(logConfig)
	if err != nil {
		fmt.Printf("Fatal: Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer appLogger.Close()

	// Initialize database connection
	db, err := initDatabase(cfg.Database)
	if err != nil {
		appLogger.Fatal("failed to connect to database", logger.WithFields(logger.Error(err)))
	}
	defer db.Close()

	// Initialize repositories
	entityRepo := repository.NewPostgresRepository(db)
	licenseRepo := repository.NewPostgresRepository(db)
	obligationRepo := repository.NewPostgresRepository(db)
	violationRepo := repository.NewPostgresRepository(db)
	penaltyRepo := repository.NewPostgresRepository(db)

	// Initialize audit client
	auditClient := NewAuditClient(cfg.AuditLog.ServiceURL, appLogger)

	// Initialize services
	entityService := service.NewEntityService(entityRepo, auditClient)
	licensingService := service.NewLicensingService(licenseRepo, entityRepo, auditClient)
	obligationService := service.NewObligationService(obligationRepo, auditClient)
	violationService := service.NewViolationService(violationRepo, penaltyRepo, entityRepo, auditClient)

	// Initialize HTTP handler
	complianceHandler := handler.NewComplianceHandler(
		entityService,
		licensingService,
		obligationService,
		violationService,
	)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(LoggingMiddleware(appLogger))

	// Health check endpoints
	router.GET("/health", complianceHandler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1/compliance")
	{
		// Entity management
		entities := v1.Group("/entities")
		{
			entities.POST("", complianceHandler.CreateEntity)
			entities.GET("", complianceHandler.ListEntities)
			entities.GET("/:id", complianceHandler.GetEntity)
			entities.PUT("/:id", complianceHandler.UpdateEntity)
			entities.POST("/:id/activate", complianceHandler.ActivateEntity)
			entities.POST("/:id/suspend", complianceHandler.SuspendEntity)
			entities.GET("/:id/violations", complianceHandler.GetOpenViolations)
		}

		// License management
		licenses := v1.Group("/licenses")
		{
			licenses.POST("", complianceHandler.CreateLicense)
			licenses.GET("", complianceHandler.ListLicenses)
			licenses.GET("/:id", complianceHandler.GetLicense)
			licenses.POST("/:id/approve", complianceHandler.ApproveLicense)
			licenses.POST("/:id/suspend", complianceHandler.SuspendLicense)
			licenses.POST("/:id/revoke", complianceHandler.RevokeLicense)
		}

		// Obligation management
		obligations := v1.Group("/obligations")
		{
			obligations.POST("", complianceHandler.CreateObligation)
			obligations.GET("", complianceHandler.ListObligations)
			obligations.GET("/:id", complianceHandler.GetObligation)
			obligations.POST("/:id/verify", complianceHandler.VerifyObligation)
			obligations.GET("/overdue", complianceHandler.GetOverdueObligations)
		}

		// Violation management
		violations := v1.Group("/violations")
		{
			violations.POST("", complianceHandler.CreateViolation)
			violations.GET("", complianceHandler.ListViolations)
			violations.GET("/:id", complianceHandler.GetViolation)
			violations.POST("/:id/penalty", complianceHandler.IssuePenalty)
			violations.POST("/:id/resolve", complianceHandler.ResolveViolation)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		appLogger.Info("compliance service starting",
			logger.WithFields(logger.Int("port", cfg.Server.Port)),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("server error", logger.WithFields(logger.Error(err)))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down compliance service")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Error("server forced shutdown", logger.WithFields(logger.Error(err)))
	}

	appLogger.Info("compliance service exited")
}

// initDatabase initializes the database connection
func initDatabase(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	return db, nil
}

// AuditClient implements port.AuditLogPort for audit logging
type AuditClient struct {
	baseURL string
	logger  *logger.Logger
}

// NewAuditClient creates a new audit client
func NewAuditClient(baseURL string, log *logger.Logger) *AuditClient {
	return &AuditClient{
		baseURL: baseURL,
		logger:  log,
	}
}

// Log sends an audit log entry
func (c *AuditClient) Log(ctx context.Context, entry *port.AuditEntry) error {
	c.logger.Info("audit log",
		logger.WithFields(
			logger.String("action", entry.Action),
			logger.String("resource_type", entry.ResourceType),
			logger.String("resource_id", entry.ResourceID),
			logger.String("entity_id", entry.EntityID),
		),
	)
	return nil
}

// LoggingMiddleware returns a gin middleware for logging
func LoggingMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Info("http request",
			logger.WithFields(
				logger.String("method", c.Request.Method),
				logger.String("path", path),
				logger.Int("status", status),
				logger.String("latency", latency.String()),
				logger.String("ip", c.ClientIP()),
			),
		)
	}
}

func getDefaultConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "csic-compliance-service",
			Version:     "1.0.0",
			Environment: "development",
		},
		Server: config.ServerConfig{
			Host:         "0.0.0.0",
			Port:         8082,
			ReadTimeout:  30,
			WriteTimeout: 30,
			MaxBodySize:  10485760,
		},
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			Username:        "csic_user",
			Password:        "change_in_production",
			Name:            "csic_compliance",
			SSLMode:         "require",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300,
		},
		Logging: config.LoggingConfig{
			Level:   "INFO",
			Format:  "json",
			Output:  "stdout",
		},
		AuditLog: config.AuditLogConfig{
			ServiceURL:    "http://localhost:8081",
			StoragePath:   "/var/lib/csic/compliance-audit",
			SealInterval:  300,
			EnableWORM:    true,
		},
	}
}
