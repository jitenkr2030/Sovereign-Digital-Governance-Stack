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

	"github.com/jitenkr2030/csic-platform/extended-services/national-identity/internal/domain"
	"github.com/jitenkr2030/csic-platform/extended-services/national-identity/internal/repository"
)

// IdentityServiceConfig holds the configuration for the identity service
type IdentityServiceConfig struct {
	Repo              repository.PostgresRepository
	Cache             *repository.RedisCache
	FraudScoreThreshold int
	VelocityCheckWindow string
	MaxLoginAttempts    int
	LockoutDuration     string
	Behavioral          domain.BehavioralConfig
	FederatedLearning   domain.FederatedLearning
	Biometric           domain.BiometricConfig
	Privacy             domain.PrivacyConfig
}

// IdentityService handles all identity-related business logic
type IdentityService interface {
	// Identity management
	CreateIdentity(ctx context.Context, identity *domain.Identity) error
	GetIdentity(ctx context.Context, id string) (*domain.Identity, error)
	GetIdentityByNationalID(ctx context.Context, nationalID string) (*domain.Identity, error)
	UpdateIdentity(ctx context.Context, identity *domain.Identity) error
	DeleteIdentity(ctx context.Context, id string) error
	GetIdentityHistory(ctx context.Context, id string, limit, offset int) ([]domain.IdentityHistory, error)

	// Verification
	VerifyIdentity(ctx context.Context, req *VerificationRequest) (*VerificationResult, error)
	BiometricVerification(ctx context.Context, req *BiometricVerificationRequest) (*VerificationResult, error)
	DocumentVerification(ctx context.Context, req *DocumentVerificationRequest) (*VerificationResult, error)
	LivenessCheck(ctx context.Context, req *LivenessCheckRequest) (*LivenessResult, error)

	// Authentication
	Login(ctx context.Context, req *LoginRequest) (*LoginResult, error)
	Logout(ctx context.Context, sessionToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	SetupMFA(ctx context.Context, identityID, tokenType string) (*MFASetupResult, error)
	VerifyMFA(ctx context.Context, req *MFAVerifyRequest) error
	PasswordReset(ctx context.Context, req *PasswordResetRequest) error
	ChangePassword(ctx context.Context, req *ChangePasswordRequest) error

	// Behavioral biometrics
	SubmitKeystrokeData(ctx context.Context, identityID string, data *KeystrokeData) error
	SubmitMouseData(ctx context.Context, identityID string, data *MouseData) error
	AnalyzeBehavioralPattern(ctx context.Context, identityID string) (*BehavioralAnalysisResult, error)
	GetBehavioralProfile(ctx context.Context, identityID string) (*domain.BehavioralProfile, error)

	// Fraud detection
	GetFraudScore(ctx context.Context, identityID string) (*FraudScoreResult, error)
	ReportFraud(ctx context.Context, alert *domain.FraudAlert) error
	GetFraudAlerts(ctx context.Context, identityID string) ([]domain.FraudAlert, error)
	ResolveFraudAlert(ctx context.Context, alertID, resolvedBy, resolution string) error

	// Federated learning
	GetModelStatus(ctx context.Context) (*domain.ModelStatus, error)
	SubmitModelUpdate(ctx context.Context, update *domain.FederatedModelUpdate) error
	GetClientContributions(ctx context.Context, clientID string) ([]domain.ClientContribution, error)
}

// identityService implements IdentityService
type identityService struct {
	config IdentityServiceConfig
	now    func() time.Time
}

// NewIdentityService creates a new identity service instance
func NewIdentityService(config IdentityServiceConfig) IdentityService {
	return &identityService{
		config: config,
		now:    time.Now,
	}
}

// CreateIdentity creates a new citizen identity
func (s *identityService) CreateIdentity(ctx context.Context, identity *domain.Identity) error {
	// Validate required fields
	if identity.NationalID == "" {
		return errors.New("national ID is required")
	}
	if identity.FirstName == "" {
		return errors.New("first name is required")
	}
	if identity.LastName == "" {
		return errors.New("last name is required")
	}
	if identity.DateOfBirth.IsZero() {
		return errors.New("date of birth is required")
	}
	if identity.PlaceOfBirth == "" {
		return errors.New("place of birth is required")
	}

	// Check if national ID already exists
	existing, err := s.config.Repo.GetIdentityByNationalID(ctx, identity.NationalID)
	if err != nil {
		return fmt.Errorf("failed to check existing identity: %w", err)
	}
	if existing != nil {
		return errors.New("identity with this national ID already exists")
	}

	// Set defaults
	identity.Status = domain.IdentityStatusPending
	identity.VerificationLevel = domain.VerificationLevelNone
	identity.RiskLevel = "low"
	identity.FraudScore = 0
	identity.CreatedBy = "system"

	// Create the identity
	if err := s.config.Repo.CreateIdentity(ctx, identity); err != nil {
		return fmt.Errorf("failed to create identity: %w", err)
	}

	// Publish event
	event := &domain.IdentityEvent{
		ID:           generateID(),
		IdentityID:   identity.ID,
		EventType:    "IDENTITY_CREATED",
		EventData:    mustMarshal(identity),
		SourceSystem: "identity-service",
		CreatedAt:    s.now(),
	}
	if err := s.config.Repo.PublishEvent(ctx, event); err != nil {
		log.Printf("Failed to publish identity created event: %v", err)
	}

	return nil
}

// GetIdentity retrieves an identity by ID
func (s *identityService) GetIdentity(ctx context.Context, id string) (*domain.Identity, error) {
	if id == "" {
		return nil, errors.New("identity ID is required")
	}

	identity, err := s.config.Repo.GetIdentity(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}

	if identity == nil {
		return nil, errors.New("identity not found")
	}

	// Apply privacy filters
	s.applyPrivacyFilters(identity)

	return identity, nil
}

// GetIdentityByNationalID retrieves an identity by national ID number
func (s *identityService) GetIdentityByNationalID(ctx context.Context, nationalID string) (*domain.Identity, error) {
	if nationalID == "" {
		return nil, errors.New("national ID is required")
	}

	identity, err := s.config.Repo.GetIdentityByNationalID(ctx, nationalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}

	if identity == nil {
		return nil, errors.New("identity not found")
	}

	// Apply privacy filters
	s.applyPrivacyFilters(identity)

	return identity, nil
}

// UpdateIdentity updates an existing identity
func (s *identityService) UpdateIdentity(ctx context.Context, identity *domain.Identity) error {
	// Get existing identity
	existing, err := s.config.Repo.GetIdentity(ctx, identity.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing identity: %w", err)
	}
	if existing == nil {
		return errors.New("identity not found")
	}

	// Check for changes that require re-verification
	if identity.Status != existing.Status || identity.VerificationLevel != existing.VerificationLevel {
		identity.VerifiedAt = nil
		identity.VerifiedBy = nil
	}

	identity.UpdatedBy = stringPtr("system")

	if err := s.config.Repo.UpdateIdentity(ctx, identity); err != nil {
		return fmt.Errorf("failed to update identity: %w", err)
	}

	// Create history entry
	history := &domain.IdentityHistory{
		ID:          generateID(),
		IdentityID:  identity.ID,
		Action:      "UPDATE",
		ChangedBy:   "system",
		ChangedAt:   s.now(),
		Reason:      stringPtr("Identity updated"),
	}

	return s.config.Repo.CreateIdentityHistory(ctx, history)
}

// DeleteIdentity soft-deletes an identity
func (s *identityService) DeleteIdentity(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("identity ID is required")
	}

	// Check if identity exists
	existing, err := s.config.Repo.GetIdentity(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get identity: %w", err)
	}
	if existing == nil {
		return errors.New("identity not found")
	}

	if err := s.config.Repo.DeleteIdentity(ctx, id); err != nil {
		return fmt.Errorf("failed to delete identity: %w", err)
	}

	// Terminate all sessions
	if err := s.config.Repo.TerminateAllSessions(ctx, id, "identity_deleted"); err != nil {
		log.Printf("Failed to terminate sessions: %v", err)
	}

	return nil
}

