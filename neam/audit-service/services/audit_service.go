package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"neam-platform/audit-service/crypto"
	"neam-platform/audit-service/models"
	"neam-platform/audit-service/repository"
)

// AuditService handles audit logging and retrieval
type AuditService interface {
	// LogEvent logs a new audit event
	LogEvent(ctx context.Context, event *models.AuditEvent) error

	// LogEventAsync logs an audit event asynchronously
	LogEventAsync(ctx context.Context, event *models.AuditEvent)

	// QueryEvents retrieves audit events based on query
	QueryEvents(ctx context.Context, query models.AuditQuery) ([]models.AuditEvent, error)

	// GetEvent retrieves a single audit event by ID
	GetEvent(ctx context.Context, eventID string) (*models.AuditEvent, error)

	// GetStatistics retrieves audit statistics
	GetStatistics(ctx context.Context) (*models.AuditStatistics, error)

	// VerifyIntegrity verifies the integrity of audit records
	VerifyIntegrity(ctx context.Context) (*models.VerificationReport, error)

	// SealEvents seals a batch of events
	SealEvents(ctx context.Context, eventIDs []string) (*models.SealRecord, error)

	// GetChain retrieves the integrity chain for an event
	GetChain(ctx context.Context, eventID string) ([]models.SealRecord, error)
}

// auditService implements AuditService
type auditService struct {
	repo          repository.AuditRepository
	sealing       SealingService
	wormStorage   WORMStorage
	opensearch    OpenSearchClient
	accessControl AccessControl
	hashService   crypto.HashService
	eventBuffer   chan *models.AuditEvent
	bufferWg      sync.WaitGroup
	bufferMu      sync.Mutex
	closed        bool
}

// NewAuditService creates a new AuditService
func NewAuditService(
	repo repository.AuditRepository,
	sealing SealingService,
	wormStorage WORMStorage,
	opensearch OpenSearchClient,
	accessControl AccessControl,
) AuditService {
	as := &auditService{
		repo:          repo,
		sealing:       sealing,
		wormStorage:   wormStorage,
		opensearch:    opensearch,
		accessControl: accessControl,
		hashService:   crypto.NewHashService(),
		eventBuffer:   make(chan *models.AuditEvent, 1000),
	}

	// Start async event processor
	for i := 0; i < 5; i++ {
		as.bufferWg.Add(1)
		go as.eventProcessor(i)
	}

	return as
}

// LogEvent logs a new audit event
func (s *auditService) LogEvent(ctx context.Context, event *models.AuditEvent) error {
	// Validate event
	if err := s.validateEvent(event); err != nil {
		return fmt.Errorf("invalid audit event: %w", err)
	}

	// Set timestamp and ID if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if event.ID == "" {
		event.ID = generateID()
	}

	// Set created time
	event.CreatedAt = time.Now().UTC()

	// Calculate integrity hash
	event.IntegrityHash = s.calculateIntegrityHash(event)

	// Get previous hash for chain continuity
	previousHash, err := s.repo.GetLastHash(ctx)
	if err != nil {
		log.Printf("Warning: Could not get previous hash: %v", err)
	}
	event.PreviousHash = previousHash

	// Set chain index
	chainIndex, err := s.repo.GetNextChainIndex(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain index: %w", err)
	}
	event.ChainIndex = chainIndex

	// Store in database
	if err := s.repo.Create(ctx, event); err != nil {
		return fmt.Errorf("failed to store audit event: %w", err)
	}

	// Send to OpenSearch for SIEM
	if s.opensearch != nil {
		if err := s.opensearch.IndexEvent(ctx, event); err != nil {
			log.Printf("Warning: Failed to index event in OpenSearch: %v", err)
		}
	}

	// Store in WORM storage
	if err := s.wormStorage.Store(ctx, event); err != nil {
		log.Printf("Warning: Failed to store in WORM: %v", err)
	}

	return nil
}

// LogEventAsync logs an audit event asynchronously
func (s *auditService) LogEventAsync(ctx context.Context, event *models.AuditEvent) {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()

	if s.closed {
		// Process synchronously if buffer is closed
		_ = s.LogEvent(ctx, event)
		return
	}

	select {
	case s.eventBuffer <- event:
	default:
		// Buffer full, process synchronously
		_ = s.LogEvent(ctx, event)
	}
}

// eventProcessor processes events from the buffer
func (s *auditService) eventProcessor(workerID int) {
	defer s.bufferWg.Done()

	for event := range s.eventBuffer {
		ctx := context.Background()
		if err := s.LogEvent(ctx, event); err != nil {
			log.Printf("Worker %d: Failed to process event %s: %v", workerID, event.ID, err)
		}
	}
}

// QueryEvents retrieves audit events based on query
func (s *auditService) QueryEvents(ctx context.Context, query models.AuditQuery) ([]models.AuditEvent, error) {
	// Apply access control
	if err := s.accessControl.CheckPermission(ctx, "audit", "read"); err != nil {
		return nil, err
	}

	// Set defaults
	if query.Limit == 0 {
		query.Limit = 100
	}
	if query.Limit > 1000 {
		query.Limit = 1000
	}
	if query.SortBy == "" {
		query.SortBy = "timestamp"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}

	return s.repo.Query(ctx, query)
}

