package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// IdentityStatus represents the status of a citizen identity
type IdentityStatus string

const (
	IdentityStatusPending    IdentityStatus = "pending"
	IdentityStatusVerified   IdentityStatus = "verified"
	IdentityStatusSuspended  IdentityStatus = "suspended"
	IdentityStatusRevoked    IdentityStatus = "revoked"
	IdentityStatusDeceased   IdentityStatus = "deceased"
	IdentityStatusExpatriate IdentityStatus = "expatriate"
)

// VerificationLevel represents the level of identity verification
type VerificationLevel string

const (
	VerificationLevelNone      VerificationLevel = "none"
	VerificationLevelBasic     VerificationLevel = "basic"
	VerificationLevelStandard  VerificationLevel = "standard"
	VerificationLevelEnhanced  VerificationLevel = "enhanced"
	VerificationLevelBiometric VerificationLevel = "biometric"
)

// Gender represents the gender of a citizen
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
	GenderUnknown Gender = "unknown"
)

// Identity represents a citizen's national identity record
type Identity struct {
	ID                 string            `json:"id" db:"id"`
	NationalID         string            `json:"national_id" db:"national_id"`
	LegacyID           *string           `json:"legacy_id,omitempty" db:"legacy_id"`
	FirstName          string            `json:"first_name" db:"first_name"`
	MiddleName         *string           `json:"middle_name,omitempty" db:"middle_name"`
	LastName           string            `json:"last_name" db:"last_name"`
	DateOfBirth        time.Time         `json:"date_of_birth" db:"date_of_birth"`
	PlaceOfBirth       string            `json:"place_of_birth" db:"place_of_birth"`
	Gender             Gender            `json:"gender" db:"gender"`
	Nationality        string            `json:"nationality" db:"nationality"`
	Ethnicity          *string           `json:"ethnicity,omitempty" db:"ethnicity"`
	MaritalStatus      *string           `json:"marital_status,omitempty" db:"marital_status"`
	
	// Address information
	CurrentAddress     Address           `json:"current_address" db:"current_address"`
	PermanentAddress   *Address          `json:"permanent_address,omitempty" db:"permanent_address"`
	
	// Contact information
	Email              *string           `json:"email,omitempty" db:"email"`
	Phone              *string           `json:"phone,omitempty" db:"phone"`
	Mobile             *string           `json:"mobile,omitempty" db:"mobile"`
	
	// Document information
	PassportNumber     *string           `json:"passport_number,omitempty" db:"passport_number"`
	PassportExpiry     *time.Time        `json:"passport_expiry,omitempty" db:"passport_expiry"`
	DriverLicense      *string           `json:"driver_license,omitempty" db:"driver_license"`
	
	// Biometric references
	BiometricID        *string           `json:"biometric_id,omitempty" db:"biometric_id"`
	PhotoURL           *string           `json:"photo_url,omitempty" db:"photo_url"`
	FingerprintHash    *string           `json:"fingerprint_hash,omitempty" db:"fingerprint_hash"`
	IrisScanHash       *string           `json:"iris_scan_hash,omitempty" db:"iris_scan_hash"`
	FacialTemplateHash *string           `json:"facial_template_hash,omitempty" db:"facial_template_hash"`
	
	// Status and verification
	Status             IdentityStatus    `json:"status" db:"status"`
	VerificationLevel  VerificationLevel `json:"verification_level" db:"verification_level"`
	VerifiedAt         *time.Time        `json:"verified_at,omitempty" db:"verified_at"`
	VerifiedBy         *string           `json:"verified_by,omitempty" db:"verified_by"`
	
	// Fraud and risk scoring
	FraudScore         int               `json:"fraud_score" db:"fraud_score"`
	RiskLevel          string            `json:"risk_level" db:"risk_level"`
	LastRiskAssessment *time.Time        `json:"last_risk_assessment,omitempty" db:"last_risk_assessment"`
	
	// Consent and privacy
	ConsentGiven       bool              `json:"consent_given" db:"consent_given"`
	ConsentDate        *time.Time        `json:"consent_date,omitempty" db:"consent_date"`
	DataOptOut         bool              `json:"data_opt_out" db:"data_opt_out"`
	
	// Audit fields
	CreatedAt          time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at" db:"updated_at"`
	CreatedBy          string            `json:"created_by" db:"created_by"`
	UpdatedBy          *string           `json:"updated_by,omitempty" db:"updated_by"`
	Version            int               `json:"version" db:"version"`
	
	// Soft delete
	DeletedAt          *time.Time        `json:"deleted_at,omitempty" db:"deleted_at"`
	IsDeleted          bool              `json:"is_deleted" db:"is_deleted"`
}

