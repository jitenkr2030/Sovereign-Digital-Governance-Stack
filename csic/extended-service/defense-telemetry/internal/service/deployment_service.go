package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"defense-telemetry/internal/domain"
	"defense-telemetry/internal/repository"

	"github.com/shopspring/decimal"
)

// DeploymentService handles unit deployment and strategic positioning
type DeploymentService struct {
	repo   *repository.PostgresRepository
	redis  *repository.RedisRepository
	config *Config
}

// NewDeploymentService creates a new deployment service
func NewDeploymentService(repo *repository.PostgresRepository, redis *repository.RedisRepository, config *Config) *DeploymentService {
	return &DeploymentService{
		repo:   repo,
		redis:  redis,
		config: config,
	}
}

// CreateDeployment creates a new deployment
func (s *DeploymentService) CreateDeployment(ctx context.Context, deployment *domain.Deployment) error {
	// Validate
	if deployment.UnitID == "" {
		return fmt.Errorf("unit_id is required")
	}
	if deployment.ZoneID == "" {
		return fmt.Errorf("zone_id is required")
	}

	// Set defaults
	if deployment.ID == "" {
		deployment.ID = fmt.Sprintf("deploy-%d", time.Now().UnixNano())
	}
	deployment.CreatedAt = time.Now()
	deployment.UpdatedAt = time.Now()
	deployment.Status = domain.DeploymentStatusPlanned

	// Check zone capacity
	currentUnits, err := s.redis.GetZoneUnitCount(ctx, deployment.ZoneID)
	if err != nil {
		return fmt.Errorf("failed to check zone capacity: %w", err)
	}

	zoneStatus, _ := s.redis.GetCachedZoneStatus(ctx, deployment.ZoneID)
	var zone domain.StrategicZone
	if zoneStatus != nil {
		json.Unmarshal(zoneStatus, &zone)
	}

	if zone.MaxUnits > 0 && int(currentUnits) >= zone.MaxUnits {
		return fmt.Errorf("zone %s is at maximum capacity", deployment.ZoneID)
	}

	// Calculate route if not provided
	if len(deployment.RoutePlanned.Waypoints) == 0 {
		if err := s.calculateRoute(deployment); err != nil {
			fmt.Printf("Warning: failed to calculate route: %v\n", err)
		}
	}

	if err := s.repo.CreateDeployment(ctx, deployment); err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	// Increment zone unit count
	if err := s.redis.IncrementZoneUnitCount(ctx, deployment.ZoneID); err != nil {
		fmt.Printf("Warning: failed to increment zone unit count: %v\n", err)
	}

	// Update unit deployment status
	unit := &domain.DefenseUnit{
		ID:             deployment.UnitID,
		IsDeployed:     true,
		DeploymentZone: deployment.ZoneID,
		LastContactAt:  time.Now(),
	}
	if err := s.repo.UpdateDefenseUnit(ctx, unit); err != nil {
		fmt.Printf("Warning: failed to update unit status: %v\n", err)
	}

	// Cache deployment
	ttl := 24 * time.Hour
	if err := s.redis.CacheDeployment(ctx, deployment.ID, deployment, ttl); err != nil {
		fmt.Printf("Warning: failed to cache deployment: %v\n", err)
	}

	return nil
}

// calculateRoute calculates the optimal route for deployment
func (s *DeploymentService) calculateRoute(deployment *domain.Deployment) error {
	// Calculate distance
	distance := s.calculateDistance(deployment.StartLocation, deployment.TargetLocation)

	// Estimate time (simple calculation)
	avgSpeed := decimal.NewFromFloat(60) // 60 km/h
	estimatedTime := time.Duration(distance.InexactFloat64() / avgSpeed.InexactFloat64() * float64(time.Hour))

	deployment.RoutePlanned = domain.RouteData{
		TotalDistance:   distance,
		EstimatedTime:   estimatedTime,
		Waypoints:       []domain.Coordinates{deployment.StartLocation, deployment.TargetLocation},
		FuelRequired:    distance.Mul(decimal.NewFromFloat(0.1)), // 0.1 liters per km
		Hazards:         []string{},
	}

	return nil
}

