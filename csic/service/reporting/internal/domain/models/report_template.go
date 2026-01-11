package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReportType represents the type of regulatory report
type ReportType string

const (
	ReportTypePeriodic          ReportType = "periodic"
	ReportTypeSAR               ReportType = "suspicious_activity"
	ReportTypeCompliance        ReportType = "compliance"
	ReportTypeAudit             ReportType = "audit"
	ReportTypeRiskAssessment    ReportType = "risk_assessment"
	ReportTypeEnforcementAction ReportType = "enforcement_action"
	ReportTypeRegulatoryFiling  ReportType = "regulatory_filing"
)

// ReportStatus represents the current status of a generated report
type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusProcessing ReportStatus = "processing"
	ReportStatusCompleted  ReportStatus = "completed"
	ReportStatusFailed     ReportStatus = "failed"
	ReportStatusArchived   ReportStatus = "archived"
)

// ReportFormat represents the output format of a report
type ReportFormat string

const (
	ReportFormatPDF   ReportFormat = "pdf"
	ReportFormatCSV   ReportFormat = "csv"
	ReportFormatXLSX  ReportFormat = "xlsx"
	ReportFormatJSON  ReportFormat = "json"
	ReportFormatXBRL  ReportFormat = "xbrl"
	ReportFormatXML   ReportFormat = "xml"
)

// ReportFrequency defines how often a report should be generated
type ReportFrequency string

const (
	FrequencyDaily   ReportFrequency = "daily"
	FrequencyWeekly  ReportFrequency = "weekly"
	FrequencyMonthly ReportFrequency = "monthly"
	FrequencyQuarterly ReportFrequency = "quarterly"
	FrequencyAnnually ReportFrequency = "annually"
	FrequencyOnDemand ReportFrequency = "on_demand"
)

