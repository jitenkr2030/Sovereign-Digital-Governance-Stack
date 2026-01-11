package service

import (
	"context"
	"fmt"
	"math"
	"smart-cities/internal/repository"
	"sync"
	"time"
)

// TrafficFlowDirection represents the direction of traffic flow
type TrafficFlowDirection string

const (
	NorthSouth TrafficFlowDirection = "NS"
	EastWest   TrafficFlowDirection = "EW"
	Both       TrafficFlowDirection = "BOTH"
)

// TrafficService handles traffic management logic
type TrafficService struct {
	postgresRepo *repository.PostgresRepository
	redisRepo    *repository.RedisRepository
	mu           sync.RWMutex
}

// NewTrafficService creates a new traffic service instance
func NewTrafficService(pg *repository.PostgresRepository, rd *repository.RedisRepository) *TrafficService {
	return &TrafficService{
		postgresRepo: pg,
		redisRepo:    rd,
	}
}

// CalculateCongestionLevel calculates congestion based on vehicle count and capacity
func (s *TrafficService) CalculateCongestionLevel(vehicleCount, capacity int) float64 {
	if capacity <= 0 {
		return 0.0
	}
	
	ratio := float64(vehicleCount) / float64(capacity)
	if ratio <= 0.5 {
		return ratio * 0.3 // Low congestion: 0-15%
	} else if ratio <= 0.8 {
		return 0.15 + (ratio-0.5)*0.5 // Medium congestion: 15-30%
	} else if ratio <= 1.0 {
		return 0.30 + (ratio-0.8)*1.0 // High congestion: 30-50%
	}
	return math.Min(1.0, 0.50+(ratio-1.0)*0.25) // Severe: 50-75%
}

// GetOptimalGreenTime calculates optimal green light duration based on traffic density
func (s *TrafficService) GetOptimalGreenTime(direction TrafficFlowDirection, vehicleCount int, baseTime int) int {
	congestion := s.CalculateCongestionLevel(vehicleCount, 100) // Assuming 100 vehicles per lane as baseline
	
	// Adjust green time based on congestion
	adjustment := int(congestion * float64(baseTime))
	maxTime := baseTime * 3 // Cap at 3x base time
	
	optimalTime := baseTime + adjustment
	if optimalTime > maxTime {
		optimalTime = maxTime
	}
	
	return optimalTime
}

// CalculateTravelTimeEstimate estimates travel time based on current traffic conditions
func (s *TrafficService) CalculateTravelTimeEstimate(ctx context.Context, startNodeID, endNodeID string) (time.Duration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Get traffic data for the route
	routeNodes, err := s.postgresRepo.GetTrafficNodesByRoute(ctx, startNodeID, endNodeID)
 {
		return 	if err != nil0, fmt.Errorf("failed to get route nodes: %w", err)
	}
	
	var totalDelay time.Duration
	baseTimePerNode := 2 * time.Minute // Base travel time per node
	
	for _, node := range routeNodes {
		congestion := s.CalculateCongestionLevel(node.CurrentVehicleCount, node.Capacity)
		delay := time.Duration(congestion * 10 * float64(time.Minute)) // Max 10 min delay per node
		totalDelay += delay
	}
	
	return baseTimePerNode * time.Duration(len(routeNodes)) + totalDelay, nil
}

// PredictTrafficPattern predicts traffic patterns based on historical data
func (s *TrafficService) PredictTrafficPattern(ctx context.Context, nodeID string, hour int) (predictedCount int, confidence float64, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Get historical data for this hour
	history, err := s.redisRepo.GetTrafficHistory(ctx, nodeID, hour)
	if err != nil {
		// Fall back to postgres if Redis doesn't have data
		history, err = s.postgresRepo.GetTrafficHistory(ctx, nodeID, hour)
		if err != nil {
			return 0, 0.0, fmt.Errorf("failed to get traffic history: %w", err)
		}
	}
	
	if len(history) == 0 {
		// Default prediction based on time of day
		return s.getDefaultPrediction(hour), 0.5, nil
	}
	
	// Calculate weighted average (recent data has more weight)
	var total, weight float64
	for i, count := range history {
		recentWeight := float64(len(history) - i)
		total += float64(count) * recentWeight
		weight += recentWeight
	}
	
	predictedCount = int(total / weight)
	
	// Confidence based on data consistency
	if len(history) >= 7 {
		confidence = 0.85
	} else if len(history) >= 3 {
		confidence = 0.7
	} else {
		confidence = 0.5
	}
	
	return predictedCount, confidence, nil
}

