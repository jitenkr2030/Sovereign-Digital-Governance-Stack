package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic-platform/services/mining/internal/adapters/handler/http"
	"github.com/csic-platform/services/mining/internal/adapters/repository/postgres"
	"github.com/csic-platform/services/mining/internal/core/ports"
	"github.com/csic-platform/services/mining/internal/core/services"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// Initialize configuration
	if err := initConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Initialize database connection
	dbConnection, err := postgres.NewConnection(logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbConnection.Close()

	// Initialize repository
	repo := postgres.NewRepository(dbConnection, logger)

	// Initialize service
	miningService := services.NewMiningService(repo, logger)

	// Initialize handlers
	handlers := http.NewHandlers(miningService, logger)

	// Initialize router
	router := http.NewRouter(handlers, logger)

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", viper.GetString("app.host"), viper.GetInt("app.port")),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Starting Mining Control Platform",
			zap.String("host", viper.GetString("app.host")),
			zap.Int("port", viper.GetInt("app.port")),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/mining-api/")

	// Set defaults
	viper.SetDefault("app.host", "0.0.0.0")
	viper.SetDefault("app.port", 8080)
	viper.SetDefault("database.host", "postgres")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("compliance.violation_threshold", 3)

	// Environment variable overrides
	viper.AutomaticEnv()

	return viper.ReadInConfig()
}

func initLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	if viper.GetBool("app.debug") {
		config = zap.NewDevelopmentConfig()
	}
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	return config.Build()
}

// Interface for dependency injection
var _ ports.MiningRepository = (*postgres.Repository)(nil)
