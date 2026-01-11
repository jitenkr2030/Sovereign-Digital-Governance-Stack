package service

import (
	"context"
	"fmt"
	"math"
	"smart-cities/internal/repository"
	"sync"
	"time"
)

// PriorityLevel represents emergency priority levels
type PriorityLevel int

const (
	PriorityLow    PriorityLevel = 1
	PriorityMedium PriorityLevel = 2
	PriorityHigh   PriorityLevel = 3
	PriorityCritical PriorityLevel = 4
)

// EmergencyService handles emergency response coordination
type EmergencyService struct {
	postgresRepo *repository.PostgresRepository
	redisRepo    *repository.RedisRepository
	mu           sync.RWMutex
}

// NewEmergencyService creates a new emergency service instance
func NewEmergencyService(pg *repository.PostgresRepository, rd *repository.RedisRepository) *EmergencyService {
	return &EmergencyService{
		postgresRepo: pg,
		redisRepo:    rd,
	}
}

// CalculateResponseTime calculates estimated response time based on conditions
func (s *EmergencyService) CalculateResponseTime(distance float64, trafficLevel float64, priority PriorityLevel) time.Duration {
	// Base speed: 60 km/h in normal conditions
	baseTime := distance / 60.0 // hours
	
	// Adjust for traffic
	trafficFactor := 1.0 + (trafficLevel * 1.5) // Heavy traffic can triple response time
	
	// Adjust for priority (higher priority = faster response)
	priorityFactor := 1.0
	switch priority {
	case PriorityCritical:
		priorityFactor = 0.6 // 40% faster
	case PriorityHigh:
		priorityFactor = 0.75
	case PriorityMedium:
		priorityFactor = 0.9
	}
	
	adjustedTime := baseTime * trafficFactor * priorityFactor
	return time.Duration(adjustedTime * float64(time.Hour))
}

// FindNearestUnit finds the nearest available emergency unit
func (s *EmergencyService) FindNearestUnit(ctx context.Context, incidentLat, incidentLon float64, unitType string) (*EmergencyUnit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	units, err := s.postgresRepo.GetEmergencyUnitsByType(ctx, unitType)
	if err != nil {
		return nil, fmt.Errorf("failed to get emergency units: %w", err)
	}
	
	var nearestUnit *EmergencyUnit
	var minDistance := math.MaxFloat64
	
	for i := range units {
		if units[i].Status != "AVAILABLE" {
			continue
		}
		
		distance := haversineDistance(
			incidentLat, incidentLon,
			units[i].Latitude, units[i].Longitude,
		)
		
		if distance < minDistance {
			minDistance = distance
			nearestUnit = &units[i]
		}
	}
	
	if nearestUnit == nil {
		return nil, fmt.Errorf("no available units of type %s", unitType)
	}
	
	return nearestUnit, nil
}

// DispatchUnit dispatches an emergency unit to an incident
func (s *EmergencyService) DispatchUnit(ctx context.Context, unitID, incidentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Get unit and incident details
	unit, err := s.postgresRepo.GetEmergencyUnit(ctx, unitID)
	if err != nil {
		return fmt.Errorf("failed to get unit: %w", err)
	}
	
	incident, err := s.postgresRepo.GetEmergencyIncident(ctx, incidentID)
	if err != nil {
		return fmt.Errorf("failed to get incident: %w", err)
	}
	
	// Calculate response time
	distance := haversineDistance(
		unit.Latitude, unit.Longitude,
		incident.Latitude, incident.Longitude,
	)
	estimatedResponse := s.CalculateResponseTime(distance, incident.TrafficLevel, incident.Priority)
	
	// Update unit status
	unit.Status = "DISPATCHED"
	unit.CurrentIncidentID = incidentID
	if err := s.postgresRepo.UpdateEmergencyUnit(ctx, unit); err != nil {
		return fmt.Errorf("failed to update unit: %w", err)
	}
	
	// Update incident status
	incident.AssignedUnitID = unitID
	incident.Status = "DISPATCHED"
	incident.EstimatedArrival = time.Now().Add(estimatedResponse)
	if err := s.postgresRepo.UpdateEmergencyIncident(ctx, incident); err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}
	
	// Cache dispatch info
	s.redisRepo.SetEmergencyDispatch(ctx, unitID, incidentID, estimatedResponse)
	
	return nil
}

// ProcessIncident processes a new emergency incident
func (s *EmergencyService) ProcessIncident(ctx context.Context, incident *EmergencyIncident) (*EmergencyIncident, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Calculate priority based on incident type and details
	incident.Priority = s.calculatePriority(incident)
	
	// Calculate traffic impact
	incident.TrafficLevel = s.assessTrafficImpact(incident)
	
	// Set initial status
	incident.Status = "PENDING"
	incident.ReportedAt = time.Now()
	
	// Create incident in database
	if err := s.postgresRepo.CreateEmergencyIncident(ctx, incident); err != nil {
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}
	
	// Find and dispatch nearest appropriate unit
	unitType := s.getUnitTypeForIncident(incident.Type)
	unit, err := s.FindNearestUnit(ctx, incident.Latitude, incident.Longitude, unitType)
	if err != nil {
		incident.Status = "NO_UNIT_AVAILABLE"
		_ = s.postgresRepo.UpdateEmergencyIncident(ctx, incident)
		return incident, nil
	}
	
	// Dispatch the unit
	_ = s.DispatchUnit(ctx, unit.ID, incident.ID)
	
	return incident, nil
}

