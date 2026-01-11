package service

import (
	"context"
	"fmt"
	"smart-cities/internal/repository"
	"sync"
	"time"
)

// WasteService handles smart waste management logic
type WasteService struct {
	postgresRepo *repository.PostgresRepository
	redisRepo    *repository.RedisRepository
	mu           sync.RWMutex
}

// NewWasteService creates a new waste service instance
func NewWasteService(pg *repository.PostgresRepository, rd *repository.RedisRepository) *WasteService {
	return &WasteService{
		postgresRepo: pg,
		redisRepo:    rd,
	}
}

// CalculateFillLevel calculates the current fill level of a waste bin
func (s *WasteService) CalculateFillLevel(bin *WasteBin) float64 {
	if bin.Capacity <= 0 {
		return 0.0
	}
	return float64(bin.CurrentFill) / float64(bin.Capacity) * 100.0
}

// PredictCollectionTime predicts when a bin will need collection
func (s *WasteService) PredictCollectionTime(bin *WasteBin) time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	fillLevel := s.CalculateFillLevel(bin)
	
	// Get historical collection data
	history, err := s.redisRepo.GetCollectionHistory(ctx, bin.ID)
	if err != nil {
		history = nil
	}
	
	// Calculate average fill rate from history
	var avgFillRate float64
	if len(history) > 0 {
		var totalRate float64
		for _, collection := range history {
			if collection.Interval > 0 {
				totalRate += float64(collection.FillAmount) / float64(collection.Interval)
			}
		}
		avgFillRate = totalRate / float64(len(history))
	}
	
	// Default fill rate if no history (percent per hour)
	if avgFillRate == 0 {
		avgFillRate = 5.0 // 5% per hour average
	}
	
	// Adjust for day of week (weekends may have different patterns)
	current := time.Now()
	var dayMultiplier float64
	switch current.Weekday() {
	case time.Saturday, time.Sunday:
		dayMultiplier = 1.2 // More waste on weekends
	default:
		dayMultiplier = 1.0
	}
	
	remainingFill := 100.0 - fillLevel
	hoursToFull := remainingFill / (avgFillRate * dayMultiplier)
	
	return time.Now().Add(time.Duration(hoursToFull) * time.Hour)
}

// GetOptimalCollectionRoute calculates the most efficient route for waste collection
func (s *WasteService) GetOptimalCollectionRoute(ctx context.Context, vehicleID string, maxBins int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Get bins that need collection (fill level > threshold)
	thresholdBins, err := s.postgresRepo.GetWasteBinsByFillLevel(ctx, 70.0)
	if err != nil {
		return nil, fmt.Errorf("failed to get bins: %w", err)
	}
	
	// Get vehicle location
	vehicle, err := s.postgresRepo.GetCollectionVehicle(ctx, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle: %w", err)
	}
	
	// Calculate distance from vehicle to each bin
	type binDist struct {
		binID   string
		distance float64
		fillLevel float64
	}
	
	var binDistances []binDist
	for _, bin := range thresholdBins {
		dist := haversineDistance(
			vehicle.Latitude, vehicle.Longitude,
			bin.Latitude, bin.Longitude,
		)
		binDistances = append(binDistances, binDist{
			binID:     bin.ID,
			distance:  dist,
			fillLevel: s.CalculateFillLevel(&bin),
		})
	}
	
	// Sort by fill level (descending) then by distance (ascending)
	// This prioritizes full bins but considers proximity
	for i := 0; i < len(binDistances)-1; i++ {
		for j := i + 1; j < len(binDistances); j++ {
			// Higher fill level gets priority
			if binDistances[j].fillLevel > binDistances[i].fillLevel {
				binDistances[i], binDistances[j] = binDistances[j], binDistances[i]
			}
		}
	}
	
	// Build route using nearest neighbor heuristic
	var route []string
	visited := make(map[string]bool)
	currentLat, currentLon := vehicle.Latitude, vehicle.Longitude
	
	for i := 0; i < len(binDistances) && len(route) < maxBins; i++ {
		bin := binDistances[i]
		if visited[bin.binID] {
			continue
		}
		
		route = append(route, bin.binID)
		visited[bin.binID] = true
		currentLat, currentLon = binDistances[i].bin.Latitude, binDistances[i].bin.Longitude
	}
	
	return route, nil
}

// CollectWaste records a waste collection event
func (s *WasteService) CollectWaste(ctx context.Context, binID string, vehicleID string, weightCollected int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	bin, err := s.postgresRepo.GetWasteBin(ctx, binID)
	if err != nil {
		return fmt.Errorf("failed to get bin: %w", err)
	}
	
	// Record collection
	collection := &WasteCollection{
		BinID:       binID,
		VehicleID:   vehicleID,
		Weight:      weightCollected,
		CollectedAt: time.Now(),
	}
	
	if err := s.postgresRepo.RecordWasteCollection(ctx, collection); err != nil {
		return fmt.Errorf("failed to record collection: %w", err)
	}
	
	// Update bin status
	bin.LastCollection = time.Now()
	bin.CurrentFill = 0
	bin.LastUpdated = time.Now()
	
	if err := s.postgresRepo.UpdateWasteBin(ctx, bin); err != nil {
		return fmt.Errorf("failed to update bin: %w", err)
	}
	
	// Update cache
	s.redisRepo.SetBinFillLevel(ctx, binID, 0, time.Hour)
	
	return nil
}