// Address represents a physical address
type Address struct {
	Street1        string  `json:"street1"`
	Street2        *string `json:"street2,omitempty"`
	City           string  `json:"city"`
	District       *string `json:"district,omitempty"`
	State          string  `json:"state"`
	PostalCode     string  `json:"postal_code"`
	Country        string  `json:"country"`
	Region         *string `json:"region,omitempty"`
	AddressType    string  `json:"address_type"`
	ValidFrom      *time.Time `json:"valid_from,omitempty"`
	ValidUntil     *time.Time `json:"valid_until,omitempty"`
	IsPrimary      bool    `json:"is_primary"`
}

// Value implements driver.Valuer for Address
func (a Address) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements sql.Scanner for Address
func (a *Address) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Address")
	}
	return json.Unmarshal(bytes, a)
}

// IdentityHistory represents the history of identity changes
type IdentityHistory struct {
	ID            string    `json:"id" db:"id"`
	IdentityID    string    `json:"identity_id" db:"identity_id"`
	Action        string    `json:"action" db:"action"`
	FieldChanged  *string   `json:"field_changed,omitempty" db:"field_changed"`
	OldValue      *string   `json:"old_value,omitempty" db:"old_value"`
	NewValue      *string   `json:"new_value,omitempty" db:"new_value"`
	ChangedBy     string    `json:"changed_by" db:"changed_by"`
	ChangedAt     time.Time `json:"changed_at" db:"changed_at"`
	Reason        *string   `json:"reason,omitempty" db:"reason"`
	IPAddress     *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent     *string   `json:"user_agent,omitempty" db:"user_agent"`
}

// IdentityVerification represents an identity verification event
type IdentityVerification struct {
	ID              string            `json:"id" db:"id"`
	IdentityID      string            `json:"identity_id" db:"identity_id"`
	VerificationType string           `json:"verification_type" db:"verification_type"`
	Method          string            `json:"method" db:"method"`
	Status          string            `json:"status" db:"status"`
	Score           float64           `json:"score" db:"score"`
	Threshold       float64           `json:"threshold" db:"threshold"`
	Passed          bool              `json:"passed" db:"passed"`
	FailureReason   *string           `json:"failure_reason,omitempty" db:"failure_reason"`
	Evidence        *VerificationEvidence `json:"evidence,omitempty" db:"evidence"`
	VerifiedAt      time.Time         `json:"verified_at" db:"verified_at"`
	ExpiresAt       *time.Time        `json:"expires_at,omitempty" db:"expires_at"`
	VerifiedBy      *string           `json:"verified_by,omitempty" db:"verified_by"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
}

// VerificationEvidence contains evidence collected during verification
type VerificationEvidence struct {
	DocumentType       string            `json:"document_type,omitempty"`
	DocumentNumber     *string           `json:"document_number,omitempty"`
	IssueDate          *time.Time        `json:"issue_date,omitempty"`
	ExpiryDate         *time.Time        `json:"expiry_date,omitempty"`
	Issuer             *string           `json:"issuer,omitempty"`
	CountryOfIssue     *string           `json:"country_of_issue,omitempty"`
	FaceMatchScore     float64           `json:"face_match_score"`
	LivenessScore      float64           `json:"liveness_score"`
	SpoofingDetected   bool              `json:"spoofing_detected"`
	OCRData            map[string]string `json:"ocr_data,omitempty"`
	MRZData            *string           `json:"mrz_data,omitempty"`
	BarcodeData        *string           `json:"barcode_data,omitempty"`
}

// Value implements driver.Valuer for VerificationEvidence
func (v VerificationEvidence) Value() (driver.Value, error) {
	return json.Marshal(v)
}

// Scan implements sql.Scanner for VerificationEvidence
func (v *VerificationEvidence) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan VerificationEvidence")
	}
	return json.Unmarshal(bytes, v)
}

// BiometricTemplate represents a biometric template stored in the system
type BiometricTemplate struct {
	ID              string    `json:"id" db:"id"`
	IdentityID      string    `json:"identity_id" db:"identity_id"`
	BiometricType   string    `json:"biometric_type" db:"biometric_type"`
	TemplateHash    string    `json:"template_hash" db:"template_hash"`
	TemplateData    []byte    `json:"template_data,omitempty" db:"template_data"`
	EncryptionKeyID *string   `json:"encryption_key_id,omitempty" db:"encryption_key_id"`
	QualityScore    float64   `json:"quality_score" db:"quality_score"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	IsActive        bool      `json:"is_active" db:"is_active"`
}

