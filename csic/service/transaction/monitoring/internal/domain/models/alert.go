package models

import (
	"time"

	"github.com/google/uuid"
)

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	SeverityCritical AlertSeverity = "critical"
	SeverityHigh     AlertSeverity = "high"
	SeverityMedium   AlertSeverity = "medium"
	SeverityLow      AlertSeverity = "low"
	SeverityInfo     AlertSeverity = "info"
)

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeSanctionsMatch      AlertType = "sanctions_match"
	AlertTypeHighRiskScore       AlertType = "high_risk_score"
	AlertTypeRapidMovement       AlertType = "rapid_movement"
	AlertTypeStructuring         AlertType = "structuring"
	AlertTypeLayering            AlertType = "layering"
	AlertTypeSmurfing            AlertType = "smurfing"
	AlertTypeUnusualPattern      AlertType = "unusual_pattern"
	AlertTypeVelocityBreach      AlertType = "velocity_breach"
	AlertTypeNewHighRiskEntity   AlertType = "new_high_risk_entity"
	AlertTypeClusterRiskIncrease AlertType = "cluster_risk_increase"
	AlertTypeTransactionAnomaly  AlertType = "transaction_anomaly"
	AlertTypeBlacklistMatch      AlertType = "blacklist_match"
	AlertTypeWhitelistException  AlertType = "whitelist_exception"
)

// AlertStatus represents alert investigation status
type AlertStatus string

const (
	AlertStatusNew          AlertStatus = "new"
	AlertStatusInProgress   AlertStatus = "in_progress"
	AlertStatusUnderReview  AlertStatus = "under_review"
	AlertStatusEscalated    AlertStatus = "escalated"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusDismissed    AlertStatus = "dismissed"
	AlertStatusFalsePositive AlertStatus = "false_positive"
)

