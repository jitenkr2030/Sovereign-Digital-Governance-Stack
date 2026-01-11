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

	"smart-cities/internal/handlers"
	"smart-cities/internal/repository"
	"smart-cities/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	App            AppConfig            `yaml:"app"`
	Database       DatabaseConfig       `yaml:"database"`
	Redis          RedisConfig          `yaml:"redis"`
	Kafka          KafkaConfig          `yaml:"kafka"`
	IoT            IoTConfig            `yaml:"iot"`
	Infrastructure InfrastructureConfig `yaml:"infrastructure"`
	Resources      ResourcesConfig      `yaml:"resources"`
	Districts      []DistrictConfig     `yaml:"districts"`
	Analytics      AnalyticsConfig      `yaml:"analytics"`
	Incidents      IncidentsConfig      `yaml:"incidents"`
	Sustainability SustainabilityConfig `yaml:"sustainability"`
	Metrics        MetricsConfig        `yaml:"metrics"`
	Logging        LoggingConfig        `yaml:"logging"`
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
	Brokers       map[string][]string `yaml:"brokers"`
	ConsumerGroup string             `yaml:"consumer_group"`
	Topics        map[string]string   `yaml:"topics"`
}

type IoTConfig struct {
	BatchSize        int           `yaml:"batch_size"`
	BatchTimeout     time.Duration `yaml:"batch_timeout"`
	RateLimitRequests int          `yaml:"rate_limit_requests"`
	RateLimitWindow  time.Duration `yaml:"rate_limit_window"`
	MaxReadingAge    time.Duration `yaml:"max_reading_age"`
	AggregationWindow time.Duration `yaml:"aggregation_window"`
}

type InfrastructureConfig struct {
	NodeTypes []string `yaml:"node_types"`
}

type ResourcesConfig struct {
	EnergyWarning      float64 `yaml:"energy_warning"`
	EnergyCritical     float64 `yaml:"energy_critical"`
	WaterFlowWarning   float64 `yaml:"water_flow_warning"`
	WaterFlowCritical  float64 `yaml:"water_flow_critical"`
	WasteWarning       float64 `yaml:"waste_warning"`
	WasteCritical      float64 `yaml:"waste_critical"`
	AQIWarning         int     `yaml:"aqi_warning"`
	AQIUnhealthy       int     `yaml:"aqi_unhealthy"`
}

type DistrictConfig struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Priority int    `yaml:"priority"`
}

type AnalyticsConfig struct {
	RealTimeWindow           time.Duration `yaml:"real_time_window"`
	HourlyWindow             time.Duration `yaml:"hourly_window"`
	DailyWindow              time.Duration `yaml:"daily_window"`
	TrendDetectionThreshold  float64       `yaml:"trend_detection_threshold"`
	AnomalyDetectionSensitivity float64     `yaml:"anomaly_detection_sensitivity"`
}

type IncidentsConfig struct {
	PriorityCritical   int           `yaml:"priority_critical"`
	PriorityHigh       int           `yaml:"priority_high"`
	PriorityMedium     int           `yaml:"priority_medium"`
	PriorityLow        int           `yaml:"priority_low"`
	AutoAssign         bool          `yaml:"auto_assign"`
	ResponseTimeCritical time.Duration `yaml:"response_time_critical"`
	ResponseTimeHigh    time.Duration `yaml:"response_time_high"`
	ResponseTimeMedium  time.Duration `yaml:"response_time_medium"`
	ResponseTimeLow     time.Duration `yaml:"response_time_low"`
}

type SustainabilityConfig struct {
	EnergyReductionTarget   float64 `yaml:"energy_reduction_target"`
	WaterConservationTarget float64 `yaml:"water_conservation_target"`
	WasteReductionTarget    float64 `yaml:"waste_reduction_target"`
	CarbonReductionTarget   float64 `yaml:"carbon_reduction_target"`
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
	iotService := service.NewIoTService(repo, redisClient, config)
	incidentService := service.NewIncidentService(repo, redisClient, config)
	analyticsService := service.NewAnalyticsService(repo, redisClient, config)
	resourceService := service.NewResourceService(repo, redisClient, config)

	// Initialize handlers
	iotHandler := handlers.NewIoTHandler(iotService)
	incidentHandler := handlers.NewIncidentHandler(incidentService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	resourceHandler := handlers.NewResourceHandler(resourceService)

	// Setup router
	router := setupRouter(iotHandler, incidentHandler, analyticsHandler, resourceHandler)

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
	iotHandler *handlers.IoTHandler,
	incidentHandler *handlers.IncidentHandler,
	analyticsHandler *handlers.AnalyticsHandler,
	resourceHandler *handlers.ResourceHandler,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "smart-cities",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// IoT endpoints
		iotHandler.RegisterRoutes(v1)

		// Incident endpoints
		incidentHandler.RegisterRoutes(v1)

		// Analytics endpoints
		analyticsHandler.RegisterRoutes(v1)

		// Resource endpoints
		resourceHandler.RegisterRoutes(v1)
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
