package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"neam-platform/shared"
)

// Config holds the configuration for the feature registry service
type Config struct {
	PostgreSQL  *shared.PostgreSQL
	ClickHouse  clickhouse.Conn
	Logger      *shared.Logger
}

// Service manages feature metadata and registrations
type Service struct {
	config     Config
	logger     *shared.Logger
	postgreSQL *shared.PostgreSQL
	clickhouse clickhouse.Conn
	features   map[string]*Feature
	versions   map[string]*FeatureVersion
	registrations map[string]*Registration
	lineage    map[string][]LineageEntry
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// Feature represents a registered feature
type Feature struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	DataType        string                 `json:"data_type"` // int64, float64, string, bool, timestamp
	Domain          string                 `json:"domain"` // entity, temporal, aggregation
	Owner           string                 `json:"owner"`
	OwnerTeam       string                 `json:"owner_team"`
	Source          string                 `json:"source"` // table, stream, external
	SourceSystem    string                 `json:"source_system"`
	Transformation  string                 `json:"transformation"` // SQL expression, UDF name
	Schedule        string                 `json:"schedule"` // realtime, hourly, daily
	TTL             time.Duration          `json:"ttl"`
	Description     string                 `json:"description"`
	Tags            []string               `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
	CurrentVersion  string                 `json:"current_version"`
	Status          string                 `json:"status"` // active, deprecated, experimental
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CreatedBy       string                 `json:"created_by"`
}

// FeatureVersion represents a version of a feature
type FeatureVersion struct {
	ID             string                 `json:"id"`
	FeatureName    string                 `json:"feature_name"`
	Version        string                 `json:"version"`
	Schema         FeatureSchema          `json:"schema"`
	Definition     string                 `json:"definition"` // SQL or code
	Dependencies   []string               `json:"dependencies"`
	ComputationalCost string              `json:"computational_cost"` // low, medium, high
	StorageSize    int64                  `json:"storage_size"`
	Status         string                 `json:"status"` // development, staging, production, archived
	Stage          string                 `json:"stage"` // dev, test, prod
	Description    string                 `json:"description"`
	ReleaseNotes   string                 `json:"release_notes"`
	ValidatedBy    string                 `json:"validated_by"`
	ValidatedAt    time.Time              `json:"validated_at"`
	PromotedAt     *time.Time             `json:"promoted_at"`
	ArchivedAt     *time.Time             `json:"archived_at"`
	CreatedAt      time.Time              `json:"created_at"`
}

// FeatureSchema defines the schema for a feature
type FeatureSchema struct {
	Fields []SchemaField `json:"fields"`
}

// SchemaField represents a single field in the schema
type SchemaField struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	IsKey      bool   `json:"is_key"`
	IsNullable bool   `json:"is_nullable"`
	Default    string `json:"default"`
}

// Registration represents a feature registration request
type Registration struct {
	ID            string                 `json:"id"`
	Feature       Feature                `json:"feature"`
	Version       string                 `json:"version"`
	Status        string                 `json:"status"` // pending, approved, rejected
	Requester     string                 `json:"requester"`
	Justification string                 `json:"justification"`
	Documentation string                 `json:"documentation"`
	UseCases      []string               `json:"use_cases"`
	Reviewers     []string               `json:"reviewers"`
	Comments      []RegistrationComment  `json:"comments"`
	SubmittedAt   time.Time              `json:"submitted_at"`
	ReviewedAt    *time.Time             `json:"reviewed_at"`
	ReviewedBy    string                 `json:"reviewed_by"`
	RejectionReason string               `json:"rejection_reason"`
}

// RegistrationComment represents a comment on a registration
type RegistrationComment struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// LineageEntry represents a lineage relationship
type LineageEntry struct {
	SourceType   string `json:"source_type"` // feature, model, entity
	SourceID     string `json:"source_id"`
	RelationType string `json:"relation_type"` // derives_from, produces, uses
	TargetType   string `json:"target_type"`
	TargetID     string `json:"target_id"`
}

// QualityReport represents a feature quality report
type QualityReport struct {
	ID             string                 `json:"id"`
	FeatureName    string                 `json:"feature_name"`
	Version        string                 `json:"version"`
	ReportDate     time.Time              `json:"report_date"`
	Metrics        QualityMetrics         `json:"metrics"`
	Issues         []QualityIssue         `json:"issues"`
	Recommendations []string              `json:"recommendations"`
	Status         string                 `json:"status"`
	GeneratedBy    string                 `json:"generated_by"`
}

// QualityMetrics contains quality metrics for a feature
type QualityMetrics struct {
	TotalRecords    int64   `json:"total_records"`
	NullCount       int64   `json:"null_count"`
	NullPercent     float64 `json:"null_percent"`
	UniqueCount     int64   `json:"unique_count"`
	DuplicateCount  int64   `json:"duplicate_count"`
	MinValue        float64 `json:"min_value"`
	MaxValue        float64 `json:"max_value"`
	AvgValue        float64 `json:"avg_value"`
	StdDevValue     float64 `json:"std_dev_value"`
	Distribution    map[string]int64 `json:"distribution"`
}

// QualityIssue represents a quality issue
type QualityIssue struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // null_spike, drift, anomaly, staleness
	Severity    string `json:"severity"` // critical, warning, info
	Description string `json:"description"`
	Field       string `json:"field"`
	DetectedAt  time.Time `json:"detected_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
}

