package connector

import (
	"fmt"

	"github.com/csic-platform/services/exchange-ingestion/internal/core/domain"
	"github.com/csic-platform/services/exchange-ingestion/internal/core/ports"
	"go.uber.org/zap"
)

// ConnectorFactory creates ExchangeConnector instances based on exchange type
type ConnectorFactory struct {
	logger *zap.Logger
}

// NewConnectorFactory creates a new ConnectorFactory
func NewConnectorFactory(logger *zap.Logger) *ConnectorFactory {
	return &ConnectorFactory{logger: logger}
}

// CreateConnector creates a new connector for the specified exchange type
func (f *ConnectorFactory) CreateConnector(config *domain.DataSourceConfig) (ports.ExchangeConnector, error) {
	if err := config.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var connector ports.ExchangeConnector

	switch config.ExchangeType {
	case domain.ExchangeTypeMock:
		connector = NewMockConnector(config, f.logger)
	case domain.ExchangeTypeBinance, domain.ExchangeTypeCoinbase,
		domain.ExchangeTypeKraken, domain.ExchangeTypeCME,
		domain.ExchangeTypeGeneric:
		connector = NewHTTPConnector(config, f.logger)
	default:
		return nil, fmt.Errorf("unsupported exchange type: %s", config.ExchangeType)
	}

	// Validate the connector configuration
	if err := connector.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("connector validation failed: %w", err)
	}

	return connector, nil
}

// GetSupportedExchangeTypes returns the list of supported exchange types
func (f *ConnectorFactory) GetSupportedExchangeTypes() []domain.ExchangeType {
	return []domain.ExchangeType{
		domain.ExchangeTypeMock,
		domain.ExchangeTypeBinance,
		domain.ExchangeTypeCoinbase,
		domain.ExchangeTypeKraken,
		domain.ExchangeTypeCME,
		domain.ExchangeTypeGeneric,
	}
}

// Ensure ConnectorFactory implements ExchangeConnectorFactory
var _ ports.ExchangeConnectorFactory = (*ConnectorFactory)(nil)
