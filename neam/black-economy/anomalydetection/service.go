package anomalydetection

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

// Config holds the configuration for the anomaly detection service
type Config struct {
	ClickHouse clickhouse.Conn
	Redis      *redis.Client
	Logger     *shared.Logger
}

// Service handles anomaly detection for black economy indicators
type Service struct {
	config     Config
	logger     *shared.Logger
	clickhouse clickhouse.Conn
	redis      *redis.Client
	anomalies  map[string]*Anomaly
	detectors  map[string]Detector
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"` // cash_anomaly, benford_violation, circular_transaction
	EntityID      string                 `json:"entity_id"`
	EntityType    string                 `json:"entity_type"` // company, individual
	Severity      string                 `json:"severity"` // critical, high, medium, low
	Score         float64                `json:"score"`
	Indicators    []Indicator            `json:"indicators"`
	DetectionDate time.Time             `json:"detection_date"`
	Status        string                 `json:"status"` // detected, investigating, resolved, dismissed
	Metadata      map[string]interface{} `json:"metadata"`
}

// Indicator represents a specific anomaly indicator
type Indicator struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Threshold   float64 `json:"threshold"`
	Deviation   float64 `json:"deviation"`
	Unit        string  `json:"unit"`
	Description string  `json:"description"`
}

// Detector defines the interface for anomaly detectors
type Detector interface {
	Detect(ctx context.Context, entityID string) ([]Anomaly, error)
	GetName() string
	GetType() string
}

// CashAnomalyDetector detects unusual cash-to-digital payment ratios
type CashAnomalyDetector struct {
	logger *shared.Logger
	redis  *redis.Client
}

// BenfordDetector analyzes transaction distributions for Benford's Law violations
type BenfordDetector struct {
	logger *shared.Logger
}

// CircularTransactionDetector identifies circular transaction patterns
type CircularTransactionDetector struct {
	logger *shared.Logger
	redis  *redis.Client
}

// NewService creates a new anomaly detection service
func NewService(cfg Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	svc := &Service{
		config:     cfg,
		logger:     cfg.Logger,
		clickhouse: cfg.ClickHouse,
		redis:      cfg.Redis,
		anomalies:  make(map[string]*Anomaly),
		detectors:  make(map[string]Detector),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Register detectors
	svc.detectors["cash_anomaly"] = &CashAnomalyDetector{
		logger: cfg.Logger,
		redis:  cfg.Redis,
	}
	svc.detectors["benford"] = &BenfordDetector{
		logger: cfg.Logger,
	}
	svc.detectors["circular"] = &CircularTransactionDetector{
		logger: cfg.Logger,
		redis:  cfg.Redis,
	}

	return svc
}

// Start begins the anomaly detection service
func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Starting anomaly detection service")

	// Start background detection loops
	s.wg.Add(1)
	go s.cashAnomalyDetectionLoop()
	s.wg.Add(1)
	go s.benfordAnalysisLoop()
	s.wg.Add(1)
	go s.circularTransactionLoop()

	s.logger.Info("Anomaly detection service started")
}

// Stop gracefully stops the service
func (s *Service) Stop() {
	s.logger.Info("Stopping anomaly detection service")
	s.cancel()
	s.wg.Wait()
	s.logger.Info("Anomaly detection service stopped")
}

// cashAnomalyDetectionLoop continuously monitors for cash anomalies
func (s *Service) cashAnomalyDetectionLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.detectCashAnomalies()
		}
	}
}

// detectCashAnomalies scans for cash-intensive business anomalies
func (s *Service) detectCashAnomalies() {
	ctx := context.Background()

	// Get all high-risk entities for cash analysis
	entities := s.getHighRiskEntities(ctx)

	for _, entityID := range entities {
		anomalies := s.detectors["cash_anomaly"].Detect(ctx, entityID)
		s.storeAnomalies(anomalies)
	}

	s.logger.Info("Cash anomaly detection completed", "entities_scanned", len(entities))
}

// benfordAnalysisLoop performs periodic Benford's Law analysis
func (s *Service) benfordAnalysisLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.performBenfordAnalysis()
		}
	}
}

