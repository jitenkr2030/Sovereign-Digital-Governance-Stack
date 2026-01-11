package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic-platform/services/reporting/internal/adapter/formatter"
	"github.com/csic-platform/services/reporting/internal/adapter/generator"
	"github.com/csic-platform/services/reporting/internal/adapter/repository"
	"github.com/csic-platform/services/reporting/internal/adapter/storage"
	"github.com/csic-platform/services/reporting/internal/core/service"
	"github.com/csic-platform/services/reporting/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds all configuration for the reporting service
type Config struct {
	AppHost      string `envconfig:"APP_HOST" default:"0.0.0.0"`
	AppPort      int    `envconfig:"APP_PORT" default:"8082"`
	DBHost       string `envconfig:"DB_HOST" default:"localhost"`
	DBPort       int    `envconfig:"DB_PORT" default:"5432"`
	DBUser       string `envconfig:"DB_USER" default:"csic"`
	DBPassword   string `envconfig:"DB_PASSWORD" default:"csic_secret"`
	DBName       string `envconfig:"DB_NAME" default:"csic_platform"`
	DBSSLMode    string `envconfig:"DB_SSLMODE" default:"disable"`
	StoragePath  string `envconfig:"STORAGE_PATH" default:"/app/reports"`
	KafkaBrokers string `envconfig:"KAFKA_BROKERS" default:"localhost:9092"`
	TopicPrefix  string `envconfig:"TOPIC_PREFIX" default:"csic"`
	LogLevel     string `envconfig:"LOG_LEVEL" default:"info"`
	Env          string `envconfig:"ENV" default:"development"`
}

func main() {
	// Initialize logger
	logger := initLogger()
	defer logger.Sync()

	logger.Info("Starting Reporting Service")

	// Load configuration (simplified - use envconfig in production)
	cfg := &Config{
		AppHost:      "0.0.0.0",
		AppPort:      8082,
		DBHost:       "localhost",
		DBPort:       5432,
		DBUser:       "csic",
		DBPassword:   "csic_secret",
		DBName:       "csic_platform",
		StoragePath:  "/app/reports",
		KafkaBrokers: "localhost:9092",
		TopicPrefix:  "csic",
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.dsn())
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Info("Connected to database")

	// Initialize storage
	storageAdapter, err := storage.NewLocalStorage(cfg.StoragePath, logger)
	if err != nil {
		logger.Fatal("Failed to initialize storage", zap.Error(err))
	}

	// Initialize repositories
	reportRepo := repository.NewPostgresReportRepository(db)
	templateRepo := repository.NewPostgresReportTemplateRepository(db)
	scheduledRepo := repository.NewPostgresScheduledReportRepository(db)
	statsRepo := repository.NewPostgresReportStatisticsRepository(db)

	// Initialize adapters
	reportGenerator := generator.NewReportGenerator(logger)
	reportFormatter := formatter.NewReportFormatter(logger)

	// Initialize services
	reportService := service.NewReportService(
		reportRepo,
		reportGenerator,
		reportFormatter,
		storageAdapter,
		statsRepo,
		nil, // metrics
		logger,
	)

	templateService := service.NewTemplateService(templateRepo, logger)
	schedulerService := service.NewSchedulerService(scheduledRepo, reportService, logger)

	// Initialize HTTP handler
	httpHandler := handler.NewHTTPHandler(
		reportService,
		templateService,
		schedulerService,
		logger,
	)

	// Setup router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.Timeout(30 * time.Second))

	httpHandler.RegisterRoutes(router)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.AppHost, cfg.AppPort),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("HTTP server starting",
			zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown of HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	logger.Info("Server stopped")
}

func (c *Config) dsn() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

func initLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	return logger
}
