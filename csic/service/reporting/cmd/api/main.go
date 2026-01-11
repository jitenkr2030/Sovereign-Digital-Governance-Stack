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

	"csic-platform/service/reporting/internal/db"
	"csic-platform/service/reporting/internal/handler/http/handler"
	"csic-platform/service/reporting/internal/handler/kafka"
	"csic-platform/service/reporting/internal/repository"
	"csic-platform/service/reporting/internal/service"
)

// Config represents the application configuration
type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Cache    CacheConfig    `yaml:"cache"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	Security SecurityConfig `yaml:"security"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// AppConfig contains application settings
type AppConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	ReadTimeout     int    `yaml:"read_timeout"`
	WriteTimeout    int    `yaml:"write_timeout"`
	ShutdownTimeout int    `yaml:"shutdown_timeout"`
	Mode            string `yaml:"mode"` // debug, release, test
}

// DatabaseConfig contains PostgreSQL settings
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // minutes
	SSLMode         string `yaml:"ssl_mode"`
}

// CacheConfig contains Redis settings
type CacheConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Password    string `yaml:"password"`
	Database    int    `yaml:"database"`
	PoolSize    int    `yaml:"pool_size"`
	TTL         int    `yaml:"ttl"` // seconds
}

// KafkaConfig contains Kafka settings
type KafkaConfig struct {
	Enabled       bool     `yaml:"enabled"`
	Brokers       []string `yaml:"brokers"`
	ConsumerGroup string   `yaml:"consumer_group"`
	TopicPrefix   string   `yaml:"topic_prefix"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	EncryptionEnabled   bool   `yaml:"encryption_enabled"`
	DefaultAlgorithm    string `yaml:"default_algorithm"`
	MaxDownloadCount    int    `yaml:"max_download_count"`
	DefaultExpiryHours  int    `yaml:"default_expiry_hours"`
	RequireAuth         bool   `yaml:"require_auth"`
	WatermarkEnabled    bool   `yaml:"watermark_enabled"`
	WatermarkText       string `yaml:"watermark_text"`
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

	// Load configuration
	cfg, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	database, err := db.NewConnection(db.ConnectionConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		Username:        cfg.Database.Username,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Database,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute,
		SSLMode:         cfg.Database.SSLMode,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.CloseConnection(database)

	// Initialize Redis cache
	var cacheRepo *repository.CacheRepository
	cacheClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Cache.Host, cfg.Cache.Port),
		Password: cfg.Cache.Password,
		DB:       cfg.Cache.Database,
		PoolSize: cfg.Cache.PoolSize,
	})

	if err := cacheClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis cache: %v", err)
	} else {
		cacheRepo = repository.NewCacheRepository(cacheClient, cfg.Cache.TTL)
		log.Println("Connected to Redis cache")
	}
	defer func() {
		if cacheClient != nil {
			cacheClient.Close()
		}
	}()

	// Initialize repositories
	templateRepo := repository.NewPostgresReportTemplateRepository(database)
	reportRepo := repository.NewPostgresGeneratedReportRepository(database)
	scheduleRepo := repository.NewPostgresScheduleRepository(database)
	exportLogRepo := repository.NewPostgresExportLogRepository(database)

	// Initialize Kafka
	var kafkaProducer *kafka.Producer
	if cfg.Kafka.Enabled {
		kafkaCfg := kafka.KafkaConfig{
			Brokers:       cfg.Kafka.Brokers,
			ConsumerGroup: cfg.Kafka.ConsumerGroup,
			TopicPrefix:   cfg.Kafka.TopicPrefix,
		}
		kafkaProducer = kafka.NewProducer(kafkaCfg)
		defer kafkaProducer.Close()
		log.Println("Kafka producer initialized")
	}

	// Initialize file storage (placeholder - would use S3, GCS, etc.)
	fileStorage := &LocalFileStorage{
		basePath: "/tmp/reports",
	}

	// Initialize metrics (placeholder)
	metrics := &MetricsRecorder{}

	// Initialize services
	reportService := service.NewReportGenerationService(
		templateRepo, reportRepo, cacheRepo, kafkaProducer, fileStorage, metrics,
	)

	exportConfig := service.SecureExportConfig{
		EncryptionEnabled:   cfg.Security.EncryptionEnabled,
		DefaultAlgorithm:    cfg.Security.DefaultAlgorithm,
		MaxDownloadCount:    cfg.Security.MaxDownloadCount,
		DefaultExpiryHours:  cfg.Security.DefaultExpiryHours,
		RequireAuth:         cfg.Security.RequireAuth,
		WatermarkEnabled:    cfg.Security.WatermarkEnabled,
		WatermarkText:       cfg.Security.WatermarkText,
	}

	exportService := service.NewExportService(
		reportRepo, exportLogRepo, cacheRepo, fileStorage, exportConfig,
	)

	var schedulerService *service.SchedulerService
	if cfg.Kafka.Enabled {
		schedulerService = service.NewSchedulerService(
			scheduleRepo, templateRepo, reportService, kafkaProducer,
		)
	}

	// Start scheduler
	if schedulerService != nil {
		go func() {
			if err := schedulerService.Start(ctx); err != nil {
				log.Printf("Failed to start scheduler: %v", err)
			}
		}()
	}

	// Initialize HTTP handler
	httpHandler := handler.NewHandler(reportService, exportService, schedulerService)

	// Set Gin mode
	gin.SetMode(cfg.App.Mode)

	// Initialize router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(httpHandler.LoggerMiddleware())
	router.Use(httpHandler.CORSMiddleware())
	router.Use(httpHandler.AuthenticationMiddleware())

	// Setup routes
	httpHandler.SetupRoutes(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.App.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.App.WriteTimeout) * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Starting CSIC Reporting Service on %s:%d", cfg.App.Host, cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Stop scheduler
	if schedulerService != nil {
		schedulerService.Stop()
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, time.Duration(cfg.App.ShutdownTimeout)*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// loadConfig loads configuration from file
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LocalFileStorage implements FileStorage using local filesystem
type LocalFileStorage struct {
	basePath string
}

// SaveReport saves a report to local filesystem
func (s *LocalFileStorage) SaveReport(ctx context.Context, reportID uuid.UUID, format domain.ReportFormat, content io.Reader) (string, int64, error) {
	// Create directory structure
	dirPath := fmt.Sprintf("%s/%s/%s", s.basePath, reportID.String()[:8], format)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", 0, err
	}

	// Create file
	filename := fmt.Sprintf("report.%s", format)
	filePath := fmt.Sprintf("%s/%s", dirPath, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	// Copy content
	size, err := io.Copy(file, content)
	if err != nil {
		return "", 0, err
	}

	return filePath, size, nil
}

// GetReport retrieves a report from local filesystem
func (s *LocalFileStorage) GetReport(ctx context.Context, filePath string) (io.ReadCloser, error) {
	return os.Open(filePath)
}

// DeleteReport deletes a report from local filesystem
func (s *LocalFileStorage) DeleteReport(ctx context.Context, filePath string) error {
	return os.Remove(filePath)
}

// MetricsRecorder implements MetricsRecorder
type MetricsRecorder struct{}

// RecordReportGenerated records report generation metrics
func (m *MetricsRecorder) RecordReportGenerated(reportType string, duration time.Duration, success bool) {
	log.Printf("Report generated: type=%s duration=%v success=%v", reportType, duration, success)
}

// RecordReportSize records report size metrics
func (m *MetricsRecorder) RecordReportSize(reportType string, size int64) {
	log.Printf("Report size: type=%s size=%d", reportType, size)
}

// Import necessary packages
import (
	"io"
	"csic-platform/service/reporting/internal/domain"
)
