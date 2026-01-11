// NEAM Industrial Adapter - Main Entry Point
// OPC UA client for industrial automation data ingestion

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"neam-platform/sensing/industrial/config"
	"neam-platform/sensing/industrial/metrics"
	"go.uber.org/zap"
)

// IndustrialAdapter represents the main industrial adapter service
type IndustrialAdapter struct {
	cfg           *config.Config
	logger        *zap.Logger
	metricsServer *metrics.Server
}

// NewIndustrialAdapter creates a new industrial adapter instance
func NewIndustrialAdapter(cfg *config.Config, logger *zap.Logger) (*IndustrialAdapter, error) {
	adapter := &IndustrialAdapter{
		cfg:    cfg,
		logger: logger,
	}

	// Initialize metrics server if enabled
	if cfg.Monitoring.Enabled {
		adapter.metricsServer = metrics.NewServer(metrics.ServerConfig{
			Port:       cfg.Monitoring.Port,
			Path:       cfg.Monitoring.Path,
			HealthPath: cfg.Monitoring.HealthPath,
		}, logger)
	}

	return adapter, nil
}

// Start begins the industrial adapter service
func (a *IndustrialAdapter) Start(ctx context.Context) error {
	a.logger.Info("Starting industrial adapter service",
		zap.String("name", a.cfg.Service.Name),
		zap.Int("port", a.cfg.Service.Port))

	// Start metrics server
	if a.metricsServer != nil {
		if err := a.metricsServer.Start(); err != nil {
			return fmt.Errorf("failed to start metrics server: %w", err)
		}
	}

	a.logger.Info("Industrial adapter service started successfully")
	return nil
}

// Stop gracefully shuts down the adapter
func (a *IndustrialAdapter) Stop() error {
	a.logger.Info("Stopping industrial adapter service")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if a.metricsServer != nil {
		if err := a.metricsServer.Stop(ctx); err != nil {
			a.logger.Error("Error stopping metrics server", zap.Error(err))
		}
	}

	a.logger.Info("Industrial adapter service stopped")
	return nil
}

// Health returns the health status of the adapter
func (a *IndustrialAdapter) Health() map[string]interface{} {
	return map[string]interface{}{
		"status":      "healthy",
		"service":     a.cfg.Service.Name,
		"discovery":   a.cfg.Discovery.Enabled,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Load configuration
	cfgPath := "config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		cfgPath = envPath
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	logger.Info("Configuration loaded",
		zap.String("service", cfg.Service.Name),
		zap.Int("port", cfg.Service.Port),
		zap.Bool("discovery_enabled", cfg.Discovery.Enabled))

	// Create adapter
	adapter, err := NewIndustrialAdapter(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create industrial adapter", zap.Error(err))
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutdown signal received")
		cancel()
	}()

	// Start adapter
	if err := adapter.Start(ctx); err != nil {
		logger.Fatal("Failed to start adapter", zap.Error(err))
	}

	// Wait for shutdown
	<-ctx.Done()

	if err := adapter.Stop(); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}
}