// BehavioralProfile represents a citizen's behavioral biometrics profile
type BehavioralProfile struct {
	ID                    string           `json:"id" db:"id"`
	IdentityID            string           `json:"identity_id" db:"identity_id"`
	
	// Keystroke dynamics
	KeystrokeProfile      *KeystrokeProfile `json:"keystroke_profile,omitempty" db:"keystroke_profile"`
	
	// Mouse dynamics
	MouseProfile          *MouseProfile     `json:"mouse_profile,omitempty" db:"mouse_profile"`
	
	// Touch dynamics (mobile)
	TouchProfile          *TouchProfile     `json:"touch_profile,omitempty" db:"touch_profile"`
	
	// Gait analysis (wearable devices)
	GaitProfile           *GaitProfile      `json:"gait_profile,omitempty" db:"gait_profile"`
	
	OverallScore          float64          `json:"overall_score" db:"overall_score"`
	AnomalyCount          int              `json:"anomaly_count" db:"anomaly_count"`
	LastAnomalyAt         *time.Time       `json:"last_anomaly_at,omitempty" db:"last_anomaly_at"`
	BaselineEstablishedAt time.Time        `json:"baseline_established_at" db:"baseline_established_at"`
	UpdatedAt             time.Time        `json:"updated_at" db:"updated_at"`
	SampleCount           int              `json:"sample_count" db:"sample_count"`
}

// KeystrokeProfile contains keystroke dynamics data
type KeystrokeProfile struct {
	AvgKeyPressDuration    float64 `json:"avg_key_press_duration"`
	AvgKeyReleaseDuration  float64 `json:"avg_key_release_duration"`
	AvgInterKeyInterval    float64 `json:"avg_inter_key_interval"`
	AvgFlightTime          float64 `json:"avg_flight_time"`
	DigraphLatencyStdDev   float64 `json:"digraph_latency_std_dev"`
	TrigraphLatencyStdDev  float64 `json:"trigraph_latency_std_dev"`
	TotalTypingTime        float64 `json:"total_typing_time"`
	ErrorRate              float64 `json:"error_rate"`
	BackspaceFrequency     float64 `json:"backspace_frequency"`
}

// Value implements driver.Valuer for KeystrokeProfile
func (k KeystrokeProfile) Value() (driver.Value, error) {
	return json.Marshal(k)
}

// Scan implements sql.Scanner for KeystrokeProfile
func (k *KeystrokeProfile) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan KeystrokeProfile")
	}
	return json.Unmarshal(bytes, k)
}

