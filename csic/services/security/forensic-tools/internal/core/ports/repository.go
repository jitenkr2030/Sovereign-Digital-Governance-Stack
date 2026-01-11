package ports

import (
	"context"
	"io"

	"github.com/csic-platform/services/security/forensic-tools/internal/core/domain"
)

// EvidenceRepository defines the interface for evidence data access operations
type EvidenceRepository interface {
	// Evidence operations
	CreateEvidence(ctx context.Context, evidence *domain.Evidence) error
	GetEvidence(ctx context.Context, id string) (*domain.Evidence, error)
	GetEvidenceByHash(ctx context.Context, hash string) (*domain.Evidence, error)
	ListEvidence(ctx context.Context, caseID string, page, pageSize int) ([]*domain.Evidence, int64, error)
	UpdateEvidence(ctx context.Context, evidence *domain.Evidence) error
	DeleteEvidence(ctx context.Context, id string) error

	// Chain of custody operations
	AddCustodyEntry(ctx context.Context, entry *domain.ChainOfCustody) error
	GetCustodyHistory(ctx context.Context, evidenceID string) ([]*domain.ChainOfCustody, error)

	// Analysis job operations
	CreateAnalysisJob(ctx context.Context, job *domain.AnalysisJob) error
	GetAnalysisJob(ctx context.Context, id string) (*domain.AnalysisJob, error)
	UpdateAnalysisJob(ctx context.Context, job *domain.AnalysisJob) error
	ListAnalysisJobs(ctx context.Context, evidenceID string, page, pageSize int) ([]*domain.AnalysisJob, int64, error)
	GetPendingJobs(ctx context.Context, limit int) ([]*domain.AnalysisJob, error)

	// Analysis result operations
	CreateAnalysisResult(ctx context.Context, result *domain.AnalysisResult) error
	GetAnalysisResults(ctx context.Context, jobID string) ([]*domain.AnalysisResult, error)
}

// BlobStorage defines the interface for evidence file storage
type BlobStorage interface {
	// Upload operations
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error
	UploadFromFile(ctx context.Context, sourcePath, objectName string) error

	// Download operations
	Download(ctx context.Context, objectName string, writer io.Writer) error
	GetFile(ctx context.Context, objectName string) (io.ReadCloser, error)

	// Delete operations
	Delete(ctx context.Context, objectName string) error
	DeleteMultiple(ctx context.Context, objectNames []string) error

	// Utility operations
	GetObjectInfo(ctx context.Context, objectName string) (*BlobObjectInfo, error)
	ObjectExists(ctx context.Context, objectName string) (bool, error)
	GetSignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
}

// BlobObjectInfo contains information about a stored object
type BlobObjectInfo struct {
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
	ETag        string    `json:"etag"`
}

// MessageProducer defines the interface for message publishing operations
type MessageProducer interface {
	PublishAnalysisJob(ctx context.Context, job *domain.AnalysisJob) error
	PublishAnalysisResult(ctx context.Context, result *domain.AnalysisResult) error
	PublishCustodyEvent(ctx context.Context, entry *domain.ChainOfCustody) error
}
