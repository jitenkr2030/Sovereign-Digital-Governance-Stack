package domain

import (
	"time"
)

// EvidenceStatus represents the status of evidence
type EvidenceStatus string

const (
	EvidenceStatusUploading   EvidenceStatus = "UPLOADING"
	EvidenceStatusUploaded    EvidenceStatus = "UPLOADED"
	EvidenceStatusProcessing  EvidenceStatus = "PROCESSING"
	EvidenceStatusAnalyzed    EvidenceStatus = "ANALYZED"
	EvidenceStatusArchived    EvidenceStatus = "ARCHIVED"
	EvidenceStatusDeleted     EvidenceStatus = "DELETED"
)

// EvidenceType represents the type of evidence
type EvidenceType string

const (
	EvidenceTypeFile       EvidenceType = "FILE"
	EvidenceTypeDiskImage  EvidenceType = "DISK_IMAGE"
	EvidenceTypeMemoryDump EvidenceType = "MEMORY_DUMP"
	EvidenceTypeNetworkCapture EvidenceType = "NETWORK_CAPTURE"
	EvidenceTypeLogFile    EvidenceType = "LOG_FILE"
	EvidenceTypeDocument   EvidenceType = "DOCUMENT"
)

// Evidence represents digital evidence
type Evidence struct {
	ID           string        `json:"id"`
	CaseID       string        `json:"case_id"`
	FileName     string        `json:"file_name"`
	FileHash     string        `json:"file_hash"` // SHA-256
	FileType     string        `json:"file_type"`
	SizeBytes    int64         `json:"size_bytes"`
	StoragePath  string        `json:"storage_path"`
	EvidenceType EvidenceType `json:"evidence_type"`
	Status       EvidenceStatus `json:"status"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	UploadedBy   string        `json:"uploaded_by"`
	UploadedAt   time.Time     `json:"uploaded_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	ArchivedAt   *time.Time    `json:"archived_at,omitempty"`
	DeletedAt    *time.Time    `json:"deleted_at,omitempty"`
}

// ChainOfCustody represents an entry in the chain of custody log
type ChainOfCustody struct {
	ID           string                 `json:"id"`
	EvidenceID   string                 `json:"evidence_id"`
	ActorID      string                 `json:"actor_id"`
	Action       ChainOfCustodyAction   `json:"action"`
	Details      map[string]interface{} `json:"details,omitempty"`
	IPAddress    string                 `json:"ip_address"`
	Timestamp    time.Time              `json:"timestamp"`
	Signature    string                 `json:"signature,omitempty"` // Digital signature for integrity
}

// ChainOfCustodyAction represents actions in the chain of custody
type ChainOfCustodyAction string

const (
	CoCActionUpload       ChainOfCustodyAction = "UPLOAD"
	CoCActionDownload     ChainOfCustodyAction = "DOWNLOAD"
	CoCActionAccess       ChainOfCustodyAction = "ACCESS"
	CoCActionAnalysis     ChainOfCustodyAction = "ANALYSIS"
	CoCActionExport       ChainOfCustodyAction = "EXPORT"
	CoCActionArchive      ChainOfCustodyAction = "ARCHIVE"
	CoCActionDelete       ChainOfCustodyAction = "DELETE"
)

// AnalysisJobStatus represents the status of an analysis job
type AnalysisJobStatus string

const (
	AnalysisJobStatusPending    AnalysisJobStatus = "PENDING"
	AnalysisJobStatusQueued     AnalysisJobStatus = "QUEUED"
	AnalysisJobStatusProcessing AnalysisJobStatus = "PROCESSING"
	AnalysisJobStatusCompleted  AnalysisJobStatus = "COMPLETED"
	AnalysisJobStatusFailed     AnalysisJobStatus = "FAILED"
	AnalysisJobStatusCancelled  AnalysisJobStatus = "CANCELLED"
)

// AnalysisJob represents a forensic analysis job
type AnalysisJob struct {
	ID           string            `json:"id"`
	EvidenceID   string            `json:"evidence_id"`
	ToolName     string            `json:"tool_name"`
	Status       AnalysisJobStatus `json:"status"`
	Priority     int               `json:"priority"` // 1-5, 5 being highest
	RequestedBy  string            `json:"requested_by"`
	SubmittedAt  time.Time         `json:"submitted_at"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
	ErrorMessage string            `json:"error_message,omitempty"`
	Progress     int               `json:"progress"` // 0-100
}

// AnalysisResult represents the result of an analysis
type AnalysisResult struct {
	ID           string                 `json:"id"`
	JobID        string                 `json:"job_id"`
	EvidenceID   string                 `json:"evidence_id"`
	ToolName     string                 `json:"tool_name"`
	ResultType   string                 `json:"result_type"`
	ResultData   map[string]interface{} `json:"result_data"`
	Summary      string                 `json:"summary"`
	Severity     string                 `json:"severity,omitempty"` // info, warning, critical
	Tags         []string               `json:"tags,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// EvidenceUploadRequest represents a request to upload evidence
type EvidenceUploadRequest struct {
	CaseID       string            `form:"case_id" binding:"required"`
	FileName     string            `form:"file_name"`
	EvidenceType EvidenceType      `form:"evidence_type"`
	Metadata     map[string]string `form:"metadata"`
}

// EvidenceListResponse represents a paginated list of evidence
type EvidenceListResponse struct {
	Evidence    []*Evidence `json:"evidence"`
	Total       int64       `json:"total"`
	Page        int         `json:"page"`
	PageSize    int         `json:"page_size"`
	TotalPages  int         `json:"total_pages"`
}

// AnalysisRequest represents a request to analyze evidence
type AnalysisRequest struct {
	EvidenceID  string   `json:"evidence_id" binding:"required"`
	Tools       []string `json:"tools"` // List of tools to run
	Priority    int      `json:"priority"`
}

// AnalysisJobResponse represents the response after submitting an analysis job
type AnalysisJobResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

// NewEvidenceListResponse creates a new paginated evidence list response
func NewEvidenceListResponse(evidence []*Evidence, total int64, page, pageSize int) *EvidenceListResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return &EvidenceListResponse{
		Evidence:   evidence,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