// ReportTemplate represents a reusable report configuration
type ReportTemplate struct {
	ID                  uuid.UUID       `json:"id" db:"id"`
	Name                string          `json:"name" db:"name"`
	Description         string          `json:"description" db:"description"`
	ReportType          ReportType      `json:"report_type" db:"report_type"`
	RegulatorID         uuid.UUID       `json:"regulator_id" db:"regulator_id"`
	Version             string          `json:"version" db:"version"`
	TemplateSchema      TemplateSchema  `json:"template_schema" db:"template_schema"`
	GenerationRules     GenerationRules `json:"generation_rules" db:"generation_rules"`
	OutputFormats       []ReportFormat  `json:"output_formats" db:"output_formats"`
	Frequency           ReportFrequency `json:"frequency" db:"frequency"`
	CronSchedule        string          `json:"cron_schedule" db:"cron_schedule"`
	IsActive            bool            `json:"is_active" db:"is_active"`
	CreatedBy           string          `json:"created_by" db:"created_by"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

// TemplateSchema defines the structure and fields of a report
type TemplateSchema struct {
	Title           string                `json:"title"`
	Version         string                `json:"version"`
	Sections        []ReportSection       `json:"sections"`
	DataSources     []DataSourceConfig    `json:"data_sources"`
	Transformations []TransformationRule  `json:"transformations"`
	ValidationRules ValidationRules       `json:"validation_rules"`
}

// ReportSection defines a section within a report
type ReportSection struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Type        string              `json:"type"` // table, chart, text, summary
	Fields      []SectionField      `json:"fields"`
	Order       int                 `json:"order"`
	Conditional ConditionalRule     `json:"conditional"`
}

// SectionField represents a field within a report section
type SectionField struct {
	Name        string      `json:"name"`
	Label       string      `json:"label"`
	DataType    string      `json:"data_type"`
	Source      string      `json:"source"`
	Format      string      `json:"format"` // currency, percentage, date, number
	Calculation string      `json:"calculation"` // sum, average, count, etc.
	Aggregate   string      `json:"aggregate"`
	Hidden      bool        `json:"hidden"`
	Order       int         `json:"order"`
}

// DataSourceConfig defines where report data comes from
type DataSourceConfig struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // database, api, kafka, calculation
	Endpoint string `json:"endpoint"`
	Query    string `json:"query"`
	Table    string `json:"table"`
	Fields   []string `json:"fields"`
	Filters  []FilterRule `json:"filters"`
}

// FilterRule defines a filter for data retrieval
type FilterRule struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, in, like
	Value    interface{} `json:"value"`
}

// TransformationRule defines how data should be transformed
type TransformationRule struct {
	SourceField string `json:"source_field"`
	TargetField string `json:"target_field"`
	Transform   string `json:"transform"` // uppercase, lowercase, format_date, calculate
	Parameters  map[string]interface{} `json:"parameters"`
}

// ValidationRules defines validation rules for report data
type ValidationRules struct {
	RequiredFields []string            `json:"required_fields"`
	CustomRules    []CustomValidation  `json:"custom_rules"`
	ThresholdRules []ThresholdRule     `json:"threshold_rules"`
}

// CustomValidation defines a custom validation rule
type CustomValidation struct {
	Name       string   `json:"name"`
	Expression string   `json:"expression"`
	Message    string   `json:"message"`
	Fields     []string `json:"fields"`
}

// ThresholdRule defines a threshold check
type ThresholdRule struct {
	Field       string  `json:"field"`
	MinValue    float64 `json:"min_value"`
	MaxValue    float64 `json:"max_value"`
	WarningOnly bool    `json:"warning_only"`
}

// ConditionalRule defines conditional display rules
type ConditionalRule struct {
	Condition   string                 `json:"condition"`
	Field       string                 `json:"field"`
	Operator    string                 `json:"operator"`
	Value       interface{}            `json:"value"`
	Actions     []ConditionalAction    `json:"actions"`
}

// ConditionalAction defines an action for a conditional rule
type ConditionalAction struct {
	Action  string                 `json:"action"` // show, hide, highlight
	Targets []string               `json:"targets"`
	Styles  map[string]interface{} `json:"styles"`
}

// GenerationRules define how the report should be generated
type GenerationRules struct {
	DataRange      DataRangeConfig     `json:"data_range"`
	Aggregation    AggregationConfig   `json:"aggregation"`
	Sorting        []SortingRule       `json:"sorting"`
	Pagination     PaginationConfig    `json:"pagination"`
	IncludeCharts  bool                `json:"include_charts"`
	IncludeSummary bool                `json:"include_summary"`
	Signatures     []SignatureConfig   `json:"signatures"`
	Watermark      WatermarkConfig     `json:"watermark"`
}

// DataRangeConfig defines the date range for data collection
type DataRangeConfig struct {
	Type        string `json:"type"` // fixed, relative, dynamic
	StartOffset string `json:"start_offset"` // e.g., "-30d", "-1m"
	EndOffset   string `json:"end_offset"` // e.g., "now", "-1d"
	FixedStart  string `json:"fixed_start"` // ISO date
	FixedEnd    string `json:"fixed_end"` // ISO date
}

// AggregationConfig defines how data should be aggregated
type AggregationConfig struct {
	GroupBy    []string               `json:"group_by"`
	Aggregates []AggregateField       `json:"aggregates"`
	DateTrunc  string                 `json:"date_trunc"` // day, week, month, quarter, year
}

// AggregateField defines an aggregation operation
type AggregateField struct {
	Field     string `json:"field"`
	Operation string `json:"operation"` // sum, avg, min, max, count, distinct
	Alias     string `json:"alias"`
}

// SortingRule defines sorting for report data
type SortingRule struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // asc, desc
	Priority  int    `json:"priority"`
}

// PaginationConfig defines pagination settings
type PaginationConfig struct {
	Enabled   bool `json:"enabled"`
	PageSize  int  `json:"page_size"`
	MaxPages  int  `json:"max_pages"`
}

// SignatureConfig defines where signatures should appear
type SignatureConfig struct {
	Position   string `json:"position"` // bottom, top, each_page
	SignerRole string `json:"signer_role"`
	SignerName string `json:"signer_name"`
	Title      string `json:"title"`
}

// WatermarkConfig defines watermark settings for the report
type WatermarkConfig struct {
	Enabled  bool   `json:"enabled"`
	Text     string `json:"text"`
	Opacity  int    `json:"opacity"` // 0-100
	Angle    int    `json:"angle"` // 0-360
	FontSize int    `json:"font_size"`
}
