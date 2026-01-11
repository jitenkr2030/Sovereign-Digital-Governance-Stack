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

	"github.com/jitenkr2030/csic-platform/extended-services/financial-crime-unit/internal/handlers"
	"github.com/jitenkr2030/csic-platform/extended-services/financial-crime-unit/internal/repository"
	"github.com/jitenkr2030/csic-platform/extended-services/financial-crime-unit/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	App              AppConfig              `yaml:"app"`
	Database         DatabaseConfig         `yaml:"database"`
	Redis            RedisConfig            `yaml:"redis"`
	Kafka            KafkaConfig            `yaml:"kafka"`
	Detection        DetectionConfig        `yaml:"detection"`
	Sanctions        SanctionsConfig        `yaml:"sanctions"`
	CaseManagement   CaseManagementConfig   `yaml:"case_management"`
	SAR              SARConfig               `yaml:"sar"`
	NetworkAnalysis  NetworkAnalysisConfig  `yaml:"network_analysis"`
	RiskScoring      RiskScoringConfig      `yaml:"risk_scoring"`
	Privacy          PrivacyConfig          `yaml:"privacy"`
	RateLimit        RateLimitConfig        `yaml:"rate_limit"`
	Observability    ObservabilityConfig    `yaml:"observability"`
	Security         SecurityConfig         `yaml:"security"`
	Health           HealthConfig           `yaml:"health"`
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
	Brokers       []string       `yaml:"brokers"`
	ConsumerGroup string         `yaml:"consumer_group"`
	Topics        KafkaTopics    `yaml:"topics"`
}

type KafkaTopics struct {
	Alerts          string `yaml:"alerts"`
	SARs            string `yaml:"sars"`
	IdentityUpdates string `yaml:"identity_updates"`
	Transactions    string `yaml:"transactions"`
	Reports         string `yaml:"reports"`
}

type DetectionConfig struct {
	Thresholds    ThresholdsConfig    `yaml:"thresholds"`
	Transaction   TransactionConfig   `yaml:"transaction"`
	Geographic    GeographicConfig    `yaml:"geographic"`
	Velocity      VelocityConfig      `yaml:"velocity"`
}

type ThresholdsConfig struct {
	HighRisk        int `yaml:"high_risk"`
	MediumRisk      int `yaml:"medium_risk"`
	LowRisk         int `yaml:"low_risk"`
	CriticalAlertScore int `yaml:"critical_alert_score"`
}

type TransactionConfig struct {
	VelocityWindow          string `yaml:"velocity_window"`
	MaxTransactionsPerWindow int   `yaml:"max_transactions_per_window"`
	MaxAmountPerWindow      int    `yaml:"max_amount_per_window"`
	StructuringThreshold    int    `yaml:"structuring_threshold"`
	RoundAmountThreshold    int    `yaml:"round_amount_threshold"`
}

type GeographicConfig struct {
	ImpossibleTravelWindow string   `yaml:"impossible_travel_window"`
	HighRiskCountries      []string `yaml:"high_risk_countries"`
	SanctionedCountries    []string `yaml:"sanctioned_countries"`
}

type VelocityConfig struct {
	DailyTransactionLimit  int `yaml:"daily_transaction_limit"`
	HourlyTransactionLimit int `yaml:"hourly_transaction_limit"`
	MaxRecipientsPerDay    int `yaml:"max_recipients_per_day"`
}

type SanctionsConfig struct {
	Enabled                      bool    `yaml:"enabled"`
	FuzzyMatchThreshold          float64 `yaml:"fuzzy_match_threshold"`
	PEPCheckEnabled              bool    `yaml:"pep_check_enabled"`
	SanctionListUpdateInterval   string  `yaml:"sanction_list_update_interval"`
	WatchlistRefreshInterval     string  `yaml:"watchlist_refresh_interval"`
}

type CaseManagementConfig struct {
	AutoCreateCasesAboveScore int           `yaml:"auto_create_cases_above_score"`
	DefaultAssignee           string        `yaml:"default_assignee"`
	Workflow                  []WorkflowState `yaml:"workflow"`
	EscalationTimeout         string        `yaml:"escalation_timeout"`
}

type WorkflowState struct {
	State      string   `yaml:"state"`
	NextStates []string `yaml:"next_states"`
}

