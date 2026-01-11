package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/jitenkr2030/csic-platform/extended-services/financial-crime-unit/internal/domain"
	"github.com/jitenkr2030/csic-platform/extended-services/financial-crime-unit/internal/repository"
)

// FCUServiceConfig holds the configuration for the FCU service
type FCUServiceConfig struct {
	Repo            repository.PostgresRepository
	Cache           *repository.RedisCache
	Detection       DetectionConfig
	Sanctions       SanctionsConfig
	CaseManagement  CaseManagementConfig
	SAR             SARConfig
	NetworkAnalysis NetworkAnalysisConfig
	RiskScoring     RiskScoringConfig
	Privacy         PrivacyConfig
	KafkaTopics     KafkaTopicsConfig
}

// KafkaTopicsConfig holds Kafka topic names
type KafkaTopicsConfig struct {
	Alerts          string
	SARs            string
	IdentityUpdates string
	Transactions    string
	Reports         string
}

// DetectionConfig holds detection settings
type DetectionConfig struct {
	Thresholds    ThresholdsConfig
	Transaction   TransactionConfig
	Geographic    GeographicConfig
	Velocity      VelocityConfig
}

type ThresholdsConfig struct {
	HighRisk           int
	MediumRisk         int
	LowRisk            int
	CriticalAlertScore int
}

type TransactionConfig struct {
	VelocityWindow           string
	MaxTransactionsPerWindow int
	MaxAmountPerWindow       int
	StructuringThreshold     int
	RoundAmountThreshold     int
}

type GeographicConfig struct {
	ImpossibleTravelWindow string
	HighRiskCountries      []string
	SanctionedCountries    []string
}

type VelocityConfig struct {
	DailyTransactionLimit  int
	HourlyTransactionLimit int
	MaxRecipientsPerDay    int
}

type SanctionsConfig struct {
	Enabled                 bool
	FuzzyMatchThreshold     float64
	PEPCheckEnabled         bool
	SanctionListUpdateInterval string
	WatchlistRefreshInterval string
}

type CaseManagementConfig struct {
	AutoCreateCasesAboveScore int
	DefaultAssignee           string
	Workflow                  []WorkflowState
	EscalationTimeout         string
}

type WorkflowState struct {
	State      string
	NextStates []string
}

type SARConfig struct {
	Enabled               bool
	TemplateFormat        string
	RegulatoryBody        string
	AutoGenerateThreshold int
	MandatoryFields       []string
	FilingDeadlineHours   int
}

type NetworkAnalysisConfig struct {
	Enabled         bool
	MaxHops         int
	AnalysisTimeout string
	HighRiskPatterns []string
}

type RiskScoringConfig struct {
	BaseScore float64
	Factors   RiskFactorsConfig
	Weights   RiskWeightsConfig
}

type RiskFactorsConfig struct {
	TransactionAmount  float64
	VelocityScore      float64
	GeographicRisk     float64
	EntityRisk         float64
	BehavioralPattern  float64
}

type RiskWeightsConfig struct {
	KYCLevel           float64
	AccountAge         float64
	TransactionHistory float64
	BlacklistHit       float64
}

type PrivacyConfig struct {
	DataRetention     string
	PIIEncryption     bool
	GDPRCompliance    bool
	AuditLogRetention string
	SARRetention      string
}

// FCUService handles all financial crime detection business logic
type FCUService interface {
	// Transaction screening
	ScreenTransaction(ctx context.Context, req *TransactionScreenRequest) (*ScreenResult, error)
	ScreenEntity(ctx context.Context, req *EntityScreenRequest) (*ScreenResult, error)
	ScreenBatch(ctx context.Context, req *BatchScreenRequest) ([]*ScreenResult, error)
	
	// Sanctions screening
	CheckSanctions(ctx context.Context, req *SanctionsCheckRequest) (*SanctionsResult, error)
	BulkSanctionsCheck(ctx context.Context, req *BulkSanctionsRequest) ([]*SanctionsResult, error)
	GetSanctionLists(ctx context.Context) ([]string, error)
	SyncSanctionLists(ctx context.Context) error
	
	// Alerts management
	GetAlerts(ctx context.Context, filter AlertFilterRequest) ([]domain.Alert, error)
	GetAlert(ctx context.Context, id string) (*domain.Alert, error)
	AcknowledgeAlert(ctx context.Context, id, userID string) error
	EscalateAlert(ctx context.Context, id, escalateTo string) error
	
	// Case management
	GetCases(ctx context.Context, filter CaseFilterRequest) ([]domain.Case, error)
	CreateCase(ctx context.Context, req *CreateCaseRequest) (*domain.Case, error)
	GetCase(ctx context.Context, id string) (*domain.Case, error)
	UpdateCase(ctx context.Context, caseObj *domain.Case) error
	TransitionCaseWorkflow(ctx context.Context, caseID, newStatus, userID string) error
	AssignCase(ctx context.Context, caseID, assigneeID string) error
	AddEvidenceToCase(ctx context.Context, caseID string, evidence *domain.Evidence) error
	
	// SAR management
	GetSARs(ctx context.Context, filter SARFilterRequest) ([]domain.SAR, error)
	CreateSAR(ctx context.Context, req *CreateSARRequest) (*domain.SAR, error)
	GetSAR(ctx context.Context, id string) (*domain.SAR, error)
	UpdateSAR(ctx context.Context, sar *domain.SAR) error
	ValidateSAR(ctx context.Context, id string) error
	SubmitSAR(ctx context.Context, id, submittedBy string) error
	GenerateAutoSAR(ctx context.Context, caseID string) (*domain.SAR, error)
	
	// Risk scoring
	GetRiskScore(ctx context.Context, entityID string) (*RiskScoreResult, error)
	GetRiskProfile(ctx context.Context, entityID string) (*domain.RiskProfile, error)
	GetRiskHistory(ctx context.Context, entityID string, limit int) ([]domain.HistoricalScore, error)
	CalculateRisk(ctx context.Context, req *CalculateRiskRequest) (*RiskScoreResult, error)
	
	// Network analysis
	AnalyzeNetwork(ctx context.Context, req *NetworkAnalysisRequest) (*domain.NetworkGraph, error)
	TraceTransactionFlow(ctx context.Context, req *TraceRequest) ([]*TransactionFlow, error)
	GetNetworkPatterns(ctx context.Context, entityID string) ([]domain.NetworkPattern, error)
	GetEntityConnections(ctx context.Context, entityID string) ([]*ConnectionInfo, error)
	
	// Entity management
	GetEntity(ctx context.Context, id string) (*domain.Entity, error)
	UpdateEntity(ctx context.Context, entity *domain.Entity) error
	GetEntityTransactions(ctx context.Context, entityID string, limit, offset int) ([]domain.Transaction, error)
	GetEntityAlerts(ctx context.Context, entityID string) ([]domain.Alert, error)
	AddToWatchlist(ctx context.Context, req *AddToWatchlistRequest) error
	RemoveFromWatchlist(ctx context.Context, entityID string) error
	GetWatchlistStatus(ctx context.Context, entityID string) (*domain.WatchlistEntry, error)
	
	// Investigation
	GetInvestigation(ctx context.Context, id string) (*domain.Investigation, error)
	StartInvestigation(ctx context.Context, req *StartInvestigationRequest) (*domain.Investigation, error)
	AddInvestigationNote(ctx context.Context, req *AddNoteRequest) error
	AddTimelineEvent(ctx context.Context, req *AddTimelineRequest) error
	
	// Reports
	GetReportSummary(ctx context.Context, req *ReportSummaryRequest) (*ReportSummary, error)
	GetTrendAnalysis(ctx context.Context, req *TrendAnalysisRequest) (*TrendAnalysis, error)
	GetDetectionEffectiveness(ctx context.Context, req *EffectivenessRequest) (*DetectionEffectiveness, error)
	GetComplianceReport(ctx context.Context, req *ComplianceReportRequest) (*ComplianceReport, error)
}