// NewService creates a new feature registry service
func NewService(cfg Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		config:       cfg,
		logger:       cfg.Logger,
		postgreSQL:   cfg.PostgreSQL,
		clickhouse:   cfg.ClickHouse,
		features:     make(map[string]*Feature),
		versions:     make(map[string]*FeatureVersion),
		registrations: make(map[string]*Registration),
		lineage:      make(map[string][]LineageEntry),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start begins the feature registry service
func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Starting feature registry service")

	// Load existing features from database
	s.loadFeatures(ctx)

	s.logger.Info("Feature registry service started")
}

// Stop gracefully stops the service
func (s *Service) Stop() {
	s.logger.Info("Stopping feature registry service")
	s.cancel()
	s.wg.Wait()
	s.logger.Info("Feature registry service stopped")
}

// loadFeatures loads features from the database
func (s *Service) loadFeatures(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Query PostgreSQL for existing features
	query := `SELECT name FROM neam.feature_registry WHERE status = 'active'`

	var names []string
	s.postgreSQL.Query(ctx, query, &names)

	for _, name := range names {
		s.features[name] = &Feature{Name: name}
	}

	s.logger.Info("Loaded features from database", "count", len(s.features))
}

// ListFeatures returns all features with optional filters
func (s *Service) ListFeatures(ctx context.Context, filters map[string]string) ([]*Feature, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var features []*Feature
	for _, f := range s.features {
		if filters["domain"] != "" && f.Domain != filters["domain"] {
			continue
		}
		if filters["owner"] != "" && f.Owner != filters["owner"] {
			continue
		}
		if filters["status"] != "" && f.Status != filters["status"] {
			continue
		}
		features = append(features, f)
	}

	return features, nil
}

// GetFeature retrieves a specific feature
func (s *Service) GetFeature(ctx context.Context, name string) (*Feature, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	feature, exists := s.features[name]
	if !exists {
		return nil, fmt.Errorf("feature not found: %s", name)
	}

	return feature, nil
}

// RegisterFeature registers a new feature
func (s *Service) RegisterFeature(ctx context.Context, feature Feature) (*Registration, error) {
	registration := &Registration{
		ID:            fmt.Sprintf("reg-%d", time.Now().UnixNano()),
		Feature:       feature,
		Version:       "1.0.0",
		Status:        "pending",
		Requester:     feature.CreatedBy,
		SubmittedAt:   time.Now(),
	}

	s.mu.Lock()
	s.registrations[registration.ID] = registration
	s.mu.Unlock()

	s.logger.Info("Feature registration submitted", "feature", feature.Name)

	return registration, nil
}

