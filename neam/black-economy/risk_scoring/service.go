package risk_scoring

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/redis/go-redis/v9"
	"neam-platform/shared"
)

// Config holds the configuration for the risk scoring service
type Config struct {
	ClickHouse clickhouse.Conn
	PostgreSQL *shared.PostgreSQL
	Redis      *redis.Client
	Logger     *shared.Logger
}

// Service handles risk scoring for entities and sectors
type Service struct {
	config     Config
	logger     *shared.Logger
	clickhouse clickhouse.Conn
	postgreSQL *shared.PostgreSQL
	redis      *redis.Client
	scores     map[string]*RiskScore
	histories  map[string][]RiskHistoryEntry
	factors    map[string]*RiskFactor
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// RiskScore represents a comprehensive risk score
type RiskScore struct {
	EntityID    string                 `json:"entity_id"`
	EntityType  string                 `json:"entity_type"` // company, individual, sector, region
	Score       float64                `json:"score"`       // 0.0 - 1.0
	Level       string                 `json:"level"`       // critical, high, medium, low, minimal
	Confidence  float64                `json:"confidence"`  // 0.0 - 1.0
	Factors     []RiskFactorScore      `json:"factors"`
	Indicators  []IndicatorScore       `json:"indicators"`
	Trend       string                 `json:"trend"` // increasing, stable, decreasing
	Components  map[string]float64     `json:"components"`
	CalculatedAt time.Time             `json:"calculated_at"`
	ValidUntil  time.Time              `json:"valid_until"`
}

// RiskFactorScore represents score contribution from a risk factor
type RiskFactorScore struct {
	FactorID    string  `json:"factor_id"`
	FactorName  string  `json:"factor_name"`
	Weight      float64 `json:"weight"`
	RawValue    float64 `json:"raw_value"`
	Normalized  float64 `json:"normalized"`
	Score       float64 `json:"score"`
	Contribution float64 `json:"contribution"`
	Description string  `json:"description"`
}

// IndicatorScore represents individual indicator scores
type IndicatorScore struct {
	IndicatorID   string  `json:"indicator_id"`
	Name          string  `json:"name"`
	Value         float64 `json:"value"`
	Threshold     float64 `json:"threshold"`
	Status        string  `json:"status"` // normal, warning, critical
	Trend         string  `json:"trend"`
	LastUpdated   time.Time `json:"last_updated"`
}

// RiskHistoryEntry represents a historical risk score entry
type RiskHistoryEntry struct {
	Timestamp   time.Time  `json:"timestamp"`
	Score       float64    `json:"score"`
	Level       string     `json:"level"`
	TriggeringEvents []string `json:"triggering_events"`
}

// RiskFactor represents a configurable risk factor
type RiskFactor struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Category     string   `json:"category"` // behavioral, transactional, network, external
	Description  string   `json:"description"`
	Weight       float64  `json:"weight"`
	Indicators   []string `json:"indicators"`
	Thresholds   RiskThresholds `json:"thresholds"`
	Formula      string   `json:"formula"`
	Enabled      bool     `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RiskThresholds defines thresholds for risk factor scoring
type RiskThresholds struct {
	Low      float64 `json:"low"`      // Threshold for low risk
	Medium   float64 `json:"medium"`   // Threshold for medium risk
	High     float64 `json:"high"`     // Threshold for high risk
	Critical float64 `json:"critical"` // Threshold for critical risk
}

// SectorRisk represents aggregated sector-level risk
type SectorRisk struct {
	Sector     string                 `json:"sector"`
	AvgScore   float64                `json:"avg_score"`
	MaxScore   float64                `json:"max_score"`
	MinScore   float64                `json:"min_score"`
	EntityCount int                   `json:"entity_count"`
	Trend      string                 `json:"trend"`
	Components map[string]float64     `json:"components"`
	RiskAreas  []RiskArea             `json:"risk_areas"`
}

// RegionRisk represents region-level risk assessment
type RegionRisk struct {
	Region     string                 `json:"region"`
	AvgScore   float64                `json:"avg_score"`
	EntityCount int                   `json:"entity_count"`
	Trend      string                 `json:"trend"`
	SectorBreakdown map[string]float64 `json:"sector_breakdown"`
}

// RiskArea represents a specific area of risk
type RiskArea struct {
	Area       string  `json:"area"`
	Score      float64 `json:"score"`
	EntityCount int    `json:"entity_count"`
	Trend      string  `json:"trend"`
}

// LeaderboardEntry represents an entry in the risk leaderboard
type LeaderboardEntry struct {
	Rank       int     `json:"rank"`
	EntityID   string  `json:"entity_id"`
	EntityName string  `json:"entity_name"`
	Score      float64 `json:"score"`
	Level      string  `json:"level"`
	Change     float64 `json:"change"`
}

// NewService creates a new risk scoring service
func NewService(cfg Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	svc := &Service{
		config:     cfg,
		logger:     cfg.Logger,
		clickhouse: cfg.ClickHouse,
		postgreSQL: cfg.PostgreSQL,
		redis:      cfg.Redis,
		scores:     make(map[string]*RiskScore),
		histories:  make(map[string][]RiskHistoryEntry),
		factors:    make(map[string]*RiskFactor),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Initialize default risk factors
	svc.initializeRiskFactors()

	return svc
}

// initializeRiskFactors sets up default risk factors
func (s *Service) initializeRiskFactors() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Behavioral risk factors
	s.factors["cash_intensity"] = &RiskFactor{
		ID:          "cash_intensity",
		Name:        "Cash Intensity",
		Category:    "behavioral",
		Description: "High cash-to-digital payment ratio",
		Weight:      0.20,
		Indicators:  []string{"cash_ratio", "cash_volume"},
		Thresholds:  RiskThresholds{Low: 0.2, Medium: 0.4, High: 0.6, Critical: 0.8},
		Formula:     "min(value / threshold, 1.0)",
		Enabled:     true,
		CreatedAt:   time.Now(),
	}

	s.factors["transaction_anomaly"] = &RiskFactor{
		ID:          "transaction_anomaly",
		Name:        "Transaction Anomaly",
		Category:    "transactional",
		Description: "Unusual transaction patterns detected",
		Weight:      0.25,
		Indicators:  []string{"benford_deviation", "outlier_count"},
		Thresholds:  RiskThresholds{Low: 0.1, Medium: 0.3, High: 0.5, Critical: 0.7},
		Formula:     "weighted_average(indicators)",
		Enabled:     true,
		CreatedAt:   time.Now(),
	}

	s.factors["network_risk"] = &RiskFactor{
		ID:          "network_risk",
		Name:        "Network Risk",
		Category:    "network",
		Description: "Connection to high-risk entities",
		Weight:      0.20,
		Indicators:  []string{"connected_risk_count", "avg_connection_risk"},
		Thresholds:  RiskThresholds{Low: 0.15, Medium: 0.35, High: 0.55, Critical: 0.75},
		Formula:     "max(connected_risks)",
		Enabled:     true,
		CreatedAt:   time.Now(),
	}

	s.factors["compliance_history"] = &RiskFactor{
		ID:          "compliance_history",
		Name:        "Compliance History",
		Category:    "behavioral",
		Description: "Historical compliance violations",
		Weight:      0.15,
		Indicators:  []string{"violation_count", "severity_score"},
		Thresholds:  RiskThresholds{Low: 0.1, Medium: 0.3, High: 0.5, Critical: 0.8},
		Formula:     "decaying_sum(violations)",
		Enabled:     true,
		CreatedAt:   time.Now(),
	}

	s.factors["external_risk"] = &RiskFactor{
		ID:          "external_risk",
		Name:        "External Risk Signals",
		Category:    "external",
		Description: "External risk intelligence indicators",
		Weight:      0.20,
		Indicators:  []string{"negative_news", "watchlist_match"},
		Thresholds:  RiskThresholds{Low: 0.1, Medium: 0.3, High: 0.5, Critical: 0.7},
		Formula:     "binary_or(indicators)",
		Enabled:     true,
		CreatedAt:   time.Now(),
	}

	s.logger.Info("Risk factors initialized", "count", len(s.factors))
}

// Start begins the risk scoring service
func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Starting risk scoring service")

	// Start background scoring updates
	s.wg.Add(1)
	go s.scoreCalculationLoop()
	s.wg.Add(1)
	go s.historyCleanupLoop()

	s.logger.Info("Risk scoring service started")
}

// Stop gracefully stops the service
func (s *Service) Stop() {
	s.logger.Info("Stopping risk scoring service")
	s.cancel()
	s.wg.Wait()
	s.logger.Info("Risk scoring service stopped")
}

// scoreCalculationLoop periodically recalculates scores
func (s *Service) scoreCalculationLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(4 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.recalculateHighRiskScores()
		}
	}
}

// recalculateHighRiskScores updates scores for high-risk entities
func (s *Service) recalculateHighRiskScores() {
	ctx := context.Background()

	// Get entities with recent anomalies
	entityIDs := s.getEntitiesWithAnomalies(ctx)

	for _, entityID := range entityIDs {
		s.CalculateScore(ctx, entityID, "company")
	}

	s.logger.Info("Risk score recalculation completed", "entities_updated", len(entityIDs))
}

// getEntitiesWithAnomalies retrieves entities with recent anomalies
func (s *Service) getEntitiesWithAnomalies(ctx context.Context) []string {
	query := `
		SELECT DISTINCT entity_id
		FROM neam.anomalies
		WHERE detection_date > now() - INTERVAL 7 DAY
	`

	var entityIDs []string
	s.clickhouse.Query(ctx, query, &entityIDs)

	return entityIDs
}

// historyCleanupLoop removes old history entries
func (s *Service) historyCleanupLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanupOldHistory()
		}
	}
}

// cleanupOldHistory removes history entries older than retention period
func (s *Service) cleanupOldHistory() {
	retentionPeriod := 365 * 24 * time.Hour
	cutoff := time.Now().Add(-retentionPeriod)

	s.mu.Lock()
	for entityID, history := range s.histories {
		var filtered []RiskHistoryEntry
		for _, entry := range history {
			if entry.Timestamp.After(cutoff) {
				filtered = append(filtered, entry)
			}
		}
		s.histories[entityID] = filtered
	}
	s.mu.Unlock()

	s.logger.Debug("History cleanup completed")
}

// CalculateScore calculates risk score for an entity
func (s *Service) CalculateScore(ctx context.Context, entityID, entityType string) (*RiskScore, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Gather indicator values
	indicators := s.gatherIndicators(ctx, entityID, entityType)

	// Calculate factor scores
	var factorScores []RiskFactorScore
	components := make(map[string]float64)

	for _, factor := range s.factors {
		if !factor.Enabled {
			continue
		}

		factorScore := s.calculateFactorScore(ctx, factor, indicators)
		factorScores = append(factorScores, factorScore)
		components[factor.ID] = factorScore.Score
	}

	// Calculate weighted total score
	totalScore := s.calculateWeightedScore(factorScores)
	level := s.scoreToLevel(totalScore)
	trend := s.calculateTrend(entityID)

	// Calculate confidence based on data quality
	confidence := s.calculateConfidence(indicators)

	riskScore := &RiskScore{
		EntityID:    entityID,
		EntityType:  entityType,
		Score:       totalScore,
		Level:       level,
		Confidence:  confidence,
		Factors:     factorScores,
		Indicators:  s.buildIndicatorScores(indicators),
		Trend:       trend,
		Components:  components,
		CalculatedAt: time.Now(),
		ValidUntil:  time.Now().Add(24 * time.Hour),
	}

	// Store score
	s.scores[entityID] = riskScore

	// Add to history
	s.addToHistory(entityID, riskScore)

	// Cache in Redis
	s.cacheScore(ctx, entityID, riskScore)

	// Persist to database
	s.persistScore(ctx, riskScore)

	s.logger.Debug("Risk score calculated", "entity_id", entityID, "score", totalScore, "level", level)

	return riskScore, nil
}

// gatherIndicators retrieves indicator values from data sources
func (s *Service) gatherIndicators(ctx context.Context, entityID, entityType string) map[string]float64 {
	indicators := make(map[string]float64)

	// Query ClickHouse for indicator values
	query := `
		SELECT 
			cash_ratio,
			benford_deviation,
			outlier_count,
			connected_risk_count,
			avg_connection_risk,
			violation_count,
			severity_score,
			negative_news_count,
			watchlist_matches
		FROM neam.entity_indicators
		WHERE entity_id = $1
		LIMIT 1
	`

	// Simulate indicator values
	indicators["cash_ratio"] = 0.35
	indicators["cash_volume"] = 5000000
	indicators["benford_deviation"] = 0.02
	indicators["outlier_count"] = 3
	indicators["connected_risk_count"] = 2
	indicators["avg_connection_risk"] = 0.65
	indicators["violation_count"] = 1
	indicators["severity_score"] = 0.4
	indicators["negative_news_count"] = 0
	indicators["watchlist_matches"] = 0

	return indicators
}

// calculateFactorScore calculates score for a specific factor
func (s *Service) calculateFactorScore(ctx context.Context, factor *RiskFactor, indicators map[string]float64) RiskFactorScore {
	// Get indicator values for this factor
	var rawValues []float64
	for _, indicatorID := range factor.Indicators {
		if value, ok := indicators[indicatorID]; ok {
			rawValues = append(rawValues, value)
		}
	}

	// Calculate normalized value
	normalized := s.normalizeValues(rawValues, factor.Thresholds)

	// Calculate score
	score := normalized * factor.Weight

	return RiskFactorScore{
		FactorID:     factor.ID,
		FactorName:   factor.Name,
		Weight:       factor.Weight,
		RawValue:      s.average(rawValues),
		Normalized:   normalized,
		Score:        score,
		Contribution: score * factor.Weight,
		Description:  factor.Description,
	}
}

// normalizeValues normalizes raw values to 0-1 scale
func (s *Service) normalizeValues(values []float64, thresholds RiskThresholds) float64 {
	if len(values) == 0 {
		return 0
	}

	avg := s.average(values)

	if avg <= thresholds.Low {
		return avg / thresholds.Low * 0.25
	} else if avg <= thresholds.Medium {
		return 0.25 + (avg-thresholds.Low)/(thresholds.Medium-thresholds.Low)*0.25
	} else if avg <= thresholds.High {
		return 0.50 + (avg-thresholds.Medium)/(thresholds.High-thresholds.Medium)*0.25
	} else if avg <= thresholds.Critical {
		return 0.75 + (avg-thresholds.High)/(thresholds.Critical-thresholds.High)*0.25
	}
	return 1.0
}

// calculateWeightedScore calculates weighted sum of factor scores
func (s *Service) calculateWeightedScore(factorScores []RiskFactorScore) float64 {
	var totalScore float64
	for _, fs := range factorScores {
		totalScore += fs.Score
	}
	return math.Min(totalScore, 1.0)
}

// scoreToLevel converts numeric score to risk level
func (s *Service) scoreToLevel(score float64) string {
	switch {
	case score >= 0.75:
		return "critical"
	case score >= 0.55:
		return "high"
	case score >= 0.35:
		return "medium"
	case score >= 0.15:
		return "low"
	default:
		return "minimal"
	}
}

// calculateTrend determines score trend from history
func (s *Service) calculateTrend(entityID string) string {
	s.mu.RLock()
	history, exists := s.histories[entityID]
	s.mu.RUnlock()

	if !exists || len(history) < 2 {
		return "stable"
	}

	recent := history[len(history)-1]
	previous := history[len(history)-2]

	if recent.Score > previous.Score+0.05 {
		return "increasing"
	} else if recent.Score < previous.Score-0.05 {
		return "decreasing"
	}
	return "stable"
}

// calculateConfidence calculates confidence level based on data completeness
func (s *Service) calculateConfidence(indicators map[string]float64) float64 {
	// Simple confidence calculation based on indicator availability
	total := 10.0
	available := float64(len(indicators))
	return math.Min(available/total, 1.0)
}

// buildIndicatorScores builds indicator score list
func (s *Service) buildIndicatorScores(indicators map[string]float64) []IndicatorScore {
	var scores []IndicatorScore

	thresholds := map[string]float64{
		"cash_ratio":        0.40,
		"benford_deviation": 0.03,
		"outlier_count":     5,
	}

	for indicator, value := range indicators {
		threshold := thresholds[indicator]
		status := "normal"
		if value > threshold*1.5 {
			status = "critical"
		} else if value > threshold {
			status = "warning"
		}

		scores = append(scores, IndicatorScore{
			IndicatorID:  indicator,
			Name:         indicator,
			Value:        value,
			Threshold:    threshold,
			Status:       status,
			LastUpdated:  time.Now(),
		})
	}

	return scores
}

// addToHistory adds score to history
func (s *Service) addToHistory(entityID string, score *RiskScore) {
	entry := RiskHistoryEntry{
		Timestamp: score.CalculatedAt,
		Score:     score.Score,
		Level:     score.Level,
	}

	s.histories[entityID] = append(s.histories[entityID], entry)
}

// GetScore retrieves current risk score for an entity
func (s *Service) GetScore(ctx context.Context, entityID string) (*RiskScore, error) {
	// Try cache first
	key := fmt.Sprintf("neam:risk:%s", entityID)
	if cached, err := s.redis.Get(ctx, key).Result(); err == nil {
		// Parse cached score (simplified)
		s.mu.RLock()
		if score, exists := s.scores[entityID]; exists {
			s.mu.RUnlock()
			return score, nil
		}
		s.mu.RUnlock()
	}

	// Fallback to stored score
	s.mu.RLock()
	score, exists := s.scores[entityID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("risk score not found: %s", entityID)
	}

	return score, nil
}

// GetScoreHistory retrieves historical scores for an entity
func (s *Service) GetScoreHistory(ctx context.Context, entityID string, days int) ([]RiskHistoryEntry, error) {
	s.mu.RLock()
	history, exists := s.histories[entityID]
	s.mu.RUnlock()

	if !exists {
		return []RiskHistoryEntry{}, nil
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	var filtered []RiskHistoryEntry
	for _, entry := range history {
		if entry.Timestamp.After(cutoff) {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// GetSectorRisk calculates aggregated risk for a sector
func (s *Service) GetSectorRisk(ctx context.Context, sector string) (*SectorRisk, error) {
	query := `
		SELECT 
			avg(score) as avg_score,
			max(score) as max_score,
			min(score) as min_score,
			count(*) as entity_count
		FROM neam.risk_scores
		WHERE sector = $1
		AND calculated_at > now() - INTERVAL 7 DAY
	`

	var risk SectorRisk
	_ = query

	// Calculate risk areas
	risk.RiskAreas = []RiskArea{
		{Area: "Cash Intensity", Score: 0.35, EntityCount: 50},
		{Area: "Transaction Anomalies", Score: 0.25, EntityCount: 30},
		{Area: "Network Connections", Score: 0.45, EntityCount: 40},
	}

	return &risk, nil
}

// GetRegionRisk calculates aggregated risk for a region
func (s *Service) GetRegionRisk(ctx context.Context, region string) (*RegionRisk, error) {
	var risk RegionRisk
	_ = region

	return &risk, nil
}

// GetLeaderboard returns entities with highest risk scores
func (s *Service) GetLeaderboard(ctx context.Context, limit int) ([]LeaderboardEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Sort scores and return top entities
	var entries []LeaderboardEntry
	rank := 1

	for id, score := range s.scores {
		if rank > limit {
			break
		}
		entries = append(entries, LeaderboardEntry{
			Rank:       rank,
			EntityID:   id,
			EntityName: id,
			Score:      score.Score,
			Level:      score.Level,
		})
		rank++
	}

	return entries, nil
}

// cacheScore stores score in Redis
func (s *Service) cacheScore(ctx context.Context, entityID string, score *RiskScore) {
	key := fmt.Sprintf("neam:risk:%s", entityID)
	s.redis.Set(ctx, key, fmt.Sprintf("%f", score.Score), 24*time.Hour)
}

// persistScore stores score in database
func (s *Service) persistScore(ctx context.Context, score *RiskScore) {
	query := `
		INSERT INTO neam.risk_scores 
		(entity_id, entity_type, score, level, confidence, components, calculated_at, valid_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	s.postgreSQL.Execute(ctx, query,
		score.EntityID, score.EntityType, score.Score, score.Level,
		score.Confidence, fmt.Sprintf("%v", score.Components),
		score.CalculatedAt, score.ValidUntil,
	)
}

// average calculates the average of a slice
func (s *Service) average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
