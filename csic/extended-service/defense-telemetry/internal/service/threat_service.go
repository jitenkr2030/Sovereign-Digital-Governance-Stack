package service

import (
	"context"
	"fmt"
	"time"

	"defense-telemetry/internal/domain"
	"defense-telemetry/internal/repository"

	"github.com/shopspring/decimal"
)

// ThreatService handles threat detection and management
type ThreatService struct {
	repo   *repository.PostgresRepository
	redis  *repository.RedisRepository
	config *Config
}

// NewThreatService creates a new threat service
func NewThreatService(repo *repository.PostgresRepository, redis *repository.RedisRepository, config *Config) *ThreatService {
	return &ThreatService{
		repo:   repo,
		redis:  redis,
		config: config,
	}
}

// CreateThreatEvent creates a new threat event
func (s *ThreatService) CreateThreatEvent(ctx context.Context, threat *domain.ThreatEvent) error {
	// Validate threat
	if err := s.validateThreat(threat); err != nil {
		return fmt.Errorf("invalid threat: %w", err)
	}

	// Set default values
	if threat.ID == "" {
		threat.ID = fmt.Sprintf("threat-%d", time.Now().UnixNano())
	}
	threat.FirstSeenAt = time.Now()
	threat.DetectionTime = time.Now()
	threat.LastUpdatedAt = time.Now()

	// Save to database
	if err := s.repo.CreateThreatEvent(ctx, threat); err != nil {
		return fmt.Errorf("failed to create threat: %w", err)
	}

	// Add to active threats set
	if threat.Status != domain.ThreatStatusResolved {
		if err := s.redis.AddThreatToActiveSet(ctx, threat.ID); err != nil {
			fmt.Printf("Warning: failed to add threat to active set: %v\n", err)
		}
	}

	// Create alert for high-severity threats
	if threat.Severity >= s.config.Threats.HighSeverityThreshold {
		alert := s.createThreatAlert(threat)
		if err := s.repo.CreateAlert(ctx, alert); err != nil {
			fmt.Printf("Warning: failed to create alert: %v\n", err)
		}
	}

	// Increment threat counter
	if err := s.redis.IncrementCounter(ctx, "threats_detected"); err != nil {
		fmt.Printf("Warning: failed to increment threat counter: %v\n", err)
	}

	return nil
}

// validateThreat validates a threat event
func (s *ThreatService) validateThreat(threat *domain.ThreatEvent) error {
	if threat.UnitID == "" && threat.ThreatSource == "" {
		return fmt.Errorf("either unit_id or threat_source is required")
	}
	if threat.Severity < 1 || threat.Severity > 5 {
		return fmt.Errorf("severity must be between 1 and 5")
	}
	return nil
}

// createThreatAlert creates an alert for a threat
func (s *ThreatService) createThreatAlert(threat *domain.ThreatEvent) *domain.Alert {
	var severity domain.AlertSeverity
	if threat.Severity >= s.config.Threats.CriticalSeverityThreshold {
		severity = domain.AlertSeverityEmergency
	} else if threat.Severity >= s.config.Threats.HighSeverityThreshold {
		severity = domain.AlertSeverityCritical
	} else {
		severity = domain.AlertSeverityWarning
	}

	return &domain.Alert{
		ID:        "",
		AlertType: domain.AlertTypeThreat,
		Severity:  severity,
		Title:     fmt.Sprintf("Threat Detected: %s", threat.ThreatType),
		Message:   fmt.Sprintf("Threat type: %s, Severity: %d, Source: %s", threat.ThreatType, threat.Severity, threat.ThreatSource),
		Source:    "threat_detection",
		EntityID:  threat.ID,
		EntityType: "threat",
	}
}