type SARConfig struct {
	Enabled                 bool     `yaml:"enabled"`
	TemplateFormat          string   `yaml:"template_format"`
	RegulatoryBody          string   `yaml:"regulatory_body"`
	AutoGenerateThreshold   int      `yaml:"auto_generate_threshold"`
	MandatoryFields         []string `yaml:"mandatory_fields"`
	FilingDeadlineHours     int      `yaml:"filing_deadline_hours"`
}

type NetworkAnalysisConfig struct {
	Enabled          bool     `yaml:"enabled"`
	MaxHops          int      `yaml:"max_hops"`
	AnalysisTimeout  string   `yaml:"analysis_timeout"`
	HighRiskPatterns []string `yaml:"high_risk_patterns"`
}

type RiskScoringConfig struct {
	BaseScore float64            `yaml:"base_score"`
	Factors   RiskFactorsConfig  `yaml:"factors"`
	Weights   RiskWeightsConfig  `yaml:"weights"`
}

type RiskFactorsConfig struct {
	TransactionAmount   float64 `yaml:"transaction_amount"`
	VelocityScore       float64 `yaml:"velocity_score"`
	GeographicRisk      float64 `yaml:"geographic_risk"`
	EntityRisk          float64 `yaml:"entity_risk"`
	BehavioralPattern   float64 `yaml:"behavioral_pattern"`
}

type RiskWeightsConfig struct {
	KYCLevel           float64 `yaml:"kyc_level"`
	AccountAge         float64 `yaml:"account_age"`
	TransactionHistory float64 `yaml:"transaction_history"`
	BlacklistHit       float64 `yaml:"blacklist_hit"`
}

type PrivacyConfig struct {
	DataRetention        string `yaml:"data_retention"`
	PIIEncryption        bool   `yaml:"pii_encryption"`
	GDPRCompliance       bool   `yaml:"gdpr_compliance"`
	AuditLogRetention    string `yaml:"audit_log_retention"`
	SARRetention         string `yaml:"sar_retention"`
}

type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
	BurstSize         int  `yaml:"burst_size"`
}

type ObservabilityConfig struct {
	Metrics   MetricsConfig   `yaml:"metrics"`
	Tracing   TracingConfig   `yaml:"tracing"`
	Logging   LoggingConfig   `yaml:"logging"`
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
	Format     string `yaml:"format"`
	PIIMasking bool   `yaml:"pii_masking"`
}

