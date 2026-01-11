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

	"github.com/csic/wallet-governance/internal/config"
	"github.com/csic/wallet-governance/internal/handler"
	"github.com/csic/wallet-governance/internal/repository"
	"github.com/csic/wallet-governance/internal/service"
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
	walletRepo, err := repository.NewPostgresWalletRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize wallet repository: %v", err)
	}
	defer walletRepo.Close()

	signatureRepo, err := repository.NewPostgresSignatureRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize signature repository: %v", err)
	}
	defer signatureRepo.Close()

	blacklistRepo, err := repository.NewPostgresBlacklistRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize blacklist repository: %v", err)
	}
	defer blacklistRepo.Close()

	whitelistRepo, err := repository.NewPostgresWhitelistRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize whitelist repository: %v", err)
	}
	defer whitelistRepo.Close()

	freezeRepo, err := repository.NewPostgresWalletFreezeRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize freeze repository: %v", err)
	}
	defer freezeRepo.Close()

	auditRepo, err := repository.NewPostgresAuditRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize audit repository: %v", err)
	}
	defer auditRepo.Close()

	// Initialize HSM service
	hsmService, err := service.NewHSMService(cfg.HSM)
	if err != nil {
		log.Fatalf("Failed to initialize HSM service: %v", err)
	}

	// Initialize service layer
	walletSvc := service.NewWalletService(walletRepo, blacklistRepo, whitelistRepo, freezeRepo, auditRepo)
	signatureSvc := service.NewSignatureService(signatureRepo, walletRepo, hsmService, auditRepo)
	governanceSvc := service.NewGovernanceService(walletRepo, signatureSvc, hsmService, auditRepo)
	freezeSvc := service.NewFreezeService(walletRepo, freezeRepo, signatureSvc, auditRepo)
	complianceSvc := service.NewComplianceService(walletRepo, blacklistRepo, whitelistRepo, freezeRepo, auditRepo)

	// Initialize handlers
	httpHandler := handler.NewHTTPHandler(walletSvc, signatureSvc, governanceSvc, freezeSvc, complianceSvc)

	// Setup Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "wallet-governance",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	api := router.Group("/api/v1/wallet")
	{
		// Wallet management endpoints
		api.POST("/wallets", httpHandler.RegisterWallet)
		api.GET("/wallets/:id", httpHandler.GetWallet)
		api.PUT("/wallets/:id", httpHandler.UpdateWallet)
		api.DELETE("/wallets/:id", httpHandler.RevokeWallet)

		// Wallet type endpoints
		api.POST("/wallets/custodial", httpHandler.RegisterCustodialWallet)
		api.POST("/wallets/multi-sig", httpHandler.RegisterMultiSigWallet)
		api.POST("/wallets/exchange-hot", httpHandler.RegisterExchangeHotWallet)
		api.POST("/wallets/exchange-cold", httpHandler.RegisterExchangeColdWallet)

		// Multi-signature governance endpoints
		api.GET("/wallets/:id/signers", httpHandler.GetWalletSigners)
		api.POST("/wallets/:id/signers", httpHandler.AddSigner)
		api.DELETE("/wallets/:id/signers/:signer_id", httpHandler.RemoveSigner)
		api.PUT("/wallets/:id/signers/:signer_id/threshold", httpHandler.UpdateSignerThreshold)
		api.POST("/wallets/:id/propose-transaction", httpHandler.ProposeTransaction)
		api.GET("/wallets/:id/pending-transactions", httpHandler.GetPendingTransactions)
		api.POST("/wallets/:id/transactions/:tx_id/approve", httpHandler.ApproveTransaction)
		api.POST("/wallets/:id/transactions/:tx_id/reject", httpHandler.RejectTransaction)

		// Signing endpoints
		api.POST("/signatures/request", httpHandler.RequestSignature)
		api.GET("/signatures/:id", httpHandler.GetSignatureStatus)
		api.POST("/signatures/:id/verify", httpHandler.VerifySignature)

		// Blacklist endpoints
		api.POST("/blacklist/addresses", httpHandler.AddToBlacklist)
		api.DELETE("/blacklist/addresses/:address", httpHandler.RemoveFromBlacklist)
		api.GET("/blacklist/addresses", httpHandler.GetBlacklist)
		api.GET("/blacklist/addresses/:address/check", httpHandler.CheckBlacklist)
		api.POST("/blacklist/batch", httpHandler.BatchAddToBlacklist)

		// Whitelist endpoints
		api.POST("/whitelist/addresses", httpHandler.AddToWhitelist)
		api.DELETE("/whitelist/addresses/:address", httpHandler.RemoveFromWhitelist)
		api.GET("/whitelist/addresses", httpHandler.GetWhitelist)
		api.GET("/whitelist/addresses/:address/check", httpHandler.CheckWhitelist)
		api.POST("/whitelist/batch", httpHandler.BatchAddToWhitelist)

		// Freeze endpoints
		api.POST("/freeze", httpHandler.FreezeWallet)
		api.POST("/freeze/emergency", httpHandler.EmergencyFreeze)
		api.POST("/unfreeze", httpHandler.UnfreezeWallet)
		api.GET("/freeze/:wallet_id", httpHandler.GetFreezeStatus)
		api.GET("/freeze/active", httpHandler.GetActiveFreezes)
		api.GET("/freeze/history/:wallet_id", httpHandler.GetFreezeHistory)

		// Compliance endpoints
		api.GET("/compliance/wallets/:id", httpHandler.GetWalletComplianceStatus)
		api.POST("/compliance/check", httpHandler.CheckWalletCompliance)
		api.GET("/compliance/audit/:wallet_id", httpHandler.GetWalletAuditTrail)

		// Recovery endpoints
		api.POST("/recovery/request", httpHandler.RequestAssetRecovery)
		api.GET("/recovery/:id", httpHandler.GetRecoveryRequest)
		api.POST("/recovery/:id/approve", httpHandler.ApproveRecovery)
		api.POST("/recovery/:id/execute", httpHandler.ExecuteRecovery)

		// Reporting endpoints
		api.GET("/reports/summary", httpHandler.GetWalletSummary)
		api.GET("/reports/by-exchange", httpHandler.GetWalletsByExchange)
		api.GET("/reports/by-type", httpHandler.GetWalletsByType)
		api.GET("/reports/freeze-summary", httpHandler.GetFreezeSummary)
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
		log.Printf("Starting Wallet Governance Service on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start background tasks
	go governanceSvc.StartTransactionExpiryChecker()
	go freezeSvc.StartFreezeExpiryChecker()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop background tasks
	governanceSvc.StopTransactionExpiryChecker()
	freezeSvc.StopFreezeExpiryChecker()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
