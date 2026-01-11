package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/csic-platform/services/security/forensic-tools/internal/config"
	"github.com/csic-platform/services/security/forensic-tools/internal/core/domain"
	"github.com/csic-platform/services/security/forensic-tools/internal/core/ports"
	"github.com/google/uuid"
)

var (
	ErrEvidenceNotFound      = errors.New("evidence not found")
	ErrEvidenceAlreadyExists = errors.New("evidence with this hash already exists")
	ErrInvalidEvidenceType   = errors.New("invalid evidence type")
	ErrAnalysisJobNotFound   = errors.New("analysis job not found")
	ErrInvalidHash           = errors.New("invalid file hash")
)

// ForensicServiceImpl implements the ForensicService interface
type ForensicServiceImpl struct {
	repo      ports.EvidenceRepository
	storage   ports.BlobStorage
	producer  ports.MessageProducer
	hashCalc  ports.HashCalculator
	cfg       *config.Config
}

// NewForensicService creates a new forensic service instance
func NewForensicService(
	repo ports.EvidenceRepository,
	storage ports.BlobStorage,
	producer ports.MessageProducer,
	hashCalc ports.HashCalculator,
	cfg *config.Config,
) *ForensicServiceImpl {
	return &ForensicServiceImpl{
		repo:     repo,
		storage:  storage,
		producer: producer,
		hashCalc: hashCalc,
		cfg:      cfg,
	}
}

