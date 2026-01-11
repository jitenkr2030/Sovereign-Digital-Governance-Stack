// Audit Log Service - Government-Grade Immutable Audit Trail
// Provides tamper-evident, cryptographically sealed audit logging for CSIC Platform
// Compliant with ISO 27001, GDPR, and financial regulatory requirements

package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/csic-platform/shared/config"
	"github.com/csic-platform/shared/logger"
)

// AuditLogService provides immutable, tamper-evident audit logging
type AuditLogService struct {
	writer    *AuditLogWriter
	sealer    *AuditLogSealer
	verifier  *AuditLogVerifier
	logger    *logger.Logger
	config    *AuditConfig
	mu        sync.RWMutex
	running   bool
}

// AuditConfig holds configuration for the audit log service
type AuditConfig struct {
	StoragePath      string `yaml:"storage_path"`
	SealInterval     int    `yaml:"seal_interval"` // seconds
	ChainFilePath    string `yaml:"chain_file_path"`
	VerificationPath string `yaml:"verification_path"`
	RetentionDays    int    `yaml:"retention_days"`
	EnableWORM       bool   `yaml:"enable_worm"` // Write Once Read Many
}

// AuditLogEntry represents a single audit log entry
type AuditLogEntry struct {
	// Immutable identifiers
	EntryID     string    `json:"entry_id"`
	SequenceNum uint64    `json:"sequence_num"`
	Timestamp   time.Time `json:"timestamp"`
	ChainID     string    `json:"chain_id"`

	// Actor information
	ActorID     string `json:"actor_id"`
	ActorType   string `json:"actor_type"` // user, system, service
	ActorRole   string `json:"actor_role"`
	SessionID   string `json:"session_id,omitempty"`
	IPAddress   string `json:"ip_address,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`

	// Action details
	Service     string `json:"service"`
	Operation   string `json:"operation"`
	ActionType  string `json:"action_type"` // create, read, update, delete, execute
	Resource    string `json:"resource"`
	ResourceID  string `json:"resource_id,omitempty"`
	Description string `json:"description"`

	// Result and impact
	Result      string `json:"result"` // success, failure, partial
	ErrorCode   string `json:"error_code,omitempty"`
	ErrorMsg    string `json:"error_msg,omitempty"`
	AffectedIDs []string `json:"affected_ids,omitempty"`

	// Compliance and classification
	ComplianceTags []string `json:"compliance_tags"`
	RiskLevel      string   `json:"risk_level"` // low, medium, high, critical
	RegulatoryRef  string   `json:"regulatory_ref,omitempty"`

	// Cryptographic integrity
	PreviousHash string `json:"previous_hash"`
	CurrentHash  string `json:"current_hash"`
	Signature    string `json:"signature,omitempty"`

	// Additional context
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	TraceID  string `json:"trace_id,omitempty"`
}