// Alert represents a compliance alert
type Alert struct {
	ID              string        `json:"id" db:"id"`
	AlertType       AlertType     `json:"alert_type" db:"alert_type"`
	Severity        AlertSeverity `json:"severity" db:"severity"`
	Status          AlertStatus   `json:"status" db:"status"`
	TargetType      string        `json:"target_type" db:"target_type"` // wallet, transaction, cluster
	TargetID        string        `json:"target_id" db:"target_id"`
	TargetValue     string        `json:"target_value" db:"target_value"`
	Title           string        `json:"title" db:"title"`
	Description     string        `json:"description" db:"description"`
	RuleID          *string       `json:"rule_id" db:"rule_id"`
	RuleName        string        `json:"rule_name" db:"rule_name"`
	TriggeredAt     time.Time     `json:"triggered_at" db:"triggered_at"`
	ResolvedAt      *time.Time    `json:"resolved_at,omitempty" db:"resolved_at"`
	AssignedTo      *string       `json:"assigned_to" db:"assigned_to"`
	AssignedTeam    string        `json:"assigned_team" db:"assigned_team"`
	CaseID          *string       `json:"case_id" db:"case_id"`
	Score           float64       `json:"score" db:"score"`
	Evidence        []string      `json:"evidence" db:"-"`
	Notes           []AlertNote   `json:"notes,omitempty" db:"-"`
	RelatedAlerts   []string      `json:"related_alerts" db:"-"`
	Metadata        map[string]interface{} `json:"metadata" db:"-"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}

// AlertNote represents a note added to an alert
type AlertNote struct {
	ID        string    `json:"id" db:"id"`
	AlertID   string    `json:"alert_id" db:"alert_id"`
	Content   string    `json:"content" db:"content"`
	Author    string    `json:"author" db:"author"`
	AuthorRole string   `json:"author_role" db:"author_role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AlertEvidence represents evidence attached to an alert
type AlertEvidence struct {
	ID          string    `json:"id" db:"id"`
	AlertID     string    `json:"alert_id" db:"alert_id"`
	EvidenceType string   `json:"evidence_type" db:"evidence_type"` // transaction, address, cluster, external
	ReferenceID string    `json:"reference_id" db:"reference_id"`
	Description string    `json:"description" db:"description"`
	Data        string    `json:"data" db:"data"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// AlertRule represents a rule that can trigger alerts
type AlertRule struct {
	ID           string                 `json:"id" db:"id"`
	Name         string                 `json:"name" db:"name"`
	Category     string                 `json:"category" db:"category"`
	Description  string                 `json:"description" db:"description"`
	AlertType    AlertType              `json:"alert_type" db:"alert_type"`
	Severity     AlertSeverity          `json:"severity" db:"severity"`
	Enabled      bool                   `json:"enabled" db:"enabled"`
	Conditions   map[string]interface{} `json:"conditions" db:"-"`
	Parameters   map[string]interface{} `json:"parameters" db:"-"`
	RiskWeight   float64                `json:"risk_weight" db:"risk_weight"`
	CooldownSecs int                    `json:"cooldown_secs" db:"cooldown_secs"`
	LastTriggered *time.Time            `json:"last_triggered,omitempty" db:"last_triggered"`
	TriggerCount int64                  `json:"trigger_count" db:"trigger_count"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// AlertStatistics represents aggregated alert statistics
type AlertStatistics struct {
	TotalAlerts        int64            `json:"total_alerts"`
	AlertsBySeverity   map[AlertSeverity]int64 `json:"alerts_by_severity"`
	AlertsByType       map[AlertType]int64    `json:"alerts_by_type"`
	AlertsByStatus     map[AlertStatus]int64   `json:"alerts_by_status"`
	AverageResolutionTime float64       `json:"average_resolution_time_hours"`
	FalsePositiveRate  float64          `json:"false_positive_rate"`
	TopRuleTriggers    []RuleTriggerStat `json:"top_rule_triggers"`
}

// RuleTriggerStat represents statistics for a triggered rule
type RuleTriggerStat struct {
	RuleID      string `json:"rule_id"`
	RuleName    string `json:"rule_name"`
	TriggerCount int64 `json:"trigger_count"`
	LastTriggered time.Time `json:"last_triggered"`
}

// NewAlert creates a new alert with generated ID
func NewAlert(alertType AlertType, severity AlertSeverity, targetType, targetID, title string) *Alert {
	return &Alert{
		ID:          uuid.New().String(),
		AlertType:   alertType,
		Severity:    severity,
		Status:      AlertStatusNew,
		TargetType:  targetType,
		TargetID:    targetID,
		Title:       title,
		TriggeredAt: time.Now(),
		Evidence:    []string{},
		Notes:       []AlertNote{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// AddNote adds a note to the alert
func (a *Alert) AddNote(content, author, authorRole string) {
	note := AlertNote{
		ID:         uuid.New().String(),
		AlertID:    a.ID,
		Content:    content,
		Author:     author,
		AuthorRole: authorRole,
		CreatedAt:  time.Now(),
	}
	a.Notes = append(a.Notes, note)
	a.UpdatedAt = time.Now()
}

// Assign assigns the alert to a user or team
func (a *Alert) Assign(assignee, team string) {
	a.AssignedTo = &assignee
	a.AssignedTeam = team
	a.Status = AlertStatusInProgress
	a.UpdatedAt = time.Now()
}

// Resolve marks the alert as resolved
func (a *Alert) Resolve() {
	now := time.Now()
	a.ResolvedAt = &now
	a.Status = AlertStatusResolved
	a.UpdatedAt = now
}

// Dismiss marks the alert as dismissed
func (a *Alert) Dismiss() {
	now := time.Now()
	a.ResolvedAt = &now
	a.Status = AlertStatusDismissed
	a.UpdatedAt = now
}

// SanctionsMatchAlert represents a sanctions screening alert
type SanctionsMatchAlert struct {
	AlertID          string    `json:"alert_id"`
	WalletAddress    string    `json:"wallet_address"`
	Network          Network   `json:"network"`
	MatchedList      string    `json:"matched_list"` // OFAC, UN, EU
	MatchedEntity    string    `json:"matched_entity"`
	MatchType        string    `json:"match_type"` // direct, indirect, owner
	MatchScore       float64   `json:"match_score"`
	SanctionedEntity string    `json:"sanctioned_entity"`
	Program          string    `json:"program"` // SDN, CONV, etc.
	Comments         string    `json:"comments"`
	CreatedAt        time.Time `json:"created_at"`
}
