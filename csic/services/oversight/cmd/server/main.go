package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpAdapter "github.com/csic/oversight/internal/adapters/handler/http"
	"github.com/csic/oversight/internal/adapters/messaging"
	postgresAdapter "github.com/csic/oversight/internal/adapters/repo/postgres"
	"github.com/csic/oversight/internal/config"
	"github.com/csic/oversight/internal/core/services"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting CSIC Exchange Oversight Service")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize PostgreSQL connection
	db, err := initDatabase(cfg.Postgres, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories
	repo := postgresAdapter.NewPostgresRepository(db, logger)
	if err := repo.InitSchema(context.Background()); err != nil {
		logger.Fatal("Failed to initialize database schema", zap.Error(err))
	}

	// Initialize Kafka connections
	kafkaWriter, kafkaReader, err := initKafka(cfg.Kafka, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Kafka connections", zap.Error(err))
	}
	defer kafkaWriter.Close()
	defer kafkaReader.Close()

	// Initialize event bus
	eventBus := messaging.NewKafkaEventBus(kafkaWriter, cfg.Kafka.AlertTopic, logger)

	// Initialize services
	abuseDetector := services.NewAbuseDetectorService(
		nil, // ruleRepo - to be implemented
		nil, // alertRepo - to be implemented
		eventBus,
		logger,
	)

	healthScorer := services.NewHealthScorerService(
		nil, // healthRepo - to be implemented
		nil, // exchangeRepo - to be implemented
		nil, // throttlePublisher - to be implemented
		nil, // cachePort - to be implemented
		logger,
	)

	// Create a combined service that implements the required interfaces
	oversightService := services.NewOversightService(repo, healthScorer, abuseDetector, logger)

	ingestionService := services.NewIngestionService(
		oversightService,
		oversightService,
		nil, // tradeAnalytics - to be implemented
		eventBus,
		logger,
	)

	// Initialize HTTP server
	httpAdapter := httpAdapter.NewHTTPServerAdapter(oversightService, eventBus, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      httpAdapter.GetRouter(),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start Kafka consumer in background
	go startKafkaConsumer(context.Background(), kafkaReader, ingestionService, logger)

	// Start HTTP server in goroutine
	go func() {
		logger.Info("HTTP server starting",
			zap.Int("port", cfg.Server.HTTPPort),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Service stopped")
}

func initLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

func initDatabase(cfg config.PostgresConfig, logger *zap.Logger) (*sql.DB, error) {
	dsn := cfg.GetDSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MinConnections)
	db.SetConnMaxLifetime(cfg.GetConnectionMaxLifetime())

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database),
	)

	return db, nil
}

func initKafka(cfg config.KafkaConfig, logger *zap.Logger) (*kafka.Writer, *kafka.Reader, error) {
	// Create Kafka writer for publishing alerts
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.AlertTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.GetBatchTimeout(),
		RequiredAcks: kafka.RequiredAcks(cfg.ProducerAck),
	}

	// Create Kafka reader for consuming trade events
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.TradeTopic,
		GroupID:        cfg.ConsumerGroup,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		MaxWait:        500 * time.Millisecond,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: time.Second,
	})

	logger.Info("Kafka connections established",
		zap.Strings("brokers", cfg.Brokers),
		zap.String("trade_topic", cfg.TradeTopic),
		zap.String("alert_topic", cfg.AlertTopic),
		zap.String("consumer_group", cfg.ConsumerGroup),
	)

	return writer, reader, nil
}

func startKafkaConsumer(ctx context.Context, reader *kafka.Reader, service *services.IngestionService, logger *zap.Logger) {
	consumer := messaging.NewKafkaConsumer(reader, &messaging.TradeEventConsumer{
		Logger:  logger,
		Service: service,
	}, logger)

	logger.Info("Starting Kafka consumer",
		zap.String("topic", reader.Config().Topic),
		zap.String("group_id", reader.Config().GroupID),
	)

	if err := consumer.Start(ctx); err != nil {
		if ctx.Err() == nil {
			logger.Error("Kafka consumer error", zap.Error(err))
		}
	}
}
