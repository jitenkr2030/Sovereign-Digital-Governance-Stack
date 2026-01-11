package service

import (
	"context"
	"fmt"
	"smart-cities/internal/repository"
	"sync"
	"time"
)

// LightingService handles smart public lighting logic
type LightingService struct {
	postgresRepo *repository.PostgresRepository
	redisRepo    *repository.RedisRepository
	mu           sync.RWMutex
}

// NewLightingService creates a new lighting service instance
func NewLightingService(pg *repository.PostgresRepository, rd *repository.RedisRepository) *LightingService {
	return &LightingService{
		postgresRepo: pg,
		redisRepo:    rd,
	}
}

// CalculateOptimalBrightness calculates optimal brightness based on conditions
func (s *LightingService) CalculateOptimalBrightness(ctx context.Context, fixtureID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	fixture, err := s.postgresRepo.GetLightingFixture(ctx, fixtureID)
	if err != nil {
		return 0, fmt.Errorf("failed to get fixture: %w", err)
	}
	
	// Get ambient light level
	ambientLevel, err := s.redisRepo.GetAmbientLightLevel(ctx, fixture.Location)
	if err != nil {
		ambientLevel = 50.0 // Default ambient level
	}
	
	// Get current time
	currentHour := time.Now().Hour()
	
	// Get traffic/pedestrian activity level
	activityLevel, err := s.redisRepo.GetActivityLevel(ctx, fixture.Location)
	if err != nil {
		activityLevel = s.getDefaultActivityLevel(currentHour)
	}
	
	// Calculate brightness
	brightness := s.calculateBrightness(ambientLevel, activityLevel, currentHour, fixture.Type)
	
	return brightness, nil
}

// calculateBrightness determines brightness percentage based on multiple factors
func (s *LightingService) calculateBrightness(ambient float64, activity float64, hour int, fixtureType string) int {
	// Base brightness based on ambient light
	baseBrightness := 100 - int(ambient)
	
	// Adjust for time of day
	var timeAdjustment int
	switch {
	case hour >= 22 || hour <= 5:
		timeAdjustment = 20 // Higher brightness at night for safety
	case hour >= 6 && hour <= 8:
		timeAdjustment = -30 // Dimmer during dawn
	case hour >= 17 && hour <= 21:
		timeAdjustment = 10 // Slightly brighter during evening peak
	default:
		timeAdjustment = 0
	}
	
	// Adjust for activity level
	activityAdjustment := int((activity - 50) * 0.4) // ±20% based on activity
	
	// Fixture type adjustment
	typeMultiplier := 1.0
	switch fixtureType {
	case "STREET_LIGHT":
		typeMultiplier = 1.0
	case "PARK_LIGHT":
		typeMultiplier = 0.8
	case "FLOOD_LIGHT":
		typeMultiplier = 1.2
	case "PATHWAY_LIGHT":
		typeMultiplier = 0.6
	}
	
	totalBrightness := float64(baseBrightness + timeAdjustment + activityAdjustment)
	totalBrightness *= typeMultiplier
	
	// Clamp between 0 and 100
	if totalBrightness < 0 {
		totalBrightness = 0
	} else if totalBrightness > 100 {
		totalBrightness = 100
	}
	
	return int(totalBrightness)
}

// getDefaultActivityLevel provides default activity levels based on hour
func (s *LightingService) getDefaultActivityLevel(hour int) float64 {
	switch {
	case hour >= 7 && hour <= 9:
		return 80.0 // Morning rush
	case hour >= 17 && hour <= 19:
		return 85.0 // Evening rush
	case hour >= 10 && hour <= 16:
		return 50.0 // Midday
	case hour >= 20 && hour <= 22:
		return 40.0 // Evening
	default:
		return 15.0 // Night
	}
}

// SetFixtureBrightness sets the brightness of a lighting fixture
func (s *LightingService) SetFixtureBrightness(ctx context.Context, fixtureID string, brightness int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if brightness < 0 || brightness > 100 {
		return fmt.Errorf("brightness must be between 0 and 100")
	}
	
	fixture, err := s.postgresRepo.GetLightingFixture(ctx, fixtureID)
	if err != nil {
		return fmt.Errorf("failed to get fixture: %w", err)
	}
	
	fixture.CurrentBrightness = brightness
	fixture.LastUpdated = time.Now()
	
	if err := s.postgresRepo.UpdateLightingFixture(ctx, fixture); err != nil {
		return fmt.Errorf("failed to update fixture: %w", err)
	}
	
	// Update cache
	s.redisRepo.SetFixtureBrightness(ctx, fixtureID, brightness, 5*time.Minute)
	
	return nil
}

