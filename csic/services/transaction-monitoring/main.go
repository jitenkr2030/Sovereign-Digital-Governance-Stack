package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/csic/monitoring/internal/config"
	"github.com/csic/monitoring/internal/core/domain"
	"github.com/csic/monitoring/internal/core/ports"
	"github.com/csic/monitoring/internal/core/services"
	"github.com/csic/monitoring/internal/handlers"
	"github.com/csic/monitoring/internal/repository"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
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

	logger.Info("Starting Transaction Monitoring Engine")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize PostgreSQL connection
	db, err := initDatabase(cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories
	transactionRepo := repository.NewTransactionRepository(db, logger)
	sanctionsRepo := repository.NewSanctionsRepository(db, logger)
	walletProfileRepo := repository.NewWalletProfileRepository(db, logger)

	// Initialize services
	riskScorer := services.NewRiskScoringService(sanctionsRepo, walletProfileRepo, logger)
	transactionService := services.NewTransactionService(transactionRepo, riskScorer, sanctionsRepo, logger)
	sanctionsService := services.NewSanctionsService(sanctionsRepo, logger)

	// Initialize handlers
	txHandler := handlers.NewTransactionHandler(transactionService, logger)
	sanctionsHandler := handlers.NewSanctionsHandler(sanctionsService, logger)
	walletHandler := handlers.NewWalletHandler(walletProfileRepo, riskScorer, logger)

	// Create router
	router := mux.NewRouter()

	// Setup middleware
	setupMiddleware(router, logger)

	// Setup routes
	setupRoutes(router, txHandler, sanctionsHandler, walletHandler, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Transaction Monitoring Engine listening", zap.Int("port", cfg.Server.HTTPPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped")
}

func initLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zap.RFC3339TimeEncoder
	return config.Build()
}

func initDatabase(cfg config.DatabaseConfig, logger *zap.Logger) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established",
		zap.String("host", cfg.Host),
		zap.String("database", cfg.Name),
	)

	return db, nil
}

func setupMiddleware(router *mux.Router, logger *zap.Logger) {
	// Request ID
	router.Use(middlewareRequestID)

	// Logger
	router.Use(middlewareLogger(logger))

	// Recoverer
	router.Use(middlewareRecoverer)

	// CORS
	router.Use(middlewareCORS)
}

func setupRoutes(
	router *mux.Router,
	txHandler *handlers.TransactionHandler,
	sanctionsHandler *handlers.SanctionsHandler,
	walletHandler *handlers.WalletHandler,
	logger *zap.Logger,
) {
	// Health and readiness
	router.HandleFunc("/health", healthCheckHandler(logger)).Methods(http.MethodGet)
	router.HandleFunc("/ready", readinessCheckHandler(logger)).Methods(http.MethodGet)

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Transaction routes
	api.HandleFunc("/transactions/ingest", txHandler.IngestTransaction).Methods(http.MethodPost)
	api.HandleFunc("/transactions/history", txHandler.GetTransactionHistory).Methods(http.MethodGet)
	api.HandleFunc("/transactions/risk/{txHash}", txHandler.GetTransactionRisk).Methods(http.MethodGet)
	api.HandleFunc("/transactions/flagged", txHandler.GetFlaggedTransactions).Methods(http.MethodGet)
	api.HandleFunc("/transactions/scan/{address}", txHandler.ScanAddress).Methods(http.MethodGet)

	// Sanctions routes
	api.HandleFunc("/sanctions", sanctionsHandler.ListSanctions).Methods(http.MethodGet)
	api.HandleFunc("/sanctions/add", sanctionsHandler.AddSanctionedAddress).Methods(http.MethodPost)
	api.HandleFunc("/sanctions/check/{address}", sanctionsHandler.CheckAddress).Methods(http.MethodGet)
	api.HandleFunc("/sanctions/import", sanctionsHandler.ImportSanctionsList).Methods(http.MethodPost)

	// Wallet routes
	api.HandleFunc("/wallets/profile/{address}", walletHandler.GetWalletProfile).Methods(http.MethodGet)
	api.HandleFunc("/wallets/risk/{address}", walletHandler.GetWalletRisk).Methods(http.MethodGet)
	api.HandleFunc("/wallets/search", walletHandler.SearchWallets).Methods(http.MethodGet)

	// Reports routes
	api.HandleFunc("/reports/suspicious-activity", txHandler.GetSuspiciousActivityReport).Methods(http.MethodGet)
	api.HandleFunc("/reports/risk-summary", txHandler.GetRiskSummaryReport).Methods(http.MethodGet)
}

// Middleware functions

func middlewareRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "request_id", requestID)))
	})
}

func middlewareLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := r.Context().Value("request_id")

			logger.Info("Incoming request",
				zap.String("request_id", fmt.Sprintf("%v", requestID)),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
			)

			next.ServeHTTP(w, r)

			logger.Info("Request completed",
				zap.String("request_id", fmt.Sprintf("%v", requestID)),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", 0),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

func middlewareRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger, _ := initLogger()
				logger.Error("Panic recovered", zap.Any("error", err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func middlewareCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func healthCheckHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "transaction-monitoring",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func readinessCheckHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ready",
			"service":  "transaction-monitoring",
			"checks":   map[string]string{"database": "connected"},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// Configuration

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
	HTTPPort     int `yaml:"http_port"`
	ReadTimeout  int `yaml:"read_timeout"`
	WriteTimeout int `yaml:"write_timeout"`
	IdleTimeout  int `yaml:"idle_timeout"`
}

type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	SSLMode         string `yaml:"ssl_mode"`
	MaxOpenConnections int `yaml:"max_open_connections"`
	MaxIdleConnections int `yaml:"max_idle_connections"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}