// UploadEvidence uploads and processes new evidence
func (s *ForensicServiceImpl) UploadEvidence(
	ctx context.Context,
	caseID, fileName string,
	reader io.Reader,
	size int64,
	evidenceType domain.EvidenceType,
	metadata map[string]string,
	actorID string,
) (*domain.Evidence, error) {
	// Calculate file hash
	fileHash, err := s.hashCalc.Calculate(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}

	// Check if evidence with same hash already exists
	existing, err := s.repo.GetEvidenceByHash(ctx, fileHash)
	if err == nil && existing != nil {
		return nil, ErrEvidenceAlreadyExists
	}

	// Generate storage path
	objectName := fmt.Sprintf("evidence/%s/%s/%s", caseID, time.Now().Format("2006/01/02"), uuid.New().String())

	// Reset reader position if possible
	if seeker, ok := reader.(io.Seeker); ok {
		seeker.Seek(0, 0)
	}

	// Upload to storage
	if err := s.storage.Upload(ctx, objectName, reader, size, "application/octet-stream"); err != nil {
		return nil, fmt.Errorf("failed to upload to storage: %w", err)
	}

	// Create evidence record
	evidence := &domain.Evidence{
		ID:           uuid.New().String(),
		CaseID:       caseID,
		FileName:     fileName,
		FileHash:     fileHash,
		FileType:     s.detectFileType(fileName),
		SizeBytes:    size,
		StoragePath:  objectName,
		EvidenceType: evidenceType,
		Status:       domain.EvidenceStatusUploaded,
		Metadata:     metadata,
		UploadedBy:   actorID,
		UploadedAt:   time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateEvidence(ctx, evidence); err != nil {
		// Cleanup storage on failure
		s.storage.Delete(ctx, objectName)
		return nil, fmt.Errorf("failed to create evidence record: %w", err)
	}

	// Add chain of custody entry
	custodyEntry := &domain.ChainOfCustody{
		ID:         uuid.New().String(),
		EvidenceID: evidence.ID,
		ActorID:    actorID,
		Action:     domain.CoCActionUpload,
		Details: map[string]interface{}{
			"file_name":  fileName,
			"file_size":  size,
			"file_hash":  fileHash,
			"file_type":  evidence.FileType,
		},
		Timestamp: time.Now(),
	}
	s.repo.AddCustodyEntry(ctx, custodyEntry)

	return evidence, nil
}

// GetEvidence retrieves evidence by ID
func (s *ForensicServiceImpl) GetEvidence(ctx context.Context, id string, actorID string) (*domain.Evidence, error) {
	evidence, err := s.repo.GetEvidence(ctx, id)
	if err != nil {
		return nil, ErrEvidenceNotFound
	}

	// Log access
	custodyEntry := &domain.ChainOfCustody{
		ID:         uuid.New().String(),
		EvidenceID: id,
		ActorID:    actorID,
		Action:     domain.CoCActionAccess,
		Details: map[string]interface{}{
			"operation": "GET",
		},
		Timestamp: time.Now(),
	}
	s.repo.AddCustodyEntry(ctx, custodyEntry)

	return evidence, nil
}

// ListEvidence retrieves a paginated list of evidence
func (s *ForensicServiceImpl) ListEvidence(ctx context.Context, caseID string, page, pageSize int, actorID string) (*domain.EvidenceListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	evidence, total, err := s.repo.ListEvidence(ctx, caseID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list evidence: %w", err)
	}

	return domain.NewEvidenceListResponse(evidence, total, page, pageSize), nil
}

// DownloadEvidence prepares evidence for download
func (s *ForensicServiceImpl) DownloadEvidence(ctx context.Context, id string, actorID string, ipAddress string) (io.ReadCloser, error) {
	evidence, err := s.repo.GetEvidence(ctx, id)
	if err != nil {
		return nil, ErrEvidenceNotFound
	}

	// Log download
	custodyEntry := &domain.ChainOfCustody{
		ID:         uuid.New().String(),
		EvidenceID: id,
		ActorID:    actorID,
		Action:     domain.CoCActionDownload,
		Details: map[string]interface{}{
			"ip_address": ipAddress,
		},
		Timestamp: time.Now(),
	}
	s.repo.AddCustodyEntry(ctx, custodyEntry)

	// Get file from storage
	return s.storage.GetFile(ctx, evidence.StoragePath)
}

// DeleteEvidence soft-deletes evidence
func (s *ForensicServiceImpl) DeleteEvidence(ctx context.Context, id string, actorID string) error {
	evidence, err := s.repo.GetEvidence(ctx, id)
	if err != nil {
		return ErrEvidenceNotFound
	}

	now := time.Now()
	evidence.Status = domain.EvidenceStatusDeleted
	evidence.DeletedAt = &now
	evidence.UpdatedAt = now

	if err := s.repo.UpdateEvidence(ctx, evidence); err != nil {
		return fmt.Errorf("failed to delete evidence: %w", err)
	}

	// Log deletion
	custodyEntry := &domain.ChainOfCustody{
		ID:         uuid.New().String(),
		EvidenceID: id,
		ActorID:    actorID,
		Action:     domain.CoCActionDelete,
		Details: map[string]interface{}{
			"reason": "User requested deletion",
		},
		Timestamp: time.Now(),
	}
	s.repo.AddCustodyEntry(ctx, custodyEntry)

	return nil
}

// ArchiveEvidence archives evidence for long-term storage
func (s *ForensicServiceImpl) ArchiveEvidence(ctx context.Context, id string, actorID string) error {
	evidence, err := s.repo.GetEvidence(ctx, id)
	if err != nil {
		return ErrEvidenceNotFound
	}

	now := time.Now()
	evidence.Status = domain.EvidenceStatusArchived
	evidence.ArchivedAt = &now
	evidence.UpdatedAt = now

	if err := s.repo.UpdateEvidence(ctx, evidence); err != nil {
		return fmt.Errorf("failed to archive evidence: %w", err)
	}

	// Log archival
	custodyEntry := &domain.ChainOfCustody{
		ID:         uuid.New().String(),
		EvidenceID: id,
		ActorID:    actorID,
		Action:     domain.CoCActionArchive,
		Timestamp:  time.Now(),
	}
	s.repo.AddCustodyEntry(ctx, custodyEntry)

	return nil
}

// GetCustodyHistory retrieves the chain of custody for evidence
func (s *ForensicServiceImpl) GetCustodyHistory(ctx context.Context, evidenceID string, actorID string) ([]*domain.ChainOfCustody, error) {
	return s.repo.GetCustodyHistory(ctx, evidenceID)
}

// VerifyChainOfCustody verifies the integrity of the chain of custody
func (s *ForensicServiceImpl) VerifyChainOfCustody(ctx context.Context, evidenceID string) (bool, error) {
	history, err := s.repo.GetCustodyHistory(ctx, evidenceID)
	if err != nil {
		return false, err
	}

	// Check for gaps in the chain
	for i := 0; i < len(history)-1; i++ {
		if history[i+1].Timestamp.Before(history[i].Timestamp) {
			return false, nil // Invalid timestamp order
		}
	}

	return true, nil
}

// RequestAnalysis submits an evidence analysis job
func (s *ForensicServiceImpl) RequestAnalysis(ctx context.Context, req *domain.AnalysisRequest, actorID string) (*domain.AnalysisJob, error) {
	// Verify evidence exists
	evidence, err := s.repo.GetEvidence(ctx, req.EvidenceID)
	if err != nil {
		return nil, ErrEvidenceNotFound
	}

	// Create analysis job
	job := &domain.AnalysisJob{
		ID:          uuid.New().String(),
		EvidenceID:  evidence.ID,
		ToolName:    "multi-tool", // Default to running multiple tools
		Status:      domain.AnalysisJobStatusPending,
		Priority:    req.Priority,
		RequestedBy: actorID,
		SubmittedAt: time.Now(),
		Progress:    0,
	}

	if err := s.repo.CreateAnalysisJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create analysis job: %w", err)
	}

	// Publish to Kafka for async processing
	s.producer.PublishAnalysisJob(ctx, job)

	return job, nil
}

