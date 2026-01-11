package ports

import (
	"context"

	"github.com/csic-platform/services/exchange-ingestion/internal/core/domain"
)

// DataSourceService defines the business logic for managing data sources
type DataSourceService interface {
	// RegisterDataSource creates a new data source configuration
	RegisterDataSource(ctx context.Context, req *RegisterDataSourceRequest) (*RegisterDataSourceResponse, error)

	// UpdateDataSource updates an existing data source configuration
	UpdateDataSource(ctx context.Context, req *UpdateDataSourceRequest) error

	// EnableDataSource enables a disabled data source
	EnableDataSource(ctx context.Context, id string) error

	// DisableDataSource disables an active data source
	DisableDataSource(ctx context.Context, id string) error

	// ListDataSources lists all data sources with optional filtering
	ListDataSources(ctx context.Context, req *ListDataSourcesRequest) (*ListDataSourcesResponse, error)

	// GetDataSource retrieves a specific data source by ID
	GetDataSource(ctx context.Context, id string) (*domain.DataSourceConfig, error)

	// TestConnection tests the connection to a data source
	TestConnection(ctx context.Context, id string) (*TestConnectionResponse, error)
}

// IngestionService defines the business logic for data ingestion operations
type IngestionService interface {
	// StartIngestion starts the ingestion process for all enabled data sources
	StartIngestion(ctx context.Context) error

	// StopIngestion stops all active ingestion processes
	StopIngestion(ctx context.Context) error

	// StartSource starts ingestion for a specific data source
	StartSource(ctx context.Context, sourceID string) error

	// StopSource stops ingestion for a specific data source
	StopSource(ctx context.Context, sourceID string) error

	// GetIngestionStatus returns the current status of all ingestion processes
	GetIngestionStatus(ctx context.Context) (*IngestionStatusResponse, error)

	// GetSourceStatus returns the status of a specific data source
	GetSourceStatus(ctx context.Context, sourceID string) (*SourceStatusResponse, error)

	// ForceSync triggers an immediate data sync for a data source
	ForceSync(ctx context.Context, sourceID string) error

	// GetStats returns ingestion statistics
	GetStats(ctx context.Context, sourceID string) (*domain.IngestionStats, error)
}

// Request and response types for DataSourceService

type RegisterDataSourceRequest struct {
	Name         string                `json:"name" validate:"required,min=1,max=100"`
	ExchangeType domain.ExchangeType   `json:"exchange_type" validate:"required"`
	Endpoint     string                `json:"endpoint" validate:"required,uri"`
	APIKey       string                `json:"api_key,omitempty"`
	APISecret    string                `json:"api_secret,omitempty"`
	Symbols      []string              `json:"symbols" validate:"required,min=1"`
	PollingRate  string                `json:"polling_rate,omitempty"` // Duration string like "1s", "5m"
	WSEnabled    bool                  `json:"ws_enabled,omitempty"`
	AuthType     domain.AuthType       `json:"auth_type,omitempty"`
	Timeout      string                `json:"timeout,omitempty"`
	RetryAttempts int                  `json:"retry_attempts,omitempty"`
}

type RegisterDataSourceResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type UpdateDataSourceRequest struct {
	ID            string   `json:"id" validate:"required"`
	Name          *string  `json:"name,omitempty"`
	Endpoint      *string  `json:"endpoint,omitempty"`
	APIKey        *string  `json:"api_key,omitempty"`
	APISecret     *string  `json:"api_secret,omitempty"`
	Symbols       []string `json:"symbols,omitempty"`
	PollingRate   *string  `json:"polling_rate,omitempty"`
	WSEnabled     *bool    `json:"ws_enabled,omitempty"`
	AuthType      *string  `json:"auth_type,omitempty"`
	Timeout       *string  `json:"timeout,omitempty"`
	RetryAttempts *int     `json:"retry_attempts,omitempty"`
}

type ListDataSourcesRequest struct {
	EnabledOnly bool   `json:"enabled_only,omitempty"`
	ExchangeType string `json:"exchange_type,omitempty"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
}

type ListDataSourcesResponse struct {
	DataSources []*domain.DataSourceConfig `json:"data_sources"`
	Total       int                        `json:"total"`
	Page        int                        `json:"page"`
	PageSize    int                        `json:"page_size"`
}

type TestConnectionResponse struct {
	Success   bool   `json:"success"`
	Latency   string `json:"latency,omitempty"`
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp"`
}

// Request and response types for IngestionService

type IngestionStatusResponse struct {
	OverallStatus  string                     `json:"overall_status"`
	ActiveSources  int                        `json:"active_sources"`
	TotalSources   int                        `json:"total_sources"`
	SourceStatuses []*SourceStatusResponse    `json:"source_statuses"`
}

type SourceStatusResponse struct {
	SourceID         string                    `json:"source_id"`
	Name             string                    `json:"name"`
	Status           domain.ConnectionStatus   `json:"status"`
	MessagesReceived int64                     `json:"messages_received"`
	MessagesFailed   int64                     `json:"messages_failed"`
	LastMessageTime  string                    `json:"last_message_time,omitempty"`
	Error            string                    `json:"error,omitempty"`
}
