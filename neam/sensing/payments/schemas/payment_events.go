// NEAM Sensing Layer - Payment Event Schemas
// Standardized schemas for payment data using Protocol Buffers structure

package schemas

import (
	"time"
)

// PaymentType represents the type of payment transaction
type PaymentType string

const (
	PaymentTypeTransfer PaymentType = "transfer"
	PaymentTypeUPI      PaymentType = "upi"
	PaymentTypeCard     PaymentType = "card"
	PaymentTypeWallet   PaymentType = "wallet"
	PaymentTypeCash     PaymentType = "cash"
	PaymentTypeCheque   PaymentType = "cheque"
	PaymentTypeNetBank  PaymentType = "net_banking"
)

// Channel represents the payment channel used
type Channel string

const (
	ChannelBank  Channel = "bank"
	ChannelUPI   Channel = "upi"
	ChannelPOS   Channel = "pos"
	ChannelATM   Channel = "atm"
	ChannelWallet Channel = "wallet"
	ChannelOnline Channel = "online"
	ChannelMobile Channel = "mobile"
)

// PaymentStatus represents the status of a payment transaction
type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusProcessing PaymentStatus = "processing"
	StatusCompleted PaymentStatus = "completed"
	StatusFailed    PaymentStatus = "failed"
	StatusReversed  PaymentStatus = "reversed"
	StatusRefunded  PaymentStatus = "refunded"
)

// PaymentEvent represents a standardized payment transaction event
type PaymentEvent struct {
	// Core identification
	TransactionID string `json:"transaction_id" bson:"transaction_id"`
	OriginalRef   string `json:"original_ref,omitempty" bson:"original_ref,omitempty"`
	
	// Timestamp
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	SettlementDate *time.Time `json:"settlement_date,omitempty" bson:"settlement_date,omitempty"`
	
	// Amount information
	Amount   float64 `json:"amount" bson:"amount"`
	Currency string  `json:"currency" bson:"currency"`
	
	// Payment type
	PaymentType PaymentType `json:"payment_type" bson:"payment_type"`
	Channel     Channel     `json:"channel" bson:"channel"`
	
	// Merchant information
	MerchantID         string `json:"merchant_id,omitempty" bson:"merchant_id,omitempty"`
	MerchantName       string `json:"merchant_name,omitempty" bson:"merchant_name,omitempty"`
	MerchantCategory   string `json:"merchant_category" bson:"merchant_category"`
	MerchantCategoryCode string `json:"merchant_category_code,omitempty" bson:"merchant_category_code,omitempty"`
	
	// Location
	Region       string  `json:"region" bson:"region"`
	District     string  `json:"district,omitempty" bson:"district,omitempty"`
	PinCode      string  `json:"pin_code,omitempty" bson:"pin_code,omitempty"`
	Latitude     float64 `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty" bson:"longitude,omitempty"`
	
	// Source information
	SourceBank      string `json:"source_bank" bson:"source_bank"`
	SourceAccount   string `json:"source_account,omitempty" bson:"source_account,omitempty"`
	DestBank        string `json:"dest_bank,omitempty" bson:"dest_bank,omitempty"`
	DestAccount     string `json:"dest_account,omitempty" bson:"dest_account,omitempty"`
	
	// Status
	Status PaymentStatus `json:"status" bson:"status"`
	
	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	
	// Processing info
	ProcessingFee  float64 `json:"processing_fee,omitempty" bson:"processing_fee,omitempty"`
	ProcessingTime int64   `json:"processing_time_ms,omitempty" bson:"processing_time_ms,omitempty"`
}

// PaymentAggregation represents aggregated payment statistics
type PaymentAggregation struct {
	Region              string    `json:"region" bson:"region"`
	Period              string    `json:"period" bson:"period"`
	StartTime           time.Time `json:"start_time" bson:"start_time"`
	EndTime             time.Time `json:"end_time" bson:"end_time"`
	
	// Volume metrics
	TotalTransactions   int64   `json:"total_transactions" bson:"total_transactions"`
	TotalAmount         float64 `json:"total_amount" bson:"total_amount"`
	AverageAmount       float64 `json:"average_amount" bson:"average_amount"`
	MaxAmount           float64 `json:"max_amount" bson:"max_amount"`
	MinAmount           float64 `json:"min_amount" bson:"min_amount"`
	
	// Count by type
	ByPaymentType map[PaymentType]int64 `json:"by_payment_type" bson:"by_payment_type"`
	ByChannel     map[Channel]int64     `json:"by_channel" bson:"by_channel"`
	ByStatus      map[PaymentStatus]int64 `json:"by_status" bson:"by_status"`
	
	// By category
	ByMerchantCategory map[string]int64 `json:"by_merchant_category" bson:"by_merchant_category"`
	
	// Velocity metrics
	TransactionsPerSecond float64 `json:"tps" bson:"tps"`
	AmountPerSecond       float64 `json:"aps" bson:"aps"`
}