// GetEvent retrieves a single audit event by ID
func (s *auditService) GetEvent(ctx context.Context, eventID string) (*models.AuditEvent, error) {
	// Apply access control
	if err := s.accessControl.CheckPermission(ctx, "audit", "read"); err != nil {
		return nil, err
	}

	event, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// GetStatistics retrieves audit statistics
func (s *auditService) GetStatistics(ctx context.Context) (*models.AuditStatistics, error) {
	// Apply access control
	if err := s.accessControl.CheckPermission(ctx, "audit", "read"); err != nil {
		return nil, err
	}

	return s.repo.GetStatistics(ctx)
}

// VerifyIntegrity verifies the integrity of audit records
func (s *auditService) VerifyIntegrity(ctx context.Context) (*models.VerificationReport, error) {
	// Apply access control
	if err := s.accessControl.CheckPermission(ctx, "audit", "verify"); err != nil {
		return nil, err
	}

	report := &models.VerificationReport{
		Timestamp:      time.Now().UTC(),
		EventsVerified: 0,
		EventsFailed:   0,
		FailedEvents:   []models.FailedEvent{},
	}

	// Get unsealed events
	events, err := s.repo.GetUnsealedEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get unsealed events: %w", err)
	}

	report.EventsVerified = int64(len(events))
	previousHash := ""

	for _, event := range events {
		// Verify integrity hash
		expectedHash := s.calculateIntegrityHash(event)
		if event.IntegrityHash != expectedHash {
			report.EventsFailed++
			report.FailedEvents = append(report.FailedEvents, models.FailedEvent{
				EventID:       event.ID,
				EventIndex:    event.ChainIndex,
				ExpectedHash:  expectedHash,
				ActualHash:    event.IntegrityHash,
				EventTime:     event.Timestamp,
				FailureReason: "integrity hash mismatch",
			})
		}

		// Verify chain continuity
		if previousHash != "" && event.PreviousHash != previousHash {
			report.EventsFailed++
			report.FailedEvents = append(report.FailedEvents, models.FailedEvent{
				EventID:       event.ID,
				EventIndex:    event.ChainIndex,
				EventTime:     event.Timestamp,
				FailureReason: "chain discontinuity",
			})
		}
		previousHash = event.IntegrityHash
	}

	// Verify sealed events using seal records
	sealRecords, err := s.repo.GetRecentSeals(ctx, 10)
	if err != nil {
		log.Printf("Warning: Failed to get recent seals: %v", err)
	} else {
		for _, seal := range sealRecords {
			if !seal.Verified {
				verified, err := s.sealing.VerifySeal(ctx, seal)
				if err != nil {
					log.Printf("Warning: Failed to verify seal %s: %v", seal.ID, err)
					continue
				}
				if verified {
					seal.Verified = true
					now := time.Now().UTC()
					seal.VerifiedAt = &now
					_ = s.repo.UpdateSeal(ctx, &seal)
				}
			}
		}
	}

	report.Verified = report.EventsFailed == 0
	report.ChainIntegrity = models.ChainIntegrityResult{
		Continuous: report.EventsFailed == 0,
		HashMatch:  report.EventsFailed == 0,
	}

	return report, nil
}

// SealEvents seals a batch of events
func (s *auditService) SealEvents(ctx context.Context, eventIDs []string) (*models.SealRecord, error) {
	// Apply access control
	if err := s.accessControl.CheckPermission(ctx, "audit", "seal"); err != nil {
		return nil, err
	}

	// Seal the events
	seal, err := s.sealing.SealEvents(ctx, eventIDs)
	if err != nil {
		return nil, err
	}

	// Update events as sealed
	now := time.Now().UTC()
	if err := s.repo.MarkSealed(ctx, eventIDs, &now); err != nil {
		return nil, fmt.Errorf("failed to mark events as sealed: %w", err)
	}

	// Store seal in WORM
	if err := s.wormStorage.StoreSeal(ctx, seal); err != nil {
		log.Printf("Warning: Failed to store seal in WORM: %v", err)
	}

	return seal, nil
}

// GetChain retrieves the integrity chain for an event
func (s *auditService) GetChain(ctx context.Context, eventID string) ([]models.SealRecord, error) {
	// Apply access control
	if err := s.accessControl.CheckPermission(ctx, "audit", "read"); err != nil {
		return nil, err
	}

	return s.repo.GetEventChain(ctx, eventID)
}

// validateEvent validates an audit event
func (s *auditService) validateEvent(event *models.AuditEvent) error {
	if event.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if event.ActorID == "" {
		return fmt.Errorf("actor_id is required")
	}
	if event.Action == "" {
		return fmt.Errorf("action is required")
	}
	if event.Resource == "" {
		return fmt.Errorf("resource is required")
	}
	if event.Outcome == "" {
		return fmt.Errorf("outcome is required")
	}
	return nil
}

// calculateIntegrityHash calculates the integrity hash for an event
func (s *auditService) calculateIntegrityHash(event *models.AuditEvent) string {
	// Create a copy without the hash fields
	eventCopy := *event
	eventCopy.IntegrityHash = ""
	eventCopy.MerkleProof = nil
	eventCopy.Sealed = false
	eventCopy.SealedAt = nil

	// Serialize and hash
	data, _ := json.Marshal(eventCopy)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Close shuts down the audit service
func (s *auditService) Close() {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()

	if s.closed {
		return
	}
	s.closed = true

	close(s.eventBuffer)
	s.bufferWg.Wait()
}

// generateID generates a unique ID
func generateID() string {
	b := make([]byte, 16)
	_, _ = crypto.RandRead(b)
	return hex.EncodeToString(b)
}