// getDefaultPrediction provides default traffic predictions based on hour
func (s *TrafficService) getDefaultPrediction(hour int) int {
	switch {
	case hour >= 7 && hour <= 9:
		return 80 // Morning rush hour
	case hour >= 17 && hour <= 19:
		return 85 // Evening rush hour
	case hour >= 10 && hour <= 16:
		return 40 // Midday
	case hour >= 20 && hour <= 22:
		return 30 // Evening
	default:
		return 15 // Night/early morning
	}
}

// SynchronizeTrafficLights synchronizes traffic lights along a route for smooth flow
func (s *TrafficService) SynchronizeTrafficLights(ctx context.Context, routeID string, offset time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	routeNodes, err := s.postgresRepo.GetTrafficNodesByRoute(ctx, "", routeID)
	if err != nil {
		return fmt.Errorf("failed to get route nodes: %w", err)
	}
	
	for i, node := range routeNodes {
		node.GreenTime = s.GetOptimalGreenTime(Both, node.CurrentVehicleCount, 30)
		node.LastUpdated = time.Now()
		
		if err := s.postgresRepo.UpdateTrafficNode(ctx, &node); err != nil {
			return fmt.Errorf("failed to update node %s: %w", node.ID, err)
		}
		
		// Cache updated node
		s.redisRepo.SetTrafficNode(ctx, &node, 5*time.Minute)
	}
	
	return nil
}

// HandleTrafficIncident processes a traffic incident and adjusts signals accordingly
func (s *TrafficService) HandleTrafficIncident(ctx context.Context, incidentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	incident, err := s.postgresRepo.GetTrafficIncident(ctx, incidentID)
	if err != nil {
		return fmt.Errorf("failed to get incident: %w", err)
	}
	
	// Update status
	incident.Status = "ACTIVE"
	incident.ResponseStarted = time.Now()
	if err := s.postgresRepo.UpdateTrafficIncident(ctx, incident); err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}
	
	// Adjust traffic signals in affected area
	affectedNodes, err := s.postgresRepo.GetTrafficNodesInRadius(ctx, incident.Latitude, incident.Longitude, 500)
	if err != nil {
		return fmt.Errorf("failed to get affected nodes: %w", err)
	}
	
	for _, node := range affectedNodes {
		// Extend green time for alternate routes
		node.GreenTime = int(float64(node.GreenTime) * 1.5)
		node.LastUpdated = time.Now()
		if err := s.postgresRepo.UpdateTrafficNode(ctx, &node); err != nil {
			return fmt.Errorf("failed to update node %s: %w", node.ID, err)
		}
	}
	
	return nil
}

// GetTrafficAnalytics provides comprehensive traffic analytics
func (s *TrafficService) GetTrafficAnalytics(ctx context.Context, nodeID string, timeRange time.Duration) (*TrafficAnalytics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Get current data
	node, err := s.postgresRepo.GetTrafficNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	
	// Get historical data
	since := time.Now().Add(-timeRange)
	history, err := s.redisRepo.GetTrafficMetrics(ctx, nodeID, since)
	if err != nil {
		history = []int{}
	}
	
	// Calculate statistics
	var avgCount, maxCount int
	for _, count := range history {
		avgCount += count
		if count > maxCount {
			maxCount = count
		}
	}
	
	if len(history) > 0 {
		avgCount /= len(history)
	}
	
	analytics := &TrafficAnalytics{
		NodeID:            nodeID,
		CurrentVehicleCount: node.CurrentVehicleCount,
		CurrentCongestion: s.CalculateCongestionLevel(node.CurrentVehicleCount, node.Capacity),
		AverageCount:      avgCount,
		PeakCount:         maxCount,
		TimeRange:         timeRange,
		LastUpdated:       time.Now(),
	}
	
	return analytics, nil
}

// TrafficAnalytics contains traffic analysis data
type TrafficAnalytics struct {
	NodeID            string
	CurrentVehicleCount int
	CurrentCongestion float64
	AverageCount      int
	PeakCount         int
	TimeRange         time.Duration
	LastUpdated       time.Time
}
