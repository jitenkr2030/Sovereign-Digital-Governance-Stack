package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic-platform/services/security/access-control/internal/adapter/consumer"
	"github.com/csic-platform/services/security/access-control/internal/adapter/repository"
	"github.com/csic-platform/services/security/access-control/internal/config"
	"github.com/csic-platform/services/security/access-control/internal/core/domain"
	"github.com/csic-platform/services/security/access-control/internal/core/service"
	"github.com/csic-platform/services/security/access-control/internal/handler"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := initLogger(cfg)
	defer logger.Sync()

	logger.Info("Starting access control service",
		zap.String("app", cfg.AppName),
		zap.String("env", cfg.Env),
	)

	// Initialize context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize PostgreSQL connection
	db, err := initPostgreSQL(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize PostgreSQL", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis client
	redisClient, err := initRedis(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize repositories
	policyRepo := repository.NewPostgresPolicyRepository(db, logger)
	auditRepo := repository.NewPostgresAuditLogRepository(db, logger)
	ownershipRepo := repository.NewPostgresResourceOwnershipRepository(db, logger)

	// Initialize services
	acService := service.NewAccessControlService(policyRepo, ownershipRepo, auditRepo, logger)

	// Initialize policy cache
	policyCache := NewPolicyCache(redisClient, cfg.RedisKeyPrefix, cfg.GetPolicyCacheTTLDuration(), logger)

	// Initialize Kafka consumer
	kafkaHandler := consumer.NewKafkaPolicyEventHandler(policyRepo, policyCache, logger)
	kafkaConsumer := consumer.NewKafkaConsumer(consumer.Config{
		Brokers:        cfg.KafkaBrokers,
		Topic:          cfg.KafkaTopic,
		ConsumerGroup:  cfg.KafkaConsumerGroup,
		MinBytes:       cfg.KafkaMinBytes,
		MaxBytes:       cfg.KafkaMaxBytes,
		MaxWait:        cfg.GetKafkaMaxWaitDuration(),
		StartOffset:    cfg.KafkaStartOffset,
	}, kafkaHandler, nil, logger)

	// Start Kafka consumer
	if err := kafkaConsumer.Start(ctx); err != nil {
		logger.Warn("Failed to start Kafka consumer", zap.Error(err))
	}

	// Initialize HTTP handler
	httpHandler := handler.NewHandler(acService, nil, logger)

	// Create router
	router := mux.NewRouter()
	httpHandler.RegisterRoutes(router)

	// Add CORS middleware
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400,
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      corsHandler.Handler(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("HTTP server starting",
			zap.String("address", srv.Addr),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down service...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop Kafka consumer
	if err := kafkaConsumer.Stop(); err != nil {
		logger.Error("Error stopping Kafka consumer", zap.Error(err))
	}

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	logger.Info("Service stopped")
}

// initLogger initializes the logger
func initLogger(cfg *config.Config) *zap.Logger {
	var level zapcore.Level
	switch cfg.LogLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	var zapConfig zap.Config
	if cfg.LogFormat == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	zapConfig.Level = zap.NewAtomicLevelAt(level)

	logger, err := zapConfig.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	return logger
}

// initPostgreSQL initializes the PostgreSQL connection
func initPostgreSQL(cfg *config.Config, logger *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.GetPostgresDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.GetConnMaxLifetimeDuration())

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL",
		zap.String("host", cfg.PostgresHost),
		zap.Int("port", cfg.PostgresPort),
		zap.String("database", cfg.PostgresDB),
	)

	return db, nil
}

// initRedis initializes the Redis client
func initRedis(cfg *config.Config, logger *zap.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("Connected to Redis",
		zap.String("address", cfg.GetRedisAddr()),
	)

	return client, nil
}

// PolicyCache implements the cache interface for policies
type PolicyCache struct {
	client    *redis.Client
	keyPrefix string
	ttl       time.Duration
	logger    *zap.Logger
}

// NewPolicyCache creates a new policy cache
func NewPolicyCache(client *redis.Client, keyPrefix string, ttl time.Duration, logger *zap.Logger) *PolicyCache {
	return &PolicyCache{
		client:    client,
		keyPrefix: keyPrefix,
		ttl:       ttl,
		logger:    logger,
	}
}

// Set sets policies in the cache
func (c *PolicyCache) Set(policies []domain.Policy) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(policies)
	if err != nil {
		return fmt.Errorf("failed to marshal policies: %w", err)
	}

	key := c.keyPrefix + "policies"
	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	c.logger.Debug("Policy cache updated", zap.Int("count", len(policies)))
	return nil
}

// Remove removes a policy from the cache
func (c *PolicyCache) Remove(policyID string) error {
	// Invalidate entire policy cache when a policy changes
	return c.Clear()
}

// Clear clears the policy cache
func (c *PolicyCache) Clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := c.keyPrefix + "policies"
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	c.logger.Debug("Policy cache cleared")
	return nil
}

// Compile-time check that PolicyCache implements the cache interface
var _ consumer.PolicyCache = (*PolicyCache)(nil)