type SecurityConfig struct {
	TLS    TLSConfig      `yaml:"tls"`
	MTLS   MTLSConfig     `yaml:"mtls"`
	APIKeys APIKeysConfig `yaml:"api_keys"`
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

type APIKeysConfig struct {
	Enabled   bool   `yaml:"enabled"`
	HeaderName string `yaml:"header_name"`
}

type HealthConfig struct {
	Endpoint            string `yaml:"endpoint"`
	ReadinessEndpoint   string `yaml:"readiness_endpoint"`
	LivenessEndpoint    string `yaml:"liveness_endpoint"`
	DependencyTimeout   string `yaml:"dependency_timeout"`
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

	// Initialize FCU service
	fcuService := service.NewFCUService(service.FCUServiceConfig{
		Repo:             repo,
		Cache:            cache,
		Detection:        config.Detection,
		Sanctions:        config.Sanctions,
		CaseManagement:   config.CaseManagement,
		SAR:              config.SAR,
		NetworkAnalysis:  config.NetworkAnalysis,
		RiskScoring:      config.RiskScoring,
		Privacy:          config.Privacy,
		KafkaTopics:      config.Kafka.Topics,
	})

	// Initialize handlers
	fcuHandler := handlers.NewFCUHandler(fcuService)
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
		// Transaction screening
		screen := v1.Group("/screen")
		{
			screen.POST("/transaction", fcuHandler.ScreenTransaction)
			screen.POST("/entity", fcuHandler.ScreenEntity)
			screen.POST("/batch", fcuHandler.ScreenBatch)
		}

		// Sanctions and PEP screening
		sanctions := v1.Group("/sanctions")
		{
			sanctions.POST("/check", fcuHandler.CheckSanctions)
			sanctions.POST("/bulk-check", fcuHandler.BulkSanctionsCheck)
			sanctions.GET("/lists", fcuHandler.GetSanctionLists)
			sanctions.POST("/lists/sync", fcuHandler.SyncSanctionLists)
		}

		// Alerts management
		alerts := v1.Group("/alerts")
		{
			alerts.GET("", fcuHandler.GetAlerts)
			alerts.GET("/:id", fcuHandler.GetAlert)
			alerts.POST("/:id/acknowledge", fcuHandler.AcknowledgeAlert)
			alerts.POST("/:id/escalate", fcuHandler.EscalateAlert)
			alerts.GET("/severity/:severity", fcuHandler.GetAlertsBySeverity)
			alerts.GET("/status/:status", fcuHandler.GetAlertsByStatus)
		}

		// Case management
		cases := v1.Group("/cases")
		{
			cases.GET("", fcuHandler.GetCases)
			cases.POST("", fcuHandler.CreateCase)
			cases.GET("/:id", fcuHandler.GetCase)
			cases.PUT("/:id", fcuHandler.UpdateCase)
			cases.POST("/:id/workflow", fcuHandler.TransitionCaseWorkflow)
			cases.POST("/:id/assign", fcuHandler.AssignCase)
			cases.POST("/:id/add-evidence", fcuHandler.AddEvidenceToCase)
			cases.GET("/assignee/:assignee", fcuHandler.GetCasesByAssignee)
			cases.GET("/status/:status", fcuHandler.GetCasesByStatus)
		}

		// SAR (Suspicious Activity Report) management
		sar := v1.Group("/sar")
		{
			sar.GET("", fcuHandler.GetSARs)
			sar.POST("", fcuHandler.CreateSAR)
			sar.GET("/:id", fcuHandler.GetSAR)
			sar.PUT("/:id", fcuHandler.UpdateSAR)
			sar.POST("/:id/validate", fcuHandler.ValidateSAR)
			sar.POST("/:id/submit", fcuHandler.SubmitSAR)
			sar.POST("/:id/generate", fcuHandler.GenerateAutoSAR)
			sar.GET("/case/:case_id", fcuHandler.GetSARsByCase)
		}

		// Risk scoring
		risk := v1.Group("/risk")
		{
			risk.GET("/score/:entity_id", fcuHandler.GetRiskScore)
			risk.GET("/profile/:entity_id", fcuHandler.GetRiskProfile)
			risk.GET("/history/:entity_id", fcuHandler.GetRiskHistory)
			risk.POST("/calculate", fcuHandler.CalculateRisk)
		}

		// Network analysis
		network := v1.Group("/network")
		{
			network.POST("/analyze", fcuHandler.AnalyzeNetwork)
			network.POST("/trace", fcuHandler.TraceTransactionFlow)
			network.GET("/patterns/:entity_id", fcuHandler.GetNetworkPatterns)
			network.GET("/connections/:entity_id", fcuHandler.GetEntityConnections)
		}

		// Entity monitoring
		entities := v1.Group("/entities")
		{
			entities.GET("/:id", fcuHandler.GetEntity)
			entities.PUT("/:id", fcuHandler.UpdateEntity)
			entities.GET("/:id/transactions", fcuHandler.GetEntityTransactions)
			entities.GET("/:id/alerts", fcuHandler.GetEntityAlerts)
			entities.POST("/:id/watchlist-add", fcuHandler.AddToWatchlist)
			entities.POST("/:id/watchlist-remove", fcuHandler.RemoveFromWatchlist)
			entities.GET("/:id/watchlist-status", fcuHandler.GetWatchlistStatus)
		}

		// Investigations
		investigation := v1.Group("/investigation")
		{
			investigation.GET("/:id", fcuHandler.GetInvestigation)
			investigation.POST("", fcuHandler.StartInvestigation)
			investigation.POST("/:id/notes", fcuHandler.AddInvestigationNote)
			investigation.POST("/:id/timeline", fcuHandler.AddTimelineEvent)
		}

		// Reports and analytics
		reports := v1.Group("/reports")
		{
			reports.GET("/summary", fcuHandler.GetReportSummary)
			reports.GET("/trends", fcuHandler.GetTrendAnalysis)
			reports.GET("/effectiveness", fcuHandler.GetDetectionEffectiveness)
			reports.GET("/compliance", fcuHandler.GetComplianceReport)
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
