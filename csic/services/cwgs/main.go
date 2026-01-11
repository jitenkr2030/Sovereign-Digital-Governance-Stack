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

	"github.com/csic/governance/internal/config"
	"github.com/csic/governance/internal/handlers"
	"github.com/csic/governance/internal/repository"
	"github.com/csic/governance/internal/services"
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

	logger.Info("Starting Custodial Wallet Governance Service")

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
	custodianRepo := repository.NewCustodianRepository(db, logger)
	walletRepo := repository.NewWalletRepository(db, logger)
	whitelistRepo := repository.NewWhitelistRepository(db, logger)
	blacklistRepo := repository.NewBlacklistRepository(db, logger)

	// Initialize services
	walletService := services.NewWalletService(walletRepo, whitelistRepo, blacklistRepo, logger)
	custodianService := services.NewCustodianService(custodianRepo, logger)
	complianceService := services.NewComplianceService(walletRepo, blacklistRepo, logger)

	// Initialize handlers
	walletHandler := handlers.NewWalletHandler(walletService, logger)
	custodianHandler := handlers.NewCustodianHandler(custodianService, logger)
	complianceHandler := handlers.NewComplianceHandler(complianceService, logger)

	// Create router
	router := mux.NewRouter()
	setupMiddleware(router, logger)
	setupRoutes(router, walletHandler, custodianHandler, complianceHandler, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	go func() {
		logger.Info("Custodial Wallet Governance listening", zap.Int("port", cfg.Server.HTTPPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
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
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	logger.Info("Database connected", zap.String("host", cfg.Host))
	return db, nil
}

func setupMiddleware(router *mux.Router, logger *zap.Logger) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Request", zap.String("method", r.Method), zap.String("path", r.URL.Path))
			next.ServeHTTP(w, r)
		})
	})
	router.Use(middlewareRecoverer)
	router.Use(middlewareCORS)
}

func middlewareRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func middlewareCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func setupRoutes(
	router *mux.Router,
	walletHandler *handlers.WalletHandler,
	custodianHandler *handlers.CustodianHandler,
	complianceHandler *handlers.ComplianceHandler,
	logger *zap.Logger,
) {
	router.HandleFunc("/health", healthCheckHandler(logger)).Methods(http.MethodGet)
	router.HandleFunc("/ready", readinessCheckHandler(logger)).Methods(http.MethodGet)

	api := router.PathPrefix("/api/v1").Subrouter()

	// Wallet routes
	api.HandleFunc("/wallets/register", walletHandler.RegisterWallet).Methods(http.MethodPost)
	api.HandleFunc("/wallets/{address}", walletHandler.GetWallet).Methods(http.MethodGet)
	api.HandleFunc("/wallets/{address}/status", walletHandler.UpdateStatus).Methods(http.MethodPut)
	api.HandleFunc("/wallets/{address}/whitelist", walletHandler.AddToWhitelist).Methods(http.MethodPost)
	api.HandleFunc("/wallets/{address}/blacklist", walletHandler.AddToBlacklist).Methods(http.MethodPost)
	api.HandleFunc("/wallets/search", walletHandler.SearchWallets).Methods(http.MethodGet)

	// Custodian routes
	api.HandleFunc("/custodians", custodianHandler.RegisterCustodian).Methods(http.MethodPost)
	api.HandleFunc("/custodians/{id}", custodianHandler.GetCustodian).Methods(http.MethodGet)
	api.HandleFunc("/custodians/{id}/wallets", custodianHandler.GetCustodianWallets).Methods(http.MethodGet)

	// Compliance routes
	api.HandleFunc("/compliance/check/{address}", complianceHandler.CheckCompliance).Methods(http.MethodGet)
	api.HandleFunc("/compliance/report", complianceHandler.GetComplianceReport).Methods(http.MethodGet)
}

func healthCheckHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"custodial-governance"}`))
	}
}

func readinessCheckHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready","service":"custodial-governance"}`))
	}
}
