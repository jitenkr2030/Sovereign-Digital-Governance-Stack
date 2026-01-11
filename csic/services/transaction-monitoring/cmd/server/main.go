package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic-platform/services/transaction-monitoring/internal/adapters/handler/http"
	"github.com/csic-platform/services/transaction-monitoring/internal/adapters/repository/postgres"
	"github.com/csic-platform/services/transaction-monitoring/internal/adapters/events/kafka"
	"github.com/csic-platform/services/transaction-monitoring/internal/core/ports"
	"github.com/csic-platform/services/transaction-monitoring/internal/core/services"
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

	// Initialize repositories
	transactionRepo := postgres.NewTransactionRepository(dbConnection, logger)
	walletProfileRepo := postgres.NewWalletProfileRepository(dbConnection, logger)
	sanctionsRepo := postgres.NewSanctionsRepository(dbConnection, logger)
	alertRepo := postgres.NewAlertRepository(dbConnection, logger)
	ruleRepo := postgres.NewMonitoringRuleRepository(dbConnection, logger)

	// Initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(logger)
	if err != nil {
		logger.Warn("Failed to initialize Kafka producer, continuing without messaging", zap.Error(err))
		kafkaProducer = nil
	}

	// Initialize services
	transactionService := services.NewTransactionAnalysisService(
		transactionRepo, walletProfileRepo, sanctionsRepo, ruleRepo, logger,
	)
	walletService := services.NewWalletProfilingService(walletProfileRepo, transactionRepo, logger)
	riskService := services.NewRiskScoringService(walletProfileRepo, transactionRepo, ruleRepo, logger)
	alertService := services.NewAlertService(alertRepo, kafkaProducer, logger)
	ruleService := services.NewRuleEngineService(ruleRepo, logger)

	// Initialize handlers
	handlers := http.NewHandlers(
		transactionService, walletService, riskService, alertService, ruleService, logger,
	)

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

	// Start Kafka consumer for transaction stream
	if kafkaConsumer, err := kafka.NewConsumer(kafkaProducer, logger); err == nil {
		go kafkaConsumer.StartConsuming(transactionService)
	}

	// Graceful shutdown
	go func() {
		logger.Info("Starting Transaction Monitoring Service",
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
	viper.AddConfigPath("/etc/transaction-monitoring/")

	// Set defaults
	viper.SetDefault("app.host", "0.0.0.0")
	viper.SetDefault("app.port", 8085)
	viper.SetDefault("database.host", "postgres")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("kafka.brokers", "localhost:9092")
	viper.SetDefault("monitoring.risk_threshold_high", 75)
	viper.SetDefault("monitoring.risk_threshold_medium", 50)
	viper.SetDefault("monitoring.max_transaction_value", 1000000.0)

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

// Interface verification
var _ ports.TransactionRepository = (*postgres.TransactionRepository)(nil)
var _ ports.WalletProfileRepository = (*postgres.WalletProfileRepository)(nil)
var _ ports.SanctionsRepository = (*postgres.SanctionsRepository)(nil)
var _ ports.AlertRepository = (*postgres.AlertRepository)(nil)
var _ ports.MonitoringRuleRepository = (*postgres.MonitoringRuleRepository)(nil)
