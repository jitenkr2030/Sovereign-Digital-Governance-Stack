package services

import (
	"context"
	"fmt"

	"github.com/csic/monitoring/internal/core/domain"
	"github.com/csic/monitoring/internal/core/ports"
	"go.uber.org/zap"
)

// SanctionsService handles sanctions list management
type SanctionsService struct {
	repo   ports.SanctionsRepository
	logger *zap.Logger
}

// NewSanctionsService creates a new sanctions service
func NewSanctionsService(repo ports.SanctionsRepository, logger *zap.Logger) *SanctionsService {
	return &SanctionsService{
		repo:   repo,
		logger: logger,
	}
}

// AddSanction adds a new sanctioned address
func (s *SanctionsService) AddSanction(ctx context.Context, sanction *domain.SanctionedAddress) error {
	// Check if already exists
	exists, err := s.repo.Exists(ctx, sanction.Address, sanction.Chain)
	if err != nil {
		return fmt.Errorf("failed to check existing sanction: %w", err)
	}
	if exists {
		return fmt.Errorf("address already in sanctions list")
	}

	// Add the sanction
	if err := s.repo.Create(ctx, sanction); err != nil {
		return fmt.Errorf("failed to add sanction: %w", err)
	}

	s.logger.Info("Sanction added",
		zap.String("address", sanction.Address),
		zap.String("chain", sanction.Chain),
		zap.String("source_list", sanction.SourceList),
		zap.String("entity_name", sanction.EntityName),
	)

	return nil
}

// ListSanctions retrieves all sanctioned addresses
func (s *SanctionsService) ListSanctions(ctx context.Context, page, pageSize int) ([]*domain.SanctionedAddress, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

// CheckAddress checks if an address is sanctioned
func (s *SanctionsService) CheckAddress(ctx context.Context, address, chain string) ([]*domain.SanctionedAddress, error) {
	return s.repo.GetByAddress(ctx, address, chain)
}

// ImportSanctionsList imports a batch of sanctioned addresses
func (s *SanctionsService) ImportSanctionsList(ctx context.Context, importReq *domain.SanctionsListImport) (imported, failed int, err error) {
	sanctions := make([]*domain.SanctionedAddress, 0, len(importReq.Addresses))

	for _, addr := range importReq.Addresses {
		sanction := &domain.SanctionedAddress{
			Address:    addr.Address,
			Chain:      addr.Chain,
			SourceList: importReq.SourceList,
			Reason:     addr.Reason,
			AddedAt:    addr.AddedAt,
		}
		sanctions = append(sanctions, sanction)
	}

	if err := s.repo.CreateBatch(ctx, sanctions); err != nil {
		return 0, len(sanctions), fmt.Errorf("failed to import sanctions: %w", err)
	}

	s.logger.Info("Sanctions list imported",
		zap.String("source_list", importReq.SourceList),
		zap.Int("imported_count", len(sanctions)),
		zap.String("imported_by", importReq.ImportedBy),
	)

	return len(sanctions), 0, nil
}

// RemoveSanction removes a sanctioned address
func (s *SanctionsService) RemoveSanction(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to remove sanction: %w", err)
	}

	s.logger.Info("Sanction removed", zap.String("id", id))

	return nil
}