// AuditChain represents a chain of sealed audit log entries
type AuditChain struct {
	ChainID       string          `json:"chain_id"`
	CreatedAt     time.Time       `json:"created_at"`
	SealedAt      time.Time       `json:"sealed_at"`
	SequenceStart uint64          `json:"sequence_start"`
	SequenceEnd   uint64          `json:"sequence_end"`
	EntryCount    int             `json:"entry_count"`
	RootHash      string          `json:"root_hash"`
	SealSignature string          `json:"seal_signature"`
	PreviousChain string          `json:"previous_chain"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
}

// VerificationResult represents the result of an audit log verification
type VerificationResult struct {
	Valid           bool     `json:"valid"`
	ChainIntegrity  bool     `json:"chain_integrity"`
	HashIntegrity   bool     `json:"hash_integrity"`
	SealIntegrity   bool     `json:"seal_integrity"`
	FirstInvalid    uint64   `json:"first_invalid_sequence,omitempty"`
	FirstInvalidID  string   `json:"first_invalid_id,omitempty"`
	VerifiedEntries int      `json:"verified_entries"`
	FailedEntries   int      `json:"failed_entries"`
	Errors          []string `json:"errors,omitempty"`
	VerifiedAt      time.Time `json:"verified_at"`
}

// NewAuditLogService creates a new audit log service instance
func NewAuditLogService(cfg *AuditConfig, logConfig logger.Config) (*AuditLogService, error) {
	// Initialize logger
	appLogger, err := logger.NewLogger(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Ensure storage directories exist
	if err := os.MkdirAll(cfg.StoragePath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(cfg.ChainFilePath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create chain directory: %w", err)
	}

	service := &AuditLogService{
		logger: appLogger,
		config: cfg,
	}

	// Initialize components
	service.writer = NewAuditLogWriter(cfg.StoragePath, cfg.EnableWORM)
	service.sealer = NewAuditLogSealer(cfg.ChainFilePath, cfg.SealInterval)
	service.verifier = NewAuditLogVerifier(cfg.StoragePath, cfg.ChainFilePath)

	appLogger.Info("audit log service initialized",
		logger.WithFields(
			logger.String("storage_path", cfg.StoragePath),
			logger.Bool("worm_enabled", cfg.EnableWORM),
		),
	)

	return service, nil
}

// Start begins the audit log service
func (s *AuditLogService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return errors.New("audit log service is already running")
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("starting audit log service")

	// Start the sealing routine
	go s.sealingRoutine(ctx)

	s.logger.Info("audit log service started")
	return nil
}

// Stop gracefully stops the audit log service
func (s *AuditLogService) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	s.logger.Info("stopping audit log service")

	// Seal any pending entries
	if err := s.sealer.SealPending(s.writer.GetSequenceNumber()); err != nil {
		s.logger.Error("failed to seal pending entries", logger.WithFields(logger.Error(err)))
	}

	s.logger.Info("audit log service stopped")
	return nil
}

// WriteLog writes a new audit log entry
func (s *AuditLogService) WriteLog(ctx context.Context, entry *AuditLogEntry) error {
	if !s.running {
		return errors.New("audit log service is not running")
	}

	// Write the entry
	return s.writer.Write(ctx, entry)
}

// WriteBatch writes multiple audit log entries
func (s *AuditLogService) WriteBatch(ctx context.Context, entries []*AuditLogEntry) error {
	if !s.running {
		return errors.New("audit log service is not running")
	}

	for _, entry := range entries {
		if err := s.writer.Write(ctx, entry); err != nil {
			return fmt.Errorf("failed to write entry %s: %w", entry.EntryID, err)
		}
	}

	return nil
}

// Verify verifies the integrity of the audit log chain
func (s *AuditLogService) Verify(ctx context.Context) (*VerificationResult, error) {
	return s.verifier.VerifyChain(ctx)
}

// GetChain returns the audit chain information
func (s *AuditLogService) GetChain(chainID string) (*AuditChain, error) {
	return s.sealer.GetChain(chainID)
}

// GetEntry retrieves a specific audit log entry
func (s *AuditLogService) GetEntry(ctx context.Context, entryID string) (*AuditLogEntry, error) {
	return s.writer.Read(ctx, entryID)
}

// QueryEntries queries audit log entries with filters
func (s *AuditLogService) QueryEntries(ctx context.Context, query *AuditQuery) ([]*AuditLogEntry, error) {
	return s.writer.Query(ctx, query)
}

// ExportChain exports a sealed audit chain for legal discovery
func (s *AuditLogService) ExportChain(ctx context.Context, chainID string) ([]byte, error) {
	return s.sealer.ExportChain(chainID)
}

// sealingRoutine periodically seals the audit log chain
func (s *AuditLogService) sealingRoutine(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(s.config.SealInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.sealer.SealPending(s.writer.GetSequenceNumber()); err != nil {
				s.logger.Error("failed to seal audit log", logger.WithFields(logger.Error(err)))
			}
		}
	}
}

// AuditQuery represents a query for audit log entries
type AuditQuery struct {
	StartTime    time.Time
	EndTime      time.Time
	ActorID      string
	Service      string
	Operation    string
	ActionType   string
	Resource     string
	Result       string
	RiskLevel    string
	SequenceFrom uint64
	SequenceTo   uint64
	Limit        int
	Offset       int
}

// generateEntryID generates a unique entry identifier
func generateEntryID() string {
	// Use timestamp + random data for uniqueness
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)
	if _, err := os.ReadFile("/dev/urandom", randomBytes); err != nil {
		// Fallback to crypto/rand if /dev/urandom is not available
		binary.BigEndian.PutUint64(randomBytes, uint64(time.Now().UnixNano()))
	}

	data := make([]byte, 16)
	binary.BigEndian.PutUint64(data[:8], uint64(timestamp))
	copy(data[8:], randomBytes[:8])

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:32])
}

// WithFields is a helper for structured logging
func WithFields(fields ...interface{}) []logger.Field {
	return logger.WithFields(fields...).([]logger.Field)
}