// calculatePriority determines the priority level of an incident
func (s *EmergencyService) calculatePriority(incident *EmergencyIncident) PriorityLevel {
	basePriority := PriorityMedium
	
	switch incident.Type {
	case "FIRE", "MEDICAL_EMERGENCY":
		basePriority = PriorityHigh
	case "HAZMAT", "STRUCTURE_COLLAPSE":
		basePriority = PriorityCritical
	case "TRAFFIC_ACCIDENT":
		if incident.InjuriesReported > 2 {
			basePriority = PriorityCritical
		} else if incident.InjuriesReported > 0 {
			basePriority = PriorityHigh
		}
	case "CRIME":
		if incident.Severity == "VIOLENT" {
			basePriority = PriorityHigh
		}
	}
	
	return basePriority
}

// assessTrafficImpact estimates the traffic impact of an incident
func (s *EmergencyService) assessTrafficImpact(incident *EmergencyIncident) float64 {
	baseImpact := 0.1
	
	switch incident.Type {
	case "TRAFFIC_ACCIDENT":
		baseImpact = 0.6
	case "FIRE", "STRUCTURE_COLLAPSE":
		baseImpact = 0.5
	case "HAZMAT":
		baseImpact = 0.8
	}
	
	// Adjust based on location type
	if incident.LocationType == "HIGHWAY" {
		baseImpact *= 1.5
	} else if incident.LocationType == "RESIDENTIAL" {
		baseImpact *= 0.7
	}
	
	return math.Min(1.0, baseImpact)
}

// getUnitTypeForIncident maps incident types to required unit types
func (s *EmergencyService) getUnitTypeForIncident(incidentType string) string {
	switch incidentType {
	case "FIRE":
		return "FIRE_TRUCK"
	case "MEDICAL_EMERGENCY":
		return "AMBULANCE"
	case "TRAFFIC_ACCIDENT":
		return "POLICE"
	case "CRIME":
		return "POLICE"
	case "HAZMAT":
		return "HAZMAT_UNIT"
	default:
		return "POLICE"
	}
}

// UpdateUnitLocation updates the location of an emergency unit
func (s *EmergencyService) UpdateUnitLocation(ctx context.Context, unitID string, lat, lon float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	unit, err := s.postgresRepo.GetEmergencyUnit(ctx, unitID)
	if err != nil {
		return fmt.Errorf("failed to get unit: %w", err)
	}
	
	unit.Latitude = lat
	unit.Longitude = lon
	unit.LastUpdated = time.Now()
	
	if err := s.postgresRepo.UpdateEmergencyUnit(ctx, unit); err != nil {
		return fmt.Errorf("failed to update unit: %w", err)
	}
	
	// Update cache
	s.redisRepo.SetUnitLocation(ctx, unitID, lat, lon, time.Minute)
	
	return nil
}

// ResolveIncident marks an incident as resolved
func (s *EmergencyService) ResolveIncident(ctx context.Context, incidentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	incident, err := s.postgresRepo.GetEmergencyIncident(ctx, incidentID)
	if err != nil {
		return fmt.Errorf("failed to get incident: %w", err)
	}
	
	incident.Status = "RESOLVED"
	incident.ResolvedAt = time.Now()
	
	if err := s.postgresRepo.UpdateEmergencyIncident(ctx, incident); err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}
	
	// Free up the unit if one was assigned
	if incident.AssignedUnitID != "" {
		unit, err := s.postgresRepo.GetEmergencyUnit(ctx, incident.AssignedUnitID)
		if err == nil {
			unit.Status = "AVAILABLE"
			unit.CurrentIncidentID = ""
			_ = s.postgresRepo.UpdateEmergencyUnit(ctx, unit)
		}
	}
	
	// Clear dispatch cache
	s.redisRepo.DeleteEmergencyDispatch(ctx, incident.AssignedUnitID)
	
	return nil
}

// GetEmergencyMetrics provides emergency response metrics
func (s *EmergencyService) GetEmergencyMetrics(ctx context.Context, timeRange time.Duration) (*EmergencyMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	since := time.Now().Add(-timeRange)
	
	incidents, err := s.postgresRepo.GetEmergencyIncidentsByTimeRange(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get incidents: %w", err)
	}
	
	metrics := &EmergencyMetrics{
		TimeRange:     timeRange,
		TotalIncidents: len(incidents),
		ResolvedCount:  0,
		AvgResponseTime: 0,
		ByType:         make(map[string]int),
		ByPriority:     make(map[PriorityLevel]int),
	}
	
	var totalResponseTime time.Duration
	
	for _, inc := range incidents {
		if inc.Status == "RESOLVED" {
			metrics.ResolvedCount++
			if inc.DispatchedAt != nil && inc.ResolvedAt != nil {
				totalResponseTime += inc.ResolvedAt.Sub(*inc.DispatchedAt)
			}
		}
		
		metrics.ByType[inc.Type]++
		metrics.ByPriority[inc.Priority]++
	}
	
	if metrics.ResolvedCount > 0 {
		metrics.AvgResponseTime = totalResponseTime / time.Duration(metrics.ResolvedCount)
	}
	
	return metrics, nil
}

// haversineDistance calculates distance between two coordinates in kilometers
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0 // km
	
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180
	
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return earthRadius * c
}

// EmergencyMetrics contains emergency response statistics
type EmergencyMetrics struct {
	TimeRange        time.Duration
	TotalIncidents   int
	ResolvedCount    int
	AvgResponseTime  time.Duration
	ByType           map[string]int
	ByPriority       map[PriorityLevel]int
}