// calculateDistance calculates distance between two coordinates in km
func (s *DeploymentService) calculateDistance(start, end domain.Coordinates) decimal.Decimal {
	// Haversine formula
	const earthRadius = 6371 // km

	lat1 := start.Latitude.Mul(decimal.NewFromFloat(math.Pi / 180))
	lat2 := end.Latitude.Mul(decimal.NewFromFloat(math.Pi / 180))
	dLat := end.Latitude.Sub(start.Latitude).Mul(decimal.NewFromFloat(math.Pi / 180))
	dLng := end.Longitude.Sub(start.Longitude).Mul(decimal.NewFromFloat(math.Pi / 180))

	a := decimal.NewFromFloat(math.Pow(math.Sin(float64(dLat)/2), 2)).
		Add(lat1.Mul(lat2).Cos().Mul(decimal.NewFromFloat(math.Pow(math.Sin(float64(dLng)/2), 2))))
	c := decimal.NewFromFloat(2 * math.Atan2(math.Sqrt(float64(a)), math.Sqrt(1-a)))

	distance := decimal.NewFromFloat(earthRadius).Mul(c)

	// Round to 2 decimal places
	return distance.Round(2)
}

// GetDeployment retrieves a deployment by ID
func (s *DeploymentService) GetDeployment(ctx context.Context, id string) (*domain.Deployment, error) {
	// Try cache first
	cached, err := s.redis.GetCachedDeployment(ctx, id)
	if err == nil && cached != nil {
		var deployment domain.Deployment
		if err := json.Unmarshal(cached, &deployment); err == nil {
			return &deployment, nil
		}
	}

	// Fall back to querying deployments by unit
	deployments, err := s.repo.GetUnitDeployments(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments: %w", err)
	}

	for _, d := range deployments {
		if d.ID == id {
			return d, nil
		}
	}

	return nil, fmt.Errorf("deployment not found: %s", id)
}

// GetUnitDeployments retrieves all deployments for a unit
func (s *DeploymentService) GetUnitDeployments(ctx context.Context, unitID string) ([]*domain.Deployment, error) {
	return s.repo.GetUnitDeployments(ctx, unitID)
}

// GetActiveDeployments retrieves all active deployments
func (s *DeploymentService) GetActiveDeployments(ctx context.Context) ([]*domain.Deployment, error) {
	deployments, err := s.repo.GetUnitDeployments(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments: %w", err)
	}

	var active []*domain.Deployment
	for _, d := range deployments {
		if d.Status == domain.DeploymentStatusEnRoute ||
			d.Status == domain.DeploymentStatusOnStation ||
			d.Status == domain.DeploymentStatusReturning {
			active = append(active, d)
		}
	}

	return active, nil
}

// UpdateDeploymentStatus updates the status of a deployment
func (s *DeploymentService) UpdateDeploymentStatus(ctx context.Context, id string, status domain.DeploymentStatus) error {
	deployments, err := s.repo.GetUnitDeployments(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get deployments: %w", err)
	}

	for _, d := range deployments {
		if d.ID == id {
			d.Status = status
			d.UpdatedAt = time.Now()

			// If completed or aborted, decrement zone unit count
			if status == domain.DeploymentStatusCompleted || status == domain.DeploymentStatusAborted {
				if err := s.redis.DecrementZoneUnitCount(ctx, d.ZoneID); err != nil {
					fmt.Printf("Warning: failed to decrement zone unit count: %v\n", err)
				}

				// Update unit status
				unit := &domain.DefenseUnit{
					ID:            d.UnitID,
					IsDeployed:    false,
					DeploymentZone: "",
				}
				_ = s.repo.UpdateDefenseUnit(ctx, unit)
			}

			return nil
		}
	}

	return fmt.Errorf("deployment not found: %s", id)
}