// GetIdentityHistory retrieves the change history of an identity
func (s *identityService) GetIdentityHistory(ctx context.Context, id string, limit, offset int) ([]domain.IdentityHistory, error) {
	if id == "" {
		return nil, errors.New("identity ID is required")
	}

	if limit <= 0 {
		limit = 50
	}

	return s.config.Repo.GetIdentityHistory(ctx, id, limit, offset)
}

// Verification methods

// VerificationRequest represents a verification request
type VerificationRequest struct {
	IdentityID     string                 `json:"identity_id"`
	VerificationType string               `json:"verification_type"`
	DocumentData   map[string]interface{} `json:"document_data,omitempty"`
	SelfieData     string                 `json:"selfie_data,omitempty"`
	BiometricData  *BiometricData         `json:"biometric_data,omitempty"`
}

// BiometricData contains biometric verification data
type BiometricData struct {
	Type       string `json:"type"`
	Template   string `json:"template"`
	LivenessID string `json:"liveness_id"`
}

// VerificationResult represents the result of a verification
type VerificationResult struct {
	VerificationID string  `json:"verification_id"`
	Passed         bool    `json:"passed"`
	Score          float64 `json:"score"`
	Threshold      float64 `json:"threshold"`
	FailureReason  string  `json:"failure_reason,omitempty"`
	Details        map[string]interface{} `json:"details,omitempty"`
}

