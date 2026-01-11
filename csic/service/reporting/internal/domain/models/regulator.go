package domain

import (
	"time"

	"github.com/google/uuid"
)

// RegulatorType represents the type of regulatory body
type RegulatorType string

const (
	RegulatorTypeCentralBank    RegulatorType = "central_bank"
	RegulatorTypeSecurities     RegulatorType = "securities_commission"
	RegulatorTypeFinancial      RegulatorType = "financial_conduct_authority"
	RegulatorTypeAML            RegulatorType = "aml_supervisor"
	RegulatorTypeTax            RegulatorType = "tax_authority"
	RegulatorTypeDataProtection RegulatorType = "data_protection"
	RegulatorTypeOther          RegulatorType = "other"
)

// Regulator represents a regulatory authority
type Regulator struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	Name              string          `json:"name" db:"name"`
	OfficialName      string          `json:"official_name" db:"official_name"`
	Type              RegulatorType   `json:"type" db:"type"`
	Jurisdiction      string          `json:"jurisdiction" db:"jurisdiction"`
	Region            string          `json:"region" db:"region"`
	
	// Contact information
	ContactEmail      string          `json:"contact_email" db:"contact_email"`
	ContactPhone      string          `json:"contact_phone" db:"contact_phone"`
	Website           string          `json:"website" db:"website"`
	SubmissionEmail   string          `json:"submission_email" db:"submission_email"`
	
	// Address
	AddressLine1      string          `json:"address_line_1" db:"address_line_1"`
	AddressLine2      string          `json:"address_line_2" db:"address_line_2"`
	City              string          `json:"city" db:"city"`
	State             string          `json:"state" db:"state"`
	PostalCode        string          `json:"postal_code" db:"postal_code"`
	Country           string          `json:"country" db:"country"`
	
	// API Configuration
	APIEndpoint       string          `json:"api_endpoint" db:"api_endpoint"`
	APIAuthType       string          `json:"api_auth_type" db:"api_auth_type"` // oauth, api_key, basic, none
	APIConfig         map[string]string `json:"api_config" db:"api_config"`
	
	// Supported report types and formats
	SupportedTypes    []ReportType    `json:"supported_types" db:"supported_types"`
	SupportedFormats  []ReportFormat  `json:"supported_formats" db:"supported_formats"`
	XBRLTaxonomy      string          `json:"xbrl_taxonomy" db:"xbrl_taxonomy"`
	
	// Requirements
	FilingRequirements string         `json:"filing_requirements" db:"filing_requirements"`
	ValidationRules   string         `json:"validation_rules" db:"validation_rules"`
	FeeStructure      string         `json:"fee_structure" db:"fee_structure"`
	
	// Timezone and business hours
	TimeZone          string          `json:"time_zone" db:"time_zone"`
	BusinessHoursStart string         `json:"business_hours_start" db:"business_hours_start"` // HH:MM
	BusinessHoursEnd   string         `json:"business_hours_end" db:"business_hours_end"` // HH:MM
	WorkingDays       []int           `json:"working_days" db:"working_days"` // 1=Monday, 7=Sunday
	
	// Holidays (JSON array of dates)
	Holidays          []string        `json:"holidays" db:"holidays"`
	
	// Custom view configuration
	ViewConfig        RegulatorViewConfig `json:"view_config" db:"view_config"`
	
	// Status
	IsActive          bool            `json:"is_active" db:"is_active"`
	IsVerified        bool            `json:"is_verified" db:"is_verified"`
	VerificationDate  *time.Time      `json:"verification_date" db:"verification_date"`
	VerificationBy    string          `json:"verification_by" db:"verification_by"`
	
	// Notes
	Notes             string          `json:"notes" db:"notes"`
	
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
	CreatedBy         string          `json:"created_by" db:"created_by"`
	UpdatedBy         string          `json:"updated_by" db:"updated_by"`
}

// RegulatorViewConfig defines custom views for a regulator
type RegulatorViewConfig struct {
	// Dashboard customization
	DashboardLayout   []DashboardWidget `json:"dashboard_layout"`
	Theme             ViewTheme         `json:"theme"`
	
	// Widget configurations
	Widgets           []WidgetConfig    `json:"widgets"`
	
	// Report preferences
	DefaultFormats    []ReportFormat   `json:"default_formats"`
	DefaultDateRange  string           `json:"default_date_range"` // last_30d, last_quarter, last_year, custom
	PreferredLanguage string           `json:"preferred_language"` // en, fr, de, etc.
	
	// Filter preferences
	VisibleFields     []string         `json:"visible_fields"`
	HiddenFields      []string         `json:"hidden_fields"`
	DefaultFilters    map[string]interface{} `json:"default_filters"`
	
	// Notification preferences
	AlertSettings     AlertSettings    `json:"alert_settings"`
	
	// Access control
	AccessLevel       string           `json:"access_level"` // read_only, restricted, full
	AllowedActions    []string         `json:"allowed_actions"`
	
	// Custom branding
	LogoURL           string           `json:"logo_url"`
	PrimaryColor      string           `json:"primary_color"`
	SecondaryColor    string           `json:"secondary_color"`
	
	// Integration settings
	IntegrationConfig IntegrationConfig `json:"integration_config"`
}

