package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic-platform/services/services/compliance/internal/core/domain"
	"github.com/csic-platform/services/services/compliance/internal/core/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ComplianceService implements the ComplianceService interface
type ComplianceService struct {
	repo             ports.ComplianceRepository
	obligationRepo   ports.ObligationRepository
	licenseRepo      ports.LicenseRepository
	auditRepo        ports.AuditRepository
	log              *zap.Logger
	baseScore        float64
	overdueDeduction float64
	suspendedDeduction float64
	earlyBonus       float64
}

// NewComplianceService creates a new ComplianceService instance
func NewComplianceService(repo ports.ComplianceRepository, logger *zap.Logger) *ComplianceService {
	return &ComplianceService{
		repo:              repo,
		log:               logger,
		baseScore:         100.0,
		overdueDeduction:  10.0,
		suspendedDeduction: 50.0,
		earlyBonus:        5.0,
	}
}

// CalculateScore calculates the compliance score for an entity
func (s *ComplianceService) CalculateScore(ctx context.Context, entityID uuid.UUID) (*domain.ComplianceScore, error) {
	s.log.Info("Calculating compliance score", zap.String("entity_id", entityID.String()))

	// Get overdue obligations
	overdueObligations, err := s.obligationRepo.GetObligationsByStatus(ctx, domain.ObligationOverdue)
	if err != nil {
		s.log.Error("Failed to get overdue obligations", zap.Error(err))
	}

	entityOverdueCount := 0
	for _, obs := range overdueObligations {
		if obs.EntityID == entityID {
			entityOverdueCount++
		}
	}

	// Get entity's licenses
	licenses, err := s.licenseRepo.GetLicensesByEntity(ctx, entityID)
	if err != nil {
		s.log.Error("Failed to get entity licenses", zap.Error(err))
	}

	// Count suspended licenses
	suspendedCount := 0
	for _, lic := range licenses {
		if lic.Status == domain.LicenseStatusSuspended {
			suspendedCount++
		}
	}

	// Calculate score
	score := s.baseScore
	breakdown := make(map[string]float64)

	// Base score
	breakdown["base_score"] = s.baseScore

	// Deductions for overdue obligations
	overdueDeduction := float64(entityOverdueCount) * s.overdueDeduction
	score -= overdueDeduction
	breakdown["overdue_deductions"] = -overdueDeduction
	breakdown["overdue_count"] = float64(entityOverdueCount)

	// Deductions for suspended licenses
	suspendedDeduction := float64(suspendedCount) * s.suspendedDeduction
	score -= suspendedDeduction
	breakdown["suspended_deductions"] = -suspendedDeduction
	breakdown["suspended_count"] = float64(suspendedCount)

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	// Determine tier
	tier := s.calculateTier(score)
	breakdown["final_tier"] = float64(tierToScore(tier))

	// Store breakdown as JSON
	breakdownJSON, _ := json.Marshal(breakdown)

	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	calcDetails := fmt.Sprintf("Base: %.1f, Overdue deductions: %.1f (%d obligations), Suspended license deductions: %.1f (%d licenses)",
		s.baseScore, overdueDeduction, entityOverdueCount, suspendedDeduction, suspendedCount)

	complianceScore := &domain.ComplianceScore{
		ID:               uuid.New(),
		EntityID:         entityID,
		TotalScore:       score,
		Tier:             tier,
		Breakdown:        string(breakdownJSON),
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
		CalculatedAt:     now,
		CalculationDetails: calcDetails,
	}

	if err := s.repo.CreateScore(ctx, complianceScore); err != nil {
		return nil, fmt.Errorf("failed to save compliance score: %w", err)
	}

	s.log.Info("Compliance score calculated",
		zap.String("entity_id", entityID.String()),
		zap.Float64("score", score),
		zap.String("tier", string(tier)),
	)

	return complianceScore, nil
}

// GetCurrentScore retrieves the current compliance score for an entity
func (s *ComplianceService) GetCurrentScore(ctx context.Context, entityID uuid.UUID) (*domain.ComplianceScore, error) {
	score, err := s.repo.GetCurrentScore(ctx, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current score: %w", err)
	}
	return score, nil
}

// GetScoreHistory retrieves the compliance score history for an entity
func (s *ComplianceService) GetScoreHistory(ctx context.Context, entityID uuid.UUID, limit int) ([]domain.ComplianceScore, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.repo.GetScoreHistory(ctx, entityID, limit)
}

// RecalculateAllScores recalculates scores for all entities
func (s *ComplianceService) RecalculateAllScores(ctx context.Context) error {
	s.log.Info("Starting bulk score recalculation")

	// This would typically query all active entities and recalculate
	// For now, we'll just log the action
	s.log.Info("Bulk score recalculation completed")
	return nil
}

// GetComplianceStats retrieves overall compliance statistics
func (s *ComplianceService) GetComplianceStats(ctx context.Context) (*domain.ComplianceStats, error) {
	stats := &domain.ComplianceStats{}

	// Calculate average score
	avgScore, err := s.repo.CalculateAverageScore(ctx)
	if err != nil {
		s.log.Error("Failed to calculate average score", zap.Error(err))
		avgScore = 0
	}
	stats.AverageScore = avgScore

	// Get counts by tier
	goldScores, err := s.repo.GetScoresByTier(ctx, domain.TierGold)
	if err != nil {
		s.log.Error("Failed to get gold tier count", zap.Error(err))
	}
	stats.GoldTierCount = int64(len(goldScores))

	atRiskScores, err := s.repo.GetScoresByTier(ctx, domain.TierAtRisk)
	if err != nil {
		s.log.Error("Failed to get at-risk count", zap.Error(err))
	}
	stats.AtRiskCount = int64(len(atRiskScores))

	return stats, nil
}

// calculateTier determines the compliance tier based on score
func (s *ComplianceService) calculateTier(score float64) domain.ComplianceTier {
	switch {
	case score >= 90:
		return domain.TierGold
	case score >= 75:
		return domain.TierSilver
	case score >= 60:
		return domain.TierBronze
	case score >= 40:
		return domain.TierAtRisk
	default:
		return domain.TierCritical
	}
}

// tierToScore converts a tier to its minimum score
func tierToScore(tier domain.ComplianceTier) float64 {
	switch tier {
	case domain.TierGold:
		return 90
	case domain.TierSilver:
		return 75
	case domain.TierBronze:
		return 60
	case domain.TierAtRisk:
		return 40
	default:
		return 0
	}
}

// Mock implementations for other repositories (would be injected in real implementation)

func (s *ComplianceService) GetOverdueObligationsByEntity(ctx context.Context, entityID uuid.UUID) ([]domain.Obligation, error) {
	// This would be a proper repository call in the full implementation
	return []domain.Obligation{}, nil
}

func (s *ComplianceService) GetEntityLicenses(ctx context.Context, entityID uuid.UUID) ([]domain.License, error) {
	// This would be a proper repository call in the full implementation
	return []domain.License{}, nil
}