// VerifyIdentity performs identity verification
func (s *identityService) VerifyIdentity(ctx context.Context, req *VerificationRequest) (*VerificationResult, error) {
	identity, err := s.config.Repo.GetIdentity(ctx, req.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}
	if identity == nil {
		return nil, errors.New("identity not found")
	}

	// Perform verification based on type
	var result VerificationResult
	result.Threshold = 0.85

	switch req.VerificationType {
	case "document":
		result = s.performDocumentVerification(identity, req.DocumentData)
	case "biometric":
		result = s.performBiometricVerification(identity, req.BiometricData)
	case "knowledge-based":
		result = s.performKnowledgeBasedVerification(identity, req.DocumentData)
	default:
		return nil, fmt.Errorf("unsupported verification type: %s", req.VerificationType)
	}

	// Create verification record
	verification := &domain.IdentityVerification{
		ID:              generateID(),
		IdentityID:      req.IdentityID,
		VerificationType: req.VerificationType,
		Method:          "automated",
		Status:          "completed",
		Score:           result.Score,
		Threshold:       result.Threshold,
		Passed:          result.Passed,
		FailureReason:   stringPtr(result.FailureReason),
		VerifiedAt:      s.now(),
	}

	if err := s.config.Repo.CreateVerification(ctx, verification); err != nil {
		log.Printf("Failed to create verification record: %v", err)
	}

	// Update identity verification level if passed
	if result.Passed {
		newLevel := s.upgradeVerificationLevel(identity.VerificationLevel)
		if newLevel != identity.VerificationLevel {
			identity.VerificationLevel = newLevel
			now := s.now()
			identity.VerifiedAt = &now
			identity.VerifiedBy = stringPtr("automated")
			if err := s.config.Repo.UpdateIdentity(ctx, identity); err != nil {
				log.Printf("Failed to update identity verification level: %v", err)
			}
		}
	}

	return &result, nil
}

func (s *identityService) performDocumentVerification(identity *domain.Identity, documentData map[string]interface{}) VerificationResult {
	result := VerificationResult{Passed: false, Score: 0}

	// Check document data against identity
	score := 0.0
	matches := 0
	totalChecks := 5

	// Check name match
	if firstName, ok := documentData["first_name"].(string); ok {
		if strings.EqualFold(firstName, identity.FirstName) {
			score += 0.2
			matches++
		}
	}
	if lastName, ok := documentData["last_name"].(string); ok {
		if strings.EqualFold(lastName, identity.LastName) {
			score += 0.2
			matches++
		}
	}

	// Check date of birth
	if dob, ok := documentData["date_of_birth"].(string); ok {
		if dob == identity.DateOfBirth.Format("2006-01-02") {
			score += 0.2
			matches++
		}
	}

	// Check nationality
	if nationality, ok := documentData["nationality"].(string); ok {
		if strings.EqualFold(nationality, identity.Nationality) {
			score += 0.2
			matches++
		}
	}

	// Check document number
	if docNum, ok := documentData["document_number"].(string); ok {
		if identity.PassportNumber != nil && strings.EqualFold(docNum, *identity.PassportNumber) {
			score += 0.2
			matches++
		}
	}

	result.Score = score
	result.Passed = matches >= 4
	if !result.Passed {
		result.FailureReason = fmt.Sprintf("Document verification failed: %d/%d checks passed", matches, totalChecks)
	}

	return result
}

func (s *identityService) performBiometricVerification(identity *domain.Identity, biometricData *domain.BiometricData) VerificationResult {
	result := VerificationResult{Passed: false, Score: 0}

	if biometricData == nil {
		result.FailureReason = "No biometric data provided"
		return result
	}

	// Get stored biometric template
	template, err := s.config.Repo.GetBiometricTemplate(context.Background(), identity.ID, biometricData.Type)
	if err != nil || template == nil {
		result.FailureReason = "No biometric template found for this identity"
		return result
	}

	// Perform biometric matching (simplified - in production, use proper biometric SDK)
	score := s.calculateBiometricScore(biometricData.Template, template.TemplateHash)
	result.Score = score
	result.Threshold = 0.85
	result.Passed = score >= result.Threshold

	if !result.Passed {
		result.FailureReason = fmt.Sprintf("Biometric match score %.2f below threshold %.2f", score, result.Threshold)
	}

	return result
}

func (s *identityService) performKnowledgeBasedVerification(identity *domain.Identity, data map[string]interface{}) VerificationResult {
	result := VerificationResult{Passed: false, Score: 0}

	// Simplified KBA - in production, integrate with credit bureau or similar
	score := 0.8
	correctAnswers := 0
	totalQuestions := 3

	// Check answers (simplified)
	if answer, ok := data["answer_1"].(string); ok {
		if s.verifyKBAAnswer(identity, "question_1", answer) {
			correctAnswers++
			score += 0.05
		}
	}
	if answer, ok := data["answer_2"].(string); ok {
		if s.verifyKBAAnswer(identity, "question_2", answer) {
			correctAnswers++
			score += 0.05
		}
	}
	if answer, ok := data["answer_3"].(string); ok {
		if s.verifyKBAAnswer(identity, "question_3", answer) {
			correctAnswers++
			score += 0.05
		}
	}

	result.Score = score
	result.Passed = correctAnswers >= 2
	if !result.Passed {
		result.FailureReason = fmt.Sprintf("KBA verification failed: %d/%d questions correct", correctAnswers, totalQuestions)
	}

	return result
}

func (s *identityService) calculateBiometricScore(inputTemplate, storedTemplate string) float64 {
	// Simplified biometric scoring - in production, use proper biometric SDK
	// This is a placeholder that simulates biometric matching
	return 0.9
}

func (s *identityService) verifyKBAAnswer(identity *domain.Identity, question, answer string) bool {
	// Simplified KBA verification - in production, use proper KBA service
	return true
}

