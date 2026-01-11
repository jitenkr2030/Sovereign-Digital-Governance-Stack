package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic/mining-control/internal/config"
	"github.com/csic/mining-control/internal/handler"
	"github.com/csic/mining-control/internal/repository"
	"github.com/csic/mining-control/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "internal/config/config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize repository layer
	poolRepo, err := repository.NewPostgresMiningPoolRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize mining pool repository: %v", err)
	}
	defer poolRepo.Close()

	machineRepo, err := repository.NewPostgresMachineRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize machine repository: %v", err)
	}
	defer machineRepo.Close()

	energyRepo, err := repository.NewTimescaleEnergyRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize energy repository: %v", err)
	}
	defer energyRepo.Close()

	hashRepo, err := repository.NewTimescaleHashRateRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize hash rate repository: %v", err)
	}
	defer hashRepo.Close()

	violationRepo, err := repository.NewPostgresViolationRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize violation repository: %v", err)
	}
	defer violationRepo.Close()

	complianceRepo, err := repository.NewPostgresComplianceRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize compliance repository: %v", err)
	}
	defer complianceRepo.Close()

	// Initialize service layer
	registrationSvc := service.NewRegistrationService(poolRepo, machineRepo, complianceRepo)
	monitoringSvc := service.NewMonitoringService(energyRepo, hashRepo, poolRepo, machineRepo, violationRepo)
	enforcementSvc := service.NewEnforcementService(poolRepo, violationRepo, complianceRepo)
	reportingSvc := service.NewReportingService(poolRepo, energyRepo, hashRepo, violationRepo)

	// Initialize handlers
	httpHandler := handler.NewHTTPHandler(registrationSvc, monitoringSvc, enforcementSvc, reportingSvc)

	// Setup Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "mining-control",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	api := router.Group("/api/v1/mining")
	{
		// Pool registration endpoints
		api.POST("/pools", httpHandler.RegisterMiningPool)
		api.GET("/pools/:id", httpHandler.GetMiningPool)
		api.PUT("/pools/:id", httpHandler.UpdateMiningPool)
		api.POST("/pools/:id/suspend", httpHandler.SuspendMiningPool)
		api.POST("/pools/:id/activate", httpHandler.ActivateMiningPool)
		api.DELETE("/pools/:id", httpHandler.RevokeMiningPoolLicense)

		// Machine registration endpoints
		api.POST("/machines", httpHandler.RegisterMiningMachine)
		api.POST("/machines/batch", httpHandler.BatchRegisterMachines)
		api.GET("/machines/:id", httpHandler.GetMiningMachine)
		api.GET("/pools/:pool_id/machines", httpHandler.GetPoolMachines)
		api.PUT("/machines/:id", httpHandler.UpdateMiningMachine)
		api.DELETE("/machines/:id", httpHandler.DecommissionMachine)

		// Telemetry endpoints
		api.POST("/telemetry/energy", httpHandler.ReportEnergyConsumption)
		api.POST("/telemetry/energy/batch", httpHandler.BatchReportEnergyConsumption)
		api.POST("/telemetry/hashrate", httpHandler.ReportHashRate)
		api.POST("/telemetry/hashrate/batch", httpHandler.BatchReportHashRate)

		// Monitoring endpoints
		api.GET("/pools/:pool_id/energy-usage", httpHandler.GetEnergyUsage)
		api.GET("/pools/:pool_id/hashrate", httpHandler.GetHashRateMetrics)
		api.GET("/pools/:pool_id/compliance-status", httpHandler.GetComplianceStatus)

		// Violation endpoints
		api.GET("/violations", httpHandler.ListViolations)
		api.GET("/violations/open", httpHandler.GetOpenViolations)
		api.GET("/violations/:id", httpHandler.GetViolation)
		api.POST("/violations/:id/resolve", httpHandler.ResolveViolation)

		// Compliance endpoints
		api.GET("/pools/:pool_id/certificate", httpHandler.GetComplianceCertificate)
		api.POST("/pools/:pool_id/certificate/renew", httpHandler.RenewComplianceCertificate)

		// Reporting endpoints
		api.GET("/reports/regional-load", httpHandler.GetRegionalEnergyLoad)
		api.GET("/reports/carbon-footprint/:pool_id", httpHandler.GetCarbonFootprintReport)
		api.GET("/reports/summary/:pool_id", httpHandler.GetMiningSummaryReport)
		api.GET("/dashboard/stats", httpHandler.GetDashboardStats)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting Mining Control Service on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start background compliance checker
	go monitoringSvc.StartComplianceChecker()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop compliance checker
	monitoringSvc.StopComplianceChecker()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
