package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jitenkr2030/csic-platform/extended-services/national-identity/internal/handlers"
	"github.com/jitenkr2030/csic-platform/extended-services/national-identity/internal/repository"
	"github.com/jitenkr2030/csic-platform/extended-services/national-identity/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	App           AppConfig           `yaml:"app"`
	Database      DatabaseConfig      `yaml:"database"`
	Redis         RedisConfig         `yaml:"redis"`
	Kafka         KafkaConfig         `yaml:"kafka"`
	Identity      IdentityConfig      `yaml:"identity"`
	Privacy       PrivacyConfig       `yaml:"privacy"`
	RateLimit     RateLimitConfig     `yaml:"rate_limit"`
	Observability ObservabilityConfig `yaml:"observability"`
	Security      SecurityConfig      `yaml:"security"`
	Health        HealthConfig        `yaml:"health"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Environment string `yaml:"environment"`
	LogLevel    string `yaml:"log_level"`
}

type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	Schema          string `yaml:"schema"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime"`
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
	Brokers       []string    `yaml:"brokers"`
	ConsumerGroup string      `yaml:"consumer_group"`
	Topics        KafkaTopics `yaml:"topics"`
}

type KafkaTopics struct {
	Events       string `yaml:"events"`
	Verifications string `yaml:"verifications"`
	FraudAlerts  string `yaml:"fraud_alerts"`
	ModelUpdates string `yaml:"model_updates"`
}

type IdentityConfig struct {
	FraudScoreThreshold  int                `yaml:"fraud_score_threshold"`
	VelocityCheckWindow  string             `yaml:"velocity_check_window"`
	MaxLoginAttempts     int                `yaml:"max_login_attempts"`
	LockoutDuration      string             `yaml:"lockout_duration"`
	Behavioral           BehavioralConfig   `yaml:"behavioral"`
	FederatedLearning    FederatedLearning  `yaml:"federated_learning"`
	Biometric            BiometricConfig    `yaml:"biometric"`
}

type BehavioralConfig struct {
	BaselineWindow               string `yaml:"baseline_window"`
	AnomalyDetectionSensitivity  string `yaml:"anomaly_detection_sensitivity"`
	KeystrokeAnalysisEnabled     bool   `yaml:"keystroke_analysis_enabled"`
	MousePatternAnalysisEnabled  bool   `yaml:"mouse_pattern_analysis_enabled"`
}

type FederatedLearning struct {
	Enabled                 bool   `yaml:"enabled"`
	RoundInterval           string `yaml:"round_interval"`
	MinClientsPerRound      int    `yaml:"min_clients_per_round"`
	ModelAggregationStrategy string `yaml:"model_aggregation_strategy"`
}

type BiometricConfig struct {
	RequiredFactors       int  `yaml:"required_factors"`
	LivenessCheckEnabled  bool `yaml:"liveness_check_enabled"`
	SpoofingDetectionEnabled bool `yaml:"spoofing_detection_enabled"`
	MaxRetryAttempts      int  `yaml:"max_retry_attempts"`
}

type PrivacyConfig struct {
	DataRetention      string `yaml:"data_retention"`
	PIIEncryption      bool   `yaml:"pii_encryption"`
	AnonymizationEnabled bool `yaml:"anonymization_enabled"`
	ConsentRequired    bool   `yaml:"consent_required"`
	GDPRCompliance     bool   `yaml:"gdpr_compliance"`
}

type RateLimitConfig struct {
	Enabled            bool `yaml:"enabled"`
	RequestsPerMinute  int  `yaml:"requests_per_minute"`
	BurstSize          int  `yaml:"burst_size"`
}