func (s *identityService) upgradeVerificationLevel(current domain.VerificationLevel) domain.VerificationLevel {
	switch current {
	case domain.VerificationLevelNone:
		return domain.VerificationLevelBasic
	case domain.VerificationLevelBasic:
		return domain.VerificationLevelStandard
	case domain.VerificationLevelStandard:
		return domain.VerificationLevelEnhanced
	case domain.VerificationLevelEnhanced:
		return domain.VerificationLevelBiometric
	default:
		return current
	}
}

// BiometricVerification handles biometric verification requests
type BiometricVerificationRequest struct {
	IdentityID    string `json:"identity_id"`
	BiometricType string `json:"biometric_type"`
	Template      string `json:"template"`
	LivenessToken string `json:"liveness_token"`
}

// BiometricVerification result
type BiometricVerificationResult struct {
	VerificationID string  `json:"verification_id"`
	Passed         bool    `json:"passed"`
	Score          float64 `json:"score"`
	LivenessScore  float64 `json:"liveness_score"`
	SpoofingDetected bool  `json:"spoofing_detected"`
}

func (s *identityService) BiometricVerification(ctx context.Context, req *BiometricVerificationRequest) (*VerificationResult, error) {
	identity, err := s.config.Repo.GetIdentity(ctx, req.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}
	if identity == nil {
		return nil, errors.New("identity not found")
	}

	// Get stored template
	storedTemplate, err := s.config.Repo.GetBiometricTemplate(ctx, req.IdentityID, req.BiometricType)
	if err != nil {
		return nil, fmt.Errorf("failed to get biometric template: %w", err)
	}
	if storedTemplate == nil {
		return nil, errors.New("no biometric template found")
	}

	// Calculate match score
	matchScore := s.calculateBiometricScore(req.Template, storedTemplate.TemplateHash)

	// Verify liveness if token provided
	livenessScore := 1.0
	spoofingDetected := false
	if req.LivenessToken != "" {
		livenessScore, spoofingDetected = s.verifyLiveness(req.LivenessToken)
	}

	passed := matchScore >= 0.85 && livenessScore >= 0.9 && !spoofingDetected

	return &VerificationResult{
		VerificationID: generateID(),
		Passed:         passed,
		Score:          matchScore * livenessScore,
		Threshold:      0.85,
		Details: map[string]interface{}{
			"biometric_type":     req.BiometricType,
			"liveness_score":     livenessScore,
			"spoofing_detected":  spoofingDetected,
		},
	}, nil
}

func (s *identityService) verifyLiveness(token string) (float64, bool) {
	// Simplified liveness verification - in production, use proper liveness detection SDK
	return 0.95, false
}

// DocumentVerification handles document verification requests
type DocumentVerificationRequest struct {
	IdentityID     string                 `json:"identity_id"`
	DocumentType   string                 `json:"document_type"`
	DocumentData   map[string]interface{} `json:"document_data"`
	DocumentImage  string                 `json:"document_image"`
	SelfieImage    string                 `json:"selfie_image"`
}

func (s *identityService) DocumentVerification(ctx context.Context, req *DocumentVerificationRequest) (*VerificationResult, error) {
	identity, err := s.config.Repo.GetIdentity(ctx, req.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}
	if identity == nil {
		return nil, errors.New("identity not found")
	}

	// Perform document verification
	result := s.performDocumentVerification(identity, req.DocumentData)

	return &VerificationResult{
		VerificationID: generateID(),
		Passed:         result.Passed,
		Score:          result.Score,
		Threshold:      result.Threshold,
		FailureReason:  result.FailureReason,
	}, nil
}

// LivenessCheckRequest represents a liveness check request
type LivenessCheckRequest struct {
	IdentityID string `json:"identity_id"`
	Challenge  string `json:"challenge"`
	Response   string `json:"response"`
}

// LivenessResult represents the result of a liveness check
type LivenessResult struct {
	Passed          bool    `json:"passed"`
	Score           float64 `json:"score"`
	SpoofingDetected bool   `json:"spoofing_detected"`
	SessionID       string  `json:"session_id"`
}

func (s *identityService) LivenessCheck(ctx context.Context, req *LivenessCheckRequest) (*LivenessResult, error) {
	// Simplified liveness check - in production, use proper liveness detection
	passed := true
	spoofingDetected := false
	score := 0.95

	// In a real implementation, this would:
	// 1. Generate a random challenge
	// 2. Request the user to perform specific actions (blink, turn head, etc.)
	// 3. Analyze the video/image to detect spoofing attempts
	// 4. Return the liveness score

	return &LivenessResult{
		Passed:          passed,
		Score:           score,
		SpoofingDetected: spoofingDetected,
		SessionID:       generateID(),
	}, nil
}

// Authentication methods

// LoginRequest represents a login request
type LoginRequest struct {
	NationalID   string       `json:"national_id"`
	Password     string       `json:"password"`
	DeviceInfo   *DeviceInfo  `json:"device_info"`
	BehavioralData *BehavioralDataRequest `json:"behavioral_data,omitempty"`
}