// MouseProfile contains mouse movement dynamics data
type MouseProfile struct {
	AvgMovementSpeed      float64 `json:"avg_movement_speed"`
	AvgAcceleration       float64 `json:"avg_acceleration"`
	AvgCurvature          float64 `json:"avg_curvature"`
	ClickAccuracy         float64 `json:"click_accuracy"`
	AvgClickDuration      float64 `json:"avg_click_duration"`
	ScrollFrequency       float64 `json:"scroll_frequency"`
	MovementPatternVar    float64 `json:"movement_pattern_variance"`
	RightClickRatio       float64 `json:"right_click_ratio"`
	DoubleClickSpeed      float64 `json:"double_click_speed"`
}

// Value implements driver.Valuer for MouseProfile
func (m MouseProfile) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements sql.Scanner for MouseProfile
func (m *MouseProfile) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan MouseProfile")
	}
	return json.Unmarshal(bytes, m)
}

// TouchProfile contains touch interaction data
type TouchProfile struct {
	AvgTouchPressure      float64 `json:"avg_touch_pressure"`
	AvgTouchDuration      float64 `json:"avg_touch_duration"`
	AvgSwipeSpeed         float64 `json:"avg_swipe_speed"`
	AvgSwipeLength        float64 `json:"avg_swipe_length"`
	TouchAccuracy         float64 `json:"touch_accuracy"`
	MultiTouchFrequency   float64 `json:"multi_touch_frequency"`
	PinchGestureRatio     float64 `json:"pinch_gesture_ratio"`
	RotationGestureRatio  float64 `json:"rotation_gesture_ratio"`
}

// Value implements driver.Valuer for TouchProfile
func (t TouchProfile) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan implements sql.Scanner for TouchProfile
func (t *TouchProfile) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan TouchProfile")
	}
	return json.Unmarshal(bytes, t)
}

// GaitProfile contains gait analysis data from wearable devices
type GaitProfile struct {
	AvgStepLength         float64 `json:"avg_step_length"`
	AvgStepDuration       float64 `json:"avg_step_duration"`
	AvgCadence            float64 `json:"avg_cadence"`
	AvgWalkingSpeed       float64 `json:"avg_walking_speed"`
	SymmetryIndex         float64 `json:"symmetry_index"`
	VariabilityIndex      float64 `json:"variability_index"`
	StanceTimeRatio       float64 `json:"stance_time_ratio"`
	SwingTimeRatio        float64 `json:"swing_time_ratio"`
	StepRegularity        float64 `json:"step_regularity"`
}

// Value implements driver.Valuer for GaitProfile
func (g GaitProfile) Value() (driver.Value, error) {
	return json.Marshal(g)
}

// Scan implements sql.Scanner for GaitProfile
func (g *GaitProfile) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan GaitProfile")
	}
	return json.Unmarshal(bytes, g)
}

// BehavioralDataPoint represents a single behavioral data submission
type BehavioralDataPoint struct {
	ID              string          `json:"id" db:"id"`
	IdentityID      string          `json:"identity_id" db:"identity_id"`
	DataType        string          `json:"data_type" db:"data_type"`
	RawData         []byte          `json:"raw_data,omitempty" db:"raw_data"`
	ProcessedData   *ProcessedData  `json:"processed_data,omitempty" db:"processed_data"`
	AnomalyScore    float64         `json:"anomaly_score" db:"anomaly_score"`
	IsAnomaly       bool            `json:"is_anomaly" db:"is_anomaly"`
	DeviceInfo      *DeviceInfo     `json:"device_info,omitempty" db:"device_info"`
	SessionID       *string         `json:"session_id,omitempty" db:"session_id"`
	SubmittedAt     time.Time       `json:"submitted_at" db:"submitted_at"`
	ProcessedAt     *time.Time      `json:"processed_at,omitempty" db:"processed_at"`
}

// ProcessedData contains processed behavioral data
type ProcessedData struct {
	Features        map[string]float64 `json:"features"`
	NormalizedData  map[string]float64 `json:"normalized_data"`
	ModelVersion    string             `json:"model_version"`
	ConfidenceScore float64            `json:"confidence_score"`
}