type ObservabilityConfig struct {
	Metrics  MetricsConfig  `yaml:"metrics"`
	Tracing  TracingConfig  `yaml:"tracing"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type MetricsConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
}

type TracingConfig struct {
	Enabled    bool    `yaml:"enabled"`
	SampleRate float64 `yaml:"sample_rate"`
}

type LoggingConfig struct {
	Format string `yaml:"format"`
}

type SecurityConfig struct {
	TLS  TLSConfig  `yaml:"tls"`
	MTLS MTLSConfig `yaml:"mtls"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type MTLSConfig struct {
	Enabled bool   `yaml:"enabled"`
	CAFile  string `yaml:"ca_file"`
}

type HealthConfig struct {
	Endpoint          string `yaml:"endpoint"`
	ReadinessEndpoint string `yaml:"readiness_endpoint"`
	LivenessEndpoint  string `yaml:"liveness_endpoint"`
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

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set log level
	log.SetLevel(mapLogLevel(config.App.LogLevel))

	// Initialize repository
	repo, err := repository.NewPostgresRepository(repository.PostgresConfig{
		Host:            config.Database.Host,
		Port:            config.Database.Port,
		Username:        config.Database.Username,
		Password:        config.Database.Password,
		Name:            config.Database.Name,
		Schema:          config.Database.Schema,
		MaxOpenConns:    config.Database.MaxOpenConns,
		MaxIdleConns:    config.Database.MaxIdleConns,
		ConnMaxLifetime: config.Database.ConnMaxLifetime,
	})
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Initialize Redis cache
	cache, err := repository.NewRedisCache(repository.RedisConfig{
		Host:      config.Redis.Host,
		Port:      config.Redis.Port,
		Password:  config.Redis.Password,
		DB:        config.Redis.DB,
		KeyPrefix: config.Redis.KeyPrefix,
		PoolSize:  config.Redis.PoolSize,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Redis cache: %v", err)
	}
	defer cache.Close()

	// Initialize identity service
	identityService := service.NewIdentityService(service.IdentityServiceConfig{
		Repo:                   repo,
		Cache:                  cache,
		FraudScoreThreshold:    config.Identity.FraudScoreThreshold,
		VelocityCheckWindow:    config.Identity.VelocityCheckWindow,
		MaxLoginAttempts:       config.Identity.MaxLoginAttempts,
		LockoutDuration:        config.Identity.LockoutDuration,
		Behavioral:             config.Identity.Behavioral,
		FederatedLearning:      config.Identity.FederatedLearning,
		Biometric:              config.Identity.Biometric,
		Privacy:                config.Privacy,
	})

	// Initialize handlers
	identityHandler := handlers.NewIdentityHandler(identityService)
	healthHandler := handlers.NewHealthHandler()

	// Setup Gin router
	if config.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())
	router.Use(corsMiddleware())

	// Health endpoints
	router.GET(config.Health.LivenessEndpoint, healthHandler.Liveness)
	router.GET(config.Health.ReadinessEndpoint, healthHandler.Readiness)
	router.GET(config.Health.Endpoint, healthHandler.Health)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Identity management
		identities := v1.Group("/identities")
		{
			identities.POST("", identityHandler.CreateIdentity)
			identities.GET("/:id", identityHandler.GetIdentity)
			identities.PUT("/:id", identityHandler.UpdateIdentity)
			identities.DELETE("/:id", identityHandler.DeleteIdentity)
			identities.GET("/:id/history", identityHandler.GetIdentityHistory)
		}

		// Verification endpoints
		verify := v1.Group("/verify")
		{
			verify.POST("", identityHandler.VerifyIdentity)
			verify.POST("/biometric", identityHandler.BiometricVerification)
			verify.POST("/document", identityHandler.DocumentVerification)
			verify.POST("/liveness", identityHandler.LivenessCheck)
		}

		// Authentication
		auth := v1.Group("/auth")
		{
			auth.POST("/login", identityHandler.Login)
			auth.POST("/logout", identityHandler.Logout)
			auth.POST("/refresh", identityHandler.RefreshToken)
			auth.POST("/mfa", identityHandler.SetupMFA)
			auth.POST("/mfa/verify", identityHandler.VerifyMFA)
			auth.POST("/password/reset", identityHandler.PasswordReset)
			auth.POST("/password/change", identityHandler.ChangePassword)
		}

		// Behavioral biometrics
		behavioral := v1.Group("/behavioral")
		{
			behavioral.POST("/ keystroke", identityHandler.SubmitKeystrokeData)
			behavioral.POST("/mouse", identityHandler.SubmitMouseData)
			behavioral.POST("/analyze", identityHandler.AnalyzeBehavioralPattern)
			behavioral.GET("/:id/profile", identityHandler.GetBehavioralProfile)
		}

		// Fraud detection
		fraud := v1.Group("/fraud")
		{
			fraud.GET("/score/:id", identityHandler.GetFraudScore)
			fraud.POST("/report", identityHandler.ReportFraud)
			fraud.GET("/alerts", identityHandler.GetFraudAlerts)
			fraud.POST("/alerts/:id/resolve", identityHandler.ResolveFraudAlert)
		}

		// Federated learning
		federated := v1.Group("/federated")
		{
			federated.GET("/model/status", identityHandler.GetModelStatus)
			federated.POST("/model/update", identityHandler.SubmitModelUpdate)
			federated.GET("/contributions/:id", identityHandler.GetClientContributions)
		}
	}

	// Metrics endpoint
	if config.Observability.Metrics.Enabled {
		router.GET(config.Observability.Metrics.Endpoint, gin.WrapH(promhttp.Handler()))
	}

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", config.App.Host, config.App.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Configure TLS if enabled
	if config.Security.TLS.Enabled {
		cert, err := tls.LoadX509KeyPair(config.Security.TLS.CertFile, config.Security.TLS.KeyFile)
		if err != nil {
			log.Fatalf("Failed to load TLS certificates: %v", err)
		}
		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting %s on %s", config.App.Name, addr)
		if config.Security.TLS.Enabled {
			if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start server: %v", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start server: %v", err)
			}
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

func mapLogLevel(level string) int {
	switch level {
	case "debug":
		return log.Ldebug
	case "info":
		return log.LstdFlags
	case "warn":
		return log.LstdFlags
	case "error":
		return log.LstdFlags | log.Lshortfile
	default:
		return log.LstdFlags
	}
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if query != "" {
			path = path + "?" + query
		}

		log.Printf("[%d] %s %s %v",
			c.Writer.Status(),
			c.Request.Method,
			path,
			latency,
		)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, Accept, X-Requested-With, X-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
