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

	"energy-optimization/internal/handlers"
	"energy-optimization/internal/repository"
	"energy-optimization/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	App         AppConfig         `yaml:"app"`
	Database    DatabaseConfig    `yaml:"database"`
	Redis       RedisConfig       `yaml:"redis"`
	Kafka       KafkaConfig       `yaml:"kafka"`
	Carbon      CarbonConfig      `yaml:"carbon"`
	LoadBalance LoadBalanceConfig `yaml:"load_balancing"`
	Meters      MetersConfig      `yaml:"meters"`
	Metrics     MetricsConfig     `yaml:"metrics"`
	Logging     LoggingConfig     `yaml:"logging"`
}

type AppConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
}

type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	Name            string        `yaml:"name"`
	TablePrefix     string        `yaml:"table_prefix"`
	SSLMode         string        `yaml:"sslmode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Password  string `yaml:"password"`
	DB        int    `yaml:"db"`
	KeyPrefix string `yaml:"key_prefix"`
	PoolSize  int    `yaml:"pool_size"`
}

type KafkaConfig struct {
	Brokers  map[string][]string `yaml:"brokers"`
	Group    string              `yaml:"consumer_group"`
	Topics   map[string]string   `yaml:"topics"`
}

type CarbonConfig struct {
	EmissionFactors map[string]decimal.Decimal `yaml:"emission_factors"`
	BudgetWarning   float64                    `yaml:"budget_warning_threshold"`
	BudgetCritical  float64                    `yaml:"budget_critical_threshold"`
}

type LoadBalanceConfig struct {
	TargetMin      float64 `yaml:"target_utilization_min"`
	TargetMax      float64 `yaml:"target_utilization_max"`
	ScaleUp        float64 `yaml:"scale_up_threshold"`
	ScaleDown      float64 `yaml:"scale_down_threshold"`
	CheckInterval  string  `yaml:"check_interval"`
}

type MetersConfig struct {
	DefaultSampleInterval time.Duration `yaml:"default_sample_interval"`
	AggregationWindow     time.Duration `yaml:"aggregation_window"`
	MaxReadingAge         time.Duration `yaml:"max_reading_age"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
	Port    int    `yaml:"port"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func main() {
	// Load configuration
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode based on environment
	if config.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize repository
	repo, err := repository.NewPostgresRepository(&config.Database)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Initialize Redis client
	redisClient, err := repository.NewRedisClient(&config.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}
	defer redisClient.Close()

	// Initialize services
	energyService := service.NewEnergyService(repo, redisClient, config)
	carbonService := service.NewCarbonService(repo, redisClient, config)
	loadBalanceService := service.NewLoadBalancingService(repo, redisClient, config)

	// Initialize handlers
	energyHandler := handlers.NewEnergyHandler(energyService)
	carbonHandler := handlers.NewCarbonHandler(carbonService)
	loadBalanceHandler := handlers.NewLoadBalancingHandler(loadBalanceService)

	// Setup router
	router := setupRouter(energyHandler, carbonHandler, loadBalanceHandler)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.App.Host, config.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start metrics server if enabled
	if config.Metrics.Enabled {
		go startMetricsServer(config.Metrics.Port)
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting %s on port %d", config.App.Name, config.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func setupRouter(energyHandler *handlers.EnergyHandler, carbonHandler *handlers.CarbonHandler, loadBalanceHandler *handlers.LoadBalancingHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "energy-optimization",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Energy readings endpoints
		energyHandler.RegisterRoutes(v1)

		// Carbon tracking endpoints
		carbonHandler.RegisterRoutes(v1)

		// Load balancing endpoints
		loadBalanceHandler.RegisterRoutes(v1)
	}

	return router
}

func startMetricsServer(port int) {
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting metrics server on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Printf("Metrics server error: %v", err)
	}
}