// Value implements driver.Valuer for ProcessedData
func (p ProcessedData) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner for ProcessedData
func (p *ProcessedData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ProcessedData")
	}
	return json.Unmarshal(bytes, p)
}

// DeviceInfo contains information about the device used
type DeviceInfo struct {
	DeviceType      string  `json:"device_type"`
	OS              string  `json:"os"`
	OSVersion       string  `json:"os_version"`
	Browser         *string `json:"browser,omitempty"`
	BrowserVersion  *string `json:"browser_version,omitempty"`
	DeviceModel     *string `json:"device_model,omitempty"`
	DeviceBrand     *string `json:"device_brand,omitempty"`
	IPAddress       string  `json:"ip_address"`
	UserAgent       string  `json:"user_agent"`
}

// Value implements driver.Valuer for DeviceInfo
func (d DeviceInfo) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements sql.Scanner for DeviceInfo
func (d *DeviceInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan DeviceInfo")
	}
	return json.Unmarshal(bytes, d)
}

// FraudAlert represents a fraud alert for an identity
type FraudAlert struct {
	ID              string    `json:"id" db:"id"`
	IdentityID      string    `json:"identity_id" db:"identity_id"`
	AlertType       string    `json:"alert_type" db:"alert_type"`
	Severity        string    `json:"severity" db:"severity"`
	Score           int       `json:"score" db:"score"`
	Indicator       string    `json:"indicator" db:"indicator"`
	Description     string    `json:"description" db:"description"`
	Evidence        *string   `json:"evidence,omitempty" db:"evidence"`
	RelatedEventID  *string   `json:"related_event_id,omitempty" db:"related_event_id"`
	Status          string    `json:"status" db:"status"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy      *string   `json:"resolved_by,omitempty" db:"resolved_by"`
	Resolution      *string   `json:"resolution,omitempty" db:"resolution"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// AuthenticationSession represents an active authentication session
type AuthenticationSession struct {
	ID              string          `json:"id" db:"id"`
	IdentityID      string          `json:"identity_id" db:"identity_id"`
	SessionToken    string          `json:"session_token" db:"session_token"`
	RefreshToken    string          `json:"refresh_token" db:"refresh_token"`
	IPAddress       string          `json:"ip_address" db:"ip_address"`
	UserAgent       string          `json:"user_agent" db:"user_agent"`
	DeviceInfo      *DeviceInfo     `json:"device_info,omitempty" db:"device_info"`
	MFAEnabled      bool            `json:"mfa_enabled" db:"mfa_enabled"`
	MFAPassed       bool            `json:"mfa_passed" db:"mfa_passed"`
	BehavioralScore float64         `json:"behavioral_score" db:"behavioral_score"`
	RiskScore       int             `json:"risk_score" db:"risk_score"`
	ExpiresAt       time.Time       `json:"expires_at" db:"expires_at"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	LastActivityAt  time.Time       `json:"last_activity_at" db:"last_activity_at"`
	TerminatedAt    *time.Time      `json:"terminated_at,omitempty" db:"terminated_at"`
	TerminationReason *string       `json:"termination_reason,omitempty" db:"termination_reason"`
}

// MFAToken represents a multi-factor authentication token
type MFAToken struct {
	ID              string    `json:"id" db:"id"`
	IdentityID      string    `json:"identity_id" db:"identity_id"`
	TokenType       string    `json:"token_type" db:"token_type"`
	TokenHash       string    `json:"token_hash" db:"token_hash"`
	Secret          *string   `json:"secret,omitempty" db:"secret"`
	PhoneNumber     *string   `json:"phone_number,omitempty" db:"phone_number"`
	EmailAddress    *string   `json:"email_address,omitempty" db:"email_address"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty" db:"verified_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	ExpiresAt       time.Time `json:"expires_at" db:"expires_at"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	FailedAttempts  int       `json:"failed_attempts" db:"failed_attempts"`
	LockedUntil     *time.Time `json:"locked_until,omitempty" db:"locked_until"`
}

// FederatedModelUpdate represents a federated learning model update from a client
type FederatedModelUpdate struct {
	ID              string    `json:"id" db:"id"`
	ClientID        string    `json:"client_id" db:"client_id"`
	ModelVersion    string    `json:"model_version" db:"model_version"`
	UpdateData      []byte    `json:"update_data,omitempty" db:"update_data"`
	UpdateHash      string    `json:"update_hash" db:"update_hash"`
	SampleCount     int       `json:"sample_count" db:"sample_count"`
	Accuracy        float64   `json:"accuracy" db:"accuracy"`
	ClientVersion   string    `json:"client_version" db:"client_version"`
	DeviceInfo      *DeviceInfo `json:"device_info,omitempty" db:"device_info"`
	SubmittedAt     time.Time `json:"submitted_at" db:"submitted_at"`
	Status          string    `json:"status" db:"status"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	Weight          float64   `json:"weight" db:"weight"`
}

// ModelStatus represents the current status of the federated learning model
type ModelStatus struct {
	ModelID         string    `json:"model_id"`
	Version         string    `json:"version"`
	GlobalAccuracy  float64   `json:"global_accuracy"`
	TotalSamples    int       `json:"total_samples"`
	TotalClients    int       `json:"total_clients"`
	LastUpdatedAt   time.Time `json:"last_updated_at"`
	CurrentRound    int       `json:"current_round"`
	RoundStatus     string    `json:"round_status"`
	NextRoundAt     *time.Time `json:"next_round_at,omitempty"`
}

// ClientContribution represents a client's contribution to the federated model
type ClientContribution struct {
	ClientID        string    `json:"client_id"`
	SampleCount     int       `json:"sample_count"`
	Accuracy        float64   `json:"accuracy"`
	Weight          float64   `json:"weight"`
	SubmittedAt     time.Time `json:"submitted_at"`
}

// ConsentRecord represents a citizen's consent record for data processing
type ConsentRecord struct {
	ID              string    `json:"id" db:"id"`
	IdentityID      string    `json:"identity_id" db:"identity_id"`
	ConsentType     string    `json:"consent_type" db:"consent_type"`
	Purpose         string    `json:"purpose" db:"purpose"`
	DataCategories  []string  `json:"data_categories" db:"data_categories"`
	ThirdParties    []string  `json:"third_parties,omitempty" db:"third_parties"`
	GeoScope        *string   `json:"geo_scope,omitempty" db:"geo_scope"`
	RetentionPeriod string    `json:"retention_period" db:"retention_period"`
	ConsentGiven    bool      `json:"consent_given" db:"consent_given"`
	ConsentDate     time.Time `json:"consent_date" db:"consent_date"`
	WithdrawalDate  *time.Time `json:"withdrawal_date,omitempty" db:"withdrawal_date"`
	IPAddress       *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent       *string   `json:"user_agent,omitempty" db:"user_agent"`
	Version         int       `json:"version" db:"version"`
}

// Value implements driver.Valuer for ConsentRecord
func (c ConsentRecord) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements sql.Scanner for ConsentRecord
func (c *ConsentRecord) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ConsentRecord")
	}
	return json.Unmarshal(bytes, c)
}

// IdentityEvent represents an event in the identity lifecycle
type IdentityEvent struct {
	ID              string          `json:"id" db:"id"`
	IdentityID      string          `json:"identity_id" db:"identity_id"`
	EventType       string          `json:"event_type" db:"event_type"`
	EventData       json.RawMessage `json:"event_data,omitempty" db:"event_data"`
	SourceSystem    string          `json:"source_system" db:"source_system"`
	CorrelationID   *string         `json:"correlation_id,omitempty" db:"correlation_id"`
	ActorID         *string         `json:"actor_id,omitempty" db:"actor_id"`
	IPAddress       *string         `json:"ip_address,omitempty" db:"ip_address"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}
