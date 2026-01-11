package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/robfig/cron/v3"

	"neam-platform/audit-service/config"
	"neam-platform/audit-service/crypto"
	"neam-platform/audit-service/handlers"
	"neam-platform/audit-service/integration"
	"neam-platform/audit-service/middleware"
	"neam-platform/audit-service/repository"
	"neam-platform/audit-service/services"
)

// @title NEAM Audit, Security & Sovereignty Layer
// @description Immutable audit trails, evidence-grade data retention, and zero-trust access control
// @version 1.0.0
// @host localhost:9090
// @BasePath /api/v1

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database connection pool
	pool, err := repository.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Initialize cryptographic services
	hashService := crypto.NewHashService()
	signatureService, err := crypto.NewSignatureService(cfg.HSMEnabled, cfg.HSMPin, cfg.HSMKeyLabel)
	if err != nil {
		log.Printf("Warning: HSM not available, using software signing: %v", err)
		signatureService = crypto.NewSoftwareSignatureService()
	}

	// Initialize Merkle Tree service
	merkleService := services.NewMerkleService(hashService)

	// Initialize WORM storage
	wormStorage := services.NewWORMStorage(cfg.WORMStoragePath, hashService, signatureService)

	// Initialize OpenSearch client for SIEM
	opensearchClient, err := integration.NewOpenSearchClient(cfg.OpenSearchURL, cfg.OpenSearchIndex)
	if err != nil {
		log.Printf("Warning: OpenSearch not available: %v", err)
		opensearchClient = nil
	}

	// Initialize audit repository
	auditRepo := repository.NewAuditRepository(pool)

	// Initialize sealing service
	sealingService := services.NewSealingService(hashService, signatureService, merkleService, wormStorage)

	// Initialize access control service
	accessControl := services.NewAccessControlService()

	// Initialize Keycloak client if enabled
	var keycloakClient *integration.KeycloakClient
	if cfg.KeycloakEnabled {
		keycloakClient, err = integration.NewKeycloakClient(
			cfg.KeycloakURL, cfg.KeycloakRealm, cfg.KeycloakClientID, cfg.KeycloakClientSecret,
		)
		if err != nil {
			log.Printf("Warning: Keycloak not available: %v", err)
		}
	}

	// Initialize audit service
	auditService := services.NewAuditService(
		auditRepo, sealingService, wormStorage, opensearchClient, accessControl,
	)

	// Initialize repositories
	userRepo := repository.NewUserRepository(pool)
	forensicRepo := repository.NewForensicRepository(pool)

	// Initialize handlers
	auditHandler := handlers.NewAuditHandler(auditService)
	authHandler := handlers.NewAuthHandler(userRepo, accessControl, keycloakClient, signatureService)
	forensicHandler := handlers.NewForensicHandler(forensicRepo, sealingService, hashService, accessControl)

	// Initialize router
	router := handlers.NewRouter(auditHandler, authHandler, forensicHandler)

	// Add middleware
	router.Use(middleware.RequestLogger())
	router.Use(middleware.Recovery())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.ZeroTrustAuth(keycloakClient, accessControl))

	// Initialize cron job for Merkle tree sealing
	cronService := cron.New()
	if _, err := cronService.AddFunc(cfg.MerkleTreeSealingCron, func() {
		if err := sealingService.SealMerkleTree(ctx); err != nil {
			log.Printf("Failed to seal Merkle tree: %v", err)
		}
	}); err != nil {
		log.Printf("Warning: Failed to schedule Merkle tree sealing: %v", err)
	}
	cronService.Start()
	defer cronService.Stop()

	// Start periodic integrity verification
	go startIntegrityVerification(ctx, auditService, cfg.IntegrityCheckInterval)

	// Start server
	serverAddr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("Starting Audit, Security & Sovereignty Layer on %s", serverAddr)
	
	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gracefully...")
		cancel()
		cronService.Stop()
		if err := router.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	if err := router.Start(serverAddr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// startIntegrityVerification runs periodic integrity checks
func startIntegrityVerification(ctx context.Context, auditService services.AuditService, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := auditService.VerifyIntegrity(ctx); err != nil {
				log.Printf("Integrity verification failed: %v", err)
			}
		}
	}
}

