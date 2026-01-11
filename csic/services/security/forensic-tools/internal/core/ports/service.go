package ports

import (
	"context"
	"io"

	"github.com/csic-platform/services/security/forensic-tools/internal/core/domain"
)

// ForensicService defines the interface for forensic tools business logic
type ForensicService interface {
	// Evidence management
	UploadEvidence(ctx context.Context, caseID, fileName string, reader io.Reader, size int64, evidenceType domain.EvidenceType, metadata map[string]string, actorID string) (*domain.Evidence, error)
	GetEvidence(ctx context.Context, id string, actorID string) (*domain.Evidence, error)
	ListEvidence(ctx context.Context, caseID string, page, pageSize int, actorID string) (*domain.EvidenceListResponse, error)
	DownloadEvidence(ctx context.Context, id string, actorID string, ipAddress string) (io.ReadCloser, error)
	DeleteEvidence(ctx context.Context, id string, actorID string) error
	ArchiveEvidence(ctx context.Context, id string, actorID string) error

	// Chain of custody
	GetCustodyHistory(ctx context.Context, evidenceID string, actorID string) ([]*domain.ChainOfCustody, error)
	VerifyChainOfCustody(ctx context.Context, evidenceID string) (bool, error)

	// Analysis operations
	RequestAnalysis(ctx context.Context, req *domain.AnalysisRequest, actorID string) (*domain.AnalysisJob, error)
	GetAnalysisJob(ctx context.Context, jobID string, actorID string) (*domain.AnalysisJob, error)
	GetAnalysisResults(ctx context.Context, jobID string, actorID string) ([]*domain.AnalysisResult, error)
	ListAnalysisJobs(ctx context.Context, evidenceID string, page, pageSize int, actorID string) ([]*domain.AnalysisJob, error)
	CancelAnalysisJob(ctx context.Context, jobID string, actorID string) error

	// Hash operations
	CalculateFileHash(reader io.Reader) (string, error)
	VerifyFileHash(ctx context.Context, evidenceID, expectedHash string) (bool, error)

	// Health check
	HealthCheck(ctx context.Context) error
}

// AnalysisEngine defines the interface for forensic analysis engines
type AnalysisEngine interface {
	// Tool identification
	Name() string
	Version() string
	IsAvailable() bool

	// Analysis execution
	Analyze(ctx context.Context, evidence *domain.Evidence, storage BlobStorage) (*domain.AnalysisResult, error)
	GetSupportedEvidenceTypes() []domain.EvidenceType
}

// HashCalculator defines the interface for hash calculation
type HashCalculator interface {
	Calculate(reader io.Reader) (string, error)
	Verify(data io.Reader, expectedHash string) (bool, error)
}

// IntegrityVerifier defines the interface for evidence integrity verification
type IntegrityVerifier interface {
	VerifyEvidence(ctx context.Context, evidence *domain.Evidence) (bool, error)
	GetIntegrityReport(ctx context.Context, evidence *domain.Evidence) (*IntegrityReport, error)
}

// IntegrityReport represents the result of an integrity check
type IntegrityReport struct {
	EvidenceID   string    `json:"evidence_id"`
	ExpectedHash string    `json:"expected_hash"`
	ActualHash   string    `json:"actual_hash"`
	Verified     bool      `json:"verified"`
	CheckedAt    time.Time `json:"checked_at"`
}