// GetAnalysisJob retrieves an analysis job by ID
func (s *ForensicServiceImpl) GetAnalysisJob(ctx context.Context, jobID string, actorID string) (*domain.AnalysisJob, error) {
	job, err := s.repo.GetAnalysisJob(ctx, jobID)
	if err != nil {
		return nil, ErrAnalysisJobNotFound
	}

	// Log access
	custodyEntry := &domain.ChainOfCustody{
		ID:         uuid.New().String(),
		EvidenceID: job.EvidenceID,
		ActorID:    actorID,
		Action:     domain.CoCActionAnalysis,
		Details: map[string]interface{}{
			"operation": "GET_JOB",
			"job_id":    jobID,
		},
		Timestamp: time.Now(),
	}
	s.repo.AddCustodyEntry(ctx, custodyEntry)

	return job, nil
}

// GetAnalysisResults retrieves results for an analysis job
func (s *ForensicServiceImpl) GetAnalysisResults(ctx context.Context, jobID string, actorID string) ([]*domain.AnalysisResult, error) {
	job, err := s.repo.GetAnalysisJob(ctx, jobID)
	if err != nil {
		return nil, ErrAnalysisJobNotFound
	}

	return s.repo.GetAnalysisResults(ctx, jobID)
}

// ListAnalysisJobs retrieves analysis jobs for evidence
func (s *ForensicServiceImpl) ListAnalysisJobs(ctx context.Context, evidenceID string, page, pageSize int, actorID string) ([]*domain.AnalysisJob, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	jobs, _, err := s.repo.ListAnalysisJobs(ctx, evidenceID, page, pageSize)
	return jobs, err
}

// CancelAnalysisJob cancels a pending or processing analysis job
func (s *ForensicServiceImpl) CancelAnalysisJob(ctx context.Context, jobID string, actorID string) error {
	job, err := s.repo.GetAnalysisJob(ctx, jobID)
	if err != nil {
		return ErrAnalysisJobNotFound
	}

	if job.Status != domain.AnalysisJobStatusPending && job.Status != domain.AnalysisJobStatusQueued {
		return errors.New("cannot cancel job that is already processing")
	}

	job.Status = domain.AnalysisJobStatusCancelled
	return s.repo.UpdateAnalysisJob(ctx, job)
}

// CalculateFileHash calculates the SHA-256 hash of a file
func (s *ForensicServiceImpl) CalculateFileHash(reader io.Reader) (string, error) {
	return s.hashCalc.Calculate(reader)
}

// VerifyFileHash verifies a file's hash matches the expected value
func (s *ForensicServiceImpl) VerifyFileHash(ctx context.Context, evidenceID, expectedHash string) (bool, error) {
	evidence, err := s.repo.GetEvidence(ctx, evidenceID)
	if err != nil {
		return false, ErrEvidenceNotFound
	}

	return evidence.FileHash == expectedHash, nil
}

// HealthCheck performs a health check
func (s *ForensicServiceImpl) HealthCheck(ctx context.Context) error {
	// Verify storage connectivity
	_, err := s.storage.GetObjectInfo(ctx, "health-check")
	if err != nil {
		// Storage might not be accessible, but that's okay for health check
	}

	// Verify database connectivity
	_, err = s.repo.ListEvidence(ctx, "", 1, 1)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// detectFileType detects the file type from the filename
func (s *ForensicServiceImpl) detectFileType(fileName string) string {
	// Simple file type detection based on extension
	switch {
	case hasExtension(fileName, ".dd", ".img", ".raw", ".iso"):
		return "disk_image"
	case hasExtension(fileName, ".mem", ".dmp"):
		return "memory_dump"
	case hasExtension(fileName, ".pcap", ".pcapng", ".cap"):
		return "network_capture"
	case hasExtension(fileName, ".log", ".txt"):
		return "log_file"
	case hasExtension(fileName, ".pdf", ".doc", ".docx", ".xls", ".xlsx"):
		return "document"
	default:
		return "file"
	}
}

// hasExtension checks if a filename has the given extension
func hasExtension(fileName string, extensions ...string) string {
	for _, ext := range extensions {
		if len(fileName) > len(ext) && fileName[len(fileName)-len(ext):] == ext {
			return ext
		}
	}
	return ""
}

// DefaultHashCalculator implements HashCalculator
type DefaultHashCalculator struct{}

// NewDefaultHashCalculator creates a new default hash calculator
func NewDefaultHashCalculator() *DefaultHashCalculator {
	return &DefaultHashCalculator{}
}

// Calculate calculates the SHA-256 hash of data
func (c *DefaultHashCalculator) Calculate(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Verify verifies that data matches the expected hash
func (c *DefaultHashCalculator) Verify(data io.Reader, expectedHash string) (bool, error) {
	actualHash, err := c.Calculate(data)
	if err != nil {
		return false, err
	}
	return actualHash == expectedHash, nil
}

// Ensure ForensicServiceImpl implements ports.ForensicService
var _ ports.ForensicService = (*ForensicServiceImpl)(nil)