// AuditEvent represents a single audit event
type AuditEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	ActorID     string                 `json:"actor_id"`
	ActorType   string                 `json:"actor_type"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id"`
	Outcome     string                 `json:"outcome"`
	Details     map[string]interface{} `json:"details,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Location    string                 `json:"location"`
	SessionID   string                 `json:"session_id"`
	IntegrityHash string               `json:"integrity_hash"`
	Sealed      bool                   `json:"sealed"`
	SealedAt    *time.Time             `json:"sealed_at,omitempty"`
	MerklePath  []string               `json:"merkle_path,omitempty"`
	PreviousHash string               `json:"previous_hash"`
}

// AuditLog represents the complete audit log structure
type AuditLog struct {
	Events      []AuditEvent   `json:"events"`
	Seal        SealRecord     `json:"seal"`
	Metadata    LogMetadata    `json:"metadata"`
}

// SealRecord represents the cryptographic seal
type SealRecord struct {
	Timestamp   time.Time        `json:"timestamp"`
	RootHash    string           `json:"root_hash"`
	Signature   string           `json:"signature"`
	SealedBy    string           `json:"sealed_by"`
	Version     string           `json:"version"`
	ChainHashes []string         `json:"chain_hashes"`
	Witnesses   []WitnessRecord  `json:"witnesses,omitempty"`
}

// WitnessRecord represents an external witness to the seal
type WitnessRecord struct {
	WitnessType string    `json:"witness_type"`
	WitnessID   string    `json:"witness_id"`
	Timestamp   time.Time `json:"timestamp"`
	Signature   string    `json:"signature"`
}

// LogMetadata contains metadata about the audit log
type LogMetadata struct {
	CreatedAt     time.Time `json:"created_at"`
	TotalEvents   int       `json:"total_events"`
	FirstEvent    time.Time `json:"first_event"`
	LastEvent     time.Time `json:"last_event"`
	StorageType   string    `json:"storage_type"`
	RetentionDays int       `json:"retention_days"`
	Compliance    []string  `json:"compliance_standards"`
}

// User represents a user in the access control system
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	Permissions  []string  `json:"permissions"`
	Groups       []string  `json:"groups"`
	Active       bool      `json:"active"`
	MFEnabled    bool      `json:"mfa_enabled"`
	CreatedAt    time.Time `json:"created_at"`
	LastLogin    time.Time `json:"last_login,omitempty"`
	LockedUntil  time.Time `json:"locked_until,omitempty"`
}

// Policy represents an access control policy
type Policy struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Subjects    []string `json:"subjects"`
	Resources   []string `json:"resources"`
	Actions     []string `json:"actions"`
	Conditions  map[string]interface{} `json:"conditions"`
	Effect      string   `json:"effect"`
	Priority    int      `json:"priority"`
	Active      bool     `json:"active"`
}

// ForensicRequest represents a forensic investigation request
type ForensicRequest struct {
	ID            string                 `json:"id"`
	RequesterID   string                 `json:"requester_id"`
	RequesterRole string                 `json:"requester_role"`
	Query         ForensicQuery          `json:"query"`
	Justification string                 `json:"justification"`
	Status        string                 `json:"status"`
	ApprovedBy    string                 `json:"approved_by,omitempty"`
	ApprovedAt    time.Time              `json:"approved_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	CompletedAt   time.Time              `json:"completed_at,omitempty"`
}

// ForensicQuery represents a query for forensic investigation
type ForensicQuery struct {
	EventTypes   []string    `json:"event_types,omitempty"`
	ActorIDs     []string    `json:"actor_ids,omitempty"`
	Resources    []string    `json:"resources,omitempty"`
	TimeRange    TimeRange   `json:"time_range"`
	Locations    []string    `json:"locations,omitempty"`
	Outcomes     []string    `json:"outcomes,omitempty"`
	Keywords     []string    `json:"keywords,omitempty"`
	IncludeChain bool        `json:"include_chain"`
	IncludeHash  bool        `json:"include_hash"`
}

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// VerificationResult represents the result of an integrity verification
type VerificationResult struct {
	Verified       bool              `json:"verified"`
	Timestamp      time.Time         `json:"timestamp"`
	EventsVerified int               `json:"events_verified"`
	EventsFailed   int               `json:"events_failed"`
	FailedEvents   []FailedEvent     `json:"failed_events,omitempty"`
	RootHash       string            `json:"root_hash"`
	ComputedHash   string            `json:"computed_hash"`
	Witnesses      []WitnessRecord   `json:"witnesses,omitempty"`
}

// FailedEvent represents an event that failed verification
type FailedEvent struct {
	EventID     string    `json:"event_id"`
	ExpectedHash string   `json:"expected_hash"`
	ActualHash  string    `json:"actual_hash"`
	Timestamp   time.Time `json:"timestamp"`
}

// calculateHash calculates SHA-256 hash for data
func calculateHash(data interface{}) string {
	jsonBytes, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:])
}

// GenerateID generates a unique ID
func GenerateID() string {
	return uuid.New().String()
}

func init() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}