// AdjustZoneLighting adjusts lighting for an entire zone
func (s *LightingService) AdjustZoneLighting(ctx context.Context, zoneID string, mode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	fixtures, err := s.postgresRepo.GetLightingFixturesByZone(ctx, zoneID)
	if err != nil {
		return fmt.Errorf("failed to get fixtures: %w", err)
	}
	
	for _, fixture := range fixtures {
		var brightness int
		switch mode {
		case "NIGHT":
			brightness = 70
		case "EVENING":
			brightness = 85
		case "DAWN":
			brightness = 50
		case "DAYTIME":
			brightness = 0
		case "EMERGENCY":
			brightness = 100
		default:
			// Calculate optimal brightness
			b, _ := s.CalculateOptimalBrightness(ctx, fixture.ID)
			brightness = b
		}
		
		fixture.CurrentBrightness = brightness
		fixture.LastUpdated = time.Now()
		
		if err := s.postgresRepo.UpdateLightingFixture(ctx, &fixture); err != nil {
			return fmt.Errorf("failed to update fixture %s: %w", fixture.ID, err)
		}
		
		// Update cache
		s.redisRepo.SetFixtureBrightness(ctx, fixture.ID, brightness, 5*time.Minute)
	}
	
	return nil
}

// HandleMotionEvent handles motion detection events
func (s *LightingService) HandleMotionEvent(ctx context.Context, fixtureID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	fixture, err := s.postgresRepo.GetLightingFixture(ctx, fixtureID)
	if err != nil {
		return fmt.Errorf("failed to get fixture: %w", err)
	}
	
	if !fixture.HasMotionSensor {
		return nil // No motion sensor on this fixture
	}
	
	// Boost brightness on motion detection
	boostedBrightness := fixture.CurrentBrightness + 30
	if boostedBrightness > 100 {
		boostedBrightness = 100
	}
	
	fixture.CurrentBrightness = boostedBrightness
	fixture.LastMotionDetected = time.Now()
	fixture.LastUpdated = time.Now()
	
	if err := s.postgresRepo.UpdateLightingFixture(ctx, fixture); err != nil {
		return fmt.Errorf("failed to update fixture: %w", err)
	}
	
	// Schedule dimming back after motion stops
	go s.scheduleDimming(ctx, fixtureID, 2*time.Minute)
	
	// Update cache
	s.redisRepo.SetFixtureBrightness(ctx, fixtureID, boostedBrightness, 5*time.Minute)
	
	return nil
}

// scheduleDimming gradually reduces brightness after motion stops
func (s *LightingService) scheduleDimming(ctx context.Context, fixtureID string, delay time.Duration) {
	time.Sleep(delay)
	
	// Check if new motion has occurred
	lastMotion, err := s.redisRepo.GetLastMotionTime(ctx, fixtureID)
	if err == nil && time.Since(lastMotion) < delay {
		return // New motion detected, skip dimming
	}
	
	// Calculate optimal brightness
	brightness, _ := s.CalculateOptimalBrightness(ctx, fixtureID)
	
	// Apply gradual dimming
	_ = s.SetFixtureBrightness(ctx, fixtureID, brightness)
}

// ActivateEmergencyLighting activates emergency lighting mode in a zone
func (s *LightingService) ActivateEmergencyLighting(ctx context.Context, zoneID string) error {
	return s.AdjustZoneLighting(ctx, zoneID, "EMERGENCY")
}

// DeactivateEmergencyLighting returns zone to normal operation
func (s *LightingService) DeactivateEmergencyLighting(ctx context.Context, zoneID string) error {
	return s.AdjustZoneLighting(ctx, zoneID, "NORMAL")
}