// OptimizeDeployment optimizes unit deployment based on current conditions
func (s *DeploymentService) OptimizeDeployment(ctx context.Context, unitID string) (*DeploymentRecommendation, error) {
	// Get current unit position
	telemetry, err := s.repo.GetLatestTelemetry(ctx, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get telemetry: %w", err)
	}
	if telemetry == nil {
		return nil, fmt.Errorf("no telemetry available for unit: %s", unitID)
	}

	// Get strategic zones
	zones, err := s.repo.GetStrategicZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get strategic zones: %w", err)
	}

	// Find optimal zone
	var optimalZone *domain.StrategicZone
	var minDistance decimal.Decimal

	for _, zone := range zones {
		if zone.Status != domain.ZoneStatusActive {
			continue
		}

		// Check zone capacity
		currentUnits, _ := s.redis.GetZoneUnitCount(ctx, zone.ID)
		if zone.MaxUnits > 0 && int(currentUnits) >= zone.MaxUnits {
			continue
		}

		// Calculate distance
		distance := s.calculateDistance(telemetry.Coordinates, zone.CenterPoint)

		if optimalZone == nil || distance.LessThan(minDistance) {
			optimalZone = zone
			minDistance = distance
		}
	}

	if optimalZone == nil {
		return nil, fmt.Errorf("no suitable zone found for optimization")
	}

	recommendation := &DeploymentRecommendation{
		UnitID:       unitID,
		CurrentZone:  "", // Would need to be tracked
		RecommendedZone: optimalZone,
		Distance:     minDistance,
		Priority:     optimalZone.Priority,
		Reason:       "Strategic coverage optimization",
		EstimatedTime: time.Duration(minDistance.InexactFloat64() / 60 * float64(time.Hour)),
	}

	return recommendation, nil
}

// DeploymentRecommendation represents a deployment optimization recommendation
type DeploymentRecommendation struct {
	UnitID          string             `json:"unit_id"`
	CurrentZone     string             `json:"current_zone,omitempty"`
	RecommendedZone *domain.StrategicZone `json:"recommended_zone"`
	Distance        decimal.Decimal    `json:"distance"`
	Priority        int                `json:"priority"`
	Reason          string             `json:"reason"`
	EstimatedTime   time.Duration      `json:"estimated_time"`
}

// GetStrategicZones retrieves all strategic zones
func (s *DeploymentService) GetStrategicZones(ctx context.Context) ([]*domain.StrategicZone, error) {
	return s.repo.GetStrategicZones(ctx)
}

// CreateStrategicZone creates a new strategic zone
func (s *DeploymentService) CreateStrategicZone(ctx context.Context, zone *domain.StrategicZone) error {
	if zone.ID == "" {
		zone.ID = fmt.Sprintf("zone-%d", time.Now().UnixNano())
	}
	zone.CreatedAt = time.Now()
	zone.UpdatedAt = time.Now()

	if zone.Status == "" {
		zone.Status = domain.ZoneStatusActive
	}

	return s.repo.CreateStrategicZone(ctx, zone)
}

// GetDeploymentStats retrieves deployment statistics
func (s *DeploymentService) GetDeploymentStats(ctx context.Context) (*DeploymentStats, error) {
	deployments, err := s.GetActiveDeployments(ctx)
	if err != nil {
		return nil, err
	}

	zoneCounts := make(map[string]int)
	for _, d := range deployments {
		zoneCounts[d.ZoneID]++
	}

	zones, _ := s.repo.GetStrategicZones(ctx)

	stats := &DeploymentStats{
		ActiveDeployments: len(deployments),
		ZoneCounts:        zoneCounts,
		TotalZones:        len(zones),
		LastUpdated:       time.Now(),
	}

	return stats, nil
}

// DeploymentStats represents deployment statistics
type DeploymentStats struct {
	ActiveDeployments int               `json:"active_deployments"`
	ZoneCounts        map[string]int    `json:"zone_counts"`
	TotalZones        int               `json:"total_zones"`
	LastUpdated       time.Time         `json:"last_updated"`
}

// json unmarshal helper
func json.Unmarshal(data []byte, v interface{}) error {
	return fmt.Errorf("json.Unmarshal not imported")
}
