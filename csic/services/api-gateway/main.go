package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic-platform/services/api-gateway/router"
	"github.com/csic-platform/services/api-gateway/middleware"
	"github.com/csic-platform/services/api-gateway/auth"
	"github.com/csic-platform/shared/config"
	"github.com/csic-platform/shared/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "internal/config/config.yaml", "Path to configuration file")
	flag.Parse()

	// Initialize configuration loader
	configLoader := config.NewConfigLoader(*configPath)
	cfg, err := configLoader.Load()
	if err != nil {
		// Fallback to default configuration if file loading fails
		cfg = getDefaultConfig()
		fmt.Printf("Warning: Failed to load config file, using defaults: %v\n", err)
	}

	// Initialize logger using shared logger package
	appLogger, err := logger.NewLogger(logger.Config{
		ServiceName:  "api-gateway",
		LogLevel:     cfg.Logging.Level,
		OutputPath:   cfg.Logging.Output,
		AuditLogPath: cfg.Logging.Path,
		Development:  cfg.App.Environment == "development",
		JSONOutput:   cfg.Logging.Format == "json",
	})
	if err != nil {
		fmt.Printf("Fatal: Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer appLogger.Close()

	// Log startup
	appLogger.Info("starting API gateway",
		logger.WithFields(
			logger.String("port", fmt.Sprintf("%d", cfg.Server.Port)),
			logger.String("environment", cfg.App.Environment),
		),
	)

	// Initialize services with shared logger
	authService := auth.NewAuthService(cfg.Security.JWT.Secret)
	authMiddleware := middleware.NewAuthMiddleware(appLogger, cfg.Security.JWT.Secret)
	rateLimiter := middleware.NewRateLimitMiddleware(appLogger)
	loggingMiddleware := middleware.NewLoggingMiddleware(appLogger)

	// Create router
	apiRouter := router.NewAPIRouter(authMiddleware, rateLimiter, loggingMiddleware)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	ginRouter := gin.New()

	// Apply global middleware
	ginRouter.Use(gin.Recovery())
	ginRouter.Use(middleware.NewSecurityHeadersMiddleware().Headers)

	// Setup API routes
	apiRouter.SetupRoutes(ginRouter)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      ginRouter,
		ReadTimeout:  cfg.Server.GetReadTimeout(),
		WriteTimeout: cfg.Server.GetWriteTimeout(),
		IdleTimeout:  cfg.Server.GetIdleTimeout(),
	}

	// Start server in goroutine
	go func() {
		appLogger.Info("API gateway listening",
			logger.WithFields(
				logger.Int("port", cfg.Server.Port),
			),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("server error",
				logger.WithFields(
					logger.Error(err),
				),
			)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down server")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Error("server forced shutdown",
			logger.WithFields(
				logger.Error(err),
			),
		)
	}

	appLogger.Info("server exited gracefully")
}

func getDefaultConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "csic-api-gateway",
			Version:     "1.0.0",
			Environment: "development",
		},
		Server: config.ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30,
			WriteTimeout: 30,
			MaxBodySize:  10485760,
		},
		Security: config.SecurityConfig{
			JWT: config.JWTConfig{
				Secret:      "default-secret-change-in-production",
				ExpiryHours: 8,
				Algorithm:   "HS256",
			},
		},
		Logging: config.LoggingConfig{
			Level:   "INFO",
			Format:  "json",
			Output:  "stdout",
		},
	}
}
