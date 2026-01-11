package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"

	"csic-platform/service/licensing/internal/config"
	"csic-platform/service/licensing/internal/db"
	"csic-platform/service/licensing/internal/handler/http/handler"
	"csic-platform/service/licensing/internal/repository"
	"csic-platform/service/licensing/internal/service"
)

// Config represents the application configuration loaded from config.yaml
type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Cache    CacheConfig    `yaml:"cache"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// AppConfig contains application-specific settings
type AppConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	ReadTimeout     int    `yaml:"read_timeout"`
	WriteTimeout    int    `yaml:"write_timeout"`
	ShutdownTimeout int    `yaml:"shutdown_timeout"`
	Mode            string `yaml:"mode"` // debug, release, test
}

// DatabaseConfig contains PostgreSQL connection settings
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // in minutes
	SSLMode         string `yaml:"ssl_mode"`
}

// CacheConfig contains Redis connection settings
type CacheConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Password    string `yaml:"password"`
	Database    int    `yaml:"database"`
	PoolSize    int    `yaml:"pool_size"`
	MinIdleConn int    `yaml:"min_idle_conn"`
	TTL         int    `yaml:"ttl"` // in seconds
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"` // json, text
}

func main() {
	// Initialize context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration from config.yaml
	cfg, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	database, err := initializeDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.CloseConnection(database); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Initialize Redis cache connection
	cacheClient, err := initializeCache(cfg.Cache)
	if err != nil {
		log.Printf("Warning: Failed to initialize cache connection: %v", err)
		// Continue without cache - service will function without it
		cacheClient = nil
	}
	defer func() {
		if cacheClient != nil {
			if err := cacheClient.Close(); err != nil {
				log.Printf("Error closing cache connection: %v", err)
			}
		}
	}()

	// Initialize repository layer
	licenseRepo := repository.NewLicenseRepository(database)
	complianceRepo := repository.NewComplianceRepository(database)
	reportRepo := repository.NewReportRepository(database)
	dashboardRepo := repository.NewDashboardRepository(database)

	// Initialize cache wrapper for caching functionality
	var cacheRepo *repository.CacheRepository
	if cacheClient != nil {
		cacheRepo = repository.NewCacheRepository(cacheClient, cfg.Cache.TTL)
	}

	// Initialize service layer
	licenseService := service.NewLicenseService(licenseRepo, cacheRepo)
	complianceService := service.NewComplianceService(complianceRepo, licenseRepo, cacheRepo)
	reportingService := service.NewReportingService(reportRepo, licenseRepo, complianceRepo, dashboardRepo)
	dashboardService := service.NewDashboardService(dashboardRepo, licenseRepo, complianceRepo, reportRepo)

	// Initialize HTTP handlers
	httpHandler := handler.NewHandler(licenseService, complianceService, reportingService, dashboardService)

	// Set Gin mode
	gin.SetMode(cfg.App.Mode)

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(httpHandler.LoggerMiddleware(cfg.Logging.Level))
	router.Use(httpHandler.CORSMiddleware())

	// Setup routes
	httpHandler.SetupRoutes(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:         formatAddress(cfg.App.Host, cfg.App.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.App.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.App.WriteTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting CSIC Licensing Service on %s:%d", cfg.App.Host, cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, time.Duration(cfg.App.ShutdownTimeout)*time.Second)
	defer shutdownCancel()

	// Gracefully shutdown the HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

// loadConfig reads the configuration from the specified YAML file
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Apply environment variable overrides
	applyEnvOverrides(&cfg)

	return &cfg, nil
}

// applyEnvOverrides allows environment variables to override config values
func applyEnvOverrides(cfg *Config) {
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		var p int
		if _, err := fmt.Sscanf(port, "%d", &p); err == nil {
			cfg.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.Username = user
	}
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		cfg.Database.Password = pass
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		cfg.Database.Database = name
	}
	if cacheHost := os.Getenv("REDIS_HOST"); cacheHost != "" {
		cfg.Cache.Host = cacheHost
	}
	if appPort := os.Getenv("APP_PORT"); appPort != "" {
		var p int
		if _, err := fmt.Sscanf(appPort, "%d", &p); err == nil {
			cfg.App.Port = p
		}
	}
}

// initializeDatabase establishes a connection to the PostgreSQL database
func initializeDatabase(cfg DatabaseConfig) (*db.Database, error) {
	return db.NewConnection(db.ConnectionConfig{
		Host:            cfg.Host,
		Port:            cfg.Port,
		Username:        cfg.Username,
		Password:        cfg.Password,
		Database:        cfg.Database,
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.ConnMaxLifetime) * time.Minute,
		SSLMode:         cfg.SSLMode,
	})
}

// formatAddress creates the server address string from host and port
func formatAddress(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

// initializeCache establishes a connection to the Redis cache
func initializeCache(cfg CacheConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
		PoolSize: cfg.PoolSize,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to Redis cache")
	return client, nil
}

