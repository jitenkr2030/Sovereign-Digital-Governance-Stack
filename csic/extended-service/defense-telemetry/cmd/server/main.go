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

	"defense-telemetry/internal/handlers"
	"defense-telemetry/internal/repository"
	"defense-telemetry/internal/service"

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
	Telemetry   TelemetryConfig   `yaml:"telemetry"`
	Equipment   EquipmentConfig   `yaml:"equipment"`
	Threats     ThreatsConfig     `yaml:"threats"`
	Maintenance MaintenanceConfig `yaml:"maintenance"`
	Deployment  DeploymentConfig  `yaml:"deployment"`
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
	Brokers      map[string][]string `yaml:"brokers"`
	ConsumerGroup string            `yaml:"consumer_group"`
	Topics       map[string]string   `yaml:"topics"`
}

type TelemetryConfig struct {
	BatchSize        int           `yaml:"batch_size"`
	BatchTimeout     time.Duration `yaml:"batch_timeout"`
	DedupWindow      time.Duration `yaml:"dedup_window"`
	DedupCacheTTL    time.Duration `yaml:"dedup_cache_ttl"`
	MaxProcessingTime time.Duration `yaml:"max_processing_time"`
	WorkerCount      int           `yaml:"worker_count"`
}

type EquipmentConfig struct {
	HullIntegrityWarning   float64 `yaml:"hull_integrity_warning"`
	HullIntegrityCritical  float64 `yaml:"hull_integrity_critical"`
	FuelLevelWarning       float64 `yaml:"fuel_level_warning"`
	FuelLevelCritical      float64 `yaml:"fuel_level_critical"`
	AmmunitionWarning      float64 `yaml:"ammunition_warning"`
	AmmunitionCritical     float64 `yaml:"ammunition_critical"`
	MaxVelocity            float64 `yaml:"max_velocity"`
	MinVelocity            float64 `yaml:"min_velocity"`
}

type ThreatsConfig struct {
	HighSeverityThreshold   int           `yaml:"high_severity_threshold"`
	CriticalSeverityThreshold int         `yaml:"critical_severity_threshold"`
	AlertCooldown           time.Duration `yaml:"alert_cooldown"`
	AutoEscalate            bool          `yaml:"auto_escalate"`
	EscalationDelay         time.Duration `yaml:"escalation_delay"`
}

type MaintenanceConfig struct {
	HealthyThreshold     float64   `yaml:"healthy_threshold"`
	WarningThreshold     float64   `yaml:"warning_threshold"`
	CriticalThreshold    float64   `yaml:"critical_threshold"`
	PredictionHorizon    time.Duration `yaml:"prediction_horizon"`
	TrendWindow          time.Duration `yaml:"trend_window"`
	Components           []string  `yaml:"components"`
}

type DeploymentConfig struct {
	Zones             []ZoneConfig `yaml:"zones"`
	MaxDistanceFromBase float64    `yaml:"max_distance_from_base"`
	OptimalCoverageRadius float64  `yaml:"optimal_coverage_radius"`
	OverlapTolerance   float64     `yaml:"overlap_tolerance"`
}

type ZoneConfig struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Priority int    `yaml:"priority"`
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
	telemetryService := service.NewTelemetryService(repo, redisClient, config)
	threatService := service.NewThreatService(repo, redisClient, config)
	maintenanceService := service.NewMaintenanceService(repo, redisClient, config)
	deploymentService := service.NewDeploymentService(repo, redisClient, config)

	// Initialize handlers
	telemetryHandler := handlers.NewTelemetryHandler(telemetryService)
	threatHandler := handlers.NewThreatHandler(threatService)
	maintenanceHandler := handlers.NewMaintenanceHandler(maintenanceService)
	deploymentHandler := handlers.NewDeploymentHandler(deploymentService)

	// Setup router
	router := setupRouter(telemetryHandler, threatHandler, maintenanceHandler, deploymentHandler)

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

func setupRouter(
	telemetryHandler *handlers.TelemetryHandler,
	threatHandler *handlers.ThreatHandler,
	maintenanceHandler *handlers.MaintenanceHandler,
	deploymentHandler *handlers.DeploymentHandler,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "defense-telemetry",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Telemetry endpoints
		telemetryHandler.RegisterRoutes(v1)

		// Threat endpoints
		threatHandler.RegisterRoutes(v1)

		// Maintenance endpoints
		maintenanceHandler.RegisterRoutes(v1)

		// Deployment endpoints
		deploymentHandler.RegisterRoutes(v1)
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
