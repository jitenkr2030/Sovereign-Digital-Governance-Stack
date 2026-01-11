package signalproc

import (
	"context"
	"sync"
	"time"

	"github.com/neamplatform/sensing/energy/config"
	"go.uber.org/zap"
)

// PointState holds the state for a single telemetry point
type PointState struct {
	LastValue         float64
	LastTimestamp     time.Time
	LastReportedValue float64
	LastReportedTime  time.Time
	ValueHistory      []float64
	SpikeCount        int
	AnomalyDetected   bool
	mu                sync.RWMutex
}

// Processor handles signal filtering and anomaly detection
type Processor struct {
	cfg        config.FilteringConfig
	logger     *zap.Logger
	pointState map[string]*PointState
	mu         sync.RWMutex
	stats      ProcessorStats
}

// ProcessorStats holds processing statistics
type ProcessorStats struct {
	PointsProcessed    int64
	PointsFiltered     int64
	SpikesDetected     int64
	AnomaliesDetected  int64
	QualityFiltered    int64
	LastProcessTime    time.Time
}

// NewProcessor creates a new signal processor
func NewProcessor(cfg config.FilteringConfig, logger *zap.Logger) *Processor {
	return &Processor{
		cfg:        cfg,
		logger:     logger,
		pointState: make(map[string]*PointState),
	}
}

// Process filters and processes incoming telemetry points
func (p *Processor) Process(input <-chan TelemetryPoint) <-chan TelemetryPoint {
	output := make(chan TelemetryPoint, 1000)

	go func() {
		defer close(output)

		for point := range input {
			processed := p.processPoint(point)

			if processed != nil {
				select {
				case output <- *processed:
				case <-context.Background().Done():
					return
				}
			}
		}
	}()

	return output
}

// processPoint applies all filters to a single point
func (p *Processor) processPoint(point TelemetryPoint) *TelemetryPoint {
	p.mu.Lock()
	state, exists := p.pointState[point.PointID]
	if !exists {
		state = &PointState{
			ValueHistory: make([]float64, 0, p.cfg.SpikeWindowSize),
		}
		p.pointState[point.PointID] = state
	}
	state.mu.Lock()
	p.mu.Unlock()

	// Quality filter
	if !p.checkQuality(point.Quality) {
		state.mu.Unlock()
		p.incrementStat("quality_filtered")
		return nil
	}

	// Range validation
	if !p.checkRange(point.Value) {
		state.mu.Unlock()
		p.logger.Warn("Value out of range",
			zap.String("point_id", point.PointID),
			zap.Float64("value", point.Value))
		p.incrementStat("filtered")
		return nil
	}

	// First point - always report
	if !exists {
		state.LastValue = point.Value
		state.LastTimestamp = point.Timestamp
		state.LastReportedValue = point.Value
		state.LastReportedTime = point.Timestamp
		state.mu.Unlock()

		p.incrementStat("processed")
		return &point
	}

	// Deadband filter
	if p.applyDeadband(point, state) {
		state.mu.Unlock()
		p.incrementStat("filtered")
		return nil
	}

	// Spike detection
	spikeDetected := p.detectSpike(point, state)
	if spikeDetected {
		state.SpikeCount++
		if state.SpikeCount >= 3 {
			state.AnomalyDetected = true
			p.incrementStat("anomalies_detected")
			p.logger.Warn("Anomaly detected",
				zap.String("point_id", point.PointID),
				zap.Float64("value", point.Value),
				zap.Float64("last_value", state.LastValue))
		}
		p.incrementStat("spikes_detected")
	} else {
		state.SpikeCount = 0
		state.AnomalyDetected = false
	}

	// Rate of change check
	if !p.checkRateOfChange(point, state) {
		state.mu.Unlock()
		p.incrementStat("filtered")
		return nil
	}

	// Update state
	state.LastValue = point.Value
	state.LastTimestamp = point.Timestamp
	state.LastReportedValue = point.Value
	state.LastReportedTime = point.Timestamp

	// Add to history
	state.ValueHistory = append(state.ValueHistory, point.Value)
	if len(state.ValueHistory) > p.cfg.SpikeWindowSize {
		state.ValueHistory = state.ValueHistory[1:]
	}

	state.mu.Unlock()
	p.incrementStat("processed")

	return &point
}

// applyDeadband checks if change is within deadband threshold
func (p *Processor) applyDeadband(point TelemetryPoint, state *PointState) bool {
	if state.LastReportedValue == 0 {
		return false
	}

	// Calculate percentage change
	change := (point.Value - state.LastReportedValue) / state.LastReportedValue * 100
	change = abs(change)

	// Check minimum reporting interval
	timeSinceLast := point.Timestamp.Sub(state.LastReportedTime)
	if timeSinceLast < time.Duration(p.cfg.MinReportingInterval) {
		return true
	}

	// Apply deadband filter
	return change <= p.cfg.DeadbandPercent
}

// detectSpike identifies sudden large changes
func (p *Processor) detectSpike(point TelemetryPoint, state *PointState) bool {
	if len(state.ValueHistory) < 2 {
		return false
	}

	// Calculate average of recent values
	sum := 0.0
	for _, v := range state.ValueHistory {
		sum += v
	}
	avg := sum / float64(len(state.ValueHistory))

	// Calculate deviation
	deviation := point.Value - avg
	if avg != 0 {
		deviation = deviation / avg * 100
	}

	return abs(deviation) >= p.cfg.SpikeThreshold
}

// checkRateOfChange validates rate of change limit
func (p *Processor) checkRateOfChange(point TelemetryPoint, state *PointState) bool {
	if state.LastTimestamp.IsZero() {
		return true
	}

	timeDelta := point.Timestamp.Sub(state.LastTimestamp)
	if timeDelta == 0 {
		return true
	}

	valueDelta := point.Value - state.LastValue
	if state.LastValue != 0 {
		valueDelta = valueDelta / state.LastValue * 100
	}

	// Calculate rate of change per second
	rateOfChange := valueDelta / timeDelta.Seconds()

	return rateOfChange <= p.cfg.RateOfChangeLimit
}

// checkQuality verifies data quality
func (p *Processor) checkQuality(quality string) bool {
	for _, valid := range p.cfg.QualityFilter {
		if quality == valid {
			return true
		}
	}
	return len(p.cfg.QualityFilter) == 0
}

// checkRange validates value is within acceptable range
func (p *Processor) checkRange(value float64) bool {
	return value >= p.cfg.MinValue && value <= p.cfg.MaxValue
}

// incrementStat atomically increments a statistic
func (p *Processor) incrementStat(stat string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch stat {
	case "processed":
		p.stats.PointsProcessed++
	case "filtered":
		p.stats.PointsFiltered++
	case "spikes_detected":
		p.stats.SpikesDetected++
	case "anomalies_detected":
		p.stats.AnomaliesDetected++
	case "quality_filtered":
		p.stats.QualityFiltered++
	}
	p.stats.LastProcessTime = time.Now()
}

// Stats returns current processing statistics
func (p *Processor) Stats() ProcessorStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats
}

// ResetStats clears processing statistics
func (p *Processor) ResetStats() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stats = ProcessorStats{}
}

// GetPointState returns the state for a specific point
func (p *Processor) GetPointState(pointID string) *PointState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pointState[pointID]
}

// ClearPointState removes state for a specific point
func (p *Processor) ClearPointState(pointID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.pointState, pointID)
}

// ClearAllState removes all point states
func (p *Processor) ClearAllState() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pointState = make(map[string]*PointState)
}

// abs returns absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
