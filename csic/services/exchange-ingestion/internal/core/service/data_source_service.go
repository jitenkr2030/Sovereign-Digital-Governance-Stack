package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/services/exchange-ingestion/internal/core/domain"
	"github.com/csic-platform/services/exchange-ingestion/internal/core/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DataSourceServiceImpl implements the DataSourceService interface
type DataSourceServiceImpl struct {
	repo       ports.DataSourceRepository
	factory    ports.ExchangeConnectorFactory
	logger     *zap.Logger
}

// NewDataSourceService creates a new DataSourceServiceImpl
func NewDataSourceService(
	repo ports.DataSourceRepository,
	factory ports.ExchangeConnectorFactory,
	logger *zap.Logger,
) *DataSourceServiceImpl {
	return &DataSourceServiceImpl{
		repo:    repo,
		factory: factory,
		logger:  logger,
	}
}

// RegisterDataSource creates a new data source configuration
func (s *DataSourceServiceImpl) RegisterDataSource(
	ctx context.Context,
	req *ports.RegisterDataSourceRequest,
) (*ports.RegisterDataSourceResponse, error) {
	s.logger.Info("Registering new data source",
		zap.String("name", req.Name),
		zap.String("exchange_type", string(req.ExchangeType)),
		zap.Strings("symbols", req.Symbols))

	// Check if a data source with the same name already exists
	existing, err := s.repo.FindByName(ctx, req.Name)
	if err == nil && existing != nil {
		return &ports.RegisterDataSourceResponse{
			Success: false,
			Message: "data source with this name already exists",
		}, domain.ErrDataSourceAlreadyExists
	}

	// Parse duration strings
	pollingRate := 5 * time.Second
	if req.PollingRate != "" {
		d, err := time.ParseDuration(req.PollingRate)
		if err != nil {
			return nil, fmt.Errorf("invalid polling rate: %w", err)
		}
		pollingRate = d
	}

	timeout := 30 * time.Second
	if req.Timeout != "" {
		d, err := time.ParseDuration(req.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout: %w", err)
		}
		timeout = d
	}

	retryAttempts := 3
	if req.RetryAttempts > 0 {
		retryAttempts = req.RetryAttempts
	}

	authType := domain.AuthTypeNone
	if req.AuthType != "" {
		authType = req.AuthType
	}

	config := &domain.DataSourceConfig{
		ID:            uuid.New(),
		Name:          req.Name,
		ExchangeType:  req.ExchangeType,
		Endpoint:      req.Endpoint,
		APIKey:        req.APIKey,
		APISecret:     req.APISecret,
		Symbols:       req.Symbols,
		Enabled:       true,
		PollingRate:   pollingRate,
		WSEnabled:     req.WSEnabled,
		AuthType:      authType,
		Timeout:       timeout,
		RetryAttempts: retryAttempts,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Validate that we can create a connector for this exchange type
	_, err = s.factory.CreateConnector(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	// Save the configuration
	if err := s.repo.Save(ctx, config); err != nil {
		s.logger.Error("Failed to save data source", zap.Error(err))
		return nil, fmt.Errorf("failed to save data source: %w", err)
	}

	s.logger.Info("Successfully registered data source",
		zap.String("id", config.ID.String()),
		zap.String("name", config.Name))

	return &ports.RegisterDataSourceResponse{
		ID:      config.ID.String(),
		Name:    config.Name,
		Success: true,
		Message: "data source registered successfully",
	}, nil
}

// UpdateDataSource updates an existing data source configuration
func (s *DataSourceServiceImpl) UpdateDataSource(
	ctx context.Context,
	req *ports.UpdateDataSourceRequest,
) error {
	s.logger.Info("Updating data source", zap.String("id", req.ID))

	config, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("data source not found: %w", err)
	}

	if req.Name != nil {
		config.Name = *req.Name
	}
	if req.Endpoint != nil {
		config.Endpoint = *req.Endpoint
	}
	if req.APIKey != nil {
		config.APIKey = *req.APIKey
	}
	if req.APISecret != nil {
		config.APISecret = *req.APISecret
	}
	if req.Symbols != nil {
		config.Symbols = req.Symbols
	}
	if req.PollingRate != nil {
		d, err := time.ParseDuration(*req.PollingRate)
		if err != nil {
			return fmt.Errorf("invalid polling rate: %w", err)
		}
		config.PollingRate = d
	}
	if req.WSEnabled != nil {
		config.WSEnabled = *req.WSEnabled
	}
	if req.AuthType != nil {
		config.AuthType = domain.AuthType(*req.AuthType)
	}
	if req.Timeout != nil {
		d, err := time.ParseDuration(*req.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}
		config.Timeout = d
	}
	if req.RetryAttempts != nil {
		config.RetryAttempts = *req.RetryAttempts
	}

	config.UpdatedAt = time.Now()

	if err := s.repo.Save(ctx, config); err != nil {
		s.logger.Error("Failed to update data source", zap.Error(err))
		return fmt.Errorf("failed to update data source: %w", err)
	}

	s.logger.Info("Successfully updated data source", zap.String("id", req.ID))
	return nil
}

// EnableDataSource enables a disabled data source
func (s *DataSourceServiceImpl) EnableDataSource(ctx context.Context, id string) error {
	s.logger.Info("Enabling data source", zap.String("id", id))

	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("data source not found: %w", err)
	}

	config.Enabled = true
	config.UpdatedAt = time.Now()

	if err := s.repo.Save(ctx, config); err != nil {
		return fmt.Errorf("failed to enable data source: %w", err)
	}

	return nil
}

