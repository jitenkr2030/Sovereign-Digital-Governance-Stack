// Audit Log Service Entry Point
// Government-grade immutable audit logging service for CSIC Platform

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

	"github.com/csic-platform/services/audit-log/handlers"
	"github.com/csic-platform/shared/config"
	"github.com/csic-platform/shared/logger"
	_ "github.com/lib/pq"
	"github.com/gin-gonic/gin"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	configLoader := config.NewConfigLoader(*configPath)
	cfg, err := configLoader.Load()
	if err != nil {
		fmt.Printf("Warning: Failed to load config file, using defaults: %v\n", err)
		cfg = getDefaultConfig()
	}

	// Initialize audit log service
	auditConfig := &AuditConfig{
		StoragePath:      cfg.AuditLog.StoragePath,
		SealInterval:     cfg.AuditLog.SealInterval,
		ChainFilePath:    cfg.AuditLog.ChainFilePath,
		VerificationPath: cfg.AuditLog.VerificationPath,
		RetentionDays:    cfg.AuditLog.RetentionDays,
		EnableWORM:       cfg.AuditLog.EnableWORM,
	}

	logConfig := logger.Config{
		ServiceName:  "audit-log-service",
		LogLevel:     cfg.Logging.Level,
		OutputPath:   cfg.Logging.Output,
		AuditLogPath: cfg.Logging.Path,
		Development:  cfg.App.Environment == "development",
		JSONOutput:   cfg.Logging.Format == "json",
	}

	auditService, err := NewAuditLogService(auditConfig, logConfig)
	if err != nil {
		fmt.Printf("Fatal: Failed to initialize audit log service: %v\n", err)
		os.Exit(1)
	}
	defer auditService.Stop()

	// Start the service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := auditService.Start(ctx); err != nil {
		fmt.Printf("Fatal: Failed to start audit log service: %v\n", err)
		os.Exit(1)
	}

	// Initialize database connection for hybrid storage
	db, err := initDatabase(cfg.Database)
	if err != nil {
		fmt.Printf("Warning: Failed to connect to database: %v\n", err)
	} else {
		defer db.Close()
		auditService.SetDatabase(db)
	}

	// Initialize HTTP handlers
	httpHandler := handlers.NewAuditLogHandler(auditService)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(httpHandler.LoggingMiddleware())

	// Health check endpoints
	router.GET("/health", httpHandler.HealthCheck)
	router.GET("/ready", httpHandler.ReadinessCheck)

	// Audit log API endpoints
	api := router.Group("/api/v1/audit")
	{
		// Write endpoints
		api.POST("/entries", httpHandler.WriteEntry)
		api.POST("/entries/batch", httpHandler.WriteBatch)

		// Query endpoints
		api.GET("/entries", httpHandler.QueryEntries)
		api.GET("/entries/:id", httpHandler.GetEntry)

		// Verification endpoints
		api.GET("/verify", httpHandler.VerifyChain)
		api.GET("/verify/report", httpHandler.GetVerificationReport)
		api.GET("/chains", httpHandler.ListChains)
		api.GET("/chains/:id", httpHandler.GetChain)
		api.GET("/chains/:id/export", httpHandler.ExportChain)

		// Summary endpoints
		api.GET("/summary", httpHandler.GetSummary)
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
		fmt.Printf("Audit Log Service starting on port %d\n", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down Audit Log Service...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	fmt.Println("Audit Log Service exited")
}

// initDatabase initializes the PostgreSQL database connection
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

func getDefaultConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "csic-audit-log-service",
			Version:     "1.0.0",
			Environment: "development",
		},
		Server: config.ServerConfig{
			Host:         "0.0.0.0",
			Port:         8081,
			ReadTimeout:  30,
			WriteTimeout: 30,
			MaxBodySize:  10485760,
		},
		Logging: config.LoggingConfig{
			Level:   "INFO",
			Format:  "json",
			Output:  "stdout",
		},
		AuditLog: config.AuditLogConfig{
			StoragePath:      "/var/lib/csic/audit-logs",
			SealInterval:     300,
			ChainFilePath:    "/var/lib/csic/audit-chains/chain.json",
			VerificationPath: "/var/lib/csic/audit-verification",
			RetentionDays:    2555, // 7 years
			EnableWORM:       true,
		},
	}
}