// PaymentVelocity represents payment velocity metrics
type PaymentVelocity struct {
	Region    string    `json:"region" bson:"region"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	
	// Current metrics
	CurrentTPS      float64 `json:"current_tps" bson:"current_tps"`
	CurrentAPS      float64 `json:"current_aps" bson:"current_aps"`
	ActiveAccounts  int64   `json:"active_accounts" bson:"active_accounts"`
	
	// Historical metrics
	AverageTPS24h   float64 `json:"avg_tps_24h" bson:"avg_tps_24h"`
	AverageTPS7d    float64 `json:"avg_tps_7d" bson:"avg_tps_7d"`
	AverageTPS30d   float64 `json:"avg_tps_30d" bson:"avg_tps_30d"`
	
	// Trend
	TrendDirection string  `json:"trend" bson:"trend"` // "up", "down", "stable"
	TrendPercent   float64 `json:"trend_percent" bson:"trend_percent"`
	
	// Anomaly indicators
	IsAnomalous   bool    `json:"is_anomalous" bson:"is_anomalous"`
	AnomalyScore  float64 `json:"anomaly_score" bson:"anomaly_score"`
	AnomalyType   string  `json:"anomaly_type,omitempty" bson:"anomaly_type,omitempty"`
}

// PaymentSummary represents a summary of payment activity
type PaymentSummary struct {
	Region           string    `json:"region" bson:"region"`
	ReportDate       time.Time `json:"report_date" bson:"report_date"`
	
	// Daily totals
	DailyTransactions int64   `json:"daily_transactions" bson:"daily_transactions"`
	DailyAmount       float64 `json:"daily_amount" bson:"daily_amount"`
	
	// Comparison with previous day
	DailyChangePercent float64 `json:"daily_change_percent" bson:"daily_change_percent"`
	DailyTrend         string  `json:"daily_trend" bson:"daily_trend"` // "up", "down", "stable"
	
	// Comparison with previous week
	WeeklyChangePercent float64 `json:"weekly_change_percent" bson:"weekly_change_percent"`
	WeeklyTrend         string  `json:"weekly_trend" bson:"weekly_trend"`
	
	// Comparison with previous month
	MonthlyChangePercent float64 `json:"monthly_change_percent" bson:"monthly_change_percent"`
	MonthlyTrend         string  `json:"monthly_trend" bson:"monthly_trend"`
	
	// Top categories
	TopCategories []CategorySummary `json:"top_categories" bson:"top_categories"`
	
	// Top regions
	TopRegions []RegionSummary `json:"top_regions" bson:"top_regions"`
}

// CategorySummary represents summary for a merchant category
type CategorySummary struct {
	Category           string  `json:"category" bson:"category"`
	Transactions       int64   `json:"transactions" bson:"transactions"`
	Amount             float64 `json:"amount" bson:"amount"`
	PercentageOfTotal  float64 `json:"percentage_of_total" bson:"percentage_of_total"`
	ChangeFromPrevious float64 `json:"change_from_previous" bson:"change_from_previous"`
}

// RegionSummary represents summary for a region
type RegionSummary struct {
	Region      string  `json:"region" bson:"region"`
	Transactions int64   `json:"transactions" bson:"transactions"`
	Amount      float64 `json:"amount" bson:"amount"`
	PercentageOfTotal float64 `json:"percentage_of_total" bson:"percentage_of_total"`
}

// PaymentFilter represents filter criteria for payment queries
type PaymentFilter struct {
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Regions      []string   `json:"regions,omitempty"`
	PaymentTypes []PaymentType `json:"payment_types,omitempty"`
	Channels     []Channel  `json:"channels,omitempty"`
	Status       []PaymentStatus `json:"status,omitempty"`
	MinAmount    *float64   `json:"min_amount,omitempty"`
	MaxAmount    *float64   `json:"max_amount,omitempty"`
	MerchantCategories []string `json:"merchant_categories,omitempty"`
	
	// Pagination
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	
	// Sorting
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"` // "asc" or "desc"
}

// PaymentQueryResult represents the result of a payment query
type PaymentQueryResult struct {
	Events      []PaymentEvent `json:"events" bson:"events"`
	TotalCount  int64          `json:"total_count" bson:"total_count"`
	QueryFilter PaymentFilter  `json:"query_filter" bson:"query_filter"`
	
	// Aggregation if requested
	Aggregation *PaymentAggregation `json:"aggregation,omitempty" bson:"aggregation,omitempty"`
	
	// Query metadata
	QueryTime   time.Duration `json:"query_time" bson:"query_time"`
	ScannedCount int64        `json:"scanned_count" bson:"scanned_count"`
}