// GetLightingMetrics provides lighting system statistics
func (s *LightingService) GetLightingMetrics(ctx context.Context, timeRange time.Duration) (*LightingMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	since := time.Now().Add(-timeRange)
	
	fixtures, err := s.postgresRepo.GetAllLightingFixtures(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get fixtures: %w", err)
	}
	
	metrics := &LightingMetrics{
		TimeRange:      timeRange,
		TotalFixtures:  len(fixtures),
		ActiveFixtures: 0,
		AvgBrightness:  0,
		EnergyUsage:    0,
		ByZone:         make(map[string]int),
		ByStatus:       make(map[string]int),
		ByType:         make(map[string]int),
	}
	
	var totalBrightness int
	var totalEnergy float64
	
	for _, fixture := range fixtures {
		totalBrightness += fixture.CurrentBrightness
		
		// Calculate energy usage (watts * hours)
		watts := float64(fixture.Wattage) * float64(fixture.CurrentBrightness) / 100.0
		totalEnergy += watts * timeRange.Hours()
		
		if fixture.CurrentBrightness > 0 {
			metrics.ActiveFixtures++
		}
		
		metrics.ByZone[fixture.Zone]++
		metrics.ByType[fixture.Type]++
		
		status := "ACTIVE"
		if fixture.CurrentBrightness == 0 {
			status = "OFF"
		} else if fixture.CurrentBrightness < 30 {
			status = "DIMMED"
		}
		if fixture.Status != "OPERATIONAL" {
			status = "FAULT"
		}
		metrics.ByStatus[status]++
	}
	
	if len(fixtures) > 0 {
		metrics.AvgBrightness = totalBrightness / len(fixtures)
	}
	metrics.EnergyUsage = totalEnergy / 1000.0 // Convert to kWh
	
	return metrics, nil
}

// PredictMaintenance predicts when maintenance will be needed
func (s *LightingService) PredictMaintenance(fixture *LightingFixture) time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Calculate hours since last maintenance
	hoursSinceMaintenance := time.Since(fixture.LastMaintenance).Hours()
	
	// Estimate remaining life based on usage patterns
	// LED fixtures typically last 50,000 hours
	usageFactor := float64(fixture.OperatingHours) / 50000.0
	
	// Adjust for environment
	environmentFactor := 1.0
	switch fixture.Environment {
	case "COASTAL":
		environmentFactor = 0.8 // Faster degradation
	case "INDUSTRIAL":
		environmentFactor = 0.85
	case "RESIDENTIAL":
		environmentFactor = 1.1 // Slower degradation
	}
	
	// Calculate estimated remaining life
	remainingLife := (1.0 - usageFactor) * environmentFactor * 50000.0
	
	// Convert to time based on average daily usage (assuming 10 hours/day)
	daysRemaining := remainingLife / 10.0
	
	return time.Now().AddDate(0, 0, int(daysRemaining))
}

// GetEnergySavings calculates energy savings from smart lighting
func (s *LightingService) GetEnergySavings(ctx context.Context, baselineEnergy float64) (*EnergySavings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Get current energy usage
	metrics, err := s.GetLightingMetrics(ctx, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	
	currentEnergy := metrics.EnergyUsage
	
	savings := &EnergySavings{
		Period:              24 * time.Hour,
		BaselineEnergy:      baselineEnergy,
		ActualEnergy:        currentEnergy,
		EnergySaved:         baselineEnergy - currentEnergy,
		SavingsPercentage:   (baselineEnergy - currentEnergy) / baselineEnergy * 100,
		CostSavings:         (baselineEnergy - currentEnergy) * 0.12, // $0.12 per kWh average
		CarbonReduction:     (baselineEnergy - currentEnergy) * 0.4, // 0.4 kg CO2 per kWh
	}
	
	return savings, nil
}

// LightingMetrics contains lighting system statistics
type LightingMetrics struct {
	TimeRange      time.Duration
	TotalFixtures  int
	ActiveFixtures int
	AvgBrightness  int
	EnergyUsage    float64 // in kWh
	ByZone         map[string]int
	ByStatus       map[string]int
	ByType         map[string]int
}

// EnergySavings contains energy savings data
type EnergySavings struct {
	Period            time.Duration
	BaselineEnergy    float64
	ActualEnergy      float64
	EnergySaved       float64
	SavingsPercentage float64
	CostSavings       float64
	CarbonReduction   float64 // in kg CO2
}