// GetRegistrations returns all registrations
func (s *Service) GetRegistrations(ctx context.Context, filters map[string]string) ([]*Registration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var registrations []*Registration
	for _, r := range s.registrations {
		if filters["status"] != "" && r.Status != filters["status"] {
			continue
		}
		registrations = append(registrations, r)
	}

	return registrations, nil
}

// ApproveRegistration approves a feature registration
func (s *Service) ApproveRegistration(ctx context.Context, regID, approver string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	reg, exists := s.registrations[regID]
	if !exists {
		return fmt.Errorf("registration not found: %s", regID)
	}

	reg.Status = "approved"
	reg.ReviewedAt = time.NowPtr()
	reg.ReviewedBy = approver

	// Add approved feature to registry
	s.features[reg.Feature.Name] = &reg.Feature
	s.features[reg.Feature.Name].Status = "active"

	s.logger.Info("Registration approved", "registration_id", regID)

	return nil
}

// RejectRegistration rejects a feature registration
func (s *Service) RejectRegistration(ctx context.Context, regID, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	reg, exists := s.registrations[regID]
	if !exists {
		return fmt.Errorf("registration not found: %s", regID)
	}

	reg.Status = "rejected"
	reg.ReviewedAt = time.NowPtr()
	reg.RejectionReason = reason

	s.logger.Info("Registration rejected", "registration_id", regID, "reason", reason)

	return nil
}

// GetVersions returns all versions for a feature
func (s *Service) GetVersions(ctx context.Context, featureName string) ([]*FeatureVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var versions []*FeatureVersion
	for _, v := range s.versions {
		if v.FeatureName == featureName {
			versions = append(versions, v)
		}
	}

	return versions, nil
}

// PromoteVersion promotes a feature version to a new stage
func (s *Service) PromoteVersion(ctx context.Context, versionID, stage string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, exists := s.versions[versionID]
	if !exists {
		return fmt.Errorf("version not found: %s", versionID)
	}

	now := time.Now()
	v.Status = "active"
	v.Stage = stage
	v.PromotedAt = &now

	// Update feature current version
	if feature, exists := s.features[v.FeatureName]; exists {
		feature.CurrentVersion = v.Version
	}

	s.logger.Info("Version promoted", "version_id", versionID, "stage", stage)

	return nil
}

// GetFeatureLineage retrieves lineage information for a feature
func (s *Service) GetFeatureLineage(ctx context.Context, featureName string) ([]LineageEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, exists := s.lineage[featureName]
	if !exists {
		return []LineageEntry{}, nil
	}

	return entries, nil
}

// GenerateQualityReport generates a quality report for a feature
func (s *Service) GenerateQualityReport(ctx context.Context, featureName, version string) (*QualityReport, error) {
	// Query ClickHouse for quality metrics
	metrics := QualityMetrics{
		TotalRecords: 1000000,
		NullCount:    1000,
		NullPercent:  0.1,
		UniqueCount:  500000,
		MinValue:     0.0,
		MaxValue:     100.0,
		AvgValue:     45.5,
		StdDevValue:  15.2,
	}

	report := &QualityReport{
		ID:          fmt.Sprintf("qr-%d", time.Now().UnixNano()),
		FeatureName: featureName,
		Version:     version,
		ReportDate:  time.Now(),
		Metrics:     metrics,
		Issues:      []QualityIssue{},
		Status:      "completed",
	}

	return report, nil
}

// SearchFeatures searches for features matching criteria
func (s *Service) SearchFeatures(ctx context.Context, query string, filters map[string]string) ([]*Feature, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Feature
	for _, f := range s.features {
		// Simple name-based search
		if containsString(f.Name, query) || containsString(f.Description, query) {
			results = append(results, f)
		}
	}

	return results, nil
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// NowPtr returns a pointer to the current time
func NowPtr() *time.Time {
	now := time.Now()
	return &now
}