// DisableDataSource disables an active data source
func (s *DataSourceServiceImpl) DisableDataSource(ctx context.Context, id string) error {
	s.logger.Info("Disabling data source", zap.String("id", id))

	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("data source not found: %w", err)
	}

	config.Enabled = false
	config.UpdatedAt = time.Now()

	if err := s.repo.Save(ctx, config); err != nil {
		return fmt.Errorf("failed to disable data source: %w", err)
	}

	return nil
}

// ListDataSources lists all data sources with optional filtering
func (s *DataSourceServiceImpl) ListDataSources(
	ctx context.Context,
	req *ports.ListDataSourcesRequest,
) (*ports.ListDataSourcesResponse, error) {
	s.logger.Debug("Listing data sources",
		zap.Bool("enabled_only", req.EnabledOnly),
		zap.String("exchange_type", req.ExchangeType))

	var dataSources []*domain.DataSourceConfig
	var err error

	if req.EnabledOnly {
		dataSources, err = s.repo.FindEnabled(ctx)
	} else {
		dataSources, err = s.repo.FindAll(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list data sources: %w", err)
	}

	// Filter by exchange type if specified
	if req.ExchangeType != "" {
		filtered := make([]*domain.DataSourceConfig, 0, len(dataSources))
		for _, ds := range dataSources {
			if ds.ExchangeType == domain.ExchangeType(req.ExchangeType) {
				filtered = append(filtered, ds)
			}
		}
		dataSources = filtered
	}

	// Apply pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	total := len(dataSources)
	start := (page - 1) * pageSize
	if start >= total {
		dataSources = []*domain.DataSourceConfig{}
	} else {
		end := start + pageSize
		if end > total {
			end = total
		}
		dataSources = dataSources[start:end]
	}

	return &ports.ListDataSourcesResponse{
		DataSources: dataSources,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// GetDataSource retrieves a specific data source by ID
func (s *DataSourceServiceImpl) GetDataSource(ctx context.Context, id string) (*domain.DataSourceConfig, error) {
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("data source not found: %w", err)
	}
	return config, nil
}

// TestConnection tests the connection to a data source
func (s *DataSourceServiceImpl) TestConnection(
	ctx context.Context,
	id string,
) (*ports.TestConnectionResponse, error) {
	s.logger.Info("Testing connection for data source", zap.String("id", id))

	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("data source not found: %w", err)
	}

	connector, err := s.factory.CreateConnector(config)
	if err != nil {
		return &ports.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("failed to create connector: %v", err),
		}, nil
	}

	start := time.Now()
	if err := connector.Connect(ctx); err != nil {
		return &ports.TestConnectionResponse{
			Success:   false,
			Latency:   time.Since(start).String(),
			Message:   fmt.Sprintf("connection failed: %v", err),
			Timestamp: start.Format(time.RFC3339),
		}, nil
	}
	defer connector.Disconnect(ctx)

	return &ports.TestConnectionResponse{
		Success:   true,
		Latency:   time.Since(start).String(),
		Message:   "connection successful",
		Timestamp: start.Format(time.RFC3339),
	}, nil
}