// performBenfordAnalysis analyzes transaction data for Benford violations
func (s *Service) performBenfordAnalysis() {
	ctx := context.Background()

	// Get high-volume transaction entities
	entities := s.getHighVolumeEntities(ctx)

	for _, entityID := range entities {
		anomalies := s.detectors["benford"].Detect(ctx, entityID)
		s.storeAnomalies(anomalies)
	}

	s.logger.Info("Benford analysis completed", "entities_analyzed", len(entities))
}

// circularTransactionLoop monitors for circular transaction patterns
func (s *Service) circularTransactionLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.detectCircularTransactions()
		}
	}
}

// detectCircularTransactions scans for circular transaction patterns
func (s *Service) detectCircularTransactions() {
	ctx := context.Background()

	entities := s.getComplexTransactionEntities(ctx)

	for _, entityID := range entities {
		anomalies := s.detectors["circular"].Detect(ctx, entityID)
		s.storeAnomalies(anomalies)
	}

	s.logger.Info("Circular transaction detection completed", "entities_scanned", len(entities))
}

// getHighRiskEntities retrieves entities with elevated risk scores
func (s *Service) getHighRiskEntities(ctx context.Context) []string {
	// Query ClickHouse for high-risk entities
	query := `
		SELECT DISTINCT entity_id
		FROM neam.entity_risk_scores
		WHERE risk_score > 0.6
		AND update_time > now() - INTERVAL 7 DAY
	`

	var entities []string
	s.clickhouse.Query(ctx, query, &entities)

	return entities
}

// getHighVolumeEntities retrieves entities with high transaction volumes
func (s *Service) getHighVolumeEntities(ctx context.Context) []string {
	query := `
		SELECT DISTINCT entity_id
		FROM neam.transactions
		WHERE timestamp > now() - INTERVAL 30 DAY
		GROUP BY entity_id
		HAVING sum(amount) > 10000000
	`

	var entities []string
	s.clickhouse.Query(ctx, query, &entities)

	return entities
}

// getComplexTransactionEntities retrieves entities with complex transaction patterns
func (s *Service) getComplexTransactionEntities(ctx context.Context) []string {
	query := `
		SELECT DISTINCT from_entity_id
		FROM neam.transactions
		WHERE timestamp > now() - INTERVAL 30 DAY
		GROUP BY from_entity_id
		HAVING count(DISTINCT to_entity_id) > 50
	`

	var entities []string
	s.clickhouse.Query(ctx, query, &entities)

	return entities
}

// storeAnomalies persists detected anomalies
func (s *Service) storeAnomalies(anomalies []Anomaly) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, anomaly := range anomalies {
		s.anomalies[anomaly.ID] = &anomaly
	}
}

// DetectCashAnomalies performs cash anomaly detection for a specific entity
func (s *Service) DetectCashAnomalies(ctx context.Context, entityID string) ([]Anomaly, error) {
	return s.detectors["cash_anomaly"].Detect(ctx, entityID)
}

// DetectBenfordViolations performs Benford's Law analysis
func (s *Service) DetectBenfordViolations(ctx context.Context, entityID string) ([]Anomaly, error) {
	return s.detectors["benford"].Detect(ctx, entityID)
}

// DetectCircularTransactions performs circular transaction detection
func (s *Service) DetectCircularTransactions(ctx context.Context, entityID string) ([]Anomaly, error) {
	return s.detectors["circular"].Detect(ctx, entityID)
}

// ListAnomalies returns anomalies with optional filters
func (s *Service) ListAnomalies(ctx context.Context, filters map[string]string) ([]*Anomaly, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var anomalies []*Anomaly
	for _, a := range s.anomalies {
		if filters["type"] != "" && a.Type != filters["type"] {
			continue
		}
		if filters["severity"] != "" && a.Severity != filters["severity"] {
			continue
		}
		if filters["status"] != "" && a.Status != filters["status"] {
			continue
		}
		anomalies = append(anomalies, a)
	}

	return anomalies, nil
}

// GetAnomaly retrieves a specific anomaly
func (s *Service) GetAnomaly(ctx context.Context, id string) (*Anomaly, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	anomaly, exists := s.anomalies[id]
	if !exists {
		return nil, fmt.Errorf("anomaly not found: %s", id)
	}

	return anomaly, nil
}