// PaymentAlert represents an alert generated from payment data
type PaymentAlert struct {
	ID          string        `json:"id" bson:"id"`
	AlertType   string        `json:"alert_type" bson:"alert_type"` // "fraud", "anomaly", "threshold", "pattern"
	Severity    string        `json:"severity" bson:"severity"` // "low", "medium", "high", "critical"
	Region      string        `json:"region" bson:"region"`
	
	Description string        `json:"description" bson:"description"`
	Details     map[string]interface{} `json:"details" bson:"details"`
	
	Timestamp   time.Time     `json:"timestamp" bson:"timestamp"`
	
	// Related transactions
	TransactionIDs []string   `json:"transaction_ids" bson:"transaction_ids"`
	
	// Status
	Status    string     `json:"status" bson:"status"` // "new", "investigating", "resolved", "dismissed"
	ResolvedAt *time.Time `json:"resolved_at,omitempty" bson:"resolved_at,omitempty"`
	ResolvedBy string     `json:"resolved_by,omitempty" bson:"resolved_by,omitempty"`
	
	// Metadata
	CreatedBy string `json:"created_by" bson:"created_by"` // "system" or user ID
	Tags      []string `json:"tags" bson:"tags"`
}

// ToMap converts PaymentEvent to map for storage
func (pe *PaymentEvent) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"transaction_id":     pe.TransactionID,
		"timestamp":          pe.Timestamp,
		"amount":             pe.Amount,
		"currency":           pe.Currency,
		"payment_type":       pe.PaymentType,
		"channel":            pe.Channel,
		"merchant_category":  pe.MerchantCategory,
		"region":             pe.Region,
		"source_bank":        pe.SourceBank,
		"status":             pe.Status,
		"metadata":           pe.Metadata,
	}
}

// Validate validates the payment event
func (pe *PaymentEvent) Validate() error {
	if pe.TransactionID == "" {
		return ErrMissingTransactionID
	}
	if pe.Timestamp.IsZero() {
		return ErrMissingTimestamp
	}
	if pe.Amount <= 0 {
		return ErrInvalidAmount
	}
	if pe.Currency == "" {
		return ErrMissingCurrency
	}
	if pe.Region == "" {
		return ErrMissingRegion
	}
	return nil
}

// Custom errors
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

var (
	ErrMissingTransactionID = &ValidationError{"transaction_id is required"}
	ErrMissingTimestamp     = &ValidationError{"timestamp is required"}
	ErrInvalidAmount        = &ValidationError{"amount must be positive"}
	ErrMissingCurrency      = &ValidationError{"currency is required"}
	ErrMissingRegion        = &ValidationError{"region is required"}
)

// Avro Schema Definition for PaymentEvent
// This can be used to generate actual Avro schemas
const PaymentEventAvroSchema = `{
  "type": "record",
  "name": "PaymentEvent",
  "namespace": "neam.sensing.payments",
  "fields": [
    {"name": "transaction_id", "type": "string"},
    {"name": "timestamp", "type": {"type": "long", "logicalType": "timestamp-millis"}},
    {"name": "amount", "type": "double"},
    {"name": "currency", "type": "string"},
    {"name": "payment_type", "type": {"type": "enum", "name": "PaymentType", "symbols": ["transfer", "upi", "card", "wallet", "cash", "cheque", "net_banking"]}},
    {"name": "channel", "type": {"type": "enum", "name": "Channel", "symbols": ["bank", "upi", "pos", "atm", "wallet", "online", "mobile"]}},
    {"name": "merchant_category", "type": "string"},
    {"name": "region", "type": "string"},
    {"name": "source_bank", "type": "string"},
    {"name": "status", "type": {"type": "enum", "name": "PaymentStatus", "symbols": ["pending", "processing", "completed", "failed", "reversed", "refunded"]}},
    {"name": "metadata", "type": {"type": "map", "values": "string"}}
  ]
}`

// Protobuf Definition for PaymentEvent
const PaymentEventProtoDefinition = `
syntax = "proto3";

package neam.sensing.payments;

option go_package = "neam-platform/sensing/payments/schemas";

import "google/protobuf/timestamp.proto";

message PaymentEvent {
  string transaction_id = 1;
  google.protobuf.Timestamp timestamp = 2;
  double amount = 3;
  string currency = 4;
  PaymentType payment_type = 5;
  Channel channel = 6;
  string merchant_category = 7;
  string region = 8;
  string source_bank = 9;
  PaymentStatus status = 10;
  map<string, string> metadata = 11;
}

enum PaymentType {
  PAYMENT_TYPE_UNSPECIFIED = 0;
  TRANSFER = 1;
  UPI = 2;
  CARD = 3;
  WALLET = 4;
  CASH = 5;
  CHEQUE = 6;
  NET_BANKING = 7;
}

enum Channel {
  CHANNEL_UNSPECIFIED = 0;
  BANK = 1;
  UPI_CH = 2;
  POS = 3;
  ATM = 4;
  WALLET_CH = 5;
  ONLINE = 6;
  MOBILE = 7;
}

enum PaymentStatus {
  PAYMENT_STATUS_UNSPECIFIED = 0;
  PENDING = 1;
  PROCESSING = 2;
  COMPLETED = 3;
  FAILED = 4;
  REVERSED = 5;
  REFUNDED = 6;
}
`