// UpdateBinFillLevel updates the fill level of a bin from sensor data
func (s *WasteService) UpdateBinFillLevel(ctx context.Context, binID string, fillLevel int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	bin, err := s.postgresRepo.GetWasteBin(ctx, binID)
	if err != nil {
		return fmt.Errorf("failed to get bin: %w", err)
	}
	
	bin.CurrentFill = fillLevel
	bin.LastUpdated = time.Now()
	bin.SensorBattery = bin.SensorBattery - 1 // Battery drains slightly with each update
	
	if bin.SensorBattery < 10 {
		bin.AlertFlags = append(bin.AlertFlags, "LOW_BATTERY")
	}
	
	if err := s.postgresRepo.UpdateWasteBin(ctx, bin); err != nil {
		return fmt.Errorf("failed to update bin: %w", err)
	}
	
	// Update cache
	s.redisRepo.SetBinFillLevel(ctx, binID, fillLevel, time.Hour)
	
	// Check if collection is needed
	fillPercentage := s.CalculateFillLevel(bin)
	if fillPercentage > 90 {
		s.triggerCollectionAlert(ctx, bin)
	}
	
	return nil
}

// triggerCollectionAlert sends alerts for bins needing immediate collection
func (s *WasteService) triggerCollectionAlert(ctx context.Context, bin *WasteBin) {
	// This would integrate with notification system
	// For now, we just log the alert
	_ = bin
	_ = ctx
	// In production, this would send SMS/email/push notifications
}

// GetWasteAnalytics provides comprehensive waste management analytics
func (s *WasteService) GetWasteAnalytics(ctx context.Context, timeRange time.Duration) (*WasteAnalytics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	since := time.Now().Add(-timeRange)
	
	collections, err := s.postgresRepo.GetWasteCollectionsByTimeRange(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get collections: %w", err)
	}
	
	bins, err := s.postgresRepo.GetAllWasteBins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bins: %w", err)
	}
	
	analytics := &WasteAnalytics{
		TimeRange:        timeRange,
		TotalCollections: len(collections),
		TotalWeight:      0,
		AvgCollectionWeight: 0,
		AvgFillLevel:     0,
		BinStatus:        make(map[string]int),
		ByWasteType:      make(map[string]int),
	}
	
	var totalWeight int
	var totalFillLevel float64
	
	for _, collection := range collections {
		totalWeight += collection.Weight
		analytics.ByWasteType[collection.WasteType] += collection.Weight
	}
	
	for _, bin := range bins {
		fillLevel := s.CalculateFillLevel(&bin)
		totalFillLevel += fillLevel
		
		status := "NORMAL"
		if fillLevel > 90 {
			status = "CRITICAL"
		} else if fillLevel > 70 {
			status = "WARNING"
		}
		analytics.BinStatus[status]++
	}
	
	analytics.TotalWeight = totalWeight
	if len(collections) > 0 {
		analytics.AvgCollectionWeight = totalWeight / len(collections)
	}
	if len(bins) > 0 {
		analytics.AvgFillLevel = totalFillLevel / float64(len(bins))
	}
	
	return analytics, nil
}

// OptimizeCollectionSchedule optimizes waste collection scheduling
func (s *WasteService) OptimizeCollectionSchedule(ctx context.Context, date time.Time) (map[string][]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	bins, err := s.postgresRepo.GetAllWasteBins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bins: %w", err)
	}
	
	// Get vehicles
	vehicles, err := s.postgresRepo.GetCollectionVehicles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicles: %w", err)
	}
	
	// Group bins by zone
	binsByZone := make(map[string][]WasteBin)
	for _, bin := range bins {
		if s.CalculateFillLevel(&bin) > 60 {
			binsByZone[bin.Zone] = append(binsByZone[bin.Zone], bin)
		}
	}
	
	// Assign bins to vehicles
	schedule := make(map[string][]string)
	vehicleIndex := 0
	
	for zone, zoneBins := range binsByZone {
		vehicleID := vehicles[vehicleIndex%len(vehicles)].ID
		
		for _, bin := range zoneBins {
			schedule[vehicleID] = append(schedule[vehicleID], bin.ID)
		}
		
		vehicleIndex++
	}
	
	return schedule, nil
}

// WasteAnalytics contains waste management statistics
type WasteAnalytics struct {
	TimeRange          time.Duration
	TotalCollections   int
	TotalWeight        int
	AvgCollectionWeight int
	AvgFillLevel       float64
	BinStatus          map[string]int
	ByWasteType        map[string]int
}