// fcuService implements FCUService
type fcuService struct {
	config FCUServiceConfig
	now    func() time.Time
}

// NewFCUService creates a new FCU service instance
func NewFCUService(config FCUServiceConfig) FCUService {
	return &fcuService{
		config: config,
		now:    time.Now,
	}
}

// Screening types and requests

type TransactionScreenRequest struct {
	TransactionID    string                 `json:"transaction_id"`
	TransactionHash  string                 `json:"transaction_hash"`
	Type             string                 `json:"type"`
	Asset            string                 `json:"asset"`
	Amount           float64                `json:"amount"`
	Currency         string                 `json:"currency"`
	SenderID         string                 `json:"sender_id"`
	SenderWallet     *string                `json:"sender_wallet,omitempty"`
	ReceiverID       string                 `json:"receiver_id"`
	ReceiverWallet   *string                `json:"receiver_wallet,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type EntityScreenRequest struct {
	EntityID   string   `json:"entity_id"`
	EntityType string   `json:"entity_type"`
	Name       string   `json:"name"`
	Email      *string  `json:"email,omitempty"`
	Phone      *string  `json:"phone,omitempty"`
	Wallets    []string `json:"wallets,omitempty"`
	NationalID *string  `json:"national_id,omitempty"`
}

type BatchScreenRequest struct {
	Transactions []TransactionScreenRequest `json:"transactions"`
	Entities     []EntityScreenRequest      `json:"entities"`
}

type ScreenResult struct {
	ID             string                 `json:"id"`
	ScreenedItem   string                 `json:"screened_item"`
	ItemID         string                 `json:"item_id"`
	Decision       string                 `json:"decision"`
	RiskScore      int                    `json:"risk_score"`
	RiskLevel      string                 `json:"risk_level"`
	TriggeredRules []string               `json:"triggered_rules"`
	Alerts         []*TriggeredAlert      `json:"alerts,omitempty"`
	Details        map[string]interface{} `json:"details,omitempty"`
	ScreenedAt     time.Time              `json:"screened_at"`
}

type TriggeredAlert struct {
	RuleID   string `json:"rule_id"`
	RuleName string `json:"rule_name"`
	Severity string `json:"severity"`
	Score    int    `json:"score"`
}

// ScreenTransaction screens a transaction for financial crime indicators
func (s *fcuService) ScreenTransaction(ctx context.Context, req *TransactionScreenRequest) (*ScreenResult, error) {
	result := &ScreenResult{
		ID:           generateID(),
		ScreenedItem: "transaction",
		ItemID:       req.TransactionID,
		ScreenedAt:   s.now(),
		TriggeredRules: []string{},
		Details:       make(map[string]interface{}),
	}

	var riskScore int
	var triggeredRules []string
	var alerts []*TriggeredAlert

	// Check velocity
	velocityScore, velocityAlerts := s.checkVelocity(ctx, req)
	riskScore += velocityScore
	triggeredRules = append(triggeredRules, velocityAlerts...)
	alerts = append(alerts, velocityAlerts...)

	// Check for structuring
	structuringScore, structuringAlerts := s.checkStructuring(req)
	riskScore += structuringScore
	triggeredRules = append(triggeredRules, structuringAlerts...)
	alerts = append(alerts, structuringAlerts...)

	// Check geographic risk
	geoScore, geoAlerts := s.checkGeographicRisk(req)
	riskScore += geoScore
	triggeredRules = append(triggeredRules, geoAlerts...)
	alerts = append(alerts, geoAlerts...)

	// Check sanctions
	sanctionsResult, err := s.CheckSanctions(ctx, &SanctionsCheckRequest{
		EntityID:   req.SenderID,
		EntityType: "individual",
		Wallets:    filterNilString(req.SenderWallet),
	})
	if err == nil && sanctionsResult.Hit {
		riskScore += 50
		triggeredRules = append(triggeredRules, "sanctions_check")
		alerts = append(alerts, &TriggeredAlert{
			RuleID:   "SANCTIONS_HIT",
			RuleName: "Sanctions List Match",
			Severity: "critical",
			Score:    100,
		})
	}

	// Check watchlist
	watchlistStatus, _ := s.config.Repo.GetWatchlistStatus(ctx, req.SenderID)
	if watchlistStatus != nil {
		riskScore += 40
		triggeredRules = append(triggeredRules, "watchlist_hit")
		alerts = append(alerts, &TriggeredAlert{
			RuleID:   "WATCHLIST_HIT",
			RuleName: "Watchlist Match",
			Severity: "critical",
			Score:    90,
		})
	}

	// Determine decision
	if riskScore >= s.config.Detection.Thresholds.HighRisk {
		result.Decision = "BLOCK"
	} else if riskScore >= s.config.Detection.Thresholds.MediumRisk {
		result.Decision = "FLAG"
	} else {
		result.Decision = "CLEAR"
	}

	result.RiskScore = riskScore
	result.RiskLevel = s.calculateRiskLevel(riskScore)
	result.TriggeredRules = triggeredRules
	result.Alerts = alerts
	result.Details["amount"] = req.Amount
	result.Details["currency"] = req.Currency

	// Create alerts for high-risk transactions
	for _, alert := range alerts {
		if alert.Severity == "critical" || alert.Severity == "high" {
			alertObj := &domain.Alert{
				ID:            generateID(),
				Type:          domain.AlertType(alert.RuleID),
				Severity:      domain.AlertSeverity(alert.Severity),
				Score:         alert.Score,
				RuleTriggered: alert.RuleName,
				EntityID:      req.SenderID,
				EntityType:    "transaction",
				TransactionID: &req.TransactionID,
				Description:   fmt.Sprintf("High-risk transaction detected: %s", alert.RuleName),
				CreatedAt:     s.now(),
			}
			s.config.Repo.CreateAlert(ctx, alertObj)

			// Auto-create case if risk is very high
			if riskScore >= s.config.CaseManagement.AutoCreateCasesAboveScore {
				s.createAutoCase(ctx, alertObj)
			}
		}
	}

	// Record velocity data
	velocityData := &domain.VelocityData{
		ID:               generateID(),
		EntityID:         req.SenderID,
		MetricType:       "transaction_count",
		WindowStart:      s.now().Add(-1 * time.Hour),
		WindowEnd:        s.now(),
		TransactionCount: 1,
		TotalAmount:      req.Amount,
		AlertTriggered:   riskScore >= s.config.Detection.Thresholds.MediumRisk,
		CreatedAt:        s.now(),
	}
	s.config.Repo.RecordVelocityData(ctx, velocityData)

	return result, nil
}

// checkVelocity checks transaction velocity
func (s *fcuService) checkVelocity(ctx context.Context, req *TransactionScreenRequest) (int, []string) {
	var triggeredRules []string
	score := 0

	// Get recent velocity data
	velocityData, _ := s.config.Repo.GetVelocityData(
		ctx, req.SenderID, "transaction_count",
		s.now().Add(-1*time.Hour), s.now(),
	)

	if velocityData != nil {
		if velocityData.TransactionCount >= s.config.Detection.Transaction.MaxTransactionsPerWindow {
			score += 30
			triggeredRules = append(triggeredRules, "velocity_high")
		}
		if velocityData.TotalAmount+req.Amount >= float64(s.config.Detection.Transaction.MaxAmountPerWindow) {
			score += 25
			triggeredRules = append(triggeredRules, "velocity_amount")
		}
	}

	return score, triggeredRules
}

// checkStructuring detects structuring patterns
func (s *fcuService) checkStructuring(req *TransactionScreenRequest) (int, []string) {
	var triggeredRules []string
	score := 0

	// Check for structuring (transactions just below reporting threshold)
	if req.Amount >= float64(s.config.Detection.Transaction.StructuringThreshold-1000) &&
		req.Amount < float64(s.config.Detection.Transaction.StructuringThreshold) {
		score += 40
		triggeredRules = append(triggeredRules, "structuring")
	}

	// Check for round amounts (potential smurfing)
	amountMod := math.Mod(req.Amount, float64(s.config.Detection.Transaction.RoundAmountThreshold))
	if amountMod < 1 || amountMod > float64(s.config.Detection.Transaction.RoundAmountThreshold-1) {
		score += 15
		triggeredRules = append(triggeredRules, "round_amount")
	}

	return score, triggeredRules
}

// checkGeographicRisk checks for geographic risk factors
func (s *fcuService) checkGeographicRisk(req *TransactionScreenRequest) (int, []string) {
	var triggeredRules []string
	score := 0

	// Simplified check - in production, use IP geolocation service
	metadata, ok := req.Metadata["geo"]
	if ok {
		if geoMap, ok := metadata.(map[string]interface{}); ok {
			if country, ok := geoMap["country"].(string); ok {
				for _, highRiskCountry := range s.config.Detection.Geographic.HighRiskCountries {
					if country == highRiskCountry {
						score += 20
						triggeredRules = append(triggeredRules, "high_risk_country")
						break
					}
				}
				for _, sanctionedCountry := range s.config.Detection.Geographic.SanctionedCountries {
					if country == sanctionedCountry {
						score += 50
						triggeredRules = append(triggeredRules, "sanctioned_country")
						break
					}
				}
			}
		}
	}

	return score, triggeredRules
}

// ScreenEntity screens an entity for sanctions and watchlist
func (s *fcuService) ScreenEntity(ctx context.Context, req *EntityScreenRequest) (*ScreenResult, error) {
	result := &ScreenResult{
		ID:           generateID(),
		ScreenedItem: "entity",
		ItemID:       req.EntityID,
		ScreenedAt:   s.now(),
		TriggeredRules: []string{},
		Details:       make(map[string]interface{}),
	}

	var riskScore int
	var triggeredRules []string

	// Check sanctions
	sanctionsResult, err := s.CheckSanctions(ctx, &SanctionsCheckRequest{
		EntityID:   req.EntityID,
		EntityType: req.EntityType,
		Name:       req.Name,
		Wallets:    req.Wallets,
	})
	if err != nil {
		return nil, fmt.Errorf("sanctions check failed: %w", err)
	}

	if sanctionsResult.Hit {
		riskScore += 80
		triggeredRules = append(triggeredRules, "sanctions_match")
		result.Details["sanctions_hit"] = sanctionsResult.MatchedEntries
	}

	// Check watchlist
	watchlistStatus, err := s.GetWatchlistStatus(ctx, req.EntityID)
	if err == nil && watchlistStatus != nil {
		riskScore += 50
		triggeredRules = append(triggeredRules, "watchlist_match")
		result.Details["watchlist_reason"] = watchlistStatus.Reason
	}

	// Get entity risk profile
	profile, _ := s.config.Repo.GetRiskProfile(ctx, req.EntityID)
	if profile != nil {
		result.Details["existing_risk_score"] = profile.OverallScore
		result.Details["existing_risk_level"] = profile.RiskLevel
	}

	// Determine decision
	if riskScore >= 70 {
		result.Decision = "BLOCK"
	} else if riskScore >= 40 {
		result.Decision = "FLAG"
	} else {
		result.Decision = "CLEAR"
	}

	result.RiskScore = riskScore
	result.RiskLevel = s.calculateRiskLevel(riskScore)
	result.TriggeredRules = triggeredRules

	return result, nil
}

func (s *fcuService) ScreenBatch(ctx context.Context, req *BatchScreenRequest) ([]*ScreenResult, error) {
	var results []*ScreenResult

	for _, tx := range req.Transactions {
		result, err := s.ScreenTransaction(ctx, &tx)
		if err != nil {
			log.Printf("Failed to screen transaction %s: %v", tx.TransactionID, err)
			continue
		}
		results = append(results, result)
	}

	for _, entity := range req.Entities {
		result, err := s.ScreenEntity(ctx, &entity)
		if err != nil {
			log.Printf("Failed to screen entity %s: %v", entity.EntityID, err)
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// Sanctions checking

type SanctionsCheckRequest struct {
	EntityID   string   `json:"entity_id"`
	EntityType string   `json:"entity_type"`
	Name       string   `json:"name,omitempty"`
	Wallets    []string `json:"wallets,omitempty"`
	NationalID *string  `json:"national_id,omitempty"`
	Email      *string  `json:"email,omitempty"`
}

type SanctionsResult struct {
	Hit            bool                     `json:"hit"`
	MatchedEntries []MatchedEntry          `json:"matched_entries,omitempty"`
	Score          float64                 `json:"score"`
	Details        map[string]interface{}  `json:"details,omitempty"`
}

type MatchedEntry struct {
	ListName   string `json:"list_name"`
	EntryID    string `json:"entry_id"`
	Name       string `json:"name"`
	MatchScore float64 `json:"match_score"`
}

// CheckSanctions checks an entity against sanctions lists
func (s *fcuService) CheckSanctions(ctx context.Context, req *SanctionsCheckRequest) (*SanctionsResult, error) {
	result := &SanctionsResult{
		Hit:            false,
		MatchedEntries: []MatchedEntry{},
		Details:        make(map[string]interface{}),
	}

	if !s.config.Sanctions.Enabled {
		return result, nil
	}

	// Check wallet addresses first (exact match)
	for _, wallet := range req.Wallets {
		sanction, _ := s.config.Repo.GetSanctionsByWallet(ctx, wallet)
		if sanction != nil {
			result.Hit = true
			result.MatchedEntries = append(result.MatchedEntries, MatchedEntry{
				ListName:   sanction.ListName,
				EntryID:    sanction.ID,
				Name:       sanction.Name,
				MatchScore: 1.0,
			})
		}
	}

	// Check name against sanctions list (fuzzy match)
	if req.Name != "" && len(result.MatchedEntries) == 0 {
		sanctions, _ := s.config.Repo.SearchSanctions(ctx, req.Name)
		for _, sanction := range sanctions {
			// Simplified fuzzy matching - in production, use proper fuzzy matching library
			matchScore := s.calculateNameSimilarity(req.Name, sanction.Name)
			if matchScore >= s.config.Sanctions.FuzzyMatchThreshold {
				result.Hit = true
				result.MatchedEntries = append(result.MatchedEntries, MatchedEntry{
					ListName:   sanction.ListName,
					EntryID:    sanction.ID,
					Name:       sanction.Name,
					MatchScore: matchScore,
				})
			}
		}
	}

	// Calculate overall score
	if result.Hit {
		var maxScore float64
		for _, entry := range result.MatchedEntries {
			if entry.MatchScore > maxScore {
				maxScore = entry.MatchScore
			}
		}
		result.Score = maxScore
	}

	result.Details["check_count"] = len(req.Wallets) + 1
	result.Details["lists_checked"] = 3

	return result, nil
}

func (s *fcuService) BulkSanctionsCheck(ctx context.Context, req *BulkSanctionsRequest) ([]*SanctionsResult, error) {
	var results []*SanctionsResult
	for _, checkReq := range req.Requests {
		result, err := s.CheckSanctions(ctx, &checkReq)
		if err != nil {
			log.Printf("Failed to check sanctions for %s: %v", checkReq.EntityID, err)
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

type BulkSanctionsRequest struct {
	Requests []SanctionsCheckRequest `json:"requests"`
}

func (s *fcuService) GetSanctionLists(ctx context.Context) ([]string, error) {
	// Return list names - in production, query actual lists
	return []string{"OFAC", "EU Sanctions", "UN Sanctions", "FATF"}, nil
}

func (s *fcuService) SyncSanctionLists(ctx context.Context) error {
	// In production, fetch and update sanctions lists from official sources
	log.Println("Syncing sanction lists...")
	return nil
}

func (s *fcuService) calculateNameSimilarity(name1, name2 string) float64 {
	// Simplified name matching - in production, use proper algorithm
	name1 = strings.ToLower(name1)
	name2 = strings.ToLower(name2)

	if name1 == name2 {
		return 1.0
	}

	// Check if one contains the other
	if strings.Contains(name1, name2) || strings.Contains(name2, name1) {
		return 0.8
	}

	// Check first name match
	name1Parts := strings.Fields(name1)
	name2Parts := strings.Fields(name2)
	if len(name1Parts) > 0 && len(name2Parts) > 0 {
		if name1Parts[0] == name2Parts[0] {
			return 0.7
		}
	}

	return 0.5
}

// Case management

type AlertFilterRequest struct {
	EntityID   string `json:"entity_id,omitempty"`
	Severity   string `json:"severity,omitempty"`
	Status     string `json:"status,omitempty"`
	DateFrom   *time.Time `json:"date_from,omitempty"`
	DateTo     *time.Time `json:"date_to,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}

type CaseFilterRequest struct {
	Status     string `json:"status,omitempty"`
	Priority   string `json:"priority,omitempty"`
	AssigneeID string `json:"assignee_id,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}

type CreateCaseRequest struct {
	Title          string              `json:"title" binding:"required"`
	Description    string              `json:"description" binding:"required"`
	SubjectID      string              `json:"subject_id" binding:"required"`
	SubjectType    string              `json:"subject_type" binding:"required"`
	Priority       string              `json:"priority,omitempty"`
	AlertIDs       []string            `json:"alert_ids,omitempty"`
	RiskScore      int                 `json:"risk_score,omitempty"`
	TotalAmount    float64             `json:"total_amount,omitempty"`
	Currency       string              `json:"currency,omitempty"`
	Tags           []string            `json:"tags,omitempty"`
	CreatedBy      string              `json:"created_by" binding:"required"`
}

func (s *fcuService) GetAlerts(ctx context.Context, filter AlertFilterRequest) ([]domain.Alert, error) {
	return s.config.Repo.GetAlerts(ctx, repository.AlertFilter{
		EntityID:  filter.EntityID,
		Severity:  filter.Severity,
		Status:    filter.Status,
		DateFrom:  filter.DateFrom,
		DateTo:    filter.DateTo,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	})
}

func (s *fcuService) GetAlert(ctx context.Context, id string) (*domain.Alert, error) {
	return s.config.Repo.GetAlert(ctx, id)
}

func (s *fcuService) AcknowledgeAlert(ctx context.Context, id, userID string) error {
	return s.config.Repo.AcknowledgeAlert(ctx, id, userID)
}

func (s *fcuService) EscalateAlert(ctx context.Context, id, escalateTo string) error {
	return s.config.Repo.EscalateAlert(ctx, id, escalateTo)
}

func (s *fcuService) GetCases(ctx context.Context, filter CaseFilterRequest) ([]domain.Case, error) {
	return s.config.Repo.GetCases(ctx, repository.CaseFilter{
		Status:     filter.Status,
		Priority:   filter.Priority,
		AssigneeID: filter.AssigneeID,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
	})
}

func (s *fcuService) CreateCase(ctx context.Context, req *CreateCaseRequest) (*domain.Case, error) {
	caseObj := &domain.Case{
		Title:          req.Title,
		Description:    req.Description,
		SubjectID:      req.SubjectID,
		SubjectType:    req.SubjectType,
		Priority:       domain.CasePriority(req.Priority),
		AlertIDs:       req.AlertIDs,
		RiskScore:      req.RiskScore,
		TotalAmount:    req.TotalAmount,
		Currency:       req.Currency,
		Tags:           req.Tags,
		CreatedBy:      req.CreatedBy,
	}

	if caseObj.Priority == "" {
		caseObj.Priority = domain.CasePriorityMedium
	}
	if req.Currency == "" {
		caseObj.Currency = "USD"
	}

	if err := s.config.Repo.CreateCase(ctx, caseObj); err != nil {
		return nil, err
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		ID:         generateID(),
		Timestamp:  s.now(),
		Action:     "CASE_CREATED",
		ActorID:    req.CreatedBy,
		ActorType:  "user",
		EntityType: "case",
		EntityID:   caseObj.ID,
		NewValue:   stringPtr(caseObj.Title),
	}
	s.config.Repo.CreateAuditLog(ctx, auditLog)

	return caseObj, nil
}

func (s *fcuService) createAutoCase(ctx context.Context, alert *domain.Alert) {
	caseObj := &domain.Case{
		Title:       fmt.Sprintf("Auto-generated case from alert: %s", alert.Type),
		Description: alert.Description,
		SubjectID:   alert.EntityID,
		SubjectType: alert.EntityType,
		Priority:    domain.CasePriorityHigh,
		AlertIDs:    []string{alert.ID},
		RiskScore:   alert.Score,
		CreatedBy:   "system",
	}

	s.config.Repo.CreateCase(ctx, caseObj)

	// Link alert to case
	alert.CaseID = &caseObj.ID
	s.config.Repo.UpdateAlert(ctx, alert)
}

func (s *fcuService) GetCase(ctx context.Context, id string) (*domain.Case, error) {
	return s.config.Repo.GetCase(ctx, id)
}

func (s *fcuService) UpdateCase(ctx context.Context, caseObj *domain.Case) error {
	caseObj.UpdatedAt = s.now()
	return s.config.Repo.UpdateCase(ctx, caseObj)
}

func (s *fcuService) TransitionCaseWorkflow(ctx context.Context, caseID, newStatus, userID string) error {
	return s.config.Repo.TransitionCaseWorkflow(ctx, caseID, newStatus, userID)
}

func (s *fcuService) AssignCase(ctx context.Context, caseID, assigneeID string) error {
	return s.config.Repo.AssignCase(ctx, caseID, assigneeID)
}

func (s *fcuService) AddEvidenceToCase(ctx context.Context, caseID string, evidence *domain.Evidence) error {
	evidence.ID = generateID()
	evidence.AddedAt = s.now()
	return s.config.Repo.AddEvidenceToCase(ctx, caseID, evidence)
}

// SAR Management

type SARFilterRequest struct {
	CaseID         string `json:"case_id,omitempty"`
	Status         string `json:"status,omitempty"`
	RegulatoryBody string `json:"regulatory_body,omitempty"`
	Limit          int    `json:"limit,omitempty"`
	Offset         int    `json:"offset,omitempty"`
}

type CreateSARRequest struct {
	CaseID            string                  `json:"case_id" binding:"required"`
	RegulatoryBody    string                  `json:"regulatory_body" binding:"required"`
	ReportType        string                  `json:"report_type" binding:"required"`
	SubjectID         string                  `json:"subject_id" binding:"required"`
	SubjectName       string                  `json:"subject_name" binding:"required"`
	SubjectType       string                  `json:"subject_type" binding:"required"`
	TotalAmount       float64                 `json:"total_amount"`
	Currency          string                  `json:"currency" binding:"required"`
	Narrative         string                  `json:"narrative" binding:"required"`
	SuspiciousActivity string                 `json:"suspicious_activity" binding:"required"`
	TransactionIDs    []string                `json:"transaction_ids,omitempty"`
	FilingInstitution string                  `json:"filing_institution" binding:"required"`
	FilingUserID      string                  `json:"filing_user_id" binding:"required"`
}

func (s *fcuService) GetSARs(ctx context.Context, filter SARFilterRequest) ([]domain.SAR, error) {
	return s.config.Repo.GetSARs(ctx, repository.SARFilter{
		CaseID:         filter.CaseID,
		Status:         filter.Status,
		RegulatoryBody: filter.RegulatoryBody,
		Limit:          filter.Limit,
		Offset:         filter.Offset,
	})
}

func (s *fcuService) CreateSAR(ctx context.Context, req *CreateSARRequest) (*domain.SAR, error) {
	sar := &domain.SAR{
		CaseID:            req.CaseID,
		RegulatoryBody:    req.RegulatoryBody,
		ReportType:        req.ReportType,
		SubjectID:         req.SubjectID,
		SubjectName:       req.SubjectName,
		SubjectType:       req.SubjectType,
		TotalAmount:       req.TotalAmount,
		Currency:          req.Currency,
		Narrative:         req.Narrative,
		SuspiciousActivity: req.SuspiciousActivity,
		TransactionDetails: s.buildTransactionDetails(ctx, req.TransactionIDs),
		FilingInstitution: req.FilingInstitution,
		FilingUserID:      req.FilingUserID,
	}

	if err := s.config.Repo.CreateSAR(ctx, sar); err != nil {
		return nil, err
	}

	return sar, nil
}

func (s *fcuService) buildTransactionDetails(ctx context.Context, txIDs []string) []domain.TransactionSummary {
	var details []domain.TransactionSummary
	for _, txID := range txIDs {
		tx, _ := s.config.Repo.GetTransaction(ctx, txID)
		if tx != nil {
			details = append(details, domain.TransactionSummary{
				TransactionID:   tx.ID,
				Date:            tx.Timestamp,
				Amount:          tx.Amount,
				Currency:        tx.Currency,
				TransactionType: tx.Type,
				Description:     fmt.Sprintf("%s transaction of %f %s", tx.Type, tx.Amount, tx.Currency),
			})
		}
	}
	return details
}

func (s *fcuService) GetSAR(ctx context.Context, id string) (*domain.SAR, error) {
	return s.config.Repo.GetSAR(ctx, id)
}

func (s *fcuService) UpdateSAR(ctx context.Context, sar *domain.SAR) error {
	sar.UpdatedAt = s.now()
	return s.config.Repo.UpdateSAR(ctx, sar)
}

func (s *fcuService) ValidateSAR(ctx context.Context, id string) error {
	sar, err := s.config.Repo.GetSAR(ctx, id)
	if err != nil {
		return err
	}

	// Check mandatory fields
	for _, field := range s.config.SAR.MandatoryFields {
		switch field {
		case "subject_id":
			if sar.SubjectID == "" {
				return fmt.Errorf("missing mandatory field: %s", field)
			}
		case "transaction_details":
			if len(sar.TransactionDetails) == 0 {
				return fmt.Errorf("missing mandatory field: %s", field)
			}
		case "narrative":
			if sar.Narrative == "" {
				return fmt.Errorf("missing mandatory field: %s", field)
			}
		case "filing_institution":
			if sar.FilingInstitution == "" {
				return fmt.Errorf("missing mandatory field: %s", field)
			}
		}
	}

	sar.Status = domain.SARStatusValidated
	sar.UpdatedAt = s.now()
	return s.config.Repo.UpdateSAR(ctx, sar)
}

func (s *fcuService) SubmitSAR(ctx context.Context, id, submittedBy string) error {
	sar, err := s.config.Repo.GetSAR(ctx, id)
	if err != nil {
		return err
	}

	if sar.Status != domain.SARStatusValidated {
		return errors.New("SAR must be validated before submission")
	}

	now := s.now()
	sar.Status = domain.SARStatusSubmitted
	sar.SubmittedAt = &now
	sar.ReviewerID = &submittedBy
	sar.UpdatedAt = now

	if err := s.config.Repo.UpdateSAR(ctx, sar); err != nil {
		return err
	}

	// In production, submit to regulatory body via secure channel
	log.Printf("SAR %s submitted to %s", sar.SARNumber, sar.RegulatoryBody)

	return nil
}

func (s *fcuService) GenerateAutoSAR(ctx context.Context, caseID string) (*domain.SAR, error) {
	caseObj, err := s.config.Repo.GetCase(ctx, caseID)
	if err != nil {
		return nil, err
	}

	// Build SAR from case data
	sar := &domain.SAR{
		CaseID:            caseID,
		RegulatoryBody:    s.config.SAR.RegulatoryBody,
		ReportType:        "Automated SAR",
		SubjectID:         caseObj.SubjectID,
		SubjectName:       caseObj.SubjectID, // In production, fetch actual name
		SubjectType:       caseObj.SubjectType,
		TotalAmount:       caseObj.TotalAmount,
		Currency:          caseObj.Currency,
		Narrative:         caseObj.Description,
		SuspiciousActivity: fmt.Sprintf("Automated report generated from case %s", caseObj.CaseNumber),
		FilingInstitution: "CSIC Platform",
		FilingUserID:      "system",
	}

	if err := s.config.Repo.CreateSAR(ctx, sar); err != nil {
		return nil, err
	}

	// Link SAR to case
	caseObj.SARID = &sar.ID
	caseObj.Status = domain.CaseStatusSARFiled
	s.config.Repo.UpdateCase(ctx, caseObj)

	return sar, nil
}

// Risk Scoring

type RiskScoreResult struct {
	EntityID     string  `json:"entity_id"`
	Score        int     `json:"score"`
	Level        string  `json:"level"`
	Components   map[string]float64 `json:"components"`
	Factors      []string `json:"factors"`
	CalculatedAt time.Time `json:"calculated_at"`
}

type CalculateRiskRequest struct {
	EntityID   string                 `json:"entity_id" binding:"required"`
	Transaction *TransactionScreenRequest `json:"transaction,omitempty"`
}

func (s *fcuService) GetRiskScore(ctx context.Context, entityID string) (*RiskScoreResult, error) {
	profile, err := s.config.Repo.GetRiskProfile(ctx, entityID)
	if err != nil {
		return nil, err
	}

	if profile == nil {
		return &RiskScoreResult{
			EntityID:     entityID,
			Score:        0,
			Level:        "unknown",
			CalculatedAt: s.now(),
		}, nil
	}

	return &RiskScoreResult{
		EntityID: entityID,
		Score:    profile.OverallScore,
		Level:    profile.RiskLevel,
		Components: map[string]float64{
			"transaction":  profile.Components.TransactionRisk,
			"velocity":     profile.Components.VelocityRisk,
			"geographic":   profile.Components.GeographicRisk,
			"entity":       profile.Components.EntityRisk,
			"behavioral":   profile.Components.BehavioralRisk,
		},
		CalculatedAt: profile.LastAssessmentAt,
	}, nil
}

func (s *fcuService) GetRiskProfile(ctx context.Context, entityID string) (*domain.RiskProfile, error) {
	return s.config.Repo.GetRiskProfile(ctx, entityID)
}

func (s *fcuService) GetRiskHistory(ctx context.Context, entityID string, limit int) ([]domain.HistoricalScore, error) {
	return s.config.Repo.GetRiskHistory(ctx, entityID, limit)
}

func (s *fcuService) CalculateRisk(ctx context.Context, req *CalculateRiskRequest) (*RiskScoreResult, error) {
	entity, _ := s.config.Repo.GetEntity(ctx, req.EntityID)

	profile := &domain.RiskProfile{
		EntityID:         req.EntityID,
		Components:       domain.RiskComponents{},
		HistoricalScores: []domain.HistoricalScore{},
		LastAssessmentAt: s.now(),
	}

	var factors []string
	var totalScore float64

	// Calculate transaction risk
	if req.Transaction != nil {
		profile.Components.TransactionRisk = s.calculateTransactionRisk(req.Transaction)
		if profile.Components.TransactionRisk > 0.5 {
			factors = append(factors, "high_value_transaction")
		}
	}

	// Get entity risk
	if entity != nil {
		profile.Components.EntityRisk = s.calculateEntityRisk(entity)
		profile.Components.SanctionsRisk = s.calculateSanctionsRisk(entity)
		profile.Components.PepRisk = s.calculatePEPRisk(entity)
		profile.Components.BlacklistRisk = s.calculateBlacklistRisk(entity)
	}

	// Get velocity risk
	velocityData, _ := s.config.Repo.GetVelocityData(ctx, req.EntityID, "transaction_count", s.now().Add(-24*time.Hour), s.now())
	if velocityData != nil && velocityData.TransactionCount > 10 {
		profile.Components.VelocityRisk = 0.6
		factors = append(factors, "high_velocity")
	}

	// Calculate overall score
	totalScore = profile.Components.TransactionRisk*0.3 +
		profile.Components.VelocityRisk*0.2 +
		profile.Components.GeographicRisk*0.15 +
		profile.Components.EntityRisk*0.25 +
		profile.Components.BehavioralRisk*0.1

	score := int(totalScore * 100)
	profile.OverallScore = score
	profile.RiskLevel = s.calculateRiskLevel(score)

	// Save profile
	s.config.Repo.SaveRiskProfile(ctx, profile)

	return &RiskScoreResult{
		EntityID: req.EntityID,
		Score:    score,
		Level:    profile.RiskLevel,
		Components: map[string]float64{
			"transaction": profile.Components.TransactionRisk,
			"velocity":    profile.Components.VelocityRisk,
			"geographic":  profile.Components.GeographicRisk,
			"entity":      profile.Components.EntityRisk,
			"behavioral":  profile.Components.BehavioralRisk,
		},
		Factors:      factors,
		CalculatedAt: s.now(),
	}, nil
}

func (s *fcuService) calculateTransactionRisk(tx *TransactionScreenRequest) float64 {
	// Simplified risk calculation based on transaction amount
	if tx.Amount > 100000 {
		return 0.9
	} else if tx.Amount > 50000 {
		return 0.7
	} else if tx.Amount > 10000 {
		return 0.5
	} else if tx.Amount > 1000 {
		return 0.3
	}
	return 0.1
}

func (s *fcuService) calculateEntityRisk(entity *domain.Entity) float64 {
	if entity.KYCLevel >= 3 {
		return 0.2
	} else if entity.KYCLevel >= 1 {
		return 0.4
	}
	return 0.7
}

func (s *fcuService) calculateSanctionsRisk(entity *domain.Entity) float64 {
	switch entity.SanctionsStatus {
	case "clear":
		return 0.0
	case "potential_match":
		return 0.5
	case "confirmed_match":
		return 1.0
	default:
		return 0.0
	}
}

func (s *fcuService) calculatePEPRisk(entity *domain.Entity) float64 {
	switch entity.PEPStatus {
	case "clear":
		return 0.0
	case "potential":
		return 0.4
	case "confirmed":
		return 0.8
	default:
		return 0.0
	}
}

func (s *fcuService) calculateBlacklistRisk(entity *domain.Entity) float64 {
	switch entity.BlacklistStatus {
	case "clear":
		return 0.0
	case "suspected":
		return 0.5
	case "confirmed":
		return 1.0
	default:
		return 0.0
	}
}

func (s *fcuService) calculateRiskLevel(score int) string {
	if score >= 85 {
		return "critical"
	} else if score >= 70 {
		return "high"
	} else if score >= 50 {
		return "medium"
	} else if score >= 25 {
		return "low"
	}
	return "minimal"
}

// Network Analysis

type NetworkAnalysisRequest struct {
	EntityID      string `json:"entity_id" binding:"required"`
	MaxHops       int    `json:"max_hops,omitempty"`
	IncludeLabels bool   `json:"include_labels,omitempty"`
}

type TraceRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
	Direction     string `json:"direction,omitempty"` // "upstream", "downstream", "both"
	MaxHops       int    `json:"max_hops,omitempty"`
}

type TransactionFlow struct {
	TransactionID   string    `json:"transaction_id"`
	FromEntity      string    `json:"from_entity"`
	ToEntity        string    `json:"to_entity"`
	Amount          float64   `json:"amount"`
	Timestamp       time.Time `json:"timestamp"`
	RiskLevel       string    `json:"risk_level"`
}

type ConnectionInfo struct {
	EntityID        string  `json:"entity_id"`
	EntityType      string  `json:"entity_type"`
	TransactionCount int    `json:"transaction_count"`
	TotalVolume     float64 `json:"total_volume"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	Relationship    string  `json:"relationship"`
}

func (s *fcuService) AnalyzeNetwork(ctx context.Context, req *NetworkAnalysisRequest) (*domain.NetworkGraph, error) {
	maxHops := req.MaxHops
	if maxHops <= 0 {
		maxHops = s.config.NetworkAnalysis.MaxHops
	}

	// Get direct connections
	connections, _ := s.GetEntityConnections(ctx, req.EntityID)

	graph := &domain.NetworkGraph{
		CentralEntityID: req.EntityID,
		Nodes:           []domain.NetworkNode{},
		Edges:           []domain.NetworkEdge{},
		Patterns:        []domain.NetworkPattern{},
		RiskScore:       0,
		AnalysisDepth:   maxHops,
		AnalyzedAt:      s.now(),
	}

	// Add central node
	centralProfile, _ := s.config.Repo.GetRiskProfile(ctx, req.EntityID)
	graph.Nodes = append(graph.Nodes, domain.NetworkNode{
		ID:          req.EntityID,
		Type:        "entity",
		Label:       req.EntityID,
		RiskScore:   centralProfile.OverallScore,
		TransactionCount: 0,
		TotalVolume: 0,
	})

	// Add connected nodes
	for _, conn := range connections {
		connProfile, _ := s.config.Repo.GetRiskProfile(ctx, conn.EntityID)
		graph.Nodes = append(graph.Nodes, domain.NetworkNode{
			ID:            conn.EntityID,
			Type:          "entity",
			Label:         conn.EntityID,
			RiskScore:     connProfile.OverallScore,
			TransactionCount: conn.TransactionCount,
			TotalVolume:   conn.TotalVolume,
		})

		graph.Edges = append(graph.Edges, domain.NetworkEdge{
			SourceID:      req.EntityID,
			TargetID:      conn.EntityID,
			TransactionCount: conn.TransactionCount,
			TotalVolume:   conn.TotalVolume,
			Relationship:  conn.Relationship,
			FirstSeenAt:   conn.FirstSeen,
			LastSeenAt:    conn.LastSeen,
		})

		graph.RiskScore += connProfile.OverallScore
	}

	// Detect patterns
	graph.Patterns = s.detectNetworkPatterns(graph)

	if len(connections) > 0 {
		graph.RiskScore /= len(connections)
	}

	return graph, nil
}

func (s *fcuService) TraceTransactionFlow(ctx context.Context, req *TraceRequest) ([]*TransactionFlow, error) {
	tx, err := s.config.Repo.GetTransaction(ctx, req.TransactionID)
	if err != nil {
		return nil, err
	}

	var flows []*TransactionFlow
	maxHops := req.MaxHops
	if maxHops <= 0 {
		maxHops = 3
	}

	// Add initial transaction
	flows = append(flows, &TransactionFlow{
		TransactionID: tx.ID,
		FromEntity:    tx.SenderID,
		ToEntity:      tx.ReceiverID,
		Amount:        tx.Amount,
		Timestamp:     tx.Timestamp,
		RiskLevel:     s.calculateRiskLevel(tx.RiskScore),
	})

	// Trace upstream (from sender)
	if req.Direction == "" || req.Direction == "upstream" {
		prevTxs, _ := s.config.Repo.GetTransactionsByEntity(ctx, tx.SenderID, maxHops, 0)
		for _, prevTx := range prevTxs {
			flows = append(flows, &TransactionFlow{
				TransactionID: prevTx.ID,
				FromEntity:    prevTx.SenderID,
				ToEntity:      prevTx.ReceiverID,
				Amount:        prevTx.Amount,
				Timestamp:     prevTx.Timestamp,
				RiskLevel:     s.calculateRiskLevel(prevTx.RiskScore),
			})
		}
	}

	// Trace downstream (to receiver)
	if req.Direction == "" || req.Direction == "downstream" {
		nextTxs, _ := s.config.Repo.GetTransactionsByEntity(ctx, tx.ReceiverID, maxHops, 0)
		for _, nextTx := range nextTxs {
			flows = append(flows, &TransactionFlow{
				TransactionID: nextTx.ID,
				FromEntity:    nextTx.SenderID,
				ToEntity:      nextTx.ReceiverID,
				Amount:        nextTx.Amount,
				Timestamp:     nextTx.Timestamp,
				RiskLevel:     s.calculateRiskLevel(nextTx.RiskScore),
			})
		}
	}

	return flows, nil
}

func (s *fcuService) GetNetworkPatterns(ctx context.Context, entityID string) ([]domain.NetworkPattern, error) {
	graph, err := s.AnalyzeNetwork(ctx, &NetworkAnalysisRequest{EntityID: entityID})
	if err != nil {
		return nil, err
	}
	return graph.Patterns, nil
}

func (s *fcuService) GetEntityConnections(ctx context.Context, entityID string) ([]*ConnectionInfo, error) {
	sentTxs, _ := s.config.Repo.GetTransactionsByEntity(ctx, entityID, 100, 0)
	receivedTxs, _ := s.config.Repo.GetTransactionsByEntity(ctx, entityID, 100, 0)

	connectionMap := make(map[string]*ConnectionInfo)

	// Process sent transactions
	for _, tx := range sentTxs {
		if _, ok := connectionMap[tx.ReceiverID]; !ok {
			connectionMap[tx.ReceiverID] = &ConnectionInfo{
				EntityID:   tx.ReceiverID,
				EntityType: "entity",
				Relationship: "recipient",
			}
		}
		connectionMap[tx.ReceiverID].TransactionCount++
		connectionMap[tx.ReceiverID].TotalVolume += tx.Amount
		if connectionMap[tx.ReceiverID].FirstSeen.IsZero() {
			connectionMap[tx.ReceiverID].FirstSeen = tx.Timestamp
		}
		connectionMap[tx.ReceiverID].LastSeen = tx.Timestamp
	}

	// Process received transactions
	for _, tx := range receivedTxs {
		if _, ok := connectionMap[tx.SenderID]; !ok {
			connectionMap[tx.SenderID] = &ConnectionInfo{
				EntityID:   tx.SenderID,
				EntityType: "entity",
				Relationship: "sender",
			}
		}
		connectionMap[tx.SenderID].TransactionCount++
		connectionMap[tx.SenderID].TotalVolume += tx.Amount
		if connectionMap[tx.SenderID].FirstSeen.IsZero() {
			connectionMap[tx.SenderID].FirstSeen = tx.Timestamp
		}
		connectionMap[tx.SenderID].LastSeen = tx.Timestamp
	}

	var connections []*ConnectionInfo
	for _, conn := range connectionMap {
		connections = append(connections, conn)
	}

	return connections, nil
}

func (s *fcuService) detectNetworkPatterns(graph *domain.NetworkGraph) []domain.NetworkPattern {
	var patterns []domain.NetworkPattern

	// Detect fan-in pattern (many sources to one target)
	inDegree := make(map[string]int)
	for _, edge := range graph.Edges {
		inDegree[edge.TargetID]++
	}
	for nodeID, degree := range inDegree {
		if degree >= 5 {
			patterns = append(patterns, domain.NetworkPattern{
				Type:   "fan_in",
				Description: fmt.Sprintf("Node %s receives transactions from %d sources", nodeID, degree),
				Severity: "high",
				EntitiesInvolved: []string{graph.CentralEntityID, nodeID},
				Confidence: 0.8,
			})
		}
	}

	// Detect fan-out pattern (one source to many targets)
	outDegree := make(map[string]int)
	for _, edge := range graph.Edges {
		outDegree[edge.SourceID]++
	}
	for nodeID, degree := range outDegree {
		if degree >= 5 {
			patterns = append(patterns, domain.NetworkPattern{
				Type:   "fan_out",
				Description: fmt.Sprintf("Node %s sends transactions to %d destinations", nodeID, degree),
				Severity: "high",
				EntitiesInvolved: []string{graph.CentralEntityID, nodeID},
				Confidence: 0.8,
			})
		}
	}

	// Detect circular pattern (self-referential or loop)
	nodeSet := make(map[string]bool)
	for _, node := range graph.Nodes {
		nodeSet[node.ID] = true
	}
	for _, edge := range graph.Edges {
		if edge.SourceID == edge.TargetID {
			patterns = append(patterns, domain.NetworkPattern{
				Type:   "circular_transaction",
				Description: "Circular transaction detected (self-loop)",
				Severity:    "critical",
				EntitiesInvolved: []string{edge.SourceID},
				Confidence: 0.95,
			})
		}
	}

	return patterns
}

// Entity Management

func (s *fcuService) GetEntity(ctx context.Context, id string) (*domain.Entity, error) {
	return s.config.Repo.GetEntity(ctx, id)
}

func (s *fcuService) UpdateEntity(ctx context.Context, entity *domain.Entity) error {
	entity.UpdatedAt = s.now()
	return s.config.Repo.UpdateEntity(ctx, entity)
}

func (s *fcuService) GetEntityTransactions(ctx context.Context, entityID string, limit, offset int) ([]domain.Transaction, error) {
	return s.config.Repo.GetTransactionsByEntity(ctx, entityID, limit, offset)
}

func (s *fcuService) GetEntityAlerts(ctx context.Context, entityID string) ([]domain.Alert, error) {
	return s.config.Repo.GetAlerts(ctx, repository.AlertFilter{
		EntityID: entityID,
		Limit:    100,
	})
}

type AddToWatchlistRequest struct {
	EntityID   string `json:"entity_id" binding:"required"`
	EntityType string `json:"entity_type" binding:"required"`
	Reason     string `json:"reason" binding:"required"`
	RiskLevel  string `json:"risk_level" binding:"required"`
	AddedBy    string `json:"added_by" binding:"required"`
	ExpiryDate *time.Time `json:"expiry_date,omitempty"`
	Notes      *string `json:"notes,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

func (s *fcuService) AddToWatchlist(ctx context.Context, req *AddToWatchlistRequest) error {
	entry := &domain.WatchlistEntry{
		EntityID:   req.EntityID,
		EntityType: req.EntityType,
		Reason:     req.Reason,
		RiskLevel:  req.RiskLevel,
		AddedBy:    req.AddedBy,
		ExpiryDate: req.ExpiryDate,
		Notes:      req.Notes,
		Tags:       req.Tags,
	}
	return s.config.Repo.AddToWatchlist(ctx, entry)
}

func (s *fcuService) RemoveFromWatchlist(ctx context.Context, entityID string) error {
	// In production, implement actual removal
	return nil
}

func (s *fcuService) GetWatchlistStatus(ctx context.Context, entityID string) (*domain.WatchlistEntry, error) {
	return s.config.Repo.GetWatchlistStatus(ctx, entityID)
}

// Investigation

type StartInvestigationRequest struct {
	CaseID         string   `json:"case_id" binding:"required"`
	Type           string   `json:"type" binding:"required"`
	LeadInvestigator string `json:"lead_investigator" binding:"required"`
	TeamMembers    []string `json:"team_members,omitempty"`
	Scope          string   `json:"scope" binding:"required"`
	Objectives     []string `json:"objectives,omitempty"`
}

func (s *fcuService) GetInvestigation(ctx context.Context, id string) (*domain.Investigation, error) {
	// Simplified - in production, implement actual retrieval
	return nil, nil
}

func (s *fcuService) StartInvestigation(ctx context.Context, req *StartInvestigationRequest) (*domain.Investigation, error) {
	investigation := &domain.Investigation{
		ID:              generateID(),
		CaseID:          req.CaseID,
		Type:            req.Type,
		Status:          "active",
		LeadInvestigator: req.LeadInvestigator,
		TeamMembers:     req.TeamMembers,
		SubjectID:       "", // In production, fetch from case
		SubjectType:     "", // In production, fetch from case
		Scope:           req.Scope,
		Objectives:      req.Objectives,
		StartedAt:       s.now(),
		CreatedAt:       s.now(),
		UpdatedAt:       s.now(),
	}

	if err := s.config.Repo.CreateInvestigation(ctx, investigation); err != nil {
		return nil, err
	}

	return investigation, nil
}

type AddNoteRequest struct {
	InvestigationID string `json:"investigation_id" binding:"required"`
	NoteType        string `json:"note_type" binding:"required"`
	Content         string `json:"content" binding:"required"`
	AuthorID        string `json:"author_id" binding:"required"`
	AuthorName      string `json:"author_name" binding:"required"`
	Confidential    bool   `json:"confidential,omitempty"`
}

func (s *fcuService) AddInvestigationNote(ctx context.Context, req *AddNoteRequest) error {
	note := &domain.InvestigationNote{
		ID:              generateID(),
		InvestigationID: req.InvestigationID,
		NoteType:        req.NoteType,
		Content:         req.Content,
		AuthorID:        req.AuthorID,
		AuthorName:      req.AuthorName,
		Confidential:    req.Confidential,
		CreatedAt:       s.now(),
	}
	return s.config.Repo.AddInvestigationNote(ctx, note)
}

type AddTimelineRequest struct {
	CaseID      string `json:"case_id" binding:"required"`
	EventType   string `json:"event_type" binding:"required"`
	Description string `json:"description" binding:"required"`
	PerformedBy string `json:"performed_by" binding:"required"`
	EntityID    *string `json:"entity_id,omitempty"`
	Metadata    *map[string]interface{} `json:"metadata,omitempty"`
}

func (s *fcuService) AddTimelineEvent(ctx context.Context, req *AddTimelineRequest) error {
	var metadata json.RawMessage
	if req.Metadata != nil {
		data, _ := json.Marshal(*req.Metadata)
		metadata = data
	}

	event := &domain.TimelineEvent{
		ID:          generateID(),
		CaseID:      req.CaseID,
		EventType:   req.EventType,
		Description: req.Description,
		PerformedBy: req.PerformedBy,
		EntityID:    req.EntityID,
		Metadata:    &metadata,
		Timestamp:   s.now(),
	}
	return s.config.Repo.AddTimelineEvent(ctx, event)
}

// Reports

type ReportSummaryRequest struct {
	DateFrom *time.Time `json:"date_from,omitempty"`
	DateTo   *time.Time `json:"date_to,omitempty"`
}

type ReportSummary struct {
	TotalAlerts       int                    `json:"total_alerts"`
	AlertsBySeverity  map[string]int         `json:"alerts_by_severity"`
	AlertsByType      map[string]int         `json:"alerts_by_type"`
	TotalCases        int                    `json:"total_cases"`
	CasesByStatus     map[string]int         `json:"cases_by_status"`
	TotalSARs         int                    `json:"total_sars"`
	AverageRiskScore  float64                `json:"average_risk_score"`
	TopRiskEntities   []string               `json:"top_risk_entities"`
}

type TrendAnalysisRequest struct {
	Period string `json:"period"` // "day", "week", "month"
}

type TrendAnalysis struct {
	Period           string                  `json:"period"`
	AlertTrend       []TrendDataPoint        `json:"alert_trend"`
	RiskTrend        []TrendDataPoint        `json:"risk_trend"`
	TopAlertTypes    []struct {
		Type  string `json:"type"`
		Count int    `json:"count"`
	} `json:"top_alert_types"`
}

type TrendDataPoint struct {
	Date   time.Time `json:"date"`
	Count  int       `json:"count"`
	Score  float64   `json:"score,omitempty"`
}

type EffectivenessRequest struct {
	DateFrom *time.Time `json:"date_from,omitempty"`
	DateTo   *time.Time `json:"date_to,omitempty"`
}

type DetectionEffectiveness struct {
	Period              string  `json:"period"`
	TotalAlerts         int     `json:"total_alerts"`
	InvestigatedAlerts  int     `json:"investigated_alerts"`
	ConfirmedAlerts     int     `json:"confirmed_alerts"`
	FalsePositives      int     `json:"false_positives"`
	ConversionRate      float64 `json:"conversion_rate"`
	AverageResolutionTime float64 `json:"average_resolution_time_hours"`
}

type ComplianceReportRequest struct {
	RegulatoryBody string `json:"regulatory_body"`
	Period         string `json:"period"`
}

type ComplianceReport struct {
	RegulatoryBody      string `json:"regulatory_body"`
	Period              string `json:"period"`
	TotalSARsFiled      int    `json:"total_sars_filed"`
	OnTimeFilingRate    float64 `json:"on_time_filing_rate"`
	AverageFilingTime   float64 `json:"average_filing_time_hours"`
	RejectedReports     int     `json:"rejected_reports"`
	PendingReports      int     `json:"pending_reports"`
}

func (s *fcuService) GetReportSummary(ctx context.Context, req *ReportSummaryRequest) (*ReportSummary, error) {
	// Simplified implementation
	return &ReportSummary{
		TotalAlerts:      0,
		AlertsBySeverity: make(map[string]int),
		AlertsByType:     make(map[string]int),
		TotalCases:       0,
		CasesByStatus:    make(map[string]int),
		TotalSARs:        0,
		AverageRiskScore: 0,
		TopRiskEntities:  []string{},
	}, nil
}

func (s *fcuService) GetTrendAnalysis(ctx context.Context, req *TrendAnalysisRequest) (*TrendAnalysis, error) {
	return &TrendAnalysis{
		Period:      req.Period,
		AlertTrend:  []TrendDataPoint{},
		RiskTrend:   []TrendDataPoint{},
		TopAlertTypes: []struct {
			Type  string `json:"type"`
			Count int    `json:"count"`
		}{},
	}, nil
}

func (s *fcuService) GetDetectionEffectiveness(ctx context.Context, req *EffectivenessRequest) (*DetectionEffectiveness, error) {
	return &DetectionEffectiveness{
		Period:               "30d",
		TotalAlerts:          0,
		InvestigatedAlerts:   0,
		ConfirmedAlerts:      0,
		FalsePositives:       0,
		ConversionRate:       0,
		AverageResolutionTime: 0,
	}, nil
}

func (s *fcuService) GetComplianceReport(ctx context.Context, req *ComplianceReportRequest) (*ComplianceReport, error) {
	return &ComplianceReport{
		RegulatoryBody:     req.RegulatoryBody,
		Period:             req.Period,
		TotalSARsFiled:     0,
		OnTimeFilingRate:   0,
		AverageFilingTime:  0,
		RejectedReports:    0,
		PendingReports:     0,
	}, nil
}

// Helper functions

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func stringPtr(s string) *string {
	return &s
}

func filterNilString(s *string) []string {
	if s == nil {
		return []string{}
	}
	return []string{*s}
}