// GetActiveThreats retrieves all active threats
func (s *ThreatService) GetActiveThreats(ctx context.Context, minSeverity int) ([]*domain.ThreatEvent, error) {
	if minSeverity <= 0 {
		minSeverity = 1
	}

	// Try cache first
	cached, err := s.redis.GetCachedActiveThreats(ctx)
	if err == nil && cached != nil {
		var threats []*domain.ThreatEvent
		if err := json.Unmarshal(cached, &threats); err == nil {
			// Filter by severity
			var filtered []*domain.ThreatEvent
			for _, t := range threats {
				if t.Severity >= minSeverity {
					filtered = append(filtered, t)
				}
			}
			return filtered, nil
		}
	}

	// Fall back to database
	threats, err := s.repo.GetActiveThreats(ctx, minSeverity)
	if err != nil {
		return nil, fmt.Errorf("failed to get active threats: %w", err)
	}

	// Cache the results
	ttl := 30 * time.Second
	if err := s.redis.CacheActiveThreats(ctx, threats, ttl); err != nil {
		fmt.Printf("Warning: failed to cache active threats: %v\n", err)
	}

	return threats, nil
}

// GetThreat retrieves a threat by ID
func (s *ThreatService) GetThreat(ctx context.Context, id string) (*domain.ThreatEvent, error) {
	threats, err := s.GetActiveThreats(ctx, 1)
	if err != nil {
		return nil, err
	}

	for _, t := range threats {
		if t.ID == id {
			return t, nil
		}
	}

	// If not found in active threats, it might be resolved
	// In a real implementation, you'd query the database directly
	return nil, fmt.Errorf("threat not found: %s", id)
}

// UpdateThreatEvent updates a threat event
func (s *ThreatService) UpdateThreatEvent(ctx context.Context, threat *domain.ThreatEvent) error {
	threat.LastUpdatedAt = time.Now()

	if err := s.repo.UpdateThreatEvent(ctx, threat); err != nil {
		return fmt.Errorf("failed to update threat: %w", err)
	}

	// If resolved, remove from active set
	if threat.Status == domain.ThreatStatusResolved || threat.Status == domain.ThreatStatusFalseAlarm {
		if err := s.redis.RemoveThreatFromActiveSet(ctx, threat.ID); err != nil {
			fmt.Printf("Warning: failed to remove threat from active set: %v\n", err)
		}
	}

	// Invalidate cache
	_ = s.redis.CacheActiveThreats(ctx, nil, 0)

	return nil
}

// ResolveThreat marks a threat as resolved
func (s *ThreatService) ResolveThreat(ctx context.Context, id string, resolution string) error {
	threat := &domain.ThreatEvent{
		ID:         id,
		Status:     domain.ThreatStatusResolved,
		Resolution: resolution,
	}

	now := time.Now()
	threat.ResolvedAt = &now

	return s.UpdateThreatEvent(ctx, threat)
}

// EscalateThreat escalates a threat to a higher level
func (s *ThreatService) EscalateThreat(ctx context.Context, id string, reason string) error {
	threat := &domain.ThreatEvent{
		ID:              id,
		EscalationLevel: 1,
		Notes:           reason,
	}

	// Auto-escalate if configured
	if s.config.Threats.AutoEscalate {
		threat.EscalationLevel = 2 // Automatic escalation
	}

	return s.UpdateThreatEvent(ctx, threat)
}

// GetThreatStats retrieves threat statistics
func (s *ThreatService) GetThreatStats(ctx context.Context) (*ThreatStats, error) {
	activeCount, err := s.redis.GetActiveThreatCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active threat count: %w", err)
	}

	detectedCount, err := s.redis.GetCounter(ctx, "threats_detected")
	if err != nil {
		return nil, fmt.Errorf("failed to get detected count: %w", err)
	}

	return &ThreatStats{
		ActiveThreats:  int(activeCount),
		TotalDetected:  int(detectedCount),
		HighSeverity:   0, // Would need to be calculated from database
		CriticalSeverity: 0,
		LastUpdated:    time.Now(),
	}, nil
}

// ThreatStats represents threat statistics
type ThreatStats struct {
	ActiveThreats      int       `json:"active_threats"`
	TotalDetected      int       `json:"total_detected"`
	HighSeverity       int       `json:"high_severity"`
	CriticalSeverity   int       `json:"critical_severity"`
	LastUpdated        time.Time `json:"last_updated"`
}

// json unmarshal helper
func json.Unmarshal(data []byte, v interface{}) error {
	return fmt.Errorf("json.Unmarshal not imported")
}