// DashboardWidget represents a widget on the regulator dashboard
type DashboardWidget struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // metrics, chart, table, alert, timeline
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Position    WidgetPosition         `json:"position"`
	Size        WidgetSize             `json:"size"`
	Config      map[string]interface{} `json:"config"`
	Visible     bool                   `json:"visible"`
	Order       int                    `json:"order"`
}

// WidgetPosition defines the position of a widget
type WidgetPosition struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// WidgetSize defines the size of a widget
type WidgetSize struct {
	Rows    int `json:"rows"`
	Columns int `json:"columns"`
}

// ViewTheme defines the visual theme for a regulator view
type ViewTheme struct {
	Mode        string            `json:"mode"` // light, dark, custom
	Colors      map[string]string `json:"colors"`
	Fonts       FontConfig        `json:"fonts"`
	Spacing     string            `json:"spacing"` // compact, comfortable, spacious
	Roundedness string            `json:"roundedness"` // none, small, medium, large
}

// FontConfig defines font settings
type FontConfig struct {
	Family   string `json:"family"`
	Size     int    `json:"size"`
	BaseSize int    `json:"base_size"`
}

// WidgetConfig defines configuration for a specific widget type
type WidgetConfig struct {
	WidgetType   string                 `json:"widget_type"`
	SourceType   string                 `json:"source_type"` // report, metrics, external
	SourceConfig map[string]interface{} `json:"source_config"`
	RefreshRate  int                    `json:"refresh_rate"` // seconds
	Exportable   bool                   `json:"exportable"`
	Drilldown    DrilldownConfig        `json:"drilldown"`
}

// DrilldownConfig defines drilldown behavior
type DrilldownConfig struct {
	Enabled    bool     `json:"enabled"`
	TargetView string   `json:"target_view"`
	Parameters []string `json:"parameters"`
}

// AlertSettings defines notification preferences
type AlertSettings struct {
	Enabled        bool              `json:"enabled"`
	EmailEnabled   bool              `json:"email_enabled"`
	WebhookEnabled bool              `json:"webhook_enabled"`
	WebhookURL     string            `json:"webhook_url"`
	AlertTypes     []string          `json:"alert_types"` // report_ready, compliance_alert, deadline_approaching
	Thresholds     []AlertThreshold  `json:"thresholds"`
	Frequency      string            `json:"frequency"` // immediate, hourly_digest, daily_digest
	Recipients     []string          `json:"recipients"`
}

// AlertThreshold defines a threshold for triggering alerts
type AlertThreshold struct {
	Metric     string      `json:"metric"`
	Operator   string      `json:"operator"` // gt, lt, eq, gte, lte
	Value      float64     `json:"value"`
	Duration   int         `json:"duration"` // seconds
	Actions    []string    `json:"actions"`
	Message    string      `json:"message"`
}

// IntegrationConfig defines integration settings for external systems
type IntegrationConfig struct {
	SSOEnabled       bool              `json:"sso_enabled"`
	SSOProvider      string            `json:"sso_provider"` // saml, oauth, oidc
	SSOConfig        map[string]string `json:"sso_config"`
	APIKeyEnabled    bool              `json:"api_key_enabled"`
	APIKeyPrefix     string            `json:"api_key_prefix"`
	WebhookEnabled   bool              `json:"webhook_enabled"`
	WebhookEvents    []string          `json:"webhook_events"`
	WebhookSecret    string            `json:"webhook_secret"`
}

// RegulatorUser represents a user associated with a regulator
type RegulatorUser struct {
	ID              uuid.UUID   `json:"id"`
	RegulatorID     uuid.UUID   `json:"regulator_id"`
	UserID          string      `json:"user_id"`
	Email           string      `json:"email"`
	Name            string      `json:"name"`
	Role            string      `json:"role"` // viewer, analyst, manager, admin
	Permissions     []string    `json:"permissions"`
	
	// Access tracking
	LastLoginAt     *time.Time  `json:"last_login_at"`
	LoginCount      int         `json:"login_count"`
	LastActivityAt  *time.Time  `json:"last_activity_at"`
	
	// Preferences
	Preferences     map[string]interface{} `json:"preferences"`
	
	IsActive        bool        `json:"is_active"`
	CreatedAt       time.Time   `json:"created_at"`
	CreatedBy       string      `json:"created_by"`
}

// RegulatorAPIKey represents an API key for regulator API access
type RegulatorAPIKey struct {
	ID            uuid.UUID   `json:"id"`
	RegulatorID   uuid.UUID   `json:"regulator_id"`
	Name          string      `json:"name"`
	KeyHash       string      `json:"key_hash"`
	KeyPrefix     string      `json:"key_prefix"`
	Permissions   []string    `json:"permissions"`
	RateLimit     int         `json:"rate_limit"` // requests per hour
	ExpiresAt     *time.Time  `json:"expires_at"`
	
	LastUsedAt    *time.Time  `json:"last_used_at"`
	UsageCount    int         `json:"usage_count"`
	
	IsActive      bool        `json:"is_active"`
	CreatedAt     time.Time   `json:"created_at"`
	CreatedBy     string      `json:"created_by"`
	RevokedAt     *time.Time  `json:"revoked_at"`
	RevokedBy     string      `json:"revoked_by"`
}