// CashAnomalyDetector implementation
func (d *CashAnomalyDetector) Detect(ctx context.Context, entityID string) ([]Anomaly, error) {
	// Simulate detection logic
	anomaly := Anomaly{
		ID:         fmt.Sprintf("cash-anomaly-%d", time.Now().UnixNano()),
		Type:       "cash_anomaly",
		EntityID:   entityID,
		EntityType: "company",
		Severity:   "high",
		Score:      0.75,
		Indicators: []Indicator{
			{
				Name:        "cash_ratio",
				Value:       0.85,
				Threshold:   0.40,
				Deviation:   2.25,
				Unit:        "ratio",
				Description: "Cash to digital payment ratio exceeds peer average",
			},
		},
		DetectionDate: time.Now(),
		Status:        "detected",
	}

	return []Anomaly{anomaly}, nil
}

func (d *CashAnomalyDetector) GetName() string {
	return "Cash Anomaly Detector"
}

func (d *CashAnomalyDetector) GetType() string {
	return "cash_anomaly"
}

// BenfordDetector implementation
func (d *BenfordDetector) Detect(ctx context.Context, entityID string) ([]Anomaly, error) {
	// Benford's Law expected distribution
	expectedDistribution := []float64{
		0.301, 0.176, 0.125, 0.097, 0.079,
		0.067, 0.058, 0.051, 0.046,
	}

	// Simulate analysis - calculate chi-square deviation
	deviation := 0.05

	if deviation > 0.03 {
		anomaly := Anomaly{
			ID:         fmt.Sprintf("benford-%d", time.Now().UnixNano()),
			Type:       "benford_violation",
			EntityID:   entityID,
			EntityType: "company",
			Severity:   "high",
			Score:      0.80,
			Indicators: []Indicator{
				{
					Name:        "benford_deviation",
					Value:       deviation,
					Threshold:   0.03,
					Deviation:   deviation / 0.03,
					Unit:        "chi-square",
					Description: "Transaction distribution deviates from Benford's Law",
				},
			},
			DetectionDate: time.Now(),
			Status:        "detected",
		}
		return []Anomaly{anomaly}, nil
	}

	return []Anomaly{}, nil
}

func (d *BenfordDetector) GetName() string {
	return "Benford's Law Detector"
}

func (d *BenfordDetector) GetType() string {
	return "benford_violation"
}

// CircularTransactionDetector implementation
func (d *CircularTransactionDetector) Detect(ctx context.Context, entityID string) ([]Anomaly, error) {
	// Simulate circular transaction detection
	anomaly := Anomaly{
		ID:         fmt.Sprintf("circular-%d", time.Now().UnixNano()),
		Type:       "circular_transaction",
		EntityID:   entityID,
		EntityType: "company",
		Severity:   "critical",
		Score:      0.90,
		Indicators: []Indicator{
			{
				Name:        "cycle_length",
				Value:       3,
				Threshold:   1,
				Deviation:   3.0,
				Unit:        "transactions",
				Description: "Circular transaction pattern detected with 3-node cycle",
			},
			{
				Name:        "total_volume",
				Value:       5000000,
				Threshold:   1000000,
				Deviation:   5.0,
				Unit:        "currency",
				Description: "High-volume circular transactions detected",
			},
		},
		DetectionDate: time.Now(),
		Status:        "detected",
	}

	return []Anomaly{anomaly}, nil
}

func (d *CircularTransactionDetector) GetName() string {
	return "Circular Transaction Detector"
}

func (d *CircularTransactionDetector) GetType() string {
	return "circular_transaction"
}

// CalculateStandardDeviation calculates standard deviation for a dataset
func CalculateStandardDeviation(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	mean := sum / float64(len(data))

	// Calculate variance
	variance := 0.0
	for _, v := range data {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(data))

	return math.Sqrt(variance)
}

// CalculateZScore calculates the z-score for a value
func CalculateZScore(value, mean, stdDev float64) float64 {
	if stdDev == 0 {
		return 0
	}
	return (value - mean) / stdDev
}