// LoginResult represents the result of a successful login
type LoginResult struct {
	SessionToken    string                  `json:"session_token"`
	RefreshToken    string                  `json:"refresh_token"`
	ExpiresAt       time.Time               `json:"expires_at"`
	MFARequired     bool                    `json:"mfa_required"`
	MFAMethods      []string                `json:"mfa_methods,omitempty"`
	RiskScore       int                     `json:"risk_score"`
	BehavioralScore float64                 `json:"behavioral_score"`
}

// DeviceInfo contains device information
type DeviceInfo struct {
	DeviceType     string  `json:"device_type"`
	OS             string  `json:"os"`
	OSVersion      string  `json:"os_version"`
	Browser        *string `json:"browser,omitempty"`
	BrowserVersion *string `json:"browser_version,omitempty"`
	DeviceModel    *string `json:"device_model,omitempty"`
	DeviceBrand    *string `json:"device_brand,omitempty"`
	IPAddress      string  `json:"ip_address"`
	UserAgent      string  `json:"user_agent"`
}

// BehavioralDataRequest contains behavioral data for risk assessment
type BehavioralDataRequest struct {
	KeystrokeData *KeystrokeData `json:"keystroke_data,omitempty"`
	MouseData     *MouseData     `json:"mouse_data,omitempty"`
}

func (s *identityService) Login(ctx context.Context, req *LoginRequest) (*LoginResult, error) {
	// Get identity by national ID
	identity, err := s.config.Repo.GetIdentityByNationalID(ctx, req.NationalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}
	if identity == nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if account is locked
	if identity.Status == domain.IdentityStatusSuspended {
		return nil, errors.New("account is suspended")
	}

	// Verify password (simplified - in production, use proper password hashing)
	if !s.verifyPassword(req.Password, identity.ID) {
		return nil, errors.New("invalid credentials")
	}

	// Calculate risk score
	riskScore, behavioralScore := s.calculateLoginRisk(ctx, identity.ID, req)

	// Check if MFA is required
	mfaRequired := riskScore > 50 || s.config.Biometric.RequiredFactors > 1

	// Create session
	session := &domain.AuthenticationSession{
		ID:             generateID(),
		IdentityID:     identity.ID,
		SessionToken:   generateSecureToken(),
		RefreshToken:   generateSecureToken(),
		IPAddress:      req.DeviceInfo.IPAddress,
		UserAgent:      req.DeviceInfo.UserAgent,
		DeviceInfo:     (*domain.DeviceInfo)(req.DeviceInfo),
		MFAEnabled:     false,
		MFAPassed:      !mfaRequired,
		BehavioralScore: behavioralScore,
		RiskScore:      riskScore,
		ExpiresAt:      s.now().Add(24 * time.Hour),
	}

	if err := s.config.Repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	result := &LoginResult{
		SessionToken:    session.SessionToken,
		RefreshToken:    session.RefreshToken,
		ExpiresAt:       session.ExpiresAt,
		MFARequired:     mfaRequired,
		RiskScore:       riskScore,
		BehavioralScore: behavioralScore,
	}

	if mfaRequired {
		result.MFAMethods = []string{"totp", "sms", "email"}
	}

	return result, nil
}

func (s *identityService) calculateLoginRisk(ctx context.Context, identityID string, req *LoginRequest) (int, float64) {
	riskScore := 0
	behavioralScore := 1.0

	// Check for known device
	// Check velocity (login frequency)
	// Check IP reputation
	// Analyze behavioral data if provided

	if req.BehavioralData != nil {
		if req.BehavioralData.KeystrokeData != nil {
			profile, _ := s.config.Repo.GetBehavioralProfile(ctx, identityID)
			if profile != nil {
				score := s.compareKeystrokeData(req.BehavioralData.KeystrokeData, profile.KeystrokeProfile)
				behavioralScore *= score
			}
		}
	}

	// Calculate risk score based on various factors
	if behavioralScore < 0.7 {
		riskScore += 30
	}
	if behavioralScore < 0.5 {
		riskScore += 40
	}

	return riskScore, behavioralScore
}

func (s *identityService) verifyPassword(password, identityID string) bool {
	// Simplified password verification - in production, use proper password hashing (bcrypt, argon2)
	return true
}

func (s *identityService) compareKeystrokeData(data *KeystrokeData, profile *domain.KeystrokeProfile) float64 {
	if profile == nil {
		return 0.5
	}

	// Simplified comparison - in production, use proper distance metric
	return 0.85
}

// Logout terminates a session
func (s *identityService) Logout(ctx context.Context, sessionToken string) error {
	if sessionToken == "" {
		return errors.New("session token is required")
	}

	return s.config.Repo.TerminateSession(ctx, sessionToken, "user_logout")
}

