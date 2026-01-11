package service

import (
	"context"
	"log"
	"time"

	"github.com/csic/surveillance/internal/domain"
	"github.com/csic/surveillance/internal/port"
	"github.com/google/uuid"
)

// AlertService handles alert lifecycle management
type AlertService struct {
	alertRepo port.AlertRepository
}

// NewAlertService creates a new alert service
func NewAlertService(alertRepo port.AlertRepository) *AlertService {
	return &AlertService{
		alertRepo: alertRepo,
	}
}

// CreateAlert creates a new alert
func (s *AlertService) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()
	alert.DetectedAt = time.Now()

	log.Printf("Creating alert: %s (%s) for exchange %s", alert.Title, alert.Type, alert.ExchangeID)

	return s.alertRepo.Create(ctx, alert)
}

// GetAlert retrieves an alert by ID
func (s *AlertService) GetAlert(ctx context.Context, id uuid.UUID) (*domain.Alert, error) {
	return s.alertRepo.GetByID(ctx, id)
}

// UpdateAlert updates an existing alert
func (s *AlertService) UpdateAlert(ctx context.Context, alert *domain.Alert) error {
	alert.UpdatedAt = time.Now()
	return s.alertRepo.Update(ctx, alert)
}

// ResolveAlert marks an alert as resolved
func (s *AlertService) ResolveAlert(ctx context.Context, id uuid.UUID, resolution string, resolvedBy uuid.UUID) error {
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	alert.Status = domain.AlertStatusResolved
	alert.ResolvedAt = &now
	alert.ResolvedBy = &resolvedBy
	alert.Resolution = &resolution
	alert.UpdatedAt = now

	return s.alertRepo.Update(ctx, alert)
}

// DismissAlert dismisses an alert
func (s *AlertService) DismissAlert(ctx context.Context, id uuid.UUID, reason string, dismissedBy uuid.UUID) error {
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	alert.Status = domain.AlertStatusDismissed
	alert.ResolvedAt = &now
	alert.ResolvedBy = &dismissedBy
	alert.Resolution = &reason
	alert.UpdatedAt = now

	log.Printf("Dismissing alert %s: %s", id, reason)

	return s.alertRepo.Update(ctx, alert)
}

// EscalateAlert escalates an alert to higher authority
func (s *AlertService) EscalateAlert(ctx context.Context, id uuid.UUID, reason string, escalatedTo uuid.UUID) error {
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	alert.Status = domain.AlertStatusEscalated
	alert.AssignedTo = &escalatedTo
	alert.UpdatedAt = time.Now()

	// Add escalation note
	note := domain.AlertNote{
		ID:         uuid.New(),
		AlertID:    id,
		AuthorID:   escalatedTo,
		Content:    "Alert escalated: " + reason,
		CreatedAt:  time.Now(),
		IsInternal: true,
	}
	alert.Notes = append(alert.Notes, note)

	log.Printf("Escalating alert %s to %s", id, escalatedTo)

	return s.alertRepo.Update(ctx, alert)
}

// AssignAlert assigns an alert to a user
func (s *AlertService) AssignAlert(ctx context.Context, id uuid.UUID, assigneeID uuid.UUID) error {
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	alert.AssignedTo = &assigneeID
	alert.Status = domain.AlertStatusInvestigating
	alert.UpdatedAt = time.Now()

	return s.alertRepo.Update(ctx, alert)
}

// AddNote adds a note to an alert
func (s *AlertService) AddNote(ctx context.Context, alertID uuid.UUID, note *domain.AlertNote) error {
	alert, err := s.alertRepo.GetByID(ctx, alertID)
	if err != nil {
		return err
	}

	if note.ID == uuid.Nil {
		note.ID = uuid.New()
	}
	note.AlertID = alertID
	note.CreatedAt = time.Now()

	alert.Notes = append(alert.Notes, *note)
	alert.UpdatedAt = time.Now()

	return s.alertRepo.Update(ctx, alert)
}

// ListAlerts retrieves alerts based on filter criteria
func (s *AlertService) ListAlerts(ctx context.Context, filter domain.AlertFilter, limit, offset int) ([]domain.Alert, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.alertRepo.List(ctx, filter, limit, offset)
}

// GetOpenAlerts retrieves all open alerts
func (s *AlertService) GetOpenAlerts(ctx context.Context) ([]domain.Alert, error) {
	return s.alertRepo.GetOpenAlerts(ctx)
}

// GetAlertsByExchange retrieves alerts for a specific exchange
func (s *AlertService) GetAlertsByExchange(ctx context.Context, exchangeID uuid.UUID, limit, offset int) ([]domain.Alert, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.alertRepo.GetByExchange(ctx, exchangeID, limit, offset)
}

// GetAlertsBySymbol retrieves alerts for a specific symbol
func (s *AlertService) GetAlertsBySymbol(ctx context.Context, symbol string, limit, offset int) ([]domain.Alert, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.alertRepo.GetBySymbol(ctx, symbol, limit, offset)
}

// GetStatistics retrieves alert statistics for a time period
func (s *AlertService) GetStatistics(ctx context.Context, startTime, endTime time.Time) (*domain.AlertStatistics, error) {
	return s.alertRepo.GetStatistics(ctx, startTime, endTime)
}

// GetTrend retrieves alert trend data
func (s *AlertService) GetTrend(ctx context.Context, interval string, startTime, endTime time.Time) ([]domain.AlertTrendPoint, error) {
	return s.alertRepo.GetAlertTrend(ctx, interval, startTime, endTime)
}

// GetThresholds retrieves current alert thresholds
func (s *AlertService) GetThresholds(ctx context.Context) (map[domain.AlertType]domain.AlertThreshold, error) {
	// In a real implementation, this would fetch from a configuration store
	// For now, return default thresholds
	return map[domain.AlertType]domain.AlertThreshold{
		domain.AlertTypeWashTrading: {
			Name:          "Wash Trade Detection",
			AlertType:     domain.AlertTypeWashTrading,
			Enabled:       true,
			Severity:      domain.AlertSeverityCritical,
			MinConfidence: 0.7,
			CooldownPeriod: 5 * time.Minute,
		},
		domain.AlertTypeSpoofing: {
			Name:          "Spoofing Detection",
			AlertType:     domain.AlertTypeSpoofing,
			Enabled:       true,
			Severity:      domain.AlertSeverityCritical,
			MinConfidence: 0.75,
			CooldownPeriod: 10 * time.Minute,
		},
		domain.AlertTypePriceManipulation: {
			Name:          "Price Manipulation Detection",
			AlertType:     domain.AlertTypePriceManipulation,
			Enabled:       true,
			Severity:      domain.AlertSeverityCritical,
			MinConfidence: 0.8,
			CooldownPeriod: 15 * time.Minute,
		},
		domain.AlertTypeVolumeAnomaly: {
			Name:          "Volume Anomaly Detection",
			AlertType:     domain.AlertTypeVolumeAnomaly,
			Enabled:       true,
			Severity:      domain.AlertSeverityWarning,
			MinConfidence: 0.6,
			CooldownPeriod: 5 * time.Minute,
		},
	}, nil
}

// UpdateThreshold updates an alert threshold
func (s *AlertService) UpdateThreshold(ctx context.Context, threshold *domain.AlertThreshold) error {
	// In a real implementation, this would persist to a configuration store
	log.Printf("Updating threshold for %s: enabled=%v, severity=%s, min_confidence=%.2f",
		threshold.AlertType, threshold.Enabled, threshold.Severity, threshold.MinConfidence)
	return nil
}