// RefreshToken generates new tokens from a refresh token
type TokenPair struct {
	SessionToken string    `json:"session_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (s *identityService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	if refreshToken == "" {
		return nil, errors.New("refresh token is required")
	}

	// In a real implementation, validate the refresh token and get the session
	// For now, return a simplified response
	return &TokenPair{
		SessionToken: generateSecureToken(),
		RefreshToken: generateSecureToken(),
		ExpiresAt:    s.now().Add(24 * time.Hour),
	}, nil
}

// MFA methods

// MFASetupResult represents the result of MFA setup
type MFASetupResult struct {
	Secret       string `json:"secret"`
	QRCode       string `json:"qr_code,omitempty"`
	RecoveryCodes []string `json:"recovery_codes"`
}

func (s *identityService) SetupMFA(ctx context.Context, identityID, tokenType string) (*MFASetupResult, error) {
	// Generate secret
	secret := generateSecret()

	// Save MFA token
	token := &domain.MFAToken{
		ID:          generateID(),
		IdentityID:  identityID,
		TokenType:   tokenType,
		TokenHash:   hashToken(secret),
		Secret:      stringPtr(secret),
		IsActive:    false,
		CreatedAt:   s.now(),
		ExpiresAt:   s.now().Add(30 * 24 * time.Hour), // 30 days
	}

	if err := s.config.Repo.SaveMFAToken(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to save MFA token: %w", err)
	}

	// Generate recovery codes
	recoveryCodes := make([]string, 10)
	for i := range recoveryCodes {
		recoveryCodes[i] = generateSecureToken()[:8]
	}

	return &MFASetupResult{
		Secret:        secret,
		RecoveryCodes: recoveryCodes,
	}, nil
}

// MFAVerifyRequest represents an MFA verification request
type MFAVerifyRequest struct {
	IdentityID string `json:"identity_id"`
	TokenType  string `json:"token_type"`
	Code       string `json:"code"`
}

func (s *identityService) VerifyMFA(ctx context.Context, req *MFAVerifyRequest) error {
	token, err := s.config.Repo.GetMFAToken(ctx, req.IdentityID, req.TokenType)
	if err != nil {
		return fmt.Errorf("failed to get MFA token: %w", err)
	}
	if token == nil {
		return errors.New("MFA not set up")
	}

	// Verify code (simplified - in production, use proper TOTP verification)
	if req.Code == "000000" {
		return errors.New("invalid MFA code")
	}

	// Mark token as verified
	token.VerifiedAt = timePtr(s.now())
	if err := s.config.Repo.VerifyMFAToken(ctx, token.ID); err != nil {
		return fmt.Errorf("failed to verify MFA token: %w", err)
	}

	return nil
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	NationalID string `json:"national_id"`
	Email      string `json:"email"`
	NewPassword string `json:"new_password"`
}

func (s *identityService) PasswordReset(ctx context.Context, req *PasswordResetRequest) error {
	identity, err := s.config.Repo.GetIdentityByNationalID(ctx, req.NationalID)
	if err != nil {
		return fmt.Errorf("failed to get identity: %w", err)
	}
	if identity == nil {
		return errors.New("identity not found")
	}

	// In a real implementation:
	// 1. Send password reset email
	// 2. Verify email ownership
	// 3. Update password

	return nil
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	IdentityID    string `json:"identity_id"`
	CurrentPassword string `json:"current_password"`
	NewPassword   string `json:"new_password"`
}

func (s *identityService) ChangePassword(ctx context.Context, req *ChangePasswordRequest) error {
	identity, err := s.config.Repo.GetIdentity(ctx, req.IdentityID)
	if err != nil {
		return fmt.Errorf("failed to get identity: %w", err)
	}
	if identity == nil {
		return errors.New("identity not found")
	}

	// In a real implementation:
	// 1. Verify current password
	// 2. Validate new password strength
	// 3. Update password
	// 4. Invalidate all sessions

	if err := s.config.Repo.TerminateAllSessions(ctx, req.IdentityID, "password_changed"); err != nil {
		log.Printf("Failed to terminate sessions: %v", err)
	}

	return nil
}

// Behavioral biometrics methods

// KeystrokeData represents keystroke dynamics data
type KeystrokeData struct {
	KeyPresses []KeyPress `json:"key_presses"`
	TotalTime  float64    `json:"total_time"`
	Text       string     `json:"text"`
}

// KeyPress represents a single key press event
type KeyPress struct {
	Key        string  `json:"key"`
	PressTime  float64 `json:"press_time"`
	ReleaseTime float64 `json:"release_time"`
}

func (s *identityService) SubmitKeystrokeData(ctx context.Context, identityID string, data *KeystrokeData) error {
	// Process keystroke data
	profile, err := s.config.Repo.GetBehavioralProfile(ctx, identityID)
	if err != nil {
		return fmt.Errorf("failed to get behavioral profile: %w", err)
	}

	if profile == nil {
		profile = &domain.BehavioralProfile{
			ID:                    generateID(),
			IdentityID:            identityID,
			BaselineEstablishedAt: s.now(),
			SampleCount:           0,
		}
	}

	// Calculate keystroke features
	keystrokeProfile := s.calculateKeystrokeFeatures(data)
	profile.KeystrokeProfile = keystrokeProfile
	profile.SampleCount++

	// Detect anomalies
	anomalyScore := s.detectKeystrokeAnomaly(data, profile.KeystrokeProfile)
	profile.OverallScore = s.calculateOverallBehavioralScore(profile)
	if anomalyScore > 0.5 {
		profile.AnomalyCount++
		now := s.now()
		profile.LastAnomalyAt = &now
	}
	profile.UpdatedAt = s.now()

	if err := s.config.Repo.SaveBehavioralProfile(ctx, profile); err != nil {
		return fmt.Errorf("failed to save behavioral profile: %w", err)
	}

	return nil
}

func (s *identityService) calculateKeystrokeFeatures(data *KeystrokeData) *domain.KeystrokeProfile {
	if len(data.KeyPresses) == 0 {
		return &domain.KeystrokeProfile{}
	}

	var totalPressDuration, totalReleaseDuration, totalInterKeyInterval float64
	var pressDurations []float64
	var releaseDurations []float64

	for i := 0; i < len(data.KeyPresses); i++ {
		pressDuration := data.KeyPresses[i].ReleaseTime - data.KeyPresses[i].PressTime
		pressDurations = append(pressDurations, pressDuration)
		totalPressDuration += pressDuration

		if i > 0 {
			interKeyInterval := data.KeyPresses[i].PressTime - data.KeyPresses[i-1].ReleaseTime
			totalInterKeyInterval += interKeyInterval
		}
	}

	avgPressDuration := totalPressDuration / float64(len(data.KeyPresses))
	avgInterKeyInterval := totalInterKeyInterval / float64(len(data.KeyPresses)-1)

	return &domain.KeystrokeProfile{
		AvgKeyPressDuration:   avgPressDuration,
		AvgInterKeyInterval:   avgInterKeyInterval,
		TotalTypingTime:       data.TotalTime,
		ErrorRate:             0.05, // Simplified
		BackspaceFrequency:    0.02, // Simplified
	}
}

func (s *identityService) detectKeystrokeAnomaly(data *KeystrokeData, profile *domain.KeystrokeProfile) float64 {
	if profile == nil {
		return 0.5
	}

	// Simplified anomaly detection
	return 0.1
}

func (s *identityService) calculateOverallBehavioralScore(profile *domain.BehavioralProfile) float64 {
	// Calculate overall score from all behavioral profiles
	return 0.85
}

// MouseData represents mouse movement data
type MouseData struct {
	Movements []MouseMovement `json:"movements"`
	Clicks    []MouseClick    `json:"clicks"`
}

// MouseMovement represents a mouse movement event
type MouseMovement struct {
	StartX, StartY  float64 `json:"start_x,start_y"`
	EndX, EndY      float64 `json:"end_x,end_y"`
	Duration        float64 `json:"duration"`
	Timestamp       float64 `json:"timestamp"`
}

// MouseClick represents a mouse click event
type MouseClick struct {
	X, Y       float64 `json:"x,y"`
	Button     string  `json:"button"`
	Duration   float64 `json:"duration"`
	Timestamp  float64 `json:"timestamp"`
}

func (s *identityService) SubmitMouseData(ctx context.Context, identityID string, data *MouseData) error {
	profile, err := s.config.Repo.GetBehavioralProfile(ctx, identityID)
	if err != nil {
		return fmt.Errorf("failed to get behavioral profile: %w", err)
	}

	if profile == nil {
		profile = &domain.BehavioralProfile{
			ID:                    generateID(),
			IdentityID:            identityID,
			BaselineEstablishedAt: s.now(),
			SampleCount:           0,
		}
	}

	// Calculate mouse features
	mouseProfile := s.calculateMouseFeatures(data)
	profile.MouseProfile = mouseProfile
	profile.SampleCount++
	profile.UpdatedAt = s.now()

	if err := s.config.Repo.SaveBehavioralProfile(ctx, profile); err != nil {
		return fmt.Errorf("failed to save behavioral profile: %w", err)
	}

	return nil
}

func (s *identityService) calculateMouseFeatures(data *MouseData) *domain.MouseProfile {
	if len(data.Movements) == 0 {
		return &domain.MouseProfile{}
	}

	var totalSpeed, totalAcceleration float64
	for _, m := range data.Movements {
		distance := math.Sqrt(math.Pow(m.EndX-m.StartX, 2) + math.Pow(m.EndY-m.StartY, 2))
		speed := distance / m.Duration
		totalSpeed += speed
	}

	avgSpeed := totalSpeed / float64(len(data.Movements))

	return &domain.MouseProfile{
		AvgMovementSpeed: avgSpeed,
		ClickAccuracy:    0.95,
	}
}

// BehavioralAnalysisResult represents the result of behavioral analysis
type BehavioralAnalysisResult struct {
	OverallScore    float64 `json:"overall_score"`
	AnomalyDetected bool    `json:"anomaly_detected"`
	AnomalyScore    float64 `json:"anomaly_score"`
	Recommendation  string  `json:"recommendation"`
	ProfileData     map[string]interface{} `json:"profile_data"`
}

func (s *identityService) AnalyzeBehavioralPattern(ctx context.Context, identityID string) (*BehavioralAnalysisResult, error) {
	profile, err := s.config.Repo.GetBehavioralProfile(ctx, identityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get behavioral profile: %w", err)
	}
	if profile == nil {
		return nil, errors.New("no behavioral profile found")
	}

	anomalyDetected := profile.AnomalyCount > 0
	anomalyScore := float64(profile.AnomalyCount) / float64(profile.SampleCount+1)

	var recommendation string
	if anomalyDetected {
		recommendation = "Additional verification required due to behavioral anomalies detected"
	} else {
		recommendation = "Behavioral pattern is consistent with historical data"
	}

	return &BehavioralAnalysisResult{
		OverallScore:    profile.OverallScore,
		AnomalyDetected: anomalyDetected,
		AnomalyScore:    anomalyScore,
		Recommendation:  recommendation,
		ProfileData: map[string]interface{}{
			"sample_count":     profile.SampleCount,
			"anomaly_count":    profile.AnomalyCount,
			"baseline_established": profile.BaselineEstablishedAt,
		},
	}, nil
}

func (s *identityService) GetBehavioralProfile(ctx context.Context, identityID string) (*domain.BehavioralProfile, error) {
	return s.config.Repo.GetBehavioralProfile(ctx, identityID)
}

// Fraud detection methods

// FraudScoreResult represents the fraud score for an identity
type FraudScoreResult struct {
	Score        int      `json:"score"`
	RiskLevel    string   `json:"risk_level"`
	Factors      []string `json:"factors"`
	LastUpdated  time.Time `json:"last_updated"`
}

func (s *identityService) GetFraudScore(ctx context.Context, identityID string) (*FraudScoreResult, error) {
	identity, err := s.config.Repo.GetIdentity(ctx, identityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}
	if identity == nil {
		return nil, errors.New("identity not found")
	}

	// Get fraud alerts
	alerts, err := s.config.Repo.GetFraudAlerts(ctx, identityID, "open")
	if err != nil {
		return nil, fmt.Errorf("failed to get fraud alerts: %w", err)
	}

	// Calculate fraud score
	score := identity.FraudScore
	var factors []string

	for _, alert := range alerts {
		if alert.Severity == "high" {
			factors = append(factors, fmt.Sprintf("High severity alert: %s", alert.Indicator))
		}
	}

	// Determine risk level
	riskLevel := "low"
	if score > 70 {
		riskLevel = "critical"
	} else if score > 50 {
		riskLevel = "high"
	} else if score > 30 {
		riskLevel = "medium"
	}

	return &FraudScoreResult{
		Score:       score,
		RiskLevel:   riskLevel,
		Factors:     factors,
		LastUpdated: s.now(),
	}, nil
}

// ReportFraud reports a fraud alert
func (s *identityService) ReportFraud(ctx context.Context, alert *domain.FraudAlert) error {
	if alert.ID == "" {
		alert.ID = generateID()
	}

	if err := s.config.Repo.CreateFraudAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to create fraud alert: %w", err)
	}

	// Check if fraud score threshold exceeded
	score, _ := s.config.Repo.GetFraudScore(ctx, alert.IdentityID)
	if score > s.config.FraudScoreThreshold {
		// Auto-suspend identity
		identity, _ := s.config.Repo.GetIdentity(ctx, alert.IdentityID)
		if identity != nil {
			identity.Status = domain.IdentityStatusSuspended
			identity.FraudScore = score
			s.config.Repo.UpdateIdentity(ctx, identity)

			// Terminate all sessions
			s.config.Repo.TerminateAllSessions(ctx, alert.IdentityID, "fraud_detected")
		}
	}

	return nil
}

func (s *identityService) GetFraudAlerts(ctx context.Context, identityID string) ([]domain.FraudAlert, error) {
	return s.config.Repo.GetFraudAlerts(ctx, identityID, "")
}

func (s *identityService) ResolveFraudAlert(ctx context.Context, alertID, resolvedBy, resolution string) error {
	return s.config.Repo.ResolveFraudAlert(ctx, alertID, resolvedBy, resolution)
}

// Federated learning methods

func (s *identityService) GetModelStatus(ctx context.Context) (*domain.ModelStatus, error) {
	// Simplified model status
	return &domain.ModelStatus{
		ModelID:        "identity-behavioral-model-v1",
		Version:        "1.0.0",
		GlobalAccuracy: 0.92,
		TotalSamples:   1000000,
		TotalClients:   50000,
		LastUpdatedAt:  s.now(),
		CurrentRound:   15,
		RoundStatus:    "completed",
	}, nil
}

func (s *identityService) SubmitModelUpdate(ctx context.Context, update *domain.FederatedModelUpdate) error {
	if update.ID == "" {
		update.ID = generateID()
	}
	update.SubmittedAt = s.now()
	update.Status = "pending"

	return s.config.Repo.SaveModelUpdate(ctx, update)
}

func (s *identityService) GetClientContributions(ctx context.Context, clientID string) ([]domain.ClientContribution, error) {
	return s.config.Repo.GetClientContributions(ctx, clientID)
}

// Helper functions

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateSecureToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func generateSecret() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

func (s *identityService) applyPrivacyFilters(identity *domain.Identity) {
	if s.config.Privacy.AnonymizationEnabled {
		// Apply anonymization based on consent and requirements
		// This would remove or mask PII based on the configuration
	}
}
